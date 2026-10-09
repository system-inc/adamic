package high_level_intermediate_representation

import "testing"

// TestJoinMemoizationLevelsMatchesUpstreamChain asserts the join against upstream's own spelling.
//
// `JoinMemoizationLevels` is a max over the constant order; upstream's `joinAliases` is a four-arm
// if-chain testing Memoized, then Conditional, then Unmemoized, then falling through to Never. Those
// are the same function only because the constants happen to be ordered by strength, and that is
// exactly the kind of coincidence a later edit breaks.
//
// So the chain is written out here and every one of the sixteen pairs is compared against it. A
// reordering of the constants fails this immediately rather than silently changing classifications.
func TestJoinMemoizationLevelsMatchesUpstreamChain(t *testing.T) {
	t.Parallel()

	// upstream's `joinAliases`, transcribed arm for arm rather than expressed as a max.
	chain := func(first, second MemoizationLevel) MemoizationLevel {
		switch {
		case first == MemoizationMemoized || second == MemoizationMemoized:
			return MemoizationMemoized
		case first == MemoizationConditional || second == MemoizationConditional:
			return MemoizationConditional
		case first == MemoizationUnmemoized || second == MemoizationUnmemoized:
			return MemoizationUnmemoized
		default:
			return MemoizationNever
		}
	}

	levels := []MemoizationLevel{
		MemoizationNever, MemoizationUnmemoized, MemoizationConditional, MemoizationMemoized,
	}
	compared := 0
	for _, first := range levels {
		for _, second := range levels {
			compared++
			want := chain(first, second)
			if got := JoinMemoizationLevels(first, second); got != want {
				t.Errorf("join(%s, %s) = %s, want %s", first, second, got, want)
			}
			// Commutativity, which upstream's chain has and a careless max-like function might not.
			if got := JoinMemoizationLevels(second, first); got != want {
				t.Errorf("join(%s, %s) = %s, want %s (join must be commutative)",
					second, first, got, want)
			}
		}
	}
	if compared != 16 {
		t.Fatalf("compared %d pairs, want 16; the level set changed and this test did not", compared)
	}
}

// TestMemoizationLevelOrderIsStrength pins the ordering the join depends on.
//
// Every classification in `pruneNonEscapingScopes` is a comparison against these constants, so the
// order is part of the contract rather than an implementation detail. Written as an explicit
// sequence so a reordering fails here with a clear message rather than in a downstream pass with an
// obscure one.
func TestMemoizationLevelOrderIsStrength(t *testing.T) {
	t.Parallel()

	if !(MemoizationNever < MemoizationUnmemoized &&
		MemoizationUnmemoized < MemoizationConditional &&
		MemoizationConditional < MemoizationMemoized) {
		t.Fatalf("the levels are not ordered by strength: never=%d unmemoized=%d conditional=%d "+
			"memoized=%d; the join is a max over this order and every classification compares "+
			"against it", MemoizationNever, MemoizationUnmemoized, MemoizationConditional,
			MemoizationMemoized)
	}

	// Each level prints distinctly, since the strings appear in every diagnostic the pass produces.
	seen := map[string]bool{}
	for _, level := range []MemoizationLevel{
		MemoizationNever, MemoizationUnmemoized, MemoizationConditional, MemoizationMemoized,
	} {
		name := level.String()
		if seen[name] {
			t.Errorf("two levels print as %q", name)
		}
		seen[name] = true
	}
}
