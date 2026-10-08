# Area views next merge decisions

The branch includes 506441dd, d5c6292b and current candidate 4885cec50290686df487b62aac47c85d871ed40c, merged at d53b51f3. Protected work is on compiler/area-views-wip2 at 2df48668eb2165ff1785c8d977336364a6cb3ca0. Fifty new conflict hunks across fifteen files were resolved individually. Their exact sides and decisions are in area-views-next-merge-hunks.json. The original approved judgments 1 through 17 remain applied. No landing push has occurred.

## Approved judgments 18 through 23

| Number | File | Area behavior | Views behavior | Proposed resolution | Witness |
| --- | --- | --- | --- | --- | --- |
| 18 | internal/native/view_callables_boxing.go | Producer fixed, optional, rest and count argument layout. | Caller-sized boxed adaptation storage. | Allocate and populate the producer layout, preserving absence, rest and counts, typed code pointers and every callback refusal. | TestCheckedViewCallableNumericLiteral; later-ranked source-file-update; isolated numeric controls passed. |
| 19 | internal/lower/phantom_*_test.go | Phantom erasure compares whole IR. | FunctionTypeTargets holds checker IDs and brand relation facts. | Compare executable IR independently and assert the retained dispatch metadata without deleting or widening target sets. | The four failing phantom erase tests. |
| 20 | internal/lower/view_callables_boxing.go | Void result certificates use a wildcard sentinel. | Undefined has representation tag 13 and a nonzero mask. | Keep certified void results at the existing wildcard mask, preserving actual undefined tag 13 and all unknown-function refusals. | Checked-view performance family; isolated proposal passed. |
| 21 | internal/fresh/fresh.go | Unknown IR nodes conservatively refuse cycles. | ArrayRecord carries existing array identity and own property references. | Follow operands and stored references through the existing freshness and cycle proofs; do not exempt cycles. | ranked11-arrow-modifiers.a. |
| 22 | internal/native/taste.go | Logical operations retain a falsy absent reference. | Record is a distinct representation; conversion to String panics. | Use the existing falsy-reference proof only in the retained && branch to emit the absent String value. Keep || and ordinary conversion boundaries. | stage3/fixtures/taste/06_localized_message.a, native panic converting tag 14 to 3. |

Judgments 18 through 23 were approved and applied in 2df48668 and c6a8c98b. The earlier observations below describe prior checkpoints. Current evidence and pending judgments are recorded at the end.

## Changes beyond conflict resolution

81a891cd replaces both array-hole wasmtime runners with Node and oracle/wasi.mjs. It preserves fatal wasm build failures and the exact byte/exit assertions. Ten array fixtures and the control, stdout mutant and exit mutant passed in 20.387 seconds. The d5c6292b native optional-write refusal control passed in 0.118 seconds; removing only its guard failed with got <nil> in 0.132 seconds.

Nine candidate refusal records were restored individually where their exact outcomes matched the measured merge. Thirteen additional negative stage3 diagnostic records preserve already approved boundaries. The existing stage3 updater refreshed compiling records only after matching recorded/current Node output and native sanitizer/leak checks. No source fixture or Node observation was changed.

## Inherited baseline failures

The detached 506441dd comparison passed lower (51.514 seconds) and flow (213.435 seconds). IR passed (26.718 seconds) after disabling VCS stamping for the comparison worktree. Full oracle finished red: TestOmittedOriginalProbePolicy, the scanner frame in TestClosureMergeRefusals, and TestCountsAreRecorded. Stage3 reproduced stale Checker records for host/09_realpath.a, host/14_getCurrentDirectory.a and host/24_useCaseSensitiveFileNames.a. Those candidate records are retained and their inherited failures are named rather than repaired. The two commits from 506441dd to d5c6292b change only fuzz optional declaration and the native optional-write test, so this oracle/lower/flow/stage3 evidence is unaffected by those source changes.

## Current validation limits

Lower still fails four phantom metadata equality tests; IR passed in 2.124 seconds. The merged full oracle is still running and red. Counts refresh failed 101 leaf fixtures in 134.178 seconds and did not refresh the table. The new 2980-file .a audit is still running. Catalog passed its previous checkpoint only; it is not claimed green for this new landing candidate. No partial unit push is authorized or made.

