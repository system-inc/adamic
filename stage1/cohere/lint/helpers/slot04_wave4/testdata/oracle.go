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

func observe(input []byte) {
	stack := []byte{}
	fmt.Printf("top %d\n", collapse.AdamicTopOfStack(stack))
	for _, index := range []int{-2147483648, -1, len(input), len(input) + 1, 2147483647} {
		fmt.Printf("peek %d %d\n", index, collapse.AdamicPeekByte(string(input), index))
	}
	for index, value := range input {
		fmt.Printf("peek %d %d\n", index, collapse.AdamicPeekByte(string(input), index))
		stack = append(stack, value)
		fmt.Printf("top %d\n", collapse.AdamicTopOfStack(stack))
	}
	if len(stack) > 0 {
		stack[len(stack)-1] = 137
		fmt.Printf("changed %d\n", collapse.AdamicTopOfStack(stack))
	}
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
