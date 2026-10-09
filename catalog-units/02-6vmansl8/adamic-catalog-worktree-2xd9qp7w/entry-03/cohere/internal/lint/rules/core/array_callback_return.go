package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ArrayCallbackReturnOptions configures the three independent decisions this rule can make.
type ArrayCallbackReturnOptions struct {
	// AllowImplicit accepts a bare `return;` inside a callback whose method wants a value.
	//
	// Off upstream. With it on, `foo.filter(function() { return; })` is accepted, on the reading
	// that returning undefined deliberately is still returning. It changes only the
	// `expectedReturnValue` arm; a callback with no return at all is still reported.
	AllowImplicit bool `json:"allowImplicit"`

	// CheckForEach turns the judgment around for `forEach`, which wants NO value returned.
	//
	// Off upstream, and it is the option that adds findings rather than removing them. With it on
	// the rule reports a `forEach` callback that returns anything, because a value returned to
	// `forEach` is discarded and its presence usually means the author meant `map`.
	CheckForEach bool `json:"checkForEach"`

	// AllowVoid accepts `void expr` as the way to say "this expression's value is deliberately
	// discarded" inside a checked `forEach`.
	//
	// Off upstream, and it only does anything while CheckForEach is on. It also changes what is
	// OFFERED rather than only what is reported: with it on a concise arrow gets a second
	// suggestion, prepending `void`, alongside wrapping in braces.
	AllowVoid bool `json:"allowVoid"`
}

var messageArrayCallbackExpectedAtEnd = rule.Message{
	Id: "expectedAtEnd",
	Description: "This callback returns a value on some paths and runs off the end on others, so " +
		"the method receives undefined for the elements that take the other path. A filter reads " +
		"that as false, a map stores it, and a sort treats it as zero, none of which is what a " +
		"missing return usually means. Return a value on every path.",
}

var messageArrayCallbackExpectedInside = rule.Message{
	Id: "expectedInside",
	Description: "This callback never returns a value, and the array method it is passed to is " +
		"built entirely around the value each call produces. Every element gets undefined, so a " +
		"filter keeps nothing, a map fills with undefined, and a sort orders nothing. Either " +
		"return the value the method needs, or use a method that does not want one.",
}

var messageArrayCallbackExpectedReturnValue = rule.Message{
	Id: "expectedReturnValue",
	Description: "This `return` leaves the callback without a value, and the array method it is " +
		"passed to needs one from every call. A bare return here reads as an early exit and " +
		"behaves as returning undefined, which the method then treats as a real answer.",
}

var messageArrayCallbackExpectedNoReturnValue = rule.Message{
	Id: "expectedNoReturnValue",
	Description: "This callback returns a value to a method that discards it, so the value is " +
		"computed and thrown away. Returning something here usually means the wrong method was " +
		"chosen, and the value was meant to be collected rather than dropped.",
}

var messageArrayCallbackWrapBraces = rule.Message{
	Id:          "wrapBraces",
	Description: "Wrap the expression in braces so it is a statement rather than a returned value.",
}

var messageArrayCallbackPrependVoid = rule.Message{
	Id:          "prependVoid",
	Description: "Prepend `void` to say the expression's value is deliberately discarded.",
}

