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
)

func TestWave02CheckerQuestions(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input.ts")
	source := "declare const promise:Promise<string>; promise; function f<T extends string>(value:T){value;} export {};"
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`), 0600); err != nil {
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
	var nodes []*ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier && node.Parent.Kind == ast.KindExpressionStatement {
			nodes = append(nodes, node)
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if len(nodes) != 2 {
		t.Fatalf("anchors: %d", len(nodes))
	}
	for i, mode := range []string{"awaited-shape", "constraint-shape"} {
		node := nodes[i]
		ask := func(question string) ([]string, error) {
			wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", question)
			if err != nil {
				return nil, err
			}
			return decodedFields(t, wire), nil
		}
		raw, err := ask("raw-shape")
		if err != nil {
			t.Fatal(err)
		}
		got, err := ask(mode + "\n" + raw[5])
		if err != nil {
			t.Fatal(err)
		}
		subject := c.GetTypeAtLocation(node)
		var expected *checker.Type
		if i == 0 {
			expected = checker.Checker_getAwaitedType(c, subject)
		} else {
			expected = checker.Checker_getBaseConstraintOfType(c, subject)
		}
		if expected == nil || got[3] != "1" || got[4] != "1" {
			t.Fatalf("%s missing: %q", mode, got)
		}
		id, err := strconv.ParseUint(got[5], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if p.typesByID[id-1] != expected {
			t.Fatalf("%s differs from direct checker operation", mode)
		}
		for _, suffix := range []string{"", "\n0", "\n01", "\nx", "\n9007199254740993", "\n1\nextra"} {
			if _, err := ask(mode + suffix); err == nil {
				t.Fatalf("%s accepted invalid suffix %q", mode, suffix)
			}
		}
		if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "NumericLiteral", mode+"\n"+raw[5]); err == nil {
			t.Fatal("accepted inexact node kind")
		}
	}
	if uint64(checker.TypeFlagsInstantiable) != uint64(524288|2097152|4194304|8388608|16777216|33554432|67108864) {
		t.Fatal("instantiable flags changed")
	}
	if uint64(checker.TypeFlagsNever) != 262144 {
		t.Fatal("never flags changed")
	}
	// A scalar has no base constraint; absence is distinct from the scalar itself.
	node := nodes[0]
	wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "raw-shape")
	if err != nil {
		t.Fatal(err)
	}
	raw := decodedFields(t, wire)
	wire, err = p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "constraint-shape\n"+raw[5])
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if strings.Join(fields[2:], ",") != "1,0,0,0,0" {
		t.Fatalf("absent constraint: %q", fields)
	}
}
