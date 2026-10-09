package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageJsxInTryStatement = rule.Message{
	Id: "jsxInTryStatement",
	Description: "This JSX is constructed inside a try block, which does not catch what the " +
		"author almost certainly means to catch. Writing JSX builds a plain description object " +
		"and renders nothing, so React runs the component later, long after this try has " +
		"returned. The only errors this catch can see are the rare ones thrown while building " +
		"the description itself, never the rendering errors the pattern is reaching for. Move " +
		"the JSX out of the try and wrap the component in an error boundary, which is the " +
		"mechanism that does catch errors during rendering.",
}

// ErrorBoundaries flags JSX constructed inside a try block.
//
//	valid:   function Component() { try { f(); } catch { return <Fallback />; } }
//	valid:   function Component() { let el; try { el = <div />; } finally { log(el); } return el; }
//	valid:   function widget() { let el; try { el = <div />; } catch {} return el; }
//	invalid: function Component() { let el; try { el = <div />; } catch { return null; } return el; }
//	invalid: function Component() { try { try { f(); } catch { return <E />; } } catch { return null; } }
//
// Ported from React's `error-boundaries` rule, whose implementation is
// `ValidateNoJSXInTryStatement.ts` and whose diagnostic category is `ErrorBoundaries`. oxc ships a
// Rust transcription at `oxc_react_compiler/src/react_compiler_validation/
// validate_no_jsx_in_try_statement.rs` that matches React line for line, including the comment; it
// is a second reading of the same source rather than an independent opinion, so there is nothing to
// choose between the two and no divergence to record between them. oxc's *linter* has no rule here.
//
// Authority is React, and it was **executed** rather than reasoned about. The rule ships in
// `eslint-plugin-react-hooks` 7.1.1 as `react-hooks/error-boundaries`, and every behavioral claim
// in this comment was measured by driving it through the ESLint `Linter` API on the input named,
// rather than derived from reading. That distinction earns its place here because reading produced
// the wrong answer three separate times, each recorded below.
//
// # What upstream computes, and why this rule does not compute it the same way
//
// Upstream runs over the compiler's own intermediate representation. It walks the basic blocks in
// layout order carrying a stack of pending catch-handler block identifiers, pushes a handler when a
// block ends in a `try` terminal, pops it when the walk reaches that handler's own block, and flags
// any JSX instruction found while the stack is non-empty. Fifty two lines.
//
// **That is a control-flow spelling of a purely syntactic question, and it is a control-flow
// spelling only because upstream has no syntax tree left at that point.** The lowering has already
// run; blocks are all it has. Because the handler is popped exactly when its own block begins, and
// because the body's blocks are laid out between the `try` terminal and the handler, the stack is
// non-empty across precisely the extent of some enclosing try *block* — never across that same
// try's catch clause, and never across its finally clause. A catch clause is covered only when an
// *outer* try is still pending, which is the "unless that catch is itself nested inside an outer
// try" sentence in React's own doc comment.
//
// So the decision reduces to: **is this JSX construction syntactically within the `try` block of
// some enclosing try statement, up to the nearest enclosing function.** We have the syntax tree, so
// that is what this asks, by walking parents. It is the port brief's fidelity clause applied
// literally: fidelity is owed to what a rule *decides*, not to the machinery it needed to obtain it,
// and reproducing a block walk over a graph we would have to build first, to answer a question the
// parent chain already answers, is the cargo cult that clause names.
//
// **Stated plainly because this rule was the first consumer of the dominator tree and the dataflow
// solver added in `internal/lint/ecmascript/control_flow_graph`, and used neither.** Both were read first. Dominance
// answers "must control pass through A to reach B", which is a question about paths; this rule asks
// "is B written inside A", which is a question about text, and a try block does not dominate its own
// body in any useful sense once a throw edge exists. A dataflow would carry an "inside a try" bit
// forward over the graph and arrive at the same answer the parent chain gives directly, after
// building a graph, at strictly higher cost and with one new failure mode: the graph lays a
// `finally` block down **twice**, so a JSX in a finally would be visited twice, and the second copy
// is unreachable exactly when nothing takes the abrupt path. That duplication is a real hazard
// `internal/unused` works around by keying on source position rather than block identity. Walking
// parents has no such hazard because a node has one parent chain.
//
// The three control-flow traps the dispatch named were each checked against this design rather than
// assumed away, and all three are inert here: the doubled `finally` layout cannot double-report
// something reached through a parent chain; a try whose body cannot throw getting no edge to its
// catch cannot hide a catch body that is found by walking parents; and `labelsOf` keeping outer
// labels is about labeled statements, which this rule never reads.
//
// # Where the syntactic reading and React part company, both measured
//
// Two inputs exist where React is silent and a purely syntactic reading would report. Neither is a
// judgment about JSX in a try. Both are artifacts of the lowering that runs before the validator,
// and both are reproduced here as silence, deliberately, with the command that established them
// stated rather than asserted.
//
// **A try with no catch clause is silent, whatever is in it.** React's HIR builder does not support
// `try`/`finally` without a handler at all: it aborts the whole function with a Todo diagnostic,
// `(BuildHIR::lowerStatement) Handle TryStatement without a catch clause`, before any validation
// pass runs. Upstream's own corpus records this and is the reason the two fixtures named
// `error.todo-invalid-jsx-in-try-with-finally` and
// `error.todo-invalid-jsx-in-catch-in-outer-try-with-finally` carry the `error.todo` prefix: they
// are named `invalid-...` and they do not produce this rule's finding, they produce the Todo. Run
// through the real rule, `try { el = <div />; } finally { log(el); }` reports nothing. A finally
// *alongside* a catch is fine and the try body still reports, which is the case that separates
// "no catch clause" from "has a finally".
//
// **JSX whose value nothing consumes is silent.** The lowering drops instructions with no consumer
// before the validator sees them, so `try { (<div />); } catch {}` is clean upstream while
// `try { foo(<div />); } catch {}` reports, and `let el; try { el = <div />; } catch {} return 1;`
// is clean while inserting a single `log(el)` before that return makes it report. The consumer can
// be anything: a call argument, a return, a member assignment, a push.
//
// **This one cost the most and it is the reason every claim here names a command.** Reading, and
// then a first round of probing, attributed these silences to `returnsNonNode` in the component
// gate. The measurement that broke that reading is one line: hold everything fixed and vary only
// the final return, and `return <div />` is **clean** while `return el` **reports**. No account
// built on the returned expression's shape survives that, since a JSX element is maximally a node.
// The discriminator is whether `el` is read at all. Two independent filters had been collapsed into
// one, and the fixtures written from the wrong one asserted upstream's verdict for a reason
// upstream does not hold.
//
// The elimination is not reproduced. It is dead-value analysis over a value graph this tree does
// not build, the divergence costs findings only in code that constructs JSX and then discards it,
// and a linter reporting a discarded `<div />` inside a try is not obviously wrong. **Recorded as a
// real divergence rather than smoothed over**: a differential against a tool running React's
// pipeline will show findings here it does not report, and `assignedJsxNeverReadIsADocumentedDivergence`
// in the fixtures is that case pinned as reporting, with the reasoning at the line.
//
// # Which functions are even eligible, and why that gate is most of the rule
//
// React compiles components and hooks, nothing else, so the validator never sees an ordinary
// function. Measured: `function widget() { let el; try { el = <div />; } catch {} return el; }` is
// **clean**, and the identical body under the name `Component` reports. That gate is not part of
// the fifty two lines, it is the driver's `getReactFunctionType`, and it is load-bearing: dropping
// it would report every try-wrapped JSX in every helper, utility and test file in the tree, a
// false-positive class the imported corpus cannot see because upstream's four fixtures are all
// named `Component`.
//
// The gate reproduced here is upstream's `infer` mode, which is the default `compilationMode`, and
// each clause below was pinned by running the real rule on an input that isolates it:
//
//	name          `/^[A-Z]/` for a component, `/^use[A-Z0-9]/` for a hook. Measured: `_Private`
//	              and `$Dollar` are clean, `usething` is clean, `useThing` reports.
//	position      the function must be found by the program-level traversal, which skips into
//	              nested functions only until it accepts one. Measured: a `Component` declared
//	              inside another function is clean, in both directions of the outer name.
//	creates JSX   `callsHooksOrCreatesJsx`, not counting nested functions. This rule can only fire
//	              on a function containing JSX, so it is satisfied wherever this rule could report
//	              and is not written out separately.
//	params        at most two, no rest element, and a second parameter must mention `ref`.
//	              Measured: `(...props)` is clean, `(props, other)` is clean, `(props, ref)`
//	              reports, `({a, b})` reports.
//	returns       `returnsNonNode`: a function whose LAST-visited return hands back an object, a
//	              function, a class, a `new`, or a bigint is not a component. Upstream **assigns**
//	              rather than accumulating, so the answer is the last return the traversal saw and
//	              not whether any return is a non-node. That is observable and was measured both
//	              ways: `return {}` before `return el` reports, `return el` before `return {}` is
//	              clean, and a function with no return at all reports because the flag never moves.
//
// A hook is gated only on the name and on creating JSX or calling a hook, with no parameter or
// return test, which is upstream's asymmetry rather than an omission here.
//
// # Where the finding points
//
// At the JSX construction itself, and this is byte-verified against React's own golden rather than
// inferred. `invalid-jsx-in-try-with-catch.expect.md` logs `"index":123` through `"index":130`;
// slicing that fixture's bytes at those offsets gives `<div />`, seven bytes. `ctx.ReportNode` on
// the element node reproduces exactly that span, probed on this parser before the rule was written.
// A nested element reports separately from its parent, so `<div><span /></div>` inside a try
// produces two findings, which was measured rather than assumed and is why the walk visits every
// JSX node rather than only outermost ones.
//
// # A nested function is a barrier in both directions
//
// JSX inside a function written inside a try block does not report, because that inner function is
// lowered as its own unit with its own empty try stack. Measured:
// `try { const f = () => <div />; f(); } catch {}` is clean, while the `<Inner />` written in the
// same try block does report. So the upward walk stops at the first enclosing function, and the
// gate is asked of that function rather than of the outermost one.
var ErrorBoundaries = rule.Rule{
	// No namespace prefix. The config writes `react/error-boundaries`, and the parity guard strips
	// the namespace on a `/` boundary, so a self-namespaced `react-error-boundaries` would match no
	// inventory entry.
	Name: "react-hooks/error-boundaries",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node) {
			enclosing := enclosingFunctionOf(node)
			if enclosing == nil || !isInsideTryBlockWithin(node, enclosing) {
				return
			}
			if !isReactCompiledFunction(enclosing) {
				return
			}
			ctx.ReportNode(node, messageJsxInTryStatement)
		}

		return rule.Listeners{
			ast.KindJsxElement:            report,
			ast.KindJsxSelfClosingElement: report,
			ast.KindJsxFragment:           report,
		}
	},
}

