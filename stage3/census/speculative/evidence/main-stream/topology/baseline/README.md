Built a census-only speculative continuation overlay with typed placeholders and failure depth.
Base: `a5630a90`, merging `69501280` into `ed6e2975`; census changes stay inside its overlay.
Measured TypeScript 6.0.3 on the pinned adapted compiler corpus; tables below are observations.
Nested depth control, no-stubs baseline, depth mutant and output isolation checks are recorded in evidence.
Coverage means speculative AST examination; native semantics and successful compilation are outside this census.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 1 | 0 |
| 1 | 2 | 0 |
| 2 | 1 | 0 |
| 3 | 1 | 1 |
| 4+ | 2 | 0 |
| Total | 7 | 1 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | an array of T | 0 | 1 | 1 | 0 | 1 | 3 |
| NotYet | a value of type T | 0 | 1 | 0 | 0 | 1 | 2 |
| NotYet | a function returning T | 0 | 0 | 0 | 1 | 0 | 1 |
| NotYet | a function returning T &#124; undefined | 1 | 0 | 0 | 0 | 0 | 1 |
| Refused | inherited library member map read as an own field | 0 | 0 | 0 | 1 | 0 | 1 |

Examined **93 / 93 TypeScript source bytes (100.000000%)**, in 1 files. Including non-TypeScript inputs, that is **93 / 93 bytes (100.000000%)**. Every AST child in every measured TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined non-TypeScript inputs:


Stock TypeScript independently matched the ancestry depth of 8 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 0 findings had no compiler-source AST identity. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: none. See method.md for their interpretation.
