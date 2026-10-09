package core

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/suppression"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// requireDescriptionFile is a .tsx name so the JSX comment case parses as JSX. The rule has no file
// gate, matching upstream's plain `create` with no source-type test.
const requireDescriptionFile = "/repository/source/RequireDescription.tsx"

// requireDescriptionCase is one row: source, the raw option object as JSON (empty for none), and the
// exact text of every span the rule must report, in order. An empty want list is a clean case.
type requireDescriptionCase struct {
	sourceText string
	options    string
	wantSpans  []string
}

// runRequireDescription drives one case through the rule's own exported decoder, so the option
// object reaches the rule the way the config layer hands it over.
func runRequireDescription(t *testing.T, testCase requireDescriptionCase) rule_testing.Result {
	t.Helper()
	decoded, err := DecodeRequireDescriptionOptions([]byte(testCase.options))
	if err != nil {
		t.Fatalf("decoding %q: %v", testCase.options, err)
	}
	return rule_testing.RunWithOptions(t, RequireDescription, requireDescriptionFile, testCase.sourceText, decoded)
}

// checkRequireDescription asserts the count, the message id and the exact span text of every finding.
//
// The span is asserted as text rather than as offsets because `rule_testing.Run` does not trim the
// fixture, so the literal and the parsed file are byte-identical and a slice of one is a slice of
// the other.
func checkRequireDescription(t *testing.T, testCase requireDescriptionCase) {
	t.Helper()
	result := runRequireDescription(t, testCase)
	if len(testCase.wantSpans) == 0 {
		rule_testing.ExpectClean(t, result)
		return
	}
	wantIds := make([]string, len(testCase.wantSpans))
	for index := range wantIds {
		wantIds[index] = "missingDescription"
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	for index, diagnostic := range result.Diagnostics {
		got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
		if got != testCase.wantSpans[index] {
			t.Errorf("finding %d spans %q, want %q", index, got, testCase.wantSpans[index])
		}
	}
}

func runRequireDescriptionCases(t *testing.T, cases []requireDescriptionCase) {
	t.Helper()
	for index, testCase := range cases {
		t.Run("case"+strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()
			checkRequireDescription(t, testCase)
		})
	}
}

// TestRequireDescriptionStaysSilentOnUpstreamPassCases is upstream's valid corpus, verbatim.
//
// Extracted from the clone's test file by loading it under a stub RuleTester rather than by
// retyping it (scratchpad `clone-suppression/corpus.cjs`), then each case run through the cloned rule
// on the installed ESLint 10.8.1: zero findings on all of them. Left out, with reasons: the two
// `eslint-env` rows, which upstream itself gates to ESLint <= 9; the `@eslint/css` language-plugin
// row, since cohere lints no CSS; and the two sibling-plugin rows, which test a cache shared with
// other rules in the plugin and whose code is already present below as plain rows.
func TestRequireDescriptionStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()
	runRequireDescriptionCases(t, []requireDescriptionCase{
		{sourceText: `/* eslint eqeqeq: "off", curly: "error" -- Here's a description about why this configuration is necessary. */`},
		{sourceText: "/* eslint-disable -- description */"},
		{sourceText: "/* eslint-enable -- description */"},
		{sourceText: "/* exported -- description */"},
		{sourceText: "/* global -- description */"},
		{sourceText: "/* globals -- description */"},
		{sourceText: "/* just eslint in a normal comment */"},
		{sourceText: "/* c8 without options */"},
		{sourceText: "// eslint-disable-line -- description"},
		{sourceText: "// eslint-disable-next-line -- description"},
		{sourceText: "/* eslint-disable-line -- description */"},
		{sourceText: "/* eslint-disable-next-line -- description */"},
		{sourceText: "// eslint-disable-line eqeqeq -- description"},
		{sourceText: "// eslint-disable-next-line eqeqeq -- description"},
		{sourceText: "/* eslint */", options: `{"ignore":["eslint"]}`},
		{sourceText: "/* eslint-enable */", options: `{"ignore":["eslint-enable"]}`},
		{sourceText: "/* eslint-disable */", options: `{"ignore":["eslint-disable"]}`},
		{sourceText: "// eslint-disable-line", options: `{"ignore":["eslint-disable-line"]}`},
		{sourceText: "// eslint-disable-next-line", options: `{"ignore":["eslint-disable-next-line"]}`},
		{sourceText: "/* eslint-disable-line */", options: `{"ignore":["eslint-disable-line"]}`},
		{sourceText: "/* eslint-disable-next-line */", options: `{"ignore":["eslint-disable-next-line"]}`},
		{sourceText: "/* exported */", options: `{"ignore":["exported"]}`},
		{sourceText: "/* global */", options: `{"ignore":["global"]}`},
		{sourceText: "/* globals */", options: `{"ignore":["globals"]}`},
		{sourceText: "/* c8 ignore next -- description */", options: `{"additionalDirectives":["c8"]}`},
		{sourceText: "/* c8 ignore next */"},
	})
}

