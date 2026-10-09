package core

import (
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noNegatedConditionFile is where the fixtures pretend to live.
const noNegatedConditionFile = "/repository/source/NoNegatedCondition.ts"

// noNegatedConditionExpectation is one finding, with the span it points at.
//
// The span is here because `ExpectFindings` asserts message ids and count and nothing else, and this
// rule reports the whole statement rather than the test inside it. A port anchoring on the test
// would pass every message-id assertion while pointing at `!a`.
type noNegatedConditionExpectation struct {
	line      int
	column    int
	endLine   int
	endColumn int
}

type noNegatedConditionCase struct {
	name     string
	source   string
	findings []noNegatedConditionExpectation
	reason   string
}

// noNegatedConditionOffsetOf converts a 1-based line and column into a byte offset.
//
// `rule_testing.Run` does NOT trim its input, unlike RunTyped, so no rebasing is needed here. Stated
// because the sibling rules in this batch that need the typed harness do rebase, and a helper copied
// between them without this line would be quietly wrong by one.
func noNegatedConditionOffsetOf(t *testing.T, source string, line int, column int) int {
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

func runNoNegatedCondition(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.Run(t, NoNegatedCondition, noNegatedConditionFile, source)
}

// TestNoNegatedConditionStaysSilent runs upstream's whole `valid` list.
func TestNoNegatedConditionStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range noNegatedConditionCleanCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoNegatedCondition(t, testCase.source))
		})
	}
}

// TestNoNegatedConditionFires runs upstream's whole `invalid` list and asserts every span.
func TestNoNegatedConditionFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range noNegatedConditionReportingCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoNegatedCondition(t, testCase.source)

			wantIds := make([]string, 0, len(testCase.findings))
			for range testCase.findings {
				wantIds = append(wantIds, messageNoNegatedCondition.Id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, expected := range testCase.findings {
				diagnostic := result.Diagnostics[index]
				wantPos := noNegatedConditionOffsetOf(t, testCase.source, expected.line, expected.column)
				wantEnd := noNegatedConditionOffsetOf(t, testCase.source, expected.endLine, expected.endColumn)
				if diagnostic.Range.Pos() != wantPos || diagnostic.Range.End() != wantEnd {
					t.Errorf("finding %d: expected span [%d,%d), got [%d,%d) %q",
						index, wantPos, wantEnd, diagnostic.Range.Pos(), diagnostic.Range.End(),
						testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()])
				}
			}
		})
	}
}

// TestNoNegatedConditionParenthesesUpstreamCannotWrite is the shape no imported fixture can see.
//
// Upstream's parser folds parentheses away before its test runs, so its corpus contains no
// parenthesized condition anywhere and could not express one if it wanted to. Our parser keeps
// `KindParenthesizedExpression`, so a port without the unwrap loop declines all of these and loses
// the finding with no fixture noticing.
//
// Every expectation was taken by driving the installed 10.8.1 build, not derived.
func TestNoNegatedConditionParenthesesUpstreamCannotWrite(t *testing.T) {
	t.Parallel()
	cases := []noNegatedConditionCase{
		{
			name:   "a parenthesized unary negation reports",
			source: `if ((!a)) { f(); } else { g(); }`,
			findings: []noNegatedConditionExpectation{
				{line: 1, column: 1, endLine: 1, endColumn: 33},
			},
			reason: "measured at columns 1 to 33 upstream",
		},
		{
			name:   "a parenthesized inequality reports",
			source: `if ((a !== b)) { f(); } else { g(); }`,
			findings: []noNegatedConditionExpectation{
				{line: 1, column: 1, endLine: 1, endColumn: 38},
			},
			reason: "the binary arm needs the same unwrap as the unary one",
		},
		{
			name:   "a doubly parenthesized negation reports, which is what makes the unwrap a loop",
			source: `if (((!a))) { f(); } else { g(); }`,
			findings: []noNegatedConditionExpectation{
				{line: 1, column: 1, endLine: 1, endColumn: 35},
			},
			reason: "a single unwrap step answers false here and this is the only case that says so",
		},
		{
			name:   "a parenthesized conditional expression reports",
			source: `(!a) ? b : c`,
			findings: []noNegatedConditionExpectation{
				{line: 1, column: 1, endLine: 1, endColumn: 13},
			},
			reason: "the conditional arm reads its condition through the same helper",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoNegatedCondition(t, testCase.source)
			wantIds := make([]string, 0, len(testCase.findings))
			for range testCase.findings {
				wantIds = append(wantIds, messageNoNegatedCondition.Id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
			for index, expected := range testCase.findings {
				diagnostic := result.Diagnostics[index]
				wantPos := noNegatedConditionOffsetOf(t, testCase.source, expected.line, expected.column)
				wantEnd := noNegatedConditionOffsetOf(t, testCase.source, expected.endLine, expected.endColumn)
				if diagnostic.Range.Pos() != wantPos || diagnostic.Range.End() != wantEnd {
					t.Errorf("expected span [%d,%d), got [%d,%d) %q. %s",
						wantPos, wantEnd, diagnostic.Range.Pos(), diagnostic.Range.End(),
						testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()], testCase.reason)
				}
			}
		})
	}
}

