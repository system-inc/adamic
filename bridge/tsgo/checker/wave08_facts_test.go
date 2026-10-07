package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

func TestWave08Facts(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	other := filepath.Join(directory, "other.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := "import {value} from './other.ts';declare function require(s:string):unknown;require('世界🌍');for(var i=0;i<2;i++){(()=>i)()}value;"
	for path, text := range map[string]string{file: source, other: "export const value=1;", filepath.Join(directory, "ambient.d.ts"): "", config: `{"compilerOptions":{"strict":true,"target":"ES2022","allowImportingTsExtensions":true,"moduleResolution":"bundler","module":"ESNext","noEmit":true}}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file, other})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var nodes []*ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		nodes = append(nodes, node)
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	ask := func(node *ast.Node, q string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), q)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	checked := 0
	for _, node := range nodes {
		if node.Pos() >= node.End() {
			continue
		}
		syntax := ask(node, "syntax-metadata")
		if syntax[2] != strconv.FormatUint(uint64(node.Flags), 10) || syntax[3] != strconv.FormatUint(uint64(node.ModifierFlags()), 10) || syntax[4] != strconv.Itoa(boolInt(ast.IsTypeNode(node))) {
			t.Fatalf("syntax differs at %s: %q", node.Kind, syntax)
		}
		if node.Kind == ast.KindForStatement {
			initializer := node.AsForStatement().Initializer
			if syntax[5] != "1" || syntax[6] != strconv.Itoa(initializer.Pos()) || syntax[7] != strconv.Itoa(initializer.End()) {
				t.Fatalf("initializer: %q", syntax)
			}
		}
		if node.Kind != ast.KindIdentifier {
			continue
		}
		symbol := c.GetSymbolAtLocation(node)
		facts := ask(node, "wave08-symbol")
		id, err := strconv.ParseUint(facts[2], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if symbol == nil {
			if id != 0 {
				t.Fatal("unresolved symbol has identity")
			}
			continue
		}
		if id == 0 || p.symbolsByID[id-1] != symbol || facts[4] != strconv.FormatUint(uint64(symbol.Flags), 10) {
			t.Fatalf("symbol identity/flags: %q", facts)
		}
		count, err := strconv.Atoi(facts[5])
		if err != nil || count != len(symbol.Declarations) {
			t.Fatalf("declaration count: %q", facts)
		}
		at := 6
		for _, declaration := range symbol.Declarations {
			if facts[at] != ast.GetSourceFileOfNode(declaration).FileName() || facts[at+1] != strings.TrimPrefix(declaration.Kind.String(), "Kind") || facts[at+2] != strconv.Itoa(declaration.Pos()) || facts[at+3] != strconv.Itoa(declaration.End()) || facts[at+4] != strconv.FormatUint(uint64(declaration.Flags), 10) {
				t.Fatalf("declaration differs: %q", facts)
			}
			at += 5
		}
		if node.Text() == "value" && scanner.GetTokenPosOfNode(node, sf, false) > 100 {
			target := c.GetAliasedSymbol(symbol)
			if facts[at+1] != target.Name || facts[at+2] != strconv.FormatUint(uint64(target.Flags), 10) || facts[at+5] != other || facts[at+6] != "VariableDeclaration" {
				t.Fatalf("alias target differs: %q", facts)
			}
			checked++
		}
	}
	if checked != 1 {
		t.Fatalf("alias positive control count %d", checked)
	}
	links := ask(sf.AsNode(), "module-links")
	count, err := strconv.Atoi(links[2])
	if err != nil {
		t.Fatal(err)
	}
	at := 3
	found := false
	for i := 0; i < count; i++ {
		path := links[at]
		at += 2
		n, _ := strconv.Atoi(links[at])
		at++
		for j := 0; j < n; j++ {
			if path == file && links[at+2] == other {
				found = true
			}
			at += 3
		}
	}
	if !found {
		t.Fatalf("resolved module absent: %q", links)
	}
	for _, question := range []string{"wave08-symbol\nbad", "syntax-metadata\nbad", "module-links\nbad"} {
		if _, err := p.Inspect(file, 0, uint64(sf.End()), "SourceFile", question); err == nil {
			t.Fatalf("suffix accepted: %s", question)
		}
	}
	if _, err := p.Inspect(file, 0, 6, "Identifier", "module-links"); err == nil {
		t.Fatal("module-links accepted a value node")
	}
	t.Logf("%d syntax nodes, alias identity and declarations, resolved module, suffix and kind refusals", len(nodes))
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
