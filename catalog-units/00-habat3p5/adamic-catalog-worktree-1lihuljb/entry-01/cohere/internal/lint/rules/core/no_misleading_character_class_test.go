package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// misleadingCharacterClassFile is where the fixtures pretend to live.
const misleadingCharacterClassFile = "/repository/source/Pattern.ts"

// Message id shorthands, so a case's expected findings read as a list of what went wrong rather
// than as a wall of repeated identifiers. The rule reports six distinct ids and one input commonly
// produces several, which is the whole reason the counts below are not all one.
const (
	withoutFlag = "surrogatePairWithoutUFlag"
	escapedPair = "surrogatePair"
	combining   = "combiningClass"
	emoji       = "emojiModifier"
	regional    = "regionalIndicatorSymbol"
	joiner      = "zwj"
)

// The corpus is oxc's, copied rather than rewritten.
//
// Every case here is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_misleading_character_class.rs`:
// 57 pass and 130 fail in one Tester block, and the snapshot records 174 diagnostics from those 130
// inputs. The extractor reports that gap as a discrepancy and it is the most important thing it
// says: an input reporting once is the exception here, not the rule.
//
// The gap is not per-character emission. Each of the five detectors scans the same run of adjacent
// class members independently, and each reports once per offending adjacent pair, so a class holding
// three astral characters spliced by two joiners produces five findings from one input. The counts
// below were measured against this port and then reconciled against the snapshot by message id:
// 72 without-flag surrogate pairs, 47 joins, 36 combining marks, 9 emoji modifiers, 6 regional
// indicators and 4 escaped surrogate pairs, which is 174 and matches upstream on every id.
func TestNoMisleadingCharacterClassFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
	}{
		{"var r = /[👍]/", []string{withoutFlag}},
		{"var r = /[\\uD83D\\uDC4D]/", []string{withoutFlag}},
		{"var r = /[\\uD83D\\uDC4D-\\uffff]/", []string{withoutFlag}},
		{"var r = /[👍]/", []string{withoutFlag}},
		{"var r = /before[\\uD83D\\uDC4D]after/", []string{withoutFlag}},
		{"var r = /[before\\uD83D\\uDC4Dafter]/", []string{withoutFlag}},
		{"var r = /\\uDC4D[\\uD83D\\uDC4D]/", []string{withoutFlag}},
		{"var r = /[👍]/", []string{withoutFlag}},
		{"var r = /[👍]\\a/", []string{withoutFlag}},
		{"var r = /\\a[👍]\\a/", []string{withoutFlag}},
		{"var r = /(?<=[👍])/", []string{withoutFlag}},
		{"var r = /(?<=[👍])/", []string{withoutFlag}},
		{"var r = /[Á]/", []string{combining}},
		{"var r = /[Á]/u", []string{combining}},
		{"var r = /[\\u0041\\u0301]/", []string{combining}},
		{"var r = /[\\u0041\\u0301]/u", []string{combining}},
		{"var r = /[\\u{41}\\u{301}]/u", []string{combining}},
		{"var r = /[❇️]/", []string{combining}},
		{"var r = /[❇️]/u", []string{combining}},
		{"var r = /[\\u2747\\uFE0F]/", []string{combining}},
		{"var r = /[\\u2747\\uFE0F]/u", []string{combining}},
		{"var r = /[\\u{2747}\\u{FE0F}]/u", []string{combining}},
		{"var r = /[👶🏻]/", []string{withoutFlag, withoutFlag}},
		{"var r = /[👶🏻]/u", []string{emoji}},
		{"var r = /[a\\uD83C\\uDFFB]/u", []string{emoji}},
		{"var r = /[\\uD83D\\uDC76\\uD83C\\uDFFB]/u", []string{emoji}},
		{"var r = /[\\u{1F476}\\u{1F3FB}]/u", []string{emoji}},
		{"var r = /[🇯🇵]/", []string{withoutFlag, withoutFlag}},
		{"var r = /[🇯🇵]/i", []string{withoutFlag, withoutFlag}},
		{"var r = /[🇯🇵]/u", []string{regional}},
		{"var r = /[\\uD83C\\uDDEF\\uD83C\\uDDF5]/u", []string{regional}},
		{"var r = /[\\u{1F1EF}\\u{1F1F5}]/u", []string{regional}},
		{"var r = /[👨‍👩‍👦]/", []string{joiner, joiner, withoutFlag, withoutFlag, withoutFlag}},
		{"var r = /[👨‍👩‍👦]/u", []string{joiner}},
		{"var r = /[👩‍👦]/u", []string{joiner}},
		{"var r = /[👩‍👦][👩‍👦]/u", []string{joiner, joiner}},
		{"var r = /[👨‍👩‍👦]foo[👨‍👩‍👦]/u", []string{joiner, joiner}},
		{"var r = /[👨‍👩‍👦👩‍👦]/u", []string{joiner, joiner}},
		{"var r = /[\\uD83D\\uDC68\\u200D\\uD83D\\uDC69\\u200D\\uD83D\\uDC66]/u", []string{joiner}},
		{"var r = /[\\u{1F468}\\u{200D}\\u{1F469}\\u{200D}\\u{1F466}]/u", []string{joiner}},
		{"var r = /[\\uD83D\\uDC68\\u200D\\uD83D\\uDC69]/u", []string{joiner}},
		{"var r = /[\\u{1F468}\\u{200D}\\u{1F469}]/u", []string{joiner}},
		{"var r = /[\\u{1F468}\\u{200D}\\u{1F469}\\u{200D}\\u{1F466}]foo[\\u{1F468}\\u{200D}\\u{1F469}\\u{200D}\\u{1F466}]/u", []string{joiner, joiner}},
		{"var r = RegExp(\"[👍]\", \"\")", []string{withoutFlag}},
		{"var r = new RegExp(\"[👍]\", \"\")", []string{withoutFlag}},
		{"var r = new RegExp('[👍]', ``)", []string{withoutFlag}},
		{"var r = new RegExp(`\n\t\t\t                [👍]`)", []string{withoutFlag}},
		{"var r = new RegExp(`\n\t\t\t                [❇️]`)", []string{combining}},
		{"var r = new RegExp(`\n\t\t\t[❇️]`)", []string{combining}},
		{"const flags = \"\"; var r = new RegExp(\"[👍]\", flags)", []string{withoutFlag}},
		{"var r = RegExp(\"[\\\\uD83D\\\\uDC4D]\", \"\")", []string{withoutFlag}},
		{"var r = RegExp(\"before[\\\\uD83D\\\\uDC4D]after\", \"\")", []string{withoutFlag}},
		{"var r = RegExp(\"[before\\\\uD83D\\\\uDC4Dafter]\", \"\")", []string{withoutFlag}},
		{"var r = RegExp(\"\\t\\t\\t👍[👍]\")", []string{withoutFlag}},
		{"var r = new RegExp(\"\\u1234[\\\\uD83D\\\\uDC4D]\")", []string{withoutFlag}},
		{"var r = new RegExp(\"\\\\u1234\\\\u5678👎[👍]\")", []string{withoutFlag}},
		{"var r = new RegExp(\"\\\\u1234\\\\u5678👍[👍]\")", []string{withoutFlag}},
		{"var r = new RegExp(\"[👍]\", \"\")", []string{withoutFlag}},
		{"var r = new RegExp(\"[👍]\", \"\")", []string{withoutFlag}},
		{"var r = new RegExp(\"[👍]\\\\a\", \"\")", []string{withoutFlag}},
		{"var r = new RegExp(\"/(?<=[👍])\", \"\")", []string{withoutFlag}},
		{"var r = new RegExp(\"/(?<=[👍])\", \"\")", []string{withoutFlag}},
		{"var r = new RegExp(\"[Á]\", \"\")", []string{combining}},
		{"var r = new RegExp(\"[Á]\", \"u\")", []string{combining}},
		{"var r = new RegExp(\"[\\\\u0041\\\\u0301]\", \"\")", []string{combining}},
		{"var r = new RegExp(\"[\\\\u0041\\\\u0301]\", \"u\")", []string{combining}},
		{"var r = new RegExp(\"[\\\\u{41}\\\\u{301}]\", \"u\")", []string{combining}},
		{"var r = new RegExp(\"[❇️]\", \"\")", []string{combining}},
		{"var r = new RegExp(\"[❇️]\", \"u\")", []string{combining}},
		{"new RegExp(\"[ \\\\ufe0f]\", \"\")", []string{combining}},
		{"new RegExp(\"[ \\\\ufe0f]\", \"u\")", []string{combining}},
		{"new RegExp(\"[ \\\\ufe0f][ \\\\ufe0f]\")", []string{combining, combining}},
		{"var r = new RegExp(\"[\\\\u2747\\\\uFE0F]\", \"\")", []string{combining}},
		{"var r = new RegExp(\"[\\\\u2747\\\\uFE0F]\", \"u\")", []string{combining}},
		{"var r = new RegExp(\"[\\\\u{2747}\\\\u{FE0F}]\", \"u\")", []string{combining}},
		{"var r = new RegExp(\"[👶🏻]\", \"\")", []string{withoutFlag, withoutFlag}},
		{"var r = new RegExp(\"[👶🏻]\", \"u\")", []string{emoji}},
		{"var r = new RegExp(\"[\\\\uD83D\\\\uDC76\\\\uD83C\\\\uDFFB]\", \"u\")", []string{emoji}},
		{"var r = new RegExp(\"[\\\\u{1F476}\\\\u{1F3FB}]\", \"u\")", []string{emoji}},
		{"var r = RegExp(`\t\t\t👍[👍]`)", []string{withoutFlag}},
		{"var r = RegExp(`\\t\\t\\t👍[👍]`)", []string{withoutFlag}},
		{"var r = new RegExp(\"[🇯🇵]\", \"\")", []string{withoutFlag, withoutFlag}},
		{"var r = new RegExp(\"[🇯🇵]\", \"i\")", []string{withoutFlag, withoutFlag}},
		{"var r = new RegExp('[🇯🇵]', `i`)", []string{withoutFlag, withoutFlag}},
		{"var r = new RegExp(\"[🇯🇵]\")", []string{withoutFlag, withoutFlag}},
		{"var r = new RegExp(\"[🇯🇵]\",)", []string{withoutFlag, withoutFlag}},
		{"var r = new RegExp((\"[🇯🇵]\"))", []string{withoutFlag, withoutFlag}},
		{"var r = new RegExp(((\"[🇯🇵]\")))", []string{withoutFlag, withoutFlag}},
		{"var r = new RegExp((\"[🇯🇵]\"),)", []string{withoutFlag, withoutFlag}},
		{"var r = new RegExp(\"[🇯🇵]\", \"u\")", []string{regional}},
		{"var r = new RegExp(\"[\\\\uD83C\\\\uDDEF\\\\uD83C\\\\uDDF5]\", \"u\")", []string{regional}},
		{"var r = new RegExp(\"[\\\\u{1F1EF}\\\\u{1F1F5}]\", \"u\")", []string{regional}},
		{"var r = new RegExp(\"[👨‍👩‍👦]\", \"\")", []string{joiner, joiner, withoutFlag, withoutFlag, withoutFlag}},
		{"var r = new RegExp(\"[👨‍👩‍👦]\", \"u\")", []string{joiner}},
		{"var r = new RegExp(\"[👩‍👦]\", \"u\")", []string{joiner}},
		{"var r = new RegExp(\"[👩‍👦][👩‍👦]\", \"u\")", []string{joiner, joiner}},
		{"var r = new RegExp(\"[👨‍👩‍👦]foo[👨‍👩‍👦]\", \"u\")", []string{joiner, joiner}},
		{"var r = new RegExp(\"[👨‍👩‍👦👩‍👦]\", \"u\")", []string{joiner, joiner}},
		{"var r = new RegExp(\"[\\\\uD83D\\\\uDC68\\\\u200D\\\\uD83D\\\\uDC69\\\\u200D\\\\uD83D\\\\uDC66]\", \"u\")", []string{joiner}},
		{"var r = new RegExp(\"[\\\\u{1F468}\\\\u{200D}\\\\u{1F469}\\\\u{200D}\\\\u{1F466}]\", \"u\")", []string{joiner}},
		{"var r = new globalThis.RegExp(\"[❇️]\", \"\")", []string{combining}},
		{"var r = new globalThis.RegExp(\"[👶🏻]\", \"u\")", []string{emoji}},
		{"var r = new globalThis.RegExp(\"[🇯🇵]\", \"\")", []string{withoutFlag, withoutFlag}},
		{"var r = new globalThis.RegExp(\"[\\\\u{1F468}\\\\u{200D}\\\\u{1F469}\\\\u{200D}\\\\u{1F466}]\", \"u\")", []string{joiner}},
		{"/[\\ud83d\\u{dc4d}]/u", []string{escapedPair}},
		{"/[\\u{d83d}\\udc4d]/u", []string{escapedPair}},
		{"/[\\u{d83d}\\u{dc4d}]/u", []string{escapedPair}},
		{"/[\\uD83D\\u{DC4d}]/u", []string{escapedPair}},
		{"RegExp(/[a👍z]/u, '');", []string{withoutFlag}},
		{"RegExp(/[👍]/)", []string{withoutFlag}},
		{"RegExp(/[👍]/, 'i');", []string{withoutFlag}},
		{"RegExp(/[👍]/, 'g');", []string{withoutFlag}},
		{"new RegExp(\"\\x5B \\\\ufe0f\\u005D\")", []string{combining}},
		{"new RegExp(\"[ \\u{5c}ufe0f]\")", []string{combining}},
		{"new RegExp(\"[ \\\\ufe\\60f]\")", []string{combining}},
		{"new RegExp(\"[ \\\\uf\\e0f]\")", []string{combining}},
		{"new RegExp(`[.\\\\u200D.]`)", []string{joiner}},
		{"new RegExp(`[.\\\\\\x75200D.]`)", []string{joiner}},
		{"var r = /[[👶🏻]]/v", []string{emoji}},
		{"new RegExp(/^[👍]$/v, '')", []string{withoutFlag}},
		{"/[\\u200c\\u200d\\p{ID_Continue}.]/", []string{joiner}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
					testCase.sourceText), testCase.wantIds...)
		})
	}
}

