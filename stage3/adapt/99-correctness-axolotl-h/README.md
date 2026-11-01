# 99h: honest config-formatting result types

JavaScript-changing in text only. formatSingleValue and formatValueOrArray explicitly return string | undefined. The two config templates consume their result with String, preserving their existing conversion of undefined to the text "undefined". The scalar helper forwards its optional result without an assertion or cast.

The array consumer is different: Array.join renders undefined as an empty field. The callback stays unchanged, and `String(value.map(v => formatSingleValue(v, map)).join(", "))` explicitly converts the joined text at interpolation. This preserves empty fields, the plugins toJSON witness, and custom map implementations that observe callback results. No cast or required-value assertion is added.

lane: run as part of Correctness's wave

Branch: typescript/axolotl-99h, from main 470e16f8315f7a83c5c7b7f6f2a77d8328db0a25. Source: stock TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Proof composition includes 99b a299eddff730973c769df5b13ff7d9fd747435a9 and 99g debc7962f9b3ea545ac9818cbe1a93c88b020f2e. Those prerequisite directories are temporary proof inputs, not changes committed on this branch. 99h requires 99g's exact formatter declaration and runs after it in a composed wave. No lane manifest changes.

## Exact changes

| Site in generateTSConfig | Before | After |
| --- | --- | --- |
| formatValueOrArray return declaration | string | string \| undefined |
| formatSingleValue return declaration | inferred string \| undefined | explicit string \| undefined |
| Commented config template | `${formattedValue}` | `${String(formattedValue)}` |
| Active config template | `${formattedValue}` | `${String(formattedValue)}` |
| Array text interpolation | `value.map(v => formatSingleValue(v, map)).join(", ")` | `String(value.map(v => formatSingleValue(v, map)).join(", "))` |

JSON.stringify returns a string or undefined. Both scalar forwarding and array formatting retain those possibilities honestly until the text consumer. For the stated domain, String on a string is the identity, and String(undefined) is exactly the old template interpolation. The array callback still returns exactly its original string or undefined to map. Join retains its empty-field rule, and String wraps the same joined text that the template previously consumed. Each helper still evaluates once per value. The 99g own-undefined literal-null branch is unchanged. No new handling of a toJSON hook, throw, map failure or ordinary scalar value is introduced.

## Adapter contract

Stock TypeScript 6.0.3 parses the source. The adapter finds the unique generateTSConfig, emitOption, formatValueOrArray and formatSingleValue declarations, checks their nesting, pins both helper headers, all two formatSingleValue calls, the unchanged scalar return, the exact array return and both templates. It requires the exact 99g formatter declaration and seven-statement emitter body. All five regions must be entirely before or entirely after. Missing 99g, half adaptation, changed spellings, moved calls and extra occurrences are refused before any write.

Every edit is planned first. The candidate result is parsed and validated again as fully adapted, and the original input is reread before one write. A second run makes zero edits and no write. The directory contains only adapt.cjs and this README; generated proofs never feed apply's input key.

## Proofs

Proof scripts, raw commands and results live on [typescript/axolotl-99h-evidence](https://github.com/system-inc/adamic/tree/typescript/axolotl-99h-evidence/stage3/evidence/99-correctness-axolotl-h). Logs are written under /workspace/scratch/99h. That evidence records the full apply runs, source scope, no-write second run, refusal fixtures, built-library witnesses and sequential unfiltered TypeScript oracle comparisons.

The measurement is main plus 99b plus 99g before, and that composition plus 99h after: **112 before and 111 after**, with exactly D036 removed and nothing added anywhere. Both checks run on fresh apply trees before upstream npm ci, as in the classification workflow. The inherited source-map-support module diagnostic is present in both complete counts and is not filtered. Upstream npm ci later supplies that dependency for the compiler builds and oracle. The earlier 99g count of 112 was measured after npm ci; comparing it directly with the fully composed fresh-tree count would conflate dependency phases. No checker option or source overlay changes.

The witnesses compare whole generated files, byte for byte, using the actual built before and after libraries. Cases include ordinary --init, --strict false, cleared outFile/strict/target/rootDir, a defined paths object whose toJSON returns undefined, and a plugin array element whose toJSON returns undefined, including a custom map that filters undefined callback results. The paths case still emits `"paths": undefined,`; the plugin case still emits `"plugins": [],`. Readback outcomes are unchanged. API controls include absent and inherited properties, null, falsy values and empty arrays. Thrown toJSON and invalid enum behavior are compared as well.

The complete unchanged TypeScript oracles run sequentially with four workers, all runners and no filter, as in 91's README. Only the existing sanctioned API-baseline failure is accepted. The proof checks exact counts, failure titles, API sanctions, full baseline-diff bytes, public declaration bytes and memory-event deltas. No test, baseline, oracle, expected count or API sanction is edited. A baseline move stops the proof rather than being accepted.

Reproduction uses fresh scratch destinations. Extract the exact prerequisite adaptation directories from the pinned branch commits into a disposable clone, build the before tree before creating 99h, then add this adaptation and build the after tree:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/99h/before > /workspace/scratch/99h/apply-before.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/99h/after > /workspace/scratch/99h/apply-after.log 2>&1
```

The evidence branch retains the exact preparation commands, source snapshots, checker invocations and full witness/oracle results. The runtime branch adds only this adaptation directory.
