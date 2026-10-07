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

func TestWave12RawCheckerQuestions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "input.a")
	config := filepath.Join(dir, "tsconfig.json")
	text := "var Object=1;function f(){var x=1;{let y=2;}}\n"
	if err := os.WriteFile(file, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"target":"ES2022","lib":["ES2022"]},"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var scopes []*ast.Node
	var identifier *ast.Node
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if ast.IsLocalsContainer(n) {
			scopes = append(scopes, n)
		}
		if n.Kind == ast.KindIdentifier && n.Text() == "Object" {
			identifier = n
		}
		n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	if len(scopes) < 4 || identifier == nil {
		t.Fatal("raw question controls absent")
	}
	wire, err := p.Inspect(file, 0, uint64(len(text)), "SourceFile", "scope-metadata")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if fields[0] != "1" || fields[1] != "scope-metadata" || fields[2] != "0" || fields[3] != strconv.Itoa(len(scopes)) {
		t.Fatalf("scope header %q", fields)
	}
	for i, n := range scopes {
		at := 4 + 3*i
		if fields[at] != strings.TrimPrefix(n.Kind.String(), "Kind") || fields[at+1] != strconv.Itoa(n.Pos()) || fields[at+2] != strconv.Itoa(n.End()) {
			t.Fatalf("scope %d differs: %q", i, fields)
		}
	}
	wire, err = p.Inspect(file, uint64(identifier.Pos()), uint64(identifier.End()), "Identifier", "global-symbol-details")
	if err != nil {
		t.Fatal(err)
	}
	fields = decodedFields(t, wire)
	symbol := c.GetGlobalSymbol("Object", ast.SymbolFlagsAll, nil)
	if symbol == nil || len(symbol.Declarations) == 0 || fields[2] != "1" || fields[5] != symbol.Name || fields[6] != strconv.Itoa(len(symbol.Declarations)) {
		t.Fatalf("global header: %q", fields)
	}
	at := 7
	for _, d := range symbol.Declarations {
		source := ast.GetSourceFileOfNode(d)
		if fields[at] != source.FileName() || fields[at+1] != strings.TrimPrefix(d.Kind.String(), "Kind") || fields[at+2] != strconv.Itoa(d.Pos()) || fields[at+3] != strconv.Itoa(d.End()) {
			t.Fatalf("global declaration: %q", fields[at:])
		}
		tags, _ := strconv.Atoi(fields[at+10])
		at += 11 + 5*tags
		parameters, _ := strconv.Atoi(fields[at])
		at += 1 + parameters
	}
	if at != len(fields) {
		t.Fatal("unexpected global trailing fields")
	}
	for _, q := range []string{"scope-metadata", "scope-metadata\nextra", "global-symbol-details\nextra"} {
		if _, err := p.Inspect(file, uint64(identifier.Pos()), uint64(identifier.End()), "Identifier", q); err == nil {
			t.Fatal("malformed question accepted", q)
		}
	}
	if _, err := p.Inspect(file, 0, uint64(len(text)), "SourceFile", "global-symbol-details"); err == nil {
		t.Fatal("global details accepted nonidentifier")
	}
}
