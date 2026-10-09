package nexus

import (
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoProcessExitAfterOutputId = "exitAfterOutput"

// correctnessNoProcessExitAfterOutputText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-process-exit-after-output.json`.
var correctnessNoProcessExitAfterOutputText = policy.MessageOf("nexus/correctness-no-process-exit-after-output", correctnessNoProcessExitAfterOutputId)

func correctnessNoProcessExitAfterOutputMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoProcessExitAfterOutputId,
		Description: correctnessNoProcessExitAfterOutputText.Render(nil),
	}
}

// CorrectnessNoProcessExitAfterOutput reports a `process.exit()` that runs after the same function
// has written to stdout or stderr, on some path through its control-flow graph.
//
//	invalid: for(const line of result.lines) console.log(line); process.exit(result.exitCode);
//	invalid: if(!report.authorized) { console.log('not linked'); process.exit(0); }
//	invalid: try { run(); } catch(error) { console.error(error); process.exit(1); }
//	invalid: console.log('starting'); try { await run(); } catch { process.exit(1); }
//	valid:   for(const line of result.lines) console.log(line); process.exitCode = result.exitCode; return;
//	valid:   if(!flag) process.exit(1); console.log('ready');
//	valid:   process.stdout.write(output, function() { process.exit(0); });
//	invalid: function showHelp() { console.log(usage); } ... showHelp(); process.exit(0);
//	valid:   printReport(report); process.exit(0);   (declared, so no body to read)
//
// # Where it came from
//
// The cross-language pass of the new-rules sweep (`#tevhg3f`, after Node's own warning on
// `process.exit` and Go's `log.Fatal` skipping deferred flushes), built as task `#gjs78wz`. Measured
// on macOS with Node 24: `console.log` of 4 MB followed by `process.exit(0)` delivered exactly 65,536
// bytes through a pipe, exit status 0, no error; to a file it was complete. ahra is a command-line
// kingdom read through pipes, and its densest file, `modules/finance/FinanceConnectionsCommandLineInterface.ts`,
// ends some twenty report commands with a block of `console.log` and `process.exit(0)`: `runSync`,
// `runStatement` (a loop printing every line of a statement), `runQuickBooksCardBalances` (an early
// report and exit, then the full report and another). `modules/godword/GodwordEntropy.ts` reaches
// the same exit through `import * as NodeProcess from 'node:process'`.
//
// # What counts as a write, and as the exit
//
// Both by symbol, never by spelling:
//
//   - **A write** is a call of `log`, `info`, `debug`, `warn`, `error`, `trace`, `table`, `dir` or
//     `dirxml` on the global `console` (an identifier whose every declaration is an ambient global
//     variable: lib.dom's, and `@types/node`'s inside `declare global`), or a call of `write` on
//     `stdout` or `stderr` where that member is the one declared on `NodeJS.Process` in a declaration
//     file. So `process.stdout.write`, `NodeProcess.stdout.write` through `import * as NodeProcess
//     from 'node:process'`, and `stdout.write` through `import { stdout } from 'node:process'` all
//     count. A local `console`, a `new Console(fileStream)` and any other stream's `write` do not.
//   - **The exit** is a call of the `exit` member declared on `NodeJS.Process` in a declaration file,
//     reached the same three ways (`process.exit`, `NodeProcess.exit`, an imported `exit`).
//
// # How "after" is decided
//
// Through the enclosing function's control-flow graph, walked forward from its entry. Each path
// carries whether a write has happened on it; an exit met on a path that carries one is reported.
// An exit ends its path, since nothing after it runs, so a write laid out after an exit never counts
// toward a later one. So:
//
//   - **A write on any one path is enough.** `if(!report.authorized) { console.log(...);
//     process.exit(0); }` reports, and so does an exit at the top of a loop whose previous turn
//     printed.
//   - **An exit before the writes is clean.** `if(!flag) process.exit(1); console.log('ready');`
//     stays silent: no path reaches the exit having written.
//   - **The shape is the condition.** Whether the bytes are still buffered when the exit runs depends
//     on how much was written, where the stream leads, and what ran in between (an `await` may let a
//     pipe drain), none of which the source says. A write followed by an exit on one path is the hazard
//     Node's documentation names, and it is the whole condition.
//
// # What "the same function" means, and the one level of callee it reads
//
// The function the exit is in, plus one level of the functions it calls. A call counts as a write
// when the checker resolves it to a function with a body whose own body, nested functions aside,
// makes a write itself: `showHelp(); process.exit(0);` in ahra's AhraCommandLineInterface.ts and the
// `reportFreshness(...)` exits in BackupCommandLineInterface.ts were the two real hazards the rule
// missed before it read callees (#4pyyvfs). One level only, so a helper that prints through another
// helper is still a missed finding. The callee is followed only where it can be read and only where
// its body surely runs before the call returns, each refusal costing a finding and never adding one:
//
//   - **In this file or in a module.** A module's function is reachable only through an import, so its
//     file is in the import closure the findings cache keys on, a dynamic `import()` included. A
//     global script's function has no such edge, and a declaration file has no body.
//   - **Not overloaded, not a generator, and awaited when async.** An overload's resolved declaration
//     is a signature with no body. A generator's body runs only when it is iterated. An async body
//     runs up to its first `await` and finishes later, so it counts only when the call is awaited
//     directly.
//   - **Not one that exits or never returns.** A callee declared `never`, or whose own body exits,
//     leaves the caller's exit as possibly dead code, so it is not read as a write.
//
// A write inside a callback (`rows.forEach((row) => console.log(row)); process.exit(0);`) is still
// not followed: the callback is not called by name, and nothing says when it runs. An exit inside a
// callback is judged in the callback alone, which is what keeps the correct form silent:
// `process.stdout.write(output, function() { process.exit(0); })` exits only once the write has been
// handed to the operating system.
//
// # Where the graph says more than happens, and what the rule does about it
//
// Three places where a path in the graph is not one that runs, each handled so it can only cost a
// finding:
//
//   - **The end of a `try` block flows into its `catch`** (ESLint's shape, kept by the graph), and the
//     first node that could throw forks there too. A write inside the `try` block can therefore reach
//     the `catch` on a path where nothing after the write threw, or where the write itself was what
//     threw. So a path entering a `catch` forgets every write made inside that statement's `try` block.
//     A write made before the `try` statement is kept: any path into the `catch` ran it.
//     `console.log('starting'); try { await run(); } catch { process.exit(1); }` reports;
//     `try { await run(); console.log('done'); } catch { process.exit(1); }` does not, though the
//     second is a real hazard whenever something after the write throws. That one is a missed finding.
//   - **A write is recorded where its call begins**, before its arguments run, because the graph has
//     no hook after a call completes. An exit inside a write's own arguments
//     (`console.log(process.exit(1))`) would then look like it ran second, so a write whose arguments
//     hold an exit is not recorded at all. An exit inside the catch clause's binding (`catch({ code =
//     process.exit(1) })`) runs before the marker the previous point relies on, so it is not
//     recorded either. As a guard against an exit the walk would pass through unseen, letting a write
//     after it count toward a later exit, a function with an unrecorded exit is skipped whole, so the
//     catch-binding exit costs its function's findings. The guard never fired on ahra, counted by a
//     probe, and a mutant removing it survives: the only exit it skips sits in a default that is
//     itself skipped whenever the error has the field, so every path past it also runs without it.
//   - **A `finally` block is laid out twice**, once for normal completion and once for the path that
//     leaves through `return` or `throw`, and both are walked. A finding is per exit call, and the
//     normal copy reaches it with the same writes whenever the `try` completes, so the second copy
//     adds nothing a reader would call false.
//
// # No fix
//
// The fix is `process.exitCode = n` and a `return`, but where the `return` goes is the author's call:
// an exit inside a loop or a nested block needs the function to leave from there, and an exit used as
// `never` (`if(!value) { console.error(...); process.exit(1); } use(value);`) narrows a type that a
// `return` has to keep narrowed. A fixer would get some of those wrong.
var CorrectnessNoProcessExitAfterOutput = rule.Rule{
	Name: "nexus/correctness-no-process-exit-after-output",

	// Symbol identity is what tells the global `console` from a local one, `NodeJS.Process.exit` from
	// any other `exit`, and the process's own `stdout` from any other stream with a `write`.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				correctnessNoProcessExitAfterOutputScanFile(ctx, sourceFile)
			},
		}
	},
}

