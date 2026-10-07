package checker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestDeclarationFileFlags(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	if err := os.WriteFile(filepath.Join(directory, "root.d.ts"), []byte("export {};"), 0600); err != nil {
		t.Fatal(err)
	}
	source := "parseInt;undefined;missing;function f(parseInt:number){parseInt;} export {};"
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	counts := map[string]int{}
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			symbol := c.GetSymbolAtLocation(node)
			wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "declaration-file-flags")
			if err != nil {
				t.Fatal(err)
			}
			got := decodedFields(t, wire)
			var expected fields
			expected.number(1)
			expected.text("declaration-file-flags")
			expected.yes(symbol != nil)
			if symbol != nil {
				expected.number(uint64(len(symbol.Declarations)))
				for _, decl := range symbol.Declarations {
					f := ast.GetSourceFileOfNode(decl)
					expected.yes(f != nil)
					expected.yes(f != nil && f.IsDeclarationFile)
				}
			}
			if wire != expected.String() {
				t.Fatalf("%s: %q != direct checker", node.Text(), got)
			}
			counts[node.Text()]++
			if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "declaration-file-flags\nextra"); err == nil {
				t.Fatal("suffix accepted")
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	for _, name := range []string{"parseInt", "undefined", "missing"} {
		if counts[name] == 0 {
			t.Fatal("missing control " + name)
		}
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "declaration-file-flags"); err == nil || !strings.Contains(err.Error(), "requires an Identifier") {
		t.Fatalf("wrong-kind refusal: %v", err)
	}
}
