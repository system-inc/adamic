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
	tree, err := parser.ParseFile(token.NewFileSet(), os.Args[1], nil, 0)
	if err != nil {
		panic(err)
	}
	seen := map[string]bool{}
	var sources []string
	ast.Inspect(tree, func(node ast.Node) bool {
		lit, ok := node.(*ast.CompositeLit)
		if !ok || len(lit.Elts) < 2 {
			return true
		}
		value, ok := lit.Elts[1].(*ast.BasicLit)
		if !ok || value.Kind != token.STRING {
			return true
		}
		text, err := strconv.Unquote(value.Value)
		if err != nil {
			panic(err)
		}
		if strings.Contains(text, ":") && strings.Contains(text, "{") && !seen[text] {
			seen[text] = true
			sources = append(sources, text)
		}
		return true
	})
	if err := json.NewEncoder(os.Stdout).Encode(sources); err != nil {
		panic(err)
	}
}
