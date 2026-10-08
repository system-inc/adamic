package checker

import (
	"context"
	"fmt"
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

func TestWave11CheckerQuestions(t *testing.T) {
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input.ts")
	source := `declare const rows:number[];declare const record:{length:number};rows;record;
declare function format(x:string):string;declare function format(x:number):number;
format('x');format('y');format(1);export {};`
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
	for _, name := range []string{"rows", "record"} {
		node := selectNode(ast.KindIdentifier, name)
		id := ask(node, "raw-shape")[5]
		got := ask(node, "number-index-type\n"+id)
		index := checker.Checker_getIndexTypeOfType(c, c.GetTypeAtLocation(node), checker.Checker_numberType(c))
		want := "0"
		if index != nil {
			want = "1"
		}
		if got[2] != want {
			t.Fatalf("%s index presence %s != direct %s", name, got[2], want)
		}
		if index != nil && got[3] != strconv.FormatUint(uint64(index.Flags()), 10) {
			t.Fatalf("numeric index flags differ: %v", got)
		}
	}
	left := selectNode(ast.KindCallExpression, "format('x')")
	for _, text := range []string{"format('y')", "format(1)"} {
		right := selectNode(ast.KindCallExpression, text)
		query := fmt.Sprintf("resolved-signature-equal\n%d\n%d\nCallExpression", right.Pos(), right.End())
		got := ask(left, query)
		want := "0"
		if c.GetResolvedSignature(left) == c.GetResolvedSignature(right) {
			want = "1"
		}
		if got[2] != want {
			t.Fatalf("signature identity %s != direct %s", got[2], want)
		}
	}
	for _, test := range []struct{ pattern, want string }{{"x", "1"}, {"[", "0"}, {"(?=x)x", "1"}, {"(x)\\1", "1"}, {"[z-a]", "0"}, {"世界🌍", "1"}, {"a\nb", "1"}} {
		if got := ask(left, "regular-expression-syntax\n"+test.pattern); got[2] != test.want {
			t.Fatalf("regex %q: %v", test.pattern, got)
		}
	}
	for _, question := range []string{"number-index-type", "number-index-type\n0", "number-index-type\n01", "number-index-type\n999999999", "resolved-signature-equal", "resolved-signature-equal\n0\n1\nIdentifier", "regular-expression-syntax"} {
		if _, err := p.Inspect(file, uint64(left.Pos()), uint64(left.End()), "CallExpression", question); err == nil {
			t.Fatalf("accepted malformed %q", question)
		}
	}
	node := selectNode(ast.KindIdentifier, "rows")
	query := fmt.Sprintf("resolved-signature-equal\n%d\n%d\nCallExpression", left.Pos(), left.End())
	if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", query); err == nil {
		t.Fatal("signature question accepted a value node")
	}
}
