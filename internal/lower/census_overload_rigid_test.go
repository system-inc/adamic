package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// censusOverload judges an overload set with its type parameters rigid, whoever asks (#6zqxygf). Lowering
// an instance of a generic call installs that call's mapper (instantiateFunctionSignature, generic.go)
// before the callee's signature reaches censusOverloads, and when the implementation has no binders of its
// own, as tsc's debug.ts assertNode has none, the set was judged at that call's types and the census kept
// the judgment for every other call. Production lowering reaches the declaration rigid first, but the
// latent census lowers unit by unit and defers generics, so there the instance came first (Correctness's
// stack on #68x1t2y). This drives that order directly: the census asked under the mapper an instance at
// T = Base installs. f's overload promises its caller's own T back, the implementation returns a fresh
// Base, so rigid it refuses, and at T = Base it would pass. Mutant: censusOverload's l.typeMapper = nil
// dropped, and the instance's judgment passes.
func TestCensusOverloadJudgesRigidUnderAnInstanceMapper(t *testing.T) {
	t.Parallel()
	const refusal = "overload 1 of f result T cannot be served by implementation result Base"
	path := filepath.Join(t.TempDir(), "main.ts")
	source := `interface Base { readonly n: number }
function f<T extends Base>(x: T): T;
function f(x: Base): Base {
    return { n: x.n };
}
const b: Base = { n: 1 };
console.log(String(f(b).n));
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	var overload, implementation, base *ast.Node
	for _, statement := range file.Statements.Nodes {
		switch {
		case statement.Kind == ast.KindInterfaceDeclaration:
			base = statement
		case statement.Kind == ast.KindFunctionDeclaration && statement.Body() == nil:
			overload = statement
		case statement.Kind == ast.KindFunctionDeclaration:
			implementation = statement
		}
	}
	if overload == nil || implementation == nil || base == nil {
		t.Fatal("missing the overload, the implementation or Base")
	}
	refuses := func(l *lowering, when string) {
		t.Helper()
		var refused *Refused
		if err := l.censusOverloads(implementation); !errors.As(err, &refused) || !strings.Contains(refused.What, refusal) {
			t.Fatalf("%s: want %q, got %v", when, refusal, err)
		}
	}

	// Asked with nothing installed, as module lowering asks, it refuses.
	refuses(&lowering{program: program, checker: checked, result: &ir.Program{}}, "rigid")

	// Asked under the mapper an instance of f at T = Base installs, it refuses the same way, and leaves
	// the instance's mapper in place for the body that goes on lowering under it.
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	parameter := checked.GetTypeAtLocation(overload.TypeParameters()[0].Name())
	baseType := checked.GetDeclaredTypeOfSymbol(checked.GetSymbolAtLocation(base.Name()))
	installed := l.instanceMapper(overload, map[*checker.Type]*checker.Type{parameter: baseType})
	if installed == nil {
		t.Fatal("instanceMapper read nothing back for T")
	}
	l.typeMapper = installed
	refuses(l, "under the instance mapper")
	if l.typeMapper != installed {
		t.Fatal("censusOverloads didn't restore the instance's mapper")
	}
}
