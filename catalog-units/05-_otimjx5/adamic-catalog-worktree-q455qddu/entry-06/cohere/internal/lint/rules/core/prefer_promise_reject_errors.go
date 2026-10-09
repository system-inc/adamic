package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

const messagePreferPromiseRejectErrorsId = "rejectAnError"

// messagePreferPromiseRejectErrors is upstream's single message, which carries no computed text.
//
// `rule.Message` is `{Id, Description}` with no interpolation layer, so there is nothing to render
// and the tests assert the id and this description directly.
var messagePreferPromiseRejectErrors = rule.Message{
	Id: messagePreferPromiseRejectErrorsId,
	Description: "Expected the Promise rejection reason to be an Error. A rejection carrying a " +
		"string or a plain object arrives at the catch with no stack, so the line that failed is " +
		"lost and the handler cannot tell one failure from another.",
}

// PreferPromiseRejectErrorsOptions is upstream's one option.
//
// It defaults to FALSE, which is what makes the generic decoder safe here: a rule handed nil
// options gets the zero value, and the zero value is the default. That is the opposite of the
// default-true trap, and it is worth saying out loud because reaching for a pointer field by habit
// would add a decoder with nothing to decide.
type PreferPromiseRejectErrorsOptions struct {
	// AllowEmptyReject permits a bare `Promise.reject()` with no argument.
	AllowEmptyReject bool `json:"allowEmptyReject"`
}

// PreferPromiseRejectErrors reports a Promise rejected with something that cannot be an Error.
//
// Two shapes report, and they are separate listeners because they find their call sites in
// completely different ways.
//
//	Promise.reject(5)                                  a static call on the global
//	new Promise((resolve, reject) => reject(5))         the executor's reject callback
//
// Valid:
//
//	Promise.reject(new Error('failed'))
//	Promise.reject(foo)
//	new Promise((resolve, reject) => reject(new Error()))
//
// Invalid:
//
//	Promise.reject('failed')
//	Promise.reject()
//	new Promise((resolve, reject) => reject({ code: 1 }))
var PreferPromiseRejectErrors = rule.Rule{
	Name:             "prefer-promise-reject-errors",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Declining the file once rather than once per node. `NeedsTypeChecker` governs the
		// registration path only, and the harness builds a context by hand, so this is the guard
		// that actually runs.
		if ctx.TypeChecker == nil {
			return nil
		}

		settings, _ := rule.OptionsAs[PreferPromiseRejectErrorsOptions](options)

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if preferPromiseRejectErrorsIsGlobalPromiseReject(ctx, node) {
					preferPromiseRejectErrorsCheckCall(ctx, node, settings)
				}
			},
			ast.KindNewExpression: func(node *ast.Node) {
				preferPromiseRejectErrorsCheckExecutor(ctx, node, settings)
			},
		}
	},
}

// preferPromiseRejectErrorsCheckCall reports a rejection whose argument cannot be an Error.
//
// Upstream's `checkRejectCall`, and the argument order matters: `allowEmptyReject` only exempts the
// no-argument case, so `Promise.reject(undefined)` still reports under it. The corpus pins that
// with a case carrying `allowEmptyReject: true` in the INVALID list.
func preferPromiseRejectErrorsCheckCall(
	ctx rule.Context,
	call *ast.Node,
	settings PreferPromiseRejectErrorsOptions,
) {
	arguments := call.Arguments()

	if len(arguments) == 0 {
		if settings.AllowEmptyReject {
			return
		}
		ctx.ReportNode(call, messagePreferPromiseRejectErrors)
		return
	}

	reason := arguments[0]
	if !preferPromiseRejectErrorsCouldBeError(reason) ||
		preferPromiseRejectErrorsIsGlobalUndefined(ctx, reason) {
		ctx.ReportNode(call, messagePreferPromiseRejectErrors)
	}
}

