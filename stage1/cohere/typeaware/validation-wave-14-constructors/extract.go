package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	f, e := parser.ParseFile(token.NewFileSet(), os.Args[1], nil, 0)
	if e != nil {
		panic(e)
	}
	dir := os.Args[2]
	if e = os.MkdirAll(dir, 0755); e != nil {
		panic(e)
	}
	var sources []string
	add := func(expr ast.Expr) {
		if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			s, e := strconv.Unquote(lit.Value)
			if e != nil {
				panic(e)
			}
			if strings.Contains(s, "RegExp") {
				sources = append(sources, s)
			}
		}
	}
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || (fn.Name.Name != "TestNoMisleadingCharacterClassFires" && fn.Name.Name != "TestNoMisleadingCharacterClassStaysSilent") {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if list, ok := n.(*ast.CompositeLit); ok && len(list.Elts) > 0 {
				if a, ok := list.Type.(*ast.ArrayType); ok {
					if id, ok := a.Elt.(*ast.Ident); ok && id.Name == "string" {
						for _, el := range list.Elts {
							add(el)
						}
					}
				} else {
					add(list.Elts[0])
				}
			}
			return true
		})
	}
	var manifest strings.Builder
	for i, s := range sources {
		p := filepath.Join(dir, fmt.Sprintf("upstream-%03d.a", i))
		if e = os.WriteFile(p, []byte(s+"\nexport {};\n"), 0600); e != nil {
			panic(e)
		}
		manifest.WriteString(p + "\n")
	}
	if e = os.WriteFile(filepath.Join(dir, "upstream.manifest"), []byte(manifest.String()), 0600); e != nil {
		panic(e)
	}
	fmt.Println(len(sources))
}
