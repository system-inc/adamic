package program_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/types/program"
)

// TestShimFieldAccessorsReadTheFieldsTheyName guards every shim accessor that is reached through an
// unsafe pointer cast rather than through a method.
//
// `shim/checker/shim.go` mirrors five upstream structs and reads eleven of their fields by casting
// the real struct to that mirror. Those eleven are the whole unsafe surface: every other accessor in
// the file is a `//go:linkname` to an upstream method and reads nothing by offset. The cast is only
// correct while the mirror reproduces upstream's field offsets exactly, and nothing about it fails
// loudly when it does not: the read returns whatever value happens to sit at the wrong offset, which
// is a perfectly valid value of the right Go type that simply is not the field asked for.
//
// That is not hypothetical. The mirror substitutes stand-in types for upstream's unexported ones,
// configured in `shim/checker/extra-shim.json`, and `checker.symbolArenaLinkStore` was substituted
// with `core.PagedLinkStore`. Upstream's type is a PagedLinkStore of pointers PLUS an Arena of
// values, so the stand-in was 24 bytes short at field 99 of 319 and every field after it read at
// the wrong offset. `Checker_numberType` returned the `null` type and `Checker_booleanType`
// returned `false`.
//
// Nothing in the tree noticed. Every package compiled, every test passed, and the only symptom was
// that `no-for-in-array` reported nothing on any input, because its whole predicate asks whether a
// type has an index signature keyed by `Checker_numberType`. It asked for a `null` index, got nil
// for every input in existence, and declined the entire corpus. A rule that is present and blind is
// the exact failure this project exists to prevent, and it arrived from a layout drift two packages
// away from any rule.
//
// Two properties of that defect set the standard every assertion here has to meet. The field NAMES
// aligned at all 319 positions and only a TYPE was wrong, so a name-by-name comparison would have
// called the mirror correct. And `numberType` returning the `null` type is non-nil, so a nil check
// passes straight through it. Every case below therefore asserts what the value IS — the checker's
// own name for a type, a specific flag set, the actual parameter names — never merely that it is
// present. Where a field's correct value cannot be pinned that way, this says so rather than
// spending a weak assertion that reads as coverage.
//
// The generator refuses a substitution whose size or alignment does not match upstream, which is a
// stronger and earlier check than this one: it fails before the wrong code exists, and it covers
// substitutions no accessor reads today. This test covers what that check cannot see. A stand-in of
// the RIGHT size in the WRONG order, or an upstream field reordering that keeps the struct the same
// size, moves these fields without changing any size the generator compares.
func TestShimFieldAccessorsReadTheFieldsTheyName(t *testing.T) {
	t.Parallel()
	fileChecker, sourceFile, release := buildLayoutSubject(t)
	defer release()

	declaredTypes, valueTypes := layoutSubjectTypes(t, fileChecker, sourceFile)

	// --- Checker: numberType, stringType, booleanType, globalRegExpType -------------------------
	//
	// TypeToString is the checker's own name for a type, so this compares what the accessor
	// returned against what the checker calls it rather than against another shim read.
	checkerCases := []struct {
		accessor string
		actual   *checker.Type
		want     string
	}{
		{"Checker_numberType", checker.Checker_numberType(fileChecker), "number"},
		{"Checker_stringType", checker.Checker_stringType(fileChecker), "string"},
		{"Checker_booleanType", checker.Checker_booleanType(fileChecker), "boolean"},
		{"Checker_globalRegExpType", checker.Checker_globalRegExpType(fileChecker), "RegExp"},
	}
	for _, testCase := range checkerCases {
		if testCase.actual == nil {
			t.Fatalf("%s returned nil, so the struct mirror in shim/checker no longer matches "+
				"upstream's Checker layout", testCase.accessor)
		}
		if got := fileChecker.TypeToString(testCase.actual); got != testCase.want {
			t.Fatalf("%s returned the %q type rather than %q.%s", testCase.accessor, got,
				testCase.want, layoutDriftAdvice)
		}
	}

	// --- Type: flags, objectFlags, symbol, alias -------------------------------------------------
	//
	// `flags` and `objectFlags` are checked against a type whose name the checker states
	// independently, so a wrong offset has to produce the exactly right flag set to pass. A literal
	// and an object type are both used because they occupy different flag bits: a single subject
	// would let a stuck value satisfy the assertion.
	numberLiteral := valueTypes["numberLiteralValue"]
	if numberLiteral == nil {
		t.Fatal("the fixture produced no type for numberLiteralValue, so nothing below is measured")
	}
	if got := fileChecker.TypeToString(numberLiteral); got != "1" {
		t.Fatalf("the fixture subject numberLiteralValue is the %q type, not the number literal "+
			"this test was written against", got)
	}
	if flags := checker.Type_flags(numberLiteral); flags&checker.TypeFlagsNumberLiteral == 0 {
		t.Fatalf("Type_flags on the `1` literal returned %v, which does not include "+
			"TypeFlagsNumberLiteral.%s", flags, layoutDriftAdvice)
	}
	if flags := checker.Type_flags(numberLiteral); flags&checker.TypeFlagsObject != 0 {
		t.Fatalf("Type_flags on the `1` literal returned %v, which wrongly includes "+
			"TypeFlagsObject.%s", flags, layoutDriftAdvice)
	}

	boxInstance := valueTypes["boxValue"]
	if boxInstance == nil {
		t.Fatal("the fixture produced no type for boxValue")
	}
	if flags := checker.Type_flags(boxInstance); flags&checker.TypeFlagsObject == 0 {
		t.Fatalf("Type_flags on an interface instance returned %v, which does not include "+
			"TypeFlagsObject.%s", flags, layoutDriftAdvice)
	}
	// `Box` is a declared interface, so its instance carries ObjectFlagsReference; a plain object
	// literal type does not. Asserting one flag present and one absent means a stuck or shifted
	// read cannot satisfy both.
	boxObjectFlags := checker.Type_objectFlags(boxInstance)
	if boxObjectFlags&checker.ObjectFlagsReference == 0 {
		t.Fatalf("Type_objectFlags on the Box<string> instance returned %v, which does not include "+
			"ObjectFlagsReference.%s", boxObjectFlags, layoutDriftAdvice)
	}
	if boxObjectFlags&checker.ObjectFlagsTuple != 0 {
		t.Fatalf("Type_objectFlags on the Box<string> instance returned %v, which wrongly includes "+
			"ObjectFlagsTuple.%s", boxObjectFlags, layoutDriftAdvice)
	}

	// `symbol` is asserted by NAME, which is the check the historical defect would have failed:
	// a shifted read yields some other valid *ast.Symbol, and only its name distinguishes them.
	boxSymbol := checker.Type_symbol(boxInstance)
	if boxSymbol == nil {
		t.Fatalf("Type_symbol on the Box<string> instance returned nil.%s", layoutDriftAdvice)
	}
	if boxSymbol.Name != "Box" {
		t.Fatalf("Type_symbol on the Box<string> instance named %q rather than \"Box\".%s",
			boxSymbol.Name, layoutDriftAdvice)
	}

	// `alias` is asserted in both directions. A type written through an alias carries one and
	// names it; a type that was never aliased carries nil. A read stuck on either answer fails one
	// of the two.
	aliasedTuple := valueTypes["pairValue"]
	if aliasedTuple == nil {
		t.Fatal("the fixture produced no type for pairValue")
	}
	tupleAlias := checker.Type_alias(aliasedTuple)
	if tupleAlias == nil {
		t.Fatalf("Type_alias on a value declared as the alias `Pair` returned nil.%s",
			layoutDriftAdvice)
	}
	if symbol := tupleAlias.Symbol(); symbol == nil || symbol.Name != "Pair" {
		t.Fatalf("Type_alias on a value declared as `Pair` named %v rather than \"Pair\".%s",
			aliasSymbolName(tupleAlias), layoutDriftAdvice)
	}
	if unaliased := checker.Type_alias(numberLiteral); unaliased != nil {
		t.Fatalf("Type_alias on the `1` literal, which was never written through an alias, "+
			"returned %v rather than nil.%s", aliasSymbolName(unaliased), layoutDriftAdvice)
	}

	// --- TupleType: combinedFlags ----------------------------------------------------------------
	//
	// Three tuple shapes that differ only in their element flags. Each expected value is distinct,
	// so no single stuck read satisfies more than one, and each is spelled with the upstream
	// constants rather than a magic number.
	tupleCases := []struct {
		subject string
		want    checker.ElementFlags
	}{
		{"pairValue", checker.ElementFlagsRequired},
		{"optionalValue", checker.ElementFlagsRequired | checker.ElementFlagsOptional},
		{"restValue", checker.ElementFlagsRequired | checker.ElementFlagsRest},
	}
	for _, testCase := range tupleCases {
		subjectType := valueTypes[testCase.subject]
		if subjectType == nil {
			t.Fatalf("the fixture produced no type for %s", testCase.subject)
		}
		if !checker.IsTupleType(subjectType) {
			t.Fatalf("the fixture subject %s is not a tuple type, so this case measures nothing",
				testCase.subject)
		}
		target := subjectType.Target()
		if target == nil {
			t.Fatalf("the tuple %s has no target type", testCase.subject)
		}
		tupleType := target.AsTupleType()
		if tupleType == nil {
			t.Fatalf("the tuple %s did not resolve to a TupleType", testCase.subject)
		}
		if got := checker.TupleType_combinedFlags(tupleType); got != testCase.want {
			t.Fatalf("TupleType_combinedFlags on %s returned %v rather than %v.%s",
				testCase.subject, got, testCase.want, layoutDriftAdvice)
		}
	}

	// --- InterfaceType: thisType -----------------------------------------------------------------
	//
	// An interface with a polymorphic `this` return has a thisType; the checker names it "this" and
	// classifies it as a type parameter. Both are asserted, because the name alone would also match
	// a shifted read that happened to land on another type the checker prints the same way.
	boxDeclared := declaredTypes["Box"]
	if boxDeclared == nil {
		t.Fatal("the fixture produced no declared type for the interface Box")
	}
	interfaceType := boxDeclared.AsInterfaceType()
	if interfaceType == nil {
		t.Fatal("the declared type of the interface Box did not resolve to an InterfaceType")
	}
	thisType := checker.InterfaceType_thisType(interfaceType)
	if thisType == nil {
		t.Fatalf("InterfaceType_thisType on an interface with a polymorphic `this` returned nil.%s",
			layoutDriftAdvice)
	}
	if got := fileChecker.TypeToString(thisType); got != "this" {
		t.Fatalf("InterfaceType_thisType returned the %q type rather than \"this\".%s",
			got, layoutDriftAdvice)
	}
	if flags := checker.Type_flags(thisType); flags&checker.TypeFlagsTypeParameter == 0 {
		t.Fatalf("InterfaceType_thisType returned a type whose flags are %v, which does not "+
			"include TypeFlagsTypeParameter as a `this` type must.%s", flags, layoutDriftAdvice)
	}

	// --- Signature: parameters, declaration -------------------------------------------------------
	//
	// The parameter names are asserted in order. That is the strongest available statement about
	// this field: a shifted read returns some other slice, and both its length and its contents
	// would have to coincide with `[first second third]` to pass.
	functionType := valueTypes["twoParams"]
	if functionType == nil {
		t.Fatal("the fixture produced no type for twoParams")
	}
	signatures := checker.Checker_getSignaturesOfType(fileChecker, functionType, checker.SignatureKindCall)
	if len(signatures) != 1 {
		t.Fatalf("the fixture subject twoParams has %d call signatures rather than 1",
			len(signatures))
	}
	parameterNames := []string{}
	for _, parameter := range checker.Signature_parameters(signatures[0]) {
		parameterNames = append(parameterNames, parameter.Name)
	}
	wantParameterNames := []string{"first", "second", "third"}
	if len(parameterNames) != len(wantParameterNames) {
		t.Fatalf("Signature_parameters returned %d parameters (%v) rather than %d.%s",
			len(parameterNames), parameterNames, len(wantParameterNames), layoutDriftAdvice)
	}
	for i, want := range wantParameterNames {
		if parameterNames[i] != want {
			t.Fatalf("Signature_parameters returned %v rather than %v.%s",
				parameterNames, wantParameterNames, layoutDriftAdvice)
		}
	}

	// `declaration` is asserted by node kind AND by the name at that node, so a shifted read has to
	// land on a function declaration that is also named `twoParams`.
	declaration := checker.Signature_declaration(signatures[0])
	if declaration == nil {
		t.Fatalf("Signature_declaration on a declared function returned nil.%s", layoutDriftAdvice)
	}
	if declaration.Kind != ast.KindFunctionDeclaration {
		t.Fatalf("Signature_declaration returned a %v node rather than a function declaration.%s",
			declaration.Kind, layoutDriftAdvice)
	}
	if name := declaration.Name(); name == nil || name.Text() != "twoParams" {
		t.Fatalf("Signature_declaration returned a function declaration named %q rather than "+
			"\"twoParams\".%s", declarationName(declaration), layoutDriftAdvice)
	}
}

