Merged main into the bridge split and switched its shared products to internal/buildcache for step 38.
Delivery: merge commit containing this report; branch compiler/test-split-bridge, parents 93cd7227 and 7a10c877.
Checks: vet, shared-product preparation, sample and compiler-corpus units and exact coverage pass.
Mutants: seven original bridge mutations remain caught; planted input fails only its owning shard, cache/coverage/budget probes pass.
Not covered: full repository gate, other platforms or cold Go compilation as a timed test unit.

## Merge meaning

Git automatically merged bridge/tsgo/bridge_test.go on the fetched tips; no text
conflict remained. Main's V1 context.Background() lowering calls stay in the
linked-checker refusal pin. The split keeps all 36 registered cases and their
independent helpers. The unchanged coverage assertion checks registration,
helper, file, round, original query positions and one owner for every active
piece at each shard count 1..40. No production compiler file was edited.

The ADAMIC_UNIT_BUDGET=1-only failure and five-minute context guard remain.
The nested linkage test now also uses a five-minute timeout instead of its old
25-second timeout, so it follows the same hang policy.

The prepared single cache adapter now calls internal/buildcache.Get with the
existing complete input digest. The obsolete local lock/publication helper was
removed. Product count and SHA-256 verification still occur after retrieval.
The shared helper provides keyed locking and atomic publication. All 19 products
are prepared once before unit timing and fetched thereafter; successful query
observations are never cached. build-products.log records misses and hits.
The corrupt-count, corrupt-digest and callback-rebuilding mutants hold the adapter.

No .a program was changed and counts.md is identical to main; no counts row is
added or changed by this test-only merge.

## Commands and observations

The commands ran sequentially with GOMAXPROCS=4, GOFLAGS=-p=2,
ADAMIC_GATE_UNCACHED=1 and ADAMIC_UNIT_BUDGET=1, with all output written to logs.
Preparation additionally uses ADAMIC_TSGO_PREPARE=1. The corpus command uses
ADAMIC_TSGO_CORPUS=/tmp/test-split-bridge-typescript, pinned to
050880ce59e30b356b686bd3144efe24f875ebc8.

```
go vet ./... # exit 0
go test ./bridge/tsgo -run '^$' -count=1 -timeout=10m # exit 0
go test ./bridge/tsgo -json -run '^TestBridge|^TestTSGoRequiresLink$' -count=1 -timeout=10m # exit 0
go test ./bridge/tsgo -json -run '^TestBridge|^TestTSGoRequiresLink$' -count=1 -timeout=10m # exit 0
python3 review/test-split-bridge/mutants.py # exit 0
```

```
sample: 19 passed, 20 skipped; maximum active test 0.180s
corpus: 34 passed, 5 skipped; maximum active test 1.260s
```

Sample mode retains 16 original active pieces, 162 query positions and five
query analyses per position. Compiler mode retains 31 active pieces, 1600
positions and five analyses per position, with all four original program roots
present in every query unit. Every active unit passed under the budget gate.
The exact test-event results are in units.json; commands and exits in checks.json.

## Planted failures and mutants

The changed native query manifest is planted through a Go overlay, changing only
sample position 0 to position 14. Shards 0/4, 1/4 and 2/4 pass; shard 3/4 fails
only TestBridgeOracleSample with a native oracle mismatch. This is the owning
unit for registration 11. The records in mutants.json include exact failed roots,
selected commands and budget mode. No compile failure counts as proof.

Missing a registration, dropping a query position, and removing the shard filter
independently fail TestBridgeUnitsCoverEveryPiece. Corrupt bytes and extra
products independently fail TestBridgeProductCacheIsVerified; invoking the
build callback after retrieval fails its built-once assertion. Poisoning the
build callback after preparation leaves the selected units passing, proving reuse.

A planted 31-second elapsed duration fails TestBridgeABI only with budget=1;
the same mutation with budget=0 passes and logs the over-budget duration.
The existing input/output length, stale handle, wrong-position type, link guard,
output free and region ownership mutations remain caught by ASan, exact assertion,
independent Go oracle, refusal pin and LSan, as applicable, in both active modes.

## Setup

Used the cloud-environment-runtime skill to inspect the enforced network policy.
The original worktree was preserved; the merge used a detached isolated worktree
and pushes its resulting commit to compiler/test-split-bridge without rewriting.
The initial Git reference attempt rejected a shallow repository, and remote
checker initialization stalled. Those setup attempts were stopped. Git object
alternates reused the existing cohere submodule objects; checkout verified
cohere 7945d102 and its TypeScript submodule d92d9bfee1. No cohere source was
copied into the project. The subsequent setup passes; its output follows.

```
go version go1.27.1 linux/amd64
v24.19.0
setup: go ready (0.022s)
setup: node ready (0.023s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.008s
setup: markdown dependencies ready (0.078s)
setup: submodules ready (0.118s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.250s)
setup: go build ready (194.689s)
setup: test binaries deferred (use --warm-tests) (194.843s)
setup: build cache warm (194.845s)
setup: build-flags commit=93cd7227cec90874c2f7a7c9d7665fe4c16722eb nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=0.34 0.12 0.03 1/234 277941 load-after=4.56 2.31 0.91 1/236 279366
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (194.881s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.UIJdtw
```

Setup reports nproc=5 and cpu.max=400000 100000, a four-CPU quota. The env file
/workspace/adamic-tools/env.sh was sourced for every build and test command.
Complete output is compressed under evidence/. Historical measurements in the
parent report remain historical; this report describes the merged tip.
