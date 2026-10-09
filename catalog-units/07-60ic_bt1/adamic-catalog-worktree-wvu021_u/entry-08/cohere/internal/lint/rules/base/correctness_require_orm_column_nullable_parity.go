package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/decorators"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// correctnessRequireOrmColumnNullableParityText is the rule's message, whose wording lives in
// `policy/messages/correctness-require-orm-column-nullable-parity.json`.
var correctnessRequireOrmColumnNullableParityText = policy.MessageOf("base/correctness-require-orm-column-nullable-parity", "columnNullableButTypeNot")

// ormNullableColumnDecorators is the source rule's own `nullableColumnDecorators`.
//
// Deliberately smaller than the `@Orm*` set `orm-column-requires-declare` keys on, and the source
// says why: only `OrmColumn` overlaps, and merging the two would widen this check onto primary-key
// and date columns it does not govern.
var ormNullableColumnDecorators = map[string]struct{}{
	"OrmColumn":       {},
	"TimestampColumn": {},
	"OrmJoinColumn":   {},
}

// ormRelationDecorators is the source layer's `RelationDecorators`.
//
// A property carrying one of these is a hydrated relation, undefined until loaded, and
// `relation-must-be-optional` owns it. This rule steps aside rather than reporting the same property
// twice under two different judgments.
var ormRelationDecorators = map[string]struct{}{
	"OrmManyToOne": {},
	"OrmOneToMany": {},
	"OrmOneToOne":  {},
}

// CorrectnessRequireOrmColumnNullableParity flags a column declared `nullable: true` whose type cannot hold null.
//
//	valid:   @OrmColumn({ nullable: true }) x!: string | null;
//	valid:   @OrmColumn({ nullable: true }) x?: string;
//	valid:   @OrmColumn({ nullable: false }) x!: string;
//	valid:   @OrmColumn({ nullable: true }) @OrmManyToOne(() => B) x!: string;
//	invalid: @OrmColumn({ nullable: true }) x!: string;
//
// This is one of Kirk's own base rules rather than an upstream port, so
// `api-phi-health/libraries/base/code-quality/lint/rules/CorrectnessRequireOrmColumnNullableParityRule.ts` is the
// specification and there is no corpus to import. Its reasoning: a row loaded with a `NULL` hands
// back `null` typed as a non-null value, and the TypeError happens later, somewhere else.
//
// # One direction only, and the other one is deliberate
//
// Column-nullable against type-not-nullable reports. The opposite, a type admitting null against a
// non-nullable column, is permitted, because the pre-insert pattern needs it: `id?: string` for a
// generated key, properties a lifecycle hook fills in. That asymmetry is upstream's and it is the
// difference between this rule and its serializable sibling, which checks both directions.
//
// # What makes a type nullable here
//
// `decorators.IsNullableType`, shared with the other three parity rules rather than re-derived. It
// admits null, undefined, any, unknown and void, recursing through unions. The top types are
// included on purpose: `nullable: false` beside a value typed `any` is a claim nothing checked, so
// the parity rules treat it as unmet rather than as satisfied.
//
// Measured against the real rule, including the rows that read as surprising:
//
//	x?: string                 clean, the question mark makes the type nullable
//	x!: any                    clean, any admits null
//	x!: unknown                clean, same
//	nullable: 'items'          clean, a non-boolean option is not this rule's question
//	["computed"]!: string      clean, the key must be a plain identifier
//	accessor x: string = 'a'   REPORTS, an accessor property is checked like a field
//
// # Cost
//
// Anchored on a decorator, which is rare, and the checker is consulted only after the decorator name
// and its `nullable: true` option have both matched.
var CorrectnessRequireOrmColumnNullableParity = rule.Rule{
	Name: "base/correctness-require-orm-column-nullable-parity",

	// The property's type decides every finding.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindDecorator: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				if _, isColumnDecorator := ormNullableColumnDecorators[decorators.CallName(node)]; !isColumnDecorator {
					return
				}

				// Only an explicit `nullable: true` is this rule's question.
				//
				// The `skip` test is SUBSUMED here and kept anyway, which a mutation sweep will
				// report as a survivor. The shelf sets `skip` only on the non-boolean path, where it
				// also returns `value=false`, so `!declaredNullable` already declines every input
				// that would skip: no shape can reach this line with `skip` true and
				// `declaredNullable` true. It is written because it states the rule's intent about
				// `nullable: 'items'`, which is a list-nullability spelling rather than a false
				// claim, and because that reasoning stops holding the moment the shelf's contract
				// changes. The serializable sibling in this package tests `skip` where it is NOT
				// subsumed, and the two lines should keep reading the same way.
				declaredNullable, skip, found := decorators.BooleanOption(node, "nullable")
				if skip || !found || !declaredNullable {
					return
				}

				owner := node.Parent
				if owner == nil {
					return
				}
				if owner.Kind != ast.KindPropertyDeclaration {
					return
				}

				// A relation property belongs to `relation-must-be-optional`, which asks a narrower
				// question about undefined specifically.
				if decorators.HasDecoratorInSet(owner, ormRelationDecorators) {
					return
				}

				name := owner.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}

				propertyType := ctx.TypeChecker.GetTypeAtLocation(name)
				if propertyType == nil || decorators.IsNullableType(propertyType) {
					return
				}

				ctx.ReportNode(name, buildColumnNullableButTypeNotMessage(
					ctx.TypeChecker.TypeToString(propertyType)))
			},
		}
	},
}

// buildColumnNullableButTypeNotMessage renders the source rule's message verbatim.
//
// The em dash in the message file is the source's own and is reproduced rather than normalised. The
// house copy rule avoids em dashes, but a user-visible message that differs from the rule being
// ported is a real divergence: the two linters print different text for the same finding. Confirmed
// against the message the real rule emits.
func buildColumnNullableButTypeNotMessage(typeText string) rule.Message {
	return rule.Message{
		Id:          correctnessRequireOrmColumnNullableParityText.Id,
		Description: correctnessRequireOrmColumnNullableParityText.Render(map[string]string{"typeText": typeText}),
	}
}
