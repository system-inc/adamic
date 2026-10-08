package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rules/structure"
	"os"
	"strings"
)

type Input struct {
	Name, Source  string
	Nil, RootOnly bool
	Depths        []int
}
type Node struct {
	Present, Jsx, Call, Switch bool
	Kind, Text                 string
	Expression, Name           int
	Children                   []int
}
type Row struct {
	Name        string
	Nodes       []Node
	Ids, Depths []int
}

func main() {
	data, e := os.ReadFile(os.Args[len(os.Args)-1])
	if e != nil {
		panic(e)
	}
	var cases []Input
	if e = json.Unmarshal(data, &cases); e != nil {
		panic(e)
	}
	rows := []Row{}
	adapted := len(os.Args) > 2 && os.Args[1] == "--cases"
	for _, c := range cases {
		depths := c.Depths
		if len(depths) == 0 {
			depths = []int{-1, 0, 19, 20, 21}
		}
		row := Row{Name: c.Name, Nodes: []Node{}, Ids: []int{-1}, Depths: depths}
		actual := []*ast.Node{}
		ids := map[*ast.Node]int{}
		if !c.Nil {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute("/repository/source/Thing.tsx"), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute("/repository/source/Thing.tsx"))}, c.Source, core.ScriptKindTSX)
			var visit func(*ast.Node) int
			visit = func(node *ast.Node) int {
				if node == nil {
					return -1
				}
				if id, ok := ids[node]; ok {
					return id
				}
				id := len(actual)
				ids[node] = id
				actual = append(actual, node)
				v := Node{Present: true, Kind: strings.TrimPrefix(node.Kind.String(), "Kind"), Expression: -1, Name: -1, Children: []int{}}
				if node.Kind == ast.KindIdentifier {
					v.Text = node.Text()
				}
				switch node.Kind {
				case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
					v.Jsx = true
				case ast.KindCallExpression:
					v.Call = true
				case ast.KindSwitchStatement:
					v.Switch = true
				}
				row.Nodes = append(row.Nodes, v)
				children := []int{}
				node.ForEachChild(func(child *ast.Node) bool { children = append(children, visit(child)); return false })
				row.Nodes[id].Children = children
				switch node.Kind {
				case ast.KindCallExpression:
					row.Nodes[id].Expression = visit(node.AsCallExpression().Expression)
				case ast.KindPropertyAccessExpression:
					row.Nodes[id].Expression = visit(node.AsPropertyAccessExpression().Expression)
					row.Nodes[id].Name = visit(node.AsPropertyAccessExpression().Name())
				case ast.KindParenthesizedExpression:
					row.Nodes[id].Expression = visit(node.AsParenthesizedExpression().Expression)
				}
				return id
			}
			visit(file.AsNode())
			for id := range actual {
				if !c.RootOnly || id == 0 {
					row.Ids = append(row.Ids, id)
				}
			}
		}
		if adapted {
			rows = append(rows, row)
			continue
		}
		for _, id := range row.Ids {
			var node *ast.Node
			if id >= 0 {
				node = actual[id]
			}
			fmt.Println(structure.HasJsxOrReactHookCalls(node))
			for _, depth := range row.Depths {
				fmt.Println(structure.AdamicSearch(node, depth))
			}
			fmt.Println(structure.AdamicDescends(node))
		}
	}
	if adapted {
		if e = json.NewEncoder(os.Stdout).Encode(rows); e != nil {
			panic(e)
		}
	}
}
