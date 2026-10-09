Merged batch 1 into the bridge split for roadmap step 38, preserving test and evidence paths.
Delivery: compiler/test-split-bridge-b2; merge parents 54c980ec and f45b69dd.
Checks: individual test selections, vet, coverage, cache reuse and planted failures are recorded here.
Mutants: original sanitizer/oracle mutants plus independent shard, coverage, cache and budget probes.
Not covered: full repository gate, remote-store transfers or integration's landing of batch 1.

The merge starts from 54c980ec, not ac75f06a. The only text conflict is
bridge/tsgo/bridge_test.go: batch 1 changed the combined TestBridge while the
feature split it into selectable units. Keep every original piece and move
ordinary/sanitized archives, the compiler and oracle into shared preparation,
using internal/buildcache.GoInputs and Get with GoBuild's names, reproducible
arguments and sanitizer environment. Overlay products remain in the keyed bundle.
Budget failures still require ADAMIC_UNIT_BUDGET=1; five-minute subprocess hang
guards inside the units remain unchanged.

Neither b05fa5c9 nor a9637488 is an ancestor of either merge parent. The published
compiler/test-split-bridge stays at ac75f06a; this delivery uses a new branch.
No old stack merge was used as a parent, and no history was rewritten.

Per the final lane correction, review/test-split-bridge remains in the tree.
The generator also remains there, matching the generated table's recorded path;
no evidence branch or generator relocation is needed. The feature's net change
against batch 1 consists only of bridge *_test.go files and review evidence.
Current main is recorded in checks.json: while main lacks batch 1, its comparison
also contains inherited train-3/buildcache/split changes. Eligibility against main
depends on integration landing that batch, as the user explicitly clarified.

The deadline instruction arrived during the original measurement. That runner
was stopped and all 78 selections were rerun with go test -timeout=75s plus a
75-second process-group kill deadline. Each selection records wall duration and
cooked=false only after completing under the limit. The unit budget gate stays
at its existing 30 seconds. No successful query observation is cached; only
hash-keyed build products are reused. Compiler corpus pin:
050880ce59e30b356b686bd3144efe24f875ebc8.

All checks passed. The 36 registered top-level Test functions stay unchanged;
no entry was added to cmd/adamic-gate/main.go's children table.

| Mode | Passed | Skipped | Maximum test body | Maximum invocation | Over 60 | Killed |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Sample | 19 | 20 | 0.180s | 5.370s | 0 | 0 |
| Compiler corpus | 34 | 5 | 1.240s | 5.837s | 0 | 0 |

selections.csv records every unit and mutant selection's actual duration and
status. All measurements completed under a 75-second process-group deadline.
The later 90-second replacement does not invalidate those shorter completed
runs. The checked-in rerunnable harnesses now use 90 seconds, classify a kill
as a failing P0, and reject measurements over the unchanged 60-second budget.
The existing unit budget still fails only with ADAMIC_UNIT_BUDGET=1.

The exact coverage pin retains 36 cases/bindings, 16 sample or 31 compiler active
pieces, 162 or 1600 query positions, five analyses per position and one owner at
every shard count 1..40. All four compiler program roots stay in each query unit.
The fresh product cache records one ordinary and one sanitized archive miss,
then 78 hits each. These include 31 preliminary active selections before the
deadline instruction and the definitive 47 active original pieces in the full
hard-limited rerun. Archive keys are 0f09cee8eac4 and d9968ae9162a.

The native manifest mutation fails only TestBridgeOracleSample in shard 3/4;
the other three shards pass. Missing a piece, query or shard filter fails coverage.
Accepting corrupt bytes, accepting an extra product or rebuilding a cached product
fails the cache pin. A planted over-budget elapsed duration fails with budget=1
and logs/passes with budget=0. A poisoned bundle callback passes after preparation,
proving reuse. All 13 probes gave the expected result, with no over-budget or
killed invocation. The two archive-key mutant ABI invocations took 10.733s and
6.452s; both passed, and their two misses failed the one-miss guard.
Original ASan, stale-handle, oracle, linkage and LSan mutants remain caught by
the independently selected units in both modes.

Commands, each redirected to logs:

```
bash cloud/setup.sh
go vet ./...
ADAMIC_TSGO_PREPARE=1 go test ./bridge/tsgo -run '^$' -count=1
python3 /tmp/bridge-b2-validation/validate.py
python3 /tmp/bridge-b2-validation/mutants75.py
python3 /tmp/bridge-b2-validation/cache-mutant.py
```

Each measurement selection used go test -json -run '^<name>$' -count=1
-timeout=75s and a process-group deadline of 75 seconds. GOMAXPROCS=4,
GOFLAGS=-p=2, ADAMIC_GATE_UNCACHED=1 and ADAMIC_UNIT_BUDGET=1 were set.
ADAMIC_BUILD_CACHE_DIR names a fresh scratch product cache and
ADAMIC_BUILD_STORE=off isolates the build-count proof from remote transfers.
Shared-product preparation ran before the deadline instruction and took 81.523s;
it had no selected test body. No such preparation test was launched afterward.
Every later selected invocation is individually timed in selections.csv.

Setup passed. nproc=5, cpu.max=400000 100000 gives a four-CPU quota.

```
go version go1.27.1 linux/amd64
setup: go ready (0.020s)
v24.19.0
setup: node ready (0.024s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.013s
setup: submodules ready (0.077s)
setup: markdown dependencies ready (0.079s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.203s)
setup: go build ready (38.713s)
setup: test binaries deferred (use --warm-tests) (38.951s)
setup: build cache warm (38.953s)
setup: build-flags commit=54c980ec6b210b47fd0bae914832da10d75802bd nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=0.00 0.50 1.14 1/269 331381 load-after=2.13 1.02 1.29 2/272 332067
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (38.991s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.7PYyR9
```
