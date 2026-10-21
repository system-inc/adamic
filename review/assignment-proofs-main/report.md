Built: rebuilt the assignment-proof acceptance suite on main as twelve top-level tests; the delivery is test-only.
Commits: adapted fixtures 6b2e806407c7fe1c3e801ba5ab722cde2172abdc from 4d17f1d4; original base aaf68770a93ec5fa8d744e5df91844ddc0d9d1c0; ccf8c172 omitted.
Commands/results: all six programs pass stock Node, both backends, ASan/UBSan, leak detection and release builds; counts update/check and integration lane checks pass.
Mutants: six source mutants compile with zero stock TypeScript diagnostics and are caught by Node stdout byte comparisons; each mutated program also agrees with both backends and native release.
Not covered: the original implementation commit's broader member/setter/throw semantics, proof-check elision, full scanner execution, whole-package tests or the full gate.

This delivers step 12 task #2dxms6q under #cvhj5fk on compiler/assignment-proofs-main. The requested criterion for omitting ccf8c172 was observed: all six acceptance programs compile on untouched origin/main and exactly match the ruled Node bytes before the fixtures commit is applied. Subsequent tests use unchanged main production code. “Redundant” here is limited to these six acceptance outcomes; it does not assert that every behavior covered by the old implementation commit is already on main.

The fixtures commit was applied with `git cherry-pick --no-commit 4d17f1d4`. Its two conflicts were resolved as follows: the missing oracle registration file became top-level tests for only the six acceptance programs, without references to the seven absent implementation fixtures; counts.md started with main's table and was regenerated. No production compiler or runtime file changed. Current landed main was merged before verification. Further landed-main merges and the final tip are identified in the delivery message; unlanded worker branches were not merged.

**Coverage and independent test leaves**

Each program and each mutant has its own top-level Test function with Parallel as its first statement. No subtest was added and no pending test needs a skip. The six `.a` programs and expectations.json are unchanged from the fixtures commit. The tests first run stock TypeScript 6.0.3, then compare the source Node output to the ruled golden, and hold generated JavaScript, sanitized native and release native to that source. Every successful native run also passes leak detection.

`internal/oracle/assignment_proofs_test.go:21` registers the six rows through the existing additionalFixtureCounts hook, keeping the programs out of the shared subtest loop. The twelve top-level tests start at `:147`; their common execution helper starts at `:84`. The source-mutant helper at `:122` requires exactly one real expression replacement, zero stock TypeScript diagnostics, normal exit and stderr, and stdout different from the original golden.

The conservative resolution assumption is that this unit retains only the requested six fixtures when the compiler commit is omitted. The seven absent implementation fixtures are not silently replaced with skips or imported from the omitted commit.

**Seconds for every new test**

One Codex instance had nproc=5 with a four-CPU quota (cpu.max=400000 100000). GOMAXPROCS=4 was set for the runs. Every row below is an individually invoked `go test` process, including test-binary build/startup, stock TypeScript, frontend lowering, Node, both native compilations, executions and leak checking. The first row includes a cold oracle test-binary rebuild. The right column records the later post-merge run with four parallel test leaves. All are below 60 seconds.

| Top-level test | Individual full command seconds | Post-merge test seconds |
|---|---:|---:|
| TestAssignmentProofScannerKeyword | 15.644 | 1.43 |
| TestAssignmentProofScannerKeywordMutant | 2.678 | 1.50 |
| TestAssignmentProofIfNarrowing | 2.825 | 1.47 |
| TestAssignmentProofIfNarrowingMutant | 2.725 | 1.43 |
| TestAssignmentProofReturnDefined | 2.829 | 1.42 |
| TestAssignmentProofReturnDefinedMutant | 2.774 | 1.43 |
| TestAssignmentProofChained | 2.728 | 1.44 |
| TestAssignmentProofChainedMutant | 2.783 | 1.29 |
| TestAssignmentProofCompound | 2.672 | 1.29 |
| TestAssignmentProofCompoundMutant | 2.731 | 1.38 |
| TestAssignmentProofWiderTarget | 2.777 | 1.30 |
| TestAssignmentProofWiderTargetMutant | 2.829 | 1.32 |

Every individual test command was `ADAMIC_GATE_UNCACHED=1 GOMAXPROCS=4 go test ./internal/oracle -run '^<test-name>$' -count=1 -v -timeout=55s`, with a separate output log. A 60-second subprocess watchdog also bounded each full command. No test was killed or timed out.

