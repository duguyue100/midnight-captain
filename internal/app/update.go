package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/bubbletea/v2"
	appfs "github.com/dgyhome/midnight-captain/internal/fs"
	"github.com/dgyhome/midnight-captain/internal/ops"
	"github.com/dgyhome/midnight-captain/internal/ui/cmdpalette"
	"github.com/dgyhome/midnight-captain/internal/ui/dialog"
	goto_ "github.com/dgyhome/midnight-captain/internal/ui/goto"
	"github.com/dgyhome/midnight-captain/internal/ui/pane"
	"github.com/dgyhome/midnight-captain/internal/ui/search"
	"github.com/dgyhome/midnight-captain/internal/ui/statusbar"
)

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is the root update function.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.propagateSizes()
		return m, nil

	case SSHConnectedMsg:
		ap := m.activePane()
		ap.FS = msg.Remote
		ap.Cwd = "/"
		ap.Reload()
		m.Statusbar.Message = "Connected: " + msg.Label
		return m, nil

	case SSHErrorMsg:
		m.Statusbar.Message = "SSH error: " + msg.Err.Error()
		return m, nil

	case statusbar.TickMsg:
		sb, cmd := m.Statusbar.Update(msg)
		m.Statusbar = sb
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case ops.ProgressMsg:
		cmd := m.handleProgress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case cmdpalette.ExecuteMsg:
		cmd := m.handleCommand(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case dialog.ConfirmResultMsg:
		cmd := m.handleConfirmResult(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case dialog.InputResultMsg:
		cmd := m.handleInputResult(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case search.ResultsMsg:
		s, cmd := m.Search.Update(msg)
		m.Search = s
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case goto_.NavigateMsg:
		ap := m.activePane()
		ap.NavigateTo(msg.Dir, "")
		return m, nil

	case search.NavigateMsg:
		ap := m.activePane()
		ap.NavigateTo(msg.Dir, msg.Name)
		return m, tea.Batch(cmds...)

	case statusMsg:
		m.Statusbar.Message = string(msg)
		return m, nil

	case tea.KeyPressMsg:
		// Confirm dialog intercepts all keys
		if m.Confirm.Visible {
			confirm, cmd := m.Confirm.Update(msg)
			m.Confirm = confirm
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Input dialog intercepts all keys
		if m.Input.Visible {
			input, cmd := m.Input.Update(msg)
			m.Input = input
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Help overlay intercepts all keys
		if m.Help.Visible {
			key := msg.String()
			if key == "esc" || key == "?" || key == "q" {
				m.Help.Close()
			}
			return m, nil
		}

		// Goto overlay intercepts all keys
		if m.Goto.Visible {
			g, cmd := m.Goto.Update(msg)
			m.Goto = g
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Search overlay intercepts all keys when open
		if m.Search.Visible {
			s, cmd := m.Search.Update(msg)
			m.Search = s
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Palette intercepts all keys when open
		if m.CmdPalette.Visible {
			palette, cmd := m.CmdPalette.Update(msg)
			m.CmdPalette = palette
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// Normal mode
		consumed, cmd := m.handleGlobalKey(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if consumed {
			return m, tea.Batch(cmds...)
		}

		// Forward to active pane
		if m.Focus == FocusLeft {
			left, cmd := m.Left.Update(msg)
			m.Left = left
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else {
			right, cmd := m.Right.Update(msg)
			m.Right = right
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	default:
		sb, cmd := m.Statusbar.Update(msg)
		m.Statusbar = sb
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

type statusMsg string

// handleGlobalKey returns (consumed bool, cmd).
func (m *Model) handleGlobalKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	key := msg.String()

	switch key {
	case keyQuit:
		return true, tea.Quit

	case keyColon:
		m.lastKey = ""
		return true, m.CmdPalette.Open()

	case keySearch:
		m.lastKey = ""
		ap := m.activePane()
		names := ap.EntryNames()
		return true, m.Search.OpenLocal(ap.Cwd, names)

	case keySlash:
		// kept as no-op to avoid forwarding to pane
		m.lastKey = ""
		return true, nil

	case keyTab:
		if m.Focus == FocusLeft {
			m.setFocus(FocusRight)
		} else {
			m.setFocus(FocusLeft)
		}
		m.lastKey = ""
		return true, nil

	case keyYank:
		ap := m.activePane()
		paths := selectedPaths(ap)
		if len(paths) > 0 {
			m.Clipboard = ops.Clipboard{Entries: paths, FS: ap.FS, Op: ops.ClipCopy}
			ap.Selected = make(map[int]bool)
			m.Statusbar.Message = fmt.Sprintf("Yanked %d item(s)", len(paths))
		}
		m.lastKey = ""
		return true, nil

	case keyCut:
		ap := m.activePane()
		paths := selectedPaths(ap)
		if len(paths) > 0 {
			m.Clipboard = ops.Clipboard{Entries: paths, FS: ap.FS, Op: ops.ClipCut}
			ap.Selected = make(map[int]bool)
			m.Statusbar.Message = fmt.Sprintf("Cut %d item(s)", len(paths))
		}
		m.lastKey = ""
		return true, nil

	case keyPaste:
		m.lastKey = ""
		return true, m.doPaste()

	case keyDelete:
		ap := m.activePane()
		names := selectedNames(ap)
		if len(names) == 0 {
			return true, nil
		}
		m.Confirm = dialog.NewConfirm("delete", fmt.Sprintf("Delete %d item(s)?", len(names)), names)
		m.Confirm.Open()
		m.lastKey = ""
		return true, nil

	case keyCreate:
		ap := m.activePane()
		// Base dir: if cursor on dir, use that dir; if on file, use its parent
		baseDir := ap.Cwd
		node, ok := ap.CurrentNode()
		if ok && node.Entry.Name != ".." {
			if node.Entry.IsDir {
				baseDir = node.FullPath
			} else {
				baseDir = filepath.Dir(node.FullPath)
			}
		}
		// Show base dir as hint in title, input starts empty
		rel, _ := filepath.Rel(ap.Cwd, baseDir)
		title := "New (path relative to: " + rel + ")"
		if baseDir == ap.Cwd {
			title = "New (in current dir)"
		}
		m.Input = dialog.NewInput("create:"+baseDir, title, "")
		m.lastKey = ""
		return true, m.Input.Open()

	case keyRename:
		ap := m.activePane()
		node, ok := ap.CurrentNode()
		if !ok {
			return true, nil
		}
		m.Input = dialog.NewInput("rename", "Rename: "+node.Entry.Name, node.Entry.Name)
		m.lastKey = ""
		return true, m.Input.Open()

	case keyOpen:
		ap := m.activePane()
		node, ok := ap.CurrentNode()
		if !ok {
			return true, nil
		}
		if !node.Entry.IsDir {
			return true, openInNvim(node.FullPath)
		}
		m.lastKey = ""
		return false, nil

	case keyGG:
		if m.lastKey == "g" {
			// gg — forward top key to pane
			m.lastKey = ""
			// We'll let pane handle 'g' for top
			return false, nil
		}
		m.lastKey = "g"
		return false, nil // forward 'g' to pane

	case keyHelp:
		m.lastKey = ""
		m.Help.Open()
		return true, nil

	default:
		m.lastKey = key
		return false, nil
	}
}

func (m *Model) doPaste() tea.Cmd {
	if len(m.Clipboard.Entries) == 0 {
		return nil
	}
	ap := m.activePane()
	destDir := ap.Cwd
	srcFS := m.Clipboard.FS
	dstFS := ap.FS
	sources := m.Clipboard.Entries
	clipOp := m.Clipboard.Op

	id := fmt.Sprintf("op-%p", &sources)

	var cmd tea.Cmd
	if clipOp == ops.ClipCut {
		m.Clipboard = ops.Clipboard{} // clear clipboard after cut-paste
		cmd = ops.Move(id, sources, destDir, srcFS, dstFS)
	} else {
		cmd = ops.Copy(id, sources, destDir, srcFS, dstFS)
	}

	m.Statusbar.Message = "Pasting…"
	return tea.Batch(cmd, m.Statusbar.StartSpinner())
}

func (m *Model) handleProgress(msg ops.ProgressMsg) tea.Cmd {
	ap := m.activePane()
	switch msg.Status {
	case ops.StatusRunning:
		pct := 0
		if msg.TotalBytes > 0 {
			pct = int(msg.DoneBytes * 100 / msg.TotalBytes)
		}
		m.Statusbar.SetProgress(true, pct)
		if !m.Statusbar.Active {
			return m.Statusbar.StartSpinner()
		}
	case ops.StatusDone:
		m.Statusbar.StopSpinner()
		m.Statusbar.Message = "Done."
		ap.Reload()
		m.inactivePane().Reload()
	case ops.StatusFailed:
		m.Statusbar.StopSpinner()
		if msg.Err != nil {
			m.Statusbar.Message = "Error: " + msg.Err.Error()
		}
	}
	return nil
}

func (m *Model) handleConfirmResult(msg dialog.ConfirmResultMsg) tea.Cmd {
	if !msg.Confirmed {
		return nil
	}
	switch msg.ID {
	case "delete":
		ap := m.activePane()
		paths := selectedPaths(ap)
		ap.Selected = make(map[int]bool)
		id := fmt.Sprintf("del-%d", len(paths))
		return ops.Delete(id, paths, ap.FS)
	}
	return nil
}

func (m *Model) handleInputResult(msg dialog.InputResultMsg) tea.Cmd {
	if msg.Cancelled || msg.Value == "" {
		return nil
	}
	ap := m.activePane()
	switch msg.ID {
	case "rename":
		node, ok := ap.CurrentNode()
		if !ok {
			return nil
		}
		oldPath := node.FullPath
		cmd := ops.Rename(oldPath, msg.Value, ap.FS)
		return tea.Batch(cmd, func() tea.Msg {
			return ops.ProgressMsg{OpID: "rename", Status: ops.StatusRunning}
		})
	default:
		// "create:<baseDir>" — smart create: dir if trailing /, else file
		if strings.HasPrefix(msg.ID, "create:") {
			baseDir := strings.TrimPrefix(msg.ID, "create:")
			m.handleCreate(ap, baseDir, msg.Value)
		}
	}
	return nil
}

func (m *Model) handleCommand(msg cmdpalette.ExecuteMsg) tea.Cmd {
	ap := m.activePane()
	switch msg.Name {
	case "hidden":
		ap.ShowHidden = !ap.ShowHidden
		ap.Reload()

	case "sort":
		arg := ""
		if len(msg.Args) > 0 {
			arg = msg.Args[0]
		}
		switch arg {
		case "size":
			ap.SortBy = 1
		case "date":
			ap.SortBy = 2
		default:
			ap.SortBy = 0
		}
		ap.Reload()

	case "ssh":
		target := ""
		if len(msg.Args) > 0 {
			target = msg.Args[0]
		}
		if target == "" {
			m.Statusbar.Message = "Usage: ssh user@host"
			return nil
		}
		m.Statusbar.Message = "Connecting to " + target + "…"
		return sshConnect(m, target)

	case "disconnect":
		if !ap.FS.IsLocal() {
			ap.FS = appfs.NewLocalFS()
			ap.Cwd, _ = os.Getwd()
			ap.Reload()
			m.Statusbar.Message = "Disconnected."
		}

	case "find":
		return m.Search.OpenRecursive(ap.Cwd)

	case "goto":
		arg := ""
		if len(msg.Args) > 0 {
			arg = msg.Args[0]
		}
		return m.Goto.Open(arg)

	case "quit":
		return tea.Quit
	}
	return nil
}

func openInNvim(path string) tea.Cmd {
	return tea.ExecProcess(exec.Command("nvim", path), func(err error) tea.Msg {
		if err != nil {
			return statusMsg("nvim: " + err.Error())
		}
		return statusMsg("")
	})
}

// handleCreate implements smart create for the `a` key.
// baseDir is the directory to create relative to.
// input rules:
//   - trailing `/`        → create directory (all components)
//   - `something/file`    → mkdir -p something, then create file
//   - `file.txt`          → create file in baseDir
func (m *Model) handleCreate(ap *pane.Model, baseDir, input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	isDir := strings.HasSuffix(input, "/")
	fullPath := filepath.Join(baseDir, input)

	if isDir {
		// Create directory (and all parents)
		if err := mkdirAllFS(ap.FS, fullPath); err != nil {
			m.Statusbar.Message = "mkdir: " + err.Error()
			return
		}
	} else {
		// Ensure parent dirs exist
		parentDir := filepath.Dir(fullPath)
		if err := mkdirAllFS(ap.FS, parentDir); err != nil {
			m.Statusbar.Message = "mkdir: " + err.Error()
			return
		}
		// Create file
		f, err := ap.FS.Create(fullPath, 0o644)
		if err != nil {
			m.Statusbar.Message = "create: " + err.Error()
			return
		}
		f.Close()
	}
	ap.Reload()
	m.Statusbar.Message = "Created: " + input
}

// mkdirAllFS creates path and all parents using FS.Mkdir, ignoring already-exists errors.
func mkdirAllFS(fsys appfs.FileSystem, path string) error {
	// Collect all path components from root down
	parts := []string{}
	p := filepath.Clean(path)
	for p != "/" && p != "." {
		parts = append([]string{p}, parts...)
		p = filepath.Dir(p)
	}
	for _, part := range parts {
		err := fsys.Mkdir(part, 0o755)
		if err != nil {
			// Ignore "already exists" errors
			if !os.IsExist(err) {
				// Try stat — if it exists as a dir, continue
				info, serr := fsys.Stat(part)
				if serr == nil {
					if info.IsDir {
						continue
					}
					return fmt.Errorf("mkdir: %s is not a directory", part)
				}
				return err
			}
		}
	}
	return nil
}
