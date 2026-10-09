package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noImplicitCoercionFile is where the fixtures pretend to live.
const noImplicitCoercionFile = "/repository/source/NoImplicitCoercion.ts"

// decodedNoImplicitCoercionOptions routes a fixture's options through the rule's own decoder.
//
// Three of the four booleans default to TRUE, so a struct built by hand in a fixture would carry
// non-nil pointers and every case would pass while the live config's bare "error" still handed the
// rule a nil it read as three judgments switched off. An empty string here is that bare
// configuration.
func decodedNoImplicitCoercionOptions(t *testing.T, raw string) any {
	t.Helper()
	if raw == "" {
		return nil
	}
	options, err := DecodeNoImplicitCoercionOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return options
}

// The corpus is ESLint's own, at `tests/lib/rules/no-implicit-coercion.js`, and it was not
// transcribed. Upstream's tester file was loaded with a stub RuleTester that captured the case
// objects, so every string is the cooked value upstream's own tester would use. The extraction was
// then replayed against the installed eslint at 10.8.1: all 141 cases agree on verdict, on message
// id, on whether a fix is applicable, and on the fix text.
//
// One case is omitted and named here rather than dropped silently: upstream writes `!!(foo + bar)`
// TWICE with opposite verdicts, separated only by `languageOptions.globals.Boolean = "off"`. The
// harness has no way to unset a global, so the second is inexpressible. It is the same judgment the
// shadowing cases below cover from the other side, and the rule's own decision is pinned by
// TestNoImplicitCoercionWithholdsTheFixWhenBooleanIsShadowed.
//
// Of the 47 imported reporting cases, 4 carry an applicable fix and 39 offer a suggestion instead.
// Asserting only the ids would leave both surfaces unchecked, so this asserts the repair on the
// fixable cases, the suggested rewrite on the rest, and the rendered message text on all of them.
func TestNoImplicitCoercionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		options          string
		wantIds          []string
		wantRecommends   []string
		wantFixed        string
		fixable          bool
		wantSuggested    string
		offersSuggestion bool
	}{
		{"!!foo", "", []string{"implicitCoercion"}, []string{"Boolean(foo)"}, "Boolean(foo)", true, "", false},
		{"!!(a, b)", "", []string{"implicitCoercion"}, []string{"Boolean((a, b))"}, "Boolean((a, b))", true, "", false},
		{"(a, b) - 0", "", []string{"implicitCoercion"}, []string{"Number((a, b))"}, "", false, "Number((a, b))", true},
		{"+(a, b)", "", []string{"implicitCoercion"}, []string{"Number((a, b))"}, "", false, "Number((a, b))", true},
		{"-(-(a, b))", "", []string{"implicitCoercion"}, []string{"Number((a, b))"}, "", false, "Number((a, b))", true},
		{"(a, b) * 1", "", []string{"implicitCoercion"}, []string{"Number((a, b))"}, "", false, "Number((a, b))", true},
		{"(a, b) + \"\"", "", []string{"implicitCoercion"}, []string{"String((a, b))"}, "", false, "String((a, b))", true},
		{"`${(a, b)}`", "{\"disallowTemplateShorthand\": true}", []string{"implicitCoercion"}, []string{"String((a, b))"}, "", false, "String((a, b))", true},
		{"!!(foo + bar)", "", []string{"implicitCoercion"}, []string{"Boolean(foo + bar)"}, "Boolean(foo + bar)", true, "", false},
		{"!!(foo + bar); var Boolean = null", "", []string{"implicitCoercion"}, []string{"Boolean(foo + bar)"}, "", false, "Boolean(foo + bar); var Boolean = null", true},
		{"~foo.indexOf(1)", "", []string{"implicitCoercion"}, []string{"foo.indexOf(1) !== -1"}, "", false, "", false},
		{"~foo.bar.indexOf(2)", "", []string{"implicitCoercion"}, []string{"foo.bar.indexOf(2) !== -1"}, "", false, "", false},
		{"+foo", "", []string{"implicitCoercion"}, []string{"Number(foo)"}, "", false, "Number(foo)", true},
		{"-(-foo)", "", []string{"implicitCoercion"}, []string{"Number(foo)"}, "", false, "Number(foo)", true},
		{"+foo.bar", "", []string{"implicitCoercion"}, []string{"Number(foo.bar)"}, "", false, "Number(foo.bar)", true},
		{"1*foo", "", []string{"implicitCoercion"}, []string{"Number(foo)"}, "", false, "Number(foo)", true},
		{"foo*1", "", []string{"implicitCoercion"}, []string{"Number(foo)"}, "", false, "Number(foo)", true},
		{"1*foo.bar", "", []string{"implicitCoercion"}, []string{"Number(foo.bar)"}, "", false, "Number(foo.bar)", true},
		{"foo.bar-0", "", []string{"implicitCoercion"}, []string{"Number(foo.bar)"}, "", false, "Number(foo.bar)", true},
		{"\"\"+foo", "", []string{"implicitCoercion"}, []string{"String(foo)"}, "", false, "String(foo)", true},
		{"``+foo", "", []string{"implicitCoercion"}, []string{"String(foo)"}, "", false, "String(foo)", true},
		{"foo+\"\"", "", []string{"implicitCoercion"}, []string{"String(foo)"}, "", false, "String(foo)", true},
		{"foo+``", "", []string{"implicitCoercion"}, []string{"String(foo)"}, "", false, "String(foo)", true},
		{"\"\"+foo.bar", "", []string{"implicitCoercion"}, []string{"String(foo.bar)"}, "", false, "String(foo.bar)", true},
		{"``+foo.bar", "", []string{"implicitCoercion"}, []string{"String(foo.bar)"}, "", false, "String(foo.bar)", true},
		{"foo.bar+\"\"", "", []string{"implicitCoercion"}, []string{"String(foo.bar)"}, "", false, "String(foo.bar)", true},
		{"foo.bar+``", "", []string{"implicitCoercion"}, []string{"String(foo.bar)"}, "", false, "String(foo.bar)", true},
		{"`${foo}`", "{\"disallowTemplateShorthand\": true}", []string{"implicitCoercion"}, []string{"String(foo)"}, "", false, "String(foo)", true},
		{"`\\\n${foo}`", "{\"disallowTemplateShorthand\": true}", []string{"implicitCoercion"}, []string{"String(foo)"}, "", false, "String(foo)", true},
		{"`${foo}\\\n`", "{\"disallowTemplateShorthand\": true}", []string{"implicitCoercion"}, []string{"String(foo)"}, "", false, "String(foo)", true},
		{"foo += \"\"", "", []string{"implicitCoercion"}, []string{"foo = String(foo)"}, "", false, "foo = String(foo)", true},
		{"foo += ``", "", []string{"implicitCoercion"}, []string{"foo = String(foo)"}, "", false, "foo = String(foo)", true},
		{"var a = !!foo", "{\"boolean\": true, \"allow\": [\"~\"]}", []string{"implicitCoercion"}, []string{"Boolean(foo)"}, "var a = Boolean(foo)", true, "", false},
		{"var a = ~foo.indexOf(1)", "{\"boolean\": true, \"allow\": [\"!!\"]}", []string{"implicitCoercion"}, []string{"foo.indexOf(1) !== -1"}, "", false, "", false},
		{"var a = 1 * foo", "{\"boolean\": true, \"allow\": [\"+\"]}", []string{"implicitCoercion"}, []string{"Number(foo)"}, "", false, "var a = Number(foo)", true},
		{"var a = +foo", "{\"boolean\": true, \"allow\": [\"*\"]}", []string{"implicitCoercion"}, []string{"Number(foo)"}, "", false, "var a = Number(foo)", true},
		{"var a = \"\" + foo", "{\"boolean\": true, \"allow\": [\"*\"]}", []string{"implicitCoercion"}, []string{"String(foo)"}, "", false, "var a = String(foo)", true},
		{"var a = `` + foo", "{\"boolean\": true, \"allow\": [\"*\"]}", []string{"implicitCoercion"}, []string{"String(foo)"}, "", false, "var a = String(foo)", true},
		{"typeof+foo", "", []string{"implicitCoercion"}, []string{"Number(foo)"}, "", false, "typeof Number(foo)", true},
		{"typeof +foo", "", []string{"implicitCoercion"}, []string{"Number(foo)"}, "", false, "typeof Number(foo)", true},
		{"let x ='' + 1n;", "", []string{"implicitCoercion"}, []string{"String(1n)"}, "", false, "let x =String(1n);", true},
		{"~foo?.indexOf(1)", "", []string{"implicitCoercion"}, []string{"foo?.indexOf(1) >= 0"}, "", false, "", false},
		{"~(foo?.indexOf)(1)", "", []string{"implicitCoercion"}, []string{"(foo?.indexOf)(1) !== -1"}, "", false, "", false},
		{"1 * a / 2", "", []string{"implicitCoercion"}, []string{"Number(a)"}, "", false, "Number(a) / 2", true},
		{"(a * 1) / 2", "", []string{"implicitCoercion"}, []string{"Number(a)"}, "", false, "(Number(a)) / 2", true},
		{"a * 1 / (b * 1)", "", []string{"implicitCoercion"}, []string{"Number(b)"}, "", false, "a * 1 / (Number(b))", true},
		{"a * 1 + 2", "", []string{"implicitCoercion"}, []string{"Number(a)"}, "", false, "Number(a) + 2", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name+" ["+testCase.options+"]", func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				testCase.name, decodedNoImplicitCoercionOptions(t, testCase.options))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			for index, diagnostic := range result.Diagnostics {
				want := noImplicitCoercionMessage(testCase.wantRecommends[index]).Description
				if got := diagnostic.Message.Description; got != want {
					t.Fatalf("finding %d message:\n got %q\nwant %q", index, got, want)
				}
			}

			fixCount := 0
			for _, diagnostic := range result.Diagnostics {
				fixCount += len(diagnostic.Fixes)
			}

			if testCase.fixable {
				if fixCount == 0 {
					t.Fatalf("upstream fixes this case and the rule proposed no repair")
				}
				// The typed harness writes each fixture as `TrimSpace(source)+"\n"`
				// (`internal/rule_testing/program.go:160`), and `ExpectFixedSource` compares the
				// WHOLE rewritten file, so upstream's `output` has to be transformed the same way
				// the input was. Padding the rule to make the comparison line up would be the
				// wrong repair.
				rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantFixed)+"\n")
				return
			}

			if fixCount != 0 {
				t.Fatalf("upstream declines to fix this case, but the rule proposed %d fix(es)", fixCount)
			}

			// The suggestion surface. `ExpectFixedSource` covers fixes only, so a suggestion's
			// rewrite has to be applied by hand or nothing checks what it would write.
			suggestions := result.Diagnostics[0].Suggestions
			if !testCase.offersSuggestion {
				if len(suggestions) != 0 {
					t.Fatalf("upstream offers no suggestion here, got %d", len(suggestions))
				}
				return
			}
			if len(suggestions) != 1 {
				t.Fatalf("expected one suggestion, got %d", len(suggestions))
			}
			// Same transform for the suggestion surface: the fixes are computed against the file
			// the harness wrote, not against the Go literal.
			applied := noImplicitCoercionApplyFixes(strings.TrimSpace(testCase.name)+"\n", suggestions[0].Fixes)
			if applied != strings.TrimSpace(testCase.wantSuggested)+"\n" {
				t.Fatalf("the suggested rewrite:\n got %q\nwant %q", applied, strings.TrimSpace(testCase.wantSuggested)+"\n")
			}
			wantSuggestionMessage := noImplicitCoercionSuggestionMessage(testCase.wantRecommends[0]).Description
			if got := suggestions[0].Message.Description; got != wantSuggestionMessage {
				t.Fatalf("suggestion message:\n got %q\nwant %q", got, wantSuggestionMessage)
			}
		})
	}
}

