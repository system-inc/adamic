Built: explicitly flagged tuple-union forEach ownership transfer; primary certification remains 5 pairs / 8 reads pending the final pair queue.
Commits: transfer patch recorded in git history, independently of optional/rest admission.
Checks: four ownership fixtures PASS 1.849s; lane 2/lane 4 array and union regressions PASS 25.236s; counts refreshed PASS 29.414s.
Mutants: skip-release leaks 24 bytes in one allocation; double-release triggers AddressSanitizer heap-use-after-free, both expected test failures.
Remaining: optional/rest contracts, then 5 primary pairs / 7 reads and two overlapping lane 4b frontier checks.

The user explicitly approved callback transfer with required ownership controls.
ArrayViewRead carries TupleUnion through finalized contract binding. Only flagged
forEach reads transfer the normalizer's owned count into the existing callback
release path. Indexed reads keep statement cleanup ownership. Ordinary lane 2
and lane 4 consumers keep their existing retain/release. Other tuple-union array
consumers remain refused. A C marker on the flagged normal release supports
precise mutations without touching any other release.

Four .a fixtures are held to source Node, JavaScript, native release, sanitized
native and shared leak checks: tuple with runtime-built string, scalar box,
throw caught outside the loop, and callback storing a tuple in a captured
variable. The throw callback declares void; stage 0 independently refuses a
never-returning forEach callback. Array push into a stored union array retains
its existing source-schema refusal, so the storage witness uses a captured
cell, whose own count survives callback cleanup.

Commands with ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations:
go test ./internal/oracle -run '^TestCheckedViewTupleForEachTransfer$'
-count=1 -v -timeout 3m (/tmp/views-tuples-transfer.log): PASS 1.849s.
ADAMIC_TUPLE_TRANSFER_MUTANT=skip on
'^TestCheckedViewTupleForEachTransferMutants$': FAIL 0.303s. Ordinary output
still matches number/exit 0; leakChecked then observes a 24-byte box leak
(/tmp/views-tuples-transfer-skip-mutant.log).
ADAMIC_TUPLE_TRANSFER_MUTANT=double on the same test: FAIL 0.273s,
AddressSanitizer heap-use-after-free in adamic_release
(/tmp/views-tuples-transfer-double-mutant.log). Neither mutant dies at compile.
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^(TestCheckedViewArrays|TestCheckedViewNativeArrays|TestCheckedViewRankedArrayContracts|TestCheckedViewRankedArrayUnionContracts|TestCheckedViewRankedParserArrayContracts|TestCheckedViewObjectPrimitiveSource|TestCheckedViewMixedSelection|TestCheckedViewObjectUnions|TestCheckedViewPrimitiveArrayPairGap|TestCheckedViewMutableArrayUnionRefusal)$'
-count=1 -v -timeout 6m (/tmp/views-tuples-transfer-regressions.log):
PASS 25.236s, 221 uncached Node observations, unchanged existing diagnostics.
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$'
-count=1 -timeout 3m -args -update-counts
(/tmp/views-tuples-transfer-counts.log): PASS 29.414s, 38 fixture rows.
Scoped IR/lower/native/JavaScript probes PASS 0.016s/0.344s/0.384s/0.199s
(/tmp/views-tuples-transfer-scoped.log). gofmt applied.

---

Built: preserve actual optional tuple producer arity without changing object ABI; certified primary total remains 5 pairs / 8 reads.
Commits: new pairs f1f6355a and 3c143841114922f672fad316938fa7ac383b9bd5; producer checkpoint recorded in git history.
Checks: six producer oracles PASS 3.315s; certified tuple regressions PASS 35.489s; unchanged 34 tuple count rows PASS 35.157s.
Mutants: typed padding produces string instead of undefined, exit 0, and fails its native output oracle; earlier pair mutants recorded below.
Remaining: 5 primary pairs / 7 reads, optional/rest forms, two overlapping lane 4b checks; shared callback transfer and range contracts await approval.

The ABI-changing length proposal was rejected and not applied. The safer
producer instead omits missing optional fields: existing shape count and native
tuple identity preserve actual source arity. JavaScript emits the same actual
array length. Existing optional field lookup handles omission in direct reads,
optional chains and destructuring. No object layout, tuple copy, or second
certificate constructor is introduced. A fixed one-position tuple view proves
the omitted producer has arity one. The shared optional/rest checked-view
constructor remains unchanged and still refuses these descriptors.

