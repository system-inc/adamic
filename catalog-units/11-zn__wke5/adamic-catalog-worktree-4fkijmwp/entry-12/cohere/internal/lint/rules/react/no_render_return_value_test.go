package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// renderReturnValueFile is where the fixtures pretend to live.
//
// A `.tsx` extension because upstream gates the whole rule on `source_type().is_jsx()` and almost
// every imported case contains JSX, which would not parse otherwise.
const renderReturnValueFile = "/repository/source/RenderReturnValue.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case in the two upstream tests below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/no_render_return_value.rs`: 9 pass, 8 fail, and the
// snapshot records 8 diagnostics from those 8 inputs, so one finding per input is measured rather
// than assumed. The extractor reported no discrepancy, one tester block, and no case carrying
// options.
//
// The imported corpus is weak in one specific way, which is why the invented cases below exist and
// why each says what measured it. Upstream's fail list covers the five parent kinds and nothing
// else, so a port implementing only the parent-kind check passes all eight fail cases and all nine
// pass cases while being silent on a whole second rule the upstream ships. That second rule is the
// arrow-expression-body scope check, and the input distinguishing it appears nowhere upstream.
func TestNoRenderReturnValueFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The eight upstream fail cases, verbatim.
		{"a variable declarator", "var Hello = ReactDOM.render(<div />, document.body);"},
		{"an object property", "\n                    var o = {\n                      inst: ReactDOM.render(<div />, document.body)\n                    };\n                  "},
		{"a return statement", "\n                    function render () {\n                      return ReactDOM.render(<div />, document.body)\n                    }\n                  "},
		{"an arrow expression body with parameters", "var render = (a, b) => ReactDOM.render(a, b)"},
		{"assignment to a member", "this.o = ReactDOM.render(<div />, document.body);"},
		{"assignment to a variable", "var v; v = ReactDOM.render(<div />, document.body);"},
		{"a variable declarator named inst", "var inst = ReactDOM.render(<div />, document.body);"},
		{"an arrow expression body with no parameters", "const render = () => ReactDOM.render(<div />, document.body)"},

		// Invented, each pinning a discrimination the imported corpus leaves unguarded. All were
		// measured on the release oxlint binary rather than reasoned from the source.
		//
		// The subscript spelling. oxc reads the property through `static_property_info()`, which
		// answers for a string key, so `ReactDOM['render']` is the same call. Measured: reports,
		// with a span of 17 bytes covering `ReactDOM['render']`. A port comparing dotted text only
		// is silent here and no upstream case says so.
		{"a subscripted property", "var q = ReactDOM['render'](<div />, x);"},
		// A compound assignment operator. oxc matches `AssignmentExpression` without inspecting the
		// operator, so `+=` counts as much as `=`. Measured: reports.
		{"a compound assignment", "var v; v += ReactDOM.render(<div />, x);"},
		// A destructuring declarator still is a declarator. Measured: reports.
		{"a destructuring declarator", "var { p } = ReactDOM.render(<div />, x);"},
		// A parameter default. Neither implementation reaches this through the arrow arm: the
		// default is spelled as an assignment and answers there instead, in oxc as an
		// `AssignmentPattern` and here as a binary expression with an equals token. Measured on
		// oxlint at one finding, and measured here at one, so the two agree by a different route.
		{"an arrow parameter default", "var a = (p = ReactDOM.render(<div />, x)) => 1;"},
		// A single-quasi template subscript. oxc's `static_property_info` reads one, at
		// `oxc_ast/src/ast_impl/js.rs:628`, so this is the same call written a third way. Measured
		// on oxlint: reports. This case exists because a sweep widening the accept set survived,
		// which is how the original narrower set was found to be a silent divergence.
		{"a templated property", "var t = ReactDOM[`render`](<div />, x);"},
		// A shadowed receiver. This parameter is not react-dom and the rule reports anyway, because
		// upstream matches the spelling rather than the binding. Measured on oxlint: reports. It is
		// the false positive a resolution-based port would decline, and it is kept because oxlint is
		// the gate this replaces.
		{"a shadowed receiver named ReactDOM", "function f(ReactDOM) { var c = ReactDOM.render(<div />, x); }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRenderReturnValue, renderReturnValueFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noRenderReturnValue")
		})
	}
}

