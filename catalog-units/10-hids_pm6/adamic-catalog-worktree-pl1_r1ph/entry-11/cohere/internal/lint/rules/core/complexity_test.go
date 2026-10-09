package core

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// complexityFile is where the fixtures pretend to live.
const complexityFile = "/repository/source/Complexity.ts"

// complexityOf builds the settings a config carrying this option would decode to.
//
// The argument is the option as JSON TEXT, exactly the bytes the config layer hands the decoder, so
// a fixture cannot bypass the decoder or test a shape a config could not deliver.
func complexityOf(optionJson string) any {
	settings, err := DecodeComplexityOptions([]byte(optionJson))
	if err != nil {
		panic(err)
	}
	return settings
}

// The corpus is ESLint's own, taken from
// /tmp/lint-sources-fresh/eslint/tests/lib/rules/complexity.js by RUNNING that file with RuleTester
// intercepted. 165 cases were captured from the single RuleTester.run call.
//
// Every expectation is what the INSTALLED ESLint 10.8.1 rule answered when driven over that case
// through the Linter interface under `sourceType: module` with the typescript-eslint parser. The
// oracle reproduced all 165 of the corpus's own declared verdicts under upstream's own
// languageOptions, and ZERO cases change verdict between that configuration and cohere's.
//
//	96 reporting cases, 69 clean cases.
func TestComplexityFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		findings int
	}{
		{source: "function a(x) {}", settings: complexityOf("0"), findings: 1},
		{source: "function foo(x) {if (x > 10) {return 'x is greater than 10';} else if (x > 5) {return 'x is greater than 5';} else {return 'x is less than 5';}}", settings: complexityOf("2"), findings: 1},
		{source: "var func = function () {}", settings: complexityOf("0"), findings: 1},
		{source: "var obj = { a(x) {} }", settings: complexityOf("0"), findings: 1},
		{source: "class Test { a(x) {} }", settings: complexityOf("0"), findings: 1},
		{source: "var a = (x) => {if (true) {return x;}}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {if (true) {return x;}}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {if (true) {return x;} else {return x+1;}}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {if (true) {return x;} else if (false) {return x+1;} else {return 4;}}", settings: complexityOf("2"), findings: 1},
		{source: "function a(x) {for(var i = 0; i < 5; i ++) {x ++;} return x;}", settings: complexityOf("1"), findings: 1},
		{source: "function a(obj) {for(var i in obj) {obj[i] = 3;}}", settings: complexityOf("1"), findings: 1},
		{source: "function a(obj) {for(var i of obj) {obj[i] = 3;}}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {for(var i = 0; i < 5; i ++) {if(i % 2 === 0) {x ++;}} return x;}", settings: complexityOf("2"), findings: 1},
		{source: "function a(obj) {if(obj){ for(var x in obj) {try {x.getThis();} catch (e) {x.getThat();}}} else {return false;}}", settings: complexityOf("3"), findings: 1},
		{source: "function a(x) {try {x.getThis();} catch (e) {x.getThat();}}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {return x === 4 ? 3 : 5;}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {return x === 4 ? 3 : (x === 3 ? 2 : 1);}", settings: complexityOf("2"), findings: 1},
		{source: "function a(x) {return x || 4;}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {x && 4;}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {x ?? 4;}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {x ||= 4;}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {x &&= 4;}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {x ??= 4;}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}", settings: complexityOf("2"), findings: 1},
		{source: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: if(x == 'foo') {5;};}}", settings: complexityOf("3"), findings: 1},
		{source: "function a(x) {while(true) {'foo';}}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {do {'foo';} while (true)}", settings: complexityOf("1"), findings: 1},
		{source: "function a(x) {(function() {while(true){'foo';}})(); (function() {while(true){'bar';}})();}", settings: complexityOf("1"), findings: 2},
		{source: "function a(x) {(function() {while(true){'foo';}})(); (function() {'bar';})();}", settings: complexityOf("1"), findings: 1},
		{source: "var obj = { a(x) { return x ? 0 : 1; } };", settings: complexityOf("1"), findings: 1},
		{source: "var obj = { a: function b(x) { return x ? 0 : 1; } };", settings: complexityOf("1"), findings: 1},
		{source: "function test (a) { if (a === 1) {} else if (a === 2) {} else if (a === 3) {} else if (a === 4) {} else if (a === 5) {} else if (a === 6) {} else if (a === 7) {} else if (a === 8) {} else if (a === 9) {} else if (a === 10) {} else if (a === 11) {} else if (a === 12) {} else if (a === 13) {} else if (a === 14) {} else if (a === 15) {} else if (a === 16) {} else if (a === 17) {} else if (a === 18) {} else if (a === 19) {} else if (a === 20) {} };", settings: nil, findings: 1},
		{source: "function test (a) { if (a === 1) {} else if (a === 2) {} else if (a === 3) {} else if (a === 4) {} else if (a === 5) {} else if (a === 6) {} else if (a === 7) {} else if (a === 8) {} else if (a === 9) {} else if (a === 10) {} else if (a === 11) {} else if (a === 12) {} else if (a === 13) {} else if (a === 14) {} else if (a === 15) {} else if (a === 16) {} else if (a === 17) {} else if (a === 18) {} else if (a === 19) {} else if (a === 20) {} };", settings: complexityOf("{}"), findings: 1},
		{source: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}", settings: complexityOf("{\"max\": 1, \"variant\": \"modified\"}"), findings: 1},
		{source: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: if(x == 'foo') {5;};}}", settings: complexityOf("{\"max\": 2, \"variant\": \"modified\"}"), findings: 1},
		{source: "function foo () { a || b; class C { x; } c || d; }", settings: complexityOf("2"), findings: 1},
		{source: "function foo () { a || b; class C { x = c; } d || e; }", settings: complexityOf("2"), findings: 1},
		{source: "function foo () { a || b; class C { [x || y]; } }", settings: complexityOf("2"), findings: 1},
		{source: "function foo () { a || b; class C { [x || y] = c; } }", settings: complexityOf("2"), findings: 1},
		{source: "function foo () { class C { [x || y]; } a || b; }", settings: complexityOf("2"), findings: 1},
		{source: "function foo () { class C { [x || y] = a; } b || c; }", settings: complexityOf("2"), findings: 1},
		{source: "function foo () { class C { [x || y]; [z || q]; } }", settings: complexityOf("2"), findings: 1},
		{source: "function foo () { class C { [x || y] = a; [z || q] = b; } }", settings: complexityOf("2"), findings: 1},
		{source: "function foo () { a || b; class C { x = c || d; } e || f; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x(){ a || b; } y = c || d || e; z() { f || g; } }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x = a || b; y() { c || d || e; } z = f || g; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x; y() { c || d || e; } z; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x = a || b; }", settings: complexityOf("1"), findings: 1},
		{source: "(class { x = a || b; })", settings: complexityOf("1"), findings: 1},
		{source: "class C { static x = a || b; }", settings: complexityOf("1"), findings: 1},
		{source: "(class { x = a ? b : c; })", settings: complexityOf("1"), findings: 1},
		{source: "class C { x = a || b || c; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x = a || b; y = b || c || d; z = e || f; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x = a || b || c; y = d || e; z = f || g || h; }", settings: complexityOf("2"), findings: 2},
		{source: "class C { x = () => a || b || c; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x = (() => a || b || c) || d; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x = () => a || b || c; y = d || e; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x = () => a || b || c; y = d || e || f; }", settings: complexityOf("2"), findings: 2},
		{source: "class C { x = function () { a || b }; y = function () { c || d }; }", settings: complexityOf("1"), findings: 2},
		{source: "class C { x = class { [y || z]; }; }", settings: complexityOf("1"), findings: 1},
		{source: "class C { x = class { [y || z] = a; }; }", settings: complexityOf("1"), findings: 1},
		{source: "class C { x = class { y = a || b; }; }", settings: complexityOf("1"), findings: 1},
		{source: "function foo () { a || b; class C { static {} } c || d; }", settings: complexityOf("2"), findings: 1},
		{source: "function foo () { a || b; class C { static { c || d; } } e || f; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { static { a || b; }  }", settings: complexityOf("1"), findings: 1},
		{source: "class C { static { a || b || c; }  }", settings: complexityOf("2"), findings: 1},
		{source: "class C { static { a || b; c || d; }  }", settings: complexityOf("2"), findings: 1},
		{source: "class C { static { a || b; c || d; e || f; }  }", settings: complexityOf("3"), findings: 1},
		{source: "class C { static { a || b; c || d; { e || f; } }  }", settings: complexityOf("3"), findings: 1},
		{source: "class C { static { if (a || b) c = d || e; } }", settings: complexityOf("3"), findings: 1},
		{source: "class C { static { if (a || b) c = (d => e || f)() || (g => h || i)(); } }", settings: complexityOf("3"), findings: 1},
		{source: "class C { x(){ a || b; } static { c || d || e; } z() { f || g; } }", settings: complexityOf("2"), findings: 1},
		{source: "class C { x = a || b; static { c || d || e; } y = f || g; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { static x = a || b; static { c || d || e; } static y = f || g; }", settings: complexityOf("2"), findings: 1},
		{source: "class C { static { a || b; } static(){ c || d || e; } static { f || g; } }", settings: complexityOf("2"), findings: 1},
		{source: "class C { static { a || b; } static static(){ c || d || e; } static { f || g; } }", settings: complexityOf("2"), findings: 1},
		{source: "class C { static { a || b; } static x = c || d || e; static { f || g; } }", settings: complexityOf("2"), findings: 1},
		{source: "class C { static { a || b || c || d; } static { e || f || g; } }", settings: complexityOf("3"), findings: 1},
		{source: "class C { static { a || b || c; } static { d || e || f || g; } }", settings: complexityOf("3"), findings: 1},
		{source: "class C { static { a || b || c || d; } static { e || f || g || h; } }", settings: complexityOf("3"), findings: 2},
		{source: "class C { x = () => a || b || c; y = f || g || h; }", settings: complexityOf("2"), findings: 2},
		{source: "function a(x) {}", settings: complexityOf("{\"max\": 0}"), findings: 1},
		{source: "const obj = { b: (a) => a?.b?.c, c: function (a) { return a?.b?.c; } };", settings: complexityOf("{\"max\": 2}"), findings: 2},
		{source: "function a(b) { b?.c; }", settings: complexityOf("{\"max\": 1}"), findings: 1},
		{source: "function a(b) { b?.['c']; }", settings: complexityOf("{\"max\": 1}"), findings: 1},
		{source: "function a(b) { b?.c; d || e; }", settings: complexityOf("{\"max\": 2}"), findings: 1},
		{source: "function a(b) { b?.c?.d; }", settings: complexityOf("{\"max\": 2}"), findings: 1},
		{source: "function a(b) { b?.['c']?.['d']; }", settings: complexityOf("{\"max\": 2}"), findings: 1},
		{source: "function a(b) { b?.c?.['d']; }", settings: complexityOf("{\"max\": 2}"), findings: 1},
		{source: "function a(b) { b?.c.d?.e; }", settings: complexityOf("{\"max\": 2}"), findings: 1},
		{source: "function a(b) { b?.c?.(); }", settings: complexityOf("{\"max\": 2}"), findings: 1},
		{source: "function a(b) { b?.c?.()?.(); }", settings: complexityOf("{\"max\": 3}"), findings: 1},
		{source: "function a(b = '') {}", settings: complexityOf("{\"max\": 1}"), findings: 1},
		{source: "function a(b) { const { c = '' } = b; }", settings: complexityOf("{\"max\": 1}"), findings: 1},
		{source: "function a(b) { const [ c = '' ] = b; }", settings: complexityOf("{\"max\": 1}"), findings: 1},
		{source: "function a(b) { const [ { c: d = '' } = {} ] = b; }", settings: complexityOf("{\"max\": 1}"), findings: 1},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, Complexity, complexityFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, testCase.findings,
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		expected := make([]string, 0, testCase.findings)
		for range testCase.findings {
			expected = append(expected, messageComplexityComplex.Id)
		}
		rule_testing.ExpectFindings(t, result, expected...)
	}
}

