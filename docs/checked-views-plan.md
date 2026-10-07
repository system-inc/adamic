# Checked views: worker plan

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

| Lane | Primary tagged | Primary untagged | Primary total | Direct tagged dependency | Direct untagged dependency |
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

## First shared API checkpoint

`internal/ir/views.go` defines one-based `ViewContractID`, logical `ViewContract`
(kind/name/representation, scalar values, object fields, union members, array
members and callable signature slots), and `ViewFieldContract` (name, child id,
optional, readonly). `Program.ViewContracts` owns the interned graph;
`Program.ViewContractTypes` maps checker type ids to contract ids. A field read
can carry `Property.ViewContract`. Zero/Unknown never authorizes erasure.

The concrete lane 2 registration points are `viewArrayContractHook` and
`viewCallableContractHook` in `internal/lower/view_contracts.go`, with signature
`func(*lowering, *ast.Node, *checker.Type, viewContractBuilder) (ir.ViewContractID, error)`.
Register from the lane-owned files' init functions, and use the builder callback
for element/argument/result contracts. Hook implementations intern their returned
contract in the same program registry. Lane 1 owns dispatch and broader read-path
wiring; registering a handler alone does not claim array/callable reads are implemented.
Tagged object-union field contracts now carry ViewUnion member ids and shared
field contracts. Nullish/mixed union contracts remain unsupported. Property.ViewTypeID
preserves the checker type id so readiness finalization can resolve a contract
interned after an earlier function body was lowered. Lane 3 must still retain checks
for zero/Unknown contracts; physical shape ids do not certify these logical types.

## Array integration prerequisite after the renewed blocker census

The overlapping target-contract census makes arrays/tuples the biggest blocker:
2,909 sites, compared with nullish members 2,903 and optional properties 2,889.
See [checked-views-blockers.md](checked-views-blockers.md) for the complete table,
limits and exact source witnesses. Lane 2 tip 2a16ce35 is merged into this feature
branch as f709a1c5. It supplies standalone helpers, not native compiler admission.

Lane 1's `internArrayViewContract` in view_array_adapter.go registers the existing
viewArrayContractHook and adapts lane 2's error-only recursive element callback.
It interns the array before its element so recursive interface/array contracts
share the same ids. Errors propagate; unknown contracts never certify elements.
The callable hook stays unregistered pending implementation certification.

Native `adamic_array.element_kind` now provides the missing physical storage byte.
Zero is unknown, 1/2 distinguish number/boolean, 7 is packed maybe-number, and 10
is a heap pointer classified at the selected read. Plain fresh literals populate
it; opaque/reused/spread/runtime producers remain unknown. Do not set it based on
a target view or treat it as an element-shape/initialization proof. The added byte
fits existing padding, with sizeof(adamic_array) still 56 here. Lane 2 can now
build native helpers in its reserved files against this field. Lane 1 wires those
helpers into shared dispatch only once all read/mutation paths are covered.
Array admission is still NotYet; this prerequisite unlocks no complete tsc sites.

The original primary lane partition above is unchanged and its dependency columns
are direct-only observations. The renewed recursive census reaches arrays and
callables through nested interface fields at many more sites; it does not reassign
those sites or claim successful lowering. Mixed primitive element descriptors now
carry explicit ViewUnion member ids instead of an ambiguous ViewScalar/Union
storage certificate. This is metadata preparation, not mixed-union admission.

## Updated lane split and delivery estimate

Lane 1 next owns nullish members and optional properties, then mixed unions.
Lane 2 next owns native array reads and mutation through the shared contract,
then callable contracts. Existing file ownership and hook signatures remain as
listed above. No other worker should edit the shared null/undefined representation
or the object readiness/read dispatch; lane 1 supplies those handoffs.

The remaining-family census in checked-views-blockers.md supersedes the earlier
direct-only dependency counts for scheduling. It records all missing families at
each site, exclusive ownership counts and a two-lane projection which does not
claim that other unresolved contracts have disappeared.