// noImplicitCoercionApplyFixes replays a suggestion's edits back to front.
func noImplicitCoercionApplyFixes(source string, fixes []rule.Fix) string {
	for index := len(fixes) - 1; index >= 0; index-- {
		fix := fixes[index]
		source = source[:fix.Range.Pos()] + fix.Text + source[fix.Range.End():]
	}
	return source
}

func TestNoImplicitCoercionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		options string
	}{
		{"Boolean(foo)", ""},
		{"foo.indexOf(1) !== -1", ""},
		{"Number(foo)", ""},
		{"parseInt(foo)", ""},
		{"parseFloat(foo)", ""},
		{"String(foo)", ""},
		{"!foo", ""},
		{"~foo", ""},
		{"-foo", ""},
		{"+1234", ""},
		{"-1234", ""},
		{"- -1234", ""},
		{"+Number(lol)", ""},
		{"-parseFloat(lol)", ""},
		{"2 * foo", ""},
		{"1 * 1234", ""},
		{"123 - 0", ""},
		{"1 * Number(foo)", ""},
		{"1 * parseInt(foo)", ""},
		{"1 * parseFloat(foo)", ""},
		{"Number(foo) * 1", ""},
		{"Number(foo) - 0", ""},
		{"parseInt(foo) * 1", ""},
		{"parseFloat(foo) * 1", ""},
		{"- -Number(foo)", ""},
		{"1 * 1234 * 678 * Number(foo)", ""},
		{"1 * 1234 * 678 * parseInt(foo)", ""},
		{"(1 - 0) * parseInt(foo)", ""},
		{"1234 * 1 * 678 * Number(foo)", ""},
		{"1234 * 1 * Number(foo) * Number(bar)", ""},
		{"1234 * 1 * Number(foo) * parseInt(bar)", ""},
		{"1234 * 1 * Number(foo) * parseFloat(bar)", ""},
		{"1234 * 1 * parseInt(foo) * parseFloat(bar)", ""},
		{"1234 * 1 * parseInt(foo) * Number(bar)", ""},
		{"1234 * 1 * parseFloat(foo) * Number(bar)", ""},
		{"1234 * Number(foo) * 1 * Number(bar)", ""},
		{"1234 * parseInt(foo) * 1 * Number(bar)", ""},
		{"1234 * parseFloat(foo) * 1 * parseInt(bar)", ""},
		{"1234 * parseFloat(foo) * 1 * Number(bar)", ""},
		{"(- -1234) * (parseFloat(foo) - 0) * (Number(bar) - 0)", ""},
		{"1234*foo*1", ""},
		{"1234*1*foo", ""},
		{"1234*bar*1*foo", ""},
		{"1234*1*foo*bar", ""},
		{"1234*1*foo*Number(bar)", ""},
		{"1234*1*Number(foo)*bar", ""},
		{"1234*1*parseInt(foo)*bar", ""},
		{"0 + foo", ""},
		{"~foo.bar()", ""},
		{"foo + 'bar'", ""},
		{"foo + `${bar}`", ""},
		{"!!foo", "{\"boolean\": false}"},
		{"~foo.indexOf(1)", "{\"boolean\": false}"},
		{"+foo", "{\"number\": false}"},
		{"-(-foo)", "{\"number\": false}"},
		{"foo - 0", "{\"number\": false}"},
		{"1*foo", "{\"number\": false}"},
		{"\"\"+foo", "{\"string\": false}"},
		{"foo += \"\"", "{\"string\": false}"},
		{"var a = !!foo", "{\"boolean\": true, \"allow\": [\"!!\"]}"},
		{"var a = ~foo.indexOf(1)", "{\"boolean\": true, \"allow\": [\"~\"]}"},
		{"var a = ~foo", "{\"boolean\": true}"},
		{"var a = 1 * foo", "{\"boolean\": true, \"allow\": [\"*\"]}"},
		{"- -foo", "{\"number\": true, \"allow\": [\"- -\"]}"},
		{"foo - 0", "{\"number\": true, \"allow\": [\"-\"]}"},
		{"var a = +foo", "{\"boolean\": true, \"allow\": [\"+\"]}"},
		{"var a = \"\" + foo", "{\"boolean\": true, \"string\": true, \"allow\": [\"+\"]}"},
		{"'' + 'foo'", ""},
		{"`` + 'foo'", ""},
		{"'' + `${foo}`", ""},
		{"'foo' + ''", ""},
		{"'foo' + ``", ""},
		{"`${foo}` + ''", ""},
		{"foo += 'bar'", ""},
		{"foo += `${bar}`", ""},
		{"`a${foo}`", "{\"disallowTemplateShorthand\": true}"},
		{"`${foo}b`", "{\"disallowTemplateShorthand\": true}"},
		{"`${foo}${bar}`", "{\"disallowTemplateShorthand\": true}"},
		{"tag`${foo}`", "{\"disallowTemplateShorthand\": true}"},
		{"`${foo}`", ""},
		{"`${foo}`", "{}"},
		{"`${foo}`", "{\"disallowTemplateShorthand\": false}"},
		{"+42", ""},
		{"'' + String(foo)", ""},
		{"String(foo) + ''", ""},
		{"`` + String(foo)", ""},
		{"String(foo) + ``", ""},
		{"`${'foo'}`", "{\"disallowTemplateShorthand\": true}"},
		{"`${`foo`}`", "{\"disallowTemplateShorthand\": true}"},
		{"`${String(foo)}`", "{\"disallowTemplateShorthand\": true}"},
		{"console.log(Math.PI * 1/4)", ""},
		{"a * 1 / 2", ""},
		{"a * 1 / b", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name+" ["+testCase.options+"]", func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				testCase.name, decodedNoImplicitCoercionOptions(t, testCase.options))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoImplicitCoercionWithholdsTheFixWhenBooleanIsShadowed pins the judgment that gates the only
// applicable fix this rule has.
//
// Upstream asks its scope manager whether `Boolean` has a declaration in source. Its corpus covers
// this from two directions and only one of them is expressible here: `!!(foo + bar); var Boolean =
// null` is imported above, while the same source under `languageOptions.globals.Boolean = "off"` is
// not, because the harness has no way to unset a global. So the shapes are enumerated here instead,
// and each was measured against the installed build at 10.8.1 before being written down.
//
// The last row is the one that separates a scope-aware answer from a textual one: a shadow in a
// SIBLING function does not reach the expression, so the fix stands.
func TestNoImplicitCoercionWithholdsTheFixWhenBooleanIsShadowed(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		wantFix    bool
	}{
		{"no shadow", "var b = !!foo;", true},
		{"a var shadow", "var Boolean; var b = !!foo;", false},
		{"a let shadow", "let Boolean = 1; var b = !!foo;", false},
		{"a parameter shadow", "function f(Boolean: any) { return !!foo; }", false},
		{"a function shadow", "function Boolean() {} var b = !!foo;", false},
		{"a class shadow", "class Boolean {} var b = !!foo;", false},
		{"a shadow in a sibling scope only", "function g(Boolean: any) {}\nvar b = !!foo;", true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, "implicitCoercion")

			fixes := len(result.Diagnostics[0].Fixes)
			suggestions := len(result.Diagnostics[0].Suggestions)
			if testCase.wantFix {
				if fixes == 0 {
					t.Fatalf("expected an applicable fix, got %d fixes and %d suggestions", fixes, suggestions)
				}
				return
			}
			if fixes != 0 {
				t.Fatalf("the fix must be withheld when Boolean is shadowed, got %d fixes", fixes)
			}
			if suggestions != 1 {
				t.Fatalf("expected the repair to be offered as a suggestion instead, got %d", suggestions)
			}
		})
	}
}

