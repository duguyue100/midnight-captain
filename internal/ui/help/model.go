package help

// Model holds visibility state for the help overlay.
type Model struct {
	Visible bool
	Width   int
	Height  int
	Version string
}

// New returns a closed help overlay.
func New() Model {
	return Model{}
}

// Open shows the overlay.
func (m *Model) Open() {
	m.Visible = true
}

// Close hides the overlay.
func (m *Model) Close() {
	m.Visible = false
}

// SetSize propagates terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}
