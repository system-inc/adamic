package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedConstantCondition = rule.Message{
	Id: "unexpected",
	Description: "This condition is always the same value, so the branch it guards either always " +
		"runs or never does. Nothing reports it: the code compiles, the dead half type-checks, and " +
		"a reader sees a decision being made where none is. It is usually a comparison that lost " +
		"an operand, or a guard someone stubbed out and did not restore.",
}

// NoConstantConditionOptions configures which loops are checked.
type NoConstantConditionOptions struct {
	// CheckLoops is "all", "allExceptWhileTrue", or "none".
	//
	// The default matches ESLint's and this tree's configuration. `while(true)` is a deliberate idiom
	// rather than a mistake, so the default exempts it while still catching `while(1)` and
	// `while(a || true)`, which are not idioms and are usually accidents.
	CheckLoops NoConstantConditionCheckLoops `json:"checkLoops"`
}

// NoConstantConditionCheckLoops is the `checkLoops` value, in the three-string spelling.
type NoConstantConditionCheckLoops string

// UnmarshalJSON reads every value upstream's schema accepts and refuses the rest.
//
// The schema is `enum: ["all", "allExceptWhileTrue", "none", true, false]`, and upstream maps the
// two booleans onto the strings before anything reads them: true is "all" and false is "none". The
// booleans are the option's older spelling, and ESLint 10.8.1 still accepts them (#d21war2). An
// unlisted string is refused rather than reaching the rule, where it would match no case and check
// every loop.
func (checkLoops *NoConstantConditionCheckLoops) UnmarshalJSON(raw []byte) error {
	// `null` decodes into a bool as false without an error, which would read it as "none".
	if string(raw) == "null" {
		return fmt.Errorf("checkLoops is null, which upstream's schema refuses")
	}
	var enabled bool
	if err := json.Unmarshal(raw, &enabled); err == nil {
		*checkLoops = "none"
		if enabled {
			*checkLoops = "all"
		}
		return nil
	}
	var spelling string
	if err := json.Unmarshal(raw, &spelling); err != nil {
		return fmt.Errorf(`checkLoops is %s, which is neither a string nor a boolean`, raw)
	}
	switch spelling {
	case "all", "allExceptWhileTrue", "none":
		*checkLoops = NoConstantConditionCheckLoops(spelling)
		return nil
	}
	return fmt.Errorf(`checkLoops is %q, which is not "all", "allExceptWhileTrue", "none", true or false`, spelling)
}

// NoConstantCondition flags a condition whose value cannot vary.
//
//	valid:   if(a) {}
//	valid:   while(true) {}                 // the idiom, exempt by default
//	invalid: if(true) {}
//	invalid: if(a = 0) {}                   // an assignment mistaken for a comparison
//	invalid: while(1) {}
//	invalid: if(a || true) {}
//
// The interesting judgment is not "is this a literal" but the boolean-position distinction that
// runs through the whole test. `[]` is constant as a condition, since an array is always truthy,
// and not constant as a value, since its elements may vary. So the same node answers differently
// depending on where it sits, and every recursive call has to say which it means.
//
// The other half is short-circuit identity. `a || true` is constant even though `a` is not, because
// the right operand alone decides the result. That is why a logical expression is not simply "both
// sides constant".
//
// Deliberate narrowing from ESLint, same shape as for-direction's and recorded for the same reason:
// ESLint asks its scope analysis whether `undefined` and `Boolean` are the globals rather than
// local shadows. We treat the names as the globals. Shadowing either is vanishingly rare and doing
// it deliberately to defeat a lint rule is not a thing anyone does, so the exposure is a false
// positive on code that redefines `undefined`, which has larger problems.
var NoConstantCondition = rule.Rule{
	Name: "no-constant-condition",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		checkLoops := NoConstantConditionCheckLoops("allExceptWhileTrue")
		if parsed, ok := rule.OptionsAs[NoConstantConditionOptions](options); ok && parsed.CheckLoops != "" {
			checkLoops = parsed.CheckLoops
		}

		reportIfConstant := func(condition *ast.Node) {
			if condition != nil && isConstantExpression(condition, true) {
				ctx.ReportNode(condition, messageUnexpectedConstantCondition)
			}
		}

		// A loop condition is checked only when the option allows it, and `while(true)` is exempt
		// under the default.
		//
		// One piece of ESLint is deliberately not reproduced. It defers every loop report until the
		// loop exits, and clears the pending set on a `yield`, so a constant-condition loop inside
		// a generator is not reported: `while(true) { yield x; }` is a legitimate infinite
		// generator. Our default already exempts `while(true)`, which is the shape that idiom is
		// always written in, so the deferral only changes the answer for a generator looping on
		// some other constant, such as `while(1) { yield x; }`.
		//
		// Reported here and not by ESLint. That is a false positive rather than a miss, so it is
		// visible and arguable rather than silent, and it needs a generator and a non-`true`
		// constant together to occur. Recorded because it is a real divergence and this comment is
		// where to start if the differential ever shows one.
		checkLoopCondition := func(condition *ast.Node, isWhileTrue bool) {
			switch checkLoops {
			case "none":
				return
			case "allExceptWhileTrue":
				if isWhileTrue {
					return
				}
			}
			reportIfConstant(condition)
		}

		return rule.Listeners{
			ast.KindIfStatement: func(node *ast.Node) {
				reportIfConstant(node.AsIfStatement().Expression)
			},

			ast.KindConditionalExpression: func(node *ast.Node) {
				reportIfConstant(node.AsConditionalExpression().Condition)
			},

			ast.KindWhileStatement: func(node *ast.Node) {
				condition := node.AsWhileStatement().Expression
				checkLoopCondition(condition, isTrueLiteral(condition))
			},

			ast.KindDoStatement: func(node *ast.Node) {
				checkLoopCondition(node.AsDoStatement().Expression, false)
			},

			ast.KindForStatement: func(node *ast.Node) {
				// A for loop with no condition is `for(;;)`, which is the same deliberate idiom as
				// `while(true)` and has no condition node to report on anyway.
				checkLoopCondition(node.AsForStatement().Condition, false)
			},
		}
	},
}