// correctnessNoProcessExitAfterOutputEventKind is what one recorded event is.
type correctnessNoProcessExitAfterOutputEventKind uint8

const (
	correctnessNoProcessExitAfterOutputWrite correctnessNoProcessExitAfterOutputEventKind = iota
	correctnessNoProcessExitAfterOutputExit
	correctnessNoProcessExitAfterOutputCatchEntered
)

// correctnessNoProcessExitAfterOutputEvent is one thing the graph recorded: a write or an exit with
// its call, or the start of a `catch` block with its `try` statement.
type correctnessNoProcessExitAfterOutputEvent struct {
	kind correctnessNoProcessExitAfterOutputEventKind
	node *ast.Node
}

// correctnessNoProcessExitAfterOutputConsoleMethods are the members of the global console that
// write to stdout or stderr every time they are called.
var correctnessNoProcessExitAfterOutputConsoleMethods = map[string]bool{
	"log": true, "info": true, "debug": true, "warn": true, "error": true, "trace": true,
	"table": true, "dir": true, "dirxml": true,
}

// correctnessNoProcessExitAfterOutputRoot is one code path root that holds an exit, with what the
// file scan found in it.
type correctnessNoProcessExitAfterOutputRoot struct {
	node     *ast.Node
	exits    []*ast.Node
	hasWrite bool
}

