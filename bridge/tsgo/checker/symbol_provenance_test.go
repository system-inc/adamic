package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSymbolProvenance(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	ambient := filepath.Join(directory, "node.d.ts")
	for path, source := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ESNext"},"files":["node.d.ts"]}`, ambient: `declare module 'node:fs' { export function existsSync(path:string):boolean; }`, file: "import {existsSync as check} from 'node:fs'; const 世界='🌍';check(世界);missing;class C{#secret=1;}"} {
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	var call, constant, private, missing, root *ast.Node
	root = p.Compiler.GetSourceFile(file).AsNode()
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			switch node.Text() {
			case "check":
				call = node
			case "世界":
				constant = node
			case "missing":
				missing = node
			}
		}
		if node.Kind == ast.KindPrivateIdentifier {
			private = node
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(root)
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	alias := ask(call, "symbol-provenance\nalias")
	own := ask(call, "symbol-provenance")
	if alias[2] == "0" || alias[2] == own[2] || alias[3] != "existsSync" || alias[4] != "1" || alias[6] != "FunctionDeclaration" || alias[9] != "1" || alias[10] != "node:fs" {
		t.Fatalf("alias provenance: %q vs %q", alias, own)
	}
	local := ask(constant, "symbol-provenance\nalias")
	if local[3] != "世界" || local[6] != "VariableDeclaration" || local[9] != "0" || local[10] != "" || local[11] != "1" || local[16] != "1" {
		t.Fatalf("const provenance: %q", local)
	}
	secret := ask(private, "symbol-provenance")
	if secret[2] == "0" || secret[6] != "PropertyDeclaration" {
		t.Fatalf("private provenance: %q", secret)
	}
	absent := ask(missing, "symbol-provenance")
	if len(absent) != 3 || absent[2] != "0" {
		t.Fatalf("unresolved provenance: %q", absent)
	}
	for _, probe := range []struct {
		node     *ast.Node
		question string
	}{{root, "symbol-provenance"}, {constant, "symbol-provenance\nwrong"}, {constant, "symbol-provenance\nalias\nextra"}} {
		if _, err := p.Inspect(file, uint64(probe.node.Pos()), uint64(probe.node.End()), strings.TrimPrefix(probe.node.Kind.String(), "Kind"), probe.question); err == nil {
			t.Fatal("malformed provenance query accepted")
		}
	}
}