// TestComplexityStaysSilent carries upstream's clean cases, which are the false positives it
// already thought about.
func TestComplexityStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
	}{
		{source: "function a(x) {}", settings: nil},
		{source: "function b(x) {}", settings: complexityOf("1")},
		{source: "function a(x) {if (true) {return x;}}", settings: complexityOf("2")},
		{source: "function a(x) {if (true) {return x;} else {return x+1;}}", settings: complexityOf("2")},
		{source: "function a(x) {if (true) {return x;} else if (false) {return x+1;} else {return 4;}}", settings: complexityOf("3")},
		{source: "function a(x) {for(var i = 0; i < 5; i ++) {x ++;} return x;}", settings: complexityOf("2")},
		{source: "function a(obj) {for(var i in obj) {obj[i] = 3;}}", settings: complexityOf("2")},
		{source: "function a(x) {for(var i = 0; i < 5; i ++) {if(i % 2 === 0) {x ++;}} return x;}", settings: complexityOf("3")},
		{source: "function a(obj) {if(obj){ for(var x in obj) {try {x.getThis();} catch (e) {x.getThat();}}} else {return false;}}", settings: complexityOf("4")},
		{source: "function a(x) {try {x.getThis();} catch (e) {x.getThat();}}", settings: complexityOf("2")},
		{source: "function a(x) {return x === 4 ? 3 : 5;}", settings: complexityOf("2")},
		{source: "function a(x) {return x === 4 ? 3 : (x === 3 ? 2 : 1);}", settings: complexityOf("3")},
		{source: "function a(x) {return x || 4;}", settings: complexityOf("2")},
		{source: "function a(x) {x && 4;}", settings: complexityOf("2")},
		{source: "function a(x) {x ?? 4;}", settings: complexityOf("2")},
		{source: "function a(x) {x ||= 4;}", settings: complexityOf("2")},
		{source: "function a(x) {x &&= 4;}", settings: complexityOf("2")},
		{source: "function a(x) {x ??= 4;}", settings: complexityOf("2")},
		{source: "function a(x) {x = 4;}", settings: complexityOf("1")},
		{source: "function a(x) {x |= 4;}", settings: complexityOf("1")},
		{source: "function a(x) {x &= 4;}", settings: complexityOf("1")},
		{source: "function a(x) {x += 4;}", settings: complexityOf("1")},
		{source: "function a(x) {x >>= 4;}", settings: complexityOf("1")},
		{source: "function a(x) {x >>>= 4;}", settings: complexityOf("1")},
		{source: "function a(x) {x == 4;}", settings: complexityOf("1")},
		{source: "function a(x) {x === 4;}", settings: complexityOf("1")},
		{source: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}", settings: complexityOf("3")},
		{source: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: if(x == 'foo') {5;};}}", settings: complexityOf("4")},
		{source: "function a(x) {while(true) {'foo';}}", settings: complexityOf("2")},
		{source: "function a(x) {do {'foo';} while (true)}", settings: complexityOf("2")},
		{source: "if (foo) { bar(); }", settings: complexityOf("3")},
		{source: "var a = (x) => {do {'foo';} while (true)}", settings: complexityOf("2")},
		{source: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: 3;}}", settings: complexityOf("{\"max\": 2, \"variant\": \"modified\"}")},
		{source: "function a(x) {switch(x){case 1: 1; break; case 2: 2; break; default: if(x == 'foo') {5;};}}", settings: complexityOf("{\"max\": 3, \"variant\": \"modified\"}")},
		{source: "function foo() { class C { x = a || b; y = c || d; } }", settings: complexityOf("2")},
		{source: "function foo() { class C { static x = a || b; static y = c || d; } }", settings: complexityOf("2")},
		{source: "function foo() { class C { x = a || b; y = c || d; } e || f; }", settings: complexityOf("2")},
		{source: "function foo() { a || b; class C { x = c || d; y = e || f; } }", settings: complexityOf("2")},
		{source: "function foo() { class C { [x || y] = a || b; } }", settings: complexityOf("2")},
		{source: "class C { x = a || b; y() { c || d; } z = e || f; }", settings: complexityOf("2")},
		{source: "class C { x() { a || b; } y = c || d; z() { e || f; } }", settings: complexityOf("2")},
		{source: "class C { x = (() => { a || b }) || (() => { c || d }) }", settings: complexityOf("2")},
		{source: "class C { x = () => { a || b }; y = () => { c || d } }", settings: complexityOf("2")},
		{source: "class C { x = a || (() => { b || c }); }", settings: complexityOf("2")},
		{source: "class C { x = class { y = a || b; z = c || d; }; }", settings: complexityOf("2")},
		{source: "class C { x = a || class { y = b || c; z = d || e; }; }", settings: complexityOf("2")},
		{source: "class C { x; y = a; static z; static q = b; }", settings: complexityOf("1")},
		{source: "function foo() { class C { static { a || b; } static { c || d; } } }", settings: complexityOf("2")},
		{source: "function foo() { a || b; class C { static { c || d; } } }", settings: complexityOf("2")},
		{source: "function foo() { class C { static { a || b; } } c || d; }", settings: complexityOf("2")},
		{source: "function foo() { class C { static { a || b; } } class D { static { c || d; } } }", settings: complexityOf("2")},
		{source: "class C { static { a || b; } static { c || d; } }", settings: complexityOf("2")},
		{source: "class C { static { a || b; } static { c || d; } static { e || f; } }", settings: complexityOf("2")},
		{source: "class C { static { () => a || b; c || d; } }", settings: complexityOf("2")},
		{source: "class C { static { a || b; () => c || d; } static { c || d; } }", settings: complexityOf("2")},
		{source: "class C { static { a } }", settings: complexityOf("1")},
		{source: "class C { static { a } static { b } }", settings: complexityOf("1")},
		{source: "class C { static { a || b; } } class D { static { c || d; } }", settings: complexityOf("2")},
		{source: "class C { static { a || b; } static c = d || e; }", settings: complexityOf("2")},
		{source: "class C { static a = b || c; static { c || d; } }", settings: complexityOf("2")},
		{source: "class C { static { a || b; } c = d || e; }", settings: complexityOf("2")},
		{source: "class C { a = b || c; static { d || e; } }", settings: complexityOf("2")},
		{source: "class C { static { a || b; c || d; } }", settings: complexityOf("3")},
		{source: "class C { static { if (a || b) c = d || e; } }", settings: complexityOf("4")},
		{source: "function b(x) {}", settings: complexityOf("{\"max\": 1}")},
		{source: "function a(b) { b?.c; }", settings: complexityOf("{\"max\": 2}")},
		{source: "function a(b = '') {}", settings: complexityOf("{\"max\": 2}")},
		{source: "function a(b) { const { c = '' } = b; }", settings: complexityOf("{\"max\": 2}")},
		{source: "function a(b) { const [ c = '' ] = b; }", settings: complexityOf("{\"max\": 2}")},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, Complexity, complexityFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != 0 {
			t.Errorf("%q with %+v: expected no findings, got %d: %v",
				testCase.source, testCase.settings, len(result.Diagnostics), result.MessageIds())
			continue
		}
		rule_testing.ExpectClean(t, result)
	}
}

