package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageForwardRefUsesRef = rule.Message{
	Id: "forwardRefUsesRef",
	Description: "A function wrapped in `forwardRef` takes only one parameter, so the ref React " +
		"passes as the second one is dropped on the floor. Every consumer writing " +
		"`<Component ref={r} />` gets a ref that never attaches, and it fails silently rather " +
		"than erroring. Accept the second parameter and forward it to whatever should receive " +
		"it, or drop the `forwardRef` wrapper, which does nothing for a component that " +
		"ignores the ref.",
}

var messageForwardRefAddRefParameter = rule.Message{
	Id:          "forwardRefAddRefParameter",
	Description: "Add a `ref` parameter.",
}

var messageForwardRefRemoveWrapper = rule.Message{
	Id:          "forwardRefRemoveWrapper",
	Description: "Remove the `forwardRef` wrapper.",
}

// ForwardRefUsesRef flags a `forwardRef` call whose function takes exactly one parameter.
//
//	valid:   forwardRef((props, ref) => null)
//	valid:   forwardRef(() => {})
//	valid:   forwardRef(function (a, b, c) {})
//	valid:   forwardRef(function (...a) {})
//	invalid: forwardRef((props) => null)
//	invalid: React.forwardRef(function Component(props) { return null; })
//
// Ported from `react/forward-ref-uses-ref`, read against oxc's `forward_ref_uses_ref.rs` and
// measured against the release oxlint binary on inputs the imported corpus does not contain.
//
// # The name says "uses ref" and the rule never asks whether the ref is used
//
// This is the first thing to get wrong and the corpus settles it in one case. Upstream counts
// parameters. It does not look at the body, does not check whether the second parameter is read,
// and does not care what it is called:
//
//	forwardRef((props, ref) => null)      passes, and `ref` is never referenced
//	forwardRef((props, whatever) => null) passes, and the parameter is not named `ref` at all
//
// A port implementing the name rather than the behavior would look for a reference to the second
// binding, and the first of those two upstream pass cases would fail. The rule is a parameter-arity
// check wearing a use-analysis name.
//
// # Exactly one, which makes zero and three both silent
//
// The condition is `parameters_count() != 1`, so the reportable shape is narrow on both sides.
// `forwardRef(() => {})` and `forwardRef(function (a, b, c) {})` are both upstream pass cases. Only
// one is wrong, because only one is unambiguously a component that forgot the ref: none is a
// component that takes no props, and more than two is not a shape upstream is willing to guess at.
//
// A rest parameter disqualifies separately, and it has to, because oxc counts rest outside the
// parameter list. `forwardRef((a, ...b) => {})` has a `parameters_count()` of one and is silent
// only because `params.rest.is_some()` short-circuits first. Our tree carries rest *inside*
// `Parameters()`, so reproducing this means excluding it explicitly rather than inheriting the
// behavior, and the two spellings would diverge on exactly that input. Measured silent on oxlint.
//
// A TypeScript `this` parameter is likewise not counted, so `forwardRef(function (this: any, a) {})`
// reports: `a` is the only real parameter. oxc gets that from its own parameter model; we get it
// from `ast.IsThisParameter`. Also measured on oxlint rather than assumed.
//
// # What counts as a `forwardRef` call, which is looser than the plugin's name suggests
//
// Upstream matches `CallExpression::callee_name()`, which is a bare identifier's name or a member
// expression's static property name, and nothing else. There is no import tracking and no receiver
// check whatsoever, so all of these report, measured on oxlint:
//
//	NotReact.forwardRef(function (a) {})    any receiver at all
//	obj.deep.forwardRef(function (a) {})    any depth
//	React["forwardRef"](function (a) {})    a computed static string
//	(React).forwardRef(function (a) {})     parens on the object are irrelevant
//
// and an aliased import is silent, because the name at the call site is what is read:
//
//	import { forwardRef as fr } from 'react'; fr(function (a) {})
//
// `internal/utilities/react/IsNamespacedMember` is the helper whose name matches this question and it
// is the wrong one twice over: it hardcodes its receiver to `React` alone, which would silence the
// four receivers above, and it requires a property access, which would silence the bare
// `forwardRef(...)` form that is most of the corpus. Its body was read rather than its name.
//
// # Parentheses, measured in three places because they fall two different ways
//
// No imported fixture writes a parenthesized anything, so each of these is pinned by a fixture
// below and every one was run against the release binary rather than reasoned about.
//
// Upstream destructures the callee and the argument directly, with no paren skipping, so wrapping
// either one makes the rule blind:
//
//	(forwardRef)(function (a) {})           silent
//	(React.forwardRef)(function (a) {})     silent
//	forwardRef((function (a) {}))           silent, the argument is parenthesized
//
// But the *statement-context* question below runs through `outermost_paren_parent`, which does see
// through parens, so there the answer flips:
//
//	(forwardRef(function (a) {}));          reports, and offers only the add-ref suggestion
//
// Reproducing all four means declining parens at the callee and the argument while skipping them
// when walking to the parent. That asymmetry is upstream's, not ours, and a port applying paren
// skipping uniformly in either direction is wrong on three of these four inputs.
//
// # Two suggestions, or one, decided by whether removing the wrapper would break the file
//
// Both repairs are suggestions rather than fixes: each changes what the code means. Removing the
// wrapper changes the component's identity, and adding a parameter changes its signature.
//
// The wrapper-removal suggestion is withheld for exactly one shape: an *anonymous* function
// expression whose call sits directly in an expression statement. Unwrapping there would leave
// `function (a) {}` as a statement, which does not parse. Every other shape gets both.
//
// Upstream guards only the anonymous case, and the named one is a real wart that is reproduced
// rather than improved on. `forwardRef(function Named(a) {});` at statement level offers a removal
// producing `function Named(a) {};`, which parses but is now a hoisted function *declaration*
// rather than an expression. That is a meaning change beyond the one intended, and it is a
// suggestion a human reads before applying, which is the only reason it is tolerable. Measured by
// running oxlint with suggestions applied rather than by reading the branch.
//
// # No options and no type checker
//
// oxc's struct is a unit struct with no configuration parsing, and its single tester block carries
// no option tuples, so the surface is empty rather than merely unused. Every question here is
// syntactic: which name the callee spells, how many parameters the function has, what kind the
// parent node is. Nothing asks what an identifier binds to, so nothing declares `NeedsTypeChecker`,
// and declaring it would not let the rule answer the aliased-import case anyway, because upstream
// does not ask that question either.
var ForwardRefUsesRef = rule.Rule{
	Name: "react/forward-ref-uses-ref",
	Run: func(ctx rule.Context, _ any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()

				// The callee is read without skipping parentheses, because upstream destructures
				// it directly. `(forwardRef)(fn)` is silent upstream and is silent here.
				if calleeName(call.Expression) != "forwardRef" {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}

				// Likewise unparenthesized: `forwardRef((function (a) {}))` is silent upstream.
				// A spread argument is not an expression upstream and is not a function-like kind
				// here, so it declines through the same switch below.
				first := call.Arguments.Nodes[0]

				var parameters []*ast.Node
				switch first.Kind {
				case ast.KindArrowFunction:
					parameters = parameterNodesOf(first.AsArrowFunction().Parameters)
				case ast.KindFunctionExpression:
					parameters = parameterNodesOf(first.AsFunctionExpression().Parameters)
				default:
					return
				}

				if !hasExactlyOneCountedParameter(parameters) {
					return
				}

				ctx.ReportNodeWithSuggestions(first, messageForwardRefUsesRef,
					suggestionsFor(ctx, node, first, parameters)...)
			},
		}
	},
}

