package ir

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestTypeTagsComplete(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "ir.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	tags := TypeTags()
	found := map[string]bool{}
	for _, declaration := range file.Decls {
		block, ok := declaration.(*ast.GenDecl)
		if !ok || block.Tok != token.CONST {
			continue
		}
		representation := false
		for _, spec := range block.Specs {
			value := spec.(*ast.ValueSpec)
			if value.Type != nil {
				identifier, ok := value.Type.(*ast.Ident)
				representation = ok && identifier.Name == "Type"
			}
			if representation {
				for _, name := range value.Names {
					found[name.Name] = true
					if _, ok := tags[name.Name]; !ok {
						t.Errorf("missing fingerprint tag %s", name.Name)
					}
				}
			}
		}
	}
	for name := range tags {
		if !found[name] {
			t.Errorf("unknown fingerprint tag %s", name)
		}
	}
}

func TestTypeTagFingerprintChanges(t *testing.T) {
	original := TypeTagFingerprint()
	for name := range TypeTags() {
		tags := TypeTags()
		tags[name]++
		if typeTagFingerprint(tags) == original {
			t.Errorf("unchanged fingerprint for %s", name)
		}
	}
}
