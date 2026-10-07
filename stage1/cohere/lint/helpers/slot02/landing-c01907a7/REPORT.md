Rebased all twenty-four completed helpers onto current origin/main c01907a7; no new helper claimed.
Commits: previous published tip 21f00cd0; rebased source 445ed4c04d0f55c24e2d5b27a3b25551a66376f5; branch codex/lint-helpers-02.
Commands: setup 65s/nproc 5; complete helper oracle PASS 493.875s; inventory PASS 6.536s; uncached input oracle PASS 2.207s; vet exit 0.
Mutants: all seventy-seven existing compiling semantic mutants caught again against actual Go; every name and independent mismatch is in evidence/mutant-witnesses.log.
Not covered: full repository gate, whole-rule findings/fixes/suggestions, arbitrary malformed adapters and prior external Tailwind gaps; no new helper or rule implementation.

# Landing result

Main advanced from f8013f0 to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 after the previous publication. Under the requested work-in-progress cap, rebasing and re-greening the existing branch is this unit. The only branch this worker has published is codex/lint-helpers-02. All forty-five commits rebased without conflicts. No helper source, fixture, test, ownership claim or readiness ledger changed. evidence/ancestry.log records the old and rebased tips and exact helper-tree comparison. evidence/helper-source-manifest.json records unchanged SHA-256 values for all twenty-four owned helper implementations.

The main update contains Stage 3 work and its external oracle hook. No stage1/cohere file changed between these main revisions; the announced allocator-aware shared leak-check replacement has not arrived in this base. Nothing was reverted or excluded. No batch 8 Diagnostic landing SHA has been named in the task. No new rule or regexp matcher is introduced, so no rule.json kinds registration or new pattern translation is needed.

The final wildcard fetch refreshed main and all origin codex/lint-helpers* branches. Main remains c01907a7 and is an ancestor of the verified rebased source. Publication targets only the owned branch, with an exact force-with-lease against 21f00cd0cf2af24438f2afccf2487c7a0f3e3037 because the requested rebase rewrites its history. This protects any intervening remote update. No main or area branch is pushed and no pull request is opened. No new claim is made.

# Fresh checks on the new main

All test output goes directly to logs, never through a pipe. Source /workspace/adamic-tools/env.sh before Go commands. Go 1.27.1, clang 20.1.8 and Node 24.19.0.

- bash cloud/setup.sh > evidence/setup.log 2>&1: PASS. Go, clang, Node and submodules ready 0s; build cache warm 65s; done 65s on five processors. nproc prints 5.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/helpers-final.log 2>&1: PASS 493.875s, all seventy-seven semantic mutants caught.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 6.536s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 2.207s, all six fixtures, zero cache hits and six probe misses.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1: exit zero, empty log.
- git diff --exit-code 21f00cd0 HEAD -- stage1/cohere/lint/helpers: exit zero before adding this landing report/evidence. All prior helper files are unchanged.
- git merge-base --is-ancestor origin/main HEAD and git diff origin/main --check: exit zero.

Actual Go cohere decides helper answers. Source Node and ASan/UBSan native agree with those answers. Batch2 through batch8 additionally compare emitted JavaScript; no extra emitted-JavaScript coverage is claimed for the inherited option/policy tests or original three slot02 tests. Every consuming rule's retained actual runtime source/options captures and the latest Unicode/query controls run again. All baseline and mutant tests pass.

# Every mutant and its independent witness

[evidence/mutant-witnesses.log](evidence/mutant-witnesses.log) lists every executed mutant name and its first actual Go mismatch. The seventy-seven names are checked against the last publication's full witness list. Every semantic mutant compiles and runs successfully before comparison; refusals, crashes, sanitizer findings or unexpected stderr are not credited. Temporary copies contain mutations; production source stays unchanged.

Detailed mutation definitions and coverage remain in these reports:

- Original slot02 REPORT.md: three name/cache mutants; inherited shared helper REPORT.md: four options/policy mutants.
- Batch2 REPORT.md: four literal-list/splitting/namespace mutants.
- Batch3 REPORT.md: five factory/math mutants.
- Batch4 REPORT.md: five prefix/namespace/underscore mutants.
- Batch5 REPORT.md: fourteen utility/root/sort value/header/alias mutants.
- Batch6 REPORT.md: eighteen stylesheet/error/whitespace mutants.
- Batch7 REPORT.md: eleven import-source/folding/attribute-matching mutants.
- Batch8 REPORT.md: thirteen class escape/identity escape/literal serialization mutants.

Observed: every retained helper baseline and all seventy-seven semantic mutants pass on current main, and all twenty-four owned helper source files retain their hashes. Inferred: this branch meets the landing cap for this main revision. No additional rule readiness or new ownership is inferred from a rebase.

# Limits

This unit runs the complete touched helper package, inventory, vet and a filtered uncached external Node oracle rather than the full repository gate. Whole-rule findings/fixes/suggestions, common AST adapter implementation, arbitrary malformed input and previously documented unavailable Tailwind live/corpus gates remain outside coverage. Earlier claim and implementation SHAs in prior reports are historical pre-rebase identifiers; their evidence is preserved unchanged. The latest regexp leaves serialize escapes and do not implement a matching engine. No new helper code is built this landing unit.
