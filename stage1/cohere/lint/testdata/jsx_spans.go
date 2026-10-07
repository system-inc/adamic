// Built through an overlay inside cohere. Prints the UTF-8 span of every JSX text in one file, or
// "empty" for a file that holds JSX with no text, so a test can pick out the captured cases that are JSX.
package main

import (
	"bufio"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		panic("usage: jsx_spans <file>")
	}
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	kind := core.ScriptKindTS
	switch filepath.Ext(path) {
	case ".tsx":
		kind = core.ScriptKindTSX
	case ".jsx":
		kind = core.ScriptKindJSX
	case ".js":
		kind = core.ScriptKindJS
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(absolute)}, string(data), kind)
	jsx := false
	texts := 0
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindJsxElement || n.Kind == ast.KindJsxSelfClosingElement || n.Kind == ast.KindJsxFragment {
			jsx = true
		}
		if n.Kind == ast.KindJsxText {
			texts++
			fmt.Fprintf(out, "%d:%d\n", n.Pos(), n.End())
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	if jsx && texts == 0 {
		fmt.Fprintln(out, "empty")
	}
}
