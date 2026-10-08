# Concrete index boxing lesson

The same workload allocates N blocks through BlockArena.allocate, prints arena.length, and disposes the arena. The numeric baseline is commit 75e1d014e7fb4f389383f8054aabecb0a6d60eec, measured with the same compiler and runtime as the concrete-index port. The historical measurement copied only that commit's arena into an isolated temporary fixture; no numeric implementation or compatibility switch is retained in the port. boxing_test.go keeps the 0, 1 and 1,000 block controls.

| Blocks | Representation | Allocations / frees | Retains | Releases | Peak live allocations | Regions |
|---:|---|---:|---:|---:|---:|---:|
| 0 | numeric baseline | 5 / 5 | 1 | 6 | 4 | 0 |
| 0 | concrete indices | 15 / 15 | 11 | 46 | 14 | 0 |
| 1 | numeric baseline | 9 / 9 | 3 | 8 | 8 | 0 |
| 1 | concrete indices | 20 / 20 | 15 | 50 | 19 | 0 |
| 1,000 | numeric baseline | 4,005 / 4,005 | 2,001 | 2,006 | 4,004 | 0 |
| 1,000 | concrete indices | 5,015 / 5,015 | 4,011 | 4,046 | 5,014 | 0 |

Removing the zero-block startup cost isolates one additional allocation/free and two additional retains/releases per block. The shared home costs another 10 startup allocations/frees, 10 retains and 40 releases. Peak live allocations rise by one per block plus 10 at startup. Every allocation is freed.

This is the compiler lesson: a sound one-field index currently boxes once per block, and importing the shared concrete-class home adds measured startup work. Keeping the private constructors, nominal classes and checked reads takes precedence over removing that cost. These controls measure block allocation only; they do not claim to isolate AST boxing, edges, traversal caches or copied stacks in the full linter.

Raw receipts are evidence/boxing-controls.log.gz and evidence/boxing-historical-controls.log.gz. testdata/boxing_baseline.json identifies the historical implementation and 1,000-block numbers.
