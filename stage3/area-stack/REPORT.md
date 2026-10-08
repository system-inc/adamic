# Group 4 generic-return clock follow-up

Built the finite optional generic plain-object intersection result proof from clock-15-20-generic-returns-t-01.
Commits: 1cdc3439c replays the topic's sole own commit 210f6ead2; 78fad25de and 8fc662bb6 strengthen null and index proof tests without changing production lowering.
Checks: lower 50.957s, IR 25.109s, flow 171.055s; JavaScript has no tests. Delivery Node oracle and output mutant pass (0.763s); counts refresh passes (82.073s); strengthened proof tests pass (0.282s).
Mutants: present-as-undefined is caught by clean Node stdout disagreement in both backends; finite-shape bypass is caught by kind and direct index proof assertions; null erasure is caught by the direct checked-signature assertion; removing dispatch is caught by the positive readonly shape's NotYet.
Not covered: the real AST root remains NotYet at its void brand field; zero census roots are claimed cleared. The earlier whole generic-returns-t topic and checked-view lanes remain out.

The only production insertion calls clockGenericReturnsT01 after the stack's normal representation, undefined-result and regex callback result handling have failed. Its proof retains every checked constituent and field. It accepts only the owned generic object/refinement pair plus undefined, with finite primitive/nested object fields and string arrays. Callable, constructable, indexed, container intersection, nested callable and distinct null shapes remain refused. No views implementation, worker merge or ABI replacement was imported. Counts were the only conflict: existing rows were retained, one disjoint row was added, then the table was regenerated.

The original null negative test let a null-erasure source mutant survive because an independent later lowering refusal still rejected the program. The original indexed probe similarly hid a finite-shape bypass. Two checked-signature tests now assert the proof's refusal before any body or call lowering. The rerun catches both mutants at those assertions. Production guard code was not relaxed or changed.

Exact commands: go test ./internal/lower ./internal/ir ./internal/flow ./internal/javascript; go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/clock_generic_returns_t_01|TestClockGenericReturnsT01Mutant' -count=1 -v; go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts; go test ./internal/lower -run '^TestClockGenericReturnsT01' -count=1 -v; python3 stage3/area-stack/clock-generic-mutants.py. All output went directly to retained logs. The source mutants use Go overlays and never edit production files.

# Group 4 host delivery

Catalog on c9cc35939dac2509ed2b08dd8c730123a4122495: exit 0, 238.959s; all 11 applicable undo patches were caught. The same five historical entries remain nonapplicable. No undo patch drifted. Controls, mutant failures and results are retained under evidence/group4-host/catalog.

Built destination-typed empty array literals and debugger statements, and added the Source representation boundary proof without changing its production admission.
Commits: 8f6674046 and 49fe63477 replay array-literal-never-element's two own commits; 6c38c3059 and 234c3791e replay debugger-statement's two own commits; 963e7c538 replays Source commit 19bcb8ee.
Checks: lower 56.045s, IR passed from its unchanged cached run, flow 166.618s, JavaScript has no tests; focused Node oracle 4.284s and refreshed counts 90.059s both passed. Catalog follows.
Mutants: one wrong empty-array element kind, six debugger mutations and three Source representation/output mutations were caught, as detailed below.
Not delivered: generic-function-value and landing/presence topics needing stricter options or checked views, class-features' failing view fixture, memoize's runtime graph regions, and error-classes' incompatible catch admissions.

Empty arrays use the contextual, enclosing or implied destination's element representation. The nonempty literal proof stays as on the stack. Four Node fixtures and two lower layout checks pass. The wrong-kind source overlay fails object/string layout assertions, and sanitized native catches 124 leaked bytes in two allocations on a positive fixture. This is one mutant held to both the layout and outside-oracle ownership checks.

Debugger remains in JavaScript and has no native trap or flow instruction. The six mutants are caught by: native artifact trap detection; Node exit-code disagreement after a trap; missing JavaScript debugger statement; missing lowerer debugger statement; the explicit debugger refusal; and the extra flow instruction. The mutated paths are byte-identical between the earlier host review and this restack, verified before relying on their catches. No mutation was active during package, oracle or counts checks.

