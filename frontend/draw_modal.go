package frontend

import (
	// "log"
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell"
)

const idealModalWidth = 120
const idealModalHeight = 40
const horizontalGutter = 2;
const verticalGutter = 1;
const showModalScrollBar = true
const modalScrollBarCharacter = '='

const closeModalMessage = "[ Esc to close ]"

var headerStyle tcell.Style = tcell.StyleDefault.
	Foreground(tcell.ColorWhite).
	Background(tcell.ColorRed)

var scrollBarHandleStyle tcell.Style = tcell.StyleDefault.
	Foreground(tcell.ColorRed).
	Background(tcell.ColorRed)
var scrollBarTrackStyle tcell.Style = tcell.StyleDefault.
	Foreground(tcell.ColorSilver).
	Background(tcell.ColorSilver)


func calculateScrollBarProperties(currentLine int, totalLines int, modalHeight int) []rune {
	scrollBarHeight := int(math.Ceil(float64(totalLines) / float64(modalHeight)))
	scrollBarPosition := int((float32(currentLine) / float32(totalLines + (2 * scrollBarHeight))) * float32(modalHeight))

	// Create an array of bits indicating whether a given item is part of the scroll bar or not.
	scrollBarEnabledForCharacter := make([]rune, modalHeight)
	for i := 0; i < modalHeight; i++ {
		if i >= scrollBarPosition && i <= scrollBarPosition + scrollBarHeight {
			// Makes up the handle of the scroll bar
			scrollBarEnabledForCharacter[i] = modalScrollBarCharacter
		} else {
			// Makes up the scroll bar handle's "track"
			scrollBarEnabledForCharacter[i] = '|'
		}
	}

	return scrollBarEnabledForCharacter
}

func (term *TerminalDisplay) DrawModal(title string, body string, scrollPosition int) {
	width, height := term.screen.Size()

	// Remove leading and trailing whitespace.
	body = strings.Trim(body, "\n\t ")

	// Given the scroll position and the body, trim away `scrollPosition` lines at the start of the
	// body.
	bodyLines := strings.Split(body, "\n")
	if len(bodyLines) > scrollPosition {
		body = strings.Join(bodyLines[scrollPosition:], "\n")
	}

	// Calculate the modal width and height
	var modalWidth int
	if width > idealModalWidth {
		modalWidth = idealModalWidth
	} else {
		modalWidth = width
	}
	var modalHeight int
	if height > idealModalHeight {
		modalHeight = idealModalHeight
	} else {
		modalHeight = height - horizontalGutter
	}

	// Caculate the upper left x/y position for the modal
	modalUpperLeftX := (width - modalWidth) / 2
	modalUpperLeftY := (height - modalHeight) / 2

	// ----------------------------------------------------------------------------
	//	Render frame
	// ------------------------------------------------------------------------------

	// Assemble modal hint
	modalHint := fmt.Sprintf(
		" %d/%d (%d%%) %s +",
		scrollPosition, // Current line
		len(bodyLines), // Total lines
		int(float32(scrollPosition) / float32(len(bodyLines)) * 100), // Percent
		closeModalMessage, // Hint on how to close the modal
	)

	// Top header
	modalHintMessagePosition := modalWidth - len(modalHint)
	for i := 0; i < modalHintMessagePosition; i++ {
		if i == 0 {
			// Left side corner
			term.WriteTextStyle(modalUpperLeftX+i, modalUpperLeftY, headerStyle, "+")
		} else {
			term.WriteTextStyle(modalUpperLeftX+i, modalUpperLeftY, headerStyle, "-")
		}
	}
	// Render modal hint
	term.WriteTextStyle(
		modalUpperLeftX+modalHintMessagePosition,
		modalUpperLeftY,
		headerStyle,
		modalHint,
	)
	// Render title
	term.WriteTextStyle(
		modalUpperLeftX+1,
		modalUpperLeftY,
		headerStyle,
		fmt.Sprintf(" %s ", title),
	)

	// Sides
	scrollBarCharacter := calculateScrollBarProperties(scrollPosition, len(bodyLines), modalHeight)
	for i := 1; i < modalHeight; i++ {
		// Left side
		term.screen.SetCell(modalUpperLeftX, modalUpperLeftY+i, tcell.StyleDefault, '|')

		// Clear the middle
		for j := 1; j < modalWidth - 1; j++ {
			term.screen.SetCell(modalUpperLeftX+j, modalUpperLeftY+i, tcell.StyleDefault, ' ');
		}

		// Right side
		var scrollBarStyle tcell.Style
		if scrollBarCharacter[i-1] == modalScrollBarCharacter {
			scrollBarStyle = scrollBarHandleStyle
		} else {
			scrollBarStyle = scrollBarTrackStyle
		}
		term.screen.SetCell(
			modalUpperLeftX+modalWidth-1,
			modalUpperLeftY+i,
			scrollBarStyle,
			scrollBarCharacter[i-1],
		)
	}

	// Bottom header
	for i := 0; i < modalWidth; i++ {
		if i == 0 || i == modalWidth-1 {
			// Corners
			term.WriteTextStyle(modalUpperLeftX+i, modalUpperLeftY+modalHeight, tcell.StyleDefault, "+")
		} else {
			term.WriteTextStyle(modalUpperLeftX+i, modalUpperLeftY+modalHeight, tcell.StyleDefault, "-")
		}
	}

	// ----------------------------------------------------------------------------
	// 	Render content
	// ------------------------------------------------------------------------------

	term.WriteParagraphStyle(
		modalUpperLeftX+horizontalGutter,
		modalUpperLeftY+verticalGutter,
		modalWidth - horizontalGutter - horizontalGutter,
		modalHeight - verticalGutter - verticalGutter,
		tcell.StyleDefault,
		body,
	)
}