// TestNoImplicitCoercionParenthesesShapes covers the node our parser keeps and upstream's folds.
//
// Upstream's corpus writes `!!(foo + bar)` and `-(-foo)`, and both went wrong here in the same way
// for the same reason: the operand is a `KindParenthesizedExpression`, which ESTree has no node for.
// Without unwrapping, the shape tests compared the wrong kind and the recommendation carried the
// parentheses twice, as `Boolean((foo + bar))`.
//
// The last two rows are the sharp end and are not in upstream's corpus as a pair: parenthesizing a
// CALLEE ends an optional chain, so the recommendation moves from `>= 0` back to `!== -1`. Measured
// against the installed build, which recommends exactly that for each.
func TestNoImplicitCoercionParenthesesShapes(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name           string
		sourceText     string
		wantRecommends string
	}{
		{"a parenthesized boolean operand", "!!(foo + bar)", "Boolean(foo + bar)"},
		{"a doubly parenthesized operand", "!!((foo))", "Boolean(foo)"},
		{"a parenthesized double negation", "-(-foo)", "Number(foo)"},
		{"a chained indexOf", "~foo?.indexOf(1)", "foo?.indexOf(1) >= 0"},
		{"a chained indexOf whose callee is parenthesized", "~(foo?.indexOf)(1)", "(foo?.indexOf)(1) !== -1"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, "implicitCoercion")
			want := noImplicitCoercionMessage(testCase.wantRecommends).Description
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Fatalf("message:\n got %q\nwant %q", got, want)
			}
		})
	}
}

