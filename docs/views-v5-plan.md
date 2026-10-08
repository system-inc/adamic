# Views V5 preparation

Roadmap step 11, checked views, task #1af9j57. Delivery branch: compiler/views-v5.

## State and boundary

The required origin/compiler/views-v2 ref is absent. This document records preparation
on codex/views-v5-dry-run, based on 54cbc125422d4e1d64c1ffe782445b2cbc2bc5b8. No compiler admission, fixture,
count, refusal expectation or gap has changed. No delivery branch has been pushed.
The source pins are dictionaries 48d166cf, dictionary joint 9a917fc5 and
intersections 040999bc. Conflict rulings are read from c923d1cc.

The inventory below selects nonmerge commits touching the dictionaries or lane7
owned fixture tree, plus the eight joint commits after 1771221b. c41eaf85 is a
separate lane 4 dictionary-selector handoff found in the dictionary tip's ancestry;
it is not a joint branch's own commit. Its dictionary hunks must be reconciled with
V2 rather than importing its unrelated primitive certification changes. This is
an inventory for review, not a promise to take every path in each commit.
No integration merge is a proposed cherry-pick. Exclude 2724dabd, fresh_refused,
graph_regions, other slices' admissions and unrelated lane 4 fixture/report updates.
Historical logs and reports are evidence, not current-machine validation.

## Conflict probes

Each inventory commit was actually attempted with git cherry-pick --no-commit in
/tmp/views-v5-replay, starting from the identical current-main tree, then aborted
or reset in that scratch worktree. The delivery checkout was never used for a
compiler cherry-pick. There are 46 attempts: 42 with unresolved paths, three
textually clean, and one directory-rename error without an unresolved index entry.
The union of unresolved paths is 465. Full per-attempt logs are
/tmp/views-v5-pick-<sha>.log. Machine-readable inventory and observations are
/tmp/views-v5-inventory.json and /tmp/views-v5-replay.json.

These independent probes expose missing prerequisite files as modify/delete
conflicts and Git's directory-rename guesses as rename conflicts. They are not a
cumulative resolved replay, and the count does not predict V2's actual conflicts.
Resolving every missing-file conflict in favor of incoming code would recreate
other slices' implementation without their boundary controls. Repeat the ordered
replay on V2 when it exists; apply the ruled hunk decisions rather than either
whole side. An unresolved dependency stays named rather than merged from another
worker's unlanded tip.

## Representation and dependency plan

Dictionary reads reuse the existing counted record wrapper and ordered map, or
probe actual fixed-object/class slots. They preserve source heap kind, scalar kind,
readiness, present undefined versus absence and null, and named plus index contracts.
Receiver then key evaluates once. String-index signatures do not admit numeric,
symbol or template index domains by implication. Dictionary read certificates do
not authorize writes, spreads or record storage operations. Unknown provenance
keeps a check or the existing demanded refusal.

Apply dictionary descriptor/source hooks first (e9a06829, c9ad59e4), then generic
index propagation (5029b110), container controls, optional receiver distinction,
erased-source casts, finite keys, and enumeration. Check the ordinary record
prerequisite from records-lowering 456c981b: the dictionary lane merged it, so its
implementation is absent from this own-commit inventory. If V2 lacks those APIs,
name that dependency and separate the V5 read hooks from ordinary-record admission;
do not import the records lane merge or claim a working dictionary slice.

Object.keys reads actual own keys without demanding values. Values and entries
snapshot keys, check each selected value through dictionaryReadContract, retain
nested contracts, and use the existing tuple result representation. The private
ArrayPush.DictionaryProduction marker applies only to checked snapshots produced
inside this helper. DictionaryEntryOrigins must reach lazy demand without altering
the actual cast-origin ledger. Unsupported reached nested tuple consumers remain
refused. No broad array admission comes from these helpers.

The joint commits add readonly dictionary conversion and rich member selection.
8861b316 and 8f5ac7cb must check the real object heap kind when a selected reference
is consumed. Storage-dependent operations remain refused when allocation origins
cannot certify them. c77f7326 depends on the prior primitive selector handoff and
array/object descriptor machinery; retain demanded refusals for any family absent
from V2. dbeb7b5c adds original CompilerOptions, OptionsBase and BuildOptions controls.

Intersections use a conjunction of nonphantom runtime obligations, never union
selection. Checker-combined fields preserve duplicate-name child intersections.
Projected fields snapshot once and retain descendant obligations. Apply component,
lazy, finite-production and conservative-demand commits in order, then tagged arms,
recursive/ bounded walks, deferred descendants, ancestor absorption and synthetic
field descriptors. Only an independently proven phantom constituent erases.
Declared-ancestor absorption requires inheritance and actual-arm assignability.
Unknown, unavailable and unsupported child contracts keep their named read boundary.
A bounded contract walk is not graph ownership admission; cycles retain the base's
refusal and no graph_regions fixture is adopted.

There is no proof yet that V5 is inseparable from a different slice: the required
base is unavailable. There are concrete conditional dependencies: record lowering
456c981b; c77f7326's rich dictionary selection needs the existing array descriptors;
48d166cf's result containers need the existing tuple consumers or their refusal;
47c6c23a onward need V1's lazy field propagation and V2's union descriptors.
040999bc explicitly records unresolved original array rows 97923 and 98493. Retain
those boundaries. Never import the array worker merely to green the lane7 witnesses.
If a required V5 subset cannot remain sound without another slice's admission,
report its exact own commits and dependency, and stop that subset.

## Ruled conflict decisions

The following original approved hunk judgments touch paths in the inventory.
They are mandatory reconciliation constraints when V5 actually edits those files;
none has been applied to compiler code in this preparation-only checkout.

