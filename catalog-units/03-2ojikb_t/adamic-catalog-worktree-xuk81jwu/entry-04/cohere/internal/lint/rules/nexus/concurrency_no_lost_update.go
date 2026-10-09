package nexus

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const concurrencyNoLostUpdateId = "lostUpdate"

// concurrencyNoLostUpdateText is the rule's message, whose wording lives in
// `policy/messages/concurrency-no-lost-update.json`. Both findings share the id, and the phrase says
// which stale read the write builds on: a variable, or an entry a read call returned.
var (
	concurrencyNoLostUpdateText           = policy.MessageOf("nexus/concurrency-no-lost-update", concurrencyNoLostUpdateId)
	concurrencyNoLostUpdateVariableRead   = concurrencyNoLostUpdateText.Option("suspendedRead", "variable")
	concurrencyNoLostUpdateStoreEntryRead = concurrencyNoLostUpdateText.Option("suspendedRead", "storeEntry")
)

// concurrencyNoLostUpdateMessage names the written target, so a reader of the finding does not have
// to work out which of several writes on a line is the one that loses an update.
func concurrencyNoLostUpdateMessage(target string) rule.Message {
	return rule.Message{
		Id:          concurrencyNoLostUpdateId,
		Description: concurrencyNoLostUpdateText.Render(map[string]string{"target": target}, concurrencyNoLostUpdateVariableRead),
	}
}

// concurrencyNoLostUpdateStoreMessage names the read call whose stale answer the write builds on.
func concurrencyNoLostUpdateStoreMessage(read string) rule.Message {
	return rule.Message{
		Id:          concurrencyNoLostUpdateId,
		Description: concurrencyNoLostUpdateText.Render(map[string]string{"read": read}, concurrencyNoLostUpdateStoreEntryRead),
	}
}

// ConcurrencyNoLostUpdate reports a read-modify-write that spans a suspension: an assignment whose
// new value is computed from the same variable or property it writes, where that value was read
// before an await or a yield that runs before the write.
//
//	invalid: count = count + await next();
//	invalid: count += await next();
//	invalid: const old = state.count; await save(); state.count = old + 1;
//	invalid: state.total = state.total + (await price());
//	valid:   state.lastRunAt = Date.now();                      not computed from its prior value
//	valid:   const original = console.log; ... finally { console.log = original; }   a restore
//	valid:   await save(); state.count = state.count + 1;       read after the suspension
//	valid:   async function f() { let total = 0; total += await next(); }   nothing else can write it
//
// # Where it came from
//
// Kirk's ruling of 2026-10-01, replacing `require-atomic-updates`, which is off in both engines.
// Upstream reports any write after an await to something read before it, and on ahra none of its 39
// findings was a read-modify-write: they were timestamps, flags, cached tokens and restores in
// `finally`. This rule asks the narrower question that names an actual lost update.
//
// # The conditions, all of which hold for a finding
//
//  1. The function can suspend: an async function or a generator. A suspension is an `await`, a
//     `yield`, or a `for await` loop's step, and only one in this function counts, never one inside a
//     nested function.
//  2. The write is a plain `=` or an arithmetic or bitwise compound assignment (`+=`, `-=`, `|=` and
//     the rest). `??=`, `||=` and `&&=` are check-then-act rather than read-modify-write and are
//     never judged. Destructuring assignments are not judged.
//     A write through a store's write call (`updateAccount(key, value)`, `cache.set(key, value)`) is
//     judged exactly as a plain `=` of its last argument to that store entry; storeAccessOf says what
//     makes a read call and a write call one store, and the read call is then the only way that
//     entry is read.
//  3. The target is a variable, or a static member path (`a.b`, `a['b']`, `this.c.d`) rooted at a
//     variable or at `this`. A computed key whose value the syntax does not settle (`a[k]`) is judged
//     only in a compound assignment, which reads and writes through one evaluated reference; under a
//     plain `=` two reads of `a[k]` are the same property only if `k` held still, which nothing here
//     can show.
//  4. Someone else can write the target during the suspension:
//     - a variable declared outside this function, or a local of it that a nested function writes;
//     - a member path rooted at `this`, at a variable declared outside this function, at one of this
//     function's own parameters (the caller holds the object), or at a local a nested function
//     references.
//     A local nobody else can see is never judged: `let total = 0; total += await next()` loses
//     nothing, because there is no one to lose it to.
//  5. The new value is computed from the target's prior value, read before a suspension that runs
//     before the write, on some path. The read is either in the assigned expression itself
//     (`x = x + await y`, with the compound form's implicit read counting as first), or reached
//     through a `const` of this function whose initializer read it (`const old = s.count; await f();
//     s.count = old + 1`), through any chain of such consts.
//  6. The value is MODIFIED rather than copied, and computed HERE. `console.log = original` writes back exactly the value
//     it saved, which is a restore, and restoring a saved handler in `finally` is the pattern the
//     upstream rule reported most on this tree. Any operation on the value (arithmetic, a call, a
//     member read, a spread into a new literal) makes it a modification. Two things are not:
//     - an awaited call's result, which is the callee's answer rather than this function's
//     computation (`token = (await refresh(token)).next` sends the old value out and writes back
//     what came back);
//     - a write after this function itself overwrote the target, between the read and the
//     write, which is the save-and-restore protocol (`const original = console.log; console.log =
//     filter; try { await run(); } finally { console.log = original.bind(console); }`).
//
// # How "before a suspension" is decided
//
// Inside the assigned expression, by evaluating it in order along each path: `&&`, `||`, `??` and
// `?:` fork, a branch's value carries only what that branch read, and a suspension marks stale every
// read already taken on its path. That is what keeps `x = c ? x + 1 : await f()` silent: the read and
// the suspension are on different paths. And `x = x || await f()` is silent too, because on the path
// that suspends the value written is `await f()`, not `x`.
//
// Across statements, through the function's control-flow graph: from the end of the const's
// declaration to the start of the write, is there a path through a suspension that does not pass the
// declaration again? Again is what keeps a const declared inside a loop honest, since each iteration
// binds a fresh one. A suspension whose expression contains the write (`await save(s.count = old + 1)`)
// runs after it and is not counted.
//
// # What it deliberately does not see
//
// A value laundered through a `let`, an alias of the object (`const snapshot = state; ...
// snapshot.count`), and a computed key under a plain `=` are all silent. So are a store read
// destructured straight into a pattern, a store whose key is a member read or a call, a store whose
// read and write functions are declared in different files, and a store object mutated in place and
// written back verbatim (`state.cursor = c; writeState(state)`). Each is a missed finding rather than
// a false one, which is the direction this rule errs in.
//
// No fix and no suggestion: the repair is a restructuring (re-read after the suspension, a lock, or
// an atomic update on the store), and which one is right is the author's call.
var ConcurrencyNoLostUpdate = rule.Rule{
	Name: "nexus/concurrency-no-lost-update",

	// Symbol identity is what makes a read and a write the same variable, and what tells an outer
	// variable from a shadowing local of the same spelling.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				var visit func(node *ast.Node) bool
				visit = func(node *ast.Node) bool {
					if control_flow_graph.IsRoot(node) && concurrencyNoLostUpdateCanSuspend(node) {
						analyzeConcurrencyNoLostUpdateRoot(ctx, node)
					}
					node.ForEachChild(visit)
					return false
				}
				sourceFile.ForEachChild(visit)
			},
		}
	},
}

