package react

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	type_checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// This file is the half of `react/jsx-no-constructed-context-values` that upstream does not have:
// a memoized provider value is stable only if every input to the memo is stable.
//
// Upstream stops at the wrapper. It sees `useMemo` and is satisfied, so
//
//	const theme = mergeTheme(Base, override);              // a new object every render
//	const value = useMemo(() => ({ open, theme }), [open, theme]);
//	return <Ctx.Provider value={value}>...
//
// is clean upstream while the memo recomputes on every render and every consumer re-renders exactly
// as if the wrapper were not there. That is the shape the review found at DialogRoot, DrawerRoot,
// TableRoot and Field: wrapping the value, which is upstream's own remedy, would have done nothing.
//
// The judgment is one-sided on purpose. A dependency is reported only when the code PROVES it can be
// a new identity on a render: it is built by an object, array, function, class, `new`, JSX or
// regular-expression expression evaluated during render, or it is the result of a call whose body
// returns such a construction on some path (or of any async function or generator, which return a new
// promise or iterator every time), or it is another memo that is itself unstable. Anything the
// analysis cannot see into is treated as stable: parameters, props, imports, module-scope constants,
// destructured hook results, `let` bindings (which can be reassigned), property paths, and calls into
// declarations with no body. Primitive-typed dependencies compare by value, so they are stable however
// they were computed.
//
// Kirk's ruling, 2026-10-01, on the cohere rule review: "go for C right out of the gate". Recorded as
// a deliberate divergence in `jsx_no_constructed_context_values.md`.

// jsxNoConstructedContextValuesStabilityDepthLimit bounds the walk through identifiers, memos and
// callee bodies. Real chains measured in ahra are at most three hops; the limit exists for cycles
// the visited set does not catch, such as a recursive function reached through two names.
const jsxNoConstructedContextValuesStabilityDepthLimit = 8

// jsxNoConstructedContextValuesInstability is why a memo's dependency changes every render.
type jsxNoConstructedContextValuesInstability struct {
	// dependency is the entry in the provider memo's dependency list, and where the finding points.
	dependency *ast.Node

	// origin is where the new identity is made: the construction, or the `return` in a callee.
	origin *ast.Node

	// reason completes the sentence "the `x` dependency ...", rendered without the line, which
	// the message appends from origin.
	reason string
}

// jsxNoConstructedContextValuesMemoHookName names the memo hook a call invokes, or "".
//
// Name-based, matching how the React rules in this package identify hooks: the bare name or the
// `React.` qualified one. A project defining its own unrelated `useMemo` is not a shape that exists
// in this tree.
func jsxNoConstructedContextValuesMemoHookName(call *ast.Node) string {
	if call == nil || call.Kind != ast.KindCallExpression {
		return ""
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	name := ""
	switch callee.Kind {
	case ast.KindIdentifier:
		name = callee.Text()
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		receiver := ast.SkipParentheses(access.Expression)
		if receiver.Kind != ast.KindIdentifier || receiver.Text() != "React" {
			return ""
		}
		if access.Name() == nil {
			return ""
		}
		name = access.Name().Text()
	}
	if name == "useMemo" || name == "useCallback" {
		return name
	}
	return ""
}

// jsxNoConstructedContextValuesUnwrap strips the wrappers that do not change identity: parentheses,
// `as`, `satisfies` and a non-null assertion.
func jsxNoConstructedContextValuesUnwrap(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		default:
			return node
		}
	}
	return nil
}

// jsxNoConstructedContextValuesNearestFunction is the closest function-like ancestor of a node.
func jsxNoConstructedContextValuesNearestFunction(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLike(current) {
			return current
		}
	}
	return nil
}

// jsxNoConstructedContextValuesRenderConst resolves an identifier to the `const` declaration it names
// when that declaration runs on every call of `scope`: declared directly in that function rather
// than in a nested one, bound to a plain name rather than a pattern.
//
// Only `const`, because a `let` can be reassigned to something stable after its initializer, and
// following the initializer alone would report a value the code later replaced.
func jsxNoConstructedContextValuesRenderConst(ctx rule.Context, identifier *ast.Node, scope *ast.Node, render *ast.Node) *ast.Node {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return nil
	}
	// The first declaration. A `const` cannot merge with anything, and a function declared with
	// overloads in a render body is a function declared during render whichever signature answers.
	declaration := symbol.Declarations[0]
	if !jsxNoConstructedContextValuesRunsWith(jsxNoConstructedContextValuesNearestFunction(declaration), scope, render) {
		return nil
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return declaration
	case ast.KindVariableDeclaration:
		// A destructured name declares a binding element rather than a variable declaration, so
		// it never reaches this arm: what a pattern pulls out of a value is not that value.
		if !ast.IsVarConst(declaration) {
			return nil
		}
		return declaration
	}
	return nil
}

