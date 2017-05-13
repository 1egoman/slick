package frontend

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji" // convert :smile: to unicode
	"regexp"

	"github.com/1egoman/slime/gateway" // The thing to interface with slack
	"github.com/1egoman/slime/status"
)

const BottomPadding = 2 // The amount of lines at the bottom of the window to leave available for status bars.

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

	for i := start; i < end; i++ {
		term.DrawBlankLine(i)
	}
}

type Display interface {
	DrawStatusBar()
	DrawCommandBar(string, int)
	DrawChannels([]gateway.Channel)
	Render()
	Close()
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
			"StatusBarConnected": tcell.StyleDefault,
			"StatusBarConnecting": tcell.StyleDefault.
				Background(tcell.ColorDarkMagenta),
			"StatusBarLog": tcell.StyleDefault,
			"StatusBarError": tcell.StyleDefault.
				Foreground(tcell.ColorDarkMagenta).
				Bold(true),

			"MessageReaction": tcell.StyleDefault,
			"MessageFile":     tcell.StyleDefault,
			"MessageActionHighlight": tcell.StyleDefault.
				Foreground(tcell.ColorRed),
			"MessageAction": tcell.StyleDefault,
			"MessageSelected": tcell.StyleDefault.
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

func (term *TerminalDisplay) Close() {
	term.screen.Fini()
}

func (term *TerminalDisplay) DrawStatusBar(
	mode string,
	connections []gateway.Connection,
	activeConnection gateway.Connection,
	stat status.Status,
) {
	_, height := term.screen.Size()
	lastRow := height - 1

	// Clear the row.
	term.DrawBlankLine(lastRow)

	// First, draw the mode (ie, chat, channel-picker, etc...)
	term.WriteText(0, lastRow, mode)

	// Then, draw a seperator
	term.WriteText(len(mode)+1, lastRow, "|")

	position := len(mode) + 3

	if stat.Show {
		// Get the color of the text on the status bar
		var style tcell.Style
		if stat.Type == status.STATUS_ERROR {
			style = term.Styles["StatusBarError"]
		} else {
			style = term.Styles["StatusBarLog"]
		}

		// Write status text
		term.WriteTextStyle(position, lastRow, style, stat.Message)
	} else {
		// Then, render each conenction
		for index, item := range connections {
			// How should the connection look?
			var style tcell.Style
			if item == activeConnection {
				style = term.Styles["StatusBarActiveConnection"]
			} else if item.Status() == gateway.CONNECTING {
				style = term.Styles["StatusBarConnecting"]
			} else if item.Status() == gateway.CONNECTED {
				style = term.Styles["StatusBarConnected"]
			} else {
				// SOme weird case has no coloring.
				style = tcell.StyleDefault
			}

			// Draw each connection
			label := fmt.Sprintf("%d: %s", index+1, item.Name())
			term.WriteTextStyle(position, lastRow, style, label)
			position += len(label) + 1
		}
	}
}

func (term *TerminalDisplay) DrawCommandBar(
	command string,
	cursorPosition int,
	currentChannel *gateway.Channel,
	currentTeamName string,
) {
	_, height := term.screen.Size()
	row := height - 2

	// Clear the row.
	term.DrawBlankLine(row)

	// Generate prefix for given team and channel
	prefix := currentTeamName
	if currentChannel != nil {
		prefix += "#" + currentChannel.Name
	}
	prefix += " >"

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
