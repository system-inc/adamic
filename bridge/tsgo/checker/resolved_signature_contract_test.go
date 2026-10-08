package checker

import (
	"context"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestResolvedSignatureMatchesChecker(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "file.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := "import { mock } from 'node:test'; mock.method({}, 'read');\n"
	foreign := "declare module 'node:test' { interface MockTracker { /** foreign comment must not be serialized */ method(target: object, name: string): number; } export const mock: MockTracker; }\n"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler"}}`, filepath.Join(directory, "node.d.ts"): foreign} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(file)))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var call *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindCallExpression {
			call = node
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if call == nil {
		t.Fatal("missing call")
	}
	sig := c.GetResolvedSignature(call)
	if sig == nil || sig.Declaration() == nil {
		t.Fatal("fixture did not resolve declaration")
	}
	declaration := sig.Declaration()
	ds := ast.GetSourceFileOfNode(declaration)
	parent := declaration.Parent
	boolText := func(v bool) string {
		if v {
			return "1"
		}
		return "0"
	}
	kind := func(node *ast.Node) string { return strings.TrimPrefix(node.Kind.String(), "Kind") }
	graph := &graph{program: p}
	want := []string{"1", "resolved-signature", "1", strconv.FormatUint(graph.id(c.GetReturnTypeOfSignature(sig)), 10), "1", ds.FileName().AsString(), kind(declaration), strconv.Itoa(declaration.Pos()), strconv.Itoa(declaration.End()), strconv.FormatUint(uint64(declaration.Flags), 10), boolText(ds.IsDeclarationFile), kind(parent), parent.Name().Text(), strconv.Itoa(parent.Pos()), strconv.Itoa(parent.End())}
	var parents []*ast.Node
	for node := parent; node != nil; node = node.Parent {
		parents = append(parents, node)
	}
	want = append(want, strconv.Itoa(len(parents)))
	for _, node := range parents {
		name, nk := "", ""
		if n := node.Name(); n != nil {
			nk = kind(n)
			if ast.IsPropertyNameLiteral(n) || n.Kind == ast.KindPrivateIdentifier {
				name = n.Text()
			}
		}
		want = append(want, kind(node), name, nk)
	}
	wire, err := p.Inspect(file, uint64(call.Pos()), uint64(call.End()), "CallExpression", "resolved-signature")
	if err != nil {
		t.Fatal(err)
	}
	got := decodedFields(t, wire)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("resolved signature got %q, actual checker requires %q", got, want)
	}
	if strings.Contains(wire, "foreign comment") || strings.Contains(wire, "target: object") {
		t.Fatal("signature metadata serialized foreign body")
	}
}
