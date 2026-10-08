# Concrete index boxing cost

The same compiler/runtime measured BlockArena.allocate followed by unit disposal against numeric baseline 75e1d014. Controls cover 0, 1 and 1,000 blocks; boxing_test.go retains them.

| Blocks | Numeric allocations/frees | Concrete allocations/frees | Numeric retains/releases | Concrete retains/releases |
|---:|---:|---:|---:|---:|
| 0 | 5/5 | 15/15 | 1/6 | 11/46 |
| 1 | 9/9 | 20/20 | 3/8 | 15/50 |
| 1,000 | 4,005/4,005 | 5,015/5,015 | 2,001/2,006 | 4,011/4,046 |

After startup subtraction, boxing costs one allocation/free and two retains/releases per block. The shared home adds 10 startup allocations/frees, 10 retains and 40 releases. Peak live allocations increase by one per block plus 10 at startup; regions remain zero. All allocations are freed. This measures block allocation, not AST/edge/traversal costs; private constructors and checked reads remain intact. testdata/boxing_baseline.json identifies the baseline.
