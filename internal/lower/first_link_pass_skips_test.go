package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// firstLinkStopsPartway holds a first-link lowering of a program whose one unit stops partway through
// lowering to that one stop: Math.random, a NotYet that traps the unit after what it holds before the call
// has lowered and been recorded, since a trap rolls nothing back. Every whole-program pass passes over what
// the trapped unit left in its tables, so the lowering returns no error, Traps is the one row, no row is
// Suspect, and no rewriting pass skipped its work.
func firstLinkStopsPartway(t *testing.T, program *ir.Program, err error, unit string) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	rows := firstLinkLocalSites(program)
	if len(rows) != 1 || rows[0].Kind != "NotYet" || rows[0].Reason != "Math.random" || rows[0].Unit != unit || rows[0].Phase != "Body" {
		t.Errorf("Traps: got %+v, want the one NotYet for Math.random in %s", rows, unit)
	}
	for _, row := range rows {
		if row.Suspect {
			t.Errorf("a Suspect row: %+v", row)
		}
	}
	if len(program.TrapRewritesSkipped) != 0 {
		t.Errorf("TrapRewritesSkipped: got %q, want none", program.TrapRewritesSkipped)
	}
}

// A for-of over a generator's call handed into a nested callee's Iterable parameter is an iterable view
// (iterable_views.go), recorded as the callee lowers, and the host it's nested in then stops at Math.random,
// trapping the callee with it. checkIterableViews passes over the trapped unit's view. Measured before the
// guard: the view's proof finds the callee's loop gone to the stub, and the pass records an Error, "the
// iterable view at main.a:3:85 has no loop to dispatch", skips its rewrite, and every later row is Suspect.
// Mutant: checkIterableViews' firstLinkSkips guards dropped, and this goes red.
func TestFirstLinkIterableViewInATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	source := "function* numbers(): Generator<number, undefined, undefined> { yield 1; yield 2; }\nfunction host(): number {\nfunction total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(numbers())));\nreturn Math.random();\n}\nconsole.log(`${host()}`);\n"
	program, err := lowerFirstLink(t, source, true)
	firstLinkStopsPartway(t, program, err, "main.a host")
}
