package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// templateCurlyInStringFile is where the fixtures pretend to live.
const templateCurlyInStringFile = "/repository/source/TemplateCurlyInString.ts"

// The corpus is ESLint's own, copied rather than rewritten.
//
// Every case below is verbatim from
// `eslint/tests/lib/rules/no-template-curly-in-string.js`: 15 pass and 7 fail, each failing case
// naming exactly one `unexpectedTemplateExpression`, so one finding per input is stated by the
// corpus rather than assumed. The cases were lifted by evaluating that test file with its
// `RuleTester` stubbed and serialising the spec object it was handed, so the strings here are the
// ones upstream runs rather than a retyping of them.
//
// Copied because a fixture a porter invents encodes the same belief as the port, and the case that
// catches a bug is the one nobody would think to write.
func TestNoTemplateCurlyInStringFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a placeholder in a single quoted string", "'Hello, ${name}'"},
		{"a placeholder in a double quoted string", "\"Hello, ${name}\""},
		{"two placeholders, reported once", "'${greeting}, ${name}'"},
		{"a placeholder holding an expression", "'Hello, ${index + 1}'"},
		{"a placeholder holding a concatenation", "'Hello, ${name + \" foo\"}'"},
		{"a placeholder holding a fallback", "'Hello, ${name || \"foo\"}'"},
		{"a placeholder holding an object access", "'Hello, ${{foo: \"bar\"}.foo}'"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoTemplateCurlyInString, templateCurlyInStringFile, testCase.sourceText),
				"unexpectedTemplateExpression")
		})
	}
}

// The clean cases are the whole discrimination, and they split into two groups.
//
// The first five are template literals, where a placeholder is doing its job. They are the reason
// the rule anchors on a string literal kind rather than on anything that holds text; a port
// listening to every literal-shaped node would report all of them. The tagged case
// (`templateFunction`) is there because a tag makes no difference: it is still a template and still
// clean.
//
// The rest are strings that hold some but not all of the placeholder syntax, and each fails the
// pattern a different way. `'$2'` has a dollar with no brace. `'${'` opens and never closes.
// `'$}'` closes without opening. `'{foo}'` has braces without a dollar. Those four are why the
// scan cannot be a search for `$` or for `{`.
func TestNoTemplateCurlyInStringStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a real template literal", "`Hello, ${name}`;"},
		{"a tagged template literal", "templateFunction`Hello, ${name}`;"},
		{"a template with no placeholder", "`Hello, name`;"},
		{"a plain string", "'Hello, name';"},
		{"a concatenation", "'Hello, ' + name;"},
		{"a template holding an expression", "`Hello, ${index + 1}`"},
		{"a template holding a concatenation", "`Hello, ${name + \" foo\"}`"},
		{"a template holding a fallback", "`Hello, ${name || \"foo\"}`"},
		{"a template holding an object access", "`Hello, ${{foo: \"bar\"}.foo}`"},
		{"a dollar and a digit", "'$2'"},
		{"an opening that never closes", "'${'"},
		{"a closing with no opening", "'$}'"},
		{"braces with no dollar", "'{foo}'"},
		{"an object written into a string", "'{foo: \"bar\"}'"},
		{"a declaration with no string at all", "const number = 3"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoTemplateCurlyInString, templateCurlyInStringFile, testCase.sourceText))
		})
	}
}

// The cases below are ours. Each exists because reading our own code or driving the installed
// ESLint build raised a question the imported corpus does not answer.

// TestNoTemplateCurlyInStringReportsTheWholeLiteral asserts where every finding points, which
// ExpectFindings cannot see.
//
// The rule carries no repair, so nothing downstream would notice a finding anchored on the wrong
// node; a rule reporting the enclosing statement or the variable declaration would pass every
// message-id fixture above. Upstream reports the `Literal` node, quotes included, which its own
// corpus records as columns 1 through 17 for a sixteen-byte literal.
//
// The expected text is the whole literal including its quotes, checked by slicing the source with
// the finding's own range, so a span short by one on either side slices to something that is not a
// complete literal.
func TestNoTemplateCurlyInStringReportsTheWholeLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantStart  int
		wantText   string
	}{
		{"'Hello, ${name}'", 0, "'Hello, ${name}'"},
		{"\"Hello, ${name}\"", 0, "\"Hello, ${name}\""},
		{"'${greeting}, ${name}'", 0, "'${greeting}, ${name}'"},
		// Not at offset zero, so a rule reporting the file or the statement would be visible.
		{"var a = 'Hello, ${name}';", 8, "'Hello, ${name}'"},
		{"foo('${a}');", 4, "'${a}'"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTemplateCurlyInString, templateCurlyInStringFile, testCase.sourceText)
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

// TestNoTemplateCurlyInStringReportsOncePerLiteral pins the count, which the imported corpus
// states only for a single input.
//
// Upstream calls `regex.test`, which answers a boolean, so a literal holding three placeholders
// still reports once. A port looping over matches would report three times and every fixture above
// would still pass, because `'${greeting}, ${name}'` is the corpus's only multi-placeholder case
// and it asserts a single id which a two-finding result would fail on count alone. These make the
// property explicit rather than resting on that one case.
//
// The second pair is the same property across separate literals: two strings in one file are two
// findings, which is what distinguishes "once per literal" from "once per file".
func TestNoTemplateCurlyInStringReportsOncePerLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantFindings int
	}{
		{"three placeholders in one literal", "'${a}${b}${c}'", 1},
		{"placeholders separated by text", "'x${a}y${b}z'", 1},
		{"two literals, two findings", "var a = '${x}'; var b = '${y}';", 2},
		{"one literal reporting beside a clean one", "var a = '${x}'; var b = 'plain';", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTemplateCurlyInString, templateCurlyInStringFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantFindings {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), testCase.wantFindings)
			}
		})
	}
}