// jsxNoConstructedContextValuesRunsWith reports whether a declaration made in `owner` is evaluated
// afresh each time `scope` runs, for the component `render`.
//
// `owner` being `scope` itself always qualifies. Beyond that, only a function nested inside the
// render body runs per render, so a memo factory or a helper declared in the component sees the
// component's own declarations as per-render values. Nothing above `render` qualifies: a component
// declared inside another function re-renders on its own state while the outer declarations keep
// their identity. And a helper outside the render (a module-level function, or one returned by a
// factory) sees only its own body, because its enclosing function ran once, not once per call.
func jsxNoConstructedContextValuesRunsWith(owner *ast.Node, scope *ast.Node, render *ast.Node) bool {
	if owner == nil {
		return false
	}
	if owner == scope {
		return true
	}
	var chain []*ast.Node
	for current := scope; current != nil; current = jsxNoConstructedContextValuesNearestFunction(current) {
		chain = append(chain, current)
		if current == render {
			break
		}
	}
	if chain[len(chain)-1] != render {
		return false
	}
	for _, function := range chain {
		if function == owner {
			return true
		}
	}
	return false
}

// jsxNoConstructedContextValuesMemoOf finds the memo call behind a provider's value expression: the
// call written inline, or the call a render-scoped `const` was initialized with.
func jsxNoConstructedContextValuesMemoOf(ctx rule.Context, value *ast.Node, scope *ast.Node) *ast.Node {
	value = jsxNoConstructedContextValuesUnwrap(value)
	if value == nil {
		return nil
	}
	if jsxNoConstructedContextValuesMemoHookName(value) != "" {
		return value
	}
	if value.Kind != ast.KindIdentifier {
		return nil
	}
	declaration := jsxNoConstructedContextValuesRenderConst(ctx, value, scope, scope)
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
		return nil
	}
	initializer := jsxNoConstructedContextValuesUnwrap(declaration.AsVariableDeclaration().Initializer)
	if jsxNoConstructedContextValuesMemoHookName(initializer) == "" {
		return nil
	}
	return initializer
}

// jsxNoConstructedContextValuesConstructionKind names a node that evaluates to a new identity every
// time it runs, in upstream's own vocabulary, or returns "".
func jsxNoConstructedContextValuesConstructionKind(node *ast.Node) string {
	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		return "object"
	case ast.KindArrayLiteralExpression:
		return "array"
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return "function expression"
	case ast.KindClassExpression:
		return "class expression"
	case ast.KindNewExpression:
		return "new expression"
	case ast.KindRegularExpressionLiteral:
		return "regular expression"
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		return "JSX element"
	case ast.KindJsxFragment:
		return "JSX fragment"
	}
	return ""
}

// jsxNoConstructedContextValuesStabilityWalk carries the state one provider's analysis shares.
//
// `active` is an in-progress set rather than a visited set: a declaration, callee or memo is marked
// on entry and cleared on exit, so it guards a cycle without making a second, independent encounter
// of the same name answer differently from the first.
type jsxNoConstructedContextValuesStabilityWalk struct {
	ctx    rule.Context
	active map[*ast.Node]bool

	// render is the component whose provider is being judged.
	render *ast.Node
}

// jsxNoConstructedContextValuesEvaluation is what one expression can evaluate to.
type jsxNoConstructedContextValuesEvaluation struct {
	// fresh is set when some evaluation is proven to be a new identity.
	fresh *jsxNoConstructedContextValuesInstability

	// other is set when some evaluation can instead be an object that already existed: a value read
	// from somewhere this walk cannot see into. When both are set, the fresh value may have been
	// stored on its way out and read back on a later call through the other path, which is exactly
	// what a cache does, so it proves nothing unless it never escapes.
	other bool

	// holders are the `const` (or assigned local) declarations the fresh value passed through in
	// the current function, which the escape check reads when `other` is set.
	holders []*ast.Node
}

