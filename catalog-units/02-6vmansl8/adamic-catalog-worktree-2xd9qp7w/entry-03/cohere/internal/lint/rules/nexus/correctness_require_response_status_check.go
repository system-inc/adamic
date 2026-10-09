package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const (
	correctnessRequireResponseStatusCheckBodyId    = "bodyReadWithoutStatusCheck"
	correctnessRequireResponseStatusCheckDiscardId = "responseDiscarded"
)

// The rule's messages, whose wording lives in `policy/messages/correctness-require-response-status-check.json`.
var (
	correctnessRequireResponseStatusCheckBodyText    = policy.MessageOf("nexus/correctness-require-response-status-check", correctnessRequireResponseStatusCheckBodyId)
	correctnessRequireResponseStatusCheckDiscardText = policy.MessageOf("nexus/correctness-require-response-status-check", correctnessRequireResponseStatusCheckDiscardId)
)

func correctnessRequireResponseStatusCheckBodyMessage() rule.Message {
	return rule.Message{Id: correctnessRequireResponseStatusCheckBodyId, Description: correctnessRequireResponseStatusCheckBodyText.Render(nil)}
}

func correctnessRequireResponseStatusCheckDiscardMessage() rule.Message {
	return rule.Message{Id: correctnessRequireResponseStatusCheckDiscardId, Description: correctnessRequireResponseStatusCheckDiscardText.Render(nil)}
}

