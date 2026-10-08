package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCheckerMapperAliasKeepsIdentity(t *testing.T) {
	// Read the actual bridge declaration, so restoring its opaque struct changes
	// what the generator sees instead of merely changing a synthetic descriptor.
	file, err := parser.ParseFile(token.NewFileSet(), "../../../../internal/lower/instantiate.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	g := generator{types: map[string]ast.Expr{}, needed: map[string]bool{}}
	for _, declaration := range file.Decls {
		if group, ok := declaration.(*ast.GenDecl); ok {
			for _, spec := range group.Specs {
				if spec, ok := spec.(*ast.TypeSpec); ok {
					g.types[spec.Name.Name] = spec.Type
				}
			}
		}
	}
	if got := g.copy(&ast.StarExpr{X: ast.NewIdent("typeMapper")}, "value.typeMapper"); got != "value.typeMapper" {
		t.Fatalf("checker mapper snapshot must preserve identity, got %s", got)
	}
	if g.needed["typeMapper"] {
		t.Fatal("checker mapper must not be deep copied as lowering-owned state")
	}
}