// calleeName answers what upstream's `CallExpression::callee_name` answers.
//
// A bare identifier gives its own name; a member access gives the static property name, which is
// the identifier after the dot or a string literal in brackets. Everything else, including a
// parenthesized callee and a computed subscript that is not a literal, answers nothing. There is no
// receiver check at any point, which is upstream's behavior rather than an omission here.
func calleeName(callee *ast.Expression) string {
	if callee == nil {
		return ""
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text()
	case ast.KindPropertyAccessExpression:
		// `Text()` panics on this kind, so the name node is read through the guarded accessor
		// rather than off the access expression itself.
		name, ok := ast.TryGetTextOfPropertyName(callee.AsPropertyAccessExpression().Name())
		if !ok {
			return ""
		}
		return name
	case ast.KindElementAccessExpression:
		// A no-substitution template is a static name upstream just as a string literal is, and
		// leaving it out was a real defect here that no imported fixture could see: the corpus
		// writes only `forwardRef` and `React.forwardRef`. A surviving mutant on this branch is
		// what surfaced it, and oxlint confirms ``React[`forwardRef`](function (a) {})`` reports.
		//
		// Anything else, including a bare identifier subscript, answers no name. That matters more
		// than it looks: `Text()` on an identifier would happily return `forwardRef` for
		// `React[forwardRef]`, which oxlint is silent on, so the kind check is the whole guard
		// rather than a formality.
		argument := callee.AsElementAccessExpression().ArgumentExpression
		if argument == nil {
			return ""
		}
		if argument.Kind != ast.KindStringLiteral &&
			argument.Kind != ast.KindNoSubstitutionTemplateLiteral {
			return ""
		}
		return argument.Text()
	}
	return ""
}

