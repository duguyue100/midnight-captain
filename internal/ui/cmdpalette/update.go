package cmdpalette

import (
	"strings"

	"charm.land/bubbletea/v2"
)

// Update handles messages for the command palette.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.Close()
			return m, nil

		case "enter":
			return m.execute()

		case "ctrl+j", "down":
			m.Cursor++
			if m.Cursor >= len(m.filtered) {
				m.Cursor = len(m.filtered) - 1
			}
			return m, nil

		case "ctrl+k", "up":
			m.Cursor--
			if m.Cursor < 0 {
				m.Cursor = 0
			}
			return m, nil

		case "tab":
			// Autocomplete: fill input with selected command name
			if len(m.filtered) > 0 {
				name := m.filtered[m.Cursor].Name
				m.input.SetValue(name + " ")
				// move cursor to end
				m.refilter()
			}
			return m, nil
		}
	}

	// Forward to textinput
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.refilter()
	return m, cmd
}

func (m *Model) refilter() {
	val := strings.TrimSpace(m.input.Value())
	if val == "" {
		m.filtered = m.commands
		m.Cursor = 0
		return
	}

	// Extract command name (first token)
	parts := strings.Fields(val)
	prefix := parts[0]

	out := make([]Command, 0, len(m.commands))
	for _, c := range m.commands {
		if strings.HasPrefix(c.Name, prefix) {
			out = append(out, c)
		}
	}
	m.filtered = out
	if m.Cursor >= len(m.filtered) {
		m.Cursor = 0
	}
}

func (m Model) execute() (Model, tea.Cmd) {
	val := strings.TrimSpace(m.input.Value())
	if val == "" {
		m.Close()
		return m, nil
	}

	// If only one match and user hits enter, use it
	parts := strings.Fields(val)
	cmdName := parts[0]
	args := parts[1:]

	// Find exact or single prefix match
	var match *Command
	if len(m.filtered) == 1 {
		match = &m.filtered[0]
	} else {
		for i := range m.filtered {
			if m.filtered[i].Name == cmdName {
				match = &m.filtered[i]
				break
			}
		}
		// Use cursor selection if no exact match
		if match == nil && len(m.filtered) > 0 {
			match = &m.filtered[m.Cursor]
			// Merge args from cursor-selected command
			parts2 := strings.Fields(val)
			if len(parts2) > 1 {
				args = parts2[1:]
			} else {
				args = nil
			}
		}
	}

	m.Close()
	if match != nil {
		return m, match.Action(args)
	}
	return m, nil
}