// CorrectnessRequireResponseStatusCheck reports a fetch Response whose body is read on a path that
// never asks whether the request succeeded, and a fetch Response awaited and thrown away.
//
//	invalid: const response = await fetch(url); return (await response.json()) as Task[];
//	invalid: await networkService.request('/api/finance/categorize', { method: 'POST', body });
//	invalid: return (await fetch(url)).json();
//	invalid: fetch(url).then(function(response) { return response.json(); });
//	invalid: const data = await response.json(); if(data.error) throw new Error(data.error); return data;
//	valid:   const response = await fetch(url); if(!response.ok) throw new Error(...); return response.json();
//	valid:   const data = await response.json(); if(!response.ok) throw new Error(data.message); return data;
//	valid:   const response = await fetch(url); return parseResponse(response);
//	valid:   const response = await fetch(url); const text = await response.text(); throw new Error(text);
//
// # Where it came from
//
// Two sites in ahra, found by the cross-language pass of the new-rules sweep (`#tevhg3f`, after go
// vet's `httpresponse`) and built as task `#3twjf7k`:
//
//   - `app/(os-layout)/_hooks/useTaskInboxRequest.ts:19`: the inbox read awaits
//     `networkService.request('/api/tasks?inbox=true')` and parses `.json()` straight away. On a 401
//     or a 500 the error body parses, `data.tasks` is `undefined`, and that is returned as the task
//     list, which the read request caches as a success. The UI fails later, far from the cause.
//   - `app/(os-layout)/finance/_components/FinanceReviewView.tsx:44`: `postCategorize` awaits the
//     categorize POST and returns nothing, so a categorize rejected with 400 or 500 reports success,
//     invalidates the cache, and the row comes back as if the click did nothing.
//
// # Which Responses it tracks
//
// A Response is tracked only when it comes from a call known not to throw on an HTTP error, decided
// by the symbol the callee resolves to, never by its spelling:
//
//   - **The platform's `fetch`**: a call to `fetch` (or `globalThis.fetch`, `window.fetch`) whose
//     symbol is declared only in declaration files, at global scope or inside `declare global`. That
//     is lib.dom's declaration, `@types/node`'s, and any other ambient typing of the WHATWG fetch,
//     which the specification has resolve on every status. An imported `fetch` resolves to its
//     import binding and is not tracked; neither is a local function named `fetch`.
//   - **`networkService.request` and `networkService.baseApiRequest`** from Structure: the methods
//     of `NetworkService` in `source/services/network/NetworkService.ts`, and of the
//     `NetworkServiceRequestExecutor` they delegate to in `internal/`, matched by the declaration
//     the call resolves to (method, class and file all three). Read before it was included: the
//     executor returns the Response whatever its status (`NetworkServiceRequestExecutor.ts:178`)
//     and only counts `!ok` toward its statistics, so a 500 reaches the caller untouched.
//
// Every other source of a Response is out, and that is deliberate rather than a gap to fill by
// widening the type test. `R2Api.signedS3Request` returns a `Promise<Response>` too and already
// throws on `!ok`, so reading its body unchecked is correct, and a rule that tracked every value of
// type `Response` would report it. A wrapper that passes a Response through unchanged is a missed
// finding here, never a false one.
//
// # What counts as reading the body, and as checking
//
// Each reference to the tracked binding is read by what is done to it, looking out through
// parentheses, `!`, `as` and `satisfies`:
//
//   - **A body read** is a call of `.json()`, `.text()`, `.arrayBuffer()`, `.blob()`, `.formData()`
//     or `.bytes()`, or a read of `.body`, the stream.
//   - **A check** is any read of `.ok`, `.status` or `.statusText`. Any read at all, not only one in a
//     condition: `throw new Error(response.statusText)` and a status logged before a branch both
//     count. Reading more as a check can only silence a site, never report one.
//   - **An escape** is everything else that hands the Response somewhere this rule cannot follow:
//     passing it to a function, returning it, storing it, destructuring it, `.clone()`, an element
//     access with a computed key, a bare truth test. An escape counts as a check. The receiver may
//     read `.ok` (`parseResponse(response)` usually does), and proving that it does would mean
//     following every helper into its body, so the rule assumes it does. A helper that does not is
//     a missed finding.
//   - Everything else is neutral: `.headers`, `.url`, `.redirected`, `.type`, `.bodyUsed`.
//
// # How "on a path" is decided
//
// Through the enclosing function's control-flow graph. A body read is reported when some path
// leaves the binding's declaration, reaches the body read, and goes on to a normal exit of the
// function (falling off the end, or a `return`) or to the binding being declared again on the next
// turn of a loop, without passing a check or an escape anywhere along the way. So:
//
//   - **Checking after reading is checking.** `const data = await response.json(); if(!response.ok)
//     throw new Error(data.message); return data;` is the usual way to surface the server's error
//     text, and every path from the read to the exit passes `.ok`. It stays silent.
//   - **Reading the body to throw it is not trusting it.** A path that leaves through an uncaught
//     `throw` is not an exit for this rule, so `const text = await response.text(); throw new
//     Error(text);` stays silent.
//   - **A check on one branch does not cover the other.** `const data = await response.json();
//     if(data.error) { log(response.status); throw ... } return data;` reports: on the path where
//     the body carries no `error` field, nothing asked whether the request succeeded, and a 404 or a
//     502 whose body has no such field is returned as data. Testing a field of the body is a guess
//     about the server's error shape; the status is the protocol's own answer.
//
// The binding must be a `const` declared from `await <tracked call>`, or the single parameter of a
// function passed straight to `.then` on a tracked call. A `let`, a destructured declaration and a
// Response laundered through `Promise.all` are not followed. A binding any closure refers to is
// dropped whole, because the closure may run before the body read or after it and the graph of the
// enclosing function cannot say which. As a guard against a reference the graph never walked, a
// binding whose references the graph did not all visit is dropped too. Each of these is a missed
// finding. That last guard has never fired: no fixture reaches it and a probe counted zero on ahra,
// www-phi-health, www-connected-app and api-phi-health, so a mutant removing it survives. It stays
// because an unvisited check or escape is exactly what would turn a silent site into a false one.
//
// # The two other shapes it reports
//
//   - **Discarded.** `await fetch(...)` or `await networkService.request(...)` as a statement on its
//     own, and a tracked `const` that is never referenced, both throw the Response away with its
//     status. An unawaited `fetch(...)` statement is not reported: that is a floating promise, and
//     the floating-promise rules own it.
//   - **Read inline.** `(await fetch(url)).json()` reads the body of a Response nothing else can ever
//     see, so nothing can have checked it.
//
// # No fix
//
// What a failure should do is the author's call: throw, return an empty result, retry, or show the
// server's message. Any one of them written in by a fixer would be wrong for most sites.
var CorrectnessRequireResponseStatusCheck = rule.Rule{
	Name: "nexus/correctness-require-response-status-check",

	// Symbol identity is what tells the platform's fetch from a local function of that name, a
	// Structure NetworkService method from any other `request`, and one binding from a shadowing one.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				correctnessRequireResponseStatusCheckScanFile(ctx, sourceFile)
			},
		}
	},
}

