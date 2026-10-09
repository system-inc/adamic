package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// octalFile is where the fixtures pretend to live.
const octalFile = "/repository/source/Octal.ts"

// There is no imported corpus for this rule, and that is the most important thing to know before
// reading the fixtures below.
//
// oxc has no `no_octal.rs`. Confirmed by grepping the whole `oxc_linter` crate for `no-octal` and
// `NoOctal` and getting zero from both, with `no-with` and `NoWith` run as controls through the
// same commands and hitting. So there is no inline Rust corpus and no snapshot diagnostic count.
// ESLint is the port target instead, and ESLint does not ship its `tests/` directory, so its corpus
// is genuinely unavailable rather than merely unread.
//
// Every case below is therefore one this port invented, which makes this a weaker artifact than a
// rule with an imported floor. The usual protection is that upstream's clean cases were each added
// when somebody hit that bug, so they catch a porter who reasoned the way the porter is about to
// reason. Nobody wrote them here.
//
// Two things were done instead of relying on that floor.
//
// First, the expectations were taken from running ESLint's own rule rather than from reading its
// regular expression and predicting. That matters: reading `/^0\d/u` and reasoning about what a
// legacy octal is produces the confident and wrong belief that `08` and `09` are clean, since they
// are legal decimals and not octal at all. Running the rule shows it reports them. So the fixtures
// below encode measured upstream behavior rather than a porter's model of it, and where they look
// surprising, that surprise is the measurement.
//
// Second, the mutation sweep was run harder than usual, because with no imported corpus these
// fixtures are the only evidence, and a surviving mutant here is a genuine hole rather than an
// upstream blind spot inherited honestly.

// A legacy octal is a leading `0` followed immediately by a decimal digit, and every one of these
// was confirmed reported by ESLint's own rule before it was written down here.
//
// The last four are the ones a porter gets wrong. `08`, `09`, `019` and `0195` are not octal: `8`
// and `9` are not octal digits, so these are NonOctalDecimalIntegerLiterals and evaluate in base
// ten. Upstream reports them anyway, and deliberately, because they are the same banned legacy
// grammar production and are a syntax error in strict mode for the same reason the octals are. A
// port that "corrects" this by checking the digits are octal-legal diverges from upstream silently
// and is the single most likely way to get this rule wrong.
func TestNoOctalFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the canonical file mode", "var a = 0755;"},
		{"a single octal digit", "var a = 07;"},
		{"all zeroes, which is still the legacy form", "var a = 000;"},
		{"a digit eight, a legal decimal upstream still reports", "var a = 08;"},
		{"a digit nine, likewise", "var a = 09;"},
		{"a two-digit non-octal decimal", "var a = 019;"},
		{"a longer non-octal decimal", "var a = 0195;"},
		{"an octal in a subscript rather than an initializer", "a[0755];"},
		{"an octal under a unary minus, which reports on the literal", "var a = -0755;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoOctal, octalFile, testCase.sourceText), "noOctal")
		})
	}
}

