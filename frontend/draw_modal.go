package frontend

import (
	// "log"
	"fmt"
	"github.com/gdamore/tcell"
)

const idealModalWidth = 120
const idealModalHeight = 40
const closeModalMessage = " [ Esc to close ] +"

func (term *TerminalDisplay) DrawModal(title string, body string) {
	width, height := term.screen.Size()

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
		modalHeight = height - 2
	}

	// Caculate the upper left x/y position for the modal
	modalUpperLeftX := (width - modalWidth) / 2
	modalUpperLeftY := (height - modalHeight) / 2

	// ----------------------------------------------------------------------------
	//	Render frame
	// ------------------------------------------------------------------------------

	// Top header
	closeModalMessagePosition := modalWidth - len(closeModalMessage)
	for i := 0; i < closeModalMessagePosition; i++ {
		if i == 0 {
			// Left side corner
			term.WriteTextStyle(modalUpperLeftX+i, modalUpperLeftY, tcell.StyleDefault, "+")
		} else {
			term.WriteTextStyle(modalUpperLeftX+i, modalUpperLeftY, tcell.StyleDefault, "-")
		}
	}
	// Render hint to close modal in upper right
	term.WriteTextStyle(
		modalUpperLeftX+closeModalMessagePosition,
		modalUpperLeftY,
		tcell.StyleDefault,
		closeModalMessage,
	)
	// Render title
	term.WriteTextStyle(
		modalUpperLeftX+1,
		modalUpperLeftY,
		tcell.StyleDefault,
		fmt.Sprintf(" %s ", title),
	)

	// Sides
	for i := 1; i < modalHeight; i++ {
		// Left side
		term.screen.SetCell(modalUpperLeftX, modalUpperLeftY+i, tcell.StyleDefault, '|')

		// Clear the middle
		for j := 1; j < modalWidth - 1; j++ {
			term.screen.SetCell(modalUpperLeftX+j, modalUpperLeftY+i, tcell.StyleDefault, ' ');
		}

		// Right side
		term.screen.SetCell(modalUpperLeftX+modalWidth-1, modalUpperLeftY+i, tcell.StyleDefault, '|')
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
		modalUpperLeftX+1,
		modalUpperLeftY+1,
		modalWidth,
		tcell.StyleDefault,
		body,
	)
}
