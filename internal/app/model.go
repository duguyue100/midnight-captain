package app

import (
	"os"

	"github.com/dgyhome/midnight-captain/internal/fs"
	"github.com/dgyhome/midnight-captain/internal/ops"
	"github.com/dgyhome/midnight-captain/internal/ui/cmdpalette"
	"github.com/dgyhome/midnight-captain/internal/ui/dialog"
	goto_ "github.com/dgyhome/midnight-captain/internal/ui/goto"
	"github.com/dgyhome/midnight-captain/internal/ui/help"
	"github.com/dgyhome/midnight-captain/internal/ui/pane"
	"github.com/dgyhome/midnight-captain/internal/ui/search"
	"github.com/dgyhome/midnight-captain/internal/ui/statusbar"
)

// FocusPane indicates which pane is active.
type FocusPane int

const (
	FocusLeft  FocusPane = 0
	FocusRight FocusPane = 1
)

// Model is the root bubbletea model.
type Model struct {
	Left       pane.Model
	Right      pane.Model
	Statusbar  statusbar.Model
	CmdPalette cmdpalette.Model
	Search     search.Model
	Confirm    dialog.ConfirmModel
	Input      dialog.InputModel
	Help       help.Model
	Goto       goto_.Model
	Clipboard  ops.Clipboard
	Focus      FocusPane
	Width      int
	Height     int
	lastKey    string
	// active operations (for progress display)
	Operations []ops.Operation
}

// NewModel creates the initial app state.
func NewModel() Model {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}

	localFS := fs.NewLocalFS()
	left := pane.New(localFS, cwd)
	left.Focused = true

	right := pane.New(localFS, cwd)
	right.Focused = false

	return Model{
		Left:       left,
		Right:      right,
		Statusbar:  statusbar.New(80),
		CmdPalette: cmdpalette.New(),
		Search:     search.New(),
		Confirm:    dialog.NewConfirm("", "", nil),
		Input:      dialog.NewInput("", "", ""),
		Help:       help.New(),
		Goto:       goto_.New(),
		Focus:      FocusLeft,
	}
}

func (m *Model) activePane() *pane.Model {
	if m.Focus == FocusRight {
		return &m.Right
	}
	return &m.Left
}

func (m *Model) inactivePane() *pane.Model {
	if m.Focus == FocusRight {
		return &m.Left
	}
	return &m.Right
}

func (m *Model) setFocus(f FocusPane) {
	m.Focus = f
	m.Left.Focused = f == FocusLeft
	m.Right.Focused = f == FocusRight
}

func (m *Model) propagateSizes() {
	dual := m.Width >= 80
	sbHeight := 1

	paneHeight := m.Height - sbHeight
	if paneHeight < 3 {
		paneHeight = 3
	}

	if dual {
		paneW := (m.Width - 1) / 2
		m.Left.SetSize(paneW, paneHeight)
		m.Right.SetSize(paneW, paneHeight)
	} else {
		m.Left.SetSize(m.Width, paneHeight)
		m.Right.SetSize(0, 0)
	}
	m.Statusbar.SetSize(m.Width)
	m.CmdPalette.SetSize(m.Width, m.Height)
	m.Search.SetSize(m.Width, m.Height)
	m.Help.SetSize(m.Width, m.Height)
	m.Goto.SetSize(m.Width, m.Height)
}

// selectedPaths returns absolute paths of selected (or cursor) entries in ap.
func selectedPaths(ap *pane.Model) []string {
	if len(ap.Selected) > 0 {
		paths := make([]string, 0, len(ap.Selected))
		for idx := range ap.Selected {
			if idx < len(ap.Nodes) {
				paths = append(paths, ap.Nodes[idx].FullPath)
			}
		}
		return paths
	}
	node, ok := ap.CurrentNode()
	if !ok {
		return nil
	}
	return []string{node.FullPath}
}

// selectedNames returns just the names (for dialog display).
func selectedNames(ap *pane.Model) []string {
	paths := selectedPaths(ap)
	names := make([]string, len(paths))
	for i, p := range paths {
		names[i] = baseName(p)
	}
	return names
}

func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
