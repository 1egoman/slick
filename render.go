package main

import (
	"log"
	"fmt"
	"sort"

	"github.com/1egoman/slime/frontend"
)

// Given application state and a frontend, render the state to the screen.
// This function is called whenever something in state is changed.
func render(state *State, term *frontend.TerminalDisplay) {
	// If the user switched connections, then refresh
	if state.ConnectionIsStale() {
		state.SyncActiveConnection()
		log.Printf("User swiching to new active connection: %s", state.ActiveConnection().Name())

		go func() {
			if err := state.ActiveConnection().Refresh(); err != nil {
				log.Fatal(err)
			}
			render(state, term)
		}()
	}

	term.DrawMessages(state.ActiveConnection().MessageHistory())

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

	if state.Mode == "pick" {
		items := []FuzzyPickerReference{}
		stringItems := []string{}

		// Accumulate all channels into `items`, and their respective labels into `stringLabels`
		for _, connection := range state.Connections {
			for _, channel := range connection.Channels() {
				// Add string representation of item to `stringItems`
				// Follows the pattern of "my-team #my-channel"
				stringItems = append(stringItems, fmt.Sprintf(
					"%s #%s",
					connection.Name(),
					channel.Name,
				))

				// Add backing representation of item to `item`
				items = append(items, FuzzyPickerReference{
					Channel: &channel,
					Connection: &connection,
				})
			}
		}

		// Fuzzy sort the items
		sorter := FuzzySorter{
			Items: items,
			StringItems: stringItems,
			Needle: string(state.Command),
		}
		sort.Sort(sorter)

		// Render all connections and channels
		term.DrawFuzzyPicker(sorter.StringItems, state.fuzzyPickerSelectedItem)
	}

	term.Render()
}
