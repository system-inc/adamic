# Retain/release inline path — October 8, 2026

The production stage 1 parse build now measures **6,627,967,853 instructions**, below the requested 6.644G ceiling. The starting runtime measures **7,293,289,097** on this machine: a reduction of **665,321,244 (9.12%)**. The exact full AST bytes match the independent pinned Go parser.

Base: `56d53e10bf0b83c953e156687f60531fb17945fe`, fetched by the requested `runtime/area-take-leftovers` refspec. Work branch: `runtime/retain-fast-path`. The requested `runtime/parse-bisect` branch was fetched by name and its report read. No AGENTS.md was present in the repository or workspace ancestors.

## What was going slow

There is no accidentally set slab flag on the ordinary parse values. The starting `ADAMIC_SLOW_COUNT` contains the shared and graph bits, not the statement-region bit. Statement regions have zero references. Comparing temporary counting builds with the pre-graph shared-only mask and the current mask gives identical dispatch counts, and zero graph/shared-flagged calls. No owned cell is used by this corpus.

The hot values are **immortal strings**, whose zero reference counts fall through the old inline wrappers. The slow helper redirects owned cells, checks graph membership, then updates ordinary or atomic counts; for these strings it ultimately does nothing. There are 20,852,831 such retains and 16,393,340 such releases. The extra cell-owner and graph support made that no-op work more expensive and changed ThinLTO inlining. The problem is not ordinary positive counts acquiring graph flags.

Temporary probes at each inline wrapper record the exact slow-dispatch predicate before count mutation. They use separate runtime snapshots and are absent from instruction measurements and production code. Per-kind slow calls on all 77 files:

| Kind | Retains before | Retains after | Releases before | Releases after |
|---|---:|---:|---:|---:|
| NULL | 0 | 0 | 3,529 | 0 |
| string | 20,852,831 | 0 | 16,678,467 | 285,127 |
| object | 0 | 0 | 199,795 | 199,795 |
| array | 0 | 0 | 475,435 | 475,435 |
| map | 0 | 0 | 2 | 2 |
| cell | 0 | 0 | 0 | 0 |
| closure | 0 | 0 | 154 | 154 |
| total | 20,852,831 | 0 | 17,357,382 | 960,513 |

All other kinds have no wrapper calls in this corpus. The 960,513 remaining slow releases are last-reference releases. Ordinary cells and closures already took the positive-count inline path; they were not the regression's hot slow callers.

## Change and ownership

`internal/native/runtime/adamic.h` makes both small wrappers `always_inline`, returns immediately for NULL, and returns for zero-count non-cell values after checking the existing shared/graph dispatch mask. This keeps static-string zero counts visible to callers even when the slow helper grows. The count instrumentation macros still execute first.

Graph members still dispatch before any plain count read. Shared counts still use the atomic slow path. Statement-region values retain their immortal behavior, including the shared-region mask check. Owned cells still reach the unchanged slow helper and redirect to their environment or async frame; positive-count standalone cells retain their existing inline behavior. Ordinary last-reference releases still use the unchanged destruction queue.

No `heap.c` change is needed: the child-destruction extraction is not the source of the recovered instructions. Self costs from complete `callgrind_annotate --show=Ir --threshold=100 --auto=no` output:

| Function | Starting self Ir | Final self Ir |
|---|---:|---:|
| adamic_retain_slow | 375,350,958 | 0 |
| adamic_release_slow | 279,582,714 | 17,289,234 |
| adamic_retain | 90,209,103 | 0 |
| adamic_release | 97,231,372 | 0 |
| adamic_heap_free_children | 241,180,584 | 241,180,584 |
| let_go | 98,257,502 | 98,257,502 |
| release_last | 109,864,976 | 109,864,976 |

Some inlined costs increase in callers; gross helper reductions are not summed as the net result. All raw self records reconcile to the measured total.

The requested async comparison uses `2eea315c9`'s parent `c4eb37d981885fa187a9a1fa14371cea58fd2d75` and `7c4dae857cc13b334e699880a5af23b2b7a83c15`, with the same current emitted C and flags on both runtime snapshots. Both ASTs equal Go's. This isolates runtime changes rather than rebuilding the historical compiler and dependency pins.

| Runtime | Ir |
|---|---:|
| async parent, c4eb37d9 | 6,644,639,177 |
| async merge, 7c4dae85 | 6,865,831,055 |
| difference | +221,191,878 |

The report's historical difference was +221,189,341. The complete fresh annotations reproduce its attribution: `let_go` +99,749,097, `release_last` -46,801,631, `Parser_kind` +28,420,977, `ParseNode_new` +22,228,651, `adamic_release` +18,110,150 and `adamic_release_slow` +7,306,397 self instructions. Source comparison confirms the added cell-owner redirect. Handling immortal values in the inline wrapper removes their need to execute that redirect check; it also restores caller inlining. The recovered total exceeds the combined endpoint regression without changing ownership or child destruction.

## Measurement

