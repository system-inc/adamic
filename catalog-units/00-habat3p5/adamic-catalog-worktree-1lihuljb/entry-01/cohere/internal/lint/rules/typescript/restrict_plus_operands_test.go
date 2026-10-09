package typescript

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus for restrict-plus-operands, taken verbatim from upstream's own tester.
//
// Extracted by PARSING the upstream test file with the TypeScript compiler and serialising each
// string through a JSON encoder, so nothing was retyped or passed through a shell, then byte-compared
// against the source with a control asserted absent: 115 of the 119 cases match the source literally
// and the remaining four match under the backtick escaping a template literal requires, because they
// carry TypeScript template LITERAL types whose `${...}` has to survive.
//
// That last part was a real defect caught by the byte check rather than a formality. A first
// extractor read a template's cooked `.text`, which evaluates `${string}` away, and produced four
// cases that were valid JSON, compiled fine, and tested something upstream never wrote. A second
// version kept the raw source and left the backslashes in, which the parser then rejected outright.
// Only the third is right, and only the byte comparison could tell the three apart.
//
// The whole corpus was additionally driven through the INSTALLED rule at 8.67.0 on a real type graph,
// honouring each case's own options. It agreed with upstream's recorded expectations on all 119
// inputs and reproduced all 56 recorded message texts exactly, so the oracle used to settle the
// questions below is pinned against the corpus rather than trusted.
const restrictPlusOperandsFile = "file.ts"

// restrictPlusOperandsBool is the address-taking helper the option struct needs.
//
// Every `allow` field is a *bool because all five default to TRUE, so a zero-valued struct would
// mean "forbid everything" and invert the rule.
func restrictPlusOperandsBool(value bool) *bool { return &value }

