package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// constructorSuperFile is where the fixtures pretend to live.
//
// A `.js` extension on purpose. The rule is about a JavaScript runtime error that TypeScript
// already refuses to compile, but it declines no file, so the extension is only about honesty:
// these are JavaScript programs.
const constructorSuperFile = "/repository/source/Subject.js"

// The corpus is oxc's, copied byte for byte rather than retyped.
//
// `oxc/crates/oxc_linter/src/rules/eslint/constructor_super.rs` carries one Tester block with 44
// pass and 43 fail inputs, and `eslint_constructor_super.snap` records exactly 43 diagnostics, so
// one finding per fail input is measured rather than assumed. Every string below was pulled out of
// the Rust source by a script and written straight to this file, because a fixture a porter retypes
// encodes the same belief as the port it is meant to check.
//
// The per-input message id was recovered by pairing each snapshot entry with the source line it
// prints, not by walking the two lists in order. That distinction is load-bearing here because the
// three trailing loop cases are near-identical and an in-order walk cannot tell a misalignment from
// a match.

// TestConstructorSuperFires is the whole fail corpus, one subtest per input, asserting the exact
// message id upstream's snapshot records for it.
//
// Grouped by message id rather than left in corpus order, because the grouping is the rule's real
// shape: four judgments, and which one an input lands in is the entire question.
func TestConstructorSuperFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messageId  string
	}{
		// missingAll: 15 inputs
		{`class A extends null { constructor() { } }`, `class A extends null { constructor() { } }`, `missingAll`},
		{`class A extends B { constructor() { } }`, `class A extends B { constructor() { } }`, `missingAll`},
		{`class A extends B { constructor() { for (var a of b) super.f`, `class A extends B { constructor() { for (var a of b) super.foo(); } }`, `missingAll`},
		{`class A extends B { constructor() { for (var i = 1; i < 10; `, `class A extends B { constructor() { for (var i = 1; i < 10; i++) super.foo(); } }`, `missingAll`},
		{`class A extends B { constructor() { var c = class extends D `, `class A extends B { constructor() { var c = class extends D { constructor() { super(); } } } }`, `missingAll`},
		{`class A extends B { constructor() { var c = () => super(); }`, `class A extends B { constructor() { var c = () => super(); } }`, `missingAll`},
		{`class A extends B { constructor() { class C extends D { cons`, `class A extends B { constructor() { class C extends D { constructor() { super(); } } } }`, `missingAll`},
		{`class A extends B { constructor() { var C = class extends D `, `class A extends B { constructor() { var C = class extends D { constructor() { super(); } } } }`, `missingAll`},
		{`class A extends B { constructor() { super(); class C extends`, `class A extends B { constructor() { super(); class C extends D { constructor() { } } } }`, `missingAll`},
		{`class A extends B { constructor() { super(); var C = class e`, `class A extends B { constructor() { super(); var C = class extends D { constructor() { } } } }`, `missingAll`},
		{`class A extends B { constructor() { a && super(); } }`, `class A extends B { constructor() { a && super(); } }`, `missingAll`},
		{`class A extends B { constructor() { return; super(); } }`, `class A extends B { constructor() { return; super(); } }`, `missingAll`},
		{`class Foo extends Bar { constructor() { for (a in b) for (c `, `class Foo extends Bar {
                            constructor() {
                                for (a in b) for (c in d);
                            }
                        }`, `missingAll`},
		{`class C extends D { constructor() { do { something(); } whil`, `class C extends D {

                            constructor() {
                                do {
                                    something();
                                } while (foo);
                            }

                        }`, `missingAll`},
		{`class C extends D { constructor() { for (let i = 1;;i++) { i`, `class C extends D {

                            constructor() {
                                for (let i = 1;;i++) {
                                    if (bar) {
                                        break;
                                    }
                                }
                            }

                        }`, `missingAll`},
		// missingSome: 13 inputs
		{`class A extends B { constructor() { if (a) super(); } }`, `class A extends B { constructor() { if (a) super(); } }`, `missingSome`},
		{`class A extends B { constructor() { if (a); else super(); } `, `class A extends B { constructor() { if (a); else super(); } }`, `missingSome`},
		{`class A extends B { constructor() { switch (a) { case 0: sup`, `class A extends B { constructor() { switch (a) { case 0: super(); } } }`, `missingSome`},
		{`class A extends B { constructor() { switch (a) { case 0: bre`, `class A extends B { constructor() { switch (a) { case 0: break; default: super(); } } }`, `missingSome`},
		{`class A extends B { constructor() { try { super(); } catch (`, `class A extends B { constructor() { try { super(); } catch (err) {} } }`, `missingSome`},
		{`class A extends B { constructor() { try { a; } catch (err) {`, `class A extends B { constructor() { try { a; } catch (err) { super(); } } }`, `missingSome`},
		{`class A extends B { constructor() { if (a) return; super(); `, `class A extends B { constructor() { if (a) return; super(); } }`, `missingSome`},
		{`class A extends B { constructor(a) { while (a) super(); } }`, `class A extends B { constructor(a) { while (a) super(); } }`, `missingSome`},
		{`class C extends D { constructor() { do { super(); } while (f`, `class C extends D {

                            constructor() {
                                do {
                                    super();
                                } while (foo);
                            }

                        }`, `missingSome`},
		{`class C extends D { constructor() { while (foo) { if (bar) {`, `class C extends D {

                            constructor() {
                                while (foo) {
                                    if (bar) {
                                        super();
                                        break;
                                    }
                                }
                            }

                        }`, `missingSome`},
		{`class A extends B { constructor() { while (condition) { if (`, `class A extends B {
            constructor() {
                while (condition) {
                    if (x) { super(); }
                }
            }
        }`, `missingSome`},
		{`class A extends B { constructor() { for (let i = 0; i < 10; `, `class A extends B {
            constructor() {
                for (let i = 0; i < 10; i++) {
                    if (i === 5) { super(); }
                }
            }
        }`, `missingSome`},
		{`class A extends B { constructor() { do { if (x) { super(); }`, `class A extends B {
            constructor() {
                do {
                    if (x) { super(); }
                } while (condition);
            }
        }`, `missingSome`},
		// badSuper: 11 inputs
		{`class A extends null { constructor() { super(); } }`, `class A extends null { constructor() { super(); } }`, `badSuper`},
		{`class A extends 100 { constructor() { super(); } }`, `class A extends 100 { constructor() { super(); } }`, `badSuper`},
		{`class A extends 'test' { constructor() { super(); } }`, `class A extends 'test' { constructor() { super(); } }`, `badSuper`},
		{`class A extends (B = 5) { constructor() { super(); } }`, `class A extends (B = 5) { constructor() { super(); } }`, `badSuper`},
		{`class A extends (B && 5) { constructor() { super(); } }`, `class A extends (B && 5) { constructor() { super(); } }`, `badSuper`},
		{`class A extends (B &&= 5) { constructor() { super(); } }`, `class A extends (B &&= 5) { constructor() { super(); } }`, `badSuper`},
		{`class A extends (B += C) { constructor() { super(); } }`, `class A extends (B += C) { constructor() { super(); } }`, `badSuper`},
		{`class A extends (B -= C) { constructor() { super(); } }`, `class A extends (B -= C) { constructor() { super(); } }`, `badSuper`},
		{`class A extends (B **= C) { constructor() { super(); } }`, `class A extends (B **= C) { constructor() { super(); } }`, `badSuper`},
		{`class A extends (B |= C) { constructor() { super(); } }`, `class A extends (B |= C) { constructor() { super(); } }`, `badSuper`},
		{`class A extends (B &= C) { constructor() { super(); } }`, `class A extends (B &= C) { constructor() { super(); } }`, `badSuper`},
		// duplicate: 4 inputs
		{`class A extends B { constructor() { super(); super(); } }`, `class A extends B { constructor() { super(); super(); } }`, `duplicate`},
		{`class A extends B { constructor() { super() || super(); } }`, `class A extends B { constructor() { super() || super(); } }`, `duplicate`},
		{`class A extends B { constructor() { if (a) super(); super();`, `class A extends B { constructor() { if (a) super(); super(); } }`, `duplicate`},
		{`class A extends B { constructor() { switch (a) { case 0: sup (2)`, `class A extends B { constructor() { switch (a) { case 0: super(); default: super(); } } }`, `duplicate`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText),
				testCase.messageId)
		})
	}
}

