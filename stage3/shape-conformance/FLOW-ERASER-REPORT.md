The subsequent [latent share report](LATENT-SHARE-REPORT.md) supersedes this report's unmeasured-census status. The production implementation and validation below remain the earlier checkpoint.

Built: shared allocation-flow query, declared field certificates, and a conservative scalar-record eraser; real tsc share remains unmeasured.
Commits: lane 1 dependency merge 40ad020d; allocation flow, certificates and scalar eraser 59db5d63.
Commands: lower 35.217s, IR 28.663s, scoped oracle 38.385s, vet and all 43 graph count assertions pass.
Mutants: ignore one nonconforming shape and ignore readiness both finish with wrong output and are caught by pinned semantic assertions.
Uncovered: broader assignability, nested/container/callback flows, staged-after-store erasure, native host view metadata, per-cast lookup emission, and the 2,936-site real share.

This is a partial lane 3 implementation, not completion or a landing claim.
The required adapted-tree measurement cannot be produced by this IR pipeline:
the completed updated census has 78 checker-rejected source entries, one Refused
entry, zero accepted IR programs, and 830 diagnostics on the combined 79-root
program. Its first failure is binder.ts:1229:13, TS7029 switch fallthrough.
No frontend options or adaptations were weakened to obtain an apparent result.

The unchanged audit ledger supplies 1,758 tagged and 1,178 untagged identities.
[site-comparison.jsonl.gz](site-comparison.jsonl.gz) retains every original
file, offset, category, and before reason, with an explicit unmeasured after state.
Certified free sites remain 0 before and 0 after. This is not a measurement of
how many sites actually conform. Conforms-ready, conforms-not-ready,
conforms-if, and dynamic-unknown counts are null in
[site-comparison-summary.json](site-comparison-summary.json).
Frontend rejection is not a dynamic unknown-allocation proof. The complete
[updated census](updated-census.jsonl.gz) and its
[summary](updated-census-summary.json) preserve the evidence.

The allocation query reuses the graph-regions implementation and numbering.
Producer collection and traversal were extracted in place, not copied.
Unknown is explicit for opaque/untracked expressions, missing or omitted
producers, callback parameters and empty cycles. A joined known-plus-unknown
set retains both its known allocations and the Unknown bit. Property/element
loads and container/callback producer edges remain untracked, conservatively.
Querying does not select graph ownership. The graph-regions caller retains its
original behavior of ignoring unknown frontiers for leak-only classification.

Field certificates store the checker type identity, declared semantic type and
optionality as compile-time IR metadata. Number and boolean certificates differ
even when native layouts coincide. Contextual declarations supply typed reserved
slots: ready: undefined! in a declared boolean field certifies boolean, not never,
and still requires initialization. Classes capture available field declarations,
but class erasure remains disabled. Certificates make no initialization claim.

The enabled eraser proves exact scalar contracts on every reaching immutable
record. Exact contract identity is a conservative subset of assignability.
Any program containing property/index mutation, Object operations or closure
calls disables this subset. Classes, spreads and broader contracts retain checks.
Read checks require the existing readiness analysis to have removed their
readiness marker; additionally no reaching initializer may reserve that field.
No second readiness bitmap or analysis was introduced. A complete scalar target
also permits removal of its admission tag test. Partial scalar field proofs can
remove individual read checks while other fields retain views. Runtime identity
and receiver evaluation are preserved. Reads use the existing backend field
lookup/cache machinery; a new dedicated one-lookup-per-cast representation is
not implemented.

The small named functions and hooks are:

- newAllocationFlowGraph and allocationFlowGraph.follow: shared producer graph
  and traversal extracted from graphFlows in graph_flow.go.
- allocationFlowGraph.ReachingAllocations: may-flow query with Sites, Unknown
  and sorted reasons, implemented in shape_flow.go.
- assignAllocationSites: graph-regions numbering, reused for programs without
  cycle seeds; already numbered programs are unchanged.
- certifyAllocationFields: allocation hook from graphAllocation, recording full
  declared field types before native layout selection.
- certifiedCheckedCast: admission hook from interfaceCast, attaching lane 1's
  complete target ViewContractID.
- eraseProvenViewChecks: deferred finalization hook from readiness, applying
  only the scalar, immutable subset after shared readiness facts are available.
