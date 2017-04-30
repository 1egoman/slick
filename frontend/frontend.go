package frontend

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji" // convert :smile: to unicode
	"regexp"

	"github.com/1egoman/slime/gateway" // The thing to interface with slack
)

const bottomPadding = 2 // The amount of lines at the bottom of the window to leave available for status bars.

var usernameRegex = regexp.MustCompile("<@[A-Z0-9]+\\|(.+)>")
var channelRegex = regexp.MustCompile("<#[A-Z0-9]+\\|(.+)>")

// Given a string to be displayed in the ui, convert it to be printable.
// 1. Convert emoji codes like :smile: to their emojis.
// 2. Replace any <@ID|username> tags with @username
// 2. Replace any <#ID|channel> tags with #channel
func makePrintWorthy(text string) string {
	text = emoji.Sprintf(text)                         // Emojis
	text = usernameRegex.ReplaceAllString(text, "@$1") // Usernames
	text = channelRegex.ReplaceAllString(text, "#$1")  // Channels
	return text
}

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
			"StatusBarActiveConnection": tcell.StyleDefault.
				Background(tcell.ColorBlue).
				Foreground(tcell.ColorWhite),
			"StatusBarConnection": tcell.StyleDefault,

			"MessageReaction":     tcell.StyleDefault,
			"MessageSelected":     tcell.StyleDefault.
				Background(tcell.ColorTeal),

			"FuzzyPickerTopBorder": tcell.StyleDefault.
				Background(tcell.ColorGray),
			"FuzzyPickerActivePrefix": tcell.StyleDefault,
			"FuzzyPickerChannelNotMember": tcell.StyleDefault.
				Foreground(tcell.ColorGray),
		},
	}
}

type TerminalDisplay struct {
	screen tcell.Screen

	Styles map[string]tcell.Style
}

func (term *TerminalDisplay) Screen() tcell.Screen {
	return term.screen
}

func (term *TerminalDisplay) Render() {
	term.screen.Show()
}

func (term *TerminalDisplay) DrawStatusBar(mode string, connections []gateway.Connection, activeConnection gateway.Connection) {
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
	term.WriteText(len(mode)+1, lastRow, "|")

	// Then, render each conenction
	position := len(mode) + 3
	for index, item := range connections {
		// How should the connection look?
		var style tcell.Style
		if item == activeConnection {
			style = term.Styles["StatusBarActiveConnection"]
		} else {
			style = term.Styles["StatusBarConnection"]
		}

		// Draw each connection
		label := fmt.Sprintf("%d: %s", index+1, item.Name())
		term.WriteTextStyle(position, lastRow, style, label)
		position += len(label) + 1
	}
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

func (term *TerminalDisplay) WriteText(x int, y int, text string) {
	term.WriteTextStyle(x, y, tcell.StyleDefault, text)
}
func (term *TerminalDisplay) WriteTextStyle(x int, y int, style tcell.Style, text string) {
	for ct, char := range text {
		term.screen.SetCell(x+ct, y, style, char)
	}
}
