package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const preferReturnThisTypeFile = "/repository/source/Returning.ts"

func preferReturnThisTypeCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestPreferReturnThisTypeStaysSilent is upstream's eight passing cases verbatim, plus shapes
// the corpus does not write whose silence was measured against the installed 8.67.0 build.
//
// Each case runs in its OWN program. That is not a stylistic choice: this rule compares types by
// REFERENCE identity, so a batched program that interned one `Foo` across several case files would
// make identity fail and hand back a clean verdict for a case that really reports. RunTyped is one
// file per call, so the fixtures are safe by construction; the oracle used to measure them had to
// be built that way deliberately.
func TestPreferReturnThisTypeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		// Upstream's valid list, byte for byte.
		"\nclass Foo {\n  f1() {}\n  f2(): Foo {\n    return new Foo();\n  }\n  f3() {\n    return this;\n  }\n  f4(): this {\n    return this;\n  }\n  f5(): any {\n    return this;\n  }\n  f6(): unknown {\n    return this;\n  }\n  f7(foo: Foo): Foo {\n    return Math.random() > 0.5 ? foo : this;\n  }\n  f10(this: Foo, that: Foo): Foo;\n  f11(): Foo {\n    return;\n  }\n  f13(this: Foo): Foo {\n    return this;\n  }\n  f14(): { f14: Function } {\n    return this;\n  }\n  f15(): Foo | this {\n    return Math.random() > 0.5 ? new Foo() : this;\n  }\n}\n    ",
		"\nclass Foo {\n  f1 = () => {};\n  f2 = (): Foo => {\n    return new Foo();\n  };\n  f3 = () => this;\n  f4 = (): this => {\n    return this;\n  };\n  f5 = (): Foo => new Foo();\n  f6 = '';\n}\n    ",
		"\nconst Foo = class {\n  bar() {\n    return this;\n  }\n};\n    ",
		"\nclass Base {}\nclass Derived extends Base {\n  f(): Base {\n    return this;\n  }\n}\n    ",
		"\nclass Foo {\n  accessor f = () => {\n    return this;\n  };\n}\n    ",
		"\nclass Foo {\n  accessor f = (): this => {\n    return this;\n  };\n}\n    ",
		"\nclass Foo {\n  f?: string;\n}\n    ",
		"\ndeclare const valueUnion: BaseUnion | string;\n\nclass BaseUnion {\n  f(): BaseUnion | string {\n    if (Math.random()) {\n      return this;\n    }\n\n    return valueUnion;\n  }\n}\n    ",

		// Measured on the installed build and silent there. Each pins a decision the corpus
		// leaves untested, and each would report under a plausible wrong port.

		// An anonymous class expression has no name to match, so the annotation `Foo` names the
		// const rather than the class. Its named sibling REPORTS and is in the fires list.
		"const Foo = class {\n  f(): Foo {\n    return this;\n  }\n};\n",

		// An intersection is invisible to upstream's union-only recursion. This is the same
		// mistake the rule exists to catch and upstream declines it, so this port declines it too.
		"class Foo {\n  f(): Foo & {} {\n    return this;\n  }\n}\n",

		// A qualified name is not an identifier, so the name test declines it.
		"namespace N {\n  export class Foo {\n    f(): N.Foo {\n      return this;\n    }\n  }\n}\n",

		// An abstract method has no body, so there is nothing to judge.
		"abstract class Foo {\n  abstract f(): Foo;\n}\n",

		// An interface is not a class and never reaches the listeners.
		"interface Foo {\n  f(): Foo;\n}\n",

		// An annotated local widens away the `this` type, so the returned value is the class type
		// rather than the receiver. Its unannotated sibling `const self = this` REPORTS and is
		// upstream's own case, which makes this pair the sharpest evidence that the judgment is a
		// type identity question rather than a syntactic one.
		"class Foo {\n  f(): Foo {\n    let self: Foo = this;\n    return self;\n  }\n}\n",

		// One return of the class type disqualifies the method however many returns of `this` sit
		// beside it. This is the `hasReturnClassType` priority, and nothing in upstream's corpus
		// writes a body with both.
		"class Foo {\n  f(): Foo {\n    if (Math.random()) {\n      return new Foo();\n    }\n    return this;\n  }\n}\n",

		// A `return this` inside a nested arrow returns from the ARROW. Neither upstream's
		// forEachReturnStatement nor ours descends into a nested function, and with the outer
		// return removed the method returns undefined. Measured silent.
		"class Foo {\n  f(): Foo | undefined {\n    [1].forEach(() => {\n      return this;\n    });\n  }\n}\n",

		// An async method returns a Promise, so the annotation is Promise<Foo> and the type
		// reference the rule matches is Promise rather than Foo.
		"class Foo {\n  async f(): Promise<Foo> {\n    return this;\n  }\n}\n",

		// An explicit `this` parameter says what the receiver is, so the author has already
		// answered the question. Upstream carries the method-signature form; this is the form with
		// a body, which is the one that could otherwise reach the return walk.
		"class Foo {\n  f(this: Foo, other: number): Foo {\n    return this;\n  }\n}\n",

		// A parenthesized `this` where the CHECKER declines. These three are the port's sharpest
		// fixtures, because an earlier revision reported all of them.
		//
		// Upstream's `this` keyword test runs on the TypeScript node, which keeps parentheses, so
		// `(this)` is never that keyword and always falls through to the type comparison. In a
		// static body `this` is the constructor, in a function expression it is `any`, and neither
		// is reference-identical to the class's this type, so the comparison declines and the case
		// is silent. Unwrapping the parentheses before the keyword test made all three report.
		//
		// The unparenthesized forms of the same three shapes REPORT and are in the fires list, so
		// this pair of lists is what holds the distinction in place from both sides.
		"class Foo {\n  static f(): Foo {\n    return (this);\n  }\n}\n",
		"class Foo {\n  f = function (): Foo {\n    return (this);\n  };\n}\n",
		"class Foo {\n  static f = (): Foo => (this);\n}\n",

		// A property initializer that is a TYPE ASSERTION, which is the shape the initializer kind
		// gate exists to keep out. It is crash protection rather than a behavioral filter: an
		// assertion node answers `Type()` (so it passes the return-type gate) and answers nil to
		// `Body()`, and `Parameters()` on a node that is not function-like dereferences nil. With
		// the gate removed these two panic, and the cohere walk recovers per FILE rather than per
		// rule, so one of them would cost every rule its verdict on that file.
		//
		// No ExpectFindings assertion can see a panic, which is why the gate survived a mutation
		// sweep until the surviving mutant was run against these inputs by hand. Both are silent
		// upstream, so they are recorded here as ordinary clean cases whose real job is to keep the
		// process alive.
		"class Foo {\n  f = <Foo>this;\n}\n",
		"class Foo {\n  f = this as Foo;\n}\n",
	}
	for index, sourceText := range cases {
		t.Run(preferReturnThisTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferReturnThisType,
				preferReturnThisTypeFile, sourceText))
		})
	}
}

