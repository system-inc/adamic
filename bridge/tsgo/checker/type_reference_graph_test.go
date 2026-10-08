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
)

func TestTypeReferenceGraphFacts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	other := filepath.Join(directory, "other.a")
	for path, text := range map[string]string{
		filepath.Join(directory, "prelude.d.ts"): `declare const probe: number;`,
		config:                                   `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["prelude.d.ts"]}`,
		file:                                     `interface Inner {[key:string]: Outer} type Wrapped = {[key:string]:Wrapped[]};`,
		other:                                    `interface Outer {[key:string]:Inner}`,
	} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file, other})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	var root *ast.Node
	symbols := map[*ast.Symbol]bool{}
	for _, path := range []string{file, other} {
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			if node.Kind == ast.KindInterfaceDeclaration && node.Name().Text() == "Inner" {
				root = node
			}
			if node.Kind == ast.KindIdentifier {
				symbols[c.GetSymbolAtLocation(node)] = true
			}
			node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(path)).AsNode())
	}
	if root == nil {
		t.Fatal("missing root")
	}
	wire, err := p.Inspect(file, uint64(root.Pos()), uint64(root.End()), "InterfaceDeclaration", "type-reference-graph")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if fields[0] != "1" || fields[1] != "type-reference-graph" || fields[2] != "1" {
		t.Fatalf("bad header %q", fields)
	}
	cursor := 4
	readNumber := func() int {
		t.Helper()
		if cursor >= len(fields) {
			t.Fatal("missing field")
		}
		value, err := strconv.Atoi(fields[cursor])
		cursor++
		if err != nil || value < 0 {
			t.Fatal("bad number")
		}
		return value
	}
	count, err := strconv.Atoi(fields[3])
	if err != nil || count < 8 {
		t.Fatalf("missing cross-file syntax: %q", fields)
	}
	kinds := map[string]int{}
	for at := 0; at < count; at++ {
		kind := fields[cursor]
		cursor++
		kinds[kind]++
		symbol := readNumber()
		if symbol > 0 && (symbol > len(p.symbolsByID) || !symbols[p.symbolsByID[symbol-1]]) {
			t.Fatal("identity differs from direct checker")
		}
		if kind == "Identifier" && symbol == 0 {
			t.Fatal("lost resolved symbol")
		}
		if body := readNumber(); body > count {
			t.Fatal("body outside graph")
		}
		for list := 0; list < 2; list++ {
			length := readNumber()
			for edge := 0; edge < length; edge++ {
				id := readNumber()
				if id == 0 || id > count {
					t.Fatal("edge outside graph")
				}
			}
		}
	}
	if cursor != len(fields) || kinds["InterfaceDeclaration"] != 2 || kinds["IndexSignature"] != 2 {
		t.Fatalf("lost syntax %v", kinds)
	}
	if _, err := p.Inspect(file, uint64(root.Pos()), uint64(root.End()), "InterfaceDeclaration", "type-reference-graph\nextra"); err == nil {
		t.Fatal("accepted graph suffix")
	}
	var wrapped *ast.Node
	source.AsNode().ForEachChild(func(node *ast.Node) bool {
		if node.Kind == ast.KindTypeAliasDeclaration {
			wrapped = node
		}
		return false
	})
	wire, err = p.Inspect(file, uint64(wrapped.Pos()), uint64(wrapped.End()), "TypeAliasDeclaration", "type-reference-graph")
	if err != nil {
		t.Fatal(err)
	}
	answer := decodedFields(t, wire)
	if !strings.Contains(strings.Join(answer, "|"), "ArrayType") {
		t.Fatal("constructor kind lost")
	}
}
