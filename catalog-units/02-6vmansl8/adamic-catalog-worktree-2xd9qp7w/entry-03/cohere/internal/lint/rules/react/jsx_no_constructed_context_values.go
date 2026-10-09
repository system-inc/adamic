package react

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// jsxNoConstructedContextValuesConstruction is what the value expression was found to be, and
// which of the four messages that selects.
//
// The `kind` string is upstream's own `type` field, rendered into the message verbatim, so the
// spellings below are the wire format rather than a description of it: `object`, `array`,
// `function expression`, `function declaration`, `class expression`, `new expression`, `regular
// expression`, `JSX element`, `JSX fragment`, `assignment expression`. Reading them as prose and
// tidying one would change a user-visible string that upstream's corpus asserts.
type jsxNoConstructedContextValuesConstruction struct {
	// kind is upstream's `type`, interpolated into the message.
	kind string

	// node is the construction itself, and where the finding points.
	node *ast.Node

	// usage is the identifier that named the construction, when the value prop was a variable
	// rather than the construction written inline. Nil selects the default message pair.
	usage *ast.Node
}

// jsxNoConstructedContextValuesIsFunctionKind reports whether a construction wants useCallback
// rather than useMemo.
//
// Upstream appends `Func` to the message id for exactly two of its ten kinds. That is a decision
// about the remedy rather than about the finding, and it is the only thing the suffix encodes.
func jsxNoConstructedContextValuesIsFunctionKind(kind string) bool {
	return kind == "function expression" || kind == "function declaration"
}

// jsxNoConstructedContextValuesMessage builds the one of four messages this finding wants.
//
// Upstream ships four message ids whose text differs along two independent axes: whether a variable
// name is available, and whether the remedy is useMemo or useCallback. The ids are combined the
// same way here, by appending `Func`, so the id a fixture asserts is upstream's own.
//
// The line numbers are one-based, matching what upstream renders. `GetLineAndCharacterOfPosition`
// is zero-based, so both are incremented at the point of use rather than inside this function,
// which takes them already rendered.
func jsxNoConstructedContextValuesMessage(
	construction jsxNoConstructedContextValuesConstruction,
	nodeLine int,
	usageLine int,
	variableName string,
) rule.Message {
	remedy := "useMemo"
	if jsxNoConstructedContextValuesIsFunctionKind(construction.kind) {
		remedy = "useCallback"
	}

	if construction.usage == nil {
		id := "defaultMsg"
		if jsxNoConstructedContextValuesIsFunctionKind(construction.kind) {
			id += "Func"
		}
		return rule.Message{
			Id: id,
			Description: fmt.Sprintf(
				"The %s passed as the value prop to the Context provider (at line %d) changes "+
					"every render, so every consumer of this context re-renders on every render "+
					"of this component even when nothing it reads has changed. Wrap it in a %s "+
					"hook so the identity is stable between renders.",
				construction.kind, nodeLine, remedy,
			),
		}
	}

	id := "withIdentifierMsg"
	if jsxNoConstructedContextValuesIsFunctionKind(construction.kind) {
		id += "Func"
	}
	return rule.Message{
		Id: id,
		Description: fmt.Sprintf(
			"The %q %s (at line %d) passed as the value prop to the Context provider (at line "+
				"%d) changes every render, so every consumer of this context re-renders on every "+
				"render of this component even when nothing it reads has changed. Wrap it in a "+
				"%s hook so the identity is stable between renders.",
			variableName, construction.kind, nodeLine, usageLine, remedy,
		),
	}
}

