package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestExportCheckerQuestions(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	helper := filepath.Join(directory, "helper.ts")
	config := filepath.Join(directory, "tsconfig.json")
	for path, text := range map[string]string{filepath.Join(directory, "prelude.d.ts"): "", config: `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["prelude.d.ts"]}`, file: `export { T, v } from './helper.js'; export * from './helper.js';`, helper: `export interface T {x:number}; export const v=1;`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file, helper})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var name, module *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier && node.Text() == "T" {
			name = node
		}
		if node.Kind == ast.KindStringLiteral {
			module = node
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	ask := func(node *ast.Node, question string) (string, error) {
		return p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
	}
	wire, err := ask(name, "export-symbol-chain")
	if err != nil {
		t.Fatal(err)
	}
	facts := decodedFields(t, wire)
	symbol := c.GetSymbolAtLocation(name)
	alias := checker.SkipAlias(symbol, c)
	if facts[2] != "2" || facts[3] != strconv.FormatUint(uint64(symbol.Flags), 10) {
		t.Fatalf("alias flags differ: %q", facts)
	}
	// The terminal interface flags are held to the checker, separately from the
	// native consumer's type/value interpretation.
	terminal := strconv.FormatUint(uint64(alias.Flags), 10)
	if !strings.Contains(strings.Join(facts, "|"), "|"+terminal+"|1|InterfaceDeclaration|0|") {
		t.Fatalf("terminal flags/declaration differ: %q", facts)
	}
	wire, err = ask(module, "export-module-properties")
	if err != nil {
		t.Fatal(err)
	}
	facts = decodedFields(t, wire)
	typ := c.GetTypeOfSymbol(c.GetSymbolAtLocation(module))
	properties := checker.Checker_getPropertiesOfType(c, typ)
	want := []string{"1", "export-module-properties", "1", "1", strconv.Itoa(len(properties))}
	for _, property := range properties {
		present := "0"
		if checker.Checker_getPropertyOfType(c, typ, property.Name) != nil {
			present = "1"
		}
		want = append(want, property.Name, present)
	}
	if strings.Join(facts, "|") != strings.Join(want, "|") {
		t.Fatalf("module properties %q want %q", facts, want)
	}
	for _, probe := range []struct {
		node     *ast.Node
		question string
	}{{sf.AsNode(), "export-symbol-chain"}, {name, "export-module-properties"}, {name, "export-symbol-chain\n1"}, {module, "export-module-properties\n1"}} {
		if _, err := ask(probe.node, probe.question); err == nil {
			t.Fatalf("accepted invalid %s", probe.question)
		}
	}
}
