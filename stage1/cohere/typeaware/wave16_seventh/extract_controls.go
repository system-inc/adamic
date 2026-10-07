//go:build ignore

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

type row struct{ Name, Function, Source string }

func text(expression ast.Expr) (string, bool) {
	switch x := expression.(type) {
	case *ast.BasicLit:
		if x.Kind == token.STRING {
			s, e := strconv.Unquote(x.Value)
			return s, e == nil
		}
	case *ast.BinaryExpr:
		if x.Op == token.ADD {
			a, ok := text(x.X)
			b, held := text(x.Y)
			return a + b, ok && held
		}
	case *ast.Ident:
		if x.Name == "provider" {
			return "<Ctx.Provider value={{a: 1}} />", true
		}
	}
	return "", false
}
func main() {
	var result []row
	for _, path := range os.Args[1:] {
		if path == "--" {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			panic(err)
		}
		for _, declaration := range f.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}
			ast.Inspect(fn, func(n ast.Node) bool {
				literal, ok := n.(*ast.CompositeLit)
				if !ok || len(literal.Elts) < 2 {
					return true
				}
				a, ok := literal.Elts[0].(ast.Expr)
				if !ok {
					return true
				}
				name, ok := text(a)
				if !ok {
					return true
				}
				b, ok := literal.Elts[1].(ast.Expr)
				if !ok {
					return true
				}
				code, ok := text(b)
				if !ok {
					return true
				}
				if strings.HasSuffix(fn.Name.Name, "ConstructionKinds") {
					code = "declare const c:any,x:any,a:any,b:any,rest:any,Foo:any;declare function makeIt():any;\nfunction Component(){return <Ctx.Provider value={" + code + "}/>;}"
				}
				if strings.HasSuffix(fn.Name.Name, "IdentifierFollowing") {
					code = "declare function makeIt():any;\nfunction Component(){" + code + "\nreturn <Ctx.Provider value={v}/>;}"
				}
				result = append(result, row{name, fn.Name.Name, code})
				return false
			})
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
