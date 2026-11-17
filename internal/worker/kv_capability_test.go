package worker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func kvCapabilityCheck(t *testing.T, name string) error {
	t.Helper()
	p, err := load.Load([]string{filepath.Join("testdata/kv_capability", name+".a")})
	if err != nil {
		t.Fatal(err)
	}
	types, release := p.Checker(context.Background(), p.Files()[0])
	// These are compiler-rule tests, not runtime ingress: the fixture's known
	// environment declaration supplies the checked type directly. The production
	// contract is tested separately and must still stop every KV field.
	var envSymbol *ast.Symbol
	for _, n := range p.Files()[0].Statements.Nodes {
		if n.Kind == ast.KindInterfaceDeclaration && n.Name().Text() == "Environment" {
			envSymbol = types.GetSymbolAtLocation(n.Name())
		}
	}
	if envSymbol == nil {
		release()
		t.Fatal("fixture environment declaration missing")
	}
	env := types.GetDeclaredTypeOfSymbol(envSymbol)
	symbol := env.Symbol()
	err = lower.CheckWorkerEnvironment(p, types, env)
	release()
	if err == nil {
		_, ordinary := lower.LowerWorker(context.Background(), p, symbol)
		if name == "get" || name == "put" || name == "delete" {
			return ordinary
		}
		if ordinary != nil {
			t.Fatalf("ordinary lowering still stops: %v", ordinary)
		}
		t.Log("ordinary Worker lowering completed")
	}
	return err
}
func TestWorkerKVCapabilityStops(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"object_literal", "structural_argument", "parameter_default", "property_initializer", "cast", "let", "module", "return", "union", "optional", "array", "record", "generic", "capture", "callback", "field_write", "parameter_write"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := kvCapabilityCheck(t, name)
			var refused *lower.Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "workers:KvNamespace") {
				t.Fatalf("KV capability boundary %s: want named Refused, got %v", name, err)
			}
			t.Log(err)
		})
	}
}
func TestWorkerKVMembersAndViewsStop(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"get", "put", "delete", "reverse_view"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := kvCapabilityCheck(t, name)
			var stop *lower.NotYet
			if !errors.As(err, &stop) || !strings.Contains(err.Error(), "workers:KvNamespace.") {
				t.Fatalf("KV member stop %s: %v", name, err)
			}
			if name == "reverse_view" && !strings.Contains(stop.What, "a member no host registration names") {
				t.Fatalf("reverse KV view missed host walk: %v", err)
			}
			t.Log(err)
		})
	}
}
func TestWorkerKVBindingReadsLower(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"const_destructure", "parameter_destructure", "local_chain", "optional_read", "alias"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := kvCapabilityCheck(t, name); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestWorkerKVFieldIdentityIsPreludeSymbol(t *testing.T) {
	t.Parallel()
	p, err := load.Load([]string{"testdata/kv_capability/fake_name.a"})
	if err != nil {
		t.Fatal(err)
	}
	if err = CheckSignature(context.Background(), p); err == nil || !strings.Contains(err.Error(), "must have type exactly string") {
		t.Fatalf("KV field identity: program lookalike admitted: %v", err)
	}

}
func TestWorkerKVImportIsWorkerOnly(t *testing.T) {
	t.Parallel()
	compiler := workerCompilerProduct(t, nil, false)
	entry, err := filepath.Abs("testdata/kv_capability/const_destructure.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"c", "js"} {
		cmd := exec.CommandContext(t.Context(), compiler, mode, entry)
		out, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(out), "adamic/workers") || !strings.Contains(string(out), "compile it with adamic worker") {
			t.Fatalf("KV %s import: %v: %s", mode, err, out)
		}
	}
}
func TestWorkerKVPurityRefusesOpaqueBindings(t *testing.T) {
	t.Parallel()
	p, err := load.Load([]string{"testdata/kv_capability/pure.a"})
	if err != nil {
		t.Fatal(err)
	}
	types, release := p.Checker(context.Background(), p.Files()[0])
	var symbol *ast.Symbol
	for _, n := range p.Files()[0].Statements.Nodes {
		if n.Kind == ast.KindInterfaceDeclaration && n.Name().Text() == "Environment" {
			symbol = types.GetSymbolAtLocation(n.Name())
		}
	}
	sites := map[ir.SourcePosition]bool{}
	err = lower.CheckWorkerEnvironment(p, types, types.GetDeclaredTypeOfSymbol(symbol), sites)
	release()
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.LowerWorker(context.Background(), p, symbol)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for i, f := range lowered.Functions {
		if f.Name == "observe" {
			found = true
			err = CheckPure(lowered, i, sites)
			if err == nil || !strings.Contains(err.Error(), "impure: holds workers:KvNamespace") {
				t.Fatalf("KV purity: %v", err)
			}
		}
	}
	if !found {
		t.Fatal("purity fixture lost observe")
	}
}

func TestWorkerKVIngressWaitsForBridge(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "worker")
	err := Build("testdata/kv_capability/const_destructure.a", out)
	var stop *lower.NotYet
	if !errors.As(err, &stop) || !strings.Contains(stop.What, `environment field "KV"`) || !strings.Contains(stop.What, "a Workers binding in the environment waits on the bridge's KV check") {
		t.Fatalf("KV ingress contract: want field-named bridge NotYet, got %v", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("KV ingress wrote output: %v", err)
	}
	t.Log(stop)
}

// Language on #bxaz0p1: an async or generator callee's result holds its KV or environment
// parameter past the handler, and rules 1 to 3 don't see it, since the call is direct and the
// result's type holds no binding. Today a neighbor rule stops each road. These pins hold those
// stops by their own words, so lifting one turns this test red before a kept generator or
// promise can use one request's bindings on a later request.
func TestWorkerKVHeldPastFrameNeighborStops(t *testing.T) {
	t.Parallel()
	for name, words := range map[string]string{
		"held_generator": "can't lower a generator function whose generator is used other than by a for-of over its call or a call as a statement",
		"held_promise":   "can't lower a promise not awaited in the same function before an escaping completion",
		"held_void":      "refuses an unawaited async task through the void operator",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p, err := load.Load([]string{filepath.Join("testdata/kv_capability", name+".a")})
			if err != nil {
				t.Fatal(err)
			}
			types, release := p.Checker(context.Background(), p.Files()[0])
			var envSymbol *ast.Symbol
			for _, n := range p.Files()[0].Statements.Nodes {
				if n.Kind == ast.KindInterfaceDeclaration && n.Name().Text() == "Environment" {
					envSymbol = types.GetSymbolAtLocation(n.Name())
				}
			}
			if envSymbol == nil {
				release()
				t.Fatal("fixture environment declaration missing")
			}
			env := types.GetDeclaredTypeOfSymbol(envSymbol)
			err = lower.CheckWorkerEnvironment(p, types, env)
			release()
			if err != nil {
				t.Fatalf("KV held past the frame %s: the environment check now stops it, so this pin moves to that rule: %v", name, err)
			}
			_, err = lower.LowerWorker(context.Background(), p, env.Symbol())
			if err == nil || !strings.Contains(err.Error(), words) {
				t.Fatalf("KV held past the frame %s: its neighbor stop moved, so a kept result now holds a binding past the handler: want %q, got %v", name, words, err)
			}
			t.Log(err)
		})
	}
}
