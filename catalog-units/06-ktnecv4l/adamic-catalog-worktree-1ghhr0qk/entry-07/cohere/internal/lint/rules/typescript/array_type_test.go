package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const arrayTypeFile = "/repository/source/Thing.ts"

// arrayTypeCaseName numbers a row so a failure names which one, since many rows differ only in an
// option or in one character of the type.
func arrayTypeCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// decodeArrayTypeOptions runs a configuration through the real decoder rather than building the
// options struct directly.
//
// The decoder holds two defaults, neither of which is a Go zero value, and a fallback of `readonly`
// onto the RESOLVED value of `default`. A fixture constructing ArrayTypeOptions by hand would leave
// both untested, and those are the two lines with no upstream counterpart. An empty configuration
// is the bare `"error"` case, which the config layer turns into nil options, and it reaches the rule
// the same way here.
func decodeArrayTypeOptions(t *testing.T, configuration string) any {
	t.Helper()

	if configuration == "" {
		// A rule configured as a bare `"error"` is handed nil, not an empty struct. Passing nil
		// here is what puts the rule's own fallback under test rather than the decoder's.
		return nil
	}
	options, err := DecodeArrayTypeOptions([]byte(configuration))
	if err != nil {
		t.Fatalf("could not decode %s: %v", configuration, err)
	}
	return options
}

// TestArrayTypeStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All ninety-four of upstream's passing inputs with the options each carries, extracted from the
// clone's test file by parsing it with the TypeScript compiler rather than by reading it, so no
// escape sequence passed through a shell or a keyboard on the way here. Every one was additionally
// run through the installed 8.67.0 build driven by the ESLint 10.8.1 Linter API, which reported
// nothing on all ninety-four.
//
// The same source appears several times under different options, which is the point: this rule's
// verdict lives above the code as often as in it.
func TestArrayTypeStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
	}{
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array\"}",
			sourceText:    "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array\"}",
			sourceText:    "let a: (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array\"}",
			sourceText:    "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"generic\"}",
			sourceText:    "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"generic\"}",
			sourceText:    "let a: (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"generic\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"generic\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array\"}",
			sourceText:    "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array\"}",
			sourceText:    "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array\"}",
			sourceText:    "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"generic\"}",
			sourceText:    "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"generic\"}",
			sourceText:    "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"generic\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"generic\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: Array<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: Array<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: Array<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: Array<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: Array<bigint> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: readonly bigint[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: readonly (string | bigint)[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: Array<bigint> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: Array<string | bigint> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: readonly bigint[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<string | bigint> = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a = new Array();",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: { foo: Bar[] }[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "function foo(a: Array<Bar>): Array<Bar> {}",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let yy: number[][] = [[4, 5], [6]];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "\nfunction fooFunction(foo: Array<ArrayClass<string>>) {\n  return foo.map(e => e.foo);\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "\nfunction bazFunction(baz: Arr<ArrayClass<String>>) {\n  return baz.map(e => e.baz);\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let fooVar: Array<(c: number) => number>;",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "type fooUnion = Array<string | number | boolean>;",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "type fooIntersection = Array<string & number>;",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "\nnamespace fooName {\n  type BarType = { bar: string };\n  type BazType<T> = Arr<T>;\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "\ninterface FooInterface {\n  '.bar': { baz: string[] };\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let yy: number[][] = [[4, 5], [6]];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let ya = [[1, '2']] as [number, string][];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "\nfunction barFunction(bar: ArrayClass<String>[]) {\n  return bar.map(e => e.bar);\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "\nfunction bazFunction(baz: Arr<ArrayClass<String>>) {\n  return baz.map(e => e.baz);\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let barVar: ((c: number) => number)[];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type barUnion = (string | number | boolean)[];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type barIntersection = (string & number)[];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "\ninterface FooInterface {\n  '.bar': { baz: string[] };\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type Unwrap<T> = T extends (infer E)[] ? E : T;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let xx: Array<Array<number>> = [[1, 2], [3]];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "type Arr<T> = Array<T>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "\nfunction fooFunction(foo: Array<ArrayClass<string>>) {\n  return foo.map(e => e.foo);\n}\n      ",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "\nfunction bazFunction(baz: Arr<ArrayClass<String>>) {\n  return baz.map(e => e.baz);\n}\n      ",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let fooVar: Array<(c: number) => number>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "type fooUnion = Array<string | number | boolean>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "type fooIntersection = Array<string & number>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "type Unwrap<T> = T extends Array<infer E> ? E : T;",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"generic\"}",
			sourceText:    "let a: ReadonlyArray<number[]> = [[]];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: readonly Array<number>[] = [[]];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: Readonly = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "const x: Readonly<string> = 'a';",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type Generic<Array extends unknown[]> = { array: Array };",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type Generic<ReadonlyArray> = { array: ReadonlyArray };",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "\ndeclare module '2' {\n  type Array<Y> = Y;\n  const y: Array<2>;\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let x: Array;",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let x: Array;",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let z: Array = [3, '4'];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let z: Array = [3, '4'];",
		},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, ArrayType,
				arrayTypeFile, testCase.sourceText,
				decodeArrayTypeOptions(t, testCase.configuration)))
		})
	}
}

