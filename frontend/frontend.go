package frontend

import (
	"log"
	"fmt"
	"regexp"
	"strings"
	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji" // convert :smile: to unicode

	"github.com/1egoman/slime/gateway" // The thing to interface with slack
)

const bottomPadding = 2; // The amount of lines at the bottom of the window to leave available for status bars.

var usernameRegex = regexp.MustCompile("<@[A-Z0-9]+\\|(.+)>")
var channelRegex = regexp.MustCompile("<#[A-Z0-9]+\\|(.+)>")

// Given a string to be displayed in the ui, convert it to be printable.
// 1. Convert emoji codes like :smile: to their emojis.
// 2. Replace any <@ID|username> tags with @username
// 2. Replace any <#ID|channel> tags with #channel
func makePrintWorthy(text string) string {
	text = emoji.Sprintf(text) // Emojis
	text = usernameRegex.ReplaceAllString(text, "@$1") // Usernames
	text = channelRegex.ReplaceAllString(text, "#$1") // Channels
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
	term.WriteText(len(mode) + 1, lastRow, "|")

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

func (term *TerminalDisplay) DrawFuzzyPicker(items []string, selectedIndex int) {
	width, height := term.screen.Size()
	bottomPadding := 2 // pad for the status bar and command bar
	startingRow := height - len(items) - bottomPadding // The top row of the fuzzy picker

	for ct, item := range items {
		row := startingRow + ct

		// Clear the row.
		for i := 0; i < width; i++ {
			char, _, style, _ := term.screen.GetContent(i, row)
			if char != ' ' || style != tcell.StyleDefault {
				term.screen.SetCell(i, row, tcell.StyleDefault, ' ')
			}
		}

		// Add prefix for selected item
		if ct == selectedIndex {
			item = "> "+item
		}

		// Draw item
		term.WriteText(0, row, item)
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



// Draw message history in the channel
func (term *TerminalDisplay) DrawMessages(messages []gateway.Message) {
	width, height := term.screen.Size()
	log.Println("start")


	for r := 0; r < height - bottomPadding; r++ {
		// Clear the row.
		for i := 0; i < width; i++ {
			char, _, style, _ := term.screen.GetContent(i, r)
			if char != ' ' || style != tcell.StyleDefault {
				term.screen.SetCell(i, r, tcell.StyleDefault, ' ')
			}
		}
	}


	// Loop from the bottom of the window to the top.
	index := len(messages) - 1
	row := height - 2
	for row > 0 && index >= 0 {
		msg := messages[index]

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

		// Calculate how many rows the message requires to render.
		messageColumnWidth := width - len(sender) - 1
		messageRows := (len(makePrintWorthy(msg.Text)) / messageColumnWidth) + 1
		if len(msg.Reactions) > 0 {
			messageRows += 1
		}
		log.Println("Message rows", messageRows)

		// Render the sender and the message
		for rowDelta, messageRow := range partitionIntoRows(
			makePrintWorthy(msg.Text), // Our message to render to the screen
			messageColumnWidth, // The width of each message row
		) {
			if rowDelta == 0 {
				// Draw the sender on the first row of a message
				term.WriteTextStyle(0, row - messageRows + 1, senderStyle, sender)

				// Render reactions after message
				if len(msg.Reactions) > 0 {
					reactionOffset := len(sender) + 1
					for _, reaction := range msg.Reactions {
						reactionEmoji := emoji.Sprintf("%d :"+reaction.Name+":", len(reaction.Users))
						term.WriteText(reactionOffset, row, reactionEmoji)
						reactionOffset += len(reactionEmoji)
					}
				}
			}

			// Render the message row.
			term.WriteTextStyle(
				len(sender) + 1,
				row - messageRows + rowDelta,
				tcell.StyleDefault,
				strings.Trim(messageRow, " "),
			)
		}

		// Subtract the message's height.
		row -= messageRows;
		index -= 1;
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




// Given a string, partition into sections of size `width`.
func partitionIntoRows(total string, width int) []string {
	partitions := []string{}
	lastIndex := 0
	for index := 0; index < len(total); index++ {
		if index % width == 0 {
			partitions = append(partitions, total[lastIndex:index])
			lastIndex = index
		}
	}

	// Add the last undersized partition if it exists
	if lastIndex < len(total) - 1 {
		partitions = append(partitions, total[lastIndex:])
	}

	return partitions
}
