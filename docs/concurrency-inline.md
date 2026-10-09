# Unshared reference count follow-up

Both requested branches were merged, not rebased: `b19a10e` has parents `9e27014` and `b9f22f0`; `be8b28e` has parents `b19a10e` and `220a915`. The release conflict was resolved by separating the last-reference draining frame and returning before touching the freeing queue otherwise, preserving atomic shared release and remote frees.

`f352efb` makes `adamic_retain` and `adamic_release` static inline in adamic.h. Positive unshared counts are read and updated plainly; nonfinal releases return immediately. Shared, immortal, NULL and last-reference cases go to out-of-line helpers. Count hooks occur once in the header, including fast operations; helpers do not duplicate them. Compiler-generated C names and the parallel ABI are unchanged.

A permanently published bit in the existing slab word dispatches before any plain count read. This is necessary: a plain read of the shared count itself would race with atomic updates. The count tag remains for shared atomic operations and region immortality. Sharing sets both bits before publication, and deallocation masks the slab bit. The header remains 16 bytes. Inspecting clang -O2 assembly of wrappers showed plain loads/stores and branches on the positive path, with no call, TLS access or atomic instruction; only other paths tail-call the slow helpers. Assembly is in `/tmp/adamic-inline-probe.s`.

## Performance

Same toolchain, main at `5d4c801`, interleaved and alternating first version each round, best of five. Baseline noise is two further interleaved sets of five executing the *same* main binary; the last column is the absolute difference between their best times, divided by set A's best. This is an observed noise estimate, not a statistical confidence bound. Every sample, output and count line is committed in `internal/native/testdata/parallel/inline-measurements.json`.

Load before/after the maps, programs and baseline noise runs: **2.40/2.84/2.00** and **1.67/2.57/1.95**. Retain measurement load before/after: **1.70/2.56/1.95** and **1.64/2.54/1.95**. Machine: Linux x86_64, `nproc` 5, quota four CPUs. A separate counted retain probe ran briefly during the program-noise phase; it did not overlap the main-versus-branch program timing phase. Prior native/oracle gates had finished before benchmarking.

| Program, seconds | main | branch | Change | Measured main/main noise |
| --- | ---: | ---: | ---: | ---: |
| word_count | 0.157568391 | 0.145805331 | -7.465% | 5.213% |
| tokenizer | 0.159572153 | 0.159862530 | +0.182% | 1.589% |
| trees | 1.613995452 | 1.610088834 | -0.242% | 2.162% |
| sort | 0.208997675 | 0.208533120 | -0.222% | 1.772% |
| retain.c, 50 million pairs | 0.183127147 | 0.041541617 | -77.315% | 1.034% |

The previous regressions are gone in this run. Retain/release is 77.315% faster and word count 7.465% faster than main. Tokenizer's +0.182% is below measured noise; trees and sort are slightly faster, with changes below noise. Inline code permits the optimizer to simplify redundant retain/release pairs; retain.c's speedup is the actual compiled loop's time, not a promise that every isolated operation has that speedup. There is no timing assertion in the tests.

The runner also repeated the existing maps: million-number map at 1/2/4 threads was 7.091/5.333/4.734 ms; string map was 137.260/70.602/37.557 ms. Raw data also preserves the extra Weak lookup measurements. This follow-up's requested performance bar covers retain.c and the four specified bench programs; contended Weak throughput remains unmeasured.

Outputs and all count fields match main for word_count, tokenizer, trees and sort. A separate `-DADAMIC_COUNT` build of retain.c produced exactly `allocations 1 frees 1 retains 50000000 releases 50000001 peak 1 regions 0`, confirming hooks on the inline path. The counted build's wall time was not used for the release timing comparison.

Reproduce:

```sh
source /workspace/adamic-tools/env.sh
python3 internal/native/testdata/parallel/measure.py /tmp/adamic-concurrency-main /tmp/adamic-inline-measure --programs word_count,tokenizer,trees,sort --noise > /tmp/adamic-inline-measure.log 2>&1
# exit 0; every output and count comparison passed
```

## Gates and mutants

Parallel harnesses now run ASan+UBSan with malloc, ASan+UBSan with sanitized size classes, counted release, explicit Malloc release, and Linux TSan. All controls passed, including nested scopes, exceptions, string caches, fresh results, remote frees and one-thread output comparisons.

```sh
go test -v -count=1 -timeout 15m ./internal/native -run TestParallel > /tmp/adamic-inline-parallel.log 2>&1
# PASS, 25.037s
go test -v -count=1 ./internal/native -run TestReleaseSharedValueAndUnsafeMutant > /tmp/adamic-inline-release-mutant.log 2>&1
# PASS, 10.201s; ASan catches heap-use-after-free
go test -count=1 -timeout 30m ./internal/native/... > /tmp/adamic-inline-native-final.log 2>&1
# PASS, 93.216s
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -parallel 4 ./internal/oracle > /tmp/adamic-inline-oracle.log 2>&1
# PASS, 158.594s; includes the merged sanitized-size-class oracle variant
go vet ./... > /tmp/adamic-inline-vet-final.log 2>&1
# exit 0, no diagnostics
git diff --check
# exit 0; gofmt clean
```

All seven concurrency mutants ran and were caught: skip sharing and plain shared counting by TSan (exit 66); non-TLS cache and direct foreign free by TSan (exit 66); reversed results by index comparison (exit 3); reused graph and missing one-thread path by harness abort. The merged release mutant also compiles and is caught by ASan on a dynamic string, compared with Node. The first broad native run failed because that merged mutant's old source anchor no longer existed. The test now routes count two through the slow path and makes the same premature-free mistake there; the complete rerun passed. No build failure or timeout was accepted as a caught mutant.

Setup on this resumed environment completed in 57s: Go, clang, Node and submodules each 0s, cache warm 57s; Go 1.27.1, clang 20.1.8, Node 24.19.0. No lowering or JavaScript edits were made. The parallel oracle variant and source-level parallel lowering remain the agreed third pass. macOS was not rerun, and no whole-repository go test gate was run; native, oracle and whole-repository vet were run.
