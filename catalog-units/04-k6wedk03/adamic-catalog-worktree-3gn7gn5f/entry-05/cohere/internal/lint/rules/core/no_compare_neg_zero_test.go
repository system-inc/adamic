package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestNoCompareNegZeroReportsComparisons is the fixture that must fire.
//
// Every comparison operator, on both sides, because the rule's whole claim is that the operator
// does not matter: all eight treat the two zeros alike, and a rule that checked only `===` would
// miss the relational cases that read as the most deliberate.
func TestNoCompareNegZeroReportsComparisons(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"x === -0;",
		"-0 === x;",
		"x == -0;",
		"-0 == x;",
		"x !== -0;",
		"-0 !== x;",
		"x != -0;",
		"-0 != x;",
		"x > -0;",
		"-0 > x;",
		"x >= -0;",
		"-0 >= x;",
		"x < -0;",
		"-0 < x;",
		"x <= -0;",
		"-0 <= x;",
	} {
		result := rule_testing.Run(t, NoCompareNegZero, "compare.ts", source)
		rule_testing.ExpectFindings(t, result, "unexpected")
	}
}

// TestNoCompareNegZeroSeesThroughParentheses pins the ESTree equivalence.
//
// ESLint's AST has no node for parentheses, so every one of these is the same expression to the
// original. Ours has the node, so the equivalence has to be written explicitly, and it has to hold
// at both levels: around the negation and around the literal.
func TestNoCompareNegZeroSeesThroughParentheses(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"x === (-0);",
		"((-0)) === x;",
		"x === -(0);",
		"x !== (((-(((0))))));",
		"x === (/* before */ - /* after */ (0));",
	} {
		result := rule_testing.Run(t, NoCompareNegZero, "parens.ts", source)
		rule_testing.ExpectFindings(t, result, "unexpected")
	}
}

// TestNoCompareNegZeroMatchesEveryZeroSpelling pins that the rule reads the value, not the text.
//
// `-0x0` is the same number as `-0` and carries the same defect, so a text comparison against "0"
// would be a hole exactly where an author reaching for a hex mask is most likely to fall in.
func TestNoCompareNegZeroMatchesEveryZeroSpelling(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"x === -0.0;",
		"x === -0e10;",
		"x === -0x0;",
		"x === -0b0;",
		"x === -0o0;",
		"x === -0_0;",
		"x === -00;",
		"x === -.0;",
	} {
		result := rule_testing.Run(t, NoCompareNegZero, "spellings.ts", source)
		rule_testing.ExpectFindings(t, result, "unexpected")
	}
}

// TestNoCompareNegZeroReportsOncePerComparison pins the count when both sides are negative zero.
//
// The count is the thing a reader trusts. Listening on the unary minus rather than the comparison
// would report `-0 === -0` twice for one defect, which is the failure mode this test exists to
// hold shut.
func TestNoCompareNegZeroReportsOncePerComparison(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoCompareNegZero, "both.ts", "-0 === -0;")
	rule_testing.ExpectFindings(t, result, "unexpected")

	nested := rule_testing.Run(t, NoCompareNegZero, "nested.ts", "x === -0 === -0;")
	rule_testing.ExpectFindings(t, nested, "unexpected", "unexpected")
}

// TestNoCompareNegZeroStaysSilentOnCorrectCode is the half that catches a rule firing on code that
// is fine.
//
// `Object.is(x, -0)` is here because it is the repair the message names: a rule that flagged the
// correct form would be worse than no rule. `-0n` is here because BigInt has no negative zero, and
// `x ** -0` because exponentiation is not a comparison however the operand is spelled.
func TestNoCompareNegZeroStaysSilentOnCorrectCode(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"x === 0;",
		"0 === x;",
		"x == 0;",
		"x === '0';",
		"'-0' === x;",
		"x == '-0';",
		"x === -1;",
		"-1 === x;",
		"x < 0;",
		"0 <= x;",
		"x > 0;",
		"0 >= x;",
		"x != 0;",
		"0 !== x;",
		"Object.is(x, -0);",
		"x === -0n;",
		"x === (-0 as number);",
		"x === (-0 satisfies number);",
		"x === -(-0);",
		"x === +(-0);",
		"x === (-0, y);",
		"x ** -0;",
		"x + -0;",
		"x = -0;",
		"x instanceof -0;",
	} {
		result := rule_testing.Run(t, NoCompareNegZero, "clean.ts", source)
		rule_testing.ExpectClean(t, result)
	}
}
