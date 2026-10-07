Built: rebased seven retained ports onto current main, preserving the clean area merge; no implementation changes or new claims.
Commits: previous pushed rule c2176226cf3bca4e581f1387ff3466eca5c79244; pre-evidence 18ec9c46d14ac63469bb7c027f562b171b109aee; helper pushed bdae8bc99.
Checks: supported parity PASS 426.882s, six rule mutants PASS 349.780s, four-helper gate PASS 23.184s, vet zero, external oracle PASS 0.564s; full witness FAIL 73.268s.
Mutants: six rule and four helper mutants caught only by Go comparison on Node, emitted JavaScript and sanitized native; annotation const mutant blocked.
Limits: multi-edit reporting/oracle and malformed-key parser recovery remain; no new helper claim, full repository gate, broader seventeen checks or fresh throughput measurement.

Rebased onto origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b using --rebase-merges, retaining origin/area/stage1-lint bb2ece564842c4b2f909b9f75c27e74c2efa4f29 as an ancestor. Both bases are present without manual shared source edits. Main changes only stage3; compiler, native, stage1 and CLAUDE sources are unchanged. The seven winners and eleven retired losers from the dedup ledger remain unchanged.

Commands use source /workspace/adamic-tools/env.sh and direct log redirection:

```sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-08-typescript go test -overlay=/tmp/wave08-unified-overlay.json ./stage1/cohere/lint -run '^(TestWave08RetainedUpstream|TestWave08RetainedCorpus|TestWave08ConstSingleEdit)$' -count=1 -v -timeout=20m > /tmp/wave08-71d-supported.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestMutants$/(core_default_suppressed|computed_replacement_wrong|foreign_read_suppressed|gating_invalid_suppressed|constraint_suggestion_wrong|enum_second_suggestion_wrong)$' -count=1 -v -timeout=20m > /tmp/wave08-71d-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=15m > /tmp/wave08-71d-witness.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave08 -count=1 -v -timeout=15m > /tmp/wave08-71d-helper.log 2>&1
go vet ./... > /tmp/wave08-71d-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave08-71d-oracle.log 2>&1
```

The rule overlay adds only the owned unified_test.go.txt virtual test, replacing no shared files. It invokes the actual unchanged registry oracle and comparator. It proves the stated supported scope and is not a full final-push receipt. The compiler checkout pin is checked as 050880ce59e30b356b686bd3144efe24f875ebc8.

Upstream: 414 supported cases, identical on actual Go/source Node/emitted JavaScript/ASan-UBSan native, 149400 bytes, PASS 96.13s. Counts: computed-key 120, gating 42, foreign propTypes 60, constraint 43, enum 21, default-param-last 128. Two malformed computed-key cases are explicitly excluded and remain blockers: ({ ['x' }); and ({ ['x': 0 });. Corpus: all 411 compiler/stage1 .ts/.a files and 2466 selected rule cases, identical 80209179 bytes, PASS 247.27s. Three const single-edit/negative controls match 1011 bytes, PASS 83.33s.

All six rule mutants compile and execute normally on all three port backends; only Go output differences kill them. core_default_suppressed suppresses parameter diagnostics; computed_replacement_wrong corrupts fix bytes; foreign_read_suppressed suppresses propTypes findings; gating_invalid_suppressed suppresses invalid configuration; constraint_suggestion_wrong corrupts a suggestion; enum_second_suggestion_wrong corrupts the second enum suggestion. const_append_wrong still needs two independent automatic annotation edits; baseline refusal is not a mutant kill.

Full TestOwnedWitnesses fails before comparison: actual Go exits 2 with panic: unexpected fix shape at adamic_lint_oracle.go:75. Fresh Node minimal probes exit 70: const x: 'x' = 'x'; hits a finding carries at most one automatic edit, and ({ ['x' }); hits parser slice expected CloseBracketToken, got CloseBraceToken at 8. No edits are collapsed and no guard is relaxed. Parser recovery is outside the shared-harness-only parking exception. Therefore stop before any new helper claim.

Helpers passed four actual-Go baselines and all four comparison-only mutants: nonzero NewTheme deadKeys, wrong ClearNamespace hyphen boundary, omitted var( refusal and reversed unresolved breakpoint ordering. The six Tailwind consumers remain documented in REPORT.md and BREAKPOINTS_REPORT.md on the helper branch; none is independently fully unblocked by these four helpers. Fresh helper logs are evidence/landing_71d_*.log there.

Setup PASS: Go 0.048s, Node 0.070s, markdown dependencies 0.177s, submodules 0.220s, clang 0.405s, cache warm 59.434s, total 59.479s; nproc 5, quota four CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. Vet exits zero with empty log; external one-byte oracle has zero cache hits. Raw rule logs are main71d_evidence/. No selected test is skipped for missing inputs. Only owned branches are pushed using exact expected-old-head leases, never main or area.
