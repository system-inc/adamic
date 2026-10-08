Built: grouped all 315 diagnostics by code and file, with per-cast dependency and unblock counts.
Commits: measured input e7f420293e94746095c2f1e2d14124484e1afb63; this report checkpoint is recorded in branch history.
Commands: diagnostic-actions.py inventory/provenance assertions pass; 315 diagnostics in 79 groups cover 2004 diagnosed sites.
Mutants: fixing one of two diagnosed producers cannot clear both; the overlapping-dependency mutant is caught.
Limits: unblock counts describe current diagnosis frontiers, not successful rechecking or free casts.

Current report: [DIAGNOSTIC-ACTIONS-REPORT.md](DIAGNOSTIC-ACTIONS-REPORT.md).

Previous checkpoints follow.

Built: record-property and constant-element allocation flow, with visible-store joins and conservative unknown frontiers.
Commits: merge checkpoint c1f4c5a70bd7d98fd543f4eb7ff96191be4a338b; implementation commit is in branch history.
Commands: full lower/IR tests, vet, 43 unchanged graph-region counts, native fixtures against Node, and independent census audits pass.
Mutants: four projection guards, six measurement guards, forged host/body proofs, and native wrong-shape/readiness erasures caught.
Limits: no tsc cast is proven free; callbacks, dynamic keys, classes, whole-array and transitive contracts remain unproven.

Current report: [PROJECTION-SHARE-REPORT.md](PROJECTION-SHARE-REPORT.md).

Previous checkpoints follow.

Built: merge main 48c05d09 with checked views, readiness and lane 3 hooks retained; fresh adapted-tree census.
Commits: previous census ffe5a90e; this merge is recorded in branch history.
Commands: full lower/IR packages, scoped uncached oracle, 43 graph-count guards, cast oracle, vet and independent census audits pass.
Mutants: native wrong-shape/readiness plus measurement type/readiness/body/provenance and host/body forgery mutants caught.
Limits: all tsc sites still need views; complete erasure and host runtime metadata remain unproven.

Current report: [MERGED-SHARE-REPORT.md](MERGED-SHARE-REPORT.md).

Previous checkpoint follows.

Built: per-function allocation-flow share over every original ledger site, measured on a checker-rejected program.
Commits: production eraser 59db5d63; lane 1 609ed395 is already merged through 40ad020d; measurement 2ab6be95.
Commands: exact AST mapping, corpus/control measurement, independent diagnostic-span audit and production graph/eraser regression pass.
Mutants: ignoring field types, readiness or diagnosed bodies is caught by independent controls; forged host/body certifications are caught by the corpus audit.
Limits: proof coverage is zero certified sites, not proof that every runtime shape fails; unsupported flows and host values retain views.

Current report: [LATENT-SHARE-REPORT.md](LATENT-SHARE-REPORT.md).

The following is the earlier production-eraser checkpoint, before the latent measurement.


Current report: [FLOW-ERASER-REPORT.md](FLOW-ERASER-REPORT.md).

The following is the earlier dependency-only checkpoint.


# Lane 3 dependency checkpoint, October 7, 2026

This branch is a checkpoint, not the requested first deliverable. No field check
has been erased by lane 3. No new shape certificate or conformance outcome is
claimed. The requested new implementation files have not been added: disconnected
helpers would not establish a proof used by compilation.

## Branch and dependencies

Started from current origin/main 71d7e491 on codex/shape-conformance and merged:

- origin/codex/interface-downcasts 55b6b4aa in e460cff4.
- origin/codex/graph-regions 198b1271 in 93a33d6c. This includes the requested
  410cb1c allocation-flow work and later pointer/map traversal fixes.
- origin/codex/non-null-check a02613ef in f88b9418, including deinitialization.

No dependency implementation was copied. Only this worker branch is a push target.
Initially docs/checked-views-plan.md was absent and lane 1 named 55b6b4aa.
A final fetch found 4dbf3b6a and 00d6006a. The plan at 4dbf3b6a was read in
full and merged. It reserves shared hooks and count regeneration to lane 1,
and requires a graph-flow extraction handoff to the allocation-flow owner.
No transitive-view implementation has landed at that tip: interface_cast.go is
unchanged from 55b6b4aa. Its view entry point supports scalar contracts and
marks fields globally.