// concurrencyNoLostUpdateCanSuspend reports whether a root is an async function or a generator.
func concurrencyNoLostUpdateCanSuspend(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
	default:
		return false
	}
	if node.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
		return true
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

// concurrencyNoLostUpdatePath is a variable, or a static member path rooted at a variable or `this`.
//
// With store set it is instead one entry of a store reached through a read call and a write call (see
// storeAccessOf): root, isThis and names then hold the receiver of a method store (`this.cache` for
// `this.cache.get(key)`) and are empty for a store of free functions, and store names the entry.
type concurrencyNoLostUpdatePath struct {
	root   *ast.Symbol
	isThis bool
	names  []string
	store  string
}

func (path concurrencyNoLostUpdatePath) key() string {
	prefix := "this"
	if !path.isThis {
		prefix = "symbol"
	}
	return prefix + "\x00" + strings.Join(path.names, "\x00") + "\x00\x00" + path.store
}

// sameRoot reports whether two paths start at the same variable, or both at `this`.
func (path concurrencyNoLostUpdatePath) sameRoot(other concurrencyNoLostUpdatePath) bool {
	if path.isThis || other.isThis {
		return path.isThis && other.isThis
	}
	return path.root == other.root
}

// sameStoreEntry reports whether two store targets name the same entry of the same store.
func (path concurrencyNoLostUpdatePath) sameStoreEntry(other concurrencyNoLostUpdatePath) bool {
	if path.store == "" || path.store != other.store || !path.sameRoot(other) || len(path.names) != len(other.names) {
		return false
	}
	for index, name := range path.names {
		if other.names[index] != name {
			return false
		}
	}
	return true
}

// concurrencyNoLostUpdateDependencies records what an evaluated value carries of the target's prior
// value: whether it was read before a suspension yet (stale) or not (fresh), and whether it is the
// value itself (verbatim) or something computed from it (modified).
type concurrencyNoLostUpdateDependencies uint8

const (
	concurrencyNoLostUpdateFreshVerbatim concurrencyNoLostUpdateDependencies = 1 << iota
	concurrencyNoLostUpdateFreshModified
	concurrencyNoLostUpdateStaleVerbatim
	concurrencyNoLostUpdateStaleModified
)

// staled is what a suspension does to a value read before it.
func (dependencies concurrencyNoLostUpdateDependencies) staled() concurrencyNoLostUpdateDependencies {
	result := dependencies &^ (concurrencyNoLostUpdateFreshVerbatim | concurrencyNoLostUpdateFreshModified)
	if dependencies&concurrencyNoLostUpdateFreshVerbatim != 0 {
		result |= concurrencyNoLostUpdateStaleVerbatim
	}
	if dependencies&concurrencyNoLostUpdateFreshModified != 0 {
		result |= concurrencyNoLostUpdateStaleModified
	}
	return result
}

// modified is what any operation on a value does to it.
func (dependencies concurrencyNoLostUpdateDependencies) modified() concurrencyNoLostUpdateDependencies {
	result := dependencies &^ (concurrencyNoLostUpdateFreshVerbatim | concurrencyNoLostUpdateStaleVerbatim)
	if dependencies&concurrencyNoLostUpdateFreshVerbatim != 0 {
		result |= concurrencyNoLostUpdateFreshModified
	}
	if dependencies&concurrencyNoLostUpdateStaleVerbatim != 0 {
		result |= concurrencyNoLostUpdateStaleModified
	}
	return result
}

// concurrencyNoLostUpdateOutcome is one evaluation path through an expression.
type concurrencyNoLostUpdateOutcome struct {
	dependencies concurrencyNoLostUpdateDependencies
	// suspended is whether a suspension has happened on this path, counting one before the expression.
	suspended bool
	// during is whether a suspension happened inside this expression, which is what stales the reads
	// its earlier siblings took.
	during bool
}

// concurrencyNoLostUpdateLocal is a const of the root that may carry the target's prior value.
type concurrencyNoLostUpdateLocal struct {
	name        *ast.Node
	initializer *ast.Node
	// elementPath is set for a binding element of an object pattern over a static path, which reads
	// that path plus the element's property at the declaration.
	elementPath *concurrencyNoLostUpdatePath
}

type concurrencyNoLostUpdateEventKind uint8

const (
	concurrencyNoLostUpdateSuspend concurrencyNoLostUpdateEventKind = iota
	concurrencyNoLostUpdateDeclarationStart
	concurrencyNoLostUpdateDeclarationEnd
	concurrencyNoLostUpdateAnchor
)

type concurrencyNoLostUpdateEvent struct {
	kind   concurrencyNoLostUpdateEventKind
	node   *ast.Node
	symbol *ast.Symbol
}

// concurrencyNoLostUpdateLocalKey memoizes what one const carries of one target.
type concurrencyNoLostUpdateLocalKey struct {
	local      *ast.Symbol
	targetRoot *ast.Symbol
	targetPath string
}

type concurrencyNoLostUpdateNestedUsage struct {
	referenced bool
	written    bool
}

// concurrencyNoLostUpdateRoot is the analysis of one suspending function.
type concurrencyNoLostUpdateRoot struct {
	ctx    rule.Context
	root   *ast.Node
	locals map[*ast.Symbol]*concurrencyNoLostUpdateLocal

	graph            *control_flow_graph.Graph[concurrencyNoLostUpdateEvent]
	between          map[concurrencyNoLostUpdateBetweenKey]bool
	writtenPaths     map[*ast.Node]*concurrencyNoLostUpdatePath
	localMemo        map[concurrencyNoLostUpdateLocalKey]concurrencyNoLostUpdateDependencies
	localInProgress  map[concurrencyNoLostUpdateLocalKey]bool
	nestedUsageCache map[*ast.Symbol]concurrencyNoLostUpdateNestedUsage
	storeAccesses    map[*ast.Node]*concurrencyNoLostUpdateStoreAccess
	stableBindings   map[*ast.Symbol]bool
}

