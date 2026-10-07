package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	checkerapi "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

func TestWave26NextRawQuestionContracts(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	text := "declare const value:number|null;value;"
	if err := os.WriteFile(file, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"module":"ESNext"},"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(file)
	var value *ast.Node
	count := 0
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		count++
		if n.Kind == ast.KindIdentifier && n.Parent.Kind == ast.KindExpressionStatement {
			value = n
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if value == nil {
		t.Fatal("missing control")
	}
	ask := func(n *ast.Node, q string) []string {
		t.Helper()
		wire, e := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), n.Kind.String()[4:], q)
		if e != nil {
			t.Fatal(e)
		}
		return decodedFields(t, wire)
	}
	structure := ask(source.AsNode(), "node-structure")
	if structure[2] != strconv.Itoa(int(core.ModuleKindESNext)) || structure[3] != strconv.Itoa(count) || structure[4] != "SourceFile" {
		t.Fatal("wrong raw syntax")
	}
	lineage := ask(value, "declaration-lineage\nnode")
	if lineage[2] == "0" || lineage[4] != "value" || lineage[5] != "1" {
		t.Fatal("wrong symbol metadata")
	}
	shape := ask(value, "raw-shape")
	id := shape[5]
	nonnull := ask(value, "transformed-shape\n"+id+"\nnon-nullable")
	if nonnull[9] != strconv.Itoa(int(checkerapi.TypeFlagsNumber)) {
		t.Fatal("wrong non-nullable graph")
	}
	constraint := ask(value, "transformed-shape\n"+id+"\nconstraint")
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	identity, _ := strconv.ParseUint(id, 10, 64)
	bound := checkerapi.Checker_getBaseConstraintOfType(c, p.typesByID[identity-1])
	release()
	if bound == nil {
		if constraint[3] != "0" {
			t.Fatal("invented constraint")
		}
	} else {
		root, err := strconv.ParseUint(constraint[5], 10, 64)
		if err != nil || root == 0 || p.typesByID[root-1] != bound {
			t.Fatal("wrong constraint identity")
		}
	}
	for _, q := range []string{"node-structure\nextra", "declaration-lineage", "declaration-lineage\nunknown", "declaration-lineage\ntype\n01", "literal-string\n0", "literal-string\n" + id, "transformed-shape\n" + id + "\nunknown", "transformed-shape\n01\nconstraint"} {
		if _, e := p.Inspect(file, uint64(value.Pos()), uint64(value.End()), "Identifier", q); e == nil {
			t.Fatalf("accepted %q", q)
		}
	}
	if uint64(checkerapi.TypeFlagsInstantiable) != 132644864 || uint64(checkerapi.TypeFlagsNever) != 262144 || uint64(ast.SymbolFlagsAlias) != 2097152 || uint64(ast.SymbolFlagsInterface) != 64 || uint64(ast.NodeFlagsConst) != 2 {
		t.Fatal("pinned flags changed")
	}
}