// parameterNodesOf flattens a parameter list, tolerating the nil an empty list can carry.
func parameterNodesOf(list *ast.NodeList) []*ast.Node {
	if list == nil {
		return nil
	}
	return list.Nodes
}

// hasExactlyOneCountedParameter reproduces `params.parameters_count() != 1 || params.rest.is_some()`.
//
// The two exclusions exist because oxc's parameter model differs from ours in both directions. A
// rest element lives outside oxc's parameter list and inside ours, so `(a, ...b)` counts as one
// there and two here, and upstream is silent on it for the separate reason that `rest.is_some()`
// short-circuits. A `this` parameter is absent from oxc's list and present in ours, so
// `function (this: any, a)` counts as one there and two here, and upstream reports it. Counting the
// list directly would get both backwards, in opposite directions, and no imported fixture writes
// either shape.
func hasExactlyOneCountedParameter(parameters []*ast.Node) bool {
	counted := 0
	for _, parameter := range parameters {
		if parameter.AsParameterDeclaration().DotDotDotToken != nil {
			return false
		}
		if ast.IsThisParameter(parameter) {
			continue
		}
		counted++
	}
	return counted == 1
}

// suggestionsFor builds the repairs, in upstream's order: remove the wrapper first where it is
// offered at all, then add the parameter.
//
// Order is load-bearing rather than cosmetic. An engine applying "the suggestion" without a human
// choosing takes the first, and upstream's is the removal, so reversing them would silently change
// what an unattended run writes.
func suggestionsFor(
	ctx rule.Context,
	call *ast.Node,
	function *ast.Node,
	parameters []*ast.Node,
) []rule.Suggestion {
	addRef := rule.Suggestion{
		Message: messageForwardRefAddRefParameter,
		Fixes:   []rule.Fix{addRefParameterFix(ctx, function, parameters)},
	}
	if !canRemoveForwardRef(function, call) {
		return []rule.Suggestion{addRef}
	}
	return []rule.Suggestion{
		{
			Message: messageForwardRefRemoveWrapper,
			Fixes: []rule.Fix{
				// `ctx.ReplaceNode` rather than `rule.ReplaceRange`, on both halves. The
				// call's own `Pos()` starts at its leading trivia, so replacing that span
				// swallows the space in `const x = forwardRef(...)` and writes
				// `const x =function(...)`. The node form routes through `TokenRange`, which
				// trims it. The replacement text is sliced the same way for the same reason:
				// the function's `Pos()` is just inside the call's `(`, so any space there
				// would be carried into the unwrapped result.
				ctx.ReplaceNode(call, functionSourceText(ctx, function)),
			},
		},
		addRef,
	}
}

// canRemoveForwardRef answers whether unwrapping would still parse.
//
// Only one shape fails: an anonymous function expression directly inside an expression statement,
// where the unwrapped `function (a) {}` would be read as a declaration missing its name. A named
// function expression is *not* guarded upstream even though unwrapping turns it into a declaration,
// which is reproduced deliberately. The walk to the parent skips parentheses, matching
// `outermost_paren_parent`, so `(forwardRef(function (a) {}));` is still statement context.
func canRemoveForwardRef(function *ast.Node, call *ast.Node) bool {
	if function.Kind != ast.KindFunctionExpression {
		return true
	}
	if function.AsFunctionExpression().Name() != nil {
		return true
	}
	return !isInExpressionStatement(call)
}

