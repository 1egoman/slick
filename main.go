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
func connect(state State, term *frontend.TerminalDisplay, connected chan struct{}) {
	// Connect to gateway, then refresh properies about it.
	state.Gateway.Connect()
	state.Gateway.Refresh()

	// Get messages for the selected channel
	selectedChannel := state.Gateway.SelectedChannel()
	if selectedChannel != nil {
		state.MessageHistory, _ = state.Gateway.FetchChannelMessages(*selectedChannel)
	}

	// Inital full render. At this point, all data has come in.
	render(state, term)

	// We're connected!
	close(connected)
}

func events(state State, term *frontend.TerminalDisplay, screen tcell.Screen, quit chan struct{}) {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				close(quit)
				return

			// case tcell.KeyEnter:
			//

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

	// GOROUTINE: Connect to the server
	connected := make(chan struct{})
	go connect(state, term, connected)

	// GOROUTINE: Handle keyboard events.
	quit := make(chan struct{})
	go events(state, term, s, quit)

	// // GOROUTINE: Handle connection incoming and outgoing messages
	go func(state State) {
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
			case "message":
				if event.Data["channel"] == state.Gateway.SelectedChannel().Id {
					messageHash := event.Data["ts"].(string)

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

			// When a pong is received, log it too.
			case "pong":
				state.MessageHistory = append(state.MessageHistory, gateway.Message{
					Sender: nil,
					Text:   "Got Pong...",
					Hash:   "pong",
				})
			}

			render(state, term)
		}
	}(state)

	<-quit
}
