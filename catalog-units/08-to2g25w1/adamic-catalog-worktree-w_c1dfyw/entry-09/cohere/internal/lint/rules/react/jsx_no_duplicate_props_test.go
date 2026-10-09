package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// duplicatePropsFile is where the fixtures pretend to live.
//
// A `.tsx` extension rather than `.ts`, and it is load-bearing rather than cosmetic. The harness
// picks its script kind from the suffix, and under `.ts` the parser reads `<App a a />` as a type
// assertion instead of JSX, producing a tree with no JsxOpeningElement in it at all. Every fixture
// below would then pass by finding nothing, which reads exactly like a rule that works.
const duplicatePropsFile = "/repository/source/Duplicate.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case in the two blocks below is verbatim from
// `oxc/crates/oxc_linter/src/rules/react/jsx_no_duplicate_props.rs`, pulled with
// oxc's inline Tester block and written into this file by a script rather than by hand, so
// no transcription step existed that could cook an escape. One Tester block, 12 pass and 9 fail,
// and the snapshot records 9 diagnostics from those 9 inputs, so one finding per input is measured
// here rather than assumed.
func TestJsxNoDuplicatePropsFires(t *testing.T) {
	t.Parallel()

	cases := []string{
		"<App a a />;",
		"<App A b c A />;",
		"<App a=\"a\" b=\"b\" a=\"a\" />;",
		"<App a=\"a\" {...this.props} b=\"b\" a=\"a\" />;",
		"<App a b=\"b\" {...this.props} a=\"a\" />;",
		"<App a={[]} b=\"b\" {...this.props} a=\"a\" />;",
		"<App a=\"a\" b=\"b\" a=\"a\" {...this.props} />;",
		"<App {...this.props} a=\"a\" b=\"b\" a=\"a\" />;",
		"\n            <App\n                a=\"a\"\n                {...this.props}\n                a={{foo: 'bar'}}\n                b=\"b\"\n            />;\n        ",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, JsxNoDuplicateProps, duplicatePropsFile, sourceText),
				"jsxNoDuplicateProps")
		})
	}
}

// The clean cases carry the whole discrimination, and three groups of them matter.
//
// The case-sensitivity group is the largest and the most easily lost: `<App a b c A />`,
// `<App A a />`, `<App A b a />` and `<App A="a" b="b" B="B" />` are four separate assertions that
// `a` and `A` are distinct props. A port that lowercased names before comparing would report all
// four, and upstream states in its own doc comment that declining them is deliberate.
//
// The spread group checks that a spread neither reports nor resets: `<App {...this.props} a b c />`
// and `<App c {...this.props} a b />` interleave one with distinct names, and the fail block has
// the matching cases where names collide across a spread.
//
// And `<App />` with no attributes at all checks the empty path, which is the one an
// over-eager index would crash on rather than decline.
func TestJsxNoDuplicatePropsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"<App />;",
		"<App {...this.props} />;",
		"<App a b c />;",
		"<App a b c A />;",
		"<App {...this.props} a b c />;",
		"<App c {...this.props} a b />;",
		"<App a=\"c\" b=\"b\" c=\"a\" />;",
		"<App {...this.props} a=\"c\" b=\"b\" c=\"a\" />;",
		"<App c=\"a\" {...this.props} a=\"c\" b=\"b\" />;",
		"<App A a />;",
		"<App A b a />;",
		"<App A=\"a\" b=\"b\" B=\"B\" />;",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, JsxNoDuplicateProps, duplicatePropsFile, sourceText))
		})
	}
}

