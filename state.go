package main

import (
	"os"
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/1egoman/slime/gateway/slack"
)

// This struct contains the main application state. I have fluxy intentions.
type State struct {
	Mode string

	Command               []rune
	CommandCursorPosition int

	// All the connections that are currently made to outside services.
	Connections      []gateway.Connection
	activeConnection int
	connectionSynced bool

	// Interacting with messages
	SelectedMessageIndex int
	BottomDisplayedItem int
	RenderedMessageNumber int

	// Fuzzy picker
	FuzzyPicker FuzzySorter
	fuzzyPickerSelectedItem int
	fuzzyPickerBottomDisplayedItem int
}

func NewInitialState() *State {
	return NewInitialStateMode("chat")
}

func NewInitialStateMode(mode string) *State {
	return &State{
		// The mode the client is in
		Mode: mode,

		// The command the user is typing
		Command:               []rune{},
		CommandCursorPosition: 0,

		// Connection to the server
		Connections: []gateway.Connection{
			gatewaySlack.New(os.Getenv("SLACK_TOKEN_TWO")), // Uncommonspace
			gatewaySlack.New(os.Getenv("SLACK_TOKEN_ONE")), // Gaus Family
		},

		// Which connection in the connections object is active
		activeConnection: 0,
		connectionSynced: false,

		// Interacting with messages
		SelectedMessageIndex: 0,
		BottomDisplayedItem: 0,
		RenderedMessageNumber: -1, // A render loop hasn't run yet.

		// Fuzzy picker data
		FuzzyPicker:       FuzzySorter{},
	}
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
	if s.activeConnection > len(s.Connections)-1 {
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
