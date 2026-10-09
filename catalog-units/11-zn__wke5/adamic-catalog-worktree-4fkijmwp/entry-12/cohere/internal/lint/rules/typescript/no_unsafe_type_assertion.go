package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnsafeTypeAssertion flags a type assertion that tells the checker something it cannot cohere,
// which is every assertion whose expression is not already assignable to the asserted type.
//
//	valid:   const a = '' as string | number;       widening, so the checker can still check it
//	valid:   const b = 1 as const;
//	invalid: const c = x as string;                 where x is `string | number`
//	invalid: const d = x as any;
//	invalid: function f<T>(x: unknown) { x as T; }  T could be instantiated with anything
//
// An assertion is the one construct that turns type checking OFF for the value it wraps. Widening is
// harmless because the checker keeps checking against the wider type, and narrowing is a promise
// nobody verifies: `event.target as HTMLElement` compiles whatever `target` really is, and the
// mistake surfaces at run time as a property read on null. That is the whole rule, and everything
// below is about telling the reader WHICH kind of unverifiable claim they made.
//
// # Five messages, and the order they are decided in is the rule
//
// Upstream tests in a fixed sequence and returns after each, so an input matching two conditions
// reports the first. Reproduced exactly, because the sequence is the only thing that decides which
// message a reader is shown:
//
// \t1. identical types           silent. `x as string` where x is already `string` asserts nothing.
// \t2. unknown asserted to any   unsafeToAnyTypeAssertion. Called out ahead of the general any test
// \t                             because `unknown` to `any` is the one direction the assignability
// \t                             test below would call safe.
// \t3. any flowing OUT           unsafeOfAnyTypeAssertion. The EXPRESSION is any, or contains an any
// \t                             in a generic position.
// \t4. any flowing IN            unsafeToAnyTypeAssertion. The ASSERTED type is the any one.
// \t5. not assignable            then three sub-cases, below.
//
// Step five splits on whether the asserted type is a type parameter, because a type parameter is
// unsafe for a reason worth naming separately: the caller chooses what it is. An unconstrained one
// gets unsafeToUnconstrainedTypeAssertion; a constrained one the expression satisfies gets
// unsafeTypeAssertionAssignableToConstraint, whose text says the caller could still pick a different
// subtype; anything else gets the general unsafeTypeAssertion.
//
// # Steps three and four are one helper called twice with its arguments swapped
//
// `type_checking.IsUnsafeAssignment` was already on the shelf, ported for the `no-unsafe-*` family,
// and it is the same helper upstream calls here. It walks generic positions rather than testing the
// top-level type, which is what makes `[x] as [string]` report when `x` is any. Calling it in both
// directions is upstream's own trick for asking "is there an any on the left" and then "is there an
// any on the right" with one implementation.
//
// The sender it hands back is what names the type in the message, and it is not always `any`:
// an error type renders as "error typed" instead, which is upstream's way of not saying `any` about
// a type that only exists because something else failed to compile.
//
// # The object literal widening, which is not an optimization
//
// An object literal's type carries its excess-property information, so asking whether
// `{ foo: 'hi', bar: 1 }` is assignable to `{ foo: string }` answers NO for a reason that has
// nothing to do with this rule. Upstream widens an object literal type before the assignability
// test for exactly that, and without it every object literal assertion in the tree would report.
//
// # A checker call that can throw, and what upstream does about it
//
// Upstream wraps its assignability call in a try and returns silently on a throw, citing
// microsoft/TypeScript#62933. Our checker does not signal failure that way, so there is nothing to
// catch and the guard has no counterpart here. Recorded rather than silently dropped, because a
// future reader comparing the two files will notice the missing try and should not have to work out
// whether it was an oversight.
//
// # Cost, and why this rule is registered but NOT enabled
//
// Measured against the ahra tree with the installed 8.67.0 build: 2,846 findings across 672 files,
// one file holding 324 of them. The rule is not auto-fixable and every site is a human decision
// about what the code really guarantees, so switching it on is a project rather than a config line.
// The port is here and correct; whether the tree adopts it is not a porter's call.
var NoUnsafeTypeAssertion = rule.Rule{
	Name: "@typescript-eslint/no-unsafe-type-assertion",

	// Every step is a type question, and two of them are assignability queries.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		checkAssertion := func(node *ast.Node, expression *ast.Node, assertedTypeNode *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}
			if expression == nil || assertedTypeNode == nil {
				return
			}

			expressionType := ctx.TypeChecker.GetTypeAtLocation(expression)
			assertedType := ctx.TypeChecker.GetTypeAtLocation(assertedTypeNode)
			if expressionType == nil || assertedType == nil {
				return
			}

			// Reference identity, matching upstream. `x as string` where x is already `string`
			// asserts nothing, and the checker interns one type per shape per program, so identity
			// is the right question. It is also why the fixtures run one file per program: a
			// batched program would intern one `T` across cases that each declare their own.
			if expressionType == assertedType {
				return
			}

			// `unknown` to `any` is tested ahead of everything else because it is the one direction
			// the assignability test would call safe. Widening to `any` is still turning checking
			// off.
			if type_checking.IsTypeAnyType(assertedType) &&
				type_checking.IsTypeUnknownType(expressionType) {
				ctx.ReportNode(node, buildUnsafeToAnyMessage("`any`"))
				return
			}

			// An any flowing OUT of the expression, including one buried in a generic position.
			if _, sender, isUnsafe := type_checking.IsUnsafeAssignment(expressionType, assertedType,
				ctx.TypeChecker, expression); isUnsafe {
				ctx.ReportNode(node, buildUnsafeOfAnyMessage(anyTypeName(sender)))
				return
			}

			// An any flowing IN from the asserted type. Same helper, arguments swapped.
			if _, sender, isUnsafe := type_checking.IsUnsafeAssignment(assertedType, expressionType,
				ctx.TypeChecker, assertedTypeNode); isUnsafe {
				ctx.ReportNode(node, buildUnsafeToAnyMessage(anyTypeName(sender)))
				return
			}

			// An object literal's type carries its excess properties, so asking whether
			// `{ foo: 'hi', bar: 1 }` is assignable to `{ foo: string }` answers no for a reason
			// that is not this rule's. Widening drops that and asks the question the rule means.
			expressionWidenedType := expressionType
			if isObjectLiteralType(expressionType) {
				expressionWidenedType = checker.Checker_getWidenedType(ctx.TypeChecker, expressionType)
			}

			// Upstream wraps this call in a try for a TypeScript crash it cites by issue number.
			// Our checker does not signal failure by panicking here, so there is nothing to catch.
			if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, expressionWidenedType, assertedType) {
				return
			}

			// A type parameter is unsafe for a reason worth naming: the CALLER picks what it is, so
			// no amount of local reasoning makes the assertion true.
			if type_checking.IsTypeParameter(assertedType) {
				constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, assertedType)
				if constraint == nil {
					ctx.ReportNode(node, buildUnsafeToUnconstrainedMessage(
						ctx.TypeChecker.TypeToString(assertedType)))
					return
				}

				// The expression satisfies the constraint, so the assertion is not absurd; it is
				// merely unprovable, because the caller can still choose a different subtype of
				// that constraint. Upstream gives this its own message and so does this port.
				if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, expressionWidenedType, constraint) {
					ctx.ReportNode(node, buildUnsafeAssignableToConstraintMessage(
						ctx.TypeChecker.TypeToString(assertedType)))
					return
				}
			}

			ctx.ReportNode(node, buildUnsafeTypeAssertionMessage(
				ctx.TypeChecker.TypeToString(assertedType)))
		}

		return rule.Listeners{
			ast.KindAsExpression: func(node *ast.Node) {
				expression := node.AsAsExpression()
				checkAssertion(node, expression.Expression, expression.Type)
			},
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				expression := node.AsTypeAssertion()
				checkAssertion(node, expression.Expression, expression.Type)
			},
		}
	},
}

