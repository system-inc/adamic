package react

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestNoUnescapedEntitiesStaysSilent runs every passing case in upstream's corpus.
//
// Extracted mechanically from
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/no-unescaped-entities.js` by
// requiring the test file with a stubbed RuleTester and emitting each `code` through
// JSON.stringify, so no escape sequence was ever typed on the way here. Verified byte for
// byte against the corpus after writing.
func TestNoUnescapedEntitiesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		// upstream valid[0]
		{name: "upstreamValid0", source: "\n        var Hello = createReactClass({\n          render: function() {\n            return (\n              <div/>\n            );\n          }\n        });\n      "},
		// upstream valid[1]
		{name: "upstreamValid1", source: "\n        var Hello = createReactClass({\n          render: function() {\n            return <div>Here is some text!</div>;\n          }\n        });\n      "},
		// upstream valid[2]
		{name: "upstreamValid2", source: "\n        var Hello = createReactClass({\n          render: function() {\n            return <div>I&rsquo;ve escaped some entities: &gt; &lt; &amp;</div>;\n          }\n        });\n      "},
		// upstream valid[3]
		{name: "upstreamValid3", source: "\n        var Hello = createReactClass({\n          render: function() {\n            return <div>first line is ok\n            so is second\n            and here are some escaped entities: &gt; &lt; &amp;</div>;\n          }\n        });\n      "},
		// upstream valid[4]
		{name: "upstreamValid4", source: "\n        var Hello = createReactClass({\n          render: function() {\n            return <div>{\">\" + \"<\" + \"&\" + '\"'}</div>;\n          },\n        });\n      "},
		// upstream valid[5], fragment
		{name: "upstreamValid5", source: "\n        var Hello = createReactClass({\n          render: function() {\n            return <>Here is some text!</>;\n          }\n        });\n      "},
		// upstream valid[6], fragment
		{name: "upstreamValid6", source: "\n        var Hello = createReactClass({\n          render: function() {\n            return <>I&rsquo;ve escaped some entities: &gt; &lt; &amp;</>;\n          }\n        });\n      "},
		// upstream valid[7], fragment
		{name: "upstreamValid7", source: "\n        var Hello = createReactClass({\n          render: function() {\n            return <>{\">\" + \"<\" + \"&\" + '\"'}</>;\n          },\n        });\n      "},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnescapedEntities, "component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUnescapedEntitiesFires runs every reporting case in upstream's corpus.
//
// Twelve cases rather than the eight a naive extraction yields. Four of them sit behind the
// corpus's `allowsInvalidJSX` flag, which upstream computes from the installed acorn-jsx
// version because espree rejects a bare `>` or `}` in JSX text as a parse error. Our parser
// accepts both, measured, so those four reach the rule here and are included. Their expected
// findings come from the corpus's own `errors` entries rather than from a run, because the
// reference implementation as installed cannot parse them.
//
// Message ids and their per-case counts are upstream's, taken from each case's `errors`
// array. Order is asserted, and it is upstream's loop order rather than source order: see
// reportEntities.
func TestNoUnescapedEntitiesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		options any
		wantIds []string
	}{
		// upstream invalid[0]
		{
			name:    "upstreamInvalid0",
			source:  "\n          var Hello = createReactClass({\n            render: function() {\n              return <div>> default parser</div>;\n            }\n          });\n        ",
			options: nil,
			wantIds: []string{"unescapedEntityAlts"},
		},
		// upstream invalid[1]
		{
			name:    "upstreamInvalid1",
			source:  "\n          var Hello = createReactClass({\n            render: function() {\n              return <div>first line is ok\n              so is second\n              and here are some bad entities: ></div>\n            }\n          });\n        ",
			options: nil,
			wantIds: []string{"unescapedEntityAlts"},
		},
		// upstream invalid[2]
		{
			name:    "upstreamInvalid2",
			source:  "\n          var Hello = createReactClass({\n            render: function() {\n              return <div>Multiple errors: '>> default parser</div>;\n            }\n          });\n        ",
			options: nil,
			wantIds: []string{"unescapedEntityAlts", "unescapedEntityAlts", "unescapedEntityAlts"},
		},
		// upstream invalid[3]
		{
			name:    "upstreamInvalid3",
			source:  "\n          var Hello = createReactClass({\n            render: function() {\n              return <div>{\"Unbalanced braces - default parser\"}}</div>;\n            }\n          });\n        ",
			options: nil,
			wantIds: []string{"unescapedEntityAlts"},
		},
		// upstream invalid[4], fragment
		{
			name:    "upstreamInvalid4",
			source:  "\n        var Hello = createReactClass({\n          render: function() {\n            return <>> babel-eslint</>;\n          }\n        });\n      ",
			options: nil,
			wantIds: []string{"unescapedEntityAlts"},
		},
		// upstream invalid[5], fragment
		{
			name:    "upstreamInvalid5",
			source:  "\n        var Hello = createReactClass({\n          render: function() {\n            return <>first line is ok\n            so is second\n            and here are some bad entities: ></>\n          }\n        });\n      ",
			options: nil,
			wantIds: []string{"unescapedEntityAlts"},
		},
		// upstream invalid[6]
		{
			name:    "upstreamInvalid6",
			source:  "\n        var Hello = createReactClass({\n          render: function() {\n            return <div>'</div>;\n          }\n        });\n      ",
			options: nil,
			wantIds: []string{"unescapedEntityAlts"},
		},
		// upstream invalid[7], fragment
		{
			name:    "upstreamInvalid7",
			source:  "\n        var Hello = createReactClass({\n          render: function() {\n            return <>{\"Unbalanced braces - babel-eslint\"}}</>;\n          }\n        });\n      ",
			options: nil,
			wantIds: []string{"unescapedEntityAlts"},
		},
		// upstream invalid[8], fragment, options {"forbid":["&"]}
		{
			name:    "upstreamInvalid8",
			source:  "\n        var Hello = createReactClass({\n          render: function() {\n            return <>foo & bar</>;\n          }\n        });\n      ",
			options: decodeNoUnescapedEntitiesOptionsForTest(t, "{\"forbid\":[\"&\"]}"),
			wantIds: []string{"unescapedEntity"},
		},
		// upstream invalid[9], options {"forbid":["&"]}
		{
			name:    "upstreamInvalid9",
			source:  "\n        var Hello = createReactClass({\n          render: function() {\n            return <span>foo & bar</span>;\n          }\n        });\n      ",
			options: decodeNoUnescapedEntitiesOptionsForTest(t, "{\"forbid\":[\"&\"]}"),
			wantIds: []string{"unescapedEntity"},
		},
		// upstream invalid[10], options {"forbid":[{"char":"&","alternatives":["&amp;"]}]}
		{
			name:    "upstreamInvalid10",
			source:  "\n        var Hello = createReactClass({\n          render: function() {\n            return <span>foo & bar</span>;\n          }\n        });\n      ",
			options: decodeNoUnescapedEntitiesOptionsForTest(t, "{\"forbid\":[{\"char\":\"&\",\"alternatives\":[\"&amp;\"]}]}"),
			wantIds: []string{"unescapedEntityAlts"},
		},
		// upstream invalid[11]
		{
			name:    "upstreamInvalid11",
			source:  "\n        <script>window.foo = \"bar\"</script>\n      ",
			options: nil,
			wantIds: []string{"unescapedEntityAlts", "unescapedEntityAlts"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(
				t, NoUnescapedEntities, "component.tsx", testCase.source, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// decodeNoUnescapedEntitiesOptionsForTest routes a corpus `options` object through the rule's
// own decoder.
//
// Building NoUnescapedEntitiesOptions by hand would leave the decoder untested, and the
// decoder is where this rule's two least upstream-shaped decisions live: the absent-versus-
// empty distinction on `forbid`, and the dispatch between a string element and an object one.
func decodeNoUnescapedEntitiesOptionsForTest(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeNoUnescapedEntitiesOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("decoding %s: %v", raw, err)
	}
	return decoded
}

// TestNoUnescapedEntitiesPointsAtTheCharacter asserts the span, which no message-id fixture sees.
//
// Upstream reports a zero-width point: every diagnostic comes back carrying `line` and `column`
// with `endLine` and `endColumn` absent, measured through the ESLint Linter API on the installed
// build. Our Diagnostic has no way to express a point, so this reports the character itself. The
// START is identical to upstream's; the end is one byte later. That is the divergence, and this is
// where it is pinned, so a later change to the span fails loudly rather than drifting.
func TestNoUnescapedEntitiesPointsAtTheCharacter(t *testing.T) {
	t.Parallel()

	const source = "const a = <div>x\"y</div>;"

	result := rule_testing.Run(t, NoUnescapedEntities, "component.tsx", source)
	rule_testing.ExpectFindings(t, result, "unescapedEntityAlts")

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != `"` {
		t.Fatalf("expected the finding to cover the quote alone, got %q", reported)
	}
	// Upstream's column for this input is 15, one-based, which is byte offset 16 zero-based.
	// Measured on the installed rule.
	if got := result.Diagnostics[0].Range.Pos(); got != 16 {
		t.Fatalf("expected the finding to start at offset 16, got %d", got)
	}
}

