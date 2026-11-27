# spread_accessor_reads.go mutants (#vjvcnnj head 3)

Each mutant is one edit to `spread_accessor_reads.go`, run against `TestSpreadPrototypeField`
(internal/oracle/spread_prototype_test.go), `TestSpreadAccessorNames`,
`TestSpreadAccessorReadChecksTheKind` and `TestSpreadAccessorReadKeepsMainsRead` (spread_accessor_reads_test.go).
M0, no mutant, is the control and is green on every row. The script is the job's command.

| mutant | edit | red on |
|---|---|---|
| M1, the getter branch dropped | the read by name is `ir.DynamicProperty`, an own-slot read that never runs a getter | pending |
| M2, the slot kind check dropped | the helper's `If` that panics with `spreadAccessorReadMismatch` is left out | pending |
| M3, acceptsUndefined's gate dropped | every read of a name in the set is read by name, observed or not | pending |
| M4, setter-only names left out | the name set takes a member only when its kind is a getter | pending |
| M5, the wide gate | spreadAccessorStops answers true for every recorded read, so each is read by name | pending |
| M6, the pass finds nothing | spreadAccessorReads returns before its walk, so every recorded read stays main's plain read | pending |
