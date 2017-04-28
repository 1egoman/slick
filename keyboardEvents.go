package main

import (
	"log"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
)

func keyboardEvents(state *State, term *frontend.TerminalDisplay, screen tcell.Screen, quit chan struct{}) {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			log.Printf("Keypress: %+v", ev.Name())
			switch {
			case ev.Key() == tcell.KeyCtrlC:
				log.Println("CLOSE QUIT 1")
				close(quit)
				return

			// Escape reverts back to chat mode.
			case ev.Key() == tcell.KeyEscape:
				state.Mode = "chat"

			// CTRL-P moves to a channel picker, which is a mode for switching teams and channels
			case ev.Key() == tcell.KeyCtrlP:
				if state.Mode != "pick" {
					state.Mode = "pick"
				} else {
					state.Mode = "chat"
				}

			// CTRL + L redraws the screen.
			case ev.Key() == tcell.KeyCtrlL:
				screen.Sync()

			//
			// MOVEMENT BETWEEN CONNECTIONS
			//
			case ev.Key() == tcell.KeyCtrlQ:
				state.SetPrevActiveConnection()
			case ev.Key() == tcell.KeyCtrlW:
				state.SetNextActiveConnection()

			//
			// MOVEMENT BETWEEN ITEMS IN THE FUZZY PICKER
			//
			case state.Mode == "pick" && ev.Key() == tcell.KeyCtrlJ:
				if state.fuzzyPickerSelectedItem > 0 {
					state.fuzzyPickerSelectedItem -= 1
				}
			case state.Mode == "pick" && ev.Key() == tcell.KeyCtrlK:
				state.fuzzyPickerSelectedItem += 1

			//
			// COMMAND BAR
			//

			case ev.Key() == tcell.KeyEnter:
				command := string(state.Command)

				if state.Mode == "chat" {
					// When in chat mode, run a command or send a message.
					switch {
					// :q or :quit closes the app
					case command == ":q", command == ":quit":
						log.Println("CLOSE QUIT 2")
						close(quit)
						return

					// By default, just send a message
					default:
						message := gateway.Message{
							Sender: state.ActiveConnection().Self(),
							Text:   command,
						}

						// Sometimes, a message could have a response. This is for example true in the
						// case of slash commands, sometimes.
						responseMessage, err := state.ActiveConnection().SendMessage(
							message,
							state.ActiveConnection().SelectedChannel(),
						)

						if err != nil {
							log.Fatal(err)
						} else if responseMessage != nil {
							// Got a response command? Append it to the message history.
							state.ActiveConnection().AppendMessageHistory(*responseMessage)
						}
					}
				} else if state.Mode == "pick" {
					// We want to choose the selected option.
					selectedItem := state.FuzzyPickerSorter.Items[state.fuzzyPickerSelectedItem]
					selectedConnectionName := selectedItem.Connection
					selectedChannelName := selectedItem.Channel

					// Find the selected connction's index in the main connection slice
					selectedConnectionIndex := -1
					for index, item := range state.Connections {
						if item.Name() == selectedConnectionName {
							selectedConnectionIndex = index
							break
						}
					}
					if selectedConnectionIndex == -1 {
						log.Fatalf("Tried to select connection %s that isn't in the slice of connections", selectedConnectionName)
					}

					// Find the selected channel's index to the channel list slice
					var selectedChannel *gateway.Channel
					for _, item := range state.Connections[selectedConnectionIndex].Channels() {
						if item.Name == selectedChannelName {
							selectedChannel = &item
							break
						}
					}
					if selectedChannel == nil {
						log.Fatalf(
							"Tried to select channel %s that isn't in the slice of channels for conenction %s",
							selectedChannelName,
							selectedConnectionName,
						)
					}

					log.Printf("Selecting connection %s and channel %s", selectedConnectionName, selectedChannel.Name)

					// Set the active connection with the discovered index, and also set a new selected
					// channel.
					state.SetActiveConnection(selectedConnectionIndex)
					state.Connections[selectedConnectionIndex].SetSelectedChannel(selectedChannel)
					state.Mode = "chat"
				}
				// Clear the command that was typed.
				state.Command = []rune{}
				state.CommandCursorPosition = 0

			// As characters are typed, add to the message.
			case ev.Key() == tcell.KeyRune:
				state.Command = append(
					append(state.Command[:state.CommandCursorPosition], ev.Rune()),
					state.Command[state.CommandCursorPosition:]...,
				)
				state.CommandCursorPosition += 1

			// Backspace removes a character.
			case ev.Key() == tcell.KeyDEL:
				if state.CommandCursorPosition > 0 {
					state.Command = append(
						state.Command[:state.CommandCursorPosition-1],
						state.Command[state.CommandCursorPosition:]...,
					)
					state.CommandCursorPosition -= 1
				}

			// Arrows right and left move the cursor
			case ev.Key() == tcell.KeyLeft:
				if state.CommandCursorPosition >= 1 {
					state.CommandCursorPosition -= 1
				}
			case ev.Key() == tcell.KeyRight:
				if state.CommandCursorPosition < len(state.Command) {
					state.CommandCursorPosition += 1
				}
			}
		case *tcell.EventResize:
			screen.Sync()
		}

		// Render after each loop
		render(state, term)
	}
}
