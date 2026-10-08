# Step 24 counts

Measured on TypeScript 6.0.3; source-site categories overlap. Native heap-event counts are unavailable because these four programs are refused.

| Fixture | Node stdout bytes | Node exit | Stage0 | Mutant caught |
| --- | --- | --- | --- | --- |
| 01_factory_clone.a | 131 | 0 | Refused | Node stdout |
| 02_parser_range.a | 113 | 0 | Refused | Node stdout |
| 03_repair_flags.a | 71 | 0 | Refused | Node stdout |
| 04_missing_presence.a | 96 | 0 | Refused | Node stdout |

Four fixtures, four fixture mutants, one corpus-presence mutant.

| Census category | Sites |
| --- | --- |
| createNodeArray call (may reuse) | 159 |
| setTextRange on array (range initialization/update) | 56 |
| NodeArray.slice (ordinary array result) | 20 |
| array/other value promoted by assertion | 3 |
| slice promoted to NodeArray | 1 |
| Direct metadata reads | 87 |
| Direct metadata writes | 10 |
| Reflection | 4 |
| Shared range-helper property accesses | 4 |
| Array-target range calls | 56 |
| Array-provider arguments | 46 |
| Other NodeArray-returning calls | 593 |
| Computed numeric element accesses | 82 |

Corpus: 81 files, 505,295 scanner tokens, no arrays observed during scanning. Each parse/incremental/debug pass has 137,742 final-tree arrays and 941,215 nodes, with zero parse diagnostics. The instrumented run records 12,511,022 observations of 441,988 distinct array objects. No pristine present-undefined field state was observed.
