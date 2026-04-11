package pane

// Key constants for pane navigation.
// Using string literals matching tea.KeyPressMsg.String() output.
const (
	keyDown         = "j"
	keyUp           = "k"
	keyRight        = "l"
	keyLeft         = "h"
	keyBackspace    = "backspace"
	keyEnter        = "enter"
	keyHalfDown     = "ctrl+d"
	keyHalfUp       = "ctrl+u"
	keyTop          = "g" // handled as double-g in app layer; pane uses single for now
	keyBottom       = "G"
	keyToggleHidden = "."
	keyVisual       = "V"
	keySpace        = " "
	keyEsc          = "esc"
)
