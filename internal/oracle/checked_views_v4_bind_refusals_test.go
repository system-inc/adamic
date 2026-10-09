package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func v4BoundSurfaceRefusal(t *testing.T, setup, invocation, operation, fix string) {
	t.Helper()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(first:string,second:string)=>string}
const raw={kind:'Receiver' as const,run:(first:'ok',second:'later'):string=>{console.log('producer');return first+':'+second;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;const bound=escaped.bind(undefined);` + setup + `console.log(` + invocation + `);}
probe(raw);`
	path := filepath.Join(t.TempDir(), "bound-surface.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "producer\nok:later\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node truth: %#v", truth)
	}
	_, err := lowered(t, path)
	var refusal *lower.Refused
	what := operation + " on a bound function"
	if operation == "bind" {
		what = "binding a bound function again"
	}
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "bound-surface.ts:") || refusal.What != what || refusal.Fix != fix {
		t.Fatalf("compile-time bound refusal with path and fix: %v", err)
	}
	t.Logf("Node succeeds; compiler refuses: %v", err)
}

func TestV4EscapeAdapterBoundCallRefusal(t *testing.T) {
	t.Parallel()
	v4BoundSurfaceRefusal(t, "", "bound.call(undefined,'ok','later')", "call", "call the bound function directly")
}
func TestV4EscapeAdapterBoundApplyRefusal(t *testing.T) {
	t.Parallel()
	v4BoundSurfaceRefusal(t, "", "bound.apply(undefined,['ok','later'])", "apply", "call the bound function directly")
}
func TestV4EscapeAdapterBoundBindRefusal(t *testing.T) {
	t.Parallel()
	v4BoundSurfaceRefusal(t, "", "bound.bind(undefined,'ok','later')()", "bind", "bind once")
}
func TestV4EscapeAdapterBoundAliasCallRefusal(t *testing.T) {
	t.Parallel()
	v4BoundSurfaceRefusal(t, "const alias:(first:string,second:string)=>string=bound;", "alias.call(undefined,'ok','later')", "call", "call the bound function directly")
}
func TestV4EscapeAdapterBoundHelperApplyRefusal(t *testing.T) {
	t.Parallel()
	v4BoundSurfaceRefusal(t, "function invoke(value:(first:string,second:string)=>string):string {return value.apply(undefined,['ok','later']);}", "invoke(bound)", "apply", "call the bound function directly")
}
func TestV4EscapeAdapterBoundHelperBindRefusal(t *testing.T) {
	t.Parallel()
	v4BoundSurfaceRefusal(t, "function invoke(value:(first:string,second:string)=>string):string {return value.bind(undefined,'ok','later')();}", "invoke(bound)", "bind", "bind once")
}

func TestV4EscapeAdapterBoundViewRefusal(t *testing.T) {
	t.Parallel()
	source := `interface Base {readonly kind:'Receiver'}
interface Target extends Base {readonly run:(first:string,second:string)=>string}
interface BoundTarget extends Base {readonly run:(second:string)=>string}
const raw={kind:'Receiver' as const,run:(first:string,second:string):string=>first+':'+second};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;const bound=escaped.bind(undefined,'ok');const holder={kind:'Receiver' as const,run:bound};const broad:Base=holder;const again=broad as BoundTarget;const next=again.run;console.log(next('later'));}
probe(raw);`
	for _, extension := range []string{".ts", ".a"} {
		path := filepath.Join(t.TempDir(), "bound-view"+extension)
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		truth := onNode(t, path)
		if truth.exitCode != 0 || string(truth.stdout) != "ok:later\n" {
			t.Fatalf("Node bound-view truth: %#v", truth)
		}
		_, err := lowered(t, path)
		var refusal *lower.Refused
		what, fix := "an escaping callable read of again.run with an unsupported value contract", "prove the callable relation or use a callable with runtime-checkable parameter and result types"
		if extension == ".a" {
			what, fix = "an unproven escaping callable read of again.run", "prove the producer parameter and result relation before this read, or call the member directly through the view"
		}
		if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "bound-view"+extension+":5:") || refusal.What != what || refusal.Fix != fix {
			t.Fatalf("bound re-view must not invent producer metadata: %v", err)
		}
		t.Logf("bound re-view %s refused with path and fix: %v", extension, err)
	}
}
