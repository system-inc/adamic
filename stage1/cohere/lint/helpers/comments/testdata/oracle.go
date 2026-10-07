package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func written(text string) string {
	var result strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != '\\' {
			result.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&result, `\u%04x`, r)
		} else {
			h, l := utf16.EncodeRune(r)
			fmt.Fprintf(&result, `\u%04x\u%04x`, h, l)
		}
	}
	return result.String()
}
func main() {
	path := os.Args[1]
	astMode := path == "--ast"
	if astMode {
		path = os.Args[2]
	}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var corpus []struct{ Name, Source string }
	if err = json.Unmarshal(data, &corpus); err != nil {
		panic(err)
	}
	var adapted []map[string]any
	for _, row := range corpus {
		if !astMode {
			fmt.Println("case " + written(row.Name))
		}
		// Match the original rule harness's ScriptKind selection exactly.
		kind := core.ScriptKindTS
		switch {
		case strings.HasSuffix(row.Name, ".tsx"):
			kind = core.ScriptKindTSX
		case strings.HasSuffix(row.Name, ".jsx"):
			kind = core.ScriptKindJSX
		case strings.HasSuffix(row.Name, ".js"):
			kind = core.ScriptKindJS
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/corpus/" + row.Name, Path: tspath.Path("/corpus/" + row.Name)}, row.Source, kind)
		if astMode {
			offsets := make(map[int]int)
			unit := 0
			for position, point := range row.Source {
				offsets[position] = unit
				if point > 65535 {
					unit += 2
				} else {
					unit++
				}
				offsets[position+utf8.RuneLen(point)] = unit
			}
			offsets[len(row.Source)] = unit
			var nodes []map[string]any
			var visit func(*ast.Node) int
			visit = func(node *ast.Node) int {
				children := []int{}
				node.ForEachChild(func(child *ast.Node) bool { children = append(children, visit(child)); return false })
				nodes = append(nodes, map[string]any{"kind": strings.TrimPrefix(node.Kind.String(), "Kind"), "pos": offsets[node.Pos()], "end": offsets[node.End()], "children": children})
				return len(nodes) - 1
			}
			root := visit(file.AsNode())
			adapted = append(adapted, map[string]any{"name": row.Name, "source": row.Source, "ast": nodes, "root": root, "goDiagnostics": len(file.Diagnostics())})
			continue
		}
		ctx := rule.Context{SourceFile: file, FileCache: rule.NewFileCache()}
		first, second := comments.ForFile(ctx), comments.ForFile(ctx)
		shared := 1
		if len(first) > 0 && &first[0] != &second[0] {
			shared = 0
		}
		fmt.Printf("cache %d %d %d\n", shared, len(comments.All(nil)), len(comments.ForFile(rule.Context{})))
		fmt.Printf("cached %d %d\n", len(first), len(comments.ForFile(rule.Context{SourceFile: file})))
		for _, c := range first {
			block := 0
			if c.IsBlock {
				block = 1
			}
			fmt.Printf("%d %d %d %d %d %d\t%s\n", c.Range.Pos(), c.Range.End(), block, c.StartLine, c.StartColumn, c.EndLine, written(c.Text))
		}
		all := comments.All(file)
		fmt.Printf("comments %d\n", len(all))
		for _, c := range all {
			block := 0
			if c.IsBlock {
				block = 1
			}
			fmt.Printf("%d %d %d %d %d %d\t%s\n", c.Range.Pos(), c.Range.End(), block, c.StartLine, c.StartColumn, c.EndLine, written(c.Text))
		}
		for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
			all[i], all[j] = all[j], all[i]
		}
		comments.AdamicSortByPosition(all)
		var positions []string
		for _, c := range all {
			positions = append(positions, fmt.Sprint(c.Range.Pos()))
		}
		fmt.Println("sorted " + strings.Join(positions, ","))
		var guard strings.Builder
		for p := range row.Source {
			if comments.AdamicCanBeginAt(row.Source, p) {
				guard.WriteByte('1')
			} else {
				guard.WriteByte('0')
			}
		}
		guard.WriteByte('0')
		fmt.Println("guard " + guard.String())
	}
	if astMode {
		if err := json.NewEncoder(os.Stdout).Encode(adapted); err != nil {
			panic(err)
		}
	}
}
