package checker

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
)

// Pattern offsets are UTF-16 units on the native side, not Go string bytes.
func TestRegexPatternFactsUnicodeAndNewlineFraming(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ question, want string }{
		{"regex-pattern-facts\nu\n😀/", "1|2|0|2|2|3|0|0"},
		{"regex-pattern-facts\n\n[a/]", "1|2|1|2|2|3|0|1|0|1|4"},
		{"regex-pattern-facts\n\n\\\n", "1|1|0|2|1|0|1|2|0"},
	} {
		out := &fields{}
		parts := strings.SplitN(row.question, "\n", 3)
		var encoded []string
		for _, unit := range utf16.Encode([]rune(parts[2])) {
			encoded = append(encoded, strconv.FormatUint(uint64(unit), 10))
		}
		question := parts[0] + "\n" + parts[1] + "\n" + strings.Join(encoded, ",")
		if err := (&Program{}).regexPatternFacts(out, question); err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(decodedFields(t, out.String()), "|"); got != row.want {
			t.Fatalf("%q: got %s, want %s", row.question, got, row.want)
		}
	}
	for _, question := range []string{"regex-pattern-facts", "regex-pattern-facts\nu", "regex-pattern-facts\nu\n65536", "regex-pattern-facts\nu\n1,,2", "regex-pattern-facts\nu\n01"} {
		if err := (&Program{}).regexPatternFacts(&fields{}, question); err == nil {
			t.Fatalf("accepted malformed %q", question)
		}
	}
}