Six inline .a probes run source Node, JavaScript, native release, ASan/UBSan
and shared leak checks: omitted, explicit undefined (with explicit undefined in
its exact-optional declared type), present, optional chain, destructuring and
actual arity. Commands, after source /workspace/adamic-tools/env.sh:
go test ./internal/oracle -run '^TestCheckedViewTupleAbsentProducerProbe$'
-count=1 -v (/tmp/views-tuples-optional-producer-final.log): PASS 3.315s.
ADAMIC_TUPLE_OPTIONAL_PRODUCER_MUTANT=padding on .../omitted:
FAIL 0.626s (/tmp/views-tuples-optional-producer-padding-mutant.log).
The mutant adds a correctly typed initialized string position to actual IR;
valid native execution exits 0 with 7/string rather than 7/undefined. An earlier
presence mutant removed the missing-position policy and was killed by UBSan;
that unsafe mutation is not counted as the semantic proof and was replaced.
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations
go test ./internal/oracle -run
'^(TestCheckedViewTupleAbsentProducerProbe|TestCheckedViewTupleOriginalRootIndex|TestCheckedViewTupleOriginalSignaturePositions|TestCheckedViewTupleOriginalOutSignature|TestCheckedViewTupleOriginalTrackedSymbols|TestCheckedViewTupleOriginalReferencedMap)$'
-count=1 -timeout 3m (/tmp/views-tuples-optional-producer-regression.log):
PASS 35.489s. This run precedes adding the optional-chain, destructuring and
arity controls; the final producer run validates those subsequent additions.
The two remaining production guard additions are absent-position flags only;
fixed tuples are unchanged.
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$'
-count=1 -timeout 3m with original declarations
(/tmp/views-tuples-optional-producer-counts.log): PASS 35.157s. Inline probes
add no repository fixture rows. New pair rows were refreshed at each pair push.

Automatic approval review rejected (1) forEach callback ownership transfer,
(2) an ABI-changing tuple-length proposal, and (3) the ABI-preserving shared
optional-range validators and Map-schema semantics. The stated concerns were
ownership, ABI/memory errors and silent miscompilation beyond prior index
approval. Rejected production changes were not applied. The accepted producer
alternative avoids the ABI change and is fully verified above. Two concrete,
gofmt-formatted, unapplied patches now live under review/: forEach ownership
transfer and optional arity contracts. Both pass git apply --check. The rest
extension is described using the same constructor and rest-child edge; it is
not implemented. Pending approval questions identify these exact boundaries.

certifications.json records the queue: 97931 forEach (1 read); 95604 position
1 (2 reads) and position 0 (1); 68230 position 1 (2) and position 0 (1).
The two lane 4b frontier checks remain queued after these; their primary IDs
97898 and 97934 overlap the census and are not counted twice. Exact reaching
coverage is not measured. The unrelated integrator failures below remain
separate from these approval blockers.

Current-session setup timing, distinct from the inherited prior-worker timing:
Node ready 0.033s; Go ready 0.043s; submodules ready 0.088s; markdown ready
0.090s (skip step 0.007s); clang ready 0.258s; Go build ready 795.430s;
test binaries deferred 795.533s; cache ready 795.535s; done 795.618s.
The cold build included lazy submodule fetching and a redundant Go build.
GOPROXY=https://proxy.golang.org|direct; nproc=5, cgroup quota=4.
No complete package or full gate is claimed.

---

Built: certified original emit-signature tuple positions, pair 97898 / 4 reads; total 5 pairs / 8 reads.
Commits: Root dispatch f1f6355a; signature pair commit recorded in git history.
Checks: positional oracle PASS 3.388s; existing outSignature and object-primitive regressions PASS 20.460s; scoped counts PASS 29.443s.
Mutants: dropping ID or signature position guards returns false/string or 7/boolean with exit 0 instead of the named refusal.
Remaining: 5 primary pairs / 7 reads, optional/rest forms, and two overlapping lane 4b frontier checks.

The original IncrementalBuildInfoEmitSignature alias is unchanged. Five fixtures
certify ID and signature positions, including string, empty tuple and singleton
tuple alternatives. The existing constructor supplies all per-position
certificates. Multiple fixed tuple alternatives now use the already-tested
mixed-union selector instead of demanding every tuple arity simultaneously.
Single-tuple alternatives retain their existing diagnostics. Selected tuple
positions admit boxed union storage only for the same tuple/scalar plan.
Node, JavaScript, native release and sanitized native pass; finishing native
fixtures pass leak checks. This covers original tuple-producing objects;
array-to-tuple cast admission remains the later lane 4b frontier obligation.