// TestNoUnescapedEntitiesRendersUpstreamsMessageText asserts the rendered strings exactly.
//
// Both messages interpolate, so `ExpectFindings` cannot see anything the format string does. The
// expected text is upstream's own, read off the installed rule rather than off our message
// constants: comparing a diagnostic against the constant the rule reported with is equality that
// moves on both sides under mutation.
func TestNoUnescapedEntitiesRendersUpstreamsMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUnescapedEntities, "component.tsx", "const a = <div>'</div>;")
	rule_testing.ExpectFindings(t, result, "unescapedEntityAlts")

	const want = "`'` can be escaped with `&apos;`, `&lsquo;`, `&#39;`, `&rsquo;`."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Fatalf("message text:\n got  %q\n want %q", got, want)
	}
}

// TestNoUnescapedEntitiesRendersTheBareStringMessage pins the other message arm's text.
//
// A bare-string `forbid` entry produces `unescapedEntity`, which offers nothing and reads
// differently. Upstream renders "HTML entity, `&` , must be escaped." with the stray space before
// the comma, an artifact of its `{{entity}}` placeholder sitting between two spaces. That spacing
// is NOT reproduced: the id is what suppressions and metrics resolve against, and the sentence is
// ours to write clearly. The divergence is stated here rather than left for a reader to find.
func TestNoUnescapedEntitiesRendersTheBareStringMessage(t *testing.T) {
	t.Parallel()

	options := decodeNoUnescapedEntitiesOptionsForTest(t, `{"forbid":["&"]}`)

	result := rule_testing.RunWithOptions(
		t, NoUnescapedEntities, "component.tsx", "const a = <div>foo & bar</div>;", options)
	rule_testing.ExpectFindings(t, result, "unescapedEntity")

	const want = "HTML entity `&` is written raw in JSX text and must be escaped."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Fatalf("message text:\n got  %q\n want %q", got, want)
	}
	if got := len(result.Diagnostics[0].Suggestions); got != 0 {
		t.Fatalf("the bare-string spelling offers no repair upstream, got %d suggestions", got)
	}
}

