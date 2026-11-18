# CPU profiles — depth-gate submission census

Base: **`c56b022995e1e960bad71626fd2074bc46aee485`**. Head: **`26d208af47f81143ff1db2280a5b43292deb96ba`**. One utilities.ts full single-file run per side, sequential base then head, on the **same unchanged adapted tree** used for [census-2](REPORT.md). Both complete normally within 30 minutes.

**The earlier cost gap is not reproduced by this profiled pair.** Wall time is 604.772 s base versus 604.656 s head; wait4 process CPU is 799.092 s versus 801.471 s. Census-2’s unprofiled wall-time means were 599.675 s versus 812.441 s (+35.5%). These profiles cannot establish where that earlier large difference was spent, and do not establish why it disappeared.

Sampled CPU totals are **787.74 s base** and **789.94 s head**: **+2.20 CPU seconds**. These are sampled CPU totals, not wall time.

## Capture and behavior check

Both own-SHA entry tools were rebuilt with the **identical scratch-only `entry_main.go` hook**: import `runtime/pprof`; if `ADAMIC_CPU_PROFILE` is set, create that file and call `pprof.StartCPUProfile` at the start of main; defer `pprof.StopCPUProfile` and file close until main returns. This profiles loading/checking, full latent lowering and stream serialization, excluding process initialization before main. The Go default CPU sampling rate is used (100 Hz / 10 ms). No lowering implementation, adaptation or ledger is patched. [Exact hook diff](profiles/build/base/profile-hook.diff.gz), [identical full overlay main](profiles/profile-overlay-main.go.txt), [preparation script](profiles/prepare.py), [overlay maps](profiles/build/), and binary hashes in run statuses are retained.

The canonical overlays are those generated from each source SHA for census-2. The base worktree now contains the earlier report commit `62eb8c2bb`; its compiler/source files are verified identical to c56b0229, with only the prior report/evidence added. The head worktree is at 26d208af. Both use the same Go toolchain as census-2. The distribution lacked a compiled `pprof` helper; its shipped `cmd/pprof` was built **after both measured runs**, so no build ran alongside either profile capture.

Environment is unchanged from census-2 apart from the additional profile path: `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1`, `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_ONLY_FILE=/workspace/depthgate2-adapted/src/compiler/utilities.ts`, `LATENT_TIMES=/dev/stderr`. Neither side sets `ADAMIC_PROBE_DEPTHGATE`. The command remains `entry-census /workspace/depthgate2-adapted/src/tsc/tsc.ts <run>/latent.jsonl`, using the rebuilt own-side binary and an absolute `ADAMIC_CPU_PROFILE=<run>/cpu.pprof`. Whole stdout/stderr and streams are archived; the 30-minute termination guard is unchanged and did not fire.

**Findings match census-2 byte for byte, on each side, against both original repeats.** The comparison extracts each top-level serialized `findings` array directly from the original JSON text, preserves every byte within it, and concatenates the arrays in stream order with newlines. It does not re-encode, sort, strip positions, deduplicate, or alter messages. Timing/unit metadata outside these arrays is not part of this finding comparison. [Comparison method](profiles/compare.py) and [exact hashes/results](profiles/behavior-comparison.json).

| Side | Findings | Finding-array bytes | SHA256 | Census-2 comparison |
| --- | --- | --- | --- | --- |
| base | 1313 | 793405 | `5dbdca0c6e87fbf5dbc3ca490dcbe46b0fcdbc58c4202f5cfdff03ff925c50ab` | identical to repeats 1 and 2 |
| head | 1297 | 784843 | `860c490e087d2b842ef43795ec442259d9c7b2442866cd9aaa95ee0d927a7b20` | identical to repeats 1 and 2 |

| Side | Wall s | wait4 user+system CPU s | Sampled CPU s | Peak RSS KiB | Machine | Exit |
| --- | --- | --- | --- | --- | --- | --- |
| base | 604.772 | 799.092 | 787.74 | 1,772,036 | ae770fcedbbc | 0 |
| head | 604.656 | 801.471 | 789.94 | 1,782,368 | ae770fcedbbc | 0 |

