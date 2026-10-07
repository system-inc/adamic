Built: verified the unchanged wave 13 area base and ran the shared parser correctness gate with its pinned TypeScript input supplied; no new claims.
Commits: code remains b1a890cd on area d65a8f93 and main 39638d9e; this report records the additional required-input check.
Commands and outputs: TestCompilerExpressionsAgree PASS 35.856s, 77 whole compiler files and 28,836,875 identical expression tree bytes; no skips.
Mutants: no mutants rerun for unchanged code; the nine rule, two export-question, lifetime and 27 metadata proofs remain in LANDING_RUNTIME_REPORT.md.
Not covered: the other required-input checks, full repository gate, three parser recovery controls and shared checker-context integration.

Fetched all origin branches. Main and area/stage1-lint are unchanged, the worktree
was clean, and the previous own push was verified at b1a890cd. No rebase was needed.
The announced required-input enforcement changes are not present in this fetched
snapshot. No check was skipped, relaxed, removed or changed by this unit.

Executed with the real pinned checkout explicitly supplied:

```
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave13-corpus go test ./stage1/typescript/parser -run '^TestCompilerExpressionsAgree$' -count=1 -timeout=15m -v > /workspace/wave13-required-parser.log 2>&1
```

The test verifies the checkout pin, enumerates all 77 compiler files and compares
the Go TypeScript parser oracle against Node and sanitized native parser output.
All 28,836,875 expression-tree bytes agree. The unset-input skip branch was not
taken. This positive compiler corpus does not contain the three invalid-syntax
recovery fixtures that still refuse before no-object-constructor can execute.
Those failures remain recorded in LANDING_RUNTIME_REPORT.md with their exact cases.

The other required-input gates, including postcss, graphql-js and postcss-parser
comparisons, were not invoked here. No full-gate green result is claimed. The
unchanged-code wave 13 oracle, sanitizer, mutant and runtime evidence remains
applicable. No new timing benchmark was run; prior native/Go measurements remain
historical. The toolchain was reused, with prior setup total 217s and nproc 5.
No source or shared harness changes were made, and no further rule was claimed.
