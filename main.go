package main

import (
	"fmt"
	"os"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
)

type State struct {
	Mode string

	Command               []rune
	CommandCursorPosition int

	Gateway gateway.Connection

	MessageHistory []gateway.Message
}

func render(state State, term *frontend.TerminalDisplay) {
	term.DrawCommandBar(
		string(state.Command),           // The command that the user is typing
		state.CommandCursorPosition,     // The cursor position
		state.Gateway.SelectedChannel(), // The selected channel
		state.Gateway.Team(),            // The selected team
	)
	term.DrawStatusBar()

	term.DrawMessages(state.MessageHistory)

	term.Render()
}

// Given a state object populated with a gateway, initialize the state with the gateway.
func connect(state *State, term *frontend.TerminalDisplay, connected chan struct{}) {
	// Connect to gateway, then refresh properies about it.
	state.Gateway.Connect()
	state.Gateway.Refresh()

	// Get messages for the selected channel
	selectedChannel := state.Gateway.SelectedChannel()
	if selectedChannel != nil {
		state.MessageHistory, _ = state.Gateway.FetchChannelMessages(*selectedChannel)
	}

	// Inital full render. At this point, all data has come in.
	render(*state, term)

	// We're connected!
	close(connected)
}

func events(state State, term *frontend.TerminalDisplay, screen tcell.Screen, quit chan struct{}) {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				close(quit)
				return

			case tcell.KeyEnter:
				command := string(state.Command)
				switch {
				case command == ":q", command == ":quit":
					close(quit)
					return
				default:
					// By default, just send a message
					state.Gateway.Outgoing() <- gateway.Event{
						Type: "message",
						Direction: "outgoing",
						Data: map[string]interface{} {
							"channel": state.Gateway.SelectedChannel().Id,
							"user": state.Gateway.Self().Id,
							"text": command,
						},
					}
				}
				// Clear the command that was typed.
				state.Command = []rune{}
				state.CommandCursorPosition = 0


			// CTRL + L redraws the screen.
			case tcell.KeyCtrlL:
				screen.Sync()

			// As characters are typed, add to the message.
			case tcell.KeyRune:
				state.Command = append(
					append(state.Command[:state.CommandCursorPosition], ev.Rune()),
					state.Command[state.CommandCursorPosition:]...,
				)
				state.CommandCursorPosition += 1

			// Backspace removes a character.
			case tcell.KeyDEL:
				if state.CommandCursorPosition > 0 {
					state.Command = append(
						state.Command[:state.CommandCursorPosition-1],
						state.Command[state.CommandCursorPosition:]...,
					)
					state.CommandCursorPosition -= 1
				}

			// Arrows right and left move the cursor
			case tcell.KeyLeft:
				if state.CommandCursorPosition >= 1 {
					state.CommandCursorPosition -= 1
				}
			case tcell.KeyRight:
				if state.CommandCursorPosition < len(state.Command) {
					state.CommandCursorPosition += 1
				}
			}
		case *tcell.EventResize:
			screen.Sync()
		}
		render(state, term)
	}
}

func main() {
	fmt.Println("Loading...")
	state := State{
		// The mode the client is in
		Mode: "normal",

		// The command the user is typing
		Command:               []rune{},
		CommandCursorPosition: 0,

		// Connection to the server
		Gateway: gateway.Slack(os.Getenv("SLACK_TOKEN")),

		// A slice of all messages in the currently active channel
		MessageHistory: []gateway.Message{},
	}

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, _ := tcell.NewScreen()
	term := frontend.NewTerminalDisplay(s)
	s.Init()
	defer s.Fini() // Make sure we clean up after tcell!

	// Initial render.
	render(state, term)

	// Connect to the server
	// Once this goroutine finishes, it closed the connected channel. This is used as a signal by
	// the gateway events goroutine to start working.
	connected := make(chan struct{})
	connect(&state, term, connected)

	// GOROUTINE: Handle keyboard events.
	quit := make(chan struct{})
	go events(state, term, s, quit)

	// // GOROUTINE: Handle connection incoming and outgoing messages
	go func(state State, connected chan struct{}) {
		// Wait to be connected before handling events.
		<-connected

		for {
			event := <-state.Gateway.Incoming()

			switch event.Type {
			case "hello":
				state.MessageHistory = append(state.MessageHistory, gateway.Message{
					Sender: nil,
					Text:   "Got Hello...",
					Hash:   "hello",
				})

				// Send an outgoing message
				state.Gateway.Outgoing() <- gateway.Event{
					Type: "ping",
					Data: map[string]interface{}{
						"foo": "bar",
					},
				}

			// When a message is received for the selected channel, add to the message history
			// "message" events come in when the gateway receives a message sent by someone else.
			// "pong" events come in when a message the user just sent is received successfully.
			case "message", "pong":
				if event.Data["channel"] == state.Gateway.SelectedChannel().Id {
					// Find a hash for the message, just use the timestamp
					// In message events, the timestamp is `ts`
					// In pong events, the timestamp is `event_ts`
					var messageHash string
					if data, ok := event.Data["ts"].(string); ok {
						messageHash = data;
					} else if data, ok := event.Data["event_ts"].(string); ok {
						messageHash = data;
					} else {
						panic("No ts or event_ts key in message or pong, so can't create a hash for this message!")
					}

					// See if the message is already in the history
					alreadyInHistory := false
					for _, msg := range state.MessageHistory {
						if msg.Hash == messageHash {
							// Message with that hash is already in the history, no need to add
							// again...
							alreadyInHistory = true
							break
						}
					}
					if alreadyInHistory {
						break
					}

					// Figure out who sent the message
					var sender *gateway.User
					if event.Data["user"] != nil {
						var err error
						sender, err = state.Gateway.UserById(event.Data["user"].(string))
						if err != nil {
							panic(err)
						}
					} else {
						sender = nil
					}

					// Add message to history
					state.MessageHistory = append(state.MessageHistory, gateway.Message{
						Sender: sender,
						Text:   event.Data["text"].(string),
						Hash:   messageHash,
					})
				}
			}

			render(state, term)
		}
	}(state, connected)

	<-quit
}
