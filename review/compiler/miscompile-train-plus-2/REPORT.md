Known survivors: historical-single-filter remains masked by its independent ABI filter; the exact original method-spread mutant remains masked by creation admission. Their adapted two-guard mutations are caught.
Built: merged fx6-cleanup onto train-plus-2, retaining rule 142 admission and turning p26/p33/p34/p35 into exact creation-refusal fixtures in both .a and .ts paths.
Commits: merge e6fb264e411585f6b7e6f8a1d53b078d2564da0a, parents 0c6b8384df2465dac8ee1dbdf7e88cd9dc47d196 and fb00e043a4a770f3041af6f5ea6ea34ed94d9174; the evidence/delivery tip is reported in the final response.
Checks: all 361 lowering tests in four shards, Node/backend fixtures, reader guard, vet, counts and committed-tip lane checks pass; 43 intended mutation obligations caught.
Limits: shared full gate not run; the two named exact survivors cannot be claimed killed. No over-refusal or unexplained count movement found.

This lands the requested train-plus-2 integration toward the views fix-forward roadmap work, including candidate 4 items 146/138 and cleanup diagnostic/dead-code work. The specified base and merge member take precedence over generic start-from-main and other-worker-branch rules. No PR or protected-branch push was made. No cohere code was copied.

The user's rule 142 ruling is applied literally: every member of a view type needs a supported, checked lowering. The four activated sources are moved unchanged from review/agree to review/refused. They are removed from the runtime acceptance/count registry. They are active refusal fixtures, not pending skips. Source Node still succeeds with stdout true+2 (p26), 1 (p33), 3 (p34), 3 (p35); that does not establish a checked lowering for their asserted view types.

| Fixture | Unsupported member | Exact creation location in both source extensions | Why no complete checked lowering exists |
|---|---|---|---|
| p26 callable_or_undefined | value | 7:14 | The callable-union builder requires every arm to have a producer signature; the optional undefined arm has no complete union adapter. |
| p33 map_in_union | value | 7:16 | The Map member lacks a checked union member contract. |
| p34 typed_array_in_union | value | 5:16 | The Uint8Array member lacks a checked union member contract. |
| p35 class_in_union | value | 6:16 | The nominal Point member lacks a checked union member contract. |

TestTrainPlusRefusalP26/P33/P34/P35 load the repository .a source and an ephemeral .ts copy with identical bytes. Both must return no IR program and a lower.Refused with the exact source path, location and What="view type has an unsupported member: value". Tests pin the creation site rather than a later demand. Assertions continue across both paths under mutants so both failures are recorded. No permanent .ts source was added. Guard implementations checkViewMembers and unsupportedViewContract are byte-for-byte the base's, and view still calls them first (resolved/admission-unchanged.json). No production admission guard was weakened.

No over-refusal was found. The old TestUnsupportedViewCallableMember used a fixed scalar callable union that candidate 4 now supports. Its original .a program is held to Node by the new TestSupportedCallableViewMember in JavaScript and native. The negative test still pins creation refusal but now uses the unsupported object-result callable union. The source fixture itself stays unchanged. Reverting scalar-union callable support makes the acceptance test fail on an unexpected creation refusal, proving that this control detects over-refusal.

p17, p18 and p20 remain acceptance witnesses. TestReviewProgramsAgreeWithNode holds them to source Node in JavaScript, release native and sanitized native, including successful-program leak checks; their leaves pass in 0.73s, 0.85s and 0.84s. The registered fixture suite also passes them. New lower witness files p18/p20/wide/p68/p69 and all callable/tuple inline controls pass both backends against Node; bad_callable remains a deliberate before-argument rejection with an observed Node control. The cleanup diagnostic sources remain correct creation refusals for values at 5:14. Their original demanded-read locations are independently held by readViewMember's AST-to-ViewWhere handoff, retaining a useful location guard without bypassing admission.

The merge's JavaScript conflict retains the train's loop-based recursive union traversal and optional-union tail handling, plus candidate 4's tuple-object WeakSet checks. The lower conflict keeps creation admission and recursive completeness checking, removing only the unused lazyReadRefusal. The train's scalar narrowing, snapshot representation and producer/result guards survive the clean overlaps. counts.md was resolved exclusively by one successful TestCountsAreRecorded -update-counts run on the merged implementation; no numeric row was chosen by hand.

Counts against 0c6b8384: seven added rows, zero changed rows, zero removed rows. All are freshly measured in the successful 54.902s regeneration. A/F/R/L/P/G is allocation/free/retain/release/peak/region.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| internal/lower/testdata/fx7_tuple_recognition/p69.a | 4 | 4 | 9 | 15 | 4 | 0 |
| internal/lower/testdata/fx7_wrong_aborts/wide.a | 5 | 5 | 6 | 11 | 5 | 0 |
| internal/oracle/testdata/review/agree/fxspptb_oct9_native_p17_callable_call_first_half.a | 11 | 11 | 14 | 19 | 9 | 0 |
| internal/oracle/testdata/review/agree/fxspptb_oct9_native_p18_cast_kind_no_field_types.a | 4 | 4 | 5 | 9 | 4 | 0 |
| internal/oracle/testdata/review/agree/fxspptb_oct9_native_p20_callable_member_call_good.a | 7 | 7 | 6 | 14 | 5 | 0 |
| internal/oracle/testdata/review/agree/fxspptb_oct9_views_p68_object_view_over_tuple.a | 4 | 4 | 8 | 14 | 4 | 0 |
| internal/oracle/testdata/review/agree/fxspptb_oct9_views_p69_tuple_view_over_tuple_union.a | 3 | 3 | 8 | 13 | 3 | 0 |

