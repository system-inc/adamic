// Records raw bridge facts for executing the owned Adamic source on Node.
// No production lint rule is called here. The findings oracle is a separate binary.
package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		panic("usage: record_context config manifest output.json.gz")
	}
	data, err := os.ReadFile(args[1])
	if err != nil {
		panic(err)
	}
	var paths []string
	for _, path := range strings.Split(string(data), "\n") {
		if path != "" {
			paths = append(paths, path)
		}
	}
	p, err := bridge.Open(args[0], paths)
	if err != nil {
		panic(err)
	}
	records := map[string]string{}
	for _, file := range p.Compiler.GetSourceFiles() {
		if file.IsDeclarationFile {
			continue
		}
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			kind := strings.TrimPrefix(node.Kind.String(), "Kind")
			modes := []string{"syntax"}
			if node.Kind == ast.KindIdentifier {
				modes = append(modes, "origin", "alias-origin")
			}
			if node.Kind == ast.KindCallExpression {
				modes = append(modes, "signature")
			}
			if node.Kind == ast.KindSourceFile {
				modes = append(modes, "program")
			}
			for _, mode := range modes {
				question := "runtime-context\n" + mode
				wire, err := p.Inspect(file.FileName(), uint64(node.Pos()), uint64(node.End()), kind, question)
				if err != nil {
					panic(err)
				}
				key := fmt.Sprintf("%s\t%d\t%d\t%s\t%s", file.FileName(), node.Pos(), node.End(), kind, question)
				records[key] = wire
			}
			node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(file.AsNode())
	}
	output, err := os.Create(args[2])
	if err != nil {
		panic(err)
	}
	compressed := gzip.NewWriter(output)
	encoder := json.NewEncoder(compressed)
	for key, wire := range records {
		if err := encoder.Encode([2]string{key, wire}); err != nil {
			panic(err)
		}
	}
	if err := compressed.Close(); err != nil {
		panic(err)
	}
	if err := output.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("%d raw fact records\n", len(records))
}
