package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const preferAsConstFile = "/repository/source/Thing.ts"

// TestPreferAsConstFires covers every upstream failing case that this port reports.
//
// Upstream ships twenty failing inputs and its snapshot holds twenty entries, but three of those
// entries are `Unexpected token` parse errors rather than rule diagnostics: the tester runs .tsx,
// where `<'bar'>'bar'` is a JSX fragment opening rather than a type assertion. That is the whole
// of the extractor's DISCREPANCY warning, 17 diagnostics against 20 inputs, and it resolves to
// exactly one finding per reporting input with nothing left to recover.
func TestPreferAsConstFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"upstream fail 1", "let foo = { bar: 'baz' as 'baz' };"},
		{"upstream fail 2", "let foo = { bar: 1 as 1 };"},
		{"upstream fail 3", "let []: 'bar' = 'bar';"},
		{"upstream fail 4", "let foo: 'bar' = 'bar';"},
		{"upstream fail 5", "let foo: 2 = 2;"},
		{"upstream fail 6", "const example: 'hello' = 'hello';"},
		{"upstream fail 7", "let foo: 'bar' = \"bar\";"},
		{"upstream fail 8", "const foo: 2 = 2;"},
		{"upstream fail 9", "\n            class foo {\n              readonly bar: 'baz' = 'baz';\n            }\n                  "},
		{"upstream fail 10", "\n            class foo {\n              static bar: 2 = 2;\n            }\n                  "},
		{"upstream fail 11", "let foo: 'bar' = 'bar' as 'bar';"},
		{"upstream fail 12", "let foo = 'bar' as 'bar';"},
		{"upstream fail 13", "let foo = 5 as 5;"},
		{"upstream fail 14", "\n            class foo {\n              bar: 'baz' = 'baz';\n            }\n                  "},
		{"upstream fail 15", "\n            class foo {\n              bar: 2 = 2;\n            }\n                  "},
		{"upstream fail 16", "\n            class foo {\n              foo = 'bar' as 'bar';\n            }\n                  "},
		{"upstream fail 17", "\n            class foo {\n              foo = 5 as 5;\n            }\n                  "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferAsConst, preferAsConstFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "preferAsConst")
		})
	}
}

// TestPreferAsConstFixesWriteWhatTheyClaim asserts the rewritten source rather than the message id.
//
// Every pair here is one of upstream's own fix vectors, extracted verbatim from `let fix: vec![]`,
// which the fixture extractor counts but does not print. They are the highest-value artifact this
// rule ships: the annotation arm writes two edits at two separated offsets, and a finding carrying
// a repair that deleted the wrong range satisfies every assertion in TestPreferAsConstFires,
// because the finding is byte-identical either way.
func TestPreferAsConstFixesWriteWhatTheyClaim(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"upstream fix vector 1", "let foo = { bar: 'baz' as 'baz' };", "let foo = { bar: 'baz' as const };"},
		{"upstream fix vector 2", "let foo = { bar: 1 as 1 };", "let foo = { bar: 1 as const };"},
		{"upstream fix vector 3", "let foo: 'bar' = 'bar' as 'bar';", "let foo: 'bar' = 'bar' as const;"},
		{"upstream fix vector 4", "let foo: 'bar' = 'bar';", "let foo = 'bar' as const;"},
		{"upstream fix vector 5", "let foo: 2 = 2;", "let foo = 2 as const;"},
		{"upstream fix vector 6", "const example: 'hello' = 'hello';", "const example = 'hello' as const;"},
		{"upstream fix vector 7", "let foo: 'bar' = \"bar\";", "let foo = \"bar\" as const;"},
		{"upstream fix vector 8", "const foo: 2 = 2;", "const foo = 2 as const;"},
		{"upstream fix vector 9", "let foo = 'bar' as 'bar';", "let foo = 'bar' as const;"},
		{"upstream fix vector 10", "let foo = 5 as 5;", "let foo = 5 as const;"},
		{"upstream fix vector 11", "\n            class foo {\n              readonly bar: 'baz' = 'baz';\n            }\n                  ", "\n            class foo {\n              readonly bar = 'baz' as const;\n            }\n                  "},
		{"upstream fix vector 12", "\n            class foo {\n              static bar: 2 = 2;\n            }\n                  ", "\n            class foo {\n              static bar = 2 as const;\n            }\n                  "},
		{"upstream fix vector 13", "\n            class foo {\n              foo = 'bar' as 'bar';\n            }\n                  ", "\n            class foo {\n              foo = 'bar' as const;\n            }\n                  "},
		{"upstream fix vector 14", "\n            class foo {\n              foo = 5 as 5;\n            }\n                  ", "\n            class foo {\n              foo = 5 as const;\n            }\n                  "},
		// Upstream omits the unmodified class property from its fix vectors, covering only the
		// `readonly` and `static` spellings. Measured against the release binary, the plain form is
		// fixed identically, so the omission is an incomplete vector list rather than a behavior.
		{"a plain class property is fixed too", "class foo {\n  bar: 'baz' = 'baz';\n}\n", "class foo {\n  bar = 'baz' as const;\n}\n"},
		// The annotation delete trims to the colon and leaves the whitespace before it, and the
		// insert lands at the initializer's end rather than after its trailing trivia. Both are
		// measured byte behaviors of upstream, not guesses about what looks tidy.
		{"whitespace before the colon survives the delete", "let a  :  'bar' = 'bar';\n", "let a   = 'bar' as const;\n"},
		{"a comment after the type survives the delete", "let a: 'bar' /*c*/ = 'bar';\n", "let a /*c*/ = 'bar' as const;\n"},
		{"the insert precedes trailing trivia", "let a: 'bar' = 'bar' /*t*/;\n", "let a = 'bar' as const /*t*/;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, PreferAsConst, preferAsConstFile, testCase.sourceText),
				testCase.wantSource)
		})
	}
}

