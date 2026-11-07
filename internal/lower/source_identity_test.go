package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Both substitutions have the same runtime representation. Their field layouts
// must still give the nested declaration distinct semantic identities.
func TestSourceIdentitySubstitutedObjectShapes(test *testing.T) {
	test.Parallel()
	entry := filepath.Join(test.TempDir(), "main.a")
	source := `type First = { padding: number; value: number };
type Second = { value: number };
function hold<T>(value: T): T { const read = <U>(item: T): T => item; return read(value); }
`
	if writeError := os.WriteFile(entry, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{entry})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	var arrow *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindArrowFunction {
			arrow = node
			return true
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	if arrow == nil {
		test.Fatal("missing nested generic arrow")
	}
	binder := checked.GetTypeAtLocation(arrow.TypeParameters()[0].Name())
	lowering := &lowering{
		program:      program,
		checker:      checked,
		substitution: map[*checker.Type]ir.Type{binder: ir.Object},
	}
	keys := []string{}
	for _, statement := range file.Statements.Nodes[:2] {
		concrete := checked.GetTypeAtLocation(statement.Name())
		lowering.typeMapper = newTypeMapper([]*checker.Type{binder}, []*checker.Type{concrete})
		identity := lowering.functionSource(arrow)
		if identity.Specialization != lowering.functionSource(arrow).Specialization {
			test.Fatal("same specialization changed identity")
		}
		keys = append(keys, identity.Specialization)
	}
	if keys[0] == keys[1] {
		test.Fatalf("distinct concrete object shapes share specialization %q", keys[0])
	}
}
