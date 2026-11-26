# spreadGetters' relation, an object literal's holder by its regular type (#z6h9s7k)

**M1, the raw holder.** spreadGetters' relation in class_accessors.go,
`!l.checker.IsTypeAssignableTo(l.relatedShape(holder), viewed)`, made
`!l.checker.IsTypeAssignableTo(holder, viewed)`, main's code before this change. An object literal's own type is fresh
and fails a narrower view by excess properties, so its getter is left out of the spread's set. Red:

- `TestSpreadRecordsTheGettersItMayRun`, its two narrower-view rows (internal/lower/spread_getters_test.go).
- `TestSpreadGettersHolderReadersAreListed` (internal/lower/spread_getters_source_test.go).
- `TestReuseSpreadAgreesWithNode`, literal_getter_narrow_view.a, literal_getter_unannotated.a and
  literal_getter_global_first.a: reuse lends the global past the getter that replaces it, a heap use after free natively
  under ASan (internal/oracle/reuse_spread_test.go).
- `TestWasmPurityInspectsANarrowSpreadsLiteralGetter`: compute is judged pure and reaches the Wasm build
  (internal/worker/purity_spread_test.go).

Must stay green under it: the four rows reuse_spread held before, and every row of TestSpreadRecordsTheGettersItMayRun
whose view is the literal's own.
