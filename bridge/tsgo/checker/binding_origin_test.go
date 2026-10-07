package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func TestBindingOriginContracts(t *testing.T) {
	d := t.TempDir()
	file := filepath.Join(d, "input.a")
	config := filepath.Join(d, "tsconfig.json")
	ambient := filepath.Join(d, "ambient.d.ts")
	source := `function f(){}f=1;({f}={});declare const local:number;local;ambient;`
	for path, text := range map[string]string{file: source, ambient: "declare const ambient:number;", config: `{"compilerOptions":{"strict":true},"files":["input.a","ambient.d.ts"]}`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	count := 0
	var nodes []*ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		count++
		if n.Kind == ast.KindIdentifier {
			nodes = append(nodes, n)
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	wire, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "binding-structure")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if fields[3] != strconv.Itoa(count) {
		t.Fatal("wrong syntax population")
	}
	for _, n := range nodes {
		wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "binding-origin")
		if err != nil {
			t.Fatal(err)
		}
		fields := decodedFields(t, wire)
		c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
		plain := c.GetSymbolAtLocation(n)
		var shorthand *ast.Symbol
		if n.Parent.Kind == ast.KindShorthandPropertyAssignment {
			shorthand = c.GetShorthandAssignmentValueSymbol(n.Parent)
		}
		cursor := 2
		for _, s := range []*ast.Symbol{plain, shorthand} {
			present := "0"
			var declarations []*ast.Node
			if s != nil {
				present = "1"
				declarations = s.Declarations
			}
			if fields[cursor] != present || fields[cursor+1] != strconv.Itoa(len(declarations)) {
				t.Fatal("wrong binding metadata")
			}
			cursor += 2
			for _, decl := range declarations {
				f := ast.GetSourceFileOfNode(decl)
				flag := "0"
				if f.IsDeclarationFile {
					flag = "1"
				}
				if fields[cursor] != f.FileName() || fields[cursor+1] != flag || fields[cursor+2] != decl.Kind.String()[4:] || fields[cursor+3] != strconv.Itoa(decl.Pos()) || fields[cursor+4] != strconv.Itoa(decl.End()) {
					t.Fatal("wrong declaration")
				}
				cursor += 5
			}
		}
		release()
		if cursor != len(fields) {
			t.Fatal("trailing facts")
		}
		for _, bad := range []string{"binding-origin\nextra", "binding-structure"} {
			if _, e := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", bad); e == nil {
				t.Fatal("accepted invalid request")
			}
		}
	}
}
