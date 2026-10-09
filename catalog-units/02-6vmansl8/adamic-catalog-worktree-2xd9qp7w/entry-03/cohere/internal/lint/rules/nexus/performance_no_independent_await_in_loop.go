package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// performanceNoIndependentAwaitInLoopText is the rule's message, whose wording lives in
// `policy/messages/performance-no-independent-await-in-loop.json`.
var performanceNoIndependentAwaitInLoopText = policy.MessageOf("nexus/performance-no-independent-await-in-loop", "independentAwaitInLoop")

func messagePerformanceNoIndependentAwaitInLoop() rule.Message {
	return rule.Message{Id: performanceNoIndependentAwaitInLoopText.Id, Description: performanceNoIndependentAwaitInLoopText.Render(nil)}
}

// PerformanceNoIndependentAwaitInLoop reports a loop over a collection whose iterations wait on
// each other for no reason the code can show.
//
//	invalid: for (const tree of trees) { results.push(await checkTree(tree)); }
//	invalid: for (const adapter of this.adapters) { try { results.push(await adapter.fetch()); } catch (error) { results.push(failed(error)); } }
//	invalid: for (let index = 0; index < files.length; index++) { await upload(files[index]); }
//	valid:   while (hasMore) { const page = await list(cursor); cursor = page.next; }
//	valid:   for (const page of pages) { cursor = await fetchPage(page, cursor); }
//	valid:   for (const section of sections) { console.log(section.name); await printTasks(section); }
//	valid:   for (const host of hosts) { if (await isUp(host)) return host; }
//	valid:   for (const item of items) { await sleep(250); await send(item); }
//
// # Why this exists rather than `no-await-in-loop`
//
// Upstream's rule reports every await inside every loop: 379 findings on ahra, and Kirk's sample put
// three quarters of them as deliberately sequential (pagination, polling, retries, ordered output).
// A rule that is wrong three times in four teaches its reader to suppress it unread. This one is
// the same question asked with the program in hand: it reports only when the loop's iterations can
// be shown not to depend on each other, so the finding is nearly always a real `Promise.all`.
//
// # When it fires, and every condition must hold
//
// The loop iterates a collection. A `for...of` that is not `for await`, or the indexed form
// `for (let i = 0; i < xs.length; i++)` whose body never writes `i`. `while`, `do`, `for...in`, a
// `for await`, and any other `for` header (an attempt counter, a cursor, `;;`) never fire, because
// those are the shapes of retry, polling and pagination.
//
// The loop itself holds an await. An await inside a nested loop belongs to that loop and is judged
// there, so a finding names the innermost loop that waits, once, however many awaits it holds. An
// await inside a nested function does not count, because the function is not called by being
// written. An await on the for-of iterable runs once and does not count either.
//
// No loop-carried state. Every identifier in the body is resolved through the checker, and a
// binding is outer when none of its declarations lies inside the loop. Then:
//
//   - a write to an outer binding anywhere in the body (assignment, compound assignment, `++`,
//     destructuring) is silence: that is a cursor, an attempt count or an accumulator;
//   - a `var` declared in the body is silence, because it is function-scoped and so outlives the
//     iteration that wrote it;
//   - in the indexed form, a write to the index in the body is silence;
//   - a compound assignment or `++` on an outer property (`summary.total += n`) is silence;
//   - a call that reads and consumes outer state (`pop`, `shift`, `splice`, `next`, `read`, `sort`,
//     `reverse` on an outer receiver) is silence.
//
// Collecting results is allowed, because it is what `Promise.all` returns: `push`, `unshift`, `set`,
// `add`, `append`, `delete` on an outer receiver, `Object.assign(outer, ...)`, a plain `=` into an
// outer property or element, and `delete outer[key]` are sinks. A sink is allowed only while the
// collection it fills is not read anywhere else in the body or by the loop's own iterable: once
// `results.length` names the archive, or `seen.has(key)` decides what runs, the iterations depend on
// each other. Paths are compared as property chains, so filling `this.cache` does not conflict with
// calling `this.fetch`, while reading `cursor` whole does conflict with writing `cursor.after`.
//
// `Object.assign(accumulator, await load())` is a sink on the same terms. Its result depends on
// order only where two loads return the same key, and `Promise.all` keeps that order too, so
// assigning the settled results in sequence reproduces it exactly. That is the
// `AhraOsCommandLineInterface` boot loader, and it reports.
//
// No exit on an awaited result. A `break`, a `return`, a `throw` not caught inside the body, or a
// labeled `continue` to an outer loop leaves the loop, and it is silence when its guard reads a
// value derived from an await: the `if` or `switch` it sits under, the test of an inner loop it sits
// in, a `catch` whose `try` awaits, a `try` that awaits before the exit and has a `catch`, an earlier
// statement that jumps on such a value, or the exit's own expression. "Derived" is a fixpoint over
// the body's own declarations and assignments, and a catch parameter whose `try` awaits counts. An
// exit whose guard reads no awaited value is not loop-carried (a validation `throw` before the first
// await), so it does not silence the rule. A plain `continue`, and a `try/catch` whose catch records
// the failure and moves on, are fine: that is the FinanceConnections shape, and it reports.
//
// No ordered side effect. Matched by name, anywhere in the body including callbacks, because these
// are about what the reader sees and a callback runs inside the iteration:
//
//   - `console.*`, `process.stdout.*`, `process.stderr.*`;
//   - `setTimeout`, `setInterval`, `setImmediate`, and any call whose name starts with `sleep`,
//     `delay`, `wait`, `pause`, `throttle`, `backoff` or `rateLimit`, which reads as pacing or
//     polling;
//   - any call whose name, or whose receiver's last segment, contains `progress` or `spinner`;
//   - `write`, `writeSync`, `appendFile`, `appendFileSync`, which write into one stream or file in
//     order;
//   - `log`, `info`, `warn`, `error`, `debug` or `trace` on a receiver whose last segment contains
//     `log` (`logger.info`, `this.log.warn`), and a bare `log(...)`;
//   - a `yield`, which hands results out in order.
//
// Inside a `catch` in the body, and one level into every function the body calls (resolved through
// the checker, imports followed), the list narrows to progress output and explicit pacing:
// `console.error`/`warn`/`debug`/`trace`, `process.stderr`, a `log`-receiver's `error`/`warn`, and
// `setTimeout` do not count there, and a callee's own `catch` blocks are not read. A failure report
// is not output whose order the reader follows, and a callee's `setTimeout` is a request timeout.
//
// These are names, not types, and the list is the whole of it: an ordered effect spelled some
// other way (a shared browser page, a migration runner executing statements in turn) is not seen
// and reports. The rule doc carries the false positives that measured on ahra.
//
// # Why syntax rather than the control-flow graph
//
// The flow graph answers "does this block reach that one", and the questions here are narrower:
// which bindings outlive an iteration (a declaration-position test on resolved symbols), and which
// conditions guard an exit (the statements between the exit and the loop body, plus earlier
// siblings that jump). Both are answered by walking up from the node, conservatively: an earlier
// sibling that holds any tainted jump counts, even when that jump targets something inside the
// sibling. Over-approximating there costs a finding, never a false one.
//
// # Where the finding points
//
// At the loop header, from `for` to the closing parenthesis, so the line a suppression goes above is
// the loop rather than one of its awaits.
//
// # No fixer
//
// The rewrite is `Promise.all`, a bounded pool, or a deliberate suppression, and which one depends
// on how large the collection is and what the far side tolerates, which the syntax cannot say.
var PerformanceNoIndependentAwaitInLoop = rule.Rule{
	Name: "nexus/performance-no-independent-await-in-loop",

	// Every identifier in the loop body is resolved to its declaration to decide whether it outlives
	// an iteration, and the binder that answers that runs only with the program.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}

		check := func(node *ast.Node) {
			analysis := newIndependentAwaitLoopAnalysis(ctx, node)
			if analysis == nil || !analysis.isIndependent() {
				return
			}
			ctx.ReportRange(independentAwaitLoopHeaderRange(ctx.SourceFile, node, analysis.body), messagePerformanceNoIndependentAwaitInLoop())
		}

		return rule.Listeners{
			ast.KindForOfStatement: check,
			ast.KindForStatement:   check,
		}
	},
}

