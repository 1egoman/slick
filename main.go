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

//
// COMMAND LINE FLAGS
//

// --log-file
var logFileFlag = flag.String(
	"log-file",
	path.Join(os.TempDir(), fmt.Sprintf("slicklog.%d", os.Getpid())),
	"Location to put a log file.",
)

// --no-config
var noConfigFlag = flag.Bool("no-config", false, "Don't load configuration from slickrc.")

// --version
var versionFlag = flag.Bool("version", false, "Display installed version of slick.")

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

	state := NewInitialState()
	quit := make(chan struct{})

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, _ := tcell.NewScreen()
	term := frontend.NewTerminalDisplay(s)
	s.Init()
	defer s.Fini() // Make sure we clean up after tcell!

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
