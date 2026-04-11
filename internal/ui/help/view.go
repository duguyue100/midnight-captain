package help

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/theme"
)

// Each pane: inner content width. Both equal for symmetry.
const colW = 38

var (
	stylePaneLeft = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderHighlight).
			Background(theme.BGFloat).
			Padding(0, 1).
			Width(colW)

	stylePaneRight = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Purple).
			Background(theme.BGFloat).
			Padding(0, 1).
			Width(colW)

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

	styleDivLine = lipgloss.NewStyle().
			Foreground(theme.FGGutter).
			Background(theme.BGFloat)

	styleDim = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat)

	styleAboutTitle = lipgloss.NewStyle().
			Foreground(theme.Magenta).
			Background(theme.BGFloat).
			Bold(true)

	styleAboutLabel = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat)

	styleAboutValue = lipgloss.NewStyle().
			Foreground(theme.FG).
			Background(theme.BGFloat)

	styleAboutAccent = lipgloss.NewStyle().
				Foreground(theme.Blue).
				Background(theme.BGFloat).
				Bold(true)

	styleAboutDim = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat)

	styleAboutHi = lipgloss.NewStyle().
			Foreground(theme.Green).
			Background(theme.BGFloat).
			Bold(true)

	styleHint = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat).
			Italic(true)
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
	{"h", "collapse dir or parent"},
	{"l / enter", "expand dir"},
	{"o", "navigate into dir"},
	{".", "toggle hidden files"},
	{"V", "visual select mode"},
	{"esc", "cancel / clear"},
}

var actionKeys = []entry{
	{"space", "fuzzy search"},
	{":", "command palette"},
	{"?", "this help"},
	{"e", "open in nvim"},
	{"a", "smart create"},
	{"r", "rename"},
	{"y", "yank (copy)"},
	{"d", "cut"},
	{"p", "paste"},
	{"x", "delete"},
	{"q", "quit"},
}

var commands = []entry{
	{":sort name|size|date", "sort entries"},
	{":hidden", "toggle hidden"},
	{":goto <path>", "jump to path"},
	{":mkdir <name>", "create directory"},
	{":touch <name>", "create file"},
	{":find", "recursive search"},
	{":ssh user@host", "connect SSH"},
	{":disconnect", "disconnect SSH"},
	{":quit", "exit"},
}

// View renders the help overlay (empty string if not visible).
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	left := strings.TrimRight(buildLeft(), "\n")
	right := strings.TrimRight(buildRight(m.Version), "\n")

	// Render once without height to measure actual rendered line counts.
	rendL := stylePaneLeft.Render(left)
	rendR := stylePaneRight.Render(right)
	lh := strings.Count(rendL, "\n") + 1
	rh := strings.Count(rendR, "\n") + 1
	outerH := lh
	if rh > outerH {
		outerH = rh
	}

	// Re-render both with the same outer height so borders align.
	row := lipgloss.JoinHorizontal(lipgloss.Top,
		stylePaneLeft.Height(outerH).Render(left),
		stylePaneRight.Height(outerH).Render(right),
	)
	return row
}

// buildLeft returns the keybindings pane content.
func buildLeft() string {
	const keyW = 16
	innerW := colW - 4 // border(2) + padding(1,1)

	var sb strings.Builder

	addSection := func(title string, entries []entry) {
		sb.WriteString(styleSection.Width(innerW).Render(title))
		sb.WriteByte('\n')
		sb.WriteString(styleDivLine.Width(innerW).Render(strings.Repeat("─", innerW)))
		sb.WriteByte('\n')
		for _, e := range entries {
			k := styleKey.Render(fmt.Sprintf("%-*s", keyW, e.key))
			d := styleDesc.Render(e.desc)
			sb.WriteString(k + " " + d)
			sb.WriteByte('\n')
		}
	}

	addSection("Navigation", navKeys)
	sb.WriteByte('\n')
	addSection("Actions", actionKeys)
	sb.WriteByte('\n')
	addSection("Commands", commands)
	sb.WriteByte('\n')
	sb.WriteString(styleHint.Width(innerW).Render("press ? or esc to close"))

	return sb.String()
}

// buildRight returns the about pane content.
func buildRight(version string) string {
	innerW := colW - 4 // border(2) + padding(1,1)

	v := version
	if v == "" {
		v = "dev"
	}

	var sb strings.Builder

	line := func(s string, style lipgloss.Style) {
		sb.WriteString(style.Width(innerW).Render(s))
		sb.WriteByte('\n')
	}
	blank := func() { sb.WriteByte('\n') }
	kv := func(label, value string, valStyle lipgloss.Style) {
		l := styleAboutLabel.Render(fmt.Sprintf("%-9s", label))
		val := valStyle.Render(value)
		sb.WriteString(l + " " + val)
		sb.WriteByte('\n')
	}

	line("  midnight-captain", styleAboutTitle)
	line("  a terminal file manager", styleAboutDim)
	blank()
	line(strings.Repeat("─", innerW), styleDivLine)
	blank()
	kv("version", "v"+v, styleAboutAccent)
	kv("license", "MIT", styleAboutValue)
	blank()
	line(strings.Repeat("─", innerW), styleDivLine)
	blank()
	line("crafted with", styleAboutDim)
	line("  duguyue100  +  OpenCode", styleAboutHi)
	blank()
	line("github.com/duguyue100/", styleAboutLabel)
	line("  midnight-captain", styleAboutAccent)
	blank()
	line(strings.Repeat("─", innerW), styleDivLine)
	blank()
	line("dual panes  ·  vim keys", styleAboutDim)
	line("tokyonight  ·  nerd fonts", styleAboutDim)
	line("ssh  ·  fuzzy search", styleAboutDim)
	blank()
	line(strings.Repeat("─", innerW), styleDivLine)
	blank()
	line("  built for navigating", styleAboutDim)
	line("  the dark at midnight", styleAboutDim)

	return sb.String()
}
