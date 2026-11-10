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
| entry-parent | running | pending | pending | pending | pending | pending |
| entry-head | running | pending | pending | pending | pending | pending |

## debug.ts, full mode

Both streams load the same checker-clean tsc entry graph (81 source files), then select only debug.ts for full-mode observation. Both end with a relation trailer and child exit 0. There are 81 debug units on each side. At locations in debug.ts itself, raw events are 20 -> 12 and unique findings are 17 -> 11; the remaining two unique findings are in core.ts. The full event streams, including two unchanged dependency findings, contain 22 -> 14 findings; unique raw findings and distinct reads are both 19 -> 13, including Boundary markers. Refused findings are 3 -> 1, NotYet findings remain 5, and Boundary markers are 11 -> 7. Two callback refusal locations and four continuation Boundary markers clear. There are no moved reasons, new findings, new sites, panics or regressions in this full-mode pair. Four attempted units become finding-free.

The predicted direction holds: assertNode and assertOptionalNode clear their callback refusal, while assertEachNode keeps it. The observed callback locations are 273:82 and 299:78, each emitted in two attempting-unit contexts. This full census therefore observes two distinct callback locations clearing, rather than three. The source has three narrowing callback overload signatures across these two functions (one assertNode and two assertOptionalNode), but the overload-group census returns at the first failed overload, so the baseline does not emit a separate overload-2 diagnostic for assertOptionalNode. The source evidence is retained in parent-census-small.go.txt and head-census-small.go.txt. Generic signature declarations remain deferred and are not counted as independently lowered bodies.

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

The two entry runs are in progress. No completed entry delta or regression conclusion is available yet.

## Evidence, checks and checkpoints

The 18 existing access-chain pairing tests passed. Both production-tree comparisons are clean in the compiler, adapter, meter and pairing directories. Every completed census must have child exit 0, a measurement header, zero checker diagnostics and a final relation-report trailer; final verification also recomputes the adapted source and archive hashes.

The initial environment transport reconnect ended the original outer setup/build/apply wrappers. The builds and apply continued with unchanged arguments and their own exit 0 was recovered from Linux kernel wait status by adopt.py. Setup was retried and completed with exit 0. Only these preparation statuses use recovered process statistics; all four census runs use direct-child resource measurement under detached supervisors. A 30-minute period without log or watched-file progress stops a child process group. No stopped run is treated as complete.

Automatic approval review initially rejected the apply checkpoint push as an unverified separate repository. Inspection established that it targets another branch of the same authorized origin, https://github.com/system-inc/adamic.git. The explicit same-repository retry was approved and succeeded.

Checkpoint snapshots are on `typescript/overload-callback-census-checkpoints` in that same repository. The final requested branch `typescript/overload-callback-census` has one report-only commit directly on `1d84074f`, touching only this directory. No main push or force-push is used. Checkpoint push logs are retained beside the report.

Raw streams are compressed with deterministic gzip even when below 20 MB. The following hashes cover the original decompressed stream and the retained archive; each run's `*-streams.json` also records timing-stream hashes.

| Raw stream archive | Original bytes | Original SHA-256 | Archive bytes | Archive SHA-256 |
| --- | ---: | --- | ---: | --- |
| [debug-parent.jsonl.gz](debug-parent.jsonl.gz) | 119642 | `ee10e7a35fd871c30951159541f935ce1fa84e6de40e8b8f00f531dad9f73835` | 8532 | `7c5ecccd0abbe957cee2781614699634ab749b913e499dbb1dac64a40dc645ef` |
| [debug-head.jsonl.gz](debug-head.jsonl.gz) | 114414 | `a0e6262afe5811e307400fdee1931b01a6244db93724419d92e5377175cc65b0` | 8390 | `b877711b3944cc0db960e032fa550ffbad5dc671be1cc31929d5746ec9c60648` |

Exact census commands, flags, exits and resources are in the four `*.status.json` files. Setup used `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh`, `npm ci --prefix stage3/api`, and `ADAMIC_TYPESCRIPT_SOURCE=/workspace/overload-typescript` at the required upstream pin. The parent applied once to `/workspace/overload-adapted`. Each compiler then generated its own latent overlay and entry extension and built with `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`.