// TestNoImplicitCoercionIndexOfShapes covers the membership-test arm's own discriminations, none of
// which upstream's corpus writes.
//
// Two mutations survived the entire imported suite before these existed: accepting a computed key
// that is not a string literal, and dropping `lastIndexOf` from the name test. Both are real gaps
// rather than equivalent mutants, and the corpus cannot see them because it writes only
// `~foo.indexOf(bar)`. Every row was measured against the installed build at 10.8.1.
//
// The static-key rule is what separates the third row from the fourth: `foo['indexOf'](x)` names
// the property statically and reports, while `foo[indexOf](x)` computes the name at runtime and
// cannot be recognised.
func TestNoImplicitCoercionIndexOfShapes(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name           string
		sourceText     string
		wantRecommends string
		wantReport     bool
	}{
		{"a plain indexOf", "~foo.indexOf(bar)", "foo.indexOf(bar) !== -1", true},
		{"lastIndexOf counts too", "~foo.lastIndexOf(bar)", "foo.lastIndexOf(bar) !== -1", true},
		{"a static computed key", "~foo['indexOf'](bar)", "foo['indexOf'](bar) !== -1", true},
		{"a runtime computed key is not a name", "~foo[indexOf](bar)", "", false},
		{"another method entirely", "~foo.someOther(bar)", "", false},
		{"the property without a call", "~foo.indexOf", "", false},
		{"a bare operand", "~foo", "", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				testCase.sourceText, nil)
			if !testCase.wantReport {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "implicitCoercion")
			want := noImplicitCoercionMessage(testCase.wantRecommends).Description
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Fatalf("message:\n got %q\nwant %q", got, want)
			}
			// This arm offers neither a fix nor a suggestion: the reader has to choose the
			// comparison, so upstream reports it bare.
			if fixes := len(result.Diagnostics[0].Fixes); fixes != 0 {
				t.Fatalf("the indexOf arm must not carry a fix, got %d", fixes)
			}
			if suggestions := len(result.Diagnostics[0].Suggestions); suggestions != 0 {
				t.Fatalf("the indexOf arm must not carry a suggestion, got %d", suggestions)
			}
		})
	}
}

