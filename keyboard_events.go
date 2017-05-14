package main

import (
	"fmt"
	"log"
	"strings"
	"errors"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
)

// FIXME: This unit is in messages, it should be in rows. The problem is that 1 message isn't always
// 1 row.
const messageScrollPadding = 7

// When the user presses a key, send a message telling slack that the user is typing.
// Never send more typing events if the outgoing channel is full.
func sendTypingIndicator(state *State) error {
	if state.ActiveConnection() != nil && state.ActiveConnection().SelectedChannel() != nil {
		outgoing := state.ActiveConnection().Outgoing()
		if len(outgoing) < cap(outgoing) {
			state.ActiveConnection().Outgoing() <- gateway.Event{
				Type: "typing",
				Data: map[string]interface{}{
					"channel": state.ActiveConnection().SelectedChannel().Id,
				},
			}
		} else {
			return errors.New("No room in outgoing channel to send typing event!")
		}
	}
	return nil
}

// When the user presses ':' or '/', enable the autocomplete menu.
func enableCommandAutocompletion(state *State, quit chan struct{}) {
	// Also, take care of autocomplete of slash commands
	// As the user types, show them above the command bar in a fuzzy picker.
	if !state.FuzzyPicker.Visible {
		// When the user presses enter, run the slash command the user typed.
		state.FuzzyPicker.Show(func(state *State) {
			err := OnCommandExecuted(state, quit)
			if err != nil {
				log.Fatalf(err.Error())
			}
		})

		// Assemble add the items to the fuzzy sorter.
		for _, command := range COMMANDS {
			if len(command.Permutations) > 0 { // Only autocomplete commands that have slash commands
				state.FuzzyPicker.Items = append(state.FuzzyPicker.Items, command)
				state.FuzzyPicker.StringItems = append(
					state.FuzzyPicker.StringItems,
					fmt.Sprintf(
						"%s%s %s\t%s - %s", // ie: "/quit (/q)        Quit - quits slime"
						string(state.Command[0]),
						strings.Join(command.Permutations, " "),
						command.Arguments,
						command.Name,
						command.Description,
					),
				)
			}
		}
	}
}

// WHen a user presses a key when they are selecting with a message, perform an action.
func OnMessageInteraction(state *State, key rune) {
	// Is a message selected?
	if state.SelectedMessageIndex >= 0 {
		switch key {
		case 'o':
			err := GetCommand("OpenFile").Handler([]string{}, state)
			if err != nil {
				state.Status.Errorf(err.Error())
			}
		case 'c':
			err := GetCommand("CopyFile").Handler([]string{}, state)
			if err != nil {
				state.Status.Errorf(err.Error())
			}
		}
	} else {
		state.Status.Printf("No message selected.")
	}
}

// When a user picks a connection / channel in the fuzzy picker
func OnPickConnectionChannel(state *State) {
	// Assert that the fuzzy picker that's active is of the right type
	if selectedItem, ok := state.FuzzyPicker.Items[state.FuzzyPicker.SelectedItem].(FuzzyPickerConnectionChannelItem); ok {
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
		log.Fatalf("In pick mode, the fuzzy picker doesn't contain FuzzyPickerConnectionChannelItem's.")
	}
}

// Given a string, create an argv array of its parts. If a part is quoted, it's all part of the same
// argument.
// `a b c d` => []string{"a", "b", "c", "d"}
// `a "b c" d` => []string{"a", "b c", "d"}
// `a \"b c\" d` => []string{`a`, `"b`, `c"`, `d`}
func CreateArgvFromString(input string) []string {
	argv := []string{""}
	argvLastIndex := 0
	insideQuotes := false
	lastItem := ' '

	for _, item := range input {
		if item == '"' && lastItem != '\\' { // Handle an unescaped quote
			insideQuotes = !insideQuotes
		} else if item == ' ' && !insideQuotes { // Handle an unquoted space
			// A space creates a new argument
			argvLastIndex += 1
			argv = append(argv, "")
		} else {
			// Add the character to the last argv item, nothing special here...
			argv[argvLastIndex] += string(item)
		}
		lastItem = item
	}

	return argv
}

// When the user presses enter in `writ` mode after typing some stuff...
func OnCommandExecuted(state *State, quit chan struct{}) error {
	// Parse the command and create a list of arguments
	args := CreateArgvFromString(string(state.Command))

	// If the command was empty, return
	if len(args) == 0 {
		return nil
	}

	// Remove the first charater (slash or colon) from the command.
	arg0 := args[0][1:]

	if arg0 == "quit" || arg0 == "q" {
		// :q or :quit closes the app, and is a special case.
		log.Println("CLOSE QUIT 2")
		close(quit)
		return nil
	} else {
		// Otherwise, find the command that the user typed.
		for _, command := range COMMANDS {
			for _, permutation := range command.Permutations {
				if permutation == arg0 {
					err := RunCommand(command, args, state)
					if err != nil {
						state.Status.Errorf("Error in running command %s: %s", arg0, err.Error())
					}
					return nil
				}
			}
		}

		// If we haven't returned by now, then the command is invalid.
		state.Status.Errorf("Unknown command %s", args[0])
	}
	return nil
}

