# require-atomic-updates

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `control_flow_graph.Build`,
[cohere/internal/lint/rules/core/require_atomic_updates.go:402](/workspace/adamic/cohere/internal/lint/rules/core/require_atomic_updates.go:402).

Shared native control-flow graph builder with read/write/suspend hooks is missing; no private copy was made.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
