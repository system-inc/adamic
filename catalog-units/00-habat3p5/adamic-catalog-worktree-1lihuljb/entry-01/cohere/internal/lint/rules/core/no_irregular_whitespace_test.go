package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// irregularWhitespaceFile is where the fixtures pretend to live. A .tsx name because the corpus
// carries JSX cases, and a .ts fixture parses `<div>` as a type assertion rather than an element,
// which would make a JSX case pass by finding nothing at all.
const irregularWhitespaceFile = "/repository/source/IrregularWhitespace.tsx"

// boolOf exists because the options are pointers and Go has no address-of for a literal.
func boolOf(value bool) *bool { return &value }

// The corpus is oxc's, copied rather than rewritten, and it needed reconciling before it could be
// copied at all.
//
// Three things about it are not visible from the rule file. There are two Tester blocks, and the
// second calls `.test()` rather than `.test_and_snapshot()`, so its two cases produce no snapshot
// entry; this tree's extractor additionally reports that block as empty, because its vectors are
// passed inline rather than through a `let` and the extractor keys on the `let`. Both of its cases
// are deliberate regressions and both are reproduced below. The snapshot holds 83 diagnostics across
// 58 failing inputs, so a fixture asserting one finding per input would be wrong on eighteen of
// them. And the per-input counts had to be recovered by walking each snapshot frame's printed source
// lines against the case they belong to, because the obvious in-order walk misattributes: two cases
// open with a byte-identical first line, and a naive alignment hands one of them both diagnostics
// and the other none.
//
// Every case states oxc's defaults explicitly rather than inheriting ours. oxc and ESLint disagree
// on three of the five defaults, and this rule follows ESLint because that is what this tree's
// resolved config asks for on every file. Passing oxc's values explicitly is what lets upstream's
// counts reproduce exactly instead of being quietly reinterpreted under different defaults.
//
// Every irregular character below is written as a codepoint escape. The characters are invisible, so
// a literal would be unreviewable in a diff, would not survive a copy and paste, and would trip this
// very rule on the file that tests it.

func TestNoIrregularWhitespaceFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    NoIrregularWhitespaceOptions
		findings   int
	}{
		{"var any \x0b = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \x0c = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u00a0 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \ufeff = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2000 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2001 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2002 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2003 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2004 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2005 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2006 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2007 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2008 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2009 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u200a = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2028 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u2029 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u202f = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u205f = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any \u3000 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var a = 'b',\u2028c = 'd',\n            e = 'f'\u2028", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"var any \u3000 = 'thing', other \u3000 = 'thing';\n            var third \u3000 = 'thing';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 3},
		{"var any = '\u3000', other = '\x0b';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(false), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"var any = `\u3000`, other = `\x0b`;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(false), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"`something ${\u3000 10} another thing`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"`something ${10\u3000} another thing`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"\u3000\n            `\u3000template`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"\u3000\n            `\u3000multiline\n            template`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"\u3000`\u3000template`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"\u3000`\u3000multiline\n            template`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"`\u3000template`\u3000", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"`\u3000multiline\n            template`\u3000", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"`\u3000template`\n            \u3000", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"`\u3000multiline\n            template`\n            \u3000", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var foo = \x0b bar;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var foo =\x0bbar;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		// A run of irregular spaces on one line is one finding, as ESLint counts these four rows (#jjfa7qb)
		{"var foo = \x0b\x0b bar;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var foo = \x0b\x0c bar;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var foo = \x0b \x0b bar;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"var foo = \x0bbar\x0b;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"\x0b", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"\u00a0\u2002\u2003", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var foo = \x0b\n            bar;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var foo =\x0b\n            \x0bbar;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"var foo = \x0c\x0b\n            \x0c\x0b\x0cbar\n            ;\x0b\x0c\n            ", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 3},
		{"var foo = \u2028bar;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var foo =\u2029 bar;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var foo = bar;\u2028", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"\u2029", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"foo\u2028\u2028", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"foo\u2029\u2028", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"foo\u2028\n            \u2028", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"foo\x0b\u2028\x0b", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 3},
		{"// \u3000", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"/* \u3000 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 1},
		{"var any = /\u3000/, other = /\u200b/;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(false), SkipJSXText: boolOf(true)}, 2},
		{"var any = `\u3000`, other = `\u200b`;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(false), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}, 2},
		{"<div>\u3000</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(false)}, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoIrregularWhitespace,
				irregularWhitespaceFile, testCase.sourceText, testCase.options)

			wanted := make([]string, testCase.findings)
			for index := range wanted {
				wanted[index] = "noIrregularWhitespace"
			}
			rule_testing.ExpectFindings(t, result, wanted...)
		})
	}
}

func TestNoIrregularWhitespaceStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    NoIrregularWhitespaceOptions
	}{
		{"'\\u000B';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u000C';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u0085';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u00A0';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u180E';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\ufeff';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2000';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2001';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2002';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2003';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2004';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2005';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2006';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2007';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2008';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2009';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u200A';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u200B';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2028';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u2029';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u202F';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u205f';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\u3000';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\x0b';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\x0c';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u0085';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u00a0';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u180e';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\ufeff';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2000';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2001';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2002';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2003';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2004';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2005';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2006';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2007';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2008';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u2009';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u200a';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u200b';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\\u2028';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\\\u2029';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u202f';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u205f';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"'\u3000';", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \x0b", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \x0c", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u0085", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u00a0", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u180e", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \ufeff", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2000", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2001", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2002", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2003", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2004", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2005", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2006", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2007", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2008", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u2009", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u200a", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u200b", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u202f", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u205f", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"// \u3000", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \x0b */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \x0c */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u0085 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u00a0 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u180e */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \ufeff */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2000 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2001 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2002 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2003 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2004 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2005 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2006 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2007 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2008 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2009 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u200a */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u200b */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2028 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u2029 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u202f */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u205f */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/* \u3000 */", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"#!/usr/bin/env\u200bnode", NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\x0b/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\x0c/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u0085/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u00a0/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u180e/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\ufeff/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2000/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2001/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2002/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2003/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2004/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2005/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2006/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2007/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2008/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u2009/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u200a/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u200b/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u202f/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u205f/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"/\u3000/", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\x0b`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\x0c`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u0085`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u00a0`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u180e`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\ufeff`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2000`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2001`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2002`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2003`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2004`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2005`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2006`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2007`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2008`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u2009`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u200a`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u200b`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u202f`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u205f`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u3000`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"`\u3000${foo}\u3000`", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"const error = ` \u3000 `;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"const error = `\n            \u3000`;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"const error = `\u3000\n            `;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"const error = `\n            \u3000\n            `;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"const error = `foo\u3000bar\n            foo\u3000bar`;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\x0b</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\x0c</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u0085</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u00a0</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u180e</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\ufeff</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2000</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2001</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2002</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2003</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2004</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2005</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2006</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2007</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2008</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u2009</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u200a</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u200b</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u202f</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u205f</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"<div>\u3000</div>;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"\ufeffconsole.log('hello BOM');", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"var any = /\u3000/, other = /\x0b/;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
		{"var any = `\u3000`, other = `\x0b`;", NoIrregularWhitespaceOptions{SkipComments: boolOf(false), SkipStrings: boolOf(true), SkipTemplates: boolOf(true), SkipRegExps: boolOf(true), SkipJSXText: boolOf(true)}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoIrregularWhitespace,
				irregularWhitespaceFile, testCase.sourceText, testCase.options))
		})
	}
}

// The second Tester block, which does not snapshot and which the extractor reports as empty.
//
// Both cases cover the two paths that carry no node: a zero width space inside a line comment, and
// one inside a hashbang. Their diagnostic counts are stated nowhere upstream, so these two are read
// from the code rather than from a snapshot, which is the reason the block exists at all.
func TestNoIrregularWhitespaceCoversTheUnsnapshottedBlock(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"a zero width space in a line comment",
			"export const ZERO_SPACE = String.fromCharCode(0x200b) // '\u200b'",
		},
		{"a zero width space in a hashbang", "#!/usr/bin/env\u200bnode"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoIrregularWhitespace,
				irregularWhitespaceFile, testCase.sourceText,
				NoIrregularWhitespaceOptions{SkipComments: boolOf(false)}),
				"noIrregularWhitespace")
		})
	}
}