Dependency integration exposed semantic conflicts beyond Git's conflict markers:

1. Graph regions skipped every uninitialized declaration. Readiness requires real
   local storage and flags. Native fixtures first failed clang with undeclared
   variables; JavaScript then failed independently pinned diagnostics.
2. Checked views add initialization and representation bytes per object field.
   Graph adoption copied only the value slots, truncating that metadata. Graph
   fixtures crashed. Adoption and diagnostic size accounting now include both
   bytes.
3. Nested forward declarations also use Declare.Uninitialized, but their locals
   are not readiness slots. Emitting those placeholders as ordinary declarations
   duplicated C declarations. The merged backends distinguish them through the
   existing Local.Uninitialized flag, preserving graph environment allocation.

41f14055, 92944877 and 3985ee2c reconcile these dependency behaviors. No eraser
behavior was added during these repairs. The failing runs are retained; failures
at clang are not counted as mutant kills.

## Exact baseline, not a new conforms measurement

The checked-in audit's 2,936 rows were independently matched to the 4,101-row
assertions ledger by file/start/end and exact text. Unique selected keys: 2,936.
Ledger SHA256:
15739c81ad33071014c88e0b4514b08c7562cb3191d0bd20e7df321fc2bc1cab.

| Existing certified outcome | Tagged | Untagged | Total |
| --- | ---: | ---: | ---: |
| Conforms | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 |
| Never | 0 | 0 | 0 |
| Undecidable | 1,758 | 1,178 | 2,936 |

These are the prior audit's outcomes. The new after measurement is **unavailable**,
not zero conforms and not a newly measured claim that every allocation is unknown.
The prior audit does not separate conforms-ready from conforms-not-ready because
it has no complete type certificates. Its source API audits pristine TypeScript
050880ce, not the adapted tree. It cannot be rerun on adapted text with unchanged
ledger offsets and called a new measurement.

[checkpoint.json](checkpoint.json) records the verified baseline and null after
measurement. The original per-site rows and reason classifications remain in
[the existing table](../interface-downcasts/shape-conformance-sites.json).
Combined baseline reasons are 2,127 parameter domains, 256 property/element
domains, 214 factories/calls, 90 destructured values, 72 bindings without
certificates, 58 opaque calls, 50 mutable bindings, 26 other expressions,
23 conditional joins, 18 prior assertions and 2 constructors. These are evidence
gaps, not proof that the actual runtime shapes fail their target contracts.

The adapted source was prepared using stage3/apply.sh at
/tmp/shape-conformance-adapted. Its generated stage3/patch-set.md change was
restored because it is outside this lane. Applying adaptations is not a native
compilation or upstream test-suite pass.

## Integration seams required for actual implementation

The listed new files alone cannot make compilation use the proof:

- graph_flow.go assigns sites only inside graphFlows, reached when cycle analysis
  selects graph types. Non-graph programs need the same allocation-site assignment.
  The producer map and backward traversal are private local values discarded by
  graphFlows. Expose that shared graph and query it from shape_flow.go; do not copy
  its traversal. Conformance must additionally propagate unknown for unsupported
  producers and open boundaries. A graph-ownership may-flow frontier is not a type
  certification.
- Allocation lowering must record the checker-declared field types before slot
  representations erase literals, unions, object structure and callable contracts.
  A side table keyed by semantic schema/allocation identity can live in a new IR
  file, but object/class allocation hooks must populate it. Native layout identity
  by names and reference bits is not semantic schema identity.
- Lane 1 must carry target and view provenance per cast/read. Its current global
  CheckedFields map cannot be cleared for a proven cast: another unknown cast
  with the same field name still needs its checks. A proven cast must retain
  object identity and propagate its certification through its own reads and aliases.
