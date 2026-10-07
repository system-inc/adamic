package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestTypeLeafFacts(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "input.a")
	config := filepath.Join(dir, "tsconfig.json")
	for path, text := range map[string]string{config: `{"files":["input.a"],"compilerOptions":{"strict":true,"target":"ESNext","lib":["ESNext"]}}`, file: `declare const items:readonly [number,string];declare const key:'世界';declare const callback:(x:number)=>string;declare const map:Map<string,number>;items;key;callback;map;export {};`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	source := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	var nodes []*ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier && n.Parent != nil && n.Parent.Kind == ast.KindExpressionStatement {
			nodes = append(nodes, n)
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	if len(nodes) != 4 {
		t.Fatal("missing expression references")
	}
	for _, n := range nodes {
		subject := c.GetTypeAtLocation(n)
		g := graph{program: p, checker: c}
		id := g.id(subject)
		wire, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "type-leaf-facts\n"+strconv.FormatUint(id, 10)+"\nlength")
		if err != nil {
			t.Fatal(err)
		}
		if !utf8.ValidString(wire) {
			t.Fatal("type-leaf-facts returned invalid UTF-8")
		}
		got := decodedFields(t, wire)
		wantArray := "0"
		if checker.Checker_isArrayOrTupleType(c, subject) {
			wantArray = "1"
		}
		if got[2] != strconv.FormatUint(uint64(subject.Flags()), 10) || got[3] != wantArray {
			t.Fatalf("flags or array: %q", got)
		}
		if got[4] != strconv.Itoa(len(c.GetSignaturesOfType(subject, checker.SignatureKindCall))) {
			t.Fatalf("signature count: %q", got)
		}
		if n.Text() == "key" && (got[5] != "1" || got[6] != subject.AsLiteralType().Value().(string)) {
			t.Fatalf("literal: %q", got)
		}
		at := 7
		if subject.Symbol() == nil {
			if got[at] != "0" {
				t.Fatalf("unexpected symbol: %q", got)
			}
			at++
		} else {
			if got[at] != "1" || got[at+1] != strings.ToValidUTF8(subject.Symbol().Name, "\ufffd") || got[at+2] != strconv.Itoa(len(subject.Symbol().Declarations)) {
				t.Fatalf("symbol origin: %q", got)
			}
			at += 3
			for _, declaration := range subject.Symbol().Declarations {
				f := ast.GetSourceFileOfNode(declaration)
				declFile := "0"
				library := "0"
				if f.IsDeclarationFile {
					declFile = "1"
				}
				if p.Compiler.IsSourceFileDefaultLibrary(f.Path()) {
					library = "1"
				}
				if got[at] != f.FileName() || got[at+1] != declFile || got[at+2] != library {
					t.Fatalf("declaration origin: %q", got)
				}
				at += 3
			}
		}
		if at != len(got)-2 {
			t.Fatal("unconsumed leaf facts")
		}
		wantProperty := "0"
		if checker.Checker_getPropertyOfType(c, subject, "length") != nil {
			wantProperty = "1"
		}
		if got[len(got)-2] != "1" || got[len(got)-1] != wantProperty {
			t.Fatalf("property: %q", got)
		}
		for _, q := range []string{"type-leaf-facts", "type-leaf-facts\n0", "type-leaf-facts\n999999999", "type-leaf-facts\n0" + strconv.FormatUint(id, 10)} {
			if _, err := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", q); err == nil {
				t.Fatalf("accepted invalid identity: %s", q)
			}
		}
	}
}
