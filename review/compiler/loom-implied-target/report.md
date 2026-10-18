Built a per-lowering, node-keyed cache for array element and conditional/logical implied targets, including nil answers.
Commit: this single commit on f93a65fba5a9d7dcd3be462391c8bc8bbef80d0a; branch compiler/loom-implied-target; task #d0bzeyh.
Measured control: Lower 32.723537201s -> 5.422958125s; impliedTarget array GetTypeAtLocation calls 4096 -> 1.
Validation: six new leaves pass; four cache-bypass mutants fail at the second query; 52 reader packages, reader guard, vet, oracle refusal/neighbor checks and existing counts checks pass.
Not covered: mutable parallelMap admission. The exact named fixture is absent from main and is refused before the hotspot on current main. No new oracle fixture or runtime feature was added.

Observed measurements:

| Program | Before Lower seconds | After Lower seconds | Before array queries | After array queries | Outcome |
|---|---:|---:|---:|---:|---|
| Exact objects.a from runtime/cycles-corpus | 0.013067581 | 0.012410636 | 0 | 0 | Identical parallelMap mutable-items refusal |
| Same literal and worker, sequential items.map control | 32.723537201 | 5.422958125 | 4096 | 1 | Accepted; all backends agree with Node |

The exact source was read from `1e6140fe5dcb45f6346f88325075195174d3a23e:internal/oracle/testdata/moves/accepted/objects.a` because f93a65fb has no file at that path. The copy and provenance are under this review directory, not registered as an oracle fixture. Its 2,048 literal elements have contiguous values 0 through 2047 and done=false. SHA-256: `848de67f6a6711fe1da741bfd93b2732248f17bbc27c396a6faa7e98206dbd6d`.

Current main refuses the exact original at its parallelMap call: items are a mutable { value: number; done: boolean; }[]. The refusal is byte-identical before and after. Thus the reported >50s original-fixture hotspot is not reachable through current main admission. Conservative assumption: leave admission unchanged and measure the same literal through an accepted sequential control. The control removes the parallelMap import and replaces that call with items.map; the worker body and literal are unchanged (control-adaptation.patch.gz). Source Node prints the same result for both programs:

```text
2048 moved objects, ordered digest 661084526
```

The timing executable loads the source first, then times exactly lower.Lower with time.Now/Since, including its checker acquisition and lowering passes but excluding load.Load and Go compilation. One measured Lower invocation per program/version; no averages. Temporary print counters around impliedTarget GetTypeAtLocation calls were built into the measurement binaries, then removed from production source. Query counts are scoped to impliedTarget, not every legitimate checker call in other compiler passes. Counter patches and measure.go.txt preserve the measurement method. Timing was on Linux amd64 with the 4-CPU cgroup quota; other verification commands may share that quota. The count is the regression criterion; wall time is observational. The sequential Lower speedup measured 6.034x; this does not claim the remaining lowering work is linear.

Implementation and inference:

- `lowering.impliedTargets` is a plain map keyed by the containing *ast.Node and holds the inferred element target for arrays or the full parent type for conditional/logical branches. It never reaches the IR program.
- The checker and AST remain fixed during a lowering. Cache hits preserve the first checker answer, without changing contextual-type queries, branch eligibility, array/tuple discrimination, or inferred function-result handling. A map membership test distinguishes a cached nil from a miss.
- Conditional and ??/||/&& cases had the same repeated parent-query shape and now use the same cache. The condition of a conditional still has no branch target.
- The optional per-lowering query function is a test seam, defaulting to checker.GetTypeAtLocation. No production counter or global test hook remains.

Focused checks actually run:

- `timeout 150 go test ./internal/lower -run` the lower-only portion of reader-command.txt, `-count=1 -v -timeout 90s` (lower-focused-final.log): pass, 6.917s. It covers TestAMutableLocationSeenWiderIsRefused (including inferred conditional, array-element and return widening), TestAViewThatCantWriteIsNotRefused, TestOptionalWideningRefused, TestOptionalWideningSpreadOverwrite, TestNodeHostOptionsRefusalBoundaries, TestNodeFSFileDoesNotAuthorizeMutableWidening, TestEnumSlotViews, TestEmptyNeverMapCannotGainWritableInhabitants, both TestProvenRelations tests and all six new tests.
- `timeout 150 go test ./internal/oracle -run "^TestStage3EnumBoundaries$|^TestClosureMergeRefusals$|^TestCheckedViewUntaggedSourceDispatch$" -count=1 -v -timeout 90s` (oracle-refusals.log): pass, 22.873s.
- Grepped the oracle tests/refusal pins for the diagnostic fragments emitted by invariance.go, including invariant-mutable, enum-literal, enum-members, method-signature-style, which can write and readonly-field text. The direct existing match is TestStage3EnumBoundaries/06_set_node_flags.a (refusal-search.log). Optional-widening pinned neighbors in TestClosureMergeRefusals and the structural/optional checked-view controls were also run.
- `timeout 150 go test ./internal/oracle -run "TestNativeAgreesWithNode/internal/oracle/testdata/(invariance_readonly|optional_widening_(class|declared|fresh)|node_options_widening)" -count=1 -v -timeout 90s` (oracle-neighbors-selected.log): all five fixtures pass against Node, JavaScript, native and sanitizer/leak checks, 4.099s. The initial selector in oracle-neighbors.log matched no fixture leaves; the corrected full path selector above ran them.
- The same five paths under `TestCountsAreRecorded` passed read-only (counts-neighbors.log), 73.114s. Main prepares count measurement products ahead of its selected rows, so this also prepared broader shard products. counts.md was not rewritten: there is no new registered fixture, and the checked existing counts are unchanged.
- The exact 52-package command is in reader-command.txt; package names are in reader-packages.txt, derived from go list Imports/TestImports/XTestImports and reverse reachability from internal/lower. Exit 0. Matching focused tests ran; packages without matches were compiled. No whole-package suite or full gate ran.
- `timeout 120 go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -v -timeout 90s` (reader-guard.log): pass, 67.10s. This existing, untouched guard uses its explicitly required 90s limit.
- `timeout 120 go vet ./internal/lower` (vet.log): pass, no output.
- `timeout 90 go build -o /tmp/implied-adamic ./cmd/adamic`, then `timeout 300 python3 review/compiler/loom-implied-target/compare-control.py`: Node, generated JavaScript, native release and ASan/UBSan builds of the sequential control agree byte-for-byte, exit 0 and empty stderr (backend-results.json).

New top-level leaf timings from the final 52-package run (setup included in each leaf):

| Test | Seconds |
|---|---:|
| TestImpliedTargetTupleCachesNil | 0.19 |
| TestImpliedTargetAndQueriesOnce | 5.31 |
| TestImpliedTargetOrQueriesOnce | 3.06 |
| TestImpliedTargetLargeArrayQueriesOnce | 2.31 |
| TestImpliedTargetNullishQueriesOnce | 8.78 |
| TestImpliedTargetConditionalQueriesOnce | 22.68 |

The count-based array test lowers 3,072 object elements plus a second literal through the checked refusal walk and declaration/statement lowerer. It observes the real checker-query seam and rejects a second query before doing it, keeping the quadratic mutant below 60 seconds. The conditional test has 128 branches and each logical-chain test 256 operands. Separate tests check cached nil tuple answers and the conditional condition boundary.

Mutants (production source restored in finally after each run):

| Mutant | Catcher | Observed failure |
|---|---|---|
| array-per-element-query | `TestImpliedTargetLargeArrayQueriesOnce` (0.09s) | second parent query detected; go test exits 1 |
| conditional-per-branch-query | `TestImpliedTargetConditionalQueriesOnce` (0.04s) | second parent query detected; go test exits 1 |
| logical-per-operand-query | `TestImpliedTargetOrQueriesOnce` (0.06s), `TestImpliedTargetAndQueriesOnce` (0.06s), `TestImpliedTargetNullishQueriesOnce` (0.06s) | second parent query detected; go test exits 1 |
| nil-cache-miss | `TestImpliedTargetTupleCachesNil` (0.03s) | second parent query detected; go test exits 1 |

These restore per-element/per-branch/per-operand queries or treat cached nil as a miss. All failures are the query-count assertions, not compiler warnings, timeouts or backend crashes. Development logs also retain an initial token-text panic in the test and an overly strict checker pointer-identity comparison; the test now uses operator kinds and compares checker type text because fresh equivalent checker types need not have identical pointers.

Toolchain setup (GOPROXY=https://proxy.golang.org|direct; emitted env file sourced):

```text
submodule cache: off
go version go1.27.1 linux/amd64
setup: go ready (0.072s)
setup: submodules ready (0.076s)
v24.19.0
setup: node ready (0.089s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.015s
setup: markdown dependencies ready (0.204s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.424s)
setup: shared cache ready (1.357s)
setup: go build ready (68.347s)
setup: test binaries deferred (use --warm-tests) (68.422s)
setup: build cache warm (68.423s)
setup: build-flags commit=f93a65fba5a9d7dcd3be462391c8bc8bbef80d0a nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/138 713 load-after=6.40 1.78 0.61 1/147 1779
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (68.451s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.AmwU7n
```

The managed environment was initially pending/offline and was awaited before repository work. Two named regenerable export caches from the preceding unit were removed before reader checks to avoid disk exhaustion; no source or review evidence was removed. No cohere code was copied. No runtime C or IR files changed, no new compiler-consuming package was added, no PR was opened, and nothing was submitted to Loom.

Storage: logs are committed as .log.gz, and the go list snapshot as packages.json.gz; decompress them to read the raw logs. Reproduction commands write the raw .log files named above.