// JsxNoConstructedContextValues flags a value handed to a context provider that is rebuilt on every
// render, which defeats the identity comparison every consumer of that context relies on.
//
//	valid:   function C() { const v = useMemo(() => ({a:1}), []); return <Ctx.Provider value={v}/>; }
//	valid:   function C(props) { return <Ctx.Provider value={props.v}/>; }
//	valid:   function C() { return <Ctx.Provider value={"s"}/>; }
//	valid:   const x = <Ctx.Provider value={{a:1}}/>;            (not inside a component)
//	invalid: function C() { return <Ctx.Provider value={{a:1}}/>; }
//	invalid: function C() { const v = {a:1}; return <Ctx.Provider value={v}/>; }
//	invalid: function C() { return <Ctx.Provider value={() => {}}/>; }
//
// Ported from `react/jsx-no-constructed-context-values` in `eslint-plugin-react`, read from the
// clone at `lib/rules/jsx-no-constructed-context-values.js`. `schema: false` means no options; four
// messages; no fixer. Everything recorded below was measured by driving the installed build (7.37.5)
// through the ESLint Linter API rather than reasoned from the source.
//
// # There is no file-suffix gate
//
// Measured under `.tsx`, `.jsx` and `.js`: all three report. There is no `.ts` row because a `.ts`
// file cannot hold JSX, so a gate would be unobservable there rather than absent, which is the
// distinction `self_closing_comp_test.go` records for a different rule.
//
// # Which elements are providers, and the two paths differ
//
// Upstream splits on the tag name's shape and only one half consults anything:
//
//	<X.Provider ...>     the property must be named `Provider`. The object is NEVER checked, so an
//	                     arbitrary `<Whatever.Provider>` reports and `<Whatever.NotProvider>` does
//	                     not. Measured both ways.
//	<X ...>              a bare identifier must resolve to a variable whose initializer is a call to
//	                     `createContext` or `React.createContext`. Measured: a bare `createContext()`
//	                     and the namespaced form both qualify, and any other callee does not.
//
// The second path is why this rule declares the checker. Upstream walks its own scope chain;
// `GetSymbolAtLocation` on the tag name answers the same question, and a probe confirmed it
// resolves to the `KindVariableDeclaration` with its `KindCallExpression` initializer intact.
//
// # The enclosing component test, measured across twenty shapes before this was written
//
// Upstream calls `utils.getParentComponent(node)`, which is the plugin's 959-line component
// detector. That detector is not in this tree, and reproducing it would be a substrate build rather
// than a rule port. It does not have to be reproduced here, and the reason is specific to this rule:
// the node being classified is ALREADY a JSX element, so every question the detector would ask
// about whether a function returns JSX is answered before it is asked. What is left is the name and
// the class heritage.
//
// So the classifier below was written as a hypothesis and probed against the installed build on
// twenty shapes before any of it went into the rule. All twenty agree:
//
//	function Component() {...}                       REPORTS   named, capitalized
//	function component() {...}                       silent    named, lowercase
//	Component() { helper() { ... } }                 silent    the NEAREST function decides
//	helper() { Component() { ... } }                 REPORTS   the nearest one again
//	Component() { const inner = () => ... }          silent    nearest is `inner`
//	const Component = () => ...                      REPORTS   the arrow borrows its binding
//	const component = () => ...                      silent
//	const thing = function Component() {...}         silent    the BINDING decides, not the
//	const Component = function thing() {...}         REPORTS   function expression's own name
//	class C extends React.Component { render() {} }  REPORTS   the class heritage decides
//	class C extends React.Component { other() {} }   REPORTS   any member, not just render
//	class C { render() {} }                          silent    no heritage
//	class C extends Foo.Bar { render() {} }          silent    not a React base
//	class c extends React.Component { render() {} }  REPORTS   class NAME is irrelevant
//	class C extends React.Component { f = () => }    REPORTS   a class field belongs to the class
//	class Thing { Render() {} }                      silent    a capitalized method is not enough
//	<top level>                                      silent    no enclosing function at all
//	React.memo(() => ...)                            REPORTS   the wrapper is transparent
//	React.forwardRef(() => ...)                      REPORTS
//
// Two of those rows are where the shelf disagrees with upstream and this rule cannot use it
// directly. `scope.NameOf` prefers a function expression's OWN name over its binding, which is the
// opposite of rows eight and nine; and it answers a method's name, where rows ten through sixteen
// show the class is what decides. Both are correct for the rules `scope.NameOf` was lifted from and
// wrong here, so the walk below is written out rather than delegated. That is a divergence in
// mechanism only, and the twenty rows are the evidence for it.
//
// # What counts as a construction
//
// Upstream's `isConstruction` is a recursive switch over the value expression. Every kind was
// measured; the interesting rows are the ones that recurse rather than answer:
//
//	{a:1} [1] () => {} function(){} class{} new F() /x/ <div/> <></>   report directly
//	c ? {a:1} : x          reports, on whichever branch constructs, LEFT first
//	x || {a:1}             the same, over both sides
//	o.a                    recurses into the OBJECT, and reports with the object as the usage
//	v = {a:1}              recurses into the right side, kind becomes `assignment expression`
//	({a:1} as any)         unwraps the assertion and reports the inner object
//	makeIt() "s" 1 `t` x   silent: a call result is not something this rule can judge
//
// An identifier recurses into what it was declared as, so `const v = {a:1}` reports with `v` as the
// usage and the object as the node. A function declaration is its own construction. A parameter is
// silent, which upstream spells as a bail-out on any definition that is not a variable or a
// function name.
//
// # Where the finding points, and it is not the value expression
//
// Every finding is anchored on the CONSTRUCTION, which for an indirect case is somewhere else
// entirely: `const v = {x:1}` two lines above the provider reports on `{x:1}`, not on `v`. Measured
// by slicing the reported range on twenty-six inputs. A message-id fixture cannot see this and the
// corpus does not assert it, so the span cases below do.
//
// # One message this rule can produce that upstream's corpus never exercises
//
// `defaultMsgFunc` appears in none of the twenty-three failing cases, though the rule can plainly
// emit it and does: an inline arrow reports it, measured. The corpus reaching only three of four
// ids is upstream's gap, not a statement that the fourth is unreachable.
//
// # Where cohere is deliberately stricter than upstream
//
// A memoized value is stable only if every input is. Upstream is satisfied by the `useMemo`
// wrapper; this rule also reports a memoized provider value whose dependency is proven to be a new
// identity on a render (`unstableDependencyMsg`) or whose memo has no dependency list at all
// (`memoWithoutDependenciesMsg`). Kirk's ruling, 2026-10-01. The analysis and its one-sided
// definition of "proven" live in jsx_no_constructed_context_values_stability.go, and the divergence
// is recorded in jsx_no_constructed_context_values.md.
var JsxNoConstructedContextValues = rule.Rule{
	Name:             "react/jsx-no-constructed-context-values",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}

		check := func(node *ast.Node, tagName *ast.Node, attributes *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}
			if !jsxNoConstructedContextValuesIsProvider(ctx, tagName) {
				return
			}

			value := jsxNoConstructedContextValuesValueExpression(attributes)
			if value == nil {
				return
			}

			construction := jsxNoConstructedContextValuesConstructionOf(ctx, value, 0)
			if construction == nil {
				// Upstream stops here. cohere goes one level further: a memoized value whose inputs
				// change every render is not stable either. See
				// jsx_no_constructed_context_values_stability.go.
				if jsxNoConstructedContextValuesInComponent(node) {
					jsxNoConstructedContextValuesCheckMemo(ctx, value, node)
				}
				return
			}

			// The component test runs LAST, matching upstream, which calls getParentComponent only
			// after a construction has been found. Order is not observable in the verdict but it is
			// observable in cost, and this is the cheaper arrangement on a tree where most
			// providers are fine.
			if !jsxNoConstructedContextValuesInComponent(node) {
				return
			}

			// `rule.TokenRange` rather than `node.Pos()`. A node's Pos includes its leading
			// trivia, so a declaration written on its own line after a blank line or a comment
			// reports the line the trivia STARTS on rather than the line the reader sees. Measured:
			// `function v() {}` on line 2 of its component rendered "at line 1" through Pos, where
			// upstream renders line 2. The span assertions could not see it, because
			// `ctx.ReportNode` already routes through TokenRange and so was pointing correctly all
			// along; only the interpolated line number was wrong, which is exactly the failure the
			// brief describes for any message built with a format string.
			nodeLine, _ := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile,
				rule.TokenRange(ctx.SourceFile, construction.node).Pos())
			usageLine, variableName := 0, ""
			if construction.usage != nil {
				usageLine, _ = scanner.GetLineAndCharacterOfPosition(ctx.SourceFile,
					rule.TokenRange(ctx.SourceFile, construction.usage).Pos())
				usageLine++
				if construction.usage.Kind == ast.KindIdentifier {
					variableName = construction.usage.Text()
				}
			}

			ctx.ReportNode(construction.node, jsxNoConstructedContextValuesMessage(
				*construction, nodeLine+1, usageLine, variableName,
			))
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement: func(node *ast.Node) {
				opening := node.AsJsxOpeningElement()
				check(node, opening.TagName, opening.Attributes)
			},
			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				element := node.AsJsxSelfClosingElement()
				check(node, element.TagName, element.Attributes)
			},
		}
	},
}

