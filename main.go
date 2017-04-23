package main

import (
	// "fmt"
	"os"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
)

type State struct {
	Mode string

	Command []rune
	CommandCursorPosition int

	Gateway gateway.Connection
}

func render(state State, term frontend.TerminalDisplay) {
	term.DrawCommandBar(
		string(state.Command), // The command that the user is typing
		state.CommandCursorPosition, // The cursor position
	)
	term.DrawStatusBar()

	term.DrawChannels(state.Gateway)

	term.Render()
}

func main() {
	state := State{
		// The mode the client is in
		Mode: "normal",

		// The command the user is typing
		Command: []rune{},
		CommandCursorPosition: 0,

		// Connection to the server
		Gateway: gateway.Slack(os.Getenv("SLACK_TOKEN")),
	}

	// Connect to gateway, then refresh properies about it.
	state.Gateway.Connect()
	state.Gateway.Refresh()

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, _ := tcell.NewScreen()
	term := frontend.NewTerminalDisplay(s)
	s.Init()

	render(state, *term)

	quit := make(chan struct{})
	go func() {
		for {
			ev := s.PollEvent()
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
					s.Sync()

				// As characters are typed, add to the message.
				case tcell.KeyRune:
					state.Command = append(
						append(state.Command[:state.CommandCursorPosition], ev.Rune()),
						state.Command[state.CommandCursorPosition:]...
					)
					state.CommandCursorPosition += 1

				// Backspace removes a character.
				case tcell.KeyDEL:
					if state.CommandCursorPosition > 0 {
						state.Command = append(
							state.Command[:state.CommandCursorPosition-1],
							state.Command[state.CommandCursorPosition:]...
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
				render(state, *term)
			case *tcell.EventResize:
				s.Sync()
			}
		}
	}()

	<-quit
	s.Fini()

	// for {
	// 	event := <-slack.Incoming
	//
	// 	switch event.Type {
	// 	case "hello":
	// 		fmt.Println("Hello!")
	//
	// 		// Send an outgoing message
	// 		slack.Outgoing <- gateway.Event{
	// 			Type: "ping",
	// 			Data: map[string]interface{} {
	// 				"foo": "bar",
	// 			},
	// 		}
	//
	// 		// List all channels
	// 		channels, _ := slack.Channels()
	// 		for _, channel := range channels {
	// 			fmt.Printf("Channel: %+v Creator: %+v\n", channel, channel.Creator)
	// 		}
	//
	// 	case "message":
	// 		fmt.Println("Message:", event.Data["text"])
	// 	case "pong":
	// 		fmt.Println("Got pong!")
	// 	}
	// }
}
