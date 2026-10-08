package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestSymbolAncestryAndAliasModes(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	platform := filepath.Join(directory, "platform.d.ts")
	config := filepath.Join(directory, "tsconfig.json")
	for path, text := range map[string]string{
		config:   `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["input.a","platform.d.ts"]}`,
		platform: "export {}; declare global { function clock():void; }\n",
		file:     "namespace Clocks {export function tick() {}} import alias=Clocks.tick; clock(); alias;\n",
	} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	var clock, alias *ast.Node
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier {
			if n.Text() == "clock" {
				clock = n
			}
			if n.Text() == "alias" {
				alias = n
			}
		}
		n.ForEachChild(visit)
		return false
	}
	visit(p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file)).AsNode())
	if clock == nil || alias == nil {
		t.Fatal("missing probe names")
	}
	ask := func(node *ast.Node, question string) []string {
		wire, e := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", question)
		if e != nil {
			t.Fatal(e)
		}
		return decodedFields(t, wire)
	}
	fields := ask(clock, "symbol-ancestry")
	if !slices.Contains(fields, "GlobalKeyword") || !slices.Contains(fields, "ModuleBlock") || !slices.Contains(fields, platform) {
		t.Fatalf("lost global declaration ancestry: %q", fields)
	}
	own := ask(alias, "symbol-ancestry\nown")
	resolved := ask(alias, "symbol-ancestry")
	if !slices.Contains(own, "ImportEqualsDeclaration") || !slices.Contains(resolved, "FunctionDeclaration") || slices.Contains(resolved, "ImportEqualsDeclaration") {
		t.Fatalf("alias modes differ from binder: own %q resolved %q", own, resolved)
	}
	if _, err := p.Inspect(file, uint64(clock.Pos()), uint64(clock.End()), "Identifier", "symbol-ancestry\nextra"); err == nil || !strings.Contains(err.Error(), "unexpected symbol-ancestry suffix") {
		t.Fatalf("accepted suffix: %v", err)
	}
}