// TestRestrictPlusOperandsStaysSilent is upstream's valid list, each under its own options.
func TestRestrictPlusOperandsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
	}{
		{"upstream valid 0", "let x = 5;", nil},
		{"upstream valid 1", "let y = '10';", nil},
		{"upstream valid 2", "let z = 8.2;", nil},
		{"upstream valid 3", "let w = '6.5';", nil},
		{"upstream valid 4", "let foo = 5 + 10;", nil},
		{"upstream valid 5", "let foo = '5.5' + '10';", nil},
		{"upstream valid 6", "let foo = parseInt('5.5', 10) + 10;", nil},
		{"upstream valid 7", "let foo = parseFloat('5.5', 10) + 10;", nil},
		{"upstream valid 8", "let foo = 1n + 1n;", nil},
		{"upstream valid 9", "let foo = BigInt(1) + 1n;", nil},
		{"upstream valid 10", "\nlet foo = 1n;\nfoo + 2n;\n    ", nil},
		{"upstream valid 11", "\nfunction test(s: string, n: number): number {\n  return 2;\n}\nlet foo = test('5.5', 10) + 10;\n    ", nil},
		{"upstream valid 12", "\nlet x = 5;\nlet z = 8.2;\nlet foo = x + z;\n    ", nil},
		{"upstream valid 13", "\nlet w = '6.5';\nlet y = '10';\nlet foo = y + w;\n    ", nil},
		{"upstream valid 14", "let foo = 1 + 1;", nil},
		{"upstream valid 15", "let foo = '1' + '1';", nil},
		{"upstream valid 16", "\nlet pair: { first: number; second: string } = { first: 5, second: '10' };\nlet foo = pair.first + 10;\n    ", nil},
		{"upstream valid 17", "\nlet pair: { first: number; second: string } = { first: 5, second: '10' };\nlet foo = pair.first + (10 as number);\n    ", nil},
		{"upstream valid 18", "\nlet pair: { first: number; second: string } = { first: 5, second: '10' };\nlet foo = '5.5' + pair.second;\n    ", nil},
		{"upstream valid 19", "\nlet pair: { first: number; second: string } = { first: 5, second: '10' };\nlet foo = ('5.5' as string) + pair.second;\n    ", nil},
		{"upstream valid 20", "\nconst foo =\n  'hello' +\n  (someBoolean ? 'a' : 'b') +\n  (() => (someBoolean ? 'c' : 'd'))() +\n  'e';\n    ", nil},
		{"upstream valid 21", "const balls = true;", nil},
		{"upstream valid 22", "balls === true;", nil},
		{"upstream valid 23", "\nfunction foo<T extends string>(a: T) {\n  return a + '';\n}\n    ", nil},
		{"upstream valid 24", "\nfunction foo<T extends 'a' | 'b'>(a: T) {\n  return a + '';\n}\n    ", nil},
		{"upstream valid 25", "\nfunction foo<T extends number>(a: T) {\n  return a + 1;\n}\n    ", nil},
		{"upstream valid 26", "\nfunction foo<T extends 1>(a: T) {\n  return a + 1;\n}\n    ", nil},
		{"upstream valid 27", "\ndeclare const a: {} & string;\ndeclare const b: string;\nconst x = a + b;\n    ", nil},
		{"upstream valid 28", "\ndeclare const a: unknown & string;\ndeclare const b: string;\nconst x = a + b;\n    ", nil},
		{"upstream valid 29", "\ndeclare const a: string & string;\ndeclare const b: string;\nconst x = a + b;\n    ", nil},
		{"upstream valid 30", "\ndeclare const a: 'string literal' & string;\ndeclare const b: string;\nconst x = a + b;\n    ", nil},
		{"upstream valid 31", "\ndeclare const a: {} & number;\ndeclare const b: number;\nconst x = a + b;\n    ", nil},
		{"upstream valid 32", "\ndeclare const a: unknown & number;\ndeclare const b: number;\nconst x = a + b;\n    ", nil},
		{"upstream valid 33", "\ndeclare const a: number & number;\ndeclare const b: number;\nconst x = a + b;\n    ", nil},
		{"upstream valid 34", "\ndeclare const a: 42 & number;\ndeclare const b: number;\nconst x = a + b;\n    ", nil},
		{"upstream valid 35", "\ndeclare const a: {} & bigint;\ndeclare const b: bigint;\nconst x = a + b;\n    ", nil},
		{"upstream valid 36", "\ndeclare const a: unknown & bigint;\ndeclare const b: bigint;\nconst x = a + b;\n    ", nil},
		{"upstream valid 37", "\ndeclare const a: bigint & bigint;\ndeclare const b: bigint;\nconst x = a + b;\n    ", nil},
		{"upstream valid 38", "\ndeclare const a: 42n & bigint;\ndeclare const b: bigint;\nconst x = a + b;\n    ", nil},
		{"upstream valid 39", "\nfunction A(s: string) {\n  return `a${s}b` as const;\n}\nconst b = A('') + '!';\n    ", nil},
		{"upstream valid 40", "\ndeclare const a: `template${string}`;\ndeclare const b: '';\nconst x = a + b;\n    ", nil},
		{"upstream valid 41", "\nconst a: `template${0}`;\ndeclare const b: '';\nconst x = a + b;\n    ", nil},
		{"upstream valid 42", "\ndeclare const a: RegExp;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(true)}},
		{"upstream valid 43", "\nconst a = /regexp/;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(true)}},
		{"upstream valid 44", "\nconst f = (a: RegExp, b: RegExp) => a + b;\n      ", RestrictPlusOperandsOptions{AllowRegExp: restrictPlusOperandsBool(true)}},
		{"upstream valid 45", "\nlet foo: string | undefined;\nfoo = foo + 'some data';\n      ", RestrictPlusOperandsOptions{AllowNullish: restrictPlusOperandsBool(true)}},
		{"upstream valid 46", "\nlet foo: string | null;\nfoo = foo + 'some data';\n      ", RestrictPlusOperandsOptions{AllowNullish: restrictPlusOperandsBool(true)}},
		{"upstream valid 47", "\nlet foo: string | null | undefined;\nfoo = foo + 'some data';\n      ", RestrictPlusOperandsOptions{AllowNullish: restrictPlusOperandsBool(true)}},
		{"upstream valid 48", "\nlet foo = '';\nfoo += 0;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false), SkipCompoundAssignments: true}},
		{"upstream valid 49", "\nlet foo = 0;\nfoo += '';\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false), SkipCompoundAssignments: true}},
		{"upstream valid 50", "\nconst f = (a: any, b: any) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true)}},
		{"upstream valid 51", "\nconst f = (a: any, b: string) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true)}},
		{"upstream valid 52", "\nconst f = (a: any, b: bigint) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true)}},
		{"upstream valid 53", "\nconst f = (a: any, b: number) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true)}},
		{"upstream valid 54", "\nconst f = (a: any, b: boolean) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true), AllowBoolean: restrictPlusOperandsBool(true)}},
		{"upstream valid 55", "\nconst f = (a: string, b: string | number) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true), AllowBoolean: restrictPlusOperandsBool(true), AllowNullish: restrictPlusOperandsBool(true), AllowNumberAndString: restrictPlusOperandsBool(true), AllowRegExp: restrictPlusOperandsBool(true)}},
		{"upstream valid 56", "\nconst f = (a: string | number, b: number) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true), AllowBoolean: restrictPlusOperandsBool(true), AllowNullish: restrictPlusOperandsBool(true), AllowNumberAndString: restrictPlusOperandsBool(true), AllowRegExp: restrictPlusOperandsBool(true)}},
		{"upstream valid 57", "\nconst f = (a: string | number, b: string | number) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true), AllowBoolean: restrictPlusOperandsBool(true), AllowNullish: restrictPlusOperandsBool(true), AllowNumberAndString: restrictPlusOperandsBool(true), AllowRegExp: restrictPlusOperandsBool(true)}},
		{"upstream valid 58", "let foo = '1' + 1n;", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(true)}}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, RestrictPlusOperands, restrictPlusOperandsFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestRestrictPlusOperandsFires is upstream's invalid list, with the ids it records per case.
//
// Several cases report TWICE, once per operand, and those are the ones that pin the ordering between
// the two passes: an expression whose operands are each individually impossible reports both operands
// and never the pair, because the pair pass is gated on the first pass having found nothing.
func TestRestrictPlusOperandsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantIds    []string
	}{
		{"upstream invalid 0", "let foo = '1' + 1;", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 1", "let foo = '1' + 1;", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 2", "let foo = [] + {};", nil, []string{"invalid", "invalid"}},
		{"upstream invalid 3", "let foo = 5 + '10';", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 4", "let foo = [] + 5;", nil, []string{"invalid"}},
		{"upstream invalid 5", "let foo = [] + [];", nil, []string{"invalid", "invalid"}},
		{"upstream invalid 6", "let foo = 5 + [3];", nil, []string{"invalid"}},
		{"upstream invalid 7", "let foo = '5' + {};", nil, []string{"invalid"}},
		{"upstream invalid 8", "let foo = 5.5 + '5';", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 9", "let foo = '5.5' + 5;", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 10", "\nlet x = 5;\nlet y = '10';\nlet foo = x + y;\n      ", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 11", "\nlet x = 5;\nlet y = '10';\nlet foo = y + x;\n      ", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 12", "\nlet x = 5;\nlet foo = x + {};\n      ", nil, []string{"invalid"}},
		{"upstream invalid 13", "\nlet y = '10';\nlet foo = [] + y;\n      ", nil, []string{"invalid"}},
		{"upstream invalid 14", "\nlet pair = { first: 5, second: '10' };\nlet foo = pair + pair;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid", "invalid"}},
		{"upstream invalid 15", "\ntype Valued = { value: number };\nlet value: Valued = { value: 0 };\nlet combined = value + 0;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 16", "let foo = 1n + 1;", nil, []string{"bigintAndNumber"}},
		{"upstream invalid 17", "let foo = 1 + 1n;", nil, []string{"bigintAndNumber"}},
		{"upstream invalid 18", "\nlet foo = 1n;\nfoo + 1;\n      ", nil, []string{"bigintAndNumber"}},
		{"upstream invalid 19", "\nlet foo = 1;\nfoo + 1n;\n      ", nil, []string{"bigintAndNumber"}},
		{"upstream invalid 20", "\nfunction foo<T extends string>(a: T) {\n  return a + 1;\n}\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 21", "\nfunction foo<T extends 'a' | 'b'>(a: T) {\n  return a + 1;\n}\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 22", "\nfunction foo<T extends number>(a: T) {\n  return a + '';\n}\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 23", "\nfunction foo<T extends 1>(a: T) {\n  return a + '';\n}\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 24", "\ndeclare const a: `template${number}`;\ndeclare const b: number;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 25", "\ndeclare const a: never;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 26", "\ndeclare const a: never & string;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 27", "\ndeclare const a: boolean & string;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 28", "\ndeclare const a: any & string;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 29", "\ndeclare const a: { a: 1 } & { b: 2 };\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 30", "\ninterface A {\n  a: 1;\n}\ndeclare const a: A;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 31", "\ninterface A {\n  a: 1;\n}\ninterface A2 extends A {\n  b: 2;\n}\ndeclare const a: A2;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 32", "\ntype A = { a: 1 } & { b: 2 };\ndeclare const a: A;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 33", "\ndeclare const a: { a: 1 } & { b: 2 };\ndeclare const b: number;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 34", "\ndeclare const a: never;\ndeclare const b: bigint;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 35", "\ndeclare const a: any;\ndeclare const b: bigint;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 36", "\ndeclare const a: { a: 1 } & { b: 2 };\ndeclare const b: bigint;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 37", "\ndeclare const a: RegExp;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 38", "\nconst a = /regexp/;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 39", "\ndeclare const a: Symbol;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 40", "\ndeclare const a: symbol;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 41", "\ndeclare const a: unique symbol;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 42", "\nconst a = Symbol('');\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 43", "\nlet foo: string | undefined;\nfoo += 'some data';\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false), SkipCompoundAssignments: false}, []string{"invalid"}},
		{"upstream invalid 44", "\nlet foo: string | null;\nfoo += 'some data';\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 45", "\nlet foo: string = '';\nfoo += 1;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 46", "\nlet foo = 0;\nfoo += '';\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"mismatched"}},
		{"upstream invalid 47", "\nconst f = (a: any, b: boolean) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true), AllowBoolean: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 48", "\nconst f = (a: any, b: []) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true)}, []string{"invalid"}},
		{"upstream invalid 49", "\nconst f = (a: any, b: boolean) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(true)}, []string{"invalid"}},
		{"upstream invalid 50", "\nconst f = (a: any, b: any) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false)}, []string{"invalid", "invalid"}},
		{"upstream invalid 51", "\nconst f = (a: any, b: string) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 52", "\nconst f = (a: any, b: bigint) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 53", "\nconst f = (a: any, b: number) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 54", "\nconst f = (a: any, b: boolean) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false)}, []string{"invalid", "invalid"}},
		{"upstream invalid 55", "\nconst f = (a: number, b: RegExp) => a + b;\n      ", RestrictPlusOperandsOptions{AllowRegExp: restrictPlusOperandsBool(true)}, []string{"invalid"}},
		{"upstream invalid 56", "\nlet foo: string | boolean;\nfoo = foo + 'some data';\n      ", RestrictPlusOperandsOptions{AllowBoolean: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 57", "\nlet foo: boolean;\nfoo = foo + 'some data';\n      ", RestrictPlusOperandsOptions{AllowBoolean: restrictPlusOperandsBool(false)}, []string{"invalid"}},
		{"upstream invalid 58", "\nconst f = (a: any, b: unknown) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true), AllowBoolean: restrictPlusOperandsBool(true), AllowNullish: restrictPlusOperandsBool(true), AllowRegExp: restrictPlusOperandsBool(true)}, []string{"invalid"}},
		{"upstream invalid 59", "let foo = '1' + 1n;", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"mismatched"}}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, RestrictPlusOperands, restrictPlusOperandsFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestRestrictPlusOperandsMessages asserts the RENDERED text of every finding, exactly.
//
// This is not optional for this rule and no id assertion above can stand in for it. All three
// messages interpolate, and one slot, `stringLike`, is assembled from which `allow` options are on,
// so the same finding reads differently under different configuration. The other slots carry a type
// rendered by the checker, where the difference between `string` and `string | boolean` is the whole
// content of the finding.
//
// The expectations are the strings the INSTALLED rule produced for these inputs, which were
// themselves checked against upstream's recorded `data` for all 56 findings that carry it.
func TestRestrictPlusOperandsMessages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		options      any
		wantMessages []string
	}{
		{"upstream invalid 0", "let foo = '1' + 1;", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `string` + `number`."}},
		{"upstream invalid 1", "let foo = '1' + 1;", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string. Got `string` + `number`."}},
		{"upstream invalid 2", "let foo = [] + {};", nil, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `never[]`.", "Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `{}`."}},
		{"upstream invalid 3", "let foo = 5 + '10';", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string. Got `number` + `string`."}},
		{"upstream invalid 4", "let foo = [] + 5;", nil, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `never[]`."}},
		{"upstream invalid 5", "let foo = [] + [];", nil, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `never[]`.", "Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `never[]`."}},
		{"upstream invalid 6", "let foo = 5 + [3];", nil, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `number[]`."}},
		{"upstream invalid 7", "let foo = '5' + {};", nil, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `{}`."}},
		{"upstream invalid 8", "let foo = 5.5 + '5';", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `number` + `string`."}},
		{"upstream invalid 9", "let foo = '5.5' + 5;", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `string` + `number`."}},
		{"upstream invalid 10", "\nlet x = 5;\nlet y = '10';\nlet foo = x + y;\n      ", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `number` + `string`."}},
		{"upstream invalid 11", "\nlet x = 5;\nlet y = '10';\nlet foo = y + x;\n      ", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `string` + `number`."}},
		{"upstream invalid 12", "\nlet x = 5;\nlet foo = x + {};\n      ", nil, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `{}`."}},
		{"upstream invalid 13", "\nlet y = '10';\nlet foo = [] + y;\n      ", nil, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `never[]`."}},
		{"upstream invalid 14", "\nlet pair = { first: 5, second: '10' };\nlet foo = pair + pair;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `{ first: number; second: string; }`.", "Invalid operand for a '+' operation. Operands must each be a number or string. Got `{ first: number; second: string; }`."}},
		{"upstream invalid 15", "\ntype Valued = { value: number };\nlet value: Valued = { value: 0 };\nlet combined = value + 0;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `Valued`."}},
		{"upstream invalid 16", "let foo = 1n + 1;", nil, []string{"Numeric '+' operations must either be both bigints or both numbers. Got `bigint` + `number`."}},
		{"upstream invalid 17", "let foo = 1 + 1n;", nil, []string{"Numeric '+' operations must either be both bigints or both numbers. Got `number` + `bigint`."}},
		{"upstream invalid 18", "\nlet foo = 1n;\nfoo + 1;\n      ", nil, []string{"Numeric '+' operations must either be both bigints or both numbers. Got `bigint` + `number`."}},
		{"upstream invalid 19", "\nlet foo = 1;\nfoo + 1n;\n      ", nil, []string{"Numeric '+' operations must either be both bigints or both numbers. Got `number` + `bigint`."}},
		{"upstream invalid 20", "\nfunction foo<T extends string>(a: T) {\n  return a + 1;\n}\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string. Got `string` + `number`."}},
		{"upstream invalid 21", "\nfunction foo<T extends 'a' | 'b'>(a: T) {\n  return a + 1;\n}\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string. Got `string` + `number`."}},
		{"upstream invalid 22", "\nfunction foo<T extends number>(a: T) {\n  return a + '';\n}\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string. Got `number` + `string`."}},
		{"upstream invalid 23", "\nfunction foo<T extends 1>(a: T) {\n  return a + '';\n}\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string. Got `number` + `string`."}},
		{"upstream invalid 24", "\ndeclare const a: `template${number}`;\ndeclare const b: number;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string. Got `string` + `number`."}},
		{"upstream invalid 25", "\ndeclare const a: never;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `never`."}},
		{"upstream invalid 26", "\ndeclare const a: never & string;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `never`."}},
		{"upstream invalid 27", "\ndeclare const a: boolean & string;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `never`."}},
		{"upstream invalid 28", "\ndeclare const a: any & string;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `any`."}},
		{"upstream invalid 29", "\ndeclare const a: { a: 1 } & { b: 2 };\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `{ a: 1; } & { b: 2; }`."}},
		{"upstream invalid 30", "\ninterface A {\n  a: 1;\n}\ndeclare const a: A;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `A`."}},
		{"upstream invalid 31", "\ninterface A {\n  a: 1;\n}\ninterface A2 extends A {\n  b: 2;\n}\ndeclare const a: A2;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `A2`."}},
		{"upstream invalid 32", "\ntype A = { a: 1 } & { b: 2 };\ndeclare const a: A;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `A`."}},
		{"upstream invalid 33", "\ndeclare const a: { a: 1 } & { b: 2 };\ndeclare const b: number;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `{ a: 1; } & { b: 2; }`."}},
		{"upstream invalid 34", "\ndeclare const a: never;\ndeclare const b: bigint;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `never`."}},
		{"upstream invalid 35", "\ndeclare const a: any;\ndeclare const b: bigint;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `any`."}},
		{"upstream invalid 36", "\ndeclare const a: { a: 1 } & { b: 2 };\ndeclare const b: bigint;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `{ a: 1; } & { b: 2; }`."}},
		{"upstream invalid 37", "\ndeclare const a: RegExp;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `RegExp`."}},
		{"upstream invalid 38", "\nconst a = /regexp/;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `RegExp`."}},
		{"upstream invalid 39", "\ndeclare const a: Symbol;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `Symbol`."}},
		{"upstream invalid 40", "\ndeclare const a: symbol;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `symbol`."}},
		{"upstream invalid 41", "\ndeclare const a: unique symbol;\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `unique symbol`."}},
		{"upstream invalid 42", "\nconst a = Symbol('');\ndeclare const b: string;\nconst x = a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `unique symbol`."}},
		{"upstream invalid 43", "\nlet foo: string | undefined;\nfoo += 'some data';\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false), SkipCompoundAssignments: false}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `string | undefined`."}},
		{"upstream invalid 44", "\nlet foo: string | null;\nfoo += 'some data';\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string. Got `string | null`."}},
		{"upstream invalid 45", "\nlet foo: string = '';\nfoo += 1;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string. Got `string` + `number`."}},
		{"upstream invalid 46", "\nlet foo = 0;\nfoo += '';\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string. Got `number` + `string`."}},
		{"upstream invalid 48", "\nconst f = (a: any, b: []) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `[]`."}},
		{"upstream invalid 49", "\nconst f = (a: any, b: boolean) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(true)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `boolean`, `null`, `RegExp`, `undefined`. Got `any`."}},
		{"upstream invalid 56", "\nlet foo: string | boolean;\nfoo = foo + 'some data';\n      ", RestrictPlusOperandsOptions{AllowBoolean: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `null`, `RegExp`, `undefined`. Got `string | boolean`."}},
		{"upstream invalid 57", "\nlet foo: boolean;\nfoo = foo + 'some data';\n      ", RestrictPlusOperandsOptions{AllowBoolean: restrictPlusOperandsBool(false)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `null`, `RegExp`, `undefined`. Got `boolean`."}},
		{"upstream invalid 58", "\nconst f = (a: any, b: unknown) => a + b;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(true), AllowBoolean: restrictPlusOperandsBool(true), AllowNullish: restrictPlusOperandsBool(true), AllowRegExp: restrictPlusOperandsBool(true)}, []string{"Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `unknown`."}},
		{"upstream invalid 59", "let foo = '1' + 1n;", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"Operands of '+' operations must be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got `string` + `bigint`."}}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, RestrictPlusOperands, restrictPlusOperandsFile, testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != len(testCase.wantMessages) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantMessages), len(result.Diagnostics))
			}
			for index, want := range testCase.wantMessages {
				if got := result.Diagnostics[index].Message.Description; got != want {
					t.Errorf("finding %d reads\n%q\nwant\n%q", index, got, want)
				}
			}
		})
	}
}

