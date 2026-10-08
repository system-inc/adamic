package checker

import (
	"context"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNarrowSymbolContracts(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "file.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := "interface M {}\ninterface M { x: number }\nlet m: M;\nconst value = 1;\n({value});\nmissing;\nPromise;\nonlyOther;\n"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["file.ts","ambient.d.ts","extra.ts"]}`, filepath.Join(directory, "ambient.d.ts"): "interface M { y: number }\n", filepath.Join(directory, "extra.ts"): "const onlyOther = 1;\n"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(file)))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	nodeAt := func(marker string) *ast.Node {
		t.Helper()
		position := strings.Index(source, marker)
		var found *ast.Node
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			if node.Kind == ast.KindIdentifier && node.Pos() <= position && node.End() > position {
				found = node
			}
			node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(sf.AsNode())
		if found == nil {
			t.Fatal("missing marker", marker)
		}
		return found
	}
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	compare := func(got, want []string) {
		t.Helper()
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("narrow contract got %q, actual checker requires %q", got, want)
		}
	}
	value := nodeAt("value});")
	own := c.GetSymbolAtLocation(value)
	short := c.GetShorthandAssignmentValueSymbol(value.Parent)
	if own == nil || short == nil || own == short {
		t.Fatal("fixture must distinguish shorthand property and value")
	}
	compare(ask(value, "symbol-identity"), []string{"1", "symbol-identity", "1", strconv.FormatUint(p.symbolID(own), 10)})
	compare(ask(value.Parent, "shorthand-value-identity"), []string{"1", "shorthand-value-identity", "1", strconv.FormatUint(p.symbolID(short), 10)})
	declaration := nodeAt("value =")
	symbol := c.GetSymbolAtLocation(declaration)
	if short != symbol {
		t.Fatal("shorthand value must name the declaration's binding")
	}
	compare(ask(declaration, "symbol-identity"), []string{"1", "symbol-identity", "1", strconv.FormatUint(p.symbolID(short), 10)})
	compare(ask(nodeAt("missing;"), "symbol-identity"), []string{"1", "symbol-identity", "0"})
	promise := nodeAt("Promise;")
	real := c.GetSymbolAtLocation(promise)
	if real == nil || len(real.Declarations) == 0 || !ast.GetSourceFileOfNode(real.Declarations[0]).IsDeclarationFile {
		t.Fatal("library fixture has no declaration")
	}
	compare(ask(promise, "first-declaration-file"), []string{"1", "first-declaration-file", "1", "1", "1"})
	compare(ask(declaration, "first-declaration-file"), []string{"1", "first-declaration-file", "1", "1", "0"})
	compare(ask(nodeAt("missing;"), "first-declaration-file"), []string{"1", "first-declaration-file", "0"})
	merged := nodeAt("M;")
	ms := c.GetSymbolAtLocation(merged)
	want := []string{"1", "local-binding-declarations", "1"}
	var local []*ast.Node
	for _, d := range ms.Declarations {
		if ast.GetSourceFileOfNode(d) == sf {
			local = append(local, d)
		}
	}
	if len(ms.Declarations) != 3 || len(local) != 2 {
		t.Fatal("fixture must have local and foreign merged declarations")
	}
	want = append(want, strconv.Itoa(len(local)))
	for _, d := range local {
		want = append(want, strings.TrimPrefix(d.Kind.String(), "Kind"), strconv.Itoa(d.Pos()), strconv.Itoa(d.End()))
	}
	compare(ask(merged, "local-binding-declarations"), want)
	localDetails := []string{"1", "local-symbol-details", "1", strconv.FormatUint(p.symbolID(ms), 10), strconv.FormatUint(uint64(ms.Flags), 10), ms.Name, strconv.Itoa(len(ms.Declarations)), strconv.Itoa(len(local))}
	for _, declaration := range local {
		localDetails = append(localDetails, strings.TrimPrefix(declaration.Kind.String(), "Kind"), strconv.Itoa(declaration.Pos()), strconv.Itoa(declaration.End()))
	}
	compare(ask(merged, "local-symbol-details"), localDetails)
	flags := []string{"1", "output-symbol", "1", strconv.Itoa(len(ms.Declarations))}
	for _, declaration := range ms.Declarations {
		if ast.GetSourceFileOfNode(declaration).IsDeclarationFile {
			flags = append(flags, "1")
		} else {
			flags = append(flags, "0")
		}
	}
	compare(ask(merged, "output-symbol\nfiles"), flags)
	compare(ask(nodeAt("onlyOther;"), "local-binding-declarations"), []string{"1", "local-binding-declarations", "1", "0"})
	// The row's requested files never shrink the tsconfig's real program.
	if p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(filepath.Join(directory, "extra.ts")))) == nil {
		t.Fatal("row roots dropped a configured source file")
	}
}
