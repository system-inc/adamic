package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"os"
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
		// The user's cycle may be refused or, with graph regions, owned by a region; either way it
		// must not be exempt as the runtime's generated type is, which would leave it uncounted.
		program, err := lowerSource(t, source)
		var refused *Refused
		claimed := err == nil && len(program.GraphTypes) == 0
		if (err != nil && (!errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable"))) || claimed {
			t.Fatalf("user class with exact generated name %s: %v", generated.Name, err)
		}
	}
}
func TestAsyncGapsNameTheMissingPiece(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, want string }{
		{`const source = {[Symbol.iterator](){return {next(){return {value:1,done:false};}}}}; for(const item of source){await Promise.resolve(); break;}`, "async for-of over a user iterator"},
		{`const source = {[Symbol.iterator](){return {next(){return {value:1,done:false};}}}}; async function f():Promise<void>{for(const item of source){await Promise.resolve(); break;}} await f();`, "async for-of over a user iterator"},
		{"import { parallelMap } from 'adamic'; async function f(): Promise<string> { const extra = 3; await Promise.resolve(); const items: readonly number[] = [1,2]; return parallelMap(items, item => item + extra).join(','); } console.log(await f());", "pool tasks capturing an async environment"},
		{"async function f(): Promise<void> { await new Promise<void>(() => {}); }\nawait f();", "Promise executors"},
		{"async function f(): Promise<void> { await Promise.all([Promise.resolve(1)]); }\nawait f();", "Promise.all"},
		{"async function f(): Promise<void> { const value = await Promise.resolve(); }\nawait f();", "type void"},
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
		if probe.want == "unawaited async task" || probe.want == "void operator" || probe.want == "refuses ==" {
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

func TestPromisePayloadCannotHideUserCycles(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "interface Box { promise: Promise<Box> | undefined; }\nfunction stash(box: Box, promise: Promise<Box>): void { box.promise = promise; }\nconst box: Box = {promise: undefined};\nconst promise = Promise.resolve(box);\nstash(box,promise);")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
		t.Fatalf("Promise payload hid the user back-reference: %v", err)
	}
}

func TestAsyncSuspensionEndsFreshConfinement(t *testing.T) {
	source, err := os.ReadFile("../oracle/testdata/async_fresh_holder.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	checked := false
	for _, write := range fresh.ProveWrites(program) {
		if write.Name == "item" {
			checked = true
			if write.Proven {
				t.Fatal("frame-held value retained fresh confinement across suspension")
			}
		}
	}
	if !checked {
		t.Fatal("holder write was not analyzed")
	}
}

func TestAsyncProgramsDisableCallFreshnessSummaries(t *testing.T) {
	program, err := lowerSource(t, `function make(): { text: string } { return { text: "held" }; }
function write(holder: { item: { text: string } | undefined }): void { holder.item = make(); }
async function run(): Promise<void> { await Promise.resolve(); }
const holder: { item: { text: string } | undefined } = { item: undefined };
write(holder);
await run();`)
	if err != nil {
		t.Fatal(err)
	}
	checked := false
	for _, write := range fresh.ProveWrites(program) {
		if write.Name == "item" {
			checked = true
			if write.Proven {
				t.Fatal("async program used a call freshness summary")
			}
		}
	}
	if !checked {
		t.Fatal("holder write was not analyzed")
	}
}

// Ordinary function values also cover the formerly unsafe compound observations.
func TestAsyncTypeOfCompoundFunctionValuesLower(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{"typeof (true ? f : f)", "typeof (f === f)"} {
		_, err := lowerSource(t, "async function f(): Promise<void> {}\nconsole.log("+expression+");\n")
		if err != nil {
			t.Fatalf("%s: %v", expression, err)
		}
	}
}

func TestAsyncReaderRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, want string
		cycle      bool
	}{
		{"async_refuse_frame_capture_cycle", "async frame capture cycle", true},
		{"async_refuse_return_thenable", "return of thenables", false},
		{"async_refuse_arrow_thenable", "return of thenables", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source, err := os.ReadFile("../oracle/testdata/async_refused/" + probe.name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
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
		})
	}
}

func TestAsyncRepeatedBindingsLower(t *testing.T) {
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
			if err != nil {
				t.Fatalf("repeated binding must lower: %s: %v", source, err)
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
