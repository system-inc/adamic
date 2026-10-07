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

func TestTypeDeclarationAncestry(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.a")
	config := filepath.Join(dir, "tsconfig.json")
	for path, text := range map[string]string{config: `{"files":["input.a"],"compilerOptions":{"strict":true,"target":"ESNext","lib":["ESNext"]}}`, file: `type Outcome<T>=(({outcome:'ok';value:T})|({outcome:'bad';error:string}));declare function result():Promise<Outcome<number>>;result();export {};`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	var n *ast.Node
	var walk func(*ast.Node)
	walk = func(current *ast.Node) {
		if current.Kind == ast.KindCallExpression {
			n = current
		}
		current.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if n == nil {
		t.Fatal("missing call")
	}
	for _, mode := range []string{"type-declaration-ancestry", "type-declaration-ancestry\nawaited"} {
		wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "CallExpression", mode)
		if err != nil {
			t.Fatal(err)
		}
		got := decodedFields(t, wire)
		at := 2
		expect := func(want string) {
			t.Helper()
			if at >= len(got) || got[at] != want {
				t.Fatalf("field %d got %q want %q", at, got, want)
			}
			at++
		}
		subject := c.GetTypeAtLocation(n)
		if strings.HasSuffix(mode, "\nawaited") {
			subject = checker.Checker_getAwaitedType(c, subject)
		}
		parts := []*checker.Type{subject}
		if subject.Flags()&checker.TypeFlagsUnion != 0 {
			parts = subject.Types()
		}
		expect(strconv.Itoa(len(parts)))
		for _, part := range parts {
			var declarations []*ast.Node
			if part.Symbol() != nil {
				declarations = part.Symbol().Declarations
			}
			expect(strconv.Itoa(len(declarations)))
			for _, declaration := range declarations {
				expect(ast.GetSourceFileOfNode(declaration).FileName())
				var chain []*ast.Node
				for current := declaration; current != nil; current = current.Parent {
					chain = append(chain, current)
				}
				expect(strconv.Itoa(len(chain)))
				for _, current := range chain {
					expect(strings.TrimPrefix(current.Kind.String(), "Kind"))
					expect(strconv.Itoa(current.Pos()))
					expect(strconv.Itoa(current.End()))
					name := ""
					if current.Name() != nil {
						name = current.Name().Text()
					}
					expect(name)
					var child *ast.Node
					if current.Kind == ast.KindTypeAliasDeclaration {
						child = current.AsTypeAliasDeclaration().Type
					}
					if current.Kind == ast.KindParenthesizedType {
						child = current.AsParenthesizedTypeNode().Type
					}
					if child == nil {
						expect("0")
					} else {
						expect("1")
						expect(strconv.Itoa(child.Pos()))
						expect(strconv.Itoa(child.End()))
					}
					count := 0
					current.ForEachChild(func(*ast.Node) bool { count++; return false })
					expect(strconv.Itoa(count))
				}
			}
		}
		if at != len(got) {
			t.Fatal("unconsumed ancestry")
		}
		if strings.HasSuffix(mode, "\nawaited") && len(parts) != 2 {
			t.Fatal("missing awaited union arms")
		}
	}
	if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "CallExpression", "type-declaration-ancestry\nextra"); err == nil {
		t.Fatal("type-declaration-ancestry accepted a suffix")
	}
}
