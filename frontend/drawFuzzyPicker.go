package frontend

import (
  // "log"
	"github.com/gdamore/tcell"
)

// The amount of rows at max that can be in the fuzzy picker.
const FuzzyPickerMaxSize = 10;

func (term *TerminalDisplay) DrawFuzzyPicker(items []string, selectedIndex int, bottomDisplayedItem int) {
	width, height := term.screen.Size()

  // If there's more than one page of items, only show one page's worth.
  if len(items) > FuzzyPickerMaxSize {
    items = items[bottomDisplayedItem:bottomDisplayedItem+FuzzyPickerMaxSize]
  }
  projectedSelectedIndex := selectedIndex - bottomDisplayedItem
	bottomPadding := 2                                 // pad for the status bar and command bar
	startingRow := height - len(items) - bottomPadding // The top row of the fuzzy picker

	// Make sure that the item that is selected is never larger then the max item.
	if projectedSelectedIndex > len(items)-1 {
		projectedSelectedIndex = len(items) - 1
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
		if ct == projectedSelectedIndex {
      term.WriteTextStyle(0, row, term.Styles["FuzzyPickerActivePrefix"], ">")
    }
    // Draw item
    term.WriteText(2, row, item)
	}
}

