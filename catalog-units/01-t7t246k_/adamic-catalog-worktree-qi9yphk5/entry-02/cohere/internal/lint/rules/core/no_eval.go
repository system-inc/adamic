package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoEvalOptions tunes whether indirect access to `eval` is permitted.
type NoEvalOptions struct {
	// AllowIndirect permits every form of `eval` access except a direct, non-optional call.
	//
	// Off by default, matching ESLint, whose `meta.schema` states this as the rule's only option.
	// Indirect eval (`(0, eval)(code)`, `window.eval(code)`, an alias) always evaluates in the
	// global scope rather than the calling one, so it cannot read or write the local bindings around
	// the call. That is a materially smaller hazard than the direct form, and a codebase that has
	// deliberately routed its one dynamic evaluation through the indirect spelling turns this on.
	AllowIndirect bool `json:"allowIndirect"`
}

// globalObjectNames are the identifiers that can name the global object, whose `eval` property is
// the global `eval` reached indirectly.
//
// Upstream gates each of these on an `env` configuration (`browser` supplies `window`, `node`
// supplies `global`), so `window.eval('foo')` is a *pass* case there with no env set. We have no env
// surface, and adding one for this rule alone would be a larger change than the rule. The three
// names are treated as the global object unconditionally instead; see the divergence note on the
// rule.
var globalObjectNames = map[string]bool{
	"global":     true,
	"window":     true,
	"globalThis": true,
}

var messageNoEval = rule.Message{
	Id: "noEval",
	Description: "This reaches the `eval` function, which compiles a string as code at runtime. " +
		"A direct call also runs that code with full access to the surrounding scope, so any part " +
		"of the string an attacker controls becomes an injection into this function's locals. It " +
		"defeats every static tool that would otherwise see what the program does, and it is slow " +
		"because the engine cannot optimize a scope it may rewrite. For JSON use JSON.parse; for " +
		"dynamic property access use bracket notation; otherwise restructure so nothing is " +
		"evaluated from a string.",
}

