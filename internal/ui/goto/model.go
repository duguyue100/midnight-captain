package goto_

import (
	"os"
	"os/user"
	"path/filepath"
	appfs "github.com/dgyhome/midnight-captain/internal/fs"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbletea/v2"
)

// NavigateMsg is emitted when the user confirms a destination.
type NavigateMsg struct {
	Dir string
}

// entry is a directory listing item.
type entry struct {
	name  string
	isDir bool
}

// Model is the goto floating prompt with live directory preview.
type Model struct {
	Visible bool
	input   textinput.Model

	// listing state
	listDir  string  // directory currently listed
	entries  []entry // filtered entries matching current typed prefix
	cursor   int     // selected entry index (-1 = none)
	notFound bool    // typed path component does not exist

	// PaneCwd is the active pane's working directory, used as base for
	// relative paths instead of the process cwd.
	PaneCwd string
	FS      appfs.FileSystem

	Width  int
	Height int
}

const maxListRows = 10

// New returns a closed goto model.
func New() Model {
	ti := textinput.New()
	ti.CharLimit = 512
	ti.SetWidth(54)
	m := Model{input: ti, cursor: -1}
	return m
}

// Open shows the prompt, optionally seeding it with an initial path.
// paneCwd is the active pane's working directory used as base for relative paths.
func (m *Model) Open(seed string, paneCwd string, fsys appfs.FileSystem) tea.Cmd {
	m.FS = fsys
	m.Visible = true
	m.cursor = -1
	m.PaneCwd = paneCwd
	m.input.Reset()
	if seed != "" {
		m.input.SetValue(seed)
		// move cursor to end
		m.input.CursorEnd()
	}
	m.refresh()
	return m.input.Focus()
}

// Close hides the prompt.
func (m *Model) Close() {
	m.Visible = false
	m.input.Blur()
}

// SetSize propagates terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// Update handles keys for the goto prompt.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.Close()
			return m, nil

		case "enter":
			return m.confirm()

		case "tab":
			m.cycleNext()
			return m, nil

		case "shift+tab":
			m.cyclePrev()
			return m, nil

		case "down", "ctrl+n":
			m.cycleNext()
			return m, nil

		case "up", "ctrl+p":
			m.cyclePrev()
			return m, nil
		}
	}

	prevVal := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != prevVal {
		m.cursor = -1
		m.refresh()
	}
	return m, cmd
}

// confirm navigates to the selected or typed path.
func (m Model) confirm() (Model, tea.Cmd) {
	// If a list entry is selected, complete into it
	if m.cursor >= 0 && m.cursor < len(m.entries) {
		e := m.entries[m.cursor]
		base := m.listDir
		target := filepath.Join(base, e.name)
		if e.isDir {
			// navigate into dir
			m.Close()
			return m, func() tea.Msg { return NavigateMsg{Dir: target} }
		}
		// file → navigate to parent
		m.Close()
		return m, func() tea.Msg { return NavigateMsg{Dir: base} }
	}
	// No selection — resolve whatever is typed
	dir := m.resolveDir(m.input.Value())
	if dir == "" {
		m.notFound = true
		return m, nil
	}
	m.Close()
	return m, func() tea.Msg { return NavigateMsg{Dir: dir} }
}

// cycleNext moves cursor forward through entries.
func (m *Model) cycleNext() {
	if len(m.entries) == 0 {
		return
	}
	m.cursor++
	if m.cursor >= len(m.entries) {
		m.cursor = 0
	}
	m.applySelection()
}

// cyclePrev moves cursor backward through entries.
func (m *Model) cyclePrev() {
	if len(m.entries) == 0 {
		return
	}
	if m.cursor <= 0 {
		m.cursor = len(m.entries) - 1
	} else {
		m.cursor--
	}
	m.applySelection()
}

// applySelection sets the input value to listDir + "/" + selected entry.
func (m *Model) applySelection() {
	if m.cursor < 0 || m.cursor >= len(m.entries) {
		return
	}
	e := m.entries[m.cursor]
	joined := filepath.Join(m.listDir, e.name)
	if e.isDir {
		joined += "/"
	}
	// Shorten home dir back to ~
	joined = contractTilde(joined)
	m.input.SetValue(joined)
	m.input.CursorEnd()
}

// refresh recomputes listDir and entries from the current input value.
func (m *Model) refresh() {
	m.notFound = false
	raw := m.input.Value()
	expanded := expandTilde(raw)

	if !filepath.IsAbs(expanded) {
		// Use pane's Cwd as base for relative paths, not process cwd
		base := m.PaneCwd
		if base == "" {
			base, _ = os.Getwd()
		}
		expanded = filepath.Join(base, expanded)
	}

	// Determine which directory to list and which prefix to filter by.
	// If remote, skip autocomplete to avoid blocking UI thread.
	if m.FS != nil && !m.FS.IsLocal() {
		m.listDir = ""
		m.entries = nil
		m.notFound = false
		return
	}

	// If raw ends with '/', list that dir with no filter.
	// Otherwise, list parent dir and filter by the last component.
	var listDir, prefix string
	info, err := os.Stat(expanded)
	if err == nil && info.IsDir() {
		// exact dir match (or ends with /)
		listDir = filepath.Clean(expanded)
		prefix = ""
	} else {
		listDir = filepath.Dir(expanded)
		prefix = strings.ToLower(filepath.Base(expanded))
		// Check parent exists
		if _, err := os.Stat(listDir); err != nil {
			m.listDir = ""
			m.entries = nil
			m.notFound = true
			return
		}
	}

	m.listDir = listDir
	m.entries = listEntries(listDir, prefix)
}

// listEntries reads a directory and returns entries matching the prefix.
func listEntries(dir, prefix string) []entry {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]entry, 0, len(des))
	for _, de := range des {
		name := de.Name()
		if strings.HasPrefix(strings.ToLower(name), prefix) {
			out = append(out, entry{name: name, isDir: de.IsDir()})
		}
	}
	// Sort: dirs first, then alpha
	sort.Slice(out, func(i, j int) bool {
		if out[i].isDir != out[j].isDir {
			return out[i].isDir
		}
		return strings.ToLower(out[i].name) < strings.ToLower(out[j].name)
	})
	return out
}

// expandTilde replaces a leading ~ with the home directory.
func expandTilde(s string) string {
	if s == "~" || strings.HasPrefix(s, "~/") {
		u, err := user.Current()
		if err != nil {
			home := os.Getenv("HOME")
			if home == "" {
				return s
			}
			return home + s[1:]
		}
		return u.HomeDir + s[1:]
	}
	return s
}

// contractTilde replaces the home directory prefix with ~.
func contractTilde(s string) string {
	u, err := user.Current()
	if err != nil {
		return s
	}
	home := u.HomeDir
	if strings.HasPrefix(s, home+"/") {
		return "~/" + s[len(home)+1:]
	}
	if s == home {
		return "~"
	}
	return s
}

// resolveDir takes raw input and returns the target directory, or "".
func (m *Model) resolveDir(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	expanded := expandTilde(raw)
	if !filepath.IsAbs(expanded) {
		base := m.PaneCwd
		if base == "" {
			base, _ = os.Getwd()
		}
		expanded = filepath.Join(base, expanded)
	}
	expanded = filepath.Clean(expanded)
	
	if m.FS != nil && !m.FS.IsLocal() {
		return expanded
	}

	info, err := os.Stat(expanded)
	if err != nil {
		return ""
	}
	if info.IsDir() {
		return expanded
	}
	return filepath.Dir(expanded)
}
