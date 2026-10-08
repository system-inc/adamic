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

func TestScopeSymbolsAtTheExactLabel(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	if err := os.WriteFile(file, []byte("const café=1;x:{let x=1;}café:{var y=1;}\nexport {};\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"files":["input.a"],"compilerOptions":{"strict":true,"target":"ES2022"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	labels := 0
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindLabeledStatement {
			labels++
			wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "LabeledStatement", "scope-symbols")
			if err != nil {
				t.Fatal(err)
			}
			got := decodedFields(t, wire)
			want := c.GetSymbolsInScope(node, ast.SymbolFlagsValue)
			if got[0] != "1" || got[1] != "scope-symbols" || got[2] != strconv.Itoa(len(want)) || len(got) != 3+2*len(want) {
				t.Fatal("scope frame differs", got)
			}
			actual := map[string]string{}
			for at := range want {
				actual[got[3+2*at]] = got[4+2*at]
			}
			for _, symbol := range want {
				if actual[strings.ToValidUTF8(symbol.Name, "�")] != strconv.FormatUint(uint64(symbol.Flags), 10) {
					t.Fatal("raw symbol differs", symbol.Name)
				}
			}
			for _, question := range []string{"scope-symbols\nsuffix", "scope-symbols\n"} {
				if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "LabeledStatement", question); err == nil {
					t.Fatal("accepted suffix")
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	if labels != 2 {
		t.Fatal("missing positive label anchors", labels)
	}
	if _, err := p.Inspect(file, 0, uint64(len(sf.Text())), "SourceFile", "scope-symbols"); err == nil {
		t.Fatal("accepted wrong kind")
	}
}
