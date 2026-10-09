package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// uselessEscapeFile is where the fixtures pretend to live.
//
// A `.tsx` extension rather than `.ts`, because six of upstream's pass cases and one of its fail
// cases are JSX, and the whole point of those seven is the distinction between a string literal
// that *is* a JSX attribute value and one sitting inside an expression container. Parsed as plain
// TypeScript the JSX never becomes JSX nodes at all, and all seven would assert nothing.
const uselessEscapeFile = "/repository/source/UselessEscape.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_useless_escape.rs`:
// one Tester block holding 204 pass and 89 fail cases, with 53 of them carrying an
// `allowRegexCharacters` option. The snapshot records 105 diagnostics from those 89 inputs, so the
// per-input count is measured from the snapshot rather than assumed to be one: nine inputs report
// more than once and one reports six times.
//
// The counts were recovered by aligning each snapshot diagnostic against the source line it prints,
// which is the only field that survives upstream's own duplicate entries. Upstream lists two of its
// multiline template cases twice, byte-identical, and a count derived by walking the snapshot in
// order assigns both diagnostics to the first copy.

// TestNoUselessEscapeFires runs every failing input and asserts the exact diagnostic count.
func TestNoUselessEscapeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		findings   int
		options    any
	}{
		{"var foo = /\\#/;", 1, nil},
		{"var foo = /\\;/;", 1, nil},
		{"var foo = \"\\'\";", 1, nil},
		{"var foo = \"\\#/\";", 1, nil},
		{"var foo = \"\\a\"", 1, nil},
		{"var foo = \"\\B\";", 1, nil},
		{"var foo = \"\\@\";", 1, nil},
		{"var foo = \"foo \\a bar\";", 1, nil},
		{"var foo = '\\\"';", 1, nil},
		{"var foo = '\\#';", 1, nil},
		{"var foo = '\\$';", 1, nil},
		{"var foo = '\\p';", 1, nil},
		{"var foo = '\\p\\a\\@';", 3, nil},
		{"<foo attr={\"\\d\"}/>", 1, nil},
		{"var foo = '\\`';", 1, nil},
		{"var foo = `\\\"`;", 1, nil},
		{"var foo = `\\'`;", 1, nil},
		{"var foo = `\\#`;", 1, nil},
		{"var foo = '\\`foo\\`';", 2, nil},
		{"var foo = `\\\"${foo}\\\"`;", 2, nil},
		{"var foo = `\\'${foo}\\'`;", 2, nil},
		{"var foo = `\\#${foo}`;", 1, nil},
		{"let foo = '\\ ';", 1, nil},
		{"let foo = /\\ /;", 1, nil},
		{"var foo = `\\$\\{{${foo}`;", 1, nil},
		{"var foo = `\\$a${foo}`;", 1, nil},
		{"var foo = `a\\{{${foo}`;", 1, nil},

		// Not upstream's. A brace escaped as the first content character of a chunk is the one input
		// that reads the seeded previous character, and upstream's corpus never writes it: every case it
		// has puts a real character or a dollar before the brace. Mutating the seed from a backtick to a
		// dollar therefore survived a green suite, silencing exactly this input while the other 292 stayed
		// clean. Modelling upstream's own scanner on it confirms it reports, so this is the case, not a
		// divergence.
		{"var foo = `\\{`;", 1, nil},
		// The same seam read from the other end: a chunk after a substitution also begins with no
		// real previous character, so a brace escaped right after the closing brace of `${foo}` is
		// judged by the seed too.
		{"var foo = `${foo}\\{`;", 1, nil},
		// Also not upstream's, and the same seam from the regex side. Outside a character class the
		// letter q is not a class escape, so this reports the `\\q` itself. It reports that and nothing
		// else, which is the half worth pinning: the shelf's scanner consumes a braced `\\q` as one
		// token wherever it appears, so the backslash inside the braces is never reached. Hardcoding
		// the escape width to two bytes made this read two findings instead of one.
		//
		// The count was written as two here from reasoning about where `\\q` is special before it was
		// measured, and the fixture caught that rather than the rule.
		{"var foo = /\\q{a\\#b}/v;", 1, nil},
		{"var foo = /[ab\\-]/", 1, nil},
		{"var foo = /[\\-ab]/", 1, nil},
		{"var foo = /[ab\\?]/", 1, nil},
		{"var foo = /[ab\\.]/", 1, nil},
		{"var foo = /[a\\|b]/", 1, nil},
		{"var foo = /\\-/", 1, nil},
		{"var foo = /[\\-]/", 1, nil},
		{"var foo = /[ab\\$]/", 1, nil},
		{"var foo = /[\\(paren]/", 1, nil},
		{"var foo = /[\\[]/", 1, nil},
		{"var foo = /[\\/]/", 1, nil},
		{"var foo = /[\\B]/", 1, nil},
		{"var foo = /[a][\\-b]/", 1, nil},
		{"var foo = /\\-[]/", 1, nil},
		{"var foo = /[a\\^]/", 1, nil},
		{"`multiline template\n\t\t\tliteral with useless \\escape`", 1, nil},
		{"`multiline template\n\t\t\tliteral with useless \\escape`", 1, nil},
		{"`template literal with line continuation \\\n\t\t\tand useless \\escape`", 1, nil},
		{"`template literal with line continuation \\\n\t\t\tand useless \\escape`", 1, nil},
		{"`template literal with mixed linebreaks \r\r\n\n\\and useless escape`", 1, nil},
		{"`template literal with mixed linebreaks in line continuations \\\n\\\r\\\r\n\\and useless escape`", 1, nil},
		{"`\\a```", 1, nil},
		{"\"use\\ strict\";", 1, nil},
		{"var foo = /\\（([^\\）\\（]+)\\）$|\\(([^\\)\\)]+)\\)$/;", 6, nil},
		{"var stringLiteralWithNextLine = \"line 1\\line 2\";", 1, nil},
		{"var stringLiteralWithNextLine = `line 1\\line 2`;", 1, nil},
		{"({ foo() { \"foo\"; \"bar\"; \"ba\\z\" } })", 1, nil},
		{"/[^\\^]/", 1, nil},
		{"/[^\\^]/u", 1, nil},
		{"/[\\$]/v", 1, nil},
		{"/[\\&\\&]/v", 1, nil},
		{"/[\\!\\!]/v", 1, nil},
		{"/[\\#\\#]/v", 1, nil},
		{"/[\\%\\%]/v", 1, nil},
		{"/[\\*\\*]/v", 1, nil},
		{"/[\\+\\+]/v", 1, nil},
		{"/[\\,\\,]/v", 1, nil},
		{"/[\\.\\.]/v", 1, nil},
		{"/[\\:\\:]/v", 1, nil},
		{"/[\\;\\;]/v", 1, nil},
		{"/[\\<\\<]/v", 1, nil},
		{"/[\\=\\=]/v", 1, nil},
		{"/[\\>\\>]/v", 1, nil},
		{"/[\\?\\?]/v", 1, nil},
		{"/[\\@\\@]/v", 1, nil},
		{"/[\\`\\`]/v", 1, nil},
		{"/[\\~\\~]/v", 1, nil},
		{"/[^\\^\\^]/v", 1, nil},
		{"/[_\\^\\^]/v", 1, nil},
		{"/[\\&\\&&\\&]/v", 1, nil},
		{"/[\\p{ASCII}--\\.]/v", 1, nil},
		{"/[\\p{ASCII}&&\\.]/v", 1, nil},
		{"/[\\.--[.&]]/v", 1, nil},
		{"/[\\.&&[.&]]/v", 1, nil},
		{"/[\\.--\\.--\\.]/v", 3, nil},
		{"/[\\.&&\\.&&\\.]/v", 3, nil},
		{"/[[\\.&]--[\\.&]]/v", 2, nil},
		{"/[[\\.&]&&[\\.&]]/v", 2, nil},
		{"var foo = \"\\#/\";", 1, NoUselessEscapeOptions{AllowRegexCharacters: []string{"#"}}},
		{"var foo = /\\#\\@/;", 1, NoUselessEscapeOptions{AllowRegexCharacters: []string{"#"}}},
		{"var foo = /[a\\@b]/", 1, NoUselessEscapeOptions{AllowRegexCharacters: []string{"#"}}},
		{"/[\\@\\@]/v", 1, NoUselessEscapeOptions{AllowRegexCharacters: []string{"#"}}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = "noUselessEscape"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.RunWithOptions(t, NoUselessEscape, uselessEscapeFile,
					testCase.sourceText, testCase.options), wantIds...)
		})
	}
}

