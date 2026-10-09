package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's, extracted by stubbing its RuleTester and capturing what the test file
// hands it, so no case was retyped and no escape was typed by hand. All 47 cases were replayed
// against the installed build first, INCLUDING running its fixer and comparing the rewritten
// source against each case's own `output`, and every one reproduced.
const selfClosingCompFile = "/repository/source/SelfClosingComp.tsx"

func TestSelfClosingCompStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		options          SelfClosingCompOptions
	}{
		{"upstream valid 0", "var HelloJohn = <Hello name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 1", "var HelloJohn = <Hello.Compound name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 2", "var Profile = <Hello name=\"John\"><img src=\"picture.png\" /></Hello>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 3", "var Profile = <Hello.Compound name=\"John\"><img src=\"picture.png\" /></Hello.Compound>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 4", "\n        <Hello>\n          <Hello name=\"John\" />\n        </Hello>\n      ", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 5", "\n        <Hello.Compound>\n          <Hello.Compound name=\"John\" />\n        </Hello.Compound>\n      ", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 6", "var HelloJohn = <Hello name=\"John\"> </Hello>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 7", "var HelloJohn = <Hello.Compound name=\"John\"> </Hello.Compound>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 8", "var HelloJohn = <Hello name=\"John\">        </Hello>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 9", "var HelloJohn = <Hello.Compound name=\"John\">        </Hello.Compound>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 10", "var HelloJohn = <div>&nbsp;</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 11", "var HelloJohn = <div>{'\u00a0'}</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 12", "var HelloJohn = <Hello name=\"John\">&nbsp;</Hello>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 13", "var HelloJohn = <Hello.Compound name=\"John\">&nbsp;</Hello.Compound>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 14", "var HelloJohn = <Hello name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 15", "var HelloJohn = <Hello.Compound name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 16", "var Profile = <Hello name=\"John\"><img src=\"picture.png\" /></Hello>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 17", "var Profile = <Hello.Compound name=\"John\"><img src=\"picture.png\" /></Hello.Compound>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 18", "\n        <Hello>\n          <Hello name=\"John\" />\n        </Hello>\n      ", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 19", "\n        <Hello.Compound>\n          <Hello.Compound name=\"John\" />\n        </Hello.Compound>\n      ", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 20", "var HelloJohn = <div> </div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 21", "var HelloJohn = <div>        </div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 22", "var HelloJohn = <div>&nbsp;</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 23", "var HelloJohn = <div>{'\u00a0'}</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 24", "var HelloJohn = <Hello name=\"John\">&nbsp;</Hello>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 25", "var HelloJohn = <Hello.Compound name=\"John\">&nbsp;</Hello.Compound>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 26", "var HelloJohn = <Hello name=\"John\"></Hello>;", SelfClosingCompOptions{Component: false, Html: true}},
		{"upstream valid 27", "var HelloJohn = <Hello.Compound name=\"John\"></Hello.Compound>;", SelfClosingCompOptions{Component: false, Html: true}},
		{"upstream valid 28", "var HelloJohn = <Hello name=\"John\">\n</Hello>;", SelfClosingCompOptions{Component: false, Html: true}},
		{"upstream valid 29", "var HelloJohn = <Hello.Compound name=\"John\">\n</Hello.Compound>;", SelfClosingCompOptions{Component: false, Html: true}},
		{"upstream valid 30", "var HelloJohn = <Hello name=\"John\"> </Hello>;", SelfClosingCompOptions{Component: false, Html: true}},
		{"upstream valid 31", "var HelloJohn = <Hello.Compound name=\"John\"> </Hello.Compound>;", SelfClosingCompOptions{Component: false, Html: true}},
		{"upstream valid 32", "var contentContainer = <div className=\"content\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 33", "var contentContainer = <div className=\"content\"><img src=\"picture.png\" /></div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream valid 34", "\n        <div>\n          <div className=\"content\" />\n        </div>\n      ", SelfClosingCompOptions{Component: true, Html: true}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, SelfClosingComp, selfClosingCompFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestSelfClosingCompFires asserts the finding AND the repair. Every invalid case upstream ships
// carries an `output`, which is the exact text its fixer must write, and those are the highest
// value artifact in a fixable rule's corpus: a fixer that repairs the right span with the wrong
// text passes any message-id assertion.
func TestSelfClosingCompFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText, wantFixed string
		options                     SelfClosingCompOptions
	}{
		{"upstream invalid 0", "var contentContainer = <div className=\"content\"></div>;", "var contentContainer = <div className=\"content\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 1", "var contentContainer = <div className=\"content\"></div>;", "var contentContainer = <div className=\"content\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 2", "var HelloJohn = <Hello name=\"John\"></Hello>;", "var HelloJohn = <Hello name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 3", "var CompoundHelloJohn = <Hello.Compound name=\"John\"></Hello.Compound>;", "var CompoundHelloJohn = <Hello.Compound name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 4", "var HelloJohn = <Hello name=\"John\">\n</Hello>;", "var HelloJohn = <Hello name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 5", "var HelloJohn = <Hello.Compound name=\"John\">\n</Hello.Compound>;", "var HelloJohn = <Hello.Compound name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 6", "var HelloJohn = <Hello name=\"John\"></Hello>;", "var HelloJohn = <Hello name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 7", "var HelloJohn = <Hello.Compound name=\"John\"></Hello.Compound>;", "var HelloJohn = <Hello.Compound name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 8", "var HelloJohn = <Hello name=\"John\">\n</Hello>;", "var HelloJohn = <Hello name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 9", "var HelloJohn = <Hello.Compound name=\"John\">\n</Hello.Compound>;", "var HelloJohn = <Hello.Compound name=\"John\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 10", "var contentContainer = <div className=\"content\"></div>;", "var contentContainer = <div className=\"content\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"upstream invalid 11", "var contentContainer = <div className=\"content\">\n</div>;", "var contentContainer = <div className=\"content\" />;", SelfClosingCompOptions{Component: true, Html: true}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, SelfClosingComp, selfClosingCompFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, "notSelfClosing")
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestSelfClosingCompMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite is a differential table.
//
// Every row was run against the installed `eslint-plugin-react` build on 2026-08-27, and both the
// finding count AND the rewritten source are what that build produced. The rows exist because
// upstream's corpus is silent on all of them, and they pin four things it cannot see:
//
//	which half a tag falls in    the split is the regular expression `/^[a-z]/` over the rendered
//	                             element type, so `<foo.bar>` is html, `<my-element>` is html,
//	                             `<_foo>` is a component, `<this.Foo>` is html, and `<SVG:rect>` is
//	                             in NEITHER half and therefore always silent
//	the newline requirement      a single space on one line stays, and the same whitespace with a
//	                             newline in it is empty
//	the non-breaking space       newline, U+00A0, newline is NOT empty, because upstream's strip
//	                             spares that character. The fix would otherwise delete it.
//	what the fix preserves       attributes and a multiline opening tag survive the repair
//
// The non-breaking space rows were constructed by code point rather than typed, after a shell path
// silently cooked one into a plain space and produced a confident wrong measurement that disagreed
// with a Go probe of the same input.
func TestSelfClosingCompMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantFindings int
		wantFixed    string
		options      SelfClosingCompOptions
	}{
		{"div empty", "var a = <div></div>;", 1, "var a = <div />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"Foo empty", "var a = <Foo></Foo>;", 1, "var a = <Foo />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"Foo.Bar", "var a = <Foo.Bar></Foo.Bar>;", 1, "var a = <Foo.Bar />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"foo.bar is html", "var a = <foo.bar></foo.bar>;", 1, "var a = <foo.bar />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"Foo.Bar.Baz leftmost decides", "var a = <Foo.Bar.Baz></Foo.Bar.Baz>;", 1, "var a = <Foo.Bar.Baz />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"foo.Bar is html, judged by the leftmost segment", "var a = <foo.Bar></foo.Bar>;", 0, "var a = <foo.Bar></foo.Bar>;", SelfClosingCompOptions{Component: true, Html: false}},
		{"foo.Bar reports under component false", "var a = <foo.Bar></foo.Bar>;", 1, "var a = <foo.Bar />;", SelfClosingCompOptions{Component: false, Html: true}},
		{"Foo.bar is a component, judged by the leftmost segment", "var a = <Foo.bar></Foo.bar>;", 1, "var a = <Foo.bar />;", SelfClosingCompOptions{Component: true, Html: false}},
		{"Foo.bar is silent under component false", "var a = <Foo.bar></Foo.bar>;", 0, "var a = <Foo.bar></Foo.bar>;", SelfClosingCompOptions{Component: false, Html: true}},
		{"three levels: Foo.bar.Baz follows the leftmost", "var a = <Foo.bar.Baz></Foo.bar.Baz>;", 1, "var a = <Foo.bar.Baz />;", SelfClosingCompOptions{Component: true, Html: false}},
		{"three levels: foo.Bar.Baz follows the leftmost", "var a = <foo.Bar.Baz></foo.Bar.Baz>;", 0, "var a = <foo.Bar.Baz></foo.Bar.Baz>;", SelfClosingCompOptions{Component: true, Html: false}},
		{"three levels: foo.Bar.Baz reports under component false", "var a = <foo.Bar.Baz></foo.Bar.Baz>;", 1, "var a = <foo.Bar.Baz />;", SelfClosingCompOptions{Component: false, Html: true}},
		{"dashed custom element is html", "var a = <my-element></my-element>;", 1, "var a = <my-element />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"namespaced svg:rect is html", "var a = <svg:rect></svg:rect>;", 1, "var a = <svg:rect />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"namespaced SVG:rect is neither half", "var a = <SVG:rect></SVG:rect>;", 0, "var a = <SVG:rect></SVG:rect>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"underscore is a component", "var a = <_foo></_foo>;", 1, "var a = <_foo />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"this.Foo renders as this", "var a = <this.Foo></this.Foo>;", 1, "var a = <this.Foo />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"text child", "var a = <div>x</div>;", 0, "var a = <div>x</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"single space on one line stays", "var a = <div> </div>;", 0, "var a = <div> </div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"spaces on one line stay", "var a = <div>   </div>;", 0, "var a = <div>   </div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"newline alone is empty", "var a = <div>\n</div>;", 1, "var a = <div />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"newline and spaces is empty", "var a = <div>\n  \n</div>;", 1, "var a = <div />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"newline and tab is empty", "var a = <div>\n\t\n</div>;", 1, "var a = <div />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"newline and NBSP is NOT empty", "var a = <div>\n\u00a0\n</div>;", 0, "var a = <div>\n\u00a0\n</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"NBSP alone is not empty", "var a = <div>\u00a0</div>;", 0, "var a = <div>\u00a0</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"expression child", "var a = <div>{a}</div>;", 0, "var a = <div>{a}</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"comment child", "var a = <div>{/* c */}</div>;", 0, "var a = <div>{/* c */}</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"two children", "var a = <div>{a}{b}</div>;", 0, "var a = <div>{a}{b}</div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"already self closing", "var a = <div />;", 0, "var a = <div />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"nested empty reports inner", "var a = <div><span></span></div>;", 1, "var a = <div><span /></div>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"attributes are preserved by the fix", "var a = <div className=\"x\" id=\"y\"></div>;", 1, "var a = <div className=\"x\" id=\"y\" />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"multiline opening tag", "var a = <div\n  className=\"x\"\n></div>;", 1, "var a = <div\n  className=\"x\"\n />;", SelfClosingCompOptions{Component: true, Html: true}},
		{"fragment is not an element", "var a = <></>;", 0, "var a = <></>;", SelfClosingCompOptions{Component: true, Html: true}},
		{"div under component false", "var a = <div></div>;", 1, "var a = <div />;", SelfClosingCompOptions{Component: false, Html: true}},
		{"Foo under component false", "var a = <Foo></Foo>;", 0, "var a = <Foo></Foo>;", SelfClosingCompOptions{Component: false, Html: true}},
		{"div under html false", "var a = <div></div>;", 0, "var a = <div></div>;", SelfClosingCompOptions{Component: true, Html: false}},
		{"Foo under html false", "var a = <Foo></Foo>;", 1, "var a = <Foo />;", SelfClosingCompOptions{Component: true, Html: false}},
		{"div under both false", "var a = <div></div>;", 0, "var a = <div></div>;", SelfClosingCompOptions{Component: false, Html: false}},
		{"Foo under both false", "var a = <Foo></Foo>;", 0, "var a = <Foo></Foo>;", SelfClosingCompOptions{Component: false, Html: false}},
		{"svg:rect under component false", "var a = <svg:rect></svg:rect>;", 1, "var a = <svg:rect />;", SelfClosingCompOptions{Component: false, Html: true}},
		{"svg:rect under html false", "var a = <svg:rect></svg:rect>;", 0, "var a = <svg:rect></svg:rect>;", SelfClosingCompOptions{Component: true, Html: false}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, SelfClosingComp, selfClosingCompFile,
				testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != testCase.wantFindings {
				t.Fatalf("installed build reports %d findings, this rule reports %d",
					testCase.wantFindings, len(result.Diagnostics))
			}
			// The repair is asserted on every REPORTING row. `ExpectFixedSource` refuses a result
			// carrying no fixes rather than treating it as an unchanged file, so the silent rows
			// assert their zero count above and nothing more; their `wantFixed` is the input, which
			// is what an unapplied fix would leave.
			if testCase.wantFindings > 0 {
				rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
			}
		})
	}
}