Commands with ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations:
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalSignaturePositions$'
-count=1 -v (/tmp/views-tuples-signature-positions.log): PASS 3.388s.
ADAMIC_TUPLE_NUMERIC_MUTANT=0 on .../signature-position-id-wrong:
FAIL 0.321s (/tmp/views-tuples-signature-mutant-id.log).
ADAMIC_TUPLE_NUMERIC_MUTANT=1 on .../signature-position-value-wrong:
FAIL 0.356s (/tmp/views-tuples-signature-mutant-value.log).
Both are caught first by JavaScript semantic output/exit checks.
go test ./internal/oracle -run
'^(TestCheckedViewTupleOriginalOutSignature|TestCheckedViewObjectPrimitiveSource)$'
-count=1 -timeout 3m (/tmp/views-tuples-signature-regression.log): PASS 20.460s.
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$'
-count=1 -timeout 3m -args -update-counts
(/tmp/views-tuples-signature-counts.log): PASS 29.443s; 34 fixture rows.

---

Built: certified original Root array index pair 97913; total 4 pairs / 4 reads, with explicitly flagged dispatch.
Commits: prior 6fefeeeea61c9e433604855645a86d43b4d2b7a5; Root pair commit recorded in git history.
Checks: Root index PASS 3.754s; lane 4/lane 2 array regressions PASS 27.314s; gap controls PASS 0.977s; counts refreshed PASS 32.451s.
Mutants: dropped required presence guard returns undefined/exit 0; scalar-as-tuple refuses a valid number/exit 70; both caught by JS and both native modes.
Remaining: 6 primary pairs / 11 reads, optional/rest forms, and two overlapping lane 4b frontier checks.

The user approved dispatch with changes. ArrayIndex.TupleUnion is set by
lowering's tupleScalarUnionType and confirmed against TupleViewMembers after
contract binding. Native and JavaScript dispatch use that flag exclusively;
Element == Union alone never selects the tuple adapter. The normalizer keeps
source storage and retains/boxes the selected slot. The emitted index owns
its snapshot through existing statement cleanup. Callback and loop consumers
still refuse pending their distinct ownership transfer.

Four original IncrementalBuildInfoRoot fixtures cover a tuple, a scalar,
bounds, and a missing required index. Source Node prints object, number,
undefined, undefined. The frontend's ordinary index is optional; the required
fixture explicitly sets the existing IR Required consumer promise on its one
flagged read, then pins the named exit-70 missing refusal. This is a required
backend contract control, not a claim that source typeof requires presence.
Both native builds and JavaScript match; finishing native programs pass the
shared leak checker. The earlier normalizer probe additionally releases the
source array before consuming the retained result.

Commands (source /workspace/adamic-tools/env.sh):
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^TestCheckedViewTupleOriginalRootIndex$' -count=1 -v
(/tmp/views-tuples-index-final.log): PASS 3.754s.
ADAMIC_TUPLE_INDEX_MUTANT=presence with .../root-index-required-missing:
FAIL 0.624s (/tmp/views-tuples-index-mutant-presence.log).
ADAMIC_TUPLE_INDEX_MUTANT=scalar with .../root-index-scalar:
FAIL 0.627s (/tmp/views-tuples-index-mutant-scalar.log).
Both mutants reach executable valid code and fail semantic exit assertions.
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^(TestCheckedViewArrays|TestCheckedViewNativeArrays|TestCheckedViewRankedArrayContracts|TestCheckedViewRankedArrayUnionContracts|TestCheckedViewRankedParserArrayContracts|TestCheckedViewObjectPrimitiveSource|TestCheckedViewMixedSelection)$'
-count=1 -v -timeout 6m (/tmp/views-tuples-array-regressions.log): PASS 27.314s,
209 uncached Node observations, including lane 4b comment arrays.
go test ./internal/oracle -run
'^(TestCheckedViewPrimitiveArrayPairGap|TestCheckedViewMutableArrayUnionRefusal)$'
-count=1 -v (/tmp/views-tuples-lane4-array-gap.log): PASS 0.977s. Lane 4's
mixed scalar array adapter is absent from this branch; its existing six
named refusal witnesses remain refused. No lane 4 work was imported.
Scoped compile probes in IR/lower/native/JavaScript PASS; gofmt applied.
Original tuple counts update passes 32.451s, recording all 29 fixture rows
(/tmp/views-tuples-index-counts.log). Integrator failures remain named below.
The old review patch below is archival; the implemented narrowed dispatch
supersedes it and its previous approval blocker is resolved.