## Additional array judgment

23. internal/lower/object.go elementType: the candidate compiles taste/07_compact.a, while the merged slotless-element guard rejects number[] | {flags:number} | undefined. Proposed resolution retains the ordinary boxed Union-array path only where no checked-view contract is erased, preserving views element-proof and unknown-array refusal boundaries. Applied as ruled, only without erasing a registered checked-view array element contract. taste/07_compact.a now agrees with Node and is recorded as Compiles.

## Stage3 path and record repair

The stage3 path validator now accepts the ruled .ts assertion fixture as well as .a, keeping portable paths and traversal rejection. Node/native agreement refreshed nested-functions/09_checker_constituent_recursion.ts. TestFixturePaths plus the complete nested-functions and objects groups passed in 8.472 seconds. One diagnostic-record edit used a brace-sensitive matcher and landed on objects/22_nested_optional_calls.a rather than objects/19_identifier_multimap.a. Both records were corrected with JSON-aware offsets and passed in that same rerun. No source fixture or Node observation changed.

Tag agreement passed in 0.079 seconds. The isolated Record=20 Go mutant failed with the exact C=14 versus Go=20 difference; all other tags remained unchanged.

## Retained optional-write expectations

fresh.a, fresh-boolean.a and fresh-undefined.a now require the replacement candidate's missing-own-field write NotYet boundary. Their exact Node observations remain checked and they stay registered as negative fixtures. All three passed in 0.584 seconds. Removing only absentOptionalWrite in an isolated overlay caused all three to fail with got <nil> in 0.334 seconds. This updates stale expectations under the explicit candidate refusal, without a production lowering change.

The final full stage3 run passed 618 test frames and failed only five fixture leaves: the three inherited host Checker records and the two held taste witnesses. The run completed in 34.212 seconds; neither taste expectation was weakened.

## Completed audit and proposal evidence

The .a audit completed all 2980 changed files in 1224.3 seconds: 2764 checked (including NotYet), 178 refused and 38 checker errors. Its expectation audit remains red on 216 files. Exact per-file diagnostics are in /tmp/area-views-next-a-check-results.json; no failures have been relabeled merely to make the audit green.

The full semantic catalog passed on 5634505954eff71e4f1961339af2c3bfbdbde0f7 in 588.849 seconds. All sixteen entries succeeded: eleven active undo patches were caught, and five historical entries were skipped for their recorded reasons. This is evidence on that checkpoint, not a claim that every check on the later landing candidate is green.

Held proposals 18 and 20 passed together in an isolated overlay against numeric callable, aggregate witness, aggregate parameter, watcher and callback-refusal controls in 59.696 seconds. Production source remains unchanged, pending the ruling.

## Completed full validation and further ruled expectations

The full oracle finished red in 2543.811 seconds: 4976 passing test frames, 598 failed frames and 52 failed top-level test families. This is the complete merged-source run, not a green result. Later targeted expectation repairs below are separately verified; no second full oracle result is claimed.

Counts refresh after the three optional-write expectation updates failed 98 leaf fixture controls in 57.675 seconds. The counts table remains provisional. Comparing all 216 .a audit mismatches with the detached candidate compiler reproduced 144 outcomes: 106 refused, 38 checker errors, 71 checked/NotYet and one other failure. None of these exact .a paths exists in the candidate; this measures compiler behavior on new fixtures and is not an inherited test failure claim. Per-file evidence is /tmp/area-views-next-a-check-baseline-comparison.json.

Ruling 10 expectation repairs pin all ten stored-marker probes plus marker/parser-cache.a, marker/discarded-required.a and marker/stored-call.a to their retained result-erasure or never-rest call boundary. Ruling 11 expectation repairs pin lazy/optional-error.a and lazy/optional-error-fields.a to the plain Error relation refusal. All source Node checks remain and independently certified immediate/lazy positive cases still execute all backends. The combined marker/lazy controls passed in 5.559 seconds. Isolated result-erasure mutants failed the eight stored result controls and the three additional marker controls; a stored-call guard mutant failed void and field-good; removing both Error relation guards failed both optional-Error controls. The first stored-call mutant had an unused local and was corrected before the semantic run; that build failure is not counted as a caught mutant.

