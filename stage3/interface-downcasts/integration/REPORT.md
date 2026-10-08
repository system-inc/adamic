Integrated checked-view lane checkpoints into codex/views-integration from ab4d6f902; further tips remain queued.
Owner e6aec805 was first; newer owner 5fd6445e was prioritized and array tip 25ed1d3f is included.
Focused compiler packages, full IR, filtered Node oracle and vet pass after every published merge.
258 explicit implementation/component mutants were caught, with production restored or untouched, alongside fixture payload mutants.
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

Newest lazy owner nullable selection, c809598e4443d2a20f4a8b2e9b260ce6069b4b5e

Two individual conflicts in lower/interface_cast.go and lower/object.go:
retain objectPrimitiveViewType, scalar/maybe representations, callable exception
and accessor refusal; expand only the Union nullish exception to include
undefined as well as null. The incoming owner guards contain both behaviors.
Fixture controls scalar-undefined, tagged-undefined and object-undefined exercise
the new route, while optional-receiver, required-missing and original-pair
controls retain the earlier policies.

The owner selector validates present scalar member literals and finite tagged
object union members after the shared presence/readiness/kind check. Nested
payload checks remain lazy. Unsupported/ambiguous member families retain named
read refusals; no producer certificate comes from the target.

All 21 owner selection fixtures passed the uncached preflight (12.253s). Two
actual backend omission mutants were executed and restored: remove only the
nullishMemberSelection call, native then JavaScript. scalar-both-wrong constructs
not-allowed at runtime with the valid string kind, and both mutants run valid
code but fail the pinned refusal. Evidence is nullable-selection-mutants.log and
its per-backend logs. Existing source-kind guards alone cannot catch this case.

Optional-boolean landing hold remains active until the user lifts it; the
priority fix branch has no published ref at this fetch. Production read census
and whole-repository gate remain unmeasured.

Final nullable-selection gates: lower 18.751s, native 26.510s, JavaScript
3.720s, full IR 31.124s, uncached filtered oracle 185.782s; vet exit 0.
Original declarations were enabled and all earlier original-pair controls ran.
The sole documented array-callback baseline exclusion is unchanged.

Priority lazy owner Map certificates, f7a784c4e094e0c97dae9dc8309a900e4cd3d666

Clean merge, no conflict hunks. Constructor schema ids are kept independently
of the target view. MapCertificatePairs permits scalar readonly covariance but
requires reverse compatibility for mutable maps and identical physical types.
Native headers and JavaScript's private WeakMap carry original evidence; nullish
selection checks a Map alternative's certificate after kind/presence/readiness.
Unsupported branded and aggregate schemas remain named demanded refusals.
The view_writes.go addition excludes ViewMap as a writable-slot certificate; it
does not settle the optional-boolean storage investigation or lift the hold.

Three actual implementation mutants were executed and restored. Native omission
of the zero-compatible-schema refusal and JavaScript certificate omission each
run scalar-both-value-schema past the expected field trap. Removing the mutable
reverse relation runs mutable-invariant past that trap. All are semantic failures
of executed witnesses, with valid compilation and no sanitizer error. The first
native mutant draft was rejected by clang's tautological-comparison warning; its
log is preserved as a build failure and is not credited as a mutant kill.

Evidence uses map-cert-* logs. Production inventory is still unmeasured; the
owner's static 2,018 pairs / 9,101 reads are not reduced by fixture success.
Optional-boolean hold remains active until explicitly lifted by the user.

Final Map gates: lower 22.851s, native 69.684s, JavaScript 4.904s,
full IR 50.966s, uncached filtered oracle 231.796s, vet exit 0.
Supplemental existing Map/Set/record regressions: native 71.781s and oracle
49.497s, both pass. Original declaration suite was enabled. No new failure or
additional exclusion; sole recorded array-callback baseline exclusion remains.


Callable source and stored markers, b6e53fb46bc623c7017ba8d86b9e995ff4a1d912

