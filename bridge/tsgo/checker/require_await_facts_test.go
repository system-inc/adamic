package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequireAwaitRawQuestions(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "source.ts")
	source := "interface Contract { run: () => Promise<number>; }\nclass Subject implements Contract { async run() { return 1; } }\nfunction identity<T>(value: T): T { return value; }\nconst callback = identity(async () => 1);\n"
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`, file: source} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	nodes := map[ast.Kind]*ast.Node{}
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		nodes[node.Kind] = node
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(file))).AsNode())
	ask := func(kind ast.Kind, question string) []string {
		t.Helper()
		node := nodes[kind]
		if node == nil {
			t.Fatal("missing selected node", kind)
		}
		wire, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(question, err)
		}
		return decodedFields(t, wire)
	}
	function := ask(ast.KindArrowFunction, "require-await\nfunction")
	if len(function) < 9 || function[2] != "function" || function[3] != "1024" || function[5] != "1" || function[6] != "NumericLiteral" {
		t.Fatalf("wrong function metadata: %q", function)
	}
	resolved := ask(ast.KindCallExpression, "require-await\nresolved")
	if len(resolved) < 8 || resolved[2] != "resolved" || resolved[3] != "1" || resolved[4] != "1" || resolved[5] != "1" {
		t.Fatalf("missing declared generic target: %q", resolved)
	}
	heritage := ask(ast.KindClassDeclaration, "require-await\nheritage")
	if len(heritage) < 8 || heritage[2] != "heritage" || heritage[5] != "1" {
		t.Fatalf("wrong raw heritage: %q", heritage)
	}
	shape := ask(ast.KindArrowFunction, "raw-shape")
	id := shape[5]
	signatures := ask(ast.KindArrowFunction, "require-await\ntype\n"+id+"\nsignatures")
	if len(signatures) < 5 || signatures[3] != "1" {
		t.Fatalf("wrong raw call signatures: %q", signatures)
	}
	for _, question := range []string{"require-await", "require-await\nunknown", "require-await\nfunction\nextra", "require-await\ntype\n0\nsignatures", "require-await\ntype\n01\nsignatures", "require-await\ntype\n" + id + "\nunknown", "require-await\ntype\n" + id + "\nproperty-global"} {
		node := nodes[ast.KindArrowFunction]
		if _, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "ArrowFunction", question); err == nil {
			t.Fatalf("accepted invalid question %q", question)
		}
	}
}
