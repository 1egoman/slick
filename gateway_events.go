package main

import (
	"log"
	"time"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
)

// Once conencted, listen for events from the active gateway. When an event comes in, act on it.
func gatewayEvents(state *State, term *frontend.TerminalDisplay) {
	cachedUsers := make(map[string]*gateway.User)

	for {
    if state.ActiveConnection() == nil {
			time.Sleep(100 * time.Millisecond) // Sleep to lower the speed of the loop for debugging reasons.
      continue
    }

		// Before events can run, confirm that the a channel is selected.
		hasFetchedChannel := state.ActiveConnection().SelectedChannel() != nil

		// Is the channel empty? If so, move to the next ieration.
		// We want the loop to always be running, so that if the reference that
		// state.ActiveConnection() points to behind the scenes changes, the we won't be blocking
		// listening for events on an old reference.
		if len(state.ActiveConnection().Incoming()) == 0 || !hasFetchedChannel {
			time.Sleep(100 * time.Millisecond) // Sleep to lower the speed of the loop for debugging reasons.
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
        if event.Data["subtype"] == "message_deleted" {
          // If a message was deleted, then delete the message from the message history
          for index, msg := range state.ActiveConnection().MessageHistory() {
            if msg.Hash == event.Data["deleted_ts"] {
              state.ActiveConnection().DeleteMessageHistory(index)
            }
          }
        } else {
          // JUST A NORMAL MESSAGE!

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

          message, err := state.ActiveConnection().ParseMessage(event.Data, cachedUsers)
          if err == nil {
            // Add message to history
            state.ActiveConnection().AppendMessageHistory(*message)

			// If the user that sent the message was typing, they aren't anymore.
			state.ActiveConnection().TypingUsers().Remove(message.Sender.Name)
          } else {
            log.Fatalf(err.Error())
          }
        }
			} else {
				log.Printf("Channel value", channel)
			}

      // Notify the user about a message if the message wasn't sent by the user.
      if event.Data["subtype"] == nil && event.Data["user"] != state.ActiveConnection().Self().Id {
        message, err := state.ActiveConnection().ParseMessage(event.Data, cachedUsers)
        if err != nil {
          log.Printf(err.Error())
        } else {
          Notification(message.Sender.Name, message.Text)
        }
      }

    // case "reaction_added":
    //   // If a message was deleted, then delete the message from the message history
    //   for index, msg := range state.ActiveConnection().MessageHistory() {
    //     if msg.Hash == event.Data["event_ts"] {
    //       for _, reaction := range msg.Reactions {
    //         event.Data["reaction"] // == "smile"
    //         event.Data["item_user"] // == "U0M9S59T2"
    //       }
    //     }
    //   }

		// When a user starts typing, then display that they are typing.
		case "user_typing":
			if state.ActiveConnection() != nil {
				if channel := state.ActiveConnection().SelectedChannel(); event.Data["channel"] == channel.Id {
					if userId, ok := event.Data["user"].(string); ok {
						user, err := state.ActiveConnection().UserById(userId)
						if err != nil {
							log.Println(err.Error())
						} else if user != nil {
							// Add user to the list of users typing.
							state.ActiveConnection().TypingUsers().Add(user.Name, time.Now())
						} else {
							log.Println("User in `user_typing` event was nil, ignoring...")
						}
					} else {
						log.Println("User id in `user_typing` raw event was not coersable to string, ignoring...")
					}
				}
			}

		case "":
			log.Printf("Unknown event received: %+v", event)
		}

		render(state, term)
	}
}
