Whole-file sharding cannot meet 30 seconds: symlinkWatching.ts takes 43.102 seconds cold.<br>
Built an input-hashed shared artifact, per-runner/file measurements and a complete observed test-to-shard census.<br>
The whole upstream run and unchanged lane verdict retain 106366 passing, one sanctioned failure and zero pending.<br>
All 54 lane fixtures pass; nine new guard mutants and both real source mutants are caught.<br>
No valid sharded gate was installed; finer partitioning inside the watcher file is required.

The branch starts at origin/main `54cbc125`. The pushed error-reporting branch
`ab59a0a1` was merged as `12963b97`. Its failure text, stacks, timing, cause
classification and baseline previews remain intact. The lane entry point,
expected counts, API sanction and acceptance predicates were not changed.

The hard constraint is impossible with a partition **by whole upstream test
file**, even with arbitrarily many shards. Upstream puts all 16 tests in
`src/testRunner/unittests/sys/symlinkWatching.ts` in one root suite and runs them
sequentially. Its directory-watcher operations deliberately wait 500 ms per
observation when CI is unset, or 1000 ms with CI set. It is predominantly waiting,
so extra CPUs do not remove that wall. The full measurement takes 42.435 seconds
for this task. Running it alone with four fresh workers takes 43.102410 seconds
including worker startup: 16 passing, zero failures. Only one worker can execute
this one file task. Shortening waits, skipping these tests, or changing expected
counts would loosen the checks and was not done.

Cold here means fresh upstream worker processes, `NODE_DISABLE_COMPILE_CACHE=1`,
no reused test results, and inputs from the prepared artifact hash. The OS page
cache was not flushed. Both observations exceed 30 seconds independently of
artifact materialization. The required exception to whole-file partitioning
would split this file's tests while preserving its original hooks and watcher
cleanup. That change has not been authorized by the brief's file partition and
is not claimed here.

The gate invocation of `stage3/lane/run.sh` is absent from this checkout. Searches
included hidden files and `cmd/adamic-gate`; the latter partitions Go packages
but does not invoke this lane. No unrelated gate entry was edited. A valid
replacement tier therefore is not provided, and neither the 30-second test bar
nor the full-gate wall target is claimed achieved.

| Piece on this instance | Observed wall seconds | Result |
|---|---:|---|
| Fresh artifact preparation | 135.247 | apply, install, compiler build and harness build all exit 0 |
| Apply within that preparation | 101.237 | adapted tree generated once |
| npm ci within preparation | 2.308 | dependencies verified by lockfile |
| Compiler build | 22.819 | exit 0 |
| Harness build | 1.069 | exit 0 |
| Verified artifact reuse | 6.522 | hash and every payload file rechecked |
| Verified reuse after private mutant views | 6.608 | artifact bytes unchanged |
| Whole upstream suite | 357.199 | 106366 passing, 1 failing, 0 pending |
| Whole oracle including install/build | 387.951 | exit 1, exactly today's sanctioned API failure |
| Existing complete lane checker, including API normalization | 0.709 | PASS |
| Independent repeat of baseline byte comparisons | 0.046 | only api/typescript.d.ts differs |
| Cold singleton watcher file | 43.102410 | 16 passing; invalid under 30 seconds |
| Candidate shard 121 control | 6.644 | 899 passing, 0 failing |
| Candidate shard 121 single-test input mutant | 7.829 | 898 passing, 1 failing |
| Neighbor shard 122 | 5.997 | 790 passing, 0 failing |
| Lane's own fixtures | 10.007 | 54 tests, OK |

This is Linux x64 / Node 24.19.0, nproc 5, with cgroup CPU quota
`400000 100000`, hence four CPUs. The whole suite keeps the existing eight
upstream workers; the cold singleton starts four. The fresh artifact preparation
overlaps part of the whole measurement, so the file timing weights include
contention. The cold singleton starts after preparation and the whole run have
finished. Verified artifact reuse briefly overlaps its timed waits. These are
observations on this box, not a promise of upper bounds on other loaded boxes.

