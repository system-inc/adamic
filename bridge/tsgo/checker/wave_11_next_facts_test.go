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
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

func TestWave11NextRawQuestions(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.ts")
	source := `type Outcome=({a:1})|(({b:2}));declare function outcome():Outcome;declare function asyncOutcome():Promise<Outcome>;outcome();asyncOutcome();'x'.trim();export {};`
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"lib":["ES2022"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	selectNode := func(kind ast.Kind, text string) *ast.Node {
		var selected *ast.Node
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			if node.Kind == kind && source[scanner.GetTokenPosOfNode(node, sf, false):node.End()] == text {
				selected = node
			}
			node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(sf.AsNode())
		if selected == nil {
			t.Fatalf("missing %s %q", kind, text)
		}
		return selected
	}
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	call := selectNode(ast.KindCallExpression, "asyncOutcome()")
	awaited := ask(call, "awaited-type-shape")
	direct := checker.Checker_getAwaitedType(c, c.GetTypeAtLocation(call))
	if awaited[3] != "1" || p.typesByID[mustUint(t, awaited[5])-1] != direct {
		t.Fatalf("awaited graph differs from checker: %v", awaited)
	}
	root := ask(call, "raw-shape")[5]
	if root == awaited[5] {
		t.Fatal("Promise was not unwrapped")
	}
	value := selectNode(ast.KindCallExpression, "outcome()")
	raw := ask(value, "raw-shape")
	rootID := mustUint(t, raw[5])
	subject := p.typesByID[rootID-1]
	for _, part := range subject.Types() {
		id := p.typeIDs[part]
		got := ask(value, "declaration-lineage\n"+strconv.FormatUint(id, 10))
		decls := part.Symbol().Declarations
		if mustUint(t, got[2]) != uint64(len(decls)) {
			t.Fatalf("declaration count differs: %v", got)
		}
		cursor := 3
		for _, decl := range decls {
			count := int(mustUint(t, got[cursor]))
			cursor++
			parent := decl
			for i := 0; i < count; i++ {
				if parent == nil {
					t.Fatal("extra lineage ancestor")
				}
				if got[cursor] != lineageKey(parent) || got[cursor+1] != strings.TrimPrefix(parent.Kind.String(), "Kind") || got[cursor+3] != file {
					t.Fatalf("lineage differs: %v", got[cursor:])
				}
				children := int(mustUint(t, got[cursor+6]))
				var directChildren []*ast.Node
				parent.ForEachChild(func(n *ast.Node) bool { directChildren = append(directChildren, n); return false })
				if children != len(directChildren) {
					t.Fatal("child count differs")
				}
				for j, child := range directChildren {
					if got[cursor+7+j] != lineageKey(child) {
						t.Fatal("child identity differs")
					}
				}
				target := got[cursor+7+children]
				if parent.Kind == ast.KindTypeAliasDeclaration && target != lineageKey(parent.AsTypeAliasDeclaration().Type) {
					t.Fatal("declared type link differs")
				}
				cursor += 8 + children
				parent = parent.Parent
			}
			if parent != nil {
				t.Fatal("missing lineage ancestor")
			}
		}
		if cursor != len(got) {
			t.Fatal("trailing lineage data")
		}
	}
	method := selectNode(ast.KindIdentifier, "trim")
	got := ask(method, "declaration-lineage")
	if got[2] == "0" || got[5] != "MethodSignature" || got[9] != "1" {
		t.Fatalf("library provenance missing: %v", got)
	}
	for _, query := range []string{"declaration-lineage\n0", "declaration-lineage\n01", "declaration-lineage\n999999999", "declaration-lineage\n1\nextra", "awaited-type-shape\nextra"} {
		if _, err := p.Inspect(file, uint64(value.Pos()), uint64(value.End()), "CallExpression", query); err == nil {
			t.Fatalf("accepted malformed %q", query)
		}
	}
}

func mustUint(t *testing.T, text string) uint64 {
	t.Helper()
	value, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
