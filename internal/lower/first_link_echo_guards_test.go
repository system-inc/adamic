package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A helper whose unit traps in its own body, after the helper was made, is used by nothing any more: the unit's
// body is its stub. firstLinkHeldStops stubs it with its maker before any pass runs, so no pass meets its code as
// live (#nxb3xh1). width's two range field reads lower as range_field_read helpers, then Math.random traps width.
// Mutant (first_link_echo_guards.mutants.md): firstLinkHeldStops' helper sweep dropped, and the helpers keep their
// code.
func TestFirstLinkHelperOrphanedByABodyTrapIsStubbed(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLinkTS(t, rangeTextRange+rangeNodeArray+"export function width(range: TextRange): number { return range.end - range.pos + Math.random(); }\nconsole.log('ok');\n", true)
	firstLinkOneTrap(t, program, err, ir.Trap{Kind: "NotYet", Reason: "Math.random", Unit: "main.ts width", Phase: "Body"})
	var stub ir.Statement
	helpers := 0
	for _, function := range program.Functions {
		if function.Name == "width" && firstLinkStubBody(function.Body) {
			stub = function.Body[0]
		}
	}
	for _, function := range program.Functions {
		if function.Name != "range_field_read" {
			continue
		}
		helpers++
		if !firstLinkStubBody(function.Body) || function.Body[0] != stub {
			t.Errorf("a range_field_read helper width made: got %+v, want width's stub %+v", function.Body, stub)
		}
	}
	if stub == nil || helpers != 2 {
		t.Fatalf("width's stub %+v and %d range_field_read helpers, want the stub and two", stub, helpers)
	}
}

// checkRangeUnionReads' stop inside a helper names the unit that made it (stopWhere), where it named no site, a Gap
// that abandoned the pass on tsc (#nxb3xh1): the stop places in width, which traps, and its helpers are stubbed
// with it, so the pass judges on. The narrow of an untyped union local the pass stops (TestRangeUnionNarrowStops)
// is put in width's first range_field_read helper.
// Mutant (first_link_echo_guards.mutants.md): stopWhere's helper arm dropped, and the stop is a Gap again.
func TestFirstLinkRangeUnionStopInAHelperTrapsItsMaker(t *testing.T) {
	t.Parallel()
	built, width, _, helpers := firstLinkHelperLowering(t)
	held := len(built.result.Locals)
	built.result.Locals = append(built.result.Locals, ir.Local{Name: "held", Type: ir.Union, Function: helpers[0]})
	built.result.Functions[helpers[0]].Body = append([]ir.Statement{ir.Evaluate{Value: ir.Narrow{Value: ir.Read{Local: held, Of: ir.Union}, To: ir.Object}}}, built.result.Functions[helpers[0]].Body...)
	if err := built.firstLinkPass("checkRangeUnionReads", false, built.checkRangeUnionReads(), built.checkRangeUnionReads); err != nil {
		t.Fatalf("the pass: got %v, want its stop recorded and nil", err)
	}
	rows := firstLinkLocalSites(&ir.Program{Traps: built.traps})
	if len(rows) != 1 || rows[0].Kind != "NotYet" || rows[0].Site != "main.ts:4:1" || rows[0].Unit != "main.ts width" || rows[0].Phase != "checkRangeUnionReads" {
		t.Fatalf("Traps: got %+v, want one NotYet at width's declaration, main.ts:4:1", rows)
	}
	if got := built.result.TrapJudgmentsAbandoned; len(got) != 0 {
		t.Fatalf("TrapJudgmentsAbandoned: got %+v, want none", got)
	}
	stub := built.result.Functions[width].Body
	for _, helper := range helpers {
		if body := built.result.Functions[helper].Body; !firstLinkStubBody(body) || body[0] != stub[0] {
			t.Errorf("helper %d's body: got %+v, want width's stub", helper, body)
		}
	}
}
