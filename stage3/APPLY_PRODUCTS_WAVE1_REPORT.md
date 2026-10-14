Built four offline adapted-tree products on main, plus a fetch-only product 99 API for the lane.
Code commit: `9cb9ac58f31cae822409505ddb6e44b9b547e3d6`; branch: `typescript/stage3-apply-products`.
Cold products: 18.60 / 39.70 / 47.66 / 23.41 seconds; whole-tree equality and all 25 apply fixtures passed.
Four input-drop mutants, a real tree-byte mutant, a parent-only deadline mutant and a write-table mutant were caught.
Not covered here: lane products, the remaining 150 Go shards, remote publication, or the full repository gate.

## Part A

Base: `origin/main` at `d072c133ec923dca88572949a9b21ea09489a4a3`. The changes from
8dd02d807a, 2888e54540 and ad37d7381f were re-cut by hand, retaining main's
supplied depth-one TypeScript source, byte-pinned installed API packages,
os.walk digest, build-cache location and disabled bytecode writes. The old
GitHub clone and npm install are gone. The executable-bit manifest-key fix
from ad37d7381f is retained, with its permission-change fixture. Old reports
were not copied as new evidence.

Each successor copies its predecessor independently. Product 99 retains the
existing verified multipart archive and parser alias behavior. `Fetch` looks
up product 99 directly and fails on a miss; lane consumers cannot build apply.
The ordinary Python CLI retains its local miss/build behavior through `Get`.
The four keys name cumulative adaptations, the source pin, parser manifest,
lock and byte pin, Python/Go recipes, cache helper, tools and execution settings.

Machine: Linux x64, Go 1.27.1, Node 24.19.0, clang 20.1.8. `nproc` returned 5;
`/sys/fs/cgroup/cpu.max` was `400000 100000`, a four-CPU quota.

| Setup command | Exit | Observed timing lines |
|---|---:|---|
| `export GOPROXY='https://proxy.golang.org\|direct'; timeout 600 bash cloud/setup.sh` | 0 | Go 0.069 s; Node 0.072 s; clang 0.467 s; markdown install step 0.715 s, ready 0.909 s; submodules 167.505 s; shared cache 168.199 s; Go build 429.989 s; test binaries deferred 430.073 s; cache warm 430.074 s; done 430.110 s |

Every build/test shell sourced `/workspace/adamic-tools/env.sh`. Gate input
preparation, outside all products: `timeout 120 git clone --depth 1 --branch
v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/hefhych-typescript`
and `timeout 90 npm ci --prefix stage3/api --ignore-scripts --no-audit
--no-fund --cache /tmp/hefhych-npm-cache`, both exit 0. The source resolved to
`050880ce59e30b356b686bd3144efe24f875ebc8`. The upstream packages for part B
were also installed outside products with `timeout 120 npm ci --prefix
/tmp/hefhych-typescript --no-audit --no-fund --cache /tmp/hefhych-upstream-npm`,
exit 0, 338 packages in 10 s.

### Cold products

For each NN, one fresh invocation, in order, with predecessors already prepared:

```sh
source /workspace/adamic-tools/env.sh
export GOWORK=off ADAMIC_TYPESCRIPT_SOURCE=/tmp/hefhych-typescript
export ADAMIC_BUILD_CACHE_DIR=/tmp/hefhych-cache-A ADAMIC_BUILD_CACHE_URL=off
timeout 90 go test ./stage3/applyproducts -run '^TestProduct_Stage3AdaptedNN$' -count=1 -timeout 90s -v > /tmp/hefhych-ANN.log 2>&1
```

The cache directory was fresh before product 10. Compiler/OS caches were not
flushed. Setup had finished; no other build ran during these four measurements.

| Product | Work | Top-level test wall s | Builder log s | Exit |
|---|---|---:|---:|---:|
| 10 | Local depth-one clone, diagnostics, adaptations 00 and 10 | 18.60 | 18.52 | 0 |
| 40 | Copy, adaptations 20 through 40 | 39.70 | 39.52 | 0 |
| 70 | Copy, adaptations 41 through 70 | 47.66 | 47.46 | 0 |
| 99 | Copy, remaining adaptations, multipart packaging | 23.41 | 23.20 | 0 |

`unshare -n true` exited 1, `Operation not permitted`. Network namespaces are
unavailable. Successful worker output is retained in each log: the only clone
is the supplied local file:// checkout, no npm install runs, and no remote
fetch is logged. Source and packages were supplied before these invocations.
No product writes a cache into HOME. The remote build store was disabled.

### Equality and checks

`timeout 90 go run ./stage3/applyproducts/fetch` exited 0 and returned
`/tmp/hefhych-cache-A/local/0b9b93822b89c19319772f142dd5db952f7309b56ce05d049794ce8c2fc0aa25`.

