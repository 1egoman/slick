package frontend_test

import (
	"testing"

	"github.com/kyokomi/emoji" // convert :smile: to unicode
	"github.com/kylelemons/godebug/pretty"

	"github.com/1egoman/slick/frontend"
	"github.com/1egoman/slick/gateway"
)

func TestParseSlackMessageContainingEverything(t *testing.T) {
	var parsedMessage gateway.PrintableMessage

	message := "<#C024BE7LR|general> foo <@U024BE7LR|ryan> bar <!here|here> baz <!channel> &amp; &lt; quux &gt; :smile: <@U012345678>"

	err := frontend.ParseSlackMessage(
		message,
		&parsedMessage,
		func (id string) (*gateway.User, error) {
			return &gateway.User{Name: "user-looked-up-by-id"}, nil
		},
	)

	if err != nil {
		t.Errorf("Error parsing slack message: %s", err)
	}

	expected := []gateway.PrintableMessagePart{
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_CHANNEL, Content: "#general"},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " foo "},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_AT_MENTION_USER, Content: "@ryan"},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " bar "},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_AT_MENTION_GROUP, Content: "@here"},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: " baz "},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_AT_MENTION_GROUP, Content: "@channel"},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: emoji.Sprintf(" & < quux > :smile: ")},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_AT_MENTION_USER, Content: "@user-looked-up-by-id"},
	}

	if diff := pretty.Compare(parsedMessage.Parts(), expected); diff != "" {
		t.Errorf("Slack message was bad: \n%s", diff)
	}
}

func TestParseEmptySlackMessage(t *testing.T) {
	var parsedMessage gateway.PrintableMessage

	err := frontend.ParseSlackMessage(
		"",
		&parsedMessage,
		func (id string) (*gateway.User, error) {
			return &gateway.User{Name: "user-looked-up-by-id"}, nil
		},
	)

	if err != nil {
		t.Errorf("Error parsing slack message: %s", err)
	}

	expected := []gateway.PrintableMessagePart{}

	if diff := pretty.Compare(parsedMessage.Parts(), expected); diff != "" {
		t.Errorf("Slack message was bad: \n%s", diff)
	}
}

func TestParseTextOnlyMessage(t *testing.T) {
	var parsedMessage gateway.PrintableMessage

	message := "foo bar baz"

	err := frontend.ParseSlackMessage(
		message,
		&parsedMessage,
		func (id string) (*gateway.User, error) {
			return &gateway.User{Name: "user-looked-up-by-id"}, nil
		},
	)

	if err != nil {
		t.Errorf("Error parsing slack message: %s", err)
	}

	expected := []gateway.PrintableMessagePart{
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "foo bar baz"},
	}

	if diff := pretty.Compare(parsedMessage.Parts(), expected); diff != "" {
		t.Errorf("Slack message was bad: \n%s", diff)
	}
}

func TestParseNewlines(t *testing.T) {
	var parsedMessage gateway.PrintableMessage

	message := "foo\nbar\nbaz"

	err := frontend.ParseSlackMessage(
		message,
		&parsedMessage,
		func (id string) (*gateway.User, error) {
			return &gateway.User{Name: "user-looked-up-by-id"}, nil
		},
	)

	if err != nil {
		t.Errorf("Error parsing slack message: %s", err)
	}

	expected := []gateway.PrintableMessagePart{
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "foo"},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_NEWLINE},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "bar"},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_NEWLINE},
		gateway.PrintableMessagePart{Type: gateway.PRINTABLE_MESSAGE_PLAIN_TEXT, Content: "baz"},
	}

	if diff := pretty.Compare(parsedMessage.Parts(), expected); diff != "" {
		t.Errorf("Slack message was bad: \n%s", diff)
	}
}