func analyzeConcurrencyNoLostUpdateRoot(ctx rule.Context, root *ast.Node) {
	analysis := &concurrencyNoLostUpdateRoot{
		ctx:              ctx,
		root:             root,
		locals:           map[*ast.Symbol]*concurrencyNoLostUpdateLocal{},
		between:          map[concurrencyNoLostUpdateBetweenKey]bool{},
		writtenPaths:     map[*ast.Node]*concurrencyNoLostUpdatePath{},
		localMemo:        map[concurrencyNoLostUpdateLocalKey]concurrencyNoLostUpdateDependencies{},
		localInProgress:  map[concurrencyNoLostUpdateLocalKey]bool{},
		nestedUsageCache: map[*ast.Symbol]concurrencyNoLostUpdateNestedUsage{},
		storeAccesses:    map[*ast.Node]*concurrencyNoLostUpdateStoreAccess{},
		stableBindings:   map[*ast.Symbol]bool{},
	}

	var assignments []*ast.Node
	var calls []*ast.Node
	suspends := false
	analysis.forEachOwnNode(func(node *ast.Node) {
		switch node.Kind {
		case ast.KindAwaitExpression, ast.KindYieldExpression:
			suspends = true
		case ast.KindForOfStatement:
			if node.AsForInOrOfStatement().AwaitModifier != nil {
				suspends = true
			}
		case ast.KindBinaryExpression:
			operator := node.AsBinaryExpression().OperatorToken.Kind
			if ast.IsAssignmentOperator(operator) && !ast.IsLogicalOrCoalescingAssignmentOperator(operator) {
				assignments = append(assignments, node)
			}
		case ast.KindCallExpression:
			calls = append(calls, node)
		case ast.KindVariableDeclaration:
			analysis.collectLocal(node)
		}
	})
	// A function that never suspends cannot hold a stale read, so it costs nothing beyond this walk.
	//
	// A cost filter rather than a judgment, and a mutant removing it survived for that reason: with no
	// suspension, no assigned expression can suspend and no path through the graph carries one.
	if !suspends {
		return
	}

	for _, assignment := range assignments {
		binary := assignment.AsBinaryExpression()
		target, isPath := analysis.pathOf(binary.Left)
		if !isPath {
			// Reached for a plain `=` too, and harmless there: no read in an assigned expression can
			// match a computed name, so such a write can never be shown computed from its target.
			target, isPath = analysis.compoundTargetOf(binary.Left)
		}
		if !isPath || !analysis.targetIsShared(target) {
			continue
		}
		if analysis.losesUpdate(assignment, target) {
			targetRange := rule.TokenRange(ctx.SourceFile, binary.Left)
			targetText := ctx.SourceFile.Text()[targetRange.Pos():targetRange.End()]
			ctx.ReportNode(assignment, concurrencyNoLostUpdateMessage(targetText))
		}
	}

	// Writes through a store's write call, judged exactly as a plain `=` is: the written value must be
	// computed from a stale read of the same entry, through that store's read call.
	for _, call := range calls {
		access := analysis.storeAccessOf(call)
		if access == nil || !access.isWrite || !analysis.targetIsShared(access.target) {
			continue
		}
		evaluator := concurrencyNoLostUpdateEvaluator{analysis: analysis, target: access.target, anchor: call}
		for _, outcome := range evaluator.evaluate(access.value, false) {
			if outcome.dependencies&concurrencyNoLostUpdateStaleModified != 0 {
				ctx.ReportNode(call, concurrencyNoLostUpdateStoreMessage(analysis.storeReadText(access.target, call)))
				break
			}
		}
	}
}

// forEachOwnNode visits every node that runs in this root, never entering a nested one.
func (analysis *concurrencyNoLostUpdateRoot) forEachOwnNode(visitor func(node *ast.Node)) {
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if control_flow_graph.IsRoot(node) {
			return false
		}
		visitor(node)
		node.ForEachChild(visit)
		return false
	}
	analysis.root.ForEachChild(visit)
}

// collectLocal records a const whose value may carry the target's prior value: one named by an
// identifier with an initializer, or each property of an object pattern destructuring a static path.
func (analysis *concurrencyNoLostUpdateRoot) collectLocal(declaration *ast.Node) {
	if !ast.IsVarConst(declaration) {
		return
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Initializer == nil {
		return
	}
	name := variable.Name()
	if name == nil {
		return
	}
	if name.Kind == ast.KindIdentifier {
		if symbol := analysis.ctx.TypeChecker.GetSymbolAtLocation(name); symbol != nil {
			analysis.locals[symbol] = &concurrencyNoLostUpdateLocal{name: name, initializer: variable.Initializer}
		}
		return
	}
	if name.Kind != ast.KindObjectBindingPattern {
		return
	}
	source, isPath := analysis.pathOf(variable.Initializer)
	if !isPath {
		return
	}
	pattern := name.AsBindingPattern()
	if pattern.Elements == nil {
		return
	}
	for _, element := range pattern.Elements.Nodes {
		binding := element.AsBindingElement()
		if binding == nil || binding.DotDotDotToken != nil {
			continue
		}
		elementName := binding.Name()
		if elementName == nil || elementName.Kind != ast.KindIdentifier {
			continue
		}
		propertyName := elementName.Text()
		if binding.PropertyName != nil {
			text, isStatic := property.Name(binding.PropertyName,
				property.Named|property.Quoted|property.Numeric|property.Computed)
			if !isStatic {
				continue
			}
			propertyName = text
		}
		symbol := analysis.ctx.TypeChecker.GetSymbolAtLocation(elementName)
		if symbol == nil {
			continue
		}
		elementPath := concurrencyNoLostUpdatePath{
			root:   source.root,
			isThis: source.isThis,
			names:  append(append([]string{}, source.names...), propertyName),
		}
		analysis.locals[symbol] = &concurrencyNoLostUpdateLocal{name: elementName, elementPath: &elementPath}
	}
}

// resolve returns the symbol an identifier occurrence reads, seeing through a shorthand property.
func (analysis *concurrencyNoLostUpdateRoot) resolve(identifier *ast.Node) *ast.Symbol {
	if parent := identifier.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
		if symbol := analysis.ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent); symbol != nil {
			return symbol
		}
	}
	return analysis.ctx.TypeChecker.GetSymbolAtLocation(identifier)
}

// pathOf reads an expression as a variable or a static member path, through parentheses and type
// assertions.
func (analysis *concurrencyNoLostUpdateRoot) pathOf(node *ast.Node) (concurrencyNoLostUpdatePath, bool) {
	node = concurrencyNoLostUpdateSkipTransparent(node)
	if node == nil {
		return concurrencyNoLostUpdatePath{}, false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		symbol := analysis.resolve(node)
		if symbol == nil {
			return concurrencyNoLostUpdatePath{}, false
		}
		return concurrencyNoLostUpdatePath{root: symbol}, true
	case ast.KindThisKeyword:
		return concurrencyNoLostUpdatePath{isThis: true}, true
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		name, isStatic := property.AccessedName(node,
			property.Named|property.Quoted|property.Templated|property.Numeric|property.Private)
		if !isStatic {
			return concurrencyNoLostUpdatePath{}, false
		}
		var object *ast.Node
		if node.Kind == ast.KindPropertyAccessExpression {
			object = node.AsPropertyAccessExpression().Expression
		} else {
			object = node.AsElementAccessExpression().Expression
		}
		objectPath, isPath := analysis.pathOf(object)
		if !isPath {
			return concurrencyNoLostUpdatePath{}, false
		}
		objectPath.names = append(append([]string{}, objectPath.names...), name)
		return objectPath, true
	}
	return concurrencyNoLostUpdatePath{}, false
}

