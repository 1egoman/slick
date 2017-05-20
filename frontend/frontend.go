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

type PrintableMessagePartType int

const (
	PLAIN_TEXT PrintableMessagePartType = iota
	AT_MENTION_USER  // (like @foo, @bar, etc)
	AT_MENTION_GROUP // (like @channel, @here, etc)
	CHANNEL          // (like #general)
	CONNECTION       // (like "my custom slack team")
)

type PrintableMessagePart struct {
	Type PrintableMessagePartType
	Content string
}

type PrintableMessage struct {
	parts []PrintableMessagePart

	// Running `Lines(width)` takes a while, and it's pure. Memoize it for speed.
	linesCacheWidth int
	linesCache [][]PrintableMessagePart
}

func (p *PrintableMessage) Parts() []PrintableMessagePart {
	return p.parts
}

func (p *PrintableMessage) SetParts(parts []PrintableMessagePart) {
	p.parts = parts
}

// Return the length of the printable message
func (p *PrintableMessage) Length() int {
	length := 0
	for _, part := range p.parts {
		length += len(part.Content)
	}
	return length
}

// Return the printable message converted to plain text string.
func (p *PrintableMessage) Plain() string {
	total := ""
	for _, part := range p.parts {
		total += part.Content
	}
	return total
}

func (p *PrintableMessage) Lines(width int) [][]PrintableMessagePart {
	// If the result has already been calculated, use it.
	if p.linesCacheWidth == width && p.linesCache != nil {
		return p.linesCache
	}

	var lines [][]PrintableMessagePart
	var messageParts []PrintableMessagePart
	var extraWords []string

	for _, part := range p.parts {
		messageParts = append(messageParts, part)
		// log.Printf("MESSAGE PARTS %+v", messageParts)

		pm := PrintableMessage{parts: messageParts}
		if pm.Length() > width {
			pm = PrintableMessage{parts: messageParts[:len(messageParts) - 1]}
			maximumLengthOfLastMessagePart := width - pm.Length()

			// log.Printf("WANTED TO DRAW %+v BUT INSTEAD ONLY DID %d", messageParts, maximumLengthOfLastMessagePart)

			// Now, cut down all words that don't fit on this line.
			// foo bar baz hello world test quux (extraWords = [])
			// foo bar baz hello world test      (extraWords = [quux])
			// foo bar baz hello world           (extraWords = [test, quux])
			// foo bar baz hello                 (extraWords = [world, test, quux])
			//                    ^
			//                 "window width"
			// (done, since the line length < window width)

			wordsInLastMessagePart := strings.Split(part.Content, " ")
			for len(strings.Join(wordsInLastMessagePart, " ")) > maximumLengthOfLastMessagePart {
				// Remove one word from the last message part until the 
				extraWords = append([]string{wordsInLastMessagePart[len(wordsInLastMessagePart) - 1]}, extraWords...)
				wordsInLastMessagePart = wordsInLastMessagePart[:len(wordsInLastMessagePart)-1]
			}
			messageParts[len(messageParts)-1].Content = strings.Join(wordsInLastMessagePart, " ")

			// log.Printf("ACTUALLY DRAWING %+v", messageParts)

			// Now that our line is below the max width, it can be drawn, so add it to the array.
			lines = append(lines, messageParts)
			// log.Printf("LINES %+v", lines)

			// Then, clear out the message parts that have been used so far.
			messageParts = make([]PrintableMessagePart, 0)

			// And start off the next line with any words that were removed from the current line to
			// make it fit.
			if len(extraWords) > 0 {
				messageParts = append(messageParts, PrintableMessagePart{
					Type: part.Type,
					Content: strings.Join(extraWords, " "),
				})
				extraWords = make([]string, 0)
			}
			// log.Printf("LINE BIT CUT OFF %+v", messageParts)
		}
	}

	// log.Printf("DONE LOOPING %+v", messageParts)

	// If there are any message parts left over, then add them at the end.
	if len(messageParts) > 0 {
		lines = append(lines, messageParts)
	}

	// log.Printf("RETURNING %+v", lines)
	p.linesCache = lines
	p.linesCacheWidth = width
	return lines
}
// Given a string to be displayed in the ui, tokenize the message and return a *PrintableMessage
// that contains each part as a token.
func parseSlackMessage(text string, printableMessage *PrintableMessage, UserById func(string) (*gateway.User, error)) error {
	text = emoji.Sprintf(text)                         // Emojis

	var parts []PrintableMessagePart

	// Iterate through each character in the message.
	// Look for tags that look like <%XXXXXXXXX>, where % is a number of symbols and X is [A-Z0-9]
	// If one is found, then turn it into a name and replace it.
	var tagType PrintableMessagePartType
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
				parts = append(parts, PrintableMessagePart{
					Type: PLAIN_TEXT,
					Content: text[startIndex:index - 1],
				})
			}

			startIndex = index
			startContentIndex = index + 2
			if nextChar == '@' { // ie, <@U5FR33U4R> for @foo
				tagType = AT_MENTION_USER
			} else if nextChar == '!' { // ie, <!UDOU3ENS> for @channel
				tagType = AT_MENTION_GROUP
			} else if nextChar == '#' { // ie, <#3IDU62ER> for #channel
				tagType = CHANNEL
			}
		} else if char == '>' && startContentIndex >= 0 {
			content := text[startContentIndex:index]
			// log.Printf("CONTENT", content, tagType)

			if tagType == AT_MENTION_USER {
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
			} else if tagType == AT_MENTION_GROUP { // content = here / channel / everyone (a group name)
				content = "@" + content
			} else if tagType == CHANNEL {
				contentParts := strings.Split(content, "|")
				if len(contentParts) == 1 { // content = general
					content = "#" + content
				} else if len(contentParts) == 2 { // content = ABCDEFJHI|general
					content = "#" + contentParts[1]
				}
			}

			parts = append(parts, PrintableMessagePart{
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
		parts = append(parts, PrintableMessagePart{
			Type: PLAIN_TEXT,
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
