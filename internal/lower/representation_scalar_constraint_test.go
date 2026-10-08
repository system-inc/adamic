package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestRepresentationScalarConstraint(t *testing.T) {
	for _, probe := range []struct {
		constraint string
		want       ir.Type
		known      bool
	}{
		{"boolean", ir.Boolean, true}, {"number", ir.Number, true}, {"string", ir.String, true},
		{"unknown", 0, false}, {"object", 0, false}, {"string | number", 0, false},
		{"{ readonly length: number }", 0, false},
	} {
		t.Run(probe.constraint, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "constraint.a")
			source := "function held<T extends " + probe.constraint + ">(value: T): T { return value; }"
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checker, release := program.Checker(context.Background(), file)
			defer release()
			var parameter *ast.Node
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindParameter {
					parameter = node.Name()
					return false
				}
				return node.ForEachChild(visit)
			}
			file.AsNode().ForEachChild(visit)
			if parameter == nil {
				t.Fatal("missing type parameter use")
			}
			l := lowering{checker: checker}
			got, known := l.representation(checker.GetTypeAtLocation(parameter))
			if got != probe.want || known != probe.known {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, known, probe.want, probe.known)
			}
		})
	}
}
