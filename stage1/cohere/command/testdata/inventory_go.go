//go:build ignore

// Exact Go AST inventories; all sites are upstream, none are port decisions.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type site struct{ Name, Location, Declaration string }

func main() {
	fset := token.NewFileSet()
	variables := []site{}
	functions := map[string][]site{}
	duplicates := map[string][]site{}
	for _, root := range []string{"cohere/command/cohere", "cohere/internal/format", "cohere/internal/lint"} {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.IsDir() {
				if d.Name() == "testdata" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, e := parser.ParseFile(fset, path, nil, 0)
			if e != nil {
				return e
			}
			for _, decl := range file.Decls {
				var rendered bytes.Buffer
				if e := format.Node(&rendered, fset, decl); e != nil {
					return e
				}
				switch n := decl.(type) {
				case *ast.GenDecl:
					if n.Tok != token.VAR {
						continue
					}
					for _, spec := range n.Specs {
						value := spec.(*ast.ValueSpec)
						for _, name := range value.Names {
							variables = append(variables, site{name.Name, fmt.Sprintf("%s:%d", path, fset.Position(name.Pos()).Line), rendered.String()})
						}
					}
				case *ast.FuncDecl:
					if n.Recv != nil {
						continue
					}
					s := site{n.Name.Name, fmt.Sprintf("%s:%d", path, fset.Position(n.Pos()).Line), rendered.String()}
					functions[n.Name.Name] = append(functions[n.Name.Name], s)
					hash := fmt.Sprintf("%x", sha256.Sum256(rendered.Bytes()))
					duplicates[hash] = append(duplicates[hash], s)
				}
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	for name, sites := range functions {
		if len(sites) < 2 {
			delete(functions, name)
		}
	}
	for hash, sites := range duplicates {
		if len(sites) < 2 {
			delete(duplicates, hash)
		}
	}
	data, e := json.MarshalIndent(struct {
		Variables          []site
		RepeatedFunctions  map[string][]site
		IdenticalFunctions map[string][]site
	}{variables, functions, duplicates}, "", "  ")
	if e != nil {
		panic(e)
	}
	if e = os.WriteFile("stage1/cohere/command/testdata/go-inventory.json", append(data, '\n'), 0644); e != nil {
		panic(e)
	}
	fmt.Printf("Go variables=%d repeated function names=%d identical function groups=%d\n", len(variables), len(functions), len(duplicates))
}
