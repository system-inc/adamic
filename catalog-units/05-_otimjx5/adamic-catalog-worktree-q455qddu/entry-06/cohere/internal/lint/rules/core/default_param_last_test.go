package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const defaultParamLastFile = "/repository/source/Parameters.ts"

// TestDefaultParamLastFires carries all 41 failing cases from the eslint corpus verbatim, from both
// of its RuleTester blocks, plus cases pinning decisions the corpus never exercises. The corpus
// states a messageId per finding, so the counts here are read off rather than recovered: 41 inputs
// producing 52 findings.
//
// Every case was replayed through the installed eslint build before it was written here, and all 96
// corpus cases agreed with the corpus on count and on span. The expectations are the SPAN TEXT
// rather than a message id, because a message id cannot see where a finding points and this rule
// carries no repair to check instead. Each span was sliced out of the source by the same script that
// extracted the case, from the columns the installed rule reported.
func TestDefaultParamLastFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		// The eslint corpus, verbatim. Both of its RuleTester blocks: the JavaScript one
		// and the TypeScript one, in source order.
		{"function f(a = 5, b) {}",
			"function f(a = 5, b) {}",
			[]string{"a = 5"}},
		{"function f(a = 5, b = 6, c) {}",
			"function f(a = 5, b = 6, c) {}",
			[]string{"a = 5", "b = 6"}},
		{"function f (a = 5, b, c = 6, d) {}",
			"function f (a = 5, b, c = 6, d) {}",
			[]string{"a = 5", "c = 6"}},
		{"function f(a = 5, b, c = 5) {}",
			"function f(a = 5, b, c = 5) {}",
			[]string{"a = 5"}},
		{"const f = (a = 5, b, ...c) => {}",
			"const f = (a = 5, b, ...c) => {}",
			[]string{"a = 5"}},
		{"const f = function f (a, b = 5, c) {}",
			"const f = function f (a, b = 5, c) {}",
			[]string{"b = 5"}},
		{"const f = (a = 5, { b }) => {}",
			"const f = (a = 5, { b }) => {}",
			[]string{"a = 5"}},
		{"const f = ({ a } = {}, b) => {}",
			"const f = ({ a } = {}, b) => {}",
			[]string{"{ a } = {}"}},
		{"const f = ({ a, b } = { a: 1, b: 2 }, c) => {}",
			"const f = ({ a, b } = { a: 1, b: 2 }, c) => {}",
			[]string{"{ a, b } = { a: 1, b: 2 }"}},
		{"const f = ([a] = [], b) => {}",
			"const f = ([a] = [], b) => {}",
			[]string{"[a] = []"}},
		{"const f = ([a, b] = [1, 2], c) => {}",
			"const f = ([a, b] = [1, 2], c) => {}",
			[]string{"[a, b] = [1, 2]"}},
		{"function foo(a = 1, b: number) {}",
			"function foo(a = 1, b: number) {}",
			[]string{"a = 1"}},
		{"function foo(a = 1, b = 2, c: number) {}",
			"function foo(a = 1, b = 2, c: number) {}",
			[]string{"a = 1", "b = 2"}},
		{"function foo(a = 1, b: number, c = 2, d: number) {}",
			"function foo(a = 1, b: number, c = 2, d: number) {}",
			[]string{"a = 1", "c = 2"}},
		{"function foo(a = 1, b: number, c = 2) {}",
			"function foo(a = 1, b: number, c = 2) {}",
			[]string{"a = 1"}},
		{"function foo(a = 1, b: number, ...c) {}",
			"function foo(a = 1, b: number, ...c) {}",
			[]string{"a = 1"}},
		{"function foo(a?: number, b: number) {}",
			"function foo(a?: number, b: number) {}",
			[]string{"a?: number"}},
		{"function foo(a: number, b?: number, c: number) {}",
			"function foo(a: number, b?: number, c: number) {}",
			[]string{"b?: number"}},
		{"function foo(a = 1, b?: number, c: number) {}",
			"function foo(a = 1, b?: number, c: number) {}",
			[]string{"a = 1", "b?: number"}},
		{"const foo = function (a = 1, b: number) {};",
			"const foo = function (a = 1, b: number) {};",
			[]string{"a = 1"}},
		{"const foo = function (a = 1, b = 2, c: number) {};",
			"const foo = function (a = 1, b = 2, c: number) {};",
			[]string{"a = 1", "b = 2"}},
		{"const foo = function (a = 1, b: number, c = 2, d: number) {};",
			"const foo = function (a = 1, b: number, c = 2, d: number) {};",
			[]string{"a = 1", "c = 2"}},
		{"const foo = function (a = 1, b: number, c = 2) {};",
			"const foo = function (a = 1, b: number, c = 2) {};",
			[]string{"a = 1"}},
		{"const foo = function (a = 1, b: number, ...c) {};",
			"const foo = function (a = 1, b: number, ...c) {};",
			[]string{"a = 1"}},
		{"const foo = function (a?: number, b: number) {};",
			"const foo = function (a?: number, b: number) {};",
			[]string{"a?: number"}},
		{"const foo = function (a: number, b?: number, c: number) {};",
			"const foo = function (a: number, b?: number, c: number) {};",
			[]string{"b?: number"}},
		{"const foo = function (a = 1, b?: number, c: number) {};",
			"const foo = function (a = 1, b?: number, c: number) {};",
			[]string{"a = 1", "b?: number"}},
		{"const foo = (a = 1, b: number) => {};",
			"const foo = (a = 1, b: number) => {};",
			[]string{"a = 1"}},
		{"const foo = (a = 1, b = 2, c: number) => {};",
			"const foo = (a = 1, b = 2, c: number) => {};",
			[]string{"a = 1", "b = 2"}},
		{"const foo = (a = 1, b: number, c = 2, d: number) => {};",
			"const foo = (a = 1, b: number, c = 2, d: number) => {};",
			[]string{"a = 1", "c = 2"}},
		{"const foo = (a = 1, b: number, c = 2) => {};",
			"const foo = (a = 1, b: number, c = 2) => {};",
			[]string{"a = 1"}},
		{"const foo = (a = 1, b: number, ...c) => {};",
			"const foo = (a = 1, b: number, ...c) => {};",
			[]string{"a = 1"}},
		{"const foo = (a?: number, b: number) => {};",
			"const foo = (a?: number, b: number) => {};",
			[]string{"a?: number"}},
		{"const foo = (a: number, b?: number, c: number) => {};",
			"const foo = (a: number, b?: number, c: number) => {};",
			[]string{"b?: number"}},
		{"const foo = (a = 1, b?: number, c: number) => {};",
			"const foo = (a = 1, b?: number, c: number) => {};",
			[]string{"a = 1", "b?: number"}},
		{"class Foo { constructor( public a: number, protected b?: number, private c...",
			"\n    class Foo {\n      constructor(\n        public a: number,\n        protected b?: number,\n        private c: number,\n      ) {}\n    }\n          ",
			[]string{"protected b?: number"}},
		{"class Foo { constructor( public a: number, protected b = 0, private c: num...",
			"\n    class Foo {\n      constructor(\n        public a: number,\n        protected b = 0,\n        private c: number,\n      ) {}\n    }\n          ",
			[]string{"protected b = 0"}},
		{"class Foo { constructor( public a?: number, private b: number, ) {} }",
			"\n    class Foo {\n      constructor(\n        public a?: number,\n        private b: number,\n      ) {}\n    }\n          ",
			[]string{"public a?: number"}},
		{"class Foo { constructor( public a = 0, private b: number, ) {} }",
			"\n    class Foo {\n      constructor(\n        public a = 0,\n        private b: number,\n      ) {}\n    }\n          ",
			[]string{"public a = 0"}},
		{"class Foo { constructor(a = 0, b: number) {} }",
			"\n    class Foo {\n      constructor(a = 0, b: number) {}\n    }\n          ",
			[]string{"a = 0"}},
		{"class Foo { constructor(a?: number, b: number) {} }",
			"\n    class Foo {\n      constructor(a?: number, b: number) {}\n    }\n          ",
			[]string{"a?: number"}},
		// Cases the corpus does not write, pinning decisions measured against the installed build.
		//
		// A rest parameter before a required one is the ONE class where core and the shipped
		// @typescript-eslint sibling disagree. Core reports it; the sibling is silent. The source is
		// illegal and reaches the rule only through parser error recovery.
		{"a rest parameter before a required one reports",
			"function f(...a: number[], b: number) {}",
			[]string{"...a: number[]"}},
		{"a rest parameter after a required one and before another reports",
			"function f(a: number, ...b: number[], c: number) {}",
			[]string{"...b: number[]"}},
		{"a default and a rest both report when a required parameter follows",
			"function f(a = 1, ...b: number[], c: number) {}",
			[]string{"a = 1", "...b: number[]"}},
		// The accessor arms. A setter takes one parameter and a getter none in the GRAMMAR, but the
		// parser recovers from the illegal source and hands back the parameters anyway, and the
		// installed build reports all four shapes.
		{"a class setter with a recovered second parameter reports",
			"class C { set x(a = 1, b: number) {} }",
			[]string{"a = 1"}},
		{"a class getter with recovered parameters reports",
			"class C { get x(a = 1, b: number) { return 1; } }",
			[]string{"a = 1"}},
		{"an object literal setter with a recovered second parameter reports",
			"const o = { set x(a = 1, b: number) {} };",
			[]string{"a = 1"}},
		{"an object literal getter with recovered parameters reports",
			"const o = { get x(a = 1, b: number) { return 1; } };",
			[]string{"a = 1"}},
		// A method and a constructor are a FunctionExpression upstream, so upstream's listener
		// reaches them. Ours needs an arm apiece.
		{"a class method reports",
			"class C { m(a = 1, b: number) {} }",
			[]string{"a = 1"}},
		{"an object literal method reports",
			"const o = { m(a = 1, b: number) {} };",
			[]string{"a = 1"}},
		{"a constructor implementation beside an overload reports on the implementation",
			"class C { constructor(a = 1, b: number); constructor(a = 1, b: number) {} }",
			[]string{"a = 1"}},
		// The parameter-property spans. typescript-go puts the modifier inside the parameter's own
		// token range, which is what makes the span match upstream's outer wrapper node.
		{"a public parameter property spans its modifier",
			"class C { constructor(public a = 0, private b: number) {} }",
			[]string{"public a = 0"}},
		{"a readonly parameter property spans its modifier",
			"class C { constructor(readonly a = 0, b: number) {} }",
			[]string{"readonly a = 0"}},
		{"a private optional parameter property spans its modifier",
			"class C { constructor(private a?: number, b: string) {} }",
			[]string{"private a?: number"}},
		// A type annotation is inside the span.
		{"a defaulted parameter with an annotation spans the annotation",
			"function f(a: number = 1, b: number) {}",
			[]string{"a: number = 1"}},
		// Findings come out in source order, which the corpus's multi-finding cases already pin.
		// This one pins it across a wider gap, with a trailing default that must NOT report.
		{"three defaults straddling two required parameters, in source order",
			"function f(a = 1, b: number, c = 2, d: number, e = 3) {}",
			[]string{"a = 1", "c = 2"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, DefaultParamLast, defaultParamLastFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantTexts))
			}
			messageIds := make([]string, len(result.Diagnostics))
			for index := range result.Diagnostics {
				messageIds[index] = "shouldBeLast"
			}
			rule_testing.ExpectFindings(t, result, messageIds...)
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

