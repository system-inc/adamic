# tracing-fs-import completed proof

Task nqp4p7g. Final commit `0334d4e3cacb729b20f5b3790cecee7f41c185be` is one commit on `2d7a1683c60339c33d3fd33ae4ff91d242f01094`
(`typescript/tracing-fs-views-2` at 2d7a1683). It changes only adaptations 47,
65, and the appended ledger entry, after both the pop/popAll and fs-view entries.

## Corrected reference contract

| Quantity | Before | After |
| --- | ---: | ---: |
| Reads in startTracing catches | 3 | 0 |
| Live export reads in run.js and typescript.js | 6 | 6 |
| Runtime export keys, including errorMessage | 2,256 | 2,256 |

All six export expressions and their complete emitted source lines are byte
identical. `_tsc.js` alone eliminates errorMessage: its sole reader was the
removed catch. See helper2-reference-census.json and public2-contracts.json.

## Emission and module effects

All ten emitted files are inventoried. emit2-hunk-classification.md lists all
68 hunks, backed by the passing symbol/token certificate. Directly reviewed
tracing import/local/loader edits are marked. Outside changes are consistent
esbuild renames, the closed helper removal, or declaration-only hostErrors
relocation. Binding shorthand normalization preserves each property key and
checks the renamed local's symbol bijectively. The empty esbuild hostErrors
initializer callback contains only a strict directive, and its binding is
used only by no-argument calls.

host2-declarations.json pins the identical source module body: function
declarations errorCode, errorMessage, propertyOf, with no other top-level
statements. 47 drops the unused errorMessage import. All 17 fs reads bind to
the namespace import. Node 24.19.0 on Linux x64 has fs built in; normal builtin
loading cannot fail in these configurations. The source module graph and
actual load stacks in catch2-positive.json show sys's fs load before tracing
can be invoked. The synthetic loader-failure counter control fires once.

## Complete serial lanes

Both unmodified measure_shards.py invocations execute all 398 units with
concurrency 2, with no other build/test workload alongside the accepted pair.
Every shard passes. Each run records 106,366 passing observations, the same
single sanctioned Public APIs failure, and zero pending. All shard runner
copies match the input hash. Memory-event counters remain unchanged.
Control: 840.363s, maximum shard 8.549s.
Final: 829.176s, maximum shard 8.018s.
The control counter observes 398 harness processes and zero catch executions.
Raw reports, observations, logs and baseline diffs are in control2-lane.tgz
and after2-lane.tgz; archive hashes are in FINAL_REPORT.json. Each completed
private adapted-tree scratch view was removed only after its child exited and
its report was complete; its runner hash was recorded before removal.

## Trace and plain Node witnesses

Two control and two final compilations use the same strict 152-line project.
The controls themselves differ: trace event counts are 30 and 31. Final counts
are 31 and 33. tracing.ts samples ph X complete events when [time,endTime)
crosses a 10 ms sampling point. Removing only that sampled kind leaves the
same 23-event inventory in every run after named timestamp/pid normalization.
Observed sampled names are checkExpression, checkVariableDeclaration,
emitDeclarationFileOrBundle, emitJsFileOrBundle, findSourceFile,
processRootFiles, transformNodes. types.json and legend.json are byte-identical
across all four runs. Raw files and hashes are retained.

Ordinary project-mode tsc does not flush legend.json. The trace-only Node
wrapper evaluates the unchanged built _tsc.js bytes and intercepts its exit
to call the existing dumpTracingLegend before returning the compiler's actual
exit code. It changes no compiler source or clock. The separate plain witness
runs the real unmodified tsc.js normally: exits, stdout, stderr, emitted JS and
declarations all match. Both public d.ts files match byte for byte.

## Application, replay, census and mutant

Fresh official control/final apply, builds, harnesses and all 79-root adamic
types checks pass. Official final replay reproduces all 747 source files;
immediate 47 and 65 completed-state replays perform zero edits. Whole-owner
drift, module collision, parameter capture and bare namespace value controls
refuse. FINAL_REPORT.json links the current source/emit identity checks.

Both complete tracing censuses have zero checker diagnostics and complete
trailers. The require refusal disappears; the readonly process.pid declaration
is unchanged. Core's new first stop is sys.ts:1654:13,
node:inspector.Session.post. Restoring the exact control tracing.ts bytes brings
back the ambient-require refusal. On the new base its physical address is
381:1, corresponding to 380:1 before the pop amendment added a line.

## Retained failed attempts

Initial and superseded stop reports remain in initial-stop.md, closure-stop.md
and Git history. The old all-before-in-catch / after-zero condition was corrected
by the user; no live exports are deleted. Earlier strict raw trace comparison
and the first missing-legend wrapper attempts are retained without passing
claims. A full-project shadow probe was killed while an assertion formatted
cyclic compiler symbols; the guard now compares identity as a boolean and all
four direct binding controls pass with small messages. The overlapping first
control lane completed 398 passing shards but was rejected by the unchanged
resource guard because oom_kill increased. That full attempt is preserved in
control2-lane-attempt1.tgz. The accepted serial pair was rerun only after the
fix and all other checks completed. No lane sanction, expectation, timeout,
fixture, manifest or memory guard was weakened.