// TestNoImplicitCoercionMultiplyByOneOperandShapes covers the operand test that upstream writes as
// a node-type comparison and this parser cannot express that way.
//
// Upstream skips an operand that is itself a `BinaryExpression`, so `a * b * 1` names the whole
// product rather than one factor. Two things make that hard to port. ESTree folds parentheses away,
// so `(index + 1) * 1` arrives already unwrapped; and ESTree splits into three node types where
// this parser has one, since `&& || ??` are a `LogicalExpression` and the comma operator is a
// `SequenceExpression`, neither of which upstream's test excludes.
//
// So a kind test alone is wrong in BOTH directions, and each direction cost a real defect. Without
// the unwrap the rule reported `(index + 1) * 1.0` in modules/tasks, which the installed build
// leaves alone; that one was found on the real tree rather than in the corpus, which writes no
// parenthesized operand here. With the unwrap but without the operator test, `(a, b) * 1` went
// silent, and that one IS in the corpus.
//
// Every row was measured against the installed build at 10.8.1.
func TestNoImplicitCoercionMultiplyByOneOperandShapes(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name           string
		sourceText     string
		wantRecommends string
		wantReport     bool
	}{
		{"a parenthesized arithmetic operand is skipped", "(i + 1) * 1", "", false},
		{"the same on the other side", "1 * (i + 1)", "", false},
		{"a parenthesized sequence still reports", "(a, b) * 1", "Number((a, b))", true},
		{"a parenthesized logical expression still reports", "(a || b) * 1", "Number(a || b)", true},
		{"a parenthesized identifier reports", "(a) * 1", "Number(a)", true},
		{"a parenthesized call reports", "(f()) * 1", "Number(f())", true},
		{"the real-tree shape that started this", "const position = (index + 1) * 1.0;", "", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				testCase.sourceText, nil)
			if !testCase.wantReport {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "implicitCoercion")
			want := noImplicitCoercionMessage(testCase.wantRecommends).Description
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Fatalf("message:\n got %q\nwant %q", got, want)
			}
		})
	}
}