// TestArrayTypeFiresOnUpstreamFailCases is the imported reporting corpus, verbatim.
//
// Ninety-nine inputs carrying one hundred diagnostics and ninety-nine before-and-after pairs. The
// pairs are the highest-value thing in this corpus: they assert what the repair writes rather than
// whether a finding appears, and a fixer that repairs the right span with the wrong text passes
// every message-id check.
//
// Four things are asserted per row, and each catches a different defect. The ids say which arm ran,
// of six that differ only in wording. The spans say where the finding points, which for a readonly
// array is the type operator rather than the array. The rendered text says what the reader is told,
// including the collapse of a non-simple element to the letter `T`, which no id can see. And the
// applied source says what the edit engine will write into the file unattended.
func TestArrayTypeFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantIds       []string
		wantSpans     []string
		wantFixed     string
	}{
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: Array<number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<number>"},
			wantFixed:     "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: Array<string | number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<string | number>"},
			wantFixed:     "let a: (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<number>"},
			wantFixed:     "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<string | number>"},
			wantFixed:     "let a: readonly (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array\"}",
			sourceText:    "let a: Array<number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<number>"},
			wantFixed:     "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array\"}",
			sourceText:    "let a: Array<string | number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<string | number>"},
			wantFixed:     "let a: (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<number>"},
			wantFixed:     "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<string | number>"},
			wantFixed:     "let a: readonly (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: Array<number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<number>"},
			wantFixed:     "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: Array<string | number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<string | number>"},
			wantFixed:     "let a: (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"ReadonlyArray<number>"},
			wantFixed:     "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"readonly (string | number)[]"},
			wantFixed:     "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"generic\"}",
			sourceText:    "let a: Array<number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<number>"},
			wantFixed:     "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"generic\"}",
			sourceText:    "let a: Array<string | number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<string | number>"},
			wantFixed:     "let a: (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"generic\"}",
			sourceText:    "let a: readonly number[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly number[]"},
			wantFixed:     "let a: ReadonlyArray<number> = [];",
		},
		{
			configuration: "{\"default\":\"array\",\"readonly\":\"generic\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly (string | number)[]"},
			wantFixed:     "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let a: Array<number> = [];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"Array<number>"},
			wantFixed:     "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let a: (string | number)[] = [];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"(string | number)[]"},
			wantFixed:     "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"ReadonlyArray<number>"},
			wantFixed:     "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"readonly (string | number)[]"},
			wantFixed:     "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array\"}",
			sourceText:    "let a: Array<number> = [];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"Array<number>"},
			wantFixed:     "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array\"}",
			sourceText:    "let a: (string | number)[] = [];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"(string | number)[]"},
			wantFixed:     "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<number>"},
			wantFixed:     "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<string | number>"},
			wantFixed:     "let a: readonly (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: Array<number> = [];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"Array<number>"},
			wantFixed:     "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: (string | number)[] = [];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"(string | number)[]"},
			wantFixed:     "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"ReadonlyArray<number>"},
			wantFixed:     "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"readonly (string | number)[]"},
			wantFixed:     "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"generic\"}",
			sourceText:    "let a: Array<number> = [];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"Array<number>"},
			wantFixed:     "let a: number[] = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"generic\"}",
			sourceText:    "let a: (string | number)[] = [];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"(string | number)[]"},
			wantFixed:     "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"generic\"}",
			sourceText:    "let a: readonly number[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly number[]"},
			wantFixed:     "let a: ReadonlyArray<number> = [];",
		},
		{
			configuration: "{\"default\":\"array-simple\",\"readonly\":\"generic\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly (string | number)[]"},
			wantFixed:     "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: number[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"number[]"},
			wantFixed:     "let a: Array<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: (string | number)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(string | number)[]"},
			wantFixed:     "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: readonly number[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly number[]"},
			wantFixed:     "let a: ReadonlyArray<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly (string | number)[]"},
			wantFixed:     "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: number[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"number[]"},
			wantFixed:     "let a: Array<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: (string | number)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(string | number)[]"},
			wantFixed:     "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<number>"},
			wantFixed:     "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array\"}",
			sourceText:    "let a: ReadonlyArray<string | number> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<string | number>"},
			wantFixed:     "let a: readonly (string | number)[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: number[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"number[]"},
			wantFixed:     "let a: Array<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: (string | number)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(string | number)[]"},
			wantFixed:     "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<number> = [];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"ReadonlyArray<number>"},
			wantFixed:     "let a: readonly number[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"readonly (string | number)[]"},
			wantFixed:     "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: number[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"number[]"},
			wantFixed:     "let a: Array<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: (string | number)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(string | number)[]"},
			wantFixed:     "let a: Array<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: readonly number[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly number[]"},
			wantFixed:     "let a: ReadonlyArray<number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: readonly (string | number)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly (string | number)[]"},
			wantFixed:     "let a: ReadonlyArray<string | number> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: bigint[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"bigint[]"},
			wantFixed:     "let a: Array<bigint> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: (string | bigint)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(string | bigint)[]"},
			wantFixed:     "let a: Array<string | bigint> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"array-simple\"}",
			sourceText:    "let a: ReadonlyArray<bigint> = [];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"ReadonlyArray<bigint>"},
			wantFixed:     "let a: readonly bigint[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: (string | bigint)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(string | bigint)[]"},
			wantFixed:     "let a: Array<string | bigint> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: readonly bigint[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly bigint[]"},
			wantFixed:     "let a: ReadonlyArray<bigint> = [];",
		},
		{
			configuration: "{\"default\":\"generic\",\"readonly\":\"generic\"}",
			sourceText:    "let a: readonly (string | bigint)[] = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly (string | bigint)[]"},
			wantFixed:     "let a: ReadonlyArray<string | bigint> = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let a: { foo: Array<Bar> }[] = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<Bar>"},
			wantFixed:     "let a: { foo: Bar[] }[] = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: Array<{ foo: Bar[] }> = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"Bar[]"},
			wantFixed:     "let a: Array<{ foo: Array<Bar> }> = [];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let a: Array<{ foo: Foo | Bar[] }> = [];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"Bar[]"},
			wantFixed:     "let a: Array<{ foo: Foo | Array<Bar> }> = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "function foo(a: Array<Bar>): Array<Bar> {}",
			wantIds:       []string{"errorStringArray", "errorStringArray"},
			wantSpans:     []string{"Array<Bar>", "Array<Bar>"},
			wantFixed:     "function foo(a: Bar[]): Bar[] {}",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let x: Array<undefined> = [undefined] as undefined[];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"Array<undefined>"},
			wantFixed:     "let x: undefined[] = [undefined] as undefined[];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let y: string[] = <Array<string>>['2'];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"Array<string>"},
			wantFixed:     "let y: string[] = <string[]>['2'];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let ya = [[1, '2']] as [number, string][];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"[number, string][]"},
			wantFixed:     "let ya = [[1, '2']] as Array<[number, string]>;",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "type Arr<T> = Array<T>;",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"Array<T>"},
			wantFixed:     "type Arr<T> = T[];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "\n// Ignore user defined aliases\nlet yyyy: Arr<Array<Arr<string>>[]> = [[[['2']]]];\n      ",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"Array<Arr<string>>[]"},
			wantFixed:     "\n// Ignore user defined aliases\nlet yyyy: Arr<Array<Array<Arr<string>>>> = [[[['2']]]];\n      ",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "\ninterface ArrayClass<T> {\n  foo: Array<T>;\n  bar: T[];\n  baz: Arr<T>;\n  xyz: this[];\n}\n      ",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"Array<T>"},
			wantFixed:     "\ninterface ArrayClass<T> {\n  foo: T[];\n  bar: T[];\n  baz: Arr<T>;\n  xyz: this[];\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "\nfunction barFunction(bar: ArrayClass<String>[]) {\n  return bar.map(e => e.bar);\n}\n      ",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"ArrayClass<String>[]"},
			wantFixed:     "\nfunction barFunction(bar: Array<ArrayClass<String>>) {\n  return bar.map(e => e.bar);\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let barVar: ((c: number) => number)[];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"((c: number) => number)[]"},
			wantFixed:     "let barVar: Array<(c: number) => number>;",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "type barUnion = (string | number | boolean)[];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"(string | number | boolean)[]"},
			wantFixed:     "type barUnion = Array<string | number | boolean>;",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "type barIntersection = (string & number)[];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"(string & number)[]"},
			wantFixed:     "type barIntersection = Array<string & number>;",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let v: Array<fooName.BarType> = [{ bar: 'bar' }];",
			wantIds:       []string{"errorStringArraySimple"},
			wantSpans:     []string{"Array<fooName.BarType>"},
			wantFixed:     "let v: fooName.BarType[] = [{ bar: 'bar' }];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let w: fooName.BazType<string>[] = [['baz']];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"fooName.BazType<string>[]"},
			wantFixed:     "let w: Array<fooName.BazType<string>> = [['baz']];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let x: Array<undefined> = [undefined] as undefined[];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<undefined>"},
			wantFixed:     "let x: undefined[] = [undefined] as undefined[];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let y: string[] = <Array<string>>['2'];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<string>"},
			wantFixed:     "let y: string[] = <string[]>['2'];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type Arr<T> = Array<T>;",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<T>"},
			wantFixed:     "type Arr<T> = T[];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "\n// Ignore user defined aliases\nlet yyyy: Arr<Array<Arr<string>>[]> = [[[['2']]]];\n      ",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<Arr<string>>"},
			wantFixed:     "\n// Ignore user defined aliases\nlet yyyy: Arr<Arr<string>[][]> = [[[['2']]]];\n      ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "\ninterface ArrayClass<T> {\n  foo: Array<T>;\n  bar: T[];\n  baz: Arr<T>;\n}\n      ",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<T>"},
			wantFixed:     "\ninterface ArrayClass<T> {\n  foo: T[];\n  bar: T[];\n  baz: Arr<T>;\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "\nfunction fooFunction(foo: Array<ArrayClass<string>>) {\n  return foo.map(e => e.foo);\n}\n      ",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<ArrayClass<string>>"},
			wantFixed:     "\nfunction fooFunction(foo: ArrayClass<string>[]) {\n  return foo.map(e => e.foo);\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let fooVar: Array<(c: number) => number>;",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<(c: number) => number>"},
			wantFixed:     "let fooVar: ((c: number) => number)[];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type fooUnion = Array<string | number | boolean>;",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<string | number | boolean>"},
			wantFixed:     "type fooUnion = (string | number | boolean)[];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type fooIntersection = Array<string & number>;",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<string & number>"},
			wantFixed:     "type fooIntersection = (string & number)[];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let x: Array<number> = [1] as number[];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"number[]"},
			wantFixed:     "let x: Array<number> = [1] as Array<number>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let y: string[] = <Array<string>>['2'];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"string[]"},
			wantFixed:     "let y: Array<string> = <Array<string>>['2'];",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let ya = [[1, '2']] as [number, string][];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"[number, string][]"},
			wantFixed:     "let ya = [[1, '2']] as Array<[number, string]>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "\n// Ignore user defined aliases\nlet yyyy: Arr<Array<Arr<string>>[]> = [[[['2']]]];\n      ",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"Array<Arr<string>>[]"},
			wantFixed:     "\n// Ignore user defined aliases\nlet yyyy: Arr<Array<Array<Arr<string>>>> = [[[['2']]]];\n      ",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "\ninterface ArrayClass<T> {\n  foo: Array<T>;\n  bar: T[];\n  baz: Arr<T>;\n}\n      ",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"T[]"},
			wantFixed:     "\ninterface ArrayClass<T> {\n  foo: Array<T>;\n  bar: Array<T>;\n  baz: Arr<T>;\n}\n      ",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "\nfunction barFunction(bar: ArrayClass<String>[]) {\n  return bar.map(e => e.bar);\n}\n      ",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"ArrayClass<String>[]"},
			wantFixed:     "\nfunction barFunction(bar: Array<ArrayClass<String>>) {\n  return bar.map(e => e.bar);\n}\n      ",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let barVar: ((c: number) => number)[];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"((c: number) => number)[]"},
			wantFixed:     "let barVar: Array<(c: number) => number>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "type barUnion = (string | number | boolean)[];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(string | number | boolean)[]"},
			wantFixed:     "type barUnion = Array<string | number | boolean>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "type barIntersection = (string & number)[];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(string & number)[]"},
			wantFixed:     "type barIntersection = Array<string & number>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "\ninterface FooInterface {\n  '.bar': { baz: string[] };\n}\n      ",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"string[]"},
			wantFixed:     "\ninterface FooInterface {\n  '.bar': { baz: Array<string> };\n}\n      ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type Unwrap<T> = T extends Array<infer E> ? E : T;",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<infer E>"},
			wantFixed:     "type Unwrap<T> = T extends (infer E)[] ? E : T;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "type Unwrap<T> = T extends (infer E)[] ? E : T;",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(infer E)[]"},
			wantFixed:     "type Unwrap<T> = T extends Array<infer E> ? E : T;",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type Foo = ReadonlyArray<object>[];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<object>"},
			wantFixed:     "type Foo = (readonly object[])[];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "const foo: Array<new (...args: any[]) => void> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<new (...args: any[]) => void>"},
			wantFixed:     "const foo: (new (...args: any[]) => void)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "const foo: ReadonlyArray<new (...args: any[]) => void> = [];",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"ReadonlyArray<new (...args: any[]) => void>"},
			wantFixed:     "const foo: readonly (new (...args: any[]) => void)[] = [];",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "const x: Readonly<string[]> = ['a', 'b'];",
			wantIds:       []string{"errorStringArrayReadonly"},
			wantSpans:     []string{"Readonly<string[]>"},
			wantFixed:     "const x: readonly string[] = ['a', 'b'];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "declare function foo<E extends Readonly<string[]>>(extra: E): E;",
			wantIds:       []string{"errorStringArraySimpleReadonly"},
			wantSpans:     []string{"Readonly<string[]>"},
			wantFixed:     "declare function foo<E extends readonly string[]>(extra: E): E;",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "type Conditional<T> = Array<T extends string ? string : number>;",
			wantIds:       []string{"errorStringArray"},
			wantSpans:     []string{"Array<T extends string ? string : number>"},
			wantFixed:     "type Conditional<T> = (T extends string ? string : number)[];",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "type Conditional<T> = (T extends string ? string : number)[];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantSpans:     []string{"(T extends string ? string : number)[]"},
			wantFixed:     "type Conditional<T> = Array<T extends string ? string : number>;",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "type Conditional<T> = (T extends string ? string : number)[];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"(T extends string ? string : number)[]"},
			wantFixed:     "type Conditional<T> = Array<T extends string ? string : number>;",
		},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
				testCase.sourceText, decodeArrayTypeOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			for index, wantSpan := range testCase.wantSpans {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
			}

			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestArrayTypeRendersTheMessageUpstreamRenders pins all six texts and their three interpolations.
