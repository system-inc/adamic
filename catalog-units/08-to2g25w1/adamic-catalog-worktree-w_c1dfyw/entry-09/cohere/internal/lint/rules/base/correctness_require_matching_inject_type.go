package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// correctnessRequireMatchingInjectTypeText is the rule's message, whose wording lives in
// `policy/messages/correctness-require-matching-inject-type.json`.
var correctnessRequireMatchingInjectTypeText = policy.MessageOf("base/correctness-require-matching-inject-type", "mismatch")

// CorrectnessRequireMatchingInjectType checks that a typed injection decorator resolves to a type the
// decorated parameter accepts.
//
//	valid:   constructor(@Inject<Service>() private s: Service) {}
//	valid:   constructor(@PlainDecorator() private anything: Other) {}   no contract to check
//	valid:   constructor(@Inject<any>() private loose: Other) {}         no concrete contract
//	invalid: constructor(@Inject<Service>() private wrong: Other) {}
//
// The contract lives on the decorator's CALL return type, as a phantom brand:
//
//	type TypedParameterDecorator<TResolved> = ParameterDecorator & {
//	    readonly __resolvedType?: TResolved;
//	}
//
// Any decorator factory returning that branded type opts into the check by doing so, and one
// returning a plain `ParameterDecorator` opts out by having no brand to read. The brand is never
// present at runtime, so the whole rule runs through the type checker.
//
// # The optional brand carries an `undefined` that is an artifact, and stripping it is the rule
//
// `__resolvedType?` is declared optional, so the symbol's type is `TResolved | undefined`. Probed
// here before writing anything: without stripping that member, the brand for `@Inject<Service>()`
// reads as `Service | undefined`, which is NOT assignable to a parameter typed `Service`, so the
// correct case reports and the rule inverts. The original strips it for exactly this reason and
// carries a shared option distinguishing this brand from a sibling whose `undefined` is meaningful.
//
// # `any` and `unknown` mean the decorator carried no contract
//
// A brand resolving to either is a site that has not been migrated to a typed injection key, and
// checking it would noise on every legacy call. The original skips both and says so; reproduced.
//
// # Three parameter shapes, and the third is where the original's own copies drifted
//
// A decorated parameter reaches the rule as a plain parameter, as a parameter property, or as one
// carrying a default. The original consolidates four drifting copies of this resolution into one
// helper and records that one of them returned the assignment pattern itself rather than its left
// side, which read the wrong node's type for defaulted parameters. Our parser puts the default on
// the parameter node rather than wrapping it, so all three shapes are one node kind here and the
// drift it fixed cannot arise. That is a parser difference rather than a decision.
//
// # Cost
//
// One type question per decorator, and only decorators on parameters are asked. A file with no
// decorated parameters costs one kind test per decorator.
var CorrectnessRequireMatchingInjectType = rule.Rule{
	Name: "base/correctness-require-matching-inject-type",

	// Every judgment is an assignability question between two resolved types.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindDecorator: func(node *ast.Node) {
				// Only a decorator on a PARAMETER carries this contract. The original resolves the
				// node whose type should be read and declines when there is none.
				parameterNode := node.Parent
				if parameterNode == nil || parameterNode.Kind != ast.KindParameter {
					return
				}

				decoratorExpression := node.AsDecorator().Expression
				if decoratorExpression == nil {
					return
				}
				decoratorType := ctx.TypeChecker.GetTypeAtLocation(decoratorExpression)
				if decoratorType == nil {
					return
				}

				resolvedType := readInjectResolvedTypeBrand(ctx, decoratorType)
				if resolvedType == nil {
					// Not a branded decorator, so there is no contract to check.
					return
				}

				// A brand of `any` or `unknown` is a decorator that carried no concrete contract.
				if type_checking.IsTypeFlagSet(resolvedType,
					checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
					return
				}

				parameterType := ctx.TypeChecker.GetTypeAtLocation(parameterNode)
				if parameterType == nil {
					return
				}
				if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, resolvedType, parameterType) {
					return
				}

				// The finding points at the PARAMETER rather than at the decorator, which is the
				// original's `node: parameterNode`. The parameter is the half the reader changes.
				//
				// The span starts at the parameter's NAME rather than at its own start. estree's
				// parameter node is the identifier, with the decorators and modifiers as siblings;
				// ours is the whole declaration and its range therefore covers `@Inject<T>() private
				// x: T` entire. Measured against the original on five reporting shapes, whose spans
				// are `wrong: Other`, `s: SubService`, `d: Other = new Other()`, `plain: Other` and
				// `n: Other`, so the span runs from the name through the end of the declaration and
				// a defaulted parameter keeps its default.
				ctx.ReportRange(core.NewTextRange(
					rule.TokenRange(ctx.SourceFile, parameterNode.Name()).Pos(),
					parameterNode.End(),
				), rule.Message{
					Id: correctnessRequireMatchingInjectTypeText.Id,
					Description: correctnessRequireMatchingInjectTypeText.Render(map[string]string{
						"decoratorName": injectDecoratorName(node),
						"resolvedType":  ctx.TypeChecker.TypeToString(resolvedType),
						"parameterType": ctx.TypeChecker.TypeToString(parameterType),
					}),
				})
			},
		}
	},
}

