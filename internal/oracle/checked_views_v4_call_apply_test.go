package oracle

import (
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"strings"
	"testing"
)

func v4SurfaceInvocationCheck(t *testing.T, operation, relation string) {
	t.Helper()
	producer, target := "(value:'ok'):string=>{console.log('producer');return 'producer'}", "(value:string)=>string"
	checkedOut, nodeOut := "read\nreceiver\nargument\n", "read\nreceiver\nargument\nproducer\nproducer\n"
	if relation == "result" {
		producer, target = "(value:string):string=>{console.log('producer');return 'bad'}", "(value:string)=>'ok'"
		checkedOut, nodeOut = "read\nreceiver\nargument\nproducer\n", "read\nreceiver\nargument\nproducer\nbad\n"
	}
	arguments := "argument()"
	if operation == "apply" {
		arguments = "[argument()]"
	}
	source := "interface Base {readonly kind:'Receiver'}\ninterface Target extends Base {readonly run:" + target + "}\nconst raw={kind:'Receiver' as const,run:" + producer + "};\nfunction receiver():{readonly label:string} {console.log('receiver');return {label:'this'};}\nfunction argument():string {console.log('argument');return 'bad'}\nfunction probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log('read');console.log(escaped." + operation + "(receiver()," + arguments + "));}\nprobe(raw);"
	program, truth := v4IdentityProgram(t, source, nodeOut)
	sanitized, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != checkedOut || !strings.Contains(string(got.stderr), "callable call failed: view.run at escaping call (read at ") || !strings.Contains(string(got.stderr), relation) || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s %s %s check/order: %#v", name, operation, relation, got)
		}
		t.Logf("%s %s %s: exit70 stdout=%q stderr=%q", name, operation, relation, got.stdout, got.stderr)
	}
	code := native.C(program)
	before := "adamic_closure_receiver_call(value, receiver, arguments, count, slots)"
	if strings.Count(code, before) != 1 {
		t.Fatal("native surface omission site moved")
	}
	got, _ := v4IdentityCounted(t, strings.Replace(code, before, "adamic_closure_receiver_call(adamic_view_adapter_underlying(value), receiver, arguments, count, slots)", 1))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("native omission must run cleanly like Node: %s", difference)
	}
	t.Logf("native/ASan/UBSan/LSan %s %s omission caught: %q", operation, relation, got.stdout)
	code = javascript.JavaScript(program)
	before = "if (adamicViewAdapterRoots.has(value)) return value.code(value, arguments_, receiver);"
	if strings.Count(code, before) != 1 {
		t.Fatal("JavaScript surface omission site moved")
	}
	got = v4IdentityJavaScript(t, strings.Replace(code, before, "value = adamicViewAdapterUnderlying(value); /* mutant: skip surface checks */", 1))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("JavaScript omission must match Node: %s", difference)
	}
	t.Logf("JavaScript %s %s omission caught: %q", operation, relation, got.stdout)
}

func TestV4EscapeAdapterCallArgument(t *testing.T) {
	t.Parallel()
	v4SurfaceInvocationCheck(t, "call", "argument")
}
func TestV4EscapeAdapterCallResult(t *testing.T) {
	t.Parallel()
	v4SurfaceInvocationCheck(t, "call", "result")
}
func TestV4EscapeAdapterApplyArgument(t *testing.T) {
	t.Parallel()
	v4SurfaceInvocationCheck(t, "apply", "argument")
}
func TestV4EscapeAdapterApplyResult(t *testing.T) {
	t.Parallel()
	v4SurfaceInvocationCheck(t, "apply", "result")
}

