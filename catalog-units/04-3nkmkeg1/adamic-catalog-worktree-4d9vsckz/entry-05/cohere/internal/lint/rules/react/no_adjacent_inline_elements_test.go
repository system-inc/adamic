package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noAdjacentInlineElementsFile is where the fixtures pretend to live.
//
// A .tsx extension because most cases hold JSX. The rule has no suffix gate, which a case below
// pins by writing reporting source to a plain `.ts` file.
const noAdjacentInlineElementsFile = "/repository/source/NoAdjacentInlineElements.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/no-adjacent-inline-elements.js, read by
// evaluating the two arrays in the tester with a stubbed RuleTester and emitting Go raw strings from
// the decoded values, so no escape sequence was typed on the way here. Upstream carries 14 valid and
// 3 invalid cases, each invalid one naming exactly one finding.
//
// All 17 were replayed against the installed build, eslint-plugin-react 7.37.5, through the ESLint
// Linter API before any Go was written, and the corpus and the running rule agreed on every one.
//
// Read the rule's doc comment before adding to these tables. Six of the passing cases look like
// evidence that the rule tests for whitespace and are not evidence about that at all, which is the
// single most misleading thing in this corpus.

// TestNoAdjacentInlineElementsFires runs the three failing cases from upstream.
func TestNoAdjacentInlineElementsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"invalid 0 two identical anchors", `<div><a></a><a></a></div>;`},
		{"invalid 1 an anchor and a span", `<div><a></a><span></span></div>;`},
		{"invalid 2 two createElement children in an array", `React.createElement("div", undefined, [React.createElement("a"), React.createElement("span")]);`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "inlineElement")
		})
	}
}

// TestNoAdjacentInlineElementsStaysSilent runs the fourteen passing cases from upstream.
func TestNoAdjacentInlineElementsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"valid 0 an element with no children", `<div />;`},
		{"valid 1 two block elements", `<div><div></div><div></div></div>;`},
		{"valid 2 a paragraph then a block", `<div><p></p><div></div></div>;`},
		{"valid 3 a paragraph then an anchor", `<div><p></p><a></a></div>;`},
		// The next six read as evidence that the rule tests for whitespace between the two
		// elements. They are not: any JSX text at all separates them, because text inside a tag is
		// JSXText rather than a Literal and upstream's isInline names no such arm. The cases in
		// TestNoAdjacentInlineElementsIgnoresWhitespaceInJsx are the ones that establish that,
		// and they are why these six cannot be read as a whitespace test.
		{"valid 4 a non breaking space between anchors", `<div><a></a>&nbsp;<a></a></div>;`},
		{"valid 5 a non breaking space and text", `<div><a></a>&nbsp;some text &nbsp; <a></a></div>;`},
		{"valid 6 a non breaking space then text", `<div><a></a>&nbsp;some text <a></a></div>;`},
		{"valid 7 a plain space between anchors", `<div><a></a> <a></a></div>;`},
		{"valid 8 anchors nested in separate list items", `<div><ul><li><a></a></li><li><a></a></li></ul></div>;`},
		{"valid 9 text between anchors", `<div><a></a> some text <a></a></div>;`},
		{"valid 10 a createElement call with only text", `React.createElement("div", null, "some text");`},
		{"valid 11 a padded string between two createElement children", `React.createElement("div", undefined, [React.createElement("a"), " some text ", React.createElement("a")]);`},
		{"valid 12 a single space between two createElement children", `React.createElement("div", undefined, [React.createElement("a"), " ", React.createElement("a")]);`},
		{"valid 13 a createElement call with no children argument", `React.createElement(a, b);`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoAdjacentInlineElementsHasNoFileSuffixGate pins that the rule reads a plain `.ts` file.
//
// Three siblings in this package gate on `.tsx`/`.jsx`, inherited from oxc. This rule is ported
// from the authority, which has no such gate, and the createElement arm needs no JSX at all.
//
// `.jsx` and `.js` are not covered because the typed program's tsconfig includes only TypeScript
// extensions, which is a fact about the harness rather than about the rule.
func TestNoAdjacentInlineElementsHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	for _, suffix := range []string{".tsx", ".ts"} {
		t.Run(suffix, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(
				t,
				NoAdjacentInlineElements,
				"/repository/source/SuffixProbe"+suffix,
				`React.createElement("div", undefined, [React.createElement("a"), React.createElement("span")]);`,
			)
			rule_testing.ExpectFindings(t, result, "inlineElement")
		})
	}
}

