package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is ESLint's own, imported verbatim from
// /tmp/lint-sources/eslint/tests/lib/rules/no-use-before-define.js by a script that parses the
// tester file with espree rather than scraping it, so no case was retyped and no escape was typed on
// the way in. 354 cases: 174 clean and 180 reporting, carrying 185 findings between them. Every
// extracted string was then compared byte for byte against the source file, which caught a real
// defect in the extractor: reading a template literal's cooked text off the wrong field emptied 117
// cases to the empty string, and an empty source reports nothing, so all 117 would have sat in the
// clean list asserting nothing at all.
//
// Upstream runs the corpus through two testers, one with the default parser at various ecmaVersions
// and one with the typescript-eslint parser. Both are imported into the same table here, because
// this tree has a single parser that accepts everything both testers cover.
//
// Cases whose source contains JSX are given a .tsx extension. Upstream signals the same thing through
// parserOptions.ecmaFeatures.jsx, which has no counterpart here.

// useBeforeDefineCase is one imported row.
type useBeforeDefineCase struct {
	name     string
	source   string
	fileName string
	// rawOptions is the options object as upstream writes it, MINUS the array wrapper. cohere's
	// config layer unwraps the [severity, options] tuple before dispatch, so a decoder receives the
	// bare object; copying ESLint's array spelling straight in would fail every row.
	rawOptions string
	findings   int
}

// runUseBeforeDefineCase drives one row through the rule, routing options through the rule's own
// exported decoder rather than building the settings struct directly.
//
// That routing is the point rather than a convenience. Six of this rule's seven options default to
// TRUE, so a test that constructed the struct itself would exercise a shape the live config never
// produces, and a default inversion or a broken "nofunc" string form would pass unnoticed.
func runUseBeforeDefineCase(t *testing.T, testCase useBeforeDefineCase) rule_testing.Result {
	t.Helper()

	if testCase.rawOptions == "" {
		return rule_testing.RunTyped(t, NoUseBeforeDefine, testCase.fileName, testCase.source)
	}

	decoded, err := DecodeNoUseBeforeDefineOptions(json.RawMessage(testCase.rawOptions))
	if err != nil {
		t.Fatalf("decoding %s: %v", testCase.rawOptions, err)
	}
	return rule_testing.RunTypedWithOptions(t, NoUseBeforeDefine, testCase.fileName, testCase.source, decoded)
}

