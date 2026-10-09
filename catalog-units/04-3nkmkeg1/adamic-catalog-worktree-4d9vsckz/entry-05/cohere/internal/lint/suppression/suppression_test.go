package suppression

import (
	"strings"
	"testing"
)

// TestNextLineSuppressesTheLineBelow covers 372 of the 387 suppressions in the codebase this was
// built against. If only one form works, it has to be this one.
func TestNextLineSuppressesTheLineBelow(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"const before = 1;",
		"// eslint-disable-next-line nexus/consistency-no-enum -- the fixture has to be a real enum.",
		"enum Suppressed {}",
		"enum Reported {}",
	}, "\n")

	index := Build(source)

	if !index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 2)) {
		t.Fatal("the line below the directive should be suppressed")
	}
	if index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 3)) {
		t.Fatal("a -next-line directive must not reach two lines down")
	}
	if index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 0)) {
		t.Fatal("a -next-line directive must not reach backward")
	}
}

// TestUnsuppressedFires is the half a violation-only corpus never has. A suppression index that
// returned true for everything would pass every test above this one.
func TestUnsuppressedFires(t *testing.T) {
	t.Parallel()
	index := Build("enum Reported {}\n")

	if index.Suppresses("nexus/consistency-no-enum", 0) {
		t.Fatal("a file with no directives must suppress nothing")
	}
}

// TestDirectiveNamingAnotherRuleDoesNotSilenceThisOne is the guard the task names explicitly. A
// directive that silenced every rule regardless of the name written would look correct on every
// single-rule fixture.
func TestDirectiveNamingAnotherRuleDoesNotSilenceThisOne(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"// eslint-disable-next-line nexus/consistency-no-abbreviated-identifier -- wire contract.",
		"enum Reported {}",
	}, "\n")

	index := Build(source)

	if index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 1)) {
		t.Fatal("a directive naming a different rule must not silence this one")
	}
	if !index.Suppresses("nexus/consistency-no-abbreviated-identifier", offsetOfLine(source, 1)) {
		t.Fatal("the directive should still silence the rule it does name")
	}
}

func TestSameLineSuppressesItsOwnLine(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"const first = 1;",
		"const value = 2; // eslint-disable-line nexus/consistency-no-ambiguous-identifier",
	}, "\n")

	index := Build(source)

	if !index.Suppresses("nexus/consistency-no-ambiguous-identifier", offsetOfLine(source, 1)) {
		t.Fatal("a -line directive should suppress its own line")
	}
	if index.Suppresses("nexus/consistency-no-ambiguous-identifier", offsetOfLine(source, 0)) {
		t.Fatal("a -line directive must not reach the line above")
	}
}

// TestFileScopeCoversFromItsOwnLineDown proves the file-level form does not reach backward. A
// directive that covered the whole file regardless of position would silence findings written above
// the decision to suppress them.
func TestFileScopeCoversFromItsOwnLineDown(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"enum Above {}",
		"/* eslint-disable nexus/consistency-no-enum */",
		"enum Below {}",
		"enum FurtherBelow {}",
	}, "\n")

	index := Build(source)

	if index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 0)) {
		t.Fatal("a file-level directive must not reach above itself")
	}
	if !index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 2)) {
		t.Fatal("a file-level directive should cover the line below it")
	}
	if !index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 3)) {
		t.Fatal("a file-level directive should cover every later line")
	}
}

// TestScopeSuffixesAreMatchedLongestFirst is the highest-consequence parse bug available here.
// Testing the bare `-disable` prefix before `-disable-next-line` reads every one of the 372
// next-line directives as a file-level disable, silencing the rest of each file with no output
// saying so.
func TestScopeSuffixesAreMatchedLongestFirst(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"// eslint-disable-next-line nexus/consistency-no-enum -- one line only.",
		"enum Suppressed {}",
		"enum MustStillFire {}",
	}, "\n")

	index := Build(source)

	if directives := index.Directives(); len(directives) != 1 || directives[0].Kind != KindNextLine {
		t.Fatalf("expected one next-line directive, got %+v", directives)
	}
	if index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 2)) {
		t.Fatal("a -next-line directive read as file-level would silence the rest of the file")
	}
}

