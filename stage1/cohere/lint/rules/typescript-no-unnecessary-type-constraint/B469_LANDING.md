Rules: clean rebase of seven retained ports onto the updated area registry/context/oracle; no owned implementation change.
Commits: previous pushed rule tip 125099b537d489234ade626a8a6656047e1598cf; pre-evidence tip 296a0e2359e2d86d25253428617ca291f20c794a; helper stays pushed at eab7ae08b051fcc9b736378790436f26678c102f.
Checks: fresh supported parity PASS 400.645s, six mutants PASS 252.142s, vet zero, external oracle PASS 0.620s; full witness gate remains red.
Mutants: six rule semantic mutants comparison-killed on Node, emitted JavaScript and ASan/UBSan native; const annotation mutant still blocked by its baseline refusal.
Limits: multi-edit reporting and malformed computed-key parser recovery prevent full parity; no new claim, broader seventeen correctness checks or throughput benchmark.

Fetched every origin head after the managed environment resumed. Current main remains c7991b900362796aefd111474e65eb5398e91953; current area is b46914832d70e00847d82d5d221ab7bb24040c53. Both bases were checked again after the comparison gates and are unchanged. Rebased cleanly onto the area branch, accepting new registered rules, context helpers and the registry-only oracle without editing shared files. The dedup ledger content is unchanged, SHA256 4c39ec0cb129b05a3971ff257c26296d0d4ce06545b42e53dc14bbf0828a526f. The seven retained and eleven retired copies remain those in AREA_LANDING.md. No new batch assignment is inferred.

Fresh commands, from this worktree after source /workspace/adamic-tools/env.sh:

```sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-08-typescript go test -overlay=/tmp/wave08-unified-overlay.json ./stage1/cohere/lint -run '^(TestWave08RetainedUpstream|TestWave08RetainedCorpus|TestWave08ConstSingleEdit)$' -count=1 -v -timeout=20m > /tmp/wave08-b469-supported.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestMutants$/(core_default_suppressed|computed_replacement_wrong|foreign_read_suppressed|gating_invalid_suppressed|constraint_suggestion_wrong|enum_second_suggestion_wrong)$' -count=1 -v -timeout=20m > /tmp/wave08-b469-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=15m > /tmp/wave08-b469-witness.log 2>&1
go vet ./... > /tmp/wave08-b469-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave08-b469-oracle.log 2>&1
```

The overlay adds only the owned unified_test.go.txt as a virtual test file; it replaces no shared comparator, oracle, registry or driver. The real shared comparison runs the actual Go oracle against source Node, emitted JavaScript and sanitized native. The TypeScript checkout is provided and the test checks its pin, 050880ce59e30b356b686bd3144efe24f875ebc8. No selected test skips for missing input.

Upstream PASS 116.90s: 414 supported cases, all four outputs identical at 149648 bytes. Counts: computed-key 120, gating 42, foreign propTypes 60, constraint 43, enum 21, core default-param-last 128. All JSX cases are included. The same two malformed computed-key sources remain excluded and explicitly blocking.

The new shared source grows the corpus to all 411 .ts/.a files and 2466 selected rule cases. Corpus PASS 240.08s, all four outputs identical at 80209179 bytes. Three const single-edit/negative controls PASS 43.48s, all four outputs identical at 1011 bytes. Combined selected gate PASS 400.645s. These observations hold only the stated supported scope, not the full gate.

Six mutants PASS 252.142s. All compile and execute normally on the three port backends, and only Go output comparison kills their semantic change: core_default_suppressed suppresses default-param-last; computed_replacement_wrong corrupts its fix; foreign_read_suppressed suppresses foreign propTypes; gating_invalid_suppressed suppresses invalid configuration; constraint_suggestion_wrong corrupts a suggestion; enum_second_suggestion_wrong corrupts the second enum suggestion. const_append_wrong still needs two automatic annotation edits, and its baseline refusal is not counted as a mutant kill. Its earlier rich-transport proof remains historical.

Whole-repository vet exits zero with empty log. The uncached external one-byte oracle PASS 0.620s, zero native/Node cache hits. Full TestOwnedWitnesses FAIL 63.108s: actual Go exits 2 with panic: unexpected fix shape at its automatic-edit guard. Fresh direct Node probes of let value: 'hello' = 'hello'; and ({ ['x' }); still exit 70 with a finding carries at most one automatic edit and parser slice expected CloseBracketToken, got CloseBraceToken at 8, respectively. Raw logs are in b469_evidence/. No shared model is edited or multi-edit repair collapsed into an inaccurate single edit. The parser gap prevents the shared-harness-only parking exception; no helper is claimed.

The helper branch remains on unchanged current main with its existing pushed green evidence: four actual-Go/Node/emitted-JavaScript/sanitized-native baselines and four comparison-only mutants PASS 108.340s, vet zero, uncached one-byte oracle PASS 45.093s. These are not freshly rerun here. The four helper mutants are nonzero NewTheme deadKeys, wrong ClearNamespace hyphen boundary, omitted var( refusal and reversed unresolved breakpoint ordering. The same six Tailwind consuming rules are listed in helpers/from_wave08/REPORT.md and BREAKPOINTS_REPORT.md; none alone removes all final blockers.

Toolchain setup in the resumed environment PASS: Go ready 0s; clang, Node and submodules ready 1s each; cache warm 92s; total 92s; nproc 5 with four-CPU quota. Tools remain Go 1.27.1, clang 20.1.8 and Node 24.19.0. No full repository gate or broader seventeen external-input correctness checks run. Test output is written directly to logs, never piped. No checks are relaxed or deleted, no new implementation file is written, and no main or area push is attempted.
