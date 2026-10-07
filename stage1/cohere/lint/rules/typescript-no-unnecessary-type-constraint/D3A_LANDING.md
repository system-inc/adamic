Rules: clean rebase of seven retained ports onto the current area; no owned implementation change.
Commits: previous pushed rule tip ded4a1b8495454b0ea9af68ce5627c5eece3d3ca; pre-evidence tip a5b7cc9187d8183d30da561c17335f984c229527; helper evidence pushed at a1d998524.
Checks: fresh supported parity PASS 542.467s; six mutants PASS 403.237s; vet zero; external oracle PASS 9.243s; full witness gate remains red.
Mutants: six rule and four helper semantic mutants freshly caught only by Go output comparison; const annotation mutant remains blocked.
Limits: multi-edit reporting and malformed-key parser recovery prevent full parity; no new claim, broader seventeen correctness checks or fresh throughput benchmark.

Fetched every origin branch. Current main is b6b1538b0cebc4ba6741ac34f1aedb60293c1d06; area/stage1-lint is d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 and contains that main. Both remote bases were rechecked after the gates and are unchanged. Rebased cleanly onto area, accepting inherited typeof lowering and union/native runtime changes without editing them. The previously read dedup ledger is byte-identical, SHA256 4c39ec0cb129b05a3971ff257c26296d0d4ce06545b42e53dc14bbf0828a526f. The seven retained and eleven retired copies remain those in AREA_LANDING.md. No new batch rule assignment is inferred and no shared harness file is edited.

Fresh commands use source /workspace/adamic-tools/env.sh and redirect directly to log files:

```sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-08-typescript go test -overlay=/tmp/wave08-unified-overlay.json ./stage1/cohere/lint -run '^(TestWave08RetainedUpstream|TestWave08RetainedCorpus|TestWave08ConstSingleEdit)$' -count=1 -v -timeout=20m > /tmp/wave08-d3a-supported.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestMutants$/(core_default_suppressed|computed_replacement_wrong|foreign_read_suppressed|gating_invalid_suppressed|constraint_suggestion_wrong|enum_second_suggestion_wrong)$' -count=1 -v -timeout=20m > /tmp/wave08-d3a-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=15m > /tmp/wave08-d3a-witness.log 2>&1
go vet ./... > /tmp/wave08-d3a-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave08-d3a-oracle.log 2>&1
```

The overlay adds only the owned unified_test.go.txt as a virtual test file, replacing no shared source. It calls the actual registry Go oracle and unmodified shared comparison. The supplied compiler checkout must match pin 050880ce59e30b356b686bd3144efe24f875ebc8, which the test checks. No selected owned test skips for missing input.

Supported upstream PASS 256.48s: 414 cases produce identical actual-Go/source-Node/emitted-JavaScript/ASan-UBSan-native output, 149648 bytes. Counts: computed-key 120, gating 42, foreign propTypes 60, constraint 43, enum 21, core default-param-last 128. JSX cases are included. The two malformed computed-key cases remain explicitly excluded and blocking.

Corpus PASS 246.54s: all 411 compiler/stage1 .ts/.a files and 2466 selected rule cases match on all four executions at 80209179 bytes. Three const single-edit/negative controls PASS 38.79s, identical at 1011 bytes. Combined selected gate PASS 542.467s. These certify the stated supported scope, not full landing readiness.

Six semantic mutants PASS 403.237s: core_default_suppressed suppresses the parameter diagnostic; computed_replacement_wrong corrupts the fix; foreign_read_suppressed suppresses foreign propTypes; gating_invalid_suppressed suppresses invalid configuration; constraint_suggestion_wrong corrupts a suggestion; enum_second_suggestion_wrong corrupts the second enum suggestion. Each compiles and executes normally on all three port backends, with only Go byte comparison killing it. const_append_wrong still depends on two automatic annotation edits; its baseline refusal is not a mutant kill. Its earlier rich-transport proof is historical.

Whole-repository vet exits zero with empty log. Uncached external one-byte oracle PASS 9.243s, zero native/Node cache hits. Full TestOwnedWitnesses FAIL 214.369s: actual Go exits 2 with panic: unexpected fix shape at its automatic-edit guard. Direct Node probes of let value: 'hello' = 'hello'; and ({ ['x' }); still exit 70 with a finding carries at most one automatic edit and parser slice expected CloseBracketToken, got CloseBraceToken at 8. Raw logs are in d3a_evidence/. The independent annotation edits remain intact; no inaccurate repair is substituted and no shared guard is relaxed. The parser gap prevents the shared-harness-only parking exception; stop before any new helper claim.

The helper branch also rebased cleanly onto current main and was pushed at a1d998524. Fresh uncached four-helper actual-Go/Node/emitted-JavaScript/sanitized-native comparison and four comparison-only mutants PASS 130.689s, vet zero, external one-byte oracle PASS 37.375s with zero cache hits. The helper mutants are nonzero NewTheme deadKeys, wrong ClearNamespace hyphen boundary, omitted var( refusal and reversed unresolved breakpoint ordering. Their exact commands and raw logs are helpers/from_wave08/LANDING.md and evidence/landing_b6b_*.log on that branch. The same six Tailwind consuming rules are documented in its reports; none of these helpers independently removes all final blockers. No new helper is claimed or written.

bash cloud/setup.sh PASS: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 2s, cache warm 271s, total 271s; nproc 5 and CPU quota four. Tools remain Go 1.27.1, clang 20.1.8 and Node 24.19.0. No full repository gate, broader seventeen external-input correctness checks or fresh throughput measurement. Test output is never piped. No check is relaxed or deleted to obtain green. Push only owned branches using exact old-head leases, never main or area.
