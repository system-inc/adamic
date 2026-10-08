Integrated every requested pushed lane into codex/views-integration from ab4d6f902.
Owner e6aec805 was first; newer owner 5fd6445e was prioritized and array tip 25ed1d3f is included.
Focused compiler packages, full IR, filtered Node oracle and vet pass after every published merge.
101 explicit implementation/component mutants were caught, with production restored or untouched, alongside fixture payload mutants.
Full repository gate and production census were not rerun; baseline failures and deferred source families remain.

Merge order follows the user's owner-first ruling. This first merge is clean,
with no conflicting hunks and no bulk or whole-file resolutions.

The integration repair in internal/lower/readiness.go applies the existing
markProgramViewArrayUse policy inside recursive struct traversal, so nested
for-of statements receive the same policy as top-level statements. Ordinary
nested loops previously retained preliminary view metadata despite no admitted
array view. The filtered Node fixtures library_array_copy_within.a,
string_views_methods.a and string_views_policy.a prove the repair. The rollback
Go overlay removes exactly that traversal repair; string_views_policy.a then
executes valid C and exits 70 with uncertified storage, against Node exit 0.
The owner's demanded callable-read refusal and tests are retained unchanged.

Validation commands (each writes full output to its named log):

```
go test ./internal/lower ./internal/ir ./internal/native -run 'TestView|TestLazyView|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestOptional' -count=1 -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View|Test.*Phantom|TestNativeAgreesWithNode/internal/oracle/testdata/.*(view|union|brand|optional|array)' -skip 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_holes_callbacks.a' -count=1 -timeout 30m
go vet ./...
```

The unskipped filtered oracle was run first: library_array_holes_callbacks.a
refuses at line 12:37 with adamic/no-type-predicate for the inline predicate
(value) => value !== undefined. The same refusal is reproduced in detached
c01ae313. This baseline fixture is explicitly excluded from the green rerun.
A broader package filter also reproduces TestPhantomArrayCastsAreErased,
TestPhantomArrayRequiredCastsAreErased and TestPhantomArrayProofs/cycle failures
on c01ae313. The parent also fails the tuple_representation refusal wording pin.
These failures were not suppressed by compiler or test changes.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh;
source /workspace/adamic-tools/env.sh. Node ready 0.072s, Go 0.098s,
clang 0.504s, Markdown 0.948s, submodules 662.275s, Go build 1177.123s,
setup done 1177.419s; nproc=5 (CPU quota 4). Stage3 API pinned Node declarations
were installed with npm ci --prefix stage3/api. Early gate attempts before
submodule readiness failed on missing cohere/TypeScript/tsc/go.mod; subsequent
concurrent cold gate attempts were interrupted and not counted as evidence.

Existing carried evidence has whitespace errors; historical logs and raw
fixture evidence are preserved. No cohere code was copied.

## Lane 4 reconciliation

Merged tip 9a385606. Eleven conflict hunks across eight files were reviewed
individually. No whole-file selection or merge strategy was used.

| File | Hunk choice and evidence |
| --- | --- |
| docs/checked-views-blockers.md | Keep both distinct dated demand inventories and their limits. |
| docs/checked-views-plan.md | Keep lazy/callable handoffs and lane-4 scope rulings; latest primitive-brand ruling supersedes earlier runtime-brand requirement. |
| internal/lower/expression.go | Keep owner resolved-result path; retain shorthand symbol resolution and lane-4 overloaded-value refusal. Phantom overload and owner result-stop fixtures prove both boundaries. |
| internal/lower/functions.go | Keep owner implementation resolution and generic overload proof; incorporate approved phantom-result exception in shared census proof. |
| internal/lower/modules.go | Keep owner bodyless-header skip and registration-time overload proofs, including unused-result refusals. |
| internal/lower/interface_cast.go | Both read hunks retain owner optional/nullish/accessor policies and named lazy refusals while using receiver-aware primitive-brand recognition. Literal hunk retains both phantom-base recursion and open numeric enums. |
| internal/lower/view_contracts.go | Retain undefined, nominal, recursive and lazy metadata; classify only approved phantom bases as scalar and retain their finite literals. |
| internal/lower/view_objects.go | Keep owner MaybeNumber/MaybeBoolean support and optional write registration, plus lane-4 brand-aware data predicate. |

