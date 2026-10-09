package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const assertedOptionalChainFile = "/repository/source/Thing.ts"

// The cases below are the original's own test file plus a run of the real plugin at 8.67.0, which
// is the only answer key this rule has: the tree cohere gates contains zero non-null assertions, so
// a clean run over it cannot distinguish a working port from an inert one.
func TestNoNonNullAssertedOptionalChainFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an assertion ending a member chain", "declare const foo: any;\nexport const r = foo?.bar!;\n"},
		{"an assertion ending a call chain", "declare const foo: any;\nexport const r = foo?.bar()!;\n"},
		{"an assertion ending an element chain", "declare const foo: any;\nexport const r = foo?.[0]!;\n"},
		{"an assertion after an optional call", "declare const foo: any;\nexport const r = foo.bar?.()!;\n"},
		{"an assertion ending a longer chain", "declare const foo: any;\nexport const r = foo?.bar.baz!;\n"},
		// The parenthesized forms. Parens invert which node wraps which, so these reach the rule
		// through the second condition rather than the first. They are the half a single-condition
		// port loses, and they fail silently.
		{"an assertion wrapping a parenthesized chain", "declare const foo: any;\nexport const r = (foo?.bar)!;\n"},
		{"a parenthesized chain then a member", "declare const foo: any;\nexport const r = (foo?.bar)!.baz;\n"},
		{"a parenthesized chain then a call", "declare const foo: any;\nexport const r = (foo?.bar)!();\n"},
		{"a parenthesized assertion at chain end", "declare const foo: any;\nexport const r = (foo?.bar!);\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoNonNullAssertedOptionalChain, assertedOptionalChainFile, testCase.sourceText),
				"noNonNullOptionalChain")
		})
	}
}

func TestNoNonNullAssertedOptionalChainStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The two cases the whole rule turns on. Both contain an assertion directly after an
		// optional access, which is the shape a naive port reports, and both are valid because the
		// chain continues past the assertion so its short-circuit still reaches the result. These
		// are the fixtures that sit exactly where the rule decides.
		{"an assertion mid-chain before a member", "declare const foo: any;\nexport const r = foo?.bar!.baz;\n"},
		{"an assertion mid-chain before a call", "declare const foo: any;\nexport const r = foo?.bar!();\n"},
		{"an assertion mid-chain before an element access", "declare const foo: any;\nexport const r = foo?.bar![0];\n"},
		{"an assertion mid-chain, longer", "declare const foo: any;\nexport const r = foo?.bar!.baz.qux;\n"},
		// An assertion inside the chain's brackets rather than on the chain: the optional link does
		// not reach through a computed property's own expression.
		{"an assertion inside a computed property", "declare const foo: any;\ndeclare const key: any;\nexport const r = foo?.[key!];\n"},
		{"an assertion in an argument", "declare const foo: any;\ndeclare const x: any;\nexport const r = foo?.bar(x!);\n"},
		// An assertion with no chain at all, and a chain with no assertion: each half of the
		// condition present without the other.
		{"an assertion with no chain", "declare const foo: any;\nexport const r = foo!;\n"},
		{"an assertion then a plain member", "declare const foo: any;\nexport const r = foo!.bar;\n"},
		{"a doubled assertion with no chain", "declare const foo: any;\nexport const r = foo!!;\n"},
		{"a chain with no assertion", "declare const foo: any;\nexport const r = foo?.bar;\n"},
		{"a chain in a parenthesis with no assertion", "declare const foo: any;\nexport const r = (foo?.bar);\n"},
		// An assertion before the chain rather than after it: `foo!?.bar` asserts on `foo`, and the
		// `?.` that follows does its own check. Redundant, which is the other rule's business, and
		// not this rule's.
		{"an assertion before an optional access", "declare const foo: any;\nexport const r = foo!?.bar;\n"},
		{"a definite assignment", "export let value!: number;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoNonNullAssertedOptionalChain, assertedOptionalChainFile, testCase.sourceText))
		})
	}
}

// What the suggestion removes, which the fixture pair above cannot see.
//
// This rule offers a suggestion that deletes a range, and nothing asserted which range. A suggestion
// removing the wrong span is a rewrite the reader was never shown, and from a fixture that only
// checks which message fired it reads identically to a correct one.
//
// The range is one character wide, the `!` itself, so an off-by-one takes the token before it and
// leaves source that no longer parses. Asserted by applying the removal and comparing the resulting
// text rather than by comparing offsets, since an offset expectation is most likely to be wrong in
// the same direction as the code that produced it.
//
// Found by a review sweep rather than by a mutation, and that is worth recording. Every mutant I
// could construct against the range helper failed to compile, and a mutant that does not compile is
// neither a catch nor a survival, so this gap stood unmeasured rather than cleared until somebody
// read the test file and saw only ExpectFindings and ExpectClean in it.
func TestNoNonNullAssertedOptionalChainSuggestsRemovingTheOperator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{
			name:       "a property access",
			sourceText: "declare const foo: { bar?: string };\nexport const a = foo?.bar!;\n",
			wantSource: "declare const foo: { bar?: string };\nexport const a = foo?.bar;\n",
		},
		{
			name:       "a call",
			sourceText: "declare const foo: { bar?: () => string };\nexport const a = foo.bar?.()!;\n",
			wantSource: "declare const foo: { bar?: () => string };\nexport const a = foo.bar?.();\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoNonNullAssertedOptionalChain,
				assertedOptionalChainFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			suggestions := result.Diagnostics[0].Suggestions
			if len(suggestions) != 1 || len(suggestions[0].Fixes) != 1 {
				t.Fatalf("wanted one suggestion carrying one fix, got %d", len(suggestions))
			}

			fix := suggestions[0].Fixes[0]
			rewritten := testCase.sourceText[:fix.Range.Pos()] + fix.Text +
				testCase.sourceText[fix.Range.End():]
			if rewritten != testCase.wantSource {
				t.Fatalf("applying the suggestion gave %q, wanted %q", rewritten, testCase.wantSource)
			}
		})
	}
}
