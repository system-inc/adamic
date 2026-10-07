Built: no new helpers; helper branch rebased onto current origin/main e8ba3d5d and revalidated; rule rebase attempted and blocked by shared-file conflicts.
Commits: helper rebase head ee606e1f plus this report commit; preserved rule head 89d25f6f; helper origin before rebase 5ee3b523.
Commands: three owned helper go tests PASS 30.965s, 56.047s and 31.473s; owned go vet PASS, empty output; setup 128s, nproc 5.
Mutants: all twelve retained semantic mutants compile, finish cleanly and differ from Go on source Node, emitted JavaScript and sanitized native; 4,470 baseline cases compare 820,676 bytes.
Not covered: rule oracle on current main, full repository gate, whole-rule integration; no new claim because the rule branch is neither landed nor landing-ready.

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
