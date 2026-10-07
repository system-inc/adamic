Built 11 production-only reverse patches and a parallel named-Go-test checker under verify/catalog.
Targets: each row records its made_against SHA; refreshed entry 07 and refusal entries target e8ba3d5d81de4d3773c723914fccd4c76248b965. Branch: devtools/regression-catalog.
Commands: original oracle pairs passed clean and failed patched; new named lowerer tests do the same in three interleaved rounds.
Mutants: nine original faults and two refusal faults caught; named checker probes reject extra failures, missing required subtests, wrong diagnostics and invalid command patterns/counts.
Not covered: entries 10 and 13-16 remain skipped. Only verify/ changes are committed; compiler faults run in disposable worktrees.

The original nine oracle patches and observations target the origin/main fetched for the initial catalog. Their whole-package counts below remain historical observations on e011f8f, not refreshed measurements on newer main. All patches apply independently, change only non-test production code, and were restored after each run. Each whole-package run uses `ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle`. Complete logs are in [logs](logs/); JSON events retain all passing and failing test names and diagnostics.

`catalog.json` is a sixteen-row array. `command` is the authoritative, shell-quoted `go test` invocation with an exact anchored `-run` pattern. `expected_tests` names every required failing test or subtest; all must fail, and failures outside those tests and their reporting parents are rejected. `expected_failure_line` supplies the diagnostic. The checker sets `ADAMIC_GATE_UNCACHED=1` and uses `-count=1`. `made_against` identifies the exact main snapshot used to prepare each patch, independent of the commit passed to the checker. For skips it identifies the assessment snapshot. Legacy `main_sha`, `test_command`, `test_args`, fixture and evidence fields remain. Skipped rows have a null command and a reason.

Run `source /workspace/adamic-tools/env.sh` on this machine, or source the environment file printed by `bash cloud/setup.sh` on another machine, then:

```sh
verify/catalog/check.sh origin/main
verify/catalog/check.sh origin/main --entry 07 -jobs 1
verify/catalog/check.sh origin/main -jobs 3
```

The checker checks patch applicability, runs the named Go test command on a clean control, applies the patch, and requires every expected test to fail with the recorded diagnostic. Source line-number drift is ignored; diagnostic text is exact. Each active entry has its own disposable worktree and pinned submodules. Workers run concurrently (`-jobs N` or `--jobs N`, default `nproc`); complete status blocks print in numeric entry order. Skips need no worktree. Each worker verifies restoration to a clean state before removing only its own worktree. Logs stay in separate numbered directories below the printed temporary log directory, with a results.json manifest. Per-entry wall times include setup, tests and cleanup; total wall time includes the entire invocation. Patch drift reports `no-longer-applies (needs a refresh)`. A red control or a different failure is an investigation result and makes the checker exit 1, never a successful reproduction.

On the original e011f8f target, the whole oracle control exited 0 with no failed tests. The table below counts failed native comparison fixtures separately from failed fixture subtests anywhere in the package, including counted builds. Counts-table row changes are separate because `TestCountsAreRecorded` reports them at its parent rather than failing each fixture subtest. Total detecting fixtures is the distinct union of failed fixture subtests and changed count rows. The complete lists, including failures in mutant self-tests, are in each row's `whole_oracle` object.

| Number | Name | Fix | Primary fixture | Expected failure line | Native fixtures failing | Distinct failed fixture subtests | Changed count rows | Total detecting fixtures |
|---|---|---|---|---|---:|---:|---:|---:|
| 01 | shared-slice-append | `7b6f9864` | `shared_slice_append.a` | `oracle_test.go:613: exit codes differ` | 1 | 1 | 0 | 1 |
| 02 | liveness-throw | `2dbee114` | `throw_keeps_old_value.a` | `oracle_test.go:613: exit codes differ` | 2 | 2 | 0 | 2 |
| 03 | defined-lent | `fe62d30a` | `borrow_defined_lent.a` | `oracle_test.go:613: exit codes differ` | 2 | 2 | 27 | 27 |
| 04 | borrowed-array-move | `95794086` | `borrow_element_super_move.a` | `oracle_test.go:613: exit codes differ` | 1 | 1 | 1 | 1 |
| 05 | spread-method-reuse | `9dd2e9ab` | `reuse_spread_method_alias.a` | `oracle_test.go:613: stdout differs` | 2 | 2 | 0 | 2 |
| 06 | constructor-capture-region | `d84813c5` | `regions_constructor_capture.a` | `oracle_test.go:613: exit codes differ` | 1 | 1 | 1 | 1 |
| 07 | borrowed-element-reads | `658dbc4f` | `borrow_element_virtual_store.a` | `oracle_test.go:613: exit codes differ` | 1 | 1 | 2 | 2 |
| 08 | narrowed-number-field | `675770b7` | `e4eec87_f1_field_narrowed.a` | `oracle_test.go:604: stdout differs` | 4 | 4 | 3 | 4 |
| 09 | literal-undefined-field | `962f412b` | `e4eec87_u01_undefined_field_widened.a` | `oracle_test.go:613: stdout differs` | 2 | 2 | 2 | 2 |