// TestRequireDescriptionFiresOnUpstreamFailCases is upstream's invalid corpus, verbatim, one finding
// each. The span is the whole comment: upstream's end column matched the comment's end on every row
// (for example `/* eslint-disable */` ends at column 21), and its start is column -1, which is the
// suppression escape documented on the rule rather than a place to point.
func TestRequireDescriptionFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()
	whole := func(sourceText string, options string) requireDescriptionCase {
		return requireDescriptionCase{sourceText: sourceText, options: options, wantSpans: []string{sourceText}}
	}
	runRequireDescriptionCases(t, []requireDescriptionCase{
		whole("/* eslint */", ""),
		whole(`/* eslint eqeqeq: "off", curly: "error" */`, ""),
		whole("/* eslint-enable */", ""),
		whole("/* eslint-enable eqeqeq */", ""),
		whole("/* eslint-disable */", ""),
		whole("/* eslint-disable eqeqeq */", ""),
		whole("// eslint-disable-line", ""),
		whole("// eslint-disable-line eqeqeq", ""),
		whole("// eslint-disable-next-line", ""),
		whole("// eslint-disable-next-line eqeqeq", ""),
		whole("/* eslint-disable-line */", ""),
		whole("/* eslint-disable-line eqeqeq */", ""),
		whole("/* eslint-disable-next-line */", ""),
		whole("/* eslint-disable-next-line eqeqeq */", ""),
		whole("/* exported */", ""),
		whole("/* global */", ""),
		whole("/* global _ */", ""),
		whole("/* globals */", ""),
		whole("/* globals _ */", ""),
		whole("/* eslint-disable-next-line eqeqeq -- */", ""),
		whole("/* c8 ignore next */", `{"additionalDirectives":["c8"]}`),
		// Gated to ESLint <= 9 upstream; measured on 10.8.1 the rule still reports both, beside
		// ESLint's own fatal "no longer supported" message.
		whole("/* eslint-env */", ""),
		whole("/* eslint-env node */", ""),
	})
}