// independentAwaitLoopAnalysis holds one loop and what the walks over its body found.
type independentAwaitLoopAnalysis struct {
	ctx  rule.Context
	loop *ast.Node
	body *ast.Node

	// collection is the expression the loop iterates: the for-of iterable, or `xs` in `xs.length`.
	collection *ast.Node

	// indexSymbol is the indexed form's counter, nil for a for-of.
	indexSymbol *ast.Symbol

	// taintedSymbols are bindings declared in the body whose value derives from an await.
	taintedSymbols map[*ast.Symbol]bool

	// orderedBodies memoizes which called function bodies make an ordered call.
	orderedBodies map[*ast.Node]bool
}

// independentAwaitLoopPath is a property chain rooted at an outer binding or at `this`.
type independentAwaitLoopPath struct {
	segments []string
	root     *ast.Node
}

func newIndependentAwaitLoopAnalysis(ctx rule.Context, loop *ast.Node) *independentAwaitLoopAnalysis {
	analysis := &independentAwaitLoopAnalysis{
		ctx:            ctx,
		loop:           loop,
		taintedSymbols: map[*ast.Symbol]bool{},
		orderedBodies:  map[*ast.Node]bool{},
	}

	switch loop.Kind {
	case ast.KindForOfStatement:
		statement := loop.AsForInOrOfStatement()
		// `for await` is asynchronous iteration: the source hands out the next item when it is ready,
		// which is the deliberate sequential case rather than the mistake.
		if statement.AwaitModifier != nil {
			return nil
		}
		analysis.body = statement.Statement
		analysis.collection = statement.Expression

	case ast.KindForStatement:
		statement := loop.AsForStatement()
		indexName, collection := independentAwaitLoopIndexedShape(statement)
		if indexName == nil {
			return nil
		}
		analysis.indexSymbol = ctx.TypeChecker.GetSymbolAtLocation(indexName)
		if analysis.indexSymbol == nil {
			return nil
		}
		analysis.body = statement.Statement
		analysis.collection = collection

	default:
		return nil
	}

	if analysis.body == nil {
		return nil
	}
	return analysis
}

// independentAwaitLoopIndexedShape recognizes `for (let i = <start>; i < xs.length; i++)` and
// answers the counter's name node and `xs`, or nil when the header is anything else.
//
// The `.length` bound is the whole point. An attempt counter (`attempt <= maximumAttempts`), a poll
// budget (`attempt < 60 && state === 'PROCESSING'`) and an infinite `for (;;)` are the same syntax
// with a different meaning, and the bound is what tells a walk over a collection from a retry.
func independentAwaitLoopIndexedShape(statement *ast.ForStatement) (*ast.Node, *ast.Node) {
	if statement.Initializer == nil || statement.Initializer.Kind != ast.KindVariableDeclarationList ||
		statement.Condition == nil || statement.Incrementor == nil {
		return nil, nil
	}
	declarations := statement.Initializer.AsVariableDeclarationList().Declarations
	if declarations == nil || len(declarations.Nodes) != 1 {
		return nil, nil
	}
	declaration := declarations.Nodes[0].AsVariableDeclaration()
	indexName := declaration.Name()
	if indexName == nil || indexName.Kind != ast.KindIdentifier || declaration.Initializer == nil {
		return nil, nil
	}

	condition := ast.SkipParentheses(statement.Condition)
	if condition.Kind != ast.KindBinaryExpression {
		return nil, nil
	}
	comparison := condition.AsBinaryExpression()
	if comparison.OperatorToken.Kind != ast.KindLessThanToken {
		return nil, nil
	}
	left := ast.SkipParentheses(comparison.Left)
	right := ast.SkipParentheses(comparison.Right)
	if left.Kind != ast.KindIdentifier || left.Text() != indexName.Text() ||
		right.Kind != ast.KindPropertyAccessExpression ||
		right.AsPropertyAccessExpression().Name().Text() != "length" {
		return nil, nil
	}

	if !independentAwaitLoopIncrementsByOne(ast.SkipParentheses(statement.Incrementor), indexName.Text()) {
		return nil, nil
	}
	return indexName, right.AsPropertyAccessExpression().Expression
}