Only their primary fixture fails in the native comparison sweep: 01, 04, 06, 07. Including counted-build fixture failures and changed count rows, entries detected only by their primary fixture are: 01, 04, 06. Entries detected only by the fixture family added with their own fix are: 01, 02, 04, 05, 06, 08, 09. This describes observations from the full uncached runs, not an inference from the targeted commands.

Entry 04 uses the original `super.run` fixture. The later fix in entry 07 masks its virtual-move companion. Entry 07 removes only the unsafe virtual-call proof; its harmless MakeError borrowing optimization is retained, and the throw fixture is a passing neighbor. Entry 08 is a checked fixture: the required inserted check must fire, so restoring the old field read can also be detected by comparison with the checked JavaScript backend. Entry 09 retains the runtime helper and later optional-field machinery still used by main, while restoring the old required reads and literal slot fitting.

Exact targeted commands (each was run three times without the patch and three times with it):

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^shared_slice_append[.]a$'
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^throw_keeps_old_value[.]a$'
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^borrow_defined_lent[.]a$'
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^borrow_element_super_move[.]a$'
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^reuse_spread_method_alias[.]a$'
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^regions_constructor_capture[.]a$'
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^borrow_element_virtual_store[.]a$'
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^e4eec87_f1_field_narrowed[.]a$'
ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^e4eec87_u01_undefined_field_widened[.]a$'
```

Commit discovery used `git log --all` after fetching all branch refs, because the checkout initially fetched only main:

```sh
git fetch origin '+refs/heads/*:refs/remotes/origin/*'
git log --all --oneline -- internal/native/runtime/string_append.c
git log --all --oneline --grep='Return object for typeof null\|Use checker type identity'
git log --all --oneline -G 'useOfThis' -- internal/lower/class.go
git log --all --oneline origin/coverage/proven-guards-relations
git show da12012 --stat
git show cbcc765c --stat
git merge-base --is-ancestor <fix> e011f8f
```

The shared-slice append fix is `7b6f9864cd22db4c9629b2d62aa4c08c2f2637b7`, found by the string_append.c history. The log subject search finds `8aa2f91c1d199a95fd64d92733bf093b8df2dd65` for typeof null and `f914cea29787df2a506d36d2f80bd0f5ff77d78a` for repeated class type arguments. The useOfThis hunk search finds `0cf467cdef1fba747a5b0019bf6f2f1b7651d30c` for constructor arrows. Their diffs and fixture registrations confirm the intended fixes. [discovery.txt](discovery.txt) preserves subjects, file lists and ancestor checks.

| Number | Name | Commit | Reason skipped |
|---|---|---|---|
| 10 | override-representation | `1f964412352567fca5ee39482205217706a078b2` | Not an ancestor of target main; its oracle fixture override_same_representation.a is a supported neighbor, not a failing refusal probe. The actual regressions are lowerer refusal tests. |
| 13 | typeof-null | `8aa2f91c1d199a95fd64d92733bf093b8df2dd65` | Fix and its oracle fixtures are not on target main. Main already has the pre-fix typeOf implementation; there is no fix to undo on this target. |
| 14 | constructor-arrow-this | `0cf467cdef1fba747a5b0019bf6f2f1b7651d30c` | Fix is not on target main and adds only lowerer refusal tests, no oracle fixture. Main already lacks the arrow capture refusal. |
| 15 | class-instance-key | `f914cea29787df2a506d36d2f80bd0f5ff77d78a` | Fix and its oracle fixtures are not on target main. Main already uses representation-based instantiation keys, so cannot undo the absent fix. |
| 16 | proven-guards-relations | `da1201283cce13e4910e706ff1ff44f1a425c381` | Supplied SHA changes only whitespace in a log. Nearby cbcc765c adds coverage, no compiler fix; mutable_kind_guard.a is an unresolved disagreement kept under notes, absent from main and from the registered oracle fixtures. |

Entry 16 is specifically not a repaired regression: `da12012` only trims whitespace in a log. Nearby `cbcc765ca953aecb6133eb7f1d2462b0c34ea99c` says it keeps the mutable-kind disagreement in notes and changes no compiler code. Its `disagreements.json` records Node/backend output `branch undefined\n` with exit 0 versus native exit 70 and `adamic: panic: compiler bug: a field the checker proved is there is missing`. Treating that as a fix would manufacture a catalog entry.

Toolchain setup printed:

```text
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
v24.19.0
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (174s)
setup: done in 174s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Setup was a single required preparation run, not a before/after measurement; its starting load was not captured. The ending machine context is recorded in `environment.json`. No new cache was added, and every oracle result run used the uncached mode.

