# nexus/performance-no-independent-await-in-loop

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `checker.SkipAlias`,
[cohere/internal/lint/rules/nexus/performance_no_independent_await_in_loop.go:479](/workspace/adamic/cohere/internal/lint/rules/nexus/performance_no_independent_await_in_loop.go:479).

No stable resolved/alias-skipped symbol identities are exposed.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