const layoutDriftAdvice = "\n" +
	"The mirror in shim/checker/shim.go has drifted from upstream's layout, so this accessor is " +
	"reading a different field.\n" +
	"Two causes are worth separating. If `go run ./internal/types/tools/generate_shims` now refuses, a substitution in " +
	"shim/checker/extra-shim.json no longer matches the size or alignment of the upstream type it " +
	"replaces, and its message names which one.\n" +
	"If the generator is happy and this still fails, the drift is one size alone cannot see: a " +
	"stand-in of the right width in the wrong field order, or an upstream reordering that left the " +
	"struct the same size. Compare the mirror against upstream's struct field by field."

func aliasSymbolName(alias *checker.TypeAlias) string {
	if alias == nil {
		return "<nil>"
	}
	if symbol := alias.Symbol(); symbol != nil {
		return symbol.Name
	}
	return "<alias with no symbol>"
}

func declarationName(node *ast.Node) string {
	if node == nil {
		return "<nil>"
	}
	if name := node.Name(); name != nil {
		return name.Text()
	}
	return "<unnamed>"
}

// layoutSubjectSource is written so that every one of the eleven unsafe reads has a subject whose
// correct value is something the checker will state independently.
const layoutSubjectSource = `
export type Pair = [string, number];
export type Rest = [string, ...number[]];
export type Optional = [string, number?];
export interface Box<T> { value: T; self(): this; }
export function twoParams(first: string, second: number, third: boolean): boolean {
	return first.length > second && third;
}
export const pairValue: Pair = ["a", 1];
export const restValue: Rest = ["a", 1, 2];
export const optionalValue: Optional = ["a"];
export const boxValue: Box<string> = { value: "v", self() { return this; } };
export const numberLiteralValue = 1;
`

