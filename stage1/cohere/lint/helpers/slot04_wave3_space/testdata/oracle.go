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

type Row struct{ Name, Source string }

func main() {
	argument := os.Args[1]
	if argument == "--full" {
		for value := rune(0); value <= 0x10ffff; value++ {
			fmt.Printf("rune %d %t %t\n", value, collapse.AdamicJavaScriptSpace(value), collapse.AdamicBlank(string(value)))
		}
		for _, value := range []rune{-2147483648, -1, 0x110000, 2147483647} {
			fmt.Printf("outside %d %t\n", value, collapse.AdamicJavaScriptSpace(value))
		}
		for value := 0; value <= 255; value++ {
			fmt.Printf("byte %d %t\n", value, collapse.AdamicValueSeparator(byte(value)))
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
	values := []string{}
	for _, row := range rows {
		values = append(values, row.Source)
		if strings.HasPrefix(row.Name, "control:") {
			continue
		}
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
				values = append(values, n.Text())
			}
			n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
		}
		visit(file.AsNode())
	}
	if adapt {
		if err = json.NewEncoder(os.Stdout).Encode(values); err != nil {
			panic(err)
		}
		return
	}
	for _, value := range values {
		fmt.Printf("blank %t\n", collapse.AdamicBlank(value))
		for _, r := range value {
			fmt.Printf("rune %d %t\n", r, collapse.AdamicJavaScriptSpace(r))
		}
		for _, b := range []byte(value) {
			fmt.Printf("byte %d %t\n", b, collapse.AdamicValueSeparator(b))
		}
	}
}
