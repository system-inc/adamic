package core

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// uselessConstructorFile is where the fixtures pretend to live.
const uselessConstructorFile = "/repository/source/UselessConstructor.ts"

// The corpus is ESLint's own, at `tests/lib/rules/no-useless-constructor.js`, extracted by
// loading that file with a stub rule tester rather than retyped: 36 clean cases and 22
// reporting ones, which is the whole upstream file.
//
// Upstream splits them across two rule testers, one per parser. Both are imported as one set,
// because our parser reads TypeScript natively and the split is a fact about espree.
func TestNoUselessConstructorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"class A { }",
		"class A { constructor(){ doSomething(); } }",
		"class A extends B { constructor(){} }",
		"class A extends B { constructor(){ super('foo'); } }",
		"class A extends B { constructor(foo, bar){ super(foo, bar, 1); } }",
		"class A extends B { constructor(){ super(); doSomething(); } }",
		"class A extends B { constructor(...args){ super(...args); doSomething(); } }",
		"class A { dummyMethod(){ doSomething(); } }",
		"class A extends B.C { constructor() { super(foo); } }",
		"class A extends B.C { constructor([a, b, c]) { super(...arguments); } }",
		"class A extends B.C { constructor(a = f()) { super(...arguments); } }",
		"class A extends B { constructor(a, b, c) { super(a, b); } }",
		"class A extends B { constructor(foo, bar){ super(foo); } }",
		"class A extends B { constructor(test) { super(); } }",
		"class A extends B { constructor() { foo; } }",
		"class A extends B { constructor(foo, bar) { super(bar); } }",
		"declare class A { constructor(options: any); }",
		"\n      declare class A {\n        constructor();\n      }\n          ",
		"\n      class A {\n        constructor();\n      }\n          ",
		"\n      abstract class A {\n        constructor();\n      }\n          ",
		"\n      class A {\n        constructor(private name: string) {}\n      }\n          ",
		"\n      class A {\n        constructor(public name: string) {}\n      }\n          ",
		"\n      class A {\n        constructor(protected name: string) {}\n      }\n          ",
		"\n      class A {\n        private constructor() {}\n      }\n          ",
		"\n      class A {\n        protected constructor() {}\n      }\n          ",
		"\n      class A extends B {\n        public constructor() {}\n      }\n          ",
		"\n      class A extends B {\n        public constructor() {\n            super();\n        }\n      }\n          ",
		"\n      class A extends B {\n        protected constructor(foo, bar) {\n          super(bar);\n        }\n      }\n          ",
		"\n      class A extends B {\n        private constructor(foo, bar) {\n          super(bar);\n        }\n      }\n          ",
		"\n      class A extends B {\n        public constructor(foo) {\n          super(foo);\n        }\n      }\n          ",
		"\n      class A extends B {\n        public constructor(foo) {}\n      }\n          ",
		"\n      class A {\n        constructor(foo);\n      }\n          ",
		"\n      class A {\n        constructor(@Foo foo) {}\n      }\n          ",
		"\n      class A {\n        constructor(@Foo foo: string) {}\n      }\n          ",
		"\n      class A extends Object {\n        constructor(@Foo foo: string) {\n          super(foo);\n        }\n      }\n          ",
		"\n      class A extends Object {\n        constructor(foo: string, @Bar() bar) {\n          super(foo, bar);\n        }\n      }\n          ",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUselessConstructor,
				uselessConstructorFile, sourceText))
		})
	}
}

