Merged stack b's shared Go build cache into the bridge unit split for roadmap step 38.
Delivery: merge commit containing this report, compiler/test-split-bridge; parents 54c980ec and b05fa5c9.
Checks: see recorded unit, vet, coverage and cache results below.
Mutants: existing planted-query, coverage, cache and budget probes, plus an archive-key rebuild mutation.
Not covered: full repository gate, other platforms or remote-store transfer.

The only conflict was bridge/tsgo/bridge_test.go: stack b changed the combined
TestBridge while this branch had replaced it with independently selectable units.
Kept all 36 registered pieces and moved stack b's archive, compiler and oracle
cache routing into TestMain's shared preparation. Each process retrieves these
four products using internal/buildcache.GoInputs and Get, with the same names,
reproducible flags and sanitizer environment as GoBuild. TestMain has no testing.TB,
so it uses these exported entry points rather than the testing wrapper.

The ordinary and sanitized archives are individually keyed shared products.
The prepared 19-product bundle includes their bytes and keeps the overlay mutants
keyed together. Every unit retrieves the shared Go products before retrieving the
bundle; no archive build occurs in a timed unit. The budget fails only with
ADAMIC_UNIT_BUDGET=1; all five-minute hang guards remain unchanged. The synthetic
cache verification test explicitly disables the remote store so fake products
cannot be published or retrieved there.

Local validation uses GOMAXPROCS=4, GOFLAGS=-p=2, ADAMIC_GATE_UNCACHED=1,
ADAMIC_UNIT_BUDGET=1 and a new ADAMIC_BUILD_CACHE_DIR. ADAMIC_BUILD_STORE=off
isolates this build-count measurement from remote transfers. Every top-level
bridge test is selected separately in sample and compiler-corpus modes.
The corpus is TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8.
No .a fixture or oracle count changed in this resolution.

Commands (each redirected to a log; all completed with exit 0):

```
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct' GOMAXPROCS=4 GOFLAGS=-p=2
bash cloud/setup.sh
go vet ./bridge/tsgo/...
go vet ./...
ADAMIC_TSGO_PREPARE=1 go test ./bridge/tsgo -run '^$' -count=1
python3 review/test-split-bridge/stack-b-merge/validate.py
python3 review/test-split-bridge/mutants.py
python3 review/test-split-bridge/stack-b-merge/cache-mutant.py
```

The validation script selects each actual top-level test with go test -json,
-count=1 and -timeout=5m. An initial discovery attempt also selected TestMain;
it exited successfully without a test event, so discovery was corrected to exclude
the harness and resumed without repeating the completed units. Both attempt logs
are included. The original mutants.json is preserved; new results are beside this report.

| Mode | Passed | Skipped | Maximum test body | Maximum invocation including shared lookup |
| --- | ---: | ---: | ---: | ---: |
| Sample | 19 | 20 | 0.180s | 8.042s |
| Compiler corpus | 34 | 5 | 1.260s | 5.803s |

Coverage retains 36 registered pieces and bindings, 16 sample or 31 compiler
active pieces, 162 sample or 1600 compiler positions, five analyses per position,
and exactly one owner for every active piece at each shard count 1..40.

The fresh-cache ledger records one ordinary archive miss (key 0f09cee8eac4)
and one sanitized archive miss (key d9968ae9162a), followed by 47 hits each:
one for every active original piece run in a separate process across both modes.
The compiler and oracle likewise have one miss each. The 19-product bundle has
one miss and subsequent hits. Product preparation is outside unit timing.
archive-counts.json records the independently asserted counts before probes.

The changed native query is caught only by TestBridgeOracleSample in shard 3/4;
shards 0/4, 1/4 and 2/4 pass. Missing a piece, query or shard filter fails the
coverage pin. Accepting corrupt bytes or extra products and rebuilding a cached
bundle fails the cache pin. A planted 31-second elapsed duration fails only with
ADAMIC_UNIT_BUDGET=1; budget=0 logs it and passes. A poisoned bundle callback
still passes after preparation, proving reuse. All 13 existing probe runs produce
the expected outcome. Original input/output-length, stale-handle, wrong-position,
linkage, output-free and region-ownership mutants remain caught by the units
in both sample and compiler modes, including ASan/UBSan/LSan observations.

The additional archive-key mutant appends the process ID to the ordinary archive
key through a Go overlay. Two independently selected ABI units pass, but their
two archive misses violate the same one-miss assertion used for the healthy run.
This proves the build-count assertion detects loss of sharing without treating
compiler failure as its catcher. See cache-mutant.json and its two event logs.

Setup passed; nproc=5, cpu.max=400000 100000 (four CPUs). Timing lines:

```
v24.19.0
go version go1.27.1 linux/amd64
setup: node ready (0.030s)
setup: go ready (0.032s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.009s
setup: submodules ready (0.086s)
setup: markdown dependencies ready (0.087s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.178s)
setup: go build ready (61.018s)
setup: test binaries deferred (use --warm-tests) (61.311s)
setup: build cache warm (61.312s)
setup: build-flags commit=54c980ec6b210b47fd0bae914832da10d75802bd nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=0.00 0.22 0.62 1/253 288400 load-after=5.02 1.62 1.08 6/283 291910
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (61.344s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.tG4fMt
```

No main or area branch was pushed. The merge preserves both parent histories.
