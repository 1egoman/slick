package main

import (
	"github.com/1egoman/slime/status"
	"github.com/1egoman/slime/gateway" // The thing to interface with slack
)

// This struct contains the main application state. I have fluxy intentions.
type State struct {
	Mode string

	Command               []rune
	CommandCursorPosition int

	// A list of all keys that have been pressed to make up the current command.
	KeyStack              []rune

	// All the connections that are currently made to outside services.
	Connections      []gateway.Connection
	activeConnection int
	connectionSynced bool

	// Interacting with messages
	SelectedMessageIndex  int
	BottomDisplayedItem   int
	RenderedMessageNumber int

	// Fuzzy picker
	FuzzyPicker                    FuzzySorter
	fuzzyPickerSelectedItem        int
	fuzzyPickerBottomDisplayedItem int

	// Status message
	Status status.Status

	// Actions to perform when a user presses a key
	KeyActions []KeyAction

	// A map of configuration options for the editor.
	Configuration map[string]string
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
		Connections: []gateway.Connection{},

		// Which connection in the connections object is active
		activeConnection: 0,
		connectionSynced: false,

		// Interacting with messages
		SelectedMessageIndex:  0,
		BottomDisplayedItem:   0,
		RenderedMessageNumber: -1, // A render loop hasn't run yet.

		// Fuzzy picker data
		FuzzyPicker: FuzzySorter{},

		// Status message
		Status: status.Status{},

		// Configuration options
		Configuration: map[string]string{
			// Should relative line numbers be shown for each message?
			"MessageList.RelativeLine": "true",
			// The format for the tiemstamp in front of each message.
			// Reference date: `Mon Jan 2 15:04:05 MST 2006`
			"MessageList.TimestampFormat": " 15:04:05",

			"CommandBar.PrefixColor": "white:blue:",
			"CommandBar.TextColor": "white::",
			"StatusBar.ActiveConnectionColor": "white:blue:",
			"StatusBar.GatewayConnectedColor": "white::",
			"StatusBar.GatewayConnectingColor": ":darkmagenta:",
			"StatusBar.GatewayFailedColor": ":red:",
			"StatusBar.LogColor": "white::",
			"StatusBar.ErrorColor": "darkmagenta::B",
			"StatusBar.TopBorderColor": ":gray:",

			"Message.ReactionColor": "::",
			"Message.LineNumbersColor": "black:silver:",
			"Message.FileColor":     "::",
			"Message.SelectedColor": ":teal:",
			"Message.Action.Color": "::",
			"Message.Action.HighlightColor": "red::",
			"Message.Attachment.TitleColor": "green::",
			"Message.Attachment.FieldTitleColor": "::B",
			"Message.Attachment.FieldValueColor": "::",
			"Message.Part.AtMentionUserColor": "red::B",
			"Message.Part.AtMentionGroupColor": "yellow::B",
			"Message.Part.ChannelColor": "blue::B",
			"Message.Part.LinkColor": "cyan::BU",
			"Message.LineNumber.Color": "black:silver:",
			"Message.LineNumber.ActiveColor": "black:white:",

			"FuzzyPicker.TopBorderColor": ":gray:",
			"FuzzyPicker.ActiveItemColor": ":teal:",
			"FuzzyPicker.ChannelNotMemberColor": "gray::",
		},
	}
}

func (s *State) ActiveConnection() gateway.Connection {
	if len(s.Connections) > 0 {
		return s.Connections[s.activeConnection]
	} else {
		return nil
	}
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
func (s *State) ActiveConnectionIndex() int {
	return s.activeConnection
}
