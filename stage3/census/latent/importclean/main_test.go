package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

// The removed function body was the only use of errors, like Lower's async pass.
func TestImportCleanDropsOnlyTheImportsARewriteLeftUnused(t *testing.T) {
	t.Parallel()
	source := []byte(`package fixture
import ("errors"; f "fmt"; _ "net/http")
func lower() error { return errors.New(f.Sprint("production")) }
`)
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "fixture.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	// Model the no-output rewrite: the production body is replaced by a measurement error.
	replacement, err := parser.ParseFile(fs, "body.go", `package fixture
 func lower() error { return f.Errorf("measurement only") }`, 0)
	if err != nil {
		t.Fatal(err)
	}
	file.Decls[1].(*ast.FuncDecl).Body = replacement.Decls[0].(*ast.FuncDecl).Body
	var rewritten bytes.Buffer
	if err := format.Node(&rewritten, fs, file); err != nil {
		t.Fatal(err)
	}
	// The check can fail: before cleaning, the rewritten file leaves errors unused.
	if unused := unusedImports(t, rewritten.Bytes()); len(unused) != 1 || !strings.Contains(unused[0], `"errors"`) {
		t.Fatalf("the rewritten fixture's unused imports are %q, want errors alone", unused)
	}
	output, err := clean(rewritten.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), `f "fmt"`) || !strings.Contains(string(output), `_ "net/http"`) {
		t.Fatal("used or side-effect import lost")
	}
	if unused := unusedImports(t, output); len(unused) != 0 {
		t.Fatalf("the cleaned fixture still leaves imports unused: %q\n%s", unused, output)
	}
}

// unusedImports type-checks source the way the compiler judges imports, without running go: each import resolves to
// an empty package, so a use reads as undefined, which is ignored, and only "imported and not used" is collected.
func unusedImports(t *testing.T, source []byte) []string {
	t.Helper()
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "fixture.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	var unused []string
	config := types.Config{
		Importer: stubImporter{},
		Error: func(err error) {
			if strings.Contains(err.Error(), "imported and not used") {
				unused = append(unused, err.Error())
			}
		},
	}
	_, _ = config.Check("fixture", fs, []*ast.File{file}, nil)
	return unused
}

type stubImporter struct{}

func (stubImporter) Import(path string) (*types.Package, error) {
	name := path[strings.LastIndex(path, "/")+1:]
	pkg := types.NewPackage(path, name)
	pkg.MarkComplete()
	return pkg, nil
}
