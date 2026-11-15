package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestAsyncGeneratedIdentityCannotBeClaimedBySource(t *testing.T) {
	t.Parallel()
	for _, generated := range ir.AsyncGeneratedTypes() {
		if !fresh.RuntimeBreaksCycles(generated) {
			t.Fatal("audited generated identity refused")
		}
		if fresh.RuntimeBreaksCycles(&ir.GeneratedType{Name: generated.Name}) {
			t.Fatal("a name claimed the runtime exemption")
		}
		source := "class " + generated.Name + " { next: " + generated.Name + " | undefined = undefined; }\nconst self = new " + generated.Name + "();\nself.next = self;\n"
		graphAccepted(t, source)
	}
}
func TestAsyncGapsNameTheMissingPiece(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, want string }{
		{"async function f(): Promise<string> { const text = 'held'; function read(): string { return text; } await Promise.resolve(); return read(); }\nconsole.log(await f());", "canonical suspension environment"},
		{"async function f(value: number): Promise<number> { return value || await Promise.resolve(7); }\nconsole.log(`${await f(0)}`);", "value-preserving logical operands"},
		{"async function f(): Promise<void> { await new Promise<void>(() => {}); }\nawait f();", "Promise executors"},
		{"async function f(): Promise<void> { await Promise.all([Promise.resolve(1)]); }\nawait f();", "Promise.all"},
		{"async function f(): Promise<number> { return Promise.resolve(1); }\nawait f();", "Promise adoption"},
		{"async function f(): Promise<number> { return await {then(resolve: (value: number) => void): void { resolve(1); }}; }\nawait f();", "thenables"},
		{"async function f(): Promise<void> {}\nf();", "unawaited async task"},
		{"async function f(): Promise<void> { await Promise.resolve(); }\nvoid f();", "void operator"},
		{"async function f(): Promise<void> { console.log(`${1 == 1}`); }\nawait f();", "refuses =="},
	} {
		_, err := lowerSource(t, probe.source)
		if err == nil || !strings.Contains(err.Error(), probe.want) {
			t.Errorf("%s: got %v, want %s", probe.source, err, probe.want)
		}
		if strings.Contains(probe.want, "task") || probe.want == "void operator" || probe.want == "refuses ==" {
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Errorf("permanent refusal became NotYet: %v", err)
			}
		} else {
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Errorf("gap became refusal: %v", err)
			}
		}
	}
}

// Async function values now use ordinary function adapters.
func TestAsyncTypeOfCompoundFunctionValuesLower(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{"typeof (true ? f : f)", "typeof (f === f)"} {
		if _, err := lowerSource(t, "async function f(): Promise<void> {}\nconsole.log("+expression+");\n"); err != nil {
			t.Fatal(err)
		}
	}
}

// Empty bodies isolate the async-generator ruling from yield and await refusals.
func TestAsyncGeneratorsRemainRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"declaration", "async function* values() {}"},
		{"expression", "const values = async function* () {};"},
		{"class method", "class Values { async *values() {} }"},
		{"object method", "const values = { async *values() {} };"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || refused.What != "an async generator function" || refused.Where == "" {
				t.Fatalf("want located async-generator refusal, got %v", err)
			}
		})
	}
}

func TestPromisePayloadCannotHideUserCycles(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "interface Box { promise: Promise<Box> | undefined; }\nfunction stash(box: Box, promise: Promise<Box>): void { box.promise = promise; }\nconst box: Box = {promise: undefined};\nconst promise = Promise.resolve(box);\nstash(box,promise);")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
		t.Fatalf("Promise payload hid the user back-reference: %v", err)
	}
}

func TestAsyncReaderRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, want string
		cycle              bool
	}{
		{"async_refuse_frame_capture_cycle", "async function make(seed: string): Promise<(n: number) => string> {\n    const text = `${seed}:held`;\n    const read = () => text;\n    await Promise.resolve();\n    return (n: number) => `${read()}:${n}`;\n}\nconst escaped = await make(`seed${3}`);\nconsole.log(escaped(7));\n", "async frame capture cycle", true},
		{"async_refuse_return_thenable", "async function f(): Promise<number> {\n    return {\n        then(resolve: (value: number) => void): void {\n            resolve(7);\n        }\n    };\n}\nconsole.log(`${await f()}`);\n", "return of thenables", false},
		{"async_refuse_arrow_thenable", "const f = async (): Promise<number> => ({\n    then(resolve: (value: number) => void): void {\n        resolve(7);\n    }\n});\nconsole.log(`${await f()}`);\n", "return of thenables", false},
	} {
		_, err := lowerSource(t, probe.source)
		if err == nil || !strings.Contains(err.Error(), probe.want) {
			t.Fatalf("reader probe must be refused by name %q: %v", probe.want, err)
		}
		var refused *Refused
		var notYet *NotYet
		if probe.cycle {
			if !errors.As(err, &refused) {
				t.Fatalf("cycle must be Refused: %v", err)
			}
		} else if !errors.As(err, &notYet) {
			t.Fatalf("unproved semantics must be NotYet: %v", err)
		}
	}
}

// asyncRepeatedLoops are the loops TestAsyncRepeatedBindingsLower puts a body in, each with an await, so
// every iteration's bindings are the frame's again.
var asyncRepeatedLoops = []string{
	"while (i < 3) { BODY await Promise.resolve(); i++; }",
	"for (; i < 3; i++) { BODY await Promise.resolve(); }",
	"do { BODY await Promise.resolve(); i++; } while (i < 3);",
}