// independentAwaitLoopIncrementsByOne accepts `i++`, `++i` and `i += 1`, the three spellings of
// stepping through a collection one element at a time.
func independentAwaitLoopIncrementsByOne(incrementor *ast.Node, indexName string) bool {
	isIndex := func(node *ast.Node) bool {
		node = ast.SkipParentheses(node)
		return node.Kind == ast.KindIdentifier && node.Text() == indexName
	}
	switch incrementor.Kind {
	case ast.KindPostfixUnaryExpression:
		unary := incrementor.AsPostfixUnaryExpression()
		return unary.Operator == ast.KindPlusPlusToken && isIndex(unary.Operand)
	case ast.KindPrefixUnaryExpression:
		unary := incrementor.AsPrefixUnaryExpression()
		return unary.Operator == ast.KindPlusPlusToken && isIndex(unary.Operand)
	case ast.KindBinaryExpression:
		binary := incrementor.AsBinaryExpression()
		right := ast.SkipParentheses(binary.Right)
		return binary.OperatorToken.Kind == ast.KindPlusEqualsToken && isIndex(binary.Left) &&
			right.Kind == ast.KindNumericLiteral && right.Text() == "1"
	}
	return false
}

// independentAwaitLoopIsBoundary answers whether a walk stops here because the code below is not
// run by being written: a function of any form, or a class static block.
func independentAwaitLoopIsBoundary(node *ast.Node) bool {
	return ast.IsFunctionLike(node) || node.Kind == ast.KindClassStaticBlockDeclaration
}

// isIndependent runs the four judgments, cheapest first.
func (analysis *independentAwaitLoopAnalysis) isIndependent() bool {
	if !analysis.holdsItsOwnAwait() {
		return false
	}
	if analysis.hasOrderedSideEffect() {
		return false
	}
	if analysis.carriesState() {
		return false
	}
	analysis.computeTaint()
	return !analysis.exitsOnAwaitedValue()
}

// holdsItsOwnAwait answers whether an await in the body runs once per iteration of THIS loop rather
// than of a loop nested inside it.
func (analysis *independentAwaitLoopAnalysis) holdsItsOwnAwait() bool {
	found := false
	var visit func(node *ast.Node, owned bool)
	visit = func(node *ast.Node, owned bool) {
		if node == nil || found || independentAwaitLoopIsBoundary(node) {
			return
		}
		switch node.Kind {
		case ast.KindAwaitExpression:
			if owned && !analysis.awaitsTheItemItself(node.AsAwaitExpression().Expression) {
				found = true
				return
			}
		case ast.KindVariableDeclarationList:
			// `await using` waits at scope exit, once per iteration of the loop it sits in.
			if owned && ast.IsVarAwaitUsing(node) {
				found = true
				return
			}
		case ast.KindForStatement:
			// The initializer runs once per outer iteration; everything else belongs to the inner loop.
			statement := node.AsForStatement()
			visit(statement.Initializer, owned)
			visit(statement.Condition, false)
			visit(statement.Incrementor, false)
			visit(statement.Statement, false)
			return
		case ast.KindForOfStatement, ast.KindForInStatement:
			statement := node.AsForInOrOfStatement()
			// A nested `for await` waits for its first item inside this iteration.
			if owned && node.Kind == ast.KindForOfStatement && statement.AwaitModifier != nil {
				found = true
				return
			}
			visit(statement.Expression, owned)
			visit(statement.Initializer, false)
			visit(statement.Statement, false)
			return
		case ast.KindWhileStatement:
			statement := node.AsWhileStatement()
			visit(statement.Expression, false)
			visit(statement.Statement, false)
			return
		case ast.KindDoStatement:
			statement := node.AsDoStatement()
			visit(statement.Statement, false)
			visit(statement.Expression, false)
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, owned)
			return found
		})
	}
	visit(analysis.body, true)
	return found
}

// awaitsTheItemItself answers whether an await's operand is the collection's own element: the
// for-of binding, or `xs[i]` in the indexed form. Such a promise was started before the loop
// began, so awaiting it in turn serializes no work, and `Promise.all` would change nothing but
// the spelling. Measured on ahra: four loops in PromiseGroup.test.ts collect already-running
// promises this way and were the largest false-positive class before this test existed.
func (analysis *independentAwaitLoopAnalysis) awaitsTheItemItself(operand *ast.Node) bool {
	operand = independentAwaitLoopSkipWrappers(operand)
	switch operand.Kind {
	case ast.KindIdentifier:
		if analysis.loop.Kind != ast.KindForOfStatement {
			return false
		}
		symbol := analysis.symbolOf(operand)
		if symbol == nil {
			return false
		}
		initializer := analysis.loop.AsForInOrOfStatement().Initializer
		for _, declaration := range symbol.Declarations {
			if declaration.Pos() >= initializer.Pos() && declaration.End() <= initializer.End() {
				return true
			}
		}
	case ast.KindElementAccessExpression:
		if analysis.indexSymbol == nil {
			return false
		}
		access := operand.AsElementAccessExpression()
		argument := independentAwaitLoopSkipWrappers(access.ArgumentExpression)
		collection := independentAwaitLoopSegments(analysis.collection)
		return argument.Kind == ast.KindIdentifier && analysis.symbolOf(argument) == analysis.indexSymbol &&
			collection != nil && strings.Join(independentAwaitLoopSegments(access.Expression), ".") == strings.Join(collection, ".")
	}
	return false
}

