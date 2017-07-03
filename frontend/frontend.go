package frontend

import (
	"github.com/gdamore/tcell"
)

// The amount of lines at the bottom of the window to leave available for status bars.
// One line for the status bar
// One line for the command bar
var BottomPadding = 2

func NewTerminalDisplay(screen tcell.Screen) *TerminalDisplay {
	return &TerminalDisplay{screen: screen}
}

type TerminalDisplay struct {
	screen tcell.Screen
}

func (term *TerminalDisplay) Screen() tcell.Screen {
	return term.screen
}

func (term *TerminalDisplay) Render() {
	term.screen.Show()
}

func (term *TerminalDisplay) Close() {
	term.screen.Fini()
}

func (term *TerminalDisplay) WriteText(x int, y int, text string) {
	term.WriteTextStyle(x, y, tcell.StyleDefault, text)
}
func (term *TerminalDisplay) WriteTextStyle(x int, y int, style tcell.Style, text string) {
	for ct, char := range text {
		term.screen.SetCell(x+ct, y, style, char)
	}
}
func (term *TerminalDisplay) WriteParagraphStyle(x int, y int, width int, style tcell.Style, text string) {
	xOffset := 0
	yOffset := 0
	for _, char := range text {
		xOffset += 1
		if char == '\n' { // If we hit a newline, then wrap the the next line.
			yOffset += 1
			xOffset = 0
		} else if xOffset > width { // If we go over the window width, then wrap to the next line.
			yOffset += 1
			xOffset = 0
		} else {
			// Otherwise, print the character.
			term.screen.SetCell(x+xOffset, y+yOffset, style, char)
		}
	}
}

// Given a line number, reset it to the default style and erase it.
func (term *TerminalDisplay) DrawBlankLine(line int) {
	width, _ := term.screen.Size()
	for j := 0; j < width; j++ {
		char, _, style, _ := term.screen.GetContent(j, line)
		if char != ' ' || style != tcell.StyleDefault {
			term.screen.SetCell(j, line, tcell.StyleDefault, ' ')
		}
	}
}
func (term *TerminalDisplay) DrawBlankLines(start int, end int) {
	_, height := term.screen.Size()

	// If end < 0, then index from the other end.
	// ie, DrawBlankLines(0, -2) would draw blank lines from the first line up until the second
	// to last.
	if end < 0 {
		end = height - end
	}

	for i := start; i <= end; i++ {
		term.DrawBlankLine(i)
	}
}
