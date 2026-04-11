package pane

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/dgyhome/midnight-captain/internal/fs"
	"github.com/dgyhome/midnight-captain/internal/theme"
	"github.com/mattn/go-runewidth"
)

var (
	styleActive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderHighlight)

	styleInactive = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.FGGutter)

	// Cursor: high-contrast bright bg + dark fg + bold
	styleCursor = lipgloss.NewStyle().
			Background(theme.Blue).
			Foreground(theme.BG).
			Bold(true)

	// Cursor when also selected
	styleCursorSel = lipgloss.NewStyle().
			Background(theme.Cyan).
			Foreground(theme.BG).
			Bold(true)

	styleSelected = lipgloss.NewStyle().
			Background(theme.BGVisual).
			Foreground(theme.FG).
			Bold(true)

	styleHeader = lipgloss.NewStyle().
			Foreground(theme.Blue).
			Bold(true)

	styleNormal = lipgloss.NewStyle().Foreground(theme.FG)
	styleDir    = lipgloss.NewStyle().Foreground(theme.DirColor).Bold(true)
	styleExec   = lipgloss.NewStyle().Foreground(theme.ExecColor)
	styleLink   = lipgloss.NewStyle().Foreground(theme.LinkColor)
	styleHide   = lipgloss.NewStyle().Foreground(theme.HiddenColor)
	styleParent = lipgloss.NewStyle().Foreground(theme.Comment).Bold(true)

	styleMeta = lipgloss.NewStyle().Foreground(theme.Comment)
)

// Column widths (fixed right-side columns)
const (
	colSizeW = 7  // "  1.2M "
	colKindW = 10 // " document "
	colDateW = 7  // "Jan 02 " / "2006   "
)

// View renders the pane as a string.
func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	innerW := m.Width - 2
	if innerW < 8 {
		innerW = 8
	}

	visibleRows := m.visibleHeight()

	var sb strings.Builder

	// Header: cwd path
	cwd := truncate(m.Cwd, innerW)
	header := styleHeader.Width(innerW).Render(cwd)
	sb.WriteString(header)
	sb.WriteByte('\n')

	// Column header row
	sb.WriteString(renderColHeader(innerW))
	sb.WriteByte('\n')

	// File list rows
	end := m.Offset + visibleRows
	if end > len(m.Nodes) {
		end = len(m.Nodes)
	}

	rendered := 0
	for i := m.Offset; i < end; i++ {
		sb.WriteString(m.renderRow(i, innerW))
		sb.WriteByte('\n')
		rendered++
	}
	// Pad empty rows
	emptyRow := strings.Repeat(" ", innerW)
	for rendered < visibleRows {
		sb.WriteString(emptyRow)
		sb.WriteByte('\n')
		rendered++
	}

	content := strings.TrimRight(sb.String(), "\n")

	boxStyle := styleInactive
	if m.Focused {
		boxStyle = styleActive
	}

	return boxStyle.Width(m.Width).Render(content)
}

func renderColHeader(width int) string {
	rightCols := colSizeW + colKindW + colDateW
	nameW := width - 3 - rightCols // 3 = icon(2 cols)+space(1)
	if nameW < 4 {
		nameW = 4
	}
	name := padRight("Name", nameW)
	size := padLeft("Size", colSizeW)
	kind := padLeft("Kind", colKindW)
	date := padLeft("Date", colDateW)
	row := "   " + name + size + kind + date
	row = colClamp(row, width)
	return styleMeta.Width(width).Render(row)
}

