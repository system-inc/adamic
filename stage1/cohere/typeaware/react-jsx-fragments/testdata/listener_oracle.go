// Reads production listener registrations with Go's parser. Imports no native rule code.
package main

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func main() {
	kinds := map[string]int{"KindJsxElement": int(ast.KindJsxElement), "KindJsxFragment": int(ast.KindJsxFragment), "KindJsxOpeningElement": int(ast.KindJsxOpeningElement), "KindJsxSelfClosingElement": int(ast.KindJsxSelfClosingElement)}
	subjects := []struct{ label, source string }{
		{"react/jsx-fragments", "react/jsx_fragments.go"},
		{"react/jsx-no-constructed-context-values", "react/jsx_no_constructed_context_values.go"},
		{"react/jsx-no-undef", "react/jsx_no_undef.go"},
	}
	for _, subject := range subjects {
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(os.Args[1], "internal/lint/rules", subject.source), nil, 0)
		if err != nil {
			panic(err)
		}
		values := []int{}
		goast.Inspect(file, func(node goast.Node) bool {
			literal, ok := node.(*goast.CompositeLit)
			if !ok {
				return true
			}
			typ, ok := literal.Type.(*goast.SelectorExpr)
			if !ok || typ.Sel.Name != "Listeners" {
				return true
			}
			for _, element := range literal.Elts {
				pair, ok := element.(*goast.KeyValueExpr)
				if !ok {
					panic("production listener is not keyed")
				}
				key, ok := pair.Key.(*goast.SelectorExpr)
				if !ok {
					panic("production listener kind is not a selector")
				}
				value, ok := kinds[key.Sel.Name]
				if !ok {
					panic("unknown production listener kind " + key.Sel.Name)
				}
				values = append(values, value)
			}
			return false
		})
		if len(values) == 0 {
			panic("missing production listeners")
		}
		sort.Ints(values)
		text := []string{}
		for _, value := range values {
			text = append(text, strconv.Itoa(value))
		}
		fmt.Printf("%s %s\n", subject.label, strings.Join(text, ","))
	}
}
