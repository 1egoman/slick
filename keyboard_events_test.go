package main_test

import (
	"fmt"
	. "github.com/1egoman/slick"
	"github.com/1egoman/slick/gateway"
	"github.com/1egoman/slick/gateway/slack"
	"github.com/gdamore/tcell"
	"github.com/jarcoal/httpmock"
	"net/http"
	"io/ioutil"
	"reflect"
	"testing"
)

func NewRuneEvent(char rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, char, tcell.ModNone)
}

func InitialFuzzyPickerState(initialSelectedIndex int, initialBottomItemIndex int) *State {
	s := NewInitialStateMode("pick")
	s.FuzzyPicker.SelectedItem = initialSelectedIndex
	s.FuzzyPicker.BottomItem = initialBottomItemIndex
	for i := 1; i <= 15; i++ {
		s.FuzzyPicker.Items = append(s.FuzzyPicker.Items, interface{}(i))
	}
	s.FuzzyPicker.StringItems = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"}
	s.FuzzyPicker.Visible = true
	return s
}

func InitialMessageHistoryState(initialSelectedIndex int, initialBottomItemIndex int) *State {
	s := NewInitialStateMode("chat")
	s.SelectedMessageIndex = initialSelectedIndex
	s.BottomDisplayedItem = initialBottomItemIndex
	s.RenderedMessageNumber = 10 // For our tests, assume that 10 messages will fit on the screen.

	// Define an active connection
	s.Connections = append(s.Connections, gatewaySlack.New("token"))
	s.SetActiveConnection(len(s.Connections) - 1)
	s.ActiveConnection().SetSelectedChannel(&gateway.Channel{Id: "channel-id", Name: "general"}) // Set a channel

	// Add messages to the active connection
	for i := 1; i <= 15; i++ {
		s.ActiveConnection().AppendMessageHistory(gateway.Message{
			Timestamp: i,
			Sender:    &gateway.User{Name: "my-user"},
			Text:      fmt.Sprintf("Hello world! %d", i),
			Hash:      string(i),
		})
	}

	// Page by 5
	s.Configuration["Message.PageAmount"] = "5"
	return s
}

