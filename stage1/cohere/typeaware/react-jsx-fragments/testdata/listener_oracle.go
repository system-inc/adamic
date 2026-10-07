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
	"strings"
)

func main() {
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
		values := []string{}
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
				values = append(values, strings.TrimPrefix(key.Sel.Name, "Kind"))
			}
			return false
		})
		if len(values) == 0 {
			panic("missing production listeners")
		}
		sort.Strings(values)
		text := []string{}
		for _, value := range values {
			text = append(text, value)
		}
		fmt.Printf("%s %s\n", subject.label, strings.Join(text, ","))
	}
}
