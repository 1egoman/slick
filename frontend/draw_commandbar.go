package frontend

import (
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
	term.WriteTextStyle(0, row, color.DeSerializeStyleTcell(config["CommandBar.PrefixColor"]), prefix)
	term.WriteTextStyle(len(prefix)+1, row, color.DeSerializeStyleTcell(config["CommandBar.TextColor"]), command)

	// Show the cursor at the cursor position
	term.screen.ShowCursor(len(prefix)+1+cursorPosition, row)
}

