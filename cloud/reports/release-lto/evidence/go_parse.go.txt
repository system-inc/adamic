package main

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"runtime"
	"strings"
)

func main() {
	manifest, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	for _, row := range strings.Split(string(manifest), "\n") {
		if row == "" {
			continue
		}
		path := strings.Split(row, "\t")[0]
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path, Path: tspath.Path(path)}, string(data), core.ScriptKindTS)
		runtime.KeepAlive(file.ECMALineMap())
		runtime.KeepAlive(file)
	}
	fmt.Println(0)
}
