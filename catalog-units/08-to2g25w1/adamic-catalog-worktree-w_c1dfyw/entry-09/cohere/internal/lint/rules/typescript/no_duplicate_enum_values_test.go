package typescript

import (
	"fmt"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const duplicateEnumValuesFile = "/repository/source/Values.ts"

// The cases below are oxc's own pass and fail vectors, emitted by the extractor rather than
// retyped, so no escape sequence passed through a hand. Every one of them is plain ASCII with
// no backslash escapes, verified byte against byte by the script that wrote this file.

func TestNoDuplicateEnumValuesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\n            enum E {\n              A,\n              B,\n            }\n        ",
		"\n            enum E {\n              A = 1,\n              B,\n            }\n        ",
		"\n            enum E {\n              A = 1,\n              B = 2,\n            }\n        ",
		"\n            enum E {\n              A = -1,\n              B = -2,\n            }\n        ",
		"\n            enum E {\n              A = +1,\n              B = +2,\n            }\n        ",
		"\n            enum E {\n              A = +1,\n              B = -1,\n            }\n        ",
		"\n            enum E {\n              A = 1,\n              B = -1,\n            }\n        ",
		"\n            enum E {\n              A = -0,\n              B = +0,\n            }\n        ",
		"\n            enum E {\n              A = -0,\n              B = 0,\n            }\n        ",
		"\n            enum E {\n              A = 1,\n              B = '1',\n            }\n        ",
		"\n            enum E {\n              A = -1,\n              B = '-1',\n            }\n        ",
		"\n            enum E {\n              A = 'A',\n              B = 'B',\n            }\n        ",
		"\n            enum E {\n              A = 'A',\n              B = 'B',\n              C,\n            }\n        ",
		"\n            enum E {\n              A = 'A',\n              B = 'B',\n              C = 2,\n              D = 1 + 1,\n            }\n        ",
		"\n            enum E {\n              A = 3,\n              B = 2,\n              C,\n            }\n        ",
		"\n            enum E {\n              A = 'A',\n              B = 'B',\n              C = 2,\n              D = foo(),\n            }\n        ",
		"\n            enum E {\n              A = '',\n              B = 0,\n            }\n        ",
		"\n            enum E {\n              A = 0,\n              B = -0,\n              C = NaN,\n            }\n        ",
		"\n            enum E {\n              A = NaN,\n              B = NaN,\n            }\n        ",
		"\n            enum E {\n              A = NaN,\n              B = -NaN,\n            }\n        ",
		"\n            enum E {\n              A = 'NaN',\n              B = NaN,\n            }\n        ",
		"\n            enum E {\n              A = -+-0,\n              B = +-+0,\n            }\n        ",
		"\n            enum E {\n              A = -'',\n              B = 0,\n            }\n        ",
		"\n            enum E {\n              A = Infinity,\n              B = Infinity,\n            }\n        ",
		"\n            const A = 'A';\n            enum E {\n              A = 'A',\n              B = `${A}`,\n            }\n        ",
	}
	for index, sourceText := range cases {
		t.Run(fmt.Sprintf("upstream pass %d", index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, sourceText))
		})
	}
}

func TestNoDuplicateEnumValuesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		findings   int
	}{
		{"\n            enum E {\n              A = 1,\n              B = 1,\n            }\n        ", 1},
		{"\n            enum E {\n              A = 0x10,\n              B = 16,\n            }\n        ", 1},
		{"\n            enum E {\n              A = 'A',\n              B = 'A',\n            }\n        ", 1},
		{"\n            enum E {\n              A = 'A',\n              B = 'A',\n              C = 1,\n              D = 1,\n            }\n        ", 2},
	}
	for index, testCase := range cases {
		t.Run(fmt.Sprintf("upstream fail %d", index), func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.findings)
			for i := range wantIds {
				wantIds[i] = "noDuplicateEnumValues"
			}
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, testCase.sourceText), wantIds...)
		})
	}
}

