Ported no-empty-static-block from batch 2 into an owned unified-registry descriptor directory.
Branch: codex/lint-port-no-empty-static-block; integration base: d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898.
Checks: registry, formatting, package and full repository vet pass; TestRulesAgree, all 41 TestMutants, and TestOwnedWitnesses pass.
Mutant: removing the comment exemption executes normally and is caught only by Go output comparisons on source Node, emitted JavaScript and sanitized native.
Not covered: the full repository test gate or separate compiler-corpus/profile tests; existing shared malformed-recovery cases remain explicit parser refusals.

Source: origin/codex/stage1-lint-batch2 at 6700c572ee3810524e356af7f92973e38ee5265f, stage1/cohere/lint/no_empty_static_block.ts. The 401-character description and message id are verbatim upstream. Cohere pin: 715ba94f3608a6500086b1076ce5cb7e51b836db.

The descriptor subscribes only to ClassStaticBlockDeclaration and takes the handed ParseNode with node: true. It reads that node's block child and reuses RuleContext.comment for parser-trivia comment handling. It adds no shared dispatch, registration, corpus, helper or harness edits. New Adamic modules are rule.a and messages.a. Upstream deliberately provides no fix or suggestion, which the comparisons preserve.

Upstream has no configurable options or named options type. The Go adapter explicitly unmarshals manifest field 5 into an empty object and returns a non-nil value even for default or null input. It preserves the options guard. TestNoEmptyStaticBlock is the prefix of both real upstream tests, TestNoEmptyStaticBlockFires and TestNoEmptyStaticBlockStaysSilent; the independent prefix run passes all nine cases, three firing and six clean.

The owned witness contains one accidental empty static block, block- and line-comment deliberate no-ops, a nonempty static block, and a comment marker in an earlier string. The mutant replaces !this.context.comment(body) with true. All three execution paths exit normally with empty stderr and differ from Go on the extra finding at the comment-explained static block, case 145 line 2766. It is not caught by compilation or sanitizers.

Validation observations: TestRulesAgree passes in 170.02s on 13,059,280 identical bytes. TestMutants passes in 1079.83s for all 41 discovered descriptors. TestOwnedWitnesses passes in 46.61s on 91,530 identical bytes under selected-rule and all-rule rows. The whole requested three-test run passes in 1296.569s without skips. These comparisons use unchanged upstream Go, source Node, Adamic-emitted JavaScript and ASan/UBSan native. The independent upstream prefix run passes in 0.055s. Existing malformed recovery refusals in other rules are separately checked by the unchanged harness and are not claimed as parser parity.

Commands, with every output written directly to /tmp/static-block-*.log:

```sh
source /workspace/adamic-tools/env.sh
gofmt -w stage1/cohere/lint/rules/no-empty-static-block/oracle.go
go run ./cmd/lint-registry
gofmt -l cmd internal stage1/cohere/lint/rules/no-empty-static-block/oracle.go
go vet ./stage1/cohere/lint/...
go vet ./...
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$|^TestMutants$|^TestRulesAgree$' -count=1 -timeout=30m -v
# From cohere:
go test ./internal/lint/rules/core -run '^TestNoEmptyStaticBlock' -count=1 -timeout=10m -v
```

Evidence files preserve complete log bytes with deterministic gzip; logs.json records uncompressed byte lengths and hashes. Only this rule directory is committed. No pull request, main push or area push.
