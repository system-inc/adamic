# Adjacent overload signatures

Migrated batch4-typescript's adjacent_overload_signatures.ts into the unified registry on area base d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898. Only this rule directory changes. The descriptor subscribes to seven container kinds and hands the parsed node to the rule; the rule does not refetch or redispatch the container. Messages are verbatim in messages.a. Identifier validation advances by Unicode code point, including supplementary-plane characters.

The upstream rule has no configurable options or fixes or suggestions. Its oracle adapter decodes the manifest JSON into the empty options schema; it does not return nil. No options sidecar is required for its firing witness.

The upstreamTest prefix TestAdjacentOverloadSignatures includes all five real upstream functions: StaysSilentOnUpstreamPassCases, FiresOnUpstreamFailCases, DiscriminatesOnCasesUpstreamDoesNotWrite, NamesTheMemberUpstreamNames, and SurvivesContainersWithNoMemberList.

Validation commands (toolchain sourced from /workspace/adamic-tools/env.sh):

- `go run ./cmd/lint-registry`: passed, registry.log.
- `gofmt -w stage1/cohere/lint/rules/typescript-adjacent-overload-signatures/oracle.go`: completed; formatting has no remaining diff.
- `go vet ./stage1/cohere/lint/...`: passed, vet.log (empty output).
- `go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -timeout=20m -v`: passed, parity.log (13,098,230 identical bytes; 49.80 seconds).
- `go test ./stage1/cohere/lint -run '^TestMutants$/adjacent_overload_separation_ignored$|^TestOwnedWitnesses$' -count=1 -timeout=20m -v`: passed, owned.log.

The witness reports a separated function overload. The mutant reverses `previous !== key` to `previous === key`; the Go comparison catches its missing finding on source JavaScript, emitted JavaScript and sanitized native execution. The full owned witness suite compares 90,584 bytes identically on all four sides.

The initial combined run passed TestRulesAgree (2,124 captured combinations, 13,096,402 identical bytes), then was interrupted during unrelated rules' mutation tests. The dedicated commands above replace that interrupted run. The full mutation suite was not completed; the new rule's mutant passed. No shared harness files or guards changed. Existing explicit malformed-parser recovery limits for method-signature-style remain recorded by the shared parity test; none belong to this rule. No type checker, dynamic RegExp, parser or Tailwind blocker was encountered by this port. No broader repository gate or throughput benchmark was run for this unit.