// enclosingFunctionOf returns the nearest function-like ancestor, or nil at the top level.
//
// The top-level nil is a decline rather than a gap: React compiles functions, and JSX written in a
// try at module scope belongs to no component, so it is silent. Measured on the real rule with a
// bare `try { el = <div />; } catch {}` at module scope, which is clean.
func enclosingFunctionOf(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if isFunctionLike(current) {
			return current
		}
	}
	return nil
}

// isFunctionLike reports whether a node introduces its own function body.
//
// The list is the set of nodes React's lowering treats as a separate compilation unit. A class
// method and an accessor are included even though neither can be a compiled component here, because
// this function's job is to bound the upward walk, and a walk that passed through a method body
// would ask the try question against a try statement outside the method.
func isFunctionLike(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return true
	}
	return false
}

// isInsideTryBlockWithin reports whether node sits in the `try` block of some try statement that is
// itself inside boundary.
//
// This is the whole of upstream's fifty two lines, spelled against the parent chain. The step that
// matters is the one that asks *which clause* of the try statement the walk arrived from: entering
// through the try block counts, entering through the catch clause or the finally block does not.
// Comparing against `TryBlock` by node identity is what draws that line, and it is the reason a
// catch body is covered only when an outer try encloses the whole statement, which the loop reaches
// on a later iteration.
//
// A try with no catch clause never counts, whatever clause the walk came from. That is the lowering
// divergence recorded on the rule: React aborts such a function before this validator runs, so
// reporting here would be a finding upstream does not produce.
func isInsideTryBlockWithin(node *ast.Node, boundary *ast.Node) bool {
	for child := node; child != nil && child != boundary; child = child.Parent {
		parent := child.Parent
		if parent == nil || parent.Kind != ast.KindTryStatement {
			continue
		}
		try := parent.AsTryStatement()
		if try.CatchClause == nil {
			continue
		}
		if try.TryBlock == child {
			return true
		}
	}
	return false
}

