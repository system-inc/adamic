package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

const v4HigherDeclarations = `interface Base {readonly run:unknown;}
interface Target {readonly run:(callback:(value:'ok')=>string)=>string;}
`

func TestV4HigherCallbackArgument(t *testing.T) {
	t.Parallel()
	source := v4HigherDeclarations + `function probe(base:Base):void {const view=base as Target;console.log(view.run((value:'ok'):string=>value));}
probe({run:(callback:(value:string)=>string):string=>callback('bad')});`
	v4Lane5Negative(t, source, "bad\n", "argument 1", "view.run callback")
	program, truth := v4IdentityProgram(t, source, "bad\n")
	changed := 0
	for _, function := range program.Functions {
		if function.Closure && len(function.CallableParameters) == 1 {
			id := function.CallableParameters[0]
			if id != 0 && program.ViewContracts[id-1].Kind == ir.ViewCallable {
				program.ViewContracts[id-1] = ir.ViewContract{Kind: ir.ViewUnknown, Of: ir.Union, Name: "unknown"}
				changed++
			}
		}
	}
	if changed != 1 {
		t.Fatalf("unwrapped callback mutant sites: %d", changed)
	}
	v4EscapeAdmitted(t, program, truth)
	t.Log("passing the callback unwrapped is caught in both backends, ASan/UBSan/LSan clean against Node")
}

func TestV4HigherCallbackNeverCalled(t *testing.T) {
	t.Parallel()
	source := v4HigherDeclarations + `function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log('read');console.log(escaped((value:'ok'):string=>value));}
probe({run:(callback:(value:string)=>string):string=>'never'});`
	v4EscapeSource(t, source, "read\nnever\n")
}

func TestV4HigherCallbackResult(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly run:unknown;}
interface Target {readonly run:(callback:(value:string)=>string)=>string;}
function probe(base:Base):void {const view=base as Target;console.log(view.run((value:string):string=>'bad'));}
probe({run:(callback:(value:string)=>'ok'):string=>callback('ok')});`
	v4Lane5Negative(t, source, "bad\n", "result", "view.run callback")
	program, truth := v4IdentityProgram(t, source, "bad\n")
	changed := 0
	for _, function := range program.Functions {
		for _, id := range function.CallableParameters {
			if id != 0 && program.ViewContracts[id-1].Kind == ir.ViewCallable {
				result := program.ViewContracts[id-1].Result
				if program.ViewContracts[result-1].Name == "\"ok\"" {
					program.ViewContracts[result-1].Allowed = nil
					changed++
				}
			}
		}
	}
	if changed != 1 {
		t.Fatalf("callback result omission sites: %d", changed)
	}
	v4EscapeAdmitted(t, program, truth)

}

func TestV4HigherCallbackArgumentOrder(t *testing.T) {
	t.Parallel()
	source := v4HigherDeclarations + `function side():string {console.log('evaluated');return 'bad';}
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log('read');console.log(escaped((value:'ok'):string=>value));}
probe({run:(callback:(value:string)=>string):string=>{console.log('outer');return callback(side());}});`
	program, _ := v4IdentityProgram(t, source, "read\nouter\nevaluated\nbad\n")
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "read\nouter\nevaluated\n" || !strings.Contains(string(got.stderr), "argument 1 expected producer \"ok\", view string") || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s callback order: exit %d stdout %q stderr %s", backend, got.exitCode, got.stdout, got.stderr)
		}
	}
}

func TestV4HigherAdamicEscapeRefusal(t *testing.T) {
	t.Parallel()
	source := v4HigherDeclarations + `function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log('read');}
probe({run:(callback:(value:string)=>string):string=>'never'});`
	v4EscapeRefusal(t, source, "read\n")
}

func TestV4HigherReturnedCallable(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly run:unknown;}
interface Target {readonly run:()=>((value:string)=>string);}
function probe(base:Base):void {const view=base as Target;const callback=view.run();console.log(callback('bad'));}
probe({run:():((value:'ok')=>string)=>(value:'ok'):string=>value});`
	v4Lane5Negative(t, source, "bad\n", "argument 1", "view.run callback")
}
