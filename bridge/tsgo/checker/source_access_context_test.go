package checker

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

func TestWave03SourceAccessContext(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.tsx")
	config := filepath.Join(directory, "tsconfig.json")
	source := `let value=1; value=2; use(value); type T=typeof value; const view=<div title="x"/>; export {value};`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","jsx":"preserve"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	yes := func(value bool) string {
		if value {
			return "1"
		}
		return "0"
	}
	positives := [4]int{}
	negatives := [4]int{}
	queries := 0
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Pos() < 0 || node.End() <= node.Pos() {
			return
		}
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), "source-access-context")
		if err != nil {
			t.Fatal(err)
		}
		got := decodedFields(t, wire)
		want := []string{yes(ast.IsDeclarationName(node)), yes(ast.IsWriteAccess(node)), yes(ast.IsPartOfTypeNode(node)), yes(node.Kind == ast.KindIdentifier && scanner.IsIntrinsicJsxName(node.Text()))}
		if len(got) != 6 || strings.Join(got[2:], "|") != strings.Join(want, "|") {
			t.Fatalf("%s access fields %q want %q", node.Kind, got, want)
		}
		for i, value := range want {
			if value == "1" {
				positives[i]++
			} else {
				negatives[i]++
			}
		}
		queries++
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	for i := range positives {
		if positives[i] == 0 || negatives[i] == 0 {
			t.Fatalf("field %s lacks both controls", strconv.Itoa(i))
		}
	}
	if _, err := p.Inspect(file, 0, uint64(sf.End()), "SourceFile", "source-access-context\nextra"); err == nil {
		t.Fatal("accepted malformed access question")
	}
	t.Logf("%d exact access queries; positive fields %v negative fields %v", queries, positives, negatives)
}
