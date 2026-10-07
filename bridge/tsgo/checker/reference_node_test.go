package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestReferenceNodeFacts(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.ts")
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`, file: "let m=JSON; m(); const object={JSON}; JSON=1; export {m};\n"} {
		if e := os.WriteFile(path, []byte(text), 0600); e != nil {
			t.Fatal(e)
		}
	}
	p, e := Open(config, []string{file})
	if e != nil {
		t.Fatal(e)
	}
	sf := p.Compiler.GetSourceFile(file)
	var jsonID, mID string
	var seenShorthand, seenExport, seenWrite bool
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier && (n.Text() == "JSON" || n.Text() == "m") {
			wire, e := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "reference-node")
			if e != nil {
				t.Fatal(e)
			}
			f := decodedFields(t, wire)
			if f[4] != "1" {
				t.Fatalf("missing symbol: %q", f)
			}
			if n.Text() == "JSON" {
				if jsonID != "" && jsonID != f[5] {
					t.Fatal("shorthand changed read symbol", jsonID, f[5])
				}
				jsonID = f[5]
				if f[8] != "1" {
					t.Fatalf("global declaration not in library: %q", f)
				}
			} else {
				if mID != "" && mID != f[5] {
					t.Fatal("local export changed read symbol", mID, f[5])
				}
				mID = f[5]
				if f[8] != "0" {
					t.Fatalf("source declaration mislabeled: %q", f)
				}
			}
			if n.Parent.Kind == ast.KindShorthandPropertyAssignment {
				seenShorthand = true
			}
			if n.Parent.Kind == ast.KindExportSpecifier {
				seenExport = true
			}
			if n.Text() == "JSON" && f[3] == "1" {
				seenWrite = true
			}
			if n.Parent.Kind == ast.KindVariableDeclaration && n.Text() == "m" && f[2] != "1" {
				t.Fatal("declaration-name flag missing")
			}
			if _, e := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "reference-node\nunknown"); e == nil {
				t.Fatal("malformed suffix accepted")
			}
		}
		n.ForEachChild(visit)
		return false
	}
	sf.AsNode().ForEachChild(visit)
	if !seenShorthand || !seenExport || !seenWrite || strings.TrimSpace(jsonID) == "" {
		t.Fatal("positive reference fact controls not reached")
	}
}