// The clean cases are the whole rule, since firing on a leading zero is easy and declining is the
// part that takes judgment. Each of these was confirmed clean by running ESLint's own rule.
//
// They fail in different ways and each pins a distinct mistake:
//
//	`0`         a bare zero has no following digit. A rule testing only `raw[0] == '0'` reports it,
//	            and zero is the most common literal in any codebase, so that mistake is loud.
//	`0.5`, `0.0` the character after the zero is a dot rather than a digit.
//	`0e5`       likewise an exponent marker, and its value is zero, so a rule reasoning about the
//	            value rather than the spelling has no way to tell it from `0`.
//	`0x1F`      hexadecimal. Modern-prefixed and not the legacy grammar.
//	`0o755`     the modern octal, which is the fix upstream wants and must never be reported. A rule
//	            that lowercases or normalizes before matching reports the very form it is asking for.
//	`0O755`     the same with a capital `O`, which is the case-folding mistake specifically.
//	`0b11`      binary, the third modern prefix.
//	`1e3`, `.5` no leading zero at all, the negative controls.
//	the string  `"0755"` has the text but is not a numeric literal, and the rule keys on the node
//	            being a number. A rule matching on source text reports it.
//	the comment likewise, and a comment is not a node at all.
func TestNoOctalStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare zero", "var a = 0;"},
		{"a fraction whose integer part is zero", "var a = 0.5;"},
		{"a zero-valued fraction", "var a = 0.0;"},
		{"a zero-valued exponent", "var a = 0e5;"},
		{"a hexadecimal literal", "var a = 0x1F;"},
		{"the modern octal, which is what upstream is asking for", "var a = 0o755;"},
		{"the modern octal with a capital prefix", "var a = 0O755;"},
		{"a binary literal", "var a = 0b11;"},
		{"an exponent with no leading zero", "var a = 1e3;"},
		{"a fraction with no leading zero", "var a = .5;"},
		{"the digits inside a string", "var a = \"0755\";"},
		{"the digits inside a single-quoted string", "var a = '0755';"},
		{"the digits inside a template literal", "var a = `0755`;"},
		{"the digits inside a line comment", "var a = 1; // 0755"},
		{"the digits inside a block comment", "/* 0755 */ var a = 1;"},
		{"the digits inside an identifier", "var a0755 = 1;"},
		// Not a legacy octal by upstream's own predicate: the character after the zero is an
		// underscore, so `/^0\d/u` returns false. Our parser recovers from it where ESLint's
		// rejects the file outright, and `tsc` reports it separately as TS6188.
		{"a leading zero followed by a numeric separator", "var a = 0_7_5_5;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoOctal, octalFile, testCase.sourceText))
		})
	}
}

// Cases written from reading *our* parser rather than upstream's, because this is where the port
// stops being a transcription.
//
// TypeScript's parser recovers from source ESLint's parser rejects outright. `01.5`, `0777.5`,
// `0755n` and `0_7_5_5` are all syntax errors in real JavaScript, so ESLint reports a fatal parse
// error and its rule never runs on them at all. Ours parses them anyway, which means this port has
// to decide what to do with inputs upstream has no opinion about, and the decision is invisible
// unless it is written down.
//
// The choice is to report, on the ground that the rule's subject is present in the source however
// the parser recovered. Each is pinned here so the choice is a fixture rather than an accident:
//
//	`01.5`     recovers as two literals, `01` and `.5`. The `01` is the legacy octal, and the rule
//	           reports exactly once because `.5` has no leading zero. A port that read the
//	           declaration's full text rather than the literal's own span would see `01.5` and could
//	           just as easily report twice or zero times.
//	`0777.5`   the same shape with more digits, which distinguishes a rule keying on length.
//	`0755n`    a BigInt suffix on a legacy octal, which is illegal. Our parser drops the `n` and
//	           hands back a plain `KindNumericLiteral` whose raw span is `0755`, so this reports for
//	           the same reason `0755` does. Worth pinning because the obvious expectation is that a
//	           BigInt literal arrives as `KindBigIntLiteral` and is therefore exempt, which is how
//	           `no-loss-of-precision` reaches its own BigInt exemption. That reasoning does not
//	           transfer here and the probe is what showed it.
//
// `0_7_5_5` is deliberately not in this list, and it is the case that corrected the port rather than
// the port correcting it. It was written here first, on the reasoning that a numeric separator is
// illegal in a legacy octal and our parser accepts it, so the recovered literal is still the legacy
// form and should report. The rule declined it and the fixture failed. The rule was right: the raw
// span is `0_7_5_5`, whose second character is an underscore rather than a digit, and running
// ESLint's own `/^0\d/u` against that string returns false. So declining is upstream-faithful and
// not a gap, `tsc` catches the input separately as TS6188, and the case lives in the silent list
// with the other declines. Written down because the failing fixture was the porter being wrong.
func TestNoOctalOnSourceOurParserAcceptsAndESLintRejects(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a legacy octal with a fractional part", "var a = 01.5;"},
		{"a longer legacy octal with a fractional part", "var a = 0777.5;"},
		{"a legacy octal carrying a BigInt suffix", "var a = 0755n;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoOctal, octalFile, testCase.sourceText), "noOctal")
		})
	}
}