// TestRestrictPlusOperandsSpans asserts WHERE each finding points, which no message assertion above
// can see.
//
// The two passes point at different nodes and that difference is the rule's whole shape. An
// individually impossible operand reports at the OPERAND, so the reader is shown which side is
// wrong; a mismatched or bigint-and-number pair reports at the WHOLE EXPRESSION, because neither
// side is wrong alone. The rows below carry both kinds and a two-operand case, so a rule that
// collapsed the distinction in either direction fails here while passing every id fixture.
//
// The expected text is sliced from upstream's recorded columns, then compared against the text our
// finding actually covers.
func TestRestrictPlusOperandsSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantSpans  []string
	}{
		{"upstream invalid 0", "let foo = '1' + 1;", RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}, []string{"'1' + 1"}},
		{"upstream invalid 2", "let foo = [] + {};", nil, []string{"[]", "{}"}},
		{"upstream invalid 4", "let foo = [] + 5;", nil, []string{"[]"}},
		{"upstream invalid 14", "\nlet pair = { first: 5, second: '10' };\nlet foo = pair + pair;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"pair", "pair"}},
		{"upstream invalid 15", "\ntype Valued = { value: number };\nlet value: Valued = { value: 0 };\nlet combined = value + 0;\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false)}, []string{"value"}},
		{"upstream invalid 16", "let foo = 1n + 1;", nil, []string{"1n + 1"}},
		{"upstream invalid 43", "\nlet foo: string | undefined;\nfoo += 'some data';\n      ", RestrictPlusOperandsOptions{AllowAny: restrictPlusOperandsBool(false), AllowBoolean: restrictPlusOperandsBool(false), AllowNullish: restrictPlusOperandsBool(false), AllowNumberAndString: restrictPlusOperandsBool(false), AllowRegExp: restrictPlusOperandsBool(false), SkipCompoundAssignments: false}, []string{"foo"}}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, RestrictPlusOperands, restrictPlusOperandsFile, testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			// The harness trims the fixture before building the program, so slice the text the
			// harness actually wrote rather than the literal above.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			for index, want := range testCase.wantSpans {
				reported := source[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != want {
					t.Errorf("finding %d points at %q, want %q", index, reported, want)
				}
			}
		})
	}
}

