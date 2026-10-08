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
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

func TestCoverageCheckerQuestions(t *testing.T) {
	directory := t.TempDir()
	config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input.ts")
	source := `interface Base { base: number }
/** @processState shared counters */
interface State extends Base { readonly fixed: number; extra?: string; method(x: number): boolean }
const state: State = {fixed:1,base:0,method(x){return x>0}};
function call(x: State): State { return x; }
const literal = "世界🌍" as const;
const truth = true as const;
type Alias = Map<string, State>;
declare const array: Alias;
array;
call(state);
state;
literal;
truth;
export {};
`
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), sf)
	defer release()
	selectNode := func(kind ast.Kind, text string) *ast.Node {
		var selected *ast.Node
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			start := scanner.GetTokenPosOfNode(node, sf, false)
			if node.Kind == kind && source[start:node.End()] == text {
				selected = node
			}
			node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(sf.AsNode())
		if selected == nil {
			t.Fatalf("missing %s %q", kind, text)
		}
		return selected
	}
	ask := func(node *ast.Node, question string) []string {
		t.Helper()
		wire, err := p.Inspect(file, uint64(node.Pos()), uint64(node.End()), strings.TrimPrefix(node.Kind.String(), "Kind"), question)
		if err != nil {
			t.Fatalf("%s: %v", question, err)
		}
		return decodedFields(t, wire)
	}
	rootID := func(node *ast.Node, question string) string {
		t.Helper()
		result := ask(node, question)
		if result[3] != "1" || result[4] != "1" {
			t.Fatalf("missing root: %q", result)
		}
		return result[5]
	}
	state := selectNode(ast.KindIdentifier, "state")
	stateID := rootID(state, "raw-shape")
	if id := rootID(state, "symbol-shape"); id != stateID {
		t.Fatalf("symbol type %s != location %s", id, stateID)
	}
	binding := ask(state, "binding-declarations")
	if binding[2] != "1" || binding[4] != "1" || binding[5] != file || binding[6] != "VariableDeclaration" {
		t.Fatalf("binding: %q", binding)
	}
	if alias := ask(state, "alias-declarations"); strings.Join(alias[2:], "|") != strings.Join(binding[2:], "|") {
		t.Fatalf("non-alias changed declarations: %q", alias)
	}
	identities := ask(state, "symbol-identities")
	own, _ := strconv.ParseUint(identities[2], 10, 64)
	if own == 0 || p.symbolsByID[own-1] != c.GetSymbolAtLocation(state) || identities[3] != "0" || identities[4] != "0" {
		t.Fatalf("symbol identities: %q", identities)
	}
	annotation := selectNode(ast.KindTypeReference, "State")
	if _, err := p.Inspect(file, uint64(state.Pos()), uint64(state.End()), "Identifier", "annotation-shape"); err == nil {
		t.Fatal("annotation-shape accepted a value node")
	}
	if rootID(annotation, "annotation-shape") != stateID {
		t.Fatal("annotation type differs")
	}
	comparison := ask(state, "identical-types\n"+stateID+"\n"+stateID)
	if comparison[2] != "1" {
		t.Fatalf("identical type refused: %q", comparison)
	}
	numberID := rootID(selectNode(ast.KindNumericLiteral, "1"), "raw-shape")
	if ask(state, "identical-types\n"+stateID+"\n"+numberID)[2] != "0" {
		t.Fatal("different types identical")
	}
	metadata := ask(state, "type-metadata\n"+stateID)
	if len(metadata) != 7 || metadata[2] != strconv.FormatUint(uint64(p.typesByID[mustID(t, stateID)-1].ObjectFlags()), 10) || metadata[6] != "0" {
		t.Fatalf("object metadata: %q", metadata)
	}
	properties := ask(state, "type-properties\n"+stateID)
	count, _ := strconv.Atoi(properties[2])
	if count != 4 {
		t.Fatalf("property count: %q", properties)
	}
	fixed, extra := false, false
	for at := 0; at < count; at++ {
		first := 3 + at*4
		name, key, flags, readonly := properties[first], properties[first+1], properties[first+2], properties[first+3]
		if ask(state, "property-exists\n"+stateID+"\n"+key)[2] != "1" {
			t.Fatalf("missing property %s", name)
		}
		symbol := p.symbolsByID[mustID(t, key)-1]
		if symbol != checker.Checker_getPropertyOfType(c, p.typesByID[mustID(t, stateID)-1], name) {
			t.Fatal("wrong property identity")
		}
		if name == "fixed" {
			fixed = readonly == "1"
		}
		if name == "extra" {
			parsed, _ := strconv.ParseUint(flags, 10, 64)
			extra = parsed&uint64(ast.SymbolFlagsOptional) != 0
		}
	}
	if !fixed || !extra {
		t.Fatalf("lost readonly or optional property: %q", properties)
	}
	literal := selectNode(ast.KindIdentifier, "literal")
	literalID := rootID(literal, "raw-shape")
	value := ask(literal, "literal-value\n"+literalID)
	if len(value) != 4 || value[2] != "string" || value[3] != "世界🌍" {
		t.Fatalf("literal: %q", value)
	}
	truth := selectNode(ast.KindIdentifier, "truth")
	value = ask(truth, "literal-value\n"+rootID(truth, "raw-shape"))
	if value[2] != "boolean" || value[3] != "true" {
		t.Fatalf("boolean literal: %q", value)
	}
	resolved := ask(state, "resolved-name\nBoolean")
	if resolved[2] != "1" || resolved[3] != "Boolean" {
		t.Fatalf("name resolution: %q", resolved)
	}
	call := selectNode(ast.KindCallExpression, "call(state)")
	if rootID(call, "contextual-argument\n0") != stateID {
		t.Fatal("wrong contextual argument type")
	}
	fn := selectNode(ast.KindFunctionDeclaration, "function call(x: State): State { return x; }")
	if rootID(fn, "annotated-return-shape") != stateID {
		t.Fatal("wrong annotated return type")
	}
	fnID := rootID(selectNode(ast.KindIdentifier, "call"), "raw-shape")
	signatures := ask(state, "function-signatures\n"+fnID)
	if signatures[2] != "1" || signatures[3] != "1" || signatures[4] != "x" || signatures[5] != "0" {
		t.Fatalf("signature: %q", signatures)
	}
	declared := selectNode(ast.KindInterfaceDeclaration, "interface State extends Base { readonly fixed: number; extra?: string; method(x: number): boolean }")
	details := ask(declared, "declaration-details")
	if details[2] != file || details[3] != "InterfaceDeclaration" || details[12] != "1" || details[14] != "processState" || !strings.Contains(details[17], "shared counters") {
		t.Fatalf("declaration and documentation: %q", details)
	}
	typeDetails := ask(state, "type-symbol-details\n"+stateID)
	if typeDetails[2] != "1" || typeDetails[5] != "State" || !strings.Contains(strings.Join(typeDetails, "|"), "processState") {
		t.Fatalf("type declaration: %q", typeDetails)
	}
	nodeDetails := ask(state, "node-symbol-details")
	if nodeDetails[2] != "1" || nodeDetails[5] != "state" {
		t.Fatalf("node symbol: %q", nodeDetails)
	}
	propertyDetails := ask(state, "property-declarations\n"+stateID+"\nmethod")
	if propertyDetails[2] != "1" || propertyDetails[5] != "method" || !strings.Contains(strings.Join(propertyDetails, "|"), "MethodSignature") {
		t.Fatalf("property declaration: %q", propertyDetails)
	}
	bases := ask(declared, "container-bases")
	if bases[4] != "1" {
		t.Fatalf("missing base: %q", bases)
	}
	array := selectNode(ast.KindIdentifier, "array")
	arrayID := rootID(array, "raw-shape")
	references := ask(array, "reference-shape\n"+arrayID)
	if references[5] != arrayID || references[12] == "0" || references[18] != "2" {
		t.Fatalf("deferred reference arguments: %q", references)
	}
	if checker.ObjectFlagsClass != 1 || checker.ObjectFlagsReference != 4 || ast.SymbolFlagsMethod != 8192 || ast.SymbolFlagsOptional != 16777216 || ast.SymbolFlagsGetAccessor != 32768 || ast.SymbolFlagsSetAccessor != 65536 {
		t.Fatal("pinned native flag constants changed")
	}
	for _, question := range []string{"literal-value\n0", "type-properties\n01", "type-metadata\n9999999", "identical-types\n0\n1", "function-signatures\n-1", "property-exists\n" + stateID + "\n0", "contextual-argument\n1"} {
		if _, err := p.Inspect(file, uint64(call.Pos()), uint64(call.End()), "CallExpression", question); err == nil {
			t.Fatalf("accepted malformed question %q", question)
		}
	}
}
func mustID(t *testing.T, text string) uint64 {
	t.Helper()
	id, err := strconv.ParseUint(text, 10, 64)
	if err != nil || id == 0 {
		t.Fatalf("invalid identity %q", text)
	}
	return id
}
