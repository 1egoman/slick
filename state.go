package main

import (
	"github.com/1egoman/slime/gateway"
)

// This struct contains the main application state. I have fluxy intentions.
type State struct {
	Mode string

	Command               []rune
	CommandCursorPosition int

	// All the connections that are currently made to outside services.
	Connections []gateway.Connection
	activeConnection int
	connectionSynced bool

	IsConnected chan bool
}

func (s *State) ActiveConnection() gateway.Connection {
	return s.Connections[s.activeConnection]
}


// Methods to manage the active connection
// When the user changes the active connection, 
func (s *State) SetActiveConnection(index int) {
	s.activeConnection = index
	s.connectionSynced = false
}
func (s *State) SetNextActiveConnection() {
	s.activeConnection += 1
	s.connectionSynced = false

	// Make sure connectino can never get larger than the amount of conenctions
	if s.activeConnection > len(s.Connections) - 1 {
		s.activeConnection = len(s.Connections) - 1
	}
}
func (s *State) SetPrevActiveConnection() {
	s.activeConnection -= 1
	s.connectionSynced = false

	// Make sure connectino can never get below 0
	if s.activeConnection < 0 {
		s.activeConnection = 0
	}
}
func (s *State) ConnectionIsStale() bool {
	return !s.connectionSynced
}
func (s *State) SyncActiveConnection() {
	s.connectionSynced = true
}