- The pass needs an orchestration call after complete producer collection and site
  assignment while checker relations remain available. lower.go currently calls
  readiness and then borrow/counters; a new file has no automatic pass hook.
- Readiness facts must be made available at the actual cast/read sites. Declared
  type assignability cannot certify initialization. The latest readiness code can
  clear initialization on undefined!/null! assignments, so an allocation-ready
  fact alone is insufficient after writes or calls.

The new plan supplies the ownership answer: lane 3 must submit named handoffs,
not edit the shared files. The needed handoffs are the shared allocation-flow
accessor, declared-field allocation metadata hooks, per-read view descriptors,
viewInitializationAtRead, and the lead-owned eraseProvenViewChecks orchestration
call. No shared hook descriptors or readiness exporter have landed yet. This is
an integration checkpoint, not a claim that conformance is impossible. The
dependency merge repairs above preceded publication of the plan and require
review by the shared-file owners; no further shared implementation is added.

## Toolchain and verification

First setup failed during dependency conflicts with Go syntax errors at the
unresolved native conflict markers. No successful setup timing is claimed for
that attempt. The retry passed using GOPROXY=https://proxy.golang.org|direct:
Go ready 0.071s; Node ready 0.070s; markdown ready 0.189s; submodules ready
0.231s; clang ready 0.475s; build ready 58.048s; cache warm 58.269s;
done 58.329s. nproc=5; cpu.max=400000 100000. Go 1.27.1, clang 20.1.8,
Node v24.19.0. Every build/test shell sources /workspace/adamic-tools/env.sh.

Initial package command:

```sh
go test ./internal/lower ./internal/ir -count=1 -timeout 30m > /tmp/shape-conformance-lower-ir.log 2>&1
```

Passed lower 50.551s and ir 51.630s. Setup logs are
/tmp/shape-conformance-setup.log and /tmp/shape-conformance-setup-retry.log.

The final dependency oracle command was:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestDefaultTaggedSourceViews|TestRequiredViewField|TestNarrowedFieldUsesSharedReadiness|TestReadinessMutants|TestDeinitializationIsNotOrdinaryStoreMutant|TestNativeAgreesWithNode/.*graph_regions/classification' -count=1 -v -timeout 30m > /tmp/shape-conformance-oracle-final.log 2>&1
```

Passed in 16.964s. Positive source fixtures are held to Node; malformed reads
are held to independent expected exit-70 diagnostics on both backends and
native release. The wrong-boolean view checks specifically reject a numeric
payload. Staged construction remains admitted until an uninitialized read.

Existing mutants rerun in this command, all caught by pinned runtime output:

- scanner-var-capture; deinitialization-values; deinitialization-entries;
  deinitialization-assign-source; deinitialization-field-alias;
  deinitialization-field-method; keep-slot-proven-after-deinitialization;
  deinitialization-through-capture; deinitialization-exception-path; drop-check;
  erase-without-proof; initialize-to-zero; miss-captured-read;
  miss-exception-path; lazy-read; weak-generic-message.
- Required-view primitive: drop wrong-type read check; drop uninitialized read
  check; remove initialization tracking. Each disagrees with the pinned expected
  exit/message; generated release C remains valid.
- Narrowed field readiness: remove initialization tracking, caught by exit/output.
- Treat a deinitializing undefined store as ordinary initialization, caught by
  exit 0 and stdout before/undefined instead of the expected exit-70 diagnostic.

These are dependency mutants. The requested lane 3 mutant that erases a cast
with one nonconforming reaching shape was not implemented or run. The requested
zero-runtime-check IR assertion, nonconforming-shape retention fixture and
host-value retention fixture are likewise outstanding.

Earlier failures are in /tmp/shape-conformance-oracle.log,
/tmp/shape-conformance-oracle-retry.log, /tmp/shape-conformance-counts.log and
/tmp/shape-conformance-counts-retry.log. The full repository gate and upstream
tsc test suite were not run. No performance result is claimed.

Counts regeneration passed in 101.025s:

```sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/shape-conformance-counts-final.log 2>&1
```

All 615 fixture rows remain present, with graph-region columns normalized and
canonical ordering regenerated. [counts-comparison.json](counts-comparison.json)
records every numeric difference against the merged pre-regeneration table.
Those values combine dependency histories and are not lane 3 optimization
results. [counts-handoff.patch](counts-handoff.patch) preserves the generated
diff for central lane 1 regeneration. The temporary counts.md edit was restored
after reading the new plan. Consequently the committed dependency table still
requires central regeneration; a counts check against that old table is not
claimed green. This branch is not ready to land. No shape-erasure performance
claim follows from these dependency numbers.

The adapted-tree stage-0 census completed. All 79 source entries fail before
IR production: 78 checker failures and one Refused hostErrors.ts (in). Three
additional non-source files are reported as errors by the tool and are excluded
from the source count. The combined 79-root check also fails at the checker.
Examples include switch fallthrough and missing-return diagnostics in binder.ts.
[adapted-census-summary.json](adapted-census-summary.json) gives the exact
frontiers; [the full census](adapted-census.jsonl.gz) preserves all diagnostics.
This does not establish that 2,936 casts are dynamically unknown: it establishes
that this IR pipeline cannot yet provide their whole-program allocation sets.
The new conforms-ready/conforms-not-ready/conforms-if share remains unmeasured.

```sh
bash stage3/apply.sh /tmp/shape-conformance-adapted > /tmp/shape-conformance-adapted.log 2>&1
go build -o /tmp/shape-conformance-census ./stage3/census/tool > /tmp/shape-conformance-census-build.log 2>&1
/tmp/shape-conformance-census /tmp/shape-conformance-adapted/src/compiler /tmp/shape-conformance-census.jsonl > /tmp/shape-conformance-census.log 2>&1
```

The full affected-package attempt was stopped after more than six minutes,
with only its Go process and three descendants terminated. Its incomplete log
/tmp/shape-conformance-packages.log is not a package pass. A scoped native
graph gate follows under the worker-gate allowance.

Final focused native command:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/javascript -run '^TestGraph|^TestCallTargetReaders' -count=1 -v -timeout 15m > /tmp/shape-conformance-native-focus.log 2>&1
```