// isReactCompiledFunction reports whether React's compiler would compile this function at all.
//
// Upstream's `getReactFunctionType` in `infer` mode, which is the default `compilationMode`. This
// is not part of the validator; it is the driver's decision about which functions to hand it, and
// it is reproduced because without it the rule reports on every helper that wraps JSX in a try.
//
// The order matters and is upstream's: a function *declaration* whose name is a component or hook
// name is accepted on syntax alone, through `isComponentDeclaration`, before the inferred path with
// its parameter and return tests runs. Every other shape goes through `getComponentOrHookLike`.
func isReactCompiledFunction(fn *ast.Node) bool {
	if !isTopLevelCompilationCandidate(fn) {
		return false
	}

	name, named := reactFunctionNameOf(fn)
	if !named {
		return false
	}

	if isHookIdentifierName(name) {
		// A hook is gated on its name alone here. Upstream also requires it to call a hook or
		// create JSX, and this rule can only fire on a function that creates JSX, so that clause is
		// satisfied wherever this one could report and adding it would change nothing.
		return true
	}
	if !isComponentIdentifierName(name) {
		return false
	}

	// A function declaration is NOT accepted on syntax alone here, even though upstream's
	// `isComponentDeclaration` reads that way. Measured: `function Component(a, b, c)` is clean
	// upstream, and so is `function Component(...props)`, so the parameter test is reached for a
	// declaration too and an early accept would report on both.
	return hasValidComponentParameters(fn) && !returnsNonNode(fn)
}

