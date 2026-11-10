# Paired latent census: overload callback use

Task 8n7eet3, Optimizer request, November 10, 2026. Report only.

Compiler parent: `1d84074f4cf6d947f7b0765997886c2c93d57cf2`. Compiler head: `8fabe4c2918fbfdf57398cf1fd6c2fb33f819ce4`. The optimizer branch was checked before work and still pointed at this head. It had not moved. Each binary uses its own SHA and guarded latent overlay, extended by its own `stage3/meter/entry_overlay.py`. No compiler or adaptation source was changed.

One adapted tree was built once by parent `1d84074f` using `stage3/apply.sh`, from TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Both sides use those same adapted bytes. `source-hashes.json` freezes all 747 source files; `patch-set.md`, `provenance.json`, `overlay-hashes.json`, and the raw child commands record reproduction inputs. CoHere and its checker dependency match on both sides.

Both entry drivers require `LATENT_ASSERT_NO_OUTPUT=1`, check that ordinary Load and Lower cannot expose output, and import no backend. Measurements concern independent lowering units, not native behavior or whole-program compilation. Deferred generic declarations are retained as deferred, not counted as successful standalone lowerings.

## Run resources

Wall time and process peak RSS come from each census child's own completed resource runner. Both sides overlap on this host and use `GOMEMLIMIT=5GiB`; these are resource observations, not a speed benchmark.

| Run | Child exit | Wall seconds | User CPU seconds | System CPU seconds | Peak RSS KiB | Peak RSS GiB |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| debug-parent | 0 | 52.038 | 57.000 | 5.794 | 883276 | 0.842 |
| debug-head | 0 | 58.162 | 61.132 | 5.996 | 893136 | 0.852 |
| entry-parent | 0 | 3725.537 | 4222.926 | 99.075 | 3889268 | 3.709 |
| entry-head | 0 | 3645.131 | 4150.187 | 95.735 | 4084944 | 3.896 |

## debug.ts, full mode

Both streams load the same checker-clean tsc entry graph (81 source files), then select only debug.ts for full-mode observation. Both end with a relation trailer and child exit 0. There are 81 debug units on each side. At locations in debug.ts itself, raw events are 20 -> 12 and unique findings are 17 -> 11; the remaining two unique findings are in core.ts. The full event streams, including two unchanged dependency findings, contain 22 -> 14 findings; unique raw findings and distinct reads are both 19 -> 13, including Boundary markers. Refused findings are 3 -> 1, NotYet findings remain 5, and Boundary markers are 11 -> 7. Two callback refusal locations and four continuation Boundary markers clear. There are no moved reasons, new findings, new sites, panics or regressions in this full-mode pair. Four attempted units become finding-free.

The predicted direction holds: assertNode and assertOptionalNode clear their callback refusal, while assertEachNode keeps it (its body passes test on to every). The observed callback locations are 273:82 and 299:78, each emitted in two attempting-unit contexts. This full census therefore observes two distinct callback locations clearing, rather than three. The source has three narrowing callback overload signatures across these two functions (one assertNode and two assertOptionalNode), but the overload-group census returns at the first failed overload, so the baseline does not emit a separate overload-2 diagnostic for assertOptionalNode. The source evidence is retained in parent-census-small.go.txt and head-census-small.go.txt. Generic signature declarations remain deferred and are not counted as independently lowered bodies.

Every assert-family finding site follows. Boundary rows are the meter's continuation markers, retained to make the raw pairing fully accountable. Locations are relative to src/compiler/debug.ts; reason strings are quoted in full.

