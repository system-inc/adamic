# Area views next merge decisions

The branch includes 506441dd and the replacement candidate d5c6292b. Fifty new conflict hunks across fifteen files were resolved individually. Their exact sides and decisions are in area-views-next-merge-hunks.json. The original approved judgments 1 through 17 remain applied. No landing push has occurred.

## Pending judgments

| Number | File | Area behavior | Views behavior | Proposed resolution | Witness |
| --- | --- | --- | --- | --- | --- |
| 18 | internal/native/view_callables_boxing.go | Producer fixed, optional, rest and count argument layout. | Caller-sized boxed adaptation storage. | Allocate and populate the producer layout, preserving absence, rest and counts, typed code pointers and every callback refusal. | TestCheckedViewCallableNumericLiteral; later-ranked source-file-update; isolated numeric controls passed. |
| 19 | internal/lower/phantom_*_test.go | Phantom erasure compares whole IR. | FunctionTypeTargets holds checker IDs and brand relation facts. | Compare executable IR independently and assert the retained dispatch metadata without deleting or widening target sets. | The four failing phantom erase tests. |
| 20 | internal/lower/view_callables_boxing.go | Void result certificates use a wildcard sentinel. | Undefined has representation tag 13 and a nonzero mask. | Keep certified void results at the existing wildcard mask, preserving actual undefined tag 13 and all unknown-function refusals. | Checked-view performance family; isolated proposal passed. |
| 21 | internal/fresh/fresh.go | Unknown IR nodes conservatively refuse cycles. | ArrayRecord carries existing array identity and own property references. | Follow operands and stored references through the existing freshness and cycle proofs; do not exempt cycles. | ranked11-arrow-modifiers.a. |
| 22 | internal/native/taste.go | Logical operations retain a falsy absent reference. | Record is a distinct representation; conversion to String panics. | Use the existing falsy-reference proof only in the retained && branch to emit the absent String value. Keep || and ordinary conversion boundaries. | stage3/fixtures/taste/06_localized_message.a, native panic converting tag 14 to 3. |

All five remain unapplied pending a ruling. The independently observed taste/07_compact.a NotYet regression is also unresolved; its positive expectation is not changed to hide the regression.

## Changes beyond conflict resolution

81a891cd replaces both array-hole wasmtime runners with Node and oracle/wasi.mjs. It preserves fatal wasm build failures and the exact byte/exit assertions. Ten array fixtures and the control, stdout mutant and exit mutant passed in 20.387 seconds. The d5c6292b native optional-write refusal control passed in 0.118 seconds; removing only its guard failed with got <nil> in 0.132 seconds.

Nine candidate refusal records were restored individually where their exact outcomes matched the measured merge. Thirteen additional negative stage3 diagnostic records preserve already approved boundaries. The existing stage3 updater refreshed compiling records only after matching recorded/current Node output and native sanitizer/leak checks. No source fixture or Node observation was changed.

## Inherited baseline failures

The detached 506441dd comparison passed lower (51.514 seconds) and flow (213.435 seconds). IR passed (26.718 seconds) after disabling VCS stamping for the comparison worktree. Full oracle finished red: TestOmittedOriginalProbePolicy, the scanner frame in TestClosureMergeRefusals, and TestCountsAreRecorded. Stage3 reproduced stale Checker records for host/09_realpath.a, host/14_getCurrentDirectory.a and host/24_useCaseSensitiveFileNames.a. Those candidate records are retained and their inherited failures are named rather than repaired. The two commits from 506441dd to d5c6292b change only fuzz optional declaration and the native optional-write test, so this oracle/lower/flow/stage3 evidence is unaffected by those source changes.

## Current validation limits

Lower still fails four phantom metadata equality tests; IR passed in 2.124 seconds. The merged full oracle is still running and red. Counts refresh failed 101 leaf fixtures in 134.178 seconds and did not refresh the table. The new 2980-file .a audit is still running. Catalog passed its previous checkpoint only; it is not claimed green for this new landing candidate. No partial unit push is authorized or made.

## Additional array judgment

23. internal/lower/object.go elementType: the candidate compiles taste/07_compact.a, while the merged slotless-element guard rejects number[] | {flags:number} | undefined. Proposed resolution retains the ordinary boxed Union-array path only where no checked-view contract is erased, preserving views element-proof and unknown-array refusal boundaries. This remains unapplied pending ruling; the positive expectation stays unchanged.

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

Judgments 18 through 23 remain unapplied. No landing push has occurred.
