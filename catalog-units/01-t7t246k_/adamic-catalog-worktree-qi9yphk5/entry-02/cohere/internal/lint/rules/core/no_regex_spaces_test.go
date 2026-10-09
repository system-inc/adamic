package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// regexSpacesFile is where the fixtures pretend to live.
const regexSpacesFile = "/repository/source/RegexSpaces.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_regex_spaces.rs`:
// 48 pass, 26 fail, and the snapshot records 26 diagnostics from those 26 inputs, so one finding
// per input is right here rather than assumed. The extractor confirms one Tester block, no options
// on any case, and a 26-pair fix vector.
//
// One finding per input is not a coincidence of the corpus, it is the rule: at most one run of
// spaces is ever reported per regex, so no input can report twice. See the rule's own doc comment.
//
// Copied because a fixture a porter invents encodes the same belief as the port, and the case that
// catches a bug is the one nobody would think to write. Two of upstream's pass cases are byte
// identical duplicates (`var foo = / /;` appears at both index 3 and index 17); both are kept, so
// this file holds the corpus rather than a tidied reading of it.
func TestNoRegexSpacesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a literal that is nothing but two spaces", "var foo = /  /;"},
		{"two spaces between words", "var foo = /bar  baz/;"},
		{"four spaces between words", "var foo = /bar    baz/;"},
		{"a run among single spaces", "var foo = / a b  c d /;"},
		{"a trailing run through the call form", "var foo = RegExp(' a b c d  ');"},
		{"a run through the call form", "var foo = RegExp('bar    baz');"},
		{"a run through the constructor", "var foo = new RegExp('bar    baz');"},
		{"a run before a counted quantifier", "var foo = /bar   {3}baz/;"},
		{"a run before an optional quantifier", "var foo = /bar    ?baz/;"},
		{"a run before a star, through the constructor", "var foo = new RegExp('bar   *baz')"},
		{"a run before a plus, through the call form", "var foo = RegExp('bar   +baz')"},
		{"a trailing run through the constructor", "var foo = new RegExp('bar    ');"},
		{"a run after an escaped backslash", `var foo = /bar\\  baz/;`},
		{"a run inside a non capturing group", "var foo = /(?:  )/;"},
		{"a run inside a lookahead", "var foo = RegExp('^foo(?=   )');"},
		{"a run after a leading escaped backslash", `var foo = /\\  /`},
		{"a run after a space and an escaped backslash", `var foo = / \\  /`},
		{"the first of two runs", "var foo = /  foo   /;"},
		{"a run after an escaped backslash and a d", `var foo = new RegExp('\\d  ')`},
		{"a run after an escaped backslash and u0041", `var foo = RegExp('\\u0041   ')`},
		{"a run after a character class", "var foo = /[   ]  /;"},
		{"a run before a character class", "var foo = /  [   ] /;"},
		{"a run before a class, through the call form", "var foo = RegExp('  [ ]');"},
		{"a run after a v flag nested class", "var foo = /[[    ]    ]    /v;"},
		{"a run after a class, through the constructor", "var foo = new RegExp('[   ]  ');"},
		{"a v flag nested class through the constructor", "var foo = new RegExp('[[    ]    ]    ', 'v');"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, testCase.sourceText), "multipleSpaces")
		})
	}
}

