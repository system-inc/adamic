package core

import (
	"encoding/json"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// groupedAccessorPairsFile is where the fixtures pretend to live.
const groupedAccessorPairsFile = "/repository/source/GroupedAccessorPairs.ts"

// decodedGroupedAccessorPairsOptions routes a fixture's options through the rule's own exported
// decoder rather than building the value directly.
//
// An empty string means the rule is configured as bare "error", which is what the live config
// does and what hands the rule nil. That row is what the decoder exists for: upstream's first
// option defaults to "anyOrder", so a decoder yielding a zero value would hand the rule an empty
// string matching no arm and silently drop the order check.
func decodedGroupedAccessorPairsOptions(t *testing.T, raw string) any {
	t.Helper()
	if raw == "" {
		return nil
	}
	options, err := DecodeGroupedAccessorPairsOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return options
}

// The corpus is ESLint's own, at `tests/lib/rules/grouped-accessor-pairs.js`, extracted by loading
// that file with a stub rule tester rather than retyped, so every string is upstream's own bytes.
//
// 79 clean cases and 57 reporting ones. Upstream ships 14 more behind `enforceForTSTypes`, which
// are in `TestGroupedAccessorPairsTypeMembers`.
func TestGroupedAccessorPairsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
	}{
		{"({})", ""},
		{"({ a })", ""},
		{"({ a(){}, b(){}, a(){} })", ""},
		{"({ a: 1, b: 2 })", ""},
		{"({ a, ...b, c: 1 })", ""},
		{"({ a, b, ...a })", ""},
		{"({ a: 1, [b]: 2, a: 3, [b]: 4 })", ""},
		{"({ a: function get(){}, b, a: function set(foo){} })", ""},
		{"({ get(){}, a, set(){} })", ""},
		{"class A {}", ""},
		{"(class { a(){} })", ""},
		{"class A { a(){} [b](){} a(){} [b](){} }", ""},
		{"(class { a(){} b(){} static a(){} static b(){} })", ""},
		{"class A { get(){} a(){} set(){} }", ""},
		{"({ get a(){} })", ""},
		{"({ set a(foo){} })", ""},
		{"({ a: 1, get b(){}, c, ...d })", ""},
		{"({ get a(){}, get b(){}, set c(foo){}, set d(foo){} })", ""},
		{"({ get a(){}, b: 1, set c(foo){} })", ""},
		{"({ set a(foo){}, b: 1, a: 2 })", ""},
		{"({ get a(){}, b: 1, a })", ""},
		{"({ set a(foo){}, b: 1, a(){} })", ""},
		{"({ get a(){}, b: 1, set [a](foo){} })", ""},
		{"({ set a(foo){}, b: 1, get 'a '(){} })", ""},
		{"({ get a(){}, b: 1, ...a })", ""},
		{"({ set a(foo){}, b: 1 }, { get a(){} })", ""},
		{"({ get a(){}, b: 1, ...{ set a(foo){} } })", ""},
		{"({ set a(foo){}, get b(){} })", "[\"getBeforeSet\"]"},
		{"({ get a(){}, set b(foo){} })", "[\"setBeforeGet\"]"},
		{"class A { get a(){} }", ""},
		{"(class { set a(foo){} })", ""},
		{"class A { static set a(foo){} }", ""},
		{"(class { static get a(){} })", ""},
		{"class A { a(){} set b(foo){} c(){} }", ""},
		{"(class { a(){} get b(){} c(){} })", ""},
		{"class A { get a(){} static get b(){} set c(foo){} static set d(bar){} }", ""},
		{"(class { get a(){} b(){} a(foo){} })", ""},
		{"class A { static set a(foo){} b(){} static a(){} }", ""},
		{"(class { get a(){} static b(){} set [a](foo){} })", ""},
		{"class A { static set a(foo){} b(){} static get ' a'(){} }", ""},
		{"(class { set a(foo){} b(){} static get a(){} })", ""},
		{"class A { static set a(foo){} b(){} get a(){} }", ""},
		{"(class { get a(){} }, class { b(){} set a(foo){} })", ""},
		{"({ get a(){}, set a(foo){} })", ""},
		{"({ a: 1, set b(foo){}, get b(){}, c: 2 })", ""},
		{"({ get a(){}, set a(foo){}, set b(bar){}, get b(){} })", ""},
		{"({ get [a](){}, set [a](foo){} })", ""},
		{"({ set a(foo){}, get 'a'(){} })", ""},
		{"({ a: 1, b: 2, get a(){}, set a(foo){}, c: 3, a: 4 })", ""},
		{"({ get a(){}, set a(foo){}, set b(bar){} })", ""},
		{"({ get a(){}, get b(){}, set b(bar){} })", ""},
		{"class A { get a(){} set a(foo){} }", ""},
		{"(class { set a(foo){} get a(){} })", ""},
		{"class A { static set a(foo){} static get a(){} }", ""},
		{"(class { static get a(){} static set a(foo){} })", ""},
		{"class A { a(){} set b(foo){} get b(){} c(){} get d(){} set d(bar){} }", ""},
		{"(class { set a(foo){} get a(){} get b(){} set b(bar){} })", ""},
		{"class A { static set [a](foo){} static get [a](){} }", ""},
		{"(class { get a(){} set [`a`](foo){} })", ""},
		{"class A { static get a(){} static set a(foo){} set a(bar){} static get a(){} }", ""},
		{"(class { static get a(){} get a(){} set a(foo){} })", ""},
		{"({ get a(){}, set a(foo){} })", "[\"anyOrder\"]"},
		{"({ set a(foo){}, get a(){} })", "[\"anyOrder\"]"},
		{"({ get a(){}, set a(foo){} })", "[\"getBeforeSet\"]"},
		{"({ set a(foo){}, get a(){} })", "[\"setBeforeGet\"]"},
		{"class A { get a(){} set a(foo){} }", "[\"anyOrder\"]"},
		{"(class { set a(foo){} get a(){} })", "[\"anyOrder\"]"},
		{"class A { get a(){} set a(foo){} }", "[\"getBeforeSet\"]"},
		{"(class { static set a(foo){} static get a(){} })", "[\"setBeforeGet\"]"},
		{"({ get a(){}, b: 1, get a(){} })", ""},
		{"({ set a(foo){}, b: 1, set a(foo){} })", ""},
		{"({ get a(){}, b: 1, set a(foo){}, c: 2, get a(){} })", ""},
		{"({ set a(foo){}, b: 1, set 'a'(bar){}, c: 2, get a(){} })", ""},
		{"class A { get [a](){} b(){} get [a](){} c(){} set [a](foo){} }", ""},
		{"(class { static set a(foo){} b(){} static get a(){} static c(){} static set a(bar){} })", ""},
		{"class A { get '#abc'(){} b(){} set #abc(foo){} }", ""},
		{"class A { get #abc(){} b(){} set '#abc'(foo){} }", ""},
		{"class A { set '#abc'(foo){} get #abc(){} }", "[\"getBeforeSet\"]"},
		{"class A { set #abc(foo){} get '#abc'(){} }", "[\"getBeforeSet\"]"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, GroupedAccessorPairs,
				groupedAccessorPairsFile, testCase.sourceText, decodedGroupedAccessorPairsOptions(t, testCase.options)))
		})
	}
}

