package lower

import (
	"errors"
	"strings"
	"testing"
)

// A for...of whose object pattern reads a checked view's array with no element check records the use, as the
// identifier road does, so checkViewArrays stops it in the same words before destructureFrom reads each element's
// fields at the view's layout (Lower's must on #rm22wzv). The array is widened to an element union whose program
// class arm has no checked member contract, so its annotation fails softly and the loop has no element check, as
// TestViewCopyRoadsStopAtTheirRead's Map arm makes it. A Map arm can't serve here: Node and a Map are held
// differently, so destructureFrom refuses the pattern before the record is read, and a class instance is held as Node
// is. Both loops read size, the field both arms hold. Without the record checkViewArrays' rewrite walk still stops the
// loop, but at the program's file with no line, so the mutant that leaves the record to the identifier road alone reds
// the pattern row here by its position.
func TestForOfObjectPatternRecordsAViewedArrayUse(t *testing.T) {
	t.Parallel()
	const prelude = `interface Node { readonly kind: number; readonly size: number; }
interface Base { readonly tag: 1; }
class Counted { readonly kind: number = 2; readonly size: number = 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw = { tag: 1 as const, _children: [7] };
const base: Base = raw;
const view = base as SyntaxList;
const values: readonly (Node | Counted)[] = view._children;
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
			if !errors.As(err, &notYet) || notYet.What != want || !strings.HasSuffix(notYet.Where, "main.ts:9:1") {
				t.Fatalf("got %v, want %q at main.ts:9:1", err, want)
			}
		})
	}
}
