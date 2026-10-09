package core

import (
	"fmt"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// thisBeforeSuperFile is where the fixtures pretend to live.
//
// A `.tsx` extension, matching the extension oxc's own snapshot records
// (`no_this_before_super.tsx`), so the fixtures parse the way upstream's did. Nothing in this rule
// reads the extension, but a corpus copied verbatim should be parsed the way it was written.
const thisBeforeSuperFile = "/repository/source/Derived.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_this_before_super.rs`.
// The extractor reports one Tester block, 40 pass and 26 fail, and the snapshot records 26
// diagnostics from those 26 fail inputs. So one finding per failing input is measured here rather
// than assumed, and it is a real property of this rule rather than an accident: upstream reports
// once per offending constructor, not once per offending `this`.
//
// That per-constructor granularity is the sharpest divergence between the two upstreams and it is
// resolved in favour of oxc deliberately. See the rule's doc comment.
func TestNoThisBeforeSuperFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// disallows all `this`/`super` if `super()` is missing.
		{"this assignment with no super call at all", "class A extends B { constructor() { this.c = 0; } }"},
		{"this call with no super call at all", "class A extends B { constructor() { this.c(); } }"},
		{"super property access with no super call at all", "class A extends B { constructor() { super.c(); } }"},
		// disallows `this`/`super` before `super()`.
		{"this assignment before super", "class A extends B { constructor() { this.c = 0; super(); } }"},
		{"this call before super", "class A extends B { constructor() { this.c(); super(); } }"},
		{"super property access before super", "class A extends B { constructor() { super.c(); super(); } }"},
		// disallows `this`/`super` in arguments of `super()`.
		{"this as a direct argument to super", "class A extends B { constructor() { super(this); } }"},
		{"this property as an argument to super", "class A extends B { constructor() { super(this.c); } }"},
		{"this nested two calls deep in a super argument", "class A extends B { constructor() { super(a(b(this.c))); } }"},
		{"a this method call as an argument to super", "class A extends B { constructor() { super(this.c()); } }"},
		{"super property as an argument to super", "class A extends B { constructor() { super(super.c); } }"},
		{"a super method call as an argument to super", "class A extends B { constructor() { super(super.c()); } }"},
		// even if is nested, reports correctly.
		{"outer constructor offends while the inner one is clean", "class A extends B { constructor() { class C extends D { constructor() { super(); this.e(); } } this.f(); super(); } }"},
		{"inner constructor offends while the outer one is clean", "class A extends B { constructor() { class C extends D { constructor() { this.e(); super(); } } super(); this.f(); } }"},
		// multi code path.
		{"super in only the consequent of an if", "class A extends B { constructor() { if (a) super(); this.a(); } }"},
		{"this in a finally after super in the try", "class A extends B { constructor() { try { super(); } finally { this.a; } } }"},
		{"this after a try whose catch could swallow a throwing super", "class A extends B { constructor() { try { super(); } catch (err) { } this.a; } }"},
		{"logical and assignment short-circuits past super", "class A extends B { constructor() { foo &&= super().a; this.c(); } }"},
		{"logical or assignment short-circuits past super", "class A extends B { constructor() { foo ||= super().a; this.c(); } }"},
		{"nullish assignment short-circuits past super", "class A extends B { constructor() { foo ??= super().a; this.c(); } }"},
		{"super nested inside an if with no else", "class A extends B { constructor() { if (foo) { if (bar) { } super(); } this.a(); }}"},
		{"super in only the alternate of an if", `class A extends B {
            constructor() {
                if (foo) {
                } else {
                    super();
                }
                this.a();
            }
        }`},
		{"this in a finally with no super anywhere", `class A extends B {
            constructor() {
                try {
                    call();
                } finally {
                    this.a();
                }
            }
        }`},
		{"this after a while loop that may run zero times", `class A extends B {
            constructor() {
                while (foo) {
                    super();
                }
                this.a();
            }
        }`},
		{"this before super inside a while body", `class A extends B {
            constructor() {
                while (foo) {
                    this.a();
                    super();
                }
            }
        }`},
		{"this before super inside an if inside a while body", `class A extends B {
            constructor() {
                while (foo) {
                    if (init) {
                        this.a();
                        super();
                    }
                }
            }
        }`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, testCase.sourceText),
				"thisBeforeSuper")
		})
	}
}