Judgments 18 through 23 are now applied. No landing push has occurred.


## Current step 11 validation

The 4885cec5 merge resolved two conflicts individually. The constituent-recursion fixture keeps its views .ts path and takes the candidate origin ?? type proof. The scanner pin takes the candidate var refusal at 69:5 after its masked assertions were rewritten. The merge commit names the constituent fixture with a Moved-result trailer. c6a8c98b names the localized-message and compact stage0 transitions with Moved-result trailers.

TestCheckedViewCallableSourceFileUpdateBoxedOptional passed in 0.821 seconds. Removing producer padding and using caller-sized slots failed under ASan with stack-buffer-overflow. Merely changing allocation capacity did not catch the mutant and is not counted as evidence. The wider producer source-file-update fixture directly exercises the boxed adapter.

The four phantom-erasure controls passed. Clearing FunctionTypeTargets fails all four with missing function-type target evidence. The original metadata remains intact; direct-call and count checks use CallTargets. TestCallTargetReaders passed in 2.181 seconds after removing two unaudited direct Call.Function reads in the new test helper.

The certified void-mask removal mutant fails performance-measure/good and optional-values. TestCheckedViewArrayRecordCycleRefused and TestCheckedViewLogicalStringNullRefused passed together in 0.177 seconds. The contained-reference mutant and null-guard mutant each fail with got <nil> instead of the required refusal. The ordinary boxed-array removal mutant fails TestFixtures/taste/07_compact.a with the old NotYet boundary. The first native string mutant did not propagate into the stage3 oracle subprocess; that green result is not counted as a caught mutant. A forwarded GOFLAGS overlay rerun is recorded separately.

The configured 4885cec5 comparison passed lower in 86.788 seconds and flow in 136.817 seconds. The first comparison lacked its declared @types/node installation and is not inherited-failure evidence. The configured comparison reproduces three stale stage3 host records: 09_realpath.a, 14_getCurrentDirectory.a and 24_useCaseSensitiveFileNames.a. Those inherited records remain unchanged.

The merged counts refresh remains red on 97 lowering leaves and one additional fixture control. Eighty-six are cycle fixtures inconsistent with the approved judgment 1 refusal boundary; every one exits successfully on Node. The complete paths, exact refusal diagnostics and Node observations are in /tmp/area-views-cycle-node-observations.json. An automatic approval review rejected converting those executable/count expectations to named refusal expectations because it broadly changes validation. No changes from that rejected command were applied. Explicit approval for the 86 named expectation updates was requested. Cycle admissions remain refused.

TestOptionalWriteErasure/fresh now requires the candidate's missing optional-own-field write NotYet, retaining all other erasure controls. The targeted optional-write and phantom controls passed in 0.890 seconds. The merged full oracle and full stage3 checks are logged separately and are not claimed green. Counts, the .a audit and the final catalog remain incomplete or red.

## Pending lowering judgments 24 through 27

| Number | File | Two sides | Proposed resolution | Witness |
| --- | --- | --- | --- | --- |
| 24 | internal/lower/collections.go | The candidate admits ordinary Union-valued Map callbacks; the added views slotless guard refuses them. | Keep the existing boxed adapter for ordinary Map callbacks only where no checked-view value contract is erased. Keep every descriptor and unknown callback refusal. | host_map_union.a, host_map_generic_union.a, host_map_iterator_union.a; entry-live-mutation.a remains refused. |
| 25 | internal/lower/detached_own.go | The candidate admits known intrinsic alias .apply calls; the views guard admits only .call. | Retain the existing intrinsic adapter only for const aliases and literal one-key argument tuples. Keep escaping aliases and unknown signatures refused. | library_method_values.a and method_coverage_object_descriptors.a. |
| 26 | internal/lower/records.go | The candidate handles Object.values on a fixed object; views contextual storage conversion rejects the ArrayLike-or-Record contextual type. | Keep the existing fixed-object intrinsic path without converting to Record. Retain ordinary storage-cast refusals. | method_coverage_object_statics.a. |
| 27 | internal/lower/census_small.go and phantom_overload_results.go | The candidate evaluates extra direct-overload arguments ignored by the implementation; views rejects the call. | Keep the existing certified direct-overload path, including evaluation of extra arguments and all overload relation and callback proofs. | census_overload_contracts.a, including next() counts; TestCensusOverloadRelation. |

