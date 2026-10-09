package core

import (
	"encoding/json"
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noBitwiseFile is where the fixtures pretend to live.
const noBitwiseFile = "/repository/source/NoBitwise.ts"

// noBitwiseExpectation is one finding, with the span and the operator the message renders.
//
// The operator is asserted because it is the only part of the message that varies, and a port that
// reported the wrong spelling would pass every message-id assertion: all thirteen operators share
// one message id.
type noBitwiseExpectation struct {
	line      int
	column    int
	endLine   int
	endColumn int
	operator  string
}

type noBitwiseCase struct {
	name        string
	source      string
	optionsJson string
	findings    []noBitwiseExpectation
	reason      string
}

// decodeNoBitwiseOptionsForTest routes a case's options through the SHIPPED decoder.
//
// Not a hand-built options struct. A fixture constructing the options directly never exercises the
// decoder, so a decoder wrong about the shape the config layer delivers passes every fixture.
func decodeNoBitwiseOptionsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return nil
	}
	decoded, err := DecodeNoBitwiseOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", optionsJson, err)
	}
	return decoded
}

// noBitwiseOffsetOf converts a 1-based line and column into a byte offset.
//
// `rule_testing.Run` does not trim its input, so no rebasing is needed.
func noBitwiseOffsetOf(t *testing.T, source string, line int, column int) int {
	t.Helper()
	offset := 0
	for current := 1; current < line; current++ {
		next := strings.IndexByte(source[offset:], '\n')
		if next < 0 {
			t.Fatalf("the fixture has no line %d", line)
		}
		offset += next + 1
	}
	return offset + column - 1
}

func runNoBitwise(t *testing.T, testCase noBitwiseCase) rule_testing.Result {
	t.Helper()
	return rule_testing.RunWithOptions(t, NoBitwise, noBitwiseFile, testCase.source,
		decodeNoBitwiseOptionsForTest(t, testCase.optionsJson))
}

// TestNoBitwiseStaysSilent runs upstream's whole `valid` list.
func TestNoBitwiseStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range noBitwiseCleanCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoBitwise(t, testCase))
		})
	}
}

// TestNoBitwiseFires runs upstream's whole `invalid` list, asserting the span and the operator.
func TestNoBitwiseFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range noBitwiseReportingCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoBitwise(t, testCase)

			wantIds := make([]string, 0, len(testCase.findings))
			for range testCase.findings {
				wantIds = append(wantIds, messageNoBitwise.Id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, expected := range testCase.findings {
				diagnostic := result.Diagnostics[index]
				wantPos := noBitwiseOffsetOf(t, testCase.source, expected.line, expected.column)
				wantEnd := noBitwiseOffsetOf(t, testCase.source, expected.endLine, expected.endColumn)
				if diagnostic.Range.Pos() != wantPos || diagnostic.Range.End() != wantEnd {
					t.Errorf("finding %d: expected span [%d,%d), got [%d,%d) %q",
						index, wantPos, wantEnd, diagnostic.Range.Pos(), diagnostic.Range.End(),
						testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()])
				}
				want := "Unexpected use of '" + expected.operator + "'."
				if !strings.HasPrefix(diagnostic.Message.Description, want) {
					t.Errorf("finding %d: expected the message to name %q, got %q",
						index, expected.operator, diagnostic.Message.Description)
				}
			}
		})
	}
}

