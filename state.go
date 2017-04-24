package main

import (
	"github.com/1egoman/slime/gateway"
)

// This struct contains the main application state. I have fluxy intentions.
type State struct {
	Mode string

	Command               []rune
	CommandCursorPosition int

	MessageHistory []gateway.Message

	// All the connections that are currently made to outside services.
	Connections []gateway.Connection
	activeConnection int
}

func (s *State) ActiveConnection() gateway.Connection {
	return s.Connections[s.activeConnection]
}
