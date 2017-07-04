package frontend_test

import (
	// "fmt"
	"testing"
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
}