// correctnessRequireResponseStatusCheckUse is what one reference does to a tracked Response.
type correctnessRequireResponseStatusCheckUse uint8

const (
	correctnessRequireResponseStatusCheckNeutral correctnessRequireResponseStatusCheckUse = iota
	correctnessRequireResponseStatusCheckBodyRead
	correctnessRequireResponseStatusCheckCheck
	correctnessRequireResponseStatusCheckEscape
)

// correctnessRequireResponseStatusCheckBodyMembers are the members of a Response that consume its
// body. All but `body` are methods and count only when called.
var correctnessRequireResponseStatusCheckBodyMembers = map[string]bool{
	"json": true, "text": true, "arrayBuffer": true, "blob": true, "formData": true, "bytes": true, "body": true,
}

// correctnessRequireResponseStatusCheckStatusMembers are the members whose read asks whether the
// request succeeded.
var correctnessRequireResponseStatusCheckStatusMembers = map[string]bool{
	"ok": true, "status": true, "statusText": true,
}

// correctnessRequireResponseStatusCheckNeutralMembers are the members that neither consume the body
// nor say anything about the status. Any member not named in one of the three sets is an escape.
var correctnessRequireResponseStatusCheckNeutralMembers = map[string]bool{
	"headers": true, "url": true, "redirected": true, "type": true, "bodyUsed": true,
}

// correctnessRequireResponseStatusCheckBinding is one tracked Response: its name node, the code path
// root it lives in, and what the analysis learns about it.
type correctnessRequireResponseStatusCheckBinding struct {
	name     *ast.Node
	symbol   *ast.Symbol
	root     *ast.Node
	producer *ast.Node

	// references are the binding's references outside type positions, in the root itself. A
	// binding with a reference inside a nested function is dropped before this is filled.
	references []*ast.Node
	dropped    bool
}

// correctnessRequireResponseStatusCheckEvent is one thing the graph recorded about a binding.
type correctnessRequireResponseStatusCheckEvent struct {
	binding *correctnessRequireResponseStatusCheckBinding
	declare bool
	use     correctnessRequireResponseStatusCheckUse
	// report is the node a body read is reported at: the call for a method, the access for `.body`.
	report *ast.Node
}

