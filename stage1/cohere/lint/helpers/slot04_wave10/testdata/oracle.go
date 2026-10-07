package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"os"
)

type Input struct {
	Name, Source        string
	Manual, ListPresent bool
	Tags                []string
}
type Row struct {
	Name                 string
	Present, ListPresent bool
	Tags                 []string
}

func main() {
	inputs := []Input{}
	data, err := os.ReadFile(os.Args[len(os.Args)-1])
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, &inputs); err != nil {
		panic(err)
	}
	adapted := len(os.Args) > 2 && os.Args[1] == "--cases"
	rows := []Row{}
	observe := func(name string, node *ast.Node, mods *ast.ModifierList, tags []string) {
		if adapted {
			rows = append(rows, Row{name, node != nil, mods != nil, tags})
			return
		}
		fmt.Printf("%t|%t|%t\n", module.IsExported(node), module.HasExportModifier(mods), module.IsExportedByName(mods))
	}
	for _, in := range inputs {
		if in.Manual {
			var mods *ast.ModifierList
			if in.ListPresent {
				mods = &ast.ModifierList{}
				for _, tag := range in.Tags {
					kind := ast.KindIdentifier
					if tag == "export-keyword" {
						kind = ast.KindExportKeyword
					}
					if tag == "default-keyword" {
						kind = ast.KindDefaultKeyword
					}
					mods.Nodes = append(mods.Nodes, &ast.Node{Kind: kind})
				}
			}
			observe(in.Name, nil, mods, in.Tags)
			continue
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/repository/source/Thing.tsx", Path: tspath.Path("/repository/source/Thing.tsx")}, in.Source, core.ScriptKindTSX)
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			if node == nil {
				return
			}
			mods := node.Modifiers()
			tags := []string{}
			if mods != nil {
				for _, modifier := range mods.Nodes {
					tag := "other"
					if modifier.Kind == ast.KindExportKeyword {
						tag = "export-keyword"
					}
					if modifier.Kind == ast.KindDefaultKeyword {
						tag = "default-keyword"
					}
					tags = append(tags, tag)
				}
			}
			observe(in.Name, node, mods, tags)
			node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
	}
	if adapted {
		if err = json.NewEncoder(os.Stdout).Encode(rows); err != nil {
			panic(err)
		}
	}
}
