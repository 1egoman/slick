package main

import (
	"log"
	"sort"

	"github.com/1egoman/slime/frontend"
)

// Given application state and a frontend, render the state to the screen.
// This function is called whenever something in state is changed.
func render(state *State, term *frontend.TerminalDisplay) {
	// If the user switched connections, then refresh
	if state.ConnectionIsStale() {
		state.SyncActiveConnection()
		log.Printf("User switching to new active connection: %s", state.ActiveConnection().Name())

		go func() {
			if err := state.ActiveConnection().Refresh(); err != nil {
				log.Fatal(err)
			}
			render(state, term)
		}()
	}

	state.RenderedMessageNumber = term.DrawMessages(
		state.ActiveConnection().MessageHistory(),                                   // List of messages
		len(state.ActiveConnection().MessageHistory())-1-state.SelectedMessageIndex, // Is a message selected?
		state.BottomDisplayedItem, // Bottommost item
	)

	term.DrawStatusBar(
		state.Mode,               // Which mode we're currently in
		state.Connections,        // A list of all connections
		state.ActiveConnection(), // Which conenction is active (to highlight the active one differently)
	)
	term.DrawCommandBar(
		string(state.Command),                      // The command that the user is typing
		state.CommandCursorPosition,                // The cursor position
		state.ActiveConnection().SelectedChannel(), // The selected channel
		state.ActiveConnection().Team(),            // The selected team
	)

	if state.FuzzyPicker.Visible {
		// Sort items by the search command
		state.FuzzyPicker.Needle = string(state.Command)
		sort.Sort(state.FuzzyPicker)

		// Render all connections and channels
		term.DrawFuzzyPicker(
			state.FuzzyPicker.StringItems,
			state.FuzzyPicker.SelectedItem,
			state.FuzzyPicker.BottomItem,
		)
	}

	term.Render()
}