func TestCommaSeparatedRuleListsAreHonored(t *testing.T) {
	t.Parallel()
	// The real line, from libraries/structure/GlobalDeclarations.d.ts:13.
	source := strings.Join([]string{
		"// eslint-disable-next-line nexus/consistency-require-type-suffix, nexus/consistency-no-abbreviated-identifier",
		"type Thing = string;",
	}, "\n")

	index := Build(source)

	position := offsetOfLine(source, 1)
	if !index.Suppresses("nexus/consistency-require-type-suffix", position) {
		t.Fatal("the first name in the list should be honored")
	}
	if !index.Suppresses("nexus/consistency-no-abbreviated-identifier", position) {
		t.Fatal("the second name in the list should be honored")
	}
	if index.Suppresses("nexus/consistency-no-enum", position) {
		t.Fatal("a rule not in the list must not be silenced")
	}
}

func TestDirectiveWithNoRuleNameCoversEveryRule(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"// eslint-disable-next-line",
		"enum Suppressed {}",
	}, "\n")

	index := Build(source)

	if !index.Suppresses("any/rule-at-all", offsetOfLine(source, 1)) {
		t.Fatal("a directive naming no rule should cover every rule")
	}
}

func TestReasonIsParsedAndItsAbsenceIsVisible(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"// eslint-disable-next-line nexus/consistency-no-enum -- EnumLike describes a real enum.",
		"enum WithReason {}",
		"// eslint-disable-next-line nexus/consistency-no-enum",
		"enum WithoutReason {}",
	}, "\n")

	directives := Build(source).Directives()
	if len(directives) != 2 {
		t.Fatalf("expected 2 directives, got %d", len(directives))
	}

	if directives[0].Reason != "EnumLike describes a real enum." {
		t.Fatalf("reason not parsed: %q", directives[0].Reason)
	}
	if !directives[0].HasReason() {
		t.Fatal("a directive with a reason should report having one")
	}
	if directives[1].HasReason() {
		t.Fatalf("a directive with no reason reported one: %q", directives[1].Reason)
	}
}

// TestSuppressionStillAppliesWithoutAReason is the decision this package makes deliberately. The
// missing reason is reported as its own finding elsewhere; it does not resurrect the suppressed one.
// Measured against the codebase cohere gates: 281 of the 306 suppressions naming one of our own
// rules state no reason, and refusing to honor them would turn one missing convention into hundreds
// of unrelated failures.
func TestSuppressionStillAppliesWithoutAReason(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"// eslint-disable-next-line nexus/consistency-no-enum",
		"enum Suppressed {}",
	}, "\n")

	if !Build(source).Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 1)) {
		t.Fatal("a reasonless suppression must still suppress")
	}
}

// TestAppliedCountSeparatesSilencedFromAbsent is what makes the coverage line honest. A run that
// suppressed forty findings and a run that found none print the same green otherwise.
func TestAppliedCountSeparatesSilencedFromAbsent(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"// eslint-disable-next-line nexus/consistency-no-enum -- deliberate.",
		"enum Suppressed {}",
		"// eslint-disable-next-line nexus/consistency-no-enum -- guards nothing.",
		"const clean = 1;",
	}, "\n")

	index := Build(source)
	index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 1))

	if index.AppliedCount(0) != 1 {
		t.Fatalf("expected the first directive to have silenced 1 finding, got %d", index.AppliedCount(0))
	}
	if index.AppliedCount(1) != 0 {
		t.Fatalf("a directive that silenced nothing should count 0, got %d", index.AppliedCount(1))
	}
}

// TestCommentTextInsideStringsIsNotADirective is the scanner's real job. A scan that read string
// contents would invent suppressions out of test data and documentation.
func TestCommentTextInsideStringsIsNotADirective(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"double quotes": "const sample = \"// eslint-disable-next-line nexus/consistency-no-enum\";\nenum Reported {}",
		"single quotes": "const sample = '// eslint-disable-next-line nexus/consistency-no-enum';\nenum Reported {}",
		"template":      "const sample = `// eslint-disable-next-line nexus/consistency-no-enum`;\nenum Reported {}",
		"regex class":   "const pattern = /[/]\\/\\/ eslint-disable-next-line nexus\\/consistency-no-enum/;\nenum Reported {}",
	}

	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			index := Build(source)
			if directives := index.Directives(); len(directives) != 0 {
				t.Fatalf("found a directive inside a literal: %+v", directives)
			}
		})
	}
}

