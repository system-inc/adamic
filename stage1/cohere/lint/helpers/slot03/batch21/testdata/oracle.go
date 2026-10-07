package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	imports "github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	next "github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

func bytesOf(text string) []int {
	out := []int{}
	for _, b := range []byte(text) {
		out = append(out, int(b))
	}
	return out
}
func units(text string) string {
	out := ""
	for _, b := range []byte(text) {
		out += strconv.Itoa(int(b)) + ","
	}
	return out
}
func bit(v bool) int {
	if v {
		return 1
	}
	return 0
}
func main() {
	f, e := os.Open(os.Args[1])
	if e != nil {
		panic(e)
	}
	z, e := gzip.NewReader(f)
	if e != nil {
		panic(e)
	}
	d := json.NewDecoder(z)
	texts := map[string]bool{}
	sources := 0
	for {
		var row struct{ Rule, File, Source string }
		e = d.Decode(&row)
		if e == io.EOF {
			break
		}
		if e != nil {
			panic(e)
		}
		sources++
		texts[row.File] = true
		texts["/"+row.File] = true
		texts[row.Source] = true
		kind := core.ScriptKindTS
		if strings.HasSuffix(row.File, "x") {
			kind = core.ScriptKindTSX
		}
		filename := path.Clean("/" + row.File)
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: filename, Path: tspath.Path(filename)}, row.Source, kind)
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			if ast.IsStringLiteralLike(n) {
				texts[n.Text()] = true
			}
			return n.ForEachChild(walk)
		}
		walk(file.AsNode())
	}
	z.Close()
	f.Close()
	for _, s := range []string{"", "x", "/", "\\", "a/b\\c", "a\\b/c", "a/", "a\\", "/a", "\\a", "a//b", "a\\\\b", "é/文\\😀/x", "😀\\x", "\xff/\xfe\\x", "\x00/\x00", "internal", "internalization", "x/internal/y", "x\\internal\\y", "_translations", "translations", "x/translations/"} {
		texts[s] = true
	}
	keys := []string{}
	for s := range texts {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	rows := []any{}
	outputs := []string{}
	for _, text := range keys {
		base, parent, trace := next.AdamicSplitPath(text)
		rows = append(rows, map[string]any{"op": "split", "path": bytesOf(text)})
		outputs = append(outputs, units(base)+"/"+units(parent)+"/"+trace)
		for _, segment := range []string{"", "internal", "_translations", "translations", "x", "😀", "\xff", "a/b", "a\\b", "é"} {
			rows = append(rows, map[string]any{"op": "segment", "path": bytesOf(text), "segment": bytesOf(segment)})
			outputs = append(outputs, strconv.Itoa(bit(imports.HasPathSegment(text, segment))))
		}
	}
	arms := collapse.AdamicArmCases()
	for _, c := range arms {
		rows = append(rows, c)
		outputs = append(outputs, strconv.Itoa(bit(c.Want)))
	}
	raw, e := json.Marshal(rows)
	if e != nil {
		panic(e)
	}
	if e = os.WriteFile(os.Args[2], raw, 0644); e != nil {
		panic(e)
	}
	stats, _ := json.Marshal(map[string]int{"sources": sources, "path_inputs": len(keys), "arm_inputs": len(arms), "total_lines": len(outputs)})
	if e = os.WriteFile(os.Args[2]+".stats.json", stats, 0644); e != nil {
		panic(e)
	}
	for _, output := range outputs {
		fmt.Println(output)
	}
}
