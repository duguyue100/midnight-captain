package pane

import (
	"github.com/dgyhome/midnight-captain/internal/fs"
)

// SortField determines sort order.
type SortField int

const (
	SortByName SortField = iota
	SortBySize
	SortByMtime
)

// Model is the file pane sub-model.
type Model struct {
	FS           fs.FileSystem
	Cwd          string
	Entries      []fs.FileEntry
	Cursor       int
	Offset       int
	Selected     map[int]bool
	VisualMode   bool
	VisualAnchor int // index where V was pressed
	Focused      bool
	Width        int
	Height       int
	SortBy       SortField
	SortAsc      bool
	ShowHidden   bool
	Err          error
}

// New creates a pane rooted at startDir using the given filesystem.
func New(filesystem fs.FileSystem, startDir string) Model {
	m := Model{
		FS:         filesystem,
		Cwd:        startDir,
		Selected:   make(map[int]bool),
		SortBy:     SortByName,
		SortAsc:    true,
		ShowHidden: true,
	}
	entries, err := filesystem.List(startDir)
	m.Err = err
	if err == nil {
		filtered := filterEntries(entries, m.ShowHidden)
		sortEntries(filtered, m.SortBy, m.SortAsc)
		if startDir != "/" {
			filtered = append([]fs.FileEntry{parentEntry()}, filtered...)
		}
		m.Entries = filtered
	}
	return m
}

// SetSize updates the pane dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// Reload re-reads the current directory.
func (m *Model) Reload() {
	entries, err := m.FS.List(m.Cwd)
	m.Err = err
	if err == nil {
		filtered := filterEntries(entries, m.ShowHidden)
		sortEntries(filtered, m.SortBy, m.SortAsc)
		if m.Cwd != "/" {
			filtered = append([]fs.FileEntry{parentEntry()}, filtered...)
		}
		m.Entries = filtered
	}
	if m.Cursor >= len(m.Entries) {
		m.Cursor = max(0, len(m.Entries)-1)
	}
	m.clampOffset()
}

// CurrentEntry returns the entry under the cursor, if any.
func (m Model) CurrentEntry() (fs.FileEntry, bool) {
	if len(m.Entries) == 0 || m.Cursor >= len(m.Entries) {
		return fs.FileEntry{}, false
	}
	return m.Entries[m.Cursor], true
}

// clampOffset adjusts scroll offset so cursor is visible.
// Must be called on the local copy after mutation.
func (m *Model) clampOffset() {
	visible := m.visibleHeight()
	if m.Cursor < m.Offset {
		m.Offset = m.Cursor
	} else if m.Cursor >= m.Offset+visible {
		m.Offset = m.Cursor - visible + 1
	}
	if m.Offset < 0 {
		m.Offset = 0
	}
}

// visibleHeight returns number of file rows shown inside the pane border.
func (m Model) visibleHeight() int {
	// Height - 2 (top+bottom border) - 1 (cwd header) - 1 (col header)
	h := m.Height - 4
	if h < 1 {
		h = 1
	}
	return h
}

// EntryCount returns number of entries in current directory.
func (m Model) EntryCount() int { return len(m.Entries) }

// SelectedCount returns number of selected entries.
func (m Model) SelectedCount() int { return len(m.Selected) }

// GetCwd returns current working directory.
func (m Model) GetCwd() string { return m.Cwd }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