The shared overload proof keeps generic alpha-renaming, strict parameters,
nullable result checks and predicate result handling. Only the existing
phantomOverloadResult proof admits a phantom result difference; nongeneric
invalid results use lane-4 refusal diagnostics. Incoming argument fitting is
applied before the owner's resolved-result conversion, so undefined checks are
not bypassed by premature fitting. One parameter test now pins the owner's
more specific diagnostic, keeping the same unsafe mutable-parameter source.
The acceptance-all-results Go overlay mutant is held to the branded-literal
refusal pin. This does not introduce a second flow graph or callable convention.

Brand-string-good/wrong and brand-string-literal-good/wrong match positive Node
controls and retain both backend exit-70 checks on malformed payloads. Their
valid-release read-removal mutants run as part of the checked-view oracle.
General intersections remain deferred; only phantomBase-approved primitive
intersections bypass that unsupported-family classification. Standalone mixed
selectors remain components; complete source union admission is not claimed.

A complete IR-package run exposed the inherited emitter-name allowlist mismatch
and the two overload readers. Both overload helpers now use Program.CallTargets;
the two permanent emitter allowlist keys name evaluateWithoutViewArrays, matching
lane 2's rename without broadening the allowlist. The complete IR rerun passes.
Focused lower/native, full IR, the filtered oracle with the same baseline skip,
and vet are the lane-4 gates. Owner overload result-stop/append controls are
included beyond the view filter. The all-results covariance mutant is caught
with a valid build: TestPhantomOverloadBrandLiteralConstraintRefused sees nil
instead of the required branded-literal refusal.

## Lane 5 reconciliation

Tip 9c4d0904 includes newer native-array lane 2 bf544d5c through 1e47e792.
Three conflict hunks across two files were individually reconciled:

- docs/checked-views-plan.md, two hunks: retain lazy/lane-4 checkpoints and both
  callable shape/read-producer handoffs as dated evidence.
- internal/native/emit_expressions.go, ArraySlice hunk: use lane 2's sparse-aware
  adamic_view_array_slice when array views exist, retaining graphArray and
  GraphTypes from the owner. Automatically merged ArrayPush keeps graph-held
  reference ownership plus sparse-aware push dispatch.

The newer array runtime consumes element_kind, certifying physical allocation
storage without a second view storage byte. Checked sparse pop/copy/join and
write-kind refusals are retained. Fixtures native-array-sparse,
native-array-sparse-pop-hole, native-array-copy-bad, native-array-write-bad,
native-array-push-bad and native-array-join-literal pin these paths, alongside
native-array-string-mutation and large-field/evaluation controls.

Fixed callable signature/read/producer adapters remain components where shared
source wiring is deferred. No unsupported callable refusal, Unknown frontier,
receiver convention or void signature was enabled by this merge. Three isolated
Go-overlay mutants were rerun and caught semantically: native-forge-producer by
TestViewCallableProducerCertificateNative/wrong-arity, javascript-forge-producer
by TestViewCallableProducerCertificateNode/wrong-arity, and eager descendants by
TestPrepareViewCallableRead/interface_Result. Production files were not mutated.

Focused lower/native/JavaScript, additional prepare-read adapter, full IR,
filtered checked-view/Node oracle (same recorded baseline exclusion), and vet
are the gates. The full repository gate remains unclaimed.

## Lane 4b reconciliation