// Where each finding points, which the message-id fixtures above cannot see.
//
// This rule carries no fix, and that is exactly why the spans need asserting rather than why they do
// not. A rule whose only defect is where it points passes a complete pair of message-id fixtures
// while being wrong at every single finding, and nothing else in this file would notice: the count
// would still be 83. So each finding is sliced back out of the source with its own range and
// compared against the character it claims to report.
//
// The multi-byte cases carry the weight. An ideographic space is three bytes in UTF-8, so a span
// built by counting characters rather than bytes points at a prefix of it, and a span of fixed width
// one is wrong the same way. Byte offsets are what a range holds here.
func TestNoIrregularWhitespaceReportsTheCharacterItself(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpans  []string
	}{
		{
			"a one-byte vertical tab",
			"var any \u000b = 'thing';",
			[]string{"\u000b"},
		},
		{
			"a two-byte no-break space",
			"var any \u00a0 = 'thing';",
			[]string{"\u00a0"},
		},
		{
			"a three-byte ideographic space",
			"var any \u3000 = 'thing';",
			[]string{"\u3000"},
		},
		{
			"one finding per run, as ESLint reports it",
			"var foo = \u000b\u000b bar;",
			[]string{"\u000b\u000b"},
		},
		{
			"a run of three different multi-byte characters is one span over all three",
			"\u00a0\u2002\u2003",
			[]string{"\u00a0\u2002\u2003"},
		},
		{
			"each line separator is its own finding, never part of a run",
			"foo\u000b\u2028\u2028\u000b",
			[]string{"\u000b", "\u2028", "\u2028", "\u000b"},
		},
		{
			"a character after a multi-byte one, so the offset has already advanced",
			"const \u3000x = \u200b1;",
			[]string{"\u3000", "\u200b"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoIrregularWhitespace,
				irregularWhitespaceFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d",
					len(testCase.wantSpans), len(result.Diagnostics))
			}
			for index, wantSpan := range testCase.wantSpans {
				diagnostic := result.Diagnostics[index]
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != wantSpan {
					t.Fatalf("finding %d reported %q, wanted %q", index, reported, wantSpan)
				}
			}
		})
	}
}

// Cases written from reading our code rather than upstream's, covering discriminations the imported
// corpus leaves untested.
//
// Three of the twenty-four characters have no failing case anywhere upstream. U+1680 appears nowhere
// in the rule file at all, in either direction, and U+0085 and U+180E appear only as passing cases
// inside strings, which pass trivially for a rule that reports nothing. So a port could omit all
// three from its set and still pass every imported case. U+180E is the sharpest: it is the character
// oxc adds by hand on top of its parser predicate, and the one character the near-miss helper in
// this tree omits, so taking the set from the wrong source drops it silently.
func TestNoIrregularWhitespaceCoversCharactersUpstreamNeverFails(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an ogham space mark, which upstream tests in neither direction", "var any \u1680 = 1;"},
		{"a next line character, which upstream only tests inside a string", "var any \u0085 = 1;"},
		{
			"a mongolian vowel separator, the character a scanner predicate omits",
			"var any \u180e = 1;",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoIrregularWhitespace, irregularWhitespaceFile, testCase.sourceText),
				"noIrregularWhitespace")
		})
	}
}

// A regular space and a tab are the two characters this rule exists to permit.
//
// Trivial to state and load-bearing to assert. The nearest helper in this tree treats both as
// whitespace, along with newline and carriage return, so a rule built on it reports every line of
// every file. Nothing in the imported corpus fails if all four leak into the set, because upstream
// never wrote a case for code containing an ordinary space; every clean case it does have would go
// red, but only once somebody ran them.
func TestNoIrregularWhitespaceAllowsOrdinaryWhitespace(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoIrregularWhitespace, irregularWhitespaceFile,
		"const first = 1;\n\tconst second = 2;\r\nconst third = 3;\n"))
}

// The byte order mark carve-out, in both directions.
//
// Upstream pins only the passing half, with a leading mark on a file that reports nothing. The
// failing half is the half a port gets wrong: a carve-out written as "skip U+FEFF" rather than "skip
// U+FEFF at offset zero" passes the imported case and then silently stops reporting the character
// everywhere else in the file.
func TestNoIrregularWhitespaceExcusesOnlyTheLeadingByteOrderMark(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoIrregularWhitespace, irregularWhitespaceFile,
		"\ufeffconsole.log('hello BOM');"))

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoIrregularWhitespace, irregularWhitespaceFile,
		"console.log('hello');\ufeff"), "noIrregularWhitespace")
}

