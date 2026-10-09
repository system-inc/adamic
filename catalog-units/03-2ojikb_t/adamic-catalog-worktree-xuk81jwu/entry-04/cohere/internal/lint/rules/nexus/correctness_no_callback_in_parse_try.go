package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoCallbackInParseTryId = "callbackInParseTry"

// correctnessNoCallbackInParseTryText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-callback-in-parse-try.json`.
var correctnessNoCallbackInParseTryText = policy.MessageOf("nexus/correctness-no-callback-in-parse-try", correctnessNoCallbackInParseTryId)

func correctnessNoCallbackInParseTryMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoCallbackInParseTryId,
		Description: correctnessNoCallbackInParseTryText.Render(nil),
	}
}

// CorrectnessNoCallbackInParseTry reports a call to a caller-supplied function inside a `try` whose
// block also parses JSON, when the `catch` discards the caught error.
//
//	invalid: try { const message = JSON.parse(line); onMessage(message); } catch { onUnparseable?.(line); }
//	invalid: try { onEvent(JSON.parse(line)); } catch { /* heartbeat or partial frame */ }
//	invalid: try { callback(null, JSON.parse(text)); } catch { callback(new Error('bad JSON')); }
//	invalid: try { return parseJsonOrThrow(read(), 'settings', isSettings); } catch { return null; }
//	valid:   let message: unknown; try { message = JSON.parse(line); } catch { onUnparseable?.(line); continue; } onMessage(message);
//	valid:   const lineParse = parseJson(line, 'stream line', isMessage); if(lineParse.outcome === 'Parsed') onMessage(lineParse.value);
//	valid:   try { onMessage(JSON.parse(line)); } catch(error) { if(!(error instanceof SyntaxError)) throw error; }
//	valid:   new Promise(function(resolve, reject) { try { resolve(JSON.parse(text)); } catch { reject(new Error('bad')); } });
//
// # Where it came from
//
// Two sites in ahra, found by the own-history pass of the new-rules sweep (`#tevhg3f`, probe P5) and
// built as task `#hspse4v`:
//
//   - `modules/claude/utilities/ClaudeUtilities.ts`, `parseStreamLines` (`#wv045mc`): one `try` held
//     the `JSON.parse` of each stream line and the call to the caller's `onMessage`, and the `catch`
//     sent the line to `onUnparseable`. A handler that threw on a well-formed line was reported as a
//     malformed line, and its own error was gone.
//   - `modules/x/XStreamApi.ts:45`, `consumeStream`: the same shape, a new copy, with an empty
//     `catch` commented "Heartbeat or partial frame; skip." Every error `onEvent` threw vanished.
//
// Both were fixed in the JSON waves (`4e385b61`, `1382cbe9`) by moving to `parseJson` and calling the
// callback only on a `'Parsed'` outcome, outside any `try`.
//
// # What the three parts mean
//
// **The parse** is a call to `JSON.parse` or to nexus `parseJsonOrThrow`, decided by the symbol the
// callee resolves to, never its spelling: `parse` declared on the `JSON` interface in a global
// declaration file (lib.es5), or a function `parseJsonOrThrow` declared in nexus's
// `source/structured-text/json/Json.ts`. Those two throw on malformed text; `parseJson` returns an
// outcome and never throws, so a `try` around it is not guarding a parse.
//
// **The callback** is a call whose callee is a bare identifier (through parentheses, `!` and `?.`)
// resolving to a parameter, or to a name destructured out of one, whose type has a call signature
// once `undefined` and `null` are set aside. That is the caller's code, which the function cannot
// know never throws. A method on a parameter (`logger.info`, `response.json`) is not a callback and
// is not counted; neither is an `any` parameter, which has no call signature to prove it is a
// function. The resolve and reject functions of a `new Promise` executor are exempt: the platform
// guarantees they never throw, so a `catch` around them cannot misread their error.
//
// **The catch that reports a parse error.** Whether a `catch` *says* "parse error" lives in its
// message text, and reading text is a name heuristic. What the AST does say exactly is whether the
// `catch` can tell the callback's error from the parse error at all: a `catch` with no binding, or
// with a binding it never reads, throws the caught value away, so whatever it does it does for both,
// and the only failure it was written for is the parse the `try` guards. That is the shape reported.
// A `catch` that reads its binding (rethrows it, tests `instanceof SyntaxError`, hands it to
// `onError`, logs it) is left alone, even when it goes on to label it a parse failure; see the
// `.md` for why that is declined rather than guessed at.
//
// # Where the calls must sit
//
// Both calls are looked for in the code the `catch` actually guards: the `try` block, not descending
// into a nested function (a closure may run after the `try` has finished) and not into the block of
// a nested `try` that has its own `catch` (that `catch` sees the error first). The nested `try`'s
// `catch` and `finally` blocks are walked, since an error thrown there does reach the outer `catch`.
// The order of the two calls inside the block does not matter: a callback run before the parse is
// misread just the same.
//
// # No fix
//
// Moving the callback out of the `try` changes which errors propagate, and the caller may have come
// to rely on the swallowing. Each site is the author's call.
var CorrectnessNoCallbackInParseTry = rule.Rule{
	Name: "nexus/correctness-no-callback-in-parse-try",

	// Symbol identity tells the platform's `JSON.parse` from a local `parse`, nexus's
	// `parseJsonOrThrow` from any function of that name, and a parameter from a local of its name.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindTryStatement: func(node *ast.Node) {
				correctnessNoCallbackInParseTryCheck(ctx, node)
			},
		}
	},
}

