package core

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// sequencesFile is where the fixtures pretend to live.
const sequencesFile = "/repository/source/Sequences.ts"

// decodedSequenceOptions routes a fixture's options through the rule's own decoder.
//
// Building a NoSequencesOptions literal instead would leave the decoder untested, and the decoder is
// the one line of this rule with no upstream counterpart: it has to turn an absent key into `true`
// where the generic helper would hand back a zero-value struct reading as `false`. A struct-built
// fixture states its own answer and cannot see that.
func decodedSequenceOptions(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeNoSequencesOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %s: %v", raw, err)
	}
	return decoded
}

// The corpus is ESLint's own, extracted from the tester rather than retyped.
//
// Every case below is verbatim from `eslint/tests/lib/rules/no-sequences.js`: 20 valid and 22
// invalid, each invalid one carrying its own column, which is the whole point of importing them.
// This rule reports at a *token* rather than at the node it matched, so a fixture asserting only the
// message id would pass over a finding pointing anywhere in the expression.
//
// Extracted by evaluating the tester with a stubbed RuleTester and serialising what it was handed,
// so no source string in this file was retyped. The columns are upstream's, converted from
// one-based columns to zero-based byte offsets at the assertion.
func TestNoSequencesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		column     int
	}{
		{"1, 2;", nil, 2},
		{"a = 1, 2", nil, 6},
		{"do {} while (doSomething(), !!test);", nil, 27},
		{"for (; doSomething(), !!test; );", nil, 21},
		{"if (doSomething(), !!test);", nil, 18},
		{"switch (doSomething(), val) {}", nil, 22},
		{"while (doSomething(), !!test);", nil, 21},
		{"with (doSomething(), val) {}", nil, 20},
		{"a => (doSomething(), a)", nil, 20},
		{"(1), 2", nil, 4},
		{"((1)) , (2)", nil, 7},
		{"while((1) , 2);", nil, 11},
		{"var foo = (1, 2);", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 13},
		{"(0,eval)(\"foo()\");", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 3},
		{"foo(a, (b, c), d);", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 10},
		{"do {} while ((doSomething(), !!test));", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 28},
		{"for (; (doSomething(), !!test); );", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 22},
		{"if ((doSomething(), !!test));", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 19},
		{"switch ((doSomething(), val)) {}", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 23},
		{"while ((doSomething(), !!test));", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 22},
		{"with ((doSomething(), val)) {}", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 21},
		{"a => ((doSomething(), a))", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 21},

		// Beyond the corpus. Chained commas are one node upstream and several here, so these pin the
		// count as well as the position: measured against the installed rule, both report exactly
		// once and both at column 2, which is the chain's FIRST comma rather than its outermost.
		{"a, b, c;", nil, 2},
		{"a, b, c, d;", nil, 2},
		{"a, b, c, d, e;", nil, 2},

		// A three-element chain in a position that is NOT exempt, paired with the five silent cases
		// below. These are the inputs that caught the chain-deduplication running in the wrong
		// direction: standing down on the outer node instead of the inner one still reports here,
		// so a fixture in this direction alone cannot see the defect and the silent half is the
		// load-bearing one. Measured against the installed rule.
		{"if (a, b, c);", nil, 6},
		{"for (; a, b, c ;);", nil, 9},
		{"a => (a, b, c)", nil, 8},
		{"function f(){ return a, b, c; }", nil, 23},
		{"`${a, b, c}`", nil, 5},

		// The same shape with a parenthesis in it, which is where the flattening stops. Upstream
		// builds two SequenceExpression nodes here and reports only on the unparenthesized one.
		// Measured: `(a, b), c` reports at 7 and `a, (b, c)` at 2.
		{"(a, b), c;", nil, 7},
		{"a, (b, c);", nil, 2},
		{"((a, b)), c;", nil, 9},

		// The `for` exemption is the initializer and the update, and NOT the condition. The corpus
		// covers the condition; these two pin that the exemption is positional rather than a blanket
		// exemption for the statement, which a structural port could easily get backwards.
		{"for (; a, b ;);", nil, 9},

		// The `for` slot comparison has to see through parentheses, and it has to happen BEFORE the
		// paren allowance. This is the inner half of the pair upstream's two `allowInParentheses:
		// false` clean cases establish: the outer sequence is the initializer and is exempt, the
		// inner one's parent is that sequence rather than the statement, and it reports. Measured
		// against the installed rule at column 8.
		{"for ((a, b), c; ; );", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 8},
		{"for (;; (a, b), c);", decodedSequenceOptions(t, `{"allowInParentheses": false}`), 11},

		// The arrow body needs two parenthesis nodes where every other position needs one, because
		// its parentheses are part of the expression rather than a statement's grammar. The corpus
		// covers both arrow forms with `doSomething()`; this is the minimal pair.
		{"a => (b, c)", nil, 8},

		// A comma inside a template substitution and inside an optional element access. Neither is
		// an argument list, so neither is a separator, and both report upstream.
		{"`${a, b}`", nil, 5},
		{"x?.[a, b]", nil, 6},

		// A sequence on the right of an assignment whose own left is parenthesized. This separates
		// "the sequence is parenthesized" from "something inside it is", which is the discrimination
		// `while((1) , 2)` makes in the corpus and which a paren check reading the wrong node would
		// invert.
		{"a = (b, c) , d", nil, 12},

		// A chain whose LAST element is a parenthesized sequence. The outer chain is unparenthesized
		// and reports at its first comma; the inner one is excused. Measured at column 2.
		{"a, b, (c, d);", nil, 2},

		// Two parenthesized sequences joined by a bare comma. Both inner ones are excused and the
		// comma between them is what reports. Written first as a clean case, on the reading that
		// everything in sight was parenthesized; the comma joining them is not. Measured at 7.
		{"(a, b), (c, d);", nil, 7},

		// A whole chain wrapped in parentheses, then extended. The parenthesized part is excused and
		// is a separate node upstream; what reports is the comma OUTSIDE the parentheses. Written
		// first with the expectation of column 3, on the reasoning that the leftmost comma of the
		// chain is reported -- which is true only when the chain is unbroken. Measured at 12, and
		// the rule was right where the reasoning was not.
		{"((a, b), c), d;", nil, 12},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoSequences, sequencesFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, "unexpectedCommaExpression")

			// Upstream's columns are one-based; the diagnostic carries a zero-based offset. Every
			// fixture here is one line, so the column is the offset plus one.
			wantOffset := testCase.column - 1
			gotRange := result.Diagnostics[0].Range
			if gotRange.Pos() != wantOffset {
				t.Errorf("reported at offset %d, want %d (upstream column %d)",
					gotRange.Pos(), wantOffset, testCase.column)
			}

			// The span is the comma token itself, one byte wide, and slicing the source is the only
			// assertion that can see a range that is the right width in the wrong place.
			reported := testCase.sourceText[gotRange.Pos():gotRange.End()]
			if reported != "," {
				t.Errorf("reported text %q, want a comma", reported)
			}
		})
	}
}