Three conflict hunks resolved separately: plan append keeps both lane hook lists;
interface_cast.go uses callableViewContract while preserving the owner's nullable
admission and guards; object.go prepares the callable property before retaining
all owner Nullish/Narrow metadata and object-primitive exclusions. No whole-file
resolution. Group1 direct/stored/callback/arity/result controls, nullable selection
controls, and stored-marker string/arity controls prove the combined behavior.

Cross-lane repairs keep nullable callable implementation proofs and add callable
producer certification after nullish readiness, presence and membership checks
in both emitters. callable-undefined-arity.a pins an arity-one producer against
() => string | undefined at the root-only read; two executed backend IR mutants
remove only its callable certificate and print present instead of the expected
exit-70 arity diagnostic. UndefinedAllowed is passed alongside optional absence
when checking signatures. Existing nullish proof refusals remain active.

Ten source implementation mutants were caught: stored-marker shape omission and
discarded string-result release omission, read-contract omission, native shape
omission, both arity omissions, JavaScript shape omission, native method-signature
omission, lower required-marker proof omission, and write-back guard omission.
Two executable component kind mutants (native number acceptance and JavaScript
number invocation) were independently caught by TestViewCallableShape*/number.
The first helper runner used the wrong Go executable and failed before executing;
that toolchain failure is not credited. All production files were restored.

Behavioral overlap preflight found only a count-row mismatch: optional-absent now
performs seven releases rather than six through the owner's nullish path. The
measured row was regenerated; allocations/frees remain balanced (1/1).
The count update passed in 4.051s. Original declarations remain enabled in the
final gate. Optional-boolean hold stays active until the user explicitly lifts it.

Final callable gates: lower 17.979s, native 41.340s, JavaScript 4.392s,
full IR 37.933s, uncached filtered oracle 208.411s; vet exit 0.
Original declaration controls ran. The sole array-callback baseline exclusion
remains; no new failure or further exclusion. Evidence: callable-* logs.


Priority lane 4, 67d34f3cfbb92c21a0986933c70a529969736a51

Five conflict hunks resolved individually in four files. docs/checked-views-plan.md
retains both appended hook/coordination sections. javascript.go adds declared
undefined payload evidence, restricts object union dispatch to object values, and
retains the owner's mapViewCertificate after dispatch. view_contracts.go keeps
the ViewNull branch and extends only the undefined condition to phantomUndefined.
object.c retains both includes and, separately, both the owner's
adamic_object_read_contract and lane 4's snapshot probe through shared readiness
and actual static-owner evidence. No whole-file pick or bulk resolution.

Preflight exposed an automatic semantic overlap: the owner's broad undefined
path intercepted approved __String string/phantom-void reads. An owned helper
viewBrandedStringUndefined selects the existing scalar read only when the physical
representation is string, null is excluded, and an approved phantom-undefined
member is present. Ordinary nullish reads remain on the owner's path. Required
missing fields still trap; explicit undefined is allowed. CompleteBrand and all
12 BrandCandidatePairs retain their original exact diagnostics and Node controls.
The first undefined-admission overlay was redundant under the intercepted path
and survived; it is preserved, not credited. Both overlays are caught after the
repair. A scratch-log relative-path error was corrected; not a mutant kill.

FiniteStringKeyComponent retains every exit-70 wrong-member refusal, with the
owner's nullable selector diagnostic now pinned. The owner already implements
open string/number/undefined selection, so stale eager compile-gap assertions
were replaced by TestCheckedViewOpenCompilerOptionKeySelection. Its twelve
source controls hold string, number, undefined and absence to Node and pin
boolean/null at value.skippedOn, expected keyof CompilerOptions | undefined.
Eight backend omission mutants print uncheckedunchecked at exit 0. The first
unboxed test replacement failed its ABI build; it was not counted. The corrected
mutant uses ir.Box and executes valid release code. Finite-key and open-key
controls pass 10.452s. Primitive mixed array admission remains refused.

This is reduced open-key/component evidence, not full original CompilerOptions
or dictionary extraction certification. Lane 4's dictionary queue is preserved
as pending 11 pairs / 38 static candidate reads. The 14 __String candidate
pairs / 511 reads have representative source witnesses; exact whole-tsc runtime
reachability is unmeasured. Optional-boolean hold remains active.

