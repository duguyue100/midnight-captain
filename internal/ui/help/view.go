package help

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/theme"
)

const boxW = 62

var (
	styleHelpBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderHighlight).
			Background(theme.BGFloat).
			Padding(0, 1)

	styleSection = lipgloss.NewStyle().
			Foreground(theme.Yellow).
			Background(theme.BGFloat).
			Bold(true)

	styleKey = lipgloss.NewStyle().
			Foreground(theme.Cyan).
			Background(theme.BGFloat)

	styleDesc = lipgloss.NewStyle().
			Foreground(theme.FGDark).
			Background(theme.BGFloat)

	styleDim = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat)
)

type entry struct {
	key  string
	desc string
}

var navKeys = []entry{
	{"j / k", "move down / up"},
	{"ctrl+d / ctrl+u", "half-page down / up"},
	{"g g", "jump to top"},
	{"G", "jump to bottom"},
	{"tab", "switch pane"},
	{"h", "collapse dir or go to parent"},
	{"l / enter", "expand dir"},
	{"o", "navigate into dir (change cwd)"},
	{".", "toggle hidden files"},
	{"V", "visual select mode"},
	{"esc", "cancel / clear selection"},
}

var actionKeys = []entry{
	{"space", "fuzzy search in current dir"},
	{":", "open command palette"},
	{"?", "show this help"},
	{"e", "open file in nvim"},
	{"a", "smart create file or dir"},
	{"r", "rename"},
	{"y", "yank (copy)"},
	{"d", "cut"},
	{"p", "paste"},
	{"x", "delete"},
	{"q", "quit"},
}

var commands = []entry{
	{":sort name|size|date", "sort entries"},
	{":hidden", "toggle hidden files"},
	{":goto <path>", "go to path (~ and / supported)"},
	{":mkdir <name>", "create directory"},
	{":touch <name>", "create empty file"},
	{":find", "recursive fuzzy search"},
	{":ssh user@host", "connect over SSH"},
	{":disconnect", "disconnect SSH pane"},
	{":quit", "exit midnight-captain"},
}

// View renders the help overlay (empty string if not visible).
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	innerW := boxW - 4 // 2 border + 2 padding
	keyW := 18

	var sb strings.Builder

	renderSection := func(title string, entries []entry) {
		sb.WriteString(styleSection.Width(innerW).Render(title))
		sb.WriteByte('\n')
		divider := strings.Repeat("─", innerW)
		sb.WriteString(styleDim.Width(innerW).Render(divider))
		sb.WriteByte('\n')
		for _, e := range entries {
			k := styleKey.Render(fmt.Sprintf("%-*s", keyW, e.key))
			d := styleDesc.Render(e.desc)
			sb.WriteString(k + " " + d)
			sb.WriteByte('\n')
		}
	}

	renderSection("Navigation", navKeys)
	sb.WriteString(styleDim.Width(innerW).Render(strings.Repeat(" ", innerW)))
	sb.WriteByte('\n')
	renderSection("Actions", actionKeys)
	sb.WriteString(styleDim.Width(innerW).Render(strings.Repeat(" ", innerW)))
	sb.WriteByte('\n')
	renderSection("Commands", commands)

	// trailing hint
	hint := styleDim.Render("  press ? or esc to close")
	sb.WriteString(hint)

	return styleHelpBox.Width(boxW).Render(sb.String())
}