// TestNoImplicitCoercionMultilineOperand pins the shape whose RECOMMENDATION spans lines.
//
// Upstream writes every case on one line, so no imported fixture produces a recommendation
// containing a newline. That matters beyond the rule: the message embeds the operand's own source
// text, so a `!!(` whose operand closes several lines later renders a multi-line message, and any
// instrument that reads cohere's output line by line loses the finding.
//
// This was reported as a missing finding at DialogRoot.tsx:175:25 and it was not one. The rule
// reports it, the live run prints it, and two extraction scripts on opposite sides of the
// comparison each dropped it: one because a wrapped continuation carried the rule tag and inflated
// the count, one because the first line did not. The two errors happened to cancel, which is why an
// earlier measurement read as exact agreement at the wrong number.
//
// The fixtures below are the rule's half of that. Nothing here can catch an extraction bug, which
// is the point worth remembering: the defect was never in the rule.
func TestNoImplicitCoercionMultilineOperand(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name           string
		sourceText     string
		wantRecommends string
	}{
		{"an operand closing on a later line", "var x = !!(\n\ta &&\n\tb\n);", "Boolean(a &&\n\tb)"},
		{"a newline after the opening parenthesis only", "var x = !!(\n\ta && b);", "Boolean(a && b)"},
		{"a newline before the closing parenthesis only", "var x = !!(a && b\n);", "Boolean(a && b)"},
		// The tree's shape at DialogRoot.tsx:175, with operands that are not already booleans. Its
		// real operand, `b === undefined && c === undefined && !d`, IS a boolean, so the rule is now
		// silent there (#vsy2eym) and that shape lives in the nothing-is-coerced table instead.
		{"the shape reported from the tree", "var x = a ?? !!(\n\tb &&\n\tc &&\n\td\n);", "Boolean(b &&\n\tc &&\n\td)"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, "implicitCoercion")
			want := noImplicitCoercionMessage(testCase.wantRecommends).Description
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Fatalf("message:\n got %q\nwant %q", got, want)
			}
			// Note which rows carry a newline in the recommendation and which do not: a break
			// INSIDE the operand survives into the message because the text is sliced from source,
			// while a break between the parenthesis and the operand is trivia and does not. Both
			// are asserted by the equality above rather than by a separate predicate, which is
			// what an earlier version of this test got wrong.
		})
	}
}

