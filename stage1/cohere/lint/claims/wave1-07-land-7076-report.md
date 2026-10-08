Integrated current area 7076b4eb and retained current main 71d7e491 ancestry; owned rule sources unchanged.
Validated implementation c5c289467e45f4ac8ed83212f074d92a6bec3a8a; final evidence pushed only to codex/lint-wave1-07-land.
Fresh witnesses, rule parity and 4 owned mutants PASS 275.478s.
All 4 mutants caught by independent Go output comparison on source Node, emitted JavaScript and sanitized native.
Full gate, full mutant registry, TestShardsAgree, standalone profiles, throughput and setup not repeated; no new claim.

Area 7076b4ebe16a296129d17f04fe0e14029d793e92 brings upstream-authored single-pass fixed-text construction and sharded manifest execution. Main 71d7e491 remains current. Clean area merge; DEDUP_LEDGER.md unchanged; no shared-file changes authored. This run verifies the ordinary driver and fixed-source result against Go. No fresh sharded-mode coverage is claimed.

Command sourced /workspace/adamic-tools/env.sh, exported GOPROXY=https://proxy.golang.org|direct, all output directly to committed tests.log:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave07-typescript go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestCompilerAndStage1Agree|TestMutants)$/^(constructor-property-name-ignored|ordering-relations-ignored|and-left-undefined-ignored|parenthesized alias lost)$' -count=1 -v -timeout 20m

    lint_test.go:445: Go, Node, emitted JavaScript, native identical: 13160434 bytes
--- PASS: TestRulesAgree (59.21s)
    lint_test.go:521: compiler and stage1: 403 files
    lint_test.go:522: Go, Node, emitted JavaScript, native identical: 20627031 bytes
--- PASS: TestCompilerAndStage1Agree (71.36s)
--- PASS: TestMutants (113.10s)
    --- PASS: TestMutants/parenthesized_alias_lost (27.50s)
    --- PASS: TestMutants/constructor-property-name-ignored (26.87s)
    --- PASS: TestMutants/ordering-relations-ignored (26.99s)
    --- PASS: TestMutants/and-left-undefined-ignored (29.12s)
    registration_test.go:45: Go, Node, emitted JavaScript, native identical: 126981 bytes
--- PASS: TestOwnedWitnesses (31.79s)

Every mutation executes successfully and differs from independent Go on all three runtimes. Option-bearing witnesses retain the previously repaired branches, and no anchors or guards were weakened. Inherited explicit parser-recovery exclusions remain outside parity coverage. The required 403-file input was supplied, with no missing-input skip. Original codex/lint-wave1-07 at f18c002f and this landing branch have identical stage1/cmd/internal executable trees excluding claim reports; this fresh run covers both. Only the four finished winning rule directories are carried on the landing branch. Other required-input packages were not run.

consistent-return and constructor-super reservations remain only on the original wave07 unfinished-work branch. No descriptors or directories for them are imported into its finished-only landing branch. No helper is claimed. Push own names only, never main or area.
