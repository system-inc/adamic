// Extract default-option source strings directly from the pinned Go test AST.
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

func text(e ast.Expr) (string, bool) {
	switch n := e.(type) {
	case *ast.BasicLit:
		if n.Kind == token.STRING {
			s, err := strconv.Unquote(n.Value)
			return s, err == nil
		}
	case *ast.BinaryExpr:
		if n.Op == token.ADD {
			a, ok := text(n.X)
			b, yes := text(n.Y)
			return a + b, ok && yes
		}
	}
	return "", false
}
func main() {
	var sources []string
	for _, path := range os.Args[1:] {
		if path == "--" {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			panic(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CompositeLit:
				if strings.Contains(path, "regex_literals_corpus") && len(node.Elts) == 4 {
					origin, a := text(node.Elts[0].(ast.Expr))
					options, b := text(node.Elts[1].(ast.Expr))
					source, c := text(node.Elts[2].(ast.Expr))
					if a && b && c && (origin == "corpus" || origin == "edge") && options == "" {
						sources = append(sources, source)
					}
				}
				if strings.Contains(path, "rest_params") && len(node.Elts) == 2 {
					_, a := text(node.Elts[0].(ast.Expr))
					source, b := text(node.Elts[1].(ast.Expr))
					if a && b {
						sources = append(sources, source)
					}
				}
			case *ast.RangeStmt:
				if !strings.Contains(path, "promise_reject") {
					break
				}
				list, ok := node.X.(*ast.CompositeLit)
				if !ok {
					break
				}
				array, ok := list.Type.(*ast.ArrayType)
				if !ok {
					break
				}
				name, ok := array.Elt.(*ast.Ident)
				if !ok || name.Name != "string" {
					break
				}
				for _, e := range list.Elts {
					source, ok := text(e.(ast.Expr))
					if ok {
						sources = append(sources, source)
					}
				}
			}
			return true
		})
	}
	if err := json.NewEncoder(os.Stdout).Encode(sources); err != nil {
		panic(err)
	}
}
