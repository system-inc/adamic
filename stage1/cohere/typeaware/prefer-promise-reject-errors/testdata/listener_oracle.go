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
	kinds := map[string]int{"KindExpressionStatement": int(ast.KindExpressionStatement), "KindVoidExpression": int(ast.KindVoidExpression), "KindCallExpression": int(ast.KindCallExpression), "KindNewExpression": int(ast.KindNewExpression), "KindSourceFile": int(ast.KindSourceFile), "KindIdentifier": int(ast.KindIdentifier)}
	subjects := []struct{ label, source string }{
		{"floating", "typescript/no_floating_promises.go"},
		{"eval", "typescript/no_implied_eval.go"},
		{"void", "typescript/no_meaningless_void_operator.go"},
		{"output", "nexus/correctness_no_process_exit_after_output.go"},
		{"timer", "nexus/correctness_no_uncleared_race_timeout.go"},
		{"blocking", "nexus/correctness_require_blocking_standard_streams.go"},
		{"promise", "core/prefer_promise_reject_errors.go"},
		{"regex", "core/prefer_regex_literals.go"},
		{"rest", "core/prefer_rest_params.go"},
		{"effect", "react/set_state_in_effect.go"},
		{"render", "react/set_state_in_render.go"},
		{"static", "react/static_components.go"},
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
