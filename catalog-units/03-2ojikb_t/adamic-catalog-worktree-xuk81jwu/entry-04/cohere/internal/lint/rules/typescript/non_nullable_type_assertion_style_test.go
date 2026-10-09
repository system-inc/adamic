package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// nonNullableTypeAssertionStyleFile names the fixture file.
//
// The rule reads no path and gates on no extension: upstream registers a bare visitor over the two
// assertion nodes with no source-type test. It ends in .ts rather than .tsx deliberately, because an
// angle-bracket assertion is not parseable in a .tsx file at all and half of upstream's corpus would
// stop being expressible.
const nonNullableTypeAssertionStyleFile = "/repository/source/Assertions.ts"

// nonNullableTypeAssertionStyleCaseName numbers a row so a failure names which one.
func nonNullableTypeAssertionStyleCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// asTheTypedHarnessWroteIt transforms an upstream `output` the way RunTyped transforms its input.
//
// rule_testing/program.go writes each fixture as strings.TrimSpace(contents)+"\n", so the file the
// rule is actually run against is not the string in the Go literal above it. Every case in this
// corpus carries a leading newline and trailing indentation, so ExpectFixedSource comparing against
// the untransformed upstream text fails on whitespace for nine byte-correct repairs, and it reads as
// nine broken fixers rather than as one harness fact. The transform is applied here rather than by
// padding the rule, which would be making the rule wrong to make the comparison line up.
func asTheTypedHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// TestNonNullableTypeAssertionStyleStaysSilentOnUpstreamPassCases is the imported clean corpus,
// verbatim.
//
// All eleven of upstream's passing inputs, extracted from the clone's test file by parsing it with
// the TypeScript compiler rather than by reading it, then byte verified against the source. Every
// one was replayed through the installed 8.x build with a real type checker, and all eleven reported
// nothing.
//
// These are the false positives upstream already thought about, and several of them are the only
// thing standing between this port and a rule that rewrites working code. Case 3 asserts to `any`,
// case 2 starts from a union containing `any`, and cases 6 and 7 assert to NonNullable<T>, which
// removes nullish from a type ALIAS rather than from the expression, so the assertion is doing real
// work a `!` would not do.
func TestNonNullableTypeAssertionStyleStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\ndeclare const original: number | string;\nconst cast = original as string;\n    ",
		"\ndeclare const original: number | undefined;\nconst cast = original as string | number | undefined;\n    ",
		"\ndeclare const original: number | any;\nconst cast = original as string | number | undefined;\n    ",
		"\ndeclare const original: number | undefined;\nconst cast = original as any;\n    ",
		"\ndeclare const original: number | null | undefined;\nconst cast = original as number | null;\n    ",
		"\ntype Type = { value: string };\ndeclare const original: Type | number;\nconst cast = original as Type;\n    ",
		"\ntype T = string;\ndeclare const x: T | number;\n\nconst y = x as NonNullable<T>;\n    ",
		"\ntype T = string | null;\ndeclare const x: T | number;\n\nconst y = x as NonNullable<T>;\n    ",
		"\nconst foo = [] as const;\n    ",
		"\nconst x = 1 as 1;\n    ",
		"\ndeclare function foo<T = any>(): T;\nconst bar = foo() as number;\n    ",
	}
	for index, sourceText := range cases {
		t.Run(nonNullableTypeAssertionStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NonNullableTypeAssertionStyle,
				nonNullableTypeAssertionStyleFile, sourceText))
		})
	}
}

