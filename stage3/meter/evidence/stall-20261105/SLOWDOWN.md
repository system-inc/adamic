# Stage 3 meter slowdown, fq9ycej

The first bad commit for the selected slowdown predicate is **`b2b32fb8914d28e5ae911ec1dd8a475933f2b10c`**. Its immediate parent is `4a8171994b6a63a42962ddaa462aa28d8e9caa3d`. The regressing unit is `tracingEnabled.dumpTypes`, `src/compiler/tracing.ts:218:5`.

Across the 1,557 units timed by both 30-minute runs, elapsed unit time increased from **776.048 s to 1,324.296 s**, or **70.65%**. Snapshot time increased from 714.347 s to 732.757 s. The net snapshot increase is only 18.410 s of the 548.248 s total increase, **3.36%**. Most of the measured increase is new lowering proof work.

## Controlled endpoint runs

Ran `f87d7884807161162c98a2d0ef3b5339e16b315f` first, then `1f7df6a2e8145391dea19da63a172d5385638e39`, sequentially on this same box. Both binaries were built before either timed run. No other census or build ran during either measurement. The input was the same immutable adapted tree from the first investigation: TypeScript v6.0.3, upstream `050880ce59e30b356b686bd3144efe24f875ebc8`, prepared with apply at `1f7df6a2`. Compiler source and its exact pinned checker dependencies were the variables, not the adaptation. Both endpoints pin cohere `06af2e399c6c1fb93fd570b774027942747400b4`.

Each command used the meter's entry overlay, `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, and `LATENT_TIMES`, with the same entry `adapted/src/tsc/tsc.ts`. Each received a 1,800-second wall budget. SIGQUIT stopped each at that budget, with a hard-kill fallback. Recorded total lifetimes, including dump/exit cleanup, were 1,800.453 s and 1,800.424 s. The first process exited before the second started. The good run completed 2,282 timed units; the bad run completed 1,557. Neither completed the whole buffered census. Unfinished attempts are excluded, not assigned their cutoff time.

Two identical scratch-only observation hooks were added to the existing overlays. `prepare.py` wraps the unchanged `latentSnapshot` implementation with wall and process-CPU counters, and adds their per-unit deltas to the existing timer. It also enables `runtime/pprof` for exactly 60 seconds, beginning 600 seconds after process start. Both profiles completed: 60.01 seconds each. No lowering or recovery decision was changed. The extra counter and profiling overhead is included in the observed runs; this is not an uninstrumented benchmark.

Identity is the exact file, line and column in the fixed adapted source, normalized to a relative path. There are no duplicate identities. All 1,557 bad-side completed identities also occur on the good side. The 725 remaining good-side identities are excluded. `slowdown/common-units.csv.gz` contains every common identity, both wall and CPU times, snapshot times, ratios, increases, and non-snapshot residuals.

| Common-unit aggregate | f87d7884 | 1f7df6a2 |
|---|---:|---:|
| Unit wall seconds | 776.048 | 1,324.296 |
| Snapshot wall seconds | 714.347 | 732.757 |
| Non-snapshot residual seconds | 61.701 | 591.539 |
| Unit process-CPU seconds | 1,182.901 | 1,713.892 |

These are sums of completed unit timers, not census wall times. They exclude load, per-file refusal scans and registration outside those timers. The residual is total unit time minus time inside `latentSnapshot`; it includes lowering, timer/bookkeeping and recovery, not exclusively expression lowering. Snapshot wall time includes allocation and GC delays occurring inside the call. The 60-second profiles sample different phases of the two full passes, so their percentages are not a whole-run allocation of time. The per-unit counters provide that split directly.

## Top 30 absolute increases

Seconds are elapsed wall time. Snapshot columns are measured within those unit totals, not additional time. Subtract snapshot from total for the non-snapshot residual. Names can repeat for overload declarations; the source position is the identity.