Executed runtime probe mutants: number-tag and boolean-payload caught by
TestCheckedViewPrimitiveHeap; skip-readiness, ignore-owner, collapse-null,
ignore-reference-storage and trust-unknown caught by PrimitiveProbe. Every
failure is semantic native release output or exit status, not a warning or
sanitizer-only failure. Source restored. Two backend undefined overlays caught
by CompleteBrand/undefined. Additional incoming oracle mutations run in the
final gate: twelve candidate-pair read bypasses (24 backend executions), six
CompleteBrand refusal bypasses and two undefined first-member substitutions
(16), two finite-key bypasses (4), and four open-key bypasses (8). Total 61 new
executed check defects, plus prior 120. Evidence is mixed-* logs.

Final lane 4 gates: lower 9.759s, native 32.670s, JavaScript 1.448s,
full IR 19.061s, uncached filtered oracle 210.849s; vet exit 0.
Original declaration controls enabled. All 52 new backend IR mutation executions
logged their caught behavior; all nine source mutations caught independently.
Sole array-callback baseline exclusion unchanged, no new failure remaining.


Priority owner structural and array Maps, 99089aa11253bd879e9b5ecad8a65c870eee8fc7

One conflict hunk in internal/ir/views.go: keep callable DiscardResult and add
ArrayReadonly; neither certificate substitutes for the other. Structural/array
Map producer descriptors remain source-owned. Readonly covariance, writable
invariance, physical storage equality, every NodeArray own-field schema and
undefined entry restrictions stay enforced. Nested phantom entries are refused
recursively, including below object fields and array elements. Nullable, mixed,
callable and tuple entries retain named demanded read refusals and unread twins.

Overlap preflight passed 41.819s: Map schemas and payload propagation, nullable
selection, NodeArray records, stored callable markers and CompleteBrand.
Structural-schema-unused-payload and array-schema-unused-payload stop at the
Map field even without reading entries, so descendant checks cannot conceal a
missing constructor certificate. Native positives retain leak checks.

Production inventory remains 2,018 pairs / 9,101 reads unverified by whole-tsc
lowering and reaching-view proof; no fixture success subtracts those sites.
Optional-boolean hold remains active until explicitly lifted by the user.

Six implementation mutations executed independently and restored:
structural-kind-only, array-kind-only, skip-own-fields, skip-readonly-source and
skip-undefined-source each let its wrong-schema root-only witness print object
at exit 0 instead of the named Map field refusal. skip-nested-brands admitted
both nested-array-brand-read and nested-object-brand-read, violating the named
compile refusal. These are semantic runtime/admission failures, not build errors
or crashes. The first five run all compiled backend calls in the oracle; the
logged failing observation is the first native mode. No extra backend count is
claimed. Evidence: owner-reference-mutants/ logs. Six new defects, total 187.


The first owner-reference oracle gate failed only map_foreach_arrays.a in both
backends: exit 70, array write expected array, found uncertified source element
contract. Producer Map schemas had interned ViewArray descriptors without any
admitted view. HasArrayViews now requires ViewOrigins before enabling the existing
program-wide checked-array policy. Actual viewed arrays retain all checks.
The regression and Map/NodeArray/optional-array controls pass 39.354s. Removing
this origin requirement restores the valid-code runtime failure in both backends;
the Node agreement oracle catches it. This seventh new defect brings the total
to 188. The original failed gate is preserved; no new test exclusion was added.

Final restored owner-reference gates: lower 11.207s, native 20.568s, JavaScript
2.218s; full IR 16.792s; uncached filtered Node oracle 213.260s; vet exit 0.
Original-declaration controls enabled. Existing array-callback proof exclusion
unchanged; no new failure remains.


Latest callable contracts, f3a2646a09db4666676eacfd09306d0f0f06bb61

Six conflicts resolved individually: two plan hunks preserve the moved lane-5
hook list once, known-void metadata and lane-4 probes/inventory. object.go keeps
the approved branded-string-undefined dispatch exception. view_callables_read.go
attaches complete producer-comparable signatures to boxed nullable Union reads.
Native/JavaScript view_nullish.go keep owner member selection and Map certificates
and use the lane-owned callable certificate afterward, guarded for null/undefined.
Known void 254 is exact, unknown zero and discarded-result 255 remain distinct.