---

Built: staged tuple-array normalization and selectors with semantic probes; certified total remains 3 pairs / 3 reads.
Commits: certified tip 0a6edef2d89f5241f7dbeb392b607fec9fc4acff; probe checkpoint recorded in git history.
Checks: combined original oracles and normalizer PASS 25.226s; selector probes PASS; optional/rest layout probe PASS 0.307s.
Mutants: native and JavaScript selector shape mutants both admit a record and fail their exit-70 oracle; prior six mutants remain recorded below.
Remaining: 7 primary pairs / 12 reads, optional/rest forms, and two overlapping lane 4b frontier checks; shared dispatch needs approval.

The two added lane 4b obligations are queued after the existing unit in
lane4b-handoff.json, from 8abb52a1518b9d8c3f0dbde993551d4f01ba10e2.
They overlap primary IDs 97898 and 97934, so they do not add production reads.
Neither added frontier is claimed certified here.

The staged normalizer is not called by production dispatch. It retains a
reference slot or boxes a selected scalar, reusing the existing array reader
for bounds, holes, undefined and storage checks. Its standalone C probe frees
the source array before consuming the normalized reference, then releases it.
Node, native release, ASan/UBSan and the shared leak checker pass. The tuple
predicate extraction preserves the existing fixed tuple certificate. The
native and JavaScript selectors reuse the mixed-union selector and existing
tuple identity and length predicate; source admission remains unchanged.

Commands (source /workspace/adamic-tools/env.sh first), output only in logs:
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^TestCheckedViewTuple(OriginalOutSignature|OriginalTrackedSymbols|OriginalReferencedMap|OriginalNumericBrandReadRefuses|ArrayNormalizerProbe)$'
-count=1 -v -timeout 3m > /tmp/views-tuples-stage-final.log 2>&1
The run passes in 25.226s with 51 uncached Node observations.
go test ./internal/ir ./internal/native ./internal/javascript -run
'^(TestTupleViewMembers.*|TestTupleHeapUnionSelectionProbe)$' -count=1 -v
> /tmp/views-tuples-selector-final.log 2>&1
IR PASS 0.011s; native PASS 0.300s; JavaScript PASS 0.163s.
ADAMIC_TUPLE_SELECTOR_MUTANT=shape with the native and JavaScript selector
probe fails both semantic exit assertions: wrong record accepted, exit 0
instead of 70 (/tmp/views-tuples-selector-mutant.log).
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations
go test ./internal/lower -run '^TestTupleOptionalRestOriginalLayoutProbe$'
-count=1 -v > /tmp/views-tuples-optional-rest-layout.log 2>&1
PASS 0.307s. This checker observation proves no runtime admission: optional
positions carry undefined; a trailing rest position carries its scalar element
type. Existing unsupported forms still return no tuple certificate.

Automatic approval review rejected shared array-union source/backend dispatch,
then index-only dispatch, then backend-only staging. The stated concern was
unvalidated ownership-sensitive shared changes risking memory errors or silent
miscompilation beyond the isolated probes. None of those patches was applied.
review/array-union-index-dispatch.patch is a concrete review artifact, checked
with git apply --check, not an implemented or validated source path. It proposes
index ownership through existing statement cleanup, leaving callback and loop
transfer refused. Approval is needed to wire that boundary and certify the
next pair. Own fixture counts are unchanged; TestCheckedViewTupleOriginalCounts passes
in 23.487s (/tmp/views-tuples-stage-counts.log), without update-counts. The unrelated integrator failures
below are not the blocker.

---

Built: certified original referenced-map tuple consumers; total 3 pairs / 3 candidate reads.
Commits: prior 49647618880c8953cbb0328044247a0d28f53345; referenced-map commit recorded in git history.
Checks: referenced-map oracle PASS 2.630s; numeric controls PASS 0.216s/0.421s; tuple counts refreshed PASS 25.401s.
Mutants: drop original FileIdListId position check fails the named refusal oracle, printing 7 then false and exiting 0.
Remaining: 7 candidate pairs / 12 reads plus optional/rest Map forms; integrator failures below do not block this worker.

