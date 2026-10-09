package typescript

import (
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noDynamicDeleteFile names the fixture file. The rule reads no path and gates on no extension:
// upstream registers a bare selector with no source-type test, and the installed 8.x build reports
// the same inputs in a `.js` file, which has its own fixture at the end of this file.
const noDynamicDeleteFile = "/repository/source/Container.ts"

// noDynamicDeleteCaseName numbers a row so a failure names which one, since several of these
// sources differ only in one token.
func noDynamicDeleteCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoDynamicDeleteStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All eleven of upstream's passing inputs, extracted from the clone's test file by parsing it with
// the TypeScript compiler rather than by reading it, so no escape sequence passed through a shell
// or a keyboard on the way here. Every one was additionally run through the installed 8.x build
// driven by the ESLint Linter API, which reported nothing on all eleven and produced no parse
// error on any of them.
func TestNoDynamicDeleteStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\nconst container: { [i: string]: 0 } = {};\ndelete container.aaa;\n    ",
		"\nconst container: { [i: string]: 0 } = {};\ndelete container.delete;\n    ",
		"\nconst container: { [i: string]: 0 } = {};\ndelete container[7];\n    ",
		"\nconst container: { [i: string]: 0 } = {};\ndelete container[-7];\n    ",
		"\nconst container: { [i: string]: 0 } = {};\ndelete container['-Infinity'];\n    ",
		"\nconst container: { [i: string]: 0 } = {};\ndelete container['+Infinity'];\n    ",
		"\nconst value = 1;\ndelete value;\n    ",
		"\nconst value = 1;\ndelete -value;\n    ",
		"\nconst container: { [i: string]: 0 } = {};\ndelete container['aaa'];\n    ",
		"\nconst container: { [i: string]: 0 } = {};\ndelete container['delete'];\n    ",
		"\nconst container: { [i: string]: 0 } = {};\ndelete container['NaN'];\n    ",
	}
	for index, sourceText := range cases {
		t.Run(noDynamicDeleteCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDynamicDelete,
				noDynamicDeleteFile, sourceText))
		})
	}
}

// TestNoDynamicDeleteFiresOnUpstreamFailCases is the imported reporting corpus, verbatim.
//
// All ten of upstream's failing inputs. Each carries exactly one error and `output: null`, so the
// count and the absence of a repair are both asserted rather than assumed.
//
// The span on every row is measured rather than transcribed: each input was run through the
// installed build and its reported line and column converted to an offset slice of that same input.
// That matters more here than on most rules, because upstream reports on the KEY rather than on the
// delete expression, and a message-id assertion cannot see the difference.
func TestNoDynamicDeleteFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
	}{
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\ndelete container['aa' + 'b'];\n      ",
			wantSpan:   "'aa' + 'b'",
		},
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\ndelete container[+7];\n      ",
			wantSpan:   "+7",
		},
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\ndelete container[-Infinity];\n      ",
			wantSpan:   "-Infinity",
		},
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\ndelete container[+Infinity];\n      ",
			wantSpan:   "+Infinity",
		},
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\ndelete container[NaN];\n      ",
			wantSpan:   "NaN",
		},
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\nconst name = 'name';\ndelete container[name];\n      ",
			wantSpan:   "name",
		},
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\nconst getName = () => 'aaa';\ndelete container[getName()];\n      ",
			wantSpan:   "getName()",
		},
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\nconst name = { foo: { bar: 'bar' } };\ndelete container[name.foo.bar];\n      ",
			wantSpan:   "name.foo.bar",
		},
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\ndelete container[+'Infinity'];\n      ",
			wantSpan:   "+'Infinity'",
		},
		{
			sourceText: "\nconst container: { [i: string]: 0 } = {};\ndelete container[typeof 1];\n      ",
			wantSpan:   "typeof 1",
		},
	}
	for index, testCase := range cases {
		t.Run(noDynamicDeleteCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDynamicDelete, noDynamicDeleteFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "dynamicDelete")

			diagnostic := result.Diagnostics[0]
			gotSpan := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}

			// Asserted against a literal typed here rather than against the rule's own message
			// constant: comparing a finding to the constant it was built from is an equality that
			// moves in both directions under mutation and therefore guards nothing.
			if diagnostic.Message.Description != "Do not delete dynamically computed property keys." {
				t.Fatalf("message: got %q", diagnostic.Message.Description)
			}
			if diagnostic.Message.Id != "dynamicDelete" {
				t.Fatalf("message id: got %q", diagnostic.Message.Id)
			}

			// `output: null` on every upstream row. The rule ships no repair, and a fixture that
			// only counted findings could not see one appearing.
			if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
				t.Fatalf("expected no repair, got %d fixes and %d suggestions",
					len(diagnostic.Fixes), len(diagnostic.Suggestions))
			}
		})
	}
}

