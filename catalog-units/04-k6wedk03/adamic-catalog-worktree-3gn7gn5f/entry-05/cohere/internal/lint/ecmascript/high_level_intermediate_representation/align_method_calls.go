// Making a method call and its property agree about scopes.
//
// This is React's `alignMethodCallScopes` (`ReactiveScopes/AlignMethodCallScopes.ts`), the first of
// three passes upstream runs immediately before `alignReactiveScopesToBlockScopes`
// (`Entrypoint/Pipeline.ts:350-384`).
//
// # What it guarantees
//
// After this pass, a `MethodCall` instruction satisfies one of two states: the call's lvalue and the
// property it reads share a scope, or neither has one. The in-between state -- one scoped, one not,
// or two different scopes -- is what it removes.
//
// # Why the in-between state is a problem rather than an imprecision
//
// `a.b()` lowers to a `PropertyLoad` producing the method and a `MethodCall` consuming it. The two
// are one expression in source and the scope machinery treats them as separate values, so they can
// land in different equivalence classes. A scope containing the call but not the property is a scope
// whose range starts after a value it depends on was computed, which is exactly the shape
// `AlignReactiveScopesToBlockScopes` then widens across a block boundary.
//
// Measured on `drawHighlightedCountryOutlines`, the one corpus function whose alignment residue is
// non-zero once declaration ids stop colliding: it is a `for...of` over a `Map` with a
// `console.log`, method calls throughout.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"sort"
)

// AlignMethodCallScopes makes every method call agree with its property about scopes.
//
// Returns a new table; the input is not mutated, matching how `AlignReactiveScopesToBlockScopes`
// takes a `ReactiveScopes` and hands back widened ranges rather than writing through.
//
// # Two mechanisms, and they are not interchangeable
//
// When both sides carry a scope, the two scopes MERGE: the survivor's range grows to cover both.
// When only one side carries one, the other side's scope ASSIGNMENT changes -- the property joins
// the call's scope, or loses its own. Upstream keeps these apart with a disjoint set for the first
// and a mapping for the second, and the distinction matters because a merge changes a range while a
// mapping changes membership.
func AlignMethodCallScopes(function *Function, scopes *ReactiveScopes) *ReactiveScopes {
	if function == nil || scopes == nil || scopes.Len() == 0 {
		return scopes
	}

	// scopeMapping records a property whose scope must become the call's, or must be dropped. A
	// `ScopeId` of zero means "no scope", which is this tree's spelling of upstream's null.
	scopeMapping := map[static_single_assignment.IdentifierId]ScopeId{}
	// merged unions scopes that must become one. Keyed by `ScopeId` widened into the identifier
	// space, because `DisjointSet` is keyed by `IdentifierId` and the two are distinct types.
	merged := &DisjointSet{}

	for _, instruction := range function.Instructions {
		if instruction == nil {
			continue
		}
		call, isMethodCall := instruction.Value.(*MethodCall)
		if !isMethodCall {
			continue
		}
		lvalueScope := scopes.ScopeOf(instruction.LValue.Identifier)
		propertyScope := scopes.ScopeOf(call.Property.Identifier)

		// # Two of these three arms are unreachable on this corpus, and that is measured
		//
		// Over 677 functions and 1,849 method calls: 1,672 calls already have both sides in one scope,
		// the property-scoped arm fires 177 times, and the other two arms fire zero times. A mutation
		// deleting either unreachable arm survives the whole suite, and the reason is the population
		// rather than a missing fixture.
		//
		// Unknown and known-mutable call results allocate, so their lvalue and property are already
		// joined by the disjoint pass. The remaining asymmetry comes from known primitive-returning
		// methods such as `includes`, `has`, `join`, and `push`: their lvalue does not allocate, while
		// the `PropertyLoad` result routinely receives a scope.
		//
		// Both unreachable arms are kept because they are upstream's and because the current
		// population is not a proof that the language cannot produce them. The verdict expires when
		// either arm receives a real input.
		switch {
		case lvalueScope != 0 && propertyScope != 0:
			if lvalueScope != propertyScope {
				merged.Union([]static_single_assignment.IdentifierId{static_single_assignment.IdentifierId(lvalueScope), static_single_assignment.IdentifierId(propertyScope)})
			}
		case lvalueScope != 0:
			// The call is scoped and the property is not, so the property joins the call's scope.
			scopeMapping[call.Property.Identifier] = lvalueScope
		case propertyScope != 0:
			// The call is unscoped, so the property does not need a scope either.
			scopeMapping[call.Property.Identifier] = 0
		}
	}

	if len(scopeMapping) == 0 && merged.Size() == 0 {
		return scopes
	}
	return rebuildWithAlignedMethodCalls(scopes, scopeMapping, merged)
}

