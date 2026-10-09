package typescript

import (
	"encoding/json"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noInferrableTypesFile is where the fixtures pretend to live.
const noInferrableTypesFile = "/repository/source/NoInferrableTypes.ts"

// noInferrableTypesPointer spells an expected rewrite in a fixture row.
func noInferrableTypesPointer(value string) *string { return &value }

// noInferrableTypesCase is one imported corpus row.
type noInferrableTypesCase struct {
	sourceText string
	options    NoInferrableTypesOptions
	wantIds    []string
	wantOutput *string
}

// runNoInferrableTypes drives one case through the rule's own exported decoder.
func runNoInferrableTypes(t *testing.T, testCase noInferrableTypesCase) rule_testing.Result {
	t.Helper()

	encoded, err := json.Marshal(testCase.options)
	if err != nil {
		t.Fatalf("could not encode options: %v", err)
	}
	decoded, err := DecodeNoInferrableTypesOptions(encoded)
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	return rule_testing.RunWithOptions(t, NoInferrableTypes, noInferrableTypesFile,
		testCase.sourceText, decoded)
}

// The corpus is typescript-eslint's own, extracted mechanically rather than retyped.
//
// `tests/rules/no-inferrable-types.test.ts` (clone, 8.69.0) was loaded with its RuleTester stubbed,
// then every case was replayed against the INSTALLED 8.67.0 rule through the ESLint Linter API. The
// expectations below are what the installed rule answered, and they agree with upstream's own
// annotations exactly: 55 valid cases report nothing, 44 invalid cases report, zero disagree.
//
// The `wantOutput` column came from the same run via `linter.verifyAndFix`, so it is what the real
// pipeline produces rather than a reconstruction. All 44 fixable rows agree with the annotated
// `output` in the corpus file.
func noInferrableTypesCases() []noInferrableTypesCase {
	return []noInferrableTypesCase{
		{"const a = 10n;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = -10n;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = BigInt(10);", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = -BigInt(10);", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = BigInt?.(10);", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = -BigInt?.(10);", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = false;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = true;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = Boolean(null);", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = Boolean?.(null);", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = !0;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = 10;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = +10;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = -10;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = Number('1');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = +Number('1');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = -Number('1');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = Number?.('1');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = +Number?.('1');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = -Number?.('1');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = Infinity;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = +Infinity;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = -Infinity;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = NaN;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = +NaN;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = -NaN;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = null;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = /a/;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = RegExp('a');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = RegExp?.('a');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = new RegExp('a');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = 'str';", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = `str`;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = String(1);", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = String?.(1);", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = Symbol('a');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = Symbol?.('a');", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = undefined;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a = void someValue;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const fn = (a = 5, b = true, c = 'foo') => {};", NoInferrableTypesOptions{}, []string{}, nil},
		{"const fn = function (a = 5, b = true, c = 'foo') {};", NoInferrableTypesOptions{}, []string{}, nil},
		{"function fn(a = 5, b = true, c = 'foo') {}", NoInferrableTypesOptions{}, []string{}, nil},
		{"function fn(a: number, b: boolean, c: string) {}", NoInferrableTypesOptions{}, []string{}, nil},
		{"\nclass Foo {\n  a = 5;\n  b = true;\n  c = 'foo';\n}\n    ", NoInferrableTypesOptions{}, []string{}, nil},
		{"\nclass Foo {\n  readonly a: number = 5;\n}\n    ", NoInferrableTypesOptions{}, []string{}, nil},
		{"\nclass Foo {\n  accessor a = 5;\n}\n    ", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a: any = 5;", NoInferrableTypesOptions{}, []string{}, nil},
		{"const fn = function (a: any = 5, b: any = true, c: any = 'foo') {};", NoInferrableTypesOptions{}, []string{}, nil},
		{"const fn = (a: number = 5, b: boolean = true, c: string = 'foo') => {};", NoInferrableTypesOptions{IgnoreParameters: true}, []string{}, nil},
		{"function fn(a: number = 5, b: boolean = true, c: string = 'foo') {}", NoInferrableTypesOptions{IgnoreParameters: true}, []string{}, nil},
		{"const fn = function (a: number = 5, b: boolean = true, c: string = 'foo') {};", NoInferrableTypesOptions{IgnoreParameters: true}, []string{}, nil},
		{"\nclass Foo {\n  a: number = 5;\n  b: boolean = true;\n  c: string = 'foo';\n}\n      ", NoInferrableTypesOptions{IgnoreProperties: true}, []string{}, nil},
		{"\nclass Foo {\n  accessor a: number = 5;\n}\n      ", NoInferrableTypesOptions{IgnoreProperties: true}, []string{}, nil},
		{"\nclass Foo {\n  a?: number = 5;\n  b?: boolean = true;\n  c?: string = 'foo';\n}\n      ", NoInferrableTypesOptions{}, []string{}, nil},
		{"\nclass Foo {\n  constructor(public a = true) {}\n}\n      ", NoInferrableTypesOptions{}, []string{}, nil},
		{"const a: bigint = 10n;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = 10n;")},
		{"const a: bigint = -10n;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = -10n;")},
		{"const a: bigint = BigInt(10);", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = BigInt(10);")},
		{"const a: bigint = -BigInt(10);", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = -BigInt(10);")},
		{"const a: bigint = BigInt?.(10);", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = BigInt?.(10);")},
		{"const a: bigint = -BigInt?.(10);", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = -BigInt?.(10);")},
		{"const a: boolean = false;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = false;")},
		{"const a: boolean = true;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = true;")},
		{"const a: boolean = Boolean(null);", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = Boolean(null);")},
		{"const a: boolean = Boolean?.(null);", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = Boolean?.(null);")},
		{"const a: boolean = !0;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = !0;")},
		{"const a: number = 10;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = 10;")},
		{"const a: number = +10;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = +10;")},
		{"const a: number = -10;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = -10;")},
		{"const a: number = Number('1');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = Number('1');")},
		{"const a: number = +Number('1');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = +Number('1');")},
		{"const a: number = -Number('1');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = -Number('1');")},
		{"const a: number = Number?.('1');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = Number?.('1');")},
		{"const a: number = +Number?.('1');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = +Number?.('1');")},
		{"const a: number = -Number?.('1');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = -Number?.('1');")},
		{"const a: number = Infinity;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = Infinity;")},
		{"const a: number = +Infinity;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = +Infinity;")},
		{"const a: number = -Infinity;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = -Infinity;")},
		{"const a: number = NaN;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = NaN;")},
		{"const a: number = +NaN;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = +NaN;")},
		{"const a: number = -NaN;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = -NaN;")},
		{"const a: null = null;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = null;")},
		{"const a: RegExp = /a/;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = /a/;")},
		{"const a: RegExp = RegExp('a');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = RegExp('a');")},
		{"const a: RegExp = RegExp?.('a');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = RegExp?.('a');")},
		{"const a: RegExp = new RegExp('a');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = new RegExp('a');")},
		{"const a: string = 'str';", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = 'str';")},
		{"const a: string = `str`;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = `str`;")},
		{"const a: string = String(1);", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = String(1);")},
		{"const a: string = String?.(1);", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = String?.(1);")},
		{"const a: symbol = Symbol('a');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = Symbol('a');")},
		{"const a: symbol = Symbol?.('a');", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = Symbol?.('a');")},
		{"const a: undefined = undefined;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = undefined;")},
		{"const a: undefined = void someValue;", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("const a = void someValue;")},
		{"const fn = (a?: number = 5) => {};", NoInferrableTypesOptions{IgnoreParameters: false}, []string{"noInferrableType"}, noInferrableTypesPointer("const fn = (a = 5) => {};")},
		{"const fn = (a: number = 5, b: boolean = true, c: string = 'foo') => {};", NoInferrableTypesOptions{IgnoreParameters: false, IgnoreProperties: false}, []string{"noInferrableType", "noInferrableType", "noInferrableType"}, noInferrableTypesPointer("const fn = (a = 5, b = true, c = 'foo') => {};")},
		{"\nclass Foo {\n  a: number = 5;\n  b: boolean = true;\n  c: string = 'foo';\n}\n      ", NoInferrableTypesOptions{IgnoreParameters: false, IgnoreProperties: false}, []string{"noInferrableType", "noInferrableType", "noInferrableType"}, noInferrableTypesPointer("\nclass Foo {\n  a = 5;\n  b = true;\n  c = 'foo';\n}\n      ")},
		{"\nclass Foo {\n  constructor(public a: boolean = true) {}\n}\n      ", NoInferrableTypesOptions{IgnoreParameters: false, IgnoreProperties: false}, []string{"noInferrableType"}, noInferrableTypesPointer("\nclass Foo {\n  constructor(public a = true) {}\n}\n      ")},
		{"\nclass Foo {\n  accessor a: number = 5;\n}\n      ", NoInferrableTypesOptions{}, []string{"noInferrableType"}, noInferrableTypesPointer("\nclass Foo {\n  accessor a = 5;\n}\n      ")},
	}
}

