package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestWave01CheckerQuestions(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	source := `/** @deprecated 世界 🌍 */ function old() {};
function partial(b:boolean) {if(b)return 1;}
function complete(b:boolean) {if(b)return 1;return 2;}
old();
function collision(a:number,b:boolean){if(b)return 1;else{let a=2;}}
`
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022"},"sourceExtensions":[".a"],"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
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
	var functions []*ast.Node
	var use, branch *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindFunctionDeclaration {
			functions = append(functions, node)
		}
		if node.Kind == ast.KindIdentifier && node.Text() == "old" && node.Parent.Kind == ast.KindCallExpression {
			use = node
		}
		if node.Kind == ast.KindIfStatement && node.AsIfStatement().ElseStatement != nil {
			branch = node
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if len(functions) != 4 || use == nil || branch == nil {
		t.Fatal("missing controls")
	}
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	for _, node := range functions {
		facts := ask(node, "function-return-flow")
		if len(facts) != 16 {
			t.Fatalf("return flow fields: %q", facts)
		}
		implicit := checker.Checker_functionHasImplicitReturn(c, node)
		if (facts[9] == "1") != implicit {
			t.Fatalf("reachability differs: %q", facts)
		}
		signature := checker.Checker_getSignatureFromDeclaration(c, node)
		flags := checker.Checker_getReturnTypeOfSignature(c, signature).Flags()
		if facts[11] != strconv.FormatUint(uint64(flags), 10) {
			t.Fatal("inferred return flags differ")
		}
	}
	if ask(functions[1], "function-return-flow")[9] != "1" || ask(functions[2], "function-return-flow")[9] != "0" {
		t.Fatal("reachability has no opposing controls")
	}
	docs := ask(use, "symbol-documentation\nnode")
	if docs[2] != "1" || docs[5] != "1" || docs[6] != "1" || docs[7] != "世界 🌍" {
		t.Fatalf("documentation: %q", docs)
	}
	call := ask(use.Parent, "symbol-documentation\nsignature")
	if call[2] != "1" || call[3] != "世界 🌍" {
		t.Fatalf("signature documentation: %q", call)
	}
	scope := ask(branch, "scope-symbol-declarations")
	if !strings.Contains(strings.Join(scope, "|"), "|a|1|1|"+file+"|0|") {
		t.Fatalf("parameter absent from scope")
	}
	for _, question := range []string{"function-return-flow\nextra", "symbol-documentation", "symbol-documentation\nproperty\n01\na", "symbol-documentation\nunknown", "scope-symbol-declarations\nextra"} {
		if _, err := p.Inspect(file, uint64(use.Pos()), uint64(use.End()), "Identifier", question); err == nil {
			t.Fatalf("accepted malformed %q", question)
		}
	}
	if checker.TypeFlagsUndefined != 4 || checker.TypeFlagsNever != 262144 || checker.TypeFlagsAny != 1 || checker.TypeFlagsStringLiteral != 1024 || checker.TypeFlagsNumberLiteral != 2048 {
		t.Fatal("native flag constants differ")
	}
}