// TestNonNullableTypeAssertionStyleFiresOnUpstreamFailCases is the imported failing corpus,
// verbatim, with the span AND the applied repair asserted on every row.
//
// The repair is half this port: meta.fixable is "code", the fix is applied unattended, and a fixer
// that repairs the right span with the wrong text satisfies every message id assertion. Upstream
// records an `output` on all nine of these, which is the specification for what the rewrite must
// write, so all nine are asserted through ExpectFixedSource rather than by eye.
//
// The spans come from replaying each case through the installed build one file per program, because
// upstream records only a start column on this rule and a start column cannot see a finding anchored
// on the wrong node that happens to begin in the right place.
func TestNonNullableTypeAssertionStyleFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
		wantFixed  string
	}{
		{
			sourceText: "\ndeclare const maybe: string | undefined;\nconst bar = maybe as string;\n      ",
			wantSpan:   "maybe as string",
			wantFixed:  "\ndeclare const maybe: string | undefined;\nconst bar = maybe!;\n      ",
		},
		{
			sourceText: "\ndeclare const maybe: string | null;\nconst bar = maybe as string;\n      ",
			wantSpan:   "maybe as string",
			wantFixed:  "\ndeclare const maybe: string | null;\nconst bar = maybe!;\n      ",
		},
		{
			sourceText: "\ndeclare const maybe: string | null | undefined;\nconst bar = maybe as string;\n      ",
			wantSpan:   "maybe as string",
			wantFixed:  "\ndeclare const maybe: string | null | undefined;\nconst bar = maybe!;\n      ",
		},
		{
			sourceText: "\ntype Type = { value: string };\ndeclare const maybe: Type | undefined;\nconst bar = maybe as Type;\n      ",
			wantSpan:   "maybe as Type",
			wantFixed:  "\ntype Type = { value: string };\ndeclare const maybe: Type | undefined;\nconst bar = maybe!;\n      ",
		},
		{
			sourceText: "\ninterface Interface {\n  value: string;\n}\ndeclare const maybe: Interface | undefined;\nconst bar = maybe as Interface;\n      ",
			wantSpan:   "maybe as Interface",
			wantFixed:  "\ninterface Interface {\n  value: string;\n}\ndeclare const maybe: Interface | undefined;\nconst bar = maybe!;\n      ",
		},
		{
			sourceText: "\ntype T = string | null;\ndeclare const x: T;\n\nconst y = x as NonNullable<T>;\n      ",
			wantSpan:   "x as NonNullable<T>",
			wantFixed:  "\ntype T = string | null;\ndeclare const x: T;\n\nconst y = x!;\n      ",
		},
		{
			sourceText: "\ntype T = string | null | undefined;\ndeclare const x: T;\n\nconst y = x as NonNullable<T>;\n      ",
			wantSpan:   "x as NonNullable<T>",
			wantFixed:  "\ntype T = string | null | undefined;\ndeclare const x: T;\n\nconst y = x!;\n      ",
		},
		{
			sourceText: "\ndeclare function nullablePromise(): Promise<string | null>;\n\nasync function fn(): Promise<string> {\n  return (await nullablePromise()) as string;\n}\n      ",
			wantSpan:   "(await nullablePromise()) as string",
			wantFixed:  "\ndeclare function nullablePromise(): Promise<string | null>;\n\nasync function fn(): Promise<string> {\n  return (await nullablePromise())!;\n}\n      ",
		},
		{
			sourceText: "\ndeclare const a: string | null;\n\nconst b = (a || undefined) as string;\n      ",
			wantSpan:   "(a || undefined) as string",
			wantFixed:  "\ndeclare const a: string | null;\n\nconst b = (a || undefined)!;\n      ",
		},
	}
	for index, testCase := range cases {
		t.Run(nonNullableTypeAssertionStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NonNullableTypeAssertionStyle,
				nonNullableTypeAssertionStyleFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "preferNonNullAssertion")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			reported := result.Diagnostics[0]
			written := asTheTypedHarnessWroteIt(testCase.sourceText)
			gotSpan := written[reported.Range.Pos():reported.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Errorf("the finding points at %q, wanted %q", gotSpan, testCase.wantSpan)
			}
			rule_testing.ExpectFixedSource(t, result, asTheTypedHarnessWroteIt(testCase.wantFixed))
		})
	}
}