// The clean cases are the whole discrimination, and they cluster into the distinctions the rule
// actually makes.
//
// Three use real tab characters rather than escapes, which pins that the rule is about the space
// character and does not generalize to whitespace. Several spell a space as ` ` or `\x20`,
// which the rule deliberately leaves alone: the objection is to a run nobody can count by eye, and
// an escape is already counted. The quantifier group (`/  +/`, `/  ?/`, `/  */`, `/  {2}/`) is the
// reason the walk carries depth at all, since a quantified space means something other than a
// repetition someone lost track of. The class group is the same argument one level in.
//
// `new RegExp('  ', flags)` and `new RegExp('[[abc]  ]', flags + 'v')` are the indeterminate flag
// guard, and they are pass cases rather than fail cases: with flags unknown the parse is unknown,
// and reporting on a guess would be worse than staying quiet.
func TestNoRegexSpacesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a literal with no spaces", "var foo = /foo/;"},
		{"a call form with no spaces", "var foo = RegExp('foo')"},
		{"a single space", "var foo = / /;"},
		{"a single space through the call form", "var foo = RegExp(' ')"},
		{"single spaces throughout", "var foo = / a b c d /;"},
		{"a space already carrying a quantifier", "var foo = /bar {3}baz/g;"},
		{"a quantified space through the call form", "var foo = RegExp('bar {3}baz', 'g')"},
		{"a quantified space through the constructor", "var foo = new RegExp('bar {3}baz')"},
		{"real tabs in a literal", "var foo = /bar\t\t\tbaz/;"},
		{"real tabs through the call form", "var foo = RegExp('bar\t\t\tbaz');"},
		{"real tabs through the constructor", "var foo = new RegExp('bar\t\t\tbaz');"},
		{"two spaces governed by a plus", "var foo = /  +/;"},
		{"two spaces governed by a question mark", "var foo = /  ?/;"},
		{"two spaces governed by a star", "var foo = /  */;"},
		{"two spaces governed by a counted quantifier", "var foo = /  {2}/;"},
		{"the same under the v flag", "var foo = /  {2}/v;"},
		{"a single space again", "var foo = / /;"},
		{"a space, an escaped backslash, a space", `var foo = /bar \\ baz/;`},
		{"escaped backslashes separating single spaces", `var foo = /bar\\ \\ baz/;`},
		{"a spelled out space between spaces", `var foo = /bar \\u0020 baz/;`},
		{"two spelled out spaces", `var foo = /bar \\u0020\\u0020baz/;`},
		{"an escaped backslash through the constructor", `var foo = new RegExp('bar \\ baz')`},
		{"escaped backslashes through the constructor", `var foo = new RegExp('bar\\ \\ baz')`},
		{"a doubled escaped backslash", `var foo = new RegExp('bar \\\\ baz')`},
		{"two spelled out spaces through the constructor", `var foo = new RegExp('bar\\u0020\\u0020baz')`},
		{"a spelled out space through the constructor", `var foo = new RegExp('bar \\u0020 baz')`},
		{"indeterminate flags", "new RegExp('  ', flags)"},
		{"two spaces inside a class", "var foo = /[  ]/;"},
		{"three spaces inside a class", "var foo = /[   ]/;"},
		{"a class between single spaces", "var foo = / [  ] /;"},
		{"two classes between single spaces", "var foo = / [  ] [  ] /;"},
		{"a class through the constructor", "var foo = new RegExp('[  ]');"},
		{"a wider class through the constructor", "var foo = new RegExp('[   ]');"},
		{"a class between spaces through the constructor", "var foo = new RegExp(' [  ] ');"},
		{"two classes through the call form", "var foo = RegExp(' [  ] [  ] ');"},
		{"an escaped bracket, so no class opens", `var foo = new RegExp(' \[   ');`},
		{"escaped brackets on both sides", `var foo = new RegExp(' \[   \] ');`},
		{"a v flag set escape holding spaces", `var foo = /[\q{    }]/v;`},
		{"an unterminated class", "var foo = new RegExp('[  ');"},
		{"indeterminate flags with a nested class", "new RegExp('[[abc]  ]', flags + 'v')"},
		{"two hex escaped spaces", `var foo = new RegExp('\\x20\\x20');`},
		{"two unicode escaped spaces", `var foo = new RegExp('\\u0020\\u0020');`},
		{"a space then a hex escaped space", `var foo = new RegExp(' \\x20');`},
		{"a hex escaped space then a space", `var foo = new RegExp('\\x20 ');`},
		{"hex escapes in a literal", `var foo = /\x20\x20/;`},
		{"unicode escapes in a literal", `var foo = /\u0020\u0020/;`},
		{"two hex escaped spaces in double quotes", `var foo = new RegExp("\\x20\\x20");`},
		{"two unicode escaped spaces in double quotes", `var foo = new RegExp("\\u0020\\u0020");`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, testCase.sourceText))
		})
	}
}

