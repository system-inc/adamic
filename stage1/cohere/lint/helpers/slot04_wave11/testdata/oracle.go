package main

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	engine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
)

type Input struct {
	Name, Source, Helper string
	Data                 []int
	Prefixes             [][]int
}

func main() {
	path := os.Args[len(os.Args)-1]
	if path == "--full" {
		engine.AdamicFull()
		return
	}
	inputs := []Input{}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, &inputs); err != nil {
		panic(err)
	}
	adapted := len(os.Args) > 2 && os.Args[1] == "--cases"
	rows := []engine.AdamicBytesCase{}
	for _, in := range inputs {
		if in.Helper != "" {
			rows = append(rows, engine.AdamicBytesCase{Name: in.Name, Helper: in.Helper, Data: in.Data, Prefixes: in.Prefixes})
			continue
		}
		add := func(value string) {
			c := engine.AdamicCase(value, []string{"", "[", "[length]", "calc(", "url(", "var("})
			c.Name = in.Name
			rows = append(rows, c)
			c.Prefixes = c.Prefixes[1:]
			rows = append(rows, c)
		}
		add(in.Source)
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/repository/source/Thing.tsx", Path: tspath.Path("/repository/source/Thing.tsx")}, in.Source, core.ScriptKindTSX)
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			switch n.Kind {
			case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
				add(n.Text())
			}
			n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
	}
	if adapted {
		if err = json.NewEncoder(os.Stdout).Encode(rows); err != nil {
			panic(err)
		}
	} else {
		for _, row := range rows {
			engine.AdamicObserve(row)
		}
	}
}