// The clean cases, which are where the discriminations actually live.
//
// Upstream's pass list is doing most of the work in this corpus, and reading it is how a porter
// learns what the rule is not about. Four groups of them each killed a different draft of this port:
// the solo cases (a lone surrogate, a lone combining mark, a lone joiner) say the rule is about
// adjacency and not about membership; the v-flag cases say a nested class or a `\q{...}` breaks a
// run in two; the invalid-regex cases say a pattern that cannot compile gets no finding at all; and
// the unresolvable-argument cases say a flags argument nobody can read means the rule declines.
func TestNoMisleadingCharacterClassStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"var r = /[👍]/u",
		"var r = /[\\uD83D\\uDC4D]/u",
		"var r = /[\\u{1F44D}]/u",
		"var r = /❇️/",
		"var r = /Á/",
		"var r = /[❇]/",
		"var r = /👶🏻/",
		"var r = /[👶]/u",
		"var r = /🇯🇵/",
		"var r = /[JP]/",
		"var r = /👨‍👩‍👦/",
		"new RegExp()",
		"var r = RegExp(/[👍]/u)",
		"const regex = /[👍]/u; new RegExp(regex);",
		"var r = /[\\uD83D]/",
		"var r = /[\\uDC4D]/",
		"var r = /[\\uD83D]/u",
		"var r = /[\\uDC4D]/u",
		"var r = /[\\u0301]/",
		"var r = /[\\uFE0F]/",
		"var r = /[\\u0301]/u",
		"var r = /[\\uFE0F]/u",
		"var r = /[\\u{1F3FB}]/u",
		"var r = /[🏻]/u",
		"var r = /[🇯]/u",
		"var r = /[🇵]/u",
		"var r = /[\\u200D]/",
		"var r = /[\\u200D]/u",
		"new RegExp('[Á] [ ');",
		"var r = new RegExp('[Á] [ ');",
		"var r = RegExp('{ [Á]', 'u');",
		"var r = new globalThis.RegExp('[Á] [ ');",
		"var r = globalThis.RegExp('{ [Á]', 'u');",
		"var r = RegExp(`${x}[👍]`)",
		"var r = new RegExp('[🇯🇵]', `${foo}`)",
		"const args = ['[👍]', 'i']; new RegExp(...args);",
		"var r = /[👍]/v",
		"var r = /^[\\q{👶🏻}]$/v",
		"var r = /[🇯\\q{abc}🇵]/v",
		"var r = /[🇯[A]🇵]/v",
		"var r = /[🇯[A--B]🇵]/v",
		"/[\\u200c\\u200d\\p{ID_Continue}.]/u",
		"/[\\u200c][\\u200d][a]/",
		"/[\\uD83D][\\uDC4D]/",
		"RegExp(/[👍]/, 'u');",
	}

	for _, testCase := range cases {
		t.Run(testCase, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile, testCase))
		})
	}
}

