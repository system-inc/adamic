Built: nexus/consistency-no-for-in migrated from batch4 onto the unified descriptor harness, with handed-node dispatch and verbatim upstream message.
Commits: base origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898; own port commit accompanies this report.
Commands/output: lint-registry, gofmt and vet pass; TestRulesAgree, all 41 TestMutants subtests and TestOwnedWitnesses PASS; package 1358.111s.
Mutant: finding ID forIn becomes forIn-mutant; compiled source Node, emitted JavaScript and sanitized native all disagree with unchanged Go.
Uncovered: whole repository gate and old wave-03 unified migration; witness option sidecars are pending integration on this base.

The only implementation edits are in this rule directory. The descriptor listens to ForInStatement and sets node: true. visit receives ParseNode and uses its end, without a kind string comparison or refetch for relevance. The shared context supplies token start and UTF-16 to byte conversion. The upstream rule emits no fixes or suggestions; the port does likewise. The message is copied verbatim from cohere/policy/messages/consistency-no-for-in.json and independently checked against it before commit. No batch helpers are needed after replacing its text lookup with the owned message constant and its kind dispatch with the descriptor.

The ledger source is origin/codex/stage1-lint-batch4:stage1/cohere/lint/nexus_consistency_no_for_in.ts. The upstream prefix TestConsistencyNoForIn captures all three actual functions: TestConsistencyNoForInReportsEveryForIn, TestConsistencyNoForInStaysSilent and TestConsistencyNoForInProposesNoFix. The shared run captures 2035 unique source/rule/options combinations. Its complete corpus comparison matches 13,072,525 bytes across Go, source Node, emitted JavaScript and ASan/UBSan native. Witness comparison matches 100,952 bytes. Inherited malformed parser-recovery cases remain the shared harness's explicit refusal checks, unchanged by this port; they are not successful lint parity claims.

Upstream explicitly has no options type or registration decoder: its Run ignores options. This adapter parses field 5 into json.RawMessage and returns a non-nil decoded payload for non-null options, preserving upstream behavior and satisfying the shared guard. The first version wrongly rejected unrelated keys on generated all-rule rows and failed with json: unknown field allowLoop; the complete failed run is retained. The corrected full run exercises these nonempty option rows and passes. empty-options.options.json is present for the witness-options integration, but this area base does not yet load sidecars, so no sidecar coverage is claimed.

After sourcing /workspace/adamic-tools/env.sh:

```sh
gofmt -w stage1/cohere/lint/rules/nexus-consistency-no-for-in/oracle.go
go run ./cmd/lint-registry
go vet ./stage1/cohere/lint/...
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -timeout 40m -v
```

All command output was redirected to logs. Successful TestRulesAgree: 82.82s; complete package: 1358.111s; TestOwnedWitnesses: 45.03s. All 41 mutants passed. The owned mutation changes only the finding ID; its process compiles and runs on every backend, and comparison catches the difference at case 30 line 525. Exact outputs and all inherited mutation results are retained compressed under evidence/, with uncompressed SHA-256 hashes. The production source has no mutation active.

Step 1 limitation: codex/typeaware-wave-03 now has a pushed PARKED.md naming react/boolean-prop-naming and the two rules stopped by the private combined React driver, with the dynamic regex reproducer. The other twelve older ports are outside unified registry discovery; their private checker green does not certify them on this harness. A passing-only landing branch for those ports was not produced. Migrating their checker dependencies is separate unfinished work; this new branch deliberately starts from lint area and carries none of that blocked driver. No main or area branch was pushed.
