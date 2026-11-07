package native

import (
	"strings"
	"testing"
)

// A literal's UTF-16 length is emitted with it (adamic.h's ADAMIC_STRING_UNITS), so the runtime never
// rescans or caches into an immortal literal. These are JavaScript's lengths for the same text.
func TestStringUnitsAreJavaScriptLengths(test *testing.T) {
	test.Parallel()
	for _, row := range []struct {
		text  string
		units int
	}{
		{"", 0},
		{"ascii", 5},
		{"é", 1},
		{"€", 1},
		{"😀", 2},
		{"a😀é€", 5},
		{"\xed\xa0\x80", 1}, // a lone surrogate, as the IR holds it
	} {
		if got := utf16Units(row.text); got != row.units {
			test.Errorf("utf16Units(%q) = %d, want %d", row.text, got, row.units)
		}
	}
	long := strings.Repeat("ab😀", 9000)
	if got := utf16Units(long); got != 9000*4 {
		test.Fatalf("long literal: %d units, want %d", got, 9000*4)
	}
}