| Declaration and unit identity | Good s | Bad s | Increase s | Ratio | Good snapshot s | Bad snapshot s |
|---|---:|---:|---:|---:|---:|---:|
| dumpTypes (`src/compiler/tracing.ts:218:5`) | 4.037 | 137.576 | 133.539 | 34.08x | 2.136 | 2.442 |
| assertOptionalToken (`src/compiler/debug.ts:316:5`) | 0.589 | 67.355 | 66.767 | 114.44x | 0.404 | 31.385 |
| createDirectoryWatcher (`src/compiler/sys.ts:422:5`) | 1.170 | 58.458 | 57.288 | 49.98x | 0.645 | 1.337 |
| assertOptionalToken (`src/compiler/debug.ts:315:5`) | 0.346 | 57.422 | 57.076 | 165.73x | 0.183 | 1.314 |
| symbolName (`src/compiler/utilitiesPublic.ts:870:1`) | 3.630 | 58.868 | 55.239 | 16.22x | 0.257 | 0.216 |
| isParameterPropertyDeclaration (`src/compiler/utilitiesPublic.ts:615:1`) | 0.803 | 45.006 | 44.202 | 56.03x | 0.078 | 0.160 |
| createScanner (`src/compiler/scanner.ts:1023:1`) | 149.153 | 190.655 | 41.502 | 1.28x | 147.511 | 188.542 |
| createUseFsEventsOnParentDirectoryWatchFile (`src/compiler/sys.ts:386:1`) | 9.242 | 50.551 | 41.309 | 5.47x | 5.347 | 3.465 |
| isEmptyBindingElement (`src/compiler/utilitiesPublic.ts:629:1`) | 1.093 | 37.272 | 36.179 | 34.11x | 0.109 | 0.126 |
| getTypeParameterOwner (`src/compiler/utilitiesPublic.ts:603:1`) | 1.482 | 22.140 | 20.658 | 14.94x | 0.847 | 0.196 |
| nodeHasName (`src/compiler/utilitiesPublic.ts:931:1`) | 0.755 | 19.281 | 18.525 | 25.52x | 0.184 | 0.134 |
| getCombinedFlags (`src/compiler/utilitiesPublic.ts:644:1`) | 1.022 | 18.158 | 17.136 | 17.78x | 0.487 | 0.502 |
| attachFlowNodeDebugInfoWorker (`src/compiler/debug.ts:495:5`) | 0.167 | 11.112 | 10.944 | 66.37x | 0.147 | 9.907 |
| buildPartRegExp (`src/compiler/semver.ts:35:1`) | 0.221 | 8.138 | 7.917 | 36.84x | 0.219 | 8.096 |
| EntityNameExpression (`src/compiler/types.ts:2965:1`) | 0.150 | 7.598 | 7.448 | 50.76x | 0.147 | 7.598 |
| attachFlowNodeDebugInfo (`src/compiler/debug.ts:531:5`) | 0.333 | 7.272 | 6.939 | 21.84x | 0.300 | 7.245 |
| Version (`src/compiler/semver.ts:46:1`) | 9.286 | 15.075 | 5.789 | 1.62x | 7.096 | 7.623 |
| Debug (`src/compiler/debug.ts:111:1`) | 3.796 | 9.429 | 5.633 | 2.48x | 3.788 | 9.419 |
| EntityNameOrEntityNameExpression (`src/compiler/types.ts:2966:1`) | 0.199 | 5.815 | 5.616 | 29.21x | 0.199 | 5.815 |
| PropertyAccessExpression (`src/compiler/types.ts:2969:1`) | 0.064 | 3.935 | 3.872 | 61.69x | 0.064 | 3.935 |
| AccessExpression (`src/compiler/types.ts:2967:1`) | 0.197 | 3.807 | 3.611 | 19.37x | 0.197 | 3.807 |
| numericIdentifierRegExp (`src/compiler/semver.ts:39:1`) | 0.323 | 3.492 | 3.169 | 10.82x | 0.322 | 3.476 |
| scanNumberFragment (`src/compiler/scanner.ts:1172:5`) | 1.649 | 4.783 | 3.134 | 2.90x | 1.638 | 4.592 |
| createDirectoryWatcher (`src/compiler/sys.ts:635:5`) | 2.421 | 5.321 | 2.900 | 2.20x | 1.889 | 4.262 |
| getNodeSystem (`src/compiler/sys.ts:1473:5`) | 12.551 | 15.226 | 2.675 | 1.21x | 10.894 | 13.217 |
| getRelativePathFromDirectory (`src/compiler/path.ts:1039:1`) | 0.455 | 2.489 | 2.034 | 5.47x | 0.394 | 2.055 |
| watchFile (`src/compiler/sys.ts:471:5`) | 1.252 | 3.240 | 1.988 | 2.59x | 1.049 | 2.986 |
| PrivateIdentifierPropertyAccessExpression (`src/compiler/types.ts:2977:1`) | 0.087 | 1.838 | 1.751 | 21.05x | 0.087 | 1.838 |
| onWatchedFileStat (`src/compiler/sys.ts:544:1`) | 0.485 | 2.232 | 1.747 | 4.61x | 0.411 | 1.907 |
| scheduleUpdateChildWatches (`src/compiler/sys.ts:763:5`) | 0.382 | 1.890 | 1.508 | 4.94x | 0.332 | 1.565 |

