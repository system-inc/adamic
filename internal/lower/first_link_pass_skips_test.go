package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// firstLinkOneTrap holds a first-link lowering to the one trap its program stops at: every whole-program pass
// passes over what the trapped unit left in its tables, so the lowering returns no error, Traps is that one
// row, no row is Suspect, and no rewriting pass skipped its work.
func firstLinkOneTrap(t *testing.T, program *ir.Program, err error, want ir.Trap) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	rows := firstLinkLocalSites(program)
	if len(rows) != 1 || rows[0].Kind != want.Kind || rows[0].Reason != want.Reason || rows[0].Unit != want.Unit || rows[0].Phase != want.Phase {
		t.Errorf("Traps: got %+v, want the one %s for %s in %s", rows, want.Kind, want.Reason, want.Unit)
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

// firstLinkStopsPartway is firstLinkOneTrap for a unit that stops partway through lowering, at Math.random, a
// NotYet that traps the unit after what it holds before the call has lowered and been recorded, since a trap
// rolls nothing back.
func firstLinkStopsPartway(t *testing.T, program *ir.Program, err error, unit string) {
	t.Helper()
	firstLinkOneTrap(t, program, err, ir.Trap{Kind: "NotYet", Reason: "Math.random", Unit: unit, Phase: "Body"})
}

// A for-of over a generator's call handed into a nested callee's Iterable parameter is an iterable view
// (iterable_views.go), recorded as the callee lowers, and the host it's nested in then stops at Math.random,
// trapping the callee with it. checkIterableViews passes over the trapped unit's view.
// Mutant: checkIterableViews' firstLinkSkips guards dropped, and the view's proof meets the callee's loop gone
// to the stub: an Echo, "a view over an iterable of unproven origin", and checkIterableViews' rewrite skipped.
func TestFirstLinkIterableViewInATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	source := "function* numbers(): Generator<number, undefined, undefined> { yield 1; yield 2; }\nfunction host(): number {\nfunction total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(numbers())));\nreturn Math.random();\n}\nconsole.log(`${host()}`);\n"
	program, err := lowerFirstLink(t, source, true)
	firstLinkStopsPartway(t, program, err, "main.a host")
}

// A for-of over a generator's call holds the generator in a local of its own, which the loop's block declares,
// and the function then stops at Math.random, so its body is the trap's stub alone and the local's Declare is
// gone with the rest. checkGeneratorRoads passes over a local of a function whose body is the stub.
// Mutant: checkGeneratorRoads' firstLinkStubBody guard dropped, and the pass records an Error, "the generator
// local generator is held outside its consuming loop", which makes every later row Suspect.
func TestFirstLinkGeneratorLocalInATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	source := "function* numbers(): Generator<number, undefined, undefined> { yield 1; yield 2; }\nfunction second(): number {\nfor (const value of numbers()) { console.log(String(value)); }\nreturn Math.random();\n}\nconsole.log(`${second()}`);\n"
	program, err := lowerFirstLink(t, source, true)
	firstLinkStopsPartway(t, program, err, "main.a second")
}

// A function constructor the refusal pass refuses, for an ==, is trapped before anything lowers, and its
// constructor lowers as the stub, which allocates no this. The new through the allocator's road lies in a
// unit that lowers. layFunctionConstructorWindows passes over the trapped constructor.
// Mutant: layFunctionConstructorWindows' firstLinkSkips guard dropped, and the stub's window is laid: an Echo,
// "a function constructor whose object may be seen before it writes pos", and the pass's rewrite skipped.
func TestFirstLinkFunctionConstructorInATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	constructor := "function NodeObject(this: Nd, kind: number, pos: number, end: number): void {\n" +
		"    this.pos = pos;\n" +
		"    this.end = end;\n" +
		"    this.kind = kind == 7 ? 7 : kind;\n" +
		"}\n"
	program, err := lowerFirstLinkTS(t, functionConstructorProgram(constructor, "", ""), true)
	firstLinkOneTrap(t, program, err, ir.Trap{Kind: "Refused", Reason: "==", Unit: "main.ts NodeObject", Phase: "Refusal"})
}

// A try around a call whose failure is a panic natively but a throw on Node, a repeat by a count that isn't a
// constant, is recorded as its function lowers, and the function then stops at Math.random. exceptions passes
// over the trapped unit's try.
// Mutant: exceptions' firstLinkSkips guard dropped, and the try is judged again: an Echo, "a try around
// repeat", and the exceptions rewrite skipped.
func TestFirstLinkTryInATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	source := "function second(count: number): number {\n\ttry {\n\t\tconsole.log('ab'.repeat(count));\n\t} catch (error) {\n\t\tconsole.log('caught');\n\t}\n\treturn Math.random();\n}\nconsole.log(`${second(2)}`);\n"
	program, err := lowerFirstLinkTS(t, source, true)
	firstLinkStopsPartway(t, program, err, "main.ts second")
}

