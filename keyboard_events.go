package main

import (
	"fmt"
	"log"
	"strings"
	"io/ioutil"
	"errors"

	"github.com/1egoman/slime/frontend" // The thing to draw to the screen
	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
	"github.com/gdamore/tcell"
	"github.com/skratchdot/open-golang/open"
)

// FIXME: This unit is in messages, it should be in rows. The problem is that 1 message isn't always
// 1 row.
const messageScrollPadding = 7

type CommandType int

const (
	NATIVE CommandType = iota
	SLACK
)

var commands = []FuzzyPickerSlashCommandItem{
	{
		Type:         NATIVE,
		Name:         "Quit",
		Description:  "Quits slime.",
		Permutations: []string{"/quit", "/q"},
	},
	{
		Type:         NATIVE,
		Name:         "Post",
		Description:  "Make a post in the current channel.",
		Arguments:    "<post file> [post name]",
		Permutations: []string{"/post"},
	},
	{
		Type:         SLACK,
		Name:         "Apps",
		Permutations: []string{"/apps"},
		Arguments:    "[search term]",
		Description:  "Search for Slack Apps in the App Directory",
	},
	{
		Type:         SLACK,
		Name:         "Away",
		Permutations: []string{"/away"},
		Arguments:    "Toggle your away status",
	},
	{
		Type:         SLACK,
		Name:         "Call",
		Permutations: []string{"/call"},
		Arguments:    "[help]",
		Description:  "Start a call",
	},
	{
		Type:         SLACK,
		Name:         "Dnd",
		Permutations: []string{"/dnd"},
		Arguments:    "[some description of time]",
		Description:  "Starts or ends a Do Not Disturb session",
	},
	{
		Type:         SLACK,
		Name:         "Feed",
		Permutations: []string{"/feed"},
		Arguments:    "help [or subscribe, list, remove...]",
		Description:  "Manage RSS subscriptions",
	},
	{
		Type:         SLACK,
		Name:         "Invite",
		Permutations: []string{"/invite"},
		Arguments:    "@user [channel]",
		Description:  "Invite another member to a channel",
	},
	{
		Type:         SLACK,
		Name:         "Invite people",
		Permutations: []string{"/invite_people"},
		Arguments:    "[name@example.com, ...]",
		Description:  "Invite people to your Slack team",
	},
	{
		Type:         SLACK,
		Name:         "Leave",
		Permutations: []string{"/leave", "/close", "/part"},
		Description:  "Leave a channel",
	},
	{
		Type:         SLACK,
		Name:         "Me",
		Permutations: []string{"/me"},
		Arguments:    "your message",
		Description:  "Displays action text",
	},
	{
		Type:         SLACK,
		Name:         "Msg",
		Permutations: []string{"/msg", "/dm"},
		Arguments:    "[your message]",
	},
	{
		Type:         SLACK,
		Name:         "Mute",
		Permutations: []string{"/mute"},
		Arguments:    "[channel]",
		Description:  "Mutes [channel] or the current channel",
	},
	{
		Type:         SLACK,
		Name:         "Remind",
		Permutations: []string{"/remind"},
		Arguments:    "[@someone or #channel] [what] [when]",
		Description:  "Set a reminder",
	},
	{
		Type:         SLACK,
		Name:         "Rename",
		Permutations: []string{"/rename"},
		Arguments:    "[new name]",
		Description:  "Rename a channel",
	},
	{
		Type:         SLACK,
		Name:         "Shrug",
		Permutations: []string{"/shrug"},
		Arguments:    "[your message]",
		Description:  "Appends ¯\\_(ツ)_/¯ to your message",
	},
	{
		Type:         SLACK,
		Name:         "Star",
		Permutations: []string{"/star"},
		Arguments:    "Stars the current channel or conversation",
	},
	{
		Type:         SLACK,
		Name:         "Status",
		Permutations: []string{"/status"},
		Arguments:    "[clear] or [:your_new_status_emoji:] [your new status message]",
		Description:  "Set or clear your custom status",
	},
	{
		Type:         SLACK,
		Name:         "Who",
		Permutations: []string{"/who"},
		Description:  "List users in the currentl channel or group",
	},
}

// When the user presses a key, send a message telling slack that the user is typing.
func sendTypingIndicator(state *State) {
	if state.ActiveConnection().SelectedChannel() != nil {
		state.ActiveConnection().Outgoing() <- gateway.Event{
			Type: "typing",
			Data: map[string]interface{}{
				"channel": state.ActiveConnection().SelectedChannel().Id,
			},
		}
	}
}

