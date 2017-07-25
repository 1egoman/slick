package frontend_test

import (
	"testing"
	"github.com/1egoman/slick/frontend"
)

func TestFuzzyPickerRenderItems(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawFuzzyPicker(
		[]string{ // All items
			"foo",
			"bar",
			"baz",
			"quux",
			"hello",
			"world",
			"the",
			"quick",
			"brown",
			"fox",
			"jumped",
			"over",
			"the",
			"lazy",
			"dog",
		},
		0, // Selected index
		0, // Bottom displayed item, used to control scrolling
		func(a string) int { return 0 },
		map[string]string{},
	)

	result, ok := screen.Compare("./tests/draw_fuzzy_picker_test/fuzzy_picker_render_items.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestFuzzyPickerRenderItemsSelectedIndex4(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawFuzzyPicker(
		[]string{ // All items
			"foo",
			"bar",
			"baz",
			"quux",
			"hello",
			"world",
			"the",
			"quick",
			"brown",
			"fox",
			"jumped",
			"over",
			"the",
			"lazy",
			"dog",
		},
		4, // Selected index
		0, // Bottom displayed item, used to control scrolling
		func(a string) int { return 0 },
		map[string]string{},
	)

	result, ok := screen.Compare("./tests/draw_fuzzy_picker_test/fuzzy_picker_render_items_selected_index_4.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}

func TestFuzzyPickerRenderItemsSelectedIndex4BottomDisplayedItem2(t *testing.T) {
	screen := frontend.NewAsciiScreen()
	term := frontend.NewTerminalDisplay(screen)

	term.DrawFuzzyPicker(
		[]string{ // All items
			"foo",
			"bar",
			"baz",
			"quux",
			"hello",
			"world",
			"the",
			"quick",
			"brown",
			"fox",
			"jumped",
			"over",
			"the",
			"lazy",
			"dog",
		},
		4, // Selected index
		2, // Bottom displayed item, used to control scrolling
		func(a string) int { return 0 },
		map[string]string{},
	)

	result, ok := screen.Compare("./tests/draw_fuzzy_picker_test/fuzzy_picker_render_items_selected_index_4_bottom_displayed_item_2.txt")
	if !ok {
		t.Errorf("Error:\n%s", result)
	}
}
