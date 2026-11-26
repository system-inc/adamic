package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
)

// A holder the cycle finder met with no node is refused with a Where, never handed to Program.Where as nil,
// which on tsc panicked in slotsOf (#ew207ss). An array and a tuple of Node, whose elements reach back through
// kids, are recorded as met with a nil node, and one unproven element write is planted, so slotsOf refuses
// each: the refusal's Where is whereOf's, a program declaration or the first file, never lib's Array.
// Mutants: whereOf returning f.where[proven] raw, and slotsOf panics in Program.Where; whereOf taking any
// declaration, and list is refused at lib.es5.d.ts.
func TestCycleRefusalOfAHolderMetWithNoNodeHasAWhere(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"list", "pair"} {
		f, types := cycleIndexTypes(t, `interface Node { kids: Node[] } declare const list: Node[]; declare const pair: [Node];`)
		holder := types[name]
		if holder == nil {
			t.Fatalf("no type for %s", name)
		}
		f.where[holder] = nil
		f.writes = []fresh.Write{{Kind: fresh.WriteElement, Function: -1, Why: "a write the test plants"}}
		var refused *Refused
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf("slotsOf(%s) panicked with the holder met with no node: %v", name, recovered)
				}
			}()
			if err := f.slotsOf(holder); !errors.As(err, &refused) {
				t.Errorf("slotsOf(%s): got %v, want the cycle refusal", name, err)
			}
		}()
		if refused == nil {
			continue
		}
		if !refused.cycle || refused.Where == "" {
			t.Errorf("slotsOf(%s): got %+v, want a cycle refusal with a Where", name, refused)
		}
		if strings.HasPrefix(refused.Where, "bundled:///") {
			t.Errorf("slotsOf(%s) refused at %s, in the library, where the program can't change anything", name, refused.Where)
		}
		t.Logf("%s refused at %s: %s", name, refused.Where, refused.What)
	}
}
