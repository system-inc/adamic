# Filter overload result-use count — checkpoint

Task #ev14ch9. Report only; no adaptation is authorized by this checkpoint.
Base: 03c140ec28a249b8b0f13eb24e34fc3d3df05266.
Upstream: TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

The unchanged adapted tree is still being constructed. Definitive counts are pending.
The checker traversal reuses supplement.cjs from typescript/filter-overload-sites,
05a2ba742a413765fd4ca636c5dbcdae1f07e948, with testRunner and typings included.
The original tool is retained as supplement-original.cjs. contexts.cjs additionally
records all four writable-result overloads (1, 2, 5, 6), their use contexts and
symbol-resolved references for local result variables. No grep discovers callers.

filter is @internal and has no declaration in upstream's public API baseline.
The implementation returns the original argument whenever no element is dropped.
No copy or checked view is proposed. Nothing is submitted to Loom.
