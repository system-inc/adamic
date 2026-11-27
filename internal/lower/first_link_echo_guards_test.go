package lower

import (
	"strings"
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
// relay, live too, hands it what trapped hands relay before Math.random traps trapped. Once live traps, the pass runs
// again and meets c inside a trapped unit, which the pass-over passes over, so the pass judges on.
// Mutants (first_link_echo_guards.mutants.md): the pass-over widened to every captured variable once any unit is
// trapped, and live's stop goes missing; the pass-over dropped, and the rerun's c is a Gap.
func TestFirstLinkLiveCapturedVariableIsStillJudged(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLinkTS(t, "function live(c: unknown): number {\n\tconst read = (): unknown => c;\n\treturn read() === undefined ? 0 : 1;\n}\nfunction relay(d: unknown): number {\n\treturn live(d);\n}\nfunction trapped(b: unknown): number {\n\treturn relay(b) + Math.random();\n}\nconsole.log(`${relay(0) + trapped(0)}`);\n", true)
	firstLinkEchoRows(t, program, err, [][2]string{{"main.ts trapped", "Body"}, {"main.ts live", "findCycles"}})
}

// looseHolder is a value with a loose field, data, so a literal spreading one into a Holder is an initialization
// the cycle finder judges (loose_slots.go's looseInits): it can reach the Holder holding it back through data.
const looseHolder = "interface Holder {\n    readonly x: number;\n    data: unknown;\n}\n"

// A literal's initialization is noted while lowering, a trapped unit's too, whose code its stub replaced, so the
// cycle finder passes over one made inside a trapped unit (#nxb3xh1). copy spreads a Holder into a Holder in a
// program with a static constructor object, so checkAccessorSpreads traps copy at the spread, then findCycles met
// the same spread, an Echo, as it met executeCommandLineWorker's spread into ParsedCommandLine on tsc
// (executeCommandLine.ts:730:19).
// Mutant (first_link_echo_guards.mutants.md): the initialization's pass-over dropped, and the Echo comes back.
func TestFirstLinkLooseInitOfATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLinkTS(t, stopWhereStatics(looseHolder+"function copy(holder: Holder): Holder {\n    return { ...holder };\n}\nconsole.log(String(copy({ x: 1, data: 0 }).x + Counter.start()));\n"), true)
	firstLinkEchoRows(t, program, err, [][2]string{{"main.ts copy", "checkAccessorSpreads"}})
}

// The initialization's pass-over is for a literal a trapped unit makes, and never for one a live unit does, though
// a trapped unit hands it a value (Compiler's condition on #nxb3xh1). copy, live, spreads a Holder into a Holder
// and still stops findCycles there, where trapped hands copy a Holder of its own before Math.random traps trapped.
// Once copy traps, the pass runs again and meets the spread inside a trapped unit, which the pass-over passes over.
// Mutants (first_link_echo_guards.mutants.md): the pass-over widened to every initialization once any unit is
// trapped, and copy's stop goes missing; the pass-over dropped, and the rerun's spread is a Gap.
func TestFirstLinkLiveLooseInitIsStillJudged(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLinkTS(t, looseHolder+"function copy(holder: Holder): Holder {\n    return { ...holder };\n}\nfunction trapped(): number {\n    return copy({ x: 2, data: 0 }).x + Math.random();\n}\nconsole.log(String(copy({ x: 1, data: 0 }).x + trapped()));\n", true)
	firstLinkEchoRows(t, program, err, [][2]string{{"main.ts trapped", "Body"}, {"main.ts copy", "findCycles"}})
	if rows := firstLinkLocalSites(program); len(rows) == 2 && !strings.HasPrefix(rows[1].Reason, "a spread into Holder") {
		t.Errorf("copy's stop: got %q, want the spread's initialization, a spread into Holder", rows[1].Reason)
	}
}

// firstLinkCatchBinding is a function name whose catch binding e is captured by read, and read is kept in an
// unknown[], so lib's Error, what's caught, reaches e's cell back through its cause and what H holds. tail ends its
// body.
func firstLinkCatchBinding(name, tail string) string {
	return "function " + name + "(): number {\n\ttry {\n\t\tconsole.log('try');\n\t} catch (e) {\n\t\tconst read = (): unknown => e;\n\t\tconst kept: unknown[] = [read];\n\t\tconsole.log(`${kept.length}`);\n\t}\n\treturn " + tail + ";\n}\n"
}

// A catch binding is read from the program's locals, a trapped unit's included, so the cycle finder passes over one
// declared inside a trapped unit, as it does a captured variable (#nxb3xh1). trapped's e, captured by read, can
// hold a caught Error that reaches its own cell back, and Math.random traps trapped.
// Mutant (first_link_echo_guards.mutants.md): the catch binding's pass-over dropped, and findCycles meets e, an
// Echo.
func TestFirstLinkCatchBindingOfATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLinkTS(t, firstLinkCatchBinding("trapped", "Math.random()")+"console.log(`${trapped()}`);\n", true)
	firstLinkEchoRows(t, program, err, [][2]string{{"main.ts trapped", "Body"}})
}

// The catch binding's pass-over is for one a trapped unit declares: live's e, the same shape, still stops findCycles
// in live while Math.random traps trapped. Once live traps, the pass runs again and passes over its e.
// Mutants (first_link_echo_guards.mutants.md): the pass-over widened to every catch binding once any unit is
// trapped, and live's stop goes missing; the pass-over dropped, and the rerun's e is a Gap.
func TestFirstLinkLiveCatchBindingIsStillJudged(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLinkTS(t, firstLinkCatchBinding("live", "0")+"function trapped(): number {\n\treturn Math.random();\n}\nconsole.log(`${live() + trapped()}`);\n", true)
	firstLinkEchoRows(t, program, err, [][2]string{{"main.ts trapped", "Body"}, {"main.ts live", "findCycles"}})
	if rows := firstLinkLocalSites(program); len(rows) == 2 && !strings.HasPrefix(rows[1].Reason, "'e', a catch binding") {
		t.Errorf("live's stop: got %q, want its catch binding e", rows[1].Reason)
	}
}
