package main

import (
  "fmt"
	"log"
  "strings"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
)

type CommandType int
const (
  NATIVE CommandType = iota
  SLACK
)

var commands = []FuzzyPickerSlashCommandItem{
  {
    Type: SLACK,
    Name: "Shrug",
    Description: "Appends a shrug to your message",
    Permutations: []string{"/shrug"},
    Arguments: "[your message]",
  },
  {
    Type: NATIVE,
    Name: "Quit",
    Description: "Quits slime.",
    Permutations: []string{"/quit", "/q"},
  },
}

// When a user picks a connection / channel
func OnPickConnectionChannel(state *State) {
  // Assert that the fuzzy picker that's active is of the right type
  if selectedItem, ok := state.FuzzyPicker.Items[state.fuzzyPickerSelectedItem].(FuzzyPickerConnectionChannelItem); ok {
    // We want to choose the selected option.
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
    state.FuzzyPicker.Hide()
  } else {
    log.Fatalln("In pick mode, the fuzzy picker doesn't contain FuzzyPickerConnectionChannelItem's.")
  }
}

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
        state.FuzzyPicker.Hide()

			// 'p' moves to a channel picker, which is a mode for switching teams and channels
			case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'p':
				if state.Mode != "pick" {
					state.Mode = "pick"
          state.FuzzyPicker.Show(OnPickConnectionChannel)

          var items []interface{}
          stringItems := []string{}

          // Accumulate all channels into `items`, and their respective labels into `stringLabels`
          for _, connection := range state.Connections {
            for _, channel := range connection.Channels() {
              // Add string representation of item to `stringItems`
              // Follows the pattern of "my-team #my-channel"
              stringItems = append(stringItems, fmt.Sprintf(
                "%s #%s",
                connection.Name(),
                channel.Name,
              ))

              // Add backing representation of item to `item`
              items = append(items, FuzzyPickerConnectionChannelItem{
                Channel:    channel.Name,
                Connection: connection.Name(),
              })
            }
          }

          // Fuzzy sort the items
          state.FuzzyPicker.Items = items
          state.FuzzyPicker.StringItems = stringItems
				} else {
					state.Mode = "chat"
          state.FuzzyPicker.Hide()
				}
				// 'e' moves to write mode. So does ':'
			case ev.Key() == tcell.KeyRune && ev.Rune() == 'w':
				state.Mode = "writ"
			case ev.Key() == tcell.KeyRune && ev.Rune() == ':':
				state.Mode = "writ"
				state.Command = []rune{':'}
				state.CommandCursorPosition = 1

			// CTRL + L redraws the screen.
			case state.Mode == "chat" && ev.Key() == tcell.KeyCtrlL:
				screen.Sync()

				//
				// MOVEMENT UP AND DOWN THROUGH MESSAGES
				//
			case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'j':
				if state.SelectedMessageIndex > 0 {
					state.SelectedMessageIndex -= 1
					log.Printf("Selecting message %s", state.SelectedMessageIndex)
				}
			case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'k':
				if state.SelectedMessageIndex < len(state.ActiveConnection().MessageHistory())-1 {
					state.SelectedMessageIndex += 1
					log.Printf("Selecting message %s", state.SelectedMessageIndex)
				}

			//
			// MOVEMENT BETWEEN CONNECTIONS
			//
			case ev.Key() == tcell.KeyCtrlZ:
				state.SetPrevActiveConnection()
			case ev.Key() == tcell.KeyCtrlX:
				state.SetNextActiveConnection()

			//
			// MOVEMENT BETWEEN ITEMS IN THE FUZZY PICKER
			//
			case state.FuzzyPicker.Visible && ev.Key() == tcell.KeyCtrlJ:
				if state.fuzzyPickerSelectedItem > 0 {
					state.fuzzyPickerSelectedItem -= 1
				}
			case state.FuzzyPicker.Visible && ev.Key() == tcell.KeyCtrlK:
				state.fuzzyPickerSelectedItem += 1

			//
			// COMMAND BAR
			//

			case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyEnter:
				command := string(state.Command)

        if state.FuzzyPicker.Visible {
          state.FuzzyPicker.OnSelected(state)
        } else if state.Mode == "writ" {
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
        }

				// Clear the command that was typed, and move back to chat mode.
				state.Command = []rune{}
				state.CommandCursorPosition = 0
				state.Mode = "chat"
        state.FuzzyPicker.Hide()

			// As characters are typed, add to the message.
			case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyRune:
				state.Command = append(
					append(state.Command[:state.CommandCursorPosition], ev.Rune()),
					state.Command[state.CommandCursorPosition:]...,
				)
				state.CommandCursorPosition += 1

        // Also, take care of autocomplete of slash commands
        // As the user types, show them above the command bar.
        if !state.FuzzyPicker.Visible && state.Command[0] == '/' {
          // Show the fuzzy picker
          state.FuzzyPicker.Show(func(state *State) {
            log.Println("Pressed enter on slash command fuzzy picker!")
          })

          // Assemble add the items to the fuzzy sorter.
          for _, command := range commands {
            state.FuzzyPicker.Items = append(state.FuzzyPicker.Items, command)
            state.FuzzyPicker.StringItems = append(
              state.FuzzyPicker.StringItems,
              fmt.Sprintf(
                "%s %s\t%s - %s", // ie: "/quit (/q)        Quit - quits slime"
                strings.Join(command.Permutations, " "),
                command.Arguments,
                command.Name,
                command.Description,
              ),
            )
          }
        }

			// Backspace removes a character.
			case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyDEL:
				if state.CommandCursorPosition > 0 {
					state.Command = append(
						state.Command[:state.CommandCursorPosition-1],
						state.Command[state.CommandCursorPosition:]...,
					)
					state.CommandCursorPosition -= 1
        } else {
          // Backspacing in an empty command box brings the user back to chat mode
          state.Mode = "chat"
        }

        // If the command box becomes empty, hide the fuzzy picker.
        if len(state.Command) == 0 {
          state.FuzzyPicker.Hide()
        }

			// Arrows right and left move the cursor
			case (state.Mode == "writ" || state.Mode == "pick") && (ev.Key() == tcell.KeyLeft || ev.Key() == tcell.KeyCtrlH):
				if state.CommandCursorPosition >= 1 {
					state.CommandCursorPosition -= 1
				}
			case (state.Mode == "writ" || state.Mode == "pick") && (ev.Key() == tcell.KeyRight || ev.Key() == tcell.KeyCtrlL):
				if state.CommandCursorPosition < len(state.Command) {
					state.CommandCursorPosition += 1
				}

			//
			// EDITING OPERATIONS
			//

			// Ctrl+w deletes a word.
			case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyCtrlW:
				lastSpaceIndex := 0
				for index := state.CommandCursorPosition - 1; index >= 0; index-- {
					if state.Command[index] == ' ' {
						lastSpaceIndex = index
						break
					}
				}

				state.Command = append(state.Command[:lastSpaceIndex], state.Command[state.CommandCursorPosition:]...)
				state.CommandCursorPosition = lastSpaceIndex

			// Ctrl+A / Ctrl+E go to the start and end of editing
			case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyCtrlA:
				state.CommandCursorPosition = 0
			case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyCtrlE:
				state.CommandCursorPosition = len(state.Command)
			}
		case *tcell.EventResize:
			screen.Sync()
		}

		// Render after each loop
		render(state, term)
	}
}
