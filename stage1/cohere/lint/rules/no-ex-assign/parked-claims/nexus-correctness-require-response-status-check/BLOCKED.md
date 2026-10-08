# nexus/correctness-require-response-status-check

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `ctx.TypeChecker.GetShorthandAssignmentValueSymbol`,
[cohere/internal/lint/rules/nexus/correctness_require_response_status_check.go:486](/workspace/adamic/cohere/internal/lint/rules/nexus/correctness_require_response_status_check.go:486).

No shorthand binding-symbol identity answer or shared control_flow_graph.Build implementation is exposed.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
