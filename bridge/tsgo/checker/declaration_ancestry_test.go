package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDeclarationAncestryAndAwaitedShape(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	for path, text := range map[string]string{
		config:                                   `{"compilerOptions":{"strict":true,"lib":["ES2022"]},"files":["prelude.d.ts"]}`,
		filepath.Join(directory, "prelude.d.ts"): "",
		file:                                     "type Result<T> = (({ok:T}) | {error:string}); declare const r:Result<number>;r;declare const p:Promise<Result<number>>;p;",
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
	var result, promise *ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier && n.Text() == "r" {
			result = n
		}
		if n.Kind == ast.KindIdentifier && n.Text() == "p" {
			promise = n
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(sf.AsNode())
	ask := func(n *ast.Node, q string) (string, error) {
		return p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", q)
	}
	raw, err := ask(result, "raw-shape")
	if err != nil {
		t.Fatal(err)
	}
	_ = raw
	direct := c.GetTypeAtLocation(result)
	if direct.Flags()&checker.TypeFlagsUnion == 0 {
		t.Fatal("control is not a union")
	}
	for _, part := range direct.Types() {
		id := p.typeIDs[part]
		wire, err := ask(result, "declaration-ancestry\n"+strconv.FormatUint(id, 10))
		if err != nil {
			t.Fatal(err)
		}
		fields := decodedFields(t, wire)
		at := 2
		count, _ := strconv.Atoi(fields[at])
		at++
		symbol := part.Symbol()
		if symbol == nil || count != len(symbol.Declarations) {
			t.Fatal("wrong declaration count")
		}
		for _, declaration := range symbol.Declarations {
			length, _ := strconv.Atoi(fields[at])
			at++
			n := declaration
			for ancestor := 0; ancestor < length; ancestor++ {
				if n == nil {
					t.Fatal("too many ancestors")
				}
				children := 0
				n.ForEachChild(func(child *ast.Node) bool { children++; return false })
				name := ""
				if named := n.Name(); named != nil {
					name = named.Text()
				}
				source := ast.GetSourceFileOfNode(n)
				first, last := 0, 0
				if n.Kind == ast.KindTypeAliasDeclaration {
					first = n.AsTypeAliasDeclaration().Type.Pos()
					last = n.AsTypeAliasDeclaration().Type.End()
				}
				expected := []string{strings.TrimPrefix(n.Kind.String(), "Kind"), name, strconv.Itoa(n.Pos()), strconv.Itoa(n.End()), string(source.FileName()), strconv.Itoa(children), strconv.Itoa(first), strconv.Itoa(last)}
				if strings.Join(fields[at:at+8], "|") != strings.Join(expected, "|") {
					t.Fatalf("ancestry mismatch: %q != %q", fields[at:at+8], expected)
				}
				at += 8
				n = n.Parent
			}
			if n != nil {
				t.Fatal("truncated ancestry")
			}
		}
		if at != len(fields) {
			t.Fatal("trailing fields")
		}
	}
	wire, err := ask(promise, "awaited-shape")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	if fields[3] != "1" || fields[4] != "1" {
		t.Fatalf("missing awaited root: %q", fields)
	}
	awaited := checker.Checker_getAwaitedType(c, c.GetTypeAtLocation(promise))
	if p.typesByID[mustID(t, fields[5])-1] != awaited {
		t.Fatal("awaited identity differs from direct checker")
	}
	for _, question := range []string{"declaration-ancestry", "declaration-ancestry\n0", "declaration-ancestry\n01", "declaration-ancestry\n999999", "declaration-ancestry\n1\nextra", "awaited-shape\nextra"} {
		if _, err := ask(result, question); err == nil {
			t.Fatal("accepted malformed question", question)
		}
	}
	t.Log("Raw ancestry, instantiated literal declarations, alias written type, awaited identity and malformed-question refusals match direct checker")
}