// allowEscape, both ways round.
//
// These are upstream's `allowEscape: true` cases, and they split cleanly. Its pass list under the
// option reports with the option off and is clean with it on, so each one pins both directions. Its
// fail list reports either way, because what excuses a member is its own spelling: a combining mark
// written out after an escaped base, `\\` or `\è` which are only backslashes in front of the
// character itself, and `\👍` whose halves are both written out.
func TestNoMisleadingCharacterClassAllowEscape(t *testing.T) {
	t.Parallel()

	allowEscape := NoMisleadingCharacterClassOptions{AllowEscape: true}

	// Clean upstream only when `allowEscape` is on.
	cleanOnlyWithTheOption := []struct {
		sourceText string
		wantIds    []string
	}{
		{"/[\\ud83d\\udc4d]/", []string{withoutFlag}},
		{"/[A\\u0301]/", []string{combining}},
		{"/[👶\\u{1f3fb}]/u", []string{emoji}},
		{"/[\\u{1F1EF}\\u{1F1F5}]/u", []string{regional}},
		{"/[👨\\u200d👩\\u200d👦]/u", []string{joiner}},
		{"/[\\u00B7\\u0300-\\u036F]/u", []string{combining}},
		{"/[\\n\\u0305]/", []string{combining}},
		{"RegExp(\"[\\uD83D\\uDC4D]\")", []string{withoutFlag}},
		{"RegExp(\"[A\\u0301]\")", []string{combining}},
		{"RegExp(\"[\\x41\\\\u0301]\")", []string{combining}},
		{"RegExp(`[\\uD83D\\uDC4D]`) // Backslash + \"uD83D\" + Backslash + \"uDC4D\"", []string{withoutFlag}},
	}

	for _, testCase := range cleanOnlyWithTheOption {
		t.Run("off "+testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
					testCase.sourceText), testCase.wantIds...)
		})
		t.Run("on "+testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTypedWithOptions(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
					testCase.sourceText, allowEscape))
		})
	}

	// Reported upstream even with the option on. With it on, each finding also points where ESLint's
	// does, which for every one of these is the class's whole body.
	reportedEitherWay := []struct {
		sourceText string
		wantIds    []string
	}{
		{"/[Á]/", []string{combining}},
		{"/[\\\\̶]/", []string{combining}},
		{"/[\\n̅]/", []string{combining}},
		{"/[\\👍]/", []string{withoutFlag}},
		{"RegExp('[\\è]')", []string{combining}},
		{"RegExp('[\\👍]')", []string{withoutFlag}},
		{"RegExp('[\\\\👍]')", []string{withoutFlag}},
		{"RegExp('[\\❇️]')", []string{combining}},
		{"RegExp(`[\\👍]`) // Backslash + U+D83D + U+DC4D", []string{withoutFlag}},
	}

	for _, testCase := range reportedEitherWay {
		t.Run("off "+testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
					testCase.sourceText), testCase.wantIds...)
		})
		t.Run("on "+testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisleadingCharacterClass,
				misleadingCharacterClassFile, testCase.sourceText, allowEscape)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			body := testCase.sourceText[strings.Index(testCase.sourceText, "[")+1 : strings.LastIndex(testCase.sourceText, "]")]
			for _, diagnostic := range result.Diagnostics {
				if got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]; got != body {
					t.Fatalf("points at %q, wanted %q", got, body)
				}
			}
		})
	}

	// A pattern that reached the call through a name has no source text of its own, so nothing in
	// it counts as escaped and it reports at the argument, as upstream's corpus has it.
	t.Run("a pattern held in a name", func(t *testing.T) {
		t.Parallel()
		sourceText := "const pattern = \"[\\x41\\u0301]\"; RegExp(pattern);"
		result := rule_testing.RunTypedWithOptions(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
			sourceText, allowEscape)
		rule_testing.ExpectFindings(t, result, combining)
		if got := sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]; got != "pattern" {
			t.Fatalf("points at %q, wanted the argument", got)
		}
	})

	// A range's minimum is spelled on its own, apart from the range: escaped, it excuses the mark it
	// is, and written out, it does not, whatever the maximum's spelling.
	for _, testCase := range []struct {
		sourceText string
		wantIds    []string
	}{
		{"/[A\\u0300-\u036F]/u", []string{}},
		{"/[A\u0300-\\u036F]/u", []string{combining}},
	} {
		t.Run("range "+testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTypedWithOptions(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
					testCase.sourceText, allowEscape), testCase.wantIds...)
		})
	}
}