Targeted invocation durations below are best of three interleaved clean/patched pairs on the same box and target commit. They include Go compilation and test execution. A faulty program can end earlier, so these numbers characterize fixture detection runs and do not imply a compiler speedup.

| Loop | Before (control) | After (reverse patch) | Instrument |
|---|---:|---:|---|
| 01 shared-slice-append | 2.395s | 2.572s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^shared_slice_append[.]a$'` |
| 02 liveness-throw | 2.384s | 2.272s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^throw_keeps_old_value[.]a$'` |
| 03 defined-lent | 2.219s | 2.280s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^borrow_defined_lent[.]a$'` |
| 04 borrowed-array-move | 2.457s | 2.275s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^borrow_element_super_move[.]a$'` |
| 05 spread-method-reuse | 2.254s | 2.212s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^reuse_spread_method_alias[.]a$'` |
| 06 constructor-capture-region | 2.090s | 2.179s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^regions_constructor_capture[.]a$'` |
| 07 borrowed-element-reads | 2.467s | 2.261s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^borrow_element_virtual_store[.]a$'` |
| 08 narrowed-number-field | 2.557s | 8.344s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^e4eec87_f1_field_narrowed[.]a$'` |
| 09 literal-undefined-field | 2.335s | 2.323s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^e4eec87_u01_undefined_field_widened[.]a$'` |

01 build flags: commit `e011f8f60899586d6373a5ccb07335ad82cfbf3c`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; node `v24.19.0`; uncached oracle results. Clean load before/after `6.72 7.57 4.31 2/163 15671` / `6.34 7.48 4.29 2/163 15753`; patched load before/after `6.34 7.48 4.29 2/163 15754` / `6.34 7.48 4.29 2/165 15844`. Harness builds: C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, separate leak checks; no changed Go flags.

02 build flags: commit `e011f8f60899586d6373a5ccb07335ad82cfbf3c`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; node `v24.19.0`; uncached oracle results. Clean load before/after `5.68 7.25 4.29 2/164 16072` / `5.68 7.25 4.29 3/164 16154`; patched load before/after `5.38 7.17 4.27 2/164 16325` / `5.38 7.17 4.27 2/164 16406`. Harness builds: C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, separate leak checks; no changed Go flags.

03 build flags: commit `e011f8f60899586d6373a5ccb07335ad82cfbf3c`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; node `v24.19.0`; uncached oracle results. Clean load before/after `5.38 7.17 4.27 2/164 16407` / `5.11 7.08 4.26 2/163 16488`; patched load before/after `4.93 6.98 4.26 2/167 16690` / `4.69 6.89 4.25 2/170 16785`. Harness builds: C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, separate leak checks; no changed Go flags.

04 build flags: commit `e011f8f60899586d6373a5ccb07335ad82cfbf3c`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; node `v24.19.0`; uncached oracle results. Clean load before/after `4.37 6.67 4.23 2/176 17332` / `4.26 6.61 4.22 2/176 17418`; patched load before/after `4.26 6.61 4.22 2/176 17419` / `4.26 6.61 4.22 1/176 17509`. Harness builds: C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, separate leak checks; no changed Go flags.

