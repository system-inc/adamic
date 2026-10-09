package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnusedExpression = rule.Message{
	Id: "unusedExpression",
	Description: "This expression is evaluated and then thrown away. Reading a property, comparing " +
		"two values, or naming a variable on a line of its own computes something the program " +
		"immediately forgets, so the line cannot affect what the code does and is almost always a " +
		"typo for something that would: a missing call, a missing `await`, or an assignment whose " +
		"left side was dropped. Use the value or delete the line.",
}

// NoUnusedExpressionsOptions configures which expression forms are tolerated in statement position.
//
// The five names are ESLint's `meta.schema` verbatim. Our rule inventory records this rule as taking
// no options, which is wrong, and the schema is the authoritative surface.
type NoUnusedExpressionsOptions struct {
	// AllowShortCircuit tolerates `a && b()`, where the logical operator is being used for control
	// flow rather than for its value. The right operand still has to be a used form, so `a && b`
	// keeps reporting.
	AllowShortCircuit bool `json:"allowShortCircuit"`

	// AllowTernary tolerates `a ? b() : c()`, the conditional spelling of the same idea. Both
	// branches have to be used forms, because either one can be the branch that runs.
	AllowTernary bool `json:"allowTernary"`

	// AllowTaggedTemplates tolerates ``tag`text` ``, where the tag is a call and may well have an
	// effect. The untagged template `` `text` `` is unaffected and keeps reporting, since it only
	// ever builds a string.
	AllowTaggedTemplates bool `json:"allowTaggedTemplates"`

	// EnforceForJSX extends the rule to `<div />` in statement position, which is off by default in
	// both upstreams because an element expression is a common intermediate while editing.
	EnforceForJSX bool `json:"enforceForJSX"`

	// IgnoreDirectives is inert in this port and deliberately so.
	//
	// ESLint runs two directive checks: one reading the parser's own `.directive` flag, applied
	// always, and a structural re-derivation applied only under this option. The two agree on every
	// input, so the option changes nothing there either. This port has the structural check alone
	// and applies it unconditionally, which reaches the same verdict by one road instead of two.
	// The field exists so that a config writing the name is accepted rather than rejected.
	IgnoreDirectives bool `json:"ignoreDirectives"`
}

// NoUnusedExpressions reports expressions in statement position whose value is discarded.
//
// Valid:
//
//	a = b;
//	f();
//	new Thing();
//	i++;
//	delete foo.bar;
//	void doSomething();
//	'use strict';
//
// Invalid:
//
//	a;
//	foo.bar;
//	a + b;
//	f(), 0;
//	`untagged template`;
//	Map<string, string>;
//	foo!;
//
// The judgment is upstream's and so is the shape: one listener on expression statements, a recursive
// verdict on the expression, and an exemption for a real directive prologue.
//
// The directive exemption is the one piece with no counterpart in oxc's rule file. oxc's parser
// hoists a directive prologue into a separate vector on the enclosing body, so a real directive
// never reaches its listener at all and its `ignore_directives` option is unreachable. TypeScript's
// parser leaves directives as ordinary expression statements, so without the exemption written here
// every `'use client'` at the top of a component file would report.
var NoUnusedExpressions = rule.Rule{
	Name: "@typescript-eslint/no-unused-expressions",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		verdict := unusedExpressionVerdict{}
		if parsed, ok := rule.OptionsAs[NoUnusedExpressionsOptions](options); ok {
			verdict.options = parsed
		}

		return rule.Listeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				if !verdict.isDisallowed(node.Expression()) {
					return
				}
				if isDirectivePrologueStatement(node) {
					return
				}
				ctx.ReportNode(node, messageUnusedExpression)
			},
		}
	},
}

// unusedExpressionVerdict answers whether one expression's value goes unused.
type unusedExpressionVerdict struct {
	options NoUnusedExpressionsOptions
}