| Function | Site | Parent | Head |
| --- | --- | --- | --- |
| assertNever | 252:5 | "Boundary: function signature or prologue failed; continued at next declaration" | "Boundary: function signature or prologue failed; continued at next declaration" |
| assertNever | 252:33 | "NotYet: a value of type never" | "NotYet: a value of type never" |
| assertEachNode | 257:86 | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" |
| assertEachNode | 261:5 | "Boundary: function signature or prologue failed; continued at next declaration" | "Boundary: function signature or prologue failed; continued at next declaration" |
| assertEachNode | 262:5 | "Boundary: function signature or prologue failed; continued at next declaration" | "Boundary: function signature or prologue failed; continued at next declaration" |
| assertNode | 273:82 | "Refused: overload 1 of assertNode parameter test cannot be served by implementation parameter test" | None |
| assertNode | 274:5 | "Boundary: function signature or prologue failed; continued at next declaration" | None |
| assertNode | 275:5 | "Boundary: function signature or prologue failed; continued at next declaration" | None |
| assertOptionalNode | 299:78 | "Refused: overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test" | None |
| assertOptionalNode | 301:5 | "Boundary: function signature or prologue failed; continued at next declaration" | None |
| assertOptionalNode | 302:5 | "Boundary: function signature or prologue failed; continued at next declaration" | None |

Every assert-family declaration, including finding-free siblings and deferred signatures:

| Function | Declaration | Eligibility, parent -> head | First non-Boundary stop, parent | First non-Boundary stop, head |
| --- | --- | --- | --- | --- |
| assert | 192:5 | attempted -> attempted | None | None |
| assertEqual | 202:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertLessThan | 209:5 | attempted -> attempted | None | None |
| assertLessThanOrEqual | 215:5 | attempted -> attempted | None | None |
| assertGreaterThanOrEqual | 221:5 | attempted -> attempted | None | None |
| assertIsDefined | 227:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachIsDefined | 239:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachIsDefined | 240:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachIsDefined | 241:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertNever | 252:5 | attempted -> attempted | "NotYet: a value of type never" | "NotYet: a value of type never" |
| assertEachNode | 257:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachNode | 258:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachNode | 259:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachNode | 260:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertEachNode | 261:5 | attempted -> attempted | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" |
| assertEachNode | 262:5 | attempted -> attempted | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" | "Refused: overload 1 of assertEachNode parameter test cannot be served by implementation parameter test" |
| assertNode | 273:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertNode | 274:5 | attempted -> attempted | "Refused: overload 1 of assertNode parameter test cannot be served by implementation parameter test" | None |
| assertNode | 275:5 | attempted -> attempted | "Refused: overload 1 of assertNode parameter test cannot be served by implementation parameter test" | None |
| assertNotNode | 286:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertNotNode | 287:5 | attempted -> attempted | None | None |
| assertNotNode | 288:5 | attempted -> attempted | None | None |
| assertOptionalNode | 299:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertOptionalNode | 300:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertOptionalNode | 301:5 | attempted -> attempted | "Refused: overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test" | None |
| assertOptionalNode | 302:5 | attempted -> attempted | "Refused: overload 1 of assertOptionalNode parameter test cannot be served by implementation parameter test" | None |
| assertOptionalToken | 313:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertOptionalToken | 314:5 | deferred_generic_instantiation -> deferred_generic_instantiation | None | None |
| assertOptionalToken | 315:5 | attempted -> attempted | None | None |
| assertOptionalToken | 316:5 | attempted -> attempted | None | None |
| assertMissingNode | 327:5 | attempted -> attempted | None | None |
| assertMissingNode | 328:5 | attempted -> attempted | None | None |

Other debug.ts sites that move: **none**. The unchanged non-assert debug findings are listed here for completeness.