//
// Upstream interpolates `className`, `type` and `readonlyPrefix` into six message strings, and the
// `type` slot is the one that hides a decision: a simple element is rendered as its own source text
// and everything else COLLAPSES to the single letter `T`. A message-id assertion cannot see that,
// nor can a span, nor can the applied fix, since the repair writes the real type either way.
//
// Every expectation here is the triple the installed build produced for that exact input, recovered
// from its message text rather than predicted.
func TestArrayTypeRendersTheMessageUpstreamRenders(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantId        string
		wantClassName string
		wantType      string
		wantReadonly  string
	}{
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let x: Array<string>;",
			wantId:        "errorStringArray",
			wantClassName: "Array",
			wantType:      "string",
			wantReadonly:  "",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let x: Array<string|number>;",
			wantId:        "errorStringArray",
			wantClassName: "Array",
			wantType:      "T",
			wantReadonly:  "",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let x: ReadonlyArray<string>;",
			wantId:        "errorStringArray",
			wantClassName: "ReadonlyArray",
			wantType:      "string",
			wantReadonly:  "readonly ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let x: ReadonlyArray<string|number>;",
			wantId:        "errorStringArray",
			wantClassName: "ReadonlyArray",
			wantType:      "T",
			wantReadonly:  "readonly ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let x: Readonly<string[]>;",
			wantId:        "errorStringArrayReadonly",
			wantClassName: "Readonly",
			wantType:      "string[]",
			wantReadonly:  "readonly ",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let x: Readonly<(string|number)[]>;",
			wantId:        "errorStringArrayReadonly",
			wantClassName: "Readonly",
			wantType:      "(string|number)[]",
			wantReadonly:  "readonly ",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let x: Array<string>;",
			wantId:        "errorStringArraySimple",
			wantClassName: "Array",
			wantType:      "string",
			wantReadonly:  "",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let x: ReadonlyArray<string>;",
			wantId:        "errorStringArraySimple",
			wantClassName: "ReadonlyArray",
			wantType:      "string",
			wantReadonly:  "readonly ",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let x: Readonly<string[]>;",
			wantId:        "errorStringArraySimpleReadonly",
			wantClassName: "Readonly",
			wantType:      "string[]",
			wantReadonly:  "readonly ",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let x: string[];",
			wantId:        "errorStringGeneric",
			wantClassName: "Array",
			wantType:      "string",
			wantReadonly:  "",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let x: readonly string[];",
			wantId:        "errorStringGeneric",
			wantClassName: "ReadonlyArray",
			wantType:      "string",
			wantReadonly:  "readonly ",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let x: (string|number)[];",
			wantId:        "errorStringGenericSimple",
			wantClassName: "Array",
			wantType:      "T",
			wantReadonly:  "",
		},
		{
			configuration: "{\"default\":\"array-simple\"}",
			sourceText:    "let x: readonly (string|number)[];",
			wantId:        "errorStringGenericSimple",
			wantClassName: "ReadonlyArray",
			wantType:      "T",
			wantReadonly:  "readonly ",
		},
		{
			configuration: "{\"default\":\"generic\"}",
			sourceText:    "let x: Foo[];",
			wantId:        "errorStringGeneric",
			wantClassName: "Array",
			wantType:      "Foo",
			wantReadonly:  "",
		},
		{
			configuration: "{\"default\":\"array\"}",
			sourceText:    "let x: Array<Foo.Bar>;",
			wantId:        "errorStringArray",
			wantClassName: "Array",
			wantType:      "Foo.Bar",
			wantReadonly:  "",
		},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
				testCase.sourceText, decodeArrayTypeOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.wantId)

			want := arrayTypeMessageForTest(testCase.wantId, testCase.wantClassName,
				testCase.wantType, testCase.wantReadonly)
			if result.Diagnostics[0].Message.Description != want.Description {
				t.Errorf("the description is %q, want %q",
					result.Diagnostics[0].Message.Description, want.Description)
			}
		})
	}
}