// TestComplexityNamesAndSpans asserts every distinct NAME rendering and the span each finding
// carries, which TestComplexityFires cannot: this rule has one message id, so a count fixture is
// blind to both the name builder and the head-location logic, and those are most of the port.
//
// Upstream builds the name with `getFunctionNameWithKind` (77 lines) and the span with
// `getFunctionHeadLoc`. The corpus exercises 14 distinct renderings across 103 findings and every
// one is below, with the source that produces it taken from the corpus and the expected name, span
// and complexity read out of the installed ESLint 10.8.1 rule's own message.
//
// Three rows are worth reading twice, because the intuitive answer is wrong:
//
//	class C { x = () => a || b || c; }        renders "Method 'x'", NOT "Arrow function"
//	const obj = { b: (a) => a?.b?.c }         renders "Method 'b'", NOT "Arrow function"
//	class C { static { } static(){ } }        a method named `static` is "Method 'static'",
//	                                          while `static static(){}` is "Static method 'static'"
//
// The first two are upstream naming a function by what it is ASSIGNED to rather than by its own
// syntax, and the third is the modifier and the name being the same word.
func TestComplexityNamesAndSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		// span is the text the finding must cover: the function HEAD, never its body.
		span string
		// message is the whole computed sentence up to the threshold clause, so both the name and
		// the counted complexity are asserted.
		message string
	}{
		{
			source: "function a(x) {}", settings: complexityOf("0"),
			span: "function a", message: "Function 'a' has a complexity of 1.",
		},
		{
			source: "var func = function () {}", settings: complexityOf("0"),
			span: "function ", message: "Function has a complexity of 1.",
		},
		{
			source: "var obj = { a(x) {} }", settings: complexityOf("0"),
			span: "a", message: "Method 'a' has a complexity of 1.",
		},
		{
			source: "var a = (x) => {if (true) {return x;}}", settings: complexityOf("1"),
			span: "=>", message: "Arrow function has a complexity of 2.",
		},
		{
			source: "class C { x(){ a || b; } y = c || d || e; z() { f || g; } }", settings: complexityOf("2"),
			span: "c || d || e", message: "Class field initializer has a complexity of 3.",
		},
		{
			source: "class C { x = a || b; y() { c || d || e; } z = f || g; }", settings: complexityOf("2"),
			span: "y", message: "Method 'y' has a complexity of 3.",
		},
		{
			// An arrow in a field position is named for the FIELD, not as an arrow function.
			source: "class C { x = () => a || b || c; }", settings: complexityOf("2"),
			span: "x = ", message: "Method 'x' has a complexity of 3.",
		},
		{
			source: "class C { static { a || b; }  }", settings: complexityOf("1"),
			span: "static", message: "Class static block has a complexity of 2.",
		},
		{
			// A method whose NAME is `static`, with no static modifier.
			source:   "class C { static { a || b; } static(){ c || d || e; } static { f || g; } }",
			settings: complexityOf("2"),
			span:     "static", message: "Method 'static' has a complexity of 3.",
		},
		{
			// The same name WITH the modifier, so both words appear.
			source:   "class C { static { a || b; } static static(){ c || d || e; } static { f || g; } }",
			settings: complexityOf("2"),
			span:     "static static", message: "Static method 'static' has a complexity of 3.",
		},
		{
			// An arrow as an object property value, named for the property.
			source: "const obj = { b: (a) => a?.b?.c };", settings: complexityOf(`{"max": 2}`),
			span: "b: ", message: "Method 'b' has a complexity of 3.",
		},
		{
			source: "const obj = { c: function (a) { return a?.b?.c; } };", settings: complexityOf(`{"max": 2}`),
			span: "c: function ", message: "Method 'c' has a complexity of 3.",
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, Complexity, complexityFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != 1 {
			t.Errorf("%q: expected exactly 1 finding, got %d %v",
				testCase.source, len(result.Diagnostics), result.MessageIds())
			continue
		}
		diagnostic := result.Diagnostics[0]
		reported := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != testCase.span {
			t.Errorf("%q: expected the span to cover %q, got %q", testCase.source, testCase.span, reported)
		}
		if !strings.HasPrefix(diagnostic.Message.Description, testCase.message) {
			t.Errorf("%q: expected the message to open with %q, got %q",
				testCase.source, testCase.message, diagnostic.Message.Description)
		}
	}
}

