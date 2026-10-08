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

Current main 48c05d09 introduces a checker-only cast preflight. `castProof.view`
routes eligible structural downcasts to the existing full `view(node,value,target)`
admission, rather than treating them as proven upcasts. The preflight does not
certify any field. Lane 2 and lane 3 hooks and ownership remain unchanged. Shared
readiness still handles staged assertions; no second state representation is added.

## Lazy admission takeover, October 7, 2026

The October 7 15:00 ruling replaces eager target-contract eligibility with lazy
admission. Working delivery estimate: **October 8, 2026 at 22:00 UTC**.
This estimate covers lazy admission and its evidence, not full tsc compilation.
The previous lane 1 worker is stopped; this work owns its view entry point and
shared read plumbing on `codex/views-lazy-admission`, based on `ab4d6f902`.

1. Merge lane 2 `codex/views-arrays-callables-parser` at `2693041a` or newer,
   keeping both cast paths, then consume lane 3's existing shared closed-world
   flow on `codex/shape-conformance`. Do not create a second flow graph.
2. Intern unsupported member descriptors without refusing the cast. Preserve
   tagged entry checks. Refuse unsupported families by field name at reads
   reachable from viewed values; unresolved flow retains a runtime tag check.
3. Cover helpers accepting both viewed and ordinary values, generic instances,
   callbacks, stored fields, aliases and transitive reads using shared may-flow.
   Unknown origins never justify trusting a read.
4. Add an unread unsupported-field fixture and the helper-read mutant, with
   source Node and both backend evidence. Run independent check-removal mutants
   and touched-package gates, with all test output captured in log files.
5. Rerun the pinned 2,936-site census, separating actual tagged and untagged
   lowering successes from contract eligibility and remaining read families.
   Merge current origin/main, recheck, and push only this branch.

Push every passing checkpoint. Any unfinished family or unmeasured census stays
explicitly unclaimed. Compiler changes begin only after reading view() and its
callers and the shared flow implementation.

## Lane 5: callable shape checks (October 7)

Lane 5 owns branch `codex/views-callables`, based on lane 2 parser tip
`2de6d1ba` (descendant of `c898009b`). The task explicitly authorizes this
territory entry. The lane 2 handoff below transfers its existing callable files
to lane 5; shared read dispatch and closure producer files remain owner handoffs.
New lane 5 files: `internal/lower/view_callables_contract.go` and its test;
`internal/native/view_callables_contract.go` and its test;
`internal/javascript/view_callables_contract.go` and its test;
`internal/native/runtime/view_callables_contract.h`;
`internal/oracle/checked_views_callable_contract_test.go`;
`stage3/interface-downcasts/lane5/` fixtures, census ranking and reports.

The October 7 ruling supersedes the earlier implementation-only admission
policy: every reached callable read must check recorded closure signature
against declared arity, parameter and result representations. A function tag
alone or a same-name program scan is insufficient. Unsupported unread members
must not refuse a cast. Missing optional members yield undefined; reserved
uninitialized slots fail using the existing readiness machinery.

Required owner handoffs, not implemented claims:

- Lane 1/lane 2: have `buildViewCallableContract` delegate signature construction
  to lane 5's builder; keep registration in its existing owner file. Fill the
  existing `ViewContract.Parameters` and `Result` fields. Preserve deferred
  unsupported signatures until a reached read.
- Closure-convention owner: record the implementation's arity, parameter and
  result representation descriptor on every created closure and method thunk.
  Native `adamic_closure` currently has only code and capture cells. JavaScript
  closures need the same immutable descriptor. Never derive it from the view.
  Unknown metadata must fail at a viewed callable read.
- Lane 1 shared read dispatch: after presence/readiness and callable-kind checks,
  call lane 5's shape helper before storing or calling the read value, including
  optional present callbacks. Pass evaluated operands and expression text once.
  Native method tables need signature descriptors too; keep receiver binding.
- Lane 1 propagation: viewed values reaching shared helpers, generics, callbacks
  and fields must retain checks; unknown flow cannot authorize trust.
- Call emission: use the reconciled closure convention after its supplied SHA is
  merged. No second calling convention is introduced by lane 5.

Working estimate for all 308 pairs / 1,503 potential reads: October 10 at
18:00 UTC (12:00 MDT), conditional on lazy-read dispatch and recorded closure
metadata handoffs. This is an estimate, not certified coverage. The ranked
ledger will retain every pair until both-backend integration and mutants prove
its read checks. Intrinsic string/map/set members appear in the callable census;
method contracts require intrinsic implementation signatures, not closure tags.

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

### Lazy admission checkpoint, October 7, 2026

Lazy unsupported descriptors now admit unread members. The shared allocation flow
checks demanded callable reads, joining helpers and retaining Unknown callbacks.
Tagged, untagged, and unread mixed-union fixtures match Node in native sanitized,
native release, and JavaScript runs. The helper mutant refuses at its opaque read.
Focused lower and oracle tests passed; array consumer coverage and the full view
suite are still pending. The earlier lane 2 and lane 3 merges preceded the new
integrator-only coordination rule. Future dependency merges use views-integration.
No integration branch was advertised by origin at this checkpoint.

### Lazy admission coverage checkpoint, October 7, 2026

The shared-flow demanded-read guard now covers wired array element consumers and
retains unsupported member families across wider helper interface type ids.
Untagged object unions without a finite discriminant retain an unsupported
descriptor and refuse at a demanded read. Existing representation/accessor
preflight refusals name their read field and family. Array consumers without
lowering adapters still fail closed. Unsupported contracts cannot certify writes.

The focused lower and checked-view oracle suite passed. Direct, generic, callback,
and stored-field helper mutants all refuse at line 9, the opaque member read.
Deleting the read-demand guard made all four oracle controls fail, along with
shared-flow helper and Unknown controls; the deletion was restored. Complete
lower package tests still report failures in the merged pre-lazy baseline.
Phantom-array, predicate marker, overload diagnostic, and nested-function failures
were reproduced against c01ae313 using a Go source overlay. Those are not claimed
as a green gate; reconciliation belongs with the integration branch.

### Lazy admission evidence complete, October 7, 2026

Core lazy descriptors and demanded read refusals have passed the focused view suite.
The exact census matches 2,936 original spans: 1,758 tagged and 1,178 untagged target
contracts intern descriptors. Unchanged production compilation is still 0 tagged,
0 untagged: Adamic checker policy stops the upstream project before lowering.
The family/read inventory, commands, mutants and limits are in
[the lazy admission report](../stage3/interface-downcasts/lazy/REPORT.md).

Direct, generic, callback and object-field helper mutants refuse at the opaque read;
a missing-name helper mutant exits 70 in sanitized native, release native and JS.
Deleting the demand guard is caught by every wired array consumer control, wider
helper and Unknown flow controls, and the source helper controls. Positive tagged,
untagged and unread mixed-union fixtures match Node in both backends.