// preferPromiseRejectErrorsCheckExecutor finds the reject callback of a `new Promise` and checks
// every call of it.
//
// Upstream reaches the call sites through the scope manager's resolved-reference index, which this
// tree does not have. The checker answers the same question one identifier at a time, so the walk
// asks `GetSymbolAtLocation` at each identifier in callee position inside the executor.
func preferPromiseRejectErrorsCheckExecutor(
	ctx rule.Context,
	node *ast.Node,
	settings PreferPromiseRejectErrorsOptions,
) {
	callee := node.Expression()
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "Promise" ||
		!preferPromiseRejectErrorsResolvesToTheGlobalPromise(ctx, callee) {
		return
	}

	arguments := node.Arguments()
	if len(arguments) == 0 || !preferPromiseRejectErrorsIsFunction(arguments[0]) {
		return
	}

	executor := arguments[0]
	parameters := executor.Parameters()
	if len(parameters) < 2 {
		return
	}

	// Upstream requires the second parameter to be an Identifier, and the corpus pins BOTH
	// directions: `function(resolve, {apply}) { apply(5) }` is clean because a destructured second
	// parameter is not tracked, while `({foo, bar, baz}, reject) => reject(5)` reports because the
	// destructure is on the FIRST parameter.
	//
	// The guard is also crash protection rather than only fidelity. `Node.Text()` panics on a
	// binding pattern, and the walk recovers per file, so one such executor would cost every rule
	// in the package its verdict on that file.
	secondName := parameters[1].Name()
	if secondName == nil || secondName.Kind != ast.KindIdentifier {
		return
	}

	targets := preferPromiseRejectErrorsRejectSymbols(ctx, parameters, secondName.Text())
	if len(targets) == 0 {
		return
	}

	preferPromiseRejectErrorsForEachCallOfSymbols(ctx, executor, targets, func(call *ast.Node) {
		preferPromiseRejectErrorsCheckCall(ctx, call, settings)
	})
}

// preferPromiseRejectErrorsRejectSymbols returns the symbols a rejection call may resolve to.
//
// Upstream finds "the first declared variable matching the second parameter's name", and its own
// comment explains why the duplicate-parameter case is safe there: the second parameter
// immediately overwrites the first, and a function with duplicate parameters cannot have
// destructuring or defaults, so nothing can observe the first binding.
//
// That reasoning is about one variable. Measured here, our checker gives the two parameters of
// `function(reject, reject) { reject(5) }` DIFFERENT symbols, and the call resolves to the FIRST
// one. So keying strictly on the second parameter's symbol goes silent on a case upstream reports,
// and no reading of either source would have shown it. Collecting every parameter spelled the same
// as the second reproduces upstream's verdict on all sixteen executor cases in the corpus.
//
// This widens nothing in the ordinary case: with distinct parameter names exactly one parameter
// matches, so the set has one element and the behavior is identical.
func preferPromiseRejectErrorsRejectSymbols(
	ctx rule.Context,
	parameters []*ast.Node,
	rejectName string,
) map[*ast.Symbol]bool {
	targets := map[*ast.Symbol]bool{}
	for _, parameter := range parameters {
		name := parameter.Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.Text() != rejectName {
			continue
		}
		if symbol := ctx.TypeChecker.GetSymbolAtLocation(name); symbol != nil {
			targets[symbol] = true
		}
	}
	return targets
}

// preferPromiseRejectErrorsForEachCallOfSymbols visits every call in the executor whose callee is
// an identifier resolving to one of the target symbols.
//
// Two properties are deliberate and both are pinned by fixtures. It recurses into nested functions,
// because a reject called from a closure inside the executor still rejects this promise and
// upstream reports it. And it only considers an identifier in CALLEE position, because passing
// reject somewhere as a value is not a rejection: `resolve(5, reject)` is one of upstream's clean
// cases.
//
// Symbol identity rather than name equality is what declines the two shadowing cases in the corpus,
// where an inner function parameter or an inner `const` is spelled `reject` and calling it rejects
// nothing.
func preferPromiseRejectErrorsForEachCallOfSymbols(
	ctx rule.Context,
	executor *ast.Node,
	targets map[*ast.Symbol]bool,
	visit func(call *ast.Node),
) {
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if ast.IsCallExpression(node) {
			callee := preferPromiseRejectErrorsUnwrap(node.Expression())
			if callee != nil && callee.Kind == ast.KindIdentifier &&
				targets[ctx.TypeChecker.GetSymbolAtLocation(callee)] {
				visit(node)
			}
		}
		node.ForEachChild(walk)
		return false
	}

	// Walking the executor's children rather than the executor itself, so a parameter default that
	// calls reject is still reached: `(resolve, reject, somethingElse = reject(5)) => {}` is one of
	// upstream's reported cases and the call sits in the parameter list rather than in the body.
	executor.ForEachChild(walk)
}

