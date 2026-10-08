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

### Untagged object union lane, integrated lazy source handoff, October 7

Merged the designated integration tip ba59427ccc7afecae29a305c41e6e9c7867e5610
by clean fast-forward. No individual lane was merged. The revised whole-family
estimate is October 14, 2026 UTC, conditional on owner hooks and a family-specific
allocation reachability inventory. lazy/ADAPTED-CENSUS.md explicitly calls exact
allocation reachability unmeasured. Its static object-union inventory is 228 pairs,
2,468 reads, including tagged unions. The old 84/186 queue is historical only.

Concrete three-hunk owner handoff is in
stage3/interface-downcasts/untagged/source-hooks.patch. It is tested using a Go
source overlay generated by source-overlay.py, without editing shared files:

1. lower/view_contracts.go: allow the demanded untagged union only when the
   lane-owned supportsUntaggedRead certifies usable member selectors. Members'
   own finite tags defer unread payloads. Structural selection requires complete,
   acyclic plain-data contracts. Other member families remain unsupported.
2. native/view_unions.go: after checking ViewUnion, dispatch only unions without
   a shared required finite tag through viewUntaggedObjectUnion. Use shared slot
   presence, initialization and physical representation metadata. No getter runs.
3. javascript/view_unions.go: the same dispatch preserves once-only evaluation,
   using shared readiness metadata and own data-property descriptors. Both
   backends refuse unsupported class/accessor selection rather than trust it.

Actual source fixtures test all three representative shapes through the compiler
and both backends under these hooks. They are reduced interfaces, not proof that
an unchanged tsc pair has lowered. Exact diagnostics are stored in source-refusals.json.
The plain source test reports the pending hooks explicitly until integrated; setting
VIEW_UNTAGGED_SOURCE_REQUIRED=1 makes missing admission a failure. Subsequent
field reads retain shared checks, not a new readiness table or a tag certificate.
The returned member id is currently selection evidence, not stored dynamic provenance.
No source pair completion is claimed until shared hook integration and its measured
per-pair fixture evidence are both present.

### Untagged object unions, production hook authorization, October 8

The user's rule change authorizes minimal shared hooks directly. Installed exactly
supportsUntaggedRead in lower/view_contracts.go and viewUntaggedObjectUnion in the
native and JavaScript view_unions.go dispatchers, as listed above. Actual source
fixtures and semantic mutants now run without any overlay. The former owner-wait
block is removed. Revised whole-family target: October 12, 2026, 23:00 UTC.
Continue against the supplied candidate inventory, explicitly labeled candidate
pairs and reads; checker-clean allocation reachability is a later audit. Common
finite-tag unions remain with shared dispatch, never charged as untagged completions.

Untagged discriminant refinement: new lane-owned internal/ir/view_unions_untagged.go
exports ViewUnionHasDiscriminant. Shared lower/view_contracts.go and the two named
backend hooks use it to require disjoint member literal sets; overlapping enum values
must use member selection. This is a minimal shared IR query, not a new representation.
The classifier uses the pinned stock checker through a cohere/TypeScript worktree,
with generated diagnostics and locked dependencies. Zero stock diagnostics; 61
candidate pairs/434 reads have no shared disjoint field, including adapter overlaps.
An own required scalar kind checks that kind contract, including open numeric enums,
without reading other enum-valued payload fields. It certifies no unread payload.

Own-data class support uses the same optional-field lookup and initialization/type
bytes for non-static class instances. JavaScript uses own data descriptors. No getter
or inherited static lookup runs during selection. Dedicated class-produced positive
and wrong-kind fixtures match Node/control expectations in both backends. Candidate
ranks 8, 11, 13 and 19 have selector-projection fixtures with exact wrong/nested pins
and eight semantic mutants; they represent 4 candidate pairs/166 candidate reads.
Remaining unvalidated selector obligations: 57 candidate pairs/268 candidate reads.
This does not certify every field of the unchanged compiler interfaces.

Untagged candidate group 2 extends own-kind projections to all 36 eligible pairs,
310 candidate reads, with a fixture set per pair in ranked order. Existing array
index dispatch already reaches the named object-union hook; it now uses the registry
name when callers omit ViewType, preserving exact expected-type diagnostics.
Remaining selector obligations: 25 candidate pairs/124 reads. Those comprise
field-only recursive Type/FlowNode contracts, unions of array contracts, and a
callable union. These are not covered by the kind projection receipts and remain
pending adapter work. No unchanged whole-tsc interface completion is claimed.

Untagged recursive and array group, October 8: minimal shared hooks are named
recursive optional descriptor references and descriptor publication in
internal/lower/view_contracts.go; untaggedArrayElement in the owned lower adapter,
called by internal/lower/object.go and view_arrays.go; array-union element interning
in view_contracts.go; object/array union dispatch in native/view_fields.go and both
backends' view_unions.go. Recursive optional descriptors retain the original ID,
not a copy of an unfinished descriptor. Runtime matching uses a conservative
128-edge depth limit, rejecting deeper input rather than overflowing. Array member
selection validates one complete member's element selectors, rejecting mixed
A/B arrays; unread own-kind payloads remain deferred and consumed indices retain
joined element contracts. No callable producer signature is inferred from heap kind.
Seven independently written array candidate projections represent 45 reads;
43 projections/355 reads now have receipts, 18 candidates/79 reads remain.
Compiler NodeArray properties and full original interfaces are not certified.

Untagged recursive FlowNode and callable controls, October 8: explicit reference
undefined uses semantic byte 13 in native/emit_objects.go fieldRepresentation.
The read-demand entry in lower/view_lazy.go calls the owned
completeUntaggedRecursiveContracts after all recursive array descriptors exist.
Callable union hooks in lower/object.go and view_callables_read.go prepare each
member signature at a read via prepareUntaggedCallableUnionRead. Fixed signatures
use the owning lane's producer metadata; native/view_callables_signature.go and
javascript/view_callables_signature.go select the expected member via the owned
untaggedCallableUnionExpected. Minimal actual-dispatch hooks are native/view_fields.go
and javascript/view_callables.go. Ordinary single callable contracts are unchanged.
Fifteen narrowed flags/id/node/antecedent FlowNode projections add 35 candidate
reads: 58/390 represented, 3/44 remain. The fixed-signature callable controls do
not complete candidate 227's optional callback and thisArg signatures. Candidates
24 (33 reads) and 72 (10 reads) require a nested TypeChecker callable adapter during
field-only structural matching; the direct callable-field hook does not provide it.
All three remaining boundaries have Node-valid source refusal receipts, separate
from positive completion counts. This is adapter work, not a checker-diagnostic
or handoff wait. Full original compiler interface completion remains unclaimed.

Recursive active-pair matching tracks (contract ID, object/array identity) on the
current path in both owned runtimes. This permits valid cycles while still checking
all other member fields; failed alternatives do not persist a success cache. A
valid cyclic source control and a cyclic wrong-field control accompany both backend
checks. Non-cyclic matching remains bounded to 128 edges and refuses deeper input.
The explicit undefined producer tag also needs a minimal shared runtime/object.c
view-write hook: replacing semantic undefined is accepted only for a physically
reference-holding slot and a reference write. Class initializer and cycle source
controls prove the write succeeds; omission is caught by a semantic mutant.
Lane 5 next group adds known-void signature metadata at the owned read adapter
and runtime shape selection, including native method producer certificates.
Void remains distinct from unknown and discarded-result contracts. Original
FileWatcher.close witnesses pin this fixed zero-argument shape.