`cloud/setup.sh` passed. Timing lines: Go ready 0.064 s, Node ready 0.099 s,
clang ready 0.509 s, Markdown dependencies ready 2.234 s (install step 1.989 s),
submodules ready 23.992 s, Go build ready 251.805 s, test binaries deferred
251.906 s, build cache warm 251.907 s, done 251.941 s. The environment file is
`/workspace/adamic-tools/env.sh`. The required GOPROXY fallback was set first.
[Setup log](evidence/shards/setup.log.gz).

The new preparation helper hashes the full non-lane/non-oracle stage3 input
files, including adapter proof data, its own code, tool versions and platform.
It rejects an arbitrary TSC_ADAPT_TYPESCRIPT override, locks by key, builds in
a private directory, checks input identity again before publishing, and verifies
the full payload on reuse. Failed preparation logs remain available; incomplete
builds never become ready artifacts. An early preparation was correctly rejected
when the helper's input identity changed during implementation. The final
artifact key is
`e41a430a94a8b7b011d31cd1c9aaabfda03d4f8c9dc3f67d13eb8859aa434b24`.
Its [manifest](evidence/shards/artifact-ready.json) records payload SHA-256 and
individual phases. Preparation is an artifact producer, not a test unit.
The separate view helper shares immutable source/build/reference inputs and
provides private local baselines and watcher scratch. Its private-case overlay
lets a mutant change real input without changing the shared artifact.

Upstream's original IPC worker protocol provides 19676 task results. A fork
observer retains task durations, passing names and original Mocha errors without
replacing upstream scheduling. Source-mapped root describe and it registration
associates all 327 unit tasks with their test files; discovery exits before any
worker starts. The full data and [per-runner totals](evidence/shards/runner-measurements.json)
are saved alongside the ordinary oracle reporter and diagnostics. Summed task
times are CPU-concurrent durations, not runner wall times.

The deterministic longest-first candidate plan groups every unit suite belonging
to one source file, then balances measured file times with lexical ties. A
10-second target gives 135 candidate shards. Its largest file alone exceeds the
hard 30-second limit, and the planner exits 1 with status `invalid`. Changing N
cannot shorten that file. This plan is diagnostic evidence, not a deployable
configuration.

The [listing](evidence/shards/tests-and-shards.jsonl.gz) contains every passing
and failing upstream test, its file, task, candidate shard and unique identity.
Repeated Mocha titles are disambiguated by occurrence within the original task.
Every observed test occurs once and every file has exactly one candidate shard.
The census sums to the unsharded reporter's 106366 passing and one failing.
Both task duplication and test-identity duplication are rejected. The real
inventory fixture checks unique IDs, file ownership and the whole-run totals.
This proves the **listed** partition's union; it does not claim all candidate
shards were executed.

| Mutant actually run | Catch |
|---|---|
| Ignore changed adapter input bytes in the key | changed-proof fixture observes unchanged key and fails |
| Ignore tool versions in the key | isolated tool-version fixture fails |
| Ignore corrupted cached payload bytes | corruption fixture's required rejection is missing |
| Ignore inputs changing during preparation | concurrent-input-change fixture fails |
| Ignore a nonzero preparation command exit | failed-preparation fixture fails |
| Ignore the 30-second file limit | oversized-single-file fixture fails |
| Split a unit source file across candidate shards | source-file ownership fixture fails |
| Ignore duplicate upstream tasks | empty-result duplicate-task fixture fails; test-ID guard cannot mask it |
| Drop original IPC observation | original-payload assertion fails in Node |
| Append const stage3ForcedMismatch = 123 to scratch Protected1.ts | three upstream baseline tests fail in candidate shard 107; 816/0 becomes 813/3; four diffs retained |
| Change marked protected prot1 to public prot1 in scratch getOccurrencesProtected1.ts | exactly one upstream fourslash test fails in candidate shard 121; 899/0 becomes 898/1; neighbor 122 stays 790/0 |

