package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// An unresolved structural constraint does not select a runtime header: a callable
// object can have the same required properties. Concrete substitutions still win.
func TestRepresentationClockChildUnresolvedConstraint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "child.a")
	source := `interface NodeView { readonly kind: number; readonly text: string; }
interface CallableNode extends NodeView { (): void; }
function render<Child extends NodeView>(child: Child): string { return child.text; }
declare const callable: CallableNode;
const view: NodeView = callable;
`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked}
	var parameter *ast.Node
	var callable, view *checker.Type
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindParameter {
			parameter = node.AsParameterDeclaration().Name()
		}
		if node.Kind == ast.KindVariableDeclaration {
			declaration := node.AsVariableDeclaration()
			switch declaration.Name().Text() {
			case "callable":
				callable = checked.GetTypeAtLocation(declaration.Name())
			case "view":
				view = checked.GetTypeAtLocation(declaration.Name())
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if parameter == nil || callable == nil || view == nil {
		t.Fatal("missing checked probe types")
	}
	if got, known := l.representation(view); !known || got != ir.Object {
		t.Fatalf("NodeView representation: %v, %t", got, known)
	}
	if got, known := l.representation(callable); !known || got != ir.Closure {
		t.Fatalf("CallableNode representation: %v, %t", got, known)
	}
	child := checked.GetTypeAtLocation(parameter)
	if child.Flags()&checker.TypeFlagsTypeParameter == 0 {
		t.Fatalf("probe resolved Child: %s", checked.TypeToString(child))
	}
	if got, known := l.representation(child); known {
		t.Fatalf("unresolved Child admitted as %v", got)
	}
	_, err = l.typeOf(parameter)
	var gap *NotYet
	if !errors.As(err, &gap) || gap.What != "a value of type Child" {
		t.Fatalf("want named Child stop, got %v", err)
	}
	for _, kind := range []ir.Type{ir.Object, ir.Closure} {
		l.substitution = map[*checker.Type]ir.Type{child: kind}
		if got, known := l.representation(child); !known || got != kind {
			t.Fatalf("concrete Child: %v, %t; want %v", got, known, kind)
		}
	}
}
