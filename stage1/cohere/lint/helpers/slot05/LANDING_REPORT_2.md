Rebased the 26 retained helpers onto current main f8013f0; no new helper or rule claimed.
SHAs: previous pushed head 76d8c61, tested rebased head 11a96b5792cd59ec564f192d0dacda8b5d2d235b; final report commit is named in the final response.
Commands: all nine helper packages PASS, vet/format PASS, uncached input oracle PASS 11.303s; setup 107s, nproc 5.
Mutants: all 69 compiled semantic variants caught again by Go comparisons; every witness is in landing-evidence-2/mutants.json.
Not covered: full repository test gate, whole-rule integration or new helpers; no compiler, shared harness or registration edits.

# Landing on current main

Branch: codex/lint-helpers-05. This is the unit's only owned pushed branch. origin/main advanced from e8ba3d5d81de4d3773c723914fccd4c76248b965 to f8013f0baac41ddc340d76f83bddde38536a8f07. The user instructed landing first and explicitly authorized rebasing the owned branch and pushing it back to its own name. No main or area branch is pushed; no pull request is opened.

Rebase command: git rebase origin/main > /tmp/lint05-landing5-rebase.log 2>&1. Exit 0. Fifty commits replayed cleanly. git range-diff e8ba3d5..76d8c61 origin/main..HEAD reports all 50 patches equal, including the inventory, original helpers and all nine batches. Evidence: landing-evidence-2/rebase.log, range-diff.txt and rebase.json. No implementation edits were needed.

Main added lowering, inheritance and map-iterator/runtime changes, so every existing helper oracle was rerun. The final fetch confirmed main remained f8013f0baac41ddc340d76f83bddde38536a8f07, and git merge-base --is-ancestor origin/main HEAD returned 0. The remote own branch still points to 76d8c61e33eeb5cbbd6afba78fb50d2b8b7951e2, the expected lease value.

## Toolchain

bash cloud/setup.sh > /tmp/lint05-landing5-setup.log 2>&1: exit 0. source /workspace/adamic-tools/env.sh before Go commands. nproc printed 5. Exact setup timing:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (107s)
setup: done in 107s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0. Go cohere remains pinned to 715ba94f3608a6500086b1076ce5cb7e51b836db.

## Fresh validation

All tests wrote logs directly, without pipes. No cache result substitutes for the rerun.

```
ADAMIC_GATE_UNCACHED=1 go test -p 1 ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot05/... -count=1 -v -timeout=20m > /tmp/lint05-landing5-helpers.log 2>&1
```

Exit 0. All nine packages passed, covering all 26 retained owned helpers and four inherited shared helpers. The earlier packages compare actual Go, source Node and sanitized native; batches 4 through 9 also compare emitted JavaScript. Native builds use ASan/UBSan and default Linux leak checking. A successful run must exit zero with empty stderr. The package order is serial to bound memory.

| Package | Observed time | Mutants caught |
|---|---:|---:|
| Shared package plus original slot 05 helpers | 139.537s | 7 |
| Batch 2 | 107.270s | 3 |
| Batch 3 | 16.503s | 3 |
| Batch 4 | 45.399s | 3 |
| Batch 5 | 68.718s | 13 |
| Batch 6 | 152.559s | 12 |
| Batch 7 | 22.848s | 11 |
| Batch 8 | 207.473s | 9 |
| Batch 9 | 50.666s | 8 |

Every mutant is built from a temporary copy. Compile failure, nonzero exit, panic or stderr is not credited. Each of the 69 final mutations compiled and ran, then produced a semantic mismatch against real Go. landing-evidence-2/mutants.json lists every test and exact observed witness with its log line; helpers.log contains the complete run. Mutation definitions remain in the test sources, with contract explanations in each existing batch's REPORT.md. Original four inherited mutants plus 65 owned mutants make 69 total.

- go vet ./... > /tmp/lint05-landing5-vet.log 2>&1: exit 0, empty log.
- gofmt -l stage1/cohere/lint/helpers > /tmp/lint05-landing5-format.log: exit 0, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > /tmp/lint05-landing5-oracle.log 2>&1: PASS 11.303s, all six input fixtures, zero probe cache hits and six misses.

See landing-evidence-2/helpers.log, vet.log, format.log and oracle.log. The full repository test gate was not run; this unit reran every owned helper package plus inherited helpers, repository-wide vet and the filtered uncached Node oracle.

## Claims and limits

No claim was added while landing was incomplete. All 26 retained helpers remain ported, tested and pushed on this branch after the landing push. Earlier published claim SHAs and timestamps in the ownership records remain historical evidence; rebasing preserves patch content and author dates and does not create a new reservation. Earlier withdrawn overlaps remain withdrawn.

Readiness is unchanged: 199 dependency occurrences removed across 66 rules; four final helper blockers removed beyond the original 46, giving conditional helper readiness 50. The individual batch consumer ledgers still apply. No new rules, rule.json listeners or dispatch changes are introduced; the user has not supplied a shared Diagnostic landing SHA in this unit.

Not covered: full linter integration and complete findings/spans/fixes/suggestions; full repository test gate; arbitrary invalid adapter inputs and byte strings beyond the existing helper contracts. No protected compiler, shared generator or shared harness file is edited. This is the landing unit required by the work-in-progress cap, so it stops after publishing the freshly green rebased branch without claiming another batch.
