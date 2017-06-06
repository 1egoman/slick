package main

import (
	"log"
	"time"

	"github.com/1egoman/slick/frontend" // The thing to draw to the screen
	"github.com/1egoman/slick/gateway"  // The thing to interface with slack
)

// Once conencted, listen for events from the active gateway. When an event comes in, act on it.
func gatewayEvents(state *State, term *frontend.TerminalDisplay) {
	cachedUsers := make(map[string]*gateway.User)

	// Keep track of how many connections are disconnected. If all connections are disconnected,
	// then exit the goroutine.
	totalDisconnected := 0

	for totalDisconnected < len(state.Connections) {
		for _, conn := range state.Connections {
			if conn == nil {
				// Sleep to lower the speed of the loop for debugging reasons.
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if conn.Status() == gateway.DISCONNECTED {
				totalDisconnected += 1
			}

			// Before events can run, confirm that the a channel is selected.
			hasFetchedChannel := conn.SelectedChannel() != nil

			// Is the channel empty? If so, move to the next ieration.
			// We want the loop to always be running, so that if the reference that
			// conn points to behind the scenes changes, the we won't be blocking
			// listening for events on an old reference.
			if len(conn.Incoming()) == 0 || !hasFetchedChannel {
				// Sleep to lower the speed of the loop for debugging reasons.
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Now that we know there are events, grab one and handle it.
			event := <-conn.Incoming()
			log.Printf("Received event: %+v", event)

			switch event.Type {
			case "hello":
				// Send an outgoing message
				conn.Outgoing() <- gateway.Event{
					Type: "ping",
					Data: map[string]interface{}{
						"type": "slick",
						"when": int(time.Now().Unix()),
					},
				}

			case "desktop_notification":
				if title, ok := event.Data["title"].(string); ok {
					if content, ok := event.Data["content"].(string); ok {
						if image, ok := event.Data["avatarImage"].(string); ok {
							Notification(title, content, image)
						}
					}
				}

			// When a message is received for the selected channel, add to the message history
			// "message" events come in when the gateway receives a message sent by someone else.
			case "message":
				if channel := conn.SelectedChannel(); event.Data["channel"] == channel.Id {
					if event.Data["subtype"] == "message_deleted" {
						// If a message was deleted, then delete the message from the message history
						for index, msg := range conn.MessageHistory() {
							if msg.Hash == event.Data["deleted_ts"] {
								conn.DeleteMessageHistory(index)
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
						for _, msg := range conn.MessageHistory() {
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

						message, err := conn.ParseMessage(event.Data, cachedUsers)
						if err == nil {
							// Add message to history
							conn.AppendMessageHistory(*message)

							// Emit event to to be handled by lua scripts
							EmitEvent(state, EVENT_MESSAGE_RECEIVED, map[string]string{
								"text":   message.Text,
								"sender": message.Sender.Name,
							})

							// If the user that sent the message was typing, they aren't anymore.
							conn.TypingUsers().Remove(message.Sender.Name)
						} else {
							log.Fatalf(err.Error())
						}
					}
				} else {
					log.Println("Channel value", channel)
				}

				// case "reaction_added":
				//   // If a message was deleted, then delete the message from the message history
				//   for index, msg := range conn.MessageHistory() {
				//     if msg.Hash == event.Data["event_ts"] {
				//       for _, reaction := range msg.Reactions {
				//         event.Data["reaction"] // == "smile"
				//         event.Data["item_user"] // == "U0M9S59T2"
				//       }
				//     }
				//   }

			// When a user starts typing, then display that they are typing.
			case "user_typing":
				if conn != nil {
					if channel := conn.SelectedChannel(); event.Data["channel"] == channel.Id {
						if userId, ok := event.Data["user"].(string); ok {
							user, err := conn.UserById(userId)
							if err != nil {
								log.Println(err.Error())
							} else if user != nil {
								// Add user to the list of users typing.
								conn.TypingUsers().Add(user.Name, time.Now())
							} else {
								log.Println("User in `user_typing` event was nil, ignoring...")
							}
						} else {
							log.Println("User id in `user_typing` raw event was not coersable to string, ignoring...")
						}
					}
				}

			// When the user's presence value changes, update the active connection
			// {"type":"presence_change","presence":"away","user":"U5FR33U4T"}
			case "presence_change":
				if conn != nil {
					// Get user presence status
					status := false
					if value, ok := event.Data["presence"].(string); ok && value == "active" {
						status = true
					}

					// Get user instance
					if userId, ok := event.Data["user"].(string); ok {
						user, err := conn.UserById(userId)
						if err != nil {
							log.Println(err.Error())
						} else {
							conn.SetUserOnline(user, status)
						}
					}
				}

			// When a reaction is added to a message, update our local copy.
			// {"type":"reaction_added","user":"U5F7KC0CQ","item":{"type":"message","channel":"C5FAJ078R","ts":"1495901274.063169"},"reaction":"grinning","item_user":"U5F7KC0CQ","event_ts":"1495912992.765337","ts":"1495912992.765337"}
			case "reaction_added":
				// First, fetch a reference to the user that is in the message.
				if userId, ok := event.Data["user"].(string); ok {
					user, err := conn.UserById(userId)
					if err != nil {
						log.Println(err.Error())
					} else {
						log.Println("No error!")

						// Next, fetch the message hash and emoji that was reacted with.
						if item, ok := event.Data["item"].(map[string]interface{}); ok {
							hash := item["ts"]
							if emoji, ok := event.Data["reaction"].(string); ok {
								// Loop through all messages to find the one that this event
								// references.
								messages := conn.MessageHistory()
								for messageIndex, message := range messages {
									if message.Hash == hash {

										// Loop through each reaction on the message. If someone
										// else already reacted with the emoji we reacted with, then
										// add our username to that reaction.
										foundReaction := false
										for reactionIndex, reaction := range message.Reactions {
											if reaction.Name == emoji {
												log.Printf("Reaction %s found for %s, so adding user %+v...", emoji, hash, user)
												messages[messageIndex].Reactions[reactionIndex].Users = append(
													reaction.Users,
													user,
												)
												foundReaction = true
												break
											}
										}

										// Otherwise, create a new reaction.
										if !foundReaction {
											log.Printf("No reaction %s found for %s, so adding...", emoji, hash)
											messages[messageIndex].Reactions = append(
												message.Reactions,
												gateway.Reaction{
													Name:  emoji,
													Users: []*gateway.User{user},
												},
											)
										}
									}
								}
								conn.SetMessageHistory(messages)
							}
						}
					}
				}

			// When a reactino is removed from a message, update our local copy.
			// {"type":"reaction_removed","user":"U5F7KC0CQ","item":{"type":"message","channel":"C5FAJ078R","ts":"1495901274.063169"},"reaction":"slightly_smiling_face","item_user":"U5F7KC0CQ","event_ts":"1495927732.484253","ts":"1495927732.484253"}
			case "reaction_removed":
				// First, fetch a reference to the user that is in the message.
				if userId, ok := event.Data["user"].(string); ok {
					user, err := conn.UserById(userId)
					if err != nil {
						log.Printf(err.Error())
					} else {
						// Next, fetch the message hash and emoji that was reacted with.
						if item, ok := event.Data["item"].(map[string]interface{}); ok {
							hash := item["ts"]
							if emoji, ok := event.Data["reaction"].(string); ok {
								// Loop through all messages to find the one that this event
								// references.
								messages := conn.MessageHistory()
								for messageIndex, message := range messages {
									if message.Hash == hash {
										// Loop through each reaction on the message. If someone
										// else already reacted with the emoji we reacted with, then
										// add our username to that reaction.
										for reactionIndex, reaction := range message.Reactions {
											if reaction.Name == emoji {
												log.Printf("Reaction %s found for %s, so removing user %+v...", emoji, hash, user)

												// Delete every instance of the user in the reaction
												for userIndex, u := range reaction.Users {
													if u.Id == user.Id {
														messages[messageIndex].Reactions[reactionIndex].Users = append(reaction.Users[:userIndex], reaction.Users[userIndex+1:]...)
													}
												}

												// If all users have been removed from a reaction,
												// then remove the reaction.
												if len(messages[messageIndex].Reactions[reactionIndex].Users) == 0 {
													log.Printf("Reaction %+v empty, so removing...", messages[messageIndex].Reactions[reactionIndex])
													messages[messageIndex].Reactions = append(
														message.Reactions[:reactionIndex],
														message.Reactions[reactionIndex+1:]...,
													)
												}
												break
											}
										}
									}
								}
								conn.SetMessageHistory(messages)
							}
						}
					}
				}

			case "":
				log.Printf("Unknown event received: %+v", event)
			}

			render(state, term)
		}
	}
}