Referenced-map certification imports IncrementalBuildInfoReferencedMap unchanged,
including both required-any numeric brands. No number or void-brand alias is
substituted. The imported numeric carrier hook proves the brand member name
absent from the primitive using the existing inventory, retains number-kind
checks, and leaves demanded any-valued reads refused. Original source hashes
are checked by the existing 78-file manifest. Five fixtures cover valid IDs,
wrong key/value types, tuple length and numeric-keyed records. Node, JavaScript,
native release and sanitized native pass; the successful native fixture passes
leak checks. The numeric position mutant is caught first in JavaScript.

Commands, with output in logs:
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^TestCheckedViewTupleOriginalReferencedMap$' -count=1 -v -timeout 3m
(/tmp/views-tuples-references.log).
ADAMIC_TUPLE_NUMERIC_MUTANT=1 on the same oracle with
'^TestCheckedViewTupleOriginalReferencedMap/references-value-wrong$'
(/tmp/views-tuples-references-mutant.log), expected exit 1.
go test ./internal/lower -run
'^(TestPhantomRefusals|TestPhantomPrimitiveNames|TestPhantomLiteralCastsStayRefused)$'
-count=1 -v -timeout 3m (/tmp/views-tuples-numeric-regression.log).
go test ./internal/oracle -run
'^TestCheckedViewTupleOriginalNumericBrandReadRefuses$' -count=1 -v -timeout 3m
(/tmp/views-tuples-numeric-any-read.log).
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$' -count=1
-timeout 3m -args -update-counts (/tmp/views-tuples-references-counts.log).

The previous checkpoint and the named integrator failures follow.

---

Built: original TrackedSymbol optional forEach consumers through the existing tuple certificate; 2 pairs / 2 candidate reads now certified.
Commits: starting f753567dd8837e90914e7c229df88d8536a3bf8b; pair 6b58eb3c02d57ea341985643ae9528b5597115ac.
Checks: outSignature PASS 26.108s; TrackedSymbol PASS 12.566s; tuple counts update/check PASS 32.037s/37.821s; complete counts refresh blocked.
Mutants: original skip/shape/nested and tracked meaning/guard all fail their semantic oracles; JavaScript catches them before native emission.
Remaining: 8 candidate pairs / 13 candidate reads; optional/rest Map forms remain refused; exact reaching-view coverage is unmeasured.

Eight tracked-*.a fixtures import unchanged TrackedSymbol from the 78 hashed
upstream declarations at 050880ce59e30b356b686bd3144efe24f875ebc8. The oracle
also asserts the full original Symbol field set. Node decides fixture behavior;
JavaScript, native release and ASan/UBSan match the valid programs and pinned
refusals. Successful native runs pass the existing leak checker. Controls cover
nested Symbol.flags, tuple meaning, tuple arity, tuple versus record identity,
undefined receivers, receiver evaluation once and conditional callback creation.

The minimal statement hook saves receiver?.forEach(callback)'s receiver once
and guards the existing ArrayVisit; callback creation remains in that branch.
It handles arrays of required-position tuples only, without null receivers,
optional calls on the method itself, or thisArg. Other optional calls retain
refusals. No second tuple representation, constructor or flow graph is added.
The hook and original counts binding are listed in docs/checked-views-plan.md.
Twenty original tuple fixture count rows are now recorded, including the twelve
previous outSignature fixtures. Original counts remeasurement remains opt-in
with the same external declaration inputs as the existing semantic oracle.

Integrator failures, not a worker blocker: the required TestCountsAreRecorded refresh fails on
fixtures outside this unit: ctor_set.a aborts with free(): invalid pointer;
census_overload_contracts.a refuses excess implementation arguments;
census_small_boolean.a and nbody_field_values.a refuse template interpolation;
maybe_number_slots.a has a native argument representation error; and
require_perf_hooks.a panics in Node.Text on a binding pattern. The three named
lowering failures reproduce with statements.go restored from the exact starting
commit using Go's source overlay, which restores the entire production source
of this checkpoint. The broader lower refusal regression also has three stale
expectations for already-supported union operations; all three reproduce with
the starting-source overlay. The relevant optional-call and tuple-storage
regressions pass (0.136s and 0.319s). No prohibited file was edited, no complete
package or full gate was run, and no successful complete counts refresh is claimed.
The tuple-only counts update and check both pass. The user confirms these
unrelated failures belong to the integrator; work continues through the queue.