The old nullable wrong-result source still refuses, now at the actual field read:
exit 70 node.value expected () => string, incompatible result representation.
This replaces its older compile refusal only because an independent producer
signature now establishes the ABI and the runtime check can reject the value.
The same-name proof exemption is paired with descriptor completion and both
backend checks, not trusted target annotations. The nullable wrong-arity fixture
reads and would call a string-returning function that never touches its missing
argument; omission evidence therefore executes valid code without an ABI fault.
Overlap oracle passes 68.322s, including Map certificates and CompleteBrand.

FileWatcher.close retains its complete original declaration. Program canonical
handoff keeps the original callable alias but reduces the receiver/helper. Four
candidate member contracts / 34 static reads are fixture-certified; 304 / 1469
remain in the static table. Whole Program and whole-tsc reachability are unclaimed.
Debug.assert and detached intrinsics remain named proof/dispatch boundaries until
separately checked after the queued predicate dependency. Optional-boolean hold
remains active. No cohere source was copied.

Six new callable implementation defects caught and restored: native and JavaScript
arity guards (both canonical-name and watcher wrong-arity witnesses), native and
JavaScript exact-result guards (watcher wrong-result exits 0, no output), and both
nullable-dispatch hook omissions (nullable-wrong-arity exits 0 with forbidden
output). Native canonical wrong-result also crashes without its guard; that is
recorded separately and is not credited as valid-code kill evidence. The watcher
witness establishes the native result guard can fail in valid code. Total 194
executed implementation/component defects. Logs: callable-latest-mutants/.

Final callable gates: lower 15.494s, native 23.472s, JavaScript 1.615s; full IR
30.111s; uncached filtered Node oracle 261.199s; vet exit 0. Original declaration
controls enabled; no new exclusions or failures. The existing array-callback
proof baseline and optional-boolean hold remain.


Prioritized proven predicates dependency, b131a36dc50163896dbbdd0ef6d304223454fd0a

Twenty conflict hunks in eleven files resolved individually. docs/0.1.md keeps
supported labels out of the refused list. ir.go preserves ViewOrigins and
MapCertificates while using the structured ordinary/overload predicate count
records. cast.go (two) and cast_proof.go (three) retain integrated deferred/lazy
views, structural certification, marker proof and writable-source guards rather
than restoring the older eager interfaceView path. invariance.go retains its
qualified-name guard; refusals.go retains deferred/view cast classification.
expression.go keeps allocation provenance and never-expression handling while
recording ordinary calls; its comparison hunk uses concrete specialization AND
retains !includesUndefined, preserving both null and undefined alternatives.

optional_widening_test.go retains checked-view admission instead of obsolete
blanket optional-widening refusals. Its census three hunks retain schema disposition
and diagnostics and adopt typed normalized file paths. field_access_paths.a keeps
the incoming readonly declaration, with the same field order and runtime output.
The four counts hunks preserve all measured integration rows, equal common rows
and the new independent predicate table at the end. No whole-file choice was used.

Parser callback/assertion proof and CLI explain-checks preflight passes: lower
10.453s, CLI 3.749s, load .007s. The three census panic controls initially fail only
computed_relation.a's stale optional-field diagnostic: integrated lazy admission
reaches the named computed-field NotYet. The pin now records that actual refusal;
no production capability or guard was weakened. Restored three controls pass .139s.

The previously excluded library_array_holes_callbacks.a now matches Node under
independent callback-body proofs. The boundary oracle passes 1.025s, also verifying
predicate count rows. Its exclusion has been removed from all subsequent required
gates. Debug.assert still stops at adamic/no-type-predicate for the property callback
declaration; merging proven direct assertions does not certify a viewed producer
with that unsupported contract. The unbound intrinsic read also remains refused.
The optional-boolean storage hold remains until the user explicitly lifts it.

