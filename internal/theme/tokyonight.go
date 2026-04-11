package theme

import "charm.land/lipgloss/v2"

var (
	// Backgrounds
	BG          = lipgloss.Color("#1a1b26")
	BGDark      = lipgloss.Color("#16161e")
	BGHighlight = lipgloss.Color("#292e42")
	BGFloat     = lipgloss.Color("#16161e")
	BGVisual    = lipgloss.Color("#283457")

	// Foregrounds
	FG       = lipgloss.Color("#c0caf5")
	FGDark   = lipgloss.Color("#a9b1d6")
	FGGutter = lipgloss.Color("#3b4261")
	Comment  = lipgloss.Color("#565f89")

	// Accents
	Blue    = lipgloss.Color("#7aa2f7")
	Cyan    = lipgloss.Color("#7dcfff")
	Green   = lipgloss.Color("#9ece6a")
	Orange  = lipgloss.Color("#ff9e64")
	Purple  = lipgloss.Color("#9d7cd8")
	Red     = lipgloss.Color("#f7768e")
	Magenta = lipgloss.Color("#bb9af7")
	Yellow  = lipgloss.Color("#e0af68")
	Teal    = lipgloss.Color("#1abc9c")

	// UI
	Border          = lipgloss.Color("#15161e")
	BorderHighlight = lipgloss.Color("#27a1b9")
	Error           = lipgloss.Color("#db4b4b")
	Warning         = lipgloss.Color("#e0af68")
	Info            = lipgloss.Color("#0db9d7")

	// Semantic file colors
	DirColor    = Blue
	ExecColor   = Green
	LinkColor   = Cyan
	HiddenColor = Comment
)
