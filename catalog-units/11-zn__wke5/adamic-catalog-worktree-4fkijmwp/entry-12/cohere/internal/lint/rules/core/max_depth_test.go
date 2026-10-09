package core

import (
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// maxDepthFile is where the fixtures pretend to live.
const maxDepthFile = "/repository/source/MaxDepth.ts"

// maxDepthOf builds the settings a config carrying this option would decode to.
//
// The argument is the option as JSON TEXT, exactly the bytes the config layer hands the decoder,
// so a fixture cannot bypass the decoder or accidentally test a shape a config could not deliver.
// That is section 7c of the standard applied at the fixture rather than only in a contract test.
func maxDepthOf(optionJson string) any {
	settings, err := DecodeMaxDepthOptions([]byte(optionJson))
	if err != nil {
		panic(err)
	}
	return settings
}

// The corpus is ESLint's own, taken from
// /tmp/lint-sources-fresh/eslint/tests/lib/rules/max-depth.js by RUNNING that file with RuleTester
// intercepted, so upstream's own cases produced the fixtures rather than a parser of mine. 28 cases
// were captured from the single RuleTester.run call.
//
// Every expectation is what the INSTALLED ESLint 10.8.1 rule answered when driven over that case
// through the Linter interface under `sourceType: module` with the typescript-eslint parser, which
// is the only configuration cohere has. The oracle reproduced all 28 of the corpus's own declared
// verdicts under upstream's own languageOptions, which is the control saying it measured the rule,
// and ZERO cases change verdict between that configuration and cohere's.
//
//	17 reporting cases, 11 clean cases.
func TestMaxDepthFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		findings int
	}{
		{source: "function foo() { if (true) { if (false) { if (true) { } } } }", settings: maxDepthOf("2"), findings: 1},
		{source: "var foo = () => { if (true) { if (false) { if (true) { } } } }", settings: maxDepthOf("2"), findings: 1},
		{source: "function foo() { if (true) {} else { for(;;) {} } }", settings: maxDepthOf("1"), findings: 1},
		{source: "function foo() { if (a) {} else if (b) {} if (c) { if (d) {} } }", settings: maxDepthOf("1"), findings: 1},
		{source: "if (a) if (b) {}", settings: maxDepthOf("1"), findings: 1},
		{source: "function foo() { while (true) { if (true) {} } }", settings: maxDepthOf("1"), findings: 1},
		{source: "function foo() { if (a) { switch (b) { case 1: foo(); } } }", settings: maxDepthOf("1"), findings: 1},
		{source: "function foo() { for (let x of foo) { if (true) {} } }", settings: maxDepthOf("1"), findings: 1},
		{source: "function foo() { while (true) { if (true) { if (false) { } } } }", settings: maxDepthOf("1"), findings: 2},
		{source: "function foo() { if (true) { if (false) { if (true) { if (false) { if (true) { } } } } } }", settings: nil, findings: 1},
		{source: "function foo() { if (true) { if (false) { if (true) { } } } }", settings: maxDepthOf("{\"max\": 2}"), findings: 1},
		{source: "function foo() { if (a) { if (b) { if (c) { if (d) { if (e) {} } } } } }", settings: maxDepthOf("{}"), findings: 1},
		{source: "function foo() { if (true) {} }", settings: maxDepthOf("{\"max\": 0}"), findings: 1},
		{source: "class C { static { if (1) { if (2) { if (3) {} } } } }", settings: maxDepthOf("2"), findings: 1},
		{source: "if (1) { class C { static { if (1) { if (2) { if (3) {} } } } } }", settings: maxDepthOf("2"), findings: 1},
		{source: "function foo() { if (1) { class C { static { if (1) { if (2) { if (3) {} } } } } } }", settings: maxDepthOf("2"), findings: 1},
		{source: "function foo() { if (1) { class C { static { if (1) { if (2) {} } } } if (2) { if (3) {} } } }", settings: maxDepthOf("2"), findings: 1},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, MaxDepth, maxDepthFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q with %+v: expected %d findings, got %d %v",
				testCase.source, testCase.settings, testCase.findings,
				len(result.Diagnostics), result.MessageIds())
			continue
		}
		expected := make([]string, 0, testCase.findings)
		for range testCase.findings {
			expected = append(expected, messageMaxDepthTooDeeply.Id)
		}
		rule_testing.ExpectFindings(t, result, expected...)
	}
}