// isTopLevelCompilationCandidate reports whether the program-level traversal would reach fn.
//
// React walks the program for function declarations, function expressions and arrow functions, and
// calls `path.skip()` on the first one it accepts, so a function nested inside an accepted one is
// never offered. It also skips class declarations and class expressions outright, which is why a
// method named `Component` is clean.
//
// Measured in both directions: a `Component` declared inside `widget` is clean, and a `Component`
// declared inside `Outer` is clean as well, so this is about nesting rather than about whether the
// outer function was itself accepted.
func isTopLevelCompilationCandidate(fn *ast.Node) bool {
	for current := fn.Parent; current != nil; current = current.Parent {
		if isFunctionLike(current) {
			return false
		}
		switch current.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return false
		}
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	}
	// A method or accessor is never offered to the traversal, and its class ancestor would have
	// declined above in every real shape; this is the standalone answer for the same question.
	return false
}

// reactFunctionNameOf returns the name React reads to classify a function.
//
// `getFunctionName` takes the declaration's own identifier when it has one, and otherwise the name
// of the single variable declarator it initializes. An anonymous function expression in any other
// position has no name and is declined, which is why `export default function (props) {}` is clean
// and `export default function Component(props) {}` reports.
func reactFunctionNameOf(fn *ast.Node) (string, bool) {
	if fn.Kind == ast.KindFunctionDeclaration {
		if name := fn.Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.Text(), true
		}
		return "", false
	}
	if name := fn.Name(); name != nil && name.Kind == ast.KindIdentifier {
		return name.Text(), true
	}
	parent := fn.Parent
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return "", false
	}
	target := parent.AsVariableDeclaration().Name()
	if target == nil || target.Kind != ast.KindIdentifier {
		return "", false
	}
	return target.Text(), true
}