Attribution for every added row:

- p18, p20 and wide: registered by 56280072 after callable contract retention and selected producer argument adaptation; p18 also exercises the checked-cast metadata change from 36d90774. Correct checked calls execute and clean up instead of the former wrong abort. Wide includes a boxed scalar argument for the selected producer ABI.
- review views p68, review views p69 and extended lower p69: registered by 636b7397 after JavaScript recognizes represented object-backed tuples while rejecting arbitrary arrays. Their native tuple representation/counts do not change.
- p17: activated by ca88bf56 after the callable repairs. The escaped callable closures and runtime suffix strings produce the measured 11/11/14/19/9/0 row. Of that commit's five activations, only p17 remains runtime acceptance under rule 142.

Against the incoming fb00e043 table, four rows p26/p33/p34/p35 disappear because they are now compile-time refusals with no native counted executable. Two class-data rows were already removed by the train's rule-142 integration d0702c8d. Six retain deltas are inherited from d3d2e838: undefined-read-write 5->4, node-indicator-good 15->14, p54_n/b/s each minus one, and p04 10->8. That snapshot reader retains only reference tags, so undefined slots no longer issue adamic_retain(NULL); p04 reads two such slots. All other columns stay the same. The node-indicator and undefined deltas are already explicitly present in d3d2e838's counts diff; the scalar/p04 stale rows were measured and attributed by the preceding train-plus unit. They are unchanged from this unit's base. undefined-read-write.a remains 2/2/4/9/2/0 and is what the merged code measured. Full numeric comparisons are in resolved/count-delta.json. No unexplained movement occurred.

Toolchain setup succeeded with GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh. Timing lines: Node 0.021s, Go 0.023s, markdown dependencies 0.075s (validated installed bytes), submodules 0.091s, clang 0.242s, build 12.698s, test binaries deferred 12.839s, build cache warm 12.841s, done 12.866s. nproc=5, cpu.max=400000 100000. Go 1.27.1, Node 24.19.0 and clang 20.1.8 were used.

Exact verification commands and observed outputs (all test output saved directly to resolved/ logs; no test output piped):

```text
timeout 240 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 210s -args -update-counts
PASS 54.902s, exactly one regeneration (counts.log).
timeout 120 go test ./internal/lower -list '^Test'
361 top-level tests enumerated, no Example/Fuzz/TestMain entries. Four disjoint shards cover every name exactly once.
timeout 540 python3 review/compiler/miscompile-train-plus-2/resolved/run-lower-shards.py
Each shard: go test ./internal/lower -run '^<recorded name alternation>$' -count=1 -v -timeout 90s, child shell limit 120s.
All PASS: 91/90/90/90 tests; command wall seconds 10.346/16.162/10.490/10.546. Exact regexes and results are in lower-shards.json and lower-results.json.
timeout 240 go test ./internal/oracle -run 'TestCheckedView|TestRequiredView|TestNarrowedField|TestViewField|TestDefaultTagged|TestFX[67]|TestScalarUnion|TestCallableProducer|TestReviewProgramsAgreeWithNode/fxspptb_oct9_(views|native)_|TestReviewProgramsRefuse/fxspptb_oct9_native_p(26|33|34|35)_' -count=1 -v -timeout 210s
PASS 29.028s (oracle.log).
timeout 180 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/lower/testdata/(callable_producers|scalar_union_views|fx7_wrong_aborts|fx7_tuple_recognition)|TestNativeAgreesWithNode/internal/oracle/testdata/review/agree/fxspptb_oct9_(native_p(04|06|08|17|18|20|54|59|75|77)|views_p(68|69))' -count=1 -v -timeout 150s
PASS 7.492s (registered.log).
timeout 150 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
PASS 0.711s; after mutation restoration PASS 0.729s (readers.log, readers-final.log).
timeout 180 go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle ./internal/ir
PASS twice, empty output (vet.log, vet-final.log).
timeout 150 go test ./internal/lower ./internal/native -run 'TestTrainPlusRefusal|TestSupportedCallableViewMember|TestUnsupportedView|TestViewDiagnosticP|TestCallableContract|TestTupleRecognition|TestCheckedCastWithoutReadSummary' -count=1 -v -timeout 90s
Restored controls PASS: lower 1.757s, native 0.458s (restored-controls.log).
git diff --cached --check -- internal stage3
PASS. An earlier unrestricted whitespace scan flagged literal diff context in inherited mutant evidence; those patches were retained faithfully.
```