// TestNoBitwiseInt32HintEdges pins the exemption's boundaries, every row measured upstream.
//
// The corpus writes exactly two `int32Hint` cases, `a|0` on and `a|0` with allow, so nothing in it
// separates "the right operand is the literal zero" from any looser reading. These do.
func TestNoBitwiseInt32HintEdges(t *testing.T) {
	t.Parallel()
	const hintOn = `{"int32Hint":true}`
	cases := []noBitwiseCase{
		{name: "x|0 is the idiom", source: `x|0`, optionsJson: hintOn, reason: "the corpus case"},
		{
			name:   "x|0.0 is exempt, because the test is on the VALUE not the spelling",
			source: `x|0.0`, optionsJson: hintOn,
			reason: "measured exempt upstream. Our parser canonicalises the numeric text to \"0\", " +
				"which is the same equivalence class as upstream's `=== 0`.",
		},
		{name: "x|0x0 is exempt", source: `x|0x0`, optionsJson: hintOn, reason: "same canonicalisation"},
		{name: "x|0e0 is exempt", source: `x|0e0`, optionsJson: hintOn, reason: "same canonicalisation"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoBitwise(t, testCase)
			if len(result.Diagnostics) != 0 {
				t.Errorf("expected silence, got %v. %s", result.MessageIds(), testCase.reason)
			}
		})
	}

	reporting := []noBitwiseCase{
		{
			name:   "0|x reports, because the idiom is directional",
			source: `0|x`, optionsJson: hintOn,
			reason: "upstream tests node.right only, so a zero on the left is not the hint",
		},
		{name: "x|1 reports", source: `x|1`, optionsJson: hintOn, reason: "the value must be zero"},
		{
			name:   "x|'0' reports, because a string is not a numeric literal",
			source: `x|'0'`, optionsJson: hintOn,
			reason: "measured reporting upstream; the kind guard is what declines it here",
		},
		{
			name:   "x|0n reports, because a BigInt zero is a different literal",
			source: `x|0n`, optionsJson: hintOn,
			reason: "KindBigIntLiteral never reaches the numeric arm",
		},
		{
			name:   "x&0 reports, because the operator must be |",
			source: `x&0`, optionsJson: hintOn,
			reason: "the hint is specific to the | coercion",
		},
		{
			name:   "x|0 reports with the hint off",
			source: `x|0`, optionsJson: "",
			reason: "the option is what makes the difference, and this is its control",
		},
	}
	for _, testCase := range reporting {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoBitwise(t, testCase)
			if len(result.Diagnostics) == 0 {
				t.Errorf("expected a finding and got none. %s", testCase.reason)
			}
		})
	}
}

// TestNoBitwiseParenthesizedZeroDivergesFromUpstream records a divergence rather than hiding it.
//
// `x|(0)` is EXEMPT upstream, measured against the installed build, because its parser folds the
// parentheses away before `node.right.type === "Literal"` is tested. Ours keeps a
// `KindParenthesizedExpression`, so the kind guard declines and the finding is reported.
//
// Not corrected, and the direction is why. This arm is an EXEMPTION: unwrapping here would make the
// rule report less, on a shape upstream's corpus never writes and so never blessed. Reporting a
// bitwise `|` that a reader wrote with redundant parentheses is the safe side of the divergence, and
// the alternative is silently widening an exemption. The sibling `no-negated-condition` in this
// batch unwraps for the opposite reason: there the parens sit on a finding rather than on an
// exemption, and declining costs findings.
func TestNoBitwiseParenthesizedZeroDivergesFromUpstream(t *testing.T) {
	t.Parallel()
	result := runNoBitwise(t, noBitwiseCase{source: `x|(0)`, optionsJson: `{"int32Hint":true}`})
	rule_testing.ExpectFindings(t, result, messageNoBitwise.Id)
}

// TestNoBitwiseDecoderAcceptsTheShapesTheConfigLayerDelivers crosses the config boundary.
//
// Every fixture above reaches the decoder with bytes this test file built; the config layer builds
// different bytes. `meta.schema` declares ONE element here, so the cohere spelling and upstream's
// coincide, and that is the fact being pinned rather than assumed.
func TestNoBitwiseDecoderAcceptsTheShapesTheConfigLayerDelivers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		raw       string
		wantAllow []string
		wantHint  bool
		wantErr   bool
	}{
		{name: "absent options, which is what a bare \"error\" delivers"},
		{name: "the cohere spelling: one option object after the severity",
			raw: `{"allow":["~"],"int32Hint":true}`, wantAllow: []string{"~"}, wantHint: true},
		{name: "allow alone", raw: `{"allow":["|","&"]}`, wantAllow: []string{"|", "&"}},
		{name: "int32Hint alone", raw: `{"int32Hint":true}`, wantHint: true},
		{name: "an empty object exempts nothing", raw: `{}`},
		{name: "an explicitly empty allow list", raw: `{"allow":[]}`, wantAllow: []string{}},
		{
			// A decoder answering "no options" to a shape it did not understand produces a rule that
			// registers, reports, and enforces something other than what the config says.
			name: "a shape that is neither must error rather than decode to a default",
			raw:  `"int32Hint"`, wantErr: true,
		},
		{
			name: "upstream's variadic spelling is not one this layer delivers",
			raw:  `[{"int32Hint":true}]`, wantErr: true,
		},
		{
			// Upstream gets this refusal from its schema. There is no schema layer here, so a rule
			// that ignored an unknown spelling would leave a project believing it had exempted
			// something it had not.
			name: "an operator outside upstream's enum is refused, not ignored",
			raw:  `{"allow":[">>>>"]}`, wantErr: true,
		},
		{
			name: "a near miss is refused too, since `>>>` and `>>>=` are different operators",
			raw:  `{"allow":["&&"]}`, wantErr: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoBitwiseOptions(json.RawMessage(testCase.raw))
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected the decoder to refuse %s, got %#v", testCase.raw, decoded)
				}
				return
			}
			if err != nil {
				t.Fatalf("the decoder refused %s: %v", testCase.raw, err)
			}
			options, ok := decoded.(NoBitwiseOptions)
			if !ok {
				t.Fatalf("the decoder returned %T rather than NoBitwiseOptions", decoded)
			}
			if options.Int32Hint != testCase.wantHint {
				t.Errorf("int32Hint: expected %v, got %v", testCase.wantHint, options.Int32Hint)
			}
			if strings.Join(options.Allow, ",") != strings.Join(testCase.wantAllow, ",") {
				t.Errorf("allow: expected %v, got %v", testCase.wantAllow, options.Allow)
			}
		})
	}
}