// TestNoInferrableTypesMatchesUpstream replays the whole corpus.
func TestNoInferrableTypesMatchesUpstream(t *testing.T) {
	t.Parallel()

	cases := noInferrableTypesCases()

	reporting, fixable := 0, 0
	for _, testCase := range cases {
		if len(testCase.wantIds) > 0 {
			reporting++
		}
		if testCase.wantOutput != nil {
			fixable++
		}
	}
	if len(cases) != 99 {
		t.Fatalf("expected 99 corpus cases, have %d", len(cases))
	}
	if reporting != 44 {
		t.Fatalf("expected 44 reporting cases, have %d", reporting)
	}
	if fixable != 44 {
		t.Fatalf("expected 44 fixable cases, have %d", fixable)
	}

	for _, testCase := range cases {
		result := runNoInferrableTypes(t, testCase)
		rule_testing.ExpectFindings(t, result, testCase.wantIds...)
	}
}

// TestNoInferrableTypesRewritesWhatUpstreamRewrites is the fixer's only real coverage.
//
// Every repair here deletes bytes, which is the shape most likely to be wrong in a way no message id
// can see: deleting the annotation without its colon writes `const a: = 5`, and deleting it without
// the `?` on an optional parameter writes `function f(a? = 5)`. Both parse-break, both report
// identically, and only comparing the rewritten file separates them.
func TestNoInferrableTypesRewritesWhatUpstreamRewrites(t *testing.T) {
	t.Parallel()

	checked := 0
	for _, testCase := range noInferrableTypesCases() {
		if testCase.wantOutput == nil {
			continue
		}
		checked++
		result := runNoInferrableTypes(t, testCase)
		rule_testing.ExpectFixedSource(t, result, *testCase.wantOutput)
	}
	if checked != 44 {
		t.Fatalf("expected to check 44 rewrites, checked %d", checked)
	}
}

