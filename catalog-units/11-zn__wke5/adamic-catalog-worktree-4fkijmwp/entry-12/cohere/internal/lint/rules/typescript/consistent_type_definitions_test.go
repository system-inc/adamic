package typescript

import (
	"encoding/json"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// consistentTypeDefinitionsFile is where the fixtures pretend to live.
const consistentTypeDefinitionsFile = "/repository/source/ConsistentTypeDefinitions.ts"

// consistentTypeDefinitionsPointer spells an expected rewrite in a fixture row. A nil expectation
// means the rule proposed no fix, which for this rule is a real verdict rather than an absence:
// upstream deliberately declines to repair an interface inside `declare global`.
func consistentTypeDefinitionsPointer(value string) *string { return &value }

// consistentTypeDefinitionsCase is one imported corpus row.
type consistentTypeDefinitionsCase struct {
	sourceText string
	style      string
	wantIds    []string
	wantOutput *string
}

// runConsistentTypeDefinitions drives one case through the rule's own exported decoder.
//
// Through the decoder because the default is not the zero value: an unconfigured rule must mean
// `interface`, and a plain string field would read empty input as "", which matches neither arm and
// silences the rule on every file while looking configured.
func runConsistentTypeDefinitions(
	t *testing.T,
	testCase consistentTypeDefinitionsCase,
) rule_testing.Result {
	t.Helper()

	encoded, err := json.Marshal(testCase.style)
	if err != nil {
		t.Fatalf("could not encode options: %v", err)
	}
	decoded, err := DecodeConsistentTypeDefinitionsOptions(encoded)
	if err != nil {
		t.Fatalf("could not decode options %q: %v", testCase.style, err)
	}
	return rule_testing.RunWithOptions(t, ConsistentTypeDefinitions,
		consistentTypeDefinitionsFile, testCase.sourceText, decoded)
}

// The corpus is typescript-eslint's own, extracted mechanically rather than retyped.
//
// `tests/rules/consistent-type-definitions.test.ts` (clone, 8.69.0) was loaded with its RuleTester
// stubbed so every case came out as data, then every case was replayed against the INSTALLED 8.67.0
// rule through the ESLint Linter API. The expectations below are what the installed rule answered.
//
// Both halves agree with upstream's own annotations exactly: 10 cases it calls valid report nothing
// here, 23 it calls invalid report, and zero disagree in either direction.
//
// The `wantOutput` column is the fixer's specification and it came from the same run --
// `linter.verifyAndFix`, so it is what the real pipeline produces rather than a reconstruction. All
// 21 fixable rows agree with the annotated `output` in the corpus file, and the 2 rows carrying
// `output: null` were confirmed to be genuine declines rather than fixers that happen to fail.
func consistentTypeDefinitionsCases() []consistentTypeDefinitionsCase {
	ptr := consistentTypeDefinitionsPointer
	_ = ptr
	return []consistentTypeDefinitionsCase{
		{"var foo = {};", "interface", []string{}, nil},
		{"interface A {}", "interface", []string{}, nil},
		{"\ninterface A extends B {\n  x: number;\n}\n      ", "interface", []string{}, nil},
		{"type U = string;", "interface", []string{}, nil},
		{"type V = { x: number } | { y: string };", "interface", []string{}, nil},
		{"\ntype Record<T, U> = {\n  [K in T]: U;\n};\n      ", "interface", []string{}, nil},
		{"type T = { x: number };", "type", []string{}, nil},
		{"type A = { x: number } & B & C;", "type", []string{}, nil},
		{"type A = { x: number } & B<T1> & C<T2>;", "type", []string{}, nil},
		{"\nexport type W<T> = {\n  x: T;\n};\n      ", "type", []string{}, nil},
		{"type T = { x: number; };", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("interface T { x: number; }")},
		{"type T={ x: number; };", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("interface T { x: number; }")},
		{"type T=                         { x: number; };", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("interface T { x: number; }")},
		{"type T /* comment */={ x: number; };", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("interface T /* comment */ { x: number; }")},
		{"\nexport type W<T> = {\n  x: T;\n};\n      ", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("\nexport interface W<T> {\n  x: T;\n}\n      ")},
		{"interface T { x: number; }", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("type T = { x: number; }")},
		{"interface T{ x: number; }", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("type T = { x: number; }")},
		{"interface T                          { x: number; }", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("type T = { x: number; }")},
		{"interface A extends B, C { x: number; };", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("type A = { x: number; } & B & C;")},
		{"interface A extends B<T1>, C<T2> { x: number; };", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("type A = { x: number; } & B<T1> & C<T2>;")},
		{"\nexport interface W<T> {\n  x: T;\n}\n      ", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("\nexport type W<T> = {\n  x: T;\n}\n      ")},
		{"\nnamespace JSX {\n  interface Array<T> {\n    foo(x: (x: number) => T): T[];\n  }\n}\n      ", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("\nnamespace JSX {\n  type Array<T> = {\n    foo(x: (x: number) => T): T[];\n  }\n}\n      ")},
		{"\nglobal {\n  interface Array<T> {\n    foo(x: (x: number) => T): T[];\n  }\n}\n      ", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("\nglobal {\n  type Array<T> = {\n    foo(x: (x: number) => T): T[];\n  }\n}\n      ")},
		{"\ndeclare global {\n  interface Array<T> {\n    foo(x: (x: number) => T): T[];\n  }\n}\n      ", "type", []string{"typeOverInterface"}, nil},
		{"\ndeclare global {\n  namespace Foo {\n    interface Bar {}\n  }\n}\n      ", "type", []string{"typeOverInterface"}, nil},
		{"\nexport default interface Test {\n  bar(): string;\n  foo(): number;\n}\n      ", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("\ntype Test = {\n  bar(): string;\n  foo(): number;\n}\nexport default Test\n      ")},
		{"\nexport declare type Test = {\n  foo: string;\n  bar: string;\n};\n      ", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("\nexport declare interface Test {\n  foo: string;\n  bar: string;\n}\n      ")},
		{"\nexport declare interface Test {\n  foo: string;\n  bar: string;\n}\n      ", "type", []string{"typeOverInterface"}, consistentTypeDefinitionsPointer("\nexport declare type Test = {\n  foo: string;\n  bar: string;\n}\n      ")},
		{"\ntype Foo = ({\n  a: string;\n});\n      ", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("\ninterface Foo {\n  a: string;\n}\n      ")},
		{"\ntype Foo = ((((((((({\n  a: string;\n})))))))));\n      ", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("\ninterface Foo {\n  a: string;\n}\n      ")},
		{"\ntype Foo = {\n  a: string;\n}\n      ", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("\ninterface Foo {\n  a: string;\n}\n      ")},
		{"\ntype Foo = {\n  a: string;\n}\ntype Bar = string;\n      ", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("\ninterface Foo {\n  a: string;\n}\ntype Bar = string;\n      ")},
		{"\ntype Foo = ((({\n  a: string;\n})))\n\nconst bar = 1;\n      ", "interface", []string{"interfaceOverType"}, consistentTypeDefinitionsPointer("\ninterface Foo {\n  a: string;\n}\n\nconst bar = 1;\n      ")},
	}
}

// TestConsistentTypeDefinitionsMatchesUpstream replays the whole corpus.
func TestConsistentTypeDefinitionsMatchesUpstream(t *testing.T) {
	t.Parallel()

	cases := consistentTypeDefinitionsCases()

	reporting := 0
	fixable := 0
	for _, testCase := range cases {
		if len(testCase.wantIds) > 0 {
			reporting++
		}
		if testCase.wantOutput != nil {
			fixable++
		}
	}
	if len(cases) != 33 {
		t.Fatalf("expected 33 corpus cases, have %d", len(cases))
	}
	if reporting != 23 {
		t.Fatalf("expected 23 reporting cases, have %d", reporting)
	}
	if fixable != 21 {
		t.Fatalf("expected 21 fixable cases, have %d", fixable)
	}

	for _, testCase := range cases {
		result := runConsistentTypeDefinitions(t, testCase)
		rule_testing.ExpectFindings(t, result, testCase.wantIds...)
	}
}

// TestConsistentTypeDefinitionsRewritesWhatUpstreamRewrites is the fixer's only real coverage.
//
// A fixer that reports in the right place and repairs wrongly is indistinguishable, at the
// message-id layer, from a correct one: every row in the test above stays green either way. So the
// whole rewritten file is compared, against text the installed 8.67.0 build produced through
// `verifyAndFix` rather than against anything reconstructed by hand.
func TestConsistentTypeDefinitionsRewritesWhatUpstreamRewrites(t *testing.T) {
	t.Parallel()

	checked := 0
	for _, testCase := range consistentTypeDefinitionsCases() {
		if testCase.wantOutput == nil {
			continue
		}
		checked++
		result := runConsistentTypeDefinitions(t, testCase)
		rule_testing.ExpectFixedSource(t, result, *testCase.wantOutput)
	}
	if checked != 21 {
		t.Fatalf("expected to check 21 rewrites, checked %d", checked)
	}
}

// TestConsistentTypeDefinitionsDeclinesInsideDeclareGlobal pins a decline, which is a verdict no
// message-id fixture can see.
//
// Upstream returns a null fixer for an interface inside `declare global`, citing its own issue
// 2707: a type alias there does not do what the interface did, so the repair would change meaning
// rather than spelling. The finding is identical whether or not a fix rides along, so only asserting
// the ABSENCE of a proposed fix can tell the two apart.
func TestConsistentTypeDefinitionsDeclinesInsideDeclareGlobal(t *testing.T) {
	t.Parallel()

	declined := runConsistentTypeDefinitions(t, consistentTypeDefinitionsCase{
		sourceText: "declare global {\n  interface Array<T> {\n    foo(): T;\n  }\n}",
		style:      "type",
	})
	rule_testing.ExpectFindings(t, declined, "typeOverInterface")
	for _, diagnostic := range declined.Diagnostics {
		if len(diagnostic.Fixes) > 0 {
			t.Errorf("expected no fix inside `declare global`, got %d", len(diagnostic.Fixes))
		}
	}

	// The control. The same interface OUTSIDE a global augmentation is repaired, so the decline
	// above is the guard rather than a fixer that never fires under this option.
	repaired := runConsistentTypeDefinitions(t, consistentTypeDefinitionsCase{
		sourceText: "interface Array2<T> {\n  foo(): T;\n}",
		style:      "type",
	})
	rule_testing.ExpectFindings(t, repaired, "typeOverInterface")
	proposed := 0
	for _, diagnostic := range repaired.Diagnostics {
		proposed += len(diagnostic.Fixes)
	}
	if proposed == 0 {
		t.Fatal("the control proposed no fix, so the decline above proves nothing")
	}
}

// TestConsistentTypeDefinitionsDecoderDefaultsToInterface covers the arm the corpus cannot reach.
//
// Every corpus row supplies an option. A rule configured as a bare "error" reaches the decoder with
// empty input, and the zero value of the settings struct is the empty string, which matches neither
// arm -- a rule that is silent on every file while looking configured, and indistinguishable from a
// clean tree in a violation count.
func TestConsistentTypeDefinitionsDecoderDefaultsToInterface(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeConsistentTypeDefinitionsOptions(nil)
	if err != nil {
		t.Fatalf("empty options should decode to the default: %v", err)
	}
	settings, ok := decoded.(ConsistentTypeDefinitionsOptions)
	if !ok {
		t.Fatalf("expected ConsistentTypeDefinitionsOptions, got %T", decoded)
	}
	if settings.Style != ConsistentTypeDefinitionsInterface {
		t.Fatalf("expected the default to be interface, got %q", settings.Style)
	}

	// The control that the default is doing work: unconfigured, the rule reports a type alias and
	// stays silent on an interface, which is the `interface` arm rather than no arm at all.
	reported := rule_testing.RunWithOptions(t, ConsistentTypeDefinitions,
		consistentTypeDefinitionsFile, "type T = { x: number };", settings)
	rule_testing.ExpectFindings(t, reported, "interfaceOverType")

	silent := rule_testing.RunWithOptions(t, ConsistentTypeDefinitions,
		consistentTypeDefinitionsFile, "interface I { x: number }", settings)
	rule_testing.ExpectClean(t, silent)

	// And a bad value fails loudly rather than silencing the rule.
	if _, err := DecodeConsistentTypeDefinitionsOptions([]byte(`"neither"`)); err == nil {
		t.Error("expected the decoder to refuse an unknown style")
	}
}

// TestConsistentTypeDefinitionsKeepsHeritageOrder pins an ordering the corpus asserts only twice and
// that a plausible implementation gets backwards.
//
// `interface A extends B, C {}` becomes `type A = {...} & B & C`, and the order of the intersection
// is the order of the `extends` list. Upstream emits one insertion per heritage type and ESLint
// applies same-position insertions in the order proposed; this tree's edit engine applies fixes back
// to front, so a run of insertions sharing a position comes out REVERSED. The result is valid
// TypeScript naming the constraints in the wrong order, which is the "repairs to something that
// means a different thing" class rather than a broken repair, and no message-id fixture can see it.
//
// Three and four heritage types are asserted as well as two, because a reversal of two is also a
// rotation and a rule that rotated rather than reversed would pass at two.
func TestConsistentTypeDefinitionsKeepsHeritageOrder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		want   string
	}{
		{"interface A extends B, C {\n  x: number;\n}", "type A = {\n  x: number;\n} & B & C"},
		{"interface A extends B, C, D {\n  x: number;\n}", "type A = {\n  x: number;\n} & B & C & D"},
		{"interface A extends B, C, D, E {\n  x: number;\n}", "type A = {\n  x: number;\n} & B & C & D & E"},
		{"interface A extends B<T1>, C<T2> {\n  x: number;\n}", "type A = {\n  x: number;\n} & B<T1> & C<T2>"},
	}
	for _, testCase := range cases {
		result := runConsistentTypeDefinitions(t, consistentTypeDefinitionsCase{
			sourceText: testCase.source,
			style:      "type",
		})
		rule_testing.ExpectFixedSource(t, result, testCase.want)
	}
}

