package lower

import (
	"errors"
	"strings"
	"testing"
)

// These programs typecheck, but stage 0 must refuse the protocols or representations it cannot
// implement soundly. The positive cases are held to Node in the slice's oracle fixtures.
func TestLibraryMapSetGapsStayRefused(t *testing.T) {
	for _, probe := range []struct{ name, source, reason string }{
		{"set-like object", `const other: ReadonlySetLike<number> = {
	size: 0,
	has: (value: number): boolean => false,
	keys: () => new Set<number>().keys(),
};
const combined = new Set<number>().union(other);
`, "union with anything but a concrete library Set"},
		{"different representations", `const combined = new Set<number>([1]).union(new Set<string>(['a']));`, "union between Sets whose elements have different representations"},
		{"dynamic thisArg", `function visit(value: number): void {}
new Set<number>([1]).forEach(visit, { label: 'context' });`, "forEach thisArg with a callback other than an arrow"},
		{"groupBy optional widening", `const values: number[] = [1];
const groups = Map.groupBy(values, (value: number | undefined): number => 0);`, "Map.groupBy grouped elements held otherwise than the input's"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q refusal, got %v", probe.reason, err)
			}
		})
	}
}
