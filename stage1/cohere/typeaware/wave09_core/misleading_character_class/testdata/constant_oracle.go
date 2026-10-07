package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"os"
	"strings"
	"unicode/utf16"
)

func written(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != 92 {
			out.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			h, l := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, h, l)
		}
	}
	return out.String()
}

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows []string
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for _, expression := range rows {
		source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/fixture.ts", Path: tspath.Path("/fixture.ts")}, "const value="+expression+";", core.ScriptKindTS)
		if len(source.Diagnostics()) != 0 {
			panic(expression)
		}
		var initializer *ast.Node
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindVariableDeclaration {
				initializer = node.AsVariableDeclaration().Initializer
				return true
			}
			return node.ForEachChild(visit)
		}
		source.AsNode().ForEachChild(visit)
		text, known := reference.ConstantString(initializer)
		fmt.Printf("%t\t%s\n", known, written(text))
	}
}