func TestNoUseBeforeDefineFires(t *testing.T) {
	t.Parallel()

	cases := []useBeforeDefineCase{
		{
			name:     "case 106",
			source:   "a++; var a=19;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 107",
			source:   "a++; var a=19;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 108",
			source:   "a++; var a=19;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 109",
			source:   "a(); var a=function() {};",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 110",
			source:   "alert(a[1]); var a=[1,3];",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 111",
			source:   "a(); function a() { alert(b); var b=10; a(); }",
			fileName: "fixture.ts",
			findings: 2,
		},
		{
			name:       "case 112",
			source:     "a(); var a=function() {};",
			fileName:   "fixture.ts",
			rawOptions: "\"nofunc\"",
			findings:   1,
		},
		{
			name:     "case 113",
			source:   "(() => { alert(a); var a = 42; })();",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 114",
			source:   "(() => a())(); function a() { }",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 116",
			source:   "a(); try { throw new Error() } catch (foo) {var a;}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 117",
			source:   "var f = () => a; var a;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 118",
			source:   "new A(); class A {};",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 119",
			source:   "function foo() { new A(); } class A {};",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 120",
			source:   "new A(); var A = class {};",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 121",
			source:   "function foo() { new A(); } var A = class {};",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 122",
			source:   "a++; { var a; }",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 123",
			source:   "\"use strict\"; { a(); function a() {} }",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 124",
			source:   "{a; let a = 1}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 125",
			source:   "switch (foo) { case 1: a();\n default: \n let a;}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 126",
			source:   "if (true) { function foo() { a; } let a;}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 127",
			source:     "a(); var a=function() {};",
			fileName:   "fixture.ts",
			rawOptions: "{\"functions\": false, \"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 128",
			source:     "new A(); class A {};",
			fileName:   "fixture.ts",
			rawOptions: "{\"functions\": false, \"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 129",
			source:     "new A(); var A = class {};",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 130",
			source:     "function foo() { new A(); } var A = class {};",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:     "case 131",
			source:   "var a = a;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 132",
			source:   "let a = a + b;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 133",
			source:   "const a = foo(a);",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 134",
			source:   "function foo(a = a) {}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 135",
			source:   "var {a = a} = [];",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 136",
			source:   "var [a = a] = [];",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 137",
			source:   "var {b = a, a} = {};",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 138",
			source:   "var [b = a, a] = {};",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 139",
			source:   "var {a = 0} = a;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 140",
			source:   "var [a = 0] = a;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 141",
			source:   "for (var a in a) {}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 142",
			source:   "for (var a of a) {}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 143",
			source:     "function foo() { bar; var bar = 1; } var bar;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 144",
			source:     "foo; var foo;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:     "case 145",
			source:   "for (let x = x;;); let x = 0",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 146",
			source:   "for (let x in xs); let xs = []",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 147",
			source:   "for (let x of xs); let xs = []",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 148",
			source:   "try {} catch ({message = x}) {} let x = ''",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 150",
			source:   "with (x); let x = {}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 154",
			source:     "class C extends C {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 155",
			source:     "const C = class extends C {};",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 156",
			source:     "class C extends (class { [C](){} }) {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 157",
			source:     "const C = class extends (class { [C](){} }) {};",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 158",
			source:     "class C extends (class { static field = C; }) {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 159",
			source:     "const C = class extends (class { static field = C; }) {};",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 160",
			source:     "class C { [C](){} }",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 161",
			source:     "(class C { [C](){} });",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 162",
			source:     "const C = class { [C](){} };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 163",
			source:     "class C { static [C](){} }",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 164",
			source:     "(class C { static [C](){} });",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 165",
			source:     "const C = class { static [C](){} };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 166",
			source:     "class C { [C]; }",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 167",
			source:     "(class C { [C]; });",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 168",
			source:     "const C = class { [C]; };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 169",
			source:     "class C { [C] = foo; }",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 170",
			source:     "(class C { [C] = foo; });",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 171",
			source:     "const C = class { [C] = foo; };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 172",
			source:     "class C { static [C]; }",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 173",
			source:     "(class C { static [C]; });",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 174",
			source:     "const C = class { static [C]; };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 175",
			source:     "class C { static [C] = foo; }",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 176",
			source:     "(class C { static [C] = foo; });",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 177",
			source:     "const C = class { static [C] = foo; };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 178",
			source:     "const C = class { static field = C; };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 179",
			source:     "const C = class { static field = class extends C {}; };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 180",
			source:     "const C = class { static field = class { [C]; } };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 181",
			source:     "const C = class { static field = class { static field = C; }; };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 182",
			source:     "class C extends D {} class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 183",
			source:     "class C extends (class { [a](){} }) {} let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 184",
			source:     "class C extends (class { static field = a; }) {} let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 185",
			source:     "class C { [a]() {} } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 186",
			source:     "class C { static [a]() {} } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 187",
			source:     "class C { [a]; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 188",
			source:     "class C { static [a]; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 189",
			source:     "class C { [a] = foo; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 190",
			source:     "class C { static [a] = foo; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 191",
			source:     "class C { static field = a; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 192",
			source:     "class C { static field = D; } class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 193",
			source:     "class C { static field = class extends D {}; } class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 194",
			source:     "class C { static field = class { [a](){} } } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 195",
			source:     "class C { static field = class { static field = a; }; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 196",
			source:     "const C = class { static { C; } };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 197",
			source:     "const C = class { static { (class extends C {}); } };",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 198",
			source:     "class C { static { a; } } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 199",
			source:     "class C { static { D; } } class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 200",
			source:     "class C { static { (class extends D {}); } } class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 201",
			source:     "class C { static { (class { [a](){} }); } } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 202",
			source:     "class C { static { (class { static field = a; }); } } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 203",
			source:     "(class C extends C {});",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 204",
			source:     "(class C extends (class { [C](){} }) {});",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 205",
			source:     "(class C extends (class { static field = C; }) {});",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:     "case 206",
			source:   "export { a }; const a = 1;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 207",
			source:     "export { a }; const a = 1;",
			fileName:   "fixture.ts",
			rawOptions: "{}",
			findings:   1,
		},
		{
			name:       "case 208",
			source:     "export { a }; const a = 1;",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": false}",
			findings:   1,
		},
		{
			name:       "case 209",
			source:     "export { a }; const a = 1;",
			fileName:   "fixture.ts",
			rawOptions: "\"nofunc\"",
			findings:   1,
		},
		{
			name:     "case 210",
			source:   "export { a as b }; const a = 1;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 211",
			source:   "export { a, b }; let a, b;",
			fileName: "fixture.ts",
			findings: 2,
		},
		{
			name:     "case 212",
			source:   "export { a }; var a;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 213",
			source:   "export { f }; function f() {}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 214",
			source:   "export { C }; class C {}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 215",
			source:     "export const foo = a; const a = 1;",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
			findings:   1,
		},
		{
			name:       "case 216",
			source:     "export default a; const a = 1;",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
			findings:   1,
		},
		{
			name:       "case 217",
			source:     "export function foo() { return a; }; const a = 1;",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
			findings:   1,
		},
		{
			name:       "case 218",
			source:     "export class C { foo() { return a; } }; const a = 1;",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
			findings:   1,
		},
		{
			name:     "case 219",
			source:   "<App />; const App = () => <div />;",
			fileName: "fixture.tsx",
			findings: 1,
		},
		{
			name:     "case 220",
			source:   "function render() { return <Widget /> }; const Widget = () => <span />;",
			fileName: "fixture.tsx",
			findings: 1,
		},
		{
			name:     "case 221",
			source:   "<Foo.Bar />; const Foo = { Bar: () => <div/> };",
			fileName: "fixture.tsx",
			findings: 1,
		},
		{
			name:     "case 290",
			source:   "\n\ta++;\n\tvar a = 19;\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 291",
			source:   "\n\ta++;\n\tvar a = 19;\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 292",
			source:   "\n\ta++;\n\tvar a = 19;\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 293",
			source:   "\n\ta();\n\tvar a = function () {};\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 294",
			source:   "\n\talert(a[1]);\n\tvar a = [1, 3];\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 295",
			source:   "\n\ta();\n\tfunction a() {\n\t  alert(b);\n\t  var b = 10;\n\t  a();\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 2,
		},
		{
			name:       "case 296",
			source:     "\n\ta();\n\tvar a = function () {};\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "\"nofunc\"",
			findings:   1,
		},
		{
			name:     "case 297",
			source:   "\n\t(() => {\n\t  alert(a);\n\t  var a = 42;\n\t})();\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 298",
			source:   "\n\t(() => a())();\n\tfunction a() {}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 299",
			source:   "\n\ta();\n\ttry {\n\t  throw new Error();\n\t} catch (foo) {\n\t  var a;\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 300",
			source:   "\n\tvar f = () => a;\n\tvar a;\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 301",
			source:   "\n\tnew A();\n\tclass A {}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 302",
			source:   "\n\tfunction foo() {\n\t  new A();\n\t}\n\tclass A {}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 303",
			source:   "\n\tnew A();\n\tvar A = class {};\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 304",
			source:   "\n\tfunction foo() {\n\t  new A();\n\t}\n\tvar A = class {};\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 305",
			source:   "\n\ta++;\n\t{\n\t  var a;\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 306",
			source:   "\n\t'use strict';\n\t{\n\t  a();\n\t  function a() {}\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 307",
			source:   "\n\t{\n\t  a;\n\t  let a = 1;\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 308",
			source:   "\n\tswitch (foo) {\n\t  case 1:\n\t\ta();\n\t  default:\n\t\tlet a;\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 309",
			source:   "\n\tif (true) {\n\t  function foo() {\n\t\ta;\n\t  }\n\t  let a;\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 310",
			source:     "\n\ta();\n\tvar a = function () {};\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false, \"functions\": false}",
			findings:   1,
		},
		{
			name:       "case 311",
			source:     "\n\tnew A();\n\tvar A = class {};\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:       "case 312",
			source:     "\n\tfunction foo() {\n\t  new A();\n\t}\n\tvar A = class {};\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
			findings:   1,
		},
		{
			name:     "case 313",
			source:   "var a = a;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 314",
			source:   "let a = a + b;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 315",
			source:   "const a = foo(a);",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 316",
			source:   "function foo(a = a) {}",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 317",
			source:   "var { a = a } = [];",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 318",
			source:   "var [a = a] = [];",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 319",
			source:   "var { b = a, a } = {};",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 320",
			source:   "var [b = a, a] = {};",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 321",
			source:   "var { a = 0 } = a;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 322",
			source:   "var [a = 0] = a;",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 323",
			source:   "\n\tfor (var a in a) {\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 324",
			source:   "\n\tfor (var a of a) {\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 325",
			source:     "\n\tinterface Bar {\n\t  type: typeof Foo;\n\t}\n\t\n\tconst Foo = 2;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": false}",
			findings:   1,
		},
		{
			name:       "case 326",
			source:     "\n\tlet var1: StringOrNumber;\n\ntype StringOrNumber = string | number;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": false, \"typedefs\": true}",
			findings:   1,
		},
		{
			name:       "case 327",
			source:     "\n\tinterface Bar {\n\t  type: typeof Foo.FOO;\n\t}\n\t\n\tclass Foo {\n\t  public static readonly FOO = '';\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": false}",
			findings:   1,
		},
		{
			name:       "case 328",
			source:     "\n\tinterface Bar {\n\t  type: typeof Foo.Bar.Baz;\n\t}\n\t\n\tconst Foo = {\n\t  Bar: {\n\t\tBaz: 1,\n\t  },\n\t};\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": false}",
			findings:   1,
		},
		{
			name:       "case 329",
			source:     "\n\tconst foo = {\n\t  bar: 'bar',\n\t} satisfies {\n\t  bar: typeof baz;\n\t};\n\t\n\tconst baz = '';\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": false}",
			findings:   1,
		},
		{
			name:       "case 330",
			source:     "\n\tfunction foo() {\n\t  bar;\n\t  var bar = 1;\n\t}\n\tvar bar;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
			findings:   1,
		},
		{
			name:       "case 331",
			source:     "\n\tclass Test {\n\t  foo(args: Foo): Foo {\n\t\treturn Foo.FOO;\n\t  }\n\t}\n\t\n\tenum Foo {\n\t  FOO,\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"enums\": true}",
			findings:   1,
		},
		{
			name:       "case 332",
			source:     "\n\tfunction foo(): Foo {\n\t  return Foo.FOO;\n\t}\n\t\n\tenum Foo {\n\t   FOO,\n\t }\n\t",
			fileName:   "fixture.ts",
			rawOptions: "{\"enums\": true}",
			findings:   1,
		},
		{
			name:       "case 333",
			source:     "\n\tconst foo = Foo.Foo;\n\t\n\tenum Foo {\n\t  FOO,\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"enums\": true}",
			findings:   1,
		},
		{
			name:     "case 334",
			source:   "\n\texport { a };\n\tconst a = 1;\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 335",
			source:     "\n\texport { a };\n\tconst a = 1;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{}",
			findings:   1,
		},
		{
			name:       "case 336",
			source:     "\n\texport { a };\n\tconst a = 1;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": false}",
			findings:   1,
		},
		{
			name:       "case 337",
			source:     "\n\texport { a };\n\tconst a = 1;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "\"nofunc\"",
			findings:   1,
		},
		{
			name:     "case 338",
			source:   "\n\texport { a as b };\n\tconst a = 1;\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 339",
			source:   "\n\texport { a, b };\n\tlet a, b;\n\t\t  ",
			fileName: "fixture.ts",
			findings: 2,
		},
		{
			name:     "case 340",
			source:   "\n\texport { a };\n\tvar a;\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 341",
			source:   "\n\texport { f };\n\tfunction f() {}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 342",
			source:   "\n\texport { C };\n\tclass C {}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 343",
			source:     "\n\texport const foo = a;\n\tconst a = 1;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
			findings:   1,
		},
		{
			name:       "case 344",
			source:     "\n\texport function foo() {\n\t  return a;\n\t}\n\tconst a = 1;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
			findings:   1,
		},
		{
			name:       "case 345",
			source:     "\n\texport class C {\n\t  foo() {\n\t\treturn a;\n\t  }\n\t}\n\tconst a = 1;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
			findings:   1,
		},
		{
			name:     "case 346",
			source:   "\n\texport { Foo };\n\t\n\tenum Foo {\n\t  BAR,\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 347",
			source:   "\n\texport { Foo };\n\t\n\tnamespace Foo {\n\t  export let bar = () => console.log('bar');\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:       "case 348",
			source:     "\n\texport { Foo, baz };\n\t\n\tenum Foo {\n\t  BAR,\n\t}\n\t\n\tlet baz: Enum;\n\tenum Enum {}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": false, \"ignoreTypeReferences\": true}",
			findings:   2,
		},
		{
			name:     "case 349",
			source:   "\n\tf();\n\tfunction f() {}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 350",
			source:   "\n\talert(a);\n\tvar a = 10;\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 351",
			source:   "\n\tf()?.();\n\tfunction f() {\n\t  return function t() {};\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 352",
			source:   "\n\talert(a?.b);\n\tvar a = { b: 5 };\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
		{
			name:     "case 353",
			source:   "\n\t@decorator\n\tclass C {\n  \t\tstatic x = \"foo\";\n  \t\t[C.x]() { }\n\t}\n\t\t  ",
			fileName: "fixture.ts",
			findings: 1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runUseBeforeDefineCase(t, testCase)

			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "usedBeforeDefined"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

func TestNoUseBeforeDefineStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []useBeforeDefineCase{
		{
			name:     "case 0",
			source:   "unresolved",
			fileName: "fixture.ts",
		},
		{
			name:     "case 1",
			source:   "Array",
			fileName: "fixture.ts",
		},
		{
			name:     "case 2",
			source:   "function foo () { arguments; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 3",
			source:   "var a=10; alert(a);",
			fileName: "fixture.ts",
		},
		{
			name:     "case 4",
			source:   "function b(a) { alert(a); }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 5",
			source:   "Object.hasOwnProperty.call(a);",
			fileName: "fixture.ts",
		},
		{
			name:     "case 6",
			source:   "function a() { alert(arguments);}",
			fileName: "fixture.ts",
		},
		{
			name:       "case 7",
			source:     "a(); function a() { alert(arguments); }",
			fileName:   "fixture.ts",
			rawOptions: "\"nofunc\"",
		},
		{
			name:     "case 8",
			source:   "(() => { var a = 42; alert(a); })();",
			fileName: "fixture.ts",
		},
		{
			name:     "case 9",
			source:   "a(); try { throw new Error() } catch (a) {}",
			fileName: "fixture.ts",
		},
		{
			name:     "case 10",
			source:   "class A {} new A();",
			fileName: "fixture.ts",
		},
		{
			name:     "case 11",
			source:   "var a = 0, b = a;",
			fileName: "fixture.ts",
		},
		{
			name:     "case 12",
			source:   "var {a = 0, b = a} = {};",
			fileName: "fixture.ts",
		},
		{
			name:     "case 13",
			source:   "var [a = 0, b = a] = {};",
			fileName: "fixture.ts",
		},
		{
			name:     "case 14",
			source:   "function foo() { foo(); }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 15",
			source:   "var foo = function() { foo(); };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 16",
			source:   "var a; for (a in a) {}",
			fileName: "fixture.ts",
		},
		{
			name:     "case 17",
			source:   "var a; for (a of a) {}",
			fileName: "fixture.ts",
		},
		{
			name:     "case 18",
			source:   "let a; class C { static { a; } }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 19",
			source:   "class C { static { let a; a; } }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 20",
			source:   "\"use strict\"; a(); { function a() {} }",
			fileName: "fixture.ts",
		},
		{
			name:       "case 21",
			source:     "\"use strict\"; { a(); function a() {} }",
			fileName:   "fixture.ts",
			rawOptions: "\"nofunc\"",
		},
		{
			name:     "case 22",
			source:   "switch (foo) { case 1:  { a(); } default: { let a; }}",
			fileName: "fixture.ts",
		},
		{
			name:     "case 23",
			source:   "a(); { let a = function () {}; }",
			fileName: "fixture.ts",
		},
		{
			name:       "case 24",
			source:     "a(); function a() { alert(arguments); }",
			fileName:   "fixture.ts",
			rawOptions: "{\"functions\": false}",
		},
		{
			name:       "case 25",
			source:     "\"use strict\"; { a(); function a() {} }",
			fileName:   "fixture.ts",
			rawOptions: "{\"functions\": false}",
		},
		{
			name:       "case 26",
			source:     "function foo() { new A(); } class A {};",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
		},
		{
			name:       "case 27",
			source:     "function foo() { bar; } var bar;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:       "case 28",
			source:     "var foo = () => bar; var bar;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:       "case 29",
			source:     "class C { static { () => foo; let foo; } }",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:     "case 30",
			source:   "class C extends (class { method() { C; } }) {}",
			fileName: "fixture.ts",
		},
		{
			name:     "case 31",
			source:   "(class extends (class { method() { C; } }) {});",
			fileName: "fixture.ts",
		},
		{
			name:     "case 32",
			source:   "const C = (class extends (class { method() { C; } }) {});",
			fileName: "fixture.ts",
		},
		{
			name:     "case 33",
			source:   "class C extends (class { field = C; }) {}",
			fileName: "fixture.ts",
		},
		{
			name:     "case 34",
			source:   "(class extends (class { field = C; }) {});",
			fileName: "fixture.ts",
		},
		{
			name:     "case 35",
			source:   "const C = (class extends (class { field = C; }) {});",
			fileName: "fixture.ts",
		},
		{
			name:     "case 36",
			source:   "class C { [() => C](){} }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 37",
			source:   "(class C { [() => C](){} });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 38",
			source:   "const C = class { [() => C](){} };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 39",
			source:   "class C { static [() => C](){} }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 40",
			source:   "(class C { static [() => C](){} });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 41",
			source:   "const C = class { static [() => C](){} };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 42",
			source:   "class C { [() => C]; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 43",
			source:   "(class C { [() => C]; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 44",
			source:   "const C = class { [() => C]; };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 45",
			source:   "class C { static [() => C]; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 46",
			source:   "(class C { static [() => C]; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 47",
			source:   "const C = class { static [() => C]; };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 48",
			source:   "class C { method() { C; } }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 49",
			source:   "(class C { method() { C; } });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 50",
			source:   "const C = class { method() { C; } };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 51",
			source:   "class C { static method() { C; } }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 52",
			source:   "(class C { static method() { C; } });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 53",
			source:   "const C = class { static method() { C; } };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 54",
			source:   "class C { field = C; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 55",
			source:   "(class C { field = C; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 56",
			source:   "const C = class { field = C; };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 57",
			source:   "class C { static field = C; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 58",
			source:   "(class C { static field = C; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 59",
			source:   "class C { static field = class { static field = C; }; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 60",
			source:   "(class C { static field = class { static field = C; }; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 61",
			source:   "class C { field = () => C; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 62",
			source:   "(class C { field = () => C; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 63",
			source:   "const C = class { field = () => C; };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 64",
			source:   "class C { static field = () => C; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 65",
			source:   "(class C { static field = () => C; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 66",
			source:   "const C = class { static field = () => C; };",
			fileName: "fixture.ts",
		},
		{
			name:     "case 67",
			source:   "class C { field = class extends C {}; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 68",
			source:   "(class C { field = class extends C {}; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 69",
			source:   "const C = class { field = class extends C {}; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 70",
			source:   "class C { static field = class extends C {}; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 71",
			source:   "(class C { static field = class extends C {}; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 72",
			source:   "class C { static field = class { [C]; }; }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 73",
			source:   "(class C { static field = class { [C]; }; });",
			fileName: "fixture.ts",
		},
		{
			name:     "case 74",
			source:   "const C = class { static field = class { field = C; }; };",
			fileName: "fixture.ts",
		},
		{
			name:       "case 75",
			source:     "class C { method() { a; } } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:       "case 76",
			source:     "class C { static method() { a; } } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:       "case 77",
			source:     "class C { field = a; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:       "case 78",
			source:     "class C { field = D; } class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
		},
		{
			name:       "case 79",
			source:     "class C { field = class extends D {}; } class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
		},
		{
			name:       "case 80",
			source:     "class C { field = () => a; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:       "case 81",
			source:     "class C { static field = () => a; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:       "case 82",
			source:     "class C { field = () => D; } class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
		},
		{
			name:       "case 83",
			source:     "class C { static field = () => D; } class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
		},
		{
			name:       "case 84",
			source:     "class C { static field = class { field = a; }; } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:     "case 85",
			source:   "class C { static { C; } }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 86",
			source:   "class C { static { C; } static {} static { C; } }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 87",
			source:   "(class C { static { C; } })",
			fileName: "fixture.ts",
		},
		{
			name:     "case 88",
			source:   "class C { static { class D extends C {} } }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 89",
			source:   "class C { static { (class { static { C } }) } }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 90",
			source:   "class C { static { () => C; } }",
			fileName: "fixture.ts",
		},
		{
			name:     "case 91",
			source:   "(class C { static { () => C; } })",
			fileName: "fixture.ts",
		},
		{
			name:     "case 92",
			source:   "const C = class { static { () => C; } }",
			fileName: "fixture.ts",
		},
		{
			name:       "case 93",
			source:     "class C { static { () => D; } } class D {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
		},
		{
			name:       "case 94",
			source:     "class C { static { () => a; } } let a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:     "case 95",
			source:   "const C = class C { static { C.x; } }",
			fileName: "fixture.ts",
		},
		{
			name:       "case 96",
			source:     "export { a }; const a = 1;",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 97",
			source:     "export { a as b }; const a = 1;",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 98",
			source:     "export { a, b }; let a, b;",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 99",
			source:     "export { a }; var a;",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 100",
			source:     "export { f }; function f() {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 101",
			source:     "export { C }; class C {}",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:     "case 102",
			source:   "const App = () => <div/>; <App />;",
			fileName: "fixture.tsx",
		},
		{
			name:     "case 103",
			source:   "let Foo, Bar; <Foo><Bar /></Foo>;",
			fileName: "fixture.tsx",
		},
		{
			name:     "case 104",
			source:   "function App() { return <div/> } <App />;",
			fileName: "fixture.tsx",
		},
		{
			name:       "case 105",
			source:     "<App />; function App() { return <div/> }",
			fileName:   "fixture.tsx",
			rawOptions: "{\"functions\": false}",
		},
		{
			name:     "case 222",
			source:   "\n\ttype foo = 1;\n\tconst x: foo = 1;\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 223",
			source:   "\n\ttype foo = 1;\n\ttype bar = foo;\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 224",
			source:   "\n\tinterface Foo {}\n\tconst x: Foo = {};\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 225",
			source:   "\n\tvar a = 10;\n\talert(a);\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 226",
			source:   "\n\tfunction b(a) {\n\t  alert(a);\n\t}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 227",
			source:   "Object.hasOwnProperty.call(a);",
			fileName: "fixture.ts",
		},
		{
			name:     "case 228",
			source:   "\n\tfunction a() {\n\t  alert(arguments);\n\t}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 229",
			source:   "declare function a();",
			fileName: "fixture.ts",
		},
		{
			name:     "case 230",
			source:   "\n\tdeclare class a {\n\t  foo();\n\t}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 231",
			source:   "const updatedAt = data?.updatedAt;",
			fileName: "fixture.ts",
		},
		{
			name:     "case 232",
			source:   "\n\tfunction f() {\n\t  return function t() {};\n\t}\n\tf()?.();\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 233",
			source:   "\n\tvar a = { b: 5 };\n\talert(a?.b);\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:       "case 234",
			source:     "\n\ta();\n\tfunction a() {\n\t  alert(arguments);\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "\"nofunc\"",
		},
		{
			name:     "case 235",
			source:   "\n\t(() => {\n\t  var a = 42;\n\t  alert(a);\n\t})();\n\t\t  ",
			fileName: "fixture.ts",
		},
		{
			name:     "case 236",
			source:   "\n\ta();\n\ttry {\n\t  throw new Error();\n\t} catch (a) {}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 237",
			source:   "\n\tclass A {}\n\tnew A();\n\t\t  ",
			fileName: "fixture.ts",
		},
		{
			name:     "case 238",
			source:   "\n\tvar a = 0,\n\t  b = a;\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 239",
			source:   "var { a = 0, b = a } = {};",
			fileName: "fixture.ts",
		},
		{
			name:     "case 240",
			source:   "var [a = 0, b = a] = {};",
			fileName: "fixture.ts",
		},
		{
			name:     "case 241",
			source:   "\n\tfunction foo() {\n\t  foo();\n\t}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 242",
			source:   "\n\tvar foo = function () {\n\t  foo();\n\t};\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 243",
			source:   "\n\tvar a;\n\tfor (a in a) {\n\t}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 244",
			source:   "\n\tvar a;\n\tfor (a of a) {\n\t}\n\t\t  ",
			fileName: "fixture.ts",
		},
		{
			name:     "case 245",
			source:   "\n\t'use strict';\n\ta();\n\t{\n\t  function a() {}\n\t}\n\t\t  ",
			fileName: "fixture.ts",
		},
		{
			name:       "case 246",
			source:     "\n\t'use strict';\n\t{\n\t  a();\n\t  function a() {}\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "\"nofunc\"",
		},
		{
			name:     "case 247",
			source:   "\n\tswitch (foo) {\n\t  case 1: {\n\t\ta();\n\t  }\n\t  default: {\n\t\tlet a;\n\t  }\n\t}\n\t\t  ",
			fileName: "fixture.ts",
		},
		{
			name:     "case 248",
			source:   "\n\ta();\n\t{\n\t  let a = function () {};\n\t}\n\t\t  ",
			fileName: "fixture.ts",
		},
		{
			name:       "case 249",
			source:     "\n\ta();\n\tfunction a() {\n\t  alert(arguments);\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"functions\": false}",
		},
		{
			name:       "case 250",
			source:     "\n\t'use strict';\n\t{\n\t  a();\n\t  function a() {}\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"functions\": false}",
		},
		{
			name:       "case 251",
			source:     "\n\tfunction foo() {\n\t  new A();\n\t}\n\tclass A {}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
		},
		{
			name:       "case 252",
			source:     "\n\tfunction foo() {\n\t  bar;\n\t}\n\tvar bar;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:       "case 253",
			source:     "\n\tvar foo = () => bar;\n\tvar bar;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"variables\": false}",
		},
		{
			name:       "case 254",
			source:     "\n\tvar x: Foo = 2;\n\ttype Foo = string | number;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"typedefs\": false}",
		},
		{
			name:       "case 255",
			source:     "\n\tvar x: Foo = {};\n\tinterface Foo {}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"typedefs\": false, \"ignoreTypeReferences\": false}",
		},
		{
			name:       "case 256",
			source:     "\n\tlet myVar: String;\n\ttype String = string;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"typedefs\": false, \"ignoreTypeReferences\": false}",
		},
		{
			name:       "case 257",
			source:     "\n\tinterface Bar {\n\t  type: typeof Foo;\n\t}\n\t\n\tconst Foo = 2;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": true}",
		},
		{
			name:       "case 258",
			source:     "\n\tinterface Bar {\n\t  type: typeof Foo.FOO;\n\t}\n\t\n\tclass Foo {\n\t  public static readonly FOO = '';\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": true}",
		},
		{
			name:       "case 259",
			source:     "\n\tinterface Bar {\n\t  type: typeof Foo.Bar.Baz;\n\t}\n\t\n\tconst Foo = {\n\t  Bar: {\n\t\tBaz: 1,\n\t  },\n\t};\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": true}",
		},
		{
			name:     "case 260",
			source:   "\n\tinterface ITest {\n\t  first: boolean;\n\t  second: string;\n\t  third: boolean;\n\t}\n\t\n\tlet first = () => console.log('first');\n\t\n\texport let second = () => console.log('second');\n\t\n\texport namespace Third {\n\t  export let third = () => console.log('third');\n\t}\n\t\t  ",
			fileName: "fixture.ts",
		},
		{
			name:     "case 261",
			source:   "\n\tfunction test(file: Blob) {\n\t  const slice: typeof file.slice =\n\t\tfile.slice || (file as any).webkitSlice || (file as any).mozSlice;\n\t  return slice;\n\t}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 262",
			source:   "\n\tinterface Foo {\n\t  bar: string;\n\t}\n\tconst bar = 'blah';\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:       "case 263",
			source:     "\n\tfunction foo(): Foo {\n\t  return Foo.FOO;\n\t}\n\t\n\tenum Foo {\n\t  FOO,\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"enums\": false}",
		},
		{
			name:       "case 264",
			source:     "\n\tlet foo: Foo;\n\t\n\tenum Foo {\n\t  FOO,\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"enums\": false}",
		},
		{
			name:       "case 265",
			source:     "\n\tclass Test {\n\t  foo(args: Foo): Foo {\n\t\treturn Foo.FOO;\n\t  }\n\t}\n\t\n\tenum Foo {\n\t  FOO,\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"enums\": false}",
		},
		{
			name:       "case 266",
			source:     "\n\texport { a };\n\tconst a = 1;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 267",
			source:     "\n\texport { a as b };\n\tconst a = 1;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 268",
			source:     "\n\texport { a, b };\n\tlet a, b;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 269",
			source:     "\n\texport { a };\n\tvar a;\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 270",
			source:     "\n\texport { f };\n\tfunction f() {}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 271",
			source:     "\n\texport { C };\n\tclass C {}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 272",
			source:     "\n\texport { Foo };\n\t\n\tenum Foo {\n\t  BAR,\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 273",
			source:     "\n\texport { Foo };\n\t\n\tnamespace Foo {\n\t  export let bar = () => console.log('bar');\n\t}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:       "case 274",
			source:     "\n\texport { Foo, baz };\n\t\n\tenum Foo {\n\t  BAR,\n\t}\n\t\n\tlet baz: Enum;\n\tenum Enum {}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"allowNamedExports\": true}",
		},
		{
			name:     "case 275",
			source:   "\n\timport * as React from 'react';\n\t\n\t<div />;\n\t\t  ",
			fileName: "fixture.tsx",
		},
		{
			name:     "case 276",
			source:   "\n\timport React from 'react';\n\t\n\t<div />;\n\t\t  ",
			fileName: "fixture.tsx",
		},
		{
			name:     "case 277",
			source:   "\n\timport { h } from 'preact';\n\t\n\t<div />;\n\t\t  ",
			fileName: "fixture.tsx",
		},
		{
			name:     "case 278",
			source:   "\n\tconst React = require('react');\n\t\n\t<div />;\n\t\t  ",
			fileName: "fixture.tsx",
		},
		{
			name:     "case 279",
			source:   "\n\ttype T = (value: unknown) => value is Id;\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 280",
			source:   "\n\tglobal.foo = true;\n\t\n\tdeclare global {\n\t  namespace NodeJS {\n\t\tinterface Global {\n\t\t  foo?: boolean;\n\t\t}\n\t  }\n\t}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 281",
			source:   "\n\t@Directive({\n\t  selector: '[rcCidrIpPattern]',\n\t  providers: [\n\t\t{\n\t\t  provide: NG_VALIDATORS,\n\t\t  useExisting: CidrIpPatternDirective,\n\t\t  multi: true,\n\t\t},\n\t  ],\n\t})\n\texport class CidrIpPatternDirective implements Validator {}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:       "case 282",
			source:     "\n\t@Directive({\n\t  selector: '[rcCidrIpPattern]',\n\t  providers: [\n\t\t{\n\t\t  provide: NG_VALIDATORS,\n\t\t  useExisting: CidrIpPatternDirective,\n\t\t  multi: true,\n\t\t},\n\t  ],\n\t})\n\texport class CidrIpPatternDirective implements Validator {}\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"classes\": false}",
		},
		{
			name:     "case 283",
			source:   "\n\tclass A {\n\t  constructor(printName) {\n\t\tthis.printName = printName;\n\t  }\n\t\n\t  openPort(printerName = this.printerName) {\n\t\tthis.tscOcx.ActiveXopenport(printerName);\n\t\n\t\treturn this;\n\t  }\n\t}\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:       "case 284",
			source:     "\n\tconst obj = {\n\t  foo: 'foo-value',\n\t  bar: 'bar-value',\n\t} satisfies {\n\t  [key in 'foo' | 'bar']: `${key}-value`;\n\t};\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": false}",
		},
		{
			name:       "case 285",
			source:     "\n\tconst obj = {\n\t  foo: 'foo-value',\n\t  bar: 'bar-value',\n\t} as {\n\t  [key in 'foo' | 'bar']: `${key}-value`;\n\t};\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": false}",
		},
		{
			name:       "case 286",
			source:     "\n\tconst obj = {\n\t  foo: {\n\t\tfoo: 'foo',\n\t  } as {\n\t\t[key in 'foo' | 'bar']: key;\n\t  },\n\t};\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": false}",
		},
		{
			name:       "case 287",
			source:     "\n\tconst foo = {\n\t  bar: 'bar',\n\t} satisfies {\n\t  bar: typeof baz;\n\t};\n\t\n\tconst baz = '';\n\t\t  ",
			fileName:   "fixture.ts",
			rawOptions: "{\"ignoreTypeReferences\": true}",
		},
		{
			name:     "case 288",
			source:   "\n\tnamespace A.X.Y {}\n\t\n\timport Z = A.X.Y;\n\t\n\tconst X = 23;\n\t\t",
			fileName: "fixture.ts",
		},
		{
			name:     "case 289",
			source:   "\n\t\tnamespace A {\n\t\t\texport namespace X {\n\t\t\t\texport namespace Y {\n\t\t\t\t\texport const foo = 40;\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\n\t\timport Z = A.X.Y;\n\n\t\tconst X = 23;\n\t\t",
			fileName: "fixture.ts",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runUseBeforeDefineCase(t, testCase)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUseBeforeDefineReportsTheUseNotTheDeclaration pins WHERE the finding points.
//
// ExpectFindings asserts ids and count and nothing else, so a rule anchored on the declaration
// rather than on the reference passes the entire imported corpus while pointing every finding at the
// wrong line. The two positions are far apart in exactly the cases this rule exists for, and
// upstream reports on the reference.
func TestNoUseBeforeDefineReportsTheUseNotTheDeclaration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		source         string
		reportedText   string
		reportedOffset int
	}{
		{"variable read above its declaration", "a++; var a=19;", "a", 0},
		{"class construction above its declaration", "new A(); class A {};", "A", 4},
		{"temporal dead zone read", "{a; let a = 1}", "a", 1},
		{"closure reads a later binding", "var f = () => a; var a;", "a", 14},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}

			diagnostic := result.Diagnostics[0]

			// Slice the source the harness actually wrote rather than the Go literal above.
			// RunTyped writes strings.TrimSpace(source)+"\n", so a span sliced from the literal is
			// off by one for any fixture carrying leading whitespace.
			written := strings.TrimSpace(testCase.source) + "\n"
			reported := written[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.reportedText {
				t.Errorf("finding text is %q, want %q", reported, testCase.reportedText)
			}

			// The offset is asserted as well as the text, because every one of these sources
			// contains the reported name twice and a text comparison alone cannot tell the
			// reference from the declaration.
			if diagnostic.Range.Pos() != testCase.reportedOffset {
				t.Errorf("finding starts at %d, want %d", diagnostic.Range.Pos(), testCase.reportedOffset)
			}
		})
	}
}

