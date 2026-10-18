Built: Views fix-forward C recut as one commit on current main, retaining both sides of the counts conflict.
Commits: source 54bbf5860566ea6042cd1e86655ca72d5623c64f; base 67733635b0252eb9d3bb3645c72262c7b18361e1; delivery SHA is the pushed compiler/loom-r-views-fx-c-ac-m tip.
Commands and outputs: full requested compiler package trees, fixture consumers, guards, vet, gofmt and Linux counts passed; commands and exits below.
Mutants: eleven semantic faults caught, one green isolation control passed, and both header predicate mutants caught.
Not covered: full repository gate, all other compiler-reader packages, opt-in lanes in skips.json, ahra/PR/Loom submission or queue changes.

Fetched newest origin/main 67733635b0252eb9d3bb3645c72262c7b18361e1 and actual source 54bbf5860566ea6042cd1e86655ca72d5623c64f. Source parent is dbdf977f5b8b0979d414ebe6f20e7aadd312e309. Applied only the source's own commit with cherry-pick --no-commit. The requested explicit-base merge-tree is preserved in merge-tree.log.

Only internal/oracle/counts.md conflicts, in two hunks. First: main's 35 callback_widening rows from slice 2 overlap the source's eight callable-producer and four callable-union-result additions. Second: main's three weak_miscompiles rows overlap the source's two stage3 snapshot rows. Discard all conflicted/auto-merged count contents, restore the entire main table, and regenerate the full counts family once with Linux house clang 20.1.8 after compilation. Both sides' rows survive; exact resulting old/new values are in counts.json and counts.patch. No count row is hand-merged.

Runtime object.c auto-merges. Preserve main's newer stored representation cases and apply exactly the original optional-boolean payload initialization: replace value.payload.boolean = boolean.boolean; with value.payload = (adamic_value){0}; and if (boolean.present) { value.payload.boolean = boolean.boolean; }. Source-own and recut added/deleted runtime lines are byte-identical. Extra runtime resolution lines: none. No main implementation line is reverted.

Assumptions: use newest fetched main at unit start and preserve the source's complete own diff, including historical review evidence. Existing refusal headers and every .a body stay unchanged unless a current compiler observation proves a header correction is required. Historical source reports remain historical; current logs and source comparisons are authoritative. Run full package trees as the latest request explicitly requires, plus fixture consumers and guards. No new package, runtime feature implementation or IR Program map is added. No cohere implementation is copied; no PR, ahra command or Loom queue update is made. The user will manage chg_k34bzw7phmpghh1khspr187jwm.

Environment recovery: prior compiler-unit caches filled both volumes. Before new work, evicted three disposable prior-unit Go export products under /tmp/adamic-gate, then preserved one root-volume cache product by moving its contents to the temporary volume and leaving a symlink at its original cache path. No source, evidence, toolchain bytes or product key changed. space-recovery.json records the named products. New test products and Go-cache writes use main's supported ADAMIC_BUILD_OUTPUT_DIR and ADAMIC_GOCACHE_DIR. Setup ran with --wasi-sdk because current main contains an unconditional WASI native test.

Reader storage recovery: main now splits the reader check into Native, Lower, Fresh and Flow, each with its own large Go export product. The initial IR run hit ENOSPC in the first three serial archive preparations. Maintenance intended to run after oracle missed the package boundary: IR had already started, so relocating the pool also caused one temporary path lookup failure. No assertion was changed. The completed oracle suite passed before that maintenance. Retired five older reproducible root-cache export snapshots through named workspace scratch, then moved the current local product pool to /workspace/scratch/views-fx-c-products-local, keeping the original /tmp cache path as a symlink. No source, product bytes, input key or existing public path changed. The later parallel owner/checker tests built all four export products successfully; the first IR command remains red because of its earlier setup failures. Rerun the complete IR package tree with products ready, and run all four TestCallTargetReaders* guards explicitly with -timeout 90s. reader-storage-recovery.json records exact paths and retirement metadata.


All source non-count additions/deletions are identical to the source commit (source-delta.json). All 63 .a files, including review evidence, remain byte-identical to that source. No new header or positional expectation correction is required. Native runtime source and recut lines match exactly (runtime-lines.json); additional runtime C/H lines for clearance: NONE.

Counts were regenerated once from main, never hand-merged. Exactly 15 additions and 3 existing-row updates, all with source-identical final values; no removed row and no unrelated change. The existing changes are node-indicator-good.a retains 15 to 14; callable-producer-wrong.a changes from 2/0/5/4/2/0 to 3/3/6/11/3/0; undefined-read-write.a retains 5 to 4. Exact complete table delta follows below. Main's callback widening and weak-reference rows survive unchanged.

