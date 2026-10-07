package checker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestAccessedPropertyFacts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "file.a")
	source := "enum Key { Catch='catch' };declare const p:Promise<number>;p[Key.Catch];p.catch;declare const o:{'世界🌍':number};o['世界🌍'];\n"
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"lib":["ES2022"]},"sourceExtensions":[".a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "fixture.d.ts"), []byte("export {};\n"), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	f := program.Compiler.GetSourceFile(file)
	c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), f)
	defer release()
	count := 0
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if ast.IsAccessExpression(node) {
			want, _ := checker.Checker_getAccessedPropertyName(c, node)
			wire, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), "accessed-property")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			if len(fields) != 3 || fields[2] != want {
				t.Fatalf("wrong property fact %q want %q", fields, want)
			}
			if _, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), "accessed-property\nextra"); err == nil {
				t.Fatal("suffix accepted")
			}
			count++
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(f.AsNode())
	if count != 4 {
		t.Fatalf("missing fact probes: %d", count)
	}
	if _, err := program.Inspect(file, uint64(f.Pos()), uint64(f.End()), "SourceFile", "accessed-property"); err == nil {
		t.Fatal("nonaccess accepted")
	}
}
