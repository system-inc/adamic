package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJsxDeclarationSyntax(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := `const F = React.Fragment; F;
const {Fragment:G} = React; G;
import {Fragment as H} from 'react'; H;
missing;
`
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["prelude.d.ts"]}`, filepath.Join(directory, "prelude.d.ts"): "declare const React:any;"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string][]string{
		"F": {"VariableDeclaration", "VariableDeclarationList", "VariableStatement", "", "", " React.Fragment"},
		"G": {"BindingElement", "ObjectBindingPattern", "VariableDeclaration", "", "", " React"},
		"H": {"ImportSpecifier", "NamedImports", "ImportClause", "Fragment", "react", ""},
	}
	seen := map[string]bool{}
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier && (n.Text() == "F" || n.Text() == "G" || n.Text() == "H" || n.Text() == "missing") {
			wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "jsx-declaration-syntax")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			if len(fields) < 3 || fields[0] != "1" || fields[1] != "jsx-declaration-syntax" {
				t.Fatal("bad header", fields)
			}
			if n.Text() == "missing" {
				if fields[2] != "0" || len(fields) != 3 {
					t.Fatal("unresolved symbol facts", fields)
				}
			} else {
				want := expected[n.Text()]
				if fields[2] != "1" || strings.Join(fields[3:], "\n") != strings.Join(want, "\n") {
					t.Fatalf("%s syntax: %q", n.Text(), fields)
				}
				seen[n.Text()] = true
			}
			if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "jsx-declaration-syntax\nextra"); err == nil {
				t.Fatal("accepted malformed JSX syntax question")
			}
		}
		n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	visit(sf.AsNode())
	if len(seen) != 3 {
		t.Fatal("missing declaration shapes", seen)
	}
	if _, err := p.Inspect(file, uint64(sf.Pos()), uint64(sf.End()), "SourceFile", "jsx-declaration-syntax"); err == nil {
		t.Fatal("accepted non-identifier")
	}
}