05 build flags: commit `e011f8f60899586d6373a5ccb07335ad82cfbf3c`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; node `v24.19.0`; uncached oracle results. Clean load before/after `3.54 6.34 4.17 2/171 17898` / `3.54 6.34 4.17 1/171 17984`; patched load before/after `3.54 6.34 4.17 1/171 17985` / `3.34 6.25 4.15 1/171 18068`. Harness builds: C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, separate leak checks; no changed Go flags.

06 build flags: commit `e011f8f60899586d6373a5ccb07335ad82cfbf3c`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; node `v24.19.0`; uncached oracle results. Clean load before/after `2.98 6.08 4.12 1/170 18272` / `2.98 6.08 4.12 1/170 18356`; patched load before/after `2.98 6.02 4.11 1/173 18525` / `2.98 6.02 4.11 1/175 18610`. Harness builds: C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, separate leak checks; no changed Go flags.

07 build flags: commit `e011f8f60899586d6373a5ccb07335ad82cfbf3c`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; node `v24.19.0`; uncached oracle results. Clean load before/after `2.76 5.83 4.08 2/192 18895` / `2.76 5.83 4.08 1/190 18998`; patched load before/after `2.76 5.83 4.08 1/190 18999` / `2.78 5.78 4.07 2/191 19089`. Harness builds: C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, separate leak checks; no changed Go flags.

08 build flags: commit `e011f8f60899586d6373a5ccb07335ad82cfbf3c`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; node `v24.19.0`; uncached oracle results. Clean load before/after `3.51 5.88 4.12 2/206 19765` / `3.39 5.82 4.10 2/206 19845`; patched load before/after `10.23 7.19 4.65 8/230 20354` / `10.70 7.34 4.72 2/205 20545`. Harness builds: C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, separate leak checks; no changed Go flags.

