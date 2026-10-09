package core

import (
	"encoding/json"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noUnneededTernaryFile is where the fixtures pretend to live.
const noUnneededTernaryFile = "/repository/source/NoUnneededTernary.ts"

// noUnneededTernaryBoolean makes an addressable bool for the one pointer option.
//
// The pointer is not decoration: `defaultAssignment` defaults to TRUE, so a plain bool field cannot
// tell an absent key from an explicit false, and the zero value would switch the second judgment on
// for anyone configuring the rule as a bare severity.
func noUnneededTernaryBoolean(value bool) *bool { return &value }

// noUnneededTernaryCase is one imported corpus row.
type noUnneededTernaryCase struct {
	sourceText string
	options    any
	wantIds    []string
	// wantFixedSource is "" when the case is reported and deliberately NOT repaired, which upstream
	// records as `output: null`.
	wantFixedSource string
}

// runNoUnneededTernary drives one case through the rule's own exported decoder.
func runNoUnneededTernary(t *testing.T, testCase noUnneededTernaryCase) rule_testing.Result {
	t.Helper()
	if testCase.options == nil {
		return rule_testing.Run(t, NoUnneededTernary, noUnneededTernaryFile, testCase.sourceText)
	}
	encoded, err := json.Marshal(testCase.options)
	if err != nil {
		t.Fatalf("could not encode options: %v", err)
	}
	decoded, err := DecodeNoUnneededTernaryOptions(encoded)
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	return rule_testing.RunWithOptions(t, NoUnneededTernary, noUnneededTernaryFile,
		testCase.sourceText, decoded)
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `tests/lib/rules/no-unneeded-ternary.js` was loaded with its RuleTester stubbed so every case came
// out as data, then replayed against the INSTALLED rule to record what it reports and what its fixer
// writes. All 47 findings and all 31 fix outcomes reproduced.
//
// Two of upstream's own cases are TypeScript (`foo as any ? false : true` and
// `foo ? foo : bar as any`), which the JavaScript oracle rejects as parse errors. Those two were
// measured through the TypeScript parser instead. Worth stating because a run that took the
// JavaScript oracle's verdict for them would have recorded two confident parse failures as data.
func noUnneededTernaryFiresCases() []noUnneededTernaryCase {
	return []noUnneededTernaryCase{
		{"var a = x === 2 ? true : false;", nil, []string{"unnecessaryConditionalExpression"}, "var a = x === 2;"},
		{"var a = x >= 2 ? true : false;", nil, []string{"unnecessaryConditionalExpression"}, "var a = x >= 2;"},
		{"var a = x ? true : false;", nil, []string{"unnecessaryConditionalExpression"}, "var a = !!x;"},
		{"var a = x === 1 ? false : true;", nil, []string{"unnecessaryConditionalExpression"}, "var a = x !== 1;"},
		{"var a = x != 1 ? false : true;", nil, []string{"unnecessaryConditionalExpression"}, "var a = x == 1;"},
		{"var a = foo() ? false : true;", nil, []string{"unnecessaryConditionalExpression"}, "var a = !foo();"},
		{"var a = !foo() ? false : true;", nil, []string{"unnecessaryConditionalExpression"}, "var a = !!foo();"},
		{"var a = foo + bar ? false : true;", nil, []string{"unnecessaryConditionalExpression"}, "var a = !(foo + bar);"},
		{"var a = x instanceof foo ? false : true;", nil, []string{"unnecessaryConditionalExpression"}, "var a = !(x instanceof foo);"},
		{"var a = foo ? false : false;", nil, []string{"unnecessaryConditionalExpression"}, "var a = false;"},
		{"var a = foo() ? false : false;", nil, []string{"unnecessaryConditionalExpression"}, ""},
		{"var a = x instanceof foo ? true : false;", nil, []string{"unnecessaryConditionalExpression"}, "var a = x instanceof foo;"},
		{"var a = !foo ? true : false;", nil, []string{"unnecessaryConditionalExpression"}, "var a = !foo;"},
		{"\n                var value = 'a'\n                var canSet = true\n                var result = value ? value : canSet ? 'unset' : 'can not set'\n            ", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "\n                var value = 'a'\n                var canSet = true\n                var result = value || (canSet ? 'unset' : 'can not set')\n            "},
		{"foo ? foo : (bar ? baz : qux)", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "foo || (bar ? baz : qux)"},
		{"function* fn() { foo ? foo : yield bar }", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "function* fn() { foo || (yield bar) }"},
		{"var a = foo ? foo : 'No';", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = foo || 'No';"},
		{"var a = ((foo)) ? (((((foo))))) : ((((((((((((((bar))))))))))))));", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = ((foo)) || ((((((((((((((bar))))))))))))));"},
		{"var a = b ? b : c => c;", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = b || (c => c);"},
		{"var a = b ? b : c = 0;", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = b || (c = 0);"},
		{"var a = b ? b : (c => c);", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = b || (c => c);"},
		{"var a = b ? b : (c = 0);", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = b || (c = 0);"},
		{"var a = b ? b : (c) => (c);", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = b || ((c) => (c));"},
		{"var a = b ? b : c, d; // this is ((b ? b : c), (d))", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = b || c, d; // this is ((b ? b : c), (d))"},
		{"var a = b ? b : (c, d);", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = b || (c, d);"},
		{"f(x ? x : 1);", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "f(x || 1);"},
		{"x ? x : 1;", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "x || 1;"},
		{"var a = foo ? foo : bar;", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = foo || bar;"},
		{"var a = foo ? foo : a ?? b;", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "var a = foo || (a ?? b);"},
		{"foo as any ? false : true", nil, []string{"unnecessaryConditionalExpression"}, "!(foo as any)"},
		{"foo ? foo : bar as any", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, []string{"unnecessaryConditionalAssignment"}, "foo || (bar as any)"},
	}
}

func noUnneededTernarySilentCases() []noUnneededTernaryCase {
	return []noUnneededTernaryCase{
		{"config.newIsCap = config.newIsCap !== false", nil, nil, ""},
		{"var a = x === 2 ? 'Yes' : 'No';", nil, nil, ""},
		{"var a = x === 2 ? true : 'No';", nil, nil, ""},
		{"var a = x === 2 ? 'Yes' : false;", nil, nil, ""},
		{"var a = x === 2 ? 'true' : 'false';", nil, nil, ""},
		{"var a = foo ? foo : bar;", nil, nil, ""},
		{"var value = 'a';var canSet = true;var result = value || (canSet ? 'unset' : 'can not set')", nil, nil, ""},
		{"var a = foo ? bar : foo;", nil, nil, ""},
		{"foo ? bar : foo;", nil, nil, ""},
		{"var a = f(x ? x : 1)", nil, nil, ""},
		{"f(x ? x : 1);", nil, nil, ""},
		{"foo ? foo : bar;", nil, nil, ""},
		{"var a = foo ? 'Yes' : foo;", nil, nil, ""},
		{"var a = foo ? 'Yes' : foo;", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, nil, ""},
		{"var a = foo ? bar : foo;", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, nil, ""},
		{"foo ? bar : foo;", NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)}, nil, ""},
	}
}

func TestNoUnneededTernaryFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range noUnneededTernaryFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoUnneededTernary(t, testCase), testCase.wantIds...)
		})
	}
}

func TestNoUnneededTernaryStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range noUnneededTernarySilentCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoUnneededTernary(t, testCase))
		})
	}
}

