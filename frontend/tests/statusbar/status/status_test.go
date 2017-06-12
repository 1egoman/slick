package frontend_test

import (
	"testing"
	"github.com/1egoman/slick/status"
	"github.com/1egoman/slick/gateway"
	"github.com/1egoman/slick/frontend"
)

func TestScreensMatch(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)
	str := status.Status{
		Type: status.STATUS_LOG,
		Message: "test",
		Show: true,
	}

	term.DrawStatusBar("chat", []gateway.Connection{}, nil, str, map[string]string{})

	result, ok := screen.Compare("./actual.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}