// TestMaxDepthStaysSilent carries upstream's clean cases, which are the false positives it already
// thought about. Each one is a class this port would otherwise ship.
func TestMaxDepthStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
	}{
		{source: "function foo() { if (true) { if (false) { if (true) { } } } }", settings: maxDepthOf("3")},
		{source: "function foo() { if (true) { } else if (false) { } else if (true) { } else if (false) {} }", settings: maxDepthOf("3")},
		{source: "var foo = () => { if (true) { if (false) { if (true) { } } } }", settings: maxDepthOf("3")},
		{source: "function foo() { if (true) { if (false) { if (true) { } } } }", settings: nil},
		{source: "function foo() { if (true) { if (false) { if (true) { } } } }", settings: maxDepthOf("{\"max\": 3}")},
		{source: "class C { static { if (1) { if (2) {} } } }", settings: maxDepthOf("2")},
		{source: "class C { static { if (1) { if (2) {} } if (1) { if (2) {} } } }", settings: maxDepthOf("2")},
		{source: "class C { static { if (1) { if (2) {} } } static { if (1) { if (2) {} } } }", settings: maxDepthOf("2")},
		{source: "if (1) { class C { static { if (1) { if (2) {} } } } }", settings: maxDepthOf("2")},
		{source: "function foo() { if (1) { class C { static { if (1) { if (2) {} } } } } }", settings: maxDepthOf("2")},
		{source: "function foo() { if (1) { if (2) { class C { static { if (1) { if (2) {} } if (1) { if (2) {} } } } } } if (1) { if (2) {} } }", settings: maxDepthOf("2")},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, MaxDepth, maxDepthFile, testCase.source, testCase.settings)
		// `ExpectClean` rather than a length check, because the guard in internal/lint/registry
		// matches the CALL NAME.
		rule_testing.ExpectClean(t, result)
	}
}

// TestMaxDepthSpansAndMessages asserts where each finding POINTS and what it says, which
// TestMaxDepthFires cannot: that test compares ids and counts, so a rule anchored on the whole
// statement instead of its keyword passes it completely.
//
// The span is the sharpest thing to get wrong here. Upstream reports
// `loc: sourceCode.getFirstToken(node).loc`, so the finding covers the KEYWORD alone.
// `rule.TokenRange` -- the obvious helper, and the one nearly every other rule in this package
// uses -- extends to `node.End()` and would underline the entire statement including its body.
// Nothing about a message id or a count would say so.
//
// Every expected span and message is what the installed ESLint 10.8.1 rule produced on the same
// source, read out of the oracle run rather than derived by hand.
func TestMaxDepthSpansAndMessages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		spans    []string
		messages []string
	}{
		{
			source:   "function foo() { if (true) { if (false) { if (true) { } } } }",
			settings: maxDepthOf("2"),
			spans:    []string{"if"},
			messages: []string{"Blocks are nested too deeply (3). Maximum allowed is 2."},
		},
		{
			// A switch keyword is six characters, so a span covering the statement rather than the
			// token is visible here in a way the two-character `if` cases could hide.
			source:   "function foo() { if (a) { switch (b) { case 1: foo(); } } }",
			settings: maxDepthOf("1"),
			spans:    []string{"switch"},
			messages: []string{"Blocks are nested too deeply (2). Maximum allowed is 1."},
		},
		{
			source:   "function foo() { for (let x of foo) { if (true) {} } }",
			settings: maxDepthOf("1"),
			spans:    []string{"if"},
			messages: []string{"Blocks are nested too deeply (2). Maximum allowed is 1."},
		},
		{
			// Two findings from one function, so the reported DEPTH has to rise between them. A
			// rule reporting a constant depth passes every count-based fixture.
			source:   "function foo() { while (true) { if (true) { if (false) { } } } }",
			settings: maxDepthOf("1"),
			spans:    []string{"if", "if"},
			messages: []string{
				"Blocks are nested too deeply (2). Maximum allowed is 1.",
				"Blocks are nested too deeply (3). Maximum allowed is 1.",
			},
		},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, MaxDepth, maxDepthFile, testCase.source, testCase.settings)
		if len(result.Diagnostics) != len(testCase.spans) {
			t.Errorf("%q: expected %d findings, got %d %v",
				testCase.source, len(testCase.spans), len(result.Diagnostics), result.MessageIds())
			continue
		}
		source := result.SourceFile.Text()
		for index, diagnostic := range result.Diagnostics {
			reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.spans[index] {
				t.Errorf("%q finding %d: expected the span to cover %q, got %q",
					testCase.source, index, testCase.spans[index], reported)
			}
			// The whole computed sentence: both the depth and the limit are interpolated, and the
			// depth is the half that a defect would get wrong.
			if !strings.HasPrefix(diagnostic.Message.Description, testCase.messages[index]) {
				t.Errorf("%q finding %d: expected the message to open with %q, got %q",
					testCase.source, index, testCase.messages[index], diagnostic.Message.Description)
			}
		}
	}
}

