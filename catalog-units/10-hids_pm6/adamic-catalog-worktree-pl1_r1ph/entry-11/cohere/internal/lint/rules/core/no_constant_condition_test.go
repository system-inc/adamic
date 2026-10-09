package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const constantConditionFile = "/repository/source/Condition.ts"

const constantConditionDeclarations = "declare let a: any, b: any;\n"

func TestNoConstantConditionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a true literal condition", "export function run() { if(true) {} }\n"},
		{"a false literal condition", "export function run() { if(false) {} }\n"},
		{"a numeric condition", "export function run() { if(1) {} }\n"},
		{"a string condition", "export function run() { if('x') {} }\n"},
		{"a null condition", "export function run() { if(null) {} }\n"},
		{"an object literal condition", "export function run() { if({}) {} }\n"},
		{"an array literal condition", "export function run() { if([]) {} }\n"},
		// An array is truthy whatever it holds, so a non-empty one is constant as a condition even
		// though it is not constant as a value. The empty case above fires either way and so does
		// not measure the boolean-position branch; this one does. Found by making arrays
		// non-constant in boolean position and watching the suite stay green.
		{"an array holding a variable, as a condition", "export function run() { if([a]) {} }\n"},
		// The same asymmetry through a negation, which reads its operand as a condition. `![a]` is
		// constant only because the flag flips; passing the position through makes it vary.
		{"a negated array holding a variable", "export function run() { if(![a]) {} }\n"},
		{"an arrow function condition", "export function run() { if(() => 1) {} }\n"},

		// The typo this rule exists for: an assignment read as a comparison.
		{"an assignment in a condition", "export function run() { if(a = 0) {} }\n"},
		{"an assignment of a constant", "export function run() { if(a = 'x') {} }\n"},

		// Short circuit. The left operand alone decides, so the whole is constant even though `a`
		// is not, which is why a logical expression is not simply both-sides-constant.
		{"true on the left of an or", "export function run() { if(true || a) {} }\n"},
		{"false on the left of an and", "export function run() { if(false && a) {} }\n"},
		{"true on the right of an or in a condition", "export function run() { if(a || true) {} }\n"},
		// A logical identity nested under the same operator. `false` decides an `&&` at any depth,
		// so `a && (b && false)` is constant.
		{"a nested identity under the same operator", "export function run() { if(a && (b && false)) {} }\n"},
		// A nested logical that is itself an identity for the outer operator. `true || false` is
		// truthy and decides an `||`, so the whole condition is constant despite `a`.
		{"a nested or that decides an outer or", "export function run() { if((true || false) || a) {} }\n"},
		{"a nested and that decides an outer and", "export function run() { if((false && true) && a) {} }\n"},

		{"a negation of a constant", "export function run() { if(!true) {} }\n"},
		{"a void expression", "export function run() { if(void a) {} }\n"},
		{"a typeof in a condition", "export function run() { if(typeof a) {} }\n"},
		{"a new expression in a condition", "export function run() { if(new Date()) {} }\n"},
		{"undefined", "export function run() { if(undefined) {} }\n"},
		{"a Boolean call with no arguments", "export function run() { if(Boolean()) {} }\n"},
		{"a Boolean call with a constant", "export function run() { if(Boolean(1)) {} }\n"},
		{"a sequence ending in a constant", "export function run() { if((a, 1)) {} }\n"},
		{"a template with a static chunk", "export function run() { if(`x${a}`) {} }\n"},
		{"a constant comparison", "export function run() { if(1 < 2) {} }\n"},

		{"a ternary condition", "export function run() { return true ? a : b; }\n"},

		// Loops other than `while(true)`, which the default exempts.
		{"a while on a numeric literal", "export function run() { while(1) {} }\n"},
		{"a while on a false literal", "export function run() { while(false) {} }\n"},
		{"a while on a short circuit", "export function run() { while(a || true) {} }\n"},
		{"a do-while on true", "export function run() { do {} while(true); }\n"},
		{"a for on a constant condition", "export function run() { for(let i = 0; true; i++) {} }\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoConstantCondition, constantConditionFile,
				constantConditionDeclarations+testCase.sourceText), "unexpected")
		})
	}
}

func TestNoConstantConditionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a variable condition", "export function run() { if(a) {} }\n"},
		{"a comparison of variables", "export function run() { if(a === b) {} }\n"},
		{"a logical of variables", "export function run() { if(a && b) {} }\n"},
		{"a call condition", "export function run() { declare function f(): boolean; if(f()) {} }\n"},
		{"a member condition", "export function run() { if(a.b) {} }\n"},

		// The idiom the default exempts. This is the whole reason the option's default is
		// allExceptWhileTrue rather than all.
		{"a while-true loop", "export function run() { while(true) { break; } }\n"},
		{"a parenthesized while-true loop", "export function run() { while((true)) { break; } }\n"},
		{"a for with no condition", "export function run() { for(;;) { break; } }\n"},

		// `in` reads a property that may not be there, so it varies between two constants. ESLint
		// excludes it explicitly, and this is the boundary a rule treating every operator alike
		// would cross.
		{"an in expression of two constants", "export function run() { if('x' in {}) {} }\n"},
		{"an instanceof", "export function run() { if(a instanceof Date) {} }\n"},

		// The right-operand short circuit applies only in a boolean position, so as a value it
		// does not make the expression constant.
		{"a variable on the left of an or", "export function run() { if(a || b) {} }\n"},

		// The operator has to match at every level, which is the case ESLint's own comment calls
		// out: `false` is an identity of `&&` and not of `||`. So `a && false` is constant on its
		// own and stops being constant once it is the left operand of an `||`, since nothing then
		// decides the result.
		//
		// I wrote this as a firing case first, reading "nested logical identity" as meaning any
		// depth rather than any depth under the same operator. The fixture failed and was right to.
		{"a constant and under an or", "export function run() { if(a && false || b) {} }\n"},
		{"a constant or under an and", "export function run() { if(a || true && b) {} }\n"},

		// The same operator-matching rule one level down, and these are the cases that measure it.
		// `true && true` is truthy, but `true` is not an identity of `&&`, so the nested expression
		// decides nothing and the outer `||` is not settled by it. Dropping the operator match
		// makes both of these fire.
		//
		// Took a probe to find. The obvious candidates (`a && false || b`) never reach the nested
		// branch at all, because the operand has to be constant before its identity is even asked
		// about, and `a && false` is not.
		{"a nested and that does not decide an outer or", "export function run() { if((true && true) || a) {} }\n"},
		{"a nested or that does not decide an outer and", "export function run() { if((false || false) && a) {} }\n"},
		{"a template with only a substitution", "export function run() { if(`${a}`) {} }\n"},
		{"an array holding a variable, as a value", "export function run() { if([a].length) {} }\n"},

		// Boolean is the only call anyone can reason about, and only with a constant argument.
		{"a Boolean call with a variable", "export function run() { if(Boolean(a)) {} }\n"},
		{"a differently named call", "export function run() { declare function Bool(v: any): boolean; if(Bool(1)) {} }\n"},

		{"a sequence ending in a variable", "export function run() { if((1, a)) {} }\n"},
		{"a negation of a variable", "export function run() { if(!a) {} }\n"},
		{"no condition at all", "export const value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoConstantCondition, constantConditionFile,
				constantConditionDeclarations+testCase.sourceText))
		})
	}
}

// TestNoConstantConditionCheckLoops pins every `checkLoops` value upstream's schema accepts, read
// through the registered decoder and run over the two loops that tell them apart.
//
// `while(true)` separates "all" from "allExceptWhileTrue", and `while(1)` separates both of them
// from "none". The booleans are upstream's older spelling: true reads as "all" and false as "none".
func TestNoConstantConditionCheckLoops(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw            string
		whileTrueFires bool
		whileOneFires  bool
	}{
		{`{}`, false, true},
		{`{"checkLoops": "allExceptWhileTrue"}`, false, true},
		{`{"checkLoops": "all"}`, true, true},
		{`{"checkLoops": "none"}`, false, false},
		{`{"checkLoops": true}`, true, true},
		{`{"checkLoops": false}`, false, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.raw, func(t *testing.T) {
			t.Parallel()
			decoded, err := rule.DecodeOptionsInto[NoConstantConditionOptions]()([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.raw, err)
			}
			for _, loop := range []struct {
				sourceText string
				fires      bool
			}{
				{"while (true) {}", testCase.whileTrueFires},
				{"while (1) {}", testCase.whileOneFires},
			} {
				result := rule_testing.RunWithOptions(t, NoConstantCondition, constantConditionFile, loop.sourceText, decoded)
				if loop.fires {
					rule_testing.ExpectFindings(t, result, "unexpected")
				} else {
					rule_testing.ExpectClean(t, result)
				}
			}
		})
	}

	for _, raw := range []string{`{"checkLoops": "sometimes"}`, `{"checkLoops": "All"}`, `{"checkLoops": 1}`, `{"checkLoops": null}`} {
		t.Run(raw+" is refused", func(t *testing.T) {
			t.Parallel()
			if _, err := rule.DecodeOptionsInto[NoConstantConditionOptions]()([]byte(raw)); err == nil {
				t.Errorf("%s decoded; upstream's schema refuses it", raw)
			}
		})
	}
}
