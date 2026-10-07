# Checked views: three-worker plan

Ruled by @system_adamic. Lane 1 stays with the interface-downcasts worker on
`codex/interface-downcasts`; due October 8 at 18:00 MDT, **00:00 UTC October 9**.
Lane 2 is worker 01a116ae on `codex/views-arrays-callables`. Lane 3 is worker
01a116af on `codex/shape-conformance`. Start from this plan tip. The checked view
remains the correctness mechanism; allocation/shape certification is its eraser.
Unknown facts retain checks, not an admission refusal. No flag enables this behavior.

## Site shares

Inputs: the unchanged 4,101-site assertions ledger and the tagged-other refinement,
with 1,758 tagged plus 1,178 untagged sites. TypeScript 6.0.3 at 050880ce is the
outside checker. These are contract dependencies, **not sites already unlocked**.

| Lane | Primary tagged | Primary untagged | Primary total | Overlapping tagged dependency | Overlapping untagged dependency |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1: objects, interfaces, unions/readiness | 1,733 | 132 | 1,865 | 1,758 | 1,171 |
| 2: arrays and callable members | 25 | 1,046 | 1,071 | 25 | 1,046 |
| 3: allocation/type certificates and erasure | all 1,758 | all 1,178 | all 2,936 | all 1,758 | all 1,178 |

Primary ownership assigns a site to lane 2 if the target is an array/tuple, or its
currently unsupported field contracts contain an array or a callable; all others
go to lane 1. This partitions 2,936 exactly. Dependency counts overlap: 25 tagged
and 1,039 untagged sites need both admission lanes. Lane 1 dependency means an
object/interface field or union/optional contract; scalar admission and common
integration remain lane 1 even for the seven targets without those dependencies.
These are direct target-contract observations, not a recursively expanded census.

Within the 1,739 tagged broader-contract sites, all need lane 1; six also need
lane 2's arrays. The other nineteen tagged targets are SourceFile callable contracts
and also need lane 1. Untagged targets expose 1,010 array dependencies and 71
callable dependencies, with 35 overlapping; their union is 1,046. Built-in array
methods account for 26 root-array callable targets and need intrinsic contracts,
not a blanket callable refusal. The earlier shape audit has **zero certified
conforming sites**; all 2,936 need allocation/type/readiness/effects evidence before
claiming erasure. 2,127 have parameter origins; 256 have property/element origins.

## Lane 1: transitive object, interface and union views (I keep this)

Own existing `internal/lower/interface_cast.go`, `cast.go`, `object.go`,
`collections.go`, `readiness.go`, and their current view tests; new
`internal/lower/view_objects.go`, `view_interfaces.go`, `view_unions.go` and tests.
Own common IR edits in `internal/ir/ir.go` and new `internal/ir/views.go`.
Own common backend dispatch/field plumbing in native `emit_expressions.go`,
`emit_statements.go`, `emit_objects.go`, `reuse.go`, `view_fields.go`, and JavaScript
`javascript.go`, `readiness.go`; new object/interface/union backend helpers use
`view_objects.go`, `view_interfaces.go`, `view_unions.go` in each backend package.
Own shared runtime `adamic.h`, `object.c`, and union representation plumbing.

Retain `(*lowering).view(node, value, target)` as the single entry point. Build a
memoized runtime contract, with cycles referenced by contract id. An object-valued
read yields the corresponding checked view through aliases, locals, parameters,
returns, captures, destructuring and containers. Optional properties permit absence
and compatible present values; uninitialized reserved slots still fail. Distinguish
null from undefined in contracts and provide a representation that preserves that
difference. Union membership is checked before member access/conversion; a tag test
selects a checked member contract without trusting its other fields. Object/interface
contracts include inherited fields. Receiver evaluation and object identity stay
unchanged. Default tagged casts keep their tag test; untagged casts have no invented
entry tag test. Reuse the non-null worker's readiness helper/state, never another bitmap.