// TestNoUnescapedEntitiesOffersEveryAlternativeAsASuggestion asserts the repairs, not just the ids.
//
// `rule_testing` can apply a fix but not a suggestion, so the applier is hand-rolled here. The expected
// output for each of the four is upstream's own, read off the installed rule's `suggestions[].fix`.
// A suggestion writing the right string over the wrong span passes any text comparison, so the
// assertion is on the rewritten whole file.
func TestNoUnescapedEntitiesOffersEveryAlternativeAsASuggestion(t *testing.T) {
	t.Parallel()

	const source = "const a = <div>x\"y</div>;"

	result := rule_testing.Run(t, NoUnescapedEntities, "component.tsx", source)
	rule_testing.ExpectFindings(t, result, "unescapedEntityAlts")

	wantDescriptions := []string{
		"Replace with `&quot;`.",
		"Replace with `&ldquo;`.",
		"Replace with `&#34;`.",
		"Replace with `&rdquo;`.",
	}
	wantSources := []string{
		"const a = <div>x&quot;y</div>;",
		"const a = <div>x&ldquo;y</div>;",
		"const a = <div>x&#34;y</div>;",
		"const a = <div>x&rdquo;y</div>;",
	}

	suggestions := result.Diagnostics[0].Suggestions
	if len(suggestions) != len(wantSources) {
		t.Fatalf("expected %d suggestions, got %d", len(wantSources), len(suggestions))
	}
	for index, suggestion := range suggestions {
		if suggestion.Message.Id != "replaceWithAlt" {
			t.Fatalf("suggestion %d: expected id replaceWithAlt, got %q", index, suggestion.Message.Id)
		}
		if got := suggestion.Message.Description; got != wantDescriptions[index] {
			t.Fatalf("suggestion %d description:\n got  %q\n want %q", index, got, wantDescriptions[index])
		}
		if got := applyOneSuggestion(t, source, suggestion); got != wantSources[index] {
			t.Fatalf("suggestion %d output:\n got  %q\n want %q", index, got, wantSources[index])
		}
	}
}

