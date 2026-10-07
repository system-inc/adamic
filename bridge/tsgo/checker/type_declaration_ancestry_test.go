package checker

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestTypeDeclarationAncestry(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.ts")
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	source := `type Outcome<T> = (({outcome:'Good';value:T}) | ({outcome:'Bad'}));declare const value:Outcome<string>;value;`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var anchor *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier && node.Parent.Kind == ast.KindExpressionStatement {
			anchor = node
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	if anchor == nil {
		t.Fatal("no anchor")
	}
	ask := func(question string) ([]string, error) {
		wire, err := p.Inspect(file, uint64(anchor.Pos()), uint64(anchor.End()), "Identifier", question)
		if err != nil {
			return nil, err
		}
		return decodedFields(t, wire), nil
	}
	if _, err := ask("raw-shape"); err != nil {
		t.Fatal(err)
	}
	parts := c.GetTypeAtLocation(anchor).Types()
	if len(parts) != 2 {
		t.Fatalf("arms %d", len(parts))
	}
	for _, part := range parts {
		id := (&graph{program: p}).id(part)
		actual, err := ask("type-declaration-ancestry\n" + strconv.FormatUint(id, 10))
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"1", "type-declaration-ancestry", strconv.Itoa(len(checker.Type_symbol(part).Declarations))}
		for _, declaration := range checker.Type_symbol(part).Declarations {
			var chain []*ast.Node
			for node := declaration; node != nil; node = node.Parent {
				chain = append(chain, node)
			}
			expected = append(expected, ast.GetSourceFileOfNode(declaration).FileName(), strconv.Itoa(len(chain)))
			for _, node := range chain {
				name := ""
				if (node.Kind == ast.KindTypeAliasDeclaration || node.Kind == ast.KindInterfaceDeclaration) && node.Name() != nil {
					name = node.Name().Text()
				}
				count := 0
				if node.Kind == ast.KindUnionType {
					count = len(node.AsUnionTypeNode().Types.Nodes)
				}
				expected = append(expected, strings.TrimPrefix(node.Kind.String(), "Kind"), strconv.Itoa(node.Pos()), strconv.Itoa(node.End()), name, strconv.Itoa(count))
				if node.Kind == ast.KindTypeAliasDeclaration && node.AsTypeAliasDeclaration().Type != nil {
					declared := node.AsTypeAliasDeclaration().Type
					expected = append(expected, "1", strconv.Itoa(declared.Pos()), strconv.Itoa(declared.End()), strings.TrimPrefix(declared.Kind.String(), "Kind"))
				} else {
					expected = append(expected, "0")
				}
			}
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("ancestry differs from direct AST facts:\n%q\n%q", actual, expected)
		}
	}
	for _, suffix := range []string{"", "\n0", "\n01", "\nx", "\n9007199254740993", "\n1\nextra"} {
		if _, err := ask("type-declaration-ancestry" + suffix); err == nil {
			t.Fatalf("accepted malformed suffix %q", suffix)
		}
	}
	raw, err := ask("raw-shape")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Inspect(file, uint64(anchor.Pos()), uint64(anchor.End()), "NumericLiteral", "type-declaration-ancestry\n"+raw[5]); err == nil {
		t.Fatal("accepted inexact node")
	}
	scalar := c.GetNumberType()
	empty, err := ask("type-declaration-ancestry\n" + strconv.FormatUint((&graph{program: p}).id(scalar), 10))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(empty, ",") != "1,type-declaration-ancestry,0" {
		t.Fatalf("scalar declarations %q", empty)
	}
}
