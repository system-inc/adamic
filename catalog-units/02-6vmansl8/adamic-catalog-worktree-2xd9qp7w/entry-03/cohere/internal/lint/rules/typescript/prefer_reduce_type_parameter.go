package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferReduceTypeParameter flags a type assertion on the initial value of `Array#reduce` where a
// type argument on `reduce` itself would say the same thing.
//
//	valid:   [1, 2, 3].reduce((sum, num) => sum + num, 0);
//	valid:   [1, 2, 3].reduce<number[]>((a, s) => a.concat(s * 2), []);
//	valid:   ['a', 'b'].reduce((accum, name) => ({ ...accum }), {} as Record<'a' | 'b', boolean>);
//	invalid: [1, 2, 3].reduce((a, s) => a.concat(s * 2), [] as number[]);
//	invalid: [1, 2, 3].reduce((a, s) => a.concat(s * 2), <number[]>[]);
//
// `reduce` is generic in its accumulator, so `reduce<T>(callback, initial)` tells the checker what
// the accumulator is and then checks the initial value AGAINST it. Writing `initial as T` instead
// tells the checker to stop checking: the assertion is believed rather than verified, so a mistake
// in the initial value survives to run time. The two spellings look equivalent and only one of them
// is still type-checked.
//
// # Why some assertions are left alone, and it is not a style judgment
//
// Upstream reports only when the assertion is UNNECESSARY, meaning the un-asserted initial value is
// already assignable to the asserted type. When it is not assignable, the assertion is doing real
// widening work and rewriting it into a type parameter would produce code that does not compile,
// so the rule declines rather than proposing a fix it knows is broken.
//
// That single test is what separates four of upstream's passing cases from cases that otherwise
// look identical:
//
//	{} as Record<'a' | 'b', boolean>            silent: {} is missing 'a' and 'b'
//	{ a: true, b: false, c: true } as Record<'a' | 'b', boolean>
//	                                            silent: excess property 'c'
//	{} as T   where T extends Record<...>       silent: T could be a narrower subtype
//	{} as Record<string, boolean>               REPORTS: {} really is assignable
//
// So the assignability question is load-bearing rather than an optimization, and a port that
// skipped it would report all four and ship a fixer that breaks the build on three of them.
//
// # The receiver has to be an array, and a union is judged constituent by constituent
//
// Upstream's `isArrayType` is `unionConstituents(t).every(part => intersectionConstituents(part)
// .every(inner => isArrayType(inner) || isTupleType(inner)))`. A tuple counts, which is why
// `declare const tuple: [number, number, number]` reports while `tuple | Reducer` does not, and an
// intersection has to be all arrays too, which is why `number[] & Reducer` is silent. The receiver
// type comes from `GetConstrainedTypeAtLocation` so `U extends T[]` walks to its constraint.
//
// # What the repair writes, measured rather than reasoned
//
// The fix removes the assertion's syntax around the value and, when the call has no type arguments
// of its own, inserts `<AssertedType>` after the callee. Three separate edits, and the third is
// conditional. Measured against the installed 8.67.0 build:
//
//	a.reduce(cb, [] as number[])              -> a.reduce<number[]>(cb, [])
//	a.reduce(cb, <number[]>[])                -> a.reduce<number[]>(cb, [])
//	a.reduce<string|undefined>(cb, x as string|undefined)
//	                                          -> a.reduce<string|undefined>(cb, x)
//	                                             the type argument is already there, so only the
//	                                             assertion is stripped
//	(a.reduce)(cb, [] as number[])            -> (a.reduce<number[]>)(cb, [])
//	                                             the insert anchors on the member access, so it
//	                                             lands INSIDE the parentheses. That is upstream's
//	                                             output byte for byte, not an artifact of this port.
//
// The asserted type's text is copied from the source rather than re-rendered through the checker.
// `typeToString` would normalize a union's spacing and expand an alias, so a fixer built on it
// would rewrite type text the author chose. Upstream copies the source text and so does this.
//
// # Two divergences from the installed build, both measured, both stated
//
// A computed member key folded through a variable is a FALSE NEGATIVE here:
//
//	const key = 'reduce'; a[key](cb, [] as number[])   REPORTS upstream, SILENT here
//
// Upstream resolves it with ESLint's `getStaticValue` over the enclosing scope, which is the
// constant-folding layer this tree does not have. A plain `.reduce` and a string-literal computed
// key `['reduce']` are both handled, and the corpus exercises the literal form in both lists. The
// gap is one shape wide and it costs findings rather than inventing them.
//
// `[] as const` is a defect upstream ships and this port does NOT reproduce:
//
//	a.reduce(cb, [] as const)   REPORTS upstream, writing `a.reduce<const>(cb, [])`
//
// `const` is not a type, so upstream's own fix produces source that does not parse. The brief's
// standing instruction is that a port is not obliged to carry a defect it can see, and a fix
// applied unattended that breaks the file is the worst version of that. A const assertion is
// declined here, at the site below. The corpus writes no `as const` anywhere, so nothing imported
// can see either behavior.
//
// # Cost
//
// The anchor is a call expression, which is common, so the body exits on the argument count and the
// member name before touching the checker. The two type queries run only for a call to something
// named `reduce` whose second argument is a type assertion.
var PreferReduceTypeParameter = rule.Rule{
	Name: "@typescript-eslint/prefer-reduce-type-parameter",

	// Both halves of the judgment are type questions: whether the receiver is an array, and whether
	// the assertion is doing work. The checker is required.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// isArrayOrTupleThroughout answers upstream's `isArrayType`: every union constituent, and
		// within each, every intersection constituent, must be an array or a tuple.
		//
		// A non-union is its own single constituent and a non-intersection likewise, so the two
		// nested loops read the same for a plain `number[]` as for `[number] | number[]`.
		isArrayOrTupleThroughout := func(receiverType *checker.Type) bool {
			for unionPart := range type_checking.UnionTypePartsSeq(receiverType) {
				for _, part := range type_checking.IntersectionTypeParts(unionPart) {
					if !checker.Checker_isArrayType(ctx.TypeChecker, part) &&
						!checker.IsTupleType(part) {
						return false
					}
				}
			}
			return true
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				call := node.AsCallExpression()
				if call.Expression == nil {
					return
				}

				// Upstream's selector is `CallExpression > MemberExpression.callee`, and estree has
				// no parenthesized-expression node, so a parenthesized callee is transparent to it.
				// Our parser does produce one, so the skip is what reproduces upstream rather than a
				// free improvement over it. Measured: `(a.reduce)(cb, [] as number[])` reports
				// upstream, and the corpus writes no parenthesized callee at all.
				//
				// SkipParentheses dereferences its argument, so the nil test above comes first.
				callee := ast.SkipParentheses(call.Expression)

				receiver, memberName, isMemberAccess := reduceMemberAccessOf(callee)
				if !isMemberAccess || memberName != "reduce" {
					return
				}

				// Upstream reads `arguments[1]` and then tests `arguments.length < 2`, so a call
				// with fewer than two arguments never reaches the assertion test. A third argument
				// is not excluded: `reduce` takes two, but a trailing argument is a type error
				// rather than something this rule judges, and the installed build reports through
				// it.
				if call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
					return
				}

				// The second argument, with parentheses skipped for the same reason as the callee.
				// estree hands the rule the assertion directly for `f(cb, ([] as number[]))`, and
				// measured against the installed build that form reports with the repair landing
				// inside the parentheses.
				secondArgument := ast.SkipParentheses(call.Arguments.Nodes[1])

				assertedValue, assertedType, isAssertion := typeAssertionParts(secondArgument)
				if !isAssertion {
					return
				}

				// `x as const` is a const assertion rather than a type assertion to a named type.
				// Upstream does not exclude it and its fixer writes `reduce<const>`, which does not
				// parse. Declined here rather than reproduced, because this rule's repair is a FIX
				// and a fix is applied with nobody watching.
				if isConstAssertionType(assertedType) {
					return
				}

				// Upstream: report only when the assertion is UNNECESSARY, meaning the value is
				// already assignable to the asserted type. An assertion doing real work is left
				// alone because rewriting it into a type parameter would stop compiling.
				initializerType := ctx.TypeChecker.GetTypeAtLocation(assertedValue)
				assertedTypeValue := ctx.TypeChecker.GetTypeAtLocation(assertedType)
				if initializerType == nil || assertedTypeValue == nil {
					return
				}
				if !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, initializerType, assertedTypeValue) {
					return
				}

				receiverType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, receiver)
				if receiverType == nil || !isArrayOrTupleThroughout(receiverType) {
					return
				}

				// The three edits, in upstream's order. The first two strip the assertion syntax
				// from around the value, which covers both spellings without either branching on
				// which one it is: for `[] as number[]` the leading range is empty and the trailing
				// one is ` as number[]`, and for `<number[]>[]` it is the other way around.
				//
				// Ranges are taken from the trimmed token bounds rather than Pos(), which includes
				// leading trivia and would make the first removal swallow the whitespace or the
				// comment before the argument.
				assertionRange := rule.TokenRange(ctx.SourceFile, secondArgument)
				valueRange := rule.TokenRange(ctx.SourceFile, assertedValue)

				fixes := []rule.Fix{
					rule.RemoveRange(core.NewTextRange(assertionRange.Pos(), valueRange.Pos())),
					rule.RemoveRange(core.NewTextRange(valueRange.End(), assertionRange.End())),
				}

				// Only when the call has no type arguments of its own. Upstream's corpus pins this
				// with `arr.reduce<string | undefined>(acc => acc, arr.shift() as string |
				// undefined)`, which reports and whose output keeps the one type argument already
				// written rather than gaining a second.
				if call.TypeArguments == nil {
					// The asserted type's own source text, copied rather than re-rendered. The
					// insert anchors on the callee BEFORE parentheses were skipped only in the
					// sense that upstream anchors on the member expression; measured, that puts the
					// text inside a parenthesized callee, which is what the installed build writes.
					typeText := ctx.SourceFile.Text()[rule.TokenRange(ctx.SourceFile, assertedType).Pos():assertedType.End()]
					fixes = append(fixes, ctx.InsertAfter(callee, "<"+typeText+">"))
				}

				ctx.ReportNodeWithFixes(secondArgument, buildPreferTypeParameterMessage(), fixes...)
			},
		}
	},
}

