Built a census-only speculative continuation overlay with typed placeholders and failure depth.
The compiler base and mapper equivalence are recorded in RESULT.json; census changes stay inside its overlay.
Measured TypeScript 6.0.3 on the pinned adapted compiler corpus; tables below are observations.
Nested depth control, no-stubs baseline, depth mutant and output isolation checks are recorded in evidence.
Coverage means speculative AST examination; native semantics and successful compilation are outside this census.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 9,232 | 3,073 |
| 1 | 2,635 | 1,000 |
| 2 | 1,097 | 351 |
| 3 | 319 | 174 |
| 4+ | 195 | 87 |
| Total | 13,478 | 4,685 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | reading SyntaxKind | 2,038 | 196 | 90 | 1 | 0 | 2,325 |
| NotYet | a generic function as a value | 1,506 | 231 | 45 | 10 | 3 | 1,795 |
| NotYet | an overloaded function as a value | 802 | 105 | 44 | 5 | 3 | 959 |
| Refused | an object refinement using an open numeric enum as a literal tag | 684 | 148 | 13 | 5 | 2 | 852 |
| NotYet | reading CharacterCodes | 558 | 17 | 2 | 0 | 0 | 577 |
| NotYet | a union of differently held members variable a function value captures | 128 | 141 | 81 | 64 | 31 | 445 |
| NotYet | an assignment value to a member | 195 | 122 | 53 | 14 | 2 | 386 |
| NotYet | a PostfixUnaryExpression | 301 | 42 | 8 | 0 | 0 | 351 |
| Refused | a cast the runtime can't check | 131 | 107 | 41 | 24 | 31 | 334 |
| Refused | inherited library member push read as an own field | 213 | 75 | 22 | 15 | 3 | 328 |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on node) | 325 | 0 | 0 | 0 | 0 | 325 |
| Refused | an unproven predicate argument for parameter test (argument "isExpression") | 207 | 79 | 2 | 0 | 0 | 288 |
| NotYet | a value of type any | 126 | 75 | 56 | 13 | 15 | 285 |
| NotYet | a value of type T | 165 | 98 | 15 | 3 | 0 | 281 |
| NotYet | a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 107 | 109 | 43 | 11 | 9 | 279 |
| NotYet | a namespace member without a lowered binding | 115 | 149 | 6 | 2 | 0 | 272 |
| NotYet | a value of type Path | 135 | 61 | 23 | 10 | 3 | 232 |
| NotYet | a union or optional field read in a program with record storage | 144 | 48 | 9 | 1 | 1 | 203 |
| NotYet | this outside a method | 5 | 30 | 120 | 8 | 2 | 165 |
| NotYet | an array of T | 32 | 59 | 35 | 13 | 7 | 146 |

Examined **3,760,099 / 10,009,820 TypeScript source bytes (37.564102%)**, in 65 files. Including non-TypeScript inputs, that is **3,760,099 / 10,615,807 bytes (35.419813%)**. Every AST child in every completed TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined inputs:

- `binder.ts`: 195,107 bytes; per-file measurement incomplete; see stream inventory.
- `builder.ts`: 117,298 bytes; per-file measurement incomplete; see stream inventory.
- `checker.ts`: 3,154,156 bytes; per-file measurement incomplete; see stream inventory.
- `commandLineParser.ts`: 183,733 bytes; per-file measurement incomplete; see stream inventory.
- `diagnosticMessages.generated.json`: 306,771 bytes; non-TypeScript input; no lowering sites.
- `diagnosticMessages.json`: 299,061 bytes; non-TypeScript input; no lowering sites.
- `emitter.ts`: 274,627 bytes; per-file measurement incomplete; see stream inventory.
- `executeCommandLine.ts`: 54,177 bytes; per-file measurement incomplete; see stream inventory.
- `factory/nodeFactory.ts`: 336,646 bytes; per-file measurement incomplete; see stream inventory.
- `factory/utilities.ts`: 76,615 bytes; per-file measurement incomplete; see stream inventory.
- `moduleNameResolver.ts`: 183,618 bytes; per-file measurement incomplete; see stream inventory.
- `parser.ts`: 541,260 bytes; per-file measurement incomplete; see stream inventory.
- `program.ts`: 271,300 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/es2015.ts`: 231,789 bytes; per-file measurement incomplete; see stream inventory.
- `tsbuildPublic.ts`: 114,655 bytes; per-file measurement incomplete; see stream inventory.
- `tsconfig.json`: 155 bytes; non-TypeScript input; no lowering sites.
- `utilities.ts`: 514,740 bytes; per-file measurement incomplete; see stream inventory.

Stock TypeScript independently matched the ancestry depth of 20,558 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 1 findings had no compiler-source AST identity. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: {"error": 30}. See method.md for their interpretation.
