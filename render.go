package main

import (
	"log"
	"sort"

	"github.com/1egoman/slick/frontend"
)

// Given application state and a frontend, render the state to the screen.
// This function is called whenever something in state is changed.
func render(state *State, term *frontend.TerminalDisplay) {
	// If the user switched connections, then refresh
	if state.ConnectionIsStale() && state.ActiveConnection() != nil {
		state.SyncActiveConnection()
		log.Printf("User switching to new active connection: %s", state.ActiveConnection().Name())

		// Emit event to to be handled by lua scripts
		EmitEvent(state, EVENT_CONNECTION_CHANGE, map[string]string{
			"name": state.ActiveConnection().Name(),
		})

		go func() {
			state.Status.Printf("Loading team %s...", state.ActiveConnection().Name())
			render(state, term)

			if err := state.ActiveConnection().Refresh(false); err != nil {
				state.Status.Errorf(err.Error())
			}

			state.Status.Clear()
			render(state, term)
		}()
	}

	// Render messages provided by the active conenction
	if state.ActiveConnection() != nil {
		state.RenderedMessageNumber, state.RenderedAllMessages = term.DrawMessages(
			state.ActiveConnection().MessageHistory(),                                   // List of messages
			len(state.ActiveConnection().MessageHistory())-1-state.SelectedMessageIndex, // Is a message selected?
			state.BottomDisplayedItem,                                                   // Bottommost item
			state.ActiveConnection().UserById,
			state.ActiveConnection().UserOnline,
			state.Configuration,
		)
	} else {
		term.DrawBlankLines(0, -1*frontend.BottomPadding)
		term.DrawInfoPage()
	}

	if state.ActiveConnection() == nil {
		term.DrawCommandBar(
			string(state.Command),       // The command that the user is typing
			state.CommandCursorPosition, // The cursor position
			nil,                  // The selected channel
			"(no active connec)", // The selected team name
			state.Configuration,
		)
	} else {
		term.DrawCommandBar(
			string(state.Command),                      // The command that the user is typing
			state.CommandCursorPosition,                // The cursor position
			state.ActiveConnection().SelectedChannel(), // The selected channel
			state.ActiveConnection().Name(),            // The selected team name
			state.Configuration,
		)
	}

	term.DrawStatusBar(
		state.Mode,               // Which mode we're currently in
		state.Connections,        // A list of all connections
		state.ActiveConnection(), // Which conenction is active (to highlight the active one differently)
		state.Status,             // Status message to display
		state.Configuration,
	)

	if state.FuzzyPicker.Visible {
		// Sort items by the search command
		state.FuzzyPicker.Needle = string(state.Command)
		sort.Sort(state.FuzzyPicker)
		if state.FuzzyPicker.OnResort != nil {
			state.FuzzyPicker.OnResort(state)
		}

		// Render all connections and channels
		// Check visibility again, because `OnResort` may have hid the fuzzy picker.
		if state.FuzzyPicker.Visible {
			term.DrawFuzzyPicker(
				state.FuzzyPicker.StringItems,
				state.FuzzyPicker.SelectedItem,
				state.FuzzyPicker.BottomItem,
				state.FuzzyPicker.Rank,
				state.Configuration,
			)
		}
	}

	term.Render()
}