// What the fixer WRITES, which is half this rule.
//
// Every arm replaces the whole ternary with reconstructed text, so a repair that lands on the right
// span with the wrong content passes every assertion above. Thirty of upstream's cases carry an
// exact expected output and one deliberately carries none.
func TestNoUnneededTernaryFixesTheSource(t *testing.T) {
	t.Parallel()

	applied := 0
	for _, testCase := range noUnneededTernaryFiresCases() {
		if testCase.wantFixedSource == "" {
			continue
		}
		applied++
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t, runNoUnneededTernary(t, testCase),
				testCase.wantFixedSource)
		})
	}
	if applied != 30 {
		t.Errorf("%d cases asserted a rewrite, wanted 30", applied)
	}
}

// The one case upstream reports and deliberately declines to repair.
//
// `f() ? true : true` collapses to `true` only if the call did not matter, and the rule cannot know
// that, so upstream returns no fix and the finding stands alone. A fixer that helpfully collapsed it
// would delete a call, which is a behaviour change no message-id fixture can see.
func TestNoUnneededTernaryDeclinesToDropACall(t *testing.T) {
	t.Parallel()

	declined := 0
	for _, testCase := range noUnneededTernaryFiresCases() {
		if testCase.wantFixedSource != "" {
			continue
		}
		declined++
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoUnneededTernary(t, testCase)
			if len(result.Diagnostics) == 0 {
				t.Fatal("wanted a finding")
			}
			for index, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d carried %d fixes, wanted none", index,
						len(diagnostic.Fixes))
				}
			}
		})
	}
	if declined != 1 {
		t.Errorf("%d cases declined a repair, wanted 1", declined)
	}
}