// TestNoUseBeforeDefineMessageText asserts the rendered description exactly, per kind of binding.
//
// The message used to be one constant telling every finding that the read "throws at runtime" (for
// `let`, `const` or a class) or "yields `undefined`" (for `var`). Measured on ahra, 10 of the 12
// findings read a FUNCTION DECLARATION, which is hoisted with its body, so the read works and the
// message described a crash that cannot happen; the other two read a module-level `let` and `const`
// from inside a function written above them, which throws only if that function is called first.
// Each row below is a kind the message now distinguishes, asserted against literals typed here.
func TestNoUseBeforeDefineMessageText(t *testing.T) {
	t.Parallel()

	const repair = " Move the declaration above its first use."
	const readingOrder = "the cost is reading order: the reader meets the name before learning what it is, " +
		"and has to jump down the file to find out."

	for _, testCase := range []struct {
		name    string
		source  string
		options string
		want    string
	}{
		{"a type is erased, so only reading order is at stake",
			"let total: Amount = 1;\nconsole.log(total);\ntype Amount = number;",
			`{"ignoreTypeReferences": false}`,
			"'Amount' is used as a type above the line that declares it. A type is erased before the " +
				"program runs, so nothing fails at runtime; " + readingOrder + repair},
		{"a class used only as a type is erased too",
			"let held: Widget | undefined;\nconsole.log(held);\nclass Widget {}",
			`{"ignoreTypeReferences": false}`,
			"'Widget' is used as a type above the line that declares it. A type is erased before the " +
				"program runs, so nothing fails at runtime; " + readingOrder + repair},
		{"a function declaration is hoisted with its body",
			"normalize('a');\nfunction normalize(text: string): string { return text; }",
			"",
			"'normalize' is read above the function declaration that defines it. A function declaration " +
				"is hoisted together with its body, so the read works at runtime; " + readingOrder + repair},
		{"a const read where it is written",
			"console.log(limit);\nconst limit = 1;",
			"",
			"'limit' is read above the `const` that declares it. The binding exists from the top of its " +
				"block but cannot be touched until its declaration runs, so this read lands in the " +
				"temporal dead zone and throws a ReferenceError." + repair},
		{"a const read inside a function written above it",
			"export function floor(): number { return shortenedAnchorFloor; }\nconst shortenedAnchorFloor = 3;",
			"",
			"'shortenedAnchorFloor' is read inside a function written above the `const` that declares it. " +
				"That is safe only while nothing calls the function before the declaration runs: an earlier " +
				"call, such as one made while the module is still loading, reaches the binding in its " +
				"temporal dead zone and throws a ReferenceError." + repair},
		{"a let read inside a function written above it",
			"export function model(): string { return dynamicLocalModel; }\nlet dynamicLocalModel = 'a';\ndynamicLocalModel = 'b';",
			"",
			"'dynamicLocalModel' is read inside a function written above the `let` that declares it. " +
				"That is safe only while nothing calls the function before the declaration runs: an earlier " +
				"call, such as one made while the module is still loading, reaches the binding in its " +
				"temporal dead zone and throws a ReferenceError." + repair},
		{"a class",
			"new Widget();\nclass Widget {}",
			"",
			"'Widget' is read above the `class` that declares it. The binding exists from the top of its " +
				"block but cannot be touched until its declaration runs, so this read lands in the " +
				"temporal dead zone and throws a ReferenceError." + repair},
		{"a var",
			"console.log(count);\nvar count = 1;",
			"",
			"'count' is read above the `var` that declares it. The binding is hoisted without its value, " +
				"so this read silently yields `undefined`, and the failure surfaces later and somewhere else." +
				repair},
		{"a var read inside a function written above it",
			"export function current(): number { return count; }\nvar count = 1;",
			"",
			"'count' is read inside a function written above the `var` that declares it. The binding is " +
				"hoisted without its value, so a call made before the declaration runs reads `undefined` " +
				"silently, and the failure surfaces later and somewhere else." + repair},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := runUseBeforeDefineCase(t, useBeforeDefineCase{
				fileName:   "a.ts",
				source:     testCase.source,
				rawOptions: testCase.options,
			})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Id; got != "usedBeforeDefined" {
				t.Errorf("message id is %q, want %q", got, "usedBeforeDefined")
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.want {
				t.Errorf("description is\n%q\nwant\n%q", got, testCase.want)
			}
		})
	}
}