// TestNoRegexSpacesReportsTheRunItself asserts where every finding points, which ExpectFindings
// cannot see.
//
// This rule's whole output is a position: the message says nothing about which run it means, and a
// fix that writes the right text over the wrong span produces source that still parses. So the span
// is checked by slicing the source with the finding's own range and comparing the text, for every
// case in the corpus rather than for a sample.
//
// The expected offsets are derived from oxc's snapshot columns rather than counted by hand, and
// every one of them slices to a run of spaces and nothing else. That self-check is why they are
// trustworthy: a wrong offset would slice to something with a non-space in it.
//
// The last case is the one place we do not match oxc. See the rule's doc comment.
func TestNoRegexSpacesReportsTheRunItself(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantStart  int
		wantText   string
	}{
		{"var foo = /  /;", 11, "  "},
		{"var foo = /bar  baz/;", 14, "  "},
		{"var foo = /bar    baz/;", 14, "    "},
		{"var foo = / a b  c d /;", 15, "  "},
		{"var foo = RegExp(' a b c d  ');", 26, "  "},
		{"var foo = RegExp('bar    baz');", 21, "    "},
		{"var foo = new RegExp('bar    baz');", 25, "    "},
		{"var foo = /bar   {3}baz/;", 14, "  "},
		{"var foo = /bar    ?baz/;", 14, "   "},
		{"var foo = new RegExp('bar   *baz')", 25, "  "},
		{"var foo = RegExp('bar   +baz')", 21, "  "},
		{"var foo = new RegExp('bar    ');", 25, "    "},
		{`var foo = /bar\\  baz/;`, 16, "  "},
		{"var foo = /(?:  )/;", 14, "  "},
		{"var foo = RegExp('^foo(?=   )');", 25, "   "},
		{`var foo = /\\  /`, 13, "  "},
		{`var foo = / \\  /`, 14, "  "},
		{"var foo = /  foo   /;", 11, "  "},
		{`var foo = new RegExp('\\d  ')`, 25, "  "},
		{`var foo = RegExp('\\u0041   ')`, 25, "   "},
		{"var foo = /[   ]  /;", 16, "  "},
		{"var foo = /  [   ] /;", 11, "  "},
		{"var foo = RegExp('  [ ]');", 18, "  "},
		{"var foo = /[[    ]    ]    /v;", 23, "    "},
		{"var foo = new RegExp('[   ]  ');", 27, "  "},
		// oxc reports this at offset 29, inside the nested class, because it never passes the
		// v flag to its constructor parser. 34 is the trailing run, which is what the same
		// pattern written as a literal reports and what ESLint reports for both spellings.
		{"var foo = new RegExp('[[    ]    ]    ', 'v');", 34, "    "},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if result.Diagnostics[0].Range.Pos() != testCase.wantStart || reported != testCase.wantText {
				t.Errorf("finding at [%d,%d) = %q, want start %d and %q",
					result.Diagnostics[0].Range.Pos(), result.Diagnostics[0].Range.End(),
					reported, testCase.wantStart, testCase.wantText)
			}
		})
	}
}

// TestNoRegexSpacesFixes is oxc's fix vector, 26 before and after pairs, asserted by applying the
// repair rather than by comparing the fix's text.
//
// The distinction matters here more than in most rules: a fix writing ` {4}` over the wrong four
// spaces produces a different file and an identical fix text, so a text comparison cannot tell the
// two apart. That is the shape of the one case where we deliberately differ from upstream.
//
// The repair is a fix rather than a suggestion because it preserves what the pattern matches: a run
// of n spaces and ` {n}` accept exactly the same input. Nothing here picks between valid answers,
// so nothing needs a human to choose it.
func TestNoRegexSpacesFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSource string
	}{
		{"var foo = /  /;", "var foo = / {2}/;"},
		{"var foo = /bar  baz/;", "var foo = /bar {2}baz/;"},
		{"var foo = /bar    baz/;", "var foo = /bar {4}baz/;"},
		{"var foo = / a b  c d /;", "var foo = / a b {2}c d /;"},
		{"var foo = RegExp(' a b c d  ');", "var foo = RegExp(' a b c d {2}');"},
		{"var foo = RegExp('bar    baz');", "var foo = RegExp('bar {4}baz');"},
		{"var foo = new RegExp('bar    baz');", "var foo = new RegExp('bar {4}baz');"},
		{"var foo = /bar   {3}baz/;", "var foo = /bar {2} {3}baz/;"},
		{"var foo = /bar    ?baz/;", "var foo = /bar {3} ?baz/;"},
		{"var foo = new RegExp('bar   *baz')", "var foo = new RegExp('bar {2} *baz')"},
		{"var foo = RegExp('bar   +baz')", "var foo = RegExp('bar {2} +baz')"},
		{"var foo = new RegExp('bar    ');", "var foo = new RegExp('bar {4}');"},
		{`var foo = /bar\\  baz/;`, `var foo = /bar\\ {2}baz/;`},
		{"var foo = /(?:  )/;", "var foo = /(?: {2})/;"},
		{"var foo = RegExp('^foo(?=   )');", "var foo = RegExp('^foo(?= {3})');"},
		{`var foo = /\\  /`, `var foo = /\\ {2}/`},
		{`var foo = / \\  /`, `var foo = / \\ {2}/`},
		{"var foo = /  foo   /;", "var foo = / {2}foo   /;"},
		{`var foo = new RegExp('\\d  ')`, `var foo = new RegExp('\\d {2}')`},
		{`var foo = RegExp('\\u0041   ')`, `var foo = RegExp('\\u0041 {3}')`},
		{"var foo = /[   ]  /;", "var foo = /[   ] {2}/;"},
		{"var foo = /  [   ] /;", "var foo = / {2}[   ] /;"},
		{"var foo = RegExp('  [ ]');", "var foo = RegExp(' {2}[ ]');"},
		{"var foo = /[[    ]    ]    /v;", "var foo = /[[    ]    ] {4}/v;"},
		{"var foo = new RegExp('[   ]  ');", "var foo = new RegExp('[   ] {2}');"},
		// oxc's fix vector bakes its own v flag bug in here, writing
		// `new RegExp('[[    ] {4}]    ', 'v')`. Ours rewrites the trailing run, matching the
		// literal spelling and ESLint.
		{"var foo = new RegExp('[[    ]    ]    ', 'v');", "var foo = new RegExp('[[    ]    ] {4}', 'v');"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, testCase.sourceText), testCase.wantSource)
		})
	}
}