Tip d11f3ae2 adds the assigned object/primitive plan and .a probes, with no
compiler admission change. The one plan hunk keeps the existing callable
read/producer handoff and integration checkpoint plus lane 4b's distinct scope,
ranked shapes and conditional estimate. Probes jsdoc-good/wrong,
option-type-good/wrong, node-indicator-good/wrong and diagnostic-good/wrong
are frontier evidence, not newly accepted source contracts. No mutant claim
is made for this documentation/probe-only merge. The same focused package,
full IR, filtered checked-view/Node oracle and vet gates are rerun.

## Lane 2 search follow-up

Tip 25ed1d3f (implementation 12482f1f) merges after lane 4b at the lead's request.
The only remaining conflict is one plan hunk: retain all dated owner/lane
checkpoints and append centralized integration coordination. The ArraySlice hunk
recorded by lane 2 against f1c91970 was already resolved in 8165756d: sparse-aware
slice dispatch plus graphArray/GraphTypes ownership. No emitter conflict recurs.
Search read metadata merges with owner readiness, including the nested ForOf fix.
Reference-source write refusals remain in both backends; search bad/literal, lazy,
sparse, object identity and evaluation fixtures pin behavior. All nine new search
mutants are rerun before publication; prior eleven array mutants remain carried
evidence unless explicitly rerun. Gates use the same recorded baseline exclusion.

## Lane 4c untagged reconciliation

Tip 6b3fc480 adds selector components without production source admission.
The sole docs/checked-views-plan.md hunk retains all integration/lane handoffs
and the incoming selector handoff. Owner lazy admission replaces the old eager
cast refusal with the named unsupported untagged-object-union field read. The
source-frontier assertion is updated to that observed refusal for binding-name,
option-element and structural view-read fixtures, preserving their rejection.
The initial old-message failure is logged. All 84 pairs / 186 reads remain pending;
component nested mutants prove the adapter seam, not compiler propagation.
Six selector semantic mutants are rerun and restored before normal gates.

## Lazy owner priority follow-up

Dictionary merge was aborted untouched when fetch revealed owner 5fd6445e.
This priority merge is clean: owner had consumed integration 4d863661, and no
previous conflict resolution is replaced. Native Error now physically owns only
name/message and stamps their actual string kinds; optional code is absent.
Optional-error and optional-error-fields positive controls match Node; number
and null code payload mutants retain exit-70 refusals in both backends.
The adapted census is carried evidence (1,758 tagged / 1,178 untagged descriptors,
zero production entries compiled, 259 checker diagnostic rows); it is not rerun
in this integration session. Its remaining runtime reachability is unmeasured.
A producer-certificate rollback mutant is run before normal gates.
The producer rollback is caught by optional-error-fields: Node exits 0, valid
native exits 70 at source.message with unsupported representation. The runner's
initial textual check expected uncertified storage, but this oracle prints byte
arrays; semantic exit evidence is verified from the captured log. Production
exceptions.c was restored in finally.

## Lane 6 dictionary reconciliation

Tip 753cca96 adds ranked demand and source probes only. One plan hunk retains
all earlier handoffs plus dictionary representation/probing/registry obligations.
No compiler check is added, so no new dictionary mutant is claimed. All 409
pairs / 2,256 reads remain pending. The frozen ranking is reproduced byte for
byte; existing record semantics are checked against Node as a representation
control, not checked dictionary admission. Full IR, focused compiler packages,
filtered checked-view/Node oracle and vet are rerun with the same baseline skip.

## Lane 7 intersection reconciliation

Tip bd7042cc adds constituent decomposition and conjunction components. The
single plan hunk retains every prior dated handoff plus the intersection
all-members discriminator/probe requirements. No shared compiler resolution
recurs, because inherited phantom-brand code is already integrated. Conjunction
requires every runtime member; the primitive phantom lane remains separate.
Production wiring is deferred, and all 198 pairs / 1,145 reads remain pending.
Six independent component mutants are rerun and restored: native/JavaScript
skip-check, any-member instead of all-members, and dropped nested adapter.
Nested mutations prove the component adapter seam, not source propagation.
Normal focused compiler packages, full IR, filtered checked-view Node oracle
and vet follow, with the same recorded parent array-callback exclusion.

