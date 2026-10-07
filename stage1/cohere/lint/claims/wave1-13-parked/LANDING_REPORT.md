Built: six retained rule ports on origin/area/stage1-lint d3a37422c, with the two blocked Nexus descriptors removed and named in PARKED.md.
Commits: prior branch d6290b64c retains both parked implementations and earlier evidence; this commit publishes only the six landing ports.
Commands: lint-registry, vet and unfiltered TestRulesAgree, TestMutants and TestOwnedWitnesses pass; exact stdout/stderr retained beside this report.
Mutants: all 46 registered semantic mutants compile, finish cleanly and are caught against Go on source Node, emitted JavaScript and sanitized native, including the six retained owned mutants.
Not covered: the two parked Nexus rules, shared explicitly unsupported parser recovery cases, other required correctness packages, new throughput and full gate.

Retained owned rules: default-case-last, for-direction, guard-for-in, no-constructor-return, no-delete-var, no-eq-null. No shared files are edited. No main or area ref is pushed.

Commands, from the repository root with source /workspace/adamic-tools/env.sh:

```
go run ./cmd/lint-registry > /tmp/wave13-land-registry.log 2>&1
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -v -timeout=20m > /tmp/wave13-land-tests.log 2>&1
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry > /tmp/wave13-land-vet.log 2>&1
```

TestRulesAgree captures 2,197 unique source/rule/options cases. Its parser recovery limits are explicit in the shared test log, not disabled or hidden by this unit. All selected tests run without skips.
