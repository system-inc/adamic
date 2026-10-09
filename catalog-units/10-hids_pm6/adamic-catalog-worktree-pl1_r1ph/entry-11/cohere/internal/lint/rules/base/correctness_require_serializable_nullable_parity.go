package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/decorators"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// CorrectnessRequireSerializableNullableParity flags a `@SerializableField` whose `optional` disagrees with its type.
//
//	valid:   @SerializableField({ optional: true }) x?: string;
//	valid:   @SerializableField({ optional: false }) x!: string;
//	valid:   @SerializableField({ optional: true, defaultValue: 'a' }) x!: string;
//	invalid: @SerializableField({ optional: true }) x!: string;
//	invalid: @SerializableField({ optional: false }) x?: string;
//
// One of Kirk's own base rules rather than an upstream port, so
// `api-phi-health/libraries/base/code-quality/lint/rules/CorrectnessRequireSerializableNullableParityRule.ts` is the
// specification. Both failure directions are real: deserializing input that omits an optional field
// produces `undefined` typed as non-null, and a non-optional field whose type admits undefined
// passes validation for missing values that downstream code treats as required.
//
// # BOTH directions, unlike its ORM sibling
//
// `orm-column-nullable-parity` checks one direction only, because the pre-insert pattern needs the
// other. This one checks both, and the two rules sit beside each other in this package with that
// difference being the whole distinction between them. Worth stating because they otherwise read as
// the same rule with different decorator names.
//
// # An ABSENT `optional` is read as false, which is the row most likely to surprise
//
// The source calls `getBooleanOption`, whose absent case returns `false` rather than a skip, and
// then guards on `=== null` which absent does not satisfy. So a bare `@SerializableField()` on a
// nullable property REPORTS:
//
//	@SerializableField() x?: string      REPORTS, typeOptionalButDecoratorNot
//	@SerializableField() x!: string      clean
//
// Measured against the real rule. Reading the source alone suggests an absent option is skipped, and
// it is not: absent and explicitly-false are the same claim here, and the shelf's three-value return
// is what lets this file express that deliberately rather than by accident.
//
// A NON-BOOLEAN value is different and does skip: `optional: 'items'` is not a claim this rule can
// check, so it declines rather than reading it as false.
//
// # `defaultValue` suppresses the whole check
//
// A field hydrated to a known value after deserialize is correctly typed non-optional even when the
// serialized form may omit it, so the presence of the key alone is enough regardless of its value.
//
// # What counts as nullable
//
// `decorators.IsNullableType`, shared with the other parity rules. It admits any and unknown, which
// is why `optional: false` beside a value typed `any` reports: the claim was never checked.
//
// # Cost
//
// Anchored on a decorator, and the checker is consulted only after the name, the `defaultValue`
// absence and the `optional` value have all been settled syntactically.
var CorrectnessRequireSerializableNullableParity = rule.Rule{
	Name: "base/correctness-require-serializable-nullable-parity",

	// The property's type decides every finding.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindDecorator: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				if decorators.CallName(node) != "SerializableField" {
					return
				}

				// A default value hydrates the field after deserialize, so the parity question does
				// not arise. Presence is the test, whatever the value is.
				if _, _, hasDefault := decorators.BooleanOption(node, "defaultValue"); hasDefault {
					return
				}

				// A non-boolean `optional` is not this rule's question. An ABSENT one is: it means
				// the same as an explicit false, which is why `found` is not tested here and is
				// tested in the ORM sibling.
				declaredOptional, skip, _ := decorators.BooleanOption(node, "optional")
				if skip {
					return
				}

				owner := node.Parent
				if owner == nil || owner.Kind != ast.KindPropertyDeclaration {
					return
				}
				name := owner.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}

				propertyType := ctx.TypeChecker.GetTypeAtLocation(name)
				if propertyType == nil {
					return
				}
				typeIsNullable := decorators.IsNullableType(propertyType)
				typeText := ctx.TypeChecker.TypeToString(propertyType)

				switch {
				case declaredOptional && !typeIsNullable:
					ctx.ReportNode(name, buildDecoratorOptionalButTypeNotMessage(typeText))
				case !declaredOptional && typeIsNullable:
					ctx.ReportNode(name, buildTypeOptionalButDecoratorNotMessage(typeText))
				}
			},
		}
	},
}

// The rule's messages, whose wording lives in `policy/messages/correctness-require-serializable-nullable-parity.json`.
var (
	correctnessRequireSerializableNullableParityDecoratorOptionalButTypeNotText = policy.MessageOf("base/correctness-require-serializable-nullable-parity", "decoratorOptionalButTypeNot")
	correctnessRequireSerializableNullableParityTypeOptionalButDecoratorNotText = policy.MessageOf("base/correctness-require-serializable-nullable-parity", "typeOptionalButDecoratorNot")
)

func buildDecoratorOptionalButTypeNotMessage(typeText string) rule.Message {
	return rule.Message{
		Id:          correctnessRequireSerializableNullableParityDecoratorOptionalButTypeNotText.Id,
		Description: correctnessRequireSerializableNullableParityDecoratorOptionalButTypeNotText.Render(map[string]string{"typeText": typeText}),
	}
}

func buildTypeOptionalButDecoratorNotMessage(typeText string) rule.Message {
	return rule.Message{
		Id:          correctnessRequireSerializableNullableParityTypeOptionalButDecoratorNotText.Id,
		Description: correctnessRequireSerializableNullableParityTypeOptionalButDecoratorNotText.Render(map[string]string{"typeText": typeText}),
	}
}