// TestNoUselessEscapeStaysSilent runs every clean input.
//
// These are the half that catch a port. The escape tables are the whole rule and each entry was
// added upstream because somebody hit it: `/\w\$\*\./` is four escapes outside a class that are
// all required, `/[a\-b]/` is a hyphen escaped in the middle of a class where escaping it is the
// only way to mean a literal, and the fifty v-flag cases pin the reserved double punctuator rule
// that no amount of reading the prose would produce correctly.
func TestNoUselessEscapeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
	}{
		{"var foo = /\\./", nil},
		{"var foo = /\\//g", nil},
		{"var foo = /\"\"/", nil},
		{"var foo = /''/", nil},
		{"var foo = /([A-Z])\\t+/g", nil},
		{"var foo = /([A-Z])\\n+/g", nil},
		{"var foo = /([A-Z])\\v+/g", nil},
		{"var foo = /\\D/", nil},
		{"var foo = /\\W/", nil},
		{"var foo = /\\w/", nil},
		{"var foo = /\\\\/g", nil},
		{"var foo = /\\w\\$\\*\\./", nil},
		{"var foo = /\\^\\+\\./", nil},
		{"var foo = /\\|\\}\\{\\./", nil},
		{"var foo = /]\\[\\(\\)\\//", nil},
		{"var foo = \"\\x123\"", nil},
		{"var foo = \"\\u00a9\"", nil},
		{"var foo = \"\\377\"", nil},
		{"var foo = \"\\\"\"", nil},
		{"var foo = \"xs\\u2111\"", nil},
		{"var foo = \"foo \\\\ bar\";", nil},
		{"var foo = \"\\t\";", nil},
		{"var foo = \"foo \\b bar\";", nil},
		{"var foo = '\\n';", nil},
		{"var foo = 'foo \\r bar';", nil},
		{"var foo = '\\v';", nil},
		{"var foo = '\\f';", nil},
		{"var foo = '\\\n';", nil},
		{"var foo = '\\\r\n';", nil},
		{"<foo attr=\"\\d\"/>", nil},
		{"<div> Testing: \\ </div>", nil},
		{"<div> Testing: &#x5C </div>", nil},
		{"<foo attr='\\d'></foo>", nil},
		{"<> Testing: \\ </>", nil},
		{"<> Testing: &#x5C </>", nil},
		{"var foo = `\\x123`", nil},
		{"var foo = `\\u00a9`", nil},
		{"var foo = `xs\\u2111`", nil},
		{"var foo = `foo \\\\ bar`;", nil},
		{"var foo = `\\t`;", nil},
		{"var foo = `foo \\b bar`;", nil},
		{"var foo = `\\n`;", nil},
		{"var foo = `foo \\r bar`;", nil},
		{"var foo = `\\v`;", nil},
		{"var foo = `\\f`;", nil},
		{"var foo = `\\\n`;", nil},
		{"var foo = `\\\r\n`;", nil},
		{"var foo = `${foo} \\x123`", nil},
		{"var foo = `${foo} \\u00a9`", nil},
		{"var foo = `${foo} xs\\u2111`", nil},
		{"var foo = `${foo} \\\\ ${bar}`;", nil},
		{"var foo = `${foo} \\b ${bar}`;", nil},
		{"var foo = `${foo}\\t`;", nil},
		{"var foo = `${foo}\\n`;", nil},
		{"var foo = `${foo}\\r`;", nil},
		{"var foo = `${foo}\\v`;", nil},
		{"var foo = `${foo}\\f`;", nil},
		{"var foo = `${foo}\\\n`;", nil},
		{"var foo = `${foo}\\\r\n`;", nil},
		{"var foo = `\\``", nil},
		{"var foo = `\\`${foo}\\``", nil},
		{"var foo = `\\${{${foo}`;", nil},
		{"var foo = `$\\{{${foo}`;", nil},
		{"var foo = String.raw`\\.`", nil},
		{"var foo = myFunc`\\.`", nil},
		{"var foo = /[\\d]/", nil},
		{"var foo = /[a\\-b]/", nil},
		{"var foo = /foo\\?/", nil},
		{"var foo = /example\\.com/", nil},
		{"var foo = /foo\\|bar/", nil},
		{"var foo = /\\^bar/", nil},
		{"var foo = /[\\^bar]/", nil},
		{"var foo = /\\(bar\\)/", nil},
		{"var foo = /[[\\]]/", nil},
		{"var foo = /[[]\\./", nil},
		{"var foo = /[\\]\\]]/", nil},
		{"var foo = /\\[abc]/", nil},
		{"var foo = /\\[foo\\.bar]/", nil},
		{"var foo = /vi/m", nil},
		{"var foo = /\\B/", nil},
		{"var foo = /\\0/", nil},
		{"var foo = /\\1/", nil},
		{"var foo = /(a)\\1/", nil},
		{"var foo = /(a)\\12/", nil},
		{"var foo = /[\\0]/", nil},
		{"var foo = 'foo \\  bar'", nil},
		{"var foo = 'foo \\  bar'", nil},
		{"/]/", nil},
		{"/\\]/", nil},
		{"/\\]/u", nil},
		{"var foo = /foo\\]/", nil},
		{"var foo = /[[]\\]/", nil},
		{"var foo = /\\[foo\\.bar\\]/", nil},
		{"var foo = /(?<a>)\\k<a>/", nil},
		{"var foo = /(\\\\?<a>)/", nil},
		{"var foo = /\\p{ASCII}/u", nil},
		{"var foo = /\\P{ASCII}/u", nil},
		{"var foo = /[\\p{ASCII}]/u", nil},
		{"var foo = /[\\P{ASCII}]/u", nil},
		{"/[^^]/", nil},
		{"/[^^]/u", nil},
		{"/[\\q{abc}]/v", nil},
		{"/[\\(]/v", nil},
		{"/[\\)]/v", nil},
		{"/[\\{]/v", nil},
		{"/[\\]]/v", nil},
		{"/[\\}]/v", nil},
		{"/[\\/]/v", nil},
		{"/[\\-]/v", nil},
		{"/[\\|]/v", nil},
		{"/[\\$$]/v", nil},
		{"/[\\&&]/v", nil},
		{"/[\\!!]/v", nil},
		{"/[\\##]/v", nil},
		{"/[\\%%]/v", nil},
		{"/[\\**]/v", nil},
		{"/[\\++]/v", nil},
		{"/[\\,,]/v", nil},
		{"/[\\..]/v", nil},
		{"/[\\::]/v", nil},
		{"/[\\;;]/v", nil},
		{"/[\\<<]/v", nil},
		{"/[\\==]/v", nil},
		{"/[\\>>]/v", nil},
		{"/[\\??]/v", nil},
		{"/[\\@@]/v", nil},
		{"/[\\``]/v", nil},
		{"/[\\~~]/v", nil},
		{"/[^\\^^]/v", nil},
		{"/[_\\^^]/v", nil},
		{"/[$\\$]/v", nil},
		{"/[&\\&]/v", nil},
		{"/[!\\!]/v", nil},
		{"/[#\\#]/v", nil},
		{"/[%\\%]/v", nil},
		{"/[*\\*]/v", nil},
		{"/[+\\+]/v", nil},
		{"/[,\\,]/v", nil},
		{"/[.\\.]/v", nil},
		{"/[:\\:]/v", nil},
		{"/[;\\;]/v", nil},
		{"/[<\\<]/v", nil},
		{"/[=\\=]/v", nil},
		{"/[>\\>]/v", nil},
		{"/[?\\?]/v", nil},
		{"/[@\\@]/v", nil},
		{"/[`\\`]/v", nil},
		{"/[~\\~]/v", nil},
		{"/[^^\\^]/v", nil},
		{"/[_^\\^]/v", nil},
		{"/[\\&&&\\&]/v", nil},
		{"/[[\\-]\\-]/v", nil},
		{"/[\\^]/v", nil},
		{"const s = \"⚓💥⚓\";\n/[\\&&]/v", nil},
		{"const s = \"⚓💥⚓\";\n/[&\\&]/v", nil},
		{"var foo = /\\#/;", NoUselessEscapeOptions{AllowRegexCharacters: []string{"#"}}},
		{"var foo = /\\;/;", NoUselessEscapeOptions{AllowRegexCharacters: []string{";"}}},
		{"var foo = /\\#\\;/;", NoUselessEscapeOptions{AllowRegexCharacters: []string{"#", ";"}}},
		{"var foo = /[ab\\-]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"-"}}},
		{"var foo = /[\\-ab]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"-"}}},
		{"var foo = /[ab\\?]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"?"}}},
		{"var foo = /[ab\\.]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
		{"var foo = /[a\\|b]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"|"}}},
		{"var foo = /\\-/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"-"}}},
		{"var foo = /[\\-]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"-"}}},
		{"var foo = /[ab\\$]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"$"}}},
		{"var foo = /[\\(paren]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"("}}},
		{"var foo = /[\\[]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"["}}},
		{"var foo = /[\\/]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"/"}}},
		{"var foo = /[\\B]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"B"}}},
		{"var foo = /[a][\\-b]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"-"}}},
		{"var foo = /\\-[]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"-"}}},
		{"var foo = /[a\\^]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"^"}}},
		{"/[^\\^]/", NoUselessEscapeOptions{AllowRegexCharacters: []string{"^"}}},
		{"/[^\\^]/u", NoUselessEscapeOptions{AllowRegexCharacters: []string{"^"}}},
		{"/[\\$]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"$"}}},
		{"/[\\&\\&]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"&"}}},
		{"/[\\!!]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"!"}}},
		{"/[\\##]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"#"}}},
		{"/[\\%%]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"%"}}},
		{"/[\\*\\*]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"*"}}},
		{"/[\\+\\+]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"+"}}},
		{"/[\\,,]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{","}}},
		{"/[\\..]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
		{"/[\\:\\:]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{":"}}},
		{"/[\\;\\;]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{";"}}},
		{"/[\\<\\<]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"<"}}},
		{"/[\\=\\=]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"="}}},
		{"/[\\>\\>]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{">"}}},
		{"/[\\?\\?]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"?"}}},
		{"/[\\@\\@]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"@"}}},
		{"/[\\``]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"`"}}},
		{"/[\\~\\~]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"~"}}},
		{"/[^\\^\\^]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"^"}}},
		{"/[_\\^\\^]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"^"}}},
		{"/[\\&\\&&\\&]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"&"}}},
		{"/[\\p{ASCII}--\\.]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
		{"/[\\p{ASCII}&&\\.]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
		// Not upstream's. A wide escape carries a payload the scan must step over whole, and nothing in
		// upstream's corpus puts a backslash inside one. Hardcoding the escape width to two bytes
		// therefore survived a green suite while reporting a phantom finding on each of these: the scan
		// resumed inside the braces and read the interior backslash as an escape of its own. A property
		// escape is one token however long its name is, and a `\\q` string disjunction is one token
		// inside a v-flag class, so the interior of both is pattern text rather than escapes.
		{"var foo = /\\p{AS\\#CII}/u;", nil},
		{"var foo = /[\\p{AS\\#CII}]/u;", nil},
		{"var foo = /[\\q{a\\#b}]/v;", nil},
		{"/[\\.--[.&]]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
		{"/[\\.&&[.&]]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
		{"/[\\.--\\.--\\.]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
		{"/[\\.&&\\.&&\\.]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
		{"/[[\\.&]--[\\.&]]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
		{"/[[\\.&]&&[\\.&]]/v", NoUselessEscapeOptions{AllowRegexCharacters: []string{"."}}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunWithOptions(t, NoUselessEscape, uselessEscapeFile,
					testCase.sourceText, testCase.options))
		})
	}
}

// TestNoUselessEscapeRepairsSource applies the fixes and compares the resulting text.
//
// Brief-mandated and it is the assertion that matters most here: this rule's whole output is a span
// and a one-character replacement, and `ExpectFindings` sees neither. A fix that writes the right
// character one byte to the left deletes the character before the backslash and leaves the backslash
// standing, and every count-based fixture passes while the rewrite corrupts the file.
//
// The expectations are upstream's own `fix` vector, verbatim.
func TestNoUselessEscapeRepairsSource(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSource string
	}{
		{"var foo = /\\#/;", "var foo = /#/;"},
		{"var foo = /\\;/;", "var foo = /;/;"},
		{"var foo = \"\\'\";", "var foo = \"'\";"},
		{"var foo = \"\\#/\";", "var foo = \"#/\";"},
		{"var foo = \"\\a\"", "var foo = \"a\""},
		{"var foo = \"\\B\";", "var foo = \"B\";"},
		{"var foo = \"\\@\";", "var foo = \"@\";"},
		{"var foo = \"foo \\a bar\";", "var foo = \"foo a bar\";"},
		{"var foo = '\\\"';", "var foo = '\"';"},
		{"var foo = '\\#';", "var foo = '#';"},
		{"var foo = '\\$';", "var foo = '$';"},
		{"var foo = '\\p';", "var foo = 'p';"},
		// Upstream expects `p\a@` here, and the difference is its fixer rather than its rule. oxc
		// treats boundary-adjacent fixes as overlapping to match ESLint, so of the three fixes this
		// input produces it applies the first and third and leaves the middle escape standing until a
		// second pass. Our engine states the opposite policy at internal/fix/apply.go: adjacent edits
		// compose safely and are admitted. Both agree on the three findings; only the number of passes
		// needed to land them differs, and one pass is the better answer.
		{"var foo = '\\p\\a\\@';", "var foo = 'pa@';"},
		{"<foo attr={\"\\d\"}/>", "<foo attr={\"d\"}/>"},
		{"var foo = '\\`';", "var foo = '`';"},
		{"var foo = `\\\"`;", "var foo = `\"`;"},
		{"var foo = `\\'`;", "var foo = `'`;"},
		{"var foo = `\\#`;", "var foo = `#`;"},
		{"var foo = '\\`foo\\`';", "var foo = '`foo`';"},
		{"var foo = `\\\"${foo}\\\"`;", "var foo = `\"${foo}\"`;"},
		{"var foo = `\\'${foo}\\'`;", "var foo = `'${foo}'`;"},
		{"var foo = `\\#${foo}`;", "var foo = `#${foo}`;"},
		{"let foo = '\\ ';", "let foo = ' ';"},
		{"let foo = /\\ /;", "let foo = / /;"},
		{"var foo = `\\$\\{{${foo}`;", "var foo = `$\\{{${foo}`;"},
		{"\"use\\ strict\";", "\"use strict\";"},
		{"({ foo() { \"foo\"; \"bar\"; \"ba\\z\" } })", "({ foo() { \"foo\"; \"bar\"; \"baz\" } })"},
		{"/[^\\^]/", "/[^^]/"},
		{"/[^\\^]/u", "/[^^]/u"},
		{"/[\\$]/v", "/[$]/v"},
		{"/[\\&\\&]/v", "/[&\\&]/v"},
		{"/[\\!\\!]/v", "/[!\\!]/v"},
		{"/[\\#\\#]/v", "/[#\\#]/v"},
		{"/[\\%\\%]/v", "/[%\\%]/v"},
		{"/[\\*\\*]/v", "/[*\\*]/v"},
		{"/[\\+\\+]/v", "/[+\\+]/v"},
		{"/[\\,\\,]/v", "/[,\\,]/v"},
		{"/[\\.\\.]/v", "/[.\\.]/v"},
		{"/[\\:\\:]/v", "/[:\\:]/v"},
		{"/[\\;\\;]/v", "/[;\\;]/v"},
		{"/[\\<\\<]/v", "/[<\\<]/v"},
		{"/[\\=\\=]/v", "/[=\\=]/v"},
		{"/[\\>\\>]/v", "/[>\\>]/v"},
		{"/[\\?\\?]/v", "/[?\\?]/v"},
		{"/[\\@\\@]/v", "/[@\\@]/v"},
		{"/[\\`\\`]/v", "/[`\\`]/v"},
		{"/[\\~\\~]/v", "/[~\\~]/v"},
		{"/[^\\^\\^]/v", "/[^^\\^]/v"},
		{"/[_\\^\\^]/v", "/[_^\\^]/v"},
		{"/[\\&\\&&\\&]/v", "/[&\\&&\\&]/v"},
		{"/[\\p{ASCII}--\\.]/v", "/[\\p{ASCII}--.]/v"},
		{"/[\\p{ASCII}&&\\.]/v", "/[\\p{ASCII}&&.]/v"},
		{"/[\\.--[.&]]/v", "/[.--[.&]]/v"},
		{"/[\\.&&[.&]]/v", "/[.&&[.&]]/v"},
		{"/[\\.--\\.--\\.]/v", "/[.--.--.]/v"},
		{"/[\\.&&\\.&&\\.]/v", "/[.&&.&&.]/v"},
		{"/[[\\.&]--[\\.&]]/v", "/[[.&]--[.&]]/v"},
		{"/[[\\.&]&&[\\.&]]/v", "/[[.&]&&[.&]]/v"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoUselessEscape, uselessEscapeFile, testCase.sourceText),
				testCase.wantSource)
		})
	}
}