// isDisallowed is upstream's `Checker.isDisallowed`, which reads inverted from its name: it answers
// true for the forms that have *no* effect and are therefore not allowed to stand alone.
//
// Upstream's default arm is "allowed", so an unrecognised node is left alone rather than reported.
// That default is load-bearing and is reproduced by the closing `return false` rather than by
// enumerating every used form: a node kind this port has never seen should go unreported, not
// reported on a guess.
func (verdict unusedExpressionVerdict) isDisallowed(expression *ast.Node) bool {
	switch expression.Kind {
	// Values with no effect at all. Every one of these is oxc's always-disallowed list, minus the
	// forms TypeScript spells with a different node.
	case ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindIdentifier,
		ast.KindThisKeyword,
		ast.KindSuperKeyword,
		ast.KindMetaProperty,
		ast.KindArrayLiteralExpression,
		ast.KindObjectLiteralExpression,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindClassExpression,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression:
		return true

	// Forms whose evaluation is the point. Listed rather than left to the default so that reading
	// this switch tells you what upstream considers used, and so that a later reader moving one out
	// of the default arm has to say why.
	//
	// A call covers more ground here than upstream's list suggests, because TypeScript has no
	// ChainExpression node: `a?.b()` is a call at the statement level, and `import('./foo')` is a
	// call too. oxc needs a ChainElement arm to reach the same answers.
	case ast.KindCallExpression,
		ast.KindNewExpression,
		ast.KindAwaitExpression,
		ast.KindYieldExpression,
		ast.KindPostfixUnaryExpression,
		ast.KindDeleteExpression,
		ast.KindVoidExpression,
		ast.KindSatisfiesExpression:
		return false

	// A prefix unary is `!a` and `+a`, which are disallowed, but also `++a`, which is an update and
	// is not. TypeScript gives all three one kind and puts the difference in the operator.
	case ast.KindPrefixUnaryExpression:
		operator := expression.AsPrefixUnaryExpression().Operator
		return operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken

	// `typeof a` is a unary that reads a value and discards it, which upstream reports. It is a
	// separate kind here rather than an operator on the unary node.
	case ast.KindTypeOfExpression:
		return true

	// The four-way collision. oxc and ESLint each get a distinct node type for assignment, sequence,
	// logical and arithmetic; TypeScript gives all four KindBinaryExpression and puts the
	// distinction in the operator token, so the discrimination has to happen here or every `a = b`
	// in the tree reports.
	case ast.KindBinaryExpression:
		binary := expression.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		switch {
		// An assignment has an effect. This arm has to be tested before the logical one, though the
		// two predicates turn out to be disjoint: `&&=` is an assignment operator and is not a
		// logical operator, measured rather than assumed.
		case ast.IsAssignmentOperator(operator):
			return false
		case ast.IsLogicalOrCoalescingBinaryOperator(operator):
			if verdict.options.AllowShortCircuit {
				return verdict.isDisallowed(binary.Right)
			}
			return true
		// A sequence and an arithmetic comparison are both always disallowed upstream, by separate
		// arms that reach the same answer. `f(), 0` reports as a whole even though `f()` is used.
		default:
			return true
		}

	case ast.KindConditionalExpression:
		if verdict.options.AllowTernary {
			conditional := expression.AsConditionalExpression()
			return verdict.isDisallowed(conditional.WhenTrue) || verdict.isDisallowed(conditional.WhenFalse)
		}
		return true

	case ast.KindTaggedTemplateExpression:
		return !verdict.options.AllowTaggedTemplates

	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return verdict.options.EnforceForJSX

	// The wrappers that carry no runtime meaning: parentheses and the type-only assertions. Each
	// hands the question to what it wraps, which is why `(a())` stays silent while `(a)` reports.
	case ast.KindParenthesizedExpression,
		ast.KindAsExpression,
		ast.KindTypeAssertionExpression,
		ast.KindNonNullExpression,
		ast.KindExpressionWithTypeArguments:
		return verdict.isDisallowed(expression.Expression())
	}

	return false
}

// isDirectivePrologueStatement answers whether a statement is a real directive in its position.
//
// This is ESLint's pair of `isTopLevelExpressionStatement` and `directives`. The position half is
// hand-written because no shelf function answers it. `ast.IsPrologueDirective` is exactly
// `Kind == KindExpressionStatement && Expression().Kind == KindStringLiteral` and nothing more,
// verified by reading its body rather than trusting its name: it is positionally blind, so it
// answers true for a string in the middle of a function body and for a string in a class static
// block, both of which upstream reports. It is used below for the shape test alone, which is all it
// can honestly answer, and the position is decided here. A port that took it for the whole question
// would silence eight of the corpus's fail cases.
//
// A template literal is deliberately not a string literal for this purpose, matching both upstreams
// and the runtime: a backticked `use strict` at the top of a function does not enable strict mode.
//
// Two conditions, per the specification's directive prologue:
//
// The container must be one that has a prologue at all: a source file, a module block, or a block
// belonging to something function-like. A class static block is a block whose parent is *not*
// function-like, which is what makes upstream report both strings in one.
//
// And the statement must be inside the unbroken leading run of string statements. The first
// non-string ends the prologue, so a string after a declaration is an ordinary unused expression.
func isDirectivePrologueStatement(node *ast.Node) bool {
	container := node.Parent
	if container == nil {
		return false
	}

	switch container.Kind {
	case ast.KindSourceFile, ast.KindModuleBlock:
	case ast.KindBlock:
		if !ast.IsFunctionLike(container.Parent) {
			return false
		}
	default:
		return false
	}

	// Upstream takes the leading run of string statements and asks whether this node is in it, so
	// the string test comes first: a node that is the container's very first statement is still not
	// a directive unless it is itself a string.
	for _, statement := range container.Statements() {
		if !ast.IsPrologueDirective(statement) {
			return false
		}
		if statement == node {
			return true
		}
	}
	return false
}
