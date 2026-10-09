package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoUnclearedRaceTimeoutId = "unclearedRaceTimeout"

// correctnessNoUnclearedRaceTimeoutText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-uncleared-race-timeout.json`.
var correctnessNoUnclearedRaceTimeoutText = policy.MessageOf("nexus/correctness-no-uncleared-race-timeout", correctnessNoUnclearedRaceTimeoutId)

func correctnessNoUnclearedRaceTimeoutMessage() rule.Message {
	return rule.Message{Id: correctnessNoUnclearedRaceTimeoutId, Description: correctnessNoUnclearedRaceTimeoutText.Render(nil)}
}

// CorrectnessNoUnclearedRaceTimeout reports a timeout armed inside a `Promise.race` whose timer
// handle nothing can reach, so the timer is never cleared when the other side wins.
//
//	invalid: await Promise.race([work(), new Promise((_resolve, reject) => { setTimeout(() => reject(new Error('timeout')), ms); })]);
//	invalid: await Promise.race([work(), new Promise((_resolve, reject) => setTimeout(() => reject(new Error('timeout')), ms))]);
//	invalid: const timeout = new Promise<never>((_resolve, reject) => { const timer = setTimeout(reject, ms); }); await Promise.race([work(), timeout]);
//	valid:   let timer; try { await Promise.race([work(), new Promise((_r, reject) => { timer = setTimeout(reject, ms); })]); } finally { clearTimeout(timer); }
//	valid:   await Promise.race([work(), new Promise((_r, reject) => { setTimeout(reject, ms).unref(); })]);
//	valid:   await Promise.race([work(), timeoutAfter(ms)]);
//
// # Where it came from
//
// `modules/phi/social/PhiSocialGenerator.ts` in ahra, the per-image render bound: `Promise.race`
// between `generateImage(...)` and a `new Promise<never>` whose executor calls `setTimeout(...)` for
// ten minutes and keeps nothing. Every image that renders in time leaves a ten-minute timer behind,
// so an unattended batch CLI lingers up to ten minutes after its last image. Found by the
// cross-language pass of the new-rules sweep (`#tevhg3f`, after go vet's `lostcancel`, which reports
// a `context.WithTimeout` whose cancel function is dropped), built as part of `#j03vwm6`.
//
// # The shape, exactly
//
//  1. A call of `race` whose symbol is the default library's `PromiseConstructor.race`, so a local
//     `race` helper or a library's `Promise`-like class is not read. Its first argument is an array
//     literal.
//  2. An element of that array is `new Promise(executor)`, either written in place or as a bare
//     reference to a `const` in the same file whose initializer is one. `Promise` is the default
//     library's. The executor is an arrow function or a function expression.
//  3. In the executor's own body, not inside a function nested in it, a call of the global
//     `setTimeout` (also through `globalThis.` or `window.`). Global means every declaration of the
//     symbol is ambient and at global scope, which is lib.dom's and `@types/node`'s
//     `declare global`; `setTimeout` imported from `node:timers/promises` is another function.
//  4. Its handle is lost: the call is a statement on its own, the expression body of the executor (the
//     Promise constructor ignores what an executor returns), the operand of `void`, or stored in a
//     plain binding (a `const`/`let` initializer, or `name = setTimeout(...)`) that nothing in the file
//     ever reads.
//
// Every other use of the handle is a way it could be cleared, and stays silent without the rule
// following it: passed to a function, returned, stored on an object (`this.timer = ...`), chained
// (`.unref()`, which also stops the hold on the process), or a binding read anywhere at all. A read
// that does not clear (`log(timer)`) is a missed finding, never a false one.
//
// # What it declines
//
// A timeout built by a helper (`Promise.race([work(), timeoutAfter(ms)])`) is not followed into the
// helper, and a timeout promise reaching the array any other way (a spread, a `let`, a property) is
// not read. Both are missed findings. `setInterval` is not this rule's shape. `AbortSignal.timeout()`
// owns its own timer and is the cleaner fix where the work accepts a signal.
//
// # No fix
//
// The repair needs a binding declared outside the race and a `finally` around the statement that
// awaits it, whose extent depends on the surrounding `try`. That is the author's edit.
var CorrectnessNoUnclearedRaceTimeout = rule.Rule{
	Name: "nexus/correctness-no-uncleared-race-timeout",

	// Symbol identity is what tells the default library's `Promise.race` from a local `race`, the global
	// `setTimeout` from an imported one, and one binding from a shadowing one.
	NeedsTypeChecker: true,

	// Whether `race` and `Promise` are the default library's, through type_checking.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				correctnessNoUnclearedRaceTimeoutCheckRace(ctx, node)
			},
		}
	},
}

