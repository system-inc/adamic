# Mutants for a parameter stored by index into an outer array (#22h1bv2)

Unrun. This change touches only internal/oracle, so the mutant below, an edit to internal/native, is for
a Linux box to run: make the edit, run the three tests in internal/oracle/region_index_store_test.go
with `ADAMIC_GATE_UNCACHED=1`, then put region.go back from a saved copy. Until a box runs it, its Red
column is what it is expected to show, not what anything has shown.

| Mutant | Edit | Red, expected |
|---|---|---|
| escapes forgets an index store | internal/native/region.go, escaping's `case ir.SetIndex:` arm: `if derived(statement.Value) { escaped = true }` removed, so storing the parameter by index no longer marks it escaping | TestRegionIndexStoreKeepsObjectOffRegion: index_store.a's C holds `ADAMIC_REGION` for `total += keep(make(7));`, the point make(7) built is allocated in that region and freed by `adamic_region_end` while outer[0] still holds it, and ASan reports heap-use-after-free at the print that reads `outer[0]?.label`. control.a and push_store.a stay green: control.a stores nothing, and push_store.a's `outer.push(p)` lowers to ir.ArrayPush, which isn't a consumer (borrow.go), so escaping's general operand check marks it |
