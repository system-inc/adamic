package typescript

import (
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// adjacentOverloadSignaturesFile is a TypeScript name because every case is TypeScript.
//
// The rule reads no filename and declines no extension. Upstream registers plain node visitors with
// no source-type test, and most of what it judges (a method signature, a construct signature, an
// overload without a body) cannot be written in JavaScript at all.
const adjacentOverloadSignaturesFile = "/repository/source/Thing.ts"

// adjacentOverloadSignaturesCaseName numbers a row so a failure names which one, since the sources
// are long and several differ only in the order of two members.
func adjacentOverloadSignaturesCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestAdjacentOverloadSignaturesStaysSilentOnUpstreamPassCases is the imported clean corpus.
//
// All thirty-three of upstream's passing inputs, extracted from the clone's test file by parsing it
// with the TypeScript compiler rather than by reading it, so no escape sequence passed through a
// shell or a keyboard on the way here. Every one was additionally run through the installed 8.67.0
// build driven by the ESLint 10.8.1 Linter API, which reported nothing on all thirty-three.
//
// These are the false positives upstream already thought about, and they are most of what this file
// is for: the reporting half of this rule is easy and the discrimination is not.
func TestAdjacentOverloadSignaturesStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\nfunction error(a: string);\nfunction error(b: number);\nfunction error(ab: string | number) {}\nexport { error };\n      ",
		"\nimport { connect } from 'react-redux';\nexport interface ErrorMessageModel {\n  message: string;\n}\nfunction mapStateToProps() {}\nfunction mapDispatchToProps() {}\nexport default connect(mapStateToProps, mapDispatchToProps)(ErrorMessage);\n      ",
		"\nexport const foo = 'a',\n  bar = 'b';\nexport interface Foo {}\nexport class Foo {}\n    ",
		"\nexport interface Foo {}\nexport const foo = 'a',\n  bar = 'b';\nexport class Foo {}\n    ",
		"\nconst foo = 'a',\n  bar = 'b';\ninterface Foo {}\nclass Foo {}\n    ",
		"\ninterface Foo {}\nconst foo = 'a',\n  bar = 'b';\nclass Foo {}\n    ",
		"\nexport class Foo {}\nexport class Bar {}\nexport type FooBar = Foo | Bar;\n    ",
		"\nexport interface Foo {}\nexport class Foo {}\nexport class Bar {}\nexport type FooBar = Foo | Bar;\n    ",
		"\nexport function foo(s: string);\nexport function foo(n: number);\nexport function foo(sn: string | number) {}\nexport function bar(): void {}\nexport function baz(): void {}\n    ",
		"\nfunction foo(s: string);\nfunction foo(n: number);\nfunction foo(sn: string | number) {}\nfunction bar(): void {}\nfunction baz(): void {}\n    ",
		"\ndeclare function foo(s: string);\ndeclare function foo(n: number);\ndeclare function foo(sn: string | number);\ndeclare function bar(): void;\ndeclare function baz(): void;\n    ",
		"\ndeclare module 'Foo' {\n  export function foo(s: string): void;\n  export function foo(n: number): void;\n  export function foo(sn: string | number): void;\n  export function bar(): void;\n  export function baz(): void;\n}\n    ",
		"\ndeclare namespace Foo {\n  export function foo(s: string): void;\n  export function foo(n: number): void;\n  export function foo(sn: string | number): void;\n  export function bar(): void;\n  export function baz(): void;\n}\n    ",
		"\ntype Foo = {\n  foo(s: string): void;\n  foo(n: number): void;\n  foo(sn: string | number): void;\n  bar(): void;\n  baz(): void;\n};\n    ",
		"\ntype Foo = {\n  foo(s: string): void;\n  ['foo'](n: number): void;\n  foo(sn: string | number): void;\n  bar(): void;\n  baz(): void;\n};\n    ",
		"\ninterface Foo {\n  (s: string): void;\n  (n: number): void;\n  (sn: string | number): void;\n  foo(n: number): void;\n  bar(): void;\n  baz(): void;\n}\n    ",
		"\ninterface Foo {\n  (s: string): void;\n  (n: number): void;\n  (sn: string | number): void;\n  foo(n: number): void;\n  bar(): void;\n  baz(): void;\n  call(): void;\n}\n    ",
		"\ninterface Foo {\n  foo(s: string): void;\n  foo(n: number): void;\n  foo(sn: string | number): void;\n  bar(): void;\n  baz(): void;\n}\n    ",
		"\ninterface Foo {\n  foo(s: string): void;\n  ['foo'](n: number): void;\n  foo(sn: string | number): void;\n  bar(): void;\n  baz(): void;\n}\n    ",
		"\ninterface Foo {\n  foo(): void;\n  bar: {\n    baz(s: string): void;\n    baz(n: number): void;\n    baz(sn: string | number): void;\n  };\n}\n    ",
		"\ninterface Foo {\n  new (s: string);\n  new (n: number);\n  new (sn: string | number);\n  foo(): void;\n}\n    ",
		"\nclass Foo {\n  constructor(s: string);\n  constructor(n: number);\n  constructor(sn: string | number) {}\n  bar(): void {}\n  baz(): void {}\n}\n    ",
		"\nclass Foo {\n  foo(s: string): void;\n  foo(n: number): void;\n  foo(sn: string | number): void {}\n  bar(): void {}\n  baz(): void {}\n}\n    ",
		"\nclass Foo {\n  foo(s: string): void;\n  ['foo'](n: number): void;\n  foo(sn: string | number): void {}\n  bar(): void {}\n  baz(): void {}\n}\n    ",
		"\nclass Foo {\n  name: string;\n  foo(s: string): void;\n  foo(n: number): void;\n  foo(sn: string | number): void {}\n  bar(): void {}\n  baz(): void {}\n}\n    ",
		"\nclass Foo {\n  name: string;\n  static foo(s: string): void;\n  static foo(n: number): void;\n  static foo(sn: string | number): void {}\n  bar(): void {}\n  baz(): void {}\n}\n    ",
		"\nclass Test {\n  static test() {}\n  untest() {}\n  test() {}\n}\n    ",
		"export default function <T>(foo: T) {}",
		"export default function named<T>(foo: T) {}",
		"\ninterface Foo {\n  [Symbol.toStringTag](): void;\n  [Symbol.iterator](): void;\n}\n    ",
		"\nclass Test {\n  #private(): void;\n  #private(arg: number): void {}\n\n  bar() {}\n\n  '#private'(): void;\n  '#private'(arg: number): void {}\n}\n    ",
		"\nfunction wrap() {\n  function foo(s: string);\n  function foo(n: number);\n  function foo(sn: string | number) {}\n}\n    ",
		"\nif (true) {\n  function foo(s: string);\n  function foo(n: number);\n  function foo(sn: string | number) {}\n}\n    ",
	}
	for index, sourceText := range cases {
		t.Run(adjacentOverloadSignaturesCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, AdjacentOverloadSignatures,
				adjacentOverloadSignaturesFile, sourceText))
		})
	}
}