// The cases below are ours, not upstream's. Each exists because reading our own code raised a
// question the imported corpus does not answer.

// TestNoRegexSpacesReportsOnlyTheFirstRun pins the behavior the upstream field name hides.
//
// oxc calls the accumulator `last_space_span`, and it does not hold the last run. Once a run of two
// or more has latched, every later run is discarded, so the finding is on the FIRST such run. The
// single space bookkeeping exists only to step past lone spaces on the way to one.
//
// Upstream's corpus contains exactly one input that can tell the two readings apart
// (`/  foo   /`), so a port that tracked the last run instead would fail that one case and pass the
// other 25. These make the distinction explicit rather than resting on a single imported case, and
// the last two are the ones upstream has no equivalent of: a lone space between two real runs, and
// three runs where the middle is widest.
func TestNoRegexSpacesReportsOnlyTheFirstRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantStart  int
		wantText   string
	}{
		{"var foo = /a  b   c/;", 12, "  "},
		{"var foo = /a   b  c/;", 12, "   "},
		{"var foo = /a  b c   d/;", 12, "  "},
		{"var foo = /a  b    c  d/;", 12, "  "},
		{"var foo = / a  b /;", 13, "  "},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if result.Diagnostics[0].Range.Pos() != testCase.wantStart || reported != testCase.wantText {
				t.Errorf("finding at [%d,%d) = %q, want start %d and %q",
					result.Diagnostics[0].Range.Pos(), result.Diagnostics[0].Range.End(),
					reported, testCase.wantStart, testCase.wantText)
			}
		})
	}
}

// TestNoRegexSpacesAgreesAcrossSpellings is the case upstream gets wrong and the reason we do not
// copy its fixture.
//
// The same pattern written three ways has to report in the same place, because the three spell one
// regex. oxc hardcodes a nil flags text into its constructor parser, so the `v` flag reaches the
// literal path and not the constructor path, and its own snapshot shows the two disagreeing by five
// bytes. Ours passes the real flags on both paths, so the three agree.
//
// The middle case is the control: without `v`, classes do not nest, `[[    ]` closes at the first
// bracket, and the depth zero run genuinely is the earlier one. A port that simply ignored flags
// everywhere would pass that case and fail the other two.
func TestNoRegexSpacesAgreesAcrossSpellings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
		wantStart  int
	}{
		{"as a literal with v", "var foo = /[[    ]    ]    /v;", "    ", 23},
		{"through the constructor with v", "var foo = new RegExp('[[    ]    ]    ', 'v');", "    ", 34},
		{"through the constructor without v", "var foo = new RegExp('[[    ]    ]    ');", "    ", 29},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if result.Diagnostics[0].Range.Pos() != testCase.wantStart || reported != testCase.wantText {
				t.Errorf("finding at [%d,%d) = %q, want start %d and %q",
					result.Diagnostics[0].Range.Pos(), result.Diagnostics[0].Range.End(),
					reported, testCase.wantStart, testCase.wantText)
			}
		})
	}
}

