package native_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	for name, source := range restPackPaths {
		t.Run(name, func(t *testing.T) {
			program := graphCompilerWitness(t, source, nil)
			if len(program.RestGraphTypes) == 0 || len(program.GraphTypes) == 0 {
				t.Fatalf("%s: no rest pack plan (rest %v, graph %v)", name, program.RestGraphTypes, program.GraphTypes)
			}
		})
	}
}

// The plan is what frees each region: without it, every path's pack is counted again and
// LeakSanitizer reports the region.
func TestRestPackWithoutItsPlanLeaks(t *testing.T) {
	for name, source := range restPackPaths {
		t.Run(name, func(t *testing.T) {
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
			if len(program.RestGraphTypes) == 0 {
				t.Fatalf("%s: no rest pack plan to remove", name)
			}
			program.RestGraphTypes = nil
			code := native.C(program)
			binary := filepath.Join(directory, "native")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			// Linux reports any failing exit, so the leak must be LeakSanitizer's, never a crash;
			// macOS's counted build says "heap values leaked".
			report := leakcheck.Report(t, code, binary)
			if !strings.Contains(report, "LeakSanitizer") && !strings.Contains(report, "heap values leaked") {
				t.Fatalf("%s: the pack without its plan didn't leak: %q", name, report)
			}
			for _, line := range strings.Split(report, "\n") {
				if strings.Contains(line, "SUMMARY") || strings.Contains(line, "heap values leaked") {
					t.Logf("%s without its plan: %s", name, line)
				}
			}
		})
	}
}

// or's fs cell seeds a graph region by type alone (cycles.go), so an or that closes no cycle is a
// graph program too, and its packs are planned in. An or of an or, packed into a region and
// freed with it, still prints what Node prints and leaks nothing.
func TestRestPackOffACycleLeaksNothing(t *testing.T) {
	graphCompilerWitness(t, restPackOr+`function isBig(n: number): boolean {
	return n > 10;
}
const inner = or(isEven, isBig);
const outer = or(inner, isEven);
console.log(`+"`${outer(3)} ${outer(4)} ${outer(11)}`"+`);
`, nil)
}