// TestComplexityNeverReportsTheProgram covers what upstream's corpus cannot.
//
// Upstream's `onCodePathEnd` filters `codePath.origin` to `function`, `class-field-initializer` and
// `class-static-block`, so the `program` path is counted and never reported. Giving the file a frame
// here would report the whole module as one enormous function, and a mutation doing exactly that
// survived all 165 imported fixtures -- upstream's corpus never puts enough top-level branching in
// one case to cross a threshold, so the difference is invisible to it.
//
// Every verdict is what the installed ESLint 10.8.1 rule answered at a threshold of 1:
//
//	if (a) {} if (b) {} if (c) {}           clean -- three top-level branches, never reported
//	a || b || c || d;                       clean
//	if (a) {} function f() { if (b) {} }    1 finding, on `f` with a complexity of 2
//
// The third row is the control, and it is the one that makes the first two mean something: it says
// the rule is reporting at all on this input, and that `f` does NOT inherit the top-level `if`.
func TestComplexityNeverReportsTheProgram(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		findings int
		message  string
	}{
		{source: "if (a) {} if (b) {} if (c) {}", findings: 0},
		{source: "a || b || c || d;", findings: 0},
		{source: "for (;;) {} while (a) {} do {} while (b);", findings: 0},
		{
			source: "if (a) {} function f() { if (b) {} }", findings: 1,
			message: "Function 'f' has a complexity of 2.",
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, Complexity, complexityFile, testCase.source, complexityOf("1"))
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q: expected %d findings, got %d %v",
				testCase.source, testCase.findings, len(result.Diagnostics), result.MessageIds())
			continue
		}
		if testCase.message == "" {
			continue
		}
		if !strings.HasPrefix(result.Diagnostics[0].Message.Description, testCase.message) {
			t.Errorf("%q: expected the message to open with %q, got %q",
				testCase.source, testCase.message, result.Diagnostics[0].Message.Description)
		}
	}
}