// TestConstructorSuperStaysSilent is upstream's 44 pass inputs, and it is the half that catches a
// port rather than confirming it.
//
// Several of them exist because somebody hit the bug they encode. `class A extends (5 && B)` is
// clean because `&&` yields its right operand, so a classifier looking at the literal on the left
// reports it wrongly. `for (const a of list) { if (a.foo) { super(a); return; } } super();` is
// clean because the loop body may run zero times and the trailing call covers that, which is the
// same reasoning that makes the loop cases in the fail list fire.
func TestConstructorSuperStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{`class A { }`, `class A { }`},
		{`class A { constructor() { } }`, `class A { constructor() { } }`},
		{`class A extends null { }`, `class A extends null { }`},
		{`class A extends B { }`, `class A extends B { }`},
		{`class A extends B { constructor() { super(); } }`, `class A extends B { constructor() { super(); } }`},
		{`class A extends B { constructor() { if (true) { super(); } e`, `class A extends B { constructor() { if (true) { super(); } else { super(); } } }`},
		{`class A extends (class B {}) { constructor() { super(); } }`, `class A extends (class B {}) { constructor() { super(); } }`},
		{`class A extends (B = C) { constructor() { super(); } }`, `class A extends (B = C) { constructor() { super(); } }`},
		{`class A extends (B &&= C) { constructor() { super(); } }`, `class A extends (B &&= C) { constructor() { super(); } }`},
		{`class A extends (B ||= C) { constructor() { super(); } }`, `class A extends (B ||= C) { constructor() { super(); } }`},
		{`class A extends (B ??= C) { constructor() { super(); } }`, `class A extends (B ??= C) { constructor() { super(); } }`},
		{`class A extends (B ||= 5) { constructor() { super(); } }`, `class A extends (B ||= 5) { constructor() { super(); } }`},
		{`class A extends (B ??= 5) { constructor() { super(); } }`, `class A extends (B ??= 5) { constructor() { super(); } }`},
		{`class A extends (B || C) { constructor() { super(); } }`, `class A extends (B || C) { constructor() { super(); } }`},
		{`class A extends (5 && B) { constructor() { super(); } }`, `class A extends (5 && B) { constructor() { super(); } }`},
		{`class A extends (false && B) { constructor() { super(); } }`, `class A extends (false && B) { constructor() { super(); } }`},
		{`class A extends (B || 5) { constructor() { super(); } }`, `class A extends (B || 5) { constructor() { super(); } }`},
		{`class A extends (B ?? 5) { constructor() { super(); } }`, `class A extends (B ?? 5) { constructor() { super(); } }`},
		{`class A extends (a ? B : C) { constructor() { super(); } }`, `class A extends (a ? B : C) { constructor() { super(); } }`},
		{`class A extends (B, C) { constructor() { super(); } }`, `class A extends (B, C) { constructor() { super(); } }`},
		{`class A { constructor() { class B extends C { constructor() `, `class A { constructor() { class B extends C { constructor() { super(); } } } }`},
		{`class A extends B { constructor() { super(); class C extends`, `class A extends B { constructor() { super(); class C extends D { constructor() { super(); } } } }`},
		{`class A extends B { constructor() { super(); class C { const`, `class A extends B { constructor() { super(); class C { constructor() { } } } }`},
		{`class A extends B { constructor() { a ? super() : super(); }`, `class A extends B { constructor() { a ? super() : super(); } }`},
		{`class A extends B { constructor() { if (a) super(); else sup`, `class A extends B { constructor() { if (a) super(); else super(); } }`},
		{`class A extends B { constructor() { switch (a) { case 0: sup`, `class A extends B { constructor() { switch (a) { case 0: super(); break; default: super(); } } }`},
		{`class A extends B { constructor() { try {} finally { super()`, `class A extends B { constructor() { try {} finally { super(); } } }`},
		{`class A extends B { constructor() { if (a) throw Error(); su`, `class A extends B { constructor() { if (a) throw Error(); super(); } }`},
		{`class A extends B { constructor() { if (true) return a; supe`, `class A extends B { constructor() { if (true) return a; super(); } }`},
		{`class A extends null { constructor() { return a; } }`, `class A extends null { constructor() { return a; } }`},
		{`class A { constructor() { return a; } }`, `class A { constructor() { return a; } }`},
		{`class A extends B { constructor(a) { super(); for (const b o`, `class A extends B { constructor(a) { super(); for (const b of a) { this.a(); } } }`},
		{`class A extends B { constructor(a) { super(); for (b in a) (`, `class A extends B { constructor(a) { super(); for (b in a) ( foo(b) ); } }`},
		{`class Foo extends Object { constructor(method) { super(); th`, `class Foo extends Object { constructor(method) { super(); this.method = method || function() {}; } }`},
		{`class A extends Object { constructor() { super(); for (let i`, `class A extends Object {
                constructor() {
                    super();
                    for (let i = 0; i < 0; i++);
                }
            }
            `},
		{`class A extends Object { constructor() { super(); for (; i <`, `class A extends Object {
                constructor() {
                    super();
                    for (; i < 0; i++);
                }
            }
            `},
		{`class A extends Object { constructor() { super(); for (let i (2)`, `class A extends Object {
                constructor() {
                    super();
                    for (let i = 0;; i++) {
                        if (foo) break;
                    }
                }
            }
            `},
		{`class A extends Object { constructor() { super(); for (let i (3)`, `class A extends Object {
                constructor() {
                    super();
                    for (let i = 0; i < 0;);
                }
            }
            `},
		{`class A extends Object { constructor() { super(); for (let i (4)`, `class A extends Object {
                constructor() {
                    super();
                    for (let i = 0;;) {
                        if (foo) break;
                    }
                }
            }
            `},
		{`class A extends B { constructor(props) { super(props); try {`, `
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
		{`class A extends obj?.prop { constructor() { super(); } }`, `class A extends obj?.prop { constructor() { super(); } }`},
		{`class A extends Base { constructor(list) { for (const a of l`, `
                        class A extends Base {
                            constructor(list) {
                                for (const a of list) {
                                    if (a.foo) {
                                        super(a);
                                        return;
                                    }
                                }
                                super();
                            }
                        }
                    `},
		{`class A extends B { constructor() { super(); try { throw new`, `class A extends B { constructor() { super(); try { throw new Error(); } catch (e) {} } }`},
		{`class A extends B { constructor() { try { mayThrow(); } catc`, `class A extends B { constructor() { try { mayThrow(); } catch (e) {} finally { super(); } } }`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText))
		})
	}
}

// TestConstructorSuperPointsAtTheRightNode is the assertion `ExpectFindings` structurally cannot
// make, and the reason it is here rather than assumed.
//
// Message ids and counts were green for the whole 87-input corpus while this was unwritten, and a
// rule that reports the right id at the wrong place passes every one of them. This rule has two
// different anchors, which doubles the chance of getting one wrong: `missingAll` and `missingSome`
// point at the constructor, because there is no single place a missing call belongs, while
// `duplicate` and `badSuper` point at the offending call, because there is one and deleting it is
// the fix. Upstream's snapshot records both and each expectation below is a column read out of it.
func TestConstructorSuperPointsAtTheRightNode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpans  []string
	}{
		{
			"a missing call anchors on the constructor, not the class",
			"class A extends B { constructor() { } }",
			[]string{"constructor() { }"},
		},
		{
			"the constructor span stops at its own closing brace",
			"class A extends B { constructor() { if (a) super(); } }",
			[]string{"constructor() { if (a) super(); }"},
		},
		{
			"an inner class reports at the inner constructor",
			"class A extends B { constructor() { super(); class C extends D { constructor() { } } } }",
			[]string{"constructor() { }"},
		},
		{
			"a bad call anchors on the call rather than the extends clause",
			"class A extends null { constructor() { super(); } }",
			[]string{"super()"},
		},
		{
			"a duplicate anchors on the second call, not the first",
			"class A extends B { constructor() { super(); super(); } }",
			[]string{"super()"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("expected %d findings, got %d: %v",
					len(testCase.wantSpans), len(result.Diagnostics), result.MessageIds())
			}
			for index, diagnostic := range result.Diagnostics {
				got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.wantSpans[index] {
					t.Errorf("finding %d: reported %q, want %q",
						index, got, testCase.wantSpans[index])
				}
			}
		})
	}
}

// TestConstructorSuperReportsTheSecondCallNotTheFirst is the span assertion that needed its own
// test, because the source holds two byte-identical `super()` calls and comparing text cannot tell
// them apart.
//
// Slicing the reported range out of the source yields `super()` either way, so this compares
// offsets instead. Upstream's snapshot pins it at column 46 for
// `class A extends B { constructor() { super(); super(); } }`, which is the second call, and a rule
// reporting the first would be green under the test above and wrong.
func TestConstructorSuperReportsTheSecondCallNotTheFirst(t *testing.T) {
	t.Parallel()

	sourceText := "class A extends B { constructor() { super(); super(); } }"
	result := rule_testing.Run(t, ConstructorSuper, constructorSuperFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected 1 finding, got %d: %v", len(result.Diagnostics), result.MessageIds())
	}

	firstCall := strings.Index(sourceText, "super()")
	secondCall := strings.Index(sourceText[firstCall+1:], "super()") + firstCall + 1
	if got := result.Diagnostics[0].Range.Pos(); got != secondCall {
		t.Errorf("reported at offset %d, want the second call at %d (the first is at %d)",
			got, secondCall, firstCall)
	}
}

// TestConstructorSuperReportsTheRightOperandOfALogicalOr is the other pair of identical calls, and
// the one where reporting the wrong half would be a real defect rather than a cosmetic one.
//
// `super() || super()` reports its *right* operand. The left one is the call that actually runs, so
// deleting it would be the wrong repair, and upstream's snapshot pins column 48 rather than 37.
func TestConstructorSuperReportsTheRightOperandOfALogicalOr(t *testing.T) {
	t.Parallel()

	sourceText := "class A extends B { constructor() { super() || super(); } }"
	result := rule_testing.Run(t, ConstructorSuper, constructorSuperFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected 1 finding, got %d: %v", len(result.Diagnostics), result.MessageIds())
	}

	rightOperand := strings.Index(sourceText, "|| super()") + len("|| ")
	if got := result.Diagnostics[0].Range.Pos(); got != rightOperand {
		t.Errorf("reported at offset %d, want the right operand at %d", got, rightOperand)
	}
}

// TestConstructorSuperRendersItsMessages asserts the rendered description exactly.
//
// Exactly, and by equality rather than by `strings.Contains`, because a predicate weaker than the
// property it guards is not a guard. A description built by concatenating string literals is the
// shape that loses a space at a seam, and a `Contains` check on a fragment would stay green through
// it.
func TestConstructorSuperRendersItsMessages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		sourceText      string
		wantId          string
		wantDescription string
	}{
		{
			"missingAll",
			"class A extends B { constructor() { } }",
			"missingAll",
			"This derived constructor never calls `super()`, so constructing the class throws " +
				"a ReferenceError before the body finishes.",
		},
		{
			"missingSome",
			"class A extends B { constructor() { if (a) super(); } }",
			"missingSome",
			"This derived constructor calls `super()` on some paths and not others, so the " +
				"paths without it throw a ReferenceError.",
		},
		{
			"duplicate",
			"class A extends B { constructor() { super(); super(); } }",
			"duplicate",
			"This `super()` call can run on a path that already called `super()`, which throws " +
				"a ReferenceError because a constructor may only call it once.",
		},
		{
			"badSuper",
			"class A extends null { constructor() { super(); } }",
			"badSuper",
			"This `super()` call has no constructor to call, because the `extends` clause names " +
				"something that cannot be constructed.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected 1 finding, got %d: %v",
					len(result.Diagnostics), result.MessageIds())
			}
			if got := result.Diagnostics[0].Message.Id; got != testCase.wantId {
				t.Errorf("message id %q, want %q", got, testCase.wantId)
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.wantDescription {
				t.Errorf("description:\n got %q\nwant %q", got, testCase.wantDescription)
			}
		})
	}
}

// TestConstructorSuperClassifiesUnconstructableExtends is the deliberate divergence from oxc,
// pinned so a later reader can see it was a decision rather than an accident.
//
// None of these inputs is in the imported corpus, which is why the corpus cannot separate the two
// classifiers. Oxc writes a blacklist and calls every one of them constructable, so it reports
// nothing. ESLint writes a whitelist and reports `badSuper` on each. ESLint is right in every case:
// none of these can be on the right of a `new`, so the `super()` throws a TypeError at runtime.
// This rule uses the whitelist, so it agrees with ESLint here and with oxc everywhere oxc has an
// opinion the corpus records.
func TestConstructorSuperClassifiesUnconstructableExtends(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"undefined", "class A extends undefined { constructor() { super(); } }"},
		{"an array literal", "class A extends [] { constructor() { super(); } }"},
		{"an object literal", "class A extends ({}) { constructor() { super(); } }"},
		{"a template literal", "class A extends `x` { constructor() { super(); } }"},
		{"a regular expression", "class A extends /re/ { constructor() { super(); } }"},
		{"a void expression", "class A extends (void 0) { constructor() { super(); } }"},
		{"an arrow function", "class A extends (() => {}) { constructor() { super(); } }"},
		// A conditional whose branches are both unconstructable. The corpus has the mixed form
		// `(a ? B : C)` as a pass case, so this is the other half of that discrimination.
		{"a conditional of two literals", "class A extends (a ? 1 : 2) { constructor() { super(); } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText),
				"badSuper")
		})
	}
}

// TestConstructorSuperAcceptsConstructableExtendsOurCodeWrites covers the shapes our own tree holds
// that the corpus does not, and it is the half that would catch the whitelist being too narrow.
//
// A whitelist errs toward reporting, so its failure mode is a false positive on valid code, and
// these are the valid spellings a TypeScript codebase actually contains. The last two exist only
// because we lint TypeScript and ESLint's list was written for JavaScript: a non-null assertion and
// an `as` cast are both erasures that leave the operand behind at runtime, so a class extending one
// is constructable exactly when the operand is.
func TestConstructorSuperAcceptsConstructableExtendsOurCodeWrites(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a namespaced identifier", "class A extends Foo.Bar { constructor() { super(); } }"},
		{"a subscripted member", "class A extends Foo['Bar'] { constructor() { super(); } }"},
		{"a mixin call", "class A extends Mixin(Base) { constructor() { super(); } }"},
		{"a new expression", "class A extends new Factory() { constructor() { super(); } }"},
		{"a tagged template", "class A extends tag`x` { constructor() { super(); } }"},
		{"a non-null assertion", "class A extends Base! { constructor() { super(); } }"},
		{"an as cast", "class A extends (Base as typeof Base) { constructor() { super(); } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText))
		})
	}
}

// TestConstructorSuperCountsOneFindingPerExtraCall pins that `duplicate` is per-call rather than
// per-constructor.
//
// The corpus has no input with three calls, so this is invented, and ESLint was run on it rather
// than guessed at: `super(); super(); super();` reports two duplicates, not one. A rule reporting
// once per constructor would be green across the whole imported corpus, because every duplicate
// case in it holds exactly two calls.
func TestConstructorSuperCountsOneFindingPerExtraCall(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, ConstructorSuper, constructorSuperFile,
			"class A extends B { constructor() { super(); super(); super(); } }"),
		"duplicate", "duplicate")
}

// TestConstructorSuperIgnoresCallsInNestedEvaluationContexts is the question a `super()` token
// inside the constructor cannot answer on its own.
//
// The corpus covers the arrow and the nested class. A function expression and a nested method are
// not in it, and they are the cases where the answer is reached by a different argument: an arrow
// inherits the enclosing `super` binding lexically and a function expression does not, so a reader
// might expect them to differ. They do not, because what the rule measures is when the body
// evaluates rather than which binding it would see, and neither body runs at the point it is
// written.
//
// The first case is not valid JavaScript, since `super()` in a plain function is a syntax error,
// but our parser produces a tree for it rather than refusing the file, so the rule still has to
// have the right opinion about it.
func TestConstructorSuperIgnoresCallsInNestedEvaluationContexts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"a function expression",
			"class A extends B { constructor() { var f = function () { super(); }; } }",
		},
		{
			"an immediately invoked arrow",
			"class A extends B { constructor() { (() => { super(); })(); } }",
		},
		{
			"a nested object method",
			"class A extends B { constructor() { var o = { m() { super(); } }; } }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText),
				"missingAll")
		})
	}
}

// TestConstructorSuperHandlesBaseAndNullExtendsSeparately pins that `extends null` is a third case
// rather than a spelling of one of the other two.
//
// It is neither a base class nor a properly derived one, and it is the only `extends` clause where
// a class is obliged to bind `this` and forbidden from calling `super()` to do it. Three of the
// four rows come from the corpus and are asserted again here as a group, because the discrimination
// only reads as a discrimination when the four sit together.
func TestConstructorSuperHandlesBaseAndNullExtendsSeparately(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a base class needs nothing", "class A { constructor() { } }", nil},
		{"extends null with no constructor is fine", "class A extends null { }", nil},
		{
			"extends null with an empty constructor binds no this",
			"class A extends null { constructor() { } }",
			[]string{"missingAll"},
		},
		{
			"extends null returning an object is the only valid form",
			"class A extends null { constructor() { return a; } }",
			nil,
		},
		{
			"extends null calling super has nothing to call",
			"class A extends null { constructor() { super(); } }",
			[]string{"badSuper"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestConstructorSuperDeclinesAConstructorOverloadSignature is a shape upstream's corpus cannot
// contain, because oxc parses these fixtures as JavaScript and an overload signature is TypeScript.
//
// A signature has no body, so there is no path through it to be missing anything on. Reporting one
// would be a false positive on every overloaded constructor in the tree, and the implementation
// right below it is the declaration that actually has to call `super()`.
func TestConstructorSuperDeclinesAConstructorOverloadSignature(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, ConstructorSuper, constructorSuperFile, `
		class A extends B {
			constructor(value: string);
			constructor(value: number);
			constructor(value: string | number) { super(value); }
		}
	`))
}

// TestConstructorSuperDoesNotBlameANestedClassForAnUnconstructableExtends is a blind spot a
// mutation sweep found, and it is the one place where the nested-context filter is load-bearing and
// nothing else covers it.
//
// The corpus does cover the filter, but only through inputs whose outer class has a constructable
// superclass, and on those the path analysis reaches the same verdict by its own separate walk. So
// a rule whose collection step forgot to stop at a nested class stayed green across all 87 of them.
// The input that separates them needs both halves at once: an outer class that cannot be
// constructed, and a nested class that legitimately calls `super()`.
//
// Without the filter the inner call is collected, the outer class reads as having a call to
// complain about, and the finding lands on the inner `super()` as `badSuper`, which is wrong twice
// over: that call is correct, and the outer constructor's real problem goes unreported. ESLint was
// run on this input rather than reasoned about, and it reports `missingAll` on the outer
// constructor.
func TestConstructorSuperDoesNotBlameANestedClassForAnUnconstructableExtends(t *testing.T) {
	t.Parallel()

	sourceText := "class A extends null { constructor() { class C extends D { constructor() { super(); } } } }"
	result := rule_testing.Run(t, ConstructorSuper, constructorSuperFile, sourceText)
	rule_testing.ExpectFindings(t, result, "missingAll")

	// The span matters as much as the id here, since the whole failure mode is the finding landing
	// on the inner class's correct call. Asserting the outer constructor is what makes this test
	// see the defect rather than merely notice a count.
	if got := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]; !strings.HasPrefix(got, "constructor() { class C") {
		t.Errorf("reported at %q, want the outer constructor", got)
	}
}

// TestConstructorSuperAcceptsAMixedShortCircuit covers the operand positions the corpus tests in
// only one direction, and each row is a mutation sweep survivor.
//
// The corpus has `(a ? B : C)` clean and, added here, `(a ? 1 : 2)` reporting, so a classifier
// demanding *both* conditional branches be constructable is green across all of it: both-true and
// both-false agree with either-true under those two inputs. The mixed forms are what separate them,
// and they are clean, because only one branch runs and the rule reports only what it can prove
// wrong.
//
// The same asymmetry runs through `||`, `??` and the sequence operator. The corpus has `(B || 5)`
// and `(B ?? 5)` with the constructable operand on the left, so nothing in it notices a classifier
// that reads only that side. `(5 || B)` and `(5 ?? B)` put it on the right, and both were run
// through ESLint rather than reasoned about.
//
// The last row is the counterexample that keeps this from proving too much. A sequence expression
// really does yield only its last operand, so `(5, B)` is clean and `(B, 5)` is not, and a
// classifier treating a comma like an `||` would get the second one wrong.
func TestConstructorSuperAcceptsAMixedShortCircuit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a conditional whose consequent is constructable",
			"class A extends (a ? B : 5) { constructor() { super(); } }"},
		{"a conditional whose alternate is constructable",
			"class A extends (a ? 5 : B) { constructor() { super(); } }"},
		{"an or whose left operand is constructable",
			"class A extends (5 || B) { constructor() { super(); } }"},
		{"a nullish coalesce whose left operand is constructable",
			"class A extends (5 ?? B) { constructor() { super(); } }"},
		{"a sequence whose last operand is constructable",
			"class A extends (5, B) { constructor() { super(); } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText))
		})
	}
}

// TestConstructorSuperTreatsLoopBodiesAsOneIteration is where oxc and ESLint knowingly part
// company, and it is two mutation sweep survivors rather than one.
//
// A loop body holding a call is reachable twice, so ESLint reports the repetition:
// `while (a) super();` is `missingSome` *and* `duplicate` there. Oxc reports only the
// `missingSome`, because its duplicate branch is gated on having more than one `super()` span and a
// body with one call holds one span. Four corpus inputs pin oxc's reading, each recording exactly
// one diagnostic, so oxc's is what this reproduces.
//
// The gate is per-span rather than per-loop, which is the distinction the first sweep survivor
// exposed. Suppressing every duplicate inside a loop body reproduces the corpus just as well and is
// wrong for a different input: `while (a) { super(); super(); }` runs both calls on a single
// iteration, so the second is a duplicate under any reading, and oxc's span gate lets it through.
// Both spellings are asserted here, because the pair is what separates "one call reached twice" from
// "two calls reached once" and neither alone does.
//
// The third row is the other survivor. A call in a loop body still sets the state for what follows
// the loop, so `super(); while (a) super();` reports the loop's call as a duplicate of the earlier
// one. A walk that discarded everything the body established would miss it.
func TestConstructorSuperTreatsLoopBodiesAsOneIteration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"one call reached twice is not a duplicate",
			"class A extends B { constructor(a) { while (a) super(); } }",
			[]string{"missingSome"},
		},
		{
			"two calls in one iteration is a duplicate",
			"class A extends B { constructor(a) { while (a) { super(); super(); } } }",
			[]string{"missingSome", "duplicate"},
		},
		{
			"a call before the loop makes the loop's call a duplicate",
			"class A extends B { constructor(a) { super(); while (a) super(); } }",
			[]string{"duplicate"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// TestConstructorSuperFindsNoDuplicateAmongUnreachableCalls is the other sweep survivor, and it is
// the input where "a second call" and "a second reachable call" come apart.
//
// `constructor() { return; super(); super(); }` holds two `super()` calls and neither runs. The
// finding is `missingAll` and nothing else, which ESLint confirms. A duplicate walk that did not
// stop at the `return` would find two calls, call the second a repeat of the first, and add a
// finding for code that never executes. The corpus has the one-call version of this and it cannot
// see the difference, because one call is never a duplicate however it is counted.
func TestConstructorSuperFindsNoDuplicateAmongUnreachableCalls(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, ConstructorSuper, constructorSuperFile,
			"class A extends B { constructor() { return; super(); super(); } }"),
		"missingAll")
}

// TestConstructorSuperReportsEveryBadCall pins that `badSuper` is per-call, which the corpus cannot
// see because every one of its eleven `badSuper` inputs holds exactly one call.
//
// A rule reporting only the first would be green across all of them. Each call is separately wrong
// and separately deletable, so each gets a finding, and that is oxc's reading: its
// unconstructable-superclass branch loops over every collected span and then returns.
//
// It is also where oxc and ESLint disagree about the *second* call. ESLint reports `badSuper` on
// the first and `duplicate` on the second, because its duplicate check runs before its
// constructable check and short-circuits. Oxc reports `badSuper` twice, which is the better answer:
// the second call is not a duplicate of a call that was never valid, it is the same error again.
// Oxc's is the one reproduced.
func TestConstructorSuperReportsEveryBadCall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"extends null", "class A extends null { constructor() { super(); super(); } }"},
		{"extends a literal", "class A extends 100 { constructor() { super(); super(); } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText),
				"badSuper", "badSuper")
		})
	}
}

// TestConstructorSuperRequiresARealValueFromAnExtendsNullConstructor covers the two ways the
// `extends null` escape hatch can be got wrong, and both are mutation sweep survivors.
//
// A class extending `null` cannot call `super()`, so the only way its constructor produces a bound
// `this` is by returning an object outright. The corpus has the working form,
// `constructor() { return a; }`, and it is clean. What it does not have is either of the near
// misses.
//
// A bare `return;` yields `undefined`, which the language then replaces with the unbound `this`, so
// it is not a substitute and the constructor is still wrong. And a `return a` written inside a
// nested function returns from that function rather than from the constructor, so it does not count
// either. Both report `missingAll`, confirmed by running ESLint on them rather than by reading.
func TestConstructorSuperRequiresARealValueFromAnExtendsNullConstructor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare return yields the unbound this",
			"class A extends null { constructor() { return; } }"},
		{"a return inside a nested function returns from that function",
			"class A extends null { constructor() { var f = function () { return a; }; } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, ConstructorSuper, constructorSuperFile, testCase.sourceText),
				"missingAll")
		})
	}
}
