// Instrument exactly lowering.statement, retaining its body and normal calls.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

func printed(n ast.Node) string {
	var b bytes.Buffer
	_ = format.Node(&b, token.NewFileSet(), n)
	return b.String()
}
func main() {
	if len(os.Args) != 3 {
		panic("usage: statementrewrite input output")
	}
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	matches := []*ast.FuncDecl{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name.Name == "latentStatementRaw" {
				panic("latent statement rewrite: lowering.statement: raw method already exists")
			}
			if fn.Name.Name == "statement" && fn.Recv != nil && len(fn.Recv.List) == 1 && printed(fn.Recv.List[0].Type) == "*lowering" {
				matches = append(matches, fn)
			}
		}
	}
	if len(matches) != 1 {
		panic(fmt.Sprintf("latent statement rewrite: lowering.statement: expected exactly one function, found %d", len(matches)))
	}
	fn := matches[0]
	if printed(fn.Type) != "func(node *ast.Node) ([]ir.Statement, error)" {
		panic("latent statement rewrite: lowering.statement: unsupported signature " + printed(fn.Type))
	}
	fn.Name.Name = "latentStatementRaw"
	wrapper, err := parser.ParseFile(fs, "wrapper.go", "package lower\nfunc (l *lowering) statement(node *ast.Node)([]ir.Statement,error){return l.latentStatement(node)}", 0)
	if err != nil {
		panic(err)
	}
	file.Decls = append(file.Decls, wrapper.Decls[0])
	var out bytes.Buffer
	if err := format.Node(&out, fs, file); err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[2], out.Bytes(), 0600); err != nil {
		panic(err)
	}
}
