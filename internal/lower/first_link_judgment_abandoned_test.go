package lower

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A judging pass whose rerun meets a stop it can't place gives up on every unit past it, so a first-link
// lowering names it in TrapJudgmentsAbandoned, and every row after is Suspect (#nxb3xh1). On tsc,
// checkAccessorSpreads and checkRangeUnionReads stopped at a NotYet whose Where named only the file, the
// rest went unjudged, and an unjudged checked view read panicked native emit. This drives a real first-link
// lowering's pass with that shape: a stop in first's body traps first, and the rerun stops at the file alone,
// a Gap. A stop in first again is then an Echo, abandoned too. A rewriting pass's give-up stays
// TrapRewritesSkipped's alone.
// Mutant: firstLinkJudgmentAbandon records nothing, and this goes red on the missing entries and the row after
// that isn't Suspect.
func TestFirstLinkJudgingPassGivingUpIsRecorded(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.a", firstLinkTwoFunctions)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	built, err := lowerWithElementKey(&viewElementKey{}, loaded, check, entry, nil, false, false, ModuleFormatEcmaScript, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(built.result.TrapJudgmentsAbandoned) != 0 {
		t.Fatalf("before the pass: TrapJudgmentsAbandoned %+v, want none", built.result.TrapJudgmentsAbandoned)
	}
	inFirst := built.program.Where(entry.Statements.Nodes[0].Body().AsBlock().Statements.Nodes[0])
	fileOnly := built.result.Source
	rows := len(built.traps)
	reruns := 0
	if err := built.firstLinkPass("checkSomething", false, &NotYet{Where: inFirst, What: "a stop in first"}, func() error {
		reruns++
		return &NotYet{Where: fileOnly, What: "a stop that names only its file"}
	}); err != nil {
		t.Fatalf("the pass: got %v, want its stops recorded and nil", err)
	}
	if reruns != 1 {
		t.Fatalf("the pass ran again %d times, want once", reruns)
	}
	recorded := built.traps[rows:]
	if len(recorded) != 2 || recorded[0].Kind != "NotYet" || recorded[0].Unit != "main.a first" || recorded[0].Suspect || recorded[1].Kind != "Gap" || recorded[1].Site != fileOnly {
		t.Fatalf("the pass's rows: got %+v, want first's NotYet, then a Gap at %s", recorded, fileOnly)
	}
	want := []ir.TrapJudgmentAbandoned{{Pass: "checkSomething", Site: fileOnly, Why: "Gap"}}
	if got := built.result.TrapJudgmentsAbandoned; len(got) != 1 || got[0] != want[0] {
		t.Fatalf("TrapJudgmentsAbandoned after the Gap: got %+v, want %+v", got, want)
	}
	if err := built.firstLinkPass("checkAgain", false, &NotYet{Where: inFirst, What: "a stop in first again"}, nil); err != nil {
		t.Fatalf("the echo: got %v, want it recorded and nil", err)
	}
	echo := built.traps[len(built.traps)-1]
	if !echo.Echo || !echo.Suspect {
		t.Fatalf("the echo's row: got %+v, want an Echo, Suspect after the abandoned judgment", echo)
	}
	want = append(want, ir.TrapJudgmentAbandoned{Pass: "checkAgain", Site: inFirst, Why: "Echo"})
	if got := built.result.TrapJudgmentsAbandoned; len(got) != 2 || got[1] != want[1] {
		t.Fatalf("TrapJudgmentsAbandoned after the Echo: got %+v, want %+v", got, want)
	}
	if err := built.firstLinkPass("rewriteSomething", true, &NotYet{Where: fileOnly, What: "a rewrite's stop that names only its file"}, nil); err != nil {
		t.Fatalf("the rewriting pass: got %v, want it recorded and nil", err)
	}
	if got := built.result.TrapJudgmentsAbandoned; len(got) != 2 {
		t.Fatalf("a rewriting pass's Gap: TrapJudgmentsAbandoned %+v, want it left to TrapRewritesSkipped", got)
	}
	if got := built.result.TrapRewritesSkipped; len(got) != 1 || got[0] != "rewriteSomething" {
		t.Fatalf("TrapRewritesSkipped: got %q, want the rewriting pass", got)
	}
}
