# View guard roads paired census: blocked by measurement coverage

Task 3zex15c. Report only. **Stopped because the requested meter latent census does not evaluate `viewPairGuard`. Paired guard counts are unavailable, not zero.** No compiler change is proposed or included.

The remote `memory/view-guard-roads` tip was checked before setup and was still `769b0e534aa9106cec787a670d0243cd3734c9c0`. The base is `16429a304` (full SHA in provenance.json). The head contains exactly the three requested commits: `878739c99`, `a00d4fa87`, `769b0e534`. Its entire production delta is `internal/lower/graph_types.go`; the other changed file is its test.

## Why this cannot answer the question

The production call paths are:

```text
lowerWithElementKey
  -> lowering.findCycles(modules)                 lower.go:205
     -> finder.graphTypes(modules)                 cycles.go:148
        -> f.viewPairGuard()                       graph_types.go:366
     -> l.selectProgramMembership(finder, modules)  cycles.go:153
        -> f.viewPairGuard()                       program_region.go:140
```

`stage3/census/latent/make_overlay.py` replaces production lowering from `files := program.Files()` through the end of that lowering pipeline with:

```go
return nil, fmt.Errorf("latent census: measurement only; no IR output")
```

That replacement removes the `findCycles` call and the intervening production lowering helpers. The entry overlay preserves that replacement. Its preliminary `lower.Lower` call checks the no-output sentinel; it does not run production lowering.

The actual entry measurement calls `lower.Latent`. Its first-error unit path calls `lowerFunction` or `statement`, then records `l.unlowerable`. Its full-mode path continues and rolls back failed statements. Neither path calls `findCycles`, `graphTypes`, `selectProgramMembership`, or `viewPairGuard`. The production membership planner is reached from `findCycles`, not from the latent unit's membership queries. Thus no guard finding from the changed code can enter either requested stream, in either mode. A zero delta from completed runs would provide no evidence that the widened guard is harmless.

This is established by reading the exact parent and head sources and generated overlays. `call-path-evidence.py` checks the unique production call locations, confirms removal of the production lowering helpers, and asserts that every generated overlay Go file has no call to any of those four functions. Its assertions passed. `call-path-evidence.json`, `call-path/`, and `overlays/` preserve the evidence. Go evidence filenames end in `.go.txt`.

The earlier method at `287b32b3:stage3/census/view-pairs-734aef5f/REPORT.md` was read and is archived in `raw/earlier-report.md`. It explicitly added `logViewPairs()` at every unit end and explicitly did not run `findCycles`. Its one unrelated object pair is a result of that additional probe, not evidence that the standard meter runs the guard. The present run did not add that probe or a new guard hook.

## Requested results and limits

| Measure | Raw | Distinct |
| --- | ---: | ---: |
| Cleared findings | unavailable | unavailable |
| New findings | unavailable | unavailable |
| Changed findings | unavailable | unavailable |
| Previously lowered units now stopped | unavailable | unavailable |
| Finding-free sites now finding | unavailable | unavailable |
| New panics | unavailable | unavailable |
| Changed first stops in already blocked units | unavailable | unavailable |

No new stop can be read or classified from this measurement. Elements, index values (including index sources under named fields), and union arms are all unobserved by the guard. The additional writable method/accessor check in the same head is unobserved as well. There are no measured pairs for which this report can judge unrelatedness, quote a new guard reason, or claim a guard bug.

The parent first-error run was intentionally terminated after the coverage defect was confirmed. Its stream contains only the checker-clean entry header, with 81 resolved source files. The runner buffers all file records until its traversal completes, so the partial stream cannot be paired. The timer ledger preserves completed unit timings; it is not a findings ledger. The head first-error run was not started. `stage3/census/pair` was not run on the incomplete input.

Step 3 was not run: there is no valid affected-file set, and full mode has the same missing guard invocation. The corrected requested settings were retained: `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, and `LATENT_ONLY_FILE` set to each affected adapted file, parent against head. Running these settings without an additional guard measurement seam would not resolve the coverage defect.

A valid continuation needs a measurement-only seam that invokes the real guard on a defined per-unit state in both modes and attributes its failure without confusing a discarded or partial unit with a successfully lowered whole program. It must specify first-error and full-mode rollback behavior and graph-program eligibility. Only then can paired streams answer the requested regression question. This report stops instead of silently changing that measurement contract.

## Setup, time, and memory

Setup used `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh` (exit 0), and `npm ci --prefix stage3/api` in each paired worktree (both exit 0). `ADAMIC_TYPESCRIPT_SOURCE=/workspace/roads-typescript` was set to TypeScript 6.0.3 source `050880ce59e30b356b686bd3144efe24f875ebc8`, verified by `git rev-parse HEAD`.

Each compiler was built from its own SHA and its own pinned submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`, TypeScript-go `d92d9bfee114c80be2c375d72edae966176e3a4f`. Both compiler worktrees were clean before measurement. The identical shared adapted tree came from the parent's apply.sh, product key `e7b65b95e17a4f32a70bb0cb987f3c27056029d053a3eed60eb571c3cc70af25`. `source-manifest.json` hashes every adapted src file. `patch-set.md` is the actual apply output. `binaries.sha256` identifies the two census executables.

Box CPU quota: 4 CPUs; memory limit: 16 GiB. Measurement ran alone after setup and both builds completed. Preparation builds overlapped, so their elapsed times are not isolated compiler performance comparisons.

| Child | Own exit | Elapsed seconds | Peak RSS KiB |
| --- | ---: | ---: | ---: |
| Parent apply | 0 | 1310.184 | 2473852 |
| Parent overlay/build wrapper | 0 | 678.967 | 2364420 |
| Head overlay/build wrapper | 0 | 675.184 | 2283552 |
| Pair tool tests, 20 tests | 0 | 52.419 | 82492 |
| Parent first-error entry census, intentionally stopped | -15 (SIGTERM) | 571.018 | 1330972 |
| Head first-error entry census | not started | unavailable | unavailable |
| Full-mode checks | not started | unavailable | unavailable |

`run.py` records actual child wait status, monotonic elapsed time, CPU time, and wait4 peak RSS in KiB. The census is a single executable, so its peak is the child's own RSS. Preparation wrapper RSS is the wait4 high-water value, not a sum of simultaneous descendant memory. The shell-facing exit for the terminated census wrapper was 143; the actual census child exit is preserved as -15. Logs are files, and no run log was piped into head or tail. Every started task was completed or terminated; the census process group was cleaned and reaped.

## Publication

The post-apply checkpoint was pushed to `typescript/view-guard-roads-census-checkpoint` at `0793e330595777c8ff08818d50afb95b45337ab7`. The stopped-run checkpoint at `a890553a448c0febc50efce538fdb6437f3097c8` preserves the coverage diagnosis. The final branch `typescript/view-guard-roads-census` has one report-only commit directly on `769b0e53`, touching only this directory. Neither main nor either source branch is pushed. No force push is used.