## Body-time bisect and confirmations

`dumpTypes` has the largest absolute increase: **4.037 s to 137.576 s, 34.08x**. Snapshot time is only **2.136 s to 2.442 s**; the residual is 1.902 s to 135.134 s. It satisfies both selection conditions, at least 2x and at least 20 seconds on the bad side.

Before testing intermediate commits, froze the predicate at the midpoint of those completed body timers: **70.8068843265 seconds**. A completed unit above that threshold is bad. The hard timeout is exactly three times the bad endpoint body time, **412.72938069 seconds**, followed by a 20-second hard-kill fallback. Load, selection, per-file registration and profile start/stop are outside the unit timer. The records confirm that exactly one unit, `dumpTypes`, was selected on every test.

Started `git bisect start 1f7df6a2 f87d7884 -- internal/lower internal/load internal/flow internal/fresh`. The full-pass profile identified the new `plainReceiver` call site, so the guided probes tested the commit introducing it and then its immediate parent instead of the default midpoint. They classified bad and good, closing the interval. `slowdown/bisect.log` is the actual Git log, including the first-bad result; no untested midpoint is presented as a measured commit. The bisect was reset afterward.

Both revisions were then confirmed in two fresh processes each. The second invocation reused each revision's identical compiled binary, not a warm lowering process. Every attempt returned before the timeout. `slowdown/test-timings.jsonl` includes setup/build time separately; only the body column determines the predicate.

| Revision | Run | Body wall s | Snapshot wall s | Body CPU s | Verdict |
|---|---|---:|---:|---:|---|
| Parent `4a8171994` | 1 | 5.206 | 2.948 | 6.472 | good |
| Parent `4a8171994` | 2 | 6.432 | 4.184 | 8.492 | good |
| First bad `b2b32fb89` | 1 | 106.223 | 3.602 | 119.500 | bad |
| First bad `b2b32fb89` | 2 | 122.500 | 2.632 | 160.948 | bad |

The first bad and parent also have CPU profiles spanning only the timed body, started before its timer and stopped after the time record. The deliberately nonexistent replay finding signature generates an expected "signature did not reproduce" message after completion. It is not the timing verdict or a claim that the compiler accepted the unit. The driver requires one selected unit.

## Which walk changed

The relevant diff is in `internal/lower/interface_cast.go` and `entries_records.go`, preserved in `slowdown/first-bad-lowering.patch.gz` with exact source snapshots beside it.

Previously `viewFieldReceiverTypeID` returned the receiver's id. The first bad commit adds `l.plainReceiver(id)` on that path. The new proof checks whether records or merged classes could reach that receiver, so bare-name checked-field entries can be avoided for a proven plain receiver. It caches requests by receiver id in the new `plainAsked` map.

