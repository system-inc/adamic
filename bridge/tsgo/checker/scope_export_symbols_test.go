package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestScopeExportSymbols(t *testing.T) {
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input.ts")
	source := `namespace A { export const x=1; export type T=number;const y=A.x;let t:A.T;namespace B {const x=2;const y=A.x;} }
import Alias=A;namespace A {const z=Alias.x;}
export {};`
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022"}}`, file: source} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	checked, shadowed, aliased := 0, false, false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		var qualifier, name *ast.Node
		if node.Kind == ast.KindPropertyAccessExpression {
			qualifier = node.AsPropertyAccessExpression().Expression
			name = node.AsPropertyAccessExpression().Name()
		}
		if node.Kind == ast.KindQualifiedName {
			qualifier = node.AsQualifiedName().Left
			name = node.AsQualifiedName().Right
		}
		if qualifier != nil && name != nil {
			accessed := c.GetSymbolAtLocation(name)
			if accessed == nil {
				t.Fatal("missing positive accessed symbol")
			}
			question := "scope-export-symbols\n" + strconv.FormatUint(uint64(accessed.Flags), 10) + "\n" + name.Text()
			wire, err := p.Inspect(file, uint64(qualifier.Pos()), uint64(qualifier.End()), strings.TrimPrefix(qualifier.Kind.String(), "Kind"), question)
			if err != nil {
				t.Fatal(err)
			}
			got := decodedFields(t, wire)
			var scoped *ast.Symbol
			for _, candidate := range c.GetSymbolsInScope(qualifier, accessed.Flags) {
				if candidate.Name == name.Text() {
					scoped = c.GetExportSymbolOfSymbol(candidate)
					break
				}
			}
			if got[len(got)-1] != strconv.FormatUint(p.symbolID(scoped), 10) {
				t.Fatalf("scope identity differs: %q", got)
			}
			own := c.GetSymbolAtLocation(qualifier)
			if got[2] != "1" || got[3] != strconv.FormatUint(p.symbolID(own), 10) {
				t.Fatalf("own symbol differs: %q", got)
			}
			if own.Flags&ast.SymbolFlagsAlias != 0 {
				alias := c.GetAliasedSymbol(own)
				if alias == nil || alias == own {
					t.Fatal("alias control unresolved")
				}
				// Own declaration details occupy a variable-length prefix; find
				// the second symbol's presence and identity through direct framing.
				expected := &fields{}
				expected.number(1)
				expected.text("scope-export-symbols")
				p.writeSymbolDetails(expected, own)
				p.writeSymbolDetails(expected, alias)
				expected.number(p.symbolID(scoped))
				if wire != expected.String() {
					t.Fatal("alias declarations differ")
				}
				aliased = true
			}
			if scoped != accessed {
				shadowed = true
			}
			checked++
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if checked != 4 || !shadowed || !aliased {
		t.Fatalf("controls missing: %d shadow=%v alias=%v", checked, shadowed, aliased)
	}
	for _, question := range []string{"scope-export-symbols", "scope-export-symbols\n02\nx", "scope-export-symbols\n-1\nx", "scope-export-symbols\n4294967296\nx"} {
		if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", question); err == nil {
			t.Fatalf("accepted malformed %q", question)
		}
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "scope-export-symbols\n2\nx"); err == nil {
		t.Fatal("accepted non-entity node")
	}
	t.Logf("%d direct-checker scope/alias controls, malformed flags and node kind rejected", checked)
}
