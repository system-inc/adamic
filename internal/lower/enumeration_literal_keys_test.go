package lower

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestEnumerationLiteralKeyBoundaries(t *testing.T) {
	t.Parallel()
	const outside = "const unrelated = { '#': 1 };\n"
	const reserved = "Object enumeration of a literal with reserved or symbol key storage"
	var chain strings.Builder
	chain.WriteString(outside + "const s0 = { key: 1 };\n")
	// Generate one spread link beyond the limit, including the operand.
	for depth := 1; depth <= enumerationSpreadLimit; depth++ {
		fmt.Fprintf(&chain, "const s%d = { ...s%d };\n", depth, depth-1)
	}
	fmt.Fprintf(&chain, "Object.values({ ...s%d });\n", enumerationSpreadLimit)
	for _, row := range []struct {
		name, source string
		deep, stops  bool
	}{
		{"parenthesized const", outside + "Object.values((({ key: 1 }) as const));", false, false},
		{"numeric key syntax", outside + "Object.values({ 12: 1 });", false, false},
		{"string prototype", outside + "Object.values({ '__proto__': 1, key: 2 });", false, true},
		{"identifier prototype", outside + "Object.values({ __proto__: 1, key: 2 });", false, true},
		{"shorthand prototype", outside + "const __proto__ = 1; Object.values({ __proto__, key: 2 });", false, true},
		{"computed string", outside + "Object.values({ ['key']: 1 });", false, true},
		{"method", outside + "Object.values({ key() { return 1; } });", false, true},
		{"clean spread", outside + "const source = { key: 1 }; Object.values({ ...source });", false, false},
		{"reserved spread", "const source = { '#': 1 }; Object.values({ ...source });", false, true},
		{"reserved nested spread", "const source = { '#': 1 }; const copy = { ...source }; Object.values({ ...copy });", false, true},
		{"prototype spread", outside + "const __proto__ = 1; const source = { __proto__ }; Object.values({ ...source });", false, true},
		{"written source", outside + "const source = { key: 1 }; Object.values({ ...source }); source.key = 2;", false, true},
		{"parameter source", outside + "function read(source: { key: number }) { return Object.values({ ...source }); }", false, true},
		{"call source", outside + "function source() { return { key: 1 }; } Object.values({ ...source() });", false, true},
		{"escaped source", outside + "function write(target: { [key: string]: number }) { target['#'] = 2; } const source = { key: 1 }; write(source); Object.values({ ...source });", false, true},
		{"depth bound", chain.String(), false, true},
		{"deep descriptor scan", outside + "Object.values({ key: 1 });", true, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(row.source), 0o644); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			var call *ast.Node
			var find ast.Visitor
			find = func(node *ast.Node) bool {
				if node.Kind == ast.KindCallExpression {
					expression := node.AsCallExpression().Expression
					if expression.Kind == ast.KindPropertyAccessExpression && expression.Name().Text() == "values" {
						call = node
						return true
					}
				}
				return node.ForEachChild(find)
			}
			file.AsNode().ForEachChild(find)
			if call == nil {
				t.Fatal("Object.values call not found")
			}
			// Ask the guard directly so a later object lowering refusal
			// (for example object.go's prototype guard) cannot hide its answer.
			lower := &lowering{program: program, checker: checked, result: &ir.Program{}}
			err = lower.enumerationDescriptors(call, call.AsCallExpression().Arguments.Nodes[0], row.deep, false)
			if !row.stops {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var stop *NotYet
			if !errors.As(err, &stop) || stop.What != reserved {
				t.Fatalf("want %q, got %v", reserved, err)
			}
		})
	}
}
