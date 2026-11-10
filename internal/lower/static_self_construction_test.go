package lower

import (
	"os"
	"strings"
	"testing"
)

// A class constructing itself in its own static initializer is judged by what the construction
// reads (#1451npc): its parameter defaults, instance field initializers, constructor body, the
// instance accessors and static methods they reach. A read of a static initialized later in source
// order stops in the future-read ruling's words, naming the field, where Node reads undefined (the
// body, parameter, field and reached rows print undefined, the getter row NaN). A call the walk
// can't resolve keeps its stop, and a class with a base keeps the escape refusal, since a base
// constructor's reach isn't followed; Node prints 1 for both. The admitted programs agree with
// Node in internal/oracle's TestStaticSelfConstructionAgreesWithNode. Mutant:
// static-self-construction.mutants.md, and every row but the base one lowers.
func TestStaticSelfConstructionStops(t *testing.T) {
	t.Parallel()
	const later = "Adamic 0.1 refuses a static field read before its initializer has run: Box.later; declare the field earlier, or move the read after that field's initialization"
	const header = "class Box {\n\tstatic readonly first: Box = new Box();\n\tstatic readonly later: number = 7;\n"
	const footer = "}\nconsole.log(String(Box.first.value));\n"
	for _, probe := range []struct{ name, source, where, want string }{
		{"a later static read by the constructor body", header + "\treadonly value: number;\n\tconstructor() {\n\t\tthis.value = Box.later;\n\t}\n" + footer, "6:16", later},
		{"a later static read by a parameter default", header + "\treadonly value: number;\n\tconstructor(value: number = Box.later) {\n\t\tthis.value = value;\n\t}\n" + footer, "5:30", later},
		{"a later static read by an instance field initializer", header + "\treadonly value: number = Box.later;\n" + footer, "4:27", later},
		{"a later static read by an instance getter the constructor reads", header + "\treadonly value: number;\n\tconstructor() {\n\t\tthis.value = 1;\n\t\tconsole.log(String(this.doubled));\n\t}\n\tget doubled(): number {\n\t\treturn Box.later * 2;\n\t}\n" + footer, "10:10", later},
		{"a later static read by a static method the constructor calls", header + "\treadonly value: number;\n\tconstructor() {\n\t\tthis.value = Box.readLater();\n\t}\n\tstatic readLater(): number {\n\t\treturn Box.later;\n\t}\n" + footer, "9:10", later},
		{"a call the walk can't resolve", "const helper: () => number = () => 1;\nclass Box {\n\tstatic readonly first: Box = new Box();\n\treadonly value: number;\n\tconstructor() {\n\t\tthis.value = helper();\n\t}\n" + footer, "6:16", "stage 0 can't lower an indirect call before static fields are initialized; move it after the field initializers yet"},
		{"a class with a base", "class Base {\n\treadonly tag: string = 'base';\n}\nclass Box extends Base {\n\tstatic readonly first: Box = new Box();\n\treadonly value: number = 1;\n" + footer, "5:35", "Adamic 0.1 refuses a constructor object escaping before static fields are initialized; publish the class value after the last static field initializer"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.HasSuffix(err.Error(), "/main.a:"+probe.where+": "+probe.want) {
				t.Errorf("got %v, want main.a:%s: %s", err, probe.where, probe.want)
			}
		})
	}
}

// staticAssert is debug.ts's assert in .a, whose asserts signature .a stops first: it calls
// verboseDebugInfo only when it's passed one and the expression is false.
const staticAssert = "function fail(message: string): never {\n\tthrow new Error(message);\n}\nfunction assert(expression: boolean, message: string, verboseDebugInfo?: string | (() => string)): void {\n\tif (!expression) {\n\t\tlet text = `False expression: ${message}`;\n\t\tif (verboseDebugInfo) {\n\t\t\ttext += typeof verboseDebugInfo === 'string' ? verboseDebugInfo : verboseDebugInfo();\n\t\t}\n\t\tfail(text);\n\t}\n}\n"

// staticAssertBox is a class whose static first constructs it, its constructor making call first.
func staticAssertBox(call string) string {
	return "class Box {\n\tstatic readonly first: Box = new Box(1);\n\tstatic readonly later: string = 'ready';\n\treadonly value: number;\n\tconstructor(value: number) {\n\t\t" + call + "\n\t\tthis.value = value;\n\t}\n}\nconsole.log(String(Box.first.value));\n"
}