func correctnessNoProcessExitAfterOutputScanFile(ctx rule.Context, sourceFile *ast.Node) {
	writers := correctnessNoProcessExitAfterOutputWriters{ctx: ctx, byDeclaration: map[*ast.Node]bool{}}
	rootsByNode := map[*ast.Node]*correctnessNoProcessExitAfterOutputRoot{}
	var roots []*correctnessNoProcessExitAfterOutputRoot
	rootFor := func(call *ast.Node) *correctnessNoProcessExitAfterOutputRoot {
		node := control_flow_graph.RootOf(call)
		if node == nil {
			return nil
		}
		root := rootsByNode[node]
		if root == nil {
			root = &correctnessNoProcessExitAfterOutputRoot{node: node}
			rootsByNode[node] = root
			roots = append(roots, root)
		}
		return root
	}

	// Two passes, exits first. Most files never call process.exit, and a call's callee is resolved
	// only inside a root that holds an exit: asking the checker for every call's signature in every
	// file took the rule's run on ahra from 2.5s to 4.2s.
	var findExits func(node *ast.Node) bool
	findExits = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression && correctnessNoProcessExitAfterOutputIsExit(ctx, node) {
			if root := rootFor(node); root != nil {
				root.exits = append(root.exits, node)
			}
		}
		node.ForEachChild(findExits)
		return false
	}
	sourceFile.ForEachChild(findExits)
	if len(roots) == 0 {
		return
	}

	var findWrites func(node *ast.Node) bool
	findWrites = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			if root := rootsByNode[control_flow_graph.RootOf(node)]; root != nil && !root.hasWrite && writers.writes(node) {
				root.hasWrite = true
			}
		}
		node.ForEachChild(findWrites)
		return false
	}
	sourceFile.ForEachChild(findWrites)

	for _, root := range roots {
		if len(root.exits) > 0 && root.hasWrite {
			correctnessNoProcessExitAfterOutputAnalyzeRoot(ctx, root, writers)
		}
	}
}