## Final integration verification

The final restored intersection gate passes: lower 16.018s, native 27.202s,
JavaScript 3.458s, full IR 31.324s, oracle 84.397s; go vet ./... exits 0.
Every gate writes its complete output to integration/logs before review.
Final commands:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript -run 'Test.*View|TestUntaggedView|TestLazyView|TestPrepareViewCallableRead|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestOptional|TestMixedUnion|TestPhantomOverload' -count=1 -timeout 30m
go test ./internal/ir -count=1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View|Test.*Phantom|TestNativeAgreesWithNode/internal/oracle/testdata/.*(view|union|brand|optional|array)' -skip 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_holes_callbacks.a' -count=1 -timeout 30m
go vet ./...
```

The sole filtered-oracle exclusion was rerun on the final tree. It still fails
at library_array_holes_callbacks.a:12:37 with the exact adamic/no-type-predicate
refusal for (value) => value !== undefined, matching detached c01ae313.
The separate failure log is views-integration-final-array-callback-baseline.log;
it is a failure, not a passing test. Other reproduced lower baseline failures
are listed above; full native graph-region failures are carried owner evidence,
not a new full-native run. No requested lane was omitted for a new failure.

The 27 explicit mutants comprise nested-loop rollback (1), unsafe overload
covariance (1), callable producer/read adapter defects (3), array search/source
write defects (9), untagged selector/adapter defects (6), Error producer rollback
(1), and intersection conjunction/adapter defects (6). All execute meaningful
refusal/output checks; none relies on a build error. Component adapter mutants
do not certify missing compiler propagation. Previous eleven array mutants are
carried lane evidence, not reruns in this integration session.

Setup cache ready 1177.301s and done 1177.419s; nproc 5, quota 4.
The compiler census and complete repository gate remain outside this run.
The branch is ready to receive subsequent lane tips through this integrator.

## Lazy owner presence evidence follow-up

Priority tip a0cc0afd merges cleanly before dictionary e9a06829. It adds
optional-presence source probes and measured refusals only, without compiler
changes or replacement of prior reconciliation. Presence operators and broader
nullish source dispatch remain explicitly incomplete. No new production check
is added, and no additional mutant claim is made for this evidence-only merge.
The usual focused package, full IR, filtered Node oracle and vet gates follow.

## Dictionary read components follow-up, October 8

Tip e9a06829 merges cleanly on the priority owner merge 566c1ec03. There are no
conflicted hunks and no bulk/whole-file choices. The plan records the user's
new permission for lanes to add their named minimal shared hooks. This tip
declares viewDictionaryContractHook, but shared source classification/indexed
dispatch and record producers remain unwired. Its descriptor stays Unsupported
dictionary source dispatch; no scalar table is reinterpreted from a target cast.

Native reads only source-certified boxed records and retains child contracts;
JavaScript uses own data descriptors and refuses accessors without invoking them.
Five options/nested/array component source controls compare with Node, including
sanitized native. Eleven executable component mutants run in the oracle gate:
skip-check, accept-wrong-shape and drop-transitive-contract each in native/JS;
native uncertified storage; JavaScript container, accessor, missing-child and
required-missing guards. Mutations are in-process harness copies, never changes
to production files. Nested mutants prove the adapter seam, not compiler flow.
The candidate ranking reproduces exactly: 29 pairs / 228 reads still pending,
with exact runtime reachability unmeasured. Formatting and source diff checks
pass. Focused compiler packages (including records against Node), full IR,
uncached filtered Node oracle and vet are the integration gates. The previously
reproduced array-callback predicate fixture is still the sole oracle exclusion.

Dictionary component gates pass: lower 10.935s, native 44.416s, JavaScript
1.303s, full IR 24.167s, uncached filtered oracle 78.664s, vet exit 0.
All eleven new component mutations execute valid code and lose the expected
refusal, as pinned by the passing mutant tests. No source pair is completed.

## Lane 4b production hook follow-up, October 8

Tip dc5e5f26 merges cleanly after dictionary 422e80502. Four minimal shared
hooks in lower/interface_cast.go, lower/object.go, native/view_fields.go and
javascript/javascript.go preserve existing guards while selecting the lane's
one-object-plus-scalar adapter. No conflict hunk required resolution.

Additional integration controls found a real optional regression: the new reader
dropped Property.Absent and Property.Optional. Native exited 70 for valid absent
optional fields and missing optional-chain receivers, against Node exit 0.
The lane helper now passes both existing flags to its native runtime and shared
JavaScript read helper. Optional receiver short-circuiting evaluates the receiver
once; required missing fields and reserved uninitialized slots still refuse.
This repairs metadata plumbing without another presence/readiness state.

Permanent .a controls optional-absent and optional-receiver match Node in both
backends; required-missing pins the original named refusal. Four valid-code IR
rollback mutants clear the optional flags independently in native/JavaScript and
are caught by the positive Node pins. The incoming twelve backend IR mutants
pin outer selection/literal and nested type/read checks. An initial test-control
excess-property annotation was fixed; the new required-missing control was
excluded from the incoming mutation selector which has no applicable nested
read. Both initial failures are logged; neither is counted as a passing gate.
All 42 original candidate pairs / 181 reads remain pending original witnesses;
reduced shapes do not establish full tsc declarations or exact reachability.

Lane 4b gates pass: lower 13.612s, native 34.968s, JavaScript 1.896s,
full IR 27.981s, uncached filtered oracle 82.549s, vet exit 0. All sixteen
backend mutants are verified in the final log (12 incoming, 4 optional-policy
rollbacks); all eleven source controls pass. The same sole parent predicate
fixture exclusion remains. Full repository/census reruns are not claimed.

## Lazy owner nullish and callable follow-up, October 8

Priority tip 36de78996 is integrated before lane 4b f4e5993e. Two conflicted
hunks are reconciled individually, with no whole-file or strategy selection.
The interface_cast.go read guard retains both nullable/callable exemptions and
objectPrimitiveViewType while preserving accessor refusal and owner family
naming. The object.go slotless guard retains both null-containing unions and
object-plus-primitive unions. Nullable Map reads remain named collection refusals.

Overlapping auto-merged dispatch needed explicit reconciliation: JavaScript's
object-primitive dispatcher excludes Nullish reads; readObjectField marks
nullish metadata only when the object-primitive adapter is not the matching
contract. This prevents the broad nullish mask from replacing finite primitive
membership selection for true|Node|undefined. Owner nullish matrix and callable
signature controls prove the new path; lane4b node-indicator-false,
optional-absent, optional-receiver and required-missing prove retained behavior.
Source producers, readiness, narrowed-kind checks and distinct null/undefined
identities remain owner behavior. No asserted type supplies a producer fact.

Validation uses the standard focused lower/native/JavaScript filter (plus
TestRecordsAgainstNode), full internal/ir, uncached filtered Node oracle and
full vet. The sole exclusion remains library_array_holes_callbacks.a, reproduced
on the parent as recorded above. No complete repository or census rerun is claimed.

Owner nullish gates pass: lower 11.092s, native 40.954s, JavaScript 1.298s,
full IR 21.062s, uncached filtered oracle 115.168s, vet exit 0. Four independent
implementation mutations were executed and restored: collapse-null is caught
by number-both Node output, allow-forbidden-undefined by number-null-opposite's
exit-70 pin, skip-kind by string-both-wrong's named field refusal, and
skip-callable-proof by the callable-signature-mutant read refusal. Every mutant
fails an executed oracle assertion, with no build failure; individual logs and
the runner summary are preserved. Source wrong/missing/opposite/nested fixtures
also run in the standard oracle, distinct from these implementation mutants.

## Newest lazy owner finite literals follow-up

Tip d115cfd5abc8274164b5ea29fbe26f9a25b78ce3 merges cleanly on f56c432e1.
No conflicted hunk or compiler change is present. Nullable finite string members
now have explicit dynamic-string source controls: literal-both agrees with Node
for present/null/undefined; literal-both-wrong constructs not-allowed at runtime
and must exit 70 at node.value naming the finite declared type in both backends.
The owner reports 57 source mutations across the nullable matrix, callable proof,
nested reads and this literal control; these are distinct from the four restored
implementation mutations rerun above. All source tests run in the filtered gate.

Finite-literal owner gates pass: lower 9.476s, native 14.583s, JavaScript
2.377s, full IR 3.172s, uncached filtered oracle 85.780s, vet exit 0.

## Lane 4b array alternatives and boxed producers follow-up

Tip f4e5993e3ed2817a9060ed53ca0b957ea5588c64 follows both priority owner
merges. Three conflicted hunks in checked_views_object_primitive_unions_test.go
are resolved separately: keep both optional and new comment/literal fixture
rows, keep optional-policy rollbacks and new outer wrong-member mutations, and
combine required-missing/literal-text-wrong exclusions for the nested mutant
selector. No refusal or Node control is removed. Backend auto-merges retain
Absent/Optional plumbing while accepting the shared ViewArray member adapter.
The named objectPrimitiveBoxedField producer hook preserves real boxed storage;
casts supply no producer kind. Shared array element and nested field obligations
remain checked. Map/callable alternatives and true NodeArray metadata stay pending.

New comment-good and literal-good controls execute both generated boxed producer
branches. comment-boolean/literal-boolean hold outer refusal and independent
wrong-member/removal mutants; comment-flags-wrong/literal-negative-wrong hold
nested type and dropped-read mutants. literal-text-wrong pins nested string
refusal. Earlier optional-absent/optional-receiver/required-missing and all their
mutants remain in the same source test. The new group contributes 12 release
backend implementation mutants, separate from payload mutants and previous runs.
Original 42 candidate pairs / 181 reads still await original witnesses; reduced
array aliases are not full upstream NodeArray certificates.

Array/boxed gates pass: lower 8.606s, native 19.043s, JavaScript 2.071s,
full IR 23.286s, uncached filtered oracle 107.157s, vet exit 0. The new twelve
backend mutants and preserved sixteen earlier lane4b backend mutants all run
in the gate; source controls include optional absence/receiver and required
missing slots. Sole parent predicate-fixture exclusion remains unchanged.

## Lane 2 NodeArray and own-slot mutation follow-up

Tip 3b1263b650b374183da1cd879b29a73dfb10135e follows lane4b 563a5bfc0
locally. Publication of that predecessor failed twice: Git could not read an
HTTPS username. The current environment reports no secret/outbound identity.
Remote remains 1d165a7c9. A complete first-parent mail patch is at
/tmp/views-integration-563a5bfc0.patch; no branch history was rewritten.

The one plan conflict retains dictionary evidence/shared-hook amendment and
the full new Lane2 hook list. Compiler hunks auto-merge without discarded
behavior. Instantiated array ancestry retains own-field descriptors; fresh
Object.assign array production copies scalar properties with original slot
certificates. Own reads/mutation use existing readiness/literal/write checks.
ArraySearch demand now participates in shared allocation flow.

The overlap probe passed all NodeArray and lane4b source cases. Optional array
wrong-kind still exits 70 with the field, declared type and found number; its
incoming exact diagnostic pin used the older field reader. Updated only that
pin to the owner nullish dispatcher message: matches no member of Declaration[]
| undefined. This preserves the refusal and validates the actual shared path.
Initial failure is saved as nodearray-overlap.log, not counted as a green gate.

Landing hold from the user: integration must not be advertised ready to land
until worker 01a115a4's optional-boolean storage check in view_writes.go is
settled. codex/view-write-optional-boolean has no pushed ref at this check.
Continue lane integration; merge its published fix before any pending lane
and explicitly record the result. This hold is independent of publication auth.

The first full filtered Node oracle exposed a new phantom-array regression:
phantom_array_brands.a:9 .sort became an unsupported checked-array consumer.
A restored diagnostic trace shows refuseOptionalWidening calls viewSchema for
BrandedReadonly before lowering. The broadened strictViewContract dispatch
registered ViewArray for the pure phantom intersection, activating global array
consumer policy. Its old erased contract path never activated that policy.
The array dispatch now excludes only phantomArrayBase-proven pure brands,
preserving the prior path while real NodeArray scalar own fields use the new
adapter. No consumer allowlist or original relation/presence proof is weakened.
The initial full failure and descriptor trace are retained as failed evidence.
The numeric wrong-element omission witness can reach an invalid pointer;
added node-array-wrong-reference instead returns a real string heap payload
through a root-only comparison. This mutant can finish valid sanitized code,
so its catcher does not depend on a later dereference or sanitizer crash.

Six NodeArray implementation mutants were executed and restored. Native and
JavaScript element-kind omission use node-array-wrong-reference and each prints
present with exit 0, caught by the exit-70 pin without sanitizer faults. Native
and JavaScript own-kind omission fail node-array-wrong-own's pin; native property
aliasing fails node-array-properties-copy's 3:12 Node output; omitting original
slot certification fails node-array-write-literal's write-site refusal pin.
A seventh rollback removes only the phantom dispatch exclusion and is caught
by phantom_array_brands.a's previously valid Node control. Earlier five optional
array mutants are carried lane evidence, not rerun against the changed owner
nullish entry point. Their source wrong/lazy/transitive controls run normally.

Final NodeArray gates pass: lower 14.534s, native 22.807s, JavaScript
1.528s, full IR 21.224s, uncached filtered oracle 146.049s, vet exit 0.
All seven implementation mutants were restored before this gate.

Lane 4b original-contract witnesses, d70bc4ddb7d71fd900c3d7a8ec682f38a1608cc4

Resolved two hunks individually. In internal/javascript/view_unions_object_primitive.go,
kept the optional-receiver short circuit as well as separate Optional and Absent
arguments. In internal/native/runtime/view_unions_object_primitive.c, kept
optional null-receiver return independent of the field Undefined flag, and
retained absent-slot permission only when both allow_absent and undefined hold.
The incoming combined condition would refuse an optional receiver when its
field type lacks undefined. optional-receiver, optional-absent, required-missing
and the uninitialized source-file witness prove these distinctions. Other
incoming code and witness additions merged unchanged.

Original declaration generation used pristine TypeScript v6.0.3 commit
050880ce59e30b356b686bd3144efe24f875ebc8. The adapter generated 78 declarations
and matched 20 + 15 read sites. ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS was set
for the oracle, so all nine original-pair controls actually ran, with manifest
and declaration validation. All 24 new backend mutants were caught: readiness
loss (2), rejection of permitted absence (2), nested-shape omission and dropped
transitive checks (8), wrong union member and outer-check omission (12).
Mutation happens on independent backend test inputs; production remains intact.

Gates: lower 12.136s, native 19.806s, JavaScript 2.223s, full IR 2.764s,
uncached filtered Node oracle 124.131s, vet exit 0. Logs use original-pairs
prefix; upstream setup and original declaration preparation are also retained.
Only library_array_holes_callbacks.a is excluded for the recorded baseline
no-type-predicate refusal. Full repository and production census were not run.
These are original-contract witnesses for two pairs and 35 candidate reads;
40 candidate pairs and 146 reads remain, with whole-program reachability
unmeasured. The optional-boolean landing hold remains active; its branch had
no published ref at this check.
