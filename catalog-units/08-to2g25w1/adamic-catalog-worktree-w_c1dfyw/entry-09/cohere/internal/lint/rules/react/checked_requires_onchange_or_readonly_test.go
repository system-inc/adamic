package react

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// checkedRequiresFile is where the fixtures pretend to live.
//
// A `.tsx` name because most cases are JSX and need a parser that reads it. It is NOT load-bearing:
// this rule has no file gate, and `TestCheckedRequiresHasNoFileGate` pins that by reporting on a
// `.ts` file too. Three shipped react rules in this package carry a `.tsx`-only gate that is oxc
// residue rather than upstream behavior (task 2abaqvt); this rule deliberately does not, because
// upstream has no such gate anywhere.
const checkedRequiresFile = "/repository/source/CheckedRequires.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/checked-requires-onchange-or-readonly.js`
// holds 24 valid and 11 invalid cases. Every string below was pulled out of that file by loading it
// with a stubbed `RuleTester` and serializing the captured object to JSON, then emitted into this
// table by a generator. No case was typed by hand, so no escape sequence passed through a shell or
// an editor on the way here.
//
// All 35 were then run against the installed build, version 7.37.5, through the ESLint Linter API,
// and its verdict agreed with the clone on every one: 24 clean and 11 reporting with the stated
// counts and message ids in the stated order. The two authorities are one for this rule.
//
// # The options column is RAW JSON, on purpose
//
// Every fixture routes through `DecodeCheckedRequiresOnChangeOrReadOnlyOptions`, the same function
// the config layer calls. Handing `RunWithOptions` a struct built by hand would leave the decoder
// untested, and the decoder is the one line with no upstream counterpart: it has to answer the
// defaults on EMPTY input where the generic helper would error. An empty string in that column
// means the rule was configured as a bare severity, which is what hands a real rule nil options,
// and 27 of the 35 cases take that path.
//
// The wire shape is the BARE object, not upstream's `[{...}]` array. cohere's config layer unwraps
// the `[severity, options]` tuple before dispatch, so the array wrapper the corpus writes is gone
// by the time a decoder sees anything.

