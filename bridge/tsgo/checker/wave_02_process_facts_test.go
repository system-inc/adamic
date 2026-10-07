package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestWave02ProcessFacts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.ts")
	config := filepath.Join(directory, "tsconfig.json")
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext"}}`, file: `import {f} from './helper'; async function main(){f(); await f();} main();`, filepath.Join(directory, "helper.ts"): `export function f(){return 1;}`} {
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
	var name *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindCallExpression && node.AsCallExpression().Expression.Kind == ast.KindIdentifier && node.AsCallExpression().Expression.Text() == "f" {
			call = node
			name = node.AsCallExpression().Expression
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
		fields := decodedFields(t, wire)
		if fields[0] != "1" || fields[1] != question {
			t.Fatal(fields)
		}
		return fields
	}
	signature := c.GetResolvedSignature(call)
	declaration := signature.Declaration()
	got := ask(call, "resolved-call")
	want := []string{"1", "resolved-call", "1", "1", ast.GetSourceFileOfNode(declaration).FileName(), strings.TrimPrefix(declaration.Kind.String(), "Kind"), strconv.Itoa(declaration.Pos()), strconv.Itoa(declaration.End()), "1", strconv.FormatUint(uint64(c.GetReturnTypeOfSignature(signature).Flags()), 10)}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("signature facts %q, want %q", got, want)
	}
	metadata := ask(call, "ast-context")
	if metadata[2] != "0" || metadata[3] != "1" || metadata[5] != "1" || metadata[6] != "FunctionDeclaration" {
		t.Fatalf("context %q", metadata)
	}
	graph := ask(sf.AsNode(), "execution-graph")
	blocks, err := strconv.Atoi(graph[2])
	if err != nil || blocks < 1 {
		t.Fatalf("graph %q", graph)
	}
	modules := ask(sf.AsNode(), "program-modules")
	if !strings.Contains(strings.Join(modules, "|"), filepath.Join(directory, "helper.ts")) {
		t.Fatal("missing resolved module")
	}
	for _, question := range []string{"ast-context\nextra", "execution-graph\nextra", "resolved-call\nextra", "program-modules\nextra", "execution-graph", "resolved-call"} {
		if _, err := p.Inspect(file, uint64(name.Pos()), uint64(name.End()), "Identifier", question); err == nil {
			t.Fatalf("accepted invalid question %q", question)
		}
	}
}
