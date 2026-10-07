// Extract literal source inputs from the pinned production Go rule tests.
package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

func main() {
	var sources []string
	seen := map[string]bool{}
	for _, path := range os.Args[1:] {
		if path == "--" {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			panic(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(literal.Value)
			if err == nil && len(text) > 4 && !seen[text] && (strings.Contains(text, ";") || strings.Contains(text, "=>") || strings.Contains(text, "{") || strings.Contains(text, "Symbol(") || strings.Contains(text, "typeof ")) {
				seen[text] = true
				sources = append(sources, text)
			}
			return true
		})
	}
	if err := json.NewEncoder(os.Stdout).Encode(sources); err != nil {
		panic(err)
	}
}
