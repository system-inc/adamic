package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
)

// React detection: does this function hold JSX or call a hook?
//
// The answer decides whether a PascalCase function is a component, and every rule about components
// rests on it. It is also where the tailwind port went wrong: JSX element names are plain
// identifiers in typescript-go's AST, and reading them as bindings produced 3,081 spurious findings
// (commit 02c720b). The lesson generalizes past that rule. An exemption the original gets free from
// its AST has to be written down explicitly in ours, because the two ASTs disagree about what a JSX
// name even is.

// jsxWalkDepthLimit bounds the search, matching the original's limit of 20.
//
// The bound is behavior rather than a safety valve, which is why it is reproduced rather than
// dropped. A deeply nested component whose JSX sits past the limit reads as "not a component" to the
// gate cohere replaces, and a Go walk with no limit would find it and report a finding the gate
// does not. Parity is the acceptance criterion, so the limit ports with the rule.
const jsxWalkDepthLimit = 20

// HasJsxOrReactHookCalls reports whether a subtree holds JSX or calls a React hook.
//
// The original walks a fixed list of eleven property names rather than every child, and reproducing
// that bound matters: an exhaustive walk finds JSX in places the original never descends into, and
// every such find is a finding the gate does not have.
//
// Naming the node kinds that carry those properties was the wrong way to reproduce it, and measuring
// is what caught it. The list looked complete and missed if, switch, for and try, because those
// carry their branches on exactly the properties the original does walk: an IfStatement's consequent
// and alternate, a loop's and a try's body. One component out of 128 was lost to that gap, and it
// was the one whose JSX all sat inside if statements.
//
// So this excludes rather than enumerates. Everything is walked except the few places the original's
// property list genuinely cannot reach, which is the same set stated as a boundary instead of as an
// inventory, and it fails toward the original rather than away from it.
func HasJsxOrReactHookCalls(node *ast.Node) bool {
	return searchForJsxOrHook(node, 0)
}

func searchForJsxOrHook(node *ast.Node, depth int) bool {
	if node == nil || depth > jsxWalkDepthLimit {
		return false
	}

	switch node.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true
	case ast.KindCallExpression:
		if react.IsHookCall(node.AsCallExpression()) {
			return true
		}
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if !descendsForJsxSearch(child) {
			return false
		}
		if searchForJsxOrHook(child, depth+1) {
			found = true
		}
		return found
	})
	return found
}

// descendsForJsxSearch reports whether the walk enters a child.
//
// One exclusion, and it earns its place by measurement rather than by sounding prudent.
//
// An earlier version of this also skipped type nodes, parameters, decorators, heritage clauses and
// import declarations, on the reasoning that JSX under a type is not a component body. Every one of
// those was speculative: removing them changed nothing in the fixtures and nothing across 3,400
// files, because the shapes they guard against do not occur. Code that cannot be shown to do
// anything is code that can only diverge from the original later, so it is gone. The original has no
// type-position check at all, and parity is the acceptance criterion.
func descendsForJsxSearch(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {

	// A switch statement's arms are genuinely unreachable to the original, and this is the one
	// exclusion that is a fact about its list rather than a judgment. ESTree puts the arms on
	// SwitchStatement.cases, and "cases" is not among the eleven names it walks. So a component that
	// returns JSX only from switch arms does not read as a component to the gate.
	//
	// Measured, not assumed: walking into them found two real components the gate does not report,
	// and both are switch-dispatch components. The rule is parity, so the blind spot ports with it.
	// SwitchCase.consequent would be reachable if anything reached the case, and nothing does.
	case ast.KindSwitchStatement:
		return false
	}
	return true
}

