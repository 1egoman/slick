package frontend

import (
	"github.com/gdamore/tcell"
	// "log"
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
func (term *TerminalDisplay) WriteParagraphStyle(x int, y int, width int, height int, style tcell.Style, text string) {
	printableMessage := ParseMarkdown(text)
	for yOffset, line := range printableMessage.Lines(width) {
		xOffset := 0
		for _, part := range line {
			term.WriteTextStyle(x+xOffset, y+yOffset, style, part.Content)
			xOffset += len(part.Content)
		}

		// Never render more content then will fit in the box.
		if yOffset > height - 1 {
			break;
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
