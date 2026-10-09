package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func v4Lane5Source(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "review/compiler/views-v4/lane5", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func v4Lane5Negative(t *testing.T, source, nodeOut, relation, member string) {
	t.Helper()
	program, truth := v4IdentityProgram(t, source, nodeOut)
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "callable call failed: "+member) || !strings.Contains(string(got.stderr), relation) || !strings.Contains(string(got.stderr), "view ") || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s first-call %s: %#v", backend, relation, got)
		}
		t.Logf("%s exit70 at first call: %s", backend, got.stderr)
	}
	if truth.exitCode != 0 {
		t.Fatal("misfit Node witness failed")
	}
}

func TestV4Lane5FactoryIdentifier(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "factory-3-wrong-overload")
	v4EscapeSource(t, source, "7\n")
	source = strings.Replace(source, `factory.createIdentifier("ok")`, `factory.createIdentifier("ok",1)`, 1)
	v4Lane5Negative(t, source, "7\n", "argument 2", "factory.createIdentifier")
}
func TestV4Lane5FactoryStringLiteral(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "factory-9-wrong-overload")
	v4EscapeSource(t, source, "7\n")
	source = strings.Replace(source, `factory.createStringLiteral("ok")`, `factory.createStringLiteral("ok",false,true)`, 1)
	v4Lane5Negative(t, source, "7\n", "argument 3", "factory.createStringLiteral")
}
func TestV4Lane5FactoryUniqueName(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "factory-24-wrong-overload")
	v4EscapeSource(t, source, "7\n")
	source = strings.Replace(source, `factory.createUniqueName("ok")`, `factory.createUniqueName("ok",1,"prefix")`, 1)
	v4Lane5Negative(t, source, "7\n", "argument 3", "factory.createUniqueName")
}
func TestV4Lane5FactoryImport(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "factory-186-wrong-overload")
	v4EscapeSource(t, source, "7\n")
	source = strings.Replace(source, `nodeFactory.createImportClause(false,undefined,undefined)`, `nodeFactory.createImportClause(1,undefined,undefined)`, 1)
	v4Lane5Negative(t, source, "7\n", "argument 1", "nodeFactory.createImportClause")
}
func TestV4Lane5FactoryYield(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "factory-189-wrong-overload")
	v4EscapeSource(t, source, "7\n")
	source = strings.Replace(source, `factory.createYieldExpression(undefined,undefined)`, `factory.createYieldExpression({value:1},{value:2})`, 1)
	v4Lane5Negative(t, source, "7\n", "argument 1", "factory.createYieldExpression")
}
func TestV4Lane5FactoryResult(t *testing.T) {
	t.Parallel()
	v4Lane5Negative(t, v4Lane5Source(t, "factory-3-wrong-result"), "read\n", "result", "factory.createIdentifier")
}
func TestV4Lane5FactoryClone(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "factory-12-wrong-producer")
	v4EscapeSource(t, source, "7\n")
	source = strings.Replace(source, `factory.cloneNode({value:7})`, `factory.cloneNode<{readonly value:number;readonly label:string}>({value:7,label:"ok"})`, 1)
	source = strings.Replace(source, `(node:Node):Node=>node`, `(node:Node):Node=>({value:7})`, 1)
	v4Lane5Negative(t, source, "7\n", "result", "factory.cloneNode")
}
func TestV4Lane5FactoryModifier(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "factory-174-wrong-producer")
	v4EscapeSource(t, source, "7\n")
	source = strings.Replace(source, `:ModifierToken<ModifierSyntaxKind>=>({value:7})`, `: {readonly value:string}=>({value:"bad"})`, 1)
	v4Lane5Negative(t, source, "bad\n", "result", "factory.createModifier")
}
func TestV4Lane5FactoryWiderWritePending(t *testing.T) {
	t.Parallel()
	t.Skip("awaits compiler/checked-wider-writes: checked write through a wider view")
}

