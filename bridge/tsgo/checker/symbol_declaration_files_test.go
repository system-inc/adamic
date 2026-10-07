package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSymbolDeclarationFiles(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := "import {String as Function} from './ambient';Function();new Symbol();missing();function f(Number:any){new Number()};\n"
	for name, text := range map[string]string{"input.a": source, "ambient.d.ts": "export declare class String{};\n", "tsconfig.json": `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"]},"files":["input.a","ambient.d.ts"]}`} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	seen := map[string]bool{}
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier && (n.Text() == "Function" || n.Text() == "Symbol" || n.Text() == "missing" || n.Text() == "Number") {
			wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "symbol-declaration-files")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			symbol := c.GetSymbolAtLocation(n)
			count := 0
			if symbol != nil {
				count = len(symbol.Declarations)
			}
			if len(fields) != 3+2*count || fields[2] != strconv.Itoa(count) {
				t.Fatalf("bad count %q", fields)
			}
			for i := 0; i < count; i++ {
				decl := ast.GetSourceFileOfNode(symbol.Declarations[i])
				want := "0"
				if decl.IsDeclarationFile {
					want = "1"
				}
				if fields[3+i*2] != decl.FileName() || fields[4+i*2] != want {
					t.Fatalf("wrong declaration %s %q", n.Text(), fields)
				}
			}
			if n.Text() == "Function" && count > 0 && fields[3] != file {
				t.Fatal("alias was followed")
			}
			if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "symbol-declaration-files\nextra"); err == nil {
				t.Fatal("extra accepted")
			}
			seen[n.Text()] = true
		}
		n.ForEachChild(visit)
		return false
	}
	sf.AsNode().ForEachChild(visit)
	for _, name := range []string{"Function", "Symbol", "missing", "Number"} {
		if !seen[name] {
			t.Fatal("missing control", name)
		}
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "symbol-declaration-files"); err == nil {
		t.Fatal("wrong kind accepted")
	}
	if !strings.HasSuffix(file, ".a") {
		t.Fatal("wrong extension")
	}
}
