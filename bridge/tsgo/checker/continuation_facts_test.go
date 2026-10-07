package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestWave15ContinuationRawFacts(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.a")
	config := filepath.Join(dir, "tsconfig.json")
	for name, text := range map[string]string{"tsconfig.json": `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext"},"files":["prelude.d.ts"]}`, "prelude.d.ts": `declare module "store" {export const value:number;} declare module "node:test" {interface MockTracker {method(object:object,key:string):void;} export const mock:MockTracker;}`, "input.a": `import * as Store from 'store';import {mock} from 'node:test';mock.method(Store,'value');declare const maybe:number|null;maybe;`} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
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
	var call, target, nullable *ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindCallExpression {
			call = n
		}
		if n.Kind == ast.KindIdentifier && n.Text() == "Store" {
			target = n
		}
		if n.Kind == ast.KindIdentifier && n.Text() == "maybe" {
			nullable = n
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	inspect := func(n *ast.Node, kind, q string) []string {
		wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), kind, q)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	chain := inspect(call, "CallExpression", "call-declaration-chain")
	if chain[2] != "99" || chain[4] != "MethodSignature" || chain[7] != "InterfaceDeclaration" || chain[8] != "MockTracker" {
		t.Fatalf("chain %q", chain)
	}
	binding := inspect(target, "Identifier", "namespace-binding")
	if binding[2] != "1" || binding[4] != "1" || binding[5] != "NamespaceImport" || binding[6] != "1" || binding[7] != "0" {
		t.Fatalf("binding %q", binding)
	}
	own := c.GetTypeAtLocation(nullable)
	id := (&graph{program: p}).id(own)
	shape := inspect(nullable, "Identifier", "nonnullable-shape\n"+strconv.FormatUint(id, 10))
	root, err := strconv.Atoi(shape[5])
	if err != nil {
		t.Fatal(err)
	}
	if p.typesByID[root-1] != c.GetNonNullableType(own) {
		t.Fatal("nonnullable identity differs from checker")
	}
	for _, test := range []struct {
		n    *ast.Node
		k, q string
	}{{target, "Identifier", "call-declaration-chain"}, {call, "CallExpression", "namespace-binding"}, {call, "CallExpression", "call-declaration-chain\nextra"}, {target, "Identifier", "namespace-binding\nextra"}, {nullable, "Identifier", "nonnullable-shape\n0"}, {nullable, "Identifier", "nonnullable-shape\n01"}} {
		if _, err := p.Inspect(file, uint64(test.n.Pos()), uint64(test.n.End()), test.k, test.q); err == nil {
			t.Fatalf("accepted %s", test.q)
		}
	}
}