The full JavaScript package passes. Full native tests fail five graph-region tests,
including sanitizer use-after-free and leaks; a c01ae313 source overlay reproduces
all five. Full lower tests likewise retain reproduced merged-baseline failures.
An extra Error.code probe exits 70 at the native read with an unsupported representation,
rather than matching Node's missing optional field result. It is a recorded limit,
not a successful positive fixture. Two old eager optional-Error compile assertions
were updated to expect lazy admission. The final focused lower and view oracle tests
pass. The integration branch was still absent from origin at the last check. This
work is available on its own branch; whole-project compilation and a green integration
gate are not claimed.

## Lane 4: mixed unions and full union admission

Worker on `codex/views-mixed-unions`, based on lane 1 b4cfd1aa (which descends
from af0a0ae7), with lane 2 0b141c26 merged. The user reserves this addition to
this plan for lane 4; earlier plan-only-lane-1 ownership does not apply here.
Lane 1 continues to own nullish representation and optional/readiness checks.
Lane 4 owns new `internal/lower/view_unions_mixed.go` and its tests, new
`internal/native/view_unions_mixed.go` and its tests, new
`internal/javascript/view_unions_mixed.go` and its tests, new runtime
`view_unions_mixed.c/.h`, new oracle `checked_views_mixed_unions_test.go`,
and `stage3/interface-downcasts/lane4/` fixtures, census, logs and reports.
Existing union files and shared dispatch remain lane 1 territory.

Required handoffs before compiler admission:

1. In `view_contracts.go`, route unions through the lane-owned
   `viewMixedUnionContractHook` before callable detection and representation
   refusal. Lane 4 declares this hook in `view_unions_mixed.go` with
   `internMixedUnionViewContract` as its implementation. The shared registry,
   recursive builder callback and rollback
   on failed member construction must retain their common contract semantics.
2. In `view_unions.go`, dispatch mixed and untagged union field registration to
   `viewMixedUnionFields(node,target,fields,seen)` before the finite tag refusal.
   In `view_objects.go` and `interface_cast.go`, permit complete union contracts
   in field reads/aliases rather than the current scalar/object/array range gate.
   In `cast.go`, route union targets to the shared view entry before refusing.
   Admission must still preserve writable-slot and nominal restrictions.
3. Native runtime: expose a non-panicking, readiness-aware slot probe returning
   presence, initialized state, logical runtime kind and normalized payload.
   It must classify original storage before conversion, distinguish null from
   undefined using lane 1 representation, respect static ownership/accessors,
   and never call a getter while testing alternatives. Current
   `adamic_object_view` panics on the first rejected member and cannot select
   between two untagged object alternatives. An object heap tag alone is not
   structural membership. Lane 4 will consume this probe; it will not duplicate
   object.c readiness or manufacture nullish evidence.
4. Shared backend field dispatch must hand the already evaluated operand,
   complete ViewUnion id, expression text and declared union name to lane 4
   selection. Keep the selected member contract on aliases and narrowed reads;
   a successful tag test never erases its transitive payload checks. Primitive
   conversion happens only after selection. Include lane 4 runtime helpers.
   Array alternatives use lane 2 contracts, including later element reads.

No census family is removed just because its contract builder or standalone
selector passes. The integration checkpoint requires actual source lowering,
source Node controls, both backend refusal pins and all three requested mutants.
An untagged object union may overlap: accept a value satisfying any complete
member; never commit to the first heap-kind match. Failed alternatives must not
panic before all alternatives have been considered.

### Lane 4 read-demand checkpoint and lazy-admission handoff, October 7

The user superseded the transitive census with read demand. Lane 4's first
checkpoint is ce2545e4, published before building further. The explicit read
inventory and conservative Unknown table live in docs/checked-views-blockers.md.
This addition does not transfer ownership of the common census/dispatch files.

