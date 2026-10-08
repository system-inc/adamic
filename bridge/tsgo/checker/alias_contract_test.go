package checker

import (
	"context"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestAliasFactsMatchChecker(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "file.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := "import { value as imported } from './module';\nimported;\n"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler"}}`, filepath.Join(directory, "module.ts"): "export const value = 1;\n"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(file)))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var selected *ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier && n.Text() == "imported" {
			selected = n
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if selected == nil {
		t.Fatal("no imported reference")
	}
	symbol := c.GetSymbolAtLocation(selected)
	target := c.GetAliasedSymbol(symbol)
	if symbol == nil || target == nil || symbol == target || checker.SkipAlias(symbol, c) != target {
		t.Fatal("fixture must be an alias")
	}
	ask := func(question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(selected.Pos()), uint64(selected.End()), "Identifier", question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	ownID, targetID := strconv.FormatUint(p.symbolID(symbol), 10), strconv.FormatUint(p.symbolID(target), 10)
	for _, question := range []string{"alias-target-identity", "skip-alias-identity"} {
		expected := []string{"1", question, "1", targetID, strconv.FormatUint(uint64(target.Flags), 10)}
		if got := ask(question); fmt.Sprint(got) != fmt.Sprint(expected) {
			t.Fatalf("narrow alias metadata %s: %q want %q", question, got, expected)
		}
	}
	expectedLocal := []string{"1", "local-alias-declarations", "1", strconv.FormatUint(uint64(target.Flags), 10), "0"}
	if got := ask("local-alias-declarations"); fmt.Sprint(got) != fmt.Sprint(expectedLocal) {
		t.Fatalf("local alias leaked foreign records: %q", got)
	}
	target = checker.SkipAlias(symbol, c)
	want := []string{"1", "alias-declarations", "1", strconv.FormatUint(uint64(target.Flags), 10), strconv.Itoa(len(target.Declarations))}
	for _, declaration := range target.Declarations {
		want = append(want, ast.GetSourceFileOfNode(declaration).FileName().AsString(), strings.TrimPrefix(declaration.Kind.String(), "Kind"), strconv.Itoa(declaration.Pos()), strconv.Itoa(declaration.End()))
	}
	if got := ask("alias-declarations"); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("SkipAlias declarations: %q want %q", got, want)
	}
	resolved := ask("output-symbol")
	own := ask("output-symbol\nown")
	if resolved[3] != targetID || own[3] != ownID || resolved[7] != "VariableDeclaration" || own[7] != "ImportSpecifier" {
		t.Fatalf("output-symbol lost alias/own distinction: %q %q", resolved, own)
	}
}