// correctnessNoCallbackInParseTryJsonFileSuffix is the nexus file that declares `parseJsonOrThrow`.
const correctnessNoCallbackInParseTryJsonFileSuffix = "/source/structured-text/json/Json.ts"

func correctnessNoCallbackInParseTryCheck(ctx rule.Context, node *ast.Node) {
	tryStatement := node.AsTryStatement()
	if tryStatement.CatchClause == nil || !correctnessNoCallbackInParseTryCatchDiscards(ctx, tryStatement.CatchClause) {
		return
	}
	parses := false
	var callbacks []*ast.Node
	correctnessNoCallbackInParseTryGuarded(tryStatement.TryBlock, func(call *ast.Node) {
		if correctnessNoCallbackInParseTryIsParse(ctx, call) {
			parses = true
		} else if correctnessNoCallbackInParseTryIsCallback(ctx, call) {
			callbacks = append(callbacks, call)
		}
	})
	if !parses {
		return
	}
	for _, call := range callbacks {
		ctx.ReportNode(call, correctnessNoCallbackInParseTryMessage())
	}
}

// correctnessNoCallbackInParseTryCatchDiscards says whether a catch clause throws the caught value
// away: it binds nothing, or binds a plain name it never refers to. A destructured binding reads the
// value by construction and does not count.
func correctnessNoCallbackInParseTryCatchDiscards(ctx rule.Context, catchClause *ast.Node) bool {
	clause := catchClause.AsCatchClause()
	if clause.VariableDeclaration == nil {
		return true
	}
	name := clause.VariableDeclaration.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil {
		return false
	}
	referenced := false
	var visit func(child *ast.Node) bool
	visit = func(child *ast.Node) bool {
		if referenced {
			return true
		}
		if child.Kind == ast.KindIdentifier && child.Text() == name.Text() {
			resolved := ctx.TypeChecker.GetSymbolAtLocation(child)
			if parent := child.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
				resolved = ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent)
			}
			// An unresolved identifier of the same name may still be the binding; count it as read.
			if resolved == nil || resolved == symbol {
				referenced = true
				return true
			}
		}
		child.ForEachChild(visit)
		return referenced
	}
	clause.Block.ForEachChild(visit)
	return !referenced
}

// correctnessNoCallbackInParseTryGuarded hands every call the try block's catch would see thrown
// through: not inside a nested function, and not inside the block of a nested try that catches.
func correctnessNoCallbackInParseTryGuarded(block *ast.Node, visitCall func(call *ast.Node)) {
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if control_flow_graph.IsRoot(node) {
			return false
		}
		switch node.Kind {
		case ast.KindTryStatement:
			nested := node.AsTryStatement()
			if nested.CatchClause != nil {
				nested.CatchClause.ForEachChild(visit)
				if nested.FinallyBlock != nil {
					visit(nested.FinallyBlock)
				}
				return false
			}
		case ast.KindCallExpression:
			visitCall(node)
		}
		node.ForEachChild(visit)
		return false
	}
	block.ForEachChild(visit)
}

// correctnessNoCallbackInParseTryIsParse says whether a call is `JSON.parse` or nexus
// `parseJsonOrThrow`, by the declaration its callee resolves to.
func correctnessNoCallbackInParseTryIsParse(ctx rule.Context, call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		// Resolved whatever it is spelled, so an aliased import of nexus's parser counts and a local
		// function that happens to share its name does not.
		return correctnessNoCallbackInParseTryIsParseJsonOrThrow(ctx, ctx.TypeChecker.GetSymbolAtLocation(callee))
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		return name.Kind == ast.KindIdentifier && correctnessNoCallbackInParseTryIsJsonParse(ctx.TypeChecker.GetSymbolAtLocation(name))
	}
	return false
}

// correctnessNoCallbackInParseTryIsJsonParse says whether a symbol is `parse` on the global `JSON`
// interface: every declaration a method signature of an interface named `JSON`, in a declaration
// file, at global scope or inside `declare global`.
func correctnessNoCallbackInParseTryIsJsonParse(symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindMethodSignature || declaration.Name() == nil || declaration.Name().Text() != "parse" {
			return false
		}
		owner := declaration.Parent
		if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || owner.Name() == nil || owner.Name().Text() != "JSON" {
			return false
		}
		if !correctnessNoCallbackInParseTryIsGlobalDeclaration(owner) {
			return false
		}
	}
	return true
}

