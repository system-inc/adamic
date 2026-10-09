package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// nonconstructorFile is where the fixtures pretend to live.
const nonconstructorFile = "/repository/source/Nonconstructor.ts"

// Every case is oxc's, copied rather than rewritten.
//
// Verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_new_native_nonconstructor.rs`: 10 pass,
// 4 fail, snapshot 4 diagnostics from 4 inputs.
//
// `RunTypedFiles` rather than `Run`, and that is required rather than preferred: this rule declares
// `NeedsTypeChecker`, and the plain harness supplies no checker, so a fixture written against it
// would take the nil path and prove nothing while passing. The typed harness fails loudly when it
// cannot build one, which is what makes these fixtures evidence.
func TestNoNewNativeNonconstructorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"new Symbol", "var foo = new Symbol('foo');"},
		{"new BigInt", "var foo = new BigInt(9007199254740991);"},
		// A function named Symbol that is returned rather than bound at this scope does not shadow
		// the global here, so the outer `new Symbol` is still the global and still reports.
		{"a nested function named Symbol does not shadow the outer call",
			"function bar() { return function Symbol() {}; } var baz = new Symbol('baz');"},
		{"a nested function named BigInt does not shadow the outer call",
			"function bar() { return function BigInt() {}; } var baz = new BigInt(9007199254740991);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTypedFiles(t, NoNewNativeNonconstructor,
					map[string]string{nonconstructorFile: testCase.sourceText}, nonconstructorFile),
				"noNewNativeNonconstructor")
		})
	}
}

// Six of these ten are shadowing, which is why this rule reads the checker.
//
// A parameter, a function declaration, and a call passing the name as an argument each defeat a
// different naive implementation. A spelling test reports the first two; a test that checked only
// the callee position would still be wrong about them; and the argument forms are not `new`
// expressions on these names at all, so they pin that the callee is what is read.
func TestNoNewNativeNonconstructorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"calling Symbol without new", "var foo = Symbol('foo');"},
		{"a parameter shadowing Symbol",
			"function bar(Symbol: any) { var baz = new Symbol('baz'); }"},
		{"a function declaration shadowing Symbol", "function Symbol() {} new Symbol();"},
		{"Symbol passed as an argument", "new foo(Symbol);"},
		{"Symbol as a later argument", "new foo(bar, Symbol);"},
		{"calling BigInt without new", "var foo = BigInt(9007199254740991);"},
		{"a parameter shadowing BigInt",
			"function bar(BigInt: any) { var baz = new BigInt(9007199254740991); }"},
		{"a function declaration shadowing BigInt", "function BigInt() {} new BigInt();"},
		{"BigInt passed as an argument", "new foo(BigInt);"},
		{"BigInt as a later argument", "new foo(bar, BigInt);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTypedFiles(t, NoNewNativeNonconstructor,
					map[string]string{nonconstructorFile: testCase.sourceText}, nonconstructorFile))
		})
	}
}

// A case written because somebody read our code rather than upstream's.
//
// An import binding the name is a shadow upstream's corpus never writes, and it is the form most
// likely in real code: a module exporting its own `Symbol` helper. It reaches the same predicate,
// because the declaration lives in source rather than in the standard library, and pinning it here
// is what keeps that generality from being accidental.
func TestNoNewNativeNonconstructorDeclinesAnImportedShadow(t *testing.T) {
	t.Parallel()

	const helperFile = "/repository/source/Helper.ts"
	files := map[string]string{
		helperFile:         "export class Symbol { constructor(_name: string) {} }\n",
		nonconstructorFile: "import { Symbol } from './Helper';\nexport const a = new Symbol('x');\n",
	}
	rule_testing.ExpectClean(t,
		rule_testing.RunTypedFiles(t, NoNewNativeNonconstructor, files, nonconstructorFile))
}
