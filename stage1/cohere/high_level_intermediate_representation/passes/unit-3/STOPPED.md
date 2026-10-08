# Unit 3 — stopped at the shared replay contract

Base: `origin/stage1-hir/wip` at `3714b3198397502297cd298d3d497835dcdeaa49`.
Fetched again before recording this stop; the shared branch had not advanced.

The lane owns only this directory. Its tagged Go adapter invokes each real Go
entry point and snapshots the graph immediately before and after it. The caller
supplies the actual prepass graph and the shared encoder; it never substitutes
an invented pipeline. Remapping also snapshots the nested graph on both sides.
Invocation sets are sorted by numeric function identity. Before/after encoder
errors propagate with the recorded input and result retained when available.
No construction printer is passed off as the missing full checkpoint codec.

Local verification: `go test -v -count=1 -timeout=3h
./stage1/cohere/high_level_intermediate_representation/passes/unit-3` passes.
The overlay checks all eight entry points (the two IIFE modes separately),
snapshot ownership, failures at both boundaries, and an actual Go DCE rewrite
including retention of the orphan instruction-table entry. All new top-level
Go tests call `t.Parallel`. These are adapter checks, not a pass certificate.

Certificate: Node **0/1,465**, emitted JavaScript **0/1,465**, sanitized native
**0/1,465**. Semantic pass mutants executed/caught: Node **0/0**, native **0/0**.
No census inputs or error paths were discarded. No full census fixtures have
been generated; all 23 Flow graphs still require the owner's Go-input replay.
No Adamic pass implementation is certified or represented as finished.

## Requests to hir-01

1. Land the shared checkpoint encoder/decoder and framing/identity reader in
   `replay/`, including the all-1,465-record success/error manifest and the 23
   Flow graphs. Publish the codec roundtrip certificate and corrupt-index mutant.
   Expose the canonical Go encoder to this lane's `unit3Encode` callback and a
   census observation hook that can call `unit3Observe` at each real boundary.
   Preservation's memo-drop input must be after outline and Go InferReactive;
   the cache path must retain its own separate Construct/drop/inclusive-inline/
   conditional-Construct checkpoints. Supply parent, nested and captures for
   remapping, including unsuccessful remap outcomes.
2. Add shared `DeclareContext`, `StartMemoize` and `FinishMemoize` variants and
   their visitor/dump/replay support. Preserve absent dependency arrays separately
   from empty arrays; global versus local roots, ordered optional path segments,
   memo IDs and finish values must survive framing. Provide AST/source handles,
   callee module origins and dependency facts rather than inferred answers.
3. Add outlined identifier-to-function identities to the shared function record
   and codec. Expose supported instruction-value replacement (currently readonly)
   and fresh typed copies of values/terminals for remapping, preserving source
   handles, orders, declaration equivalence classes and orphan table entries.
   Fresh identifiers using an existing declaration must not mint an extra
   declaration: Go `NewIdentifier` preserves that declaration without allocation.
4. Publish the public `ForFunctionWithoutManualMemoization` cache hook contract;
   lane-local memo cache implementation will import it. SSA and mutation_aliasing
   stay imports by reference. No shared index class is requested for the current
   lane-local maps; existing Function/Block/Instruction/Identifier indices suffice.

No owner-owned file was changed. Merge the framing update when it lands (no
rebase), then implement and certify against fresh replay arenas. The previously
reported native CloneFunction spread limitation is an integration gap, not an
excuse to exclude any record from this isolated lane. No new compiler language
gap or shortest reproducer was discovered in these Go adapter checks.
