# Landing cap verification

No new helper or rule was claimed in this unit. Fetched every origin head and checked both pushed branches. Neither old tip was an ancestor of main, and neither contained current main. Target main is `e011f8f60899586d6373a5ccb07335ad82cfbf3c`.

## Helper branch

`codex/lint-helpers-from-lint-wave1-08` rebased cleanly from pushed tip `7de34d523d08c8f52379ec9b848be2df5e32d9f5` onto target main. Sixteen commits were replayed, including inherited helper/inventory foundations. Rebased implementation tip before this evidence commit: `3eced48119c9eb76e1abf104363685c1471f3e70`. A local recovery ref preserves the old tip. No helper source change was required. The explicit user instruction to rebase and push authorizes the history update; the push uses a full expected-old-head lease, preventing overwrite of any intervening remote work.

Fresh uncached validation after rebase:

```
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/wave08-landing-setup.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave08 -count=1 -v -timeout=15m > /tmp/wave08-helper-landing-tests.log 2>&1
go vet ./... > /tmp/wave08-helper-landing-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave08-helper-landing-oracle.log 2>&1
```

Owned helper package PASS in 25.099s. All four helper baselines matched actual Go on source Node, emitted JavaScript and native under ASan/UBSan; all four compiling semantic mutants were caught on every port backend: nonzero initial deadKeys, hyphen-boundary namespace clearing, missing var( width refusal and reversed one-unresolved comparison. Every successful baseline/mutant exits zero with empty stderr; only output comparisons kill mutants. Whole-repository vet exited zero with empty log. Filtered external oracle PASS in 7.269s, zero cache hits and one native/Node miss each.

Setup PASS: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 171s, total 171s on five processors. Tools: Go 1.27.1, clang 20.1.8, Node 24.19.0. No full repository test gate or inherited helper suite was run; scope is the four owned helpers, repository vet and filtered external oracle. Consumer and integration limits remain in the earlier reports.

## Rule branch blocker

`codex/lint-wave1-08` remains at its pushed tip `6a66e2d3e42854d4385801d032f17cff314f55b7`. An isolated worktree attempted rebase onto target main. The first replayed foundation commit, `29175443` (Discover lint rules from independent directories), conflicted in:

- stage1/cohere/lint/README.md
- stage1/cohere/lint/lint.ts
- stage1/cohere/lint/lint_test.go
- stage1/cohere/lint/testdata/oracle.go

Main has newer static rule dispatch and oracle behavior; the registration foundation cannot be applied as an automatic patch over those files. Ahra's standing territory instruction says to keep changes in owned rule directories and not edit shared registration or harness files, and to state blockers and stop. These are shared foundation conflicts, so no out-of-territory resolution was made. The rebase was aborted, leaving all old commits and the pushed branch intact. Its pre-rebase tests remain historical evidence; it has not been re-certified against current main and is not landing-ready. The shared foundation owner must reconcile these files before this rule branch can pass the landing cap.

The rule blocker keeps the cap closed. No new claim, helper implementation or pull request was opened.
