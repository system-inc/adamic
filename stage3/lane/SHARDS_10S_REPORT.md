Replanned 151 shards into 377, using a 9-second target and retaining the existing limits.\
All 377 passed serially on the normal Go product path: maximum 11.785 seconds.\
Executed union and clean full lane preserve 106,366 passes plus the sanctioned API failure.\
Four-active-shard load comparison: new maximum 36.141 seconds; old maximum 50.383 seconds, four timeouts.\
Acceptance remains incomplete: no eight-CPU allocation, legacy shared-artifact timings exceed 12 seconds, and the repository Go gate did not pass.

# Shard budget report: task #x2jdg33

Branch `typescript/stage3-lane-shards-10s`, cut from origin/main
`23ff023b9ed986834d20404d43e2c363667edf9b`. All committed changes are in
`stage3/lane/`. No split assignments, per-test deadlines, shard.py limits,
compiler files, expected lane counts or sanctioned API declarations changed.

## Plan and expected split counts

| Kind | Before | After |
|---|---:|---:|
| Whole-file bins | 134 | 348 |
| Exact groups from four expensive suites | 0 | 12 |
| Existing watcher cases | 16 | 16 |
| Public APIs byte comparison | 1 | 1 |
| Total shards | 151 | 377 |

The generator uses deterministic longest-processing-time assignment for every
ordinary file. The retained historical cold ledger's maximum is 24.649585 seconds;
`10 * 9 / 24.649585` gives a weight target of 3.651177 seconds. This is a
conservative planning calibration, not an assertion that the historical ledger
was serial. Its SHA-256 is recorded in the plan's `planning` object.

Whole-file probes still exceeded the desired bound for project-reference source
maps and convert-to-async-function. Those suites and the two noEmit suites now
have three exact groups apiece. Duplicate full titles stay together. The noEmit
compilation and clean-build comparison retain their shared describe family;
separating those tests actually failed the upstream baseline oracle. Every grep
is anchored and checked against the complete original execution inventory.

There are 393 declared Go tests in this package: one product producer and 392
other tests, including 377 shards and 15 existing checks. At 18 splits, expect
20-21 shard tests, or 21-22 runnable non-producer tests, per split before Loom's
placement. I recommend considering 49 splits if the desired placement remains
about eight runnable tests per split. This is a placement recommendation; no split
was edited, and it does not increase available CPU capacity.

## Alone-time distributions

All numbers are seconds; p95 uses the nearest rank. Fresh processes, disabled
Node compile cache, private per-unit cache directories and no result reuse were
used. The operating-system page cache was retained. Measurements include process,
view, harness and verdict time, not just Mocha test-body time.

| Measurement | n | Min | Median | p95 | Max |
|---|---:|---:|---:|---:|---:|
| Historical 151, shared artifact, **four concurrent** | 151 | 9.776 | 19.584 | 22.718 | 24.650 |
| Old problem shards, serial normal product | 8 | 10.487 | 12.181 | 17.162 | 17.162 |
| New complete plan, serial normal product | 377 | 2.054 | 5.804 | 7.928 | 11.785 |
| Old problem shards, serial legacy artifact | 8 | 21.929 | 23.998 | 29.768 | 29.768 |
| New preliminary plan sample, serial legacy artifact | 6 | 17.221 | 17.958 | 23.947 | 23.947 |

The old normal-product sample is exactly the eight shards named in the brief,
not a complete serial run of all 151. The historical ledger has `concurrency: 4`;
it cannot honestly be relabeled an alone-time distribution.

The complete new ledger has exactly 377 unique units, matching current input
hashes, all successful and all below 12 seconds. Slowest:
`stage3-lane-file-183`, 315 passing executions, 11.785379 seconds.
The coordinator took 2,216.315 seconds. The other 376 units were also executed,
not estimated from weights or copied from earlier results.

**Method limitation:** these 377 measurements use `STAGE3_ADAPTED_TREE`, the
normal Go product path. The older `STAGE3_ARTIFACT` path performs expensive
verification before running upstream tests on this box. Directory birth-time
observations put that preliminary work around 13-17 seconds in the retained
samples. The legacy samples did not meet 12 seconds (two old samples hit their
30-second standalone limit). I did not silently omit that startup from their
walls, change any limit, or claim a complete legacy-path under-12 proof.

The artifact and normal product had 82,824 identical source/test/reference/build
files, with zero missing, extra or differing files after excluding .git,
node_modules and tsbuildinfo. Their full dependency digests also matched:
`8c377575441538d71d25802d215a056088d3c4f0117dd4fd151c635603367fc8`.
This supports the same-input comparison; it does not erase the timing-path
limitation.