`timeout 240 python3 stage3/applyproducts/prove_tree.py
/workspace/scratch/hefhych-main-apply PRODUCT99` exited 0. It executes
origin/main's original `build_product`, without the chain. Plain main apply
took **100.461376 s**. Every non-.git path is sorted and hashed with its kind,
full mode, symlink target or file bytes, including empty directories.
Both hashes were:

`b122537d07680ddad789bf9e543ac807efdeb6fe6748b49d325e83764306901b`

| Verification command | Exit | Result |
|---|---:|---|
| `timeout 90 go test ./stage3/applyproducts ./stage3/applyproducts/fetch -run 'TracksInputs\|Deadline' -count=1 -timeout 90s -v` | 0 | Four TracksInputs checks and deadline proof; 1.360 s |
| `timeout 90 go test ./stage3/applyproducts ./stage3/applyproducts/fetch -count=1 -timeout 90s -v` | 0 | Entire changed packages; 1.003 s plus fetch package with no tests |
| `timeout 90 python3 stage3/test_apply.py` | 0 | All 25 fixtures, 13.886 s; detached checkout remains clean |
| `timeout 90 go vet ./stage3/applyproducts ./stage3/applyproducts/fetch` | 0 | No diagnostics |
| `git diff --check` | 0 | No diagnostics |
| Prescribed root lane-check command, after code commit | 0 | 12.5 s; gofmt/tools on 3 Go files, parallel analyzer on 1 package; vet exceeded its 10 s cap, completed separately above |

The prescribed command was bounded by `timeout 90`:
`git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show
origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`.
This clone initially tracked main only. Its first invocation printed an invalid
remote-ref error; the pipeline's exit 0 was vacuous. Explicitly fetching the
two missing tracking refs fixed that. A pre-commit run also inspected zero Go
files and is not counted as validation of the committed diff.

### Mutants

| Mutant | Command / exit | Intended catcher |
|---|---|---|
| Product 10 drops source.json | `go test ./stage3/applyproducts -run '^TestStage3Adapted10TracksInputs$' -count=1 -timeout 90s -v`, 1 | `changed source pin reused product key` |
| Product 40 drops source.json | Same command with 40, 1 | Same assertion |
| Product 70 drops source.json | Same command with 70, 1 | Same assertion |
| Product 99 drops source.json | Same command with 99, 1 | Same assertion |
| Append bytes to reference `src/compiler/core.ts` | Equality script, 0 for successful rejection | Whole-tree hashes differ; reference restored |
| Kill worker parent only | `go test ./stage3/applyproducts -run '^TestProductDeadlineKillsDescendants$' -count=1 -timeout 90s -v`, 1 | `deadline left the worker's child running`; child killed by cleanup |
| Force `--write-table` | `STAGE3_APPLY_MUTANT=1 python3 stage3/test_apply.py ApplyCheckoutTests.test_ordinary_apply_leaves_git_clean`, 1 | `ordinary apply dirtied the source checkout`, 13.214 s |

The four input mutants run through `timeout 90 python3
stage3/applyproducts/prove_inputs.py`, exit 0 for four intended failures;
each nested Go invocation has `-timeout 90s`. The recipe is restored in a
finally block. The parent-only and write-table runs also had outer hard limits.
Existing fixture mutants additionally reject corrupted archive parts, payload
hashes, manifest keys, parser bytes/modes/links and missing packages.

### Failed and superseded attempts

The first key-test invocation exited 1 before tests because setup had not yet
populated `cohere/TypeScript/tsc/go.mod`; the GOWORK=off retry passed. Early
focused Python runs passed 22 and then 24 fixtures as the upstream tests were
restored. Early product 10/40 recipes passed at 20.90/41.01 s. A later product
10 passed at 49.11 s during setup; another passed at 57.27 s and product 40 took
71.96 s under setup contention, exceeding the unit budget. Those are failures
for sizing, not the final uncontended proof. A superseded product 70 invocation
was stopped with its worker process group (143) before using its timing.

The first equality attempt exited 1 at adaptation 47 with ENOSPC. The first
all-fixture attempt exited 1 when its worktree could not be written. `/tmp` is
an 8.8 GB tmpfs, separately from the 32 GB workspace. Only this task's obsolete
cache directories and partial reference tree were removed. The fresh reference
was moved to the workspace volume; equality and all fixtures were rerun and
passed as recorded above.

Compressed verification logs are in `applyproducts/evidence/`. No full upstream
oracle or full repository gate was run for part A. Source-only delivery; remote
build-store publishing, traced settlement and cold remote restoration are not
claimed by these local-cache measurements. Both commits use Task: #hefhych.
