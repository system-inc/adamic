package base

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// correctnessRequireVerifyArrayParityFile is where the fixtures pretend to live.
const correctnessRequireVerifyArrayParityFile = "/repository/source/Entity.ts"

// verifyArrayDecoratorPreamble declares the decorator factories the fixtures apply.
//
// One name from each of the three sets the rule distinguishes, plus an unrecognised custom rule,
// because the third set only exists to handle that last case.
const verifyArrayDecoratorPreamble = "declare function VerifyIsArray(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifyArrayMinimumSize(size: number): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifyIsNotEmpty(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifyIsString(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifyIsNumber(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifyIsOptional(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifyBy(check: unknown): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifySomethingCustom(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function NotAVerifyDecorator(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function OrmColumn(): PropertyDecorator & ParameterDecorator;\n"

// TestVerifyArrayParityFires covers every input the rule must report.
//
// Written rather than imported, since this is our own rule. Each case names which of the three
// decorator sets it exercises, and the whole set was checked against the existing ESLint rule.
func TestVerifyArrayParityFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name     string
		source   string
		reported string
		message  string
	}{
		{
			name:     "a value-level rule on a plain array",
			source:   "class Entity { @VerifyIsString() names: string[]; }",
			reported: "names",
			message:  "Array-typed property 'names' has value-level @Verify rules but no array-level rule.",
		},
		{
			name:     "a value-level rule on a readonly array",
			source:   "class Entity { @VerifyIsString() names: readonly string[]; }",
			reported: "names",
			message:  "Array-typed property 'names' has value-level @Verify rules but no array-level rule.",
		},
		{
			name:     "a value-level rule on a tuple",
			source:   "class Entity { @VerifyIsString() pair: [string, string]; }",
			reported: "pair",
			message:  "Array-typed property 'pair' has value-level @Verify rules but no array-level rule.",
		},
		{
			name:     "a value-level rule on a generic Array",
			source:   "class Entity { @VerifyIsNumber() counts: Array<number>; }",
			reported: "counts",
			message:  "Array-typed property 'counts' has value-level @Verify rules but no array-level rule.",
		},
		{
			name: "a nullable array is still an array",
			// The union case, which neither checker predicate answers for directly. This is the
			// ordinary way to write an optional list, so it is not an edge case.
			source:   "class Entity { @VerifyIsString() names: string[] | null; }",
			reported: "names",
			message:  "Array-typed property 'names' has value-level @Verify rules but no array-level rule.",
		},
		{
			name:     "an optional array is still an array",
			source:   "class Entity { @VerifyIsString() names?: string[]; }",
			reported: "names",
			message:  "Array-typed property 'names' has value-level @Verify rules but no array-level rule.",
		},
		{
			name: "the optional sentinel does not satisfy array parity",
			// VerifyIsOptional is in the known-non-value set, so it neither arms nor suppresses.
			source:   "class Entity { @VerifyIsOptional() @VerifyIsString() names?: string[]; }",
			reported: "names",
			message:  "Array-typed property 'names' has value-level @Verify rules but no array-level rule.",
		},
		{
			name: "a non-Verify decorator does not suppress either",
			// The case that makes the prefix test load-bearing. Without it a foreign decorator
			// falls into the unrecognised-rule arm and suppresses, and the clean case asserting
			// that a foreign decorator alone does not ARM the rule cannot tell the two paths
			// apart. Measured against the real rule: the finding still fires here.
			source:   "class Entity { @NotAVerifyDecorator() @VerifyIsString() names: string[]; }",
			reported: "names",
			message:  "Array-typed property 'names' has value-level @Verify rules but no array-level rule.",
		},
		{
			name: "an ORM decorator beside a validation rule still reports",
			// The realistic spelling of the case above: a persistence decorator sitting beside a
			// validation one is ordinary in this codebase. Measured against the real rule.
			source:   "class Entity { @OrmColumn() @VerifyIsString() names: string[]; }",
			reported: "names",
			message:  "Array-typed property 'names' has value-level @Verify rules but no array-level rule.",
		},
		{
			name: "VerifyBy does NOT suppress",
			// The case this port got wrong first. `VerifyBy` sits in the known-non-value set, so
			// it falls through all three branches: it neither arms the finding nor suppresses it,
			// and a value-level rule beside it still reports. The fixture originally asserted the
			// opposite, on the plausible reasoning that a custom rule might be array-level, and
			// only driving the real ESLint rule settled it. An UNRECOGNISED name does suppress;
			// `VerifyBy` is recognised, and that is the whole difference between the two sets.
			source:   "class Entity { @VerifyBy(() => true) @VerifyIsString() names: string[]; }",
			reported: "names",
			message:  "Array-typed property 'names' has value-level @Verify rules but no array-level rule.",
		},
		{
			name:     "a parameter property is judged like a property",
			source:   "class Entity { constructor(@VerifyIsString() public names: string[]) {} }",
			reported: "names",
			message:  "Array-typed property 'names' has value-level @Verify rules but no array-level rule.",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := verifyArrayDecoratorPreamble + testCase.source
			result := rule_testing.RunTyped(t, CorrectnessRequireVerifyArrayParity, correctnessRequireVerifyArrayParityFile, source)
			rule_testing.ExpectFindings(t, result, "missingArrayRule")

			// The span: the original reports the property's KEY, so a port pointing at the whole
			// declaration would pass the id assertion and send the reader to the wrong column.
			onDisk := strings.TrimSpace(source) + "\n"
			reported := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.reported {
				t.Errorf("finding should point at %q, got %q", testCase.reported, reported)
			}

			// The rendered text, which interpolates the property name.
			if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, testCase.message) {
				t.Errorf("message should begin %q, got %q", testCase.message, got)
			}
		})
	}
}

