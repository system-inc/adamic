package typescript

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// requireArraySortCompareFile names the fixture file. The rule reads no path and gates on no
// extension.
const requireArraySortCompareFile = "/repository/source/Sorting.ts"

func requireArraySortCompareCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// requireArraySortCompareOptionsFor routes a case's options through the rule's OWN decoder.
//
// This is the line that matters most in this file. Building the options struct directly would leave
// the decoder untested, and this rule's decoder is the one most likely to be wrong: the single
// option DEFAULTS TO TRUE, so a generic `DecodeOptionsInto` would hand back a zero-value struct that
// silently inverts the rule, and every fixture built from a struct would pass anyway.
//
// An empty specifier means the case ran with no options at all, which is upstream's default and is
// also what cohere hands a rule configured as a bare "error".
func requireArraySortCompareOptionsFor(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return DefaultRequireArraySortCompareSettings()
	}
	decoded, err := DecodeRequireArraySortCompareOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding %s: %v", optionsJson, err)
	}
	return decoded
}

// TestRequireArraySortCompareStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All twenty of upstream's passing inputs, extracted from the clone's test file by parsing it with
// the TypeScript compiler. Every one was additionally run through the installed 8.x build against a
// real program, which reported nothing and produced no parse error on any of them.
//
// Seven of the twenty carry explicit options, and both settings of the single option appear, so the
// exemption is exercised in both directions rather than only at its default.
func TestRequireArraySortCompareStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
	}{
		{
			sourceText:  "\nfunction f(a: any[]) {\n  a.sort(undefined);\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\nfunction f(a: any[]) {\n  a.sort((a, b) => a - b);\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\nfunction f(a: Array<string>) {\n  a.sort(undefined);\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\nfunction f(a: Array<number>) {\n  a.sort((a, b) => a - b);\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\nfunction f(a: { sort(): void }) {\n  a.sort();\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\nclass A {\n  sort(): void {}\n}\nfunction f(a: A) {\n  a.sort();\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\ninterface A {\n  sort(): void;\n}\nfunction f(a: A) {\n  a.sort();\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\ninterface A {\n  sort(): void;\n}\nfunction f<T extends A>(a: T) {\n  a.sort();\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\nfunction f(a: any) {\n  a.sort();\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\nnamespace UserDefined {\n  interface Array {\n    sort(): void;\n  }\n  function f(a: Array) {\n    a.sort();\n  }\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\nfunction f(a: any[]) {\n  a?.sort((a, b) => a - b);\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\nnamespace UserDefined {\n  interface Array {\n    sort(): void;\n  }\n  function f(a: Array) {\n    a?.sort();\n  }\n}\n    ",
			optionsJson: "",
		},
		{
			sourceText:  "\n['foo', 'bar', 'baz'].sort();\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
		},
		{
			sourceText:  "\nfunction getString() {\n  return 'foo';\n}\n[getString(), getString()].sort();\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
		},
		{
			sourceText:  "\nconst foo = 'foo';\nconst bar = 'bar';\nconst baz = 'baz';\n[foo, bar, baz].sort();\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
		},
		{
			sourceText:  "\ndeclare const x: string[];\nx.sort();\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
		},
		{
			sourceText:  "\nfunction f<T extends string[]>(a: T) {\n  a.sort();\n}\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
		},
		{
			sourceText:  "\nfunction f<T extends Array<string>>(a: T) {\n  a.sort();\n}\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
		},
		{
			sourceText:  "\nfunction f<T extends string>(a: T[]) {\n  a.sort();\n}\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
		},
		{
			sourceText:  "\nfunction f(a: number[]) {\n  a.toSorted((a, b) => a - b);\n}\n      ",
			optionsJson: "",
		},
	}
	for index, testCase := range cases {
		t.Run(requireArraySortCompareCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, RequireArraySortCompare,
				requireArraySortCompareFile, testCase.sourceText,
				requireArraySortCompareOptionsFor(t, testCase.optionsJson)))
		})
	}
}