// isObjectLiteralType answers upstream's `isObjectType(t) && isObjectFlagSet(t, ObjectLiteral)`.
//
// The object-literal flag lives in the object-flags word, which is only meaningful on a type that is
// an object in the first place, so both halves are read.
//
// The flag half is EQUIVALENT here and is kept for fidelity rather than for effect. Measured: a
// mutant dropping it survives every fixture, and probing `getWidenedType` directly says why. On
// `Date`, on a class instance, on a plain object type, on an array and on a function type it returns
// the SAME type object, so widening a non-literal is the identity and the assignability question
// below is unchanged. Only a fresh object literal widens to something different. So the flag saves a
// checker call on the common path and decides nothing, and no fixture can prove it; this comment is
// the record instead.
func isObjectLiteralType(t *checker.Type) bool {
	return type_checking.IsObjectType(t) &&
		checker.Type_objectFlags(t)&checker.ObjectFlagsObjectLiteral != 0
}

// anyTypeName renders the type the message names, reproducing upstream's `getAnyTypeName`.
//
// An error type is called "error typed" rather than `any`, because it is only `any`-shaped as a
// consequence of something else failing to compile, and telling the reader it is `any` would send
// them looking for an annotation that is not there.
func anyTypeName(sender *checker.Type) string {
	if sender != nil && type_checking.IsIntrinsicErrorType(sender) {
		return "error typed"
	}
	return "`any`"
}

func buildUnsafeOfAnyMessage(typeName string) rule.Message {
	return rule.Message{
		Id: "unsafeOfAnyTypeAssertion",
		Description: "Unsafe assertion from " + typeName + " detected: consider using type " +
			"guards or a safer assertion.",
	}
}

func buildUnsafeToAnyMessage(typeName string) rule.Message {
	return rule.Message{
		Id: "unsafeToAnyTypeAssertion",
		Description: "Unsafe assertion to " + typeName + " detected: consider using a more " +
			"specific type to ensure safety.",
	}
}

func buildUnsafeToUnconstrainedMessage(typeName string) rule.Message {
	return rule.Message{
		Id: "unsafeToUnconstrainedTypeAssertion",
		Description: "Unsafe type assertion: '" + typeName + "' could be instantiated with an " +
			"arbitrary type which could be unrelated to the original type.",
	}
}

func buildUnsafeAssignableToConstraintMessage(typeName string) rule.Message {
	return rule.Message{
		Id: "unsafeTypeAssertionAssignableToConstraint",
		Description: "Unsafe type assertion: the original type is assignable to the constraint " +
			"of type '" + typeName + "', but '" + typeName + "' could be instantiated with a " +
			"different subtype of its constraint.",
	}
}

func buildUnsafeTypeAssertionMessage(typeName string) rule.Message {
	return rule.Message{
		Id: "unsafeTypeAssertion",
		Description: "Unsafe type assertion: type '" + typeName + "' is more narrow than the " +
			"original type.",
	}
}