// The clean cases carry the whole discrimination, and each one fails a different way.
//
// They are worth reading as a group rather than a list, because they encode four distinctions the
// rule has to draw and none of which a naive source-position comparison draws:
//
//   - a class with no `extends`, or `extends null`, has no `super()` to precede
//   - `this` inside a nested function, arrow, or nested class belongs to that scope
//   - a field initializer always evaluates after `super()` regardless of where it is written
//   - `super()` reached on *every* path counts, even when it is written twice
func TestNoThisBeforeSuperStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// if the class has no extends or `extends null`, just ignore.
		// those classes cannot call `super()`.
		{"a class with no body", "class A { }"},
		{"a base class with an empty constructor", "class A { constructor() { } }"},
		{"a base class assigning this", "class A { constructor() { this.b = 0; } }"},
		{"a base class calling a this method", "class A { constructor() { this.b(); } }"},
		{"extends null with no constructor", "class A extends null { }"},
		{"extends null with an empty constructor", "class A extends null { constructor() { } }"},
		// allows `this`/`super` after `super()`.
		{"a derived class with no constructor", "class A extends B { }"},
		{"a constructor that only calls super", "class A extends B { constructor() { super(); } }"},
		{"a class expression thrown from a catch, with this in the outer finally", `
        function f() {
            try {
                return a();
            }
            catch (err) {
                throw new class CustomError extends Error {
                    constructor() {
                        super(err);
                    }
                };
            }
            finally {
                this.b();
            }
        }
        `},
		{"this on both sides of an assignment after super", "class A extends B { constructor() { super(); this.c = this.d; } }"},
		{"a this call after super", "class A extends B { constructor() { super(); this.c(); } }"},
		{"a super property call after super", "class A extends B { constructor() { super(); super.c(); } }"},
		{"super in both arms of an if", "class A extends B { constructor() { if (true) { super(); } else { super(); } this.c(); } }"},
		{"super as the right side of an assignment", "class A extends B { constructor() { foo = super(); this.c(); } }"},
		{"super inside an addition assignment", "class A extends B { constructor() { foo += super().a; this.c(); } }"},
		{"super inside a bitwise or assignment", "class A extends B { constructor() { foo |= super().a; this.c(); } }"},
		{"super inside a bitwise and assignment", "class A extends B { constructor() { foo &= super().a; this.c(); } }"},
		// allows `this`/`super` in nested executable scopes, even if before `super()`.
		{"a nested class declaration before super", "class A extends B { constructor() { class B extends C { constructor() { super(); this.d = 0; } } super(); } }"},
		{"a nested class expression before super", "class A extends B { constructor() { var B = class extends C { constructor() { super(); this.d = 0; } }; super(); } }"},
		{"a function declaration using this before super", "class A extends B { constructor() { function c() { this.d(); } super(); } }"},
		{"a function expression using this before super", "class A extends B { constructor() { var c = function c() { this.d(); }; super(); } }"},
		{"an arrow using this before super", "class A extends B { constructor() { var c = () => this.d(); super(); } }"},
		// ignores out of constructors.
		{"this in a base class method", "class A { b() { this.c = 0; } }"},
		{"this in a derived class method", "class A extends B { c() { this.d = 0; } }"},
		{"this in a plain function", "function a() { this.b = 0; }"},
		// multi code path.
		{"super and this in both arms of an if", "class A extends B { constructor() { if (a) { super(); this.a(); } else { super(); this.b(); } } }"},
		{"super in both unbraced arms of an if", "class A extends B { constructor() { if (a) super(); else super(); this.a(); } }"},
		{"super in a try with an empty finally", "class A extends B { constructor() { try { super(); } finally {} this.a(); } }"},
		// https://github.com/eslint/eslint/issues/5261
		{"a for-of loop using this after super", "class A extends B { constructor(a) { super(); for (const b of a) { this.a(); } } }"},
		{"a for-of loop with no this before super", "class A extends B { constructor(a) { for (const b of a) { foo(b); } super(); } }"},
		// https://github.com/eslint/eslint/issues/5319
		{"this inside a logical chain after super", "class A extends B { constructor(a) { super(); this.a = a && function(){} && this.foo; } }"},
		// https://github.com/eslint/eslint/issues/5394
		{"a bare this after a for loop after super", `class A extends Object {
                constructor() {
                    super();
                    for (let i = 0; i < 0; i++);
                    this;
                }
            }`},
		// https://github.com/eslint/eslint/issues/5894
		{"this after a return in a base class", "class A { constructor() { return; this; } }"},
		{"this after a return in a derived class", "class A extends B { constructor() { return; this; } }"},
		// https://github.com/eslint/eslint/issues/8848
		{"a try containing a for-of after super", `
            class A extends B {
                constructor(props) {
                    super(props);

                    try {
                        let arr = [];
                        for (let a of arr) {
                        }
                    } catch (err) {
                    }
                }
            }
        `},
		// Class field initializers are always evaluated after `super()`.
		{"a field initializer in a base class", "class C { field = this.toString(); }"},
		{"a field initializer in a derived class with no constructor", "class C extends B { field = this.foo(); }"},
		{"a field initializer alongside a clean constructor", "class C extends B { field = this.foo(); constructor() { super(); } }"},
		// < in this case, initializers are never evaluated.
		{"a field initializer alongside a constructor that never calls super", "class C extends B { field = this.foo(); constructor() { } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, testCase.sourceText))
		})
	}
}

