package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/decorators"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// graphQlOperationDecorators scopes this rule to methods that really are GraphQL operations.
//
// A parameter decorated with `@InjectGraphQlOperationContext` on a method carrying none of these is
// out of scope: the injection is doing something else and there is no operation return type to
// compare against.
var graphQlOperationDecorators = map[string]struct{}{
	"GraphQlQuery":         {},
	"GraphQlMutation":      {},
	"GraphQlFieldResolver": {},
}

// GraphQlOperationContextMatchesReturn holds the generic on an injected operation context to the
// return type of the operation it is injected into.
//
//	valid:   @GraphQlQuery(() => Thing)
//	         async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<Thing>):
//	             Promise<Thing>
//	invalid: the same with GraphQlOperationContext<Other>
//	invalid: the same with the parameter typed as something that is not an operation context
//
// The context object carries the selection set the caller asked for, and its generic is what tells
// the resolver which fields it may read. When that generic and the method's return type disagree,
// the resolver is reading a selection set for a different shape than the one it returns, and
// nothing else notices: both types exist, both compile, and the mismatch surfaces as fields
// silently missing at run time.
//
// # What this rule deliberately does not check, because a sibling already does
//
// Whether the parameter's BASE type is an operation context at all is `inject-type-matches-parameter`'s
// question, answered through the decorator's branded resolved type. This rule reports a wrong base
// type anyway, under its own message, because it has already resolved the type and the alternative
// is silence on the shape that most needs saying something. The original's header comment states
// this split, and the second message id exists to keep the two findings distinguishable.
//
// # The comparison is bidirectional assignability, not identity
//
// Both directions are asserted, which is stricter than either alone and weaker than reference
// identity. A subtype passing one direction is not enough: `Derived` and `Base` are mutually
// unassignable in exactly one direction each, and a resolver returning `Base` while reading a
// `Derived` selection set is the defect this rule exists for. Identity would be too strict, because
// the two types are reached by different routes and a structurally identical alias should pass.
//
// # Both sides are unwrapped to their element type first
//
// `Promise<Readonly<Thing[]>>` and `Thing` describe the same selection set, so the comparison is
// made after stripping the wrappers that do not change it. That unwrap is recursive and its exact
// set is the original's: promise, readonly alias, a nullable union with exactly one non-nullish
// member, and array or readonly array. Anything else is left alone.
//
// # No repair
//
// The original ships none and neither does this. The two ways to satisfy the rule are to change the
// generic or to change the return type, and which one is right depends on what the resolver is
// meant to do, which is not recoverable from the source.
var GraphQlOperationContextMatchesReturn = rule.Rule{
	Name: "base/correctness-require-matching-operation-context",

	// Both halves are type questions: what the parameter's generic resolves to, and what the
	// method's call signature returns.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindDecorator: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if decorators.CallName(node) != "InjectGraphQlOperationContext" {
					return
				}

				// The decorated parameter. Our parser hands the parameter directly as the
				// decorator's parent, so the original's `resolveDecoratedParameterNode` has no
				// counterpart here: it exists to unwrap estree's TSParameterProperty and
				// AssignmentPattern, and probed on all three shapes our parser answers
				// KindParameter for each. That is fidelity to what the rule decides rather than to
				// how the original obtained it.
				//
				// The KIND GUARD is the one part of that helper that survives, and it is
				// load-bearing rather than defensive. A decorator can sit on a class, a method, a
				// property or an accessor, and all four reach this listener with a non-parameter
				// parent, measured. Most are then declined by the enclosing-method walk below,
				// because a class member has no enclosing method, which is why the obvious
				// placements cannot prove the guard. The shape that can is a decorator on a member
				// of a LOCAL CLASS declared inside an operation method: the walk finds the outer
				// method, and without this guard the rule reports a bogus mismatch on its property
				// and a bogus wrongBaseType on its method. Both are silent upstream, where
				// `resolveDecoratedParameterNode` answers null for a non-parameter owner, and both
				// are pinned as fixtures.
				parameterNode := node.Parent
				if parameterNode == nil || parameterNode.Kind != ast.KindParameter {
					return
				}

				methodDeclaration := enclosingMethodOf(parameterNode)
				if methodDeclaration == nil {
					return
				}
				if !decorators.HasDecoratorInSet(methodDeclaration, graphQlOperationDecorators) {
					return
				}

				parameterType := ctx.TypeChecker.GetTypeAtLocation(parameterNode)
				if parameterType == nil {
					return
				}

				// The base type has to BE an operation context. Reported rather than skipped, for
				// the reason in the doc comment.
				symbol := checker.Type_symbol(parameterType)
				if symbol == nil || symbol.Name != "GraphQlOperationContext" {
					ctx.ReportNode(parameterNode, buildWrongBaseTypeMessage(
						ctx.TypeChecker.TypeToString(parameterType)))
					return
				}

				typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, parameterType)
				if len(typeArguments) == 0 {
					return
				}
				parameterGenericType := unwrapToElementType(ctx, typeArguments[0])

				returnType := methodReturnTypeOf(ctx, methodDeclaration)
				if returnType == nil {
					return
				}
				returnType = unwrapToElementType(ctx, returnType)

				if parameterGenericType == nil || returnType == nil {
					return
				}

				// Bidirectional assignability. Each direction alone admits a subtype, which is the
				// case this rule exists to catch.
				if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, parameterGenericType, returnType) &&
					checker.Checker_isTypeAssignableTo(ctx.TypeChecker, returnType, parameterGenericType) {
					return
				}

				ctx.ReportNode(parameterNode, buildMismatchMessage(
					ctx.TypeChecker.TypeToString(parameterGenericType),
					ctx.TypeChecker.TypeToString(returnType)))
			},
		}
	},
}

