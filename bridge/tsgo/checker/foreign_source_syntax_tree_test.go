package checker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestWave03ForeignSyntaxAndSymbol(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := "const value:ConstrainBoolean = true;"
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"target":"ES2022","lib":["ES2022","DOM"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	anchor := program.Compiler.GetSourceFile(file)
	var library *ast.SourceFile
	for _, sf := range program.Compiler.GetSourceFiles() {
		if strings.HasSuffix(sf.FileName(), "lib.dom.d.ts") {
			library = sf
			break
		}
	}
	if library == nil {
		t.Fatal("DOM library missing")
	}
	query := "foreign-source-syntax-tree\n" + library.FileName()
	got, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", query)
	if err != nil {
		t.Fatal(err)
	}
	var expected fields
	expected.number(1)
	expected.text("foreign-source-syntax-tree")
	want, err := program.sourceSyntaxTree(&expected, library.AsNode(), "source-syntax-tree")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatal("foreign AST differs from loaded compiler AST")
	}
	var identifier *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if identifier != nil {
			return
		}
		if node.Kind == ast.KindIdentifier && node.Text() == "ConstrainBooleanParameters" {
			identifier = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return identifier != nil })
	}
	walk(library.AsNode())
	if identifier == nil {
		t.Fatal("library alias reference missing")
	}
	query = fmt.Sprintf("foreign-node-symbol-context\n%s\n%d\n%d\nIdentifier", library.FileName(), identifier.Pos(), identifier.End())
	got, err = program.Inspect(file, 0, uint64(len(source)), "SourceFile", query)
	if err != nil {
		t.Fatal(err)
	}
	checker, release := program.Compiler.GetTypeCheckerForFile(context.Background(), anchor)
	defer release()
	var symbol fields
	symbol.number(1)
	symbol.text("foreign-node-symbol-context")
	want, err = program.symbolContext(&symbol, checker, identifier, "symbol-context")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatal("foreign symbol differs from checker")
	}
	for _, question := range []string{"foreign-source-syntax-tree", "foreign-source-syntax-tree\nmissing", "foreign-source-syntax-tree\n" + library.FileName() + "\nextra", "foreign-node-symbol-context", "foreign-node-symbol-context\n" + library.FileName() + "\n01\n2\nIdentifier", "foreign-node-symbol-context\n" + library.FileName() + "\n1\n2\nIdentifier", "foreign-node-symbol-context\nmissing\n0\n1\nIdentifier", fmt.Sprintf("foreign-node-symbol-context\n%s\n0%d\n%d\nIdentifier", library.FileName(), identifier.Pos(), identifier.End())} {
		if _, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", question); err == nil {
			t.Fatalf("accepted %q", question)
		}
	}
	t.Log("bundled AST and symbol fields match their compiler sources; malformed foreign requests refused")
}
