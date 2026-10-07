package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

func TestWave03FinalNumericKinds(t *testing.T) {
	data, err := os.ReadFile("../../../stage1/cohere/typeaware/wave_03_final/syntax_kinds.a")
	if err != nil {
		t.Fatal(err)
	}
	entries := regexp.MustCompile(`result.set\('([^']+)', (\d+)\)`).FindAllSubmatch(data, -1)
	if len(entries) < 300 {
		t.Fatal("kind table incomplete")
	}
	entries = append(entries, regexp.MustCompile(`export const K([A-Za-z0-9]+) = (\d+);`).FindAllSubmatch(data, -1)...)
	for _, entry := range entries {
		number, err := strconv.Atoi(string(entry[2]))
		if err != nil {
			t.Fatal(err)
		}
		if got := ast.Kind(number).String(); got != "Kind"+string(entry[1]) {
			t.Fatalf("numeric kind %s=%d differs from parser %s", entry[1], number, got)
		}
	}
}
func TestWave03FinalRawQuestions(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := `declare function factory<T>(value:()=>T):T;const tuple:[Promise<number>]=[Promise.resolve(1)];factory(async()=>1);`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var call, tuple, function *ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindCallExpression && n.Expression().Kind == ast.KindIdentifier && n.Expression().Text() == "factory" {
			call = n
		}
		if n.Kind == ast.KindVariableDeclaration && n.Name().Text() == "tuple" {
			tuple = n.Name()
		}
		if n.Kind == ast.KindFunctionDeclaration {
			function = n
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if call == nil || tuple == nil || function == nil {
		t.Fatal("fixture nodes missing")
	}
	query := func(node *ast.Node, question string) ([]string, error) {
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), node.Kind.String()[4:], question)
		if err != nil {
			return nil, err
		}
		return decodedFields(t, wire), nil
	}
	values, err := query(call, "generic-signature-shape")
	if err != nil {
		t.Fatal(err)
	}
	number := func(index int) uint64 {
		value, err := strconv.ParseUint(values[index], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	// The raw argument bounds and kind come from the parser, independently of Adamic.
	if number(2) != 1 || number(3) != uint64(call.Arguments()[0].Pos()) || number(4) != uint64(call.Arguments()[0].End()) || number(5) != uint64(call.Arguments()[0].Kind) || number(6) != 0 || number(7) != 1 || number(8) != 1 {
		t.Fatal("generic argument or signature metadata differs")
	}
	resolved := c.GetResolvedSignature(call)
	declared := resolved.Target()
	typeParameter := number(9)
	if p.typesByID[typeParameter-1] != declared.TypeParameters()[0] {
		t.Fatal("declared type parameter differs")
	}
	if number(10) != 0 || number(11) != 1 || p.typesByID[number(12)-1] != checker.Checker_getTypeOfSymbol(c, declared.Parameters()[0]) {
		t.Fatal("declared parameter differs")
	}
	if number(13) != 0 || number(14) != 0 || number(15) != 1 || p.typesByID[number(16)-1] != checker.Checker_getTypeOfSymbol(c, resolved.Parameters()[0]) {
		t.Fatal("instantiated parameter differs")
	}
	if number(17) != 0 || p.typesByID[number(18)-1] != c.GetReturnTypeOfSignature(declared) {
		t.Fatal("declared return differs")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	tupleType := c.GetTypeAtLocation(tuple)
	tupleID := g.id(tupleType)
	functionType := c.GetTypeAtLocation(function)
	functionID := g.id(functionType)
	indexed, err := query(call, "indexed-type-shape\n"+strconv.FormatUint(tupleID, 10)+"\nnumber")
	if err != nil {
		t.Fatal(err)
	}
	if indexed[3] != "1" || indexed[4] != "1" {
		t.Fatal("number index missing")
	}
	id, err := strconv.ParseUint(indexed[5], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	if p.typesByID[id-1] != checker.Checker_getIndexTypeOfType(c, tupleType, checker.Checker_numberType(c)) {
		t.Fatal("indexed type differs from direct checker")
	}
	returns, err := query(call, "callable-return-shape\n"+strconv.FormatUint(functionID, 10))
	if err != nil {
		t.Fatal(err)
	}
	if returns[4] != "1" {
		t.Fatal("callable signature missing")
	}
	id, err = strconv.ParseUint(returns[5], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	if p.typesByID[id-1] != c.GetReturnTypeOfSignature(c.GetSignaturesOfType(functionType, checker.SignatureKindCall)[0]) {
		t.Fatal("callable return differs from direct checker")
	}
	for _, question := range []string{"generic-signature-shape\n0", "indexed-type-shape\n0\nnumber", "indexed-type-shape\n01\nnumber", "indexed-type-shape\n999999\nnumber", "indexed-type-shape\n1\ninvalid", "indexed-type-shape\n1", "callable-return-shape\n0", "callable-return-shape\n01", "callable-return-shape\n999999", "callable-return-shape\n1\nextra"} {
		if _, err := query(call, question); err == nil {
			t.Fatalf("accepted malformed raw question %q", question)
		}
	}
	if _, err := query(tuple, "generic-signature-shape"); err == nil {
		t.Fatal("generic signature accepted a non-call")
	}
	t.Log("raw generic, index and return facts match direct checker; ten malformed questions and non-call refused")
}