Profiles: **[base.pprof](profiles/base.pprof)** and **[head.pprof](profiles/head.pprof)**. Status metadata preserves timestamps, CPU/memory limits, selected-source hash, binary hash and exact environment. This is one profiled observation per side; profiling overhead and sampling variation apply.

## Where the extra CPU sits

Each row sums the **CPU nanosecond weight of every selected sample once** after `pprof -focus`/`-ignore`. Recursive cumulative rows are not added together. “Extra” is head minus base, and the final percentage divides by the total sampled CPU difference. These are stack attribution measurements, not an allocation-causality estimate.

| Selected stacks | Base CPU s | Head CPU s | Extra CPU s | Share of total extra CPU |
| --- | --- | --- | --- | --- |
| mentionsHeld, cumulative (any call stack) | 1.62 | 1.67 | +0.05 | +2.3% |
| containerWhereOtherGoes and callees, excluding mentionsHeld | 0.38 | 0.37 | -0.01 | -0.5% |
| All remaining stacks (disjoint remainder) | 785.74 | 787.90 | +2.16 | +98.2% |
| typeNesting and callees, globally (overlaps above) | 0.63 | 0.73 | +0.10 | +4.5% |
| typeNesting within walk, excluding mentionsHeld (subset) | 0.00 | 0.00 | +0.00 | +0.0% |
| Walk excluding mentionsHeld and typeNesting (subset) | 0.38 | 0.37 | -0.01 | -0.5% |

The first three rows partition the total. The percentages use the small **2.20 s difference in this profile pair**, not census-2’s earlier gap; they are sensitive to sampling/run variation. A measured zero means **no observed samples**, not proof of zero work. **typeNesting has no observed samples inside the walk on either side** (globally it changes 0.63 → 0.73 s). The strict outer-frame and expanded recursive-pair-anchor views agree exactly: **0 sampled CPU seconds** have a pair frame but omit the outer frame after excluding mentionsHeld, so no ancestry correction is needed here. The Go runtime’s CPU stack limit is 64 PCs; stack-length diagnostics are retained.

**typeNesting is overlapping evidence**, principally useful as a subset of the walk; do not add it to the walk row. The final two subset rows partition the walk excluding mentionsHeld. Background GC and other runtime work without the selected functions on its stack remain in the disjoint remainder; this analysis does not assign that work back to allocation sites. Exact integer weights and partition checks are in [attribution.json](profiles/attribution.json).

Commands, repeated with each side’s binary/profile:

```sh
GO=/workspace/adamic-tools/go/bin/go
BINARY=/workspace/depthgate2-profiles/HEAD_OR_BASE/entry-census
PROFILE=/workspace/depthgate2-profile-runs/HEAD_OR_BASE-utilities-r1/cpu.pprof
"$GO" tool pprof -raw "$BINARY" "$PROFILE"
"$GO" tool pprof -raw -focus=mentionsHeld "$BINARY" "$PROFILE"
"$GO" tool pprof -raw -focus=containerWhereOtherGoes -ignore=mentionsHeld "$BINARY" "$PROFILE"
"$GO" tool pprof -raw -focus=typeNesting "$BINARY" "$PROFILE"
"$GO" tool pprof -raw -focus=containerWhereOtherGoes -ignore='mentionsHeld|typeNesting' "$BINARY" "$PROFILE"
"$GO" tool pprof -raw -ignore='containerWhereOtherGoes|mentionsHeld' "$BINARY" "$PROFILE"
```

Use lowercase `base` / `head` in those paths. The same filters were also run with `-top -cum` for readable focused reports. To obtain the typeNesting-in-walk intersection, the analyzer reads the walk-without-mentionsHeld raw samples and sums those whose raw location stack contains typeNesting; it does not OR the two focus expressions. [All exact commands](profiles/commands.txt), [analysis script](profiles/analyze.py), [focused/raw reports](profiles/analysis/).

