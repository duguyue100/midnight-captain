package statusbar

import (
	"time"

	"charm.land/bubbletea/v2"
)

// spinner frames for the rotating bar during operations.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// TickMsg drives the spinner.
type TickMsg struct{}

// SetMessageMsg sets a transient status message.
type SetMessageMsg struct {
	Message string
}

// Model holds statusbar state.
type Model struct {
	Message string
	Width   int

	// progress state
	Active      bool // true while an op is running
	Pct         int  // 0–100
	spinFrame   int
	CurrentFile string // currently processing file
}

// New creates a statusbar model.
func New(width int) Model {
	return Model{Width: width}
}

// SetSize updates width.
func (m *Model) SetSize(w int) {
	m.Width = w
}

// SetProgress updates the active operation progress (pct 0–100, active=true).
// Call with active=false to clear.
func (m *Model) SetProgress(active bool, pct int, file string) {
	m.Active = active
	m.Pct = pct
	m.CurrentFile = file
}

// tickCmd sends a TickMsg after 100ms to animate the spinner.
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
		return TickMsg{}
	})
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetMessageMsg:
		m.Message = msg.Message
	case TickMsg:
		if m.Active {
			m.spinFrame = (m.spinFrame + 1) % len(spinnerFrames)
			return m, tickCmd()
		}
	}
	return m, nil
}

// StartSpinner begins spinner animation.
func (m *Model) StartSpinner() tea.Cmd {
	m.Active = true
	m.spinFrame = 0
	return tickCmd()
}

// StopSpinner halts the spinner.
func (m *Model) StopSpinner() {
	m.Active = false
	m.Pct = 0
	m.CurrentFile = ""
}

// SpinnerFrame returns the current spinner glyph.
func (m Model) SpinnerFrame() string {
	return spinnerFrames[m.spinFrame]
}