//
// ALL OTHER KEY EVENTS
//
var keyEvents = []struct {
	Name         string
	InitialState *State
	Keys         []*tcell.EventKey
	Passed       func(*State) bool
}{
	//
	// SWITCHING MODES
	//

	// chat <=> writ
	{
		"Pressing w in chat mode moves to writ mode.",
		NewInitialStateMode("writ"), []*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone)},
		func(state *State) bool { return state.Mode == "writ" },
	},
	{
		"Pressing esc in writ mode brings user to chat mode.",
		NewInitialStateMode("writ"), []*tcell.EventKey{tcell.NewEventKey(tcell.KeyEscape, ' ', tcell.ModNone)},
		func(state *State) bool { return state.Mode == "chat" },
	},
	{
		"Backspacing past the start of the line brings the user to chat mode.",
		NewInitialStateMode("writ"), []*tcell.EventKey{tcell.NewEventKey(tcell.KeyDEL, ' ', tcell.ModNone)},
		func(state *State) bool { return state.Mode == "chat" },
	},
	{
		"Pressing : moves to writ mode with a colon prepopulated.",
		NewInitialStateMode("writ"), []*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone)},
		func(state *State) bool {
			return state.Mode == "writ" &&
				state.Command[0] == ':' && state.CommandCursorPosition == 1
		},
	},
	{
		"Pressing / moves to writ mode with a slash prepopulated.",
		NewInitialStateMode("writ"), []*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone)},
		func(state *State) bool {
			return state.Mode == "writ" &&
				state.Command[0] == '/' && state.CommandCursorPosition == 1
		},
	},

	// chat <=> pick
	{
		"Pressing p in chat mode brings user to pick mode.",
		NewInitialState(), []*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone)},
		func(state *State) bool { return state.Mode == "pick" },
	},
	{
		"Pressing esc in pick mode brings user to chat mode.",
		NewInitialStateMode("pick"), []*tcell.EventKey{tcell.NewEventKey(tcell.KeyEscape, ' ', tcell.ModNone)},
		func(state *State) bool { return state.Mode == "chat" },
	},

	// Fuzzy picker movement
	{
		"In pick mode, ctrl+j moves down",
		InitialFuzzyPickerState(1, 0), // Selecting index 1, and item 0 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone)},
		func(state *State) bool { return state.FuzzyPicker.SelectedItem == 0 },
	},
	{
		"In pick mode, ctrl+k moves down",
		InitialFuzzyPickerState(1, 0), // Selecting index 1, and item 0 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone)},
		func(state *State) bool { return state.FuzzyPicker.SelectedItem == 2 },
	},
	{
		"In pick mode, a user can force the fuzzy picker to scroll up to show off-screen items.",
		InitialFuzzyPickerState(9, 0), // Selecting index 9, and item 0 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone)},
		func(state *State) bool {
			return state.FuzzyPicker.SelectedItem == 10 &&
				state.FuzzyPicker.BottomItem == 1
		},
	},
	{
		"In pick mode, a user can force the fuzzy picker to scroll down to show off-screen items.",
		InitialFuzzyPickerState(3, 3), // Selecting index 3, and item 3 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone)},
		func(state *State) bool {
			return state.FuzzyPicker.SelectedItem == 2 &&
				state.FuzzyPicker.BottomItem == 2
		},
	},
	{
		"In pick mode, the fuzzy picker can't scroll higher than the number of items",
		InitialFuzzyPickerState(14, 4), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone)},
		func(state *State) bool { return state.FuzzyPicker.SelectedItem == 14 },
	},
	{
		"In pick mode, the fuzzy picker can't scroll below the last item",
		InitialFuzzyPickerState(0, 0), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone)},
		func(state *State) bool { return state.FuzzyPicker.SelectedItem == 0 },
	},

	// Editing operations
	{
		"In writ mode, ctrl+w deletes an entire word when at the end of the word",
		func() *State {
			s := NewInitialStateMode("writ")
			s.Command = []rune("foo bar baz")
			s.CommandCursorPosition = len(s.Command)
			return s
		}(),
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlW, ' ', tcell.ModNone)},
		func(state *State) bool { return string(state.Command) == "foo bar" },
	},
	{
		"In writ mode, ctrl+w deletes the rest of a word when in the middle of a word",
		func() *State {
			s := NewInitialStateMode("writ")
			s.Command = []rune("foo bar baz")
			s.CommandCursorPosition = len(s.Command) - 1 // -1 means to leave the z in baz alone
			return s
		}(),
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlW, ' ', tcell.ModNone)},
		func(state *State) bool { return string(state.Command) == "foo barz" },
	},
	{
		"In writ mode, ctrl+a goes to the start of the command",
		func() *State {
			s := NewInitialStateMode("writ")
			s.Command = []rune("foo bar baz")
			s.CommandCursorPosition = len(s.Command)
			return s
		}(),
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlA, ' ', tcell.ModNone)},
		func(state *State) bool { return state.CommandCursorPosition == 0 },
	},
	{
		"In writ mode, ctrl+e goes to the end of the command",
		func() *State {
			s := NewInitialStateMode("writ")
			s.Command = []rune("foo bar baz")
			s.CommandCursorPosition = 0
			return s
		}(),
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlE, ' ', tcell.ModNone)},
		func(state *State) bool { return state.CommandCursorPosition == len(state.Command) },
	},
}

func TestHandleKeyboardEvent(t *testing.T) {
	for _, test := range keyEvents {
		// fmt.Printf("Test `%s` Running.", test.Name)

		// Create fresh state for the test.
		quit := make(chan struct{}, 1)

		// Run the test.
		for _, key := range test.Keys {
			HandleKeyboardEvent(key, test.InitialState, nil, quit)
		}

		// Verify it passed.
		if !test.Passed(test.InitialState) {
			t.Errorf("Test `%s` failed.", test.Name)
		} else {
			fmt.Printf(".")
		}

		httpmock.DeactivateAndReset()
	}
}

