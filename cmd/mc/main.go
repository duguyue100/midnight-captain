package main

import (
	"fmt"
	"os"

	"charm.land/bubbletea/v2"
	"github.com/dgyhome/midnight-captain/internal/app"
)

var version = "dev"

func main() {
	m := app.NewModel(version)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
