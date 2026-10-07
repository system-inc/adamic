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

## Latest main recheck

Fetched current main e8ba3d5d81de4d3773c723914fccd4c76248b965. Its entire
stage1/cohere/lint tree is unchanged from e011f8f6 (git diff --exit-code exits 0).
The helper evidence branch rebases cleanly again. Both archived comparison
validators and their compiling mutants pass on all three Adamic paths against Go
with unchanged case and byte counts. Repository vet passes with empty output;
the filtered external oracle passes in 13.263s. Missing live consumer fixtures
remain explicitly excluded from that green comparison claim.

The rule rebase against this exact new main fails again at foundation commit
29175443 in the same four shared files. See evidence/rule-landing-rebase.log.
The rebase is aborted and the rule branch stays at 9d2c673b, unlanded and not
landing-ready. No shared conflict file is resolved or committed. The explicit
ownership limit still blocks preserving main's thirty rules while migrating
the five-rule registration foundation. No new helper or rule is claimed, and
no push targets main or an area branch.

## Main f8013f0b recheck

Fetched origin/main f8013f0baac41ddc340d76f83bddde38536a8f07. The shared
lint tree is still unchanged from e8ba3d5d, while compiler lowering/runtime
changes have landed. Rebased the helper evidence branch cleanly and reran
both archived validators: 4,761 integer inputs / 160,778 bytes and 1,012
static declaration cases / 139,208 bytes match actual Go on source Node,
emitted JavaScript and sanitized native. Both clean compiling semantic
mutants are caught on every path again. Vet passes with empty output, and
the filtered external oracle passes in 17.392s. Live consumer fixtures
remain missing and are not credited as passing.

A fresh rule rebase against f8013f0b again fails at registration foundation
29175443 in the same four unowned shared files; the updated exact output is
in evidence/rule-landing-rebase.log. Aborted without resolving or committing
those files. Rule branch remains 9d2c673b and is not landing-ready. No new
helper or rule is claimed, and no main/area branch is pushed. The finding
model landing SHA has not been supplied.

## Recheck after stage3 landing

Rebased onto current main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Bounded withdrawn leadingInteger and nodesFromStaticDeclarations comparisons pass again on source Node, emitted JavaScript and sanitized native: 4,761 inputs / 160,778 bytes and 1,012 cases / 139,208 bytes. Both compiling semantic mutants are again caught only by Go comparison on all three paths. They remain non-executable archived witnesses, not delivered helpers; the six missing live Tailwind consumer paths remain explicitly unpassed. No active helper reservation has been added before this rebase and push.

## Current compiler landing b8fb957aa
Rebased and re-green after inherited static-field emission landed. Retained fixed-hex helper: 68,844 unique inputs / 2,149,768 bytes, actual Go private helper compared with Node, emitted JavaScript and sanitized native; width mutant caught on all three. All four original consumer suites pass, with additional real hex-path calls in all four consumers. Archived withdrawn integer/declaration witnesses pass again without delivery credit. Full consumer findings on Adamic require the surrounding regexp/rule integration and are not claimed by this leaf result. See HEX_REPORT.md.

## Latest main 39638d9e2
Rebased without conflict. Retained fixed-hex helper rebuilt and re-green on all three runtimes with the Go oracle and width mutant. Original four consumer suites and targeted actual leaf calls pass again. Filtered inherited-field and one-byte oracles pass. Shared rule-harness blockers remain on the parked rule branch; no new helper claimed during landing-first work.