## pprof top reports — first 40 lines each

The excerpts below are the first 40 stdout lines from the literal commands shown, without adding a node-count or changing default thresholds. Full stdout/stderr is retained under `profiles/analysis/`.

### base — -top -cum

```sh
/workspace/adamic-tools/go/bin/go tool pprof -top -cum /workspace/depthgate2-profiles/base/entry-census /workspace/depthgate2-profile-runs/base-utilities-r1/cpu.pprof
```

```text
File: entry-census
Build ID: 731b31acb9220e0f71187c52f0d91c7fa703431e
Type: cpu
Time: 2026-11-18 13:19:39 UTC
Duration: 604.22s, Total samples = 787.74s (130.37%)
Showing nodes accounting for 568.71s, 72.20% of 787.74s total
Dropped 3016 nodes (cum <= 3.94s)
      flat  flat%   sum%        cum   cum%
     0.01s 0.0013% 0.0013%    427.77s 54.30%  github.com/system-inc/adamic/internal/lower.latentFullSelected
         0     0% 0.0013%    422.66s 53.65%  github.com/system-inc/adamic/internal/lower.latentFull (inline)
         0     0% 0.0013%    422.66s 53.65%  github.com/system-inc/adamic/internal/lower.latentRun
         0     0% 0.0013%    417.28s 52.97%  github.com/system-inc/adamic/internal/lower.latentFullSelected.func7
         0     0% 0.0013%    412.08s 52.31%  github.com/system-inc/adamic/internal/lower.LatentWithRelationChecks
         0     0% 0.0013%    406.72s 51.63%  github.com/system-inc/adamic/internal/lower.Latent
         0     0% 0.0013%    403.20s 51.18%  main.main
         0     0% 0.0013%    397.40s 50.45%  runtime.main
     0.01s 0.0013% 0.0025%    331.14s 42.04%  github.com/system-inc/adamic/internal/lower.(*lowering).latentStatement
         0     0% 0.0025%    328.93s 41.76%  github.com/system-inc/adamic/internal/lower.(*lowering).statement (inline)
         0     0% 0.0025%    323.60s 41.08%  github.com/system-inc/adamic/internal/lower.(*lowering).statements
         0     0% 0.0025%    316.63s 40.19%  github.com/system-inc/adamic/internal/lower.(*lowering).lowerFunction
         0     0% 0.0025%    311.21s 39.51%  github.com/system-inc/adamic/internal/lower.(*lowering).lowerBody
     0.02s 0.0025% 0.0051%    284.34s 36.10%  github.com/system-inc/adamic/internal/lower.(*lowering).latentStatementRaw
         0     0% 0.0051%    242.81s 30.82%  github.com/system-inc/adamic/internal/lower.(*lowering).enumNeverValue
         0     0% 0.0051%    241.38s 30.64%  github.com/system-inc/adamic/internal/lower.(*lowering).value
     0.01s 0.0013% 0.0063%    240.74s 30.56%  github.com/system-inc/adamic/internal/lower.(*lowering).latentExpressionRaw
     0.01s 0.0013% 0.0076%    238.86s 30.32%  github.com/system-inc/adamic/internal/lower.(*lowering).latentSpecExpression
         0     0% 0.0076%    237.02s 30.09%  github.com/system-inc/adamic/internal/lower.(*lowering).expression (inline)
     0.06s 0.0076% 0.015%    236.06s 29.97%  github.com/system-inc/adamic/internal/lower.(*lowering).latentSnapshot
     0.02s 0.0025% 0.018%       236s 29.96%  github.com/system-inc/adamic/internal/lower.latentCopyPointer_lowering
     0.03s 0.0038% 0.022%    235.96s 29.95%  github.com/system-inc/adamic/internal/lower.latentCopy_lowering
         0     0% 0.022%    222.22s 28.21%  github.com/system-inc/adamic/internal/lower.latentCopyIR[go.shape.*uint8]
    18.11s  2.30%  2.32%    222.16s 28.20%  github.com/system-inc/adamic/internal/lower.latentCopyIR[go.shape.*uint8].func1
    20.03s  2.54%  4.86%    207.67s 26.36%  github.com/microsoft/TypeScript/tsc/internal/ast.(*Node).ForEachChild
     7.05s  0.89%  5.76%    204.26s 25.93%  github.com/microsoft/TypeScript/tsc/internal/ast.visit (inline)
         0     0%  5.76%    198.89s 25.25%  github.com/system-inc/adamic/internal/lower.(*lowering).call
     0.12s 0.015%  5.77%    198.64s 25.22%  runtime.systemstack
     7.85s     1%  6.77%    197.92s 25.13%  github.com/microsoft/TypeScript/tsc/internal/ast.visitNodeList (inline)
     6.15s  0.78%  7.55%    197.83s 25.11%  github.com/microsoft/TypeScript/tsc/internal/ast.visitNodes (inline)
         0     0%  7.55%    197.57s 25.08%  github.com/system-inc/adamic/internal/lower.(*lowering).callOrMethod
         0     0%  7.55%    186.36s 23.66%  runtime.gcBgMarkWorker
```

