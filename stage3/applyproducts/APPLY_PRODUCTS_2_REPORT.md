Built six apply products (10, 20, 40, 65, 70, 99) and a named ten-minute step deadline.
Code commit: 4654ecd9a317191e044b1536a570c637b70a6041; branch typescript/stage3-apply-products-2, based on 46ca8671.
Cold tests: 20.51 / 35.16 / 16.82 / 22.10 / 38.50 / 26.79 seconds; slowest traced test 159.24 seconds, exit 0; tree equality passes.
Caught six source-pin omissions, final metadata omission, parent-only kill, changed tree bytes and forced write-table, each by its intended check.
Not covered: Workshop timings or Loom settlement, all six traced products, lane products, full upstream oracle or full repository gate.

The report and compressed evidence commit follows the code commit. Every commit
has Task: #hefhych. No files under stage3/lane were edited; apply.py needed no
change because its existing step-worker accepts arbitrary prefix boundaries.
The requested stack base and ancestry are retained. Only this unit's changes
under stage3/applyproducts are added on top.

The worker retains process-group cancellation. The deadline remains injectable
through workerCommand's context, so TestProductDeadlineKillsDescendants still
uses 500 milliseconds. Production uses stepDeadline = 10 * time.Minute. The
comment records that tracing exceeds the old 80 seconds and Loom separately
bounds hangs with go test's 14-minute timeout and its tracer's 15-minute bound.
Final-product lookups and metadata selection now use len(boundaries)-1 rather
than index 3. Each new boundary has its own product and TracksInputs function.
READS.md records their cumulative subsets; the package read union is unchanged.

Machine and preparation

Linux x64; nproc=5, cpu.max=400000 100000 (four-CPU quota). Go 1.27.1,
Node 24.19.0, npm 11.9.0, clang 20.1.8. All Go/test shells source
/workspace/adamic-tools/env.sh and set GOWORK=off.

`export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh`
exited 0. Timing lines: Node ready 0.088s; Go ready 0.150s; clang ready 0.608s;
markdown install step 1.024s, ready 1.289s; submodules ready 282.289s; shared
cache ready 283.304s; Go build ready 544.370s; test binaries deferred 544.457s;
build cache warm 544.458s; done 544.487s. The submodule download and initial Go
warming account for the long setup; it completed within the 600-second limit.

Gate inputs were prepared outside products:
`timeout 120 git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/hefhych-2-typescript`
and `timeout 90 npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund`,
both exit 0. Source commit: 050880ce59e30b356b686bd3144efe24f875ebc8.

Measurements and cut choice

Before choosing the cuts, `timeout 240 python3 -B /tmp/hefhych-2-profile.py`
ran current main's unmodified plain build_product with an external timing
wrapper around run(). It exited 0 and retained the reference tree. This run
overlapped setup's submodule download, so it guides the cuts rather than serving
as the final uncontended product proof. Its wrapper and full log are archived.
Adaptation 20 alone took 30.136s and 70 alone took 31.330s. Adaptations 30 through
40 totaled 6.812s, and 41 through 65 totaled 9.750s. These observations led to
the six boundaries, with 20 and 70 isolated. No adaptation was split internally.

For final measurements, setup and other tests had finished. The product cache
/tmp/hefhych-2-cache was fresh before 10. Source/parser preparation and Go
compilation caches were supplied; OS page caches were not flushed. The exact
fresh invocation for each NN was:

```sh
source /workspace/adamic-tools/env.sh
export GOWORK=off ADAMIC_TYPESCRIPT_SOURCE=/workspace/hefhych-2-typescript
export ADAMIC_BUILD_CACHE_DIR=/tmp/hefhych-2-cache ADAMIC_BUILD_CACHE_URL=off
go test ./stage3/applyproducts -run '^TestProduct_Stage3AdaptedNN$' -count=1 -timeout 90s -v > /tmp/hefhych-2-cold-NN.log 2>&1
```

