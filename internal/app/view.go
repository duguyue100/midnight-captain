package app

import (
	"strings"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/dgyhome/midnight-captain/internal/theme"
	"github.com/mattn/go-runewidth"
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

	if m.CmdPalette.Visible {
		base = placeOverlay(base, m.CmdPalette.View(), m.Width, m.Height)
	}
	if m.Search.Visible {
		base = placeOverlay(base, m.Search.View(), m.Width, m.Height)
	}
	if m.Confirm.Visible {
		base = placeOverlay(base, m.Confirm.View(), m.Width, m.Height)
	}
	if m.Input.Visible {
		base = placeOverlay(base, m.Input.View(), m.Width, m.Height)
	}
	if m.Help.Visible {
		base = placeOverlay(base, m.Help.View(), m.Width, m.Height)
	}
	if m.Goto.Visible {
		base = placeOverlay(base, m.Goto.View(), m.Width, m.Height)
	}

	return base
}

func (m Model) renderBase() string {
	dual := m.Width >= 80
	sbHeight := 1
	paneHeight := m.Height - sbHeight

	var sb strings.Builder

	if dual {
		divider := lipgloss.NewStyle().
			Foreground(theme.FGGutter).
			Render("│")

		leftView := m.Left.View()
		rightView := m.Right.View()

		leftLines := strings.Split(leftView, "\n")
		rightLines := strings.Split(rightView, "\n")

		leftLines = clampLines(leftLines, paneHeight)
		rightLines = clampLines(rightLines, paneHeight)

		for i := 0; i < paneHeight; i++ {
			l := lineAt(leftLines, i, m.Left.Width)
			r := lineAt(rightLines, i, m.Right.Width)
			sb.WriteString(l + divider + r + "\n")
		}
	} else {
		ap := m.activePane()
		lines := strings.Split(ap.View(), "\n")
		lines = clampLines(lines, paneHeight)
		for i := 0; i < paneHeight; i++ {
			sb.WriteString(lineAt(lines, i, m.Width))
			sb.WriteByte('\n')
		}
	}

	sb.WriteString(m.Statusbar.View(m.activePane()))
	// no trailing newline — content is exactly m.Height lines

	return sb.String()
}

// placeOverlay merges overlay into base by splicing each overlay line into the
// corresponding base line using ANSI-aware Cut/Truncate so base colors are preserved.
func placeOverlay(base, overlay string, w, h int) string {
	overlayLines := strings.Split(overlay, "\n")
	baseLines := strings.Split(base, "\n")

	// Measure overlay visual width (widest line, stripped of ANSI)
	overlayH := len(overlayLines)
	overlayW := 0
	for _, l := range overlayLines {
		ww := runewidth.StringWidth(ansi.Strip(l))
		if ww > overlayW {
			overlayW = ww
		}
	}

	// Center position (0-indexed into baseLines)
	startRow := (h - overlayH) / 2
	startCol := (w - overlayW) / 2
	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	for i, ol := range overlayLines {
		row := startRow + i
		if row >= len(baseLines) {
			break
		}
		bl := baseLines[row]
		// Splice: left portion of base | overlay line | right portion of base
		left := ansi.Truncate(bl, startCol, "")
		right := ansi.TruncateLeft(bl, startCol+overlayW, "")
		baseLines[row] = left + ol + right
	}

	return strings.Join(baseLines, "\n")
}

func clampLines(lines []string, max int) []string {
	if len(lines) <= max {
		return lines
	}
	return lines[:max]
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