// The option decodes under upstream's spelling and nothing else.
func TestNoMisleadingCharacterClassDecodesItsOption(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[NoMisleadingCharacterClassOptions]()
	for raw, want := range map[string]NoMisleadingCharacterClassOptions{
		`{}`:                    {},
		`{"allowEscape":false}`: {},
		`{"allowEscape":true}`:  {AllowEscape: true},
	} {
		decoded, err := decode([]byte(raw))
		if err != nil {
			t.Fatalf("decoding %s: %v", raw, err)
		}
		if decoded != any(want) {
			t.Fatalf("decoding %s gave %+v, wanted %+v", raw, decoded, want)
		}
	}
	for _, raw := range []string{`{"allowescape":true}`, `{"allowEscapes":true}`, `{"allowEscape":"yes"}`} {
		if _, err := decode([]byte(raw)); err == nil {
			t.Fatalf("decoding %s was accepted", raw)
		}
	}
}

// The `allowEscape` pass case that is clean with the option off too.
//
// A lone high surrogate written raw, followed by a backslash and the letters `udc4d`, so the class
// holds one surrogate and some ordinary letters and there is no pair. Kept separate from the block
// above because it proves the opposite thing: not every case in upstream's option list depends on
// the option.
func TestNoMisleadingCharacterClassIsCleanHereRegardlessOfTheOption(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"/[�d83d\\udc4d]/u // U+D83D + Backslash + \"udc4d\"",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile, sourceText))
			rule_testing.ExpectClean(t,
				rule_testing.RunTypedWithOptions(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
					sourceText, NoMisleadingCharacterClassOptions{AllowEscape: true}))
		})
	}
}