// correctnessNoProcessExitAfterOutputIsExit says whether a call is `NodeJS.Process.exit`, reached
// through any receiver or an imported binding.
func correctnessNoProcessExitAfterOutputIsExit(ctx rule.Context, call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	name := correctnessNoProcessExitAfterOutputMemberName(callee)
	return name != nil && name.Text() == "exit" && correctnessNoProcessExitAfterOutputIsProcessMember(ctx, name)
}

// correctnessNoProcessExitAfterOutputIsWrite says whether a call writes to stdout or stderr: a
// writing method of the global console, or `write` on the process's own `stdout` or `stderr`.
func correctnessNoProcessExitAfterOutputIsWrite(ctx rule.Context, call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression && callee.Kind != ast.KindElementAccessExpression {
		return false
	}
	member, named := property.AccessedName(callee, property.Textual)
	if !named {
		return false
	}
	var receiver *ast.Node
	if callee.Kind == ast.KindPropertyAccessExpression {
		receiver = ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	} else {
		receiver = ast.SkipParentheses(callee.AsElementAccessExpression().Expression)
	}
	if correctnessNoProcessExitAfterOutputConsoleMethods[member] && receiver.Kind == ast.KindIdentifier &&
		receiver.Text() == "console" {
		return correctnessNoProcessExitAfterOutputIsGlobalVariable(ctx.TypeChecker.GetSymbolAtLocation(receiver))
	}
	if member != "write" {
		return false
	}
	stream := correctnessNoProcessExitAfterOutputMemberName(receiver)
	return stream != nil && (stream.Text() == "stdout" || stream.Text() == "stderr") &&
		correctnessNoProcessExitAfterOutputIsProcessMember(ctx, stream)
}

// correctnessNoProcessExitAfterOutputWriters answers whether a call writes, directly or through one
// callee, remembering each callee's answer for the file.
type correctnessNoProcessExitAfterOutputWriters struct {
	ctx           rule.Context
	byDeclaration map[*ast.Node]bool
}

// writes says whether a call writes to stdout or stderr itself, or runs a callee whose own body does.
func (writers correctnessNoProcessExitAfterOutputWriters) writes(call *ast.Node) bool {
	if correctnessNoProcessExitAfterOutputIsWrite(writers.ctx, call) {
		return true
	}
	callee := writers.followedCallee(call)
	if callee == nil {
		return false
	}
	answer, known := writers.byDeclaration[callee]
	if !known {
		answer = correctnessNoProcessExitAfterOutputBodyWrites(writers.ctx, callee)
		writers.byDeclaration[callee] = answer
	}
	return answer
}

// followedCallee is the function a call runs, when the rule reads its body: see the rule's doc
// comment for each refusal. Nil means the call is not followed.
func (writers correctnessNoProcessExitAfterOutputWriters) followedCallee(call *ast.Node) *ast.Node {
	signature := writers.ctx.TypeChecker.GetResolvedSignature(call)
	if signature == nil {
		return nil
	}
	declaration := signature.Declaration()
	if declaration == nil {
		return nil
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
	default:
		return nil
	}
	if declaration.Body() == nil {
		return nil
	}
	file := ast.GetSourceFileOfNode(declaration)
	// A declaration file needs no test of its own: it has no bodies, so the check above refused it
	if file == nil || (file != writers.ctx.SourceFile && !ast.IsExternalModule(file)) {
		return nil
	}
	flags := ast.GetFunctionFlags(declaration)
	if flags&ast.FunctionFlagsGenerator != 0 {
		return nil
	}
	if flags&ast.FunctionFlagsAsync != 0 && !correctnessNoProcessExitAfterOutputIsAwaited(call) {
		return nil
	}
	returnType := writers.ctx.TypeChecker.GetReturnTypeOfSignature(signature)
	if returnType != nil && returnType.Flags()&checker.TypeFlagsNever != 0 {
		return nil
	}
	return declaration
}

