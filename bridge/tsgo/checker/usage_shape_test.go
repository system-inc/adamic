package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestUsageShapeFactsAndRefusals(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := "function identity<T extends string>(x:T):T {return x;}\nidentity;\n"
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	owner := sf.Statements.Nodes[0]
	wire, err := p.Inspect(file, uint64(owner.Pos()), uint64(owner.End()), "FunctionDeclaration", "usage-shape")
	if err != nil {
		t.Fatal(err)
	}
	values := decodedFields(t, wire)
	if len(values) < 4 || values[0] != "1" || values[1] != "usage-shape" {
		t.Fatalf("bad header: %q", values)
	}
	root, _ := strconv.Atoi(values[2])
	length, _ := strconv.Atoi(values[3])
	if root != 1 || length < 4 {
		t.Fatalf("missing type graph: %q", values)
	}
	type record struct {
		flags      uint64
		start, end int
		file       string
		constraint int
		lists      [][]int
	}
	records := []record{}
	at := 4
	natural := func() int {
		n, e := strconv.Atoi(values[at])
		at++
		if e != nil || n < 0 {
			t.Fatal("bad natural")
		}
		return n
	}
	for id := 0; id < length; id++ {
		r := record{flags: uint64(natural())}
		natural()
		r.start = natural()
		r.end = natural()
		r.file = values[at]
		at++
		r.constraint = natural()
		natural()
		at += 4
		for list := 0; list < 9; list++ {
			count := natural()
			var ids []int
			for i := 0; i < count; i++ {
				n := natural()
				if n > length {
					t.Fatal("invalid graph edge")
				}
				ids = append(ids, n)
			}
			r.lists = append(r.lists, ids)
		}
		records = append(records, r)
	}
	if at != len(values) {
		t.Fatal("trailing graph fields")
	}
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	signature := c.GetSignatureFromDeclaration(owner)
	parameterType := c.GetTypeOfSymbol(signature.Parameters()[0])
	typeParameter := signature.TypeParameters()[0]
	if parameterType != typeParameter || c.GetReturnTypeOfSignature(signature) != typeParameter {
		t.Fatal("control does not share its parameter and return type")
	}
	found := false
	foundSignature := false
	for _, r := range records {
		if r.flags == uint64(checker.TypeFlagsTypeParameter) {
			declaration := typeParameter.Symbol().Declarations[0]
			if r.start != declaration.Pos() || r.end != declaration.End() || r.file != file || r.constraint == 0 || records[r.constraint-1].flags != uint64(checker.TypeFlagsString) {
				t.Fatalf("parameter topology differs: %+v", r)
			}
			found = true
		}
		if r.flags == 0 && len(r.lists[6]) == 1 && len(r.lists[7]) == 1 && len(r.lists[8]) == 1 && r.lists[6][0] == r.lists[7][0] && r.lists[7][0] == r.lists[8][0] {
			foundSignature = true
		}
	}
	if !foundSignature {
		t.Fatal("no signature shares its parameter, type parameter and return identity")
	}
	if !found {
		t.Fatal("no type parameter declaration")
	}
	for _, q := range []string{"usage-shape\nextra", "unknown-question"} {
		if _, err := p.Inspect(file, uint64(owner.Pos()), uint64(owner.End()), "FunctionDeclaration", q); err == nil {
			t.Fatalf("accepted %q", q)
		}
	}
	identifier := sf.Statements.Nodes[1].AsExpressionStatement().Expression
	if identifier.Kind != ast.KindIdentifier {
		t.Fatal("missing value control")
	}
	if _, err := p.Inspect(file, uint64(identifier.Pos()), uint64(identifier.End()), "Identifier", "usage-shape"); err == nil || !strings.Contains(err.Error(), "requires a signature") {
		t.Fatalf("usage-shape accepted a value node: %v", err)
	}
}