// The clean cases are what separate the comma OPERATOR from the four separators spelled the same
// way: array elements, object properties, argument lists, and the declarators of one `var`. Each is
// a class of false positive a rule matching on the token would ship.
func TestNoSequencesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
	}{
		{"var arr = [1, 2];", nil},
		{"var obj = {a: 1, b: 2};", nil},
		{"var a = 1, b = 2;", nil},
		{"var foo = (1, 2);", nil},
		{"(0,eval)(\"foo()\");", nil},
		{"for (i = 1, j = 2;; i++, j++);", nil},
		{"foo(a, (b, c), d);", nil},
		{"do {} while ((doSomething(), !!test));", nil},
		{"for ((doSomething(), somethingElse()); (doSomething(), !!test); );", nil},
		{"if ((doSomething(), !!test));", nil},
		{"switch ((doSomething(), val)) {}", nil},
		{"while ((doSomething(), !!test));", nil},
		{"with ((doSomething(), val)) {}", nil},
		{"a => ((doSomething(), a))", nil},
		{"var foo = (1, 2);", decodedSequenceOptions(t, `{}`)},
		{"var foo = (1, 2);", decodedSequenceOptions(t, `{"allowInParentheses": true}`)},
		{"for ((i = 0, j = 0); test; );", decodedSequenceOptions(t, `{"allowInParentheses": false}`)},
		{"for (; test; (i++, j++));", decodedSequenceOptions(t, `{"allowInParentheses": false}`)},
		{"const foo = () => { return ((bar = 123), 10) }", nil},
		{"const foo = () => (((bar = 123), 10));", nil},

		// Beyond the corpus. The initializer and the update of a `for`, separately, because the
		// corpus only ever writes them together and a port could exempt one and not the other.
		{"for (a, b;;);", nil},
		{"for (;; a, b);", nil},

		// The exemption reaches through two parenthesis levels as well as one, because upstream's
		// comparison is made on a tree where parentheses are not nodes at all. Measured silent.
		{"for (((i = 0, j = 0)); test; );", decodedSequenceOptions(t, `{"allowInParentheses": false}`)},

		// A three-element chain in each exempt position. These are the cases a mutation sweep found
		// and the imported corpus could not, because upstream never writes a chain longer than two.
		// Deduplicating the chain at the inner node rather than the outer one reports every one of
		// them: the inner node's parent is the outer comma expression rather than the statement, so
		// the `for` exemption and the paren allowance are both asked about the wrong node and both
		// decline. Measured silent against the installed rule.
		{"for (a, b, c;;);", nil},
		{"for (;; a, b, c);", nil},
		{"(a, b, c);", nil},
		{"if ((a, b, c));", nil},
		{"a => ((a, b, c))", nil},
		{"var x = (a, b, c);", nil},
		{"while ((a, b, c));", nil},
		{"switch ((a, b, c)) {}", nil},
		{"function f(){ return (a, b, c); }", nil},
		{"for (var i = (a, b, c);;);", nil},

		// The arrow body with its second parenthesis, the silent half of the minimal pair above.
		{"a => ((b, c))", nil},

		// Positions where one parenthesis is enough and the corpus does not write them: an
		// assignment right-hand side, a template substitution, an optional element access, a
		// `throw`, a computed class member, and a spread.
		{"x = (a, b);", nil},
		{"`${(a, b)}`", nil},
		{"x?.[(a, b)]", nil},
		{"throw (a, b);", nil},
		{"class C { [(a, b)]() {} }", nil},
		{"({...(a, b)})", nil},

		// Arrow PARAMETERS are a separator list rather than a sequence, and they parse into a shape
		// that has no comma operator in it at all. Written because a rule reaching for a comma token
		// would find one here.
		{"(a, b) => c", nil},

		// `new C(a, b)` and `foo(a, b)` are argument separators, which upstream's comment calls out
		// as never reaching the rule. The corpus writes the call form only with a parenthesized
		// sequence inside it.
		{"new C(a, b)", nil},
		{"[a, b]", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunWithOptions(t, NoSequences, sequencesFile, testCase.sourceText, testCase.options))
		})
	}
}

