package color_test

import (
	. "github.com/1egoman/slick/color"
	"testing"
)

func assertSerialized(t *testing.T, a string, b string) {
	if a != b {
		t.Errorf("%s != %s", a, b)
	}
}

func TestSerializingStyles(t *testing.T) {
	assertSerialized(t, SerializeStyle("red", "green", STYLE_PLAIN), "red:green:")
	assertSerialized(t, SerializeStyle("red", "green", STYLE_BOLD), "red:green:B")
	assertSerialized(t, SerializeStyle("red", "green", STYLE_UNDERLINE), "red:green:U")
	assertSerialized(t, SerializeStyle("red", "green", STYLE_BOLD|STYLE_UNDERLINE), "red:green:BU")

	assertSerialized(t, SerializeStyle("teal", "", STYLE_BOLD), "teal::B")
	assertSerialized(t, SerializeStyle("", "#FF0000", STYLE_PLAIN), ":#FF0000:")
	assertSerialized(t, SerializeStyle("", "", STYLE_PLAIN), "::")
}

func assertDeSerialized(t *testing.T, serializedStyle, realFg, realBg string, realStyle StyleFormattingMask) {
	fg, bg, style, err := DeSerializeStyle(serializedStyle)

	if err != nil {
		t.Errorf("Deserializing style %s failed with error: %s", serializedStyle, err)
	} else if fg != realFg {
		t.Errorf("Deserialized foreground color %s != actual foreground color %s", realFg, fg)
	} else if bg != realBg {
		t.Errorf("Deserialized backgrond color %s != actual backgrond color %s", realBg, bg)
	} else if fg != realFg {
		t.Errorf("Deserialized style %+v != style color %+v", realStyle, style)
	}
}

func TestDeserializingStyles(t *testing.T) {
	assertDeSerialized(t, "red:green:", "red", "green", STYLE_PLAIN)
	assertDeSerialized(t, "red:green:B", "red", "green", STYLE_BOLD)
	assertDeSerialized(t, "red:green:U", "red", "green", STYLE_UNDERLINE)
	assertDeSerialized(t, "red:green:BU", "red", "green", STYLE_BOLD|STYLE_UNDERLINE)

	assertDeSerialized(t, "teal::B", "teal", "", STYLE_BOLD)
	assertDeSerialized(t, ":#FF0000:", "", "#FF0000", STYLE_PLAIN)
	assertDeSerialized(t, "::", "", "", STYLE_PLAIN)

	// Verify that deserializing a style with invalid seperators (anything but two) produces an error.
	_, _, _, err := DeSerializeStyle(":")
	if err.Error() != "Less than or greater than three colon-seperated parts in style formatting string." {
		t.Errorf("Single colon seperated part in color doesn't produce the right error: %s", err)
	}
	_, _, _, err = DeSerializeStyle(":::")
	if err.Error() != "Less than or greater than three colon-seperated parts in style formatting string." {
		t.Errorf("Triple colon seperated part in color doesn't produce the right error: %s", err)
	}
}
