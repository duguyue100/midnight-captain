package pane

import (
	"path/filepath"

	"github.com/dgyhome/midnight-captain/internal/fs"
)

// SortField determines sort order.
type SortField int

const (
	SortByName SortField = iota
	SortBySize
	SortByMtime
)

// TreeNode is one visible row in the flat expanded tree.
type TreeNode struct {
	Entry    fs.FileEntry
	Depth    int    // 0 = top-level (Cwd children), 1 = one level expanded, etc.
	FullPath string // absolute path
	Expanded bool   // only meaningful for dirs
}

// Model is the file pane sub-model.
type Model struct {
	FS           fs.FileSystem
	Cwd          string
	Nodes        []TreeNode // flat visible list (replaces Entries)
	Cursor       int
	Offset       int
	Selected     map[int]bool
	VisualMode   bool
	VisualAnchor int
	Focused      bool
	Width        int
	Height       int
	SortBy       SortField
	SortAsc      bool
	ShowHidden   bool
	Err          error

	expandState map[string]bool // path → expanded; persists across reloads
}

// New creates a pane rooted at startDir using the given filesystem.
func New(filesystem fs.FileSystem, startDir string) Model {
	m := Model{
		FS:          filesystem,
		Cwd:         startDir,
		Selected:    make(map[int]bool),
		SortBy:      SortByName,
		SortAsc:     true,
		ShowHidden:  true,
		expandState: make(map[string]bool),
	}
	m.Reload()
	return m
}

// SetSize updates the pane dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// Reload re-reads the current directory and rebuilds the flat node list.
func (m *Model) Reload() {
	if m.expandState == nil {
		m.expandState = make(map[string]bool)
	}
	m.Nodes = m.buildNodes(m.Cwd, 0)
	if m.Cursor >= len(m.Nodes) {
		m.Cursor = max(0, len(m.Nodes)-1)
	}
	m.clampOffset()
}

// buildNodes builds the flat visible list for dir at given depth.
// Recursively expands dirs whose paths are in expandState.
func (m *Model) buildNodes(dir string, depth int) []TreeNode {
	entries, err := m.FS.List(dir)
	if err != nil {
		if depth == 0 {
			m.Err = err
		}
		return nil
	}
	if depth == 0 {
		m.Err = nil
	}

	filtered := filterEntries(entries, m.ShowHidden)
	sortEntries(filtered, m.SortBy, m.SortAsc)

	var nodes []TreeNode

	// ".." only at top level
	if depth == 0 && dir != "/" {
		nodes = append(nodes, TreeNode{
			Entry:    parentEntry(),
			Depth:    0,
			FullPath: filepath.Dir(dir),
		})
	}

	for _, e := range filtered {
		fullPath := filepath.Join(dir, e.Name)
		expanded := e.IsDir && m.expandState[fullPath]
		node := TreeNode{
			Entry:    e,
			Depth:    depth,
			FullPath: fullPath,
			Expanded: expanded,
		}
		nodes = append(nodes, node)
		if expanded {
			children := m.buildNodes(fullPath, depth+1)
			nodes = append(nodes, children...)
		}
	}
	return nodes
}

// toggleExpand expands or collapses the dir at cursor index.
// Returns the new cursor position (stays on same node).
func (m *Model) toggleExpand(idx int) {
	if idx < 0 || idx >= len(m.Nodes) {
		return
	}
	node := m.Nodes[idx]
	if !node.Entry.IsDir || node.Entry.Name == ".." {
		return
	}
	if node.Expanded {
		m.expandState[node.FullPath] = false
	} else {
		m.expandState[node.FullPath] = true
	}
	m.Reload()
	// Re-find the same node after rebuild
	for i, n := range m.Nodes {
		if n.FullPath == node.FullPath {
			m.Cursor = i
			m.clampOffset()
			return
		}
	}
}

// collapseAtCursor collapses an expanded dir OR jumps to parent node.
// Returns true if handled, false if should fall through to goParent.
func (m *Model) collapseAtCursor() bool {
	if m.Cursor < 0 || m.Cursor >= len(m.Nodes) {
		return false
	}
	node := m.Nodes[m.Cursor]

	// If cursor is on an expanded dir, collapse it
	if node.Entry.IsDir && node.Expanded {
		m.expandState[node.FullPath] = false
		m.Reload()
		for i, n := range m.Nodes {
			if n.FullPath == node.FullPath {
				m.Cursor = i
				m.clampOffset()
				return true
			}
		}
		return true
	}

	// If cursor is a child (depth > 0), jump to parent node
	if node.Depth > 0 {
		parentPath := filepath.Dir(node.FullPath)
		for i, n := range m.Nodes {
			if n.FullPath == parentPath {
				m.Cursor = i
				m.clampOffset()
				return true
			}
		}
		return true
	}

	return false // let goParent handle it
}

// CurrentEntry returns the entry under the cursor, if any.
func (m Model) CurrentEntry() (fs.FileEntry, bool) {
	if len(m.Nodes) == 0 || m.Cursor >= len(m.Nodes) {
		return fs.FileEntry{}, false
	}
	return m.Nodes[m.Cursor].Entry, true
}

// CurrentNode returns the node under the cursor.
func (m Model) CurrentNode() (TreeNode, bool) {
	if len(m.Nodes) == 0 || m.Cursor >= len(m.Nodes) {
		return TreeNode{}, false
	}
	return m.Nodes[m.Cursor], true
}

// clampOffset adjusts scroll offset so cursor is visible.
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
	h := m.Height - 4 // border(2) + cwd header(1) + col header(1)
	if h < 1 {
		h = 1
	}
	return h
}

// EntryCount returns number of visible nodes.
func (m Model) EntryCount() int { return len(m.Nodes) }

// SelectedCount returns number of selected entries.
func (m Model) SelectedCount() int { return len(m.Selected) }

// GetCwd returns current working directory.
func (m Model) GetCwd() string { return m.Cwd }

// EntryNames returns names of top-level entries (excludes "..", excludes expanded children).
func (m Model) EntryNames() []string {
	names := make([]string, 0, len(m.Nodes))
	for _, n := range m.Nodes {
		if n.Depth == 0 && n.Entry.Name != ".." {
			names = append(names, n.Entry.Name)
		}
	}
	return names
}

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