func (m Model) renderRow(idx, width int) string {
	node := m.Nodes[idx]
	entry := node.Entry
	isCursor := idx == m.Cursor
	isSel := m.Selected[idx]

	// Indent: 2 spaces per depth level
	indent := strings.Repeat("  ", node.Depth)
	indentW := runewidth.StringWidth(indent)

	// Expand/collapse indicator for dirs (replaces plain icon space)
	var expandIndicator string
	if entry.IsDir && entry.Name != ".." {
		if node.Expanded {
			expandIndicator = "▼"
		} else {
			expandIndicator = "▸"
		}
	}

	rightCols := colSizeW + colKindW + colDateW
	// nameW accounts for indent + icon(2) + space(1) + expand indicator(1)
	iconCols := 2 + 1 // icon width + space
	if expandIndicator != "" {
		iconCols += runewidth.StringWidth(expandIndicator)
	}
	nameW := width - indentW - iconCols - rightCols
	if nameW < 4 {
		nameW = 4
	}

	icon := nerdIcon(entry)
	name := entry.Name
	if entry.IsDir && name != ".." {
		name += "/"
	}
	name = truncateCols(name, nameW)
	namePadded := padRight(name, nameW)

	size := formatSize(entry)
	kind := formatKind(entry)
	date := formatDate(entry.ModTime)

	sizeCol := padLeft(size, colSizeW)
	kindCol := padLeft(kind, colKindW)
	dateCol := padLeft(date, colDateW)

	var row string
	if expandIndicator != "" {
		row = indent + expandIndicator + icon + " " + namePadded + sizeCol + kindCol + dateCol
	} else {
		row = indent + icon + " " + namePadded + sizeCol + kindCol + dateCol
	}
	row = colClamp(row, width)

	return colorRow(row, entry, isCursor, isSel)
}

func colorRow(row string, entry fs.FileEntry, isCursor, isSel bool) string {
	switch {
	case isCursor && isSel:
		return styleCursorSel.Render(row)
	case isCursor:
		return styleCursor.Render(row)
	case isSel:
		return styleSelected.Render(row)
	case entry.Name == "..":
		return styleParent.Render(row)
	case entry.IsLink:
		return styleLink.Render(row)
	case entry.IsDir:
		return styleDir.Render(row)
	case entry.Mode&0o111 != 0:
		return styleExec.Render(row)
	case strings.HasPrefix(entry.Name, "."):
		return styleHide.Render(row)
	default:
		return styleNormal.Render(row)
	}
}

// nerdIcon returns a 2-char wide nerd font icon for the entry.
// Falls back to ASCII if terminal doesn't support nerd fonts.
func nerdIcon(e fs.FileEntry) string {
	if e.Name == ".." {
		return " "
	}
	if e.IsLink {
		return " "
	}
	if e.IsDir {
		return " "
	}
	// File type by extension
	ext := fileExt(e.Name)
	switch ext {
	case "go":
		return " "
	case "rs":
		return " "
	case "py":
		return " "
	case "js", "mjs", "cjs":
		return " "
	case "ts", "tsx":
		return " "
	case "jsx":
		return " "
	case "html", "htm":
		return " "
	case "css", "scss", "sass":
		return " "
	case "json":
		return " "
	case "yaml", "yml":
		return " "
	case "toml":
		return " "
	case "md", "markdown":
		return " "
	case "txt":
		return " "
	case "sh", "bash", "zsh", "fish":
		return " "
	case "vim", "lua":
		return " "
	case "c", "h":
		return " "
	case "cpp", "cc", "cxx", "hpp":
		return " "
	case "java":
		return " "
	case "rb":
		return " "
	case "php":
		return " "
	case "swift":
		return " "
	case "kt":
		return " "
	case "pdf":
		return " "
	case "png", "jpg", "jpeg", "gif", "svg", "webp", "ico":
		return " "
	case "mp4", "mkv", "avi", "mov", "webm":
		return " "
	case "mp3", "flac", "wav", "ogg", "m4a":
		return " "
	case "zip", "tar", "gz", "bz2", "xz", "7z", "rar":
		return " "
	case "deb", "rpm", "pkg", "dmg", "exe", "msi":
		return " "
	case "lock":
		return " "
	case "env", "cfg", "conf", "ini":
		return " "
	case "git":
		return " "
	case "dockerfile":
		return " "
	case "makefile":
		return " "
	case "sql":
		return " "
	case "db", "sqlite", "sqlite3":
		return " "
	case "key", "pem", "crt", "cer":
		return " "
	}
	// Executable (no extension but +x)
	if e.Mode&0o111 != 0 {
		return " "
	}
	return " "
}

func fileExt(name string) string {
	// Special names
	lower := strings.ToLower(name)
	if lower == "dockerfile" || lower == "makefile" || lower == "gemfile" {
		return lower
	}
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			if i == 0 {
				return ""
			}
			return strings.ToLower(name[i+1:])
		}
	}
	return ""
}