// enclosingMethodOf walks up from a parameter to the class method that owns it.
//
// Nil when the parameter belongs to a plain function, an arrow, or a constructor.
//
// The loop is EQUIVALENT to a single step today and is written as a loop anyway. Probed: a decorated
// parameter's parent is always the method or constructor directly, with nothing between them, so a
// mutant stopping after one iteration survives every fixture. It is kept because the original is a
// loop and because a future parser change that introduced an intermediate node would silently
// narrow a single step, which is the failure this shape cannot have.
//
// The constructor case is a real narrowing against the original and it costs nothing, which was
// measured rather than argued. estree calls a constructor a MethodDefinition, so the original's walk
// reaches it while this one does not. But a GraphQL operation decorator on a constructor is a PARSE
// ERROR, so the scope gate would decline it anyway: driving the real rule over a seeded constructor
// carrying one reports `Decorators are not valid here` and nothing else, and over a legal
// constructor parameter with no operation decorator it reports nothing. The two versions reach the
// same set through different routes.
func enclosingMethodOf(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindMethodDeclaration {
			return current
		}
	}
	return nil
}

// methodReturnTypeOf reads the return type off a method's call signature.
//
// Through the signature rather than off the annotation, because an unannotated method still has a
// return type the checker inferred and the original reads it the same way.
func methodReturnTypeOf(ctx rule.Context, methodDeclaration *ast.Node) *checker.Type {
	methodType := ctx.TypeChecker.GetTypeAtLocation(methodDeclaration)
	if methodType == nil {
		return nil
	}
	signatures := checker.Checker_getSignaturesOfType(ctx.TypeChecker, methodType, checker.SignatureKindCall)
	if len(signatures) == 0 {
		return nil
	}
	return checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signatures[0])
}