// TestNoUnescapedEntitiesRepairsTheRightLineOfAMultiLineNode pins the fixer's line arithmetic.
//
// Upstream's fixer rebuilds `node.raw` line by line and substitutes only on the line the finding
// landed on, and the index it substitutes at is relative to a slice whose first line starts at the
// node's own start column. Getting that wrong on a node that begins mid-line produces a repair that
// lands one column off, which no message-id fixture and no single-line case can see. Expected
// output measured on the installed rule.
func TestNoUnescapedEntitiesRepairsTheRightLineOfAMultiLineNode(t *testing.T) {
	t.Parallel()

	const source = "const a = <div>ok\n  bad: \"q\"\n  more</div>;"

	result := rule_testing.Run(t, NoUnescapedEntities, "component.tsx", source)
	rule_testing.ExpectFindings(t, result, "unescapedEntityAlts", "unescapedEntityAlts")

	const wantFirst = "const a = <div>ok\n  bad: &quot;q\"\n  more</div>;"
	if got := applyOneSuggestion(t, source, result.Diagnostics[0].Suggestions[0]); got != wantFirst {
		t.Fatalf("first finding's repair:\n got  %q\n want %q", got, wantFirst)
	}
	const wantSecond = "const a = <div>ok\n  bad: \"q&quot;\n  more</div>;"
	if got := applyOneSuggestion(t, source, result.Diagnostics[1].Suggestions[0]); got != wantSecond {
		t.Fatalf("second finding's repair:\n got  %q\n want %q", got, wantSecond)
	}
}