// TestNoUseBeforeDefineDecodesTheStringForm covers the option spelling no struct unmarshal accepts.
//
// Upstream's schema is a oneOf, so the option may arrive as the bare string "nofunc" rather than an
// object, meaning exactly {"functions": false}. It reaches the decoder as a JSON string, which a
// struct unmarshal rejects outright, so without the string arm the config layer would surface a
// legal spelling as a configuration error.
func TestNoUseBeforeDefineDecodesTheStringForm(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeNoUseBeforeDefineOptions(json.RawMessage(`"nofunc"`))
	if err != nil {
		t.Fatalf("decoding the string form: %v", err)
	}

	settings := resolveNoUseBeforeDefineSettings(decoded)
	if settings.functions {
		t.Error(`"nofunc" left the functions option on`)
	}
	// Everything else must keep its default, which is what makes this a shorthand for one option
	// rather than a reset.
	if !settings.classes || !settings.variables || !settings.enums || !settings.typedefs {
		t.Errorf(`"nofunc" disturbed an option other than functions: %+v`, settings)
	}
}

// TestNoUseBeforeDefineDefaultsSurviveAnAbsentConfiguration is the guard against the failure mode
// this rule is most exposed to.
//
// A rule configured as a bare "error" is handed nil options. DecodeOptionsInto errors on empty
// input, the config layer turns that into nil for a non-required rule, and a type assertion on nil
// yields the zero value, which for a struct of bools is false everywhere. Six of this rule's seven
// options default to TRUE, so that path would leave a rule registered on every file and judging
// almost nothing, with every fixture above still green because every one of them reaches the rule
// through the decoder.
func TestNoUseBeforeDefineDefaultsSurviveAnAbsentConfiguration(t *testing.T) {
	t.Parallel()

	for _, options := range []any{nil, any(nil), NoUseBeforeDefineOptions{}} {
		settings := resolveNoUseBeforeDefineSettings(options)

		if !settings.functions || !settings.classes || !settings.variables ||
			!settings.enums || !settings.typedefs || !settings.ignoreTypeReferences {
			t.Errorf("an absent configuration switched an option off: %+v", settings)
		}
		if settings.allowNamedExports {
			t.Errorf("an absent configuration switched allowNamedExports on: %+v", settings)
		}
	}

	// And the rule must actually report through that path, not merely hold the right settings.
	result := rule_testing.RunTypedWithOptions(t, NoUseBeforeDefine, "fixture.ts", "a++; var a=19;", nil)
	rule_testing.ExpectFindings(t, result, "usedBeforeDefined")
}