## Loaded before and after

`nproc` reports 5; affinity is CPUs 0-4, but cgroup `cpu.max` is
`400000 100000`: a **four-CPU quota**. No eight-CPU machine was available.
I ran an eight-active-shard pilot on that allocation and retained its failed
calibrations. It is not the requested eight-CPU proof.

The completed analogue uses four active shards and five Python busy loops with
that same inherited CPU affinity and quota. The after cohort was fixed as
File020 through File040 before timing (21 shards, about one current 18-way split).
The before cohort is the eight old problem shards. Both used the same five busy
loops and existing 50-second shard limit. Median per-shard slowdown in the new
cohort was **2.210560 times the unloaded four-concurrent cohort**, or
**3.556064 times serial alone**. This is not a claimed 2.5-times-alone calibration.

| Cohort | n | Min | Median | p95 | Max |
|---|---:|---:|---:|---:|---:|
| New, four active, no added load | 21 | 5.810 | 8.803 | 12.732 | 13.202 |
| New, four active, five busy loops | 21 | 9.970 | 18.255 | 26.305 | 36.141 |
| Old, four active, no added load | 8 | 16.000 | 20.549 | 30.897 | 30.897 |
| Old, four active, five busy loops | 8 | 36.818 | 49.944 | 50.383 | 50.383 |

All 21 new shards passed below 40 seconds. Four of eight old shards timed out;
the old maximum was 50.382969 seconds. Individual observations follow.

### New cohort

| Shard | Serial alone | Four active, unloaded | Four active, loaded | Loaded verdict |
|---|---:|---:|---:|---|
| stage3-lane-file-020 | 5.619 | 12.732 | 36.141 | pass |
| stage3-lane-file-021 | 7.803 | 13.202 | 20.827 | pass |
| stage3-lane-file-022 | 4.364 | 8.803 | 20.822 | pass |
| stage3-lane-file-023 | 3.066 | 7.948 | 14.475 | pass |
| stage3-lane-file-024 | 6.911 | 9.677 | 20.793 | pass |
| stage3-lane-file-025 | 5.144 | 6.129 | 9.970 | pass |
| stage3-lane-file-026 | 3.399 | 6.468 | 15.810 | pass |
| stage3-lane-file-027 | 4.818 | 8.224 | 12.344 | pass |
| stage3-lane-file-028 | 2.054 | 5.810 | 13.798 | pass |
| stage3-lane-file-029 | 4.681 | 7.759 | 26.305 | pass |
| stage3-lane-file-030 | 3.983 | 6.394 | 13.290 | pass |
| stage3-lane-file-031 | 4.840 | 8.805 | 16.803 | pass |
| stage3-lane-file-032 | 3.459 | 8.170 | 18.651 | pass |
| stage3-lane-file-033 | 4.926 | 9.099 | 17.516 | pass |
| stage3-lane-file-034 | 4.494 | 6.795 | 18.255 | pass |
| stage3-lane-file-035 | 4.956 | 9.141 | 20.939 | pass |
| stage3-lane-file-036 | 5.924 | 12.085 | 19.907 | pass |
| stage3-lane-file-037 | 5.820 | 10.592 | 23.414 | pass |
| stage3-lane-file-038 | 4.620 | 10.151 | 14.698 | pass |
| stage3-lane-file-039 | 5.756 | 7.827 | 18.048 | pass |
| stage3-lane-file-040 | 7.229 | 9.228 | 22.128 | pass |
### Old cohort

| Shard | Serial alone | Four active, unloaded | Four active, loaded | Loaded verdict |
|---|---:|---:|---:|---|
| stage3-lane-file-007 | 15.548 | 30.854 | 36.818 | pass |
| stage3-lane-file-010 | 17.162 | 30.897 | 50.383 | fail |
| stage3-lane-file-018 | 12.423 | 16.000 | 49.535 | pass |
| stage3-lane-file-036 | 11.839 | 23.775 | 41.058 | pass |
| stage3-lane-file-049 | 11.057 | 21.824 | 50.152 | fail |
| stage3-lane-file-057 | 10.487 | 17.218 | 50.163 | fail |
| stage3-lane-file-063 | 12.632 | 17.528 | 49.736 | pass |
| stage3-lane-file-109 | 11.939 | 19.273 | 50.244 | fail |

### Rejected calibrations and interpretation