// TestNoUnescapedEntitiesReportsInSourceOrder pins the loop nesting, which is the one place this
// port deliberately does not reproduce upstream's code.
//
// Upstream emits entity-major, so it would emit the `>` in `'>x` before the `'`. That order is not
// observable through ESLint, which sorts diagnostics by position; measured by reversing the
// `forbid` array on the same input and watching the reported order not move. Emission order IS
// observable here, so the loops emit in source order, which is also what upstream's own corpus
// asserts for its three-finding case.
//
// Two entity kinds and three findings, because a single-kind case cannot see the difference and a
// two-finding case of one kind cannot either.
func TestNoUnescapedEntitiesReportsInSourceOrder(t *testing.T) {
	t.Parallel()

	const source = "const a = <div>Multiple errors: '>> default parser</div>;"

	result := rule_testing.Run(t, NoUnescapedEntities, "component.tsx", source)
	rule_testing.ExpectFindings(t, result,
		"unescapedEntityAlts", "unescapedEntityAlts", "unescapedEntityAlts")

	// Upstream's corpus lists this case's errors in exactly this order.
	wantSpans := []string{"'", ">", ">"}
	for index, want := range wantSpans {
		reported := source[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
		if reported != want {
			t.Fatalf("finding %d covered %q, expected %q", index, reported, want)
		}
	}
	for index := 1; index < len(result.Diagnostics); index++ {
		if result.Diagnostics[index-1].Range.Pos() >= result.Diagnostics[index].Range.Pos() {
			t.Fatalf("findings are not in source order: %d at %d follows %d",
				index, result.Diagnostics[index].Range.Pos(), result.Diagnostics[index-1].Range.Pos())
		}
	}
}

// TestNoUnescapedEntitiesUsesTheDefaultsWhenHandedNilOptions is the fixture that bypasses the
// decoder.
//
// A rule configured as a bare "error" is handed nil, because `configuration.OptionsRegistry.Decode` turns
// a decode of empty input into nil for a rule whose options are not required. That is the path the
// entire real tree takes, and a rule reaching the scan with an empty entity list there registers on
// every file and finds nothing, which looks exactly like a clean tree. Every other fixture in this
// file reaches the rule through options, so nothing else can see it.
func TestNoUnescapedEntitiesUsesTheDefaultsWhenHandedNilOptions(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(
		t, NoUnescapedEntities, "component.tsx", "const a = <div>'</div>;", nil)
	rule_testing.ExpectFindings(t, result, "unescapedEntityAlts")
}

// TestNoUnescapedEntitiesSeparatesAnAbsentForbidFromAnEmptyOne is the decoder's load-bearing case.
//
// Absent means upstream's four defaults; `[]` means nothing is forbidden and the rule is silent.
// A decoder holding a plain slice collapses the two into one nil and would report on a
// configuration that explicitly asked for silence. Measured on the installed rule: `{forbid: []}`
// is clean on source the defaults report on.
func TestNoUnescapedEntitiesSeparatesAnAbsentForbidFromAnEmptyOne(t *testing.T) {
	t.Parallel()

	const source = "const a = <div>a'b\"c</div>;"

	absent := decodeNoUnescapedEntitiesOptionsForTest(t, `{}`)
	rule_testing.ExpectFindings(t,
		rule_testing.RunWithOptions(t, NoUnescapedEntities, "component.tsx", source, absent),
		"unescapedEntityAlts", "unescapedEntityAlts")

	empty := decodeNoUnescapedEntitiesOptionsForTest(t, `{"forbid":[]}`)
	rule_testing.ExpectClean(t,
		rule_testing.RunWithOptions(t, NoUnescapedEntities, "component.tsx", source, empty))
}

// TestNoUnescapedEntitiesDeclinesTheForbidEntriesUpstreamCannotUse pins three decoder declines.
//
// A multi-character string can never match, because upstream compares one character against the
// whole string; measured clean on `{forbid: ["oo"]}` over `<div>foo bar</div>`. An object with no
// `char` never matches either, measured clean. An object with no `alternatives` is the deliberate
// divergence: upstream CRASHES on it with "Cannot read properties of undefined (reading 'map')",
// measured, and this drops it instead.
func TestNoUnescapedEntitiesDeclinesTheForbidEntriesUpstreamCannotUse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		options string
		source  string
	}{
		{"multiCharacterString", `{"forbid":["oo"]}`, "const a = <div>foo bar</div>;"},
		{"emptyString", `{"forbid":[""]}`, "const a = <div>foo bar</div>;"},
		{"objectWithoutChar", `{"forbid":[{"alternatives":["x"]}]}`, "const a = <div>a&b</div>;"},
		{"objectWithoutAlternatives", `{"forbid":[{"char":"&"}]}`, "const a = <div>a&b</div>;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := decodeNoUnescapedEntitiesOptionsForTest(t, testCase.options)
			rule_testing.ExpectClean(t,
				rule_testing.RunWithOptions(t, NoUnescapedEntities, "component.tsx", testCase.source, options))
		})
	}
}