//
// MESSAGE HISTORY MOVEMENT (j, k, gg, G, ctrl-u, ctrl-d, etc...)
//
var messageHistoryKeyEvents = []struct {
	Name         string
	InitialState *State
	Keys         []*tcell.EventKey
	Passed       func(*State) bool
}{
	{
		"In chat mode, j moves messages down",
		InitialMessageHistoryState(1, 0), // Selecting index 1, and item 0 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone)},
		func(state *State) bool { return state.SelectedMessageIndex == 0 },
	},
	{
		"In chat mode, k moves messages up",
		InitialMessageHistoryState(1, 0), // Selecting index 1, and item 0 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone)},
		func(state *State) bool { return state.SelectedMessageIndex == 2 },
	},
	{
		"In chat mode, a user can force the message history to scroll up to show off-screen items.",
		InitialMessageHistoryState(9, 0), // Selecting index 9, and item 0 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone)},
		func(state *State) bool {
			return state.SelectedMessageIndex == 10 &&
				state.BottomDisplayedItem == 1
		},
	},
	{
		"In chat mode, a user can force the message history to scroll down to show off-screen items.",
		InitialMessageHistoryState(3, 3), // Selecting index 3, and item 3 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone)},
		func(state *State) bool {
			return state.SelectedMessageIndex == 2 &&
				state.BottomDisplayedItem == 2
		},
	},
	{
		"In chat mode, the message history can't go past the last item in the history",
		InitialMessageHistoryState(14, 4), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone)},
		func(state *State) bool { return state.SelectedMessageIndex == 14 },
	},
	{
		"In chat mode, the message history can't scroll below the most recent message",
		InitialMessageHistoryState(0, 0), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone)},
		func(state *State) bool { return state.SelectedMessageIndex == 0 },
	},
	{
		"In chat mode, the message history can't scroll below the most recent message",
		InitialMessageHistoryState(0, 0), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone)},
		func(state *State) bool { return state.SelectedMessageIndex == 0 },
	},
	{
		"In chat mode, the message history can be scrolled all the way to the top by pressing 'gg'.",
		InitialMessageHistoryState(0, 0), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{
			tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
			tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),
		},
		func(state *State) bool {
			return state.SelectedMessageIndex == 14 &&
				state.BottomDisplayedItem == 7
		},
	},
	{
		"In chat mode, the message history can be scrolled back to the bottom by pressing 'G'.",
		InitialMessageHistoryState(3, 3), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone)},
		func(state *State) bool {
			return state.SelectedMessageIndex == 0 &&
				state.BottomDisplayedItem == 0
		},
	},
	{
		"In chat mode, the message history can be scrolled up in bulk with Ctrl+U",
		InitialMessageHistoryState(3, 3), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlU, ' ', tcell.ModNone)},
		func(state *State) bool {
			// Note: we are rendering 10 items per page, 8 == state.Configuration["Message.PageAmount"] + 3
			return state.SelectedMessageIndex == 8 &&
				state.BottomDisplayedItem == 8
		},
	},
	{
		"In chat mode, the message history can't be scrolled up when already at the oldest.",
		InitialMessageHistoryState(14, 4), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlU, ' ', tcell.ModNone)},
		func(state *State) bool {
			return state.SelectedMessageIndex == 14 &&
				state.BottomDisplayedItem == 4
		},
	},
	{
		"In chat mode, the message history can be scrolled up in bulk with Ctrl+D",
		InitialMessageHistoryState(14, 4), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlD, ' ', tcell.ModNone)},
		func(state *State) bool {
			// Note: we are rendering 10 items per page, 9 == 14 - state.Configuration["Message.PageAmount"]
			return state.SelectedMessageIndex == 9 &&
				state.BottomDisplayedItem == 0
		},
	},
	{
		"In chat mode, the message history can't be scrolled down when already at the oldest.",
		InitialMessageHistoryState(0, 0), // Selecting index 0, and item 0 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlD, ' ', tcell.ModNone)},
		func(state *State) bool {
			return state.SelectedMessageIndex == 0 &&
				state.BottomDisplayedItem == 0
		},
	},
}

func TestHandleMessageMovementKeyboardEvents(t *testing.T) {
	for _, test := range messageHistoryKeyEvents {
		func() { // (Function closure exists so that we can defer inside.)
			defer httpmock.DeactivateAndReset()

			// When moving between messages, if we move to the last message, then slick will try to load
			// the next page of events. We don't want this to happen.
			if test.InitialState.ActiveConnection() != nil {
				selectedChannelId := test.InitialState.ActiveConnection().SelectedChannel().Id
				httpmock.Activate()
				httpmock.RegisterResponder(
					"GET",
					"https://slack.com/api/channels.history?token=token&channel="+selectedChannelId+"&count=100",
					httpmock.NewStringResponder(200, `{"ok": true, "messages": []}`),
				)
			}

			// fmt.Printf("Test `%s` Running.", test.Name)

			// Create fresh state for the test.
			quit := make(chan struct{}, 1)

			// Run the test.
			for _, key := range test.Keys {
				HandleKeyboardEvent(key, test.InitialState, nil, quit)
			}

			// Verify it passed.
			if !test.Passed(test.InitialState) {
				t.Errorf("Test `%s` failed.", test.Name)
			} else {
				fmt.Printf(".")
			}
		}()
	}
}

