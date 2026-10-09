package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// correctnessRequireMatchingProviderReturnText is the rule's message, whose wording lives in
// `policy/messages/correctness-require-matching-provider-return.json`.
var correctnessRequireMatchingProviderReturnText = policy.MessageOf("base/correctness-require-matching-provider-return", "mismatch")

// expectedReturnTypeBrand is the phantom property a typed method decorator carries.
//
// Purely a type-level marker: `readonly __expectedReturnType?: TExpectedReturn` is never present at
// run time, so this rule reads it entirely through the checker. A decorator factory returning a
// plain MethodDecorator carries no such property and opts out of the check by that absence.
const expectedReturnTypeBrand = "__expectedReturnType"

// CorrectnessRequireMatchingProviderReturn checks a decorated method's return type against the contract its
// decorator declares.
//
//	valid:   @Provider(token) method(): string {}        where the token expects string | undefined
//	valid:   @Plain() method(): number {}                a plain MethodDecorator declares nothing
//	invalid: @Provider(token) method(): number {}        where the token expects string | undefined
//
// # The contract is a phantom brand on the decorator's call return type
//
// A decorator factory opts in by returning `BrandedMethodDecorator<TExpectedReturn>`, which is
// `MethodDecorator` intersected with `{ readonly __expectedReturnType?: TExpectedReturn }`. The rule
// reads that property off the type of the decorator expression and asks whether the method's
// declared return type is assignable to it. A decorator whose type carries no such property has
// declared no contract, so there is nothing to check and the rule is silent.
//
// The brand is never a runtime value. Everything here happens in the type checker, which is why this
// rule declares NeedsTypeChecker and why a syntax-only port of it would be silent on every input.
//
// # Undefined is NOT stripped here, and that is the difference from its sibling
//
// `api-phi-health` has two brand readers with deliberately different undefined handling.
// `CorrectnessRequireMatchingInjectType` reads `__resolvedType` and strips `undefined`, because the brand
// field is declared optional and the `undefined` in its type is an artifact of that declaration.
// This rule reads `__expectedReturnType` and does NOT strip, because `Provider<T>`'s contract
// intrinsically includes `undefined` through its opt-out path. Stripping it would reject every valid
// provider returning `X | undefined`, which is most of them.
//
// That is the single most reversible decision in this file, so it is stated rather than left to be
// inferred from an absent argument, and a fixture pins it.
//
// # A brand resolving to any or unknown carries no contract
//
// When the type argument is not concrete, assignability answers yes for everything and the check
// would be theatre. The original skips those, and so does this.
//
// # Where this port reads the return type
//
// The original resolves the `MethodDefinition`'s `value`, a FunctionExpression, and reads the return
// type off its call signature. Our parser has no separate value node: a method declaration IS the
// function, so the type at the method declaration carries the call signature directly. Measured, a
// method declaration yields exactly one signature and its return type is what the method declares.
// This is fidelity to the decision rather than to the tree shape, the same way the pagination rule's
// parameter walk collapses.
//
// # No fix, matching the original
//
// A mismatch means reconciling the method with the token's contract, which is a judgment about what
// the method should return rather than a spelling change.
var CorrectnessRequireMatchingProviderReturn = rule.Rule{
	Name: "base/correctness-require-matching-provider-return",

	// Every judgment in this rule is a question for the checker: what type the decorator has, what
	// its brand carries, and whether one type is assignable to another.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindDecorator: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				checkProviderReturn(ctx, node)
			},
		}
	},
}

