# no-proto unified port

Ported the ledger's batch2:no_proto.ts onto the unified registry from origin/area/stage1-lint. Only this descriptor directory changes. The descriptor uses node: true and subscribes to PropertyAccessExpression and ElementAccessExpression. The entry uses the supplied node; only property-key children are fetched. The diagnostic message and id are verbatim. Static bracket keys unwrap parentheses as pinned Go property.AccessedName does. No fixes or suggestions are produced upstream, and their empty wire representations match.

Upstream has no options type or configurable behavior. The adapter decodes and retains any supplied JSON instead of dropping it; empty/default/null rows use upstream defaults. No shared generator, harness, RuleContext or dispatch files were edited.

The upstreamTest prefix TestNoProto captures all six actual functions in core/no_proto_test.go: TestNoProtoFires, TestNoProtoStaysSilent, TestNoProtoReportsTheWholeAccess, TestNoProtoReportsWhyItMatters, TestNoProtoDeclinesShapesTheCorpusOmits and TestNoProtoReportsThroughOptionalChaining. The shared capture filters by registered rule name. Witnesses cover reads/writes, quotes, templates, escapes, optional chains, nested parentheses, quiet dynamic keys/interpolated templates/object literals, and an astral character before the finding.

## Verification

All commands ran after sourcing /workspace/adamic-tools/env.sh. Test output was written directly to logs.

- bash cloud/setup.sh: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, cache warm 47s, total 47s. nproc: 5; cgroup CPU quota: 4 CPUs.
- go run ./cmd/lint-registry: passed; generated files remain ignored.
- gofmt on oracle.go: passed.
- go vet ./stage1/cohere/lint/...: passed.
- go test -v -count=1 -timeout=30m ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$': PASS, 823.372s. TestRulesAgree covered 2,031 unique upstream source/rule/options combinations; 13,057,439 output bytes matched Go, source Node, emitted JavaScript and ASan/UBSan native. All 41 registered mutants were caught. TestOwnedWitnesses matched 100,110 output bytes, including selected-rule and all-rule runs for both owned witnesses.
- ADAMIC_TYPESCRIPT_SOURCE=/tmp/no-proto-typescript ADAMIC_LINT_BENCH=1 go test -v -count=1 -timeout=20m ./stage1/cohere/lint -run '^(TestCompilerAndStage1Agree|TestThroughput)$': PASS, 120.214s. The checkout was fetched at the harness's exact pinned TypeScript commit. Compiler/repository parity covered 397 files and 20,439,895 output bytes on all four implementations.

The owned mutant changes the key comparison from __proto__ to __never_proto__. It compiles and executes, and Go comparison catches its missing findings on source Node, emitted JavaScript and sanitized native. Its full differences are retained in evidence/unified.log.

Throughput is for all 41 registered rules on 77 compiler files, not isolated no-proto cost. Best of five: Go 0.770242s, native 2.771740s, Node 1.816030s; each reported 15,027 findings. Native is 3.60 times Go's time on this machine. Builds are excluded and the timing binary is unsanitized.

## Limits and parked work

All 18 previous wave-15 rules are parked with raw source reproducers in stage1/cohere/typeaware/claims/wave-15-parked.md on codex/typeaware-wave-15. None had complete unified certification, so there was no passing owned subset to land in step 1. This branch carries none of that blocked implementation or its private harness.

No no-proto blocker was encountered. The full repository gate and its other 17 correctness-input checks were not run. TestRulesAgree retains the inherited explicit refusal tests for malformed method-signature recovery; these are not no-proto cases and were neither skipped nor relaxed. No checker bridge handles are used by this syntax-only rule. The witnesses have no option sidecar because upstream no-proto has no options. Detailed logs are in evidence/.