// correctnessNoProcessExitAfterOutputIsAwaited says whether a call is the operand of an `await`,
// through parentheses.
func correctnessNoProcessExitAfterOutputIsAwaited(call *ast.Node) bool {
	parent := call.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindAwaitExpression
}

// correctnessNoProcessExitAfterOutputBodyWrites says whether a function's own body, nested functions
// aside, makes a write and never exits. A body that exits is not read as a write: the caller's exit
// after it may never run.
func correctnessNoProcessExitAfterOutputBodyWrites(ctx rule.Context, function *ast.Node) bool {
	wrote, exits := false, false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if exits || control_flow_graph.IsRoot(node) {
			return exits
		}
		if node.Kind == ast.KindCallExpression {
			if correctnessNoProcessExitAfterOutputIsExit(ctx, node) {
				exits = true
				return true
			}
			if correctnessNoProcessExitAfterOutputIsWrite(ctx, node) {
				wrote = true
			}
		}
		node.ForEachChild(visit)
		return exits
	}
	// An arrow's expression body is itself the thing to judge, not something to look inside
	body := function.Body()
	if body.Kind == ast.KindBlock {
		body.ForEachChild(visit)
	} else {
		visit(body)
	}
	return wrote && !exits
}

// correctnessNoProcessExitAfterOutputMemberName is the identifier that names the member an
// expression reads: the name of `receiver.member`, or the identifier itself for a bare binding that
// may be an import of the member. Anything else answers nil.
func correctnessNoProcessExitAfterOutputMemberName(expression *ast.Node) *ast.Node {
	switch expression.Kind {
	case ast.KindPropertyAccessExpression:
		name := expression.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name
		}
	case ast.KindIdentifier:
		return expression
	}
	return nil
}

// correctnessNoProcessExitAfterOutputIsProcessMember says whether a name resolves, through an import
// alias when it is one, to a member declared on `NodeJS.Process` in a declaration file and nowhere
// else: the interface `@types/node` gives the global `process`.
func correctnessNoProcessExitAfterOutputIsProcessMember(ctx rule.Context, name *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = ctx.TypeChecker.GetAliasedSymbol(symbol)
	}
	if symbol == nil || len(symbol.Declarations) == 0 || !rule.IsDeclaredOnlyInDeclarationFiles(symbol) {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindMethodSignature && declaration.Kind != ast.KindPropertySignature {
			return false
		}
		owner := declaration.Parent
		if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || owner.Name() == nil || owner.Name().Text() != "Process" {
			return false
		}
		block := owner.Parent
		if block == nil || block.Kind != ast.KindModuleBlock || block.Parent == nil || block.Parent.Kind != ast.KindModuleDeclaration {
			return false
		}
		namespace := block.Parent.Name()
		if namespace == nil || namespace.Kind != ast.KindIdentifier || namespace.Text() != "NodeJS" {
			return false
		}
	}
	return true
}

// correctnessNoProcessExitAfterOutputIsGlobalVariable says whether a symbol is an ambient global
// variable: every declaration a `var` in a declaration file, at the top of a script or inside
// `declare global`. That is lib.dom's `console` and `@types/node`'s, and never a binding a source
// file declares or imports.
func correctnessNoProcessExitAfterOutputIsGlobalVariable(symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 || !rule.IsDeclaredOnlyInDeclarationFiles(symbol) {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindVariableDeclaration {
			return false
		}
		list := declaration.Parent
		if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Parent == nil ||
			list.Parent.Kind != ast.KindVariableStatement {
			return false
		}
		container := list.Parent.Parent
		switch {
		case container != nil && container.Kind == ast.KindSourceFile:
			if ast.IsExternalModule(container.AsSourceFile()) {
				return false
			}
		case container != nil && container.Kind == ast.KindModuleBlock && ast.IsGlobalScopeAugmentation(container.Parent):
		default:
			return false
		}
	}
	return true
}