// TestNonNullableTypeAssertionStyleOnShapesUpstreamsCorpusDoesNotWrite covers the repair decisions
// and the type-parameter branches upstream's eleven-and-nine corpus leaves unexercised.
//
// Every expected verdict and every expected repair below was measured by running the installed 8.x
// build on that exact source in its own program, alongside a control that reported, so a silent row
// is a verdict rather than a harness that never ran. The whole point of the table is the fix column:
// upstream's corpus writes parentheses in only two of its nine failing rows and never writes a bare
// call, a bare await, or a redundant parenthesis, so the wrapping decision is almost entirely
// untested by the imported cases.
func TestNonNullableTypeAssertionStyleOnShapesUpstreamsCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantSpan   string
		wantFixed  string
	}{
		{
			name:       "bare-await-wraps",
			why:        "an await written without parentheses is precedence 16, below unary, so the repair has to ADD them; this is the input on which reading our parenthesized node directly and reading upstream's inner node agree, and it pins the wrapping half of the decision",
			sourceText: "declare function p(): Promise<string | null>;\nasync function fn(): Promise<string> {\n  return await p() as string;\n}\n",
			wantSpan:   "await p() as string",
			wantFixed:  "declare function p(): Promise<string | null>;\nasync function fn(): Promise<string> {\n  return (await p())!;\n}\n",
		},
		{
			name:       "angle-bracket-await-wraps",
			why:        "the same wrapping decision reached through the angle-bracket spelling, whose listener is a separate arm here",
			sourceText: "declare function p(): Promise<string | null>;\nasync function fn(): Promise<string> {\n  return <string>await p();\n}\n",
			wantSpan:   "<string>await p()",
			wantFixed:  "declare function p(): Promise<string | null>;\nasync function fn(): Promise<string> {\n  return (await p())!;\n}\n",
		},
		{
			name:       "parenthesized-conditional",
			why:        "parentheses already in the source around a conditional, which must survive into the output or the repair changes what the code means",
			sourceText: "declare const c: boolean;\ndeclare const a: string | null;\nconst b = (c ? a : a) as string;\n",
			wantSpan:   "(c ? a : a) as string",
			wantFixed:  "declare const c: boolean;\ndeclare const a: string | null;\nconst b = (c ? a : a)!;\n",
		},
		{
			name:       "comma-expression",
			why:        "the same for a comma expression, where dropping the parentheses would change the meaning of the surrounding assignment",
			sourceText: "declare const a: string | null;\nconst b = (0, a) as string;\n",
			wantSpan:   "(0, a) as string",
			wantFixed:  "declare const a: string | null;\nconst b = (0, a)!;\n",
		},
		{
			name:       "redundant-parens-around-identifier",
			why:        "the ONE shape where reading our node directly disagrees with upstream: upstream folds the redundant parentheses away and writes a!, and without the unwrap in the rule this writes (a)!",
			sourceText: "declare const a: string | null;\nconst b = (a) as string;\n",
			wantSpan:   "(a) as string",
			wantFixed:  "declare const a: string | null;\nconst b = a!;\n",
		},
		{
			name:       "call-expression",
			why:        "a call is precedence 20, above unary, so nothing is added; the corpus writes no bare call and this is the row that separates the wrap branch from the no-wrap one on an unparenthesized input",
			sourceText: "declare const o: { m(): string | null };\nconst b = o.m() as string;\n",
			wantSpan:   "o.m() as string",
			wantFixed:  "declare const o: { m(): string | null };\nconst b = o.m()!;\n",
		},
		{
			name:       "unconstrained-type-parameter",
			why:        "an unconstrained type parameter could be instantiated with null by a caller, so couldBeNullish answers true on a nil constraint and the rule declines",
			sourceText: "function f<T>(x: T) {\n  return x as string;\n}\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "constrained-nullable-type-parameter",
			why:        "the mirror of the row above, kept beside it so the nil-constraint branch is a measurement rather than an argument",
			sourceText: "function f<T extends string | null>(x: T) {\n  return x as string;\n}\n",
			wantSpan:   "x as string",
			wantFixed:  "function f<T extends string | null>(x: T) {\n  return x!;\n}\n",
		},
		{
			name:       "const-assertion-on-nullable",
			why:        "a const assertion on a nullable expression, where the asserted type IS the original type, so every part of the type comparison agrees and only the syntactic guard declines it",
			sourceText: "declare const a: string | null;\nconst b = a as const;\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "angle-bracket-plain",
			why:        "the plain angle-bracket spelling, proving the second listener reports rather than only the as spelling",
			sourceText: "declare const a: string | null;\nconst b = <string>a;\n",
			wantSpan:   "<string>a",
			wantFixed:  "declare const a: string | null;\nconst b = a!;\n",
		},
		{
			name:       "control-clean",
			why:        "the control: nothing nullish in the original, so a rule that had stopped comparing types at all would still be visible here",
			sourceText: "declare const a: string;\nconst b = a as string;\n",
			wantSpan:   "",
			wantFixed:  "",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NonNullableTypeAssertionStyle,
				nonNullableTypeAssertionStyleFile, testCase.sourceText)
			if testCase.wantSpan == "" {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "preferNonNullAssertion")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d, on the case that covers: %s",
					len(result.Diagnostics), testCase.why)
			}
			reported := result.Diagnostics[0]
			written := asTheTypedHarnessWroteIt(testCase.sourceText)
			gotSpan := written[reported.Range.Pos():reported.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Errorf("the finding points at %q, wanted %q (%s)",
					gotSpan, testCase.wantSpan, testCase.why)
			}
			rule_testing.ExpectFixedSource(t, result, asTheTypedHarnessWroteIt(testCase.wantFixed))
		})
	}
}

