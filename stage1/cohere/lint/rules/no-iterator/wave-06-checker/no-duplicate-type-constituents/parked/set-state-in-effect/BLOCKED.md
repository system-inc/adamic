# react-hooks/set-state-in-effect

Stopped on area `9b7547976bea1b00f8b04508b9102d03511f65b3`.

Required Go call: `high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)` at `cohere/internal/lint/rules/react/set_state_in_effect.go:267`.

Shared source-to-HIR lowering, SSA, captures and manual memoization erasure are absent. The checker alone does not supply that analysis.

Reproducer: `blocked.ts.txt`. This is an analysis input, not a successful parity witness. No descriptor is registered and no upstream parity or mutant result is claimed. Shared checker and helper files remain untouched.