Source remains unresolved until its concrete instantiation is known. A structural CompilerType constraint admits both Object and Closure, and a length-only constraint admits both Array and String. No unresolved parameter is assigned an Object ABI. Two production overlays that do so fail the checked-type assertions. The ignored-payload IR mutant finishes without leaks but disagrees with Node stdout in both backends. 19bcb8ee adds that evidence and fixture only; it does not admit an unsafe constraint.

Generic-function-value is skipped as a whole. Its own 7d70c5728 modifies internal/lower/sparse_arrays.go and stage3/stricter-options/{REPORT.md,indexed_test.go}, absent on this stack. The topic's sparse-array change needs the stricter-options/ArrayHoles infrastructure. That group is explicitly held for checked views. The reviewed prefix put generic identity checking before every intrinsic early return and retained the census signature proof, but the entire topic was aborted at the dependency; no prefix is delivered.

Class-features-review remains skipped as a whole: combining the descendant and every-return-receiver proofs still leaves its own positive iterators_sym_keys_view.a:2:66 at the SymbolKey/checked-view boundary. Devirtualize's two own commits were already delivered at 2c7d9fd2 and remain.

Error-classes-counts-2 is skipped. Its positive error_classes.a contains a try around repeat(-1) at 87:1 and toFixed(101) at 79:1. The stack retains both try-around-runtime-failure refusals and its real-Error allocation/throw proof; the incoming topic removes those refusals and makes the failures catchable. The preserved-repeat reduction is held/error_classes_repeat.a. Observed: Node prints caught; production lowering stops at 1:1 with the exact try-around-repeat NotYet. Inference: keeping that required refusal prevents the topic's positive operation from lowering. No full error topic pass or mutant pass is claimed. Its full conflict diff records the class/static/accessor, throw, narrowing, object helper and Node exit convention sides. Proposal if the ruling is changed: first prove equivalent catchable runtime failure ownership and nominal Error allocation ancestry, then add the constructor/cause rules while preserving static/accessor/private layouts and counted dispatch. Under the current ruling every existing refusal stays, so the whole topic is out.

Memoize-regions remains skipped for runtime's captured-cell graph regions and environment adoption. Landing-batch-3 and optional-field-write-2 remain skipped for checked views, as ruled; optional-field-write-2 is the held group 1 prerequisite. No view or predicate lane was imported.

Commands: go test ./internal/lower ./internal/ir ./internal/flow ./internal/javascript; go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(array_literal_empty_|debugger_|representation_clock_source)|TestDebuggerNativeEmitsNothing|TestRepresentationClockSourceMutant' -count=1 -v; go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts. Mutants: internal/lower/testdata/debugger/run-mutants.py; two Source production overlays running TestRepresentationClockSourceCheckedTypes; an object.go wrong-kind overlay running the two empty-layout tests and the four registered Node fixtures. The isolated review references the pinned SDK and Node types installed by cloud/setup.sh. Its initial package run lacked the Node types; it was stopped and retried after making the installed pinned types available. Its reviewed production and test files are identical to the delivery files, excluding the regenerated counts table and report evidence.

Setup: nproc 5, CPU quota 4; Go 1.27.1, Node 24.19, clang 20.1.8. Recorded setup timings: Go 0.024s, Node 0.027s, submodules 0.061s, markdown 0.076s, step 0.008s, clang 0.180s, build 49.385s, deferred 49.521s, cache 49.522s, total 49.555s. GOPROXY used https://proxy.golang.org|direct.

# Group 3 ruled delivery

Catalog on 2442285a445a5b553e619ee3f5b3c2ab664305ae: exit 0, 276.825s; all 11 applicable undo patches were caught. The same five historical entries remain nonapplicable. No undo patch drifted. Per-entry controls, mutants and results are retained under evidence/group3-ruled/catalog.

