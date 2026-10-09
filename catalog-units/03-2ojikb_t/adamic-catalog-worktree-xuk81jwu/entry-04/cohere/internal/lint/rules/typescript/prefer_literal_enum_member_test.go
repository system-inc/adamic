package typescript

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// preferLiteralEnumMemberFile names the fixture file. The rule reads no path and gates on no
// extension: upstream registers a bare TSEnumMember visitor with no source-type test.
const preferLiteralEnumMemberFile = "/repository/source/Flags.ts"

// preferLiteralEnumMemberCaseName numbers a row so a failure names which one.
func preferLiteralEnumMemberCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// decodePreferLiteralEnumMemberOptionsForTest routes a fixture through the rule's own decoder.
//
// Building the options struct directly would leave the decoder untested, and the decoder is where
// the wire shape lives: cohere strips ESLint's [severity, options] tuple, so what arrives here is
// the bare object rather than upstream's one-element array.
func decodePreferLiteralEnumMemberOptionsForTest(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodePreferLiteralEnumMemberOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return decoded
}

// TestPreferLiteralEnumMemberStaysSilentOnUpstreamPassCases is the imported clean corpus at the
// rule's default settings, verbatim.
//
// Ten of upstream's eighteen passing inputs, extracted from the clone's test file by parsing it with
// the TypeScript compiler rather than by reading it. Every one was additionally run through the
// installed 8.x build, which reported nothing on all ten.
//
// Two of these are the shapes our parser answers differently from upstream's, so they are the cases
// that pay for themselves here: a regular expression initializer is one node upstream calls a
// Literal, and a bare backtick string is a NoSubstitutionTemplateLiteral rather than a
// TemplateLiteral with an empty expression list.
func TestPreferLiteralEnumMemberStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\nenum ValidRegex {\n  A = /test/,\n}\n    ",
		"\nenum ValidString {\n  A = 'test',\n}\n    ",
		"\nenum ValidLiteral {\n  A = `test`,\n}\n    ",
		"\nenum ValidNumber {\n  A = 42,\n}\n    ",
		"\nenum ValidNumber {\n  A = -42,\n}\n    ",
		"\nenum ValidNumber {\n  A = +42,\n}\n    ",
		"\nenum ValidNull {\n  A = null,\n}\n    ",
		"\nenum ValidPlain {\n  A,\n}\n    ",
		"\nenum ValidQuotedKey {\n  'a',\n}\n    ",
		"\nenum ValidQuotedKeyWithAssignment {\n  'a' = 1,\n}\n    ",
	}
	for index, sourceText := range cases {
		t.Run(preferLiteralEnumMemberCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, PreferLiteralEnumMember,
				preferLiteralEnumMemberFile, sourceText))
		})
	}
}

// TestPreferLiteralEnumMemberStaysSilentWhenBitwiseExpressionsAreAllowed is the imported clean
// corpus under allowBitwiseExpressions, verbatim.
//
// The options go through the rule's own decoder rather than being built as a struct, because this
// rule's only key defaults to FALSE and reads TRUE here, so a decoder that dropped the key would
// leave every one of these eight cases reporting and the failure would name the rule rather than the
// decoder.
//
// The last three carry the shapes that separate this port from upstream's. Upstream's parser folds a
// parenthesis away entirely, so `Foo.A | (Foo.B & ~Foo.C)` reaches its rule as a bare binary
// expression; ours produces a real KindParenthesizedExpression, and without the unwrap in the rule
// this case reports and upstream does not.
func TestPreferLiteralEnumMemberStaysSilentWhenBitwiseExpressionsAreAllowed(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\nenum Foo {\n  A = 1 << 0,\n  B = 1 >> 0,\n  C = 1 >>> 0,\n  D = 1 | 0,\n  E = 1 & 0,\n  F = 1 ^ 0,\n  G = ~1,\n}\n      ",
		"\nenum Foo {\n  A = 1 << 0,\n  B = 1 >> 0,\n  C = A | B,\n}\n      ",
		"\nenum Foo {\n  A = 1 << 0,\n  B = 1 >> 0,\n  C = Foo.A | Foo.B,\n}\n      ",
		"\nenum Foo {\n  A = 1 << 0,\n  B = 1 >> 0,\n  C = Foo['A'] | B,\n}\n      ",
		"\nenum Foo {\n  A = 1 << 0,\n  B = 1 << 1,\n  C = 1 << 2,\n  D = A | B | C,\n}\n      ",
		"\nenum Foo {\n  A = 1 << 0,\n  B = 1 << 1,\n  C = 1 << 2,\n  D = Foo.A | Foo.B | Foo.C,\n}\n      ",
		"\nenum Foo {\n  A = 1 << 0,\n  B = 1 << 1,\n  C = 1 << 2,\n  D = Foo.A | (Foo.B & ~Foo.C),\n}\n      ",
		"\nenum Foo {\n  A = 1 << 0,\n  B = 1 << 1,\n  C = 1 << 2,\n  D = Foo.A | -Foo.B,\n}\n      ",
	}
	for index, sourceText := range cases {
		t.Run(preferLiteralEnumMemberCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, PreferLiteralEnumMember,
				preferLiteralEnumMemberFile, sourceText, decodePreferLiteralEnumMemberOptionsForTest(t, `{"allowBitwiseExpressions": true}`)))
		})
	}
}