// TestRestrictPlusOperandsDecoder routes configuration through the rule's own exported decoder,
// which is the only thing that puts the default inversion under test.
//
// Every `allow` option DEFAULTS TO TRUE, so the zero-valued struct means the opposite of the
// default and a rule reading it would report most of a real tree. Building the options struct
// directly in a fixture, as every other test in this file does, cannot see that: the struct is
// already correct by the time the rule reads it. Only the wire format can be wrong.
//
// So these rows start from JSON, exactly as the config layer delivers it, and assert three things
// the struct-built fixtures structurally cannot:
//
//   - an ABSENT key falls back to true rather than to the zero value
//   - an explicit `false` is distinguishable from an absent key, which is why the fields are *bool
//   - a rule configured as a bare "error", which arrives as NIL options, gets the defaults
func TestRestrictPlusOperandsDecoder(t *testing.T) {
	t.Parallel()

	decode := func(t *testing.T, raw string) restrictPlusOperandsSettings {
		t.Helper()
		decoded, err := rule.DecodeOptionsInto[RestrictPlusOperandsOptions]()(json.RawMessage(raw))
		if err != nil {
			t.Fatalf("decoding %s: %v", raw, err)
		}
		return restrictPlusOperandsSettingsFrom(decoded)
	}

	t.Run("an empty object leaves every allow at its true default", func(t *testing.T) {
		t.Parallel()
		got := decode(t, `{}`)
		if got != DefaultRestrictPlusOperandsSettings() {
			t.Errorf("got %+v, want the defaults %+v", got, DefaultRestrictPlusOperandsSettings())
		}
	})

	t.Run("nil options are the defaults rather than the zero struct", func(t *testing.T) {
		t.Parallel()
		if got := restrictPlusOperandsSettingsFrom(nil); got != DefaultRestrictPlusOperandsSettings() {
			t.Errorf("got %+v, want the defaults %+v", got, DefaultRestrictPlusOperandsSettings())
		}
	})

	t.Run("an explicit false is carried through", func(t *testing.T) {
		t.Parallel()
		got := decode(t, `{"allowNumberAndString": false}`)
		if got.allowNumberAndString {
			t.Error("allowNumberAndString decoded as true from an explicit false")
		}
		if !got.allowAny || !got.allowBoolean || !got.allowNullish || !got.allowRegExp {
			t.Errorf("setting one option to false disturbed the others: %+v", got)
		}
	})

	t.Run("the decoded options reach the rule and change its verdict", func(t *testing.T) {
		t.Parallel()
		// The end-to-end half. Byte-identical source, opposite verdicts, separated only by what
		// came off the wire.
		const source = "let foo = '1' + 1;"

		permissive := decode(t, `{}`)
		strict := decode(t, `{"allowNumberAndString": false}`)
		if permissive.allowNumberAndString == strict.allowNumberAndString {
			t.Fatal("the two configurations decode the same, so this proves nothing")
		}

		rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, RestrictPlusOperands, restrictPlusOperandsFile, source,
			RestrictPlusOperandsOptions{}))
		rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, RestrictPlusOperands, restrictPlusOperandsFile, source,
			RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)}), "mismatched")
	})
}

