package lower

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// profilerDeclaration is performance.ts's line 15 in tsc, the onProfilerEvent row's declaration
// (internal/load/host_bindings.json), Absent on every target.
const profilerDeclaration = "declare let onProfilerEvent: ((eventName: string) => void) | undefined;\n"

// profilerDeadRead is the panic hostBindingDeadRead puts where onProfilerEvent's guard proves a read dead.
const profilerDeadRead = "compiler bug: the typeof guard of onProfilerEvent, a host binding absent on every target, was taken"

// profilerModules is a program whose performance.ts declares onProfilerEvent on line 1 and holds body
// in run(name), from line 3, which main.ts calls.
func profilerModules(body string) map[string]string {
	return map[string]string{
		"performance.ts": profilerDeclaration + "export function run(name: string): void {\n" + body + "}\n",
		"main.ts":        "import { run } from './performance.ts';\nrun('one');\nconsole.log('ran');\n",
	}
}

// An Absent host binding lowers by its row (host_bindings.go): typeof is the constant "undefined", and a
// read lowers only in the three dominating shapes, tsc's own if among them, each as the unreachable
// placeholder. The oracle holds what these run to on every backend (internal/oracle/host_bindings_test.go).
func TestHostBindingAbsentLowersUnderItsGuard(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, body string }{
		{"tsc's if", "\tif (typeof onProfilerEvent === \"function\") {\n\t\tonProfilerEvent(name);\n\t}\n"},
		{"an if without a block, testing !== \"undefined\"", "\tif (typeof onProfilerEvent !== \"undefined\") onProfilerEvent(name);\n"},
		{"an if inside the guarded branch", "\tif (typeof onProfilerEvent === \"function\") {\n\t\tif (name.length > 0) {\n\t\t\tonProfilerEvent(name);\n\t\t}\n\t}\n"},
		{"&& as a statement", "\ttypeof onProfilerEvent === \"function\" && onProfilerEvent(name);\n"},
		{"&& as a boolean", "\tconst long = typeof onProfilerEvent === \"function\" && onProfilerEvent.length > 0;\n\tconsole.log(String(long));\n"},
		{"?:", "\tconst count = typeof onProfilerEvent !== \"undefined\" ? onProfilerEvent.length : -1;\n\tconsole.log(String(count));\n"},
		{"parentheses around the guard", "\tif ((typeof (onProfilerEvent) === \"function\")) {\n\t\tonProfilerEvent(name);\n\t}\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerModules(t, profilerModules(probe.body))
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if !slices.Contains(program.Strings, profilerDeadRead) {
				t.Errorf("no unreachable placeholder for the guarded read among the program's strings %q", program.Strings)
			}
		})
	}
	t.Run("typeof alone", func(t *testing.T) {
		t.Parallel()
		program, err := lowerModules(t, profilerModules("\tconsole.log(typeof onProfilerEvent + name);\n"))
		if err != nil {
			t.Fatalf("want it to lower, got %v", err)
		}
		if !slices.Contains(program.Strings, "undefined") {
			t.Errorf("typeof isn't the constant \"undefined\": strings %q", program.Strings)
		}
	})
}

// Every shape the ruling leaves out stops as not yet at the read, never folded: ||, an early return, a
// negated guard, a guard bound to a local, a guard in a called function, and the else branch of a test
// for "undefined". Mutant (internal/oracle/testdata/host_bindings/mutants.md, (a)): hostBindingDominated
// admits a read with no guard, and these lower, or stop elsewhere, instead.
func TestHostBindingReadsOutsideTheirGuardAreNotYet(t *testing.T) {
	t.Parallel()
	const stop = "stage 0 can't lower a read of onProfilerEvent, a host binding absent on every target, outside the then-branch of if, the right of && or the true arm of ?: under a typeof guard that proves it dead, yet"
	for _, probe := range []struct{ name, body, where string }{
		{"||", "\ttypeof onProfilerEvent !== \"function\" || onProfilerEvent(name);\n", "performance.ts:3:43"},
		{"an early return", "\tif (typeof onProfilerEvent !== \"function\") {\n\t\treturn;\n\t}\n\tonProfilerEvent(name);\n", "performance.ts:6:2"},
		{"a negated guard", "\tif (!(typeof onProfilerEvent === \"undefined\")) {\n\t\tonProfilerEvent(name);\n\t}\n", "performance.ts:4:3"},
		{"a guard bound to a local", "\tconst present = typeof onProfilerEvent === \"function\";\n\tif (present) {\n\t\tonProfilerEvent(name);\n\t}\n", "performance.ts:5:3"},
		{"a guard in a called function", "\tif (present()) {\n\t\tonProfilerEvent?.(name);\n\t}\n}\nfunction present(): boolean {\n\treturn typeof onProfilerEvent === \"function\";\n", "performance.ts:4:3"},
		{"the else branch of a test for undefined", "\tif (typeof onProfilerEvent === \"undefined\") {\n\t\treturn;\n\t} else {\n\t\tonProfilerEvent(name);\n\t}\n", "performance.ts:6:3"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, profilerModules(probe.body))
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), probe.where+": "+stop) {
				t.Errorf("got %v, want %s: %s", err, probe.where, stop)
			}
		})
	}
}