09 build flags: commit `e011f8f60899586d6373a5ccb07335ad82cfbf3c`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; node `v24.19.0`; uncached oracle results. Clean load before/after `7.57 6.94 4.70 7/212 21103` / `7.36 6.91 4.71 2/206 21194`; patched load before/after `7.36 6.91 4.71 2/206 21195` / `7.36 6.91 4.71 2/206 21278`. Harness builds: C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`, separate leak checks; no changed Go flags.

Original checker output on e011f8f:

```text
target e011f8f60899586d6373a5ccb07335ad82cfbf3c
logs /tmp/adamic-gate/adamic-catalog-logs-qdg0u4q3
01 shared-slice-append: applies-and-fails-as-recorded
02 liveness-throw: applies-and-fails-as-recorded
03 defined-lent: applies-and-fails-as-recorded
04 borrowed-array-move: applies-and-fails-as-recorded
05 spread-method-reuse: applies-and-fails-as-recorded
06 constructor-capture-region: applies-and-fails-as-recorded
07 borrowed-element-reads: applies-and-fails-as-recorded
08 narrowed-number-field: applies-and-fails-as-recorded
09 literal-undefined-field: applies-and-fails-as-recorded
10 override-representation: skipped (Not an ancestor of target main; its oracle fixture override_same_representation.a is a supported neighbor, not a failing refusal probe. The actual regressions are lowerer refusal tests.)
11 refuse-definite-assignment: skipped (Fix is on main but adds only internal/lower/definite_assignment_test.go inline probes, no oracle fixture or Node comparison. Cannot meet the required oracle command without changing tests.)
12 refuse-suppression-directives: skipped (Fix is on main but adds only internal/lower/suppression_directives_test.go inline probes, no oracle fixture or Node comparison. Cannot meet the required oracle command without changing tests.)
13 typeof-null: skipped (Fix and its oracle fixtures are not on target main. Main already has the pre-fix typeOf implementation; there is no fix to undo on this target.)
14 constructor-arrow-this: skipped (Fix is not on target main and adds only lowerer refusal tests, no oracle fixture. Main already lacks the arrow capture refusal.)
15 class-instance-key: skipped (Fix and its oracle fixtures are not on target main. Main already uses representation-based instantiation keys, so cannot undo the absent fix.)
16 proven-guards-relations: skipped (Supplied SHA changes only whitespace in a log. Nearby cbcc765c adds coverage, no compiler fix; mutable_kind_guard.a is an unresolved disagreement kept under notes, absent from main and from the registered oracle fixtures.)
```

Checker negative probes:

```text
05 spread-method-reuse: applies-but-failure-not-as-recorded (needs investigation)
Wrong expected diagnostic mutant: rejected
05 spread-method-reuse: no-longer-applies (needs a refresh)
Drifted patch mutant: rejected as no-longer-applies
05 spread-method-reuse: baseline-fails-or-fixture-missing (needs investigation)
Already-red baseline mutant: rejected
Checker mutant probes passed; worktree restored
05 spread-method-reuse: invalid-patch (production code only)
Test-file patch mutant: rejected
05 spread-method-reuse: baseline-fails-or-fixture-missing (needs investigation)
Missing fixture mutant: rejected
Cached command mutant: rejected
Additional checker probes passed
```

Validation scope: complete oracle package on clean main and once per each of the nine reverse patches, each primary fixture in three interleaved control/mutant pairs, checker negative probes and final clean-worktree check. The full repository gate was not run: this unit adds verification artifacts only.

Named lowerer entries on e8ba3d5d81de4d3773c723914fccd4c76248b965:

| Entry | Command | Recorded failure line | Required failing subtests |
|---|---|---|---:|
| 11 | `go test -count=1 ./internal/lower -run '^TestDefiniteAssignmentIsRefused$'` | `definite_assignment_test.go:22: got <nil>, want definite assignment refusal` | 4 |
| 12 | `go test -count=1 ./internal/lower -run '^TestSuppressionDirectivesAreRefused$'` | `suppression_directives_test.go:23: got <nil>, want directive refusal` | 24 |

Each new command passed without its patch and failed with it in three interleaved rounds. The whole lowerer control passed; each whole-package mutant run failed only its named refusal test family, with all four definite-assignment cases and all twenty-four suppression cases failing. Test sources and fixtures are unchanged. Complete output is in `logs/named-*.log`, and full failure-name inventories are in `whole_lowerer`.

Entry 10 was rechecked after `git fetch origin`: `git merge-base --is-ancestor 1f96441 origin/main` returned 1 for e8ba3d5d81de4d3773c723914fccd4c76248b965, so it remains skipped as requested. Entry 07 is now refreshed for this main: it restores trust in the static `expression.Function` instead of proving safety for all `CallTargets`. Closure-target safety and MakeError borrowing remain intact. Its original patch and observations remain under `history` in the row, with the patch in `history/07-borrowed-element-reads-e011f8f.patch`.

`--entry NN` filters before creating the test loop. The entry-10 CLI probe printed only that skip and its wall time. The earlier sequential checker measured patch checks, its clean control, patched test and restoration; worktree/submodule setup was outside that per-entry timer. The parallel checker now includes setup and cleanup and prints their separate durations. Skips also print wall time. The checker prints commit/toolchain/build context and load before/after for these invocation observations. They are not before/after performance benchmarks. Go JSON output is decoded for diagnostic matching while its raw log is retained.

| Loop | Before (control), best of 3 | After (reverse patch), best of 3 | Instrument |
|---|---:|---:|---|
| 11 | 1.766s | 1.793s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/lower -run '^TestDefiniteAssignmentIsRefused$'` |
| 12 | 2.329s | 2.350s | `ADAMIC_GATE_UNCACHED=1 go test -count=1 ./internal/lower -run '^TestSuppressionDirectivesAreRefused$'` |

11 build flags: commit `e8ba3d5d81de4d3773c723914fccd4c76248b965`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; Node `v24.19.0`; `go test -count=1`; GOFLAGS ``; `ADAMIC_GATE_UNCACHED=1`. Clean load before/after `5.84 4.08 4.05 1/219 122610` / `5.45 4.03 4.03 1/219 122662`; mutant load before/after `5.45 4.03 4.03 1/219 122663` / `5.45 4.03 4.03 1/219 122713`. These lowerer tests do not build native binaries.

12 build flags: commit `e8ba3d5d81de4d3773c723914fccd4c76248b965`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; Node `v24.19.0`; `go test -count=1`; GOFLAGS ``; `ADAMIC_GATE_UNCACHED=1`. Clean load before/after `4.46 3.88 3.98 1/219 122945` / `4.51 3.90 3.99 1/219 122998`; mutant load before/after `4.55 3.92 3.99 1/219 123109` / `4.55 3.92 3.99 1/219 123160`. These lowerer tests do not build native binaries.

Named checker probes used the actual definite-assignment mutation. JSON output was accepted with all four expected failures. Injecting an unrelated failing test, requiring a missing subtest, or changing the expected diagnostic was rejected. Unanchored patterns and `-count=2` were rejected before running Go. The last probe checks the repeat count; specifying `-count=2` itself does not enable a cache.

