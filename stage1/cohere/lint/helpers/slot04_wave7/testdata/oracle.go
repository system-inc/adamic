package main

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"strings"
)

type Entry struct {
	Key  []int
	Kind string
}
type Row struct {
	Name, Source string
	Bytes, Root  []int
	Entries      []Entry
}
type Case struct {
	Root    []int
	Entries []Entry
}

func ints(s string) []int {
	result := []int{}
	for _, b := range []byte(s) {
		result = append(result, int(b))
	}
	return result
}
func text(v []int) string {
	result := []byte{}
	for _, b := range v {
		if b < 0 || b > 255 {
			panic("byte")
		}
		result = append(result, byte(b))
	}
	return string(result)
}
func observeCase(c Case) {
	names, kinds := []string{}, []string{}
	for _, e := range c.Entries {
		names = append(names, text(e.Key))
		kinds = append(kinds, e.Kind)
	}
	collapse.AdamicObserveVariants(names, kinds, text(c.Root))
}
func makeCase(input []byte) Case {
	value := ints(string(input))
	return Case{value, []Entry{{value, "static"}}}
}
func observe(input []byte) { observeCase(makeCase(input)) }
func main() {
	argument := os.Args[1]
	if argument == "--full" {
		observe(nil)
		for first := 0; first <= 255; first++ {
			observe([]byte{byte(first)})
			for last := 0; last <= 255; last++ {
				observe([]byte{byte(first), byte(last)})
			}
		}
		return
	}
	adapt := argument == "--cases"
	if adapt {
		argument = os.Args[2]
	}
	data, err := os.ReadFile(argument)
	if err != nil {
		panic(err)
	}
	var rows []Row
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	values := [][]byte{}
	cases := []Case{}
	for _, row := range rows {
		if row.Name == "control:call" {
			cases = append(cases, Case{row.Root, row.Entries})
			continue
		}
		if strings.HasPrefix(row.Name, "control:") {
			bytes := []byte{}
			for _, value := range row.Bytes {
				if value < 0 || value > 255 {
					panic("invalid byte control")
				}
				bytes = append(bytes, byte(value))
			}
			values = append(values, bytes)
			continue
		}
		values = append(values, []byte(row.Source))
		kind := core.ScriptKindTS
		switch {
		case strings.HasSuffix(row.Name, ".tsx"):
			kind = core.ScriptKindTSX
		case strings.HasSuffix(row.Name, ".jsx"):
			kind = core.ScriptKindJSX
		case strings.HasSuffix(row.Name, ".js"):
			kind = core.ScriptKindJS
		}
		path := tspath.NormalizePath("/corpus/" + row.Name)
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path, Path: tspath.Path(path)}, row.Source, kind)
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			if n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNoSubstitutionTemplateLiteral {
				values = append(values, []byte(n.Text()))
			}
			n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
		}
		visit(file.AsNode())
	}
	for _, value := range values {
		cases = append(cases, makeCase(value))
	}
	if adapt {
		if err = json.NewEncoder(os.Stdout).Encode(cases); err != nil {
			panic(err)
		}
		return
	}
	for _, c := range cases {
		observeCase(c)
	}
}