// concurrencyNoLostUpdateComputedName stands for a member whose key the syntax does not settle. It
// can never equal a real property name, so no read in an assigned expression matches it.
const concurrencyNoLostUpdateComputedName = "\x00computed"

// compoundTargetOf reads a compound assignment's member target whose key is computed (`a[k].b += 1`)
// as its root plus one name that matches nothing.
//
// A compound assignment evaluates its target reference once and both reads and writes through it, so
// a computed key cannot name two different properties between the read and the write. That is what
// makes `state.byKey[key] += await next()` judgeable when `state.byKey[key] = state.byKey[key] +
// await next()` is not: there the two `key` reads are separate, and nothing here can show `key` held
// still. Only the root matters, for condition 4.
func (analysis *concurrencyNoLostUpdateRoot) compoundTargetOf(node *ast.Node) (concurrencyNoLostUpdatePath, bool) {
	current := concurrencyNoLostUpdateSkipTransparent(node)
	if current == nil || (current.Kind != ast.KindPropertyAccessExpression && current.Kind != ast.KindElementAccessExpression) {
		return concurrencyNoLostUpdatePath{}, false
	}
	for current.Kind == ast.KindPropertyAccessExpression || current.Kind == ast.KindElementAccessExpression {
		if current.Kind == ast.KindPropertyAccessExpression {
			current = concurrencyNoLostUpdateSkipTransparent(current.AsPropertyAccessExpression().Expression)
		} else {
			current = concurrencyNoLostUpdateSkipTransparent(current.AsElementAccessExpression().Expression)
		}
		if current == nil {
			return concurrencyNoLostUpdatePath{}, false
		}
	}
	root, isPath := analysis.pathOf(current)
	if !isPath {
		return concurrencyNoLostUpdatePath{}, false
	}
	root.names = []string{concurrencyNoLostUpdateComputedName}
	return root, true
}

// concurrencyNoLostUpdateSkipTransparent sees through the wrappers that change neither a value nor
// which reference it is.
func concurrencyNoLostUpdateSkipTransparent(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		default:
			return node
		}
	}
	return nil
}

// targetIsShared is condition 4: someone other than this line can write the target while the
// function is suspended.
func (analysis *concurrencyNoLostUpdateRoot) targetIsShared(target concurrencyNoLostUpdatePath) bool {
	if target.isThis {
		return true
	}
	if target.store != "" && target.root == nil {
		// A store of free functions lives outside this call by construction.
		return true
	}
	declaration := target.root.ValueDeclaration
	if declaration == nil && len(target.root.Declarations) > 0 {
		declaration = target.root.Declarations[0]
	}
	if declaration == nil {
		return false
	}
	if !concurrencyNoLostUpdateWithin(declaration, analysis.root) {
		return true
	}
	usage := analysis.nestedUsage(target.root)
	// A method store's entry lives in the object its receiver holds, so it is judged as a member is.
	if len(target.names) == 0 && target.store == "" {
		// The binding itself. A parameter or a local is this call's own, so only a nested function
		// writing it can race this one.
		return usage.written
	}
	if declaration.Kind == ast.KindParameter && declaration.Parent == analysis.root {
		return true
	}
	return usage.referenced
}

