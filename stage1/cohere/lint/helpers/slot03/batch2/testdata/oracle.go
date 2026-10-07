package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"os"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	f, e := os.Open(os.Args[1])
	must(e)
	defer f.Close()
	z, e := gzip.NewReader(f)
	must(e)
	defer z.Close()
	scan := bufio.NewScanner(z)
	scan.Buffer(make([]byte, 4096), 16<<20)
	rows := []map[string]any{}
	names := []string{"a", "img", "script", "link", "input", "iframe", "html", "body", "Image", "", "div", "é", "😀"}
	observe := func(n *ast.Node) {
		kind, text, id := "", "", -1
		if n != nil {
			id = 0
			kind = strings.TrimPrefix(n.Kind.String(), "Kind")
			if n.Kind == ast.KindIdentifier || n.Kind == ast.KindStringLiteral || n.Kind == ast.KindPrivateIdentifier {
				text = n.Text()
			}
		}
		queries := append([]string{}, names...)
		queries = append(queries, text)
		for _, name := range queries {
			rows = append(rows, map[string]any{"node": id, "kind": kind, "text": text, "name": name})
			fmt.Println(jsx.IsIntrinsicElementNamed(n, name))
		}
	}
	observe(nil)
	parse := func(file, source string) {
		if !strings.HasPrefix(file, "/") {
			file = "/" + file
		}
		kind := core.ScriptKindTS
		if strings.HasSuffix(file, ".tsx") {
			kind = core.ScriptKindTSX
		}
		if strings.HasSuffix(file, ".jsx") {
			kind = core.ScriptKindJSX
		}
		parsed := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: file, Path: tspath.Path(file)}, source, kind)
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			observe(n)
			n.ForEachChild(walk)
			return false
		}
		walk(parsed.AsNode())
	}
	for scan.Scan() {
		var row struct{ File, Source string }
		must(json.Unmarshal(scan.Bytes(), &row))
		parse(row.File, row.Source)
	}
	must(scan.Err())
	parse("controls.tsx", `const x = <><a/><A/><Foo.a/><svg:a/><img/><Image/><script/><link/><é/><😀/></>; const a = 'a'; const escaped = \u0061;`)
	data, e := json.Marshal(map[string]any{"intrinsic": rows})
	must(e)
	must(os.WriteFile(os.Args[2], data, 0644))
}
