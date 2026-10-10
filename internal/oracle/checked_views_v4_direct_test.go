package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Each source control is run on Node; runtime negatives are the deliberate
// checked-view divergence, pinned to exit 70 after their observable arguments.
func v4Direct(t *testing.T, producer, target, argument, nodeOut, checkedOut, relation string, mutant bool) {
	t.Helper()
	source := "interface Base { readonly kind: 'Receiver'; }\ninterface Target extends Base { readonly run: " + target + "; }\nfunction argument(): string { console.log('argument'); return " + argument + "; }\nfunction probe(base: Base): void { const view=base as Target; console.log(view.run(argument())); }\nconst raw={kind: 'Receiver' as const, run: " + producer + "}; probe(raw);\n"
	path := filepath.Join(t.TempDir(), "direct.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != nodeOut || len(truth.stderr) != 0 {
		t.Fatalf("Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	var call ir.CallClosure
	find := func(value ir.Expression) ir.Expression {
		if c, ok := value.(ir.CallClosure); ok && c.CallContract != 0 {
			if p, ok := c.Closure.(ir.Property); ok && p.Name == "run" {
				call = c
			}
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), find)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), find)
	if call.CallContract == 0 {
		t.Fatal("no checked direct call")
	}
	p := call.Closure.(ir.Property)
	message := ""
	if relation == "argument" {
		message = fmt.Sprintf("callable call failed: %s at %s argument 1 expected producer \"ok\", view string", p.View, call.CallWhere)
	}
	if relation == "result" {
		message = fmt.Sprintf("callable call failed: %s at %s result expected view \"ok\", producer string", p.View, call.CallWhere)
	}
	want := run{stdout: []byte(checkedOut)}
	if relation != "" {
		want.exitCode = 70
		want.stderr = []byte("adamic: panic: " + message + "\n")
	}
	sanitized, binary := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: %s; %#v", name, difference, got)
		}
	}
	if relation == "" {
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
	if !mutant {
		return
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
		id := program.ViewContracts[call.CallContract-1].Result
		program.ViewContracts[id-1].Allowed = nil
		changed++
	}
	if changed != 1 {
		t.Fatalf("mutant changed %d domains", changed)
	}
	gotSanitized, mutantBinary := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": gotSanitized, "javascript": onJavaScriptBackend(t, program)} {
		if strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("mutant requires semantic kill: %#v", got)
		}
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s mutant must match Node: %s", name, difference)
		}
		if disagreement(want, got) == "" {
			t.Fatal("check omission survived")
		}
		t.Logf("%s %s check omission caught: exit %d output %q", name, relation, got.exitCode, got.stdout)
	}
	if report := leaksUncached(t, program, mutantBinary); report != "" {
		t.Fatal(report)
	}
}

func TestV4DirectArgument(t *testing.T) {
	t.Parallel()
	v4Direct(t, "(value: 'ok'): string => 'producer'", "(value: string) => string", "'bad'", "argument\nproducer\n", "argument\n", "argument", true)
}
func TestV4DirectResult(t *testing.T) {
	t.Parallel()
	v4Direct(t, "(value: string): string => 'bad'", "(value: string) => 'ok'", "'ok'", "argument\nbad\n", "argument\n", "result", true)
}
func TestV4DirectCompatible(t *testing.T) {
	t.Parallel()
	v4Direct(t, "(value: 'ok'): string => value", "(value: string) => string", "'ok'", "argument\nok\n", "argument\nok\n", "", false)
}

func v4DirectSource(t *testing.T, source, stdout, failure string, nodeStdout ...string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "boundary.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	expectedNode := stdout
	if len(nodeStdout) != 0 {
		expectedNode = nodeStdout[0]
	}
	if strings.Contains(failure, "has no checkable producer signature") {
		if truth.exitCode == 0 || string(truth.stdout) != expectedNode || !strings.Contains(string(truth.stderr), "TypeError") {
			t.Fatalf("Node callee failure: %#v", truth)
		}
	} else if truth.exitCode != 0 || string(truth.stdout) != expectedNode || len(truth.stderr) != 0 {
		t.Fatalf("Node: %#v", truth)
	}
	t.Logf("source Node: exit=%d stdout=%q", truth.exitCode, truth.stdout)
	if failure == "" && (truth.exitCode != 0 || string(truth.stdout) != stdout) {
		t.Fatalf("Node: %#v", truth)
	}
	want := run{stdout: []byte(stdout)}
	sanitized, binary := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if failure == "" {
			if difference := disagreement(want, got); difference != "" {
				t.Fatalf("%s: %s; %#v", name, difference, got)
			}
		} else {
			if got.exitCode != 70 || string(got.stdout) != stdout || !strings.Contains(string(got.stderr), failure) {
				t.Fatalf("%s: exit=%d stdout=%q stderr=%q", name, got.exitCode, got.stdout, got.stderr)
			}
		}
	}
	if failure == "" {
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
	if strings.HasPrefix(failure, "argument 1 expected producer") {
		changed := 0
		var erase func(ir.ViewContractID)
		erase = func(id ir.ViewContractID) {
			if id == 0 {
				return
			}
			c := &program.ViewContracts[id-1]
			if c.Name == "\"ok\"" && len(c.Allowed) != 0 {
				c.Allowed = nil
				changed++
			}
			for _, field := range c.Fields {
				erase(field.Contract)
			}
			erase(c.Element)
		}
		for _, f := range program.Functions {
			for _, id := range f.CallableParameters {
				erase(id)
			}
		}
		if changed != 1 {
			t.Fatalf("aggregate boundary mutant changed %d domains", changed)
		}
		sanitized, binary := nativelyUncached(t, program)
		for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("%s mutant must execute like Node: %s", name, difference)
			}
			t.Logf("%s aggregate argument check omission caught: exit %d output %q; expected exit 70", name, got.exitCode, got.stdout)
		}
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
}

