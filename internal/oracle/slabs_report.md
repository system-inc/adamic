Exported `native.Options.Slabs` and added selectable `TestNativeSlabsAgreeWithNode`. The malloc comparison, release comparison and leak check in `TestNativeAgreesWithNode` are unchanged. The slab lane runs every lowerable fixture, including panic and inserted-check paths. No allocation-heavy subset was needed. Native observations are always fresh, so no new result cache was added. Node observations use the existing harness cache, bypassed by `ADAMIC_GATE_UNCACHED=1`.

All 322 lowerable fixtures from origin/main `c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06` passed uncached with sanitized slabs. Fixtures flagged: none. There are no first ASan error lines to list. Runtime sources and emitted ownership logic are unchanged from that main commit.

Build-flags line for every timing below: measurement commit `04774faf957145981a79959c69e8b9bdb5cefaba`; nproc 5; cgroup cpu.max `400000 100000`; Go `go1.27.1 linux/amd64`; clang `20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; Node `v24.19.0`; oracle observations uncached, Go build cache warm, runtime libraries warm and separately keyed by complete flags and runtime contents. Both binaries use `native.Flags(Options{Sanitize:true})`: C11, warnings as errors, `-ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`; slabs additionally uses `-DADAMIC_SLABS`. Both execute with `ASAN_OPTIONS=detect_leaks=0`. Four parallel fixture workers. Loads below are 1/5/15 minute averages.

| Loop | Before: malloc seconds | After: slabs seconds | Instrument |
|---|---:|---:|---|
| 1 | 31.136718 | 31.823665 | M then S |
| 2 | 30.073828 | 30.107482 | M then S |
| 3 | 30.113199 | 31.978085 | M then S |

Exact instruments, each redirected to a separate log:

```sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_SLAB_MEASURE=1 go test -count=1 -timeout 30m -parallel 4 ./internal/oracle -run '^TestNativeMallocAgreesWithNode$' > /tmp/adamic-slab-time-N-Malloc.log 2>&1
ADAMIC_GATE_UNCACHED=1 ADAMIC_SLAB_MEASURE=1 go test -count=1 -timeout 30m -parallel 4 ./internal/oracle -run '^TestNativeSlabsAgreeWithNode$' > /tmp/adamic-slab-time-N-Slabs.log 2>&1
```

M is the first command; S the second. The instrument is Python `time.monotonic()` around each complete command, including Go startup/build. Raw records preserve exact seconds, commands, exit statuses and full load readings in `slabs_evidence.jsonl.gz`.

| Loop | Lane | Load before | Load after |
|---|---|---|---|
| 1 | malloc | 2.92 3.35 1.54 | 3.48 3.44 1.63 |
| 1 | slabs | 3.48 3.44 1.63 | 4.14 3.60 1.76 |
| 2 | malloc | 4.14 3.60 1.76 | 4.38 3.71 1.85 |
| 2 | slabs | 4.38 3.71 1.85 | 4.55 3.82 1.95 |
| 3 | malloc | 4.55 3.82 1.95 | 5.04 4.00 2.07 |
| 3 | slabs | 5.04 4.00 2.07 | 5.19 4.14 2.18 |

Best of three: malloc 30.073828 s, slabs 30.107482 s, difference 0.033654 s (+0.112%). Run variation is larger than that difference. The new all-fixture shard costs about 30 seconds on this box. This is a comparison-only malloc control with the same work as the slab shard, not the complete existing oracle: the existing test also runs the JavaScript backend, release build and LeakSanitizer. It remains necessary for leaks. This experiment does not measure scheduling savings from assigning the new shard to another worker.

Three proofs were run:

* `TestSlabLaneCatchesEarlyRelease`: `string_append.a`'s runtime-built `doubled` loses its count before the self-append. Slabs reports `ERROR: AddressSanitizer: use-after-poison`; malloc reports `ERROR: AddressSanitizer: heap-use-after-free`. Ordinary early release is caught by both.
* `TestSlabLaneCatchesRecycledRelease`: a deliberately constructed allocator-dependent mutant seeds a freed seven-byte string slot before the fixture builds `doubled`. It releases `doubled` one count early only when it occupies that recycled address. Slabs reports `ERROR: AddressSanitizer: use-after-poison`; malloc's ASan quarantine prevents address recycling, and the complete fixture matches Node stdout, stderr and exit code. No compiler or allocator flag test selects the mutant branch. This proves a path dependent on allocator reuse can escape the malloc lane; it is not evidence of an existing compiler bug.
* Cache-key omission mutant: a temporary Go overlay filters `-DADAMIC_SLABS` out of `runtimeKey`. `go test -count=1 -overlay=/tmp/adamic-slab-key-overlay.json ./internal/native -run '^TestRuntimeKeyIncludesEveryInput$'` fails with `omitting -DADAMIC_SLABS did not change the key`. The existing library key already covers the exported option through `Flags`; library.go did not need a change. The mutant is absent from committed code.

The existing `TestFreedValuesAreCaughtWithSlabs` also passed. Freed slots already poison their whole size, including the reference count; no heap.c change was needed. Important limit: `take` unpoisons a recycled slot before using it. ASan cannot distinguish an old pointer from the new live object at that address. The lane can catch slab-specific free-slot errors and allocator-dependent paths, and output comparison can catch observable corruption, but it does not guarantee detection of a stale read after reallocation.

Setup printed Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 110s, done 110s on 5 processors (quota four CPUs). `/workspace/adamic-tools/env.sh` was sourced for every toolchain command.

Validation: gofmt, git diff --check, native package uncached, oracle slab lane uncached over all fixtures, both mutant proof tests uncached, and repository go vet. The complete oracle package passed, and the original string_append comparison plus both persistent mutant proofs passed uncached. The full repository test gate and specialized input/output/weak test harnesses under slabs were not run. The slab lane covers the ordinary fixture list, including the named runtime and sweep fixtures.

Final validation commands, all exited 0:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/native > /tmp/adamic-slab-native.log 2>&1
go vet ./... > /tmp/adamic-slab-vet-all.log 2>&1
go test -count=1 -timeout 30m ./internal/oracle > /tmp/adamic-slab-oracle-package.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/string_append.a$|^TestSlabLaneCatches' > /tmp/adamic-slab-filtered.log 2>&1
gofmt -l cmd internal > /tmp/adamic-slab-fmt.log
git diff --check
```

The complete oracle package used the existing observation cache; its slab executions always rebuilt and ran. The all-fixture slab runs and filtered original comparison used uncached observations. The two proof tests are checked in for reruns without overlays. The full unmutated slab-run log in the evidence names all 322 fixtures. Setup's own 110-second status was retained, but setup load averages were not captured; it is not a before/after benchmark.
