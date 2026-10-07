package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"strings"
)

type Row struct {
	Name, Source string
	Bytes        []int
}

func byteText(s string) string {
	pieces := []string{}
	for _, b := range []byte(s) {
		pieces = append(pieces, fmt.Sprint(b))
	}
	return strings.Join(pieces, ",")
}
func leaf(n *collapse.Node) {
	fmt.Printf("node %s|%s|%s|%s|%s|%s|%t|%t|%t|%t\n", n.Kind, byteText(n.Selector), byteText(n.Name), byteText(n.Params), byteText(n.Property), byteText(n.Value), n.ValuePresent, n.Important, n.Context == nil, n.Nodes == nil)
}
func observe(input []byte) {
	value := string(input)
	fmt.Printf("bucket %s\n", byteText(collapse.AdamicBreakpointBucket(value)))
	comment := collapse.Comment(value)
	other := collapse.Comment(value)
	leaf(comment)
	comment.Value = ""
	comment.Important = true
	leaf(other)
	declaration := collapse.Declaration(value, value)
	fresh := collapse.Declaration(value, value)
	leaf(declaration)
	declaration.ValuePresent = false
	declaration.Property = ""
	leaf(fresh)
	leaf(declaration)
}
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
	for _, row := range rows {
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
	if adapt {
		output := [][]int{}
		for _, value := range values {
			row := []int{}
			for _, b := range value {
				row = append(row, int(b))
			}
			output = append(output, row)
		}
		if err = json.NewEncoder(os.Stdout).Encode(output); err != nil {
			panic(err)
		}
		return
	}
	for _, value := range values {
		observe(value)
	}
}