// TestRequireDescriptionReadsTheSeparatorAsESLintDoes is the reason grammar, every row measured on
// the cloned rule through ESLint 10.8.1 (scratchpad `clone-suppression/probe-cases.json`).
//
// The `--` followed only by whitespace is the row the task names: a separator is not a reason.
func TestRequireDescriptionReadsTheSeparatorAsESLintDoes(t *testing.T) {
	t.Parallel()
	runRequireDescriptionCases(t, []requireDescriptionCase{
		// Reports: no reason.
		{sourceText: "// eslint-disable-next-line foo -- ", wantSpans: []string{"// eslint-disable-next-line foo -- "}},
		{sourceText: "// eslint-disable-next-line foo --", wantSpans: []string{"// eslint-disable-next-line foo --"}},
		{sourceText: "// eslint-disable-next-line foo --\t", wantSpans: []string{"// eslint-disable-next-line foo --\t"}},
		{sourceText: "// eslint-disable-next-line foo--bar", wantSpans: []string{"// eslint-disable-next-line foo--bar"}},
		{sourceText: "/* eslint-disable -- \t */", wantSpans: []string{"/* eslint-disable -- \t */"}},
		// Clean: a reason, however it is spaced.
		{sourceText: "// eslint-disable-next-line foo -- bar"},
		{sourceText: "// eslint-disable-next-line foo ---  bar"},
		{sourceText: "// eslint-disable-next-line foo -- -- "},
		{sourceText: "// eslint-disable-next-line foo -- --bar"},
		{sourceText: "/* eslint-disable-next-line\tfoo\t--\treason */"},
		// JavaScript's \s includes the no-break space; Go's does not, which is why the class is spelled out.
		{sourceText: "// eslint-disable-next-line foo -- bar"},
	})
}

// TestRequireDescriptionMatchesUpstreamOnCommentShapes is every other shape the oracle was asked
// about where upstream's parser decides, so both engines agree.
func TestRequireDescriptionMatchesUpstreamOnCommentShapes(t *testing.T) {
	t.Parallel()
	runRequireDescriptionCases(t, []requireDescriptionCase{
		// A trailing directive: upstream reports from column -1, here the comment alone.
		{sourceText: "x; // eslint-disable-line foo", wantSpans: []string{"// eslint-disable-line foo"}},
		{sourceText: "const a = 1;\n    /* eslint-disable */\nconst b = 2;", wantSpans: []string{"/* eslint-disable */"}},
		{sourceText: "const a = <div>{/* eslint-disable-line */}</div>;", wantSpans: []string{"/* eslint-disable-line */"}},
		{sourceText: "/*eslint-disable*/", wantSpans: []string{"/*eslint-disable*/"}},
		{sourceText: "//eslint-disable-line", wantSpans: []string{"//eslint-disable-line"}},
		// A multi-line block without star continuations is still a directive to ESLint.
		{sourceText: "/*\n eslint-disable\n */", wantSpans: []string{"/*\n eslint-disable\n */"}},
		{sourceText: "/* eslint-disable-next-line\n */", wantSpans: []string{"/* eslint-disable-next-line\n */"}},
		// Two directives, one described: only the bare one reports.
		{
			sourceText: "/* eslint-disable */\nfoo();\n// eslint-disable-next-line -- why\nbar();\n/* eslint-enable */",
			wantSpans:  []string{"/* eslint-disable */", "/* eslint-enable */"},
		},
		// Not directives in either engine.
		{sourceText: "/** eslint-disable */"},
		{sourceText: "/* eslint-disabled */"},
		{sourceText: "const s = '/* eslint-disable */';"},
		{sourceText: "// eslint eqeqeq: 0"},
		{sourceText: "// global foo"},
		{sourceText: "// a note about eslint-disable"},
		// additionalDirectives: a named word is a directive in a line comment too, and only as a whole word.
		{sourceText: "// c8 ignore next", options: `{"additionalDirectives":["c8"]}`, wantSpans: []string{"// c8 ignore next"}},
		{sourceText: "// c8 ignore next -- covered elsewhere", options: `{"additionalDirectives":["c8"]}`},
		{sourceText: "// c8 ignore next"},
		{sourceText: "/* istanbul ignore next */", options: `{"additionalDirectives":["c8","istanbul"]}`, wantSpans: []string{"/* istanbul ignore next */"}},
		{sourceText: "/* c8x ignore */", options: `{"additionalDirectives":["c8"]}`},
		// An unterminated block is a parse error to ESLint, which then reports nothing at all.
		{sourceText: "x;\n/* eslint-disable"},
		// An ignore naming a different kind ignores nothing else.
		{sourceText: "/* eslint-disable */", options: `{"ignore":["eslint-disable-line"]}`, wantSpans: []string{"/* eslint-disable */"}},
		{sourceText: "/* eslint-enable */", options: `{"ignore":["eslint-disable"]}`, wantSpans: []string{"/* eslint-enable */"}},
		{sourceText: "/* eslint-disable */", options: `{}`, wantSpans: []string{"/* eslint-disable */"}},
	})
}