// The finding points at the earlier occurrence's name, and nothing weaker than a span check can
// see that.
//
// `ExpectFindings` asserts ids and count only, so a rule reporting the whole attribute, the whole
// element, or the *later* copy passes every fixture above. All three are plausible ports: ESLint
// reports the whole `JSXAttribute`, and reporting the copy you just found is the more obvious
// implementation than holding on to the one before it. oxc emits two labels and renders the earlier
// one as the diagnostic's position, which is what the imported snapshot prints and what the release
// oxlint binary prints for the same inputs.
func TestJsxNoDuplicatePropsReportsEarlierOccurrence(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		// The byte offset each finding should point at, in order. Offsets rather than substrings,
		// because every one of these inputs writes the same name more than once and a substring
		// search would find the first copy whichever one was reported.
		wantOffsets []int
	}{
		// `<App a a />;`  the second `a` is at 7, the first at 5. Upstream's snapshot puts this
		// diagnostic at column 6, one-based, which is offset 5.
		{"<App a a />;", []int{5}},
		// `<App A b c A />;`  offset 5 is the first `A`, offset 11 the second.
		{"<App A b c A />;", []int{5}},
		// The spread case whose snapshot column upstream states outright: `1:22` for
		// `<App {...this.props} a="a" b="b" a="a" />;` is offset 21, the first `a` after the spread.
		{"<App {...this.props} a=\"a\" b=\"b\" a=\"a\" />;", []int{21}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, JsxNoDuplicateProps, duplicatePropsFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantOffsets) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantOffsets))
			}
			for index, want := range testCase.wantOffsets {
				got := result.Diagnostics[index].Range.Pos()
				if got != want {
					t.Errorf("finding %d points at offset %d (%q), want offset %d (%q)",
						index, got, testCase.sourceText[got:result.Diagnostics[index].Range.End()],
						want, testCase.sourceText[want:want+1])
				}
			}
		})
	}
}

// The rendered message is asserted by equality rather than by containment.
//
// A `strings.Contains` predicate on an interpolated value is weaker than the property it guards and
// stays green while the text is wrong, so the whole string is compared here. This message
// interpolates nothing, which is itself a divergence worth pinning: oxc writes the prop name into
// the text and our harness renders a fixed Description, so a later change that started
// interpolating would need to update this assertion rather than sliding past it.
func TestJsxNoDuplicatePropsMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, JsxNoDuplicateProps, duplicatePropsFile, "<App a a />;")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	want := "This prop is written more than once on the same element. React builds the props " +
		"object by assigning each attribute in order, so every earlier copy is overwritten and " +
		"only the last one survives. The discarded values look like they are being passed and " +
		"are not. Remove one of them, or rename them so each prop is distinct."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Errorf("message text\n got %q\nwant %q", got, want)
	}
}

// Cases upstream does not cover, each pinning a decision the imported corpus cannot see.
//
// Every fail case upstream ships writes at most two copies of one name, and every one of them is
// self-closing, so several distinct rules produce identical results on all nine. These inputs are
// the ones that separate them, and each was measured against the release oxlint binary at
// `~/Projects/system/oxc/target/release/oxlint` before being written down.
func TestJsxNoDuplicatePropsBeyondUpstreamCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		// A third copy reports again, because the map is rewritten on every hit rather than only
		// on the first. Once per extra occurrence, not once per duplicated name. Measured: oxlint
		// prints two findings here and three for the four-copy case below. A rule reporting once
		// per name would give one and one, and would pass all nine upstream fail cases.
		{"three copies report twice", "<App a a a />;", 2},
		{"four copies report three times", "<App a a a a />;", 3},

		// Which copies are compared, not just how many findings there are. With the map rewritten
		// each time, the third `a` is compared against the second, so the two findings point at the
		// first and second `a`. A rule that kept the *first* occurrence forever would report the
		// same count while pointing at the first `a` twice. The offsets are asserted in the
		// dedicated test below.

		// A non-self-closing element. Upstream has no such fail case at all, so the
		// JsxOpeningElement arm of this rule is guarded by nothing in the imported corpus.
		{"open and close tags", "<App a a></App>;", 1},

		// The two element kinds do not both fire for one element. If they did, every fail fixture
		// above would report twice and this suite would already be red, but the input that would
		// show it directly is the one with a closing tag, above, which reports once.

		// A spread does not suppress what follows it. Upstream covers a spread *between* two
		// copies; this covers a spread before an untouched pair.
		{"spread then duplicate pair", "<App {...p} b b />;", 1},

		// A name inside a spread is invisible to both tools, and this is the reproduced gap rather
		// than a case the rule gets wrong. oxlint is silent here.
		{"collision inside a spread is not seen", "<App a {...{a: 1}} />;", 0},

		// Namespaced names never enter the map, so two identical ones do not collide. oxlint is
		// silent. A port that compared attribute *text* rather than requiring an identifier name
		// would report this.
		{"namespaced duplicate", "<svg xlink:href=\"x\" xlink:href=\"y\" />;", 0},

		// And a namespaced name does not collide with the plain name it ends in, which is the
		// sharper half of the same decision: reading `xlink:href` as `href` would make this report.
		{"namespaced beside plain", "<svg href=\"x\" xlink:href=\"y\" />;", 0},

		// A dashed name is a single identifier and does participate, which is the case that keeps
		// the namespaced decline from being implemented as "decline anything unusual".
		{"dashed names collide", "<App data-x data-x />;", 1},

		// An element with no attributes at all, and one with a lone spread. The empty path is the
		// one an implementation indexing before checking would crash on rather than decline.
		{"no attributes", "<App />;", 0},
		{"only a spread", "<App {...this.props} />;", 0},

		// Two elements in one file each report on their own. The map is built per element, so a
		// name used once in each is not a collision. A rule hoisting the map out of the listener
		// would report the second element and pass every single-element fixture upstream ships.
		{"separate elements do not share a map", "<A a /><B a />;", 0},

		// Nested elements, same shape. The inner element's props must not reach the outer map.
		{"nested elements do not share a map", "<A a><B a /></A>;", 0},

		// Case sensitivity, stated by upstream as deliberate and covered by four clean cases there.
		// This is the same judgment written as a direct assertion so the reason survives if those
		// four are ever read as incidental.
		{"case differs so no duplicate", "<App foo Foo />;", 0},
		{"same case does duplicate", "<App foo foo />;", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, JsxNoDuplicateProps, duplicatePropsFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantCount {
				for _, diagnostic := range result.Diagnostics {
					t.Logf("  finding at %q",
						testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()])
				}
				t.Errorf("%q got %d findings, want %d",
					testCase.sourceText, len(result.Diagnostics), testCase.wantCount)
			}
		})
	}
}