// TestRequireArraySortCompareFiresOnUpstreamFailCases is the imported reporting corpus, verbatim.
//
// All seventeen of upstream's failing inputs, each carrying exactly one finding and no `output`
// field, since the rule ships no repair.
//
// The span is measured rather than transcribed, and it matters here: upstream reports
// `callee.parent`, which is the whole CALL expression rather than the member access, so the finding
// underlines `a.sort()` with its parentheses. Both readings are defensible and the message id
// cannot tell them apart.
//
// The harness writes each fixture as `strings.TrimSpace(source)+"\n"`, so every span assertion here
// slices the TRIMMED text rather than the Go literal. Every case in this corpus carries a leading
// newline, so comparing against the literal would be off by one byte on all seventeen and would read
// exactly like an off-by-one in the rule.
func TestRequireArraySortCompareFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
		wantSpan    string
	}{
		{
			sourceText:  "\nfunction f(a: Array<any>) {\n  a.sort();\n}\n      ",
			optionsJson: "",
			wantSpan:    "a.sort()",
		},
		{
			sourceText:  "\nfunction f(a: number[]) {\n  a.sort();\n}\n      ",
			optionsJson: "",
			wantSpan:    "a.sort()",
		},
		{
			sourceText:  "\nfunction f(a: number[]) {\n  a.sort();\n}\n      ",
			optionsJson: "{\"ignoreStringArrays\": false}",
			wantSpan:    "a.sort()",
		},
		{
			sourceText:  "\nfunction f(a: number | number[]) {\n  if (Array.isArray(a)) a.sort();\n}\n      ",
			optionsJson: "",
			wantSpan:    "a.sort()",
		},
		{
			sourceText:  "\nfunction f(a: string | string[]) {\n  if (Array.isArray(a)) a.sort();\n}\n      ",
			optionsJson: "{\"ignoreStringArrays\": false}",
			wantSpan:    "a.sort()",
		},
		{
			sourceText:  "\nfunction f(a: number[] | string[]) {\n  a.sort();\n}\n      ",
			optionsJson: "",
			wantSpan:    "a.sort()",
		},
		{
			sourceText:  "\nfunction f<T extends number[]>(a: T) {\n  a.sort();\n}\n      ",
			optionsJson: "",
			wantSpan:    "a.sort()",
		},
		{
			sourceText:  "\nfunction f<T extends string[]>(a: T) {\n  a.sort();\n}\n      ",
			optionsJson: "{\"ignoreStringArrays\": false}",
			wantSpan:    "a.sort()",
		},
		{
			sourceText:  "\nfunction f<T, U extends T[]>(a: U) {\n  a.sort();\n}\n      ",
			optionsJson: "",
			wantSpan:    "a.sort()",
		},
		{
			sourceText:  "\nfunction f(a: number[]) {\n  a?.sort();\n}\n      ",
			optionsJson: "",
			wantSpan:    "a?.sort()",
		},
		{
			sourceText:  "\n[1, 2, 3].sort();\n      ",
			optionsJson: "",
			wantSpan:    "[1, 2, 3].sort()",
		},
		{
			sourceText:  "\nfunction getNumber() {\n  return 1;\n}\n[getNumber(), getNumber()].sort();\n      ",
			optionsJson: "",
			wantSpan:    "[getNumber(), getNumber()].sort()",
		},
		{
			sourceText:  "\nconst foo = 1;\nconst bar = 2;\nconst baz = 3;\n[foo, bar, baz].sort();\n      ",
			optionsJson: "",
			wantSpan:    "[foo, bar, baz].sort()",
		},
		{
			sourceText:  "\n[2, 'bar', 'baz'].sort();\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
			wantSpan:    "[2, 'bar', 'baz'].sort()",
		},
		{
			sourceText:  "\nfunction getNumber() {\n  return 2;\n}\n[2, 3].sort();\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
			wantSpan:    "[2, 3].sort()",
		},
		{
			sourceText:  "\nconst one = 1;\nconst two = 2;\nconst three = 3;\n[one, two, three].sort();\n      ",
			optionsJson: "{\"ignoreStringArrays\": true}",
			wantSpan:    "[one, two, three].sort()",
		},
		{
			sourceText:  "\nfunction f(a: number[]) {\n  a.toSorted();\n}\n      ",
			optionsJson: "",
			wantSpan:    "a.toSorted()",
		},
	}
	for index, testCase := range cases {
		t.Run(requireArraySortCompareCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, RequireArraySortCompare,
				requireArraySortCompareFile, testCase.sourceText,
				requireArraySortCompareOptionsFor(t, testCase.optionsJson))
			rule_testing.ExpectFindings(t, result, "requireCompare")

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]
			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}

			if diagnostic.Message.Description != "Require 'compare' argument." {
				t.Fatalf("message: got %q", diagnostic.Message.Description)
			}
			if diagnostic.Message.Id != "requireCompare" {
				t.Fatalf("message id: got %q", diagnostic.Message.Id)
			}
			if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
				t.Fatalf("expected no repair, got %d fixes and %d suggestions",
					len(diagnostic.Fixes), len(diagnostic.Suggestions))
			}
		})
	}
}

