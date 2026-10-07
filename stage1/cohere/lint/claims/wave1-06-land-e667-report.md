Integrated area e667e3e1 with current main 71d7e491 ancestry; owned rule sources unchanged.
Validated implementation 01da2c81e7ed6b31bd28f7335a4c4765196816f5; final evidence commit pushed only to codex/wave1-06-land.
Fresh witnesses, rule parity and 7 owned mutants PASS 320.862s.
All 7 semantic mutants caught by independent Go comparison on source Node, emitted JavaScript and sanitized native.
Full gate, other required-input packages, standalone profiles, throughput and setup not repeated; no new helper claim.

Fetched all origin heads. Area e667e3e1dbdfd1b9125c3256961bfbc8ec31946b adds upstream-authored path-preserving fixture copies and several automatic edits per finding. Main remains 71d7e491b3c9724f7a0e2ee754592149e7f9790b. Clean area merge; no authored shared harness, registry, oracle or compiler changes. DEDUP_LEDGER.md unchanged. Winning copies retained, losing copies absent. The option-bearing witnesses and handed-node declarations from the previous correction remain unchanged.

Command sourced /workspace/adamic-tools/env.sh; all output directly to committed tests.log:
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$/^(wrong-diagnostic-id(#01)?|wrong-global-type-name|omit-empty-export-fix|wrong-enclosing-scope|configured-restriction-ignored|report-left-reference)$' -count=1 -v -timeout 20m

Mutants: wrong-diagnostic-id (beforeInteractive), wrong-diagnostic-id#01 (CSS tags), wrong-global-type-name, omit-empty-export-fix, wrong-enclosing-scope, configured-restriction-ignored, report-left-reference. Every mutant executes successfully and differs from independent upstream Go on all three runtimes. Findings, ranges, messages, fixes and suggestions are compared byte for byte. Inherited explicit method-signature-style parser-recovery exclusions remain outside parity coverage; no guard was relaxed or deleted. Required compiler corpus validated separately on wave1-07; not rerun on this branch.

    lint_test.go:443: Go, Node, emitted JavaScript, native identical: 13203441 bytes
--- PASS: TestRulesAgree (71.70s)
--- PASS: TestMutants (214.96s)
    --- PASS: TestMutants/wrong-diagnostic-id (28.74s)
    --- PASS: TestMutants/wrong-diagnostic-id#01 (29.60s)
    --- PASS: TestMutants/wrong-global-type-name (29.16s)
    --- PASS: TestMutants/omit-empty-export-fix (31.18s)
    --- PASS: TestMutants/wrong-enclosing-scope (32.63s)
    --- PASS: TestMutants/configured-restriction-ignored (32.49s)
    --- PASS: TestMutants/report-left-reference (28.41s)
    registration_test.go:45: Go, Node, emitted JavaScript, native identical: 120417 bytes
--- PASS: TestOwnedWitnesses (34.17s)

consistent-return still lacks compatible Judge/EndReachable integration through shared RuleContext. constructor-super remains reserved and unimplemented, without attributing it to that same missing API. No new rule or helper claimed. Both existing branches re-greened and pushed to their own names; never main or area. Logs preserve the actual captured paths and exclusions.