func concurrencyNoLostUpdateWithin(node *ast.Node, ancestor *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

// nestedUsage reports whether a function nested in this root references or writes a symbol.
func (analysis *concurrencyNoLostUpdateRoot) nestedUsage(symbol *ast.Symbol) concurrencyNoLostUpdateNestedUsage {
	if usage, cached := analysis.nestedUsageCache[symbol]; cached {
		return usage
	}
	usage := concurrencyNoLostUpdateNestedUsage{}
	var visit func(node *ast.Node, nested bool) bool
	visit = func(node *ast.Node, nested bool) bool {
		if control_flow_graph.IsRoot(node) {
			nested = true
		}
		if nested && node.Kind == ast.KindIdentifier && analysis.resolve(node) == symbol {
			usage.referenced = true
			if reference.WritesToBinding(node) {
				usage.written = true
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			return visit(child, nested)
		})
		return false
	}
	analysis.root.ForEachChild(func(child *ast.Node) bool {
		return visit(child, false)
	})
	analysis.nestedUsageCache[symbol] = usage
	return usage
}

// losesUpdate is condition 5 and 6 for one candidate write.
func (analysis *concurrencyNoLostUpdateRoot) losesUpdate(assignment *ast.Node, target concurrencyNoLostUpdatePath) bool {
	binary := assignment.AsBinaryExpression()
	evaluator := concurrencyNoLostUpdateEvaluator{analysis: analysis, target: target, anchor: assignment}
	outcomes := evaluator.evaluate(binary.Right, false)
	if binary.OperatorToken.Kind == ast.KindEqualsToken {
		for _, outcome := range outcomes {
			if outcome.dependencies&concurrencyNoLostUpdateStaleModified != 0 {
				return true
			}
		}
		return false
	}
	// A compound assignment reads its target before the right-hand side runs, and always computes.
	for _, outcome := range outcomes {
		if outcome.during {
			return true
		}
		if outcome.dependencies&concurrencyNoLostUpdateStaleModified != 0 ||
			outcome.dependencies&concurrencyNoLostUpdateStaleVerbatim != 0 {
			return true
		}
	}
	return false
}

// concurrencyNoLostUpdateEvaluator walks one expression in evaluation order for one target.
//
// anchor is where the expression starts in the control-flow graph: the write being judged, or the
// declaration of the const whose initializer is being judged. A const read inside the expression is
// stale when a suspension lies between the const's declaration and the anchor.
type concurrencyNoLostUpdateEvaluator struct {
	analysis *concurrencyNoLostUpdateRoot
	target   concurrencyNoLostUpdatePath
	anchor   *ast.Node
}

func concurrencyNoLostUpdateDeduplicate(outcomes []concurrencyNoLostUpdateOutcome) []concurrencyNoLostUpdateOutcome {
	seen := map[concurrencyNoLostUpdateOutcome]bool{}
	unique := outcomes[:0:0]
	for _, outcome := range outcomes {
		if !seen[outcome] {
			seen[outcome] = true
			unique = append(unique, outcome)
		}
	}
	return unique
}

func (evaluator concurrencyNoLostUpdateEvaluator) leaf(
	dependencies concurrencyNoLostUpdateDependencies,
	suspended bool,
) []concurrencyNoLostUpdateOutcome {
	return []concurrencyNoLostUpdateOutcome{{dependencies: dependencies, suspended: suspended}}
}

// sequence evaluates nodes left to right and unions their values. A suspension in a later node
// stales what earlier nodes read.
func (evaluator concurrencyNoLostUpdateEvaluator) sequence(nodes []*ast.Node, suspended bool) []concurrencyNoLostUpdateOutcome {
	states := evaluator.leaf(0, suspended)
	for _, node := range nodes {
		if node == nil {
			continue
		}
		var next []concurrencyNoLostUpdateOutcome
		for _, state := range states {
			for _, result := range evaluator.evaluate(node, state.suspended) {
				dependencies := state.dependencies
				if result.during {
					dependencies = dependencies.staled()
				}
				next = append(next, concurrencyNoLostUpdateOutcome{
					dependencies: dependencies | result.dependencies,
					suspended:    result.suspended,
					during:       state.during || result.during,
				})
			}
		}
		states = concurrencyNoLostUpdateDeduplicate(next)
	}
	return states
}

func concurrencyNoLostUpdateModifiedAll(outcomes []concurrencyNoLostUpdateOutcome) []concurrencyNoLostUpdateOutcome {
	for index := range outcomes {
		outcomes[index].dependencies = outcomes[index].dependencies.modified()
	}
	return concurrencyNoLostUpdateDeduplicate(outcomes)
}

// then evaluates second after each outcome of first, keeping only second's value.
func (evaluator concurrencyNoLostUpdateEvaluator) then(
	first []concurrencyNoLostUpdateOutcome,
	second *ast.Node,
) []concurrencyNoLostUpdateOutcome {
	var outcomes []concurrencyNoLostUpdateOutcome
	for _, state := range first {
		for _, result := range evaluator.evaluate(second, state.suspended) {
			result.during = result.during || state.during
			outcomes = append(outcomes, result)
		}
	}
	return outcomes
}

// evaluate returns every path through node, given whether a suspension already happened on the path.
func (evaluator concurrencyNoLostUpdateEvaluator) evaluate(node *ast.Node, suspended bool) []concurrencyNoLostUpdateOutcome {
	if node == nil {
		return evaluator.leaf(0, suspended)
	}
	if control_flow_graph.IsRoot(node) || node.Kind == ast.KindClassExpression {
		// A nested function's body runs later, if at all; its value carries nothing read now.
		return evaluator.leaf(0, suspended)
	}
	if ast.IsTypeNode(node) {
		return evaluator.leaf(0, suspended)
	}

	switch node.Kind {
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression,
		ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
		return evaluator.evaluate(concurrencyNoLostUpdateSkipTransparent(node), suspended)

	case ast.KindIdentifier:
		return evaluator.leaf(evaluator.identifier(node, suspended), suspended)

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if path, isPath := evaluator.analysis.pathOf(node); isPath {
			if matched := evaluator.match(path); matched != 0 {
				return evaluator.leaf(matched, suspended)
			}
			// A static path that is not the target can still be a member read of a const carrying it.
			// Nothing in a static path can suspend, so only its root identifier needs evaluating.
			root := node
			for {
				root = concurrencyNoLostUpdateSkipTransparent(root)
				if root.Kind == ast.KindPropertyAccessExpression {
					root = root.AsPropertyAccessExpression().Expression
					continue
				}
				if root.Kind == ast.KindElementAccessExpression {
					root = root.AsElementAccessExpression().Expression
					continue
				}
				break
			}
			if root.Kind != ast.KindIdentifier {
				return evaluator.leaf(0, suspended)
			}
			return evaluator.leaf(evaluator.identifier(root, suspended).modified(), suspended)
		}
		if node.Kind == ast.KindPropertyAccessExpression {
			return concurrencyNoLostUpdateModifiedAll(
				evaluator.evaluate(node.AsPropertyAccessExpression().Expression, suspended))
		}
		access := node.AsElementAccessExpression()
		return concurrencyNoLostUpdateModifiedAll(
			evaluator.sequence([]*ast.Node{access.Expression, access.ArgumentExpression}, suspended))

	case ast.KindAwaitExpression:
		operand := node.AsAwaitExpression().Expression
		outcomes := evaluator.evaluate(operand, suspended)
		// An awaited call's value is the callee's answer, produced across the suspension, and not a
		// computation made here from what this function handed it. Measured on ahra: the OAuth refresh
		// in `modules/spotify/SpotifyApi.ts:91` and `modules/finance/connections/QuickBooksClient.ts:156`
		// sends the old refresh token in a `fetch` body and writes back the rotated one from the
		// response, and carrying the argument through the call reported both as computed from the old
		// token. Awaiting a value that is not a call (`await count`) still carries it.
		producedByCallee := false
		// An awaited read of the target store is the exception: the callee's answer IS the entry. It
		// is read as the promise settles, so this await does not stale it; a later one does.
		readOfTarget := false
		switch skipped := concurrencyNoLostUpdateSkipTransparent(operand); {
		case skipped == nil:
		case evaluator.isTargetStoreRead(skipped):
			readOfTarget = true
		case skipped.Kind == ast.KindCallExpression, skipped.Kind == ast.KindNewExpression,
			skipped.Kind == ast.KindTaggedTemplateExpression:
			producedByCallee = true
		}
		for index := range outcomes {
			if producedByCallee {
				outcomes[index].dependencies = 0
			}
			if !readOfTarget {
				outcomes[index].dependencies = outcomes[index].dependencies.staled()
			}
			outcomes[index].suspended = true
			outcomes[index].during = true
		}
		return concurrencyNoLostUpdateDeduplicate(outcomes)

	case ast.KindYieldExpression:
		// The value of a yield is whatever the caller sends back, so nothing read here survives it.
		outcomes := evaluator.evaluate(node.AsYieldExpression().Expression, suspended)
		for index := range outcomes {
			outcomes[index].dependencies = 0
			outcomes[index].suspended = true
			outcomes[index].during = true
		}
		return concurrencyNoLostUpdateDeduplicate(outcomes)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		switch {
		case operator == ast.KindAmpersandAmpersandToken || operator == ast.KindBarBarToken ||
			operator == ast.KindQuestionQuestionToken:
			left := evaluator.evaluate(binary.Left, suspended)
			outcomes := append([]concurrencyNoLostUpdateOutcome{}, left...)
			outcomes = append(outcomes, evaluator.then(left, binary.Right)...)
			return concurrencyNoLostUpdateDeduplicate(outcomes)
		case operator == ast.KindCommaToken:
			return concurrencyNoLostUpdateDeduplicate(
				evaluator.then(evaluator.evaluate(binary.Left, suspended), binary.Right))
		case operator == ast.KindEqualsToken:
			// A nested assignment's value is its right-hand side.
			return evaluator.evaluate(binary.Right, suspended)
		}
		return concurrencyNoLostUpdateModifiedAll(evaluator.sequence([]*ast.Node{binary.Left, binary.Right}, suspended))

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		condition := evaluator.evaluate(conditional.Condition, suspended)
		outcomes := evaluator.then(condition, conditional.WhenTrue)
		outcomes = append(outcomes, evaluator.then(condition, conditional.WhenFalse)...)
		return concurrencyNoLostUpdateDeduplicate(outcomes)

	case ast.KindObjectLiteralExpression:
		var parts []*ast.Node
		for _, member := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch member.Kind {
			case ast.KindPropertyAssignment:
				assignment := member.AsPropertyAssignment()
				if name := assignment.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					parts = append(parts, name.AsComputedPropertyName().Expression)
				}
				parts = append(parts, assignment.Initializer)
			case ast.KindShorthandPropertyAssignment:
				parts = append(parts, member.AsShorthandPropertyAssignment().Name())
			case ast.KindSpreadAssignment:
				parts = append(parts, member.AsSpreadAssignment().Expression)
			}
		}
		return concurrencyNoLostUpdateModifiedAll(evaluator.sequence(parts, suspended))

	case ast.KindCallExpression:
		if evaluator.isTargetStoreRead(node) {
			// The arguments run first and carry nothing of the entry; the call's answer is the entry.
			var children []*ast.Node
			node.ForEachChild(func(child *ast.Node) bool {
				children = append(children, child)
				return false
			})
			outcomes := evaluator.sequence(children, suspended)
			for index := range outcomes {
				outcomes[index].dependencies = concurrencyNoLostUpdateFreshVerbatim
			}
			return concurrencyNoLostUpdateDeduplicate(outcomes)
		}
	}

	var children []*ast.Node
	node.ForEachChild(func(child *ast.Node) bool {
		children = append(children, child)
		return false
	})
	return concurrencyNoLostUpdateModifiedAll(evaluator.sequence(children, suspended))
}

