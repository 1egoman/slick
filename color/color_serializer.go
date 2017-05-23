package color

import (
	"fmt"
	"strings"
	"errors"
	"github.com/gdamore/tcell"
)

type StyleFormattingMask int
const (
	STYLE_PLAIN StyleFormattingMask = 0
	STYLE_BOLD = 1 << iota
	STYLE_UNDERLINE
)

// Given a foreground color, background color, and formatting mask, return a serialized version of
// the style.
// SerializeStyle("red", "green", STYLE_BOLD) // => "red:green:B"
// SerializeStyle("red", "", STYLE_BOLD | STYLE_UNDERLINE) // => "red::B"
// SerializeStyle("red", "#FF00FF", STYLE_PLAIN) // => "red:#FF00FF:"
func SerializeStyle(foreground string, background string, formatting StyleFormattingMask) string {
	formattingString := ""
	if formatting & STYLE_BOLD > 0 {
		formattingString += "B"
	}
	if formatting & STYLE_UNDERLINE > 0 {
		formattingString += "U"
	}

	return fmt.Sprintf("%s:%s:%s", foreground, background, formattingString)
}

// Given a serialized style string, deserialize into a forgeground color, background color, and
// formatting mask. Also returns an error parsing if there was one.
// DeSerializeStyle("red:green:B") // => "red", "green", STYLE_BOLD, nil
// DeSerializeStyle(":") // => "", "", STYLE_PLAIN, error("Less than or greater than three colon-seperated parts in style formatting string.")
func DeSerializeStyle(styleString string) (string, string, StyleFormattingMask, error) {
	parts := strings.Split(styleString, ":")

	if len(parts) != 3 {
		return "", "", STYLE_PLAIN, errors.New("Less than or greater than three colon-seperated parts in style formatting string.")
	}

	// Create the formatting mask given the last part of the formatting string
	var formattingMask StyleFormattingMask
	isBold := strings.Index(parts[2], "B") >= 0
	if isBold {
		formattingMask = formattingMask | STYLE_BOLD
	}
	isUnderline := strings.Index(parts[2], "U") >= 0
	if isUnderline {
		formattingMask = formattingMask | STYLE_UNDERLINE
	}

	return parts[0], parts[1], formattingMask, nil
}

// Deserialize a serialized style string into a tcell.Style.
func DeSerializeStyleTcell(styleString string) (tcell.Style) {
	foreground, background, mask, err := DeSerializeStyle(styleString)

	// If we fail, use default style.
	if err != nil {
		return tcell.StyleDefault
	}

	style := tcell.StyleDefault

	if len(foreground) > 0 {
		style = style.Foreground(tcell.GetColor(foreground))
	}
	if len(background) > 0 {
		style = style.Background(tcell.GetColor(background))
	}

	return style.Bold(mask & STYLE_BOLD > 0).Underline(mask & STYLE_UNDERLINE > 0)
}