The initial eight-active-shard pilot on the four-CPU quota tried 12, 10, 8 and
6 busy loops. Maxima were 51.018, 51.608, 50.680 and 50.665 seconds respectively;
none satisfied the after bound. Median slowdowns against the unloaded concurrent
cohort were 1.954, 1.773, 1.808 and 1.885. The pilot exited 1 without a success
proof. The four-active run's first 12-loop attempt was also rejected (4.438-times
concurrent slowdown; maximum 50.217 seconds); five loops gave the accepted
analogue above. All attempts and failures are retained.

Observed: all new shards alone are below 12 seconds on the normal product path,
and the completed four-active comparison distinguishes the old timeout behavior
from the new cohort. Inferred: 11.785379 * 2.7 = 31.820524 seconds, below 40.
That arithmetic is not a substitute for the requested eight-CPU run. The exact
eight-CPU, about-2.5-times-alone acceptance test remains outstanding.

## Executed union, full lane and API

`merge_shards.py` compared all actual shard observations with the clean same-artifact
unfiltered lane, including execution names and their multiplicities. It accepted
exactly **106,367 executions: 106,366 pass, one fail, zero skip**, with 377 units.
The only failure is
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
The only baseline difference is `api/typescript.d.ts`.

The clean full-lane command passed its own checker in 517.962 seconds, with
106,366 passing, one sanctioned failure and zero pending. It ran all runners,
eight upstream workers and no test filter. The API diff is byte-identical to
main's retained API diff, SHA-256:
`3da14ced9f2c7eea0ae8b4dabbe6067a91510b0ab0d950ddb83903a9eb50fa28`.

The first full run overlapped Go setup and failed an additional
`relationComplexityError` before-all hook at its unchanged 40-second timeout:
106,360 pass, two fail, 896.641 seconds. It is retained as a failed experiment;
the clean rerun, not that run, is the union reference.

## Commands and results

Commands below ran from `/workspace/adamic`, with output redirected to files.
Scratch result directories must be absent before reproducing coordinators.
The per-child argv and hard limits, including every load and mutant invocation,
are in [commands.jsonl.gz](evidence/budget-10s/commands.jsonl.gz).

The toolchain environment was:

```sh
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/x2jdg33-typescript
export STAGE3_ARTIFACT=/tmp/x2jdg33-artifacts/ca99f25d13bfb84fe9c89a4ba0eaf088b5a529c393a80eba07bdc326cbe48557
export STAGE3_ADAPTED_TREE=/workspace/x2jdg33/products/local/04d979148f440864934c43358444f2a06603bb7149ddfd1153bf9360df7c4c56/tree
export PYTHONDONTWRITEBYTECODE=1
```

Setup: `timeout 840 bash cloud/setup.sh > /tmp/x2jdg33-setup.log 2>&1`
exited 124 after the hard limit. Its cumulative timing lines were node .057, go .067, clang .514,
markdown 1.324, submodules 578.542 and shared cache 614.565 seconds.
The initial markdown npm-ci step took 1.130 seconds. Retrying the same bounded setup completed:
node .261, go .284, submodules .511, markdown .572, clang 1.094, shared cache
2.862, go build 490.113, tests deferred 490.245, warm 490.248, done **490.345**
seconds (cumulative lines). Markdown's step was .095 seconds and skipped as
already installed. Tool versions: Node v24.19.0, Go go1.27.1, clang 20.1.8;
`nproc` = 5. Logs preserve both attempts.

