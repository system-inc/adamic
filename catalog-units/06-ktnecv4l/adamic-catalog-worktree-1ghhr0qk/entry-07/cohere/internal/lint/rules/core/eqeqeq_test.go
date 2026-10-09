package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// eqeqeqFile is where the fixtures pretend to live.
const eqeqeqFile = "/repository/source/Eqeqeq.ts"

// eqeqeqRepair says which KIND of repair a finding must carry.
//
// This is the whole rule, and it is the one property no message-id fixture can see. Every finding
// this rule produces carries exactly one repair, and whether that repair is applied unattended or
// offered to a human is the difference between a linter that tidies spelling and one that changes
// what a program computes. A port shipping the suggestion arm as a fix would rewrite `a == b` to
// `a === b` across the tree without asking.
type eqeqeqRepair int

const (
	// eqeqeqRepairFix is applied unattended, and is correct only where coercion is impossible.
	eqeqeqRepairFix eqeqeqRepair = iota
	// eqeqeqRepairSuggestion is offered to a human, because applying it changes behaviour.
	eqeqeqRepairSuggestion
)

// eqeqeqCase is one imported corpus row.
type eqeqeqCase struct {
	sourceText string
	options    any
	// wantRepairs is one entry per expected finding, in order. nil means the case is clean.
	wantRepairs []eqeqeqRepair
	// wantFixedSource is what the fix phase writes, or "" when every finding is suggestion-only
	// and the source therefore comes back unchanged.
	wantFixedSource string
}

// runEqeqeq drives one case, routing options through the rule's own exported decoder.
//
// Through the decoder rather than by building the struct, because the decoder is where upstream's
// mode-DEPENDENT default lives -- the null policy defaults to Always under Always and is forced to
// Ignore under Smart -- and where an unknown spelling is rejected. Neither line has an upstream
// counterpart to inherit correctness from.
func runEqeqeq(t *testing.T, testCase eqeqeqCase) rule_testing.Result {
	t.Helper()
	if testCase.options == nil {
		return rule_testing.Run(t, Eqeqeq, eqeqeqFile, testCase.sourceText)
	}
	decoded, err := DecodeEqeqeqOptions([]byte(eqeqeqUpstreamList(testCase.options.(EqeqeqOptions))))
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	return rule_testing.RunWithOptions(t, Eqeqeq, eqeqeqFile, testCase.sourceText, decoded)
}