// Every reporting case, with the suggestion output upstream asserts for each.
//
// The suggestion removes the constructor entirely, and whether it must leave a semicolon behind
// is a real decision rather than formatting: a class body member that would otherwise continue
// the previous expression needs one. Upstream computes that and its outputs pin it.
func TestNoUselessConstructorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantOutput string
	}{
		{"class A { constructor(){} }", "class A {  }"},
		{"class A { constructor     (){} }", "class A {  }"},
		{"class A { 'constructor'(){} }", "class A {  }"},
		{"class A extends B { constructor() { super(); } }", "class A extends B {  }"},
		{"class A extends B { constructor(foo){ super(foo); } }", "class A extends B {  }"},
		{"class A extends B { constructor(foo, bar){ super(foo, bar); } }", "class A extends B {  }"},
		{"class A extends B { constructor(...args){ super(...args); } }", "class A extends B {  }"},
		{"class A extends B.C { constructor() { super(...arguments); } }", "class A extends B.C {  }"},
		{"class A extends B { constructor(a, b, ...c) { super(...arguments); } }", "class A extends B {  }"},
		{"class A extends B { constructor(a, b, ...c) { super(a, b, ...c); } }", "class A extends B {  }"},
		{"class A {\n  foo = 'bar'\n  constructor() { }\n  [0]() { }\n}", "class A {\n  foo = 'bar'\n  ;\n  [0]() { }\n}"},
		{"class A {\n  foo = 'bar'\n  constructor() { }\n  *baz() {}\n}", "class A {\n  foo = 'bar'\n  ;\n  *baz() {}\n}"},
		{"class A {\n  foo = 'bar'\n  constructor() { }\n  in\n}", "class A {\n  foo = 'bar'\n  ;\n  in\n}"},
		{"class A {\n  foo = 'bar'\n  constructor() { }\n  instanceof\n}", "class A {\n  foo = 'bar'\n  ;\n  instanceof\n}"},
		{"class A {\n  foo = 'bar'\n  constructor() { }\n  #instanceof\n}", "class A {\n  foo = 'bar'\n  \n  #instanceof\n}"},
		{"class A {\n  foo\n  constructor() { }\n  [0]() { }\n}", "class A {\n  foo\n  \n  [0]() { }\n}"},
		{"class A {\n  \"foo\"\n  constructor() { }\n  [0]() { }\n}", "class A {\n  \"foo\"\n  \n  [0]() { }\n}"},
		{"class A {\n  42\n  constructor() { }\n  [0]() { }\n}", "class A {\n  42\n  \n  [0]() { }\n}"},
		{"class A {\n  [foo]\n  constructor() { }\n  [0]() { }\n}", "class A {\n  [foo]\n  \n  [0]() { }\n}"},
		{"class A {\n  #foo\n  constructor() { }\n  [0]() { }\n}", "class A {\n  #foo\n  \n  [0]() { }\n}"},
		{"\n            class A {\n                constructor() {}\n            }\n              ", "\n            class A {\n                \n            }\n              "},
		{"\n            class A {\n                public constructor() {}\n            }\n        ", "\n            class A {\n                \n            }\n        "},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConstructor, uselessConstructorFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noUselessConstructor")

			suggestions := result.Diagnostics[0].Suggestions
			if len(suggestions) != 1 {
				t.Fatalf("offered %d suggestions, want 1", len(suggestions))
			}
			if suggestions[0].Message.Id != "removeConstructor" {
				t.Errorf("suggestion is %q, want %q", suggestions[0].Message.Id, "removeConstructor")
			}
			if got := applyUselessConstructorSuggestion(t, result.SourceFile.Text(), suggestions[0]); got != testCase.wantOutput {
				t.Errorf("applying the suggestion produced:\n  %q\nwant:\n  %q", got, testCase.wantOutput)
			}
		})
	}
}

// applyUselessConstructorSuggestion replays a suggestion's fixes, back to front.
//
// `rule_testing` applies a fix but has no suggestion support at all, so this lives here. Same
// ordering the fix engine uses, for the same reason.
func applyUselessConstructorSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	if len(suggestion.Fixes) == 0 {
		t.Fatalf("suggestion %q carries no fixes", suggestion.Message.Id)
	}
	ordered := make([]rule.Fix, len(suggestion.Fixes))
	copy(ordered, suggestion.Fixes)
	sort.Slice(ordered, func(first, second int) bool {
		return ordered[first].Range.Pos() > ordered[second].Range.Pos()
	})
	for _, fix := range ordered {
		if fix.Range.Pos() < 0 || fix.Range.End() > len(source) || fix.Range.Pos() > fix.Range.End() {
			t.Fatalf("fix range [%d,%d) is outside the source, which is a defect in the rule",
				fix.Range.Pos(), fix.Range.End())
		}
		source = source[:fix.Range.Pos()] + fix.Text + source[fix.Range.End():]
	}
	return source
}

// The span upstream reports, measured against the installed eslint at 10.8.1.
//
// It runs from the member start through the token before the parameter list, so it covers any
// accessibility modifier: `public constructor` rather than `constructor`. Upstream builds it
// explicitly from `node.loc.start` and the token before the opening paren, so a port reporting
// the whole member would include the body and a port reporting the name alone would drop the
// modifier. Both would pass every message-id fixture.
func TestNoUselessConstructorSpansTheHead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantReported string
	}{
		{"class A { constructor() {} }", "constructor"},
		{"class A extends B { constructor() { super(); } }", "constructor"},
		{"class A { public constructor() {} }", "public constructor"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConstructor, uselessConstructorFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			source := result.SourceFile.Text()
			if got := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]; got != testCase.wantReported {
				t.Errorf("the finding points at %q, want %q", got, testCase.wantReported)
			}
		})
	}
}