// correctnessNoProcessExitAfterOutputHoldsExit says whether an exit runs inside node, in the same
// code path root: an exit inside a nested function runs when that function is called, not here.
func correctnessNoProcessExitAfterOutputHoldsExit(ctx rule.Context, node *ast.Node) bool {
	held := false
	var visit func(child *ast.Node) bool
	visit = func(child *ast.Node) bool {
		if held || control_flow_graph.IsRoot(child) {
			return held
		}
		if child.Kind == ast.KindCallExpression && correctnessNoProcessExitAfterOutputIsExit(ctx, child) {
			held = true
			return true
		}
		child.ForEachChild(visit)
		return held
	}
	node.ForEachChild(visit)
	return held
}

// correctnessNoProcessExitAfterOutputInCatchBinding says whether node sits in the binding of a catch
// clause, which runs before the catch block's own first statement.
func correctnessNoProcessExitAfterOutputInCatchBinding(node *ast.Node, root *ast.Node) bool {
	previous := node
	for current := node.Parent; current != nil && current != root; current = current.Parent {
		if current.Kind == ast.KindCatchClause && current.AsCatchClause().VariableDeclaration == previous {
			return true
		}
		previous = current
	}
	return false
}

// correctnessNoProcessExitAfterOutputAnalyzeRoot builds one root's graph, walks it from the entry,
// and reports each exit some path reaches having written.
func correctnessNoProcessExitAfterOutputAnalyzeRoot(
	ctx rule.Context,
	root *correctnessNoProcessExitAfterOutputRoot,
	writers correctnessNoProcessExitAfterOutputWriters,
) {
	type builder = control_flow_graph.Builder[correctnessNoProcessExitAfterOutputEvent]
	recorded := map[*ast.Node]bool{}
	graph := control_flow_graph.Build(root.node, control_flow_graph.Hooks[correctnessNoProcessExitAfterOutputEvent]{
		Expression: func(b *builder, node *ast.Node) {
			if node.Kind != ast.KindCallExpression || correctnessNoProcessExitAfterOutputInCatchBinding(node, root.node) {
				return
			}
			if correctnessNoProcessExitAfterOutputIsExit(ctx, node) {
				recorded[node] = true
				b.Emit(correctnessNoProcessExitAfterOutputEvent{kind: correctnessNoProcessExitAfterOutputExit, node: node})
				return
			}
			if writers.writes(node) && !correctnessNoProcessExitAfterOutputHoldsExit(ctx, node) {
				b.Emit(correctnessNoProcessExitAfterOutputEvent{kind: correctnessNoProcessExitAfterOutputWrite, node: node})
			}
		},
		Statement: func(b *builder, node *ast.Node) {
			if node.Kind == ast.KindBlock && node.Parent != nil && node.Parent.Kind == ast.KindCatchClause {
				b.Emit(correctnessNoProcessExitAfterOutputEvent{kind: correctnessNoProcessExitAfterOutputCatchEntered, node: node.Parent.Parent})
			}
		},
	})
	for _, exit := range root.exits {
		if !recorded[exit] {
			return
		}
	}

	walk := correctnessNoProcessExitAfterOutputWalk{root: root.node, chains: map[*ast.Node]string{}, tries: map[*ast.Node]int{}}
	reported := map[*ast.Node]bool{}
	type position struct {
		block *control_flow_graph.Block[correctnessNoProcessExitAfterOutputEvent]
		state correctnessNoProcessExitAfterOutputState
	}
	visited := make([]map[string]bool, len(graph.Blocks))
	queue := []position{{block: graph.Blocks[0]}}
	visited[0] = map[string]bool{"": true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		state := current.state
		exited := false
		for _, event := range current.block.Events {
			switch event.kind {
			case correctnessNoProcessExitAfterOutputWrite:
				state = state.with(walk.chainOf(event.node))
			case correctnessNoProcessExitAfterOutputCatchEntered:
				state = state.entering(walk.tryId(event.node))
			case correctnessNoProcessExitAfterOutputExit:
				if len(state) > 0 {
					reported[event.node] = true
				}
				exited = true
			}
			if exited {
				break
			}
		}
		if exited {
			continue
		}
		key := state.key()
		for _, successor := range current.block.Successors {
			if successor == nil || !successor.Reachable {
				continue
			}
			index := successor.Index()
			if visited[index] == nil {
				visited[index] = map[string]bool{}
			}
			if visited[index][key] {
				continue
			}
			visited[index][key] = true
			queue = append(queue, position{block: successor, state: state})
		}
	}

	for _, exit := range root.exits {
		if reported[exit] {
			ctx.ReportNode(exit, correctnessNoProcessExitAfterOutputMessage())
		}
	}
}