// jsxNoConstructedContextValuesUnknown is an evaluation the walk cannot see into.
var jsxNoConstructedContextValuesUnknown = jsxNoConstructedContextValuesEvaluation{other: true}

// combine merges two evaluations that are alternatives, as the two branches of `?:` or the two
// sides of `??` are. The left finding wins for the message; the holders of both are kept, so an
// escape on either side silences the pair.
func (evaluation jsxNoConstructedContextValuesEvaluation) combine(
	alternative jsxNoConstructedContextValuesEvaluation,
) jsxNoConstructedContextValuesEvaluation {
	fresh := evaluation.fresh
	if fresh == nil {
		fresh = alternative.fresh
	}
	return jsxNoConstructedContextValuesEvaluation{
		fresh:   fresh,
		other:   evaluation.other || alternative.other,
		holders: append(append([]*ast.Node{}, evaluation.holders...), alternative.holders...),
	}
}

// isPrimitive reports whether every part of a node's type compares by value, so a fresh computation
// of it is still equal to the last one.
func (walk *jsxNoConstructedContextValuesStabilityWalk) isPrimitive(node *ast.Node) bool {
	t := walk.ctx.TypeChecker.GetTypeAtLocation(node)
	if t == nil {
		return false
	}
	const primitive = checker.TypeFlagsStringLike | checker.TypeFlagsNumberLike |
		checker.TypeFlagsBigIntLike | checker.TypeFlagsBooleanLike | checker.TypeFlagsEnumLike |
		checker.TypeFlagsESSymbolLike | checker.TypeFlagsNull | checker.TypeFlagsUndefined |
		checker.TypeFlagsVoid
	for part := range type_checking.UnionTypePartsSeq(t) {
		if !type_checking.IsTypeFlagSet(part, primitive) {
			return false
		}
	}
	return true
}

// memoFinding explains why a memo's result is a new identity on a render, or returns nil.
//
// Both halves are required. The memo has to recompute (a dependency proven unstable, or no
// dependency list at all), and what it recomputes has to be new: `useCallback` always hands back the
// callback it was given, which is a new function on every render, but a `useMemo` factory that
// returns an object that already existed hands back that same object however often it runs. The
// returned `other` is the factory's, for the caller's escape check.
//
// A dependency list has to be an array literal to be judged. A list passed by name is a shape this
// analysis cannot see into, so it is stable by the one-sided rule. The returned finding's
// `dependency` is nil when the list is missing.
func (walk *jsxNoConstructedContextValuesStabilityWalk) memoFinding(
	call *ast.Node,
	scope *ast.Node,
	depth int,
) (*jsxNoConstructedContextValuesInstability, bool) {
	if depth > jsxNoConstructedContextValuesStabilityDepthLimit || walk.active[call] {
		return nil, true
	}
	walk.active[call] = true
	defer delete(walk.active, call)

	arguments := call.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return nil, true
	}
	var finding *jsxNoConstructedContextValuesInstability
	if len(arguments.Nodes) == 1 {
		finding = &jsxNoConstructedContextValuesInstability{origin: call}
	} else {
		list := jsxNoConstructedContextValuesUnwrap(arguments.Nodes[1])
		if list == nil || list.Kind != ast.KindArrayLiteralExpression {
			return nil, true
		}
		for _, element := range list.AsArrayLiteralExpression().Elements.Nodes {
			if found := walk.dependencyFinding(element, scope, depth+1); found != nil {
				finding = found
				break
			}
		}
		if finding == nil {
			return nil, true
		}
	}

	if jsxNoConstructedContextValuesMemoHookName(call) == "useCallback" {
		return finding, false
	}
	factory := jsxNoConstructedContextValuesUnwrap(arguments.Nodes[0])
	if factory == nil || (factory.Kind != ast.KindArrowFunction && factory.Kind != ast.KindFunctionExpression) {
		return nil, true
	}
	result := walk.functionEvaluation(factory, depth+1)
	if result.fresh == nil {
		return nil, true
	}
	return finding, result.other
}

// dependencyFinding explains why one entry of a memo's dependency list is a new identity on a
// render, or returns nil. This is where a render-scoped holder's escape is checked.
func (walk *jsxNoConstructedContextValuesStabilityWalk) dependencyFinding(
	element *ast.Node,
	scope *ast.Node,
	depth int,
) *jsxNoConstructedContextValuesInstability {
	result := walk.evaluate(element, scope, depth)
	if result.fresh == nil {
		return nil
	}
	if result.other && walk.anyEscapes(result.holders, true) {
		return nil
	}
	finding := *result.fresh
	finding.dependency = element
	return &finding
}

