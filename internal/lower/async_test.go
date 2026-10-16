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
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
			t.Fatalf("user class with exact generated name %s: %v", generated.Name, err)
		}
	}
}
func TestAsyncGapsNameTheMissingPiece(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, want string }{
		{"async function f(): Promise<string> { const text = 'held'; function read(): string { return text; } await Promise.resolve(); return read(); }\nconsole.log(await f());", "canonical suspension environment"},
		{"async function f(): Promise<void> { exit: { await Promise.resolve(); break exit; } }\nawait f();", "async labeled regions"},
		{"async function f(value: number): Promise<number> { return value || await Promise.resolve(7); }\nconsole.log(`${await f(0)}`);", "value-preserving logical operands"},
		{"async function f(): Promise<void> { await new Promise<void>(() => {}); }\nawait f();", "Promise executors"},
		{"async function f(): Promise<void> { await Promise.all([Promise.resolve(1)]); }\nawait f();", "Promise.all"},
		{"async function f(): Promise<void> { try { throw new Error('why'); } catch { await Promise.resolve(); } }\nawait f();", "await in catch or finally"},
		{"async function f(): Promise<void> { try { await Promise.resolve(); } finally { await Promise.resolve(); } }\nawait f();", "await in catch or finally"},
		{"async function f(): Promise<number> { try { return await Promise.resolve(1); } finally { console.log('clean'); } }\nawait f();", "async finally completion routing"},
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
		{"async_refuse_loop_body_capture", "async function run(): Promise<void> {\n    const readers: (() => string)[] = [];\n    let i = 0;\n    while (i < 3) {\n        const held = `item${i}`;\n        readers.push(() => held);\n        await Promise.resolve();\n        i++;\n    }\n    const a = readers[0] ?? (() => \"missing\");\n    const b = readers[1] ?? (() => \"missing\");\n    const c = readers[2] ?? (() => \"missing\");\n    console.log(`${a()} ${b()} ${c()}`);\n}\nawait run();\n", "async per-iteration captured cells", false},
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

func TestAsyncRepeatedBindingRefusals(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"if (true) { const held = `item${i}`; readers.push(() => held); }",
		"try { throw new Error(`item${i}`); } catch (held) { if (held instanceof Error) { readers.push(() => held.message); } }",
	} {
		for _, loop := range []string{
			"while (i < 3) { BODY await Promise.resolve(); i++; }",
			"for (; i < 3; i++) { BODY await Promise.resolve(); }",
			"do { BODY await Promise.resolve(); i++; } while (i < 3);",
		} {
			source := "async function run(): Promise<void> { const readers: (() => string)[] = []; let i = 0; " + strings.ReplaceAll(loop, "BODY", body) + " } await run();"
			_, err := lowerSource(t, source)
			if err == nil || !strings.Contains(err.Error(), "async per-iteration captured cells") {
				t.Fatalf("repeated binding must be refused: %s: %v", source, err)
			}
		}
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
	for _, probe := range []struct{ name, source string }{
		{"unused", "async function f(): Promise<void> { const value = await Promise.resolve(); }\nawait f();"},
		{"observed", "async function f(): Promise<void> { const value = await Promise.resolve(); console.log(typeof value); } await f();"},
	} {
		if _, err := lowerSource(t, probe.source); err != nil {
			t.Fatal(err)
		}
	}
}
