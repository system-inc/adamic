package lower

import (
	"context"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// A first-link lowering whose checker holds an overflowed relation records the RelationOverflow row and goes
// on, and every row is Suspect, the rows recorded before it included: past the budget every relation lowering
// asked answered not related wherever it was asked, a cycle the finder didn't follow among them (#c0eypws).
// The overflow is there before the judging passes run, so the row is read at their baseline, and each pass
// after it abandons with the spent budget's Error row (TestFirstLinkBudgetSpentAtTheBaselineAbandonsEveryPass).
// The control lowers the same program first, with no overflow, and its one row, host's Math.random, isn't
// Suspect. The mark is planted as TestAnOverflowedRelationMakesLoweringNotYet plants it.
// Mutant: lowerCheckedWithCycleTypes' firstLinkSuspectEvery dropped, and no row is Suspect.
func TestFirstLinkRelationOverflowMarksEveryRowSuspect(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	checked, err := load.Load([]string{writeAgreementSource(t, "main.a", "function host(): number {\nreturn Math.random();\n}\nconsole.log(`${host()}`);\n")})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	typeChecker, release := checked.Checker(ctx, checked.Files()[0])
	defer release()
	control, err := LowerWithOptions(ctx, checked, Options{FirstLinkTraps: true})
	firstLinkStopsPartway(t, control, err, "main.a host")
	markRelationOverflow(t, typeChecker)
	program, err := LowerWithOptions(ctx, checked, Options{FirstLinkTraps: true})
	if err != nil {
		t.Fatalf("a first-link lowering with an overflow marked: %v, want the RelationOverflow row", err)
	}
	rows := firstLinkLocalSites(program)
	if len(rows) < 3 || rows[0].Reason != "Math.random" || rows[1].Phase != "RelationOverflow" || rows[1].Kind != "NotYet" || rows[1].Unit != "" || !strings.Contains(rows[1].Reason, "gave up on as too complex") {
		t.Fatalf("Traps: got %+v, want host's Math.random, the RelationOverflow row, then the passes' rows", rows)
	}
	for _, row := range rows[2:] {
		if row.Kind != "Error" || !strings.HasPrefix(row.Reason, "the checker's relation budget is spent: ") {
			t.Errorf("a row past the RelationOverflow row that isn't a pass abandoned at the spent budget: %+v", row)
		}
	}
	for _, row := range rows {
		if !row.Suspect {
			t.Errorf("a row that isn't Suspect after the relation overflowed: %+v", row)
		}
	}
}