// evaluate says what an expression evaluated in `scope` can be: proven new, possibly an existing
// object, or both.
func (walk *jsxNoConstructedContextValuesStabilityWalk) evaluate(
	node *ast.Node,
	scope *ast.Node,
	depth int,
) jsxNoConstructedContextValuesEvaluation {
	node = jsxNoConstructedContextValuesUnwrap(node)
	if node == nil {
		return jsxNoConstructedContextValuesEvaluation{}
	}
	if depth > jsxNoConstructedContextValuesStabilityDepthLimit {
		return jsxNoConstructedContextValuesUnknown
	}
	if walk.isPrimitive(node) {
		return jsxNoConstructedContextValuesEvaluation{}
	}

	if kind := jsxNoConstructedContextValuesConstructionKind(node); kind != "" {
		return jsxNoConstructedContextValuesEvaluation{fresh: &jsxNoConstructedContextValuesInstability{
			origin: node,
			reason: fmt.Sprintf("is a new %s built during render", kind),
		}}
	}

	switch node.Kind {
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return walk.evaluate(conditional.WhenTrue, scope, depth+1).
			combine(walk.evaluate(conditional.WhenFalse, scope, depth+1))

	case ast.KindBinaryExpression:
		// A logical operator evaluates to either side. A comma evaluates to its right side, and so
		// does a plain assignment, but an assignment is also a store: into a local it adds a holder,
		// and into anything else (a property, an element, an outer variable) it is an escape, which
		// proves nothing. Every other binary operator produces a primitive, which the type check
		// above has already let through.
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return walk.evaluate(binary.Left, scope, depth+1).combine(walk.evaluate(binary.Right, scope, depth+1))
		case ast.KindCommaToken:
			return walk.evaluate(binary.Right, scope, depth+1)
		case ast.KindEqualsToken:
			target := jsxNoConstructedContextValuesUnwrap(binary.Left)
			local := jsxNoConstructedContextValuesLocalDeclaration(walk.ctx, target, scope)
			if local == nil {
				return jsxNoConstructedContextValuesUnknown
			}
			result := walk.evaluate(binary.Right, scope, depth+1)
			if result.fresh != nil {
				result.holders = append(result.holders, local)
			}
			return result
		}
		return jsxNoConstructedContextValuesUnknown

	case ast.KindCallExpression:
		if name := jsxNoConstructedContextValuesMemoHookName(node); name != "" {
			inner, other := walk.memoFinding(node, scope, depth+1)
			if inner == nil {
				return jsxNoConstructedContextValuesUnknown
			}
			reason := fmt.Sprintf("comes from a %s with no dependency list", name)
			if inner.dependency != nil {
				reason = fmt.Sprintf("comes from a %s whose own dependency %q %s",
					name, jsxNoConstructedContextValuesDependencyText(walk.ctx, inner.dependency), inner.reason)
			}
			return jsxNoConstructedContextValuesEvaluation{
				fresh: &jsxNoConstructedContextValuesInstability{origin: inner.origin, reason: reason},
				other: other,
			}
		}
		return walk.callEvaluation(node, depth+1)

	case ast.KindIdentifier:
		declaration := jsxNoConstructedContextValuesRenderConst(walk.ctx, node, scope, walk.render)
		if declaration == nil || walk.active[declaration] {
			return jsxNoConstructedContextValuesUnknown
		}
		if declaration.Kind == ast.KindFunctionDeclaration {
			return jsxNoConstructedContextValuesEvaluation{
				fresh: &jsxNoConstructedContextValuesInstability{
					origin: declaration,
					reason: "is a function declared during render",
				},
				holders: []*ast.Node{declaration},
			}
		}
		walk.active[declaration] = true
		result := walk.evaluate(declaration.AsVariableDeclaration().Initializer, scope, depth+1)
		delete(walk.active, declaration)
		if result.fresh != nil {
			result.holders = append(result.holders, declaration)
		}
		return result
	}
	return jsxNoConstructedContextValuesUnknown
}