// A write of an object field through a checked view is recorded for the source-slot certificate as the view
// is made, and the function holding both then stops at Math.random. A read of the field through a type
// parameter elsewhere would give back the write's stop (sourceSlotGenericReads). certifySourceSlots passes over
// the trapped unit's write, which never runs.
// Mutant: certifySourceSlots' firstLinkSkips guard dropped, and the generic read gives back the write's stop: an
// Echo, "writing a checked object field without its source-slot type certificate", and the pass's rewrite
// skipped.
func TestFirstLinkSourceSlotWriteInATrappedUnitIsPassedOver(t *testing.T) {
	t.Parallel()
	source := "interface Base { readonly kind: 'box' | 'other' }\n" +
		"interface Box extends Base { readonly kind: 'box'; child: { ready: boolean } }\n" +
		"function readyOf<T extends Box>(box: T): boolean { return box.child.ready; }\n" +
		"function second(): number {\n" +
		"\tconst raw = { kind: 'box' as const, child: { ready: true, other: 42 } };\n" +
		"\tconst held: Base = raw;\n" +
		"\tconst view = held as Box;\n" +
		"\tview.child = { ready: true };\n" +
		"\tconsole.log(String(raw.child.other));\n" +
		"\treturn Math.random();\n" +
		"}\n" +
		"console.log(`${second()}`);\n"
	program, err := lowerFirstLinkTS(t, source, true)
	firstLinkStopsPartway(t, program, err, "main.ts second")
}

// firstLinkBuilt lowers a .a program by lowerWithElementKey, as a first-link lowering when traps is set, and
// hands back the lowering, so a test drives one of its passes by hand.
func firstLinkBuilt(t *testing.T, source string, traps bool) (*lowering, *load.Program) {
	t.Helper()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.a", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	built, err := lowerWithElementKey(&viewElementKey{}, loaded, check, entry, nil, false, false, ModuleFormatEcmaScript, traps, nil)
	if err != nil {
		t.Fatal(err)
	}
	return built, loaded
}

// A Go panic inside a whole-program pass is a row of kind "Error" in a first-link lowering, never the end of
// it: firstLinkRun recovers it on the pass's first run and on a run again past a stop, in the panic's words,
// with no unit, and every row after it is Suspect. No source road reaches a pass's panic honestly (each would
// be a compiler bug), so this drives a real first-link lowering's firstLinkRun with a pass that panics. Without
// the option, the same pass's panic reaches the caller as it always has.
// Mutant: firstLinkRun repanicking what it recovers, and this panics, "a pass's nil dereference".
func TestFirstLinkPassPanicIsAnErrorRow(t *testing.T) {
	t.Parallel()
	built, loaded := firstLinkBuilt(t, firstLinkTwoFunctions, true)
	if err := built.firstLinkRun("checkSomething", false, func() error { panic("a pass's nil dereference") }); err != nil {
		t.Fatalf("a pass's panic: got %v, want it recorded and nil", err)
	}
	row := built.traps[len(built.traps)-1]
	if want := (ir.Trap{Kind: "Error", Reason: "a pass's nil dereference", Phase: "checkSomething"}); row != want {
		t.Fatalf("the pass's panic's row: got %+v, want %+v", row, want)
	}
	// A pass that stops in a unit runs again past it, and its panic there is recovered too.
	first := loaded.Files()[0].Statements.Nodes[0]
	runs := 0
	again := func() error {
		runs++
		if runs == 1 {
			return built.notYet(first, "a stop before the panic")
		}
		panic("a pass's bad index")
	}
	if err := built.firstLinkRun("checkOther", false, again); err != nil {
		t.Fatalf("a pass's panic on its run again: got %v, want it recorded and nil", err)
	}
	rows := built.traps[len(built.traps)-2:]
	if rows[0].Kind != "NotYet" || rows[0].Unit != "main.a first" || !rows[0].Suspect {
		t.Fatalf("the stop before the panic: got %+v, want a Suspect NotYet in main.a first", rows[0])
	}
	if want := (ir.Trap{Kind: "Error", Reason: "a pass's bad index", Phase: "checkOther", Suspect: true}); rows[1] != want {
		t.Fatalf("the panic on the run again: got %+v, want %+v", rows[1], want)
	}
	off, _ := firstLinkBuilt(t, "console.log('ran');\n", false)
	recovered := func() (recovered any) {
		defer func() { recovered = recover() }()
		if err := off.firstLinkRun("checkSomething", false, func() error { panic("a pass's nil dereference") }); err != nil {
			t.Errorf("without the option: got %v, want the panic", err)
		}
		return nil
	}()
	if text, isText := recovered.(string); !isText || !strings.Contains(text, "nil dereference") {
		t.Fatalf("without the option: recovered %v, want the pass's panic", recovered)
	}
	if err := off.firstLinkRun("checkSomething", false, func() error { return errors.New("lower: compiler bug: a plain error") }); err == nil {
		t.Fatal("without the option: a pass's plain error was swallowed")
	}
}