// TestNoInferrableTypesExemptsReadonlyForACompilerReason pins an exemption that reads as an
// oversight and is not.
//
// `class C { readonly a: number = 5 }` is clean. Upstream cites Microsoft/TypeScript#14416: without
// the annotation the property's type narrows to the literal `5` rather than to `number`, so removing
// it changes what the class means. This is the same family as an `output: null` -- a case where the
// repair would alter behaviour -- expressed by declining to report at all.
//
// The control is the same property without `readonly`, which reports. Without it this test also
// passes for a rule that never looks at class properties.
func TestNoInferrableTypesExemptsReadonlyForACompilerReason(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"class C { readonly a: number = 5; }",
		"class C { readonly a: string = 'x'; }",
		"class C { a?: number = 5; }",
	} {
		result := runNoInferrableTypes(t, noInferrableTypesCase{sourceText: source})
		rule_testing.ExpectClean(t, result)
	}

	control := runNoInferrableTypes(t, noInferrableTypesCase{
		sourceText: "class C { a: number = 5; }",
	})
	rule_testing.ExpectFindings(t, control, "noInferrableType")
}

// TestNoInferrableTypesOptionsSuppressEachVisitor covers the two flags independently.
//
// Each turns off one listener, so a rule that wired them to the wrong visitor -- or to neither --
// passes every corpus row that configures only one of them. The source below carries a violation in
// all three positions at once, so each flag's effect is visible as a change in COUNT rather than as
// a silence that could mean anything.
func TestNoInferrableTypesOptionsSuppressEachVisitor(t *testing.T) {
	t.Parallel()

	const source = "const a: number = 5;\nfunction f(b: number = 5) {}\nclass C { c: number = 5; }"

	all := runNoInferrableTypes(t, noInferrableTypesCase{sourceText: source})
	rule_testing.ExpectFindings(t, all, "noInferrableType", "noInferrableType", "noInferrableType")

	withoutParameters := runNoInferrableTypes(t, noInferrableTypesCase{
		sourceText: source,
		options:    NoInferrableTypesOptions{IgnoreParameters: true},
	})
	rule_testing.ExpectFindings(t, withoutParameters, "noInferrableType", "noInferrableType")

	withoutProperties := runNoInferrableTypes(t, noInferrableTypesCase{
		sourceText: source,
		options:    NoInferrableTypesOptions{IgnoreProperties: true},
	})
	rule_testing.ExpectFindings(t, withoutProperties, "noInferrableType", "noInferrableType")

	neither := runNoInferrableTypes(t, noInferrableTypesCase{
		sourceText: source,
		options:    NoInferrableTypesOptions{IgnoreParameters: true, IgnoreProperties: true},
	})
	rule_testing.ExpectFindings(t, neither, "noInferrableType")
}