// callEvaluation says what a call can return, by reading the body of the function the checker
// resolves it to.
//
// The callee has to be source the program holds: every signature that can be read has a body (a
// function declaration, method, function expression or arrow), and every one that cannot (an
// interface member, a function type, an overload, anything in a declaration file, so every library
// and every `declare function`) is unknown. An async callee or a generator is new outright.
func (walk *jsxNoConstructedContextValuesStabilityWalk) callEvaluation(
	call *ast.Node,
	depth int,
) jsxNoConstructedContextValuesEvaluation {
	signature := walk.ctx.TypeChecker.GetResolvedSignature(call)
	if signature == nil {
		return jsxNoConstructedContextValuesUnknown
	}
	callee := signature.Declaration()
	if callee == nil || callee.Body() == nil {
		return jsxNoConstructedContextValuesUnknown
	}

	name := jsxNoConstructedContextValuesCalleeName(call)
	// An async function returns a new promise from every call, and a generator a new iterator,
	// whatever their bodies return. That is the strongest proof this walk has.
	if ast.HasSyntacticModifier(callee, ast.ModifierFlagsAsync) {
		return jsxNoConstructedContextValuesEvaluation{fresh: &jsxNoConstructedContextValuesInstability{
			origin: callee,
			reason: fmt.Sprintf("is the result of `%s(...)`, an async function, which returns a new promise on every call", name),
		}}
	}
	if jsxNoConstructedContextValuesIsGenerator(callee) {
		return jsxNoConstructedContextValuesEvaluation{fresh: &jsxNoConstructedContextValuesInstability{
			origin: callee,
			reason: fmt.Sprintf("is the result of `%s(...)`, a generator, which returns a new iterator on every call", name),
		}}
	}

	result := walk.functionEvaluation(callee, depth+1)
	if result.fresh != nil {
		result.fresh = &jsxNoConstructedContextValuesInstability{
			origin: result.fresh.origin,
			reason: fmt.Sprintf("is the result of `%s(...)`, which can return a new value on every call", name),
		}
	}
	return result
}

// functionEvaluation says what a function body returns, across every `return` it owns (not those of
// nested functions), or its expression body.
//
// "Some path returns a new value" proves a new identity on that path when no other path can return
// an existing object. When one can, the new value proves nothing if it escapes before it is
// returned: passed to a call, stored into a property, an element or anything outside the function,
// captured by a closure, aliased, or used as a method receiver. That is the caching helper, whose
// first call builds and stores and whose later calls read the stored one back, and it is checked
// here, in the function's own scope, before the holders are dropped for the caller.
func (walk *jsxNoConstructedContextValuesStabilityWalk) functionEvaluation(
	function *ast.Node,
	depth int,
) jsxNoConstructedContextValuesEvaluation {
	if depth > jsxNoConstructedContextValuesStabilityDepthLimit || walk.active[function] {
		return jsxNoConstructedContextValuesUnknown
	}
	body := function.Body()
	if body == nil {
		return jsxNoConstructedContextValuesUnknown
	}
	walk.active[function] = true
	defer delete(walk.active, function)

	var result jsxNoConstructedContextValuesEvaluation
	if body.Kind != ast.KindBlock {
		result = walk.evaluate(body, function, depth+1)
	} else {
		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindReturnStatement {
				result = result.combine(walk.evaluate(node.AsReturnStatement().Expression, function, depth+1))
				return false
			}
			if ast.IsFunctionLike(node) || ast.IsClassLike(node) {
				return false
			}
			return node.ForEachChild(visit)
		}
		body.ForEachChild(visit)
	}

	// Only the holders this function declared are judged here, against this function's body. A holder
	// from an enclosing function (a render `const` a memo factory or a nested helper returns) travels
	// on to the dependency, where it is judged against the body that declared it.
	var own, enclosing []*ast.Node
	for _, holder := range result.holders {
		if jsxNoConstructedContextValuesNearestFunction(holder) == function {
			own = append(own, holder)
		} else {
			enclosing = append(enclosing, holder)
		}
	}
	if result.fresh != nil && result.other && walk.anyEscapes(own, false) {
		result.fresh = nil
	}
	result.holders = enclosing
	return result
}