// correctnessNoUnclearedRaceTimeoutCheckRace reports the lost timers in every timeout promise of one
// `Promise.race([...])` call.
func correctnessNoUnclearedRaceTimeoutCheckRace(ctx rule.Context, node *ast.Node) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	name := callee.AsPropertyAccessExpression().Name()
	if name.Kind != ast.KindIdentifier || name.Text() != "race" {
		return
	}
	if !correctnessNoUnclearedRaceTimeoutIsLibraryMember(ctx, ctx.TypeChecker.GetSymbolAtLocation(name), "PromiseConstructor") {
		return
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return
	}
	array := ast.SkipParentheses(call.Arguments.Nodes[0])
	if array.Kind != ast.KindArrayLiteralExpression {
		return
	}
	for _, element := range array.AsArrayLiteralExpression().Elements.Nodes {
		promise := correctnessNoUnclearedRaceTimeoutPromise(ctx, ast.SkipParentheses(element))
		if promise == nil {
			continue
		}
		for _, timer := range correctnessNoUnclearedRaceTimeoutLostTimers(ctx, promise) {
			ctx.ReportNode(timer, correctnessNoUnclearedRaceTimeoutMessage())
		}
	}
}

// correctnessNoUnclearedRaceTimeoutIsLibraryMember says whether a symbol is a member of the named
// interface in the default library, with every declaration there.
func correctnessNoUnclearedRaceTimeoutIsLibraryMember(ctx rule.Context, symbol *ast.Symbol, interfaceName string) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		owner := declaration.Parent
		if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || owner.Name() == nil || owner.Name().Text() != interfaceName {
			return false
		}
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !type_checking.IsSourceFileDefaultLibrary(ctx.Program, file) {
			return false
		}
	}
	return true
}

// correctnessNoUnclearedRaceTimeoutPromise is the `new Promise(executor)` a race element is, written
// in place or named by a `const` of this file initialized to one, and nil for anything else.
func correctnessNoUnclearedRaceTimeoutPromise(ctx rule.Context, element *ast.Node) *ast.Node {
	if element.Kind == ast.KindIdentifier {
		symbol := ctx.TypeChecker.GetSymbolAtLocation(element)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return nil
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
			return nil
		}
		list := declaration.Parent
		if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsConst == 0 {
			return nil
		}
		initializer := declaration.AsVariableDeclaration().Initializer
		if initializer == nil {
			return nil
		}
		element = ast.SkipParentheses(initializer)
	}
	if element.Kind != ast.KindNewExpression {
		return nil
	}
	constructor := ast.SkipParentheses(element.AsNewExpression().Expression)
	if constructor.Kind != ast.KindIdentifier || constructor.Text() != "Promise" ||
		!type_checking.IsSymbolFromDefaultLibrary(ctx.Program, ctx.TypeChecker.GetSymbolAtLocation(constructor)) {
		return nil
	}
	return element
}

// correctnessNoUnclearedRaceTimeoutLostTimers returns the global `setTimeout` calls made directly in a
// promise's executor whose handle nothing can reach.
func correctnessNoUnclearedRaceTimeoutLostTimers(ctx rule.Context, promise *ast.Node) []*ast.Node {
	arguments := promise.AsNewExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return nil
	}
	executor := ast.SkipParentheses(arguments.Nodes[0])
	if executor.Kind != ast.KindArrowFunction && executor.Kind != ast.KindFunctionExpression {
		return nil
	}
	body := executor.Body()
	if body == nil {
		return nil
	}

	var lost []*ast.Node
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) || node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression {
			// A nested function may run later, more than once, or never; its timers are not the
			// executor's.
			return false
		}
		if node.Kind == ast.KindCallExpression && correctnessNoUnclearedRaceTimeoutIsGlobalSetTimeout(ctx, node) &&
			correctnessNoUnclearedRaceTimeoutHandleIsLost(ctx, executor, node) {
			lost = append(lost, node)
		}
		node.ForEachChild(visit)
		return false
	}
	visit(body)
	return lost
}

// correctnessNoUnclearedRaceTimeoutIsGlobalSetTimeout says whether a call is to the global
// `setTimeout`, directly or through `globalThis.` or `window.`.
func correctnessNoUnclearedRaceTimeoutIsGlobalSetTimeout(ctx rule.Context, call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	var name *ast.Node
	switch callee.Kind {
	case ast.KindIdentifier:
		name = callee
	case ast.KindPropertyAccessExpression:
		receiver := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
		if receiver.Kind != ast.KindIdentifier || (receiver.Text() != "globalThis" && receiver.Text() != "window") {
			return false
		}
		name = callee.AsPropertyAccessExpression().Name()
	default:
		return false
	}
	if name.Kind != ast.KindIdentifier || name.Text() != "setTimeout" {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if !correctnessNoUnclearedRaceTimeoutIsGlobalTimer(ctx, declaration) {
			return false
		}
	}
	return true
}

