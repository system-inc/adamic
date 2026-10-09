package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/scanner"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// blockScopedVarFile is where the fixtures pretend to live.
const blockScopedVarFile = "/repository/source/BlockScopedVar.ts"

// blockScopedVarCase is one imported corpus row.
type blockScopedVarCase struct {
	sourceText string
	wantIds    []string
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `tests/lib/rules/block-scoped-var.js` was loaded with its RuleTester stubbed so every case came
// out as data, then each was replayed against the INSTALLED rule in `node_modules` to record how
// many findings it produces. All 106 reproduced, so the counts below are measurements. Four cases
// were byte-identical duplicates of another row and are dropped, leaving 102.
//
// The rule takes no options at all -- `schema: []` upstream -- so there is no option surface here
// and nothing to route through a decoder.
func blockScopedVarFiresCases() []blockScopedVarCase {
	return []blockScopedVarCase{
		{"function f(){ x; { var x; } }", []string{"outOfScope"}},
		{"function f(){ { var x; } x; }", []string{"outOfScope"}},
		{"function f() { var a; { var b = 0; } a = b; }", []string{"outOfScope"}},
		{"function f() { try { var a = 0; } catch (e) { var b = a; } }", []string{"outOfScope"}},
		{"function a() { for(var b in {}) { var c = b; } c; }", []string{"outOfScope"}},
		{"function a() { for(var b of {}) { var c = b; } c; }", []string{"outOfScope"}},
		{"function f(){ switch(2) { case 1: var b = 2; b; break; default: b; break;} b; }", []string{"outOfScope"}},
		{"for (var a = 0;;) {} a;", []string{"outOfScope"}},
		{"for (var a in []) {} a;", []string{"outOfScope"}},
		{"for (var a of []) {} a;", []string{"outOfScope"}},
		{"{ var a = 0; } a;", []string{"outOfScope"}},
		{"if (true) { var a; } a;", []string{"outOfScope"}},
		{"if (true) { var a = 1; } else { var a = 2; }", []string{"outOfScope", "outOfScope"}},
		{"for (var i = 0;;) {} for(var i = 0;;) {}", []string{"outOfScope", "outOfScope"}},
		{"class C { static { if (bar) { var foo; } foo; } }", []string{"outOfScope"}},
		{"{ var foo,\n  bar; } bar;", []string{"outOfScope"}},
		{"{ var { foo,\n  bar } = baz; } bar;", []string{"outOfScope"}},
		{"if (foo) { var a = 1; } else if (bar) { var a = 2; } else { var a = 3; }", []string{"outOfScope", "outOfScope", "outOfScope", "outOfScope", "outOfScope", "outOfScope"}},
	}
}

func blockScopedVarSilentCases() []blockScopedVarCase {
	return []blockScopedVarCase{
		{"function f() { } f(); var exports = { f: f };", []string{}},
		{"var f = () => {}; f(); var exports = { f: f };", []string{}},
		{"!function f(){ f; }", []string{}},
		{"function f() { var a, b; { a = true; } b = a; }", []string{}},
		{"var a; function f() { var b = a; }", []string{}},
		{"function f(a) { }", []string{}},
		{"!function(a) { };", []string{}},
		{"!function f(a) { };", []string{}},
		{"function f(a) { var b = a; }", []string{}},
		{"!function f(a) { var b = a; };", []string{}},
		{"function f() { var g = f; }", []string{}},
		{"function f() { } function g() { var f = g; }", []string{}},
		{"function f() { var hasOwnProperty; { hasOwnProperty; } }", []string{}},
		{"function f(){ a; b; var a, b; }", []string{}},
		{"function f(){ g(); function g(){} }", []string{}},
		{"if (true) { var a = 1; a; }", []string{}},
		{"var a; if (true) { a; }", []string{}},
		{"for (var i = 0; i < 10; i++) { i; }", []string{}},
		{"var i; for(i; i; i) { i; }", []string{}},
		{"function myFunc(foo) {  \"use strict\";  var { bar } = foo;  bar.hello();}", []string{}},
		{"function myFunc(foo) {  \"use strict\";  var [ bar ]  = foo;  bar.hello();}", []string{}},
		{"function myFunc(...foo) {  return foo;}", []string{}},
		{"var f = () => { var g = f; }", []string{}},
		{"class Foo {}\nexport default Foo;", []string{}},
		{"new Date", []string{}},
		{"var eslint = require('eslint');", []string{}},
		{"var fun = function({x}) {return x;};", []string{}},
		{"var fun = function([,x]) {return x;};", []string{}},
		{"function f(a) { return a.b; }", []string{}},
		{"var a = { \"foo\": 3 };", []string{}},
		{"var a = { foo: 3 };", []string{}},
		{"var a = { foo: 3, bar: 5 };", []string{}},
		{"var a = { set foo(a){}, get bar(){} };", []string{}},
		{"function f(a) { return arguments[0]; }", []string{}},
		{"function f() { }; var a = f;", []string{}},
		{"var a = f; function f() { };", []string{}},
		{"function f(){ for(var i; i; i) i; }", []string{}},
		{"function f(){ for(var a=0, b=1; a; b) a, b; }", []string{}},
		{"function f(){ for(var a in {}) a; }", []string{}},
		{"function f(){ switch(2) { case 1: var b = 2; b; break; default: b; break;} }", []string{}},
		{"a:;", []string{}},
		{"foo: while (true) { bar: for (var i = 0; i < 13; ++i) {if (i === 7) break foo; } }", []string{}},
		{"foo: while (true) { bar: for (var i = 0; i < 13; ++i) {if (i === 7) continue foo; } }", []string{}},
		{"const React = require(\"react/addons\");const cx = React.addons.classSet;", []string{}},
		{"var v = 1;  function x() { return v; };", []string{}},
		{"import * as y from \"./other.js\"; y();", []string{}},
		{"import y from \"./other.js\"; y();", []string{}},
		{"import {x as y} from \"./other.js\"; y();", []string{}},
		{"var x; export {x};", []string{}},
		{"var x; export {x as v};", []string{}},
		{"export {x} from \"./other.js\";", []string{}},
		{"export {x as v} from \"./other.js\";", []string{}},
		{"class Test { myFunction() { return true; }}", []string{}},
		{"class Test { get flag() { return true; }}", []string{}},
		{"var Test = class { myFunction() { return true; }}", []string{}},
		{"var doStuff; let {x: y} = {x: 1}; doStuff(y);", []string{}},
		{"function foo({x: y}) { return y; }", []string{}},
		{"!function f(){}; f", []string{}},
		{"var f = function foo() { }; foo(); var exports = { f: foo };", []string{}},
		{"var f = () => { x; }", []string{}},
		{"function f(){ x; }", []string{}},
		{"function f(a) { return a[b]; }", []string{}},
		{"function f() { return b.a; }", []string{}},
		{"var a = { foo: bar };", []string{}},
		{"var a = { foo: foo };", []string{}},
		{"var a = { bar: 7, foo: bar };", []string{}},
		{"var a = arguments;", []string{}},
		{"function x(){}; var a = arguments;", []string{}},
		{"function z(b){}; var a = b;", []string{}},
		{"function z(){var b;}; var a = b;", []string{}},
		{"function f(){ try{}catch(e){} e }", []string{}},
		{"a:b;", []string{}},
		{"/*global React*/ let {PropTypes, addons: {PureRenderMixin}} = React; let Test = React.createClass({mixins: [PureRenderMixin]});", []string{}},
		{"/*global prevState*/ const { virtualSize: prevVirtualSize = 0 } = prevState;", []string{}},
		{"const { dummy: { data, isLoading }, auth: { isLoggedIn } } = this.props;", []string{}},
		{"function a(n) { return n > 0 ? b(n - 1) : \"a\"; } function b(n) { return n > 0 ? a(n - 1) : \"b\"; }", []string{}},
		{"(function () { foo(); })(); function foo() {}", []string{}},
		{"class C { static { var foo; foo; } }", []string{}},
		{"class C { static { foo; var foo; } }", []string{}},
		{"class C { static { if (bar) { foo; } var foo; } }", []string{}},
		{"var foo; class C { static { foo; } } ", []string{}},
		{"class C { static { foo; } } var foo;", []string{}},
		{"var foo; class C { static {} [foo]; } ", []string{}},
		{"foo; class C { static {} } var foo; ", []string{}},
	}
}

func TestBlockScopedVarFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range blockScopedVarFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, BlockScopedVar,
				blockScopedVarFile, testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestBlockScopedVarStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range blockScopedVarSilentCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, BlockScopedVar,
				blockScopedVarFile, testCase.sourceText))
		})
	}
}

