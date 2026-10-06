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
