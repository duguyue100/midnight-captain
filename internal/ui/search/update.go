package search

import (
	"path/filepath"

	"charm.land/bubbletea/v2"
)

// Update handles messages for the search overlay.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case ResultsMsg:
		m.loading = false
		if msg.Done {
			m.allFiles = msg.Files
		}
		m.results = m.filter(m.input.Value())
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.Close()
			return m, nil

		case "enter":
			if len(m.results) > 0 {
				rel := m.results[m.cursor].Path
				abs := filepath.Join(m.baseDir, rel)
				dir := filepath.Dir(abs)
				name := filepath.Base(abs)
				m.Close()
				return m, func() tea.Msg {
					return NavigateMsg{Dir: dir, Name: name}
				}
			}

		case "ctrl+j", "down":
			m.cursor++
			if m.cursor >= len(m.results) {
				m.cursor = len(m.results) - 1
			}
			return m, nil

		case "ctrl+k", "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 0
			}
			return m, nil
		}
	}

	// Forward to textinput
	var cmd tea.Cmd
	prevVal := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != prevVal {
		m.results = m.filter(m.input.Value())
		m.cursor = 0
	}
	return m, cmd
}
