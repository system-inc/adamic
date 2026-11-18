package native

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// The three ownership plans run on every program, and only a suspending function's body is left out
// of them (#kha2ea0): one setTimeout or Promise used to empty them for the whole program, so graph.ts
// with its timer read 0 regions and 1.9 times Node's user time, where its twin without the timer read
// 10,485,750 regions and 0.83 times.

// suspensionProgram lowers source as main.a.
func suspensionProgram(t *testing.T, source string) *ir.Program {
	t.Helper()
	return suspensionProgramAs(t, source, "main.a")
}

// suspensionProgramAs lowers source as the file named.
func suspensionProgramAs(t *testing.T, source, name string) *ir.Program {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

// suspensionRun builds program sanitized and runs it, failing unless it prints want and exits 0
// with nothing on stderr: ASan's report of a read of a freed region is that stderr.
func suspensionRun(t *testing.T, program *ir.Program, want string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "native")
	if err := Build(C(program), binary, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	var out, errs bytes.Buffer
	command.Stdout, command.Stderr = &out, &errs
	err := command.Run()
	if err != nil || out.String() != want || errs.Len() != 0 {
		t.Fatalf("native: %v, stdout %q, stderr %q; want exit 0 and %q, as Node prints", err, out.String(), errs.String(), want)
	}
}

// statementsOf is every statement in body, nested ones too, by the address the plans key them by.
func statementsOf(body []ir.Statement) map[*ir.Statement]bool {
	set := map[*ir.Statement]bool{}
	var visit func([]ir.Statement)
	visit = func(list []ir.Statement) {
		for index := range list {
			set[&list[index]] = true
			walkStatement(list[index], func(ir.Expression) {}, visit)
		}
	}
	visit(body)
	return set
}

func functionNamed(t *testing.T, program *ir.Program, name string) int {
	t.Helper()
	for index, function := range program.Functions {
		if function.Name == name {
			return index
		}
	}
	t.Fatalf("no function %s", name)
	return -1
}

// A region made in a statement and handed to a suspending function's parameter would end with the
// statement, while the function reads it after it resumes (Memory's rule 1). countLater reads p after
// its await, and relay hands its own p on to countLater, so its p escapes only through the fixed
// point. weigh doesn't suspend, and measure's statement handing it make(5) keeps its region in a program
// that holds promises. The top-level awaits make Main a suspending module function, so the controls
// live in functions that don't suspend. Mutant: the Suspends() clause dropped from the escape seed, and countLater's and
// relay's p flow nowhere, start's and startRelay's statements take regions, and ASan reports the read.
func TestSuspendingCalleeParametersEscape(t *testing.T) {
	t.Parallel()
	program := suspensionProgram(t, `interface P {
    readonly x: number;
}
function make(n: number): P {
    return { x: n };
}
// Each field is read into arithmetic held in a local, as graph.ts's sum reads its node: a field
// returned or handed on alone counts as the parameter escaping today (region.go's escaping, #1jch5ez),
// which would hide the region this pins.
function weigh(p: P): number {
    const weight = 1 + p.x;
    return weight;
}
async function countLater(p: P): Promise<number> {
    await Promise.resolve(0);
    const seen = 1 + p.x;
    console.log("later " + String(seen));
    return seen;
}
function relay(p: P): Promise<number> {
    return countLater(p);
}
function start(): Promise<number> {
    const pending = countLater(make(7));
    return pending;
}
function startRelay(): Promise<number> {
    const pending = relay(make(9));
    return pending;
}
function measure(): number {
    let weight = 0;
    weight += weigh(make(5));
    return weight;
}
console.log("sync " + String(measure()));
await start();
await startRelay();
`)
	plan := planRegions(program)
	for _, name := range []string{"countLater", "relay"} {
		if !plan.escapes[functionNamed(t, program, name)][0] {
			t.Errorf("%s's p flows nowhere, though a suspending body reads it after the call returns", name)
		}
	}
	for _, name := range []string{"start", "startRelay"} {
		for statement := range statementsOf(program.Functions[functionNamed(t, program, name)].Body) {
			if plan.statements[statement] {
				t.Errorf("%s's statement takes a region handed to a suspending parameter", name)
			}
		}
	}
	measured := statementsOf(program.Functions[functionNamed(t, program, "measure")].Body)
	regions := 0
	for statement := range plan.statements {
		if measured[statement] {
			regions++
		}
	}
	if regions == 0 {
		t.Error("measure's weigh(make(5)) took no region in a program holding promises")
	}
	suspensionRun(t, program, "sync 6\nlater 8\nlater 10\n")
}

// A suspending body keeps its locals in frame cells, which reuse's entries and a borrowed element's
// binding don't name, so the plans leave its body out (#kha2ea0): work, an async function, and steps,
// a generator, each hold a statement that would take a region, a spread that would be reused and an
// element that would be borrowed. grow and summary don't suspend and keep theirs: a spread reused in
// grow, and a region in summary (its own maker, since make's literals also go into the global boxes,
// which keeps them out of a region), which also consumes the generator, since a consumer can't sit in the
// suspending module function the top-level await makes Main. Mutant: the Suspends() skips dropped from planRegions' statements, planReuse
// and planElementBorrows, and the plans reach into work and steps.
func TestSuspendingBodiesPlanNothing(t *testing.T) {
	t.Parallel()
	program := suspensionProgram(t, `interface Box {
    readonly n: number;
    readonly tag: string;
}
function make(n: number): Box {
    return { n: n, tag: "t" + String(n) };
}
function weigh(b: Box): number {
    return b.n;
}
interface P {
    readonly x: number;
}
function makeP(n: number): P {
    return { x: n };
}
function weighP(p: P): number {
    const weight = 1 + p.x;
    return weight;
}
function grow(b: Box): Box {
    let c = { ...b, n: b.n + 1 };
    c = { ...c, n: c.n + 1 };
    return c;
}
async function work(xs: Box[]): Promise<number> {
    let total = 0;
    total += weigh(make(1));
    let b = make(2);
    await Promise.resolve(0);
    b = { ...b, n: b.n + 10 };
    for (const x of xs) {
        total += x.n;
    }
    const first = xs[0];
    await Promise.resolve(0);
    total += b.n + (first === undefined ? 0 : first.n);
    return total;
}
function* steps(xs: Box[]) {
    let b = make(3);
    yield 0;
    b = { ...b, n: b.n + 1 };
    for (const x of xs) {
        yield x.n + b.n;
    }
}
async function report(xs: Box[]): Promise<void> {
    const total = await work(xs);
    console.log("total " + String(total));
}
const boxes: Box[] = [make(4), make(5)];
function summary(): string {
    let seen = 0;
    for (const v of steps(boxes)) {
        seen += v;
    }
    let weight = 0;
    weight += weighP(makeP(6));
    return "seen " + String(seen) + " grown " + String(grow(make(1)).n) + " weight " + String(weight);
}
console.log(summary());
await report(boxes);
`)
	borrows, lending := planElementBorrows(program)
	reuse := planReuse(program, lending)
	regions := planRegions(program)
	for _, name := range []string{"work", "steps", "report"} {
		index := functionNamed(t, program, name)
		if !program.Functions[index].Suspends() {
			t.Fatalf("%s doesn't suspend", name)
		}
		for statement := range statementsOf(program.Functions[index].Body) {
			if regions.statements[statement] || borrows[statement] || reuse.spreads[statement] != nil || reuse.arrays[statement] != nil || reuse.moves[statement] != nil {
				t.Errorf("a plan reaches into %s's body", name)
			}
		}
	}
	grown := statementsOf(program.Functions[functionNamed(t, program, "grow")].Body)
	reused := false
	for statement := range reuse.spreads {
		reused = reused || grown[statement]
	}
	if !reused {
		t.Error("grow's spread isn't reused in a program holding promises and a generator")
	}
	summarized := statementsOf(program.Functions[functionNamed(t, program, "summary")].Body)
	region := false
	for statement := range regions.statements {
		region = region || summarized[statement]
	}
	if !region {
		t.Error("summary's weighP(makeP(6)) took no region in a program holding promises and a generator")
	}
	suspensionRun(t, program, "seen 17 grown 3 weight 7\ntotal 26\n")
}

// Runtime Performance's graph.ts (#ahnc64w), made small: a graph built fresh and summed in one
// statement, ten rounds of it in the original, and one trailing setTimeout. The timer used to cost the
// program every region; graph's statement now takes its region, and the emitted C opens and ends one.
// Mutant: the program-wide gate put back in cProgram, and the C holds no region.
func TestTimerProgramKeepsItsRegions(t *testing.T) {
	t.Parallel()
	program := suspensionProgramAs(t, `import 'node:timers';
interface Node {
	readonly value: number;
	readonly next: Node | undefined;
	readonly back: Node | undefined;
}
function build(depth: number, value: number): Node {
	if (depth === 0) {
		return { value, next: undefined, back: undefined };
	}
	return { value, next: build(depth - 1, value + 1), back: build(depth - 1, value + 2) };
}
function sum(node: Node): number {
	let total = 1 + node.value;
	if (node.next !== undefined) {
		total += sum(node.next);
	}
	if (node.back !== undefined) {
		total += sum(node.back);
	}
	return total;
}
function graph(seed: number): number {
	const total = sum(build(3, seed));
	return total;
}
let checksum = 0;
for (let round = 0; round < 3; round += 1) {
	checksum += graph(round);
}
console.log(`+"`graph checksum ${checksum}`"+`);
setTimeout(() => {
	console.log('quiet');
}, 1);
`, "main.ts")
	if !program.HasPromises() {
		t.Fatal("the timer no longer makes HasPromises true, so this pin no longer holds the gate")
	}
	graph := statementsOf(program.Functions[functionNamed(t, program, "graph")].Body)
	planned := false
	for statement := range planRegions(program).statements {
		planned = planned || graph[statement]
	}
	if !planned {
		t.Error("graph's statement took no region in a program holding a timer")
	}
	if !strings.Contains(C(program), "adamic_region_end(") {
		t.Error("the emitted C ends no region in a program holding a timer")
	}
	suspensionRun(t, program, "graph checksum 243\nquiet\n")
}

// A fresh function emitted just before an async one: emitting the fresh one's region version left the
// emitter in it, and the async function after it was named and signed as a region version, so nothing
// defined the one its callers call (async_methods.a's split link, where Worker.allocate precedes
// Worker.work). Every program that holds promises could reach it once the plans run there.
// Mutant: cProgram's reset before a suspending function dropped, and clang finds later never defined.
func TestAsyncAfterAFreshFunction(t *testing.T) {
	t.Parallel()
	program := suspensionProgram(t, `interface P {
    readonly x: number;
}
function make(n: number): P {
    return { x: n };
}
async function later(n: number): Promise<number> {
    await Promise.resolve(0);
    return n + 1;
}
function weigh(p: P): number {
    const weight = 1 + p.x;
    return weight;
}
function measure(): number {
    let weight = 0;
    weight += weigh(make(2));
    return weight;
}
console.log(String(measure()));
console.log(String(await later(1)));
`)
	if !planRegions(program).fresh[functionNamed(t, program, "make")] {
		t.Fatal("make isn't fresh, so this pin no longer puts a region version before an async function")
	}
	if make, later := functionNamed(t, program, "make"), functionNamed(t, program, "later"); later != make+1 {
		t.Fatalf("later (%d) doesn't follow make (%d), so this pin no longer emits it right after a region version", later, make)
	}
	suspensionRun(t, program, "3\n2\n")
}

// A fresh function's region version can't hand one of its allocations to a suspending callee today
// (Memory's road on #kha2ea0): the call's promise would have to be awaited where it's made, and a fresh
// function returns an object literal, so it can neither await nor return it. Lowering stops it there,
// before any plan. If this road ever lowers, the pin goes red, and the fresh function's escape has to be
// pinned under ASan with the Suspends() seed dropped as its mutant.
func TestFreshFunctionCannotHandItsAllocationToASuspendingCallee(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(`interface P {
    readonly x: number;
}
interface Box {
    readonly y: number;
}
function make(n: number): P {
    return { x: n };
}
async function countLater(p: P): Promise<number> {
    await Promise.resolve(0);
    const seen = 1 + p.x;
    return seen;
}
function holder(n: number): Box {
    const pending = countLater(make(n));
    return { y: n };
}
function weighBox(b: Box): number {
    const weight = 1 + b.y;
    return weight;
}
console.log(String(weighBox(holder(4))));
`), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	if err == nil || !strings.Contains(err.Error(), "a promise not awaited in the same function") {
		t.Fatalf("want lowering's stop for a promise holder neither awaits nor returns, got %v", err)
	}
}