// TestDecodeRequireArraySortCompareOptionsKeepsTheDefaultWhenTheKeyIsAbsent pins the inversion trap.
//
// This rule's single option DEFAULTS TO TRUE, which makes the generic `rule.DecodeOptionsInto`
// actively wrong here rather than merely insufficient: it yields a zero-value struct, so an absent
// key would read as `ignoreStringArrays: false` and the rule would report every string array
// upstream deliberately exempts.
//
// The failure is invisible from the corpus. Thirteen of upstream's thirty-seven cases name the
// option explicitly and would keep passing, and the twenty-four that do not name it are mostly
// non-string arrays whose verdict the option cannot change. Only a string array with no options at
// all separates the two decoders, which is the third case below.
func TestDecodeRequireArraySortCompareOptionsKeepsTheDefaultWhenTheKeyIsAbsent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "emptyObject", raw: `{}`, want: true},
		{name: "explicitTrue", raw: `{"ignoreStringArrays": true}`, want: true},
		{name: "explicitFalse", raw: `{"ignoreStringArrays": false}`, want: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeRequireArraySortCompareOptions(json.RawMessage(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.raw, err)
			}
			options, isOptions := decoded.(RequireArraySortCompareOptions)
			if !isOptions {
				t.Fatalf("the decoder returned %T rather than the options struct", decoded)
			}
			if options.IgnoreStringArrays != testCase.want {
				t.Fatalf("ignoreStringArrays: expected %v, got %v", testCase.want, options.IgnoreStringArrays)
			}
		})
	}
}

// TestRequireArraySortCompareFallsBackToTheDefaultOnNilOptions pins the OTHER half of the trap.
//
// A rule configured as a bare "error" is handed nil options, and `options.(T)` on nil yields the
// zero value rather than failing, which for this rule is the inverse of upstream's default. Every
// fixture above reaches the rule through the decoder, so none of them can see this; the rule's own
// fallback is what makes a plainly-configured rule behave as documented.
//
// A string array with no options is the separating input. Under the default it is exempt; under a
// zeroed struct it reports.
func TestRequireArraySortCompareFallsBackToTheDefaultOnNilOptions(t *testing.T) {
	t.Parallel()

	const stringArray = "['foo', 'bar', 'baz'].sort();"

	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, RequireArraySortCompare,
		requireArraySortCompareFile, stringArray, nil))

	// The control: the same input with the exemption explicitly off must report, so the clean
	// verdict above is the option working rather than the rule being inert.
	reported := rule_testing.RunTypedWithOptions(t, RequireArraySortCompare,
		requireArraySortCompareFile, stringArray,
		RequireArraySortCompareOptions{IgnoreStringArrays: false})
	rule_testing.ExpectFindings(t, reported, "requireCompare")
}