// TestPreferAsConstReportsTheDestructuringCaseWithoutARepair pins the one reporting shape that
// offers nothing to apply.
//
// Upstream passes `can_fix` as `matches!(id, BindingPattern::BindingIdentifier(_))`, so a binding
// pattern reports and proposes no edit: deleting the annotation off `let []: 'bar' = 'bar'` and
// appending `as const` would leave a destructure of a string, which is a different program. The
// absence of this input from upstream's fourteen fix vectors is the corpus agreeing.
//
// ExpectFixedSource returns the source unchanged when a diagnostic carries no fix, so asserting
// the input back is the assertion that no repair was offered.
func TestPreferAsConstReportsTheDestructuringCaseWithoutARepair(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an array binding pattern", "let []: 'bar' = 'bar';"},
		// Upstream writes only the array form. The object form travels the same branch, and a port
		// that special-cased arrays would pass every imported fixture while diverging here.
		{"an object binding pattern", "let {}: 'bar' = 'bar';"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferAsConst, preferAsConstFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "preferAsConst")
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Fatalf("expected no fix on a binding pattern, got %d", len(result.Diagnostics[0].Fixes))
			}
			if len(result.Diagnostics[0].Suggestions) != 0 {
				t.Fatalf("expected no suggestion on a binding pattern, got %d", len(result.Diagnostics[0].Suggestions))
			}
		})
	}
}

// TestPreferAsConstPointsAtTheLiteralType asserts where the finding lands, which no message-id
// assertion can see.
//
// Upstream labels the literal type node, never the annotation and never the initializer. The
// distinction is invisible to ExpectFindings and visible in every snapshot column: `let foo: 'bar'
// = 'bar';` underlines `'bar'` at column 10, the type, not the value at column 18.
//
// The expected text is a literal typed here rather than a reference to the rule's own message or
// span helper, so that a mutation moving the report site cannot move the assertion with it.
func TestPreferAsConstPointsAtTheLiteralType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"a variable annotation", "let foo: 'bar' = 'bar';", "'bar'"},
		{"a numeric variable annotation", "let foo: 2 = 2;", "2"},
		// The value and the type differ in quote style, so a span taken from the initializer instead
		// of the type would still be five bytes and still spell bar. Only the quote marks separate them.
		{"quote styles differ between type and value", "let foo: 'bar' = \"bar\";", "'bar'"},
		{"an as expression", "let foo = 'bar' as 'bar';", "'bar'"},
		{"a class property", "class foo {\n  bar: 'baz' = 'baz';\n}\n", "'baz'"},
		{"a binding pattern", "let []: 'bar' = 'bar';", "'bar'"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, PreferAsConst, preferAsConstFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "preferAsConst")
			finding := result.Diagnostics[0]
			if got := testCase.sourceText[finding.Range.Pos():finding.Range.End()]; got != testCase.wantText {
				t.Fatalf("finding points at %q, want %q", got, testCase.wantText)
			}
			if finding.Message.Id != "preferAsConst" {
				t.Fatalf("message id %q, want %q", finding.Message.Id, "preferAsConst")
			}
		})
	}
}

func TestPreferAsConstStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"upstream pass 1", "let foo = 'baz' as const;"},
		{"upstream pass 2", "let foo = 1 as const;"},
		{"upstream pass 3", "let foo = { bar: 'baz' as const };"},
		{"upstream pass 4", "let foo = { bar: 1 as const };"},
		{"upstream pass 5", "let foo = { bar: 'baz' };"},
		{"upstream pass 6", "let foo = { bar: 2 };"},
		{"upstream pass 7", "let foo = 'bar' as string;"},
		{"upstream pass 8", "let foo = `bar` as `bar`;"},
		{"upstream pass 9", "let foo = `bar` as `foo`;"},
		{"upstream pass 10", "let foo = `bar` as 'bar';"},
		{"upstream pass 11", "let foo: string = 'bar';"},
		{"upstream pass 12", "let foo: number = 1;"},
		{"upstream pass 13", "let foo: 'bar' = baz;"},
		{"upstream pass 14", "let foo: 'bar' = 'baz';"},
		{"upstream pass 15", "let foo: 2 = 3;"},
		{"upstream pass 16", "let foo = 'bar';"},
		{"upstream pass 17", "let foo: 'bar';"},
		{"upstream pass 18", "let foo = { bar };"},
		{"upstream pass 19", "let foo: 'baz' = 'baz' as const;"},
		{"upstream pass 20", "\n                  class foo {\n                    bar = 'baz';\n                  }\n                "},
		{"upstream pass 21", "\n                  class foo {\n                    bar: 'baz';\n                  }\n                "},
		{"upstream pass 22", "\n                  class foo {\n                    bar;\n                  }\n                "},
		{"upstream pass 23", "\n                  class foo {\n                    bar: string = 'baz';\n                  }\n                "},
		{"upstream pass 24", "\n                  class foo {\n                    bar: number = 1;\n                  }\n                "},
		{"upstream pass 25", "\n                  class foo {\n                    bar = 'baz' as const;\n                  }\n                "},
		{"upstream pass 26", "\n                  class foo {\n                    bar = 2 as const;\n                  }\n                "},
		{"upstream pass 27", "\n                  class foo {\n                    get bar(): 'bar' {}\n                    set bar(bar: 'bar') {}\n                  }\n                "},
		{"upstream pass 28", "\n                  class foo {\n                    bar = () => 'bar' as const;\n                  }\n                "},
		{"upstream pass 29", "\n                  type BazFunction = () => 'baz';\n                  class foo {\n                    bar: BazFunction = () => 'bar';\n                  }\n                "},
		{"upstream pass 30", "\n                  class foo {\n                    bar(): void {}\n                  }\n                "},

		// The three angle-bracket assertions upstream lists as failures. They are failures in its
		// tester only because it parses .tsx, where the snapshot records `Unexpected token` rather
		// than a rule diagnostic. Measured in .ts against the release binary, with a control rule
		// firing on a neighbouring file, all three are silent: oxc registers no TSTypeAssertion arm
		// at all. This is a real gap against typescript-eslint, which does register one and would
		// report. Upstream knows: its own fix vectors for these three are commented out. Reproduced
		// as silence because oxc is what the differential harness compares against.
		{"an angle bracket type assertion", "\n            class foo {\n              foo = <'bar'>'bar';\n            }\n                  "},
		{"an angle bracket type assertion", "let foo = <'bar'>'bar';"},
		{"an angle bracket type assertion", "let foo = <4>4;"},

		// Positions that carry a literal annotation and an equal literal value, and are still silent.
		// Upstream registers exactly three arms, so every other annotated position is out of scope,
		// and each of these was measured rather than reasoned about.
		{"a parameter default", "function f(p: 'bar' = 'bar') {}"},
		{"a generic type argument", "declare function g<T>(x: T): T;\nconst r = g<'bar'>('bar');"},
		// A parenthesized type parses to ParenthesizedType rather than LiteralType, so it declines
		// without a paren skip. Adding one reads as a free correctness improvement and would be a
		// divergence: measured silent upstream.
		{"a parenthesized literal type", "let a: ('bar') = 'bar';"},
		// Literal kinds upstream matches on are strings and numbers alone. Its match arms fall
		// through for every other TSLiteral, so these three are silent despite reading as the
		// same shape.
		{"a boolean literal type", "let a: true = true;"},
		{"a bigint literal type", "let a: 2n = 2n;"},
		{"a negated numeric literal type", "let a: -2 = -2;"},
		// A type and a value of different literal kinds. The kinds must agree, not merely the text.
		// A definite assignment carrying an initializer. TypeScript rejects the combination, and
		// upstream is silent on it: measured against the release binary with a control firing on a
		// neighbouring file. This is also the input that would have made the annotation delete
		// swallow the `!`, since the token before the type is an exclamation rather than a colon.
		{"a definite assignment with an initializer", "let a!: 'bar' = 'bar';"},
		{"a definite assignment property with an initializer", "class C { p!: 'baz' = 'baz'; }"},
		{"a numeric type against a string value", "let a: 2 = '2';"},
		{"a string type against a numeric value", "let a: '2' = 2;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, PreferAsConst, preferAsConstFile, testCase.sourceText))
		})
	}
}

