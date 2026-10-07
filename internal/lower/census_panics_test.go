package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestCensusComputedRelationPanicsStayNamed(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, where string }{
		{"refusal_scan", "main.a:2:17"},
		{"lowering_unit", "main.a:3:19"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("testdata/census_panics/" + probe.name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			t.Run("full_lower", func(t *testing.T) {
				_, err := lowerSource(t, string(source))
				assertComputedFieldBoundary(t, err, probe.where)
			})

			t.Run("expression_lowering", func(t *testing.T) {
				// The census attempts the unit even after its refusal scan fails.
				// Exercise expression lowering without that scan as a separate pin.
				path := filepath.Join(t.TempDir(), "main.a")
				if err := os.WriteFile(path, source, 0o644); err != nil {
					t.Fatal(err)
				}
				program, err := load.Load([]string{path})
				if err != nil {
					t.Fatal(err)
				}
				file := program.Files()[0]
				checked, release := program.Checker(context.Background(), file)
				defer release()
				var expression *ast.Node
				var visit ast.Visitor
				visit = func(node *ast.Node) bool {
					if node.Kind == ast.KindSatisfiesExpression {
						expression = node
						return true
					}
					return node.ForEachChild(visit)
				}
				file.AsNode().ForEachChild(visit)
				if expression == nil {
					t.Fatal("missing satisfies expression in reduction")
				}
				lowering := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
				_, err = lowering.expression(expression)
				assertComputedFieldBoundary(t, err, probe.where)
			})
		})
	}
}

func assertComputedFieldBoundary(t *testing.T, err error, where string) {
	t.Helper()
	var notYet *NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("want named computed-field boundary, got %v", err)
	}
	if notYet.What != "a computed field name" || !strings.HasSuffix(notYet.Where, where) {
		t.Fatalf("want %s: stage 0 can't lower a computed field name yet, got %v", where, err)
	}
}

func TestComputedRelationFixPreservesExactFields(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"item", "'item'"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, "interface Item { readonly name: string; readonly count?: number } const checked = { "+key+": { name: 'a' } } satisfies { readonly item: Item }; console.log(checked.item.name);")
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