// TestArrayTypeDeclinesAHeritageClause pins the false-positive class the corpus cannot see.
//
// A heritage clause is not a type position upstream's visitor can reach: TSESTree gives an
// interface's `extends` and a class's `implements` their own node types, so `TSTypeReference` never
// fires on them. typescript-go reuses `KindTypeReference` under a heritage clause, so the listener
// does fire and the rule would propose a rewrite that does not even parse.
//
// This shipped in the first draft and no fixture could see it. It was found by running the rule
// against the real tree and diffing: cohere reported 713 findings where ESLint reported 711, and
// both extra were `extends Array<...>` in one file. All five rows below are the installed 8.67.0
// build's verdict, and the last is the control that separates "declined the heritage clause" from
// "never reached the shape".
func TestArrayTypeDeclinesAHeritageClause(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantCount  int
		reason     string
	}{
		{"interface I extends Array<string> {}", 0, "an interface extends clause is not a type position"},
		{"interface I extends ReadonlyArray<string> {}", 0, "nor is a readonly one"},
		{"class C implements Array<string> {}", 0, "nor is a class implements clause"},
		{"class C extends Array<string> {}", 0, "a class extends clause is a different node kind and was already declined"},
		{"let x: Array<string>;", 1, "the control: an ordinary type position does report"},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
				testCase.sourceText, decodeArrayTypeOptions(t, "{\"default\": \"array\"}"))
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "errorStringArray"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestArrayTypeDeclinesAGenericWithTheWrongArgumentCount pins an exact-count guard.
//
// A rewrite from `Array<A, B>` to a suffix form has no meaning, so upstream declines anything that
// does not take exactly one type argument. The guard reads as a defensive count and is a real
// verdict: `Array<string, number>` is illegal TypeScript that the parser still produces a node for,
// and a rule reporting it would propose a repair that loses an argument.
//
// A mutant relaxing the test from "exactly one" to "at least one" survived the imported corpus,
// which writes no two-argument case at all. All four rows were run through the installed 8.67.0
// build.
func TestArrayTypeDeclinesAGenericWithTheWrongArgumentCount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantCount  int
		reason     string
	}{
		{"let x: Array<string, number>;", 0, "two arguments have no suffix spelling"},
		{"let x: ReadonlyArray<string, number>;", 0, "and neither does a readonly one"},
		{"let x: Readonly<string[], number>;", 0, "nor the Readonly form"},
		{"let x: Array<string>;", 1, "the control: exactly one argument does report"},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
				testCase.sourceText, decodeArrayTypeOptions(t, "{\"default\": \"array\"}"))
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "errorStringArray"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestArrayTypeWrapsAReadonlyRewriteOnlyWhenItHasTo pins the outer-parenthesis decision.
//
// A `readonly` rewrite sitting directly inside another suffix array needs its own parentheses,
// because `readonly string[][]` would read as an array of readonly arrays only by accident. When the
// element is ALREADY parenthesized the rewrite must not add a second pair, and the reason it does
// not is subtle: the node's own range starts inside the existing parentheses, so the fix that
// replaces it never sees them.
//
// A mutant forcing the wrap on survived the whole imported corpus, which writes no parenthesized
// element in this position. These rows separate the two, and each asserts the applied source rather
// than the fix text, because a fix writing the right string over the wrong span passes a text
// comparison.
//
// All four outputs are what the installed 8.67.0 build's `cohereAndFix` wrote.
func TestArrayTypeWrapsAReadonlyRewriteOnlyWhenItHasTo(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantId     string
		wantFixed  string
		reason     string
	}{
		{
			sourceText: "let x: ReadonlyArray<string>[];",
			wantId:     "errorStringArray",
			wantFixed:  "let x: (readonly string[])[];",
			reason:     "an unparenthesized element gets parentheses from the rewrite",
		},
		{
			sourceText: "let x: (ReadonlyArray<string>)[];",
			wantId:     "errorStringArray",
			wantFixed:  "let x: (readonly string[])[];",
			reason:     "an already-parenthesized one does not, and the two agree on the result",
		},
		{
			sourceText: "let x: readonly ReadonlyArray<string>[];",
			wantId:     "errorStringArray",
			wantFixed:  "let x: readonly (readonly string[])[];",
			reason:     "the outer readonly changes nothing about the inner decision",
		},
		{
			// A different arm, and the id says so: `Readonly<T[]>` already ends in brackets, so its
			// repair does not append another pair and upstream gives it its own message.
			sourceText: "let x: (Readonly<string[]>)[];",
			wantId:     "errorStringArrayReadonly",
			wantFixed:  "let x: (readonly string[])[];",
			reason:     "and the Readonly spelling reaches the same place by a different arm",
		},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
				testCase.sourceText, decodeArrayTypeOptions(t, "{\"default\": \"array\"}"))
			rule_testing.ExpectFindings(t, result, testCase.wantId)
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestArrayTypeJudgesSimplicityTheWayUpstreamDoes pins the two arms the corpus cannot reach.
//
// `isSimpleType` has one arm that treats `Array` unlike every other name, and one behavior that
// exists only because our parser keeps parentheses. Mutants removing each survived the whole
// imported corpus, which writes neither shape under `array-simple`.
//
// The `Array` arm: a reference is simple only when it takes NO type arguments, EXCEPT that `Array`
// is simple when its single argument is. So `Array<Array<string>>` is simple all the way down and
// reports twice under `array-simple`, once per level, while `Foo<string>` is not simple at all.
//
// The parenthesis behavior: TSESTree has no parenthesized-type node, so `(string)[]`'s element type
// IS the string keyword and is simple. Ours produces a real node that would answer false without
// the skip, turning a clean case into a reported one.
//
// Every field is what the installed 8.67.0 build produced for that exact input.
func TestArrayTypeJudgesSimplicityTheWayUpstreamDoes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantIds       []string
		wantFixed     string
		reason        string
	}{
		{
			configuration: "{\"default\": \"array-simple\"}",
			sourceText:    "let x: Array<Array<string>>;",
			wantIds:       []string{"errorStringArraySimple", "errorStringArraySimple"},
			wantFixed:     "let x: string[][];",
			reason:        "Array recurses into its argument, so both levels are simple and both report",
		},
		{
			configuration: "{\"default\": \"array-simple\"}",
			sourceText:    "let x: Array<Array>;",
			wantIds:       []string{"errorStringArraySimple"},
			wantFixed:     "let x: Array[];",
			reason:        "and a bare Array is simple with no arguments to recurse into",
		},
		{
			configuration: "{\"default\": \"array-simple\"}",
			sourceText:    "let x: (string)[];",
			wantIds:       nil,
			wantFixed:     "",
			reason:        "a parenthesized simple element is still simple, so this is clean",
		},
		{
			configuration: "{\"default\": \"array-simple\"}",
			sourceText:    "let x: (string|number)[];",
			wantIds:       []string{"errorStringGenericSimple"},
			wantFixed:     "let x: Array<string|number>;",
			reason:        "the control: a parenthesized union is not simple and does report",
		},
		{
			configuration: "{\"default\": \"generic\"}",
			sourceText:    "let x: (string)[];",
			wantIds:       []string{"errorStringGeneric"},
			wantFixed:     "let x: Array<string>;",
			reason:        "and the repair drops the parentheses, anchoring inside them rather than around",
		},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
				testCase.sourceText, decodeArrayTypeOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(testCase.wantIds) > 0 {
				rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
			}
		})
	}
}