// The accessibility rule, which is the subtlest thing this rule decides and which upstream's
// corpus covers only partly.
//
// `private` and `protected` always make a constructor useful, because they restrict who may
// construct. `public` is useful ONLY on a subclass, where it widens the visibility the parent
// declared; on a base class it says what was already true. All four measured against the
// installed build.
func TestNoUselessConstructorAccessibility(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		wantFinding bool
	}{
		{"class A { private constructor() {} }", false},
		{"class A { protected constructor() {} }", false},
		{"class A { public constructor() {} }", true},
		{"class A extends B { public constructor() { super(); } }", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConstructor, uselessConstructorFile, testCase.sourceText)
			if testCase.wantFinding {
				rule_testing.ExpectFindings(t, result, "noUselessConstructor")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The message text, asserted against a literal typed here rather than against the rule's own
// constant, because comparing a diagnostic to the constant it was built from moves both sides
// together under mutation.
func TestNoUselessConstructorMessages(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUselessConstructor, uselessConstructorFile,
		"class A { constructor() {} }")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "noUselessConstructor" {
		t.Errorf("message id is %q, want %q", got, "noUselessConstructor")
	}
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, "This constructor does nothing") {
		t.Errorf("message description starts %q, which is not the sentence this rule reports", got)
	}
}

// Forwarding shapes that look like a pass-through and are not.
//
// Two mutants survived the whole imported corpus here. Dropping the test that a rest parameter must
// be met by a SPREAD left `constructor(...args) { super(args); }` reading as useless, and it is not:
// `super(args)` passes the array itself where `super(...args)` passes its elements, so the two calls
// do different things. Dropping the comparison of argument NAMES left `super(y)` forwarding `x`.
//
// Upstream's corpus writes neither, because both are mistakes nobody makes on purpose. All five rows
// measured against the installed eslint at 10.8.1.
func TestNoUselessConstructorForwardingMustMatchExactly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		wantFinding bool
	}{
		// A rest met by a bare identifier passes the array, not its elements.
		{"class A extends B { constructor(...args) { super(args); } }", false},
		// The same shape with the spread IS a pass-through, and is the control: without it the four
		// clean rows could pass for a rule that never reports a rest at all.
		{"class A extends B { constructor(...args) { super(...args); } }", true},
		// A different name forwarded.
		{"class A extends B { constructor(x) { super(y); } }", false},
		// The right names in the wrong order.
		{"class A extends B { constructor(x, y) { super(y, x); } }", false},
		// A rest spread under a different name.
		{"class A extends B { constructor(...a) { super(...b); } }", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConstructor, uselessConstructorFile, testCase.sourceText)
			if testCase.wantFinding {
				rule_testing.ExpectFindings(t, result, "noUselessConstructor")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// A class that only implements an interface, and a body whose one statement is not a super call.
//
// Two more mutants survived the imported corpus. Treating any heritage clause as a superclass made
// `class A implements I { constructor() {} }` judge itself by the subclass rule and go clean; it is
// a base class, and upstream reports it. And dropping the requirement that the sole statement's
// callee be `super` let any single call read as a pass-through.
//
// Upstream's corpus writes no `implements` at all, because its default parser cannot parse one.
// All five rows measured against the installed eslint at 10.8.1 with the TypeScript parser.
func TestNoUselessConstructorHeritageAndSoleStatement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		wantFinding bool
	}{
		// Implements is not extends, so this is judged as a base class: an empty body is useless.
		{"class A implements I { constructor() {} }", true},
		// And by the same token a `super()` here is NOT a pass-through, it is a statement in a base
		// class constructor, which makes the constructor do something.
		{"class A implements I { constructor() { super(); } }", false},
		// Both clauses together: the extends is what counts, so this is the subclass rule again.
		{"class A extends B implements I { constructor() { super(); } }", true},
		// A sole statement that is a call but not to super.
		{"class A extends B { constructor() { doSomething(); } }", false},
		// A sole statement that is not a call at all.
		{"class A extends B { constructor() { this.x = 1; } }", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConstructor, uselessConstructorFile, testCase.sourceText)
			if testCase.wantFinding {
				rule_testing.ExpectFindings(t, result, "noUselessConstructor")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}