// preferPromiseRejectErrorsIsGlobalPromiseReject answers whether a call is `Promise.reject(...)` on
// the global Promise.
//
// Upstream is `isSpecificMemberAccess(callee, "Promise", "reject")` plus a global check on the
// receiver, and it accepts every spelling whose property name the syntax settles. The corpus
// reports `Promise['reject'](5)` shapes through optional chaining and parentheses, so the property
// name is read through the shelf rather than by matching a property access alone.
func preferPromiseRejectErrorsIsGlobalPromiseReject(ctx rule.Context, call *ast.Node) bool {
	callee := preferPromiseRejectErrorsUnwrap(call.Expression())
	if callee == nil {
		return false
	}

	// `property.Static` rather than `Textual`, measured against the installed rule: a template key
	// reports upstream, and `Textual` excludes templates. `Static` also accepts a private name,
	// which is correct here for a reason worth stating, since it looks like a hole: `Text()` on a
	// private identifier carries the leading hash, so `Promise.#reject(5)` answers "#reject" and
	// can never equal "reject". That is upstream's verdict too, and it is one of its clean cases.
	name, ok := property.AccessedName(callee, property.Static)
	if !ok || name != "reject" {
		return false
	}

	var receiver *ast.Node
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = callee.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		receiver = callee.AsElementAccessExpression().Expression
	default:
		return false
	}

	receiver = preferPromiseRejectErrorsUnwrap(receiver)
	if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != "Promise" {
		return false
	}
	return preferPromiseRejectErrorsResolvesToTheGlobalPromise(ctx, receiver)
}

// preferPromiseRejectErrorsResolvesToTheGlobalPromise asks upstream's `isGlobalReference` question
// through the checker: is this `Promise` the global one, or a binding shadowing it?
//
// A declaration in a declaration file is the global. Measured across the five shadowing shapes the
// corpus writes before this was built on: the global answers five ambient declarations, while a
// `let`, a hoisted `var`, a parameter and a block-scoped class each answer exactly one
// non-ambient declaration. Every declaration must be ambient rather than just the first, which is
// the index-zero trap this tree has been bitten by before.
func preferPromiseRejectErrorsResolvesToTheGlobalPromise(ctx rule.Context, identifier *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !file.IsDeclarationFile {
			return false
		}
	}
	return true
}