### base — -top

```sh
/workspace/adamic-tools/go/bin/go tool pprof -top /workspace/depthgate2-profiles/base/entry-census /workspace/depthgate2-profile-runs/base-utilities-r1/cpu.pprof
```

```text
File: entry-census
Build ID: 731b31acb9220e0f71187c52f0d91c7fa703431e
Type: cpu
Time: 2026-11-18 13:19:39 UTC
Duration: 604.22s, Total samples = 787.74s (130.37%)
Showing nodes accounting for 568.71s, 72.20% of 787.74s total
Dropped 3016 nodes (cum <= 3.94s)
      flat  flat%   sum%        cum   cum%
    53.49s  6.79%  6.79%     79.54s 10.10%  reflect.(*rtype).Field
    30.98s  3.93% 10.72%     68.65s  8.71%  runtime.tryDeferToSpanScan
    28.70s  3.64% 14.37%     28.70s  3.64%  runtime.spanClass.sizeclass (inline)
    26.28s  3.34% 17.70%    117.53s 14.92%  runtime.scanObject
    22.82s  2.90% 20.60%     22.82s  2.90%  internal/runtime/maps.ctrlGroup.matchH2 (inline)
    20.03s  2.54% 23.14%    207.67s 26.36%  github.com/microsoft/TypeScript/tsc/internal/ast.(*Node).ForEachChild
    19.15s  2.43% 25.57%     30.01s  3.81%  runtime.findObject
    18.27s  2.32% 27.89%     18.27s  2.32%  runtime.memmove
    18.11s  2.30% 30.19%    222.16s 28.20%  github.com/system-inc/adamic/internal/lower.latentCopyIR[go.shape.*uint8].func1
    15.91s  2.02% 32.21%     15.91s  2.02%  runtime.memclrNoHeapPointers
    11.68s  1.48% 33.69%     25.81s  3.28%  reflect.(*structType).Field
    11.64s  1.48% 35.17%     16.93s  2.15%  reflect.Value.Field
    10.99s  1.40% 36.57%     22.45s  2.85%  github.com/system-inc/adamic/internal/lower.(*lowering).refuseViewElementReads.func1
    10.92s  1.39% 37.95%     10.92s  1.39%  internal/runtime/maps.memHashAES
     9.87s  1.25% 39.21%     30.17s  3.83%  runtime.mapaccess2_fast64
     8.73s  1.11% 40.31%      8.95s  1.14%  internal/runtime/gc/scan.scanSpanPackedAVX512
     8.36s  1.06% 41.38%     13.69s  1.74%  reflect.Value.assignTo
     8.15s  1.03% 42.41%     65.59s  8.33%  runtime.scanSpan
        8s  1.02% 43.43%     17.17s  2.18%  github.com/system-inc/adamic/internal/lower.(*lowering).writtenAcrossFunctions.func1
     7.92s  1.01% 44.43%     99.51s 12.63%  github.com/system-inc/adamic/internal/lower.latentPrivateIRField
     7.85s     1% 45.43%    197.92s 25.13%  github.com/microsoft/TypeScript/tsc/internal/ast.visitNodeList (inline)
     7.83s  0.99% 46.42%      7.86s     1%  runtime.ifaceeq
     7.05s  0.89% 47.32%    204.26s 25.93%  github.com/microsoft/TypeScript/tsc/internal/ast.visit (inline)
     6.81s  0.86% 48.18%     17.89s  2.27%  runtime.scanObjectsSmall
     6.59s  0.84% 49.02%     13.62s  1.73%  runtime.scanObjectSmall
     6.54s  0.83% 49.85%      6.63s  0.84%  internal/abi.Name.IsExported (inline)
     6.17s  0.78% 50.63%     27.66s  3.51%  github.com/microsoft/TypeScript/tsc/internal/binder.(*NameResolver).Resolve
     6.15s  0.78% 51.41%    197.83s 25.11%  github.com/microsoft/TypeScript/tsc/internal/ast.visitNodes (inline)
     6.14s  0.78% 52.19%      6.14s  0.78%  runtime.typePointers.nextFast (inline)
     6.09s  0.77% 52.96%     44.82s  5.69%  reflect.Value.Set
     5.93s  0.75% 53.72%      6.08s  0.77%  runtime.spanOf (inline)
     5.85s  0.74% 54.46%     24.59s  3.12%  runtime.mapaccess2_faststr
```

