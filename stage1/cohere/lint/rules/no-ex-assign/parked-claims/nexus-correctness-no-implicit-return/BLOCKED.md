# nexus/correctness-no-implicit-return

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `checker.Checker_functionHasImplicitReturn`,
[cohere/internal/lint/rules/nexus/correctness_no_implicit_return.go:96](/workspace/adamic/cohere/internal/lint/rules/nexus/correctness_no_implicit_return.go:96).

No function end-flow/implicit-return answer is exposed.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
