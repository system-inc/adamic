Integrated current area b28757f33 while retaining current main 71d7e491 ancestry; seven canonical rule directories unchanged.
Validated implementation fc7cbe4c3718371e876f3be165768bcb3ebbefc5; evidence commit follows on codex/lint-wave1-06-land.
TestRulesAgree and TestOwnedWitnesses PASS 106.234s; TestNodeTableIsLinkOnly PASS 32.882s.
Seven previously caught mutants remain recorded in wave1-06-land-main-report.md; no fresh mutant run is claimed.
Full gate, setup, throughput, vet and compiler corpus not repeated here; no new helper claimed.

Fetched all origin heads. Area brings upstream fixes for reparsing fixes under the original file path, exhausted fix-pass serialization, and the link-only node table guard. DEDUP_LEDGER.md unchanged. Clean merge with no authored shared-file changes. Main unchanged. All seven retained rules remain present; losing copies and duplicated baseline rules remain absent.

Commands sourced /workspace/adamic-tools/env.sh, with output directly to the committed logs:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave07-typescript go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout 15m
go test ./stage1/cohere/lint -run '^TestNodeTableIsLinkOnly$' -count=1 -v -timeout 10m

Rules: 13,188,475 identical bytes, 72.26s. Positive owned witnesses: 120,417 identical bytes, 33.95s. Independent Go, source Node, emitted JavaScript and sanitized native agree on findings and fixes. Link guard: 2,559 rows and 13,206,430 identical bytes with and without unattached node rows. Inherited explicit parser-recovery exclusions remain outside parity coverage. No gate was relaxed. Earlier mutant evidence is historical, not a fresh result for the changed shared harness. Original wave1-07 consistent-return still lacks Judge/EndReachable integration; constructor-super remains unfinished. No claim beyond those prerequisites is made. Push only the own landing branch; never main or area.