Fourteen new predicate/conflict mutations executed and restored. Six proof mutants:
callback body bypass (ParserCallbackLie), assertion normal-return bypass and helper
termination bypass (ParserAssertionProofErasure), erased specialized null check,
erased specialized undefined/null distinction and dropped scalar operand evaluation
(ParserAssertionBackends, both backends except undefined distinction native-only).
Four census guards: qualified cast, qualified widening, computed-name and early
generic cleanup; their reduced controls catch the original Go compiler panic.
Two report erasures fail exact CLI stderr goldens for ordinary/callback and assertion
calls. Union-slot admission fails both MixedUnionCallbackABIIsPending refusal pins.
Removing !includesUndefined fails source Node agreement for string-both in valid
native code. Total 208 implementation/component mutants. Panic witnesses are kept
separate from valid emitted-code observations. The ambiguous union-runner anchor
stopped before mutation, was corrected and only the two pending cases rerun; it is
not a mutant kill. Evidence: predicates-mutants/ and predicates-extra-mutants/.

Final predicate required gates: lower 13.307s, native 18.174s, JavaScript 1.967s;
full IR 23.823s; uncached filtered Node oracle 276.136s; vet exit 0. Array-callback
fixture is included with no skip. Full CLI 15.401s and load 8.229s pass; restored
predicate/parser/assertion/census proof selection passes 20.392s.

Supplemental predicate/census/generic/regex/output oracle fails three fixtures in
48.390s. An isolated unmodified pushed-parent 8afdc710 reproduces all three in
.098s after its cold checker build. census_overload_contracts.a:13:16 retains the
more-arguments-than-implementation NotYet; census_small_boolean.a:19:16 retains
unsupported template interpolation. maybe_number_slots.a:29:33 previously stops
at the unproven inferred predicate argument `(value) => value === undefined`.
Its independent proof now passes, exposing a later native C build failure: passing
adamic_object* to show(adamic_maybe_number). No binary is admitted by that failed
build. This is a changed frontier on an already-failing supplemental fixture,
not a green full-oracle claim. Both logs are retained. The required checked-view
selection has no failure or exclusion; whole repository remains unclaimed.

Publication note: callable merge 8afdc710 initially failed HTTPS username lookup
twice. The requested first-parent fallback patch remains at
/tmp/views-integration-8afdc7102bbf5c9cdad9f4471923003cac7f8a6f.patch.
The installed GitHub credential helper then pushed it successfully. That helper
is configured only in this repository; no credential value was read or logged.


## Lazy admission 055333e1 left out

Newest fetched owner tip 055333e1ea0c278bc809b0dbbd292fbb0e9b56a1 contains
85e958e2deed551a539a665a35fd492a02a7f4ea. Its automatic merge changes
censusCallableSlotless to admit all boxed union callable slots. The existing
TestMixedUnionCallbackABIIsPending fails both mixed_union_callback and
mixed_union_callback_variance: each expected a refusal but received nil (0.073s).
Preserving the existing union refusal instead makes the new Map certificate
entry-live-mutation fixture fail at line 5:82: stage 0 cannot lower a function
value taking string | number | null (0.068s). A global admission therefore cannot
retain both obligations without a separate callable ABI adapter or scoped proof.
The item was aborted and left out, as authorized; no incoming production code
was committed. Exact evidence is owner-055333-abi.log and
owner-055333-preserved-abi.log. The branch remained at 6a7f1bf3a83e61069b1a380739ed513c0f9b0591.

## Callable scalar witnesses 15b30747

Merged source tip 15b30747f83f57f926fb1b291cda5ba3bc39a27a adds
Scanner.hasPrecedingLineBreak and performance.mark fixtures. One conflict hunk
in internal/oracle/counts.md retains all eight incoming allocation rows before
the complete existing predicate direction table, whose end-of-file placement
is independently checked. The plan additions merge without conflict. No shared
production hook changes. Original signatures and reads retain reduced receivers;
these fixtures do not prove complete compiler contexts or runtime reachability.