### head — -top -cum

```sh
/workspace/adamic-tools/go/bin/go tool pprof -top -cum /workspace/depthgate2-profiles/head/entry-census /workspace/depthgate2-profile-runs/head-utilities-r1/cpu.pprof
```

```text
File: entry-census
Build ID: d446c7b0d8ec818d0b19f0bb784ec1e810d2e824
Type: cpu
Time: 2026-11-18 13:29:44 UTC
Duration: 604.02s, Total samples = 789.94s (130.78%)
Showing nodes accounting for 565.34s, 71.57% of 789.94s total
Dropped 3033 nodes (cum <= 3.95s)
      flat  flat%   sum%        cum   cum%
         0     0%     0%    422.04s 53.43%  github.com/system-inc/adamic/internal/lower.latentFullSelected
         0     0%     0%    417.22s 52.82%  github.com/system-inc/adamic/internal/lower.latentFull (inline)
         0     0%     0%    417.22s 52.82%  github.com/system-inc/adamic/internal/lower.latentRun
     0.01s 0.0013% 0.0013%    410.97s 52.03%  github.com/system-inc/adamic/internal/lower.latentFullSelected.func7
         0     0% 0.0013%    406.73s 51.49%  github.com/system-inc/adamic/internal/lower.LatentWithRelationChecks
         0     0% 0.0013%    401.34s 50.81%  github.com/system-inc/adamic/internal/lower.Latent
         0     0% 0.0013%    397.06s 50.26%  main.main
         0     0% 0.0013%    391.36s 49.54%  runtime.main
     0.01s 0.0013% 0.0025%    331.72s 41.99%  github.com/system-inc/adamic/internal/lower.(*lowering).latentStatement
         0     0% 0.0025%    329.27s 41.68%  github.com/system-inc/adamic/internal/lower.(*lowering).statement (inline)
         0     0% 0.0025%    323.61s 40.97%  github.com/system-inc/adamic/internal/lower.(*lowering).statements
         0     0% 0.0025%    316.41s 40.05%  github.com/system-inc/adamic/internal/lower.(*lowering).lowerFunction
         0     0% 0.0025%    310.39s 39.29%  github.com/system-inc/adamic/internal/lower.(*lowering).lowerBody
         0     0% 0.0025%    283.11s 35.84%  github.com/system-inc/adamic/internal/lower.(*lowering).latentStatementRaw
     0.02s 0.0025% 0.0051%    244.11s 30.90%  github.com/system-inc/adamic/internal/lower.(*lowering).enumNeverValue
     0.01s 0.0013% 0.0063%    242.24s 30.67%  github.com/system-inc/adamic/internal/lower.(*lowering).value
     0.01s 0.0013% 0.0076%    241.88s 30.62%  github.com/system-inc/adamic/internal/lower.(*lowering).latentExpressionRaw
     0.02s 0.0025%  0.01%    240.02s 30.38%  github.com/system-inc/adamic/internal/lower.(*lowering).latentSpecExpression
         0     0%  0.01%    237.93s 30.12%  github.com/system-inc/adamic/internal/lower.(*lowering).expression (inline)
     0.07s 0.0089% 0.019%    235.25s 29.78%  github.com/system-inc/adamic/internal/lower.(*lowering).latentSnapshot
         0     0% 0.019%    235.18s 29.77%  github.com/system-inc/adamic/internal/lower.latentCopyPointer_lowering
     0.02s 0.0025% 0.022%    235.17s 29.77%  github.com/system-inc/adamic/internal/lower.latentCopy_lowering
         0     0% 0.022%    220.79s 27.95%  github.com/system-inc/adamic/internal/lower.latentCopyIR[go.shape.*uint8]
    18.11s  2.29%  2.31%    220.74s 27.94%  github.com/system-inc/adamic/internal/lower.latentCopyIR[go.shape.*uint8].func1
    21.43s  2.71%  5.03%    207.57s 26.28%  github.com/microsoft/TypeScript/tsc/internal/ast.(*Node).ForEachChild
     8.11s  1.03%  6.05%    204.36s 25.87%  github.com/microsoft/TypeScript/tsc/internal/ast.visit (inline)
     0.13s 0.016%  6.07%    201.40s 25.50%  runtime.systemstack
     8.25s  1.04%  7.11%    199.22s 25.22%  github.com/microsoft/TypeScript/tsc/internal/ast.visitNodeList (inline)
     6.44s  0.82%  7.93%    199.16s 25.21%  github.com/microsoft/TypeScript/tsc/internal/ast.visitNodes (inline)
         0     0%  7.93%    199.01s 25.19%  github.com/system-inc/adamic/internal/lower.(*lowering).call
         0     0%  7.93%    197.81s 25.04%  github.com/system-inc/adamic/internal/lower.(*lowering).callOrMethod
         0     0%  7.93%    189.81s 24.03%  runtime.gcBgMarkWorker
```

