package frontend

import (
	"strings"

	"github.com/1egoman/slick/color"
	"github.com/1egoman/slick/gateway" // The thing to interface with slack
)

func (term *TerminalDisplay) DrawCommandBar(
	command string,
	cursorPosition int,
	currentChannel *gateway.Channel,
	currentTeamName string,
	config map[string]string,
) {
	_, height := term.screen.Size()
	commandLines := strings.Split(command, "\n")
	row := height - 1 - len(commandLines)

	// Clear the rows.
	term.DrawBlankLines(row, row+len(commandLines)-1)

	// Generate prefix for given team and channel
	prefix := currentTeamName
	if currentChannel != nil {
		prefix += "#" + currentChannel.Name
	}
	prefix += " >"

	// Write what the user is typing
	term.WriteTextStyle(0, row, color.DeSerializeStyleTcell(config["CommandBar.PrefixColor"]), prefix)
	for index, line := range commandLines {
		term.WriteTextStyle(
			len(prefix)+1,
			row+index,
			color.DeSerializeStyleTcell(config["CommandBar.TextColor"]),
			line,
		)
	}

	// Show the cursor at the cursor position
	x := 0
	y := 0
	for index, item := range command {
		if index >= cursorPosition {
			break
		}

		// Newlines reset the cursor back to the starting position
		if item == '\n' {
			y += 1
			x = 0
		} else {
			x += 1
		}
	}
	term.screen.ShowCursor(len(prefix)+1+x, row+y)
}
