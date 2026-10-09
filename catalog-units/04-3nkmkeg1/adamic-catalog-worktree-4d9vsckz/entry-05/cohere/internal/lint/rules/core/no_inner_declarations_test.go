package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is ESLint's own, 39 valid and 29 invalid cases, extracted by loading its test file
// with a stubbed rule tester so nothing was retyped.
//
// Two invalid cases are not here and their absence is deliberate. Both configure
// `ecmaVersion: 5`, where a function declaration inside a block has no block scoping at all, so the
// `allow` setting cannot exempt it and the input reports even in strict code. Our parser has no
// such mode and cannot be asked for one, so those two are carried below as a recorded absence
// rather than greened or silently dropped.
//
// Six cases upstream configures with `sourceType: module` carry that in the source here instead,
// since our harness decides module-ness from the file's own imports and exports. Three of them
// already write an `export`; the other three are marked at the line.
//
// Options are routed through DecodeNoInnerDeclarationsOptions in every case, because the second
// option defaults to TRUE and a fixture building the struct directly would leave that inversion
// untested.

// decodeNoInnerDeclarationsForTest turns a fixture's JSON into options the way the config layer
// does. An empty string stands for a rule configured as a bare severity.
func decodeNoInnerDeclarationsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	decoded, err := DecodeNoInnerDeclarationsOptions([]byte(optionsJson))
	if err != nil {
		t.Fatalf("decoding %q: %v", optionsJson, err)
	}
	return decoded
}

func TestNoInnerDeclarationsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		source      string
		optionsJson string
	}{
		{name: "valid0", source: "function doSomething() { }", optionsJson: ""},
		{name: "valid1", source: "function doSomething() { function somethingElse() { } }", optionsJson: ""},
		{name: "valid2", source: "(function() { function doSomething() { } }());", optionsJson: ""},
		{name: "valid3", source: "if (test) { var fn = function() { }; }", optionsJson: ""},
		{name: "valid4", source: "if (test) { var fn = function expr() { }; }", optionsJson: ""},
		{name: "valid5", source: "function decl() { var fn = function expr() { }; }", optionsJson: ""},
		{name: "valid6", source: "function decl(arg) { var fn; if (arg) { fn = function() { }; } }", optionsJson: ""},
		{name: "valid7", source: "var x = {doSomething() {function doSomethingElse() {}}}", optionsJson: ""},
		{name: "valid8", source: "function decl(arg) { var fn; if (arg) { fn = function expr() { }; } }", optionsJson: ""},
		{name: "valid9", source: "function decl(arg) { var fn; if (arg) { fn = function expr() { }; } }", optionsJson: ""},
		{name: "valid10", source: "if (test) { var foo; }", optionsJson: ""},
		{name: "valid11", source: "if (test) { let x = 1; }", optionsJson: `["both"]`},
		{name: "valid12", source: "if (test) { const x = 1; }", optionsJson: `["both"]`},
		// Upstream sets sourceType module for this case; expressed by making the file one.
		{name: "valid13", source: "export {};\nif (test) { using x = 1; }", optionsJson: `["both"]`},
		// Upstream sets sourceType module for this case; expressed by making the file one.
		{name: "valid14", source: "export {};\nif (test) { await using x = 1; }", optionsJson: `["both"]`},
		{name: "valid15", source: "function doSomething() { while (test) { var foo; } }", optionsJson: ""},
		{name: "valid16", source: "var foo;", optionsJson: `["both"]`},
		{name: "valid17", source: "var foo = 42;", optionsJson: `["both"]`},
		{name: "valid18", source: "function doSomething() { var foo; }", optionsJson: `["both"]`},
		{name: "valid19", source: "(function() { var foo; }());", optionsJson: `["both"]`},
		{name: "valid20", source: "foo(() => { function bar() { } });", optionsJson: ""},
		{name: "valid21", source: "var fn = () => {var foo;}", optionsJson: `["both"]`},
		{name: "valid22", source: "var x = {doSomething() {var foo;}}", optionsJson: `["both"]`},
		{name: "valid23", source: "export var foo;", optionsJson: `["both"]`},
		{name: "valid24", source: "export function bar() {}", optionsJson: `["both"]`},
		{name: "valid25", source: "export default function baz() {}", optionsJson: `["both"]`},
		{name: "valid26", source: "exports.foo = () => {}", optionsJson: `["both"]`},
		{name: "valid27", source: "exports.foo = function(){}", optionsJson: `["both"]`},
		{name: "valid28", source: "module.exports = function foo(){}", optionsJson: `["both"]`},
		{name: "valid29", source: "class C { method() { function foo() {} } }", optionsJson: `["both"]`},
		{name: "valid30", source: "class C { method() { var x; } }", optionsJson: `["both"]`},
		{name: "valid31", source: "class C { static { function foo() {} } }", optionsJson: `["both"]`},
		{name: "valid32", source: "class C { static { var x; } }", optionsJson: `["both"]`},
		{name: "valid33", source: "'use strict' \n if (test) { function doSomething() { } }", optionsJson: "[\"functions\", {\"blockScopedFunctions\": \"allow\"}]"},
		{name: "valid34", source: "'use strict' \n if (test) { function doSomething() { } }", optionsJson: `["functions"]`},
		{name: "valid35", source: "function foo() {'use strict' \n if (test) { function doSomething() { } } }", optionsJson: "[\"functions\", {\"blockScopedFunctions\": \"allow\"}]"},
		// Upstream sets sourceType module for this case; expressed by making the file one.
		{name: "valid36", source: "export {};\nfunction foo() { { function bar() { } } }", optionsJson: "[\"functions\", {\"blockScopedFunctions\": \"allow\"}]"},
		{name: "valid37", source: "class C { method() { if(test) { function somethingElse() { } } } }", optionsJson: "[\"functions\", {\"blockScopedFunctions\": \"allow\"}]"},
		{name: "valid38", source: "const C = class { method() { if(test) { function somethingElse() { } } } }", optionsJson: "[\"functions\", {\"blockScopedFunctions\": \"allow\"}]"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInnerDeclarations,
				"file.ts", testCase.source, decodeNoInnerDeclarationsForTest(t, testCase.optionsJson)))
		})
	}
}

