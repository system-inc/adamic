package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const extraNonNullFile = "/repository/source/Thing.ts"

func TestNoExtraNonNullAssertionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		count      int
	}{
		// The three selectors of the original, one case each.
		{"an assertion on an assertion", "declare const foo: any;\nexport const r = foo!!;\n", 1},
		{"an assertion on an assertion, then a member", "declare const foo: any;\nexport const r = foo!!.bar;\n", 1},
		{"an assertion before an optional member access", "declare const foo: any;\nexport const r = foo!?.bar;\n", 1},
		{"an assertion before an optional call", "declare const foo: any;\nexport const r = foo!?.();\n", 1},
		{"an assertion before an optional element access", "declare const foo: any;\nexport const r = foo!?.[0];\n", 1},
		// Parenthesized forms. ESTree cannot see these as distinct from the unparenthesized ones,
		// so the original has no fixture for them and gets them right for free. We materialize a
		// node for the parens, so each of these is a real decision the walk-up has to make, and
		// each fails silently in the direction of reporting nothing.
		{"a parenthesized assertion inside an assertion", "declare const foo: any;\nexport const r = (foo!)!;\n", 1},
		{"a parenthesized assertion before an optional access", "declare const foo: any;\nexport const r = (foo!)?.bar;\n", 1},
		// Three assertions produce two findings: the inner two are each redundant against the one
		// outside them, and the outermost is redundant against nothing.
		{"three assertions", "declare const foo: any;\nexport const r = foo!!!;\n", 2},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoExtraNonNullAssertion, extraNonNullFile, testCase.sourceText)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "noExtraNonNullAssertion"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoExtraNonNullAssertionFixesWriteWhatTheyClaim asserts the rewritten source rather than the
// message id.
//
// This is the only rule of the three that carries a real fix rather than a suggestion, so it is the
// only one that rewrites a file unattended. An id assertion proves the least exactly where the blast
// radius is largest: a fix that deleted the wrong byte would satisfy every assertion above, because
// the finding it is attached to is identical either way.
func TestNoExtraNonNullAssertionFixesWriteWhatTheyClaim(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"an assertion on an assertion", "declare const foo: any;\nexport const r = foo!!;\n", "declare const foo: any;\nexport const r = foo!;\n"},
		{"an assertion on an assertion, then a member", "declare const foo: any;\nexport const r = foo!!.bar;\n", "declare const foo: any;\nexport const r = foo!.bar;\n"},
		{"an assertion before an optional member access", "declare const foo: any;\nexport const r = foo!?.bar;\n", "declare const foo: any;\nexport const r = foo?.bar;\n"},
		{"an assertion before an optional call", "declare const foo: any;\nexport const r = foo!?.();\n", "declare const foo: any;\nexport const r = foo?.();\n"},
		{"an assertion before an optional element access", "declare const foo: any;\nexport const r = foo!?.[0];\n", "declare const foo: any;\nexport const r = foo?.[0];\n"},
		// Three assertions produce two fixes at two different offsets. Applied front to back the
		// second would land in the wrong place, so this is also the case that pins the ordering.
		{"three assertions", "declare const foo: any;\nexport const r = foo!!!;\n", "declare const foo: any;\nexport const r = foo!;\n"},
		// The `!` is one ASCII byte, so subtracting one from the node's end cannot land inside a
		// character. This is the fixture that would catch it if that reasoning were wrong.
		{"multibyte text around the assertion", "declare const foo: any;\nexport const r = { '\u00e9\u00e8\u00ea': foo!! };\n", "declare const foo: any;\nexport const r = { '\u00e9\u00e8\u00ea': foo! };\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoExtraNonNullAssertion, extraNonNullFile, testCase.sourceText),
				testCase.wantSource)
		})
	}
}

func TestNoExtraNonNullAssertionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The rule's decision is "is this assertion redundant against another one", and the clean
		// cases have to sit where that decision is actually made rather than where no assertion
		// exists at all. Each of these contains a real assertion that is not redundant.
		{"a single assertion", "declare const foo: any;\nexport const r = foo!;\n"},
		{"a single assertion then a member", "declare const foo: any;\nexport const r = foo!.bar;\n"},
		{"a single assertion then a call", "declare const foo: any;\nexport const r = foo!();\n"},
		// The regression from upstream issue #2166, and the single most valuable fixture here.
		// The assertion sits on a link that is part of an optional chain but does not own a `?.`.
		// Reading the chain flag instead of the chain root reports this, which is wrong: the `.baz`
		// after the `!` is not optional, so nothing duplicates the assertion.
		{"an assertion mid-chain before a plain access", "declare const foo: any;\nexport const r = foo?.bar!.baz;\n"},
		{"an assertion mid-chain before a plain call", "declare const foo: any;\nexport const r = foo?.bar!();\n"},
		{"an assertion mid-chain, longer", "declare const checksCounter: any;\nexport const r = checksCounter?.textContent!.trim();\n"},
		// Upstream issue #2732. The assertion is on the computed property rather than on the
		// object, so the optional link does not reach through it and nothing is duplicated. A guard
		// that checked only "parent is an optional element access" without checking which position
		// the node occupies reports this.
		{"an assertion inside a computed property", "declare const foo: any;\ndeclare const key: any;\nexport const r = foo?.[key!];\n"},
		// An assertion on an argument rather than on the callee, for the same reason.
		{"an assertion in an argument of an optional call", "declare const foo: any;\ndeclare const x: any;\nexport const r = foo?.(x!);\n"},
		// An optional chain with no assertion at all, and an assertion with no chain: the two
		// halves of the condition, each present without the other.
		{"an optional chain alone", "declare const foo: any;\nexport const r = foo?.bar;\n"},
		{"a definite assignment", "export let value!: number;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoExtraNonNullAssertion, extraNonNullFile, testCase.sourceText))
		})
	}
}
