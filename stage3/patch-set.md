# Patch set

Compared with the pinned pristine tree after upstream diagnostic generation.
Rows measure incremental edits; total measures the final tree against pristine.

| Adaptation | Files | Lines added | Lines removed |
|---|---:|---:|---:|
| 00-setup | 0 | 0 | 0 |
| 10-type-imports | 73 | 3720 | 3720 |
| 20-optional-declarations | 24 | 371 | 371 |
| 30-indexed-reads | 11 | 209 | 209 |
| 31-indexed-reads-checker | 1 | 394 | 394 |
| 32-indexed-reads-program | 20 | 135 | 135 |
| 33-indexed-reads-emit | 16 | 119 | 119 |
| 40-explicit-any | 17 | 106 | 106 |
| 41-explicit-any-remaining | 10 | 40 | 37 |
| 45-regex-captures | 5 | 12 | 10 |
| 46-fix-pragma-empty-argument | 1 | 1 | 1 |
| 47-host-errors | 3 | 16 | 5 |
| 48-memoize | 1 | 2 | 2 |
| 50-temporary-scanner-implicit-returns | 1 | 2 | 0 |
| 51-temporary-scanner-fallthrough | 0 | 0 | 0 |
| 60-temporary-factory-local-symbol | 1 | 1 | 1 |
| 61-temporary-factory-type-expression | 1 | 1 | 1 |
| 62-temporary-tracing-write | 1 | 1 | 1 |
| 63-temporary-tracing-legend | 1 | 1 | 1 |
| 64-temporary-parser-range-read | 1 | 1 | 1 |
| 65-temporary-node-builtins | 2 | 7 | 2 |
| 70-readonly-views | 7 | 22 | 22 |
| 71-writable-views | 20 | 88 | 88 |
| 75-optional-widening | 1 | 17 | 6 |
| 76-truthful-casts | 1 | 1 | 1 |
| **Total** | pending | pending | pending |

65-temporary-node-builtins is temporary: it retires when Adamic accepts require
with a literal specifier and Node builtin types. Its two-file edit erases on Node;
core already has the required local process view on current main.

76-truthful-casts is permanent. It gives the private assertion cache its proven
six-key type; the truthful sameMap union proposal is not applied because it
exposes builder's DiagnosticMessageChain.next contract errors.


The merged batch total is pending: adaptation 71 rejects the createTypeChecker
runtime fingerprint after adaptation 41. Rows above retain the source branches'
individual measurements; the combined pipeline has not completed.
