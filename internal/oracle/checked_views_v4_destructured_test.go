package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestV4DestructuredCallableArgument(t *testing.T) {
	t.Parallel()
	source := v4EscapeDeclarations + `const raw={kind:'Receiver' as const,run:(value:'ok'):string=>value};
function probe(base:Base):void {const view=base as Target;const {run}=view;console.log(run('bad'));}
probe(raw);`
	v4Lane5Negative(t, source, "bad\n", "argument 1", "run (field run)")
	program, truth := v4IdentityProgram(t, source, "bad\n")
	code := native.C(program)
	before := "adamic_closure_receiver_call(closure, receiver, arguments, count, slots)"
	if strings.Count(code, before) != 1 {
		t.Fatal("native adapter bypass site moved")
	}
	got, _ := v4IdentityCounted(t, strings.Replace(code, before, "adamic_closure_receiver_call(adamic_view_adapter_underlying(closure), receiver, arguments, count, slots)", 1))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatal(difference)
	}
	code = javascript.JavaScript(program)
	before = "const adamicCall = (closure, values) => closure.code(closure, values);"
	if strings.Count(code, before) != 1 {
		t.Fatal("JavaScript adapter bypass site moved")
	}
	got = v4IdentityJavaScript(t, strings.Replace(code, before, "const adamicCall = (closure, values) => { closure = adamicViewAdapterUnderlying(closure); return closure.code(closure, values); };", 1))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatal(difference)
	}
	t.Log("destructured adapter omission caught in native ASan/UBSan/LSan and JavaScript")
}

func TestV4DestructuredCallableReadOnly(t *testing.T) {
	t.Parallel()
	source := v4EscapeDeclarations + `const raw={kind:'Receiver' as const,run:(value:'ok'):string=>value};
function probe(base:Base):void {const view=base as Target;const {run}=view;console.log(run.name);}
probe(raw);`
	v4EscapeSource(t, source, "run\n")
}

func TestV4DestructuredCallableAdamicRefusal(t *testing.T) {
	t.Parallel()
	source := v4EscapeDeclarations + `const raw={kind:'Receiver' as const,run:(value:'ok'):string=>value};
function probe(base:Base):void {const view=base as Target;const {run}=view;console.log('read');}
probe(raw);`
	path := filepath.Join(t.TempDir(), "destructured.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "read\n" {
		t.Fatalf("Node: %#v", truth)
	}
	_, err := lowered(t, path)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "destructured.a:4:66") || refusal.What != "an unproven escaping callable read of run (field run)" || refusal.Fix != "prove the producer parameter and result relation before this read, or call the member directly through the view" {
		t.Fatalf("destructured refusal with path and fix: %v", err)
	}
}