Fixtures under `stage3/interface-downcasts/lane1/`: nested Node/Identifier fields;
recursive parent/child interfaces; inherited interface fields; malformed nested
payload; cast before a later store; undefined!/null! slots and clone/spread readiness;
optional absent/present/wrong type; nullable object; string|number and boxed writes;
tagged Identifier|Literal in a field; side-effectful receiver; aliased and returned
nested values. Add lane-owned oracle `checked_views_objects_test.go`,
`checked_views_interfaces_test.go`, `checked_views_unions_test.go`.

Push after **objects, interfaces, unions**, in that order. Each report gives actual
source lowering counts against the pinned 1,758/1,178 site set, new sites versus the
previous checkpoint, remaining reasons, and both-backend oracle/mutant evidence.
Contract eligibility alone is never reported as successful lowering. Array/callable
blockers can remain until lane 2 arrives; failed proof never disables a required check.

## Lane 2: arrays and callable members

Own new `internal/lower/view_arrays.go`, `view_callables.go` and their tests; new
`internal/native/view_arrays.go`, `view_callables.go`; new
`internal/javascript/view_arrays.go`, `view_callables.go`; new runtime
`view_arrays.c/.h`, `view_callables.c/.h`. Do not edit lane 1's shared dispatch,
readiness, IR or object runtime files. Submit named hook calls for lane 1 to wire.
Existing array/iteration/call emitters needing changes are also shared handoffs,
not permission for both workers to edit them.

Array views check element reads lazily and propagate object/union contracts to
returned elements; cover readonly arrays, tuples, holes, iteration, callbacks,
mutating aliases and representation conversion. Readonly does not prove another
alias cannot mutate. Preserve existing bounds/undefined semantics. Writes need
source/target contract compatibility; array mutators cannot bypass it.

For callable members, presence/initialization and typeof-function establish only
that a callable exists. Admit class methods, intrinsic methods and closures whose
implementation identity/signature/body is already proven compatible, preserving
this binding. Check arguments/results only when their complete runtime contracts
can certify the advertised operation; result checks alone do not prove arbitrary
callback effects or variance. Refuse an unproven opaque signature that cannot be
certified; report exact target/member and count. Do not treat inherited array
built-ins as arbitrary uncheckable callbacks.

Fixtures under `stage3/interface-downcasts/lane2/`: primitive and object arrays;
readonly array with mutable alias; a malformed second element reached lazily;
union elements; sparse array; iterator/map callback; tuple; method receiver binding;
proven class method and record closure; missing/uninitialized callable; wrong
argument/result; opaque host callback refusal. Own oracle
`checked_views_arrays_test.go`, `checked_views_callables_test.go`.

## Lane 3: allocation flow and typed-shape erasure

Own new `internal/lower/shape_flow.go`, `shape_conformance.go`, corresponding tests;
new `internal/ir/shape_conformance.go`, `internal/native/shape_conformance.go`,
`internal/javascript/shape_conformance.go`, and runtime
`shape_conformance.c/.h`. No edits to lane 1 or lane 2 files. Any required metadata
field or emitter call in a shared file is a small named handoff to lane 1.

Reuse graph-regions, inspected at **198b1271e7c9b48b20fe150c223222fe6ece9bbd**;
it is not in fetched main 71d7e491 yet. Its `internal/lower/graph_flow.go` already
provides `graphAllocationSites` and `(*cycleFinder).graphFlows`: unique negative
allocation ids, backward may-flow through joined assignments/results, known
call/closure arguments, and aggregate write seeds. Consume the same allocation
ids, do not build a second numbering system or copy the worker's file. Coordinate
its dependency merge into this feature branch before enabling this eraser. If a
reusable accessor is needed in graph_flow.go, lane 3 supplies the named extraction
to that file's owner; it does not fork the machinery. An opaque closure frontier
that graph regions can tolerate for leak-only handling must be **Unknown** here,
never an empty allocation set that proves conformance vacuously.

Extend may-flow to nested field/element stores and loads, captures, recursion,
branches, callbacks, factories, spreads/clones and host frontiers. Certify each
allocation's full declared field contracts and proven callable identities.
Names/reference bits alone are not type certificates: number and boolean currently
share a physical shape. Supply an immutable semantic schema id without changing
observable object identity. Dynamic record key sets need membership evidence.