// TestPreferLiteralEnumMemberFiresOnUpstreamFailCases is the imported failing corpus at the rule's
// default settings, verbatim, with the span of every finding asserted.
//
// The span matters more here than the count. Upstream reports on `node.id`, the member's NAME, while
// the thing being judged is its initializer, so a port that anchored on the member or on the
// initializer would satisfy every message id in this table and point somewhere upstream never
// points. Each expected span below is sliced out of upstream's own line and column numbers rather
// than typed, so the assertion cannot drift from what the corpus recorded.
func TestPreferLiteralEnumMemberFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			sourceText: "\nenum InvalidObject {\n  A = {},\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"A"},
		},
		{
			sourceText: "\nenum InvalidArray {\n  A = [],\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"A"},
		},
		{
			sourceText: "\nenum InvalidTemplateLiteral {\n  A = `foo ${0}`,\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"A"},
		},
		{
			sourceText: "\nenum InvalidConstructor {\n  A = new Set(),\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"A"},
		},
		{
			sourceText: "\nenum InvalidExpression {\n  A = 2 + 2,\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"A"},
		},
		{
			sourceText: "\nenum InvalidExpression {\n  A = delete 2,\n  B = -a,\n  C = void 2,\n  D = ~2,\n  E = !0,\n}\n      ",
			wantIds:    []string{"notLiteral", "notLiteral", "notLiteral", "notLiteral", "notLiteral"},
			wantSpans:  []string{"A", "B", "C", "D", "E"},
		},
		{
			sourceText: "\nconst variable = 'Test';\nenum InvalidVariable {\n  A = 'TestStr',\n  B = 2,\n  C,\n  V = variable,\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"V"},
		},
		{
			sourceText: "\nenum InvalidEnumMember {\n  A = 'TestStr',\n  B = A,\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"B"},
		},
		{
			sourceText: "\nconst Valid = { A: 2 };\nenum InvalidObjectMember {\n  A = 'TestStr',\n  B = Valid.A,\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"B"},
		},
		{
			sourceText: "\nenum Valid {\n  A,\n}\nenum InvalidEnumMember {\n  A = 'TestStr',\n  B = Valid.A,\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"B"},
		},
		{
			sourceText: "\nconst obj = { a: 1 };\nenum InvalidSpread {\n  A = 'TestStr',\n  B = { ...a },\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"B"},
		},
		{
			sourceText: "\nenum Foo {\n  A,\n  B = +A,\n}\n      ",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"B"},
		},
	}
	for index, testCase := range cases {
		t.Run(preferLiteralEnumMemberCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferLiteralEnumMember,
				preferLiteralEnumMemberFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q", findingIndex, gotSpan, wantSpan)
				}
			}
		})
	}
}

// TestPreferLiteralEnumMemberFiresWhenBitwiseExpressionsAreAllowed is the imported failing corpus
// under allowBitwiseExpressions, verbatim, with spans.
//
// The option does not merely widen what passes, it changes which message the rule reports, so these
// rows assert the notLiteralOrBitwiseExpression id that no default-settings case can reach. The
// second row is the one that pins the "part of a bitwise computation" carve-out: an enum member may
// name a sibling member only inside a bitwise expression, so `x >> Foo.A` reports for `x` while
// `Foo.A | Foo.B` two rows above does not.
func TestPreferLiteralEnumMemberFiresWhenBitwiseExpressionsAreAllowed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			sourceText: "\nconst x = 1;\nenum Foo {\n  A = x << 0,\n  B = x >> 0,\n  C = x >>> 0,\n  D = x | 0,\n  E = x & 0,\n  F = x ^ 0,\n  G = ~x,\n}\n      ",
			wantIds:    []string{"notLiteralOrBitwiseExpression", "notLiteralOrBitwiseExpression", "notLiteralOrBitwiseExpression", "notLiteralOrBitwiseExpression", "notLiteralOrBitwiseExpression", "notLiteralOrBitwiseExpression", "notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"A", "B", "C", "D", "E", "F", "G"},
		},
		{
			sourceText: "\nconst x = 1;\nenum Foo {\n  A = 1 << 0,\n  B = x >> Foo.A,\n  C = x >> A,\n}\n      ",
			wantIds:    []string{"notLiteralOrBitwiseExpression", "notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"B", "C"},
		},
	}
	for index, testCase := range cases {
		t.Run(preferLiteralEnumMemberCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, PreferLiteralEnumMember,
				preferLiteralEnumMemberFile, testCase.sourceText, decodePreferLiteralEnumMemberOptionsForTest(t, `{"allowBitwiseExpressions": true}`))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q", findingIndex, gotSpan, wantSpan)
				}
			}
		})
	}
}