// hasOrderedSideEffect looks for output, pacing, progress and yields anywhere the iteration runs,
// callbacks included.
//
// Inside a `catch` the question narrows to the one asked of a callee: a `console.warn` reporting
// that one item failed is not output whose order the reader follows. Measured on ahra: a recall
// sample of upstream sites this rule declined found GmailApi.ts:583 silenced only by
// `console.warn('[Gmail] Failed to fetch draft:', ...)` in its catch, a loop that is otherwise the
// same shape as EmailApi.ts:504, which reports.
func (analysis *independentAwaitLoopAnalysis) hasOrderedSideEffect() bool {
	found := false
	var visit func(node *ast.Node, insideFunction bool, insideCatch bool)
	visit = func(node *ast.Node, insideFunction bool, insideCatch bool) {
		if node == nil || found {
			return
		}
		switch node.Kind {
		case ast.KindYieldExpression:
			if !insideFunction {
				found = true
				return
			}
		case ast.KindCallExpression:
			callee := node.AsCallExpression().Expression
			ordered := independentAwaitLoopIsOrderedCall(callee)
			if insideCatch {
				ordered = independentAwaitLoopIsCalleeOrderedCall(callee)
			}
			if ordered || analysis.calleeMakesOrderedCall(callee) {
				found = true
				return
			}
		}
		nextInsideFunction := insideFunction || independentAwaitLoopIsBoundary(node)
		nextInsideCatch := insideCatch || node.Kind == ast.KindCatchClause
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, nextInsideFunction, nextInsideCatch)
			return found
		})
	}
	visit(analysis.body, false, false)
	return found
}

// calleeMakesOrderedCall follows a call to the function it names, one level, and answers whether
// that function's own body makes an ordered call.
//
// The loop body is not where a CLI prints. `await generateFacetMonthly(facet, month)` prints a
// multi-line block per month from inside the callee, and `announceDelivery(recipient, await ...)`
// prints a line per recipient, so the order the loop imposes is the order the reader sees.
// Measured on ahra: three of the first pass's thirteen false positives were exactly this, each with
// `console.log` directly in the callee.
//
// One level, deliberately. Following calls further reaches the shared helpers every module calls,
// and a log line three frames down in a request wrapper would silence nearly every API loop. The
// callee is resolved through the checker, so an imported function is followed to its source; a
// declaration file has no body and ends the walk.
func (analysis *independentAwaitLoopAnalysis) calleeMakesOrderedCall(callee *ast.Node) bool {
	callee = independentAwaitLoopSkipWrappers(callee)
	var name *ast.Node
	switch callee.Kind {
	case ast.KindIdentifier:
		name = callee
	case ast.KindPropertyAccessExpression:
		name = callee.AsPropertyAccessExpression().Name()
	default:
		return false
	}
	symbol := analysis.ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil {
		return false
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = checker.SkipAlias(symbol, analysis.ctx.TypeChecker)
		if symbol == nil {
			return false
		}
	}
	for _, declaration := range symbol.Declarations {
		if body := independentAwaitLoopFunctionBody(declaration); body != nil && analysis.bodyMakesOrderedCall(body) {
			return true
		}
	}
	return false
}

// independentAwaitLoopFunctionBody answers the body a declaration runs when called, or nil: a
// function or method directly, or a variable or class field initialized with a function.
func independentAwaitLoopFunctionBody(declaration *ast.Node) *ast.Node {
	if sourceFile := ast.GetSourceFileOfNode(declaration); sourceFile == nil || sourceFile.IsDeclarationFile {
		return nil
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return declaration.Body()
	case ast.KindVariableDeclaration, ast.KindPropertyDeclaration:
		var initializer *ast.Node
		if declaration.Kind == ast.KindVariableDeclaration {
			initializer = declaration.AsVariableDeclaration().Initializer
		} else {
			initializer = declaration.AsPropertyDeclaration().Initializer
		}
		if initializer == nil {
			return nil
		}
		initializer = independentAwaitLoopSkipWrappers(initializer)
		if initializer.Kind == ast.KindFunctionExpression || initializer.Kind == ast.KindArrowFunction {
			return initializer.Body()
		}
	}
	return nil
}