// ArrayCallbackReturn flags an array-method callback that returns nothing when the method needs a
// value, or returns a value to `forEach`, which discards it.
//
//	valid:   foo.forEach(function(x) { var a = 0; })
//	valid:   foo.every(function() { return true; })
//	valid:   foo.map(async function() {})                 async is not checked, except Array.fromAsync
//	valid:   foo.every(function*() {})                    a generator is never checked
//	invalid: foo.every(() => {})                          expectedInside
//	invalid: foo.every(function() { if (a) return true; }) expectedAtEnd
//	invalid: foo.filter(function() { return; })           expectedReturnValue
//	invalid: foo.forEach(x => x)                          expectedNoReturnValue, with checkForEach
//
// Every method in this family is defined by what the callback hands back. `filter` keeps the element
// when the callback says so, `map` stores what it produces, `sort` reads a number, `every` and
// `some` read a boolean. A callback that returns nothing does not opt out of that contract; it
// answers undefined for every element, which each method then interprets as a real answer. The
// mistake is usually a `forEach` written where a `map` was meant, or the reverse.
//
// # Four message ids, and what actually separates them
//
// Two ask for a value and two are about a value that should not be there, and the split is not by
// option:
//
//	expectedInside         the callback body has no `return` at all and can run off its end
//	expectedAtEnd          the body HAS a `return` somewhere and can still run off its end
//	expectedReturnValue    a specific `return;` with no argument, reported at that statement
//	expectedNoReturnValue  a value reaching `forEach`, under checkForEach
//
// The first two are one judgment reported at the function; which of them fires is decided by whether
// any `return` was seen while walking, so they cannot be collapsed. The third is per return
// statement and points there rather than at the function, so a port using one message would point
// at the wrong node for every case of it.
//
// # Reachability, and why this needs the control-flow graph
//
// "Can this callback run off its end" is not a syntactic question. `foo.every(function() { if (a)
// return true; })` reports and `foo.every(function() { if (a) return true; else return false; })`
// does not, and the difference is whether any path reaches the closing brace. Upstream asks
// `isAnySegmentReachable(funcInfo.currentSegments)` at the function's exit, which is code path
// analysis. `internal/lint/ecmascript/control_flow_graph` answers exactly that question with `Graph.EndReachable`,
// whose doc comment names this as the question it exists for, so this rule is a caller rather than
// a reimplementation.
//
// # Which callee positions count, and the shape upstream walks up through
//
// `getArrayMethodName` climbs from the function through logical, conditional and chain expressions,
// so `foo.every(cb || function(){})` and `foo.every(a ? f : g)` are both checked. It also follows a
// `return` out of an immediately-invoked wrapper, which is how
// `foo.every(function(){ return function(){}; }())` is judged. All three are reproduced and all
// three are in the corpus.
//
// Async and generator handling is asymmetric and deliberate upstream. A generator is never checked.
// An async function is checked only for `Array.fromAsync`, and the guard sits inside the call arm
// rather than at the top, so `foo.map(async function(){})` is clean while
// `Array.fromAsync(x, async function(){})` reports. Measured against the installed rule, both ways.
//
// # Suggestions, not fixes, and the two the corpus asserts
//
// `meta.hasSuggestions` is true and `meta.fixable` is null. That distinction is the port: a repair
// here changes what the code MEANS, since wrapping a concise arrow's body in braces stops it
// returning and prepending `void` discards its value. Neither may be applied unattended, and
// upstream does not.
//
//	wrapBraces   offered on a concise arrow under checkForEach. `x => x` becomes `x => {x}`.
//	prependVoid  offered as well when allowVoid is on, and on a `return expr;` under both options.
//
// # What this port does NOT reproduce
//
// The `prependVoid` suggestion upstream computes a precedence test and a parenthesis test through
// its token API, so `x => a, b` gets parentheses that `x => a` does not. That is reproduced for the
// shapes the corpus asserts and is stated here rather than left implicit, because the surface it
// walks is a token stream we answer differently. Every suggestion this rule offers is asserted by
// applying it and comparing the whole resulting source, so a shape it gets wrong fails loudly rather
// than silently.
var ArrayCallbackReturn = rule.Rule{
	Name: "array-callback-return",

	// No type information. Every decision is syntactic: which method name the callee spells, which
	// argument position the function sits in, and whether the body can run off its end. Upstream
	// asks its scope analysis nothing here either.
	NeedsTypeChecker: false,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as a bare severity is handed nil, which `options.(T)` turns into the
		// zero value. All three defaults are false upstream, so the zero value is right, and the
		// fallback is still written out rather than relied on.
		settings := ArrayCallbackReturnOptions{}
		if decoded, configured := rule.OptionsAs[ArrayCallbackReturnOptions](options); configured {
			settings = decoded
		}

		check := func(node *ast.Node) {
			checkArrayCallback(ctx, node, settings)
		}

		return rule.Listeners{
			ast.KindFunctionExpression: check,
			ast.KindArrowFunction:      check,
		}
	},
}