// TestNonNullableTypeAssertionStyleRendersUpstreamsMessageText asserts what a reader is told.
//
// rule.Message is {Id, Description} with no interpolation, so there is nothing to render and nothing
// a format string could get wrong; what there is to get wrong is the text, and every other fixture
// here goes through ExpectFindings, which compares ids and count and nothing else. The wanted string
// is typed as a literal rather than read from the rule's own constant, because a comparison against
// the constant moves with any mutation of it.
func TestNonNullableTypeAssertionStyleRendersUpstreamsMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NonNullableTypeAssertionStyle,
		nonNullableTypeAssertionStyleFile,
		"declare const maybe: string | undefined;\nconst bar = maybe as string;\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	reported := result.Diagnostics[0]
	if reported.Message.Id != "preferNonNullAssertion" {
		t.Errorf("reported id %q, wanted %q", reported.Message.Id, "preferNonNullAssertion")
	}
	const wantMessage = "Use a ! assertion to more succinctly remove null and undefined from the type."
	if reported.Message.Description != wantMessage {
		t.Errorf("reported message %q, wanted %q", reported.Message.Description, wantMessage)
	}
}

// TestNonNullableTypeAssertionStyleRequiresTheTypedHarness pins the checker guard.
//
// Every listener starts with a nil check, and the shim answers nil from a type query on a nil
// checker rather than panicking, so a rule missing that guard does not crash, it goes SILENT. A
// vacuous green is the more dangerous of the two failures because nothing announces it, so this
// asserts the untyped harness produces no finding on an input the typed one reports, which makes a
// later revert to rule_testing.Run fail loudly rather than quietly.
func TestNonNullableTypeAssertionStyleRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "declare const maybe: string | undefined;\nconst bar = maybe as string;\n"
	rule_testing.ExpectClean(t, rule_testing.Run(t, NonNullableTypeAssertionStyle,
		nonNullableTypeAssertionStyleFile, sourceText))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NonNullableTypeAssertionStyle,
		nonNullableTypeAssertionStyleFile, sourceText), "preferNonNullAssertion")
}

