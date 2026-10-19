Built a census-only speculative continuation overlay with typed placeholders and failure depth.
The compiler base and mapper equivalence are recorded in RESULT.json; census changes stay inside its overlay.
Measured TypeScript 6.0.3 on the pinned adapted compiler corpus; tables below are observations.
Nested depth control, no-stubs baseline, depth mutant and output isolation checks are recorded in evidence.
Coverage means speculative AST examination; native semantics and successful compilation are outside this census.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 8 | 0 |
| 1 | 10 | 0 |
| 2 | 2 | 0 |
| 3 | 0 | 0 |
| 4+ | 0 | 0 |
| Total | 20 | 0 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | a function returning T | 5 | 2 | 0 | 0 | 0 | 7 |
| NotYet | a value of type T | 0 | 5 | 0 | 0 | 0 | 5 |
| NotYet | a union of differently held members variable a function value captures | 0 | 2 | 2 | 0 | 0 | 4 |
| NotYet | a generic function as a value | 2 | 0 | 0 | 0 | 0 | 2 |
| NotYet | a function returning U | 1 | 0 | 0 | 0 | 0 | 1 |
| NotYet | a value of type U | 0 | 1 | 0 | 0 | 0 | 1 |

Examined **406 / 591 TypeScript source bytes (68.697124%)**, in 2 files. Including non-TypeScript inputs, that is **406 / 591 bytes (68.697124%)**. Deferred generic body bytes: 185, neither examined nor blocked. Never reached (unexamined coverage gaps): [{'where': '/workspace/adamic/review/compiler/loom-census-deferral/deferred/source/a.a:2:1', 'name': 'unused', 'body_start': 68, 'body_end': 131}, {'where': '/workspace/adamic/review/compiler/loom-census-deferral/deferred/source/a.a:4:1', 'name': 'diagnosed', 'body_start': 230, 'body_end': 277}, {'where': '/workspace/adamic/review/compiler/loom-census-deferral/deferred/source/a.a:2:42', 'name': 'unusedNested', 'body_start': 97, 'body_end': 115}]. Every AST child in every completed TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined inputs:


Stock TypeScript independently matched the ancestry depth of 20 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 0 findings had no compiler-source AST identity. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: none. See method.md for their interpretation.
