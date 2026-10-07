package checker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestWave16TextQuestions(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := "declare const numeric:`${number}-${number}`,textual:`id-${string}`,literal:\"O'Brien\",truth:true; const command='echo '+literal; command; function f<T extends 'x'|'y'>(value:T){value;} declare const lone:'\\uD800'; lone; declare const loneTemplate:`${number}\\uD800`; loneTemplate; export {};\n"
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	nodes := map[string]*ast.Node{}
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n.Kind == ast.KindIdentifier {
			nodes[n.Text()] = n
		}
		n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	ask := func(node *ast.Node, q string) ([]string, error) {
		out := &fields{}
		out.number(1)
		out.text(q)
		wire, handled, err := p.additionalAnswer(out, c, node, q)
		if !handled {
			t.Fatal("unregistered question")
		}
		if err != nil {
			return nil, err
		}
		return decodedFields(t, wire), nil
	}
	for _, name := range []string{"numeric", "textual", "literal", "truth", "command", "value"} {
		node := nodes[name]
		if node == nil {
			t.Fatal("missing identifier")
		}
		fields, err := ask(node, "text-type-facts")
		if err != nil {
			t.Fatal(err)
		}
		at := 2
		natural := func() uint64 {
			if at >= len(fields) {
				t.Fatal("truncated type facts")
			}
			value, err := strconv.ParseUint(fields[at], 10, 64)
			at++
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		field := func() string {
			if at >= len(fields) {
				t.Fatal("truncated type facts")
			}
			s := fields[at]
			at++
			return s
		}
		rootID := natural()
		constrainedID := natural()
		root := c.GetTypeAtLocation(node)
		constraint := checker.Checker_getBaseConstraintOfType(c, root)
		if constraint == nil {
			constraint = root
		}
		if rootID == 0 || p.typesByID[rootID-1] != root || constrainedID == 0 || p.typesByID[constrainedID-1] != constraint {
			t.Fatal("text root identity differs")
		}
		count := natural()
		for i := uint64(0); i < count; i++ {
			id := natural()
			if id == 0 || int(id) > len(p.typesByID) {
				t.Fatal("unknown type identity")
			}
			subject := p.typesByID[id-1]
			flags := natural()
			errorFlag := field()
			kind := field()
			literal := field()
			boundID := natural()
			partCount := natural()
			if flags != uint64(subject.Flags()) {
				t.Fatal("text type flags differ")
			}
			errorWant := "0"
			if subject.Flags()&checker.TypeFlagsIntrinsic != 0 && subject.AsIntrinsicType().IntrinsicName() == "error" {
				errorWant = "1"
			}
			if errorFlag != errorWant {
				t.Fatal("error type flag differs")
			}
			literalWant, kindWant := "", ""
			if subject.Flags()&checker.TypeFlagsLiteral != 0 {
				switch value := subject.AsLiteralType().Value().(type) {
				case string:
					kindWant = "string"
					literalWant = value
				case bool:
					kindWant = "boolean"
					literalWant = strconv.FormatBool(value)
				case fmt.Stringer:
					kindWant = "number"
					literalWant = value.String()
				}
			}
			if kind != kindWant || literal != literalWant {
				t.Fatal("literal value differs")
			}
			bound := checker.Checker_getBaseConstraintOfType(c, subject)
			if (bound == nil && boundID != 0) || (bound != nil && (boundID == 0 || p.typesByID[boundID-1] != bound)) {
				t.Fatal("constraint identity differs")
			}
			var parts []*checker.Type
			if subject.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
				parts = subject.Types()
			}
			if int(partCount) != len(parts) {
				t.Fatal("constituent count differs")
			}
			for _, part := range parts {
				id := natural()
				if id == 0 || p.typesByID[id-1] != part {
					t.Fatal("constituent identity differs")
				}
			}
			var texts []string
			var holes []*checker.Type
			if subject.Flags()&checker.TypeFlagsTemplateLiteral != 0 {
				texts = subject.AsTemplateLiteralType().Texts()
				holes = subject.AsTemplateLiteralType().Types()
			}
			if int(natural()) != len(texts) {
				t.Fatal("template text count differs")
			}
			for _, text := range texts {
				if field() != text {
					t.Fatal("template text differs")
				}
			}
			if int(natural()) != len(holes) {
				t.Fatal("template hole count differs")
			}
			for _, hole := range holes {
				id := natural()
				if id == 0 || p.typesByID[id-1] != hole {
					t.Fatal("template hole identity differs")
				}
			}
		}
		if at != len(fields) {
			t.Fatal("trailing type facts")
		}
	}
	node := nodes["command"]
	value, err := ask(node, "value-declaration")
	if err != nil {
		t.Fatal(err)
	}
	symbol := c.GetSymbolAtLocation(node)
	decl := symbol.ValueDeclaration
	want := []string{"1", "value-declaration", "1", strconv.FormatUint(uint64(symbol.Flags), 10), "1", file, strings.TrimPrefix(decl.Kind.String(), "Kind"), strconv.Itoa(decl.Pos()), strconv.Itoa(decl.End())}
	if strings.Join(value, "|") != strings.Join(want, "|") {
		t.Fatal("value declaration differs")
	}
	if _, err := ask(sf.AsNode(), "value-declaration"); err == nil {
		t.Fatal("value declaration accepted wrong node kind")
	}
	if _, err := ask(nodes["lone"], "text-type-facts"); err == nil || !strings.Contains(err.Error(), "non-UTF-8") {
		t.Fatalf("non-UTF-8 literal accepted: %v", err)
	}
	if _, err := ask(nodes["loneTemplate"], "text-type-facts"); err == nil || !strings.Contains(err.Error(), "non-UTF-8") {
		t.Fatalf("non-UTF-8 template accepted: %v", err)
	}
	t.Log("raw template holes/texts, literal values, constraints and value declaration match direct checker; malformed node and non-UTF-8 refusals pass")
}