func correctnessRequireResponseStatusCheckScanFile(ctx rule.Context, sourceFile *ast.Node) {
	bindingsByRoot := map[*ast.Node][]*correctnessRequireResponseStatusCheckBinding{}
	var roots []*ast.Node
	track := func(binding *correctnessRequireResponseStatusCheckBinding) {
		if _, seen := bindingsByRoot[binding.root]; !seen {
			roots = append(roots, binding.root)
		}
		bindingsByRoot[binding.root] = append(bindingsByRoot[binding.root], binding)
	}

	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindExpressionStatement:
			awaited := ast.SkipParentheses(node.AsExpressionStatement().Expression)
			if awaited.Kind == ast.KindAwaitExpression {
				call := ast.SkipParentheses(awaited.AsAwaitExpression().Expression)
				if correctnessRequireResponseStatusCheckIsProducer(ctx, call) {
					ctx.ReportNode(node, correctnessRequireResponseStatusCheckDiscardMessage())
				}
			}
		case ast.KindAwaitExpression:
			call := ast.SkipParentheses(node.AsAwaitExpression().Expression)
			if correctnessRequireResponseStatusCheckIsProducer(ctx, call) {
				if binding := correctnessRequireResponseStatusCheckConstBinding(ctx, node); binding != nil {
					track(binding)
				} else if use, report := correctnessRequireResponseStatusCheckClassify(node); use == correctnessRequireResponseStatusCheckBodyRead {
					ctx.ReportNode(report, correctnessRequireResponseStatusCheckBodyMessage())
				}
			}
		case ast.KindCallExpression:
			if binding := correctnessRequireResponseStatusCheckThenBinding(ctx, node); binding != nil {
				track(binding)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	sourceFile.ForEachChild(visit)

	for _, root := range roots {
		correctnessRequireResponseStatusCheckAnalyzeRoot(ctx, root, bindingsByRoot[root])
	}
}

// correctnessRequireResponseStatusCheckIsProducer says whether a node is a call to the platform's
// fetch or to Structure's NetworkService request methods, the calls known not to throw on an HTTP
// error status.
func correctnessRequireResponseStatusCheckIsProducer(ctx rule.Context, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	var name *ast.Node
	switch callee.Kind {
	case ast.KindIdentifier:
		name = callee
	case ast.KindPropertyAccessExpression:
		name = callee.AsPropertyAccessExpression().Name()
	default:
		return false
	}
	if name.Kind != ast.KindIdentifier {
		return false
	}
	switch name.Text() {
	case "fetch":
		return correctnessRequireResponseStatusCheckIsGlobalFetch(ctx.TypeChecker.GetSymbolAtLocation(name))
	case "request", "baseApiRequest":
		return callee.Kind == ast.KindPropertyAccessExpression &&
			correctnessRequireResponseStatusCheckIsNetworkServiceMethod(ctx.TypeChecker.GetSymbolAtLocation(name))
	}
	return false
}

// correctnessRequireResponseStatusCheckIsGlobalFetch says whether a symbol is the global `fetch`:
// every declaration ambient, in a declaration file, at global scope or inside `declare global`.
// Every declaration rather than the first, because lib.dom and `@types/node` each contribute one
// and a project file merging in a third would make the symbol something this rule has not read.
func correctnessRequireResponseStatusCheckIsGlobalFetch(symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration {
			return false
		}
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !file.IsDeclarationFile {
			return false
		}
		container := declaration.Parent
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

// correctnessRequireResponseStatusCheckNetworkServiceDeclarations are the request methods of
// Structure's network service read for this rule, by the file that declares each class. Both
// classes return the Response whatever its status.
var correctnessRequireResponseStatusCheckNetworkServiceDeclarations = []struct {
	fileSuffix string
	className  string
}{
	{"/source/services/network/NetworkService.ts", "NetworkService"},
	{"/source/services/network/internal/NetworkServiceRequestExecutor.ts", "NetworkServiceRequestExecutor"},
}

// correctnessRequireResponseStatusCheckIsNetworkServiceMethod says whether a symbol is
// `request` or `baseApiRequest` declared on one of the two classes above, in its own file.
func correctnessRequireResponseStatusCheckIsNetworkServiceMethod(symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindMethodDeclaration {
			return false
		}
		class := declaration.Parent
		if class == nil || class.Kind != ast.KindClassDeclaration || class.Name() == nil {
			return false
		}
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil {
			return false
		}
		matched := false
		for _, known := range correctnessRequireResponseStatusCheckNetworkServiceDeclarations {
			if class.Name().Text() == known.className && strings.HasSuffix(file.FileName().AsString(), known.fileSuffix) {
				matched = true
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// correctnessRequireResponseStatusCheckConstBinding is the binding `const response = await
// <producer>(...)` declares, when the await is that whole initializer and the name is a plain
// identifier.
func correctnessRequireResponseStatusCheckConstBinding(ctx rule.Context, await *ast.Node) *correctnessRequireResponseStatusCheckBinding {
	declaration := await.Parent
	for declaration != nil && declaration.Kind == ast.KindParenthesizedExpression {
		declaration = declaration.Parent
	}
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration ||
		ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer) != await {
		return nil
	}
	name := declaration.Name()
	list := declaration.Parent
	if name == nil || name.Kind != ast.KindIdentifier || list == nil || list.Kind != ast.KindVariableDeclarationList ||
		list.Flags&ast.NodeFlagsConst == 0 {
		return nil
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	root := control_flow_graph.RootOf(declaration)
	if symbol == nil || root == nil {
		return nil
	}
	return &correctnessRequireResponseStatusCheckBinding{name: name, symbol: symbol, root: root, producer: await}
}

// correctnessRequireResponseStatusCheckThenBinding is the parameter of `producer(...).then(function(
// response) { ... })`, a callback passed straight to `.then` on a tracked call, when that parameter
// is a plain identifier with no default.
func correctnessRequireResponseStatusCheckThenBinding(ctx rule.Context, call *ast.Node) *correctnessRequireResponseStatusCheckBinding {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	access := callee.AsPropertyAccessExpression()
	if access.Name().Kind != ast.KindIdentifier || access.Name().Text() != "then" ||
		!correctnessRequireResponseStatusCheckIsProducer(ctx, ast.SkipParentheses(access.Expression)) {
		return nil
	}
	arguments := call.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return nil
	}
	callback := ast.SkipParentheses(arguments.Nodes[0])
	if callback.Kind != ast.KindArrowFunction && callback.Kind != ast.KindFunctionExpression {
		return nil
	}
	parameters := callback.Parameters()
	if len(parameters) == 0 {
		return nil
	}
	parameter := parameters[0].AsParameterDeclaration()
	name := parameter.Name()
	if name == nil || name.Kind != ast.KindIdentifier || parameter.Initializer != nil || parameter.DotDotDotToken != nil {
		return nil
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil {
		return nil
	}
	return &correctnessRequireResponseStatusCheckBinding{name: name, symbol: symbol, root: callback, producer: call}
}

// correctnessRequireResponseStatusCheckClassify reads what one expression of a Response is used for,
// looking out through the wrappers that do not change the value, and returns the node a body read
// reports at.
func correctnessRequireResponseStatusCheckClassify(expression *ast.Node) (correctnessRequireResponseStatusCheckUse, *ast.Node) {
	outer := expression
	for outer.Parent != nil {
		switch outer.Parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression,
			ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
			outer = outer.Parent
			continue
		}
		break
	}
	parent := outer.Parent
	if parent == nil {
		return correctnessRequireResponseStatusCheckEscape, nil
	}
	isAccess := (parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == outer) ||
		(parent.Kind == ast.KindElementAccessExpression && parent.AsElementAccessExpression().Expression == outer)
	if !isAccess {
		return correctnessRequireResponseStatusCheckEscape, nil
	}
	member, named := property.AccessedName(parent, property.Textual)
	if !named {
		return correctnessRequireResponseStatusCheckEscape, nil
	}
	switch {
	case correctnessRequireResponseStatusCheckStatusMembers[member]:
		return correctnessRequireResponseStatusCheckCheck, nil
	case member == "body":
		return correctnessRequireResponseStatusCheckBodyRead, parent
	case correctnessRequireResponseStatusCheckBodyMembers[member]:
		call := parent.Parent
		if call != nil && call.Kind == ast.KindCallExpression && call.AsCallExpression().Expression == parent {
			return correctnessRequireResponseStatusCheckBodyRead, call
		}
		// `response.json` handed somewhere uncalled, or bound: the body may be read anywhere.
		return correctnessRequireResponseStatusCheckEscape, nil
	case correctnessRequireResponseStatusCheckNeutralMembers[member]:
		return correctnessRequireResponseStatusCheckNeutral, nil
	}
	return correctnessRequireResponseStatusCheckEscape, nil
}

// correctnessRequireResponseStatusCheckAnalyzeRoot collects each binding's references, builds the
// root's graph once for all of them, and reports every body read on an unchecked path.
func correctnessRequireResponseStatusCheckAnalyzeRoot(
	ctx rule.Context,
	root *ast.Node,
	bindings []*correctnessRequireResponseStatusCheckBinding,
) {
	bySymbol := map[*ast.Symbol]*correctnessRequireResponseStatusCheckBinding{}
	byName := map[*ast.Node]*correctnessRequireResponseStatusCheckBinding{}
	names := map[string]bool{}
	for _, binding := range bindings {
		bySymbol[binding.symbol] = binding
		byName[binding.name] = binding
		names[binding.name.Text()] = true
	}
	resolve := func(identifier *ast.Node) *correctnessRequireResponseStatusCheckBinding {
		if identifier.Kind != ast.KindIdentifier || !names[identifier.Text()] || byName[identifier] != nil {
			return nil
		}
		if parent := identifier.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
			if symbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent); symbol != nil {
				return bySymbol[symbol]
			}
		}
		return bySymbol[ctx.TypeChecker.GetSymbolAtLocation(identifier)]
	}

	// Every reference, with the ones inside a nested function dropping their binding and the ones in
	// a type position (`typeof response`) left out, since nothing in a type runs.
	var collect func(node *ast.Node, nested bool) bool
	collect = func(node *ast.Node, nested bool) bool {
		if binding := resolve(node); binding != nil && !correctnessRequireResponseStatusCheckInType(node, root) {
			if nested {
				binding.dropped = true
			} else {
				binding.references = append(binding.references, node)
			}
		}
		innerNested := nested || (node != root && control_flow_graph.IsRoot(node))
		node.ForEachChild(func(child *ast.Node) bool {
			return collect(child, innerNested)
		})
		return false
	}
	collect(root, false)

	for _, binding := range bindings {
		if !binding.dropped && len(binding.references) == 0 {
			ctx.ReportNode(binding.producer, correctnessRequireResponseStatusCheckDiscardMessage())
			binding.dropped = true
		}
	}

	type builder = control_flow_graph.Builder[correctnessRequireResponseStatusCheckEvent]
	walked := map[*ast.Node]bool{}
	graph := control_flow_graph.Build(root, control_flow_graph.Hooks[correctnessRequireResponseStatusCheckEvent]{
		Read: func(b *builder, node *ast.Node) {
			binding := resolve(node)
			if binding == nil || binding.dropped || correctnessRequireResponseStatusCheckInType(node, root) {
				return
			}
			walked[node] = true
			use, report := correctnessRequireResponseStatusCheckClassify(node)
			if use != correctnessRequireResponseStatusCheckNeutral {
				b.Emit(correctnessRequireResponseStatusCheckEvent{binding: binding, use: use, report: report})
			}
		},
		Write: func(b *builder, node *ast.Node) {
			if binding := byName[node]; binding != nil && !binding.dropped {
				b.Emit(correctnessRequireResponseStatusCheckEvent{binding: binding, declare: true})
			}
		},
	})

	final := make([]bool, len(graph.Blocks))
	for _, block := range graph.FinalBlocks {
		final[block.Index()] = true
	}
	for _, binding := range bindings {
		if binding.dropped {
			continue
		}
		unwalked := false
		for _, reference := range binding.references {
			if !walked[reference] {
				unwalked = true
			}
		}
		if unwalked {
			continue
		}
		for _, report := range correctnessRequireResponseStatusCheckUncheckedReads(graph, final, binding) {
			ctx.ReportNode(report, correctnessRequireResponseStatusCheckBodyMessage())
		}
	}
}

// correctnessRequireResponseStatusCheckInType says whether a reference sits in a type position
// below the root, where it names a type and runs nothing.
func correctnessRequireResponseStatusCheckInType(node *ast.Node, root *ast.Node) bool {
	for current := node.Parent; current != nil && current != root; current = current.Parent {
		if current.Kind == ast.KindTypeQuery {
			return true
		}
	}
	return false
}

// correctnessRequireResponseStatusCheckPosition is a place in the graph: an event index in a block.
type correctnessRequireResponseStatusCheckPosition struct {
	block *control_flow_graph.Block[correctnessRequireResponseStatusCheckEvent]
	start int
}

// correctnessRequireResponseStatusCheckUncheckedReads returns the body reads of one binding that lie
// on an unchecked path: reachable from the declaration without passing a check or an escape, and
// leading on to a normal exit, or to the declaration again, without passing one either.
func correctnessRequireResponseStatusCheckUncheckedReads(
	graph *control_flow_graph.Graph[correctnessRequireResponseStatusCheckEvent],
	final []bool,
	binding *correctnessRequireResponseStatusCheckBinding,
) []*ast.Node {
	var starts []correctnessRequireResponseStatusCheckPosition
	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		for index, event := range block.Events {
			if event.binding == binding && event.declare {
				starts = append(starts, correctnessRequireResponseStatusCheckPosition{block: block, start: index + 1})
			}
		}
	}

	// The body reads reachable from a declaration with nothing checked yet.
	var candidates []correctnessRequireResponseStatusCheckPosition
	correctnessRequireResponseStatusCheckWalk(starts, func(block *control_flow_graph.Block[correctnessRequireResponseStatusCheckEvent], index int) bool {
		event := block.Events[index]
		if event.binding != binding {
			return true
		}
		if event.declare || event.use != correctnessRequireResponseStatusCheckBodyRead {
			return false
		}
		candidates = append(candidates, correctnessRequireResponseStatusCheckPosition{block: block, start: index})
		return true
	}, nil)

	var reports []*ast.Node
	reported := map[*ast.Node]bool{}
	for _, candidate := range candidates {
		report := candidate.block.Events[candidate.start].report
		if reported[report] {
			continue
		}
		unchecked := false
		correctnessRequireResponseStatusCheckWalk([]correctnessRequireResponseStatusCheckPosition{{block: candidate.block, start: candidate.start + 1}},
			func(block *control_flow_graph.Block[correctnessRequireResponseStatusCheckEvent], index int) bool {
				event := block.Events[index]
				if event.binding != binding || unchecked {
					return !unchecked
				}
				if event.declare {
					// The next turn of a loop binds a fresh Response; this one was used and dropped.
					unchecked = true
					return false
				}
				return event.use == correctnessRequireResponseStatusCheckBodyRead
			},
			func(block *control_flow_graph.Block[correctnessRequireResponseStatusCheckEvent]) {
				if final[block.Index()] {
					unchecked = true
				}
			})
		if unchecked {
			reported[report] = true
			reports = append(reports, report)
		}
	}
	return reports
}

// correctnessRequireResponseStatusCheckWalk visits the events reachable from the starts, in order
// along each path. visitEvent returns whether the path goes on past that event; completeBlock, when
// given, runs for every block a path walks to its end. Each block is entered from its top at most
// once, which is enough because the state a path carries is only whether it is still going.
func correctnessRequireResponseStatusCheckWalk(
	starts []correctnessRequireResponseStatusCheckPosition,
	visitEvent func(block *control_flow_graph.Block[correctnessRequireResponseStatusCheckEvent], index int) bool,
	completeBlock func(block *control_flow_graph.Block[correctnessRequireResponseStatusCheckEvent]),
) {
	queue := append([]correctnessRequireResponseStatusCheckPosition(nil), starts...)
	entered := map[int]bool{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		stopped := false
		for index := current.start; index < len(current.block.Events); index++ {
			if !visitEvent(current.block, index) {
				stopped = true
				break
			}
		}
		if stopped {
			continue
		}
		if completeBlock != nil {
			completeBlock(current.block)
		}
		for _, successor := range current.block.Successors {
			if successor == nil || !successor.Reachable || entered[successor.Index()] {
				continue
			}
			entered[successor.Index()] = true
			queue = append(queue, correctnessRequireResponseStatusCheckPosition{block: successor})
		}
	}
}
