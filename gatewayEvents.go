package main

import (
	"log"
	"time"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
)

// Once conencted, listen for events from the active gateway. When an event comes in, act on it.
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
					messageHash = data
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
				if text, ok := event.Data["text"].(string); ok {
					state.ActiveConnection().AppendMessageHistory(gateway.Message{
						Sender: sender,
						Text:   text,
						Hash:   messageHash,
					})
				} else {
					log.Printf("WARN: Tried to append a message without a text element in the body: %+v", event)
				}
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
