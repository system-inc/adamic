// Built through a cohere overlay: the parser, not the lint verdict, is the oracle.
package main

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func main() {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/repository/Thing.tsx"}, "export const Thing = () => <p>hi</p>;\n", core.ScriptKindTSX)
	if len(file.Diagnostics()) != 0 {
		panic("Go did not parse JSX")
	}
	found := false
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindJsxElement {
			found = true
		}
		node.ForEachChild(visit)
		return false
	}
	visit(file.AsNode())
	fmt.Println(found)
}
