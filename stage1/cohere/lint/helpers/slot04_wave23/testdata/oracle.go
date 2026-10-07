package main

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	cfg "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"os"
)

type Input struct {
	Name, Source string
	Nil          bool
}
type Row struct {
	Name  string
	Nodes []cfg.AdamicNode
}

func main() {
	data, e := os.ReadFile(os.Args[len(os.Args)-1])
	if e != nil {
		panic(e)
	}
	var inputs []Input
	if e = json.Unmarshal(data, &inputs); e != nil {
		panic(e)
	}
	rows := []Row{}
	for _, input := range inputs {
		nodes := []*ast.Node{}
		ids := map[*ast.Node]int{}
		if !input.Nil {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/repository/source/Thing.tsx", Path: tspath.Path("/repository/source/Thing.tsx")}, input.Source, core.ScriptKindTSX)
			var visit func(*ast.Node)
			visit = func(n *ast.Node) {
				if n == nil {
					return
				}
				ids[n] = len(nodes)
				nodes = append(nodes, n)
				n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
			}
			visit(file.AsNode())
		}
		if len(os.Args) > 2 && os.Args[1] == "--cases" {
			row := Row{Name: input.Name, Nodes: []cfg.AdamicNode{}}
			for _, n := range nodes {
				row.Nodes = append(row.Nodes, cfg.AdamicView(n, ids))
			}
			rows = append(rows, row)
		} else {
			cfg.AdamicObserve(nodes, ids)
		}
	}
	if len(os.Args) > 2 && os.Args[1] == "--cases" {
		if e = json.NewEncoder(os.Stdout).Encode(rows); e != nil {
			panic(e)
		}
	}
}
