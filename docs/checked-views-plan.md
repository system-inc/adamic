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