func TestV4Lane5BoxedParameter(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "view-parameter-wrong")
	v4EscapeSource(t, source, "5\n")
	source = strings.Replace(source, `callback(5)`, `callback("bad")`, 1)
	v4Lane5Negative(t, source, "bad\n", "argument 1", "viewed.run")
}
func TestV4Lane5BoxedParameterMembers(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "view-parameter-wrong-members")
	v4EscapeSource(t, source, "5\n")
	source = strings.Replace(source, `callback(5)`, `callback("bad")`, 1)
	v4Lane5Negative(t, source, "bad\n", "argument 1", "viewed.run")
}
func TestV4Lane5BoxedResult(t *testing.T) {
	t.Parallel()
	v4Lane5Negative(t, v4Lane5Source(t, "view-result-wrong"), "true\n", "result", "viewed.run")
}
func TestV4Lane5BoxedResultMembers(t *testing.T) {
	t.Parallel()
	v4Lane5Negative(t, v4Lane5Source(t, "view-result-wrong-members"), "true\n", "result", "viewed.run")
}

func TestV4Lane5SkipAdapterMutant(t *testing.T) {
	t.Parallel()
	source := v4EscapeDeclarations + `const raw={kind:'Receiver' as const,run:(value:'ok'):string=>value};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;console.log(escaped('bad'));}
probe(raw);`
	v4Lane5Negative(t, source, "bad\n", "argument 1", "view.run")
	program, truth := v4IdentityProgram(t, source, "bad\n")
	code := native.C(program)
	before := "adamic_closure_receiver_call(closure, receiver, arguments, count, slots)"
	if strings.Count(code, before) != 1 {
		t.Fatal("native adapter bypass site moved")
	}
	got, _ := v4IdentityCounted(t, strings.Replace(code, before, "adamic_closure_receiver_call(adamic_view_adapter_underlying(closure), receiver, arguments, count, slots)", 1))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("native skip-adapter must run cleanly like Node: %s", difference)
	}
	code = javascript.JavaScript(program)
	before = "const adamicCall = (closure, values) => closure.code(closure, values);"
	if strings.Count(code, before) != 1 {
		t.Fatal("JavaScript adapter bypass site moved")
	}
	got = v4IdentityJavaScript(t, strings.Replace(code, before, "const adamicCall = (closure, values) => { closure = adamicViewAdapterUnderlying(closure); return closure.code(closure, values); };", 1))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("JavaScript skip-adapter must run cleanly like Node: %s", difference)
	}
	t.Log("skip-adapter mutant caught in native ASan/UBSan/LSan and JavaScript")
}

func TestV4Lane5BoxedAdamicRefusal(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "view-result-wrong")
	path := filepath.Join(t.TempDir(), "boxed.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "true\n" {
		t.Fatalf("Node: %#v", truth)
	}
	_, err := lowered(t, path)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "boxed.a:5:") || refusal.What != "an unproven escaping callable read of viewed.run" || refusal.Fix != "prove the producer parameter and result relation before this read, or call the member directly through the view" {
		t.Fatalf(".a keeps the unproven-read refusal: %v", err)
	}
	t.Logf("readonly unknown storage grants no .a signature proof: %v", err)
}
func TestV4Lane5FactoryWiderWriteRefusal(t *testing.T) {
	t.Parallel()
	source := v4Lane5Source(t, "factory-165-wrong-arity")
	path := filepath.Join(t.TempDir(), "wider-write.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "done\n" {
		t.Fatalf("Node: %#v", truth)
	}
	_, err := lowered(t, path)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "wider-write.ts:6:48") || !strings.Contains(refusal.What, "readonly field onEmitNode becomes writable") || refusal.Fix != "keep onEmitNode readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)" {
		t.Fatalf("wider-write keeps path and fix: %v", err)
	}
	t.Logf("rank165 refusal: %v", err)
}

func TestV4Lane5FactoryReadonlyTargetControl(t *testing.T) {
	t.Parallel()
	// Rank165's writable-target refusal remains pinned above; the readonly target
	// needs a descriptor for its higher-order emitCallback parameter before admission.
	t.Skip("awaits compiler/views-v4: runtime-checkable higher-order callable parameter contracts")
}
