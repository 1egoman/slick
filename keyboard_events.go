package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/1egoman/slick/frontend" // The thing to draw to the screen
	"github.com/1egoman/slick/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji" // convert :smile: to unicode
)

// FIXME: This unit is in messages, it should be in rows. The problem is that 1 message isn't always
// 1 row.
const messageScrollPadding = 7

// When the user presses a key, send a message telling slack that the user is typing.
// Never send more typing events if the outgoing channel is full.
func sendTypingIndicator(state *State) error {
	if state.ActiveConnection() != nil && state.ActiveConnection().SelectedChannel() != nil {
		outgoing := state.ActiveConnection().Outgoing()
		if len(outgoing) < cap(outgoing)/2 { // Only send typing indicator if we have plenty of room in the queue.
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
func enableCommandAutocompletion(state *State, term *frontend.TerminalDisplay, quit chan struct{}) {
	// Also, take care of autocomplete of slash commands
	// As the user types, show them above the command bar in a fuzzy picker.
	if !state.FuzzyPicker.Visible {
		// When the user presses enter, run the slash command the user typed.
		state.FuzzyPicker.Hide()
		state.FuzzyPicker.Show(func(state *State) {
			err := OnCommandExecuted(state, term, quit)
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
						"%s%s %s\t%s - %s", // ie: "/quit (/q)        Quit - quits slick"
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

// Given a keystack, extract the preceding quantity.
func KeystackQuantityParser(keystack []rune) (int, []rune, error) {
	quantityRunes := []rune{}

	// Fetch the preceding int before the command
	for _, key := range keystack {
		if key >= '0' && key <= '9' {
			quantityRunes = append(quantityRunes, key)
		} else {
			break
		}
		keystack = keystack[1:] // Remove a rune of the quantity from the front of the keystack
	}

	// Convert the quantity to an int
	if len(quantityRunes) == 0 {
		return 1, keystack, nil
	} else {
		quantity, err := strconv.Atoi(string(quantityRunes))
		return quantity, keystack, err
	}
}

// WHen a user presses a key when they are selecting with a message, perform an action.
func OnMessageInteraction(state *State, key rune, quantity int) {
	// Is a message selected?
	if state.SelectedMessageIndex >= 0 {
		switch key {
		case 'o': // Open a file
			err := GetCommand("OpenFile").Handler([]string{}, state)
			if err != nil {
				state.Status.Errorf(err.Error())
			}
		case 'c': // Copy the link to a file
			err := GetCommand("CopyFile").Handler([]string{}, state)
			if err != nil {
				state.Status.Errorf(err.Error())
			}
		case 'l': // Open link in attachment
			err := GetCommand("OpenAttachmentLink").Handler(
				[]string{"__INTERNAL__", fmt.Sprintf("%d", quantity)},
				state,
			)
			if err != nil {
				state.Status.Errorf(err.Error())
			}
		case 'm': // Open link in message
			err := GetCommand("OpenMessageLink").Handler(
				[]string{"__INTERNAL__", fmt.Sprintf("%d", quantity)},
				state,
			)
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
				"Tried to select channel %s that isn't in the slice of channels for connection %s",
				selectedChannelName,
				selectedConnectionName,
			)
		}

		log.Printf("Selecting connection %s and channel %s", selectedConnectionName, selectedChannel.Name)

		// Set the active connection with the discovered index, and also set a new selected
		// channel.
		state.SetActiveConnection(selectedConnectionIndex)
		state.Connections[selectedConnectionIndex].SetSelectedChannel(selectedChannel)
		EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "chat"})
		state.Mode = "chat"
		state.SelectedMessageIndex = 0
		state.BottomDisplayedItem = 0
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
func OnCommandExecuted(state *State, term *frontend.TerminalDisplay, quit chan struct{}) error {
	// Parse the command and create a list of arguments
	args := CreateArgvFromString(string(state.Command))

	// If the command was empty, return
	if len(args) == 0 {
		return nil
	}

	// Remove the first charater (slash or colon) from the command.
	arg0 := args[0][1:]

	// Emit event to to be handled by lua scripts
	EmitEvent(state, EVENT_COMMAND_RUN, map[string]string{
		"raw":     string(state.Command),
		"command": arg0,
	})

	// SPECIAL CASES
	// Since these commands need access to "privileged" things, they are harded here.
	// `quit` - needs to be able to close the `quit` channel
	// `require` - needs `term` to pass to `ParseScript`.
	if arg0 == "quit" || arg0 == "q" {
		// :q or :quit closes the app, and is a special case.
		log.Println("CLOSE QUIT 2")
		close(quit)
		return nil
	} else if arg0 == "require" || arg0 == "r" || arg0 == "source" {
		var luaPath string
		if len(args) == 2 { // /require foo.lua
			luaPath = args[1]
		} else {
			state.Status.Errorf("Please use more arguments. /require foo.lua")
			return nil
		}

		// Read post from filesystem
		luaContents, err := ioutil.ReadFile(luaPath)
		if err != nil {
			state.Status.Errorf("Couldn't readfile %s: %s", luaPath, err.Error())
			return nil
		}

		err = ParseScript(string(luaContents), state, term)
		if err != nil {
			state.Status.Errorf("lua error: %s", err.Error())
		} else {
			return nil
		}
	} else {
		// Otherwise, find the command that the user typed.
		for _, command := range COMMANDS {
			for _, permutation := range command.Permutations {
				if permutation == arg0 {
					err := RunCommand(command, args, state)
					if err != nil {
						state.Status.Errorf("Error in running command %s: %s", arg0, err.Error())
						render(state, term)
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

// After the user runs a command, reset the key stack.
func resetKeyStack(state *State) {
	state.KeyStack = []rune{}
}

// Fetch more messages when the user has scrolled to the end of the previous message list.
func FetchMessageHistoryScrollback(state *State) error {
	msgHistory := state.ActiveConnection().MessageHistory()
	messages, err := state.ActiveConnection().FetchChannelMessages(
		*state.ActiveConnection().SelectedChannel(), // Channel
		&(msgHistory[0].Hash),                       // *string
	)

	if err != nil {
		return err
	}

	for msgIndex := len(messages) - 1; msgIndex >= 0; msgIndex-- {
		state.ActiveConnection().PrependMessageHistory(messages[msgIndex])
	}

	return nil
}

// Break out function to handle only keyboard events. Called by `keyboardEvents`.
func HandleKeyboardEvent(ev *tcell.EventKey, state *State, term *frontend.TerminalDisplay, quit chan struct{}) error {
	// Did the user press a key in the keymap?
	if state.Mode == "chat" && ev.Key() == tcell.KeyRune {
		// Add pressed key to the stack of keys
		state.KeyStack = append(state.KeyStack, ev.Rune())

		// Did the user press the key combo?
		for _, action := range state.EventActions {
			if action.Type == EVENT_KEYMAP && string(action.Key) == string(state.KeyStack) {
				err := action.Handler(state, nil)
				if err != nil {
					state.Status.Errorf(err.Error())
				}
				resetKeyStack(state)
			}
		}
	}

	quantity, keystackCommand, _ := KeystackQuantityParser(state.KeyStack)
	log.Printf("Keystack: %v", state.KeyStack)
	switch {
	case ev.Key() == tcell.KeyCtrlC:
		log.Println("CLOSE QUIT 1")
		close(quit)
		return nil

	// Escape reverts back to chat mode and clears the key stack.
	case ev.Key() == tcell.KeyEscape:
		EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "chat"})
		state.Mode = "chat"
		state.FuzzyPicker.Hide()
		resetKeyStack(state)
		state.Status.Clear()

	// 'p' moves to a channel picker, which is a mode for switching teams and channels
	case state.Mode == "chat" && len(keystackCommand) == 1 && keystackCommand[0] == 'p':
		if state.Mode != "pick" {
			EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "pick"})
			state.Mode = "pick"
			state.FuzzyPicker.Hide()
			state.FuzzyPicker.Show(OnPickConnectionChannel)

			var items []interface{}
			stringItems := []string{}
			var accessories string

			// Accumulate all channels into `items`, and their respective labels into `stringLabels`
			for _, connection := range state.Connections {
				for _, channel := range connection.Channels() {
					accessories = ""
					if channel.IsArchived {
						accessories += "(archived) "
					}
					if !channel.IsMember {
						accessories += "(not a member) "
					}

					// Add string representation of item to `stringItems`
					// Follows the pattern of "my-team #my-channel"
					stringItems = append(stringItems, fmt.Sprintf(
						"#%s %s\t%s",
						channel.Name,
						connection.Name(),
						accessories,
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
			EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "chat"})
			state.Mode = "chat"
			state.FuzzyPicker.Hide()
		}
		resetKeyStack(state)

	// 'e' moves to write mode. So does ':' and '/'
	case state.Mode == "chat" && len(keystackCommand) == 1 && keystackCommand[0] == 'w':
		EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "writ"})
		state.Mode = "writ"
		resetKeyStack(state)
	case state.Mode == "chat" && len(keystackCommand) == 1 && keystackCommand[0] == ':':
		EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "writ"})
		state.Mode = "writ"
		state.Command = []rune{':'}
		state.CommandCursorPosition = 1
		enableCommandAutocompletion(state, term, quit)
		resetKeyStack(state)
	case state.Mode == "chat" && len(keystackCommand) == 1 && keystackCommand[0] == '/':
		EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "writ"})
		state.Mode = "writ"
		state.Command = []rune{'/'}
		state.CommandCursorPosition = 1
		enableCommandAutocompletion(state, term, quit)
		resetKeyStack(state)

	//
	// TAB-COMPLETE FOR FILE PATHS
	//
	case state.Mode == "writ" && ev.Key() == tcell.KeyTab && len(state.Command) > 0 && state.CommandCursorPosition > 0 && state.Command[state.CommandCursorPosition-1] == '/':
		EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "pick"})
		state.Mode = "pick"

		// Open the fuzzy picker
		state.FuzzyPicker.Hide()
		state.FuzzyPicker.Show(func(state *State) {
			render(state, term)
		})
		state.FuzzyPicker.Resort(func(state *State) {
			// If the command moves before the starting point, hide the fuzzy picker.
			if len(state.Command) <= state.FuzzyPicker.ThrowAwayPrefix-1 {
				log.Println("User moved into already chosen path, aborting...")
				state.FuzzyPicker.Hide()
				EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "writ"})
				state.Mode = "writ"
				return
			}

			// Once we get to a new path segment, update with new values.
			if len(state.Command)-1 > state.FuzzyPicker.ThrowAwayPrefix-1 && state.Command[len(state.Command)-1] == '/' {
				// Clear items, and recalculate
				state.FuzzyPicker.Items = []interface{}{}
				state.FuzzyPicker.StringItems = []string{}
				state.FuzzyPicker.SelectedItem = 0
				state.FuzzyPicker.BottomItem = 0
				pathCommand := string(state.Command)
				for {
					beginningOfPath := strings.LastIndex(pathCommand, " ")
					// Make sure the space isn't at the end of the command
					if beginningOfPath == len(pathCommand) {
						continue
					}
					// Make sure that we didn't go through all spaces in the command
					// If we did, then the slash isn't in a good spot. Cancel the fuzzy picker.
					if beginningOfPath == -1 {
						state.FuzzyPicker.Hide()
						EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "writ"})
						state.Mode = "writ"
						break
					}
					// After the space, is there a slash?
					if index := strings.Index(pathCommand[beginningOfPath+1:], "/"); index > 1 {
						pathCommand = pathCommand[:beginningOfPath] // If not, try again from before the indexed space
						continue
					}

					// At this point, all tests have been passed. Get the path and use it for the rest of the calculations
					if beginningOfPath == -1 {
						// If no space, then the path is at the start of the phrase.
						beginningOfPath = 0
					} else {
						// Otherwise, start on the slash after the space.
						beginningOfPath += 1
					}

					path := pathCommand[beginningOfPath:]

					// Does the path have a tilda before it? If so, replace with $HOME.
					if len(path) > 0 && path[0] == '~' {
						path = os.Getenv("HOME") + path[1:]
					}

					state.FuzzyPicker.ThrowAwayPrefix = beginningOfPath + 1

					// Construct a list of items to show in the fuzzy picker
					files, err := ioutil.ReadDir(path)
					if err != nil {
						state.Status.Errorf("Error fetching path items: %s", err.Error())
						state.FuzzyPicker.Hide()
						EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "writ"})
						state.Mode = "writ"
						return
					}

					for _, file := range files {
						state.FuzzyPicker.Items = append(state.FuzzyPicker.Items, file.Name())

						var displayName string
						if file.IsDir() {
							displayName = file.Name() + "/" // (directories end in a slash)
						} else {
							displayName = file.Name()
						}
						state.FuzzyPicker.StringItems = append(
							state.FuzzyPicker.StringItems,
							displayName,
						)
					}

					state.FuzzyPicker.ThrowAwayPrefix = state.CommandCursorPosition
					log.Printf("Got contents of new directory %s", string(state.Command[state.FuzzyPicker.ThrowAwayPrefix:]))
					break
				}
			}

			// User just typed a space? Then close the fuzzy picker.
			if state.Command[len(state.Command)-1] == ' ' {
				state.FuzzyPicker.Hide()
				EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "writ"})
				state.Mode = "writ"
			}
		})

	//
	// MOVEMENT UP AND DOWN THROUGH MESSAGES AND ACTIONS ON THE MESSAGES
	//
	// `j` moves down a message.
	case state.Mode == "chat" && len(keystackCommand) == 1 && keystackCommand[0] == 'j':
		for i := 0; i < quantity; i++ {
			err := GetCommand("MoveBackMessage").Handler([]string{}, state)
			if err != nil {
				state.Status.Errorf(err.Error())
			}
		}
		resetKeyStack(state)

	// `k` moves up a message.
	case state.Mode == "chat" && len(keystackCommand) == 1 && keystackCommand[0] == 'k': // Up a message
		for i := 0; i < quantity; i++ {
			err := GetCommand("MoveForwardMessage").Handler([]string{}, state)
			if err != nil {
				state.Status.Errorf(err.Error())
			}
		}
		resetKeyStack(state)

	// `G` will go to the bottom (newest) of the message history
	case state.Mode == "chat" && len(keystackCommand) == 1 && keystackCommand[0] == 'G': // Select first message
		if state.ActiveConnection() != nil && len(state.ActiveConnection().MessageHistory()) > 0 {
			state.SelectedMessageIndex = 0
			state.BottomDisplayedItem = 0
			log.Printf("Selecting first message")
		} else {
			state.Status.Errorf("No active connection or message history!")
		}
		resetKeyStack(state)

	// `gg` will go to the top (oldest) of the message history
	case state.Mode == "chat" && len(keystackCommand) == 2 && string(keystackCommand) == "gg":
		if state.ActiveConnection() != nil && len(state.ActiveConnection().MessageHistory()) > 0 {
			state.SelectedMessageIndex = len(state.ActiveConnection().MessageHistory()) - 1
			state.BottomDisplayedItem = state.SelectedMessageIndex - messageScrollPadding
			log.Printf("Selecting last message")

			// Now that we're at the top, fetch more messages.
			msgHistory := state.ActiveConnection().MessageHistory()
			log.Printf("Last message loaded: %s", msgHistory[0].Hash)
		} else {
			state.Status.Errorf("No active connection or message history!")
		}
		resetKeyStack(state)

	// `zz` will center the viewport on a message:
	case state.Mode == "chat" && len(keystackCommand) == 2 && string(keystackCommand) == "zz": // Center on a mezzage
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
		resetKeyStack(state)

	case state.Mode == "chat" && ev.Key() == tcell.KeyCtrlU: // Up a message page
		pageAmount, err := strconv.Atoi(state.Configuration["Message.PageAmount"])
		if err != nil {
			state.Status.Errorf("Cannot parse Message.PageAmount as int: %s", state.Configuration["Message.PageAmount"])
			return nil
		}

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
		resetKeyStack(state)
	case state.Mode == "chat" && ev.Key() == tcell.KeyCtrlD: // Down a message page
		pageAmount, err := strconv.Atoi(state.Configuration["Message.PageAmount"])
		if err != nil {
			state.Status.Errorf("Cannot parse Message.PageAmount as int: %s", state.Configuration["Message.PageAmount"])
			return nil
		}

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
		resetKeyStack(state)
	case state.Mode == "chat" && (string(keystackCommand) == "o" ||
		string(keystackCommand) == "c" ||
		string(keystackCommand) == "l" ||
		string(keystackCommand) == "m"): // Message interaction
		// When a user presses a key to interact with a message, handle it.
		OnMessageInteraction(state, keystackCommand[0], quantity)
		resetKeyStack(state)

	//
	// MOVEMENT BETWEEN CONNECTIONS
	//
	case ev.Key() == tcell.KeyCtrlZ:
		state.SetPrevActiveConnection()
	case ev.Key() == tcell.KeyCtrlX:
		state.SetNextActiveConnection()
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() > '0' && ev.Rune() <= '9':
		index := int(ev.Rune() - '1')
		log.Printf("Switch to connection index %d", index)
		if index < len(state.Connections) {
			state.SetActiveConnection(index)
		} else {
			state.Status.Errorf("No such connection with index %d", index)
		}

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

			// If the command starts with a slash or colon, then run it.
		} else if state.Command[0] == '/' ||
			// Make sure the command doesn't start with :emoji: - Fixes #18.
			emoji.Sprint(state.Command)[0] == ':' {
			err := OnCommandExecuted(state, term, quit)
			if err != nil {
				log.Fatalf(err.Error())
			}

			// Otherwise, send as a message.
		} else if state.Mode == "writ" && state.ActiveConnection() != nil {
			// Emit event to to be handled by lua scripts
			EmitEvent(state, EVENT_MESSAGE_SENT, map[string]string{
				"sender": state.ActiveConnection().Self().Name,
				"text":   string(state.Command),
			})

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
		EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "chat"})
		state.Mode = "chat"
		state.FuzzyPicker.Hide()

	case state.Mode == "pick" && state.FuzzyPicker.Visible && len(state.FuzzyPicker.StringItems) > 0 && ev.Key() == tcell.KeyTab:
		// Pressing tab when in the fuzzy picker takes the displayed item and updates the command
		// bar with its contents
		displayItem := state.FuzzyPicker.StringItems[state.FuzzyPicker.SelectedItem]
		state.Command = append(state.Command[:state.FuzzyPicker.ThrowAwayPrefix], []rune(displayItem)...)
		state.CommandCursorPosition = len(state.Command)

	//
	// EDITING OPERATIONS
	//

	// Backslash adds a newline to a message.
	case state.Mode == "writ" && ev.Key() == tcell.KeyRune && ev.Rune() == '\\':
		state.Command = append(
			append(state.Command[:state.CommandCursorPosition], '\n'),
			state.Command[state.CommandCursorPosition:]...,
		)
		state.CommandCursorPosition += 1

	// As characters are typed, add to the message.
	case (state.Mode == "writ" || state.Mode == "pick") && ev.Key() == tcell.KeyRune:
		state.Command = append(
			append(state.Command[:state.CommandCursorPosition], ev.Rune()),
			state.Command[state.CommandCursorPosition:]...,
		)
		state.CommandCursorPosition += 1

		// Send a message on the outgoing channel that the user is typing.
		// (Only send events when the user is typing a message, not when they try to send a command)
		if state.Mode == "writ" && !state.FuzzyPicker.Visible {
			err := sendTypingIndicator(state)
			if err != nil {
				state.Status.Errorf(err.Error())
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
			// Send a message on the outgoing channel that the user is typing.
			sendTypingIndicator(state)
		} else {
			// Backspacing in an empty command box brings the user back to chat mode
			EmitEvent(state, EVENT_MODE_CHANGE, map[string]string{"from": state.Mode, "to": "chat"})
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

	// If the user has scrolled to the end of the list of messages in their active channel, then load more
	if state.ActiveConnection() != nil &&
		state.SelectedMessageIndex > len(state.ActiveConnection().MessageHistory())-1-messageScrollPadding &&
		len(state.ActiveConnection().MessageHistory()) > 0 {
		go func(state *State) {
			err := FetchMessageHistoryScrollback(state)
			if err != nil {
				state.Status.Errorf("Error fetching more messages: %s", err)
			}
		}(state)
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
				err := HandleKeyboardEvent(ev, state, term, quit)
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