// TestNoUnescapedEntitiesReportsAnObjectEntryWithNoAlternatives pins the message upstream renders
// for an empty alternatives list.
//
// `{char: "&", alternatives: []}` is a different configuration from `{char: "&"}` and from `"&"`,
// and upstream reports it as `unescapedEntityAlts` with nothing to offer, rendering as
// "`&` can be escaped with ." Reproduced including the space before the period.
func TestNoUnescapedEntitiesReportsAnObjectEntryWithNoAlternatives(t *testing.T) {
	t.Parallel()

	options := decodeNoUnescapedEntitiesOptionsForTest(t, `{"forbid":[{"char":"&","alternatives":[]}]}`)

	result := rule_testing.RunWithOptions(
		t, NoUnescapedEntities, "component.tsx", "const a = <div>a&b</div>;", options)
	rule_testing.ExpectFindings(t, result, "unescapedEntityAlts")

	const want = "`&` can be escaped with ."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Fatalf("message text:\n got  %q\n want %q", got, want)
	}
	if got := len(result.Diagnostics[0].Suggestions); got != 0 {
		t.Fatalf("an empty alternatives list offers nothing, got %d suggestions", got)
	}
}

// TestNoUnescapedEntitiesIgnoresTextThatIsNotJsxText pins the false-positive class the listener
// choice creates.
//
// Upstream's selector names `Literal` alongside `JSXText`, which reads as "strings inside JSX are
// checked too" and is the opposite of the truth: the `isJSX(node.parent)` guard rejects every
// literal, because a Literal is never a direct child of a JSX element or fragment. Measured both
// ways on the installed rule, and both of these are clean there.
func TestNoUnescapedEntitiesIgnoresTextThatIsNotJsxText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{"attributeString", `const a = <div title="a>b" />;`},
		{"expressionContainerString", `const a = <div>{"has > and quote \""}</div>;`},
		{"plainString", `const a = "he said \"hi\"";`},
		{"templateLiteral", "const a = `it's here`;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUnescapedEntities, "component.tsx", testCase.source))
		})
	}
}

// TestNoUnescapedEntitiesScansBytesWithoutMisreadingMultiByteText guards the byte scan.
//
// The scan walks bytes rather than runes, which is safe only because every forbidden character is
// ASCII and no UTF-8 continuation byte is below 0x80. This pins that: text carrying an accented
// letter and an astral emoji reports exactly the two ASCII quotes and points at them, so a scan
// that misread a continuation byte would fail here rather than in the tree.
func TestNoUnescapedEntitiesScansBytesWithoutMisreadingMultiByteText(t *testing.T) {
	t.Parallel()

	const source = "const a = <div>café \"q\" \U0001F600 'y'</div>;"

	result := rule_testing.Run(t, NoUnescapedEntities, "component.tsx", source)
	rule_testing.ExpectFindings(t, result,
		"unescapedEntityAlts", "unescapedEntityAlts", "unescapedEntityAlts", "unescapedEntityAlts")

	wantSpans := []string{`"`, `"`, "'", "'"}
	for index, want := range wantSpans {
		reported := source[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
		if reported != want {
			t.Fatalf("finding %d covered %q, expected %q", index, reported, want)
		}
	}
}

// TestNoUnescapedEntitiesLeavesEscapedEntitiesAlone is the discrimination the whole rule rests on.
//
// The node's raw source is what is scanned, and an already-escaped entity is a sequence of harmless
// characters in it. This would still pass if the rule scanned a decoded value instead, on `&gt;`
// alone, because `>` decoded would report. It is the `&amp;` case that separates them: decoded it
// is `&`, which is not forbidden by default, so a decoded scan reports on the `&gt;` and not on
// this. Both are here for that reason.
func TestNoUnescapedEntitiesLeavesEscapedEntitiesAlone(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoUnescapedEntities, "component.tsx", "const a = <div>&gt;&amp;&quot;&apos;&#125;</div>;"))
}