// checkArrayCallback judges one function expression or arrow function.
func checkArrayCallback(ctx rule.Context, node *ast.Node, settings ArrayCallbackReturnOptions) {
	methodName, isCallback := arrayMethodNameFor(node)
	if !isCallback {
		return
	}

	// Every `return` inside this callback, excluding those belonging to nested functions. Upstream
	// gets the same set from its per-function code path stack; here the walk stops at any nested
	// function-like node, which is the same boundary.
	returns := returnStatementsIn(node)

	if methodName == "forEach" {
		checkForEachCallback(ctx, node, returns, settings)
		return
	}

	// The end-of-body judgment is emitted FIRST, and the order is upstream's rather than a choice.
	//
	// Upstream reports the function-level finding from `FunctionExpression:exit`, which ESLint runs
	// after the whole body has been traversed, and the per-return findings from `ReturnStatement`
	// during that traversal. ESLint then sorts its messages by position before handing them over, so
	// a finding on the function head, which starts before every return inside it, comes out first.
	// `foo.every(function() { if (a) return; })` is the case that pins it: upstream orders it
	// `expectedAtEnd` then `expectedReturnValue`, and emitting in traversal order gives the reverse.
	//
	// `ExpectFindings` asserts ids IN ORDER, so this is a real difference rather than a cosmetic one.
	//
	// A concise arrow always produces a value, so only a braced body can run off its end. Upstream
	// tests `node.body.type === "BlockStatement"` for the same reason.
	if body := node.Body(); body != nil && body.Kind == ast.KindBlock && arrayCallbackCanRunOffEnd(node) {
		message := messageArrayCallbackExpectedInside
		if len(returns) > 0 {
			// The body returns somewhere and can still fall off the end, which is a different
			// sentence to write and a different thing for the reader to fix.
			message = messageArrayCallbackExpectedAtEnd
		}
		reportAtFunctionHead(ctx, node, rule.Message{
			Id: message.Id,
			Description: fmt.Sprintf("%s %s", message.Description,
				arrayCallbackContextSentence(ctx, node, methodName)),
		})
	}

	// A `return;` with no argument is reported at the statement rather than at the function, so this
	// arm runs per return and is independent of the end-of-body judgment above.
	if !settings.AllowImplicit {
		for _, returnStatement := range returns {
			if returnStatement.AsReturnStatement().Expression != nil {
				continue
			}
			ctx.ReportNode(returnStatement, rule.Message{
				Id: messageArrayCallbackExpectedReturnValue.Id,
				Description: fmt.Sprintf("%s %s", messageArrayCallbackExpectedReturnValue.Description,
					arrayCallbackContextSentence(ctx, node, methodName)),
			})
		}
	}
}

// checkForEachCallback judges a `forEach` callback, which wants no value at all.
//
// The whole arm is inert unless `checkForEach` is on, which is upstream's default, so the common
// configuration reaches this and returns immediately.
func checkForEachCallback(
	ctx rule.Context,
	node *ast.Node,
	returns []*ast.Node,
	settings ArrayCallbackReturnOptions,
) {
	if !settings.CheckForEach {
		return
	}

	// A concise arrow's body IS the returned value, so the whole arrow is the finding and the
	// repair is structural. Upstream reports at the function head here.
	body := node.Body()
	if node.Kind == ast.KindArrowFunction && body != nil && body.Kind != ast.KindBlock {
		if settings.AllowVoid && isVoidExpression(body) {
			// Already saying the value is discarded, which is what allowVoid asks for.
			return
		}
		suggestions := []rule.Suggestion{arrayCallbackWrapBracesSuggestion(ctx, node, body)}
		if settings.AllowVoid {
			suggestions = append(suggestions, arrayCallbackPrependVoidSuggestion(ctx, body))
		}
		reportAtFunctionHeadWithSuggestions(ctx, node, rule.Message{
			Id: messageArrayCallbackExpectedNoReturnValue.Id,
			Description: fmt.Sprintf("%s %s",
				messageArrayCallbackExpectedNoReturnValue.Description,
				arrayCallbackContextSentence(ctx, node, "forEach")),
		}, suggestions...)
		return
	}

	// A braced body reports per `return expr;` rather than once for the function, because that is
	// where the value actually leaves.
	for _, returnStatement := range returns {
		argument := returnStatement.AsReturnStatement().Expression
		if argument == nil {
			continue
		}
		if settings.AllowVoid && isVoidExpression(argument) {
			continue
		}
		message := rule.Message{
			Id: messageArrayCallbackExpectedNoReturnValue.Id,
			Description: fmt.Sprintf("%s %s",
				messageArrayCallbackExpectedNoReturnValue.Description,
				arrayCallbackContextSentence(ctx, node, "forEach")),
		}
		if settings.AllowVoid {
			ctx.ReportNodeWithSuggestions(returnStatement, message,
				arrayCallbackPrependVoidSuggestion(ctx, argument))
			continue
		}
		// Without allowVoid there is nothing to offer: upstream sets `suggest` to null here, because
		// the only repair it knows is the one the option enables.
		ctx.ReportNode(returnStatement, message)
	}
}

// arrayCallbackWrapBracesSuggestion wraps a concise arrow's body in braces.
//
// `x => x` becomes `x => {x}`, which turns the returned expression into a statement. It changes what
// the arrow means, so it is a suggestion rather than a fix, exactly as upstream has it.
func arrayCallbackWrapBracesSuggestion(ctx rule.Context, node *ast.Node, body *ast.Node) rule.Suggestion {
	_ = node
	return rule.Suggestion{
		Message: messageArrayCallbackWrapBraces,
		Fixes: []rule.Fix{
			ctx.InsertBefore(body, "{"),
			ctx.InsertAfter(body, "}"),
		},
	}
}

