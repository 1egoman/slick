package frontend_test

import (
	"testing"
	// "time"
	"github.com/1egoman/slick/status"
	"github.com/1egoman/slick/gateway"
	"github.com/1egoman/slick/frontend"
	"github.com/1egoman/slick/gateway/slack"
)

func TestStatusbarConnectionList(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)
	str := status.Status{Type: status.STATUS_LOG, Message: "", Show: false}

	term.DrawStatusBar("chat", []gateway.Connection{
		gatewaySlack.NewWithName("helloworld", "token"),
		gatewaySlack.NewWithName("example", "token"),
	}, nil, str, map[string]string{})

	result, ok := screen.Compare("./tests/draw_statusbar_test/statusbar_connection_list.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestStatusbarStatusMessage(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)
	str := status.Status{
		Type: status.STATUS_LOG,
		Message: "test",
		Show: true,
	}

	term.DrawStatusBar("chat", []gateway.Connection{}, nil, str, map[string]string{})

	result, ok := screen.Compare("./tests/draw_statusbar_test/statusbar_status_message.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestStatusbarMultilineStatus(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)
	str := status.Status{Type: status.STATUS_LOG, Message: "hello\nworld\nfoo\nbar", Show: true}

	term.DrawStatusBar("pick", []gateway.Connection{
		gatewaySlack.NewWithName("helloworld", "token"),
		gatewaySlack.NewWithName("example", "token"),
	}, nil, str, map[string]string{})

	result, ok := screen.Compare("./tests/draw_statusbar_test/statusbar_multiline_status.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestStatusbarTypingUsers(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)
	str := status.Status{Type: status.STATUS_LOG, Message: "", Show: false}

	activeConnection := gatewaySlack.NewWithName("helloworld", "token")
	activeConnection.TypingUsers().Add("foo", time.Now().Add(-1 * time.Second))
	activeConnection.TypingUsers().Add("bar", time.Now())

	term.DrawStatusBar("chat", []gateway.Connection{
		activeConnection,
		gatewaySlack.NewWithName("example", "token"),
	}, activeConnection, str, map[string]string{})

	result, ok := screen.Compare("./tests/draw_statusbar_test/statusbar_typing_users.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}
