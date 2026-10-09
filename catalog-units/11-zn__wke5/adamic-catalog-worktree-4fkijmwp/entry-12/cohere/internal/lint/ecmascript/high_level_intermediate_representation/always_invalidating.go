// Which values invalidate on every render, asked of the resident checker rather than inferred.
//
// This is React's `isAlwaysInvalidatingType` (`ReactiveScopes/MergeReactiveScopesThatInvalidate
// Together.ts:505` at `react_conformance.UpstreamSha`). It answers one question: is this value of a
// type that produces a fresh identity every render, so that memoizing it alone buys nothing.
//
// Upstream's answer is a shape-registry lookup -- `type.kind === 'Object'` with a `shapeId` of
// `BuiltInArray`, `BuiltInObject`, `BuiltInFunction` or `BuiltInJsx`, or `type.kind === 'Function'`.
// That registry exists because Babel has no type checker, so upstream carries ~1,450 lines of local
// inference to reconstruct what a checker already knows.
//
// # We do not owe that port, and this file is the measurement rather than the claim
//
// `Identifier.Node` carries the real TypeScript node -- when measured, 60,819 of 61,738 corpus
// identifiers had one, 98.5% -- and cohere runs the checker in the same process. So the question goes to
// `GetTypeAtLocation` and comes back with the answer the compiler actually computed, through
// imports and generics, which no amount of local inference recovers. `reactive.go` established this
// pattern; this is its second consumer.
//
// # What the checker reports, measured before this was written
//
//	const arr = [1,2,3]   flags=Object        symbol="Array"          callSignatures=0
//	const obj = {a:1}     flags=Object        symbol="\xfeobject"     callSignatures=0
//	const fn = () => 1    flags=Object        symbol="\xfefunction"   callSignatures=1
//	const num = 42        flags=NumberLiteral                         callSignatures=0
//	const str = "s"       flags=StringLiteral                         callSignatures=0
//
// The four invalidating shapes are separable and the primitives are separable from all of them,
// which is the whole requirement. The two internal symbol names are deliberately NOT matched on:
// `\xfeobject` and `\xfefunction` are the checker's own anonymous-type placeholders and are not a
// documented surface. Object-ness is read from the type flag and function-ness from the presence of
// a call signature, both of which are stable API.
package high_level_intermediate_representation

import (
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/static_single_assignment"
)

// IsAlwaysInvalidatingType reports whether a value of this type gets a fresh identity every render.
//
// True for arrays, plain objects, functions and JSX -- the four upstream names -- because each
// allocates on every evaluation, so `Object.is` on two renders' values is always false and a scope
// producing only such a value invalidates unconditionally.
//
// False when the checker is absent or the node is nil, which is the conservative direction: a false
// answer means "do not treat this scope as always-invalidating", so a missing type declines to
// merge rather than merging on a guess. That matters because the two upstream call sites both use a
// true answer to permit a merge.
func IsAlwaysInvalidatingType(function *Function, id static_single_assignment.IdentifierId,
	typeChecker *shimchecker.Checker) bool {
	if function == nil || typeChecker == nil {
		return false
	}
	if int(id) >= len(function.Identifiers) {
		return false
	}
	identifier := function.Identifiers[id]
	if identifier == nil || identifier.Node == nil {
		// Not defensive: `GetTypeAtLocation` dereferences the node and segfaults on nil. Measured
		// by deleting this guard, which crashes the corpus test rather than returning a wrong
		// answer. When measured, 645 of 40,241 corpus identifiers had no node -- the values `high_level_intermediate_representation.go`
		// documents as having no direct syntactic source.
		return false
	}
	valueType := typeChecker.GetTypeAtLocation(identifier.Node)
	if valueType == nil {
		return false
	}

	// # One flag answers all four shapes, and the function arm upstream needs is redundant here
	//
	// Upstream tests `type.kind === 'Function'` separately from its four object shapes, because its
	// registry models functions as their own kind. In TypeScript's type system a function IS an
	// object type: measured across an arrow, a function declaration, a type-alias callback, a
	// callable interface and an overloaded signature, every one reports the Object flag set, with
	// call signature counts of 1, 1, 1, 1 and 2. A separate call-signature arm was written first
	// and removed after a mutation sweep showed deleting it changed no answer.
	//
	// So this is one check rather than two, and the reduction is measured rather than assumed. If a
	// callable form is ever found that lacks the Object flag, the arm comes back with that form as
	// its fixture.
	//
	// Arrays, plain objects and JSX elements carry the same flag. A number, string, boolean, null
	// and `any` do not, which is exactly the separation the predicate needs: those are comparable
	// with `Object.is` across renders and the four invalidating shapes are not.
	//
	// `any` is worth naming because it is the one case where upstream and this could plausibly
	// differ: it reports flags=Any, so this answers false, and upstream's registry has no entry for
	// it either. Both decline, which is the conservative direction.
	return shimchecker.Type_flags(valueType)&shimchecker.TypeFlagsObject != 0
}
