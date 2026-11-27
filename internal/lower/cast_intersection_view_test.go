package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A .ts cast from an intersection of plain objects that only reads a field narrower (castFieldRefinement, item 19
// on #bbssdcy) is castRoad's structural view, as one object's is (#gbr6vs2): tsc's toBuilderStateFileInfoForMultiEmit
// casts an Omit<FileInfo, "signature"> & { readonly signature: string | false | undefined } to FileInfo, whose
// signature is string | undefined, after a string test. No CheckedCast wraps it: every read through the view takes
// the slot as stored and checks it against the target's field (cast_intersection_view's oracle). These targets keep
// the refined field readonly. One that makes it writable is refused by the view-site rule, so the arm leaves it at
// its NotYet (TestCastStopNamesItsCause's field refinement row, tsc's own shape). Mutants: castRoad's gate without its
// intersection arm, and structuralViewCast refusing an intersection source again, each red these rows.
func TestIntersectionFieldRefinementLowersAsAStructuralView(t *testing.T) {
	t.Parallel()
	info := "interface Info { readonly version: string; readonly signature: string | undefined; readonly global: true | undefined } type Encoded = Omit<Info, \"signature\"> & { readonly signature: string | false | undefined }; "
	for _, probe := range []struct{ name, source string }{
		{"tsc's guarded cast", info + "function f(x: Encoded): Info { return typeof x.signature === 'string' ? x as Info : { version: x.version, signature: undefined, global: x.global }; } console.log(f({ version: 'v', signature: 's', global: undefined }).version);"},
		{"an unguarded cast", info + "function f(x: Encoded): Info { return x as Info; } console.log(String(f({ version: 'v', signature: undefined, global: undefined }).signature));"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowered, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			for _, function := range lowered.Functions {
				walk(function.Body, func(node any) bool {
					if cast, ok := node.(ir.CheckedCast); ok {
						t.Fatalf("a CheckedCast on %s: a field refinement is a view with no test at the cast", cast.Field)
					}
					return true
				})
			}
		})
	}
}

// What the intersection arm leaves stopped: a cast that also widens a writable field (castWidenedTarget's refusal,
// asked before the refinement), an intersection holding a primitive, and in .a, where a checked downcast's checks
// are refused. A row with no words holds only that the cast stays stopped. Mutant: castRoad's intersection arm
// without its castFieldRefinement condition reds "a target that widens another field".
func TestIntersectionFieldRefinementKeepsItsStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, extension string
		refused                 bool
		want                    string
	}{
		{"a target that widens another field", "interface Info { readonly version: string; signature: string | undefined; count: number } type Encoded = Omit<Info, \"signature\" | \"count\"> & { readonly signature: string | false | undefined; count: number | string }; function f(x: Encoded): Info { return x as Info; }", ".ts", true, "which Encoded can write string | number where number is read"},
		{"in .a", "interface Info { readonly version: string; signature: string | undefined } type Encoded = Omit<Info, \"signature\"> & { readonly signature: string | false | undefined }; function f(x: Encoded): Info { return x as Info; }", ".a", true, "in .a whose checks no proof erases"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var refused *Refused
			var notYet *NotYet
			switch {
			case err == nil:
				t.Fatal("lowered: the arm must leave this cast stopped")
			case probe.want == "":
				if !errors.As(err, &refused) && !errors.As(err, &notYet) {
					t.Fatalf("want a stop, got %v", err)
				}
			case probe.refused && (!errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want)):
				t.Fatalf("want a refusal holding %q, got %v", probe.want, err)
			case !probe.refused && (!errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.want)):
				t.Fatalf("want NotYet holding %q, got %v", probe.want, err)
			}
		})
	}
}