// TestArrayTypeReadsOnlyTheReadonlyOperator separates `readonly` from the other type operators.
//
// `readonly T[]` is a type operator wrapping an array, and so are `keyof T[]` and `unique symbol[]`.
// Only the first makes the array readonly, so only the first is judged on the `readonly` axis, is
// reported on the operator rather than on the array, and repairs to `ReadonlyArray`. A mutant
// dropping the token test survived the whole imported corpus, which writes no `keyof` over an array
// at all; these rows are what it takes to see it.
//
// Every field is what the installed 8.67.0 build produced for that exact input.
func TestArrayTypeReadsOnlyTheReadonlyOperator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantIds       []string
		wantSpans     []string
		wantFixed     string
		reason        string
	}{
		{
			configuration: "{\"default\": \"generic\"}",
			sourceText:    "let x: keyof string[];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"string[]"},
			wantFixed:     "let x: keyof Array<string>;",
			reason:        "keyof is not readonly, so the span is the array and the repair keeps keyof",
		},
		{
			configuration: "{\"default\": \"generic\"}",
			sourceText:    "let x: unique symbol[];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"symbol[]"},
			wantFixed:     "let x: unique Array<symbol>;",
			reason:        "nor is unique",
		},
		{
			configuration: "{\"default\": \"generic\"}",
			sourceText:    "let x: readonly string[];",
			wantIds:       []string{"errorStringGeneric"},
			wantSpans:     []string{"readonly string[]"},
			wantFixed:     "let x: ReadonlyArray<string>;",
			reason:        "the control: readonly IS, so the span is the operator and the repair replaces it",
		},
		{
			configuration: "{\"default\": \"array\", \"readonly\": \"generic\"}",
			sourceText:    "let x: keyof string[];",
			wantIds:       nil,
			wantSpans:     nil,
			wantFixed:     "",
			reason:        "and keyof is judged on the default axis, so a readonly-only setting misses it",
		},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
				testCase.sourceText, decodeArrayTypeOptions(t, testCase.configuration))
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			for index, wantSpan := range testCase.wantSpans {
				reported := testCase.sourceText[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
			}

			if len(testCase.wantIds) > 0 {
				rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
			}
		})
	}
}

