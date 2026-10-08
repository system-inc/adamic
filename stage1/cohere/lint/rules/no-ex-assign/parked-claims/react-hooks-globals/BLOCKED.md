# react-hooks/globals

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation`,
[cohere/internal/lint/rules/react/globals.go:617](/workspace/adamic/cohere/internal/lint/rules/react/globals.go:617).

No resolved identifier declaration/lexical ownership answer is exposed.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
