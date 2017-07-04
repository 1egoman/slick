package frontend_test

import (
	// "fmt"
	"testing"
	"github.com/kyokomi/emoji"
	"reflect"
	. "github.com/1egoman/slick/frontend"
	"github.com/1egoman/slick/gateway"
)


func TestParseMarkdown(t *testing.T) {
	markdownOne := ParseMarkdown("hello *world*")
	if !reflect.DeepEqual(
		markdownOne,
		gateway.NewPrintableMessage(
			[]gateway.PrintableMessagePart{
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "hello "},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_FORMATTING_BOLD, Content: "world"},
			},
		),
	) {
		t.Errorf("The markdown 'hello *world*' wasn't parsed properly: %+v", markdownOne)
	}

	markdownTwo := ParseMarkdown("hello *world* with stuff at end")
	if !reflect.DeepEqual(
		markdownTwo,
		gateway.NewPrintableMessage(
			[]gateway.PrintableMessagePart{
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "hello "},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_FORMATTING_BOLD, Content: "world"},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " with stuff at end"},
			},
		),
	) {
		t.Errorf("The markdown 'hello *world* with stuff at end' wasn't parsed properly: %+v", markdownTwo)
	}

	markdownThree := ParseMarkdown("hello *world* (you _forgot_ about me!)")
	if !reflect.DeepEqual(
		markdownThree,
		gateway.NewPrintableMessage(
			[]gateway.PrintableMessagePart{
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "hello "},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_FORMATTING_BOLD, Content: "world"},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " (you "},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_FORMATTING_ITALIC, Content: "forgot"},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " about me!)"},
			},
		),
	) {
		t.Errorf("The markdown 'hello *world* (you _forgot_ about me!)' wasn't parsed properly: %+v", markdownThree)
	}

	markdownFour := ParseMarkdown("I _like_ to *write* `code` too")
	if !reflect.DeepEqual(
		markdownFour,
		gateway.NewPrintableMessage(
			[]gateway.PrintableMessagePart{
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "I "},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_FORMATTING_ITALIC, Content: "like"},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " to "},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_FORMATTING_BOLD, Content: "write"},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " "},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_FORMATTING_CODE, Content: "code"},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " too"},
			},
		),
	) {
		t.Errorf("The markdown 'I _like_ to *write* `code` too' wasn't parsed properly: %+v", markdownFour)
	}

	markdownFive := ParseMarkdown("Ending in a single trailing *whitespace* ")
	if !reflect.DeepEqual(
		markdownFive,
		gateway.NewPrintableMessage(
			[]gateway.PrintableMessagePart{
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "Ending in a single trailing "},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_FORMATTING_BOLD, Content: "whitespace"},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " "},
			},
		),
	) {
		t.Errorf("The markdown 'I _like_ to *write* `code` too' wasn't parsed properly: %+v", markdownFive)
	}

	markdownSix := ParseMarkdown("With preformatting: ```this is preformatted```")
	if !reflect.DeepEqual(
		markdownSix,
		gateway.NewPrintableMessage(
			[]gateway.PrintableMessagePart{
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "With preformatting: "},
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_FORMATTING_PREFORMATTED, Content: "this is preformatted"},
			},
		),
	) {
		t.Errorf("The markdown 'With preformatting: ```this is preformatted```' wasn't parsed properly: %+v", markdownSix)
	}

	markdownSeven := ParseMarkdown("I :heart: emoji!")
	if !reflect.DeepEqual(
		markdownSeven,
		gateway.NewPrintableMessage(
			[]gateway.PrintableMessagePart{
				gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: emoji.Sprintf("I :heart: emoji!")},
			},
		),
	) {
		t.Errorf("The markdown 'I :heart: emoji!' wasn't parsed properly: %+v", markdownSeven)
	}
}
