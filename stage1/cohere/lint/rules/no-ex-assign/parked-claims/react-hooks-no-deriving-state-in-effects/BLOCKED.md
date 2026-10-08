# react-hooks/no-deriving-state-in-effects

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `high_level_intermediate_representation.ForFunctionWithoutManualMemoization`,
[cohere/internal/lint/rules/react/no_deriving_state_in_effects.go:142](/workspace/adamic/cohere/internal/lint/rules/react/no_deriving_state_in_effects.go:142).

Shared memo-erased React HIR/SSA/capture analysis remains missing.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
