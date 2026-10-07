Rebased all 32 retained helpers onto current main c01907a7; no new helper or rule claimed.
SHAs: previous pushed head 9c1a8401; full-gate tested rebased head dc5f70ad; final rebased head 522cb3aab568143d015e69b25fb0743bf1b8746c; report SHA is named in the final response.
Commands and outputs: all eleven helper packages PASS, vet/format PASS, filtered uncached oracle PASS 1.195s then 0.881s; setup 33s, nproc 5.
Mutants: all 89 compiling semantic variants caught by Go comparisons; eight batch11 variants caught again after the final documentation-only rebase; every witness is archived.
Not covered: full repository test gate, whole-rule integration or new helpers; no compiler, shared harness or registration edits.

# Landing on current main

The work-in-progress cap made landing readiness this unit's task. codex/lint-helpers-05 is the only branch owned and pushed by this unit. All previously retained helpers were completed and pushed at 9c1a8401c186156e241b48e3a045bea95937bcc1. Main advanced from f8013f0baac41ddc340d76f83bddde38536a8f07 to a62e1f9de6c91080ddabfa375479fb437a46ed0c, adding Stage 3 work and internal/oracle/stage3_hook_test.go. The announced shared allocator leak-check changes are not on this main. No upstream diff was reverted.

The first rebase replayed all 59 branch commits cleanly. Every patch compared equal in git range-diff. Full helper validation ran at dc5f70ad0375e46861fe43c63f3d2ad1e0e0e63c. During that gate main advanced once more to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06, changing only one line in documentation/velocity/landings.csv. A second clean rebase preserved all 59 patches. The final rebased head is 522cb3aab568143d015e69b25fb0743bf1b8746c.

landing-evidence-3/oracle-input-identity.json compares Git object IDs for internal, cmd, cohere, go.mod, oracle, the entire helper tree and cloud before and after the final rebase. Every input is identical; go.sum is absent on both sides. final-delta.log names only the velocity CSV. Thus the complete eleven-package results apply to unchanged compiler and helper inputs. The filtered uncached oracle and batch11 were also rerun after the final rebase. Their final results are PASS 0.881s and PASS 39.137s respectively, including all eight batch11 mutants again. These additional executions are not counted as eight new distinct mutants.

# Toolchain and commands

Setup succeeded with Go 1.27.1, clang 20.1.8 and Node 24.19.0. nproc printed 5. Exact timing:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (33s)
setup: done in 33s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

All test output went directly to logs. With /workspace/adamic-tools/env.sh sourced:

```
bash cloud/setup.sh > /tmp/lint05-batch12-setup.log 2>&1
git rebase origin/main > /tmp/lint05-landing3-rebase.log 2>&1
git range-diff f8013f0..9c1a840 origin/main..HEAD > /tmp/lint05-landing3-range-diff.log
ADAMIC_GATE_UNCACHED=1 go test -p=2 ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot05/... -count=1 -v -timeout=30m > /tmp/lint05-landing3-helpers.log 2>&1
go vet ./... > /tmp/lint05-landing3-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-landing3-oracle.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05 > /tmp/lint05-landing3-format.log 2>&1
git rebase origin/main > /tmp/lint05-landing3-final-rebase.log 2>&1
git range-diff f8013f0..9c1a840 origin/main..HEAD > /tmp/lint05-landing3-final-range-diff.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-landing3-final-oracle.log 2>&1
go test ./stage1/cohere/lint/helpers/slot05/batch11 -count=1 -v -timeout=20m > /tmp/lint05-landing3-final-helper.log 2>&1
```

All commands exited zero. Repository-wide vet and owned formatting produced empty logs. Both filtered oracle runs passed all six input fixtures, with zero probe cache hits and six misses. The complete helper gate covers all 32 owned helpers plus four inherited prerequisites. Earlier packages compare actual Go, source Node and sanitized native; batches 4 through 11 also compare emitted JavaScript. Native uses ASan/UBSan and default Linux leak checking. Package concurrency was capped at two; tests inside each package retain their declared serial memory bounds.

| Package | Observed time |
| --- | ---: |
| Shared and original slot 05 | 139.190s |
| Shared and original slot 05/slot05/batch10 | 126.645s |
| Shared and original slot 05/slot05/batch11 | 41.013s |
| Shared and original slot 05/slot05/batch2 | 110.674s |
| Shared and original slot 05/slot05/batch3 | 17.839s |
| Shared and original slot 05/slot05/batch4 | 46.273s |
| Shared and original slot 05/slot05/batch5 | 70.344s |
| Shared and original slot 05/slot05/batch6 | 155.248s |
| Shared and original slot 05/slot05/batch7 | 23.209s |
| Shared and original slot 05/slot05/batch8 | 210.408s |
| Shared and original slot 05/slot05/batch9 | 51.378s |

# Mutants, readiness and limits

All 89 distinct semantic variants compiled, ran with exit zero and empty stderr, and differed from actual Go output or dependency traces. Compile failures, panic and sanitizer errors are not credited. landing-evidence-3/mutants.json records every test, exact mismatch and log line. Mutation definitions and contracts remain in the original helper test files and batch reports. Counts by package are shared/original 7; batch2 3; batch3 3; batch4 3; batch5 13; batch6 12; batch7 11; batch8 9; batch9 8; batch10 12; batch11 8. Four inherited variants and 85 owned variants total 89.

No new reservation was made while landing was incomplete. Historical claim timestamps and ownership resolutions remain unchanged, and withdrawn overlaps stay withdrawn. Readiness is unchanged: 32 retained helpers remove 223 prerequisite occurrences across 66 rules; the frozen conditional helper-readiness count remains fifty. No new listener, rule.json kinds, finding-position conversion, Diagnostic integration or regex matcher was added. The user has not named a Diagnostic landing SHA here.

This landing unit stops after pushing the freshly green rebased branch. Integration owns main and area branches; neither is pushed by this worker. No pull request is opened. The own-branch push uses the explicitly authorized rebase and an exact lease on previous head 9c1a8401c186156e241b48e3a045bea95937bcc1.

Not covered: full repository test gate, complete rule findings/spans/fixes/suggestions, whole-linter integration, dependency implementations beyond each existing adapter contract or arbitrary invalid inputs outside documented helpers. Every archived log is in landing-evidence-3. No protected compiler, shared harness or registration file is changed.
