package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnsafeOptionalChain = rule.Message{
	Id: "unsafeOptionalChain",
	Description: "This optional chain can evaluate to `undefined`, and the code around it then " +
		"reaches into that `undefined` rather than short-circuiting with it. The `?.` protects " +
		"only the links inside its own chain: once the chain ends, the result is an ordinary " +
		"value and the next access, call, spread, or destructuring throws a TypeError. Either " +
		"extend the chain so the whole expression short-circuits, or supply a fallback with `??`.",
}

var messageUnsafeArithmetic = rule.Message{
	Id: "unsafeArithmetic",
	Description: "This optional chain can evaluate to `undefined`, and arithmetic on `undefined` " +
		"produces NaN rather than throwing. So this does not fail where the value is missing, it " +
		"silently poisons every number computed from it, and the NaN surfaces somewhere far from " +
		"the cause. Supply a fallback with `??` before doing arithmetic on it.",
}

// NoUnsafeOptionalChainingOptions mirrors the single switch ESLint's rule carries.
//
// Confirmed against `meta.schema` rather than an inventory column: one object with one boolean
// property and `additionalProperties: false`. A plain bool rather than a pointer because the
// default is false, so the zero value and the default agree and an options object that omits the
// key behaves like no options at all. That equivalence is asserted by fixture.
//
// The option is off by default upstream, and it is the looser setting rather than the stricter
// one, because arithmetic on a missing value is a weaker signal than reaching into it: `?? 0` is
// often already implied by the surrounding code.
type NoUnsafeOptionalChainingOptions struct {
	DisallowArithmeticOperators bool `json:"disallowArithmeticOperators"`
}

// arithmeticOperators are the binary operators for which an `undefined` operand yields NaN.
//
// Exactly ESLint's UNSAFE_ARITHMETIC_OPERATORS. Bitwise and shift operators are deliberately
// absent even though they also coerce: they coerce `undefined` to 0 rather than to NaN, so the
// result is a number and the failure mode this message describes does not arise.
var arithmeticOperators = map[ast.Kind]bool{
	ast.KindPlusToken:             true,
	ast.KindMinusToken:            true,
	ast.KindSlashToken:            true,
	ast.KindAsteriskToken:         true,
	ast.KindPercentToken:          true,
	ast.KindAsteriskAsteriskToken: true,
}

// arithmeticAssignmentOperators are the compound assignments carrying the same coercion.
//
// Exactly ESLint's UNSAFE_ASSIGNMENT_OPERATORS. A plain `=` is absent because assigning
// `undefined` to a binding is legal and is one of upstream's clean cases.
var arithmeticAssignmentOperators = map[ast.Kind]bool{
	ast.KindPlusEqualsToken:             true,
	ast.KindMinusEqualsToken:            true,
	ast.KindSlashEqualsToken:            true,
	ast.KindAsteriskEqualsToken:         true,
	ast.KindPercentEqualsToken:          true,
	ast.KindAsteriskAsteriskEqualsToken: true,
}

// relationalOperators are the operators whose right operand is reached into rather than compared.
//
// `in` looks up a key on its right operand and `instanceof` reads its prototype, so both throw on
// `undefined`. The ordering comparisons (`<`, `>=`) are absent: they coerce rather than reach, so
// `foo?.bar < baz` is merely false and is one of upstream's clean cases.
var relationalOperators = map[ast.Kind]bool{
	ast.KindInKeyword:         true,
	ast.KindInstanceOfKeyword: true,
}

