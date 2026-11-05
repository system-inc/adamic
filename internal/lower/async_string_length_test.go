package lower

import (
	"errors"
	"strings"
	"testing"
)

// A template or a concatenation outside a try can throw past V8's longest string, so an async function
// that makes one may reject, and the promise consumption proof keeps that edge (Adamic's ruling on
// #c3bk0fp). word below rejects only that way. Two of its promises pending, awaited in one template,
// stop as NotYet named for the string-length limit, at that template: if the first await throws,
// the second's rejection goes unobserved. One pending promise awaited first in a template still lowers,
// since nothing in the statement can throw before that await observes it. A callee that throws for
// another reason keeps the general stop.
func TestAsyncStringLengthConsumption(t *testing.T) {
	t.Parallel()
	const word = "async function word(text: string): Promise<string> {\n\treturn `${text}!`;\n}\n"
	const general = "a promise not awaited in the same function before an escaping completion"
	for _, probe := range []struct {
		name, source, where, want string
	}{
		{"two pending awaits in one template", word + "const first = word('a');\nconst second = word('b');\nconsole.log(`${await first}|${await second}`);\n", "main.a:6:", stringLengthConsumption},
		{"one pending await first in a template", word + "const first = word('a');\nconsole.log(`${await first}|`);\n", "", ""},
		{"a callee that throws otherwise", "async function fail(text: string): Promise<string> {\n\tif (text === 'b') {\n\t\tthrow new Error(text);\n\t}\n\treturn text;\n}\nconst first = fail('a');\nconst second = fail('b');\nconsole.log(`${await first}|${await second}`);\n", "main.a:8:", general},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if probe.want == "" {
				if err != nil {
					t.Fatalf("must lower: %v", err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != probe.want || !strings.HasPrefix(notYet.Where, probe.where) {
				t.Fatalf("want NotYet %q at %s, got %v", probe.want, probe.where, err)
			}
		})
	}
}