// TypeScript syntax inside the replaced span, which is what this fixer risks destroying.
//
// Every arm of this repair REPLACES THE WHOLE TERNARY with reconstructed text, which is the exact
// shape that widened eight declarations to `any` in `no-undef-init` and dropped a `: void` in
// `no-arrow-function-lifecycle`. Both of those rebuilt their replacement from node FIELDS. This one
// slices the source by RANGE, so anything inside the span survives verbatim, and these rows turn
// that argument into a measurement rather than leaving it as a claim in a doc comment.
//
// Upstream's corpus contains two TypeScript cases and cannot contain these: the gap between its
// corpus and our tree is TypeScript syntax. Every verdict was taken from the installed rule driven
// through the TypeScript parser.
//
// The `as`/`satisfies` rows are the load-bearing ones. Upstream's precedence table has no entry for
// either, so a port inheriting it unchanged gives them the TIGHTEST precedence and emits
// `a || b as any`, which parses as `(a || b) as any` and asserts the type of the wrong expression.
func TestNoUnneededTernaryPreservesTypeSyntax(t *testing.T) {
	t.Parallel()

	defaultAssignmentOff := NoUnneededTernaryOptions{
		DefaultAssignment: noUnneededTernaryBoolean(false),
	}
	cases := []noUnneededTernaryCase{
		// A type annotation on the assignment target sits outside the replaced span and survives.
		{"declare const foo: unknown; const x: boolean = foo ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"},
			"declare const foo: unknown; const x: boolean = !!foo;"},
		// `satisfies` inside the test, which must be parenthesized under the negation.
		{"declare const foo: boolean; const x = foo satisfies boolean ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"},
			"declare const foo: boolean; const x = !!(foo satisfies boolean);"},
		// A parenthesized `as`, whose own parentheses are kept as written.
		{"declare const foo: unknown; const x = (foo as string) ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"},
			"declare const foo: unknown; const x = !!((foo as string));"},
		// A non-null assertion, which our AST hangs off the operand as its own node.
		{"declare const foo: string | undefined; const x = foo! ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"},
			"declare const foo: string | undefined; const x = !!(foo!);"},
		// Explicit type arguments on a call, which a JavaScript parser reads as two comparisons.
		{"declare function f<T>(): T; const x = f<number>() ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"},
			"declare function f<T>(): T; const x = !!f<number>();"},
		// An angle-bracket cast, which has no JavaScript equivalent at all.
		{"declare const a: unknown; const x = <boolean>a ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"},
			"declare const a: unknown; const x = !!(<boolean>a);"},
		// The default-assignment arm with an `as` alternate: parenthesized, because `as` binds
		// LOOSER than the `||` the repair introduces.
		{"declare const a: unknown; declare const b: unknown; const x = a ? a : b as any;",
			defaultAssignmentOff, []string{"unnecessaryConditionalAssignment"},
			"declare const a: unknown; declare const b: unknown; const x = a || (b as any);"},
		{"declare const a: unknown; declare const b: unknown; " +
			"const x = a ? a : b satisfies unknown;",
			defaultAssignmentOff, []string{"unnecessaryConditionalAssignment"},
			"declare const a: unknown; declare const b: unknown; " +
				"const x = a || (b satisfies unknown);"},
		// The controls for those two: a `||` alternate is NOT parenthesized, because it binds at
		// the same precedence, while a `??` alternate is, because the grammar forbids mixing them.
		// Without these rows a rule that parenthesized everything would look correct.
		{"declare const a: unknown; declare const b: unknown; declare const c: unknown; " +
			"const x = a ? a : b || c;",
			defaultAssignmentOff, []string{"unnecessaryConditionalAssignment"},
			"declare const a: unknown; declare const b: unknown; declare const c: unknown; " +
				"const x = a || b || c;"},
		{"declare const a: unknown; declare const b: unknown; declare const c: unknown; " +
			"const x = a ? a : b ?? c;",
			defaultAssignmentOff, []string{"unnecessaryConditionalAssignment"},
			"declare const a: unknown; declare const b: unknown; declare const c: unknown; " +
				"const x = a || (b ?? c);"},
		// An `as` in a branch rather than in the test is not a boolean literal, so the rule
		// declines entirely. This is the row that separates "sees through everything" from "reads
		// the branch correctly".
		{"declare const foo: unknown; const x = foo ? true : false as boolean;", nil, nil, ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoUnneededTernary(t, testCase)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
		})
	}
}