| File and original hunk | Area side | Views side | Approved resolution | Witness fixtures |
| --- | --- | --- | --- | --- |
| `internal/ir/ir.go` h3 (original line 223) | Typed arrays occupy representation tags 12, 13, 14 and are references. | Tags 12/13 describe null/undefined storage and Record is 14. | Keep views semantic slot tags and Record=14; move typed-array representations to distinct 15,16,17 and audit every numeric consumer. Retain both reference predicates. | `internal/oracle/testdata/typed_arrays_uint8array.a`<br>`internal/oracle/testdata/typed_arrays_float64array.a`<br>`stage3/interface-downcasts/dictionaries/source/values-fixed-object-good.a` |
| `internal/ir/ir.go` h4 (original line 265) | Typed arrays occupy representation tags 12, 13, 14 and are references. | Tags 12/13 describe null/undefined storage and Record is 14. | Keep views semantic slot tags and Record=14; move typed-array representations to distinct 15,16,17 and audit every numeric consumer. Retain both reference predicates. | `internal/oracle/testdata/typed_arrays_uint8array.a`<br>`internal/oracle/testdata/typed_arrays_float64array.a`<br>`stage3/interface-downcasts/dictionaries/source/values-fixed-object-good.a` |
| `internal/javascript/javascript.go` h1 (original line 54) | Typed counted/uncounted code pointers, exact optional/rest/count layouts and host dispatch. | Boxed signature adaptation, checked callable reads/results and owned result disposal. | Keep area typed pointer shapes and argument layouts; adapt existing views certificates/boxed values around that dispatch, comparing the correct code-pointer member and never casting a function pointer. Keep owned key/value holds and all callback refusals. | `internal/oracle/testdata/closure_convention_host24.a`<br>`internal/oracle/testdata/closure_convention_receiver_rest.a`<br>`internal/oracle/testdata/arguments_length_callbacks.a`<br>`stage3/interface-downcasts/lane5/ranked-callables/return-statement/good.a`<br>`stage3/interface-downcasts/nullish/maps/entry-live-mutation.a` |
| `internal/javascript/javascript.go` h2 (original line 200) | Reuse canonical interior environment cells and their readiness. | Emit readiness for noncaptured uninitialized parameters and skip interior capture allocation. | Retain both readiness cases and area interior-cell early return; no duplicate allocation or readiness variable. | `internal/oracle/testdata/non_null_uninitialized_capture.ts`<br>`internal/oracle/testdata/closure_convention_nested.a` |
| `internal/javascript/javascript.go` h3 (original line 445) | Reuse canonical interior environment cells and their readiness. | Emit readiness for noncaptured uninitialized parameters and skip interior capture allocation. | Retain both readiness cases and area interior-cell early return; no duplicate allocation or readiness variable. | `internal/oracle/testdata/non_null_uninitialized_capture.ts`<br>`internal/oracle/testdata/closure_convention_nested.a` |
| `internal/javascript/javascript.go` h4 (original line 1129) | Pass through proven Narrow values and emit Array.isArray. | Validate narrowed representation, undefined storage and tuple identity. | Keep views narrowing checks and the separate area ArrayIsArray case. | `stage3/interface-downcasts/tuples/emit-string-good.a`<br>`internal/oracle/testdata/regexp_null_narrowed.a` |
| `internal/javascript/javascript.go` h5 (original line 1279) | Typed counted/uncounted code pointers, exact optional/rest/count layouts and host dispatch. | Boxed signature adaptation, checked callable reads/results and owned result disposal. | Keep area typed pointer shapes and argument layouts; adapt existing views certificates/boxed values around that dispatch, comparing the correct code-pointer member and never casting a function pointer. Keep owned key/value holds and all callback refusals. | `internal/oracle/testdata/closure_convention_host24.a`<br>`internal/oracle/testdata/closure_convention_receiver_rest.a`<br>`internal/oracle/testdata/arguments_length_callbacks.a`<br>`stage3/interface-downcasts/lane5/ranked-callables/return-statement/good.a`<br>`stage3/interface-downcasts/nullish/maps/entry-live-mutation.a` |
| `internal/lower/cast.go` h1 (original line 33) | Dispatch interface view casts directly. | Dispatch classified deferred casts and preserve the deferred refusal when no plan applies. | Keep classified deferred dispatch and existing interface/class proofs without bypassing sameKeeping or a deferred refusal. | `stage3/interface-downcasts/lane1/unions-objects-good.a`<br>`stage3/interface-downcasts/lane1/unions-objects-wrong-payload.a` |
| `internal/lower/cast_proof.go` h1 (original line 101) | No special discarded-marker shortcut. | Recognize independently checked discarded callable markers. | Keep shortcut only under the existing validated marker contract; retain ordinary cast proofs and qualified-name guard. | `internal/oracle/testdata/census_never_rest_marker.a` |
| `internal/lower/collections.go` h1 (original line 355) | Allow known writable Union slots in destructuring. | Require primitive-union read proof and build the view contract. | Keep view-contract construction and both storage safety requirements; refuse any unsupported overlap rather than interpreting a slot by its apparent type. | `internal/oracle/testdata/tuple_values.a`<br>`stage3/interface-downcasts/tuples/emit-string-good.a` |
| `internal/lower/collections.go` h2 (original line 375) | Initialize through readiness-aware initializeLocal. | Prepare primitive and intersection reads before a plain Declare. | Keep both read preparations and feed their result through the area readiness-aware initializer. | `internal/oracle/testdata/switch_case_declarations/shared.a`<br>`stage3/interface-downcasts/tuples/emit-string-good.a` |
| `internal/lower/expression.go` h1 (original line 25) | Recognize exact plain objects, declared undefined field storage, void and observed array-predicate types. | Give never a storage representation whose evaluation traps. | Keep views never storage handling plus all area special cases and array-predicate observation. | `internal/oracle/testdata/never_reached.a`<br>`internal/oracle/testdata/taste_void.a` |
| `internal/lower/expression.go` h2 (original line 54) | Recognize fixed-width typed arrays. | Recognize checked and phantom array bases. | Keep the typed-array recognition and both distinct views array recognizers. | `internal/oracle/testdata/typed_arrays_uint8array.a`<br>`internal/oracle/testdata/phantom_array_brands.a` |
| `internal/lower/expression.go` h3 (original line 124) | Undefined/void share reference storage; most nullable references use NULL. | Null has distinct boxed identity; nullable views use Union, with specialized RegExp compatibility. | Keep area void support and views distinct null/undefined identities with the established RegExp exception; audit every nullable conversion. | `internal/oracle/testdata/typeof_null.a`<br>`internal/oracle/testdata/regexp_null_narrowed.a`<br>`internal/lower/testdata/optional_widening/storage_coherence/boolean-nullish.a` |
| `internal/lower/expression.go` h4 (original line 153) | Undefined/void share reference storage; most nullable references use NULL. | Null has distinct boxed identity; nullable views use Union, with specialized RegExp compatibility. | Keep area void support and views distinct null/undefined identities with the established RegExp exception; audit every nullable conversion. | `internal/oracle/testdata/typeof_null.a`<br>`internal/oracle/testdata/regexp_null_narrowed.a`<br>`internal/lower/testdata/optional_widening/storage_coherence/boolean-nullish.a` |
| `internal/lower/expression.go` h5 (original line 275) | Check contextual unknown views and preserve effects when adapting an undefined-only value. | Attach graph ownership, record predicate directions, and stop further conversion for never. | Preserve contextual checks and undefined effects, retain graph allocation and predicate bookkeeping, and retain never terminal behavior. | `internal/oracle/testdata/never_call.a`<br>`internal/oracle/testdata/graph_regions/classification_return.a`<br>`internal/oracle/testdata/undefined_references.a` |
| `internal/lower/expression.go` h6 (original line 551) | Dispatch typed arrays, namespaces, library method values and arguments.length. | Separate never trapping from unchecked value lowering and recognize process values. | Keep never trapping outside uncheckedValue; retain all area dispatch helpers and views process recognition within uncheckedValue. | `internal/oracle/testdata/never_call.a`<br>`internal/oracle/testdata/arguments_length.a`<br>`internal/oracle/testdata/process_shadow.a` |
| `internal/lower/expression.go` h7 (original line 678) | Mark synthetic checked union helpers for non-null operand recovery. | Mark tuple identity on the Narrow operation. | Keep both the checked-helper flag and tuple identity on that operation. | `internal/oracle/testdata/non_null_union.ts`<br>`stage3/interface-downcasts/tuples/emit-string-good.a` |
| `internal/lower/expression.go` h8 (original line 963) | Use raw checker nullability and distinguish unboxed nullable reference checks. | Use concrete generic nullability and distinct null/undefined identities. | Use concrete generic evidence without erasing an observable null or undefined test, keeping sentinel-aware behavior. | `internal/oracle/testdata/typeof_null.a`<br>`internal/oracle/testdata/generic_values.a` |
| `internal/lower/expression.go` h9 (original line 996) | Use raw checker nullability and distinguish unboxed nullable reference checks. | Use concrete generic nullability and distinct null/undefined identities. | Use concrete generic evidence without erasing an observable null or undefined test, keeping sentinel-aware behavior. | `internal/oracle/testdata/typeof_null.a`<br>`internal/oracle/testdata/generic_values.a` |
| `internal/lower/expression.go` h10 (original line 1368) | Forward named function rest slots and arguments counts. | Forward callable producer masks. | Keep rest/count forwarding and producer masks on the same forwarder. | `internal/oracle/testdata/arguments_length_value_count.a`<br>`internal/oracle/testdata/census_never_rest_marker.a` |
| `internal/lower/expression.go` h11 (original line 1455) | Decide result storage from concrete generic type. | Decide from raw checker type. | Retain the area concrete result decision; keep views checked result adaptation around the resulting representation. | `internal/oracle/testdata/generic_values.a` |
| `internal/lower/generic.go` h1 (original line 115) | Preserve nested substitution, closure stack and closure registration; restore context after depth increment. | Restore caller context also on earlier validation exits. | Keep nested context and registration, use the earlier context restoration defer, and decrement generic depth only after increment. | `internal/oracle/testdata/nested_generic_capture.a`<br>`stage3/drivers/scanner/probes/nested-generic-sibling-call.a` |
| `internal/lower/generic.go` h2 (original line 151) | Preserve nested substitution, closure stack and closure registration; restore context after depth increment. | Restore caller context also on earlier validation exits. | Keep nested context and registration, use the earlier context restoration defer, and decrement generic depth only after increment. | `internal/oracle/testdata/nested_generic_capture.a`<br>`stage3/drivers/scanner/probes/nested-generic-sibling-call.a` |
| `internal/lower/library_object.go` h1 (original line 12) | Route through explicit objectCallArguments for spread-aware calls. | Recognize dictionary values/keys, records and checked array-record production. | Keep the explicit-argument wrapper and views recognizers, requiring a supported written-argument plan before using a recognizer. | `stage3/interface-downcasts/dictionaries/source/values-fixed-object-good.a`<br>`internal/oracle/testdata/library_object_keys.a` |
| `internal/lower/library_string.go` h1 (original line 74) | Evaluate void/undefined effects before spelling undefined. | Spell phantom members using their certified primitive value. | Keep both distinct spelling branches with effects evaluated once. | `internal/oracle/testdata/taste_void.a`<br>`internal/oracle/testdata/phantom_brands.a` |
| `internal/lower/library_string.go` h2 (original line 92) | Allow writable types or proven dynamic scalar property. | Use the non-nullable type or proven primitive dictionary value. | Retain each existing certified scalar path and refuse unsupported reference members; preserve nullable spelling. | `internal/oracle/testdata/records_scalars.a`<br>`internal/oracle/testdata/library_string_conversion.a` |
| `internal/lower/object.go` h1 (original line 338) | Reject erased unknown array storage and allow packed Union elements. | Require primitive/nominal contracts or tuple-scalar union plans for slotless elements. | Keep erased-unknown refusal and views certified element plans; preserve Union storage only where the contract proves its representation. | `stage3/interface-downcasts/tuples/emit-string-good.a`<br>`internal/oracle/testdata/typed_arrays_uint8array.a` |
| `internal/lower/object.go` h2 (original line 1318) | Record callback checker type for typed calling conventions. | Record checked element read plan. | Retain CallbackType and the original ViewRead on the same ArrayMap/ArrayVisit/ArrayReduce constructor. | `internal/oracle/testdata/arguments_length_callbacks.a`<br>`internal/oracle/testdata/library_array_holes_callbacks.a` |
| `internal/lower/object.go` h3 (original line 1514) | Record callback checker type for typed calling conventions. | Record checked element read plan. | Retain CallbackType and the original ViewRead on the same ArrayMap/ArrayVisit/ArrayReduce constructor. | `internal/oracle/testdata/arguments_length_callbacks.a`<br>`internal/oracle/testdata/library_array_holes_callbacks.a` |
| `internal/lower/object.go` h4 (original line 1550) | Record callback checker type for typed calling conventions. | Record checked element read plan. | Retain CallbackType and the original ViewRead on the same ArrayMap/ArrayVisit/ArrayReduce constructor. | `internal/oracle/testdata/arguments_length_callbacks.a`<br>`internal/oracle/testdata/library_array_holes_callbacks.a` |
| `internal/lower/object.go` h5 (original line 1577) | Reject unsupported slotless Map values unless writable Union. | Reject only unknown representation at this point. | Retain the area unsupported-slot refusal unless an existing Map storage certificate proves the value; preserve all 47 callback refusal pins. | `stage3/interface-downcasts/nullish/maps/entry-live-mutation.a`<br>`internal/oracle/testdata/map_foreach_keys.a` |
| `internal/lower/object.go` h6 (original line 1866) | Record comparator callback type. | Validate and record optional comparator behavior. | Keep callback type and the existing optional comparator proof on the same sort constructor. | `internal/oracle/testdata/arguments_length_callbacks.a`<br>`internal/oracle/testdata/library_array_holes_callbacks.a` |
| `internal/lower/refusals.go` h1 (original line 27) | Refuse index signatures; namespaces and void have proved area paths. | Refuse namespaces and void; dictionaries admit supported index signatures. | Keep proved area namespace/void paths, retain refusal of unsupported index signatures, and admit only the existing dictionary contract slice after ruling. | `internal/oracle/testdata/params_namespaces_order.a`<br>`internal/oracle/testdata/taste_void.a`<br>`stage3/interface-downcasts/dictionaries/source/values-fixed-object-good.a` |
| `internal/lower/refusals.go` h2 (original line 89) | Predicate and Node library refusal checks. | Dictionary, phantom, detached-own and record refusal checks, then predicate check. | Retain every existing guard and diagnostic precedence; avoid duplicate Node-library checks already present after the hunk. | `internal/oracle/testdata/records_prototype_set.a`<br>`internal/oracle/testdata/phantom_brands.a` |
| `internal/lower/refusals.go` h3 (original line 218) | Refuse shorthand escapes of library aliases and permit certified library method reads. | Permit certified process observations and detached hasOwn methods. | Retain shorthand escape refusal and each independently certified method-read exception; do not broaden unbound-method admission. | `internal/oracle/testdata/detached_own_records.a`<br>`internal/oracle/testdata/process_observations.a`<br>`internal/oracle/testdata/library_function_expressions.a` |
| `internal/native/emit_expressions.go` h1 (original line 397) | Typed counted/uncounted code pointers, exact optional/rest/count layouts and host dispatch. | Boxed signature adaptation, checked callable reads/results and owned result disposal. | Keep area typed pointer shapes and argument layouts; adapt existing views certificates/boxed values around that dispatch, comparing the correct code-pointer member and never casting a function pointer. Keep owned key/value holds and all callback refusals. | `internal/oracle/testdata/closure_convention_host24.a`<br>`internal/oracle/testdata/closure_convention_receiver_rest.a`<br>`internal/oracle/testdata/arguments_length_callbacks.a`<br>`stage3/interface-downcasts/lane5/ranked-callables/return-statement/good.a`<br>`stage3/interface-downcasts/nullish/maps/entry-live-mutation.a` |
| `internal/native/emit_expressions.go` h2 (original line 537) | Typed counted/uncounted code pointers, exact optional/rest/count layouts and host dispatch. | Boxed signature adaptation, checked callable reads/results and owned result disposal. | Keep area typed pointer shapes and argument layouts; adapt existing views certificates/boxed values around that dispatch, comparing the correct code-pointer member and never casting a function pointer. Keep owned key/value holds and all callback refusals. | `internal/oracle/testdata/closure_convention_host24.a`<br>`internal/oracle/testdata/closure_convention_receiver_rest.a`<br>`internal/oracle/testdata/arguments_length_callbacks.a`<br>`stage3/interface-downcasts/lane5/ranked-callables/return-statement/good.a`<br>`stage3/interface-downcasts/nullish/maps/entry-live-mutation.a` |
| `internal/native/runtime/adamic.h` h1 (original line 35) | Close conditional environment-kind enum block. | Add distinct null kind. | Keep conditional environment kind and one unconditional distinct null kind, paired with a single views null-sentinel definition. | `internal/oracle/testdata/typeof_null.a`<br>`internal/oracle/testdata/regexp_null_narrowed.a` |
| `internal/native/runtime/adamic.h` h2 (original line 97) | Cell owner pointer exists only with canonical closures. | Cell owner pointer always exists for graph ownership. | Keep owner storage where either canonical or graph ownership requires it; preserve ordinary layout when neither requires it if the graph ABI permits. | `internal/oracle/testdata/closure_convention_nested.a`<br>`internal/oracle/testdata/graph_regions_generic_capture.a` |
| `internal/native/runtime/adamic.h` h3 (original line 325) | Typed counted/uncounted code pointers, exact optional/rest/count layouts and host dispatch. | Boxed signature adaptation, checked callable reads/results and owned result disposal. | Keep area typed pointer shapes and argument layouts; adapt existing views certificates/boxed values around that dispatch, comparing the correct code-pointer member and never casting a function pointer. Keep owned key/value holds and all callback refusals. | `internal/oracle/testdata/closure_convention_host24.a`<br>`internal/oracle/testdata/closure_convention_receiver_rest.a`<br>`internal/oracle/testdata/arguments_length_callbacks.a`<br>`stage3/interface-downcasts/lane5/ranked-callables/return-statement/good.a`<br>`stage3/interface-downcasts/nullish/maps/entry-live-mutation.a` |