// Each case asserts the count and the two values the message interpolates. ExpectFindings sees
// neither: both slots are filled from the same node, so swapping them renders a sentence that still
// reads as English and names the wrong scope.
func TestNoInnerDeclarationsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		source      string
		optionsJson string
		want        [][2]string
	}{
		{name: "invalid0", source: "if (test) { function doSomething() { } }", optionsJson: `["both"]`, want: [][2]string{{"function", "program"}}},
		{name: "invalid1", source: "if (foo) var a; ", optionsJson: `["both"]`, want: [][2]string{{"variable", "program"}}},
		{name: "invalid2", source: "if (foo) /* some comments */ var a; ", optionsJson: `["both"]`, want: [][2]string{{"variable", "program"}}},
		{name: "invalid3", source: "if (foo){ function f(){ if(bar){ var a; } } }", optionsJson: `["both"]`, want: [][2]string{{"function", "program"}, {"variable", "function body"}}},
		{name: "invalid4", source: "if (foo) function f(){ if(bar) var a; }", optionsJson: `["both"]`, want: [][2]string{{"function", "program"}, {"variable", "function body"}}},
		{name: "invalid5", source: "if (foo) { var fn = function(){} } ", optionsJson: `["both"]`, want: [][2]string{{"variable", "program"}}},
		{name: "invalid6", source: "if (foo)  function f(){} ", optionsJson: "", want: [][2]string{{"function", "program"}}},
		{name: "invalid7", source: "function bar() { if (foo) function f(){}; }", optionsJson: `["both"]`, want: [][2]string{{"function", "function body"}}},
		{name: "invalid8", source: "function bar() { if (foo) var a; }", optionsJson: `["both"]`, want: [][2]string{{"variable", "function body"}}},
		{name: "invalid9", source: "if (foo) { var a; }", optionsJson: `["both"]`, want: [][2]string{{"variable", "program"}}},
		{name: "invalid10", source: "function doSomething() { do { function somethingElse() { } } while (test); }", optionsJson: "", want: [][2]string{{"function", "function body"}}},
		{name: "invalid11", source: "(function() { if (test) { function doSomething() { } } }());", optionsJson: "", want: [][2]string{{"function", "function body"}}},
		{name: "invalid12", source: "while (test) { var foo; }", optionsJson: `["both"]`, want: [][2]string{{"variable", "program"}}},
		{name: "invalid13", source: "function doSomething() { if (test) { var foo = 42; } }", optionsJson: `["both"]`, want: [][2]string{{"variable", "function body"}}},
		{name: "invalid14", source: "(function() { if (test) { var foo; } }());", optionsJson: `["both"]`, want: [][2]string{{"variable", "function body"}}},
		{name: "invalid15", source: "const doSomething = () => { if (test) { var foo = 42; } }", optionsJson: `["both"]`, want: [][2]string{{"variable", "function body"}}},
		{name: "invalid16", source: "class C { method() { if(test) { var foo; } } }", optionsJson: `["both"]`, want: [][2]string{{"variable", "function body"}}},
		{name: "invalid17", source: "class C { static { if (test) { var foo; } } }", optionsJson: `["both"]`, want: [][2]string{{"variable", "class static block body"}}},
		{name: "invalid18", source: "class C { static { if (test) { function foo() {} } } }", optionsJson: "[\"both\", {\"blockScopedFunctions\": \"disallow\"}]", want: [][2]string{{"function", "class static block body"}}},
		{name: "invalid19", source: "class C { static { if (test) { if (anotherTest) { var foo; } } } }", optionsJson: `["both"]`, want: [][2]string{{"variable", "class static block body"}}},
		{name: "invalid21", source: "if (test) { function doSomething() { } }", optionsJson: "[\"both\", {\"blockScopedFunctions\": \"disallow\"}]", want: [][2]string{{"function", "program"}}},
		{name: "invalid22", source: "'use strict' \n if (test) { function doSomething() { } }", optionsJson: "[\"both\", {\"blockScopedFunctions\": \"disallow\"}]", want: [][2]string{{"function", "program"}}},
		{name: "invalid23", source: "'use strict' \n if (test) { function doSomething() { } }", optionsJson: "[\"both\", {\"blockScopedFunctions\": \"disallow\"}]", want: [][2]string{{"function", "program"}}},
		{name: "invalid25", source: "function foo() {'use strict' \n { function bar() { } } }", optionsJson: "[\"both\", {\"blockScopedFunctions\": \"disallow\"}]", want: [][2]string{{"function", "function body"}}},
		{name: "invalid26", source: "function foo() {'use strict' \n { function bar() { } } }", optionsJson: "[\"both\", {\"blockScopedFunctions\": \"disallow\"}]", want: [][2]string{{"function", "function body"}}},
		{name: "invalid27", source: "function doSomething() { 'use strict' \n do { function somethingElse() { } } while (test); }", optionsJson: "[\"both\", {\"blockScopedFunctions\": \"disallow\"}]", want: [][2]string{{"function", "function body"}}},
		{name: "invalid28", source: "{ function foo () {'use strict' \n console.log('foo called'); } }", optionsJson: `["both"]`, want: [][2]string{{"function", "program"}}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoInnerDeclarations, "file.ts",
				testCase.source, decodeNoInnerDeclarationsForTest(t, testCase.optionsJson))
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("expected %d findings, got %d", len(testCase.want), len(result.Diagnostics))
			}
			for index, finding := range result.Diagnostics {
				if finding.Message.Id != "moveDeclToRoot" {
					t.Errorf("finding %d id: got %q", index, finding.Message.Id)
				}
				wantPrefix := fmt.Sprintf("Move %s declaration to %s root. ",
					testCase.want[index][0], testCase.want[index][1])
				if got := finding.Message.Description; !strings.HasPrefix(got, wantPrefix) {
					t.Errorf("finding %d message:\n got %q\nwant prefix %q", index, got, wantPrefix)
				}
			}
		})
	}
}

