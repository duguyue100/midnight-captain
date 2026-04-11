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

	searchLoading = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat).
			Italic(true)
)

const searchW = 70

// View renders the search overlay or "".
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	innerW := searchW - 4 // border(2) + padding(2)
	var sb strings.Builder

	// Input line
	sb.WriteString(searchInput.Width(innerW).Render(m.input.View()))
	sb.WriteByte('\n')
	sb.WriteString(searchDivider.Width(innerW).Render(strings.Repeat("─", innerW)))

	if m.loading {
		sb.WriteByte('\n')
		sb.WriteString(searchLoading.Width(innerW).Render("  scanning…"))
	} else {
		limit := len(m.results)
		maxRows := m.Height/2 - 4
		if maxRows < 5 {
			maxRows = 5
		}
		if limit > maxRows {
			limit = maxRows
		}
		for i := 0; i < limit; i++ {
			r := m.results[i]
			line := fmt.Sprintf("  %s", r.Path)
			runes := []rune(line)
			if len(runes) > innerW {
				line = "  …" + string(runes[len(runes)-(innerW-3):])
			}
			sb.WriteByte('\n')
			if i == m.cursor {
				sb.WriteString(searchResultActive.Width(innerW).Render(line))
			} else {
				sb.WriteString(searchResult.Width(innerW).Render(line))
			}
		}
		if len(m.results) == 0 && !m.loading {
			sb.WriteByte('\n')
			sb.WriteString(searchLoading.Width(innerW).Render("  no results"))
		}
	}

	return searchBox.Width(searchW).Render(sb.String())
}