// isTrueLiteral reports the literal `true`, which is the exempt loop idiom.
func isTrueLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node != nil && node.Kind == ast.KindTrueKeyword
}

// isConstantExpression reports whether an expression's value cannot vary.
//
// inBooleanPosition is the parameter that makes this rule work, and it changes the answer rather
// than merely relaxing it. In a condition, `[]` is constant because every array is truthy; as a
// value, `[a]` varies with a. So the flag has to be threaded through every recursion and set
// correctly at each step, and getting it wrong in either direction produces findings on correct
// code or silence on broken code.
func isConstantExpression(node *ast.Node, inBooleanPosition bool) bool {
	// A hole in a sparse array (`[, a]`) has no node, and ESLint treats it as constant: it is
	// always undefined.
	if node == nil {
		return true
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return true
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindArrowFunction, ast.KindFunctionExpression,
		ast.KindClassExpression, ast.KindObjectLiteralExpression:
		// An object or class is truthy whatever it holds. ESLint notes that a `toString` returning
		// a variable makes this technically wrong outside a boolean position, and opts not to
		// handle it; reproduced, since diverging would report code the gate does not.
		return true

	case ast.KindIdentifier:
		// `undefined` is the only identifier with a fixed value. See the rule comment on why this
		// does not check that it is the global one.
		return node.Text() == "undefined"

	case ast.KindTemplateExpression:
		return isConstantTemplate(node, inBooleanPosition)

	case ast.KindArrayLiteralExpression:
		// Truthy whatever it holds, so constant as a condition. As a value it is constant only if
		// every element is.
		if inBooleanPosition {
			return true
		}
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			if !isConstantExpression(element, false) {
				return false
			}
		}
		return true

	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindExclamationToken:
			// Negation reads its operand as a condition, so the flag flips to true rather than
			// being passed through.
			return isConstantExpression(unary.Operand, true)
		}
		return isConstantExpression(unary.Operand, false)

	case ast.KindVoidExpression:
		// `void anything` is undefined.
		return true

	case ast.KindTypeOfExpression:
		// typeof always yields a non-empty string, so it is truthy, but the string itself varies.
		if inBooleanPosition {
			return true
		}
		return isConstantExpression(node.AsTypeOfExpression().Expression, false)

	case ast.KindNewExpression:
		// A constructed object is truthy, and that is all that can be said about it.
		return inBooleanPosition

	case ast.KindBinaryExpression:
		return isConstantBinaryExpression(node, inBooleanPosition)

	case ast.KindCallExpression:
		return isConstantBooleanCall(node)

	case ast.KindSpreadElement:
		return isConstantExpression(node.AsSpreadElement().Expression, inBooleanPosition)
	}

	return false
}

