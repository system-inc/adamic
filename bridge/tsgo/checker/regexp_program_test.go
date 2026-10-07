package checker

import (
	"strconv"
	"testing"
)

func TestRegexpProgramFacts(t *testing.T) {
	for _, pattern := range []string{"", `\Afoo\z`, `(?i)k`, `(?m)^foo$`, `\p{Greek}+`, `(?:a?)*`, `a{2,4}`} {
		out := &fields{}
		out.number(1)
		out.text("regexp-program")
		wire, err := regexpProgramFacts(out, "regexp-program\n"+pattern)
		if err != nil {
			t.Fatal(err)
		}
		values := decodedFields(t, wire)
		if values[2] != "1" {
			t.Fatalf("valid pattern %q refused: %v", pattern, values)
		}
		start, err := strconv.Atoi(values[3])
		if err != nil {
			t.Fatal(err)
		}
		count, err := strconv.Atoi(values[4])
		if err != nil || start >= count || count == 0 {
			t.Fatalf("invalid program: %v", values)
		}
	}
	for _, question := range []string{"regexp-program", "regexp-program\n[", "regexp-program\n(a)\\1"} {
		out := &fields{}
		out.number(1)
		out.text("regexp-program")
		wire, err := regexpProgramFacts(out, question)
		if err != nil {
			t.Fatal(err)
		}
		values := decodedFields(t, wire)
		if len(values) != 4 || values[2] != "0" || values[3] == "" {
			t.Fatalf("invalid pattern accepted: %v", values)
		}
	}
}