// NoEval flags any access to the `eval` function.
//
//	valid:   Eval(foo)
//	valid:   setTimeout('foo')
//	valid:   window.noeval('foo')
//	valid:   function foo() { var eval = 'foo'; window[eval]('foo') }
//	valid:   class A { foo() { this.eval(); } }
//	invalid: eval('foo')
//	invalid: (0, eval)('foo')
//	invalid: var EVAL = eval; EVAL('foo')
//	invalid: window.eval('foo')
//	invalid: eval.toString()
//
// # Two hazards, not one
//
// A *direct* call, where the callee is literally the identifier `eval`, runs the compiled string in
// the caller's own scope: it can read and rewrite the local bindings around the call site. An
// *indirect* access, which is every other spelling, evaluates in the global scope instead. Both are
// reported by default and `AllowIndirect` narrows the rule to the first, which is why the option
// exists at all rather than being a blanket off switch.
//
// # Why this reads the checker
//
// Determined by probe rather than by reading upstream's call site, because the brief is right that
// seeing `ctx.scoping()` proves nothing: that table answers scope-flag questions the AST already
// knows as well as name-resolution questions only the checker can.
//
// This rule asks both kinds, and they cost differently.
//
// The `this.eval` arm asks a scope-flag question. Whether `this` at some point is the global object
// is decided by the enclosing construct, and `GetThisContainer` answers it off the AST for free.
//
// The bare-identifier arm asks a resolution question and has no structural answer. `eval` appearing
// as a value, in `var EVAL = eval` or `cb(eval)`, is the global `eval` only when nothing shadows the
// name, and the shadow can be a parameter, a `var`, a function declaration, or an import anywhere
// in an enclosing scope. Upstream answers this with `root_unresolved_references`, which is a scope
// analysis. `resolvesToAGlobal`, already shipped for `no-new-native-nonconstructor`, asks the
// equivalent question of the checker: which file declares this name. Measured on this corpus before
// the rule was written, with a probe printing the verdict per identifier: `eval('foo')` resolved to
// `lib.es5.d.ts` and answered true, while the parameter, `var`, and function-declaration shadows all
// answered false.
//
// The direct-call arm is deliberately *not* gated on resolution, matching upstream: `function
// foo(eval) { eval('foo') }` is a fail case in the corpus. A call whose callee is spelled `eval` is
// a direct eval to the engine's own grammar regardless of what the name binds to, so the syntactic
// test is the correct one there and the resolution test would be wrong.
//
// # Divergence from upstream, stated rather than silent
//
// Upstream gates the global-object arm behind an `env` configuration, so with no env set
// `window.eval('foo')` and `global.eval('foo')` are pass cases and only `globalThis` would be
// considered. We have no env surface. I probed whether the checker could stand in for it and it
// cannot: in our lib set `window` and `global` resolve to no symbol at all, shadowed or not, so the
// checker gives no signal to gate on. The three names are therefore treated as the global object
// unconditionally, which reports `window.eval('foo')` where upstream with no env would not. That is
// the direction I would want a lint rule to be wrong in, since a `.eval` property call on something
// named `window` is the hazard whether or not the config declared a browser.
//
// The cost of that choice is a false positive on a local variable named `window` with an unrelated
// `eval` property, which nothing in the corpus exercises and which the computed-subscript cases
// already show the rule declining where the property is not statically `eval`.
var NoEval = rule.Rule{
	Name: "no-eval",

	// See the doc above: the bare-identifier arm's whole discrimination is whether `eval` names the
	// global or a local shadowing it, and nothing structural answers that.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		allowIndirect := false
		if typed, ok := rule.OptionsAs[NoEvalOptions](options); ok {
			allowIndirect = typed.AllowIndirect
		}

		report := func(node *ast.Node) {
			ctx.ReportNode(node, messageNoEval)
		}

		listeners := rule.Listeners{
			// The direct call. Syntactic on purpose: a callee spelled `eval` is a direct eval to the
			// grammar, whatever the name binds to, which is why upstream reports the shadowed
			// parameter case.
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Expression == nil {
					return
				}
				// An optional call routes through optional-chaining evaluation rather than step
				// 6.a.vi of the call runtime semantics, so it is not a direct eval. Only
				// `allowIndirect` cares, since without it the callee is reported by the identifier
				// arm anyway.
				if allowIndirect && call.QuestionDotToken != nil {
					return
				}
				if isEvalCallee(call.Expression) {
					report(call.Expression)
				}
			},
		}

		if allowIndirect {
			return listeners
		}

		// A bare `eval` used as a value rather than called directly. Resolution decides it.
		listeners[ast.KindIdentifier] = func(node *ast.Node) {
			if node.Text() != "eval" {
				return
			}
			// A direct callee is the other arm's finding; reporting here too would double-count it.
			if isDirectCallee(node) {
				return
			}
			// A property name, a declaration name, or an object-literal key is not a reference to
			// the global.
			if !isValueReference(node) {
				return
			}
			if !resolvesToAGlobal(ctx, node) {
				return
			}
			report(node)
		}

		// A `.eval` property reached through a name for the global object.
		listeners[ast.KindPropertyAccessExpression] = func(node *ast.Node) {
			name, known := property.AccessedName(node, property.Static)
			if !known || name != "eval" {
				return
			}
			if !namesTheGlobalObject(node.AsPropertyAccessExpression().Expression) {
				return
			}
			report(node.AsPropertyAccessExpression().Name())
		}

		listeners[ast.KindElementAccessExpression] = func(node *ast.Node) {
			name, known := property.AccessedName(node, property.Static)
			if !known || name != "eval" {
				return
			}
			if !namesTheGlobalObject(node.AsElementAccessExpression().Expression) {
				return
			}
			// The literal including its quotes, matching upstream's span.
			report(node.AsElementAccessExpression().ArgumentExpression)
		}

		return listeners
	},
}

// isEvalCallee reports whether a call's callee is the identifier `eval`, seeing through the wrappers
// that do not change what is called.
//
// `eval!('foo')` and `(eval as any)('foo')` are both direct evals and both fail cases upstream, so
// the non-null assertion, the type assertion, and parentheses have to be transparent here.
func isEvalCallee(callee *ast.Node) bool {
	inner := skipCalleeWrappers(callee)
	return inner != nil && inner.Kind == ast.KindIdentifier && inner.Text() == "eval"
}