// TestNoAdjacentInlineElementsIgnoresWhitespaceInJsx is the test the corpus cannot be.
//
// Six of upstream's passing cases separate two anchors with whitespace, so they read as a whitespace
// test. Applying the discipline of changing the one thing the shape suggests matters, the verdict
// does NOT move: a single non-space character is just as clean. Text written inside a tag is
// `JSXText` in the tree upstream walks, its `isInline` names no such arm, so it falls through to
// false and breaks the adjacency run whatever it holds.
//
// A port that read those six cases as a whitespace test would pass all seventeen imported cases and
// be wrong about the whole JSX arm. That is the exact failure the corpus cannot catch, because every
// fixture that could distinguish the two readings happens to use whitespace.
func TestNoAdjacentInlineElementsIgnoresWhitespaceInJsx(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a space separates", `<div><a></a> <a></a></div>;`, nil},
		// The case that gives it away. If whitespace were what mattered, this would report.
		{"a single non space character separates just as well", `<div><a></a>x<a></a></div>;`, nil},
		{"a non breaking space separates", `<div><a></a>&nbsp;<a></a></div>;`, nil},
		{"an expression container separates", `<div><a></a>{x}<a></a></div>;`, nil},
		{"a comment container separates", `<div><a></a>{/* c */}<a></a></div>;`, nil},
		{"a newline separates", "<div><a></a>\n<a></a></div>;", nil},
		// Nothing between them is the only reporting shape on this arm.
		{"nothing between them reports", `<div><a></a><a></a></div>;`, []string{"inlineElement"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoAdjacentInlineElementsWhitespaceMattersInCreateElement covers the arm where the regex is
// reachable.
//
// In a children array a string IS a Literal, so `/(?:^\s|\s$)/` decides. The corpus writes two of
// these and the rest were measured, including the empty string, which is the sharp one: it has no
// whitespace to find, so it counts as inline and adjacency holds.
func TestNoAdjacentInlineElementsWhitespaceMattersInCreateElement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a single space separates", `React.createElement("div", undefined, [React.createElement("a"), " ", React.createElement("a")]);`, nil},
		{"trailing whitespace is enough", `React.createElement("div", undefined, [React.createElement("a"), "x ", React.createElement("a")]);`, nil},
		{"leading whitespace is enough", `React.createElement("div", undefined, [React.createElement("a"), " x", React.createElement("a")]);`, nil},
		{"a bare word does not separate", `React.createElement("div", undefined, [React.createElement("a"), "x", React.createElement("a")]);`, []string{"inlineElement"}},
		// An empty string has no edge whitespace, so it is itself inline and the run continues.
		{"an empty string does not separate", `React.createElement("div", undefined, [React.createElement("a"), "", React.createElement("a")]);`, []string{"inlineElement"}},
		// A number stringifies to something with no whitespace, so it counts as inline too.
		{"a number is inline", `React.createElement("div", undefined, [1, React.createElement("a")]);`, []string{"inlineElement"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoAdjacentInlineElementsDoesNotCrashWhereUpstreamDoes records the one divergence.
//
// Upstream's `isInline` tests that a child is a call and then indexes `node.arguments[0].value` with
// no guard on either the argument count or the callee. Both inputs below throw
// `Cannot read properties of undefined (reading 'value')` on the installed build, measured by
// catching around the Linter call.
//
// This port guards and reaches the verdict upstream would have reached: a call with no arguments is
// not inline. Reproducing the crash would cost far more than the divergence, because the walk
// recovers per FILE rather than per rule, so one panic takes every rule's findings for that file.
//
// The crash is unreachable from the JSX arm, where a call only appears inside an expression
// container, which upstream never inspects. The third case pins that.
func TestNoAdjacentInlineElementsDoesNotCrashWhereUpstreamDoes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a no argument call before an inline child", `React.createElement("div", undefined, [foo(), React.createElement("a")]);`, nil},
		{"a no argument call after an inline child", `React.createElement("div", undefined, [React.createElement("a"), bar()]);`, nil},
		{"a call inside a jsx expression container is not a child upstream reads", `<div><a></a>{foo()}<a></a></div>;`, nil},
		// Two no-argument calls in a row are both non-inline, so nothing reports for that reason
		// too, which keeps the guard from being a blanket silence.
		{"two guarded calls still report nothing", `React.createElement("div", undefined, [foo(), bar()]);`, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoAdjacentInlineElementsReportsOncePerContainer pins upstream's early return.
//
// `validate` returns as soon as it reports, so a run of three inline elements yields ONE finding and
// a container holding two separated pairs also yields one. Both measured. A loop that kept going
// would report twice on inputs upstream reports once, and the corpus's longest failing case holds a
// single pair, so nothing in it could see the difference.
func TestNoAdjacentInlineElementsReportsOncePerContainer(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"three in a row report once", `<div><a></a><a></a><a></a></div>;`},
		{"two separated pairs report once", `<div><a></a><a></a> <a></a><a></a></div>;`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "inlineElement")
		})
	}
}

// TestNoAdjacentInlineElementsNameMatching covers which names are on the inline list.
//
// The list holds lowercase HTML names, so a component reference is a different name whatever it
// renders, and a member-named tag has no plain name to compare at all. Both measured.
func TestNoAdjacentInlineElementsNameMatching(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"capitalised components are not on the list", `<div><A></A><A></A></div>;`, nil},
		{"a member named tag has no plain name", `<div><a.b></a.b><a.b></a.b></div>;`, nil},
		{"a block element is not on the list", `<div><p></p><p></p></div>;`, nil},
		{"a block before an inline does not report", `<div><p></p><a></a></div>;`, nil},
		{"an inline before a block does not report", `<div><a></a><p></p></div>;`, nil},
		// Self-closing is the shape most of these are written in and it never produces a
		// JsxOpeningElement here, so it is worth its own case.
		{"self closing inline elements report", `<div><br /><br /></div>;`, []string{"inlineElement"}},
		// An entry on upstream's list that does not read as "inline text", carried because MDN
		// classes it that way. Deleting it would silence the rule on something somebody added
		// deliberately.
		{"button is on upstream's list", `<div><button></button><button></button></div>;`, []string{"inlineElement"}},
		{"input is on upstream's list", `<div><input /><input /></div>;`, []string{"inlineElement"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoAdjacentInlineElementsCreateElementShapes covers the outer call's gates.
//
// The outer call must be a createElement call and its third argument must be an array. Neither is
// stated by the corpus beyond one case, and the asymmetry with the INNER children is the surprising
// part: upstream gates the outer call and never checks the inner ones at all.
func TestNoAdjacentInlineElementsCreateElementShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a foreign outer namespace declines", `NotReact.createElement("div", undefined, [NotReact.createElement("a"), NotReact.createElement("span")]);`, nil},
		{"a bare outer call with no import declines", `createElement("div", undefined, [createElement("a"), createElement("span")]);`, nil},
		{"a destructured react import reports", "import {createElement} from 'react';\ncreateElement(\"div\", undefined, [createElement(\"a\"), createElement(\"span\")]);", []string{"inlineElement"}},
		{"a third argument that is not an array yields no children", `React.createElement("div", undefined, React.createElement("a"));`, nil},
		{"only two arguments yields no children", `React.createElement("div", undefined);`, nil},
		// The asymmetry: the outer gate is the only place the callee is checked, so a foreign inner
		// call still counts as inline once the outer call has qualified.
		{"foreign inner calls still count as inline", `React.createElement("div", undefined, [Whatever.createElement("a"), Other.createElement("span")]);`, []string{"inlineElement"}},
		{"an inner call whose tag is an identifier is not inline", `React.createElement("div", undefined, [React.createElement(Foo), React.createElement("a")]);`, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoAdjacentInlineElementsAnchorsOnTheContainer asserts where the finding points.
//
// The corpus asserts message ids only. Upstream reports on the CONTAINER rather than on either of
// the two adjacent children, which is a real choice: the reader is being told about a relationship,
// and neither child is wrong on its own.
func TestNoAdjacentInlineElementsAnchorsOnTheContainer(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"the jsx container, not either child", `<div><a></a><a></a></div>;`, `<div><a></a><a></a></div>`},
		{"the whole createElement call", `React.createElement("div", undefined, [React.createElement("a"), React.createElement("span")]);`, `React.createElement("div", undefined, [React.createElement("a"), React.createElement("span")])`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "inlineElement")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			got := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if got != testCase.wantText {
				t.Errorf("finding spans %q, want %q", got, testCase.wantText)
			}
		})
	}
}

// TestNoAdjacentInlineElementsRequiresTheTypedHarness pins the checker declaration.
//
// The bare-identifier branch of the createElement resolution asks the checker, so under the plain
// harness that arm guards and goes silent while the JSX arm keeps reporting. A later revert of
// `NeedsTypeChecker` fails here rather than producing a vacuous green.
func TestNoAdjacentInlineElementsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "import {createElement} from 'react';\ncreateElement(\"div\", undefined, [createElement(\"a\"), createElement(\"span\")]);"

	untyped := rule_testing.Run(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, source)
	rule_testing.ExpectClean(t, untyped)

	typed := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, source)
	rule_testing.ExpectFindings(t, typed, "inlineElement")

	jsxUntyped := rule_testing.Run(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, `<div><a></a><a></a></div>;`)
	rule_testing.ExpectFindings(t, jsxUntyped, "inlineElement")
}

// TestNoAdjacentInlineElementsMessageExplainsTheRendering asserts the rendered text.
//
// The rule builds no message with a format verb, so nothing interpolates. What is worth pinning is
// that the description explains what the reader sees on the page rather than restating the rule
// name. Asserted against a literal typed here rather than against the rule's own constant, so both
// sides cannot move together under mutation.
func TestNoAdjacentInlineElementsMessageExplainsTheRendering(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoAdjacentInlineElements, noAdjacentInlineElementsFile, `<div><a></a><a></a></div>;`)
	rule_testing.ExpectFindings(t, result, "inlineElement")
	message := result.Diagnostics[0].Message
	if message.Id != "inlineElement" {
		t.Errorf("message id is %q", message.Id)
	}
	if !strings.HasPrefix(message.Description, "These two inline elements sit directly against each other") {
		t.Errorf("description does not open on the rendering: %q", message.Description)
	}
}