// Break out function to handle only keyboard events. Called by `keyboardEvents`.
func HandleKeyboardEvent(ev *tcell.EventKey, state *State, quit chan struct{}) error {
	// Prior to executing a keyboard command, clear the status.
	state.Status.Clear()

	// Did the user press a key in the keymap?
	if state.Mode == "chat" && ev.Key() == tcell.KeyRune {
		// Add pressed key to the stack of keys
		state.KeyStack = append(state.KeyStack, ev.Rune())

		// Did the user press the key combo?
		for _, key := range state.KeyActions {
			if string(key.Key) == string(state.KeyStack) {
				err := key.Handler(state)
				if err != nil {
					state.Status.Errorf(err.Error())
				}
				state.KeyStack = []rune{}
			}
		}
	}

	switch {
	case ev.Key() == tcell.KeyCtrlC:
		log.Println("CLOSE QUIT 1")
		close(quit)
		return nil

	// Escape reverts back to chat mode and clears the key stack.
	case ev.Key() == tcell.KeyEscape:
		state.Mode = "chat"
		state.FuzzyPicker.Hide()
		state.KeyStack = []rune{}

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
						"#%s %s",
						channel.Name,
						connection.Name(),
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

	// 'e' moves to write mode. So does ':' and '/'
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'w':
		state.Mode = "writ"
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == ':':
		state.Mode = "writ"
		state.Command = []rune{':'}
		state.CommandCursorPosition = 1
		enableCommandAutocompletion(state, quit)
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == '/':
		state.Mode = "writ"
		state.Command = []rune{'/'}
		state.CommandCursorPosition = 1
		enableCommandAutocompletion(state, quit)


	//
	// MOVEMENT UP AND DOWN THROUGH MESSAGES AND ACTIONS ON THE MESSAGES
	//
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'j': // Down a message
		err := GetCommand("MoveBackMessage").Handler([]string{}, state)
		if err != nil {
			state.Status.Errorf(err.Error())
		}
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'k': // Up a message
		err := GetCommand("MoveForwardMessage").Handler([]string{}, state)
		if err != nil {
			state.Status.Errorf(err.Error())
		}
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'G': // Select first message
		if state.ActiveConnection() != nil && len(state.ActiveConnection().MessageHistory()) > 0{
			state.SelectedMessageIndex = 0
			state.BottomDisplayedItem = 0
			log.Printf("Selecting first message")
		} else {
			state.Status.Errorf("No active connection or message history!")
		}
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'g': // Select last message loaded
		if state.ActiveConnection() != nil && len(state.ActiveConnection().MessageHistory()) > 0{
			state.SelectedMessageIndex = len(state.ActiveConnection().MessageHistory()) - 1
			state.BottomDisplayedItem = state.SelectedMessageIndex - messageScrollPadding
			log.Printf("Selecting last message")
		} else {
			state.Status.Errorf("No active connection or message history!")
		}
	case state.Mode == "chat" && ev.Key() == tcell.KeyCtrlU: // Up a message page
		pageAmount := state.RenderedMessageNumber / 2
		if state.ActiveConnection() != nil && state.SelectedMessageIndex < len(state.ActiveConnection().MessageHistory())-1 {
			state.SelectedMessageIndex += pageAmount
			state.BottomDisplayedItem += pageAmount
			log.Printf("Selecting message %d, bottom index %d", state.SelectedMessageIndex, state.BottomDisplayedItem)

			// Clamp BottomDisplayedItem at zero.
			largestMessageIndex := len(state.ActiveConnection().MessageHistory()) - 1
			if state.BottomDisplayedItem > largestMessageIndex {
				state.BottomDisplayedItem = largestMessageIndex
			}
			// Clamp SelectedMessageIndex at zero.
			if state.SelectedMessageIndex > largestMessageIndex {
				state.SelectedMessageIndex = largestMessageIndex
			}
		} else {
			state.Status.Errorf("No active connection, or message history too short!")
		}
	case state.Mode == "chat" && ev.Key() == tcell.KeyCtrlD: // Down a message page
		pageAmount := state.RenderedMessageNumber / 2
		if state.ActiveConnection() != nil && state.SelectedMessageIndex > 0 {
			state.SelectedMessageIndex -= pageAmount
			state.BottomDisplayedItem -= pageAmount

			// Clamp BottomDisplayedItem at zero.
			if state.BottomDisplayedItem < 0 {
				state.BottomDisplayedItem = 0
			}
			// Clamp SelectedMessageIndex at zero.
			if state.SelectedMessageIndex < 0 {
				state.SelectedMessageIndex = 0
			}
			log.Printf("Selecting message %d, bottom index %d", state.SelectedMessageIndex, state.BottomDisplayedItem)
		} else {
			state.Status.Errorf("No active connection, or message history too short!")
		}
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && (ev.Rune() == 'o' || ev.Rune() == 'c'):
		// When a user presses a key to interact with a message, handle it.
		OnMessageInteraction(state, ev.Rune())
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'Z':
		// Center the selected message
		if state.ActiveConnection() != nil {
			state.BottomDisplayedItem = state.SelectedMessageIndex - (state.RenderedMessageNumber / 4)

			// Clamp BottomDisplayedItem at zero.
			if state.BottomDisplayedItem < 0 {
				state.BottomDisplayedItem = 0
			}
		} else {
			state.Status.Errorf("No active connection, or message history too short!")
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
		if state.FuzzyPicker.SelectedItem > 0 {
			state.FuzzyPicker.SelectedItem -= 1
			// If we select an item off the screen, show it on the screen by changing the bottommost
			// item.
			if state.FuzzyPicker.SelectedItem < state.FuzzyPicker.BottomItem {
				state.FuzzyPicker.BottomItem -= 1
			}
		}
	case state.FuzzyPicker.Visible && ev.Key() == tcell.KeyCtrlK:
		topDisplayedItem := state.FuzzyPicker.BottomItem + frontend.FuzzyPickerMaxSize - 1
		if state.FuzzyPicker.SelectedItem < len(state.FuzzyPicker.Items)-1 {
			state.FuzzyPicker.SelectedItem += 1
			// If we select an item off the screen, show it on the screen by changing the bottommost
			// item.
			if state.FuzzyPicker.SelectedItem > topDisplayedItem {
				state.FuzzyPicker.BottomItem += 1
			}
		}

	//
	// COMMAND BAR
	//

	case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyEnter:
		log.Println("Enter pressed")
		if state.FuzzyPicker.Visible {
			state.FuzzyPicker.OnSelected(state)
		} else if state.Mode == "writ" && state.ActiveConnection() != nil {
			// Just send a normal message!
			message := gateway.Message{
				Sender: state.ActiveConnection().Self(),
				Text:   string(state.Command),
			}

			// Sometimes, a message could have a response. This is for example true in the
			// case of slash commands, sometimes.
			responseMessage, err := state.ActiveConnection().SendMessage(
				message,
				state.ActiveConnection().SelectedChannel(),
			)

			if err != nil {
				return err
			} else if responseMessage != nil {
				// Got a response command? Append it to the message history.
				state.ActiveConnection().AppendMessageHistory(*responseMessage)
			}
		}

		// Clear the command that was typed, and move back to chat mode. Also hide the fuzzy picker
		// is its open.
		state.Command = []rune{}
		state.CommandCursorPosition = 0
		state.Mode = "chat"
		state.FuzzyPicker.Hide()

	//
	// EDITING OPERATIONS
	//

	// As characters are typed, add to the message.
	case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyRune:
		state.Command = append(
			append(state.Command[:state.CommandCursorPosition], ev.Rune()),
			state.Command[state.CommandCursorPosition:]...,
		)
		state.CommandCursorPosition += 1

		// Send a message on the outgoing channel that the user is typing.
		err := sendTypingIndicator(state)
		if err != nil {
			state.Status.Errorf(err.Error())
		}

	// Backspace removes a character.
	case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyDEL:
		if state.CommandCursorPosition > 0 {
			state.Command = append(
				state.Command[:state.CommandCursorPosition-1],
				state.Command[state.CommandCursorPosition:]...,
			)
			state.CommandCursorPosition -= 1
			// Send a message on the outgoing channel that the user is typing.
			sendTypingIndicator(state)
		} else {
			// Backspacing in an empty command box brings the user back to chat mode
			state.Mode = "chat"
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

	return nil
}

func keyboardEvents(state *State, term *frontend.TerminalDisplay, screen tcell.Screen, quit chan struct{}) {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			log.Printf("Keypress: %+v", ev.Name())

			// CTRL + L redraws the screen.
			if state.Mode == "chat" && ev.Key() == tcell.KeyCtrlL {
				screen.Sync()
			} else {
				err := HandleKeyboardEvent(ev, state, quit)
				if err != nil {
					log.Fatalf(err.Error())
				}
			}
		case *tcell.EventResize:
			screen.Sync()
		}

		// Render after each loop
		render(state, term)
	}
}
