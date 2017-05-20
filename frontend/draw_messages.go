package frontend

import (
	"fmt"
	"strings"
	"time"
	"log"

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

func getAttachmentHeight(attachment gateway.Attachment) int {
	lines := 0

	// One line for the title
	lines += 1

	// One line for each field
	lines += len(attachment.Fields)

	return lines
}

// Given a pointer to a file and a row to render it on, render it.
func renderFile(term *TerminalDisplay, file *gateway.File, isSelected bool, row int, leftOffset int) {
	if file != nil {
		var messageActions []string
		if isSelected {
			messageActions = []string{"Open", "Copy"}
		}

		fileRow := fmt.Sprintf(
			"| %s: %s",
			file.Name,
			formatAbreviatedLink(file.Permalink),
		)
		term.WriteTextStyle(leftOffset, row, term.Styles["MessageFile"], fileRow)

		// Render actions after the file
		messageActionOffset := leftOffset + len(fileRow) + 1 // Add a space netween file and actions
		renderActions(term, messageActions, messageActionOffset, row)
	}
}

func renderAttachment(
	term *TerminalDisplay,
	attachment gateway.Attachment,
	isSelected bool,
	row int,
	leftOffset int,
	windowWidth int,
	index int,
) {
	selectedActions := []string{"Link"}
	selectedActionsWidth := len(strings.Join(selectedActions, " "))

	maxAttachmentWidth := windowWidth - leftOffset - 1
	if isSelected {
		maxAttachmentWidth -= selectedActionsWidth + 2
	}

	attachmentColor := tcell.StyleDefault.
		Foreground(tcell.GetColor("#" + attachment.Color)).
		Bold(true)

	title := attachment.Title
	if len(title) > maxAttachmentWidth {
		title = title[:maxAttachmentWidth]
	}

	term.WriteTextStyle(leftOffset, row, attachmentColor, "+ ")

	// Draw the attachment title.
	term.WriteTextStyle(
		leftOffset + 2,
		row,
		tcell.StyleDefault,
		title,
	)

	// Render actions after the attachment title
	if isSelected {
		actionsPositionOnEndOfRow := windowWidth - selectedActionsWidth - 1
		actionsPositionOnEndOfText := leftOffset + 2 + len(title) + 1
		var actionPosition int

		if actionsPositionOnEndOfRow > actionsPositionOnEndOfText {
			actionPosition = actionsPositionOnEndOfText
		} else {
			actionPosition = actionsPositionOnEndOfRow
		}

		renderActions(term, selectedActions, actionPosition, row)
	}

	// Draw each attachment field.
	for index, field := range attachment.Fields {
		term.WriteTextStyle(leftOffset, row + index + 1, attachmentColor, "|")
		term.WriteTextStyle(
			leftOffset + 2,
			row + index + 1,
			term.Styles["MessageAttachmentFieldTitle"],
			field.Title+":",
		)
		term.WriteTextStyle(
			leftOffset + 2 + len(field.Title) + 2,
			row + index + 1,
			term.Styles["MessageAttachmentFieldValue"],
			field.Value,
		)
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

// Draw message history in the channel
func (term *TerminalDisplay) DrawMessages(
	messages []gateway.Message, // A list of messages to render
	selectedMessageIndex int, // Index of selected message (-1 for no selected message)
	bottomDisplayedItem int, // The bottommost message. If 0, bottommost message is most recent.
	userById func(string) (*gateway.User, error),
) int { // Return how many messages were rendered.
	width, height := term.screen.Size()

	for r := 0; r < height-BottomPadding; r++ {
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
	row := height - BottomPadding - 1
	for row >= 0 && index >= 0 {
		msg := messages[index]

		// Get sender information
		sender, senderStyle := getSenderInfo(msg)

		timestamp := time.Unix(int64(msg.Timestamp), 0).Format("15:04:05")
		prefixWidth := len(timestamp) + 1 + len(sender) + 1

		// Is the message selected?
		var selectedStyle tcell.Style
		if index == selectedMessageIndex {
			selectedStyle = term.Styles["MessageSelected"]
		} else {
			selectedStyle = tcell.StyleDefault
		}

		parsedMessage, err := parseSlackMessage(msg.Text, userById)
		if err != nil {
			// FIXME: Probably should return an error here? And not return 0?
			log.Println("Error making message print-worthy (probably because fetching user id => user name failed):", err)
			return 0
		}

		// Calculate how many rows the message requires to render.
		messageColumnWidth := width - prefixWidth
		messageRows := (parsedMessage.Length() / messageColumnWidth) + 1
		accessoryRow := row         // The row to start rendering "message accessories" on
		if len(msg.Text) == 0 {
			accessoryRow -= 1
		}
		if len(msg.Reactions) > 0 { // Reactions need one row
			messageRows += 1
			accessoryRow -= 1
		}
		if msg.File != nil { // Files need one row
			messageRows += 1
			accessoryRow -= 1
		}
		if msg.Attachments != nil { // Attachments need a lot of rows. :(
			// Collect the total attachment height
			var attachmentSize int
			for _, attach := range *msg.Attachments {
				attachmentSize += getAttachmentHeight(attach)
			}

			messageRows += attachmentSize
			accessoryRow -= attachmentSize
		}

		// Draw the sender and timestamp on the first row of a message
		term.WriteTextStyle(0, row-messageRows+1, selectedStyle, timestamp)
		term.WriteTextStyle(len(timestamp)+1, row-messageRows+1, senderStyle, sender)

		// Render reactions and file attachment after message
		// log.Printf("attach %+v", msg.Attachments)
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

		if msg.Attachments != nil {
			for attachmentIndex, attachment := range *msg.Attachments {
				accessoryRow += 1
				renderAttachment(
					term,
					attachment,                    // Attachment to render
					index == selectedMessageIndex, // Is the current message selected?
					accessoryRow,                  // Which row to render the attachment on?
					prefixWidth,                   // How far to the left should the first attachment be offset?
					width,                         // Width of the window
					attachmentIndex,               // The index of the given attachent in the selected message
				)
				accessoryRow += getAttachmentHeight(attachment) - 1
			}
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


		// Render the sender and the message
		totalWidth := 0
		rowDelta := -1
		for _, part := range parsedMessage.Parts() {
			// If the content won't fit on this line, move to the next line (y++) and reset x to 0
			if (totalWidth + len(part.Content)) > messageColumnWidth {
				rowDelta += 1
				totalWidth = 0
			}

			var style tcell.Style
			if part.Type == PLAIN_TEXT {
				style = tcell.StyleDefault
			} else if part.Type == AT_MENTION_USER {
				style = tcell.StyleDefault.
					Foreground(tcell.ColorRed).
					Bold(true)
			} else if part.Type == AT_MENTION_GROUP {
				style = tcell.StyleDefault.
					Foreground(tcell.ColorYellow).
					Bold(true)
			} else if part.Type == CHANNEL {
				style = tcell.StyleDefault.
					Foreground(tcell.ColorBlue).
					Bold(true)
			}

			// Render the next message part
			term.WriteTextStyle(
				prefixWidth + totalWidth,
				row-messageRows-rowDelta,
				style,
				part.Content,
			)

			totalWidth += len(part.Content)
		}

		// Subtract the message's height.
		row -= messageRows
		index -= 1
	}

	// Return how many messages were rendered to the screen
	return (len(messages) - 1) - index
}
