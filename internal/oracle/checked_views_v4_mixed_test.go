package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

const v4MixedDeclarations = `interface Payload {readonly value:number;}
interface Base {readonly run:unknown;}
interface Target {readonly run:(value:string|Payload)=>number;}
`

func TestV4MixedCallableFieldBoxing(t *testing.T) {
	t.Parallel()
	source := v4MixedDeclarations + `function probe(base:Base):void {const view=base as Target;const holder:{readonly arg:string|Payload}={arg:'ok'};console.log(` + "`${view.run(holder.arg)}`" + `);}
probe({run:(value:string|Payload):number=>typeof value==='string'?value.length:value.value});`
	v4EscapeSource(t, source, "2\n")
}

func TestV4MixedCallableFieldArgument(t *testing.T) {
	t.Parallel()
	source := v4MixedDeclarations + `function probe(base:Base):void {const view=base as Target;const holder:{readonly arg:string|Payload}={arg:'ok'};console.log(` + "`${view.run(holder.arg)}`" + `);}
probe({run:(value:Payload):number=>7});`
	v4Lane5Negative(t, source, "7\n", "argument 1", "view.run")
	program, truth := v4IdentityProgram(t, source, "7\n")
	changed := 0
	for _, function := range program.Functions {
		if function.Closure && len(function.CallableParameters) == 1 {
			id := function.CallableParameters[0]
			if id != 0 && program.ViewContracts[id-1].Name == "Payload" {
				program.ViewContracts[id-1] = ir.ViewContract{Kind: ir.ViewUnknown, Of: ir.Union, Name: "unknown"}
				changed++
			}
		}
	}
	if changed != 1 {
		t.Fatalf("producer argument omission count: %d", changed)
	}
	v4EscapeAdmitted(t, program, truth)
	t.Log("mixed aggregate producer-domain omission caught against Node in both backends with sanitizers")
}

func TestV4MixedLongFieldDiagnostic(t *testing.T) {
	t.Parallel()
	name := "Aggregate" + strings.Repeat("Payload", 40)
	source := "interface " + name + " {readonly value:number;}\n" + `interface Base {readonly kind:'base';}
interface Target extends Base {readonly payload:` + name + `;}
function probe(base:Base):void {const view=base as Target;console.log(` + "`${view.payload.value}`" + `);}
const raw={kind:'base' as const,payload:7};probe(raw);`
	program, _ := v4IdentityProgram(t, source, "undefined\n")
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "field read failed") || !strings.Contains(string(got.stderr), name) || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s long diagnostic: exit %d stdout %q stderr %s", backend, got.exitCode, got.stdout, got.stderr)
		}
	}
}