// TestNoNegatedConditionShapesThatMustStaySilent pins the boundaries, measured the same way.
//
// The first two are the ones a reader would expect to report and that upstream deliberately does
// not, and both were confirmed against the installed build rather than argued from the source.
func TestNoNegatedConditionShapesThatMustStaySilent(t *testing.T) {
	t.Parallel()
	cases := []noNegatedConditionCase{
		{
			name:   "an equality against false is not a negation",
			source: `if (a === false) { f(); } else { g(); }`,
			reason: "upstream names three operators literally and `===` is not one of them",
		},
		{
			name:   "an else-if chain ending in an else is still clean",
			source: `if (!a) { f(); } else if (b) { g(); } else { h(); }`,
			reason: "the trailing else belongs to the INNER if, whose test `b` is not negated. The " +
				"outer if's else IS an IfStatement, so hasElseWithoutCondition declines it.",
		},
		{
			name:   "a negated condition with no else at all is clean",
			source: `if (!a) { f(); }`,
			reason: "the whole rule is about the else the reader has to invert into",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoNegatedCondition(t, testCase.source)
			if len(result.Diagnostics) != 0 {
				t.Errorf("expected silence, got %v. %s", result.MessageIds(), testCase.reason)
			}
		})
	}
}

// TestNoNegatedConditionCompoundNegationReports is the case that reads like a compound condition.
//
// `!(a && b)` is a `!` and reports, which is worth a case of its own because it looks like a
// parenthesized binary rather than a negation and the rule's three-operator list does not obviously
// cover it. Measured upstream at columns 1 to 38.
func TestNoNegatedConditionCompoundNegationReports(t *testing.T) {
	t.Parallel()
	result := runNoNegatedCondition(t, `if (!(a && b)) { f(); } else { g(); }`)
	rule_testing.ExpectFindings(t, result, messageNoNegatedCondition.Id)
}

// noNegatedConditionCleanCases are upstream's `valid` list, verbatim.
var noNegatedConditionCleanCases = []noNegatedConditionCase{
	{name: "upstream valid[0]", source: `if (a) {}`},
	{name: "upstream valid[1]", source: `if (a) {} else {}`},
	{name: "upstream valid[2]", source: `if (!a) {}`},
	{name: "upstream valid[3]", source: `if (!a) {} else if (b) {}`},
	{name: "upstream valid[4]", source: `if (!a) {} else if (b) {} else {}`},
	{name: "upstream valid[5]", source: `if (a == b) {}`},
	{name: "upstream valid[6]", source: `if (a == b) {} else {}`},
	{name: "upstream valid[7]", source: `if (a != b) {}`},
	{name: "upstream valid[8]", source: `if (a != b) {} else if (b) {}`},
	{name: "upstream valid[9]", source: `if (a != b) {} else if (b) {} else {}`},
	{name: "upstream valid[10]", source: `if (a !== b) {}`},
	{name: "upstream valid[11]", source: `if (a === b) {} else {}`},
	{name: "upstream valid[12]", source: `a ? b : c`},
}

// noNegatedConditionReportingCases are upstream's `invalid` list, verbatim.
//
// Upstream asserts only a count on every one of these, so the spans were measured by driving
// the installed 10.8.1 build rather than copied.
var noNegatedConditionReportingCases = []noNegatedConditionCase{
	{
		name:   "upstream invalid[0]",
		source: `if (!a) {;} else {;}`,
		findings: []noNegatedConditionExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 21},
		},
	},
	{
		name:   "upstream invalid[1]",
		source: `if (a != b) {;} else {;}`,
		findings: []noNegatedConditionExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 25},
		},
	},
	{
		name:   "upstream invalid[2]",
		source: `if (a !== b) {;} else {;}`,
		findings: []noNegatedConditionExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 26},
		},
	},
	{
		name:   "upstream invalid[3]",
		source: `!a ? b : c`,
		findings: []noNegatedConditionExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 11},
		},
	},
	{
		name:   "upstream invalid[4]",
		source: `a != b ? c : d`,
		findings: []noNegatedConditionExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 15},
		},
	},
	{
		name:   "upstream invalid[5]",
		source: `a !== b ? c : d`,
		findings: []noNegatedConditionExpectation{
			{line: 1, column: 1, endLine: 1, endColumn: 16},
		},
	},
}

// TestNoNegatedConditionOnlyTheBangOperatorCounts pins the unary arm against every other prefix.
//
// Upstream tests `operator === "!"` and its corpus never writes another prefix operator in a
// condition that has an else, so a mutation widening the test to accept ANY prefix unary survives
// the whole imported corpus. Every row here was measured against the installed 10.8.1 build.
//
// `~` is the one worth noting: it is a bitwise negation and reads as a negation in English, and it
// is clean. `!!a` reports because its OUTER operator is `!`, which is the row that shows the test is
// on the operator rather than on whether the value is inverted.
func TestNoNegatedConditionOnlyTheBangOperatorCounts(t *testing.T) {
	t.Parallel()
	clean := []struct {
		name   string
		source string
	}{
		{"typeof", `if (typeof a) { f(); } else { g(); }`},
		{"void", `if (void a) { f(); } else { g(); }`},
		{"unary minus", `if (-a) { f(); } else { g(); }`},
		{"unary plus", `if (+a) { f(); } else { g(); }`},
		{"bitwise not", `if (~a) { f(); } else { g(); }`},
		{"delete", `if (delete a.b) { f(); } else { g(); }`},
		{"prefix increment", `if (++a) { f(); } else { g(); }`},
		{"unary minus in a conditional expression", `-a ? b : c`},
	}
	for _, testCase := range clean {
		t.Run(testCase.name+" is clean", func(t *testing.T) {
			t.Parallel()
			result := runNoNegatedCondition(t, testCase.source)
			if len(result.Diagnostics) != 0 {
				t.Errorf("expected silence, got %v: only `!` counts as a negation here",
					result.MessageIds())
			}
		})
	}

	t.Run("a double negation reports, because the outer operator is still !", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, runNoNegatedCondition(t, `if (!!a) { f(); } else { g(); }`),
			messageNoNegatedCondition.Id)
	})
}
