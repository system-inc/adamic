package markdown

import (
	"testing"
	"unicode/utf16"
)

// TestUnitsMatchEncoding checks firstUnit, lastUnit and utf16Length against the text encoded whole, as
// they were written before #vbjv3d6, on ASCII, astral characters at either end, and invalid UTF-8.
func TestUnitsMatchEncoding(t *testing.T) {
	t.Parallel()
	texts := []string{
		"", "a", "ab", "é", "日本", "😀", "a😀", "😀a", "😀😀", "\U0010FFFF", "�", "a�",
		"\xff", "a\xff", "\xffa", "\xe2\x82", "a\xe2\x82", "\xe2\x82\xac", "\xe2\xe2\x82\xac", "\xf0\x9f\x98",
		"\xf0\x9f\x98\x80\x80", "\x80\xf0\x9f\x98\x80",
	}
	for _, text := range texts {
		units := utf16.Encode([]rune(text))
		first, last := rune(-1), rune(-1)
		if len(units) > 0 {
			first, last = rune(units[0]), rune(units[len(units)-1])
		}
		if actual := firstUnit(text); actual != first {
			t.Errorf("firstUnit(%q) = %U, encoding says %U", text, actual, first)
		}
		if actual := lastUnit(text); actual != last {
			t.Errorf("lastUnit(%q) = %U, encoding says %U", text, actual, last)
		}
		if actual := utf16Length(text); actual != len(units) {
			t.Errorf("utf16Length(%q) = %d, encoding says %d", text, actual, len(units))
		}
	}
}
