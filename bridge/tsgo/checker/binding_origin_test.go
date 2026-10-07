package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestBindingOriginFacts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	if err := os.WriteFile(file, []byte("function f(){ arguments; arguments.length; const record={arguments}; } const R=RegExp; new R('世界'); export { R }; undeclared;"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"sourceExtensions":[".a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "fixture.d.ts"), []byte("export {};\n"), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := program.Compiler.GetSourceFile(file)
	c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	probes, implicit, shorthand, ambient, missing := 0, 0, 0, 0, 0
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			wire, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "binding-origin")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			at := 2
			original := c.GetSymbolAtLocation(node)
			read := original
			if parent := node.Parent; parent != nil {
				if parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
					read = c.GetShorthandAssignmentValueSymbol(parent)
				}
				if parent.Kind == ast.KindExportSpecifier {
					read = c.GetExportSpecifierLocalTargetSymbol(parent)
				}
			}
			for _, symbol := range []*ast.Symbol{original, read} {
				id, err := strconv.ParseUint(fields[at], 10, 64)
				if err != nil {
					t.Fatal(err)
				}
				at++
				if symbol == nil {
					if id != 0 || fields[at] != "0" {
						t.Fatal("missing symbol", fields)
					}
					missing++
					at++
					continue
				}
				if id == 0 || fields[at] != strconv.Itoa(len(symbol.Declarations)) {
					t.Fatal("symbol declaration count", fields)
				}
				at++
				if node.Text() == "arguments" && len(symbol.Declarations) == 0 {
					implicit++
				}
				for _, decl := range symbol.Declarations {
					f := ast.GetSourceFileOfNode(decl)
					if fields[at] != f.FileName() || fields[at+1] != strings.TrimPrefix(decl.Kind.String(), "Kind") || fields[at+2] != strconv.Itoa(decl.Pos()) || fields[at+3] != strconv.Itoa(decl.End()) {
						t.Fatal("declaration mismatch", fields)
					}
					want := "0"
					if f.IsDeclarationFile {
						want = "1"
						ambient++
					}
					if fields[at+4] != want {
						t.Fatal("ambient mismatch", fields)
					}
					at += 5
				}
			}
			if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment && read != original {
				shorthand++
			}
			if at != len(fields) {
				t.Fatal("trailing fields", fields)
			}
			probes++
			if _, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "binding-origin\nextra"); err == nil {
				t.Fatal("suffix accepted")
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if probes < 10 || implicit == 0 || shorthand == 0 || ambient == 0 || missing == 0 {
		t.Fatal("vacuous binding probes", probes, implicit, shorthand, ambient, missing)
	}
	if _, err := program.Inspect(file, uint64(source.Pos()), uint64(source.End()), "SourceFile", "binding-origin"); err == nil {
		t.Fatal("wrong kind accepted")
	}
}