Later numbered rulings supersede earlier pending descriptions. Apply these by
name where their file is touched, preserving the base's other admissions:

- 1 and the cycles.go judgment: retain the area cycle refusal; 2724dabd and the
  declined 86 cycle-expectation proposals are excluded.
- 7: .a postfix assertions stay refused; do not migrate new V5 sources to .ts.
- 8: only independently producer-certified callable boxing can escape the unknown
  function guard; V5 supplies no new callable certificate.
- 9: the dictionary cast must check Union-to-Object heap identity and preserve the
  selected Object/Record target. A structural {} type is not a source heap proof.
- 10 and 11: preserve stored-marker/result-erasure and plain Error optional-host
  field refusals. Dictionary/intersection dispatch must not bypass them.
- 12: preserve overloaded shorthand implementation-identity refusal.
- 13: finite computed-key selection requires an independently registered receiver
  view contract; ordinary unviewed objects retain their boundary.
- 14: nullable-reference-to-unknown requires the non-null source representation;
  retain the refusal if the proof is unavailable.
- 15: callable masks normalize only existing phantomBase/phantomArrayBase proofs.
- 16: preserve the Union field guard and independently certified length:3 control.
- 18 and 20: retain producer-sized typed callable adapters and zero void certificate
  masks if the base provides them; V5 does not add callable admission.
- 21: existing freshness follows contained references without exempting cycles.
- 23: ordinary boxed Union-array storage survives only without erasing a registered
  checked-view contract; dictionary result construction is a narrower private case.
- 24: collections.go keeps the ordinary boxed Union Map.forEach adapter only
  without erasing checked-view value contracts. Preserve entry-live-mutation.a.
- 26: records.go keeps fixed-object Object.values, entries and keys dispatch without
  contextual conversion into Record. Ordinary storage conversions remain refused.

Rulings 17, 19, 22, 25 and 27 concern files outside these lane commits' changed
paths; preserve the base implementations. If reconciliation requires editing such
files, add their named rulings and independent witnesses before changing them.
Protected orchestration files remain untouched. The lane production hooks explicitly
avoid lower.go, native/emit.go, native/native.go and oracle/oracle_test.go.

## Required real-build verification

When the remote base appears, fetch it and create compiler/views-v5 from its exact
SHA. Preserve this plan, take only reviewed V5 own commits and scoped handoffs,
and record actual taken SHAs separately from this inventory. First collect same-
machine base observations and then replay exactly the same commands on V5:

```sh
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
go test -json -count=1 -timeout 30m ./internal/lower ./internal/ir ./internal/flow ./internal/javascript ./internal/oracle > /tmp/views-v5-base-packages.jsonl 2>&1
(cd stage3/api && npm ci) > /tmp/views-v5-base-npm.log 2>&1
go test -json -count=1 -timeout 30m ./stage3/fixtures > /tmp/views-v5-base-stage3.jsonl 2>&1
go test -json -count=1 -timeout 30m ./stage1/... -run 'Gap|Gaps|Probes' > /tmp/views-v5-base-gaps.jsonl 2>&1
```

Repeat with V5 log names. Record each leaf transition; every moved existing result
needs its named Moved-result trailer. New fixtures must stay in V5 scope. Review
counts.md against the saved base, regenerate TestCountsAreRecorded with
-update-counts, and accept only V5 rows. Move actually closed stage1 gap sources to
the package's existing closed form and update its GAPS.md. Run stage3 once more
with only its host platform guard lifted locally on each tree; restore and verify
the guard bytes before committing.

cloud/fast-gate/run.py is absent both from current main and c923d1cc, so its exact
aCheck implementation cannot be inspected on these trees. Read it from V2 when
available. Audit every added/changed external-package .a: successful checker or
first-line exact expected refusal/type-error plus an asserting test. Original-tsc
imports require the lane's pinned declaration adapters, not copied cohere files.
Do not call declaration-resolution failure a successful refusal audit.

For each new check, run a compiling removal mutant caught by its intended fixture:
source heap kind, scalar/reference storage, presence, readiness, allowed literal,
finite key, optional receiver, descendant field, union arm/conjunction, bounded
walk, ancestor proof, synthetic-field binding, record-operation conversion guard,
actual keys, per-enumerated-value selection, private result ownership and derived
entry origin demand. Operand evaluation must have a side-effect witness. Use
Node source first; compare both backends, sanitized native and finishing leak
checks. Existing lane-mutant reports are recipes, not this machine's proof.

No new runtime check has been installed here, so no removal mutant or compiler
package comparison is claimed. No Node language ruling is invented from a text
conflict. Required inspections remain conditional on the actual V2 compiler.
Unresolved array/tuple consumers remain refused rather than adding another slice.

## Preparation commands and outcomes

Fetch of main and the four named lane refs succeeded. Remote checks for
refs/heads/compiler/views-v2 returned no ref on each preparation pass. One check
without network escalation failed to connect to the sandbox proxy; the authorized
network retry succeeded and still returned no ref.

Toolchain setup command: export GOPROXY='https://proxy.golang.org|direct';
bash cloud/setup.sh > /tmp/views-v5-setup.log 2>&1. Exit 0. Timing lines:
Node 0.053s, Go 0.073s, markdown dependencies 0.128s, submodules 0.139s,
clang 0.384s, Go build 81.581s, deferred test binaries 81.749s,
cache warm 81.752s, done 81.800s. nproc: 5; cgroup quota: 4 CPUs.
The printed environment is /workspace/adamic-tools/env.sh.

No implementation test, count refresh, stage3 run, stage1 run or mutant has been
run for a nonexistent V2 comparison. This preparatory document adds no .a files.
No compiler/views-v5 push is authorized as a finished green unit yet.

## Own-commit inventory and exact probe conflicts

Changes list production/test Go, C and header paths, plus a count of the remaining
fixture, report and evidence files. All unresolved paths are named, including Git
rename guesses; inspect the raw log before treating a rename guess as a source
conflict. The lane-order lists are chronological within each lane. Joint follows
dictionary descriptor availability; intersection hooks precede other member
dispatch. Reconcile shared hunks individually on the actual base.

### e88908f4: Rank dictionary demand and record the source admission frontier

Lane: dictionaries. Own nonmerge commit: `e88908f4094d28b8a4a4a4409763ff410adbcfe2`.

Changes: fixture/report/evidence only. 12 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`

### 753cca96: Preserve dictionary checkpoint setup and observation logs

Lane: dictionaries. Own nonmerge commit: `753cca960f3abeef5184032e13acb0f1ce2ec2fb`.

Changes: fixture/report/evidence only. 18 other paths.

Cherry-pick exit 0: textual application only, not a compile or execution result.

### e9a06829: Prepare dictionary read contracts on the integrated lazy baseline

Lane: dictionaries. Own nonmerge commit: `e9a0682932f89442a4c06ff36a593ed019cfc8d0`.

Changes: `internal/javascript/view_dictionaries.go`, `internal/lower/view_dictionaries.go`, `internal/lower/view_dictionaries_test.go`, `internal/native/runtime/view_dictionaries.c`, `internal/native/runtime/view_dictionaries.h`, `internal/oracle/checked_views_dictionaries_test.go`. 24 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-compiler-build.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-components.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-diff-check.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-frontier.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-guards.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-initial-components.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-javascript-package.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-node-dependencies.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-regression-initial.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-regression.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group1-vet.log`
- `stage3/drivers/parser/evidence/front15/new-probes/integration-setup.log`
- `stage3/drivers/parser/evidence/front15/new-probes/lazy-ranking-check.log`
- `stage3/drivers/parser/evidence/front15/new-probes/lazy-ranking.log`