// The span, which no message assertion can see. Upstream reports the whole declaration.
func TestNoInnerDeclarationsSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		source      string
		optionsJson string
		wantText    string
	}{
		{"functionDeclaration", "if (foo) function f(){}", `["both"]`, "function f(){}"},
		{"variableStatement", "if (foo) var a;", `["both"]`, "var a;"},
		{"insideABareBlock", "{ var a; }", `["both"]`, "var a;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoInnerDeclarations, "file.ts",
				testCase.source, decodeNoInnerDeclarationsForTest(t, testCase.optionsJson))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected 1 finding, got %d", len(result.Diagnostics))
			}
			finding := result.Diagnostics[0]
			if got := testCase.source[finding.Range.Pos():finding.Range.End()]; got != testCase.wantText {
				t.Errorf("reported text: got %q, want %q", got, testCase.wantText)
			}
		})
	}
}

// The decoder, whose defaults are the two lines most likely to have no upstream counterpart.
func TestDecodeNoInnerDeclarationsOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                     string
		optionsJson              string
		wantBoth                 bool
		wantBlockScopedFunctions bool
	}{
		{"absentMeansFunctionsAndAllow", "", false, true},
		{"bareFunctions", `["functions"]`, false, true},
		{"bareBoth", `["both"]`, true, true},
		{"bothWithAllow", `["both", {"blockScopedFunctions": "allow"}]`, true, true},
		{"bothWithDisallow", `["both", {"blockScopedFunctions": "disallow"}]`, true, false},
		{"functionsWithDisallow", `["functions", {"blockScopedFunctions": "disallow"}]`, false, false},
		{"emptySecondObjectKeepsTheDefault", `["both", {}]`, true, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoInnerDeclarationsOptions([]byte(testCase.optionsJson))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.optionsJson, err)
			}
			settings, isSettings := decoded.(NoInnerDeclarationsOptions)
			if !isSettings {
				t.Fatalf("decoded to %T", decoded)
			}
			if settings.Both == nil || settings.BlockScopedFunctions == nil {
				t.Fatalf("a nil field would leave the rule reading a default it did not choose")
			}
			if *settings.Both != testCase.wantBoth {
				t.Errorf("both: got %v, want %v", *settings.Both, testCase.wantBoth)
			}
			if *settings.BlockScopedFunctions != testCase.wantBlockScopedFunctions {
				t.Errorf("blockScopedFunctions: got %v, want %v",
					*settings.BlockScopedFunctions, testCase.wantBlockScopedFunctions)
			}
		})
	}

	for _, bad := range []string{
		`["neither"]`,
		`["both", {"blockScopedFunctions": "maybe"}]`,
		// A bare string, which the config layer never delivers to a list rule.
		`"both"`,
		// A key the second element's schema does not declare, and a third element.
		`["both", {"blockScopedFunction": "allow"}]`,
		`["both", {"blockScopedFunctions": "allow"}, "functions"]`,
	} {
		if _, err := DecodeNoInnerDeclarationsOptions([]byte(bad)); err == nil {
			t.Errorf("decoding %q should have failed, and a silent fallback to the default would "+
				"turn a typo into a configuration nobody wrote", bad)
		}
	}
}