// Where the finding points, which no message-id fixture can see.
//
// Upstream reports the whole conditional expression, and the repair replaces exactly that span. A
// repair whose range were wider than the finding would edit code the reader was never shown.
func TestNoUnneededTernaryPointsAtTheWholeConditional(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		wantSpan   string
	}{
		{"var a = x === 2 ? true : false;", nil, "x === 2 ? true : false"},
		{"var a = x ? true : false;", nil, "x ? true : false"},
		{"var a = foo ? foo : bar;",
			NoUnneededTernaryOptions{DefaultAssignment: noUnneededTernaryBoolean(false)},
			"foo ? foo : bar"},
		{"var a = f() ? true : true;", nil, "f() ? true : true"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoUnneededTernary(t, noUnneededTernaryCase{
				sourceText: testCase.sourceText, options: testCase.options})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			if got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]; got !=
				testCase.wantSpan {
				t.Errorf("finding pointed at %q, wanted %q", got, testCase.wantSpan)
			}
			// The repair, where there is one, must cover exactly the reported span.
			for index, fix := range diagnostic.Fixes {
				if fix.Range.Pos() != diagnostic.Range.Pos() ||
					fix.Range.End() != diagnostic.Range.End() {
					t.Errorf("fix %d spans [%d,%d) while the finding spans [%d,%d)", index,
						fix.Range.Pos(), fix.Range.End(), diagnostic.Range.Pos(),
						diagnostic.Range.End())
				}
			}
		})
	}
}