All four remain unapplied pending rulings. No landing push has occurred.

## 34c20abb merge

7302410ecfe4d54c6321fde785f67bf39fc5dcc4 is protected on compiler/area-views-wip3. The current candidate is 34c20abb6e0982c07df4394eb8e03c3a7aacbf63. One comment hunk in omitted_arguments_test.go was resolved individually in favor of the candidate's .ts probe. The candidate's TestOmittedOriginalProbePolicy now checks the unchanged Node stdout 11 against native, JavaScript, sanitizers and leaks. The .a refusal stays in TestNonNullAdamicRefusal and TestNonNullAssertionIsRefusedInAdamic. No production lowering or refusal changed. The candidate's two recorded count moves are retained.

TestNativeAgreesWithNode passed all seven pending-judgment witnesses on the configured 4885cec5 baseline in 2.691 seconds. Judgments 24 through 27 remain pending. The forwarded native AbsentString mutant failed TestFixtures/taste/06_localized_message.a/native with the original conversion panic. TestFixtures passed 625 frames and failed only the three inherited host Checker records. The merged full oracle is still running and red; counts, the .a audit and final catalog are not green.

## Boxed-callable declaration and package validation

internal/native/view_callables_boxing.go now declares memset through string.h in each generated adapter declaration. This retains judgment 18's producer-sized layout, missing optional slots, actual count and rest tail. The focused boxed-callable, original mixed callback ABI, nullable nominal array, nested payload and approved 21/22 controls passed 47 tests, 48 pass frames, in 36.708 seconds. The repeated caller-layout mutant fails TestCheckedViewCallableSourceFileUpdateBoxedOptional under ASan with stack-buffer-overflow in 2.444 seconds. No function-pointer cast or new callback admission was introduced.

The required post-34c20abb packages passed lower in 140.388 seconds, IR in 19.710 seconds and JavaScript in 4.339 seconds. Flow remains red in 214.382 seconds with 163 failed frames; 4885cec5's configured lower and flow comparison was green. These merged failures are not inherited package failures. The expanded map certificate check still refuses entry-nominal-optional-mutable-control.a at writing a possibly absent optional own field; that candidate refusal remains unchanged. The full oracle is still running and red. No landing push has occurred.

## Ruled callable counts and optional expectations

TestCheckedViewCallableCounts now requires the exact judgment 10 marker-result or never-rest call refusal and unchanged Node stdout before excluding each named negative from runtime counts. Nine obsolete stored-marker runtime rows were removed. All thirteen marker/stored-marker negative controls remain independently tested. The new boxed-optional row measures 11 allocations, 11 frees, 17 retains, 24 releases and peak 11. Six prefix-operand rows gained one measured retain/release each; allocations and frees did not change. The updater preserves the separate predicate section. Final lane refresh passed in 203.895 seconds and the read-only count control passed in 103.576 seconds. Removing the result guard fails on discarded-required with the wrong refusal boundary; removing the stored-call guard fails field-good with nil error. Neither mutant fails compilation.

optional_checked_test.go updates only exact optional-read diagnostics to the admitted nullable-union helper's existing matches-no-member text. The Boolean, String, null and numeric-literal sources, their exact Node observations, empty native stdout and required exit 70 remain unchanged. entry-nominal-optional-mutable-control.a now requires the candidate's precise optional-own-field write NotYet, with unchanged exact Node stdout. The optional suites and complete map certificate suite passed in 163.968 seconds. Removing absentOptionalWrite fails the map control with nil error in 0.126 seconds. The optional nullable-adapter mutant admits literal 1 and exits 0 with stdout 1, failing the required checked assertion; the other mutated scalar cases stop at the independent union-narrowing guard and are not claimed as wrong-result witnesses. The first object-primitive adapter mutant did not touch the nullable path and is not counted as proof.

The current compiler's a-check passed the three new approved fixtures: array-record-cycle and logical-string-null refused their exact named rules; boxed-optional checked. This is a three-file control, not a green claim for the full 2980-file audit. The new stage3 run passed 625 frames and reproduces only the same three inherited host Checker records in 50.789 seconds.

