# Mutants for the finder's own relation for selection (#nxb3xh1 head 6)

Each is one edit, run on a house box through Loom against the pins in shared_allocation_test.go and
internal/oracle's graph_shared_allocation_test.go, then put back.

| Mutant | File | Old text | New text | Red |
|---|---|---|---|---|
| A union side cut by its first member alone (Compiler's) | shared_allocation.go | `oneShape, otherShape := f.allocationShapeOf(one), f.allocationShapeOf(other)` in mayShareAllocation | the same, after `one` and `other` are each taken as their first member when a union | PENDING |
| viewAlsoTable widened to mayShareAllocation alone (Memory's condition A) | cycles.go | `viewAlsoTable(f.l, f.related, f.isFunction, f.shapes)` | `viewAlsoTable(f.l, f.mayShareAllocation, f.isFunction, f.shapes)` | PENDING |
| The edges built from every candidate, the queued set dropped | graph_types.go | the queue seeded from `seeds`, each link queued once by `queued` | the queue seeded from every candidate, a node built once by its edges key, every link queued | cost only: PENDING |