Artifact preparation needed a corrected `ADAMIC_TYPESCRIPT_SOURCE` environment;
the first attempt failed apply, and a second correctly rejected concurrent
input changes during dependency installation. The third passed in 98.645 seconds.
Producer apply/install/build/harness phases were 17.517/3.981/44.821/3.354 seconds.
TypeScript pin: 6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`.
The API input digest is
`1512f1871eebe448dac94546c5a6337daa1cb64ff20c80f3d9e93f85a0f48dce`.

| Proof/check | Command (stdout and stderr saved) | Result |
|---|---|---|
| Normal product | `ADAMIC_BUILD_CACHE_DIR=/workspace/x2jdg33/products timeout 840 go test -v ./stage3/lane -run '^TestProduct_Stage3AdaptedTree$' -count=1 -timeout 13m > /tmp/x2jdg33-product.log 2>&1` | Pass, 232.098 seconds |
| Clean full reference | `timeout 1200 bash stage3/lane/run.sh /workspace/x2jdg33/whole-clean --artifact "$STAGE3_ARTIFACT" > /tmp/x2jdg33-whole-clean.log 2>&1` | Pass, 106366/1/0; upstream oracle exits 1 for the sanction |
| All shards alone | `timeout 5400 python3 -B stage3/lane/measure_shards.py /workspace/x2jdg33/cold --concurrency 1 > /tmp/x2jdg33-cold.log 2>&1` | 377/377 pass and <12 seconds |
| Executed union | `timeout 180 python3 -B stage3/lane/merge_shards.py /workspace/x2jdg33/cold --whole /workspace/x2jdg33/whole-clean --output /workspace/x2jdg33/proofs-4cpu/merged.json > /workspace/x2jdg33/proofs-4cpu/merge.log 2>&1` | Pass, exact 106367 execution multiset |
| Four-active load, union and real mutants | `timeout 2400 python3 -B stage3/lane/prove_budget.py /workspace/x2jdg33/cold /workspace/x2jdg33/whole-clean /tmp/x2jdg33-before-plan.json /workspace/x2jdg33/proofs-4cpu --concurrency 4 > /tmp/x2jdg33-budget-proof-4cpu.log 2>&1` | Exit 0 for implemented checks; hardware limitation explicit |
| Eight-active pilot, quota four | `timeout 2400 python3 -B stage3/lane/prove_budget.py /workspace/x2jdg33/cold /workspace/x2jdg33/whole-clean /tmp/x2jdg33-before-plan.json /workspace/x2jdg33/proofs > /tmp/x2jdg33-budget-proof.log 2>&1` | Exit 1; calibration rejected |
| Guard mutants | `timeout 120 python3 -B stage3/lane/prove_shard_guards.py > /tmp/x2jdg33-shard-guards-final.log 2>&1` | Five mutants caught, source/production hashes unchanged |
| Python fixtures | Each `timeout 120 python3 -B -m unittest discover -s stage3/lane -p FILE -v`, one of the eight `test_*.py` files per invocation, separate log | All 70 pass; exact argv, durations and per-file counts in fixture-ledger.json |
| Generated declarations | `timeout 60 python3 -B stage3/lane/generate_shards.py --check > /tmp/x2jdg33-generation-verified.log 2>&1` | Pass |
| Go manifest | `timeout 180 go test ./stage3/lane -run '^TestShardsMatchManifest$' -count=1 -timeout 2m > /tmp/x2jdg33-go-manifest-verified.log 2>&1` | Pass, .090 seconds |
| Static checks | `timeout 180 go vet ./... > /tmp/x2jdg33-vet-all-verified.log 2>&1`; `gofmt -l cmd internal stage3/lane > /tmp/x2jdg33-gofmt-all.log` | Vet exit 0; gofmt output empty |

The final API/main comparison was also run explicitly (exit 0):

```sh
timeout 30 python3 -B - <<'PYTHON' > /tmp/x2jdg33-api-main-final.log 2>&1
import gzip,hashlib,subprocess
from pathlib import Path
base='23ff023b9ed986834d20404d43e2c363667edf9b'
main=gzip.decompress(subprocess.check_output(['git','show',base+':stage3/lane/evidence/shards/whole-baseline.diff.gz']))
actual=Path('/workspace/x2jdg33/whole-clean/oracle/baseline.diff').read_bytes()
assert main==actual
print('PASS exact API diff bytes unchanged from main',hashlib.sha256(actual).hexdigest())
PYTHON
```

Final syntax checks parsed all six changed/new Python files with `ast.parse`.
The final integrity check also verified current runtime/plan/manifest identities
and byte equality to HEAD for `shard.py`, `merge_shards.py`, `generate_shards.py`,
`expected.json` and `sanctioned-api.json`. `git diff --check` passed.

### Repository-wide Go gate

The required ordinary worker gate was run with this environment and hard limit:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/x2jdg33-typescript
export ADAMIC_BUILD_CACHE_DIR=/workspace/x2jdg33/products
export TMPDIR=/tmp
timeout 1900 go test -count=1 -timeout 30m ./... > /tmp/x2jdg33-go-gate.log 2>&1
```

Interrupted with SIGTERM (exit 143) after confirmed disk exhaustion; this is not a complete or passing gate. The complete output is retained in
[the gate log](evidence/budget-10s/commands/x2jdg33-go-gate.log.gz).
This run permits ordinary worker oracle caching; the direct shard measurements
above used fresh processes and no reused test results.

Observed failure summary:

```text
FAIL	github.com/system-inc/adamic/bridge/tsgo	1185.801s
FAIL	github.com/system-inc/adamic/cmd/adamic	56.933s
FAIL	github.com/system-inc/adamic/cmd/adamic-test262	235.025s
FAIL	github.com/system-inc/adamic/internal/buildcache	632.932s
FAIL	github.com/system-inc/adamic/internal/dry	103.647s
```

