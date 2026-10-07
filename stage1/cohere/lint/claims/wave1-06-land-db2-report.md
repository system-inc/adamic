Integrated current area db2ecc004 and retained current main 71d7e491 ancestry; owned rules unchanged.
Validated source ce9338060e44fd7c18c23897876a2f2b35b17250; final evidence tip pushed on codex/lint-wave1-06-land only.
Fresh TestRulesAgree, seven owned TestMutants and TestOwnedWitnesses PASS 310.361s.
All seven mutants caught by independent Go output comparison on source Node, emitted JavaScript and sanitized native.
Full gate, full mutant registry, sharded mode, required compiler corpus, standalone profiles, setup and throughput not rerun on this branch.

Area db2ecc00447f9ebe8adecb190f71ac222e5db860 imports the next rule batch and parallelizes TestMutants. Main remains 71d7e491b3c9724f7a0e2ee754592149e7f9790b. Clean merge; no shared files authored. DEDUP_LEDGER.md unchanged. All seven finished winning takeover rules retained; losing copies remain absent.

Command sourced /workspace/adamic-tools/env.sh and exported GOPROXY=https://proxy.golang.org|direct; all output directly to committed tests.log:
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$/^(wrong-diagnostic-id(#01)?|wrong-global-type-name|omit-empty-export-fix|wrong-enclosing-scope|configured-restriction-ignored|report-left-reference)$' -parallel 2 -count=1 -v -timeout 20m

The filter was checked against all descriptor mutant names: it selects exactly the seven owned rules. -parallel 2 bounds simultaneous native builds without altering any assertion or witness. Each mutant executes successfully and differs from independent Go on all three runtimes. Positive owned witnesses compare 290,670 identical bytes. Expanded registry findings and fixes agree; exact byte counts and timings remain in the raw log. Inherited explicit parser-recovery exclusions remain outside parity coverage. No check relaxed or deleted. No new rule or helper claimed; unfinished wave07 work remains on its original branch only. Never push main or area.
