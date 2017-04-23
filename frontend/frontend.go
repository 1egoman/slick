package frontend

import (
	"github.com/gdamore/tcell"

	"github.com/1egoman/slime/gateway"  // The thing to interface with slack
)

type Display interface {
	DrawStatusBar()
	DrawCommandBar(string, int)
	Render()

	// Lower level primatives
	WriteText(tcell.Screen, int, int, string)
}

func NewTerminalDisplay(screen tcell.Screen) *TerminalDisplay {
	return &TerminalDisplay{
		screen: screen,
		Styles: map[string]tcell.Style{
			"CommandBarPrefix": tcell.StyleDefault.
				Background(tcell.ColorRed).
				Foreground(tcell.ColorWhite),
			"CommandBarText": tcell.StyleDefault,
		},
	}
}

type TerminalDisplay struct {
	screen tcell.Screen

	Styles map[string]tcell.Style
}

func (term *TerminalDisplay) WriteText(x int, y int, text string) {
	term.WriteTextStyle(x, y, tcell.StyleDefault, text)
}
func (term *TerminalDisplay) WriteTextStyle(x int, y int, style tcell.Style, text string) {
	for ct, char := range text {
		term.screen.SetCell(x+ct, y, style, char)
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

func (term *TerminalDisplay) DrawCommandBar(command string, cursorPosition int) {
	width, height := term.screen.Size()
	row := height - 2

	for i := 0; i < width; i++ {
		char, _, style, _ := term.screen.GetContent(i, row)
		if char != ' ' || style != tcell.StyleDefault {
			term.screen.SetCell(i, row, tcell.StyleDefault, ' ')
		}
	}

	prefix := "gausfamily#general >"

	// Write what the user is typing
	term.WriteTextStyle(0, row, term.Styles["CommandBarPrefix"], prefix)
	term.WriteTextStyle(len(prefix)+1, row, term.Styles["CommandBarText"], command)

	// Show the cursor at the cursor position
	term.screen.ShowCursor(len(prefix) + 1 + cursorPosition, row)
}

func (term *TerminalDisplay) DrawChannels(conn gateway.Connection) {
	channels, _ := conn.GetChannels()
	for ct, i := range channels {
		term.WriteText(0, ct, i.Name)
	}
}
