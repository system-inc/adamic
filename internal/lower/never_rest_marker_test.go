package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
)

// tsc's debug.ts hands assertEachIsDefined, a generic overload set, to assertIsDefined as
// `stackCrawlMark || assertEachIsDefined`, where stackCrawlMark?: AnyFunction and AnyFunction is
// (...args: never[]) => void. A function whose every value, by the program-wide flow proof, only goes
// into never-rest slots, is compared, or is handed to Error.captureStackTrace is lowered as its identity
// marker (never_rest_marker.go, #avqasqt): one closure per function whose body never runs the
// declaration. These pins hold the shape and the proof's edges. tsc's shape lowers and agrees with Node,
// natively too, in .ts, in .a without asserts signatures, and through Error.captureStackTrace; two
// markers are === to themselves and not to each other; a cycle handing the mark back and forth proves;
// an export read as a value in another module proves; a one-signature generic proves where it stopped
// without a concrete contextual signature. Any use off the proof's list keeps main's road and main's
// stop: a call through the slot, a field, a variable, an ordinary function slot, typeof or .length
// through the parameter, and a shorthand property, whose identity then agrees with Node through the
// forwarder. Mutants, as patch files beside the brief: neverRestOnly answering yes for every function
// declaration reddens the off-list pins (3, 4, 5, 9, 7b); one marker for every function reddens the
// identity pin (2); (a) accepting without following the parameter reddens the call and typeof pins (3,
// 9); the proof accepting a followed parameter's call with callClosure's never-rest stop removed reddens
// pin 3, and TestNeverRestMarkerFence, run under ADAMIC_MARKER_FENCE=1, then sees the stub's panic; the
// binding's never-rest branch removed reddens tsc's shape at the refusal pass's binding stop.

// neverRestTsc is tsc's shape (debug.ts's fail, assertIsDefined and assertEachIsDefined), with the
// mark's presence said in fail's message, since captureStackTrace's stack is nothing a program reads.
const neverRestTsc = "interface Node {\n" +
	"\treadonly kind: number;\n" +
	"}\n" +
	"type AnyFunction = (...args: never[]) => void;\n" +
	"function fail(message: string, mark?: AnyFunction): never {\n" +
	"\tif (mark === undefined) {\n" +
	"\t\tthrow new Error(message);\n" +
	"\t}\n" +
	"\tthrow new Error(message + ' (marked)');\n" +
	"}\n" +
	"function assertIsDefined<T>(value: T, message?: string, mark?: AnyFunction): asserts value is NonNullable<T> {\n" +
	"\tif (value === undefined || value === null) {\n" +
	"\t\tfail(message ?? 'undefined', mark || assertIsDefined);\n" +
	"\t}\n" +
	"}\n" +
	"function assertEachIsDefined<T extends Node>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly T[];\n" +
	"function assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly NonNullable<T>[];\n" +
	"function assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction) {\n" +
	"\tfor (const v of value) {\n" +
	"\t\tassertIsDefined(v, message, mark || assertEachIsDefined);\n" +
	"\t}\n" +
	"}\n"

// neverRestTscA is neverRestTsc in .a, without the asserts signatures, whose predicate .a stops first.
const neverRestTscA = "interface Node {\n" +
	"\treadonly kind: number;\n" +
	"}\n" +
	"type AnyFunction = (...args: never[]) => void;\n" +
	"function fail(message: string, mark?: AnyFunction): never {\n" +
	"\tif (mark === undefined) {\n" +
	"\t\tthrow new Error(message);\n" +
	"\t}\n" +
	"\tthrow new Error(message + ' (marked)');\n" +
	"}\n" +
	"function assertIsDefined<T>(value: T, message?: string, mark?: AnyFunction): void {\n" +
	"\tif (value === undefined) {\n" +
	"\t\tfail(message ?? 'undefined', mark || assertIsDefined);\n" +
	"\t}\n" +
	"}\n" +
	"function assertEachIsDefined<T extends Node>(value: readonly T[], message?: string, mark?: AnyFunction): void;\n" +
	"function assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction): void;\n" +
	"function assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction): void {\n" +
	"\tfor (const v of value) {\n" +
	"\t\tassertIsDefined(v, message, mark || assertEachIsDefined);\n" +
	"\t}\n" +
	"}\n"