// TestNonNullableTypeAssertionStyleOnTypeParametersAndPartialUnions covers the branches upstream's
// corpus leaves untouched because of what it never writes.
//
// Upstream asserts FROM a type parameter twice and TO one never, so the whole couldBeNullish
// recursion is invisible to the imported cases; and every one of its failing rows asserts to exactly
// the original minus nullish, so neither half of the two-way membership comparison is separable from
// the other there. Four mutations survived the imported corpus and these rows are what kill them.
//
// Every verdict and every repair was measured against the installed 8.x build on that exact source in
// its own program, with a reporting control in the same run.
func TestNonNullableTypeAssertionStyleOnTypeParametersAndPartialUnions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantSpan   string
		wantFixed  string
	}{
		{
			name:       "assert-to-unconstrained-parameter",
			why:        "asserting TO an unconstrained type parameter: a caller could instantiate it with null, so couldBeNullish answers true on a nil constraint and the rule declines. Upstream's whole corpus asserts FROM a type parameter and never to one, so its nine failing and eleven passing rows leave this branch untouched",
			sourceText: "function f<T>(x: T | null) {\n  return x as T;\n}\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "assert-to-non-nullable-constrained-parameter",
			why:        "the same shape with a constraint that cannot be null, which REPORTS; this is the row that separates the type-parameter branch from a version that always answered nullish",
			sourceText: "function f<T extends string>(x: T | null) {\n  return x as T;\n}\n",
			wantSpan:   "x as T",
			wantFixed:  "function f<T extends string>(x: T | null) {\n  return x!;\n}\n",
		},
		{
			name:       "assert-to-nullable-constrained-parameter",
			why:        "a constraint that CAN be null, which recurses one level and declines; the row above and this one bracket the recursion from both sides",
			sourceText: "function f<T extends string | null>(x: T | undefined) {\n  return x as T;\n}\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "assert-to-parameter-union",
			why:        "a union of two parameters where one constraint is nullable, declining through the union arm of couldBeNullish rather than through its type-parameter arm",
			sourceText: "function f<T extends string | null, U>(x: T | U | null) {\n  return x as T | U;\n}\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "assert-to-two-non-nullable-parameters",
			why:        "the mirror where both constraints are non-nullable and the whole union reports, so the union arm is measured in both directions",
			sourceText: "function f<T extends string, U extends number>(x: T | U | null) {\n  return x as T | U;\n}\n",
			wantSpan:   "x as T | U",
			wantFixed:  "function f<T extends string, U extends number>(x: T | U | null) {\n  return x!;\n}\n",
		},
		{
			name:       "assert-narrower-than-original",
			why:        "the asserted type drops a non-nullish constituent as well as the nullish one, so a ! would not produce the same type and the rule declines; this is the every-non-nullish-original-is-asserted clause and nothing in upstream's corpus exercises it alone",
			sourceText: "declare const a: string | number | null;\nconst b = a as string;\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "assert-wider-than-original",
			why:        "the asserted type ADDS a constituent the original never had, which the every-asserted-is-an-original clause declines",
			sourceText: "declare const a: string | null;\nconst b = a as string | number;\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "control-tp",
			why:        "the control: the plainest reporting shape, so a rule that had stopped reporting entirely would still be visible in this table",
			sourceText: "declare const a: string | null;\nconst b = a as string;\n",
			wantSpan:   "a as string",
			wantFixed:  "declare const a: string | null;\nconst b = a!;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NonNullableTypeAssertionStyle,
				nonNullableTypeAssertionStyleFile, testCase.sourceText)
			if testCase.wantSpan == "" {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "preferNonNullAssertion")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d, on the case that covers: %s",
					len(result.Diagnostics), testCase.why)
			}
			reported := result.Diagnostics[0]
			written := asTheTypedHarnessWroteIt(testCase.sourceText)
			gotSpan := written[reported.Range.Pos():reported.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Errorf("the finding points at %q, wanted %q (%s)",
					gotSpan, testCase.wantSpan, testCase.why)
			}
			rule_testing.ExpectFixedSource(t, result, asTheTypedHarnessWroteIt(testCase.wantFixed))
		})
	}
}