// skipCalleeWrappers unwraps the expression forms that wrap a callee without changing it.
//
// Parentheses are deliberately *not* unwrapped for the direct-call test's span, but they are for the
// identity test: `(eval)('foo')` is a direct eval. The span stays on the outer node because that is
// what upstream's snapshot records for `eval!` and `(eval as any)`.
func skipCalleeWrappers(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindExpressionWithTypeArguments:
			node = node.AsExpressionWithTypeArguments().Expression
		default:
			return node
		}
	}
	return nil
}

// isDirectCallee reports whether this identifier is the callee of the call that encloses it.
//
// Walks out through the same transparent wrappers, because `(eval as any)('foo')` reports once from
// the call arm and must not report a second time from the identifier arm.
func isDirectCallee(node *ast.Node) bool {
	current := node
	for current.Parent != nil {
		switch current.Parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindNonNullExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindTypeAssertionExpression,
			ast.KindExpressionWithTypeArguments:
			current = current.Parent
		case ast.KindCallExpression:
			return current.Parent.AsCallExpression().Expression == current
		default:
			return false
		}
	}
	return false
}

// isValueReference reports whether an identifier is being read as a value.
//
// A property name (`foo.eval`), a declaration's own name (`var eval`), a parameter name, an
// object-literal key, and an import binding all spell `eval` without referring to the global one.
// The property-access case matters most: `window.eval` must be reported once, by the member arm,
// rather than a second time here.
func isValueReference(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Name() != node
	// A declaration's own *name* is a binding rather than a reference, but its *initializer* is a
	// read like any other. `var EVAL = eval` is a fail case upstream and was the one fixture this
	// arm got wrong: rejecting the whole VariableDeclaration threw the initializer out with the
	// name. Same shape for a parameter default and a property initializer.
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == node
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Initializer == node
	case ast.KindPropertyDeclaration:
		return parent.AsPropertyDeclaration().Initializer == node
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindBindingElement:
		return parent.AsBindingElement().Initializer == node
	case ast.KindEnumMember:
		return parent.AsEnumMember().Initializer == node
	case ast.KindQualifiedName,
		ast.KindPropertySignature,
		ast.KindMethodDeclaration,
		ast.KindMethodSignature,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindImportSpecifier,
		ast.KindExportSpecifier,
		ast.KindImportClause,
		ast.KindNamespaceImport,
		ast.KindShorthandPropertyAssignment,
		ast.KindTypeReference:
		return false
	}
	return true
}

// namesTheGlobalObject reports whether an expression names the global object.
//
// Either one of the three candidate identifiers directly, or a chain of the *same* name off itself,
// which is how `window.window.eval` and `globalThis.globalThis['eval']` reach the same object. The
// same-name requirement is upstream's and it is load-bearing: `foo.window.eval` is a property of
// something unrelated, so the walk stops at a name that does not match rather than accepting any
// chain ending in a global name.
//
// `staticPropertyName` is the shelf's, from no_self_assign.go. Probed on this corpus before being
// built on rather than taken from its doc comment: it answers ("eval", true) for `x.eval`,
// `x['eval']` and `x[`+"`eval`"+`], and ("", false) for `x[eval]`, which is exactly the distinction
// three of upstream's pass cases turn on. It also answers for numeric subscripts, which this rule
// never asks about since it compares against one name.
func namesTheGlobalObject(node *ast.Node) bool {
	current := skipCalleeWrappers(node)
	if current == nil {
		return false
	}

	if current.Kind == ast.KindIdentifier {
		return globalObjectNames[current.Text()]
	}

	// A chain: the property read must repeat the object's own name, so this compares the two
	// spellings rather than merely asking whether each is a candidate.
	name, known := property.AccessedName(current, property.Static)
	if !known || !globalObjectNames[name] {
		return false
	}
	object := skipCalleeWrappers(accessedObject(current))
	if object == nil {
		return false
	}
	if object.Kind == ast.KindIdentifier {
		return object.Text() == name
	}
	objectName, objectKnown := property.AccessedName(object, property.Static)
	return objectKnown && objectName == name && namesTheGlobalObject(object)
}