// TestVerifyArrayParityStaysSilent covers the inputs the rule must decline.
func TestVerifyArrayParityStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		reason string
	}{
		{
			name:   "a value-level rule on a non-array",
			source: "class Entity { @VerifyIsString() name: string; }",
			reason: "The type question is what limits the rule to arrays.",
		},
		{
			name:   "VerifyIsArray satisfies parity",
			source: "class Entity { @VerifyIsArray() @VerifyIsString() names: string[]; }",
			reason: "The canonical array-level rule.",
		},
		{
			name:   "a sizing rule satisfies parity",
			source: "class Entity { @VerifyArrayMinimumSize(1) @VerifyIsString() names: string[]; }",
			reason: "Every array-level rule suppresses, not only VerifyIsArray.",
		},
		{
			name: "VerifyIsNotEmpty satisfies parity",
			// The original places this among the array-level rules deliberately: emptiness is a
			// whole-container check, so on an array it evaluates the array itself.
			source: "class Entity { @VerifyIsNotEmpty() @VerifyIsString() names: string[]; }",
			reason: "Emptiness is a container-level question.",
		},
		{
			name: "an unrecognised custom rule suppresses",
			// The third set exists for exactly this: the custom rule's target cannot be read
			// statically and may be array-level, so reporting here would be a false positive.
			source: "class Entity { @VerifySomethingCustom() @VerifyIsString() names: string[]; }",
			reason: "An unknown Verify* name may itself be an array-level rule.",
		},
		{
			name:   "no value-level rule at all",
			source: "class Entity { @VerifyIsOptional() names: string[]; }",
			reason: "Nothing to be inconsistent with; the finding needs a known value-level rule.",
		},
		{
			name:   "no decorators at all",
			source: "class Entity { names: string[]; }",
			reason: "An unvalidated property is outside the rule's scope.",
		},
		{
			name:   "a non-Verify decorator does not arm the rule",
			source: "class Entity { @NotAVerifyDecorator() names: string[]; }",
			reason: "The prefix test limits the rule to the validation engine's decorators.",
		},
		{
			name:   "a Set is not an array",
			source: "class Entity { @VerifyIsString() names: Set<string>; }",
			reason: "Measured: neither checker predicate answers true for a Set, which is what keeps the rule from reporting every iterable.",
		},
		{
			name:   "an any-typed property is not known to be an array",
			source: "class Entity { @VerifyIsString() names: any; }",
			reason: "Measured false. The rule declines rather than guessing, which is the conservative direction for a rule with no fix.",
		},
		{
			name:   "a bare identifier decorator is not a call",
			source: "class Entity { @VerifyIsString names: string[]; }",
			reason: "The original's verifyDecoratorName guards on CallExpression, and the shared CallName reproduces that by answering empty.",
		},
		{
			name:   "a plain constructor parameter is not a property",
			source: "class Entity { constructor(@VerifyIsString() names: string[]) {} }",
			reason: "Without an accessibility or readonly modifier the parameter declares no property.",
		},
		{
			name:   "a computed key is not read",
			source: "const key = 'names';\nclass Entity { @VerifyIsString() [key]: string[]; }",
			reason: "The original guards on an Identifier key.",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessRequireVerifyArrayParity,
				correctnessRequireVerifyArrayParityFile, verifyArrayDecoratorPreamble+testCase.source))
		})
	}
}

// TestVerifyArrayParityRequiresTheTypedHarness asserts the rule declines a nil checker, with a
// control proving the same input reports through the typed harness.
func TestVerifyArrayParityRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = "class Entity { @VerifyIsString() names: string[]; }"

	rule_testing.ExpectClean(t, rule_testing.Run(t, CorrectnessRequireVerifyArrayParity,
		correctnessRequireVerifyArrayParityFile, verifyArrayDecoratorPreamble+source))

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, CorrectnessRequireVerifyArrayParity,
		correctnessRequireVerifyArrayParityFile, verifyArrayDecoratorPreamble+source), "missingArrayRule")
}