### c9ad59e4: Wire checked dictionary source reads and preserve transitive guards

Lane: dictionaries. Own nonmerge commit: `c9ad59e4dd76497571be6065bd6d449d757324b2`.

Changes: `internal/flow/infer.go`, `internal/ir/ir.go`, `internal/ir/view_dictionaries.go`, `internal/ir/views.go`, `internal/javascript/javascript.go`, `internal/javascript/view_dictionaries.go`, `internal/javascript/view_dictionary_dispatch.go`, `internal/lower/object.go`, `internal/lower/readiness.go`, `internal/lower/refusals.go`, `internal/lower/shape_conformance.go`, `internal/lower/view_dictionaries.go`, `internal/lower/view_dictionaries_test.go`, `internal/lower/view_lazy.go`, `internal/lower/view_writes.go`, `internal/native/emit_expressions.go`, `internal/native/runtime/adamic.h`, `internal/native/runtime/record.c`, `internal/native/runtime/view_dictionaries.c`, `internal/native/runtime/view_dictionaries.h`, `internal/native/view_dictionaries.go`, `internal/oracle/checked_view_dictionary_source_test.go`, `internal/oracle/checked_views_dictionaries_test.go`. 12 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/ir/views.go`
- `internal/javascript/javascript.go`
- `internal/javascript/view_dictionaries.go`
- `internal/lower/readiness.go`
- `internal/lower/refusals.go`
- `internal/lower/shape_conformance.go`
- `internal/lower/view_dictionaries.go`
- `internal/lower/view_dictionaries_test.go`
- `internal/lower/view_lazy.go`
- `internal/lower/view_writes.go`
- `internal/native/runtime/view_dictionaries.c`
- `internal/native/runtime/view_dictionaries.h`
- `internal/oracle/checked_views_dictionaries_test.go`
- `stage3/a-check-headers/GROUP2.md`
- `stage3/a-check-headers/components/array-read-wrong.a`
- `stage3/a-check-headers/components/options-read-wrong.a`
- `stage3/a-check-headers/logs/group2-components.log`
- `stage3/a-check-headers/logs/group2-ranking.log`
- `stage3/a-check-headers/logs/group2-regression-initial.log`
- `stage3/a-check-headers/logs/group2-regression.log`
- `stage3/a-check-headers/logs/group2-source-mutants-initial.log`
- `stage3/a-check-headers/logs/group2-source-mutants.log`
- `stage3/a-check-headers/logs/group2-vet.log`
- `stage3/a-check-headers/logs/group2-whitespace.log`

### 5029b110: Check dictionary array elements and generic index arguments

Lane: dictionaries. Own nonmerge commit: `5029b110e5ce468ad90946745174e77394914dd2`.

Changes: `internal/ir/view_dictionaries.go`, `internal/lower/generic.go`, `internal/lower/refusals.go`, `internal/lower/view_dictionaries.go`, `internal/lower/view_objects.go`, `internal/native/view_dictionaries.go`, `internal/oracle/checked_view_dictionary_propagation_test.go`. 21 other paths.

Cherry-pick exit 1; unresolved paths:

- `internal/ir/view_dictionaries.go`
- `internal/lower/refusals.go`
- `internal/lower/view_dictionaries.go`
- `internal/lower/view_objects.go`
- `internal/native/view_dictionaries.go`
- `stage3/a-check-headers/completed-read-pairs.json`
- `stage3/a-check-headers/source/flow-callback-good.a`
- `stage3/a-check-headers/source/flow-callback-wrong.a`
- `stage3/a-check-headers/source/flow-field-good.a`
- `stage3/a-check-headers/source/flow-field-wrong.a`
- `stage3/a-check-headers/source/flow-generic-good.a`
- `stage3/a-check-headers/source/flow-generic-wrong.a`
- `stage3/a-check-headers/source/flow-helper-good.a`
- `stage3/a-check-headers/source/flow-helper-wrong.a`
- `stage3/a-check-headers/source/paths-container-wrong.a`
- `stage3/a-check-headers/source/paths-element-wrong.a`
- `stage3/a-check-headers/source/paths-good.a`
- `stage3/a-check-headers/source/paths-key-missing.a`
- `stage3/a-check-headers/source/paths-missing.a`
- `stage3/a-check-headers/source/paths-wrong.a`
- `stage3/a-check-headers/source/root-good.a`
- `stage3/a-check-headers/source/root-wrong.a`
- `stage3/a-check-headers/source/write-guard.a`
- `stage3/interface-downcasts/dictionaries/candidate-pair-progress.json`
- `stage3/interface-downcasts/dictionaries/candidate-summary.json`
- `stage3/interface-downcasts/dictionaries/rank-lazy-demand.py`

### 26e4694d: Record dictionary storage merge evidence and preserve raw logs

Lane: dictionaries. Own nonmerge commit: `26e4694ddddc3c345f3808f1c71ab198feb3dd69`.

Changes: fixture/report/evidence only. 2 other paths.

Cherry-pick exit 1; unresolved paths:

- `stage3/interface-downcasts/dictionaries/GROUP3.md`

### ce4eeaa4: Certify lazy dictionary container reads before rich element unions

Lane: dictionaries. Own nonmerge commit: `ce4eeaa45750a49d1fd44f0881473513d3d82eb8`.

Changes: `internal/oracle/checked_view_dictionary_container_test.go`. 9 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `stage3/a-check-headers/GROUP4.md`
- `stage3/a-check-headers/source/command-options-good.a`
- `stage3/a-check-headers/source/command-options-missing.a`
- `stage3/a-check-headers/source/command-options-wrong.a`
- `stage3/interface-downcasts/dictionaries/candidate-pair-progress.json`
- `stage3/interface-downcasts/dictionaries/candidate-summary.json`
- `stage3/interface-downcasts/dictionaries/completed-read-pairs.json`
- `stage3/interface-downcasts/dictionaries/logs/.gitattributes`

### dda25682: Certify remaining direct dictionary container shapes and mixed paths

Lane: dictionaries. Own nonmerge commit: `dda25682419c46983dc8531984985dca3d52f6da`.

Changes: `internal/oracle/checked_view_dictionary_container_test.go`, `internal/oracle/checked_view_dictionary_propagation_test.go`. 36 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/oracle/checked_view_dictionary_container_test.go`
- `internal/oracle/checked_view_dictionary_propagation_test.go`
- `stage3/a-check-headers/GROUP5.md`
- `stage3/a-check-headers/source/builder-state-options-good.a`
- `stage3/a-check-headers/source/builder-state-options-missing.a`
- `stage3/a-check-headers/source/builder-state-options-wrong.a`
- `stage3/a-check-headers/source/compiler-paths-container-wrong.a`
- `stage3/a-check-headers/source/compiler-paths-element-wrong.a`
- `stage3/a-check-headers/source/compiler-paths-good.a`
- `stage3/a-check-headers/source/compiler-paths-key-missing.a`
- `stage3/a-check-headers/source/compiler-paths-missing.a`
- `stage3/a-check-headers/source/compiler-paths-wrong.a`
- `stage3/a-check-headers/source/incremental-bundle-options-good.a`
- `stage3/a-check-headers/source/incremental-bundle-options-missing.a`
- `stage3/a-check-headers/source/incremental-bundle-options-wrong.a`
- `stage3/a-check-headers/source/incremental-multi-options-good.a`
- `stage3/a-check-headers/source/incremental-multi-options-missing.a`
- `stage3/a-check-headers/source/incremental-multi-options-wrong.a`
- `stage3/a-check-headers/source/incremental-options-good.a`
- `stage3/a-check-headers/source/incremental-options-missing.a`
- `stage3/a-check-headers/source/incremental-options-wrong.a`
- `stage3/a-check-headers/source/reusable-state-options-good.a`
- `stage3/a-check-headers/source/reusable-state-options-missing.a`
- `stage3/a-check-headers/source/reusable-state-options-wrong.a`
- `stage3/a-check-headers/source/version-paths-good.a`
- `stage3/a-check-headers/source/version-paths-missing.a`
- `stage3/a-check-headers/source/version-paths-wrong.a`
- `stage3/a-check-headers/source/watch-options-good.a`
- `stage3/a-check-headers/source/watch-options-missing.a`
- `stage3/a-check-headers/source/watch-options-wrong.a`
- `stage3/a-check-headers/source/wildcard-directories-good.a`
- `stage3/a-check-headers/source/wildcard-directories-missing.a`
- `stage3/a-check-headers/source/wildcard-directories-wrong.a`
- `stage3/interface-downcasts/dictionaries/candidate-pair-progress.json`
- `stage3/interface-downcasts/dictionaries/candidate-summary.json`
- `stage3/interface-downcasts/dictionaries/completed-read-pairs.json`
- `stage3/interface-downcasts/dictionaries/logs/.gitattributes`

### 07603363: Preserve raw dictionary container oracle and mutant evidence

Lane: dictionaries. Own nonmerge commit: `076033637068ea64be78fc47f467c896255f6dec`.

Changes: fixture/report/evidence only. 5 other paths.

Cherry-pick exit 1: ambiguous directory rename split of dictionaries/logs; no unmerged index path.

### 8a660bcd: Check required dictionary fields through optional receivers

Lane: dictionaries. Own nonmerge commit: `8a660bcd50fa547e64dc148c13c3af334d1e834a`.