// WHen a user presses a key when they are selecting with a message, perform an action.
func OnMessageInteraction(state *State, key rune) {
	// Is a message selected?
	if state.SelectedMessageIndex >= 0 {
		selectedMessageIndex := len(state.ActiveConnection().MessageHistory()) - 1 - state.SelectedMessageIndex
		selectedMessage := state.ActiveConnection().MessageHistory()[selectedMessageIndex]

		switch key {
		case 'o':
			// Open the private image url in the browser
			if selectedMessage.File != nil {
				open.Run(selectedMessage.File.Permalink)
			} else {
				log.Println("o pressed, but currently selected message has no file to Open")
			}
		case 'c':
			// Open the private image url in the browser
			if selectedMessage.File != nil {
				// FIXME: actually write to the clipboard
				log.Println("FIXME: will eventually copy to clipboard:", selectedMessage.File.Permalink)
			} else {
				log.Println("o pressed, but currently selected message has no file to Open")
			}
		}
	} else {
		log.Println("o pressed, but no message selected")
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
	command := string(state.Command)
	args := CreateArgvFromString(command)
	switch {
	// :q or :quit closes the app
	case command == ":q", command == ":quit":
		log.Println("CLOSE QUIT 2")
		close(quit)
		return nil

	case args[0] == ":post":
		if len(args) > 2 { // /post path/to/post.txt "post title"
			postPath := args[1]
			postTitle := args[2]
			postContent, err := ioutil.ReadFile(postPath)
			if err != nil {
				return err
			}
			if err = state.ActiveConnection().PostText(postTitle, string(postContent)); err != nil {
				return err
			}
		} else {
			return errors.New("Please use more arguments. /post path/to/post.txt \"post title\"")
		}

	// Unknown command!
	default:
		return errors.New(fmt.Sprintf("Unknown command %s", args[0]))
	}
	return nil
}

// Break out function to handle only keyboard events. Called by `keyboardEvents`.
func HandleKeyboardEvent(ev *tcell.EventKey, state *State, quit chan struct{}) error {
	switch {
	case ev.Key() == tcell.KeyCtrlC:
		log.Println("CLOSE QUIT 1")
		close(quit)
		return nil

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
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == '/':
		state.Mode = "writ"
		state.Command = []rune{'/'}
		state.CommandCursorPosition = 1


	//
	// MOVEMENT UP AND DOWN THROUGH MESSAGES AND ACTIONS ON THE MESSAGES
	//
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'j': // Down a message
		if state.SelectedMessageIndex > 0 {
			state.SelectedMessageIndex -= 1
			if state.BottomDisplayedItem > 0 && state.SelectedMessageIndex < state.BottomDisplayedItem+messageScrollPadding {
				state.BottomDisplayedItem -= 1
			}
			log.Printf("Selecting message %s", state.SelectedMessageIndex)
		}
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'k': // Up a message
		if state.SelectedMessageIndex < len(state.ActiveConnection().MessageHistory())-1 {
			state.SelectedMessageIndex += 1
			if state.SelectedMessageIndex >= state.RenderedMessageNumber-messageScrollPadding {
				state.BottomDisplayedItem += 1
			}
			log.Printf("Selecting message %d, bottom index %d", state.SelectedMessageIndex, state.BottomDisplayedItem)
		}
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'G': // Select first message
		state.SelectedMessageIndex = 0
		state.BottomDisplayedItem = 0
		log.Printf("Selecting first message")
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && ev.Rune() == 'g': // Select last message loaded
		state.SelectedMessageIndex = len(state.ActiveConnection().MessageHistory()) - 1
		state.BottomDisplayedItem = state.SelectedMessageIndex - messageScrollPadding
		log.Printf("Selecting first message")
	case state.Mode == "chat" && ev.Key() == tcell.KeyCtrlU: // Up a message page
		pageAmount := state.RenderedMessageNumber / 2
		if state.SelectedMessageIndex < len(state.ActiveConnection().MessageHistory())-1 {
			state.SelectedMessageIndex += pageAmount
			if state.SelectedMessageIndex >= state.RenderedMessageNumber-messageScrollPadding {
				state.BottomDisplayedItem += pageAmount
			}
			log.Printf("Selecting message %d, bottom index %d", state.SelectedMessageIndex, state.BottomDisplayedItem)
		}
	case state.Mode == "chat" && ev.Key() == tcell.KeyRune && (ev.Rune() == 'o' || ev.Rune() == 'c'):
		// When a user presses a key to interact with a message, handle it.
		OnMessageInteraction(state, ev.Rune())

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
		} else if state.Mode == "writ" {
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
		sendTypingIndicator(state)

		// Also, take care of autocomplete of slash commands
		// As the user types, show them above the command bar in a fuzzy picker.
		if !state.FuzzyPicker.Visible && (state.Command[0] == '/' || state.Command[0] == ':') {
			// When the user presses enter, run the slash command the user typed.
			state.FuzzyPicker.Show(func(state *State) {
				err := OnCommandExecuted(state, quit)
				if err != nil {
					log.Fatalf(err.Error())
				}
			})

			// Assemble add the items to the fuzzy sorter.
			for _, command := range commands {
				state.FuzzyPicker.Items = append(state.FuzzyPicker.Items, command)
				state.FuzzyPicker.StringItems = append(
					state.FuzzyPicker.StringItems,
					fmt.Sprintf(
						"%s %s\t%s - %s", // ie: "/quit (/q)        Quit - quits slime"
						strings.Replace(strings.Join(command.Permutations, " "), "/", string(state.Command[0]), -1),
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
