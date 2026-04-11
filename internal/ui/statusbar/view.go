package statusbar

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/theme"
)

var (
	styleBar = lipgloss.NewStyle().
			Background(theme.BGDark).
			Foreground(theme.FGDark)

	styleHint = lipgloss.NewStyle().
			Background(theme.BGDark).
			Foreground(theme.Comment)
)

// PaneInfo abstracts pane data needed by statusbar.
type PaneInfo interface {
	EntryCount() int
	SelectedCount() int
	GetCwd() string
}

// View renders the status bar (2 lines).
func (m Model) View(left, right PaneInfo) string {
	if m.Width == 0 {
		return ""
	}

	// Line 1: info
	sel := left.SelectedCount()
	selStr := ""
	if sel > 0 {
		selStr = fmt.Sprintf(" · %d selected", sel)
	}
	leftInfo := fmt.Sprintf(" %d items%s", left.EntryCount(), selStr)

	hint := " Tab · : · q "
	if m.Message != "" {
		hint = " " + m.Message + " "
	}

	gap := m.Width - len(leftInfo) - len(hint)
	if gap < 0 {
		gap = 0
	}

	line1 := styleBar.Render(leftInfo) +
		styleBar.Render(strings.Repeat(" ", gap)) +
		styleHint.Render(hint)

	// Line 2: key hints
	hints := " j/k navigate · h/l expand · o enter dir · a create · Space search · : commands · q quit"
	hints = truncate(hints, m.Width)
	line2 := styleHint.Width(m.Width).Render(hints)

	return line1 + "\n" + line2
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}
