package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/decorators"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// correctnessRequireOptionalRelationText is the rule's message, whose wording lives in
// `policy/messages/correctness-require-optional-relation.json`.
var correctnessRequireOptionalRelationText = policy.MessageOf("base/correctness-require-optional-relation", "mustBeOptional")

// relationDecorators is the source rule's `RelationDecorators` set.
//
// Kept as its own value rather than inlined because the source library keeps it that way and says
// why: the nullability-parity rules read the same set to skip properties this rule governs instead.
// Those rules are separate ports, and two copies of a three-name set is exactly how the source
// repository's five drifting copies of `getDecoratorName` came about.
var relationDecorators = map[string]struct{}{
	"OrmManyToOne": {},
	"OrmOneToMany": {},
	"OrmOneToOne":  {},
}

// CorrectnessRequireOptionalRelation requires an ORM relation property to admit `undefined`.
//
//	valid:   @OrmManyToOne() profile?: Profile;
//	valid:   @OrmManyToOne() profile: Profile | undefined;
//	valid:   @OrmManyToOne() profile: any;
//	invalid: @OrmManyToOne() profile: Profile;
//	invalid: @OrmManyToOne() profile: Profile | null;
//
// The ORM does not eagerly load a relation, so reading the property before it has been hydrated
// gives `undefined`. A type that does not say so hands every call site false confidence and a
// TypeError at runtime.
//
// # `| null` alone does not satisfy, and that is the whole point of the narrow test
//
// The source library keeps two masks and this rule reads the narrower one: undefined, any and
// unknown count, while null and void do not. The reason is stated there and measured here: the
// runtime returns `undefined` for an unloaded relation, never `null`, so a property typed
// `Profile | null` is still lying about its load state.
//
// Measured against the source rule across sixteen shapes rather than inferred. `| null` reports;
// `| null | undefined` does not; `any` and `unknown` are silent because they erase the distinction
// rather than answer it; and `void` REPORTS, which is the one that reads as a surprise and is
// deliberate, since void is absent from the narrow mask.
//
// # The decorator must be CALLED, and that is a real discrimination
//
// The source's `hasDecoratorInSet` requires a call expression with an identifier callee, so
// `@OrmManyToOne()` matches and a bare `@OrmManyToOne` does not. Measured: the bare form is silent.
// That differs from the sibling rule for `declare`, whose helper accepts the bare form too, and the
// two are deliberately not merged. This rule therefore uses the shelf helper and that one does not.
//
// # Cost
//
// One listener on a property declaration, and the type is asked for only after a relation decorator
// is found, which is a rare shape.
var CorrectnessRequireOptionalRelation = rule.Rule{
	Name: "base/correctness-require-optional-relation",

	// The finding is decided by whether the property's type admits undefined, so there is no
	// syntactic subset of this rule.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				// The shelf's HasDecoratorInSet, which was measured to agree with the source's
				// `hasDecoratorInSet` exactly: both require a call expression with a bare identifier
				// callee, so `@OrmManyToOne()` matches while `@OrmManyToOne` and `@Orm.ManyToOne()`
				// do not. The sibling rule in this package deliberately does NOT use it, for the
				// reason recorded there.
				if !decorators.HasDecoratorInSet(node, relationDecorators) {
					return
				}

				// The source rule requires an Identifier key, so a computed key is skipped. Measured
				// silent rather than assumed, and it matters: the message interpolates the name and
				// a computed key has none to print.
				name := node.AsPropertyDeclaration().Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}

				propertyType := ctx.TypeChecker.GetTypeAtLocation(name)
				if propertyType == nil {
					return
				}
				if typeIncludesUndefined(propertyType) {
					return
				}

				ctx.ReportNode(name, rule.Message{
					Id: correctnessRequireOptionalRelationText.Id,
					Description: correctnessRequireOptionalRelationText.Render(map[string]string{
						"propertyName": name.AsIdentifier().Text,
						"propertyType": ctx.TypeChecker.TypeToString(propertyType),
					}),
				})
			},
		}
	},
}

// typeIncludesUndefined is the source library's `typeIncludesUndefined`.
//
// The mask is deliberately narrower than its `isNullableType` sibling: undefined, any and unknown
// count, while null and void do not. Recursive across a union, because a union admits undefined when
// any constituent does.
func typeIncludesUndefined(subjectType *checker.Type) bool {
	if type_checking.IsTypeFlagSet(subjectType,
		checker.TypeFlagsUndefined|checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		return true
	}
	if type_checking.IsUnionType(subjectType) {
		for part := range type_checking.UnionTypePartsSeq(subjectType) {
			if typeIncludesUndefined(part) {
				return true
			}
		}
	}
	return false
}