// TestNoImplicitCoercionTypeScriptShapes covers syntax upstream's corpus structurally cannot carry.
//
// Its tests are JavaScript, so no imported case puts a type assertion, a non-null operator or a
// generic call inside a coerced operand. This rule's repairs are built by SLICING source text
// rather than re-rendering the expression, which is what makes these safe, and that is a property
// worth pinning: the natural way to write the recommendation is to rebuild it from parts, and that
// is how this project lost type information twice.
func TestNoImplicitCoercionTypeScriptShapes(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name           string
		sourceText     string
		wantRecommends string
	}{
		{"a non-null assertion inside a double negation", "!!foo!", "Boolean(foo!)"},
		{"an as-expression inside a double negation", "!!(foo as unknown)", "Boolean(foo as unknown)"},
		{"a generic call under a unary plus", "+f<number>(1)", "Number(f<number>(1))"},
		{"a satisfies expression added to an empty string", "'' + (foo satisfies string)", "String(foo satisfies string)"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, "implicitCoercion")
			want := noImplicitCoercionMessage(testCase.wantRecommends).Description
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Fatalf("message:\n got %q\nwant %q", got, want)
			}
		})
	}
}

// TestDecodeNoImplicitCoercionOptions is where this rule's decoder is hard.
//
// Three of the four booleans default to TRUE, so an absent option read as a zero value would leave
// three of the four judgments switched off, and every fixture built from a struct rather than routed
// through the decoder would pass anyway.
func TestDecodeNoImplicitCoercionOptions(t *testing.T) {
	t.Parallel()

	settingsFrom := func(t *testing.T, raw string) NoImplicitCoercionOptions {
		t.Helper()
		decoded, err := DecodeNoImplicitCoercionOptions(json.RawMessage(raw))
		if err != nil {
			t.Fatalf("the decoder refused %s: %v", raw, err)
		}
		options, isOptions := decoded.(NoImplicitCoercionOptions)
		if !isOptions {
			t.Fatalf("expected NoImplicitCoercionOptions, got %T", decoded)
		}
		return options
	}

	t.Run("empty input enables the three default-true judgments", func(t *testing.T) {
		t.Parallel()
		options := settingsFrom(t, "")
		for name, value := range map[string]*bool{
			"boolean": options.Boolean, "number": options.Number, "string": options.String,
		} {
			if value == nil || !*value {
				t.Fatalf("%s should default to true, got %v", name, value)
			}
		}
		if options.DisallowTemplateShorthand == nil || *options.DisallowTemplateShorthand {
			t.Fatalf("disallowTemplateShorthand should default to false, got %v", options.DisallowTemplateShorthand)
		}
	})

	t.Run("an explicit false is kept", func(t *testing.T) {
		t.Parallel()
		options := settingsFrom(t, `{"boolean": false}`)
		if options.Boolean == nil || *options.Boolean {
			t.Fatalf("an explicit false must survive, got %v", options.Boolean)
		}
		if options.Number == nil || !*options.Number {
			t.Fatalf("an absent key must keep its default, got %v", options.Number)
		}
	})

	t.Run("nil options reach the rule with the defaults", func(t *testing.T) {
		t.Parallel()
		// Past the decoder rather than through it: what a bare "error" produces after the config
		// layer turns the decoder's error into nil. A zero-valued struct would go silent here.
		result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
			"var b = !!foo;", nil)
		rule_testing.ExpectFindings(t, result, "implicitCoercion")
	})
}

