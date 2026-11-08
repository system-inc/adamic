package native_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// A rest parameter's array is packed by its caller, after lowering, so it takes its graph plan
// from Program.RestGraphTypes. Before #2zhhwaz it took none: a cycle that runs through a pack held
// the graph region from a counted array the region held back, and the whole region leaked while
// the program printed what Node prints. Each program below closes a cycle that runs through a pack,
// and between them they pack by every path: a direct call (direct, push, kept), a spread call
// forwarded through a second rest function, a closure called as a value, and a named function
// called through its value adapter.

const restPackOr = `type Test = (n: number) => boolean | undefined;
function or(...fs: Test[]): Test {
	return (n: number): boolean | undefined => {
		let lastResult: boolean | undefined;
		for (const f of fs) {
			lastResult = f(n);
			if (lastResult) {
				return lastResult;
			}
		}
		return lastResult;
	};
}
function isEven(n: number): boolean {
	return n % 2 === 0;
}
`

var restPackPaths = map[string]string{
	// An element of fs captures slot, and slot is then given or's result, so the cycle runs
	// result, fs's cell, the pack, watcher, slot's cell, and back.
	"direct": restPackOr + `function run(): void {
	let slot: Test | undefined;
	const watcher = (n: number): boolean => slot !== undefined && n > 100;
	const g = or(watcher, isEven);
	slot = g;
	console.log(` + "`${g(3)} ${g(4)} ${g(101)}`" + `);
}
run();
`,
	// The pack the cycle runs through is the second one: either's spread repacks fs for or.
	"spread": restPackOr + `function either(...fs: Test[]): Test {
	return or(...fs);
}
function run(): void {
	let slot: Test | undefined;
	const watcher = (n: number): boolean => slot !== undefined && n > 100;
	const g = either(watcher, isEven);
	slot = g;
	console.log(` + "`${g(3)} ${g(4)} ${g(101)}`" + `);
}
run();
`,
	// A closure with a rest parameter, called as a value: the caller packs for its layout.
	"closure": `type Test = (n: number) => boolean | undefined;
const or = (...fs: Test[]): Test => {
	return (n: number): boolean | undefined => {
		let lastResult: boolean | undefined;
		for (const f of fs) {
			lastResult = f(n);
			if (lastResult) {
				return lastResult;
			}
		}
		return lastResult;
	};
};
function isEven(n: number): boolean {
	return n % 2 === 0;
}
function run(): void {
	let slot: Test | undefined;
	const watcher = (n: number): boolean => slot !== undefined && n > 100;
	const g = or(watcher, isEven);
	slot = g;
	console.log(` + "`${g(3)} ${g(4)} ${g(101)}`" + `);
}
run();
`,
	// or pushes its own result into fs through an alias.
	"push": `type Test = (n: number) => boolean | undefined;
function or(...fs: Test[]): Test {
	const result = (n: number): boolean | undefined => {
		let lastResult: boolean | undefined;
		for (const f of fs) {
			lastResult = f(n);
			if (lastResult) {
				return lastResult;
			}
		}
		return lastResult;
	};
	const a = fs;
	a.push(result);
	return result;
}
function isEven(n: number): boolean {
	return n % 2 === 0;
}
function isBig(n: number): boolean {
	return n > 10;
}
const either = or(isEven, isBig);
console.log(` + "`${either(4)} ${either(11)}`" + `);
`,
	// A closure over fs is kept in a let that an element of fs captured.
	"kept": `type Test = (n: number) => boolean | undefined;
function or(keep: (count: () => number) => void, ...fs: Test[]): Test {
	const count = (): number => fs.length;
	keep(count);
	return (n: number): boolean | undefined => {
		let lastResult: boolean | undefined;
		for (const f of fs) {
			lastResult = f(n);
			if (lastResult) {
				return lastResult;
			}
		}
		return lastResult;
	};
}
function isEven(n: number): boolean {
	return n % 2 === 0;
}
function run(): void {
	let slot: (() => number) | undefined;
	const watcher = (n: number): boolean => slot !== undefined && n > 100;
	const keep = (count: () => number): void => {
		slot = count;
	};
	const g = or(keep, watcher, isEven);
	console.log(` + "`${g(3)} ${g(4)} ${g(101)} ${slot === undefined ? 0 : slot()}`" + `);
}
run();
`,
	// or used as a value: the call goes through its value adapter, which forwards to or, so the
	// pack's type is read through the forwarding chain.
	"adapter": restPackOr + `function run(make: (...fs: Test[]) => Test): void {
	let slot: Test | undefined;
	const watcher = (n: number): boolean => slot !== undefined && n > 100;
	const g = make(watcher, isEven);
	slot = g;
	console.log(` + "`${g(3)} ${g(4)} ${g(101)}`" + `);
}
run(or);
`,
}

