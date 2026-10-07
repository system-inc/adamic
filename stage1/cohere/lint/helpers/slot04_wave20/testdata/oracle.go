package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rules/structure"
	"os"
)

type Input struct {
	Name, Source  string
	Nil, RootOnly bool
	Depths        []int
}
type Node struct {
	Present, Jsx, Call, Switch, Hook bool
	Children                         []int
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
		if !c.Nil {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/repository/source/Thing.tsx", Path: tspath.Path("/repository/source/Thing.tsx")}, c.Source, core.ScriptKindTSX)
			var visit func(*ast.Node) int
			visit = func(node *ast.Node) int {
				if node == nil {
					return -1
				}
				id := len(actual)
				actual = append(actual, node)
				v := Node{Present: true, Children: []int{}}
				switch node.Kind {
				case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
					v.Jsx = true
				case ast.KindCallExpression:
					v.Call = true
					v.Hook = react.IsHookCall(node.AsCallExpression())
				case ast.KindSwitchStatement:
					v.Switch = true
				}
				row.Nodes = append(row.Nodes, v)
				children := []int{}
				node.ForEachChild(func(child *ast.Node) bool { children = append(children, visit(child)); return false })
				row.Nodes[id].Children = children
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
			fmt.Printf("%t|%t", structure.HasJsxOrReactHookCalls(node), structure.AdamicDescends(node))
			for _, depth := range depths {
				fmt.Printf("|%t", structure.AdamicSearch(node, depth))
			}
			fmt.Println()
		}
	}
	if adapted {
		if e = json.NewEncoder(os.Stdout).Encode(rows); e != nil {
			panic(e)
		}
	}
}