// Two reaches the static initialization walk follows, ruled by Adamic on #1451npc so tsc's
// Version.zero lowers. A call of a parameter is judged per call edge: assert reached with no
// verboseDebugInfo never calls it, and lowers; reached with a function, the function's reach is
// walked, and a later static read there stops naming the field, though Node prints 1 for each,
// since assert's expression holds. A function declared outside the class that reads it stops first
// at the dead-zone rule, since the class's outer binding isn't initialized while its statics run
// (Node throws a ReferenceError there if the function is called), a static method read as a value is
// refused as unbound, a static arrow field reading a later static is refused at its own initializer by
// the source-order rule, and an arrow written in the constructor is walked as the constructor's own. Each of
// those names the class; the per-edge judgment's witness reaches the later static without naming it
// (TestStaticWalkPerEdgeWitness). A member call on a library global runs no program code when both
// resolve by symbol to Node's Error.captureStackTrace, which lowers; a program object's
// captureStackTrace is program code and keeps the indirect call's stop (Node prints program code
// object, then 1). Mutants: static-self-construction.mutants.md.
func TestStaticWalkReachAdmissions(t *testing.T) {
	t.Parallel()
	const later = "Adamic 0.1 refuses a static field read before its initializer has run: Box.later; declare the field earlier, or move the read after that field's initialization"
	for _, probe := range []struct{ name, source, where, want string }{
		{"assert reached with no verboseDebugInfo", staticAssert + staticAssertBox("assert(value > 0, 'value');"), "", ""},
		{"assert reached with an outer function reading the class", staticAssert + "function describeLater(): string {\n\treturn Box.later;\n}\n" + staticAssertBox("assert(value > 0, 'value', describeLater);"), "14:9", "stage 0 can't lower a class used from a function that may run before the class's declaration (its temporal dead zone) yet"},
		{"assert reached with a static arrow field reading a later static", staticAssert + "class Box {\n\tstatic readonly describe = (): string => Box.later;\n\tstatic readonly first: Box = new Box(1);\n\tstatic readonly later: string = 'ready';\n\treadonly value: number;\n\tconstructor(value: number) {\n\t\tassert(value > 0, 'value', Box.describe);\n\t\tthis.value = value;\n\t}\n}\nconsole.log(String(Box.first.value));\n", "14:43", later},
		{"assert reached with a closure reading a later static", staticAssert + staticAssertBox("assert(value > 0, 'value', () => Box.later);"), "18:36", later},
		{"Node's Error.captureStackTrace", "import 'node:process';\n" + staticAssertBox("const target = new Error('target');\n\t\tError.captureStackTrace(target);"), "", ""},
		{"a program object's captureStackTrace", "const tracer = {\n\tcaptureStackTrace(target: object): void {\n\t\tconsole.log(`program code ${typeof target}`);\n\t},\n};\n" + staticAssertBox("tracer.captureStackTrace({});"), "11:3", "stage 0 can't lower an indirect call before static fields are initialized; move it after the field initializers yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if probe.want == "" {
				if err != nil {
					t.Fatalf("must lower: %v", err)
				}
				return
			}
			if err == nil || !strings.HasSuffix(err.Error(), "/main.a:"+probe.where+": "+probe.want) {
				t.Errorf("got %v, want main.a:%s: %s", err, probe.where, probe.want)
			}
		})
	}
}

// Error.captureStackTrace runs no program code only while a program can neither install
// Error.prepareStackTrace, the one function of its own V8 would call, nor replace
// Error.captureStackTrace itself, which tsc accepts and Node then runs in the host's place (Language's
// attack on #1451npc). So the walk's admission (staticLibraryCallsNothing) rests on these stops: the
// hook's typed write, its write through any and its define, and the member's direct write and its
// write through a view of Error. Node prints installed for each. Lifting any of them has to answer
// that admission first, and every later entry in libraryMembersCallingNothing pins the same roads
// for its own member and hooks here.
func TestPrepareStackTraceStaysStopped(t *testing.T) {
	t.Parallel()
	const prepare = "(error: Error, frames: unknown[]): string => `${error.name} ${frames.length}`"
	const replace = "(target: object): void => { console.log(`replaced ${typeof target}`); }"
	for _, probe := range []struct{ name, source, where, want string }{
		{"a typed write", "import 'node:process';\nError.prepareStackTrace = " + prepare + ";\nconsole.log('installed');\n", "2:1", "stage 0 can't lower node:globals.ErrorConstructor.prepareStackTrace yet"},
		{"a write through any", "import 'node:process';\n(Error as any).prepareStackTrace = " + prepare + ";\nconsole.log('installed');\n", "2:2", "Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)"},
		{"Object.defineProperty", "import 'node:process';\nObject.defineProperty(Error, 'prepareStackTrace', { value: " + prepare + " });\nconsole.log('installed');\n", "2:1", "Adamic 0.1 refuses Object.defineProperty; property descriptors can change the presence, type or access behavior of fields; Adamic fields have a fixed shape and are plain loads and stores"},
		{"captureStackTrace replaced", "import 'node:process';\nError.captureStackTrace = " + replace + ";\nconsole.log('installed');\n", "2:1", "stage 0 can't lower node:globals.ErrorConstructor.captureStackTrace written (fix: read or call Error.captureStackTrace) yet"},
		{"captureStackTrace replaced through a view of Error", "import 'node:process';\n(Error as { captureStackTrace: (target: object) => void }).captureStackTrace = " + replace + ";\nconsole.log('installed');\n", "2:2", "stage 0 can't lower reading Error yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.HasSuffix(err.Error(), "/main.a:"+probe.where+": "+probe.want) {
				t.Errorf("got %v, want main.a:%s: %s", err, probe.where, probe.want)
			}
		})
	}
}

