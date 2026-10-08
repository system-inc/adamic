# Unit 8 stopped on the shared terminal schema

Base: 9addb0e8a48b1cf60621ae24ef2196c860ff33ce, origin/stage1-hir/wip.
Branch: hir/unit-8. No reactive pass implementation or census certificate is claimed.

## Observed blocker

The Go builder consumes Scope at reactive_build.go:693, Sequence at :725 and
MaybeThrow at :729. These are missing from the shared TerminalType at
stage1/cohere/high_level_intermediate_representation/core.ts:89 and from
replay/decode.ts:113. The existing shared decoder panics at :134 for each.

The smallest replay inputs live in replay_gap.a:11, :13 and :15. The function
owns three checked block handles; the input refers to its second and third
blocks. These are terminal-schema probes, not census records or inferred scopes.
The test imports ../../replay/index.ts and executes all three inputs on source
Node, emitted JavaScript and sanitized native. All nine runs produce the exact
missing-terminal panic and exit 70. evidence/schema.log records the observations.
The test fails when any input stops producing that refusal, so a later owner fix
cannot silently leave this stopped note looking current.

## Request to hir-01

Add these Go terminal variants to the shared typed graph, decoder, encoder,
copy/remap and visitor interfaces, preserving their exact fields and evaluation
orders:

* Scope: ScopeId, Block, Fallthrough, Order; terminal.go:324.
  ScopeId must resolve through the checkpoint identity reader to ScopeIndex;
  Block and Fallthrough must resolve to checked BlockIndex handles.
* Sequence: Block, Fallthrough, Order; terminal.go:225.
* MaybeThrow: Continuation, Handler, Order; terminal.go:285.

The referenced Go paths are under
cohere/internal/lint/ecmascript/high_level_intermediate_representation/.
Unit 8 will own its reactive-tree types, sidecar codec, builder, visitors and
transforms. It needs no new arena class: ReactiveIndex and ScopeIndex already
exist. No shared file was edited and no private terminal implementation was
invented. The owner must supply the shared variants before an isolated Go scope
checkpoint can be decoded as the required ConstructedHIR input.

## Certificate and limits

Pass comparison: native 0/1,465; source Node 0/1,465; emitted JavaScript 0/1,465.
Flow pass replay: 0/23. The nine successful refusal observations above are not
successful pass comparisons. No semantic pass mutant ran, because no pass was
implemented. No input or error path was sampled or counted as a pass.

Commands:

```
source /workspace/adamic-tools/env.sh
go test -v -count=1 -timeout 20m ./stage1/cohere/high_level_intermediate_representation/passes/unit-8
gofmt -l stage1/cohere/high_level_intermediate_representation/passes/unit-8/replay_gap_test.go
go vet ./stage1/cohere/high_level_intermediate_representation/passes/unit-8
```

The gap test passed in 27.332s. Test output is in evidence/schema.log.
The initial emitted-JavaScript run omitted oracle/node.mjs and could not resolve
adamic; the test was corrected to use the repository runtime loader and rerun.
This was a test-launch error, not a language gap or an owner blocker.

Setup passed in 40.649s; go ready 0.022s, Node ready 0.024s, submodules ready
0.063s, Markdown ready 0.072s, clang ready 0.162s, Go build ready 40.474s.
nproc is 5, cgroup cpu.max is 400000 100000. evidence/setup.log retains the
complete timing lines. The full HIR package and all 1,465 pass comparisons were
not run because the required prepass terminal representation is unavailable.