// neverRestTscRun runs tsc's shape on an honest array, then on one with a hole, whose fail throws with
// the mark present: the marker is truthy, so `mark || assertIsDefined` keeps it.
const neverRestTscRun = "const nodes: readonly Node[] = [{ kind: 1 }, { kind: 2 }];\n" +
	"assertEachIsDefined(nodes);\n" +
	"console.log('ok');\n" +
	"const holes: readonly (Node | undefined)[] = [{ kind: 1 }, undefined];\n" +
	"try {\n" +
	"\tassertEachIsDefined(holes, 'a hole');\n" +
	"} catch (error) {\n" +
	"\tif (error instanceof Error) {\n" +
	"\t\tconsole.log(error.message);\n" +
	"\t}\n" +
	"}\n" +
	"console.log('after');\n"

// neverRestOverloaded declares f, an overload set main reads as a value only by functionValue's stop, in
// six lines, so each off-list pin's stop is at a line of its own after them.
const neverRestOverloaded = "type AnyFunction = (...args: never[]) => void;\n" +
	"function f(): void;\n" +
	"function f(label: string): void;\n" +
	"function f(label?: string): void {\n" +
	"\tconsole.log(label ?? 'f');\n" +
	"}\n"

// neverRestOverloadedStop is main's stop for f read as a value, at a line and column.
func neverRestOverloadedStop(where string) string {
	return "main.ts:" + where + ": stage 0 can't lower an overloaded function as a value yet"
}

// neverRestAgrees lowers source as a file with extension, agrees with Node on the JavaScript backend, and
// builds it natively, under the sanitizers, wanting want on its stdout, what Node prints.
func neverRestAgrees(t *testing.T, source, extension, want string) {
	t.Helper()
	program := lowersAndAgreesWithNodeAs(t, source, extension)
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != want {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want %q, as Node prints", code, stdout, stderr, want)
	}
}

// neverRestStops lowers source as main.ts and wants main's NotYet, its whole text ending want.
func neverRestStops(t *testing.T, source, want string) {
	t.Helper()
	_, err := lowerSourceAs(t, source, ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Error(), want) {
		t.Fatalf("got %v, want an error ending %q", err, want)
	}
}

// Pins 1, 1b and 10: tsc's shape, the mark handed through two levels of `mark || f` into fail's
// never-rest parameter, lowers and agrees on both backends, the hole's failure caught with the mark
// present; in .a without asserts signatures too; and through Error.captureStackTrace(e, mark || fail),
// tsc's own fail, which evaluates its arguments and nothing else.
func TestNeverRestMarkerLowersTscShape(t *testing.T) {
	t.Parallel()
	const want = "ok\na hole (marked)\nafter\n"
	t.Run("ts", func(t *testing.T) {
		t.Parallel()
		neverRestAgrees(t, neverRestTsc+neverRestTscRun, ".ts", want)
	})
	t.Run("a", func(t *testing.T) {
		t.Parallel()
		neverRestAgrees(t, neverRestTscA+neverRestTscRun, ".a", want)
	})
	t.Run("captureStackTrace", func(t *testing.T) {
		t.Parallel()
		neverRestAgrees(t, "import 'node:process';\n"+
			"type AnyFunction = (...args: never[]) => void;\n"+
			"function fail(message: string, mark?: AnyFunction): never {\n"+
			"\tconst e = new Error(message);\n"+
			"\tError.captureStackTrace(e, mark || fail);\n"+
			"\tthrow e;\n"+
			"}\n"+
			"function check(value: number | undefined, mark?: AnyFunction): void {\n"+
			"\tif (value === undefined) {\n"+
			"\t\tfail('missing', mark || check);\n"+
			"\t}\n"+
			"}\n"+
			"check(1);\n"+
			"try {\n"+
			"\tcheck(undefined);\n"+
			"} catch (error) {\n"+
			"\tif (error instanceof Error) {\n"+
			"\t\tconsole.log(error.message);\n"+
			"\t}\n"+
			"}\n"+
			"console.log('done');\n", ".ts", "missing\ndone\n")
	})
}

// Pin 2: a marker is === to every read of its own function and to no other's. f, an overload set, and
// h, a plain function, are both proven, each its own marker.
func TestNeverRestMarkerIdentity(t *testing.T) {
	t.Parallel()
	neverRestAgrees(t, neverRestOverloaded+
		"function h(): void {\n"+
		"\tconsole.log('h');\n"+
		"}\n"+
		"function g(mark?: AnyFunction): void {\n"+
		"\tconsole.log(String(mark === f));\n"+
		"\tconsole.log(String(mark !== h));\n"+
		"}\n"+
		"g(f);\n"+
		"g(h);\n"+
		"g();\n"+
		"f('called');\n"+
		"h();\n", ".ts", "true\ntrue\nfalse\nfalse\nfalse\ntrue\ncalled\nh\n")
}

