Ported no-self-compare to the unified harness as an owned Adamic listener.
Base: d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898; this branch contains no parked checker rules.
Registry, gofmt, vet, TestOwnedWitnesses and TestRulesAgree PASS; assigned TestMutants subtest PASS.
The equality-reversed mutant compiled and was caught by Go comparisons on Node, emitted JavaScript and sanitized native.
Not covered: unrelated mutant subtests, full repository gate, or dedicated throughput profiling.

The batch2 message is copied verbatim. The descriptor hands the BinaryExpression directly to visit; the operator is its second child. Operand signatures preserve punctuation and literal contents while ignoring trivia, using RuleContext.signature. Empty array/object literals receive the same special case as upstream hasSameTokens. No checker, regular expression engine, shared helper, dispatcher or harness change is required.

The upstream prefix TestNoSelfCompare includes every real function: TestNoSelfCompareStaysSilent, TestNoSelfCompareFires, TestNoSelfCompareDeclinesInAndInstanceof, and TestNoSelfCompareReportsTheWholeComparison. Upstream has no options struct or configurable settings. The adapter decodes field 5 as JSON when supplied, preserving its value instead of returning nil for an options-bearing row; default rows return upstream's nil default.

Witnesses cover a reporting comparison, whitespace-different call chains, empty arrays, literal comment contents, differing operands, in/instanceof exclusions, and private names. Shared TestOwnedWitnesses verifies both selected and all-rule runs against unchanged Go. TestRulesAgree matched 13,061,920 bytes on all four backends (71.38 seconds); TestOwnedWitnesses matched 94,306 bytes (32.55 seconds). The assigned semantic mutant passed its catch test in 27.96 seconds. Native is built with ASan/UBSan by the unchanged harness.

An initial run began before the operator-child correction and failed agreement. It was stopped and superseded by the final agreement/witness run and assigned mutant run recorded here. Existing malformed method-signature fixtures retain the harness's explicit parser-recovery refusal checks; no skip or recovery exception was added for this rule.

Setup: bash cloud/setup.sh completed in 88s; cumulative lines were Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s and build cache warm 88s. nproc=5; cgroup quota=4 processors. Sourced /workspace/adamic-tools/env.sh. Removed completed private-gate ELF/archive binaries to recover disk space, preserving all source fixtures and recorded evidence.

Commands, from the repository root after sourcing the toolchain:

```sh
go run ./cmd/lint-registry > /tmp/no-self-compare-registry-final.log 2>&1
gofmt -l stage1/cohere/lint/rules/no-self-compare/oracle.go > /tmp/no-self-compare-gofmt-final.log
go vet ./cmd/lint-registry ./stage1/cohere/lint/... > /tmp/no-self-compare-vet-final.log 2>&1
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout 30m > /tmp/no-self-compare-agreement-final.log 2>&1
go test ./stage1/cohere/lint -run '^TestMutants/self-comparison-equality-reversed$' -count=1 -v -timeout 10m > /tmp/no-self-compare-mutant-final.log 2>&1
```

Exact final logs are in evidence/*.txt. Only this descriptor directory is changed. The parked private checker analyses and their reproducers remain in stage1/cohere/typeaware/claims/wave-20-parked.md on codex/typeaware-wave-20; all existing private analyses are outside this landing branch.
