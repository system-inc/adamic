Built: three retained helpers unchanged; both owned branches rebased onto origin/main b8fb957a; no new claims.
Commits: rule branch pushed at 4a97dd6de; helper evidence commit follows this current-main rebase.
Commands: owned helper tests and vet PASS again on b8fb957a; 4,470 cases and 820,676 matching Go bytes; six fresh rule witness/mutant suites and uncached inherited-field oracle also PASS.
Mutants: twelve helper and seventeen rule semantic mutants compile, finish cleanly and are caught on source Node, emitted JavaScript and sanitized native again.
Not covered: full gate, fresh broad rule corpus/performance on the new backend, normal harness integration and eleven frontend/capture exclusions; no new helper claim.

## Landing state

Fetched origin/main and inspected both branches previously pushed by this unit. Neither was an ancestor of main, and neither contained the current main. The helper branch rebased cleanly with sixteen commits replayed onto e8ba3d5d81de4d3773c723914fccd4c76248b965. A second fetch before publication found the same main SHA. No helper source changed. The current compiler reran the owned direct-Go comparisons and all mutants successfully. The branch is published under its existing name, codex/lint-helpers-from-lint-wave1-13. The explicitly requested rebase rewrites that owned branch; publication uses a lease pinned to its observed old head 5ee3b523a1ab4a5e77817156db47db59654fa2bc to reject any concurrent remote change. No main or area/ ref is written.

Tests:

```
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/lint/helpers/from-wave1-13/at-rule ./stage1/cohere/lint/helpers/from-wave1-13/modifier ./stage1/cohere/lint/helpers/from-wave1-13/resolver -count=1 -v -timeout=10m > /tmp/lint-wave13-landing-helper-tests.log 2>&1
go vet ./stage1/cohere/lint/helpers/from-wave1-13/at-rule ./stage1/cohere/lint/helpers/from-wave1-13/modifier ./stage1/cohere/lint/helpers/from-wave1-13/resolver > /tmp/lint-wave13-landing-helper-vet.log 2>&1
```

Baseline observations still match pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. The three helper ports remove eighteen recorded dependency occurrences across the same six Tailwind rules, zero final blockers. Their direct helper observation contracts and explicit dependencies remain as documented in their individual reports. No whole-rule parity claim is added.

## Rule rebase blocker

Created an isolated worktree for codex/lint-wave1-13 and attempted git rebase origin/main. The first replayed commit, 29175443 (Discover lint rules from independent directories), conflicts in:

- stage1/cohere/lint/README.md
- stage1/cohere/lint/lint.ts
- stage1/cohere/lint/lint_test.go
- stage1/cohere/lint/testdata/oracle.go

These are shared dispatcher, harness and oracle files inherited from the registration foundation, outside the unit's rule directories. Ahra's ownership instruction says to keep changes inside owned rule directories and not edit the shared registration generator or test harness; it says to identify other blockers and stop rather than editing shared files. Resolving this foundation conflict would require choosing or combining shared implementations. The rebase was therefore aborted without resolving or editing any shared file, and the original branch remains at 89d25f6f888c7a737c876d6a1b51a4907162bbd8. It has not been re-greened against current main and is not declared landing-ready. Integration must resolve or land the registration foundation before this unit can replay its owned rule work without that conflict.

The work-in-progress cap remains active because of that rule branch. No helper eligibility scan, new claim, new implementation or new branch follows this blocker. The helper verification is independent of the aborted rule rebase. Test and rebase output is saved under evidence/landing-*.log. No PR is opened.

## Current main follow-up

Rebased cleanly onto origin/main f8013f0baac41ddc340d76f83bddde38536a8f07. The same three owned helper packages passed uncached go test -count=1 -v -timeout=10m and go vet. All 4,470 baseline cases and twelve semantic mutants passed their comparisons on source Node, emitted JavaScript and sanitized native. The rule rebase was retried against this main and encountered the same four shared-file conflicts at registration commit 291754439; aborted without shared edits. No new claims, numeric listener changes or full gate followed the blocker. Current output is retained in evidence/current-helper-tests.log, current-helper-vet.log and current-rule-rebase.log.

## Current parking assessment on c01907a7

The new parking instruction allowed replaying only the 29 owned rule commits and retaining current main shared files. That branch is now rebased and pushed at edb7991340b30ba4dc260d6a864cdcc36d4d074e. Its supported corpus and all seventeen compiling mutants passed; fresh current-main witness/mutant runs also passed after main advanced to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Compiler and stage1 inputs did not change between f8013f0b and c01907a7; the new Stage 3 oracle hook was not run. The owned rule PARKING_REPORT.md retains details and throughput. Shared registration remains absent on main, and eleven parser/adapter exclusions prevent a claim of full Go parity beyond that harness gap. No new helper is claimed under the only-harness condition.

The helper branch also rebased cleanly to c01907a7. All three owned helper tests were rerun with -count=1 -v -timeout=10m, all twelve semantic mutants caught on all three modes, and owned helper go vet passed with empty output. New logs are evidence/c019-helper-tests.log and evidence/c019-helper-vet.log. All historical sections above describe earlier observed states.

## Corrected listener contract

Rule descriptors use validated typescript-go ast.Kind names, not numeric kinds. The owned rule parking report was corrected and pushed at b6ae2ce0f294d1595ef97dda8e07ddcd484f7973. No code or test input changed, so the current-main oracle evidence remains applicable. The named shared harness ab70f38d4 includes JSX parser work but is not on main c01907a7; its effects on the remaining exclusions are unmeasured. No new helper or rule is claimed. New regex ports use the shared translated JS RegExp table or literals, with Unicode-mode constructors for option patterns.

## Latest current main revalidation

Main advanced to b8fb957aa839a9e8cb0b54279dd9864fa317bd30 with the inherited static field read backend fix. Both owned branches rebased cleanly and retained that upstream change. All three owned helper tests passed again with -count=1 -v -timeout=10m, including all twelve compiling semantic mutants; owned helper vet passed with empty output. Logs: evidence/b8-helper-tests.log and evidence/b8-helper-vet.log. The rule branch was freshly revalidated with six raw-witness/mutant suites and overlay vet, then pushed at 4a97dd6de22d0932c9c290ccc9f36d4f141aa8e1. Its uncached inherited-field Node oracle also passed. Broad rule corpus and throughput observations above remain historical for the preceding backend. The named shared harness ab70f38d4 is not on main. No new claim follows the eleven frontend/capture exclusions, which extend beyond the only-harness parking condition on this tested main.