// TestPreferReturnThisTypeFires is upstream's thirteen reporting cases verbatim with their own
// spans and their own repairs, plus eleven shapes upstream does not write whose behavior was
// measured against the installed 8.67.0 build.
//
// Every case reports exactly once, which is asserted by passing a single id: a port that reported
// an overload signature as well as its implementation would fail on the count alone.
func TestPreferReturnThisTypeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
		wantFixed  string
	}{
		{
			sourceText: "\nclass Foo {\n  f(): Foo {\n    return this;\n  }\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  f(): this {\n    return this;\n  }\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  f = function (): Foo {\n    return this;\n  };\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  f = function (): this {\n    return this;\n  };\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  f(): Foo {\n    const self = this;\n    return self;\n  }\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  f(): this {\n    const self = this;\n    return self;\n  }\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  f = (): Foo => {\n    return this;\n  };\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  f = (): this => {\n    return this;\n  };\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  f = (): Foo => {\n    const self = this;\n    return self;\n  };\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  f = (): this => {\n    const self = this;\n    return self;\n  };\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  f = (): Foo => this;\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  f = (): this => this;\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  accessor f = (): Foo => {\n    return this;\n  };\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  accessor f = (): this => {\n    return this;\n  };\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  accessor f = (): Foo => this;\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  accessor f = (): this => this;\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  f1(): Foo | undefined {\n    return this;\n  }\n  f2(): this | undefined {\n    return this;\n  }\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  f1(): this | undefined {\n    return this;\n  }\n  f2(): this | undefined {\n    return this;\n  }\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  bar(): Foo | undefined {\n    if (Math.random() > 0.5) {\n      return this;\n    }\n  }\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  bar(): this | undefined {\n    if (Math.random() > 0.5) {\n      return this;\n    }\n  }\n}\n      ",
		},
		{
			sourceText: "\nclass Foo {\n  bar(num: 1 | 2): Foo {\n    switch (num) {\n      case 1:\n        return this;\n      case 2:\n        return this;\n    }\n  }\n}\n      ",
			wantSpan:   "Foo",
			wantFixed:  "\nclass Foo {\n  bar(num: 1 | 2): this {\n    switch (num) {\n      case 1:\n        return this;\n      case 2:\n        return this;\n    }\n  }\n}\n      ",
		},
		{
			sourceText: "\nclass Animal<T> {\n  eat(): Animal<T> {\n    console.log(\"I'm moving!\");\n    return this;\n  }\n}\n      ",
			wantSpan:   "Animal<T>",
			wantFixed:  "\nclass Animal<T> {\n  eat(): this {\n    console.log(\"I'm moving!\");\n    return this;\n  }\n}\n      ",
		},
		{
			sourceText: "\ndeclare const valueUnion: number | string;\n\nclass BaseUnion {\n  f(): BaseUnion | string {\n    if (Math.random()) {\n      return this;\n    }\n\n    return valueUnion;\n  }\n}\n      ",
			wantSpan:   "BaseUnion",
			wantFixed:  "\ndeclare const valueUnion: number | string;\n\nclass BaseUnion {\n  f(): this | string {\n    if (Math.random()) {\n      return this;\n    }\n\n    return valueUnion;\n  }\n}\n      ",
		},
		// Beyond upstream's corpus. Each was measured on the installed build before it was
		// written, and each pins a discrimination the imported cases cannot reach.

		// A getter is a MethodDefinition in estree, so upstream's selector matches it. The corpus
		// writes no accessor at all, and a port anchored only on KindMethodDeclaration is silent here.
		{
			sourceText: "class Foo {\n  get f(): Foo {\n    return this;\n  }\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  get f(): this {\n    return this;\n  }\n}\n",
		},
		// A static method is in scope. `this` in a static body is the constructor, and the pair of
		// types being compared is still the class type and its this type.
		{
			sourceText: "class Foo {\n  static f(): Foo {\n    return this;\n  }\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  static f(): this {\n    return this;\n  }\n}\n",
		},
		// A NAMED class expression has a name to match, so the annotation is that name rather than the
		// const's. Its anonymous sibling is silent and sits in the other list.
		{
			sourceText: "const Foo = class Bar {\n  f(): Bar {\n    return this;\n  }\n};\n",
			wantSpan:   "Bar",
			wantFixed:  "const Foo = class Bar {\n  f(): this {\n    return this;\n  }\n};\n",
		},
		// A parenthesized return argument. Upstream's parser deletes the node before its rule sees it
		// and ours keeps it, so the `this` keyword fast path needs an unwrap to fire. The checker would
		// rescue this door on its own, which is why the concise-body case below is the one that proves
		// the unwrap is load-bearing rather than cosmetic.
		{
			sourceText: "class Foo {\n  f(): Foo {\n    return (this);\n  }\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  f(): this {\n    return (this);\n  }\n}\n",
		},
		// A parenthesized CONCISE BODY, and this is the case the unwrap exists for. That branch compares
		// the body's type against the class this type, and without the unwrap the fast path never fires
		// on a shape upstream reports. Measured on the installed build.
		{
			sourceText: "class Foo {\n  f = (): Foo => (this);\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  f = (): this => (this);\n}\n",
		},
		// Doubly parenthesized, because the unwrap is a loop rather than a single step.
		{
			sourceText: "class Foo {\n  f = (): Foo => ((this));\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  f = (): this => ((this));\n}\n",
		},
		// An overload signature followed by an implementation reports ONCE. The signature has no body,
		// so it is declined before the return walk, and a port missing that test reports twice.
		{
			sourceText: "class Foo {\n  f(): Foo;\n  f(): Foo {\n    return this;\n  }\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  f(): Foo;\n  f(): this {\n    return this;\n  }\n}\n",
		},
		// A type reference carrying type arguments. The span and the replaced range are the WHOLE
		// reference, so the repair writes `this` rather than `this<T>`.
		{
			sourceText: "class Foo<T> {\n  f(): Foo<T> {\n    return this;\n  }\n}\n",
			wantSpan:   "Foo<T>",
			wantFixed:  "class Foo<T> {\n  f(): this {\n    return this;\n  }\n}\n",
		},
		// A parenthesized union member inside a union. Type position keeps parens in our parser too, so
		// the union recursion needs its own unwrap, which is a site distinct from the expression one.
		{
			sourceText: "class Foo {\n  f(): (Foo | undefined) | null {\n    return this;\n  }\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  f(): (this | undefined) | null {\n    return this;\n  }\n}\n",
		},
		// The class name twice in one union reports ONCE and rewrites ONCE. Upstream's tester shows
		// `this | this` only because cohereAndFix re-lints until the text stops changing; a single
		// application of the same messages writes `this | Foo`, measured, and this port is one pass.
		{
			sourceText: "class Foo {\n  f(): Foo | Foo {\n    return this;\n  }\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  f(): this | Foo {\n    return this;\n  }\n}\n",
		},
		// A nested class shadowing the name does not change what the annotation resolves to, because
		// the annotation is read in the outer scope.
		{
			sourceText: "class Foo {\n  f(): Foo {\n    class Foo {}\n    return this;\n  }\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  f(): this {\n    class Foo {}\n    return this;\n  }\n}\n",
		},

		// Reaches unionContainsType with a NON-UNION type, which is the crash this rule shipped
		// briefly. `this` inside a plain function expression is `any`, an intrinsic, and
		// `AsUnionType` on an intrinsic panics rather than returning nil. Upstream's own corpus
		// writes this exact shape and could not see the defect, because the `this` keyword fast
		// path consumes the return before the union test is reached. This case is here so that a
		// later removal of that fast path, which is a pure cost optimization and therefore a
		// plausible thing to delete, fails loudly instead of taking every rule's verdict on the
		// file down with it.
		//
		// It is reported through ExpectFindings like any other case; what makes it a crash fixture
		// is the companion test below, which drives the same shape with the fast path bypassed.
		{
			sourceText: "class Foo {\n  f = function (): Foo {\n    return this;\n  };\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  f = function (): this {\n    return this;\n  };\n}\n",
		},

		// A setter and a constructor carrying a return type. The GRAMMAR forbids both, and the
		// PARSER produces both anyway under error recovery, which is the distinction that decided
		// these two arms. Probed directly: the parser hands back a KindSetAccessor and a
		// KindConstructor each with a KindTypeReference return type. Upstream reports both,
		// measured, because its selector is `ClassBody > MethodDefinition` and estree calls all
		// three of these a MethodDefinition.
		//
		// Both arms were added by analogy with the getter before any of this was measured, and both
		// survived a mutation sweep until these two cases existed, which is the imported corpus
		// being unable to express a shape upstream had no reason to write.
		{
			sourceText: "class Foo {\n  set f(v: Foo): Foo {\n    return this;\n  }\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  set f(v: Foo): this {\n    return this;\n  }\n}\n",
		},
		{
			sourceText: "class Foo {\n  constructor(): Foo {\n    return this;\n  }\n}\n",
			wantSpan:   "Foo",
			wantFixed:  "class Foo {\n  constructor(): this {\n    return this;\n  }\n}\n",
		},
	}
	for index, testCase := range cases {
		t.Run(preferReturnThisTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferReturnThisType, preferReturnThisTypeFile,
				testCase.sourceText)

			rule_testing.ExpectFindings(t, result, "useThisType")

			// The harness writes the fixture trimmed, so both the span slice and the expected
			// rewrite are taken against that text rather than against the Go literal above.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]

			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}

			// Asserted against a literal typed here rather than against the rule's own message
			// constant, because comparing a finding to the constant it was built from moves both
			// sides together under mutation and proves nothing.
			if diagnostic.Message.Description != "Use `this` type instead." {
				t.Fatalf("message: expected %q, got %q", "Use `this` type instead.",
					diagnostic.Message.Description)
			}
			if diagnostic.Message.Id != "useThisType" {
				t.Fatalf("message id: expected %q, got %q", "useThisType", diagnostic.Message.Id)
			}

			rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantFixed)+"\n")
		})
	}
}

// TestPreferReturnThisTypeRequiresTheTypedHarness pins that this rule cannot work without a checker,
// and that it declines rather than crashing when handed none.
//
// The nil guard at the top of the listener is unreachable from every other test in this file:
// RunTyped always supplies a live checker, so a mutant neutralizing that guard survives the whole
// fixture set. The plain harness is the only instrument that can see it. Without the guard this
// dereferences a nil checker; with it the rule is silent, which is the documented contract for a
// rule declaring NeedsTypeChecker.
//
// The case used here REPORTS under RunTyped and is asserted clean here, so a revert that dropped
// NeedsTypeChecker would fail rather than pass vacuously.
func TestPreferReturnThisTypeRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !PreferReturnThisType.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker: every judgment it makes is a type identity question")
	}

	sourceText := "class Foo {\n  f(): Foo {\n    return this;\n  }\n}\n"

	rule_testing.ExpectClean(t, rule_testing.Run(t, PreferReturnThisType,
		preferReturnThisTypeFile, sourceText))

	// The control: the same source under the typed harness reports, so the silence above is the
	// guard declining rather than the rule being unable to see this shape at all.
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferReturnThisType,
		preferReturnThisTypeFile, sourceText), "useThisType")
}