// TestRestrictPlusOperandsStringLikeRenderings pins all three spellings of the interpolated slot.
//
// The SINGULAR spelling is the one worth writing this test for: it is reachable, and no case in
// upstream's corpus renders it, because none leaves exactly one `allow` option on. So it is a live
// branch with no imported coverage, which is precisely the shape that ships wrong and stays wrong.
// The configuration below is constructed rather than imported for that reason.
//
// The plural rendering also pins the ORDER, which is upstream's and is not alphabetical: `null` and
// `undefined` are both gated on allowNullish and sit on either side of `RegExp`.
func TestRestrictPlusOperandsStringLikeRenderings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		settings restrictPlusOperandsSettings
		want     string
	}{
		{
			"everything allowed, which is the default",
			DefaultRestrictPlusOperandsSettings(),
			"string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`",
		},
		{
			"nothing allowed",
			restrictPlusOperandsSettings{},
			"string",
		},
		{
			"exactly one allowed, which upstream's corpus never renders",
			restrictPlusOperandsSettings{allowRegExp: true},
			"string, allowing a string + `RegExp`",
		},
		{
			"allowNullish contributes two entries on either side of RegExp",
			restrictPlusOperandsSettings{allowNullish: true, allowRegExp: true},
			"string, allowing a string + any of: `null`, `RegExp`, `undefined`",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := restrictPlusOperandsStringLike(testCase.settings); got != testCase.want {
				t.Errorf("rendered\n%q\nwant\n%q", got, testCase.want)
			}
		})
	}
}