// TestNoUseBeforeDefineRequiresTheTypedHarness pins the checker guard.
//
// NeedsTypeChecker governs the registration path and says nothing about a Context built by hand,
// which TestNoRegisteredRuleCrashesOnAbsentOptionalNodes does. Measured on this rule: through the
// untyped harness GetSymbolAtLocation would be reached on a nil checker, so the guard is what stands
// between the rule and either a panic or, worse, a vacuous silence. This case asserts the decline is
// silent rather than fatal, so a later revert of the guard fails loudly here.
func TestNoUseBeforeDefineRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUseBeforeDefine, "fixture.ts", "a++; var a=19;")
	rule_testing.ExpectClean(t, result)

	// The control: the same source through the typed harness must report, or the case above is
	// passing because the rule is broken rather than because the guard fired.
	typed := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", "a++; var a=19;")
	rule_testing.ExpectFindings(t, typed, "usedBeforeDefined")
}

// TestNoUseBeforeDefineSubstrateDivergences records the five imported cases this port does not
// report, and why each one is the substrate rather than the rule.
//
// All five share a single measured cause: `GetSymbolAtLocation` returns nil for the reference, so
// there is no binding to compare a position against and the rule has nothing to judge. That was
// established with a probe printing the resolution result for each shape, alongside controls that do
// resolve, rather than inferred from the silence.
//
// They are kept here as reporting-zero rather than deleted, because a deleted case is indis-
// tinguishable from a case nobody imported. If the checker ever resolves these, this test fails and
// names the rows to move back into the Fires table.
//
// The `with` family. TypeScript declines to resolve identifiers inside a `with` block at all,
// because the object supplies bindings dynamically and no static analysis can say whether `x` names
// the later `let` or a property of `obj`. That refusal is correct, and upstream reports these only
// because eslint-scope resolves lexically and ignores the dynamic scope entirely. Reporting them
// here would require asserting a binding the checker explicitly declines to name.
//
// The block-scoped function. A function declared inside a block is not visible outside it under
// ES2015 semantics, so the checker resolves the outer call to nothing. Upstream reports it because
// eslint-scope models the legacy web-compatibility hoisting that predates block scoping. The control
// separating these is `a(); function a() {}` at top level, which resolves and which this rule
// reports.
func TestNoUseBeforeDefineSubstrateDivergences(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{"with statement, bare body", "with (obj) x; let x = {}"},
		{"with statement, block body", "with (obj) { x } let x = {}"},
		{"with statement, nested block", "with (obj) { if (a) { x } } let x = {}"},
		{"with statement, arrow inside", "with (obj) { (() => { if (a) { x } })() } let x = {}"},
		{"function declared in a block", `"use strict"; a(); { function a() {} }`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}

	// The control. Without it every row above passes for any reason at all, including a rule that
	// reports nothing anywhere, which is the failure mode this whole file exists to catch.
	control := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", "a(); function a() {}")
	rule_testing.ExpectFindings(t, control, "usedBeforeDefined")
}

