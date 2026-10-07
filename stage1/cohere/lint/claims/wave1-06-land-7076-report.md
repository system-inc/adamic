Integrated current area 7076b4eb and retained current main 71d7e491 ancestry; owned rule sources unchanged.
Validated implementation cc8f813b800f245c6830b0fbc858c959aa521c7f; final evidence pushed only to codex/lint-wave1-06-land.
Fresh witnesses, rule parity and 7 owned mutants PASS 304.046s.
All 7 mutants caught by independent Go output comparison on source Node, emitted JavaScript and sanitized native.
Full gate, full mutant registry, TestShardsAgree, standalone profiles, throughput and setup not repeated; no new claim.

Area 7076b4ebe16a296129d17f04fe0e14029d793e92 brings upstream-authored single-pass fixed-text construction and sharded manifest execution. Main 71d7e491 remains current. Clean area merge; DEDUP_LEDGER.md unchanged; no shared-file changes authored. This run verifies the ordinary driver and fixed-source result against Go. No fresh sharded-mode coverage is claimed.

Command sourced /workspace/adamic-tools/env.sh, exported GOPROXY=https://proxy.golang.org|direct, all output directly to committed tests.log:
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$/^(wrong-diagnostic-id(#01)?|wrong-global-type-name|omit-empty-export-fix|wrong-enclosing-scope|configured-restriction-ignored|report-left-reference)$' -count=1 -v -timeout 20m

    lint_test.go:445: Go, Node, emitted JavaScript, native identical: 13203441 bytes
--- PASS: TestRulesAgree (64.44s)
--- PASS: TestMutants (206.55s)
    --- PASS: TestMutants/wrong-diagnostic-id (28.85s)
    --- PASS: TestMutants/wrong-diagnostic-id#01 (28.18s)
    --- PASS: TestMutants/wrong-global-type-name (28.50s)
    --- PASS: TestMutants/omit-empty-export-fix (28.61s)
    --- PASS: TestMutants/wrong-enclosing-scope (28.79s)
    --- PASS: TestMutants/configured-restriction-ignored (30.73s)
    --- PASS: TestMutants/report-left-reference (30.13s)
    registration_test.go:45: Go, Node, emitted JavaScript, native identical: 120217 bytes
--- PASS: TestOwnedWitnesses (33.03s)

Every mutation executes successfully and differs from independent Go on all three runtimes. Option-bearing witnesses retain the previously repaired branches, and no anchors or guards were weakened. Inherited explicit parser-recovery exclusions remain outside parity coverage. Required compiler corpus was run on wave07 separately, not on this takeover branch. Other required-input packages were not run.

consistent-return and constructor-super reservations remain only on the original wave07 unfinished-work branch. No descriptors or directories for them are imported into its finished-only landing branch. No helper is claimed. Push own names only, never main or area.
