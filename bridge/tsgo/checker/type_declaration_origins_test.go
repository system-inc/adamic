package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

func TestTypeDeclarationOrigins(t *testing.T) {
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input.ts")
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"lib":["ES2022"]},"files":["input.ts","ambient.d.ts"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "ambient.d.ts"), []byte(`declare module "witness-package" { export interface Ambient { value: string } }`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("import type { Ambient } from 'witness-package';\nclass Local {}\ndeclare const local: Local;\ndeclare const ambient: Ambient;\ndeclare const error: Error;\nlocal; ambient; error;"), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file, filepath.Join(directory, "ambient.d.ts")})
	if err != nil {
		t.Fatal(err)
	}
	source := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	found := map[string]*ast.Node{}
	var walk func(*ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			found[node.Text()] = node
		}
		node.ForEachChild(walk)
		return false
	}
	walk(source.AsNode())
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		value, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, value)
	}
	for _, name := range []string{"local", "ambient", "error"} {
		node := found[name]
		if node == nil {
			t.Fatal("fixture node absent", name)
		}
		shape := ask(node, "raw-shape")
		origins := ask(node, "type-declaration-origins\n"+shape[5])
		if len(origins) < 7 || origins[2] != directory || origins[3] == "0" {
			t.Fatalf("%s origins: %q", name, origins)
		}
		switch name {
		case "local":
			if origins[4] != file || origins[5] != "0" || origins[6] != "" {
				t.Fatalf("local origins: %q", origins)
			}
		case "ambient":
			if origins[4] != filepath.Join(directory, "ambient.d.ts") || origins[5] != "0" || origins[6] != "witness-package" {
				t.Fatalf("ambient origins: %q", origins)
			}
		case "error":
			if origins[5] != "1" || !strings.HasSuffix(origins[4], "lib.es5.d.ts") {
				t.Fatalf("library origins: %q", origins)
			}
		}
	}

	foreign := filepath.Join(directory, "ambient.d.ts")
	foreignSource := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(foreign))
	direct, err := program.Inspect(foreign, uint64(foreignSource.Pos()), uint64(foreignSource.End()), "SourceFile", "scope-locals")
	if err != nil {
		t.Fatal(err)
	}
	delegated, err := program.Inspect(file, uint64(source.Pos()), uint64(source.End()), "SourceFile", fmt.Sprintf("other-file\n%s\n%d\n%d\nSourceFile\nscope-locals", foreign, foreignSource.Pos(), foreignSource.End()))
	if err != nil || direct != delegated {
		t.Fatalf("foreign facts: error=%v delegated=%q direct=%q", err, delegated, direct)
	}
	for _, question := range []string{"type-declaration-origins", "type-declaration-origins\n0", "type-declaration-origins\n01", "type-declaration-origins\n1\nextra"} {
		node := found["local"]
		if _, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", question); err == nil {
			t.Fatalf("accepted %q", question)
		}
	}
}