// TestCheckedRequiresFires asserts the ids, the count, AND the order.
//
// Order is asserted rather than incidental: three of the eleven reporting cases produce two findings
// on one node, and upstream reports `exclusiveCheckedAttribute` before `missingProperty` because it
// asks the exclusive question first. A rule that asked them the other way round would satisfy a
// set comparison and diverge from the corpus, so `ExpectFindings` taking an ordered list is doing
// real work here.
func TestCheckedRequiresFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
	}{
		{"upstream invalid-0", "<input type=\"radio\" checked />", "", []string{"missingProperty"}},
		{"upstream invalid-1", "<input type=\"radio\" checked={true} />", "", []string{"missingProperty"}},
		{"upstream invalid-2", "<input type=\"checkbox\" checked />", "", []string{"missingProperty"}},
		{"upstream invalid-3", "<input type=\"checkbox\" checked={true} />", "", []string{"missingProperty"}},
		{"upstream invalid-4", "<input type=\"checkbox\" checked={condition ? true : false} />", "", []string{"missingProperty"}},
		{"upstream invalid-5", "<input type=\"checkbox\" checked defaultChecked />", "", []string{"exclusiveCheckedAttribute", "missingProperty"}},
		{"upstream invalid-6", "React.createElement(\"input\", { checked: false })", "", []string{"missingProperty"}},
		{"upstream invalid-7", "React.createElement(\"input\", { checked: true, defaultChecked: true })", "", []string{"exclusiveCheckedAttribute", "missingProperty"}},
		{"upstream invalid-8", "<input type=\"checkbox\" checked defaultChecked />", "{\"ignoreMissingProperties\":true}", []string{"exclusiveCheckedAttribute"}},
		{"upstream invalid-9", "<input type=\"checkbox\" checked defaultChecked />", "{\"ignoreExclusiveCheckedAttribute\":true}", []string{"missingProperty"}},
		{"upstream invalid-10", "<input type=\"checkbox\" checked defaultChecked />", "{\"ignoreMissingProperties\":false,\"ignoreExclusiveCheckedAttribute\":false}", []string{"exclusiveCheckedAttribute", "missingProperty"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runCheckedRequires(t, testCase.sourceText, testCase.rawOptions)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestCheckedRequiresStaysSilent runs upstream's clean cases.
//
// These are the false positives upstream already thought about, and two of them are the ones a port
// is most likely to get wrong: `React.createElement('input')` with no second argument at all, and
// `(()=>{})()`, an immediately-invoked arrow whose callee is neither an identifier nor a member
// access, which is exactly the shape that panics a rule reaching for a callee name without checking
// the kind first.
func TestCheckedRequiresStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"upstream valid-0", "<input type=\"checkbox\" />", ""},
		{"upstream valid-1", "<input type=\"checkbox\" onChange={noop} />", ""},
		{"upstream valid-2", "<input type=\"checkbox\" readOnly />", ""},
		{"upstream valid-3", "<input type=\"checkbox\" checked onChange={noop} />", ""},
		{"upstream valid-4", "<input type=\"checkbox\" checked={true} onChange={noop} />", ""},
		{"upstream valid-5", "<input type=\"checkbox\" checked={false} onChange={noop} />", ""},
		{"upstream valid-6", "<input type=\"checkbox\" checked readOnly />", ""},
		{"upstream valid-7", "<input type=\"checkbox\" checked={true} readOnly />", ""},
		{"upstream valid-8", "<input type=\"checkbox\" checked={false} readOnly />", ""},
		{"upstream valid-9", "<input type=\"checkbox\" defaultChecked />", ""},
		{"upstream valid-10", "React.createElement('input')", ""},
		{"upstream valid-11", "React.createElement('input', { checked: true, onChange: noop })", ""},
		{"upstream valid-12", "React.createElement('input', { checked: false, onChange: noop })", ""},
		{"upstream valid-13", "React.createElement('input', { checked: true, readOnly: true })", ""},
		{"upstream valid-14", "React.createElement('input', { checked: true, onChange: noop, readOnly: true })", ""},
		{"upstream valid-15", "React.createElement('input', { checked: foo, onChange: noop, readOnly: true })", ""},
		{"upstream valid-16", "<input type=\"checkbox\" checked />", "{\"ignoreMissingProperties\":true}"},
		{"upstream valid-17", "<input type=\"checkbox\" checked={true} />", "{\"ignoreMissingProperties\":true}"},
		{"upstream valid-18", "<input type=\"checkbox\" onChange={noop} checked defaultChecked />", "{\"ignoreExclusiveCheckedAttribute\":true}"},
		{"upstream valid-19", "<input type=\"checkbox\" onChange={noop} checked={true} defaultChecked />", "{\"ignoreExclusiveCheckedAttribute\":true}"},
		{"upstream valid-20", "<input type=\"checkbox\" onChange={noop} checked defaultChecked />", "{\"ignoreMissingProperties\":true,\"ignoreExclusiveCheckedAttribute\":true}"},
		{"upstream valid-21", "<span/>", ""},
		{"upstream valid-22", "React.createElement('span')", ""},
		{"upstream valid-23", "(()=>{})()", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runCheckedRequires(t, testCase.sourceText, testCase.rawOptions))
		})
	}
}

// runCheckedRequires drives the rule through its own decoder, on a typed program.
//
// `RunTyped` rather than `Run` because the rule declares `NeedsTypeChecker`: the bare
// `createElement(...)` arm asks the checker which binding that name is. The plain harness hands a
// rule a nil checker, and the failure mode there is SILENCE rather than a crash. Every StaysSilent
// case would pass vacuously, so `TestCheckedRequiresNeedsTheTypedHarness` pins the requirement
// with a case that only the typed path can report.
func runCheckedRequires(t *testing.T, sourceText string, rawOptions string) rule_testing.Result {
	t.Helper()
	decoded, err := DecodeCheckedRequiresOnChangeOrReadOnlyOptions([]byte(rawOptions))
	if err != nil {
		t.Fatalf("decoding %q: %v", rawOptions, err)
	}
	return rule_testing.RunTypedWithOptions(t, CheckedRequiresOnChangeOrReadOnly,
		checkedRequiresFile, sourceText, decoded)
}