// TestNoImplicitCoercionIsSilentWhenNothingIsCoerced is where cohere is deliberately quieter than
// ESLint.
//
// Upstream is type-blind, so `1 * zoom` reports and recommends `Number(zoom)` even when `zoom` is
// already a `number`. Nothing is coerced there, and the advice is wrong. cohere asks the checker and
// stays silent on the four number arms (`+x`, `-(-x)`, `1 * x`, `x - 0`) when every constituent of the
// operand's type is already a number, and on `-(-x)` when it is a bigint, the one number arm that is an
// identity on a bigint rather than a TypeError.
//
// The first two silent rows are the real sites in ahra, four findings in ESLint:
// `libraries/structure/source/components/maps/Map.tsx:747/765/784` (`context.lineWidth = 1 * zoom`)
// and `MapDrawing.ts:220` (`const dotRadius = 1.0 * zoom`), with `zoom: number`.
//
// The reporting rows keep the proof tight: a string, `number | undefined`, `any`, `unknown`, and a
// bigint under `+`, `* 1` and `- 0` (each a compile error and a runtime TypeError, so still not
// "already a number"), plus the string and boolean arms, which this change does not touch.
func TestNoImplicitCoercionIsSilentWhenNothingIsCoerced(t *testing.T) {
	t.Parallel()

	const declarations = "declare const zoom: number;\n" +
		"declare const context: { lineWidth: number };\n" +
		"declare const level: 1 | 2 | 3;\n" +
		"declare const branded: number & { readonly unit: 'pixels' };\n" +
		"declare const big: bigint;\n" +
		"declare const text: string;\n" +
		"declare const maybe: number | undefined;\n" +
		"declare const loose: any;\n" +
		"declare const opaque: unknown;\n" +
		"declare const flag: boolean;\n" +
		"declare const yes: true | false;\n" +
		"declare const maybeFlag: boolean | null;\n" +
		"declare const maybeText: string | undefined;\n" +
		"declare const label: `item-${number}`;\n" +
		"enum Size { Small = 1, Large = 2 }\n" +
		"declare const size: Size;\n"

	for _, testCase := range []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{name: "Map.tsx:747, one times a number", sourceText: "context.lineWidth = 1 * zoom;"},
		{name: "MapDrawing.ts:220, one-point-zero times a number", sourceText: "const dotRadius = 1.0 * zoom;"},
		{name: "a number times one", sourceText: "const n = zoom * 1;"},
		{name: "unary plus on a number", sourceText: "const n = +zoom;"},
		{name: "double negation on a number", sourceText: "const n = -(-zoom);"},
		{name: "a number minus zero", sourceText: "const n = zoom - 0;"},
		{name: "a union of number literals", sourceText: "const n = 1 * level;"},
		{name: "a branded number", sourceText: "const n = +branded;"},
		{name: "a numeric enum", sourceText: "const n = 1 * size;"},
		{name: "double negation on a bigint", sourceText: "const n = -(-big);"},
		{name: "a constrained type parameter", sourceText: "function f<T extends number>(value: T) { return 1 * value; }"},
		{name: "one times a non-null-asserted number", sourceText: "const n = 1 * maybe!;"},

		{name: "one times a string", sourceText: "const n = 1 * text;", wantIds: []string{"implicitCoercion"}},
		{name: "unary plus on a string", sourceText: "const n = +text;", wantIds: []string{"implicitCoercion"}},
		{name: "double negation on a string", sourceText: "const n = -(-text);", wantIds: []string{"implicitCoercion"}},
		{name: "a string minus zero", sourceText: "const n = text - 0;", wantIds: []string{"implicitCoercion"}},
		{name: "unary plus on a possibly undefined number", sourceText: "const n = +maybe;", wantIds: []string{"implicitCoercion"}},
		{name: "one times any", sourceText: "const n = 1 * loose;", wantIds: []string{"implicitCoercion"}},
		{name: "unary plus on unknown", sourceText: "const n = +opaque;", wantIds: []string{"implicitCoercion"}},
		{name: "unary plus on a bigint", sourceText: "const n = +big;", wantIds: []string{"implicitCoercion"}},
		{name: "one times a bigint", sourceText: "const n = 1 * big;", wantIds: []string{"implicitCoercion"}},
		{name: "a bigint minus zero", sourceText: "const n = big - 0;", wantIds: []string{"implicitCoercion"}},

		// The string and boolean arms ask the same question (#vsy2eym). These four were reporting
		// controls while only the number arms asked it.
		{name: "an empty string plus a string", sourceText: "const s = '' + text;"},
		{name: "a string plus an empty string", sourceText: "const s = text + '';"},
		{name: "a string plus-equals an empty string", sourceText: "let s = text;\ns += '';"},
		{name: "a template literal type plus an empty string", sourceText: "const s = '' + label;"},
		{name: "double not on a boolean", sourceText: "const b = !!flag;"},
		{name: "double not on a boolean literal union", sourceText: "const b = !!yes;"},
		{name: "DialogRoot.tsx:175, double not on a boolean expression", sourceText: "const b = loose ?? !!(\n\ttext === undefined &&\n\tflag === undefined &&\n\t!flag\n);"},

		{name: "an empty string plus a possibly undefined string", sourceText: "const s = '' + maybeText;", wantIds: []string{"implicitCoercion"}},
		{name: "an empty string plus a number", sourceText: "const s = '' + zoom;", wantIds: []string{"implicitCoercion"}},
		{name: "a number plus-equals an empty string", sourceText: "let n: number | string = zoom;\nn += '';", wantIds: []string{"implicitCoercion"}},
		{name: "an empty string plus any", sourceText: "const s = '' + loose;", wantIds: []string{"implicitCoercion"}},
		{name: "double not on a nullable boolean", sourceText: "const b = !!maybeFlag;", wantIds: []string{"implicitCoercion"}},
		{name: "double not on a string", sourceText: "const b = !!text;", wantIds: []string{"implicitCoercion"}},
		{name: "double not on unknown", sourceText: "const b = !!opaque;", wantIds: []string{"implicitCoercion"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				declarations+testCase.sourceText, nil)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}