```text
11 refuse-definite-assignment: applies-and-fails-as-recorded
Actual JSON Go test command: accepted with all four expected failing subtests
11 refuse-definite-assignment: applies-but-failure-not-as-recorded (needs investigation)
Unrelated failing test output mutant: rejected only by unexpected-test check
11 refuse-definite-assignment: applies-but-failure-not-as-recorded (needs investigation)
Missing required failing subtest mutant: rejected
11 refuse-definite-assignment: applies-but-failure-not-as-recorded (needs investigation)
Wrong named-test diagnostic mutant: rejected
Unanchored pattern mutant: rejected
Cached Go command mutant: rejected
Named checker probes passed; worktree restored
Label clarification: the "Cached Go command" probe changed -count=1 to -count=2. This is an incorrect repeat count, not an enabled cache. The repeat-count guard rejected it.
```

Resumed toolchain setup output:

```text
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
v24.19.0
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (45s)
setup: done in 45s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

The new validation scope is the two named commands in three clean/mutant rounds, the complete lowerer package clean and under each refusal patch, actual JSON and negative checker probes, and CLI entry selection. No new oracle sweep or full repository gate was run for this extension. After pushing this branch, the checker is run against the latest fetched origin/main and its output is reported in the handoff.

Refreshed entry 07 evidence on e8ba3d5d81de4d3773c723914fccd4c76248b965:

The original pre-fix fault trusted a virtual call's static signature even when an override removed elements. The refreshed patch retains that intent through main's newer CallTargets API. All three clean runs passed and all three patched runs failed with:

```text
oracle_test.go:618: exit codes differ
```

The refreshed full uncached oracle sweep detects two native comparison failures: borrow_element_virtual_store.a and call_targets_element.a. Including changed counted-build rows adds borrow_element_virtual_move.a, for three detecting fixtures. Entry 07 is no longer detected only by its primary native fixture. Original whole-package counts above remain explicitly historical. Raw new evidence is in logs/refresh-07-*.

The checker adds `-trimpath` to recorded Go commands to remove temporary checkout paths from compiled artifacts. Go's existing build cache retains its source, dependency, toolchain and build-flags keys; no custom cache or test-result cache was added. `ADAMIC_GATE_UNCACHED=1` and `-count=1` apply to both controls and mutants. Effective commands and individual invocation loads/durations are saved beside every raw log in .timing.json files. The plain-Go sequential run and the trimpath parallel run both require the same complete diagnostics and expected test failures, which checks observed output equivalence on every active entry.

Parallel checker timing on the same box and same main commit, before then after:

| Loop | Before, sequential | After, 5 workers | Instrument |
|---|---:|---:|---|
| 1, whole catalog including setup/cleanup | 333.646s | 146.446s | `python3 /workspace/scratch/check-before.py e8ba3d5d81de4d3773c723914fccd4c76248b965` then `bash verify/catalog/check.sh e8ba3d5d81de4d3773c723914fccd4c76248b965`; Python monotonic wall timer |

The sequential run exceeded five minutes, so this uses the requested single-run exception rather than best of three. The old checker and portable measurement instrument are preserved in history/check-sequential.py and history/measure-parallel.py; the original instrument executed the equivalent scratch copy. This comparison measures both parallel execution and -trimpath artifact reuse; it does not isolate the scheduler speedup. Both runs use the existing Go artifact cache: the sequential checkout path requires rebuilding dependencies, while trimpath can reuse artifacts warmed during entry 07 validation. The first trimpath build cost is retained in that entry's first control record, rather than hidden. Raw commands, timing records, controls and mutants are preserved in parallel-measurements.json and logs/parallel-*-1*.

before build flags: commit `e8ba3d5d81de4d3773c723914fccd4c76248b965`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; Node `v24.19.0`; GOFLAGS ``; `ADAMIC_GATE_UNCACHED=1`; native C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`. go test -count=1; first entry control also -x for build profiling. Load before/after `1.31 3.65 3.59 1/263 143061` / `2.68 3.61 3.68 1/269 146650`.

