package app

import (
	"strings"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/theme"
)

var styleRoot = lipgloss.NewStyle().Background(theme.BG)

// View renders the full terminal UI.
func (m Model) View() tea.View {
	content := m.renderContent()
	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m Model) renderContent() string {
	if m.Width == 0 {
		return "Loading..."
	}

	base := m.renderBase()

	// Overlay palette if visible
	if m.CmdPalette.Visible {
		overlay := m.CmdPalette.View()
		base = placeOverlay(base, overlay, m.Width, m.Height)
	}

	// Overlay search if visible
	if m.Search.Visible {
		overlay := m.Search.View()
		base = placeOverlay(base, overlay, m.Width, m.Height)
	}

	// Overlay confirm dialog if visible
	if m.Confirm.Visible {
		overlay := m.Confirm.View()
		base = placeOverlay(base, overlay, m.Width, m.Height)
	}

	// Overlay input dialog if visible
	if m.Input.Visible {
		overlay := m.Input.View()
		base = placeOverlay(base, overlay, m.Width, m.Height)
	}

	return base
}

func (m Model) renderBase() string {
	dual := m.Width >= 80

	var sb strings.Builder

	if dual {
		divider := lipgloss.NewStyle().
			Foreground(theme.FGGutter).
			Render("│")

		leftView := m.Left.View()
		rightView := m.Right.View()

		leftLines := strings.Split(leftView, "\n")
		rightLines := strings.Split(rightView, "\n")

		maxLines := max(len(leftLines), len(rightLines))
		for i := 0; i < maxLines; i++ {
			l := lineAt(leftLines, i, m.Left.Width)
			r := lineAt(rightLines, i, m.Right.Width)
			sb.WriteString(l + divider + r + "\n")
		}
	} else {
		ap := m.activePane()
		sb.WriteString(ap.View())
		sb.WriteByte('\n')
	}

	sb.WriteString(m.Statusbar.View(&m.Left, &m.Right))

	return sb.String()
}

// placeOverlay centers the overlay string on top of the base string.
// Uses lipgloss.Place to position, then overlays line-by-line.
func placeOverlay(base, overlay string, w, h int) string {
	// Use lipgloss Place to get the overlay positioned on a blank canvas
	placed := lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, overlay)

	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(placed, "\n")

	out := make([]string, max(len(baseLines), len(overlayLines)))
	for i := range out {
		bl := ""
		if i < len(baseLines) {
			bl = baseLines[i]
		}
		ol := ""
		if i < len(overlayLines) {
			ol = overlayLines[i]
		}
		// Overlay line wins if it contains non-space content
		if strings.TrimSpace(ol) != "" {
			out[i] = ol
		} else {
			out[i] = bl
		}
	}
	return strings.Join(out, "\n")
}

func lineAt(lines []string, i, width int) string {
	if i < len(lines) {
		return lines[i]
	}
	return strings.Repeat(" ", width)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