// TestMaxDepthDecoderReadsEveryOptionShape pins the option surface, which the fixtures above
// exercise only where the corpus happened to configure it.
//
// Upstream's schema is a `oneOf` and its object form carries a JavaScript quirk that no reading of
// the source alone predicts: the `hasOwn` guard decides whether the branch is ENTERED and the `||`
// inside decides the VALUE, and the two disagree exactly when a zero is written. An earlier draft
// of the decoder read it as a plain truthiness fallback and got both zero rows wrong.
//
// Every expected limit is read out of the installed rule's own message, which states maxDepth
// directly, rather than inferred from a finding count.
func TestMaxDepthDecoderReadsEveryOptionShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		// optionJson is exactly the bytes the config layer hands the decoder.
		optionJson string
		maximum    int
		disabled   bool
	}{
		{optionJson: `2`, maximum: 2},
		{optionJson: `0`, maximum: 0},
		{optionJson: `{"maximum": 2}`, maximum: 2},
		{optionJson: `{"max": 2}`, maximum: 2},
		// Neither key present, so upstream never enters the branch and the default of 4 stands.
		{optionJson: `{}`, maximum: 4},
		// `undefined || 0` is 0, so a lone `max` of zero WINS and reports every block.
		{optionJson: `{"max": 0}`, maximum: 0},
		// `0 || undefined` is undefined, so every comparison is false and the rule goes silent.
		// No integer limit can express this, which is why it is a flag.
		{optionJson: `{"maximum": 0}`, disabled: true},
		// The `||` consults maximum first, so a truthy maximum wins over any max. The zero rows
		// below do NOT show that precedence -- whichever key is non-zero wins under either order,
		// so swapping the two branches survives them. These two rows are the ones that
		// discriminate, because both keys are non-zero and different:
		//
		//	{"max": 5, "maximum": 2}   limit 2   measured against the installed rule
		//	{"maximum": 5, "max": 2}   limit 5
		{optionJson: `{"max": 5, "maximum": 2}`, maximum: 2},
		{optionJson: `{"maximum": 5, "max": 2}`, maximum: 5},
		{optionJson: `{"max": 0, "maximum": 3}`, maximum: 3},
		{optionJson: `{"maximum": 0, "max": 3}`, maximum: 3},
	}

	for _, testCase := range cases {
		decoded, err := DecodeMaxDepthOptions([]byte(testCase.optionJson))
		if err != nil {
			t.Errorf("decoding %s: %v", testCase.optionJson, err)
			continue
		}
		settings, ok := decoded.(MaxDepthSettings)
		if !ok {
			t.Errorf("decoding %s: expected MaxDepthSettings, got %T", testCase.optionJson, decoded)
			continue
		}
		if settings.Disabled != testCase.disabled {
			t.Errorf("decoding %s: expected disabled=%v, got %v",
				testCase.optionJson, testCase.disabled, settings.Disabled)
			continue
		}
		if !testCase.disabled && settings.Maximum != testCase.maximum {
			t.Errorf("decoding %s: expected a limit of %d, got %d",
				testCase.optionJson, testCase.maximum, settings.Maximum)
		}
	}

	// And end to end, because a decoder row is not a verdict. The disabled shape must report
	// NOTHING on source that a limit of 0 would report five times.
	nested := "function f() { if (a) { if (b) { if (c) { if (d) { if (e) {} } } } } }"
	result := rule_testing.RunWithOptions(t, MaxDepth, maxDepthFile, nested, maxDepthOf(`{"maximum": 0}`))
	rule_testing.ExpectClean(t, result)

	// The control: the same source under the sibling zero spelling reports every block, so the
	// silence above is the option rather than a rule that cannot fire on this input.
	result = rule_testing.RunWithOptions(t, MaxDepth, maxDepthFile, nested, maxDepthOf(`{"max": 0}`))
	if len(result.Diagnostics) != 5 {
		t.Errorf("expected the control to report all 5 blocks, got %d: %v",
			len(result.Diagnostics), result.MessageIds())
	}

	// A rule handed nil options enforces upstream's default of 4 rather than doing nothing.
	result = rule_testing.RunWithOptions(t, MaxDepth, maxDepthFile, nested, nil)
	if len(result.Diagnostics) != 1 {
		t.Errorf("expected the default limit of 4 to report once, got %d: %v",
			len(result.Diagnostics), result.MessageIds())
	}
}

