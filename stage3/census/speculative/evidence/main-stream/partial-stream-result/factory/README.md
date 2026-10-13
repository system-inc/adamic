Built a census-only speculative continuation overlay with typed placeholders and failure depth.
The compiler base and mapper equivalence are recorded in RESULT.json; census changes stay inside its overlay.
Measured TypeScript 6.0.3 on the pinned adapted compiler corpus; tables below are observations.
Nested depth control, no-stubs baseline, depth mutant and output isolation checks are recorded in evidence.
Coverage means speculative AST examination; native semantics and successful compilation are outside this census.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 524 | 327 |
| 1 | 80 | 8 |
| 2 | 8 | 4 |
| 3 | 0 | 1 |
| 4+ | 0 | 0 |
| Total | 612 | 340 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | reading SyntaxKind | 316 | 1 | 0 | 0 | 0 | 317 |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on node) | 228 | 0 | 0 | 0 | 0 | 228 |
| Refused | an object refinement using an open numeric enum as a literal tag | 59 | 4 | 0 | 0 | 0 | 63 |
| NotYet | a generic function as a value | 29 | 5 | 0 | 0 | 0 | 34 |
| NotYet | an assignment value to a member | 5 | 27 | 1 | 0 | 0 | 33 |
| NotYet | a value of type T | 1 | 23 | 0 | 0 | 0 | 24 |
| NotYet | a function returning T | 23 | 0 | 0 | 0 | 0 | 23 |
| NotYet | reading identity | 19 | 0 | 0 | 0 | 0 | 19 |
| NotYet | reading cast | 14 | 0 | 0 | 0 | 0 | 14 |
| NotYet | reading skipPartiallyEmittedExpressions | 10 | 0 | 0 | 0 | 0 | 10 |
| NotYet | reading notImplemented | 9 | 0 | 0 | 0 | 0 | 9 |
| NotYet | a namespace member without a lowered binding | 8 | 0 | 0 | 0 | 0 | 8 |
| NotYet | reading Debug | 0 | 8 | 0 | 0 | 0 | 8 |
| Refused | inherited library member get read as an own field | 4 | 1 | 1 | 1 | 0 | 7 |
| NotYet | a call through ?. (an optional call) | 4 | 1 | 0 | 0 | 0 | 5 |
| NotYet | new a ParenthesizedExpression | 5 | 0 | 0 | 0 | 0 | 5 |
| NotYet | reading isLeftHandSideExpression | 5 | 0 | 0 | 0 | 0 | 5 |
| NotYet | reading isNodeArray | 5 | 0 | 0 | 0 | 0 | 5 |
| NotYet | reading objectAllocator | 0 | 5 | 0 | 0 | 0 | 5 |
| NotYet | reading sameMap | 5 | 0 | 0 | 0 | 0 | 5 |

Examined **99,148 / 581,593 TypeScript source bytes (17.047660%)**, in 7 files. Including non-TypeScript inputs, that is **99,148 / 581,593 bytes (17.047660%)**. Every AST child in every completed TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined inputs:

- `emitHelpers.ts`: 69,184 bytes; per-file measurement incomplete; see stream inventory.
- `nodeFactory.ts`: 336,646 bytes; per-file measurement incomplete; see stream inventory.
- `utilities.ts`: 76,615 bytes; per-file measurement incomplete; see stream inventory.

Stock TypeScript independently matched the ancestry depth of 954 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 0 findings had no compiler-source AST identity. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: none. See method.md for their interpretation.
