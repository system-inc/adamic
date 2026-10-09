package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoEmptyPatternOptions tunes whether an empty object pattern is allowed in parameter position.
type NoEmptyPatternOptions struct {
	// AllowObjectPatternsAsParameters permits `function foo({}) {}` and `({} = {}) => {}`, the
	// shape used to accept an options object and ignore it while keeping the arity.
	//
	// Deliberately narrow, following ESLint: it covers object patterns only, and only in parameter
	// position. An empty array pattern parameter still reports, because `[]` there destructures an
	// iterable and throws on a non-iterable argument, so it does real work and getting it wrong is
	// a runtime error rather than a no-op.
	AllowObjectPatternsAsParameters bool `json:"allowObjectPatternsAsParameters"`
}

var messageEmptyObjectPattern = rule.Message{
	Id: "unexpectedObject",
	Description: "This destructuring pattern is empty, so it binds nothing. It reads like an " +
		"object literal but is a pattern, which is why the mistake survives review: `const {} = foo` " +
		"declares no variables at all and only asserts that `foo` is not null. Name the properties " +
		"being pulled out, or drop the destructuring.",
}

var messageEmptyArrayPattern = rule.Message{
	Id: "unexpectedArray",
	Description: "This destructuring pattern is empty, so it binds nothing. `const [] = foo` " +
		"declares no variables and only forces `foo` to be iterated. Name the elements being pulled " +
		"out, or drop the destructuring.",
}

// NoEmptyPattern flags a destructuring pattern that declares no bindings.
//
//	valid:   const { a } = foo;
//	valid:   const { a = {} } = foo;
//	valid:   const [a = []] = foo;
//	invalid: const {} = foo;
//	invalid: const [] = foo;
//	invalid: const { a: {} } = foo;
//	invalid: function foo({}) {}
//
// An empty pattern is almost always a default value that was written in the wrong position. The two
// look identical at a glance, which is the whole reason the rule exists: `const { a = {} } = foo`
// gives `a` a default of `{}`, while `const { a: {} } = foo` uses `a` only as a location to descend
// through and creates nothing. The second is valid syntax that does no work, so nothing downstream
// complains.
//
// The two shapes carry different messages because they fail differently and a reader has to know
// which one they are looking at. An empty object pattern is inert apart from a null check; an empty
// array pattern still iterates its subject, so it throws on a non-iterable where the object form
// would not.
//
// AllowObjectPatternsAsParameters covers the one place an empty pattern is sometimes deliberate: a
// function that accepts an options object it does not read yet, where `({})` documents the shape and
// holds the arity. Off by default, matching ESLint, and it does not extend to array patterns.
//
// Where the option is on, ESLint still reports a parameter whose default is anything other than an
// empty object literal, and we match that. It looks fussy until you see what it separates:
// `({} = {})` is a parameter that may be omitted entirely, which is the pattern the option exists
// for, while `({} = bar)` reads `bar` at every call for a destructuring that binds nothing, so the
// only observable effect is whatever evaluating `bar` does. That is a bug wearing the option's
// clothes.
//
// Nested empty patterns report even in parameter position and even with the option on, because the
// option is about a parameter that takes no bindings, and `function foo({ a: {} })` is a parameter
// that does take a binding and then throws it away.
//
// No fix. Whether the author meant `{ a = {} }` or meant to name properties they never wrote is not
// recoverable from the code, and both repairs change what the function receives.
var NoEmptyPattern = rule.Rule{
	Name: "no-empty-pattern",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings := NoEmptyPatternOptions{}
		if configured, isConfigured := rule.OptionsAs[NoEmptyPatternOptions](options); isConfigured {
			settings = configured
		}

		return rule.Listeners{
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				if hasBindingElements(node) {
					return
				}
				if settings.AllowObjectPatternsAsParameters && isAllowedEmptyParameterPattern(node) {
					return
				}
				ctx.ReportNode(node, messageEmptyObjectPattern)
			},

			ast.KindArrayBindingPattern: func(node *ast.Node) {
				if hasBindingElements(node) {
					return
				}
				ctx.ReportNode(node, messageEmptyArrayPattern)
			},
		}
	},
}

// hasBindingElements says whether a binding pattern declares anything.
func hasBindingElements(node *ast.Node) bool {
	bindingPattern := node.AsBindingPattern()
	if bindingPattern == nil || bindingPattern.Elements == nil {
		return false
	}
	return len(bindingPattern.Elements.Nodes) > 0
}

// isAllowedEmptyParameterPattern says whether an empty object pattern sits in the one position the
// option permits.
//
// The pattern has to be the parameter itself rather than nested inside one, which is what reading
// only the immediate parent enforces: for `function foo({ a: {} })` the inner pattern's parent is a
// BindingElement, not a Parameter, so it falls through and reports. That single check is the whole
// nesting exemption, and writing it as a parent-kind test rather than an ancestor walk is what makes
// it correct.
func isAllowedEmptyParameterPattern(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindParameter {
		return false
	}
	parameter := parent.AsParameterDeclaration()
	if parameter == nil {
		return false
	}
	if parameter.Initializer == nil {
		return true
	}
	return isEmptyObjectLiteral(parameter.Initializer)
}

// isEmptyObjectLiteral says whether an expression is the literal `{}`.
func isEmptyObjectLiteral(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	objectLiteral := node.AsObjectLiteralExpression()
	if objectLiteral == nil || objectLiteral.Properties == nil {
		return true
	}
	return len(objectLiteral.Properties.Nodes) == 0
}