### head — -top

```sh
/workspace/adamic-tools/go/bin/go tool pprof -top /workspace/depthgate2-profiles/head/entry-census /workspace/depthgate2-profile-runs/head-utilities-r1/cpu.pprof
```

```text
File: entry-census
Build ID: d446c7b0d8ec818d0b19f0bb784ec1e810d2e824
Type: cpu
Time: 2026-11-18 13:29:44 UTC
Duration: 604.02s, Total samples = 789.94s (130.78%)
Showing nodes accounting for 565.34s, 71.57% of 789.94s total
Dropped 3033 nodes (cum <= 3.95s)
      flat  flat%   sum%        cum   cum%
    54.96s  6.96%  6.96%     82.08s 10.39%  reflect.(*rtype).Field
    31.66s  4.01% 10.97%     69.41s  8.79%  runtime.tryDeferToSpanScan
    29.32s  3.71% 14.68%     29.32s  3.71%  runtime.spanClass.sizeclass (inline)
    28.39s  3.59% 18.27%    119.08s 15.07%  runtime.scanObject
    21.43s  2.71% 20.98%    207.57s 26.28%  github.com/microsoft/TypeScript/tsc/internal/ast.(*Node).ForEachChild
    21.22s  2.69% 23.67%     21.22s  2.69%  internal/runtime/maps.ctrlGroup.matchH2 (inline)
    18.98s  2.40% 26.07%     29.81s  3.77%  runtime.findObject
    18.11s  2.29% 28.37%    220.74s 27.94%  github.com/system-inc/adamic/internal/lower.latentCopyIR[go.shape.*uint8].func1
    16.31s  2.06% 30.43%     16.31s  2.06%  runtime.memclrNoHeapPointers
    16.22s  2.05% 32.48%     16.22s  2.05%  runtime.memmove
    11.61s  1.47% 33.95%     26.97s  3.41%  reflect.(*structType).Field
    11.43s  1.45% 35.40%     11.43s  1.45%  internal/runtime/maps.memHashAES
    11.36s  1.44% 36.84%     16.52s  2.09%  reflect.Value.Field
    10.62s  1.34% 38.18%     23.08s  2.92%  github.com/system-inc/adamic/internal/lower.(*lowering).refuseViewElementReads.func1
    10.33s  1.31% 39.49%     28.63s  3.62%  runtime.mapaccess2_fast64
     8.95s  1.13% 40.62%     67.76s  8.58%  runtime.scanSpan
     8.43s  1.07% 41.69%      8.77s  1.11%  internal/runtime/gc/scan.scanSpanPackedAVX512
     8.28s  1.05% 42.74%    102.17s 12.93%  github.com/system-inc/adamic/internal/lower.latentPrivateIRField
     8.25s  1.04% 43.78%    199.22s 25.22%  github.com/microsoft/TypeScript/tsc/internal/ast.visitNodeList (inline)
     8.14s  1.03% 44.81%     16.91s  2.14%  github.com/system-inc/adamic/internal/lower.(*lowering).writtenAcrossFunctions.func1
     8.11s  1.03% 45.84%    204.36s 25.87%  github.com/microsoft/TypeScript/tsc/internal/ast.visit (inline)
     7.90s  1.00% 46.84%      7.94s  1.01%  runtime.ifaceeq
     7.80s  0.99% 47.83%     13.04s  1.65%  reflect.Value.assignTo
     7.69s  0.97% 48.80%     15.01s  1.90%  runtime.scanObjectSmall
     6.75s  0.85% 49.66%     17.85s  2.26%  runtime.scanObjectsSmall
     6.62s  0.84% 50.49%      6.76s  0.86%  internal/abi.Name.IsExported (inline)
     6.44s  0.82% 51.31%    199.16s 25.21%  github.com/microsoft/TypeScript/tsc/internal/ast.visitNodes (inline)
     6.07s  0.77% 52.08%     42.44s  5.37%  reflect.Value.Set
     5.95s  0.75% 52.83%     23.93s  3.03%  runtime.mapaccess2_faststr
     5.81s  0.74% 53.57%     14.75s  1.87%  github.com/system-inc/adamic/internal/lower.(*lowering).entriesRecordStorageTypes.func1
     5.62s  0.71% 54.28%      6.20s  0.78%  github.com/microsoft/TypeScript/tsc/internal/ast.IsFunctionLikeKind (inline)
     5.46s  0.69% 54.97%     26.90s  3.41%  github.com/microsoft/TypeScript/tsc/internal/binder.(*NameResolver).Resolve
```

## Evidence and scope

[Validation](profiles/validation.txt) verifies source hashes, identical hooks, own-SHA source provenance, run completion, byte-for-byte findings, profile readability, stack partition identities and lossless archives. [SHA256SUMS](profiles/SHA256SUMS) covers every other profile-evidence file. Only PROFILE.md and this new profile evidence are added to the existing report branch; no candidate code, adaptation, ledger or `internal/` change is committed.