// A rule configured as a bare severity is handed nil, and the zero value of the options struct
// carries a nil BlockScopedFunctions. Since that option defaults to TRUE, reading nil as false
// would invert the rule and report every block-scoped function in the tree. Every fixture above
// routes through the decoder, so nothing there could see it.
func TestNoInnerDeclarationsHandlesNilOptions(t *testing.T) {
	t.Parallel()

	// Strict code, default options: a block-scoped function declaration is exempt.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInnerDeclarations, "file.ts",
		"export {};\nif (test) { function doSomething() { } }", nil))

	// Sloppy code, default options: it reports, so the nil path is not simply inert.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoInnerDeclarations, "file.ts",
		"if (test) { function doSomething() { } }", nil), "moveDeclToRoot")

	// And `var` is NOT checked under the default, which is the other half of the nil path.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInnerDeclarations, "file.ts",
		"if (test) { var foo; }", nil))
}

// A partially built options struct, which is the shape no other test here can produce.
//
// Every decoder path fills both pointers and the nil path falls back to the defaults, so the rule's
// `BlockScopedFunctions == nil` reading is unreachable through the wiring as it stands. A mutant
// inverting it survived the whole corpus for exactly that reason. It is still the line the brief
// names as most likely to be wrong, because the option defaults to TRUE and the zero value of a
// bool is false, so a caller assembling the struct by hand and leaving that field alone would flip
// the rule silently.
//
// Pinned here by handing the rule the shape only such a caller produces. The alternative was to
// delete the reading as unreachable, and the verdict would have been correct today and expired the
// first time anything constructed these options anywhere but the decoder.
func TestNoInnerDeclarationsPartialOptionsKeepTheDefault(t *testing.T) {
	t.Parallel()

	checkBoth := true
	partial := NoInnerDeclarationsOptions{Both: &checkBoth}

	// Strict code with the block-scoped-functions field left unset. The default is allow, so this
	// is clean; read as disallow it reports.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInnerDeclarations, "file.ts",
		"export {};\nif (test) { function doSomething() { } }", partial))

	// A control on the same struct, so the clean verdict above is not vacuous: `var` is checked
	// because Both was set, which proves the options reached the rule at all.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoInnerDeclarations, "file.ts",
		"export {};\nif (test) { var foo; }", partial), "moveDeclToRoot")
}

