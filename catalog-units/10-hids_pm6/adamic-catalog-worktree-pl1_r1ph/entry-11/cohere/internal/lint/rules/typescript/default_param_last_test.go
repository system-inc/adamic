package typescript

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const defaultParamLastFile = "/repository/source/Parameters.ts"

// TestDefaultParamLastFires carries all 45 failing cases from the typescript-eslint corpus
// verbatim, plus cases pinning decisions the corpus never exercises. The corpus states a messageId
// per finding, so the counts here are read off rather than recovered: 45 inputs producing 54
// findings. Every case was replayed through the installed @typescript-eslint build before it was
// written here, and all 45 agreed with the corpus on count and on span.
func TestDefaultParamLastFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messageIds []string
	}{
		// The corpus, verbatim.
		{"function foo(a = 1, b: number) {}",
			"function foo(a = 1, b: number) {}",
			[]string{"shouldBeLast"}},
		{"function foo(a = 1, b = 2, c: number) {}",
			"function foo(a = 1, b = 2, c: number) {}",
			[]string{"shouldBeLast", "shouldBeLast"}},
		{"function foo(a = 1, b: number, c = 2, d: number) {}",
			"function foo(a = 1, b: number, c = 2, d: number) {}",
			[]string{"shouldBeLast", "shouldBeLast"}},
		{"function foo(a = 1, b: number, c = 2) {}",
			"function foo(a = 1, b: number, c = 2) {}",
			[]string{"shouldBeLast"}},
		{"function foo(a = 1, b: number, ...c) {}",
			"function foo(a = 1, b: number, ...c) {}",
			[]string{"shouldBeLast"}},
		{"function foo(a?: number, b: number) {}",
			"function foo(a?: number, b: number) {}",
			[]string{"shouldBeLast"}},
		{"function foo(a: number, b?: number, c: number) {}",
			"function foo(a: number, b?: number, c: number) {}",
			[]string{"shouldBeLast"}},
		{"function foo(a = 1, b?: number, c: number) {}",
			"function foo(a = 1, b?: number, c: number) {}",
			[]string{"shouldBeLast", "shouldBeLast"}},
		{"function foo(a = 1, { b }) {}",
			"function foo(a = 1, { b }) {}",
			[]string{"shouldBeLast"}},
		{"function foo({ a } = {}, b) {}",
			"function foo({ a } = {}, b) {}",
			[]string{"shouldBeLast"}},
		{"function foo({ a, b } = { a: 1, b: 2 }, c) {}",
			"function foo({ a, b } = { a: 1, b: 2 }, c) {}",
			[]string{"shouldBeLast"}},
		{"function foo([a] = [], b) {}",
			"function foo([a] = [], b) {}",
			[]string{"shouldBeLast"}},
		{"function foo([a, b] = [1, 2], c) {}",
			"function foo([a, b] = [1, 2], c) {}",
			[]string{"shouldBeLast"}},
		{"const foo = function (a = 1, b: number) {};",
			"const foo = function (a = 1, b: number) {};",
			[]string{"shouldBeLast"}},
		{"const foo = function (a = 1, b = 2, c: number) {};",
			"const foo = function (a = 1, b = 2, c: number) {};",
			[]string{"shouldBeLast", "shouldBeLast"}},
		{"const foo = function (a = 1, b: number, c = 2, d: number) {};",
			"const foo = function (a = 1, b: number, c = 2, d: number) {};",
			[]string{"shouldBeLast", "shouldBeLast"}},
		{"const foo = function (a = 1, b: number, c = 2) {};",
			"const foo = function (a = 1, b: number, c = 2) {};",
			[]string{"shouldBeLast"}},
		{"const foo = function (a = 1, b: number, ...c) {};",
			"const foo = function (a = 1, b: number, ...c) {};",
			[]string{"shouldBeLast"}},
		{"const foo = function (a?: number, b: number) {};",
			"const foo = function (a?: number, b: number) {};",
			[]string{"shouldBeLast"}},
		{"const foo = function (a: number, b?: number, c: number) {};",
			"const foo = function (a: number, b?: number, c: number) {};",
			[]string{"shouldBeLast"}},
		{"const foo = function (a = 1, b?: number, c: number) {};",
			"const foo = function (a = 1, b?: number, c: number) {};",
			[]string{"shouldBeLast", "shouldBeLast"}},
		{"const foo = function (a = 1, { b }) {};",
			"const foo = function (a = 1, { b }) {};",
			[]string{"shouldBeLast"}},
		{"const foo = function ({ a } = {}, b) {};",
			"const foo = function ({ a } = {}, b) {};",
			[]string{"shouldBeLast"}},
		{"const foo = function ({ a, b } = { a: 1, b: 2 }, c) {};",
			"const foo = function ({ a, b } = { a: 1, b: 2 }, c) {};",
			[]string{"shouldBeLast"}},
		{"const foo = function ([a] = [], b) {};",
			"const foo = function ([a] = [], b) {};",
			[]string{"shouldBeLast"}},
		{"const foo = function ([a, b] = [1, 2], c) {};",
			"const foo = function ([a, b] = [1, 2], c) {};",
			[]string{"shouldBeLast"}},
		{"const foo = (a = 1, b: number) => {};",
			"const foo = (a = 1, b: number) => {};",
			[]string{"shouldBeLast"}},
		{"const foo = (a = 1, b = 2, c: number) => {};",
			"const foo = (a = 1, b = 2, c: number) => {};",
			[]string{"shouldBeLast", "shouldBeLast"}},
		{"const foo = (a = 1, b: number, c = 2, d: number) => {};",
			"const foo = (a = 1, b: number, c = 2, d: number) => {};",
			[]string{"shouldBeLast", "shouldBeLast"}},
		{"const foo = (a = 1, b: number, c = 2) => {};",
			"const foo = (a = 1, b: number, c = 2) => {};",
			[]string{"shouldBeLast"}},
		{"const foo = (a = 1, b: number, ...c) => {};",
			"const foo = (a = 1, b: number, ...c) => {};",
			[]string{"shouldBeLast"}},
		{"const foo = (a?: number, b: number) => {};",
			"const foo = (a?: number, b: number) => {};",
			[]string{"shouldBeLast"}},
		{"const foo = (a: number, b?: number, c: number) => {};",
			"const foo = (a: number, b?: number, c: number) => {};",
			[]string{"shouldBeLast"}},
		{"const foo = (a = 1, b?: number, c: number) => {};",
			"const foo = (a = 1, b?: number, c: number) => {};",
			[]string{"shouldBeLast", "shouldBeLast"}},
		{"const foo = (a = 1, { b }) => {};",
			"const foo = (a = 1, { b }) => {};",
			[]string{"shouldBeLast"}},
		{"const foo = ({ a } = {}, b) => {};",
			"const foo = ({ a } = {}, b) => {};",
			[]string{"shouldBeLast"}},
		{"const foo = ({ a, b } = { a: 1, b: 2 }, c) => {};",
			"const foo = ({ a, b } = { a: 1, b: 2 }, c) => {};",
			[]string{"shouldBeLast"}},
		{"const foo = ([a] = [], b) => {};",
			"const foo = ([a] = [], b) => {};",
			[]string{"shouldBeLast"}},
		{"const foo = ([a, b] = [1, 2], c) => {};",
			"const foo = ([a, b] = [1, 2], c) => {};",
			[]string{"shouldBeLast"}},
		{"class Foo { constructor( public a: number, protected b?: number, private c:...",
			"\nclass Foo {\n  constructor(\n    public a: number,\n    protected b?: number,\n    private c: number,\n  ) {}\n}\n      ",
			[]string{"shouldBeLast"}},
		{"class Foo { constructor( public a: number, protected b = 0, private c: numb...",
			"\nclass Foo {\n  constructor(\n    public a: number,\n    protected b = 0,\n    private c: number,\n  ) {}\n}\n      ",
			[]string{"shouldBeLast"}},
		{"class Foo { constructor( public a?: number, private b: number, ) {} }",
			"\nclass Foo {\n  constructor(\n    public a?: number,\n    private b: number,\n  ) {}\n}\n      ",
			[]string{"shouldBeLast"}},
		{"class Foo { constructor( public a = 0, private b: number, ) {} }",
			"\nclass Foo {\n  constructor(\n    public a = 0,\n    private b: number,\n  ) {}\n}\n      ",
			[]string{"shouldBeLast"}},
		{"class Foo { constructor(a = 0, b: number) {} }",
			"\nclass Foo {\n  constructor(a = 0, b: number) {}\n}\n      ",
			[]string{"shouldBeLast"}},
		{"class Foo { constructor(a?: number, b: number) {} }",
			"\nclass Foo {\n  constructor(a?: number, b: number) {}\n}\n      ",
			[]string{"shouldBeLast"}},
		// Beyond the corpus. Each was measured against the installed rule before being written.

		// A class method is a MethodDefinition whose value is a FunctionExpression upstream, so
		// upstream's FunctionExpression listener reaches it. Our parser gives it its own kind and
		// it needs its own listener. The corpus writes no class method other than a constructor,
		// so nothing imported can see a port that omits the arm. Measured: reports at col 20.
		{"a class method", "class Foo { method(a = 1, b: number) {} }", []string{"shouldBeLast"}},
		// A static class method, to pin that staticness is not a discrimination. Upstream reads no
		// static flag. Measured: reports at col 22.
		{"a static class method", "class Foo { static m(a = 1, b: number) {} }", []string{"shouldBeLast"}},
		// An object literal method, the other shape our parser gives KindMethodDeclaration.
		// Measured: reports at col 15.
		{"an object literal method", "const o = { m(a = 1, b: number) {} };", []string{"shouldBeLast"}},
		// A function expression assigned to an object property, which is upstream's own
		// FunctionExpression shape. Measured: reports at col 26.
		{"a function expression in an object literal",
			"const o = { m: function (a = 1, b: number) {} };", []string{"shouldBeLast"}},
		// A generator and an async function, to pin that neither modifier changes the judgment.
		// Both are the same node kind here and upstream. Measured: both report.
		{"a generator function", "function* gen(a = 1, b: number) {}", []string{"shouldBeLast"}},
		{"an async function", "async function foo(a = 1, b: number) {}", []string{"shouldBeLast"}},
		{"an async arrow function", "const foo = async (a = 1, b: number) => {};", []string{"shouldBeLast"}},
		// A `this` parameter is an ordinary parameter in both trees and counts as plain, so it
		// neither reports nor exempts what follows. Measured: reports on `a = 1` at col 28, which
		// is the second parameter, so the `this` parameter was passed over silently.
		{"a this parameter ahead of the defaulted one",
			"function foo(this: Window, a = 1, b: number) {}", []string{"shouldBeLast"}},
		// Nested functions each get their own judgment. Measured: two findings.
		{"a nested function with its own violation",
			"function outer(x = 1, y) { function inner(p = 1, q) {} }",
			[]string{"shouldBeLast", "shouldBeLast"}},
		// A parameter carrying both a type annotation and a default. The annotation is not part of
		// the classification and the span covers it. Measured: reports spanning `a: number = 1`.
		{"a defaulted parameter with a type annotation",
			"function foo(a: number = 1, b: number) {}", []string{"shouldBeLast"}},
		// An optional parameter property, which combines the wrapper span with the question-token
		// classification. The corpus writes optional and defaulted parameter properties separately
		// but never an optional one with a `private` modifier. Measured: spans `private a?: number`.
		{"an optional parameter property",
			"class Foo { constructor(private a?: number, b: number) {} }", []string{"shouldBeLast"}},
		// Four violations in one list, alternating, to pin that the sweep does not stop at the
		// first finding and that the region boundary is the LAST plain parameter rather than the
		// first. Measured: two findings, on `a = 1` and `c = 2`, with `e = 3` clean because no
		// plain parameter follows it.
		{"three defaults straddling two required parameters",
			"function foo(a = 1, b: number, c = 2, d: number, e = 3) {}",
			[]string{"shouldBeLast", "shouldBeLast"}},

		// An accessor with an ILLEGAL parameter list. These four exist because a mutation emptying
		// the getter and setter arms survived the whole suite, and the reason it survived was that
		// the reasoning behind those arms was wrong rather than that nothing could reach them.
		//
		// The argument that looked airtight: a setter takes exactly one parameter and a getter
		// none, so no plain parameter can follow a defaulted one and the loop body never runs.
		// That is the GRAMMAR. Our parser recovers from source that violates it and attaches the
		// extra parameters anyway, so a two-parameter setter is a live node here. Probed directly:
		// all four inputs below give the accessor two parameters and a body.
		//
		// Upstream reports every one of them, measured against the installed build, because its
		// parser recovers the same way and the node is a FunctionExpression to its listener. So
		// the arms are load-bearing and were always correct; only the fixtures were blind.
		{"a setter with two parameters, the first defaulted",
			"class Foo { set x(a = 1, b: number) {} }", []string{"shouldBeLast"}},
		{"a getter with two parameters, the first defaulted",
			"class Foo { get x(a = 1, b: number) { return 1; } }", []string{"shouldBeLast"}},
		{"an object literal setter with two parameters",
			"const o = { set x(a = 1, b: number) {} };", []string{"shouldBeLast"}},
		{"an object literal getter with two parameters",
			"const o = { get x(a = 1, b: number) { return 1; } };", []string{"shouldBeLast"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, DefaultParamLast, defaultParamLastFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestDefaultParamLastStaysSilent carries all 42 passing cases from the typescript-eslint corpus
// verbatim, plus the clean side of every decision the fires table exercises. Upstream's clean cases
// are the false positives it already thought about; the added ones cover the two places our tree
// differs from the one upstream reads.
func TestDefaultParamLastStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The corpus, verbatim.
		{"function foo() {}",
			"function foo() {}"},
		{"function foo(a: number) {}",
			"function foo(a: number) {}"},
		{"function foo(a = 1) {}",
			"function foo(a = 1) {}"},
		{"function foo(a?: number) {}",
			"function foo(a?: number) {}"},
		{"function foo(a: number, b: number) {}",
			"function foo(a: number, b: number) {}"},
		{"function foo(a: number, b: number, c?: number) {}",
			"function foo(a: number, b: number, c?: number) {}"},
		{"function foo(a: number, b = 1) {}",
			"function foo(a: number, b = 1) {}"},
		{"function foo(a: number, b = 1, c = 1) {}",
			"function foo(a: number, b = 1, c = 1) {}"},
		{"function foo(a: number, b = 1, c?: number) {}",
			"function foo(a: number, b = 1, c?: number) {}"},
		{"function foo(a: number, b?: number, c = 1) {}",
			"function foo(a: number, b?: number, c = 1) {}"},
		{"function foo(a: number, b = 1, ...c) {}",
			"function foo(a: number, b = 1, ...c) {}"},
		{"const foo = function () {};",
			"const foo = function () {};"},
		{"const foo = function (a: number) {};",
			"const foo = function (a: number) {};"},
		{"const foo = function (a = 1) {};",
			"const foo = function (a = 1) {};"},
		{"const foo = function (a?: number) {};",
			"const foo = function (a?: number) {};"},
		{"const foo = function (a: number, b: number) {};",
			"const foo = function (a: number, b: number) {};"},
		{"const foo = function (a: number, b: number, c?: number) {};",
			"const foo = function (a: number, b: number, c?: number) {};"},
		{"const foo = function (a: number, b = 1) {};",
			"const foo = function (a: number, b = 1) {};"},
		{"const foo = function (a: number, b = 1, c = 1) {};",
			"const foo = function (a: number, b = 1, c = 1) {};"},
		{"const foo = function (a: number, b = 1, c?: number) {};",
			"const foo = function (a: number, b = 1, c?: number) {};"},
		{"const foo = function (a: number, b?: number, c = 1) {};",
			"const foo = function (a: number, b?: number, c = 1) {};"},
		{"const foo = function (a: number, b = 1, ...c) {};",
			"const foo = function (a: number, b = 1, ...c) {};"},
		{"const foo = () => {};",
			"const foo = () => {};"},
		{"const foo = (a: number) => {};",
			"const foo = (a: number) => {};"},
		{"const foo = (a = 1) => {};",
			"const foo = (a = 1) => {};"},
		{"const foo = (a?: number) => {};",
			"const foo = (a?: number) => {};"},
		{"const foo = (a: number, b: number) => {};",
			"const foo = (a: number, b: number) => {};"},
		{"const foo = (a: number, b: number, c?: number) => {};",
			"const foo = (a: number, b: number, c?: number) => {};"},
		{"const foo = (a: number, b = 1) => {};",
			"const foo = (a: number, b = 1) => {};"},
		{"const foo = (a: number, b = 1, c = 1) => {};",
			"const foo = (a: number, b = 1, c = 1) => {};"},
		{"const foo = (a: number, b = 1, c?: number) => {};",
			"const foo = (a: number, b = 1, c?: number) => {};"},
		{"const foo = (a: number, b?: number, c = 1) => {};",
			"const foo = (a: number, b?: number, c = 1) => {};"},
		{"const foo = (a: number, b = 1, ...c) => {};",
			"const foo = (a: number, b = 1, ...c) => {};"},
		{"class Foo { constructor(a: number, b: number, c: number) {} }",
			"\nclass Foo {\n  constructor(a: number, b: number, c: number) {}\n}\n    "},
		{"class Foo { constructor(a: number, b?: number, c = 1) {} }",
			"\nclass Foo {\n  constructor(a: number, b?: number, c = 1) {}\n}\n    "},
		{"class Foo { constructor(a: number, b = 1, c?: number) {} }",
			"\nclass Foo {\n  constructor(a: number, b = 1, c?: number) {}\n}\n    "},
		{"class Foo { constructor( public a: number, protected b: number, private c: ...",
			"\nclass Foo {\n  constructor(\n    public a: number,\n    protected b: number,\n    private c: number,\n  ) {}\n}\n    "},
		{"class Foo { constructor( public a: number, protected b?: number, private c ...",
			"\nclass Foo {\n  constructor(\n    public a: number,\n    protected b?: number,\n    private c = 10,\n  ) {}\n}\n    "},
		{"class Foo { constructor( public a: number, protected b = 10, private c?: nu...",
			"\nclass Foo {\n  constructor(\n    public a: number,\n    protected b = 10,\n    private c?: number,\n  ) {}\n}\n    "},
		{"class Foo { constructor( a: number, protected b?: number, private c = 0, ) ...",
			"\nclass Foo {\n  constructor(\n    a: number,\n    protected b?: number,\n    private c = 0,\n  ) {}\n}\n    "},
		{"class Foo { constructor( a: number, b?: number, private c = 0, ) {} }",
			"\nclass Foo {\n  constructor(\n    a: number,\n    b?: number,\n    private c = 0,\n  ) {}\n}\n    "},
		{"class Foo { constructor( a: number, private b?: number, c = 0, ) {} }",
			"\nclass Foo {\n  constructor(\n    a: number,\n    private b?: number,\n    c = 0,\n  ) {}\n}\n    "},
		// Beyond the corpus, and this block is the substance of the port.
		//
		// Every bodyless function-like construct is a DIFFERENT node type in the tree upstream
		// reads, so none of them reaches its three listeners. Our parser draws no such distinction
		// for the first four: a `declare function` is an ordinary KindFunctionDeclaration whose
		// Body() is nil, and an overload signature is an ordinary KindConstructor whose Body() is
		// nil. Without the nil-body guard the rule reports every ambient declaration in the tree,
		// and NOT ONE of the 42 clean corpus cases above can see it, because the corpus writes no
		// ambient declaration anywhere. All six were measured clean against the installed rule.
		{"an ambient function declaration",
			"declare function foo(a = 1, b: number): void;"},
		{"a constructor overload signature",
			"class Foo { constructor(a = 1, b: number); constructor(a?: any, b?: any) {} }"},
		{"an abstract method signature",
			"abstract class Foo { abstract m(a = 1, b: number): void; }"},
		{"a method in an ambient class",
			"declare class Foo { m(a = 1, b: number): void; }"},
		// These last two our parser gives their own kinds, so they are unreachable through the
		// listeners rather than declined by the guard. They are pinned anyway, because a later
		// reader adding a KindMethodSignature or KindFunctionType arm would be adding a divergence
		// and this is where it fails.
		{"a method signature in an interface",
			"interface I { m(a?: number, b: number): void }"},
		{"a function type alias",
			"type T = (a?: number, b: number) => void;"},

		// An accessor is a FunctionExpression upstream and is visited. These two are the LEGAL
		// shapes, where no violation is expressible: a setter takes exactly one parameter and a
		// getter none, so no plain parameter can follow a defaulted one. The illegal shapes, which
		// our parser does produce and which do report, are in the fires table above. Measured:
		// both of these clean.
		{"a setter with a defaulted parameter", "class Foo { set x(v = 1) {} }"},
		{"a getter", "class Foo { get x(): number { return 1; } }"},

		// A rest parameter is neither plain nor reportable, so it sets no flag and is never
		// condemned. The corpus covers the reportable direction (`a = 1, b: number, ...c` reports);
		// these cover the other. Measured: all three clean.
		{"a default followed only by a rest", "function foo(a = 1, ...b) {}"},
		{"an optional followed only by a rest", "function foo(a?: number, ...b) {}"},
		{"a default, an optional, and a rest", "function foo(a = 1, b?, ...c) {}"},

		// The clean side of the multi-violation case: every default sits after every required
		// parameter, so the sweep's region is empty. Measured: clean.
		{"defaults only after every required parameter",
			"function foo(a: number, b: number, c = 1, d = 2) {}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, DefaultParamLast, defaultParamLastFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestDefaultParamLastSpans asserts where each finding points, by slicing the source the harness
// actually wrote and comparing the text.
//
// The span is the load-bearing half of this port. Upstream reports the OUTER TSParameterProperty
// node for a parameter property, so its finding covers the `public` or `readonly` keyword;
// typescript-go has no wrapper and the modifiers live on the parameter, so the same span arrives
// for a different structural reason and only an assertion records that the two agree. Nothing in
// ExpectFindings can see it: every one of these cases carries the same single message id.
//
// The expected texts are the spans MEASURED against the installed @typescript-eslint build over the
// same 45 inputs, converted from line and column to byte offsets over the trimmed source. All 45
// agreed with the corpus's own stated columns.
//
// rule_testing.Run does not trim, so the literal here and the file on disk are the same bytes and the
// slice is not offset. The typed harness would trim; this rule needs no checker.
func TestDefaultParamLastSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		{"function foo(a = 1, b: number) {} (2)",
			"function foo(a = 1, b: number) {}",
			[]string{"a = 1"}},
		{"function foo(a = 1, b = 2, c: number) {} (2)",
			"function foo(a = 1, b = 2, c: number) {}",
			[]string{"a = 1", "b = 2"}},
		{"function foo(a = 1, b: number, c = 2, d: number) {} (2)",
			"function foo(a = 1, b: number, c = 2, d: number) {}",
			[]string{"a = 1", "c = 2"}},
		{"function foo(a = 1, b: number, c = 2) {} (2)",
			"function foo(a = 1, b: number, c = 2) {}",
			[]string{"a = 1"}},
		{"function foo(a = 1, b: number, ...c) {} (2)",
			"function foo(a = 1, b: number, ...c) {}",
			[]string{"a = 1"}},
		{"function foo(a?: number, b: number) {} (2)",
			"function foo(a?: number, b: number) {}",
			[]string{"a?: number"}},
		{"function foo(a: number, b?: number, c: number) {} (2)",
			"function foo(a: number, b?: number, c: number) {}",
			[]string{"b?: number"}},
		{"function foo(a = 1, b?: number, c: number) {} (2)",
			"function foo(a = 1, b?: number, c: number) {}",
			[]string{"a = 1", "b?: number"}},
		{"function foo(a = 1, { b }) {} (2)",
			"function foo(a = 1, { b }) {}",
			[]string{"a = 1"}},
		{"function foo({ a } = {}, b) {} (2)",
			"function foo({ a } = {}, b) {}",
			[]string{"{ a } = {}"}},
		{"function foo({ a, b } = { a: 1, b: 2 }, c) {} (2)",
			"function foo({ a, b } = { a: 1, b: 2 }, c) {}",
			[]string{"{ a, b } = { a: 1, b: 2 }"}},
		{"function foo([a] = [], b) {} (2)",
			"function foo([a] = [], b) {}",
			[]string{"[a] = []"}},
		{"function foo([a, b] = [1, 2], c) {} (2)",
			"function foo([a, b] = [1, 2], c) {}",
			[]string{"[a, b] = [1, 2]"}},
		{"const foo = function (a = 1, b: number) {}; (2)",
			"const foo = function (a = 1, b: number) {};",
			[]string{"a = 1"}},
		{"const foo = function (a = 1, b = 2, c: number) {}; (2)",
			"const foo = function (a = 1, b = 2, c: number) {};",
			[]string{"a = 1", "b = 2"}},
		{"const foo = function (a = 1, b: number, c = 2, d: number) {}; (2)",
			"const foo = function (a = 1, b: number, c = 2, d: number) {};",
			[]string{"a = 1", "c = 2"}},
		{"const foo = function (a = 1, b: number, c = 2) {}; (2)",
			"const foo = function (a = 1, b: number, c = 2) {};",
			[]string{"a = 1"}},
		{"const foo = function (a = 1, b: number, ...c) {}; (2)",
			"const foo = function (a = 1, b: number, ...c) {};",
			[]string{"a = 1"}},
		{"const foo = function (a?: number, b: number) {}; (2)",
			"const foo = function (a?: number, b: number) {};",
			[]string{"a?: number"}},
		{"const foo = function (a: number, b?: number, c: number) {}; (2)",
			"const foo = function (a: number, b?: number, c: number) {};",
			[]string{"b?: number"}},
		{"const foo = function (a = 1, b?: number, c: number) {}; (2)",
			"const foo = function (a = 1, b?: number, c: number) {};",
			[]string{"a = 1", "b?: number"}},
		{"const foo = function (a = 1, { b }) {}; (2)",
			"const foo = function (a = 1, { b }) {};",
			[]string{"a = 1"}},
		{"const foo = function ({ a } = {}, b) {}; (2)",
			"const foo = function ({ a } = {}, b) {};",
			[]string{"{ a } = {}"}},
		{"const foo = function ({ a, b } = { a: 1, b: 2 }, c) {}; (2)",
			"const foo = function ({ a, b } = { a: 1, b: 2 }, c) {};",
			[]string{"{ a, b } = { a: 1, b: 2 }"}},
		{"const foo = function ([a] = [], b) {}; (2)",
			"const foo = function ([a] = [], b) {};",
			[]string{"[a] = []"}},
		{"const foo = function ([a, b] = [1, 2], c) {}; (2)",
			"const foo = function ([a, b] = [1, 2], c) {};",
			[]string{"[a, b] = [1, 2]"}},
		{"const foo = (a = 1, b: number) => {}; (2)",
			"const foo = (a = 1, b: number) => {};",
			[]string{"a = 1"}},
		{"const foo = (a = 1, b = 2, c: number) => {}; (2)",
			"const foo = (a = 1, b = 2, c: number) => {};",
			[]string{"a = 1", "b = 2"}},
		{"const foo = (a = 1, b: number, c = 2, d: number) => {}; (2)",
			"const foo = (a = 1, b: number, c = 2, d: number) => {};",
			[]string{"a = 1", "c = 2"}},
		{"const foo = (a = 1, b: number, c = 2) => {}; (2)",
			"const foo = (a = 1, b: number, c = 2) => {};",
			[]string{"a = 1"}},
		{"const foo = (a = 1, b: number, ...c) => {}; (2)",
			"const foo = (a = 1, b: number, ...c) => {};",
			[]string{"a = 1"}},
		{"const foo = (a?: number, b: number) => {}; (2)",
			"const foo = (a?: number, b: number) => {};",
			[]string{"a?: number"}},
		{"const foo = (a: number, b?: number, c: number) => {}; (2)",
			"const foo = (a: number, b?: number, c: number) => {};",
			[]string{"b?: number"}},
		{"const foo = (a = 1, b?: number, c: number) => {}; (2)",
			"const foo = (a = 1, b?: number, c: number) => {};",
			[]string{"a = 1", "b?: number"}},
		{"const foo = (a = 1, { b }) => {}; (2)",
			"const foo = (a = 1, { b }) => {};",
			[]string{"a = 1"}},
		{"const foo = ({ a } = {}, b) => {}; (2)",
			"const foo = ({ a } = {}, b) => {};",
			[]string{"{ a } = {}"}},
		{"const foo = ({ a, b } = { a: 1, b: 2 }, c) => {}; (2)",
			"const foo = ({ a, b } = { a: 1, b: 2 }, c) => {};",
			[]string{"{ a, b } = { a: 1, b: 2 }"}},
		{"const foo = ([a] = [], b) => {}; (2)",
			"const foo = ([a] = [], b) => {};",
			[]string{"[a] = []"}},
		{"const foo = ([a, b] = [1, 2], c) => {}; (2)",
			"const foo = ([a, b] = [1, 2], c) => {};",
			[]string{"[a, b] = [1, 2]"}},
		{"class Foo { constructor( public a: number, protected b?: number, private c:... (2)",
			"\nclass Foo {\n  constructor(\n    public a: number,\n    protected b?: number,\n    private c: number,\n  ) {}\n}\n      ",
			[]string{"protected b?: number"}},
		{"class Foo { constructor( public a: number, protected b = 0, private c: numb... (2)",
			"\nclass Foo {\n  constructor(\n    public a: number,\n    protected b = 0,\n    private c: number,\n  ) {}\n}\n      ",
			[]string{"protected b = 0"}},
		{"class Foo { constructor( public a?: number, private b: number, ) {} } (2)",
			"\nclass Foo {\n  constructor(\n    public a?: number,\n    private b: number,\n  ) {}\n}\n      ",
			[]string{"public a?: number"}},
		{"class Foo { constructor( public a = 0, private b: number, ) {} } (2)",
			"\nclass Foo {\n  constructor(\n    public a = 0,\n    private b: number,\n  ) {}\n}\n      ",
			[]string{"public a = 0"}},
		{"class Foo { constructor(a = 0, b: number) {} } (2)",
			"\nclass Foo {\n  constructor(a = 0, b: number) {}\n}\n      ",
			[]string{"a = 0"}},
		{"class Foo { constructor(a?: number, b: number) {} } (2)",
			"\nclass Foo {\n  constructor(a?: number, b: number) {}\n}\n      ",
			[]string{"a?: number"}},
		// The parameter-property spans, called out separately because they are the reason this test
		// exists. Measured against the installed rule: the modifier is inside the span.
		{"a public parameter property spans its modifier",
			"class Foo { constructor(public a = 0, private b: number) {} }",
			[]string{"public a = 0"}},
		{"a readonly parameter property spans its modifier",
			"class Foo { constructor(readonly a = 0, b: number) {} }",
			[]string{"readonly a = 0"}},
		{"an optional parameter property spans its modifier",
			"class Foo { constructor(private a?: number, b: number) {} }",
			[]string{"private a?: number"}},
		// A type annotation is inside the span.
		{"a defaulted parameter with an annotation spans the annotation",
			"function foo(a: number = 1, b: number) {}",
			[]string{"a: number = 1"}},
		// Findings come out in source order, which the four multi-finding corpus cases above
		// already pin. This one pins it across a wider gap.
		{"three defaults straddling two required parameters, in source order",
			"function foo(a = 1, b: number, c = 2, d: number, e = 3) {}",
			[]string{"a = 1", "c = 2"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, DefaultParamLast, defaultParamLastFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantTexts))
			}
			for index, wantText := range testCase.wantTexts {
				diagnostic := result.Diagnostics[index]
				gotText := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotText != wantText {
					t.Errorf("finding %d spans %q, want %q", index, gotText, wantText)
				}
			}
		})
	}
}

// TestDefaultParamLastMessageReadsAsWritten asserts the message against literal strings typed here
// rather than against the rule's own constants, so a mutation moving a constant moves only one side
// of the comparison.
//
// The rule interpolates nothing, so there is no rendered text and no format string to guard; a
// rule.Message is {Id, Description} and both halves are asserted directly.
func TestDefaultParamLastMessageReadsAsWritten(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, DefaultParamLast, defaultParamLastFile,
		"function foo(a = 1, b: number) {}")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	message := result.Diagnostics[0].Message
	if message.Id != "shouldBeLast" {
		t.Errorf("message id is %q, want %q", message.Id, "shouldBeLast")
	}
	// The description says why the code is wrong rather than restating the rule name. Asserted on
	// the opening clause and on the closing one, so a mutation to either end is visible.
	if !strings.HasPrefix(message.Description, "A parameter with a default sits before a parameter") {
		t.Errorf("description opens with %q", message.Description)
	}
	if !strings.HasSuffix(message.Description, "where omitting it is what selects the default.") {
		t.Errorf("description closes with %q", message.Description)
	}
}
