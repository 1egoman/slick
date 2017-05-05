package frontend

import (
	// "log"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji" // convert :smile: to unicode

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

const abbrevitionThreshhold = 30

func formatAbreviatedLink(link string) string {
	if strings.HasPrefix(link, "https://") {
		link = link[8:]
	}
	if strings.HasPrefix(link, "http://") {
		link = link[7:]
	}
	if len(link) < abbrevitionThreshhold {
		return link
	}

	return fmt.Sprintf(
		"%s...%s",
		link[:abbrevitionThreshhold/2],
		link[len(link)-(abbrevitionThreshhold/2):],
	)
}

// Given a pointer to a file and a row to render it on, render it.
func renderFile(term *TerminalDisplay, file *gateway.File, isSelected bool, row int, leftOffset int) {
	if file != nil {
		var messageStyle tcell.Style
		var messageActions []string
		if isSelected {
			messageStyle = term.Styles["MessageSelected"]
			messageActions = []string{"Open", "Copy"}
		} else {
			messageStyle = term.Styles["MessageFile"]
		}

		fileRow := fmt.Sprintf(
			"| %s: %s",
			file.Name,
			formatAbreviatedLink(file.Permalink),
		)
		term.WriteTextStyle(leftOffset, row, messageStyle, fileRow)

		// Render actions after the file
		messageActionOffset := leftOffset + len(fileRow) + 1 // Add a space netween file and actions
		renderActions(term, messageActions, messageActionOffset, row)
	}
}

// Render actions that can be done with the message
// Uppercase letters in the actions are highlighted in a different color (they are the key to
// press to do the thing)
func renderActions(term *TerminalDisplay, actions []string, leftOffset int, row int) {
	for _, action := range actions {
		for _, char := range action {
			if char >= 'A' && char <= 'Z' {
				term.WriteTextStyle(leftOffset, row, term.Styles["MessageActionHighlight"], string(char))
			} else {
				term.WriteTextStyle(leftOffset, row, term.Styles["MessageAction"], string(char))
			}
			leftOffset += 1 // Add one space between actions
		}
		leftOffset += 1 // Add one space between actions
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
	bottomDisplayedItem int, // The bottommost message. If 0, bottommost message is most recent.
) int { // Return how many messages were rendered.
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
	// Start at the `bottomDisplayedItem`th item and loop until no more items can be rendered.
	index := len(messages) - 1 - bottomDisplayedItem
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
		accessoryRow := row         // The row to start rendering "message accessories" on
		if len(msg.Reactions) > 0 { // Reactions need one row
			messageRows += 1
			accessoryRow -= 1
		}
		if msg.File != nil { // Files need one row
			messageRows += 1
			accessoryRow -= 1
		}

		// Render the sender and the message
		for rowDelta, messageRow := range partitionIntoRows(
			makePrintWorthy(msg.Text), // Our message to render to the screen
			messageColumnWidth,        // The width of each message row
		) {
			if rowDelta == 0 {
				// Draw the sender on the first row of a message
				term.WriteText(0, row-messageRows+1, timestamp)
				term.WriteTextStyle(len(timestamp)+1, row-messageRows+1, senderStyle, sender)

				// Render reactions and file attachment after message
				if msg.File != nil {
					accessoryRow += 1
					renderFile(
						term,
						msg.File,                      // File to render
						index == selectedMessageIndex, // Is the current message selected?
						accessoryRow,                  // Which row should the file be rendered on?
						prefixWidth,                   // How far from the left should the first reaction be offset?
					)
				}
				if len(msg.Reactions) > 0 {
					accessoryRow += 1
					renderReactions(
						term,
						msg.Reactions, // Reactions to render
						accessoryRow,  // Which row should the reactions be rendered on?
						prefixWidth,   // How far from the left should the first reaction be offset?
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

	// Return how many messages were rendered to the screen
	return (len(messages) - 1) - index
}