// A nested pair under `allowInParentheses: false` reports twice, and this pins both spans.
//
// It is asserted as a SET rather than as a sequence because the emission order is this port's one
// measured divergence from upstream: our pre-order walk reaches the outer node first, and the outer
// node's reported comma is the later one, so we emit later-then-earlier where upstream emits
// earlier-then-later. Identical count, identical spans, and invisible downstream because the
// reporter sorts by position. Asserting the order here would pin the divergence as if it were the
// decision, which it is not.
func TestNoSequencesReportsBothHalvesOfANestedPair(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantOffset []int
	}{
		{"(a, b), c;", []int{2, 6}},
		{"a = (b, c) , d", []int{6, 11}},
		{"((a, b)), c;", []int{3, 8}},
		{"!(a, b), c", []int{3, 7}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoSequences, sequencesFile, testCase.sourceText,
				decodedSequenceOptions(t, `{"allowInParentheses": false}`))
			rule_testing.ExpectFindings(t, result,
				"unexpectedCommaExpression", "unexpectedCommaExpression")

			got := []int{}
			for _, diagnostic := range result.Diagnostics {
				got = append(got, diagnostic.Range.Pos())
			}
			sort.Ints(got)
			if len(got) != len(testCase.wantOffset) {
				t.Fatalf("reported %d findings, want %d", len(got), len(testCase.wantOffset))
			}
			for index := range got {
				if got[index] != testCase.wantOffset[index] {
					t.Errorf("offsets %v, want %v", got, testCase.wantOffset)
					break
				}
			}
		})
	}
}

