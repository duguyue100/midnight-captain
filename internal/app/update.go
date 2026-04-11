package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"charm.land/bubbletea/v2"
	appfs "github.com/dgyhome/midnight-captain/internal/fs"
	"github.com/dgyhome/midnight-captain/internal/ops"
	"github.com/dgyhome/midnight-captain/internal/ui/cmdpalette"
	"github.com/dgyhome/midnight-captain/internal/ui/dialog"
	"github.com/dgyhome/midnight-captain/internal/ui/search"
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

	case keySlash:
		m.lastKey = ""
		ap := m.activePane()
		return true, m.Search.Open(ap.Cwd)

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

	case keyRename:
		ap := m.activePane()
		entry, ok := ap.CurrentEntry()
		if !ok {
			return true, nil
		}
		m.Input = dialog.NewInput("rename", "Rename: "+entry.Name, entry.Name)
		m.lastKey = ""
		return true, m.Input.Open()

	case keyOpen:
		ap := m.activePane()
		entry, ok := ap.CurrentEntry()
		if !ok {
			return true, nil
		}
		if !entry.IsDir {
			path := ap.Cwd + "/" + entry.Name
			return true, openInNvim(path)
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
		return true, nil // help overlay — TODO Phase 5

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
	return cmd
}

func (m *Model) handleProgress(msg ops.ProgressMsg) tea.Cmd {
	ap := m.activePane()
	switch msg.Status {
	case ops.StatusDone:
		m.Statusbar.Message = "Done."
		ap.Reload()
		m.inactivePane().Reload()
	case ops.StatusFailed:
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
		entry, ok := ap.CurrentEntry()
		if !ok {
			return nil
		}
		oldPath := ap.Cwd + "/" + entry.Name
		cmd := ops.Rename(oldPath, msg.Value, ap.FS)
		return tea.Batch(cmd, func() tea.Msg {
			return ops.ProgressMsg{OpID: "rename", Status: ops.StatusRunning}
		})
	case "mkdir":
		path := filepath.Join(ap.Cwd, msg.Value)
		if err := ap.FS.Mkdir(path, 0o755); err != nil {
			m.Statusbar.Message = "mkdir: " + err.Error()
		} else {
			ap.Reload()
		}
	case "touch":
		path := filepath.Join(ap.Cwd, msg.Value)
		f, err := ap.FS.Create(path, 0o644)
		if err != nil {
			m.Statusbar.Message = "touch: " + err.Error()
		} else {
			f.Close()
			ap.Reload()
		}
	}
	return nil
}

func (m *Model) handleCommand(msg cmdpalette.ExecuteMsg) tea.Cmd {
	ap := m.activePane()
	switch msg.Name {
	case "/hidden":
		ap.ShowHidden = !ap.ShowHidden
		ap.Reload()

	case "/sort":
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

	case "/mkdir":
		name := ""
		if len(msg.Args) > 0 {
			name = msg.Args[0]
		}
		if name != "" {
			path := filepath.Join(ap.Cwd, name)
			if err := ap.FS.Mkdir(path, 0o755); err != nil {
				m.Statusbar.Message = "mkdir: " + err.Error()
			} else {
				ap.Reload()
			}
		} else {
			m.Input = dialog.NewInput("mkdir", "New directory name:", "")
			return m.Input.Open()
		}

	case "/touch":
		name := ""
		if len(msg.Args) > 0 {
			name = msg.Args[0]
		}
		if name != "" {
			path := filepath.Join(ap.Cwd, name)
			f, err := ap.FS.Create(path, 0o644)
			if err != nil {
				m.Statusbar.Message = "touch: " + err.Error()
			} else {
				f.Close()
				ap.Reload()
			}
		} else {
			m.Input = dialog.NewInput("touch", "New file name:", "")
			return m.Input.Open()
		}

	case "/ssh":
		target := ""
		if len(msg.Args) > 0 {
			target = msg.Args[0]
		}
		if target == "" {
			m.Statusbar.Message = "Usage: /ssh user@host"
			return nil
		}
		m.Statusbar.Message = "Connecting to " + target + "…"
		return sshConnect(m, target)

	case "/disconnect":
		if !ap.FS.IsLocal() {
			ap.FS = appfs.NewLocalFS()
			ap.Cwd, _ = os.Getwd()
			ap.Reload()
			m.Statusbar.Message = "Disconnected."
		}

	case "/quit":
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
