package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageConstantShortCircuit = rule.Message{
	Id: "constantShortCircuit",
	Description: "The left side of this short-circuit always has the same truthiness or " +
		"nullishness, so the operator never chooses: one side always wins and the other is dead. " +
		"The usual cause is a guard written against the wrong thing, such as testing an object " +
		"that is always present rather than the field inside it.",
}

var messageConstantBinaryOperand = rule.Message{
	Id: "constantBinaryOperand",
	Description: "This comparison always gives the same answer, because one side can never equal " +
		"the other. Nothing reports it: the code compiles and the branch it guards simply never " +
		"changes. It is usually a comparison against the wrong thing, such as testing an array or " +
		"a function for null when neither can be.",
}

var messageAlwaysNew = rule.Message{
	Id: "alwaysNew",
	Description: "This side of the comparison builds a fresh object every time it runs, and " +
		"strict equality on objects compares identity rather than contents, so the comparison is " +
		"always false. The intent is nearly always to compare contents, which needs an explicit " +
		"check rather than an operator.",
}

var messageBothAlwaysNew = rule.Message{
	Id: "bothAlwaysNew",
	Description: "Both sides build a fresh object, and two fresh objects are never the same " +
		"object, so this comparison is always false even under loose equality: two objects are " +
		"compared by identity whichever equality operator is used.",
}

var messageConstantRelationalComparison = rule.Message{
	Id: "constantRelationalComparison",
	Description: "Both sides of this comparison are literals, so the answer is fixed and could be " +
		"written as true or false. A comparison that decides nothing is usually a leftover from " +
		"code that used to compare something real.",
}

// NoConstantBinaryExpressionOptions configures the relational arm.
type NoConstantBinaryExpressionOptions struct {
	// CheckRelationalComparisons turns on reporting `1 < 2`.
	//
	// Off in this tree, matching ESLint's default. Implemented anyway rather than skipped, because
	// a rule that silently lacks an option is indistinguishable from one whose option is off, and
	// the difference only surfaces when someone turns it on.
	CheckRelationalComparisons bool `json:"checkRelationalComparisons"`
}

// NoConstantBinaryExpression flags comparisons and short-circuits whose result cannot vary.
//
//	valid:   a ?? b
//	valid:   a === b
//	invalid: [] ?? a                  an array is never nullish, so ?? never chooses
//	invalid: a === []                 nothing equals a fresh array
//	invalid: [] == true               an empty array's boolean comparison is fixed
//	invalid: {} === {}                two fresh objects are never the same object
//
// The judgment is per operator and per side rather than one notion of constant, which is what makes
// this rule long. `??` asks whether the left side's nullishness is fixed, `&&` asks whether its
// truthiness is, `==` against a boolean asks whether the other side's loose coercion is fixed, and
// `===` against an object asks whether either side is freshly built. Four different questions, and
// collapsing them would report correct code.
//
// ESLint asks its scope analysis whether `undefined`, `Boolean`, `String`, and the built-in
// constructors are the globals rather than local shadows, and so does this, through the checker:
// a name declared in no source file of the program is the global. `function foo(undefined) {
// undefined ?? bar }` and a local `function Boolean(n)` are ordinary values, and reading the names
// as the globals reported 9 of ESLint's corpus rows that it pins as clean (#jjfa7qb).
//
// The relational arm is off in this tree, so it cannot be validated against the differential. It is
// implemented because a missing option and a disabled one look identical until someone enables it.
var NoConstantBinaryExpression = rule.Rule{
	Name: "no-constant-binary-expression",

	// Whether `undefined`, `Boolean` and the built-in constructors are the globals
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		judge := constantJudge{isGlobal: func(identifier *ast.Node) bool {
			return ctx.TypeChecker == nil || !rule.IsDeclaredInASourceFile(ctx.TypeChecker.GetSymbolAtLocation(identifier))
		}}
		checkRelationalComparisons := false
		if parsed, ok := rule.OptionsAs[NoConstantBinaryExpressionOptions](options); ok {
			checkRelationalComparisons = parsed.CheckRelationalComparisons
		}

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil {
					return
				}
				left := ast.SkipParentheses(binary.Left)
				right := ast.SkipParentheses(binary.Right)
				if left == nil || right == nil {
					return
				}

				switch binary.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
					// The left side alone decides whether the operator ever chooses.
					if isConstantExpression(left, true) {
						ctx.ReportNode(left, messageConstantShortCircuit)
					}

				case ast.KindQuestionQuestionToken:
					if judge.hasConstantNullishness(left, false) {
						ctx.ReportNode(left, messageConstantShortCircuit)
					}

				case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
					if operand := judge.constantComparisonOperand(left, right, true); operand != nil {
						ctx.ReportNode(operand, messageConstantBinaryOperand)
					} else if operand := judge.constantComparisonOperand(right, left, true); operand != nil {
						ctx.ReportNode(operand, messageConstantBinaryOperand)
					} else if judge.isAlwaysNew(left) {
						ctx.ReportNode(left, messageAlwaysNew)
					} else if judge.isAlwaysNew(right) {
						ctx.ReportNode(right, messageAlwaysNew)
					}

				case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
					if operand := judge.constantComparisonOperand(left, right, false); operand != nil {
						ctx.ReportNode(operand, messageConstantBinaryOperand)
					} else if operand := judge.constantComparisonOperand(right, left, false); operand != nil {
						ctx.ReportNode(operand, messageConstantBinaryOperand)
					} else if judge.isAlwaysNew(left) && judge.isAlwaysNew(right) {
						// Both sides only. Under loose equality a single fresh object can still
						// equal a primitive through coercion, so one is not enough.
						ctx.ReportNode(left, messageBothAlwaysNew)
					}

				case ast.KindLessThanToken, ast.KindLessThanEqualsToken,
					ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
					if checkRelationalComparisons && judge.isStaticLiteral(left) && judge.isStaticLiteral(right) {
						ctx.ReportNode(node, messageConstantRelationalComparison)
					}
				}
			},
		}
	},
}

