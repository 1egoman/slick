package gateway

import (
	"strings"
)

type PrintableMessagePartType int

const (
	PRINTABLE_MESSAGE_PLAIN_TEXT PrintableMessagePartType = iota
	PRINTABLE_MESSAGE_FORMATTING_BOLD
	PRINTABLE_MESSAGE_FORMATTING_ITALIC
	PRINTABLE_MESSAGE_AT_MENTION_USER  // (like @foo, @bar, etc)
	PRINTABLE_MESSAGE_AT_MENTION_GROUP // (like @channel, @here, etc)
	PRINTABLE_MESSAGE_CHANNEL          // (like #general)
	PRINTABLE_MESSAGE_CONNECTION       // (like "my custom slack team")
	PRINTABLE_MESSAGE_LINK             // (like http://example.com)
)

type PrintableMessagePart struct {
	Type PrintableMessagePartType
	Content string
	Metadata map[string]interface{}
}

type PrintableMessage struct {
	parts []PrintableMessagePart

	// Running `Lines(width)` takes a while, and it's pure. Memoize it for speed.
	linesCacheWidth int
	linesCache [][]PrintableMessagePart
}

func NewPrintableMessage(parts []PrintableMessagePart) PrintableMessage {
	return PrintableMessage{parts: parts}
}

func (p *PrintableMessage) Parts() []PrintableMessagePart {
	return p.parts
}

func (p *PrintableMessage) SetParts(parts []PrintableMessagePart) {
	p.parts = parts
}

// Return the length of the printable message
func (p *PrintableMessage) Length() int {
	length := 0
	for _, part := range p.parts {
		length += len(part.Content)
	}
	return length
}

// Return the printable message converted to plain text string.
func (p *PrintableMessage) Plain() string {
	total := ""
	for _, part := range p.parts {
		total += part.Content
	}
	return total
}

func (p *PrintableMessage) Lines(width int) [][]PrintableMessagePart {
	// If the result has already been calculated, use it.
	if p.linesCacheWidth == width && p.linesCache != nil {
		return p.linesCache
	}

	var lines [][]PrintableMessagePart
	var messageParts []PrintableMessagePart
	var extraWords []string

	for _, part := range p.parts {
		messageParts = append(messageParts, part)
		// log.Printf("MESSAGE PARTS %+v", messageParts)

		pm := PrintableMessage{parts: messageParts}
		if pm.Length() > width {
			pm = PrintableMessage{parts: messageParts[:len(messageParts) - 1]}
			maximumLengthOfLastMessagePart := width - pm.Length()

			// log.Printf("WANTED TO DRAW %+v BUT INSTEAD ONLY DID %d", messageParts, maximumLengthOfLastMessagePart)

			// Now, cut down all words that don't fit on this line.
			// foo bar baz hello world test quux (extraWords = [])
			// foo bar baz hello world test      (extraWords = [quux])
			// foo bar baz hello world           (extraWords = [test, quux])
			// foo bar baz hello                 (extraWords = [world, test, quux])
			//                    ^
			//                 "window width"
			// (done, since the line length < window width)

			wordsInLastMessagePart := strings.Split(part.Content, " ")
			for len(strings.Join(wordsInLastMessagePart, " ")) > maximumLengthOfLastMessagePart {
				if len(wordsInLastMessagePart) < 1 { // Make sure that there are words to split on. Fixes #9.
					break
				}

				// Remove one word from the last message part
				extraWords = append([]string{wordsInLastMessagePart[len(wordsInLastMessagePart) - 1]}, extraWords...)
				wordsInLastMessagePart = wordsInLastMessagePart[:len(wordsInLastMessagePart)-1]
			}

			// If the last message part in a line has content to append, then add it. However, if
			// it's empty (probably because all the words were moved to the next line) then delete
			// it (since its content would just be "" anyway)
			if len(wordsInLastMessagePart) > 0 {
				messageParts[len(messageParts)-1].Content = strings.Join(wordsInLastMessagePart, " ")
			} else {
				messageParts = messageParts[:len(messageParts)-1]
			}

			// log.Printf("ACTUALLY DRAWING %+v", messageParts)

			// Now that our line is below the max width, it can be drawn, so add it to the array.
			if len(messageParts) > 0 {
				lines = append(lines, messageParts)
			}
			// log.Printf("LINES %+v", lines)

			// Then, clear out the message parts that have been used so far.
			messageParts = make([]PrintableMessagePart, 0)

			// And start off the next line with any words that were removed from the current line to
			// make it fit.
			if len(extraWords) > 0 {
				messageParts = append(messageParts, PrintableMessagePart{
					Type: part.Type,
					Content: strings.Join(extraWords, " "),
				})
				extraWords = make([]string, 0)
			}
			// log.Printf("LINE BIT CUT OFF %+v", messageParts)
		}
	}

	// log.Printf("DONE LOOPING %+v", messageParts)

	// If there are any message parts left over, then add them at the end.
	if len(messageParts) > 0 {
		lines = append(lines, messageParts)
	}

	// log.Printf("RETURNING %+v", lines)
	p.linesCache = lines
	p.linesCacheWidth = width
	return lines
}
