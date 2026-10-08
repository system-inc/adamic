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

func TestRepresentationClockStringNullUndefinedAdmission(t *testing.T) {
	for _, probe := range []struct {
		constraint string
		want       ir.Type
		known      bool
	}{
		{"string | null | undefined", ir.NullishString, true},
		{"'text' | '' | null | undefined", ir.NullishString, true},
		{"string | undefined", ir.String, true},
		{"string | null", 0, false},
		{"null | undefined", 0, false},
		{"string | number | null | undefined", 0, false},
		{"number | null | undefined", 0, false},
	} {
		t.Run(probe.constraint, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "constraint.a")
			source := "function held(value: " + probe.constraint + "): void {}"
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