// TestNonNullableTypeAssertionStyleOnNullishFlagsAndUnionAssertions covers the two branches of
// couldBeNullish that upstream's corpus cannot reach.
//
// couldBeNullish is asked only about ASSERTED types, and every asserted type in upstream's twenty
// cases is a single non-nullish type, so neither its union arm nor its leaf flag test decides
// anything there. Both can be neutralized with all twenty imported cases staying green.
//
// The route to them is narrower than it first looks, and the first fixtures written for it did not
// work. unionConstituentsIfNotLoose splits the asserted type BEFORE couldBeNullish is called, so the
// top-level call never receives a union at all: asserting to `string | undefined` hands the leaf test
// a bare `undefined`, and both survivors lived through fixtures built on that shape. The union arm is
// reachable only by recursing into a type parameter's CONSTRAINT, and only a constraint carrying no
// nullish part runs the loop to completion and reaches its false return. The rows below are built on
// that shape instead, and they are what killed both mutants.
//
// The leaf test is the one that matters in practice: without TypeFlagsUndefined this rule rewrites
// `x as T` into `x!` where T is constrained to `string | undefined`, widening the type unattended,
// because meta.fixable is code. Measured against the installed 8.x build with controls in the same
// runs.
func TestNonNullableTypeAssertionStyleOnNullishFlagsAndUnionAssertions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantSpan   string
		wantFixed  string
	}{
		{
			name:       "union-asserted-no-nullish",
			why:        "the asserted type is a UNION with no nullish part, so couldBeNullish walks its constituents and reaches the union arm's false return. Nothing in upstream's corpus asserts to a multi-constituent union that reports, so that return can be rewritten to true and every imported case stays green",
			sourceText: "declare const a: string | number | undefined;\nconst b = a as string | number;\n",
			wantSpan:   "a as string | number",
			wantFixed:  "declare const a: string | number | undefined;\nconst b = a!;\n",
		},
		{
			name:       "union-asserted-no-nullish-null",
			why:        "the same through null rather than undefined, so the union arm is measured against both nullish flags",
			sourceText: "declare const a: string | number | null;\nconst b = a as string | number;\n",
			wantSpan:   "a as string | number",
			wantFixed:  "declare const a: string | number | null;\nconst b = a!;\n",
		},
		{
			name:       "undefined-only-simple",
			why:        "the simplest undefined case, kept beside the two rows above as the shape they are variations of",
			sourceText: "declare const a: string | undefined;\nconst b = a as string;\n",
			wantSpan:   "a as string",
			wantFixed:  "declare const a: string | undefined;\nconst b = a!;\n",
		},
		{
			name:       "void-in-union",
			why:        "void is NOT nullish for this rule and a union containing it is left alone, which is the boundary of the leaf flag test on the far side from null and undefined",
			sourceText: "declare const a: string | void;\nconst b = a as string;\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "control-n",
			why:        "the control for this group",
			sourceText: "declare const a: string | null;\nconst b = a as string;\n",
			wantSpan:   "a as string",
			wantFixed:  "declare const a: string | null;\nconst b = a!;\n",
		},
		{
			name:       "asserted-contains-undefined",
			why:        "the ASSERTED type contains undefined, which is the only arrangement where the leaf flag test decides anything: drop TypeFlagsUndefined from it and this rewrites a as string | undefined into a!, silently changing the type. Every asserted type in upstream's corpus is nullish-free, so the corpus cannot separate the two flags",
			sourceText: "declare const a: string | number | undefined;\nconst b = a as string | undefined;\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "asserted-is-undefined",
			why:        "the degenerate version of the row above where the asserted type is undefined and nothing else",
			sourceText: "declare const a: string | undefined;\nconst b = a as undefined;\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "parameter-constrained-to-non-nullish-union",
			why:        "the ONLY arrangement that reaches the union arm's false return. unionConstituentsIfNotLoose has already split the asserted type before couldBeNullish sees it, so the top-level call never receives a union; the arm is reached only by recursing into a type parameter's CONSTRAINT, and only a constraint with no nullish part gets past the loop to the return",
			sourceText: "function f<T extends string | number>(x: T | null) {\n  return x as T;\n}\n",
			wantSpan:   "x as T",
			wantFixed:  "function f<T extends string | number>(x: T | null) {\n  return x!;\n}\n",
		},
		{
			name:       "parameter-constrained-to-undefined-union",
			why:        "the same route with undefined in the constraint, which is the only way the leaf flag test is ever asked about undefined. Drop TypeFlagsUndefined and this reports and rewrites x as T into x!, widening the type silently",
			sourceText: "function f<T extends string | undefined>(x: T | null) {\n  return x as T;\n}\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "parameter-constrained-to-null-union",
			why:        "the null half of the row above, so the leaf test is measured on both flags through the same route",
			sourceText: "function f<T extends string | null>(x: T | undefined) {\n  return x as T;\n}\n",
			wantSpan:   "",
			wantFixed:  "",
		},
		{
			name:       "control-n17",
			why:        "the control for this group",
			sourceText: "declare const a: string | null;\nconst b = a as string;\n",
			wantSpan:   "a as string",
			wantFixed:  "declare const a: string | null;\nconst b = a!;\n",
		},
		{
			name:       "control-n18",
			why:        "the control for this group, differing from the first row only in that its asserted union drops the undefined",
			sourceText: "declare const a: string | number | undefined;\nconst b = a as string | number;\n",
			wantSpan:   "a as string | number",
			wantFixed:  "declare const a: string | number | undefined;\nconst b = a!;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NonNullableTypeAssertionStyle,
				nonNullableTypeAssertionStyleFile, testCase.sourceText)
			if testCase.wantSpan == "" {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "preferNonNullAssertion")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d, on the case that covers: %s",
					len(result.Diagnostics), testCase.why)
			}
			reported := result.Diagnostics[0]
			written := asTheTypedHarnessWroteIt(testCase.sourceText)
			gotSpan := written[reported.Range.Pos():reported.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Errorf("the finding points at %q, wanted %q (%s)",
					gotSpan, testCase.wantSpan, testCase.why)
			}
			rule_testing.ExpectFixedSource(t, result, asTheTypedHarnessWroteIt(testCase.wantFixed))
		})
	}
}