// TestComplexityNamesModifiersTheCorpusNeverCrosses covers seven more renderings that upstream's
// corpus cannot reach.
//
// The name builder assembles `static`, `private`, `async` and `generator` before the kind word, and
// the corpus never writes a generator or an async function whose complexity crosses a threshold --
// so a mutation dropping the generator word survived all 165 imported fixtures. These rows exercise
// every modifier and every combination the builder can produce.
//
// Two are not guessable and were measured rather than derived:
//
//	class C { #p() { a || b; } }        renders "Private method #p" -- the private name is
//	                                    UNQUOTED, where every other name is quoted
//	async function* i() {}              renders "Async generator function 'i'", so the modifier
//	                                    order is static, private, async, generator, then the kind
//
// Every expected name and span is what the installed ESLint 10.8.1 rule produced at a threshold
// of 1.
func TestComplexityNamesModifiersTheCorpusNeverCrosses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source  string
		span    string
		message string
	}{
		{
			source: "function* g() { a || b; }",
			span:   "function* g", message: "Generator function 'g' has a complexity of 2.",
		},
		{
			source: "async function h() { a || b; }",
			span:   "async function h", message: "Async function 'h' has a complexity of 2.",
		},
		{
			source: "async function* i() { a || b; }",
			span:   "async function* i", message: "Async generator function 'i' has a complexity of 2.",
		},
		{
			source: "class C { async *m() { a || b; } }",
			span:   "async *m", message: "Async generator method 'm' has a complexity of 2.",
		},
		{
			source: "class C { static async *m() { a || b; } }",
			span:   "static async *m", message: "Static async generator method 'm' has a complexity of 2.",
		},
		{
			// A private name is rendered WITHOUT quotes, unlike every other name.
			source: "class C { #p() { a || b; } }",
			span:   "#p", message: "Private method #p has a complexity of 2.",
		},
		{
			source: "class C { static #p() { a || b; } }",
			span:   "static #p", message: "Static private method #p has a complexity of 2.",
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, Complexity, complexityFile, testCase.source, complexityOf("1"))
		if len(result.Diagnostics) != 1 {
			t.Errorf("%q: expected exactly 1 finding, got %d %v",
				testCase.source, len(result.Diagnostics), result.MessageIds())
			continue
		}
		diagnostic := result.Diagnostics[0]
		reported := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != testCase.span {
			t.Errorf("%q: expected the span to cover %q, got %q", testCase.source, testCase.span, reported)
		}
		if !strings.HasPrefix(diagnostic.Message.Description, testCase.message) {
			t.Errorf("%q: expected the message to open with %q, got %q",
				testCase.source, testCase.message, diagnostic.Message.Description)
		}
	}
}

