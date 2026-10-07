// Extracts literal fixture inputs from the pinned Go test source, without the shared harness.
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
	f, err := parser.ParseFile(token.NewFileSet(), os.Args[1], nil, 0)
	if err != nil {
		panic(err)
	}
	values := map[string]string{}
	var eval func(ast.Expr) (string, bool)
	eval = func(e ast.Expr) (string, bool) {
		switch x := e.(type) {
		case *ast.BasicLit:
			if x.Kind != token.STRING {
				return "", false
			}
			s, err := strconv.Unquote(x.Value)
			return s, err == nil
		case *ast.Ident:
			s, ok := values[x.Name]
			return s, ok
		case *ast.BinaryExpr:
			if x.Op != token.ADD {
				return "", false
			}
			a, ok := eval(x.X)
			b, yes := eval(x.Y)
			return a + b, ok && yes
		case *ast.ParenExpr:
			return eval(x.X)
		}
		return "", false
	}
	for _, decl := range f.Decls {
		if g, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range g.Specs {
				if v, ok := spec.(*ast.ValueSpec); ok {
					for i, n := range v.Names {
						if i < len(v.Values) {
							if s, ok := eval(v.Values[i]); ok {
								values[n.Name] = s
							}
						}
					}
				}
			}
		}
	}
	inputs := []string{}
	seen := map[string]bool{}
	add := func(s string) {
		if !seen[s] && (strings.Contains(s, "function ") || strings.Contains(s, "=>")) && !strings.Contains(s, "declare module") && !strings.Contains(s, "export declare function") {
			seen[s] = true
			inputs = append(inputs, s)
		}
	}
	for _, decl := range f.Decls {
		if g, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range g.Specs {
				if v, ok := spec.(*ast.ValueSpec); ok {
					for _, e := range v.Values {
						if s, ok := eval(e); ok {
							add(s)
						}
					}
				}
			}
		}
	}
	ast.Inspect(f, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.KeyValueExpr:
			if k, ok := x.Key.(*ast.Ident); ok && k.Name == "source" {
				if s, ok := eval(x.Value); ok {
					add(s)
				}
			}
		case *ast.AssignStmt:
			for i, e := range x.Lhs {
				if n, ok := e.(*ast.Ident); ok && i < len(x.Rhs) {
					if s, ok := eval(x.Rhs[i]); ok {
						values[n.Name] = s
						if n.Name == "source" {
							add(s)
						}
					}
				}
			}
		case *ast.CompositeLit:
			if a, ok := x.Type.(*ast.ArrayType); ok {
				if name, ok := a.Elt.(*ast.Ident); ok && name.Name == "string" {
					for _, expression := range x.Elts {
						if text, ok := eval(expression); ok {
							add(text)
						}
					}
				}
			}
		case *ast.CallExpr:
			if fn, ok := x.Fun.(*ast.Ident); ok && (strings.Contains(fn.Name, "Fixture") || strings.HasPrefix(fn.Name, "runSetState")) && len(x.Args) >= 2 {
				if s, ok := eval(x.Args[1]); ok {
					add(s)
				}
			}
		}
		return true
	})
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Constants map[string]string
		Inputs    []string
	}{values, inputs}); err != nil {
		panic(err)
	}
}
