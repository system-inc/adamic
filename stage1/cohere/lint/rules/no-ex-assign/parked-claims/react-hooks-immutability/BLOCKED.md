# react-hooks/immutability

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `high_level_intermediate_representation.MayHoldComponentOrHook`,
[cohere/internal/lint/rules/react/immutability.go:166](/workspace/adamic/cohere/internal/lint/rules/react/immutability.go:166).

Shared React HIR/SSA/capture analysis remains missing.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
