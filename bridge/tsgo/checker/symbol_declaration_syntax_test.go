package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Raw declaration facts must retain syntax identity and initializer roles. The
// production checker, rather than any React classification, supplies the oracle.
func TestSymbolDeclarationSyntax(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.ts")
	config := filepath.Join(dir, "tsconfig.json")
	source := `import {Fragment as F} from 'react';const {Fragment:D}=React;const V=React.Fragment;F;D;V;missing;export {};`
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022"}}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	byRange := map[[3]int]*ast.Node{}
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		byRange[[3]int{n.Pos(), n.End(), int(n.Kind)}] = n
		n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
	}
	visit(sf.AsNode())
	checked, initializers, absent := 0, 0, false
	visit = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier && n.Parent.Kind == ast.KindExpressionStatement {
			wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "symbol-declaration-syntax")
			if err != nil {
				t.Fatal(err)
			}
			f := decodedFields(t, wire)
			at := 0
			take := func() string {
				if at >= len(f) {
					t.Fatal("truncated declaration graph")
				}
				v := f[at]
				at++
				return v
			}
			number := func() int {
				v, e := strconv.Atoi(take())
				if e != nil {
					t.Fatal(e)
				}
				return v
			}
			if take() != "1" || take() != "symbol-declaration-syntax" {
				t.Fatal("wrong protocol")
			}
			present := take()
			roots := number()
			for i := 0; i < roots; i++ {
				number()
			}
			count := number()
			type role struct {
				node        *ast.Node
				initializer int
			}
			var records []role
			for i := 0; i < count; i++ {
				if take() != file {
					t.Fatal("wrong declaration source")
				}
				kind := number()
				take()
				pos, end := number(), number()
				take()
				number()
				number()
				initializer := number()
				original := byRange[[3]int{pos, end, kind}]
				if original == nil || int(original.Kind) != kind {
					t.Fatalf("lost raw syntax identity: %d %d %d", kind, pos, end)
				}
				records = append(records, role{original, initializer})
				children := number()
				for j := 0; j < children; j++ {
					id := number()
					if id < 0 || id >= count {
						t.Fatal("invalid child index")
					}
				}
			}
			if at != len(f) {
				t.Fatal("trailing declaration fields")
			}
			for _, r := range records {
				if r.initializer != 0 {
					if r.node.Kind != ast.KindVariableDeclaration || r.initializer > len(records) || records[r.initializer-1].node != r.node.AsVariableDeclaration().Initializer {
						t.Fatal("wrong initializer role")
					}
					initializers++
				}
			}
			if n.Text() == "missing" {
				if present != "0" || roots != 0 || count != 0 {
					t.Fatal("fabricated missing declaration")
				}
				absent = true
			} else {
				if present != "1" || roots == 0 || count == 0 {
					t.Fatal("lost binding declaration")
				}
				checked++
			}
			if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "symbol-declaration-syntax\nextra"); err == nil {
				t.Fatal("accepted question suffix")
			}
		}
		n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
	}
	visit(sf.AsNode())
	if checked != 3 || initializers < 2 || !absent {
		t.Fatalf("incomplete controls: %d %d %v", checked, initializers, absent)
	}
}