**Each mutant and its catcher**

All six mutation-check tests pass because the deliberately wrong source produces different bytes from its original ruled golden. Unlike a compiler-error-only mutation, each mutated source passes strict TypeScript and successfully runs in Node, JavaScript, ASan/UBSan and release native; no sanitizer or compiler error substitutes for the intended stdout comparison.

| Mutant | Node golden | Changed source stdout | Catcher |
|---|---|---|---|
| 01: keyword return/store becomes Identifier | Starts `128 128`, `83 83` | Both become `80 80` | Node stdout bytes |
| 02: absent RHS becomes “missing” | Second call `-1 2` | Second call `7 2` | Node stdout bytes |
| 03: omit cache write | `value1 value1 1`, `value2 value2 2` | Cache prints undefined on both calls | Node stdout bytes |
| 04: omit b's assignment | `value1 value1 value1 1` | `value1 value1 undefined 1` | Node stdout bytes |
| 05: += becomes = | `12 12 1`, `14 14 2` | `2 2 1`, `2 2 2` | Node stdout bytes |
| 06: omit wide's write | `128 128 1` | `128 0 1` | Node stdout bytes |

**Counts and commands**

The required update ran `GOMAXPROCS=4 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m -args -update-counts`; it passed in 66.561s and added exactly six rows. This is the existing counts aggregator, whose code was not touched, rather than one of the new test leaves. The later no-update verification passed in 45.393s. Every new fixture has allocations equal to frees:

| Fixture | Allocations | Frees | Retains | Releases | Peak |
|---|---:|---:|---:|---:|---:|
| 01_scanner_keyword | 18 | 18 | 18 | 32 | 6 |
| 02_if_narrowing | 6 | 6 | 5 | 13 | 3 |
| 03_return_defined | 8 | 8 | 6 | 18 | 4 |
| 04_chained | 4 | 4 | 6 | 15 | 3 |
| 05_compound | 8 | 8 | 0 | 8 | 4 |
| 06_wider_target | 4 | 4 | 0 | 4 | 4 |

After merging main, the focused suite ran:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 GOMAXPROCS=4 go test ./internal/oracle -run '^TestAssignmentProof(ScannerKeyword|IfNarrowing|ReturnDefined|Chained|Compound|WiderTarget)(Mutant)?$' -count=1 -parallel=4 -v -timeout=55s > /workspace/scratch/assignment-proofs-main/merged-leaves.log 2>&1
GOMAXPROCS=4 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m > /workspace/scratch/assignment-proofs-main/counts-check.log 2>&1
```

All twelve leaves passed. Every test output went to a file, never through a pipe. Native flags were clang `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all` for sanitizers and `-O2` for release/count builds. Node was 24.19.0, Go 1.27.1, clang 20.1.8. These timings describe test execution, not performance claims about emitted programs.

**Integration lane check**

After committing, integration's exact command ran from the repository root with the toolchain environment sourced:

```sh
(git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -) > /workspace/scratch/assignment-proofs-main/lane-checks.log 2>&1
```

It returned 0 and printed:

```text
lane checks 1.8 s: gofmt and tools on 1 Go files, t.Parallel on 1 test packages; vet 1 packages
```

The only new process command in the changed Go test file is declared `node`; integration also checked gofmt, all twelve Parallel declarations and vet for internal/oracle. No lane violation needed a fix. After the report commit and the next landed-main merge, the exact command passed again: `lane checks 0.9 s: gofmt and tools on 1 Go files, t.Parallel on 1 test packages; vet 1 packages`. Its log is preserved as final-lane-checks.log.gz.

**Toolchain setup**

Setup used GOPROXY=https://proxy.golang.org|direct and passed. Go ready 0.025s; Node ready 0.025s; submodules 0.099s; Markdown ready 0.111s (validated installed bytes skipped, step 0.027s); clang ready 0.172s; Go build ready 44.612s; test binaries deferred 44.823s; build cache warm 44.824s; done 44.851s. Its environment was /workspace/adamic-tools/env.sh. This one-time environment preparation is separate from the per-test setup included in the table. No Go-module 403 occurred. Raw current-main baseline results, per-test commands/timings and compressed logs are included beside this report.