Failed in 11.599s at four old diagnostic byte assertions: TestGraphRegionsRuntime,
TestGraphRegionsMillion, TestGraphContainerBoundary and TestGraphLazyRegions.
TestGraphClosureEnvironment and TestGraphUnreleasedAnchorCounted pass. The
intentional unreleased-anchor leak is detected (383 bytes in five allocations).
The four diagnostic expectations predate checked-view metadata. Examples:
three three-field objects report 210 bytes instead of 192; a million such
objects report 70,000,000 bytes instead of 64,000,000. The owning tests were
not edited. [graph-size-handoff.patch](graph-size-handoff.patch) proposes the
four observed diagnostic expectation updates for owner review; it is unapplied
and is not claimed to make this gate pass. Native safety and leak assertions
remain unchanged by the proposal.

Final gofmt, vet and git diff --check pass. Formatting/vet logs are empty:
/tmp/shape-conformance-format.log and /tmp/shape-conformance-vet-final.log.
Final fetched origin/main is still 71d7e491 and is already an ancestor.
Only the plan commit 4dbf3b6a from the newly advanced lane 1 branch is needed
for this checkpoint; no transitive checked-view implementation is claimed.
The branch remains unready to land because of the native assertions, central
counts regeneration and outstanding lane 3 implementation.

Explicit allocation-flow dependency check:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run '^TestGraphAllocationFlow' -count=1 -v -timeout 15m > /tmp/shape-conformance-graph-flow-final.log 2>&1
```

Passed classification and return/conditional/mixed/override behavior, source
Node agreement, ASan/UBSan, shared leak checks and counted teardown; oracle
2.711s. This explicit selection supersedes the earlier path filter: that
filter entered TestNativeAgreesWithNode but selected no classification leaf
fixtures. No leaf-fixture evidence is attributed to the earlier filter. The
other view/readiness tests in its command did run and pass.