// TestCheckedRequiresResolvesEveryPragmaShape covers the four bindings upstream accepts.
//
// The corpus writes NONE of these: every one of its `createElement` cases is namespaced
// `React.createElement`, so the entire bare-identifier arm, and the checker declaration that arm
// forces, is invisible to the imported fixtures. These rows exist because the reference
// implementation has a whole helper for them, and each verdict below was measured on the installed
// build before it was written here rather than read off the source.
//
// The three declining rows are the ones that make the arm a discrimination rather than a rubber
// stamp. A bare `createElement` with nothing declaring it is silent; the same call importing from
// `preact` is silent, because upstream compares the module against the lowercased pragma; and a
// locally-defined function of that name is silent, which is the control separating "resolved to the
// pragma" from "resolved to anything at all".
func TestCheckedRequiresResolvesEveryPragmaShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"import from react",
			"import { createElement } from 'react';\ncreateElement('input', { checked: true });\n",
			[]string{"missingProperty"},
		},
		{
			"import from preact declines",
			"import { createElement } from 'preact';\ncreateElement('input', { checked: true });\n",
			nil,
		},
		{
			"destructured from the pragma",
			"declare const React: any;\nconst { createElement } = React;\ncreateElement('input', { checked: true });\n",
			[]string{"missingProperty"},
		},
		{
			"destructured from require",
			"declare function require(name: string): any;\nconst { createElement } = require('react');\ncreateElement('input', { checked: true });\n",
			[]string{"missingProperty"},
		},
		{
			"aliased off the pragma",
			"declare const React: any;\nconst createElement = React.createElement;\ncreateElement('input', { checked: true });\n",
			[]string{"missingProperty"},
		},
		{
			"aliased off require",
			"declare function require(name: string): any;\nconst createElement = require('react').createElement;\ncreateElement('input', { checked: true });\n",
			[]string{"missingProperty"},
		},
		{
			"undeclared bare call declines",
			"declare const createElementSomethingElse: unknown;\ncreateElement('input', { checked: true });\n",
			nil,
		},
		{
			"local function of the same name declines",
			"function createElement(tag: string, properties: object) { return tag; }\ncreateElement('input', { checked: true });\n",
			nil,
		},
		// The five below pin the INITIALIZER side of the binding, which the accepting rows above
		// cannot see on their own: each accepting row proves the shape is reached, and only a
		// matching declining row proves the pragma test inside it decides anything. Three separate
		// mutants survived the whole suite without them, one per line these cover: the identifier
		// test in `initializerComesFromPragma`, the callee-name test in `isPragmaRequireCall`, and
		// its module-name test. All five measured silent on the installed build.
		{
			"destructured from a different namespace declines",
			"declare const Preact: any;\nconst { createElement } = Preact;\ncreateElement('input', { checked: true });\n",
			nil,
		},
		{
			"destructured from a different module declines",
			"declare function require(name: string): any;\nconst { createElement } = require('preact');\ncreateElement('input', { checked: true });\n",
			nil,
		},
		{
			"destructured from a call that is not require declines",
			"declare function load(name: string): any;\nconst { createElement } = load('react');\ncreateElement('input', { checked: true });\n",
			nil,
		},
		{
			"aliased off a different namespace declines",
			"declare const Preact: any;\nconst createElement = Preact.createElement;\ncreateElement('input', { checked: true });\n",
			nil,
		},
		{
			"aliased off a different module declines",
			"declare function require(name: string): any;\nconst createElement = require('preact').createElement;\ncreateElement('input', { checked: true });\n",
			nil,
		},
		// The three below cover DECLARATION MERGING, where the symbol carries more than one
		// declaration and the import is not necessarily first. A mutant replacing the loop in
		// `bindsToPragmaImport` with the first declaration only SURVIVED the rest of this suite,
		// and probing the checker showed why: an ambient declaration written above the import puts
		// the import at index 1, which is the index-zero trap the port brief names for six other
		// rules here.
		//
		// The third row is a stated DIVERGENCE rather than a match. Upstream resolves the LATEST
		// definition, so an import followed by an ambient declaration is silent there; this loop
		// finds the import wherever it sits and reports. Measured on the installed build with the
		// TypeScript parser. The reasoning for reporting rather than reproducing is at
		// `bindsToPragmaImport`.
		{
			"an ambient declaration above the import still reports",
			"declare function createElement(a: string): void;\nimport { createElement } from 'react';\ncreateElement('input', { checked: true });\n",
			[]string{"missingProperty"},
		},
		{
			"an interface above the import still reports",
			"interface createElement { a: number }\nimport { createElement } from 'react';\ncreateElement('input', { checked: true });\n",
			[]string{"missingProperty"},
		},
		{
			"an ambient declaration BELOW the import reports, where upstream is silent",
			"import { createElement } from 'react';\ndeclare function createElement(a: string): void;\ncreateElement('input', { checked: true });\n",
			[]string{"missingProperty"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runCheckedRequires(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestCheckedRequiresMatchesUpstreamOnShapesTheCorpusOmits pins ten measured verdicts.
//
// Each row was run against the installed build, 7.37.5, before it was written, and each covers a
// discrimination the imported corpus never exercises. They are not invented cases: they are the
// answers to the questions reading the reference implementation raised.
//
// The four silent `createElement` spellings are the reason this rule does not call the shelf's
// `react.IsCreateElementCall`, which answers true on all four. The two silent key spellings are
// upstream reading `prop.key.name`, which a string literal and a computed key do not have. The
// property is unmistakably `checked` to a reader and invisible to the rule, and that is reproduced
// rather than improved.
func TestCheckedRequiresMatchesUpstreamOnShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a computed member on the pragma declines",
			"declare const React: any;\nReact['createElement']('input', { checked: true });\n",
			nil,
		},
		{
			"a different namespace declines",
			"declare const Preact: any;\nPreact.createElement('input', { checked: true });\n",
			nil,
		},
		{
			"document.createElement declines",
			"document.createElement('input');\n",
			nil,
		},
		{
			"a parenthesized pragma callee still reports",
			"declare const React: any;\n(React.createElement)('input', { checked: true });\n",
			[]string{"missingProperty"},
		},
		{
			"a string-literal key is invisible",
			"declare const React: any;\nReact.createElement('input', { 'checked': true });\n",
			nil,
		},
		{
			"a computed key is invisible",
			"declare const React: any;\nReact.createElement('input', { ['checked']: true });\n",
			nil,
		},
		{
			"a spread beside a literal checked still reports",
			"declare const React: any;\ndeclare const rest: object;\nReact.createElement('input', { checked: true, ...rest });\n",
			[]string{"missingProperty"},
		},
		{
			"a template-literal tag declines",
			"declare const React: any;\nReact.createElement(`input`, { checked: true });\n",
			nil,
		},
		{
			"attribute names are case sensitive",
			"const element = <input CHECKED />;\n",
			nil,
		},
		{
			"a lowercase readonly does not satisfy the rule",
			"declare const React: any;\nReact.createElement('input', { readonly: true, checked: true });\n",
			[]string{"missingProperty"},
		},
		// The next five pin the TAG gate, which upstream's corpus cannot see. Its only non-input
		// cases are `<span/>` and `React.createElement('span')`, neither of which writes `checked`,
		// so a rule that dropped the tag test entirely passes all 35 imported cases. A mutant
		// replacing `IsIntrinsicElementNamed(tagName, "input")` with a bare nil check SURVIVED the
		// whole imported set and these are what kill it. Every verdict below measured silent on the
		// installed build first.
		{
			"a component that is not input declines",
			"declare const Checkbox: any;\nconst element = <Checkbox checked />;\n",
			nil,
		},
		{
			"a different intrinsic element declines",
			"const element = <select checked />;\n",
			nil,
		},
		{
			"a member-expression tag declines",
			"declare const input: any;\nconst element = <input.Nested checked />;\n",
			nil,
		},
		{
			"createElement of a non-input tag declines",
			"declare const React: any;\nReact.createElement('select', { checked: true });\n",
			nil,
		},
		{
			"createElement of a component name declines",
			"declare const React: any;\nReact.createElement('Checkbox', { checked: true });\n",
			nil,
		},
		// The next two pin the PROPERTY name on the namespaced callee. The corpus writes only
		// `React.createElement`, so a rule accepting any property on the pragma passes all 35
		// imported cases; a mutant dropping that name test SURVIVED the whole set. Both verdicts
		// measured silent on the installed build.
		{
			"another React member declines",
			"declare const React: any;\nReact.cloneElement('input', { checked: true });\n",
			nil,
		},
		{
			"createFactory declines",
			"declare const React: any;\nReact.createFactory('input', { checked: true });\n",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runCheckedRequires(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestCheckedRequiresReportsOnTheWholeElement asserts WHERE each finding points.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule anchoring on the attribute
// rather than on the element passes a complete fixture pair while pointing somewhere the reader was
// never shown. Upstream reports on the `JSXOpeningElement` and on the `CallExpression`, which are
// the whole opening tag and the whole call. The expected text is sliced from the SOURCE THE HARNESS
// WROTE rather than from the literal above, because `RunTyped` trims the fixture and a span sliced
// from an untrimmed literal is off by one.
func TestCheckedRequiresReportsOnTheWholeElement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		{
			"self-closing element",
			"<input type=\"checkbox\" checked />",
			[]string{"<input type=\"checkbox\" checked />"},
		},
		{
			"opening element of a paired tag",
			"<input checked></input>",
			[]string{"<input checked>"},
		},
		{
			"both findings point at the same element",
			"<input checked defaultChecked />",
			[]string{"<input checked defaultChecked />", "<input checked defaultChecked />"},
		},
		{
			"the whole call, not the argument",
			"declare const React: any;\nReact.createElement('input', { checked: true });",
			[]string{"React.createElement('input', { checked: true })"},
		},
		{
			"a nested call reports on each call separately",
			"declare const React: any;\nReact.createElement('input', { checked: 1 }, React.createElement('input', { checked: 2 }));",
			[]string{
				"React.createElement('input', { checked: 1 }, React.createElement('input', { checked: 2 }))",
				"React.createElement('input', { checked: 2 })",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runCheckedRequires(t, testCase.sourceText, "")
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantTexts))
			}
			source := result.SourceFile.Text()
			for index, diagnostic := range result.Diagnostics {
				got := source[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.wantTexts[index] {
					t.Errorf("finding %d points at %q, want %q", index, got, testCase.wantTexts[index])
				}
			}
		})
	}
}

