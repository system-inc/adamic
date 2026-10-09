Finding and build: the exact original spread-method mutant survives through masking by creation admission; the train includes both requested fixes with both guards preserved.
Commits: train 2a570424; depth merge 81692df9; callable merge and single counts regeneration 4083680c; interaction expectations 1e5c2646; final evidence SHA is reported with the push.
Commands and outputs: full lower PASS 62.156s, scoped oracle PASS 48.873s, readers PASS 1.712s, lane PASS 11.2s; exact bounded commands and logs are here.
Mutants: all 29 adapted fx6 obligations, ten depth/text obligations, constructor, admission, snapshot and callable-result obligations are caught; the exact original spread-method revert is separately recorded as masked.
Not covered: the whole repository gate, full native/oracle packages, other platforms, or new unsupported-family admission. Integration owns the gate.

The branch starts exactly at the requested train and merges 2fa2a97f then a5e685cf. Only counts.md conflicts. Both count inputs were retained mechanically with merge-file --union pending a single TestCountsAreRecorded -update-counts run on the fully merged compiler. That run passed in 204.290s. The final table contains no duplicated input row. No production code was edited beyond the member merges. No cohere implementation was copied. The explicit train brief takes precedence over the general current-main recipe, and no newer main was merged. This advances #aecx10a and #1fk58py, including items 140, 151, 152 and 154, with the depth guard toward step 24.

The original spread-method-bypass.diff only removes the own-slot check. On this train, checkViewMembers independently refuses its method signature at creation, so TestViewSpreadMethodRefused stays green: this is an observed survivor, not a claimed kill. The adapted obligation removes both checks and fails with got <nil>. The unsupported-view mutant independently proves creation admission. Likewise, p32/p33 no longer reach the legacy name-wide write check: receiver certificates allow the ordinary producer writes, then checked reads stop on number instead of string. Their pinned diagnostics now describe those reads. TestFX7DeclaredWriteText independently activates the retained legacy CheckedFields name entry on a lowered temporary program and pins the write's declared string | number text in JavaScript, release native and sanitized native. Both original diagnostic source mutations fail that control. This is an emitter boundary control, not a new admission rule.

The unsupported-view saved revert failed to apply at its historical context. The rerun removes the same admission block and its two helper functions from the current sources. Snapshot reruns apply scoped changes to current files rather than overwriting the train with historical complete source snapshots. Every mutation restores its exact original source in finally. There are no remaining production mutations. The new test leaf was 0.42s in isolation and 0.78s in the final fixture suite; touched p32/p33 were 0.95s/1.33s in that suite. The method-spread leaf was 0.03s in isolation. No new persistent fixture row was introduced by the temporary emitter control.

undefined-read-write.a is allocations/frees/retains/releases/peak/regions 2/2/4/9/2/0. Commit d3d2e838 changes retains 5 to 4: readUnionView now retains only string-through-function reference tags and produces NULL for undefined. Its one former adamic_retain(NULL) call disappears; count.h counts calls even on NULL. Direct builds at the original train and the merged compiler produce 4. Restoring only the former reference reader produces 5. This value is measured behavior, not a chosen conflict side.

The regenerated table also repairs six stale train rows. This was reported as a finding before attribution. Direct counted builds at 2a570424 already produce every regenerated value, and generated C for these six witnesses and undefined-read-write.a is byte-identical at the train and merged compiler. Empty .diff files and the full C/count receipts are in count-attribution/.

| Existing row | Train table A/F/R/L/P/G | Actual train and merged A/F/R/L/P/G | Isolated cause |
|---|---|---|---|
| fxspptb_oct9_native_p04_undefined_slot_write.a | 5/5/10/17/3/0 | 5/5/8/17/3/0 | d3d2e838 removes two null snapshot retains; restoring its old reader restores 10. |
| p54_n.a | 5/5/6/14/4/0 | 5/5/5/14/4/0 | d3d2e838 removes one null snapshot retain; counterfactual restores 6. |
| p54_b.a | 3/3/6/12/2/0 | 3/3/5/12/2/0 | Same guard; counterfactual restores 6. |
| p54_s.a | 4/4/9/14/3/0 | 4/4/8/14/3/0 | Same guard; counterfactual restores 9. |
| snapshot-maybe-boolean.a | 6/6/9/24/2/0 | 9/9/9/27/3/0 | 239a3a8f checked spread returns an owned helper result, so each of three calls copies rather than reuses source. |
| snapshot-maybe-number.a | 10/10/9/21/4/0 | 13/13/9/24/5/0 | Same three copies; bypassing only checked spread restores the table's original 10/10/9/21/4/0. |

For both snapshot rows, bypassing only the checked-spread boundary restores their isolated-branch counts exactly. There are no remaining unattributed count movements. Thirteen new rows cover the nine depth/text fixtures and four callable-result fixtures. p32/p33 record releases 11 rather than their isolated depth branch's 9: ordinary raw writes complete their cleanup before the guarded read stops, preserving both merged receiver boundaries.

