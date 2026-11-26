package lower

import (
	"errors"
	"strings"
	"testing"
)

// A for...of whose object pattern reads a checked view's array with no element check records the use, as the
// identifier road does, so checkViewArrays stops it in the same words before destructureFrom reads each element's
// fields at the view's layout (Lower's must on #rm22wzv). The array's element union holds an arm with no contract, a
// Map, so its annotation fails softly and the loop has no element check, as TestViewCopyRoadsStopAtTheirRead's for-of
// row makes it. Both loops read size, the field both arms hold. Mutant: the record left to the identifier road alone,
// so the pattern loop lowers and reads the lie raw, reds the pattern row here.
func TestForOfObjectPatternRecordsAViewedArrayUse(t *testing.T) {
	t.Parallel()
	const prelude = `interface Node { readonly kind: number; readonly size: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw = { tag: 1 as const, _children: [7] };
const base: Base = raw;
const view = base as SyntaxList;
const values: readonly (Node | Map<string, number>)[] = view._children;
`
	for _, row := range []struct{ name, loop string }{
		{"identifier", "for (const item of values) { console.log(String(item.size)); }\n"},
		{"object pattern", "for (const { size } of values) { console.log(String(size)); }\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, prelude+row.loop, ".ts")
			want := "a checked view's array read through for-of (not checked yet)"
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != want || !strings.HasSuffix(notYet.Where, "main.ts:8:1") {
				t.Fatalf("got %v, want %q at main.ts:8:1", err, want)
			}
		})
	}
}
