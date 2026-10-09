// How strongly a value needs memoizing, which is the lattice `pruneNonEscapingScopes` walks.
//
// This is React's `MemoizationLevel` and `joinAliases`
// (`ReactiveScopes/PruneNonEscapingScopes.ts` at `react_conformance.UpstreamSha`).
//
// # Four levels rather than a boolean, and the difference is not cosmetic
//
// The obvious model is "does this value escape". That model passes every fixture where a returned
// value must be memoized and is wrong on every fixture where memoization propagates: a value that
// does not escape may still need memoizing because something that does escape depends on it, and a
// value that escapes may still be free to recompute because it is cheaply comparable.
//
// The four levels separate those questions. `Memoized` needs memoizing if it escapes at all;
// `Conditional` needs it only when something it feeds does; `Unmemoized` is not comparable with
// `Object.is` but is left alone unless forced; `Never` is cheap to compare and is never worth
// memoizing.
package high_level_intermediate_representation

// MemoizationLevel is how a value's memoization is decided relative to its dependencies.
//
// The constant order is the join order and is load-bearing: `JoinMemoizationLevels` takes the
// maximum, so a value appearing as an lvalue at two levels gets the stronger one. Reordering these
// silently changes every classification.
type MemoizationLevel uint8

const (
	// MemoizationNever is a value cheaply comparable with `Object.is`, so memoizing it buys nothing.
	MemoizationNever MemoizationLevel = iota
	// MemoizationUnmemoized is a value that cannot be compared with `Object.is` but which is left
	// alone unless memoization is forced.
	MemoizationUnmemoized
	// MemoizationConditional is a value memoized only when its dependencies are.
	//
	// The propagating shapes: a ternary, a logical, a sequence, a load. They produce no new identity
	// of their own, so whether they need memoizing is entirely a question about what flows through
	// them.
	MemoizationConditional
	// MemoizationMemoized is a value that needs memoizing if it escapes.
	//
	// The allocating shapes: array and object literals, calls, JSX, `new`. Each produces a fresh
	// identity every evaluation, so anything comparing across renders sees a change unless the value
	// is held.
	MemoizationMemoized
)

// JoinMemoizationLevels returns the stronger of two levels.
//
// Upstream's `joinAliases`, which is written as a four-arm if-chain in the same order. An identifier
// can be an lvalue more than once -- reassigned, or destructured alongside another binding -- and the
// final level is the strongest any of those assignments demanded.
//
// A max over the constant order rather than a chain, because the two are the same function and the
// max says why: the levels are ordered by strength and the join is "whichever demanded more".
func JoinMemoizationLevels(first, second MemoizationLevel) MemoizationLevel {
	if first > second {
		return first
	}
	return second
}

func (level MemoizationLevel) String() string {
	switch level {
	case MemoizationNever:
		return "never"
	case MemoizationUnmemoized:
		return "unmemoized"
	case MemoizationConditional:
		return "conditional"
	case MemoizationMemoized:
		return "memoized"
	default:
		return "<unknown memoization level>"
	}
}
