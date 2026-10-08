package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestTypeAliasInfo(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"lib":["ES2022"]},"files":["prelude.d.ts"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "prelude.d.ts"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	source := "declare const readonlyError: Readonly<Error>; readonlyError;\ndeclare const p:{then(...callbacks:((x:number)=>void)[]):void};p;\n"
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var node, receiver *ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier && n.Text() == "readonlyError" {
			node = n
		}
		if n.Kind == ast.KindIdentifier && n.Text() == "p" {
			receiver = n
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if node == nil {
		t.Fatal("missing control")
	}
	ask := func(question string) (string, error) {
		return p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", question)
	}
	raw, err := ask("raw-shape")
	if err != nil {
		t.Fatal(err)
	}
	id := decodedFields(t, raw)[5]
	direct := c.GetTypeAtLocation(node)
	alias := checker.Type_alias(direct)
	wire, err := ask("type-alias-info\n" + id)
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if alias == nil || len(alias.TypeArguments()) != 1 || fields[2] != "1" || fields[3] != alias.Symbol().Name {
		t.Fatalf("wrong alias: %q", fields)
	}
	count, err := strconv.Atoi(fields[4])
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < count; i++ {
		f := ast.GetSourceFileOfNode(alias.Symbol().Declarations[i])
		at := 5 + i*3
		if fields[at] != string(f.FileName()) || fields[at+1] != "1" || fields[at+2] != "1" {
			t.Fatalf("wrong origin: %q", fields)
		}
	}
	at := 5 + count*3
	if strings.Join(fields[at:at+3], "|") != "1|1|1" || p.typesByID[mustID(t, fields[at+3])-1] != alias.TypeArguments()[0] {
		t.Fatalf("wrong argument graph: %q", fields)
	}
	for _, suffix := range []string{"", "\n0", "\n01", "\n999999", "\n" + id + "\nextra"} {
		if _, err := ask("type-alias-info" + suffix); err == nil {
			t.Fatal("accepted malformed identity", suffix)
		}
	}

	thenWire, err := p.Inspect(file, uint64(receiver.Pos()), uint64(receiver.End()), "Identifier", "then-callback-parameters")
	if err != nil {
		t.Fatal(err)
	}
	thenFields := decodedFields(t, thenWire)
	if thenFields[4] != "1" {
		t.Fatalf("missing rest callback: %q", thenFields)
	}
	property := checker.Checker_getPropertyOfType(c, c.GetTypeAtLocation(receiver), "then")
	signature := c.GetSignaturesOfType(c.GetTypeOfSymbolAtLocation(property, receiver), checker.SignatureKindCall)[0]
	parameter := checker.Signature_parameters(signature)[0]
	parameterType := checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(parameter, receiver))
	callbackType := checker.Checker_getIndexTypeOfType(c, parameterType, checker.Checker_numberType(c))
	if p.typesByID[mustID(t, thenFields[5])-1] != callbackType {
		t.Fatal("wrong rest callback identity")
	}
	if _, err := p.Inspect(file, uint64(receiver.Pos()), uint64(receiver.End()), "Identifier", "then-callback-parameters\nextra"); err == nil {
		t.Fatal("accepted suffix")
	}
	t.Log("Rest callback element type agrees with direct checker")
	t.Log("Readonly alias, default-library declaration and argument identity match direct checker; malformed identities refused")
}