// reduceMemberAccessOf reads a callee as a member access, returning the receiver and member name.
//
// A property access supplies the name directly. A computed access supplies it only when the key is
// a string literal. Upstream's `isStaticMemberAccessOfValue` additionally folds a computed key
// through a `const` binding using ESLint's scope analysis, which is the divergence recorded in the
// rule's doc comment.
//
// Optional-chain forms are accepted rather than excluded: the corpus carries `[1, 2, 3]?.reduce(...)`
// in the reporting list and `[1, 2, 3]?.[null](...)` in the passing one, so the chain is not the
// discriminator and the key is.
func reduceMemberAccessOf(callee *ast.Node) (receiver *ast.Node, memberName string, isMemberAccess bool) {
	if ast.IsPropertyAccessExpression(callee) {
		access := callee.AsPropertyAccessExpression()
		if access.Name() == nil || !ast.IsIdentifier(access.Name()) {
			return nil, "", false
		}
		return access.Expression, access.Name().Text(), true
	}

	if ast.IsElementAccessExpression(callee) {
		access := callee.AsElementAccessExpression()
		key := access.ArgumentExpression
		if key == nil || key.Kind != ast.KindStringLiteral {
			return nil, "", false
		}
		return access.Expression, key.Text(), true
	}

	return nil, "", false
}

