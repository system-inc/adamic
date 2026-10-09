package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// deferredAtomicEvent is one event waiting to be placed, carrying the construct it belongs to and
// where it sat in the block's original event order.
type deferredAtomicEvent struct {
	within *ast.Node
	// originalIndex is how many kept events preceded this deferral when the graph emitted it. It is
	// the floor on where the deferral can land: an event the graph put before it stays before it.
	originalIndex int
	event         atomicEvent
}

// placeDeferredEvents moves the two event kinds whose real position the graph's hooks cannot
// express, leaving every other event in the order the graph emitted it.
//
// # The suspension
//
// An await or a yield makes every variable read BEFORE it stale, and upstream runs `makeOutdated` at
// the expression's `:exit`. The Expression hook is pre-order and it is the only hook the graph offers
// for these nodes, so the marker is emitted before the operand and has to move past it.
//
// The corpus pins both directions and they are one token apart: `foo = await foo` REPORTS, because
// the read inside the awaited operand is made stale by that very await, and `foo = await bar + foo`
// is CLEAN, because the read outside it evaluates afterwards.
//
// # The property store
//
// The graph's Write hook fires where a BINDING is stored into, and `foo.bar = x` stores into a
// property, so `foo` is only ever read there and no Write event exists. Upstream resolves the same
// asymmetry in `getWriteExpr`, walking UP from a read reference through enclosing member expressions
// to an assignment the chain is the left side of. The read is the anchor in both implementations.
//
// The store is judged once the right-hand side has evaluated, which is upstream's deferral to
// `:expression:exit`. Placing it beside the read is wrong and the corpus says so at once:
// `foo.bar += await baz` emits `R(foo) SUSPEND`, so a store next to the read is judged before the
// suspension that makes it stale, and 21 of upstream's 37 findings went silent under that placement.
//
// # Why the key is an original index plus containment, and not either alone
//
// Three keys were tried and the first two failed the corpus the same way, silence on 29 of 37
// findings, which is why the reasoning is recorded rather than the answer:
//
//	containment alone      an await's operand often emits NO events. `await amount` on an
//	                       undeclared global contributes nothing, so there is no contained event to
//	                       sit after and the suspension falls to the front of the block, ahead of
//	                       the read it was supposed to make stale.
//	a raw offset test      "starts before the await ends" also matches the assignment's own
//	                       left-hand side, which sits far to the LEFT of the operand while its store
//	                       happens to the right, so the suspension is pushed past the write instead.
//
// What is actually true is both at once: a deferral lands after every event the graph already put
// before it, and additionally after any later event that lies inside its construct. The first half
// is what the operand-with-no-events case needs, and the second is what moves a suspension past the
// reads inside its own operand.
func placeDeferredEvents(
	ctx rule.Context,
	graph *control_flow_graph.Graph[atomicEvent],
	escapes map[*ast.Symbol]bool,
	root *ast.Node,
) {
	for _, block := range graph.Blocks {
		var suspensions []deferredAtomicEvent
		var propertyStores []deferredAtomicEvent
		var kept []atomicEvent

		for _, event := range block.Events {
			switch event.kind {
			case atomicSuspend:
				if event.target == nil {
					kept = append(kept, event)
					continue
				}
				suspensions = append(suspensions, deferredAtomicEvent{
					within:        event.target,
					originalIndex: len(kept),
					// The node is kept on the placed event. Phase two decides where a property store
					// goes by asking which events lie inside its assignment, and a suspension with
					// no node is invisible to that test, which leaves the store ahead of it.
					event: atomicEvent{kind: atomicSuspend, target: event.target},
				})

			case atomicRead, atomicStoreAnchor:
				// An anchor is dropped rather than kept: it says where a store hangs and means
				// nothing to the lattice, which would otherwise treat it as a refresh.
				if event.kind == atomicRead {
					kept = append(kept, event)
				}
				if event.target == nil {
					continue
				}
				assignment, isProperty := propertyAssignmentHeadedBy(event.target)
				if !isProperty {
					continue
				}
				// The property arm asks the escape question with isMemberAccess true, which is what
				// makes a parameter report here and stay clean on the variable side.
				if !escapesEnclosingFunction(ctx, event.symbol, root, true, escapes) {
					continue
				}
				propertyStores = append(propertyStores, deferredAtomicEvent{
					within: assignment,
					// A store is placed against the stream that already carries the suspensions, so
					// its floor is counted there rather than here. Recomputed in phase two.
					originalIndex: 0,
					event: atomicEvent{
						kind:       atomicWrite,
						symbol:     event.symbol,
						target:     event.target,
						assignment: assignment,
						isProperty: true,
					},
				})

			default:
				kept = append(kept, event)
			}
		}

		if len(suspensions) == 0 && len(propertyStores) == 0 {
			block.Events = kept
			continue
		}

		// Two phases, because a property store belongs after EVERYTHING in its assignment and a
		// suspension inside the right-hand side is part of that. Placing both against the same
		// stream leaves a store that has no contained event of its own ahead of the suspension that
		// stales it, and the property arm goes largely silent: measured, 17 of upstream's 37
		// findings.
		//
		// The ORDER of the two phases is NOT load-bearing, which was measured rather than assumed
		// after a mutant swapping them survived. Each key is containment against a different range,
		// and an assignment contains its own await, so a store placed first still ends up after the
		// suspension once that suspension is placed. Four property shapes, including the doubled
		// `foo[bar].baz = await (foo.bar += await foo[bar].baz)`, produce identical findings under
		// both orders. The phases are kept separate because a store's floor is counted against the
		// stream that carries the suspensions, not because one must run first.
		withSuspensions := placeDeferrals(kept, suspensions)
		block.Events = placeDeferrals(withSuspensions, propertyStores)
	}
}