// The finding has to point at the offending constructor, and nothing above asserts that.
//
// `ExpectFindings` checks message ids and count only, so a rule reporting the right number of
// findings at entirely the wrong place passes every case in this file. That is not hypothetical:
// this rule anchors on a node it does not receive as a listener argument in the naive version, and
// the two most available wrong answers, the class and the constructor body, both produce exactly
// one finding per offending constructor.
//
// The expected text is taken from oxc's snapshot, which renders the caret under the whole method
// definition starting at `constructor`.
func TestNoThisBeforeSuperReportsAtTheConstructor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{
			"a one-line constructor",
			"class A extends B { constructor() { this.c = 0; } }",
			"constructor() { this.c = 0; }",
		},
		{
			"the caret excludes the class and the trailing brace",
			"class A extends B { constructor() { super.c(); super(); } }",
			"constructor() { super.c(); super(); }",
		},
		{
			"a constructor taking parameters",
			"class A extends B { constructor(a, b) { this.c = a; super(); } }",
			"constructor(a, b) { this.c = a; super(); }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Errorf("finding points at %q, want %q", reported, testCase.want)
			}
		})
	}
}

// A comment before the constructor must not drag the caret onto the comment.
//
// `rule.TokenRange` is what prevents this, by starting the range at the first token rather than at
// the node's `Pos()`, which includes leading trivia. Asserted here rather than trusted, because a
// port building the range from `node.Pos()` directly passes every other test in this file and
// breaks `-next-line` suppressions written above the constructor.
func TestNoThisBeforeSuperSkipsLeadingTrivia(t *testing.T) {
	t.Parallel()

	source := "class A extends B {\n  // build it\n  constructor() { this.c = 0; }\n}"
	result := rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
	}
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "constructor() { this.c = 0; }" {
		t.Errorf("finding points at %q, want the constructor with no leading comment", reported)
	}
}

