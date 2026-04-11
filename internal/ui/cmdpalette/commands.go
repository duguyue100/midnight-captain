package cmdpalette

import "charm.land/bubbletea/v2"

// ExecuteMsg is sent when a command is executed.
type ExecuteMsg struct {
	Name string
	Args []string
}

func builtinCommands() []Command {
	return []Command{
		{
			Name:        "sort",
			Description: "Sort by: name | size | date",
			Action: func(args []string) tea.Cmd {
				arg := ""
				if len(args) > 0 {
					arg = args[0]
				}
				return func() tea.Msg {
					return ExecuteMsg{Name: "sort", Args: []string{arg}}
				}
			},
		},
		{
			Name:        "hidden",
			Description: "Toggle hidden files",
			Action: func(args []string) tea.Cmd {
				return func() tea.Msg {
					return ExecuteMsg{Name: "hidden"}
				}
			},
		},
		{
			Name:        "mkdir",
			Description: "Create directory: mkdir <name>",
			Action: func(args []string) tea.Cmd {
				arg := ""
				if len(args) > 0 {
					arg = args[0]
				}
				return func() tea.Msg {
					return ExecuteMsg{Name: "mkdir", Args: []string{arg}}
				}
			},
		},
		{
			Name:        "touch",
			Description: "Create empty file: touch <name>",
			Action: func(args []string) tea.Cmd {
				arg := ""
				if len(args) > 0 {
					arg = args[0]
				}
				return func() tea.Msg {
					return ExecuteMsg{Name: "touch", Args: []string{arg}}
				}
			},
		},
		{
			Name:        "ssh",
			Description: "Connect SSH: ssh user@host",
			Action: func(args []string) tea.Cmd {
				arg := ""
				if len(args) > 0 {
					arg = args[0]
				}
				return func() tea.Msg {
					return ExecuteMsg{Name: "ssh", Args: []string{arg}}
				}
			},
		},
		{
			Name:        "disconnect",
			Description: "Disconnect SSH on active pane",
			Action: func(args []string) tea.Cmd {
				return func() tea.Msg {
					return ExecuteMsg{Name: "disconnect"}
				}
			},
		},
		{
			Name:        "goto",
			Description: "Go to path: goto <path>",
			Action: func(args []string) tea.Cmd {
				arg := ""
				if len(args) > 0 {
					arg = args[0]
				}
				return func() tea.Msg {
					return ExecuteMsg{Name: "goto", Args: []string{arg}}
				}
			},
		},
		{
			Name:        "find",
			Description: "Recursive fuzzy search from current dir",
			Action: func(args []string) tea.Cmd {
				return func() tea.Msg {
					return ExecuteMsg{Name: "find"}
				}
			},
		},
		{
			Name:        "quit",
			Description: "Exit midnight-captain",
			Action: func(args []string) tea.Cmd {
				return tea.Quit
			},
		},
	}
}
