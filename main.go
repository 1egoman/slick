package main

import (
	"os"
	"log"
	"time"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
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

func keyboardEvents(state *State, term *frontend.TerminalDisplay, screen tcell.Screen, quit chan struct{}) {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			log.Printf("Keypress: %+v", ev.Name())
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				close(quit)
				return

			// Escape reverts back to chat mode.
			case ev.Key() == tcell.KeyEscape:
				state.Mode = "chat"

			// CTRL-P moves to a channel picker, which is a mode for switching teams and channels
			case ev.Key() == tcell.KeyCtrlP:
				if state.Mode != "picker" {
					state.Mode = "picker"
				} else {
					state.Mode = "chat"
				}

			// CTRL + L redraws the screen.
			case ev.Key() == tcell.KeyCtrlL:
				screen.Sync()

			//
			// MOVEMENT BETWEEN CONNECTIONS
			//
			case ev.Key() == tcell.KeyCtrlQ:
				state.SetPrevActiveConnection()
			case ev.Key() == tcell.KeyCtrlW:
				state.SetNextActiveConnection()

			//
			// COMMAND BAR
			//

			case ev.Key() == tcell.KeyEnter:
				command := string(state.Command)
				switch {
				// :q or :quit closes the app
				case command == ":q", command == ":quit":
					close(quit)
					return
				default:
					// By default, just send a message
					message := gateway.Message{
						Sender: state.ActiveConnection().Self(),
						Text: command,
					}

					// Sometimes, a message could have a response. This is for example true in the
					// case of slash commands, sometimes.
					responseMessage, err := state.ActiveConnection().SendMessage(message, state.ActiveConnection().SelectedChannel())

					if err != nil {
						log.Fatal(err)
					} else if responseMessage != nil {
						// Got a response command? Append it to the message history.
						state.ActiveConnection().AppendMessageHistory(*responseMessage)
					}
				}
				// Clear the command that was typed.
				state.Command = []rune{}
				state.CommandCursorPosition = 0


			// As characters are typed, add to the message.
			case ev.Key() == tcell.KeyRune:
				state.Command = append(
					append(state.Command[:state.CommandCursorPosition], ev.Rune()),
					state.Command[state.CommandCursorPosition:]...,
				)
				state.CommandCursorPosition += 1

			// Backspace removes a character.
			case ev.Key() == tcell.KeyDEL:
				if state.CommandCursorPosition > 0 {
					state.Command = append(
						state.Command[:state.CommandCursorPosition-1],
						state.Command[state.CommandCursorPosition:]...,
					)
					state.CommandCursorPosition -= 1
				}

			// Arrows right and left move the cursor
			case ev.Key() == tcell.KeyLeft:
				if state.CommandCursorPosition >= 1 {
					state.CommandCursorPosition -= 1
				}
			case ev.Key() == tcell.KeyRight:
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

func gatewayEvents(state *State, term *frontend.TerminalDisplay, connected chan struct{}) {
	// Wait to be connected before handling events.
	<-connected

	for {
		// Before events can run, confirm that the a channel is selected.
		hasFetchedChannel := state.ActiveConnection().SelectedChannel() != nil

		// Is the channel empty? If so, move to the next ieration.
		// We want the loop to always be running, so that if the reference that
		// state.ActiveConnection() points to behind the scenes changes, the we won't be blocking
		// listening for events on an old reference.
		if len(state.ActiveConnection().Incoming()) == 0 || !hasFetchedChannel {
			time.Sleep(100 * time.Millisecond) // Sleep to lower the speef od the loop for debugging reasons.
			continue
		}

		// Now that we know there are events, grab one and handle it.
		event := <-state.ActiveConnection().Incoming()
		log.Printf("Received event: %+v", event)

		switch event.Type {
		case "hello":
			state.ActiveConnection().AppendMessageHistory(gateway.Message{
				Sender: nil,
				Text:   "Got Hello...",
				Hash:   "hello",
			})

			// Send an outgoing message
			state.ActiveConnection().Outgoing() <- gateway.Event{
				Type: "ping",
				Data: map[string]interface{}{
					"foo": "bar",
				},
			}

		// When a message is received for the selected channel, add to the message history
		// "message" events come in when the gateway receives a message sent by someone else.
		case "message":
			if channel := state.ActiveConnection().SelectedChannel(); event.Data["channel"] == channel.Id {
				// Find a hash for the message, just use the timestamp
				// In message events, the timestamp is `ts`
				// In pong events, the timestamp is `event_ts`
				var messageHash string
				if data, ok := event.Data["ts"].(string); ok {
					messageHash = data;
				} else {
					log.Fatal("No ts key in message so can't create a hash for this message!")
				}

				// See if the message is already in the history
				alreadyInHistory := false
				for _, msg := range state.ActiveConnection().MessageHistory() {
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
					sender, err = state.ActiveConnection().UserById(event.Data["user"].(string))
					if err != nil {
						log.Fatal(err)
					}
				} else {
					sender = nil
				}

				// Add message to history
				state.ActiveConnection().AppendMessageHistory(gateway.Message{
					Sender: sender,
					Text:   event.Data["text"].(string),
					Hash:   messageHash,
				})
			} else {
				// 1
				log.Printf("Channel value", channel)
			}

		case "":
			log.Printf("Unknown event received: %+v", event)
		}

		render(state, term)
	}
}

func main() {
	// Configure logger to log to file
	logFile, err := os.Create("./log")
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile)
	log.Println("Starting Slime...")

	state := &State{
		// The mode the client is in
		Mode: "chat",

		// The command the user is typing
		Command:               []rune{},
		CommandCursorPosition: 0,

		// Connection to the server
		Connections: []gateway.Connection{
			gateway.Slack(os.Getenv("SLACK_TOKEN_TWO")), // Uncommonspace
			gateway.Slack(os.Getenv("SLACK_TOKEN_ONE")), // Gaus Family
		},

		// Which connection in the connections object is active
		activeConnection: 0,
		connectionSynced: false,
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
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.Fini()
				panic(r)
			}
		}()
		connect(state, term, connected)
	}()

	// GOROUTINE: Handle events coming from the input device (ie, keyboard).
	quit := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.Fini()
				panic(r)
			}
		}()
		keyboardEvents(state, term, s, quit)
	}()

	// GOROUTINE: Handle events coming from slack.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.Fini()
				panic(r)
			}
		}()

		gatewayEvents(state, term, connected)
	}()

	<-quit
	log.Println("Quitting gracefully...")
}
