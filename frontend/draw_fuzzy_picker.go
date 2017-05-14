package frontend

import (
	"github.com/gdamore/tcell"
)

// The amount of rows at max that can be in the fuzzy picker.
const FuzzyPickerMaxSize = 10

func (term *TerminalDisplay) DrawFuzzyPicker(
	preItems []string,
	selectedIndex int,
	bottomDisplayedItem int,
	rank func(string)int,
) {
	width, height := term.screen.Size()

	// Filter items to remove those that have a negitive rank.
	var items []string
	for _, item := range preItems {
		if rank(item) >= 0 {
			items = append(items, item)
		}
	}


	// If there's more than one page of items, only show one page's worth.
	if len(items) > FuzzyPickerMaxSize {
		items = items[bottomDisplayedItem : bottomDisplayedItem+FuzzyPickerMaxSize]
	}
	projectedSelectedIndex := selectedIndex - bottomDisplayedItem
	startingRow := height - len(items) - BottomPadding // The top row of the fuzzy picker

	// Make sure that the item that is selected is never larger then the max item.
	if projectedSelectedIndex > len(items)-1 {
		projectedSelectedIndex = len(items) - 1
	}

	// Above the top of the picker, draw a border.
	for i := 0; i < width; i++ {
		term.screen.SetCell(i, startingRow-1, term.Styles["FuzzyPickerTopBorder"], ' ')
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
