// Reads production cohere listener declarations, using its pinned numeric AST kinds.
package main

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func main() {
	kinds := map[string]ast.Kind{
		"KindCallExpression":          ast.KindCallExpression,
		"KindElementAccessExpression": ast.KindElementAccessExpression,
		"KindVariableDeclaration":     ast.KindVariableDeclaration,
		"KindTryStatement":            ast.KindTryStatement,
		"KindBinaryExpression":        ast.KindBinaryExpression,
		"KindExpressionStatement":     ast.KindExpressionStatement,
		"KindSourceFile":              ast.KindSourceFile,
	}
	for _, name := range os.Args[2:] {
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(os.Args[1], "internal/lint/rules/nexus", name+".go"), nil, 0)
		if err != nil {
			panic(err)
		}
		var numbers []string
		maps := 0
		goast.Inspect(file, func(node goast.Node) bool {
			literal, ok := node.(*goast.CompositeLit)
			if !ok {
				return true
			}
			typ, ok := literal.Type.(*goast.SelectorExpr)
			if !ok || typ.Sel.Name != "Listeners" {
				return true
			}
			owner, ok := typ.X.(*goast.Ident)
			if !ok || owner.Name != "rule" {
				return true
			}
			maps++
			for _, element := range literal.Elts {
				pair, ok := element.(*goast.KeyValueExpr)
				if !ok {
					panic("listener without a key")
				}
				key, ok := pair.Key.(*goast.SelectorExpr)
				if !ok {
					panic("nonconstant listener key")
				}
				value, ok := kinds[key.Sel.Name]
				if !ok {
					panic("unknown numeric kind: " + key.Sel.Name)
				}
				numbers = append(numbers, fmt.Sprint(int(value)))
			}
			return false
		})
		if maps != 1 || len(numbers) == 0 {
			panic("missing or ambiguous listener map: " + name)
		}
		fmt.Printf("%s %s\n", name, strings.Join(numbers, ","))
	}
}
