Built: rebased seven retained rule ports onto current main and merged current area cleanly; no implementation change or new claim.
Commits: previous pushed rule eff78d579910221d72e0db22f7655874675d573e; pre-evidence 0f6a2d46951240968842cd1a759dd947f5fce44e; helper pushed 58ccae567.
Checks: supported parity PASS 372.918s, six mutants PASS 295.766s, vet zero, one-byte oracle PASS 13.553s; full owned witness FAIL 51.732s.
Mutants: six rule and four helper mutants caught only by Go byte comparison on Node, emitted JavaScript and sanitized native; const annotation mutant remains blocked.
Limits: shared multi-edit finding/oracle and malformed computed-key parser recovery prevent full parity; no new helper claim, broader seventeen checks or fresh throughput benchmark.

Main is 4e0bfda50a19c705a1aac0d9932e08483806d61c. Area is bb2ece564842c4b2f909b9f75c27e74c2efa4f29. Rebased cleanly onto main, then cleanly merged area because its newer witness-options validation is not yet in main. No manual shared-file resolution or shared source edit was made. Both heads are ancestors of this branch. Read the dedup ledger before the rebase; its seven retained winners and eleven retired copies are unchanged.

Commands, all after source /workspace/adamic-tools/env.sh, with direct log redirection:

```sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-08-typescript go test -overlay=/tmp/wave08-unified-overlay.json ./stage1/cohere/lint -run '^(TestWave08RetainedUpstream|TestWave08RetainedCorpus|TestWave08ConstSingleEdit)$' -count=1 -v -timeout=20m > /tmp/wave08-4e0-supported.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestMutants$/(core_default_suppressed|computed_replacement_wrong|foreign_read_suppressed|gating_invalid_suppressed|constraint_suggestion_wrong|enum_second_suggestion_wrong)$' -count=1 -v -timeout=20m > /tmp/wave08-4e0-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=15m > /tmp/wave08-4e0-witness.log 2>&1
go vet ./... > /tmp/wave08-4e0-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave08-4e0-oracle.log 2>&1
```

The owned overlay adds unified_test.go.txt as a virtual test, replacing no shared source. It calls the actual shared Go oracle and comparator. This is supported-scope evidence, not a final-push receipt or full correctness assertion. The supplied TypeScript checkout is pinned to 050880ce59e30b356b686bd3144efe24f875ebc8.

Upstream: 414 cases match on all four executions, 149648 bytes, PASS 88.51s. Counts are computed-key 120, gating 42, foreign propTypes 60, constraint 43, enum 21, default-param-last 128. Exactly two malformed computed-key cases remain explicitly excluded and blocking: ({ ['x' }); and ({ ['x': 0 });. Corpus: all 411 compiler/stage1 .ts/.a files, 2466 selected rule cases, match at 80209179 bytes, PASS 241.88s. Three single-edit/negative const controls match at 1009 bytes, PASS 42.47s.

Rule mutants: core_default_suppressed suppresses parameter diagnostics; computed_replacement_wrong corrupts replacement bytes; foreign_read_suppressed suppresses propTypes diagnostics; gating_invalid_suppressed suppresses invalid configuration; constraint_suggestion_wrong corrupts a suggestion; enum_second_suggestion_wrong corrupts the second enum suggestion. Each compiles and executes successfully on all three port backends; only different Go output catches it. const_append_wrong still needs two automatic annotation edits: its blocked baseline is not a kill.

Full TestOwnedWitnesses fails before comparison: Go exits 2 with panic: unexpected fix shape at adamic_lint_oracle.go:75. Minimal source const x: 'x' = 'x'; exits 70 on Node with a finding carries at most one automatic edit. Minimal ({ ['x' }); exits 70 with parser slice expected CloseBracketToken, got CloseBraceToken at 8. The original independent Go edits are retained and no shared guard is bypassed. This is not the shared-harness-only parking exception because parser recovery also blocks full parity. Stop before any additional helper claim.

The helper branch rebased onto current main and pushed 58ccae567. Fresh uncached four-helper Go/Node/emitted-JavaScript/ASan-UBSan-native comparison and four semantic mutants PASS 57.933s. Mutants: nonzero NewTheme deadKeys, incorrect ClearNamespace hyphen boundary, omitted var( refusal, reversed unresolved breakpoint comparison. Six Tailwind consumers are listed in its reports; none is independently fully unblocked by these helpers.

Whole-repository vet exits zero with empty log; external one-byte oracle PASS 13.553s with zero cache hits. Setup timing: Go 0.068s, Node 0.118s, submodules 0.324s, clang 0.641s, markdown dependencies 1.665s, cache warm 92.691s, total 92.831s. nproc 5; cgroup quota four CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. No full repository gate, broader seventeen external-input correctness checks or fresh throughput measurement. Historical performance evidence remains in the owned claims and rule reports. Raw logs are main4e0_evidence/. Only owned branches are pushed with exact old-head leases; never main or area.

The newly integrated final-push checker was run on committed candidate ef7c27b05 using ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-08-typescript python3 cloud/lint-wave-check.py --claim stage1/cohere/lint/claims/wave1-08.md with output directly redirected. It exits 1 before tests: lint-wave-check: FAIL origin: unreadable legacy selection origin/codex/lint-helpers-from-lint-wave1-05:stage1/cohere/lint/helpers/wave05/evidence/cache-refresh-selection.json. No receipt is produced and this report does not assert final readiness. Do not edit the other worker's file or shared checker to pass. The user explicitly requests the owned rebased work be pushed, so publish this blocked candidate to its own branch while keeping the work-in-progress cap closed.