// TestSelfClosingCompDoesNotPanicOnJsxText pins a crash shape that no findings assertion can see.
//
// `ast.Node.Text()` PANICS on a JsxText node with "Unhandled case in Node.Text", and this rule has
// to read exactly that node to decide whether a whitespace-only child counts as empty. The first
// draft called `Text()` and took the whole test binary down on upstream's own passing case 6.
//
// The cost of getting this wrong in production is larger than one rule: the walk recovers per FILE
// rather than per rule, so one such call loses every rule's findings in that file while the run
// still prints a plausible summary. The field is read directly instead.
//
// Every shape below reaches the text-reading path. A findings assertion would pass vacuously if the
// rule crashed, so this asserts by running at all.
func TestSelfClosingCompDoesNotPanicOnJsxText(t *testing.T) {
	t.Parallel()

	sources := []string{
		"var a = <div>\n</div>;",
		"var a = <div> </div>;",
		"var a = <div>text</div>;",
		"var a = <div>\n  \n</div>;",
		"var a = <div>{expression}text</div>;",
		"var a = <Foo>\n</Foo>;",
	}
	for _, sourceText := range sources {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			// Reaching the next line at all is the assertion.
			rule_testing.RunWithOptions(t, SelfClosingComp, selfClosingCompFile, sourceText,
				DefaultSelfClosingCompOptions())
		})
	}
}

