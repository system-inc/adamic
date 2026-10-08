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

func TestGlobalBinding(t *testing.T) {
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input.ts")
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022"}}`, file: `Object=1;({Object}= {});function f(Object){Object=2;}missing;globalThis;export {};`} {
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
	checked := 0
	shorthand := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			for _, mode := range []string{"node", "value"} {
				wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "global-binding\n"+mode)
				if err != nil {
					t.Fatal(err)
				}
				fields := decodedFields(t, wire)
				symbol := c.GetSymbolAtLocation(node)
				if mode == "value" && node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
					symbol = c.GetShorthandAssignmentValueSymbol(node.Parent)
					shorthand = true
				}
				cursor := 2
				for _, symbol := range []*ast.Symbol{symbol, node.LocalSymbol()} {
					present := "0"
					count := 0
					if symbol != nil {
						present = "1"
						count = len(symbol.Declarations)
					}
					if fields[cursor] != present || fields[cursor+1] != strconv.Itoa(count) {
						t.Fatalf("binding presence/count differs: %q", fields)
					}
					cursor += 2
					if symbol != nil {
						for _, declaration := range symbol.Declarations {
							source := ast.GetSourceFileOfNode(declaration)
							present, dts := "0", "0"
							if source != nil {
								present = "1"
								if source.IsDeclarationFile {
									dts = "1"
								}
							}
							if fields[cursor] != present || fields[cursor+1] != dts {
								t.Fatalf("origins differ: %q", fields)
							}
							cursor += 2
						}
					}
				}
				if cursor != len(fields) {
					t.Fatalf("unexpected trailing fields: %q", fields)
				}
				checked++
			}
			for _, question := range []string{"global-binding", "global-binding\nnode\nextra", "global-binding\nunknown"} {
				if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", question); err == nil {
					t.Fatalf("accepted malformed global-binding %q", question)
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	if checked < 8 || !shorthand {
		t.Fatal("missing positive/shorthand controls")
	}
	if _, err := p.Inspect(file, uint64(sf.Pos()), uint64(sf.End()), strings.TrimPrefix(sf.Kind.String(), "Kind"), "global-binding\nnode"); err == nil {
		t.Fatal("accepted SourceFile binding request")
	}
}