// preferPromiseRejectErrorsIsGlobalUndefined answers whether an expression is the global
// `undefined` rather than a binding shadowing it.
//
// Written as the COMPLEMENT of the ambient test above, and that asymmetry is the whole point.
// Measured: the real `undefined` resolves to a symbol with ZERO declarations, while
// `function f(undefined) { ... }` gives its parameter one. So the shipped `resolvesToAGlobal`
// exemplar, which requires at least one ambient declaration, answers FALSE on every input this
// arm must report. A rule built on it would pass its whole corpus while being silently inert on
// `Promise.reject(undefined)`, which is upstream's own reported case.
//
// The corpus pins both directions: the bare form is invalid and the shadowed form is valid.
func preferPromiseRejectErrorsIsGlobalUndefined(ctx rule.Context, expression *ast.Node) bool {
	if expression == nil || expression.Kind != ast.KindIdentifier ||
		expression.Text() != "undefined" {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(expression)
	// No symbol, or a symbol declared nowhere in source, is the global. A shadowing binding always
	// carries its own declaration.
	if symbol == nil {
		return true
	}
	return len(symbol.Declarations) == 0
}

// preferPromiseRejectErrorsCouldBeError is upstream's `couldBeError` from its shared ast utilities.
//
// The judgment is deliberately optimistic: it asks whether the expression COULD produce an Error,
// not whether it does. An identifier, a call, a member read and an await are all opaque, so they
// pass. Everything the syntax settles as not-an-Error reports.
func preferPromiseRejectErrorsCouldBeError(node *ast.Node) bool {
	node = preferPromiseRejectErrorsUnwrap(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindCallExpression,
		ast.KindNewExpression,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindTaggedTemplateExpression,
		ast.KindYieldExpression,
		ast.KindAwaitExpression:
		// Possibly an error object. Upstream also names ChainExpression here, which our parser does
		// not produce: optional chaining is a question token on the access itself, so `obj?.foo`
		// arrives as a property access and is already covered by the arm above. The corpus's two
		// optional-chaining clean cases pin it.
		return true

	case ast.KindBinaryExpression:
		return preferPromiseRejectErrorsBinaryCouldBeError(node.AsBinaryExpression())

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return preferPromiseRejectErrorsCouldBeError(conditional.WhenTrue) ||
			preferPromiseRejectErrorsCouldBeError(conditional.WhenFalse)
	}

	return false
}

// preferPromiseRejectErrorsBinaryCouldBeError handles the assignment, logical and comma operators,
// which our parser folds into one binary node where upstream has three node types.
func preferPromiseRejectErrorsBinaryCouldBeError(binary *ast.BinaryExpression) bool {
	switch binary.OperatorToken.Kind {
	// Upstream's AssignmentExpression with `=` or `&&=`: the expression takes the right side's
	// value.
	case ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
		return preferPromiseRejectErrorsCouldBeError(binary.Right)

	// Upstream's `||=` and `??=`: either side can be the result.
	case ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
		return preferPromiseRejectErrorsCouldBeError(binary.Left) ||
			preferPromiseRejectErrorsCouldBeError(binary.Right)

	// Upstream's LogicalExpression with `&&`. If it short-circuits the left side was falsy and
	// therefore not an error, and otherwise it takes the right side, so only the right side can be
	// a plausible error. That asymmetry is why `Promise.reject(foo && 5)` reports while
	// `Promise.reject(5 && foo)` does not, and the corpus carries both.
	case ast.KindAmpersandAmpersandToken:
		return preferPromiseRejectErrorsCouldBeError(binary.Right)

	// Upstream's `||` and `??`: either side can be the result.
	case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return preferPromiseRejectErrorsCouldBeError(binary.Left) ||
			preferPromiseRejectErrorsCouldBeError(binary.Right)

	// Upstream's SequenceExpression, which our parser spells as a comma binary operator. The value
	// is the last expression's, and upstream reads `exprs.at(-1)` after requiring a non-empty list.
	// A comma binary node always has both sides, so the emptiness test has no counterpart here.
	case ast.KindCommaToken:
		return preferPromiseRejectErrorsCouldBeError(binary.Right)
	}

	// Every remaining assignment operator is arithmetic or bitwise, and such an expression either
	// evaluates to a primitive or throws, so it cannot be an Error. Every remaining non-assignment
	// operator is a comparison or an arithmetic operator, which upstream's switch falls through to
	// its default for. The corpus pins six of the mathematical assignments explicitly.
	return false
}

// preferPromiseRejectErrorsIsFunction answers upstream's `isFunction` for an executor argument.
//
// A function declaration cannot appear in an argument position, but it is named here because
// upstream's helper accepts it and a parse recovering from malformed source can produce shapes the
// grammar forbids.
func preferPromiseRejectErrorsIsFunction(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindFunctionDeclaration:
		return true
	}
	return false
}

// preferPromiseRejectErrorsUnwrap strips parentheses and the type-only wrappers.
//
// Written as a loop with its own nil check rather than calling `ast.SkipParentheses`, which
// dereferences its argument. This matters twice over here: our parser keeps a parenthesis node that
// upstream's folds away, so `(Promise?.reject)(5)` reaches the callee test wrapped where upstream
// sees it bare, and both parenthesized forms are in the corpus as reported cases.
//
// The four TypeScript wrappers are here for the same reason and it is not an improvement on
// upstream, it is what upstream already does. ESTree has no node for any of them, so upstream's
// switch never mentions them and its verdict is decided by the value INSIDE. Measured on the
// parenthesis it does have: `Promise.reject((5))` reports and `Promise.reject((error))` does not,
// so the wrapper is transparent and the inner expression decides.
//
// Left out, they would fall to the switch's default and report, which is how this was found: the
// dry run's one finding on the real tree was `reject(error as Error)`, source that asserts the
// value IS an Error. `no_unsafe_optional_chaining.go:317` and `no_misleading_character_class.go:418`
// take the same set for the same stated reason.
func preferPromiseRejectErrorsUnwrap(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindTypeAssertionExpression, ast.KindNonNullExpression:
			node = node.Expression()
		default:
			return node
		}
	}
	return node
}