The final scoped oracle includes checked-view fixtures, fx6/fx7 controls, scalar and callable producer witnesses, snapshot and callable-union results, and review agreement/refusal witnesses. The separate registered fixture run passed in 81.198s; stage3 assertions passed in 20.725s. View runtime controls passed in native and JavaScript. All four callable results match source Node; native includes release and sanitizers. Wrong-view witnesses use the existing ruled exit-70 and pinned text contracts, and unsupported fixtures remain compile-time refusals; these intentionally are not assertions that unsafe source Node and admitted native should agree. The complete member .a census is member-fixtures.json. The additional saved p19 invocation probe now prints 6 on source Node, JavaScript, release native and sanitized native; p70 source Node prints 2 and Adamic retains its explicit tuple-view NotYet.

Every fx6 obligation and its catcher is listed below; individual logs and generated mutation diffs are in fx6/.

| Mutant | Observed catcher |
|---|---|
| nullable-receiver | Optional read boundary and optional-receiver assertions |
| generic-receiver | Generic wrong-view exit comparison |
| union-receiver | Stage3 identifier-kind status assertion |
| revert-receiver | TestFX6P37 JavaScript stdout |
| revert-destructure-type-id | TestFX6P53 JavaScript stdout |
| revert-p70 | TestTupleObjectViewRefused |
| element-bypass | TestCheckedViewElementP05 exit comparison |
| destructure-bypass | TestCheckedViewDestructuredUnion exit comparison |
| skip-conversion-check | Both scalar misfit native and sanitized exit comparisons |
| unfiltered-native | Producer registry emitted-C assertion about distinct pointer types, not a clang kill |
| direct-producer-certificate, adapted | Closure-thunk certificate assertion, both redundant filters removed |
| assignable-direct-certificate | Same certificate assertion as a separate recorded member obligation |
| exact-identity | Literal-return JavaScript stdout |
| skip-adapter-refusal | FewerParameters/MethodShorthand/ExtraOptional refusal assertions |
| skip-result-registry-bound | DiscardedObjectResult refusal assertion |
| eager-tagged-callable | TaggedUnreadMethod acceptance assertion |
| 147-name-wide-undefined-write | TestFX7P04/P06 native stdout |
| 148-name-wide-read | TestFX7P75/P77 backend stdout |
| 149-name-wide-null-write | TestFX7P08/P59 native stdout |
| final-spread | TestCheckedViewSpread runtime exit |
| final-in | TestCheckedViewIn runtime exit |
| final-keys | TestCheckedViewKeys runtime exit |
| final-keys-alias | TestCheckedViewKeysAlias runtime exit |
| final-values | TestCheckedViewValues runtime exit |
| final-entries | TestCheckedViewEntries runtime exit |
| p19-read | TestCheckedViewElementP19Read runtime exit |
| p72-input | TestCheckedViewElementP72 runtime exit; input control |
| in-empty-selector | TestViewInEmptyKeyControl JavaScript stdout |
| spread-method, adapted | TestViewSpreadMethodRefused with both admission and own-slot guards removed |

Ten depth/text source mutants are separately caught: native-depth-128 and javascript-depth-128 by Depth65; native-stack-guard by the forced stack control; union-found-kind by TextP67/P72; scalar-found-kind by TextScalarTuple; native-declared-name and javascript-declared-name by DeclaredWriteText; javascript-optional-frames by Depth2000; native-cycle and javascript-cycle by their respective cycle controls. The runner rejects build failures as kills. Exact receipts are depth-mutants.log and depth/.

Extra receipts are in extra/results.json: constructor-revert fails the four constructor Node/native exit comparisons; unsupported-view-revert fails the creation/refusal expectations; unsupported-stage3-revert fails the interface-kind status assertion; snapshot-runtime-mutant and snapshot-compiler-mutant fail Node/native behavior with sanitizer evidence; snapshot-reference-mutant fails only the fallback payload probe's exit comparison while preserving its pinned diagnostic. Runtime-alone and compiler-alone remain green controls. Callable zero and six reference-slot mutations run inside recordNativeAgreementFailure: their enclosing tests pass because the mutations fail the intended Node comparison. Zero is a stdout mismatch, undefined versus 0; six is a native stdout/exit mismatch, empty versus 6. Neither is a compiler warning or link-failure kill.

Setup used the required GOPROXY and sourced /workspace/adamic-tools/env.sh. nproc=5, cgroup quota four CPUs. Initial setup reported Go 0.319s, Node 0.430s, clang 1.130s, markdown 2.196s and submodules 21.230s, then its build census overlapped merges and failed with missing lowering methods and DeclaredType. Stable rerun passed: Go 0.186s, Node 0.169s, clang 0.914s, markdown 0.367s, submodules 0.399s, build cache warm 80.221s, done 80.390s. Pinned stage3/api npm dependencies were installed. The isolated baseline initially failed VCS stamping; -buildvcs=false built it successfully without changing compiler behavior. All outputs went to logs, never through a pipeline.

Integration lane output: lane checks 11.2 s: gofmt and tools on 51 Go files, t.Parallel on 4 test packages; a-check 13 .a files; vet 4 packages. Supplemental bounded go vet on ir/lower/native/javascript/oracle passed with no diagnostics. git diff --check passed. No PR was opened. The final evidence-only commit receives the same lane check before the single finished-unit push to compiler/miscompile-train-plus.