// constantJudge carries the one question the judgments below need of the program: whether a name is
// the global rather than a local that shares its spelling
type constantJudge struct {
	isGlobal func(identifier *ast.Node) bool
}

// isGlobalBooleanCall is isConstantBooleanCall asked of the global `Boolean` only, so a local function
// named Boolean returns whatever it returns
func (judge constantJudge) isGlobalBooleanCall(node *ast.Node) bool {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	return callee != nil && judge.isGlobal(callee) && isConstantBooleanCall(node)
}

// constantComparisonOperand returns the operand whose comparison against the other is fixed.
//
// Two ways that happens, and they are different questions. Comparing against null or undefined asks
// whether the other side's nullishness is fixed: an array is never nullish, so `a === null` on an
// array is always false. Comparing against a boolean asks whether the other side's coercion to
// boolean is fixed, which is a different and looser question under `==` than under `===`.
func (judge constantJudge) constantComparisonOperand(known *ast.Node, other *ast.Node, strict bool) *ast.Node {
	if judge.isNullishValue(known) && judge.hasConstantNullishness(other, false) {
		return other
	}
	if judge.isStaticBoolean(known) {
		if strict {
			if judge.hasConstantStrictBooleanComparison(other) {
				return other
			}
		} else if judge.hasConstantLooseBooleanComparison(other) {
			return other
		}
	}
	return nil
}

// isNullishValue reports the two nullish values, however they are spelled.
//
// Named apart from prefer_spread's isNullOrUndefined rather than merged with it, because the two
// ask different questions despite the near-identical name. That one is looking for "no receiver"
// in a `.call(null, ...)` position and accepts only `void 0`; this one is asking whether a value is
// nullish and accepts any `void`, since `void anything` evaluates to undefined. Merging them would
// widen a rule nobody asked me to change.
func (judge constantJudge) isNullishValue(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNullKeyword, ast.KindVoidExpression:
		return true
	case ast.KindIdentifier:
		return node.Text() == "undefined" && judge.isGlobal(node)
	}
	return false
}

// isStaticBoolean reports a literal true or false, or a Boolean call with a constant argument.
func (judge constantJudge) isStaticBoolean(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindCallExpression:
		return judge.isGlobalBooleanCall(node)
	}
	return false
}