Validation uses GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1, -count=1, -parallel 4 and -p 1, with JSON output written directly to logs. Whole package commands have -timeout 12m and an 840-second outer limit; explicit reader/maps and parallel guards have -timeout 90s. No test was removed, weakened or rerun after exceeding its leaf budget. test-times.json/tsv distinguishes grouping parents from leaves and retains all measured leaf times, including deliberate mutation failures. Full oracle holds source Node, native release, ASan/UBSan/LeakSanitizer, backend JavaScript and recorded counts, including the source's producers, union results and snapshot fixtures.

Full internal/ir includes SortedMap and call-target readers; internal/lower includes native-agreement registration; internal/oracle includes exact runtime-set witnesses. TestCallTargetReaders and TestEveryMapAProgramHoldsIsSorted are also recorded with the requested 90-second guard limit. The cmd/adamic-gate parallel guard passed. Full internal/javascript, flow/fresh fixture consumers and stage3/fixtures passed. No new compiler-consuming Go package, IR Program map, runtime feature set or product recipe was introduced.

A-check: the unchanged README-pinned Gate.aCheck invokes this branch's actual compiler on every one of the 63 changed .a paths, including review files. scan.py retains actual exit statuses, full diagnostics, source hashes and the predicate's judgments. Results: {"refused": 4, "checked": 54, "NotYet": 5}. The four existing refusal headers remain correct. Removed-header and wrong-reason mutants use the README's explicitly documented replay of the current compiler's captured real diagnostic; both fail the unchanged predicate and restore the real review witness byte for byte. This is a predicate proof, not a claim to rerun compiler diagnostics on the temporary header variants.

Mutants: source's eleven corrected faults and one isolation control run as precise changes over the current files, preserving main, with every original byte restored in finally blocks. Only test assertion/behavior failures count; Go/C build errors and timeouts do not. Separate read-through Go cache writes for the mutant phase use root-volume scratch while reading immutable successful actions from the positive run's cache (mutant-cache.json), avoiding a second gigabyte export product and the earlier unit's disk-full failure. The packed snapshot fault is held directly by payload bytes; the combined retain fault is held by the Node/native snapshot fixture; the green control proves the retain guard masks a stale payload in isolation. Exact patches, commands, exits and catchers follow.

Setup: GOPROXY=https://proxy.golang.org|direct; timeout 600 bash cloud/setup.sh --wasi-sdk; source /workspace/adamic-tools/env.sh. Linux clang 20.1.8 (SDK clang 20.1.8-wasi-sdk), Go 1.27.1, Node 24.19.0; nproc=5, cpu.max=400000 100000 (four CPUs). Exact timing lines:

```text
setup: go ready (0.024s)
setup: node ready (0.023s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.010s
setup: markdown dependencies ready (0.067s)
setup: submodules ready (0.069s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.153s)
setup: wasi sdk ready (/workspace/adamic-tools/wasi-sdk, linking with /workspace/adamic-tools/bin/clang) (0.244s)
setup: shared cache ready (0.899s)
setup: go build ready (44.892s)
setup: test binaries deferred (use --warm-tests) (44.972s)
setup: build cache warm (44.973s)
setup: build-flags commit=67733635b0252eb9d3bb3645c72262c7b18361e1 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.13 0.04 0.02 1/1010 580376 load-after=3.06 0.84 0.30 1/1015 581075
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (45.002s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.sUTqH4
```

Commands and results:

- compiler-build: exit 0, 12.625s; `go build -o /tmp/views-fx-c-adamic ./cmd/adamic`.
- counts-regenerate: exit 0, 88.222s; `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 12m -parallel 4 -p 1 -json -args -update-counts`.
- lower-all: exit 0, 60.093s; `go test ./internal/lower/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- native-all: exit 0, 287.188s; `go test ./internal/native/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- oracle-all: exit 0, 439.051s; `go test ./internal/oracle/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- ir-all: exit 1, 67.101s; `go test ./internal/ir/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- ir-all-ready: exit 0, 10.13s; `go test ./internal/ir/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- javascript-all-ready: exit 0, 3.125s; `go test ./internal/javascript/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- flow-all-ready: exit 0, 68.382s; `go test ./internal/flow/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- fresh-all-ready: exit 0, 16.921s; `go test ./internal/fresh/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- stage3-fixtures: exit 0, 62.925s; `go test ./stage3/fixtures/... -count=1 -timeout 12m -parallel 4 -p 1 -json`.
- readers-maps: exit 0, 7.74s; `go test ./internal/ir -run ^Test(CallTargetReaders.*|EveryMapAProgramHoldsIsSorted)$ -count=1 -timeout 90s -parallel 4 -json`.
- parallel-guard: exit 0, 1.272s; `go test ./cmd/adamic-gate -run ^TestEveryTestIsParallelOrSaysWhy$ -count=1 -timeout 90s -parallel 4 -json`.
- vet: exit 0, 5.661s; `go vet ./internal/ir/... ./internal/lower/... ./internal/native/... ./internal/oracle/... ./internal/javascript/... ./internal/flow/... ./internal/fresh/... ./stage3/fixtures/...`.

