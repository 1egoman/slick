package frontend

import (
	"github.com/kyokomi/emoji" // convert :smile: to unicode
	"strings"
	"fmt"

	"github.com/1egoman/slick/gateway" // The thing to interface with slack
)

// Given a string to be displayed in the ui, tokenize the message and return a *PrintableMessage
// that contains each part as a token.
func ParseSlackMessage(text string, printableMessage *gateway.PrintableMessage, UserById func(string) (*gateway.User, error)) error {
	text = emoji.Sprintf(text) // Emojis
	text = strings.Replace(text, "&amp;", "&", -1)
	text = strings.Replace(text, "&gt;", ">", -1)
	text = strings.Replace(text, "&lt;", "<", -1)

	var parts []gateway.PrintableMessagePart

	// Iterate through each character in the message.
	// Look for tags that look like <%XXXXXXXXX>, where % is a number of symbols and X is [A-Z0-9]
	// If one is found, then turn it into a name and replace it.
	var tagType gateway.PrintableMessagePartType
	startIndex := 0 // Start at the beginning of the message text
	startContentIndex := -1
	for index, char := range text {
		var nextChar rune

		if index+1 < len(text)-1 {
			nextChar = rune(text[index+1])
		} else {
			nextChar = ' ' // Placeholder character.
		}

		if char == '\n' {
			// Enforce newlines.

			// Add any text before the newline to the printable message slice.
			parts = append(parts, gateway.PrintableMessagePart{
				Type:    gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
				Content: text[startIndex:index],
			})
			startIndex = index + 1

			// Add the newline.
			parts = append(parts, gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_NEWLINE})
		} else if char == '<' {
			// Since we just discovered the boundary of the next bit of interest, then add the
			// previous plain text bit (before this tag) to the parts slice.
			if index > 0 {
				parts = append(parts, gateway.PrintableMessagePart{
					Type:    gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
					Content: text[startIndex:index],
				})
			}

			startIndex = index
			startContentIndex = index + 2
			if nextChar == '@' { // ie, <@U5FR33U4R> for @foo
				tagType = gateway.PRINTABLE_MESSAGE_AT_MENTION_USER
			} else if nextChar == '!' { // ie, <!UDOU3ENS> for @channel
				tagType = gateway.PRINTABLE_MESSAGE_AT_MENTION_GROUP
			} else if nextChar == '#' { // ie, <#3IDU62ER> for #channel
				tagType = gateway.PRINTABLE_MESSAGE_CHANNEL
			} else {
				tagType = gateway.PRINTABLE_MESSAGE_LINK
				startContentIndex -= 1 // Links don't have a "idenfiying" character, so one less char is needed
			}
		} else if char == '>' && startContentIndex >= 0 {
			content := text[startContentIndex:index]
			metadata := make(map[string]interface{})

			// log.Printf("CONTENT", content, tagType)

			if tagType == gateway.PRINTABLE_MESSAGE_AT_MENTION_USER {
				contentParts := strings.Split(content, "|")
				if len(contentParts) == 1 { // content = ABCDEFGHI
					user, err := UserById(content)
					if err != nil {
						// Couldn't fetch user info, instead of exploding just render the user id.
						// FIXME: better way to do this? A bit of a compromise.
						content = fmt.Sprintf("@<%s>", content)
					} else {
						content = "@" + user.Name
					}
				} else if len(contentParts) == 2 { // content = ABCDEFJHI|username
					content = "@" + contentParts[1]
				}
			} else if tagType == gateway.PRINTABLE_MESSAGE_AT_MENTION_GROUP { // content = here / channel / everyone (a group name)
				content = "@" + content
			} else if tagType == gateway.PRINTABLE_MESSAGE_CHANNEL {
				contentParts := strings.Split(content, "|")
				if len(contentParts) == 1 { // content = general
					content = "#" + content
				} else if len(contentParts) == 2 { // content = ABCDEFJHI|general
					content = "#" + contentParts[1]
				}
			} else if tagType == gateway.PRINTABLE_MESSAGE_LINK {
				// Links have meta
				contentParts := strings.Split(content, "|")
				metadata["Href"] = contentParts[0]
				if len(contentParts) == 2 { // content = http://example.com|label
					content = contentParts[1]
				}
			}

			parts = append(parts, gateway.PrintableMessagePart{
				Type:     tagType,
				Content:  content,
				Metadata: metadata,
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
		parts = append(parts, gateway.PrintableMessagePart{
			Type:    gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
			Content: text[startIndex:],
		})
	}

	printableMessage.SetParts(parts)
	return nil
}