Lane 7 integrated-lazy checkpoint, October 7: integration ba59427c includes this
lane's earlier tip and merges by fast-forward. Do not merge individual lanes.
The newer lazy census lists 45 intersection candidates / 769 reads across owners;
lazy/REPORT.md explicitly leaves allocation-exact reachability UNMEASURED.
Do not report these candidate counts as exact or subtract component fixtures.

Concrete shared-hook handoff is lane7/shared-hooks.patch, generated without
editing shared files by lane7/make-integration-overlay.py. The integrator owns
four small hooks: view_lazy.go recognizes structural intersections; view_contracts.go
routes them to internStructuralViewIntersection; view_objects.go admits their data
representation; expression.go selects object storage without the old scalar-only
field restriction. The lane-owned builder uses ViewObject plus Members as a
conjunction and checker-combined Fields for duplicate names. This representation
supports lazy structural field reads, rather than eager payload certification.
It excludes callable, indexed, array, tuple, nominal and primitive constituents;
phantom-only arms contribute no runtime fields. Unsupported descendants retain
lazy obligations. The component all-members matcher is not used by this path.

Source tests enable the handoff via ADAMIC_INTERSECTION_HOOKS=1 with the generated
Go overlay. They cover the first directly structural declared intersection,
GeneratedIdentifier.emitNode (4 candidate reads), with reduced contract fragments,
shared-helper reads, a wrong first-arm scalar, a wrong second-arm nested scalar,
and optional absence. Generic Named & Counted controls remain separate. These
fixtures do not certify the full upstream interface pair or exact reachability.
Production completions remain zero pending integrator hooks. Revised whole-family
working date: October 12, 2026, 23:00 UTC, conditional on hooks landing promptly
and remaining compound families being supported. This is an estimate, not a
promise established by measured throughput.

Lane 7 correctness review: the four-hook patch is INCOMPLETE and must not enable
production admission. Native and JavaScript viewObjectUnion currently ignore
Members when Kind is ViewObject. Thus metadata preserves obligations on later
field reads but does not check all members at an intersection-valued root read.
RootConjunctionProbe is an explicit red probe for this gap. Required fifth/sixth
shared seams are native/view_unions.go and javascript/view_unions.go: recognize
the conjunctive descriptor and call lane-owned all-member snapshot matchers.
Do not certify a shape with only the heap tag or flattened field registration.
Until this dispatch exists, the original unsupported intersection refusal stays
active in the production tree. The overlay is experimental evidence only.


Lane 7 production hook group, October 8, authorized by the shared-hook rule:
shared lower/expression.go, view_lazy.go, view_contracts.go and view_objects.go
now call structuralViewIntersection and internStructuralViewIntersection.
IR views.go adds the named Intersection discriminator; union selection is kept
separate. Native/view_unions.go and javascript/view_unions.go dispatch this flag
to lane-owned viewObjectIntersection methods. These use checker-combined Fields
to check all nonphantom member obligations, including duplicate-name child
intersections, and preserve one snapshot per projected field. The existing
presence/readiness and expected-type diagnostics are reused. Deferred unsupported
descendants retain lazy read obligations; a recursive backedge retains checks on
subsequent projected reads. Native and JavaScript code generation lives in the
lane's owned view_intersections files. These seven named hooks supersede the
incomplete overlay handoff. Production tests no longer skip or need an overlay.

Candidate queue remains the lazy inventory, labeled candidate, across all owners.
No exact counts are inferred. First production group verifies GeneratedIdentifier
emitNode shape fragments, root-only reads, helper reads, optional absence,
object phantom brands and duplicate-field nested intersections. Full upstream
pair certification remains distinct from reduced fragments. Revised whole-family
working date: October 11, 2026, 23:00 UTC. This estimate now has no shared-hook
handoff dependency; compound union, recursive, array and callable intersections
still require additional implementation and validation.

Lane 7 conservative scope hook: view_lazy.go also calls the named
viewIntersectionReadFamily classifier after descriptor construction. Read demand
refuses union selection containing intersections, deep recursive payloads and
supported compound child families that the finite object/scalar matcher cannot
validate yet. Casts stay lazy, and already-unsupported descendants keep their
own read obligations. This prevents the new general intersection classification
from accidentally admitting unwired compound runtime paths. Until those adapters
are implemented, these are explicitly pending candidate families.

Lane 7 scoped candidate accounting: the 45 / 769 shared queue separates into
15 explicit object-intersection candidates / 64 reads, 23 primitive-brand
candidates / 692 reads delegated outside this lane, and 7 private builder/array
alias overlaps / 13 reads whose ownership/runtime classification is unresolved.
This partitions the static candidates, not allocation-exact reachability and
not phantom-brand proof from display text. The object keys are preserved in
lane7/lazy-pair-progress.json. Full upstream pairs certified remain zero;
reduced source shapes are not subtracted. Recursive checker-type traversal also
prevents optional recursive descriptor copies with temporarily empty Fields
from bypassing the demanded-read refusal.


Lane 7 selected-arm hook group, October 8: IR IntersectionTag and the native /
JavaScript view_unions dispatch call lane-owned viewIntersectionUnion methods.
view_lazy calls viewIntersectionReadFamily with the checker type. A finite
object union with intersection arms, or a previously refused enum-tag alias,
can select disjoint literal-tag arms and validate every selected field with one
tag snapshot. Equal field obligations may be coalesced; overlapping unequal
arms, recursive union payloads and untagged selection remain refused at demand.
Existing supported plain union dispatch keeps its diagnostics. No protected
emit.go, lower.go, native.go or oracle_test.go edits are needed.

Original source witness metadata verifies 22 candidate pairs / 77 reads against
pinned TypeScript git objects without copying upstream files. Builder declaration
spans classify five overlap pairs / 11 reads as numeric-brand or array work;
two overlap pairs / 2 reads involve FileInfo object intersections behind callback
unions and remain in lane 7. The revised queue is 17 object candidates / 66 reads,
28 delegated candidates / 703 reads, zero unclassified overlaps. Full upstream
certifications still zero: finite declaration fragments do not cover complete
tsc interfaces. Original-read-witnesses.json preserves exact UTF-16 spans and
source hashes; overlap-classification.json records the declaration evidence.
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


Lane 7 recursive hook group, October 8: IR IntersectionRecursive selects the
lane-owned recursive object walker. ObjectPresent in optional descriptors points
to their completed canonical object; view_contracts.go sets this link when copying
a present object descriptor. Lane-owned IR RecursiveIntersectionObjects emits
only the graph reachable from this read. Native/runtime/adamic.h includes the
new owned view_intersections_recursive.h; its C walker reads each combined field
through shared presence/readiness/physical-type validation, checks literals,
recurses with a full field path and tracks (object, contract) active pairs.
Paths are explicitly allocated and freed; successful native fixtures pass leak
checks. JavaScript uses the same descriptor obligations and active-pair rule.

Two further minimal shared hooks, native/view_nullish.go and
javascript/view_nullish.go, call viewIntersectionNullishRead after physical
nullish admission. A present optional or null-allowing intersection must retain
its payload checks. The new optional-root test caught the former bypass in both
backends. Null and undefined controls remain handled by their owning lane.

Plain recursive object/scalar intersections are now supported. Recursive union,
array, callable, nullable descendant and nominal payloads remain pending or lazy
unsupported obligations at descendant reads. Cyclic allocated source programs
are not claimed as covered. Reduced recursive fixtures do not decrement full
upstream candidate counts: 17 pairs / 66 reads remain. Date remains October 11,
2026, 23:00 UTC. Original leading tsc witness spans are in the prior group.
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