// NoUnsafeOptionalChaining flags an optional chain whose `undefined` result would be used in a
// position that cannot accept it.
//
//	valid:   obj?.foo.bar
//	valid:   (obj?.foo ?? bar).baz
//	valid:   const foo = obj?.bar
//	valid:   var bar = {...foo?.bar}
//	invalid: (obj?.foo).bar
//	invalid: (obj?.foo)()
//	invalid: const {foo} = obj?.bar
//	invalid: foo in obj?.bar
//	invalid: const a = [...obj?.foo]
//
// Ported from ESLint's `no-unsafe-optional-chaining`, whose structure this follows, cross-checked
// against oxc's port for the TypeScript-only wrapper forms.
//
// # The rule is about where a chain ends, not where it starts
//
// `?.` short-circuits the whole chain it belongs to, so `obj?.foo.bar.baz` is safe: if `obj` is
// nullish the entire expression is `undefined` and `.bar` never runs. The danger begins where the
// chain stops and an ordinary operation consumes its result. Parentheses are the usual way that
// happens, which is why `(obj?.foo).bar` reports while `obj?.foo.bar` does not.
//
// # Anchoring: IsOutermostOptionalChain, and why not the other two
//
// ESLint anchors on `ChainExpression`, an ESTree node with no counterpart here. typescript-go
// instead flags every link of a chain with NodeFlagsOptionalChain, and offers three predicates.
// A probe over this rule's own corpus separated them:
//
//	IsOptionalChain          true for every link of a chain, including ones carrying no `?.`
//	IsOptionalChainRoot      true only for the link owning a `?.` token
//	IsOutermostOptionalChain true for the last link of a chain, AND for any node that is not
//	                         part of a chain at all
//
// That third row is the one worth stating, because the name does not suggest it and reading the
// name alone cost a working port: `IsOutermostOptionalChain` returns true for a plain binary
// expression and for a parenthesized expression. It answers "is nothing above this continuing a
// chain", not "is this a chain". So the ChainExpression anchor is the conjunction,
// `IsOptionalChain(node) && IsOutermostOptionalChain(node)`, and that pair is the boundary where
// short-circuiting stops. Every one of the twenty diagnostics in upstream's snapshot underlines
// exactly that node.
//
// IsOptionalChainRoot is what translates ESLint's per-node `optional` boolean, used below to ask
// whether a member access or call is itself part of the chain.
//
// A sweep measured something worth recording about that guard. Removing it entirely is caught by
// fifteen fixture lines, so it is load-bearing. But *weakening* it from IsOptionalChainRoot to
// IsOptionalChain survives every fixture, and a probe over twenty four chain shapes found zero
// inputs that could distinguish them. The reason is structural rather than a gap in the corpus: the
// substitution only changes behavior at a node with chain=true and root=false, which is a link in
// the middle of a chain, and such a link's operand is by construction the previous link of that same
// chain and therefore never the outermost one. The anchor conjunction below already declines it. So
// the two predicates are interchangeable *here* while the guard itself is not removable, and the
// faithful one is kept because the equivalence rests on the anchor rather than on anything local. This distinction is the documented
// trap in this tree, recorded at `internal/rules/typescript/no_extra_non_null_assertion.go`:
// reading IsOptionalChain where the original reads `optional` over-reports, because it is true for
// links with no `?.` of their own. Here it would make `obj?.foo.bar` look like a chain consumed by
// an optional access, which is the right answer by accident, and `(obj?.foo).bar` look the same
// way, which is wrong. The probe rather than the reasoning is what settled it.
//
// # Reported per chain rather than per unsafe use
//
// `(obj?.foo && obj?.baz).bar` reports twice. The short circuit propagates through both operands
// of `&&`, so both chains reach the unsafe access and each is separately something the author
// must fix. This is why upstream's 18 failing inputs produce 20 diagnostics.
var NoUnsafeOptionalChaining = rule.Rule{
	Name: "no-unsafe-optional-chaining",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		disallowArithmeticOperators := false
		if settings, hasSettings := rule.OptionsAs[NoUnsafeOptionalChainingOptions](options); hasSettings {
			disallowArithmeticOperators = settings.DisallowArithmeticOperators
		}

		reportUnsafeUsage := func(node *ast.Node) {
			reportChainsReaching(ctx, node, messageUnsafeOptionalChain)
		}
		reportUnsafeArithmetic := func(node *ast.Node) {
			reportChainsReaching(ctx, node, messageUnsafeArithmetic)
		}

		listeners := rule.Listeners{
			// Reaching into the result: a property or element access whose object is a chain, and
			// which is not itself part of that chain.
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if !ast.IsOptionalChainRoot(node) {
					reportUnsafeUsage(node.Expression())
				}
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if !ast.IsOptionalChainRoot(node) {
					reportUnsafeUsage(node.Expression())
				}
			},
			// Calling the result. `obj?.foo?.()` is safe because the call carries its own `?.`.
			ast.KindCallExpression: func(node *ast.Node) {
				if !ast.IsOptionalChainRoot(node) {
					reportUnsafeUsage(node.Expression())
				}
			},
			// `new` has no optional form, so there is nothing to exempt.
			ast.KindNewExpression: func(node *ast.Node) {
				reportUnsafeUsage(node.Expression())
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				reportUnsafeUsage(node.AsTaggedTemplateExpression().Tag)
			},
			// `for (x of undefined)` throws; `for (x in undefined)` does not, and ESLint listens
			// only for the of-form.
			ast.KindForOfStatement: func(node *ast.Node) {
				reportUnsafeUsage(node.AsForInOrOfStatement().Expression)
			},
			ast.KindWithStatement: func(node *ast.Node) {
				reportUnsafeUsage(node.AsWithStatement().Expression)
			},
			// Extending `undefined` throws. Both a declaration and an expression reach here,
			// since typescript-go gives them a shared heritage shape.
			ast.KindClassDeclaration: func(node *ast.Node) {
				reportUnsafeUsage(classHeritageExpression(node))
			},
			ast.KindClassExpression: func(node *ast.Node) {
				reportUnsafeUsage(classHeritageExpression(node))
			},
			// Destructuring `undefined` throws, whether the pattern is an object or an array.
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if isDestructuringPattern(declaration.Name()) {
					reportUnsafeUsage(declaration.Initializer)
				}
			},
			// A default value in a parameter or pattern position, ESLint's AssignmentPattern.
			ast.KindParameter: func(node *ast.Node) {
				parameter := node.AsParameterDeclaration()
				if isDestructuringPattern(parameter.Name()) {
					reportUnsafeUsage(parameter.Initializer)
				}
			},
			ast.KindBindingElement: func(node *ast.Node) {
				element := node.AsBindingElement()
				if isDestructuringPattern(element.Name()) {
					reportUnsafeUsage(element.Initializer)
				}
			},
			// Array spread reads the iterator off its argument and throws on `undefined`, so a
			// chain spread into an array literal or a call argument list is unsafe.
			//
			// Object spread is safe (`{...undefined}` is an empty object) and ESLint expresses
			// that as a guard on the spread's parent. No such guard is needed here, and writing
			// one would be dead code: typescript-go gives object spread its own node kind,
			// KindSpreadAssignment, so this listener never sees it. A sweep removing the parent
			// check survived every fixture, which is what sent me to look; the branch was
			// unreachable rather than untested, so it is deleted rather than covered.
			ast.KindSpreadElement: func(node *ast.Node) {
				reportUnsafeUsage(node.Expression())
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				operator := binary.OperatorToken.Kind

				// `in` and `instanceof` reach into their right operand.
				if relationalOperators[operator] {
					reportUnsafeUsage(binary.Right)
				}

				// Destructuring assignment, `({foo} = obj?.bar)`. ESLint reaches this through its
				// AssignmentExpression selector; here an assignment is a binary expression.
				//
				// This is checked before the arithmetic option is consulted, and the ordering is
				// load-bearing rather than stylistic: destructuring `undefined` throws whatever
				// the option says, so gating it behind `disallowArithmeticOperators` made the
				// whole arm dead under the default configuration. Upstream's fail case for it
				// carries no options at all, and that fixture is what caught this.
				if operator == ast.KindEqualsToken && isDestructuringPattern(binary.Left) {
					reportUnsafeUsage(binary.Right)
					return
				}

				if !disallowArithmeticOperators {
					return
				}
				// Arithmetic poisons the result from either side, so both operands are checked.
				if arithmeticOperators[operator] {
					reportUnsafeArithmetic(binary.Right)
					reportUnsafeArithmetic(binary.Left)
					return
				}
				// A compound assignment reads its left operand and writes the result, so only the
				// right side introduces the `undefined`.
				if arithmeticAssignmentOperators[operator] {
					reportUnsafeArithmetic(binary.Right)
				}
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				if !disallowArithmeticOperators {
					return
				}
				unary := node.AsPrefixUnaryExpression()
				// Only `+x` and `-x` coerce to a number. `!x`, `~x`, `typeof x` and the update
				// operators are either safe on `undefined` or not arithmetic.
				if unary.Operator == ast.KindPlusToken || unary.Operator == ast.KindMinusToken {
					reportUnsafeArithmetic(unary.Operand)
				}
			},
		}

		return listeners
	},
}