// TestCheckedRequiresMessagesReadAsWritten pins the rendered text.
//
// Neither message interpolates, so there is nothing for a format string to get wrong, but the
// Description is what a reader actually sees and a silent edit to it is otherwise invisible to
// every other test in this file. Asserted against literals typed here rather than against the
// rule's own constants, because comparing a diagnostic to the constant it was reported with is an
// equality both sides of which move together under mutation.
func TestCheckedRequiresMessagesReadAsWritten(t *testing.T) {
	t.Parallel()

	result := runCheckedRequires(t, "<input checked defaultChecked />", "")
	if len(result.Diagnostics) != 2 {
		t.Fatalf("got %d findings, want 2", len(result.Diagnostics))
	}

	if result.Diagnostics[0].Message.Id != "exclusiveCheckedAttribute" {
		t.Errorf("first id is %q", result.Diagnostics[0].Message.Id)
	}
	if !strings.HasPrefix(result.Diagnostics[0].Message.Description,
		"This `input` writes both `checked` and `defaultChecked`.") {
		t.Errorf("exclusive description reads %q", result.Diagnostics[0].Message.Description)
	}

	if result.Diagnostics[1].Message.Id != "missingProperty" {
		t.Errorf("second id is %q", result.Diagnostics[1].Message.Id)
	}
	if !strings.HasPrefix(result.Diagnostics[1].Message.Description,
		"This `input` is `checked` with no `onChange` and no `readOnly`.") {
		t.Errorf("missing description reads %q", result.Diagnostics[1].Message.Description)
	}
}

