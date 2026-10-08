# no-else-return

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `ctx.TypeChecker.GetSymbolsInScope`,
[cohere/internal/lint/rules/core/no_else_return.go:372](/workspace/adamic/cohere/internal/lint/rules/core/no_else_return.go:372).

scope-locals exposes binder tables, not GetSymbolsInScope at an arbitrary if with declarations and lexical scope ownership.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
