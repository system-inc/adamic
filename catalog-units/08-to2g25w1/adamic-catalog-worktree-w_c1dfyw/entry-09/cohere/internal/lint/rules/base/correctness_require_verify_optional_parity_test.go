package base

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// correctnessRequireVerifyOptionalParityFile is where the fixtures pretend to live.
const correctnessRequireVerifyOptionalParityFile = "/repository/source/Entity.ts"

// verifyDecoratorPreamble declares the decorator factories the fixtures apply.
//
// Ambient declarations rather than real implementations, because the rule reads decorator NAMES and
// the property's type, and neither depends on what the factory returns. Written once here so a
// fixture's source is only the shape under test.
const verifyDecoratorPreamble = "declare function VerifyIsOptional(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifyIsString(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifyIsNumber(): PropertyDecorator & ParameterDecorator;\n" +
	"declare function VerifyBy(check: unknown): PropertyDecorator & ParameterDecorator;\n" +
	"declare function NotAVerifyDecorator(): PropertyDecorator & ParameterDecorator;\n" +
	"declare const namespaced: { Verify(): PropertyDecorator };\n"

// TestVerifyOptionalParityFires covers every input the rule must report.
//
// These fixtures are written rather than imported, because this is our own rule and there is no
// upstream corpus. That inverts the usual risk: an invented fixture encodes the same belief as the
// port, so each case below states which half of the rule it exercises, and the whole set was
// checked against the existing ESLint rule running over api-phi-health.
func TestVerifyOptionalParityFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name      string
		source    string
		messageId string
		// reported is the source text the finding should cover.
		reported string
		// message is the rendered prefix, which carries the interpolated type text. Asserted
		// because the two messages differ only in wording once the id is satisfied: a mutation
		// swapping one format string for the other kept every id assertion green.
		message string
	}{
		{
			name:      "the sentinel with a plain type",
			source:    "class Entity { @VerifyIsOptional() name: string; }",
			messageId: "optionalButTypeNot",
			reported:  "name",
			message:   "@VerifyIsOptional() is present but type 'string' does not include null/undefined.",
		},
		{
			name:      "the sentinel beside another rule with a plain type",
			source:    "class Entity { @VerifyIsString() @VerifyIsOptional() name: string; }",
			messageId: "optionalButTypeNot",
			reported:  "name",
			message:   "@VerifyIsOptional() is present but type 'string' does not include null/undefined.",
		},
		{
			name:      "a nullable type with a validation rule and no sentinel",
			source:    "class Entity { @VerifyIsString() name?: string; }",
			messageId: "typeOptionalButNoVerify",
			reported:  "name",
			message:   "Type 'string | undefined' is nullable but @VerifyIsOptional() is missing.",
		},
		{
			name:      "a union with null counts as nullable",
			source:    "class Entity { @VerifyIsString() name: string | null; }",
			messageId: "typeOptionalButNoVerify",
			reported:  "name",
			message:   "Type 'string | null' is nullable but @VerifyIsOptional() is missing.",
		},
		{
			name:      "a union with undefined counts as nullable",
			source:    "class Entity { @VerifyIsString() name: string | undefined; }",
			messageId: "typeOptionalButNoVerify",
			reported:  "name",
		},
		{
			name: "any counts as nullable, because it erases the question",
			// This is the widest arm of the shared mask and the one a narrower predicate would miss.
			source:    "class Entity { @VerifyIsString() name: any; }",
			messageId: "typeOptionalButNoVerify",
			reported:  "name",
			message:   "Type 'any' is nullable but @VerifyIsOptional() is missing.",
		},
		{
			name:      "unknown counts as nullable for the same reason",
			source:    "class Entity { @VerifyIsString() name: unknown; }",
			messageId: "typeOptionalButNoVerify",
			reported:  "name",
		},
		{
			name:      "a parameter property is judged like a property",
			source:    "class Entity { constructor(@VerifyIsOptional() public name: string) {} }",
			messageId: "optionalButTypeNot",
			reported:  "name",
		},
		{
			name:      "a readonly parameter property is judged too",
			source:    "class Entity { constructor(@VerifyIsString() readonly name?: string) {} }",
			messageId: "typeOptionalButNoVerify",
			reported:  "name",
		},
		{
			name: "a custom VerifyBy rule counts as another rule",
			// VerifyBy is a validation rule like any other for this judgment, so it satisfies the
			// "at least one other" test that the second half needs.
			source:    "class Entity { @VerifyBy(() => true) name?: string; }",
			messageId: "typeOptionalButNoVerify",
			reported:  "name",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := verifyDecoratorPreamble + testCase.source
			result := rule_testing.RunTyped(t, CorrectnessRequireVerifyOptionalParity, correctnessRequireVerifyOptionalParityFile, source)
			rule_testing.ExpectFindings(t, result, testCase.messageId)

			// The span. Upstream reports the property's KEY rather than the whole property, so a
			// port pointing at the declaration would pass every message-id assertion while sending
			// the reader to the wrong column.
			//
			// `RunTyped` writes the fixture as TrimSpace(contents)+"\n", so the literal is
			// transformed the same way before slicing rather than sliced directly.
			onDisk := strings.TrimSpace(source) + "\n"
			reported := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.reported {
				t.Errorf("finding should point at %q, got %q", testCase.reported, reported)
			}

			// The rendered text, asserted as an exact prefix rather than with a substring test:
			// the type name is interpolated, and a `strings.Contains` cannot see a doubled or
			// misplaced interpolation.
			if testCase.message != "" {
				if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, testCase.message) {
					t.Errorf("message should begin %q, got %q", testCase.message, got)
				}
			}
		})
	}
}