### Lane 2 mutable, recursive and nullable ranked arrays

The fourth ranked group (CaseBlock.clauses, ParsedCommandLine.fileNames,
DiagnosticMessageChain.next, FlowLabel.antecedent, CommaListExpression.elements)
needed no new hook. It merged views-integration for lane 1's nullish member
reads and changed three shared places:

- `arrayLiteralElement` in `internal/lower/graph_flow.go`, called from
  `graphFlows` for array literals, joins the element types of a union slot's
  array members instead of asking a union for type arguments, which panicked.
- Native `adoptGraphObject` sizes graph adoption with `adamic_object_size`, so
  the per-slot contract IDs are copied into graph storage.
- Native checked view writes name object, array and Map types as the
  JavaScript backend does.

An array written into a slot made holding undefined, and pushes of records
outside the flat scalar subset, remain named exit-70 refusals. Fixtures are
`lane2/ranked4-*.a`, held by TestCheckedViewRanked4ArrayContracts; the report is
`lane2/RANKED4_ARRAYS_REPORT.md`.

### Lane 2 fifth ranked arrays

The fifth ranked group (ResolvedType.constructSignatures, ClassDeclaration.modifiers,
ExpressionWithTypeArguments.typeArguments, TupleTypeNode.elements, JSDoc.tags,
FunctionLikeDeclaration.parameters) needed no new hook and changed no shared file.
Whole-array assignment through a mutable field keeps the original alias checked:
an assigned array whose elements break the alias's contract is refused at that
alias's next read in both backends.

FunctionLikeDeclaration.parameters is read through the union after member casts;
a direct cast to the union is refused by cast_proof.go and pinned as a frontier
until lane 4 routes union targets to the shared view entry. That pair is not
credited. Tuples moved to lane 4c; TupleTypeNode.elements stays here because the
census files it under array contracts. Fixtures are `lane2/ranked5-*.a`, held by
TestCheckedViewRanked5ArrayContracts and TestCheckedViewRanked5UnionCastFrontier;
the report is `lane2/RANKED5_ARRAYS_REPORT.md`.

### Lane 2 sixth ranked arrays

The sixth ranked group (ResolvedType.callSignatures, TemplateLiteralType.texts,
TupleType.labeledElementDeclarations, ClassDeclaration.members, HasJSDoc.jsDoc,
JsxAttributes.properties, and the `ClassDeclaration | ClassExpression` members
read) changed one lane 2 runtime line: `adamic_view_array_at` names a present NULL
reference `undefined`, as the JavaScript backend and the earlier array-undefined
pin do. The union cast is a pinned frontier like FunctionLikeDeclaration's, and a
JSDocArray built with its own jsDocCache field is a pinned NotYet. Fixtures are
`lane2/ranked6-*.a`, held by TestCheckedViewRanked6ArrayContracts and
TestCheckedViewRanked6Frontiers; the report is `lane2/RANKED6_ARRAYS_REPORT.md`.

### Lane 2 seventh ranked arrays

The seventh ranked group (ClassDeclaration.heritageClauses, HeritageClause.types,
SetAccessorDeclaration and ConstructorDeclaration parameters, Type.aliasTypeArguments,
TemplateLiteralType.types, SourceFile.bindDiagnostics, Diagnostic.relatedInformation,
ParsedCommandLine.projectReferences) needed no new hook. Pushes of records with object
or undefined-typed fields remain named runtime refusals, and assigning a fresh array to
a field of a Diagnostic read from a viewed element is a pinned NotYet. Fixtures are
`lane2/ranked7-*.a`, held by TestCheckedViewRanked7ArrayContracts and
TestCheckedViewRanked7Frontiers; the report is `lane2/RANKED7_ARRAYS_REPORT.md`.

### Lane 2 eighth ranked arrays

The eighth ranked group (JsxElement.children, Method and Function declaration
parameters, ParameterDeclaration.modifiers, EnumDeclaration.members,
Bundle.sourceFiles, SourceFile.imports, Signature.compositeSignatures,
GenericType.typeParameters) needed no new hook. A recursive union element type is
read through its members, and a flat record push is checked against the original
element contract. Fixtures are `lane2/ranked8-*.a`, held by
TestCheckedViewRanked8ArrayContracts; the report is `lane2/RANKED8_ARRAYS_REPORT.md`.

### Lane 2 ninth ranked arrays

The ninth ranked group (fourteen pairs: HasType member parameters and modifiers,
ImportDeclaration.modifiers, NamedImports.elements, ModuleBlock.statements,
NewExpression.arguments, TemplateExpression.templateSpans,
NodeBuilderContext.typeStack, ResolvedType.indexInfos and ResolvedType.properties)
needed no new hook. Push, pop and searches on a mutable `number[]` go through the
view's element contract. Fixtures are `lane2/ranked9-*.a`, held by
TestCheckedViewRanked9ArrayContracts; the report is `lane2/RANKED9_ARRAYS_REPORT.md`.

### Lane 2 tenth ranked arrays

The tenth ranked group (fifteen pairs: ArrowFunction.typeParameters, accessor
modifiers, NamedExports.elements, SourceFile reference arrays, binding pattern
elements, CallExpression.typeArguments, ClassExpression heritage and members,
JSDocTemplateTag and InterfaceType type parameters,
TypeReference.resolvedTypeArguments, Symbol.declarations through `Symbol |
undefined`) needed no new hook. Assigning an array to an optional array field
through a view is refused before lowering because `slotContract`
(internal/lower/view_writes.go) has no array certificate; it is pinned. Fixtures
are `lane2/ranked10-*.a`, held by TestCheckedViewRanked10ArrayContracts and
TestCheckedViewRanked10Frontiers; the report is `lane2/RANKED10_ARRAYS_REPORT.md`.

### Lane 2 eleventh ranked arrays

The eleventh ranked group (fourteen pairs: ClassDeclaration.typeParameters,
declaration modifiers, FunctionExpression.parameters, ImportAttributes.elements,
IndexInfo.components, InterfaceDeclaration heritage and members,
SourceFile.libReferenceDirectives, NodeBuilderContext.reverseMappedStack) needed no
new hook. Arrays of intersections check the intersection tag at the element read.
CallExpression | NewExpression and HasDecorators are union-target casts for lane 4;
CompilerOptions.lib sits on an index-signature declaration 0.1 refuses. Fixtures are
`lane2/ranked11-*.a`, held by TestCheckedViewRanked11ArrayContracts; the report is
`lane2/RANKED11_ARRAYS_REPORT.md`.

### Lane 2 twelfth ranked arrays

The twelfth ranked group (sixteen pairs: type parameters and modifiers of several
declarations, JSDocFunctionType.parameters, ParsedCommandLine.errors, SourceFile
moduleAugmentations and packageJsonLocations, JsonSourceFile.statements, template
literal type spans, TypeLiteralNode.members, CaseClause.statements,
NodeWithTypeArguments.typeArguments and JSDoc.comment as a string-or-array union)
needed no new hook. Native checks a nullable union read twice (runtime mask, then
emitted member selection). Fixtures are `lane2/ranked12-*.a`, held by
TestCheckedViewRanked12ArrayContracts; the report is
`lane2/RANKED12_ARRAYS_REPORT.md`.

### Lane 2 thirteenth ranked arrays