Built boxed callable unions, registered class constructor values and closed lexical constructor caches, and library-small operations through the existing regex callback protocol.
Commits: own function-values, new-expression and library-small commits replayed individually after 0a167e089; the void-value topic was withdrawn after a base-fixture regression.
Checks: lower 48.477s, IR 27.634s, flow 164.577s, JavaScript has no tests; required results and refreshed counts are in evidence/group3-ruled; the focused Node oracle passed in both backends (9.263s), and counts refreshed successfully (75.720s).
Mutants: five function boundary/source mutants, eleven constructor mutants and thirteen library proof/runtime/IR checks were caught; catches and commands follow.
Not delivered: topics failing a retained refusal, checked-view dependencies, and resolutions held by automatic approval review.

Function-values preserves the current Map union fixtures and drops the topic's reinstated Map-union refusal witness. The five mutants cover return/argument union fitting, invariant reference views and call/return boxing. Both native and JavaScript are compared with Node; the view controls require a loud stop.

New-expression uses CallTargets for the lexical initializer and requires every known target to have no closure or environment. Zero arguments and additional-capture rules remain. Six class-value mutants cover eager caching, missing cache storage, static allocators, callee re-read, repeated construction and repeated arguments. Three local-cache mutants cover eager initialization, missing storage and extra captures; the cache-or eager-init and Uint16 admission mutants complete eleven catches. Uint16 stays withdrawn pending its runtime representation. Its two deliberate refusal controls and the extra-capture control are excluded from flow's runnable program list by exact filenames, while their explicit refusal tests stay registered.

Library-small retains regexReplacement's runtime callback protocol. Seven runtime mutations cover match content, literal replacement, Unicode, global iteration, offset, lastIndex reset and collection order; two offset/input controls, three string bound/position mutations and the capture-free proof mutation complete thirteen checks. Both script entry points invoke these current runtime checks. Literal object-entry inference uses the topic's const-assertion proof. No native runtime ABI was changed by this delivery.

Void-value is skipped: its new helper refuses the existing positive stage3/fixtures/taste/13_void_callback.a:6:16 callable result. Its own three positives and seven mutants passed before withdrawal, but that does not justify regressing the stack. The final oracle includes the preserved taste fixture.

Optional-call is skipped: optional_call_collections.a:18:17 hits the retained single-intrinsic-operand receiver refusal. Its seven other positive fixtures passed, including the preserved literal receiver behavior.

Element-type is skipped: element_type_array_unions.a:8:32 reaches the retained uncheckable object-tag refusal; element_type_unions.a:14:16 reaches the template object/array/map/function/undefined boundary.

Rest-args is skipped: rest_tuple_create.a:4:17 and rest_tuple_update.a:4:17 retain the non-array rest refusal; rest_union_token.a:8:62, rest_union_diagnostic.a:7:60 and rest_callable_write.a:10:27 retain the unpackable-spread refusal. A review-only signature fix turns the checker panic into the exact named stop; neither the topic nor that fix is delivered.

This-outside is skipped: union_object_kind.a:3:13 and union_object_kind_stale.a:5:13 retain the uncheckable object-tag refusal; the five generic receiver positives passed.

Generic-returns-t is skipped: notyet_signature_object.a:4:28 needs the opaque host collection/object checked-view proof; eleven other positive reductions passed.

Case-declaration, void-or-undefined and for-in remain held by automatic approval review. File-by-file stack and topic behaviors and concrete proposals are in REMAINING-PROPOSALS.md. No rejected edit ran. Method-values and binary remain held by the earlier proposals. For-of-object, class-construction, class-set-property, object-property, destructuring, object-small and representations remain skipped for checked views as ruled. Overloads, statements-small, element-access and syntax-kinds were delivered at e0e234ea and stay in the stack.

Commands: go test ./internal/lower ./internal/ir ./internal/flow ./internal/javascript; the exact focused oracle and counts commands are recorded in COMMANDS.md. Mutation scripts: run-function-values-mutants.py, run-function-values-boundary-mutants.py, run-class-value-mutants.py, run-cache-local-mutants.py, run-cache-or-mutant.py, run-uint16-notyet-mutant.py, run-library-small-regex-mutants.py, the library runtime and IR mutation tests, and the capture-proof source overlay. The withdrawn void-value script was also run and caught seven mutants. All retained logs are compressed without changing their content.

