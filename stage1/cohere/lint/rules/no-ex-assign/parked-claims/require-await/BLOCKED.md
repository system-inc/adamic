# require-await

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `checker.Checker_getResolvedSignature; resolved.Target()`,
[cohere/internal/lint/rules/core/require_await.go:988](/workspace/adamic/cohere/internal/lint/rules/core/require_await.go:988).

signature-shape does not expose the declared generic target and its type parameters/substitutions needed to avoid inferred self-demand.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