// The cases from here down are not upstream's. Each one pins a behavior measured against the release
// oxlint binary at ~/Projects/system/oxc/target/release/oxlint, with a firing control alongside every
// silent verdict, and each names the command's answer at the line. They exist because the imported
// corpus cannot separate the readings: its longest fail case has one repeat of one value, so every
// question below has at least two answers that agree on all four fail inputs.

// TestNoDuplicateEnumValuesReportsOncePerExtraCopy pins the counting reading.
//
// Three plausible readings exist for three copies of one value: one finding for the whole value, one
// per extra copy, or one per pair. They give 1, 2 and 3 for a triple, and the corpus contains no
// input with more than one repeat, so it votes on none of them. Measured: `enum E { A = 1, B = 1,
// C = 1 }` reports twice and adding `D = 1` makes it three, so it is one per extra copy.
func TestNoDuplicateEnumValuesReportsOncePerExtraCopy(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, "enum E { A = 1, B = 1, C = 1 }\n"),
		"noDuplicateEnumValues", "noDuplicateEnumValues")

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, "enum E { A = 1, B = 1, C = 1, D = 1 }\n"),
		"noDuplicateEnumValues", "noDuplicateEnumValues", "noDuplicateEnumValues")

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, "enum E { A = 'x', B = 'x', C = 'x' }\n"),
		"noDuplicateEnumValues", "noDuplicateEnumValues")
}

// TestNoDuplicateEnumValuesPointsAtTheEarlierInitializer pins where a finding lands.
//
// oxc renders one diagnostic with two labels and the earlier initializer is its primary position.
// Message-id assertions cannot see this, and pointing at the later member is the more natural design,
// so it needs its own assertion. Measured: the release binary prints `2:7` for this input, which is
// the `1` on line two.
func TestNoDuplicateEnumValuesPointsAtTheEarlierInitializer(t *testing.T) {
	t.Parallel()

	sourceText := "enum E {\n  A = 1,\n  B = 1,\n}\n"
	result := rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, sourceText)
	rule_testing.ExpectFindings(t, result, "noDuplicateEnumValues")

	reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "1" {
		t.Fatalf("expected the finding on the earlier initializer `1`, got %q", reported)
	}
	// The position, not just the text: both initializers are the byte `1`, so a text comparison
	// alone cannot tell the first from the second.
	if got := result.Diagnostics[0].Range.Pos(); got != 15 {
		t.Fatalf("expected the finding at offset 15 (line 2's initializer), got %d", got)
	}
}

// TestNoDuplicateEnumValuesAnchorsNumbersAtTheFirstAndStringsAtThePrevious pins upstream's asymmetry.
//
// This is the finding of this port and it is not guessable from either implementation's shape. oxc
// keeps numbers in a vector it appends to only when the value is new, so the recorded position never
// moves and a third copy reports against the *first*. It keeps strings in a map written on every hit,
// so the recorded position advances and a third copy reports against the *second*. Measured on the
// release binary over the same triple written both ways: the numeric enum reports twice at `2:7` and
// `2:7`, the string enum at `7:7` and then `8:7`.
//
// The imported corpus cannot see this at all, and neither can a fixture asserting only message ids.
func TestNoDuplicateEnumValuesAnchorsNumbersAtTheFirstAndStringsAtThePrevious(t *testing.T) {
	t.Parallel()

	numeric := "enum N {\n  A = 1,\n  B = 1,\n  C = 1,\n}\n"
	numericResult := rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, numeric)
	rule_testing.ExpectFindings(t, numericResult, "noDuplicateEnumValues", "noDuplicateEnumValues")
	// Both findings anchor on `A`'s initializer. Line two starts at 9 and its `1` sits at 15.
	if first, second := numericResult.Diagnostics[0].Range.Pos(), numericResult.Diagnostics[1].Range.Pos(); first != 15 || second != 15 {
		t.Fatalf("expected both numeric findings anchored at offset 15, got %d and %d", first, second)
	}

	stringly := "enum S {\n  A = 'x',\n  B = 'x',\n  C = 'x',\n}\n"
	stringResult := rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, stringly)
	rule_testing.ExpectFindings(t, stringResult, "noDuplicateEnumValues", "noDuplicateEnumValues")
	// The anchor advances: `B` reports against `A`, then `C` reports against `B`.
	first, second := stringResult.Diagnostics[0].Range.Pos(), stringResult.Diagnostics[1].Range.Pos()
	if first != 15 {
		t.Fatalf("expected the first string finding anchored on A's initializer at 15, got %d", first)
	}
	if second != 26 {
		t.Fatalf("expected the second string finding anchored on B's initializer at 26, got %d", second)
	}
}