// TestRequireArraySortCompareNeedsTheTypedHarness pins the checker declaration.
//
// Under the plain harness the checker is nil and this rule returns immediately, so every clean
// fixture would pass vacuously and every reporting one would fail in a way that reads like a rule
// defect. Asserting the declaration directly means a later revert fails loudly rather than going
// quietly green.
func TestRequireArraySortCompareNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !RequireArraySortCompare.NeedsTypeChecker {
		t.Fatal("the rule resolves the receiver's type, so it must declare NeedsTypeChecker")
	}

	// The guard's other half: handed no checker, the rule must stay silent rather than panic.
	rule_testing.ExpectClean(t, rule_testing.Run(t, RequireArraySortCompare,
		requireArraySortCompareFile, "function f(a: number[]) {\n  a.sort();\n}\n"))
}

// TestRequireArraySortCompareStaysSilentOnShapesTheCorpusDoesNotWrite covers clean divergences.
//
// Every verdict was measured against the installed 8.x build over a real program. The tuple pair is
// the one worth reading twice: an all-string tuple is exempt through the exemption's tuple branch,
// while a mixed tuple is silent for a DIFFERENT reason, that a tuple is not an array type and so
// fails the report test. Two clean verdicts, two mechanisms, and a port could get one right by
// accident while breaking the other.
func TestRequireArraySortCompareStaysSilentOnShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{
			// An explicit undefined behaves identically at run time and is silent, because the argument count is part of upstream's selector.
			sourceText: "function f(a: number[]) {\n  a.sort(undefined);\n}",
		},
		{
			// An all-string tuple is exempt under the default option, through the tuple branch of the exemption.
			sourceText: "function f(a: [string, string]) {\n  a.sort();\n}",
		},
		{
			// A mixed tuple fails the exemption, then fails the array test too, since a tuple is not an array type. Silent for the second reason rather than the first.
			sourceText: "function f(a: [string, number]) {\n  a.sort();\n}",
		},
		{
			// A plain string array, exempt under the default.
			sourceText: "function f(a: string[]) {\n  a.sort();\n}",
		},
		{
			// Computed keys that are not string literals. Upstream folds a key only when it can
			// read a static value, and none of these is `sort`, so all four are clean there too.
			//
			// These exist because a mutant accepting EVERY computed key kind survived the whole
			// corpus. `a[0]()` is the one that separates the two versions without crashing: a
			// numeric literal's Text() is "0", which is simply not a member name the rule wants,
			// so the mutant reports a call that has nothing to do with sorting.
			sourceText: "function f(a: number[], i: number) {\n  a[i]();\n}",
		},
		{
			sourceText: "function f(a: number[]) {\n  a[0]();\n}",
		},
		{
			// A template literal key. Upstream reads it as a static value and it is not `sort`
			// either way, and our port declines it on the kind test.
			sourceText: "function f(a: any) {\n  a[`sort`]();\n}",
		},
		{
			// A symbol-valued key. Upstream's accessor can return a symbol and compares it against
			// the two names, so this is clean; here the kind test declines it before any of that.
			sourceText: "function f(a: number[]) {\n  a[Symbol.iterator]();\n}",
		},
		{
			// A zero-argument call to some OTHER array method. The member name is the only thing
			// declining these, so they are what separates a name test from no name test at all.
			// Added after a mutant that accepted every member name survived the whole corpus:
			// upstream writes no such case, because it never occurred to anyone that a rule about
			// sort might fire on reverse.
			sourceText: "function f(a: number[]) {\n  a.reverse();\n}",
		},
		{
			sourceText: "function f(a: number[]) {\n  a.pop();\n}",
		},
		{
			sourceText: "function f(a: number[]) {\n  a.flat();\n}",
		},
		{
			// A union with a non-array member fails the every-constituent test.
			sourceText: "function f(a: number[] | string) {\n  a.sort();\n}",
		},
	}
	for index, testCase := range cases {
		t.Run(requireArraySortCompareCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, RequireArraySortCompare,
				requireArraySortCompareFile, testCase.sourceText,
				DefaultRequireArraySortCompareSettings()))
		})
	}
}