# Group 2 ruled statics delivery

Catalog on 7941cee0fcc51626e00b11ac761d071bb2c31122: exit 0, 255.426s, all 11 applicable undo patches applied and were caught; five recorded nonapplicable entries stayed skipped. Command: `bash verify/catalog/check.sh HEAD --jobs 2`. No undo patch needed refresh.

Built structural static-method dispatch through the existing counted method ABI; conditions remain delivered.
Commits: d69d05ca0 and 36441fa62 replay b6aa4f00 and 1db2b324; 829446f68 adapts the mutant anchors to counted dispatch.
Checks: lower 46.956s, IR 28.961s, flow 172.762s; six topic oracle fixtures passed against Node in both backends (1.555s); counts refresh passed (82.739s).
Mutants: blanket guard, missing static map, guessed static side, missing inherited map, missing optional short circuit, missing generic static map, JavaScript missing map and JavaScript late map all failed their intended controls.
Not delivered: logical-assignment-comma, method-values, binary and computed-name destructuring; reasons follow.

Commands: `go test ./internal/lower ./internal/ir ./internal/flow ./internal/javascript`; `go test ./internal/oracle -run 'TestNativeAgreesWithNode/.*/structural_statics_' -count=1 -v`; `python3 cloud/notyet-statics-mutants.py`; `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts`. Raw results and individual mutant catches are in [evidence/group2-statics](evidence/group2-statics).

Logical-assignment-comma is skipped as a whole: its positive `logical_or_assignment.a:223:17` reaches the retained accessor logical-assignment refusal. The ruling requires that refusal to remain.

Method-values remains held as described in [METHOD-VALUES-PROPOSAL.md](METHOD-VALUES-PROPOSAL.md). Automatic approval review rejected its unverified native closure ABI conflict resolution. No rejected edit ran.

Binary remains held. Automatic approval review rejected the proposed combined ownership, freshness and assignment IR resolution as an unverified semantic merge that risks miscompilation. Files: `internal/flow/build.go`, `internal/fresh/fresh.go`, `internal/ir/ir.go`, `internal/lower/assignment_value.go`, `internal/lower/expression.go`. The stack has compound-assignment ownership and write guards; the topic introduces expression-valued plain assignments and freshness tracking. Proposal: retain the stack's compound operations and IR fields, introduce plain assignment values through the existing assignment lowering and all its guards, and combine both freshness/ownership analyses. No rejected edit ran; the topic cherry-pick was aborted.

Computed-name destructuring remains skipped for checked views, as ruled. Namespace and enum admission remain held for the two namespace judgments recorded in [PARSER-REVIEW.md](PARSER-REVIEW.md); no views code was imported.

# Current parser and scanner review

See [PARSER-REVIEW.md](PARSER-REVIEW.md) for the fixture-base merge, every parser assertion outcome, exact enum and mutable-namespace dependencies, current checks and group 4 additions. Namespace prerequisite work is retained on compiler/area-stack-namespace-held pending two rulings; it is absent from this delivery. Checked-view topics remain skipped.

# Compiler area stack review

Delivery branch: `compiler/area-stack`. Base `337aa466`; ruled runtime drop `5c05a776` merged as `1d60ab525`.

