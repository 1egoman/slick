package frontend

import (
	"github.com/gdamore/tcell"
)

type Display interface {
	DrawStatusBar()
	Render()

	// Lower level primatives
	WriteText(tcell.Screen, int, int, string)
}

func NewTerminalDisplay(screen tcell.Screen) *TerminalDisplay {
	return &TerminalDisplay{screen: screen}
}

type TerminalDisplay struct {
	screen tcell.Screen
}

func (term *TerminalDisplay) WriteText(x int, y int, text string) {
	for ct, char := range text {
		term.screen.SetCell(x+ct, y, tcell.StyleDefault, char)
	}
}

func (term *TerminalDisplay) Render() {
	term.screen.Show()
}

func (term *TerminalDisplay) DrawStatusBar() {
	_, height := term.screen.Size()
	lastRow := height - 1

	term.WriteText(0, lastRow, "Foo Bar!")
}
