package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestIdentifierBindingAndSourceJsx(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	file := filepath.Join(directory, "input.tsx")
	config := filepath.Join(directory, "tsconfig.json")
	source := "import {String} from './ambient';function f(){arguments;const value=1;const object={value};return <div/>};\n"
	for name, text := range map[string]string{"input.tsx": source, "ambient.d.ts": "export declare class String{};\n", "tsconfig.json": `{"compilerOptions":{"strict":true,"target":"ES2022","jsx":"preserve","lib":["ES2022"]},"files":["input.tsx","ambient.d.ts"]}`} {
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
	foundShorthand, foundArguments := false, false
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier {
			wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "identifier-binding")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			symbol := c.GetSymbolAtLocation(n)
			if n.Parent != nil && n.Parent.Kind == ast.KindShorthandPropertyAssignment {
				symbol = c.GetShorthandAssignmentValueSymbol(n.Parent)
				foundShorthand = true
			}
			if symbol == nil {
				if len(fields) != 3 || fields[2] != "0" {
					t.Fatalf("unresolved %q", fields)
				}
			} else {
				if fields[2] != "1" || fields[3] != strconv.FormatUint(p.symbolID(symbol), 10) || fields[4] != strconv.Itoa(len(symbol.Declarations)) || len(fields) != 5+len(symbol.Declarations)*5 {
					t.Fatalf("binding %q", fields)
				}
				for i, d := range symbol.Declarations {
					decl := ast.GetSourceFileOfNode(d)
					bit := "0"
					if decl.IsDeclarationFile {
						bit = "1"
					}
					if fields[5+i*5] != decl.FileName() || fields[6+i*5] != bit || fields[8+i*5] != strconv.Itoa(d.Pos()) || fields[9+i*5] != strconv.Itoa(d.End()) {
						t.Fatalf("declaration %q", fields)
					}
				}
				if n.Text() == "arguments" {
					foundArguments = true
					if len(symbol.Declarations) != 0 {
						t.Fatal("implicit arguments declared")
					}
				}
			}
			if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "identifier-binding\nextra"); err == nil {
				t.Fatal("extra accepted")
			}
		}
		n.ForEachChild(visit)
		return false
	}
	sf.AsNode().ForEachChild(visit)
	if !foundArguments || !foundShorthand {
		t.Fatal("missing controls")
	}
	wire, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "source-has-jsx")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if len(fields) != 3 || fields[2] != "1" {
		t.Fatalf("jsx %q", fields)
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "source-has-jsx\nextra"); err == nil {
		t.Fatal("extra jsx accepted")
	}
	if _, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", "identifier-binding"); err == nil {
		t.Fatal("wrong kind accepted")
	}
}
