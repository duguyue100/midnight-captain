package cmdpalette

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/theme"
)

var (
	styleBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderHighlight).
			Background(theme.BGFloat).
			Padding(0, 1)

	styleInput = lipgloss.NewStyle().
			Foreground(theme.FG).
			Background(theme.BGFloat)

	styleSuggestion = lipgloss.NewStyle().
			Foreground(theme.FGDark).
			Background(theme.BGFloat)

	styleSuggestionActive = lipgloss.NewStyle().
				Foreground(theme.FG).
				Background(theme.BGHighlight).
				Bold(true)

	styleDesc = lipgloss.NewStyle().
			Foreground(theme.Comment).
			Background(theme.BGFloat)
)

const (
	paletteW   = 60
	maxSuggest = 10
)

// View renders the floating palette overlay, or "" if not visible.
func (m Model) View() string {
	if !m.Visible {
		return ""
	}

	innerW := paletteW - 4 // 2 border + 2 padding

	var sb strings.Builder

	// Input line
	inputView := m.input.View()
	sb.WriteString(styleInput.Width(innerW).Render(inputView))

	// Divider
	sb.WriteByte('\n')
	sb.WriteString(styleDesc.Width(innerW).Render(strings.Repeat("─", innerW)))

	// Suggestions
	limit := len(m.filtered)
	if limit > maxSuggest {
		limit = maxSuggest
	}

	// Always render maxSuggest lines to keep height fixed
	for i := 0; i < maxSuggest; i++ {
		sb.WriteByte('\n')
		if i < limit {
			cmd := m.filtered[i]
			line := fmt.Sprintf("%-12s  %s", cmd.Name, cmd.Description)
			if len([]rune(line)) > innerW {
				line = string([]rune(line)[:innerW])
			}
			if i == m.Cursor {
				sb.WriteString(styleSuggestionActive.Width(innerW).Render(line))
			} else {
				sb.WriteString(styleSuggestion.Width(innerW).Render(line))
			}
		} else if i == 0 && limit == 0 {
			sb.WriteString(styleDesc.Width(innerW).Render("  no matching commands"))
		} else {
			sb.WriteString(styleSuggestion.Width(innerW).Render(""))
		}
	}

	box := styleBox.Width(paletteW).Render(sb.String())
	return box
}
