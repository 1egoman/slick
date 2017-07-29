package frontend_test

import (
	"testing"
	"github.com/1egoman/slick/gateway"
	"github.com/1egoman/slick/frontend"
)

func TestCommandbarEmptyCommand(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawCommandBar(
		"",
		0,
		&gateway.Channel{Name: "bar", IsMember: true, IsArchived: false},
		"foo",
		false,
		map[string]string{},
	)

	result, ok := screen.Compare("./tests/draw_commandbar_test/commandbar_empty_command.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestCommandbarArchivedNotAMember(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawCommandBar(
		"",
		0,
		&gateway.Channel{Name: "bar", IsMember: false, IsArchived: true},
		"foo",
		false,
		map[string]string{},
	)

	result, ok := screen.Compare("./tests/draw_commandbar_test/commandbar_archived_not_a_member.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestCommandbarOffline(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawCommandBar(
		"",
		0,
		&gateway.Channel{Name: "bar", IsMember: true, IsArchived: false},
		"foo",
		true,
		map[string]string{},
	)

	result, ok := screen.Compare("./tests/draw_commandbar_test/commandbar_offline.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestCommandbarOfflineNotAMember(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawCommandBar(
		"",
		0,
		&gateway.Channel{Name: "bar", IsMember: false, IsArchived: true},
		"foo",
		true,
		map[string]string{},
	)

	result, ok := screen.Compare("./tests/draw_commandbar_test/commandbar_offline_not_a_member.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestCommandbarLongCommand(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawCommandBar(
		"hello world - The quick brown fox jumped over the lazy dog",
		0,
		&gateway.Channel{Name: "bar", IsMember: true, IsArchived: false},
		"foo",
		false,
		map[string]string{},
	)

	result, ok := screen.Compare("./tests/draw_commandbar_test/commandbar_long_command.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestCommandbarLongCommandWithNewlines(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawCommandBar(
		"hello world - The quick brown \nfox jumped over the lazy dog",
		0,
		&gateway.Channel{Name: "bar", IsMember: true, IsArchived: false},
		"foo",
		false,
		map[string]string{},
	)

	result, ok := screen.Compare("./tests/draw_commandbar_test/commandbar_long_command_with_newlines.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}