// TestComplexityDecoderReadsEveryOptionShape pins the option surface, including the JavaScript
// quirk this rule shares with max-depth.
//
// Upstream's threshold handling gates on PRESENCE (`hasOwn`) and picks by TRUTHINESS (`||`), and
// the two disagree exactly when a zero is written, so the two zero spellings do opposite things.
// The table below is the same one `DecodeMaxDepthOptions` carries, verified against the installed
// complexity rule rather than assumed to transfer.
func TestComplexityDecoderReadsEveryOptionShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		// optionJson is exactly the bytes the config layer hands the decoder.
		optionJson string
		maximum    int
		variant    ComplexityVariant
		// silent marks the shape that leaves upstream comparing against `undefined`.
		silent bool
	}{
		{optionJson: `2`, maximum: 2, variant: ComplexityClassic},
		{optionJson: `0`, maximum: 0, variant: ComplexityClassic},
		{optionJson: `{"max": 2}`, maximum: 2, variant: ComplexityClassic},
		{optionJson: `{"maximum": 2}`, maximum: 2, variant: ComplexityClassic},
		// Neither key present, so the branch is never entered and the default of 20 stands.
		{optionJson: `{}`, maximum: 20, variant: ComplexityClassic},
		{optionJson: `{"variant": "modified"}`, maximum: 20, variant: ComplexityModified},
		{optionJson: `{"max": 2, "variant": "modified"}`, maximum: 2, variant: ComplexityModified},
		// `undefined || 0` is 0, so a lone `max` of zero wins and reports every function.
		{optionJson: `{"max": 0}`, maximum: 0, variant: ComplexityClassic},
		// `0 || undefined` is undefined, so every comparison is false and the rule goes silent.
		{optionJson: `{"maximum": 0}`, silent: true, variant: ComplexityClassic},
		// The `||` consults maximum first, so a truthy maximum wins over any max.
		{optionJson: `{"max": 5, "maximum": 2}`, maximum: 2, variant: ComplexityClassic},
		{optionJson: `{"maximum": 5, "max": 2}`, maximum: 5, variant: ComplexityClassic},
	}

	for _, testCase := range cases {
		decoded, err := DecodeComplexityOptions([]byte(testCase.optionJson))
		if err != nil {
			t.Errorf("decoding %s: %v", testCase.optionJson, err)
			continue
		}
		settings, ok := decoded.(ComplexitySettings)
		if !ok {
			t.Errorf("decoding %s: expected ComplexitySettings, got %T", testCase.optionJson, decoded)
			continue
		}
		if settings.Variant != testCase.variant {
			t.Errorf("decoding %s: expected variant %q, got %q",
				testCase.optionJson, testCase.variant, settings.Variant)
		}
		if testCase.silent {
			if settings.Maximum != complexityNeverReports {
				t.Errorf("decoding %s: expected the silencing threshold, got %d",
					testCase.optionJson, settings.Maximum)
			}
			continue
		}
		if settings.Maximum != testCase.maximum {
			t.Errorf("decoding %s: expected a threshold of %d, got %d",
				testCase.optionJson, testCase.maximum, settings.Maximum)
		}
	}

	// End to end, because a decoder row is not a verdict. The silencing shape must report NOTHING
	// on source that a threshold of 0 reports on.
	branchy := "function f() { if (a) {} if (b) {} }"
	result := rule_testing.RunWithOptions(t, Complexity, complexityFile, branchy, complexityOf(`{"maximum": 0}`))
	rule_testing.ExpectClean(t, result)

	// The control: the sibling zero spelling reports, so the silence above is the option rather
	// than a rule that cannot fire on this input.
	result = rule_testing.RunWithOptions(t, Complexity, complexityFile, branchy, complexityOf(`{"max": 0}`))
	if len(result.Diagnostics) != 1 {
		t.Errorf("expected the control to report, got %d: %v",
			len(result.Diagnostics), result.MessageIds())
	}

	// A rule handed nil options enforces upstream's default of 20 rather than doing nothing, so a
	// two-branch function is clean and a deliberately deep one is not.
	result = rule_testing.RunWithOptions(t, Complexity, complexityFile, branchy, nil)
	rule_testing.ExpectClean(t, result)

	// The variant, end to end: a switch with three cases is complexity 4 classic and 2 modified.
	switchy := "function f() { switch (a) { case 1: break; case 2: break; case 3: break; } }"
	result = rule_testing.RunWithOptions(t, Complexity, complexityFile, switchy, complexityOf("3"))
	if len(result.Diagnostics) != 1 {
		t.Errorf("expected classic counting to report at threshold 3, got %d", len(result.Diagnostics))
	}
	result = rule_testing.RunWithOptions(t, Complexity, complexityFile, switchy,
		complexityOf(`{"max": 3, "variant": "modified"}`))
	rule_testing.ExpectClean(t, result)
}

