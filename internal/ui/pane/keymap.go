package pane

// Key constants for pane navigation.
// Using string literals matching tea.KeyPressMsg.String() output.
const (
	keyDown         = "j"
	keyUp           = "k"
	keyRight        = "l" // expand/collapse dir
	keyLeft         = "h" // collapse or jump to parent node or goParent
	keyOpenDir      = "o" // navigate into dir (change Cwd)
	keyBackspace    = "backspace"
	keyEnter        = "enter" // expand/collapse dir (same as l)
	keyHalfDown     = "ctrl+f"
	keyHalfUp       = "ctrl+u"
	keyTop          = "g"
	keyBottom       = "G"
	keyToggleHidden = "."
	keyVisual       = "V"
	keyEsc          = "esc"
)
