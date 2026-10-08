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

func TestWave14NextRawQuestions(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := `declare const input:HTMLInputElement|null;document.addEventListener('click',event=>input);`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"files":["input.a"],"compilerOptions":{"strict":true,"module":"ESNext","lib":["ES2022","DOM"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var input, document, call *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			switch node.Text() {
			case "input":
				input = node
			case "document":
				document = node
			}
		}
		if node.Kind == ast.KindCallExpression {
			call = node
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	origin := ask(document, "symbol-context")
	if origin[2] != strconv.FormatUint(p.symbolID(c.GetSymbolAtLocation(document)), 10) || origin[4] != "1" || origin[6] != "VariableDeclaration" || origin[9] != "1" {
		t.Fatalf("global declaration %q", origin)
	}
	shape := ask(input, "raw-shape")
	metadata := ask(input, "interface-bases\n"+shape[5])
	id := mustID(t, metadata[3])
	if p.typesByID[id-1] != c.GetNonNullableType(c.GetTypeAtLocation(input)) {
		t.Fatal("nonnullable identity differs")
	}
	if metadata[2] != strconv.FormatUint(uint64(c.GetTypeAtLocation(input).Flags()), 10) {
		t.Fatal("raw flags differ")
	}
	signature := ask(call, "signature-context")
	if signature[2] != "99" || signature[3] == "0" || signature[4] != "MethodSignature" {
		t.Fatalf("signature ancestors %q", signature)
	}
	for _, question := range []string{"interface-bases", "interface-bases\n0", "interface-bases\n01", "interface-bases\n999999", "symbol-context\nsuffix", "signature-context\nsuffix", "signature-context"} {
		if _, err := p.Inspect(file, uint64(input.Pos()), uint64(input.End()), "Identifier", question); err == nil {
			t.Fatal("invalid question accepted", question)
		}
	}
}