// Pin 6: a two-function cycle handing the mark back and forth proves, each parameter answering yes where
// the walk reaches it again (the greatest fixpoint), and agrees.
func TestNeverRestMarkerCycleProves(t *testing.T) {
	t.Parallel()
	neverRestAgrees(t, "type AnyFunction = (...args: never[]) => void;\n"+
		"function a(n: number, mark?: AnyFunction): void {\n"+
		"\tif (n <= 0) {\n"+
		"\t\tconsole.log(String(mark === a) + ' ' + String(mark === b));\n"+
		"\t\treturn;\n"+
		"\t}\n"+
		"\tb(n - 1, mark || a);\n"+
		"}\n"+
		"function b(n: number, mark?: AnyFunction): void {\n"+
		"\tif (n <= 0) {\n"+
		"\t\tconsole.log(String(mark === a) + ' ' + String(mark === b));\n"+
		"\t\treturn;\n"+
		"\t}\n"+
		"\ta(n - 1, mark || b);\n"+
		"}\n"+
		"a(3);\n"+
		"b(2);\n"+
		"a(0);\n", ".ts", "true false\nfalse true\nfalse false\n")
}

// Pin 7: an overload set exported from one module and read as a value in another, into a never-rest
// parameter there, proves across modules, since its reads are found by symbol, through the import, and
// agrees. The same function also stored in a local by the second module keeps main's road, which stops
// at its first value read, in check's body.
func TestNeverRestMarkerAcrossModules(t *testing.T) {
	t.Parallel()
	const library = "export type AnyFunction = (...args: never[]) => void;\n" +
		"export function stamp(): void;\n" +
		"export function stamp(label: string): void;\n" +
		"export function stamp(label?: string): void {\n" +
		"\tconsole.log(label ?? 'stamp');\n" +
		"}\n"
	const check = "import { stamp, type AnyFunction } from './library.ts';\n" +
		"function check(mark?: AnyFunction): void {\n" +
		"\tconsole.log(String(mark === stamp));\n" +
		"}\n"
	t.Run("proven", func(t *testing.T) {
		t.Parallel()
		directory := t.TempDir()
		for name, source := range map[string]string{"library.ts": library, "main.ts": check + "check(stamp);\ncheck();\nstamp('called');\n"} {
			if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		program := agreeEntry(t, filepath.Join(directory, "main.ts"), nil, 0, nil, nil)
		if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "true\nfalse\ncalled\n" {
			t.Fatalf("native: exit %d, stdout %q, stderr %q", code, stdout, stderr)
		}
	})
	t.Run("stored", func(t *testing.T) {
		t.Parallel()
		_, err := lowerModules(t, map[string]string{"library.ts": library, "main.ts": check + "const held: AnyFunction = stamp;\ncheck(stamp);\ncheck(held);\n"})
		var notYet *NotYet
		if want := neverRestOverloadedStop("3:30"); !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Error(), want) {
			t.Fatalf("got %v, want an error ending %q", err, want)
		}
	})
}

// Pin 8: a one-signature generic passed only into never-rest parameters, and called directly, lowers
// and agrees, where main stopped without a concrete contextual signature; and its === isn't the
// identity stop for specialized generic function values (hasGenericFunctionValues), since a marker is
// one closure, never an instance.
func TestNeverRestMarkerOneSignatureGeneric(t *testing.T) {
	t.Parallel()
	neverRestAgrees(t, "interface Node {\n"+
		"\treadonly kind: number;\n"+
		"}\n"+
		"type AnyFunction = (...args: never[]) => void;\n"+
		"function each<T extends Node>(value: readonly T[]): void {\n"+
		"\tfor (const v of value) {\n"+
		"\t\tconsole.log(String(v.kind));\n"+
		"\t}\n"+
		"}\n"+
		"function same(mark?: AnyFunction): boolean {\n"+
		"\treturn mark === each;\n"+
		"}\n"+
		"console.log(String(same(each)));\n"+
		"console.log(String(same()));\n"+
		"each([{ kind: 3 }]);\n", ".ts", "true\nfalse\n3\n")
}

// neverRestCalledThroughSlot is pin 3's program: f handed to a never-rest parameter that's called with
// no arguments. run's body is lowered before g's, so main's stop is at f's value.
const neverRestCalledThroughSlot = neverRestOverloaded +
	"function run(): void {\n" +
	"\tg(f);\n" +
	"}\n" +
	"function g(mark: AnyFunction): void {\n" +
	"\tmark();\n" +
	"}\n" +
	"run();\n" +
	"console.log('reached');\n"