// TestArrayTypeDeclinesAShadowedName is the scope walk, which has no shelf to lean on.
//
// Upstream asks eslint-scope; this tree has no scope layer, so the walk is written in the rule and
// is the largest thing in the port that is not a transcription. Every row was run through the
// installed 8.67.0 build and carries the verdict that build produced.
//
// The rows come in pairs on purpose. A silence that is not paired with a control is not a
// measurement: it reads identically whether the shadow was seen or the rule simply never reached
// the shape.
func TestArrayTypeDeclinesAShadowedName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantCount  int
		reason     string
	}{
		{
			sourceText: "type Array<Y> = Y; const y: Array<2>;",
			wantCount:  0,
			reason:     "a file-scope type alias shadows",
		},
		{
			sourceText: "const y: Array<2>;",
			wantCount:  1,
			reason:     "the control for the row above",
		},
		{
			sourceText: "interface Array<Y> { z: Y } const y: Array<2>;",
			wantCount:  0,
			reason:     "an interface shadows",
		},
		{
			sourceText: "class Array<Y> { z: Y } const y: Array<2>;",
			wantCount:  0,
			reason:     "a class shadows",
		},
		{
			sourceText: "enum Array { A } const y: Array<2>;",
			wantCount:  0,
			reason:     "an enum shadows",
		},
		{
			sourceText: "declare const Array: unknown; const y: Array<2>;",
			wantCount:  0,
			reason:     "a value declaration shadows a type position",
		},
		{
			sourceText: "declare let Array: unknown; const y: Array<2>;",
			wantCount:  0,
			reason:     "and so does a let",
		},
		{
			sourceText: "declare function Array(): void; const y: Array<2>;",
			wantCount:  0,
			reason:     "and a function declaration",
		},
		{
			sourceText: "import { Array } from \"m\"; const y: Array<2>;",
			wantCount:  0,
			reason:     "and a named import",
		},
		{
			sourceText: "import Array from \"m\"; const y: Array<2>;",
			wantCount:  0,
			reason:     "and a default import",
		},
		{
			sourceText: "import * as Array from \"m\"; const y: Array<2>;",
			wantCount:  0,
			reason:     "and a namespace import",
		},
		{
			sourceText: "namespace Array { export const z = 1; } const y: Array<2>;",
			wantCount:  0,
			reason:     "and a namespace",
		},
		{
			sourceText: "type G<Array> = { a: Array<string> };",
			wantCount:  0,
			reason:     "a type parameter shadows",
		},
		{
			sourceText: "type G<T> = { a: Array<string> };",
			wantCount:  1,
			reason:     "the control for the row above",
		},
		{
			sourceText: "declare module \"2\" { type Array<Y> = Y; const y: Array<2>; }",
			wantCount:  0,
			reason:     "a module-scope alias shadows inside it",
		},
		{
			sourceText: "declare module \"2\" { const y: Array<2>; }",
			wantCount:  1,
			reason:     "the control for the row above",
		},
		{
			sourceText: "function f() { type Array<Y> = Y; const y: Array<2> = 1 as any; }",
			wantCount:  0,
			reason:     "a function-scope alias shadows inside it",
		},
		{
			sourceText: "function f() { const y: Array<2> = [] as any; }",
			wantCount:  1,
			reason:     "the control for the row above",
		},
		{
			sourceText: "{ type Array<Y> = Y; } const y: Array<2>;",
			wantCount:  1,
			reason:     "a SIBLING block does not shadow, so this reports",
		},
		{
			sourceText: "const y: Array<2>; type Array<Y> = Y;",
			wantCount:  0,
			reason:     "a later declaration in the same scope still binds",
		},
		{
			sourceText: "class C { Array: unknown; m(): Array<2> { return null as any; } }",
			wantCount:  1,
			reason:     "a class member is not a variable",
		},
		{
			sourceText: "interface I { Array: unknown; m(): Array<2>; }",
			wantCount:  1,
			reason:     "nor is an interface member",
		},
		{
			sourceText: "type T = { Array: unknown; m(): Array<2> };",
			wantCount:  1,
			reason:     "nor a type literal member",
		},
		{
			sourceText: "try {} catch (Array) { const y: Array<2> = 1 as any; }",
			wantCount:  0,
			reason:     "a catch parameter shadows",
		},
		{
			sourceText: "for (let Array = 0; ; ) { const y: Array<2> = 1 as any; }",
			wantCount:  0,
			reason:     "a for-statement binding shadows",
		},
		{
			sourceText: "function f(Array: unknown) { const y: Array<2> = 1 as any; }",
			wantCount:  0,
			reason:     "a parameter shadows",
		},
		{
			sourceText: "declare const x: any; function f() { const { Array } = x; const y: Array<2> = 1 as any; }",
			wantCount:  0,
			reason:     "an object destructuring binding shadows",
		},
		{
			sourceText: "declare const x: any; function f() { const [Array] = x; const y: Array<2> = 1 as any; }",
			wantCount:  0,
			reason:     "an array destructuring binding shadows",
		},
		{
			sourceText: "declare const x: any; function f() { const y: Array<2> = 1 as any; }",
			wantCount:  1,
			reason:     "the control for the two rows above",
		},
		{
			sourceText: "type ReadonlyArray<Y> = Y; const y: ReadonlyArray<2>;",
			wantCount:  0,
			reason:     "the walk keys on the name, so ReadonlyArray shadows too",
		},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
				testCase.sourceText, decodeArrayTypeOptions(t, "{\"default\": \"array\"}"))
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "errorStringArray"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestArrayTypeDecoderResolvesBothAxes puts the two lines with no upstream counterpart under test.
//
// `default` defaults to `array` rather than to the empty string a Go zero value would give, and
// `readonly` falls back to the RESOLVED `default` rather than to a constant. Both are invisible to
// any fixture that builds the options struct directly, which is why every other test in this file
// routes through the decoder and why this one asserts the decoder's output.
func TestArrayTypeDecoderResolvesBothAxes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		wantDefault   ArrayTypeSetting
		wantReadonly  ArrayTypeSetting
	}{
		{"{}", ArrayTypeArray, ArrayTypeArray},
		{"{\"default\": \"generic\"}", ArrayTypeGeneric, ArrayTypeGeneric},
		{"{\"default\": \"array-simple\"}", ArrayTypeArraySimple, ArrayTypeArraySimple},
		{"{\"default\": \"generic\", \"readonly\": \"array\"}", ArrayTypeGeneric, ArrayTypeArray},
		{"{\"readonly\": \"generic\"}", ArrayTypeArray, ArrayTypeGeneric},
	}
	for index, testCase := range cases {
		t.Run(arrayTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeArrayTypeOptions([]byte(testCase.configuration))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.configuration, err)
			}
			options, ok := decoded.(ArrayTypeOptions)
			if !ok {
				t.Fatalf("the decoder returned %T rather than ArrayTypeOptions", decoded)
			}
			if options.Default != testCase.wantDefault {
				t.Errorf("default resolved to %q, want %q", options.Default, testCase.wantDefault)
			}
			if options.Readonly != testCase.wantReadonly {
				t.Errorf("readonly resolved to %q, want %q", options.Readonly, testCase.wantReadonly)
			}
		})
	}
}

