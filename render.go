package main

import (
	"github.com/1egoman/slime/frontend"
)

// Given application state and a frontend, render the state to the screen.
// This function is called whenever something in state is changed.
func render(state State, term *frontend.TerminalDisplay) {
	term.DrawMessages(state.MessageHistory)

	term.DrawStatusBar(state.Mode)
	term.DrawCommandBar(
		string(state.Command),           // The command that the user is typing
		state.CommandCursorPosition,     // The cursor position
		state.Gateway.SelectedChannel(), // The selected channel
		state.Gateway.Team(),            // The selected team
	)

	if state.Mode == "picker" {
		term.DrawFuzzyPicker([]string{"abc", "def", "ghi"}, 1)
	}

	term.Render()
}