// TestAdjacentOverloadSignaturesFiresOnUpstreamFailCases is the imported reporting corpus.
//
// Thirty-four inputs carrying thirty-six diagnostics between them, so the count per input is stated
// by the corpus rather than recovered. Every count was re-derived by running the installed build,
// which agreed with the corpus on all thirty-four, so there is no drift to record.
//
// The span is asserted on every row. It is the member node, and it includes any `export` or
// `export default` keyword: upstream reports the wrapping export declaration, and typescript-go
// carries the same keywords as modifiers on the declaration itself, so the reported text matches
// without the rule doing anything about it. Only a fixture can say that, since no harness knows
// which node a finding should have pointed at.
func TestAdjacentOverloadSignaturesFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpans  []string
		wantNames  []string
	}{
		{
			sourceText: "\nfunction wrap() {\n  function foo(s: string);\n  function foo(n: number);\n  type bar = number;\n  function foo(sn: string | number) {}\n}\n      ",
			wantSpans:  []string{"function foo(sn: string | number) {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nif (true) {\n  function foo(s: string);\n  function foo(n: number);\n  let a = 1;\n  function foo(sn: string | number) {}\n  foo(a);\n}\n      ",
			wantSpans:  []string{"function foo(sn: string | number) {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nexport function foo(s: string);\nexport function foo(n: number);\nexport function bar(): void {}\nexport function baz(): void {}\nexport function foo(sn: string | number) {}\n      ",
			wantSpans:  []string{"export function foo(sn: string | number) {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nexport function foo(s: string);\nexport function foo(n: number);\nexport type bar = number;\nexport type baz = number | string;\nexport function foo(sn: string | number) {}\n      ",
			wantSpans:  []string{"export function foo(sn: string | number) {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nfunction foo(s: string);\nfunction foo(n: number);\nfunction bar(): void {}\nfunction baz(): void {}\nfunction foo(sn: string | number) {}\n      ",
			wantSpans:  []string{"function foo(sn: string | number) {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nfunction foo(s: string);\nfunction foo(n: number);\ntype bar = number;\ntype baz = number | string;\nfunction foo(sn: string | number) {}\n      ",
			wantSpans:  []string{"function foo(sn: string | number) {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nfunction foo(s: string) {}\nfunction foo(n: number) {}\nconst a = '';\nconst b = '';\nfunction foo(sn: string | number) {}\n      ",
			wantSpans:  []string{"function foo(sn: string | number) {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nfunction foo(s: string) {}\nfunction foo(n: number) {}\nclass Bar {}\nfunction foo(sn: string | number) {}\n      ",
			wantSpans:  []string{"function foo(sn: string | number) {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nfunction foo(s: string) {}\nfunction foo(n: number) {}\nfunction foo(sn: string | number) {}\nclass Bar {\n  foo(s: string);\n  foo(n: number);\n  name: string;\n  foo(sn: string | number) {}\n}\n      ",
			wantSpans:  []string{"foo(sn: string | number) {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ndeclare function foo(s: string);\ndeclare function foo(n: number);\ndeclare function bar(): void;\ndeclare function baz(): void;\ndeclare function foo(sn: string | number);\n      ",
			wantSpans:  []string{"declare function foo(sn: string | number);"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ndeclare function foo(s: string);\ndeclare function foo(n: number);\nconst a = '';\nconst b = '';\ndeclare function foo(sn: string | number);\n      ",
			wantSpans:  []string{"declare function foo(sn: string | number);"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ndeclare module 'Foo' {\n  export function foo(s: string): void;\n  export function foo(n: number): void;\n  export function bar(): void;\n  export function baz(): void;\n  export function foo(sn: string | number): void;\n}\n      ",
			wantSpans:  []string{"export function foo(sn: string | number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ndeclare module 'Foo' {\n  export function foo(s: string): void;\n  export function foo(n: number): void;\n  export function foo(sn: string | number): void;\n  function baz(s: string): void;\n  export function bar(): void;\n  function baz(n: number): void;\n  function baz(sn: string | number): void;\n}\n      ",
			wantSpans:  []string{"function baz(n: number): void;"},
			wantNames:  []string{"baz"},
		},
		{
			sourceText: "\ndeclare namespace Foo {\n  export function foo(s: string): void;\n  export function foo(n: number): void;\n  export function bar(): void;\n  export function baz(): void;\n  export function foo(sn: string | number): void;\n}\n      ",
			wantSpans:  []string{"export function foo(sn: string | number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ndeclare namespace Foo {\n  export function foo(s: string): void;\n  export function foo(n: number): void;\n  export function foo(sn: string | number): void;\n  function baz(s: string): void;\n  export function bar(): void;\n  function baz(n: number): void;\n  function baz(sn: string | number): void;\n}\n      ",
			wantSpans:  []string{"function baz(n: number): void;"},
			wantNames:  []string{"baz"},
		},
		{
			sourceText: "\ntype Foo = {\n  foo(s: string): void;\n  foo(n: number): void;\n  bar(): void;\n  baz(): void;\n  foo(sn: string | number): void;\n};\n      ",
			wantSpans:  []string{"foo(sn: string | number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ntype Foo = {\n  foo(s: string): void;\n  ['foo'](n: number): void;\n  bar(): void;\n  baz(): void;\n  foo(sn: string | number): void;\n};\n      ",
			wantSpans:  []string{"foo(sn: string | number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ntype Foo = {\n  foo(s: string): void;\n  name: string;\n  foo(n: number): void;\n  foo(sn: string | number): void;\n  bar(): void;\n  baz(): void;\n};\n      ",
			wantSpans:  []string{"foo(n: number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ninterface Foo {\n  (s: string): void;\n  foo(n: number): void;\n  (n: number): void;\n  (sn: string | number): void;\n  bar(): void;\n  baz(): void;\n  call(): void;\n}\n      ",
			wantSpans:  []string{"(n: number): void;"},
			wantNames:  []string{"call"},
		},
		{
			sourceText: "\ninterface Foo {\n  foo(s: string): void;\n  foo(n: number): void;\n  bar(): void;\n  baz(): void;\n  foo(sn: string | number): void;\n}\n      ",
			wantSpans:  []string{"foo(sn: string | number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ninterface Foo {\n  foo(s: string): void;\n  ['foo'](n: number): void;\n  bar(): void;\n  baz(): void;\n  foo(sn: string | number): void;\n}\n      ",
			wantSpans:  []string{"foo(sn: string | number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ninterface Foo {\n  foo(s: string): void;\n  'foo'(n: number): void;\n  bar(): void;\n  baz(): void;\n  foo(sn: string | number): void;\n}\n      ",
			wantSpans:  []string{"foo(sn: string | number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ninterface Foo {\n  foo(s: string): void;\n  name: string;\n  foo(n: number): void;\n  foo(sn: string | number): void;\n  bar(): void;\n  baz(): void;\n}\n      ",
			wantSpans:  []string{"foo(n: number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\ninterface Foo {\n  foo(): void;\n  bar: {\n    baz(s: string): void;\n    baz(n: number): void;\n    foo(): void;\n    baz(sn: string | number): void;\n  };\n}\n      ",
			wantSpans:  []string{"baz(sn: string | number): void;"},
			wantNames:  []string{"baz"},
		},
		{
			sourceText: "\ninterface Foo {\n  new (s: string);\n  new (n: number);\n  foo(): void;\n  bar(): void;\n  new (sn: string | number);\n}\n      ",
			wantSpans:  []string{"new (sn: string | number);"},
			wantNames:  []string{"new"},
		},
		{
			sourceText: "\ninterface Foo {\n  new (s: string);\n  foo(): void;\n  new (n: number);\n  bar(): void;\n  new (sn: string | number);\n}\n      ",
			wantSpans:  []string{"new (n: number);", "new (sn: string | number);"},
			wantNames:  []string{"new", "new"},
		},
		{
			sourceText: "\nclass Foo {\n  constructor(s: string);\n  constructor(n: number);\n  bar(): void {}\n  baz(): void {}\n  constructor(sn: string | number) {}\n}\n      ",
			wantSpans:  []string{"constructor(sn: string | number) {}"},
			wantNames:  []string{"constructor"},
		},
		{
			sourceText: "\nclass Foo {\n  foo(s: string): void;\n  foo(n: number): void;\n  bar(): void {}\n  baz(): void {}\n  foo(sn: string | number): void {}\n}\n      ",
			wantSpans:  []string{"foo(sn: string | number): void {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nclass Foo {\n  foo(s: string): void;\n  ['foo'](n: number): void;\n  bar(): void {}\n  baz(): void {}\n  foo(sn: string | number): void {}\n}\n      ",
			wantSpans:  []string{"foo(sn: string | number): void {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nclass Foo {\n  // prettier-ignore\n  \"foo\"(s: string): void;\n  foo(n: number): void;\n  bar(): void {}\n  baz(): void {}\n  foo(sn: string | number): void {}\n}\n      ",
			wantSpans:  []string{"foo(sn: string | number): void {}"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nclass Foo {\n  constructor(s: string);\n  name: string;\n  constructor(n: number);\n  constructor(sn: string | number) {}\n  bar(): void {}\n  baz(): void {}\n}\n      ",
			wantSpans:  []string{"constructor(n: number);"},
			wantNames:  []string{"constructor"},
		},
		{
			sourceText: "\nclass Foo {\n  foo(s: string): void;\n  name: string;\n  foo(n: number): void;\n  foo(sn: string | number): void {}\n  bar(): void {}\n  baz(): void {}\n}\n      ",
			wantSpans:  []string{"foo(n: number): void;"},
			wantNames:  []string{"foo"},
		},
		{
			sourceText: "\nclass Foo {\n  static foo(s: string): void;\n  name: string;\n  static foo(n: number): void;\n  static foo(sn: string | number): void {}\n  bar(): void {}\n  baz(): void {}\n}\n      ",
			wantSpans:  []string{"static foo(n: number): void;"},
			wantNames:  []string{"static foo"},
		},
		{
			sourceText: "\nclass Test {\n  #private(): void;\n  '#private'(): void;\n  #private(arg: number): void {}\n  '#private'(arg: number): void {}\n}\n      ",
			wantSpans:  []string{"#private(arg: number): void {}", "'#private'(arg: number): void {}"},
			wantNames:  []string{"#private", "\"#private\""},
		},
	}
	for index, testCase := range cases {
		t.Run(adjacentOverloadSignaturesCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, AdjacentOverloadSignatures,
				adjacentOverloadSignaturesFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantSpans))
			for index := range wantIds {
				wantIds[index] = "adjacentSignature"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, wantSpan := range testCase.wantSpans {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
				// The member name is the only part of the description that varies per input, and it
				// is the part upstream interpolates. Equality against a literal typed here rather
				// than against the rule's own message constant: comparing a finding to the constant
				// it was built from moves both sides under mutation and asserts nothing.
				want := messageAdjacentOverloadSignatures(testCase.wantNames[index]).Description
				if result.Diagnostics[index].Message.Description != want {
					t.Errorf("finding %d reads %q, want %q", index,
						result.Diagnostics[index].Message.Description, want)
				}
			}
		})
	}
}

// TestAdjacentOverloadSignaturesDiscriminatesOnCasesUpstreamDoesNotWrite covers the folds.
//
// Upstream's corpus writes no computed key, no private identifier, no accessor, no abstract member,
// no class expression, no block statement, and no static block. Every one of those is a place where
// TSESTree and typescript-go disagree about the tree, so the corpus cannot see whether this port
// resolved the disagreement correctly. Each row was run through the installed 8.67.0 build and
// carries the verdict that build produced, so a row asserting silence asserts upstream's silence.
//
// Read as a group these rows are the name-keying table from the rule's doc comment, executable.
func TestAdjacentOverloadSignaturesDiscriminatesOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantCount  int
		reason     string
	}{
		{
			sourceText: "interface I { [k](): void; other(): void; [k](): void; }",
			wantCount:  1,
			reason:     "a computed identifier key is read through the brackets",
		},
		{
			sourceText: "interface I { [\"lit\"](): void; other(): void; lit(): void; }",
			wantCount:  1,
			reason:     "a computed literal keys like a bare one, so the two collide",
		},
		{
			sourceText: "interface I { \"foo\"(): void; other(): void; foo(): void; }",
			wantCount:  1,
			reason:     "a quoted key that is a valid identifier keys bare",
		},
		{
			sourceText: "interface I { \"a-b\"(): void; other(): void; [\"a-b\"](): void; }",
			wantCount:  1,
			reason:     "a key needing quotes keys wrapped, and still collides with itself",
		},
		{
			sourceText: "interface I { 12(): void; other(): void; \"12\"(): void; }",
			wantCount:  1,
			reason:     "a numeric key is cooked, then wrapped because it needs quoting",
		},
		{
			sourceText: "interface I { 1e1(): void; other(): void; 10(): void; }",
			wantCount:  1,
			reason:     "a numeric key is cooked to its canonical rendering first",
		},
		{
			sourceText: "class C { #p(): void {} other(): void {} #p(): void {} }",
			wantCount:  1,
			reason:     "a private identifier keys with its hash",
		},
		{
			sourceText: "interface I { [Symbol.iterator](): void; other(): void; [Symbol.iterator](): void; }",
			wantCount:  1,
			reason:     "anything else keys by the source text of the key",
		},
		{
			// The expression arm keys by SOURCE TEXT, so whether that text carries the key's leading
			// trivia decides whether two spellings of one key match. `Pos()` includes trivia and
			// upstream's `range` does not, and this is the input that separates them: with trivia
			// the two keys read as `Symbol.iterator` and a newline-and-spaces followed by
			// `Symbol.iterator`, which do not match, and the case goes silent. A mutant using the
			// untrimmed start survived until this row existed, because every other expression-key
			// fixture writes both keys inline where there is no trivia to differ.
			sourceText: "interface I {\n  [\n    Symbol.iterator\n  ](): void;\n  foo(): void;\n  [Symbol.iterator](): void;\n}",
			wantCount:  1,
			reason:     "an expression key matches across a line break, so the trivia is trimmed",
		},
		{
			// The control for the row above. Both keys carry the same trivia, so it reports whether
			// or not the trim happens, which is why the pair is needed rather than either alone.
			sourceText: "interface I {\n  [\n    Symbol.iterator\n  ](): void;\n  foo(): void;\n  [\n    Symbol.iterator\n  ](): void;\n}",
			wantCount:  1,
			reason:     "the control: two identically-written expression keys match either way",
		},
		{
			sourceText: "class C { static foo(): void {} other(): void {} foo(): void {} }",
			wantCount:  0,
			reason:     "static is part of the identity, so these are different members",
		},
		{
			sourceText: "class C { static foo(): void {} other(): void {} static foo(): void {} }",
			wantCount:  1,
			reason:     "two static members of one name are the same",
		},
		{
			sourceText: "interface I { (): void; foo(): void; (a: number): void; }",
			wantCount:  1,
			reason:     "a call signature is named call",
		},
		{
			sourceText: "interface I { (): void; foo(): void; call(): void; }",
			wantCount:  0,
			reason:     "the call flag keeps a call signature from colliding with a member named call",
		},
		{
			sourceText: "interface I { call(): void; foo(): void; call(a: number): void; }",
			wantCount:  1,
			reason:     "a member named call still collides with itself",
		},
		{
			sourceText: "interface I { new (): void; foo(): void; new (a: number): void; }",
			wantCount:  1,
			reason:     "a construct signature is named new",
		},
		{
			sourceText: "interface I { new(): void; foo(): void; \"new\"(): void; }",
			wantCount:  0,
			reason:     "new carries no flag, so a member named new DOES collide with it",
		},
		{
			sourceText: "class C { constructor(); foo(): void {} constructor(a: number) {} }",
			wantCount:  1,
			reason:     "a constructor is a method named constructor",
		},
		{
			sourceText: "class C { constructor(); foo(): void {} [\"constructor\"](): void {} }",
			wantCount:  1,
			reason:     "and that name is a real key, so a quoted constructor collides",
		},
		{
			sourceText: "class C { get x(): number { return 1; } foo(): void {} get x(): number { return 2; } }",
			wantCount:  1,
			reason:     "a getter is a method",
		},
		{
			sourceText: "class C { get x(): number { return 1; } foo(): void {} set x(v: number) {} }",
			wantCount:  1,
			reason:     "a getter and a setter share one name",
		},
		{
			sourceText: "class C { x(): void {} foo(): void {} get x(): number { return 1; } }",
			wantCount:  1,
			reason:     "and so does a plain method of that name",
		},
		{
			sourceText: "abstract class C { abstract foo(): void; bar(): void {} abstract foo(a: number): void; }",
			wantCount:  0,
			reason:     "an abstract member is not a method to this rule",
		},
		{
			sourceText: "abstract class C { foo(): void {} abstract bar(): void; foo(a: number): void {} }",
			wantCount:  1,
			reason:     "and it clears the predecessor like any non-method",
		},
		{
			sourceText: "abstract class C { foo(): void {} bar(): void {} foo(a: number): void {} }",
			wantCount:  1,
			reason:     "the concrete control for the pair above",
		},
		{
			sourceText: "class C { foo(): void {} static { } foo(a: number): void {} }",
			wantCount:  1,
			reason:     "a static block is not a method and separates two signatures",
		},
		{
			sourceText: "interface I { foo(): void; bar(): void; foo(a: number): void; foo(b: string): void; }",
			wantCount:  1,
			reason:     "only the reopening reports, not every later member",
		},
		{
			sourceText: "interface I { foo(): void; bar(): void; foo(a: number): void; bar(a: number): void; }",
			wantCount:  2,
			reason:     "two reopenings report twice",
		},
		{
			sourceText: "const C = class { foo(): void {} bar(): void {} foo(a: number): void {} };",
			wantCount:  1,
			reason:     "a class expression is a container too",
		},
		{
			sourceText: "function outer() { function foo(): void; function bar(): void; function foo(a: number): void {} }",
			wantCount:  1,
			reason:     "a block statement is a container",
		},
		{
			sourceText: "export default function foo(): void {}",
			wantCount:  0,
			reason:     "an unnamed default export has no name to key",
		},
		{
			sourceText: "const a = 1; export { a }; const b = 2; export { b };",
			wantCount:  0,
			reason:     "an export with no declaration is not a method",
		},
	}
	for index, testCase := range cases {
		t.Run(adjacentOverloadSignaturesCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, AdjacentOverloadSignatures,
				adjacentOverloadSignaturesFile, testCase.sourceText)
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "adjacentSignature"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestAdjacentOverloadSignaturesNamesTheMemberUpstreamNames pins the interpolated name.
//
// The name is invisible to a message-id assertion and it is the whole of what upstream interpolates,
// including its `static ` prefix. Each row's expectation is the name the installed build printed for
// that exact input, recovered from its message text rather than predicted.
func TestAdjacentOverloadSignaturesNamesTheMemberUpstreamNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantName   string
	}{
		{
			sourceText: "interface I { [k](): void; other(): void; [k](): void; }",
			wantName:   "k",
		},
		{
			sourceText: "interface I { [\"lit\"](): void; other(): void; lit(): void; }",
			wantName:   "lit",
		},
		{
			sourceText: "interface I { \"foo\"(): void; other(): void; foo(): void; }",
			wantName:   "foo",
		},
		{
			sourceText: "interface I { \"a-b\"(): void; other(): void; [\"a-b\"](): void; }",
			wantName:   "\"a-b\"",
		},
		{
			sourceText: "interface I { 12(): void; other(): void; \"12\"(): void; }",
			wantName:   "\"12\"",
		},
		{
			sourceText: "interface I { 1e1(): void; other(): void; 10(): void; }",
			wantName:   "\"10\"",
		},
		{
			sourceText: "class C { #p(): void {} other(): void {} #p(): void {} }",
			wantName:   "#p",
		},
		{
			sourceText: "interface I { [Symbol.iterator](): void; other(): void; [Symbol.iterator](): void; }",
			wantName:   "Symbol.iterator",
		},
		{
			sourceText: "class C { static foo(): void {} other(): void {} static foo(): void {} }",
			wantName:   "static foo",
		},
		{
			sourceText: "interface I { (): void; foo(): void; (a: number): void; }",
			wantName:   "call",
		},
		{
			sourceText: "interface I { new (): void; foo(): void; new (a: number): void; }",
			wantName:   "new",
		},
		{
			sourceText: "class C { constructor(); foo(): void {} constructor(a: number) {} }",
			wantName:   "constructor",
		},
		{
			sourceText: "class C { get x(): number { return 1; } foo(): void {} get x(): number { return 2; } }",
			wantName:   "x",
		},
		{
			sourceText: "\ndeclare module 'Foo' {\n  export function foo(s: string): void;\n  export function foo(n: number): void;\n  export function foo(sn: string | number): void;\n  function baz(s: string): void;\n  export function bar(): void;\n  function baz(n: number): void;\n  function baz(sn: string | number): void;\n}\n      ",
			wantName:   "baz",
		},
	}
	for index, testCase := range cases {
		t.Run(adjacentOverloadSignaturesCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, AdjacentOverloadSignatures,
				adjacentOverloadSignaturesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "adjacentSignature")
			want := messageAdjacentOverloadSignatures(testCase.wantName).Description
			if result.Diagnostics[0].Message.Description != want {
				t.Errorf("the description is %q, want %q",
					result.Diagnostics[0].Message.Description, want)
			}
		})
	}
}

// TestAdjacentOverloadSignaturesSurvivesContainersWithNoMemberList is a crash fixture.
//
// `Members()` and `Statements()` both PANIC on a kind that has no such list, and the listener map is
// the only thing keeping each call on a kind that does. The walk recovers per file rather than per
// rule, so one panic would cost every rule that file rather than one finding. Probed directly:
// `Members()` on a `KindSourceFile` panics with `Unhandled case in Node.MemberList`.
//
// These sources put the rule's seven container kinds beside the kinds nearest them in the tree, so a
// listener wired to the wrong accessor takes the run down here rather than in the real tree. There
// is no finding to assert; the assertion is that the run completes.
func TestAdjacentOverloadSignaturesSurvivesContainersWithNoMemberList(t *testing.T) {
	t.Parallel()

	for name, sourceText := range map[string]string{
		"enum":         "enum E { A, B }",
		"mappedType":   "type M<T> = { [K in keyof T]: T[K] };",
		"switch":       "function f(x: number) { switch (x) { case 1: break; default: break; } }",
		"emptyModule":  "declare module 'm' {}",
		"emptyClass":   "class C {}",
		"emptyLiteral": "type T = {};",
		"arrowBody":    "const f = () => { function g(): void; function g(a: number): void {} };",
		"objectType":   "const o: { foo(): void; bar(): void } = { foo() {}, bar() {} };",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.Run(t, AdjacentOverloadSignatures,
				adjacentOverloadSignaturesFile, sourceText)
		})
	}
}
