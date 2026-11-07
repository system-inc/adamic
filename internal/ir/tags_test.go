package ir

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestTypeTagsComplete(test *testing.T) {
	test.Parallel()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		test.Fatal(err)
	}
	var declarations []ast.Decl
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			test.Fatal(err)
		}
		declarations = append(declarations, file.Decls...)
	}
	tags := TypeTags()
	found := map[string]bool{}
	for _, declaration := range declarations {
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
						test.Errorf("missing fingerprint tag %s", name.Name)
					}
				}
			}
		}
	}
	for name := range tags {
		if !found[name] {
			test.Errorf("unknown fingerprint tag %s", name)
		}
	}
}

func TestTypeTagFingerprintChanges(test *testing.T) {
	test.Parallel()
	original := TypeTagFingerprint()
	for name := range TypeTags() {
		tags := TypeTags()
		tags[name]++
		if typeTagFingerprint(tags) == original {
			test.Errorf("unchanged fingerprint for %s", name)
		}
	}
}