// TestDivisionIsNotARegularExpression guards the other side of the same scan. Reading `a / b` as a
// pattern would swallow source up to the next slash, hiding every directive in between.
func TestDivisionIsNotARegularExpression(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"const ratio = width / height;",
		"const other = total / count;",
		"// eslint-disable-next-line nexus/consistency-no-enum -- still visible.",
		"enum Suppressed {}",
	}, "\n")

	index := Build(source)

	if len(index.Directives()) != 1 {
		t.Fatalf("division swallowed the directive: %+v", index.Directives())
	}
	if !index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 3)) {
		t.Fatal("the directive after two divisions should still apply")
	}
}

func TestJsxCommentFormIsHonored(t *testing.T) {
	t.Parallel()
	// The shape 14 next/no-img-element suppressions take in .tsx files.
	source := strings.Join([]string{
		"const element = (",
		"  <div>",
		"    {/* eslint-disable-next-line @next/next/no-img-element -- served-route bytes */}",
		"    <img src={source} />",
		"  </div>",
		");",
	}, "\n")

	index := Build(source)

	directives := index.Directives()
	if len(directives) != 1 {
		t.Fatalf("expected 1 directive, got %+v", directives)
	}
	if directives[0].Reason != "served-route bytes" {
		t.Fatalf("reason not parsed from a JSX comment: %q", directives[0].Reason)
	}
	if !index.Suppresses("@next/next/no-img-element", offsetOfLine(source, 3)) {
		t.Fatal("a JSX-wrapped directive should suppress the line below it")
	}
}

func TestBlockCommentNextLineFormIsHonored(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"/* eslint-disable-next-line @next/next/no-img-element -- served-route bytes */",
		"const element = 1;",
	}, "\n")

	index := Build(source)

	directives := index.Directives()
	if len(directives) != 1 || directives[0].Kind != KindNextLine {
		t.Fatalf("block-comment next-line form not recognized: %+v", directives)
	}
	if directives[0].Reason != "served-route bytes" {
		t.Fatalf("reason not parsed: %q", directives[0].Reason)
	}
}

// TestNearMissesAreNotDirectives keeps the parser from claiming comments that merely talk about
// suppression. Documentation mentioning the syntax is common and must not become a suppression.
func TestNearMissesAreNotDirectives(t *testing.T) {
	t.Parallel()
	sources := []string{
		"// eslint-disable-nonsense nexus/consistency-no-enum\nenum Reported {}",
		"// prettier-ignore\nenum Reported {}",
		"// this line explains eslint-disable-next-line without being one\nenum Reported {}",
	}

	for _, source := range sources {
		index := Build(source)
		if directives := index.Directives(); len(directives) != 0 {
			t.Fatalf("near miss claimed as a directive in %q: %+v", source, directives)
		}
	}
}

func TestStackedDirectivesBothApply(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"// eslint-disable-next-line nexus/consistency-no-enum -- first.",
		"// eslint-disable-next-line nexus/consistency-no-abbreviated-identifier -- second.",
		"enum Thing {}",
	}, "\n")

	index := Build(source)

	// The first directive covers line 1, which is the second comment rather than the code. Only the
	// second directive reaches the declaration, and that is the behavior eslint has: a stacked pair
	// does not both reach past an intervening comment line.
	position := offsetOfLine(source, 2)
	if !index.Suppresses("nexus/consistency-no-abbreviated-identifier", position) {
		t.Fatal("the directive directly above the code should apply")
	}
	if index.Suppresses("nexus/consistency-no-enum", position) {
		t.Fatal("a directive separated from the code by another comment must not reach it")
	}
}

func TestEmptyFileHasNoDirectives(t *testing.T) {
	t.Parallel()
	index := Build("")

	if len(index.Directives()) != 0 {
		t.Fatal("an empty file should produce no directives")
	}
	if index.Suppresses("any/rule", 0) {
		t.Fatal("an empty file should suppress nothing")
	}
}

func TestUnterminatedBlockCommentDoesNotHang(t *testing.T) {
	t.Parallel()
	index := Build("/* eslint-disable nexus/consistency-no-enum\nenum Thing {}")

	if len(index.Directives()) != 1 {
		t.Fatalf("an unterminated block comment should still parse as one directive: %+v", index.Directives())
	}
}

func TestOxlintAndVerifySpellingsAreAccepted(t *testing.T) {
	t.Parallel()
	for _, tool := range []string{"eslint", "oxlint", "cohere"} {
		source := "// " + tool + "-disable-next-line nexus/consistency-no-enum -- accepted.\nenum Thing {}"
		if !Build(source).Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 1)) {
			t.Fatalf("%s-disable-next-line was not honored", tool)
		}
	}
}