// TestRequireDescriptionCoversWhatCohereHonors is the second arm: comments upstream cannot see, or
// sees and declines, that this tree's suppression index acts on.
//
// The `eslint-` rows here are the one measured divergence from upstream, and each was measured
// silent on the cloned rule: ESLint ignores a line-comment `// eslint-disable`, a star-continued
// block, and a multi-line `eslint-disable-line`, while cohere suppresses with every one of them.
func TestRequireDescriptionCoversWhatCohereHonors(t *testing.T) {
	t.Parallel()
	runRequireDescriptionCases(t, []requireDescriptionCase{
		// cohere's own spellings, every scope, and the enable.
		{sourceText: "// cohere-disable-next-line no-alert", wantSpans: []string{"// cohere-disable-next-line no-alert"}},
		{sourceText: "x; // cohere-disable-line", wantSpans: []string{"// cohere-disable-line"}},
		{sourceText: "/* cohere-disable */", wantSpans: []string{"/* cohere-disable */"}},
		{sourceText: "/* cohere-enable */", wantSpans: []string{"/* cohere-enable */"}},
		{sourceText: "// verify-disable-next-line foo", wantSpans: []string{"// verify-disable-next-line foo"}},
		{sourceText: "// oxlint-disable-next-line foo", wantSpans: []string{"// oxlint-disable-next-line foo"}},
		{sourceText: "/* oxlint-enable foo */", wantSpans: []string{"/* oxlint-enable foo */"}},
		{sourceText: "{/* cohere-disable-next-line */}", wantSpans: []string{"/* cohere-disable-next-line */"}},
		// The same separator grammar as the eslint spelling, not cohere's laxer one.
		{sourceText: "// cohere-disable-next-line foo -- ", wantSpans: []string{"// cohere-disable-next-line foo -- "}},
		{sourceText: "// cohere-disable-next-line foo--bar", wantSpans: []string{"// cohere-disable-next-line foo--bar"}},
		{sourceText: "// cohere-disable-next-line no-alert -- the kiosk build has no other channel"},
		{sourceText: "/* cohere-disable -- generated */"},
		{sourceText: "/* cohere-enable -- generated */"},
		// Near misses the suppression index refuses.
		{sourceText: "// cohere-disabled"},
		{sourceText: "// a note about cohere-disable"},
		// The ignore option reaches a cohere spelling through its shared kind.
		{sourceText: "// cohere-disable-next-line foo", options: `{"ignore":["eslint-disable-next-line"]}`},
		{sourceText: "/* cohere-enable */", options: `{"ignore":["eslint-enable"]}`},
		{sourceText: "/* cohere-disable */", options: `{"ignore":["eslint-disable-line"]}`, wantSpans: []string{"/* cohere-disable */"}},
		// A file-scope disable and an enable written as `//` are prose to ESLint, and since Kirk's
		// ruling of 2026-10-01 to cohere too, so there is no directive to describe.
		{sourceText: "// eslint-disable"},
		{sourceText: "// eslint-disable eqeqeq"},
		{sourceText: "// eslint-enable"},
		{sourceText: "/*\n * eslint-disable\n */", wantSpans: []string{"/*\n * eslint-disable\n */"}},
		{sourceText: "/* eslint-disable-line\n */", wantSpans: []string{"/* eslint-disable-line\n */"}},
		// A scope word must end at whitespace or the end of the comment, as in ESLint, so this is not a
		// directive to either engine.
		{sourceText: "/* eslint-disable-next-line, foo */"},
		{sourceText: "// eslint-disable -- generated", options: ``},
		{sourceText: "// eslint-disable", options: `{"ignore":["eslint-disable"]}`},
	})
}

