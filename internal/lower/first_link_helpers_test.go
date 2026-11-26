package lower

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// firstLinkHelperProgram reads end and pos through a range type's union in width, each read lowering as a
// range_field_read helper of its own (rangeFieldDispatch), and other is a unit with no read.
const firstLinkHelperProgram = rangeTextRange + rangeNodeArray + "export function width(range: TextRange): number { return range.end - range.pos; }\nexport function other(): number { return 1; }\nconsole.log('ok');\n"

// firstLinkHelperLowering is a first-link lowering of firstLinkHelperProgram, with width's and other's function
// indexes and width's two helpers, each held to having been noted as width's.
func firstLinkHelperLowering(t *testing.T) (built *lowering, width, other int, helpers []int) {
	t.Helper()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.ts", firstLinkHelperProgram)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	built, err = lowerWithElementKey(&viewElementKey{}, loaded, check, entry, nil, false, false, ModuleFormatEcmaScript, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(built.traps) != 0 {
		t.Fatalf("Traps before the pass: %+v, want none", built.traps)
	}
	width, other = -1, -1
	for index, node := range built.functionNodes {
		switch node {
		case entry.Statements.Nodes[3]:
			width = index
		case entry.Statements.Nodes[4]:
			other = index
		}
	}
	for index, function := range built.result.Functions {
		if function.Name == "range_field_read" && built.functionNodes[index] == nil {
			helpers = append(helpers, index)
			if built.firstLinkHelperUnits[index] != entry.Statements.Nodes[3] {
				t.Fatalf("helper %d is noted as %v's, want width's", index, built.firstLinkHelperUnits[index])
			}
		}
	}
	if width < 0 || other < 0 || len(helpers) != 2 {
		t.Fatalf("width %d, other %d, helpers %v: want both bodies lowered and two range_field_read helpers", width, other, helpers)
	}
	return built, width, other, helpers
}

// firstLinkHelperPass is a judging pass modeled on checkViewWidenedReads at tsc's utilitiesPublic.ts:370:24: it
// stops at site, a read inside width, while any of helpers still holds its read, never a stub.
func firstLinkHelperPass(built *lowering, site string, helpers []int) func() error {
	return func() error {
		for _, helper := range helpers {
			if !firstLinkStubBody(built.result.Functions[helper].Body) {
				return &NotYet{Where: site, What: "a checked view read in range_field_read"}
			}
		}
		return nil
	}
}

// A range field read through a range type's union lowers as a helper with no declaration, whose code is its unit's,
// so when a judging pass's stop at the read traps the unit, the helper is stubbed with it, as the unit's own stub,
// and the pass judges on: one trap, no abandonment (#nxb3xh1). Without that the helper kept its read, the rerun met
// the same site again, and the pass gave up at a Gap, as checkViewWidenedReads did at 370:24 on tsc.
// Mutant (first_link_helpers.mutants.md): firstLinkStubHelpers' stubbing loop dropped; the Gap and the abandonment
// come back.
func TestFirstLinkHelperIsStubbedWithItsUnit(t *testing.T) {
	t.Parallel()
	built, width, _, helpers := firstLinkHelperLowering(t)
	declaration := built.functionNodes[width]
	site := built.program.Where(declaration.Body().AsBlock().Statements.Nodes[0])
	pass := firstLinkHelperPass(built, site, helpers)
	if err := built.firstLinkPass("checkHelperReads", false, pass(), pass); err != nil {
		t.Fatalf("the pass: got %v, want its stop recorded and nil", err)
	}
	if len(built.traps) != 1 || built.traps[0].Kind != "NotYet" || built.traps[0].Unit != "main.ts width" || built.traps[0].Phase != "checkHelperReads" {
		t.Fatalf("Traps: got %+v, want width's one NotYet", built.traps)
	}
	if got := built.result.TrapJudgmentsAbandoned; len(got) != 0 {
		t.Fatalf("TrapJudgmentsAbandoned: got %+v, want none", got)
	}
	stub := built.result.Functions[width].Body
	for _, helper := range helpers {
		if body := built.result.Functions[helper].Body; !firstLinkStubBody(body) || body[0] != stub[0] {
			t.Errorf("helper %d's body: got %+v, want width's stub %+v", helper, body, stub)
		}
	}
}

// A helper a live unit uses too is never stubbed when only one of its owners traps: its code is still the live
// unit's, so the pass meets its read again, a Gap, and gives its judgment up loudly, where a stub would trap the live
// unit at run time with the trapped one's reason. range_field_read is made per read site, so no two units share one,
// and this row is built in the IR: other is given a call of width's first helper. Width's second helper, which only
// width uses, is stubbed with it.
// Mutant (first_link_helpers.mutants.md): firstLinkStubHelpers' check of every owner dropped, so the trapped maker
// alone stubs the helper; the pass judges on and this goes red.
func TestFirstLinkHelperWithALiveOwnerIsAGap(t *testing.T) {
	t.Parallel()
	built, width, other, helpers := firstLinkHelperLowering(t)
	held := len(built.result.Locals)
	built.result.Locals = append(built.result.Locals, ir.Local{Name: "held", Type: ir.Union, Function: other})
	shared := built.result.Functions[helpers[0]]
	kept := shared.Body
	built.result.Functions[other].Body = append([]ir.Statement{ir.Evaluate{Value: ir.Call{Function: helpers[0], Arguments: []ir.Expression{ir.Read{Local: held, Of: ir.Union}}, Returns: shared.Returns}}}, built.result.Functions[other].Body...)
	declaration := built.functionNodes[width]
	site := built.program.Where(declaration.Body().AsBlock().Statements.Nodes[0])
	pass := firstLinkHelperPass(built, site, helpers)
	if err := built.firstLinkPass("checkHelperReads", false, pass(), pass); err != nil {
		t.Fatalf("the pass: got %v, want its stops recorded and nil", err)
	}
	if len(built.traps) != 2 || built.traps[0].Unit != "main.ts width" || built.traps[1].Kind != "Gap" || built.traps[1].Site != site {
		t.Fatalf("Traps: got %+v, want width's NotYet, then a Gap at %s", built.traps, site)
	}
	want := ir.TrapJudgmentAbandoned{Pass: "checkHelperReads", Site: site, Why: "Gap"}
	if got := built.result.TrapJudgmentsAbandoned; len(got) != 1 || got[0] != want {
		t.Fatalf("TrapJudgmentsAbandoned: got %+v, want %+v", got, want)
	}
	if body := built.result.Functions[helpers[0]].Body; firstLinkStubBody(body) || len(body) != len(kept) {
		t.Errorf("the helper other uses too: got %+v, want its code kept", body)
	}
	if !firstLinkStubBody(built.result.Functions[helpers[1]].Body) {
		t.Errorf("the helper only width uses: got %+v, want width's stub", built.result.Functions[helpers[1]].Body)
	}
}
