package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageComparisonWithNaN = rule.Message{
	Id: "comparisonWithNaN",
	Description: "Comparing against NaN is always false, including `NaN === NaN`, because NaN is " +
		"the one value in the language that is not equal to itself. So this comparison does not " +
		"test what it reads as testing: it is a check for NaN that can never succeed, and the " +
		"branch behind it is dead. Use Number.isNaN(value) instead, which is the only reliable " +
		"test. Note the Number. prefix: the global isNaN coerces its argument first, so " +
		"isNaN('hello') is true.",
}

// comparisonOperators are the operators for which a NaN operand makes the result constant.
//
// Relational operators are included alongside equality. `x < NaN` is always false too, and writing
// NaN there signals an intent the operator cannot carry just as much as `x === NaN` does.
var comparisonOperators = map[ast.Kind]bool{
	ast.KindLessThanToken:                true,
	ast.KindLessThanEqualsToken:          true,
	ast.KindGreaterThanToken:             true,
	ast.KindGreaterThanEqualsToken:       true,
	ast.KindEqualsEqualsToken:            true,
	ast.KindEqualsEqualsEqualsToken:      true,
	ast.KindExclamationEqualsToken:       true,
	ast.KindExclamationEqualsEqualsToken: true,
}

// UseIsNaNOptions mirrors the two switches ESLint's rule carries.
//
// EnforceForSwitchCase defaults to true, matching both ESLint 9 and the live config for this tree.
// A `case NaN:` never matches for the same reason `=== NaN` is never true, so the default is the
// one that catches the defect rather than the one that is quieter.
//
// EnforceForIndexOf reports `list.indexOf(NaN)` and `list.lastIndexOf(NaN)`, which always return -1
// because both search with strict equality. Default false, as upstream's. It was decoded and never
// read until #6esg2nx, so a config turning it on changed nothing.
type UseIsNaNOptions struct {
	EnforceForSwitchCase *bool `json:"enforceForSwitchCase"`
	EnforceForIndexOf    bool  `json:"enforceForIndexOf"`
}

// buildIndexOfNaNMessage is upstream's indexOfNaN, naming the method.
func buildIndexOfNaNMessage(methodName string) rule.Message {
	replacement := "findIndex"
	if methodName == "lastIndexOf" {
		replacement = "findLastIndex"
	}
	return rule.Message{
		Id: "indexOfNaN",
		Description: "`" + methodName + "` cannot find NaN, so this call returns -1 whatever the array " +
			"holds. It searches with strict equality, and NaN is not equal to itself. Search with " +
			"Number.isNaN instead: `" + replacement + "(Number.isNaN)`.",
	}
}

var messageCaseWithNaN = rule.Message{
	Id: "caseWithNaN",
	Description: "This `case NaN:` can never match. A switch compares with strict equality, and " +
		"NaN is not equal to itself, so the branch is dead exactly as `value === NaN` would be. " +
		"Test for it before the switch with Number.isNaN(value), since there is no case label " +
		"that can catch it.",
}

var messageSwitchOnNaN = rule.Message{
	Id: "switchOnNaN",
	Description: "This switch tests NaN against each case, and NaN equals nothing including " +
		"itself, so every case is dead and only the default runs. Whatever this switch was " +
		"written to dispatch on, it dispatches on nothing.",
}

