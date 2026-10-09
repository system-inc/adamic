Built a census-only speculative continuation overlay with typed placeholders and failure depth.
Base: `a5630a90`, merging `69501280` into `ed6e2975`; census changes stay inside its overlay.
Measured TypeScript 6.0.3 on the pinned adapted compiler corpus; tables below are observations.
Nested depth control, no-stubs baseline, depth mutant and output isolation checks are recorded in evidence.
Coverage means speculative AST examination; native semantics and successful compilation are outside this census.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 1,020 | 294 |
| 1 | 98 | 44 |
| 2 | 2 | 1 |
| 3 | 0 | 0 |
| 4+ | 0 | 0 |
| Total | 1,120 | 339 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| Refused | an object refinement using an open numeric enum as a literal tag | 119 | 33 | 1 | 0 | 0 | 153 |
| NotYet | reading SyntaxKind | 120 | 8 | 0 | 0 | 0 | 128 |
| NotYet | reading visitNode | 64 | 3 | 0 | 0 | 0 | 67 |
| NotYet | reading setTextRange | 49 | 2 | 0 | 0 | 0 | 51 |
| NotYet | reading isExpression | 40 | 2 | 0 | 0 | 0 | 42 |
| Refused | an unproven predicate argument for parameter test (argument "isExpression") | 40 | 2 | 0 | 0 | 0 | 42 |
| NotYet | reading append | 30 | 7 | 0 | 0 | 0 | 37 |
| Refused | inherited library member push read as an own field | 34 | 3 | 0 | 0 | 0 | 37 |
| NotYet | an assignment value to a member | 20 | 1 | 0 | 0 | 0 | 21 |
| NotYet | reading visitNodes | 20 | 1 | 0 | 0 | 0 | 21 |
| NotYet | reading EmitFlags | 20 | 0 | 0 | 0 | 0 | 20 |
| NotYet | reading ModuleKind | 17 | 3 | 0 | 0 | 0 | 20 |
| NotYet | reading visitEachChild | 20 | 0 | 0 | 0 | 0 | 20 |
| NotYet | reading isStatement | 19 | 0 | 0 | 0 | 0 | 19 |
| Refused | an unproven predicate argument for parameter test (argument "isStatement") | 19 | 0 | 0 | 0 | 0 | 19 |
| NotYet | a namespace member without a lowered binding | 16 | 1 | 0 | 0 | 0 | 17 |
| NotYet | reading Debug | 0 | 16 | 1 | 0 | 0 | 17 |
| NotYet | reading setOriginalNode | 16 | 1 | 0 | 0 | 0 | 17 |
| NotYet | a SpreadElement | 11 | 5 | 0 | 0 | 0 | 16 |
| NotYet | checked view field expression of type Expression | 16 | 0 | 0 | 0 | 0 | 16 |

Examined **219,710 / 219,710 TypeScript source bytes (100.000000%)**, in 4 files. Including non-TypeScript inputs, that is **219,710 / 219,710 bytes (100.000000%)**. Every AST child in every measured TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined non-TypeScript inputs:


Stock TypeScript independently matched the ancestry depth of 1,459 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 0 findings had no compiler-source AST identity. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: none. See method.md for their interpretation.
