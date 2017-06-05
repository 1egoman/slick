package frontend

import (
	// "log"
	"fmt"
	"github.com/gdamore/tcell"
	"strings"

	"github.com/1egoman/slick/color"
	"github.com/1egoman/slick/gateway" // The thing to interface with slack
	"github.com/1egoman/slick/status"
)

func (term *TerminalDisplay) DrawStatusBar(
	mode string,
	connections []gateway.Connection,
	activeConnection gateway.Connection,
	stat status.Status,
	config map[string]string,
) {
	width, height := term.screen.Size()
	lastRow := height - 1

	// If the status bar needs to display mutiple lines to view a larger, multiline message, then
	// adjust the last row acccordingly.
	messages := strings.Split(stat.Message, "\n")
	if stat.Show && len(messages) > 1 {
		lastRow -= len(messages) - 1 // For each additional message line
		lastRow -= 1                 // For the "Press any key to continue" bit
	}

	// Clear the row.
	defaultColor := color.DeSerializeStyleTcell(config["StatusBar.Color"])
	for j := 0; j < width; j++ {
		char, _, style, _ := term.screen.GetContent(j, lastRow)
		if char != ' ' || style != defaultColor {
			term.screen.SetCell(j, lastRow, defaultColor, ' ')
		}
	}

	// First, draw the mode (ie, chat, channel-picker, etc...)
	term.WriteTextStyle(
		0, lastRow,
		color.DeSerializeStyleTcell(config["StatusBar.ModeColor"]),
		mode,
	)

	// Then, draw a separator
	term.WriteTextStyle(len(mode)+1, lastRow, defaultColor, "|")

	position := len(mode) + 3

	if stat.Show {
		// Get the color of the text on the status bar
		var style tcell.Style
		if stat.Type == status.STATUS_ERROR {
			style = color.DeSerializeStyleTcell(config["StatusBar.ErrorColor"])
		} else {
			style = color.DeSerializeStyleTcell(config["StatusBar.LogColor"])
		}

		// Write status text
		if len(messages) == 1 {
			// Just render one row, nothing special.
			term.WriteTextStyle(position, lastRow, style, stat.Message)
		} else {
			// Rendering multiple rows is more involved.
			// Above the top of the picker, draw a border.
			for i := 0; i < width; i++ {
				term.screen.SetCell(i, lastRow-1, color.DeSerializeStyleTcell(config["StatusBar.TopBorderColor"]), ' ')
			}
			for ct, line := range messages {
				term.DrawBlankLine(lastRow + ct)
				term.WriteTextStyle(position, lastRow+ct, style, line)
			}
			term.WriteTextStyle(position, lastRow+len(messages), style, "Press any key to continue...")
		}
	} else {
		// Otherwise, render each connection
		for index, item := range connections {
			// How should the connection look?
			var style tcell.Style
			if item.Status() == gateway.CONNECTING {
				style = color.DeSerializeStyleTcell(config["StatusBar.GatewayConnectingColor"])
			} else if item.Status() == gateway.FAILED {
				style = color.DeSerializeStyleTcell(config["StatusBar.GatewayFailedColor"])
			} else if item == activeConnection {
				style = color.DeSerializeStyleTcell(config["StatusBar.ActiveConnectionColor"])
			} else if item.Status() == gateway.CONNECTED {
				style = color.DeSerializeStyleTcell(config["StatusBar.GatewayConnectedColor"])
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
