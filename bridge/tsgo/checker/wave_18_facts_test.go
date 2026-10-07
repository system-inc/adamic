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
)

func TestWave18CompilerFacts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := `class Base {get x(){return 0;}}class Child extends Base {get x(){return 1;}}
 declare const tuple:[Promise<number>,number];tuple;
 declare const sync:Iterable<number>;sync;
 declare const async:AsyncIterable<number>;async;
 declare const dispose:AsyncDisposable;dispose;
 declare const callback:(...fs:((value:number)=>void)[])=>void;callback;
 export {};`
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ESNext","lib":["ESNext"]},"files":["anchor.d.ts"]}`, filepath.Join(directory, "anchor.d.ts"): "export {};"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	selectNode := func(kind ast.Kind, text string) *ast.Node {
		var selected *ast.Node
		var walk func(*ast.Node)
		walk = func(n *ast.Node) {
			if n.Kind == kind && source[scanner.GetTokenPosOfNode(n, sf, false):n.End()] == text {
				selected = n
			}
			n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
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
	root := func(n *ast.Node) string { return ask(n, "raw-shape")[5] }
	tuple := selectNode(ast.KindIdentifier, "tuple")
	tupleID := root(tuple)
	values := ask(tuple, "iteration-type-facts\n"+tupleID+"\nvalues")
	if values[2] != "1" || values[5] != "2" {
		t.Fatalf("tuple facts %q", values)
	}
	expected := c.GetTypeArguments(c.GetTypeAtLocation(tuple))
	for i := 0; i < 2; i++ {
		if p.typesByID[mustID(t, values[6+i])-1] != expected[i] {
			t.Fatal("tuple member identity differs from checker")
		}
	}
	for _, name := range []string{"sync", "async", "dispose"} {
		n := selectNode(ast.KindIdentifier, name)
		id := root(n)
		for _, key := range []string{"iterator", "asyncIterator", "asyncDispose"} {
			fields := ask(n, "iteration-type-facts\n"+id+"\n"+key)
			direct := checker.Checker_getPropertyOfType(c, c.GetTypeAtLocation(n), checker.Checker_getPropertyNameForKnownSymbolName(c, key)) != nil
			if len(fields) != 3 || (fields[2] == "1") != direct {
				t.Fatalf("%s %s differs: %q", name, key, fields)
			}
		}
	}
	callback := selectNode(ast.KindIdentifier, "callback")
	callbackID := root(callback)
	callbacks := ask(callback, "iteration-type-facts\n"+callbackID+"\ncallbacks")
	if callbacks[4] != "1" {
		t.Fatalf("missing callback %q", callbacks)
	}
	signature := c.GetSignaturesOfType(c.GetTypeAtLocation(callback), checker.SignatureKindCall)[0]
	param := checker.Signature_parameters(signature)[0]
	direct := checker.Checker_getIndexTypeOfType(c, checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(param, callback)), checker.Checker_numberType(c))
	if p.typesByID[mustID(t, callbacks[5])-1] != direct {
		t.Fatal("rest callback element differs from checker")
	}
	getter := selectNode(ast.KindGetAccessor, "get x(){return 1;}")
	bases := ask(getter, "base-member-facts")
	if len(bases) != 10 || bases[2] != "1" || bases[4] != "1" || bases[5] != strconv.FormatUint(uint64(c.GetPropertyOfType(c.GetBaseTypes(c.GetDeclaredTypeOfSymbol(getter.Parent.Symbol()))[0], "x").Flags), 10) || bases[8] != "1" || bases[9] != "0" {
		t.Fatalf("base getter metadata %q", bases)
	}
	// Malformed identities, operations and inappropriate nodes must be refused.
	for _, question := range []string{"iteration-type-facts", "iteration-type-facts\n0\niterator", "iteration-type-facts\n01\niterator", "iteration-type-facts\n999999\nvalues", "iteration-type-facts\n" + tupleID + "\nunknown", "base-member-facts", "base-member-facts\nextra"} {
		if _, err := p.Inspect(file, uint64(tuple.Pos()), uint64(tuple.End()), "Identifier", question); err == nil {
			t.Fatalf("accepted malformed question %q", question)
		}
	}
}
