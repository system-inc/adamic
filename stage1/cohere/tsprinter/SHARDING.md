# Whole-corpus process sharding

Branch `stage1-format/tsprinter-shards`, based on
`stage1-guards/printers` at `5dcbeb7606db4b561c4a4d4fcfc1485535ce91f8`.
Only the tsprinter test harness changes. The printer, parser, independent Go
selectors and pinned upstream outcomes are unchanged.

`shards_test.go` assigns fragments to `min(runtime.NumCPU(), 8)` processes.
It sorts by descending input protocol bytes, ties by original case number,
and assigns each case to the shard with the least assigned bytes, tying by
shard number. Documents move as complete reset/build/print blocks. Each
shard receives both the protocol and its matching JSON specifications.

Every corpus pass uses the same plan: source Node, sanitized native,
JavaScript backend, separate leak execution, native release and each existing
Prettier comparison. The 29 existing mutants also run their complete corpora
in shards. Tiny upstream-only probes and loud gap proofs remain unchanged.
Each process retains childguard, its environment and command arguments.
The macOS leaks tool adds its own stdout report, so each leak invocation's
exit status is checked separately; the ordinary native run checks its answers.

The merger requires exactly one answer per assigned case, a terminating
newline, and a unique answer for every original position. Missing, extra and
repeated positions fail. Reassembly restores original corpus order before
existing byte comparisons and the upstream outcome census. A differing port
answer now names its originating file and fragment. Count failures name the
shard and the expected case at the answer-count boundary; the protocol does
not identify which earlier answer was dropped if a child drops one in the
middle. Any such drop still fails the count check.

## Controls

`TestCorpusShardAssignment` pins the largest-first tie-breaking.
`TestCorpusShardCoverage` rejects missing, extra and unterminated answers.
`TestCorpusShardTransport` runs a real Node child over 23 independent cases,
including cases crossing shard boundaries, and compares reconstructed bytes.

Two uncommitted scratch copies under `review/tsprinter-shard-controls/`
mutated that child's actual output, without changing expected answers:

- `byte`: replaced the first byte of one result. The child exited 0; the test
  exited 1 naming `fixture-00.ts:0` and its byte difference.
- `drop`: omitted the last assigned answer in shard 0. The child exited 0;
  the test exited 1 naming `fixture-20.ts:0`: four answers, expected five.

Logs: `/tmp/tsprinter-shards-byte-mutant.log` and
`/tmp/tsprinter-shards-drop-mutant.log`. Scratch files are not committed.

## Measurement

The complete package uses both required pins:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/tsprinter-tsc/typescript
export ADAMIC_TS_PRETTIER=/tmp/estree-deadlines-library
go test -json -count=1 -parallel=4 -timeout 30m ./stage1/cohere/tsprinter \
  > /tmp/tsprinter-shards-unloaded.log 2>&1
```

Bash `time` records command wall, user and system time separately. The loaded
run starts ten `yes > /dev/null` children, records their PIDs, and stops and
reaps only those children after the same complete package finishes. This is
twice `nproc`, not twice the cgroup quota. No test output is piped. The input
pins are TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8` and
Prettier 3.9.6; the tests independently enforce both pins.

This box reports `nproc=5`, `cpu.max=400000 100000`, 17.6 GB memory,
Go 1.27.1, clang 20.1.8 and Node 24.19.0. Setup logged node/go ready at
0.032s, submodules at 0.079s, markdown dependencies at 0.090s, clang at
0.201s, Go build at 36.295s, cache warm at 36.451s and done at 36.478s.
Test-binary warming was deferred. Setup log:
`/tmp/tsprinter-shards-setup.log`.

| Complete package | Command wall | Package elapsed | User CPU | System CPU | Result |
|---|---:|---:|---:|---:|---|
| Unloaded | 893.135s (14m 53s) | 891.013s | 2864.628s | 178.926s | 51 pass, 0 fail, 0 skip |
| Ten CPU burners | 2360.994s (39m 21s) | 2338.084s | 2920.308s | 134.882s | 51 pass, 0 fail, 0 skip |

Each result includes 15 top-level tests and 36 subtests. All 29 original
printer mutants were caught on both Node and native by normal-exit byte
mismatches. Full corpora are 169,887 expression fragments, 57,799
statement/program fragments and 5,072 documents. The focused tsc audit also
checks 1,729 expressions and 874 statements from the same deliberate root.
The existing upstream outcome census, 13 complete files and single recorded
full-file parse refusal remain unchanged.

**The loaded-under-25-minutes target was not met.** The loaded measurement
uses `-timeout 60m` to finish and expose the complete result. No corpus,
comparison, sanitizer or mutant was removed. The sharding-only implementation
keeps native compilation unchanged. The user's unsharded observations were
995s unloaded and 3,467s loaded; these are supplied baseline observations,
not fresh before measurements made in this unit.

Unloaded load averages before/after: `0.78 1.00 0.49` / `6.85 7.96 5.15`.
Loaded before/after: `2.18 6.31 4.78` / `9.42 13.61 13.34`.
All ten burners were alive at the end and had each consumed approximately
610 to 664 CPU seconds; only these PIDs were killed and reaped afterwards.
The setup reports 17.6 GB memory; `memory.max` is 17,179,869,184 bytes,
exactly 16 GiB. There was no competing test run during either measurement.

Logs: `/tmp/tsprinter-shards-{unloaded,loaded}.log`, matching `-time.log`
and `-machine.log` files. Raw timing and validation facts are in
[results/shard-timing.json](results/shard-timing.json).

## Rejected parallel-build experiment

After the two green measurements, an attempt to reach 25 minutes used the
compiler's existing `Options{Split: true, Jobs: 5}` in this harness, without
changing `internal/`. An await-mutant trial passed in 41.269s wall time.
The full experiment set `ADAMIC_GATE_UNCACHED=1`, so every C object rebuilt.
It **failed** in 766.771s wall time: 49 passed and 2 failed (one mutant and
its parent), with no skips. `return_and_throw_lose_their_keyword` lost a
native shard with exit -1 and empty stderr, naming
`stage1/cohere/yaml/widthTables.ts:3150:KindArrayLiteralExpression`.
The cgroup recorded one OOM kill and a peak of 17,179,897,856 bytes at its
16 GiB limit. A killed child is a failure, never a mutant catch.

That build-path change was reverted completely. The committed implementation
is the one held by both green full-package runs above. The experiment is
reported, not included in successful timing results. Logs:
`/tmp/tsprinter-shards-split-trial.log` and
`/tmp/tsprinter-shards-final-unloaded.log`.

## Final checks and limits

`go vet ./stage1/cohere/tsprinter`, `gofmt -l stage1/cohere/tsprinter` and
`git diff --check` are clean. The first optional race build failed without
a diagnostic. The serialized retry,
`go test -race -p=1 -v ./stage1/cohere/tsprinter -run '^TestCorpusShard' -count=1`,
passed all six tests/subtests in 1.082s, with its complete output in
`/tmp/tsprinter-shards-race-retry.log`. This unit did not run the full repository gate or measure
macOS. Its macOS leaks report handling retains the existing leak verdict;
only the Linux path was exercised here.