Changes: `internal/javascript/readiness.go`, `internal/native/runtime/object.c`, `internal/oracle/checked_view_dictionary_optional_test.go`. 24 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/javascript/readiness.go`
- `internal/native/runtime/object.c`
- `stage3/a-check-headers/GROUP6.md`
- `stage3/a-check-headers/logs/group6-final-gate.log`
- `stage3/a-check-headers/logs/group6-initial-cast.log`
- `stage3/a-check-headers/logs/group6-initial-presence.log`
- `stage3/a-check-headers/source/optional-options-good.a`
- `stage3/a-check-headers/source/optional-options-missing.a`
- `stage3/a-check-headers/source/optional-options-receiver-missing.a`
- `stage3/a-check-headers/source/optional-options-undefined.a`
- `stage3/a-check-headers/source/optional-options-wrong.a`
- `stage3/a-check-headers/source/optional-watch-good.a`
- `stage3/a-check-headers/source/optional-watch-missing.a`
- `stage3/a-check-headers/source/optional-watch-receiver-missing.a`
- `stage3/a-check-headers/source/optional-watch-undefined.a`
- `stage3/a-check-headers/source/optional-watch-wrong.a`
- `stage3/a-check-headers/source/optional-wildcard-good.a`
- `stage3/a-check-headers/source/optional-wildcard-missing.a`
- `stage3/a-check-headers/source/optional-wildcard-receiver-missing.a`
- `stage3/a-check-headers/source/optional-wildcard-undefined.a`
- `stage3/a-check-headers/source/optional-wildcard-wrong.a`
- `stage3/interface-downcasts/dictionaries/candidate-pair-progress.json`
- `stage3/interface-downcasts/dictionaries/candidate-summary.json`
- `stage3/interface-downcasts/dictionaries/completed-read-pairs.json`
- `stage3/interface-downcasts/dictionaries/logs/.gitattributes`

### 45b334d0: Check erased dictionary sources and generic lookup descriptors

Lane: dictionaries. Own nonmerge commit: `45b334d08dbe6fa6355d054f1f82b95ace40d847`.

Changes: `internal/lower/cast_proof.go`, `internal/lower/readiness.go`, `internal/lower/view_dictionaries.go`, `internal/oracle/checked_view_dictionary_lookup_test.go`. 31 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/lower/view_dictionaries.go`
- `stage3/a-check-headers/GROUP7.md`
- `stage3/a-check-headers/logs/group7-gate.log`
- `stage3/a-check-headers/logs/group7-initial-admission.log`
- `stage3/a-check-headers/logs/group7-initial-producers.log`
- `stage3/a-check-headers/logs/group7-initial-readiness.log`
- `stage3/a-check-headers/logs/group7-vet.log`
- `stage3/a-check-headers/source/map-generic-good.a`
- `stage3/a-check-headers/source/map-generic-missing.a`
- `stage3/a-check-headers/source/map-generic-producer-good.a`
- `stage3/a-check-headers/source/map-generic-producer-missing.a`
- `stage3/a-check-headers/source/map-generic-producer-wrong.a`
- `stage3/a-check-headers/source/map-generic-wrong.a`
- `stage3/a-check-headers/source/map-string-good.a`
- `stage3/a-check-headers/source/map-string-missing.a`
- `stage3/a-check-headers/source/map-string-producer-good.a`
- `stage3/a-check-headers/source/map-string-producer-missing.a`
- `stage3/a-check-headers/source/map-string-producer-wrong.a`
- `stage3/a-check-headers/source/map-string-wrong.a`
- `stage3/a-check-headers/source/package-paths-good.a`
- `stage3/a-check-headers/source/package-paths-missing.a`
- `stage3/a-check-headers/source/package-paths-nested-wrong.a`
- `stage3/a-check-headers/source/package-paths-producer-good.a`
- `stage3/a-check-headers/source/package-paths-producer-missing.a`
- `stage3/a-check-headers/source/package-paths-producer-nested-wrong.a`
- `stage3/a-check-headers/source/package-paths-producer-wrong.a`
- `stage3/a-check-headers/source/package-paths-wrong.a`
- `stage3/interface-downcasts/dictionaries/candidate-pair-progress.json`
- `stage3/interface-downcasts/dictionaries/candidate-summary.json`
- `stage3/interface-downcasts/dictionaries/completed-read-pairs.json`
- `stage3/interface-downcasts/dictionaries/logs/.gitattributes`

### deec3c93: Check finite dictionary keys and optional named fields

Lane: dictionaries. Own nonmerge commit: `deec3c933543c7b5e7ba0d9db0964265d03c494c`.

Changes: `internal/lower/object.go`, `internal/lower/view_dictionaries.go`, `internal/lower/view_dictionary_lookup.go`, `internal/oracle/checked_view_dictionary_finite_test.go`, `internal/oracle/checked_view_dictionary_propagation_test.go`. 49 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/lower/object.go`
- `internal/lower/view_dictionaries.go`
- `internal/oracle/checked_view_dictionary_propagation_test.go`
- `stage3/a-check-headers/GROUP8.md`
- `stage3/a-check-headers/logs/group8-extra.log`
- `stage3/a-check-headers/logs/group8-gate.log`
- `stage3/a-check-headers/logs/group8-mutants.log`
- `stage3/a-check-headers/logs/group8-vet.log`
- `stage3/a-check-headers/source/enum-map-good.a`
- `stage3/a-check-headers/source/enum-map-missing.a`
- `stage3/a-check-headers/source/enum-map-open-number.a`
- `stage3/a-check-headers/source/enum-map-producer-good.a`
- `stage3/a-check-headers/source/enum-map-producer-missing.a`
- `stage3/a-check-headers/source/enum-map-producer-open-number.a`
- `stage3/a-check-headers/source/enum-map-producer-wrong.a`
- `stage3/a-check-headers/source/enum-map-wrong.a`
- `stage3/a-check-headers/source/indexed-cache-good.a`
- `stage3/a-check-headers/source/indexed-cache-missing.a`
- `stage3/a-check-headers/source/indexed-cache-nested-wrong.a`
- `stage3/a-check-headers/source/indexed-cache-wrong-array.a`
- `stage3/a-check-headers/source/indexed-cache-wrong.a`
- `stage3/a-check-headers/source/iterable-cache-good.a`
- `stage3/a-check-headers/source/iterable-cache-missing.a`
- `stage3/a-check-headers/source/iterable-cache-nested-wrong.a`
- `stage3/a-check-headers/source/iterable-cache-wrong-array.a`
- `stage3/a-check-headers/source/iterable-cache-wrong.a`
- `stage3/a-check-headers/source/iteration-cache-good.a`
- `stage3/a-check-headers/source/iteration-cache-missing.a`
- `stage3/a-check-headers/source/iteration-cache-nested-wrong.a`
- `stage3/a-check-headers/source/iteration-cache-wrong-array.a`
- `stage3/a-check-headers/source/iteration-cache-wrong.a`
- `stage3/a-check-headers/source/optional-paths-container-wrong.a`
- `stage3/a-check-headers/source/optional-paths-element-wrong.a`
- `stage3/a-check-headers/source/optional-paths-good.a`
- `stage3/a-check-headers/source/optional-paths-key-missing.a`
- `stage3/a-check-headers/source/optional-paths-missing.a`
- `stage3/a-check-headers/source/optional-paths-once.a`
- `stage3/a-check-headers/source/optional-paths-receiver-missing.a`
- `stage3/a-check-headers/source/optional-paths-wrong.a`
- `stage3/a-check-headers/source/optional-required-paths-good.a`
- `stage3/a-check-headers/source/optional-required-paths-missing.a`
- `stage3/a-check-headers/source/optional-required-paths-receiver-missing.a`
- `stage3/a-check-headers/source/optional-required-paths-undefined.a`
- `stage3/a-check-headers/source/signature-cache-good.a`
- `stage3/a-check-headers/source/signature-cache-missing.a`
- `stage3/a-check-headers/source/signature-cache-nested-wrong.a`
- `stage3/a-check-headers/source/signature-cache-wrong-array.a`
- `stage3/a-check-headers/source/signature-cache-wrong.a`
- `stage3/interface-downcasts/dictionaries/candidate-pair-progress.json`
- `stage3/interface-downcasts/dictionaries/candidate-summary.json`
- `stage3/interface-downcasts/dictionaries/completed-read-pairs.json`

### 04de3f66: Certify string and Path indexing on checked views integration

Lane: dictionaries. Own nonmerge commit: `04de3f6667a002682ed2cdf7a347dcae0d256f43`.

Changes: `internal/oracle/checked_view_dictionary_recheck_test.go`. 18 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `stage3/drivers/parser/evidence/front15/new-probes/group9-integration-dictionaries.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group9-string-final.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group9-vet.log`
- `stage3/interface-downcasts/dictionaries/candidate-pair-progress.json`
- `stage3/interface-downcasts/dictionaries/candidate-summary.json`
- `stage3/interface-downcasts/dictionaries/completed-read-pairs.json`

### 7a3ea2a8: Enumerate actual dictionary keys without demanding values

Lane: dictionaries. Own nonmerge commit: `7a3ea2a8155cfad376e55f9bd4114e8a8b16d9d8`.

Changes: `internal/ir/records.go`, `internal/lower/library_object.go`, `internal/lower/view_dictionary_enumeration.go`, `internal/native/emit_records.go`, `internal/native/runtime/view_dictionaries.c`, `internal/native/runtime/view_dictionaries.h`, `internal/oracle/checked_view_dictionary_enumeration_test.go`. 33 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/ir/records.go`
- `internal/lower/library_object.go`
- `internal/native/emit_records.go`
- `internal/native/runtime/view_dictionaries.c`
- `internal/native/runtime/view_dictionaries.h`
- `stage3/drivers/parser/evidence/front15/new-probes/group10-blocked-js-final.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group10-gate.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group10-keys.log`
- `stage3/drivers/parser/evidence/front15/new-probes/group10-vet.log`

### 48d166cf: Check dictionary value and entry snapshots lazily

Lane: dictionaries. Own nonmerge commit: `48d166cf7ba1954de46ca2a4b30d72f6872153a9`.

Changes: `internal/ir/ir.go`, `internal/javascript/javascript.go`, `internal/lower/library_object.go`, `internal/lower/view_dictionary_enumeration.go`, `internal/lower/view_dictionary_enumeration_test.go`, `internal/lower/view_lazy.go`, `internal/native/emit_expressions.go`, `internal/oracle/checked_view_dictionary_enumeration_test.go`. 64 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/ir/ir.go`
- `internal/javascript/javascript.go`
- `internal/lower/library_object.go`
- `internal/lower/view_dictionary_enumeration.go`
- `internal/lower/view_lazy.go`
- `internal/native/emit_expressions.go`
- `internal/oracle/checked_view_dictionary_enumeration_test.go`
- `stage3/interface-downcasts/dictionaries/recheck-results.json`

### 5733fdd0: Build intersection conjunction components while keeping source admission deferred

Lane: intersections. Own nonmerge commit: `5733fdd0ea93bba7122f9f61acc1870514b0bdf0`.

Changes: `internal/javascript/view_intersections.go`, `internal/lower/view_intersections.go`, `internal/lower/view_intersections_test.go`, `internal/native/runtime/view_intersections.c`, `internal/native/runtime/view_intersections.h`, `internal/oracle/checked_views_intersections_test.go`. 7 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`

