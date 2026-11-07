package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The removed function body was the only use of errors, like Lower's async pass.
func TestImportCleanDropsOnlyTheImportsARewriteLeftUnused(test *testing.T) {
	test.Parallel()
	source := []byte(`package fixture
import ("errors"; f "fmt"; _ "net/http")
func lower() error { return errors.New(f.Sprint("production")) }
`)
	fs := token.NewFileSet()
	file, parseFileError := parser.ParseFile(fs, "fixture.go", source, parser.ParseComments)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	// Model the no-output rewrite: the production body is replaced by a measurement error.
	replacement, parseFileError := parser.ParseFile(fs, "body.go", `package fixture
 func lower() error { return f.Errorf("measurement only") }`, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	file.Decls[1].(*ast.FuncDecl).Body = replacement.Decls[0].(*ast.FuncDecl).Body
	var rewritten bytes.Buffer
	if nodeError := format.Node(&rewritten, fs, file); nodeError != nil {
		test.Fatal(nodeError)
	}
	// The check can fail: before cleaning, the rewritten file leaves errors unused.
	if unused := unusedImports(test, rewritten.Bytes()); len(unused) != 1 || !strings.Contains(unused[0], `"errors"`) {
		test.Fatalf("the rewritten fixture's unused imports are %q, want errors alone", unused)
	}
	output, cleanError := clean(rewritten.Bytes())
	if cleanError != nil {
		test.Fatal(cleanError)
	}
	if !strings.Contains(string(output), `f "fmt"`) || !strings.Contains(string(output), `_ "net/http"`) {
		test.Fatal("used or side-effect import lost")
	}
	if unused := unusedImports(test, output); len(unused) != 0 {
		test.Fatalf("the cleaned fixture still leaves imports unused: %q\n%s", unused, output)
	}
}

// Main's own lowering files, the stable input under every overlay rewrite, use each import they declare, so
// importclean must take each one without an error and give it back unchanged, blank imports like _ "unsafe" for
// go:linkname included. An import form it can't read or misjudges reds here, at its author's change, not at the meter.
func TestImportCleanKeepsMainLowerImports(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/*.go
	paths, globError := filepath.Glob("../../../../internal/lower/*.go")
	if globError != nil {
		test.Fatal(globError)
	}
	cleaned := 0
	for _, path := range paths {
		// make_overlay.py hands importclean only the package's own files; tests are never overlaid.
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		source, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		output, cleanError := clean(source)
		if cleanError != nil {
			test.Errorf("%s doesn't clean: %v", filepath.Base(path), cleanError)
			continue
		}
		file, parseFileError := parser.ParseFile(token.NewFileSet(), filepath.Base(path), output, parser.ImportsOnly)
		if parseFileError != nil {
			test.Errorf("%s cleans to source that doesn't parse: %v", filepath.Base(path), parseFileError)
			continue
		}
		original, parseFileError := parser.ParseFile(
			token.NewFileSet(),
			filepath.Base(path),
			source,
			parser.ImportsOnly,
		)
		if parseFileError != nil {
			test.Fatal(parseFileError)
		}
		if imports, want := importSpecs(file), importSpecs(original); !slices.Equal(imports, want) {
			test.Errorf("%s's imports after cleaning are %q, want %q", filepath.Base(path), imports, want)
		} else if !bytes.Equal(output, source) {
			test.Errorf("%s comes back changed though every import is used", filepath.Base(path))
		}
		cleaned++
	}
	if cleaned == 0 {
		test.Fatal("no internal/lower file was read")
	}
}

// importSpecs reads each import as its name, if any, and quoted path, in source order.
func importSpecs(file *ast.File) []string {
	var specs []string
	for _, spec := range file.Imports {
		if spec.Name != nil {
			specs = append(specs, spec.Name.Name+" "+spec.Path.Value)
			continue
		}
		specs = append(specs, spec.Path.Value)
	}
	return specs
}

// unusedImports type-checks source the way the compiler judges imports, without running go: each import resolves to
// an empty package, so a use reads as undefined, which is ignored, and only "imported and not used" is collected.
func unusedImports(test *testing.T, source []byte) []string {
	test.Helper()
	fs := token.NewFileSet()
	file, parseFileError := parser.ParseFile(fs, "fixture.go", source, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	var unused []string
	config := types.Config{
		Importer: stubImporter{},
		Error: func(typeCheckError error) {
			if strings.Contains(typeCheckError.Error(), "imported and not used") {
				unused = append(unused, typeCheckError.Error())
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