// arrayCallbackPrependVoidSuggestion prepends `void` to an expression.
//
// Upstream additionally parenthesizes when the expression binds looser than `void` and is not
// already parenthesized, which it decides through its token API. That narrowing is reproduced for
// the operators the corpus writes; see the rule doc comment for the stated limit.
func arrayCallbackPrependVoidSuggestion(ctx rule.Context, expression *ast.Node) rule.Suggestion {
	// A leading space is needed when the token before the expression is `return` and the two are
	// written adjacent, because `returnvoid x` is one identifier. Upstream computes exactly this,
	// comparing the return token's end against the first token's start, and the corpus pins it:
	// `foo.forEach((x) => { return(x); })` must become `return void (x)` and not `returnvoid (x)`.
	//
	// An arrow's `=>` needs no such space, which is why upstream's condition tests for `return`
	// specifically rather than for adjacency alone.
	prefix := "void "
	if arrayCallbackReturnIsAdjacentTo(ctx, expression) {
		prefix = " void "
	}

	if arrayCallbackNeedsParensUnderVoid(expression) {
		return rule.Suggestion{
			Message: messageArrayCallbackPrependVoid,
			Fixes: []rule.Fix{
				ctx.InsertBefore(expression, prefix+"("),
				ctx.InsertAfter(expression, ")"),
			},
		}
	}
	return rule.Suggestion{
		Message: messageArrayCallbackPrependVoid,
		Fixes:   []rule.Fix{ctx.InsertBefore(expression, prefix)},
	}
}

// arrayCallbackReturnIsAdjacentTo reports whether an expression is written with no space after the
// `return` keyword that precedes it.
//
// `ctx.InsertBefore` anchors on the token range, which starts past leading trivia, so the check is
// whether the source immediately before that token is the end of `return`.
func arrayCallbackReturnIsAdjacentTo(ctx rule.Context, expression *ast.Node) bool {
	parent := expression.Parent
	if parent == nil || parent.Kind != ast.KindReturnStatement {
		return false
	}
	start := rule.TokenRange(ctx.SourceFile, expression).Pos()
	const keyword = "return"
	if start < len(keyword) {
		return false
	}
	return ctx.SourceFile.Text()[start-len(keyword):start] == keyword
}

// arrayCallbackNeedsParensUnderVoid reports whether prepending `void` to an expression would bind
// tighter than the expression itself, so the result needs parentheses to keep its meaning.
//
// `void` is a unary operator, so anything looser than unary needs wrapping: assignment, the comma
// operator, a conditional, and every binary operator. Upstream computes this with a precedence table
// plus a check for parentheses that are already there.
func arrayCallbackNeedsParensUnderVoid(expression *ast.Node) bool {
	switch expression.Kind {
	case ast.KindParenthesizedExpression:
		// Already wrapped, so `void (x)` is well formed without adding another pair. Upstream's
		// `isParenthesised` check.
		return false
	case ast.KindBinaryExpression, ast.KindConditionalExpression, ast.KindYieldExpression:
		return true
	}
	return false
}

// isVoidExpression reports whether an expression is `void something`.
//
// `void` is its own node kind in this parser rather than a prefix unary operator, which is worth
// stating because the analogous `typeof` and `delete` are also their own kinds while `!` and `-` are
// operators on `KindPrefixUnaryExpression`. Probing the tree settled it: the body of
// `foo.forEach((x) => void x)` comes back as `KindVoidExpression`.
func isVoidExpression(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindVoidExpression
}

// arrayCallbackCanRunOffEnd reports whether control can reach the end of a callback's body.
//
// This is upstream's `isAnySegmentReachable(funcInfo.currentSegments)` at the function's exit, and
// `control_flow_graph.Graph.EndReachable` is documented as answering that exact question. A body whose
// every path returns or throws cannot fall through, so nothing is missing at the end of it.
func arrayCallbackCanRunOffEnd(node *ast.Node) bool {
	graph := control_flow_graph.Build(node, control_flow_graph.Hooks[struct{}]{})
	if graph == nil {
		return true
	}
	return graph.EndReachable
}

