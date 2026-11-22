package lower

import (
	"errors"
	"strings"
	"testing"
)

// A function whose result is typed any is Refused, as any is anywhere (open question 16, refuseAnyValue), not
// NotYet as a representation stage 0 hasn't learned: tsc's tryParseJson (utilities.ts:7808) is the shape
// (#cxhztpd). A result stage 0 doesn't hold yet stays NotYet (not_any, the Seats scout's phantom-string-key
// reduction), so the guard reads any alone. Mutant: the guard dropped from signature() (functions.go), and the
// any rows read NotYet.
func TestFunctionAnyResultIsRefused(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name, source, at string
		refused          bool
	}{
		{"declaration", "function tryParseJson(text: string): any {\n\treturn text;\n}\nconsole.log(String(tryParseJson('5')));\n", "main.ts:1:10", true},
		{"arrow", "const parse = (text: string): any => text;\nconsole.log(String(parse('5')));\n", "main.ts:1:15", true},
		{"not_any", "type Key<T> = T & { __canonicalKey: never };\nfunction toKey(text: string): Key<string> {\n\tconsole.log(text.toLowerCase());\n\tthrow new Error('end');\n}\ntry { toKey('ABC'); } catch {}\n", "main.ts:2:10", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".ts")
			var refused *Refused
			var notYet *NotYet
			switch {
			case errors.As(err, &refused):
				if !row.refused || !strings.HasSuffix(refused.Where, row.at) || !strings.HasPrefix(refused.What, "a value of type any") {
					t.Fatalf("Refused at %s: %q, want refused=%v at %s", refused.Where, refused.What, row.refused, row.at)
				}
			case errors.As(err, &notYet):
				if row.refused || !strings.HasSuffix(notYet.Where, row.at) {
					t.Fatalf("NotYet at %s: %q, want refused=%v at %s", notYet.Where, notYet.What, row.refused, row.at)
				}
			default:
				t.Fatalf("lowered with %v, want a stop at %s", err, row.at)
			}
		})
	}
}
