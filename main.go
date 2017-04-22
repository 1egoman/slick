package main

import (
	"fmt"
	"os"
	"time"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
)

func main() {
	slack := gateway.Slack(os.Getenv("SLACK_TOKEN"))
	slack.Connect()

	fmt.Println(slack)

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, _ := tcell.NewScreen()
	term := frontend.NewTerminalDisplay(s)

	s.Init()
	term.DrawStatusBar()
	s.Show()

	quit := make(chan struct{})
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyCtrlL:
					s.Sync()
				}
			case *tcell.EventResize:
				s.Sync()
			}
		}
	}()

	time.Sleep(2000 * time.Millisecond)
	term.DrawStatusBar()
	s.Show()

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