// TestPreferLiteralEnumMemberFiresWhenBitwiseExpressionsAreExplicitlyDisallowed is upstream's one
// case written with the option spelled out as false.
//
// It is kept separate from the default-settings table on purpose. An explicit false and an absent
// key are the same verdict for this rule and they arrive at the decoder as different inputs, so this
// row is what proves the decoder does not treat a written false as a missing key and fall back to a
// default that happened to agree.
func TestPreferLiteralEnumMemberFiresWhenBitwiseExpressionsAreExplicitlyDisallowed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			sourceText: "\nenum Foo {\n  A = 1 << 0,\n  B = 1 >> 0,\n  C = 1 >>> 0,\n  D = 1 | 0,\n  E = 1 & 0,\n  F = 1 ^ 0,\n  G = ~1,\n}\n      ",
			wantIds:    []string{"notLiteral", "notLiteral", "notLiteral", "notLiteral", "notLiteral", "notLiteral", "notLiteral"},
			wantSpans:  []string{"A", "B", "C", "D", "E", "F", "G"},
		},
	}
	for index, testCase := range cases {
		t.Run(preferLiteralEnumMemberCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, PreferLiteralEnumMember,
				preferLiteralEnumMemberFile, testCase.sourceText, decodePreferLiteralEnumMemberOptionsForTest(t, `{"allowBitwiseExpressions": false}`))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q", findingIndex, gotSpan, wantSpan)
				}
			}
		})
	}
}