// eqeqeqUpstreamList writes a case's options the way upstream's config writes them, so every case
// crosses the decoder on the list shape the config layer delivers rather than on a struct.
func eqeqeqUpstreamList(options EqeqeqOptions) string {
	mode := strings.ToLower(string(options.Mode))
	if mode == "" {
		mode = "always"
	}
	if options.Null == "" {
		return `["` + mode + `"]`
	}
	return `["` + mode + `", {"null": "` + strings.ToLower(string(options.Null)) + `"}]`
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `tests/lib/rules/eqeqeq.js` was loaded with its RuleTester stubbed so every case came out as data
// with its options attached, then each was replayed against the INSTALLED rule in `node_modules` to
// record what it reports, which repair each finding carries, and what the fix phase writes. All 77
// findings and all 46 fix outcomes reproduced, so everything below is a measurement rather than a
// transcription of the `errors` arrays.
//
// Upstream's options are a positional list whose legal second element depends on the first
// (`["always", {null}]` is valid, `["smart", {null}]` is not). Each case below names the struct and
// `eqeqeqUpstreamList` writes it back as that list, so the decoder reads upstream's spelling. Its
// deprecated `"allow-null"` spelling is rendered as the combination it stands for, `Always` plus
// `Null: Ignore`, and `TestDecodeEqeqeqOptions` pins that it decodes to the same.
func eqeqeqFiresCases() []eqeqeqCase {
	return []eqeqeqCase{
		{"a == b", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"a != b", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"typeof a == 'number'", nil, []eqeqeqRepair{eqeqeqRepairFix}, "typeof a === 'number'"},
		{"typeof a == 'number'", EqeqeqOptions{Mode: EqeqeqAlways}, []eqeqeqRepair{eqeqeqRepairFix}, "typeof a === 'number'"},
		{"'string' != typeof a", nil, []eqeqeqRepair{eqeqeqRepairFix}, "'string' !== typeof a"},
		{"true == true", nil, []eqeqeqRepair{eqeqeqRepairFix}, "true === true"},
		{"2 == 3", nil, []eqeqeqRepair{eqeqeqRepairFix}, "2 === 3"},
		{"2 == 3", EqeqeqOptions{Mode: EqeqeqAlways}, []eqeqeqRepair{eqeqeqRepairFix}, "2 === 3"},
		{"'hello' != 'world'", nil, []eqeqeqRepair{eqeqeqRepairFix}, "'hello' !== 'world'"},
		{"'hello' != 'world'", EqeqeqOptions{Mode: EqeqeqAlways}, []eqeqeqRepair{eqeqeqRepairFix}, "'hello' !== 'world'"},
		{"`hello` == `world`", nil, []eqeqeqRepair{eqeqeqRepairFix}, "`hello` === `world`"},
		{"`hello` != 'world'", nil, []eqeqeqRepair{eqeqeqRepairFix}, "`hello` !== 'world'"},
		{"a == null", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"a == null", EqeqeqOptions{Mode: EqeqeqAlways}, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"null != a", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"true == 1", EqeqeqOptions{Mode: EqeqeqSmart}, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"0 != '1'", EqeqeqOptions{Mode: EqeqeqSmart}, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"'wee' == /wee/", EqeqeqOptions{Mode: EqeqeqSmart}, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"`hello${world}` == `hello`", EqeqeqOptions{Mode: EqeqeqSmart}, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"typeof a == 'number'", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, []eqeqeqRepair{eqeqeqRepairFix}, "typeof a === 'number'"},
		{"'string' != typeof a", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, []eqeqeqRepair{eqeqeqRepairFix}, "'string' !== typeof a"},
		{"'hello' != 'world'", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, []eqeqeqRepair{eqeqeqRepairFix}, "'hello' !== 'world'"},
		{"2 == 3", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, []eqeqeqRepair{eqeqeqRepairFix}, "2 === 3"},
		{"true == true", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, []eqeqeqRepair{eqeqeqRepairFix}, "true === true"},
		{"true == null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullAlways}, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"true != null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullAlways}, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"null == null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullAlways}, []eqeqeqRepair{eqeqeqRepairFix}, "null === null"},
		{"null != null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullAlways}, []eqeqeqRepair{eqeqeqRepairFix}, "null !== null"},
		{"true === null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"true !== null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"null === null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, []eqeqeqRepair{eqeqeqRepairFix}, "null == null"},
		{"null !== null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, []eqeqeqRepair{eqeqeqRepairFix}, "null != null"},
		{"a\n==\nb", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"(a) == b", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"(a) != b", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"a == (b)", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"a != (b)", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"(a) == (b)", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"(a) != (b)", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"(a == b) == (c)", nil, []eqeqeqRepair{eqeqeqRepairSuggestion, eqeqeqRepairSuggestion}, ""},
		{"(a != b) != (c)", nil, []eqeqeqRepair{eqeqeqRepairSuggestion, eqeqeqRepairSuggestion}, ""},
		{"a == b;", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"a!=b;", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"(a + b) == c;", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"(a + b)  !=  c;", nil, []eqeqeqRepair{eqeqeqRepairSuggestion}, ""},
		{"((1) )  ==  (2);", nil, []eqeqeqRepair{eqeqeqRepairFix}, "((1) )  ===  (2);"},
	}
}

func eqeqeqSilentCases() []eqeqeqCase {
	return []eqeqeqCase{
		{"a === b", nil, nil, ""},
		{"a !== b", nil, nil, ""},
		{"a === b", EqeqeqOptions{Mode: EqeqeqAlways}, nil, ""},
		{"typeof a == 'number'", EqeqeqOptions{Mode: EqeqeqSmart}, nil, ""},
		{"'string' != typeof a", EqeqeqOptions{Mode: EqeqeqSmart}, nil, ""},
		{"'hello' != 'world'", EqeqeqOptions{Mode: EqeqeqSmart}, nil, ""},
		{"`hello` != `world`", EqeqeqOptions{Mode: EqeqeqSmart}, nil, ""},
		{"`hello` == 'hello'", EqeqeqOptions{Mode: EqeqeqSmart}, nil, ""},
		{"2 == 3", EqeqeqOptions{Mode: EqeqeqSmart}, nil, ""},
		{"true == true", EqeqeqOptions{Mode: EqeqeqSmart}, nil, ""},
		{"null == a", EqeqeqOptions{Mode: EqeqeqSmart}, nil, ""},
		{"a == null", EqeqeqOptions{Mode: EqeqeqSmart}, nil, ""},
		{"null == a", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, nil, ""},
		{"a == null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, nil, ""},
		{"a == null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, nil, ""},
		{"a != null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, nil, ""},
		{"a !== null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullIgnore}, nil, ""},
		{"a === null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullAlways}, nil, ""},
		{"a !== null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullAlways}, nil, ""},
		{"null === null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullAlways}, nil, ""},
		{"null !== null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullAlways}, nil, ""},
		{"a == null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
		{"a != null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
		{"null == null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
		{"null != null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
		{"a >= null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
		{"a + null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
		{"null + null", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
		{"null instanceof Foo", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
		{"foo === /abc/u", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
		{"foo === 1n", EqeqeqOptions{Mode: EqeqeqAlways, Null: EqeqeqNullNever}, nil, ""},
	}
}

func TestEqeqeqFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range eqeqeqFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runEqeqeq(t, testCase)
			wantIds := make([]string, len(testCase.wantRepairs))
			for index := range testCase.wantRepairs {
				wantIds[index] = "unexpected"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

func TestEqeqeqStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range eqeqeqSilentCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runEqeqeq(t, testCase))
		})
	}
}

// Which KIND of repair each finding carries, which is the rule's central judgment.
//
// Twenty of upstream's findings carry a Fix and twenty-eight carry a Suggestion, and none carries
// both. The split is not stylistic: a Fix is applied by the edit engine with nobody watching, and
// `a == b` becoming `a === b` can change what the program computes. A port that got the judgment
// right and the ARM wrong would pass every count assertion above while rewriting the tree.
func TestEqeqeqSplitsFixesFromSuggestions(t *testing.T) {
	t.Parallel()

	fixArm, suggestionArm := 0, 0
	for _, testCase := range eqeqeqFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runEqeqeq(t, testCase)
			if len(result.Diagnostics) != len(testCase.wantRepairs) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantRepairs),
					len(result.Diagnostics))
			}
			for index, want := range testCase.wantRepairs {
				diagnostic := result.Diagnostics[index]
				switch want {
				case eqeqeqRepairFix:
					if len(diagnostic.Fixes) != 1 || len(diagnostic.Suggestions) != 0 {
						t.Errorf("finding %d carried %d fixes and %d suggestions, wanted "+
							"exactly one fix and no suggestion", index, len(diagnostic.Fixes),
							len(diagnostic.Suggestions))
					}
				case eqeqeqRepairSuggestion:
					if len(diagnostic.Suggestions) != 1 || len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d carried %d suggestions and %d fixes, wanted "+
							"exactly one suggestion and no fix", index,
							len(diagnostic.Suggestions), len(diagnostic.Fixes))
					}
				}
			}
		})
		for _, want := range testCase.wantRepairs {
			if want == eqeqeqRepairFix {
				fixArm++
			} else {
				suggestionArm++
			}
		}
	}
	// The totals, so a wholesale drift from one arm to the other fails even if every individual
	// row were somehow edited to agree with it.
	if fixArm != 20 || suggestionArm != 28 {
		t.Errorf("the corpus carried %d fix-arm and %d suggestion-arm findings, wanted 20 and 28",
			fixArm, suggestionArm)
	}
}

// What the fix phase WRITES, for the twenty cases that carry an applicable repair.
//
// `ExpectFixedSource` replays every fix into the source and compares the whole rewritten file, so it
// catches a repair that lands on the right span with the wrong text. Suggestions are deliberately
// not applied here: the harness cannot apply them, and more importantly upstream's engine does not
// either, which is what `output: null` records for those twenty-six cases.
//
// The expectation is upstream's `output` verbatim, with no trailing-newline transform. That is
// harness-specific and worth stating: `rule_testing.Run` does not trim, while `RunTyped` writes each
// fixture as `TrimSpace(source)+"\n"` (`rule_testing/program.go`), so a typed rule's expectation
// must be transformed the same way. This rule needs no checker, so it is compared as written.
func TestEqeqeqFixesTheSource(t *testing.T) {
	t.Parallel()

	applied := 0
	for _, testCase := range eqeqeqFiresCases() {
		if testCase.wantFixedSource == "" {
			continue
		}
		applied++
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t, runEqeqeq(t, testCase), testCase.wantFixedSource)
		})
	}
	if applied != 20 {
		t.Errorf("%d cases asserted a rewrite, wanted 20", applied)
	}
}

// The twenty-six cases upstream reports and deliberately declines to apply.
//
// Their `output` is null because every finding on them is suggestion-only, so the fix phase leaves
// the source alone. Asserted as carrying NO fix rather than as a fix that happens to be a no-op:
// those are different artifacts and only the first is what upstream ships.
func TestEqeqeqDeclinesToApplyASuggestion(t *testing.T) {
	t.Parallel()

	declined := 0
	for _, testCase := range eqeqeqFiresCases() {
		if testCase.wantFixedSource != "" {
			continue
		}
		declined++
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runEqeqeq(t, testCase)
			for index, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d carried %d fixes, wanted none so the fix phase leaves "+
						"the source alone", index, len(diagnostic.Fixes))
				}
			}
		})
	}
	if declined != 26 {
		t.Errorf("%d cases declined a rewrite, wanted 26", declined)
	}
}

// TypeScript shapes upstream's corpus cannot contain, because its corpus is JavaScript.
//
// Two fixers shipped tonight that were right about JavaScript and destroyed type information here,
// because they built a replacement span from a node's neighbour and our AST puts `Type`,
// `ExclamationToken` and type arguments where ESTree does not. This rule is structurally safe from
// that -- its repair replaces the OPERATOR TOKEN and nothing else, so nothing can be swallowed --
// and these rows are what turns that argument into a measurement.
//
// Every verdict was taken from the installed rule driven through the TypeScript parser, not from
// the JavaScript oracle, which rejects all of these as parse errors and would have reported a
// confident zero.
func TestEqeqeqPreservesTypeSyntax(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		wantRepair eqeqeqRepair
		// wantFixedSource is asserted only on the fix arm; a suggestion is never applied.
		wantFixedSource string
	}{
		// A type annotation sits between the binding name and the initializer in our AST. The
		// repair never goes near it, so the annotation survives.
		{"declare const a: unknown; typeof a == 'string'", nil, eqeqeqRepairFix,
			"declare const a: unknown; typeof a === 'string'"},
		// An `as` expression on the left of the comparison.
		{"const x = 1 as number; typeof x == 'number'", nil, eqeqeqRepairFix,
			"const x = 1 as number; typeof x === 'number'"},
		// Explicit type arguments on a call, which is syntax the JavaScript parser reads as two
		// comparisons and ours reads as one call.
		{"declare function f<T>(): T; typeof f<number>() == 'number'", nil, eqeqeqRepairFix,
			"declare function f<T>(): T; typeof f<number>() === 'number'"},
		// The suggestion arm, where nothing is applied at all: a union-typed operand against null.
		{"declare const a: string | undefined; a == null", nil, eqeqeqRepairSuggestion, ""},
		// A parenthesized `as` expression, which is the shape closest to a span-building mistake.
		{"declare const a: unknown; (a as string) == 'x'", nil, eqeqeqRepairSuggestion, ""},
		// A non-null assertion, which our AST hangs off the operand as its own node.
		{"declare const a: number; a! == 1", nil, eqeqeqRepairSuggestion, ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runEqeqeq(t, eqeqeqCase{
				sourceText: testCase.sourceText, options: testCase.options})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			switch testCase.wantRepair {
			case eqeqeqRepairFix:
				if len(diagnostic.Fixes) != 1 || len(diagnostic.Suggestions) != 0 {
					t.Fatalf("wanted one fix and no suggestion, got %d and %d",
						len(diagnostic.Fixes), len(diagnostic.Suggestions))
				}
				rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
			case eqeqeqRepairSuggestion:
				if len(diagnostic.Suggestions) != 1 || len(diagnostic.Fixes) != 0 {
					t.Fatalf("wanted one suggestion and no fix, got %d and %d",
						len(diagnostic.Suggestions), len(diagnostic.Fixes))
				}
			}
		})
	}
}