Working estimate for lane 1 nullish plus optional: October 8 22:00 UTC (16:00 MDT),
with native null/undefined separation and mutation/alias parity still a risk.
This is an estimate, not a completion claim; each family is pushed only after its
both-backend oracle and semantic mutants pass.


## Lane 5 callable handoff after c898009b

The lead moved callable work to lane 5 on codex/views-callables, starting from
c898009b. Lane 2 now owns native arrays only, including array intrinsics and
mutation; its current integration branch is codex/views-arrays-callables-parser.
The earlier combined array/callable shares and territories above are historical.
Lane 2 will not edit callable files or certification hooks further. Lane 1 owns
the cast merge conflict and will push lazy admission with Lane 2's tip merged;
Lane 2 merges that SHA rather than resolving that conflict locally.

Lane 5 owns internal/lower/view_callables.go and its tests,
internal/native/view_callables.go, internal/javascript/view_callables.go, and
internal/native/runtime/view_callables.c/.h. The callable fixtures currently live
under stage3/interface-downcasts/lane2; future callable fixtures should use a lane5
subdirectory. checked_views_callables_test.go contains the existing callable
oracles as well as the parser array probes; leave it to lane 5 and add subsequent
array tests in separate array-owned oracle files to avoid concurrent edits.

Existing callable integration points at c898009b:

- buildViewCallableContract registers viewCallableContractHook in init and interns
  a ViewCallable/Closure kind contract. It does not build argument/result contracts
  or certify arbitrary signatures; construct signatures remain NotYet.
- viewCallableFieldUses, called by viewObjectFields, currently scans same-named
  implementations and escaping reads across the module graph. It requires a body
  and viewSignatureCompatible, which checks strict parameters/results and writable
  relations. It is a conservative predecessor to per-read allocation certification.
- viewCallableCall/checkViewCallableSignature pin the refusal naming the member:
  its signature cannot be checked at runtime and no compatible implementation is
  proven. Runtime function kind never supplies signatureProven.
- viewProvenClassCast is invoked by deferredViewCast in view_cast_preflight.go.
  Its current tag-to-construction scan requires every matching construction to
  have the compatible nominal class identity before free method dispatch.
- Native emitViewCallableRead calls adamic_view_callable. The runtime distinguishes
  an own closure from a shape method and preserves readiness/kind checks and this.
  JavaScript emitViewCallableProperty uses emitViewCallableRead/Call and the
  adamicViewCallableRead/Call helper text, preserving explicit method receivers.
- Shared metadata uses ViewContract.Kind=ViewCallable, Of=Closure and
  Property.Method/Of/View/ViewContract and Readiness. Lane 1 owns the shared
  dispatch/readiness wiring.

Array elements whose declared type is callable remain a dependency on lane 5:
viewArrayContract delegates the element to the common builder callback, while the
current viewArrayFields uses viewCallableCall for an unproven element signature.
Lane 2 may check an element's physical closure kind, but will not interpret that
kind as a compatible callable signature. Array intrinsic contracts remain lane 2;
user-defined function-valued array elements and fields remain lane 5.

Read-based callable demand is 308 type/field pairs and 1,503 source reads in the
pinned full compiler, from the c898009b-following census at 2de6d1ba. These are
partial demand, not successful lowerings or a claim every read is blocked. Exact
pairs and source witnesses are in stage3/interface-downcasts/lane2/read-census.
Existing controls include function-field, class-method, callable missing/kind/
uninitialized fixtures and the pinned opaque-signature refusal. The previous
signature mutant was caught; its evidence is retained in parser-return-logs.

## Centralized checked-view integration

The lead assigned Codex 01a11882-3830 on codex/views-integration to merge every
lane tip and resolve conflicts once, hunk by hunk. This supersedes all direct
cross-lane merge instructions above, including the Lane 1 cast-conflict handoff.
Each worker pushes only its own branch and obtains another lane's code by merging
the newest codex/views-integration SHA when needed. Workers blocked on lazy
admission build against that integration branch where possible; otherwise they
report the blocker and rest until the lead wakes them. Lane 2 remains native
arrays only, and Lane 5 retains callable ownership and certification hooks.