// TestRequireDescriptionIsRegisteredAsADirectiveSubject checks the second registration in
// require_description_register.go through the real suppression index. The fixture harness above
// never applies suppressions, so without this nothing here would notice that line going missing,
// and every blanket `// eslint-disable-line` in a real run would silence its own finding.
//
// The control is an ordinary rule name at the same offset, which the same directive does silence.
func TestRequireDescriptionIsRegisteredAsADirectiveSubject(t *testing.T) {
	t.Parallel()
	source := "x; // eslint-disable-line\n"
	if !suppression.Build(source).Suppresses("no-alert", 3) {
		t.Fatal("control: a blanket same-line directive should silence an ordinary rule on its line")
	}
	if suppression.Build(source).Suppresses(RequireDescription.Name, 3) {
		t.Fatal("a same-line directive silenced the finding about itself; the rule is not registered as a directive subject")
	}
}

// TestRequireDescriptionWithNoOptionsAtAll bypasses the decoder, which is how a bare severity
// reached a rule as nil and left another rule inert across the whole tree.
func TestRequireDescriptionWithNoOptionsAtAll(t *testing.T) {
	t.Parallel()
	result := rule_testing.Run(t, RequireDescription, requireDescriptionFile, "// eslint-disable-line\n")
	rule_testing.ExpectFindings(t, result, "missingDescription")
}

// TestRequireDescriptionDecoderRefusesWhatTheSchemaRefuses covers `additionalProperties: false`,
// the `ignore` enum and `uniqueItems`, each of which would otherwise decode to "ignore nothing".
func TestRequireDescriptionDecoderRefusesWhatTheSchemaRefuses(t *testing.T) {
	t.Parallel()
	refused := map[string]string{
		`{"ignored":["eslint"]}`:                       "unknown field",
		`{"ignore":["eslint-disable-file"]}`:           "wanted one of",
		`{"ignore":["eslint", "eslint"]}`:              "twice",
		`{"ignore":"eslint"}`:                          "cannot unmarshal",
		`{"additionalDirectives":[1]}`:                 "cannot unmarshal",
		`{"ignore":["cohere-disable-next-line"]}`:      "wanted one of",
		`{"ignore":["ESLINT-DISABLE"],"extra":true}`:   "unknown field",
		`{"ignore":["Eslint-Disable-Next-Line"]}`:      "wanted one of",
		`{"additionalDirectives":["c8"],"x":null}`:     "unknown field",
		`{"ignore":["global","globals","global"]}`:     "twice",
		`{"ignore":["eslint-disable-next-line", 3]}`:   "cannot unmarshal",
		`{"ignore":[null]}`:                            "wanted one of",
		`{"additionalDirectives":"c8"}`:                "cannot unmarshal",
		`{"ignore":["exported"],"ignore ":["global"]}`: "unknown field",
	}
	for raw, wantFragment := range refused {
		_, err := DecodeRequireDescriptionOptions([]byte(raw))
		if err == nil || !strings.Contains(err.Error(), wantFragment) {
			t.Errorf("decoding %s: got %v, want an error containing %q", raw, err, wantFragment)
		}
	}

	decoded, err := DecodeRequireDescriptionOptions(nil)
	if err != nil || len(decoded.(RequireDescriptionOptions).Ignore) != 0 {
		t.Errorf("a bare severity should decode to the zero options, got %#v, %v", decoded, err)
	}
	decoded, err = DecodeRequireDescriptionOptions([]byte(`{"ignore":["eslint-disable","eslint-enable"],"additionalDirectives":["c8"]}`))
	options, _ := decoded.(RequireDescriptionOptions)
	if err != nil || len(options.Ignore) != 2 || len(options.AdditionalDirectives) != 1 {
		t.Errorf("a valid object should decode whole, got %#v, %v", decoded, err)
	}
}
