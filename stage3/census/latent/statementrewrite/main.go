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
	"strings"
)

func printed(node ast.Node) string {
	var buffer bytes.Buffer
	_ = format.Node(&buffer, token.NewFileSet(), node)
	return buffer.String()
}

// fieldTypes prints a parameter or result list's types in order, one per declared name, without the names.
func fieldTypes(fields *ast.FieldList) string {
	types := []string{}
	if fields != nil {
		for _, field := range fields.List {
			for range max(1, len(field.Names)) {
				types = append(types, printed(field.Type))
			}
		}
	}
	return strings.Join(types, ", ")
}

func main() {
	if len(os.Args) != 3 {
		panic("usage: statementrewrite input output")
	}
	fs := token.NewFileSet()
	file, parseFileError := parser.ParseFile(fs, os.Args[1], nil, parser.ParseComments)
	if parseFileError != nil {
		panic(parseFileError)
	}
	matches := []*ast.FuncDecl{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name.Name == "latentStatementRaw" {
				panic("latent statement rewrite: lowering.statement: raw method already exists")
			}
			if fn.Name.Name == "statement" &&
				fn.Recv != nil &&
				len(fn.Recv.List) == 1 &&
				printed(fn.Recv.List[0].Type) == "*lowering" {
				matches = append(matches, fn)
			}
		}
	}
	if len(matches) != 1 {
		panic(fmt.Sprintf(
			"latent statement rewrite: lowering.statement: expected exactly one function, found %d",
			len(matches),
		))
	}
	fn := matches[0]
	// The types are the contract, not the names: Diagnostics' debug source lines (f2cf1ebb) name the results so a
	// defer can prepend a marker, and the raw method keeps its own names under the unnamed wrapper (#mdaed6m).
	if fieldTypes(fn.Type.Params) != "*ast.Node" || fieldTypes(fn.Type.Results) != "[]ir.Statement, error" {
		panic("latent statement rewrite: lowering.statement: unsupported signature " + printed(fn.Type))
	}
	fn.Name.Name = "latentStatementRaw"
	wrapper, parseFileError := parser.ParseFile(
		fs,
		"wrapper.go",
		"package lower\nfunc (l *lowering) statement(node *ast.Node)([]ir.Statement,error){return l.latentStatement(node)}",
		0,
	)
	if parseFileError != nil {
		panic(parseFileError)
	}
	file.Decls = append(file.Decls, wrapper.Decls[0])
	var out bytes.Buffer
	if nodeError := format.Node(&out, fs, file); nodeError != nil {
		panic(nodeError)
	}
	if writeError := os.WriteFile(os.Args[2], out.Bytes(), 0600); writeError != nil {
		panic(writeError)
	}
}
