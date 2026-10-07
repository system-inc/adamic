Probed the parked 12-rule branch against published shared harness ab70f38d47de1d4974082b38f84a56af2368b7af in a detached worktree.
Current main remains c01907a7; rules before this evidence commit are 03ef9828e, helpers f24b1c8dd. No shared files on the owned branches change.
Registry discovers all 12 owned .a rules and the lint package compiles; TestOwnedWitnesses fails in Go's serializer with unexpected fix shape after 94.410s.
Existing rule/declaration and helper semantic mutants remain recorded in landing-c01907a7; this probe stops before backend comparisons and proves no new parity or mutant result.
Remaining shared blocker: multiple automatic fixes are rejected by both the Go oracle and reportRange; @typescript-eslint/prefer-as-const's real two-fix witness demonstrates it. No new claims.

The user named origin/lint-rules/harness ab70f38d4 while it is being merged into area/stage1-lint. It does not contain current origin/main and is not yet on main. Both owned pushed branches remain based on current main and their prior current-main oracles are green. This probe prepares integration without pushing to main or any area branch.

Detached worktree /tmp/adamic-wave15-harness-probe was created from codex/lint-wave1-15. The named harness was merged without a commit. Three conflicts in lint.ts, lint_test.go and testdata/oracle.go were resolved by taking their exact ab70f38d4 published versions, rather than writing shared changes. The unchanged cohere submodule was exposed through a temporary symlink for builds. No probe merge is pushed.

Commands (source /workspace/adamic-tools/env.sh):

- git fetch origin '+refs/heads/*:refs/remotes/origin/*' --no-recurse-submodules
- git worktree add --detach /tmp/adamic-wave15-harness-probe codex/lint-wave1-15
- git -C /tmp/adamic-wave15-harness-probe merge --no-commit ab70f38d4
- git restore --source=ab70f38d4 --stage --worktree stage1/cohere/lint/lint.ts stage1/cohere/lint/lint_test.go stage1/cohere/lint/testdata/oracle.go
- go run ./cmd/lint-registry: PASS, all 12 owned modules appear among 27 registered rules.
- go test ./stage1/cohere/lint -run '^TestRegistered' -count=1 -v -timeout=20m: compiles, but warning no tests to run. This is not oracle evidence; the filter was corrected below.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=20m: FAIL in the Go oracle, before source Node, emitted JavaScript or native comparison.

The exact shared failure is panic: unexpected fix shape at testdata/oracle.go:81, from its len(d.Fixes) != 1 guard. The shared context's reportRange also explicitly panics when edits.length > 1. The harness now supports .a discovery, seven-argument report returning Finding, independent single-edit ranges, suggestions and optional node:true callbacks. The older registration and profile_test.go build blockers are therefore obsolete for this published snapshot, but multiple automatic edits remain unsupported.

Actual unchanged Go cohere proof for the owned prefer-as-const witness let value: 'bar' = 'bar';:

- One preferAsConst diagnostic at bytes 11..16.
- Automatic fix 9..16 inserts an empty string, removing the annotation.
- Automatic fix 24..24 inserts " as const".
- Applied source is let value = 'bar' as const; followed by a newline.

The complete-proposal Go oracle already used by this unit reports both edits; its raw output is retained in prefer-as-const.txt. These are two independently ranged fixes, not one suggestion. Combining them into one wider replacement would change the fix representation and does not satisfy the requested byte-for-byte fix comparison. No edit is discarded to make the shared serializer pass.

Our owned complete-proposal results and all prior supported four-way comparisons remain available. The current owned publish bridge deliberately refuses unsupported fix shapes rather than dropping data. Adapting it to the new APIs must be validated after the shared automatic-edit representation is complete; no unlanded API import is introduced on the current-main branch. The ten prior parser exclusions are not declared resolved merely because this harness adds JSX parsing; this failed test never reaches their backend comparison.

Stop at this named shared-file boundary, as instructed. The rule branch remains parked. DesignSystemForProgram's separate missing-Mutex prerequisite remains recorded and released on the helper branch. Other helper work is not claimed, and the queue is not declared exhausted. No allocator leak-check changes are reverted; current main has not advanced since the previous landing proof.