Before checking those types, each first request for an id calls `moduleOrder(l.program.Files()[0])`. In this cyclic tsc graph, `moduleOrder` rebuilds the ESM order, checks the modules, and calls `proveModuleReads`. That existing proof walks module ASTs and resolves global reads through the checker. Its result is rebuilt on each call. The enclosing `view` visitor itself walks property accesses throughout the program; its calls to `checkedViewReceiverField` now feed distinct receiver ids into this new proof. Consequently the same module-read walk is requested repeatedly. This is additional whole-program work per receiver, not an infinite generator resume or an exponential container/type-argument walk established by these measurements.

The overlay exposes this path while independently lowering `dumpTypes`, including its structural casts of `Type` to specific type interfaces. It continues after unsupported statements. Its ordinary snapshot rollback also restores the new owned `plainAsked` map after a failed statement, so proof requests recorded only by that failed attempt do not survive. The generated copier explicitly copies that map. The expensive samples are in the production proof path, not in generating or copying that map.

The main bad-side profile was captured during `dumpTypes`. Its cumulative CPU samples are:

| Frame | Sample CPU s | Cumulative share |
|---|---:|---:|
| `plainReceiver` | 37.77 | 84.20% |
| `moduleOrder` | 37.25 | 83.04% |
| `proveModuleReads` | 34.96 | 77.93% |
| `Checker.GetSymbolAtLocation` | 17.05 | 38.01% |
| `NameResolver.Resolve` | 8.30 | 18.50% |

The call chain is `view` / `checkedViewReceiverField` -> `viewFieldReceiverTypeID` -> `plainReceiver` -> `moduleOrder` -> `proveModuleReads` -> checker symbol lookup / binder resolution, with the AST visitors between those frames. Percentages overlap along the chain and must not be added. The full profile files, flat tables and cumulative tables are included.

The controlled first-bad body profile independently has `plainReceiver` at 81.59%, `moduleOrder` at 80.97%, and `proveModuleReads` at 76.16%. The parent body profile has no new `plainReceiver` path; `latentSnapshot` accounts for 45.21% of its sampled CPU. The good full-pass 60-second sample, taken in a different phase, has `latentSnapshot` at 53.69%. This agrees with the direct counters: snapshots are an existing substantial meter cost, but they do not account for this unit's regression.

## Plain compiler control

Built the standard `stage3/census/tool` without a lowering overlay at both endpoints and ran it on `adapted/src/compiler/tracing.ts`, the file containing `dumpTypes`. This tool emits both a per-file record and a final whole-program record, even for one root. Every invocation completed within the same 412.729-second hard cap.

| Endpoint | Per-file s | Final one-root program s | Result before body lowering |
|---|---:|---:|---|
| f87d7884 | 58.936 | 43.342 | Refused at `core.ts:332:1`, a generator function |
| 1f7df6a2 | 42.745 | 24.577 | Refused at `core.ts:381:26`, a mutable type-parameter view |

Those times include normal load/check and the ordinary lowering gate. Neither run reaches `dumpTypes`: `lowerCheckedWithCycleTypes` returns from its refusal loop before `declareModule` lowers function bodies. Thus the ordinary per-file gate is faster in these samples, and provides no measured body time for this unit. It would be incorrect to call the latent 34x result an observed slowdown of an admitted plain compiler program.

The new work is in production lowering code. The latent overlay's continuation past refusals makes it observable here and can amplify repeated proofs through rollback. The measured regression is not a slower overlay snapshot implementation; it is an expensive new lowering proof on the latent path. These controls do not establish the speed of an admitted real program that reaches the same proof.

## Scope and artifacts

The bisect identifies the first bad commit under the frozen selected-unit predicate. The aggregate endpoint comparison spans all commits in the window; no claim is made that this one commit explains every unit's increase or the exact duration of an uncapped daily run. The runs are right-censored at 30 minutes, and no unfinished unit is included in the common-unit ranking. No fix was made or proposed.

`slowdown/` contains both full partial timing streams, the common-unit CSV, the top-30 data, six CPU profiles, their flat/cumulative reports, raw per-commit records and timings, the exact predicate, actual bisect log, relevant diff and source, plain compiler records, binary hashes, environment details, and the scripts that produced the observations. All census processes started for this follow-up have exited or were stopped before publication.