// Three copies report against the copy before them, not against the first one forever.
//
// This is the assertion that separates "keep the earliest occurrence" from "keep the most recent
// one" as the map's value. Both report twice on `<App a a a />`, so a count check cannot tell them
// apart; only the offsets can. Upstream rewrites the map entry on every insert, so the second
// finding points at the second `a` rather than at the first again.
func TestJsxNoDuplicatePropsThirdCopyComparesAgainstSecond(t *testing.T) {
	t.Parallel()

	const sourceText = "<App a a a />;"
	result := rule_testing.Run(t, JsxNoDuplicateProps, duplicatePropsFile, sourceText)
	if len(result.Diagnostics) != 2 {
		t.Fatalf("got %d findings, want 2", len(result.Diagnostics))
	}
	// The three `a` characters sit at offsets 5, 7 and 9.
	wantOffsets := []int{5, 7}
	for index, want := range wantOffsets {
		if got := result.Diagnostics[index].Range.Pos(); got != want {
			t.Errorf("finding %d points at offset %d, want %d", index, got, want)
		}
	}
}

// Upstream's `ignoreCase` rows, the four that pass the option, read through the registered decoder.
//
// The three failing ones are clean without the option, which the clean table above pins; under it
// they report once each. The namespaced `a:b` stays clean under it, as upstream's clean case says.
func TestJsxNoDuplicatePropsIgnoreCase(t *testing.T) {
	t.Parallel()

	decoded, err := rule.DecodeOptionsInto[JsxNoDuplicatePropsOptions]()([]byte(`{"ignoreCase": true}`))
	if err != nil {
		t.Fatalf("decoding ignoreCase: %v", err)
	}
	cases := []struct {
		sourceText string
		wantCount  int
	}{
		{"<App A a />;", 1},
		{"<App a b c A />;", 1},
		{`<App A="a" b="b" B="B" />;`, 1},
		{`<App a:b="c" />;`, 0},
		{`<svg a:b="c" A:b="d" />;`, 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, JsxNoDuplicateProps, duplicatePropsFile, testCase.sourceText, decoded)
			if len(result.Diagnostics) != testCase.wantCount {
				t.Errorf("got %d findings, want %d", len(result.Diagnostics), testCase.wantCount)
			}
		})
	}

	if _, err := rule.DecodeOptionsInto[JsxNoDuplicatePropsOptions]()([]byte(`{"ignorecase": true}`)); err == nil {
		t.Error("a misspelled ignoreCase decoded")
	}
}