// Every reporting case, with its own message ids in upstream's own order. Nine of them report
// more than once, so the counts are upstream data rather than an assumption.
func TestGroupedAccessorPairsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		wantIds    []string
	}{
		{"({ get a(){}, b:1, set a(foo){} })", "", []string{"notGrouped"}},
		{"({ set 'abc'(foo){}, b:1, get 'abc'(){} })", "", []string{"notGrouped"}},
		{"({ get [a](){}, b:1, set [a](foo){} })", "", []string{"notGrouped"}},
		{"class A { get abc(){} b(){} set abc(foo){} }", "", []string{"notGrouped"}},
		{"(class { set abc(foo){} b(){} get abc(){} })", "", []string{"notGrouped"}},
		{"class A { static set a(foo){} b(){} static get a(){} }", "", []string{"notGrouped"}},
		{"(class { static get 123(){} b(){} static set 123(foo){} })", "", []string{"notGrouped"}},
		{"class A { static get [a](){} b(){} static set [a](foo){} }", "", []string{"notGrouped"}},
		{"class A { get '#abc'(){} b(){} set '#abc'(foo){} }", "", []string{"notGrouped"}},
		{"class A { get #abc(){} b(){} set #abc(foo){} }", "", []string{"notGrouped"}},
		{"({ set a(foo){}, get a(){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"({ get 123(){}, set 123(foo){} })", "[\"setBeforeGet\"]", []string{"invalidOrder"}},
		{"({ get [a](){}, set [a](foo){} })", "[\"setBeforeGet\"]", []string{"invalidOrder"}},
		{"class A { set abc(foo){} get abc(){} }", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"(class { get [`abc`](){} set [`abc`](foo){} })", "[\"setBeforeGet\"]", []string{"invalidOrder"}},
		{"class A { static get a(){} static set a(foo){} }", "[\"setBeforeGet\"]", []string{"invalidOrder"}},
		{"(class { static set 'abc'(foo){} static get 'abc'(){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"class A { static set [abc](foo){} static get [abc](){} }", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"class A { set '#abc'(foo){} get '#abc'(){} }", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"class A { set #abc(foo){} get #abc(){} }", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"({ get a(){}, b: 1, set a(foo){} })", "[\"anyOrder\"]", []string{"notGrouped"}},
		{"({ get a(){}, b: 1, set a(foo){} })", "[\"setBeforeGet\"]", []string{"notGrouped"}},
		{"({ get a(){}, b: 1, set a(foo){} })", "[\"getBeforeSet\"]", []string{"notGrouped"}},
		{"class A { set a(foo){} b(){} get a(){} }", "[\"getBeforeSet\"]", []string{"notGrouped"}},
		{"(class { static set a(foo){} b(){} static get a(){} })", "[\"setBeforeGet\"]", []string{"notGrouped"}},
		{"({ get 'abc'(){}, d(){}, set 'abc'(foo){} })", "", []string{"notGrouped"}},
		{"({ set ''(foo){}, get [''](){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"class A { set abc(foo){} get 'abc'(){} }", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"(class { set [`abc`](foo){} get abc(){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"({ set ['abc'](foo){}, get [`abc`](){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"({ set 123(foo){}, get [123](){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"class A { static set '123'(foo){} static get 123(){} }", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"(class { set [a+b](foo){} get [a+b](){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"({ set [f(a)](foo){}, get [f(a)](){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"({ get a(){}, b: 1, set a(foo){}, set c(foo){}, d(){}, get c(){} })", "", []string{"notGrouped", "notGrouped"}},
		{"({ get a(){}, set b(foo){}, set a(bar){}, get b(){} })", "", []string{"notGrouped", "notGrouped"}},
		{"({ get a(){}, set [a](foo){}, set a(bar){}, get [a](){} })", "", []string{"notGrouped", "notGrouped"}},
		{"({ a(){}, set b(foo){}, ...c, get b(){}, set c(bar){}, get c(){} })", "[\"getBeforeSet\"]", []string{"notGrouped", "invalidOrder"}},
		{"({ set [a](foo){}, get [a](){}, set [-a](bar){}, get [-a](){} })", "[\"getBeforeSet\"]", []string{"invalidOrder", "invalidOrder"}},
		{"class A { get a(){} constructor (){} set a(foo){} get b(){} static c(){} set b(bar){} }", "", []string{"notGrouped", "notGrouped"}},
		{"(class { set a(foo){} static get a(){} get a(){} static set a(bar){} })", "", []string{"notGrouped", "notGrouped"}},
		{"class A { get a(){} set a(foo){} static get b(){} static set b(bar){} }", "[\"setBeforeGet\"]", []string{"invalidOrder", "invalidOrder"}},
		{"(class { set [a+b](foo){} get [a-b](){} get [a+b](){} set [a-b](bar){} })", "", []string{"notGrouped", "notGrouped"}},
		{"({ get a(){}, set a(foo){}, get b(){}, c: function(){}, set b(bar){} })", "", []string{"notGrouped"}},
		{"({ get a(){}, get b(){}, set a(foo){} })", "", []string{"notGrouped"}},
		{"({ set a(foo){}, get [a](){}, get a(){} })", "", []string{"notGrouped"}},
		{"({ set [a](foo){}, set a(bar){}, get [a](){} })", "", []string{"notGrouped"}},
		{"({ get a(){}, set a(foo){}, set b(bar){}, get b(){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"class A { get a(){} static set b(foo){} static get b(){} set a(foo){} }", "", []string{"notGrouped"}},
		{"(class { static get a(){} set a(foo){} static set a(bar){} })", "", []string{"notGrouped"}},
		{"class A { set a(foo){} get a(){} static get a(){} static set a(bar){} }", "[\"setBeforeGet\"]", []string{"invalidOrder"}},
		{"({ get a(){}, a: 1, set a(foo){} })", "", []string{"notGrouped"}},
		{"({ a(){}, set a(foo){}, get a(){} })", "[\"getBeforeSet\"]", []string{"invalidOrder"}},
		{"class A { get a(){} a(){} set a(foo){} }", "", []string{"notGrouped"}},
		{"class A { get a(){} a; set a(foo){} }", "", []string{"notGrouped"}},
		{"({ get a(){},\n    b: 1,\n    set a(foo){}\n})", "", []string{"notGrouped"}},
		{"class A { static set a(foo){} b(){} static get \n a(){}\n}", "", []string{"notGrouped"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, GroupedAccessorPairs,
				groupedAccessorPairsFile, testCase.sourceText, decodedGroupedAccessorPairsOptions(t, testCase.options)),
				testCase.wantIds...)
		})
	}
}

// The span upstream reports, for every finding that states a column.
//
// It is the accessor HEAD rather than the whole member: `set a`, `static get a`, `set [a]`, from
// the `get`/`set` or `static` keyword through the key, stopping before the parameter list.
// Upstream computes it with `getFunctionHeadLoc` over the accessor's function VALUE; our tree has
// no separate value node, so the same span is the member's own token start through its key's end,
// which was probed against seven shapes including a static, a computed key, a private name, a
// numeric key and a string key before being written.
//
// Every expectation here was measured against the installed eslint at 10.8.1 rather than read off
// this rule. Pointing at the whole member instead is a defensible reading of the same rule with
// an identical message id, and no id fixture could tell the two apart.
func TestGroupedAccessorPairsSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		index      int
		wantSpan   string
	}{
		{"({ get a(){}, b:1, set a(foo){} })", "", 0, "set a"},
		{"({ set 'abc'(foo){}, b:1, get 'abc'(){} })", "", 0, "get 'abc'"},
		{"({ get [a](){}, b:1, set [a](foo){} })", "", 0, "set [a]"},
		{"class A { get abc(){} b(){} set abc(foo){} }", "", 0, "set abc"},
		{"(class { set abc(foo){} b(){} get abc(){} })", "", 0, "get abc"},
		{"class A { static set a(foo){} b(){} static get a(){} }", "", 0, "static get a"},
		{"(class { static get 123(){} b(){} static set 123(foo){} })", "", 0, "static set 123"},
		{"class A { static get [a](){} b(){} static set [a](foo){} }", "", 0, "static set [a]"},
		{"class A { get '#abc'(){} b(){} set '#abc'(foo){} }", "", 0, "set '#abc'"},
		{"class A { get #abc(){} b(){} set #abc(foo){} }", "", 0, "set #abc"},
		{"({ set a(foo){}, get a(){} })", "[\"getBeforeSet\"]", 0, "get a"},
		{"({ get 123(){}, set 123(foo){} })", "[\"setBeforeGet\"]", 0, "set 123"},
		{"({ get [a](){}, set [a](foo){} })", "[\"setBeforeGet\"]", 0, "set [a]"},
		{"class A { set abc(foo){} get abc(){} }", "[\"getBeforeSet\"]", 0, "get abc"},
		{"(class { get [`abc`](){} set [`abc`](foo){} })", "[\"setBeforeGet\"]", 0, "set [`abc`]"},
		{"class A { static get a(){} static set a(foo){} }", "[\"setBeforeGet\"]", 0, "static set a"},
		{"(class { static set 'abc'(foo){} static get 'abc'(){} })", "[\"getBeforeSet\"]", 0, "static get 'abc'"},
		{"class A { static set [abc](foo){} static get [abc](){} }", "[\"getBeforeSet\"]", 0, "static get [abc]"},
		{"class A { set '#abc'(foo){} get '#abc'(){} }", "[\"getBeforeSet\"]", 0, "get '#abc'"},
		{"class A { set #abc(foo){} get #abc(){} }", "[\"getBeforeSet\"]", 0, "get #abc"},
		{"({ get a(){}, b: 1, set a(foo){}, set c(foo){}, d(){}, get c(){} })", "", 0, "set a"},
		{"({ get a(){}, b: 1, set a(foo){}, set c(foo){}, d(){}, get c(){} })", "", 1, "get c"},
		{"({ get a(){}, set b(foo){}, set a(bar){}, get b(){} })", "", 0, "set a"},
		{"({ get a(){}, set b(foo){}, set a(bar){}, get b(){} })", "", 1, "get b"},
		{"({ get a(){}, set [a](foo){}, set a(bar){}, get [a](){} })", "", 0, "set a"},
		{"({ get a(){}, set [a](foo){}, set a(bar){}, get [a](){} })", "", 1, "get [a]"},
		{"({ a(){}, set b(foo){}, ...c, get b(){}, set c(bar){}, get c(){} })", "[\"getBeforeSet\"]", 0, "get b"},
		{"({ a(){}, set b(foo){}, ...c, get b(){}, set c(bar){}, get c(){} })", "[\"getBeforeSet\"]", 1, "get c"},
		{"({ set [a](foo){}, get [a](){}, set [-a](bar){}, get [-a](){} })", "[\"getBeforeSet\"]", 0, "get [a]"},
		{"({ set [a](foo){}, get [a](){}, set [-a](bar){}, get [-a](){} })", "[\"getBeforeSet\"]", 1, "get [-a]"},
		{"class A { get a(){} constructor (){} set a(foo){} get b(){} static c(){} set b(bar){} }", "", 0, "set a"},
		{"class A { get a(){} constructor (){} set a(foo){} get b(){} static c(){} set b(bar){} }", "", 1, "set b"},
		{"(class { set a(foo){} static get a(){} get a(){} static set a(bar){} })", "", 0, "get a"},
		{"(class { set a(foo){} static get a(){} get a(){} static set a(bar){} })", "", 1, "static set a"},
		{"class A { get a(){} set a(foo){} static get b(){} static set b(bar){} }", "[\"setBeforeGet\"]", 0, "set a"},
		{"class A { get a(){} set a(foo){} static get b(){} static set b(bar){} }", "[\"setBeforeGet\"]", 1, "static set b"},
		{"(class { set [a+b](foo){} get [a-b](){} get [a+b](){} set [a-b](bar){} })", "", 0, "get [a+b]"},
		{"(class { set [a+b](foo){} get [a-b](){} get [a+b](){} set [a-b](bar){} })", "", 1, "set [a-b]"},
		{"({ get a(){}, set a(foo){}, get b(){}, c: function(){}, set b(bar){} })", "", 0, "set b"},
		{"({ get a(){}, get b(){}, set a(foo){} })", "", 0, "set a"},
		{"({ set a(foo){}, get [a](){}, get a(){} })", "", 0, "get a"},
		{"({ set [a](foo){}, set a(bar){}, get [a](){} })", "", 0, "get [a]"},
		{"({ get a(){}, set a(foo){}, set b(bar){}, get b(){} })", "[\"getBeforeSet\"]", 0, "get b"},
		{"class A { get a(){} static set b(foo){} static get b(){} set a(foo){} }", "", 0, "set a"},
		{"(class { static get a(){} set a(foo){} static set a(bar){} })", "", 0, "static set a"},
		{"class A { set a(foo){} get a(){} static get a(){} static set a(bar){} }", "[\"setBeforeGet\"]", 0, "static set a"},
		{"({ get a(){}, a: 1, set a(foo){} })", "", 0, "set a"},
		{"({ a(){}, set a(foo){}, get a(){} })", "[\"getBeforeSet\"]", 0, "get a"},
		{"class A { get a(){} a(){} set a(foo){} }", "", 0, "set a"},
		{"class A { get a(){} a; set a(foo){} }", "", 0, "set a"},
		{"({ get a(){},\n    b: 1,\n    set a(foo){}\n})", "", 0, "set a"},
		{"class A { static set a(foo){} b(){} static get \n a(){}\n}", "", 0, "static get \n a"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, GroupedAccessorPairs, groupedAccessorPairsFile,
				testCase.sourceText, decodedGroupedAccessorPairsOptions(t, testCase.options))
			if testCase.index >= len(result.Diagnostics) {
				t.Fatalf("want a finding at index %d, got %d findings", testCase.index, len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[testCase.index]
			source := result.SourceFile.Text()
			if got := source[diagnostic.Range.Pos():diagnostic.Range.End()]; got != testCase.wantSpan {
				t.Errorf("finding %d points at %q, want %q", testCase.index, got, testCase.wantSpan)
			}
		})
	}
}

// Upstream's `enforceForTSTypes` option: the same judgment over the accessor signatures in an
// interface body or a type literal.
//
// Upstream's 14 cases for it, 9 clean and 5 reporting, extracted with the rest of the corpus, plus
// five shapes measured against the installed eslint at 10.8.1 through the typescript-eslint parser:
// a string key pairing with a computed literal one, a non-literal computed key, a method signature
// of the same name between the pair, a duplicate getter, and a type literal nested in an interface.
// The span column is the measured one, the accessor head, as for class members.
func TestGroupedAccessorPairsTypeMembers(t *testing.T) {
	t.Parallel()

	const enforced = `["anyOrder", {"enforceForTSTypes": true}]`
	cases := []struct {
		sourceText string
		options    string
		wantId     string
		wantSpan   string
	}{
		{"interface I { get prop(): any, set prop(value: any): void }", enforced, "", ""},
		{"interface I { set prop(value: any): void, get prop(): any }", enforced, "", ""},
		{"interface I { get a(): any, between: true, set b(value: any): void }", enforced, "", ""},
		{"interface I { before: true, get prop(): any, set prop(value: any): void, after: true }", `["getBeforeSet", {"enforceForTSTypes": true}]`, "", ""},
		{"interface I { set prop(value: any): void, get prop(): any }", `["setBeforeGet", {"enforceForTSTypes": true}]`, "", ""},
		{"type T = { get prop(): any, set prop(value: any): void }", enforced, "", ""},
		{"type T = { set prop(value: any): void, get prop(): any }", `["setBeforeGet", {"enforceForTSTypes": true}]`, "", ""},
		{"interface I { get prop(): any, between: true, set prop(value: any): void }", "", "", ""},
		{"type T = { get prop(): any, between: true, set prop(value: any): void }", "", "", ""},
		{"interface I { get a(): any, between: true, set a(value: any): void }", enforced, "notGrouped", "set a"},
		{"interface I { get a(): any, set a(value: any): void }", `["setBeforeGet", {"enforceForTSTypes": true}]`, "invalidOrder", "set a"},
		{"interface I { set a(value: any): void, get a(): any }", `["getBeforeSet", {"enforceForTSTypes": true}]`, "invalidOrder", "get a"},
		{"type T = { get a(): any, between: true, set a(value: any): void }", enforced, "notGrouped", "set a"},
		{"type T = { get a(): any, set a(value: any): void }", `["setBeforeGet", {"enforceForTSTypes": true}]`, "invalidOrder", "set a"},

		// Measured, not in upstream's corpus.
		{"interface I { get 'a'(): any, x: 1, set ['a'](v: any): void }", enforced, "notGrouped", "set ['a']"},
		{"interface I { get [k](): any, x: 1, set [k](v: any): void }", enforced, "notGrouped", "set [k]"},
		{"interface I { get a(): any, a(): void, set a(v: any): void }", enforced, "notGrouped", "set a"},
		{"interface I { get a(): any, x: 1, set a(v: any): void, get a(): any }", enforced, "", ""},
		{"interface O { m: { get a(): any, x: 1, set a(v: any): void } }", enforced, "notGrouped", "set a"},

		// The option is what turns the check on: the same reporting source, off and at its default.
		{"interface I { get a(): any, x: 1, set a(v: any): void }", `["anyOrder", {"enforceForTSTypes": false}]`, "", ""},
		{"interface I { get a(): any, x: 1, set a(v: any): void }", `["anyOrder"]`, "", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText+" "+testCase.options, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, GroupedAccessorPairs, groupedAccessorPairsFile,
				testCase.sourceText, decodedGroupedAccessorPairsOptions(t, testCase.options))
			if testCase.wantId == "" {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantId)
			diagnostic := result.Diagnostics[0]
			if got := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]; got != testCase.wantSpan {
				t.Errorf("the finding points at %q, want %q", got, testCase.wantSpan)
			}
		})
	}
}

// The rendered names for type members, measured against the installed eslint at 10.8.1.
//
// The last row is the kept divergence the rule's doc comment names: upstream renders a non-literal
// computed signature key as `getter 'null'`, and this renders it as the bare `getter` a class
// accessor with the same key gets upstream too.
func TestGroupedAccessorPairsTypeMemberMessageText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		want       string
	}{
		{"interface I { get a(): any, between: true, set a(value: any): void }", `["anyOrder", {"enforceForTSTypes": true}]`, "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"type T = { get a(): any, set a(value: any): void }", `["setBeforeGet", {"enforceForTSTypes": true}]`, "Expected setter 'a' to be before getter 'a'."},
		{"interface I { get 'a'(): any, x: 1, set ['a'](v: any): void }", `["anyOrder", {"enforceForTSTypes": true}]`, "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"interface I { get [k](): any, x: 1, set [k](v: any): void }", `["anyOrder", {"enforceForTSTypes": true}]`, "Accessor pair getter and setter should be grouped."},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, GroupedAccessorPairs, groupedAccessorPairsFile,
				testCase.sourceText, decodedGroupedAccessorPairsOptions(t, testCase.options))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.want {
				t.Errorf("message %q, want %q", got, testCase.want)
			}
		})
	}
}

// The decoder is the line with no upstream counterpart, and this rule is the shape the brief
// warns about: the first option is a bare string whose default is "anyOrder", so a decoder
// yielding a zero value hands the rule an empty string matching no arm.
func TestDecodeGroupedAccessorPairsOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"empty input falls back to the default", "", "anyOrder"},
		{"an explicit anyOrder", `["anyOrder"]`, "anyOrder"},
		{"getBeforeSet", `["getBeforeSet"]`, "getBeforeSet"},
		{"setBeforeGet", `["setBeforeGet"]`, "setBeforeGet"},
		{"a second element at its default is accepted", `["setBeforeGet", {"enforceForTSTypes": false}]`, "setBeforeGet"},
		{"a second element turning on type members", `["getBeforeSet", {"enforceForTSTypes": true}]`, "getBeforeSet"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeGroupedAccessorPairsOptions(json.RawMessage(testCase.raw))
			if err != nil {
				t.Fatalf("the decoder refused %q: %v", testCase.raw, err)
			}
			options, isOptions := decoded.(GroupedAccessorPairsOptions)
			if !isOptions {
				t.Fatalf("the decoder returned %T rather than GroupedAccessorPairsOptions", decoded)
			}
			if string(options.Order) != testCase.want {
				t.Errorf("the order decodes to %q, want %q", options.Order, testCase.want)
			}
		})
	}
}