// TestRestrictPlusOperandsRequiresTheTypedHarness asserts the rule declares the checker and that the
// plain harness cannot prove it, so a later revert to rule_testing.Run fails loudly.
func TestRestrictPlusOperandsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !RestrictPlusOperands.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so every typed fixture would run against a nil checker")
	}

	const source = "let foo = 1n + 1;"

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, RestrictPlusOperands, restrictPlusOperandsFile, source), "bigintAndNumber")
	rule_testing.ExpectClean(t, rule_testing.Run(t, RestrictPlusOperands, restrictPlusOperandsFile, source))
}

// TestRestrictPlusOperandsSurvivesMalformedExpressions runs the rule over `+` shapes where an
// operand is absent or is not what the well-formed grammar produces.
//
// A panic costs every rule its verdict on the whole file rather than costing this rule one finding,
// and `+` is among the most common tokens in any tree, so this rule is offered far more nodes than
// most. The shapes are deliberately malformed, because error recovery synthesizes nodes that
// well-formed source never produces.
func TestRestrictPlusOperandsSurvivesMalformedExpressions(t *testing.T) {
	t.Parallel()

	sources := []string{
		"const x = 1 + ;\n",
		"const x =  + 1;\n",
		"const x = + ;\n",
		"const x = 1 +;\n",
		"let x; x += ;\n",
		"let x; x + ;\n",
		"const x = (1) + (2);\n",
		"const x = ((1 + 2)) + ((3));\n",
		"const x = 1 + 2 + 3 + 4 + 5;\n",
		"declare const a: never; const x = a + a;\n",
		"declare const a: unknown; const x = a + a;\n",
		"const x = void 0 + void 0;\n",
		"const x = `a` + `b`;\n",
		"const x = 1 + `${1}`;\n",
		"enum E { A } const x = E.A + E.A;\n",
		"const x = null + undefined;\n",
		"class C { z = 1 + 1; }\n",
		"const x = [1] + [2];\n",
	}

	for index, source := range sources {
		t.Run(fmt.Sprintf("shape-%d", index), func(t *testing.T) {
			t.Parallel()
			// A panic fails the test. Findings are deliberately unasserted: what the rule concludes
			// about a malformed shape belongs in its own fixture.
			rule_testing.RunTyped(t, RestrictPlusOperands, restrictPlusOperandsFile, source)
		})
	}
}

