package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

func TestWave19CheckerQuestions(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(strconv.FormatBool(enabled), func(t *testing.T) {
			directory := t.TempDir()
			file := filepath.Join(directory, "input.a")
			config := filepath.Join(directory, "tsconfig.json")
			source := `declare const o:{[key:string]:number;fixed:number};o['dynamic'];o['fixed'];declare const n:{[key:number]:number};n['missing'];declare const a:any;a['missing'];export {};`
			if err := os.WriteFile(file, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"isolatedDeclarations":`+strconv.FormatBool(enabled)+`,"noPropertyAccessFromIndexSignature":`+strconv.FormatBool(enabled)+`},"files":["input.a"]}`), 0600); err != nil {
				t.Fatal(err)
			}
			p, err := Open(config, []string{file})
			if err != nil {
				t.Fatal(err)
			}
			sf := p.Compiler.GetSourceFile(file)
			c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
			defer release()
			ask := func(node *ast.Node, question string) []string {
				t.Helper()
				wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
				if err != nil {
					t.Fatal(err)
				}
				return decodedFields(t, wire)
			}
			expected := "0"
			if enabled {
				expected = "1"
			}
			if got := ask(sf.AsNode(), "isolated-declarations"); len(got) != 3 || got[2] != expected {
				t.Fatalf("isolated option: %q", got)
			}
			if got := ask(sf.AsNode(), "index-signature-access"); len(got) != 3 || got[2] != expected {
				t.Fatalf("index option: %q", got)
			}
			count := 0
			var walk func(*ast.Node)
			walk = func(node *ast.Node) {
				if node.Kind == ast.KindElementAccessExpression {
					count++
					access := node.AsElementAccessExpression()
					got := ask(node, "index-signature-access")
					subject := c.GetNonNullableType(c.GetTypeAtLocation(access.Expression))
					property := c.GetSymbolAtLocation(access.ArgumentExpression)
					if property == nil && access.ArgumentExpression.Kind == ast.KindStringLiteral {
						for _, candidate := range c.GetPropertiesOfType(subject) {
							if candidate.Name == access.ArgumentExpression.Text() {
								property = candidate
								break
							}
						}
					}
					presence := "0"
					if property != nil {
						presence = "1"
					}
					infos := c.GetIndexInfosOfType(subject)
					if got[2] != expected || got[3] != presence || got[4] != strconv.Itoa(len(infos)) || len(got) != 5+len(infos) {
						t.Fatalf("access facts %q", got)
					}
					for i, info := range infos {
						if got[5+i] != strconv.FormatUint(uint64(info.KeyType().Flags()), 10) {
							t.Fatal("wrong index key flags")
						}
					}
					token := scanner.GetTokenPosOfNode(node, sf, false)
					if source[token:node.End()] == "o['dynamic']" && (presence != "0" || len(infos) != 1) {
						t.Fatal("missing string-index positive")
					}
					for _, question := range []string{"isolated-declarations", "isolated-declarations\nextra", "index-signature-access\nextra"} {
						if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "ElementAccessExpression", question); err == nil {
							t.Fatalf("accepted invalid %q", question)
						}
					}
				}
				if node.Kind == ast.KindIdentifier {
					got := ask(node, "node-symbol-origin")
					symbol := c.GetSymbolAtLocation(node)
					if symbol == nil {
						if len(got) != 3 || got[2] != "0" {
							t.Fatalf("absent symbol: %q", got)
						}
					} else {
						if got[2] != "1" || got[3] != symbol.Name || got[4] != strconv.Itoa(len(symbol.Declarations)) || len(got) != 5+3*len(symbol.Declarations) {
							t.Fatalf("symbol origins: %q", got)
						}
						for i, declaration := range symbol.Declarations {
							origin := ast.GetSourceFileOfNode(declaration)
							isDeclaration, isLibrary := "0", "0"
							if origin.IsDeclarationFile {
								isDeclaration = "1"
							}
							if p.Compiler.IsSourceFileDefaultLibrary(origin.Path()) {
								isLibrary = "1"
							}
							if got[5+3*i] != origin.FileName() || got[6+3*i] != isDeclaration || got[7+3*i] != isLibrary {
								t.Fatalf("wrong declaration origin: %q", got)
							}
						}
					}
					if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "node-symbol-origin\nextra"); err == nil {
						t.Fatal("accepted symbol-origin suffix")
					}

					if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "index-signature-access"); err == nil {
						t.Fatal("accepted Identifier")
					}
				}
				node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
			}
			walk(sf.AsNode())
			if count != 4 {
				t.Fatalf("checked %d accesses", count)
			}
		})
	}
}