// The span, which every message-id assertion above is structurally unable to see.
//
// This rule carries no fix and no suggestion, so the instinct is that a span cannot be wrong in a
// way that matters. That instinct is what shipped a rule whose 187 green fixtures all pointed at the
// wrong place. The finding's range is the whole user-visible output here, so it is the thing to
// assert, and it is asserted by slicing the source with the finding's own range rather than by
// comparing offsets computed the same way the rule computed them.
//
// The literal alone is the right answer, matching ESLint's own columns: `var a = 0755;` reports
// columns 9 through 13, which is `0755` and not the declaration, not the initializer, and not the
// statement. `-0755` reports at column 10, so the sign is outside the span, which is a consequence
// of the parser making the minus a unary operator over a positive literal rather than part of it.
func TestNoOctalReportsTheLiteralSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"a plain octal", "var a = 0755;", "0755"},
		{"a single digit", "var a = 07;", "07"},
		{"under a unary minus, the sign is not part of it", "var a = -0755;", "0755"},
		{"in a subscript", "a[0755];", "0755"},
		{"not the fractional tail the parser split off", "var a = 01.5;", "01"},
		{"the raw digits rather than the BigInt spelling", "var a = 0755n;", "0755"},
		{"a non-octal decimal", "var a = 0195;", "0195"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoOctal, octalFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noOctal")
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Errorf("reported span = %q, want %q", reported, testCase.want)
			}
		})
	}
}

// The rule must read raw source text and not the parsed value, and this is the case that proves it.
//
// Our parser hands back a *cooked* `Text()`: `0755` arrives as `"493"` and `0x1F` arrives as `"31"`,
// measured with a probe rather than assumed. So a rule asking the node what it says gets the decimal
// expansion, in which the leading zero the rule exists to find has already been erased. Reading
// `.Text()` would make this rule silent on every input it is for, and every clean case above would
// still pass, vacuously.
//
// The pairs below are chosen so the cooked spellings collide across the fire and clean cases: `0755`
// and `0o755` both cook to `493`, so no rule reading cooked text can separate them, and one of them
// is upstream's recommended fix. That is what makes this a discrimination rather than a restatement.
func TestNoOctalReadsRawTextRatherThanTheCookedValue(t *testing.T) {
	t.Parallel()

	// Both cook to "493". The legacy spelling reports and the modern one must not.
	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoOctal, octalFile, "var a = 0755;"), "noOctal")
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoOctal, octalFile, "var a = 0o755;"))

	// `07` cooks to "7", indistinguishable from a plain `7`, which must stay clean.
	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoOctal, octalFile, "var a = 07;"), "noOctal")
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoOctal, octalFile, "var a = 7;"))

	// `000` and `0e5` both cook to "0", as does a bare `0`. Only the first is the legacy form.
	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoOctal, octalFile, "var a = 000;"), "noOctal")
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoOctal, octalFile, "var a = 0e5;"))
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoOctal, octalFile, "var a = 0;"))
}

// Several literals in one file report several times, and each finding points at its own literal.
//
// Every other fixture here is a single-literal file, so a rule that reported once per file, or that
// cached the first raw span it computed and reused it, would pass all of them. This is the case that
// separates per-node work from per-file work.
func TestNoOctalReportsEachLiteralSeparately(t *testing.T) {
	t.Parallel()

	source := "var a = 0755, b = 0o755, c = 07, d = 0, e = 08;"
	result := rule_testing.Run(t, NoOctal, octalFile, source)
	rule_testing.ExpectFindings(t, result, "noOctal", "noOctal", "noOctal")

	want := []string{"0755", "07", "08"}
	for index, diagnostic := range result.Diagnostics {
		reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != want[index] {
			t.Errorf("finding %d reported %q, want %q", index, reported, want[index])
		}
	}
}
