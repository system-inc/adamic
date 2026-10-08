// Parse Go test fixtures as Go syntax, preserving their file and line evidence.
package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
)

type fixture struct {
	Text string `json:"text"`
	File string `json:"file"`
	Line int    `json:"line"`
}

func main() {
	paths, err := filepath.Glob("cohere/internal/format/css/selector/*_test.go")
	if err != nil {
		panic(err)
	}
	files := token.NewFileSet()
	results := []fixture{}
	for _, name := range paths {
		file, err := parser.ParseFile(files, name, nil, 0)
		if err != nil {
			panic(err)
		}
		add := func(expression ast.Expr) {
			literal, ok := expression.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				panic(err)
			}
			results = append(results, fixture{value, name, files.Position(literal.Pos()).Line})
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if spec, ok := node.(*ast.ValueSpec); ok {
				for _, name := range spec.Names {
					if name.Name == "selectorFixtures" {
						ast.Inspect(spec, func(n ast.Node) bool {
							if field, ok := n.(*ast.KeyValueExpr); ok {
								if key, ok := field.Key.(*ast.Ident); ok && key.Name == "text" {
									add(field.Value)
								}
							}
							return true
						})
					}
				}
			}
			if call, ok := node.(*ast.CallExpr); ok {
				if name, ok := call.Fun.(*ast.Ident); ok && name.Name == "mustParse" && len(call.Args) == 2 {
					add(call.Args[1])
				}
			}
			return true
		})
	}
	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		panic(err)
	}
}
