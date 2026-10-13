package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"strings"
	"testing"
)

func v4AdapterBlame(t *testing.T, relation, operation string) {
	t.Helper()
	producer, target := "(value:'ok'):string=>{console.log('producer');return 'producer';}", "(value:string)=>string"
	nodeOut, checkedOut := "read\nforwarded\nargument\nproducer\nproducer\n", "read\nforwarded\nargument\n"
	if relation == "result" {
		producer, target = "(value:string):string=>{console.log('producer');return 'bad';}", "(value:string)=>'ok'"
		nodeOut = "read\nforwarded\nargument\nproducer\nbad\n"
		checkedOut += "producer\n"
	}
	invocation := "callback(argument())"
	if operation == "call" {
		invocation = "callback.call(undefined,argument())"
	}
	if operation == "apply" {
		invocation = "callback.apply(undefined,[argument()])"
	}
	if operation == "bound" {
		invocation = "const bound=callback.bind(undefined);\nreturn bound(argument())"
	} else {
		invocation = "return " + invocation
	}
	source := "interface Base {readonly kind:'Receiver'}\ninterface Target extends Base {readonly run:" + target + "}\nconst raw={kind:'Receiver' as const,run:" + producer + "};\nfunction acquire(base:Base):" + target + " {\nconst view=base as Target;\nreturn view.run;\n}\nfunction forward(callback:" + target + "):" + target + " {console.log('forwarded');return callback;}\nfunction argument():string {console.log('argument');return 'bad';}\nfunction distant(callback:" + target + "):string {\n" + invocation + ";\n}\nconst escaped=acquire(raw);console.log('read');\nconst forwarded=forward(escaped);\nconsole.log(distant(forwarded));"
	program, truth := v4IdentityProgram(t, source, nodeOut)
	var read ir.Property
	callSite := ""
	inspect := func(expression ir.Expression) ir.Expression {
		if property, ok := expression.(ir.Property); ok && property.ViewEscape {
			read = property
		}
		if call, ok := expression.(ir.CallClosure); ok && !call.CheckBound && call.ArgumentCount == nil {
			if local, ok := program.ClosureTargets(call).Value.(ir.Read); ok && (program.Locals[local.Local].Name == "callback" || program.Locals[local.Local].Name == "bound") && call.CallWhere != "" {
				callSite = call.CallWhere
			}
		}
		return expression
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), inspect)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), inspect)
	if read.ViewWhere == "" || callSite == "" || read.ViewWhere == callSite {
		t.Fatal("lost distinct read/call source sites")
	}
	message := fmt.Sprintf("callable call failed: %s at escaping call (read at %s) argument 1 expected producer \"ok\", view string call at %s", read.View, read.ViewWhere, callSite)
	if relation == "result" {
		message = fmt.Sprintf("callable call failed: %s at escaping call (read at %s) result expected view \"ok\", producer string call at %s", read.View, read.ViewWhere, callSite)
	}
	want := run{exitCode: 70, stdout: []byte(checkedOut), stderr: []byte("adamic: panic: " + message + "\n")}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s two-site blame: %s; %#v", backend, difference, got)
		}
		t.Logf("%s %s/%s read=%s call=%s exit=70", backend, operation, relation, read.ViewWhere, callSite)
	}
	changed := 0
	omit := func(expression ir.Expression) ir.Expression {
		if property, ok := expression.(ir.Property); ok && property.ViewEscape {
			property.ViewWhere = ""
			changed++
			return property
		}
		return expression
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), omit)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), omit)
	if changed != 1 {
		t.Fatalf("read-site omission changed %d reads", changed)
	}
	sanitized, _ = nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != checkedOut || !strings.Contains(string(got.stderr), "call at "+callSite) || strings.Contains(string(got.stderr), read.ViewWhere) || strings.Contains(string(got.stderr), "Sanitizer") || disagreement(want, got) == "" {
			t.Fatalf("%s read-site omission must only lose blame: %#v", backend, got)
		}
		t.Logf("%s read-site omission caught; same exit/output, missing read site", backend)
	}
	if truth.exitCode != 0 {
		t.Fatal("unchecked Node witness failed")
	}
}

func TestV4EscapeAdapterBlameArgument(t *testing.T) {
	t.Parallel()
	v4AdapterBlame(t, "argument", "direct")
}
func TestV4EscapeAdapterBlameResult(t *testing.T) {
	t.Parallel()
	v4AdapterBlame(t, "result", "direct")
}
func TestV4EscapeAdapterBlameCall(t *testing.T)  { t.Parallel(); v4AdapterBlame(t, "argument", "call") }
func TestV4EscapeAdapterBlameApply(t *testing.T) { t.Parallel(); v4AdapterBlame(t, "result", "apply") }
func TestV4EscapeAdapterBlameBound(t *testing.T) {
	t.Parallel()
	v4AdapterBlame(t, "argument", "bound")
}

func TestV4EscapeAdapterBlameNestedResult(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(value:string)=>'ok'}
function acquire(base:Base):(value:string)=>'ok' {
const view=base as Target;
return view.run;
}
const good={kind:'Receiver' as const,run:(value:string):'ok'=>'ok'};
const inner=acquire(good);
const raw={kind:'Receiver' as const,run:(value:string):string=>{console.log(inner('ok'));return 'bad';}};
const escaped=acquire(raw);
function distant(callback:(value:string)=>'ok'):void {
console.log(callback('ok'));
}
distant(escaped);`
	program, _ := v4IdentityProgram(t, source, "ok\nbad\n")
	var read ir.Property
	callSite := ""
	inspect := func(expression ir.Expression) ir.Expression {
		if p, ok := expression.(ir.Property); ok && p.ViewEscape {
			read = p
		}
		if c, ok := expression.(ir.CallClosure); ok {
			if r, ok := program.ClosureTargets(c).Value.(ir.Read); ok && program.Locals[r.Local].Name == "callback" {
				callSite = c.CallWhere
			}
		}
		return expression
	}
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), inspect)
	message := fmt.Sprintf("adamic: panic: callable call failed: %s at escaping call (read at %s) result expected view \"ok\", producer string call at %s\n", read.View, read.ViewWhere, callSite)
	want := run{exitCode: 70, stdout: []byte("ok\n"), stderr: []byte(message)}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s nested call context was not restored: %s", backend, difference)
		}
	}
}

func TestV4EscapeAdapterBlameRuntimeCallback(t *testing.T) {
	t.Parallel()
	// Library callback IR has no source call position; do not invent one from the read site.
	t.Skip("awaits compiler/views-v4: source call sites on runtime callback IR")
}