// The decoder is the one line of this rule with no upstream counterpart, so it gets its own test.
//
// An absent key and an explicit `true` must both mean allowed, an explicit `false` must mean not
// allowed, and NIL OPTIONS -- which is what the live config produces for a rule configured as bare
// "error" -- must mean allowed. That last one is the case no fixture routed through the decoder can
// reach, because it is the shape the config layer produces when the decoder is never called.
func TestNoSequencesOptionDefaultsToAllowingParentheses(t *testing.T) {
	t.Parallel()

	parenthesized := "var foo = (1, 2);"

	t.Run("nil options, the shape the live config produces", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoSequences, sequencesFile, parenthesized, nil))
	})

	t.Run("an empty object", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoSequences, sequencesFile, parenthesized,
			decodedSequenceOptions(t, `{}`)))
	})

	t.Run("a zero-value struct still allows", func(t *testing.T) {
		t.Parallel()
		// Not a redundant spelling of the empty object: this is the value `rule.DecodeOptionsInto`
		// would have produced, and it is the reason the field is a pointer. If AllowInParentheses
		// were a plain bool this case would report and every other fixture would stay green.
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoSequences, sequencesFile, parenthesized,
			NoSequencesOptions{}))
	})

	t.Run("explicitly false", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoSequences, sequencesFile, parenthesized,
			decodedSequenceOptions(t, `{"allowInParentheses": false}`)), "unexpectedCommaExpression")
	})

	t.Run("empty input decodes to the default rather than erroring", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoSequencesOptions(nil)
		if err != nil {
			t.Fatalf("decoding empty input: %v", err)
		}
		options, ok := decoded.(NoSequencesOptions)
		if !ok {
			t.Fatalf("decoded to %T, want NoSequencesOptions", decoded)
		}
		if !options.allowInParentheses() {
			t.Error("empty input decoded to disallowing parentheses, which inverts the rule")
		}
	})

	t.Run("malformed input is an error rather than a silent default", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeNoSequencesOptions([]byte("[")); err == nil {
			t.Error("malformed options decoded without an error")
		}
	})
}

// The message is a value with no interpolation, so there is nothing to render and the assertion is
// on the constant's own fields. Asserted against literals typed here rather than against the rule's
// own constant, which would move with it under mutation.
func TestNoSequencesMessage(t *testing.T) {
	t.Parallel()

	if messageUnexpectedCommaExpression.Id != "unexpectedCommaExpression" {
		t.Errorf("message id is %q", messageUnexpectedCommaExpression.Id)
	}
	if messageUnexpectedCommaExpression.Description == "" {
		t.Error("message carries no description")
	}
}

// A guard against the decoder growing a second spelling of the same option, which is how a serde
// alias silently widens a rule. Upstream's schema names exactly one property and forbids the rest,
// so a misspelling is refused by name, as ESLint refuses it (#4a4yse4); it used to be ignored.
func TestNoSequencesDecoderRefusesAMisspelledKey(t *testing.T) {
	t.Parallel()

	if _, err := DecodeNoSequencesOptions([]byte(`{"allowInParentheses": false}`)); err != nil {
		t.Fatalf("baseline: upstream's spelling was refused: %v", err)
	}
	_, err := DecodeNoSequencesOptions([]byte(`{"allowInParenthesis": false}`))
	if err == nil || !strings.Contains(err.Error(), "allowInParenthesis") {
		t.Fatalf("want a refusal naming the misspelled key, got %v", err)
	}
}
