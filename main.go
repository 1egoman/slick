package main

import (
	"fmt"
	"os"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
)

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

func keyboardEvents(state *State, term *frontend.TerminalDisplay, screen tcell.Screen, quit chan struct{}) {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC:
				close(quit)
				return

			// Escape reverts back to chat mode.
			case tcell.KeyEscape:
				state.Mode = "chat"

			// CTRL-P moves to a channel picker, which is a mode for switching teams and channels
			case tcell.KeyCtrlP:
				if state.Mode != "picker" {
					state.Mode = "picker"
				} else {
					state.Mode = "chat"
				}

			case tcell.KeyEnter:
				command := string(state.Command)
				switch {
				// :q or :quit closes the app
				case command == ":q", command == ":quit":
					close(quit)
					return
				default:
					// By default, just send a message
					message := gateway.Message{
						Sender: state.Gateway.Self(),
						Text: command,
					}

					// Sometimes, a message could have a response. This is for example true in the
					// case of slash commands, sometimes.
					responseMessage, err := state.Gateway.SendMessage(message, state.Gateway.SelectedChannel())

					if err != nil {
						panic(err)
					} else if responseMessage != nil {
						// Got a response command? Append it to the message history.
						state.MessageHistory = append(state.MessageHistory, *responseMessage)
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
		render(*state, term)
	}
}

func gatewayEvents(state *State, term *frontend.TerminalDisplay, connected chan struct{}) {
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
		case "message":
			if event.Data["channel"] == state.Gateway.SelectedChannel().Id {
				// Find a hash for the message, just use the timestamp
				// In message events, the timestamp is `ts`
				// In pong events, the timestamp is `event_ts`
				var messageHash string
				if data, ok := event.Data["ts"].(string); ok {
					messageHash = data;
				} else {
					panic("No ts key in message so can't create a hash for this message!")
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

		render(*state, term)
	}
}

func main() {
	fmt.Println("Loading...")
	state := State{
		// The mode the client is in
		Mode: "chat",

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
	go connect(&state, term, connected)

	// GOROUTINE: Handle events coming from the input device (ie, keyboard).
	quit := make(chan struct{})
	go keyboardEvents(&state, term, s, quit)

	// GOROUTINE: Handle events coming from slack.
	go gatewayEvents(&state, term, connected)

	<-quit
}
