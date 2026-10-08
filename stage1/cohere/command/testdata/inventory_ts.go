//go:build ignore

// Raw declaration/body identity includes methods. It is a duplication candidate,
// not permission to merge different modules' nominal classes or imported names.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
)

type occurrence struct{ Name, Location string }

func main() {
	root, e := filepath.Abs(".")
	if e != nil {
		panic(e)
	}
	p, e := load.Load([]string{filepath.Join(root, "stage1/cohere/command/main.a")})
	if e != nil {
		panic(e)
	}
	declarations := map[string][]occurrence{}
	bodies := map[string][]occurrence{}
	count := 0
	for _, file := range p.CompilerProgram().GetSourceFiles() {
		name := p.FileName(file)
		if !strings.HasPrefix(name, filepath.Join(root, "stage1")) {
			continue
		}
		relative, e := filepath.Rel(root, name)
		if e != nil {
			panic(e)
		}
		source := file.Text()
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			switch n.Kind {
			case ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				body := n.Body()
				if body != nil {
					count++
					label := "anonymous"
					if n.Name() != nil {
						label = n.Name().Text()
					}
					pos := n.Pos()
					s := occurrence{label, fmt.Sprintf("%s:%d", relative, 1+strings.Count(source[:pos], "\n"))}
					decl := strings.TrimSpace(source[pos:n.End()])
					bodyText := strings.TrimSpace(source[body.Pos():body.End()])
					a := fmt.Sprintf("%x", sha256.Sum256([]byte(decl)))
					b := fmt.Sprintf("%x", sha256.Sum256([]byte(bodyText)))
					declarations[a] = append(declarations[a], s)
					bodies[b] = append(bodies[b], s)
				}
			}
			n.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
	}
	for hash, sites := range declarations {
		if len(sites) < 2 {
			delete(declarations, hash)
		}
	}
	for hash, sites := range bodies {
		if len(sites) < 2 {
			delete(bodies, hash)
		}
	}
	data, e := json.MarshalIndent(struct {
		CallableSites                          int
		IdenticalDeclarations, IdenticalBodies map[string][]occurrence
	}{count, declarations, bodies}, "", "  ")
	if e != nil {
		panic(e)
	}
	if e = os.WriteFile("stage1/cohere/command/testdata/ts-helper-inventory.json", append(data, '\n'), 0644); e != nil {
		panic(e)
	}
	fmt.Printf("TS callable sites=%d identical declaration groups=%d identical body groups=%d\n", count, len(declarations), len(bodies))
}
