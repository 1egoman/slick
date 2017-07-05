package frontend

import (
	"github.com/kyokomi/emoji" // convert :smile: to unicode
	"strings"
	// "fmt"

	"github.com/1egoman/slick/gateway" // The thing to interface with slack
)

func ParseMarkdown(text string) gateway.PrintableMessage {
	text = emoji.Sprintf(text) // Emojis
	text = strings.Replace(text, "&amp;", "&", -1)
	text = strings.Replace(text, "&gt;", ">", -1)
	text = strings.Replace(text, "&lt;", "<", -1)

	var parts []gateway.PrintableMessagePart

	// Iterate through each character in the message.
	// Look for bits of markdown like *this* or _this_.
	var tagType gateway.PrintableMessagePartType
	startIndex := 0 // Start at the beginning of the message text
	lastTagEndIndex := 0
	for index, char := range text {
		if char == '\n' {
			// Enforce newlines.

			// Add any text before the newline to the printable message slice.
			parts = append(parts, gateway.PrintableMessagePart{
				Type:    gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
				Content: text[startIndex:index],
			})
			startIndex = index + 1
			lastTagEndIndex = index+1

			// Add the newline.
			parts = append(parts, gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_NEWLINE})

		// BOLD
		} else if char == '*' && tagType == gateway.PRINTABLE_MESSAGE_FORMATTING_BOLD {
			// Add any text before the newline to the printable message slice.
			parts = append(parts, gateway.PrintableMessagePart{
				Type:    gateway.PRINTABLE_MESSAGE_FORMATTING_BOLD,
				Content: text[startIndex:index],
			})
			lastTagEndIndex = index+1
		} else if char == '*' {
			// Since we just discovered the boundary of the next bit of interest, then add the
			// previous plain text bit (before this tag) to the parts slice.
			if index - lastTagEndIndex > 0 {
				parts = append(parts, gateway.PrintableMessagePart{
					Type:    gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
					Content: text[lastTagEndIndex:index],
				})
			}

			tagType = gateway.PRINTABLE_MESSAGE_FORMATTING_BOLD
			startIndex = index+1

		// ITALICS
		} else if char == '_' && tagType == gateway.PRINTABLE_MESSAGE_FORMATTING_ITALIC {
			// Add any text before the newline to the printable message slice.
			parts = append(parts, gateway.PrintableMessagePart{
				Type:    gateway.PRINTABLE_MESSAGE_FORMATTING_ITALIC,
				Content: text[startIndex:index],
			})
			lastTagEndIndex = index+1
		} else if char == '_' {
			// Since we just discovered the boundary of the next bit of interest, then add the
			// previous plain text bit (before this tag) to the parts slice.
			if index - lastTagEndIndex > 0 {
				parts = append(parts, gateway.PrintableMessagePart{
					Type:    gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
					Content: text[lastTagEndIndex:index],
				})
			}

			tagType = gateway.PRINTABLE_MESSAGE_FORMATTING_ITALIC
			startIndex = index+1

		// PREFORMATTED
		} else if index < len(text) - 2 && char == '`' && text[index + 1] == '`' && text[index + 2] == '`' && tagType == gateway.PRINTABLE_MESSAGE_FORMATTING_PREFORMATTED {
			// Add any text before the newline to the printable message slice.
			parts = append(parts, gateway.PrintableMessagePart{
				Type:    gateway.PRINTABLE_MESSAGE_FORMATTING_PREFORMATTED,
				Content: text[startIndex:index],
			})
			lastTagEndIndex = index+3
		} else if index < len(text) - 2 && char == '`' && text[index + 1] == '`' && text[index + 2] == '`' {
			// Since we just discovered the boundary of the next bit of interest, then add the
			// previous plain text bit (before this tag) to the parts slice.
			if index - lastTagEndIndex > 0 {
				parts = append(parts, gateway.PrintableMessagePart{
					Type:    gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
					Content: text[lastTagEndIndex:index],
				})
			}

			tagType = gateway.PRINTABLE_MESSAGE_FORMATTING_PREFORMATTED
			startIndex = index+3

		// CODE
		} else if char == '`' && tagType == gateway.PRINTABLE_MESSAGE_FORMATTING_CODE {
			// Add any text before the newline to the printable message slice.
			parts = append(parts, gateway.PrintableMessagePart{
				Type:    gateway.PRINTABLE_MESSAGE_FORMATTING_CODE,
				Content: text[startIndex:index],
			})
			lastTagEndIndex = index+1
		} else if char == '`' && tagType != gateway.PRINTABLE_MESSAGE_FORMATTING_PREFORMATTED {
			// Since we just discovered the boundary of the next bit of interest, then add the
			// previous plain text bit (before this tag) to the parts slice.
			if index - lastTagEndIndex > 0 {
				parts = append(parts, gateway.PrintableMessagePart{
					Type:    gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
					Content: text[lastTagEndIndex:index],
				})
			}

			tagType = gateway.PRINTABLE_MESSAGE_FORMATTING_CODE
			startIndex = index+1
		}
	}

	// Add the final plain text part to the message.
	// bla bla #general foo bar
	//                  ^^^^^^ = This bit
	if lastTagEndIndex < len(text) {
		parts = append(parts, gateway.PrintableMessagePart{
			Type:    gateway.PRINTABLE_MESSAGE_PLAIN_TEXT,
			Content: text[lastTagEndIndex:],
		})
	}

	return gateway.NewPrintableMessage(parts)
}
