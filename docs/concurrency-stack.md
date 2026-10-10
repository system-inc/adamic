Merged the four requested branches onto 0f67284 and then current main e8ba3d5.
The scheduler keeps both moved ownership and adaptive grain, with the race hooks intact.
All thirteen runtime controls passed 50/50; all five runtime races were caught on 150/150 executions.
Both moves race witnesses passed 50/50; all fifteen accepted sources passed TSan at one and default threads.
Async and host promises were left out; the full gate hit environment/resource failures recorded below.

## Branches and commits

| Branch | Fetched tip | Stack merge | No-op |
|---|---|---|---|
| codex/race-checks-reliable | abd12e246b5d37080e022832c35b0d18377777e9 | 6191992 | no |
| cloud/grok-parallel | bbf9defb2e5932f9ec8ddf6af8518f78585b9bba | 35caece | no |
| codex/concurrency-moves | 8aa7b628f97e306986a689e34a68bc9f387440b5 | efb1652 | no |
| codex/fuzz-moves | 441bfce9a1699cf13eccf6abe3556652e8ed7735 | 3fbe4a1 | no |
| origin/main | e8ba3d5d81de4d3773c723914fccd4c76248b965 | 4a12618 | no |

The pushed branch is codex/concurrency-stack only. No area branch or main was changed.

## Conflicts

The moves merge conflicted only in parallel.c's scope initializer. It now initializes
`.moved = moved` and `.grain = adamic_parallel_grain(items->length, thread_count)`.
Both APIs use adaptive ranges. The move API preserves exclusive, plain-counted item
and result graphs; the ordinary API shares them. Publication and claim hooks remain.
The merge commit explains this resolution.

Bringing the stack up to main produced eight conflicts, resolved in 4a12618:

- fresh.go: keep main's conservative closure-target operand interpretation and the parallel-map effects case.
- fuzz_test.go: retain the generated parallel-refusal contract and main's available regex support.
- generate.go: retain parallel and moves alongside inheritance, map keys, regex and bitwise features.
- prelude.d.ts: retain parallelMap alongside the checker-bridge declarations.
- refusals.go: run parser suppression refusal, then parallel preflight.
- adamic.h: retain static-class dispatch before the atomic packed field cache.
- counts.md: retain both fixture sets; regenerate separately.
- oracle_test.go: retain both fixture registration sets.

Main also introduced old cache fields in code Git merged without conflicts. Static,
accessor, optional-field and numeric-view helpers now use packed caches. Metadata
indices come from local slot pointers; they never reread a cache another worker can
replace. Optional absence and oversized-slot misses remain correct. The merge message
records these adaptations. string_append.c retains adamic_reference_count and
`length + added <= capacity`.

## Separate fixes

- 5f8eb5e: regenerate Linux counts. Only twelve rows moved in order; no numeric count changed.
- b3fbaf5: give TSan oracle executions a five-minute bound. The loaded one-thread benchmark
  had been killed at the ordinary 60-second bound, returning -1 with empty output. Other
  variants retain one minute. Three fresh executions, exact Node comparisons, race rejection
  and process-group cancellation are unchanged.
- 9380b4b: search 300 seeds for the parallel fuzzer's required witnesses. Main changed the
  generator stream; the first captured-let witness is now seed 114, outside the old 1..80
  window. The actual refusal and both matcher mutants remain required.
- 0a5f703: publish the executable cache-test wrapper before ambient parallel tests fork.
  A loaded run hit ETXTBSY. Avoiding inherited writable descriptors is the reason for the
  isolation; the original 32 simultaneous cache callers, one compilation, identical archives,
  32 links/runs and no temporary entries remain required.

## Proofs

Raw runtime observations are in concurrency-stack-mutants.log; structured rates, all
100 moves race reports, fixture/thread pairs and runtime hashes are in concurrency-stack-results.json.
No runtime or compiler source changed after the main merge while these proofs ran.

| Runtime control | Checks caught | Detector |
|---|---:|---|
| skip_items_share | 50/50, 150/150 executions | TSan data race |
| lazy_cache | 50/50, 150/150 executions | TSan data race |
| plain_shared_count | 50/50, 150/150 executions | TSan data race |
| field_cache | 50/50, 150/150 executions | TSan data race |
| remote_free | 50/50, 150/150 executions | TSan data race |
| result_order | 50/50 | ordered harness comparison |
| reused_graph | 50/50 | graph lifecycle assertion |
| oversized_slot | 50/50 | slot assertion |
| fixed_grain | 50/50 | adaptive-grain assertion |
| eager_strings | 50/50 | lazy-cache assertion |
| one_worker | 50/50 | no-worker assertion |
| pointer_guard | 50/50 | required panic status/message |
| loser_free | 50/50 | LeakSanitizer |