The root overlay was full (32G total, 30G used, zero available); bridge
product creation reported no space left on device. I stopped the known gate
process subtree rather than continue a run already failing for exhausted disk.
Removing 447 disposable entries from this task-only scratch product cache freed
8.6G; the verified adapted product and all measured evidence were preserved.
The log also records a bridge subprocess deadline, missing prebuilt command
runtimes, test262 verdict-coverage failures, a header-identity expectation failure,
a DNS-fixture lookup failure and missing cohere mirror entries. These failures
were not baseline-compared, and no source outside the lane was changed.

The repository gate is an additional validation gap. No compiler or other
package source was changed to work around it; the completed scoped proofs above
are reported separately.

The saved before-plan contains main's original 151 unit definitions and tasks.
The measurement helper rebinds their artifact/runtime input hashes to today's
verified inputs without reassigning tasks. Child standalone limit is 30 seconds;
loaded comparisons use 50 seconds, matching the existing Go-wrapper allowance.
No production default or Go deadline changed. The generated runtime hash is
`6a50ed7453428e6bcacfc6e4302c13e0036c3c153c13e912271147fa7ba884f3`.
The plan, manifest, Go declarations and ownership ledger were regenerated after
the last edit to a RUNTIME-listed file.

## Mutants and negative witnesses

| Mutant | Check that caught it | Observed result |
|---|---|---|
| Real one-execution deletion from File034: `codeFixAddMissingParam9.ts`; local expectation/digest/hash rebound | Executed global union and counts | Mutated shard itself passed. Merge alone rejected missing execution: 106365 pass + 1 fail instead of 106366 + 1; exact-union error, no local-verdict or identity rejection |
| Real-work oversized shard, combined original tasks; 298 executions, complete global union/count preserved | Common under-12 budget assertion, then existing loaded timeout | Alone passed upstream in 24.367258 seconds; 12-second assertion caught it. Under same five-loop load it failed in 50.004117 seconds: timed out after 49.5 seconds past its 50-second limit. All seven ordinary companions passed below 40 |
| Disable exactly-one-upstream-task guard in memory | Invalid-inventory fixture | Intended assertion fails; mutant caught |
| Disable enough-selectable-families guard in memory | Invalid-inventory fixture | Intended assertion fails; mutant caught |
| Remove anchors from case regexes in private plan | Complete inventory and suffix-exclusion fixture | Intended assertion fails; mutant caught |
| Pop one planned task in private plan | Complete execution/task preservation fixture | Intended assertion fails; mutant caught |
| Alter one generated `t.Parallel()` declaration in scratch output | `generate_shards.py --check --output SCRATCH` | Exit 1, declarations differ; mutant caught |
| Preliminary naive two-way noEmit split separates compilation from clean-build comparison | Real upstream baseline oracle and existing expected per-unit verdicts | tsbuild groups produced 40 pass/4 fail and 4 pass/40 fail; rejected and replaced with describe-family grouping |

The timing mutant contains real tests, not sleeps or forged ledger times. The
one-case-deletion mutant was arranged to stay locally green so that a timeout,
input-hash error or unrelated local check could not stand in for the union proof.
The five metadata/declaration mutants run against isolated objects or scratch
files and restore the production namespace; no mutant is committed as a live plan.

## Evidence

[Machine-readable proof](evidence/budget-10s/proof.json),
[complete cold ledger](evidence/budget-10s/cold-ledger.json),
[full-lane verdict](evidence/budget-10s/whole-verdict.json),
[fixture ledger](evidence/budget-10s/fixture-ledger.json), and
[context](evidence/budget-10s/context.json) carry the summary results.
`cold-units.tar.gz` retains each unit's actual observations, oracle logs, API and
baseline evidence. `whole-reference.tar.gz` retains the unfiltered reference.
`proof-runs.tar.gz` retains individual loaded/before/mutant runs, saved plans and
child command logs. `exploratory-runs.tar.gz` retains failed experiments, legacy
path samples, earlier candidate probes and both fixture passes. Adapted trees,
node_modules, caches and symlinks are excluded; dropped-case unchanged units
resolve to their corresponding cold archive units. `before-plan.json.gz` retains
the original plan; `baseline.diff.gz` retains the exact API diff. Command logs
are individually gzip-compressed under `commands/`. `sha256.json` hashes all
committed evidence files except itself.

Remaining work is an actual eight-CPU loaded proof with about 2.5-times-alone
calibration, and a same-legacy-path complete <12-second proof if that path is a
required production acceptance target. I chose to preserve both limitations and
push the reviewable lane-only result rather than invent either measurement.
