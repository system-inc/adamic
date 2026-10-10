package oracle

import (
	"errors"
	"fmt"
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
			t.Fatalf("control/admission differs in %s: %s; %#v", backend, difference, got)
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
	if err != nil {
		t.Fatal(err)
	}
	v4EscapeAdmitted(t, program, truth)
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
	v4EscapeAdapterCheck(t, "argument")
}

func TestV4EscapeAdapterResult(t *testing.T) {
	t.Parallel()
	v4EscapeAdapterCheck(t, "result")
}

func TestV4EscapeAdapterIdentity(t *testing.T) {
	t.Parallel()
	v4EscapeAdapterIdentity(t)
}

func TestV4EscapeAdapterNoStacking(t *testing.T) {
	t.Parallel()
	v4EscapeAdapterNoStacking(t)
}

// Source Node is the unchecked truth. The checked divergence is required only
// when the escaped value is invoked, after the observable argument effects.
func v4EscapeAdapterCheck(t *testing.T, relation string) {
	t.Helper()
	producer, target := "(value:'ok'):string=>{ console.log('producer'); return 'producer'; }", "(value:string)=>string"
	checkedOut, nodeOut := "read\nargument\n", "read\nargument\nproducer\nproducer\n"
	if relation == "result" {
		producer, target = "(value:string):string=>{ console.log('producer'); return 'bad'; }", "(value:string)=>'ok'"
		checkedOut, nodeOut = "read\nargument\nproducer\n", "read\nargument\nproducer\nbad\n"
	}
	source := "interface Base { readonly kind:'Receiver'; }\ninterface Target extends Base { readonly run:" + target + "; }\nconst raw={kind:'Receiver' as const,run:" + producer + "};\nfunction argument():string { console.log('argument'); return 'bad'; }\nfunction probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log('read'); console.log(escaped(argument())); }\nprobe(raw);"
	path := filepath.Join(t.TempDir(), "adapter.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != nodeOut || len(truth.stderr) != 0 {
		t.Fatalf("source Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	var read ir.Property
	callWhere := ""
	find := func(expression ir.Expression) ir.Expression {
		if p, ok := expression.(ir.Property); ok && p.ViewEscape && p.View != "" {
			read = p
		}
		if c, ok := expression.(ir.CallClosure); ok && c.CallWhere != "" {
			callWhere = c.CallWhere
		}
		return expression
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), find)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), find)
	if read.ViewEscapeContract == 0 || read.ViewEscapeAdamic {
		t.Fatalf("lost adapter demand: %#v", read)
	}
	message := fmt.Sprintf("callable call failed: %s at escaping call (read at %s) argument 1 expected producer \"ok\", view string", read.View, read.ViewWhere)
	if relation == "result" {
		message = fmt.Sprintf("callable call failed: %s at escaping call (read at %s) result expected view \"ok\", producer string", read.View, read.ViewWhere)
	}
	message += " call at " + callWhere
	want := run{exitCode: 70, stdout: []byte(checkedOut), stderr: []byte("adamic: panic: " + message + "\n")}
	sanitized, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: %s; %#v", name, difference, got)
		}
		t.Logf("%s %s check: exit=%d stdout=%q stderr=%q", name, relation, got.exitCode, got.stdout, got.stderr)
	}
	changed := 0
	if relation == "argument" {
		for _, f := range program.Functions {
			for _, id := range f.CallableParameters {
				if id != 0 && program.ViewContracts[id-1].Name == "\"ok\"" {
					program.ViewContracts[id-1].Allowed = nil
					changed++
				}
			}
		}
	} else {
		id := program.ViewContracts[read.ViewEscapeContract-1].Result
		program.ViewContracts[id-1].Allowed = nil
		changed++
	}
	if changed != 1 {
		t.Fatalf("omission changed %d domains", changed)
	}
	sanitized, binary := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if strings.Contains(string(got.stderr), "Sanitizer") || disagreement(truth, got) != "" || disagreement(want, got) == "" {
			t.Fatalf("%s check omission must run cleanly like source Node: %#v", name, got)
		}
		t.Logf("%s %s omission caught: exit=%d stdout=%q", name, relation, got.exitCode, got.stdout)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func v4EscapeSource(t *testing.T, source, output string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "control.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != output || len(truth.stderr) != 0 {
		t.Fatalf("Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	v4EscapeAdmitted(t, program, truth)
}

func TestV4EscapeAdapterUnusedMisfit(t *testing.T) {
	t.Parallel()
	v4EscapeSource(t, `interface Base { readonly kind:'Receiver'; }
interface Target extends Base { readonly run:(value:string)=>'ok'; }
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>'bad'};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log('read'); }
probe(raw);`, "read\n")
}

func TestV4EscapeAdapterReturned(t *testing.T) {
	t.Parallel()
	v4EscapeSource(t, v4EscapeDeclarations+`
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>value};
function probe(base:Base):(value:string)=>string { const view=base as Target; return view.run; }
function consume(callback:(value:string)=>string):void { console.log(callback('ok')); console.log(callback('ok')); }
consume(probe(raw));`, "ok\nok\n")
}

func TestV4EscapeAdapterBoxed(t *testing.T) {
	t.Parallel()
	v4EscapeSource(t, `interface Base { readonly kind:'Receiver'; }
interface Target extends Base { readonly run:(value:number|string)=>number|string; }
const raw={kind:'Receiver' as const,run:(value:number):number=>value+1};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log(`+"`${escaped(4)}`"+`); }
probe(raw);`, "5\n")
}

func TestV4EscapeAdapterCounted(t *testing.T) {
	t.Parallel()
	v4EscapeSource(t, `interface Base { readonly kind:'Receiver'; }
interface Target extends Base { readonly run:(value?:string)=>string; }
const raw={kind:'Receiver' as const,run:function(value?:string):string { return `+"`${arguments.length}:${value ?? 'absent'}`"+`; }};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log(escaped()); console.log(escaped('ok')); }
probe(raw);`, "0:absent\n1:ok\n")
}

func TestV4EscapeAdapterUncheckableRefusal(t *testing.T) {
	t.Parallel()
	source := `interface Base { readonly kind:'Receiver'; }
interface Link { readonly text:string; readonly next?:Link; }
interface Target extends Base { readonly run:(value:Link)=>string; }
const raw={kind:'Receiver' as const,run:(value:{readonly text:string}):string=>'ok'};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log('read'); }
probe(raw);`
	path := filepath.Join(t.TempDir(), "unsupported.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "read\n" {
		t.Fatalf("Node: %#v", truth)
	}
	_, err := lowered(t, path)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "unsupported.ts:5:") || !strings.Contains(refusal.What, "escaping callable read of view.run with an unsupported value contract") || !strings.Contains(refusal.Fix, "runtime-checkable parameter and result types") {
		t.Fatalf("unsupported relation needs path and fix: %v", err)
	}
}

func TestV4EscapeAdapterNull(t *testing.T) {
	t.Parallel()
	v4EscapeSource(t, `interface Base { readonly kind:'Receiver'; }
interface Target extends Base { readonly run:(value:null)=>null; }
const raw={kind:'Receiver' as const,run:(value:null):null=>value};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; console.log(`+"`${escaped(null)===null}`"+`); }
probe(raw);`, "true\n")
}

func TestV4EscapeAdapterVoid(t *testing.T) {
	t.Parallel()
	v4EscapeSource(t, `interface Base { readonly kind:'Receiver'; }
interface Target extends Base { readonly run:(value:string)=>void; }
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>value+'!'};
function probe(base:Base):void { const view=base as Target; const escaped=view.run; escaped('ok'); console.log('done'); }
probe(raw);`, "done\n")
}