// TestNoDynamicDeleteStaysSilentOnShapesTheCorpusDoesNotWrite covers the clean divergences.
//
// Upstream's corpus writes no parenthesized form and no optional chain anywhere, so nothing in it
// can see either decision. Both are places where the estree grammar and our parse tree differ, and
// a port that matched our kinds directly would report every row here.
//
// Every verdict below was measured by running the input through the installed 8.x build rather than
// derived from reading the rule, because reading the rule file suggests the opposite for the
// parenthesized rows: it tests `property.type === Literal`, and a paren is a real node here.
func TestNoDynamicDeleteStaysSilentOnShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{
			// estree has no parenthesized node, so upstream's Literal test sees through the parens. This is the case a port testing our kinds without skipping parens would report.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[(7)];",
		},
		{
			// The paren skip is needed under the minus arm too, which no reading of the rule file suggests.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[-(7)];",
		},
		{
			// Same skip for a string key.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[('aaa')];",
		},
		{
			// An optional chain is a ChainExpression upstream, so the rule returns before it looks at the key.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c?.[x];",
		},
		{
			// The same input with an ACCEPTABLE key, which is what proves the exclusion is structural rather than a consequence of the index test.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c?.[7];",
		},
		{
			// The OUTERMOST access is a property access, so the operand test declines. The computed access is not the thing being deleted.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[x].a;",
		},
		{
			// A numeric separator is still a numeric literal.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[1_0];",
		},
		{
			// A hex literal is still a numeric literal.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[0x10];",
		},
	}
	for index, testCase := range cases {
		t.Run(noDynamicDeleteCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDynamicDelete,
				noDynamicDeleteFile, testCase.sourceText))
		})
	}
}

// TestNoDynamicDeleteFiresOnShapesTheCorpusDoesNotWrite covers the reporting divergences.
//
// The estree `Literal` node covers string, number, boolean, null, regex and bigint under one type
// and discriminates on `typeof property.value`. Our parser gives each its own kind, so the four
// non-accepted literal kinds are the ones most likely to be waved through by a port that reasoned
// about "is it a literal" rather than about which two values upstream accepts. None of them is in
// the corpus; all four were measured against the installed build.
//
// The remaining rows pin where the finding LANDS on nested and parenthesized shapes, which is the
// half of this rule a message-id assertion cannot see.
func TestNoDynamicDeleteFiresOnShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
	}{
		{
			// A template literal is not a Literal in estree even with no substitutions.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[`aaa`];",
			wantSpan:   "`aaa`",
		},
		{
			// typeof true is 'boolean', which is not in upstream's accepted pair.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[true];",
			wantSpan:   "true",
		},
		{
			// typeof null is 'object'.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[null];",
			wantSpan:   "null",
		},
		{
			// typeof a regex is 'object'.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[/re/];",
			wantSpan:   "/re/",
		},
		{
			// typeof 7n is 'bigint'.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[7n];",
			wantSpan:   "7n",
		},
		{
			// Parens around the whole operand are skipped, and the span excludes them.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete ((c['x' + 'y']));",
			wantSpan:   "'x' + 'y'",
		},
		{
			// A sequence expression is not a literal, and the reported span is the sequence without its parens.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[(y, 'b')];",
			wantSpan:   "y, 'b'",
		},
		{
			// Nested unary minus: the inner operand is a unary expression rather than a numeric literal, so the accept arm declines.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[- -7];",
			wantSpan:   "- -7",
		},
		{
			// Parens on the receiver rather than on the operand or the key.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete (c)[x];",
			wantSpan:   "x",
		},
		{
			// A computed access whose receiver is a property access still reports.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c.a[x];",
			wantSpan:   "x",
		},
		{
			// A `this` receiver reports like any other.
			sourceText: "const c: { [i: string]: 0 } = {};\nclass K { m() { delete this[x]; } }",
			wantSpan:   "x",
		},
		{
			// Nested computed accesses: only the outermost key is the one being deleted, so exactly one finding on `y`.
			sourceText: "const c: { [i: string]: 0 } = {};\ndelete c[x][y];",
			wantSpan:   "y",
		},
	}
	for index, testCase := range cases {
		t.Run(noDynamicDeleteCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDynamicDelete, noDynamicDeleteFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "dynamicDelete")

			gotSpan := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
		})
	}
}

// TestNoDynamicDeleteReportsInJavaScriptToo pins the absence of a file gate.
//
// Upstream registers a bare selector with no source-type test, and the installed build reports the
// same input in a `.js` file. The gate is worth a fixture rather than a comment: three sibling rules
// in this tree carry a `.tsx`-only gate upstream does not have, which cost every finding in every
// other extension until somebody measured it.
func TestNoDynamicDeleteReportsInJavaScriptToo(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoDynamicDelete, "/repository/source/Container.js",
		"const container = {};\ndelete container[name];")
	rule_testing.ExpectFindings(t, result, "dynamicDelete")
}
