package gateway

import (
	"log"
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
	PRINTABLE_MESSAGE_NEWLINE
)

type PrintableMessagePart struct {
	Type     PrintableMessagePartType
	Content  string
	Metadata map[string]interface{}
}

type PrintableMessage struct {
	parts []PrintableMessagePart

	// Running `Lines(width)` takes a while, and it's pure. Memoize it for speed.
	linesCacheWidth int
	linesCache      [][]PrintableMessagePart
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
	var lines [][]PrintableMessagePart
	
	var lineBeingAssembled []PrintableMessagePart
	lineWidth := 0
	for _, part := range p.parts {
		if lineWidth + len(part.Content) > width {
			widthRemainingInLine := width - lineWidth

			if part.Type == PRINTABLE_MESSAGE_NEWLINE {
				lines = append(lines, lineBeingAssembled)
				lineBeingAssembled = make([]PrintableMessagePart, 0)
			}
			
			content := part.Content
			for len(content) > width {
				// The goal is to split the part in two - the first bit goes on the current line, the second bit is saved for the next iteration.
				
				// Attempt to split the part at a space, if possible. If not, just split in the middle of a word.
				// (Look for the last space in the "first bit" of the string, and split at that marker)
				amountOfLineUsed := strings.LastIndex(content[:widthRemainingInLine], " ")+1
				if amountOfLineUsed <= 0 {
					amountOfLineUsed = widthRemainingInLine
				}
				
				// Append the first bit to the current line
				firstBit := PrintableMessagePart{Type: part.Type, Content: content[:amountOfLineUsed], Metadata: part.Metadata}
				lineBeingAssembled = append(lineBeingAssembled, firstBit)
			
				// Append the current line to the lines collection.
				lines = append(lines, lineBeingAssembled)
				lineBeingAssembled = make([]PrintableMessagePart, 0)
				
				// Remove the chunk already used from the message content.
				content = content[amountOfLineUsed:]
				
				// Parts after the first part should be full width - ie:
				//          |         | <= Partial width
				// #general The quick b
				// |                  | <= Full width
				// rown fox jumps over 
				// the lazy dog
				widthRemainingInLine = width
			}
			
			// Finally, append the final bit that was left unhandled.
			finalBit := PrintableMessagePart{Type: part.Type, Content: content, Metadata: part.Metadata}
			lineBeingAssembled = append(lineBeingAssembled, finalBit)
		} else {
			// Add the part to the end of the line if it fits on the line.
			lineWidth += len(part.Content)
			lineBeingAssembled = append(lineBeingAssembled, part)
		}
	}
	
	// Append final line to the collection.
	lines = append(lines, lineBeingAssembled)
	
	return lines
}

func (p *PrintableMessage) Print(width int) {
	for _, line := range p.Lines(width) {
		lineContent := ""
		for _, part := range line {
			lineContent += part.Content
		}
		fmt.Println(lineContent)
	}
}