// TestNoUseBeforeDefineMergedDeclarations pins which declaration of a merged symbol the rule judges
// against, which the imported corpus does not cover at all.
//
// A merged symbol carries several declarations and their order in `symbol.Declarations` is NOT
// source order. Measured with a probe on this tree: for `interface Foo {} function Foo() {}` the
// checker returns the FUNCTION first even though the interface is written above it, so an
// index-zero read compares the reference against a position 16 bytes later than the real one.
//
// The rule takes the earliest by position instead, because the question it asks is "when does this
// name become usable" and the answer is at the first declaration. Upstream was driven on all three
// shapes below through the ESLint Linter API with the typescript-eslint parser, and reports every
// one of them at column 1; this rule agrees only because of that choice. A mutation replacing it
// with `Declarations[0]` survives the entire 349-row corpus, which is why these cases exist.
func TestNoUseBeforeDefineMergedDeclarations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		// The one that separates the two readings: the value declaration sorts ahead of the type
		// declaration written above it, so index zero and earliest-by-position disagree.
		{"interface above function", "Foo; interface Foo {} function Foo() {}"},
		{"function above interface", "Foo; function Foo() {} interface Foo {}"},
		{"enum declared twice", "Foo; enum Foo { A } enum Foo { B }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", testCase.source)
			rule_testing.ExpectFindings(t, result, "usedBeforeDefined")

			// The span matters as much as the count here. Reporting once against the wrong
			// declaration produces the same id and the same total.
			if len(result.Diagnostics) == 1 && result.Diagnostics[0].Range.Pos() != 0 {
				t.Errorf("finding starts at %d, want 0", result.Diagnostics[0].Range.Pos())
			}
		})
	}

	// The rows above establish that a merged symbol reports at all. They do NOT separate the two
	// readings, and finding that out cost a round: with the reference above BOTH declarations, index
	// zero and earliest-by-position give different declarations and the same verdict, because every
	// declaration is below the reference either way.
	//
	// The input that separates them puts the reference BETWEEN the two declarations. Upstream was
	// driven on each of these and reports zero: the binding is usable from the first declaration, so
	// a reference after it is fine even though a later declaration of the same name follows. Reading
	// index zero here compares against the LATER declaration and reports a false positive.
	for _, clean := range []struct {
		name   string
		source string
	}{
		{"between interface and function", "interface Foo {} Foo; function Foo() {}"},
		{"between interface and function, type position", "interface Foo {} type B = Foo; function Foo() {}"},
		{"between two enum declarations", "enum Foo { A } Foo; enum Foo { B }"},
		{"between class and namespace", "class Foo {} Foo; namespace Foo { export let x = 1; }"},
	} {
		t.Run(clean.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", clean.source)
			rule_testing.ExpectClean(t, result)
		})
	}

	// The control for the block above: the same merged shape with the reference below EVERY
	// declaration is also clean, so a rule reporting nothing at all would pass those four rows. The
	// reporting rows earlier in this function are what make the pair meaningful.
	control := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", "interface Foo {} function Foo() {} Foo;")
	rule_testing.ExpectClean(t, control)
}