// TestArrayTypeNilOptionsFallsBackToUpstreamDefaults bypasses the decoder entirely.
//
// A rule configured as a bare `"error"` is handed nil, and a bare type assertion on nil yields the
// zero value, which for this rule is two empty settings matching no arm. That shape registers on
// every file and reports nothing while every decoder-routed fixture stays green, which is the exact
// failure this project has shipped before. So the nil path gets its own row rather than being
// covered by implication.
func TestArrayTypeNilOptionsFallsBackToUpstreamDefaults(t *testing.T) {
	t.Parallel()

	// `array` is the default, so the generic spelling reports and the suffix one does not.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
		"let x: Array<string>;", nil), "errorStringArray")
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile,
		"let x: string[];", nil))
}

// TestArrayTypeSurvivesMalformedTypes is a crash fixture.
//
// `ast.SkipTypeParentheses` dereferences its argument on the first loop test, so a nil element type
// reaches it as a panic rather than as a nil result, and error recovery is where a nil element type
// comes from. The walk recovers per FILE rather than per rule, so one such panic would cost every
// rule that file rather than one finding.
//
// There is no finding to assert; the assertion is that the run completes. Upstream cannot reach
// most of these at all, since its parser rejects what ours recovers from, so asserting a verdict
// would be inventing one.
func TestArrayTypeSurvivesMalformedTypes(t *testing.T) {
	t.Parallel()

	for name, sourceText := range map[string]string{
		"emptyBrackets":   "let x: [];",
		"unclosedGeneric": "let x: Array<;",
		"emptyGeneric":    "let x: Array<>;",
		"strayBracket":    "let x: [;",
		"readonlyNothing": "let x: readonly;",
		"readonlyGeneric": "let x: readonly Array<>;",
		"nestedUnclosed":  "let x: Array<Array<;",
		"parenNothing":    "let x: ()[];",
		"onlyReadonly":    "type T = readonly;",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.RunWithOptions(t, ArrayType, arrayTypeFile, sourceText,
				decodeArrayTypeOptions(t, "{\"default\": \"generic\"}"))
		})
	}
}

