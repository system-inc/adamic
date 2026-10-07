package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSymbolsInScopeFacts(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.a")
	config := filepath.Join(dir, "tsconfig.json")
	for name, text := range map[string]string{"tsconfig.json": `{"compilerOptions":{"strict":true},"files":["prelude.d.ts"]}`, "prelude.d.ts": "", "input.a": "const outside=1;label:{let inside=1;}export {};"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(file)
	var label *ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindLabeledStatement {
			label = n
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	wire, err := p.Inspect(file, uint64(label.Pos()), uint64(label.End()), "LabeledStatement", "symbols-in-scope")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	symbols := c.GetSymbolsInScope(label, ast.SymbolFlagsValue)
	if fields[2] != strconv.Itoa(len(symbols)) {
		t.Fatal("scope count differs")
	}
	want := make(map[string]string)
	for _, symbol := range symbols {
		want[symbol.Name] = strconv.FormatUint(uint64(symbol.Flags), 10)
	}
	for at := range symbols {
		if want[fields[3+at*2]] != fields[4+at*2] {
			t.Fatal("scope symbol differs")
		}
	}
	for _, q := range []string{"symbols-in-scope\nextra", "symbols-in-scope"} {
		node := label
		kind := "LabeledStatement"
		if q == "symbols-in-scope" {
			node = source.AsNode()
			kind = "SourceFile"
		}
		if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), kind, q); err == nil {
			t.Fatalf("accepted %s %s", kind, q)
		}
	}
}