func formatSize(e fs.FileEntry) string {
	if e.Name == ".." {
		return "      "
	}
	if e.IsDir {
		return "     -"
	}
	s := e.Size
	switch {
	case s >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(s)/(1<<30))
	case s >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(s)/(1<<20))
	case s >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(s)/(1<<10))
	default:
		return fmt.Sprintf("%dB", s)
	}
}

func formatKind(e fs.FileEntry) string {
	if e.Name == ".." {
		return "    parent"
	}
	if e.IsLink {
		return "   symlink"
	}
	if e.IsDir {
		return " directory"
	}
	ext := fileExt(e.Name)
	switch ext {
	case "go":
		return "        Go"
	case "rs":
		return "      Rust"
	case "py":
		return "    Python"
	case "js", "mjs", "cjs":
		return "        JS"
	case "ts":
		return "        TS"
	case "tsx", "jsx":
		return "     React"
	case "html", "htm":
		return "      HTML"
	case "css":
		return "       CSS"
	case "json":
		return "      JSON"
	case "yaml", "yml":
		return "      YAML"
	case "toml":
		return "      TOML"
	case "md", "markdown":
		return "  Markdown"
	case "sh", "bash", "zsh", "fish":
		return "     Shell"
	case "c", "h":
		return "         C"
	case "cpp", "cc", "cxx", "hpp":
		return "       C++"
	case "java":
		return "      Java"
	case "rb":
		return "      Ruby"
	case "swift":
		return "     Swift"
	case "pdf":
		return "       PDF"
	case "png", "jpg", "jpeg", "gif", "svg", "webp":
		return "     Image"
	case "mp4", "mkv", "avi", "mov", "webm":
		return "     Video"
	case "mp3", "flac", "wav", "ogg":
		return "     Audio"
	case "zip", "tar", "gz", "bz2", "xz", "7z", "rar":
		return "  Archive"
	case "sql":
		return "       SQL"
	case "db", "sqlite", "sqlite3":
		return "  Database"
	case "key", "pem", "crt":
		return "      Cert"
	case "env":
		return "       Env"
	case "txt":
		return "      Text"
	case "lock":
		return "    Lockfile"
	case "makefile":
		return "  Makefile"
	case "dockerfile":
		return "    Docker"
	}
	if e.Mode&0o111 != 0 {
		return "      Exec"
	}
	return "      File"
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "      "
	}
	now := time.Now()
	if t.Year() == now.Year() {
		return t.Format("Jan 02")
	}
	return t.Format("2006  ")
}

// colWidth returns the terminal display width of s (accounts for double-wide chars).
func colWidth(s string) int {
	return runewidth.StringWidth(s)
}

// padRight pads s to exactly w terminal columns (right-padding with spaces).
func padRight(s string, w int) string {
	cw := colWidth(s)
	if cw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-cw)
}

// padLeft pads s to exactly w terminal columns (left-padding with spaces).
func padLeft(s string, w int) string {
	cw := colWidth(s)
	if cw >= w {
		return s
	}
	return strings.Repeat(" ", w-cw) + s
}

// colClamp hard-clamps s to w terminal columns, padding or truncating as needed.
func colClamp(s string, w int) string {
	cw := colWidth(s)
	if cw == w {
		return s
	}
	if cw < w {
		return s + strings.Repeat(" ", w-cw)
	}
	// Truncate: walk runes until we hit w cols
	total := 0
	for i, r := range s {
		rw := runewidth.RuneWidth(r)
		if total+rw > w {
			return s[:i] + strings.Repeat(" ", w-total)
		}
		total += rw
	}
	return s
}

// truncateCols truncates s to at most w terminal columns, adding "…" if cut.
func truncateCols(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if colWidth(s) <= w {
		return s
	}
	if w <= 1 {
		return "…"
	}
	// Build truncated string fitting w-1 cols, then append "…"
	total := 0
	for i, r := range s {
		rw := runewidth.RuneWidth(r)
		if total+rw > w-1 {
			return s[:i] + "…"
		}
		total += rw
	}
	return s
}

// truncate is kept for non-column uses (cwd header, etc.)
func truncate(s string, max int) string {
	return truncateCols(s, max)
}
