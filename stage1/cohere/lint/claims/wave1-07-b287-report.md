Integrated current area b28757f33 while retaining current main 71d7e491 ancestry; four winning rule directories unchanged.
Validated implementation 38d66e87a8d25e4e9b29fb27cc7bf6db34ecfcab; evidence commit follows on codex/lint-wave1-07.
Rules, required compiler corpus and positive witnesses PASS 187.424s; node-table guard PASS 31.056s.
Four previously caught mutants remain in wave1-07-current-main-report.md; mutants not rerun against the changed harness.
Full gate, setup, throughput and vet not repeated; consistent-return and constructor-super unfinished, so no helpers claimed.

Fetched every origin head. Main is unchanged at 71d7e491. Area adds upstream fix reparsing under the original file path, fix-pass exhaustion serialization, and the link-only node table guard. Clean merge, no authored shared-file changes, DEDUP_LEDGER.md unchanged. Losing copies remain retired.

Commands sourced /workspace/adamic-tools/env.sh with all output directly to the committed logs:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave07-typescript go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestCompilerAndStage1Agree)$' -count=1 -v -timeout 15m
go test ./stage1/cohere/lint -run '^TestNodeTableIsLinkOnly$' -count=1 -v -timeout 10m

Rules: 13,139,824 identical bytes, 66.54s. Required TypeScript compiler and stage1 corpus: 403 files, 20,623,918 identical bytes, 90.26s. Positive witnesses: 121,785 identical bytes, 30.60s. Independent Go, source Node, emitted JavaScript and sanitized native agree on findings and fixes. Node-table guard: 2,489 rows, 13,158,874 identical bytes with and without unattached node rows. Required compiler input was supplied; no missing-input skip. Inherited explicit parser-recovery exclusions remain outside parity coverage. Other required-input packages and full repository gate were not run. No guard relaxed.

Historical mutants: ordering-relations-ignored, and-left-undefined-ignored, constructor-property-name-ignored, parenthesized alias lost. No fresh mutant or performance result is claimed. consistent-return still lacks compatible Judge and independently built EndReachable integration through RuleContext; constructor-super remains reserved and unimplemented, without attributing it to that same dependency. No helper branch or claim created, and no queue exhaustion claimed. Push only the own branch, never main or area.