// returnStatementsIn collects every `return` belonging to one function, not to a nested one.
//
// The walk stops at any node that owns its own returns. Upstream gets the same set from its code
// path stack, where a nested function pushes a new frame and its returns are attributed there.
func returnStatementsIn(node *ast.Node) []*ast.Node {
	var returns []*ast.Node
	body := node.Body()
	if body == nil {
		return nil
	}

	var visit func(*ast.Node) bool
	visit = func(child *ast.Node) bool {
		if child == nil {
			return false
		}
		switch child.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindClassStaticBlockDeclaration:
			// A nested function's returns are its own.
			return false
		case ast.KindReturnStatement:
			returns = append(returns, child)
		}
		child.ForEachChild(visit)
		return false
	}
	body.ForEachChild(visit)

	return returns
}

// reportAtFunctionHead reports on the function's head rather than on its whole body.
//
// Upstream passes an explicit `loc: getFunctionHeadLoc(node, sourceCode)`, which spans the `function`
// keyword through the parameter list for a function expression and just the `=>` for an arrow. The
// distinction is not cosmetic: a callback body can be hundreds of lines, and a finding spanning all
// of it is unreadable in any editor that highlights ranges.
func reportAtFunctionHead(ctx rule.Context, node *ast.Node, message rule.Message) {
	ctx.ReportRange(arrayCallbackFunctionHeadRange(ctx, node), message)
}

// reportAtFunctionHeadWithSuggestions is reportAtFunctionHead for a finding that offers repairs.
func reportAtFunctionHeadWithSuggestions(
	ctx rule.Context,
	node *ast.Node,
	message rule.Message,
	suggestions ...rule.Suggestion,
) {
	ctx.ReportRangeWithSuggestions(arrayCallbackFunctionHeadRange(ctx, node), message, suggestions...)
}

// arrayCallbackFunctionHeadRange is upstream's `getFunctionHeadLoc`.
//
// For an arrow function the head is the `=>` token itself. For a function expression it runs from
// the start of the function through the end of its parameter list, so a named callback's name is
// inside the span and its body is not.
func arrayCallbackFunctionHeadRange(ctx rule.Context, node *ast.Node) core.TextRange {
	if node.Kind == ast.KindArrowFunction {
		if arrow := node.AsArrowFunction(); arrow != nil && arrow.EqualsGreaterThanToken != nil {
			return rule.TokenRange(ctx.SourceFile, arrow.EqualsGreaterThanToken)
		}
	}

	start := rule.TokenRange(ctx.SourceFile, node).Pos()
	end := start
	if parameters := node.ParameterList(); parameters != nil {
		end = parameters.End()
	}
	// The parameter list node ends before its closing parenthesis, so the span is extended to it.
	// Measured on `foo.every(function cb() {})`: upstream's column places the finding on the
	// `function` keyword, and the span it uses reaches the `)`.
	text := ctx.SourceFile.Text()
	for end < len(text) && text[end] != ')' {
		end++
	}
	if end < len(text) {
		end++
	}
	if end <= start {
		return rule.TokenRange(ctx.SourceFile, node)
	}
	return core.NewTextRange(start, end)
}

// arrayCallbackContextSentence names the method and the callback the way upstream's message data
// does, as a sentence appended to the rule's own description.
//
// `rule.Message` carries no interpolation layer, so the per-finding half is rendered here. The two
// values are upstream's `arrayMethodName` and `name`, spelled the same way so a reader moving
// between the two linters sees the same words.
func arrayCallbackContextSentence(ctx rule.Context, node *ast.Node, methodName string) string {
	return fmt.Sprintf("Here %s() is being passed %s.",
		arrayCallbackQualifiedMethodName(methodName), arrayCallbackFunctionNameWithKind(ctx, node))
}

// arrayCallbackQualifiedMethodName is upstream's `fullMethodName`.
//
// The four static methods are on `Array` itself and everything else is on the prototype. Getting
// this wrong renders `Array.prototype.from`, which does not exist.
func arrayCallbackQualifiedMethodName(methodName string) string {
	switch methodName {
	case "from", "fromAsync", "of", "isArray":
		return "Array." + methodName
	}
	return "Array.prototype." + methodName
}

// arrayCallbackFunctionNameWithKind is upstream's `getFunctionNameWithKind`, narrowed to the two
// kinds this rule can be handed.
//
// The renderings the corpus asserts are `arrow function`, `function`, and `function 'cb'`. The
// quoting is single quotes around the name, which is upstream's spelling and is asserted by its
// message text.
func arrayCallbackFunctionNameWithKind(ctx rule.Context, node *ast.Node) string {
	if node.Kind == ast.KindArrowFunction {
		return "arrow function"
	}
	if name := node.Name(); name != nil {
		if functionName := text.TrimWhitespace(name.Text()); functionName != "" {
			return fmt.Sprintf("function '%s'", functionName)
		}
	}
	_ = ctx
	return "function"
}