// bodyMakesOrderedCall scans one function body for a call ordered by name, memoized per body
// because the same helper is called from many loops in one file.
func (analysis *independentAwaitLoopAnalysis) bodyMakesOrderedCall(body *ast.Node) bool {
	if answer, known := analysis.orderedBodies[body]; known {
		return answer
	}
	found := false
	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		// A failure path is not the output a reader watches. Measured on ahra: the first version of
		// this scan counted catch blocks and silenced thirteen loops, among them every boot loader
		// (`console.error` when a domain fails to load), FinanceSync (`delay` before its one retry)
		// and AhraOsTriggers (`delay` before a gh retry). All of them sit in a catch.
		if node.Kind == ast.KindCatchClause {
			return
		}
		if node.Kind == ast.KindCallExpression && independentAwaitLoopIsCalleeOrderedCall(node.AsCallExpression().Expression) {
			found = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(body)
	analysis.orderedBodies[body] = found
	return found
}

// independentAwaitLoopCalleeQuietConsoleMethods are the console methods a callee uses to report a
// problem rather than to show progress, so they do not make its caller's loop ordered.
var independentAwaitLoopCalleeQuietConsoleMethods = map[string]bool{
	"error": true, "warn": true, "debug": true, "trace": true, "assert": true,
}

// independentAwaitLoopIsCalleeOrderedCall is the narrower question asked inside a called function:
// progress output and explicit pacing only.
//
// Two exclusions from the loop-body list. A callee's `console.error` and `console.warn` report a
// failure (`sendOnMeta`'s advisory safety warning is the measured case), and a callee's `setTimeout`
// is almost always a request timeout rather than a sleep (`FrameTvApi.sendRequest` arms one per
// request), so neither says the caller's loop must run in order.
func independentAwaitLoopIsCalleeOrderedCall(callee *ast.Node) bool {
	callee = independentAwaitLoopSkipWrappers(callee)
	switch callee.Kind {
	case ast.KindIdentifier:
		if name := callee.Text(); name == "setTimeout" || name == "setInterval" || name == "setImmediate" {
			return false
		}
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		receiver := independentAwaitLoopSegments(access.Expression)
		method := access.Name().Text()
		if len(receiver) > 0 && receiver[0] == "console" && independentAwaitLoopCalleeQuietConsoleMethods[method] {
			return false
		}
		if len(receiver) >= 2 && receiver[0] == "process" && receiver[1] == "stderr" {
			return false
		}
		if len(receiver) > 0 && strings.Contains(strings.ToLower(receiver[len(receiver)-1]), "log") &&
			(method == "error" || method == "warn" || method == "debug" || method == "trace") {
			return false
		}
		if method == "setTimeout" || method == "setInterval" || method == "setImmediate" {
			return false
		}
	}
	return independentAwaitLoopIsOrderedCall(callee)
}

var independentAwaitLoopPacingPrefixes = []string{"sleep", "delay", "wait", "pause", "throttle", "backoff", "ratelimit", "print"}

var independentAwaitLoopStreamWrites = map[string]bool{
	"write": true, "writeSync": true, "appendFile": true, "appendFileSync": true,
}

var independentAwaitLoopLogMethods = map[string]bool{
	"log": true, "info": true, "warn": true, "error": true, "debug": true, "trace": true,
}

// independentAwaitLoopIsOrderedCall answers whether a call's callee names an effect whose order the
// reader can see. Names only; the doc comment on the rule carries the whole list.
func independentAwaitLoopIsOrderedCall(callee *ast.Node) bool {
	callee = independentAwaitLoopSkipWrappers(callee)
	name := ""
	var receiver []string
	switch callee.Kind {
	case ast.KindIdentifier:
		name = callee.Text()
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		name = access.Name().Text()
		receiver = independentAwaitLoopSegments(access.Expression)
	default:
		return false
	}

	if len(receiver) > 0 && receiver[0] == "console" {
		return true
	}
	if len(receiver) >= 2 && receiver[0] == "process" && (receiver[1] == "stdout" || receiver[1] == "stderr") {
		return true
	}
	if name == "setTimeout" || name == "setInterval" || name == "setImmediate" {
		return true
	}

	lowerName := strings.ToLower(name)
	for _, prefix := range independentAwaitLoopPacingPrefixes {
		if strings.HasPrefix(lowerName, prefix) {
			return true
		}
	}
	if strings.Contains(lowerName, "progress") || strings.Contains(lowerName, "spinner") {
		return true
	}
	if independentAwaitLoopStreamWrites[name] {
		return true
	}

	lastReceiver := ""
	if len(receiver) > 0 {
		lastReceiver = strings.ToLower(receiver[len(receiver)-1])
	}
	if strings.Contains(lastReceiver, "progress") || strings.Contains(lastReceiver, "spinner") {
		return true
	}
	if independentAwaitLoopLogMethods[name] && strings.Contains(lastReceiver, "log") {
		return true
	}
	return callee.Kind == ast.KindIdentifier && name == "log"
}

// independentAwaitLoopSkipWrappers removes parentheses and non-null assertions, which change
// nothing about which value an expression names.
func independentAwaitLoopSkipWrappers(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		default:
			return node
		}
	}
	return node
}

// independentAwaitLoopSegments spells a property chain as its names, `this.cache.entries` as
// `this`, `cache`, `entries`, or nil when the expression is not a plain chain.
func independentAwaitLoopSegments(node *ast.Node) []string {
	node = independentAwaitLoopSkipWrappers(node)
	switch node.Kind {
	case ast.KindIdentifier:
		return []string{node.Text()}
	case ast.KindThisKeyword:
		return []string{"this"}
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		base := independentAwaitLoopSegments(access.Expression)
		if base == nil {
			return nil
		}
		return append(append([]string{}, base...), access.Name().Text())
	}
	return nil
}

// independentAwaitLoopRoot answers the identifier or `this` a property chain starts from.
func independentAwaitLoopRoot(node *ast.Node) *ast.Node {
	node = independentAwaitLoopSkipWrappers(node)
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword:
		return node
	case ast.KindPropertyAccessExpression:
		return independentAwaitLoopRoot(node.AsPropertyAccessExpression().Expression)
	}
	return nil
}

// symbolOf resolves an identifier occurrence to its binding, reading a shorthand property's value
// rather than the property it creates.
func (analysis *independentAwaitLoopAnalysis) symbolOf(identifier *ast.Node) *ast.Symbol {
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment &&
		identifier.Parent.AsShorthandPropertyAssignment().Name() == identifier {
		return analysis.ctx.TypeChecker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	}
	return analysis.ctx.TypeChecker.GetSymbolAtLocation(identifier)
}

// isOuter answers whether a binding outlives one iteration: no declaration of it lies inside the
// loop. An unresolved name is a global and outer by definition.
func (analysis *independentAwaitLoopAnalysis) isOuter(symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return true
	}
	for _, declaration := range symbol.Declarations {
		if ast.GetSourceFileOfNode(declaration) == analysis.ctx.SourceFile &&
			declaration.Pos() >= analysis.loop.Pos() && declaration.End() <= analysis.loop.End() {
			return false
		}
	}
	return true
}

// rootIsOuter answers whether a property chain's root outlives an iteration.
func (analysis *independentAwaitLoopAnalysis) rootIsOuter(root *ast.Node) bool {
	if root == nil {
		return false
	}
	if root.Kind == ast.KindThisKeyword {
		return true
	}
	return analysis.isOuter(analysis.symbolOf(root))
}

// independentAwaitLoopSinkMethods fill a collection without reading it back, which is what
// `Promise.all` returns anyway.
var independentAwaitLoopSinkMethods = map[string]bool{
	"push": true, "unshift": true, "set": true, "add": true, "append": true, "delete": true,
}

// independentAwaitLoopIsSinkMethod answers whether a method fills its receiver: one of the exact
// names above, or a collection-building verb followed by a capitalized noun (`addPage`,
// `appendChild`, `insertRow`). The prefix form exists because a shared document built page by page
// is the same shape as a pushed array, and the one ahra instance (`document.addPage` per slide,
// with `document` also handed to the await) read as independent until it was a sink whose receiver
// the body reads back. `address` and `inserted` stay out because the letter after the verb must be
// upper case.
func independentAwaitLoopIsSinkMethod(method string) bool {
	if independentAwaitLoopSinkMethods[method] {
		return true
	}
	for _, verb := range []string{"add", "append", "insert"} {
		if len(method) > len(verb) && strings.HasPrefix(method, verb) && method[len(verb)] >= 'A' && method[len(verb)] <= 'Z' {
			return true
		}
	}
	return false
}

