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
			Background(theme.BGVisual).
			Foreground(theme.Cyan).
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
	if m.Err != nil {
		errStr := truncateCols(" Error: "+m.Err.Error(), innerW)
		sb.WriteString(lipgloss.NewStyle().Foreground(theme.Red).Width(innerW).Render(errStr))
		sb.WriteByte('\n')
		rendered++
	} else {
		for i := m.Offset; i < end; i++ {
			sb.WriteString(m.renderRow(i, innerW))
			sb.WriteByte('\n')
			rendered++
		}
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

	rightCols := colSizeW + colKindW + colDateW

	// Icon slot: nerdIcon returns glyph+space = 2 terminal cols.
	// Dirs (non-..) prepend ▸/▼ indicator (1 col) → total 3 cols before name.
	// Others: just icon (2 cols) before name.
	var iconStr string
	var iconSlotW int
	if entry.IsDir && entry.Name != ".." {
		indicator := "▸"
		icon := iconFolder
		if node.Expanded {
			indicator = "▼"
			icon = iconFolderOpen
		}
		iconStr = indicator + icon // 1 + 2 = 3 cols
		iconSlotW = 3
	} else {
		iconStr = nerdIcon(entry) // 2 cols (glyph+space)
		iconSlotW = 2
	}

	nameW := width - indentW - iconSlotW - rightCols
	if nameW < 4 {
		nameW = 4
	}

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

	row := indent + iconStr + namePadded + sizeCol + kindCol + dateCol
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

// Nerd font icons as explicit Unicode codepoints (NF v3 / cod range).
// All are private-use area glyphs rendered 2 terminal columns wide by nerd-patched fonts.
const (
	iconParent     = "\uf07c " // nf-fa-folder_open  ..
	iconLink       = "\uf0c1 " // nf-fa-link
	iconFolder     = "\uf07b " // nf-fa-folder (closed)
	iconFolderOpen = "\uf07c " // nf-fa-folder_open (expanded)
	iconGo         = "\ue627 " // nf-dev-go
	iconRust       = "\ue7a8 " // nf-dev-rust
	iconPython     = "\ue606 " // nf-dev-python
	iconJS         = "\ue74e " // nf-dev-javascript
	iconTS         = "\ue628 " // nf-dev-typescript
	iconReact      = "\ue7ba " // nf-dev-react
	iconHTML       = "\ue736 " // nf-dev-html5
	iconCSS        = "\ue749 " // nf-dev-css3
	iconJSON       = "\ue60b " // nf-dev-javascript (reuse)
	iconYAML       = "\ue6d5 " // nf-dev-code (yaml)
	iconTOML       = "\ue6b2 " // nf-custom-toml
	iconMarkdown   = "\ue73e " // nf-dev-markdown
	iconText       = "\uf15c " // nf-fa-file_text
	iconShell      = "\uf489 " // nf-dev-terminal
	iconVim        = "\ue62b " // nf-dev-vim
	iconC          = "\ue61e " // nf-dev-c
	iconCpp        = "\ue61d " // nf-dev-cplusplus
	iconJava       = "\ue738 " // nf-dev-java
	iconRuby       = "\ue739 " // nf-dev-ruby
	iconPHP        = "\ue73d " // nf-dev-php
	iconSwift      = "\ue755 " // nf-dev-swift
	iconKotlin     = "\ue70e " // nf-dev-kotlin
	iconPDF        = "\uf1c1 " // nf-fa-file_pdf
	iconImage      = "\uf1c5 " // nf-fa-file_image
	iconVideo      = "\uf03d " // nf-fa-film
	iconAudio      = "\uf001 " // nf-fa-music
	iconArchive    = "\uf410 " // nf-oct-package
	iconPackage    = "\uf187 " // nf-fa-archive
	iconLock       = "\uf023 " // nf-fa-lock
	iconConfig     = "\ue615 " // nf-dev-aptana (config)
	iconGit        = "\ue702 " // nf-dev-git
	iconDocker     = "\ue7b0 " // nf-dev-docker
	iconMakefile   = "\uf0ad " // nf-fa-wrench
	iconSQL        = "\uf1c0 " // nf-fa-database
	iconDB         = "\uf1c0 " // nf-fa-database
	iconKey        = "\uf084 " // nf-fa-key
	iconExec       = "\uf489 " // nf-dev-terminal
	iconFile       = "\uf15b " // nf-fa-file
)

// nerdIcon returns a nerd font icon string (glyph + space = 2 terminal cols) for the entry.
func nerdIcon(e fs.FileEntry) string {
	if e.Name == ".." {
		return iconParent
	}
	if e.IsLink {
		return iconLink
	}
	if e.IsDir {
		return iconFolder
	}
	// File type by extension
	ext := fileExt(e.Name)
	switch ext {
	case "go":
		return iconGo
	case "rs":
		return iconRust
	case "py":
		return iconPython
	case "js", "mjs", "cjs":
		return iconJS
	case "ts":
		return iconTS
	case "tsx", "jsx":
		return iconReact
	case "html", "htm":
		return iconHTML
	case "css", "scss", "sass":
		return iconCSS
	case "json":
		return iconJSON
	case "yaml", "yml":
		return iconYAML
	case "toml":
		return iconTOML
	case "md", "markdown":
		return iconMarkdown
	case "txt":
		return iconText
	case "sh", "bash", "zsh", "fish":
		return iconShell
	case "vim", "lua":
		return iconVim
	case "c", "h":
		return iconC
	case "cpp", "cc", "cxx", "hpp":
		return iconCpp
	case "java":
		return iconJava
	case "rb":
		return iconRuby
	case "php":
		return iconPHP
	case "swift":
		return iconSwift
	case "kt":
		return iconKotlin
	case "pdf":
		return iconPDF
	case "png", "jpg", "jpeg", "gif", "svg", "webp", "ico":
		return iconImage
	case "mp4", "mkv", "avi", "mov", "webm":
		return iconVideo
	case "mp3", "flac", "wav", "ogg", "m4a":
		return iconAudio
	case "zip", "tar", "gz", "bz2", "xz", "7z", "rar":
		return iconArchive
	case "deb", "rpm", "pkg", "dmg", "exe", "msi":
		return iconPackage
	case "lock":
		return iconLock
	case "env", "cfg", "conf", "ini":
		return iconConfig
	case "git":
		return iconGit
	case "dockerfile":
		return iconDocker
	case "makefile":
		return iconMakefile
	case "sql":
		return iconSQL
	case "db", "sqlite", "sqlite3":
		return iconDB
	case "key", "pem", "crt", "cer":
		return iconKey
	}
	// Executable (no extension but +x)
	if e.Mode&0o111 != 0 {
		return iconExec
	}
	return iconFile
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
