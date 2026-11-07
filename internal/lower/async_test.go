package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestAsyncGeneratedIdentityCannotBeClaimedBySource(test *testing.T) {
	test.Parallel()
	for _, generated := range ir.AsyncGeneratedTypes() {
		if !fresh.RuntimeBreaksCycles(generated) {
			test.Fatal("audited generated identity refused")
		}
		if fresh.RuntimeBreaksCycles(&ir.GeneratedType{Name: generated.Name}) {
			test.Fatal("a name claimed the runtime exemption")
		}
		source := "class " +
			generated.Name +
			" { next: " +
			generated.Name +
			" | undefined = undefined; }\nconst self = new " +
			generated.Name +
			"();\nself.next = self;\n"
		graphAccepted(test, source)
	}
}
func TestAsyncGapsNameTheMissingPiece(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ source, want string }{
		{
			"async function f(): Promise<string> { const text = 'held'; function read(): string { return text; } await Promise.resolve(); return read(); }\nconsole.log(await f());",
			"canonical suspension environment",
		},
		{
			"async function f(value: number): Promise<number> { return value || await Promise.resolve(7); }\nconsole.log(`${await f(0)}`);",
			"value-preserving logical operands",
		},
		{"async function f(): Promise<void> { await new Promise<void>(() => {}); }\nawait f();", "Promise executors"},
		{"async function f(): Promise<void> { await Promise.all([Promise.resolve(1)]); }\nawait f();", "Promise.all"},
		{"async function f(): Promise<number> { return Promise.resolve(1); }\nawait f();", "Promise adoption"},
		{
			"async function f(): Promise<number> { return await {then(resolve: (value: number) => void): void { resolve(1); }}; }\nawait f();",
			"thenables",
		},
		{"async function f(): Promise<void> {}\nf();", "unawaited async task"},
		{"async function f(): Promise<void> { await Promise.resolve(); }\nvoid f();", "void operator"},
		{"async function f(): Promise<void> { console.log(`${1 == 1}`); }\nawait f();", "refuses =="},
	} {
		_, lowerSourceError := lowerSource(test, probe.source)
		if lowerSourceError == nil || !strings.Contains(lowerSourceError.Error(), probe.want) {
			test.Errorf("%s: got %v, want %s", probe.source, lowerSourceError, probe.want)
		}
		if strings.Contains(probe.want, "task") || probe.want == "void operator" || probe.want == "refuses ==" {
			var refused *Refused
			if !errors.As(lowerSourceError, &refused) {
				test.Errorf("permanent refusal became NotYet: %v", lowerSourceError)
			}
		} else {
			var notYet *NotYet
			if !errors.As(lowerSourceError, &notYet) {
				test.Errorf("gap became refusal: %v", lowerSourceError)
			}
		}
	}
}

// Async function values now use ordinary function adapters.
func TestAsyncTypeOfCompoundFunctionValuesLower(test *testing.T) {
	test.Parallel()
	for _, expression := range []string{"typeof (true ? f : f)", "typeof (f === f)"} {
		if _, lowerSourceError := lowerSource(
			test,
			"async function f(): Promise<void> {}\nconsole.log("+expression+");\n",
		); lowerSourceError != nil {
			test.Fatal(lowerSourceError)
		}
	}
}

// Empty bodies isolate the async-generator ruling from yield and await refusals.
func TestAsyncGeneratorsRemainRefused(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"declaration", "async function* values() {}"},
		{"expression", "const values = async function* () {};"},
		{"class method", "class Values { async *values() {} }"},
		{"object method", "const values = { async *values() {} };"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var refused *Refused
			if !errors.As(lowerSourceError, &refused) ||
				refused.What != "an async generator function" ||
				refused.Where == "" {
				test.Fatalf("want located async-generator refusal, got %v", lowerSourceError)
			}
		})
	}
}

func TestPromisePayloadCannotHideUserCycles(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(
		test,
		"interface Box { promise: Promise<Box> | undefined; }\nfunction stash(box: Box, promise: Promise<Box>): void { box.promise = promise; }\nconst box: Box = {promise: undefined};\nconst promise = Promise.resolve(box);\nstash(box,promise);",
	)
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) || !strings.Contains(lowerSourceError.Error(), "adamic/cycle-capable") {
		test.Fatalf("Promise payload hid the user back-reference: %v", lowerSourceError)
	}
}