### bd7042cc: Record intersection mutant evidence and defer integration to the shared integrator

Lane: intersections. Own nonmerge commit: `bd7042cc6e4a759323d5a735e488501aa2e24902`.

Changes: fixture/report/evidence only. 18 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`

### 058160aa: Prepare lazy intersection descriptors and source contract probes

Lane: intersections. Own nonmerge commit: `058160aa6c8587a329dc908ad38dddbcaedaeb38`.

Changes: `internal/lower/view_intersections.go`, `internal/lower/view_intersections_test.go`, `internal/oracle/checked_views_intersections_test.go`. 20 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/lower/view_intersections.go`
- `internal/lower/view_intersections_test.go`
- `internal/oracle/checked_views_intersections_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/RESUME-REPORT.md`
- `stage1/cohere/lint/regex/testdata/shapes/emit-absent.a`
- `stage1/cohere/lint/regex/testdata/shapes/emit-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/emit-nested.a`
- `stage1/cohere/lint/regex/testdata/shapes/emit-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/lazy-pair-progress.json`
- `stage1/cohere/lint/regex/testdata/shapes/make-integration-overlay.py`
- `stage1/cohere/lint/regex/testdata/shapes/resume-evidence/intersections-baseline-focused.log`
- `stage1/cohere/lint/regex/testdata/shapes/resume-evidence/intersections-final-focused.log`
- `stage1/cohere/lint/regex/testdata/shapes/resume-evidence/intersections-overlay-regression.log`
- `stage1/cohere/lint/regex/testdata/shapes/resume-evidence/intersections-resume-setup.log`
- `stage1/cohere/lint/regex/testdata/shapes/resume-evidence/intersections-root-conjunction.log`
- `stage1/cohere/lint/regex/testdata/shapes/resume-evidence/intersections-source-mutants.log`
- `stage1/cohere/lint/regex/testdata/shapes/resume-evidence/nested.log`
- `stage1/cohere/lint/regex/testdata/shapes/resume-evidence/shape.log`
- `stage1/cohere/lint/regex/testdata/shapes/resume-evidence/skip.log`
- `stage1/cohere/lint/regex/testdata/shapes/root-only-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/run-source-mutants.py`
- `stage1/cohere/lint/regex/testdata/shapes/shared-hooks.patch`

### 47c6c23a: Check structural intersection contracts in production reads

Lane: intersections. Own nonmerge commit: `47c6c23a4a5f9223c1b5b93f5dc719e5c19774c6`.

Changes: `internal/ir/views.go`, `internal/javascript/view_intersections.go`, `internal/javascript/view_unions.go`, `internal/lower/expression.go`, `internal/lower/view_contracts.go`, `internal/lower/view_intersections.go`, `internal/lower/view_intersections_test.go`, `internal/lower/view_lazy.go`, `internal/lower/view_objects.go`, `internal/native/view_intersections.go`, `internal/native/view_unions.go`, `internal/oracle/checked_views_intersections_test.go`. 21 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/ir/views.go`
- `internal/javascript/view_intersections.go`
- `internal/javascript/view_unions.go`
- `internal/lower/expression.go`
- `internal/lower/view_contracts.go`
- `internal/lower/view_intersections.go`
- `internal/lower/view_intersections_test.go`
- `internal/lower/view_lazy.go`
- `internal/lower/view_objects.go`
- `internal/native/view_unions.go`
- `internal/oracle/checked_views_intersections_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/PRODUCTION-REPORT.md`
- `stage1/cohere/lint/regex/testdata/shapes/brand-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/brand-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/duplicate-absent.a`
- `stage1/cohere/lint/regex/testdata/shapes/duplicate-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/duplicate-literal.a`
- `stage1/cohere/lint/regex/testdata/shapes/duplicate-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/emit-root-nested.a`
- `stage1/cohere/lint/regex/testdata/shapes/emit-root-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/intersections-extra-shapes.log`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/intersections-hooks-setup.log`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/intersections-oracle-regression.log`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/intersections-package-regression.log`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/intersections-production-focused.log`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/intersections-production-mutants.log`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/intersections-vet.log`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/nested.log`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/shape.log`
- `stage1/cohere/lint/regex/testdata/shapes/production-evidence/skip.log`
- `stage3/interface-downcasts/lane7/run-source-mutants.py`

### 60d50afc: Refuse unimplemented intersection combinations at demanded reads

Lane: intersections. Own nonmerge commit: `60d50afce8aa6e57dca9aabcc6596d9f4b4c653b`.

Changes: `internal/lower/view_intersections.go`, `internal/lower/view_lazy.go`, `internal/oracle/checked_views_intersections_test.go`. 15 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/lower/view_intersections.go`
- `internal/lower/view_lazy.go`
- `internal/oracle/checked_views_intersections_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/compound-read.a`
- `stage1/cohere/lint/regex/testdata/shapes/compound-unread.a`
- `stage1/cohere/lint/regex/testdata/shapes/followup-evidence/intersections-compound-demand.log`
- `stage1/cohere/lint/regex/testdata/shapes/followup-evidence/intersections-compound-mutant.log`
- `stage1/cohere/lint/regex/testdata/shapes/followup-evidence/intersections-final-oracle.log`
- `stage1/cohere/lint/regex/testdata/shapes/followup-evidence/intersections-final-packages.log`
- `stage1/cohere/lint/regex/testdata/shapes/followup-evidence/intersections-final-scope.log`
- `stage1/cohere/lint/regex/testdata/shapes/followup-evidence/intersections-final-vet.log`
- `stage1/cohere/lint/regex/testdata/shapes/followup-evidence/intersections-post-merge.log`
- `stage1/cohere/lint/regex/testdata/shapes/followup-evidence/intersections-recursive-mutant.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-read.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-unread.a`
- `stage3/interface-downcasts/lane7/PRODUCTION-REPORT.md`
- `stage3/interface-downcasts/lane7/lazy-pair-progress.json`

### 9c8c720e: Check finite tagged intersection arms and verify original witnesses

Lane: intersections. Own nonmerge commit: `9c8c720e32ee3cb798951bd7c3ef4a646d9af128`.

Changes: `internal/ir/views.go`, `internal/javascript/view_intersections.go`, `internal/javascript/view_unions.go`, `internal/lower/view_intersections.go`, `internal/lower/view_lazy.go`, `internal/native/view_intersections.go`, `internal/native/view_unions.go`, `internal/oracle/checked_views_intersections_test.go`. 18 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/ir/views.go`
- `internal/javascript/view_intersections.go`
- `internal/javascript/view_unions.go`
- `internal/lower/view_intersections.go`
- `internal/lower/view_lazy.go`
- `internal/native/view_intersections.go`
- `internal/native/view_unions.go`
- `internal/oracle/checked_views_intersections_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/COMPOUND-REPORT.md`
- `stage1/cohere/lint/regex/testdata/shapes/compound-evidence/checked-views.log`
- `stage1/cohere/lint/regex/testdata/shapes/compound-evidence/nested-mutant.log`
- `stage1/cohere/lint/regex/testdata/shapes/compound-evidence/packages.log`
- `stage1/cohere/lint/regex/testdata/shapes/compound-evidence/shape-mutant.log`
- `stage1/cohere/lint/regex/testdata/shapes/compound-evidence/skip-mutant.log`
- `stage1/cohere/lint/regex/testdata/shapes/compound-evidence/vet.log`
- `stage1/cohere/lint/regex/testdata/shapes/compound-untagged-read.a`
- `stage1/cohere/lint/regex/testdata/shapes/leading-access-element.a`
- `stage1/cohere/lint/regex/testdata/shapes/leading-access-identifier.a`
- `stage1/cohere/lint/regex/testdata/shapes/leading-access-missing.a`
- `stage1/cohere/lint/regex/testdata/shapes/leading-access-property.a`
- `stage1/cohere/lint/regex/testdata/shapes/leading-access-unknown-tag.a`
- `stage1/cohere/lint/regex/testdata/shapes/leading-access-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/original-read-witnesses.json`
- `stage1/cohere/lint/regex/testdata/shapes/overlap-classification.json`
- `stage3/interface-downcasts/lane7/lazy-pair-progress.json`

### 2b3a5cee: Check recursive object intersection payloads through nullable dispatch

Lane: intersections. Own nonmerge commit: `2b3a5ceea209e578cc4ad79ad03ef6fa8790f4c0`.

Changes: `internal/ir/view_intersections.go`, `internal/ir/views.go`, `internal/javascript/view_intersections.go`, `internal/javascript/view_intersections_recursive.go`, `internal/javascript/view_nullish.go`, `internal/lower/view_contracts.go`, `internal/lower/view_intersections.go`, `internal/native/runtime/adamic.h`, `internal/native/runtime/view_intersections_recursive.c`, `internal/native/runtime/view_intersections_recursive.h`, `internal/native/view_intersections.go`, `internal/native/view_intersections_recursive.go`, `internal/native/view_nullish.go`, `internal/oracle/checked_views_intersections_test.go`. 31 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/ir/views.go`
- `internal/javascript/view_intersections.go`
- `internal/javascript/view_nullish.go`
- `internal/lower/view_contracts.go`
- `internal/lower/view_intersections.go`
- `internal/native/view_intersections.go`
- `internal/native/view_nullish.go`
- `internal/oracle/checked_views_intersections_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/RECURSIVE-REPORT.md`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-absent.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-boolean-literal.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-deep-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/canonical.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/checked-views.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/final-focused.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/literal-code.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/literal-enabled.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/literal-mode.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/mutant-summary.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/nested.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/optional.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/packages.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/recursive-cases.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/shape.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/skip.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/unknown-tag-mutant.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-evidence/vet.log`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-literal-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-literal-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-missing.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-null-root-absent.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-null-root-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-number-literal.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-object-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-optional-absent.a`
- `stage1/cohere/lint/regex/testdata/shapes/recursive-optional-root.a`
- `stage1/cohere/lint/regex/testdata/shapes/run-recursive-mutants.py`

### 59eb65c8: Record integrated intersection gates and reproducible arm mutants

