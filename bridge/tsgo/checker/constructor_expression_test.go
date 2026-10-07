package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestConstructorExpressionContract(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := "/* 世界 */new String('x');new ((Symbol))();new Function;"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true},"files":["input.a"]}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindNewExpression {
			count++
			expression := node.AsNewExpression().Expression
			wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "NewExpression", "constructor-expression")
			if err != nil {
				t.Fatal(err)
			}
			got := decodedFields(t, wire)
			expected := []string{"1", "constructor-expression", "1", strings.TrimPrefix(expression.Kind.String(), "Kind"), strconv.Itoa(expression.Pos()), strconv.Itoa(expression.End())}
			if strings.Join(got, "|") != strings.Join(expected, "|") {
				t.Fatalf("metadata: %v want %v", got, expected)
			}
			for _, question := range []string{"constructor-expression\nextra", "constructor-expression\n"} {
				if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "NewExpression", question); err == nil {
					t.Fatal("accepted extra question fields")
				}
			}
			if _, err := p.Inspect(file, uint64(expression.Pos()), uint64(expression.End()), strings.TrimPrefix(expression.Kind.String(), "Kind"), "constructor-expression"); err == nil {
				t.Fatal("accepted a nonconstructor")
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(p.Compiler.GetSourceFile(file).AsNode())
	if count != 3 {
		t.Fatalf("constructors: %d", count)
	}
}