// typeAssertionParts reads either spelling of a type assertion, returning the value and the type.
//
// `x as T` and `<T>x` are different nodes with the same meaning, and upstream's `isTypeAssertion`
// accepts both. `x satisfies T` is deliberately NOT one: it checks without changing the type, so
// there is nothing unnecessary to remove, and the installed build is silent on it. Measured.
func typeAssertionParts(node *ast.Node) (value *ast.Node, assertedType *ast.Node, isAssertion bool) {
	switch node.Kind {
	case ast.KindAsExpression:
		expression := node.AsAsExpression()
		if expression.Expression == nil || expression.Type == nil {
			return nil, nil, false
		}
		return expression.Expression, expression.Type, true
	case ast.KindTypeAssertionExpression:
		expression := node.AsTypeAssertion()
		if expression.Expression == nil || expression.Type == nil {
			return nil, nil, false
		}
		return expression.Expression, expression.Type, true
	}
	return nil, nil, false
}

// isConstAssertionType answers whether an asserted type is the `const` of a const assertion.
//
// The parser gives `as const` a type reference whose name is the identifier `const`, which is not a
// type anyone can write elsewhere, so the name test cannot collide with a real type called `const`.
func isConstAssertionType(assertedType *ast.Node) bool {
	if !ast.IsTypeReferenceNode(assertedType) {
		return false
	}
	typeName := assertedType.AsTypeReferenceNode().TypeName
	return typeName != nil && ast.IsIdentifier(typeName) && typeName.Text() == "const"
}

func buildPreferTypeParameterMessage() rule.Message {
	return rule.Message{
		Id: "preferTypeParameter",
		Description: "Unnecessary assertion: Array#reduce accepts a type parameter for the " +
			"default value.",
	}
}