// The decoder, whose only option defaults to TRUE.
//
// A rule configured as a bare severity is handed nil, and `options.(T)` on nil yields the zero
// value, whose false would switch the second judgment ON for everyone. Every fixture above reaches
// the rule through the decoder, so none of them can see that.
func TestDecodeNoUnneededTernaryOptions(t *testing.T) {
	t.Parallel()

	t.Run("nil input keeps the true default", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoUnneededTernaryOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !decoded.(NoUnneededTernaryOptions).resolve() {
			t.Error("defaultAssignment resolved to false, which turns the second judgment on")
		}
	})

	t.Run("an explicit false overrides it", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoUnneededTernaryOptions([]byte(`{"defaultAssignment":false}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if decoded.(NoUnneededTernaryOptions).resolve() {
			t.Error("an explicit false was read as the default true")
		}
	})
}

// With nil options the default-assignment judgment stays OFF and the boolean one stays on.
func TestNoUnneededTernaryWithNilOptionsUsesTheDefault(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoUnneededTernary, noUnneededTernaryFile,
		"var a = x ? true : false;"), "unnecessaryConditionalExpression")
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnneededTernary, noUnneededTernaryFile,
		"var a = foo ? foo : bar;"))
}

// The precedence question upstream asks, which our AST answers differently by default.
//
// Espree produces no node for parentheses, so upstream's `getPrecedence(node.test)` on `(a + b)` is
// asking about the ADDITION and answers 12, which is below unary precedence, so the negation wraps
// it again and the output carries both pairs. Our parser gives the parentheses a node that
// genuinely binds tightest, so the shared precedence table answers 20 and one pair is lost.
//
// And an unknown node type answers -1 upstream rather than 20, deliberately, so that a fixer wraps
// what it does not recognise. `foo!` is a `TSNonNullExpression`, which is not in eslint's visitor
// keys, so it takes that arm. The wrong answer here produces `!!foo!`, which PARSES, so nothing
// downstream would ever have complained.
//
// Both were found by a failing fixture rather than by reading the table, and neither can be reached
// from upstream's JavaScript corpus.
func TestNoUnneededTernaryAsksPrecedenceOfWhatTheParenthesesHold(t *testing.T) {
	t.Parallel()

	cases := []noUnneededTernaryCase{
		// A parenthesized loose-binding test keeps its own pair AND gains one.
		{"const x = (a + b) ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"}, "const x = !!((a + b));"},
		// A parenthesized tight-binding test keeps only its own pair, which is the control that
		// separates "asks about the inside" from "always wraps a parenthesized test".
		{"const x = (foo) ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"}, "const x = !!(foo);"},
		{"const x = ((foo)) ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"}, "const x = !!((foo));"},
		// The unparenthesized form of the first row, which needs exactly one pair.
		{"const x = a + b ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"}, "const x = !!(a + b);"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoUnneededTernary(t, testCase)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
		})
	}
}

// Shapes upstream's corpus exercises but whose distinguishing rows the imported set missed.
//
// Three surviving mutants pointed here, and each names a different gap:
//
//	the inverse table       only two of its four rows were reached, so swapping one survived
//	the branch paren skip   no imported case parenthesizes a boolean branch, so removing the
//	                        skip left every row green while `foo ? (true) : false` went silent
//	the relational set      `in` and `instanceof` produce a boolean, so a test using either is
//	                        emitted bare rather than double-negated
//
// Every verdict was measured against the installed rule before the row was written.
func TestNoUnneededTernaryCoversEveryInverseAndParenthesizedBranch(t *testing.T) {
	t.Parallel()

	cases := []noUnneededTernaryCase{
		// All four equality operators have a true inverse, and each row pins one direction.
		{"var a = x == 2 ? false : true;", nil,
			[]string{"unnecessaryConditionalExpression"}, "var a = x != 2;"},
		{"var a = x != 2 ? false : true;", nil,
			[]string{"unnecessaryConditionalExpression"}, "var a = x == 2;"},
		{"var a = x === 2 ? false : true;", nil,
			[]string{"unnecessaryConditionalExpression"}, "var a = x !== 2;"},
		{"var a = x !== 2 ? false : true;", nil,
			[]string{"unnecessaryConditionalExpression"}, "var a = x === 2;"},
		// A parenthesized boolean branch is still a boolean branch. Espree gives parentheses no
		// node, so upstream reads the literal directly and the skip is fidelity rather than a
		// widening -- the same correction `no-eq-null` documents for its operands.
		{"var a = foo ? (true) : false;", nil,
			[]string{"unnecessaryConditionalExpression"}, "var a = !!foo;"},
		{"var a = foo ? true : (false);", nil,
			[]string{"unnecessaryConditionalExpression"}, "var a = !!foo;"},
		{"var a = foo ? (true) : (true);", nil,
			[]string{"unnecessaryConditionalExpression"}, "var a = true;"},
		// `in` and `instanceof` always produce a boolean, so the test is emitted bare.
		{"var a = x in y ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"}, "var a = x in y;"},
		{"var a = x instanceof y ? true : false;", nil,
			[]string{"unnecessaryConditionalExpression"}, "var a = x instanceof y;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoUnneededTernary(t, testCase)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
		})
	}
}