func TestRestPackOnACycleIsAGraphAllocation(t *testing.T) {
	t.Parallel()
	for name, source := range restPackPaths {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			program := graphCompilerWitness(t, source, nil)
			if len(program.RestGraphTypes) == 0 || len(program.GraphTypes) == 0 {
				t.Fatalf("%s: no rest pack plan (rest %v, graph %v)", name, program.RestGraphTypes, program.GraphTypes)
			}
			// The witness's own counted check takes a value still in a region at exit as held, so
			// this asks the stricter question: every allocation freed, no region left.
			if allocations, frees, regions := restPackCounts(t, source, false); allocations != frees || regions != 0 {
				t.Fatalf("%s: the region outlived the program (allocations %d, frees %d, in regions %d)", name, allocations, frees, regions)
			}
		})
	}
}

// The plan is what frees each region: without it, every path's pack is counted again and holds
// its region from outside. LeakSanitizer scans conservatively and misses a region a stale stack
// word still points at, so it saw these leaks on some runs and not others; the counted build sees
// every allocation that isn't freed, so it's the judge here.
func TestRestPackWithoutItsPlanLeaks(t *testing.T) {
	t.Parallel()
	for name, source := range restPackPaths {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			allocations, frees, regions := restPackCounts(t, source, true)
			if allocations == frees && regions == 0 {
				t.Fatalf("%s: the pack without its plan freed everything (allocations %d, frees %d)", name, allocations, frees)
			}
			t.Logf("%s without its plan: allocations %d, frees %d, in regions %d", name, allocations, frees, regions)
		})
	}
}

// restPackCounts lowers source, clears its rest pack plan when asked, builds it counted, runs it,
// and returns the counted build's allocations, frees and values left in regions.
func restPackCounts(t *testing.T, source string, withoutPlan bool) (int64, int64, int64) {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "main.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if withoutPlan {
		if len(program.RestGraphTypes) == 0 {
			t.Fatal("no rest pack plan to remove")
		}
		program.RestGraphTypes = nil
	}
	binary := filepath.Join(directory, "counted")
	if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("counted run: %v\n%s", err, stderr.String())
	}
	match := leakcheck.CountsLine.FindSubmatch(stderr.Bytes())
	if match == nil {
		t.Fatalf("no counts line:\n%s", stderr.String())
	}
	counts := [3]int64{}
	for index, group := range []int{1, 2, 6} {
		count, err := strconv.ParseInt(string(match[group]), 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		counts[index] = count
	}
	return counts[0], counts[1], counts[2]
}

// or's fs cell seeds a graph region by type alone (cycles.go), so an or that closes no cycle is a
// graph program too, and its packs are planned in. An or of an or, packed into a region and
// freed with it, still prints what Node prints and leaks nothing.
func TestRestPackOffACycleLeaksNothing(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, restPackOr+`function isBig(n: number): boolean {
	return n > 10;
}
const inner = or(isEven, isBig);
const outer = or(inner, isEven);
console.log(`+"`${outer(3)} ${outer(4)} ${outer(11)}`"+`);
`, nil)
}
