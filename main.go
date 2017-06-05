package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/1egoman/slick/frontend" // The thing to draw to the screen
	"github.com/1egoman/slick/version"
	"github.com/gdamore/tcell"
)

var (
	// Main screen to draw to
	screen tcell.Screen

	// A central place to store state.
	// This struct is passed around to basically everything,
	// and used to set and reference state in a predictable way.
	state *State = NewInitialState()

	// A channel to prop the app open. When the app is to be
	// closed, this channel is closed, which causes everything
	// to come to a halt.
	quit chan struct{} = make(chan struct{})

	// An abstraction on top of the display that handles
	// drawing the status bar, command bar, messages, etc
	term *frontend.TerminalDisplay
)

//
// COMMAND LINE FLAGS
//

var (
	// --log-file
	logFileFlag *string = flag.String(
		"log-file",
		path.Join(os.TempDir(), "slick.log"),
		"Location to put a log file.",
	)

	// --no-config
	noConfigFlag *bool = flag.Bool("no-config", false, "Don't load configuration from slickrc.")

	// --version
	versionFlag *bool = flag.Bool("version", false, "Display installed version of slick.")
)

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(version.Version())
		return
	}

	// Configure logger to log to file
	// 1. Create the folder for the log file.
	// 2. Create the log file.
	// 3. Pass the log file instance to the logger.
	if err := os.MkdirAll(path.Dir(*logFileFlag), 0755); err != nil {
		log.Fatal(err)
	}
	logFile, err := os.Create(*logFileFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile)
	log.Println("Starting Slick...")

	// Instantiate the screen and terminal display
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, _ = tcell.NewScreen()
	screen.Init()
	defer screen.Fini()
	term = frontend.NewTerminalDisplay(screen)

	// Initial render.
	render(state, term)

	if !*noConfigFlag {
		log.Println("Reading config files...")
		for _, data := range GetConfigFileContents() {
			err := ParseScript(data, state, term)
			if err != nil {
				state.Status.Errorf("lua error: %s", err.Error())
				log.Printf("lua error: %s", err.Error())
			}
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
				screen.Fini()
				panic(r)
			}
		}()
		keyboardEvents(state, term, screen, quit)
	}()

	// GOROUTINE: Handle events coming from slack.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				screen.Fini()
				panic(r)
			}
		}()

		gatewayEvents(state, term)
	}()

	<-quit
	log.Println("Quitting gracefully...")

	// Save each connection and close it in turn.
	if _, ok := state.Configuration["Connection.Cache"]; ok {
		err := os.MkdirAll(PathToSavedConnections(), 0755)
		if err != nil {
			log.Printf("Error creating folder to store connection state: %s", err)
			return
		}
		for _, connection := range state.Connections {
			err := SaveConnection(connection)
			if err != nil {
				log.Printf("Error saving connection state: %s", err)
				continue
			}

			connection.Disconnect()
		}
	}
}
