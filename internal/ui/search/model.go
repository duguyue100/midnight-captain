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
	input   textinput.Model
	Visible bool
	results []SearchResult
	cursor  int
	loading bool
	Width   int
	Height  int
	baseDir string // directory being searched
}

// New creates a search model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "fuzzy search files…"
	ti.CharLimit = 128
	return Model{input: ti}
}

// Open shows the search overlay rooted at dir.
func (m *Model) Open(dir string) tea.Cmd {
	m.Visible = true
	m.baseDir = dir
	m.results = nil
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
