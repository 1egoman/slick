package frontend

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/kyokomi/emoji" // convert :smile: to unicode

	"github.com/1egoman/slick/color"
	"github.com/1egoman/slick/gateway" // The thing to interface with slack
)

// Given an array of reactions and a row to render them on, render them.
func renderReactions(
	term *TerminalDisplay,
	config map[string]string,

	reactions []gateway.Reaction,
	row int,
	leftOffset int,
) {
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
			color.DeSerializeStyleTcell(config["Message.Reaction.Color"]),
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
func renderFile(
	term *TerminalDisplay,
	config map[string]string,

	file *gateway.File,
	isSelected bool,
	row int,
	leftOffset int,
) {
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
		term.WriteTextStyle(leftOffset, row, color.DeSerializeStyleTcell(config["Message.FileColor"]), fileRow)

		// Render actions after the file
		messageActionOffset := leftOffset + len(fileRow) + 1 // Add a space netween file and actions
		renderActions(term, config, messageActions, messageActionOffset, row)
	}
}

func renderAttachment(
	term *TerminalDisplay,
	config map[string]string,

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
		leftOffset+2,
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

		renderActions(term, config, selectedActions, actionPosition, row)
	}

	// Draw each attachment field.
	for index, field := range attachment.Fields {
		term.WriteTextStyle(leftOffset, row+index+1, attachmentColor, "|")
		term.WriteTextStyle(
			leftOffset+2,
			row+index+1,
			color.DeSerializeStyleTcell(config["Message.Attachment.FieldTitleColor"]),
			field.Title+":",
		)
		term.WriteTextStyle(
			leftOffset+2+len(field.Title)+2,
			row+index+1,
			color.DeSerializeStyleTcell(config["Message.Attachment.FieldValueColor"]),
			field.Value,
		)
	}
}

