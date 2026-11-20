# Paired Stage 3 census for #ad161ms

Base `df1c19ae40b5fc5f6eed5cb7daa5c92841674e83` is the exact parent of head `b8f6109bf1bc284de5b3dd0b446c26b2053cadfd`. Both refs were fetched by branch name. Fetched `main` was at `b576987b23511ff4e3fceab0a98cacd563bdc1f1`.

## 1. New findings

| Latent pass | New | Gone | Changed | Reached |
| --- | ---: | ---: | ---: | ---: |
| tsc entry | 0 | 0 | 1 | 0 |
| compiler directory | 0 | 0 | 1 | 0 |

Zero new findings under any reason in either pass.

## 2. Gone, changed and reached findings

No findings disappeared in either pass.

### Changed in tsc entry and compiler directory

Base reports `Refused` in `refusal_scan`, unit `transformESDecorators` at `src/compiler/transformers/esDecorators.ts:296:1`, finding at `src/compiler/transformers/esDecorators.ts:1307:49`.

> Adamic 0.1 refuses a value of type { referencedName: StringLiteral; name: PropertyName; } | { referencedName: Identifier; name: ComputedPropertyName; } seen as { referencedName: Expression | undefined; name: PropertyName | undefined; }, which can write Expression | undefined where StringLiteral is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)

Head reports `Refused` in `refusal_scan`, unit `transformESDecorators` at `src/compiler/transformers/esDecorators.ts:296:1`, finding at `src/compiler/transformers/esDecorators.ts:1307:49`.

> Adamic 0.1 refuses a value of type { referencedName: StringLiteral; name: ComputedPropertyName | Identifier | NumericLiteral | PrivateIdentifier | StringLiteral; } | ... seen as { referencedName: Expression | undefined; name: PropertyName | undefined; }, which can write Expression | undefined where StringLiteral is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)

The diagnostic is at the call to `visitReferencedPropertyName(member.name)`. That helper passes `isPropertyName` as the test to `visitNode` at `src/compiler/transformers/esDecorators.ts:1930` and `:1936`. The narrowed predicate changes the printed type of `name`. Both messages identify the same invariant-mutable write concern. The head message contains the checker’s `...` elision.

Neither pass exposed a later stop after an earlier one cleared.

Each pass recorded the same 730 family-1 findings on both trees. The reason is “an object refinement using an open numeric enum as a literal tag.”

## 3. Units that lowered at base and stop at head, and new panics

| Latent pass | Newly stopped units | New tail stops in units that lowered at base | New panics |
| --- | ---: | ---: | ---: |
| tsc entry | 0 | 0 | 0 |
| compiler directory | 0 | 0 | 0 |

## 4. First stops across all roots

0 of 79 roots moved to a different first stop.

| Roots | First stop on both trees |
| ---: | --- |
| 73 | `NotYet` at `src/compiler/tracing.ts:76:15`. a read of mode narrowed to "project" \| "server" where a call may have rewritten it (mode is assigned inside startTracing). |
| 2 | Accepted by lowering. `src/compiler/corePublic.ts`, `src/compiler/hostErrors.ts`. |
| 1 | `NotYet` at `src/compiler/semver.ts:71:9`. a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration. |
| 1 | `NotYet` at `src/compiler/commandLineParser.ts:293:35`. reading an enum before its runtime initialization; move the call after the enum declaration. |
| 1 | `NotYet` at `src/compiler/diagnosticInformationMap.generated.ts:13:45`. reading an enum before its runtime initialization; move the call after the enum declaration. |
| 1 | `Refused` at `src/compiler/utilitiesPublic.ts:779:16` → `src/compiler/utilitiesPublic.ts:780:16`. a cast the runtime can't check. |

The `tracing.ts` root reports the same cast after the import edits shifted its line by one. The table uses the ordinary per-root census from the Stage 3 next-stop reports. The scan-all binary was built from base. Its continuation probes were stopped after spending several minutes on the first root. This table covers the ordinary first-stop boundary. [Full paired 79-root table](evidence/next-stop-table.json.gz) · [Prior Stage 3 census](https://github.com/system-inc/adamic/blob/df1c19ae40b5fc5f6eed5cb7daa5c92841674e83/stage3/adapt/99-correctness-axolotl-q/CENSUS.md).

## Inputs and evidence

Each detached checkout ran `git submodule update --init --recursive` and its own `stage3/apply.sh`. Both used the meter’s shallow tag clone of TypeScript `v6.0.3` at `050880ce59e30b356b686bd3144efe24f875ebc8`. Cohere was pinned to `5360753e19e4417675a90e92a1823bea66b13918` and its TypeScript submodule to `d92d9bfee114c80be2c375d72edae966176e3a4f`. Comparing every tracked path and file byte in the completed adapted trees found one changed file, `src/compiler/utilitiesPublic.ts`. The generated patch-set summaries record each tree’s own adaptations.

Both latent binaries and the ordinary census binary were built once from base. Both latent passes set `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1` and `CENSUS_MODULE_FORMAT=CommonJSBundle`. The entry binary used `src/tsc/tsc.ts`. The directory binary used `src/compiler`. The ordinary census used `CENSUS_MODULE_FORMAT=CommonJSBundle`.

Pairing used the location and access-chain normalization from `stage3/census/pair` on each tree separately. The comparison mapped corresponding lines in the edited file and paired by kind, phase, reason and message with positions stripped. For findings that kept their unit and site, changes in the diagnostic words are recorded under changed. A head-only stop reached after an earlier base stop cleared is recorded under reached.

| Pass | Base | Head | Pair |
| --- | --- | --- | --- |
| tsc entry latent | [Stream](evidence/base/entry.jsonl.gz) · [Times](evidence/base/entry-times.jsonl.gz) | [Stream](evidence/head/entry.jsonl.gz) · [Times](evidence/head/entry-times.jsonl.gz) | [Delta](evidence/entry-pair.json.gz) |
| Compiler-directory latent | [Stream](evidence/base/compiler.jsonl.gz) · [Times](evidence/base/compiler-times.jsonl.gz) | [Stream](evidence/head/compiler.jsonl.gz) · [Times](evidence/head/compiler-times.jsonl.gz) | [Delta](evidence/compiler-pair.json.gz) |
| Ordinary next-stop census | [Stream](evidence/base/next-stop.jsonl.gz) | [Stream](evidence/head/next-stop.jsonl.gz) | [79-root table](evidence/next-stop-table.json.gz) |
