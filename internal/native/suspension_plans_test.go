package native

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The three ownership plans run on every program, and only a suspending function's body is left out
// of them (#kha2ea0): one setTimeout or Promise used to empty them for the whole program, so graph.ts
// with its timer read 0 regions and 1.9 times Node's user time, where its twin without the timer read
// 10,485,750 regions and 0.83 times.

// suspensionProgram lowers source as main.a.
func suspensionProgram(t *testing.T, source string) *ir.Program {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
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
	if err := Build(C(program), binary, Options{Sanitize: true}); err != nil {
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
// point. weigh doesn't suspend, and Main's statement handing it make(5) keeps its region in a program
// that holds promises. Mutant: the Suspends() clause dropped from the escape seed, and countLater's and
// relay's p flow nowhere, start's and startRelay's statements take regions, and ASan reports the read.
func TestSuspendingCalleeParametersEscape(t *testing.T) {
	t.Parallel()
	program := suspensionProgram(t, `interface P {
    readonly x: number;
}
function make(n: number): P {
    return { x: n };
}
function weigh(p: P): number {
    return p.x;
}
async function countLater(p: P): Promise<number> {
    await Promise.resolve(0);
    console.log("later " + String(p.x));
    return p.x;
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
let weight = 0;
weight += weigh(make(5));
const first = start();
const second = startRelay();
console.log("sync " + String(weight));
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
	main := statementsOf(program.Main)
	regions := 0
	for statement := range plan.statements {
		if main[statement] {
			regions++
		}
	}
	if regions == 0 {
		t.Error("Main's weigh(make(5)) took no region in a program holding promises")
	}
	suspensionRun(t, program, "sync 5\nlater 7\nlater 9\n")
}

// A suspending body keeps its locals in frame cells, which reuse's entries and a borrowed element's
// binding don't name, so the plans leave its body out (#kha2ea0): work, an async function, and steps,
// a generator, each hold a statement that would take a region, a spread that would be reused and an
// element that would be borrowed. grow and Main don't suspend and keep theirs: a spread reused in grow,
// and a region in Main. Mutant: the Suspends() skips dropped from planRegions' statements, planReuse
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
let seen = 0;
for (const v of steps(boxes)) {
    seen += v;
}
let weight = 0;
weight += weigh(make(6));
console.log("seen " + String(seen) + " grown " + String(grow(make(1)).n) + " weight " + String(weight));
const done = report(boxes);
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
	main := statementsOf(program.Main)
	region := false
	for statement := range regions.statements {
		region = region || main[statement]
	}
	if !region {
		t.Error("Main's weigh(make(6)) took no region in a program holding promises and a generator")
	}
	suspensionRun(t, program, "seen 17 grown 3 weight 6\ntotal 26\n")
}
