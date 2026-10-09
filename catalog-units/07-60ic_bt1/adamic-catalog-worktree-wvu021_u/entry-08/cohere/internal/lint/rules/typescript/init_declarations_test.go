package typescript

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// initDeclarationsFile names the fixture file.
//
// The rule reads no path and gates on no extension.
const initDeclarationsFile = "/repository/source/Declarations.ts"

// initDeclarationsCaseName numbers a row so a failure names which one.
func initDeclarationsCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// decodeInitDeclarationsOptionsForTest routes a fixture through the rule's own decoder.
//
// The wire shape is the unusual part of this rule and the reason every option fixture goes through
// the decoder rather than building the struct. Upstream's schema is a positional list whose first
// element is a bare mode string, and the config layer hands this rule that list whole. It used to
// store only the first element after the severity, so the first version of this decoder read a list,
// every fixture passed, and it failed at run time on the first real configuration. The fixtures here
// pass the list, which is now what the config layer delivers; the decoder test pins that it does.
func decodeInitDeclarationsOptionsForTest(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeInitDeclarationsOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return decoded
}

// runInitDeclarations runs one case, with or without options.
func runInitDeclarations(t *testing.T, sourceText string, options string) rule_testing.Result {
	t.Helper()
	if options == "" {
		return rule_testing.Run(t, InitDeclarations, initDeclarationsFile, sourceText)
	}
	return rule_testing.RunWithOptions(t, InitDeclarations, initDeclarationsFile, sourceText,
		decodeInitDeclarationsOptionsForTest(t, options))
}

// TestInitDeclarationsStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All fifty-one of upstream's passing inputs, extracted from the clone's test file by parsing it
// with the TypeScript compiler rather than by reading it, then byte verified along with every one of
// their option arrays. Every case was replayed through the installed 8.x build, which reported
// nothing on all fifty-one.
//
// The same source appears in this list and in the failing one under the opposite mode, which is why
// the options travel with each row rather than being set once for the file.
func TestInitDeclarationsStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
	}{
		{sourceText: "var foo = null;", options: ""},
		{sourceText: "foo = true;", options: ""},
		{sourceText: "\nvar foo = 1,\n  bar = false,\n  baz = {};\n    ", options: ""},
		{sourceText: "\nfunction foo() {\n  var foo = 0;\n  var bar = [];\n}\n    ", options: ""},
		{sourceText: "var fn = function () {};", options: ""},
		{sourceText: "var foo = (bar = 2);", options: ""},
		{sourceText: "for (var i = 0; i < 1; i++) {}", options: ""},
		{sourceText: "\nfor (var foo in []) {\n}\n    ", options: ""},
		{sourceText: "\nfor (var foo of []) {\n}\n      ", options: ""},
		{sourceText: "let a = true;", options: "[\"always\"]"},
		{sourceText: "const a = {};", options: "[\"always\"]"},
		{sourceText: "\nfunction foo() {\n  let a = 1,\n    b = false;\n  if (a) {\n    let c = 3,\n      d = null;\n  }\n}\n      ", options: "[\"always\"]"},
		{sourceText: "\nfunction foo() {\n  const a = 1,\n    b = true;\n  if (a) {\n    const c = 3,\n      d = null;\n  }\n}\n      ", options: "[\"always\"]"},
		{sourceText: "\nfunction foo() {\n  let a = 1;\n  const b = false;\n  var c = true;\n}\n      ", options: "[\"always\"]"},
		{sourceText: "var foo;", options: "[\"never\"]"},
		{sourceText: "var foo, bar, baz;", options: "[\"never\"]"},
		{sourceText: "\nfunction foo() {\n  var foo;\n  var bar;\n}\n      ", options: "[\"never\"]"},
		{sourceText: "let a;", options: "[\"never\"]"},
		{sourceText: "const a = 1;", options: "[\"never\"]"},
		{sourceText: "\nfunction foo() {\n  let a, b;\n  if (a) {\n    let c, d;\n  }\n}\n      ", options: "[\"never\"]"},
		{sourceText: "\nfunction foo() {\n  const a = 1,\n    b = true;\n  if (a) {\n    const c = 3,\n      d = null;\n  }\n}\n      ", options: "[\"never\"]"},
		{sourceText: "\nfunction foo() {\n  let a;\n  const b = false;\n  var c;\n}\n      ", options: "[\"never\"]"},
		{sourceText: "for (var i = 0; i < 1; i++) {}", options: "[\"never\", {\"ignoreForLoopInit\": true}]"},
		{sourceText: "\nfor (var foo in []) {\n}\n      ", options: "[\"never\", {\"ignoreForLoopInit\": true}]"},
		{sourceText: "\nfor (var foo of []) {\n}\n      ", options: "[\"never\", {\"ignoreForLoopInit\": true}]"},
		{sourceText: "\nfunction foo() {\n  var bar = 1;\n  let baz = 2;\n  const qux = 3;\n}\n      ", options: "[\"always\"]"},
		{sourceText: "declare const foo: number;", options: "[\"always\"]"},
		{sourceText: "declare const foo: number;", options: "[\"never\"]"},
		{sourceText: "\ndeclare namespace myLib {\n  let numberOfGreetings: number;\n}\n      ", options: "[\"always\"]"},
		{sourceText: "\ndeclare namespace myLib {\n  let numberOfGreetings: number;\n}\n      ", options: "[\"never\"]"},
		{sourceText: "\ninterface GreetingSettings {\n  greeting: string;\n  duration?: number;\n  color?: string;\n}\n      ", options: ""},
		{sourceText: "\ninterface GreetingSettings {\n  greeting: string;\n  duration?: number;\n  color?: string;\n}\n      ", options: "[\"never\"]"},
		{sourceText: "type GreetingLike = string | (() => string) | Greeter;", options: ""},
		{sourceText: "type GreetingLike = string | (() => string) | Greeter;", options: "[\"never\"]"},
		{sourceText: "\nfunction foo() {\n  var bar: string;\n}\n      ", options: "[\"never\"]"},
		{sourceText: "var bar: string;", options: "[\"never\"]"},
		{sourceText: "\nvar bar: string = function (): string {\n  return 'string';\n};\n      ", options: "[\"always\"]"},
		{sourceText: "\nvar bar: string = function (arg1: string): string {\n  return 'string';\n};\n      ", options: "[\"always\"]"},
		{sourceText: "function foo(arg1: string = 'string'): void {}", options: "[\"never\"]"},
		{sourceText: "const foo: string = 'hello';", options: "[\"never\"]"},
		{sourceText: "\nconst class1 = class NAME {\n  constructor() {\n    var name1: string = 'hello';\n  }\n};\n      ", options: ""},
		{sourceText: "\nconst class1 = class NAME {\n  static pi: number = 3.14;\n};\n      ", options: ""},
		{sourceText: "\nconst class1 = class NAME {\n  static pi: number = 3.14;\n};\n      ", options: "[\"never\"]"},
		{sourceText: "\ninterface IEmployee {\n  empCode: number;\n  empName: string;\n  getSalary: (number) => number; // arrow function\n  getManagerName(number): string;\n}\n      ", options: ""},
		{sourceText: "\ninterface IEmployee {\n  empCode: number;\n  empName: string;\n  getSalary: (number) => number; // arrow function\n  getManagerName(number): string;\n}\n      ", options: "[\"never\"]"},
		{sourceText: "const foo: number = 'asd';", options: "[\"always\"]"},
		{sourceText: "const foo: number;", options: "[\"never\"]"},
		{sourceText: "\nnamespace myLib {\n  let numberOfGreetings: number;\n}\n      ", options: "[\"never\"]"},
		{sourceText: "\nnamespace myLib {\n  let numberOfGreetings: number = 2;\n}\n      ", options: "[\"always\"]"},
		{sourceText: "\ndeclare namespace myLib1 {\n  const foo: number;\n  namespace myLib2 {\n    let bar: string;\n    namespace myLib3 {\n      let baz: object;\n    }\n  }\n}\n      ", options: "[\"always\"]"},
		{sourceText: "\ndeclare namespace myLib1 {\n  const foo: number;\n  namespace myLib2 {\n    let bar: string;\n    namespace myLib3 {\n      let baz: object;\n    }\n  }\n}\n      ", options: "[\"never\"]"},
	}
	for index, testCase := range cases {
		t.Run(initDeclarationsCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runInitDeclarations(t, testCase.sourceText, testCase.options))
		})
	}
}