// correctnessNoUnclearedRaceTimeoutIsGlobalTimer says whether one declaration of `setTimeout` is the
// platform's timer: ambient, in a declaration file, at global scope or inside `declare global`, or
// the default library's `WindowOrWorkerGlobalScope` member. lib.dom declares the global function and
// that member (`window.setTimeout` resolves to both at once); `@types/node` declares the function,
// and a namespace merged with it, inside `declare global`.
func correctnessNoUnclearedRaceTimeoutIsGlobalTimer(ctx rule.Context, declaration *ast.Node) bool {
	file := ast.GetSourceFileOfNode(declaration)
	if file == nil || !file.IsDeclarationFile {
		return false
	}
	container := declaration.Parent
	switch {
	case container == nil:
		return false
	case container.Kind == ast.KindSourceFile:
		return !ast.IsExternalModule(container.AsSourceFile())
	case container.Kind == ast.KindModuleBlock:
		return ast.IsGlobalScopeAugmentation(container.Parent)
	case container.Kind == ast.KindInterfaceDeclaration:
		return container.Name() != nil && container.Name().Text() == "WindowOrWorkerGlobalScope" &&
			type_checking.IsSourceFileDefaultLibrary(ctx.Program, file)
	}
	return false
}

// correctnessNoUnclearedRaceTimeoutHandleIsLost says whether the value a `setTimeout` call returns
// can never be reached: dropped where it is made, or kept in a binding nothing reads.
func correctnessNoUnclearedRaceTimeoutHandleIsLost(ctx rule.Context, executor *ast.Node, call *ast.Node) bool {
	outer := call
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	switch {
	case parent == nil:
		return false
	case parent.Kind == ast.KindExpressionStatement, parent.Kind == ast.KindVoidExpression:
		return true
	case parent == executor:
		// The executor's expression body. The Promise constructor discards what the executor returns.
		return true
	case parent.Kind == ast.KindVariableDeclaration && parent.AsVariableDeclaration().Initializer == outer:
		name := parent.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return correctnessNoUnclearedRaceTimeoutNeverRead(ctx, ctx.TypeChecker.GetSymbolAtLocation(name))
	case parent.Kind == ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindEqualsToken || binary.Right != outer {
			return false
		}
		target := ast.SkipParentheses(binary.Left)
		if target.Kind != ast.KindIdentifier {
			return false
		}
		// The assignment's own value is a read of the handle (`a = b = setTimeout(...)`).
		assignment := parent
		for assignment.Parent != nil && assignment.Parent.Kind == ast.KindParenthesizedExpression {
			assignment = assignment.Parent
		}
		if assignment.Parent == nil || assignment.Parent.Kind != ast.KindExpressionStatement {
			return false
		}
		return correctnessNoUnclearedRaceTimeoutNeverRead(ctx, ctx.TypeChecker.GetSymbolAtLocation(target))
	}
	return false
}

// correctnessNoUnclearedRaceTimeoutNeverRead says whether a binding of this file is never read: its
// every reference is its own declaration name or the target of a plain `=` assignment. A binding
// declared outside this file, or with more than one declaration, is assumed read.
func correctnessNoUnclearedRaceTimeoutNeverRead(ctx rule.Context, symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
		return false
	}
	declared := declaration.Name()
	name := declared.Text()
	read := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if read {
			return true
		}
		if node.Kind == ast.KindIdentifier && node.Text() == name && node != declared {
			resolved := ctx.TypeChecker.GetSymbolAtLocation(node)
			if parent := node.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
				if valueSymbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent); valueSymbol != nil {
					resolved = valueSymbol
				}
			}
			if resolved == symbol && !correctnessNoUnclearedRaceTimeoutIsPlainAssignmentTarget(node) {
				read = true
				return true
			}
		}
		node.ForEachChild(visit)
		return false
	}
	ctx.SourceFile.AsNode().ForEachChild(visit)
	return !read
}

// correctnessNoUnclearedRaceTimeoutIsPlainAssignmentTarget says whether an identifier is the whole
// left side of a plain `=` assignment, which writes the binding without reading it.
func correctnessNoUnclearedRaceTimeoutIsPlainAssignmentTarget(identifier *ast.Node) bool {
	outer := identifier
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	return binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Left == outer
}
