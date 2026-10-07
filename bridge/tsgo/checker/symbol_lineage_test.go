package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSymbolLineage(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	if err := os.WriteFile(config, []byte(`{"files":["input.a"],"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("declare const callback:((value:number)=>void)|undefined;callback;export {};\n"), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	var node *ast.Node
	var walk func(*ast.Node)
	walk = func(current *ast.Node) {
		if current.Kind == ast.KindIdentifier && current.Text() == "callback" {
			node = current
		}
		current.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if node == nil {
		t.Fatal("missing callback")
	}
	wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "symbol-lineage")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if fields[0] != "1" || fields[1] != "symbol-lineage" || fields[2] != "1" || fields[4] != "1" || fields[5] != file || fields[6] != "VariableDeclaration" {
		t.Fatalf("lost declaration: %q", fields)
	}
	identity, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil || identity == 0 || p.symbolsByID[identity-1] != c.GetSymbolAtLocation(node) {
		t.Fatal("wrong symbol identity")
	}
	want := len(c.GetSignaturesOfType(c.GetNonNullableType(c.GetTypeAtLocation(node)), checker.SignatureKindCall))
	if want != 1 || fields[len(fields)-1] != strconv.Itoa(want) {
		t.Fatalf("nonnullable signatures differ: %q", fields)
	}
	if fields[19] != "VariableDeclarationList" || fields[20] != strconv.FormatUint(uint64(c.GetSymbolAtLocation(node).Declarations[0].Parent.Flags), 10) || fields[23] != "VariableStatement" {
		t.Fatalf("lost ancestry: %q", fields)
	}
	if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "symbol-lineage\nextra"); err == nil {
		t.Fatal("symbol-lineage accepted a suffix")
	}
}