Remaining production candidates: four original numeric-brand pairs / seven
reads (97898, 97913, 97926, 97931), and four optional-position receiver pairs /
six reads (95604 fields 0/1; 68230 fields 0/1). A scratch unchanged
IncrementalBuildInfoEmitSignature probe still refuses its representation at
stage 0. No required-any numeric brand was substituted or erased. Optional/rest
Map gap controls pass as named compile refusals, not implementation completion;
no rest production pair has been found in the supplied inventory.

Commands actually run (all test output redirected to separate log files):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/views-tuples-setup-retry.log 2>&1
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/tuples/prepare.cjs /tmp/views-tuples-pinned /tmp/views-tuples-original-declarations > /tmp/views-tuples-prepare.log 2>&1
export ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewTupleOriginalOutSignature$' -count=1 -v -timeout 3m > /tmp/views-tuples-original-baseline.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewTupleOriginalTrackedSymbols$' -count=1 -v -timeout 3m > /tmp/views-tuples-tracked-final.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewMapStorageGaps$/^(optional-tuple|rest-tuple)$' -count=1 -v -timeout 3m > /tmp/views-tuples-map-gaps.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 15m -args -update-counts > /tmp/views-tuples-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$' -count=1 -timeout 3m -args -update-counts > /tmp/views-tuples-owned-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$' -count=1 -timeout 3m > /tmp/views-tuples-owned-counts-check.log 2>&1
go test ./internal/lower -run '^(TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat|TestATupleSeenAsAnArrayIsNotYet)$' -count=1 -v -timeout 3m > /tmp/views-tuples-lower-regression.log 2>&1
go test ./internal/lower -run '^TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat$/^a_(method|function_value)_called_through' -count=1 -v -timeout 3m > /tmp/views-tuples-optional-call-regression.log 2>&1
go test ./internal/lower -run '^TestATupleSeenAsAnArrayIsNotYet$' -count=1 -v -timeout 3m > /tmp/views-tuples-storage-regression.log 2>&1
```

Every mutant is a separate go test ./internal/oracle -count=1 -v -timeout 3m:

- ADAMIC_TUPLE_ORIGINAL_MUTANT=skip with
  ^TestCheckedViewTupleOriginalOutSignature/emit-wrong-kind-noread$:
  exits 0 and prints boolean instead of the expected exit-70 named refusal.
- ADAMIC_TUPLE_ORIGINAL_MUTANT=shape with
  ^TestCheckedViewTupleOriginalOutSignature/emit-record-noread$:
  accepts a record and prints object instead of the named tuple refusal.
- ADAMIC_TUPLE_ORIGINAL_MUTANT=nested with
  ^TestCheckedViewTupleOriginalOutSignature/emit-helper-wrong$:
  prints false instead of the named nested-position refusal.
- ADAMIC_TUPLE_TRACKED_MUTANT=meaning with
  ^TestCheckedViewTupleOriginalTrackedSymbols/tracked-meaning-wrong$:
  prints 1 then false and exits 0 instead of the pinned SymbolFlags refusal.
- ADAMIC_TUPLE_TRACKED_MUTANT=guard with
  ^TestCheckedViewTupleOriginalTrackedSymbols/tracked-evaluation-undefined$:
  improperly creates the callback and exits 70 with a TypeError; Node and the
  unchanged program print only receiver and exit 0. The callback stdout and
  exit-code oracle catch the missing guard; no clang or sanitizer failure is
  credited as a mutant kill.

Logs: /tmp/views-tuples-mutant-{skip,shape,nested}.log and
/tmp/views-tuples-tracked-mutant-{meaning,guard}.log. Starting-source checks:
/tmp/views-tuples-start-counts-failures.log and
/tmp/views-tuples-start-lower-failures.log, using
/tmp/views-tuples-start-overlay.json.

Setup finished successfully after working around the submodule fetch. Its
initial automatic TypeScript checkout fetched full history and was interrupted
(exit 143). A racing manual shallow fetch failed with 'shallow file has changed
since we read it'; sequential exact shallow fetch and checkout of
cohere/TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f then succeeded.
Retry timing lines: Node ready 0.033s; Go 0.043s; submodules 0.088s;
markdown-width skipped, ready 0.090s (step 0.007s); clang 0.258s; Go build
795.430s; test binaries deferred 795.533s; cache warm 795.535s; done 795.618s.
nproc=5; cgroup quota=4; Go 1.27.1, clang 20.1.8, Node 24.19.0. The long cold
build also included redundant initial compilations, which were stopped before
sequential scoped reruns. No timeout cutoff was used for this checkpoint.

Previous checkpoint report follows unchanged.

---

Built: original EmitSignature field views using lane 1's single tuple certificate path and lane 2 array read hooks.
Commits: territory b3bc7138; lane 1 merge 13486c7c (includes 65d8a138).
Checks: original pair oracle PASS 11.507s; scoped IR/lower/native/JavaScript PASS 0.011s/4.593s/39.285s/0.875s.
Mutants: skip root, accept record, drop helper position each fail the pinned refusal oracle; none is killed by clang.
Remaining: 9 candidate pairs / 14 candidate reads after this push; optional/rest Map forms are additional fixture obligations.

The combined candidate queue is 10 pairs / 15 reads. Lane 2 supplies 6/9;
lane 1's nullish table supplies four optional-position receiver pairs / six
reads. No rest receiver is identifiable in that table. The two optional/rest
Map gap fixtures are tracked separately from production counts. Exact reaching
view coverage is not measured. This push covers pair 97934, outSignature, one
candidate read, against unreduced declarations from upstream commit
050880ce59e30b356b686bd3144efe24f875ebc8.

Preparation reuses lane 4b's original declaration emitter, verifies all nine
lane 2 read sites, and hashes all 78 generated declaration files. Tests require
the original IncrementalBundleEmitBuildInfo field set; no reduced interface is
substituted. All twelve fixtures run their unchanged source on Node, generated
JavaScript, native release and native ASan/UBSan. Successful native runs also
pass the shared leak checker. Undefined is admitted; the required field being
missing refuses. Wrong primitive, tuple length, record identity and nested
position diagnostics are pinned literally in the oracle.

Lane 1's constructor is generalized only to accept the existing recursive child
builders. Map producers still require complete descriptors and reject callable
tuple descendants. The array adapter retains lazy descendants. Empty tuples
have an explicit layout marker; Map schemas distinguish them from ordinary
records. Native producers carry tuple identity independently of casts, and
ordinary object copies do not acquire it. These hooks are listed in the plan.

Reproduction (source /workspace/adamic-tools/env.sh first):

```sh
node stage3/interface-downcasts/tuples/prepare.cjs /tmp/untagged-typescript /tmp/views-tuples-original-declarations > /tmp/views-tuples-prepare.log 2>&1
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewTupleOriginalOutSignature$' -count=1 -v > /tmp/views-tuples-pair1-final.log 2>&1
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'View|Map|Contract|Tuple' -count=1 -timeout 10m > /tmp/views-tuples-packages-handoff.log 2>&1
```

With the same original-declaration environment, each mutant is a separate
`go test ./internal/oracle -count=1 -v` run. `ADAMIC_TUPLE_ORIGINAL_MUTANT=skip`
on `^TestCheckedViewTupleOriginalOutSignature/emit-wrong-kind-noread$` prints
boolean and exits 0 instead of the named exit-70 refusal. `shape` on
`.../emit-record-noread$` admits an ordinary record and exits 0. `nested` on
`.../emit-helper-wrong$` returns the wrong position value and exits 0. Logs:
/tmp/views-tuples-mutant-skip-noread.log,
/tmp/views-tuples-mutant-shape-handoff.log,
/tmp/views-tuples-mutant-nested-handoff.log. IR mutations do not alter production
sources. The earlier skip probe was masked by secondary narrowing and was
replaced by the consumer that does not narrow.

Setup: GOPROXY=https://proxy.golang.org|direct; submodules 0.077s; markdown-width
skipped (ready 0.090s); clang ready 0.182s; Go build 37.033s; deferred test
binaries 37.214s; cache ready 37.216s; done 37.247s. nproc=5, cgroup quota=4.
Go 1.27.1, clang 20.1.8, Node 24.19.0. No complete repository gate is claimed.

Target for the optional/rest adapters is October 8. A completion date for the
whole original-declaration family is contingent on certifying original branded
numeric IDs: the highest-ranked tuple uses IncrementalBuildInfoFileId, whose
original required brand is any. The unchanged original read currently reports
NotYet for that numeric intersection. No number alias or void-brand replacement
will be counted as certification of that original declaration.