// TestNoInferrableTypesAnchorsAParameterPropertyOnItsName pins a span the corpus cannot see.
//
// Upstream unwraps a `TSParameterProperty` to `param.parameter` before reporting, so the finding on
// `constructor(public readonly a: string = ”)` starts at `a`, not at `public`. This tree gives the
// whole thing one flat `KindParameter`, so reporting the node -- the obvious translation -- includes
// the modifiers.
//
// Upstream's own corpus has parameter-property cases and none of them declares a column, so
// `RuleTester` never compares one and every id fixture passes either way. A differential over 89
// real files caught it: 202 of 204 findings matched exactly and both mismatches were parameter
// properties differing only in column.
func TestNoInferrableTypesAnchorsAParameterPropertyOnItsName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source     string
		wantAnchor string
	}{
		{"class C { constructor(public readonly a: string = '') {} }", "a"},
		{"class C { constructor(public a: boolean = true) {} }", "a"},
		{"class C { constructor(private readonly count: number = 0) {} }", "count"},
		// A plain parameter has no modifiers, so it keeps the whole-node anchor upstream uses for
		// an AssignmentPattern. This is the control: without it the assertions above also pass for
		// a rule that always anchors on the name.
		{"function f(a: number = 5) {}", "a: number = 5"},
	}
	for _, testCase := range cases {
		result := runNoInferrableTypes(t, noInferrableTypesCase{sourceText: testCase.source})
		if len(result.Diagnostics) != 1 {
			t.Fatalf("%q: expected 1 finding, got %d", testCase.source, len(result.Diagnostics))
		}
		span := result.Diagnostics[0].Range
		reported := testCase.source[span.Pos():span.End()]
		if reported != testCase.wantAnchor {
			t.Errorf("%q: anchored on %q, want %q", testCase.source, reported, testCase.wantAnchor)
		}
	}
}
