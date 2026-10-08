Merged lint area c4bdc23fa86d55cf7e579989201c11258f4d3a62 into the owned branch; no rebase or shared checks removed.
Rule witness and message mutant PASS on Go, source Node, emitted JavaScript and sanitized native.
Full lint package: 29 PASS, 5 FAIL, 1 SKIP; all optional inputs supplied; JSX storage failure retry PASS.
Registry semantic mutants caught: 76; no-explicit-any remains blocked on recovered interface-constructor parsing.
Full-gate wall 1619.753s; nproc 5; one-minute load min 0.73, max 9.76, final 1.68.

This syntax-only rule uses no private checker stand-in or recorded checker answers.
The actual Go adapter continues decoding its own NoExplicitAnyOptions and the
upstream prefix remains TestNoExplicitAny, capturing all 14 test names.

The parser blocker is cohere/internal/lint/rules/typescript/no_explicit_any_test.go:54,
TestNoExplicitAnyFires (upstream_fail_15), reporting unexpectedAny through
NoExplicitAny at no_explicit_any.go:120. evidence/reproducer.ts.txt is unchanged.
Source Node and native reject CloseBraceToken at byte 54; no recovered case was
removed or weakened. This also fails TestNodeTableIsLinkOnly, TestShardsAgree and
TestProfileSnapshotsAgree. Full upstream/options/fix certification remains blocked.

TestJsxLintTrees initially failed writing the module cache on a full workspace.
The same check passed after moving that cache to scratch storage, on 63 captured
sources and 51432 identical whole-tree bytes. The original full-run counts remain
reported rather than rewriting a failure as a pass.

TestCheckerBridgeRefusalPending is the single skip at
stage1/cohere/lint/checker_pending_test.go:49. It requires TSGoError in the library
prelude and the tsgoInspect C error-buffer result from codex/tsgo-errors-as-values.
Supplying test inputs cannot satisfy that API dependency. Its check is intact.

Commands and complete outputs are retained under evidence/unpark/:
- go test -json -count=1 -timeout=90m ./stage1/cohere/lint
- go test ./stage1/cohere/lint -run '^TestRulesAgree$|^TestOwnedWitnesses$|^TestMutants$/typescript-no-explicit-any-message$' -count=1 -timeout=30m -v
- go test ./stage1/cohere/lint -run '^TestJsxLintTrees$' -count=1 -timeout=15m -v
- go run ./cmd/lint-registry; go vet ./stage1/cohere/lint/...
Environment: ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript,
ADAMIC_LINT_BENCH=1, ADAMIC_LINT_PROFILE_DIR and
ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/unpark-any-profiles, GOMAXPROCS=4.
TypeScript corpus pin: 050880ce59e30b356b686bd3144efe24f875ebc8.
Setup succeeded in 168.770s (cache warm 168.744s); no new claims or PR.