// UseIsNaN flags a comparison against NaN.
//
//	valid:   Number.isNaN(value)
//	valid:   const NaN = 1; value === NaN   (a local is not the global; see isNaNReference)
//	invalid: value === NaN
//	invalid: NaN !== value
//	invalid: value < NaN
//
// The failure is total rather than partial: every comparison against NaN evaluates to false (or true
// for `!==`), so the branch is unreachable and the check silently does nothing. That makes this one
// of the few rules where a finding is almost certainly a real defect rather than a style preference.
//
// `Number.NaN` is the same value reached through a property and is reported the same way.
//
// No fix. Rewriting `x === NaN` to `Number.isNaN(x)` is right for equality and wrong for the
// relational operators, where there is no equivalent, and the negated form needs `!Number.isNaN(x)`.
// Getting that wrong changes a branch that never runs into one that runs sometimes, which is a
// behavior change wearing a lint fix.
var UseIsNaN = rule.Rule{
	Name: "use-isnan",
	// Asked only whether a name spelled NaN or Number is declared in this file, so its findings key
	// on the closure's shapes. A NaN anywhere is rare, so the checker is consulted rarely.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Defaults to on, matching ESLint 9 and this tree's configuration. An option that only relaxes the
		// rule gets the strict reading when the config says nothing, so a misconfiguration cannot
		// quietly disable half the rule.
		enforceForSwitchCase := true
		settings, hasSettings := rule.OptionsAs[UseIsNaNOptions](options)
		if hasSettings && settings.EnforceForSwitchCase != nil {
			enforceForSwitchCase = *settings.EnforceForSwitchCase
		}

		return rule.Listeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				if !enforceForSwitchCase {
					return
				}

				statement := node.AsSwitchStatement()

				// Upstream's checkSwitchStatement, spans included: switchNaN on the whole statement and
				// caseNaN on the whole clause, and a NaN case under a NaN switch is reported as well,
				// since each is its own dead code. Measured on ESLint 10.8.1's corpus (#9g0v4j6).
				if isNaNReference(ctx, statement.Expression) {
					ctx.ReportNode(node, messageSwitchOnNaN)
				}

				if statement.CaseBlock == nil {
					return
				}
				for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
					if clause.Kind != ast.KindCaseClause {
						continue
					}
					if isNaNReference(ctx, clause.AsCaseOrDefaultClause().Expression) {
						ctx.ReportNode(clause, messageCaseWithNaN)
					}
				}
			},

			// Upstream's checkCallExpression: a method named `indexOf` or `lastIndexOf`, by a static
			// name (`.indexOf`, `['indexOf']`, `` [`indexOf`] ``, through `?.` and parentheses), called
			// with one or two arguments, the first of them NaN. A third argument is not the array
			// method's signature and is left alone, and so is a spread, which may not be NaN at all.
			// ESLint 10.8.1's 73 rows for this option replay through here; see the test.
			ast.KindCallExpression: func(node *ast.Node) {
				if !settings.EnforceForIndexOf {
					return
				}
				call := node.AsCallExpression()
				methodName := staticMethodName(ast.SkipParentheses(call.Expression))
				if methodName != "indexOf" && methodName != "lastIndexOf" {
					return
				}
				arguments := call.Arguments.Nodes
				if len(arguments) == 0 || len(arguments) > 2 || !isNaNReference(ctx, arguments[0]) {
					return
				}
				ctx.ReportNode(node, buildIndexOfNaNMessage(methodName))
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil || !comparisonOperators[binary.OperatorToken.Kind] {
					return
				}
				if isNaNReference(ctx, binary.Left) || isNaNReference(ctx, binary.Right) {
					ctx.ReportNode(node, messageComparisonWithNaN)
				}
			},
		}
	},
}

// staticMethodName is the name a member callee reads by, when the source spells it out: `.name`, or a
// string or substitution-free template in brackets. Anything computed answers "".
func staticMethodName(callee *ast.Node) string {
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		if name := callee.Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	case ast.KindElementAccessExpression:
		argument := ast.SkipParentheses(callee.AsElementAccessExpression().ArgumentExpression)
		if argument != nil && (argument.Kind == ast.KindStringLiteral || argument.Kind == ast.KindNoSubstitutionTemplateLiteral) {
			return argument.Text()
		}
	}
	return ""
}

// isNaNReference reports whether an expression names the global NaN, bare or through Number.
//
// A local named NaN or Number is not the global, and upstream's sourceCode.isGlobalReference stays
// silent on it: `function f(NaN) { return x === NaN; }` and `let Number; x === Number.NaN` are clean
// in ESLint 10.8.1's corpus. This rule matched by spelling and reported both, an accepted false
// positive until #9g0v4j6 asked the checker instead, through identifierIsShadowed: a name declared
// in source is a shadow, and one declared only in a declaration file, as the lib's NaN and Number
// are, is the global. Without a checker the answer is the global, which reports.
func isNaNReference(ctx rule.Context, node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text() == "NaN" && !identifierIsShadowed(ctx, node)

	case ast.KindPropertyAccessExpression:
		// `Number.NaN` is the same value spelled through the constructor.
		access := node.AsPropertyAccessExpression()
		if access.Expression == nil || access.Expression.Kind != ast.KindIdentifier {
			return false
		}
		if access.Expression.Text() != "Number" || identifierIsShadowed(ctx, access.Expression) {
			return false
		}
		name := access.Name()
		return name != nil && name.Text() == "NaN"

	case ast.KindElementAccessExpression:
		// `Number['NaN']` is the same read in brackets, which upstream's isSpecificMemberAccess
		// accepts through its static property name.
		access := node.AsElementAccessExpression()
		if access.Expression == nil || access.Expression.Kind != ast.KindIdentifier || access.Expression.Text() != "Number" ||
			identifierIsShadowed(ctx, access.Expression) {
			return false
		}
		argument := ast.SkipParentheses(access.ArgumentExpression)
		return argument != nil && (argument.Kind == ast.KindStringLiteral || argument.Kind == ast.KindNoSubstitutionTemplateLiteral) &&
			argument.Text() == "NaN"

	case ast.KindParenthesizedExpression:
		return isNaNReference(ctx, node.AsParenthesizedExpression().Expression)

	case ast.KindBinaryExpression:
		// A sequence's value is its last expression, so `(sideEffect(), NaN)` is NaN. Upstream's
		// isNaNIdentifier reads a SequenceExpression's last element for every check, the comparison
		// and the switch as well as indexOf.
		binary := node.AsBinaryExpression()
		return binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindCommaToken &&
			isNaNReference(ctx, binary.Right)
	}
	return false
}