after build flags: commit `e8ba3d5d81de4d3773c723914fccd4c76248b965`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; Node `v24.19.0`; GOFLAGS ``; `ADAMIC_GATE_UNCACHED=1`; native C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`. go test -trimpath -count=1; jobs default nproc; one worktree per active entry. Load before/after `2.68 3.61 3.68 1/269 146650` / `6.13 4.93 4.18 1/326 150727`.

Entry 01's first control took 167.846s, but Go reported only 1.050s executing the test. Its -x trace recorded 74 compile commands, including cohere's TypeScript checker. Compilation/linking/startup dominates this first invocation; fixture execution accounts for under one percent of it. This explains the earlier 172s observation. Build flags are the before flags above, with the profiled invocation's load before/after `1.18 3.50 3.54 1/268 143208` / `4.95 4.54 3.95 1/257 144194`. The trace inventory and raw trace location are in logs/parallel-entry01-build-profile.json.

| Loop | Before, control best of 3 | After, entry 07 reverse patch best of 3 | Instrument |
|---|---:|---:|---|
| Refreshed 07 | 2.225s | 2.330s | `ADAMIC_GATE_UNCACHED=1 go test -trimpath -count=1 ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^borrow_element_virtual_store[.]a$'` |

Refreshed 07 build flags: commit `e8ba3d5d81de4d3773c723914fccd4c76248b965`; nproc `5`; cgroup cpu.max `400000 100000`; `go version go1.27.1 linux/amd64`; `clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)`; Node `v24.19.0`; GOFLAGS ``; `ADAMIC_GATE_UNCACHED=1`; native C11 strict warnings, `-ffp-contract=off -fno-optimize-sibling-calls`, release `-O2`, sanitizer `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`. `go test -trimpath -count=1`. Clean load before/after `4.60 4.53 3.71 1/247 132095` / `4.60 4.53 3.71 1/247 132178`; mutant `4.60 4.53 3.71 1/247 132179` / `4.31 4.47 3.69 1/249 132261`. All three interleaved pairs, including the initial artifact build, are retained in catalog.json.

Latest required setup output (raw log: logs/setup-parallel.log):