// arrayTypeMessageForTest rebuilds one message from its id and its three interpolations.
//
// The switch exists so a fixture can name the arm it expects by upstream's own id rather than by
// calling one of six Go constructors, which keeps the fixture table readable as the upstream table
// it came from. It is deliberately NOT a function the rule uses: a test that reached into the rule
// for its expectation would move both sides together under mutation and assert nothing.
func arrayTypeMessageForTest(id string, className string, elementType string, readonlyPrefix string) rule.Message {
	switch id {
	case "errorStringArray":
		return messageArrayTypeArray(className, elementType, readonlyPrefix)
	case "errorStringArrayReadonly":
		return messageArrayTypeArrayReadonly(className, elementType, readonlyPrefix)
	case "errorStringArraySimple":
		return messageArrayTypeArraySimple(className, elementType, readonlyPrefix)
	case "errorStringArraySimpleReadonly":
		return messageArrayTypeArraySimpleReadonly(className, elementType, readonlyPrefix)
	case "errorStringGeneric":
		return messageArrayTypeGeneric(className, elementType, readonlyPrefix)
	case "errorStringGenericSimple":
		return messageArrayTypeGenericSimple(className, elementType, readonlyPrefix)
	}
	return rule.Message{Id: "unknown", Description: "no such message id: " + id}
}

// The decoder reads empty input and `null` as upstream's defaults and each of upstream's three
// spellings as itself, under either key, and refuses a spelling outside them, naming the key and the
// value (#rfbha44). Each refused value used to fall back to the default, so a config naming one loaded
// clean and ran a setting nobody wrote; upstream's schema is an enum and refuses every one at load.
func TestArrayTypeDecoderRefusesWhatUpstreamRefuses(t *testing.T) {
	t.Parallel()

	for _, accepted := range []struct {
		wire               string
		default_, readonly ArrayTypeSetting
	}{
		{"", ArrayTypeArray, ArrayTypeArray},
		{"null", ArrayTypeArray, ArrayTypeArray},
		{`{}`, ArrayTypeArray, ArrayTypeArray},
		{`{"default": "generic"}`, ArrayTypeGeneric, ArrayTypeGeneric},
		{`{"default": "array-simple", "readonly": "array"}`, ArrayTypeArraySimple, ArrayTypeArray},
	} {
		decoded, err := DecodeArrayTypeOptions([]byte(accepted.wire))
		if err != nil {
			t.Errorf("%s was refused: %v", accepted.wire, err)
			continue
		}
		options := decoded.(ArrayTypeOptions)
		if options.Default != accepted.default_ || options.Readonly != accepted.readonly {
			t.Errorf("%s decoded to %+v, want %s and %s", accepted.wire, options, accepted.default_, accepted.readonly)
		}
	}
	for _, refused := range []struct{ wire, named string }{
		{`{"default": "generics"}`, "generics"},
		{`{"default": "array", "readonly": "Array"}`, "Array"},
		{`{"readonly": "simple"}`, "simple"},
		{`"generic"`, "generic"},
	} {
		_, err := DecodeArrayTypeOptions([]byte(refused.wire))
		if err == nil {
			t.Errorf("%s decoded; it must be refused", refused.wire)
			continue
		}
		if !strings.Contains(err.Error(), refused.named) {
			t.Errorf("%s was refused without naming %s: %v", refused.wire, refused.named, err)
		}
	}
}
