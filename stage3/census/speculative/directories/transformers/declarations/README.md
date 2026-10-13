Built a census-only speculative continuation overlay with typed placeholders and failure depth.
Base: `a5630a90`, merging `69501280` into `ed6e2975`; census changes stay inside its overlay.
Measured TypeScript 6.0.3 on the pinned adapted compiler corpus; tables below are observations.
Nested depth control, no-stubs baseline, depth mutant and output isolation checks are recorded in evidence.
Coverage means speculative AST examination; native semantics and successful compilation are outside this census.

| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 346 | 23 |
| 1 | 13 | 8 |
| 2 | 58 | 0 |
| 3 | 0 | 0 |
| 4+ | 0 | 0 |
| Total | 417 | 31 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | reading Diagnostics | 103 | 0 | 29 | 0 | 0 | 132 |
| NotYet | reading SyntaxKind | 64 | 0 | 29 | 0 | 0 | 93 |
| NotYet | reading createDiagnosticForNode | 21 | 0 | 0 | 0 | 0 | 21 |
| NotYet | reading SymbolAccessibility | 17 | 0 | 0 | 0 | 0 | 17 |
| Refused | a cast the runtime can't check | 9 | 8 | 0 | 0 | 0 | 17 |
| NotYet | an ElementAccessExpression | 16 | 0 | 0 | 0 | 0 | 16 |
| NotYet | reading addRelatedInfo | 9 | 0 | 0 | 0 | 0 | 9 |
| NotYet | a namespace member without a lowered binding | 8 | 0 | 0 | 0 | 0 | 8 |
| NotYet | reading Debug | 0 | 8 | 0 | 0 | 0 | 8 |
| NotYet | reading isStatic | 8 | 0 | 0 | 0 | 0 | 8 |
| Refused | an object refinement using an open numeric enum as a literal tag | 8 | 0 | 0 | 0 | 0 | 8 |
| NotYet | reading getTextOfNode | 5 | 0 | 0 | 0 | 0 | 5 |
| NotYet | reading isExportAssignment | 5 | 0 | 0 | 0 | 0 | 5 |
| NotYet | reading isSetAccessor | 5 | 0 | 0 | 0 | 0 | 5 |
| NotYet | checked view field modifiers of type NodeArray&lt;ModifierLike&gt; &#124; undefined | 2 | 2 | 0 | 0 | 0 | 4 |
| NotYet | reading findAncestor | 4 | 0 | 0 | 0 | 0 | 4 |
| NotYet | reading isJSDocTypeAlias | 4 | 0 | 0 | 0 | 0 | 4 |
| NotYet | reading isConstructorDeclaration | 3 | 0 | 0 | 0 | 0 | 3 |
| NotYet | reading isGetAccessor | 3 | 0 | 0 | 0 | 0 | 3 |
| NotYet | reading isMethodDeclaration | 3 | 0 | 0 | 0 | 0 | 3 |

Examined **46,404 / 46,404 TypeScript source bytes (100.000000%)**, in 1 files. Including non-TypeScript inputs, that is **46,404 / 46,404 bytes (100.000000%)**. Every AST child in every measured TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined non-TypeScript inputs:


Stock TypeScript independently matched the ancestry depth of 450 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 1 findings had no compiler-source AST identity. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: {"error": 3}. See method.md for their interpretation.
