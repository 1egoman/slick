package main

import (
	"log"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
)

// Given a state object populated with a gateway, initialize the state with the gateway.
func connect(state *State, term *frontend.TerminalDisplay, connected chan struct{}) {
	// Connect to all gateways on start.
	for _, connection := range state.Connections {
		if err := connection.Connect(); err != nil {
			log.Fatal(err)
		}
	}

	// Render initial state.
	render(state, term)

	// We're connected!
	close(connected)
}

