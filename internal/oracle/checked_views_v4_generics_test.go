package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV4GenericProducerInstantiation(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(value:string)=>string}
function make<T extends string>(expected:T):(value:T)=>T {return (value:T):T=>{console.log('producer');return expected;};}
const raw={kind:'Receiver' as const,run:make<'ok'>('ok')};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log('read');console.log(escaped('bad'));}
probe(raw);`
	program, truth := v4IdentityProgram(t, source, "read\nproducer\nok\n")
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "read\n" || !strings.Contains(string(got.stderr), `argument 1 expected producer "ok", view string`) || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s instantiated producer domain: %#v", backend, got)
		}
		t.Logf("%s actual instantiation checked: %s", backend, got.stderr)
	}
	changed := 0
	for _, function := range program.Functions {
		if !function.Closure {
			continue
		}
		for _, id := range function.CallableParameters {
			if id != 0 && program.ViewContracts[id-1].Name == `"ok"` {
				program.ViewContracts[id-1].Allowed = nil
				changed++
			}
		}
	}
	if changed != 1 {
		t.Fatalf("instantiation omission changed %d domains", changed)
	}
	v4EscapeAdmitted(t, program, truth)
	t.Log("instantiation domain omission caught in both backends and sanitized native")
}

func TestV4GenericErasedReadRefusal(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:<T extends string>(value:T)=>T}
const raw={kind:'Receiver' as const,run:(value:string):string=>'bad'};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log('read');}
probe(raw);`
	// The generic promise has no runtime witness at this escaping read, in .ts too.
	path := v4GenericSource(t, source)
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "read\n" {
		t.Fatalf("Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err == nil {
		v4EscapeAdmitted(t, program, truth)
	}
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "erased.ts:4:") || !strings.Contains(refusal.What, "escaping callable read of view.run") || !strings.Contains(refusal.Fix, "runtime-checkable parameter and result types") {
		t.Fatalf("erased generic read requires path and fix: %v", err)
	}
	t.Logf("erased relation refused: %v", err)
}

func v4GenericSource(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "erased.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestV4GenericViewCallInstantiation(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:<T extends string>(value:T)=>T}
const raw={kind:'Receiver' as const,run:(value:string):string=>{console.log('producer');return 'bad';}};
function invoke(base:Base):string {const view=base as Target;return view.run<'ok'>('ok');}
console.log(invoke(raw));`
	program, truth := v4IdentityProgram(t, source, "producer\nbad\n")
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "producer\n" || !strings.Contains(string(got.stderr), `result expected view "ok", producer string`) || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s instantiated view result: %#v", backend, got)
		}
		t.Logf("%s view call's instantiation checked: %s", backend, got.stderr)
	}
	changed := 0
	for _, contract := range program.ViewContracts {
		if contract.Kind == ir.ViewCallable && contract.Result != 0 && program.ViewContracts[contract.Result-1].Name == `"ok"` {
			program.ViewContracts[contract.Result-1].Allowed = nil
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("view instantiation omission changed %d domains", changed)
	}
	v4EscapeAdmitted(t, program, truth)
}