// TestNoTemplateCurlyInStringMatchesTheCookedValue pins that the subject is the parsed value and
// not the raw source text.
//
// The first case writes its dollar sign as an escape, so the raw text between the quotes is
// `\${a}` and holds no unescaped `${`, while the cooked value is `${a}`. Both halves of that were
// measured rather than assumed: our parser hands back that raw and that cooked pair, and the
// installed ESLint build (version 10.8.1, driven through the Linter API) reports the input, because
// upstream matches `node.value`, which is already cooked. A port scanning `TokenRange` text instead
// would be silent here, and silent on exactly the string that renders with a live-looking
// placeholder in it.
//
// The second case is the same dollar written as a unicode escape, which no reading of the raw text
// could find at all. It reports upstream; measured.
//
// The reverse direction has no expressible case in our tree, and that is worth stating rather than
// leaving as an absence: a raw `${` that cooks away would need an escape that erases characters,
// and the only one is a line continuation, which cannot sit between the dollar and the brace and
// still leave them adjacent in the value.
func TestNoTemplateCurlyInStringMatchesTheCookedValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		// Raw-quoted so the backslash reaches the fixture rather than being read by Go's unquoter.
		sourceText string
	}{
		{"a dollar written as a backslash escape", `'\${a}'`},
		{"a dollar written as a unicode escape", `'\u0024{a}'`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoTemplateCurlyInString, templateCurlyInStringFile, testCase.sourceText),
				"unexpectedTemplateExpression")
		})
	}
}

// TestNoTemplateCurlyInStringPatternEdges covers the three ways the pattern is narrower than a
// reader expects, each measured against the installed build rather than inferred from the regex.
//
// `'${}'` is clean because `[^}]+` demands at least one character; `'${ }'` reports because a space
// is a character. That single space is the entire difference between the two, and a port writing
// `[^}]*` would report both while passing every imported fixture, since the corpus contains
// neither.
//
// `'${{a}}'` reports: the class is negated on the brace rather than lazy, so the match is `${{a}`
// and the trailing brace is left over. `'$ {a}'` is clean because the opening must be adjacent.
//
// The last two are the backtracking cases. An opening that fails to close does not end the search,
// because a later opening may succeed, and a scan that gave up at the first failure would be silent
// on both.
func TestNoTemplateCurlyInStringPatternEdges(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"an empty placeholder body", "'${}'", false},
		{"a placeholder body of one space", "'${ }'", true},
		{"a brace inside the body", "'${{a}}'", true},
		{"a space between the dollar and the brace", "'$ {a}'", false},
		{"an unclosed opening before a closed one", "'${a${b}'", true},
		{"an empty placeholder before a real one", "'${}${a}'", true},
		{"two empty placeholders and nothing else", "'x${}y${}z'", false},
		{"a body holding an escape", `'${a\n b}'`, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTemplateCurlyInString, templateCurlyInStringFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "unexpectedTemplateExpression")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoTemplateCurlyInStringDeclinesOtherLiteralKinds pins the surface.
//
// Upstream anchors on `Literal` and then tests `typeof node.value === "string"`, which declines a
// regular expression even though its value is truthy. Measured: `var a = /\${x}/;` is clean on the
// installed build. Our tree gives regular expressions, templates, and strings three separate kinds,
// so the same discrimination lands in the listener map rather than in a value test, and this is
// what pins that the translation did not widen the surface on the way.
//
// The template cases matter most: a template is where a placeholder belongs, so reporting one would
// be a false positive on correct code rather than a missed finding.
func TestNoTemplateCurlyInStringDeclinesOtherLiteralKinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a regular expression holding the syntax", `var a = /\${x}/;`},
		{"a template with a live placeholder", "var a = `${x}`;"},
		{"a template with no placeholder", "var a = `plain`;"},
		{"a tagged template", "var a = tag`${x}`;"},
		{"a number", "var a = 3;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoTemplateCurlyInString, templateCurlyInStringFile, testCase.sourceText))
		})
	}
}

// TestNoTemplateCurlyInStringMessage asserts the reported text exactly.
//
// A `rule.Message` is `{Id, Description}` with no interpolation, so there is nothing to render
// wrong here the way a format string can be. What this catches instead is the id drifting from what
// upstream calls the finding, which every other assertion in this file would survive: the fixtures
// pass the id they expect straight through, so an id renamed in both places at once stays green.
// The literal below is typed here rather than read from the rule's own constant, so both sides
// cannot move together.
func TestNoTemplateCurlyInStringMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoTemplateCurlyInString, templateCurlyInStringFile, "'${a}'")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "unexpectedTemplateExpression" {
		t.Errorf("message id = %q, want %q", got, "unexpectedTemplateExpression")
	}
	if result.Diagnostics[0].Message.Description == "" {
		t.Error("the message carries no description, so the finding says nothing about why")
	}
}
