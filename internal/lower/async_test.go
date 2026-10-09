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
		{"async function f(): Promise<void> { await Promise.resolve(); throw new Error(); }\nawait f();", "Error construction without one explicit message"},
		{"async function f(): Promise<void> { await new Promise<void>(() => {}); }\nawait f();", "Promise executors"},
		{"async function f(): Promise<void> { await Promise.all([Promise.resolve(1)]); }\nawait f();", "full Promise surface"},
		{"async function f(): Promise<void> { try { await Promise.resolve(); } finally { console.log('cleanup'); } }\nawait f();", "try/catch/finally"},
		{"async function f(): Promise<void> { const value = await Promise.resolve(); }\nawait f();", "void-valued locals"},
		{"async function f(): Promise<void> { await undefined; }\nawait f();", "undefined/void-valued expressions"},
		{"async function f(): Promise<object> { return {}; }\nawait f();", "object, union"},
		{"async function f(): Promise<void> { await f(); }\nawait f();", "recursive async call graphs"},
		{"async function f(): Promise<void> { while (false) { await Promise.resolve(); } }\nawait f();", "control flow"},
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

// Compound operands still need an async function value; they must name the gap rather than
// indexing the empty synchronous function table. A typeof observation never calls its operand.
func TestAsyncTypeOfCompoundFunctionValueIsNotYet(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{"typeof (true ? f : f)", "typeof (f === f)"} {
		_, err := lowerSource(t, "async function f(): Promise<void> {}\nconsole.log("+expression+");\n")
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "async function f as a value") || notYet.Where == "" {
			t.Fatalf("%s: want a located async function value NotYet, got %v", expression, err)
		}
	}
}
