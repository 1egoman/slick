package main

import (
	// "log"
	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
)

// Given a state object populated with a gateway, initialize the state with the gateway.
func connect(state *State, term *frontend.TerminalDisplay, connected chan struct{}) error {
	// Connect to all gateways on start.
	for _, connection := range state.Connections {
		if err := connection.Connect(); err != nil {
			return err
		}
	}

	// Render initial state.
	render(state, term)

	// We're connected!
	close(connected)

	// Preload a list of channels for each connection.
	// WIthout this, a user can't fuzzy pick another connection's channel, since they won't be
	// loaded yet :O
	for _, connection := range state.Connections {
		_, err := connection.FetchChannels()
		if err != nil {
			return err
		}
	}

	// Render any changes due to the added connections
	render(state, term)
	return nil
}