Lane: intersections. Own nonmerge commit: `59eb65c8f90a26bc99e88c3737bbc7f0360ea48a`.

Changes: fixture/report/evidence only. 20 other paths.

Cherry-pick exit 1; unresolved paths:

- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/arm-nested.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/arm-shape.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/arm-skip.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/arm-summary.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/arm-tag.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/checked-views.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/full-ir.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/packages.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/recursive-canonical.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/recursive-literal-code.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/recursive-literal-enabled.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/recursive-literal-mode.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/recursive-nested.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/recursive-optional.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/recursive-shape.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/recursive-skip.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/recursive-summary.log`
- `stage1/cohere/lint/regex/testdata/shapes/latest-integration-evidence/vet.log`
- `stage1/cohere/lint/regex/testdata/shapes/run-arm-mutants.py`
- `stage3/interface-downcasts/lane7/RECURSIVE-REPORT.md`

### 9e4e13fb: Certify original tracker intersection field with complete declarations

Lane: intersections. Own nonmerge commit: `9e4e13fb1bb9c755c1743eb0937906eeea988de8`.

Changes: `internal/javascript/readiness.go`, `internal/javascript/view_intersections.go`, `internal/lower/view_intersections.go`, `internal/lower/view_lazy.go`, `internal/native/view_intersections.go`, `internal/oracle/checked_views_intersections_original_test.go`. 16 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/javascript/readiness.go`
- `internal/javascript/view_intersections.go`
- `internal/lower/view_intersections.go`
- `internal/lower/view_lazy.go`
- `internal/native/view_intersections.go`
- `stage3/interface-downcasts/lane7/lazy-pair-progress.json`

### 4a5bc26b: Certify original Bindable intersection reads with a bounded walk

Lane: intersections. Own nonmerge commit: `4a5bc26befb0a00451a7a18ee0458ca5d636b9c7`.

Changes: `internal/ir/view_intersections.go`, `internal/ir/view_intersections_test.go`, `internal/ir/views.go`, `internal/javascript/view_intersections.go`, `internal/javascript/view_intersections_recursive.go`, `internal/lower/view_intersections.go`, `internal/lower/view_lazy.go`, `internal/native/runtime/view_intersections_recursive.c`, `internal/native/runtime/view_intersections_recursive.h`, `internal/native/view_intersections.go`, `internal/native/view_intersections_recursive.go`, `internal/oracle/checked_views_intersections_original_test.go`. 23 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/ir/view_intersections.go`
- `internal/ir/views.go`
- `internal/javascript/view_intersections.go`
- `internal/javascript/view_intersections_recursive.go`
- `internal/lower/view_intersections.go`
- `internal/lower/view_lazy.go`
- `internal/native/runtime/view_intersections_recursive.c`
- `internal/native/runtime/view_intersections_recursive.h`
- `internal/native/view_intersections.go`
- `internal/native/view_intersections_recursive.go`
- `internal/oracle/checked_views_intersections_original_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/BINDABLE-REPORT.md`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-access-chain-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-access-expression-this.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-access-expression-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-access-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-access-left-kind.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-access-left-symbol.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-access-left-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-element-argument-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-element-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-static-chain-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-static-expression-this.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-static-expression-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-static-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-static-helpers-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-static-helpers-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-static-left-kind.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/bindable-static-left-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/run-bindable-mutants.py`
- `stage3/interface-downcasts/lane7/lazy-pair-progress.json`
- `stage3/interface-downcasts/lane7/original/certification.json`
- `stage3/interface-downcasts/lane7/original/prepare.cjs`

### 990007bc: Certify original emitNode, JSDoc heritage and JSDoc owner reads

Lane: intersections. Own nonmerge commit: `990007bccb58642c1565bfcb687e43b707518aef`.

Changes: `internal/oracle/checked_views_intersections_original_test.go`. 29 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/oracle/checked_views_intersections_original_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/NODES-REPORT.md`
- `stage1/cohere/lint/regex/testdata/shapes/original/class-augments-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/class-augments-kind.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/class-augments-name.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/class-augments-symbol.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/class-implements-arguments.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/class-implements-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/class-implements-kind.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/class-implements-symbol.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/emit-original-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/emit-original-missing.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/emit-original-range-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/emit-original-ranges-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/emit-original-wrong.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/jsdoc-parent-function-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/jsdoc-parent-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/jsdoc-parent-kind.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/jsdoc-parent-parameters.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/jsdoc-parent-symbol.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/jsdoc-root-good.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/jsdoc-root-kind.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/jsdoc-root-symbol.a`
- `stage1/cohere/lint/regex/testdata/shapes/original/run-pair-mutants.py`
- `stage1/cohere/lint/regex/testdata/shapes/probes/plain-union-optional-chain.a`
- `stage3/interface-downcasts/lane7/lazy-pair-progress.json`
- `stage3/interface-downcasts/lane7/original/certification.json`
- `stage3/interface-downcasts/lane7/original/prepare.cjs`

### b26025fe: Check deferred intersection members at every read path

Lane: intersections. Own nonmerge commit: `b26025fefdef283bb41667bce8477abaeb8a6df9`.

Changes: `internal/ir/view_intersections.go`, `internal/lower/collections.go`, `internal/lower/view_intersections.go`, `internal/oracle/checked_views_intersections_deferred_test.go`, `internal/oracle/interface_cast_counts_test.go`. 41 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/ir/view_intersections.go`
- `internal/lower/collections.go`
- `internal/lower/view_intersections.go`
- `internal/oracle/counts.md`
- `internal/oracle/interface_cast_counts_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/DEFERRED-REPORT.md`
- `stage1/cohere/lint/regex/testdata/shapes/deferred-member-callback.a`
- `stage1/cohere/lint/regex/testdata/shapes/deferred-member-destructure.a`
- `stage1/cohere/lint/regex/testdata/shapes/deferred-member-helper.a`
- `stage1/cohere/lint/regex/testdata/shapes/deferred-member-read.a`
- `stage1/cohere/lint/regex/testdata/shapes/deferred-member-unread.a`
- `stage3/interface-downcasts/lane7/original/bindable-access-chain-good.a`
- `stage3/interface-downcasts/lane7/original/bindable-access-expression-this.a`
- `stage3/interface-downcasts/lane7/original/bindable-access-expression-wrong.a`
- `stage3/interface-downcasts/lane7/original/bindable-access-good.a`
- `stage3/interface-downcasts/lane7/original/bindable-access-left-kind.a`
- `stage3/interface-downcasts/lane7/original/bindable-access-left-symbol.a`
- `stage3/interface-downcasts/lane7/original/bindable-access-left-wrong.a`
- `stage3/interface-downcasts/lane7/original/bindable-element-argument-wrong.a`
- `stage3/interface-downcasts/lane7/original/bindable-element-good.a`
- `stage3/interface-downcasts/lane7/original/bindable-static-chain-good.a`
- `stage3/interface-downcasts/lane7/original/bindable-static-expression-this.a`
- `stage3/interface-downcasts/lane7/original/bindable-static-expression-wrong.a`
- `stage3/interface-downcasts/lane7/original/bindable-static-good.a`
- `stage3/interface-downcasts/lane7/original/bindable-static-helpers-good.a`
- `stage3/interface-downcasts/lane7/original/bindable-static-helpers-wrong.a`
- `stage3/interface-downcasts/lane7/original/bindable-static-left-kind.a`
- `stage3/interface-downcasts/lane7/original/bindable-static-left-wrong.a`
- `stage3/interface-downcasts/lane7/original/class-augments-good.a`
- `stage3/interface-downcasts/lane7/original/class-augments-kind.a`
- `stage3/interface-downcasts/lane7/original/class-augments-name.a`
- `stage3/interface-downcasts/lane7/original/class-augments-symbol.a`
- `stage3/interface-downcasts/lane7/original/class-implements-arguments.a`
- `stage3/interface-downcasts/lane7/original/class-implements-good.a`
- `stage3/interface-downcasts/lane7/original/class-implements-kind.a`
- `stage3/interface-downcasts/lane7/original/class-implements-symbol.a`
- `stage3/interface-downcasts/lane7/original/jsdoc-parent-function-good.a`
- `stage3/interface-downcasts/lane7/original/jsdoc-parent-good.a`
- `stage3/interface-downcasts/lane7/original/jsdoc-parent-kind.a`
- `stage3/interface-downcasts/lane7/original/jsdoc-parent-parameters.a`
- `stage3/interface-downcasts/lane7/original/jsdoc-parent-symbol.a`
- `stage3/interface-downcasts/lane7/original/jsdoc-root-good.a`
- `stage3/interface-downcasts/lane7/original/jsdoc-root-kind.a`
- `stage3/interface-downcasts/lane7/original/jsdoc-root-symbol.a`

### 99ec3e0a: Certify original Node union reads through declared ancestry

Lane: intersections. Own nonmerge commit: `99ec3e0a56b92276980151d7edc9a9b8d8d35b34`.

Changes: `internal/lower/view_intersections.go`, `internal/lower/view_intersections_test.go`, `internal/oracle/checked_views_intersections_absorption_test.go`, `internal/oracle/checked_views_intersections_original_test.go`. 4 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/lower/view_intersections.go`
- `internal/lower/view_intersections_test.go`
- `internal/oracle/checked_views_intersections_original_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/ABSORPTION-REPORT.md`
- `stage3/interface-downcasts/lane7/original/certification.json`
- `stage3/interface-downcasts/lane7/original/prepare.cjs`

### 503cd3f9: Certify synthetic heritage intersection field reads

Lane: intersections. Own nonmerge commit: `503cd3f950d76e31d3f476617fb030b2fce4fadd`.

Changes: `internal/lower/object.go`, `internal/lower/view_intersections.go`, `internal/oracle/checked_views_intersections_heritage_union_test.go`, `internal/oracle/checked_views_intersections_original_test.go`. 4 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/lower/object.go`
- `internal/lower/view_intersections.go`
- `internal/oracle/checked_views_intersections_original_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/HERITAGE-UNION-REPORT.md`
- `stage3/interface-downcasts/lane7/original/certification.json`
- `stage3/interface-downcasts/lane7/original/prepare.cjs`

### 01a1eda4: Record original array and Identifier certification blockers

Lane: intersections. Own nonmerge commit: `01a1eda4a6bb4e86031176a63a282f85ccc6e124`.