// match compares a static path read against the target: the target itself is a verbatim read, a
// longer path through it is a read of the target that something is then computed from.
func (evaluator concurrencyNoLostUpdateEvaluator) match(path concurrencyNoLostUpdatePath) concurrencyNoLostUpdateDependencies {
	// A store entry is read only through its read call, never through a path.
	if evaluator.target.store != "" {
		return 0
	}
	if !path.sameRoot(evaluator.target) || len(path.names) < len(evaluator.target.names) {
		return 0
	}
	for index, name := range evaluator.target.names {
		if path.names[index] != name {
			return 0
		}
	}
	if len(path.names) == len(evaluator.target.names) {
		return concurrencyNoLostUpdateFreshVerbatim
	}
	return concurrencyNoLostUpdateFreshModified
}

// identifier evaluates one identifier read: the target variable itself, or a const carrying it.
func (evaluator concurrencyNoLostUpdateEvaluator) identifier(node *ast.Node, suspended bool) concurrencyNoLostUpdateDependencies {
	symbol := evaluator.analysis.resolve(node)
	if symbol == nil {
		return 0
	}
	if !evaluator.target.isThis && evaluator.target.store == "" && evaluator.target.root == symbol &&
		len(evaluator.target.names) == 0 {
		return concurrencyNoLostUpdateFreshVerbatim
	}
	local := evaluator.analysis.locals[symbol]
	if local == nil || local.name == node {
		return 0
	}
	dependencies := evaluator.analysis.localDependencies(symbol, local, evaluator.target)
	if dependencies == 0 {
		return 0
	}
	if suspended || evaluator.analysis.suspensionBetween(symbol, evaluator.anchor, evaluator.target) {
		dependencies = dependencies.staled()
	}
	return dependencies
}

// localDependencies is what a const carries of the target at the end of its declaration.
func (analysis *concurrencyNoLostUpdateRoot) localDependencies(
	symbol *ast.Symbol,
	local *concurrencyNoLostUpdateLocal,
	target concurrencyNoLostUpdatePath,
) concurrencyNoLostUpdateDependencies {
	evaluator := concurrencyNoLostUpdateEvaluator{analysis: analysis, target: target, anchor: local.name}
	if local.elementPath != nil {
		return evaluator.match(*local.elementPath)
	}
	key := concurrencyNoLostUpdateLocalKey{local: symbol, targetRoot: target.root, targetPath: target.key()}
	if dependencies, memoized := analysis.localMemo[key]; memoized {
		return dependencies
	}
	if analysis.localInProgress[key] {
		return 0
	}
	analysis.localInProgress[key] = true
	var dependencies concurrencyNoLostUpdateDependencies
	for _, outcome := range evaluator.evaluate(local.initializer, false) {
		dependencies |= outcome.dependencies
	}
	delete(analysis.localInProgress, key)
	analysis.localMemo[key] = dependencies
	return dependencies
}

// concurrencyNoLostUpdateBetweenKey memoizes one control-flow question.
type concurrencyNoLostUpdateBetweenKey struct {
	symbol     *ast.Symbol
	anchor     *ast.Node
	targetRoot *ast.Symbol
	targetPath string
}

// suspensionBetween reports whether some path runs from the end of a const's declaration through a
// suspension to anchor without declaring the const again and without this function writing the
// target in between.
//
// That last clause is the save-and-restore protocol, and it is why a restore is not an update even
// when what it writes back is derived from the saved value rather than identical to it. Measured on
// ahra: `modules/phi/social/PhiSocialTerminal.ts` and `modules/art/ArtTerminal.ts` save
// `process.stdout.write.bind(process.stdout)`, replace `process.stdout.write` with a filter, await the
// callback, and put the bound original back in `finally`. The `.bind` call made the saved value a
// "modification" and the restore reported. What makes it a restore is the replacement between: the
// function itself overwrote the target after reading it, so the later write puts back the state from
// before this function's own change rather than building on a value someone else may have changed.
func (analysis *concurrencyNoLostUpdateRoot) suspensionBetween(
	symbol *ast.Symbol,
	anchor *ast.Node,
	target concurrencyNoLostUpdatePath,
) bool {
	memoKey := concurrencyNoLostUpdateBetweenKey{symbol: symbol, anchor: anchor, targetRoot: target.root, targetPath: target.key()}
	if answer, memoized := analysis.between[memoKey]; memoized {
		return answer
	}
	graph := analysis.controlFlowGraph()

	type position struct {
		block     *control_flow_graph.Block[concurrencyNoLostUpdateEvent]
		start     int
		suspended bool
	}
	var queue []position
	visited := map[[2]int]bool{}
	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		for index, event := range block.Events {
			if event.kind == concurrencyNoLostUpdateDeclarationEnd && event.symbol == symbol {
				queue = append(queue, position{block: block, start: index + 1})
			}
		}
	}

	answer := false
search:
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		suspended := current.suspended
		stopped := false
		for _, event := range current.block.Events[current.start:] {
			switch event.kind {
			case concurrencyNoLostUpdateSuspend:
				// A suspension whose expression contains the anchor runs after it.
				if event.node == nil || !concurrencyNoLostUpdateContains(event.node, anchor) {
					suspended = true
				}
			case concurrencyNoLostUpdateDeclarationEnd:
				// Declaring the const again binds a fresh one, so this path says nothing about the
				// value the anchor reads. Stopping at the declaration's start instead would be
				// equivalent, and a mutant doing so survived: nothing inside a const's own initializer
				// can read it, so no anchor lies between the two.
				if event.symbol == symbol {
					stopped = true
				}
			case concurrencyNoLostUpdateDeclarationStart:
				// A const anchor is its declaration's START. A suspension inside its own initializer
				// is the evaluator's to judge, path by path, because the initializer can fork.
				if event.node == anchor && suspended {
					answer = true
					break search
				}
			case concurrencyNoLostUpdateAnchor:
				if event.node == anchor {
					if suspended {
						answer = true
						break search
					}
				} else if analysis.writesTarget(event.node, target) {
					stopped = true
				}
			}
			if stopped {
				break
			}
		}
		if stopped {
			continue
		}
		for _, successor := range current.block.Successors {
			if successor == nil || !successor.Reachable {
				continue
			}
			key := [2]int{successor.Index(), 0}
			if suspended {
				key[1] = 1
			}
			if visited[key] {
				continue
			}
			visited[key] = true
			queue = append(queue, position{block: successor, suspended: suspended})
		}
	}
	analysis.between[memoKey] = answer
	return answer
}