//
// CreateArgvFromString
//

var argvTests = []struct {
	Input  string
	Output []string
}{
	{`a b c d`, []string{`a`, `b`, `c`, `d`}},     // no quotes or escapes
	{`a b "c d"`, []string{`a`, `b`, `c d`}},      // quotes at end
	{`a "b c" d`, []string{`a`, `b c`, `d`}},      // quotes in middle
	{`a b "c d`, []string{`a`, `b`, `c d`}},       // mismatched quotes
	{`a \"b c d`, []string{`a`, `\"b`, `c`, `d`}}, // escaped quote
	{`a \b c d`, []string{`a`, `\b`, `c`, `d`}},   // backslash in the args
}

func TestCreateArgvFromString(t *testing.T) {
	for index, argv := range argvTests {
		value := CreateArgvFromString(argv.Input)
		if !reflect.DeepEqual(value, argv.Output) {
			t.Errorf("Test index %d failed! Should have gotten %+v, really got %+v", index, argv.Output, value)
		}
	}
}

//
// keystackQuantityParser
//
func TestKeystackQuantityParser(t *testing.T) {
	for input, output := range map[string]struct {
		Quantity int
		Keystack string
	}{
		"5a":   {5, "a"},
		"123a": {123, "a"},
		"2abc": {2, "abc"},
		"0a":   {0, "a"},
		"a":    {1, "a"},
		"a123": {1, "a123"},
		"123":  {123, ""},
	} {
		quant, stack, err := KeystackQuantityParser([]rune(input))
		if err != nil {
			t.Errorf("Error on %s: %s", input, err)
		}

		if quant != output.Quantity {
			t.Errorf("Error: quantity doesn't match up for %s", input)
		}
		if string(stack) != output.Keystack {
			t.Errorf("Error: keystack doesn't match up for %s, %s != %s", input, string(stack), output.Keystack)
		}
	}
}

// Sending commands
func TestEmojiStartingMessageWontMeTreatedAsCommand(t *testing.T) {
	// Mock slack
	userSentMessage := false
	httpmock.Activate()
	httpmock.RegisterResponder(
		"GET",
		"https://slack.com/api/chat.postMessage?token=token&channel=channel-id&text=%3Asmile%3A&link_names=true&parse=full&unfurl_links=true&as_user=true",
		func(req *http.Request) (*http.Response, error) {
			userSentMessage = true
			return httpmock.NewStringResponse(200, `{"ok": true}`), nil
		},
	)
	defer httpmock.DeactivateAndReset()

	quit := make(chan struct{}, 1)

	// Create fresh state for the test.
	state := NewInitialStateMode("writ")
	state.Command = []rune(":smile:")
	state.CommandCursorPosition = 0

	state.Connections = append(state.Connections, gatewaySlack.New("token"))
	state.SetActiveConnection(len(state.Connections) - 1)
	state.ActiveConnection().SetSelectedChannel(&gateway.Channel{Id: "channel-id", Name: "general"}) // Set a channel

	initialMessageHistoryLength := len(state.ActiveConnection().MessageHistory())

	// Run the test.
	HandleKeyboardEvent(
		tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
		state,
		nil,
		quit,
	)

	// Verify it passed - a message should have been posted, not a command run.
	if len(state.ActiveConnection().MessageHistory()) != initialMessageHistoryLength || !userSentMessage {
		t.Errorf("Test failed.")
	}
}

