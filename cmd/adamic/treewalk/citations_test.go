package treewalk

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// docs/cli.md cites Go by symbol, never by line, so a citation goes stale only when its declaration does, and then
// this fails on the change that moved it rather than on whichever change next reads the page. Every `path.go:Name`
// must name a top-level declaration of that file (a function, Type.Method, a type, var or const), and every Go file the
// page names must exist. The page cites Go wherever it lives, outside cmd/adamic's closure too, and this package reads
// every Go file, so an edit that breaks a citation reruns it (#hbq28nr). cmd/adamic checks the page's other citations.
// Mutants: rename a cited function, cite a Go file by line, or name a Go file that isn't there.
func TestDocumentedGoCitationsResolve(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..", "..")
	page, err := os.ReadFile(filepath.Join(root, "docs", "cli.md"))
	if err != nil {
		t.Fatal(err)
	}
	declared := map[string]map[string]bool{}
	symbols := 0
	for _, match := range Citation.FindAllStringSubmatch(string(page), -1) {
		path, anchor := match[1], match[2]
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		if anchor == "" {
			if _, err := os.Stat(filepath.Join(root, path)); err != nil {
				t.Errorf("docs/cli.md names %s, which isn't in the tree", path)
			}
			continue
		}
		if anchor[0] >= '0' && anchor[0] <= '9' {
			t.Errorf("docs/cli.md cites %s:%s by line; cite Go by the symbol, as path.go:Name", path, anchor)
			continue
		}
		names, ok := declared[path]
		if !ok {
			names = declarationsOf(t, filepath.Join(root, path))
			declared[path] = names
		}
		for _, name := range strings.Split(anchor, ",") {
			symbols++
			if !names[name] {
				t.Errorf("docs/cli.md cites %s:%s, which %s doesn't declare", path, name, path)
			}
		}
	}
	// A page with no symbol cites would pass every check above, so it fails here instead.
	if symbols == 0 {
		t.Fatal("found no Go symbol citations in docs/cli.md; the pattern looks at the wrong text")
	}
}

// declarationsOf names every top-level declaration of a Go file, methods as Type.Method.
func declarationsOf(t *testing.T, path string) map[string]bool {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			name := declaration.Name.Name
			if declaration.Recv != nil && len(declaration.Recv.List) == 1 {
				receiver := declaration.Recv.List[0].Type
				if star, ok := receiver.(*ast.StarExpr); ok {
					receiver = star.X
				}
				if index, ok := receiver.(*ast.IndexExpr); ok {
					receiver = index.X
				}
				if identifier, ok := receiver.(*ast.Ident); ok {
					name = identifier.Name + "." + name
				}
			}
			names[name] = true
		case *ast.GenDecl:
			for _, spec := range declaration.Specs {
				switch spec := spec.(type) {
				case *ast.ValueSpec:
					for _, identifier := range spec.Names {
						names[identifier.Name] = true
					}
				case *ast.TypeSpec:
					names[spec.Name.Name] = true
				}
			}
		}
	}
	return names
}