// correctnessNoProcessExitAfterOutputWalk names the `try` statements of one root and the chain of
// them each write sits in.
type correctnessNoProcessExitAfterOutputWalk struct {
	root   *ast.Node
	chains map[*ast.Node]string
	tries  map[*ast.Node]int
}

func (walk *correctnessNoProcessExitAfterOutputWalk) tryId(statement *ast.Node) int {
	id, seen := walk.tries[statement]
	if !seen {
		id = len(walk.tries)
		walk.tries[statement] = id
	}
	return id
}

// chainOf is the set of `try` statements, with a `catch`, whose `try` block holds the write: the
// statements whose `catch` a path may enter with the write made or with the write the thing that
// threw. It is written as sorted ids, `,`-separated, with a `,` at each end so a membership test is
// a substring test.
func (walk *correctnessNoProcessExitAfterOutputWalk) chainOf(write *ast.Node) string {
	if chain, seen := walk.chains[write]; seen {
		return chain
	}
	var ids []int
	previous := write
	for current := write.Parent; current != nil && current != walk.root; current = current.Parent {
		if current.Kind == ast.KindTryStatement {
			statement := current.AsTryStatement()
			if statement.TryBlock == previous && statement.CatchClause != nil {
				ids = append(ids, walk.tryId(current))
			}
		}
		previous = current
	}
	sort.Ints(ids)
	var chain strings.Builder
	chain.WriteString(",")
	for _, id := range ids {
		chain.WriteString(strconv.Itoa(id))
		chain.WriteString(",")
	}
	walk.chains[write] = chain.String()
	return walk.chains[write]
}

// correctnessNoProcessExitAfterOutputState is what one path carries: the chains of the writes made on
// it that a `catch` could still take back, kept minimal. A chain that is a subset of another is
// forgotten by every `catch` the other is, so the larger one is dropped; the empty chain, `,`, is a
// write no `catch` takes back, and once a path holds it the path has written for good.
type correctnessNoProcessExitAfterOutputState []string

func correctnessNoProcessExitAfterOutputChainHolds(outer string, inner string) bool {
	for _, id := range strings.Split(strings.Trim(inner, ","), ",") {
		if id != "" && !strings.Contains(outer, ","+id+",") {
			return false
		}
	}
	return true
}

func (state correctnessNoProcessExitAfterOutputState) with(chain string) correctnessNoProcessExitAfterOutputState {
	next := correctnessNoProcessExitAfterOutputState{}
	for _, held := range state {
		if correctnessNoProcessExitAfterOutputChainHolds(chain, held) {
			// Something already held is forgotten only where this one is too.
			return state
		}
		if !correctnessNoProcessExitAfterOutputChainHolds(held, chain) {
			next = append(next, held)
		}
	}
	next = append(next, chain)
	sort.Strings(next)
	return next
}

func (state correctnessNoProcessExitAfterOutputState) entering(try int) correctnessNoProcessExitAfterOutputState {
	member := "," + strconv.Itoa(try) + ","
	next := correctnessNoProcessExitAfterOutputState{}
	for _, held := range state {
		if !strings.Contains(held, member) {
			next = append(next, held)
		}
	}
	return next
}

func (state correctnessNoProcessExitAfterOutputState) key() string {
	return strings.Join(state, "|")
}
