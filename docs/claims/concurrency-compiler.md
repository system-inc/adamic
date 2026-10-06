# Concurrency compiler claims (#p286ycm)

Branch: codex/concurrency-compiler, from origin/main (5d4c801).
Contract: origin/codex/concurrency:docs/concurrency.md.

Create internal/lower/shareable.go and internal/lower/parallel.go for structural
Shareable inference, capture checks and conservative task effect proofs.
Create internal/native/parallel.go for the runtime ABI call.
Add small ParallelMap hooks following ReadTextFile in internal/ir/ir.go,
internal/lower/input.go, internal/javascript/javascript.go,
internal/fresh/fresh.go and internal/native/emit_expressions.go.
Add concurrency oracle fixtures, exact refusal checks and tests for these proofs.
Add a small native gate in internal/oracle/oracle_test.go naming codex/concurrency.
Claim any necessary embedded adamic declaration and Node oracle shim hooks after
inspection, before editing them. Reuse cycles.go's traversal; any changes there
will be reported to its owner explicitly.

Do not edit internal/native/runtime/, internal/native/emit_objects.go,
internal/native/emit.go, internal/lower/lower.go or internal/native/native.go.

Inspection additions: internal/load/prelude.d.ts (the public signature),
oracle/adamic.mjs (the Node sequential witness), and a three-line entry hook in
internal/lower/refusals.go so task async/effects diagnostics precede general ones.
Create internal/oracle/concurrency_test.go, internal/lower/parallel_test.go and
internal/native/parallel_compiler_test.go. Concurrency fixtures use their own
oracle list until runtime integration, so existing native counts remain intact.

Exception propagation requires adding ParallelMap beside ArrayMap in
internal/lower/exceptions.go's closure-call case. This is a one-token hook;
without it a throwing task's caller would not carry the exception onward.