// IsLikelyReactComponent reports whether a function-like node looks like a React component.
//
// This is a different question from HasJsxOrReactHookCalls and the two are not interchangeable,
// which is worth stating plainly because reaching for the wrong one produced a false positive on the
// live tree. That one asks "does this subtree contain JSX anywhere", walking a bounded property
// list. This one is the original's `isLikelyReactComponent`, and it asks two much narrower things:
//
//	the first parameter is named `properties` or `props`, or
//	a top-level return in the body returns JSX, or a ternary with a JSX branch
//
// The parameter-name arm is the load-bearing half and it is checked first, before any JSX is looked
// for. A component whose JSX is unreachable to a bounded walk is still a component under this test
// as long as it takes `properties`, which is exactly the case that caught this: a switch-dispatch
// component returns JSX only from switch arms, and the property walk cannot reach a switch arm.
//
// The consequence differs by which question a rule is asking. A rule asking "is this a component,
// flag it" loses a finding when detection misses, which is a quiet parity gap. A rule asking "does
// this file export a component by name" gains a false finding, because the component it failed to
// see was the exported one. Only the second polarity turns a detection miss into a report.
func IsLikelyReactComponent(functionLike *ast.Node) bool {
	if functionLike == nil {
		return false
	}

	// The kind check is not defensive, it is reachable, and a fixture found it by crashing.
	//
	// Callers pass a variable's initializer, which is any expression at all: `const Colors = {...}`
	// hands this an object literal. `Parameters()` and `Body()` are switches over function-like
	// kinds and dereference data that an object literal does not have, so a nil check on their
	// results is too late. The guard has to be on the kind, before either call.
	switch functionLike.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration:
	default:
		return false
	}

	// The parameter arm. `function Thing(properties)` is a component by convention here regardless
	// of what its body does, which is what makes it reach shapes a JSX search cannot.
	if parameters := functionLike.Parameters(); len(parameters) > 0 {
		first := parameters[0]
		if first != nil {
			name := first.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				if text := name.Text(); text == "properties" || text == "props" {
					return true
				}
			}
		}
	}

	body := functionLike.Body()
	if body == nil {
		return false
	}

	// A block body: only statements directly in it are examined, never nested ones. The original
	// iterates `body.body` and looks at each statement's own kind, so a return inside an if, a loop
	// or a switch arm is not reached. That bound is behavior rather than an oversight, and widening
	// it here would find components the gate does not.
	if body.Kind == ast.KindBlock {
		found := false
		for _, statement := range body.AsBlock().Statements.Nodes {
			if statement.Kind != ast.KindReturnStatement {
				continue
			}
			if returnArgumentLooksLikeJsx(statement.AsReturnStatement().Expression) {
				found = true
				break
			}
		}
		return found
	}

	// An expression-bodied arrow function: the body is the returned value.
	return returnArgumentLooksLikeJsx(body)
}

// returnArgumentLooksLikeJsx reports JSX, a fragment, or a ternary with a JSX branch.
//
// # Parentheses have to be skipped, and this is not a detail
//
// `return (\n    <div />\n);` is how almost every component in this codebase is written, and it
// parses here as a ParenthesizedExpression wrapping the JSX. ESTree builds no node for parentheses,
// so the original sees the JSX element directly and never has to think about it.
//
// Without the skip, only the unparenthesized single-line form reads as a component. Measured on the
// live tree before the skip was added: fifteen files reported, every one of them a file that does
// export its component by name, because the exported component wrapped its JSX in parentheses and
// therefore did not read as a component at all.
//
// The ternary branches are unwrapped for the same reason: `return flag ? (<a />) : (<b />)` is
// ordinary formatting.
func returnArgumentLooksLikeJsx(expression *ast.Node) bool {
	// The nil check comes before SkipParentheses, which dereferences its argument. A bare `return;`
	// carries no expression, and that is an ordinary shape rather than an exotic one.
	if expression == nil {
		return false
	}
	expression = ast.SkipParentheses(expression)
	if expression == nil {
		return false
	}
	if isJsxValue(expression) {
		return true
	}
	if expression.Kind == ast.KindConditionalExpression {
		conditional := expression.AsConditionalExpression()
		return jsxAfterParentheses(conditional.WhenTrue) || jsxAfterParentheses(conditional.WhenFalse)
	}
	return false
}

// jsxAfterParentheses reports JSX once any wrapping parentheses are removed, guarding the nil that
// SkipParentheses would dereference.
func jsxAfterParentheses(expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	return isJsxValue(ast.SkipParentheses(expression))
}

// isJsxValue reports a JSX element or fragment.
//
// A self-closing element is included where the original names only JSXElement, because ESTree has no
// separate self-closing kind: `<div />` is a JSXElement there with no children, and typescript-go
// gives it its own kind. Omitting it would make `return <div />` not read as a component, which is
// the single most common component shape there is.
func isJsxValue(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindJsxElement, ast.KindJsxFragment, ast.KindJsxSelfClosingElement:
		return true
	}
	return false
}
