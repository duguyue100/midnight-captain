package goto_

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/theme"
)

const (
	gotoInnerW = 54
	gotoFieldW = gotoInnerW + 4 // border(2) + padding(2)
	gotoBoxW   = gotoFieldW + 6 // box border(2) + padding(2×2)
)

var (
	styleGotoBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderHighlight).
			Background(theme.BGFloat).
			Padding(1, 2)

	styleGotoTitle = lipgloss.NewStyle().
			Foreground(theme.Blue).
			Bold(true).
			Background(theme.BGFloat)

	styleGotoField = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.FGGutter).
			Background(theme.BGFloat).
			Padding(0, 1)

	styleGotoDivider = lipgloss.NewStyle().
				Foreground(theme.FGGutter).
				Background(theme.BGFloat)

	styleGotoDir = lipgloss.NewStyle().
			Foreground(theme.Blue).
			Background(theme.BGFloat)

	styleGotoFile = lipgloss.NewStyle().
			Foreground(theme.FGDark).
			Background(theme.BGFloat)

	styleGotoSelected = lipgloss.NewStyle().
				Foreground(theme.FG).
				Background(theme.BGHighlight).
				Bold(true)

	styleGotoErr = lipgloss.NewStyle().
			Foreground(theme.Red).
			Background(theme.BGFloat)

	styleGotoDim = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat)

	styleGotoHint = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat)
)

const (
	dirIcon  = " " // nerd font folder
	fileIcon = " " // nerd font file
)

// View renders the goto overlay, or "" if not visible.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	var sb strings.Builder

	// Title
	sb.WriteString(styleGotoTitle.Width(gotoFieldW).Render("Go to"))
	sb.WriteByte('\n')
	sb.WriteByte('\n')

	// Input field
	sb.WriteString(styleGotoField.Width(gotoFieldW).Render(m.input.View()))
	sb.WriteByte('\n')

	// Divider with listing dir label
	label := ""
	if m.listDir != "" {
		home := ""
		if m.FS != nil {
			home = m.FS.Home()
		}
		label = contractTilde(m.listDir, home)
	}
	divLine := fmt.Sprintf("── %s ", label)
	if len([]rune(divLine)) < gotoFieldW {
		divLine += strings.Repeat("─", gotoFieldW-len([]rune(divLine)))
	}
	sb.WriteString(styleGotoDivider.Width(gotoFieldW).Render(divLine))
	sb.WriteByte('\n')

	// Entries list
	raw := strings.TrimSpace(m.input.Value())
	switch {
	case m.notFound:
		sb.WriteString(styleGotoErr.Width(gotoFieldW).Render("  path does not exist"))
		sb.WriteByte('\n')

	case raw == "" && len(m.entries) == 0:
		sb.WriteString(styleGotoDim.Width(gotoFieldW).Render("  type a path  ·  ~ or / to start"))
		sb.WriteByte('\n')

	case len(m.entries) == 0:
		sb.WriteString(styleGotoErr.Width(gotoFieldW).Render("  no matches"))
		sb.WriteByte('\n')

	default:
		limit := len(m.entries)
		if limit > maxListRows {
			limit = maxListRows
		}
		for i := 0; i < limit; i++ {
			e := m.entries[i]
			icon := fileIcon
			baseStyle := styleGotoFile
			if e.isDir {
				icon = dirIcon
				baseStyle = styleGotoDir
			}
			name := e.name
			if e.isDir {
				name += "/"
			}
			line := fmt.Sprintf(" %s %-*s", icon, gotoFieldW-5, name)
			if i == m.cursor {
				sb.WriteString(styleGotoSelected.Width(gotoFieldW).Render(line))
			} else {
				sb.WriteString(baseStyle.Width(gotoFieldW).Render(line))
			}
			sb.WriteByte('\n')
		}
		if len(m.entries) > maxListRows {
			more := fmt.Sprintf("  … %d more", len(m.entries)-maxListRows)
			sb.WriteString(styleGotoDim.Width(gotoFieldW).Render(more))
			sb.WriteByte('\n')
		}
	}

	// Footer hint
	sb.WriteString(styleGotoHint.Width(gotoFieldW).Render("  tab/↑↓ select  ·  enter confirm  ·  esc cancel"))

	return styleGotoBox.Width(gotoBoxW).Render(sb.String())
}