| Function | Site | Parent | Head |
| --- | --- | --- | --- |
| formatSymbol | 350:9 | "Boundary: failed statement; state rolled back; continued at next statement or declaration" | "Boundary: failed statement; state rolled back; continued at next statement or declaration" |
| formatSymbol | 350:135 | "NotYet: a template interpolating an object, an array, a map, a function or undefined" | "NotYet: a template interpolating an object, an array, a map, a function or undefined" |
| getEnumMembers | 400:13 | "Boundary: failed statement; state rolled back; continued at next statement or declaration" | "Boundary: failed statement; state rolled back; continued at next statement or declaration" |
| getEnumMembers | 400:27 | "NotYet: a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls)" | "NotYet: a readonly view read by a run-time key in a program with a class getter that does more than return its fields and literals joined by operators (the getter runs inside the read, unseen by the passes that follow calls)" |
| getEnumMembers | 401:13 | "Boundary: failed statement; state rolled back; continued at next statement or declaration" | "Boundary: failed statement; state rolled back; continued at next statement or declaration" |
| getEnumMembers | 401:24 | "NotYet: reading value" | "NotYet: reading value" |

Full-mode dependency findings: the comparator finding at core.ts:1049:82 and its Boundary at core.ts:1049:5 are unchanged. A debug-only observation can report a dependency stop; these are not additional changed debug.ts sites.

The exact access-chain pair output is in [debug-paired.json](debug-paired.json) and [debug-paired.md](debug-paired.md).

## tsc entry, first-error mode

Both first-error streams load the same checker-clean tsc entry and resolved 81-file graph. LATENT_FULL, LATENT_SPECULATIVE and LATENT_ONLY_FILE are unset. Both ledgers have 4,562 units: 4,058 attempted and 504 deferred generic declarations, with no checker-body skips. The raw event count is 5,758 on each side; raw unique findings are 4,997 -> 4,996 and distinct findings are 4,795 -> 4,794. Both relation-overflow endpoint checks are empty. The only moved unit remains blocked: its new binder stop was already present in the parent stream under other attempting units, so it introduces no new global finding. The parent first-error stream never reaches an assertOptionalNode callback refusal; the full debug pair above is the evidence for that function. The pairing tool verified equal headers, unit identities and eligibility, and normalized locations to the outermost contiguous property/element access chain.

A cleared finding below leaves no remaining finding at that identity. A changed reason is a removed reason with another reason still at the identity. New findings include exposed reasons at already-blocked sites; new sites have no parent finding. Changed first stops in already-blocked units are reported separately from regressions.

| Measure | Raw | Distinct |
| --- | ---: | ---: |
| Cleared, no remaining finding | 1 | 1 |
| Changed reason, still blocked | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| All unit first-stop changes | 1 | 1 |
| Regressions | 0 | 0 |
| All removed findings/reasons | 1 | 1 |
| Already-blocked units with a changed first stop | 1 | 1 |
| Previously blocked units now finding-free | 0 | 0 |

| Finding kind | Raw parent | Raw head | Distinct parent | Distinct head |
| --- | ---: | ---: | ---: | ---: |
| NotYet | 535 | 535 | 528 | 528 |
| Refused | 4462 | 4461 | 4267 | 4266 |

Every unit whose first stop changes, with whole reasons:

| Unit | Declaration | Parent first stop | Head first stop | Classification |
| --- | --- | --- | --- | --- |
| flattenDestructuringBinding | src/compiler/transformers/destructuring.ts:244:1 | "Refused: overload 1 of assertNode parameter test cannot be served by implementation parameter test" at src/compiler/debug.ts:273:82 | "NotYet: writing a checked object field without its source-slot type certificate" at src/compiler/binder.ts:1109:17 | already blocked, changed stop |

Outside-debug units with a moving overload-parameter reason (all quoted whole, including newly exposed stops):

- `src/compiler/transformers/destructuring.ts:244:1` (`flattenDestructuringBinding`): "Refused: overload 1 of assertNode parameter test cannot be served by implementation parameter test" at src/compiler/debug.ts:273:82 -> "NotYet: writing a checked object field without its source-slot type certificate" at src/compiler/binder.ts:1109:17.

### Regressions

A regression is a previously finding-free attempted unit that now stops, a newly finding-bearing scan site, or a new panic. Exposed first stops in already-blocked units are excluded.