// jsxNoConstructedContextValuesLocalDeclaration resolves an identifier to a variable declared
// directly in `scope`, of any kind, or nil. It is what an assignment target has to be for the
// assignment to be a holder rather than an escape.
func jsxNoConstructedContextValuesLocalDeclaration(ctx rule.Context, target *ast.Node, scope *ast.Node) *ast.Node {
	if target == nil || target.Kind != ast.KindIdentifier {
		return nil
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(target)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration ||
		jsxNoConstructedContextValuesNearestFunction(declaration) != scope {
		return nil
	}
	return declaration
}

// anyEscapes reports whether any holder is referenced anywhere in the function that declared it,
// nested functions included, in a way that could let the value outlive the call.
func (walk *jsxNoConstructedContextValuesStabilityWalk) anyEscapes(holders []*ast.Node, render bool) bool {
	for _, holder := range holders {
		scope := jsxNoConstructedContextValuesNearestFunction(holder)
		if scope == nil || scope.Body() == nil {
			return true
		}
		body := scope.Body()
		name := holder.Name()
		if name == nil {
			return true
		}
		symbol := walk.ctx.TypeChecker.GetSymbolAtLocation(name)
		if symbol == nil {
			return true
		}
		escaped := false
		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if escaped {
				return true
			}
			if node.Kind == ast.KindIdentifier && node != name {
				reference := walk.ctx.TypeChecker.GetSymbolAtLocation(node)
				if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
					reference = walk.ctx.TypeChecker.GetShorthandAssignmentValueSymbol(node.Parent)
				}
				if reference == symbol && !jsxNoConstructedContextValuesStaysHome(node, scope, render) {
					escaped = true
					return true
				}
			}
			return node.ForEachChild(visit)
		}
		body.ForEachChild(visit)
		if escaped {
			return true
		}
	}
	return false
}

// jsxNoConstructedContextValuesStaysHome reports whether one reference to a holder leaves the value
// where it was: returned by its own function, read through (a field read or write, not a method
// call), compared, tested, or reassigned. In a render scope two more places are home: the dependency
// list of a memo, and the object a memo factory returns, which becomes the provider value itself.
// Everything else is treated as an escape, including a reference this list never thought of.
func jsxNoConstructedContextValuesStaysHome(reference *ast.Node, scope *ast.Node, render bool) bool {
	expression := reference
	for expression.Parent != nil {
		switch expression.Parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindNonNullExpression:
			expression = expression.Parent
			continue
		}
		break
	}
	parent := expression.Parent
	if parent == nil {
		return false
	}
	ownFunction := func(function *ast.Node) bool {
		return function == scope || (render && jsxNoConstructedContextValuesIsMemoFactory(function))
	}

	switch parent.Kind {
	case ast.KindReturnStatement:
		return ownFunction(jsxNoConstructedContextValuesNearestFunction(expression))
	case ast.KindArrowFunction:
		return parent.Body() == expression && ownFunction(parent)
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression == expression &&
			!jsxNoConstructedContextValuesIsCallee(parent)
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression == expression &&
			!jsxNoConstructedContextValuesIsCallee(parent)
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken:
			return binary.Left == expression
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken,
			ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindInstanceOfKeyword:
			return true
		}
		return false
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindTypeOfExpression:
		return true
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == expression
	case ast.KindArrayLiteralExpression:
		return render && jsxNoConstructedContextValuesIsDependencyList(parent)
	case ast.KindShorthandPropertyAssignment:
		return render && jsxNoConstructedContextValuesIsMemoResult(parent.Parent)
	case ast.KindPropertyAssignment:
		return render && parent.AsPropertyAssignment().Initializer == expression &&
			jsxNoConstructedContextValuesIsMemoResult(parent.Parent)
	}
	return false
}

// jsxNoConstructedContextValuesIsCallee reports whether an access is the callee of a call, which
// hands the object to the method as its receiver.
func jsxNoConstructedContextValuesIsCallee(access *ast.Node) bool {
	parent := access.Parent
	return parent != nil && parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == access
}

// jsxNoConstructedContextValuesIsMemoFactory reports whether a function is the factory of a memo.
func jsxNoConstructedContextValuesIsMemoFactory(function *ast.Node) bool {
	if function == nil || function.Parent == nil || jsxNoConstructedContextValuesMemoHookName(function.Parent) == "" {
		return false
	}
	arguments := function.Parent.AsCallExpression().Arguments
	return arguments != nil && len(arguments.Nodes) > 0 && arguments.Nodes[0] == function
}

// jsxNoConstructedContextValuesIsDependencyList reports whether an array literal is a memo's list.
func jsxNoConstructedContextValuesIsDependencyList(list *ast.Node) bool {
	if list.Parent == nil || jsxNoConstructedContextValuesMemoHookName(list.Parent) == "" {
		return false
	}
	arguments := list.Parent.AsCallExpression().Arguments
	return len(arguments.Nodes) > 1 && arguments.Nodes[1] == list
}