// hasConstantNullishness reports whether an expression is always nullish or always is not.
//
// nonNullish says the caller has already established the node is not itself null or undefined,
// which is what stops `null ?? a` reporting on a left side that is exactly the value `??` exists to
// handle.
func (judge constantJudge) hasConstantNullishness(node *ast.Node, nonNullish bool) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if nonNullish && judge.isNullishValue(node) {
		return false
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression,
		ast.KindNewExpression,
		ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression,
		ast.KindPostfixUnaryExpression:
		// Objects, arrays, functions, classes, constructed values, literals, template strings, and
		// the result of ++ or -- are never nullish, and a nullish literal is always nullish. Either
		// way the answer does not vary.
		return true

	case ast.KindIdentifier:
		return node.Text() == "undefined" && judge.isGlobal(node)

	case ast.KindVoidExpression, ast.KindTypeOfExpression, ast.KindPrefixUnaryExpression,
		ast.KindDeleteExpression, ast.KindAwaitExpression:
		// void is always undefined; typeof is always a string; the unary math and logical
		// operators all yield numbers, strings, or booleans. Await is ESLint's addition and is
		// deliberately not claimed here, so it falls through to false below.
		switch node.Kind {
		case ast.KindAwaitExpression:
			return false
		}
		return true

	case ast.KindCallExpression:
		// The five wrapper constructors always return a value of their own type.
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee == nil || callee.Kind != ast.KindIdentifier {
			return false
		}
		if !judge.isGlobal(callee) {
			return false
		}
		switch callee.Text() {
		case "Boolean", "String", "Number", "Symbol", "BigInt":
			return true
		}
		return false

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindQuestionQuestionToken:
			// `a ?? b` is nullish exactly when b is, since a nullish a yields b.
			return judge.hasConstantNullishness(binary.Right, true)
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			// Either operand may be the result, so nothing fixed can be said.
			return false
		case ast.KindEqualsToken:
			return judge.hasConstantNullishness(binary.Right, nonNullish)
		case ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken,
			ast.KindQuestionQuestionEqualsToken:
			// A logical assignment yields one side or the other; reasoning about it needs the
			// variable's value, which ESLint declines to do and so do we.
			return false
		}
		// Every other binary operator yields a number, string, or boolean.
		return true
	}

	return false
}

// hasConstantLooseBooleanComparison reports whether `x == true` always gives the same answer.
//
// Looser than the strict version because `==` coerces, so the question is whether the coercion is
// fixed rather than whether the node can be a boolean at all. `[]` coerces to `0`, `[1]` coerces to
// `1`, so an array of one element is not fixed while an empty one or a longer one is.
func (judge constantJudge) hasConstantLooseBooleanComparison(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindClassExpression,
		ast.KindArrowFunction, ast.KindFunctionExpression,
		ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return true

	case ast.KindIdentifier:
		return node.Text() == "undefined" && judge.isGlobal(node)

	case ast.KindArrayLiteralExpression:
		// `[x]` coerces to whatever x stringifies to, which could be "0" or "1", so a single
		// element is the one case that varies. Empty or several elements always coerce to
		// something that is neither.
		elements := node.AsArrayLiteralExpression().Elements.Nodes
		nonSpread := 0
		for _, element := range elements {
			if element != nil && element.Kind != ast.KindSpreadElement {
				nonSpread++
			}
		}
		return len(elements) == 0 || nonSpread > 1

	case ast.KindTemplateExpression:
		// A template with substitutions could produce "0" or "1".
		return len(node.AsTemplateExpression().TemplateSpans.Nodes) == 0

	case ast.KindVoidExpression, ast.KindTypeOfExpression:
		return true

	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindExclamationToken {
			return isConstantExpression(unary.Operand, true)
		}
		// Not reasoned about: +, -, and ~ could coerce to 0 or 1.
		return false

	case ast.KindNewExpression:
		// A constructed object may define valueOf or toString.
		return false

	case ast.KindCallExpression:
		return judge.isGlobalBooleanCall(node)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken:
			return judge.hasConstantLooseBooleanComparison(binary.Right)
		case ast.KindEqualsToken:
			return judge.hasConstantLooseBooleanComparison(binary.Right)
		}
		return false
	}

	return false
}

