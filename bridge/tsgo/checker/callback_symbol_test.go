package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCallbackSymbolFacts(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.a")
	config := filepath.Join(dir, "tsconfig.json")
	for name, text := range map[string]string{"tsconfig.json": `{"compilerOptions":{"strict":true},"files":["input.a"]}`, "input.a": `declare const key:unique symbol;const object={[key](){arguments;}};function f(){return f;}`} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	ids := map[*ast.Symbol]string{}
	seen := 0
	var identifier *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			identifier = node
			wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "callback-symbol")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			symbol := c.GetSymbolAtLocation(node)
			count := 0
			if symbol != nil {
				count = len(symbol.Declarations)
			}
			if len(fields) != 4 || fields[2] == "0" && symbol != nil || fields[2] != "0" && symbol == nil || fields[3] != strconv.Itoa(count) {
				t.Fatalf("wrong raw symbol facts: %v", fields)
			}
			if old, ok := ids[symbol]; ok && old != fields[2] {
				t.Fatal("identity changed")
			}
			ids[symbol] = fields[2]
			seen++
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if seen < 5 {
		t.Fatal("missing identifiers")
	}
	if _, err := p.Inspect(file, uint64(identifier.Pos()), uint64(identifier.End()), "Identifier", "callback-symbol\nextra"); err == nil {
		t.Fatal("accepted suffix on identifier")
	}
	for _, question := range []string{"callback-symbol", "callback-symbol\nextra"} {
		if _, err := p.Inspect(file, 0, uint64(source.End()), "SourceFile", question); err == nil {
			t.Fatal("accepted wrong kind or suffix")
		}
	}
}