func TestLineOfHandlesBoundaries(t *testing.T) {
	t.Parallel()
	source := "a\nb\nc"

	index := Build(source)

	// Keyed by byte offset, valued by the zero-based line that offset falls on. In "a\nb\nc" the
	// three characters sit at offsets 0, 2, and 4.
	for offset, expectedLine := range map[int]int{0: 0, 2: 1, 4: 2} {
		if got := index.LineOf(offset); got != expectedLine {
			t.Fatalf("offset %d: expected line %d, got %d", offset, expectedLine, got)
		}
	}
}

// offsetOfLine returns the byte offset a zero-based line begins at.
func offsetOfLine(source string, line int) int {
	offset := 0
	for current := 0; current < line; current++ {
		next := strings.IndexByte(source[offset:], '\n')
		if next < 0 {
			return len(source)
		}
		offset += next + 1
	}
	return offset
}

// Every honored spelling suppresses, and they share one grammar.
//
// Four spellings reach this parser and the corpus above exercises only `eslint-disable`, so a
// spelling could be dropped from `directives.disableDirectives` and every other test here would still pass.
//
// `cohere-disable` is what new code writes. `verify-disable` is what this tool's own directives were
// called before the rename, and it is honored so a comment written under the old name does not start
// failing because the binary was renamed. `eslint-disable` and `oxlint-disable` are the corpus and
// the gate being replaced, 387 of the first alone, and they are permanent rather than transitional.
//
// The rule name, the ` -- ` reason and the scope suffix are identical across all four, which is the
// property that keeps this from becoming four code paths that drift. Asserted here rather than
// stated, since nothing downstream of parsing knows which spelling it read.
func TestEveryHonoredSpellingSuppresses(t *testing.T) {
	t.Parallel()
	for _, directive := range []string{"cohere-disable", "verify-disable", "eslint-disable", "oxlint-disable"} {
		t.Run(directive, func(t *testing.T) {
			t.Parallel()
			source := strings.Join([]string{
				"const before = 1;",
				"// " + directive + "-next-line nexus/consistency-no-enum -- the fixture has to be a real enum.",
				"enum Suppressed {}",
				"enum Reported {}",
			}, "\n")

			index := Build(source)

			if !index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 2)) {
				t.Fatalf("`%s-next-line` did not suppress the line below it", directive)
			}
			if index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 3)) {
				t.Fatalf("`%s-next-line` reached two lines down", directive)
			}
			if index.Suppresses("nexus/other-rule", offsetOfLine(source, 2)) {
				t.Fatalf("`%s` naming one rule silenced another", directive)
			}
		})
	}
}

// A block directive closes on its own spelling's `-enable`.
//
// The disable and enable lists are separate slices rather than pairs, so a spelling can be added to
// one and forgotten in the other. That would leave a block that opens and never closes, which reads
// as a working suppression while silencing the rest of the file.
func TestEveryHonoredSpellingClosesItsBlock(t *testing.T) {
	t.Parallel()
	for _, directive := range []string{"cohere", "cohere", "eslint", "oxlint"} {
		t.Run(directive, func(t *testing.T) {
			t.Parallel()
			source := strings.Join([]string{
				"/* " + directive + "-disable nexus/consistency-no-enum */",
				"enum Suppressed {}",
				"/* " + directive + "-enable nexus/consistency-no-enum */",
				"enum Reported {}",
			}, "\n")

			index := Build(source)

			if !index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 1)) {
				t.Fatalf("`%s-disable` did not suppress inside its own block", directive)
			}
			if index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 3)) {
				t.Fatalf("`%s-enable` did not close the block, so the rest of the file is silenced", directive)
			}
		})
	}
}