// TestNoDuplicateEnumValuesComparesTheParsedValueNotTheSpelling pins numeric normalization.
//
// oxc compares `f64`, so every spelling of one double collides. This port compares the parser's
// canonical rendering of that double instead, which is a bijection onto it, and these are the
// spellings that would separate the two approaches if it were not. `0x10` against `16` is upstream's
// own fail case; the rest were measured on the release binary and all report.
func TestNoDuplicateEnumValuesComparesTheParsedValueNotTheSpelling(t *testing.T) {
	t.Parallel()

	cases := []string{
		"enum E { A = 1, B = 0x1 }\n",
		"enum E { A = 1, B = 1.0 }\n",
		"enum E { A = 1, B = 1e0 }\n",
		"enum E { A = 1, B = 0b1 }\n",
		"enum E { A = 1, B = 0o1 }\n",
		// A numeric separator is not part of the value. `1_0` is ten, so it collides with `10`.
		"enum E { A = 10, B = 1_0 }\n",
	}
	for index, sourceText := range cases {
		t.Run(fmt.Sprintf("spelling %d", index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, sourceText),
				"noDuplicateEnumValues")
		})
	}

	// The control for the group above: `1_0` is ten rather than one, so it must not collide with
	// `1`. Without this, a port that stripped separators textually would pass every case above.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile,
		"enum E { A = 1, B = 1_0 }\n"))
}

// TestNoDuplicateEnumValuesComparesTheCookedStringValue pins that escapes are decoded before matching.
//
// oxc compares `s.value`, the cooked text. Measured: `enum E { A = "aAb", B = 'aAb' }` reports
// on the release binary, and its message renders the cooked `'aAb'`. A port comparing raw source
// would be silent here. The escape is built rather than typed, so nothing on the path to this file
// could cook it early.
func TestNoDuplicateEnumValuesComparesTheCookedStringValue(t *testing.T) {
	t.Parallel()

	sourceText := "enum E { A = " + "\"a\\u0041b\"" + ", B = 'aAb' }\n"
	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, sourceText),
		"noDuplicateEnumValues")

	// The quote style is not part of the value either.
	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, "enum E { A = 'x', B = \"x\" }\n"),
		"noDuplicateEnumValues")
}

// TestNoDuplicateEnumValuesKeepsNumbersAndStringsApart pins the two-table decision.
//
// `enum E { A = 1, B = '1' }` is an upstream passing case, so one table keyed by text would fail the
// corpus. This restates it directly because the two-table structure is the thing being guarded and
// the corpus case reads as being about coercion.
func TestNoDuplicateEnumValuesKeepsNumbersAndStringsApart(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile,
		"enum E { A = 1, B = '1', C = 0, D = '0' }\n"))
}