// TestVerifyOptionalParityStaysSilent covers the inputs the rule must decline.
//
// These are the false positives the rule has to avoid, and several encode a distinction nothing
// else would suggest.
func TestVerifyOptionalParityStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		reason string
	}{
		{
			name:   "no decorators at all",
			source: "class Entity { name?: string; }",
			reason: "A property the validation engine never looks at is outside the rule's scope.",
		},
		{
			name:   "a non-Verify decorator only",
			source: "class Entity { @NotAVerifyDecorator() name?: string; }",
			reason: "The prefix test is what limits the rule to the validation engine's decorators.",
		},
		{
			name:   "the sentinel with an optional type",
			source: "class Entity { @VerifyIsOptional() name?: string; }",
			reason: "Both halves agree, which is the shape the sentinel exists for.",
		},
		{
			name:   "the sentinel with an explicit undefined union",
			source: "class Entity { @VerifyIsOptional() name: string | undefined; }",
			reason: "The same agreement written the other way; the checker collapses them.",
		},
		{
			name:   "the sentinel with a null union",
			source: "class Entity { @VerifyIsOptional() name: string | null; }",
			reason: "Null satisfies the shared mask alongside undefined.",
		},
		{
			name:   "a validation rule with a plain type",
			source: "class Entity { @VerifyIsString() name: string; }",
			reason: "Both halves agree in the other direction.",
		},
		{
			name: "a nullable type carrying ONLY the sentinel",
			// The case that makes the "at least one other rule" test load-bearing: dropping it
			// reports here, and this is exactly what a correctly-annotated optional field looks
			// like.
			source: "class Entity { @VerifyIsOptional() name: string | null; }",
			reason: "There is no other rule to be inconsistent with.",
		},
		{
			name:   "a bare identifier decorator is not a call",
			source: "class Entity { @VerifyIsString name?: string; }",
			reason: "The original guards on CallExpression, so a decorator applied without parentheses is not matched. The shared CallName answers empty for it, which reproduces that.",
		},
		{
			name:   "a qualified decorator name is a different symbol",
			source: "class Entity { @namespaced.Verify() name?: string; }",
			reason: "A member-expression callee is not resolved, matching the original's default.",
		},
		{
			name:   "a plain constructor parameter is not a property",
			source: "class Entity { constructor(@VerifyIsString() name?: string) {} }",
			reason: "Without an accessibility or readonly modifier the parameter declares no property, so there is nothing for the engine to validate. This is the guard the estree original gets for free from its TSParameterProperty node kind.",
		},
		{
			name:   "a computed key is not read",
			source: "const key = 'name';\nclass Entity { @VerifyIsString() [key]?: string; }",
			reason: "The original guards on an Identifier key. A computed name is not knowable statically.",
		},
		{
			name:   "a method is not a property",
			source: "class Entity { @VerifyIsString() name(): string { return ''; } }",
			reason: "The rule anchors on property declarations and parameters, not on methods.",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessRequireVerifyOptionalParity,
				correctnessRequireVerifyOptionalParityFile, verifyDecoratorPreamble+testCase.source))
		})
	}
}

// TestVerifyOptionalParityRequiresTheTypedHarness asserts the rule declines a nil checker.
//
// A typed rule handed the plain harness goes silent rather than crashing, which makes every
// StaysSilent case pass vacuously and every Fires case look like a rule defect. The control below
// is what separates "correctly guarded" from "cannot fire at all".
func TestVerifyOptionalParityRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = "class Entity { @VerifyIsOptional() name: string; }"

	rule_testing.ExpectClean(t, rule_testing.Run(t, CorrectnessRequireVerifyOptionalParity,
		correctnessRequireVerifyOptionalParityFile, verifyDecoratorPreamble+source))

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, CorrectnessRequireVerifyOptionalParity,
		correctnessRequireVerifyOptionalParityFile, verifyDecoratorPreamble+source), "optionalButTypeNot")
}
