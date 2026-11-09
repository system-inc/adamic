package lower

import (
	"context"
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
// performance is judged by its model.
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
// global, and through a union or intersection holding it, a type parameter constrained to it, or an arm
// of ?: that is it even where the checker reduced the union away (other's type a supertype of it). An
// alias of it stops before, as the global object as a value (global_object.go). Mutants (e), (f) and
// (g), internal/oracle/testdata/host_bindings/mutants.md: the receiver judged by its spelling again, by
// its type's own symbol alone, and without its operands.
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
		{"a union with the global object", map[string]string{"main.ts": "import process from 'node:process';\nconst other: { process: typeof process; tag: string } = { process, tag: 'other' };\n(process.argv.length > 9 ? globalThis : other).process = process;\n"}, "main.ts:3:1: Adamic 0.1 refuses a write through globalThis to the host binding process; " + fix},
		{"a union the checker reduces to its other member", map[string]string{"main.ts": "import process from 'node:process';\nconst other: { process: typeof process } = { process };\n(process.argv.length > 9 ? globalThis : other).process = other.process;\n"}, "main.ts:3:1: Adamic 0.1 refuses a write through globalThis to the host binding process; " + fix},
		{"an intersection with the global object", map[string]string{"main.ts": "const fire = (eventName: string): void => {\n\tconsole.log(eventName);\n};\n(globalThis as typeof globalThis & { onProfilerEvent?: (eventName: string) => void }).onProfilerEvent = fire;\n"}, "main.ts:4:1: Adamic 0.1 refuses a write through globalThis to the host binding onProfilerEvent; " + fix},
		{"a parameter of a union with the global object", map[string]string{"main.ts": "import process from 'node:process';\nfunction put(host: typeof globalThis | { process: typeof process; tag: string }): void {\n\thost.process = process;\n}\nput(globalThis);\n"}, "main.ts:3:2: Adamic 0.1 refuses a write through globalThis to the host binding process; " + fix},
		{"a parameter of an intersection with the global object", map[string]string{"main.ts": "const fire = (eventName: string): void => {\n\tconsole.log(eventName);\n};\nfunction put(host: typeof globalThis & { onProfilerEvent?: (eventName: string) => void }): void {\n\thost.onProfilerEvent = fire;\n}\nput(globalThis);\n"}, "main.ts:5:2: Adamic 0.1 refuses a write through globalThis to the host binding onProfilerEvent; " + fix},
		{"a type parameter constrained to the global object", map[string]string{"main.ts": "import process from 'node:process';\nfunction put<G extends typeof globalThis>(host: G): void {\n\thost.process = process;\n}\nput(globalThis);\n"}, "main.ts:3:2: Adamic 0.1 refuses a write through globalThis to the host binding process; " + fix},
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
		{"require of another type than core.ts's", "core.ts", "declare const require: (specifier: string) => unknown;\nconsole.log(typeof require);\n"},
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

// tsc's core.ts declares require beside process and reads it only in isNodeLikeSystem's typeof test. Its
// row is Absent on every target, Node's answer under the oracle's module mode (Adamic on #baobab, Nov 8),
// so the declaration declares nothing, typeof is the constant "undefined", isNodeLikeSystem's test is
// false at run time, as it is on the oracle's Node, and a read under the guard is the unreachable
// placeholder. The oracle holds the run (TestHostBindingsHeldToNode
// and TestHostBindingsAgreeWithNode's node_like). Mutant (k), internal/oracle/testdata/host_bindings/
// mutants.md: the row removed, and the declaration is refused again as an ambient value declaration.
func TestHostBindingRequireIsAbsent(t *testing.T) {
	t.Parallel()
	program, err := lowerModules(t, map[string]string{
		"core.ts": "declare const require: unknown;\nexport function isNodeLikeSystem(): boolean {\n\treturn typeof require !== \"undefined\";\n}\nexport function loader(): string {\n\tif (typeof require !== \"undefined\") {\n\t\treturn String(require);\n\t}\n\treturn \"no loader\";\n}\n",
		"main.ts": "import { isNodeLikeSystem, loader } from './core.ts';\nconsole.log(String(isNodeLikeSystem()));\nconsole.log(loader());\n",
	})
	if err != nil {
		t.Fatalf("want core.ts's require to lower by its row, got %v", err)
	}
	const deadRead = "compiler bug: the typeof guard of require, a host binding absent on every target, was taken"
	if !slices.Contains(program.Strings, deadRead) {
		t.Errorf("no unreachable placeholder for require's guarded read among the program's strings %q", program.Strings)
	}
}

// A row's Node answer is measured in ES module mode, so a build that declares one CommonJS bundle, where
// Node answers typeof require "function", stops at core.ts's require rather than folding it Absent, at
// its declaration, and so does a row of any name, Absent or routed (WebAssembly's must on #00qn5az, #redxes2). The
// same program lowers in the default ES build. Mutant (l), internal/oracle/testdata/host_bindings/
// mutants.md: hostBindingStopIn's format test dropped, and the bundle build lowers require as Absent.
func TestHostBindingRowsStopInACommonJSBundle(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, file, source, want string }{
		{"core.ts's require", "core.ts", "declare const require: unknown;\nexport function isNodeLikeSystem(): boolean {\n\treturn typeof require !== \"undefined\";\n}\n",
			"core.ts:1:15: stage 0 can't lower the host binding require in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), yet"},
		{"performance.ts's onProfilerEvent", "performance.ts", profilerDeclaration + "export function run(name: string): void {\n\tif (typeof onProfilerEvent === \"function\") {\n\t\tonProfilerEvent(name);\n\t}\n}\n",
			"performance.ts:1:13: stage 0 can't lower the host binding onProfilerEvent in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), yet"},
		// A routed row too: process is Provided and routes its reads through node:process in the ES build, and a
		// bundle build stops at its declaration before any read is routed (Compiler's read of #00qn5az's seam with
		// #yewmv46).
		{"core.ts's process, a routed row", "core.ts", "declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;\nconsole.log(typeof process);\n",
			"core.ts:1:15: stage 0 can't lower the host binding process in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2), yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			main := "import './" + probe.file + "';\n"
			if strings.Contains(probe.name, "process") {
				// The routed row's model export is node:process's, which the program has to load (processModules).
				main = "import 'node:process';\n" + main
			}
			for name, source := range map[string]string{probe.file: probe.source, "main.ts": main} {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if _, err := LowerWithOptions(context.Background(), program, Options{}); err != nil {
				t.Fatalf("the ES build: want it to lower, got %v", err)
			}
			_, err = LowerWithOptions(context.Background(), program, Options{ModuleFormat: ModuleFormatCommonJSBundle})
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), probe.want) {
				t.Errorf("the bundle build: got %v, want %s", err, probe.want)
			}
		})
	}
}
