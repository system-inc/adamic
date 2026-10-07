package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReferenceSymbolOrigins(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "config.json")
	ambient := filepath.Join(directory, "environment.d.ts")
	for path, source := range map[string]string{config: `{"compilerOptions":{"target":"ES2022","lib":["ES2022"]},"files":["environment.d.ts"]}`, ambient: `declare const foo: number;`, file: `foo; const object={foo}; export {foo as renamed}; function f(){var undefined; ({undefined}=object); missing;}`} {
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	var bare, shorthand, exported, decl, write, missing *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			switch node.Text() {
			case "foo":
				if node.Parent.Kind == ast.KindShorthandPropertyAssignment {
					shorthand = node
				} else if node.Parent.Kind == ast.KindExportSpecifier {
					exported = node
				} else {
					bare = node
				}
			case "undefined":
				if node.Parent.Kind == ast.KindVariableDeclaration {
					decl = node
				} else {
					write = node
				}
			case "missing":
				missing = node
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	root := p.Compiler.GetSourceFile(file).AsNode()
	visit(root)
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	for _, node := range []*ast.Node{bare, shorthand, exported} {
		got := ask(node, "reference-symbol-origins")
		if len(got) != 9 || got[2] == "0" || got[3] != "1" || got[4] != ambient || got[5] != "VariableDeclaration" || got[8] != "1" {
			t.Fatalf("ambient read: %q", got)
		}
	}
	own := ask(decl, "reference-symbol-origins")
	target := ask(write, "reference-symbol-origins")
	if strings.Join(own, "|") != strings.Join(target, "|") || own[4] != file || own[8] != "0" {
		t.Fatalf("binding identity: %q vs %q", own, target)
	}
	absent := ask(missing, "reference-symbol-origins")
	if len(absent) != 3 || absent[2] != "0" {
		t.Fatalf("unresolved: %q", absent)
	}
	for _, probe := range []struct {
		node     *ast.Node
		question string
	}{{root, "reference-symbol-origins"}, {bare, "reference-symbol-origins\nalias"}} {
		if _, err := p.Inspect(file, uint64(probe.node.Pos()), uint64(probe.node.End()), strings.TrimPrefix(probe.node.Kind.String(), "Kind"), probe.question); err == nil {
			t.Fatal("invalid query accepted")
		}
	}
}