// jsxNoConstructedContextValuesIsMemoResult reports whether an object literal is what a memo factory
// returns, as its expression body or as the operand of one of its own returns.
func jsxNoConstructedContextValuesIsMemoResult(object *ast.Node) bool {
	expression := object
	for expression.Parent != nil && expression.Parent.Kind == ast.KindParenthesizedExpression {
		expression = expression.Parent
	}
	parent := expression.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindArrowFunction:
		return parent.Body() == expression && jsxNoConstructedContextValuesIsMemoFactory(parent)
	case ast.KindReturnStatement:
		return jsxNoConstructedContextValuesIsMemoFactory(jsxNoConstructedContextValuesNearestFunction(parent))
	}
	return false
}

// jsxNoConstructedContextValuesIsGenerator reports whether a function-like node carries `*`.
func jsxNoConstructedContextValuesIsGenerator(function *ast.Node) bool {
	switch function.Kind {
	case ast.KindFunctionDeclaration:
		return function.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return function.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return function.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

// jsxNoConstructedContextValuesCalleeName is the callee as the reader wrote it, for the message.
func jsxNoConstructedContextValuesCalleeName(call *ast.Node) string {
	callee := jsxNoConstructedContextValuesUnwrap(call.AsCallExpression().Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text()
	case ast.KindPropertyAccessExpression:
		if name := callee.AsPropertyAccessExpression().Name(); name != nil {
			return name.Text()
		}
	}
	return "the call"
}

// jsxNoConstructedContextValuesDependencyText is a dependency's source text, trimmed of trivia.
func jsxNoConstructedContextValuesDependencyText(ctx rule.Context, dependency *ast.Node) string {
	span := rule.TokenRange(ctx.SourceFile, dependency)
	return ctx.SourceFile.Text()[span.Pos():span.End()]
}

// jsxNoConstructedContextValuesCheckMemo reports a provider value that is memoized over an input
// that changes every render, or memoized with no dependency list at all.
func jsxNoConstructedContextValuesCheckMemo(ctx rule.Context, value *ast.Node, provider *ast.Node) {
	// The caller has already established the provider sits in a component, so there is a function.
	scope := jsxNoConstructedContextValuesNearestFunction(provider)
	memo := jsxNoConstructedContextValuesMemoOf(ctx, value, scope)
	if memo == nil {
		return
	}
	hook := jsxNoConstructedContextValuesMemoHookName(memo)
	lineOf := func(node *ast.Node) int {
		line, _ := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile, rule.TokenRange(ctx.SourceFile, node).Pos())
		return line + 1
	}

	walk := &jsxNoConstructedContextValuesStabilityWalk{ctx: ctx, active: map[*ast.Node]bool{}, render: scope}
	found, _ := walk.memoFinding(memo, scope, 0)
	if found == nil {
		return
	}
	if found.dependency == nil {
		ctx.ReportNode(memo.AsCallExpression().Expression, rule.Message{
			Id: "memoWithoutDependenciesMsg",
			Description: fmt.Sprintf(
				"The %s that builds the Context provider's value (at line %d) has no dependency list, "+
					"so it recomputes on every render and every consumer of this context re-renders on "+
					"every render of this component, exactly as if it were not memoized. Pass the "+
					"dependency list.",
				hook, lineOf(memo),
			),
		})
		return
	}

	dependencyText := jsxNoConstructedContextValuesDependencyText(ctx, found.dependency)
	location := ""
	if ast.GetSourceFileOfNode(found.origin) == ctx.SourceFile {
		location = fmt.Sprintf(" (at line %d)", lineOf(found.origin))
	}
	ctx.ReportNode(found.dependency, rule.Message{
		Id: "unstableDependencyMsg",
		Description: fmt.Sprintf(
			"The %q dependency of the %s that builds the Context provider's value (at line %d) %s%s, "+
				"so the %s recomputes on every render and every consumer of this context re-renders "+
				"on every render of this component, exactly as if it were not memoized. A memoized "+
				"value is only stable when every input is: make %q stable too, by memoizing it or by "+
				"building it outside the component.",
			dependencyText, hook, lineOf(memo), found.reason, location, hook, dependencyText,
		),
	})
}
