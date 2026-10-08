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
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

func TestTypeOperations(t *testing.T) {
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input.ts")
	source := `interface B { f():void };class D implements B { field=1;async f(){} };const d=new D();const m=new Map<string,number>();declare function call(cb:()=>void):void;call(async()=>{});const é=1;const object={é};missing;export {};`
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`, file: source} {
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
	pick := func(kind ast.Kind, text string) *ast.Node {
		var result *ast.Node
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			if node.Kind == kind && source[scanner.GetTokenPosOfNode(node, sf, false):node.End()] == text {
				result = node
			}
			node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(sf.AsNode())
		if result == nil {
			t.Fatalf("missing %s %q", kind, text)
		}
		return result
	}
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatalf("%s: %v", question, err)
		}
		return decodedFields(t, wire)
	}
	root := func(node *ast.Node) string { fields := ask(node, "raw-shape"); return fields[5] }
	d, m, call := pick(ast.KindIdentifier, "d"), pick(ast.KindIdentifier, "m"), pick(ast.KindCallExpression, "call(async()=>{})")
	metadata := ask(d, "type-operations\nmetadata\n"+root(d))
	// Member resolution adds a cache bit during serialization, independently
	// in each leased checker. Compare semantic flags, not resolution timing.
	typ := c.GetTypeAtLocation(d)
	if metadata[2] != strconv.FormatUint(uint64(typ.ObjectFlags()&0xfffff), 10) || metadata[3] != strconv.Itoa(len(checker.Checker_getPropertiesOfType(c, typ))) || metadata[7] != "D" || metadata[9] != "ClassDeclaration" {
		t.Fatalf("metadata differs: %q; direct flags %d properties %d", metadata, typ.ObjectFlags(), len(checker.Checker_getPropertiesOfType(c, typ)))
	}
	iterator := ask(m, "type-operations\niterator\n"+root(m))
	if iterator[2] != "1" {
		t.Fatalf("Map iterator absent: %q", iterator)
	}
	signatures := ask(call.Expression(), "type-operations\nsignatures\n"+root(call.Expression()))
	if signatures[2] != "1" || signatures[4] != "1" || signatures[6] != "0" {
		t.Fatalf("signature parameters differ: %q", signatures)
	}
	contextual := ask(call, "type-operations\nargument\n0")
	if contextual[5] != "1" {
		t.Fatalf("contextual argument absent: %q", contextual)
	}
	heritage := ask(pick(ast.KindClassDeclaration, "class D implements B { field=1;async f(){} }"), "type-operations\nheritage\n0")
	if heritage[5] != "1" {
		t.Fatalf("heritage absent: %q", heritage)
	}
	for _, question := range []string{"type-operations", "type-operations\nraw\n00", "type-operations\nraw\n0\nextra", "type-operations\nmetadata\n0", "type-operations\niterator\n999999999", "type-operations\nargument\n2", "type-operations\nproperty\n" + root(m), "type-operations\nheritage\n0", "type-operations\naccess-name\n0"} {
		if _, err := p.Inspect(file, uint64(d.Pos()), uint64(d.End()), "Identifier", question); err == nil {
			t.Fatalf("accepted invalid type operation %q", question)
		}
	}
}
