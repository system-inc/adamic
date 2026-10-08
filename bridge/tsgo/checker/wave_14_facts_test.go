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

func TestWave14CheckerQuestions(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	source := "class Token {} function register(x:new()=>unknown){} register(Token); declare const obj:{}; String(obj); declare const rows:{}[]; rows.join();"
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"files":["input.a"],"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	var token, obj, rows *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			switch node.Text() {
			case "Token":
				token = node
			case "obj":
				obj = node
			case "rows":
				rows = node
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(sf.AsNode())
	if token == nil || obj == nil || rows == nil {
		t.Fatal("missing positive controls")
	}
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	context := ask(token, "reference-context")
	if context[2] != strconv.FormatUint(p.symbolID(c.GetSymbolAtLocation(token)), 10) || context[3] != "0" || context[4] == context[5] || context[6] != "1" || context[7] != "1" {
		t.Fatalf("contextual constructor: %q", context)
	}
	for _, node := range []*ast.Node{obj, rows} {
		shape := ask(node, "raw-shape")
		id := shape[5]
		metadata := ask(node, "stringification-type\n"+id)
		subject := p.typesByID[mustID(t, id)-1]
		if metadata[2] != strconv.FormatUint(uint64(subject.Flags()), 10) || metadata[3] != c.TypeToString(subject) {
			t.Fatalf("type metadata: %q", metadata)
		}
		if metadata[6] != strconv.Itoa(boolInt(checker.Checker_isArrayType(c, subject))) {
			t.Fatalf("array metadata: %q", metadata)
		}
		element := checker.Checker_getIndexTypeOfType(c, subject, checker.Checker_numberType(c))
		parsed := mustIDOrZero(t, metadata[8])
		if (parsed == 0) != (element == nil) || parsed != 0 && p.typesByID[parsed-1] != element {
			t.Fatal("number index differs")
		}
	}
	for _, question := range []string{"stringification-type", "stringification-type\n0", "stringification-type\n01", "stringification-type\n999999", "reference-context\nsuffix"} {
		if _, err := p.Inspect(file, uint64(obj.Pos()), uint64(obj.End()), "Identifier", question); err == nil {
			t.Fatal("invalid question accepted", question)
		}
	}
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
func mustIDOrZero(t *testing.T, text string) uint64 {
	t.Helper()
	id, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