// TestNoRenderReturnValueStaysSilent holds upstream's nine clean cases plus the ones that pin where
// this rule deliberately stops.
//
// The last four are the important ones and they are the reason this rule cannot be written as "is
// the return value used". Every one of them uses the return value in a way any reader would call a
// use, and upstream reports none of them, because the rule is a fixed list of parent kinds rather
// than a use analysis. Measured on oxlint: all silent.
func TestNoRenderReturnValueStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The nine upstream pass cases, verbatim. Two are byte-identical duplicates upstream, kept
		// as written rather than deduplicated.
		{"a bare call statement", "ReactDOM.render(<div />, document.body);"},
		{"a ref callback assigning a local", "\n                    let node;\n                    ReactDOM.render(<div ref={ref => node = ref}/>, document.body);\n                  "},
		{"a ref callback assigning a member", "ReactDOM.render(<div ref={ref => this.node = ref}/>, document.body);"},
		{"React.render rather than ReactDOM", "React.render(<div ref={ref => this.node = ref}/>, document.body);"},
		{"React.render rather than ReactDOM, again", "React.render(<div ref={ref => this.node = ref}/>, document.body);"},
		{"React.render assigned", "var foo = React.render(<div />, root);"},
		{"a bare render call assigned", "var foo = render(<div />, root)"},
		{"a misspelled property on a similar name", "var foo = ReactDom.renderder(<div />, root)"},
		{"an unrelated ReactDOM method", "export const foo = () => ({ destroy: ({ dom }) => { ReactDOM.unmountComponentAtNode(dom); } });"},

		// Invented. These four are genuine uses of the return value that upstream does not report,
		// and they are the whole answer to what this rule decides. A port that inferred its parent
		// set from "what counts as using a value" reports all four and passes every upstream case
		// while doing so. Each measured silent on oxlint at file scope.
		{"an argument position", "foo(ReactDOM.render(<div />, x));"},
		{"a property read off the result", "ReactDOM.render(<div />, x).foo;"},
		{"a condition", "if (ReactDOM.render(<div />, x)) {}"},
		{"a comparison", "ReactDOM.render(<div />, x) === null;"},
		// A class field looks like a declarator and is not one. oxc calls it a `PropertyDefinition`,
		// which is absent from its match list. Measured silent, against the declarator control
		// above which reports.
		{"a class field initializer", "class C { f = ReactDOM.render(<div />, x); }"},
		// The scope check stops at the nearest function boundary rather than walking to any
		// ancestor arrow. Measured silent, against the `return` form which reports.
		{"a function expression body inside an arrow", "var a = () => (function () { ReactDOM.render(<div />, x); });"},
		// A block-bodied arrow is not an expression-bodied one, so neither check reaches it.
		{"a block-bodied arrow", "var b = () => { ReactDOM.render(<div />, x); };"},
		// The receiver must be `ReactDOM` exactly. A subscripted receiver is not an identifier.
		{"a subscripted receiver", "var a = globals['ReactDOM'].render(<div />, x);"},
		// A variable subscript names whatever the variable holds, which is not knowable before it
		// runs. Upstream's `static_property_info` answers nothing for it and `AccessedName` declines
		// a bare identifier subscript whatever the accept set says. Measured silent on oxlint,
		// against the templated form above which reports.
		{"a variable subscript", "var b = ReactDOM[render](<div />, x);"},
		// An alias of the real thing. `RD` holds `ReactDOM` and this consumes the return value, so
		// it is a genuine miss rather than a case upstream thinks is fine. Measured silent on
		// oxlint, in a file that imports react-dom and reports on the unaliased call two lines
		// above. Reproduced rather than improved on; see the checker note on the rule.
		{"an aliased receiver", "import ReactDOM from 'react-dom';\nvar RD = ReactDOM;\nvar b = RD.render(<div />, x);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRenderReturnValue, renderReturnValueFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoRenderReturnValueReportsTwiceInsideANestedArrow pins upstream's double report.
//
// `() => () => ReactDOM.render(...)` satisfies both of upstream's checks at once: the call's parent
// is an arrow, and the parent's enclosing scope is also an arrow with an expression body. oxc
// diagnoses in both branches without a guard against having already reported, so the input produces
// two identical findings at the same span.
//
// This appears nowhere in the imported corpus, and it is not obviously intentional upstream. It is
// reproduced rather than deduplicated because a differential run against oxlint is what this port
// has to survive, and collapsing it to one finding would read as a difference on real code. Measured
// on oxlint: two findings, both at column 21 of the same line.
func TestNoRenderReturnValueReportsTwiceInsideANestedArrow(t *testing.T) {
	t.Parallel()

	const sourceText = "var f = () => () => ReactDOM.render(<div />, x);"
	result := rule_testing.Run(t, NoRenderReturnValue, renderReturnValueFile, sourceText)
	rule_testing.ExpectFindings(t, result, "noRenderReturnValue", "noRenderReturnValue")
}

// TestNoRenderReturnValueReachesAnywhereInsideAnArrowExpressionBody is the second half of the rule.
//
// Inside `() => <expression>`, position stops mattering entirely: the scope check asks only which
// scope the call's parent sits in, so an argument, a property read, an array element and a
// comparison all report, though every one of them is silent at file scope. The pairs below are the
// same four inputs as the silent cases above, wrapped in an arrow.
//
// This is the discrimination the imported corpus cannot see at all, and a port built only from the
// parent-kind list is silent on all of it while passing all seventeen upstream cases. Measured on
// oxlint, one finding each.
func TestNoRenderReturnValueReachesAnywhereInsideAnArrowExpressionBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an argument position", "var d = () => foo(ReactDOM.render(<div />, x));"},
		{"a property read off the result", "var c = () => ReactDOM.render(<div />, x).foo;"},
		{"an array element", "var e = () => [ReactDOM.render(<div />, x)];"},
		{"a comparison", "var g = () => ReactDOM.render(<div />, x) === null;"},
		// A parenthesized body. The parent is a parenthesized expression rather than the arrow, so
		// the parent-kind check misses it and only the scope check answers. Measured: reports once.
		{"a parenthesized arrow body", "var c = () => (ReactDOM.render(<div />, x));"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRenderReturnValue, renderReturnValueFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noRenderReturnValue")
		})
	}
}

// TestNoRenderReturnValueSpan asserts where the finding points and what it says.
//
// `ExpectFindings` sees neither, and this rule's span is not the obvious one: upstream merges the
// receiver identifier's span with the property's, so the finding covers `ReactDOM.render` and stops
// before the argument list. A port reporting the call expression, which is what the neighbouring
// `no-is-mounted` does, passes every fixture above while pointing at the wrong range.
//
// The subscript case is the one that pins the merge rather than a text length: `ReactDOM['render']`
// is 18 bytes where `ReactDOM.render` is 15, and both are measured from the oxlint snapshot rather
// than counted by hand.
func TestNoRenderReturnValueSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"a dotted access", "var Hello = ReactDOM.render(<div />, document.body);", "ReactDOM.render"},
		{"a subscripted access", "var q = ReactDOM['render'](<div />, x);", "ReactDOM['render']"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRenderReturnValue, renderReturnValueFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noRenderReturnValue")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d diagnostics, want 1", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			if reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]; reported != testCase.want {
				t.Errorf("finding covers %q, want %q", reported, testCase.want)
			}
		})
	}
}