// TestDecodeSelfClosingCompOptions routes configuration through the rule's own decoder.
//
// The tables above build the options struct directly, which leaves the decoder entirely untested,
// and a mutation sweep confirmed that: inverting both of its nil checks survived every fixture in
// this file. The decoder is where this rule's most dangerous line lives, because BOTH options
// default to true. A generic `rule.DecodeOptionsInto` would hand back a zero-value struct on absent
// input, both halves false, and the rule would silently report nothing while every struct-built
// fixture kept passing.
//
// The empty-input row is the one that matters most: a rule configured as a bare `"error"` reaches
// the decoder with no bytes at all.
func TestDecodeSelfClosingCompOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		raw           string
		wantComponent bool
		wantHtml      bool
		wantErr       bool
	}{
		{name: "empty input is the bare error configuration", raw: "", wantComponent: true, wantHtml: true},
		{name: "empty object keeps both defaults", raw: `{}`, wantComponent: true, wantHtml: true},
		{name: "component false alone", raw: `{"component": false}`, wantComponent: false, wantHtml: true},
		{name: "html false alone", raw: `{"html": false}`, wantComponent: true, wantHtml: false},
		{name: "both false", raw: `{"component": false, "html": false}`, wantComponent: false, wantHtml: false},
		{name: "explicit true is not distinguishable from absent, and should not be", raw: `{"component": true, "html": true}`, wantComponent: true, wantHtml: true},
		{name: "wrong type is an error", raw: `{"component": "yes"}`, wantErr: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeSelfClosingCompOptions([]byte(testCase.raw))
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("want an error, got %#v", decoded)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			options, isOptions := decoded.(SelfClosingCompOptions)
			if !isOptions {
				t.Fatalf("decoder returned %T, want SelfClosingCompOptions", decoded)
			}
			if options.Component != testCase.wantComponent {
				t.Errorf("Component is %v, want %v", options.Component, testCase.wantComponent)
			}
			if options.Html != testCase.wantHtml {
				t.Errorf("Html is %v, want %v", options.Html, testCase.wantHtml)
			}
		})
	}
}