// isInExpressionStatement walks out through parentheses to the first node that is not one.
func isInExpressionStatement(node *ast.Node) bool {
	current := node.Parent
	for current != nil && current.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	return current != nil && current.Kind == ast.KindExpressionStatement
}

// addRefParameterFix rewrites the parameter list to end with `, ref)`.
//
// The span is the parameter list including its parentheses where it has them, which an arrow with a
// single unparenthesized parameter does not. Upstream rebuilds the text rather than inserting,
// because it has to normalize three things at once: a trailing `)` that is stripped and re-added, a
// trailing comma with any whitespace after it, and a missing opening paren. `forwardRef(a => {})`
// becoming `forwardRef((a, ref) => {})` needs all three, and `function(a, ) {}` needs the middle
// one, which is why upstream ships fixture pairs for both.
func addRefParameterFix(ctx rule.Context, function *ast.Node, parameters []*ast.Node) rule.Fix {
	span := parameterListRange(ctx, function, parameters)
	text := ctx.SourceFile.Text()[span.Pos():span.End()]
	text = strings.TrimSuffix(text, ")")
	text = strings.TrimRight(text, " \t\r\n")
	text = strings.TrimSuffix(text, ",")
	if !strings.HasPrefix(text, "(") {
		text = "(" + text
	}
	return rule.ReplaceRange(span, text+", ref)")
}

// parameterListRange finds the span upstream calls `params.span`.
//
// Our tree has no node for the parameter list, so the range is recovered from the source. It runs
// from the opening parenthesis, when there is one, to its matching close, and otherwise covers just
// the single bare parameter of an arrow. Scanning back from the first parameter rather than forward
// from the function keeps `function Component(a)` from finding the paren-free stretch before the
// name, and scanning forward from the last parameter to the next `)` is safe because a default
// value or a type annotation is inside the last parameter's own range.
func parameterListRange(ctx rule.Context, function *ast.Node, parameters []*ast.Node) core.TextRange {
	text := ctx.SourceFile.Text()

	// `Pos()` on the first parameter starts at its leading trivia rather than at the parameter, so
	// both branches below have to step over that whitespace rather than trusting it. Getting this
	// wrong is invisible in the common spelling and shows up only when the source has a space where
	// most code does not: `forwardRef( a => {})` would otherwise produce `forwardRef(( a, ref) =>
	// {})`, inserting the opening paren before the space instead of after it.
	start := parameters[0].Pos()
	for start < len(text) && isSpaceByte(text[start]) {
		start++
	}

	// An opening paren immediately before the first parameter means the list is parenthesized, so
	// the span covers the parens too. Scanning back from the parameter rather than forward from the
	// function keeps `function Component(a)` from matching the paren-free stretch before the name.
	open := start
	for open > function.Pos() && isSpaceByte(text[open-1]) {
		open--
	}
	if open > function.Pos() && text[open-1] == '(' {
		end := parameters[len(parameters)-1].End()
		for end < len(text) && text[end] != ')' {
			end++
		}
		if end < len(text) {
			end++
		}
		return core.NewTextRange(open-1, end)
	}

	// A lone arrow parameter with no parens. oxc's `params.span` starts at the parameter itself,
	// measured from the diagnostic label offset on `const x = forwardRef( a => {});`, which oxlint
	// reports at the `a` rather than at the space before it.
	return core.NewTextRange(start, parameters[len(parameters)-1].End())
}

func isSpaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// functionSourceText is the unwrapped function's own text, without leading trivia.
//
// Split out rather than inlined because the trivia trim has to match the one on the span being
// replaced: a slice from `Pos()` would carry whatever whitespace sits after the call's `(`.
func functionSourceText(ctx rule.Context, function *ast.Node) string {
	span := rule.TokenRange(ctx.SourceFile, function)
	return ctx.SourceFile.Text()[span.Pos():span.End()]
}