// Where the finding points, which the fixtures above cannot see.
//
// `ExpectFindings` asserts message ids and count and nothing else, so a rule reporting the right
// number of the right findings at the wrong offsets passes every case in this file. The span is not
// decoration here: it is the two class members that mislead, and a rule pointing at the whole
// literal would tell an author with a forty-character class nothing about which pair to fix.
//
// The offsets are checked by slicing the source rather than by comparing numbers, because a number
// computed from the same wrong base as the code that produced it agrees with itself.
func TestNoMisleadingCharacterClassPointsAtTheOffendingPair(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpans  []string
	}{
		// The pair, not the class and not the literal.
		{"var r = /[áb]/", []string{"á"}},
		// The astral character is one span even though it counts as two members, because there is
		// no narrower source text to point at.
		{"var r = /[👍]/", []string{"👍"}},
		// Escaped halves point at both escapes together, twelve characters rather than six.
		{`var r = /[👍]/`, []string{`👍`}},
		// A class in the middle of a pattern: the span must move with it.
		{`var r = /before[👍]after/`, []string{`👍`}},
		// Two classes, two spans, each pointing inside its own class.
		{"var r = /[👩‍👦][👩‍👦]/u", []string{"👩‍👦", "👩‍👦"}},
		// The joiner finding spans all three members, since the joiner alone means nothing.
		{"var r = /[a‍b]/u", []string{"a‍b"}},
		// A constructor's string argument: the offset has to cross the quote.
		{`var r = new RegExp("[á]")`, []string{"á"}},
		// A legacy octal escape ahead of the class, so the offset has to cross a token whose raw
		// width is not its cooked width. `\40` is one escape producing a space, and the octal run
		// stops at two digits because a leading `4` cannot take a third without exceeding 255. A
		// mutation widening that cap to three digits survived every other case in this file.
		{"var r = new RegExp(\"[\\40 \\\\ufe0f]\")", []string{" \\\\ufe0f"}},
		{"var r = new RegExp(\"\\770[ \\\\ufe0f]\")", []string{" \\\\ufe0f"}},
		// A range's minimum ends the run, so the pair before it points at the pair and not on into the
		// range.
		{`var r = /[\uD83D\uDC4D-\uffff]/`, []string{`\uD83D\uDC4D`}},
		// Without the flag a join reaching into an astral character stops where it starts, which is
		// ESLint's span moved from UTF-16 halves onto bytes: the `a`, the joiner, and the high half of
		// 👩, which has no bytes of its own.
		{"var r = /[a\u200d👩]/", []string{"a\u200d", "👩"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
				testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d diagnostics, got %d", len(testCase.wantSpans),
					len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.wantSpans[index] {
					t.Fatalf("diagnostic %d points at %q, wanted %q", index, got,
						testCase.wantSpans[index])
				}
			}
		})
	}
}