func TestV4EscapeAdapterCallApplyReceiver(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(this:{readonly label:string},value:string)=>string}
const raw={kind:'Receiver' as const,run:function(this:{readonly label:string},value:'ok'):string {console.log(this.label);return value+'!';}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped.call({label:'call'},'ok'));console.log(escaped.apply({label:'apply'},['ok']));}
probe(raw);`
	v4EscapeSurface(t, source, "call\nok!\napply\nok!\n", false)
}

func TestV4EscapeAdapterCallApplyCount(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(value?:string)=>string}
const raw={kind:'Receiver' as const,run:function(value?:string):string {return ` + "`${arguments.length}:${value??'missing'}`" + `;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped.call(undefined));console.log(escaped.apply(undefined,[]));console.log(escaped.call(undefined,'ok'));console.log(escaped.apply(undefined,['ok']));}
probe(raw);`
	v4EscapeSurface(t, source, "0:missing\n0:missing\n1:ok\n1:ok\n", false)
}

func TestV4EscapeAdapterCallRefusal(t *testing.T) {
	t.Parallel()
	v4EscapeRefusal(t, v4EscapeDeclarations+`const raw={kind:'Receiver' as const,run:(value:'ok'):string=>'producer'};function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped.call(undefined,'bad'));}probe(raw);`, "producer\n")
}

func TestV4EscapeAdapterSurfaceSortCallback(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(left:number,right:number)=>number}
const raw={kind:'Receiver' as const,run:(left:number,right:number):number=>left-right};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;const values=[3,1,2];values.sort(escaped);console.log(values.join(','));}
probe(raw);`
	v4EscapeSurface(t, source, "1,2,3\n", false)
}

func TestV4EscapeAdapterSurfaceRegExpCallback(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(value:string)=>string}
const raw={kind:'Receiver' as const,run:(value:'x'):string=>'ok'};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log('x-x'.replace(/x/g,escaped));}
probe(raw);`
	v4EscapeSurface(t, source, "ok-ok\n", false)
}

func TestV4EscapeAdapterCallReceiverCheck(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(this:{readonly label:string},value:string)=>string}
const raw={kind:'Receiver' as const,run:function(this:{readonly label:'ok'},value:string):string {console.log(this.label);return value;}};
function argument():string {console.log('argument');return 'value';}
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log('read');console.log(escaped.call({label:'bad'},argument()));}
probe(raw);`
	program, truth := v4IdentityProgram(t, source, "read\nargument\nbad\nvalue\n")
	sanitized, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "read\nargument\n" || !strings.Contains(string(got.stderr), "thisArg expected producer") || !strings.Contains(string(got.stderr), ", view ") || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s receiver check: %#v", name, got)
		}
	}
	changed := 0
	for _, function := range program.Functions {
		if function.CallableReceiver != 0 {
			for _, field := range program.ViewContracts[function.CallableReceiver-1].Fields {
				if field.Name == "label" {
					program.ViewContracts[field.Contract-1].Allowed = nil
					changed++
				}
			}
		}
	}
	if changed != 1 {
		t.Fatalf("receiver omission changed %d domains", changed)
	}
	got, _ := v4IdentityCounted(t, native.C(program))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatal(difference)
	}
	got = onJavaScriptBackend(t, program)
	if difference := disagreement(truth, got); difference != "" {
		t.Fatal(difference)
	}
	t.Log("receiver-domain omission caught in native/ASan/UBSan/LSan and JavaScript")
}

func TestV4EscapeAdapterCallReceiverRefusal(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(this:{readonly label:string},value:string)=>string}
const raw={kind:'Receiver' as const,run:function(this:{readonly label:'ok'},value:string):string {return this.label;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped.call({label:'bad'},'value'));}
probe(raw);`
	v4EscapeRefusal(t, source, "bad\n")
}

func TestV4EscapeAdapterCallApplyZeroReceiver(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(this:{readonly label:string})=>string}
const raw={kind:'Receiver' as const,run:function(this:{readonly label:string}):string {return this.label;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped.call({label:'call'}));console.log(escaped.apply({label:'apply'},[]));}
probe(raw);`
	v4EscapeSurface(t, source, "call\napply\n", false)
}
