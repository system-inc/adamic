package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestWave03CheckerQuestions(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := `interface Callable extends Function {():void;():string;new():object} declare const call:Callable; type T=Callable;export {type Callable};call();`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"jsxFactory":"h.create","jsxFragmentFactory":"h.Fragment"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	var call *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			got := ask(node, "node-binding")
			own, _ := strconv.ParseUint(got[2], 10, 64)
			exported, _ := strconv.ParseUint(got[3], 10, 64)
			if own != p.symbolID(c.GetSymbolAtLocation(node)) || exported != p.symbolID(c.GetExportSpecifierLocalTargetSymbol(node)) || got[4] != strconv.Itoa(map[bool]int{false: 0, true: 1}[ast.IsPartOfTypeNode(node)]) {
				t.Fatalf("binding facts differ: %q", got)
			}
			if node.Text() == "call" {
				call = node
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if call == nil {
		t.Fatal("missing call")
	}
	graph := ask(call, "raw-shape")
	id := graph[5]
	got := ask(call, "signature-kinds\n"+id)
	subject := c.GetTypeAtLocation(call)
	constructs := c.GetSignaturesOfType(subject, checker.SignatureKindConstruct)
	calls := c.GetSignaturesOfType(subject, checker.SignatureKindCall)
	if got[2] != strconv.Itoa(len(constructs)) || got[3] != strconv.Itoa(len(calls)) || len(constructs) != 1 || len(calls) != 2 {
		t.Fatalf("signature counts differ %q", got)
	}
	for i, signature := range calls {
		result := checker.Checker_getReturnTypeOfSignature(c, signature)
		if got[4+i*2] != "1" || got[5+i*2] != strconv.FormatUint(uint64(result.Flags()), 10) {
			t.Fatalf("return flags differ %q", got)
		}
	}
	options := ask(sf.AsNode(), "import-runtime-options")
	if strings.Join(options[2:], "|") != "0|h.create|h.Fragment" {
		t.Fatalf("options differ %q", options)
	}
	for _, question := range []string{"signature-kinds", "signature-kinds\n0", "signature-kinds\n01", "signature-kinds\n999999", "signature-kinds\n" + id + "\nextra", "node-binding\nextra", "import-runtime-options"} {
		if _, err := p.Inspect(file, uint64(call.Pos()), uint64(call.End()), "Identifier", question); err == nil {
			t.Fatalf("accepted malformed question %q", question)
		}
	}
	if _, err := p.Inspect(file, 0, uint64(sf.End()), "SourceFile", "node-binding"); err == nil {
		t.Fatal("node-binding accepted a source file")
	}
}