// TestCheckedRequiresDecodesItsOptions exercises the decoder the config layer calls.
//
// The empty-input row is the one with no upstream counterpart and the reason the decoder is
// hand-written: `rule.DecodeOptionsInto` ERRORS on empty input, and a rule configured as a bare
// `"error"` is handed exactly that. A rule whose decoder errored there would be handed nil options,
// `options.(T)` would yield the zero value, and because both defaults here happen to BE the zero
// value, the rule would still behave correctly while its decoder was broken. So this asserts the
// decoder rather than inferring it from a passing fixture.
func TestCheckedRequiresDecodesItsOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		raw                 string
		wantMissingIgnored  bool
		wantExclusiveIgnore bool
	}{
		{"empty input answers the defaults", "", false, false},
		{"an empty object answers the defaults", "{}", false, false},
		{"missing properties ignored", `{"ignoreMissingProperties":true}`, true, false},
		{"exclusive attribute ignored", `{"ignoreExclusiveCheckedAttribute":true}`, false, true},
		{"both ignored", `{"ignoreMissingProperties":true,"ignoreExclusiveCheckedAttribute":true}`, true, true},
		{"explicit false is still false", `{"ignoreMissingProperties":false}`, false, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeCheckedRequiresOnChangeOrReadOnlyOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.raw, err)
			}
			settings, ok := decoded.(CheckedRequiresOnChangeOrReadOnlyOptions)
			if !ok {
				t.Fatalf("decoded to %T", decoded)
			}
			if settings.IgnoreMissingProperties != testCase.wantMissingIgnored {
				t.Errorf("IgnoreMissingProperties is %v", settings.IgnoreMissingProperties)
			}
			if settings.IgnoreExclusiveCheckedAttribute != testCase.wantExclusiveIgnore {
				t.Errorf("IgnoreExclusiveCheckedAttribute is %v", settings.IgnoreExclusiveCheckedAttribute)
			}
		})
	}
}