The sequential scratch runner bounds each invocation at 95 seconds, including
compiler startup, and has an outer timeout 540. Each product misses once;
all its predecessors hit. Worker logs contain no source/package downloads.

| Boundary | Work | Top-level test s | Whole Go invocation s | Exit |
|---|---|---:|---:|---:|
| 10 | Local clone, diagnostics, 00 and 10 | 20.51 | 21.892 | 0 |
| 20 | Copy, adaptation 20 alone | 35.16 | 35.972 | 0 |
| 40 | Copy, adaptations 30 through 40 | 16.82 | 17.626 | 0 |
| 65 | Copy, adaptations 41 through 65 | 22.10 | 22.986 | 0 |
| 70 | Copy, adaptation 70 alone | 38.50 | 40.190 | 0 |
| 99 | Copy, remaining adaptations and packaging | 26.79 | 27.773 | 0 |

All six are below 60 seconds locally. Products **20 and 70 cannot meet the
roughly 34-second local sizing target with this recipe even when isolated**.
There is no further legal boundary cut for them. Workshop was not accessible;
these measurements do not prove its new wall times or the 45-second aim there.
The smaller blocks remove the work previously sharing those products and leave
more margin than the old 40/70 blocks. No claim of new Workshop timings is made.

Ptrace

Product 70 was the slowest. Its cached result was moved outside the cache's
local directory on the same filesystem, retaining all predecessors. After
`timeout 90 go test -c -o /workspace/hefhych-2-apply.test ./stage3/applyproducts`
(exit 0), this ran from stage3/applyproducts with the same environment:

```sh
strace -f -o /dev/null /workspace/hefhych-2-apply.test \
  -test.run '^TestProduct_Stage3Adapted70$' -test.count=1 -test.timeout=14m -test.v
```

An external Python monotonic timer bounded the subprocess at 835 seconds;
the outer command had timeout 840. This uses the explicitly requested traced
budget, not the 90-second untraced sizing budget. There was no competing build.
Unfiltered ptrace succeeded: builder 156.42s; top-level test **159.24s**;
whole traced invocation **159.297s**, exit 0. The healthy traced step exceeds
80 seconds and survives the new deadline. Only this slowest product was traced.

Whole-tree equality

The plain reference executed `origin/main:stage3/apply.py` at
97f003d15f6728fa9bf5c9615ca4970ad7da8fa6, using the same pinned adapters and parser
(the source/adapt/API paths have no diff against main). The comparison imports
prove_tree.tree_hash and hashes every sorted non-.git path: kind, mode,
symlink target or file bytes, including empty directories. Both hashes are:

`b122537d07680ddad789bf9e543ac807efdeb6fe6748b49d325e83764306901b`

The comparison, bounded by timeout 90, exited 0. It then appended bytes to the
real reference src/compiler/core.ts, verified the hashes differ and restored
the original bytes in a finally block. Product 99 is at
/tmp/hefhych-2-cache/local/79d532961a354db4451d8ecf2e83da063c489c1dce92166b3ffef0c3709bdf3a.
`timeout 90 go run ./stage3/applyproducts/fetch` returns that final directory;
it does not stop at the former index 3.

Checks and mutants

Every Go test command below includes -count=1 -timeout 90s. Outputs are complete
log files, never piped from a running test.

| Command | Exit | Result |
|---|---:|---|
| `timeout 90 go test ./stage3/applyproducts ./stage3/applyproducts/fetch -run 'Tracks\|Deadline' -count=1 -timeout 90s -v` | 0 | Six pin checks, metadata and descendant deadline; 0.982s |
| `timeout 90 go test ./stage3/applyproducts ./stage3/applyproducts/fetch -count=1 -timeout 90s -v` | 0 | Entire packages, prepared products; 1.256s |
| `timeout 90 python3 -B stage3/test_apply.py` with supplied source/cache | 0 | All 25 fixtures, 15.378s |
| `timeout 90 go vet ./stage3/applyproducts ./stage3/applyproducts/fetch` | 0 | No diagnostics |
| `git diff --check` | 0 | No diagnostics |
| Prescribed committed root lane checks | 0 | 2.9s; 3 Go files, 1 parallel test package, 1 vetted package |