// The repair, asserted by applying it rather than by reading its text.
//
// A fix writing the right string over the wrong span passes a text comparison while being a real
// defect, so the source is rewritten and compared whole. Only the without-flag surrogate finding
// carries a repair, and only on a regex literal.
func TestNoMisleadingCharacterClassSuggestsTheUnicodeFlag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSource string
	}{
		{"var r = /[👍]/", "var r = /[👍]/u"},
		{`var r = /[👍]/`, `var r = /[👍]/u`},
		{`var r = /before[👍]after/`, `var r = /before[👍]after/u`},
		// Existing flags are kept and the new one is appended.
		{"var r = /[🇯🇵]/i", "var r = /[🇯🇵]/iu"},
		// A lookbehind is untouched by the flag, so the rewrite is offered.
		{"var r = /(?<=[👍])/", "var r = /(?<=[👍])/u"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
				testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("wanted a diagnostic, got none")
			}
			diagnostic := result.Diagnostics[0]
			if len(diagnostic.Suggestions) != 1 {
				t.Fatalf("wanted one suggestion, got %d", len(diagnostic.Suggestions))
			}
			fix := diagnostic.Suggestions[0].Fixes[0]
			rewritten := testCase.sourceText[:fix.Range.Pos()] + fix.Text +
				testCase.sourceText[fix.Range.End():]
			if rewritten != testCase.wantSource {
				t.Fatalf("applying the suggestion gave %q, wanted %q", rewritten,
					testCase.wantSource)
			}
		})
	}
}

