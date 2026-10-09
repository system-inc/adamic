package core

import (
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noUselessConcatFile is where the fixtures pretend to live.
const noUselessConcatFile = "/repository/source/NoUselessConcat.ts"

// The corpus is ESLint's own, at `tests/lib/rules/no-useless-concat.js`, copied rather than
// rewritten: all 12 clean cases and all 8 reporting ones, which is the whole upstream file. The
// rule declares `schema: []`, so no case carries options.
//
// Every reporting case states its own line and column per finding, and those were converted to
// byte offsets here and checked to land on a `+` character before being written down. That
// conversion is asserted rather than assumed, because a span fixture derived from the wrong
// arithmetic would pin the wrong answer just as confidently as the right one.
//
// Two cases carry TWO findings, which is why the expectation is a list rather than a count: a
// three-literal chain reports at both operators, and so does the parenthesized pair.
func TestNoUselessConcatFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name              string
		sourceText        string
		wantOperatorStart []int
	}{
		{"two strings", "'a' + 'b'", []int{4}},
		{"a chain whose first operand is on another line", "'a' +\n'b' + 'c'", []int{10}},
		{"an identifier then two strings", "foo + 'a' + 'b'", []int{10}},
		{"three strings", "'a' + 'b' + 'c'", []int{4, 10}},
		{"two parenthesized groups", "(foo + 'a') + ('b' + 'c')", []int{12, 19}},
		{"a template and a string", "`a` + 'b'", []int{4}},
		{"two templates", "`a` + `b`", []int{4}},
		{"an identifier then two templates", "foo + `a` + `b`", []int{10}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConcat, noUselessConcatFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantOperatorStart))
			for index := range wantIds {
				wantIds[index] = "unexpectedConcat"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// The span is the whole rule's other half. Upstream points at the `+` OPERATOR rather
			// than at the expression, so a port reporting the node passes every id assertion while
			// pointing at column 1 of a chain.
			//
			// The spans are compared as a SORTED set rather than in emission order, and that is a
			// fact about the harness rather than a weakening of the assertion. A chain like
			// `'a' + 'b' + 'c'` parses with the second `+` as the outer node, the walk here is
			// pre-order, so the rule emits the outer operator first while upstream lists column 5
			// before column 11. Nothing is lost by sorting: `internal/report/report.go:51` sorts
			// every diagnostic by file and then by position before a reader or a differential ever
			// sees it, so the order this test would be pinning is one that production does not
			// preserve. What matters, and what is asserted, is that the exact set of operators is
			// reported and that each finding is one character wide.
			gotStarts := make([]int, 0, len(result.Diagnostics))
			for _, diagnostic := range result.Diagnostics {
				reported := diagnostic.Range
				if reported.End() != reported.Pos()+1 {
					t.Fatalf("expected a single-character span, got [%d,%d) which is %q",
						reported.Pos(), reported.End(), testCase.sourceText[reported.Pos():reported.End()])
				}
				if got := testCase.sourceText[reported.Pos():reported.End()]; got != "+" {
					t.Fatalf("a finding should point at the operator, got %q", got)
				}
				gotStarts = append(gotStarts, reported.Pos())
			}
			sort.Ints(gotStarts)

			wantStarts := append([]int(nil), testCase.wantOperatorStart...)
			sort.Ints(wantStarts)
			if len(gotStarts) != len(wantStarts) {
				t.Fatalf("expected operators at %v, got %v", wantStarts, gotStarts)
			}
			for index := range wantStarts {
				if gotStarts[index] != wantStarts[index] {
					t.Fatalf("expected operators at %v, got %v", wantStarts, gotStarts)
				}
			}
		})
	}
}

func TestNoUselessConcatStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"two numbers", "var a = 1 + 1;"},
		{"a number times a string", "var a = 1 * '2';"},
		{"subtraction", "var a = 1 - 2;"},
		{"two identifiers", "var a = foo + bar;"},
		{"a string and an identifier", "var a = 'foo' + bar;"},
		{"a concatenation broken across lines", "var foo = 'foo' +\n 'bar';"},
		{"a parenthesized sum beside a string", "var string = (number + 1) + 'px';"},
		{"a string and a number", "'a' + 1"},
		{"a number and a string", "1 + '1'"},
		{"a number and a template", "1 + `1`"},
		{"a template and a number", "`1` + 1"},
		{"a parenthesized sum with a unary beside a template", "(1 + +2) + `b`"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConcat, noUselessConcatFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUselessConcatSkipsParentheses covers the difference between the two parsers, and it is the
// difference most likely to make this port silently narrower than upstream.
//
// ESTree gives a parenthesized expression no node at all, so upstream's `node.left` is already the
// literal inside the parentheses and it never has to skip anything. typescript-go produces a real
// node, so without a skip every parenthesized operand reads as a non-literal.
//
// The brief warns against guessing at parenthesis behaviour, so none of this was guessed: upstream's
// own corpus asserts two findings on `(foo + 'a') + ('b' + 'c')`, and the three simpler shapes below
// were measured against the installed build at 10.8.1, where all three report.
func TestNoUselessConcatSkipsParentheses(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
	}{
		{"a parenthesized left operand", "('a') + 'b'"},
		{"a parenthesized right operand", "'a' + ('b')"},
		{"both operands parenthesized", "('a') + ('b')"},
		{"deeply nested parentheses", "((('a'))) + 'b'"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConcat, noUselessConcatFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpectedConcat")
		})
	}
}

