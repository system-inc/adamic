package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Test the independent registration modules without changing shared dispatch.
func TestWave16QuestionFacts(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.a")
	source := `interface MockTracker {method(x:number):void} declare const tracker:MockTracker; tracker.method(1);
 declare const count:number; count; 0n; document;
 document.addEventListener('click',e=>{(e.target as HTMLInputElement).value; if(e.target instanceof HTMLInputElement){(e.target as HTMLInputElement).value;}});
 export {};
 `
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","lib":["ES2022","DOM"]},"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(file)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	selectNodes := func(kind ast.Kind, text string) []*ast.Node {
		var matches []*ast.Node
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n.Kind == kind && source[scanner.GetTokenPosOfNode(n, sf, false):n.End()] == text {
				matches = append(matches, n)
			}
			n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
		visit(sf.AsNode())
		if len(matches) == 0 {
			t.Fatalf("missing %s %s", kind, text)
		}
		return matches
	}
	ask := func(n *ast.Node, q string) []string {
		out := &fields{}
		out.number(1)
		out.text(q)
		wire, handled, err := p.additionalAnswer(out, c, n, q)
		if !handled || err != nil {
			t.Fatalf("%s: %t %v", q, handled, err)
		}
		return decodedFields(t, wire)
	}
	if got := ask(sf.AsNode(), "emit-module-kind"); got[2] != "99" {
		t.Fatalf("emit module: %q", got)
	}
	call := selectNodes(ast.KindCallExpression, "tracker.method(1)")[0]
	ancestry := ask(call, "signature-ancestry")
	declaration := c.GetResolvedSignature(call).Declaration()
	var expected []string
	for current := declaration; current != nil; current = current.Parent {
		kind := strings.TrimPrefix(current.Kind.String(), "Kind")
		name, nameKind := "", ""
		if current.Kind != ast.KindSourceFile && current.Name() != nil {
			name = current.Name().Text()
			nameKind = strings.TrimPrefix(current.Name().Kind.String(), "Kind")
		}
		expected = append(expected, kind, name, nameKind)
	}
	if ancestry[2] != "1" || ancestry[3] != strconv.Itoa(len(expected)/3) || strings.Join(ancestry[4:], "|") != strings.Join(expected, "|") {
		t.Fatalf("signature ancestry: %q", ancestry)
	}
	doc := selectNodes(ast.KindIdentifier, "document")[0]
	origins := ask(doc, "symbol-origins")
	symbol := c.GetSymbolAtLocation(doc)
	if origins[2] != strconv.FormatUint(p.symbolID(symbol), 10) || origins[3] != strconv.FormatUint(uint64(symbol.Flags), 10) || origins[4] != "document" || origins[5] != strconv.Itoa(len(symbol.Declarations)) {
		t.Fatalf("symbol identity: %q", origins)
	}
	for i, d := range symbol.Declarations {
		if origins[6+i*2] != strings.TrimPrefix(d.Kind.String(), "Kind") || origins[7+i*2] != "1" {
			t.Fatalf("default library origin: %q", origins)
		}
	}
	for i, assertion := range selectNodes(ast.KindAsExpression, "e.target as HTMLInputElement") {
		facts := ask(assertion, "listener-type-facts")
		known := c.GetTypeAtLocation(assertion.Expression())
		target := c.GetTypeFromTypeNode(assertion.Type())
		want := "0"
		if c.IsTypeAssignableTo(c.GetNonNullableType(known), c.GetNonNullableType(target)) {
			want = "1"
		}
		if facts[2] != want || (i == 0 && want != "0") || (i == 1 && want != "1") {
			t.Fatalf("assignability: %q", facts)
		}
		id, err := strconv.ParseUint(facts[3], 10, 64)
		if err != nil || id == 0 || p.typesByID[id-1] != target {
			t.Fatalf("asserted type identity: %q", facts)
		}
		if !strings.Contains(strings.Join(facts, "|"), "|HTMLInputElement|") {
			t.Fatalf("missing asserted symbol: %q", facts)
		}
	}
	count := selectNodes(ast.KindIdentifier, "count")[1]
	number := ask(count, "render-type-facts")
	if number[3] != "1" || number[5] != strconv.FormatUint(uint64(checker.TypeFlagsNumber), 10) || number[6] != "0" || number[7] != "" || number[8] != "0" || number[9] != "0" {
		t.Fatalf("number graph: %q", number)
	}
	literal := ask(selectNodes(ast.KindBigIntLiteral, "0n")[0], "render-type-facts")
	if literal[6] != "1" || literal[7] != "0" {
		t.Fatalf("bigint literal value: %q", literal)
	}
	for question := range additionalQuestions {
		out := &fields{}
		if _, handled, err := p.additionalAnswer(out, c, count, question+"\nextra"); !handled || err == nil {
			t.Fatalf("suffix accepted: %s", question)
		}
	}
	for _, question := range []string{"signature-ancestry", "emit-module-kind", "listener-type-facts"} {
		out := &fields{}
		if _, _, err := p.additionalAnswer(out, c, count, question); err == nil {
			t.Fatalf("wrong kind accepted: %s", question)
		}
	}
	t.Log("five independent fact modules match checker identities, declaration ancestry, origins, narrowing and bigint values; suffix and node-kind refusals pass")
}