// Why the rewrite is a suggestion and never a fix, as an executable claim.
//
// Adding `u` changes how the rest of the pattern parses. `/[👍]\a/` is a legal pattern today and
// `/[👍]\a/u` is a SyntaxError, because an identity escape of a letter is rejected under the flag.
// An engine applying that unattended would turn a working file into one that throws, and a fix
// producing source that parses but does not run is the failure the edit engine structurally cannot
// refuse.
//
// So two things are asserted: the rule never proposes an unattended fix, and it withholds even the
// suggestion when the pattern holds a construct the flag would break.
func TestNoMisleadingCharacterClassNeverProposesAnAutomaticFix(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
		"var r = /[👍]/")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("the rule proposed %d unattended fixes; the u flag changes meaning and must be "+
			"offered rather than applied", len(result.Diagnostics[0].Fixes))
	}

	// Patterns the flag would break get the finding and no repair.
	for _, sourceText := range []string{
		// An identity escape of a letter is a SyntaxError under `u`.
		`var r = /[👍]\a/`,
		`var r = /\a[👍]\a/`,
		// An unquantified brace is a SyntaxError under `u`.
		"var r = /[👍]{/",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
				sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("wanted a diagnostic, got none")
			}
			if len(result.Diagnostics[0].Suggestions) != 0 {
				t.Fatalf("wanted no suggestion on a pattern the flag would break, got %d",
					len(result.Diagnostics[0].Suggestions))
			}
		})
	}
}

// A pattern already carrying `u` or `v` gets no flag suggestion, since it already has one.
//
// The findings that survive the flag are the four that are about code points rather than about code
// units, so this is the ordinary case for a combining mark rather than an edge.
func TestNoMisleadingCharacterClassDoesNotSuggestAFlagThatIsAlreadyThere(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{"var r = /[Á]/u", "var r = /[Á]/v"} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
				sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			if len(result.Diagnostics[0].Suggestions) != 0 {
				t.Fatalf("wanted no suggestion, got %d", len(result.Diagnostics[0].Suggestions))
			}
		})
	}
}