// jsxNoConstructedContextValuesIsProvider reports whether a tag name names a context provider.
//
// The two paths are not symmetric and the asymmetry is upstream's. A qualified name only has to end
// in `Provider`, with the object never examined, so `<Anything.Provider>` qualifies. A bare
// identifier has to resolve to a `createContext` call. Both measured on the installed build.
func jsxNoConstructedContextValuesIsProvider(ctx rule.Context, tagName *ast.Node) bool {
	if tagName == nil {
		return false
	}

	switch tagName.Kind {
	case ast.KindPropertyAccessExpression:
		name := tagName.AsPropertyAccessExpression().Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "Provider"

	case ast.KindIdentifier:
		return jsxNoConstructedContextValuesResolvesToCreateContext(ctx, tagName)
	}
	return false
}

// jsxNoConstructedContextValuesResolvesToCreateContext reports whether an identifier was declared
// as a variable initialized with a call to createContext.
//
// Upstream walks its own scope chain to the binding and reads the initializer. The checker answers
// the same question directly, which is the step 2b substitution: the decision reproduced exactly,
// the means different because we have a program and upstream has one file.
//
// The declarations are looped rather than indexed, because a symbol can carry more than one and an
// index-zero read goes silent on merged shapes. Upstream reads `defs[0]`; looping can only be
// wider, and for this predicate wider means "any declaration initializes it as a context", which is
// the same answer whenever only one declaration has an initializer at all.
func jsxNoConstructedContextValuesResolvesToCreateContext(ctx rule.Context, identifier *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
			continue
		}
		initializer := declaration.AsVariableDeclaration().Initializer
		if initializer == nil {
			continue
		}
		// The initializer is unwrapped in a loop rather than through `ast.SkipParentheses`, which
		// dereferences its argument; an initializer is optional and has already been checked for
		// nil above, but the loop shape is the one this tree standardized on after a nil panic cost
		// 167 files.
		for initializer.Kind == ast.KindParenthesizedExpression {
			initializer = initializer.AsParenthesizedExpression().Expression
			if initializer == nil {
				break
			}
		}
		if initializer == nil || initializer.Kind != ast.KindCallExpression {
			continue
		}
		if jsxNoConstructedContextValuesIsCreateContextCallee(
			initializer.AsCallExpression().Expression,
		) {
			return true
		}
	}
	return false
}