Every changed count row (A/F/retain/release/peak/region):

```text
before: absent
after:  | internal/lower/testdata/callable_producers/p21.a | 3 | 3 | 7 | 13 | 3 | 0 |
before: absent
after:  | internal/lower/testdata/callable_producers/p60.a | 3 | 3 | 7 | 13 | 3 | 0 |
before: absent
after:  | internal/lower/testdata/callable_producers/p61.a | 4 | 4 | 7 | 14 | 4 | 0 |
before: absent
after:  | internal/lower/testdata/callable_producers/p23.a | 3 | 3 | 7 | 13 | 3 | 0 |
before: absent
after:  | internal/lower/testdata/callable_producers/p22_adapted.a | 3 | 3 | 6 | 11 | 3 | 0 |
before: absent
after:  | internal/lower/testdata/callable_producers/p59_adapted.a | 4 | 4 | 7 | 11 | 4 | 0 |
before: absent
after:  | internal/lower/testdata/callable_producers/p62_adapted.a | 4 | 4 | 7 | 11 | 4 | 0 |
before: absent
after:  | internal/lower/testdata/callable_producers/p135_tagged_unread.a | 4 | 4 | 7 | 11 | 4 | 0 |
before: absent
after:  | internal/oracle/testdata/callable_union_result_number.a | 4 | 4 | 1 | 6 | 4 | 0 |
before: absent
after:  | internal/oracle/testdata/callable_union_result_string.a | 4 | 4 | 2 | 7 | 3 | 0 |
before: absent
after:  | internal/oracle/testdata/callable_union_result_zero.a | 4 | 4 | 1 | 6 | 4 | 0 |
before: absent
after:  | internal/oracle/testdata/callable_union_result_both.a | 17 | 17 | 5 | 26 | 7 | 0 |
before: absent
after:  | internal/oracle/testdata/loom_views_c/snapshot-original.a | 3 | 3 | 8 | 15 | 3 | 0 |
before: absent
after:  | stage3/interface-downcasts/v2/snapshot-maybe-boolean.a | 6 | 6 | 9 | 24 | 2 | 0 |
before: absent
after:  | stage3/interface-downcasts/v2/snapshot-maybe-number.a | 10 | 10 | 9 | 21 | 4 | 0 |
before: | stage3/interface-downcasts/lane4b/fixtures/node-indicator-good.a | 5 | 5 | 15 | 22 | 4 | 0 |
after:  | stage3/interface-downcasts/lane4b/fixtures/node-indicator-good.a | 5 | 5 | 14 | 22 | 4 | 0 |
before: | stage3/interface-downcasts/v2/callable-producer-wrong.a | 2 | 0 | 5 | 4 | 2 | 0 |
after:  | stage3/interface-downcasts/v2/callable-producer-wrong.a | 3 | 3 | 6 | 11 | 3 | 0 |
before: | stage3/interface-downcasts/v2/undefined-read-write.a | 2 | 2 | 5 | 9 | 2 | 0 |
after:  | stage3/interface-downcasts/v2/undefined-read-write.a | 2 | 2 | 4 | 9 | 2 | 0 |
```

Mutant results:

- packed-snapshot-stale: exit 1, expected green=False, verified=True, 21.635s; TestViewSnapshotPayloadOwnership.
- native-direct-code: exit 1, expected green=False, verified=True, 12.574s; TestCallableProducerRegistryFiltersDirectFunctions.
- exact-producer-type: exit 1, expected green=False, verified=True, 18.384s; TestCheckedViewV2CallableProducerMutant.
- ambiguous-producer: exit 1, expected green=False, verified=True, 11.368s; TestCallableProducerAmbiguousRefused.
- missing-abi-refusal: exit 1, expected green=False, verified=True, 11.468s; TestCallableProducerExtraOptionalRefused, TestCallableProducerMethodShorthandRefused.
- raw-union-result: exit 1, expected green=False, verified=True, 15.956s; TestCallableUnionResultNumber.
- lower-direct-code-fixed-mutant: exit 1, expected green=False, verified=True, 11.69s; TestCallableProducerCertificatesUseClosureThunks.
- snapshot-retain-unmasked: exit 1, expected green=False, verified=True, 29.103s; TestLoomViewsCOriginal.
- tagged-lazy-check: exit 1, expected green=False, verified=True, 11.421s; TestCallableProducerTaggedUnreadMethod.
- unrecorded-unique-result: exit 1, expected green=False, verified=True, 4.331s; TestCallableProducerUniqueObjectResultRefused.
- unique-certificate-fixed-witness: exit 1, expected green=False, verified=True, 11.182s; TestCallableProducerAmbiguousCertificate.
- snapshot-retain-control: exit 0, expected green=True, verified=True, 16.093s; .

Measured leaves: 10186; slowest 24.09 seconds.