// writesTarget reports whether an assignment writes exactly the target path, or a store's write call
// writes exactly the target entry.
func (analysis *concurrencyNoLostUpdateRoot) writesTarget(assignment *ast.Node, target concurrencyNoLostUpdatePath) bool {
	if assignment.Kind == ast.KindCallExpression {
		access := analysis.storeAccessOf(assignment)
		return access != nil && access.isWrite && access.target.sameStoreEntry(target)
	}
	if target.store != "" {
		return false
	}
	written, cached := analysis.writtenPaths[assignment]
	if !cached {
		if path, isPath := analysis.pathOf(assignment.AsBinaryExpression().Left); isPath {
			written = &path
		}
		analysis.writtenPaths[assignment] = written
	}
	if written == nil || !written.sameRoot(target) || len(written.names) != len(target.names) {
		return false
	}
	for index, name := range target.names {
		if written.names[index] != name {
			return false
		}
	}
	return true
}

func concurrencyNoLostUpdateContains(outer *ast.Node, inner *ast.Node) bool {
	return inner.Pos() >= outer.Pos() && inner.End() <= outer.End()
}

// controlFlowGraph builds the root's graph on first need, recording suspensions, the declarations of
// tracked consts, and every assignment as an anchor.
func (analysis *concurrencyNoLostUpdateRoot) controlFlowGraph() *control_flow_graph.Graph[concurrencyNoLostUpdateEvent] {
	if analysis.graph != nil {
		return analysis.graph
	}
	type builder = control_flow_graph.Builder[concurrencyNoLostUpdateEvent]
	declarationSymbol := func(node *ast.Node) *ast.Symbol {
		parent := node.Parent
		if parent == nil || node.Kind != ast.KindIdentifier {
			return nil
		}
		switch parent.Kind {
		case ast.KindVariableDeclaration, ast.KindBindingElement:
			if parent.Name() != node {
				return nil
			}
		default:
			return nil
		}
		symbol := analysis.ctx.TypeChecker.GetSymbolAtLocation(node)
		if local := analysis.locals[symbol]; local == nil || local.name != node {
			return nil
		}
		return symbol
	}
	isForAwait := func(node *ast.Node) bool {
		return node.Kind == ast.KindForOfStatement && node.AsForInOrOfStatement().AwaitModifier != nil
	}
	analysis.graph = control_flow_graph.Build(analysis.root, control_flow_graph.Hooks[concurrencyNoLostUpdateEvent]{
		Expression: func(b *builder, node *ast.Node) {
			switch node.Kind {
			case ast.KindAwaitExpression, ast.KindYieldExpression:
				b.Emit(concurrencyNoLostUpdateEvent{kind: concurrencyNoLostUpdateSuspend, node: node})
			case ast.KindBinaryExpression:
				if ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
					b.Emit(concurrencyNoLostUpdateEvent{kind: concurrencyNoLostUpdateAnchor, node: node})
				}
			case ast.KindCallExpression:
				if access := analysis.storeAccessOf(node); access != nil && access.isWrite {
					b.Emit(concurrencyNoLostUpdateEvent{kind: concurrencyNoLostUpdateAnchor, node: node})
				}
			}
		},
		Statement: func(b *builder, node *ast.Node) {
			// The first step of a `for await` suspends before the body runs. Recorded with no node, so
			// no anchor is exempted for being inside it.
			//
			// The step between iterations is not recorded, and a mutant recording it survived every
			// fixture for a reason rather than for want of one: the only const a path can carry across
			// the back edge is one declared before the loop, and every path from that declaration into
			// the body already crosses this first step. A const declared in the body is declared again
			// before anything in the next iteration can read it.
			if isForAwait(node) {
				b.Emit(concurrencyNoLostUpdateEvent{kind: concurrencyNoLostUpdateSuspend})
			}
		},
		Read: func(b *builder, node *ast.Node) {
			if symbol := declarationSymbol(node); symbol != nil {
				b.Emit(concurrencyNoLostUpdateEvent{kind: concurrencyNoLostUpdateDeclarationStart, node: node, symbol: symbol})
			}
		},
		Write: func(b *builder, node *ast.Node) {
			if symbol := declarationSymbol(node); symbol != nil {
				b.Emit(concurrencyNoLostUpdateEvent{kind: concurrencyNoLostUpdateDeclarationEnd, node: node, symbol: symbol})
			}
		},
	})
	return analysis.graph
}

// concurrencyNoLostUpdateStoreAccess is one call into a store: a read of an entry, or a write of a
// new value to it.
type concurrencyNoLostUpdateStoreAccess struct {
	target  concurrencyNoLostUpdatePath
	isWrite bool
	// value is the written value, a write call's last argument.
	value *ast.Node
}

var (
	concurrencyNoLostUpdateReadVerbs  = []string{"get", "read", "load"}
	concurrencyNoLostUpdateWriteVerbs = []string{"set", "update", "write", "save", "put", "store"}
)

// concurrencyNoLostUpdateSplitVerb splits `getAccountCredentials` into its verb and its noun,
// `AccountCredentials`, when it starts with one of verbs at a word boundary.
func concurrencyNoLostUpdateSplitVerb(name string, verbs []string) (string, bool) {
	for _, verb := range verbs {
		if !strings.HasPrefix(name, verb) {
			continue
		}
		noun := name[len(verb):]
		if noun == "" || (noun[0] >= 'A' && noun[0] <= 'Z') {
			return noun, true
		}
	}
	return "", false
}

