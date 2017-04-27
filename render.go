package main

import (
	"log"
	"github.com/1egoman/slime/frontend"
	"github.com/1egoman/slime/gateway"
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

	// term.DrawMessages(state.ActiveConnection().MessageHistory())
	term.DrawMessages([]gateway.Message{
		gateway.Message{Sender: &gateway.User{Name: "Ryan"}, Text: "Foo!"},
		gateway.Message{Sender: &gateway.User{Name: "Ryan"}, Text: "ghjmkmnjbhgvsbhnjkmnjbhvgrhbjnkmgnjbhvgsjnhgfvakdvg adkvja gdiuvkadbgskbghskubjh kuvjhsu jbhvskufjxb hsiuj"},
	})

	term.DrawStatusBar(
		state.Mode, // Which mode we're currently in
		state.Connections, // A list of all connections
		state.ActiveConnection(), // Which conenction is active (to highlight the active one differently)
	)
	term.DrawCommandBar(
		string(state.Command),           // The command that the user is typing
		state.CommandCursorPosition,     // The cursor position
		state.ActiveConnection().SelectedChannel(), // The selected channel
		state.ActiveConnection().Team(),            // The selected team
	)

	if state.Mode == "picker" {
		term.DrawFuzzyPicker([]string{"abc", "def", "ghi"}, 1)
	}

	term.Render()
}