// TestMaxDepthClassMembersAreEachTheirOwnFrame covers what upstream's corpus cannot.
//
// Upstream's `startFunction` set has no class-member entry and needs none: in ESTree a method's
// body is a FunctionExpression, which the set already covers, so the frame is pushed by the body.
// Our parser hangs the body directly off the member -- probed, there is no wrapping node at all --
// so this port needs an explicit arm for methods, accessors and constructors. It survived a
// mutation sweep against all 28 imported fixtures, because upstream's corpus for this rule writes
// no classes.
//
// # The obvious discriminating shape is the wrong one, and it cost a round
//
// The first version of this test used sibling methods, on the argument that without a frame the
// second would inherit the first one's depth. That argument is wrong: the depth DECREMENTS on the
// way out, so siblings never accumulate either way, and the mutation survived these fixtures too.
//
// A frame RESETS a count, so it can only show where there is a count to reset: a class nested
// inside an already-deep block. Every row below is what the installed ESLint 10.8.1 rule answered,
// and deleting the arm makes all four fail at once while the sibling rows stay green.
func TestMaxDepthClassMembersAreEachTheirOwnFrame(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		limit    string
		findings int
	}{
		// The four rows that discriminate. Without the member arm the inner `if` inherits the
		// enclosing block's depth instead of starting over, and each reports once.
		{source: "function f() { if (a) { class C { m() { if (b) {} } } } }", limit: "1", findings: 0},
		{source: "function f() { if (a) { if (b) { class C { m() { if (c) {} } } } } }", limit: "2", findings: 0},
		{source: "function f() { if (a) { class C { get g() { if (b) {} } } } }", limit: "1", findings: 0},
		{source: "function f() { if (a) { class C { constructor() { if (b) {} } } } }", limit: "1", findings: 0},

		// The sibling rows, kept as a record of what does NOT discriminate: they pass with the arm
		// and without it, because the depth decrements on exit. A reader who deletes the arm and
		// sees these stay green should look at the rows above.
		{source: "class C { a() { if (x) {} } b() { if (y) {} } }", limit: "1", findings: 0},
		{source: "class C { get g() { if (a) {} } set s(v) { if (b) {} } }", limit: "1", findings: 0},

		// The control: a frame is a frame rather than a blanket exemption, so nesting twice inside
		// one member still reports.
		{source: "class C { m() { if (a) { if (b) {} } } }", limit: "1", findings: 1},
		{source: "class C { constructor() { if (a) { if (b) {} } } }", limit: "1", findings: 1},
		// A static block is in upstream's own set, so this passes without the member arm. It is the
		// control that says the walk reaches class bodies at all.
		{source: "class C { static { if (a) { if (b) {} } } }", limit: "1", findings: 1},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, MaxDepth, maxDepthFile, testCase.source,
			maxDepthOf(testCase.limit))
		if len(result.Diagnostics) != testCase.findings {
			t.Errorf("%q at limit %s: expected %d findings, got %d %v",
				testCase.source, testCase.limit, testCase.findings,
				len(result.Diagnostics), result.MessageIds())
		}
	}
}