// Path auto complete:
// This is the functionality that lets users hit tab after typing a slash and get auto complete for
// those paths, ie I type / and hit tab, I'll see `etc`, `var`, `tmp`, etc...
func TestPathAutoComplete(t *testing.T) {
	quit := make(chan struct{}, 1)

	// A directory in `/`
	// TODO: look into fs mocking solutions
	directory := "var"

	// Create fresh state for the test.
	state := NewInitialStateMode("writ")
	state.Command = []rune("foo bar baz /") // Trying to autocomplete a file path
	state.CommandCursorPosition = len(state.Command)

	// Press tab, and verify that the fuzzy picker is filled with a bunch of items.
	HandleKeyboardEvent(
		tcell.NewEventKey(tcell.KeyTab, ' ', tcell.ModNone),
		state,
		nil,
		quit,
	)
	state.FuzzyPicker.OnResort(state)

	if state.Mode != "pick" {
		t.Errorf("Not picking file path after presing / and tab")
	}

	// Get all files and folders in "/"
	files, err := ioutil.ReadDir("/")
	if err != nil {
		t.Errorf(err.Error())
	}
	var fileItems []interface{}
	for _, file := range files {
		fileItems = append(fileItems, file.Name())
	}

	// Verify that's what's in the fuzzy picker right now
	if !reflect.DeepEqual(state.FuzzyPicker.Items, fileItems) {
		t.Errorf("%+v != %+v", state.FuzzyPicker.Items, fileItems)
	}

	// Type out that directory name
	for _, char := range directory {
		HandleKeyboardEvent(
			tcell.NewEventKey(tcell.KeyRune, char, tcell.ModNone),
			state,
			nil,
			quit,
		)
		state.FuzzyPicker.OnResort(state)
	}

	// Then type a slash
	HandleKeyboardEvent(
		tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
		state,
		nil,
		quit,
	)
	state.FuzzyPicker.OnResort(state)

	// Then verify that the fuzzy picker has everything in that directory
	files, err = ioutil.ReadDir("/"+directory)
	if err != nil {
		t.Errorf(err.Error())
	}
	fileItems = nil // Clear slice
	for _, file := range files {
		fileItems = append(fileItems, file.Name())
	}
	if !reflect.DeepEqual(state.FuzzyPicker.Items, fileItems) {
		t.Errorf("%+v != %+v", state.FuzzyPicker.Items, fileItems)
	}

	// Press tab to autocomplete the selected item
	commandLength := len(state.Command)
	HandleKeyboardEvent(
		tcell.NewEventKey(tcell.KeyTab, ' ', tcell.ModNone),
		state,
		nil,
		quit,
	)
	state.FuzzyPicker.OnResort(state)

	if commandLength >= len(state.Command) {
		t.Errorf("Tab to autocomplete file path did nothing")
	}

	// Finally, press backspace until the most recent slash.
	for state.Command[len(state.Command)-1] != '/' {
		HandleKeyboardEvent(
			tcell.NewEventKey(tcell.KeyBackspace, ' ', tcell.ModNone),
			state,
			nil,
			quit,
		)
		state.FuzzyPicker.OnResort(state)
	}

	// And make sure that the fuzzy picker has gone away
	if state.Mode != "writ" {
		t.Errorf("Fuzzy picker didn't dissapear when previous slash was removed.")
	}



	//
	// Reset the command. Trying another test
	//
	state.Command = []rune("foo bar baz ./") // Trying to autocomplete a file path in $HOME
	state.CommandCursorPosition = len(state.Command)

	// Press tab to start autocomplete
	HandleKeyboardEvent(
		tcell.NewEventKey(tcell.KeyTab, ' ', tcell.ModNone),
		state,
		nil,
		quit,
	)
	state.FuzzyPicker.OnResort(state)

	// Move into `./color`
	for _, char := range "color" {
		HandleKeyboardEvent(
			tcell.NewEventKey(tcell.KeyRune, char, tcell.ModNone),
			state,
			nil,
			quit,
		)
		state.FuzzyPicker.OnResort(state)
	}

	// Press slash
	HandleKeyboardEvent(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone), state, nil, quit)
	state.FuzzyPicker.OnResort(state)

	// Press tab to autocomplete
	HandleKeyboardEvent(tcell.NewEventKey(tcell.KeyTab, ' ', tcell.ModNone), state, nil, quit)
	state.FuzzyPicker.OnResort(state)

	// Verify that there are items in `./color`
	if len(state.FuzzyPicker.Items) == 0 {
		t.Errorf("Fuzzy picker items empty")
	}

	// Press space
	HandleKeyboardEvent(
		tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone),
		state,
		nil,
		quit,
	)
	state.FuzzyPicker.OnResort(state)

	// Make sure fuzzy picker is gone
	if state.Mode != "writ" {
		t.Errorf("Didn't remove file path picker after pressing space")
	}


	//
	// Reset the command. Trying another test
	//
	state.Command = []rune("badprefix/") // Trying to autocomplete a file path that is bad
	state.CommandCursorPosition = len(state.Command)

	// Press tab to start autocomplete
	HandleKeyboardEvent(
		tcell.NewEventKey(tcell.KeyTab, ' ', tcell.ModNone),
		state,
		nil,
		quit,
	)
	state.FuzzyPicker.OnResort(state)

	// Make sure fuzzy picker didn't pop up with bad prefix.
	if state.Mode != "writ" {
		t.Errorf("Autocompleted bad prefixed path")
	}
}
