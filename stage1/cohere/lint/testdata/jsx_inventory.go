// Built through an overlay inside cohere. Reports JSX membership for every path
// in a manifest, using the same unmodified Go parser and criterion as jsx_spans.go.
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
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		panic("usage: jsx_inventory <manifest>")
	}
	manifest, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for _, path := range strings.Split(strings.TrimSpace(string(manifest)), "\n") {
		value := 0
		if hasJsx(path) {
			value = 1
		}
		fmt.Fprintf(out, "%s\t%d\n", path, value)
	}
}

func hasJsx(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	kind := core.ScriptKindTS
	switch filepath.Ext(path) {
	case ".tsx":
		kind = core.ScriptKindTSX
	case ".jsx":
		kind = core.ScriptKindJSX
	case ".js":
		kind = core.ScriptKindJS
	}
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(path), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(path))}, string(data), kind)
	jsx := false
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindJsxElement || n.Kind == ast.KindJsxSelfClosingElement || n.Kind == ast.KindJsxFragment {
			jsx = true
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	return jsx
}