// applyOneSuggestion rewrites the source with one suggestion's fixes and returns the result.
//
// `rule_testing` applies fixes and has no equivalent for suggestions, so this exists rather than being
// imported. Back to front, the same order the fix engine uses, so an earlier edit cannot move the
// offsets a later one was computed against.
func applyOneSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()

	ordered := append([]rule.Fix(nil), suggestion.Fixes...)
	sort.Slice(ordered, func(a, b int) bool { return ordered[a].Range.Pos() > ordered[b].Range.Pos() })

	rewritten := source
	for _, fix := range ordered {
		if fix.Range.Pos() < 0 || fix.Range.End() > len(rewritten) || fix.Range.Pos() > fix.Range.End() {
			t.Fatalf("suggestion proposed an out-of-range edit %d..%d over %d bytes",
				fix.Range.Pos(), fix.Range.End(), len(rewritten))
		}
		rewritten = rewritten[:fix.Range.Pos()] + fix.Text + rewritten[fix.Range.End():]
	}
	return rewritten
}

// TestNoUnescapedEntitiesRepairsCoverTheWholeNode pins the ONE thing that separates upstream's edit
// from the obvious simplification.
//
// Replacing just the offending character produces byte-identical output for any single applied
// suggestion, so nothing that asserts a rewritten file can tell the two apart. A mutant scoping the
// fix to the character survived the entire suite for exactly that reason.
//
// The difference is overlap. Upstream replaces the whole JsxText node, so two findings inside one
// node carry the SAME range and the engine sees them as mutually exclusive. Measured on the
// installed rule: `<div>a'b"c</div>` yields two findings whose first suggestions both carry
// range [13,18], the node's own span. Character-scoped fixes would be disjoint and independently
// applicable, which is a different proposal to an engine that reasons about overlap.
//
// So this asserts the range rather than the output, because the output cannot see it.
func TestNoUnescapedEntitiesRepairsCoverTheWholeNode(t *testing.T) {
	t.Parallel()

	const source = "const a = <div>a'b\"c</div>;"

	result := rule_testing.Run(t, NoUnescapedEntities, "component.tsx", source)
	rule_testing.ExpectFindings(t, result, "unescapedEntityAlts", "unescapedEntityAlts")

	// The JsxText node is `a'b"c`, which the parser places at offsets 15..20.
	const wantPos, wantEnd = 15, 20
	for findingIndex, diagnostic := range result.Diagnostics {
		for suggestionIndex, suggestion := range diagnostic.Suggestions {
			if len(suggestion.Fixes) != 1 {
				t.Fatalf("finding %d suggestion %d: expected one fix, got %d",
					findingIndex, suggestionIndex, len(suggestion.Fixes))
			}
			fixRange := suggestion.Fixes[0].Range
			if fixRange.Pos() != wantPos || fixRange.End() != wantEnd {
				t.Fatalf("finding %d suggestion %d: fix covers %d..%d, expected the whole JsxText %d..%d",
					findingIndex, suggestionIndex, fixRange.Pos(), fixRange.End(), wantPos, wantEnd)
			}
		}
	}

	// And the two findings' repairs overlap, which is the property the range choice exists for.
	firstRange := result.Diagnostics[0].Suggestions[0].Fixes[0].Range
	secondRange := result.Diagnostics[1].Suggestions[0].Fixes[0].Range
	if firstRange.Pos() != secondRange.Pos() || firstRange.End() != secondRange.End() {
		t.Fatalf("two findings in one JsxText proposed non-identical ranges %d..%d and %d..%d",
			firstRange.Pos(), firstRange.End(), secondRange.Pos(), secondRange.End())
	}
}