// TestInitDeclarationsFiresOnUpstreamFailCases is the imported failing corpus, verbatim, with the
// span and the rendered message of every finding asserted.
//
// The span is the whole reason this rule exists as a TypeScript wrapper rather than as the core rule
// alone. Under "always" the finding covers the NAME only, so `let arr: string;` reports three
// characters and leaves the annotation out; under "never" it covers the whole declarator including
// the annotation and the initializer. Two spans from one rule, decided by the mode, and no message
// id assertion can tell them apart.
//
// The message interpolates the variable name, so it is asserted as text too. That is the one place a
// format string could be wrong while every id and every span stayed right.
func TestInitDeclarationsFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		options      string
		wantIds      []string
		wantSpans    []string
		wantMessages []string
	}{
		{
			sourceText:   "var foo;",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized"},
			wantSpans:    []string{"foo"},
			wantMessages: []string{"Variable 'foo' should be initialized on declaration."},
		},
		{
			sourceText:   "for (var a in []) var foo;",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized"},
			wantSpans:    []string{"foo"},
			wantMessages: []string{"Variable 'foo' should be initialized on declaration."},
		},
		{
			sourceText:   "\nvar foo,\n  bar = false,\n  baz;\n      ",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized", "initialized"},
			wantSpans:    []string{"foo", "baz"},
			wantMessages: []string{"Variable 'foo' should be initialized on declaration.", "Variable 'baz' should be initialized on declaration."},
		},
		{
			sourceText:   "\nfunction foo() {\n  var foo = 0;\n  var bar;\n}\n      ",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized"},
			wantSpans:    []string{"bar"},
			wantMessages: []string{"Variable 'bar' should be initialized on declaration."},
		},
		{
			sourceText:   "\nfunction foo() {\n  var foo;\n  var bar = foo;\n}\n      ",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized"},
			wantSpans:    []string{"foo"},
			wantMessages: []string{"Variable 'foo' should be initialized on declaration."},
		},
		{
			sourceText:   "let a;",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized"},
			wantSpans:    []string{"a"},
			wantMessages: []string{"Variable 'a' should be initialized on declaration."},
		},
		{
			sourceText:   "\nfunction foo() {\n  let a = 1,\n    b;\n  if (a) {\n    let c = 3,\n      d = null;\n  }\n}\n      ",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized"},
			wantSpans:    []string{"b"},
			wantMessages: []string{"Variable 'b' should be initialized on declaration."},
		},
		{
			sourceText:   "\nfunction foo() {\n  let a;\n  const b = false;\n  var c;\n}\n      ",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized", "initialized"},
			wantSpans:    []string{"a", "c"},
			wantMessages: []string{"Variable 'a' should be initialized on declaration.", "Variable 'c' should be initialized on declaration."},
		},
		{
			sourceText:   "var foo = (bar = 2);",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"foo = (bar = 2)"},
			wantMessages: []string{"Variable 'foo' should not be initialized on declaration."},
		},
		{
			sourceText:   "var foo = true;",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"foo = true"},
			wantMessages: []string{"Variable 'foo' should not be initialized on declaration."},
		},
		{
			sourceText:   "\nvar foo,\n  bar = 5,\n  baz = 3;\n      ",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized", "notInitialized"},
			wantSpans:    []string{"bar = 5", "baz = 3"},
			wantMessages: []string{"Variable 'bar' should not be initialized on declaration.", "Variable 'baz' should not be initialized on declaration."},
		},
		{
			sourceText:   "\nfunction foo() {\n  var foo;\n  var bar = foo;\n}\n      ",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"bar = foo"},
			wantMessages: []string{"Variable 'bar' should not be initialized on declaration."},
		},
		{
			sourceText:   "let a = 1;",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"a = 1"},
			wantMessages: []string{"Variable 'a' should not be initialized on declaration."},
		},
		{
			sourceText:   "\nfunction foo() {\n  let a = 'foo',\n    b;\n  if (a) {\n    let c, d;\n  }\n}\n      ",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"a = 'foo'"},
			wantMessages: []string{"Variable 'a' should not be initialized on declaration."},
		},
		{
			sourceText:   "\nfunction foo() {\n  let a;\n  const b = false;\n  var c = 1;\n}\n      ",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"c = 1"},
			wantMessages: []string{"Variable 'c' should not be initialized on declaration."},
		},
		{
			sourceText:   "for (var i = 0; i < 1; i++) {}",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"i = 0"},
			wantMessages: []string{"Variable 'i' should not be initialized on declaration."},
		},
		{
			sourceText:   "\nfor (var foo in []) {\n}\n      ",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"foo"},
			wantMessages: []string{"Variable 'foo' should not be initialized on declaration."},
		},
		{
			sourceText:   "\nfor (var foo of []) {\n}\n      ",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"foo"},
			wantMessages: []string{"Variable 'foo' should not be initialized on declaration."},
		},
		{
			sourceText:   "\nfunction foo() {\n  var bar;\n}\n      ",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized"},
			wantSpans:    []string{"bar"},
			wantMessages: []string{"Variable 'bar' should be initialized on declaration."},
		},
		{
			sourceText:   "let arr: string[] = ['arr', 'ar'];",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"arr: string[] = ['arr', 'ar']"},
			wantMessages: []string{"Variable 'arr' should not be initialized on declaration."},
		},
		{
			sourceText:   "let arr: string = function () {};",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"arr: string = function () {}"},
			wantMessages: []string{"Variable 'arr' should not be initialized on declaration."},
		},
		{
			sourceText:   "\nconst class1 = class NAME {\n  constructor() {\n    var name1: string = 'hello';\n  }\n};\n      ",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"name1: string = 'hello'"},
			wantMessages: []string{"Variable 'name1' should not be initialized on declaration."},
		},
		{
			sourceText:   "let arr: string;",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized"},
			wantSpans:    []string{"arr"},
			wantMessages: []string{"Variable 'arr' should be initialized on declaration."},
		},
		{
			sourceText:   "\nnamespace myLib {\n  let numberOfGreetings: number;\n}\n      ",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized"},
			wantSpans:    []string{"numberOfGreetings"},
			wantMessages: []string{"Variable 'numberOfGreetings' should be initialized on declaration."},
		},
		{
			sourceText:   "\nnamespace myLib {\n  let numberOfGreetings: number = 2;\n}\n      ",
			options:      "[\"never\"]",
			wantIds:      []string{"notInitialized"},
			wantSpans:    []string{"numberOfGreetings: number = 2"},
			wantMessages: []string{"Variable 'numberOfGreetings' should not be initialized on declaration."},
		},
		{
			sourceText:   "\nnamespace myLib1 {\n  const foo: number;\n  namespace myLib2 {\n    let bar: string;\n    namespace myLib3 {\n      let baz: object;\n    }\n  }\n}\n      ",
			options:      "[\"always\"]",
			wantIds:      []string{"initialized", "initialized", "initialized"},
			wantSpans:    []string{"foo", "bar", "baz"},
			wantMessages: []string{"Variable 'foo' should be initialized on declaration.", "Variable 'bar' should be initialized on declaration.", "Variable 'baz' should be initialized on declaration."},
		},
	}
	for index, testCase := range cases {
		t.Run(initDeclarationsCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := runInitDeclarations(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q", findingIndex, gotSpan, wantSpan)
				}
				if reported.Message.Description != testCase.wantMessages[findingIndex] {
					t.Errorf("finding %d reads %q, wanted %q", findingIndex,
						reported.Message.Description, testCase.wantMessages[findingIndex])
				}
			}
		})
	}
}