// jsxNoConstructedContextValuesIsCreateContextCallee accepts the two spellings of the factory.
//
// Upstream tests a bare `createContext` identifier, or a member expression whose object is named
// `React` and whose property is `createContext`. Both measured reporting.
func jsxNoConstructedContextValuesIsCreateContextCallee(callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	for callee.Kind == ast.KindParenthesizedExpression {
		callee = callee.AsParenthesizedExpression().Expression
		if callee == nil {
			return false
		}
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == "createContext"

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		receiver := access.Expression
		for receiver != nil && receiver.Kind == ast.KindParenthesizedExpression {
			receiver = receiver.AsParenthesizedExpression().Expression
		}
		if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "createContext"
	}
	return false
}

// jsxNoConstructedContextValuesValueExpression returns the expression inside a `value={...}`
// attribute, or nil.
//
// Three shapes are declined and each is a measured clean case: no `value` attribute at all, a
// boolean shorthand `value` with no initializer, and a string-literal `value="x"` which is not an
// expression container.
func jsxNoConstructedContextValuesValueExpression(attributes *ast.Node) *ast.Node {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return nil
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return nil
	}
	for _, property := range properties.Nodes {
		if property.Kind != ast.KindJsxAttribute {
			continue
		}
		attribute := property.AsJsxAttribute()
		name := attribute.Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "value" {
			continue
		}
		initializer := attribute.Initializer
		if initializer == nil || initializer.Kind != ast.KindJsxExpression {
			// A boolean shorthand, or a string literal. Both silent upstream.
			return nil
		}
		return initializer.AsJsxExpression().Expression
	}
	return nil
}