The thirteenth ranked group (seventeen pairs: type parameters, modifiers and type
arguments of more declarations, MethodSignature.parameters, MappedTypeNode.members,
JSDoc tag comments, ConditionalRoot type parameters, AnonymousType.aliasTypeArguments,
SourceFile commentDirectives and parseDiagnostics) needed no new hook. Four casts to
union aliases are lane 4's and MapLike's dynamic key is an index signature. Fixtures
are `lane2/ranked13-*.a`, held by TestCheckedViewRanked13ArrayContracts; the report is
`lane2/RANKED13_ARRAYS_REPORT.md`.

### Lane 2 fourteenth ranked arrays

The fourteenth ranked group (eighteen credited pairs: ConstructorTypeNode members,
signature and declaration modifiers, JSDocSignature.parameters, JsxFragment.children,
UnionTypeNode.types, DefaultClause.statements, EmitNode helpers and
tokenSourceMapRanges, InterfaceType type parameters and declaredProperties,
TransientSymbol.declarations, CircularBuildOrder.circularDiagnostics) needed no new
hook. view_lazy.go's field-name fallback refuses any viewed `.text` read once an
interface declares tsc's callable-or-string EmitHelper.text; that is pinned as a
frontier for routing. Fixtures are `lane2/ranked14-*.a`, held by
TestCheckedViewRanked14ArrayContracts and TestCheckedViewRanked14Frontiers; the report
is `lane2/RANKED14_ARRAYS_REPORT.md`.



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

### Lane 4 named primitive read hooks under the direct-hook ruling

The user now authorizes minimal named shared hooks per lane, reconciled by the
integrator. Lane 4 adds viewStringUndefined(property) in its native and JavaScript
view_unions_mixed.go files. Shared native viewField calls it to select the existing
optional payload reader while retaining the required-field presence flags.
Shared JavaScript Property dispatch supplies its result as undefinedMember to
adamicViewField; the readiness helper accepts undefined payloads independently
of optional absence and retains null rejection. Tagged object-union dispatch is
restricted to object-valued fields. These are executable source-read hooks,
not flow erasure or a second readiness state.

Shared strictViewContract classifies phantomUndefined as ViewUndefined;
unsupportedViewFamily permits that approved phantom erasure, and viewDataType
recognizes it in read preflight. No other intersection, brand member, nominal,
callable or dictionary contract is enabled by these hooks. The two leading
candidate pairs are pinned by TestCheckedViewCompleteBrand and both backend
undefined-admission overlay mutants. Integration consumes this lane tip through
the user; no individual lane branch is merged.

### Lane 4 staged boxed primitive probe

Lane 4 owns the named adamic_view_union_heap normalizer in its mixed-union
runtime and adamic_object_view_union_snapshot in shared object.c, declared in
view_unions_mixed.h. The latter calls the existing readiness/static-owner resolver,
then classifies only source storage evidence. Optional absence stays separate
from present uninitialized storage. Unknown storage remains unknown. No cast
admission or read dispatch is enabled by these probe-only changes. Dedicated
release and sanitizer probes cover storage formats and inherited owners before
source dispatch is enabled. The integrator reconciles these minimal named hooks.

Lane 4 also owns internal/ir/view_primitives.go and its tests: the named
PrimitiveViewMembers planner preserves literal/undefined alternatives and
rejects missing, unsupported, object and recursive member graphs. It does not
enable source admission or a second flow graph. The primitive array adapter
remains an unapplied review artifact in lane4/primitive-array-adapter-review.patch,
with its staged hooks and validation gate in lane4/ADAPTER-REVIEW.md. Automatic
approval review rejected enabling that cross-layer work on memory-safety and
silent-miscompile risk. Accepted probe work remains separately reviewable.

Census correction: the reduced finite CompilerOptions key fixtures in 0bae12f8
omitted tsc's string index signature and therefore did not complete either
skippedOn pair. The corrected lane 4 candidate ledger has nine mixed primitive
pairs / twenty-seven reads remaining. The __String family remains complete,
fourteen candidate pairs / five hundred eleven reads. Candidate counts are not
certified production reachability. The full key gaps inherit a library Record
index signature and checker-validate a numeric keyof key without constructing
runtime dictionary storage. They remain compile refusals, not completions.

### Lane 4 dictionary scalar and nullish element selection

Lane 6 GROUP4 at ce4eeaa4 hands lane 4 the scalar/nullish selection portion of
CompilerOptions and BuildOptions dynamic element extraction, two candidate pairs /
eleven reads. Lane 6 retains lookup, absence, enumeration and storage; lane 4b
retains object/array alternatives and descendant contracts. The ranked lane 4
ledger now has eleven pending scheduling pairs / thirty-eight candidate reads,
including nine original mixed primitive pairs / twenty-seven reads. Full shared
extraction remains pending; container certification is not child-read coverage.
See stage3/interface-downcasts/lane4/DICTIONARY-HANDOFF.md. No new hook or second
flow analysis is added by this scheduling checkpoint.

Lane 4 integration reconciliation: lower/view_unions_mixed.go owns
viewBrandedStringUndefined, called by lower/object.go to retain the scalar
undefined-payload certificate for approved string/phantom-void contracts.
Null-containing and ordinary nullable contracts retain the lazy owner's reader.
Open-key reduced witnesses use that owner's now-present primitive member selector;
they do not certify the full original CompilerOptions dictionary extraction.

Lane 5 integration nullable-callable hook: native/view_nullish.go and
javascript/view_nullish.go invoke owned viewCallableNullishCertificate helpers
after the owner presence/readiness/kind checks. Non-null callable values retain
the same signature check; permitted null/undefined preserve their identities.

The nullable-callable reconciliation also attaches complete fixed signatures to
boxed Union reads in prepareViewCallableProperty. Otherwise the new same-name
proof exemption would admit nullable function reads without a signature check.
TestCheckedViewNullishCallableSignatureMutant now pins the runtime wrong-result
refusal in both backends rather than the older compile-time refusal.

Untagged final candidate projections, October 8: reuse lane 5 at tip 15b30747
through integration 6a7f1bf3 (merge 976f8bc6). Revised candidate-projection target
is October 9, 2026, 23:00 UTC; full unchanged TypeChecker remains outside this
claim. New fixtures are pair-24, pair-72 and pair-227 .a projections and the owned
run-nested-callable-mutants.py harness. Minimal named shared hooks:
- view_callable_signatures_match/adamicViewCallableSignaturesMatch extracts the
  existing lane-5 representation predicate for nested non-panicking selection.
- viewCallableProducers/viewCallableRecorded reuses lane-5 immutable code lookup;
  the JS record includes its function index. No closure convention is added.
- completeViewCallableShapeContract accepts represented optional parameters;
  generics, rest and overloads still refuse. The scalar read filter stays strict.
- lower.untaggedCallableTargets and certifyUntaggedCallableProducers reuse existing
  closureRecords and exact checker identity to populate ViewContract.Functions;
  ProducerCertified requires that whitelist, including an empty whitelist. This
  is conservative for higher-order parameters and object results.
- checkLazyViewReads runs that certificate hook and bypasses field-name fallback
  only for a fully producer-certified callable read; receiver refusals remain.
- prepareUntaggedStructuralRead is called at object-union and array-element reads
  to demand nested supported scalar callable certificates lazily.
The native and JS owned structural matchers use the shared producer certificates.
The native declaration includes lane-5 header before its generated signatures.
Both callable read emitters retain lane-5 named refusals and enforce logical
producer membership for unions. Compiler-wide reachability remains unmeasured;
fixtures deliberately narrow TypeChecker to one scalar check method and forEach
thisArg to optional number. Intrinsic array forEach and its any thisArg are not
certified by these fixtures. Full original-interface completions remain zero.

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