// TestComplexityNameBuilderBranches tests `complexityNameOf` DIRECTLY rather than through the
// rule's messages.
//
// The two tables above assert renderings end to end, and they are the ones that caught three real
// defects. But this rule has a single message id and a 20-branch name builder, so when an end-to-end
// row fails it says the rendering is wrong without saying which branch produced it. Calling the
// builder on a located node names the branch.
//
// Each row is one branch of the builder, and the expected string is what the installed
// ESLint 10.8.1 rule rendered for that shape. The `find` field locates the node to name, so a row
// tests one branch rather than whatever the walk reached first.
func TestComplexityNameBuilderBranches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		// find selects which function-like node to name.
		find func(node *ast.Node) bool
		want string
	}{
		{
			source: "function a() {}",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindFunctionDeclaration },
			want:   "Function 'a'",
		},
		{
			source: "var f = function () {};",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindFunctionExpression },
			want:   "Function",
		},
		{
			source: "function* g() {}",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindFunctionDeclaration },
			want:   "Generator function 'g'",
		},
		{
			source: "async function h() {}",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindFunctionDeclaration },
			want:   "Async function 'h'",
		},
		{
			source: "async function* i() {}",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindFunctionDeclaration },
			want:   "Async generator function 'i'",
		},
		{
			source: "var a = () => {};",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindArrowFunction },
			want:   "Arrow function",
		},
		{
			source: "class C { m() {} }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindMethodDeclaration },
			want:   "Method 'm'",
		},
		{
			source: "class C { static m() {} }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindMethodDeclaration },
			want:   "Static method 'm'",
		},
		{
			source: "class C { get g() {} }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindGetAccessor },
			want:   "Getter 'g'",
		},
		{
			source: "class C { set s(v) {} }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindSetAccessor },
			want:   "Setter 's'",
		},
		{
			source: "class C { constructor() {} }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindConstructor },
			want:   "Constructor",
		},
		{
			// The private name is unquoted, unlike every other name.
			source: "class C { #p() {} }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindMethodDeclaration },
			want:   "Private method #p",
		},
		{
			source: "class C { static #p() {} }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindMethodDeclaration },
			want:   "Static private method #p",
		},
		{
			source: "class C { static { } }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindClassStaticBlockDeclaration },
			want:   "Class static block",
		},
		{
			source: "class C { x = 1; }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindPropertyDeclaration },
			want:   "Class field initializer",
		},
		{
			// A function that BORROWS its name from what it is assigned to, which is the branch
			// three end-to-end defects came from.
			source: "class C { x = () => 1; }",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindArrowFunction },
			want:   "Method 'x'",
		},
		{
			source: "var o = { b: (a) => 1 };",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindArrowFunction },
			want:   "Method 'b'",
		},
		{
			source: "var o = { c: function () {} };",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindFunctionExpression },
			want:   "Method 'c'",
		},
		{
			// A named function expression in a property position is named for the PROPERTY, not
			// for itself: upstream tries the property name first and falls back to the function's
			// own id only when the property name is not static. Measured against the installed
			// rule, which renders "Method 'd'" here.
			source: "var o = { d: function named() {} };",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindFunctionExpression },
			want:   "Method 'd'",
		},
		{
			source: "var o = { e() {} };",
			find:   func(n *ast.Node) bool { return n.Kind == ast.KindMethodDeclaration },
			want:   "Method 'e'",
		},
	}

	for _, testCase := range cases {
		var located *ast.Node
		probe := rule.Rule{
			Name: "probe/complexity-name",
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindSourceFile: func(file *ast.Node) {
						var walk func(*ast.Node)
						walk = func(node *ast.Node) {
							if node == nil || located != nil {
								return
							}
							if testCase.find(node) {
								located = node
								return
							}
							node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
						}
						walk(file)
						if located != nil {
							if got := complexityNameOf(ctx, located); got != testCase.want {
								t.Errorf("%q: expected the name builder to render %q, got %q",
									testCase.source, testCase.want, got)
							}
						}
					},
				}
			},
		}
		rule_testing.Run(t, probe, complexityFile, testCase.source)
		if located == nil {
			t.Errorf("%q: the probe found no node to name, so this row proved nothing", testCase.source)
		}
	}
}