// hasConstantStrictBooleanComparison reports whether `x === true` always gives the same answer.
//
// The question is simply whether the node can ever be a boolean at all: an array is never a
// boolean, so `[] === true` is always false whatever the array holds. That makes it broader than
// the loose version rather than narrower, which reads backwards until you see why.
func (judge constantJudge) hasConstantStrictBooleanComparison(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression,
		ast.KindNewExpression, ast.KindTemplateExpression,
		ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindPostfixUnaryExpression,
		ast.KindVoidExpression, ast.KindTypeOfExpression:
		return true

	case ast.KindIdentifier:
		return node.Text() == "undefined" && judge.isGlobal(node)

	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindExclamationToken {
			return isConstantExpression(unary.Operand, true)
		}
		// The math operators all yield numbers, which are never booleans.
		return true

	case ast.KindDeleteExpression:
		// delete yields a boolean, so it can be one.
		return false

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee == nil || callee.Kind != ast.KindIdentifier {
			return false
		}
		if !judge.isGlobal(callee) {
			return false
		}
		switch callee.Text() {
		case "String", "Number", "BigInt", "Symbol":
			// Never return a boolean.
			return true
		case "Boolean":
			return isConstantBooleanCall(node)
		}
		return false

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken, ast.KindEqualsToken:
			return judge.hasConstantStrictBooleanComparison(binary.Right)
		case ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken,
			ast.KindQuestionQuestionEqualsToken:
			return false
		}
		return isNumericOrStringBinaryOperator(binary.OperatorToken.Kind)
	}

	return false
}

// isNumericOrStringBinaryOperator reports the operators that always yield a number or a string.
//
// Deliberately excludes the comparison and logical operators, which can yield a boolean, and `in`
// and `instanceof`, which always do.
func isNumericOrStringBinaryOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindMinusToken, ast.KindPlusToken, ast.KindSlashToken, ast.KindAsteriskToken,
		ast.KindPercentToken, ast.KindBarToken, ast.KindCaretToken, ast.KindAmpersandToken,
		ast.KindAsteriskAsteriskToken, ast.KindLessThanLessThanToken,
		ast.KindGreaterThanGreaterThanToken, ast.KindGreaterThanGreaterThanGreaterThanToken:
		return true
	}
	return false
}

// isAlwaysNew reports whether an expression builds a fresh value every time it runs.
//
// Fresh means identity comparison can never succeed against anything, which is what makes
// `a === []` always false. A regular expression counts: it is an object.
//
// A `new` call counts only for the built-in constructors. A user-defined one may return a sentinel
// from its constructor, which is a real pattern and would make the comparison meaningful.
func (judge constantJudge) isAlwaysNew(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression,
		ast.KindRegularExpressionLiteral:
		return true

	case ast.KindNewExpression:
		callee := ast.SkipParentheses(node.AsNewExpression().Expression)
		if callee == nil || callee.Kind != ast.KindIdentifier {
			return false
		}
		return isEcmaScriptGlobalConstructor(callee.Text()) && judge.isGlobal(callee)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken, ast.KindEqualsToken:
			return judge.isAlwaysNew(binary.Right)
		}
		return false

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return judge.isAlwaysNew(conditional.WhenTrue) && judge.isAlwaysNew(conditional.WhenFalse)
	}

	return false
}

// ecmaScriptGlobalConstructors are the built-ins whose `new` always yields a fresh object.
//
// The boxed primitives are the ones worth catching: `new Number(1) === 1` is false, which surprises
// people, and it is exactly the mistake this arm exists for.
var ecmaScriptGlobalConstructors = map[string]bool{
	"Array": true, "ArrayBuffer": true, "BigInt64Array": true, "BigUint64Array": true,
	"Boolean": true, "DataView": true, "Date": true, "Error": true, "EvalError": true,
	"Float32Array": true, "Float64Array": true, "Function": true, "Int8Array": true,
	"Int16Array": true, "Int32Array": true, "Map": true, "Number": true, "Object": true,
	"Promise": true, "Proxy": true, "RangeError": true, "ReferenceError": true, "RegExp": true,
	"Set": true, "String": true, "SyntaxError": true, "TypeError": true, "URIError": true,
	"Uint8Array": true, "Uint8ClampedArray": true, "Uint16Array": true, "Uint32Array": true,
	"WeakMap": true, "WeakRef": true, "WeakSet": true,
}

func isEcmaScriptGlobalConstructor(name string) bool {
	return ecmaScriptGlobalConstructors[name]
}

// isStaticLiteral reports a value knowable without running anything.
func (judge constantJudge) isStaticLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return true

	case ast.KindIdentifier:
		return node.Text() == "undefined" && judge.isGlobal(node)

	case ast.KindTemplateExpression:
		return len(node.AsTemplateExpression().TemplateSpans.Nodes) == 0

	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindMinusToken, ast.KindPlusToken, ast.KindTildeToken:
			operand := ast.SkipParentheses(unary.Operand)
			return operand != nil &&
				(operand.Kind == ast.KindNumericLiteral || operand.Kind == ast.KindBigIntLiteral ||
					operand.Kind == ast.KindStringLiteral)
		}
	}

	return false
}
