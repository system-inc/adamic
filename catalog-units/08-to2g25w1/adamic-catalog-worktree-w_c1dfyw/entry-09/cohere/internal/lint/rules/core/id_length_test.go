package core

import (
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// idLengthFile is where the fixtures pretend to live.
const idLengthFile = "/repository/source/IdLength.ts"

// idLengthOf builds the settings a config carrying this option object would decode to.
//
// The argument is the option as JSON TEXT, exactly the bytes the config layer hands the decoder, so
// a fixture cannot bypass the decoder or test a shape a config could not deliver. That is section
// 7c of the standard applied at the fixture rather than only in a contract test.
func idLengthOf(optionJson string) any {
	settings, err := DecodeIdLengthOptions([]byte(optionJson))
	if err != nil {
		panic(err)
	}
	return settings
}

// The corpus is ESLint's own, taken from
// /tmp/lint-sources-fresh/eslint/tests/lib/rules/id-length.js by RUNNING that file with RuleTester
// intercepted. 181 cases were captured from the single RuleTester.run call.
//
// Every expectation is what the INSTALLED ESLint 10.8.1 rule answered when driven over that case
// through the Linter interface under `sourceType: module` with the typescript-eslint parser. The
// oracle reproduced all 181 of the corpus's own declared verdicts under upstream's own
// languageOptions, and ZERO cases change verdict between that configuration and cohere's.
//
// 35 of these cases carry non-ASCII identifiers, which is what makes the grapheme count testable:
// see text.GraphemeCount, which is scored against Intl.Segmenter.
//
//	84 reporting cases, 97 clean cases.
func TestIdLengthFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		// settings is what the rule's own decoder produces for this case's options.
		settings any
		// findings is one message id per expected finding, taken from what the installed rule
		// produced: this rule has four ids and no count could tell them apart.
		findings []string
	}{
		{source: "var x = 1;", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var x;", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "obj.e = document.body;", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function x() {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function xyz(a) {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var obj = { a: 1, bc: 2 };", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "try { blah(); } catch (e) { /* pass */ }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var handler = function (e) {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "for (var i=0; i < 10; i++) { console.log(i); }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var j=0; while (j > -10) { console.log(--j); }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var [i] = arr;", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var [,i,a] = arr;", settings: nil, findings: []string{messageIdLengthTooShort.Id, messageIdLengthTooShort.Id}},
		{source: "function foo([a]) {}", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "import x from 'module';", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "import { x as z } from 'module';", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "import { foo as z } from 'module';", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "import { 'foo' as z } from 'module';", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "import * as x from 'module';", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "import longName from 'module';", settings: idLengthOf("{\"max\": 5}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "import * as longName from 'module';", settings: idLengthOf("{\"max\": 5}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "import { foo as longName } from 'module';", settings: idLengthOf("{\"max\": 5}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "var _$xt_$ = Foo(42)", settings: idLengthOf("{\"min\": 2, \"max\": 4}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "var _$x$_t$ = Foo(42)", settings: idLengthOf("{\"min\": 2, \"max\": 4}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "var toString;", settings: idLengthOf("{\"max\": 5}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "(a) => { a * a };", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function foo(x = 0) { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "class x { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "class Foo { x() {} }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function foo(...x) { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function foo({x}) { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function foo({x: a}) { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function foo({x: a, longName}) { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function foo({ longName: a }) {}", settings: idLengthOf("{\"min\": 3, \"max\": 5}"), findings: []string{messageIdLengthTooShort.Id}},
		{source: "function foo({ prop: longName }) {};", settings: idLengthOf("{\"min\": 3, \"max\": 5}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "function foo({ a: b }) {};", settings: idLengthOf("{\"exceptions\": [\"a\"]}"), findings: []string{messageIdLengthTooShort.Id}},
		{source: "var hasOwnProperty;", settings: idLengthOf("{\"max\": 10, \"exceptions\": []}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "function foo({ a: { b: { c: d, e } } }) { }", settings: nil, findings: []string{messageIdLengthTooShort.Id, messageIdLengthTooShort.Id}},
		{source: "var { x} = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { x: a} = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { a: a} = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { prop: a } = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { longName: a } = {};", settings: idLengthOf("{\"min\": 3, \"max\": 5}"), findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { prop: [x] } = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { prop: [[x]] } = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { prop: longName } = {};", settings: idLengthOf("{\"min\": 3, \"max\": 5}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "var { x: a} = {};", settings: idLengthOf("{\"exceptions\": [\"x\"]}"), findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { a: { b: { c: d } } } = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { a: { b: { c: d, e } } } = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id, messageIdLengthTooShort.Id}},
		{source: "var { a: { b: { c, e: longName } } } = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { a: { b: { c: d, e: longName } } } = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { a, b: { c: d, e: longName } } = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id, messageIdLengthTooShort.Id}},
		{source: "import x from 'y';", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "export var x = 0;", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "({ a: obj.x.y.z } = {});", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "({ prop: obj.x } = {});", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var x = 1;", settings: idLengthOf("{\"properties\": \"never\"}"), findings: []string{messageIdLengthTooShort.Id}},
		{source: "var {prop: x} = foo;", settings: idLengthOf("{\"properties\": \"never\"}"), findings: []string{messageIdLengthTooShort.Id}},
		{source: "var foo = {x: prop};", settings: idLengthOf("{\"properties\": \"always\"}"), findings: []string{messageIdLengthTooShort.Id}},
		{source: "function BEFORE_send() {};", settings: idLengthOf("{\"min\": 3, \"max\": 5}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "function NOTMATCHED_send() {};", settings: idLengthOf("{\"min\": 3, \"max\": 5, \"exceptionPatterns\": [\"^BEFORE_\"]}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "function N() {};", settings: idLengthOf("{\"min\": 3, \"max\": 5, \"exceptionPatterns\": [\"^BEFORE_\"]}"), findings: []string{messageIdLengthTooShort.Id}},
		{source: "class Foo { #x() {} }", settings: nil, findings: []string{messageIdLengthTooShortPrivate.Id}},
		{source: "class Foo { x = 1 }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "class Foo { #x = 1 }", settings: nil, findings: []string{messageIdLengthTooShortPrivate.Id}},
		{source: "class Foo { #abcdefg() {} }", settings: idLengthOf("{\"max\": 3}"), findings: []string{messageIdLengthTooLongPrivate.Id}},
		{source: "class Foo { abcdefg = 1 }", settings: idLengthOf("{\"max\": 3}"), findings: []string{messageIdLengthTooLong.Id}},
		{source: "class Foo { #abcdefg = 1 }", settings: idLengthOf("{\"max\": 3}"), findings: []string{messageIdLengthTooLongPrivate.Id}},
		{source: "var \U00020b9f = 2", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var \u845b\U000e0100 = 2", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var myObj = { \U00010318: 1 };", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "(\U00010318) => { \U00010318 * \U00010318 };", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "class \U00020b9f { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "class Foo { \U00010318() {} }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "class Foo1 { #\U00010318() {} }", settings: nil, findings: []string{messageIdLengthTooShortPrivate.Id}},
		{source: "class Foo2 { \U00010318 = 1 }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "class Foo3 { #\U00010318 = 1 }", settings: nil, findings: []string{messageIdLengthTooShortPrivate.Id}},
		{source: "function foo1(...\U00010318) { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function foo([\U00010318]) { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var [ \U00010318 ] = arr;", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { prop: [\U00010318]} = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "function foo({\U00010318}) { }", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { \U00010318 } = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "var { prop: \U00010318} = {};", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
		{source: "({ prop: obj.\U00010318 } = {});", settings: nil, findings: []string{messageIdLengthTooShort.Id}},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, IdLength, idLengthFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != len(testCase.findings) {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, len(testCase.findings),
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		rule_testing.ExpectFindings(t, result, testCase.findings...)
	}
}

// TestIdLengthStaysSilent carries upstream's clean cases, which are the false positives it already
// thought about. Each one is a class this port would otherwise ship.
func TestIdLengthStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
	}{
		{source: "var xyz;", settings: nil},
		{source: "var xy = 1;", settings: nil},
		{source: "function xyz() {};", settings: nil},
		{source: "function xyz(abc, de) {};", settings: nil},
		{source: "var obj = { abc: 1, de: 2 };", settings: nil},
		{source: "var obj = { 'a': 1, bc: 2 };", settings: nil},
		{source: "var obj = {}; obj['a'] = 2;", settings: nil},
		{source: "abc = d;", settings: nil},
		{source: "try { blah(); } catch (err) { /* pass */ }", settings: nil},
		{source: "var handler = function ($e) {};", settings: nil},
		{source: "var _a = 2", settings: nil},
		{source: "var _ad$$ = new $;", settings: nil},
		{source: "var xyz = new \u03a3\u03a3();", settings: nil},
		{source: "unrelatedExpressionThatNeedsToBeIgnored();", settings: nil},
		{source: "var obj = { 'a': 1, bc: 2 }; obj.tk = obj.a;", settings: nil},
		{source: "var query = location.query.q || '';", settings: nil},
		{source: "var query = location.query.q ? location.query.q : ''", settings: nil},
		{source: "let {a: foo} = bar;", settings: nil},
		{source: "let foo = { [a]: 1 };", settings: nil},
		{source: "let foo = { [a + b]: 1 };", settings: nil},
		{source: "var x = Foo(42)", settings: idLengthOf("{\"min\": 1}")},
		{source: "var x = Foo(42)", settings: idLengthOf("{\"min\": 0}")},
		{source: "foo.$x = Foo(42)", settings: idLengthOf("{\"min\": 1}")},
		{source: "var lalala = Foo(42)", settings: idLengthOf("{\"max\": 6}")},
		{source: "for (var q, h=0; h < 10; h++) { console.log(h); q++; }", settings: idLengthOf("{\"exceptions\": [\"h\", \"q\"]}")},
		{source: "(num) => { num * num };", settings: nil},
		{source: "function foo(num = 0) { }", settings: nil},
		{source: "class MyClass { }", settings: nil},
		{source: "class Foo { method() {} }", settings: nil},
		{source: "function foo(...args) { }", settings: nil},
		{source: "var { prop } = {};", settings: nil},
		{source: "var { [a]: prop } = {};", settings: nil},
		{source: "var { a: foo } = {};", settings: idLengthOf("{\"min\": 3}")},
		{source: "var { prop: foo } = {};", settings: idLengthOf("{\"max\": 3}")},
		{source: "var { longName: foo } = {};", settings: idLengthOf("{\"min\": 3, \"max\": 5}")},
		{source: "var { foo: a } = {};", settings: idLengthOf("{\"exceptions\": [\"a\"]}")},
		{source: "var { a: { b: { c: longName } } } = {};", settings: nil},
		{source: "({ a: obj.x.y.z } = {});", settings: idLengthOf("{\"properties\": \"never\"}")},
		{source: "import something from 'y';", settings: nil},
		{source: "export var num = 0;", settings: nil},
		{source: "import * as something from 'y';", settings: nil},
		{source: "import { x } from 'y';", settings: nil},
		{source: "import { x as x } from 'y';", settings: nil},
		{source: "import { 'x' as x } from 'y';", settings: nil},
		{source: "import { x as foo } from 'y';", settings: nil},
		{source: "import { longName } from 'y';", settings: idLengthOf("{\"max\": 5}")},
		{source: "import { x as bar } from 'y';", settings: idLengthOf("{\"max\": 5}")},
		{source: "({ prop: obj.x.y.something } = {});", settings: nil},
		{source: "({ prop: obj.longName } = {});", settings: nil},
		{source: "var obj = { a: 1, bc: 2 };", settings: idLengthOf("{\"properties\": \"never\"}")},
		{source: "var obj = { [a]: 2 };", settings: idLengthOf("{\"properties\": \"never\"}")},
		{source: "var obj = {}; obj.a = 1; obj.bc = 2;", settings: idLengthOf("{\"properties\": \"never\"}")},
		{source: "({ prop: obj.x } = {});", settings: idLengthOf("{\"properties\": \"never\"}")},
		{source: "var obj = { aaaaa: 1 };", settings: idLengthOf("{\"max\": 4, \"properties\": \"never\"}")},
		{source: "var obj = {}; obj.aaaaa = 1;", settings: idLengthOf("{\"max\": 4, \"properties\": \"never\"}")},
		{source: "({ a: obj.x.y.z } = {});", settings: idLengthOf("{\"max\": 4, \"properties\": \"never\"}")},
		{source: "({ prop: obj.xxxxx } = {});", settings: idLengthOf("{\"max\": 4, \"properties\": \"never\"}")},
		{source: "var arr = [i,j,f,b]", settings: nil},
		{source: "function foo([arr]) {}", settings: nil},
		{source: "var {x} = foo;", settings: idLengthOf("{\"properties\": \"never\"}")},
		{source: "var {x, y: {z}} = foo;", settings: idLengthOf("{\"properties\": \"never\"}")},
		{source: "let foo = { [a]: 1 };", settings: idLengthOf("{\"properties\": \"always\"}")},
		{source: "let foo = { [a + b]: 1 };", settings: idLengthOf("{\"properties\": \"always\"}")},
		{source: "function BEFORE_send() {};", settings: idLengthOf("{\"min\": 3, \"max\": 5, \"exceptionPatterns\": [\"^BEFORE_\"]}")},
		{source: "function BEFORE_send() {};", settings: idLengthOf("{\"min\": 3, \"max\": 5, \"exceptionPatterns\": [\"^BEFORE_\", \"send$\"]}")},
		{source: "function BEFORE_send() {};", settings: idLengthOf("{\"min\": 3, \"max\": 5, \"exceptionPatterns\": [\"^BEFORE_\", \"^A\", \"^Z\"]}")},
		{source: "function BEFORE_send() {};", settings: idLengthOf("{\"min\": 3, \"max\": 5, \"exceptionPatterns\": [\"^A\", \"^BEFORE_\", \"^Z\"]}")},
		{source: "var x = 1 ;", settings: idLengthOf("{\"min\": 3, \"max\": 5, \"exceptionPatterns\": [\"[x-z]\"]}")},
		{source: "class Foo { #xyz() {} }", settings: nil},
		{source: "class Foo { xyz = 1 }", settings: nil},
		{source: "class Foo { #xyz = 1 }", settings: nil},
		{source: "class Foo { #abc() {} }", settings: idLengthOf("{\"max\": 3}")},
		{source: "class Foo { abc = 1 }", settings: idLengthOf("{\"max\": 3}")},
		{source: "class Foo { #abc = 1 }", settings: idLengthOf("{\"max\": 3}")},
		{source: "var \U00020b9f = 2", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "var \u845b\U000e0100 = 2", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "var a = { \U00010318: 1 };", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "(\U00010318) => { \U00010318 * \U00010318 };", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "class \U00020b9f { }", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "class F { \U00010318() {} }", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "class F { #\U00010318() {} }", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "class F { \U00010318 = 1 }", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "class F { #\U00010318 = 1 }", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "function f(...\U00010318) { }", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "function f([\U00010318]) { }", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "var [ \U00010318 ] = a;", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "var { p: [\U00010318]} = {};", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "function f({\U00010318}) { }", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "var { \U00010318 } = {};", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "var { p: \U00010318} = {};", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "({ prop: o.\U00010318 } = {});", settings: idLengthOf("{\"min\": 1, \"max\": 1}")},
		{source: "import foo from 'foo.json' with { type: 'json' }", settings: idLengthOf("{\"min\": 1, \"max\": 3, \"properties\": \"always\"}")},
		{source: "export * from 'foo.json' with { type: 'json' }", settings: idLengthOf("{\"min\": 1, \"max\": 3, \"properties\": \"always\"}")},
		{source: "export { default } from 'foo.json' with { type: 'json' }", settings: idLengthOf("{\"min\": 1, \"max\": 3, \"properties\": \"always\"}")},
		{source: "import('foo.json', { with: { type: 'json' } })", settings: idLengthOf("{\"min\": 1, \"max\": 3, \"properties\": \"always\"}")},
		{source: "import('foo.json', { 'with': { type: 'json' } })", settings: idLengthOf("{\"min\": 1, \"max\": 3, \"properties\": \"always\"}")},
		{source: "import('foo.json', { with: { type } })", settings: idLengthOf("{\"min\": 1, \"max\": 3, \"properties\": \"always\"}")},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, IdLength, idLengthFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != 0 {
			t.Errorf("%q with %+v: expected no findings, got %d: %v",
				testCase.source, testCase.settings, len(result.Diagnostics), result.MessageIds())
			continue
		}
		rule_testing.ExpectClean(t, result)
	}
}

// TestIdLengthCountsGraphemesNotRunes covers what upstream's corpus cannot.
//
// The count skips two classes of rune that continue a cluster rather than starting one: combining
// marks, and Hangul conjoining vowel and trail jamo. The corpus reaches the first -- `葛󠄀` is a base
// plus a variation selector, and a mutation disabling the mark skip is caught by it. It cannot
// reach the second: upstream writes no Hangul, so a mutation deleting the jamo skip survived all
// 181 imported fixtures.
//
// Decomposed Hangul is the distinguishing shape and it is not exotic: `각` written as three
// conjoining jamo is ONE grapheme and three runes, it carries no combining-mark property, and the
// parser accepts it as an identifier. Every verdict below is what the installed ESLint 10.8.1 rule
// answered:
//
//	var 각 = 1;                    min 2         tooShort  (1 grapheme)
//	var 각 = 1;                    min 1, max 1  clean
//	var 각 = 1;                                min 1, max 1  clean     (precomposed)
//	var 각각 = 1;  min 1, max 1  tooLong   (2 graphemes)
//
// The second row is the one that fails without the jamo skip: three runes exceed a maximum of one.
// The last row is the control -- it says the skip is not simply "count Hangul as one", since two
// decomposed syllables still count as two.
//
// The source LITERALS are written as escapes rather than raw characters, so nothing in the editing
// path can normalise a decomposed sequence into its precomposed form -- which would leave the rows
// looking identical while silently destroying what they test. The glyphs above appear raw in this
// comment only, for readability; the first draft of this file wrote them raw in the literals too,
// which is how it got checked.
func TestIdLengthCountsGraphemesNotRunes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		findings []string
	}{
		{
			source:   "var \u1100\u1161\u11a8 = 1;",
			settings: idLengthOf(`{"min": 2}`),
			findings: []string{messageIdLengthTooShort.Id},
		},
		{
			source:   "var \u1100\u1161\u11a8 = 1;",
			settings: idLengthOf(`{"min": 1, "max": 1}`),
			findings: nil,
		},
		{
			source:   "var \uac01 = 1;",
			settings: idLengthOf(`{"min": 1, "max": 1}`),
			findings: nil,
		},
		{
			source:   "var \uac01 = 1;",
			settings: idLengthOf(`{"min": 2}`),
			findings: []string{messageIdLengthTooShort.Id},
		},
		{
			source:   "var \u1100\u1161\u11a8\u1100\u1161\u11a8 = 1;",
			settings: idLengthOf(`{"min": 1, "max": 1}`),
			findings: []string{messageIdLengthTooLong.Id},
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, IdLength, idLengthFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != len(testCase.findings) {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, len(testCase.findings),
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		if len(testCase.findings) > 0 {
			rule_testing.ExpectFindings(t, result, testCase.findings...)
		}
	}
}

// TestIdLengthDestructuringArmsCoverWhatTheCorpusDoesNot pins two arms of the naming-site table
// that survived a mutation sweep against all 181 imported fixtures.
//
// Both are cases where upstream's corpus happens not to combine the shape with the option that
// would reveal it, so deleting the arm changes nothing any imported case can see.
//
// **The destructuring key.** In `var { a: b } = x` only `b` is a name this file chose; `a` is the
// source object's own property. The corpus never writes a key that is too SHORT beside a binding
// that is fine, so both survive a rule that checks the key. Measured against the installed rule at
// min 2:
//
//	var { a: longEnough } = x;   clean     -- the short key is not ours
//	var { a: b } = x;            1, on `b` -- only the binding
//
// **An array pattern with `properties: "never"`.** Upstream lists `ArrayPattern: true`,
// unconditional, so a positional element is not a property and the option does not gate it. Every
// array-pattern case in the corpus runs with properties left at its default, so an arm that gated
// them would pass all of them. Measured:
//
//	var [ a ] = x;         properties never   1 finding
//	function f([a]) {}     properties never   2 findings, on `f` and `a`
//	var { p: [a] } = x;    properties never   1 finding, on `a`
func TestIdLengthDestructuringArmsCoverWhatTheCorpusDoesNot(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		findings int
		// reported names the identifier the single finding must cover, so a rule reporting the KEY
		// instead of the binding fails even where the count matches.
		reported string
	}{
		{source: "var { a: longEnough } = x;", settings: idLengthOf(`{"min": 2}`), findings: 0},
		{source: "var { a: b } = x;", settings: idLengthOf(`{"min": 2}`), findings: 1, reported: "b"},
		{source: "var [ a ] = x;", settings: idLengthOf(`{"min": 2, "properties": "never"}`), findings: 1, reported: "a"},
		{source: "var [ a ] = x;", settings: idLengthOf(`{"min": 2}`), findings: 1, reported: "a"},
		{source: "var { p: [a] } = x;", settings: idLengthOf(`{"min": 2, "properties": "never"}`), findings: 1, reported: "a"},
		{source: "function f([a]) {}", settings: idLengthOf(`{"min": 2, "properties": "never"}`), findings: 2},

		// A rest element is never a property name either, so `properties` does not gate it. Every
		// rest case in the corpus runs at the default, so an arm gating them passes all of them.
		// Measured: 1 finding on `a` under both settings, and 2 on `function f(...a) {}`.
		{source: "var { ...a } = x;", settings: idLengthOf(`{"min": 2, "properties": "never"}`), findings: 1, reported: "a"},
		{source: "var { ...a } = x;", settings: idLengthOf(`{"min": 2}`), findings: 1, reported: "a"},
		{source: "function f(...a) {}", settings: idLengthOf(`{"min": 2, "properties": "never"}`), findings: 2},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, IdLength, idLengthFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, testCase.findings,
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		if testCase.reported == "" {
			continue
		}
		source := result.SourceFile.Text()
		reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != testCase.reported {
			t.Errorf("%q: expected the finding on %q, got %q", testCase.source, testCase.reported, reported)
		}
	}
}

// TestIdLengthSpansAndMessages asserts where each finding POINTS and what it says, which
// TestIdLengthFires cannot: that test compares ids and counts.
//
// The private rows earn their place twice over. `Text()` carries the leading `#`, so a port
// measuring the hashed spelling counts one character too many, and the message must render the
// hash back. And upstream's own two private messages are INCONSISTENT about where the hash goes --
// `'#{{name}}'` for too-short but `#'{{name}}'` for too-long -- which is reproduced here because
// the message text is a verdict a fixture asserts, not a thing to tidy.
//
// Every expected span and message is what the installed ESLint 10.8.1 rule produced.
func TestIdLengthSpansAndMessages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		spans    []string
		messages []string
	}{
		{
			source:   "var a = 1;",
			settings: idLengthOf(`{"min": 2}`),
			spans:    []string{"a"},
			messages: []string{"Identifier name 'a' is too short (< 2)."},
		},
		{
			source:   "var longName = 1;",
			settings: idLengthOf(`{"max": 5}`),
			spans:    []string{"longName"},
			messages: []string{"Identifier name 'longName' is too long (> 5)."},
		},
		{
			// The write half of the member split, so the finding lands on the property.
			source:   "obj.a = 1;",
			settings: idLengthOf(`{"min": 2}`),
			spans:    []string{"a"},
			messages: []string{"Identifier name 'a' is too short (< 2)."},
		},
		{
			// The span covers the hash while the LENGTH was measured without it.
			source:   "class Klass { #a = 1; }",
			settings: idLengthOf(`{"min": 2}`),
			spans:    []string{"#a"},
			messages: []string{"Identifier name '#a' is too short (< 2)."},
		},
		{
			// Upstream puts the hash outside the quotes for the too-long message only.
			source:   "class Klass { #longName = 1; }",
			settings: idLengthOf(`{"max": 5}`),
			spans:    []string{"#longName"},
			messages: []string{"Identifier name #'longName' is too long (> 5)."},
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, IdLength, idLengthFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != len(testCase.spans) {
			t.Errorf("%q: expected %d findings, got %d %v",
				testCase.source, len(testCase.spans), len(result.Diagnostics), result.MessageIds())
			continue
		}
		source := result.SourceFile.Text()
		for index, diagnostic := range result.Diagnostics {
			reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.spans[index] {
				t.Errorf("%q finding %d: expected the span to cover %q, got %q",
					testCase.source, index, testCase.spans[index], reported)
			}
			if !strings.HasPrefix(diagnostic.Message.Description, testCase.messages[index]) {
				t.Errorf("%q finding %d: expected the message to open with %q, got %q",
					testCase.source, index, testCase.messages[index], diagnostic.Message.Description)
			}
		}
	}
}

// TestIdLengthDecoderReadsEveryOption pins the option surface, including the two defaults that a
// zero-value struct gets wrong.
//
// A rule handed nil options must enforce a minimum of 2 with properties ON. The zero value says
// minimum 0 and properties off, which is not a weaker rule but a silent one, and every fixture
// above would still pass because they all configure explicitly.
func TestIdLengthDecoderReadsEveryOption(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeIdLengthOptions([]byte(`{"min": 3, "max": 9, "properties": "never", "exceptions": ["x"], "exceptionPatterns": ["^_"]}`))
	if err != nil {
		t.Fatalf("decoding a full option object: %v", err)
	}
	settings, ok := decoded.(IdLengthSettings)
	if !ok {
		t.Fatalf("expected IdLengthSettings, got %T", decoded)
	}
	if settings.Minimum != 3 || settings.Maximum != 9 || !settings.HasMaximum {
		t.Errorf("expected min 3 and max 9, got %+v", settings)
	}
	if settings.CheckProperties {
		t.Error(`expected properties "never" to turn property checking off`)
	}
	if !settings.Exceptions["x"] || len(settings.ExceptionPatterns) != 1 {
		t.Errorf("expected the exception list and pattern to be read, got %+v", settings)
	}

	// The defaults, which the zero value does not supply.
	decoded, _ = DecodeIdLengthOptions(nil)
	settings, _ = decoded.(IdLengthSettings)
	if settings.Minimum != 2 || !settings.CheckProperties || settings.HasMaximum {
		t.Errorf("expected the documented defaults (min 2, properties on, no maximum), got %+v", settings)
	}

	// An uncompilable exception pattern is an error rather than a silently dropped exemption,
	// which would report names the author asked to exempt.
	if _, err := DecodeIdLengthOptions([]byte(`{"exceptionPatterns": ["^[a-z"]}`)); err == nil {
		t.Error("expected an uncompilable exceptionPattern to be reported, got no error")
	}

	// End to end: a rule handed nil options enforces the default minimum rather than nothing.
	result := rule_testing.RunWithOptions(t, IdLength, idLengthFile, "var a = 1;", nil)
	if len(result.Diagnostics) != 1 {
		t.Errorf("expected the default minimum of 2 to report once, got %d: %v",
			len(result.Diagnostics), result.MessageIds())
	}
	// And the control: with properties on by default, an object key is checked too.
	result = rule_testing.RunWithOptions(t, IdLength, idLengthFile, "var obj = { a: 1 };", nil)
	if len(result.Diagnostics) != 1 {
		t.Errorf("expected properties to be checked by default, got %d: %v",
			len(result.Diagnostics), result.MessageIds())
	}
}
