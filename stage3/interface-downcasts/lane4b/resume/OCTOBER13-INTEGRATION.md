Built: advanced only to views-integration ffe428ab26eefb73154adbf570ad99e9c1b4f872; four obsolete compile-frontier pins now hold exact runtime obligations.
Commits: prior lane tip 8abb52a1518b9d8c3f0dbde993551d4f01ba10e2; integration was a clean fast-forward, with no conflict choices needed.
Checks: all original pairs/frontiers PASS 55.966s; other lane 4b controls, mutants, leaks and golden counts PASS 88.437s; setup PASS 33.048s, nproc 5, quota four.
Mutants: each of four new runtime pins catches a demanded-path omission after successful Node-equivalent execution in all three modes; 12 semantic catches.
Uncovered: seven inherited intersection certificates are freshly validated and await ledger adoption; nine pairs / 23 reads still compile-refuse; no full gate.

The attempted non-null compiler-area merge c41c0e062e99da37820f822968d4df1b48cdaee7
was aborted immediately on the user's correction. No compiler-area merge commit
or change was pushed. It is not an ancestor of this branch. No resolution edits
from that attempt survive. Automatic review had rejected a broad heuristic
conflict script; narrow explicit edits were subsequently discarded by the abort.
There was no rejected action left needing approval.

Four reduced original producers now lower but lack required Identifier fields.
The exact exit-70 messages name node.left.expression.escapedText or
node.parent.escapedText, expected __String, found missing. Those checks cannot be
replaced by the earlier unsupported-intersection compile refusal. Their mutants
remove guards along the demanded path, including a possible earlier ancestor
check. They execute the original Node output in sanitized native, release native
and JavaScript, and pass a separate native leak check. Independent per-pair
member/outer/nested mutants remain in lane 7's complete original witnesses.

All 12 remaining-frontier probes were rerun, including the two JSDoc parent
probes integration already converted into runtime obligations. All earlier
original object-plus-primitive certificates remain passing. The 16 comments,
PackageJsonInfoContents pairs, reduced shape controls and three defect controls
also passed, with existing golden counts unchanged and declaration tests enabled.
The initial recheck failed only on four obsolete compile-refusal expectations;
its complete log is retained as an observation, not an inherited-failure claim.

Commands, tool environment sourced and output directly redirected to logs:

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
git fetch origin '+refs/heads/codex/views-integration:refs/remotes/origin/codex/views-integration' '+refs/heads/codex/non-null-checked-area:refs/remotes/origin/codex/non-null-checked-area' '+refs/heads/codex/views-object-primitive-unions:refs/remotes/origin/codex/views-object-primitive-unions'
git merge --no-edit origin/codex/views-integration
git merge --no-edit origin/codex/non-null-checked-area
git merge --abort
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitiveOriginalPairs$|^TestCheckedViewObjectPrimitiveRemainingFrontiers$' -count=1 -v -timeout 15m
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitive(Source|FixCounts|Fixes|CommentPairs|PackagePairs)$' -count=1 -v -timeout 15m
```

Setup timing: Node .030s, Go .039s, markdown .088s, submodules .115s, clang .310s,
Go build 32.893s, cache warm 33.010s, done 33.048s. No setup failure.