// correctnessNoCallbackInParseTryIsParseJsonOrThrow says whether a symbol, through an import, is the
// function `parseJsonOrThrow` declared in nexus's Json.ts.
func correctnessNoCallbackInParseTryIsParseJsonOrThrow(ctx rule.Context, symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = checker.SkipAlias(symbol, ctx.TypeChecker)
	}
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil ||
			declaration.Name().Text() != "parseJsonOrThrow" {
			return false
		}
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !strings.HasSuffix(file.FileName().AsString(), correctnessNoCallbackInParseTryJsonFileSuffix) {
			return false
		}
	}
	return true
}

// correctnessNoCallbackInParseTryIsGlobalDeclaration says whether a declaration sits in a
// declaration file at global scope, or inside `declare global`.
func correctnessNoCallbackInParseTryIsGlobalDeclaration(declaration *ast.Node) bool {
	file := ast.GetSourceFileOfNode(declaration)
	if file == nil || !file.IsDeclarationFile {
		return false
	}
	container := declaration.Parent
	switch {
	case container != nil && container.Kind == ast.KindSourceFile:
		return !ast.IsExternalModule(container.AsSourceFile())
	case container != nil && container.Kind == ast.KindModuleBlock && ast.IsGlobalScopeAugmentation(container.Parent):
		return true
	}
	return false
}

// correctnessNoCallbackInParseTryIsCallback says whether a call runs a caller-supplied function: its
// callee is a name bound by a parameter, directly or destructured, whose type is callable once the
// nullish part is set aside, and which is not a Promise executor's resolve or reject.
func correctnessNoCallbackInParseTryIsCallback(ctx rule.Context, call *ast.Node) bool {
	callee := call.AsCallExpression().Expression
	for callee.Kind == ast.KindParenthesizedExpression || callee.Kind == ast.KindNonNullExpression {
		if callee.Kind == ast.KindParenthesizedExpression {
			callee = callee.AsParenthesizedExpression().Expression
		} else {
			callee = callee.AsNonNullExpression().Expression
		}
	}
	if callee.Kind != ast.KindIdentifier {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(callee)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	parameter := correctnessNoCallbackInParseTryOwningParameter(symbol.Declarations[0])
	if parameter == nil || correctnessNoCallbackInParseTryIsPromiseExecutorParameter(ctx, parameter) {
		return false
	}
	// An `any` callee has no call signature, so it is not counted here.
	calleeType := ctx.TypeChecker.GetTypeAtLocation(callee)
	if calleeType == nil {
		return false
	}
	return len(ctx.TypeChecker.GetSignaturesOfType(ctx.TypeChecker.GetNonNullableType(calleeType), checker.SignatureKindCall)) > 0
}

// correctnessNoCallbackInParseTryOwningParameter is the parameter a declaration belongs to: the
// declaration itself when it is one, or the parameter whose binding pattern it is destructured from.
// Nil for anything else.
func correctnessNoCallbackInParseTryOwningParameter(declaration *ast.Node) *ast.Node {
	current := declaration
	for current != nil {
		switch current.Kind {
		case ast.KindParameter:
			return current
		case ast.KindBindingElement, ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			current = current.Parent
		default:
			return nil
		}
	}
	return nil
}

// correctnessNoCallbackInParseTryIsPromiseExecutorParameter says whether a parameter is the first or
// second of a function passed straight as the executor of `new Promise(...)`, the global Promise.
// Those are the resolve and reject functions, which never throw.
func correctnessNoCallbackInParseTryIsPromiseExecutorParameter(ctx rule.Context, parameter *ast.Node) bool {
	executor := parameter.Parent
	if executor == nil || (executor.Kind != ast.KindArrowFunction && executor.Kind != ast.KindFunctionExpression) {
		return false
	}
	parameters := executor.Parameters()
	if len(parameters) == 0 || (parameters[0] != parameter && (len(parameters) < 2 || parameters[1] != parameter)) {
		return false
	}
	outer := executor.Parent
	argument := executor
	for outer != nil && outer.Kind == ast.KindParenthesizedExpression {
		argument = outer
		outer = outer.Parent
	}
	if outer == nil || outer.Kind != ast.KindNewExpression {
		return false
	}
	newExpression := outer.AsNewExpression()
	if newExpression.Arguments == nil || len(newExpression.Arguments.Nodes) == 0 || newExpression.Arguments.Nodes[0] != argument {
		return false
	}
	constructor := ast.SkipParentheses(newExpression.Expression)
	if constructor.Kind != ast.KindIdentifier || constructor.Text() != "Promise" {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(constructor)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		owner := declaration
		if declaration.Kind == ast.KindVariableDeclaration {
			// `declare var Promise: PromiseConstructor` sits in a variable statement's list.
			owner = declaration.Parent.Parent
		}
		if !correctnessNoCallbackInParseTryIsGlobalDeclaration(owner) {
			return false
		}
	}
	return true
}