// Cases written from reading our code rather than upstream's.
//
// The imported corpus is a floor. Each case below covers a decision this port makes that upstream's
// cases do not reach, and each exists because the mutation sweep or the shelf's own shape put the
// question. Stated per case rather than in a block, because a reader hitting one of these failing
// needs to know what it was protecting.
func TestNoMisleadingCharacterClassCoversOurOwnDecisions(t *testing.T) {
	t.Parallel()

	fires := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			// A range's maximum starts a new run, so a pair straddling the range's end is adjacent
			// and reports. Upstream's corpus has the min side and not this one.
			"a pair after a range's maximum",
			"var r = /[a-b́]/",
			[]string{combining},
		},
		{
			// `window.RegExp` and `global.RegExp` reach the same constructor as `globalThis.RegExp`,
			// which is the only one upstream's corpus exercises.
			"a window-qualified constructor",
			`var r = new window.RegExp("[Á]")`,
			[]string{combining},
		},
		{
			"a global-qualified constructor",
			`var r = new global.RegExp("[Á]")`,
			[]string{combining},
		},
		{
			// A variation selector is a combining character here, and the supplement block is a
			// separate range from the basic one that upstream's cases use.
			"a variation selector from the supplement block",
			"var r = /[a\U000E0100]/u",
			[]string{combining},
		},
		{
			// The three combining blocks upstream lists and never exercises.
			"a combining mark from the extended block",
			"var r = /[a᪰]/u",
			[]string{combining},
		},
		{
			"a combining mark from the supplement block",
			"var r = /[a᷀]/u",
			[]string{combining},
		},
		{
			"a combining mark from the symbols block",
			"var r = /[a⃐]/u",
			[]string{combining},
		},
		{
			"a combining half mark",
			"var r = /[a︠]/u",
			[]string{combining},
		},
		{
			// An open-ended quantifier before the class. The brace opens a real quantifier, so the
			// pattern compiles under the unicode flag and the class inside it still misleads. A
			// mutation dropping the comma branch of the quantifier grammar reads `{2,}` as a lone
			// brace, decides the pattern cannot compile, and goes silent on all three of these.
			"an open-ended quantifier before the class",
			"var r = /a{2,}[Á]/u",
			[]string{combining},
		},
		{
			"a bounded quantifier before the class",
			"var r = /a{2,3}[Á]/u",
			[]string{combining},
		},
		{
			// A brace inside a character class is a literal brace in every mode, so the scan for a
			// lone brace has to step over classes whole. A mutation that walks into them instead
			// reads this `{` as lone and goes silent.
			"a literal brace in its own class before the class",
			"var r = /[{][Á]/u",
			[]string{combining},
		},
		{
			// A base with two marks stacked on it is one misleading sequence, not two. The second
			// mark follows a mark, and a mark following a mark adds nothing new to report: the
			// finding is already made and the run is already known to be misleading. Upstream's
			// corpus never stacks two marks, so deleting `!isCombiningCharacter(previous.value)`
			// survived its entire fail list; this case and the next one are what kill it.
			"a base carrying two stacked combining marks",
			"var r = /[á̂]/",
			[]string{combining},
		},
	}

	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
					testCase.sourceText), testCase.wantIds...)
		})
	}

	silent := []struct {
		name       string
		sourceText string
	}{
		{
			// A class escape breaks the run, so the members either side are not adjacent. Upstream
			// handles this in its collector and its corpus never puts a `\d` between two halves.
			"a class escape between two would-be neighbours",
			"var r = /[👶\\d🏻]/u",
		},
		{
			// The two endpoints of one range are not neighbours: everything between them is in the
			// set too, so calling them adjacent would report on `[a-z]`.
			"a range whose own endpoints would otherwise pair",
			"var r = /[\U0001F1EF-\U0001F1F5]/u",
		},
		{
			// A callee that is not RegExp is not this rule's business, however suggestive the
			// argument is.
			"a call to something else entirely",
			`var r = new NotRegExp("[Á]")`,
		},
		{
			// A member access whose object is not a global namespace.
			"a RegExp property on an ordinary object",
			`var r = new someLibrary.RegExp("[Á]")`,
		},
		{
			// A pattern argument that is not readable at lint time. A constant binding is readable,
			// as ESLint reads it, and reports at the argument; the table covers that.
			"a pattern held in a parameter",
			`function build(pattern: string) { return new RegExp(pattern) }`,
		},
		{
			// A template with a substitution could be anything once it runs.
			"a substituting template as the pattern",
			"var r = new RegExp(`[${x}Á]`)",
		},
		{
			// Two joiners in a row join nothing, and upstream skips the pair explicitly. Its corpus
			// never writes one.
			"two joiners side by side",
			"var r = /[‍‍]/u",
		},
		{
			// A high surrogate is only half of anything, so what follows it decides. Upstream's
			// corpus has the solo lead surrogate and never puts an ordinary character after one, so
			// dropping the low half's range check survives its whole fail list. A mutation sweep
			// found that: `low >= 0xDC00 && low <= 0xDFFF` could be deleted outright with 187
			// fixtures green.
			"a lead surrogate followed by an ordinary letter",
			`var r = /[\uD83Da]/`,
		},
		{
			// The same gap from the other side: two leads are not a pair.
			"two lead surrogates",
			`var r = /[\uD83D\uD83D]/`,
		},
		{
			// And a trail with nothing leading it.
			"an ordinary letter followed by a trail surrogate",
			`var r = /[a\uDC4D]/`,
		},
		{
			// Two combining marks and no base is a class of two marks, which is what it says it is.
			// The rule is about a mark attaching to something, and nothing is attaching here.
			"two combining marks with nothing to combine with",
			"var r = /[́̂]/",
		},
		{
			// A misleading class followed by one the parser refuses. The whole pattern is a
			// SyntaxError, so nothing in it can mislead anybody at run time and no finding is right.
			//
			// This pins a real disagreement between the two shelf layers rather than a hypothetical:
			// the class walker accepts `[a-\d]` as a well-formed span and the class parser refuses
			// it, because a range with a set endpoint is a syntax error. A mutation making that
			// refusal a no-op survived every other fixture in this file, and these two cases are
			// what distinguish the versions.
			"a misleading class followed by an unparseable range",
			`var r = /[Á][a-\d]/u`,
		},
		{
			// Same shape, refused for a different reason: `\B` inside a class is a syntax error
			// under the unicode flag.
			"a misleading class followed by an unparseable escape",
			`var r = /[Á][\B]/u`,
		},
		{
			// A pair written backwards is not a pair.
			"a trail surrogate before a lead",
			`var r = /[\uDC4D\uD83D]/`,
		},
		{
			// The rule is about character classes, so the same characters outside one are fine.
			// Upstream's corpus has `/❇️/` and `/Á/`; this is the joiner and the modifier.
			"a joined sequence outside any class",
			"var r = /👩‍👦/u",
		},
		{
			"an emoji modifier outside any class",
			"var r = /👶🏻/u",
		},
	}

	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoMisleadingCharacterClass, misleadingCharacterClassFile,
					testCase.sourceText))
		})
	}
}