// The nested case has to pick the constructor that actually offends, not merely some constructor.
//
// Upstream's snapshot proves it does: two fail cases differ only in which of the two constructors
// is clean, and the caret moves from column 21 to column 57 between them. A rule reporting the
// outermost constructor unconditionally produces the right count on both and the right span on
// exactly one, so counting cannot see this and only the span can.
func TestNoThisBeforeSuperPicksTheOffendingNestedConstructor(t *testing.T) {
	t.Parallel()

	outerOffends := "class A extends B { constructor() { class C extends D { constructor() { super(); this.e(); } } this.f(); super(); } }"
	result := rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, outerOffends)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("outer case: expected one finding, got %d", len(result.Diagnostics))
	}
	reported := outerOffends[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if want := "constructor() { class C extends D { constructor() { super(); this.e(); } } this.f(); super(); }"; reported != want {
		t.Errorf("outer case points at %q, want %q", reported, want)
	}

	innerOffends := "class A extends B { constructor() { class C extends D { constructor() { this.e(); super(); } } super(); this.f(); } }"
	result = rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, innerOffends)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("inner case: expected one finding, got %d", len(result.Diagnostics))
	}
	reported = innerOffends[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if want := "constructor() { this.e(); super(); }"; reported != want {
		t.Errorf("inner case points at %q, want %q", reported, want)
	}
}

// The message text is asserted exactly, because the fixtures above cannot see it.
//
// `ExpectFindings` matches the id, so a rule reporting the right id with a description that says
// the wrong thing, or renders a placeholder unsubstituted, passes everything else here.
func TestNoThisBeforeSuperMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile,
		"class A extends B { constructor() { this.c = 0; } }")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
	}
	want := "This derived constructor reads `this` or `super` on a path that has not called " +
		"`super()` yet, which throws a ReferenceError because the `this` binding does not exist " +
		"until `super()` returns."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Errorf("description is %q, want %q", got, want)
	}
	if got := result.Diagnostics[0].Message.Id; got != "thisBeforeSuper" {
		t.Errorf("message id is %q, want %q", got, "thisBeforeSuper")
	}
}