// The repair replaces exactly the operator, and nothing on either side of it.
//
// The rows above assert the rewritten source, which would still pass if the repair happened to
// reproduce the neighbouring bytes it deleted. This asserts the SPAN directly: the fix range must
// cover the operator token and no more. That is the property that makes this rule structurally
// unable to eat a type annotation, so it is worth stating as a property rather than as six examples.
func TestEqeqeqRepairSpansOnlyTheOperator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
	}{
		{"a == b", "=="},
		{"a != b", "!="},
		{"typeof a == 'number'", "=="},
		{"a\n  ==\n  b", "=="},
		{"a /* c */ == /* d */ b", "=="},
		// Parenthesized operands, where a span built from a neighbour would swallow a paren.
		{"((1) )  ==  (2);", "=="},
		{"declare const a: unknown; typeof a == 'string'", "=="},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runEqeqeq(t, eqeqeqCase{sourceText: testCase.sourceText})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			// The finding points at the operator, which is upstream's `loc: operatorToken.loc`
			// rather than the whole binary expression.
			if got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]; got !=
				testCase.wantSpan {
				t.Errorf("finding pointed at %q, wanted %q", got, testCase.wantSpan)
			}
			repairs := diagnostic.Fixes
			if len(repairs) == 0 && len(diagnostic.Suggestions) == 1 {
				repairs = diagnostic.Suggestions[0].Fixes
			}
			if len(repairs) != 1 {
				t.Fatalf("wanted exactly one repair, got %d", len(repairs))
			}
			if got := testCase.sourceText[repairs[0].Range.Pos():repairs[0].Range.End()]; got !=
				testCase.wantSpan {
				t.Errorf("the repair replaces %q, wanted %q", got, testCase.wantSpan)
			}
		})
	}
}