// reportChainsReaching reports every optional chain whose `undefined` could reach the given
// position.
//
// This is ESLint's checkUndefinedShortCircuit, and the recursion is the whole subtlety of the
// rule. A chain does not have to sit directly in the unsafe position to be dangerous: several
// operators pass an `undefined` operand through to their own result, so the search descends
// through each of them and reports at every chain it lands on.
//
//	a ?? b, a || b     only the right survives when the left is present, so only the right
//	                   propagates. This is why `(obj?.foo ?? bar).baz` is clean: the chain is on
//	                   the left, and `bar` is what reaches `.baz` when the chain short-circuits.
//	a && b             either operand can be the result, so both propagate.
//	a, b               the value of a sequence is its last expression, so only that one.
//	c ? a : b          either branch can be the result, so both.
//	await a            awaiting `undefined` yields `undefined`.
//
// Anything else stops the descent, which is what makes the default arm a decision rather than an
// oversight: reaching a call or a literal means whatever `undefined` was there has been consumed
// by something that handled it.
func reportChainsReaching(ctx rule.Context, node *ast.Node, message rule.Message) {
	if node == nil {
		return
	}

	// Parentheses and the type-only wrappers are transparent to evaluation. ESTree has no node for
	// any of them, so ESLint sees through them for free and its switch never mentions them; here
	// they are real nodes and a descent that did not skip them would stop at the wrapper and
	// report nothing. This is the arm that carries every parenthesized case in the corpus.
	switch node.Kind {
	case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
		ast.KindTypeAssertionExpression, ast.KindNonNullExpression, ast.KindExpressionWithTypeArguments:
		reportChainsReaching(ctx, node.Expression(), message)
		return
	case ast.KindAwaitExpression:
		reportChainsReaching(ctx, node.AsAwaitExpression().Expression, message)
		return
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		reportChainsReaching(ctx, conditional.WhenTrue, message)
		reportChainsReaching(ctx, conditional.WhenFalse, message)
		return
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			reportChainsReaching(ctx, binary.Right, message)
		case ast.KindAmpersandAmpersandToken:
			reportChainsReaching(ctx, binary.Left, message)
			reportChainsReaching(ctx, binary.Right, message)
		case ast.KindCommaToken:
			// The value of a comma expression is its right side; the left is evaluated and
			// discarded, which is why `(foo?.bar, bar)()` is clean.
			reportChainsReaching(ctx, binary.Right, message)
		}
		return
	}

	// The chain itself.
	//
	// Both predicates are needed and neither is redundant, which a probe established rather than
	// reading: IsOutermostOptionalChain answers "is this the last link of the chain it belongs
	// to" and returns true for a node belonging to no chain at all, including a plain binary
	// expression and a parenthesized one. IsOptionalChain is what asks whether there is a chain
	// here. Anchoring on the outermost predicate alone reported `foo?.bar in {}` and
	// `(obj?.foo ?? bar).baz`, both of which upstream marks clean, and it was upstream's clean
	// cases rather than any invented fixture that caught it.
	if ast.IsOptionalChain(node) && ast.IsOutermostOptionalChain(node) {
		ctx.ReportNode(node, message)
	}
}