// TestNoDuplicateEnumValuesDeclinesTheShapesUpstreamDoesNotMatch pins the reproduced gaps.
//
// Every case here is silent upstream and silent by the same mechanism: oxc matches only the
// `NumericLiteral` and `StringLiteral` variants, so anything wrapping a literal is a different node
// and falls through. All five were measured on the release binary with a firing control in the same
// run, because a silent probe and a broken probe look identical.
func TestNoDuplicateEnumValuesDeclinesTheShapesUpstreamDoesNotMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// A parenthesis makes the initializer a KindParenthesizedExpression. This is the case a
		// free-looking `SkipParentheses` would have got wrong, and the corpus writes no parens.
		{"a parenthesized number", "enum E { A = 1, B = (1) }\n"},
		{"a parenthesized string", "enum E { A = 'x', B = ('x') }\n"},
		// A sign is a prefix unary applied to the literal. Upstream ships this input commented out
		// in its *fail* vector, which is upstream saying it knows and has not fixed it.
		{"a negated number", "enum E { A = -1, B = -1 }\n"},
		// A substitution-free template is a separate node kind. Upstream marks this with a literal
		// TODO above two commented-out fail cases; @typescript-eslint resolves it and oxc does not.
		{"a template literal", "enum E { A = 'A', B = `A` }\n"},
		{"two template literals", "enum E { A = `A`, B = `A` }\n"},
		// A BigInt is KindBigIntLiteral rather than KindNumericLiteral.
		{"a bigint", "enum E { A = 1n, B = 1n }\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, testCase.sourceText))
		})
	}
}

// TestNoDuplicateEnumValuesScopesToOneEnum pins that the tables do not leak across declarations.
//
// A rule gathering into one map for the file would report the second enum's first member. Measured:
// the release binary reports twice on this input, once inside each enum, rather than three times.
func TestNoDuplicateEnumValuesScopesToOneEnum(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile,
			"enum E { A = 1, B = 1 }\nenum F { C = 1, D = 1 }\n"),
		"noDuplicateEnumValues", "noDuplicateEnumValues")

	// Two enums sharing a value with no repeat inside either is clean.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile,
		"enum E { A = 1 }\nenum F { B = 1 }\n"))
}

// TestNoDuplicateEnumValuesReachesConstAndDeclareEnums pins the declaration forms.
//
// Both are `KindEnumDeclaration` with a modifier rather than distinct kinds, so this costs the rule
// nothing, but a listener anchored on something narrower would miss them. Measured: the release
// binary reports on both.
func TestNoDuplicateEnumValuesReachesConstAndDeclareEnums(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, "const enum E { A = 1, B = 1 }\n"),
		"noDuplicateEnumValues")

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, "declare enum E { A = 1, B = 1 }\n"),
		"noDuplicateEnumValues")
}

// TestNoDuplicateEnumValuesReadsComputedAndQuotedMemberNames pins that an unusual member name is not
// a reason to decline.
//
// oxc reads the name through `id.static_name()`, which resolves a computed key holding a literal, and
// the name is only used for the help text so it never gates the finding. Measured on the release
// binary: `enum E { "A" = 1, ["B"] = 1 }` reports and its help says `Give B a unique value`. This
// also guards against reading the name with `Node.Text()`, which panics on a ComputedPropertyName.
func TestNoDuplicateEnumValuesReadsComputedAndQuotedMemberNames(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, "enum E { \"A\" = 1, [\"B\"] = 1 }\n"),
		"noDuplicateEnumValues")
}

// TestNoDuplicateEnumValuesMessageSaysWhyRatherThanRestatingTheName guards the message text.
//
// The literal below is typed here rather than referenced from the rule's own constant. Comparing
// against `messageNoDuplicateEnumValues` would be equality that looks correct and sees nothing: both
// sides move together under mutation, so a mutant rewriting the message survives.
func TestNoDuplicateEnumValuesMessageSaysWhyRatherThanRestatingTheName(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoDuplicateEnumValues, duplicateEnumValuesFile, "enum E { A = 1, B = 1 }\n")
	rule_testing.ExpectFindings(t, result, "noDuplicateEnumValues")

	if got := result.Diagnostics[0].Message.Id; got != "noDuplicateEnumValues" {
		t.Fatalf("unexpected message id %q", got)
	}
	want := "Two members of this enum are initialized to the same value. TypeScript permits it, but " +
		"a reader expects members of one enum to name distinct things, and a reverse lookup by " +
		"value can only return one of them, so the later member silently wins. The usual cause is " +
		"a copied line whose value was never changed. Give each member its own value."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Fatalf("message description changed:\n got: %q\nwant: %q", got, want)
	}
}
