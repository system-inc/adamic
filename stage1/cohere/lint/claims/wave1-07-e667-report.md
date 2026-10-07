Integrated area e667e3e1 with current main 71d7e491 ancestry; owned rule sources unchanged.
Validated implementation 5a0ee758286a70aa59d52a4ea5483d3d482d773a; final evidence commit pushed only to codex/wave1-07.
Fresh witnesses, rule parity and 4 owned mutants PASS 309.460s.
All 4 semantic mutants caught by independent Go comparison on source Node, emitted JavaScript and sanitized native.
Full gate, other required-input packages, standalone profiles, throughput and setup not repeated; no new helper claim.

Fetched all origin heads. Area e667e3e1dbdfd1b9125c3256961bfbc8ec31946b adds upstream-authored path-preserving fixture copies and several automatic edits per finding. Main remains 71d7e491b3c9724f7a0e2ee754592149e7f9790b. Clean area merge; no authored shared harness, registry, oracle or compiler changes. DEDUP_LEDGER.md unchanged. Winning copies retained, losing copies absent. The option-bearing witnesses and handed-node declarations from the previous correction remain unchanged.

Command sourced /workspace/adamic-tools/env.sh; all output directly to committed tests.log:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave07-typescript go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestCompilerAndStage1Agree|TestMutants)$/^(constructor-property-name-ignored|ordering-relations-ignored|and-left-undefined-ignored|parenthesized alias lost)$' -count=1 -v -timeout 20m

Mutants: parenthesized alias lost, constructor-property-name-ignored, ordering-relations-ignored, and-left-undefined-ignored. Every mutant executes successfully and differs from independent upstream Go on all three runtimes. Findings, ranges, messages, fixes and suggestions are compared byte for byte. Inherited explicit method-signature-style parser-recovery exclusions remain outside parity coverage; no guard was relaxed or deleted. Required compiler input was supplied; 403 files pass with no missing-input skip.

    lint_test.go:443: Go, Node, emitted JavaScript, native identical: 13162339 bytes
--- PASS: TestRulesAgree (66.82s)
    lint_test.go:519: compiler and stage1: 403 files
    lint_test.go:520: Go, Node, emitted JavaScript, native identical: 20626096 bytes
--- PASS: TestCompilerAndStage1Agree (92.68s)
--- PASS: TestMutants (119.32s)
    --- PASS: TestMutants/parenthesized_alias_lost (29.37s)
    --- PASS: TestMutants/constructor-property-name-ignored (29.77s)
    --- PASS: TestMutants/ordering-relations-ignored (30.21s)
    --- PASS: TestMutants/and-left-undefined-ignored (27.41s)
    registration_test.go:45: Go, Node, emitted JavaScript, native identical: 127189 bytes
--- PASS: TestOwnedWitnesses (30.62s)

consistent-return still lacks compatible Judge/EndReachable integration through shared RuleContext. constructor-super remains reserved and unimplemented, without attributing it to that same missing API. No new rule or helper claimed. Both existing branches re-greened and pushed to their own names; never main or area. Logs preserve the actual captured paths and exclusions.
