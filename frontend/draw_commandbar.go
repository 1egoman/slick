package frontend

import (
	// "log"
	// "strings"

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
	width, height := term.screen.Size()

	// Generate prefix for given team and channel
	prefix := currentTeamName
	if currentChannel != nil {
		prefix += "#" + currentChannel.Name

		if currentChannel.IsMember == false {
			prefix += " (not a member)"
		}
		if currentChannel.IsArchived {
			prefix += " (archived)"
		}
	}
	prefix += " >"

	// How many chars at max can be in a line?
	maxLineWidth := width - 2 - len(prefix) - 1

	commandLines := make([]string, 0)
	lastBreak := 0
	currentLineIndex := 0
	for index, char := range command {
		if char == '\n' || currentLineIndex > maxLineWidth {
			commandLines = append(commandLines, command[lastBreak:index+1])
			lastBreak = index+1
			currentLineIndex = 0
		} else {
			currentLineIndex += 1
		}
	}
	commandLines = append(commandLines, command[lastBreak:])
	// commandLines := strings.Split(command, "\n")

	row := height - 1 - len(commandLines)

	// Clear the rows.
	term.DrawBlankLines(row, row+len(commandLines)-1)

	// Write what the user is typing
	term.WriteTextStyle(0, row, color.DeSerializeStyleTcell(config["CommandBar.PrefixColor"]), prefix)
	for index, line := range commandLines {
		term.WriteTextStyle(
			len(prefix)+1,
			row+index,
			color.DeSerializeStyleTcell(config["CommandBar.TextColor"]),
			line,
		)

		// Render a backslash before each line.
		if len(line) > 0 && line[len(line)-1] == '\n' {
			term.WriteTextStyle(
				len(prefix)+1+len(line)-1,
				row+index,
				color.DeSerializeStyleTcell(config["CommandBar.NewLineColor"]),
				"\\",
			)
		}
	}

	// Show the cursor at the cursor position
	x := 0
	y := 0
	totalCharactersCounted := 0
	for yPos, line := range commandLines {
		// Is the cursor on the given line of the output?
		if cursorPosition <= totalCharactersCounted + len(line) {
			// If so, the y position is the index of the line, and the x positino is the difference
			// in the number of the characters traversed and the actual cursor position.
			y = yPos
			x = cursorPosition - totalCharactersCounted

			// When thre's a newline at the end of the current line, render on the next line.
			if cursorPosition == totalCharactersCounted + len(line) && len(line) > 0 && line[len(line) - 1] == '\n' {
				y += 1
				x = 0
			}
			break
		}
		totalCharactersCounted += len(line)
	}
	term.screen.ShowCursor(len(prefix)+1+x, row+y)
}
