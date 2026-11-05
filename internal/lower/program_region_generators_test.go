package lower

import (
	"errors"
	"strings"
	"testing"
)

// A program with generators may have a Program region (#8xcz03h) only while no generator frame can join
// it or reach a member by a road the cycle walk doesn't see. These pins hold the roads that would let one
// (lower.go names them where the guard stood): a frame's captured cell, and a generator value in a slot.
// Each program has a self-referential Tree with two linked literals, so the region really engages, and
// lowers with it on. When a slot can hold a generator (#qk8rztp S2c-4), each pin that goes green is a
// road the cycle walk must follow through the frame's links first.

// programRegionGeneratorTree is the self-referential type and the generator walking it that every pin
// shares. TReturn is undefined, not void, so a generator used as a value reaches lowering's stop rather
// than the void rules' refusal.
const programRegionGeneratorTree = "interface Tree { name: string; next: Tree | undefined; }\n" +
	"function* walk(start: Tree): Generator<Tree, undefined, undefined> { let at: Tree | undefined = start; while (at !== undefined) { yield at; at = at.next; } }\n"

// programRegionGeneratorRun is a run() whose two linked Tree literals are the region's members, with
// body after them.
func programRegionGeneratorRun(body string) string {
	return "function run(): void {\n const root: Tree = { name: `root${1}`, next: undefined };\n const branch: Tree = { name: `branch${2}`, next: root };\n" + body + "\n}\nrun();\n"
}

// programRegionGeneratorStops lowers source with the Program region on and wants a NotYet saying reason.
func programRegionGeneratorStops(t *testing.T, source, reason string) {
	t.Helper()
	_, err := programLowerSource(t, source)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, reason) {
		t.Fatalf("got %v, want a NotYet saying %q", err, reason)
	}
}

// The control: the same tree and generator, consumed by a for-of, lower with the region on and the Tree
// literals in it, so each pin below is stopped by its own road and not by the region.
func TestProgramRegionGeneratorProgramEngages(t *testing.T) {
	t.Parallel()
	program, err := programLowerSource(t, programRegionGeneratorTree+programRegionGeneratorRun(" for (const tree of walk(branch)) { console.log(tree.name); }"))
	if err != nil {
		t.Fatal(err)
	}
	if !program.HasGenerators() || len(program.ProgramTypes) == 0 {
		t.Fatalf("generators %t, Program types %d: want a generator and the region engaged", program.HasGenerators(), len(program.ProgramTypes))
	}
}

// Invariant 1: a generator's frame has no captured cell, so FrameEnvironment, which holds only captured
// locals, is empty for it and the region's environment walk never reaches a frame.
func TestProgramRegionGeneratorCapturedLocalStops(t *testing.T) {
	t.Parallel()
	source := "interface Tree { name: string; next: Tree | undefined; }\n" +
		"function* named(start: Tree): Generator<string, void, undefined> { const label = start.name; const tell = (): string => label; yield tell(); }\n" +
		programRegionGeneratorRun(" for (const name of named(branch)) { console.log(name); }")
	programRegionGeneratorStops(t, source, "whose local 'label' a closure captures")
}

func TestProgramRegionGeneratorNestedFunctionStops(t *testing.T) {
	t.Parallel()
	source := "interface Tree { name: string; next: Tree | undefined; }\n" +
		"function* named(start: Tree): Generator<string, void, undefined> { function suffix(): string { return '!'; } yield start.name + suffix(); }\n" +
		programRegionGeneratorRun(" for (const name of named(branch)) { console.log(name); }")
	programRegionGeneratorStops(t, source, "declaring a nested function")
}

// Invariant 3: a generator value never reaches a slot, a field, an array element or a closure's capture.
func TestProgramRegionGeneratorStoredInFieldStops(t *testing.T) {
	t.Parallel()
	source := programRegionGeneratorTree + programRegionGeneratorRun(" const holder = { steps: walk(branch) };\n for (const tree of holder.steps) { console.log(tree.name); }")
	programRegionGeneratorStops(t, source, "whose generator is used other than by a for-of over its call or a call as a statement")
}

func TestProgramRegionGeneratorPushedIntoArrayStops(t *testing.T) {
	t.Parallel()
	source := programRegionGeneratorTree + programRegionGeneratorRun(" const held: Generator<Tree, undefined, undefined>[] = [];\n held.push(walk(branch));\n console.log(String(held.length));")
	programRegionGeneratorStops(t, source, "whose generator is used other than by a for-of over its call or a call as a statement")
}

// A generator held in a local stops there, at the call that would fill the local, before any closure could
// capture it.
func TestProgramRegionGeneratorHeldInLocalStops(t *testing.T) {
	t.Parallel()
	source := programRegionGeneratorTree + programRegionGeneratorRun(" const steps = walk(branch);\n const read = (): Generator<Tree, undefined, undefined> => steps;\n for (const tree of read()) { console.log(tree.name); }")
	programRegionGeneratorStops(t, source, "whose generator is used other than by a for-of over its call or a call as a statement")
}

// The callee side of the one road S2c-1 opened: a generator passed into an Iterable<T> parameter, which
// the callee then puts in a slot, an object's field or a closure's capture, instead of one for-of over it.
// Both stop at the call, at checkGeneratorArguments' first test: the parameter's one use isn't a for-of.
func TestProgramRegionGeneratorParameterStoredInFieldStops(t *testing.T) {
	t.Parallel()
	source := programRegionGeneratorTree +
		"function names(trees: Iterable<Tree>): void { const holder = { t: trees }; for (const tree of holder.t) { console.log(tree.name); } }\n" +
		programRegionGeneratorRun(" names(walk(branch));")
	programRegionGeneratorStops(t, source, "passed into the parameter trees, whose one use isn't a for-of over it")
}

func TestProgramRegionGeneratorParameterCapturedByClosureStops(t *testing.T) {
	t.Parallel()
	source := programRegionGeneratorTree +
		"function names(trees: Iterable<Tree>): void { const read = (): Iterable<Tree> => trees; for (const tree of read()) { console.log(tree.name); } }\n" +
		programRegionGeneratorRun(" names(walk(branch));")
	programRegionGeneratorStops(t, source, "passed into the parameter trees, whose one use isn't a for-of over it")
}