// A skipped literal excuses its own text and not the trivia in front of it.
//
// A node's Pos sits before its leading trivia, so excusing the span from Pos to End covers whatever
// whitespace and comments precede the literal. Under the default options that is directly wrong: the
// ideographic space here stands in code position, before the string begins, and skipStrings must not
// reach backwards over it. Upstream has no case for this because oxc never computes a skip region at
// all, so nothing on that side could have caught it.
func TestNoIrregularWhitespaceDoesNotExcuseTriviaBeforeASkippedLiteral(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoIrregularWhitespace, irregularWhitespaceFile,
		"const value =\u3000'text';"), "noIrregularWhitespace")
}

// The interpolation inside a skipped template stays reportable.
//
// Upstream pins this with two failing cases under defaults that turn skipTemplates on, and it earns
// a case of our own because the two implementations reach it differently. oxc excuses the quasis one
// at a time and never sees the template expression as a unit, while this walks the tree and could
// excuse the whole span in a single line. That version passes every other template case here.
func TestNoIrregularWhitespaceReportsInsideASkippedTemplatesInterpolation(t *testing.T) {
	t.Parallel()

	skipTemplates := NoIrregularWhitespaceOptions{SkipTemplates: boolOf(true)}

	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoIrregularWhitespace,
		irregularWhitespaceFile, "`before ${\u3000 value} after`", skipTemplates),
		"noIrregularWhitespace")

	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoIrregularWhitespace,
		irregularWhitespaceFile, "`before\u3000 ${value} after`", skipTemplates))
}

// The defaults this rule applies when the config says nothing, one option at a time.
//
// Every imported case states all five explicitly, so none of them can see what happens with no
// options at all. Four of the five defaults are Go's zero value and would survive the options block
// being deleted entirely; skipStrings is the one that would not, and it is the one where "the config
// said false" and "the config said nothing" mean opposite things.
func TestNoIrregularWhitespaceDefaults(t *testing.T) {
	t.Parallel()

	// skipStrings defaults to on, so a string stays quiet.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoIrregularWhitespace, irregularWhitespaceFile,
		"var any = '\u3000';"))

	// The other four default to off, so each of these reports.
	for _, sourceText := range []string{
		"// a comment holding \u3000",
		"var any = `\u3000`;",
		"var any = /\u3000/;",
		"var any = <div>\u3000</div>;",
	} {
		rule_testing.ExpectFindings(t, rule_testing.Run(t, NoIrregularWhitespace,
			irregularWhitespaceFile, sourceText), "noIrregularWhitespace")
	}
}

// A comment opener inside a string is not a comment, and a string is not excused by skipComments.
//
// This pins the reason the rule reads parser trivia rather than matching on a slash pair. The
// measured behavior of the shelf on this exact input is that it returns no comment at all, so the
// ideographic space here is inside a string and nothing else. Under skipComments with skipStrings
// off it must still report, which a rule finding its comments by scanning text would get wrong.
func TestNoIrregularWhitespaceDoesNotTreatACommentOpenerInAStringAsAComment(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoIrregularWhitespace,
		irregularWhitespaceFile, "var s = '// not a comment \u3000';",
		NoIrregularWhitespaceOptions{SkipComments: boolOf(true), SkipStrings: boolOf(false)}),
		"noIrregularWhitespace")
}

// A multi-byte character standing immediately before an irregular one, with no byte between them.
//
// This exists for the guard rather than for the rule. The guard walks bytes and steps over a decoded
// character's width, and that arithmetic has to be exact: advancing one byte too far lands inside
// the following character, decodes the replacement rune instead of it, and the guard concludes the
// file holds nothing, which silently suppresses every finding in that file.
//
// Adjacency is the whole point and it is easy to lose. An earlier version of this case put a space
// between the two characters, and a mis-stepping guard landed back on that ASCII byte and recovered,
// so the case passed against the broken version and proved nothing. The two must share a boundary:
// an e-acute is two bytes and an ideographic space is three, so the irregular character begins
// exactly at the byte an off-by-one guard skips. Every other case in this file puts ASCII in front
// of its irregular character, where a mis-stepped index happens to land correctly.
func TestNoIrregularWhitespaceFindsAnIrregularCharacterAbuttingAMultiByteOne(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoIrregularWhitespace, irregularWhitespaceFile,
		"const \u00e9\u3000x = 1;"), "noIrregularWhitespace")

	// A three-byte character abutting a three-byte one, so a guard mis-stepping by a different
	// amount is caught too.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoIrregularWhitespace, irregularWhitespaceFile,
		"const \u4e2d\u3000x = 1;"), "noIrregularWhitespace")
}
