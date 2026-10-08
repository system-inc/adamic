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

func TestBindingState(t *testing.T) {
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input.ts")
	source := `const é=1;const object={é};function f(p:number){return p+é;}missing;export {};`
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
	checked, shorthand, missing := 0, false, false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "binding-state\nextra"); err == nil {
				t.Fatal("accepted binding-state suffix on valid Identifier")
			}
			wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "binding-state")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			symbol := c.GetSymbolAtLocation(node)
			if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
				symbol = c.GetShorthandAssignmentValueSymbol(node.Parent)
				shorthand = true
			}
			id, err := strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			if symbol == nil {
				if id != 0 || fields[5] != "0" {
					t.Fatalf("unresolved identifier supplied declarations: %q", fields)
				}
				missing = true
			} else {
				if id == 0 || p.symbolsByID[id-1] != symbol {
					t.Fatalf("wrong binding identity: %q", fields)
				}
				checked++
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	if checked < 5 || !shorthand || !missing {
		t.Fatalf("controls incomplete: %d %v %v", checked, shorthand, missing)
	}
	for _, question := range []string{"binding-state", "binding-state\nextra"} {
		if _, err := p.Inspect(file, uint64(sf.Pos()), uint64(sf.End()), strings.TrimPrefix(sf.Kind.String(), "Kind"), question); err == nil {
			t.Fatalf("accepted invalid binding request %q", question)
		}
	}
}