// TestNoUselessConcatTemplatesWithSubstitutions pins a piece of upstream behaviour that reads like a
// defect and is not.
//
// `astUtils.isStringLiteral` accepts ANY template literal, so a template carrying a substitution
// counts even though it is not a constant and the concatenation is not actually useless. Measured
// against the installed build rather than inferred from the helper's name. Reproduced deliberately;
// narrowing it to constant templates would be a silent improvement on upstream, which is the thing
// a port must not do quietly.
func TestNoUselessConcatTemplatesWithSubstitutions(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
	}{
		{"a substituted template and a string", "`a${x}` + 'b'"},
		{"two substituted templates", "`a${x}` + `${y}b`"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConcat, noUselessConcatFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unexpectedConcat")
		})
	}
}

// TestNoUselessConcatSameLine states the line rule, including the shapes that separate the two ways
// of asking it.
//
// Upstream compares the left operand's END line against the right operand's START line. A template
// literal spanning lines is what makes that different from "are the two operands on one line": a
// template opening on line 1 and closing on line 2 leaves the operator on line 2 beside a literal
// that starts on line 1, and upstream reports it. Every verdict below was measured against the
// installed build at 10.8.1.
func TestNoUselessConcatSameLine(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		wantReport bool
	}{
		{"a line break before the right operand", "'a' +\n'b'", false},
		{"a line break before the operator", "'a'\n+ 'b'", false},
		{"a block comment holding a newline before the operator", "'a' /* \n */ + 'b'", false},
		{"a block comment holding a newline after the operator", "'a' + /* \n */ 'b'", false},
		{"a multiline template on the left, which ENDS on the operator's line", "`a\nb` + 'c'", true},
		{"a multiline template on the right, which STARTS on the operator's line", "'c' + `a\nb`", true},
		{"multiline templates on both sides", "`a\nb` + `c\nd`", true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConcat, noUselessConcatFile, testCase.sourceText)
			if testCase.wantReport {
				rule_testing.ExpectFindings(t, result, "unexpectedConcat")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUselessConcatMessage asserts the message against a literal typed here, on a plain pair and
// on a template carrying a substitution.
//
// The second row is why this is not equality on the rule's own constant. The message used to say
// the join produces "a constant the source could have spelled directly", which is false when one
// side is a template with an expression: `${orgFlag}` + ' --dir' joins into a template, one literal
// but no constant. Two of the four ahra findings were that shape (`PlanetScaleApi.ts:627`,
// `GraphQlOperationsMetadataPlugin.ts:476`), and comparing against the constant could never notice.
func TestNoUselessConcatMessage(t *testing.T) {
	t.Parallel()

	const want = "These two literals are joined at runtime where the source could have written one " +
		"literal. Nothing is computed between them, so the concatenation only adds an operator for a " +
		"reader to follow and a chance for the two halves to drift apart."

	for _, source := range []string{"'a' + 'b'", "`--org ${orgFlag}` + ' --dir'"} {
		result := rule_testing.Run(t, NoUselessConcat, noUselessConcatFile, source)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("expected one finding for %s, got %d", source, len(result.Diagnostics))
		}
		if got := result.Diagnostics[0].Message.Id; got != "unexpectedConcat" {
			t.Errorf("expected id %q, got %q", "unexpectedConcat", got)
		}
		if got := result.Diagnostics[0].Message.Description; got != want {
			t.Errorf("description for %s is\n%q\nwant\n%q", source, got, want)
		}
	}
}

// TestNoUselessConcatOtherOperators covers the operator guard.
//
// This listener sees every binary expression, and several non-`+` operators take two operands that
// would both pass the literal test. Upstream's corpus carries subtraction and multiplication; the
// comparison forms were added here because they are the shapes most common in real code.
func TestNoUselessConcatOtherOperators(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"'a' - 'b'",
		"'a' * 'b'",
		"'a' === 'b'",
		"'a' == 'b'",
		"'a' < 'b'",
		"'a' , 'b'",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessConcat, noUselessConcatFile, sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}