// A value outside the enum is refused rather than folded into one of the three arms.
//
// Upstream gets that refusal from its schema before the rule runs. There is no schema layer here,
// so it lives in the decoder, and a rule silently reading an unknown string as "anyOrder" would
// be the inert shape the config layer's own doc comment describes.
func TestDecodeGroupedAccessorPairsOptionsRefusesAValueOutsideTheEnum(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		`["sideways"]`, `["GetBeforeSet"]`, `[""]`, `[123]`,
		// A bare string, which the config layer never delivers to a list rule.
		`"getBeforeSet"`,
		// The second element: an unknown key and a third element are refused the way upstream's
		// schema refuses them.
		`["getBeforeSet", {"enforceForTypes": false}]`,
		`["getBeforeSet", {"enforceForTSTypes": false}, "anyOrder"]`,
	} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			if _, err := DecodeGroupedAccessorPairsOptions(json.RawMessage(raw)); err == nil {
				t.Errorf("the decoder accepted %s, which is outside upstream's enum", raw)
			}
		})
	}
}

// A rule configured as bare "error" reaches Run with nil rather than with an options struct. This
// asserts nil behaves as "anyOrder" rather than as no arm at all, which is the defect the
// hand-rolled decoder exists to prevent and which no fixture routing through the decoder can see.
func TestGroupedAccessorPairsWithNilOptionsDefaultsToAnyOrder(t *testing.T) {
	t.Parallel()

	// Ungrouped reports under every order, including the default.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, GroupedAccessorPairs,
		groupedAccessorPairsFile, "({ get a(){}, b:1, set a(foo){} })", nil), "notGrouped")
	// Grouped but set-before-get is clean under anyOrder and would report under getBeforeSet, so
	// this row is what separates the default from the other two arms.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, GroupedAccessorPairs,
		groupedAccessorPairsFile, "({ set a(foo){}, get a(){} })", nil))
}

