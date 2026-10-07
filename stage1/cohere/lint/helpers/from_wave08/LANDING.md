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

## Refreshed main e8ba3d5d

Rebased the helper branch again onto current origin/main e8ba3d5d. Fresh uncached owned-package comparison PASS in 19.450s, all four baselines and all four semantic mutants on Node, emitted JavaScript and sanitized native. Whole-repository vet exit zero; filtered uncached external oracle PASS in 0.465s. Commands match those above; refresh logs are in evidence/landing_refresh_*.log. No helper source change or new claim. The rule rebase still conflicts in the same four shared foundation files on the first registration commit and was aborted again; its tip remains 6a66e2d3. Main and area branches were never pushed.

## Main f8013f0ba

Rebased the helper branch onto refreshed origin/main f8013f0baac41ddc340d76f83bddde38536a8f07. Fresh uncached four-helper Go/Node/emitted-JavaScript/sanitized-native comparison and all four semantic mutants PASS in 25.241s. Whole-repository vet exit zero, empty log; filtered uncached external oracle PASS in 7.649s with zero cache hits. Commands and scope match the earlier landing verification; logs are evidence/landing_f801_*.log. No helper source change or new claim. Retried the rule branch rebase against this main: the first registration foundation commit still conflicts in README.md, lint.ts, lint_test.go and testdata/oracle.go. Aborted that rebase and preserved 6a66e2d3. Shared harness/registration ownership keeps the rule branch blocked and the cap closed. Pushed only the owned helper branch, using a lease against its previous remote tip 1021f3f1. No main or area branch was pushed; no Diagnostic integration SHA was supplied.

## Current main c01907a70

Main advanced during the parking gate. Rebased this branch onto c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; nineteen commits replayed cleanly. Fresh uncached four-helper comparison and all four semantic mutants PASS in 19.617s on Go, source Node, emitted JavaScript and ASan/UBSan native. Whole-repository vet exit zero, empty log; filtered uncached external oracle PASS in 0.501s, zero native/Node cache hits. Commands match the f801 verification; logs are evidence/landing_c019_*.log. No helper source change or new claim.

The rule branch now excludes shared foundation commits and rebases its owned commits cleanly, rather than resolving shared conflicts. Scratch comparisons hold all eighteen ports over 886 supported upstream cases, the compiler/stage1 corpus and eighteen semantic mutants; its owned PARKING.md names #zmh9v36. The explicit TSX gap probe still fails with expected GreaterThanToken, got Identifier at 15. Other JSX/legacy-number exclusions remain. Thus full fixture parity has parser gaps beyond the shared harness, and this unit does not assert eligibility for the shared-harness-only parking exception. No new helper is claimed while that strict cap remains unresolved.

## Current main b8fb957aa

Rebased this branch from pushed tip 6982efe8d5993dd549c42a1651b22b919bdc75e8 onto b8fb957aa839a9e8cb0b54279dd9864fa317bd30. Rebase was clean; pre-evidence tip 55646fd1c59960c250fc842d38fb41a139053d58. Main changes native fieldSlot handling for inherited static fields and adds its oracle fixture. No owned helper source change, shared harness edit or new claim.

Fresh `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave08 -count=1 -v -timeout=15m` PASS in 17.689s. All four baselines and all four compiling semantic mutants are compared with Go on source Node, emitted JavaScript and ASan/UBSan native. NewTheme nonzero dead count, ClearNamespace hyphen boundary, missing var( width refusal and reversed unresolved comparison are caught only by output comparison; successful executions exit zero with empty stderr. The same six consuming rules and zero final blockers alone remain listed in REPORT.md and BREAKPOINTS_REPORT.md; no rule is marked ported by this rebase.

Whole-repository `go vet ./...` exited zero with empty log. Uncached one-byte oracle PASS in 0.541s. The inherited-static-field fixture was selected separately with `go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^inherited_static_field_read[.]a$' -count=1 -v -timeout=10m`: PASS in 0.781s, native misses 3 and Node misses 2, no cache hits.

`bash cloud/setup.sh` PASS: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 212s, total 212s. nproc prints 5; cgroup quota is four CPUs. Tools remain Go 1.27.1, clang 20.1.8 and Node 24.19.0. Evidence is landing_b8fb_*.log. No full repository test suite or new throughput measurement. The owned rule branch is also rebasing and recertifying; its existing recovered-key and legacy-number parser gaps still close the strict landing cap beyond the harness-only parking exception.

## Current main 39638d9e

Rebased cleanly from remote a055d8d1c6a8589edf7af8cac0420beff4b5a24b onto origin/main 39638d9e278d38bb5aeae887f46d55a70e47aaad. Implementation tip before this evidence commit: 33f5d769975e22d29c4706485ebc8d998103449e. No helper source change or new claim.

Fresh uncached owned-package comparison PASS in 26.323s on actual Go, source Node, emitted JavaScript and ASan/UBSan native. All four comparison-only semantic mutants pass: nonzero NewTheme deadKeys, ClearNamespace hyphen boundary, omitted var( refusal and reversed unresolved breakpoint ordering. Whole-repository vet exited zero with empty log. Filtered uncached TestTheOracleCatchesOneByte PASS in 0.677s, zero cache hits. Commands match the prior section; evidence is landing_396_*.log. Four helpers still support the same six Tailwind consumers listed in REPORT.md and BREAKPOINTS_REPORT.md; none is independently the final missing helper. No full repository test suite or fresh throughput benchmark.

The rule branch has now rebased onto area/stage1-lint 7481e0324e34a2537aafa9db7eeacda50405611b, which contains current main. It applies the dedup ledger and removes eleven losing copies. Its prefer-as-const two-edit fix is refused by the unified Go oracle (unexpected fix shape) and the shared finding model; malformed computed-key parser recovery is also blocked. Therefore the landing cap remains closed and no additional helper is claimed. This evidence supersedes the old shared-registration conflict, which no longer blocks rebase. No main or area branch is pushed.