// Render actions that can be done with the message
// Uppercase letters in the actions are highlighted in a different color (they are the key to
// press to do the thing)
func renderActions(
	term *TerminalDisplay,
	config map[string]string,

	actions []string,
	leftOffset int,
	row int,
) {
	for _, action := range actions {
		for _, char := range action {
			if char >= 'A' && char <= 'Z' {
				term.WriteTextStyle(leftOffset, row, color.DeSerializeStyleTcell(config["Message.Action.HighlightColor"]), string(char))
			} else {
				term.WriteTextStyle(leftOffset, row, color.DeSerializeStyleTcell(config["Message.Action.Color"]), string(char))
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

func getRelativeLineNumber(activeLine int, currentLine int) int {
	value := currentLine - activeLine
	if value < 0 {
		return -1 * value
	} else {
		return value
	}
}

// Draw message history in the channel
func (term *TerminalDisplay) DrawMessages(
	messages []gateway.Message, // A list of messages to render
	selectedMessageIndex int, // Index of selected message (-1 for no selected message)
	bottomDisplayedItem int, // The bottommost message. If 0, bottommost message is most recent.
	userById func(string) (*gateway.User, error),
	userOnline func(user *gateway.User) bool,
	config map[string]string,
) (renderedMessageNumber int, renderedAllMessages bool) { // Return how many messages were rendered.
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

		// Calculate the width of the message prefix.
		timestamp := time.Unix(int64(msg.Timestamp), 0).Format(config["Message.TimestampFormat"])
		prefixWidth := len(timestamp) + 1 + len(sender) + 1
		var relativeLineWidth int
		if _, ok := config["Message.RelativeLine"]; ok {
			// The relative line gutter width should be the same length as the height. If we've got
			// over one hundred lines then we're going to have three digit relative line numbers.
			relativeLineWidth = len(fmt.Sprintf("%d", height)) + 1
			prefixWidth += relativeLineWidth
		}

		// Is the message selected?
		var selectedStyle tcell.Style
		if index == selectedMessageIndex {
			selectedStyle = color.DeSerializeStyleTcell(config["Message.SelectedColor"])
		} else {
			selectedStyle = tcell.StyleDefault
		}

		// Take our message text and convert it to message parts
		// TODO: cache this somehow. It's slow as hell.
		var parsedMessage gateway.PrintableMessage
		err := ParseSlackMessage(msg.Text, &parsedMessage, userById)
		if err != nil {
			// FIXME: Probably should return an error here? And not return 0?
			log.Println("Error making message print-worthy (probably because fetching user id => user name failed):", err)
			return 0, false
		}

		// Calculate how many rows the message requires to render.
		messageColumnWidth := width - prefixWidth
		parsedMessageLines := parsedMessage.Lines(messageColumnWidth)
		messageRows := len(parsedMessageLines)
		accessoryRow := row // The row to start rendering "message accessories" on
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

		// The number of characters from the left that messages should be offset by.
		messageOffset := 0

		// Draw a relative line number at the start of the line, if requested.
		var lineNumberStyle tcell.Style
		if _, ok := config["Message.RelativeLine"]; ok {
			relativeLineNumber := getRelativeLineNumber(selectedMessageIndex, index)

			// Color the active line number different than the rest
			if selectedMessageIndex == index {
				lineNumberStyle = color.DeSerializeStyleTcell(config["Message.LineNumber.ActiveColor"])
			} else {
				lineNumberStyle = color.DeSerializeStyleTcell(config["Message.LineNumber.Color"])
			}

			// Convert width to a string.
			w := fmt.Sprintf("%d", relativeLineWidth)

			// Fill in each cell that belongs to the message.
			// The top cel has the line number in it.
			for i := 0; i <= messageRows; i++ {
				var content string
				if i == 1 {
					content = fmt.Sprintf("%"+w+"d ", relativeLineNumber)
				} else {
					content = fmt.Sprintf("%"+w+"s ", "") // a string that is `w` characters long.
				}
				term.WriteTextStyle(
					messageOffset,
					row-messageRows+i,
					lineNumberStyle,
					content,
				)
			}


			messageOffset += relativeLineWidth + 1 // Each line number needs this many columns, +1 padding
		}

		// Draw the sender, the sender's online status, and the timestamp on the first row of a message
		term.WriteTextStyle(messageOffset, row-messageRows+1, selectedStyle, timestamp)
		messageOffset += len(timestamp) + 1

		if msg.Sender != nil && userOnline(msg.Sender) {
			// Render online status for sender
			term.WriteTextStyle(
				messageOffset,
				row-messageRows+1,
				color.DeSerializeStyleTcell(config["Message.Sender.OnlinePrefixColor"]),
				config["Message.Sender.OnlinePrefix"],
			)
			messageOffset += len(config["Message.Sender.OnlinePrefix"])
		} else if msg.Sender != nil {
			// Render offline status for sender
			term.WriteTextStyle(
				messageOffset,
				row-messageRows+1,
				color.DeSerializeStyleTcell(config["Message.Sender.OfflinePrefixColor"]),
				config["Message.Sender.OfflinePrefix"],
			)
			messageOffset += len(config["Message.Sender.OfflinePrefix"])
		}

		term.WriteTextStyle(messageOffset, row-messageRows+1, senderStyle, sender)
		messageOffset += len(sender)

		// Render optional reactions, file, or attachment after message
		if msg.File != nil {
			accessoryRow += 1
			renderFile(
				term, config,
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
					term, config,
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
				term, config,
				msg.Reactions, // Reactions to render
				accessoryRow,  // Which row should the reactions be rendered on?
				prefixWidth,   // How far from the left should the first reaction be offset?
			)
		}

		// Render the sender and the message
		for lineIndex, line := range parsedMessageLines {
			totalWidth := 0
			for _, part := range line {
				// How should this part be styled?
				var style tcell.Style
				if part.Type == gateway.PRINTABLE_MESSAGE_PLAIN_TEXT {
					style = tcell.StyleDefault
				} else if part.Type == gateway.PRINTABLE_MESSAGE_AT_MENTION_USER {
					style = color.DeSerializeStyleTcell(config["Message.Part.AtMentionUserColor"])
				} else if part.Type == gateway.PRINTABLE_MESSAGE_AT_MENTION_GROUP {
					style = color.DeSerializeStyleTcell(config["Message.Part.AtMentionGroupColor"])
				} else if part.Type == gateway.PRINTABLE_MESSAGE_CHANNEL {
					style = color.DeSerializeStyleTcell(config["Message.Part.ChannelColor"])
				} else if part.Type == gateway.PRINTABLE_MESSAGE_LINK {
					style = color.DeSerializeStyleTcell(config["Message.Part.LinkColor"])
				}

				// Render the next message part
				term.WriteTextStyle(
					messageOffset+1+totalWidth,
					row-messageRows+lineIndex+1,
					style,
					part.Content,
				)

				totalWidth += len(part.Content)
			}
		}

		// Subtract the message's height.
		row -= messageRows
		index -= 1
	}

	// Return how many messages were rendered to the screen
	return (len(messages) - 1) - index, (index >= 0)
}
