package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

func main() {
	f, e := parser.ParseFile(token.NewFileSet(), os.Args[1], nil, 0)
	if e != nil {
		panic(e)
	}
	vars := map[string]string{}
	eval := func(x ast.Expr) string { return "" }
	eval = func(x ast.Expr) string {
		switch n := x.(type) {
		case *ast.BasicLit:
			if n.Kind == token.STRING {
				s, _ := strconv.Unquote(n.Value)
				return s
			}
		case *ast.Ident:
			return vars[n.Name]
		case *ast.BinaryExpr:
			if n.Op == token.ADD {
				return eval(n.X) + eval(n.Y)
			}
		}
		return ""
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if a, ok := n.(*ast.AssignStmt); ok && len(a.Lhs) == 1 && len(a.Rhs) == 1 {
			if id, ok := a.Lhs[0].(*ast.Ident); ok {
				vars[id.Name] = eval(a.Rhs[0])
			}
		}
		return true
	})
	out := map[string]string{}
	seen := map[string]bool{}
	add := func(s string) {
		if (strings.Contains(s, "async") || strings.Contains(s, "await ")) && !seen[s] {
			seen[s] = true
			out[fmt.Sprintf("require_await-upstream-%03d.ts", len(out))] = s
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if s, ok := c.Fun.(*ast.SelectorExpr); ok && s.Sel.Name == "RunTyped" && len(c.Args) == 4 {
				add(eval(c.Args[3]))
				return false
			}
		}
		if b, ok := n.(*ast.BinaryExpr); ok && b.Op == token.ADD {
			add(eval(b))
			return false
		}
		if b, ok := n.(*ast.BasicLit); ok && b.Kind == token.STRING {
			add(eval(b))
		}
		return true
	})
	json.NewEncoder(os.Stdout).Encode(out)
}
