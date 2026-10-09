package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// selfCompareFile is where the fixtures pretend to live.
const selfCompareFile = "/repository/source/SelfCompare.ts"

// The corpus is upstream's, extracted from its tester rather than retyped.
//
// `/tmp/lint-sources/eslint/tests/lib/rules/no-self-compare.js` carries 6 valid cases and 15
// invalid ones, every invalid case naming exactly one `comparingToSelf`. Rendered to Go literals by
// a script that reads that file through a stub RuleTester, so nothing was retyped and no escape
// could be cooked on the way.
//
// The last two clean cases are the sharpest thing in the corpus and they are why this rule compares
// tokens rather than text: `this.#field` and `this['#field']` read almost the same characters and
// are different tokens, accessing different things. Both orderings are written, which is upstream
// pinning that the distinction is not an artifact of which side the private name sits on.
func TestNoSelfCompareStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"if (x === y) { }"},
		{"if (1 === 2) { }"},
		{"y=x*x"},
		{"foo.bar.baz === foo.bar.qux"},
		{"class C { #field; foo() { this.#field === this['#field']; } }"}, // lang={"ecmaVersion":2022}
		{"class C { #field; foo() { this['#field'] === this.#field; } }"}, // lang={"ecmaVersion":2022}
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoSelfCompare, selfCompareFile, testCase.sourceText))
		})
	}
}

// Every reporting case names one `comparingToSelf`, which is upstream's own per-input count.
//
// Two of these carry the whole discrimination. `foo.bar().baz.qux >= foo.bar ().baz .qux` differs in
// whitespace on both sides and reports, so a port comparing source text is silent on it. And
// `this.#field === this.#field` reports while the two clean private-name cases above do not, which
// is the pair that separates token equality from character equality.
func TestNoSelfCompareFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"if (x === x) { }"},      // errors=["comparingToSelf"]
		{"if (x !== x) { }"},      // errors=["comparingToSelf"]
		{"if (x > x) { }"},        // errors=["comparingToSelf"]
		{"if ('x' > 'x') { }"},    // errors=["comparingToSelf"]
		{"do {} while (x === x)"}, // errors=["comparingToSelf"]
		{"x === x"},               // errors=["comparingToSelf"]
		{"x !== x"},               // errors=["comparingToSelf"]
		{"x == x"},                // errors=["comparingToSelf"]
		{"x != x"},                // errors=["comparingToSelf"]
		{"x > x"},                 // errors=["comparingToSelf"]
		{"x < x"},                 // errors=["comparingToSelf"]
		{"x >= x"},                // errors=["comparingToSelf"]
		{"x <= x"},                // errors=["comparingToSelf"]
		{"foo.bar().baz.qux >= foo.bar ().baz .qux"},                   // errors=["comparingToSelf"]
		{"class C { #field; foo() { this.#field === this.#field; } }"}, // errors=["comparingToSelf"] lang={"ecmaVersion":2022}
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoSelfCompare, selfCompareFile, testCase.sourceText), "comparingToSelf")
		})
	}
}

// Operators upstream deliberately leaves out, which the imported corpus never writes.
//
// `in` and `instanceof` are comparisons in the grammar and are absent from upstream's operator set,
// because neither is a tautology when both sides are written the same: `x in x` asks whether the
// value of `x` has a property named by the value of `x`, and `x instanceof x` asks whether `x` is an
// instance of itself. A port widening the set to "the relational operators" reports both, and every
// imported fixture stays green while it does, because upstream never wrote them.
//
// The equality operators are the control: same shape, same identical operands, and they report.
func TestNoSelfCompareDeclinesInAndInstanceof(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"in", "declare const x: string; export const a = x in x;"},
		{"instanceof", "declare const x: any; export const a = x instanceof x;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoSelfCompare, selfCompareFile, testCase.sourceText))
		})
	}

	// The control, so the silence above is a measurement about the operator rather than about the
	// declare-heavy shape of the inputs.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoSelfCompare, selfCompareFile,
		"declare const x: any; export const a = x === x;"), "comparingToSelf")
}

// The reported span is the whole comparison, which no message-id fixture above can see.
//
// Upstream passes `node`, the BinaryExpression, rather than either operand or the operator token. A
// port anchoring on the operator satisfies every assertion above and points a reader at two
// characters in the middle of an expression.
func TestNoSelfCompareReportsTheWholeComparison(t *testing.T) {
	t.Parallel()

	const sourceText = "if (foo.bar >= foo.bar) { }"
	result := rule_testing.Run(t, NoSelfCompare, selfCompareFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "foo.bar >= foo.bar" {
		t.Fatalf("reported %q, wanted the whole comparison", reported)
	}
}