func buildLayoutSubject(t *testing.T) (*checker.Checker, *ast.SourceFile, func()) {
	t.Helper()
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "Subject.ts")
	if err := os.WriteFile(sourcePath, []byte(layoutSubjectSource), 0o644); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}
	configPath := filepath.Join(directory, "tsconfig.json")
	config := `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"types":[]},"include":["*.ts"]}`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("writing the tsconfig: %v", err)
	}

	graph, err := program.Build(program.Options{
		ConfigFileName:   configPath,
		CurrentDirectory: directory,
		SingleThreaded:   true,
	})
	if err != nil {
		t.Fatalf("building the type graph: %v", err)
	}
	files := graph.ProjectFiles()
	if len(files) == 0 {
		t.Fatal("the program contains no project files, so this test would prove nothing")
	}
	fileChecker, release := graph.CheckerForFile(context.Background(), files[0])
	return fileChecker, files[0], release
}

// layoutSubjectTypes returns the declared types of the fixture's type and interface declarations,
// and the value types of its variables and functions, both keyed by name.
func layoutSubjectTypes(
	t *testing.T,
	fileChecker *checker.Checker,
	sourceFile *ast.SourceFile,
) (declared map[string]*checker.Type, values map[string]*checker.Type) {
	t.Helper()
	declared = map[string]*checker.Type{}
	values = map[string]*checker.Type{}

	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
			if name := node.Name(); name != nil {
				if symbol := fileChecker.GetSymbolAtLocation(name); symbol != nil {
					declared[name.Text()] = checker.Checker_getDeclaredTypeOfSymbol(fileChecker, symbol)
				}
			}
		case ast.KindVariableDeclaration, ast.KindFunctionDeclaration:
			if name := node.Name(); name != nil {
				if symbol := fileChecker.GetSymbolAtLocation(name); symbol != nil {
					values[name.Text()] = checker.Checker_getTypeOfSymbol(fileChecker, symbol)
				}
			}
		}
		node.ForEachChild(walk)
		return false
	}
	sourceFile.AsNode().ForEachChild(walk)

	if len(declared) == 0 || len(values) == 0 {
		t.Fatal("the fixture yielded no types, so every assertion below would pass vacuously")
	}
	return declared, values
}