// tsc's real Version.zero (semver.ts:47) with the debug.ts and core.ts functions its constructor reaches
// passes the static walk through both admissions: Error.captureStackTrace runs no program code, and
// assert's verboseDebugInfo() is judged by the calls reaching assert, none of which passes one. It then
// stops at a later wall, debug.ts's assertIsDefined, an asserts predicate over a generic, which isn't
// this walk's (Node prints 0.0.0-0 first). Main stops it at 92:38, the constructor escape this change
// lifts; mutant "Parameter calls never known" stops it at 67:111, the indirect call. When the predicate
// wall falls, it moves to internal/oracle's TestStaticSelfConstructionAgreesWithNode.
func TestTscVersionZeroPassesTheStaticWalk(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/static_self_construction/tsc_version_zero.ts")
	if err != nil {
		t.Fatal(err)
	}
	const want = "/main.ts:73:2: stage 0 can't lower checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } yet"
	_, err = lowerSourceAs(t, string(source), ".ts")
	if err == nil || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want ...%s", err, want)
	}
}

// The per-edge judgment's witness (Language, on #1451npc): a function declaration handed to a function
// that calls it reaches a later static without naming the class, through an instance getter on the instance
// it's handed, so neither the dead-zone rule nor the source-order rule sees it, and only staticParameterCalls
// walks reader's body from run's read(version) edge. Node prints undefined for the number-typed static; the
// control, later declared first, lowers (Node prints 7). Compiler's road, an arrow written in the static
// initializer and handed to the constructor, is the third row. Mutant "No per-edge check" in
// static-self-construction.mutants.md makes the first row lower.
func TestStaticWalkPerEdgeWitness(t *testing.T) {
	t.Parallel()
	const reader = "function run(read: (version: Version) => number, version: Version): number {\n\treturn read(version);\n}\nfunction reader(version: Version): number {\n\treturn version.size;\n}\n"
	const getter = "\tget size(): number {\n\t\treturn Version.later;\n\t}\n}\nconsole.log(String(Version.zero));\n"
	for _, row := range []struct{ name, source, where, want string }{
		{"a declared function reading a later static through a getter", reader + "class Version {\n\tstatic zero: number = run(reader, new Version());\n\tstatic later: number = 7;\n" + getter, "11:10", "Adamic 0.1 refuses a static field read before its initializer has run: Version.later; declare the field earlier, or move the read after that field's initialization"},
		{"the same with the static declared first", reader + "class Version {\n\tstatic later: number = 7;\n\tstatic zero: number = run(reader, new Version());\n" + getter, "", ""},
		{"an arrow in the static initializer handed to the constructor", "class Box {\n\tstatic zero: Box = new Box(() => Box.later);\n\tstatic later: number = 5;\n\tconstructor(read: () => number) {\n\t\tconsole.log(String(read()));\n\t}\n}\nconsole.log(typeof Box.zero);\n", "2:35", "Adamic 0.1 refuses a static field read before its initializer has run: Box.later; declare the field earlier, or move the read after that field's initialization"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, row.source)
			if row.want == "" {
				if err != nil {
					t.Fatalf("must lower: %v", err)
				}
				return
			}
			if err == nil || !strings.HasSuffix(err.Error(), "/main.a:"+row.where+": "+row.want) {
				t.Errorf("got %v, want main.a:%s: %s", err, row.where, row.want)
			}
		})
	}
}
