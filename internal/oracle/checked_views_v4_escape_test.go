package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

const v4EscapeDeclarations = `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run: (value: string) => string; }
`

func v4EscapeRefusal(t *testing.T, source, output string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "escape.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != output || len(truth.stderr) != 0 {
		t.Fatalf("Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err == nil {
		// The guard mutant must be a semantic admission, not a clang or sanitizer failure.
		v4EscapeAdmitted(t, program, truth)
	}
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "escape.a:") || refusal.What != "an unproven escaping callable read of view.run" || refusal.Fix != "prove the producer parameter and result relation before this read, or call the member directly through the view" {
		t.Fatalf("escaping read refusal with path and fix: %v", err)
	}
}

func v4EscapeAdmitted(t *testing.T, program *ir.Program, truth run) {
	t.Helper()
	sanitized, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("guard mutant is not a clean admission in %s: %s", backend, difference)
		}
		t.Logf("guard mutant admitted in %s: exit %d output %q", backend, got.exitCode, got.stdout)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestV4EscapeArgumentRefusal(t *testing.T) {
	t.Parallel()
	v4EscapeRefusal(t, v4EscapeDeclarations+`
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>'producer'};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log('read'); console.log(escaped('bad')); }
probe(raw);`, "read\nproducer\n")
}

func TestV4EscapeResultRefusal(t *testing.T) {
	t.Parallel()
	v4EscapeRefusal(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run: (value:string) => 'ok'; }
const raw={kind:'Receiver' as const,run:(value:string):string=>'bad'};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log('read'); }
probe(raw);`, "read\n")
}

func TestV4EscapePassedRefusal(t *testing.T) {
	t.Parallel()
	v4EscapeRefusal(t, v4EscapeDeclarations+`
function consume(callback:(value:string)=>string):void { console.log(callback('bad')); }
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>'producer'};
function probe(base:Base):void { const view=base as Target; consume(view.run); }
probe(raw);`, "producer\n")
}

func TestV4EscapeReturnedRefusal(t *testing.T) {
	t.Parallel()
	v4EscapeRefusal(t, v4EscapeDeclarations+`
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>'producer'};
function probe(base:Base):(value:string)=>string { const view=base as Target; return view.run; }
console.log(probe(raw)('bad'));`, "producer\n")
}

func TestV4EscapeMixedProducerRefusal(t *testing.T) {
	t.Parallel()
	v4EscapeRefusal(t, v4EscapeDeclarations+`
const good={kind:'Receiver' as const,run:(value:string):string=>value};
const bad={kind:'Receiver' as const,run:(value:'ok'):string=>'producer'};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log('read'); }
probe(good); probe(bad);`, "read\nread\n")
}

func TestV4EscapeCompatible(t *testing.T) {
	t.Parallel()
	source := v4EscapeDeclarations + `
const raw={kind:'Receiver' as const,run:(value:string):string=>value};
function probe(base:Base):void { const view=base as Target; const first=view.run; const second=view.run; console.log(first===second?'same':'different'); console.log(first('ok')); console.log(second('again')); }
probe(raw);`
	path := filepath.Join(t.TempDir(), "proven.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	escapes := 0
	viewTypeID := 0
	inspect := func(expression ir.Expression) ir.Expression {
		if property, ok := expression.(ir.Property); ok && property.ViewEscape && property.Name == "run" {
			escapes++
			if viewTypeID == 0 {
				viewTypeID = property.ViewTypeID
			}
			if viewTypeID == 0 || property.ViewTypeID != viewTypeID {
				t.Fatal("same view type lost its cache key")
			}
			if !property.ViewEscapeAdamic || property.ViewEscapeContract == 0 {
				t.Fatalf("missing escape metadata: %#v", property)
			}
			contract := program.ViewContracts[property.ViewEscapeContract-1]
			if len(contract.Parameters) != 1 || program.ViewContracts[contract.Parameters[0]-1].Name != "string" || program.ViewContracts[contract.Result-1].Name != "string" {
				t.Fatalf("lost invocation contract: %#v", contract)
			}
		}
		return expression
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), inspect)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), inspect)
	if escapes != 2 {
		t.Fatalf("got %d escaping reads", escapes)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "same\nok\nagain\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node: %#v", truth)
	}
	sanitized, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %s", backend, difference)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestV4EscapeTypeScriptBoundary(t *testing.T) {
	t.Parallel()
	// Temporary .ts input exercises the policy distinction; repository programs remain .a.
	path := filepath.Join(t.TempDir(), "escape.ts")
	source := v4EscapeDeclarations + `
const raw={kind:'Receiver' as const,run:(value:string):string=>value};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log('read'); }
probe(raw);`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "read\n" {
		t.Fatalf("Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err == nil {
		v4EscapeAdmitted(t, program, truth)
	}
	var pending *lower.NotYet
	if !errors.As(err, &pending) || !strings.Contains(pending.Where, "escape.ts:") || !strings.Contains(pending.What, "weak adapter cache per underlying function and view type in internal/native/runtime/closure.c") {
		t.Fatalf("adapter must remain explicitly unavailable: %v", err)
	}
}

func TestV4EscapeOrdinaryRead(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, v4EscapeDeclarations+`
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>value};
const base:Base=raw; const view=base as Target;
const ordinary={run:(value:string):string=>value};
const escaped=ordinary.run; console.log(escaped('ok'));`, "ok\n", "")
}

func TestV4EscapeAdapterArgument(t *testing.T) {
	t.Parallel()
	t.Skip("awaits compiler/views-v4: weak adapter cache per underlying function and view type, then escaped argument checks and their omission mutant")
}

func TestV4EscapeAdapterResult(t *testing.T) {
	t.Parallel()
	t.Skip("awaits compiler/views-v4: weak adapter cache per underlying function and view type, then escaped result checks and their omission mutant")
}

func TestV4EscapeAdapterIdentity(t *testing.T) {
	t.Parallel()
	t.Skip("awaits compiler/views-v4: weak adapter cache and underlying identity for ===, Object.is, Map and Set keys, with unequal-read mutant")
}

func TestV4EscapeAdapterNoStacking(t *testing.T) {
	t.Parallel()
	t.Skip("awaits compiler/views-v4: adapter unwrapping or composition and bounded allocation and retain counts, with unconditional-wrap mutant")
}