- rewriteShapeStatements: lane-owned IR transformation for these proof results.

Fixtures in stage3/interface-downcasts/lane3 are .a source programs:

- proven.a: allocation through factory return, parameter, return alias and cast;
  IR asserts zero CheckedCast nodes and zero checked field reads; source Node,
  native and JavaScript print hello.
- nonconforming.a: the same cast receives boolean and number payload shapes;
  the view stays and the second call stops with exit 70, naming ready, boolean
  and number. A number must not certify a boolean merely by its native layout.
- uninitialized.a: a declared boolean slot remains uninitialized at the cast;
  the view stays and stops with exit 70, naming ready and uninitialized.
- host.a: readTextFile supplies an untracked object; IR asserts retained cast
  and checked reads. Source Node prints true, but the existing native host
  object's tag lacks the view runtime's representation metadata and stops with
  exit 70: field read failed: kind is not a string; expected string, found
  unsupported representation. The failed host oracle attempt is preserved in
  logs/erasure-final.log. No host runtime agreement is claimed; the passing
  execution oracle deliberately covers the other three fixtures. The host
  count row records this failure frontier, not a successful host execution.

Source-level mutants are reproducible with eraser-mutants.py from the repository
root after sourcing the toolchain. Ignoring contract mismatch prints true then
false natively instead of failing; ignoring readiness prints false instead of
failing. Both return exit 0 with empty stderr, compile valid C, and are caught by
the ordinary fixture oracle's pinned exit/message. Independent IR mutants also
run on release native and JavaScript: malformed input produces true/false versus
true/0, and the staged input prints false. Sanitizers or compiler warnings are
not the catch. The initial readiness source mutant escaped because the staged
certificate was incorrectly never; contextual certificate capture was corrected,
a declaration assertion added, and both source mutants rerun and caught.

Validation commands (all test output captured to files):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestShapeGraphCountSnapshot$' -count=1 -v
# Same test on isolated pre-refactor dependency merge 40ad020d and candidate.
go test ./internal/lower ./internal/ir -count=1 -timeout 15m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView|^TestShapeErasureCountRows|^TestShapeGraphCountSnapshot|^TestGraphAllocationFlow' -count=1 -v
go vet ./internal/lower ./internal/ir ./internal/oracle
python3 stage3/shape-conformance/eraser-mutants.py
```

All 43 graph-region fixture count rows match the pre-refactor merge exactly:
allocations, frees, retains, releases, peak, regions, graph regions and merges.
The guard includes the million-object fixture.
[graph-counts-guard.json](graph-counts-guard.json) records every row; before and
after logs are logs/graph-counts-before.log and logs/erasure-oracle-final.log.
The explicit graph allocation oracle also verifies Node behavior, ASan/UBSan,
shared leak checking and counted teardown for return/conditional/mixed/override.
The final scoped oracle includes lane 1 object/interface/tagged-object-union tests.
Final package, oracle and vet logs are logs/erasure-packages-restored.log,
logs/erasure-oracle-final.log and logs/erasure-vet-restored.log. Mutant failures
are logs/ignore-nonconforming-shape.log and logs/ignore-readiness.log.
Observed final outputs: lower passes in 35.217s, IR in 28.663s, and the
scoped oracle in 38.385s. Vet and formatting output are empty. The graph
snapshot now asserts all 43 rows against the committed baseline automatically;
its final execution is logs/graph-counts-pinned.log.
The full repository gate was not rerun. Earlier inherited native diagnostic-byte
assertions and the central mixed-format counts table remain owner handoffs;
this lane does not claim they pass. New measured count rows are handed off in
[erasure-count-rows.md](erasure-count-rows.md), without editing the shared table.

Setup from the initial lane turn: bash cloud/setup.sh passed in 58.329s;
GOPROXY=https://proxy.golang.org|direct; environment
/workspace/adamic-tools/env.sh; nproc=5, quota=4 CPUs; Go 1.27.1, clang 20.1.8,
Node 24.19.0. nproc was rechecked as 5. Refreshed origin/main remains 71d7e491,
already an ancestor; lane 1 is merged at 609ed395. Pushes target only
codex/shape-conformance. No pull request is opened.