// rebuildWithAlignedMethodCalls applies the merges and the mapping to produce a new scope table.
//
// The two are applied in that order because a mapping names a scope by id, and a merge changes which
// id survives. Applying the mapping first would assign a property to a scope that the merge is about
// to fold away.
func rebuildWithAlignedMethodCalls(scopes *ReactiveScopes, scopeMapping map[static_single_assignment.IdentifierId]ScopeId,
	merged *DisjointSet) *ReactiveScopes {
	// survivorOf resolves a scope through the merge, or returns it unchanged.
	survivorOf := func(scope ScopeId) ScopeId {
		if scope == 0 || !merged.Has(static_single_assignment.IdentifierId(scope)) {
			return scope
		}
		return ScopeId(merged.RepresentativeOf(static_single_assignment.IdentifierId(scope)))
	}

	rebuilt := &ReactiveScopes{
		byIdentifier: make(map[static_single_assignment.IdentifierId]ScopeId, len(scopes.byIdentifier)),
		ranges:       map[ScopeId]mutation_aliasing.MutableRange{},
		members:      map[ScopeId][]static_single_assignment.IdentifierId{},
	}

	for identifier, scope := range scopes.byIdentifier {
		target := survivorOf(scope)
		if mapped, remapped := scopeMapping[identifier]; remapped {
			target = survivorOf(mapped)
		}
		if target == 0 {
			continue
		}
		rebuilt.byIdentifier[identifier] = target
	}
	// A property that had no scope and joins the call's is not in `byIdentifier` yet.
	for identifier, mapped := range scopeMapping {
		if mapped == 0 {
			continue
		}
		rebuilt.byIdentifier[identifier] = survivorOf(mapped)
	}

	// Ranges: a merged scope takes the union of its members' ranges, which is upstream's
	// `Math.min(start)` and `Math.max(end)` over the disjoint set.
	for _, scope := range scopes.Ids() {
		target := survivorOf(scope)
		existing, present := rebuilt.ranges[target]
		incoming := scopes.RangeOf(scope)
		if !present {
			rebuilt.ranges[target] = incoming
			continue
		}
		if incoming.Start < existing.Start {
			existing.Start = incoming.Start
		}
		if incoming.End > existing.End {
			existing.End = incoming.End
		}
		rebuilt.ranges[target] = existing
	}

	// Membership and order are derived from the rebuilt index rather than copied, so a property that
	// changed scopes appears in exactly one members list.
	for _, scope := range scopes.Ids() {
		target := survivorOf(scope)
		if _, seen := rebuilt.members[target]; !seen {
			rebuilt.members[target] = nil
			rebuilt.order = append(rebuilt.order, target)
		}
	}
	for identifier, scope := range rebuilt.byIdentifier {
		rebuilt.members[scope] = append(rebuilt.members[scope], identifier)
	}
	for scope := range rebuilt.members {
		sort.Slice(rebuilt.members[scope], func(a, b int) bool {
			return rebuilt.members[scope][a] < rebuilt.members[scope][b]
		})
	}
	return rebuilt
}