Additional lane 4 territory: lane4/read-demand.cjs, read-demand-*.json.gz,
read-demand-summary.json, audit-read-demand.py, make-read-flow-overlay.py,
map-read-flow-input.py, export-read-flow.py, read-flow artifacts and
read-fixtures/*.a. The new oracle remains checked_views_mixed_unions_test.go.
The scratch overlay extends lane 3's source adapter by receiver queries only;
allocation numbering, argument/return joins and solving use its original
assignAllocationSites/newAllocationFlowGraph/ReachingAllocations code. Neither
cohere code nor the flow engine is copied into lane 4.

Concrete lane 1 hooks needed for sound lazy contracts:

1. Split view admission from descendant support. view() currently invokes
   viewObjectFields and recursive viewContract before constructing the cast.
   Reserve a target/tag descriptor without rejecting every descendant contract.
   Keep an unsupported child as a deferred obligation with its declared family
   and refusal reason, never as a successfully validated ViewUnknown value.
2. At each read, consult the receiver's reaching-view status through lane 3's
   shared graph. Known viewed and Unknown receivers retain a cheap tag test and
   the field's declared contract check; a helper's interface parameter cannot
   lose the check because its body lacks a lexical cast. Preserve this through
   generics, callbacks, aliases and object/array-held values. Unmodeled producer
   edges remain Unknown. Missing flow cannot certify a receiver ordinary.
3. Remove the eager global scan keyed only by field name as an admission gate.
   Make unsupported-member refusal a read-site result naming the receiver field,
   declared family and reason. Compile refusal at that read is acceptable until
   a runtime descriptor/probe is available; unread unsupported fields must not
   reject a tag-valid cast. Read support must precede enabling lazy admission.
4. Resolve a deferred member contract on demand; retain the selected child view
   for transitive reads. No first-member selection, blanket object-tag proof,
   null/undefined conflation or phantom-brand erasure is allowed. Native slot
   probes still need lane 1's readiness and presence machinery, not a copy.

Lane 4's helper oracle passes ordinary and viewed Box values to the same helper.
The malformed boolean payload stops at value.unsupported with exit 70 in both
backends. A mutant removing only that helper Property.View marker runs valid
release code and is caught by the pinned refusal. The unsupported mixed-union
helper remains a compile refusal. This proves the existing scalar helper guard,
not completion of lazy mixed-union/nullish/optional admission.

Lane 1 9b85eada was fetched; merging it into this lane conflicts in the shared
cast.go and the plan. The merge was aborted without editing shared compiler
files. Lane 2 remains merged at 0b141c26. Lane 3 c1f4c5a7 is used in an isolated
measurement worktree rather than merged through those conflicts. The lane 1
owner must reconcile cast dispatch before a landed/re-greened integration or a
completion date for its nullish/optional families can be claimed by lane 4.

### Lane 4 selector checkpoint from bc87b103

The user's new work queue is frozen at 213 `(type, field)` pairs and 1,111
explicit read sites, ranked in lane4/pair-progress.json. It does not silently
remove branded/intersection or other-owner obligations. The first two pairs,
Identifier.escapedText and Symbol.escapedName, account for 448 reads and require
__String brand proof; accepting every runtime string is not that proof.

Lane 4 now supplies runtime/view_unions_mixed.c/.h and
javascript/view_unions_mixed.go. The entry points are:

- C: adamic_view_mixed_union_select(snapshot,members,count,match,context,
  expression,declared), returning the selected member index.
- JavaScript: adamicViewMixedUnionSelect(snapshot,members,match,expression,
  declared), included through javascript.MixedUnionRuntime().

The snapshot has an explicit logical kind and borrowed normalized payload.
Logical kinds are neither raw heap kinds nor physical field storage bytes.
Null and undefined have distinct logical kinds; the selector never manufactures
that distinction from a null pointer. Shared probes must establish presence,
readiness and kind before entry. Selection transfers no ownership or readiness
proof. Convert/retain only after it succeeds, preserving the member contract on
later field/element/call reads and aliases.

Member records carry logical kind, optional literal constraint and a nonzero
shared contract id for reference members. The adapter match callback must be a
complete, pure, non-panicking contract test using shared metadata/probes, never
getters or callable bodies. Missing reference adapters/contracts and unknown
logical values do not pass. Failed alternatives allow later members to be tried.
An unavailable member contract must stay a named read-site NotYet/deferred
obligation until its adapter is available; it cannot be emitted as a complete
member that guesses conformance from a heap-kind match.

Additional concrete lane 1 wiring required, alongside lazy admission:

1. Include view_unions_mixed.h in native program assembly, and append
   javascript.MixedUnionRuntime() to its helper text before field read emission.
2. Route union-typed reads with complete member descriptors to the selector
   after the already-evaluated receiver's shared normalized slot probe. Supply
   expression text and the exact declared union. Apply optional/required field
   presence and readiness through the common path, not through this selector.
3. Supply the shared registry member-contract matcher for object/array/Map/
   callable alternatives. No adapter may panic on a failed alternative or copy
   readiness traversal into lane 4. Do not admit unsupported deferred bodies.
4. Feed the selected member id into the existing propagation/read path, retaining
   transitive checks. Lane 4's four-family component oracle is not authorization
   to admit source casts before these paths are covered.

The component oracle covers string|number|undefined, string|Identifier, two
untagged objects with different text contracts, and false|string|undefined.
Each has source .a Node controls for every member plus a wrong-value exit-70
pin for both implementations. Six independent release mutations are caught:
skip member testing, take the first untested member and skip object transitive
matching, separately in C and JavaScript. This proves selector semantics;
frontend/source integration remains pending and removes zero pairs.

Planning target for all 213 integrated pairs: October 16, 2026, 17:00 MDT
(23:00 UTC), conditional on lane 1 lazy admission plus normalized slot probes,
and the required brand/schema proof support, arriving by October 9. This is a
conditional working estimate, not an unconditional promise to erase phantom
brands or admit incomplete dictionaries/callable signatures. Completion evidence
and remaining pair/read totals are updated at every push.

## Untagged object union unit, October 7

Worker branch `codex/views-untagged-object-unions` starts at lane 4 583d19b7.
The user authorizes this plan addition. Own only new files named
`view_unions_untagged.go` and `view_unions_untagged_test.go` in lower/native/
JavaScript, runtime `view_unions_untagged.c/.h`, oracle
`checked_views_untagged_unions_test.go`, and the new
`stage3/interface-downcasts/untagged/` subtree. Existing mixed-union files,
shared contracts/dispatch/IR/readiness and counts remain with their owners.

Frozen demand: 84 pairs and 186 explicit reads. Rank by reads descending,
then checker receiver type id and field. Family labels overlap and include
narrowed/intersection receivers and array/callable alternatives; keep these in
the queue rather than silently reclassifying them. Component tests do not
remove pairs. A pair completes only with actual source read lowering and both
backend positive/negative pins plus semantic mutants.

Required owner handoff: shared lazy read dispatch supplies the evaluated value,
its union member contracts and non-panicking presence/readiness/logical-kind
probes. Member-specific kind tags select candidates, never certify payloads.
Structural alternatives must permit later candidates after a failed match;
selected member contracts must survive nested reads and aliases. Reuse existing
mixed-union selection and shared probes rather than duplicate readiness state.
Any hook needed in an existing file will be recorded here before implementation.

Working whole-family estimate: October 12, 2026 UTC, conditional on lazy read
hooks and normalized probes arriving by October 9. Unsupported descendants and
other-lane alternatives remain named read obligations; no cast-time eagerness
or trusting Unknown is permitted. This supersedes the provisional October 10
estimate made before reading the lane handoff documents.
### Lane 4 scope split, October 7, 2026

This ruling supersedes lane 4's earlier all-213 assignment and delivery estimate.
Lane 4 on `codex/views-mixed-unions` now owns only these two families:

| Family | Assigned pairs | Read sites | Delivery target |
| --- | ---: | ---: | --- |
| __String primitive phantom brands, including nullish members | 30 | 543 | October 9, 2026, 17:00 MDT (23:00 UTC) |
| Mixed primitive unions | 63 | 690 | October 13, 2026, 17:00 MDT (23:00 UTC) |

These are the user's assigned family totals, not an additive partition: branded
strings also occur in the mixed-primitive census. Neither total denotes completed
source integration. The old 213-pair ledger remains historical evidence until
exact pair membership is reconciled with this split; no completed pairs are claimed.

Object-plus-primitive unions belong to lane 4b, worker 01a11877-e6c0, on
`codex/views-object-primitive-unions`. Untagged object unions belong to lane 4c,
worker 01a11878-0860, on `codex/views-untagged-object-unions`. Non-brand
intersections belong to lane 7, worker 01a11878-4afa. Lane 4 stops building those
families. Its existing selector component evidence remains available to those
owners; it does not establish their compiler integration.

For __String, the latest ruling explicitly erases the primitive phantom brand
and checks string membership at each potentially viewed read. This supersedes
this document's earlier requirement for runtime brand proof and its prohibition
on phantom-brand erasure. Merge `codex/phantom-brands` at d90994da, retaining its
15:28 overload rule. A nullish alternative needs its own membership check.

Lane 4's first compiler checkpoint remains the two leading pairs with 448 reads,
then the most-read mixed primitive shapes. Wire the C and JavaScript selectors
from 583d19b7 into the compiler for this scope, coordinating shared hooks with
lane 1 and merging lazy admission as it lands. Presence/readiness still use the
shared machinery; Unknown flow retains checks, including helper reads. Every
completion needs actual source fixtures held to Node, both backend refusal pins
and semantic mutants. Report family pairs and reads remaining after every push.

Working target for both families is October 13 at 17:00 MDT. Nullish encoding,
optional presence and shared dispatch merges remain integration risks; report an
observed blocker immediately rather than counting a component test as completion.

## Integration branch checkpoint, October 7, 2026

The dated lane checkpoints above retain their original evidence and limits.
The integrator has merged lazy owner e6aec805 first on codex/views-integration;
its first published SHA is f1c919701ec9b42b78e104fe9cc1ce96db9d7d09.
Lane-4 primitive phantom scalar reads are reconciled with that lazy entry point;
standalone selector evidence does not imply production union admission.
See stage3/interface-downcasts/integration/REPORT.md for conflict choices,
validation, reproduced baseline failures and mutants.

Lane 5 helper checkpoint: `completeViewCallableShapeContract(node,target,id,build)`
is a reached-read hook, not an admission hook. It fills the placeholder's existing
Parameters/Result slots for fixed, nongeneric, single signatures and propagates
child failures. Optional/rest/generic/overload signatures remain unknown until
implemented. Native `adamic_view_callable_shape(value,recorded,expected,expression,
optional)` and JavaScript `adamicViewCallableShape` compare producer metadata;
they introduce no call convention. Native NULL is confirmed undefined from the
shared presence/nullish check, never permission to treat null as absent.
Lane 1 wires completion only at reads and wires shape checks after readiness;
the closure owner supplies metadata at creation. These helpers alone certify
zero census reads. Lane 1 tip ab4d6f90 has 13 shared code/count merge conflicts;
per the lane 2 handoff, lane 1 resolves and publishes the reconciled lazy tip.
Both backends also provide `emitViewCallableShape(value,recorded,expected,
expression,optional)`, inserting the helper around already evaluated operands.
It returns the same callable identity and adds no ownership or receiver wrapper.
The caller must preserve the field-read's existing lifetime and route subsequent
calls through the reconciled convention.
Producer handoff locations on this base: native closure creation is in
`emit_expressions.go` (adamic_closure_new), and JavaScript creation is in
`javascript.go` (new AdamicClosure). Both are shared owner files. Use the
implementation Function.Parameters/Returns, excluding a hidden receiver from
advertised arity. Function.Returns zero means void in existing IR; it must not
be copied as the helper's zero/unknown signature code. A void descriptor needs
an explicit agreed representation before enabling that family. Count-row and
real-source read-time failure registration remain central owner handoffs.

## Lane 5 read adapter handoff to views-integration

Coordination supersedes the individual-lane merge instructions above: lane 5
merges only codex/views-integration after its integrator publishes a tip. On this
checkpoint that remote branch is not yet published. Do not merge fetched lazy or
closure tips directly. The 11,648 explicit witnesses remain an Unknown-fallback
upper bound; remeasure what viewed values can reach after lazy provenance lands.

New lane 5 territory: internal/lower/view_callables_read.go and tests;
internal/native/view_callables_signature.go and tests;
internal/javascript/view_callables_signature.go and tests;
lane5/run-read-adapter-mutants.py. JavaScript's owned viewCallablesRuntime now
includes the shape runtime. The native certificate emitter includes its header
in generated declarations. No shared owner file was changed.

Concrete integration hooks requested:

- At the shared callable read with KnownViewed or Unknown provenance, call
  `l.prepareViewCallableRead(node, declared)` and propagate its error at that read.
  Install the returned id on Property.ViewContract. This strips only optional
  undefined and builds representation descriptors without walking object members
  of argument/result types. It must not be called from cast admission or from the
  global same-name field scan. Keep the existing refusal until all its reads
  have equivalent runtime coverage; removing the scan alone is unsound.
- After shared presence, readiness and storage-kind validation, snapshot the
  callable, then use `e.emitViewCallableCertificate(property, value)` on both
  backends before capturing or calling it. Native `value` must be an evaluated
  C name; JS snapshots with a one-argument arrow wrapper. Check every read path,
  including native viewField and dynamic method-call lookup. Existing owned
  callable lookup still follows the previous proof path until these hooks land.
- Optional absence must be returned by the shared read, not passed through the
  old required-field helper. Native NULL may represent confirmed undefined only;
  present null cannot be excused as optional. Pass Property.Absent appropriately.
- The adapter selects immutable producer descriptors by generated closure code
  identity and IR parameter/result representations. It neither guesses arity
  from calls nor uses declared view metadata as a producer certificate. Unknown
  code, Receiver closures and class method thunks remain uncertified. Use the
  reconciled closure ABI for eventual receiver/method certificates; do not treat
  the current fixed ordinary-closure adapter as coverage for them. Void remains
  unknown. Table emission is deliberately a correctness-first prototype, with
  per-read tables; compact it after the integration protocol is stable.

All adapters are tested, but none certifies a corpus pair without the shared
source-read hooks. Three new semantic mutants catch forged producer descriptors
(native and JavaScript) and eager argument/result object-member traversal.

## Lane 4b: object plus primitive unions, October 7, 2026

Branch `codex/views-object-primitive-unions` starts from lane 4 tip 9ecdda53,
which contains selector checkpoint 583d19b7. Assigned demand is 73 pairs and
267 explicit reads. Own only new `internal/lower/view_unions_object_primitive.go`
and its tests, `internal/native/view_unions_object_primitive.go` and its tests,
`internal/javascript/view_unions_object_primitive.go` and its tests, runtime
`view_unions_object_primitive.c/.h`, oracle
`checked_views_object_primitive_unions_test.go`, and
`stage3/interface-downcasts/lane4b/` fixtures, ranking, scripts and logs.
This plan addition is explicitly authorized by the unit. Existing lane files,
common dispatch, readiness and representation remain with their owners.

Reuse lane 4's mixed-union selector and shared contract ids. Each read selects
primitive membership by logical runtime kind or the reference member by its
contract; preserve that contract on subsequent object/element reads. Unknown
provenance retains checks, including helpers, generics, callbacks and fields.
Unsupported members remain obligations at reachable reads, never cast failures.

Required owner handoffs: lane 1 supplies a presence/readiness-aware normalized
slot probe, a pure non-panicking reference-member matcher, selector dispatch
and selected-member propagation; lazy admission resolves read contracts without
eager descendant refusal. Lane 2 supplies array/Map/callable member adapters
where required. Existing selectors are components, not evidence of source
admission. No lane 4b helper may guess structural membership from a heap tag.

Rank individual pairs by reads and aggregate identical declared shapes for
fixtures. First shapes: JSDoc string|NodeArray comment (24 pairs, 63 reads),
CommandLineOption literal-string|Map type (1 leading pair, 33 reads),
true|Node|undefined, then string|DiagnosticMessageChain. Every completed shape
needs source Node controls, both backend exit-70 field/type/found pins and
independent skip-check, wrong-shape and dropped-nested-check mutants.

Working whole-family target: October 14, 2026, 23:00 UTC, conditional on shared
lazy admission, normalized probes and member adapters arriving by October 9.
The initial October 9 estimate was revised after inspection of pending hooks.
No integrated pair is complete at this territory checkpoint: 73 pairs and
267 reads remain. Lazy-admission branch was not yet present on origin.

### Lane 4b resumed on integrated lazy admission, October 7

Merged integration ba59427ccc7afecae29a305c41e6e9c7867e5610. Superseding
inventory: 42 candidate pairs / 181 candidate reads from lazy/census; exact
production reachability remains unmeasured pending checker-clean TypeScript.
The old 73/267 table remains historical. Ranked ledger and source controls are
in lane4b/resume. No original tsc pair is yet marked complete.

The lead now authorizes minimal shared hooks. This lane adds named dispatch to
objectPrimitiveViewType in lower/object.go and lower/interface_cast.go, and
viewObjectPrimitive in native/view_fields.go and javascript/javascript.go.
The two admission guards allow one structural object plus scalar primitives
and optional undefined; accessors remain refused. Backend dispatch preserves
the shared union contract and descendant read obligations. New runtime files
normalize concrete scalar and boxed slots, check presence/readiness, select
literal primitive membership, and retain the single object alternative.
Selection of that alternative is not a structural-conformance certificate:
its fields remain lazy obligations and every reachable descendant read must
use the shared view machinery. No object member is erased into Unknown.

Production source controls cover reduced true|Node|undefined and string|Chain
shapes, matching the first two ranked shapes (five candidate pairs / 43 reads).
They are not original-pair coverage. Both backends pin outer and nested wrong
values, and valid values; the optional shape also pins undefined. Independent
release mutants remove outer selection, accept a wrong literal/nested scalar,
and remove a transitive read check. Arrays, Map, callables, intersections,
multiple object alternatives, indexed reads, inherited static slots and packed
optional numeric/boolean storage remain outside this adapter checkpoint.

Revised whole-family working date: October 16, 2026, 23:00 UTC. This includes
remaining member adapters and original-pair witnesses; exact reachability will
still require the checker-clean program. Only integration is merged, and the
lead forwards this lane's tip to the integrator for hunk reconciliation.

### Lane 4b scalar and array union checkpoint, October 8

The next direct shape is string|number|PseudoBigInt (rank 4, 10 candidate reads).
Rank 3 is an object union with an intersection and remains uncovered. Array
alternatives now dispatch through the existing shared lazy element contract;
source controls model string|NodeArray<JSDocComment>|undefined using a readonly
array alias. Full NodeArray inherited metadata is not certified by this alias.
These two added shapes bring reduced controls to 22 candidate pairs / 103 reads;
20 candidate pairs / 78 reads have no reduced shape control. Original-pair
witnesses remain pending for all 42 pairs / 181 reads, and exact production
reachability remains unmeasured.

Minimal added shared hook: lower/object.go calls objectPrimitiveBoxedField for
ir.Union property initializers. The helper admits physical boxed storage only
for scalar unions or the supported single object/array alternative, preserving
runtime tags and source slots. It does not widen writes or prove payload types.
Both positive source controls exercise boxed union field producers and both
backends check the member at the read. Native sanitized/leak-checked runs pass.
Independent outer removal, wrong-member acceptance, wrong nested-shape acceptance
and dropped nested checks all fail pinned refusals in executable release code.

Revised working whole-family date: October 13, 2026, 23:00 UTC, brought forward
from October 16 after this production checkpoint. Remaining risks include true
NodeArray representation/metadata, object intersections and unions, dynamic
keys, tuple/element reads, callable alternatives and original-pair context
witnesses. This is a delivery estimate, not completed production reachability.

### Lane 4b original-pair witness checkpoint, October 8

Original declarations are generated outside the repository from TypeScript pin
050880ce59e30b356b686bd3144efe24f875ebc8 using original/prepare.cjs. No cohere
code is copied. The generator rejects tracked source drift, audits every UTF-16
read-site slice for the two leading pairs, emits declarations, and records their
hashes and complete SourceFile/Node/Diagnostic/DiagnosticMessageChain field sets.
The oracle binds type-only imports to those declarations, runs normal strict
load diagnostics, and checks every root/member field name against the manifest.

Certified standalone original field-contract pairs: SourceFile.externalModuleIndicator
(20 candidate reads) and Diagnostic.messageText (15 candidate reads). These are
complete original declarations, with valid/wrong/absent controls as applicable,
both-backend refusal pins and independent semantic mutants. SourceFile also has
generic, callback and stored-alias controls. Its optional missing field returns
undefined; an uninitialized present slot still refuses. These are pair-contract
witnesses and site provenance audits, not execution of the 35 whole-tsc contexts
or exact production allocation reachability. Remaining: 40 candidate pairs /
146 candidate reads. Whole-family working date remains October 13, 23:00 UTC.

Only lane-owned emitter/runtime files changed for original optional admission:
pass receiver optionality and field absence separately; allow a missing optional
slot while retaining the shared readiness guard. Existing shared hook hunks remain
those already listed. Original schemas retain all unread unsupported descendants
as lazy obligations. No blanket shape certificate is claimed.

## Centralized checked-view integration

The lead assigned Codex 01a11882-3830 on codex/views-integration to merge every
lane tip and resolve conflicts once, hunk by hunk. This supersedes all direct
cross-lane merge instructions above, including the Lane 1 cast-conflict handoff.
Each worker pushes only its own branch and obtains another lane's code by merging
the newest codex/views-integration SHA when needed. Workers blocked on lazy
admission build against that integration branch where possible; otherwise they
report the blocker and rest until the lead wakes them. Lane 2 remains native
arrays only, and Lane 5 retains callable ownership and certification hooks.

### Untagged unit selector handoff

The new `lower.UntaggedViewMembers(contracts, unionID)` returns ordered member
contract ids and each member's own required finite scalar tag fields. Members
without tags retain structural matching. Optional tags never exclude a member.
No descendant contract is replaced or certified by these tag descriptors.

Concrete owner wiring needed before source admission:

1. The lazy read dispatcher resolves this plan on an object-union read, leaving
   other-family or deferred members as named read obligations. Existing
   `viewUnionFields` must stop demanding a common finite discriminant for this
   family. Do not call the new plan builder at the cast as an eager gate.
2. Native emits `adamic_view_untagged_member` descriptors and calls
   `adamic_view_untagged_union_select` with the common evaluated snapshot,
   non-panicking slot probe and complete member matcher. Include
   `view_unions_untagged.h`; runtime embedding already includes its new files.
3. JavaScript appends `MixedUnionRuntime()` and `UntaggedUnionRuntime()` and
   calls `adamicViewUntaggedUnionSelect` with the same contracts and adapters.
4. Both selectors return the selected contract id. The shared dispatch must
   keep that id through aliases, helpers, generics, callbacks and stored values,
   and check each later read; selection is no permission to erase checks.
5. The three ordinary absent controls have measured count rows in
   `untagged/count-row-additions.md`, for the shared counts owner to register.

Merging lazy admission 5002bfe0 into d15b4206 conflicts in the plan, cast.go,
interface_cast.go, readiness.go, view_objects.go and native/view_fields.go.
The attempt was aborted without editing those shared compiler files. Lane 4
9ecdda53 merged with both plan additions retained. Source fixtures still
compile-refuse at their casts for lack of a common finite discriminant. The
component selector is therefore not a source admission checkpoint: all 84 pairs
and 186 reads remain pending. The nested mutant exercises the supplied component
matcher; compiler-wide transitive read propagation remains unproven by this unit.

Coordination superseded by the user: stop direct lane merges. Dependency code
now arrives through `codex/views-integration`, whose owner resolves shared hunks.
The earlier merge observations above remain historical evidence. At this unit's
check, `git ls-remote --heads origin codex/views-integration` returned no ref.
Do not resolve another lane's compiler conflicts locally while waiting for it.


## Dictionary contracts unit, October 7, 2026

Branch `codex/views-dictionaries` starts at lane 4 `9ecdda53`, after `583d19b7`.
This unit owns new `internal/lower/view_dictionaries.go` and tests,
`internal/native/view_dictionaries.go` and tests,
`internal/javascript/view_dictionaries.go` and tests,
`internal/native/runtime/view_dictionaries.c/.h`,
`internal/oracle/checked_views_dictionaries_test.go`, and
`stage3/interface-downcasts/dictionaries/` (ranking, fixtures, logs, reports).
The user explicitly authorizes this plan addition. Existing lane files and
protected orchestration files remain their owners' territory.

The frozen family is 409 pairs and 2,256 explicit reads. Numeric-indexed
NodeArray types appear in this dictionary inventory; preserve their pair ids
and demand rather than claiming they are supported or silently reassigning them.
Array dispatch must precede dictionary dispatch. String-key objects with named
properties (CompilerOptions) need both named field and index contracts.

Reuse `runtime/record.c`: its record is an ordinary counted object wrapping the
existing ordered Map. The current base contains that runtime and its component
oracle, but no record IR kind or frontend record operations. Its untagged
adamic_value slots and reference_values bit cannot distinguish number from
boolean or prove a nested object's logical type. A target cast must never
supply the missing source certificate.

Concrete handoffs required from the shared owners before source admission:

1. Reconcile lazy-admission `5002bfe0` with this base in cast.go,
   interface_cast.go, readiness.go, view_objects.go and native/view_fields.go.
   The attempted merge was aborted with all shared code restored. Initial setup
   failed on checker APIs; restoring both exact submodule checkouts recovered
   the compiler build without a committed pin or loader change.
2. Add ViewDictionary to the common contract kinds, with Element referring to
   the shared recursive child registry. Add a viewDictionaryContractHook to
   view_contracts.go after array/callable dispatch, before ordinary object
   admission. The lane helper will use string index information, intern before
   recursion, and preserve named fields and unsupported/deferred children.
3. Supply the existing record lowering through shared IR and expression dispatch.
   Dictionary field reads must preserve the record identity and element contract
   on aliases, helper parameters, returns, callbacks, generics and stored fields.
   An indexed read with viewed or Unknown provenance must call the dictionary
   read hook even when no lexical cast appears in its function.
4. Supply a source-certified normalized record slot probe: presence, initialized
   state, logical kind, borrowed payload and child schema/contract. Reuse existing
   readiness, null/undefined and object-view metadata. Do not infer kind from the
   target or add another record table/bitmap. Every alias write/delete/overwrite
   must maintain source facts; Unknown keeps the check. Existing get returns
   untagged storage and is insufficient for a runtime type check.
5. Wire emitViewDictionaryRead in both backend indexed/property read dispatches,
   passing the evaluated receiver/key, element contract, expression and declared
   type once. A missing key is undefined only when the declared read permits it;
   a present undefined is distinct from an uninitialized slot. Preserve current
   own-key prototype refusals until inherited semantics are implemented soundly.
   Nested object results feed shared transitive view propagation at later reads.

Whole-family working estimate: October 16, 2026 at 23:00 UTC, conditional on
these hooks and reconciled lazy admission by October 9 at 23:00 UTC. The earlier
October 14 estimate was provisional before inspecting the source representation.
No unconditional completion date is defensible while shared source probing and
record compiler lowering are absent. No pair is completed by a ranking or a
component oracle. Each push records actual source support separately.

Coordination update: consume other lanes only through `codex/views-integration`.
Both earlier direct merges were aborted. The integration ref was absent at this
unit's final remote check. Dictionary source admission remains blocked on the
registry, record compiler operations, source probes and transitive indexed-read
hooks above. The recovered exact submodule checkout builds successfully; setup
rerun exits 0 in 24.177s and the existing record package passes in 11.344s.


## Lane 7: nonprimitive intersection field contracts, October 7

Branch `codex/views-intersections` starts at lane 4 tip 9ecdda53, after
583d19b7, and merges phantom-brands d90994da without copying cohere code.
The user authorizes this lane-specific plan addition. Own only new
`internal/lower/view_intersections.go` and `view_intersections_test.go`,
`internal/native/view_intersections.go` and `view_intersections_test.go`,
`internal/javascript/view_intersections.go` and `view_intersections_test.go`,
`internal/native/runtime/view_intersections.c/.h`,
`internal/oracle/checked_views_intersections_test.go`, and
`stage3/interface-downcasts/lane7/` fixtures, ranking, evidence and reports.
Existing shared files remain with their owners, including counts.md.

The frozen family ledger contains 198 pairs and 1,145 reads. Primitive brands
are tracked separately, never counted as lane 7 completions. Lane 4 owns
__String (30 direct-display pairs, 543 reads); other primitive branded aliases
such as Path need ownership reconciliation. Type displays alone cannot establish
that an intersection constituent is phantom. Reuse phantomField from the merged
brand implementation and preserve the overload result rule.

Priority follows the ranked pair ledger, starting with object refinement fields,
then nested intersections, interface composition, and array/callable overlaps.
An intersection checks every runtime constituent, ignoring only a proven phantom
constituent. A bare object heap tag never proves constituent fields. Nested
results retain their intersection contract through helper, generic, callback,
container and alias reads. Optional absence and readiness use shared machinery.
Unknown provenance retains checks; unread unsupported children do not refuse casts.

Concrete handoff to the lazy-admission/shared-dispatch owner:

1. Route TypeFlagsIntersection before generic representation/callable dispatch
   through lane 7's viewIntersectionContractHook. The lane will supply the hook
   and builder in its new lower file. Reserve recursive ids in the common registry.
2. Supply a distinct intersection descriptor (all member ids, not union selection)
   in common IR, or an explicit all-members discriminator. ViewUnion cannot encode
   conjunction. ViewObject flattened fields alone must preserve repeated-field
   conjunction and each member's readonly/optional obligations.
3. At each potentially viewed intersection read, validate shared presence/readiness
   once and hand the evaluated snapshot plus all constituent ids to the lane's
   matcher. Keep child contracts on subsequent reads; no getter reevaluation.
4. Expose a pure non-panicking shared member matcher and normalized slot probes.
   They must distinguish missing, uninitialized, undefined and null. Array and
   callable constituents require the respective owners' adapters. Unsupported
   members remain named read-site obligations, not successful Unknown contracts.

Observed base has no intersection hook or descriptor. Lazy-admission branch is
not yet published at this checkpoint. These handoffs are prerequisites for source
integration; a standalone helper must not be counted as completed pairs.
Working estimate for the whole nonprimitive family: October 16, 2026, 23:00 UTC,
conditional on shared hooks and lazy admission by October 9. This is a planning
estimate, not an observed delivery. No runtime completion is claimed yet.

Lane 7 component API checkpoint: `(*lowering).viewIntersectionContracts(node,
target,build)` returns all nonphantom constituent ids, propagating unsupported
member errors. It does not reserve an aggregate descriptor. The shared owner
should call it from the intersection dispatch hook after defining the all-members
IR discriminator. Runtime entry points are `adamic_view_intersection_matches`
and `adamic_view_intersection_require`; JavaScript helper text is exported by
`javascript.IntersectionRuntime()`. They reuse lane 4's normalized snapshot type
and accept a pure member matcher from shared dispatch. Production source admission
is not enabled. Four source Node controls and six C/JS adapter mutations pass;
these do not prove frontend transitive propagation or optional/readiness behavior.
The ranked frozen ledger still has zero completed pairs and zero completed reads.

Lane 7 coordination update: the user now assigns all cross-lane merges/conflict
resolution to `codex/views-integration` (worker 01a11882-3830). Lane 7's lazy,
callable and shape merge attempts were aborted without editing conflict hunks.
No individual lane is merged after this ruling. The integration branch was not
published during this checkpoint. Source c/js compilation of lane7/good.a still
refuses Named & Counted at value, so zero pairs/reads are complete. Native and
JavaScript package tests pass; lower retains the documented mixed-array adapter
failure. The lane-owned component work, six executable mutants and exact blocker
handoff are recorded in lane7/REPORT.txt. Rest pending integrated lazy admission.


## Lane 5 source hooks on ba59427 (October 8)

The user's hook rule supersedes the prior handoff-only restrictions. Lane 5
merges integration only and owns these minimal named shared hooks:

- lower/object.go: prepareViewCallableProperty attaches fixed scalar signature
  descriptors at reads; casts remain lazy. lower/interface_cast.go recognizes
  optional callable unions through callableViewContract.
- native/view_fields.go: emitViewCallableCertificate runs after presence,
  readiness and kind checks. native/emit_functions.go checks resolved viewed
  method pointers through emitViewCallableMethodCertificate before calling.
- Owned native and JavaScript view_callables files check dynamic method lookup
  and ordinary closure fields. New native/view_callables_methods.go selects
  independently recorded method signatures by actual thunk identity.
- lower/invariance.go: viewCallableVoidMarker treats a discarded function result
  as a value contract, while widenedProperties retains reverse writable-slot
  checks. New lower/view_callables_marker.go validates immediate discarded,
  zero-argument marker calls against their original scalar source signature.
  lower/cast_proof.go and lower/expression.go use that narrowly named proof;
  stored erased-marker calls retain their existing refusal.
- oracle/checked_views_arrays_test.go updates callable signature diagnostic pins;
  oracle/checked_views_lazy_test.go pins runtime refusal for now-supported scalar
  helper reads. oracle/interface_cast_counts_test.go registers viewCallableCounts.
  New callable source/count tests and lane5/group1 and lane5/marker fixtures are
  lane 5 territory. counts.md records only the measured added fixture rows.

Coverage ledger labels counts as candidates. Group1 covers the two Program
string-directory contracts, 11 static candidate reads; 306 pairs / 1,492 static
candidate reads remain. Actual corpus reachable-read certification is unmeasured.
Higher-ranked intrinsic, optional/rest, overloaded and aggregate-return shapes
remain unclaimed. Working family date: October 12, 2026, 18:00 UTC, subject to
remeasurement after the checker-clean exact table and closure reconciliation.

## Lane 5 stored marker calls and upstream witnesses (October 8)

New minimal hooks: ir/views.go adds DiscardResult to callable contracts;
ir/ir.go adds CheckedDiscard and its contract/label to CallClosure. lower/expression.go routes erased
never-rest calls through lower/view_callables_marker.go, requiring a zero-argument
call whose result is discarded. Native emit_functions.go dispatches that flag to
owned view_callables_discard.go; it rechecks producer identity and arity and frees
reference results according to the independently recorded result representation.
Owned signature/read adapters recognize only this explicit discarded-result
contract. Zero remains unknown; known void is separately recorded as metadata.
Normal valued result contracts remain exact. javascript/javascript.go dispatches
the same call flag to its owned view_callables_discard.go. No closure ABI is changed.

Original candidate witnesses are verified against upstream TypeScript commit
050880ce59e30b356b686bd3144efe24f875ebc8 in an independent source checkout, never
by copying cohere files. Intrinsic and stored-callable evidence will be reported
separately and counts remain candidates until exact reachability is measured.

Lane 5 next group adds known-void signature metadata at the owned read adapter
and runtime shape selection, including native method producer certificates.
Void remains distinct from unknown and discarded-result contracts. Original
FileWatcher.close witnesses pin this fixed zero-argument shape.
### Dictionary boxed-record helper checkpoint after ba59427c

The dictionary lane supplies `runtime/view_dictionaries.c/.h` and
`javascript.DictionaryRuntime()`. Native `adamic_view_dictionary_read` consumes
only source-certified existing records whose table values use existing union
boxes (`reference_values=true`). It classifies each selected box at the read,
checks a logical-kind mask and returns a borrowed normalized value with the
nonzero child contract on reference results. JavaScript's corresponding helper
checks an own data property without invoking getters. Neither scans unread keys.
Native null has no distinct boxed encoding yet and stays unsupported.

Shared wiring remains required: a source-certified record producer, descriptor
registration ahead of unsupportedViewFamily's dictionary fallback, record IR and
indexed-read dispatch, and propagation of the result's child contract. Existing
raw scalar records must keep their current representation; they cannot be passed
as boxed_storage=true on the word of a target cast. Shared source dispatch must
retain deferred refusals until the complete producer/read path exists.

The lazy census reports 18 dictionary-contract pairs / 174 candidate reads and
14 dictionary-read pairs / 66 candidate reads; their deduplicated union is 29
pairs / 228 reads. This is a candidate table, not exact runtime reachability.
ADAPTED-CENSUS.md explicitly says runtime reachability is unmeasured while
checker diagnostics prevent production IR. Do not replace that Unknown with zero
or claim exact pair completion from component evidence.

`internal/lower/view_dictionaries.go` now declares viewDictionaryContractHook and
prepares recursive index/named-field descriptors, rolling back new ids on failure.
The root remains Unsupported dictionary source dispatch until shared source
wiring is complete. New records are not created by this helper. It does not change
refusals.go's unconditional index-signature rule or install a read dispatcher.
The native helper currently supports broad kind membership on source-certified
boxed tables; finite literals and full array/callable/union/null contracts require
their complete owning adapters, never a kind-mask shortcut.

Revised whole-family working date: October 19, 2026 at 23:00 UTC, conditional
on source producer and shared dispatch handoffs by October 9. Group1.md records
the proven helper paths and all unimplemented source obligations. Candidate
progress remains 29 / 228, with exact production reachability unmeasured.

### Shared-hook coordination amendment, October 8, 2026

The user now authorizes each lane to add its own minimal shared-code hooks
listed under that lane's plan section. This supersedes earlier owner-only
wiring restrictions for those named hooks. The integration worker reconciles
each overlapping hunk individually, preserving all refusals and Node agreements.
Lanes still consume other lanes through codex/views-integration; a hook or
component alone is not evidence of complete source admission.

## Lane 2 native array hooks, October 8

Under the lead's October 8 rule, lane 2 adds minimal shared hooks directly;
`codex/views-integration` reconciles overlapping hunks. Callables belong to lane
5 and are unchanged. The native array family estimate is October 10, 12:00 MDT.
Use the frozen candidate inventory until checker-clean exact reachability exists.

- `viewOptionalArrayContract` in `view_array_adapter.go`, dispatched from
  `strictViewContract`, preserves the array/element descriptor through undefined
  alternatives. Published in 5ca21a57.
- `viewArrayBase`, `viewArrayOwnProperties`, and `viewArrayElementType` in
  `view_array_types.go` recognize instantiated array ancestry and array/record
  intersections. Representation, `elementType`, collection iteration,
  `viewDataType`, `unsupportedViewFamily`, `strictViewContract`, and existing
  array read metadata call these hooks. Tuple/class ancestry and overridden
  intrinsics are excluded.
- `internArrayViewContract` adds own-field child descriptors without scanning
  elements. Unsupported children remain lazy read obligations.
- `viewArrayRecordProduction` is called from `objectCall`: Object.assign of a
  fresh array literal and one exact scalar record produces a fixed own-field
  shape. Other array assignments refuse explicitly.
- `viewArrayOwnReceiver` is called from `property`; `viewArrayOwnWriteReceiver`
  from `setProperty` and `updateProperty`. Own writes register the shared original
  slot certificate, including finite literal constraints.
- New IR `ArrayRecord` and `ArrayProperties` carry production and selected own
  storage through the existing walkers. Native `emitViewArrayRecord` and
  `emitViewArrayProperties`, and JavaScript counterparts, are dispatched from
  native `evaluate` and JavaScript `value`. Native reuses the existing owned
  `adamic_array.properties`; no header layout or destructor change is needed.
- `checkLazyViewReads` dispatches `ArraySearch` through its existing named
  `arrayRead` hook, as it already does for map/join/other selected consumers.

Fixtures live under stage3/interface-downcasts/lane2/node-array-*.a; their oracle
is TestCheckedViewNodeArrayRecords. This checkpoint covers scalar own-field
reads/writes and lazy object element reads. It does not certify reference element
writes, arbitrary array shape mutation, all consumers, or whole-tsc compilation.

Lane 5 integration nullable-callable hook: native/view_nullish.go and
javascript/view_nullish.go invoke owned viewCallableNullishCertificate helpers
after the owner presence/readiness/kind checks. Non-null callable values retain
the same signature check; permitted null/undefined preserve their identities.

The nullable-callable reconciliation also attaches complete fixed signatures to
boxed Union reads in prepareViewCallableProperty. Otherwise the new same-name
proof exemption would admit nullable function reads without a signature check.
TestCheckedViewNullishCallableSignatureMutant now pins the runtime wrong-result
refusal in both backends rather than the older compile-time refusal.

## Lane 5 non-predicate scalar witnesses (October 8)

Lane 5 adds internal/oracle/checked_views_callable_scalar_witness_test.go and
fixtures under lane5/scanner and lane5/performance. These preserve original tsc
member signatures and call expressions with reduced receivers and implementations.
No shared production hook is needed for these fixed scalar signatures. The lane
counts helper adds these directories; candidate counts remain separate by census.

Lane 5 also pins gaps/mixed-scalar-boxed-result.a against Node. It reproduces
the predicates worker ABI boundary: a number-returning producer needs a boxed
union result adapter. This adds no blanket callable admission. Aggregate callable
results also need field-demand propagation before their shapes can be admitted.

## Lane 5 scalar and boxed callable adapters (October 8)

Lane 5 owns new native/view_callables_boxing.go and lower/view_callables_boxing.go,
plus oracle/checked_views_callable_boxing_test.go and lane5/boxing fixtures.
Minimal shared hooks: censusCallableSlotless admits Union's single reference word;
native/emit_functions.go routes closure invocation through the named boxed adapter.
The adapter selects actual producer code, converts borrowed parameters, releases
new parameter boxes on normal and exceptional returns, and converts owned results.
No closure calling convention or callable identity is replaced. Read certification
will use independently recorded representation member masks for union variance.

Callable union metadata hooks: ir.Function.CallableMasks is recorded by the
function signature producer; ir.ViewContract.RepresentationMask records requested
members. Native and JavaScript helpers compare result subsets and reversed
parameter subsets before admitting a read. Scalars retain their previous checks.

Callback adapter hooks use the same named helper in native emit_expressions.go,
emit_arrays.go, emit_maps.go, from.go, library_array_holes.go, and reuse.go. Both
normal and reused array maps, visits, reduce, from, Map/Set visits carry explicit
runtime input and result representations. Sort's existing direct-function
comparator is distinct and remains outside this closure adapter.

Sort callback hooks: native/view_arrays.go and emit_expressions.go use the
owned boxed comparator adapter; direct named-function comparators retain their
existing exact-element admission. lower/expression.go copies producer masks into
named function forwarders. The former mixed-result boundary is now an oracle
positive in its existing lane-owned test; counts adds the boxing directory.

JavaScript adds view_callables_boxing.go and a named call-dispatch hook in
javascript.go. It shares native unknown-producer refusal when union conversion
would require metadata; JavaScript otherwise preserves its normal values.