// The decoder, which reads upstream's option list and carries a MODE-DEPENDENT default.
//
// A fixture that built the options struct directly would leave the list reading, the mode-dependent
// second element, and the rejection of an unknown spelling completely untested.
func TestDecodeEqeqeqOptions(t *testing.T) {
	t.Parallel()

	t.Run("nil input selects upstream's defaults", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeEqeqeqOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		settings := decoded.(EqeqeqOptions).resolve()
		if settings.mode != EqeqeqAlways || settings.null != EqeqeqNullAlways {
			t.Errorf("defaults resolved to %+v, wanted Always and Always", settings)
		}
	})

	t.Run("the second element is read: always plus null ignore", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeEqeqeqOptions([]byte(`["always", {"null": "ignore"}]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if settings := decoded.(EqeqeqOptions).resolve(); settings.mode != EqeqeqAlways ||
			settings.null != EqeqeqNullIgnore {
			t.Errorf("resolved to %+v, wanted Always and Ignore", settings)
		}
	})

	t.Run("allow-null is always plus null ignore, as upstream reads it", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeEqeqeqOptions([]byte(`["allow-null"]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if settings := decoded.(EqeqeqOptions).resolve(); settings.mode != EqeqeqAlways ||
			settings.null != EqeqeqNullIgnore {
			t.Errorf("resolved to %+v, wanted Always and Ignore", settings)
		}
	})

	t.Run("smart takes no second element, so one is refused rather than discarded", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeEqeqeqOptions([]byte(`["smart", {"null": "never"}]`)); err == nil {
			t.Error("a null policy beside smart decoded, and the rule would never read it")
		}
	})

	t.Run("a third element is refused", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeEqeqeqOptions([]byte(`["always", {"null": "ignore"}, "smart"]`)); err == nil {
			t.Error("a third element decoded and would have been dropped")
		}
	})

	t.Run("an unknown mode is rejected rather than disabling the rule", func(t *testing.T) {
		t.Parallel()
		// The Go constant's spelling, which is not upstream's.
		if _, err := DecodeEqeqeqOptions([]byte(`["Always"]`)); err == nil {
			t.Error("an unrecognized mode decoded; every arm here is selected by equality, so " +
				"an unrecognized string would report nothing at all")
		}
	})

	t.Run("an unknown null policy is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeEqeqeqOptions([]byte(`["always", {"null": "sometimes"}]`)); err == nil {
			t.Error("an unrecognized null policy decoded")
		}
	})

	t.Run("an unknown key in the second element is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeEqeqeqOptions([]byte(`["always", {"nul": "ignore"}]`)); err == nil {
			t.Error("a misspelled key decoded, so the null policy it carried would be ignored")
		}
	})

	t.Run("the retired object spelling is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeEqeqeqOptions([]byte(`{"mode":"Smart","null":"Never"}`)); err == nil {
			t.Error("the pre-list object spelling decoded, which the config layer never delivers")
		}
	})
}

