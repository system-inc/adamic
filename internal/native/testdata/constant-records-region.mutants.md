# constantRecords' region refusal, and its mutant (#j2qkavx)

constantRecords (internal/native/constant_tables.go) batches a run of closed records of one layout into a static table and one noinline loop, now in functions as well as in module initialization. Region placement is emitter state, not on the IR, so it refuses whenever `e.inRegion` or `e.regionLiteralDepth` is set (Emit's ruling on #j2qkavx).

Measured through `ahra loom job` J-0zy1hq2 at 55dfd948e6bac4eab97b7cad1583e8c31999dfe0, on a rented machine. The edit was applied once (perl counts the substitution and refuses otherwise), the test ran, and the file was restored from a saved copy.

| Mutant | Exact edit | Red |
| --- | --- | --- |
| The region refusal dropped | constant_tables.go, `constantRecords`: `if e.inRegion \|\| e.regionLiteralDepth != 0 \|\| e.fieldTypesNeeded()` made `if e.fieldTypesNeeded()` | TestConstantRecordsStayOutOfRegions: "a fresh function's region version: constantRecords took the table true, want false", "a literal handed a region: constantRecords took the table true, want false", and "build's region version batched its records into the table" |

Under the mutant, the test's ASan and leak witness stays green: allocations 67, frees 64, regions 3, balanced, with retains and releases 0 and 4 where the control has 60 and 64. That's because an array literal is never region-made. Fresh functions return ir.Object (region.go), and handRegion sends a field that isn't a call through e.value, so the array build_in makes is a heap array either way. The refusal fails closed rather than preventing a wrong run, and its red is the emitted body and the state probes.