The one-test mutant produces `getOccurrencesProtected1.baseline.jsonc` with the
full upstream error and stack retained. The file belongs only to candidate shard
121 and its private input overlay is invisible to other views. Replacing that
one task in the whole-run evidence yields 106365 passing and two failing. The
**unchanged** checker rejects the exact API diff, passing/failing counts, reporter
counts, failure titles and baseline path set. The saved
[proof](evidence/shards/proof.json) and
[rejected report](evidence/shards/merged-mutant-report.json) explicitly call this
an arithmetic replay. Only the affected shard and its neighbor were executed;
there is no new complete merge command or full executed-shard union proof.

Every measurement/test command wrote stdout and stderr to a log. Commands run:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step34-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /workspace/scratch/step34-tree > /tmp/step34-apply.log 2>&1
STAGE3_TASK_PROFILE=/workspace/scratch/step34-tasks.jsonl \
 NODE_OPTIONS='--enable-source-maps --require=/workspace/adamic/stage3/oracle/profile-tasks.cjs' \
 bash stage3/oracle/run.sh /workspace/scratch/step34-tree /workspace/scratch/step34-whole \
 > /tmp/step34-whole.log 2>&1
python3 stage3/lane/artifact.py /workspace/scratch/step34-artifacts
# First attempt rejected changing inputs; final preparation and two cache-hit probes logged separately.
# Discovery-only native host runs populate root-suite/root-test source mapping.
python3 stage3/lane/measure_plan.py /workspace/scratch/step34-tasks.jsonl \
 /workspace/scratch/step34-whole /workspace/scratch/step34-plan
NODE_DISABLE_COMPILE_CACHE=1 node stage3/oracle/run-tasks.cjs \
 /workspace/scratch/step34-cold-view /workspace/scratch/step34-watch-tasks.json \
 /workspace/scratch/step34-cold-watch 4 > /tmp/step34-cold-watch.log 2>&1
# The same run-tasks command with one worker measured control/mutant/neighbor views,
# first for conformance Protected1, then for the single-test fourslash mutation.
# Python replayed check_results and baseline bytes, wrote mutant diffs, and asserted rejection.
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s stage3/lane -p 'test_*.py' -v \
 > /tmp/step34-verified-tests.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/lane/prove_artifact.py \
 > /tmp/step34-verified-artifact-mutants.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/lane/prove_plan.py \
 > /tmp/step34-verified-plan-mutants.log 2>&1
node --check stage3/oracle/profile-tasks.cjs > /tmp/step34-syntax.log 2>&1
node --check stage3/oracle/run-tasks.cjs >> /tmp/step34-syntax.log 2>&1
bash -n stage3/lane/run.sh stage3/oracle/run.sh >> /tmp/step34-syntax.log 2>&1
git diff --check
```

Earlier fixture runs passed 40 tests after merging diagnostics, 5 initial artifact
fixtures, 52 fixtures before adding the observer fixtures, and the two observer
fixtures. Artifact and planner mutant runners each exit 0 only when every mutant
is killed by its intended assertion. The final run is 54 tests, OK, and both
runners exit 0. The existing dropped-error-text mutant is included in these lane
fixtures. No new Adamic fixture was added; lane counts.md was refreshed.

Not covered: a valid 30-second whole-file sharding configuration (impossible for
the measured watcher file), execution of every candidate shard, a real sharded
merge/tier, or the full Adamic gate. JavaScript-equivalence checks still run inside
the unchanged adapters during apply; they were not extracted or independently
timed, since that would require adapter territory outside this unit. API guarding
and the post-run byte scan were independently timed. Only Linux x64 was measured;
no macOS or separate artificially saturated cold-run promise is made.
