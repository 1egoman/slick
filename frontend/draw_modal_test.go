package frontend_test

import (
	"testing"
	"github.com/1egoman/slick/frontend"
)

func TestRenderEmptyModal(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawModal("title", "", 0, false)
	// term.DrawModal("title", "body text\ngoes here\nfoo bar baz")

	result, ok := screen.Compare("./tests/draw_modal_test/modal_empty.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestRenderModalWithContent(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawModal("title", "body text\ngoes here\nfoo bar baz", 0, false)

	result, ok := screen.Compare("./tests/draw_modal_test/modal_with_content.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestRenderModalWithContentScroll(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawModal("title", "body text\ngoes here\nfoo bar baz", 1, false)

	result, ok := screen.Compare("./tests/draw_modal_test/modal_with_content_scroll.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}
