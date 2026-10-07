Built: rebased helper evidence onto current main; rule branch rebase blocked by unowned shared-file conflicts.
Commits: main e011f8f60899586d6373a5ccb07335ad82cfbf3c; helper rebase head 3b2a3c60; rule branch remains 9d2c673b.
Checks: archived helper oracle comparisons rerun successfully on Node source, emitted JavaScript and sanitized native.
Mutants: ignored negative sign and erased value presence compile, exit zero with clean stderr, and are caught by real Go on all three paths again.
Limits: rule branch is not landing-ready; no new helper is claimed under the work-in-progress cap.

The two pushed branches owned by this worker were checked after fetching every origin head. Neither old tip is an ancestor of current main. Rebase of codex/lint-helpers-from-codex-lint-wave1-14 onto origin/main completes cleanly. Independent validators produce exactly the same 4,761 integer inputs / 160,778 bytes and 1,012 declaration cases / 139,208 bytes as before. No helper is delivered from these candidates: both remain yielded to their earlier owners. All six live consumer suites still record zero calls due to the already documented missing private Tailwind installation and stylesheet; canonical and unknown coverage guards still fail, so this is only the archived oracle comparison pass.

The rule branch rebase was attempted in /workspace/adamic-landing-wave14. Its first replayed registration-foundation commit, 29175443, conflicts in:

- stage1/cohere/lint/README.md
- stage1/cohere/lint/lint.ts
- stage1/cohere/lint/lint_test.go
- stage1/cohere/lint/testdata/oracle.go

Main's shared driver and oracle contain thirty rules; this old foundation extracts only five into directory registration. Resolving by taking the old foundation would discard newer main rules. Preserving all main rules requires migrating unowned shared driver/oracle/test code to directory registration, which the unit instructions explicitly assign elsewhere and prohibit this worker editing. The rebase was aborted successfully, preserving 9d2c673b and leaving its checkout clean. This is an observed rebase/integration blocker, not a failed rule comparison. No new rule comparison against current main is claimed because a valid rebased rule checkout was not obtained. The integration owner must land/rebase the registration foundation while retaining main's rules before this branch can be made landing-ready within its ownership boundary.

Reproduction commands, all test output sent directly to files:

```sh
git fetch origin '+refs/heads/*:refs/remotes/origin/*'
git rebase origin/main
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot14/validate.py > /tmp/helper14-landing-integer.log 2>&1
python3 stage1/cohere/lint/helpers/slot14/validate_nodes.py > /tmp/helper14-landing-nodes.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot14/evidence/landing-vet.log 2>&1
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestCountsAreRecorded' -count=1 -timeout 30m > stage1/cohere/lint/helpers/slot14/evidence/landing-filtered-oracle.log 2>&1
gofmt -l cmd internal > stage1/cohere/lint/helpers/slot14/evidence/landing-format.log
```

The first rebase above runs on the helper branch. The separately attempted rule rebase uses git worktree add /workspace/adamic-landing-wave14 codex/lint-wave1-14 followed by git -C /workspace/adamic-landing-wave14 rebase origin/main, then git rebase --abort after the four conflicts. No shared conflict file was resolved or committed. Existing .generated artifacts are excluded. The rebased helper push uses an exact old-tip force-with-lease because the user explicitly requested rebasing already pushed branches; no unrelated branch or unobserved remote change is overwritten.

Post-rebase bounded gate: repository vet exits zero with empty output; gofmt output is empty; filtered external oracle passes in 16.987s. Full repository and live consumer gates are not claimed.