Machine: Linux x86_64, Intel Xeon Platinum 8573C, 5 visible CPUs, cgroup `cpu.max=400000 100000` (4 CPU quota), 17.6 GB memory. Setup: `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh --wasi-sdk`, then source `/workspace/adamic-tools/env.sh`. Node **v24.19.0**, Go **1.27.1**, clang/LLVM/lld **20.1.8**, WASI SDK **27**, Valgrind/Cachegrind **3.27.1**. The submodule's redundant history fetch was stopped; its exact pinned commit was shallow-fetched before rerunning setup successfully. Valgrind was extracted into scratch from Debian's `valgrind_3.27.1-0.2_amd64.deb`.

Corpus: batch 8's 77 compiler files, TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`, in `stage1/profiles/benchmarks.json` order. Every file's SHA256 was checked. The emitted C is byte-identical before and after:
`866e672cf5073a12b6474a0389182eaf65713b7baca2f742e7d57622e37871ac`.

An independent adapter overlays `stage1/typescript/parser/testdata/oracle.go` into the pinned Go parser and runs `--manifest compiler.txt --whole`. Every instruction-measured binary runs `--manifest compiler.txt --ast` and matches all **44,766,682 bytes**, SHA256 `8ae015600498b915cc25abab82730299451ae990b50478980d5a3bc465801bfe`, with empty program stderr.

Final production build:

```sh
go run ./cmd/adamic-stage1 -driver parse -policy thin -emit shipping.c -o shipping
VALGRIND_LIB=tools/usr/libexec/valgrind tools/usr/bin/valgrind \
  --tool=cachegrind --cache-sim=yes --branch-sim=yes \
  --I1=32768,8,64 --D1=32768,8,64 --LL=268435456,1,64 \
  --log-file=shipping-cachegrind.log --cachegrind-out-file=shipping.cachegrind \
  shipping --manifest compiler.txt --count
```

The shipped flags apply to the runtime, driver and link: C11 and warning flags, `-ffp-contract=off -fno-optimize-sibling-calls -pthread -O2 -flto=thin`, with `-fuse-ld=lld` at link. No CPU override, training profile, sanitizers or `ADAMIC_COUNT` is used in instruction measurements. `--count` is the driver's output mode. Every profiled output is exactly `0\n` with empty program stderr. All 13 Cachegrind event totals reconcile to raw self costs using `stage1/profiles/measure.py`'s accounting checks.

| Build | Ir | Load before (1/5/15 min) | Load after |
|---|---:|---|---|
| starting runtime | 7,293,289,097 | 7.77 / 9.27 / 4.61 | 3.18 / 7.56 / 4.35 |
| scratch candidate | 6,627,967,839 | 1.74 / 5.37 / 3.95 | 1.91 / 4.71 / 3.82 |
| async parent | 6,644,639,177 | 1.74 / 5.37 / 3.95 | 1.91 / 4.71 / 3.82 |
| async merge | 6,865,831,055 | 1.60 / 4.41 / 3.74 | 4.36 / 4.47 / 3.82 |
| final production command | 6,627,967,853 | 3.76 / 4.35 / 3.80 | 1.96 / 3.70 / 3.61 |

The scratch candidate and async-parent profiles overlapped; the final profile overlapped targeted tests. These are simulated instruction counts, not elapsed-time or hardware-cycle claims. The fresh starting count is 169,582 instructions above the previous report's 7,293,119,515; no claim is made that this small cross-session difference is fully attributed. The production-command build is 14 instructions above the scratch candidate.

Final measured binary SHA256: `0668782518f05def7d6e22d7416364960ed864ed3ea4b74aaab98c8b4ec458f4`.

## Tests and mutants

Targeted runtime tests passed (43.648 seconds):

```sh
go test -v -count=1 ./internal/native -run '^(TestGraphMembersCountOnTheirRegion|TestEnvironmentCellsCountTheirEnvironment|TestGraphValueIsNeverShared|TestGraphRegionsRuntime|TestGraphClosureEnvironment|TestReleaseSharedValueAndUnsafeMutant|TestRuntimeReleasePaths)$'
```

The two existing mutants compile and are caught:

- `TestGraphMembersCountOnTheirRegion`: replacing the graph/shared mask with the shared-only mask sends a graph member inline; allocation/free imbalance catches the unfreed region.
- `TestEnvironmentCellsCountTheirEnvironment`: removing `counted_heap`'s cell-owner redirect causes the retained interior cell to outlive a freed environment; ASan catches heap-use-after-free.

Targeted oracle tests passed (96.161 seconds), including 96 selected native fixture cases, the normal/release/slab sanitizer comparisons, leak checks, accepted concurrency TSan variants, graph counts, concurrency refusals and specialized async mutants:

```sh
go test -v -count=1 -parallel=4 -timeout=30m ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(graph_regions_.*|async_.*|concurrency)$|^(TestGraphRegionsCountsAndFree|TestGraphRegionsCompiledMillion|TestGraphAllocationFlowIsLeakClean|TestGraphParallelMapRefusesRegions|TestConcurrencyAgreesWithNode|TestConcurrencyRefusals|TestAsync.*)$'
```

No fixture or allocation-count expectation changed, and no mutant was added; `counts.md` requires no regeneration. No whole-package test run was performed. `git diff --check` passed.

Raw profiles, complete annotations, per-kind counter outputs, full ASTs, emitted C, binaries, runner scripts, exact commands/load records, accounting JSON and test logs remain in `/workspace/scratch/retain-fast-path/`. Only the runtime header and this report are committed.