// storeAccessOf reads a call as a read or a write of one store entry, or returns nil.
//
// A store is a pair of calls the syntax ties together and nothing else: the same noun after a read verb
// and a write verb (`getAccountCredentials` and `updateAccountCredentials`, `loadSettings` and
// `saveSettings`, `cache.get` and `cache.set`), the same entry named by the same leading arguments, and
// the written value as the write's one extra, last argument. Three things keep the pair from being a
// coincidence of names:
//
//   - free functions must both resolve, through imports, to declarations in the same file, and their
//     noun must not be empty (a bare `get` and `set` from one module say nothing about a shared entry);
//   - methods must be called on the same static receiver path, the object that holds the entry;
//   - every key argument is a literal or a binding that holds still for the whole function (a const,
//     an import, or a parameter or local nothing in the function writes), so two calls naming `key`
//     name the same entry. Anything else (a member read, a call, a spread) is not judged.
//
// This is the Reddit token refresh of 2026-10-01: `modules/reddit/RedditClient.ts` reads the account
// through `getAccountCredentials(accountKey)`, awaits the token endpoint, and writes `{ oauth2: {
// ...account.oauth2, ...newTokens } }` back through `updateAccountCredentials(accountKey, ...)`, so a
// concurrent refresh of the other token is overwritten with the copy read before the await. No `=`
// is involved, which is why the assignment half of this rule could not see it.
func (analysis *concurrencyNoLostUpdateRoot) storeAccessOf(call *ast.Node) *concurrencyNoLostUpdateStoreAccess {
	if access, cached := analysis.storeAccesses[call]; cached {
		return access
	}
	access := analysis.computeStoreAccess(call)
	analysis.storeAccesses[call] = access
	return access
}

func (analysis *concurrencyNoLostUpdateRoot) computeStoreAccess(call *ast.Node) *concurrencyNoLostUpdateStoreAccess {
	if call.Kind != ast.KindCallExpression {
		return nil
	}
	expression := call.AsCallExpression()
	callee := concurrencyNoLostUpdateSkipTransparent(expression.Expression)
	if callee == nil {
		return nil
	}

	var name string
	var target concurrencyNoLostUpdatePath
	var owner string
	switch callee.Kind {
	case ast.KindIdentifier:
		name = callee.Text()
		symbol := analysis.resolve(callee)
		if symbol == nil {
			return nil
		}
		symbol = shimchecker.SkipAlias(symbol, analysis.ctx.TypeChecker)
		declaration := symbol.ValueDeclaration
		if declaration == nil && len(symbol.Declarations) > 0 {
			declaration = symbol.Declarations[0]
		}
		if declaration == nil {
			return nil
		}
		sourceFile := ast.GetSourceFileOfNode(declaration)
		if sourceFile == nil {
			return nil
		}
		owner = ("function\x00" + sourceFile.FileName()).AsString()
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		nameNode := access.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return nil
		}
		name = nameNode.Text()
		receiver, isPath := analysis.pathOf(access.Expression)
		if !isPath {
			return nil
		}
		target = receiver
		owner = "method"
	default:
		return nil
	}

	noun, isRead := concurrencyNoLostUpdateSplitVerb(name, concurrencyNoLostUpdateReadVerbs)
	isWrite := false
	if !isRead {
		noun, isWrite = concurrencyNoLostUpdateSplitVerb(name, concurrencyNoLostUpdateWriteVerbs)
		if !isWrite {
			return nil
		}
	}
	if noun == "" && owner != "method" {
		return nil
	}

	var arguments []*ast.Node
	if expression.Arguments != nil {
		arguments = expression.Arguments.Nodes
	}
	keyArguments := arguments
	var value *ast.Node
	if isWrite {
		if len(arguments) == 0 {
			return nil
		}
		keyArguments = arguments[:len(arguments)-1]
		value = arguments[len(arguments)-1]
		if value.Kind == ast.KindSpreadElement {
			return nil
		}
	}
	keys := make([]string, 0, len(keyArguments))
	for _, argument := range keyArguments {
		key, isKey := analysis.storeKeyOf(argument)
		if !isKey {
			return nil
		}
		keys = append(keys, key)
	}

	target.store = owner + "\x00" + noun + "\x00" + strings.Join(keys, "\x00")
	return &concurrencyNoLostUpdateStoreAccess{target: target, isWrite: isWrite, value: value}
}

// storeKeyOf canonicalizes one key argument: a literal by its text, a binding that holds still by its
// symbol.
func (analysis *concurrencyNoLostUpdateRoot) storeKeyOf(argument *ast.Node) (string, bool) {
	argument = concurrencyNoLostUpdateSkipTransparent(argument)
	if argument == nil {
		return "", false
	}
	switch {
	case ast.IsStringLiteralLike(argument):
		return "string\x00" + argument.Text(), true
	case ast.IsNumericLiteral(argument):
		return "number\x00" + argument.Text(), true
	case argument.Kind == ast.KindIdentifier:
		symbol := analysis.resolve(argument)
		if symbol == nil || !analysis.bindingHoldsStill(symbol) {
			return "", false
		}
		return fmt.Sprintf("binding\x00%p", symbol), true
	}
	return "", false
}

// bindingHoldsStill reports whether a binding names one value for the whole of this function: a
// const or an import, or a parameter or local of this function that nothing in it, nested functions
// included, writes.
func (analysis *concurrencyNoLostUpdateRoot) bindingHoldsStill(symbol *ast.Symbol) bool {
	if holds, cached := analysis.stableBindings[symbol]; cached {
		return holds
	}
	holds := false
	declaration := symbol.ValueDeclaration
	if declaration == nil && len(symbol.Declarations) > 0 {
		declaration = symbol.Declarations[0]
	}
	switch {
	case symbol.Flags&ast.SymbolFlagsAlias != 0:
		holds = true
	case declaration == nil:
	case declaration.Kind == ast.KindVariableDeclaration && ast.IsVarConst(declaration):
		holds = true
	case (declaration.Kind == ast.KindParameter || declaration.Kind == ast.KindVariableDeclaration) &&
		concurrencyNoLostUpdateWithin(declaration, analysis.root):
		holds = true
		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if !holds {
				return true
			}
			if node.Kind == ast.KindIdentifier && node != declaration.Name() && analysis.resolve(node) == symbol &&
				reference.WritesToBinding(node) {
				holds = false
				return true
			}
			node.ForEachChild(visit)
			return !holds
		}
		analysis.root.ForEachChild(visit)
	}
	analysis.stableBindings[symbol] = holds
	return holds
}

// isTargetStoreRead reports whether a node is a read call of exactly the entry being judged.
func (evaluator concurrencyNoLostUpdateEvaluator) isTargetStoreRead(node *ast.Node) bool {
	if evaluator.target.store == "" || node.Kind != ast.KindCallExpression {
		return false
	}
	access := evaluator.analysis.storeAccessOf(node)
	return access != nil && !access.isWrite && access.target.sameStoreEntry(evaluator.target)
}

// storeReadText is the source of the first read of the target entry before the write, for the
// message.
func (analysis *concurrencyNoLostUpdateRoot) storeReadText(target concurrencyNoLostUpdatePath, write *ast.Node) string {
	var read *ast.Node
	evaluator := concurrencyNoLostUpdateEvaluator{analysis: analysis, target: target}
	analysis.forEachOwnNode(func(node *ast.Node) {
		if read == nil && node.Pos() < write.End() && evaluator.isTargetStoreRead(node) {
			read = node
		}
	})
	if read == nil {
		read = write.AsCallExpression().Expression
	}
	textRange := rule.TokenRange(analysis.ctx.SourceFile, read)
	return analysis.ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
}