// Which identifier the finding points at, and which declaration it names.
//
// The message-id fixtures above see neither. This rule reports the USE, not the declaration, and
// the message names the declaration by line and column, so a port pointing at the declaration and
// naming the use would satisfy every count above while inverting the whole finding.
//
// Every expected span and every line/column pair was read off the installed rule's own output.
func TestBlockScopedVarPointsAtTheUseAndNamesTheDeclaration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		// The 1-based column each finding must POINT at. Columns rather than the spelled text,
		// because every identifier in these cases spells the same one or two letters: a test
		// comparing text passes just as happily when the finding points at the declaration
		// instead of the use, which is the exact inversion this test exists to refuse. A mutant
		// swapping the reported node survived a text comparison and fails this.
		wantColumns []int
		// The declaration coordinates the message must carry, one per finding, in order.
		wantDeclarations [][2]int
	}{
		{"{ var a = 0; } a;", []int{16}, [][2]int{{1, 7}}},
		{"function f(){ x; { var x; } }", []int{15}, [][2]int{{1, 24}}},
		{"function f() { var a; { var b = 0; } a = b; }", []int{42}, [][2]int{{1, 29}}},
		{"for (var i = 0;;) {} for(var i = 0;;) {}", []int{10, 30},
			[][2]int{{1, 30}, {1, 10}}},
		// A destructuring pattern introduces a name the declaration's own `Name()` is not, so a
		// port reading only that goes silent here. Upstream reports it.
		{"{ var { foo,\n  bar } = baz; } bar;", []int{18}, [][2]int{{2, 3}}},
		// The six-finding case: three declarations of one variable, each outside the other two
		// blocks, so each declaration name is reported twice. The columns and the coordinates
		// disagree on every row, which is what makes this the case that pins the pairing.
		{"if (foo) { var a = 1; } else if (bar) { var a = 2; } else { var a = 3; }",
			[]int{16, 16, 45, 45, 65, 65},
			[][2]int{{1, 45}, {1, 65}, {1, 16}, {1, 65}, {1, 16}, {1, 45}}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, BlockScopedVar, blockScopedVarFile,
				testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantColumns) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantColumns),
					len(result.Diagnostics))
			}
			for index, wantColumn := range testCase.wantColumns {
				span := result.Diagnostics[index].Range
				// Columns are read from the file the harness actually wrote, which is
				// `TrimSpace(source)+"\n"` under the typed harness, so an offset taken from the Go
				// literal would be off by one on any case with leading whitespace.
				_, column := scanner.GetLineAndCharacterOfPosition(result.SourceFile, span.Pos())
				if column+1 != wantColumn {
					t.Errorf("finding %d pointed at column %d, wanted %d", index, column+1,
						wantColumn)
				}
				line, column := testCase.wantDeclarations[index][0],
					testCase.wantDeclarations[index][1]
				// Assert the coordinates through the rendered text, because the format string is
				// the thing an id assertion cannot see and the thing most likely to be off by one.
				wantFragment := fmt.Sprintf("on line %d column %d", line, column)
				if got := result.Diagnostics[index].Message.Description; !strings.Contains(got,
					wantFragment) {
					t.Errorf("finding %d rendered\n  %q\nwhich does not carry %q", index, got,
						wantFragment)
				}
			}
		})
	}
}

