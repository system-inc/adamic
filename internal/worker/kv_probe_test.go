package worker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Scratch probe, never committed: each candidate's check and ordinary lowering, printed.
func TestKVProbe(t *testing.T) {
	names, _ := filepath.Glob("testdata/kv_probe/*.a")
	for _, name := range names {
		p, err := load.Load([]string{name})
		if err != nil {
			t.Logf("PROBE %s %s load: %v", os.Getenv("PROBE_MODE"), filepath.Base(name), err)
			continue
		}
		types, release := p.Checker(context.Background(), p.Files()[0])
		var envSymbol *ast.Symbol
		for _, n := range p.Files()[0].Statements.Nodes {
			if n.Kind == ast.KindInterfaceDeclaration && n.Name().Text() == "Environment" {
				envSymbol = types.GetSymbolAtLocation(n.Name())
			}
		}
		env := types.GetDeclaredTypeOfSymbol(envSymbol)
		symbol := env.Symbol()
		check := lower.CheckWorkerEnvironment(p, types, env)
		release()
		var lowered error
		if check == nil {
			_, lowered = lower.LowerWorker(context.Background(), p, symbol)
		}
		t.Logf("PROBE %s %s check=%v lower=%v", os.Getenv("PROBE_MODE"), filepath.Base(name), check, lowered)
	}
}