// TestInitDeclarationsLeavesDestructuringPatternsAlone pins the identifier-only guard.
//
// Upstream reports only on an identifier binding and says at the line that its span narrowing
// depends on that, which is true here too: the narrowed span is the identifier's own text and a
// pattern has none to take. But its corpus never writes a pattern WITHOUT an initializer, so the
// guard decides nothing there and a mutation removing it survives all seventy-seven imported rows.
//
// Measured against the installed 8.x build, with a reporting control for each mode in the same runs.
func TestInitDeclarationsLeavesDestructuringPatternsAlone(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		options    string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "object-pattern-uninitialized",
			why:        "an object destructuring pattern with no initializer, which upstream declines because it reports only on an identifier binding. Its corpus never writes a pattern without one, so the guard is unexercised there and a port without it reports on a shape that has no name to underline",
			sourceText: "let { a, b };\n",
			options:    "[\"always\"]",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "array-pattern-uninitialized",
			why:        "the array form of the row above, which is a different node kind here",
			sourceText: "let [a, b];\n",
			options:    "[\"always\"]",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "object-pattern-in-for-of",
			why:        "a pattern in a for-of head, where the loop supplies the value and the pattern guard and the loop-position test would both have to fail for this to report",
			sourceText: "for (const { a } of list) {\n}\n",
			options:    "[\"always\"]",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "object-pattern-initialized-never",
			why:        "a pattern WITH an initializer under never, which is the mode where the whole declarator is the span and where a port reaching for the name would have none to take",
			sourceText: "let { a, b } = obj;\n",
			options:    "[\"never\"]",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "array-pattern-initialized-never",
			why:        "the array form of the row above",
			sourceText: "let [a, b] = arr;\n",
			options:    "[\"never\"]",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "control-d5-always",
			why:        "the control for the always rows, an ordinary identifier that reports",
			sourceText: "let foo;\n",
			options:    "[\"always\"]",
			wantIds:    []string{"initialized"},
			wantSpans:  []string{"foo"},
		},
		{
			name:       "control-d5-never",
			why:        "the control for the never rows",
			sourceText: "let foo = 1;\n",
			options:    "[\"never\"]",
			wantIds:    []string{"notInitialized"},
			wantSpans:  []string{"foo = 1"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runInitDeclarations(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestInitDeclarationsExemptsEveryConstantBindingUnderNever pins the third and fourth members of a
// set upstream keeps and its corpus only half exercises.
//
// The core rule holds const, using and await using together, because a binding in any of those forms
// must carry an initializer and reporting it for doing so would be asking for source that does not
// compile. Upstream writes only the const case, so a port exempting const alone passes every
// imported row, and our parser puts all three on one flag field where await using is the using flag
// with a second bit set, which makes a naive flag test miss it specifically.
//
// Measured against the installed 8.x build with a reporting control in the same run.
func TestInitDeclarationsExemptsEveryConstantBindingUnderNever(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "using-initialized-never",
			why:        "a using binding under never, which upstream exempts alongside const because such a binding MUST be initialized and reporting it would be asking for source that does not compile. Its corpus writes neither using nor await using, so a port exempting const alone survives all seventy-seven imported rows",
			sourceText: "using x = f();\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "await-using-initialized-never",
			why:        "the await using form, which our parser spells as the using flag with another bit set rather than as a separate one, so a flag test naming only using would miss it",
			sourceText: "async function g() {\n  await using y = f();\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "const-initialized-never",
			why:        "the const half, which upstream's corpus does cover, kept beside the two above so the three read as one decision",
			sourceText: "const z = 1;\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "let-initialized-never",
			why:        "the control: a let binding in the same shape, which reports",
			sourceText: "let w = 1;\n",
			wantIds:    []string{"notInitialized"},
			wantSpans:  []string{"w = 1"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runInitDeclarations(t, testCase.sourceText, `["never"]`)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := testCase.sourceText[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestInitDeclarationsDecoderReadsThePositionalArray pins the lines of the decoder that have no
// upstream counterpart.
//
// The wire shape is a positional list rather than an object, so the mode has no key name and the
// second element is optional. Three of those lines decide something no rule fixture can see: an
// empty list, a value outside the schema's enum, and an absent second element. Two mutations of
// them survived the whole imported corpus, because every option row there names a mode explicitly.
func TestInitDeclarationsDecoderReadsThePositionalArray(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw  string
		want InitDeclarationsOptions
	}{
		// Measured on the installed build: `["error"]` with no mode reports nothing at all on
		// `var foo; var bar = 1;`, the same as a bare `"error"`.
		{raw: `[]`, want: InitDeclarationsOptions{Mode: InitDeclarationsUnconfigured}},
		{raw: `["always"]`, want: InitDeclarationsOptions{Mode: InitDeclarationsAlways}},
		{raw: `["never"]`, want: InitDeclarationsOptions{Mode: InitDeclarationsNever}},
		{
			raw:  `["never", {"ignoreForLoopInit": true}]`,
			want: InitDeclarationsOptions{Mode: InitDeclarationsNever, IgnoreForLoopInit: true},
		},
		{
			raw:  `["never", {"ignoreForLoopInit": false}]`,
			want: InitDeclarationsOptions{Mode: InitDeclarationsNever},
		},
		{raw: `["never", {}]`, want: InitDeclarationsOptions{Mode: InitDeclarationsNever}},
	}
	for _, testCase := range cases {
		t.Run(testCase.raw, func(t *testing.T) {
			t.Parallel()
			decoded := decodeInitDeclarationsOptionsForTest(t, testCase.raw)
			settings, isSettings := decoded.(InitDeclarationsOptions)
			if !isSettings {
				t.Fatalf("the decoder returned %T rather than the rule's own options type", decoded)
			}
			if settings != testCase.want {
				t.Errorf("decoded to %+v, wanted %+v", settings, testCase.want)
			}
		})
	}
}

// TestInitDeclarationsDecoderRefusesWhatUpstreamRefuses is the other half of the table above.
//
// A mode outside the schema's enum is a CONFIGURATION error upstream and never reaches the rule:
// ESLint refuses the key outright rather than running with a fallback. Measured. This decoder used to
// resolve it to the inert unconfigured mode, the closest spelling of "this rule never ran" it had,
// which is silence; the config layer now surfaces a decoder error at startup, so it is refused.
func TestInitDeclarationsDecoderRefusesWhatUpstreamRefuses(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		`["bogus"]`,
		// A second element beside "always", which upstream's anyOf does not allow.
		`["always", {"ignoreForLoopInit": true}]`,
		// A key upstream does not declare, and a third element.
		`["never", {"ignoreForLoopInits": true}]`,
		`["never", {"ignoreForLoopInit": true}, "always"]`,
		// The two one-slot workarounds from when the config layer delivered only the first element.
		`"never"`,
		`{"mode": "never", "ignoreForLoopInit": true}`,
	} {
		if decoded, err := DecodeInitDeclarationsOptions(json.RawMessage(raw)); err == nil {
			t.Errorf("%s decoded to %+v; it must be refused", raw, decoded)
		}
	}
}

// TestInitDeclarationsIsInertWhenNamedWithoutAMode reproduces upstream's most consequential behavior,
// which is that it does nothing.
//
// Upstream declares `defaultOptions: ['always']` as a createRule property rather than as
// `meta.defaultOptions`, and ESLint 10 applies only the latter, so a rule named as a bare severity
// string is handed no mode and reports nothing. Measured across four spellings of the same
// configuration on `var foo; var bar = 1;`:
//
//	"error"                 no findings
//	["error"]               no findings
//	["error", "always"]     reports foo
//	["error", "never"]      reports bar
//
// This test exists because the natural port is the wrong one. Reading `defaultOptions: ['always']` and
// defaulting to always passes all seventy-seven imported cases, since every one of them names its
// mode, and it puts 512 findings on the real tree that the gate being replaced does not report. The
// first version of this port did exactly that, and only the dry run disagreed.
//
// `EnableRule.ts` writes the bare severity spelling, so this IS how the rule is configured here.
func TestInitDeclarationsIsInertWhenNamedWithoutAMode(t *testing.T) {
	t.Parallel()

	const sourceText = "var foo;\nvar bar = 1;\n"

	// Handed nil, which is what the config layer passes for a rule named as a bare severity.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, InitDeclarations,
		initDeclarationsFile, sourceText, nil))

	// Handed an empty array, which is the other spelling of the same configuration.
	rule_testing.ExpectClean(t, runInitDeclarations(t, sourceText, `[]`))

	// And the controls, so the three silences above are verdicts rather than a rule that never ran.
	rule_testing.ExpectFindings(t, runInitDeclarations(t, sourceText, `["always"]`), "initialized")
	rule_testing.ExpectFindings(t, runInitDeclarations(t, sourceText, `["never"]`), "notInitialized")
}