// The whole rendered sentence, asserted by equality on one case.
//
// The test above checks the coordinates with `strings.Contains`, which is deliberately a weaker
// predicate than the property it guards and would stay green if the rest of the sentence were
// wrong. This closes that: one case, compared against a literal typed here rather than against the
// rule's own constant, since a comparison to the constant moves with the code under mutation.
func TestBlockScopedVarRendersTheWholeMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, BlockScopedVar, blockScopedVarFile, "{ var a = 0; } a;")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	want := "`a` is declared with `var` inside a block on line 1 column 7, and used here, " +
		"outside it. That works because `var` is function-scoped rather than block-scoped, so " +
		"the braces around the declaration say one thing and the binding does another. Declare " +
		"it with `let` in the scope that actually needs it, or move the declaration out to where " +
		"it is used."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Errorf("rendered\n  %q\nwanted\n  %q", got, want)
	}
	if got := result.Diagnostics[0].Message.Id; got != "outOfScope" {
		t.Errorf("message id was %q, wanted outOfScope", got)
	}
}

// The typed harness is required, and a revert to `Run` must fail loudly rather than go quiet.
//
// `GetSymbolAtLocation` on a nil checker returns nil rather than crashing, so this rule would find
// no references at all and every silent fixture would pass vacuously while every firing one failed.
// The guard at the top of the listener makes that explicit, and this pins it.
func TestBlockScopedVarNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "{ var a = 0; } a;"
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, BlockScopedVar, blockScopedVarFile,
		source), "outOfScope")
	rule_testing.ExpectClean(t, rule_testing.Run(t, BlockScopedVar, blockScopedVarFile, source))
}

// `let` and `const` are already block-scoped, so nothing about them is out of scope.
//
// Written for a surviving mutant: forcing the declaration-kind test to accept everything left all
// 102 imported cases green. The corpus does carry `let` and `const`, but never declared inside a
// block and read outside it, which is the only shape where the two kinds disagree. So the rule
// could have reported every block-scoped declaration in the tree and nothing here would have said
// so.
//
// Each row is paired with the `var` spelling of the same shape, so a row that went silent for the
// wrong reason -- a parse failure, a shape the container walk does not recognise -- fails its
// partner rather than passing quietly as coverage. Every verdict was measured against the installed
// rule.
func TestBlockScopedVarIgnoresBlockScopedDeclarations(t *testing.T) {
	t.Parallel()

	cases := []blockScopedVarCase{
		{"function f(){ { let a = 1; } a; }", []string{}},
		{"function f(){ { var a = 1; } a; }", []string{"outOfScope"}},
		{"function f(){ { const a = 1; } a; }", []string{}},
		{"for (let i = 0;;) {} i;", []string{}},
		{"for (var i = 0;;) {} i;", []string{"outOfScope"}},
		{"if (true) { let a; } a;", []string{}},
		{"if (true) { var a; } a;", []string{"outOfScope"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, BlockScopedVar, blockScopedVarFile,
				testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}