The full oracle started before the 34c20abb merge and later test repairs, and finished red in 1952.661 seconds: 5147 passing frames and 427 failed frames. It is not a completed validation of the later tree. The requested 4885cec5 comparison of additional failure families found only TestUnknownNarrowingMutants present; its two mutants passed in 0.558 seconds. Process and regexp-surrogate suites are absent from that base and their failures are not called inherited. Current unknown-narrowing skip-inner-typeof stops at an inserted check, while its stale mutant expectation demands only a stdout difference; no expectation change is applied to it here. Judgments 24 through 27 and explicit approval of the 86 named cycle expectations remain pending.

## Review artifacts on e3440508 plus f0c6e6fc

Judgments 24 through 27 remain proposals. The 86 individual cycle proposals and exact measured Node observations are in [area-views-next-cycle-updates.json](area-views-next-cycle-updates.json). No production lowering or cycle expectation update is made by these review artifacts. Baseline 4885cec5 evidence below is distinguished from the new f0c6e6fc comparison in area-views-next-failure-comparison.json.

### Judgment 24: internal/lower/collections.go

Area record: The configured 4885cec5 candidate lowered these ordinary Union-valued Map.forEach fixtures and native output agreed with Node. The candidate path uses the existing boxed callback adapter.

Views record: The added views slotless-value guard stops each fixture before callback lowering: stage 0 can't lower a Map forEach callback with a slotless value representation yet.

Proposal: Use the existing ordinary boxed Map callback adapter only when no checked-view value contract is erased. Preserve dynamic descriptor checks and unknown-signature/callback refusals, including entry-live-mutation.a.

Reason: Union storage alone does not erase a checked-view contract. The ordinary adapter already agrees with Node; its admission must remain separate from certified checked-view callbacks.

Fixtures and measured Node observations (exit 0, empty stderr):

- `internal/oracle/testdata/host_map_union.a`: stdout `"1 x undefined\none:0:number\ntext::string\nnew:new:string\nvalue:0:number\nvalue::string\nvalue:new:string\ncopy:one:0\ncopy:text:\ncopy:new:new\nvisit:one:0:9\nvisit:text::9\nvisit:new:new:9\nafter:one:9\nafter:text:9\nafter:new:9\nfalse undefined\nfalse:false:boolean\nmissing:undefined:undefined\nzero:0:number\ntext:x:string\ntrue undefined undefined\n"`.
- `internal/oracle/testdata/host_map_generic_union.a`: stdout `"true false undefined\n0 \none:0:number\ntwo::string\nvalue:0:number\nvalue::string\nvisit:one:0\nvisit:two:\nfirst:one:0\ntrue false undefined\n7 8\none:7:number\ntwo:8:number\nvalue:7:number\nvalue:8:number\nvisit:one:7\nvisit:two:8\nfirst:one:7\n"`.
- `internal/oracle/testdata/host_map_iterator_union.a`: stdout `"entry:one:1:number\nentry:text:x:string\nkey:one\nkey:text\nvalue:1:number\nvalue:x:string\nmap:one:1:number\nmap:text:x:string\nvisit:one:1:number\nvisit:text:x:string\nentry:one:0:number\nentry:text::string\nkey:one\nkey:text\nvalue:0:number\nvalue::string\nmap:one:0:number\nmap:text::string\nvisit:one:0:number\nvisit:text::string\nentry:zero:0:number\nentry:empty::string\nentry:false:false:boolean\nentry:missing:undefined:undefined\nkey:zero\nkey:empty\nkey:false\nkey:missing\nvalue:0:number\nvalue::string\nvalue:false:boolean\nvalue:undefined:undefined\nmap:zero:0:number\nmap:empty::string\nmap:false:false:boolean\nmap:missing:undefined:undefined\nvisit:zero:0:number\nvisit:empty::string\nvisit:false:false:boolean\nvisit:missing:undefined:undefined\nlive:one:0\nlive:text:\nlive:later:3\nnext:0\nremaining:\nremaining:3\n"`.

Required controls after approval: the named positive witnesses must agree with Node; existing negative boundaries must still refuse. No changed lowering is authorized by this ledger entry.

### Judgment 25: internal/lower/detached_own.go

Area record: The configured 4885cec5 candidate lowered const aliases of Object.prototype.hasOwnProperty called through .apply with a literal single-key tuple; native agreed with Node.