// isConstantTemplate reports whether a template literal's value cannot vary.
//
// As a condition it is enough that one static chunk is non-empty, since the result is then a
// non-empty string whatever the substitutions produce. As a value every substitution must itself
// be constant.
func isConstantTemplate(node *ast.Node, inBooleanPosition bool) bool {
	template := node.AsTemplateExpression()

	if inBooleanPosition {
		if template.Head != nil && len(template.Head.Text()) > 0 {
			return true
		}
		for _, span := range template.TemplateSpans.Nodes {
			literal := span.AsTemplateSpan().Literal
			if literal != nil && len(literal.Text()) > 0 {
				return true
			}
		}
	}

	for _, span := range template.TemplateSpans.Nodes {
		if !isConstantExpression(span.AsTemplateSpan().Expression, false) {
			return false
		}
	}
	return true
}

// isConstantBinaryExpression handles the operators, including the short-circuit cases.
func isConstantBinaryExpression(node *ast.Node, inBooleanPosition bool) bool {
	binary := node.AsBinaryExpression()
	if binary.OperatorToken == nil {
		return false
	}
	operator := binary.OperatorToken.Kind

	switch operator {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
		leftConstant := isConstantExpression(binary.Left, inBooleanPosition)
		rightConstant := isConstantExpression(binary.Right, inBooleanPosition)

		// Short circuit: an operand that alone decides the result makes the whole expression
		// constant even when the other operand varies. `a || true` is always truthy.
		if leftConstant && isLogicalIdentity(binary.Left, operator) {
			return true
		}
		if inBooleanPosition && rightConstant && isLogicalIdentity(binary.Right, operator) {
			return true
		}
		return leftConstant && rightConstant

	case ast.KindCommaToken:
		// A sequence yields its last operand, so only that one decides.
		return isConstantExpression(binary.Right, inBooleanPosition)

	case ast.KindEqualsToken:
		// `if(a = 0)` is the classic typo this rule exists for: the condition is whatever was
		// assigned, so it is constant when the value is.
		return isConstantExpression(binary.Right, inBooleanPosition)

	case ast.KindBarBarEqualsToken:
		return inBooleanPosition && isLogicalIdentity(binary.Right, ast.KindBarBarToken)

	case ast.KindAmpersandAmpersandEqualsToken:
		return inBooleanPosition && isLogicalIdentity(binary.Right, ast.KindAmpersandAmpersandToken)

	case ast.KindInKeyword:
		// `in` reads a property that may or may not be there, so it varies even between two
		// constants. ESLint excludes it explicitly and so do we.
		return false

	case ast.KindInstanceOfKeyword:
		return false
	}

	// Every other binary operator produces a value determined by its operands, so it is constant
	// exactly when both are. The operands are values rather than conditions regardless of where
	// the expression itself sits.
	return isConstantExpression(binary.Left, false) && isConstantExpression(binary.Right, false)
}

// isLogicalIdentity reports whether a node alone decides a logical expression's result.
//
// `true` decides an `||` and `false` decides an `&&`. The recursive cases matter more than they
// look: `a && false || b` has a left operand that is an identity for `&&` and not for `||`, so the
// operator has to match at every level.
func isLogicalIdentity(node *ast.Node, operator ast.Kind) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindTrueKeyword:
		return operator == ast.KindBarBarToken
	case ast.KindFalseKeyword, ast.KindNullKeyword:
		return operator == ast.KindAmpersandAmpersandToken

	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		if len(node.Text()) > 0 {
			return operator == ast.KindBarBarToken
		}
		return operator == ast.KindAmpersandAmpersandToken

	case ast.KindNumericLiteral, ast.KindBigIntLiteral:
		if numericLiteralSign(node.Text()) == 0 {
			return operator == ast.KindAmpersandAmpersandToken
		}
		return operator == ast.KindBarBarToken

	case ast.KindRegularExpressionLiteral, ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindFunctionExpression:
		// Always truthy.
		return operator == ast.KindBarBarToken

	case ast.KindVoidExpression:
		// void yields undefined, which is falsy, so it decides an `&&`.
		return operator == ast.KindAmpersandAmpersandToken

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			return binary.OperatorToken.Kind == operator &&
				(isLogicalIdentity(binary.Left, operator) || isLogicalIdentity(binary.Right, operator))
		case ast.KindBarBarEqualsToken:
			return operator == ast.KindBarBarToken && isLogicalIdentity(binary.Right, operator)
		case ast.KindAmpersandAmpersandEqualsToken:
			return operator == ast.KindAmpersandAmpersandToken && isLogicalIdentity(binary.Right, operator)
		}
	}

	return false
}

// isConstantBooleanCall reports `Boolean(...)` with a constant or absent argument.
//
// The only call anyone can reason about without knowing what a function does. `Boolean()` is false
// and `Boolean(1)` is true; anything else is a call whose result nobody here can predict.
func isConstantBooleanCall(node *ast.Node) bool {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "Boolean" {
		return false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return true
	}
	return isConstantExpression(call.Arguments.Nodes[0], true)
}