// placeDeferrals inserts each deferral after the later of two floors: where the graph originally put
// it, and the last event contained in its construct. The kept events keep their own order.
func placeDeferrals(kept []atomicEvent, deferrals []deferredAtomicEvent) []atomicEvent {
	placed := make([][]atomicEvent, len(kept)+1)

	for _, deferral := range deferrals {
		slot := deferral.originalIndex
		for index := slot; index < len(kept); index++ {
			target := kept[index].target
			if target == nil {
				continue
			}
			if target.Pos() >= deferral.within.Pos() && target.End() <= deferral.within.End() {
				slot = index + 1
			}
		}
		placed[slot] = append(placed[slot], deferral.event)
	}

	rebuilt := make([]atomicEvent, 0, len(kept)+len(deferrals))
	rebuilt = append(rebuilt, placed[0]...)
	for index, event := range kept {
		rebuilt = append(rebuilt, event)
		rebuilt = append(rebuilt, placed[index+1]...)
	}
	return rebuilt
}

// propertyAssignmentHeadedBy reports whether an identifier is the object at the head of a member
// chain that is the target of an assignment, and returns that assignment.
//
// This is upstream's `getWriteExpr` loop transcribed: from the identifier, climb while the node is
// the OBJECT of a member expression, and stop at an assignment whose left side the chain is. Being
// the object is the load-bearing half. In `bar.foo = 1` the identifier `foo` is the member NAME, not
// the object, so it heads nothing and the chain is `bar`'s.
//
// Both member forms are accepted because both occur in the corpus: `foo.bar` is a property access
// and `foo[bar].baz` puts an element access in the middle of the chain.
//
// # The object-position tests are unreachable through the graph, and are kept anyway
//
// A mutant dropping the property-access object test survived the whole corpus, and the reason is the
// graph rather than the fixtures: `controlflow`'s Read hook fires on the OBJECT of a member
// expression and never on its name. Measured directly on
// `const a = {}; const foo = {}; async function x() { a.foo = a.foo + await baz; }`, chosen because
// a variable and a member share the spelling `foo`, the hook produced reads for `a` at 50, `a` at 58
// and `baz`, and never for either `foo`. So no member name can reach this function at all.
//
// The tests are kept because this is called from exactly two places, both handing it an identifier
// straight off the Read hook: the read arm and the plain-target anchor arm in
// `analyzeRootNonAtomicUpdates`. That enumeration is what the verdict rests on, so a third caller
// passing an arbitrary identifier voids it, and the guards are what would keep such a caller
// correct. Deleting them would trade a measured-inert branch for a trap.
func propertyAssignmentHeadedBy(identifier *ast.Node) (*ast.Node, bool) {
	node := identifier
	sawMember := false

	for node != nil {
		parent := node.Parent
		if parent == nil {
			return nil, false
		}
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression != node {
				return nil, false
			}
			sawMember = true
			node = parent

		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression != node {
				return nil, false
			}
			sawMember = true
			node = parent

		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.Left != node || !isAssignmentOperatorToken(binary.OperatorToken) {
				return nil, false
			}
			// A chain with no member step is a plain variable assignment, which the Write hook
			// already recorded. Requiring at least one member step is what keeps the two arms from
			// both reporting the same store.
			if !sawMember {
				return nil, false
			}
			return parent, true

		default:
			return nil, false
		}
	}
	return nil, false
}