Changes: `internal/oracle/checked_views_brands_original_gaps_test.go`, `internal/oracle/checked_views_intersections_array_blockers_test.go`, `internal/oracle/checked_views_intersections_identifier_test.go`, `internal/oracle/checked_views_intersections_original_test.go`. 5 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/oracle/checked_views_brands_original_gaps_test.go`
- `internal/oracle/checked_views_intersections_original_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/REMAINING-REPORT.md`
- `stage3/interface-downcasts/lane4/original/BLOCKERS.md`
- `stage3/interface-downcasts/lane7/original/certification.json`
- `stage3/interface-downcasts/lane7/original/prepare.cjs`

### d5c2ecce: Certify the original Identifier helper member read

Lane: intersections. Own nonmerge commit: `d5c2ecce91c545f2766fd4937e617eeb4e12abcd`.

Changes: `internal/oracle/checked_views_intersections_identifier_test.go`. 5 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/oracle/checked_views_intersections_identifier_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/IDENTIFIER-REPORT.md`
- `stage1/cohere/lint/regex/testdata/shapes/counts.md`
- `stage3/interface-downcasts/lane4/original/BLOCKERS.md`
- `stage3/interface-downcasts/lane7/original/certification.json`

### fd73f8b6: Certify isolated intersection helper member reads

Lane: intersections. Own nonmerge commit: `fd73f8b6f51d6ca55a0636b3ec62dd649d9f8e5b`.

Changes: `internal/oracle/checked_views_intersections_isolated_members_test.go`, `internal/oracle/checked_views_intersections_original_test.go`. 5 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/oracle/checked_views_intersections_original_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/ISOLATED-REPORT.md`
- `stage3/interface-downcasts/lane7/counts.md`
- `stage3/interface-downcasts/lane7/original/certification.json`
- `stage3/interface-downcasts/lane7/original/prepare.cjs`

### 4c3c3009: Verify assigned intersection overlaps and original array blockers

Lane: intersections. Own nonmerge commit: `4c3c3009c1ab74e4da902ffa9357b81ec0cf7d95`.

Changes: `internal/oracle/checked_views_intersections_array_blockers_test.go`. 3 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/oracle/checked_views_intersections_array_blockers_test.go`
- `stage1/cohere/lint/regex/testdata/shapes/ASSIGNED-REPORT.md`
- `stage1/cohere/lint/regex/testdata/shapes/assigned-lane4b-certificates.json`

### e8e97d74: Certify original literal import argument after views integration

Lane: intersections. Own nonmerge commit: `e8e97d742e10b815163ab3a188c206344d2e724b`.

Changes: `internal/oracle/checked_views_intersections_ranked_test.go`. 6 other paths.

Cherry-pick exit 1; unresolved paths:

- `stage3/interface-downcasts/lane7/counts.md`
- `stage3/interface-downcasts/lane7/lazy-pair-progress.json`

### 6c66af6b: Certify two original numeric operand intersections

Lane: intersections. Own nonmerge commit: `6c66af6b2d183531715692546b0895cf714bbf4e`.

Changes: `internal/oracle/checked_views_intersections_ranked_test.go`. 6 other paths.

Cherry-pick exit 1; unresolved paths:

- `internal/oracle/checked_views_intersections_ranked_test.go`
- `stage3/interface-downcasts/lane7/counts.md`
- `stage3/interface-downcasts/lane7/original/prepare-ranked.cjs`

### ee1257a3: Complete ranked intersection certificates and record remaining dependencies

Lane: intersections. Own nonmerge commit: `ee1257a3c4a6f026d8773973c6bf8cf6efbbd6a9`.

Changes: `internal/oracle/checked_views_intersections_ranked_test.go`. 7 other paths.

Cherry-pick exit 1; unresolved paths:

- `internal/oracle/checked_views_intersections_ranked_test.go`
- `stage3/interface-downcasts/lane7/counts.md`
- `stage3/interface-downcasts/lane7/original/prepare-ranked.cjs`
- `stage3/interface-downcasts/lane7/pair-progress.json`
- `stage3/interface-downcasts/lane7/rank-pairs.py`

### 040999bc: Record morning intersection recheck and array lane dependency

Lane: intersections. Own nonmerge commit: `040999bcec89e1e29c707f954880cc9142346a72`.

Changes: fixture/report/evidence only. 1 other paths.

Cherry-pick exit 0: textual application only, not a compile or execution result.

### c41eaf85: Select primitive and nullish dictionary arms without erasing reference obligations

Lane: dictionary joint. Own nonmerge commit: `c41eaf85f7417bf3d33532fce30457d70be1440e`.

Changes: `internal/ir/ir.go`, `internal/ir/view_dictionaries.go`, `internal/ir/view_dictionary_primitives_test.go`, `internal/javascript/view_dictionaries.go`, `internal/lower/library_string.go`, `internal/lower/readiness.go`, `internal/lower/records.go`, `internal/lower/view_dictionaries.go`, `internal/lower/view_dictionary_primitives.go`, `internal/lower/view_dictionary_primitives_test.go`, `internal/lower/view_lazy.go`, `internal/native/runtime/view_dictionaries.c`, `internal/native/view_dictionaries.go`, `internal/oracle/checked_views_primitive_dictionaries_test.go`. 50 other paths.

Cherry-pick exit 1; unresolved paths:

- `docs/checked-views-plan.md`
- `internal/ir/ir.go`
- `internal/ir/view_dictionaries.go`
- `internal/javascript/view_dictionaries.go`
- `internal/lower/library_string.go`
- `internal/lower/readiness.go`
- `internal/lower/records.go`
- `internal/lower/view_dictionaries.go`
- `internal/lower/view_lazy.go`
- `internal/native/runtime/view_dictionaries.c`
- `internal/native/view_dictionaries.go`
- `stage3/interface-downcasts/lane4/original/certification.json`
- `stage3/interface-downcasts/lane4/primitive-original/REPORT.md`
- `stage3/interface-downcasts/lane4/primitive-original/certification.json`

### 8861b316: Carry readonly dictionary consumers to checked lookup

Lane: dictionary joint. Own nonmerge commit: `8861b3160eb4d700e74bcdff897d3dfbea6541df`.

Changes: `internal/ir/ir.go`, `internal/javascript/javascript.go`, `internal/lower/cast.go`, `internal/lower/cast_proof.go`, `internal/lower/readiness.go`, `internal/lower/records.go`, `internal/lower/view_dictionary_conversion.go`, `internal/native/union.go`. 0 other paths.

Cherry-pick exit 1; unresolved paths:

- `internal/ir/ir.go`
- `internal/javascript/javascript.go`
- `internal/lower/cast.go`
- `internal/lower/readiness.go`
- `internal/lower/records.go`
- `internal/native/union.go`

### c77f7326: Check rich dictionary selection and transitive reads

Lane: dictionary joint. Own nonmerge commit: `c77f73261e9f9fa5e505c76fa9945d7b1252417e`.

Changes: `internal/ir/view_dictionaries.go`, `internal/lower/view_dictionaries.go`, `internal/lower/view_dictionary_references.go`, `internal/lower/view_lazy.go`, `internal/oracle/checked_view_dictionary_joint_test.go`. 14 other paths.

Cherry-pick exit 1; unresolved paths:

- `internal/ir/view_dictionaries.go`
- `internal/lower/view_dictionaries.go`
- `internal/lower/view_lazy.go`

### 29ffa565: Keep primitive controls consistent with lazy reference selection

Lane: dictionary joint. Own nonmerge commit: `29ffa56577dd1f3921d62196d057d8b978955554`.

Changes: `internal/oracle/checked_views_primitive_dictionaries_test.go`. 0 other paths.

Cherry-pick exit 1; unresolved paths:

- `internal/oracle/checked_views_primitive_dictionaries_test.go`

### bd3ede0d: Record lane 4b dictionary certification and remaining consumers

Lane: dictionary joint. Own nonmerge commit: `bd3ede0d5c57b9939324aa4e10387667a6d8c4c6`.

Changes: fixture/report/evidence only. 1 other paths.

Cherry-pick exit 0: textual application only, not a compile or execution result.

### 8f5ac7cb: Check dictionary representation at reference consumption

Lane: dictionary joint. Own nonmerge commit: `8f5ac7cbba11559698aafbd6ee935a22555589af`.

Changes: `internal/ir/ir.go`, `internal/javascript/javascript.go`, `internal/javascript/view_dictionaries.go`, `internal/lower/view_dictionary_conversion.go`, `internal/lower/view_dictionary_conversion_test.go`, `internal/lower/view_lazy.go`, `internal/native/runtime/view_dictionaries.c`, `internal/native/runtime/view_dictionaries.h`, `internal/native/union.go`. 0 other paths.

Cherry-pick exit 1; unresolved paths:

- `internal/ir/ir.go`
- `internal/javascript/javascript.go`
- `internal/javascript/view_dictionaries.go`
- `internal/lower/view_dictionary_conversion.go`
- `internal/lower/view_lazy.go`
- `internal/native/runtime/view_dictionaries.c`
- `internal/native/runtime/view_dictionaries.h`
- `internal/native/union.go`

### dbeb7b5c: Certify dictionary consumers across the three views lanes

Lane: dictionary joint. Own nonmerge commit: `dbeb7b5c030b48f2b80e23b9855c9dc4a9ce159c`.

Changes: `internal/oracle/checked_view_dictionary_joint_test.go`. 15 other paths.

Cherry-pick exit 1; unresolved paths:

- `internal/oracle/checked_view_dictionary_joint_test.go`
- `stage3/interface-downcasts/dictionaries/joint/buildoptions-wrong-representation.a`
- `stage3/interface-downcasts/dictionaries/joint/compileroptions-wrong-representation.a`
- `stage3/interface-downcasts/dictionaries/joint/counts.md`

### a1b0298c: Record lane 6 checked dictionary recheck results

Lane: dictionary joint. Own nonmerge commit: `a1b0298cad1bd30cbfb641066970c69ce08a567c`.

Changes: fixture/report/evidence only. 1 other paths.

Cherry-pick exit 1; unresolved paths:

- `stage3/interface-downcasts/dictionaries/recheck-results.json`

### 9a917fc5: Report dictionary joint certification and exclusions

Lane: dictionary joint. Own nonmerge commit: `9a917fc54cb131167a08e44272f9dc7c633a5426`.

Changes: fixture/report/evidence only. 7 other paths.

Cherry-pick exit 1; unresolved paths:

- `stage3/interface-downcasts/dictionaries/joint/REPORT.md`
