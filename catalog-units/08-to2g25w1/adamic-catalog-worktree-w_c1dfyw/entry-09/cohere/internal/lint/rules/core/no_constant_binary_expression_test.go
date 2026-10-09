package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const constantBinaryFile = "/repository/source/Compare.ts"

const constantBinaryDeclarations = "declare let a: any, b: any;\n"

func TestNoConstantBinaryExpressionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		// Nullish short circuit: an array can never be nullish, so ?? never chooses.
		{"an array on the left of nullish coalescing", "export const v = [] ?? a;\n", "constantShortCircuit"},
		{"an object on the left of nullish coalescing", "export const v = {} ?? a;\n", "constantShortCircuit"},
		{"a string on the left of nullish coalescing", "export const v = 'x' ?? a;\n", "constantShortCircuit"},
		{"a template on the left of nullish coalescing", "export const v = `x${a}` ?? b;\n", "constantShortCircuit"},
		{"a wrapper call on the left of nullish coalescing", "export const v = String(a) ?? b;\n", "constantShortCircuit"},
		// A nullish literal on the left is reported too, and I first wrote these as clean cases
		// reading "?? exists for null" as meaning null is exempt there. It is not: the operator
		// still never chooses, it just always picks the right side. ESLint passes nonNullish=false
		// at this call site precisely so a nullish left side counts as constant.
		//
		// The nonNullish guard exists for the recursive case instead: `a ?? (b ?? null)` asks
		// whether the inner right side is constantly nullish, and there a null answer means the
		// chain can still produce it.
		{"null on the left of nullish coalescing", "export const v = null ?? a;\n", "constantShortCircuit"},
		{"undefined on the left of nullish coalescing", "export const v = undefined ?? a;\n", "constantShortCircuit"},
		{"an arithmetic expression on the left of nullish coalescing", "export const v = (a + b) ?? a;\n", "constantShortCircuit"},
		// A nested `??` whose right side is never nullish makes the whole chain never nullish, so
		// the outer operator never chooses either. No other fixture reaches the recursive arm.
		//
		// The right side has to be non-nullish rather than nullish, which is the opposite of what
		// I first wrote. `(a ?? null) ?? b` is correctly silent: the chain can still produce null,
		// so the outer `??` really does choose. That asymmetry is what the recursive nonNullish
		// argument exists for, and the silent case for it is in the other table.
		{"a nested nullish chain ending in a non-nullish value", "export const v = (a ?? []) ?? b;\n", "constantShortCircuit"},
		// The strict boolean question is broader than the loose one, which reads backwards: it asks
		// only whether the node can ever be a boolean. A one-element array can coerce to 0 or 1, so
		// the loose comparison is not fixed, while the strict one is, since an array is never a
		// boolean. This pair is what separates the two tests.
		{"a one-element array strictly compared to true", "export const v = [a] === true;\n", "constantBinaryOperand"},

		// Truthiness short circuit.
		{"a constant on the left of an and", "export const v = true && a;\n", "constantShortCircuit"},
		{"a constant on the left of an or", "export const v = [] || a;\n", "constantShortCircuit"},

		// Nullish comparison: neither side can be null, so the answer is fixed.
		{"an array compared to null", "export const v = ([] === null);\n", "constantBinaryOperand"},
		{"an object compared to undefined", "export const v = ({}) === undefined;\n", "constantBinaryOperand"},
		{"null compared to a function", "export const v = null === (() => 1);\n", "constantBinaryOperand"},

		// Boolean comparison: an array is never a boolean, so === true is always false.
		{"an array strictly compared to true", "export const v = [] === true;\n", "constantBinaryOperand"},
		{"a string strictly compared to false", "export const v = 'x' === false;\n", "constantBinaryOperand"},
		{"true strictly compared to an object", "export const v = true === ({});\n", "constantBinaryOperand"},

		// The loose form is a different question: whether the coercion is fixed. An empty array
		// coerces to 0 and a two-element one to a string with a comma, so both are fixed.
		{"an empty array loosely compared to true", "export const v = [] == true;\n", "constantBinaryOperand"},
		{"a two-element array loosely compared to true", "export const v = [a, b] == true;\n", "constantBinaryOperand"},

		// Fresh objects under strict equality compare by identity.
		{"an object compared to a variable", "export const v = a === ({});\n", "alwaysNew"},
		{"an array compared to a variable", "export const v = [] === a;\n", "alwaysNew"},
		{"a regular expression compared to a variable", "export const v = /x/ === a;\n", "alwaysNew"},
		{"a boxed number compared to a variable", "export const v = new Number(1) === a;\n", "alwaysNew"},

		// Both sides fresh, so even loose equality compares identity.
		{"two fresh objects loosely compared", "export const v = ({}) == ({});\n", "bothAlwaysNew"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoConstantBinaryExpression, constantBinaryFile,
				constantBinaryDeclarations+testCase.sourceText), testCase.wantId)
		})
	}
}

func TestNoConstantBinaryExpressionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a variable on the left of nullish coalescing", "export const v = a ?? b;\n"},
		{"a member on the left of nullish coalescing", "export const v = a.b ?? b;\n"},
		// The other half of the recursive arm. The chain can still produce null, so the outer `??`
		// genuinely chooses and must not be reported. Pairs with the firing case above.
		{"a nested nullish chain ending in null", "export const v = (a ?? null) ?? b;\n"},
		{"a call on the left of nullish coalescing", "export function run(f: () => any) { return f() ?? a; }\n"},

		{"a variable on the left of an and", "export const v = a && b;\n"},
		{"a comparison of two variables", "export const v = a === b;\n"},
		{"a variable compared to null", "export const v = a === null;\n"},
		{"a variable compared to true", "export const v = a === true;\n"},

		// The single-element array is the boundary the loose form turns on: `[x]` coerces to
		// whatever x stringifies to, which could be "0" or "1", so it is not fixed.
		{"a one-element array loosely compared to true", "export const v = [a] == true;\n"},

		// A user-defined constructor may return a sentinel, so it is not always new.
		{"a user constructor compared to a variable", "declare class Thing {}\nexport const v = new Thing() === a;\n"},

		// One fresh object under loose equality can still equal a primitive through coercion.
		{"one fresh object loosely compared to a variable", "export const v = ({}) == a;\n"},

		// Relational comparisons are off by default in this tree.
		{"a constant relational comparison", "export const v = 1 < 2;\n"},

		{"an in expression", "export const v = 'x' in ({});\n"},
		{"nothing binary at all", "export const value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoConstantBinaryExpression, constantBinaryFile,
				constantBinaryDeclarations+testCase.sourceText))
		})
	}
}

// The relational arm is off in this tree and on when configured, and both halves are pinned.
//
// Worth its own test because a rule that silently lacks an option is indistinguishable from one
// whose option is off, and the difference surfaces only when someone turns it on. That is the exact
// shape use-isnan shipped with.
func TestNoConstantBinaryExpressionRelationalArm(t *testing.T) {
	t.Parallel()

	source := constantBinaryDeclarations + "export const v = 1 < 2;\n"

	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoConstantBinaryExpression, constantBinaryFile, source,
		NoConstantBinaryExpressionOptions{CheckRelationalComparisons: false}))

	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoConstantBinaryExpression, constantBinaryFile, source,
		NoConstantBinaryExpressionOptions{CheckRelationalComparisons: true}),
		"constantRelationalComparison")

	// A variable operand is not a literal, so the arm stays silent even when enabled.
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoConstantBinaryExpression, constantBinaryFile,
		constantBinaryDeclarations+"export const v = a < 2;\n",
		NoConstantBinaryExpressionOptions{CheckRelationalComparisons: true}))
}