// unwrapToElementType strips the wrappers that do not change a value's selection-set element type.
//
//	Promise<T>              -> T
//	Readonly<T>             -> T   a mapped type, so it is reached through the ALIAS symbol
//	T | null | undefined    -> T   only when exactly one member is non-nullish
//	T[] / ReadonlyArray<T>  -> T
//
// Recursive, so `Promise<Readonly<T[]>>` reaches T. Anything matching none of these is returned
// unchanged. The set is the original's exactly; widening it would make two genuinely different
// selection sets compare equal.
//
// The union arm requires exactly ONE non-nullish member rather than taking the first: a real union
// of two object types is a different selection set from either, and collapsing it would compare the
// wrong thing.
func unwrapToElementType(ctx rule.Context, subject *checker.Type) *checker.Type {
	for depth := 0; subject != nil && depth < unwrapDepthLimit; depth++ {
		symbol := checker.Type_symbol(subject)

		if symbol != nil && symbol.Name == "Promise" {
			if typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, subject); len(typeArguments) > 0 {
				subject = typeArguments[0]
				continue
			}
		}

		// EQUIVALENT for this rule, and kept for fidelity. Measured twice over: the arm is
		// reachable and does unwrap `Readonly<T>` to `T`, and a mutant deleting it still changes no
		// verdict, because `Readonly<T>` and `T` are MUTUALLY ASSIGNABLE while not being identical.
		// The bidirectional comparison below therefore accepts the pair whether or not the wrapper
		// was stripped, so no input can distinguish the two versions.
		//
		// Worth stating precisely, because the obvious fixture does not reach here at all: the
		// checker resolves `Readonly<T[]>` to `ReadonlyArray<T>`, symbol `ReadonlyArray` and no
		// alias, so the array arm handles it. Only `Readonly<T>` over a non-array keeps the alias.
		if aliasSymbol := checker.Type_alias(subject); aliasSymbol != nil {
			if aliasSymbol.Symbol() != nil && aliasSymbol.Symbol().Name == "Readonly" {
				if aliasArguments := aliasSymbol.TypeArguments(); len(aliasArguments) > 0 {
					subject = aliasArguments[0]
					continue
				}
			}
		}

		if subject.IsUnion() {
			var nonNullish []*checker.Type
			for _, member := range subject.AsUnionType().Types() {
				if !type_checking.IsTypeFlagSet(member,
					checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
					nonNullish = append(nonNullish, member)
				}
			}
			if len(nonNullish) == 1 {
				subject = nonNullish[0]
				continue
			}
		}

		if symbol != nil && (symbol.Name == "Array" || symbol.Name == "ReadonlyArray") {
			if typeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, subject); len(typeArguments) > 0 {
				subject = typeArguments[0]
				continue
			}
		}

		return subject
	}
	return subject
}

// unwrapDepthLimit bounds the unwrap loop.
//
// The original recurses without one and relies on types being finite, which they are for every
// shape this rule sees. The bound is here because a rule that loops forever takes the whole run
// down rather than reporting badly, and a type deeper than this is not a selection set anybody
// wrote by hand.
const unwrapDepthLimit = 64

// The rule's messages, whose wording lives in `policy/messages/correctness-require-matching-operation-context.json`.
var (
	graphQlOperationContextMatchesReturnMismatchText      = policy.MessageOf("base/correctness-require-matching-operation-context", "mismatch")
	graphQlOperationContextMatchesReturnWrongBaseTypeText = policy.MessageOf("base/correctness-require-matching-operation-context", "wrongBaseType")
)

func buildMismatchMessage(parameterGeneric string, returnType string) rule.Message {
	return rule.Message{
		Id: graphQlOperationContextMatchesReturnMismatchText.Id,
		Description: graphQlOperationContextMatchesReturnMismatchText.Render(map[string]string{
			"parameterGeneric": parameterGeneric,
			"returnType":       returnType,
		}),
	}
}

func buildWrongBaseTypeMessage(parameterType string) rule.Message {
	return rule.Message{
		Id:          graphQlOperationContextMatchesReturnWrongBaseTypeText.Id,
		Description: graphQlOperationContextMatchesReturnWrongBaseTypeText.Render(map[string]string{"parameterType": parameterType}),
	}
}
