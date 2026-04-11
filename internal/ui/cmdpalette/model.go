package cmdpalette

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbletea/v2"
)

// Command is a registered command in the palette.
type Command struct {
	Name        string
	Description string
	Action      func(args []string) tea.Cmd
}

// Model holds command palette state.
type Model struct {
	input      textinput.Model
	Visible    bool
	justOpened bool // drop first key event after open (prevents trigger char leaking into input)
	commands   []Command
	filtered   []Command
	Cursor     int
	Width      int
	Height     int
}

// New creates a command palette model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 256

	m := Model{
		input: ti,
	}
	m.commands = builtinCommands()
	m.filtered = m.commands
	return m
}

// Open shows the palette and focuses the input.
func (m *Model) Open() tea.Cmd {
	m.Visible = true
	m.justOpened = true
	m.Cursor = 0
	m.input.Reset()
	m.filtered = m.commands
	return m.input.Focus()
}

// Close hides the palette.
func (m *Model) Close() {
	m.Visible = false
	m.input.Blur()
}

// SetSize updates dimensions used for rendering.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}