// independentAwaitLoopConsumingMethods read outer state and change it in one motion, so the next
// iteration sees what this one left.
var independentAwaitLoopConsumingMethods = map[string]bool{
	"pop": true, "shift": true, "splice": true, "next": true, "read": true, "sort": true, "reverse": true,
}

// isNameOnly answers whether an identifier occurrence names something other than a binding read or
// write: a declaration's own name, a member name, a label, or a type position.
func independentAwaitLoopIsNameOnly(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return true
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Name() == identifier
	case ast.KindBreakStatement, ast.KindContinueStatement, ast.KindLabeledStatement,
		ast.KindJsxAttribute, ast.KindQualifiedName:
		return true
	case ast.KindShorthandPropertyAssignment:
		return false
	}
	return ast.IsDeclarationName(identifier) || ast.IsPartOfTypeNode(identifier)
}

// carriesState looks for anything one iteration leaves for the next.
func (analysis *independentAwaitLoopAnalysis) carriesState() bool {
	var mutated []independentAwaitLoopPath
	var reads []independentAwaitLoopPath
	sinkRoots := map[*ast.Node]bool{}
	carried := false

	recordSink := func(target *ast.Node) {
		root := independentAwaitLoopRoot(target)
		if root == nil || !analysis.rootIsOuter(root) {
			return
		}
		mutated = append(mutated, independentAwaitLoopPath{segments: independentAwaitLoopSegments(target), root: root})
		sinkRoots[root] = true
	}

	// outerPropertyTarget answers the outer chain a property or element write lands in: the whole
	// chain for `a.b.c = x`, the object for `a.b[key] = x`. Nil when the target is not outer.
	outerPropertyTarget := func(target *ast.Node) *ast.Node {
		target = independentAwaitLoopSkipWrappers(target)
		switch target.Kind {
		case ast.KindPropertyAccessExpression:
			if analysis.rootIsOuter(independentAwaitLoopRoot(target)) {
				return target
			}
		case ast.KindElementAccessExpression:
			object := target.AsElementAccessExpression().Expression
			if analysis.rootIsOuter(independentAwaitLoopRoot(object)) {
				return object
			}
		}
		return nil
	}

	var visit func(node *ast.Node, insideFunction bool)
	visit = func(node *ast.Node, insideFunction bool) {
		if node == nil || carried {
			return
		}
		switch node.Kind {
		case ast.KindVariableDeclarationList:
			// A `var` is function-scoped, so the value one iteration writes is there for the next.
			if !insideFunction && node.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConst|ast.NodeFlagsUsing) == 0 {
				carried = true
				return
			}

		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				if target := outerPropertyTarget(binary.Left); target != nil {
					if binary.OperatorToken.Kind != ast.KindEqualsToken {
						carried = true
						return
					}
					recordSink(target)
				}
			}

		case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
			var operator ast.Kind
			var operand *ast.Node
			if node.Kind == ast.KindPrefixUnaryExpression {
				operator, operand = node.AsPrefixUnaryExpression().Operator, node.AsPrefixUnaryExpression().Operand
			} else {
				operator, operand = node.AsPostfixUnaryExpression().Operator, node.AsPostfixUnaryExpression().Operand
			}
			if (operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken) &&
				outerPropertyTarget(operand) != nil {
				carried = true
				return
			}

		case ast.KindDeleteExpression:
			if target := outerPropertyTarget(node.AsDeleteExpression().Expression); target != nil {
				recordSink(target)
			}

		case ast.KindCallExpression:
			call := node.AsCallExpression()
			callee := independentAwaitLoopSkipWrappers(call.Expression)
			if callee.Kind == ast.KindPropertyAccessExpression {
				access := callee.AsPropertyAccessExpression()
				method := access.Name().Text()
				receiverRoot := independentAwaitLoopRoot(access.Expression)
				if receiverRoot != nil && analysis.rootIsOuter(receiverRoot) {
					if independentAwaitLoopConsumingMethods[method] {
						carried = true
						return
					}
					if independentAwaitLoopIsSinkMethod(method) {
						recordSink(access.Expression)
					}
				}
				// `Object.assign(target, ...)` fills its first argument.
				if method == "assign" && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					if objectSegments := independentAwaitLoopSegments(access.Expression); len(objectSegments) == 1 &&
						objectSegments[0] == "Object" {
						recordSink(call.Arguments.Nodes[0])
					}
				}
			}

		case ast.KindIdentifier:
			if independentAwaitLoopIsNameOnly(node) {
				break
			}
			symbol := analysis.symbolOf(node)
			if reference.WritesToBinding(node) {
				// The indexed form's counter is declared in the header, so it reads as local, and a
				// write to it in the body is exactly the loop-carried state this exists to catch.
				if analysis.isOuter(symbol) || (analysis.indexSymbol != nil && symbol == analysis.indexSymbol) {
					carried = true
					return
				}
				break
			}
			if analysis.isOuter(symbol) {
				reads = append(reads, independentAwaitLoopPath{segments: independentAwaitLoopReadSegments(node), root: node})
			}

		case ast.KindThisKeyword:
			reads = append(reads, independentAwaitLoopPath{segments: independentAwaitLoopReadSegments(node), root: node})
		}

		nextInsideFunction := insideFunction || independentAwaitLoopIsBoundary(node)
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, nextInsideFunction)
			return carried
		})
	}
	visit(analysis.body, false)
	if carried {
		return true
	}
	if len(mutated) == 0 {
		return false
	}

	// The iterable is read by the loop itself, so a body that grows the collection it walks is
	// a work queue rather than a map.
	var collectionReads []independentAwaitLoopPath
	var collect func(node *ast.Node)
	collect = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier && !independentAwaitLoopIsNameOnly(node) ||
			node.Kind == ast.KindThisKeyword {
			collectionReads = append(collectionReads, independentAwaitLoopPath{segments: independentAwaitLoopReadSegments(node), root: node})
		}
		node.ForEachChild(func(child *ast.Node) bool {
			collect(child)
			return false
		})
	}
	collect(analysis.collection)

	for _, sink := range mutated {
		for _, read := range append(reads, collectionReads...) {
			if sinkRoots[read.root] {
				continue
			}
			if independentAwaitLoopPathsOverlap(sink.segments, read.segments) {
				return true
			}
		}
	}
	return false
}

