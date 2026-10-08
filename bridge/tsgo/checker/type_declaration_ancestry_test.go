package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestTypeDeclarationAncestry(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := `export type Result = (({outcome:'Ok';value:string})) | ({outcome:'Error';message:string});
declare function produce():Promise<Result>;
produce();`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	var call *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindCallExpression {
			call = node
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file)).AsNode())
	if call == nil {
		t.Fatal("call absent")
	}
	ask := func(question string) (string, error) {
		return p.Inspect(file, uint64(call.Pos()), uint64(call.End()), "CallExpression", question)
	}
	wire, err := ask("raw-shape")
	if err != nil {
		t.Fatal(err)
	}
	id := decodedFields(t, wire)[5]
	raw, err := ask("type-declaration-ancestry\n" + id + "\n0")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(decodedFields(t, raw), "|"), "InterfaceDeclaration|Promise") {
		t.Fatal("raw type lost its Promise declaration")
	}
	awaited, err := ask("type-declaration-ancestry\n" + id + "\n1")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, awaited)
	if fields[0] != "1" || fields[1] != "type-declaration-ancestry" || fields[2] != "2" {
		t.Fatalf("awaited union framing: %q", fields)
	}
	joined := strings.Join(fields, "|")
	for _, want := range []string{"TypeLiteral", "ParenthesizedType", "UnionType", "TypeAliasDeclaration|Result", "SourceFile", file} {
		if !strings.Contains(joined, want) {
			t.Fatalf("ancestry missing %q", want)
		}
	}
	for _, question := range []string{"type-declaration-ancestry", "type-declaration-ancestry\n0\n0", "type-declaration-ancestry\n01\n0", "type-declaration-ancestry\n9999999\n0", "type-declaration-ancestry\n" + id + "\n2", "type-declaration-ancestry\n" + id + "\n0\nextra", "unknown-question"} {
		if _, err := ask(question); err == nil {
			t.Fatalf("malformed question accepted: %q", question)
		}
	}
}
