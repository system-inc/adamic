package lower

import (
	"errors"
	"testing"
)

// A RegExp match destructures into plain names, holes and defaults; what stays NotYet keeps its words,
// and an array of anything else is still not a tuple.
func TestMatchDestructureLowersOnlyPlainNames(t *testing.T) {
	t.Parallel()
	const found = "const match = /(a)(b)?/.exec(\"a\");\nif (match) {\n"
	for _, probe := range []struct {
		name, source, what string
	}{
		{"names, a hole and a default", found + "\tconst [, first, second = \"d\"] = match;\n\tconsole.log(`${first}${second}`);\n}\n", ""},
		{"a rest element", found + "\tconst [, ...others] = match;\n\tconsole.log(`${others.length}`);\n}\n", "a rest element when destructuring a RegExp match"},
		{"a nested pattern", found + "\tconst [, [first] = \"xy\"] = match;\n\tconsole.log(`${first}`);\n}\n", "a nested pattern when destructuring a RegExp match"},
		{"a plain array", "const words: string[] = [\"a\", \"b\"];\nconst [first] = words;\nconsole.log(`${first}`);\n", "destructuring anything but a tuple into [names]"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".a")
			var notYet *NotYet
			switch {
			case probe.what == "" && err != nil:
				t.Fatalf("want it to lower, got %v", err)
			case probe.what != "" && !errors.As(err, &notYet):
				t.Fatalf("want NotYet %s, got %v", probe.what, err)
			case probe.what != "" && notYet.What != probe.what:
				t.Fatalf("want NotYet %s, got NotYet %s", probe.what, notYet.What)
			}
		})
	}
}