// TestNoRegexSpacesShapesOurCorpusDoesNotCover comes from reading our tree rather than upstream's.
//
// A regex literal preceded by a comment is the trap `no_empty_character_class` documents: a node's
// Pos sits before its leading trivia, so a start derived from Pos rather than from End puts the
// finding on the comment, on a line no suppression directive the author can write is able to reach.
// Upstream's corpus has no commented case at all, so nothing imported can catch it.
//
// The remaining cases are call shapes our tree contains and upstream's corpus does not: a
// parenthesized callee, a member call that merely ends in the right name, a shadowing local, and a
// call with no arguments.
func TestNoRegexSpacesShapesOurCorpusDoesNotCover(t *testing.T) {
	t.Parallel()

	t.Run("a literal preceded by a comment reports inside the pattern", func(t *testing.T) {
		t.Parallel()
		sourceText := "// a comment about the pattern\nvar foo = /bar  baz/;"
		result := rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, sourceText)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
		}
		reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != "  " {
			t.Errorf("finding = %q, want the two spaces inside the pattern", reported)
		}
	})

	t.Run("a parenthesized callee is still the RegExp constructor", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t,
			rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, "var foo = new (RegExp)('bar  baz');"), "multipleSpaces")
	})

	silent := []struct {
		name       string
		sourceText string
	}{
		{"a member call ending in the name", "var foo = my.RegExp('bar  baz');"},
		{"a different callee entirely", "var foo = NotRegExp('bar  baz');"},
		{"a call with no arguments", "var foo = RegExp();"},
		{"a non literal pattern argument", "var foo = new RegExp(pattern);"},
		{"a template literal pattern argument", "var foo = new RegExp(`bar  baz`);"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, testCase.sourceText))
		})
	}
}

// TestNoRegexSpacesParsesTheCookedPatternAndReportsRawOffsets is the discrimination this port got
// wrong on its first pass, and the one nothing in the imported corpus explains.
//
// Two of upstream's pass cases are `new RegExp(' \[   ')` and `new RegExp(' \[   \] ')`. Both hold
// three adjacent spaces at what looks like depth zero, and a port that hands the RAW body to the
// walk reports them, because raw `\[` reads as a regex escape and no class opens. Cooked, the `\[`
// is a JS string escape, the pattern is ` [   `, and that bracket opens a class running off the end
// of the pattern. Upstream and ESLint are both silent for that reason, and the corpus records the
// silence without recording the cause.
//
// The complementary half is the literal spelling, which has no JS string layer at all: `/\[   /`
// really is an escaped bracket, really has no class, and really does report. So the same six
// characters are clean in one spelling and a finding in the other, and only a rule that cooks the
// constructor's argument gets both.
//
// The offsets then have to come back out raw. `new RegExp('\\d  ')` cooks two bytes shorter than it
// was written, so a cooked offset applied to the file lands two bytes early and the fix eats the
// `d`. That case is in the imported corpus and its fix pair is what pins it, but it pins it only
// once; these make the general rule explicit.
func TestNoRegexSpacesParsesTheCookedPatternAndReportsRawOffsets(t *testing.T) {
	t.Parallel()

	t.Run("an escaped bracket cooks to a real class opener", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoRegexSpaces, regexSpacesFile,
			`var foo = new RegExp(' \[   ');`))
	})

	t.Run("the same pattern as a literal has no string layer and reports", func(t *testing.T) {
		t.Parallel()
		sourceText := `var foo = /\[   /;`
		result := rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, sourceText)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
		}
		reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != "   " {
			t.Errorf("finding = %q, want the three spaces", reported)
		}
	})

	t.Run("a cooked escape shifts the raw offset", func(t *testing.T) {
		t.Parallel()
		// Raw body `\\d  ` is five bytes; cooked `\d  ` is four. The run sits at cooked index 2
		// and raw index 3, so a rule reporting the cooked offset would slice `d ` instead.
		sourceText := `var foo = new RegExp('\\d  ')`
		result := rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, sourceText)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
		}
		reported := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != "  " {
			t.Errorf("finding = %q, want the two spaces and not the character before them", reported)
		}
	})

	t.Run("a unicode escape before the run shifts it further", func(t *testing.T) {
		t.Parallel()
		// Raw `\\u0041   ` is ten bytes, cooked `A   ` is nine.
		rule_testing.ExpectFixedSource(t,
			rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, `var foo = RegExp('\\u0041   ')`),
			`var foo = RegExp('\\u0041 {3}')`)
	})

	t.Run("a cooked escape that is itself a space is not a run", func(t *testing.T) {
		t.Parallel()
		// `\x20\x20` cooks to two real adjacent spaces, so a rule walking only the cooked value
		// would report a run nobody wrote. The raw gate is what refuses it, and this is why that
		// gate runs on the raw body rather than on the cooked one.
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoRegexSpaces, regexSpacesFile,
			`var foo = new RegExp('\x20\x20');`))
	})
}