// A guard is matched to the table's declaration by symbol, never by name: a local that shadows a host
// binding's name is judged as the local (Language's condition 2 on #00qn5az). Mutant (c): hostBindingOf
// takes the row by name, and the parameter's guarded call becomes the placeholder, while the local
// performance stops as the pending row.
func TestHostBindingShadowingLocalIsTheLocal(t *testing.T) {
	t.Parallel()
	t.Run("a parameter named onProfilerEvent beside the declaration", func(t *testing.T) {
		t.Parallel()
		program, err := lowerModules(t, map[string]string{
			"performance.ts": profilerDeclaration + "export function run(onProfilerEvent: ((eventName: string) => void) | undefined, name: string): void {\n\tif (typeof onProfilerEvent === \"function\") {\n\t\tonProfilerEvent(name);\n\t}\n}\n",
			"main.ts":        "import { run } from './performance.ts';\nrun((eventName: string): void => {\n\tconsole.log(eventName);\n}, 'fired');\n",
		})
		if err != nil {
			t.Fatalf("want it to lower, got %v", err)
		}
		if slices.Contains(program.Strings, profilerDeadRead) {
			t.Error("the parameter's guarded call lowered as the host binding's unreachable placeholder")
		}
	})
	t.Run("a local named performance", func(t *testing.T) {
		t.Parallel()
		if _, err := lowerModules(t, map[string]string{
			"main.ts": "const performance = { now(): number { return 1; } };\nconsole.log(String(typeof performance === \"object\" && performance.now() > 0));\n",
		}); err != nil {
			t.Fatalf("want it to lower as the local, got %v", err)
		}
	})
}

// Nothing in the program may define a host binding's name (Language's condition 1 on #00qn5az): a write
// to a `declare let` is refused, and so is a write through the global object whose key names a row, by
// property or by element, or whose key could name one. The global object is judged by its type, typeof
// globalThis, through any name: globalThis, a parameter of that type, globalThis.globalThis and Node's
// global (an alias of it stops before as the global object as a value, global_object.go). Mutant
// (e), internal/oracle/testdata/host_bindings/mutants.md: the receiver judged by its spelling again.
func TestHostBindingWritesAreRefused(t *testing.T) {
	t.Parallel()
	const fix = "a host binding is the host's to define, and nothing in the program may; keep the value in a variable of your own"
	for _, probe := range []struct {
		name    string
		modules map[string]string
		want    string
	}{
		{"a write to the declare let", profilerModules("\tonProfilerEvent = (eventName: string): void => {\n\t\tconsole.log(eventName + name);\n\t};\n"), "performance.ts:3:2: Adamic 0.1 refuses a write to the host binding onProfilerEvent; " + fix},
		{"globalThis.process", map[string]string{"main.ts": "import process from 'node:process';\nglobalThis.process = process;\n"}, "main.ts:2:1: Adamic 0.1 refuses a write through globalThis to the host binding process; " + fix},
		{"globalThis['performance']", map[string]string{"main.ts": "import { performance } from 'node:perf_hooks';\nglobalThis['performance'] = performance;\n"}, "main.ts:2:1: Adamic 0.1 refuses a write through globalThis to the host binding performance; " + fix},
		{"a parameter of the global object's type", map[string]string{"main.ts": "import process from 'node:process';\nfunction put(host: typeof globalThis): void {\n\thost.process = process;\n}\nput(globalThis);\n"}, "main.ts:3:2: Adamic 0.1 refuses a write through globalThis to the host binding process; " + fix},
		{"globalThis.globalThis", map[string]string{"main.ts": "import process from 'node:process';\nglobalThis.globalThis.process = process;\n"}, "main.ts:2:1: Adamic 0.1 refuses a write through globalThis to the host binding process; " + fix},
		{"Node's global", map[string]string{"main.ts": "import process from 'node:process';\nglobal.process = process;\n"}, "main.ts:2:1: Adamic 0.1 refuses a write through globalThis to the host binding process; " + fix},
		{"globalThis[key]", map[string]string{"main.ts": "import process from 'node:process';\nconst key: 'process' | 'performance' = process.argv.length > 9 ? 'process' : 'performance';\ndelete globalThis[key];\n"}, "main.ts:3:8: Adamic 0.1 refuses a write through globalThis whose key may name a host binding; " + fix},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, probe.modules)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), probe.want) {
				t.Errorf("got %v, want %s", err, probe.want)
			}
		})
	}
	// globalThis.onProfilerEvent can't be written in a program the checker passes: every Adamic file is a
	// module (load.go's ModuleDetection), so a declare let is the module's, never a property of globalThis.
	directory := t.TempDir()
	main := filepath.Join(directory, "performance.ts")
	if err := os.WriteFile(main, []byte(profilerDeclaration+"globalThis.onProfilerEvent = (eventName: string): void => {\n\tconsole.log(eventName);\n};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := load.Load([]string{main}); err == nil || !strings.Contains(err.Error(), "performance.ts:2:12: error TS7017: Element implicitly has an 'any' type because type 'typeof globalThis' has no index signature.") {
		t.Errorf("Load: got %v, want the checker's error that globalThis has no onProfilerEvent", err)
	}
}