// TestNoBitwiseAllowActuallyReachesTheRule proves the decoded list changes what is reported.
//
// A decoder that round-trips and a rule that ignores what it decoded look identical from the test
// above. One source, two configurations, opposite verdicts.
func TestNoBitwiseAllowActuallyReachesTheRule(t *testing.T) {
	t.Parallel()
	allowed := runNoBitwise(t, noBitwiseCase{source: `~a`, optionsJson: `{"allow":["~"]}`})
	if len(allowed.Diagnostics) != 0 {
		t.Errorf("with `~` allowed the rule must be silent, got %v", allowed.MessageIds())
	}

	// The control, and it is the load-bearing half: an allow list for a DIFFERENT operator must not
	// exempt this one, which separates "the list was read" from "the list was read correctly".
	other := runNoBitwise(t, noBitwiseCase{source: `~a`, optionsJson: `{"allow":["|"]}`})
	rule_testing.ExpectFindings(t, other, messageNoBitwise.Id)

	none := runNoBitwise(t, noBitwiseCase{source: `~a`})
	rule_testing.ExpectFindings(t, none, messageNoBitwise.Id)
}

// TestNoBitwiseLogicalOperatorsStayClean pins the omissions from upstream's list.
//
// The list is thirteen bitwise operators, and the four a reader most expects to see in it are
// absent: `&&`, `||`, `??` and `!`. The corpus asserts the assignment forms; the plain forms and the
// remaining arithmetic operators are stated here so a mutation adding a kind to the map fails.
func TestNoBitwiseLogicalOperatorsStayClean(t *testing.T) {
	t.Parallel()
	sources := []string{
		`a && b`, `a || b`, `a ?? b`, `!a`, `a + b`, `a - b`, `a * b`, `a / b`, `a % b`,
		`a ** b`, `a += b`, `a -= b`, `a &&= b`, `a ||= b`, `a ??= b`, `a === b`, `a < b`, `a > b`,
	}
	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			result := runNoBitwise(t, noBitwiseCase{source: source})
			if len(result.Diagnostics) != 0 {
				t.Errorf("expected silence, got %v: this operator is not bitwise", result.MessageIds())
			}
		})
	}
}

// noBitwiseCleanCases are upstream's `valid` list, verbatim.
var noBitwiseCleanCases = []noBitwiseCase{
	{name: "upstream valid[0]", source: `a + b`, optionsJson: ""},
	{name: "upstream valid[1]", source: `!a`, optionsJson: ""},
	{name: "upstream valid[2]", source: `a && b`, optionsJson: ""},
	{name: "upstream valid[3]", source: `a || b`, optionsJson: ""},
	{name: "upstream valid[4]", source: `a += b`, optionsJson: ""},
	{name: "upstream valid[5]", source: `a &&= b`, optionsJson: ""},
	{name: "upstream valid[6]", source: `a ||= b`, optionsJson: ""},
	{name: "upstream valid[7]", source: `a ??= b`, optionsJson: ""},
	{name: "upstream valid[8]", source: `~[1, 2, 3].indexOf(1)`, optionsJson: `{"allow": ["~"]}`},
	{name: "upstream valid[9]", source: `~1<<2 === -8`, optionsJson: `{"allow": ["~", "<<"]}`},
	{name: "upstream valid[10]", source: `a|0`, optionsJson: `{"int32Hint": true}`},
	{name: "upstream valid[11]", source: `a|0`, optionsJson: `{"allow": ["|"], "int32Hint": false}`},
}

