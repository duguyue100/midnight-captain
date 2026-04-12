package search

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbletea/v2"
)

// SearchResult holds a single fuzzy match.
type SearchResult struct {
	Path  string
	Score int
}

// ResultsMsg is sent from the background walk goroutine.
type ResultsMsg struct {
	Files   []string // full file list from walk (only set when Done=true)
	Results []SearchResult
	Done    bool
}

// NavigateMsg is sent when user selects a result.
type NavigateMsg struct {
	Dir  string
	Name string
}

// Model holds search overlay state.
type Model struct {
	input     textinput.Model
	Visible   bool
	recursive bool // true = recursive walk mode
	results   []SearchResult
	allFiles  []string // owned by Model, never accessed from goroutines
	cursor    int
	loading   bool
	Width     int
	Height    int
	baseDir   string // directory being searched
}

// New creates a search model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 128
	return Model{input: ti}
}

// OpenLocal shows the search overlay with current dir entries only (instant, no walk).
func (m *Model) OpenLocal(dir string, names []string) tea.Cmd {
	m.Visible = true
	m.recursive = false
	m.baseDir = dir
	m.cursor = 0
	m.loading = false
	m.input.Reset()
	m.allFiles = names
	m.results = m.filter("")
	return m.input.Focus()
}

// OpenRecursive shows the search overlay with a full recursive walk.
func (m *Model) OpenRecursive(dir string) tea.Cmd {
	m.Visible = true
	m.recursive = true
	m.baseDir = dir
	m.results = nil
	m.allFiles = nil
	m.cursor = 0
	m.loading = true
	m.input.Reset()
	return tea.Batch(m.input.Focus(), startWalk(dir))
}

// Close hides the search overlay.
func (m *Model) Close() {
	m.Visible = false
	m.input.Blur()
}

// SetSize updates dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}
