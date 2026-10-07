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
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
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
	edgesRows := []map[string]any{}
	edgesWant := []string{}
	observeEdges := func(before, after string, first, last, leading, trailing bool) {
		edgesRows = append(edgesRows, map[string]any{"before": before, "after": after, "first": fmt.Sprint(first), "last": fmt.Sprint(last), "leading": fmt.Sprint(leading), "trailing": fmt.Sprint(trailing)})
		a, b := tailwind.AdamicHoleEdges(before, after, first, last, leading, trailing)
		edgesWant = append(edgesWant, fmt.Sprintf("%t,%t", a, b))
	}
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
			if n.Kind == ast.KindTemplateExpression {
				template := n.AsTemplateExpression()
				if template.Head != nil && template.TemplateSpans != nil {
					before := template.Head.Text()
					spans := template.TemplateSpans.Nodes
					for i, spanNode := range spans {
						span := spanNode.AsTemplateSpan()
						if span != nil && span.Literal != nil {
							after := span.Literal.Text()
							for flags := 0; flags < 4; flags++ {
								observeEdges(before, after, i == 0, i == len(spans)-1, flags&1 != 0, flags&2 != 0)
							}
							before = after
						}
					}
				}
			}
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
	for _, before := range []string{"", "x", " ", "\t", "\n", "\r", "\v", "\f", "x ", " x", "é", "😀", "\u00a0", "\u2003", "\x00"} {
		for _, after := range []string{"", "x", " ", "\t", "\n", "\r", "\v", "\f", "x ", " x", "é", "😀", "\u00a0", "\u2003", "\x00"} {
			for flags := 0; flags < 16; flags++ {
				observeEdges(before, after, flags&1 != 0, flags&2 != 0, flags&4 != 0, flags&8 != 0)
			}
		}
	}
	for _, line := range edgesWant {
		fmt.Println(line)
	}
	data, e := json.Marshal(map[string]any{"intrinsic": rows, "edges": edgesRows})
	must(e)
	must(os.WriteFile(os.Args[2], data, 0644))
}
