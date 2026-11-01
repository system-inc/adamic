# 96c: generators.ts's required indexed reads

Axolotl (#axolotl), task #vy7ncrg. Byte-identical in JavaScript, with no public API change. The adapter edits only src/compiler/transformers/generators.ts.

Adaptation 33 owns transformers/ and declined these 19 reads with "its invariant review remains unfinished" (33's sites.json). This finishes that review.

- **The edit:** each read gets a `!` at its own evaluation point, 33's method.
- **On Node:** the `!` erases.
- **In Adamic:** it's a loud required-value check, so a native build stops where Node would carry `undefined` on. That changes behavior only when the invariant below is broken. None is broken by any program tsc's parser produces.

## The sites, by invariant

| Function | Read | Invariant |
| --- | --- | --- |
| transformAndEmitStatements | `statements[i]` | U-loop: `i` runs from `start` to `statements.length - 1` over a readonly `NodeArray`, which has no holes. |
| transformAndEmitVariableDeclarationList | `variables[i]` | U-loop: `i < numVariables`, which is `variables.length`, over a `NodeArray`. |
| transformAndEmitForInStatement, visitForInStatement | `initializer.declarations[0]` | U-endpoint: a for-in head's `VariableDeclarationList`, which the grammar requires to hold one declaration. Upstream already reads `[0].name` unguarded, so an empty list, if error recovery ever made one, throws on Node too. The `!` makes that same failure loud, never a new one. |
| transformAndEmitSwitchStatement | `caseBlock.clauses[i]`, three reads | U-loop: each loop is bounded by `numClauses`, which is `caseBlock.clauses.length`, over a `NodeArray`. |
| transformAndEmitSwitchStatement | `clauseLabels[i]`, two reads | U-parallel: the first loop pushes exactly one label for each clause, before any read, so `clauseLabels.length === numClauses`. |
| transformAndEmitSwitchStatement | `clauseLabels[defaultClauseIndex]` | U-position: read only under `defaultClauseIndex >= 0`, and that index was a clause index from the same first loop. |
| hasImmediateContainingLabeledBlock | `blockStack![j]` | U-loop: `j` runs down from `start` to 0. Both callers pass `i - 1` for an `i` below `blockStack.length`, and `blockStack` is built by `push`. |
| findBreakTarget, findContinueTarget | `blockStack[i]`, two reads each | U-loop: `i` runs from `blockStack.length - 1` down to 0, under `if (blockStack)`, over a push-built stack. |
| appendLabel | `withBlockStack[i]` | U-loop: `i` runs from `withBlockStack.length - 1` down to 0, under `if (withBlockStack)`, over a push-built stack. |
| tryEnterLabel | `labelNumbers[labelNumber]`, third read | U-grid: the else branch of the same slot's `=== undefined` test, whose then branch allocates it. Only this read is asserted. The test and the store stay as they are. |
| tryEnterOrLeaveBlock | `blockOffsets![blockIndex]`, `blocks[blockIndex]` | U-parallel: beginBlock and endBlock write `blockActions`, `blockOffsets` and `blocks` at the same index together (2142-2144, 2157-2159), and the loop is bounded by `blockActions.length`. |

The three TS2339 `.expression` errors in transformAndEmitSwitchStatement (1881, 1887, 1889) are the `clause.kind === SyntaxKind.CaseClause` narrowing, which a possibly absent `clause` blocks. They go with the `!` on `caseBlock.clauses[i]`.

## Proof

- **The shape:** the adapter parses with stock TypeScript 6.0.3. It finds each read by its function, its exact text and its count in that function, and refuses any other count. It never asserts a store, and refuses a half-adapted file. It plans every edit before its single write.
- **Diagnostics:** `adamic types` on the adapted tree, on 96a's tree at main a23ba096:
  - whole program, 140 before and 113 after
  - generators.ts own-file, 27 before and 0 after
  - The 27 removed are exactly its own, with no new diagnostic anywhere.
- **JavaScript:** `ts.transpileModule` of generators.ts before and after gives byte-identical output (100,215 bytes each).
- **Idempotence:** a second run asserts nothing.

**Lane:** run as part of Correctness's wave, which carries the manifest regeneration. The gap: no lane run before the wave's own, since no box could run one on Nov 1. Loom's run of the wave is the proof.