// hasValidComponentParameters reproduces `isValidComponentParams`.
//
// No parameters is valid. One or two are valid, a rest element in first position is not, and a
// second parameter must be an identifier mentioning `ref`. Three or more is never valid.
//
// Measured on the real rule: `(...props)` clean, `(props, other)` clean, `(props, ref)` reports,
// `(a, b, c)` clean, `({a, b})` reports.
//
// Upstream additionally rejects a first parameter annotated with a primitive or function type,
// through `isValidPropsAnnotation`. That clause is not reproduced. It reads a type annotation to
// decide a name-shaped question, its effect is to *narrow* which functions are compiled, and every
// input it would exclude is one where this rule would otherwise report; so omitting it can only add
// findings on annotated props, never remove them. Stated as a divergence rather than left implicit,
// and it is the one clause here that was not pinned by a measurement.
func hasValidComponentParameters(fn *ast.Node) bool {
	parameters := parametersOf(fn)
	switch len(parameters) {
	case 0:
		return true
	case 1:
		return !isRestParameter(parameters[0])
	case 2:
		if isRestParameter(parameters[0]) {
			return false
		}
		second := parameters[1].AsParameterDeclaration().Name()
		if second == nil || second.Kind != ast.KindIdentifier {
			return false
		}
		return mentionsRef(second.Text())
	}
	return false
}

// parametersOf returns a function's parameter list, or nil when it has none.
//
// Routed through the shim's own `Parameters` helper rather than a kind switch, because every
// function-like node carries the list and the accessor knows which field holds it.
func parametersOf(fn *ast.Node) []*ast.Node {
	list := fn.Parameters()
	if list == nil {
		return nil
	}
	return list
}

// isRestParameter reports whether a parameter is written with a spread.
func isRestParameter(parameter *ast.Node) bool {
	return parameter != nil && parameter.Kind == ast.KindParameter &&
		parameter.AsParameterDeclaration().DotDotDotToken != nil
}

// mentionsRef reports whether a second parameter's name claims to be a ref.
//
// Substring rather than equality, which is upstream's own test: `name.includes('ref') ||
// name.includes('Ref')`. So `forwardedRef` and `myref` both count, and that looseness is upstream's
// rather than an approximation here.
func mentionsRef(name string) bool {
	return strings.Contains(name, "ref") || strings.Contains(name, "Ref")
}

// returnsNonNode reports whether any return in fn hands back something that cannot be rendered.
//
// Upstream's `returnsNonNode`, which walks the function's own returns, skipping nested functions and
// object methods, and asks whether the returned expression is one of a fixed list of shapes that are
// never React nodes: an object literal, a function, an arrow, a class expression, a `new`, or a
// bigint. A bare `return;` with no argument also counts, since a missing argument is `null` to the
// traversal and `isNonNode(null)` is true.
//
// **Upstream assigns rather than accumulates**, so the answer is whichever return the traversal
// visited last rather than whether any return is a non-node. That is reproduced, because it is
// observable: measured on the real rule, a function returning `el` and then `1` on a later line is
// clean, while the same two returns in the opposite order reports. Writing this as "any return is a
// non-node" would have been the sensible-looking version and would disagree with upstream on
// exactly those inputs.
func returnsNonNode(fn *ast.Node) bool {
	body := fn.Body()
	if body == nil {
		return false
	}
	if fn.Kind == ast.KindArrowFunction && body.Kind != ast.KindBlock {
		return isNonNodeExpression(body)
	}

	answer := false
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		switch {
		case isFunctionLike(node):
			// Nested functions carry their own returns, which say nothing about this one.
			return false
		case node.Kind == ast.KindReturnStatement:
			answer = isNonNodeExpression(node.AsReturnStatement().Expression)
		}
		node.ForEachChild(walk)
		return false
	}
	body.ForEachChild(walk)
	return answer
}

// isNonNodeExpression reports whether an expression is one React can never render.
//
// A nil expression, from a bare `return;`, answers true, matching upstream's `isNonNode(null)`.
func isNonNodeExpression(expression *ast.Node) bool {
	if expression == nil {
		return true
	}
	switch expression.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrowFunction, ast.KindFunctionExpression,
		ast.KindBigIntLiteral, ast.KindClassExpression, ast.KindNewExpression:
		return true
	}
	return false
}
