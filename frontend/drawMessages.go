package frontend

import (
	// "log"
	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji" // convert :smile: to unicode
	"strings"

	"github.com/1egoman/slime/gateway" // The thing to interface with slack
)

// Given an array of reactions and a row to render them on, render them.
func renderReactions(term *TerminalDisplay, reactions []gateway.Reaction, row int, leftOffset int) {
	reactionOffset := leftOffset

	// Add a prefix to the reactino list
	term.WriteText(reactionOffset, row, "| ")
	reactionOffset += 2

	for _, reaction := range reactions {
		// Render the reaction
		reactionEmoji := emoji.Sprintf("%d :"+reaction.Name+":", len(reaction.Users))

		// Draw the reaction
		term.WriteTextStyle(
			reactionOffset,
			row, // Last row, below the message
			term.Styles["MessageReaction"],
			reactionEmoji,
		)

		// Offset the next reaction.
		reactionOffset += len(reactionEmoji) + 1
	}
}

// Given a message, return the sender's name and the color to make the sender's name
func getSenderInfo(msg gateway.Message) (string, tcell.Style) {
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
	return sender, senderStyle
}

// Draw message history in the channel
func (term *TerminalDisplay) DrawMessages(messages []gateway.Message) {
	width, height := term.screen.Size()

	for r := 0; r < height-bottomPadding; r++ {
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
	row := height - bottomPadding - 1
	for row >= 0 && index >= 0 {
		msg := messages[index]

		// Get sender information
		sender, senderStyle := getSenderInfo(msg)

		// Calculate how many rows the message requires to render.
		messageColumnWidth := width - len(sender) - 1
		messageRows := (len(makePrintWorthy(msg.Text)) / messageColumnWidth) + 1
		if len(msg.Reactions) > 0 {
			messageRows += 1
		}

		// Render the sender and the message
		for rowDelta, messageRow := range partitionIntoRows(
			makePrintWorthy(msg.Text), // Our message to render to the screen
			messageColumnWidth,        // The width of each message row
		) {
			if rowDelta == 0 {
				// Draw the sender on the first row of a message
				term.WriteTextStyle(0, row-messageRows+1, senderStyle, sender)

				// Render reactions after message
				if len(msg.Reactions) > 0 {
					renderReactions(
						term,
						msg.Reactions, // Reactions to render
						row,           // Which row should the reactions be rendered on?
						len(sender)+1, // How far from the left should the first reaction be offset?
					)
				}
			}

			// Render the message row.
			term.WriteTextStyle(
				len(sender)+1,
				row-messageRows+rowDelta,
				tcell.StyleDefault,
				strings.Trim(messageRow, " "),
			)
		}

		// Subtract the message's height.
		row -= messageRows
		index -= 1
	}
}