// asyncRepeatedSource is a loop of asyncRepeatedLoops holding body, in an async function that keeps readers.
func asyncRepeatedSource(loop, body string) string {
	return "async function run(): Promise<void> { const readers: (() => string)[] = []; let i = 0; " + strings.ReplaceAll(loop, "BODY", body) + " } await run();"
}

// The second body (rows 004 to 006) captures the catch binding itself in a closure kept in readers,
// Language's u00: it lowers by the catch narrowing (loose_slots.go's looseCaughtRoad), since every throw
// is a new Error and nothing writes readers into a loose slot, so what's caught can't reach readers. The
// third is Adamic's twin (#kaz4pk8, 03:12Z), which captures a string read from the caught value instead.
func TestAsyncRepeatedBindingsLower(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"if (true) { const held = `item${i}`; readers.push(() => held); }",
		"try { throw new Error(`item${i}`); } catch (held) { if (held instanceof Error) { readers.push(() => held.message); } }",
		"try { throw new Error(`item${i}`); } catch (held) { const message = held instanceof Error ? held.message : \"thrown\"; readers.push(() => message); }",
	} {
		for _, loop := range asyncRepeatedLoops {
			source := asyncRepeatedSource(loop, body)
			_, err := lowerSource(t, source)
			if err != nil {
				t.Fatalf("repeated binding must lower: %s: %v", source, err)
			}
		}
	}
}

// asyncCatchRoads load the readers array into a caught value, roads around the catch narrowing (Runtime
// and Adamic on #kaz4pk8 at 03:13Z, and Lower's thenable, whose then calls reject by a parameter's name). Each must still stop. Today each stops at an older wall
// before the narrowing judges it, one of the seven its premise stands on (looseCaughtRoad); the day that
// wall lifts, the value it lets into the thrown value enters H, and the row must stop at the loose-slot
// stop instead, by the finder's own reach, or go red here.
var asyncCatchRoads = []struct{ name, wall, source string }{
	{"a thrown Error subclass holding readers", "a base that isn't a declared class",
		"class Carrier extends Error { readers: (() => string)[] = []; }\nasync function run(): Promise<void> { const readers: (() => string)[] = []; try { const carrier = new Carrier(\"carried\"); carrier.readers = readers; throw carrier; } catch (held) { if (held instanceof Error) { readers.push(() => held.message); } } await Promise.resolve(); console.log(String(readers.length)); }\nawait run();\n"},
	{"an Error whose cause is readers", "writing a possibly absent optional own field",
		"async function run(): Promise<void> { const readers: (() => string)[] = []; try { const error = new Error(\"carried\"); error.cause = readers; throw error; } catch (held) { if (held instanceof Error) { readers.push(() => held.message); } } await Promise.resolve(); console.log(String(readers.length)); }\nawait run();\n"},
	{"Promise.reject(readers) into an awaited catch", "Promise rejection with a non-Error",
		"async function run(): Promise<void> { const readers: (() => string)[] = []; try { await Promise.reject(readers); } catch (held) { if (held instanceof Error) { readers.push(() => held.message); } } console.log(String(readers.length)); }\nawait run();\n"},
	{"an awaited thenable whose then rejects with readers", "thenables",
		"async function run(): Promise<void> { const readers: (() => string)[] = []; try { await { then(resolve: (value: number) => void, reject: (reason: unknown) => void): void { reject(readers); } }; } catch (held) { if (held instanceof Error) { readers.push(() => held.message); } } console.log(String(readers.length)); }\nawait run();\n"},
}

// TestAsyncCatchRoadsStop holds each of asyncCatchRoads to a stop: its wall today, or the loose-slot stop
// once that wall lifts.
func TestAsyncCatchRoadsStop(t *testing.T) {
	t.Parallel()
	for _, road := range asyncCatchRoads {
		t.Run(road.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, road.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !(strings.Contains(notYet.What, road.wall) || strings.HasSuffix(notYet.What, looseStopWords)) {
				t.Fatalf("got %v, want a stop at %q or the loose-slot stop", err, road.wall)
			}
			t.Logf("stop: %v", err)
		})
	}
}

func TestAsyncReturnThenableShapes(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`class Then { then(resolve: (value: number) => void): void { resolve(7); } }
async function f(): Promise<number> { return new Then(); }
const result = await f();`,
		`async function f(flag: boolean): Promise<number> {
return flag ? 4 : { then(resolve: (value: number) => void): void { resolve(7); } };
}
const result = await f(false);`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "return of thenables") {
			t.Fatalf("class or union thenable must be refused by name: %v", err)
		}
	}
}

// Async unit 2 represents a void await's value as undefined. Keep both the
// formerly refused unused binding and the observable registered fixture covered.
func TestAsyncVoidAwaitValuesLower(t *testing.T) {
	t.Parallel()
	if _, err := lowerSource(t, "async function f(): Promise<void> { const value = await Promise.resolve(); }\nawait f();"); err != nil {
		t.Fatal(err)
	}
	// Reading the void binding is H2's refusal (#x2gt39p): a Promise<void> can resolve to anything a
	// value-returning function hands it.
	_, err := lowerSource(t, "async function f(): Promise<void> { const value = await Promise.resolve(); console.log(typeof value); } await f();")
	if err == nil || !strings.Contains(err.Error(), "a read of a void value") {
		t.Fatalf("want the observed void binding refused, got %v", err)
	}
}
