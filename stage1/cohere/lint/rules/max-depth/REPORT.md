Built: max-depth from batch4 on the unified harness, with .a modules and a node-aware descriptor.
Commits: based on lint integration d3a37422c; final branch SHA is reported after push.
Commands: lint-registry, gofmt and vet pass; TestRulesAgree, TestOwnedWitnesses and the owned TestMutants subtest pass.
Mutant: extending the keyword span to node.end compiled and ran cleanly, then failed byte comparison on all three backends.
Uncovered: full repository gate and unrelated mutants were not rerun; existing malformed-parser recovery cases remain explicit refusals.

The source is origin/codex/stage1-lint-batch4:stage1/cohere/lint/max_depth.ts,
as named by DEDUP_LEDGER.md. The shared driver now dispatches only the nine
counted statement kinds and hands the existing ParseNode to visit. There is no
SourceFile recursion or shared registration/harness change. The rule preserves
else-if exemption, reset at functions, methods/accessors/constructors and static
blocks, configured maximum/disabled behavior, and keyword-only spans.

The batch message was moved verbatim into messages.a. The integer decoder is a
rule-local helper; the remaining helpers already exist on RuleContext. No fix
or suggestion is invented because upstream offers neither. The Go adapter
returns the unmodified core.MaxDepth and decodes field 5 to its upstream
MaxDepthSettings with default maximum 4. Captured JSON represents decoded
settings, including the zero and Disabled distinctions.

The upstreamTest prefix TestMaxDepth includes every actual upstream function:
TestMaxDepthFires, TestMaxDepthStaysSilent, TestMaxDepthSpansAndMessages,
TestMaxDepthDecoderReadsEveryOptionShape and
TestMaxDepthClassMembersAreEachTheirOwnFrame. Their asserted captured cases enter
the real shared comparison. The owned UTF-8 witness has five nested constructs
and reports at the default limit, so no pending witness-options harness update
is required. Its only finding covers the switch keyword.

Commands, with outputs redirected to evidence logs:

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
gofmt -w stage1/cohere/lint/rules/max-depth/oracle.go
gofmt -l stage1/cohere/lint/rules/max-depth/oracle.go
go run ./cmd/lint-registry
go vet ./stage1/cohere/lint/...
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$|^TestMutants$/^max-depth-keyword-span$' -count=1 -v -timeout=30m
```

TestRulesAgree compares 13,067,617 bytes between production Go cohere, source
Node, emitted JavaScript on Node and sanitized native, and passes. The harness
also checks existing unsupported-recovery refusals belonging to other rules;
these are not max-depth blockers or certification of parser recovery. The
TestMutants filter runs only this unit's mutant. execute rejects nonzero exits
or any stderr, so its three caught differences are comparison-only failures.
TestOwnedWitnesses checks every descriptor witness in selected and all-rule
modes. Setup took 132s; nproc is 5. Full output is in evidence/gate.log.
