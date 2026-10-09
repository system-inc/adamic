package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const forDirectionFile = "/repository/source/Loop.ts"

func TestForDirectionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"less-than with a decrement", "export function run() { for(let i = 0; i < 10; i--) {} }\n"},
		{"less-than-or-equal with a decrement", "export function run() { for(let i = 0; i <= 10; i--) {} }\n"},
		{"greater-than with an increment", "export function run() { for(let i = 10; i > 0; i++) {} }\n"},
		{"greater-than-or-equal with an increment", "export function run() { for(let i = 10; i >= 0; i++) {} }\n"},
		{"a prefix decrement", "export function run() { for(let i = 0; i < 10; --i) {} }\n"},
		{"a prefix increment on a descending loop", "export function run() { for(let i = 10; i > 0; ++i) {} }\n"},

		// The counter on the right flips what counts as wrong, and a rule checking only the left
		// operand is silent on every loop written this way.
		{"the counter on the right of less-than", "export function run() { for(let i = 10; 10 < i; i++) {} }\n"},
		{"the counter on the right of greater-than", "export function run() { for(let i = 0; 0 > i; i--) {} }\n"},

		// Compound assignment with a statically known sign.
		{"minus-equals on an ascending loop", "export function run() { for(let i = 0; i < 10; i -= 1) {} }\n"},
		{"plus-equals on a descending loop", "export function run() { for(let i = 10; i > 0; i += 1) {} }\n"},
		{"plus-equals with a negated literal", "export function run() { for(let i = 0; i < 10; i += -1) {} }\n"},
		{"minus-equals with a negated literal", "export function run() { for(let i = 10; i > 0; i -= -1) {} }\n"},
		// Number(true) is 1, and ESLint accepts booleans as a static step. Nobody writes this, but
		// the fixture exists so the accepted type set is measured rather than asserted: dropping
		// the true arm otherwise changes nothing any test can see.
		{"minus-equals with true", "export function run() { for(let i = 0; i < 10; i -= true) {} }\n"},

		// A single modification wrapped in a sequence is still a single modification.
		{"a sequence with one modification", "export function run() { let x = 0; for(let i = 0; i < 10; x++, i--) {} }\n"},

		// A nonzero mantissa with an exponent does move the counter, so this must fire. Pairs with
		// the silent 0e5 case above: together they pin that the mantissa decides, not the exponent.
		{"minus-equals with an exponent literal", "export function run() { for(let i = 0; i < 10; i -= 1e2) {} }\n"},
		{"a parenthesized update", "export function run() { for(let i = 0; i < 10; (i--)) {} }\n"},
		{"a parenthesized condition", "export function run() { for(let i = 0; (i < 10); i--) {} }\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ForDirection, forDirectionFile, testCase.sourceText),
				"incorrectDirection")
		})
	}
}

func TestForDirectionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The shapes the rule is asking for.
		{"an ascending loop", "export function run() { for(let i = 0; i < 10; i++) {} }\n"},
		{"a descending loop", "export function run() { for(let i = 10; i > 0; i--) {} }\n"},
		{"plus-equals on an ascending loop", "export function run() { for(let i = 0; i < 10; i += 1) {} }\n"},
		{"minus-equals on a descending loop", "export function run() { for(let i = 10; i > 0; i -= 1) {} }\n"},
		{"the counter on the right, moving correctly", "export function run() { for(let i = 0; 10 > i; i++) {} }\n"},

		// Two modifications make the loop ambiguous rather than wrong, and ESLint deliberately
		// says nothing. This is the boundary a rule taking the first modification would cross.
		{"two modifications of the counter", "export function run() { for(let i = 0; i < 10; i++, i--) {} }\n"},
		{"two decrements on an ascending loop", "export function run() { for(let i = 0; i < 10; i--, i--) {} }\n"},

		// An unknown sign is silence. Reporting these would flag correct loops.
		{"plus-equals with a variable step", "export function run() { let step = -1; for(let i = 0; i < 10; i += step) {} }\n"},
		{"plus-equals with a call result", "export function run() { declare function step(): number; for(let i = 0; i < 10; i += step()) {} }\n"},
		{"plus-equals with zero", "export function run() { for(let i = 0; i < 10; i -= 0) {} }\n"},
		// Number(false) is 0, so this is no movement and stays silent. Pairs with the true case in
		// the firing table to pin both halves of the boolean arm.
		{"minus-equals with false", "export function run() { for(let i = 0; i < 10; i -= false) {} }\n"},
		{"plus-equals with a zero in another base", "export function run() { for(let i = 0; i < 10; i -= 0x0) {} }\n"},
		// A zero mantissa stays zero whatever the exponent, so this is still no movement and still
		// silence. Found by treating an exponent as nonzero and watching the suite stay green.
		{"plus-equals with a zero mantissa and an exponent", "export function run() { for(let i = 0; i < 10; i -= 0e5) {} }\n"},
		{"plus-equals with a zero decimal", "export function run() { for(let i = 0; i < 10; i -= 0.0) {} }\n"},

		// Assignment operators that carry no direction.
		{"a plain assignment", "export function run() { for(let i = 0; i < 10; i = i + 1) {} }\n"},
		{"a multiply-assign", "export function run() { for(let i = 1; i < 10; i *= 2) {} }\n"},

		// The update does not touch the counter at all, so the loop's progress is elsewhere.
		{"an update to a different variable", "export function run() { let x = 0; for(let i = 0; i < 10; x++) {} }\n"},

		// Conditions the rule has no opinion about.
		{"an equality condition", "export function run() { for(let i = 0; i !== 10; i--) {} }\n"},
		{"a non-binary condition", "export function run() { declare const going: boolean; for(let i = 0; going; i--) {} }\n"},
		{"neither operand an identifier", "export function run() { for(let i = 0; 0 < 10; i--) {} }\n"},

		// Loops with a missing clause cannot be judged, and must not crash.
		{"no condition", "export function run() { for(let i = 0; ; i--) {} }\n"},
		{"no update", "export function run() { for(let i = 0; i < 10; ) { i++; } }\n"},
		{"an empty for", "export function run() { for(;;) { break; } }\n"},

		{"a while loop is not this rule's business", "export function run() { let i = 0; while(i < 10) { i--; } }\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ForDirection, forDirectionFile, testCase.sourceText))
		})
	}
}

// Both operands are examined, and at most one report is emitted per loop.
//
// The early return that enforces this is unreachable, and it took three wrong arguments and a
// direct measurement to establish that, which is why it is written down rather than left as a
// comment saying "defensive".
//
// The reason both sides cannot qualify: for `a < b`, the wrong direction is -1 on the left and +1
// on the right. One update expression moves a given identifier one way. So a left match needs a
// decrement of `a` and a right match needs an increment of `b`, and while an update could contain
// both (`a--, b++`), that is a loop where the condition is `a < b` with `a` falling and `b`
// rising, which is a loop that terminates less often, not more. Reaching a genuine double report
// needs both operands to name the same identifier, and then the two required directions are
// opposite and one update cannot supply both.
//
// What settled it was probing each side alone rather than reasoning further. My first two
// arguments were confident and wrong, and both would have gone into a commit message as fact.
// The measurement cost one test and thirty seconds.
//
// The return is kept: it makes the single-report behavior local rather than a property a reader
// has to reconstruct, and one line is cheaper than this comment.
func TestForDirectionReportsOncePerLoop(t *testing.T) {
	t.Parallel()

	// The same counter on both sides. Qualifies on the left only, since the wrong direction
	// differs by side and one update cannot be both.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, ForDirection, forDirectionFile,
		"export function run() { for(let i = 0; i < i; i--) {} }\n"), "incorrectDirection")

	// Different counters, one modified. Qualifies on the left only.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, ForDirection, forDirectionFile,
		"export function run() { let j = 0; for(let i = 0; i < j; i--) {} }\n"), "incorrectDirection")

	// Both counters modified, still one finding. The right-hand operand does not qualify: for
	// `i < j` the wrong direction on the right is an increment, and `j--` is a decrement.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, ForDirection, forDirectionFile,
		"export function run() { let j = 0; for(let i = 0; i < j; i--, j--) {} }\n"), "incorrectDirection")

	// The right-hand operand in isolation, which is what pins the direction mapping rather than
	// leaving it inferred from the case above.
	rule_testing.ExpectClean(t, rule_testing.Run(t, ForDirection, forDirectionFile,
		"export function run() { let j = 0; for(let i = 0; i < j; j--) {} }\n"))

	// And the shape that does qualify on the right, so the mapping is pinned in both directions.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, ForDirection, forDirectionFile,
		"export function run() { let j = 0; for(let i = 0; i < j; j++) {} }\n"), "incorrectDirection")
}
