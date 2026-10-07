package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func wave20NextProgram(t *testing.T) (*Program, string) {
	t.Helper()
	directory := t.TempDir()
	file := filepath.Join(directory, "file.a")
	config := filepath.Join(directory, "tsconfig.json")
	for name, text := range map[string]string{
		"file.a":        "import { helper } from './helper.js'; import type { T } from './helper.js'; declare function missing():void; console.log('世界🌍'); helper(); missing(); import('./helper.js'); import(something); require('./helper.js'); const { log: destructured } = console; destructured('ok'); function binding({ value }: { value: number }) { return value; } class Computed { [something]() {} } export {};\n",
		"helper.ts":     "export type T=number;export function helper():void{console.log('hello');}\n",
		"fixture.d.ts":  "declare function require(name:string):unknown;declare const something:string;\n",
		"tsconfig.json": `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022","DOM"]},"sourceExtensions":[".a"]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	return program, file
}
func TestDeclarationChainFacts(t *testing.T) {
	t.Parallel()
	program, file := wave20NextProgram(t)
	source := program.Compiler.GetSourceFile(file)
	c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	count := 0
	structuredNames := 0
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			wire, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "declaration-chain")
			if err != nil {
				t.Fatal(err)
			}
			fields := decodedFields(t, wire)
			symbol := c.GetSymbolAtLocation(node)
			if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
				symbol = c.GetAliasedSymbol(symbol)
			}
			if symbol == nil {
				if strings.Join(fields[2:], ",") != "0,0,,0" {
					t.Fatal(fields)
				}
			} else {
				if fields[3] != strconv.FormatUint(uint64(symbol.Flags), 10) || fields[4] != symbol.Name || fields[5] != strconv.Itoa(len(symbol.Declarations)) {
					t.Fatalf("symbol mismatch %q", fields)
				}
				at := 6
				for _, decl := range symbol.Declarations {
					f := ast.GetSourceFileOfNode(decl)
					if fields[at] != "1" || fields[at+1] != f.FileName() {
						t.Fatal("source mismatch", fields)
					}
					depth, err := strconv.Atoi(fields[at+5])
					if err != nil {
						t.Fatal(err)
					}
					at += 6
					seen := 0
					for current := decl; current != nil; current = current.Parent {
						if fields[at] != strings.TrimPrefix(current.Kind.String(), "Kind") || fields[at+2] != strconv.Itoa(current.Pos()) || fields[at+3] != strconv.Itoa(current.End()) || fields[at+4] != strconv.FormatUint(uint64(current.Flags), 10) {
							t.Fatal("ancestry mismatch", fields)
						}
						if named := current.Name(); named != nil && (named.Kind == ast.KindObjectBindingPattern || named.Kind == ast.KindArrayBindingPattern || named.Kind == ast.KindComputedPropertyName) {
							structuredNames++
							if fields[at+1] != "" {
								t.Fatal("structured name has text", fields)
							}
						}
						at += 6
						seen++
					}
					if seen != depth {
						t.Fatal("wrong ancestry depth")
					}
				}
				if at != len(fields) {
					t.Fatal("trailing fields")
				}
			}
			count++
			if _, err := program.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "declaration-chain\nextra"); err == nil {
				t.Fatal("suffix accepted")
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if count < 6 || structuredNames < 2 {
		t.Fatal("vacuous node probes")
	}
}