// jsxNoConstructedContextValuesRecursionLimit bounds the identifier-following walk.
//
// Upstream has no limit and relies on scope analysis terminating; following declarations through
// the checker can in principle cycle on `const a = a`, which the parser accepts. The bound is well
// above anything real: the deepest indirection measured on the installed build is two hops.
const jsxNoConstructedContextValuesRecursionLimit = 16

// jsxNoConstructedContextValuesConstructionOf is upstream's `isConstruction`, recursive.
//
// Every arm was measured against the installed build; see the rule doc for the table. The arms that
// recurse rather than answer are the interesting ones, and the order inside a conditional or a
// logical expression is observable: the LEFT side is tried first, so `c ? {a:1} : {b:2}` reports on
// `{a:1}`.
func jsxNoConstructedContextValuesConstructionOf(
	ctx rule.Context,
	node *ast.Node,
	depth int,
) *jsxNoConstructedContextValuesConstruction {
	if node == nil || depth > jsxNoConstructedContextValuesRecursionLimit {
		return nil
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		return &jsxNoConstructedContextValuesConstruction{kind: "object", node: node}

	case ast.KindArrayLiteralExpression:
		return &jsxNoConstructedContextValuesConstruction{kind: "array", node: node}

	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return &jsxNoConstructedContextValuesConstruction{kind: "function expression", node: node}

	case ast.KindClassExpression:
		return &jsxNoConstructedContextValuesConstruction{kind: "class expression", node: node}

	case ast.KindNewExpression:
		return &jsxNoConstructedContextValuesConstruction{kind: "new expression", node: node}

	case ast.KindRegularExpressionLiteral:
		// Upstream reaches this through `Literal` with a non-null `regex` field. Our parser gives a
		// regular expression its own kind, so the other literal kinds simply do not match here and
		// no separate test for them is needed.
		return &jsxNoConstructedContextValuesConstruction{kind: "regular expression", node: node}

	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		return &jsxNoConstructedContextValuesConstruction{kind: "JSX element", node: node}

	case ast.KindJsxFragment:
		return &jsxNoConstructedContextValuesConstruction{kind: "JSX fragment", node: node}

	case ast.KindParenthesizedExpression:
		// Upstream never sees this node: its parser folds parentheses away before the rule runs.
		// Recursing through it reproduces upstream's verdict on source upstream cannot represent,
		// and `({a:1})` as a value is ordinary in this tree.
		return jsxNoConstructedContextValuesConstructionOf(
			ctx, node.AsParenthesizedExpression().Expression, depth+1,
		)

	case ast.KindAsExpression, ast.KindSatisfiesExpression:
		// Upstream's `TSAsExpression` arm. `satisfies` did not exist when that arm was written and
		// takes the same shape, so it is treated the same; recorded here because it is an addition
		// rather than a reproduction.
		return jsxNoConstructedContextValuesConstructionOf(
			ctx, node.AsAsExpression().Expression, depth+1,
		)

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		if found := jsxNoConstructedContextValuesConstructionOf(ctx, conditional.WhenTrue, depth+1); found != nil {
			return found
		}
		return jsxNoConstructedContextValuesConstructionOf(ctx, conditional.WhenFalse, depth+1)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			// Upstream's AssignmentExpression arm: recurse into the right side, then relabel the
			// kind and set the whole assignment as the usage.
			found := jsxNoConstructedContextValuesConstructionOf(ctx, binary.Right, depth+1)
			if found == nil {
				return nil
			}
			return &jsxNoConstructedContextValuesConstruction{
				kind:  "assignment expression",
				node:  found.node,
				usage: node,
			}
		}
		if !jsxNoConstructedContextValuesIsLogicalOperator(binary.OperatorToken.Kind) {
			// Arithmetic, comparison and comma are all silent upstream, which only has a
			// LogicalExpression arm and no BinaryExpression one.
			return nil
		}
		if found := jsxNoConstructedContextValuesConstructionOf(ctx, binary.Left, depth+1); found != nil {
			return found
		}
		return jsxNoConstructedContextValuesConstructionOf(ctx, binary.Right, depth+1)

	case ast.KindPropertyAccessExpression:
		found := jsxNoConstructedContextValuesConstructionOf(
			ctx, node.AsPropertyAccessExpression().Expression, depth+1,
		)
		if found == nil {
			return nil
		}
		// Upstream sets the usage to the OBJECT rather than to the whole member expression.
		return &jsxNoConstructedContextValuesConstruction{
			kind:  found.kind,
			node:  found.node,
			usage: node.AsPropertyAccessExpression().Expression,
		}

	case ast.KindIdentifier:
		return jsxNoConstructedContextValuesFollowIdentifier(ctx, node, depth)
	}

	return nil
}

