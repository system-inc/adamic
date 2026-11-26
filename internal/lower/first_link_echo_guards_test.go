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

// firstLinkEchoRows holds a first-link lowering to its rows: no Echo among them, no judgment abandoned, and each
// row's unit and phase as want gives them, in order.
func firstLinkEchoRows(t *testing.T, program *ir.Program, err error, want [][2]string) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	rows := firstLinkLocalSites(program)
	matches := len(rows) == len(want)
	for index := 0; matches && index < len(rows); index++ {
		matches = rows[index].Unit == want[index][0] && rows[index].Phase == want[index][1] && !rows[index].Echo
	}
	if !matches {
		t.Errorf("Traps: got %+v, want units and phases %q, no Echo", rows, want)
	}
	if len(program.TrapJudgmentsAbandoned) != 0 {
		t.Errorf("TrapJudgmentsAbandoned: got %+v, want none", program.TrapJudgmentsAbandoned)
	}
}

// Every construct lowering holds as unlowerable where it can't return an error is held in a first-link lowering,
// and one inside a unit trapped already is passed over, its code the stub (firstLinkHeldStops, #nxb3xh1). inner,
// hoisted in outer, holds one, a function value capturing the variable its own initializer declares, then
// Math.random traps outer; live holds another. Only live's traps, as unlowerable, with no Echo: on tsc
// createScanner's scanNumber was the first held, inside createScanner trapped by its var, an Echo that abandoned the
// judgment, and live's was never held at all.
// Mutants (first_link_echo_guards.mutants.md): the trapped unit's pass-over dropped, an Echo and an abandoned
// judgment; every construct held but the first dropped, and live's row goes missing.
func TestFirstLinkUnlowerableInATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLink(t, "function outer(): number {\n\tfunction inner(): number {\n\t\tconst self = (): number => self.length;\n\t\treturn self();\n\t}\n\treturn Math.random() + inner();\n}\nfunction live(): number {\n\tconst me = (): number => me.length;\n\treturn me();\n}\nconsole.log(`${outer() + live()}`);\n", true)
	firstLinkEchoRows(t, program, err, [][2]string{{"main.a outer.inner.self.anonymous_1", "Body"}, {"main.a outer", "Body"}, {"main.a live.me.anonymous_1", "Body"}, {"main.a live", "unlowerable"}})
}

// A function declared in a unit and lowered before the unit's body stopped, a hoisted nested declaration, is the
// unit's code, so the unit's trap stubs it too (firstLinkStubInside), and no pass meets it as live (#nxb3xh1). copy,
// hoisted in outer, spreads a parameter in a program with a static constructor object, then Math.random traps
// outer: checkAccessorSpreads met copy's spread, inside the trapped outer, an Echo, as it met
// copyFileReferenceAsSynthetic's in transformRoot on tsc, and checkViewWidenedReads isPinnedCommentLocal's read in
// emitDetachedComments.
// Mutant (first_link_echo_guards.mutants.md): firstLinkBody's stub of what it lowered inside dropped, and the Echo
// comes back.
func TestFirstLinkNestedFunctionOfABodyTrapIsStubbed(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLinkTS(t, stopWhereStatics("function outer(point: Point): number {\n    function copy(): Point {\n        return { ...point };\n    }\n    return copy().x + Math.random();\n}\nconsole.log(String(outer({ x: 1 }) + Counter.start()));\n"), true)
	firstLinkEchoRows(t, program, err, [][2]string{{"main.ts outer", "Body"}})
	for _, function := range program.Functions {
		if function.Name == "copy" && !firstLinkStubBody(function.Body) {
			t.Errorf("copy's body: got %+v, want outer's stub", function.Body)
		}
	}
}

// The cycle finder reads every captured variable from the program's locals, a trapped unit's included, whose cells
// its stub never makes, so a first-link lowering passes over a variable declared inside a trapped unit (#nxb3xh1).
// trapped's parameter b, of type unknown, is captured by read, and assigning read to it traps trapped: findCycles
// met b, an Echo, as it met isJsonEqual's b on tsc.
// Mutant (first_link_echo_guards.mutants.md): loose_slots.go's pass-over dropped, and the Echo comes back.
func TestFirstLinkCapturedVariableOfATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLinkTS(t, "function trapped(b: unknown): number {\n\tconst read = (): unknown => b;\n\tb = read;\n\treturn Math.random();\n}\nconsole.log(`${trapped(0)}`);\n", true)
	firstLinkEchoRows(t, program, err, [][2]string{{"main.ts trapped", "Body"}})
}

// The cycle finder's pass-over is for a variable a trapped unit declares, and never for one a live unit does, though
// a trapped unit hands it a value (Compiler's condition on #nxb3xh1: no input a trapped unit gives a live unit's
// judgment is dropped). live's parameter c, of type unknown, captured by read, still stops findCycles in live, where
// relay, live too, hands it what trapped hands relay before Math.random traps trapped.
// Mutant (first_link_echo_guards.mutants.md): the pass-over widened to every captured variable once any unit is
// trapped, and live's stop goes missing.
func TestFirstLinkLiveCapturedVariableIsStillJudged(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLinkTS(t, "function live(c: unknown): number {\n\tconst read = (): unknown => c;\n\treturn read() === undefined ? 0 : 1;\n}\nfunction relay(d: unknown): number {\n\treturn live(d);\n}\nfunction trapped(b: unknown): number {\n\treturn relay(b) + Math.random();\n}\nconsole.log(`${relay(0) + trapped(0)}`);\n", true)
	firstLinkEchoRows(t, program, err, [][2]string{{"main.ts trapped", "Body"}, {"main.ts live", "findCycles"}})
}
