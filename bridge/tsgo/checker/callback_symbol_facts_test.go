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

func TestCallbackSymbolFacts(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := `/* 世界 🌍 */
const RegExp = 1;
const object = {RegExp, ['computed']: function callback(){ callback(); }};
export {RegExp as exported};
undefined; missing; declaredInDTS; Error;
declare const local: number; local;
function f(arguments: unknown) { arguments; }
`
	for path, text := range map[string]string{
		config:                                   `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["prelude.d.ts"]}`,
		file:                                     source,
		filepath.Join(directory, "prelude.d.ts"): "declare const declaredInDTS: number;\n",
	} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	identifiers := 0
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			identifiers++
			for _, variant := range []string{"", "\nread"} {
				want := c.GetSymbolAtLocation(node)
				if variant != "" && node.Parent != nil {
					if node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Parent.Name() == node {
						want = c.GetShorthandAssignmentValueSymbol(node.Parent)
					}
					if node.Parent.Kind == ast.KindExportSpecifier {
						want = c.GetExportSpecifierLocalTargetSymbol(node.Parent)
					}
				}
				wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "callback-symbol-facts"+variant)
				if err != nil {
					t.Fatal(err)
				}
				fields := decodedFields(t, wire)
				if fields[0] != "1" || fields[1] != "callback-symbol-facts" {
					t.Fatalf("wrong header: %q", fields)
				}
				id, err := strconv.ParseUint(fields[2], 10, 64)
				if err != nil {
					t.Fatal(err)
				}
				count := 0
				if want != nil {
					count = len(want.Declarations)
				}
				if fields[3] != strconv.Itoa(count) || len(fields) != 4+count {
					t.Fatalf("wrong declaration count: %q", fields)
				}
				if (want == nil && id != 0) || (want != nil && (id == 0 || p.symbolsByID[id-1] != want)) {
					t.Fatalf("wrong symbol identity for %s", node.Text())
				}
				if want != nil {
					for i, declaration := range want.Declarations {
						expected := "0"
						if ast.GetSourceFileOfNode(declaration).IsDeclarationFile {
							expected = "1"
						}
						if fields[4+i] != expected {
							t.Fatalf("wrong declaration-file flag: %q", fields)
						}
					}
				}
				for _, suffix := range []string{"\ninvalid", "\nread\nextra"} {
					if _, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), "Identifier", "callback-symbol-facts"+suffix); err == nil {
						t.Fatal("accepted malformed question", suffix)
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if identifiers < 15 {
		t.Fatal("insufficient controls", identifiers)
	}
	if _, err := p.Inspect(file, uint64(sf.AsNode().Pos()), uint64(sf.AsNode().End()), strings.TrimPrefix(sf.AsNode().Kind.String(), "Kind"), "callback-symbol-facts"); err == nil {
		t.Fatal("accepted non-identifier")
	}
	t.Logf("%d identifiers, both raw/read variants agree with direct checker; computed names, shorthand/export targets and malformed questions covered", identifiers)
}
