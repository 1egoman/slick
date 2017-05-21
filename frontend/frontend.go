package frontend

import (
	// "log"
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji" // convert :smile: to unicode
	"strings"

	"github.com/1egoman/slime/gateway" // The thing to interface with slack
	"github.com/1egoman/slime/status"
)

const BottomPadding = 2 // The amount of lines at the bottom of the window to leave available for status bars.

// Given a string to be displayed in the ui, tokenize the message and return a *PrintableMessage
// that contains each part as a token.
func parseSlackMessage(text string, printableMessage *gateway.PrintableMessage, UserById func(string) (*gateway.User, error)) error {
	text = emoji.Sprintf(text)                         // Emojis
	text = strings.Replace(text, "&amp;", "&", -1)
	text = strings.Replace(text, "&gt;", ">", -1)
	text = strings.Replace(text, "&gt;", "<", -1)

	var parts []gateway.PrintableMessagePart

	// Iterate through each character in the message.
	// Look for tags that look like <%XXXXXXXXX>, where % is a number of symbols and X is [A-Z0-9]
	// If one is found, then turn it into a name and replace it.
	var tagType gateway.PrintableMessagePartType
	startIndex := 0 // Start at the beginning of the message text
	startContentIndex := -1
	for index, char := range text {
		var nextChar rune

		if index + 1 < len(text) - 1 {
			nextChar = rune(text[index+1])
		} else {
			nextChar = ' ' // Placeholder character.
		}

		if char == '<' {
			// Since we just discovered the boundry of the next bit of interest, then add the
			// previous plain text bit (before this tag) to the parts slice.
			if index > 0 {
				parts = append(parts, gateway.PrintableMessagePart{
					Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
					Content: text[startIndex:index],
				})
			}

			startIndex = index
			startContentIndex = index + 2
			if nextChar == '@' { // ie, <@U5FR33U4R> for @foo
				tagType = gateway.PRINTABLE_MESSAGE_AT_MENTION_USER
			} else if nextChar == '!' { // ie, <!UDOU3ENS> for @channel
				tagType = gateway.PRINTABLE_MESSAGE_AT_MENTION_GROUP
			} else if nextChar == '#' { // ie, <#3IDU62ER> for #channel
				tagType = gateway.PRINTABLE_MESSAGE_CHANNEL
			} else {
				tagType = gateway.PRINTABLE_MESSAGE_LINK
				startContentIndex -= 1 // Links don't have a "idenfiying" character, so one less char is needed
			}
		} else if char == '>' && startContentIndex >= 0 {
			content := text[startContentIndex:index]
			metadata := make(map[string]interface{})

			// log.Printf("CONTENT", content, tagType)

			if tagType == gateway.PRINTABLE_MESSAGE_AT_MENTION_USER {
				contentParts := strings.Split(content, "|")
				if len(contentParts) == 1 { // content = ABCDEFGHI
					user, err := UserById(content)
					if err != nil {
						return err
					} else {
						content = "@" + user.Name
					}
				} else if len(contentParts) == 2 { // content = ABCDEFJHI|username
					content = "@" + contentParts[1]
				}
			} else if tagType == gateway.PRINTABLE_MESSAGE_AT_MENTION_GROUP { // content = here / channel / everyone (a group name)
				content = "@" + content
			} else if tagType == gateway.PRINTABLE_MESSAGE_CHANNEL {
				contentParts := strings.Split(content, "|")
				if len(contentParts) == 1 { // content = general
					content = "#" + content
				} else if len(contentParts) == 2 { // content = ABCDEFJHI|general
					content = "#" + contentParts[1]
				}
			} else if tagType == gateway.PRINTABLE_MESSAGE_LINK {
				// Links have meta
				contentParts := strings.Split(content, "|")
				metadata["Href"] = contentParts[0]
				if len(contentParts) == 1 { // content = http://example.com
					content = content // No change
				} else if len(contentParts) == 2 { // content = http://example.com|label
					content = contentParts[1]
				}
			}

			parts = append(parts, gateway.PrintableMessagePart{
				Type: tagType,
				Content: content,
			})

			// Reset the start indicies
			startContentIndex = -1
			startIndex = index + 1
		}
	}

	// Add the final plain text part to the message.
	// bla bla #general foo bar
	//                  ^^^^^^ = This bit
	if len(text) > 0 {
		parts = append(parts, gateway.PrintableMessagePart{
			Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
			Content: text[startIndex:],
		})
	}

	printableMessage.SetParts(parts)
	return nil
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
				Background(tcell.ColorBlue).
				Foreground(tcell.ColorWhite),
			"CommandBarText": tcell.StyleDefault,
			"StatusBarActiveConnection": tcell.StyleDefault.
				Background(tcell.ColorBlue).
				Foreground(tcell.ColorWhite),
			"StatusBarGatewayConnected": tcell.StyleDefault,
			"StatusBarGatewayConnecting": tcell.StyleDefault.
				Background(tcell.ColorDarkMagenta),
			"StatusBarGatewayFailed": tcell.StyleDefault.
				Background(tcell.ColorRed),
			"StatusBarLog": tcell.StyleDefault,
			"StatusBarError": tcell.StyleDefault.
				Foreground(tcell.ColorDarkMagenta).
				Bold(true),
			"StatusBarTopBorder": tcell.StyleDefault.
				Background(tcell.ColorGray),

			"MessageReaction": tcell.StyleDefault,
			"MessageFile":     tcell.StyleDefault,
			"MessageActionHighlight": tcell.StyleDefault.
				Foreground(tcell.ColorRed),
			"MessageAction": tcell.StyleDefault,
			"MessageSelected": tcell.StyleDefault.
				Background(tcell.ColorTeal),
			"MessageAttachmentTitle": tcell.StyleDefault.
				Foreground(tcell.ColorGreen),
			"MessageAttachmentFieldTitle": tcell.StyleDefault.
				Bold(true),
			"MessageAttachmentFieldValue": tcell.StyleDefault,
			"MessagePartAtMentionUser": tcell.StyleDefault.
				Foreground(tcell.ColorRed).
				Bold(true),
			"MessagePartAtMentionGroup": tcell.StyleDefault.
				Foreground(tcell.ColorYellow).
				Bold(true),
			"MessagePartChannel": tcell.StyleDefault.
				Foreground(tcell.ColorBlue).
				Bold(true),
			"MessagePartLink": tcell.StyleDefault.
				Foreground(tcell.ColorDarkCyan).
				Underline(true).
				Bold(true),

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
	width, height := term.screen.Size()
	lastRow := height - 1

	// If the status bar needs to display mutiple lines to view a larger, multiline message, then
	// adjust the last row acccordingly.
	messages := strings.Split(stat.Message, "\n")
	if stat.Show && len(messages) > 1 {
		lastRow -= len(messages) - 1 // For each additional message line
		lastRow -= 1 // For the "Press any key to continue" bit
	}

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
		if len(messages) == 1 {
			// Just render one row, nothing special.
			term.WriteTextStyle(position, lastRow, style, stat.Message)
		} else {
			// Rendering multiple rows is more involved.
			// Above the top of the picker, draw a border.
			for i := 0; i < width; i++ {
				term.screen.SetCell(i, lastRow - 1, term.Styles["StatusBarTopBorder"], ' ')
			}
			for ct, line := range messages {
				term.DrawBlankLine(lastRow + ct)
				term.WriteTextStyle(position, lastRow+ct, style, line)
			}
			term.WriteTextStyle(position, lastRow+len(messages), style, "Press any key to continue...")
		}
	} else {
		// Otherwise, render each conenction
		for index, item := range connections {
			// How should the connection look?
			var style tcell.Style
			if item == activeConnection {
				style = term.Styles["StatusBarActiveConnection"]
			} else if item.Status() == gateway.CONNECTING {
				style = term.Styles["StatusBarGatewayConnecting"]
			} else if item.Status() == gateway.CONNECTED {
				style = term.Styles["StatusBarGatewayConnected"]
			} else if item.Status() == gateway.FAILED {
				style = term.Styles["StatusBarGatewayFailed"]
			} else {
				// SOme weird case has no coloring.
				style = tcell.StyleDefault
			}

			// Draw each connection
			label := fmt.Sprintf("%d: %s", index+1, item.Name())
			term.WriteTextStyle(position, lastRow, style, label)
			position += len(label) + 1
		}

		// And the users that are currently typing
		if activeConnection != nil && activeConnection.TypingUsers() != nil {
			// Create a slice of user names from the slice of users
			userList := activeConnection.TypingUsers().Users()

			var typingUsers string
			if len(userList) > 0 {
				if len(userList) <= 3 {
					typingUsers = fmt.Sprintf("%s typing...", strings.Join(userList, ", "))
				} else {
					typingUsers = "Several typing..."
				}
				typingUsersXPos := width - len(typingUsers) - 1
				term.WriteTextStyle(typingUsersXPos, lastRow, tcell.StyleDefault, typingUsers)
			}
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
