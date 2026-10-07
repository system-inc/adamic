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
		{"async function f(): Promise<void> { await new Promise<void>(() => {}); }\nawait f();", "Promise executors"},
		{"async function f(): Promise<void> { await Promise.all([Promise.resolve(1)]); }\nawait f();", "Promise.all"},
		{"async function f(): Promise<void> { try { throw new Error('why'); } catch { await Promise.resolve(); } }\nawait f();", "await in catch or finally"},
		{"async function f(): Promise<void> { try { await Promise.resolve(); } finally { await Promise.resolve(); } }\nawait f();", "await in catch or finally"},
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

func TestPromisePayloadCannotHideUserCycles(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "interface Box { promise: Promise<Box> | undefined; }\nfunction stash(box: Box, promise: Promise<Box>): void { box.promise = promise; }\nconst box: Box = {promise: undefined};\nconst promise = Promise.resolve(box);\nstash(box,promise);")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
		t.Fatalf("Promise payload hid the user back-reference: %v", err)
	}
}