// TestRequireArraySortCompareFiresOnShapesTheCorpusDoesNotWrite covers reporting divergences.
//
// The corpus writes no computed member access and no parenthesized callee anywhere, so nothing in
// it can see how the member name is read. Both matter: a computed string key must resolve, and a
// parenthesized callee must be seen through, because estree has no parenthesized node and upstream's
// child selector therefore matches straight through one.
//
// That last case is here because the first version of this rule got it wrong from a confident
// argument. The doc comment asserted `(a.sort)()` was silent upstream on the reasoning that a paren
// breaks the child relation. Measured, it reports, and the port was silently missing it.
func TestRequireArraySortCompareFiresOnShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
	}{
		{
			// A string-literal computed key resolves without any scope analysis, so it reports.
			sourceText: "function f(a: number[]) {\n  a['sort']();\n}",
			wantSpan:   "a['sort']()",
		},
		{
			// A parenthesized callee. estree has no paren node, so upstream's child selector sees through it and reports. This case is why the callee is paren-skipped; the first version of this rule reasoned its way to the opposite answer and was wrong.
			sourceText: "function f(a: number[]) {\n  (a.sort)();\n}",
			wantSpan:   "(a.sort)()",
		},
		{
			// The second member name the rule matches.
			sourceText: "function f(a: number[]) {\n  a.toSorted();\n}",
			wantSpan:   "a.toSorted()",
		},
		{
			// toSorted through a computed string key.
			sourceText: "function f(a: number[]) {\n  a['toSorted']();\n}",
			wantSpan:   "a['toSorted']()",
		},
		{
			// A readonly array is still an array to the checker.
			sourceText: "function f(a: readonly number[]) {\n  a.toSorted();\n}",
			wantSpan:   "a.toSorted()",
		},
		{
			// An unknown array is still an array.
			sourceText: "function f(a: unknown[]) {\n  a.sort();\n}",
			wantSpan:   "a.sort()",
		},
		{
			// A never array is an array with no type arguments to check, so the exemption's every-loop is vacuously true on the element side while the array test still reports. It reports because the receiver is not a string array by the exemption's own test.
			sourceText: "function f(a: never[]) {\n  a.sort();\n}",
			wantSpan:   "a.sort()",
		},
		{
			// An optional computed access with a literal key.
			sourceText: "function f(a: number[]) {\n  a?.['sort']();\n}",
			wantSpan:   "a?.['sort']()",
		},
		{
			// A top-level array with an inferred element type.
			sourceText: "const a: number[] = []; a.sort();",
			wantSpan:   "a.sort()",
		},
	}
	for index, testCase := range cases {
		t.Run(requireArraySortCompareCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, RequireArraySortCompare,
				requireArraySortCompareFile, testCase.sourceText,
				DefaultRequireArraySortCompareSettings())
			rule_testing.ExpectFindings(t, result, "requireCompare")

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			gotSpan := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
		})
	}
}

// TestRequireArraySortCompareDoesNotConstantFoldAComputedKey pins the one stated divergence.
//
// Upstream resolves a computed member key through ESLint's `getStaticValue` over the enclosing
// scope, so a key held in a const folds to its value and `a[key]()` reports. That is a
// scope-analysis and constant-folding layer this tree does not have, and building it for one rule
// would be a substrate project rather than a port.
//
// So this is recorded as a FALSE NEGATIVE rather than papered over. It is the safe direction for a
// rule that ships no repair, and it is pinned as a fixture so that if the folding layer ever lands,
// this test fails and tells whoever built it that a rule wants rewiring.
//
// The controls matter as much as the case: the literal-key spelling of the same access DOES report
// here, so the silence below is specifically about folding through a binding rather than about
// computed access being unsupported.
func TestRequireArraySortCompareDoesNotConstantFoldAComputedKey(t *testing.T) {
	t.Parallel()

	const folded = "function f(a: number[]) {\n  const key = 'sort';\n  a[key]();\n}"
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, RequireArraySortCompare,
		requireArraySortCompareFile, folded, DefaultRequireArraySortCompareSettings()))

	const literalKey = "function f(a: number[]) {\n  a['sort']();\n}"
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, RequireArraySortCompare,
		requireArraySortCompareFile, literalKey, DefaultRequireArraySortCompareSettings()),
		"requireCompare")
}
