package frontend

import (
	// "log"
	"strings"
  "time"

	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji" // convert :smile: to unicode

	"github.com/1egoman/slime/gateway" // The thing to interface with slack
)

// Given an array of reactions and a row to render them on, render them.
func renderReactions(term *TerminalDisplay, reactions []gateway.Reaction, row int, leftOffset int) {
	reactionOffset := leftOffset

	// Add a prefix to the reaction list
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

func renderAttachments(term *TerminalDisplay, attachments []gateway.Attachment, row int, leftOffset int) {
  // TODO: write me?
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

// Given a string, partition into sections of size `width`.
func partitionIntoRows(total string, width int) []string {
	partitions := []string{}
	lastIndex := 0
	for index := 0; index < len(total); index++ {
		if index%width == 0 {
			partitions = append(partitions, total[lastIndex:index])
			lastIndex = index
		}
	}

	// Add the last undersized partition if it exists
	if lastIndex < len(total)-1 {
		partitions = append(partitions, total[lastIndex:])
	}

	return partitions
}

// Draw message history in the channel
func (term *TerminalDisplay) DrawMessages(
  messages []gateway.Message, // A list of messages to render
  selectedMessageIndex int, // Index of selected message (-1 for no selected message)
) {
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

    timestamp := time.Unix(int64(msg.Timestamp), 0).Format("15:04:05")
    prefixWidth := len(timestamp) + 1 + len(sender) + 1

    // Is the message selected?
    var messageStyle tcell.Style
    if index == selectedMessageIndex {
      messageStyle = term.Styles["MessageSelected"]
    } else {
      messageStyle = tcell.StyleDefault
    }

		// Calculate how many rows the message requires to render.
		messageColumnWidth := width - prefixWidth
		messageRows := (len(makePrintWorthy(msg.Text)) / messageColumnWidth) + 1
		if len(msg.Reactions) > 0 {
			messageRows += 1
		}
    messageRows += len(msg.Attachments)

		// Render the sender and the message
		for rowDelta, messageRow := range partitionIntoRows(
			makePrintWorthy(msg.Text), // Our message to render to the screen
			messageColumnWidth,        // The width of each message row
		) {
			if rowDelta == 0 {
				// Draw the sender on the first row of a message
				term.WriteText(0, row-messageRows+1, timestamp)
				term.WriteTextStyle(len(timestamp) + 1, row-messageRows+1, senderStyle, sender)

				// Render reactions after message
				if len(msg.Reactions) > 0 {
					renderReactions(
						term,
						msg.Reactions, // Reactions to render
						row,           // Which row should the reactions be rendered on?
						prefixWidth, // How far from the left should the first reaction be offset?
					)
				}
			}

			// Render the message row.
			term.WriteTextStyle(
				prefixWidth, // Sender and timestamp go before message.
				row-messageRows+rowDelta,
				messageStyle,
				strings.Trim(messageRow, " "),
			)
		}

		// Subtract the message's height.
		row -= messageRows
		index -= 1
	}
}