Lane 7 latest integration reconciliation: viewDataType retains both structural
intersection recognition and viewArrayBase ancestry recognition; plan appends
from both lanes are preserved. Central tip d50c1d37 is merged, never a peer tip.


Lane 7 original-certification group: new files are
internal/oracle/checked_views_intersections_original_test.go and
stage3/interface-downcasts/lane7/original/*.a, prepare.cjs and run-mutants.py.
They reference complete declarations emitted outside Adamic from pristine
microsoft/TypeScript at 050880ce, using the integrated lane4b emission adapter.
All emitted files are hash checked; complete original receiver, component and
combined intersection field sets are asserted. No cohere file is copied.

Named minimal hooks for this group: the owned intersection classifier admits
ViewCallable child kind checks, and both owned field walkers retain callable
presence/kind checks. Shared javascript/readiness.go adamicViewField recognizes
representation 8 as a real function or AdamicClosure. Signature/body proof stays
at the owning callable read/call boundary; a function tag is never that proof.
Shared lower/view_lazy.go consults viewIntersectionReadChecks before the
field-name fallback. A resolved supported conjunctive read keeps its emitted
checks instead of inheriting a different wider carrier's unsupported descriptor.
Unsupported or unresolved reads retain the existing conservative fallback.
The original missing-optional control caught that fallback collision.


Intersection integration presence hook: native object_optional_view separates
required property presence, optional receiver presence and permitted undefined
payloads. The explicit adamic_object_optional_view_undefined entry point and
native viewField preserve the lane 4 branded-string undefined member flag.
Finite and recursive intersection walkers pass field.Optional independently of
child.Undefined; JavaScript uses the existing ninth undefinedMember argument.
RequiredUndefinedPresence source controls refuse absent required fields while
permitting explicitly initialized undefined values. Recursive tables carry both
flags. This fixes an observed all-backend root-read gap without weakening any
callable signature, Map certificate or lazy descendant obligation. Nullable
intersection hooks handle direct conjunctions; the lazy owner's union selector
already validates tagged intersection arms, so that dispatch is not duplicated.


### Lane 2 reference element writes

Lane 2 adds the following named shared hooks for the flat required scalar-record
subset. Callable declarations remain Lane 5's territory.

- `viewArrayLiteralSourceContract` and `viewArrayObjectSourceContract` in
  `internal/lower/view_array_writes.go` are called from array/object literal
  production. They retain original allocation declarations, excluding nominal,
  optional, nested and callable records. Contextual scalar slot declarations are
  retained without trusting contextual readonly views as allocation evidence.
- `markViewArrayScalarSlotWrite` is called from `setProperty` and
  `updateProperty`; `prepareViewArrayReferenceWrites` runs first in readiness.
  Together they preserve scalar RHS contracts and activate original-slot guards
  for certified records, including alias writes discovered before array origins.
- IR metadata: `ArrayLiteral.ElementContract`,
  `ObjectLiteral.ArrayWriteContract`, `SetProperty.ArraySlotWriteContract`.
  `FlatArrayRecordContract` and `ArrayRecordWritePairs` in a new IR file use the
  shared `ScalarWriteContracts` structural and mutable-field invariant rule.
- Native/JavaScript `viewArraySourceCertificate` is dispatched from their existing
  array-view value hooks. Native `viewArrayReferenceWrite` and JavaScript
  `adamicArrayReferenceWrite` are used by checked element replacement and push.
  The existing emitters pass the already-evaluated incoming value to the named
  mutation hook; verification precedes retaining and changing the array.
- New native runtime files `view_array_writes.c/.h` provide certificate setters
  and the compatibility guard. `adamic_array.element_contract` and
  `adamic_object.array_write_contract` are unsigned IDs, separate from physical
  `element_kind`. Constructors (including region objects) initialize zero.
  Dense/checked sparse slice preserves the array ID; plain record copying
  preserves its object ID. Source literal hooks establish certificates; zero
  means unsupported production and refuses writes. No destructor ownership
  change is introduced. JavaScript mirrors the IDs with WeakMaps.
- Generated scalar-record verification checks incoming fields for presence,
  readiness, scalar kind and finite literal constraints at the write. Array
  field reads still do not scan elements. Unsupported original/incoming record
  contracts remain named exit-70 refusals.

Fixtures are `lane2/reference-array-*.a`, held by
`TestCheckedViewArrayReferenceWrites`. This is a flat-record write foundation,
not certification of complete tsc Type/Symbol/Diagnostic shapes, class elements,
optional/nested record writes, array/map/function elements or every mutator.

### Lane 2 readonly array unions

- `viewReadonlyArrayUnionBase` in `internal/lower/view_array_unions.go` is called
  from `viewArrayBaseSeen`. Every alternative must have valid, non-overridden
  ReadonlyArray ancestry. The common base identifies array storage and intrinsic
  names; it does not select an element alternative or inspect array contents.
- `viewArrayUnionElementType` joins every alternative's declared element type
  with the checker and is called from `viewArrayElementType`. Shared `elementType`
  uses this named hook instead of replacing the type by the first array base.
  `internArrayViewContract` and existing read/consumer metadata therefore retain
  the union element descriptor, including its existing tagged object checks.
- `viewMutableArrayUnion` is called from `unsupportedViewFamily`. Mutable array
  unions remain a named lazy field-read refusal; an unread member does not block
  cast admission. This checkpoint does not erase their distinct write contracts.
- `adamic_object_view`'s existing diagnostic allocation accounts for both copies
  of the expected type in its format string. BindingPattern's long array-union
  name exposed the former out-of-bounds panic text read under ASan. Diagnostics
  are unchanged; restoring the old sizing is an executed mutant witness.

No new IR node, array header metadata, callable hook or element scan is added.
The new oracle is TestCheckedViewRankedArrayUnionContracts, with Node controls,
both backends and exact kind/tag failure pins. The separate
TestCheckedViewMutableArrayUnionRefusal pins the unsupported read diagnostic.


Array integration counts hook: interfaceCastCounts also records the measured
ranked and reference-array source fixtures through viewRankedArrayCounts.
The dedicated checked-view count test updates only those measured rows before
the existing predicate direction table. The deliberately refused mutable array
union read has no runtime artifact and therefore no runtime allocation row.
These are runtime allocation counts for fixtures, not production read coverage.

Object plus primitive lane, original LiteralType checkpoint (October 8):
New owned witnesses: lane4b/original/literal-{good,boolean,negative-wrong,text-wrong}.a
and bindable-expression-{unread,frontier}.a. Original LiteralType.value (10
candidate reads) imports complete LiteralType and PseudoBigInt declarations.
Minimal named shared hook: internal/lower/library_object.go objectIntersection
calls objectPrimitiveIntersectionStorage in the lane-owned lower helper. This
recognizes object-reference storage for plain structural unions/intersections
with aggregate fields; it proves no conjunction or member contract. Existing
lazy admission still refuses the unsupported bindable expression at its read.
Cast-only original bindable witness admits in both backends; demanded read pins
the named representation-conversion refusal. This pair remains uncertified.
Existing intersection oracle checks are included in validation. No change to
view_lazy.go or shared contract acceptance is included.

Object plus primitive lane, next original batch (October 8):
New owned .a witnesses in lane4b/original use complete upstream
DiagnosticRelatedInformation, AutoGenerateInfo/GeneratedNamePart, ConditionalType/Type,
DiagnosticWithLocation and DiagnosticWithDetachedLocation. Owned *-probe.a files
record uncertified higher-ranked intersection, metadata-array and dictionary
frontiers. No production or shared-hook changes are added in this batch.

Dictionary handoff from ce4eeaa4 GROUP4.md is accepted: lane 4b owns object/array
member selection inside CompilerOptionsValue and the TsConfigSourceFile
alternative, with descendants retained as lazy checked reads. Lane 6 owns lookup,
absence and enumeration; lane 4 owns scalar/nullish selection. Original candidate
rows CompilerOptions dynamic-key (10, rank 6) and BuildOptions dynamic-key (1,
rank 36) are already in our 42/181 inventory; they are not added or certified
again. Owned resume/dictionary-member-handoff.json records this split and counts.
Full dynamic lookup validation awaits lane 6's hooks in views-integration,
currently ba59427c locally; no dictionary branch was merged.


### Dictionary source hooks, October 8 rule change

The user now authorizes minimal named shared hooks. This supersedes the waiting
on owner handoffs above. Dictionary source read group owns these shared hunks:

- `ir/views.go`: ViewDictionary, a read descriptor, never a writable-slot proof.
- `ir/ir.go`: Property.DictionaryKey, evaluated after the receiver exactly once.
- `ir/view_dictionaries.go`: DictionaryReadKinds, common fail-closed adapter boundary.
- `lower/view_lazy.go`: stringDictionary dispatch to viewDictionaryContractHook.
- `lower/refusals.go`: defer index-signature obligations to producer/read lowering.
- `lower/object.go`: named and indexed dictionaryRead hooks before tuple dispatch.
- `lower/readiness.go`: preserve DictionaryKey guards through final rewriting.
- `lower/shape_conformance.go`: exclude dynamic dictionary reads from scalar erasure.
- `lower/view_writes.go`: dictionary read descriptors cannot certify writes.
- `flow/infer.go`: visit dictionary key effects after receiver effects.
- `native/emit_expressions.go` and `javascript/javascript.go`: dictionaryRead dispatch.
- JavaScript assembly includes DictionaryRuntime after shared readiness metadata.
- `runtime/record.c` and `adamic.h`: adamic_record_is checks the actual producer
  shape; adamic_record_check_missing_member reuses the existing own-key policy.

The new native source adapter probes actual fixed-shape slot type/readiness before
normalizing a selected value. It preserves identity and does not copy or reinterpret
an ordinary object as a table. Actual record wrappers route through existing record.c
and independently check the table's reference-value representation. No second record
representation is introduced. Source dictionary literals currently retain their
existing fixed shapes; dynamic record producers/writes are the next owned group.

Seven compiler-source controls and six source semantic mutants supplement the
original component tests. Rich element adapters remain demanded-read refusals.
The accepted measurement is 29 pairs / 228 reads, labeled static candidates.
No production candidate pair is credited from representative fixtures alone.

### Dictionary shared storage and propagation hooks, October 8

This lane merges records-lowering 456c981b at the user's explicit request. It
supersedes the earlier handoff for record production; other lane changes still
arrive through views-integration. No parallel table representation is introduced.

Owned minimal shared hunks in this group:

- `lower/generic.go`: inferDictionaryIndexTypes carries instantiated string-index
  element arguments into generic helper bodies.
- `lower/view_objects.go`: structural dictionary views use the existing view path.
- `ir/ir.go`: Record uses storage tag 14, preserving reserved null/undefined tags
  12 and 13 already used by checked field writes.
- `ir/records.go`: read demand metadata and an optional DictionaryRead descriptor.
- `lower/records.go`: prepare checked record reads, reuse finite-partial storage,
  permit boxed primitive unions, and defer structural casts to checked admission.
- `lower/view_dictionaries.go`: internDictionaryViewContract describes Record
  producers; dictionaryProducerRefusal checks unsupported actual producers;
  dictionaryWriteRefusal separates ordinary records from uncertified view writes;
  regexDictionaryReceiver keeps named RegExp groups on their owning adapter.
- `lower/view_lazy.go`: finite-partial dictionary descriptors, Record aggregate
  reachability, and demanded refusals for unsupported reads/enumeration.
- `lower/view_contracts.go` and `lower/interface_cast.go`: Record-valued optional
  fields retain their dictionary contracts and lazy admission.
- `native/emit_records.go` and `javascript/records.go`: select the checked read
  descriptor conservatively when the program has view origins; plain programs
  retain the original record operations. Native producers call record_new_typed.
- `runtime/record.c` and `adamic.h`: record_new_typed stores an independent
  element representation certificate in the existing wrapper's numeric slot 1;
  record_storage_check rejects operations through mismatched storage.
- `runtime/view_dictionaries.c`: scalar/maybe-number reads decode that source
  certificate; reference tables still classify existing heap values and boxes.
- `runtime/object.c`, `native/view_fields.go`, `javascript/readiness.go`: Record
  container reads check the ordinary object heap kind before read dispatch.
- `fresh/fresh.go`: dynamic dictionary reads use the existing element edge.
- `lower/cycles.go`: unsafe writes into counted record tables retain cycle
  refusals until the graph ownership lane supplies record cleanup.
- Merge compatibility: callableDataDeclaration in `lower/view_callables.go`
  prevents a destructuring BindingPattern from reaching Node.Text; the existing
  flow oracle for require_node_perf_hooks exercises the repaired path.
- `native/record_test.go`: preserve semantic mutants after the exported missing
  member helper rename; `oracle/records_test.go`: use the current leak checker.

Group 3's MapLike<string[]> candidate has source reads over both fixed objects
and genuine record producers, with transitive array checks. It credits 1 pair / 6
candidate reads; 28 pairs / 222 reads remain. Enumeration and uncertified writes
remain demanded refusals. Rich CompilerOptions element unions are unfinished.

### Dictionary group 4: lazy rich container read

New lane files: `internal/oracle/checked_view_dictionary_container_test.go` and
`dictionaries/source/command-options-{good,wrong,missing}.a`. No shared code hook
was needed. The required ParsedCommandLine.options container read checks object
kind and presence without demanding unread rich union elements. This credits
the 111-read candidate container pair only, not CompilerOptions dynamic reads.

Union selection overlap proposed to lane 4: scalar and nullish alternatives at
dictionary element extraction. Lane 4b: object/array alternatives in
CompilerOptionsValue and TsConfigSourceFile unions. Dictionary storage, lookup,
absence and enumeration adapters remain this lane's responsibility.

### Dictionary group 5: remaining direct container shapes

Lane files: extend `internal/oracle/checked_view_dictionary_container_test.go`
and `checked_view_dictionary_propagation_test.go`; new `.a` witnesses under
`dictionaries/source/` named watch-options, wildcard-directories, version-paths,
incremental-{multi-options,bundle-options,options}, {reusable,builder}-state-options,
and compiler-paths, each with valid, wrong and absent cases. No shared hook
changed. Parent container checks retain unsupported unread union descriptors;
CompilerOptions.paths uses the mixed named/index-signature read dispatch and
checks extracted arrays transitively. Container bypass is a semantic C/JS mutant.

User accepted lane 4's primitive/nullish and lane 4b's object/array/TsConfigSourceFile
element selection handoffs. Storage, lookup, absence and enumeration remain
with dictionaries. Candidate string and Path dynamic rows are string indexing
according to their census sites; counts stay pending until disposition evidence.

### Dictionary group 6: optional receiver presence

New lane files: `internal/oracle/checked_view_dictionary_optional_test.go` and
`dictionaries/source/optional-{watch,options,wildcard}-{good,wrong,missing,receiver-missing}.a`.
Minimal shared hooks: `native/runtime/object.c:adamic_object_optional_view`
distinguishes an absent receiver from a missing required field on a present
receiver; `javascript/readiness.go:adamicViewField` permits undefined because
of optional chaining only when the receiver is absent. The compiling C/JS
optional-absence mutant deliberately restores that conflation.

### Dictionary group 7: erased-source lookup admission

New lane files: `internal/oracle/checked_view_dictionary_lookup_test.go` and
`dictionaries/source/{map-string,map-generic,package-paths}-*.a`. Named hook
`lower/view_dictionaries.go:dictionaryCastNeedsView` is called minimally from
`lower/cast_proof.go` before apparent upcast admission. An erased structural
source declaring no keys cannot certify dictionary storage by assignability;
its reads install the ordinary lazy dictionary view. Nominal and writable-slot
restrictions remain required. Existing shared record storage is reused.

Group 7 also adds the minimal dictionary-property pointer traversal in
`lower/readiness.go:readinessStatement`. Its record read descriptor must receive
the same parameter readiness facts as its enclosing operation; otherwise a
generic dictionary helper emits a nonexistent parameter readiness flag.

### Dictionary group 8: finite keys and optional named lookup

New lane file: `lower/view_dictionary_lookup.go`, containing named hooks
finiteDictionaryKeys and optionalDictionaryField. Minimal dispatch hooks in
`lower/object.go` select these adapters; `lower/view_dictionaries.go` exposes
dictionaryReadContract to preserve the selected declaration through an optional
receiver. Finite string keys select ordinary own fields through the existing
dictionary probe. Optional named lookup uses an ordinary IR conditional helper,
evaluates the receiver once and distinguishes receiver absence from field absence.
New source witnesses are optional-paths and finite-cache shapes; the propagation
oracle gains the optional-paths controls. No record representation is added.

Group 8 also owns `internal/oracle/checked_view_dictionary_finite_test.go`, finite-cache
and enum-map source controls, optional required-field controls, and compiling C/JS
finite-key and optional short-circuit mutants. Numeric enums use the existing adapter.

Dictionary integration cleanup hook: native/view_callables_discard.go releases
known Record producer results along with the existing reference result kinds.
lane5/stored-marker/record.a exercises a dynamically allocated string in that
record; TestCheckedViewStoredMarkerCalls pins Node agreement and native cleanup.
The producer result certificate determines ownership; no valued callable shape
is admitted by this discarded-result cleanup.
Untagged integration refresh 80a6921c: four conflicts resolved explicitly. Keep
both lanes' plan sections. Run intersection and intersection-tag dispatch before
untagged object/array union dispatch in both backends. Lazy field-name fallback
is skipped when either intersection read checks or producer-certified callable
read checks prove a runtime obligation, while explicit receiver/member refusals
still apply. Preserve incoming required-undefined presence metadata and owned
callable ProducerCertified metadata in the merged IR descriptor.

Post-refresh untagged repair: viewOptionalArrayContract in view_array_adapter.go
retains the canonical array ID in ObjectPresent. completeUntaggedRecursiveContracts
fills an optional array's missing element only from that completed canonical
array descriptor. The incoming intersection hook changed recursive construction
order and exposed the formerly copied unfinished element in candidate 92; no
unsupported element is certified by this repair. Callable read emitters retain
physical mismatch diagnostics and require logical membership whenever the
physical signature otherwise matches. Logical mismatches remain named refusals.

Untagged integration array-union reconciliation: untaggedPlainArrayUnion keeps
separate whole-member contracts for plain readonly array alternatives in
strictViewContract. NodeArray alternatives with own fields keep lane 2's array
adapter and lazy selected-element contracts. The shared elementType hook joins
all alternatives and retains its ordinary array fallback. The mixed-array
refusal, NodeArray lazy control and unsupported mutable union refusal all pass;
omitting this dispatch distinction is an executed semantic mutant.

Untagged undefined producer integration: native dictionary source slot decoding
recognizes semantic tag 13 while retaining the allowed-kind mask for required
fields. fieldInitialRepresentation in native/emit_objects.go preserves physical
Union 10 on reserved uninitialized slots; readiness still refuses reads, and
initialized undefined uses semantic 13. Both fixes have executed omission
mutants and prior-lane Node controls.

### Lane 4 original branded-read certification on the new branch

Lane 4 resumes from integration 4e67894a on codex/views-mixed-unions-2.
Its original/prepare.cjs verifies pinned upstream and the shared complete
declaration output, plus the original thirty-pair / 543-read candidate inventory.
checked_views_brands_original_test.go imports those complete declarations,
requires every original receiver field in the view contract, compares Node with
release native, sanitized native and JavaScript, and runs a member-check bypass
for each certified pair. Unread unsupported fields remain lazy obligations.
This checkpoint adds no shared compiler hook and no separate flow analysis.

Lane 4 original certification now also owns
internal/oracle/checked_views_brand_arrays_original_test.go and
internal/oracle/checked_views_brands_original_gaps_test.go. The array oracle
verifies the complete original primitive-member descriptor and dynamic index
read, including a release-mode check-removal mutant. The gap suite retains
original full declarations and named compile refusals for the three remaining
shared blockers. Private namespace declarations are generated from original
checker.ts AST, and private ActiveLabel from original binder.ts AST; the inferred
anonymous renamed-binding type is emitted by the stock checker. Every original
field and source hash is retained. No source admission hook or flow solver is
added by these tests. Field-name fallback currently blocks two supported name
reads with an unrelated never descriptor; recursive intersection admission
blocks the third. Scoped admission must preserve checks through wider helpers.

### Lane 4 direct tagged union target admission

Named hook viewUnionTargetProof in lower/view_union_targets.go is called by
cast_proof.go for a broad object source and union target. Each target arm must
preserve source slots and carry a distinct, readonly, non-accessor literal tag.
Existing cast.go emits its multi-value tag guard and enters the shared lazy
view. No payload is trusted by the tag test. Lane 4 owns the corresponding
union-target oracle and original FunctionLikeDeclaration/ClassDeclaration union
fixtures. The integrator reconciles this small cast-proof dispatch hook.
The existing certifiedCheckedCast hook also retains the union target contract
on cast.go's multi-tag IR node. The tag proof remains distinct from payload proof.

### Lane 4 fallback collision cause and scoped demand

Original ArrowFunction declares name: never. Interning the complete Node graph
therefore creates that descriptor even for ActiveLabel and renamed bindings.
The global field-name fallback previously applied that obligation to every
viewed allocation named name; similarly a Map members descriptor blocked class
array members. New lower/view_demand_scopes.go and tests attach existing
CheckedCast.ViewContract targets to the actual origins, then query lane 3's
allocation graph and projection index to carry target contracts over actual
field/element producers and stores. No second allocation solver is introduced.
The named scopedViewFieldFamilies hook in view_lazy.go scopes only the fallback;
direct unsupported reads, wider helpers receiving a viewed unsupported member,
unknown callbacks, untracked origins, opaque projections and mutations retain
refusals. Cast.go uses existing certifiedCheckedCast for union target metadata.
Primitive nullish payloads do not invent aggregate allocations. Original name
oracles assert ArrowFunction.name: never is still in the complete descriptor
graph and run field-check mutants. The wider-helper forged-string mutant must
be caught by the never obligation despite passing the helper's primitive check.
The intersected Identifier pair remains lane 7's work.

### Lane 4 first nonbrand primitive certification

Owned original primitive fixtures and declaration witness preparation live in
stage3/interface-downcasts/lane4/primitive-original; oracle tests live in
internal/oracle/checked_views_primitives_original_test.go. Four pairs / 69
candidate reads are certified using existing integrated nullable selectors.
EmitNode direct destructuring is explicitly not credited by a property fixture.
No production hook is added in this checkpoint.

### Lane 4 primitive destructuring hook

New lower/view_primitive_reads.go names viewPrimitiveUnionRead and
preparePrimitiveDestructuredRead. Minimal hooks in interface_cast.go and
collections.go admit only complete declared primitive union descriptors and
attach existing runtime selector metadata to object binding field reads.
Unsupported, missing and recursive descriptors retain refusal; global unknown
flow guard remains unchanged. No runtime ABI or emitter change.

The original binding fixture also exposed a panic in view_callables.go: its
producer scan called Text on an unrelated VariableDeclaration binding pattern.
Named viewCallableProducerDeclaration limits producer-name extraction to the
four existing producer declaration kinds; callable proofs remain unchanged.

Ordinary binding control exposed an unsafe erasure of the primitive-to-box
conversion. Named primitiveBindingConversion in readiness.go retains the
existing selector only for union binding reads with complete primitive
descriptors. This is representation conversion, necessary with or without a
view; ordinary helpers are held to Node in native, JS and sanitizer runs.

### Lane 4 pure primitive property selectors

Named preparePrimitivePropertyRead reuses the existing nullable selector with
no admitted nullish alternative for pure primitive unions. Minimal object.go
and interface_cast.go hooks require complete primitive descriptors. Readiness
retains primitive-to-box conversion for property reads as well as bindings,
including ordinary values. No runtime layout or emitter change. Original
StringLiteralType | NumberLiteralType.value is the next six-read candidate.

The lower NotYet table drops only its now-supported primitive field-read row;
union field writes, class fields and array/closure storage refusals remain.
An ordinary pure-primitive field fixture supplies Node/backend coverage.
The original union helper is tested with viewed values of each full member;
direct untagged union admission remains lane 4c and is not credited here.

### Lane 4 nullable finite-member diagnostic alignment

Minimal hook in javascript/view_nullish.go: nullishMemberSelection names
"matches no member of" the declared union, matching the native selector when a
runtime kind exists but a finite literal member rejects it. Resolved.originalPath
false versus its true-only member exposed the mismatch. No check is removed.
Complete original Resolved is printed from its private source interface and
hash-pinned separately; all five fields, including PackageId, are retained.

### Lane 4 private primitive receiver certification

Owned primitive-original/prepare.cjs now prints complete original FlowGraphNode,
FlowGraphEdge and watch presence interfaces, retaining dependencies and hashing
the generated declarations. No new production hook in this batch. Builder
signature receiver intersection remains lane 7, with a named refusal pin.
Owned measure-tuple-candidates.cjs verifies two scalar tuple census corrections
for lane 1: two pairs / four candidate reads, no certification credit.
## Lane 5 scalar and boxed callable adapters (October 8)

Lane 5 owns new native/view_callables_boxing.go and lower/view_callables_boxing.go,
plus oracle/checked_views_callable_boxing_test.go and lane5/boxing fixtures.
Minimal shared hook: native/emit_functions.go routes closure invocation through
the named boxed adapter. The existing callable slot guard already admits Union.
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

## Lane 5 aggregate callable demand (October 8)

New lower/view_callables_aggregate.go owns deferred payload schema collection and
registration. Minimal shared hooks: lower.go stores checker payload references;
view_callables_read.go records payload type IDs while comparing representations;
functions.go records producer payload types; view_lazy.go completes aggregate
schemas before demand and follows aggregate call results. interface_cast.go's
schema visitor follows signature payloads, parameters and results. collections.go
fills callable and type metadata on destructured reads. ir.ViewContract adds
Payload/PayloadTypeID metadata. Payload schemas are descriptors and are built
only when a checked callable can demand them, never as signature proof at cast.

The named concrete-read predicate in view_lazy.go keeps a checked scalar or
completed callable read from being poisoned by an unrelated wider producer member
with the same name. Direct unsupported descriptors and unsupported receivers
still refuse. Unknown nominal/dictionary producer parameters have no certificate.

Class method certificates keep boxed result/parameter variance unsupported until
their thunk invocation uses an adapter. Native view_callables_methods.go and
JavaScript view_callables_signature.go preserve the same read refusal. Aggregate
return demand reads evaluation operands via ir.ClosureOperands, and result targets
remain resolved by the shared graph and ir.Program.ClosureTargets.

Lane 5 optional callable hook: view_callables_contract.go and
view_callables_read.go admit represented optional parameters while preserving
exact recorded parameter count. Missing numeric optional arguments use the
existing closure argument count and MaybeNumber representation. A producer
requiring the omitted parameter still fails its representation/member check.

Lane 5 long-signature refusal hook: native/runtime/object.c allocates space for
both copies of the declared type in adamic_object_view's failure message. Rank 53's
wrong-value fixture pins its complete long signature on sanitized and release
native, preventing truncated diagnostics and reads beyond the message buffer.

Lane 5 boxed argument field-production hook: lower/object.go calls the named
viewCallableBoxedRecordField in view_callables_boxing.go. A supported contextual
scalar/object union field stores the initializer in the declared boxed form.
This fixes rank 50's scalar argument field being retained as a pointer before
callable dispatch. Unknown, nominal and dictionary members remain unsupported.

## Lane 5 callable array payloads (October 8)

Owned hooks in lower/view_callables_boxing.go and view_callables_aggregate.go
recognize represented array parameters/results and complete their deferred
element schemas before lazy demand. Array extraction continues through lane 2's
named metadata hooks; no callback ABI guard or owner-mutation refusal is removed.
Schema creation certifies neither array contents nor descendant object fields.

### Lane 4 integration reconciliation after private primitive batch

Merge d718a9ff retains both dictionary and primitive admission/readiness hooks,
both untagged-object and primitive property preparation, and scoped fallback
with integration callable exemptions. Global unknown fallback remains intact.
Integration callable implementation-name validation supersedes the narrower
producer-name helper, retaining computed-name refusal and binding-pattern safety.

### Lane 4 read-only primitive array indexes

Named primitiveArrayIndexRead in object.go recognizes only complete primitive
member descriptors at an indexed read. elementType, producers, writes and other
consumers stay unchanged. Owned view_array_primitives.go files in lower/native/
javascript plus runtime/view_array_primitives.{h,c} retain the receiver before
index evaluation, validate physical producer metadata, snapshot one slot, select
its declared member, and return an owned scalar box. No allocation layout or ABI
rewrite. Native evaluate and JS array-read dispatch are minimal named hooks.

PrimitiveArrayReads IR metadata enables existing producer storage certificates
for ordinary primitive array reads too. emit_slots.go's arrayIndexSlot uses the
same selector and owned box for typeof, so typeof cannot bypass selection.

### Lane 4 dictionary primitive and nullish selection

Named DictionaryReadKinds ViewNull admission and native kind mask now retain
null separately from undefined. The dictionary slot normalizers classify the
existing adamic_null heap sentinel, and adamic_view_dictionary_box preserves
that sentinel for both scalar null slots and boxed records. No lookup, absence,
enumeration, allocation layout or ownership ABI is changed. Lane 6 owns those
operations; lane 4 certifies only the scalar/nullish component of rich original
CompilerOptions, OptionsBase and BuildOptions element unions.
