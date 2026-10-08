package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestWave05CoreProjectionFacts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	for path, text := range map[string]string{
		config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["input.a"]}`,
		file:   "function id<T>(value:T):T{return value;}id(1);interface I{f():Promise<number>;}class C implements I{async f(){return 1;}}declare const xs:number[];xs;export {};\n",
	} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	var call, class, array *ast.Node
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n.Kind == ast.KindCallExpression {
			call = n
		}
		if n.Kind == ast.KindClassDeclaration {
			class = n
		}
		if n.Kind == ast.KindIdentifier && n.Text() == "xs" {
			array = n
		}
		n.ForEachChild(walk)
		return false
	}
	walk(p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file)).AsNode())
	if call == nil || class == nil || array == nil {
		t.Fatal("missing fixture nodes")
	}
	ask := func(n *ast.Node, q string) []string {
		wire, e := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), strings.TrimPrefix(n.Kind.String(), "Kind"), q)
		if e != nil {
			t.Fatal(e)
		}
		return decodedFields(t, wire)
	}
	declared := ask(call, "declared-call-signature")
	if len(declared) != 12 || declared[2] != "1" || declared[3] != "1" || declared[4] != declared[6] || declared[4] != declared[10] || declared[9] != "0" {
		t.Fatalf("declared generic identities: %q", declared)
	}
	inherited := ask(class, "type-projection\nheritage")
	if len(inherited) != 4 || inherited[2] != "1" || inherited[3] == "0" {
		t.Fatalf("heritage: %q", inherited)
	}
	property := ask(class, "type-projection\n"+inherited[3]+"\nproperty\nf")
	if len(property) != 5 || property[2] != "2" || property[3] == "0" || property[4] != "0" {
		t.Fatalf("property: %q", property)
	}
	sig := ask(class, "type-projection\n"+property[3]+"\nsignatures")
	if len(sig) != 6 || sig[2] != "1" || sig[3] == "0" || sig[4] != "0" || sig[5] != "0" {
		t.Fatalf("signatures: %q", sig)
	}
	shape := ask(array, "raw-shape")
	indexed := ask(array, "type-projection\n"+shape[5]+"\nindex")
	if len(indexed) != 4 || indexed[2] != "1" || indexed[3] == "0" {
		t.Fatalf("index: %q", indexed)
	}
	for _, q := range []string{"declared-call-signature\nextra", "type-projection\n01\nsignatures", "type-projection\n0\nindex", "type-projection\n" + property[3] + "\nsignatures\nextra", "type-projection\n" + property[3] + "\nunknown"} {
		if _, e := p.Inspect(file, uint64(call.Pos()), uint64(call.End()), "CallExpression", q); e == nil {
			t.Fatalf("accepted invalid question %q", q)
		}
	}
	for _, q := range []string{"declared-call-signature", "type-projection\nheritage"} {
		if _, e := p.Inspect(file, uint64(array.Pos()), uint64(array.End()), "Identifier", q); e == nil {
			t.Fatalf("accepted invalid node for %q", q)
		}
	}
}