func TestAsyncReaderRefusals(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name, source, want string
		cycle              bool
	}{
		{
			"async_refuse_frame_capture_cycle",
			"async function make(seed: string): Promise<(n: number) => string> {\n    const text = `${seed}:held`;\n    const read = () => text;\n    await Promise.resolve();\n    return (n: number) => `${read()}:${n}`;\n}\nconst escaped = await make(`seed${3}`);\nconsole.log(escaped(7));\n",
			"async frame capture cycle",
			true,
		},
		{
			"async_refuse_return_thenable",
			"async function f(): Promise<number> {\n    return {\n        then(resolve: (value: number) => void): void {\n            resolve(7);\n        }\n    };\n}\nconsole.log(`${await f()}`);\n",
			"return of thenables",
			false,
		},
		{
			"async_refuse_arrow_thenable",
			"const f = async (): Promise<number> => ({\n    then(resolve: (value: number) => void): void {\n        resolve(7);\n    }\n});\nconsole.log(`${await f()}`);\n",
			"return of thenables",
			false,
		},
	} {
		_, lowerSourceError := lowerSource(test, probe.source)
		if lowerSourceError == nil || !strings.Contains(lowerSourceError.Error(), probe.want) {
			test.Fatalf("reader probe must be refused by name %q: %v", probe.want, lowerSourceError)
		}
		var refused *Refused
		var notYet *NotYet
		if probe.cycle {
			if !errors.As(lowerSourceError, &refused) {
				test.Fatalf("cycle must be Refused: %v", lowerSourceError)
			}
		} else if !errors.As(lowerSourceError, &notYet) {
			test.Fatalf("unproved semantics must be NotYet: %v", lowerSourceError)
		}
	}
}

func TestAsyncRepeatedBindingsLower(test *testing.T) {
	test.Parallel()
	for _, body := range []string{
		"if (true) { const held = `item${i}`; readers.push(() => held); }",
		"try { throw new Error(`item${i}`); } catch (held) { if (held instanceof Error) { readers.push(() => held.message); } }",
	} {
		for _, loop := range []string{
			"while (i < 3) { BODY await Promise.resolve(); i++; }",
			"for (; i < 3; i++) { BODY await Promise.resolve(); }",
			"do { BODY await Promise.resolve(); i++; } while (i < 3);",
		} {
			source := "async function run(): Promise<void> { const readers: (() => string)[] = []; let i = 0; " +
				strings.ReplaceAll(loop, "BODY", body) +
				" } await run();"
			_, lowerSourceError := lowerSource(test, source)
			if lowerSourceError != nil {
				test.Fatalf("repeated binding must lower: %s: %v", source, lowerSourceError)
			}
		}
	}
}

func TestAsyncReturnThenableShapes(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`class Then { then(resolve: (value: number) => void): void { resolve(7); } }
async function f(): Promise<number> { return new Then(); }
const result = await f();`,
		`async function f(flag: boolean): Promise<number> {
return flag ? 4 : { then(resolve: (value: number) => void): void { resolve(7); } };
}
const result = await f(false);`,
	} {
		_, lowerSourceError := lowerSource(test, source)
		var notYet *NotYet
		if !errors.As(lowerSourceError, &notYet) || !strings.Contains(lowerSourceError.Error(), "return of thenables") {
			test.Fatalf("class or union thenable must be refused by name: %v", lowerSourceError)
		}
	}
}

// Async unit 2 represents a void await's value as undefined. Keep both the
// formerly refused unused binding and the observable registered fixture covered.
func TestAsyncVoidAwaitValuesLower(test *testing.T) {
	test.Parallel()
	if _, lowerSourceError := lowerSource(
		test,
		"async function f(): Promise<void> { const value = await Promise.resolve(); }\nawait f();",
	); lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	// Reading the void binding is H2's refusal (#x2gt39p): a Promise<void> can resolve to anything a
	// value-returning function hands it.
	_, lowerSourceError := lowerSource(
		test,
		"async function f(): Promise<void> { const value = await Promise.resolve(); console.log(typeof value); } await f();",
	)
	if lowerSourceError == nil || !strings.Contains(lowerSourceError.Error(), "a read of a void value") {
		test.Fatalf("want the observed void binding refused, got %v", lowerSourceError)
	}
}
