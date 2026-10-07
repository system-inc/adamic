Rebased all twenty-one retained helpers onto current origin/main f8013f0; no new helper claimed.
Commits: previous published tip 64a72a8; rebased source 088d78c56e70906956e54b3cbdfe0cfc7162a044; branch codex/lint-helpers-02.
Commands: setup 169s/nproc 5; complete helper oracle PASS 567.703s; inventory PASS 15.079s; uncached input oracle PASS 58.701s; vet exit 0.
Mutants: all sixty-four existing compiling semantic mutants caught again against actual Go; every mutant name and mismatch is in evidence/mutant-witnesses.log.
Not covered: full repository gate, integrated rule findings/fixes/suggestions, invalid adapters and prior documented external Tailwind gaps; no new helper implementation.

# Landing result

Main advanced from e8ba3d5 to f8013f0baac41ddc340d76f83bddde38536a8f07 after the previous publication. Under the requested work-in-progress cap, rebasing and re-greening the existing branch is this unit. The only branch this worker has published is codex/lint-helpers-02. Its forty commits rebased without conflicts; no helper source, test, fixture, dependency ledger or prior report changed. evidence/ancestry.log records the old and rebased tips and the exact helper-tree comparison. evidence/helper-source-manifest.json records unchanged SHA-256 values for every one of the twenty-one retained helper implementations.

The final wildcard fetch checked main and all origin codex/lint-helpers* branches. Current main remains f8013f0 and is an ancestor of the rebased source. Publication uses only the owned branch, with an exact force-with-lease against 64a72a8dd40b8293b6a9005ff5619f9f1fd60d56 because rebasing rewrites its history. This protects any intervening remote update. No main or area branch is pushed and no pull request is opened. No new claim is made.

# Fresh checks on the new main

All test output goes directly to logs, never through a pipe. The toolchain environment is /workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8 and Node 24.19.0.

- bash cloud/setup.sh > evidence/setup.log 2>&1: PASS. Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 1s, build cache warm 169s, done 169s on five processors. nproc prints 5.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/helpers-final.log 2>&1: PASS 567.703s, all sixty-four semantic mutants caught.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 15.079s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 58.701s, all six fixtures, zero cache hits and six probe misses.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1: exit zero, empty log.
- git diff --exit-code 64a72a8 HEAD -- stage1/cohere/lint/helpers: exit zero before adding this landing report/evidence. All prior helper files are unchanged.
- git merge-base --is-ancestor origin/main HEAD and git diff origin/main --check: exit zero.

Actual Go cohere decides helper answers. Source Node and ASan/UBSan native agree with those answers. Batch2 through batch7 additionally compare emitted JavaScript; no extra emitted-JavaScript coverage is claimed for the inherited option/policy tests or the original three slot02 tests. The latest JSX batch includes every consuming rule's explicit query names, full-scalar Unicode table/version checks and eleven compiling semantic mutants. The source fixtures and Go dependency contracts are the same as the last publication.

# Every mutant and its independent witness

[evidence/mutant-witnesses.log](evidence/mutant-witnesses.log) lists all sixty-four executed mutant test names and the first actual Go mismatch that caught each. Every semantic mutant compiles and runs successfully before comparison; a refusal, crash, sanitizer finding or unexpected stderr fails the test and is not credited. Temporary copies contain the mutations; production sources remain untouched.

The mutation definitions and detailed scope are retained in the first-batch and individual batch reports:

- [Original names/cache helpers](../REPORT.md): three mutants.
- Inherited helper REPORT.md at ../../REPORT.md: four options/policy mutants.
- [Batch2](../batch2/REPORT.md): four literal-list/splitting/namespace mutants.
- [Batch3](../batch3/REPORT.md): five factory/math mutants.
- [Batch4](../batch4/REPORT.md): five prefix/namespace/underscore mutants.
- [Batch5](../batch5/REPORT.md): fourteen lookup/root/sort value/header/alias mutants.
- [Batch6](../batch6/REPORT.md): eighteen stylesheet/error/whitespace mutants.
- [Batch7](../batch7/REPORT.md): eleven import-source/folding/attribute-matching mutants.

Observed: every retained helper's baseline and all sixty-four semantic mutants pass on the new main compiler/runtime; all twenty-one helper source files retain their previous hashes. Inferred: the branch meets the landing cap for current main. No additional readiness, rule implementation or new helper ownership is inferred from a rebase.

# Limits

This unit uses the complete touched helper package, the inventory package, vet and a filtered uncached external oracle rather than the full repository gate. Whole-rule findings/fixes/suggestions, common AST adapter implementation, arbitrary malformed input and the previously documented unavailable Tailwind live/corpus gates remain outside coverage. Earlier claim and implementation SHAs in existing reports are historical pre-rebase identifiers; their evidence is preserved unchanged. No new rule is introduced, so there is no rule.json kinds registration or shared Diagnostic integration change in this unit.