// isDestructuringPattern reports whether a node is an object or array binding pattern.
//
// ESLint's isDestructuringPattern, which decides whether an assignment or declaration reaches
// into its right side rather than merely storing it. Both binding forms (a declaration's pattern)
// and literal forms (an assignment's target, which parses as an object or array literal) count,
// because `({foo} = obj?.bar)` and `const {foo} = obj?.bar` are the same operation written in two
// positions and upstream reports both.
func isDestructuringPattern(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern,
		ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		return true
	}
	return false
}

// classHeritageExpression returns the expression a class extends, or nil when it extends nothing.
//
// The extends clause is reached through the heritage clause list rather than a direct field, and
// an interface's `implements` list lives in the same place under a different token, so the
// keyword check is load-bearing rather than defensive.
func classHeritageExpression(node *ast.Node) *ast.Node {
	heritageClauses := node.ClassLikeData().HeritageClauses
	if heritageClauses == nil {
		return nil
	}
	for _, clause := range heritageClauses.Nodes {
		heritage := clause.AsHeritageClause()
		if heritage.Token != ast.KindExtendsKeyword {
			continue
		}
		if heritage.Types == nil || len(heritage.Types.Nodes) == 0 {
			return nil
		}
		return heritage.Types.Nodes[0]
	}
	return nil
}
