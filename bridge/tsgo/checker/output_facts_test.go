package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputCheckerQuestions(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	helper := filepath.Join(directory, "helper.ts")
	config := filepath.Join(directory, "tsconfig.json")
	for path, text := range map[string]string{
		config:                                `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["node.d.ts"]}`,
		filepath.Join(directory, "node.d.ts"): `declare namespace NodeJS { interface Process { exit(code?:number):never; stdout:{write(s:string):void}; } } declare var process:NodeJS.Process;`,
		helper:                                `export function print():void {}`,
		file:                                  `import {print} from './helper.js';print();process.exit(0);`,
	} {
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
	var exit, call *ast.Node
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier && n.Text() == "exit" {
			exit = n
		}
		if n.Kind == ast.KindCallExpression && n.AsCallExpression().Expression.Kind == ast.KindIdentifier {
			call = n
		}
		n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(source.AsNode())
	ask := func(n *ast.Node, q string) (string, error) {
		return p.Inspect(file, uint64(n.Pos()), uint64(n.End()), strings.TrimPrefix(n.Kind.String(), "Kind"), q)
	}
	wire, err := ask(exit, "platform-symbol")
	if err != nil {
		t.Fatal(err)
	}
	f := decodedFields(t, wire)
	if !strings.Contains(strings.Join(f, "|"), "MethodSignature|exit|") || !strings.Contains(strings.Join(f, "|"), "InterfaceDeclaration|Process|") || !strings.Contains(strings.Join(f, "|"), "ModuleDeclaration|NodeJS|") {
		t.Fatalf("missing raw ancestry: %q", f)
	}
	wire, err = ask(call, "call-declaration")
	if err != nil {
		t.Fatal(err)
	}
	f = decodedFields(t, wire)
	declaration := c.GetResolvedSignature(call).Declaration()
	if f[2] != "1" || f[3] != ast.GetSourceFileOfNode(declaration).FileName() || !strings.Contains(strings.Join(f, "|"), "FunctionDeclaration|print|") {
		t.Fatalf("wrong resolved call: %q", f)
	}
	wire, err = ask(source.AsNode(), "program-modules")
	if err != nil {
		t.Fatal(err)
	}
	f = decodedFields(t, wire)
	if !strings.Contains(strings.Join(f, "|"), "import|0|1|"+helper) {
		t.Fatalf("resolved module absent: %q", f)
	}
	wire, err = ask(source.AsNode(), "output-flow")
	if err != nil {
		t.Fatal(err)
	}
	f = decodedFields(t, wire)
	if f[2] == "0" || !strings.Contains(strings.Join(f, "|"), "expression|CallExpression|") {
		t.Fatalf("raw graph missing call: %q", f)
	}
	wire, err = ask(source.AsNode(), "source-context")
	if err != nil {
		t.Fatal(err)
	}
	f = decodedFields(t, wire)
	if strings.Join(f, "|") != "1|source-context|1|0" {
		t.Fatalf("source context: %q", f)
	}
	if _, err := ask(source.AsNode(), "platform-symbol"); err == nil {
		t.Fatal("accepted wrong platform node")
	}
	for _, q := range []string{"platform-symbol", "call-declaration", "program-modules", "output-flow", "source-context"} {
		var n *ast.Node
		switch q {
		case "platform-symbol":
			n = exit
		case "call-declaration":
			n = call
		default:
			n = source.AsNode()
		}
		if _, err := ask(n, q+"\nextra"); err == nil {
			t.Fatalf("accepted suffix: %s", q)
		}
	}
	for _, q := range []string{"platform-symbol", "call-declaration", "program-modules", "output-flow", "source-context"} {
		if _, err := ask(exit, q); err == nil && q != "platform-symbol" {
			t.Fatalf("accepted wrong node: %s", q)
		}
	}
}
