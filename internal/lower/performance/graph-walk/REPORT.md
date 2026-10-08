# Graph allocation-site walk cost

Measured on f409709 plus the pointer/map traversal patch, with current main b6b1538, area/runtime 94a9c832 and developer-tools f6eef5df merged. The compiler source hash, benchmark hash, all input manifests/hashes, individual samples and log paths are in [measurements.json](measurements.json).

The original 78-file TypeScript compiler source corpus is absent from this checkout. These are the available whole parser, the largest stage-1 entry by transitive source bytes (lint), and the largest by transitive implementation-file count (the checker volume suite). Imports were followed including type imports; prelude/library declarations are excluded from these file/line totals.

| Input | Files | Lines | Source bytes | Functions | Locals |
|---|---:|---:|---:|---:|---:|
| stage1/typescript/parser/main.ts | 9 | 4492 | 187566 | 155 | 784 |
| stage1/cohere/lint/main.ts | 17 | 8718 | 343341 | 257 | 1358 |
| stage1/cohere/typeaware/volume_suite.ts | 31 | 8267 | 342208 | 359 | 1859 |

Go 1.27.1 linux/amd64, default `go test -c` optimization, no race instrumentation; Xeon Platinum 8573C, nproc 5, cgroup cpu.max 400000/100000, GOMAXPROCS 5. Source loading, type checking and binary compilation are outside timed intervals. Each sample reuses one loaded checker and warms it with one complete lowering before the timer. Variants run sequentially in alternating order, three rounds of three lowerings each. No gate jobs ran during sampling.

The baseline is a scratch Go source overlay that skips only the allocation-site tree copy/IDs and their selection writes. The rest of the flow pass runs. This baseline is for compiler timing; it intentionally omits classification needed to prevent the known leaks and was not used for native execution. There is no production switch to disable the walk.

All three original entries prove their cycles and have no graph types, so they never invoke the walk. The graph-enabled variant appends a small real escaping self-cycle through LoadOverlay, with two extra functions and two extra locals; it forces the existing graph pass to run over each complete entry. No stage-1 file is changed.

| Input | Graph component added | With walk, median ms | Without walk, median ms |
|---|---|---:|---:|
| stage1/typescript/parser/main.ts | no | 655.22 | 720.98 |
| stage1/typescript/parser/main.ts | yes | 948.67 | 1121.18 |
| stage1/cohere/lint/main.ts | no | 762.65 | 813.87 |
| stage1/cohere/lint/main.ts | yes | 1372.61 | 1672.30 |
| stage1/cohere/typeaware/volume_suite.ts | no | 1026.91 | 984.50 |
| stage1/cohere/typeaware/volume_suite.ts | yes | 2600.16 | 2600.67 |

These elapsed-time ranges overlap between variants for every input. Even the no-graph controls differ despite not running the walk. Negative differences are not evidence that copying speeds up lowering; this small comparison cannot resolve its net effect on full lowering. The graph-enabled runs also include the existing cycle/type proof and all other lowering work.

To isolate the requested cost, BenchmarkGraphAllocationSiteWalk copies the lowered original Main and function bodies with the same production helper, while loading/checking/lowering stay outside its timer. Five runs of at least one timed second each, medians below. These original entries skip this work in production; this is the cost when their IR trees are walked. B/op is total transient Go memory allocated per walk, not peak RSS or native object metadata.

| Input | Walk ms | Go bytes/walk | Go allocations/walk |
|---|---:|---:|---:|
| stage1/typescript/parser/main.ts | 9.29 | 1814890 | 64678 |
| stage1/cohere/lint/main.ts | 15.08 | 2750870 | 97650 |
| stage1/cohere/typeaware/volume_suite.ts | 16.54 | 3235882 | 111916 |

The pointer/map change adds no native runtime work. Current IR body trees do not carry allocations through those two container kinds; existing source fixtures retain their allocation-site order and recorded counts. The added traversal guard wraps all 15 real allocation-site-bearing IR types in pointers, map values, pointer map values and map keys, checks nil containers and input preservation, and inventories GraphTypes declarations in the IR source so a new allocation type requires coverage.

Three mutants in [graph-walk-mutants.py](../../testdata/graph-walk-mutants.py) each exit 1: skipping Ptr loses pointer sites; skipping Map loses map sites; adding a valid future IR allocation without a probe fails the source inventory. Sources are restored in finally. This unit changes the site-assignment walk, not the separate producer semantics for new IR expressions.

Reproduce from the repository root, with the setup environment sourced:

```text
ADAMIC_GRAPH_BENCH_ROUNDS=3 ADAMIC_GRAPH_BENCH_ITERATIONS=3 python3 internal/lower/testdata/graph-walk-measure.py /tmp/graph-walk-final-bench > /tmp/graph-walk-final-measure.log 2>&1
python3 internal/lower/testdata/graph-walk-mutants.py > /tmp/graph-walk-final-mutants-results.log 2>&1
```

Setup for this unit: go, clang, Node and submodule readiness each 0s; build-cache warm 205s; total 205s, nproc 5. The timing covers these stage-1 inputs only, not tsc itself or a general complexity bound. Concurrency remains deferred pending its integration signal.

Final verification on the restored tree:

```text
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/fresh -count=1 -timeout 30m > /tmp/graph-walk-final-packages.log 2>&1
# PASS: lower 33.597s, fresh 40.103s
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/ir -run '^TestGraph|^TestCallTargetReaders$' -count=1 -v -timeout 15m > /tmp/graph-walk-final-native-guard.log 2>&1
# PASS: all six graph tests, native 6.817s; call-target guard, ir 30.883s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/.*(regions|weak|fresh|nested|string_views)|TestFreshWriteProbesUseRegions|TestGraphRegions|TestGraphAllocationFlow|TestWeakRegionReview|TestNested.*|TestCountsAreRecorded' -count=1 -v -timeout 30m > /tmp/graph-walk-final-oracle.log 2>&1
# PASS 107.226s; native misses 944, Node misses 254, zero cache hits
python3 internal/lower/testdata/graph-walk-mutants.py > /tmp/graph-walk-final-mutants-results.log 2>&1
# skip-ptr 1, skip-map 1, new-allocation 1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower -run '^TestGraphAllocation' -count=1 -v > /tmp/graph-walk-final-restored.log 2>&1
gofmt -l cmd internal > /tmp/graph-walk-final-format.log 2>&1
go vet ./... > /tmp/graph-walk-final-vet.log 2>&1
```

The complete counts check passes without a table edit: all 543 numeric rows,
including the override fixture, remain unchanged. Format/vet logs are empty and
git diff --check passes. The full repository gate and an actual macOS run were
not performed. The previous repair f409709 includes the shared leak helper and
an independent run of Darwin's count predicate on Linux; no macOS leak-tool
execution is claimed here.
