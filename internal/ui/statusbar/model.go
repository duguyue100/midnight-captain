package statusbar

import "charm.land/bubbletea/v2"

// SetMessageMsg sets a transient status message.
type SetMessageMsg struct {
	Message string
}

// Model holds statusbar state.
type Model struct {
	Message string
	Width   int
	// Future: operations []ops.Operation
}

// New creates a statusbar model.
func New(width int) Model {
	return Model{Width: width}
}

// SetSize updates width.
func (m *Model) SetSize(w int) {
	m.Width = w
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetMessageMsg:
		m.Message = msg.Message
	}
	return m, nil
}