// jsxNoConstructedContextValuesIsLogicalOperator reports whether an operator is one upstream's
// LogicalExpression arm covers.
//
// ESTree's LogicalExpression is exactly these three; every other binary operator lands on a node
// type upstream's switch has no arm for and is therefore silent.
func jsxNoConstructedContextValuesIsLogicalOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindAmpersandAmpersandToken,
		ast.KindBarBarToken,
		ast.KindQuestionQuestionToken:
		return true
	}
	return false
}

// jsxNoConstructedContextValuesFollowIdentifier resolves an identifier to what it was declared as
// and recurses into that.
//
// Upstream reads its scope's binding, takes the LAST definition, and bails out unless that
// definition is a variable or a function name. A parameter is the case that bail-out exists for and
// it is measured silent.
//
// The checker answers the same question. `LocalSymbol` is consulted as a fallback because
// `GetSymbolAtLocation` is truncated on an exported declaration, which would otherwise go silent on
// `export const v = {a:1}`, a shape the corpus does not write and this tree does constantly.
func jsxNoConstructedContextValuesFollowIdentifier(
	ctx rule.Context,
	identifier *ast.Node,
	depth int,
) *jsxNoConstructedContextValuesConstruction {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return nil
	}

	declarations := symbol.Declarations
	if len(declarations) == 0 {
		return nil
	}

	// Upstream takes the LAST definition rather than the first, which matters for a name declared
	// more than once. Reproduced rather than looped, because this arm asks "what single thing is
	// this name" and a loop would be wider than upstream.
	declaration := declarations[len(declarations)-1]
	if declaration == nil {
		return nil
	}

	// Only a declaration evaluated each time the function using it runs is a construction. A
	// module-scope `const v = {...}` is evaluated once and keeps its identity forever, and so does a
	// declaration in a function enclosing the component (a factory that returns the component): the
	// component re-renders on its own state while the outer value stands still.
	//
	// Upstream gets the same answer by a narrower test: it resolves the name only in the scope the
	// JSX itself sits in (`callScope.set`, no walk up the chain). Measured on 7.37.5: a module-scope
	// object or function and an outer-function object are silent, and so is a component-level object
	// used by a provider inside an `if` block, which is a true positive upstream misses because the
	// block is its own scope. cohere compares functions rather than scopes, so it stays silent on the
	// first three and reports the fourth. The divergence is recorded in the rule's `.md`.
	if jsxNoConstructedContextValuesNearestFunction(declaration) != jsxNoConstructedContextValuesNearestFunction(identifier) {
		return nil
	}

	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return &jsxNoConstructedContextValuesConstruction{
			kind:  "function declaration",
			node:  declaration,
			usage: identifier,
		}

	case ast.KindVariableDeclaration:
		initializer := declaration.AsVariableDeclaration().Initializer
		if initializer == nil {
			return nil
		}
		found := jsxNoConstructedContextValuesConstructionOf(ctx, initializer, depth+1)
		if found == nil {
			return nil
		}
		return &jsxNoConstructedContextValuesConstruction{
			kind:  found.kind,
			node:  found.node,
			usage: identifier,
		}
	}

	// A parameter, an import, a class, anything else: upstream's bail-out.
	return nil
}

