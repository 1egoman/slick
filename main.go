package main

import (
	"log"
	"os"
	"path"

	"github.com/1egoman/slick/frontend" // The thing to draw to the screen
	"github.com/1egoman/slick/version"
	"github.com/gdamore/tcell"
)

func main() {
	// Configure logger to log to file
	logFile, err := os.Create(path.Join(os.Getenv("HOME"), ".slicklog"))
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile)
	log.Println("Starting Slick...")

	state := NewInitialState()
	quit := make(chan struct{})

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, _ := tcell.NewScreen()
	term := frontend.NewTerminalDisplay(s)
	s.Init()
	defer s.Fini() // Make sure we clean up after tcell!

	// Initial render.
	render(state, term)

	log.Println("Reading config files...")
	for _, data := range GetConfigFileContents() {
		err := ParseScript(data, state, term)
		if err != nil {
			state.Status.Errorf("lua error: %s", err.Error())
			log.Printf("lua error: %s", err.Error())
		}
	}

	// GOROUTINE: On start, check for a new release and if found update to it.
	go func() {
		if _, ok := state.Configuration["AutoUpdate"]; ok {
			log.Println("Checking for update...")
			if updatedVersion := version.DoUpdate(); updatedVersion != nil {
				state.Status.Printf("Updated to slick %s! Restart to complete.", *updatedVersion)
				render(state, term)
				return
			} else {
				log.Println("No update, continuing...")
			}
		}
	}()

	// GOROUTINE: Handle events coming from the input device (ie, keyboard).
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

		gatewayEvents(state, term)
	}()

	<-quit
	log.Println("Quitting gracefully...")

	// Save each connection and close it in turn.
	if _, ok := state.Configuration["Connection.Cache"]; ok {
		os.MkdirAll(PathToSavedConnections(), 0755)
		for _, connection := range state.Connections {
			err := SaveConnection(connection)
			if err != nil {
				log.Printf("Error saving connection state: %s", err)
			}

			connection.Disconnect()
		}
	}
}