// Cases upstream does not cover, each guarding a decision this port made on its own.
//
// Upstream's corpus exercises `if`, `try`, `while` and the logical assignments, and says nothing
// about the rest of the statement grammar. Every case below names a construct the scanner routes
// deliberately, and each one exists because the routing is a choice rather than a consequence.
func TestNoThisBeforeSuperFiresOnCasesUpstreamOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// A switch with no default can be skipped entirely, so the clause that calls grants
		// nothing. Measured against both upstreams rather than reasoned: this and the four
		// switch cases below were run through `oxlint` and ESLint's Linter API, which agreed.
		{
			"a switch with no default clause grants no credit",
			"class A extends B { constructor(x) { switch (x) { case 1: super(); break; } this.a(); } }",
		},
		// One clause escaping the switch without calling is enough to sink the whole thing, even
		// with a default present. This is the case that makes the analysis a must-analysis over
		// the clause list rather than a search for one calling clause.
		{
			"a switch where one clause escapes without calling",
			"class A extends B { constructor(x) { switch (x) { case 1: super(); break; case 2: bar(); break; default: super(); } this.a(); } }",
		},
		// Falling out of a calling clause into an uncalling default leaves the default's state,
		// not the clause's. Guards the fallthrough carry against being an unconditional win.
		{
			"a calling clause falling through into a default that does not call",
			"class A extends B { constructor(x) { switch (x) { case 1: super(); default: bar(); } this.a(); } }",
		},
		// A violation inside one clause reports even when another clause is clean.
		{
			"this in a switch clause before super in another",
			"class A extends B { constructor(x) { switch (x) { case 1: this.a(); break; default: super(); } } }",
		},
		// An offending constructor inside a class *expression*, which upstream's corpus reaches
		// only in the clean direction. Its `var B = class extends C { constructor() { super();
		// this.d = 0; } }` is a pass case, so a gate that dropped class expressions entirely
		// would keep it green while making the rule silent on every one of them. A mutation
		// narrowing the class-kind check to declarations survived the whole suite until this
		// case existed, which is exactly the blind spot a clean-only corpus leaves behind.
		// Confirmed against `oxlint`, which reports it.
		{
			"an offending constructor in a class expression",
			"var A = class extends B { constructor() { this.c = 0; super(); } };",
		},
		// A labeled statement is transparent: the label changes where `break` goes and not whether
		// the body runs, so a violation inside one has to surface.
		{
			"a labeled block hiding a this before super",
			"class A extends B { constructor() { outer: { this.a(); } super(); } }",
		},
		// A getter inside an object literal in the constructor is a separate evaluation context,
		// so a `super()` written in it must not credit the constructor.
		{
			"super inside a nested accessor does not credit the constructor",
			"class A extends B { constructor() { var o = { get x() { return 1; } }; this.a(); super(); } }",
		},
		// The optional-call spelling of `super?.()` is not a super call the rule credits, because
		// it is not valid in a derived constructor to begin with; what matters is that the walk
		// does not credit something merely because a `super` token appeared next to a call.
		{
			"a super property read through an element access",
			"class A extends B { constructor() { super['c'](); super(); } }",
		},
		// A ternary is not an `if`: neither arm is guaranteed, so credit from either is unsound.
		// The generic child walk handles it, and this pins that the generic path does not credit.
		{
			"super in one arm of a conditional expression grants no credit",
			"class A extends B { constructor() { foo ? super() : bar(); this.a(); } }",
		},
		// A logical `&&` short-circuits exactly like `&&=`, and it reaches a different code path:
		// `&&=` is caught by the operator check while `&&` falls through to the generic walk.
		// Upstream reports this too, for the same reason.
		{
			"super behind a logical and grants no credit",
			"class A extends B { constructor() { foo && super(); this.a(); } }",
		},
		// A parameter default evaluates before the body, but a `this` in it is still before
		// `super()`. This one is a real language rule and upstream's CFG catches it; the scanner
		// here reaches it because the constructor's parameters are children of the constructor.
		{
			"this in a parameter default",
			"class A extends B { constructor(a = this.b) { super(); } }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, testCase.sourceText),
				"thisBeforeSuper")
		})
	}
}