// jsxNoConstructedContextValuesInComponent reports whether a JSX element sits inside something
// upstream's component detector would accept.
//
// See the rule doc for the twenty measured rows this reproduces and for why the shelf's
// `scope.NameOf` cannot be used: it prefers a function expression's own name over its binding, and
// it answers a method's name where the enclosing class is what decides.
//
// The walk stops at the FIRST function-like ancestor, which is upstream's behaviour rather than an
// optimization: a provider inside a lowercase helper nested in a real component is silent, measured.
func jsxNoConstructedContextValuesInComponent(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration:
			name := current.AsFunctionDeclaration().Name()
			return name != nil && utilsreact.IsLikelyComponentName(name.Text())

		case ast.KindFunctionExpression, ast.KindArrowFunction:
			// A class field holding an arrow belongs to its class rather than to a binding, and the
			// arrow is nearer than the property declaration, so it would otherwise answer first.
			// Measured: `class C extends React.Component { f = () => <Ctx.Provider .../> }` reports.
			if current.Parent != nil && current.Parent.Kind == ast.KindPropertyDeclaration {
				return jsxNoConstructedContextValuesEnclosingClassIsComponent(current.Parent)
			}
			return jsxNoConstructedContextValuesBindingIsComponent(current)

		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor, ast.KindPropertyDeclaration:
			return jsxNoConstructedContextValuesEnclosingClassIsComponent(current)
		}
	}
	return false
}

// jsxNoConstructedContextValuesBindingIsComponent reads the name a function expression or arrow is
// bound to.
//
// Upstream's detector attributes an anonymous function to whatever it is assigned to, and prefers
// that binding even when the function expression carries its own name. Both directions measured:
// `const thing = function Component(){}` is silent and `const Component = function thing(){}`
// reports.
//
// A call wrapper is transparent, so `React.memo(() => ...)` and `forwardRef(() => ...)` reach the
// binding one level further out. Measured reporting for both.
func jsxNoConstructedContextValuesBindingIsComponent(function *ast.Node) bool {
	parent := function.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		name := parent.AsVariableDeclaration().Name()
		return name != nil && name.Kind == ast.KindIdentifier &&
			utilsreact.IsLikelyComponentName(name.Text())

	case ast.KindCallExpression:
		return jsxNoConstructedContextValuesBindingIsComponent(parent)
	}
	return false
}

// jsxNoConstructedContextValuesEnclosingClassIsComponent reports whether the class owning a member
// is a React component class.
//
// The heritage decides and nothing else does: measured, a member of a class extending
// `React.Component` reports whatever the member is named, a class with no heritage is silent even
// from `render`, and the class's own name is irrelevant.
func jsxNoConstructedContextValuesEnclosingClassIsComponent(member *ast.Node) bool {
	for current := member.Parent; current != nil; current = current.Parent {
		if ast.IsClassLike(current) {
			return utilsreact.IsEs6ComponentClass(current)
		}
	}
	return false
}
