package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
)

type Row struct{ Name, Source string }
type Node struct {
	Kind       string `json:"kind"`
	Text       string `json:"text"`
	Name       int    `json:"name"`
	TagName    int    `json:"tagName"`
	Attributes int    `json:"attributes"`
}

func main() {
	path := os.Args[1]
	adapt := path == "--ast"
	if adapt {
		path = os.Args[2]
	}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var rows []Row
	if err := json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	adapted := []map[string]any{}
	for _, row := range rows {
		kind := core.ScriptKindTS
		switch {
		case strings.HasSuffix(row.Name, ".tsx"):
			kind = core.ScriptKindTSX
		case strings.HasSuffix(row.Name, ".jsx"):
			kind = core.ScriptKindJSX
		case strings.HasSuffix(row.Name, ".js"):
			kind = core.ScriptKindJS
		}
		fileName := tspath.NormalizePath("/corpus/" + row.Name)
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: fileName, Path: tspath.Path(fileName)}, row.Source, kind)
		nodes := []*ast.Node{}
		ids := map[*ast.Node]int{nil: -1}
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			if _, ok := ids[n]; ok {
				return
			}
			ids[n] = len(nodes)
			nodes = append(nodes, n)
			n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
		index := func(node *ast.Node) int {
			value, exists := ids[node]
			if !exists {
				panic("AST field was absent from traversal")
			}
			return value
		}
		projected := []Node{}
		for _, n := range nodes {
			item := Node{Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), Name: -1, TagName: -1, Attributes: -1}
			// Projection reads raw AST fields, independently of the helper oracle.
			switch n.Kind {
			case ast.KindJsxOpeningElement:
				item.TagName = index(n.AsJsxOpeningElement().TagName)
				item.Attributes = index(n.AsJsxOpeningElement().Attributes)
			case ast.KindJsxSelfClosingElement:
				item.TagName = index(n.AsJsxSelfClosingElement().TagName)
				item.Attributes = index(n.AsJsxSelfClosingElement().Attributes)
			case ast.KindJsxAttribute:
				item.Name = index(n.AsJsxAttribute().Name())
			}
			if n.Kind == ast.KindIdentifier {
				item.Text = n.Text()
			}
			projected = append(projected, item)
			if !adapt {
				tag, attributes := jsx.ElementParts(n)
				fmt.Printf("parts %d %d\n", index(tag), index(attributes))
			}
		}
		if adapt {
			adapted = append(adapted, map[string]any{"name": row.Name, "ast": projected})
		}
	}
	if adapt {
		if err := json.NewEncoder(os.Stdout).Encode(adapted); err != nil {
			panic(err)
		}
	}
}