// TestCheckedRequiresRejectsMalformedOptions asserts the decoder surfaces bad input.
func TestCheckedRequiresRejectsMalformedOptions(t *testing.T) {
	t.Parallel()

	if _, err := DecodeCheckedRequiresOnChangeOrReadOnlyOptions([]byte(`{"ignoreMissingProperties":`)); err == nil {
		t.Fatal("truncated JSON decoded without error")
	}
	if _, err := DecodeCheckedRequiresOnChangeOrReadOnlyOptions(json.RawMessage(`"error"`)); err == nil {
		t.Fatal("a string decoded into the options struct without error")
	}
}

// TestCheckedRequiresNeedsTheTypedHarness pins the checker declaration.
//
// The rule guards with `if ctx.TypeChecker == nil`, so on the untyped harness the bare-identifier
// arm goes silent rather than panicking, and silence is the more dangerous failure, because every
// StaysSilent fixture would pass vacuously while the arm did nothing. This asserts the difference
// directly: one source that reports under `RunTyped` and is silent under `Run`. A later revert of
// `NeedsTypeChecker` fails here rather than quietly halving the rule.
//
// The control matters as much as the subject. The JSX arm needs no checker at all, so it must
// report under BOTH harnesses; if it did not, this test would be measuring a broken harness rather
// than a real dependency.
func TestCheckedRequiresNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	const bareCall = "import { createElement } from 'react';\ncreateElement('input', { checked: true });\n"

	typed := rule_testing.RunTypedWithOptions(t, CheckedRequiresOnChangeOrReadOnly,
		checkedRequiresFile, bareCall, CheckedRequiresOnChangeOrReadOnlyOptions{})
	rule_testing.ExpectFindings(t, typed, "missingProperty")

	untyped := rule_testing.RunWithOptions(t, CheckedRequiresOnChangeOrReadOnly,
		checkedRequiresFile, bareCall, CheckedRequiresOnChangeOrReadOnlyOptions{})
	rule_testing.ExpectClean(t, untyped)

	const jsxOnly = "const element = <input checked />;\n"
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, CheckedRequiresOnChangeOrReadOnly,
		checkedRequiresFile, jsxOnly, CheckedRequiresOnChangeOrReadOnlyOptions{}), "missingProperty")
}

