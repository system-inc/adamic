package checker

import (
	"context"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestLeakedNumericAndConstraintFacts(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.a")
	config := filepath.Join(dir, "tsconfig.json")
	for name, text := range map[string]string{"tsconfig.json": `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["prelude.d.ts"]}`, "prelude.d.ts": "", "input.a": `0n;1n;0;2;function f<T extends number>(value:T){value;}export {};`} {
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
	count := 0
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		tpe := c.GetTypeAtLocation(n)
		if n.Kind == ast.KindBigIntLiteral || n.Kind == ast.KindNumericLiteral {
			id := (&graph{program: p}).id(tpe)
			wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), n.Kind.String()[4:], "numeric-literal\n"+strconv.FormatUint(id, 10))
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			value, held := tpe.AsLiteralType().Value().(fmt.Stringer)
			if !held || fields[2] != "1" || fields[3] != value.String() {
				t.Fatalf("literal differs: %q", fields)
			}
			count++
		}
		if n.Kind == ast.KindIdentifier && n.Text() == "value" {
			id := (&graph{program: p}).id(tpe)
			wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "base-constraint-shape\n"+strconv.FormatUint(id, 10))
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			root, err := strconv.Atoi(fields[5])
			if err != nil {
				t.Fatal(err)
			}
			if p.typesByID[root-1] != checker.Checker_getBaseConstraintOfType(c, tpe) {
				t.Fatal("constraint differs")
			}
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if count != 4 {
		t.Fatal(count)
	}
	for _, q := range []string{"numeric-literal", "numeric-literal\n0", "numeric-literal\n01", "base-constraint-shape", "base-constraint-shape\n0", "base-constraint-shape\n01"} {
		if _, err := p.Inspect(file, 0, uint64(source.End()), "SourceFile", q); err == nil {
			t.Fatalf("accepted %s", q)
		}
	}
}