```text
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
v24.19.0
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (66s)
setup: done in 66s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

This setup is preparation rather than a timing comparison. nproc was 5. No full repository gate was run. The refreshed 07 whole oracle sweep, every active named control/mutant command, and the checker restoration probe provide this unit's validation.

The restoration probe runs the real entry 07 clean control and fault, then suppresses only `git apply -R`. The checker detects the dirty worktree and marks the entry failed despite a correctly detected compiler fault. It still removes its own worktree. Measured manifests and output are also audited for numeric ordering and unique worktree paths; reversed-order and shared-path data mutants fail those audits. Run `PYTHONDONTWRITEBYTECODE=1 python3 verify/catalog/probe-parallel.py origin/main` after the benchmark to reproduce these checks.

Final check after another fetch of origin/main, exit 0:

```text
target e8ba3d5d81de4d3773c723914fccd4c76248b965
logs /tmp/adamic-gate/adamic-catalog-logs-hfl1bopf
build flags: commit=e8ba3d5d81de4d3773c723914fccd4c76248b965; nproc=5; cpu.max=400000 100000; go version go1.27.1 linux/amd64; clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261); v24.19.0; ADAMIC_GATE_UNCACHED=1; go test -trimpath -count=1; native C11 strict warnings, -ffp-contract=off -fno-optimize-sibling-calls; release -O2; sanitizer -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all
jobs 5; Go -trimpath shares only content-keyed build artifacts; test results uncached
01 shared-slice-append: applies-and-fails-as-recorded
01 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
01 wall=31.075s setup=17.153s check=10.651s cleanup=1.091s load=2.63 4.15 3.95 2/335 151106 -> 3.10 4.10 3.94 3/375 152890
02 liveness-throw: applies-and-fails-as-recorded
02 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
02 wall=30.466s setup=16.206s check=11.947s cleanup=0.546s load=2.63 4.15 3.95 1/335 151106 -> 3.11 4.12 3.95 3/377 152868
03 defined-lent: applies-and-fails-as-recorded
03 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
03 wall=29.472s setup=17.850s check=9.974s cleanup=0.496s load=2.63 4.15 3.95 3/338 151109 -> 3.11 4.12 3.95 4/445 152650
04 borrowed-array-move: applies-and-fails-as-recorded
04 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
04 wall=32.187s setup=17.640s check=9.994s cleanup=1.603s load=2.63 4.15 3.95 3/338 151109 -> 3.10 4.10 3.94 1/379 152944
05 spread-method-reuse: applies-and-fails-as-recorded
05 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
05 wall=31.585s setup=17.474s check=10.497s cleanup=1.308s load=2.63 4.15 3.95 2/339 151110 -> 3.10 4.10 3.94 2/371 152899
06 constructor-capture-region: applies-and-fails-as-recorded
06 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
06 wall=29.419s setup=15.201s check=11.294s cleanup=0.574s load=3.11 4.12 3.95 6/445 152650 -> 4.09 4.26 4.00 5/438 154332
07 borrowed-element-reads: applies-and-fails-as-recorded
07 made against=e8ba3d5d81de4d3773c723914fccd4c76248b965
07 wall=31.462s setup=19.819s check=9.786s cleanup=0.605s load=3.11 4.12 3.95 3/378 152869 -> 4.09 4.25 4.00 1/396 154582
08 narrowed-number-field: applies-and-fails-as-recorded
08 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
08 wall=30.441s setup=18.412s check=9.707s cleanup=0.422s load=3.10 4.10 3.94 4/375 152890 -> 4.09 4.25 4.00 1/396 154581
09 literal-undefined-field: applies-and-fails-as-recorded
09 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
09 wall=27.842s setup=14.799s check=11.300s cleanup=0.930s load=3.10 4.10 3.94 2/371 152899 -> 4.09 4.26 4.00 5/433 154333
10 override-representation: skipped (Fix 1f96441 is still not an ancestor of origin/main (e8ba3d5d81de4d3773c723914fccd4c76248b965); leave skipped pending integration.)
10 made against=e8ba3d5d81de4d3773c723914fccd4c76248b965
10 wall=0.000s setup=0.000s check=0.000s cleanup=0.000s load=3.10 4.10 3.94 2/380 152945 -> 3.10 4.10 3.94 2/380 152945
11 refuse-definite-assignment: applies-and-fails-as-recorded
11 made against=e8ba3d5d81de4d3773c723914fccd4c76248b965
11 wall=28.349s setup=18.021s check=7.289s cleanup=0.564s load=3.10 4.10 3.94 2/380 152945 -> 4.09 4.25 4.00 3/439 154480
12 refuse-suppression-directives: applies-and-fails-as-recorded
12 made against=e8ba3d5d81de4d3773c723914fccd4c76248b965
12 wall=20.352s setup=11.937s check=6.876s cleanup=0.398s load=4.09 4.26 4.00 4/438 154332 -> 3.71 4.16 3.97 1/394 154817
13 typeof-null: skipped (Fix and its oracle fixtures are not on target main. Main already has the pre-fix typeOf implementation; there is no fix to undo on this target.)
13 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
13 wall=0.000s setup=0.000s check=0.000s cleanup=0.000s load=4.09 4.26 4.00 5/433 154333 -> 4.09 4.26 4.00 5/433 154333
14 constructor-arrow-this: skipped (Fix is not on target main and adds only lowerer refusal tests, no oracle fixture. Main already lacks the arrow capture refusal.)
14 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
14 wall=0.000s setup=0.000s check=0.000s cleanup=0.000s load=4.09 4.26 4.00 5/433 154333 -> 4.09 4.26 4.00 5/433 154333
15 class-instance-key: skipped (Fix and its oracle fixtures are not on target main. Main already uses representation-based instantiation keys, so cannot undo the absent fix.)
15 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
15 wall=0.000s setup=0.000s check=0.000s cleanup=0.000s load=4.09 4.26 4.00 4/433 154333 -> 4.09 4.26 4.00 5/434 154334
16 proven-guards-relations: skipped (Supplied SHA changes only whitespace in a log. Nearby cbcc765c adds coverage, no compiler fix; mutable_kind_guard.a is an unresolved disagreement kept under notes, absent from main and from the registered oracle fixtures.)
16 made against=e011f8f60899586d6373a5ccb07335ad82cfbf3c
16 wall=0.000s setup=0.000s check=0.000s cleanup=0.000s load=4.09 4.26 4.00 5/434 154334 -> 4.09 4.26 4.00 5/434 154334
total wall=79.306s load=2.63 4.15 3.95 1/333 151094 -> 3.71 4.16 3.97 1/389 154817
```

The explicit `--entry 07 -jobs 1` CLI check also exited 0 and selected only entry 07; output and flags are in logs/parallel-entry07-cli.log. All disposable worktrees were removed.
