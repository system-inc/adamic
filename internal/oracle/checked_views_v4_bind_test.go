package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"strings"
	"testing"
)

func TestV4EscapeAdapterBindCompatible(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(first:string,second:string)=>string}
const raw={kind:'Receiver' as const,run:(first:'ok',second:'later'):string=>{console.log('producer');return first+':'+second;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;const bound=escaped.bind(undefined,'ok');console.log('bound');console.log(bound('later'));console.log(bound('later'));}
probe(raw);`
	v4EscapeSurface(t, source, "bound\nproducer\nok:later\nproducer\nok:later\n", false)
}

func TestV4EscapeAdapterBindPrefix(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(first:string,second:string)=>string}
const raw={kind:'Receiver' as const,run:(first:'ok',second:'later'):string=>{console.log('producer');return first+second;}};
function receiver():{readonly label:string} {console.log('receiver');return {label:'this'};}
function first():string {console.log('first');return 'bad';}
function second():string {console.log('second');return 'later';}
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log('read');const bound=escaped.bind(receiver(),first(),second());console.log('bound');}
probe(raw);`
	program, truth := v4IdentityProgram(t, source, "read\nreceiver\nfirst\nsecond\nbound\n")
	sanitized, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "read\nreceiver\nfirst\nsecond\n" || !strings.Contains(string(got.stderr), "argument 1 expected producer") || !strings.Contains(string(got.stderr), "view string") || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s bind prefix/order: %#v", name, got)
		}
		t.Logf("%s prefix check: %q", name, got.stderr)
	}
	changed := 0
	for i := range program.Functions {
		for j, statement := range program.Functions[i].Body {
			if evaluate, ok := statement.(ir.Evaluate); ok {
				if call, ok := evaluate.Value.(ir.CallClosure); ok && call.CheckBound {
					program.Functions[i].Body[j] = ir.Evaluate{Value: ir.NumberConstant{Value: 0}}
					changed++
				}
			}
		}
	}
	if changed != 1 {
		t.Fatalf("prefix omission changed %d checks", changed)
	}
	v4EscapeAdmitted(t, program, truth)
	t.Log("prefix omission caught in both backends, native ASan/UBSan/LeakSanitizer clean; producer never invoked")
}

func v4BindLaterCheck(t *testing.T, relation string) {
	t.Helper()
	producer, target := "(first:'ok',second:'later'):string=>{console.log('producer');return 'value'}", "(first:string,second:string)=>string"
	checkedOut, nodeOut := "bound\nargument\n", "bound\nargument\nproducer\nvalue\n"
	if relation == "result" {
		producer, target = "(first:'ok',second:string):string=>{console.log('producer');return 'bad'}", "(first:string,second:string)=>'ok'"
		checkedOut, nodeOut = "bound\nargument\nproducer\n", "bound\nargument\nproducer\nbad\n"
	}
	source := "interface Base {readonly kind:'Receiver'}\ninterface Target extends Base {readonly run:" + target + "}\nconst raw={kind:'Receiver' as const,run:" + producer + "};\nfunction argument():string {console.log('argument');return 'bad'}\nfunction probe(base:Base):void {const view=base as Target;const escaped=view.run;const bound=escaped.bind(undefined,'ok');console.log('bound');console.log(bound(argument()));}\nprobe(raw);"
	program, truth := v4IdentityProgram(t, source, nodeOut)
	sanitized, _ := nativelyUncached(t, program)
	for name, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != checkedOut || !strings.Contains(string(got.stderr), relation) || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s later %s check: %#v", name, relation, got)
		}
	}
	changed := 0
	if relation == "argument" {
		for _, function := range program.Functions {
			for _, id := range function.CallableParameters {
				if id != 0 && program.ViewContracts[id-1].Name == "\"later\"" {
					program.ViewContracts[id-1].Allowed = nil
					changed++
				}
			}
		}
	} else {
		inspect := func(expression ir.Expression) ir.Expression {
			if property, ok := expression.(ir.Property); ok && property.ViewEscape && !property.ViewEscapeAdamic {
				id := program.ViewContracts[property.ViewEscapeContract-1].Result
				program.ViewContracts[id-1].Allowed = nil
				changed++
			}
			return expression
		}
		mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), inspect)
		mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), inspect)
	}
	if changed != 1 {
		t.Fatalf("later omission changed %d domains", changed)
	}
	v4EscapeAdmitted(t, program, truth)
	t.Logf("later %s omission caught in native/ASan/UBSan/LeakSanitizer and JavaScript", relation)
}
func TestV4EscapeAdapterBindRemainingArgument(t *testing.T) {
	t.Parallel()
	v4BindLaterCheck(t, "argument")
}
func TestV4EscapeAdapterBindResult(t *testing.T) { t.Parallel(); v4BindLaterCheck(t, "result") }

func TestV4EscapeAdapterBindReceiverAndCapture(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(this:{readonly label:string},first:string,second:string)=>string}
const raw={kind:'Receiver' as const,run:function producer(this:{readonly label:string},first:'ok',second:'later'):string {return this.label+':'+first+':'+second;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;let receiver={label:'original'};let prefix='ok';const bound=escaped.bind(receiver,prefix);receiver={label:'changed'};prefix='changed';console.log(bound('later'));console.log(bound('later'));}
probe(raw);`
	v4EscapeSurface(t, source, "original:ok:later\noriginal:ok:later\n", false)
}

func TestV4EscapeAdapterBindCountAndSurface(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(first:string,second?:string,third?:string)=>string}
const raw={kind:'Receiver' as const,run:function producer(first:'ok',second?:string,third?:string):string {return ` + "`${arguments.length}:${first}:${second??'missing'}:${third??'missing'}`" + `;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;const bound=escaped.bind(undefined,'ok');const name=bound.name;console.log(name);console.log(` + "`${bound.length}`" + `);console.log(bound());console.log(bound('second'));console.log(bound('second','third'));}
probe(raw);`
	v4EscapeSurface(t, source, "bound producer\n2\n1:ok:missing:missing\n2:ok:second:missing\n3:ok:second:third\n", false)
}

func TestV4EscapeAdapterBindBoxed(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(first:number|string,second:number|string)=>string}
const raw={kind:'Receiver' as const,run:(first:1,second:'later'):string=>{console.log('producer');return ` + "`${first}:${second}`" + `;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;const bound=escaped.bind(undefined,1);console.log('bound');console.log(bound('later'));}
probe(raw);`
	v4EscapeSurface(t, source, "bound\nproducer\n1:later\n", false)
}

func TestV4EscapeAdapterBindAllArguments(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(first:string,second:string)=>string}
const raw={kind:'Receiver' as const,run:(first:'ok',second:'later'):string=>{console.log('producer');return first+':'+second;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;const bound=escaped.bind(undefined,'ok','later');console.log('bound');console.log(bound());}
probe(raw);`
	v4EscapeSurface(t, source, "bound\nproducer\nok:later\n", false)
}

func TestV4EscapeAdapterBindNameOwnership(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(first:string)=>string}
const raw={kind:'Receiver' as const,run:function producer(first:'ok'):string {console.log('producer');return first;}};
function show(value:(first:string)=>string):void {const name=value.name;console.log(name);console.log(name);}
function probe(base:Base):void {const view=base as Target;const escaped=view.run;const bound=escaped.bind(undefined);show(bound);}
probe(raw);`
	v4EscapeSurface(t, source, "bound producer\nbound producer\n", false)
}