New/touched leaf observations from the full shards: TestTrainPlusRefusalP26 0.13s, P33 0.14s, P34 0.15s, P35 0.11s; TestSupportedCallableViewMember 0.70s. Each is a parallel top-level test, including both source-path loads where relevant. All new leaves meet the unit's grain limit. Other changed leaves and imported member leaves are recorded in the shard/control logs.

Mutation evidence was regenerated on this resolved merge. 43 intended obligations are caught; source edits are always restored in finally, with Go and outer command limits. No build failure is counted as a kill:

| Mutant group | Exact obligations and catchers |
|---|---|
| Receiver (3) | nullable-receiver: optional checked read escaped; generic-receiver: Node/backend exit differs; union-receiver: stage3 gap changed. |
| Valid (3) | revert-receiver: TestFX6P37 stdout; revert-destructure-type-id: TestFX6P53 stdout; revert-p70: TestTupleObjectViewRefused missing located refusal. |
| Routing (2) | element-bypass: TestCheckedViewElementP05; destructure-bypass: TestCheckedViewDestructuredUnion; checked-stop exits disappear. |
| Scalar (1) | skip-conversion-check: misfit number/boolean native and sanitized exit comparisons. |
| Callable (2) | unfiltered-native: registry test finds distinct pointer-type comparison; adapted direct-producer-certificate: certificate includes a direct function. |
| Assignable (5) | assignable-direct-certificate; exact-identity (literal-return stdout); skip-adapter-refusal (fewer params/method shorthand/extra optional); skip-result-registry-bound (discarded object result refusal); eager-tagged-callable (unread-method acceptance). |
| Name (3) | 147 name-wide undefined write: P04/P06 native stdout; 148 name-wide read: P75/P77 backend stdout; 149 name-wide null write: P08/P59 native stdout. |
| Member syntax (10) | final spread/in/keys/keys alias/values/entries and P19/P72: expected checked stops vanish; in-empty-selector: unrelated-field stdout check; adapted spread-method: missing own-slot refusal. |
| Supplemental (8) | Constructor revert: four original Node/native exit comparisons. Tuple reject/marker: P68 stdout. Admit arbitrary arrays: rejection disappears. Clear callable contract: pre-argument check fails. Source argument layout and missing producer mask: wide producer native stdout differs. Spread refusal: missing located refusal. |
| Isolated metadata (1) | Same metadata removal caught by TestCheckedCastWithoutReadSummary's native exit 70. Lowered p18 remains independently masked by read summaries, as recorded by candidate 4. |
| Cleanup location (1) | Blank ViewWhere fails P17/P48/P64 independent read-source-location assertions. |
| New refusal proofs (3) | Omit creation admission, wrong creation location, wrong unsupported member: all four TestTrainPlusRefusal tests fail on both .a and .ts paths for each mutation. |
| Supported callable control (1) | Revert fixed scalar union parameter support: TestSupportedCallableViewMember fails on an unexpected creation refusal. |

Commands: timeout 840 python3 .../run-members.py; timeout 600 python3 .../run-extra.py; timeout 300 python3 .../run-isolated.py; timeout 180 python3 .../run-location-mutant.py; timeout 150 python3 .../run-original-spread.py; timeout 480 python3 .../run-refusal-mutants.py. The final callable-support mutation uses the recorded .patch and /workspace/adamic-tools/go/bin/go test ./internal/lower -run '^TestSupportedCallableViewMember$' -count=1 -v -timeout 90s (subprocess limit 120s). Results are in resolved/fx6/, extra-mutants/, refusal-mutants/ and their runner logs. The array mutant's old .some context is adapted to the train's for-loop without changing mutation semantics. Certificate and spread obligations retain the previously recorded two-guard adaptations; the exact masked variants are also run explicitly.

Known survivors are preserved explicitly: historical-single-filter exit 0 because the independent ABI filter still removes direct functions, and exact original-spread-method exit 0 because creation admission rejects the unsupported method before the own-slot guard. Their original and adapted results are distinct, not relabeled kills. The lowered metadata experiment also passes, but the same source mutation is independently killed by its existing isolated IR test, so its obligation is proved. All 38 candidate-4 obligations, cleanup's location obligation, three new refusal checks and the new supported-control check are caught.

Committed-tip lane command:

```text
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

Executed under timeout 180 after merge commit e6fb264e411585f6b7e6f8a1d53b078d2564da0a. PASS: lane checks 7.1s, gofmt/tools on 69 Go files, t.Parallel on four test packages, a-check 13 .a files, vet four packages (resolved/lane-initial.log). Remote lane/tool refs were explicitly refreshed first because this checkout's normal refspec tracks only main. An earlier check following the evidence-whitespace interruption ran against the old HEAD and is retained as lane-base-only.log; it is not used as merged-tip evidence. The checker is rerun on the final evidence commit before the one delivery push, with its local output in resolved/lane-final.log. The shared full gate was not run.

The preceding blocked attempt is preserved separately in blocked-REPORT.md and the original logs/attempted-merge.patch. Those describe that attempt, not this delivered result. The ruling settled fixture classification; no contract implementation or weakened admission was introduced to make the train green.
