Built: landing refresh of all seven delivered helpers onto origin/main b8fb957aa839a9e8cb0b54279dd9864fa317bd30; no ninth helper was claimed.
Commits: pre-rebase pushed helper 2b917cbb24fa112ddf8537814e7790f91e4877a2; rebased implementation bcd3074e3c947519e99748f3becbaadde38fab72 before this evidence commit; rule sibling rebased 79ed6a17bc1a17f445cb4daad5c56471dbd73eb7 and being rechecked independently.
Checks: complete owned helper package PASS 131.391s, vet clean; all seven contracts retain actual Go/source Node/emitted JS/sanitized-native parity. Main's inherited-static-field filtered oracle also PASS 0.918s uncached.
Mutants: all 20 existing helper controls are caught again, including three literal-regex, six truthiness/label and the earlier bigint/predicate/allocator/segment/counter controls; semantic mutants compile/run before comparison catches them.
Uncovered: exact counter bigint-return lowering remains blocked; full CFG/rule integration and repository-wide gate are not claimed; other unclaimed helpers remain.

The final pre-claim fetch observed main advancing from c01907a7 to b8fb957a. The ancestry check aborted before the candidate decorator claim file could be written; no candidate helper code or reservation exists. Under the user's landing-first instruction the unit became rebasing and re-greening the two existing branches. This branch rebases cleanly, retains the inherited-static-field emitter fix and introduces no helper implementation changes. The complete owned package was nevertheless rerun uncached by count=1 and each helper oracle compares actual Go anew; no old result is relabeled as a new run.

The four CFG consumers remain array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. The five delivered CFG helpers remove five dependencies from each, zero final blockers. The two delivered Tailwind helpers still serve their six documented consumers. The counter support probe still agrees with Go on Node and refuses at gaps/atomic-counter.a:4:10: stage 0 can't lower a function returning bigint yet. No production counter, exact native counter, concurrent counter or full graph builder is delivered.

Main's new delta changes only internal/native/emit_objects.go and the inherited_static_field_read.a oracle fixture/count/registration. No shared lint harness or the announced leak-check helper changes have landed in this delta. The filtered upstream static-field oracle runs source Node, emitted JS and sanitized/release native and confirms the new fix remains intact. The exact same compiler delta is retained by both rebased branches. Shared lint helpers' comparison/context/main/generator files are not modified by this unit.

Commands after sourcing /workspace/adamic-tools/env.sh:

```sh
git rebase origin/main
go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-b8fb-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-b8fb-helper-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/inherited_static_field_read.a$' -count=1 -v -timeout=10m > /tmp/wave12-b8fb-static-oracle.log 2>&1
```

Every command writes test output to a log. The helper branch is pushed only to codex/lint-helpers-from-lint-wave1-12 with the exact pre-rebase tip as force-with-lease, per the user's explicit rebase/push instruction. Never to main or any area branch. Setup Go/clang/Node/submodules ready 0s, cache warm 53s, total 53s; nproc 5. No new throughput sample or complete gate claim.
