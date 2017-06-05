package gateway_test

import (
	// "fmt"
	. "github.com/1egoman/slick/gateway"
	"reflect"
	"testing"
)

func plainText(text string) PrintableMessagePart {
	return PrintableMessagePart{Type: PRINTABLE_MESSAGE_PLAIN_TEXT, Content: text}
}
func channel(text string) PrintableMessagePart {
	return PrintableMessagePart{Type: PRINTABLE_MESSAGE_AT_MENTION_GROUP, Content: text}
}

func TestPrintableMessageLines(t *testing.T) {
	for _, test := range []struct {
		MessageParts  []PrintableMessagePart
		Width         int
		WrappedResult [][]PrintableMessagePart
	}{
		// The simple case: a one line message should come back unmodified.
		{
			MessageParts: []PrintableMessagePart{plainText("hello world"), plainText("foo")},
			Width:        15,
			WrappedResult: [][]PrintableMessagePart{
				[]PrintableMessagePart{plainText("hello world"), plainText("foo")},
			},
		},
		// Evenly wrap a message at it's part boundary
		{
			MessageParts: []PrintableMessagePart{plainText("hello world"), plainText("foo")},
			Width:        len("hello world"),
			WrappedResult: [][]PrintableMessagePart{
				[]PrintableMessagePart{plainText("hello world")},
				[]PrintableMessagePart{plainText("foo")},
			},
		},
		// Ensure that a `PrintableMessagePart` can be broken at a line boundary to wrap to the next
		// line.
		{
			MessageParts: []PrintableMessagePart{plainText("hello world"), plainText("foo bar")},
			Width:        15,
			WrappedResult: [][]PrintableMessagePart{
				[]PrintableMessagePart{plainText("hello world"), plainText("foo")},
				[]PrintableMessagePart{plainText("bar")},
			},
		},
		// Formatting should still work, ie, a channel should stay a channel.
		{
			MessageParts: []PrintableMessagePart{channel("general"), plainText("test")},
			Width:        15,
			WrappedResult: [][]PrintableMessagePart{
				[]PrintableMessagePart{channel("general"), plainText("test")},
			},
		},
		// Wrapping at a part boundary of a non-plaintext part should make both "half parts"
		{
			MessageParts: []PrintableMessagePart{plainText("foo bar baz"), channel("quux hello world")},
			Width:        len("foo bar baz quux"),
			WrappedResult: [][]PrintableMessagePart{
				[]PrintableMessagePart{plainText("foo bar baz"), channel("quux")},
				[]PrintableMessagePart{channel("hello world")},
			},
		},

		// Test for issue #9
		{
			MessageParts: []PrintableMessagePart{plainText("foo")},
			Width:        -1, // This shouldn't ever happen, I just wanted to get `maximumLengthOfLastMessagePart` < 0
			WrappedResult: [][]PrintableMessagePart{
				[]PrintableMessagePart{plainText("foo")},
			},
		},
	} {
		pm := NewPrintableMessage(test.MessageParts)
		lines := pm.Lines(test.Width)

		if !reflect.DeepEqual(lines, test.WrappedResult) {
			t.Errorf("%+v != %+v\n", lines, test.WrappedResult)
		}
	}
}
