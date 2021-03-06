package main

import (
	"github.com/1egoman/slick/gateway" // The thing to interface with slack
	"github.com/1egoman/slick/modal"
	"github.com/1egoman/slick/status"
)

// This struct contains the main application state. I have fluxy intentions.
type State struct {
	Mode string

	Command               []rune
	CommandCursorPosition int

	// Is the current session of the client offline?
	Offline bool

	// A list of all keys that have been pressed to make up the current command.
	KeyStack []rune

	// All the connections that are currently made to outside services.
	Connections      []gateway.Connection
	activeConnection int
	connectionSynced bool

	// Interacting with messages
	SelectedMessageIndex  int
	BottomDisplayedItem   int
	RenderedMessageNumber int
	RenderedAllMessages   bool

	// Fuzzy picker
	SelectionInput                    SelectionInput
	SelectionInputSelectedItem        int
	SelectionInputBottomDisplayedItem int

	// Status message
	Status status.Status

	// Modal
	Modal modal.Modal

	// Handlers to bind to specific actions. For example, when the user presses some keys,  when we
	// switch connections, etc...
	EventActions []EventAction

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

		// Starts online
		Offline: false,

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
		RenderedAllMessages:   false,

		// Selection Input data
		SelectionInput: SelectionInput{},

		// Status message
		Status: status.Status{},

		// Modal
		Modal: modal.Modal{},

		// Configuration options
		Configuration: map[string]string{
			// Disable connection caching
			"Connection.Cache": "true",
			// Should relative line numbers be shown for each message?
			// "Message.RelativeLine": "true",

			// Disable auto-updates
			// "AutoUpdate": "true",

			// The format for the tiemstamp in front of each message.
			// Reference date: `Mon Jan 2 15:04:05 MST 2006`
			"Message.TimestampFormat": " Jan 2 15:04:05",

			// How many messages should Ctrl-U / Ctrl-D page by?
			"Message.PageAmount": "12",

			// User online status settings
			"Message.Sender.OnlinePrefix":       "*",
			"Message.Sender.OnlinePrefixColor":  "green::",
			"Message.Sender.OfflinePrefix":      "*",
			"Message.Sender.OfflinePrefixColor": "silver::",

			"Message.ReactionColor":              "::",
			"Message.FileColor":                  "::",
			"Message.SelectedColor":              ":teal:",
			"Message.Action.Color":               "::",
			"Message.Action.HighlightColor":      "red::",
			"Message.Attachment.TitleColor":      "green::",
			"Message.Attachment.FieldTitleColor": "::B",
			"Message.Attachment.FieldValueColor": "::",
			"Message.Part.AtMentionUserColor":    "red::B",
			"Message.Part.AtMentionGroupColor":   "yellow::B",
			"Message.Part.ChannelColor":          "blue::B",
			"Message.Part.LinkColor":             "cyan::BU",
			"Message.LineNumber.Color":           "white::",
			"Message.LineNumber.ActiveColor":     "teal::",
			"Message.UnconfirmedColor":           "gray::",

			"CommandBar.PrefixColor":  "::",
			"CommandBar.TextColor":    "::",
			"CommandBar.NewLineColor": "gray::B",

			"StatusBar.Color":                  "::",
			"StatusBar.ModeColor":              "::",
			"StatusBar.ActiveConnectionColor":  "white:blue:",
			"StatusBar.GatewayConnectedColor":  "white::",
			"StatusBar.GatewayConnectingColor": ":darkmagenta:",
			"StatusBar.GatewayFailedColor":     ":red:",
			"StatusBar.LogColor":               "white::",
			"StatusBar.ErrorColor":             "darkmagenta::B",
			"StatusBar.TopBorderColor":         ":gray:",

			"FuzzyPicker.TopBorderColor":  ":gray:",
			"FuzzyPicker.ActiveItemColor": "::B",
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