// A declaration the table doesn't name, exactly, stays refused as an ambient value declaration: another
// type, another file, another keyword, an exported one, and one beside a declarator the table names.
func TestHostBindingUnmatchedDeclarationIsRefused(t *testing.T) {
	t.Parallel()
	const refusal = "Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values"
	for _, probe := range []struct {
		name, file, source string
	}{
		{"another type", "core.ts", "declare const process: number;\nconsole.log(typeof process);\n"},
		{"another file", "main.ts", profilerDeclaration + "console.log(typeof onProfilerEvent);\n"},
		{"another keyword", "performance.ts", "declare var onProfilerEvent: ((eventName: string) => void) | undefined;\nconsole.log(typeof onProfilerEvent);\n"},
		{"exported", "performance.ts", "export " + profilerDeclaration + "console.log(typeof onProfilerEvent);\n"},
		{"beside an unnamed declarator", "performance.ts", "declare let onProfilerEvent: ((eventName: string) => void) | undefined, other: number;\nconsole.log(typeof onProfilerEvent);\n"},
		{"tsc's require", "core.ts", "declare const require: unknown;\nconsole.log(typeof require);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			modules := map[string]string{probe.file: probe.source}
			if probe.file != "main.ts" {
				modules["main.ts"] = "import './" + probe.file + "';\n"
			}
			_, err := lowerModules(t, modules)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), probe.file+":1:1: "+refusal) {
				t.Errorf("got %v, want %s:1:1: %s", err, probe.file, refusal)
			}
		})
	}
}

// process and performance are Provided on Node, but their native column waits on Runtime's clearance,
// their wasm32 column on Runtime's and Platforms', and performance's reads on #3sh9398's host model, so
// each stops as not yet at its declaration, naming what it waits on, never Node's answer printed where a
// host has nothing.
func TestHostBindingPendingRowsAreNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, file, source, want string }{
		{"process", "core.ts", "declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;\nexport function isNodeLikeSystem(): boolean {\n\treturn typeof process !== \"undefined\" && !!process.nextTick && !process.browser;\n}\n",
			"core.ts:1:15: stage 0 can't lower the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, yet"},
		{"performance", "performanceCore.ts", "interface Performance {\n\tnow(): number;\n}\ndeclare const performance: Performance | undefined;\nexport function now(): number {\n\treturn typeof performance === \"object\" ? performance.now() : 0;\n}\n",
			"performanceCore.ts:4:15: stage 0 can't lower the host binding performance, pending Runtime's clearance of its native column, Runtime and Platforms' clearance of its wasm32 column and #3sh9398's host model, yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, map[string]string{probe.file: probe.source, "main.ts": "import './" + probe.file + "';\n"})
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), probe.want) {
				t.Errorf("got %v, want %s", err, probe.want)
			}
		})
	}
}
