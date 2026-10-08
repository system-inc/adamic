package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestWave07SymbolContext(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	for path, text := range map[string]string{
		config:                                   `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["es2022"]},"files":["ambient.d.ts"]}`,
		filepath.Join(directory, "ambient.d.ts"): `export {}; declare global { function setTimeout(callback:()=>void): number; }`,
		file:                                     `function f(){return 1;} f(); setTimeout(()=>{});`,
	} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	identifiers, calls := 0, 0
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier || node.Kind == ast.KindCallExpression {
			wire, err := p.wave07SymbolContext(c, node, "wave07-symbol-context")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			if fields[0] != "1" || fields[1] != "wave07-symbol-context" {
				t.Fatal(fields)
			}
			if fields[2] != strconv.FormatUint(p.symbolID(c.GetSymbolAtLocation(node)), 10) {
				t.Fatalf("symbol identity differs from direct checker: %v", fields)
			}
			if node.Kind == ast.KindIdentifier {
				identifiers++
			}
			if node.Kind == ast.KindCallExpression {
				calls++
				if node.Expression().Kind == ast.KindIdentifier && node.Expression().Text() == "f" {
					declaration := c.GetResolvedSignature(node).Declaration()
					if !strings.Contains(wire, source.Text()[declaration.Body().Pos():declaration.Body().End()]) {
						t.Fatal("selected body absent")
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if calls != 2 || identifiers < 3 {
		t.Fatalf("coverage %d calls %d identifiers", calls, identifiers)
	}
	if _, err := p.wave07SymbolContext(c, source.AsNode(), "wave07-symbol-context\nextra"); err == nil {
		t.Fatal("suffix accepted")
	}
}
