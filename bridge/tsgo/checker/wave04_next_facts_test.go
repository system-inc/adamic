package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWave04NextRawFactsAndDispatchGap(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := "function write(): void {}\nwrite();\nexport {};\n"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["prelude.d.ts"]}`, filepath.Join(directory, "prelude.d.ts"): "declare const positiveControl: number;"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var call *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindCallExpression {
			call = node
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	if call == nil {
		t.Fatal("positive control call absent")
	}
	var contextFields fields
	wire, err := p.wave04NextDeclarationContext(&contextFields, c, call.Expression(), "wave04-next-declaration-context")
	if err != nil {
		t.Fatal(err)
	}
	got := decodedFields(t, wire)
	if len(got) < 9 || got[0] != "1" || got[3] != "write" || got[4] != "1" || got[5] != file || got[6] != "0" || got[7] != "0" || got[8] != "1" {
		t.Fatalf("wrong raw declaration context: %v", got)
	}
	var signatureFields fields
	wire, err = p.wave04NextResolvedSignatureDeclaration(&signatureFields, c, call, "wave04-next-resolved-signature-declaration")
	if err != nil {
		t.Fatal(err)
	}
	got = decodedFields(t, wire)
	if len(got) < 10 || got[0] != "1" || got[1] != "16" || got[2] != "1" || got[3] != file || got[len(got)-3] != "1" {
		t.Fatalf("wrong selected signature or body: %v", got)
	}
	for _, question := range []string{"wave04-next-declaration-context", "wave04-next-resolved-signature-declaration"} {
		_, err := p.Inspect(file, uint64(call.Pos()), uint64(call.End()), "CallExpression", question)
		if err == nil {
			t.Logf("%s: public dispatch is registered", question)
			continue
		}
		if !strings.Contains(err.Error(), "unsupported checker question") {
			t.Fatalf("expected dispatch integration gap for %s: %v", question, err)
		}
		t.Logf("%s: direct raw facts pass; public dispatch rejects: %v", question, err)
	}
}
