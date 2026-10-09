# Ambiguous identifier port

Ports `nexus/consistency-no-ambiguous-identifier` from batch 3 (`fa9781c2c0911c52312002ba97ffd4b560d178ae`, `stage1/cohere/lint/rules/no_ambiguous_identifier.ts`) to the unified descriptor harness. Messages are verbatim from the upstream policy. The Identifier listener receives its node directly. The shared binding-ownership helper lives on RuleContext; comparator and event helpers remain in this rule directory. Regex literals use the shared translation table, without stateful global matching.

The upstream prefix `TestConsistencyNoAmbiguousIdentifier` captures all six real tests: Fires, StaysSilent, InfersContext, ProposesNoFix, ForeignNames and JudgesThisProperty. Upstream has no options struct and ignores options; the adapter nevertheless decodes field 5 as JSON and preserves its value. The firing witness supplies `{}` through the options sidecar. Witness-options support was adopted from the integration branch rather than editing the harness.

## Verification

Commands used after sourcing `/workspace/adamic-tools/env.sh`; each test wrote directly to its corresponding log (compressed copies in `evidence/`):

```sh
go run ./cmd/lint-registry
gofmt -w stage1/cohere/lint/rules/nexus-consistency-no-ambiguous-identifier/oracle.go
go vet ./stage1/cohere/lint/...
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=30m
go test ./stage1/cohere/lint -run '^TestMutants$/^ambiguous-identifier-span$' -count=1 -v -timeout=30m
go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=30m
```

All passed. Owned witnesses matched 104,680 bytes on Go, Node, emitted JavaScript and sanitized native (31.670 seconds). RulesAgree captured 2,050 unique source/rule/options combinations and matched 13,199,345 bytes (102.123 seconds). The end-position +1 mutant compiled and ran, then the Go comparison caught it on Node, emitted JavaScript and native (21.550 seconds). Only this rule's mutant was selected.

The harness retains its inherited explicit malformed-source parser-recovery limits for method-signature-style; no limits were added or relaxed. The full repository gate and unrelated mutants were not run. Test durations are whole-test observations, not a native-versus-Go benchmark.
