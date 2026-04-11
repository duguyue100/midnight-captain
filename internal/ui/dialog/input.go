package dialog

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/theme"
)

// InputResultMsg is sent when user submits or cancels the input dialog.
type InputResultMsg struct {
	ID        string
	Value     string
	Cancelled bool
}

// InputModel is a single-line text input dialog (for rename, mkdir, touch).
type InputModel struct {
	ID      string
	Title   string
	input   textinput.Model
	Visible bool
}

// NewInput creates a text input dialog.
func NewInput(id, title, initial string) InputModel {
	ti := textinput.New()
	ti.CharLimit = 255
	ti.SetValue(initial)
	ti.SetWidth(52)
	return InputModel{ID: id, Title: title, input: ti}
}

// Open shows the dialog and focuses the input.
func (m *InputModel) Open() tea.Cmd {
	m.Visible = true
	return m.input.Focus()
}

// Close hides the dialog.
func (m *InputModel) Close() {
	m.Visible = false
	m.input.Blur()
}

// Update handles input for the dialog.
func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			val := strings.TrimSpace(m.input.Value())
			m.Close()
			return m, func() tea.Msg { return InputResultMsg{ID: m.ID, Value: val} }
		case "esc":
			m.Close()
			return m, func() tea.Msg { return InputResultMsg{ID: m.ID, Cancelled: true} }
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

var (
	inputBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderHighlight).
			Background(theme.BGFloat).
			Padding(1, 2)

	inputTitle = lipgloss.NewStyle().
			Foreground(theme.Blue).
			Bold(true).
			Background(theme.BGFloat)

	inputField = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.FGGutter).
			Background(theme.BGFloat).
			Padding(0, 1)
)

// View renders the input dialog or "".
func (m InputModel) View() string {
	if !m.Visible {
		return ""
	}

	const (
		inputW = 52         // matches ti.SetWidth
		fieldW = inputW + 4 // field border(2) + padding(2)
		boxW   = fieldW + 6 // box border(2) + padding(2×2)
	)
	var sb strings.Builder

	sb.WriteString(inputTitle.Width(fieldW).Render(m.Title))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(inputField.Width(fieldW).Render(m.input.View()))

	return inputBox.Width(boxW).Render(sb.String())
}