// Pins 3, 4, 5 and 9: a use off the proof's list keeps main's road for every read of the function, and
// main's stop at its first value read: a call through the never-rest parameter, a variable, a field, an
// ordinary function slot, and typeof or .length through the parameter.
func TestNeverRestMarkerOffListKeepsMainsStop(t *testing.T) {
	t.Parallel()
	const optional = "function g(mark?: AnyFunction): void {\n" +
		"\tconsole.log(String(mark === undefined));\n" +
		"}\n"
	for _, probe := range []struct{ name, source, where string }{
		{"called through the slot", neverRestCalledThroughSlot, "8:4"},
		{"a variable", neverRestOverloaded + optional + "const k: AnyFunction = f;\ng(f);\ng(k);\n", "10:24"},
		{"a field", neverRestOverloaded + optional + "const o = { m: f };\ng(f);\ng(o.m);\n", "10:16"},
		{"an ordinary slot", neverRestOverloaded + "interface Node {\n\treadonly kind: number;\n}\nfunction visitAll(visit: (value: readonly Node[]) => void): void {\n\tvisit([]);\n}\n" + optional + "visitAll(f);\ng(f);\n", "16:10"},
		{"typeof through the parameter", neverRestOverloaded + "function run(): void {\n\tg(f);\n}\nfunction g(mark?: AnyFunction): void {\n\tconsole.log(typeof mark);\n}\nrun();\n", "8:4"},
		{"length through the parameter", neverRestOverloaded + "function run(): void {\n\tg(f);\n}\nfunction g(mark: AnyFunction): void {\n\tconsole.log(String(mark.length));\n}\nrun();\n", "8:4"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			neverRestStops(t, probe.source, neverRestOverloadedStop(probe.where))
		})
	}
}

// Pin 4, its shorthand: a function also read as { h } keeps main's road, so the forwarder the shorthand
// makes is the value check is handed too, and === agrees with Node. Were the shorthand missed by the
// proof, check would be handed a marker and print false.
func TestNeverRestMarkerShorthandKeepsForwarder(t *testing.T) {
	t.Parallel()
	neverRestAgrees(t, "type AnyFunction = (...args: never[]) => void;\n"+
		"function h(): void {\n"+
		"\tconsole.log('h');\n"+
		"}\n"+
		"const o = { h };\n"+
		"function check(mark?: AnyFunction): void {\n"+
		"\tconsole.log(String(mark === o.h));\n"+
		"}\n"+
		"check(h);\n"+
		"o.h();\n", ".ts", "true\nh\n")
}

// Pin 11: the binding's own pin. tsc's assertEachIsDefined stored in a variable, which the proof answers
// no for, keeps main's road, and stops past the binding, which reads never for the never-rest slot
// (type_argument_bindings.go), rather than at a type argument that can't be read.
func TestNeverRestMarkerBindingReadsNever(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, neverRestTsc+"const m: AnyFunction = assertEachIsDefined;\nconsole.log(String(m === assertEachIsDefined));\n", ".ts")
	if err == nil || strings.Contains(err.Error(), "can't be read") {
		t.Fatalf("want a stop past the binding, got %v", err)
	}
	t.Logf("main's road stops at %v", err)
}

// The fence behind the proof, for the mutant that lets pin 3's program lower (m4-fence.diff): a call
// that ever reaches a marker runs its stub, which panics in its own words and exits 70 on both
// backends, never the declaration's body. Run only under ADAMIC_MARKER_FENCE=1, with the mutant applied.
func TestNeverRestMarkerFence(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_MARKER_FENCE") != "1" {
		t.Skip("ADAMIC_MARKER_FENCE=1 runs the fence under m4-fence.diff, which lets a call reach a marker")
	}
	program, err := lowerSourceAs(t, neverRestCalledThroughSlot, ".ts")
	if err != nil {
		t.Fatalf("the mutant must admit the program: %v", err)
	}
	const panicked = "adamic: panic: " + neverRestMarkerCalled
	generated := filepath.Join(t.TempDir(), "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := runAgreementNode(t, generated); got.code != 70 || !strings.Contains(string(got.stderr), panicked) || len(got.stdout) != 0 {
		t.Fatalf("JavaScript backend: exit %d, stdout %q, stderr %q; want exit 70 and %q", got.code, got.stdout, got.stderr, panicked)
	}
	if code, stdout, stderr := brandedNative(t, program); code != 70 || !strings.Contains(stderr, panicked) || stdout != "" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want exit 70 and %q", code, stdout, stderr, panicked)
	}
}
