package main_test

import (
  "fmt"
  "testing"
	"github.com/gdamore/tcell"
  . "github.com/1egoman/slime"
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

var keyEvents = []struct{
  Name string
  InitialState *State
  Key *tcell.EventKey
  Passed func(*State) bool
}{
  //
  // SWITCHING MODES
  //

  // chat <=> writ
  {
    "Pressing w in chat mode moves to writ mode.",
    NewInitialStateMode("writ"), tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),
    func(state *State) bool { return state.Mode == "writ" },
  },
  {
    "Pressing esc in writ mode brings user to chat mode.",
    NewInitialStateMode("writ"), tcell.NewEventKey(tcell.KeyEscape, ' ', tcell.ModNone),
    func(state *State) bool { return state.Mode == "chat" },
  },
  {
    "Backspacing past the start of the line brings the user to chat mode.",
    NewInitialStateMode("writ"), tcell.NewEventKey(tcell.KeyDEL, ' ', tcell.ModNone),
    func(state *State) bool { return state.Mode == "chat" },
  },
  {
    "Pressing : moves to writ mode with a colon prepopulated.",
    NewInitialStateMode("writ"), tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),
    func(state *State) bool { return state.Mode == "writ" &&
      state.Command[0] == ':' && state.CommandCursorPosition == 1 },
  },

  // chat <=> pick
  {
    "Pressing p in chat mode brings user to pick mode.",
    NewInitialState(), tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
    func(state *State) bool { return state.Mode == "pick" },
  },
  {
    "Pressing esc in pick mode brings user to chat mode.",
    NewInitialStateMode("pick"), tcell.NewEventKey(tcell.KeyEscape, ' ', tcell.ModNone),
    func(state *State) bool { return state.Mode == "chat" },
  },

  // Fuzzy picker movement
  {
    "In pick mode, ctrl+j moves down",
    InitialFuzzyPickerState(1, 0), // Selecting index 1, and item 0 is on the bottom
    tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone),
    func(state *State) bool { return state.FuzzyPicker.SelectedItem == 0 },
  },
  {
    "In pick mode, ctrl+k moves down",
    InitialFuzzyPickerState(1, 0), // Selecting index 1, and item 0 is on the bottom
    tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
    func(state *State) bool { return state.FuzzyPicker.SelectedItem == 2 },
  },
  {
    "In pick mode, a user can force the fuzzy picker to scroll up to show off-screen items.",
    InitialFuzzyPickerState(9, 0), // Selecting index 9, and item 0 is on the bottom
    tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
    func(state *State) bool {
      return state.FuzzyPicker.SelectedItem == 10 &&
        state.FuzzyPicker.BottomItem == 1
    },
  },
  {
    "In pick mode, a user can force the fuzzy picker to scroll down to show off-screen items.",
    InitialFuzzyPickerState(3, 3), // Selecting index 3, and item 3 is on the bottom
    tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone),
    func(state *State) bool {
      return state.FuzzyPicker.SelectedItem == 2 &&
        state.FuzzyPicker.BottomItem == 2
    },
  },
  {
    "In pick mode, the fuzzy picker can't scroll higher than the number of items",
    InitialFuzzyPickerState(14, 4), // Selecting index 14, and item 4 is on the bottom
    tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
    func(state *State) bool { return state.FuzzyPicker.SelectedItem == 14 },
  },
  {
    "In pick mode, the fuzzy picker can't scroll below the last item",
    InitialFuzzyPickerState(0, 0), // Selecting index 14, and item 4 is on the bottom
    tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone),
    func(state *State) bool { return state.FuzzyPicker.SelectedItem == 0 },
  },
}

func TestHandleKeyboardEvent(t *testing.T) {
  for _, test := range keyEvents {
    // Create fresh state for the test.
    quit := make(chan struct{}, 1)

    // Run the test.
    HandleKeyboardEvent(test.Key, test.InitialState, quit)

    // Verify it passed.
    if !test.Passed(test.InitialState) {
      t.Errorf("Test `%s` failed.", test.Name)
    } else {
      fmt.Printf(".")
    }
  }
}