// TestPreferLiteralEnumMemberOnShapesUpstreamsCorpusDoesNotWrite covers the inputs where our tree
// and upstream's differ about the parse rather than about the rule.
//
// Every expected verdict below was measured by running the installed 8.x build on that exact
// source, alongside a control input that reported, so a silent row is a verdict rather than a rule
// that never ran. None of these appear in upstream's test file: its parser cannot produce a
// parenthesis node at all, so no case it writes could have found the unwrap missing.
func TestPreferLiteralEnumMemberOnShapesUpstreamsCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		options    string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "paren-one",
			why:        "a parenthesis is a node here and is folded away upstream, so without the unwrap this reports and upstream does not",
			sourceText: "enum Foo {\n  A = (1),\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "paren-nested",
			why:        "the unwrap has to loop, because a parenthesis nests",
			sourceText: "enum Foo {\n  A = ((2)),\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "paren-string",
			why:        "the unwrap runs before the literal arm rather than only before the bitwise one",
			sourceText: "enum Foo {\n  A = ('x'),\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "paren-bad-inside",
			why:        "the unwrap is transparent in both directions, so a bad initializer inside parentheses still reports",
			sourceText: "enum Foo {\n  A = (2 + 2),\n}\n",
			options:    "",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"A"},
		},
		{
			name:       "quoted-key-ident",
			why:        "a quoted member key is reachable by a bare identifier, because the comparison is on the cooked text",
			sourceText: "enum Foo {\n  'A' = 1 << 0,\n  B = A | 1,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "template-computed",
			why:        "a computed access whose key is a backtick string resolves the same way as a quoted one",
			sourceText: "enum Foo {\n  A = 1 << 0,\n  B = 1 << 1,\n  C = Foo[`A`] | Foo.B,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "minus-self-opt",
			why:        "the plus and minus arm passes the bitwise flag THROUGH rather than setting it, so a sibling under a minus is still not allowed",
			sourceText: "enum Foo {\n  A = 1 << 0,\n  B = -A,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{"notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"B"},
		},
		{
			name:       "tilde-self-opt",
			why:        "the tilde arm SETS the flag, which is what makes this the opposite verdict from the row above it",
			sourceText: "enum Foo {\n  A = 1 << 0,\n  B = ~A,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "bigint",
			why:        "a bigint is one of upstream's Literal shapes and has its own kind here",
			sourceText: "enum Foo {\n  A = 1n,\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "boolean",
			why:        "true and false are Literal upstream and are two separate keyword kinds here",
			sourceText: "enum Foo {\n  A = true,\n  B = false,\n}\n",
			options:    "",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "paren-bitwise-nested-opt",
			why:        "the unwrap runs on the recursion rather than only at the top, which is what a parenthesized operand needs",
			sourceText: "enum Foo {\n  A = 1 << 0,\n  B = ((A) | 1),\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "other-enum-name",
			why:        "the member access receiver has to be THIS enum's name, so a sibling enum's member is not exempt",
			sourceText: "enum Bar {\n  A = 1,\n}\nenum Foo {\n  B = 1 << 0,\n  C = Bar.A | B,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{"notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"C"},
		},
		{
			name:       "computed-nonstatic-key",
			why:        "a computed key that is not a static string declines, matching what a null answer does upstream",
			sourceText: "enum Foo {\n  A = 1 << 0,\n  B = Foo[someVar] | 1,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{"notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"B"},
		},
		{
			name:       "foreign-receiver-colliding-name",
			why:        "the receiver-name check is the ONLY thing declining here, because the property name does collide with a member of this enum; the weaker version of this case above declines on the name instead and cannot see the check at all",
			sourceText: "enum Bar {\n  A = 1,\n}\nenum Foo {\n  A = 1 << 0,\n  B = Bar.A | 1,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{"notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"B"},
		},
		{
			name:       "foreign-receiver-computed-colliding",
			why:        "the same distinguishing input through the computed-access branch, which carries its own copy of the receiver check",
			sourceText: "enum Bar {\n  A = 1,\n}\nenum Foo {\n  A = 1 << 0,\n  B = Bar['A'] | 1,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{"notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"B"},
		},
		{
			name:       "object-receiver-colliding",
			why:        "a plain object rather than an enum, so the receiver is not even an enum name",
			sourceText: "const Baz = { A: 1 };\nenum Foo {\n  A = 1 << 0,\n  B = Baz.A | 1,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{"notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"B"},
		},
		{
			name:       "self-receiver-control",
			why:        "the control: this enum's own name in the same shape, which must stay silent or the three rows above would be measuring nothing",
			sourceText: "enum Bar {\n  A = 1,\n}\nenum Foo {\n  A = 1 << 0,\n  B = Foo.A | 1,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "nonbitwise-binary-under-option",
			why:        "a binary operator that is not bitwise still reports with the option on; upstream's corpus writes only bitwise operators under this option, so nothing in it can tell the operator set from a set that accepts everything",
			sourceText: "enum Foo {\n  A = 1 + 1,\n  B = 1 * 2,\n  C = 1 - 1,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{"notLiteralOrBitwiseExpression", "notLiteralOrBitwiseExpression", "notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"A", "B", "C"},
		},
		{
			name:       "bad-right-operand",
			why:        "only the RIGHT operand is bad; upstream's corpus puts the bad operand on the left in every one of its rows, so a port that checked the left alone passes all of them",
			sourceText: "const x = 1;\nenum Foo {\n  A = 1 << x,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{"notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"A"},
		},
		{
			name:       "bad-left-operand",
			why:        "the mirror of the row above, kept beside it so the pair reads as one measurement rather than as a case whose side happens to matter",
			sourceText: "const x = 1;\nenum Foo {\n  A = x << 1,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{"notLiteralOrBitwiseExpression"},
			wantSpans:  []string{"A"},
		},
		{
			name:       "good-both-control",
			why:        "the control for the two rows above: both operands literal, so a rule that had stopped checking operands entirely would still be visible here",
			sourceText: "enum Foo {\n  A = 1 << 1,\n}\n",
			options:    "{\"allowBitwiseExpressions\": true}",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "control-fires",
			why:        "the control: without it every silent row above would be indistinguishable from a rule that never ran",
			sourceText: "enum Foo {\n  A = 2 + 2,\n}\n",
			options:    "",
			wantIds:    []string{"notLiteral"},
			wantSpans:  []string{"A"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var result rule_testing.Result
			if testCase.options == "" {
				result = rule_testing.Run(t, PreferLiteralEnumMember,
					preferLiteralEnumMemberFile, testCase.sourceText)
			} else {
				result = rule_testing.RunWithOptions(t, PreferLiteralEnumMember,
					preferLiteralEnumMemberFile, testCase.sourceText,
					decodePreferLiteralEnumMemberOptionsForTest(t, testCase.options))
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestPreferLiteralEnumMemberDecoderKeepsAnAbsentKeyDistinctFromAnExplicitFalse pins the two lines
// of the decoder that have no upstream counterpart.
//
// The wire field is a pointer so an absent key and a written false stay distinguishable, and the
// default is read from DefaultPreferLiteralEnumMemberSettings rather than from a zero value. Both
// happen to agree for this rule, whose only key defaults to false, and the test exists because that
// agreement is a coincidence: the next key added here may default to true, and nothing else in this
// file would notice a decoder that dropped it.
func TestPreferLiteralEnumMemberDecoderKeepsAnAbsentKeyDistinctFromAnExplicitFalse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw  string
		want bool
	}{
		{raw: `{}`, want: false},
		{raw: `{"allowBitwiseExpressions": false}`, want: false},
		{raw: `{"allowBitwiseExpressions": true}`, want: true},
	}
	for _, testCase := range cases {
		t.Run(testCase.raw, func(t *testing.T) {
			t.Parallel()
			decoded := decodePreferLiteralEnumMemberOptionsForTest(t, testCase.raw)
			settings, isSettings := decoded.(PreferLiteralEnumMemberOptions)
			if !isSettings {
				t.Fatalf("the decoder returned %T rather than the rule's own options type", decoded)
			}
			if settings.AllowBitwiseExpressions != testCase.want {
				t.Errorf("allowBitwiseExpressions decoded to %v, wanted %v",
					settings.AllowBitwiseExpressions, testCase.want)
			}
		})
	}
}

// TestPreferLiteralEnumMemberFallsBackWhenHandedNilOptions covers the path a rule configured as a
// bare severity string takes.
//
// A rule named as "error" with no object is handed nil, which the type assertion in Run cannot
// satisfy, so the fallback is the only thing standing between that configuration and a zero value.
// Every other fixture in this file reaches the rule through the decoder and none of them can see
// this line.
func TestPreferLiteralEnumMemberFallsBackWhenHandedNilOptions(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, PreferLiteralEnumMember,
		preferLiteralEnumMemberFile, "enum Foo {\n  A = 1 << 0,\n}\n", nil)
	rule_testing.ExpectFindings(t, result, "notLiteral")
}

