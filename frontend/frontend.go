package frontend

import (
	"fmt"
	"github.com/gdamore/tcell"

	"github.com/1egoman/slime/gateway" // The thing to interface with slack
)

type Display interface {
	DrawStatusBar()
	DrawCommandBar(string, int)
	DrawChannels([]gateway.Channel)
	Render()
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

func (term *TerminalDisplay) Render() {
	term.screen.Show()
}

func (term *TerminalDisplay) DrawStatusBar(mode string) {
	width, height := term.screen.Size()
	lastRow := height - 1

	// Clear the row.
	for i := 0; i < width; i++ {
		char, _, style, _ := term.screen.GetContent(i, lastRow)
		if char != ' ' || style != tcell.StyleDefault {
			term.screen.SetCell(i, lastRow, tcell.StyleDefault, ' ')
		}
	}

	// First, draw the mode (ie, chat, channel-picker, etc...)
	term.WriteText(0, lastRow, mode)

	// Then, draw a seperator
	term.WriteText(len(mode) + 1, lastRow, "|")
}

func (term *TerminalDisplay) DrawCommandBar(
	command string,
	cursorPosition int,
	currentChannel *gateway.Channel,
	currentTeam *gateway.Team,
) {
	width, height := term.screen.Size()
	row := height - 2

	// Clear the row.
	for i := 0; i < width; i++ {
		char, _, style, _ := term.screen.GetContent(i, row)
		if char != ' ' || style != tcell.StyleDefault {
			term.screen.SetCell(i, row, tcell.StyleDefault, ' ')
		}
	}

	var prefix string
	if currentTeam != nil && currentChannel != nil {
		prefix = currentTeam.Name + "#" + currentChannel.Name + " >"
	} else {
		prefix = "(loading...) >"
	}

	// Write what the user is typing
	term.WriteTextStyle(0, row, term.Styles["CommandBarPrefix"], prefix)
	term.WriteTextStyle(len(prefix)+1, row, term.Styles["CommandBarText"], command)

	// Show the cursor at the cursor position
	term.screen.ShowCursor(len(prefix)+1+cursorPosition, row)
}

// Draw message history in the channel
func (term *TerminalDisplay) DrawMessages(messages []gateway.Message) {
	width, _ := term.screen.Size()
	for row, msg := range messages {
		// Clear the row.
		for i := 0; i < width; i++ {
			char, _, style, _ := term.screen.GetContent(i, row)
			if char != ' ' || style != tcell.StyleDefault {
				term.screen.SetCell(i, row, tcell.StyleDefault, ' ')
			}
		}

		// Get the name of the sender, and the sender's color
		sender := "(anon)"
		senderStyle := tcell.StyleDefault
		if msg.Sender != nil {
			sender = msg.Sender.Name
			// If the sender has a color associated, use that!
			if len(msg.Sender.Color) > 0 {
				senderStyle = senderStyle.Foreground(tcell.GetColor("#" + msg.Sender.Color))
			}
		}

		sender = fmt.Sprintf("%d %s", row, sender)

		// Write sender and message to the screen
		term.WriteTextStyle(0, row, senderStyle, sender)
		term.WriteText(len(sender)+1, row, msg.Text)
	}
}

func (term *TerminalDisplay) WriteText(x int, y int, text string) {
	term.WriteTextStyle(x, y, tcell.StyleDefault, text)
}
func (term *TerminalDisplay) WriteTextStyle(x int, y int, style tcell.Style, text string) {
	for ct, char := range text {
		term.screen.SetCell(x+ct, y, style, char)
	}
}
