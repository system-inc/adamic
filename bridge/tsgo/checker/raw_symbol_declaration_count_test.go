package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRawSymbolDeclarationCount(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	for path, text := range map[string]string{filepath.Join(directory, "globals.d.ts"): "declare const positiveControl: number;\n", config: `{"compilerOptions":{"strict":true,"target":"ES2022"}}`, file: "arguments;function f(){arguments;const value={arguments};} export {};\n"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	var identifiers []*ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier && node.Text() == "arguments" {
			identifiers = append(identifiers, node)
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(program.Compiler.GetSourceFile(tspath.RootedFilePath(file)).AsNode())
	if len(identifiers) != 3 {
		t.Fatal("missing direct/shorthand controls", len(identifiers))
	}
	for at, node := range identifiers {
		wire, err := program.InspectWave24(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "raw-symbol-declaration-count")
		if err != nil {
			t.Fatal(err)
		}
		fields := decodedFields(t, wire)
		if len(fields) != 4 {
			t.Fatal(fields)
		}
		if at == 0 && fields[2] != "0" {
			t.Fatal("unresolved top-level arguments", fields)
		}
		if at == 1 && (fields[2] == "0" || fields[3] != "0") {
			t.Fatal("implicit arguments", fields)
		}
		if at == 2 && (fields[2] == "0" || fields[3] != "1") {
			t.Fatal("direct shorthand property", fields)
		}
	}
	node := identifiers[1]
	if _, err := program.InspectWave24(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "raw-symbol-declaration-count\nwrong"); err == nil {
		t.Fatal("accepted suffix")
	}
	root := program.Compiler.GetSourceFile(tspath.RootedFilePath(file)).AsNode()
	if _, err := program.InspectWave24(file, uint64(root.Pos()), uint64(root.End()), strings.TrimPrefix(root.Kind.String(), "Kind"), "raw-symbol-declaration-count"); err == nil {
		t.Fatal("accepted wrong kind")
	}
}
