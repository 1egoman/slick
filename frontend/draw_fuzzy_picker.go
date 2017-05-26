package frontend

import (
	"strings"
	"github.com/gdamore/tcell"
	"github.com/1egoman/slick/color"
)

// The amount of rows at max that can be in the fuzzy picker.
const FuzzyPickerMaxSize = 10

func (term *TerminalDisplay) DrawFuzzyPicker(
	preItems []string,
	selectedIndex int,
	bottomDisplayedItem int,
	rank func(string)int,
	config map[string]string,
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
		term.screen.SetCell(i, startingRow-1, color.DeSerializeStyleTcell(config["FuzzyPicker.TopBorderColor"]), ' ')
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
		style := tcell.StyleDefault
		if ct == projectedSelectedIndex {
			term.WriteTextStyle(0, row, color.DeSerializeStyleTcell(config["FuzzyPicker.ActiveItemColor"]), ">")
			style = color.DeSerializeStyleTcell(config["FuzzyPicker.ActiveItemColor"])
		}

		// Draw item
		// If a tab is present, then the part after the tab should on on the right
		tabIndex := strings.Index(item, "\t")
		if tabIndex >= 0 {
			term.WriteTextStyle(width - (len(item) - tabIndex) - 1, row, style, item[tabIndex:]) // right bit
			term.WriteTextStyle(2, row, style, item[:tabIndex]) // left bit
		} else {
			term.WriteTextStyle(2, row, style, item)
		}
	}
}