// TestPreferAsConstMatchesNumericValuesNotSpellings pins the comparison oxc actually makes.
//
// oxc compares the parsed f64 values, so `let a: 0x10 = 16` reports. typescript-eslint compares
// `raw` text and declines all four of these. The imported corpus writes only inputs where the two
// agree, so no upstream fixture separates them, and a port that string-compared the source
// spelling would pass the whole corpus while diverging on every line below.
//
// Our NumericLiteral.Text already holds the canonical rendering of the double, so string equality
// on it reproduces the float comparison without parsing anything.
func TestPreferAsConstMatchesNumericValuesNotSpellings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a hexadecimal type against a decimal value", "let a: 0x10 = 16;"},
		{"a trailing zero", "let a: 2 = 2.0;"},
		{"a numeric separator", "let a: 1_0 = 10;"},
		{"exponent notation", "let a: 1e2 = 100;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, PreferAsConst, preferAsConstFile, testCase.sourceText),
				"preferAsConst")
		})
	}
}

// TestPreferAsConstReportsOnceWhenBothArmsCouldApply pins a case that reads like a double report
// and is not.
//
// `let foo: 'bar' = 'bar' as 'bar';` has a literal annotation and a literal assertion, so both the
// declarator arm and the as-expression arm look eligible. Only the second fires: the declarator
// arm compares its type against the initializer, and the initializer is an AsExpression rather
// than a literal, so the kinds disagree and it declines. Upstream's snapshot records exactly one
// diagnostic, at the assertion's type.
//
// This matters beyond the count. The two arms being mutually exclusive on one declarator is what
// makes the annotation arm's two-part edit safe to ship as a plain fix: the pair can never
// compete with this rule's own other repair for the same bytes.
func TestPreferAsConstReportsOnceWhenBothArmsCouldApply(t *testing.T) {
	t.Parallel()

	sourceText := "let foo: 'bar' = 'bar' as 'bar';"
	result := rule_testing.Run(t, PreferAsConst, preferAsConstFile, sourceText)
	rule_testing.ExpectFindings(t, result, "preferAsConst")
	finding := result.Diagnostics[0]
	if got := sourceText[finding.Range.Pos():finding.Range.End()]; got != "'bar'" {
		t.Fatalf("finding points at %q, want the assertion type", got)
	}
	if finding.Range.Pos() != 26 {
		t.Fatalf("finding starts at %d, want 26, the assertion type rather than the annotation", finding.Range.Pos())
	}
}