Erase a field check only if **every** reaching allocation establishes that field's
presence and target type, plus readiness at this read and effects that preserve
those facts. A staged declared-T slot is not an initialized-T proof. A one-time
narrowing of a wider writable slot is not stable across aliases. Preserve tag tests
unless separately proven. Unknown/external origins, unconstrained effects, phantom
brands and unsupported shape facts leave the corresponding checks in place.

Fixtures under `stage3/interface-downcasts/lane3/`: a fully proven factory through
parameter/return aliases (checks erased); safe joined shapes; one malformed branch
(check retained); zero/unknown allocation set; loop and recursion; nested alias
write; opaque callback; clone; staged read before/after a store; identical physical
layouts with different scalar types; dictionary missing key; callable override.
Own oracle `checked_views_erasure_test.go` and verify emitted checks as well as Node.

## Hook contract and no-overlap integration

The following are **proposed named hooks**, not already implemented APIs. Lane 1
owns wiring in view() and shared backend dispatch. Lane 2/3 own hook implementations
in their reserved new files. Lane 1 first publishes minimal descriptors and hook
adapters; until a family is merged its adapter returns NotYet or no proof. Workers
can build/test their helpers directly without editing another lane's adapter.

- `viewContract(node, target)` (lane 1): intern a complete `ir.ViewContractID`;
  optional/union first routes to `viewUnionContract` (lane 1), arrays/tuples to
  `viewArrayContract` (lane 2), call/construct signatures to `viewCallableContract`
  (lane 2), and ordinary structural objects to `viewObjectContract` (lane 1).
  Each family receives a recursive contract-builder callback, so lane 2 can return
  object/union element contracts without implementing lane 1's rules.
- `view(value, target)` uses that contract; `markViewRead` (lane 1) propagates the
  id and expression text through fields and object-valued results. Use the existing
  full `(node, value, target)` signature; the shorthand here is not a new entry point.
- `emitViewArrayRead` and `emitViewCallableRead/Call` (lane 2, each backend): receive
  the contract id, source text and already evaluated operand; do not evaluate twice.
  Lane 1 wires them into the common read/call dispatch and includes lane 2's runtime
  headers/JavaScript helper text.
- `viewReachingAllocations(program)` and `certifyViewShapes(program, contracts)`
  (lane 3): return a may-flow set **with an explicit Unknown bit**, and per-allocation,
  per-field semantic certificates. Graph allocation ids are preserved.
- `viewFieldProof(readSite, allocationSet, field, contract)` (lane 3): returns
  separate presence/type/readiness/effect facts, with no-proof as the default.
  `eraseProvenViewChecks(program, proofs)` (lane 3) applies only the justified parts;
  lane 1 wires this after view propagation/readiness and allocation ids are established.
  Lane 1 exports `viewInitializationAtRead` from the shared readiness analysis so
  lane 3 reuses the state rather than inventing initialization tracking.

Only lane 1 edits common IR structures, common dispatch, shared object/header files,
`internal/oracle/counts.md`, existing view test files and this plan. Each lane owns
its new fixture/test/log subtree and supplies count-row additions for lane 1 to
regenerate centrally. All three avoid protected `internal/native/emit.go`,
`internal/lower/lower.go`, `internal/native/native.go`,
`internal/oracle/oracle_test.go`; any indispensable orchestration edit there goes
to the lead as a concrete minimal patch, not an uncoordinated worker edit.

All families: positive executions match source Node byte for byte; negative checks
pin exit 70, expression/property, expected contract and found category on both
backends. Mutants independently drop field, readiness, transitive/element or
callable-contract checks, erase without proof, omit a reaching shape, and evaluate
an operand twice. Valid release C and semantic assertions catch them, not clang or
a sanitizer. Capture test output in logs, regenerate counts, push own branches only.
Complete admission before benchmarking erased reads; report retained read-site and
dynamic-read counts separately. Unknown proofs never reduce checks.
