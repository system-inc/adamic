package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// buildNoUselessCallMessage renders the finding, which names the method being called.
//
// Built per report rather than held as a constant, because the method name is interpolated and
// `rule.Message` has no rendering layer. A fixture asserting the id cannot see anything this format
// string does, so the rendered text is asserted directly in the test.
func buildNoUselessCallMessage(methodName string) rule.Message {
	return rule.Message{
		Id: "unnecessaryCall",
		Description: fmt.Sprintf(
			"This `.%s()` sets a `this` the call would already have had, so the only thing it "+
				"changes is how the line reads. Call the function directly. The indirection also "+
				"breaks the moment somebody adds a method named `%s` to the receiver, and it hides "+
				"the call target from every tool that reads call sites.", methodName, methodName),
	}
}

// NoUselessCall flags a `.call()` or non-variadic `.apply()` whose `this` argument changes nothing.
//
//	valid:   foo.apply(obj, [1, 2]);
//	valid:   obj.foo.apply(otherObj, [1, 2]);
//	valid:   foo.apply(null, args);
//	valid:   var call; foo[call](null, 1, 2);
//	invalid: foo.call(undefined, 1, 2);
//	invalid: obj.foo.apply(obj, [1, 2]);
//	invalid: [].concat.apply([ ], [1, 2]);
//
// # Two shapes, and `.apply()` only in its non-variadic form
//
// `.call(thisArg, a, b)` is watched whenever it has at least one argument. `.apply(thisArg, arr)` is
// watched only when it has exactly two arguments AND the second is an array LITERAL, because that is
// the case where the argument list is known at the call site and the rewrite is mechanical.
// `foo.apply(null, args)` is left alone here: the array is not a literal, so removing `.apply()`
// would need spread syntax, and that is `prefer-spread`'s finding rather than this one. The two
// rules deliberately partition the same syntax by whether the second argument is a literal, and each
// rule's doc names the other.
//
// # `computed === false` is required, and it is NOT what the sibling helper does
//
// Upstream requires the property access to be non-computed, so `foo["call"](undefined, 1, 2)` is
// CLEAN. Measured on the installed rule, both for `call` and for `apply`. This package already has
// `isApplyMemberAccess`, written for `prefer-spread`, and it deliberately ACCEPTS a string-literal
// subscript because `foo['apply'](null, args)` is the same call. Reusing it here unchanged would
// report a whole class of input upstream passes, and no imported fixture would catch it because
// upstream's own clean cases for the computed form use a variable key rather than a literal.
//
// So this rule tests the access shape itself rather than borrowing that helper. It is the one place
// in this port where the neighbouring rule's answer is the wrong answer.
//
// # The receiver comparison, and one stated divergence
//
// For a callee reached through a member access, the redundant `this` is the object it was reached
// through, and "the same object" is decided by comparing tokens, which is upstream's `equalTokens`.
// The shelf already holds that oracle as `hasSameTokens`.
//
// One upstream reporting case does not reproduce, and it is the divergence `prefer-spread` already
// records at its own line: `(obj?.foo).bar.call(obj?.foo, 1, 2)`. ESTree has no node for
// parentheses, so upstream reads both sides as the same tokens and reports. Our parser keeps the
// node, and the two expressions are genuinely different: `(obj?.foo).bar` throws when `obj` is null
// while `obj?.foo.bar` short-circuits. Measured with a signature probe over all nine optional-chain
// shapes in the corpus; that one is the only disagreement, and reporting it would be a false
// finding. Recorded as a case that stays silent rather than deleted, so the next reader sees the
// decision instead of a gap.
//
// # No fix
//
// Upstream ships none, and the rewrite is not mechanical: removing `.call(obj, ...)` from
// `a[i++].foo.call(a[i++], 1)` changes how many times the side effect runs, and the array-literal
// `.apply()` form has to be unpacked into an argument list whose spacing carries intent.
var NoUselessCall = rule.Rule{
	Name: "no-useless-call",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpression := node.AsCallExpression()
				if callExpression == nil || callExpression.Arguments == nil {
					return
				}
				arguments := callExpression.Arguments.Nodes

				callee := ast.SkipParentheses(callExpression.Expression)
				methodName, isMethod := nonComputedMemberName(callee)
				if !isMethod {
					return
				}

				switch methodName {
				case "call":
					// Upstream asks for at least one argument, which is the `this`. `foo.call()`
					// passes no receiver at all and is a different call.
					if len(arguments) < 1 {
						return
					}
				case "apply":
					// Exactly two arguments, the second an array literal. Anything else is either
					// variadic, which is `prefer-spread`'s concern, or not an `.apply()` shape.
					if len(arguments) != 2 ||
						ast.SkipParentheses(arguments[1]).Kind != ast.KindArrayLiteralExpression {
						return
					}
				default:
					return
				}

				// `memberAccessObject` cannot return nil here: `nonComputedMemberName` has already
				// established that the callee is a property access. The guard is against the
				// invariant being widened later rather than against anything reachable today, and
				// `ast.SkipParentheses` dereferences its argument, so a nil reaching it panics and
				// takes the whole file's walk down with it.
				appliedFunction := ast.SkipParentheses(memberAccessObject(callee))
				if appliedFunction == nil {
					return
				}
				thisArgument := ast.SkipParentheses(arguments[0])
				if thisArgument == nil {
					return
				}

				// A callee that is not itself a member access has no implicit receiver, so the only
				// `this` values that change nothing are the empty ones. `void 0` counts alongside
				// `null` and `undefined`, and upstream's corpus writes all three.
				expectedReceiver := memberAccessObject(appliedFunction)
				if expectedReceiver == nil {
					if isNullOrUndefined(thisArgument) {
						ctx.ReportNode(node, buildNoUselessCallMessage(methodName))
					}
					return
				}

				if hasSameTokens(ctx.SourceFile, expectedReceiver, thisArgument) {
					ctx.ReportNode(node, buildNoUselessCallMessage(methodName))
				}
			},
		}
	},
}

// nonComputedMemberName returns the property name of a NON-computed member access.
//
// Upstream's predicate is `callee.property.type === "Identifier" && callee.computed === false`, and
// the computed half is load-bearing rather than incidental: `foo["call"](undefined, 1, 2)` is clean
// on the installed rule. That is the one place this rule must NOT reuse `isApplyMemberAccess` from
// `prefer-spread`, which accepts a string-literal subscript on purpose.
//
// A private identifier answers false. `foo.#call(undefined, 1, 2)` is a different property that
// merely spells the same word, and it is upstream's own clean case.
//
// That check is SUBSUMED rather than load-bearing, and this is a measurement rather than an
// argument: typescript-go keeps the sigil in the identifier's text, so `#call` compares unequal to
// both `call` and `apply` at the switch below and the access is declined there anyway. Probed
// directly, `foo.#call(...)` yields `kind=KindPrivateIdentifier text="#call"` while `this.call(...)`
// yields `kind=KindIdentifier text="call"`. A mutation removing the guard survived all 49 fixtures,
// which is what sent the probe.
//
// Kept rather than deleted, matching what `prefer-spread` decided about the same shape for the same
// reason: the comparison it guards is against a bare name, and nothing at the call site tells a
// reader that the sigil survives into it. Deleting it would make the next reader re-derive this.
func nonComputedMemberName(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return "", false
	}
	propertyAccess := node.AsPropertyAccessExpression()
	if propertyAccess == nil || propertyAccess.Name() == nil {
		return "", false
	}
	name := propertyAccess.Name()
	if name.Kind != ast.KindIdentifier {
		return "", false
	}
	return name.Text(), true
}