// A rule configured as a bare severity is handed nil options and must still enforce the defaults.
//
// `options.(T)` on nil yields the zero value, whose empty mode and empty null policy match no arm.
// Every fixture above reaches the rule through the decoder, so none of them can see this.
func TestEqeqeqWithNilOptionsUsesTheDefaults(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, Eqeqeq, eqeqeqFile, "a == b"), "unexpected")
	// The null policy defaults to Always, so a null comparison reports rather than being exempt.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, Eqeqeq, eqeqeqFile, "a == null"),
		"unexpected")
	// And the mode defaults to Always rather than Smart, so a typeof comparison still reports.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, Eqeqeq, eqeqeqFile, "typeof a == 'x'"),
		"unexpected")
	rule_testing.ExpectClean(t, rule_testing.Run(t, Eqeqeq, eqeqeqFile, "a === b"))
}

// Two shapes upstream's corpus does not write, each found by a surviving mutant.
//
// Upstream's cases never pair a number with a bigint, and never parenthesize a `typeof`. Both gaps
// let a real discrimination be deleted with all 77 imported rows staying green:
//
//	the bigint arm          collapsing `typeof 1n` into "number" makes `1 == 1n` take the FIX arm,
//	                        which would rewrite an operator whose operands genuinely differ in type
//	the parenthesis skip    reading the operand directly puts `(typeof a) == 'number'` on the
//	                        SUGGESTION arm, so a provably safe repair stops being applied
//
// The first is the dangerous direction and the second is the quiet one. Every verdict below was
// measured against the installed rule before the row was written, because reading the code cannot
// settle whether upstream's `typeof node.value` separates a number from a bigint.
func TestEqeqeqSeparatesBigIntAndSeesThroughParentheses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText      string
		wantRepair      eqeqeqRepair
		wantFixedSource string
	}{
		// A number and a bigint are different `typeof` classes, so coercion is possible and the
		// repair is only offered.
		{"1 == 1n", eqeqeqRepairSuggestion, ""},
		// Two bigints are the same class, so the repair is applied.
		{"1n == 1n", eqeqeqRepairFix, "1n === 1n"},
		// A parenthesized `typeof` on either side still takes the fix arm.
		{"(typeof a) == 'number'", eqeqeqRepairFix, "(typeof a) === 'number'"},
		{"typeof a == ('number')", eqeqeqRepairFix, "typeof a === ('number')"},
		// Both sides `typeof`, which upstream's corpus also omits.
		{"typeof a == typeof b", eqeqeqRepairFix, "typeof a === typeof b"},
		// Parenthesized literals, which reach the same-typed-literals predicate rather than the
		// typeof one.
		{"(1) == (2)", eqeqeqRepairFix, "(1) === (2)"},
		{"('a') == ('b')", eqeqeqRepairFix, "('a') === ('b')"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runEqeqeq(t, eqeqeqCase{sourceText: testCase.sourceText})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			if testCase.wantRepair == eqeqeqRepairFix {
				if len(diagnostic.Fixes) != 1 || len(diagnostic.Suggestions) != 0 {
					t.Fatalf("wanted the fix arm, got %d fixes and %d suggestions",
						len(diagnostic.Fixes), len(diagnostic.Suggestions))
				}
				rule_testing.ExpectFixedSource(t, result, testCase.wantFixedSource)
				return
			}
			if len(diagnostic.Suggestions) != 1 || len(diagnostic.Fixes) != 0 {
				t.Fatalf("wanted the suggestion arm, got %d suggestions and %d fixes",
					len(diagnostic.Suggestions), len(diagnostic.Fixes))
			}
		})
	}
}