// TestPreferLiteralEnumMemberRendersUpstreamsMessageText asserts what a reader is actually told.
//
// rule.Message is {Id, Description} with no interpolation, so there is nothing to render and nothing
// a format string could get wrong. What there IS to get wrong is the text itself, and every other
// fixture in this file goes through ExpectFindings, which compares message ids and count and nothing
// else. A mutation rewriting both descriptions survived the entire suite until this test existed.
//
// The wanted strings are typed here as literals rather than read from the rule's own constants. A
// comparison against the constant is equality, it looks correct, and both sides move together under
// mutation, which is the specific way this assertion fails to be an assertion.
func TestPreferLiteralEnumMemberRendersUpstreamsMessageText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		options     string
		wantId      string
		wantMessage string
	}{
		{
			name:        "default settings",
			options:     "",
			wantId:      "notLiteral",
			wantMessage: "Explicit enum value must only be a literal value (string or number).",
		},
		{
			name:    "allowBitwiseExpressions",
			options: `{"allowBitwiseExpressions": true}`,
			wantId:  "notLiteralOrBitwiseExpression",
			wantMessage: "Explicit enum value must only be a literal value (string or number) " +
				"or a bitwise expression.",
		},
	}
	const sourceText = "enum Foo {\n  A = 2 + 2,\n}\n"
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var result rule_testing.Result
			if testCase.options == "" {
				result = rule_testing.Run(t, PreferLiteralEnumMember,
					preferLiteralEnumMemberFile, sourceText)
			} else {
				result = rule_testing.RunWithOptions(t, PreferLiteralEnumMember,
					preferLiteralEnumMemberFile, sourceText,
					decodePreferLiteralEnumMemberOptionsForTest(t, testCase.options))
			}
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			reported := result.Diagnostics[0]
			if reported.Message.Id != testCase.wantId {
				t.Errorf("reported id %q, wanted %q", reported.Message.Id, testCase.wantId)
			}
			if reported.Message.Description != testCase.wantMessage {
				t.Errorf("reported message %q, wanted %q",
					reported.Message.Description, testCase.wantMessage)
			}
		})
	}
}