// checkProviderReturn judges one decorator against the method it sits on.
func checkProviderReturn(ctx rule.Context, node *ast.Node) {
	// A method decorator only. The original resolves the decorator's parent and requires a
	// MethodDefinition; a decorator on a class, a property, an accessor or a parameter reaches this
	// listener too and none of them is this rule's business.
	method := node.Parent
	if method == nil || method.Kind != ast.KindMethodDeclaration {
		return
	}

	decorator := node.AsDecorator()
	if decorator == nil || decorator.Expression == nil {
		return
	}

	// The type of the decorator value being applied. For `@Provider(token)` that is the factory's
	// return type; for a bare `@Foo` it is `Foo` itself. Both are read the same way, which is what
	// makes the brand the whole opt-in mechanism rather than a naming convention.
	decoratorType := ctx.TypeChecker.GetTypeAtLocation(decorator.Expression)
	if decoratorType == nil {
		return
	}

	brandSymbol := ctx.TypeChecker.GetPropertyOfType(decoratorType, expectedReturnTypeBrand)
	if brandSymbol == nil {
		// Not a branded decorator, so it declared no contract and there is nothing to check.
		return
	}
	expectedType := ctx.TypeChecker.GetTypeOfSymbol(brandSymbol)
	if expectedType == nil {
		return
	}

	// A brand whose type argument is any or unknown carries no real contract.
	//
	// SUBSUMED rather than load bearing, and kept because it is the original's own line and it
	// states the intent where a reader looks for it. Every type is assignable to `any` and to
	// `unknown`, so `IsTypeAssignableTo` below already answers yes and this guard cannot change a
	// verdict. Measured over six shapes under both spellings, including `void` and `never`, which
	// are the returns most likely to break an assignability rule: byte identical findings with the
	// guard present and absent.
	//
	// It stays because it is cheaper than a checker call and because it says out loud that such a
	// decorator declared nothing, which is a fact about the contract rather than about assignability
	// happening to be permissive.
	if checker.Type_flags(expectedType)&checker.TypeFlagsAnyOrUnknown != 0 {
		return
	}

	returnType := declaredReturnTypeOfMethod(ctx, method)
	if returnType == nil {
		return
	}

	if ctx.TypeChecker.IsTypeAssignableTo(returnType, expectedType) {
		return
	}

	ctx.ReportNode(method, rule.Message{
		Id: correctnessRequireMatchingProviderReturnText.Id,
		Description: correctnessRequireMatchingProviderReturnText.Render(map[string]string{
			"decoratorName": decoratorNameOf(decorator),
			"expectedType":  ctx.TypeChecker.TypeToString(expectedType),
			"returnType":    ctx.TypeChecker.TypeToString(returnType),
		}),
	})
}

// declaredReturnTypeOfMethod reads the return type off a method's call signature.
//
// The original reaches the FunctionExpression a MethodDefinition carries and takes the first call
// signature. Our parser has no separate value node, so the type at the method declaration carries
// the signature directly; measured, a method declaration yields exactly one.
func declaredReturnTypeOfMethod(ctx rule.Context, method *ast.Node) *checker.Type {
	methodType := ctx.TypeChecker.GetTypeAtLocation(method)
	if methodType == nil {
		return nil
	}
	signatures := ctx.TypeChecker.GetSignaturesOfType(methodType, checker.SignatureKindCall)
	if len(signatures) == 0 {
		return nil
	}
	return ctx.TypeChecker.GetReturnTypeOfSignature(signatures[0])
}

// decoratorNameOf gives a printable name for a decorator, for the message.
//
// The original passes `resolveMemberExpression: true` and a `<decorator>` fallback, so a
// `@Namespace.provider(...)` callee resolves to its property name and a shape with no readable name
// still renders. Both are reproduced.
//
// Every branch checks the node kind before reading text. `Text()` panics on a property access, and
// the walk recovers per FILE rather than per rule, so one such call costs every rule in this package
// every finding in that file while the run still prints a plausible summary.
func decoratorNameOf(decorator *ast.Decorator) string {
	expression := decorator.Expression
	if expression == nil {
		return "<decorator>"
	}
	if expression.Kind == ast.KindCallExpression {
		expression = expression.AsCallExpression().Expression
		if expression == nil {
			return "<decorator>"
		}
	}
	switch expression.Kind {
	case ast.KindIdentifier:
		return expression.Text()
	case ast.KindPropertyAccessExpression:
		// The property name only, matching resolveMemberExpression. The receiver is deliberately
		// not rendered, since the original prints the decorator's own name rather than its path.
		name := expression.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	}
	return "<decorator>"
}