// TestCheckedRequiresHandlesNilOptions covers the path the fixtures cannot reach.
//
// Every other test here routes through the decoder, which always produces a struct. A rule can still
// be handed nil: `config.OptionsRegistry.Decode` turns a decoder error into nil for a non-required
// rule, and `options.(T)` on nil yields the zero value. Both defaults ARE the zero value for this
// rule, so nil must behave exactly like an absent option object, and this asserts it rather than
// assuming it. A rule whose defaults were not zero would be silently inverted on this path, which is
// a failure 21 of 21 imported cases could not see on a rule that hit it.
func TestCheckedRequiresHandlesNilOptions(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedWithOptions(t, CheckedRequiresOnChangeOrReadOnly,
		checkedRequiresFile, "<input checked defaultChecked />", nil)
	rule_testing.ExpectFindings(t, result, "exclusiveCheckedAttribute", "missingProperty")
}

// TestCheckedRequiresHasNoFileGate reports on a `.ts` file, not only on `.tsx`.
//
// Three shipped rules in this package USED to gate on `.tsx`/`.jsx` through `isJsxFileName`, which
// was oxc residue rather than upstream behavior: real ESLint reports on `.ts` too, measured by
// probing eslint-plugin-react with the same source under four suffixes, all four of which reported.
// Filed as task 2abaqvt and removed in 4cff5fe, 206b628 and 9dbd234, so no rule in this package
// gates on the suffix any more. This case still pins the absence, so nobody reintroduces one.
//
// `.ts` is the extension that matters and it is also the only one available. The typed harness
// writes a tsconfig including `**/*.ts` and `**/*.tsx` (`internal/rule_testing/program.go:30`), so a
// `.jsx` or `.mjsx` fixture is not in the program at all and the run fails with TS18003 rather than
// measuring the rule. That is a fact about the harness rather than about this rule, and the two
// extensions below are the whole reachable surface. The `createElement` arm is used rather than a
// JSX one because JSX in a `.ts` file is a syntax error, so it is the arm that can actually
// distinguish the two extensions.
func TestCheckedRequiresHasNoFileGate(t *testing.T) {
	t.Parallel()

	const source = "declare const React: any;\nReact.createElement('input', { checked: true });\n"
	for _, fileName := range []string{
		"/repository/source/CheckedRequires.tsx",
		"/repository/source/CheckedRequires.ts",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, CheckedRequiresOnChangeOrReadOnly,
				fileName, source, CheckedRequiresOnChangeOrReadOnlyOptions{})
			rule_testing.ExpectFindings(t, result, "missingProperty")
		})
	}
}

// TestCheckedRequiresSurvivesShapesThatWouldPanic drives the callee shapes with no name to read.
//
// The walk recovers per FILE rather than per rule, so one nil dereference in this rule takes the
// whole file away from every other rule in the tree. Upstream's own corpus carries exactly one of
// these, `(()=>{})()`, and the rest are ordinary code this rule will meet on the real tree: a
// `super()` call whose callee is a bare keyword, a dynamic `import()`, an optional call, and a
// tagged template. Each reaches the CallExpression listener with a callee that is neither an
// identifier nor a property access.
func TestCheckedRequiresSurvivesShapesThatWouldPanic(t *testing.T) {
	t.Parallel()

	sources := []string{
		"(()=>{})();\n",
		"class Base { constructor() {} }\nclass Derived extends Base { constructor() { super(); } }\n",
		"async function load() { await import('./other'); }\n",
		"declare const maybe: undefined | ((tag: string) => void);\nmaybe?.('input');\n",
		"declare function tag(strings: TemplateStringsArray): void;\ntag`input`;\n",
		"declare const factory: { make(): (tag: string) => void };\nfactory.make()('input');\n",
		"declare const React: any;\nReact.createElement();\n",
		"declare const React: any;\nReact.createElement('input');\n",
		"declare const React: any;\nReact.createElement('input', null);\n",
		"const element = <input {...{ checked: true }} />;\n",
		"const element = <input xlink:href=\"x\" checked readOnly />;\n",
	}

	for index, sourceText := range sources {
		t.Run(string(rune('a'+index)), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runCheckedRequires(t, sourceText, ""))
		})
	}
}