func TestV4DirectMethod(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { run(value: string): string; }
class Producer { readonly kind='Receiver' as const; run(value: 'ok'): string { return value; } }
function probe(base:Base):void { const view=base as Target; console.log(view.run('ok')); }
probe(new Producer());`, "ok\n", "")
}

func TestV4DirectObjectArgument(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:(value:{readonly text:string})=>string; }
const raw={kind:'Receiver' as const,run:(value:{readonly text:'ok'}):string=>'producer'};
function probe(base:Base):void { const view=base as Target; console.log(view.run({text:'bad'})); }
probe(raw);`, "", "argument 1 expected producer { readonly text: \"ok\"; }, view { readonly text: string; }", "producer\n")
}

func TestV4DirectArrayArgument(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:(value:readonly string[])=>string; }
const raw={kind:'Receiver' as const,run:(value:readonly 'ok'[]):string=>'producer'};
function probe(base:Base):void { const view=base as Target; console.log(view.run(['bad'])); }
probe(raw);`, "", "argument 1 expected producer readonly \"ok\"[], view readonly string[]", "producer\n")
}

func TestV4DirectCalleeOrder(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:(value:string)=>string; }
const raw={kind:'Receiver' as const,run:7};
function argument():string { console.log('argument');return 'ok'; }
function probe(base:Base):void { const view=base as Target; console.log(view.run(argument())); }
probe(raw);`, "argument\n", "has no checkable producer signature")
}

func TestV4DirectBoxed(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:(value:number|string)=>number|string; }
function producer(value:number):number { return value+1; }
const raw={kind:'Receiver' as const,run:producer};
function probe(base:Base):void { const view=base as Target; console.log(`+"`${view.run(4)}`"+`); }
probe(raw);`, "5\n", "")
}

func TestV4DirectUnused(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:(value:string)=>'ok'; }
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>'bad'};
function probe(base:Base):void { const view=base as Target; console.log(view.kind); }
probe(raw);`, "Receiver\n", "")
}

func TestV4DirectOptionalReceiver(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:(value:string)=>string; }
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>value};
function argument():string { console.log('argument');return 'ok'; }
function call(view:Target|undefined):void { console.log(`+"`${view?.run(argument())}`"+`); }
function probe(base:Base):void { call(base as Target); } call(undefined);probe(raw);`, "undefined\nargument\nok\n", "")
}

func TestV4DirectVoid(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:(value:string)=>void; }
const raw={kind:'Receiver' as const,run:(value:'ok'):string=>value};
function probe(base:Base):void { const view=base as Target; view.run('ok'); console.log('done'); }
probe(raw);`, "done\n", "")
}

func TestV4DirectNull(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:(value:null)=>null; }
const raw={kind:'Receiver' as const,run:(value:null):null=>value};
function probe(base:Base):void { const view=base as Target; console.log(`+"`${view.run(null)===null}`"+`); }
probe(raw);`, "true\n", "")
}

func TestV4DirectUnknownProducerParameter(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:(value:string)=>string; }
const raw={kind:'Receiver' as const,run:(value:unknown):string=>'ok'};
function probe(base:Base):void { const view=base as Target; console.log(view.run('ok')); }
probe(raw);`, "ok\n", "")
}

func TestV4DirectClassResult(t *testing.T) {
	t.Parallel()
	v4DirectSource(t, `interface Base { readonly kind: 'Receiver'; }
interface Target extends Base { readonly run:()=>{readonly text:string}; }
class Result { readonly text='ok'; }
const raw={kind:'Receiver' as const,run:():Result=>new Result()};
function probe(base:Base):void { const view=base as Target; console.log(view.run().text); }
probe(raw);`, "ok\n", "")
}

func TestV4DirectRecursiveRefusal(t *testing.T) {
	t.Parallel()
	source := `interface Base { readonly kind: 'Receiver'; }
interface Link { readonly text:string; readonly next?:Link; }
interface Target extends Base { readonly run:(value:Link)=>string; }
const raw={kind:'Receiver' as const,run:(value:{readonly text:string}):string=>'ok'};
function probe(base:Base):void { const view=base as Target; console.log(view.run({text:'ok'})); }
probe(raw);`
	path := filepath.Join(t.TempDir(), "recursive.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "ok\n" {
		t.Fatalf("Node: %#v", truth)
	}
	_, err := lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "recursive.a:5:") || !strings.Contains(err.Error(), "checked view call to member run with an unsupported value contract") || !strings.Contains(err.Error(), "prove the callable relation") {
		t.Fatalf("refusal with path and fix: %v", err)
	}
}

func TestV4DirectProducerRefusal(t *testing.T) {
	t.Parallel()
	source := `interface Base { readonly kind: 'Receiver'; }
interface Link { readonly text:string; readonly next?:Link; }
interface Target extends Base { readonly run:(value:{readonly text:string})=>string; }
const raw={kind:'Receiver' as const,run:(value:Link):string=>'ok'};
function probe(base:Base):void { const view=base as Target; console.log(view.run({text:'ok'})); }
probe(raw);`
	path := filepath.Join(t.TempDir(), "producer.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "ok\n" {
		t.Fatalf("Node: %#v", truth)
	}
	_, err := lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "producer.a:5:") || !strings.Contains(err.Error(), "checked view call to member run with an unsupported value contract") || !strings.Contains(err.Error(), "prove the callable relation") {
		t.Fatalf("refusal with path and fix: %v", err)
	}
}
