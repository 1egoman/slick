package main

import (
	"os"
	"log"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/1egoman/slime/gateway/slack"
	"github.com/gdamore/tcell"
)

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
			gatewaySlack.New(os.Getenv("SLACK_TOKEN_TWO")), // Uncommonspace
			gatewaySlack.New(os.Getenv("SLACK_TOKEN_ONE")), // Gaus Family
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
