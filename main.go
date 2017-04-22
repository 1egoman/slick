package main

import (
	"github.com/1egoman/slime/gateway"
	"os"
	"fmt"
)

func main() {
	slack := gateway.Slack(os.Getenv("SLACK_TOKEN"))
	slack.Connect()

	fmt.Println(slack)

	for {
		event := <-slack.Incoming

		switch event.Type {
		case "hello":
			fmt.Println("Hello!")

			// Send an outgoing message
			slack.Outgoing <- gateway.Event{
				Type: "ping",
				Data: map[string]interface{} {
					"foo": "bar",
				},
			}

			// List all channels
			channels, _ := slack.Channels()
			for _, channel := range channels {
				fmt.Printf("Channel: %+v Creator: %+v\n", channel, channel.Creator)
			}

		case "message":
			fmt.Println("Message:", event.Data["text"])
		case "pong":
			fmt.Println("Got pong!")
		}
	}
}
