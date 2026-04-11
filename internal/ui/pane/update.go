package pane

import (
	"charm.land/bubbletea/v2"
)

// RefreshMsg tells the pane to reload its directory.
type RefreshMsg struct{}

// Update handles bubbletea messages for the pane.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case RefreshMsg:
		m.Reload()
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case keyMatches(msg, keyDown):
		m.Cursor++
		if m.Cursor >= len(m.Entries) {
			m.Cursor = max(0, len(m.Entries)-1)
		}
		m.clampOffset()
		if m.VisualMode {
			m.extendVisualSelection()
		}

	case keyMatches(msg, keyUp):
		m.Cursor--
		if m.Cursor < 0 {
			m.Cursor = 0
		}
		m.clampOffset()
		if m.VisualMode {
			m.extendVisualSelection()
		}

	case keyMatches(msg, keyHalfDown):
		m.Cursor += m.visibleHeight() / 2
		if m.Cursor >= len(m.Entries) {
			m.Cursor = max(0, len(m.Entries)-1)
		}
		m.clampOffset()
		if m.VisualMode {
			m.extendVisualSelection()
		}

	case keyMatches(msg, keyHalfUp):
		m.Cursor -= m.visibleHeight() / 2
		if m.Cursor < 0 {
			m.Cursor = 0
		}
		m.clampOffset()
		if m.VisualMode {
			m.extendVisualSelection()
		}

	case keyMatches(msg, keyTop):
		m.Cursor = 0
		m.Offset = 0
		if m.VisualMode {
			m.extendVisualSelection()
		}

	case keyMatches(msg, keyBottom):
		m.Cursor = max(0, len(m.Entries)-1)
		m.clampOffset()
		if m.VisualMode {
			m.extendVisualSelection()
		}

	case keyMatches(msg, keyEnter), keyMatches(msg, keyRight):
		return m.enterCurrent()

	case keyMatches(msg, keyLeft), keyMatches(msg, keyBackspace):
		return m.goParent()

	case keyMatches(msg, keyToggleHidden):
		m.ShowHidden = !m.ShowHidden
		m.Reload()

	case keyMatches(msg, keyVisual):
		m.VisualMode = !m.VisualMode
		if m.VisualMode {
			m.VisualAnchor = m.Cursor
			m.Selected[m.Cursor] = true
		}

	case keyMatches(msg, keySpace):
		m.Selected[m.Cursor] = !m.Selected[m.Cursor]
		m.Cursor++
		if m.Cursor >= len(m.Entries) {
			m.Cursor = max(0, len(m.Entries)-1)
		}
		m.clampOffset()

	case keyMatches(msg, keyEsc):
		m.VisualMode = false
		m.Selected = make(map[int]bool)
	}
	return m, nil
}

// extendVisualSelection selects all entries between VisualAnchor and Cursor.
func (m *Model) extendVisualSelection() {
	m.Selected = make(map[int]bool)
	lo, hi := m.VisualAnchor, m.Cursor
	if lo > hi {
		lo, hi = hi, lo
	}
	for i := lo; i <= hi; i++ {
		m.Selected[i] = true
	}
}

func (m Model) enterCurrent() (Model, tea.Cmd) {
	entry, ok := m.CurrentEntry()
	if !ok {
		return m, nil
	}
	if entry.Name == ".." {
		return m.goParent()
	}
	if entry.IsDir {
		m.Cwd = m.Cwd + "/" + entry.Name
		m.Cursor = 0
		m.Offset = 0
		m.Selected = make(map[int]bool)
		m.Reload()
	}
	return m, nil
}

func (m Model) goParent() (Model, tea.Cmd) {
	if m.Cwd == "/" {
		return m, nil
	}
	lastName := baseName(m.Cwd)
	m.Cwd = parentDir(m.Cwd)
	m.Cursor = 0
	m.Offset = 0
	m.Selected = make(map[int]bool)
	m.Reload()
	for i, e := range m.Entries {
		if e.Name == lastName {
			m.Cursor = i
			m.clampOffset()
			break
		}
	}
	return m, nil
}

// NavigateTo navigates the pane to dir and positions cursor on the named file.
func (m *Model) NavigateTo(dir, name string) {
	m.Cwd = dir
	m.Cursor = 0
	m.Offset = 0
	m.Selected = make(map[int]bool)
	m.Reload()
	for i, e := range m.Entries {
		if e.Name == name {
			m.Cursor = i
			m.clampOffset()
			break
		}
	}
}

func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return "/"
}

func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func keyMatches(msg tea.KeyPressMsg, key string) bool {
	return msg.String() == key
}