// TestSelfClosingCompNilOptionsFallsBackToBothHalvesOn bypasses the decoder entirely.
//
// A rule configured as a bare `"error"` can reach `Run` with nil options rather than a decoded
// struct, and `options.(T)` on nil yields the ZERO value, which for this rule means both halves off
// and a rule that reports nothing. The explicit fallback is what prevents that, and nothing else in
// this file would notice if it were removed.
func TestSelfClosingCompNilOptionsFallsBackToBothHalvesOn(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{"var a = <div></div>;", "var a = <Foo></Foo>;"} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, SelfClosingComp, selfClosingCompFile, sourceText, nil)
			rule_testing.ExpectFindings(t, result, "notSelfClosing")
		})
	}
}

// TestSelfClosingCompHasNoFileSuffixGate pins the absence of the gate three siblings in this
// package carry.
//
// Those siblings were ported from oxc, which gates on the file being read as JSX. This rule's
// authority has no filename condition anywhere.
//
// # Why `.ts` is absent from this list, and why that is not the gate
//
// Measured on 2026-08-27, the same source under four suffixes: `.tsx`, `.jsx` and `.js` all report,
// and `.ts` reports nothing. That zero is the PARSER rather than a rule gate: in a `.ts` file
// `<div>` is a type assertion and no JSX node is produced at all, so there is nothing for any JSX
// rule to see. The `.js` row is what proves the absence of a gate, since an oxc-style
// `isJsxFileName` check would decline it while the parser happily produces JSX there. Three other
// rules in this package carried exactly that check until 4cff5fe, 206b628 and 9dbd234 removed it.
//
// This distinction matters because the two zeros look identical from a findings assertion, and
// asserting a finding on `.ts` would be asserting something false about the parser.
func TestSelfClosingCompHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	sourceText := "var a = <div></div>;\n"
	for _, fileName := range []string{
		"/repository/source/Suffix.tsx",
		"/repository/source/Suffix.jsx",
		"/repository/source/Suffix.js",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, SelfClosingComp, fileName, sourceText,
				DefaultSelfClosingCompOptions())
			rule_testing.ExpectFindings(t, result, "notSelfClosing")
		})
	}
}
