package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func TestCycleShapeRelationsKeepLiteralDistinctions(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := `const first = { name: 'first', values: [1] };
const second = { name: 'second', values: [2] };
const different = { name: 'third', values: ['three'] };
const spread = { ...first };
const method = { name: 'method', values: [3], size() { return this.values.length; } };
const frozen = { name: 'frozen', values: [4] } as const;
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
	types := map[string]*checker.Type{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			declaration := node.AsVariableDeclaration()
			types[declaration.Name().Text()] = checked.GetTypeAtLocation(declaration.Initializer)
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	finder := &cycleFinder{l: &lowering{checker: checked}}
	first := finder.shapeRepresentative(types["first"])
	if first != finder.shapeRepresentative(types["second"]) {
		t.Fatal("equivalent plain literals did not share relation cache")
	}
	for _, name := range []string{"different", "spread", "method", "frozen"} {
		if finder.shapeRepresentative(types[name]) == first {
			t.Fatalf("%s lost its literal distinction", name)
		}
	}
	if finder.shapeRepresentative(types["spread"]) != types["spread"] {
		t.Fatal("spread members must retain their declaration identity")
	}
	for _, from := range types {
		for _, to := range types {
			if got, want := finder.shapeAssignable(from, to), checked.IsTypeAssignableTo(from, to); got != want {
				t.Fatalf("cached relation %s -> %s = %t, checker = %t", checked.TypeToString(from), checked.TypeToString(to), got, want)
			}
		}
	}
}
