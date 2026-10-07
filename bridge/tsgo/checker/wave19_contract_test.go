package checker

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestWave19ContractFacts(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "input.a")
	config := filepath.Join(directory, "tsconfig.json")
	text := `function identity<T>(value:T):T{return value} const f:()=>Promise<number>=async()=>1; const item={run:async()=>1}; const result=identity(item); interface I{run():Promise<number>} class C implements I{async run(){return 1;}} export {};`
	if err := os.WriteFile(source, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["input.a"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{source})
	if err != nil {
		t.Fatal(err)
	}
	file := p.Compiler.GetSourceFile(source)
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), file)
	defer release()
	var arrow, call, method *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindArrowFunction && checker.Checker_getContextualType(c, node, checker.ContextFlagsNone) != nil {
			arrow = node
		}
		if node.Kind == ast.KindCallExpression {
			call = node
		}
		if node.Kind == ast.KindMethodDeclaration && node.Parent.Kind == ast.KindClassDeclaration {
			method = node
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	if arrow == nil || call == nil || method == nil {
		t.Fatal("missing typed probes")
	}
	frame := func(mode string) *fields { out := &fields{}; out.number(1); out.text(mode); return out }
	identity := func(subject *checker.Type) string {
		return strconv.FormatUint((&graph{program: p, checker: c}).id(subject), 10)
	}
	contextual := checker.Checker_getContextualType(c, arrow, checker.ContextFlagsNone)
	signature := c.GetSignaturesOfType(contextual, checker.SignatureKindCall)[0]
	expectedReturn := identity(c.GetReturnTypeOfSignature(signature))
	wire, err := p.wave19TypeSignatures(frame("wave19-type-signatures"), c, arrow, "wave19-type-signatures\n"+identity(contextual))
	if err != nil {
		t.Fatal(err)
	}
	got := decodedFields(t, wire)
	if got[2] != "1" || got[3] != expectedReturn || got[4] != "0" || got[5] != "0" || got[6] != "0" {
		t.Fatalf("wrong signature facts %q", got)
	}
	wire, err = p.wave19GenericCall(frame("wave19-generic-call"), c, call, "wave19-generic-call")
	if err != nil {
		t.Fatal(err)
	}
	got = decodedFields(t, wire)
	resolved := c.GetResolvedSignature(call)
	declared := resolved.Target()
	if got[2] != "0" || got[3] != strconv.Itoa(len(declared.TypeParameters())) || got[4] != identity(declared.TypeParameters()[0]) || got[5] != "1" || got[6] != identity(checker.Checker_getTypeOfSymbol(c, declared.Parameters()[0])) {
		t.Fatalf("wrong generic facts %q", got)
	}
	object := checker.Checker_getTypeOfSymbol(c, declared.Parameters()[0])
	_ = object
	argument := call.Arguments()[0]
	subject := c.GetTypeAtLocation(argument)
	property := checker.Checker_getPropertyOfType(c, subject, "run")
	expectedProperty := identity(checker.Checker_getTypeOfSymbol(c, property))
	wire, err = p.wave19TypeMembers(frame("wave19-type-members"), c, call, "wave19-type-members\n"+identity(subject)+"\nrun")
	if err != nil {
		t.Fatal(err)
	}
	got = decodedFields(t, wire)
	if got[2] != expectedProperty || got[3] != "0" || got[4] != "0" || got[5] != "0" || got[6] != "0" {
		t.Fatalf("wrong member facts %q", got)
	}
	wire, err = p.wave19HeritageMembers(frame("wave19-heritage-members"), c, method, "wave19-heritage-members")
	if err != nil {
		t.Fatal(err)
	}
	got = decodedFields(t, wire)
	heritage := method.Parent.AsClassDeclaration().HeritageClauses.Nodes[0].AsHeritageClause().Types.Nodes[0]
	base := c.GetTypeAtLocation(heritage)
	member := checker.Checker_getPropertyOfType(c, base, "run")
	expectedMember := identity(c.GetTypeOfSymbolAtLocation(member, method))
	if got[4] != "1" || got[5] != expectedMember {
		t.Fatalf("wrong heritage facts %q", got)
	}
	for _, question := range []string{"wave19-type-signatures\n0", "wave19-type-signatures\n01", "wave19-type-signatures\n999999"} {
		if _, err := p.wave19TypeSignatures(frame("wave19-type-signatures"), c, arrow, question); err == nil {
			t.Fatal("invalid identity accepted")
		}
	}
	if _, err := p.wave19GenericCall(frame("wave19-generic-call"), c, arrow, "wave19-generic-call"); err == nil {
		t.Fatal("function accepted as call")
	}
	if _, err := p.wave19HeritageMembers(frame("wave19-heritage-members"), c, arrow, "wave19-heritage-members"); err == nil {
		t.Fatal("arrow accepted as heritage member")
	}
}