// The strictness gate, computed here rather than read off a scope table. Each source of strictness
// is pinned separately, because a rule that got only one of them right would pass most of the
// corpus.
func TestNoInnerDeclarationsStrictness(t *testing.T) {
	t.Parallel()

	allow := `["both", {"blockScopedFunctions": "allow"}]`

	exempt := []struct {
		name   string
		source string
	}{
		{"moduleIsStrict", "export {};\nif (test) { function doSomething() { } }"},
		{"programDirective", "'use strict'\nif (test) { function doSomething() { } }"},
		{"functionDirective", "function foo() { 'use strict'\n { function bar() { } } }"},
		{"classBodyIsStrict", "class C { method() { if (test) { function f() { } } } }"},
		// A prologue can hold several directives, so the scan continues past one that is not
		// `use strict` rather than stopping at it. Measured clean upstream.
		{"anotherDirectiveBeforeUseStrict", "'use asm'\n'use strict'\nif (test) { function f() { } }"},
		{"staticBlockIsStrict", "class C { static { if (test) { function f() { } } } }"},
	}
	for _, testCase := range exempt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInnerDeclarations,
				"file.ts", testCase.source, decodeNoInnerDeclarationsForTest(t, allow)))
		})
	}

	reports := []struct {
		name   string
		source string
	}{
		// A control, so the five clean verdicts above mean something.
		{"sloppyProgramReports", "if (test) { function doSomething() { } }"},
		// A directive that is not first is an ordinary expression statement and makes nothing
		// strict. No upstream case writes this and getting it wrong is silent.
		//
		// Three spellings, because they exercise DIFFERENT guards in the prologue scan and a
		// fixture covering only the first leaves the other reachable. `foo();` is an expression
		// statement whose expression is not a string, while `var a;` and `if(1){}` are not
		// expression statements at all, and only the latter two can see a scan that fails to stop
		// at the first non-directive. Found by a surviving mutant, then measured against the
		// installed build with a control that fired.
		{"nonDirectiveExpressionFirst", "foo();\n'use strict'\nif (test) { function f() { } }"},
		{"variableStatementFirst", "var a;\n'use strict'\nif (test) { function f() { } }"},
		{"ifStatementFirst", "if(1){}\n'use strict'\nif (test) { function f() { } }"},
		// A directive inside a NESTED block is not a prologue either.
		{"directiveInsideABlockIsNotAPrologue", "if (test) { 'use strict'\n function f() { } }"},
	}
	for _, testCase := range reports {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoInnerDeclarations,
				"file.ts", testCase.source, decodeNoInnerDeclarationsForTest(t, allow)),
				"moveDeclToRoot")
		})
	}
}

// Two upstream invalid cases this port cannot express, recorded rather than dropped.
//
// Both configure `ecmaVersion: 5`, where a function declaration inside a block has no block scoping
// at all. Upstream's `allow` setting is gated on `ecmaVersion >= 2015`, so under ES5 the exemption
// does not apply and the input reports even in strict code:
//
//	["both", {blockScopedFunctions: "allow"}]  ecmaVersion 5
//	  if (test) { function doSomething() { } }                          reports
//	  a 'use strict' directive above the same if statement           reports
//
// Our parser has one language edition and no way to be asked for an older one, so the gate has no
// surface here and is not implemented. Under every edition this tool can parse, the two inputs
// above behave as the fixtures in TestNoInnerDeclarationsStrictness assert. This is a parser
// difference rather than a rule difference and is stated so the next reader does not add the gate
// and find nothing can reach it.
func TestNoInnerDeclarationsCannotExpressTheEcmaScript5Gate(t *testing.T) {
	t.Parallel()

	allow := `["both", {"blockScopedFunctions": "allow"}]`
	// The sloppy half of each pair still reports here, for the ordinary reason.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoInnerDeclarations, "file.ts",
		"if (test) { function doSomething() { } }", decodeNoInnerDeclarationsForTest(t, allow)),
		"moveDeclToRoot")
	// The strict half is where the editions differ: upstream reports it under ES5 and is silent
	// from ES2015 on. This tool can only produce the second answer.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInnerDeclarations, "file.ts",
		"'use strict'\nif (test) { function doSomething() { } }",
		decodeNoInnerDeclarationsForTest(t, allow)))
}