Topics below are skipped as whole topics. No unresolved compiler hunk was accepted. Each linked diff records every conflicting file and both sides: the HEAD block before `=======` is the stack side, and the block after it is the incoming topic side (Git's leading combined-diff columns are bookkeeping); the proposal applies to every semantic hunk in that topic. Count-only conflicts retain existing rows and add disjoint witnesses, followed by counts regeneration.

## enum-init-reach

Files: `internal/lower/enum_initialization.go`, `internal/lower/namespaces.go`, `internal/lower/refusals.go`.

Both sides, hunk by hunk: [enum-init-reach.diff](enum-init-reach.diff).

Proposal: Keep the stack's indirect-call and pending-initializer refusals. Adopt shared enum/namespace reachability only after namespaceCallGraph prerequisites are ruled, retaining class/readiness checks and dead-branch refusal policy.

## logical-assignment-comma

Files: `internal/lower/expression.go`, `internal/lower/logical_assignment.go`, `internal/lower/refusals.go`, `internal/native/slots.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [logical-assignment-comma.diff](logical-assignment-comma.diff).

Proposal: Keep IR.LogicalAssignment ownership and the counted/packed ABI; add expression-scope/captured stores without replacing those operations. Retain accessor and unsupported target refusals. Requires a lowering ruling.

## method-values

Files: `internal/javascript/javascript.go`, `internal/native/emit_expressions.go`, `internal/native/emit_objects.go`, `internal/native/runtime/adamic.h`, `internal/native/runtime/closure.c`, `internal/native/runtime/heap.c`.

Both sides, hunk by hunk: [method-values.diff](method-values.diff).

Proposal: Keep counted and uncounted method entries and canonical receiver closures. Add bound/unbound adapters only with the same counted dispatch and ownership. Requires a closure ABI ruling.

## statics

Files: `internal/javascript/javascript.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [statics.diff](statics.diff).

Proposal: Retain counted class method calls and declared argument conventions while adding dynamic static-method selection. Requires a receiver/dispatch lowering ruling.

## binary

Files: `internal/flow/build.go`, `internal/fresh/fresh.go`, `internal/ir/ir.go`, `internal/lower/assignment_value.go`, `internal/lower/expression.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [binary.diff](binary.diff).

Proposal: Keep identifier Effects/read assignment lowering and all current target refusals; add property/index AssignmentValue operations only after ownership and program tracking are reconciled. Requires a lowering ruling.

## void-value

Files: `internal/lower/expression.go`.

Both sides, hunk by hunk: [void-value.diff](void-value.diff).

Proposal: Keep evaluation effects and erased-callable refusals; add represented void results only where their evaluation/return representation is proved. Requires a void lowering ruling.

## optional-call

Files: `internal/ir/ir.go`, `internal/javascript/javascript.go`, `internal/lower/expression.go`, `internal/lower/lower_test.go`, `internal/native/emit_functions.go`, `internal/native/runtime/adamic.h`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [optional-call.diff](optional-call.diff).

Proposal: Keep counted/direct/spread closure dispatch and adamic_method_entry. Add optional lookup/guard adapters using that ABI. Requires a closure ABI and evaluation-order ruling.

## case-declaration

Files: `internal/javascript/javascript.go`, `internal/lower/locals.go`, `internal/lower/object.go`, `internal/native/emit_locals.go`, `internal/oracle/counts.md`, `internal/oracle/switch_case_declaration_test.go`.

Both sides, hunk by hunk: [case-declaration.diff](case-declaration.diff).

Proposal: Keep switch scope, discriminant lifetime, captured ready cells and checks. Add source case-declaration support without weakening these checks. Requires a scope/readiness lowering ruling.

## this-outside

Files: `internal/fresh/fresh.go`, `internal/javascript/javascript.go`, `internal/lower/expression.go`, `internal/native/emit_expressions.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [this-outside.diff](this-outside.diff).

Proposal: Keep observed-type array predicate narrowing and undefined admission; add general runtime kind tests only with equivalent checker-observation proofs. Requires a narrowing ruling.

## void-or-undefined

Files: `internal/lower/expression.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [void-or-undefined.diff](void-or-undefined.diff).

Proposal: Keep existing void/never call result erasure and concrete substitution. Introduce optional void results only with explicit ABI/representation proofs. Requires a lowering ruling.

## rest-args

Files: `internal/lower/expression.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [rest-args.diff](rest-args.diff).

Proposal: Keep existing spread handling, forwarded argument counts and defaulted slots. Add resolved-signature rest packing without packing twice or losing counts. Requires an argument ABI ruling.

## class-construction

Files: `internal/lower/expression.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [class-construction.diff](class-construction.diff).

Proposal: Keep contextual unknownView/undefined checks; admit class-interface values only after checked-view dispatch is ruled. Dependency: checked views.

## class-set-property

Files: `internal/lower/census_small.go`, `internal/lower/object.go`, `internal/native/emit_expressions.go`, `internal/native/emit_objects.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [class-set-property.diff](class-set-property.diff).

Proposal: Keep property View/readiness/literal guards and shape kinds. Add union-field stores with checked-view ownership and runtime brand checks. Dependency: checked-view field integration; runtime 17b5a053 equivalent is 6a026ee4.

## object-property

Files: `internal/lower/expression.go`.

Both sides, hunk by hunk: [object-property.diff](object-property.diff).

Proposal: Keep contextual unknownView/undefined conversions; add optional-property-chain lowering while preserving both check order and absence behavior. Requires checked-view/optional lowering judgment.

## destructuring

Files: `internal/lower/collections.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [destructuring.diff](destructuring.diff).

Proposal: Keep declared field layouts, readiness and allowed-view guards. Add nested/default binding statements without bypassing those checks. Dependency: checked-view binding integration.

## for-in

Files: `internal/fresh/library_language.go`, `internal/native/emit_functions.go`, `internal/native/emit_objects.go`, `internal/native/emit_statements.go`, `internal/native/reuse.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [for-in.diff](for-in.diff).

Proposal: Keep dropped-extra-argument effects and property readiness checks. Add enumeration-argument metadata and writes without changing their order or ownership. Requires emitter/checked-view lowering judgment.

## library-small

Files: `internal/lower/regexp.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [library-small.diff](library-small.diff).

Proposal: Keep shared regexReplacement callback protocol, checked arguments and refusals; extend library coverage without replacing it with a one-argument callback loop. Requires callback lowering judgment.

## object-small

Files: `internal/lower/census_small.go`, `internal/lower/object.go`, `internal/native/emit_expressions.go`, `internal/native/emit_objects.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [object-small.diff](object-small.diff).

Proposal: Keep checked property reads and runtime shape kinds. Add union object operations with the same guards and ownership. Dependency: checked-view object integration.

## element-type

Files: `internal/lower/object.go`.

Both sides, hunk by hunk: [element-type.diff](element-type.diff).

Proposal: Keep existing Union element admission and runtime storage. Add optional boolean/scalar brands without removing Union acceptance. Requires representation/refusal judgment.

## function-values

Files: `internal/lower/census_small.go`, `internal/lower/lower_test.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [function-values.diff](function-values.diff).

Proposal: Keep existing Map union support and its tests; add boxed callable unions. Topic reinstates the Map union refusal witness, which conflicts with current support. Requires a refusal ruling.

## generic-returns-t

Files: `internal/lower/functions.go`.

Both sides, hunk by hunk: [generic-returns-t.diff](generic-returns-t.diff).

Proposal: Keep the special checked regex callback result representation for null/undefined. Apply concrete signature results without replacing that protocol. Requires return lowering ruling.

## generic-function-value

Files: `internal/lower/expression.go`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [generic-function-value.diff](generic-function-value.diff).

Proposal: Keep optional intrinsic handling and all call refusals. Add generic function identity checks with a ruled order relative to intrinsic early return. Requires refusal ordering judgment; later nullable/record changes also depend on record/view lanes.

## class-features-review

Files: `internal/lower/iteration_origin.go`.

Both sides, hunk by hunk: [class-features-review.diff](class-features-review.diff).

Proposal: Keep dynamic subclass receiver-origin safety. The stack detects any receiver return and checks descendants; the topic requires every return to be a receiver and introduces nominal convention fitting. Reconcile both proofs before changing iterator admission/refusal.

## landing-batch-3

Files: `internal/lower/expression.go`.

Both sides, hunk by hunk: [landing-batch-3.diff](landing-batch-3.diff).

Proposal: Do not import merged workers. The own reconciliation commits refer to absent Record IR/recordElement and discriminant/view fixtures. Dependency: records-lowering and checked-view lanes.

## optional-field-write-2

Files: `internal/javascript/javascript.go`, `internal/lower/invariance.go`, `internal/lower/lower_test.go`, `internal/lower/object.go`, `internal/lower/refusals.go`, `internal/lower/statements.go`, `internal/lower/taste_representation_test.go`, `internal/native/emit_objects.go`, `internal/native/reuse.go`, `internal/native/runtime/library_object.c`, `internal/native/runtime/object.c`, `internal/native/runtime/region.c`, `internal/oracle/counts.md`.

Both sides, hunk by hunk: [optional-field-write-2.diff](optional-field-write-2.diff).

Proposal: Keep existing checked-field runtime brands and view writes. Add optional presence/readiness/unknown guards only against a ruled shared object layout. Dependency: checked-view optional-write integration.

## error-classes-counts-2

Files: `internal/lower/class.go`, `internal/lower/class_inheritance.go`, `internal/lower/exceptions.go`, `internal/lower/expression.go`, `internal/lower/lower_test.go`, `internal/lower/narrowed.go`, `internal/lower/object.go`, `oracle/adamic.mjs`.

Both sides, hunk by hunk: [error-classes-counts-2.diff](error-classes-counts-2.diff).

Proposal: Keep accessor/private/static method lowering and every existing refusal. Add Error constructor/cause/stack rules and guarded property writes with equivalent checked-view/readiness/exception ownership. Requires class/exception lowering judgment.

## Dependency-only skip

`enum-tag-narrowing-2` (`64952252`) modifies absent `internal/lower/enum_tag_views.go` and relies on the checked-views merges in its ancestry. It stays outside this stack.

## Topic commit selection

Only non-merge topic commits were replayed, in chronological order. Explicitly authorized non-null and runtime-drop histories were merged. Paired integration merges/reverts are not imported, nor are inherited predicate/view/map-key lanes. The Uint16 new-expression experiment and its own withdrawal were both replayed; Uint16 remains refused.

## Dependencies exposed by group-three controls

`for-of-object` is skipped as a whole topic: `for_of_library_view.a:27:24` stops at `for...of over a union of differently held members`. The private Iterable view needs its origin and representation proofs reconciled with this base's dynamic `object`/unknown handling. The positive library-iterator control fails; no mutant failure is claimed as proof. Earlier object/tuple/generator controls passed, but were not delivered as a partial topic.

`representations` is skipped as a whole topic: `representation_object_view.a:9:27` retains the opaque host/collection-to-object refusal, and `representation_generic_wrapped_value.a:12:49` retains the uncheckable narrowed object-tag refusal. It needs the checked-view/object-tag lane. No new view code or refusal weakening is introduced to unblock those controls.

Element-access's original `paths` boundary remains correct after excluding representations; no diagnostic-only adjustment was retained.

## New-expression call-target judgment

File: `internal/lower/new_class_value.go`, `classConstructorValue`, lexical-cache initializer `ir.Call` case.

Stack side: `internal/ir/call_targets_guard_test.go:TestCallTargetReaders` prohibits unapproved `Call.Function` reads and requires analyses to use `CallTargets`/`ClosureTargets`.

Topic side: `initializer := l.result.Functions[value.Function]`; admits a zero-argument initializer when that one statically named function has no closure/environment. The required IR package fails at `new_class_value.go:159:39` with `unapproved call-target read ...; use CallTargets or ClosureTargets`.

Proposal: read `l.result.CallTargets(value)` once; retain the zero-argument requirement and admit the initializer only when the target set is known/nonempty and every possible target has no closure/environment. Preserve the additional-capture refusal otherwise. This may reject an initializer the topic's single-target check admitted, so it needs a ruling. Do not add an audit allowlist exception merely to green the test. Skip the entire new-expression topic until ruled, including its constructor-cache fixes and the experiment/withdrawal pair.

## Integration bookkeeping excluded

The statements topic's checked-views trial merge and its revert `1ed18fad0` were excluded together: they import/undo another worker's integration, not a statements lesson. Their net effect on the topic snapshot is zero. The five statements-only commits (`1b185548`, `f5987c61`, `564175fb`, `041239f6`, `742bada3`) were replayed in order. No merged views code or reverse patch removing this base's views behavior was imported. The for-of trial merge/revert pair would receive the same treatment, but that whole topic is held for its failing library-view control.

## Conflict evidence format

Readable `.diff` copies normalize tabs and trailing spaces solely for presentation. Every adjacent `.diff.gz` preserves the original combined diff byte for byte, including both conflict sides and all hunk context. These are review artifacts, not patches applied to the compiler.

## Requested additions and reorder

The first three groups were already pushed before this steering. The following topics remain held as whole topics; therefore no remote history was rewritten. Their requested positions are group 1 immediately after non-null for optional presence, group 1b for stricter options, and group 2 for the new destructuring topic.

### Optional presence, reordered group 1

Own commits, in order: 5bb775ca, 03980678, 86b3fe99, 739d6e10, 83edd7ec, 173164be, 7e7464e6. Trial at non-null tip 93ebcf52 reproduces semantic conflicts. The ruled runtime-drop commit changes unrelated runtime files and cannot supply these missing lowering contracts.

Every conflict hunk and both sides: [optional-presence-reordered.diff](optional-presence-reordered.diff), with original bytes in its adjacent gzip file. The previously listed optional-field-write judgment remains applicable. In object.go the base refuses spread-added fields; the topic admits contextually optional fields. Proposal: admit only represented optional slots after preserving the existing field-kind and readiness checks; keep other spreads refused. In invariance.go/refusals.go the incoming side imports optional-view and record checks absent from this area. Proposal: retain predicate refusals and current structural checks; require the record/view owner prerequisites rather than copying their implementation into this stack. In optional_write.go the base lacks the file and the topic modifies its earlier implementation: name that dependency instead of silently restoring an unlanded lowering. All remaining emitter/runtime layout hunks require preserving field-kind/readiness and ownership while introducing independent presence; the earlier full topic judgment documents the proposed integration. The topic is skipped pending those rulings and prerequisites.

### Stricter options, group 1b

Requested tip 8f32e51e is topic-only on 337aa466, but depends on the held presence topic. It is skipped as a whole dependency, without importing checked views or map-keys code. Its [own handoff](evidence/stricter-source-handoff.md) explicitly reports **106 scheduled / 67 remaining errors**, and says optional guard commits 58e743bf, 9af86993, 59ca7f85, 6db50e57, 4b1922bd, a022b1e1 and 8c81fd11 were omitted, with expression-cleanup dependency a593d04e also requiring reconciliation. Thus importing this ref plus presence alone does not establish the requested 173/0 result. Its historical [census log](evidence/stricter-source-census.log.gz) is source-worker evidence, not a new census on this delivery SHA. This branch has no stricter-options census implementation to rerun while the topic is held; no 173/0 claim is made.

### Namespace initialization, group 1

The exact original stricter-indexed-all base is aee98c837. Namespace-owned non-merge commits after it are f11380454, f8fea8a78, be7c88fc2, 4318e17b2, 5ad36d2cf, 162fd0812, efea3488a, d8387c33f. Main's intervening non-merge commits are excluded. The first implementation excludes ambient declarations from executable initialization. Its fixture registration can be combined by adding only the two new controls; its documentation can preserve the area text and append only its own host-namespace section. However its new lower tests require namespaceGraphForTest and SCC namespaceCallGraph infrastructure absent from this area. Those belong to the earlier held namespace reachability dependency. The topic remains skipped as a whole, rather than importing the stricter/view base. Exact host reductions also retain independent node:fs.native and node:process.Process.cwd NotYet; the source worker does not claim those programs run natively.

### Destructuring topic, group 2

Own commits in order: d2cf13cf3, 1f3c1894b, b988a5900, d516d4e1. Trial at group 2 tip d40016bd applies the computed-name commit cleanly, then conflicts in collections.go and object.go. Entire topic skipped; computed-name-only partial work is not delivered.

Every conflicting hunk and both sides: [destructuring-topic-group2.diff](destructuring-topic-group2.diff). In collections.go the base requires matching field/local representation and supplies View/ViewType/ViewAllowed metadata; the topic uses declaredStatements for nested/default binding initialization. Proposal: transfer all existing view and representation checks into the new initializer, retaining each failure boundary. In object.go the base refuses slotless union fields while the topic boxes unions and fits the declared union slot. Proposal: allow only proven boxed union storage with existing readiness and contextual field guards. Both proposals change lowering/admission and remain pending a ruling.
