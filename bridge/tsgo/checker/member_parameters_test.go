package checker

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

func TestMemberParameterSourceText(t *testing.T) {
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input.a")
	source := `interface Methods { indexOf(世界: unknown, from?: number): number; includes(世界: unknown, from?: number): boolean; }
declare const methods: Methods; methods.indexOf('x');`
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["prelude.d.ts"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "prelude.d.ts"), []byte("declare const marker: number;"), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	var selected *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindPropertyAccessExpression {
			selected = node.AsPropertyAccessExpression().Name()
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(file))).AsNode())
	wire, err := program.Inspect(file, uint64(selected.Pos()), uint64(selected.End()), "Identifier", "member-parameters")
	if err != nil {
		t.Fatal(err)
	}
	got := decodedFields(t, wire)
	want := []string{"1", "member-parameters", "1", "1", "2", "世界: unknown", "from?: number", "1", "1", "2", "世界: unknown", "from?: number"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parameter source text: %q != %q", got, want)
	}
	if _, err := program.Inspect(file, uint64(selected.Pos()), uint64(selected.End()), "Identifier", "member-parameters\nextra"); err == nil || !strings.Contains(err.Error(), "suffix") {
		t.Fatalf("unexpected suffix accepted: %v", err)
	}
}
