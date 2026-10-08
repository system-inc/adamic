# Views V6: remove only proven checks

Task #t389n3b, roadmap step 11 under #a03mesg. This is the pre-build plan,
not a completed compiler/views-v6 delivery. Current main is
54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8. Origin does not advertise
compiler/views-v5 in the completed base checks. No V6 admissions or runtime
changes are applied to main or this planning branch.

When compiler/views-v5 exists, create compiler/views-v6 from its exact tip.
Do not use the private dry-run branch as the delivery base. Preserve every
base acceptance, refusal and NotYet result except removal of independently
proven view checks. No optional-write, callable, array, dictionary, union or
cyclic-ownership admission is part of V6.

## Sources and lane commits

Read CLAUDE.md and the language/memory rules before implementation. Source
checkpoints are compiler/area-views-wip3 at 7302410e and
codex/views-integration at 432d4913. Conflict judgments are read from
codex/views-next-pending at c923d1cc. The lane-3 section of
[checked-views-plan.md](https://github.com/system-inc/adamic/blob/7302410ecfe4d54c6321fde785f67bf39fc5dcc4/docs/checked-views-plan.md) owns allocation flow, semantic
certificates and erasure; its hook names describe the intended interfaces,
not additional admissions. The implemented subset is smaller: exact scalar
contracts on closed immutable records.

The following non-merge lane commits were cherry-picked in order onto current
main in the private worktree /workspace/scratch/views-v6-lane-replay, branch
codex/views-v6-lane-replay. Complete hashes and changed paths are in
[the lane manifest](views-v6/lane-commits.json).

| Commit | What it changes |
|---|---|
| 55b6b4aa | Shape-conformance audit, physical-layout collision/readiness witnesses, source ledger and boundaries. |
| bb6c0395 | Allocation-flow prerequisite report, checker-rejected census, count and graph-size handoffs. |
| a0385814 | Preserve checkpoint verification logs and report. |
| dee69981 | Record graph-flow dependency verification against Node/leaks. |
| 59db5d63 | Extract shared allocation query in graph_flow.go; add AllocationSet and FieldTypeCertificate; attach declared field certificates; attach complete cast contract; run scalar-record erasure after readiness; add proof tests, four source probes, oracle and graph-count guard. |
| 4a668ef4 | Record implementation identity and final validation. |
| 2ab6be95 | Measurement-only diagnosed-body adapter, provenance audit, control mutants and census artifacts. No production admission. |
| ffe5a90e | Record share identity and historical main-merge conflicts. |
| e7f42029 | Query property/constant-element producers; join visible stores; retain Unknown for opaque stores/calls, missing fields, recursion, spread and array mutation; add two source probes and projection mutants; update measurement artifacts. |
| b8316acd | Group diagnosed-body dependencies by file and record actions. No production admission. |

Only 59db5d63 and e7f42029 add the production eraser/query implementation.
Historical reports and measurements remain attributed to their original tree;
they are not evidence that V6 passes on its new base. Do not import merges,
including 40ad020d or c1f4c5a7, to obtain prerequisites. c1f4c5a7 also contains
checkedViewTarget integration; retain V5's actual cast entry instead of taking
that merge's unrelated changes. c9ad59e4 is a dictionary-slice commit, not a
lane-3 commit: preserve only its already-integrated safety boundary that dynamic
DictionaryKey reads cannot use this scalar eraser. 1c410de9 is a cross-lane
assertion-fixture migration, not a V6 source admission or whole-commit dependency.

Exclude runtime commit 2724dabd and all fresh_refused/graph_regions fixture
admissions. The historical 43-row graph-count test from 59db5d63 depends on
runtime's fixture set and is not a V6 acceptance-test addition. Compare whatever
graph/ownership controls actually exist on V5 without importing absent runtime
fixtures or changing their refusal expectations. V5 must provide the shared
allocation-query substrate without enabling excluded cyclic ownership.

## Observed dry-run conflicts

Ten lane commits were replayed without their merges. Fourteen paths conflicted.
The exact cherry-pick diagnostics and conflicted diffs are preserved in
[dry-run-conflicts.json](views-v6/dry-run-conflicts.json).

59db5d63 has three production conflicts:

- internal/lower/graph_flow.go: modify/delete; main has no shared graph file.
- internal/lower/graph_types.go: modify/delete; main has no graph-allocation hook.
- internal/lower/readiness.go: the lane's defer hook shares a hunk with historical
  deinitialization analysis absent from this main. Add only the final erasure
  call to the eventual V5 readiness analysis; do not replace its analysis.

The eleven e7f42029 conflicts are historical artifact context from the excluded
merge checkpoint, rather than additional production changes:

- stage3/shape-conformance/REPORT.md
- stage3/shape-conformance/latent/lower.go.txt
- stage3/shape-conformance/latent/mutants.py
- stage3/shape-conformance/logs/latent-drop-diagnostic-provenance-build.log
- stage3/shape-conformance/logs/latent-drop-diagnostic-provenance.log
- stage3/shape-conformance/logs/latent-ignore-diagnosed-body-build.log
- stage3/shape-conformance/logs/latent-ignore-diagnosed-body.log
- stage3/shape-conformance/logs/latent-ignore-field-type-build.log
- stage3/shape-conformance/logs/latent-ignore-field-type.log
- stage3/shape-conformance/logs/latent-ignore-readiness-build.log
- stage3/shape-conformance/logs/latent-ignore-readiness.log

For discovery only, the private worktree retained the lane side of conflicts so
later cherry-picks could be attempted. That choice is not a real merge decision,
an admission, or a validated implementation. Real V6 must preserve V5's source
and reapply only the lane-owned changes and ruled boundaries. No production file
is changed on the planning branch.

## Conflict judgments relevant to V6

Apply each relevant decision from both ledgers; preserve the decisions already
landed in V5 rather than cherry-picking the monolithic integration branch.

- area-views-merge-judgments.md, ir.go h3/h4: retain semantic null/undefined tags,
  Record=14 and distinct typed-array tags 15/16/17. The V6 certificate fields do
  not renumber representations or change reference predicates.
- The same ledger, cycles.go h1 and the later cycle ruling: query allocation
  metadata without selecting graph ownership. Keep every base cycle refusal,
  captured-reference boundary and call-count proof. Runtime admission is excluded.
- Source-extension ruling 7: .a postfix assertions remain refused. The original
  staged lane probe contains undefined!; use an existing ruled .ts fixture from
  V5 if present, or construct the readiness proof witness in IR. Do not add a new
  positive .ts file or weaken .a refusal to make this old probe executable.
- area-views-next-merge-judgments.md, judgment 19: preserve FunctionTypeTargets
  and other dispatch metadata; executable-IR comparisons must distinguish
  compile-time certificates/checker IDs from executable behavior. Assert retained
  metadata separately instead of deleting it to satisfy equality tests.
- The next ledger's retained optional-write expectations: fresh optional-own-field
  writes keep V5's exact NotYet outcome. Allocation evidence never admits writes.
- Dictionary integration's shape_conformance.go boundary: DictionaryKey != nil
  retains its check; a scalar declared field does not prove dynamic key membership.

There are no judgment rows for graph_flow.go, graph_types.go, readiness.go or
interface_cast.go in those two ledgers. Their V6 edits are the named lane hook
handoffs, not permission to replace unrelated base logic. If indispensable
wiring touches a protected file, prepare the minimal handoff for its owner;
do not edit lower.go, native/emit.go, native/native.go or oracle/oracle_test.go.
Judgments about callable adapters, dictionaries and other slices stay with those
slices unless V6 actually changes their files.

## Representation and erasure rules

Reuse the base allocation IDs and call-target machinery. Never create a second
numbering system, ownership selector, readiness bitmap or callable convention.
Allocation queries return all possible Sites plus explicit Unknown and reasons.
An empty set, omitted argument, recursive unresolved producer or opaque frontier
cannot certify anything. Known sites survive alongside Unknown.

Certificates preserve checker type identity, declared semantic type and
optionality before physical layout loses number/boolean distinctions. A declared
T reserved slot is not initialized T. Capture contextual declarations for staged
slots, but obtain readiness solely from the base analysis.

The initial eraser remains the lane's conservative immutable scalar-record subset:

1. Reject proof if any origin is unknown/absent, its contract is missing, or the
   field contract is not a supported scalar contract.
2. Every reaching allocation must have that field, a nonoptional declaration,
   the exact target contract and an initialized slot. Classes and spreads retain
   checks. Every join member must pass, not just one conforming member.
3. A read retains checks for unresolved readiness, optional/absent/method status,
   dynamic dictionary key or unsupported effects. Any property/index mutation,
   object operation or opaque closure call conservatively disables this subset.
4. Clear only the proven read's view-check metadata. Remove a CheckedCast only
   when the complete target object's fields are independently proven; partial
   field evidence must not erase a tag or whole-object obligation.
5. Preserve every operand evaluation, identity, original field representation,
   dispatch target, descendant obligation and ownership action unrelated to the
   check. Projection joins initial values and every visible store, with Unknown
   for initially absent fields, opaque aliases, mutation or unfinished recursion.

Do not extend erasure to transitive object contracts, arrays, callables, dictionaries,
phantom brands, mutable aliases or post-store readiness merely because previous
slices admit their checked forms. Unsupported proof facts leave checks in place.
The four planned API names in the worker plan correspond here to shared
ReachingAllocations, certifyAllocationFields, the scalar fieldProof predicate and
eraseProvenViewChecks; broader certificates are not silently supplied.

## Source observations and language questions

Fresh Node observations and source hashes are in
[node-observations.json](views-v6/node-observations.json). All six original
sources exit 0 with empty stderr. proven.a and proven-property.a print hello;
nonconforming.a and nonconforming-property.a print true then 0; host.a prints
true; uninitialized.a prints undefined. The old oracle deliberately expects
checked-view panics for malformed/staged reads rather than Node's unchecked
cast behavior. Those diagnostics must be compared on both backends, with Node's
original output recorded independently.

No new language question was found in the pre-build. The staged source's
undefined! is already ruled: keep .a refusal and the existing .ts boundary.
Its Node behavior does not authorize an Adamic admission. The host probe's old
unsupported-representation failure is an inherited checkpoint observation;
measure its actual V5 outcome before assigning an expected result. Unknown host
origins retain checks even if another slice now makes the checked execution work.
Do not repair host admission in V6. A future question must include the complete
program, fresh Node output and a concrete proposal before being sent for ruling.

## Validation once V5 exists

Run base and V6 on this same machine, with the same pinned dependencies and
fixture bytes. The unit-specific comparison requirements authorize the named
complete package runs; do not run the repository-wide gate. Write every result
to a separate log and retain complete per-test output for comparison.

- Run internal/lower, internal/ir, internal/flow and internal/javascript on both
  trees; run the full internal/oracle on both trees uncached. Compare test leaves,
  exact refusal/NotYet diagnostics, source Node outputs and both backend behavior.
  Every changed result needs its exact name and a Moved-result commit trailer.
  Unexpected admissions or refusals block delivery; do not refresh expectations.
- Test existing and new V6 proven factories/aliases/joins and unsafe origins.
  Inspect emitted checks and IR in addition to Node, release native, sanitizers
  and leaks. Preserve base readiness and tag failures byte for byte.
- Refresh TestCountsAreRecorded for V6 rows only. Any other row difference must
  be independently attributed to V6 erasure or reverted; do not import historic
  count handoffs or runtime's missing rows.
- npm ci --prefix stage3/api first. Run stage3/fixtures on each tree normally and
  again with only its platform guard lifted locally. Restore the file before
  committing; do not update Linux host records to conceal changed behavior.
- Run go test ./stage1/... -run 'Gap|Gaps|Probes' on both trees. If a gap closes,
  move it to its package's closed form and update that GAPS.md, naming its result.
- Inspect cloud/fast-gate/run.py aCheck on the real base, then apply its exact
  rules to every added/changed .a outside a Go package. That script is absent from
  current main, so this pre-build does not invent a substitute or claim a-check.
  Negative first-line headers must name the exact refusal/type error and have
  an asserting test. No new .ts Adamic source is permitted.

Commands for the required comparisons, each redirected to a named tree-specific log:

```sh
go test -count=1 -timeout 30m -json ./internal/lower ./internal/ir ./internal/flow ./internal/javascript
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -json ./internal/oracle
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
npm ci --prefix stage3/api
go test ./stage3/fixtures -count=1 -timeout 30m -json
go test ./stage1/... -run 'Gap|Gaps|Probes' -count=1 -timeout 30m -json
```

## Mutants to run, not prior evidence claimed as new

Each new proof guard needs its independent removal mutant, with valid compiled
code and the named fixture/assertion catching it. Existing lane runners cover
wrong-contract erasure and readiness erasure, plus projection store, opaque,
missing-field and array-mutation guards. Rerun them on V6 after adaptation;
clang/build failures do not count.

Add independent controls for omitted reaching allocations, Unknown with known
origins, empty origin sets, missing/optional declarations, scalar contract kind,
unsupported effects, class/spread exclusions, dictionary key membership and
partial-target/tag preservation wherever the implementation adds a check.
Positive erasure controls must also fail when the eraser is disabled, proving
that source/backend equality alone is not masking retained checks. Preserve
number/boolean layout collision, branch joins, parameters/returns, staged IR,
opaque callbacks, recursive projections and constant elements in the controls.

No production mutant has been run on a real V6 base in this pre-build. The
historical lane logs are not counted as current validation.

## Work performed and current block

Setup succeeded with GOPROXY=https://proxy.golang.org|direct, sourcing
/workspace/adamic-tools/env.sh. Timing lines: Node ready 0.020s; Go ready 0.021s;
markdown skipped 0.007s, markdown/submodules ready 0.064s; clang ready 0.153s;
build ready 44.156s; test binaries deferred 44.287s; cache warm 44.290s;
done 44.321s. nproc=5, CPU quota=4; Go 1.27.1, clang 20.1.8, Node 24.19.0.

The scratch worktree initially lacked its cohere submodule. Referencing the
already-initialized repository submodule fixed that loader error without copying
cohere code. The focused dry-run command then failed at
internal/ir/ir.go:593:17: undefined: ViewContractID, confirming the missing V1
frame dependency. Log: /tmp/views-v6-dry-run-build.log. Command:

```sh
go test -buildvcs=false ./internal/lower -run 'TestShapeFlow|TestShapeProjection|TestShapeErasure|TestShapeCertificates' -count=1 > /tmp/views-v6-dry-run-build.log 2>&1
```

This is not a mutant kill or a green package comparison. Full package/oracle,
counts, stage3 normal/platform-lifted, stage1 gaps and a-check comparisons remain
unrun because the required compiler/views-v5 base is absent. No result move is
claimed. The private cherry-pick replay is not suitable for delivery. No push
to compiler/views-v6 is made until that base exists and every required comparison
and mutant passes. Origin is rechecked on each work pass.