// noBitwiseReportingCases are upstream's `invalid` list, verbatim.
//
// Upstream asserts only a count, so the spans and the rendered operator were measured by
// driving the installed 10.8.1 build.
var noBitwiseReportingCases = []noBitwiseCase{
	{
		name:        "upstream invalid[0]",
		source:      `a ^ b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 6, operator: "^"}},
	},
	{
		name:        "upstream invalid[1]",
		source:      `a | b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 6, operator: "|"}},
	},
	{
		name:        "upstream invalid[2]",
		source:      `a & b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 6, operator: "&"}},
	},
	{
		name:        "upstream invalid[3]",
		source:      `a << b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 7, operator: "<<"}},
	},
	{
		name:        "upstream invalid[4]",
		source:      `a >> b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 7, operator: ">>"}},
	},
	{
		name:        "upstream invalid[5]",
		source:      `a >>> b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 8, operator: ">>>"}},
	},
	{
		name:        "upstream invalid[6]",
		source:      `a|0`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 4, operator: "|"}},
	},
	{
		name:        "upstream invalid[7]",
		source:      `~a`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 3, operator: "~"}},
	},
	{
		name:        "upstream invalid[8]",
		source:      `a ^= b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 7, operator: "^="}},
	},
	{
		name:        "upstream invalid[9]",
		source:      `a |= b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 7, operator: "|="}},
	},
	{
		name:        "upstream invalid[10]",
		source:      `a &= b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 7, operator: "&="}},
	},
	{
		name:        "upstream invalid[11]",
		source:      `a <<= b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 8, operator: "<<="}},
	},
	{
		name:        "upstream invalid[12]",
		source:      `a >>= b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 8, operator: ">>="}},
	},
	{
		name:        "upstream invalid[13]",
		source:      `a >>>= b`,
		optionsJson: "",
		findings:    []noBitwiseExpectation{{line: 1, column: 1, endLine: 1, endColumn: 9, operator: ">>>="}},
	},
}

// TestNoBitwiseEveryAcceptedSpellingActuallyExempts ties the decoder to the rule.
//
// The decoder decides which spellings a project may write in `allow`; the operator map decides which
// operators the rule can act on. Nothing structural forced those to agree while they were two
// literals, and a mutation changing one spelling in the decoder's list SURVIVED the entire fixture
// set — because each side is internally consistent and every fixture exercises one side or the
// other.
//
// The two failure directions are both silent:
//
//	accepted but not in the map   a project writes it in `allow`, the config loads, nothing is exempted
//	in the map but not accepted   the operator is unexemptable and the config is refused at load
//
// This walks every spelling the decoder accepts and requires it to actually exempt its own operator,
// which is the only assertion that fails when the two drift. The count is asserted too, so removing
// an entry from either side fails here rather than quietly testing less.
func TestNoBitwiseEveryAcceptedSpellingActuallyExempts(t *testing.T) {
	t.Parallel()
	// One source per operator, chosen so the operator under test is the only one present.
	sourceFor := map[string]string{
		"^": `a ^ b`, "|": `a | b`, "&": `a & b`,
		"<<": `a << b`, ">>": `a >> b`, ">>>": `a >>> b`,
		"^=": `a ^= b`, "|=": `a |= b`, "&=": `a &= b`,
		"<<=": `a <<= b`, ">>=": `a >>= b`, ">>>=": `a >>>= b`,
		"~": `~a`,
	}

	accepted := noBitwiseAllowedSpellings()
	if len(accepted) != 13 {
		t.Fatalf("expected upstream's thirteen operators, got %d: %v", len(accepted), accepted)
	}

	for _, spelling := range accepted {
		source, known := sourceFor[spelling]
		if !known {
			t.Errorf("the decoder accepts %q but this test has no source exercising it, so nothing "+
				"proves the rule can act on it", spelling)
			continue
		}
		t.Run(spelling, func(t *testing.T) {
			t.Parallel()
			// It reports with no options...
			rule_testing.ExpectFindings(t,
				runNoBitwise(t, noBitwiseCase{source: source}), messageNoBitwise.Id)

			// ...and allowing this exact spelling silences it. A spelling the decoder accepts that
			// the map does not carry fails right here.
			allowed := runNoBitwise(t, noBitwiseCase{
				source:      source,
				optionsJson: `{"allow":[` + string(mustJson(t, spelling)) + `]}`,
			})
			if len(allowed.Diagnostics) != 0 {
				t.Errorf("allowing %q did not exempt %q: the decoder accepts a spelling the "+
					"operator map does not carry", spelling, source)
			}
		})
	}
}

// mustJson quotes a string the way the config would, so the option text is built rather than typed.
func mustJson(t *testing.T, value string) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encoding %q: %v", value, err)
	}
	return encoded
}