// TestNoUseBeforeDefineQualifiedNames pins the interaction between a namespace path and the
// ignoreTypeReferences option, which the imported corpus does not exercise and which two plausible
// readings of upstream answer differently.
//
// Upstream's exemption is an IMMEDIATE-parent test: `identifier.parent.type === "TSTypeReference"`.
// So `Foo` in `let x: Foo` is exempt and `A` in `let x: A.B` is not, because `A`'s parent is the
// qualified name rather than the type reference. Walking upward through the qualified name reads as
// the obvious correction and is wrong, and it fails silently: it exempts every namespaced type
// reference in the tree while the entire 349-row corpus stays green, because upstream writes no
// case combining a namespace path with a forward reference.
//
// Every expectation below was measured by driving the installed ESLint rule through the Linter API
// with the typescript-eslint parser, not read off the source.
func TestNoUseBeforeDefineQualifiedNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		findings int
	}{
		// The direct type reference IS exempt: the identifier's parent is the type reference itself.
		{"plain type reference is exempt", "let x: Foo; type Foo = 1;", 0},

		// The namespaced ones are NOT, and this is the whole point of the immediate-parent test.
		{"namespaced type reference reports", "let x: A.B; namespace A { export type B = 1 }", 1},
		{"deeper namespaced type reference reports", "let x: A.B.C; namespace A { export namespace B { export type C = 1 } }", 1},

		// A value-position namespace path reports for the ordinary reason, and only its leftmost
		// segment names a binding: `B` and `C` are members of `A`, not bindings of their own.
		{"namespaced value reference reports once", "A.B.C; namespace A { export namespace B { export const C = 1 } }", 1},
		{"member access on a namespace reports once", "const y = A.B; namespace A { export const B = 1 }", 1},

		// A type query is exempt whatever its shape, which is the other half of the option and a
		// separate predicate from the type-reference test.
		{"type query over a namespace path is exempt", "let x: typeof A.B; namespace A { export const B = 1 }", 0},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", testCase.source)
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "usedBeforeDefined"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoUseBeforeDefineJsxAttributes separates a JSX attribute NAME, which names nothing, from a JSX
// component reference, which does.
//
// The corpus cannot express this difference. eslint-scope creates no reference for an attribute
// name, so upstream never had a reason to write a case, and its eleven JSX rows are all about
// component names. Our checker resolves an attribute name to whatever declaration the surrounding
// types associate with it, which reports whenever that declaration sits later in the file. Measured
// on the real tree before the guard existed: 1,671 findings, with the two largest clusters being an
// icon file full of `key=` and a component file full of ordinary props.
//
// Every expectation was measured against the installed ESLint rule with the typescript-eslint parser
// and jsx enabled.
func TestNoUseBeforeDefineJsxAttributes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		findings int
	}{
		// Attribute names are not bindings, whatever they resolve to.
		{"key attribute is not a reference", `const a = <div key="x" />; export {};`, 0},
		{"several attributes are not references", `const a = <div className="x" id="y" />; export {};`, 0},

		// Component references ARE, and this is the pair the guard has to preserve. Losing these
		// would make the guard look correct while switching the rule off for every JSX file.
		{"component used above its declaration", "const a = <Section />; function Section() { return <div/>; } export {};", 1},
		{"component used from a function above its declaration", "function App() { return <Section/>; } function Section() { return <div/>; } export {};", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.tsx", testCase.source)
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "usedBeforeDefined"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoUseBeforeDefineShorthandProperties pins the accessor a shorthand property needs.
//
// `GetSymbolAtLocation` on the `form` in `{ form }` resolves to the PROPERTY's own symbol, whose
// only declaration is the shorthand node itself, so comparing positions against it compares the
// reference with itself. The recovery is GetShorthandAssignmentValueSymbol.
//
// The corpus writes no shorthand property anywhere in its 354 cases, so nothing imported can see
// this. On the real tree it was 3,243 findings, and the first three read at source were all this.
func TestNoUseBeforeDefineShorthandProperties(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		findings int
	}{
		// The pair that makes the defect visible: same binding, same line, different spelling.
		{"shorthand reads an earlier binding", "const form = 1; const o = { form };", 0},
		{"longhand reads an earlier binding", "const form = 1; const o = { form: form };", 0},
		{"shorthand beside another property", "const a = 1; const o = { a, b: 2 };", 0},
		{"shorthand inside a function", "function f() { const form = 1; return { form }; }", 0},

		// The control: a shorthand reading a LATER binding is a genuine forward reference and must
		// still report, or the guard above has switched the rule off for object literals.
		{"shorthand reads a later binding", "const o = { form }; const form = 1;", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", testCase.source)
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "usedBeforeDefined"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoUseBeforeDefineDestructuringPropertyNames separates the two halves of a renamed
// destructuring pattern.
//
// In `const { sorted: sortedTypes } = f()`, `sorted` names a property of the right-hand value and
// `sortedTypes` is the binding. Our parser keeps the property name in its own slot on the binding
// element, and the checker resolves it to whatever declares that property, which is often later in
// the file, so without a guard the rule reports the property name as a forward reference.
//
// `isDeclaringName` cannot cover it: that helper already answers true for the binding NAME of a
// binding element, and the property name is a different slot on the same node. Upstream's corpus
// writes destructuring DEFAULTS (`var {a = 0, b = a} = {}`) but never a rename, so nothing imported
// can see this. It was found by differencing the real tree against the installed rule.
//
// Every count below was measured against that rule.
func TestNoUseBeforeDefineDestructuringPropertyNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		findings int
	}{
		// One finding each, on the BINDING or on the source value, never on the property name.
		{"renamed binding from a later function", "const { sorted: sortedTypes } = f(); function f() { return { sorted: 1 }; }", 1},
		{"shorthand binding from a later function", "const { sorted } = f(); function f() { return { sorted: 1 }; }", 1},
		{"two renamed bindings from a later object", "const { a: b, c: d } = o; const o = { a: 1, c: 2 };", 1},

		// The control: destructuring that reads nothing declared later is clean, so a guard that
		// switched the rule off for binding elements entirely would not pass the rows above.
		{"renamed binding from an earlier value", "const o = { a: 1 }; const { a: b } = o;", 0},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", testCase.source)
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "usedBeforeDefined"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoUseBeforeDefineObjectMethodParameters pins the sentinel set against a parser difference.
//
// Upstream's walk stops at
// `(Function|Class)(Declaration|Expression)|ArrowFunctionExpression|CatchClause|Import...`, matched
// against ESTree, where an object literal's method, a class method, an accessor and a constructor
// are ALL FunctionExpression nodes. Our parser gives each its own kind, so translating that regex
// literally misses every one of them.
//
// What it costs is not subtle: a parameter of an object method walks past the method, past the
// object literal, and reaches the `const` the object is assigned to, whose initializer contains the
// whole method body, so every use of every such parameter reads as evaluated during that variable's
// initialization. On the real tree that was 44 findings in one test file alone.
//
// The pair that makes it visible is the last two rows: byte-equivalent code, one spelled as a
// shorthand method and one as a function expression, which upstream's parser renders identically and
// ours does not.
func TestNoUseBeforeDefineObjectMethodParameters(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"const o = { start(controller) { controller.close(); } };",
		"function f(controller) { controller.close(); }",
		"const f = function(controller) { controller.close(); };",
		"const o = { start: function(c) { c.x(); } };",
		"class K { m(c) { c.x(); } }",
		"class K { get p() { return 1; } set p(v) { this.q = v; } }",
		"class K { constructor(c) { this.c = c; } }",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", source)
			rule_testing.ExpectClean(t, result)
		})
	}

	// The control: a parameter is clean, but a genuine forward reference from inside an object
	// method still reports. Without this, dropping the whole initialization arm would pass the rows
	// above.
	control := rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", "const o = { start() { return later; } }; const later = 1;")
	rule_testing.ExpectFindings(t, control, "usedBeforeDefined")
}