The 50-repeat runtime command took 1713.708 seconds under overlapping gate load.
The moves aliased-element and flattened-nested-graph compiler mutants each produced
TSan data races and exit 66 on all 50 executions. The existing moves prover also caught
use-after-move, wrong fix, wrong nested path and gratuitous sharing. Plain/shared graph
counts were 2050/0 for the control and 1/2049 for the sharing mutant.

All 30 exact moves refusal messages/fixes remain byte-identical to the moves branch
and passed. TSan covered all 13 accepted parallel fixtures, the accepted moves fixture
and parallel_files.a, three fresh processes at each of one and default threads, each
matching Node. Focused packages took 7.492, 223.880 and 153.170 seconds respectively.
Benchmark execution batches alone took 102.77 seconds at one thread and 39.47 at default.
The corrected fuzz witness test, including its matcher mutants, passed in 48.466 seconds.

## Commands and final gate

All test output went directly to /workspace/concurrency-stack-*.log files.

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts
go test -count=50 -timeout 30m -v -run '^TestParallel(ChecksCatchMutants|ScalingGuardMutants)$' ./internal/native
python3 internal/oracle/testdata/moves/prove.py
go test -count=1 -timeout 15m -v -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^concurrency$/^accepted$/.*$/^tsan$' ./internal/oracle
go test -count=1 -timeout 15m -v -run '^TestMovesAgreesWithNode$/^tsan$' ./internal/oracle
go test -count=1 -timeout 15m -v -run '^TestNativeAgreesWithNode$/^bench$/^parallel_files.a$/^tsan$' ./internal/oracle
go test -count=1 -timeout 10m -v -run '^TestParallelRunnerAgreesAndCanFail$' ./internal/fuzz
go test -count=50 -timeout 10m -run '^TestRuntimeCacheConcurrentBuilders$' ./internal/native
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -parallel 1 -timeout 30m ./stage1/cohere/json
npm install --prefix /tmp/adamic-markdown-width --ignore-scripts --no-audit --no-fund emoji-regex@10.6.0 get-east-asian-width@1.6.0 narrow-emojis@0.0.3
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 15m -v -run '^TestMarkdownUnicodeWidths$' ./stage1/cohere/markdownblocks
```

Setup: Go, clang, Node and submodules each 0 seconds; warm cache 13 seconds; total
13 seconds. nproc=5, cgroup cpu.max=400000 100000. Linux amd64, clang 20.1.8,
Go 1.27.1 and Node 24.19.0. Gate runs overlapped at host load around 7 to 21.
Formatting and vet passed with empty logs. Counts regeneration passed in 133.308 seconds.

The full uncached gate finished with exit 1. Native passed in 248.403 seconds,
oracle in 773.702, lower in 34.979, fuzz in 101.366, and typeaware in 1292.572.
All packages other than JSON and markdownblocks passed. JSON passed on retry.
Markdown Unicode widths first failed because its documented scratch npm packages
were absent. The package then reached the 30-minute bound with CodeBlockLayout,
LeafComposition, RootLayout, StructureLayout and TableLayout still running. That
whole package is not claimed green; its width oracle was retried separately after
installing the three documented pinned packages. The isolated width retry passed all 1,217,505 cases and its three output mutants in 265.347 seconds. The five timed-out Markdown layout tests were not rerun; the full gate is not claimed green.

Cache publication passed all 50 repetitions in 224.309 seconds, retaining 32
simultaneous builders in each repetition.

The unrestricted full gate hit `clang: error: unable to execute command: Killed`
in stage1/cohere/json/TestSingleFileStdoutDriver. The cgroup recorded one OOM kill
and a peak of 17,180,229,632 bytes against a 17,179,869,184-byte limit. Memory
pressure is the supported explanation for the compiler termination. The rest of
that gate was allowed to finish. JSON was retried uncached with `-parallel 1`,
without changing source or removing any tests. Its result is recorded below.
JSON serial retry: exit 0, 783.916 seconds. All tests ran uncached; no source changed.

An environment refresh stopped the first final gate and cache stress run before either
finished; both were restarted. Earlier superseded gates exposed the documented failures,
so none is claimed as passing. Finished mutant and fixture evidence survived the refresh.
macOS/arm64 and a new performance measurement were not run. The 300-seed generation
coverage and default move campaign are exercised by package tests; a fresh full 300-seed
execution campaign was not requested or run. Async and host-promises branches were excluded.