// TestDefaultParamLastStaysSilent carries all 55 clean cases from the eslint corpus verbatim, plus
// the bodyless shapes the corpus never writes. Every clean case upstream ships is a false positive
// somebody already hit.
func TestDefaultParamLastStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The eslint corpus's clean cases, verbatim.
		{"function f() {}",
			"function f() {}"},
		{"function f(a) {}",
			"function f(a) {}"},
		{"function f(a = 5) {}",
			"function f(a = 5) {}"},
		{"function f(a, b) {}",
			"function f(a, b) {}"},
		{"function f(a, b = 5) {}",
			"function f(a, b = 5) {}"},
		{"function f(a, b = 5, c = 5) {}",
			"function f(a, b = 5, c = 5) {}"},
		{"function f(a, b = 5, ...c) {}",
			"function f(a, b = 5, ...c) {}"},
		{"const f = () => {}",
			"const f = () => {}"},
		{"const f = (a) => {}",
			"const f = (a) => {}"},
		{"const f = (a = 5) => {}",
			"const f = (a = 5) => {}"},
		{"const f = function f() {}",
			"const f = function f() {}"},
		{"const f = function f(a) {}",
			"const f = function f(a) {}"},
		{"const f = function f(a = 5) {}",
			"const f = function f(a = 5) {}"},
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
			"\n    class Foo {\n      constructor(a: number, b: number, c: number) {}\n    }\n        "},
		{"class Foo { constructor(a: number, b?: number, c = 1) {} }",
			"\n    class Foo {\n      constructor(a: number, b?: number, c = 1) {}\n    }\n        "},
		{"class Foo { constructor(a: number, b = 1, c?: number) {} }",
			"\n    class Foo {\n      constructor(a: number, b = 1, c?: number) {}\n    }\n        "},
		{"class Foo { constructor( public a: number, protected b: number, private c:...",
			"\n    class Foo {\n      constructor(\n        public a: number,\n        protected b: number,\n        private c: number,\n      ) {}\n    }\n        "},
		{"class Foo { constructor( public a: number, protected b?: number, private c...",
			"\n    class Foo {\n      constructor(\n        public a: number,\n        protected b?: number,\n        private c = 10,\n      ) {}\n    }\n        "},
		{"class Foo { constructor( public a: number, protected b = 10, private c?: n...",
			"\n    class Foo {\n      constructor(\n        public a: number,\n        protected b = 10,\n        private c?: number,\n      ) {}\n    }\n        "},
		{"class Foo { constructor( a: number, protected b?: number, private c = 0, )...",
			"\n    class Foo {\n      constructor(\n        a: number,\n        protected b?: number,\n        private c = 0,\n      ) {}\n    }\n        "},
		{"class Foo { constructor( a: number, b?: number, private c = 0, ) {} }",
			"\n    class Foo {\n      constructor(\n        a: number,\n        b?: number,\n        private c = 0,\n      ) {}\n    }\n        "},
		{"class Foo { constructor( a: number, private b?: number, c = 0, ) {} }",
			"\n    class Foo {\n      constructor(\n        a: number,\n        private b?: number,\n        c = 0,\n      ) {}\n    }\n        "},
		// An all-required parameter list, written here rather than taken from the corpus. Upstream
		// has valid cases of this shape, but a rule whose predicate answered "carries a default"
		// for a plain parameter would light up every function in the tree, and that failure is
		// worth a case somebody wrote deliberately. My first draft of this rule had exactly that
		// defect: it kept upstream's loop bound and dropped its skip over a required parameter, so
		// every parameter in a list like these reported. Inverting the predicate turns all four
		// red.
		{"two plain required parameters",
			"function componentPrefixFromInterfaceName(interfaceName: string, suffix: string): string | null {\n\treturn null;\n}"},
		{"three plain required parameters",
			"function f(a: string, b: number, c: boolean) { return [a, b, c]; }"},
		{"a required parameter followed by a destructured one carrying no default",
			"function f(a: string, { b }: { b: number }) { return [a, b]; }"},
		{"a method whose second parameter is an array binding pattern carrying no default",
			"const o = {\n\tcreate(context, [options]) {\n\t\treturn [context, options];\n\t},\n};"},
		// Cases the corpus does not write. Every one is a shape our parser gives an ordinary
		// function kind while the TypeScript ESTree gives it a node type outside upstream's three
		// listeners, so without the nil-body guard each would report. Measured clean against the
		// installed build.
		{"an ambient function declaration is exempt",
			"declare function f(a = 1, b: number): void;"},
		{"a constructor overload signature is exempt",
			"class C { constructor(a = 1, b: number); constructor(x?: number) {} }"},
		{"an abstract method signature is exempt",
			"abstract class C { abstract m(a?: number, b: number): void; }"},
		{"a method in an ambient class is exempt",
			"declare class C { m(a = 1, b: number): void; }"},
		{"a function overload signature is exempt",
			"function f(a = 1, b: number): void;\nfunction f(x?: number): void {}"},
		// These two our parser gives their own kinds, which are simply not listened for. They are
		// here so that adding an arm for either fails loudly.
		{"an interface method signature is exempt",
			"interface I { m(a?: number, b: number): void }"},
		{"a function type is exempt",
			"type T = (a?: number, b: number) => void;"},
		// A rest parameter in its legal position stays clean, which is what separates the rest
		// cases above from a rule that simply reports every rest parameter.
		{"a rest parameter after a default is clean",
			"function f(a = 1, ...b: number[]) {}"},
		{"a rest parameter after a required parameter is clean",
			"function f(a: number, ...b: number[]) {}"},
		{"a lone rest parameter is clean",
			"function f(...a: number[]) {}"},
		// A destructured parameter's name is a binding pattern, and Node.Text panics on one. These
		// pin that the rule never reads a parameter name; without the corpus's own destructured
		// cases above this would be the only coverage.
		{"a destructured parameter last is clean",
			"function f(a: number, { b } = {}) {}"},
		{"an array-destructured parameter last is clean",
			"function f(a: number, [b] = []) {}"},
		{"a destructured rest parameter is clean",
			"function f(a: number, ...{ length }: number[]) {}"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, DefaultParamLast, defaultParamLastFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
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

	result := rule_testing.Run(t, DefaultParamLast, defaultParamLastFile, "function f(a = 5, b) {}")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	message := result.Diagnostics[0].Message
	if message.Id != "shouldBeLast" {
		t.Errorf("message id is %q, want %q", message.Id, "shouldBeLast")
	}
	if message.Description != "A parameter carrying a default or an optional marker sits before a "+
		"parameter without one, so the default can never be taken: reaching the later parameter "+
		"obliges every caller to pass something here. Move it after the required parameters, "+
		"where omitting it is what selects the default." {
		t.Errorf("description reads %q", message.Description)
	}
}
