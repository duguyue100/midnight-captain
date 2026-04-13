package dialog

import (
	"strings"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/theme"
)

// ConfirmResultMsg is sent when user answers the confirm dialog.
type ConfirmResultMsg struct {
	ID        string
	Confirmed bool
}

// ConfirmModel is a yes/no confirmation dialog.
type ConfirmModel struct {
	ID      string
	Title   string
	Items   []string // files listed in the body
	Visible bool
	choice  bool // true = Yes
}

// NewConfirm creates a confirm dialog.
func NewConfirm(id, title string, items []string) ConfirmModel {
	return ConfirmModel{ID: id, Title: title, Items: items, choice: false}
}

// Open shows the dialog.
func (m *ConfirmModel) Open() { m.Visible = true; m.choice = false }

// Close hides the dialog.
func (m *ConfirmModel) Close() { m.Visible = false }

// Update handles key input.
func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "y", "Y":
			m.Close()
			return m, func() tea.Msg { return ConfirmResultMsg{ID: m.ID, Confirmed: true} }
		case "n", "N", "esc", "q":
			m.Close()
			return m, func() tea.Msg { return ConfirmResultMsg{ID: m.ID, Confirmed: false} }
		case "enter":
			confirmed := m.choice
			m.Close()
			return m, func() tea.Msg { return ConfirmResultMsg{ID: m.ID, Confirmed: confirmed} }
		case "h", "left":
			m.choice = true
		case "l", "right":
			m.choice = false
		case "tab", "shift+tab":
			m.choice = !m.choice
		}
	}
	return m, nil
}

var (
	confirmBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderHighlight).
			Background(theme.BGFloat).
			Padding(1, 2)

	confirmTitle = lipgloss.NewStyle().
			Foreground(theme.Yellow).
			Bold(true).
			Background(theme.BGFloat)

	confirmItem = lipgloss.NewStyle().
			Foreground(theme.FGDark).
			Background(theme.BGFloat)

	confirmBtnActive = lipgloss.NewStyle().
				Background(theme.Blue).
				Foreground(theme.BG).
				Bold(true).
				Padding(0, 2)

	confirmBtnInactive = lipgloss.NewStyle().
				Background(theme.BGHighlight).
				Foreground(theme.FGDark).
				Padding(0, 2)
)

// View renders the confirm dialog or "".
func (m ConfirmModel) View() string {
	if !m.Visible {
		return ""
	}

	const innerW = 40
	var sb strings.Builder

	sb.WriteString(confirmTitle.Width(innerW).Render(m.Title))
	sb.WriteByte('\n')
	sb.WriteByte('\n')

	limit := len(m.Items)
	if limit > 8 {
		limit = 8
	}
	for _, item := range m.Items[:limit] {
		sb.WriteString(confirmItem.Width(innerW).Render("  " + item))
		sb.WriteByte('\n')
	}
	if len(m.Items) > 8 {
		sb.WriteString(confirmItem.Width(innerW).Render(
			"  … and more",
		))
		sb.WriteByte('\n')
	}

	sb.WriteByte('\n')

	// Buttons: Yes / No
	var yes, no string
	if m.choice {
		yes = confirmBtnActive.Render("Yes")
		no = confirmBtnInactive.Render("No")
	} else {
		yes = confirmBtnInactive.Render("Yes")
		no = confirmBtnActive.Render("No")
	}
	btns := lipgloss.NewStyle().Background(theme.BGFloat).Width(innerW).
		Align(lipgloss.Center).Render(yes + "   " + no)
	sb.WriteString(btns)

	return confirmBox.Width(innerW + 4).Render(sb.String())
}
