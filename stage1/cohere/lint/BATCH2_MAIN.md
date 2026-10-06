# Batch 2 integration with current main

Batch 2 was merged with origin/main at 5d4c801, preserving the published history and the compiler hot-file split. The merge commit is 90f0ce00295fddcf1986e74a4f5303518d4fb4db. No conflict resolution or compiler-source edits were needed. Main is an ancestor of this branch after the merge.

The serial diagnostic run measures load, lowering, C emission, clang and execution independently. It was deliberately cancelled after representative rebuild measurements; it is not a passing full-suite result. Before this follow-up, the package ran 20 batch-2, 17 baseline and 10 volume mutant rebuilds sequentially. Each rebuild lowers the entire linter, not merely the changed rule. Representative generated C is about 3.38 MB. Logs separate these compilation costs from lint execution.

The harness now runs isolated mutant children in parallel and limits full-program native builds to four. Each child still creates its own source copy, binary and output files, compiles with the unchanged ASan/UBSan flags, executes successfully, and must disagree with the independent pinned Go oracle on both Node and native. Parent tests wait for their children, so throughput samples still follow completed builds. Existing semantic checks and fixtures are retained. Timing annotations remain in the harness for future triage.

## Final verification

Harness commit: `a314e17612cc9ccbcce57eecd535e10713cfba1e`. The complete package command, with no test-name filter, was:

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/lint-batch2/typescript-6.0.3 go test ./stage1/cohere/lint -count=1 -v -timeout 30m > /tmp/batch2-refresh-suite.log 2>&1
```

A Python subprocess wrapper redirected stdout/stderr directly to that file and measured monotonic elapsed time plus `resource.getrusage(RUSAGE_CHILDREN)` after completion. [suite-time.log](batch2_main_evidence/suite-time.log) records:

```text
exit=0 wall=933.395s user=2408.278s system=164.740s maxrss=2258620KiB
```

Wall time is **15 minutes 33.395 seconds**; the Go package reports PASS in 930.632s. Maximum RSS is the largest individual child observation, not simultaneous total memory. The full [suite.log](batch2_main_evidence/suite.log) records all executed tests and differences that caught mutants. Parallel parent test durations printed by Go exclude their parallel children's work; use the outer wall measurement for the complete suite.

| Comparison | Result |
| --- | --- |
| Batch 2 positive and option controls | 23,811 identical bytes; PASS 31.22s |
| Batch 2 upstream | 615 admitted cases, 176,497 identical bytes; PASS 71.62s |
| Combined upstream and generated fixtures | 1,888 captured combinations, 934,817 identical bytes; PASS 64.93s |
| Pinned compiler and current stage1 sources | 77 compiler plus 136 stage1 files, 19,452,499 identical bytes; PASS 144.61s |
| Batch 2 family mutants | All twenty caught on Node and sanitized native |
| Baseline family mutants | All seventeen caught on both engines |
| Volume family mutants | All ten caught on both engines |
| Decoration, scalar fold and position mutants | All three caught on both engines |
| Count-only mutant | Both engines print 3 instead of Go's 2 while ordinary output stays identical |
| Constructor, optional-index, numeric-or and positioned-lastIndexOf gap probes | All existing refusal checks pass |

Every one of the 51 mutants compiled and completed successfully before its wrong answer was counted. Findings, displayed messages, ranges, repair payloads and iterative fixed source still compare with unmodified Go Cohere. TypeScript source is pinned to `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3); Cohere remains pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`.

## Why this exceeds ten minutes

There are **55 full sanitized native builds**, comprising 51 mutants and four ordinary parity builds. Each lowers and emits the whole linter and parser, approximately 3.38 MB of generated C, then invokes clang with the unchanged O1, debug, ASan and UBSan flags. A mutant's changed rule does not reduce the rest of the linked program.

[phases.json](batch2_main_evidence/phases.json) summarizes the measured phase annotations:

| Phase | Sum of per-build elapsed seconds |
| --- | ---: |
| Load | 22.146 |
| Lower | 33.430 |
| Emit C | 18.690 |
| clang build and link | 2,195.891 |

clang is **96.73% of the sum of these build phases**. These elapsed durations overlap during parallel builds and must not be added to the outer wall time. clang median is 38.358s per build, minimum 25.331s, maximum 67.511s. The serial diagnostic run measured individual mutant clang times of 23.804s, 35.104s and 26.308s, versus less than a second for their load/lower/emit work and milliseconds for native controls. It was cancelled at 141.212s and is explicitly marked incomplete in [baseline-cancelled.log](batch2_main_evidence/baseline-cancelled.log); [baseline.log](batch2_main_evidence/baseline.log) is diagnostic evidence only.

The corpus also has real execution cost: sanitized native took **86.266s** to produce its complete findings and fixed-source output over 213 files. The complete corpus test takes 144.61s including compilation, Go and Node. Combined and batch-only upstream checks capture and run actual Go rule tests as well.

Observed cause of this run's excess is repeated whole-program compilation plus corpus verification, dominated by clang within the build phases. The old serial mutant scheduling plausibly explains the triage timeout: the prior batch-only report already measured twenty rebuilds at 798.47s, before the other mutant families and corpora. That is an inference about triage, whose exact command, load and timeout stack were not supplied. No exact reproduction of its thirty-minute timeout or same-machine speedup ratio is claimed. The bounded concurrent scheduling preserves all checks and completes this isolated package in 15m33s; it does not bring it below ten minutes. Further reduction needs less compilation work or CI shards, rather than removal of oracle comparisons.

## Other checks and setup

Commands below also redirected all output directly to the linked files and exited zero:

```bash
go vet ./... > /tmp/batch2-refresh-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 10m > /tmp/batch2-refresh-oracle.log 2>&1
gofmt -l stage1/cohere/lint > /tmp/batch2-refresh-gofmt.log
git diff --check > /tmp/batch2-refresh-diffcheck.log 2>&1
git diff --cached --check > /tmp/batch2-refresh-source-diffcheck.log 2>&1
```

[vet.log](batch2_main_evidence/vet.log), [gofmt.log](batch2_main_evidence/gofmt.log) and diff logs are empty. The [external oracle](batch2_main_evidence/oracle.log) passes in 9.549s, with native and Node cache hits zero and misses one each. The initial separate [gap probe run](batch2_main_evidence/gaps-before.log) passes in 1.267s; those probes are also rerun in the complete suite.

`bash cloud/setup.sh` succeeded; [setup.log](batch2_main_evidence/setup.log) prints Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 1s, build-cache warm 159s, total 159s. Environment: `/workspace/adamic-tools/env.sh`. Go 1.27.1, clang 20.1.8, Node 24.19.0. Direct nproc is 5, cgroup CPU quota is four cores (`400000 100000`), reported memory 17.6 GB. A [process snapshot](batch2_main_evidence/processes.log) records active builds and the quota.

## Limits

Nine inherited unsupported-recovery combinations remain explicit bounded refusal checks, not successful lint parity. Two malformed negative-zero cases and the inherited admitted malformed fixtures retain their findings-only contract. No syntax or parser coverage is silently added or dropped. General type-aware rules, JSX and recovery remain outside this slice.

Throughput and historical profile/snapshot tests remain opt-in and were skipped in this complete default suite. No new findings-per-second measurement is claimed. Full repository `go test ./...` was not run: this follow-up ran the complete lint package with the compiler corpus enabled, repository vet and the named uncached external oracle. No TypeScript rule, parser, compiler or runtime source was manually edited beyond incorporating main's published history.