// TestNoRegexSpacesDropsAPatternItCannotMap covers the one branch a mutation sweep found nothing
// else reaching: a string literal holding an escape form the cooked-to-raw mapping does not model.
//
// A line continuation is the reachable case. `'a\<newline>  b'` cooks to `a  b`, so the backslash
// and the newline together produce no cooked character at all, and the one-raw-escape to
// one-cooked-character correspondence the mapping is built on does not hold across it. Every offset
// after that point would be shifted, which moves a finding rather than losing one, so the case is
// dropped instead of guessed at.
//
// This is silence by design rather than a gap worth closing later. The run really is there and a
// perfect rule would report it, but a finding at a position derived from a correspondence already
// known to be wrong is worse than no finding: the fix would rewrite bytes nobody pointed at.
func TestNoRegexSpacesDropsAPatternItCannotMap(t *testing.T) {
	t.Parallel()

	// A real line continuation inside the argument, written as an actual newline in the source.
	// Here it sits BEFORE the run, so the walk desynchronizes while cooked characters remain and
	// the loop's own bail is what fires.
	t.Run("a continuation before the run", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoRegexSpaces, regexSpacesFile,
			"var foo = new RegExp('bar\\\n    baz');"))
	})

	// And here it sits AFTER the run, which is a different branch and the only input that reaches
	// the final length check. Cooked is exhausted while raw still holds the continuation, so the
	// loop never notices: without the check the offsets look complete and well formed, and the rule
	// reports a run whose position rests on a correspondence that stopped being true. A mutation
	// sweep is what found this, and the case before it does not cover it.
	t.Run("a continuation after the run", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.Run(t, NoRegexSpaces, regexSpacesFile,
			"var foo = new RegExp('a  b\\\n');"))
	})
}

// TestNoRegexSpacesMapsThroughEveryEscapeForm exists because a mutation sweep found the escape
// width table unmeasured: changing `\x41` from four raw bytes to three, or `\u0041` from six to
// five, left every imported fixture green.
//
// The corpus is why. It is full of `\\x20` and `\\u0020`, which are an escaped BACKSLASH followed
// by ordinary characters, so they exercise the two-byte generic branch and never the hex or
// unicode branches at all. A real single-backslash escape sitting before a run of spaces is a shape
// upstream simply does not test, and it is the shape where a wrong width moves the finding: in
// `'\x41  b'` the run begins at raw index 4, and reading the escape as three bytes would point at
// the last digit of it instead.
//
// The brace form is here for the same reason and is the one whose width is not a constant at all.
func TestNoRegexSpacesMapsThroughEveryEscapeForm(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{
			"a hex escape before the run",
			`var foo = new RegExp('\x41  b');`,
			`var foo = new RegExp('\x41 {2}b');`,
		},
		{
			"a unicode escape before the run",
			`var foo = new RegExp('\u0041  b');`,
			`var foo = new RegExp('\u0041 {2}b');`,
		},
		{
			"a braced unicode escape before the run",
			`var foo = new RegExp('\u{41}  b');`,
			`var foo = new RegExp('\u{41} {2}b');`,
		},
		{
			"a single character escape before the run",
			`var foo = new RegExp('\t  b');`,
			`var foo = new RegExp('\t {2}b');`,
		},
		{
			"two escapes before the run",
			`var foo = new RegExp('\x41\u0041  c');`,
			`var foo = new RegExp('\x41\u0041 {2}c');`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoRegexSpaces, regexSpacesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "multipleSpaces")
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != "  " {
				t.Errorf("finding = %q, want exactly the two spaces", reported)
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantSource)
		})
	}
}
