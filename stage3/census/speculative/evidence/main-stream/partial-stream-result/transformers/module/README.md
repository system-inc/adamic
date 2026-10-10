Built a census-only speculative continuation overlay with typed placeholders and failure depth.
The compiler base and mapper equivalence are recorded in RESULT.json; census changes stay inside its overlay.
Measured TypeScript 6.0.3 on the pinned adapted compiler corpus; tables below are observations.
Nested depth control, no-stubs baseline, depth mutant and output isolation checks are recorded in evidence.
Coverage means speculative AST examination; native semantics and successful compilation are outside this census.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 15 | 3 |
| 1 | 1 | 0 |
| 2 | 0 | 0 |
| 3 | 0 | 0 |
| 4+ | 0 | 0 |
| Total | 16 | 3 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | an assignment value to a member | 4 | 0 | 0 | 0 | 0 | 4 |
| NotYet | reading ModuleKind | 3 | 0 | 0 | 0 | 0 | 3 |
| NotYet | reading SyntaxKind | 3 | 0 | 0 | 0 | 0 | 3 |
| NotYet | reading isSourceFile | 3 | 0 | 0 | 0 | 0 | 3 |
| Refused | an object refinement using an open numeric enum as a literal tag | 3 | 0 | 0 | 0 | 0 | 3 |
| NotYet | a namespace member without a lowered binding | 1 | 0 | 0 | 0 | 0 | 1 |
| NotYet | reading Debug | 0 | 1 | 0 | 0 | 0 | 1 |
| NotYet | reading map | 1 | 0 | 0 | 0 | 0 | 1 |

Examined **3,750 / 219,710 TypeScript source bytes (1.706795%)**, in 1 files. Including non-TypeScript inputs, that is **3,750 / 219,710 bytes (1.706795%)**. Every AST child in every completed TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined inputs:

- `esnextAnd2015.ts`: 18,686 bytes; per-file measurement incomplete; see stream inventory.
- `module.ts`: 111,793 bytes; per-file measurement incomplete; see stream inventory.
- `system.ts`: 85,481 bytes; per-file measurement incomplete; see stream inventory.

Stock TypeScript independently matched the ancestry depth of 19 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 0 findings had no compiler-source AST identity. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: none. See method.md for their interpretation.