// escapesEnclosingFunction reports whether anything outside this code path root can observe the
// binding, and therefore whether a race on it is possible at all.
//
// This is upstream's `isLocalVariableWithoutEscape`, inverted so the name states the condition under
// which the rule REPORTS. Upstream:
//
//	if (!variable) return false;                                   an unresolved global escapes
//	if (isMemberAccess && defs.some(Parameter)) return false;       a parameter's object escapes
//	return references.every(r => r.from.variableScope === functionScope);
//
// # The three answers, and the shape of each
//
// **A binding declared outside this function escapes.** Its declaration lives in another scope, so
// any other function sharing that scope can write it while this one is suspended. Upstream reaches
// this through `variable.scope.variableScope !== functionScope` failing the `every`.
//
// **A binding captured by a closure escapes.** Even declared here, a nested function holding a
// reference can be called during the suspension. Upstream's `every` fails on the closure's own
// reference, whose `from.variableScope` is the closure.
//
// **A parameter escapes for the PROPERTY question only.** The object a parameter names was
// constructed by the caller and the caller still holds it, so a property write on it is observable
// outside regardless of what this function does with the binding. The binding itself is a fresh
// local slot. Upstream returns false early for exactly this pair, which is why the predicate answers
// two different things about one parameter.
//
// # How the scope question is asked here
//
// Upstream reads `variable.references` from `eslint-scope`, a resolved reference index we do not
// have. The equivalent question over the checker is asked from the declaration side instead: a
// binding whose declaration sits outside this root is reachable from outside it, and a binding
// declared inside is reachable from outside only if some occurrence of it lives in a nested root.
// Both are computed from `control_flow_graph.RootOf`, which is the same boundary the graph itself uses, so
// the answer cannot disagree with the graph about what "this function" means.
func escapesEnclosingFunction(
	ctx rule.Context,
	symbol *ast.Symbol,
	root *ast.Node,
	isMemberAccess bool,
	memo map[*ast.Symbol]bool,
) bool {
	if symbol == nil {
		// Upstream's `if (!variable) return false`, so an unresolved name is not judged. Our
		// resolution never reaches here with nil because the callers check, but the guard is kept
		// so the predicate is total and reads the same as upstream's.
		return false
	}
	if len(symbol.Declarations) == 0 {
		return false
	}

	if isMemberAccess && symbolIsParameter(symbol) {
		// The early return that makes a parameter answer differently on the two arms. It is checked
		// BEFORE the memo, because the memo is keyed by symbol alone and this answer depends on
		// which arm is asking.
		return true
	}

	if cached, seen := memo[symbol]; seen {
		return cached
	}
	escaped := computeEscape(ctx, symbol, root)
	memo[symbol] = escaped
	return escaped
}

// symbolIsParameter reports whether any declaration of a binding is a parameter.
//
// Upstream asks `variable.defs.some(d => d.type === "Parameter")`, which is a disjunction over every
// definition rather than a question about one, so this loops rather than indexing. A binding can
// carry more than one declaration and the index-zero shortcut would answer about whichever the
// checker happened to order first.
func symbolIsParameter(symbol *ast.Symbol) bool {
	for _, declaration := range symbol.Declarations {
		if declaration != nil && declaration.Kind == ast.KindParameter {
			return true
		}
	}
	return false
}

// computeEscape answers the declaration-side and capture-side halves of the escape question.
func computeEscape(ctx rule.Context, symbol *ast.Symbol, root *ast.Node) bool {
	// The declaration side. A binding declared outside this root is shared with whatever else lives
	// in that scope, so it escapes.
	declaredInside := false
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		if control_flow_graph.RootOf(declaration) == root {
			declaredInside = true
			break
		}
	}
	if !declaredInside {
		return true
	}

	// The capture side. A binding declared here still escapes if a nested function names it, because
	// that function can run during the suspension.
	return symbolIsReadFromNestedRoot(ctx, symbol, root)
}

// symbolIsReadFromNestedRoot reports whether any occurrence of a binding lives in a code path root
// nested inside this one.
//
// The name filter is what keeps this cheap and it cannot err in the unsafe direction: two bindings
// sharing a name are still separated by symbol identity, and an occurrence of a differently named
// binding can never be a capture of this one.
//
// Measured against the installed rule on the pair the corpus ships for exactly this:
//
//	async function x() { let foo; bar(() => foo); foo += await amount; }        reports
//	async function x() { let foo; bar(() => baz += 1); foo += await amount; }   clean
//
// The second is the control. A closure exists in both, and only the one that names the binding
// changes the verdict, which is what proves this is testing capture rather than the presence of a
// nested function.
func symbolIsReadFromNestedRoot(ctx rule.Context, symbol *ast.Symbol, root *ast.Node) bool {
	name := symbol.Name
	if name == "" {
		return false
	}

	captured := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil || captured {
			return false
		}
		if node.Kind == ast.KindIdentifier && node.Text() == name {
			if !isNonReferenceIdentifier(node) && control_flow_graph.RootOf(node) != root {
				if resolveOccurrence(ctx, node) == symbol {
					captured = true
					return true
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)

	return captured
}