// The rendered message text upstream asserts, for every finding in the corpus.
//
// Both messages interpolate two accessor names, so a message-id assertion cannot see anything the
// format string does. The text below is upstream's own `data` rendered through its own
// `meta.messages` templates, so it is upstream's bytes rather than this rule's.
//
// The naming has five forms and they are the reason this test is long: a plain `getter 'a'`, a
// bare `getter` when the key is computed and cannot be named, a `static setter 'a'`, and a
// `private getter #p` which carries no quotes at all.
func TestGroupedAccessorPairsMessageText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		index      int
		wantId     string
		wantText   string
	}{
		{"({ get a(){}, b:1, set a(foo){} })", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"({ set 'abc'(foo){}, b:1, get 'abc'(){} })", "", 0, "notGrouped", "Accessor pair setter 'abc' and getter 'abc' should be grouped."},
		{"({ get [a](){}, b:1, set [a](foo){} })", "", 0, "notGrouped", "Accessor pair getter and setter should be grouped."},
		{"class A { get abc(){} b(){} set abc(foo){} }", "", 0, "notGrouped", "Accessor pair getter 'abc' and setter 'abc' should be grouped."},
		{"(class { set abc(foo){} b(){} get abc(){} })", "", 0, "notGrouped", "Accessor pair setter 'abc' and getter 'abc' should be grouped."},
		{"class A { static set a(foo){} b(){} static get a(){} }", "", 0, "notGrouped", "Accessor pair static setter 'a' and static getter 'a' should be grouped."},
		{"(class { static get 123(){} b(){} static set 123(foo){} })", "", 0, "notGrouped", "Accessor pair static getter '123' and static setter '123' should be grouped."},
		{"class A { static get [a](){} b(){} static set [a](foo){} }", "", 0, "notGrouped", "Accessor pair static getter and static setter should be grouped."},
		{"class A { get '#abc'(){} b(){} set '#abc'(foo){} }", "", 0, "notGrouped", "Accessor pair getter '#abc' and setter '#abc' should be grouped."},
		{"class A { get #abc(){} b(){} set #abc(foo){} }", "", 0, "notGrouped", "Accessor pair private getter #abc and private setter #abc should be grouped."},
		{"({ set a(foo){}, get a(){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter 'a' to be before setter 'a'."},
		{"({ get 123(){}, set 123(foo){} })", "[\"setBeforeGet\"]", 0, "invalidOrder", "Expected setter '123' to be before getter '123'."},
		{"({ get [a](){}, set [a](foo){} })", "[\"setBeforeGet\"]", 0, "invalidOrder", "Expected setter to be before getter."},
		{"class A { set abc(foo){} get abc(){} }", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter 'abc' to be before setter 'abc'."},
		{"(class { get [`abc`](){} set [`abc`](foo){} })", "[\"setBeforeGet\"]", 0, "invalidOrder", "Expected setter 'abc' to be before getter 'abc'."},
		{"class A { static get a(){} static set a(foo){} }", "[\"setBeforeGet\"]", 0, "invalidOrder", "Expected static setter 'a' to be before static getter 'a'."},
		{"(class { static set 'abc'(foo){} static get 'abc'(){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected static getter 'abc' to be before static setter 'abc'."},
		{"class A { static set [abc](foo){} static get [abc](){} }", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected static getter to be before static setter."},
		{"class A { set '#abc'(foo){} get '#abc'(){} }", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter '#abc' to be before setter '#abc'."},
		{"class A { set #abc(foo){} get #abc(){} }", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected private getter #abc to be before private setter #abc."},
		{"({ get a(){}, b: 1, set a(foo){} })", "[\"anyOrder\"]", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"({ get a(){}, b: 1, set a(foo){} })", "[\"setBeforeGet\"]", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"({ get a(){}, b: 1, set a(foo){} })", "[\"getBeforeSet\"]", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"class A { set a(foo){} b(){} get a(){} }", "[\"getBeforeSet\"]", 0, "notGrouped", "Accessor pair setter 'a' and getter 'a' should be grouped."},
		{"(class { static set a(foo){} b(){} static get a(){} })", "[\"setBeforeGet\"]", 0, "notGrouped", "Accessor pair static setter 'a' and static getter 'a' should be grouped."},
		{"({ get 'abc'(){}, d(){}, set 'abc'(foo){} })", "", 0, "notGrouped", "Accessor pair getter 'abc' and setter 'abc' should be grouped."},
		{"({ set ''(foo){}, get [''](){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter '' to be before setter ''."},
		{"class A { set abc(foo){} get 'abc'(){} }", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter 'abc' to be before setter 'abc'."},
		{"(class { set [`abc`](foo){} get abc(){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter 'abc' to be before setter 'abc'."},
		{"({ set ['abc'](foo){}, get [`abc`](){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter 'abc' to be before setter 'abc'."},
		{"({ set 123(foo){}, get [123](){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter '123' to be before setter '123'."},
		{"class A { static set '123'(foo){} static get 123(){} }", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected static getter '123' to be before static setter '123'."},
		{"(class { set [a+b](foo){} get [a+b](){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter to be before setter."},
		{"({ set [f(a)](foo){}, get [f(a)](){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter to be before setter."},
		{"({ get a(){}, b: 1, set a(foo){}, set c(foo){}, d(){}, get c(){} })", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"({ get a(){}, b: 1, set a(foo){}, set c(foo){}, d(){}, get c(){} })", "", 1, "notGrouped", "Accessor pair setter 'c' and getter 'c' should be grouped."},
		{"({ get a(){}, set b(foo){}, set a(bar){}, get b(){} })", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"({ get a(){}, set b(foo){}, set a(bar){}, get b(){} })", "", 1, "notGrouped", "Accessor pair setter 'b' and getter 'b' should be grouped."},
		{"({ get a(){}, set [a](foo){}, set a(bar){}, get [a](){} })", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"({ get a(){}, set [a](foo){}, set a(bar){}, get [a](){} })", "", 1, "notGrouped", "Accessor pair setter and getter should be grouped."},
		{"({ a(){}, set b(foo){}, ...c, get b(){}, set c(bar){}, get c(){} })", "[\"getBeforeSet\"]", 0, "notGrouped", "Accessor pair setter 'b' and getter 'b' should be grouped."},
		{"({ a(){}, set b(foo){}, ...c, get b(){}, set c(bar){}, get c(){} })", "[\"getBeforeSet\"]", 1, "invalidOrder", "Expected getter 'c' to be before setter 'c'."},
		{"({ set [a](foo){}, get [a](){}, set [-a](bar){}, get [-a](){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter to be before setter."},
		{"({ set [a](foo){}, get [a](){}, set [-a](bar){}, get [-a](){} })", "[\"getBeforeSet\"]", 1, "invalidOrder", "Expected getter to be before setter."},
		{"class A { get a(){} constructor (){} set a(foo){} get b(){} static c(){} set b(bar){} }", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"class A { get a(){} constructor (){} set a(foo){} get b(){} static c(){} set b(bar){} }", "", 1, "notGrouped", "Accessor pair getter 'b' and setter 'b' should be grouped."},
		{"(class { set a(foo){} static get a(){} get a(){} static set a(bar){} })", "", 0, "notGrouped", "Accessor pair setter 'a' and getter 'a' should be grouped."},
		{"(class { set a(foo){} static get a(){} get a(){} static set a(bar){} })", "", 1, "notGrouped", "Accessor pair static getter 'a' and static setter 'a' should be grouped."},
		{"class A { get a(){} set a(foo){} static get b(){} static set b(bar){} }", "[\"setBeforeGet\"]", 0, "invalidOrder", "Expected setter 'a' to be before getter 'a'."},
		{"class A { get a(){} set a(foo){} static get b(){} static set b(bar){} }", "[\"setBeforeGet\"]", 1, "invalidOrder", "Expected static setter 'b' to be before static getter 'b'."},
		{"(class { set [a+b](foo){} get [a-b](){} get [a+b](){} set [a-b](bar){} })", "", 0, "notGrouped", "Accessor pair setter and getter should be grouped."},
		{"(class { set [a+b](foo){} get [a-b](){} get [a+b](){} set [a-b](bar){} })", "", 1, "notGrouped", "Accessor pair getter and setter should be grouped."},
		{"({ get a(){}, set a(foo){}, get b(){}, c: function(){}, set b(bar){} })", "", 0, "notGrouped", "Accessor pair getter 'b' and setter 'b' should be grouped."},
		{"({ get a(){}, get b(){}, set a(foo){} })", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"({ set a(foo){}, get [a](){}, get a(){} })", "", 0, "notGrouped", "Accessor pair setter 'a' and getter 'a' should be grouped."},
		{"({ set [a](foo){}, set a(bar){}, get [a](){} })", "", 0, "notGrouped", "Accessor pair setter and getter should be grouped."},
		{"({ get a(){}, set a(foo){}, set b(bar){}, get b(){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter 'b' to be before setter 'b'."},
		{"class A { get a(){} static set b(foo){} static get b(){} set a(foo){} }", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"(class { static get a(){} set a(foo){} static set a(bar){} })", "", 0, "notGrouped", "Accessor pair static getter 'a' and static setter 'a' should be grouped."},
		{"class A { set a(foo){} get a(){} static get a(){} static set a(bar){} }", "[\"setBeforeGet\"]", 0, "invalidOrder", "Expected static setter 'a' to be before static getter 'a'."},
		{"({ get a(){}, a: 1, set a(foo){} })", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"({ a(){}, set a(foo){}, get a(){} })", "[\"getBeforeSet\"]", 0, "invalidOrder", "Expected getter 'a' to be before setter 'a'."},
		{"class A { get a(){} a(){} set a(foo){} }", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"class A { get a(){} a; set a(foo){} }", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"({ get a(){},\n    b: 1,\n    set a(foo){}\n})", "", 0, "notGrouped", "Accessor pair getter 'a' and setter 'a' should be grouped."},
		{"class A { static set a(foo){} b(){} static get \n a(){}\n}", "", 0, "notGrouped", "Accessor pair static setter 'a' and static getter 'a' should be grouped."},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, GroupedAccessorPairs, groupedAccessorPairsFile,
				testCase.sourceText, decodedGroupedAccessorPairsOptions(t, testCase.options))
			if testCase.index >= len(result.Diagnostics) {
				t.Fatalf("want a finding at index %d, got %d findings", testCase.index, len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[testCase.index]
			if diagnostic.Message.Id != testCase.wantId {
				t.Errorf("finding %d is %q, want %q", testCase.index, diagnostic.Message.Id, testCase.wantId)
			}
			// Equality rather than a prefix or a contains: an interpolated message is exactly
			// where a weaker predicate stops being a guard.
			if diagnostic.Message.Description != testCase.wantText {
				t.Errorf("rendered:\n  %q\nwant:\n  %q", diagnostic.Message.Description, testCase.wantText)
			}
		})
	}
}

// A computed key whose expression is NOT a literal, beside a static key that reads the same way.
//
// A mutation tagging the token-text fallback as though it were a static name survived the whole
// imported corpus, because upstream never writes a computed key beside a plain key of the same
// spelling. These three do, and all three measured CLEAN against the installed eslint at 10.8.1:
// `get [a]` names whichever property the variable `a` holds, which is not knowable before it runs,
// so it can never be shown to be the same property as `a` or `'a'`.
//
// The fourth row is the control. Two computed keys spelled the same way DO pair, through upstream's
// token-list comparison, so a rule that simply declined every computed key would pass the first
// three for the wrong reason.
func TestGroupedAccessorPairsComputedKeysDoNotPairWithStaticOnes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
	}{
		{"({ get [a](){}, b:1, set 'a'(foo){} })", nil},
		{"({ get [a](){}, b:1, set a(foo){} })", nil},
		{"({ get [a.b](){}, b:1, set ['a.b'](foo){} })", nil},
		// The control: two non-literal computed keys with the same text pair through the token
		// comparison, which is what the fallback is for.
		{"({ get [a](){}, b:1, set [a](foo){} })", []string{"notGrouped"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, GroupedAccessorPairs, groupedAccessorPairsFile,
				testCase.sourceText, nil)
			if testCase.wantIds == nil {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}