// The other half of the same set: constructs that must stay silent.
//
// These are the ones that catch an over-eager scanner, and three of them guard a specific way this
// port could have gone wrong that upstream's corpus does not reach.
func TestNoThisBeforeSuperStaysSilentOnCasesUpstreamOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// A constructor overload signature has no body at all. Dereferencing it is a panic rather
		// than a wrong answer, which is the failure mode a fixture is most valuable against.
		{
			"an overload signature with no body",
			"class A extends B { constructor(a: string); constructor(a: any) { super(a); } }",
		},
		// `implements` is not `extends`. A class implementing an interface is still a base class
		// with no `super()` to call, and a heritage-clause check that did not distinguish the two
		// would report every `this` in it.
		{
			"a base class implementing an interface",
			"class A implements I { constructor() { this.c = 0; } }",
		},
		// A class expression assigned in a base constructor: the outer constructor is not derived,
		// and the inner one is clean. Guards against the class-kind check answering on the wrong
		// class when the two are nested.
		{
			"a derived class expression inside a base constructor",
			"class A { constructor() { var C = class extends D { constructor() { super(); this.x(); } }; this.y(); } }",
		},
		// A static block is a separate evaluation context and runs before any instance exists, so
		// a `this` in it refers to the class rather than the instance.
		{
			"a static block in a derived class",
			"class A extends B { static { this.registry = []; } constructor() { super(); } }",
		},
		// A `try` with both a catch and a finally, where every path calls super before the this.
		// The scanner scans the finally from the entry state, so a `super()` inside the finally
		// itself is what makes this clean rather than any credit carried out of the try.
		{
			"a finally that calls super before touching this",
			"class A extends B { constructor() { try { foo(); } finally { super(); this.a(); } } }",
		},
		// Credit carried into a loop body from before the loop must survive, or every `this` after
		// a `super()` inside any loop would report. Upstream's `for (const b of a) { this.a(); }`
		// pass case covers for-of; this covers the plain `for` and the `while`.
		{
			"this inside a while body after super",
			"class A extends B { constructor() { super(); while (foo) { this.a(); } } }",
		},
		// Credit from a `for` initializer is kept, since it runs exactly once and before the body.
		{
			"super in a for initializer credits the rest",
			"class A extends B { constructor() { for (var i = super(); false; ) { } this.a(); } }",
		},
		// A nested arrow inside a nested function, two levels deep. The walk has to stop at the
		// first boundary rather than at the deepest one.
		{
			"this two evaluation contexts deep",
			"class A extends B { constructor() { function f() { return () => this.d(); } super(); } }",
		},
		// A field initializer using `super.x` rather than `this`, in a derived class. Field
		// initializers run after `super()` regardless of which of the two they read.
		{
			"a field initializer reading a super property",
			"class C extends B { field = super.toString(); constructor() { super(); } }",
		},
		// A do-while body runs at least once, so its `super()` is guaranteed and the credit
		// survives. This port first grouped do-while with the other loops and reported here;
		// both `oxlint` and ESLint disagreed, which is how the grouping was found to be wrong.
		{
			"a do-while calling super in its body",
			"class A extends B { constructor() { do { super(); } while (foo); this.a(); } }",
		},
		// A switch with a default where every clause that can leave has called.
		{
			"a switch whose every clause calls super",
			"class A extends B { constructor(x) { switch (x) { case 1: super(); break; default: super(); } this.a(); } }",
		},
		// An empty clause falls through into the one below and inherits its verdict, so `case 1:`
		// contributing nothing does not sink the switch.
		{
			"a switch with an empty clause falling through into a calling one",
			"class A extends B { constructor(x) { switch (x) { case 1: case 2: super(); break; default: super(); } this.a(); } }",
		},
		// A switch that is only a default is unconditional.
		{
			"a switch that is only a default clause",
			"class A extends B { constructor(x) { switch (x) { default: super(); } this.a(); } }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, testCase.sourceText))
		})
	}
}

// TestNoThisBeforeSuperKeepsTheCallAcrossElselessIfs holds #7t7c7g7: an `if` with no `else` keeps
// the state it was entered with. It reset the state to "not called" after every such `if`, so a
// constructor calling `super()` first and then assigning `this` in two or more sequential `if` blocks
// reported: BaseError.ts:174 and eight FormComponentData constructors in api-phi-health. The
// reporter's repro at one, two and six blocks stays silent; the nearest real violations still report.
func TestNoThisBeforeSuperKeepsTheCallAcrossElselessIfs(t *testing.T) {
	t.Parallel()

	constructor := func(blocks int) string {
		source := "class A extends B { constructor(flag) { super();"
		for index := 0; index < blocks; index++ {
			source += " if (flag) { this.name = 'n'; }"
		}
		return source + " } }"
	}
	for _, blocks := range []int{1, 2, 6} {
		t.Run(fmt.Sprintf("super first, then %d else-less ifs", blocks), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, constructor(blocks)))
		})
	}

	reporting := []struct{ name, sourceText string }{
		// `this` inside the second if, with super() only after both.
		{"this in the second if, before super", "class A extends B { constructor(flag) { if (flag) { this.a = 1; } if (flag) { this.b = 2; } super(); } }"},
		// super() only in an else-less if, so the second if's this may run without it.
		{"super in an else-less if, this in the next", "class A extends B { constructor(flag) { if (flag) { super(); } if (flag) { this.b = 2; } } }"},
	}
	for _, testCase := range reporting {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoThisBeforeSuper, thisBeforeSuperFile, testCase.sourceText), "thisBeforeSuper")
		})
	}
}
