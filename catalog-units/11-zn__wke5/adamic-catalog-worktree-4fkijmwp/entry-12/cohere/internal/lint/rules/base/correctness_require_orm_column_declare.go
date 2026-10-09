package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// correctnessRequireOrmColumnDeclareText is the rule's message, whose wording lives in
// `policy/messages/correctness-require-orm-column-declare.json`.
var correctnessRequireOrmColumnDeclareText = policy.MessageOf("base/correctness-require-orm-column-declare", "missingDeclare")

// declareRequiringColumnDecorators is the source rule's set of the same name.
//
// Eight entries, and the source keeps it rule-local rather than sharing it with the column set that
// `orm-column-nullable-parity` reads, which has three. Only `OrmColumn` overlaps, and the comment
// there says outright that they are kept apart so the two are never wrong-merged. That separation is
// reproduced: a shared set would be smaller and would silently change what both rules mean.
var declareRequiringColumnDecorators = map[string]bool{
	"OrmColumn":            true,
	"OrmPrimaryKey":        true,
	"OrmPrimaryAutoColumn": true,
	"OrmCreateDateColumn":  true,
	"OrmUpdateDateColumn":  true,
	"OrmColumnIndex":       true,
	"OrmColumnUnique":      true,
	"OrmColumnUniqueIndex": true,
}

// CorrectnessRequireOrmColumnDeclare requires `declare` on a property carrying an ORM column decorator.
//
//	valid:   @OrmColumn() declare name: string;
//	valid:   @OrmManyToOne() profile?: Profile;   a relation decorator is a different rule
//	invalid: @OrmColumn() name: string;
//	invalid: @OrmPrimaryKey() id: number;
//
// Under `useDefineForClassFields`, a class field definition is emitted at construction and overwrites
// whatever the decorator wrote into the property's metadata. Marking the field `declare` tells
// TypeScript the property is set elsewhere and skips that emit, which is what preserves the
// decorator's behavior at runtime. So this is a correctness rule wearing the clothes of a style one.
//
// # It accepts a decorator written WITHOUT parentheses, unlike its relation sibling
//
// The two ORM rules in this package use different helpers from the source library and the difference
// is real rather than incidental. `relation-must-be-optional` goes through `hasDecoratorInSet`, which
// requires a call expression, so a bare `@OrmManyToOne` is silent there. This one goes through
// `getDecoratorName`, which also resolves a bare identifier expression, so `@OrmColumn` with no
// parentheses reports here. Both measured against the source rules on the same shape.
//
// A namespaced `@Orm.Column()` is silent, because the source calls `getDecoratorName` with its
// member-expression resolution left at the default of false. That option exists and two other rules
// in the library pass true; this one does not, so the port does not either.
//
// # A key that is not a plain identifier is reported as `<computed>`, and its span needs unwrapping
//
// The source names the property in its message and falls back to the literal string `<computed>` when
// the key is not an Identifier. Both a computed key and a private name take that path, because a
// private name is its own node type in the source's tree rather than an identifier. Measured: both
// report, and both render as `<computed>` in the message while the finding still points at the key.
//
// The SPAN needs one adjustment our tree makes necessary. A computed key is one node here, a
// KindComputedPropertyName carrying the brackets, and two in the source's tree, where `key` is the
// inner expression. Reporting our node directly would underline `['computed']` where the source
// underlines `'computed'`, so the inner expression is what the finding anchors on. That was the one
// row of nineteen where the port and the source disagreed, and the fixture caught it.
//
// # Cost
//
// One listener on a property declaration, no checker, no program, and the decorator scan stops at the
// first match, which is what makes the message name the triggering decorator rather than the last one.
var CorrectnessRequireOrmColumnDeclare = rule.Rule{
	Name: "base/correctness-require-orm-column-declare",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				modifiers := node.Modifiers()
				if modifiers == nil {
					return
				}
				// `declare` already present, so the field emit is already skipped.
				if ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient) {
					return
				}

				triggeringDecorator := ""
				for _, modifier := range modifiers.Nodes {
					if modifier.Kind != ast.KindDecorator {
						continue
					}
					name := ormColumnDecoratorName(modifier)
					if name != "" && declareRequiringColumnDecorators[name] {
						triggeringDecorator = name
						break
					}
				}
				if triggeringDecorator == "" {
					return
				}

				key := node.AsPropertyDeclaration().Name()
				if key == nil {
					return
				}
				// A computed key is one node here and two in the source's tree: its `key` is the
				// inner expression, ours wraps that in a KindComputedPropertyName carrying the
				// brackets. Unwrapping puts the finding on the same span the source reports, which
				// is `'computed'` rather than `['computed']`. Measured on both spellings rather than
				// inferred, and it is the one row of nineteen where the two disagreed.
				if key.Kind == ast.KindComputedPropertyName {
					if inner := key.AsComputedPropertyName().Expression; inner != nil {
						key = inner
					}
				}
				// The source's `node.key.type === 'Identifier' ? node.key.name : '<computed>'`. A
				// private name takes the fallback too, because it is not an Identifier there.
				propertyName := "<computed>"
				if key.Kind == ast.KindIdentifier {
					propertyName = key.AsIdentifier().Text
				}

				ctx.ReportNode(key, rule.Message{
					Id: correctnessRequireOrmColumnDeclareText.Id,
					Description: correctnessRequireOrmColumnDeclareText.Render(map[string]string{
						"propertyName":  propertyName,
						"decoratorName": triggeringDecorator,
					}),
				})
			},
		}
	},
}

// ormColumnDecoratorName is the source library's `getDecoratorName` at its DEFAULT options.
//
// Two shapes resolve: a call with a bare identifier callee, and a bare identifier expression. A
// member-expression callee does not, because `resolveMemberExpression` defaults to false and this
// rule does not override it. Returning the empty string stands in for the source's null fallback,
// which is what this caller passes.
//
// # Why this is not decorators.CallName from the shelf
//
// The shelf has a helper for reading a decorator's name and this rule cannot use it. Measured on the
// same three shapes rather than read off its doc comment: `CallName` answers "" for a bare
// `@OrmColumn` while this answers "OrmColumn", and the source rule REPORTS on that shape. Using the
// shelf here would make the rule silent on a decorator written without parentheses.
//
// The difference is real rather than an oversight in either place. The source library keeps two
// helpers on purpose, and its comment says which rules used which: the call-only variant for the
// predicate rules, the wider one for the rules that print a name. The sibling rule in this package
// goes through the shelf because its source helper is the call-only one. `CallName`'s own comment
// says every rule using it keys on a bare identifier, which was true when it was written and is the
// kind of claim that stops being true as rules arrive; this is the one that does not.
//
// The name is prefixed rather than bare because this package is one namespace shared with every
// other base rule, and a sibling already has a `decoratorNameOf`. Two helpers one rename apart is
// how a package-level collision arrives.
func ormColumnDecoratorName(decorator *ast.Node) string {
	expression := decorator.AsDecorator().Expression
	if expression == nil {
		return ""
	}
	if expression.Kind == ast.KindCallExpression {
		callee := expression.AsCallExpression().Expression
		if callee != nil && callee.Kind == ast.KindIdentifier {
			return callee.AsIdentifier().Text
		}
		return ""
	}
	if expression.Kind == ast.KindIdentifier {
		return expression.AsIdentifier().Text
	}
	return ""
}
