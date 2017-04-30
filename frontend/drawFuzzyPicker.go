package frontend

import (
	"github.com/gdamore/tcell"
)

// The amount of rows at max that can be in the fuzzy picker.
const fuzzyPickerMaxSize = 10;

func (term *TerminalDisplay) DrawFuzzyPicker(items []string, selectedIndex int) {
	width, height := term.screen.Size()

  // If there are over 5 items, truncate the list to 5 items
  if len(items) > fuzzyPickerMaxSize {
    items = items[:fuzzyPickerMaxSize]
  }
	bottomPadding := 2                                 // pad for the status bar and command bar
	startingRow := height - len(items) - bottomPadding // The top row of the fuzzy picker

	// Make sure that the item that is selected is never larger then the max item.
	if selectedIndex > len(items)-1 {
		selectedIndex = len(items) - 1
	}

  // Above the top of the picker, draw a border.
  for i := 0; i < width; i++ {
    term.screen.SetCell(i, startingRow - 1, term.Styles["FuzzyPickerTopBorder"], ' ')
  }

	for ct, item := range items {
		row := startingRow + (len(items) - 1) - ct

		// Clear the row.
		for i := 0; i < width; i++ {
			char, _, style, _ := term.screen.GetContent(i, row)
			if char != ' ' || style != tcell.StyleDefault {
				term.screen.SetCell(i, row, tcell.StyleDefault, ' ')
			}
		}

		// Add selected prefix for selected item
		if ct == selectedIndex {
      term.WriteTextStyle(0, row, term.Styles["FuzzyPickerActivePrefix"], ">")
    }
    // Draw item
    term.WriteText(2, row, item)
	}
}

