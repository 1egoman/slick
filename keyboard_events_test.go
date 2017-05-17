package main_test

import (
	"fmt"
	"reflect"
	. "github.com/1egoman/slime"
	"github.com/1egoman/slime/gateway"
	"github.com/1egoman/slime/gateway/slack"
	"github.com/gdamore/tcell"
	"github.com/jarcoal/httpmock"
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
			Sender: &gateway.User{Name: "my-user"},
			Text:   fmt.Sprintf("Hello world! %d", i),
			Hash: string(i),
		})
	}
	return s
}

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
			HandleKeyboardEvent(key, test.InitialState, quit)
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


// Message history movement
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
		"In chat mode, the message history can be scrolled up a half page at a time with Ctrl+U",
		InitialMessageHistoryState(3, 3), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlU, ' ', tcell.ModNone)},
		func(state *State) bool {
			// Note: we are rendering 10 items per page, 8 == (10 / 2) + 3
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
		"In chat mode, the message history can be scrolled down a half page at a time with Ctrl+D",
		InitialMessageHistoryState(14, 4), // Selecting index 14, and item 4 is on the bottom
		[]*tcell.EventKey{tcell.NewEventKey(tcell.KeyCtrlD, ' ', tcell.ModNone)},
		func(state *State) bool {
			// Note: we are rendering 10 items per page, 9 == 14 - (10 / 2)
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

			// When moving between messages, if we move to the last message, then slime will try to load
			// the next page of events. We don't want this to happen.
			if test.InitialState.ActiveConnection() != nil {
				selectedChannelId := test.InitialState.ActiveConnection().SelectedChannel().Id
				httpmock.Activate()
				httpmock.RegisterResponder(
					"GET",
					"https://slack.com/api/channels.history?token=token&channel=" + selectedChannelId + "&count=100",
					httpmock.NewStringResponder(200, `{"ok": true, "messages": []}`),
				)
			}

			// fmt.Printf("Test `%s` Running.", test.Name)

			// Create fresh state for the test.
			quit := make(chan struct{}, 1)

			// Run the test.
			for _, key := range test.Keys {
				HandleKeyboardEvent(key, test.InitialState, quit)
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

var argvTests  = []struct{
	Input string
	Output []string
}{
	{`a b c d`, []string{`a`, `b`, `c`, `d`}}, // no quotes or escapes
	{`a b "c d"`, []string{`a`, `b`, `c d`}}, // quotes at end
	{`a "b c" d`, []string{`a`, `b c`, `d`}}, // quotes in middle
	{`a b "c d`, []string{`a`, `b`, `c d`}}, // mismatched quotes
	{`a \"b c d`, []string{`a`, `\"b`, `c`, `d`}}, // escaped quote
	{`a \b c d`, []string{`a`, `\b`, `c`, `d`}}, // backslash in the args
}

func TestCreateArgvFromString(t *testing.T) {
	for index, argv := range argvTests {
		value := CreateArgvFromString(argv.Input)
		if !reflect.DeepEqual(value, argv.Output) {
			t.Errorf("Test index %d failed! Should have gotten %+v, really got %+v", index, argv.Output, value)
		}
	}
}