// readInjectResolvedTypeBrand reads the phantom `__resolvedType` brand, with its artifact
// `undefined` removed.
//
// Returns nil when the decorator's type carries no such property, which is how a plain
// `ParameterDecorator` opts out.
//
// The strip is not a tidy-up. Measured before this rule was written: the brand field is declared
// optional, so its type is `T | undefined`, and comparing that against a parameter typed `T` answers
// NOT assignable. Without the strip the rule reports every correct site and is silent on none of
// them, which is the exact inversion of its purpose.
func readInjectResolvedTypeBrand(ctx rule.Context, decoratorType *checker.Type) *checker.Type {
	symbol := ctx.TypeChecker.GetPropertyOfType(decoratorType, "__resolvedType")
	if symbol == nil {
		return nil
	}
	brandType := ctx.TypeChecker.GetTypeOfSymbol(symbol)
	if brandType == nil || !type_checking.IsUnionType(brandType) {
		return brandType
	}

	kept := []*checker.Type{}
	for _, member := range brandType.Types() {
		if !type_checking.IsTypeFlagSet(member, checker.TypeFlagsUndefined) {
			kept = append(kept, member)
		}
	}
	switch len(kept) {
	case 0:
		// A brand of exactly `undefined`, which no real declaration produces. Answering the
		// original union keeps this conservative: the assignability test then decides, rather than
		// this helper inventing a type.
		return brandType
	case 1:
		return kept[0]
	}
	// More than one member survived, so the contract is a genuine union and the original recombines
	// it. The original reaches for a checker method not on the public type and falls back to the
	// unstripped union when it is unavailable; that fallback is the behaviour reproduced here, since
	// a union that still carries `undefined` is what its own fallback path produces too.
	return brandType
}

// injectDecoratorName derives the name printed in the message.
//
// The original resolves a member-expression callee to its property name, so `@Namespace.inject(...)`
// prints `inject`, and falls back to a printable sentinel so a report always has a name in it.
//
// # Why this is not `decorators.CallName`
//
// The shelf has a decorator-name reader and three base rules use it, so not calling it deserves a
// reason rather than looking like an oversight. It answers a narrower question, which its own doc
// states and a probe confirmed: `@Foo()` gives "Foo", while `@ns.Foo()` and a bare `@Foo` both give
// the empty string, because every rule using it keys on a bare identifier and treats a qualified
// name as a different symbol.
//
// This rule does not key on the name at all. It PRINTS it, in a message the reader uses to find the
// decorator, so the empty string is the one answer that helps nobody. Measured against the original:
// `@Namespace.inject<Service>()` prints `@inject`, which a fixture asserts, and `CallName` would
// make that `@`. The sentinel exists for the same reason, so a report always names something.
//
// If a second rule ever wants this wider reading, that is the moment to lift it beside CallName
// rather than widening CallName underneath its three existing callers.
func injectDecoratorName(decorator *ast.Node) string {
	const fallback = "<decorator>"
	expression := decorator.AsDecorator().Expression
	if expression == nil {
		return fallback
	}

	if expression.Kind == ast.KindCallExpression {
		callee := expression.AsCallExpression().Expression
		if callee == nil {
			return fallback
		}
		switch callee.Kind {
		case ast.KindIdentifier:
			return callee.Text()
		case ast.KindPropertyAccessExpression:
			name := callee.AsPropertyAccessExpression().Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				return name.Text()
			}
		}
		return fallback
	}
	if expression.Kind == ast.KindIdentifier {
		return expression.Text()
	}
	return fallback
}