Regression conclusions use the normalized distinct-read identities. The raw column retains unnormalized candidate counts from the pairing tool.

**None.**

Panics: 0 parent, 0 head. The exact cleared, changed, new and regression records are retained in [entry-paired.json](entry-paired.json) and [entry-paired.md](entry-paired.md).

## Evidence, checks and checkpoints

The 18 existing access-chain pairing tests passed. Both production-tree comparisons are clean in the compiler, adapter, meter and pairing directories. Every completed census must have child exit 0, a measurement header, zero checker diagnostics and a final relation-report trailer; [verification.json](verification.json) records the final recomputation of adapted source and archive hashes.

The initial environment transport reconnect ended the original outer setup/build/apply wrappers. The builds and apply continued with unchanged arguments and their own exit 0 was recovered from Linux kernel wait status by adopt.py. Setup was retried and completed with exit 0. A second transport reconnect during the entry runs left both detached supervisors and their children alive with unchanged commands; entry-reconnect.json records their identities. Only the preparation statuses use recovered process statistics; all four census runs use direct-child resource measurement under detached supervisors. A 30-minute period without log or watched-file progress stops a child process group. No stopped run is treated as complete.

Automatic approval review initially rejected the apply checkpoint push as an unverified separate repository. Inspection established that it targets another branch of the same authorized origin, https://github.com/system-inc/adamic.git. The explicit same-repository retry was approved and succeeded.

Checkpoint snapshots are on `typescript/overload-callback-census-checkpoints` in that same repository. The final requested branch `typescript/overload-callback-census` has one report-only commit directly on `1d84074f`, touching only this directory. No main push or force-push is used. [checkpoints.json](checkpoints.json) records all five pushed snapshots, and push logs are retained beside the report. [process-cleanup.json](process-cleanup.json) confirms that no measurement, supervisor or recovery process remains active.

Raw streams are compressed with deterministic gzip even when below 20 MB. The following hashes cover the original decompressed stream and the retained archive; each run's `*-streams.json` also records timing-stream hashes.

| Raw stream archive | Original bytes | Original SHA-256 | Archive bytes | Archive SHA-256 |
| --- | ---: | --- | ---: | --- |
| [debug-parent.jsonl.gz](debug-parent.jsonl.gz) | 119642 | `ee10e7a35fd871c30951159541f935ce1fa84e6de40e8b8f00f531dad9f73835` | 8532 | `7c5ecccd0abbe957cee2781614699634ab749b913e499dbb1dac64a40dc645ef` |
| [debug-head.jsonl.gz](debug-head.jsonl.gz) | 114414 | `a0e6262afe5811e307400fdee1931b01a6244db93724419d92e5377175cc65b0` | 8390 | `b877711b3944cc0db960e032fa550ffbad5dc671be1cc31929d5746ec9c60648` |
| [entry-parent.jsonl.gz](entry-parent.jsonl.gz) | 17573179 | `0f2d8558a320b7648ab31f0bc872c65a9a1b1c54f6d2f429bd007b79ed8ec354` | 1234131 | `03280df3dc421c51ea4ea4f4997681c1e24a5ec608fadab23bc72bb7abad0865` |
| [entry-head.jsonl.gz](entry-head.jsonl.gz) | 17573041 | `330ffb01b33f41d68e1f7f557af6d4c1085b83d91282c70e9c7595907398464a` | 1234120 | `83a78266597fcc54aff8e1f4f2939e81845729b82d5f875561762b2649a9df49` |

[commands.md](commands.md) records reproduction commands. Exact census commands, flags, exits and resources are in the four `*.status.json` files. Setup used `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh`, `npm ci --prefix stage3/api`, and `ADAMIC_TYPESCRIPT_SOURCE=/workspace/overload-typescript` at the required upstream pin. The parent applied once to `/workspace/overload-adapted`. Each compiler then generated its own latent overlay and entry extension and built with `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`.
