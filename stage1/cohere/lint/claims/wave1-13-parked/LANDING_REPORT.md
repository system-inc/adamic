Built: seven retained rule ports on shared harness fix 4f18a05c9, with only the parser-blocked delay descriptor removed and named in PARKED.md.
Commits: prior branch d6290b64c retains both parked implementations and earlier evidence; this publication restores return-void after the shared JSX fix and retains the six other landing ports.
Commands: lint-registry, vet and unfiltered TestRulesAgree, TestMutants and TestOwnedWitnesses pass; exact stdout/stderr retained beside this report.
Mutants: all 46 registered semantic mutants compile, finish cleanly and are caught against Go on source Node, emitted JavaScript and sanitized native, including the six retained owned mutants.
Not covered: the still-parked Nexus delay rule, shared explicitly unsupported parser recovery cases, other required correctness packages, new throughput and full gate.

Retained owned rules: default-case-last, for-direction, guard-for-in, no-constructor-return, no-delete-var, no-eq-null. No shared files are edited. No main or area ref is pushed.

Commands, from the repository root with source /workspace/adamic-tools/env.sh:

```
go run ./cmd/lint-registry > /tmp/wave13-land-registry.log 2>&1
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -v -timeout=20m > /tmp/wave13-land-tests.log 2>&1
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry > /tmp/wave13-land-vet.log 2>&1
```

TestRulesAgree captures 2,197 unique source/rule/options cases. Its parser recovery limits are explicit in the shared test log, not disabled or hidden by this unit. All selected tests run without skips.

## Shared fixed-source repair landing

Rebased onto 4f18a05c9ee0fc3c17a3a769329bb5c705875027 as explicitly requested. It includes the witness-options integration and keeps both the Go oracle and the port's unconverged fix-engine answer. Shared files are inherited unchanged. Restored nexus/consistency-no-return-void and added the exact formerly failing JSX source as a firing witness. Only nexus/consistency-no-hand-rolled-delay remains parked for its top-level-await parser failure.

Fresh commands use the setup environment and direct log redirection:

```
go run ./cmd/lint-registry > /tmp/wave13-fixed-registry.log 2>&1
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$/(default-case-last-semantic-mutant|for-direction-semantic-mutant|guard-for-in-semantic-mutant|no-constructor-return-semantic-mutant|no-delete-var-semantic-mutant|no-eq-null-semantic-mutant|return-void-unicode-whitespace-mutant)$' -count=1 -v -timeout=20m > /tmp/wave13-fixed-tests.log 2>&1
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry > /tmp/wave13-fixed-vet.log 2>&1
```

All three selected tests pass on this base, including all seven owned mutants caught against Go in all three execution modes. Witnesses match 140,033 bytes. No selected test skips. Complete output is fixed-tests.log, fixed-registry.log and fixed-vet.log. The earlier unfiltered sweep of all 46 registry mutants is retained in landing-tests.log; it predates this shared fix. The bb2ece564 witness-options revalidation is retained separately as options-tests.log, options-registry.log and options-vet.log. Other correctness packages, fresh throughput and full gate are not claimed.
