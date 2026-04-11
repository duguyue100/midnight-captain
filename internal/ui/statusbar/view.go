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

	styleAccent = lipgloss.NewStyle().
			Background(theme.BGDark).
			Foreground(theme.Blue)

	styleSep = lipgloss.NewStyle().
			Background(theme.BGDark).
			Foreground(theme.FGGutter)

	styleHint = lipgloss.NewStyle().
			Background(theme.BGDark).
			Foreground(theme.Comment)

	styleMsg = lipgloss.NewStyle().
			Background(theme.BGDark).
			Foreground(theme.Yellow)

	styleSpinner = lipgloss.NewStyle().
			Background(theme.BGDark).
			Foreground(theme.Cyan)
)

const sep = " | "

// PaneInfo abstracts pane data needed by statusbar.
type PaneInfo interface {
	EntryCount() int
	SelectedCount() int
	GetCwd() string
}

// View renders the status bar (1 line).
// Layout: [count] [| spinner pct%] [| message]      [tab | ? | q]
func (m Model) View(left, right PaneInfo) string {
	if m.Width == 0 {
		return ""
	}

	// --- Left section: item count + optional selection ---
	count := left.EntryCount()
	sel := left.SelectedCount()

	countStr := fmt.Sprintf(" %d items", count)
	if sel > 0 {
		countStr += fmt.Sprintf(" (%d selected)", sel)
	}
	left_ := styleAccent.Render(countStr)

	// --- Middle section: spinner+pct OR message ---
	middle := ""
	if m.Active {
		spin := styleSpinner.Render(m.SpinnerFrame())
		pct := styleBar.Render(fmt.Sprintf(" %d%%", m.Pct))
		middle = styleSep.Render(sep) + spin + pct
	} else if m.Message != "" {
		middle = styleSep.Render(sep) + styleMsg.Render(m.Message)
	}

	// --- Right section: key hints ---
	right_ := styleHint.Render("tab | ? | q ")

	// Measure plain widths for gap calculation
	leftW := len([]rune(countStr))
	middleW := 0
	if m.Active {
		middleW = len([]rune(sep)) + len([]rune(m.SpinnerFrame())) + len([]rune(fmt.Sprintf(" %d%%", m.Pct)))
	} else if m.Message != "" {
		middleW = len([]rune(sep)) + len([]rune(m.Message))
	}
	rightW := len([]rune("tab | ? | q "))

	gap := m.Width - leftW - middleW - rightW
	if gap < 1 {
		gap = 1
	}

	line := left_ + middle + styleBar.Render(strings.Repeat(" ", gap)) + right_
	return line
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
