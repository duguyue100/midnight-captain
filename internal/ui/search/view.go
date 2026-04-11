package search

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/theme"
)

var (
	searchBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderHighlight).
			Background(theme.BGFloat).
			Padding(0, 1)

	searchInput = lipgloss.NewStyle().
			Foreground(theme.FG).
			Background(theme.BGFloat)

	searchDivider = lipgloss.NewStyle().
			Foreground(theme.FGGutter).
			Background(theme.BGFloat)

	searchResult = lipgloss.NewStyle().
			Foreground(theme.FGDark).
			Background(theme.BGFloat)

	searchResultActive = lipgloss.NewStyle().
				Foreground(theme.FG).
				Background(theme.BGHighlight).
				Bold(true)

	searchEmpty = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat).
			Italic(true)
)

const (
	searchW       = 70
	searchMaxRows = 12 // fixed result rows — never changes height
)

// View renders the search overlay with fixed dimensions.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	innerW := searchW - 4 // border(2) + padding(2)
	var sb strings.Builder

	// Input line with mode label
	label := "search  "
	if m.recursive {
		label = "find    "
	}
	inputLine := fmt.Sprintf("%s%s", label, m.input.View())
	sb.WriteString(searchInput.Width(innerW).Render(inputLine))
	sb.WriteByte('\n')
	sb.WriteString(searchDivider.Width(innerW).Render(strings.Repeat("─", innerW)))

	blank := searchEmpty.Width(innerW).Render("")

	if m.loading {
		sb.WriteByte('\n')
		sb.WriteString(searchEmpty.Width(innerW).Render("  scanning…"))
		// pad remaining rows
		for i := 1; i < searchMaxRows; i++ {
			sb.WriteByte('\n')
			sb.WriteString(blank)
		}
	} else {
		for i := 0; i < searchMaxRows; i++ {
			sb.WriteByte('\n')
			if i < len(m.results) {
				r := m.results[i]
				line := fmt.Sprintf("  %s", r.Path)
				runes := []rune(line)
				if len(runes) > innerW {
					line = "  …" + string(runes[len(runes)-(innerW-3):])
				}
				if i == m.cursor {
					sb.WriteString(searchResultActive.Width(innerW).Render(line))
				} else {
					sb.WriteString(searchResult.Width(innerW).Render(line))
				}
			} else if i == 0 && len(m.results) == 0 {
				sb.WriteString(searchEmpty.Width(innerW).Render("  no results"))
			} else {
				sb.WriteString(blank)
			}
		}
	}

	return searchBox.Width(searchW).Render(sb.String())
}