`timeout 150 python3 -B stage3/applyproducts/prove_inputs.py` exited 0 for six
independent intended failures. Each drops source.json from the real recipe for
only that boundary, runs its own TestStage3AdaptedNNTracksInputs, and restores
the recipe in finally. The deadline and metadata scratch runner is archived.

| Mutant | Command / status | Intended catcher |
|---|---|---|
| 10 drops source.json | `go test ./stage3/applyproducts -run '^TestStage3Adapted10TracksInputs$' -count=1 -timeout 90s -v`, 1 | changed source pin reused product key |
| 20 drops source.json | Same with 20, 1 | TestStage3Adapted20TracksInputs, same assertion |
| 40 drops source.json | Same with 40, 1 | TestStage3Adapted40TracksInputs, same assertion |
| 65 drops source.json | Same with 65, 1 | TestStage3Adapted65TracksInputs, same assertion |
| 70 drops source.json | Same with 70, 1 | TestStage3Adapted70TracksInputs, same assertion |
| 99 drops source.json | Same with 99, 1 | TestStage3Adapted99TracksInputs, same assertion |
| Drop final top-level metadata selection | `go test ./stage3/applyproducts -run '^TestStage3Adapted99TracksMetadata$' -count=1 -timeout 90s -v`, 1 | changed source pin reused product key |
| Kill worker parent only | `go test ./stage3/applyproducts -run '^TestProductDeadlineKillsDescendants$' -count=1 -timeout 90s -v`, 1 | deadline left the worker's child running; cleanup kills child |
| Append bytes to real reference core.ts | Bounded hash comparison, 0 for successful rejection | whole-tree hashes differ |
| Force --write-table | `timeout 90 env STAGE3_APPLY_MUTANT=1 python3 -B stage3/test_apply.py ApplyCheckoutTests.test_ordinary_apply_leaves_git_clean`, 1 | ordinary apply dirtied the source checkout; 14.358s |

All source mutants were restored. Existing archive/parser corruption fixtures
also passed as part of the Python suite; no additional adapter mutant was added.

Current-main proof

A separate detached worktree at main 97f003d1 merged code commit 4654ecd9 without
committing, cleanly. Merge tree: 5f4eb17eb1a6129a29fff6da2076e70ff971623b.
With supplied parser/source and prepared products, both entire Go packages passed
(1.006s), and `timeout 90 python3 -B stage3/test_apply.py` passed all 25 fixtures
(14.641s). The branch retains the specifically requested base rather than
rewriting the parent's history. The following report/evidence commit changes no
recipe or runtime behavior. Final committed-tip checks run before the only push.

Root lane command, after sourcing the toolchain, bounded by timeout 90:

```sh
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

The tracking refs were explicitly fetched first because this clone initially
tracked only main. The lane invocation used pipefail so a missing script could
not count as a pass. No PR is opened and main is never pushed.

Failed preparation and limits

The first base checkout attempt could not resolve 46ca8671 in the shallow clone;
explicitly fetching the named branches resolved it. The first Python fixture
run omitted ADAMIC_TYPESCRIPT_SOURCE and failed its checkout fixture, 24 other
fixtures passing. The corrected full suite above is the validation. Initial
ptrace preparation tried a cross-filesystem rename and /usr/bin/time, which is
absent; no product ran under that attempt. A same-filesystem move and Python
timer corrected it, then the single actual traced product passed.

No new .a source fixtures or counts rows were needed. No lane ownership changes,
remote publication, Loom keyed settlement, complete upstream oracle or complete
repository gate is claimed. The roughly 34-second target remains unmet for
isolated adaptations 20 and 70; they must not be split internally.

Full logs and scratch measurement/mutant runners are compressed under
[evidence/apply-products-2](evidence/apply-products-2/).
