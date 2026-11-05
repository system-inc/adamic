# API sanction wrapper split

Task: #wqzsq36

Baseline: origin/main at d40c2470eb76fc4cd89d30235bc52650881c9c2e.
Linux amd64, Go 1.27.1, Node 24.19.0, clang 20.1.8. The box has
five allowed CPUs and a four-CPU cgroup quota. Setup completed in 380.255 s.
The pinned TypeScript input is 050880ce59e30b356b686bd3144efe24f875ebc8.

## Measurement and choice

[cases.tsv](cases.tsv) records every unchanged Python case alone, with
GOMAXPROCS=1 and affinity to CPU 0, then GOMAXPROCS=4 and affinity to CPUs
0-3. Affinity matters because GOMAXPROCS alone does not restrict Node.
These measurements followed setup and ran without another test workload.
All 44 invocations passed. Totals: 74.060 s on one CPU, 48.918 s on four.
The longest individual case was 6.609 s on one CPU.
Earlier measurements overlapping setup were excluded.

No case builds TypeScript. Each generator subprocess loads TypeScript and
reads and parses the pinned declarations and proof sources again. Acceptance
and undo cases also run a separate Node normalizer through the API guard. A Node
CPU profile of before_mismatch_names_row recorded 195 ms in module wrapSafe,
146 ms in the generator's visit, 111 ms in garbage collection, 92 ms in
TypeScript scan, 89 ms in forEachChildRecursively, 82 ms in spawnSync and
70 ms in ownerPath. These are sampled self times, not complete phase times.

A wrapper-local NODE_COMPILE_CACHE experiment did not improve runtime.
Three one-CPU rejection invocations took 6.445 s without the cache and
6.619 s sharing it. Explicitly prewarming TypeScript also did not help:
6.622 s without versus 6.617 s with, excluding the extra warmup cost.
The cache mechanism was therefore removed. Sharing a live generator would
require a new protocol and isolation for its mutable claimed-owner and edit
state. The chosen change preserves the existing independent CLI executions
and partitions their work into eight descriptive wrappers, each measured
under 25 s. Each wrapper calls t.Parallel(). No Python case or assertion
changed.

The kill stays at runPythonCommand's 55 s. These runs measured the hand's
first cut, which also lowered it to 25 s for the generator wrappers.
Correctness dropped that hunk before submitting: the slowest wrapper reads
19.35 s under the whole lane's parallel load, and a fixed kill that close to
a loaded time is how TestHostStatusMutants went red on an innocent change
(#xz91ar4). The bar is held by the split, and a slower run fails at 55 s.

The coverage check discovers Python tests and checks their exact-once union
against groups whose Go wrappers actually exist. A new Python case, a missing
case, a duplicate selector, a deleted wrapper or an empty wrapper fails it.

## Mutant

Apply the patch with zero context, then run:

```sh
git apply --unidiff-zero stage3/lane/drop-api-case-mutant.patch
go test ./stage3/lane -run '^TestLaneGenerateApiPythonGroupsCoverEveryTestOnce$' -count=1 -v
```

The mutant drops test_api_sanction_line_applied_and_guard_accepts from the
wrapper union. It exited 1, failing the coverage test in 0.09 s with:

```text
test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_line_applied_and_guard_accepts covered 0 times, want exactly once
```

The runner restored the source in a finally block. The same test then
passed in 0.09 s. Raw logs are in /workspace/scratch/drop-api-case-mutant.log
and /workspace/scratch/api-union-restored.log.

## Whole-suite validation

Products were built ahead with `go test -json -count=1 -timeout=30m -run
'^TestProduct_' ./stage3/lane/... ./cmd/adamic-gate/...`. All passed;
TestProduct_Stage3AdaptedTree took 283.99 s, the census exports 6.04 s,
and the selector probe 1.60 s.

Each complete command used `go test -json -count=1 -timeout=90m` with the
package pattern below. Loaded means GOMAXPROCS=1 and taskset -c 0. Normal
means GOMAXPROCS unset and the box's ordinary CPU affinity. The package
limit allows all 398 upstream shards to run. In these runs each API
wrapper's own kill was the first cut's 25 s; the submitted change keeps
55 s, so none of these results could be cut shorter by it.

| Command package pattern | One CPU wall seconds | Normal wall seconds |
| --- | ---: | ---: |
| `./stage3/lane/...` | 2600.045 | 799.918 |
| `./cmd/adamic-gate/...` | 2.868 | 1.644 |

Every command exited 0. [units.tsv](units.tsv) records every test and package
result, with its elapsed time in both modes. The gate includes the repository
name and parallelism censuses. No test was skipped.

| API wrapper | One CPU seconds | Normal seconds |
| --- | ---: | ---: |
| TestLaneGenerateApiClaimRejectionsPython | 8.65 | 14.60 |
| TestLaneGenerateApiInsertDeclarationRejectionsPython | 11.02 | 17.64 |
| TestLaneGenerateApiInsertFieldRejectionsPython | 4.61 | 6.27 |
| TestLaneGenerateApiInsertGuardsPython | 12.80 | 19.35 |
| TestLaneGenerateApiLineRejectionsPython | 8.70 | 15.25 |
| TestLaneGenerateApiMemberGuardsPython | 11.13 | 15.67 |
| TestLaneGenerateApiOwnerRejectionsPython | 6.97 | 11.30 |
| TestLaneGenerateApiPropertyGuardsPython | 12.51 | 18.26 |

The coverage test passed in both complete runs. Before the broad runs, a
focused one-CPU API run also passed all eight wrappers and the coverage
check in 79.322 s; its slowest wrapper took 13.74 s.

Raw command logs and the independent case/profile/cache experiments are
under /workspace/scratch. The final whole-suite logs are lane-loaded.jsonl,
gate-loaded.jsonl, lane-normal.jsonl and gate-normal.jsonl. An early broad
attempt was stopped before its selected 30-minute package timeout could
truncate the one-CPU suite; its log is separate and is not included in the
final results or units.tsv.
