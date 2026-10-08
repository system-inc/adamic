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

func TestLibraryProvenanceMatchesChecker(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "file.ts")
	config := filepath.Join(directory, "tsconfig.json")
	source := "new Date().toISOString();\n"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ES2022"}}`} {
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
	var selected *ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier && n.Text() == "toISOString" {
			selected = n
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	if selected == nil {
		t.Fatal("no library member")
	}
	symbol := c.GetSymbolAtLocation(selected)
	if symbol == nil {
		t.Fatal("no library symbol")
	}
	question := fmt.Sprintf("library-member-provenance\n%d\n%d\nIdentifier", selected.Pos(), selected.End())
	wire, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", question)
	if err != nil {
		t.Fatal(err)
	}
	got := decodedFields(t, wire)
	want := []string{"1", "library-member-provenance", "1", strconv.Itoa(len(symbol.Declarations))}
	for _, d := range symbol.Declarations {
		lib := "0"
		if p.Compiler.IsSourceFileDefaultLibrary(ast.GetSourceFileOfNode(d).PathKey()) {
			lib = "1"
		}
		want = append(want, lib, strings.TrimPrefix(d.Kind.String(), "Kind"), d.Name().Text(), strings.TrimPrefix(d.Parent.Kind.String(), "Kind"), d.Parent.Name().Text())
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("library provenance: %q want %q", got, want)
	}
	defaultQuestion := fmt.Sprintf("symbol-default-library\n%d\n%d\nIdentifier", selected.Pos(), selected.End())
	defaultWire, err := p.Inspect(file, 0, uint64(len(source)), "SourceFile", defaultQuestion)
	if err != nil {
		t.Fatal(err)
	}
	library := false
	for _, declaration := range symbol.Declarations {
		library = library || p.Compiler.IsSourceFileDefaultLibrary(ast.GetSourceFileOfNode(declaration).PathKey())
	}
	if !library {
		t.Fatal("fixture lacks a default-library declaration")
	}
	if got := decodedFields(t, defaultWire); fmt.Sprint(got) != fmt.Sprint([]string{"1", "symbol-default-library", "1"}) {
		t.Fatalf("default-library symbol fact: %q", got)
	}
	if _, err := p.Inspect(file, uint64(selected.Pos()), uint64(selected.End()), "Identifier", question); err == nil {
		t.Fatal("library file question accepted a node selector")
	}
}
