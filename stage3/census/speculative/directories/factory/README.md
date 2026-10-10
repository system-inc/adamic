Built a census-only speculative continuation overlay with typed placeholders and failure depth.
Base: `a5630a90`, merging `69501280` into `ed6e2975`; census changes stay inside its overlay.
Measured TypeScript 6.0.3 on the pinned adapted compiler corpus; tables below are observations.
Nested depth control, no-stubs baseline, depth mutant and output isolation checks are recorded in evidence.
Coverage means speculative AST examination; native semantics and successful compilation are outside this census.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 3,182 | 642 |
| 1 | 880 | 270 |
| 2 | 126 | 18 |
| 3 | 22 | 4 |
| 4+ | 5 | 11 |
| Total | 4,215 | 945 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | an assignment value to a member | 953 | 56 | 16 | 1 | 0 | 1,026 |
| NotYet | reading SyntaxKind | 787 | 34 | 33 | 0 | 0 | 854 |
| NotYet | a generic function as a value | 501 | 192 | 2 | 0 | 0 | 695 |
| Refused | an object refinement using an open numeric enum as a literal tag | 141 | 209 | 1 | 1 | 1 | 353 |
| NotYet | reading TransformFlags | 33 | 228 | 9 | 0 | 0 | 270 |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on node) | 244 | 0 | 0 | 0 | 0 | 244 |
| NotYet | a union of differently held members variable a function value captures | 84 | 55 | 0 | 0 | 0 | 139 |
| NotYet | a BinaryExpression with a value and a value | 87 | 0 | 0 | 0 | 0 | 87 |
| NotYet | a namespace member without a lowered binding | 55 | 9 | 3 | 0 | 0 | 67 |
| NotYet | reading Debug | 0 | 55 | 9 | 1 | 0 | 65 |
| Refused | a cast the runtime can't check | 27 | 15 | 6 | 2 | 6 | 56 |
| NotYet | a boolean &#124; undefined variable a function value captures | 38 | 15 | 0 | 0 | 0 | 53 |
| NotYet | a value of type T | 8 | 38 | 0 | 0 | 0 | 46 |
| NotYet | a function returning T | 40 | 0 | 0 | 0 | 0 | 40 |
| Refused | inherited library member push read as an own field | 36 | 0 | 0 | 0 | 0 | 36 |
| NotYet | reading ModifierFlags | 25 | 2 | 0 | 0 | 0 | 27 |
| NotYet | reading NodeFlags | 10 | 16 | 0 | 1 | 0 | 27 |
| NotYet | reading EmitFlags | 19 | 4 | 1 | 0 | 0 | 24 |
| NotYet | a value of type T["kind"] | 17 | 6 | 0 | 0 | 0 | 23 |
| NotYet | reading GeneratedIdentifierFlags | 23 | 0 | 0 | 0 | 0 | 23 |

Examined **581,593 / 581,593 TypeScript source bytes (100.000000%)**, in 10 files. Including non-TypeScript inputs, that is **581,593 / 581,593 bytes (100.000000%)**. Every AST child in every measured TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined non-TypeScript inputs:


Stock TypeScript independently matched the ancestry depth of 5,171 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 0 findings had no compiler-source AST identity. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: {"error": 3}. See method.md for their interpretation.