All four executable omissions ran: native and JavaScript shape bypass, native
and JavaScript wrong-arity acceptance. Each caused both original member tests
to lose their exact exit-70 refusal in valid emitted code (exit 0); restored
production guards are unchanged. Total implementation/component mutants: 212.
The mixed scalar/boxed result boundary remains refused and has a Node control.
The optional-boolean landing hold remains until the user explicitly lifts it.

Final callable-15 required gates: lower 15.322s, native 20.377s, JavaScript 2.834s; full IR 3.380s; uncached filtered Node oracle 247.320s; go vet ./... exit 0. No oracle fixture exclusion. Gate logs retained under integration/logs.


## Intersection 9e4e13fb reconciliation

Incoming tip 9e4e13fb1bb9c755c1743eb0937906eeea988de8 carries finite,
tagged-arm and recursive intersection dispatch and an original SymbolTracker
witness. Five conflict hunks resolved individually in four files: both plan
appends retain every lane hook/history; JavaScript readiness retains the ninth
undefinedMember flag, null diagnostic and owner Map code while admitting actual
callable kind 8; both nullish emitters retain owner member selection, per-member
Map certificates and callable signature certificates before direct conjunction
validation. Tagged intersection arms already run through owner union selection;
the new conjunction hook excludes that case to avoid duplicate slot reads.
Automatic strictViewContract merging retains phantomUndefined erasure and adds
structural intersections and canonical optional descriptors. The shared allocation
flow graph, lazy descendant refusals and writable-slot certification stay intact.

New source root-read witnesses expose a required-property gap: finite and
recursive string | undefined required fields missing from actual objects print
true and exit 0 in sanitized native, release native and generated JavaScript.
Explicitly present undefined controls also print true as Node does. The repair
separates field.Optional and child.Undefined in both walkers and descriptor tables.
Native adds object_optional_view's explicit payload flag and the named
adamic_object_optional_view_undefined wrapper; viewField forwards the branded
string undefined flag. Existing wrapper uses the same readiness/presence validator.
Optional receiver absence does not authorize absent required fields on a present
receiver, or present undefined payloads without a declared value member. JavaScript
uses the existing ninth flag. Four restored controls pass, and CompleteBrand
controls pass. Before/after logs preserve observed output; no check is removed.

The new presence counterfactual makes each required label optional in production
IR. Both finite and recursive missing cases execute valid code with true/exit 0
in all three modes, failing six independent refusal assertions. This is four
backend-specific defects (sanitizer repetition is not separately counted).
The source tests construct temporary .a witnesses from inline source without
copying upstream/cohere declarations. Original lane tests enable all 78 declaration
hashes and complete unchanged field sets through ADAMIC_INTERSECTION_ORIGINAL_DECLS.
Both original-declaration opt-in suites remain enabled in all later required gates.

Supplemental TestOptionalClassReads fails class_derived, class_generic,
class_generic_compound and class_transitive with stderr differs; class_expression
passes. The isolated unmodified prior integration 8afdc710 reproduces the same
four failures (3.075s). These are existing supplemental failures, not a green
full-oracle claim. Presence/brand repair selection passes its new controls but
that supplemental class selection remains red (combined 29.737s). The optional-
boolean landing hold remains until the user explicitly lifts it; property
presence and permitted undefined values do not establish storage compatibility.

Intersection counterfactuals complete: source dispatch/shape/nested (six backend
checks), selected-arm skip/shape/nested/tag (eight), recursive skip/shape/nested/
canonical/optional/three literals (sixteen), and original skip/shape/nested/
presence/outer/absence (twelve). Each driver shows valid true/exit-0 execution
losing the independent refusal pin, except forbidden optional absence produces
the intended counterfactual exit 70 on its positive control. Together with the
four new required-undefined presence checks, 46 newly executed implementation
counterfactuals bring the cumulative total to 258. Sanitizer/release repetitions
are retained in logs but not separately counted. All descriptor mutants are
confined to each test's IR; production sources are unmutated.

Final intersection required gates: lower 16.587s, native 52.707s, JavaScript 2.807s; full IR 30.949s; go vet ./... exit 0; uncached filtered Node oracle 376.970s, both original-declaration suites enabled, no fixture exclusion. Protected assembly/oracle files unchanged.