Views record: The views upfront detached-own guard permits .call only and refuses .apply: Adamic 0.1 refuses detached Object.prototype.hasOwnProperty; bind it to a const and use only .call(target, key), or use that call directly. The fixture failures occur at library_method_values.a:26:41 and method_coverage_object_descriptors.a:9:47.

Proposal: Retain the existing intrinsic adapter only for proven const aliases and literal one-key argument tuples. Keep receiver/key evaluation order, and refuse escaping aliases, unknown signatures and unsupported receiver shapes.

Reason: This preserves the independently proven intrinsic call path without admitting arbitrary function values or widening views calling-convention boundaries.

Fixtures and measured Node observations (exit 0, empty stderr):

- `internal/oracle/testdata/library_method_values.a`: stdout `"0 2 2\n2 0 true\n2|1\n1|2|1|NaN\nhello\nbcd\n[]\ntrue false true\ntrue false true\n[object Object] [object Array] [object Number] [object String] [object Boolean]\n12 255 5 1.25\n1|0|NaN|Infinity\ntrue|false\n10|NaN|2|3|4\n10|NaN|2\n1.25|2.5\n1 rsp\nhello  changed\nsaved\nsaved|saved\nSTRASSE\n1.3 1 ff 42\n4 5\nAB\n42\n0 0\n[object Array]\n1\n[object Null]\n3 3\n3|2|1\nx  \nx\ntrue false\n2 rs\nx|undefined|y\n"`.
- `internal/oracle/testdata/method_coverage_object_descriptors.a`: stdout `"true true false\ntrue true false\ntrue false false\nfalse false false\ntrue false true\ntrue true true\ntrue true true\nfalse false false\nfalse false false\nfalse false false\nfalse false false\nfalse false false\nfalse false false false\n"`.

Required controls after approval: the named positive witnesses must agree with Node; existing negative boundaries must still refuse. No changed lowering is authorized by this ledger entry.

### Judgment 26: internal/lower/records.go

Area record: The configured 4885cec5 candidate lowered Object.values(object) through the existing fixed-object intrinsic path; native agreed with Node.

Views record: Views contextual storage conversion stops method_coverage_object_statics.a:12:20: stage 0 can't lower a { '2': number; x: number; '1': number; } seen as ArrayLike<number> | { [s: string]: number; } (fixed objects and records have different storage; copy explicitly) yet.

Proposal: Retain the existing fixed-object Object.values intrinsic path without converting the object to Record. Keep every ordinary fixed-object/Record storage-cast refusal.

Reason: Object.values already consumes the fixed object directly and preserves Node key order. Avoiding an incidental contextual storage cast does not authorize that cast for ordinary storage.

Fixtures and measured Node observations (exit 0, empty stderr):

- `internal/oracle/testdata/method_coverage_object_statics.a`: stdout `"true false\ntrue false true\n1|2|x\n1|2|3\n1=1\n2=2\nx=3\n"`.

Required controls after approval: the named positive witnesses must agree with Node; existing negative boundaries must still refuse. No changed lowering is authorized by this ledger entry.

### Judgment 27: internal/lower/census_small.go; internal/lower/phantom_overload_results.go

Area record: The configured 4885cec5 candidate lowered the certified direct overload whose implementation ignores an extra actual argument; native agreed with Node. next() still runs for that ignored argument.

Views record: Views stops census_overload_contracts.a:13:16: stage 0 can't lower an overloaded call with more arguments than its implementation yet. TestCensusOverloadRelation also requires this certified path.

Proposal: Keep the existing certified direct-overload path and evaluate all actual arguments in source order before discarding extras. Preserve implementation/overload relation proofs and every callback/refused overload boundary.

Reason: The ignored argument has observable side effects. Node prints 2:2 after fixed('xx', next()), proving the second next() ran; dropping evaluation would silently change defaults from 2 to 1.

Fixtures and measured Node observations (exit 0, empty stderr):

- `internal/oracle/testdata/census_overload_contracts.a`: stdout `"3\nname\n2:2:1\n2:2\n5\n"`.

Required controls after approval: the named positive witnesses must agree with Node; existing negative boundaries must still refuse. No changed lowering is authorized by this ledger entry.