// TestRestrictPlusOperandsIndividualComplaintSuppressesThePair pins the ordering between the two
// passes, which upstream's own corpus never separates.
//
// An expression can be wrong in both ways at once: an operand that is individually impossible AND a
// pair the second pass would also complain about. Upstream reports the operand and stops, because
// its `hadIndividualComplaint` flag gates the pair passes entirely. Removing that gate survived all
// 174 imported rows, so no case upstream ships is wrong in both ways at the same time.
//
// The first input tried for it was NOT distinguishing, and the reason is worth keeping. `string |
// symbol` added to a number has an impossible operand and looks like a string-plus-number mismatch,
// but `allowNumberAndString` DEFAULTS TO TRUE, so the mismatched arm never fires under defaults and
// both versions report once. The bigint arm is the one no option gates, so it is the only second-pass
// arm reachable under the default configuration. Measured with the gate removed:
//
//	string | symbol + 1     [invalid]  both versions       does NOT distinguish
//	number | {} + 1n        [invalid] against [invalid bigintAndNumber]   distinguishes
//
// The non-distinguishing row is kept below as the control that makes that visible.
func TestRestrictPlusOperandsIndividualComplaintSuppressesThePair(t *testing.T) {
	t.Parallel()

	t.Run("an impossible operand suppresses the bigint pair complaint", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTypedWithOptions(t, RestrictPlusOperands, restrictPlusOperandsFile,
			"declare const a: number | {};\nconst x = a + 1n;", nil)
		rule_testing.ExpectFindings(t, result, "invalid")
	})

	t.Run("the mismatched arm is off by default, so this shape cannot see the gate", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTypedWithOptions(t, RestrictPlusOperands, restrictPlusOperandsFile,
			"declare const a: string | symbol;\nconst x = a + 1;", nil)
		rule_testing.ExpectFindings(t, result, "invalid")
	})

	t.Run("with the mismatched arm on, the same suppression holds", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTypedWithOptions(t, RestrictPlusOperands, restrictPlusOperandsFile,
			"declare const a: string | symbol;\nconst x = a + 1;",
			RestrictPlusOperandsOptions{AllowNumberAndString: restrictPlusOperandsBool(false)})
		rule_testing.ExpectFindings(t, result, "invalid")
	})
}
