package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

func TestSymbolDeclarationProvenance(t *testing.T) {
	dir := t.TempDir()
	config, file, ambient := filepath.Join(dir, "tsconfig.json"), filepath.Join(dir, "input.ts"), filepath.Join(dir, "ambient.d.ts")
	for path, source := range map[string]string{config: `{"compilerOptions":{"strict":true,"lib":["ES2022"]}}`, file: `const own=1;own;ambient;missing;const ref={RegExp};export {};`, ambient: `declare const ambient:string;`} {
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file, ambient})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	selected := map[string]*ast.Node{}
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			selected[node.Text()] = node
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	for _, name := range []string{"own", "ambient", "missing", "RegExp"} {
		node := selected[name]
		if node == nil {
			t.Fatal("missing identifier", name)
		}
		symbol := c.GetSymbolAtLocation(node)
		if name == "RegExp" {
			value := c.GetShorthandAssignmentValueSymbol(node.Parent)
			if value == nil || value == symbol {
				t.Fatal("shorthand value symbol not distinct")
			}
			symbol = value
		}
		id := p.symbolID(symbol)
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "symbol-declaration-provenance\n"+strconv.FormatUint(id, 10))
		if err != nil {
			t.Fatal(err)
		}
		got := decodedFields(t, wire)
		var declarations []*ast.Node
		if symbol != nil {
			declarations = symbol.Declarations
		}
		if len(got) != 3+2*len(declarations) || mustUint(t, got[2]) != uint64(len(declarations)) {
			t.Fatalf("wrong count: %v", got)
		}
		for at, declaration := range declarations {
			source := ast.GetSourceFileOfNode(declaration)
			expected := "0"
			if source.IsDeclarationFile {
				expected = "1"
			}
			if got[3+2*at] != source.FileName().AsString() || got[4+2*at] != expected {
				t.Fatalf("checker provenance differs: %v", got)
			}
		}
		if name == "own" && (len(declarations) != 1 || got[4] != "0") {
			t.Fatal("source binding not exercised")
		}
		if (name == "ambient" || name == "RegExp") && (len(declarations) == 0 || got[4] != "1") {
			t.Fatal("declaration-file binding not exercised")
		}
		if name == "missing" && len(declarations) != 0 {
			t.Fatal("nil symbol not exercised")
		}
	}
	node := selected["own"]
	for _, suffix := range []string{"", "\n01", "\n-1", "\n999999999", "\n0\nextra"} {
		if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "symbol-declaration-provenance"+suffix); err == nil {
			t.Fatalf("accepted malformed identity %q", suffix)
		}
	}
}