// independentAwaitLoopReadSegments spells the longest property chain a read starts, so that
// `results.length` is one read of `results.length` rather than a read of `results`.
func independentAwaitLoopReadSegments(root *ast.Node) []string {
	segments := independentAwaitLoopSegments(root)
	current := root
	for {
		parent := current.Parent
		for parent != nil && (parent.Kind == ast.KindParenthesizedExpression || parent.Kind == ast.KindNonNullExpression) {
			current = parent
			parent = current.Parent
		}
		if parent == nil || parent.Kind != ast.KindPropertyAccessExpression ||
			parent.AsPropertyAccessExpression().Expression != current {
			return segments
		}
		segments = append(segments, parent.AsPropertyAccessExpression().Name().Text())
		current = parent
	}
}

// independentAwaitLoopPathsOverlap answers whether one chain is a prefix of the other: reading
// `cursor` whole sees a write to `cursor.after`, and reading `results.length` sees a push to
// `results`, while `this.cache` and `this.fetch` are unrelated.
func independentAwaitLoopPathsOverlap(first []string, second []string) bool {
	if len(first) == 0 || len(second) == 0 {
		return false
	}
	shorter := min(len(first), len(second))
	for index := range shorter {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

// computeTaint finds the bindings declared in the body whose value derives from an await, to a
// fixpoint, so `const response = await f(); const ok = response.ok;` taints both.
func (analysis *independentAwaitLoopAnalysis) computeTaint() {
	taintNames := func(name *ast.Node) bool {
		changed := false
		var mark func(node *ast.Node)
		mark = func(node *ast.Node) {
			if node == nil {
				return
			}
			if node.Kind == ast.KindIdentifier {
				if symbol := analysis.ctx.TypeChecker.GetSymbolAtLocation(node); symbol != nil && !analysis.taintedSymbols[symbol] {
					analysis.taintedSymbols[symbol] = true
					changed = true
				}
				return
			}
			node.ForEachChild(func(child *ast.Node) bool {
				// A default value inside a pattern is an expression, not a name being bound.
				if child.Kind == ast.KindIdentifier && node.Kind == ast.KindBindingElement &&
					node.AsBindingElement().Initializer == child {
					return false
				}
				mark(child)
				return false
			})
		}
		mark(name)
		return changed
	}

	for range 8 {
		changed := false
		var visit func(node *ast.Node)
		visit = func(node *ast.Node) {
			if node == nil || independentAwaitLoopIsBoundary(node) {
				return
			}
			switch node.Kind {
			case ast.KindVariableDeclaration:
				declaration := node.AsVariableDeclaration()
				if declaration.Initializer != nil && analysis.containsTaint(declaration.Initializer) {
					changed = taintNames(declaration.Name()) || changed
				}
			case ast.KindBinaryExpression:
				binary := node.AsBinaryExpression()
				if ast.IsAssignmentOperator(binary.OperatorToken.Kind) && analysis.containsTaint(binary.Right) {
					changed = taintNames(binary.Left) || changed
				}
			case ast.KindCatchClause:
				clause := node.AsCatchClause()
				if clause.VariableDeclaration != nil && analysis.containsAwait(node.Parent.AsTryStatement().TryBlock, -1) {
					changed = taintNames(clause.VariableDeclaration.AsVariableDeclaration().Name()) || changed
				}
			case ast.KindForOfStatement, ast.KindForInStatement:
				statement := node.AsForInOrOfStatement()
				if analysis.containsTaint(statement.Expression) {
					changed = taintNames(statement.Initializer) || changed
				}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				visit(child)
				return false
			})
		}
		visit(analysis.body)
		if !changed {
			return
		}
	}
}

// containsTaint answers whether an expression awaits, or reads a binding derived from an await.
func (analysis *independentAwaitLoopAnalysis) containsTaint(node *ast.Node) bool {
	found := false
	var visit func(current *ast.Node)
	visit = func(current *ast.Node) {
		if current == nil || found || independentAwaitLoopIsBoundary(current) {
			return
		}
		switch current.Kind {
		case ast.KindAwaitExpression:
			found = true
			return
		case ast.KindIdentifier:
			if !independentAwaitLoopIsNameOnly(current) && analysis.taintedSymbols[analysis.symbolOf(current)] {
				found = true
				return
			}
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(node)
	return found
}

// containsAwait answers whether a subtree awaits, optionally only before a position (-1 for
// anywhere).
func (analysis *independentAwaitLoopAnalysis) containsAwait(node *ast.Node, before int) bool {
	found := false
	var visit func(current *ast.Node)
	visit = func(current *ast.Node) {
		if current == nil || found || independentAwaitLoopIsBoundary(current) {
			return
		}
		if current.Kind == ast.KindAwaitExpression && (before < 0 || current.Pos() < before) {
			found = true
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(node)
	return found
}

// exitsOnAwaitedValue looks for a jump out of the loop whose guard reads an awaited value.
func (analysis *independentAwaitLoopAnalysis) exitsOnAwaitedValue() bool {
	found := false
	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || found || independentAwaitLoopIsBoundary(node) {
			return
		}
		switch node.Kind {
		case ast.KindBreakStatement, ast.KindContinueStatement, ast.KindReturnStatement, ast.KindThrowStatement:
			if analysis.exitsLoop(node) && analysis.isGuardedByTaint(node, analysis.body, true) {
				found = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(analysis.body)
	return found
}

// exitsLoop answers whether a jump leaves this loop rather than a construct inside it.
func (analysis *independentAwaitLoopAnalysis) exitsLoop(jump *ast.Node) bool {
	switch jump.Kind {
	case ast.KindReturnStatement:
		return true

	case ast.KindThrowStatement:
		// Caught inside the body, a throw is a jump to the handler, not out of the loop.
		for child, parent := jump, jump.Parent; parent != nil && child != analysis.body; child, parent = parent, parent.Parent {
			if parent.Kind == ast.KindTryStatement {
				try := parent.AsTryStatement()
				if try.TryBlock == child && try.CatchClause != nil {
					return false
				}
			}
		}
		return true

	case ast.KindBreakStatement, ast.KindContinueStatement:
		var label *ast.Node
		if jump.Kind == ast.KindBreakStatement {
			label = jump.AsBreakStatement().Label
		} else {
			label = jump.AsContinueStatement().Label
		}
		isBreak := jump.Kind == ast.KindBreakStatement

		if label == nil {
			// The nearest loop (or, for break, switch) is the target.
			for current := jump.Parent; current != nil; current = current.Parent {
				if current == analysis.loop {
					return isBreak
				}
				if ast.IsIterationStatement(current, false) || (isBreak && current.Kind == ast.KindSwitchStatement) {
					return false
				}
			}
			return false
		}

		for current := jump.Parent; current != nil && current != analysis.loop; current = current.Parent {
			if current.Kind == ast.KindLabeledStatement && current.AsLabeledStatement().Label.Text() == label.Text() {
				return false
			}
		}
		for current := analysis.loop.Parent; current != nil && current.Kind == ast.KindLabeledStatement; current = current.Parent {
			if current.AsLabeledStatement().Label.Text() == label.Text() {
				return isBreak
			}
		}
		// A label outside this loop: the jump leaves it for an outer one.
		return true
	}
	return false
}

// isGuardedByTaint climbs from a jump to `limit`, asking at each step whether what decides that
// the jump runs reads an awaited value. `withSiblings` also asks whether an earlier statement jumps
// on such a value, which would decide whether this statement is reached at all.
func (analysis *independentAwaitLoopAnalysis) isGuardedByTaint(jump *ast.Node, limit *ast.Node, withSiblings bool) bool {
	switch jump.Kind {
	case ast.KindReturnStatement:
		if analysis.containsTaint(jump.AsReturnStatement().Expression) {
			return true
		}
	case ast.KindThrowStatement:
		if analysis.containsTaint(jump.AsThrowStatement().Expression) {
			return true
		}
	}

	for child, parent := jump, jump.Parent; parent != nil && child != limit; child, parent = parent, parent.Parent {
		switch parent.Kind {
		case ast.KindIfStatement:
			if analysis.containsTaint(parent.AsIfStatement().Expression) {
				return true
			}
		case ast.KindCaseClause, ast.KindDefaultClause:
			caseBlock := parent.Parent
			switchStatement := caseBlock.Parent.AsSwitchStatement()
			if analysis.containsTaint(switchStatement.Expression) {
				return true
			}
			for _, clause := range caseBlock.AsCaseBlock().Clauses.Nodes {
				if analysis.containsTaint(clause.AsCaseOrDefaultClause().Expression) {
					return true
				}
			}
		case ast.KindWhileStatement:
			if analysis.containsTaint(parent.AsWhileStatement().Expression) {
				return true
			}
		case ast.KindDoStatement:
			if analysis.containsTaint(parent.AsDoStatement().Expression) {
				return true
			}
		case ast.KindForStatement:
			if analysis.containsTaint(parent.AsForStatement().Condition) {
				return true
			}
		case ast.KindForOfStatement, ast.KindForInStatement:
			if analysis.containsTaint(parent.AsForInOrOfStatement().Expression) {
				return true
			}
		case ast.KindCatchClause:
			// Reached only when the try threw, and an await is what throws asynchronously.
			if analysis.containsAwait(parent.Parent.AsTryStatement().TryBlock, -1) {
				return true
			}
		case ast.KindTryStatement:
			// `try { return await probe(); } catch {}` exits only when the await resolved.
			try := parent.AsTryStatement()
			if try.TryBlock == child && try.CatchClause != nil && analysis.containsAwait(try.TryBlock, jump.Pos()) {
				return true
			}
		case ast.KindBlock:
			if withSiblings && analysis.earlierSiblingJumpsOnTaint(parent.AsBlock().Statements.Nodes, child) {
				return true
			}
		}
		if parent.Kind == ast.KindCaseClause || parent.Kind == ast.KindDefaultClause {
			if withSiblings && analysis.earlierSiblingJumpsOnTaint(parent.AsCaseOrDefaultClause().Statements.Nodes, child) {
				return true
			}
		}
	}
	return false
}

// earlierSiblingJumpsOnTaint answers whether a statement before `child` holds a jump guarded by an
// awaited value. Conservative: any such jump counts, even one whose target lies inside the sibling.
func (analysis *independentAwaitLoopAnalysis) earlierSiblingJumpsOnTaint(statements []*ast.Node, child *ast.Node) bool {
	for _, sibling := range statements {
		if sibling == child || sibling.Pos() >= child.Pos() {
			return false
		}
		found := false
		var visit func(node *ast.Node)
		visit = func(node *ast.Node) {
			if node == nil || found || independentAwaitLoopIsBoundary(node) {
				return
			}
			switch node.Kind {
			case ast.KindBreakStatement, ast.KindContinueStatement, ast.KindReturnStatement, ast.KindThrowStatement:
				if analysis.isGuardedByTaint(node, sibling, false) {
					found = true
					return
				}
			}
			node.ForEachChild(func(grandchild *ast.Node) bool {
				visit(grandchild)
				return found
			})
		}
		visit(sibling)
		if found {
			return true
		}
	}
	return false
}

// independentAwaitLoopHeaderRange spans `for (...)`, from the keyword to the closing parenthesis.
func independentAwaitLoopHeaderRange(sourceFile *ast.SourceFile, loop *ast.Node, body *ast.Node) core.TextRange {
	start := rule.TokenRange(sourceFile, loop).Pos()
	bodyStart := rule.TokenRange(sourceFile, body).Pos()
	header := strings.TrimRight(sourceFile.Text()[start:bodyStart], " \t\r\n")
	return core.NewTextRange(start, start+len(header))
}