// TestALineCommentFileDisableSuppressesNothing is Kirk's ruling of 2026-10-01: a file-scope disable
// and an enable are directives only as block comments, which is ESLint's grammar (measured on
// 10.8.1). As `//` they are prose, and phi web's build scripts hold exactly that prose, which
// previously disabled every rule for the rest of the file.
//
// Each case pairs the line form, which must change nothing, with the block form, which must still
// work, so the test cannot pass by breaking file scope altogether.
func TestALineCommentFileDisableSuppressesNothing(t *testing.T) {
	t.Parallel()
	prose := strings.Join([]string{
		"// eslint-disable + generated banner keep the linter and future readers out.",
		"enum Reported {}",
		"// eslint-disable nexus/consistency-no-enum",
		"enum AlsoReported {}",
	}, "\n")
	index := Build(prose)
	if len(index.Directives()) != 0 {
		t.Fatalf("a `//` file-scope disable was read as %d directive(s)", len(index.Directives()))
	}
	for _, line := range []int{1, 3} {
		if index.Suppresses("nexus/consistency-no-enum", offsetOfLine(prose, line)) {
			t.Fatalf("line %d was suppressed by a `//` file-scope disable", line)
		}
	}

	block := strings.Join([]string{
		"/* eslint-disable nexus/consistency-no-enum */",
		"enum Suppressed {}",
		"// eslint-enable nexus/consistency-no-enum",
		"enum StillSuppressed {}",
	}, "\n")
	blockIndex := Build(block)
	if !blockIndex.Suppresses("nexus/consistency-no-enum", offsetOfLine(block, 1)) {
		t.Fatal("control: the block form stopped suppressing")
	}
	if !blockIndex.Suppresses("nexus/consistency-no-enum", offsetOfLine(block, 3)) {
		t.Fatal("a `//` enable closed a block, which ESLint does not do")
	}
}

// An interpolation body is code, and the scan has to read it as code or lose its place for the rest of
// the file (#59rmvb5). Base's SqliteAdapter.ts quotes identifiers with the first case's line: the
// regex's `"` was read as a string, the quotes paired wrongly, and every directive after it vanished,
// so a block disable ran to the end of the file and two next-line disables suppressed nothing.
//
// Each case puts something inside `${...}` and then, on later lines, a block enable and a next-line
// disable, so a scan that loses sync anywhere inside the interpolation loses both.
func TestAnInterpolationIsLexedAsCode(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"a regex holding a quote":    "    return `\"${key.replace(/\"/g, '\"\"')}\"`;",
		"an object literal's braces": "    return `${ {width: 1}.width } and ${ {a: {b: 2}}.a.b }`;",
		"a nested template":          "    return `outer ${ `inner ${key} '` } done`;",
		"a division then a regex":    "    return `${total / count} ${/'/.test(key)}`;",
		"a comment holding a quote":  "    return `${key /* it's a key */}`;",
	}

	for name, line := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source := strings.Join([]string{
				"/* eslint-disable nexus/consistency-no-enum -- the block under test */",
				"function quote(key: string): string {",
				line,
				"}",
				"/* eslint-enable nexus/consistency-no-enum -- the block ends here */",
				"enum Reported {}",
				"// eslint-disable-next-line nexus/performance-no-independent-await-in-loop -- sequential by design.",
				"for(const item of items) { await work(item); }",
			}, "\n")

			index := Build(source)

			if !index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 3)) {
				t.Fatal("the control failed: the block disable does not cover its own body, so nothing below proves anything")
			}
			if index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 5)) {
				t.Errorf("the block enable was lost after the interpolation, so the disable ran past it: %+v", index.Directives())
			}
			if !index.Suppresses("nexus/performance-no-independent-await-in-loop", offsetOfLine(source, 7)) {
				t.Errorf("the next-line disable after the interpolation was lost: %+v", index.Directives())
			}
		})
	}
}

// A directive's text inside a string in an interpolation is still a string. Lexing the body as code
// must not turn its literals into comments.
func TestCommentTextInsideAnInterpolatedStringIsNotADirective(t *testing.T) {
	t.Parallel()
	source := "const sample = `${'// eslint-disable-next-line nexus/consistency-no-enum'}`;\nenum Reported {}"
	if directives := Build(source).Directives(); len(directives) != 0 {
		t.Fatalf("found a directive inside a string inside an interpolation: %+v", directives)
	}
}

// The interpolation closes on its own brace, not an object literal's. The fixture for this needs a
// comment after the object literal: a scan that left the interpolation at the literal's `}` reads the
// rest as template text and resynchronizes by the end of the line, so a later directive alone cannot
// tell the two apart, and an earlier fixture built only that way let the brace count be deleted.
func TestAnInterpolationClosesOnItsOwnBrace(t *testing.T) {
	t.Parallel()
	source := strings.Join([]string{
		"const label = `${ {width: 1}.width /* eslint-disable-next-line nexus/consistency-no-enum -- inside the interpolation */ }`;",
		"enum Suppressed {}",
	}, "\n")

	index := Build(source)
	if !index.Suppresses("nexus/consistency-no-enum", offsetOfLine(source, 1)) {
		t.Fatalf("a comment after an object literal inside an interpolation was read as template text: %+v", index.Directives())
	}
}