// TestConsistentTypeDefinitionsKeepsTheDeclarationModifiers pins what the rewrite must not delete.
//
// Swept after `consistent-indexed-object-style` shipped an interface-to-alias repair that replaced
// the declaration from its first modifier and rebuilt only `type <name> = `, unexporting six
// interfaces. This rule rewrites only the keyword and the punctuation around the body, so `export`,
// `declare`, the JSDoc above and a comment between the modifier and the keyword all stay where they
// were, in both directions. Every expected output is byte-identical to upstream's installed build.
func TestConsistentTypeDefinitionsKeepsTheDeclarationModifiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		style      string
		sourceText string
		want       string
	}{
		{"type", "export interface Foo { a: string }\n", "export type Foo = { a: string }\n"},
		{"type", "declare interface Foo { a: string }\n", "declare type Foo = { a: string }\n"},
		{"type", "/** Doc. */\nexport interface Foo { a: string }\n", "/** Doc. */\nexport type Foo = { a: string }\n"},
		{"type", "export /* c */ interface Foo { a: string }\n", "export /* c */ type Foo = { a: string }\n"},
		{"type", "namespace N {\n  export interface Foo { a: string }\n}\n", "namespace N {\n  export type Foo = { a: string }\n}\n"},
		{"interface", "export type Foo = { a: string };\n", "export interface Foo { a: string }\n"},
		{"interface", "declare type Foo = { a: string };\n", "declare interface Foo { a: string }\n"},
		{"interface", "/** Doc. */\nexport type Foo = { a: string };\n", "/** Doc. */\nexport interface Foo { a: string }\n"},
		{"interface", "export /* c */ type Foo = { a: string };\n", "export /* c */ interface Foo { a: string }\n"},
	}
	for _, testCase := range cases {
		decoded, err := DecodeConsistentTypeDefinitionsOptions([]byte(`"` + testCase.style + `"`))
		if err != nil {
			t.Fatal(err)
		}
		result := rule_testing.RunWithOptions(t, ConsistentTypeDefinitions,
			consistentTypeDefinitionsFile, testCase.sourceText, decoded)
		rule_testing.ExpectFixedSource(t, result, testCase.want)
	}
}
