Built a topic-only branch on main 6998ebc24ae353193cb1495d3d51308131a4b5c7, retaining only my eight non-merge commits and adding morning union coverage and a dynamic-length refusal control.
Commits carried: 7fcfd775, d7be3de1, 659cac36, c1464903, 20f2087f, 541255f2, 295b8027, 7753bb27; the final topic adaptation is the branch tip.
Checks: 17 topic fixture cases, focused lower/native/flow/IR checks, JavaScript package build and required counts refresh pass; no whole package or full gate run.
Mutants: all 22 targeted mutants fail fixture assertions or sanitizer/oracle comparisons; each restored; runner and measured results accompany this report.
Uncovered: six dynamic/compound array lengths need a checked no-holes proof; branded Path needs checked views; any remains a ruling refusal; runtime scalar views await 17b5a053 on main.

This section supersedes the historical report below for codex/notyet-class-set-property-topic. The branch has no merge commits above main and contains no runtime-owner commit, compiler-area merge, checked non-null merge or census-tool commit. Main was fetched before creating this branch. Toolchain setup from the earlier session remains installed: setup total 880.607s; nproc 5; every Go invocation sourced /workspace/adamic-tools/env.sh (GOPROXY=https://proxy.golang.org|direct).

The morning input is codex/stage3-notyet-table e8c283b5, stage3/notyet-table/rerun-0730/after/roots.csv: 38 sites in nine relevant kinds. Counts below associate sites with a rule, not a claim that every site lowered. Twelve representative sites were replayed with the 9a1f14c5 replay tool, built as an untracked scratch overlay against this branch and removed afterward. No worker history or measurement hook lands in production. Replay returns 0 when the old finding still matches, 1 when it is absent. Measurement is on the same checker-rejected adapted entry project as before and does not prove whole-compiler compilation. Failed statement boundaries can hide later sites.

| Morning kind | Census sites | Result and limit |
| --- | ---: | --- |
| replacing a represented method at runtime | 17 | Lowered literal and structural own-slot replacement. Inherited/class/library replacements remain refused or guarded. core:1544/1545 are echoes of the earlier uncheckable cast at 1543; sys:1382 reaches a function value taking boolean/undefined at 1382:34. No claim that all 17 complete. |
| assigning a field of a value | 11 | Five clearing/decrement forms covered (checker:1974, core:312, sourcemap:316, tracing:164/171). Six dynamic or compound forms skipped pending a checked no-holes proof (checker:8129/8132, core:307, emitNode:308, parser:2283, sys:234). core:307 still reproduces; tracing replay encounters reading Debug first. |
| storing true / Node / undefined | 3 | Lowered boxed owned fields. parser:1341 has no selected-unit findings. utilities:8987 reaches the RHS binary at 8987:48 and direct case declaration at 8998:13. |
| storing false / Type | 2 | Lowered boxed boolean/object fields; checker replay reaches binaries at 15220:20 and 15237:17. |
| storing Path / undefined | 1 | Skipped branded-string checked views. builder:668 is hidden inside the failed statement at 642, a value-as-condition refusal; its missing old signature is not evidence that Path lowered. |
| storing false / Symbol | 1 | Lowered storage rule, covered by SymbolRecord reduction. checker:34075 is behind failed bindings, including reading links at 34074. |
| storing string / number | 1 | Lowered owned scalar boxes; emitNode:242 has no selected-unit findings. |
| storing false / string[] / undefined | 1 | Lowered boxed array ownership and checked narrowed tag, covered by the new morning fixture. moduleNameResolver:2277 reaches the RHS binary at 2277:59; earlier optional-boolean/structural-call stops remain. |
| storing any | 1 | Refused for a ruling; tsbuildPublic:2080 reproduces storing any in a field. |

The new morning fixture is held to Node in JavaScript, sanitized native (ASan/UBSan), release native and ownership/leak checks. Seventeen registered topic cases pass: thirteen lower (three intentional guard-firing controls compare the two backends), four pin compile-time refusals. The new dynamic-array refusal has a separate Node control. Existing narrowed-union and inherited-method guard fixtures remain checks of deliberate refusals, not proof of JavaScript acceptance of those invalidating aliases.

Runtime dependency: runtime/shape-field-kinds 17b5a053 is not an ancestor of main, so its owner commit was not carried. The separate class_set_property_union.c retains adamic_union_slot_owner; reference slots retain through shape->references, generated scalar layouts use exact emitted shape identity, and unknown runtime scalar layouts panic with the dependency name. No name matching exists. RegExp/stat/iterator scalar-view source fixtures remain committed but their registrations are deferred, the counted stat test skips with the dependency name, and their count rows are removed. Re-enable these only after shape kinds land on main; the wrong-kind debug mutant is deferred with that dependency. Separate runtime helpers requiring owner review: adamic_union_slot_owner and adamic_union_runtime_field in internal/native/runtime/class_set_property_union.c.

Minimal compatibility hooks replace the compiler-area slotless helper locally: boxed union literal initialization, field reads and setProperty stores are allowed; MaybeBoolean slots remain refused. Typed-array tag support is deferred because main has no typed-array IR. The counts refresh changes only topic rows and library_string_raw, whose retains/releases move 50/84 to 54/88 through the restored owned union-read hook; allocations/frees remain 45/45.

Commands and observations (all test output written to logs):

- go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/class_set_property' -count=1 -v: PASS, 17 cases, 6.531s; /tmp/adamic-topic-final-oracle-fixtures.log.
- go test ./internal/oracle -run '^TestClassSetProperty' -count=1: PASS, 1.389s; /tmp/adamic-topic-final-boundaries.log (shape stat dependency skipped).
- go test ./internal/lower ./internal/native ./internal/flow ./internal/ir ./internal/javascript -run '^(TestClassSetProperty|TestUnionRuntimeFieldGuards|TestCallTargetReaders|TestCallTargetsIncludeEveryDescendant|TestClosureTargetsBoundOnlyProvenValues)' -count=1: PASS; lower 0.271s, native 1.113s, flow 1.797s, IR 8.279s; JavaScript has no Go test files and is exercised by the oracle; /tmp/adamic-topic-final-packages.log.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: PASS, 45.088s; /tmp/adamic-topic-counts.log.
- python3 /tmp/adamic-topic-mutants.py and /tmp/adamic-topic-boundary-mutants.py: 22 fixture failures, no build-failure substitutes; combined reproducible runner topic-mutants.py and results topic-mutants.json.
- Scratch replay overlay: python3 .topic-census-tool/make_overlay.py /workspace/adamic /tmp/adamic-topic-replay-overlay; go build -buildvcs=false -overlay=/tmp/adamic-topic-replay-overlay/overlay.json -o /tmp/adamic-topic-replay ./.topic-census-tool/replay/worker. Twelve runs recorded in topic-replays.json; command list in topic-replays.py. Tool scratch files removed before commit.

| Mutant | What caught it |
| --- | --- |
| structural-slot | structural_method cannot lower (exit 1) |
| own-method-native | inherited_method backend guard disagreement (exit 1) |
| own-method-javascript | inherited_method backend guard disagreement (exit 1) |
| literal-write | literal_method cannot lower (exit 1) |
| array-clear | array_length Node output disagreement (exit 1) |
| array-decrement | array_length Node output disagreement (exit 1) |
| array-empty-range | array_length RangeError/output disagreement (exit 1) |
| array-receiver-once | array_length extra receiver evaluation (exit 1) |
| union-storage | union_morning cannot lower (exit 1) |
| union-literal-fit | union_morning UBSan invalid boolean and oracle disagreement (exit 1) |
| union-write-fit | union_morning ASan heap-buffer-overflow (exit 1) |
| union-narrow-check | union_narrow missing panic/backend disagreement (exit 1) |
| union-array-tag | union_morning wrong object tag panic (exit 1) |
| runtime-retain | union ownership/sanitizer disagreement (exit 1) |
| runtime-scalar-stop | TestUnionRuntimeFieldGuards requires named scalar panic 70 (exit 1) |
| runtime-foreign-slot | TestUnionRuntimeFieldGuards requires named foreign-slot panic 70 (exit 1) |
| undefined-write | undefined_frozen missing frozen-write panic (exit 1) |
| intersection-agreement | TestClassSetPropertyIntersectionSlotAgreement rejects mismatched slots (exit 1) |
| accessor-own-method | accessor_method_guard backend guard disagreement (exit 1) |
| shape-type | union_views number/boolean layout conflation (exit 1) |
| dynamic-length-accepted | TestClassSetPropertyDynamicLengthBoundary requires NotYet (exit 1) |
| any-accepted | TestClassSetPropertyAnyBoundary requires storing any NotYet (exit 1) |

Historical report for the former compiler-area-based branch follows; its merges and proofs are not part of the topic branch dependency claim.

Built own-slot method replacement, owned mixed-union fields, array length removal, consistently held intersections and undefined fields.
Commits: 08b76587, 5475f479, 863babed, ecfead29, d0464e43; main landing merge 4249054a; requested non-null merge b6415dda includes c41c0e06.
Checks: Node, JavaScript, native ASan/UBSan and release, ownership/leaks, focused lower/native/flow/IR checks pass; counts refreshed; no full package or full gate run.
Mutants: all 37 targeted mutants fail their controls; each restored; individual catches and logs below.
Uncovered: any needs a ruling (2 roots); 3 census echoes cancelled; keyof replay stops at Path before its assignment; inherited method expandos and unrepresented runtime scalar views remain guarded.

Compiler base: b410340dc8f889b5799c3bc519117c63def3aa24, newest origin/area/compiler resolved at setup. Census replay 9a1f14c5d994aa855625e7cfa295677060348fec was merged at f943bdf03fbe492b65731714a7437a7980db228c. User-requested c41c0e062e99da37820f822968d4df1b48cdaee7 was merged without conflicts at b6415ddab8eeca895d29284d13e1da332fac3ecb; pending work was stashed and restored. All 14 representative examples were replayed again after that merge. replay-final.json contains the final compact findings, including the later intersection replay.

The adapted census input matches all 81 source hashes in stage3/meter/runs/20261008T035244Z.latent-full/tsc/source-manifest.json. Replay observes a checker-rejected entry-root program, and does not prove whole-compiler compilation. A replay exit of 1 means its requested stop was absent; the findings name the remaining stops. The 27 counts below are the table's covered kind counts, not a claim that all 27 sites were independently executed.

| Kind | Roots | Result and representative next stop |
| --- | ---: | --- |
| Replacing a represented method at runtime | 14 | Lowered literal and structural own slots. sys.ts:1382 reaches structural call 1387:49. core.ts:1544/1545 are cancelled echoes of an unchecked Map cast at 1543:17; next `reading map`. Inherited/class/library methods remain refused or checked for a ruling. |
| true / Node / undefined | 3 | Lowered. parser.ts:1341 has no selected-unit findings; utilities.ts retains binary 8987:48 and direct case declaration 8998:13. |
| any | 2 | Refused for a ruling. tsbuildPublic.ts:2080 retains `storing any in a field`; 2068 is preceded by any at 2065 and structural call 2068:44. |
| Assigning a field of a value | 1 | Lowered array `.length--` and literal `.length = 0`; tracing.ts:164 retains dependency unknown at debug.ts:213:28. |
| VariableDeclaration intersected with named Identifier | 1 | Lowered consistently held reference slots and optional scalar pairs. JSX retains structural call 113:29 and unchecked cast 114:48. |
| false / Type | 1 | Lowered; next binaries checker.ts:15220:20 and 15237:17. |
| false / VersionPaths / undefined | 1 | Lowered; next binaries moduleNameResolver.ts:2409:49 and 2411:12. |
| keyof CompilerOptions / undefined | 1 | Lowered mixed string/numeric key fixture; actual site replay is blocked earlier by Path in builder.ts:1502:71. No claim that 1509 was reached. |
| string / number | 1 | Lowered; emitNode.ts:242 has no selected-unit findings. |
| string / number / PseudoBigInt | 1 | Cancelled census echo of unchecked cast checker.ts:20277:22; next `reading type` at 20278. General storage rule is held to Node by the mixed-union fixture. |
| undefined | 1 | Lowered; nodeFactory.ts:2357 retains readonly TextRange upcast refusals in utilities.ts:10655 and 10645. |

Method writes evaluate receiver then RHS exactly once. A literal method already has an owned data slot. Structural signatures require an existing own slot in both backends; replacing an inherited method would add an expando, which the fixed-shape doctrine refuses. A minimal finishAccessors hook preserves this guard through setter dispatch. Source Node deliberately proceeds in the checked inherited-method controls; both emitted backends stop with the same named guard.

Mixed-union fields hold owned boxes using their declared storage type. Narrowed reads hold once and check the current member before casting, including object/array/map/typed-array kinds. Readonly widened views consult the actual slot owner's representation; shape identity includes field types, rather than reference bits alone. Class spread public layouts, inherited constructor static storage and runtime RegExp scalars are covered. Missing optional fields and optional receivers remain distinct. Unknown runtime scalar layouts stop with a named guard instead of interpreting scalar bits as heap pointers.

Plain object intersections admit scalar fields, consistently held reference fields and optional scalar pairs. Unused unrepresented phantom members do not prevent carrying an object pointer; accessing them still needs a representation proof. Boxed-union-to-object refinements remain stopped. The first constituent-agreement mutant survived because an earlier expression stop masked it; a direct helper check against the same Node fixture then killed it. That failed initial probe is not counted as a caught mutant.

Exactly-undefined fields keep an ordinary slot containing the existing null representation. Undefined assignments preserve property effects, including the frozen-object TypeError. Array length removal uses existing ArrayPop to release removed elements, preserves aliases, and throws a catchable RangeError with `Invalid array length` on decrementing an empty array. Growth, holes and general array property assignment remain stopped.

Setup: GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh completed in 880.607s. Timing lines: Go .248s, Node .782s, clang 1.830s, markdown 5.354s, ready 6.252s, submodule 57.013s, build 880.305s, tests deferred 880.522s, cache 880.526s. nproc=5; CPU quota 400000/100000. Environment: /workspace/adamic-tools/env.sh; log /tmp/adamic-class-set-property-setup.log.

Final focused commands, with output redirected to /tmp/adamic-class-set-property-final-<package>.log:

```
go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/class_set_property_|^TestClassSetProperty|^TestCheckedNonNullTypeScript$' -count=1
go test ./internal/lower -run '^(TestClassSetProperty|TestClassFeatures|TestOptionalWidening|TestObjectRefusalsExplainSoundness|TestObjectUnprovenShapesStayNotYet|TestCheckedCastProofAndElision|TestNonNullAssertion)' -count=1
go test ./internal/native -run '^(TestUnionRuntimeFieldGuards|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestInheritanceMemoryPlans)' -count=1
go test ./internal/flow -run '^TestClassSetPropertyFlow$|^TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/class_set_property_|^Test(LivenessHoldsOnEveryPath|EveryMutationIsInItsRange)/programs/../oracle/testdata/class_set_property_' -count=1
go test ./internal/ir -run '^Test(CallTargetsIncludeEveryDescendant|ClosureTargetsBoundOnlyProvenValues|CallTargetReaders)$' -count=1
go test ./internal/javascript -run '^TestClassSetProperty' -count=1
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
```

Outputs before landing: oracle 14.965s, lower 10.870s, native 3.004s, flow 2.599s, IR 44.713s; JavaScript has no standalone tests and is exercised by the oracle. The first flow selector skipped nested liveness/mutation subtests, triggering their nonvacuity checks; the corrected command above passes. Counts refreshes passed: methods 46.121s, array/any 52.567s, intersection/undefined 41.027s; runtime/views 38.517s.

Replay command (site and exact reason come from the census table; every replay's output is a log):

```
go run ./stage3/census/latent/replay -project /tmp/adamic-class-set-property-census/src/tsc/tsc.ts -where /tmp/adamic-class-set-property-census/src/compiler/<site> -kind NotYet -reason '<exact reason>'
```

Each mutant was applied separately and restored. The five mutants-*.json files retain every exit code and exact /tmp log path. Tests read the same reduced .a fixtures used by the Node oracle; native layout-guard probes additionally build a sanitized C harness.

| Mutant | What caught it |
| --- | --- |
| method-slot | Literal-method fixture stops with NotYet |
| method-write-not-read | Assignment LHS wrongly triggers unbound-method Refused |
| union-storage | Union storage fixture stops with NotYet |
| union-read | Backend/native disagreement from wrong slot layout |
| union-narrow-check | Checked alias-narrowing fixture continues instead of stopping |
| union-shape-type | Number and boolean view stdout differs |
| union-own-read | ASan heap-use-after-free |
| union-object-tag | Object member stdout differs |
| union-optional-slot | ASan invalid field access |
| union-optional-receiver | UBSan null member access |
| structural-slot | Structural method fixture stops with NotYet |
| own-method-ir | Checked inherited-method backend/native disagreement |
| own-method-native | Checked inherited-method backend/native disagreement |
| own-method-javascript | Checked inherited-method backend/native disagreement |
| any-accepted | Exact any-boundary assertion fails |
| array-empty-range | Caught RangeError output differs |
| array-decrement | Remaining array/output differs |
| array-clear | Cleared array/output differs |
| array-receiver-once | Receiver-call count/output differs |
| undefined-symbol | Undefined field fixture stops with NotYet |
| undefined-write-elided | Frozen write proceeds instead of TypeError |
| intersection-kept | Named intersection fixture stops with NotYet |
| intersection-packed | Optional-pair intersection fixture stops with NotYet |
| intersection-phantom | Unused phantom member stops the intersection fixture |
| intersection-agreement | Same Node fixture fails direct constituent slot agreement assertion |
| union-write-fit | Mixed-union fixture catches ASan heap-buffer-overflow and exit disagreement |
| runtime-retain | ASan heap-use-after-free |
| runtime-boolean | RegExp view prints 1 instead of true |
| static-owner | Static view wrongly stops instead of printing 7 then 9 |
| class-public-shape | Private-field class spread wrongly stops with unknown layout |
| runtime-unknown-scalar | Sanitized harness returns 0 instead of named panic 70 |
| runtime-foreign-slot | Sanitized harness returns 0 instead of named panic 70 |
| accessor-own-method | Accessor-rewritten inherited method produces backend/native disagreement |

Runtime-owner review: the only runtime C edit is the NEW separate helper file internal/native/runtime/class_set_property_union.c. Existing runtime C files were not edited. Its helpers locate the owning slot layout, retain reference slots, box known runtime scalar layouts, and fail closed on unknown scalars or foreign slots.

Ownership: origin/codex/notyet-* histories were checked after the compiler base. Other workers' property, arrayMethodArguments, representation, callClosure, combine and callOrMethod functions were not edited. Shared hooks and every file outside setProperty are named in each commit message. The flow fixture filter excludes only the three compile-time refusal controls; positive fixtures retain graph, trace, liveness and mutable-range verification.

Rulings and limits: any lacks a storage proof and remains refused by doctrine; Map/inherited-method expandos need a language ruling; direct class/library method replacement remains refused. General length growth/holes and runtime scalar layouts without representation metadata are not covered. Path and named next stops belong to other lessons. New fixture files are all .a, no cohere code was copied, no PR was opened, and pushes target only codex/notyet-class-set-property.

Landing: origin/main efe9f4042049234e5a52639fe77b47c311fd530c was merged without conflicts at 4249054a3b82e4343bd2d1e7011f718c0c4c8220. A final fetch confirmed main was still at that SHA. All 14 representative replays were repeated and their findings are unchanged. Focused commands above pass after landing: oracle 7.661s, lower 11.068s, native 4.675s, flow 11.560s, IR 4.365s; JavaScript has no standalone tests. `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1` verifies the recorded counts in 63.758s. Landing logs: /tmp/adamic-class-set-property-landing-<package>.log. The final documentation commit changes only this report and replay log paths.

Runtime shape-kind review follow-up:

The requested range `git log origin/area/runtime..17b5a053` contains only 17b5a053a19b40cad6d08edf2f2aadb6564e865c. It was cherry-picked at 6a026ee4, without merging the runtime area. Conflict resolution kept this branch's three-argument closure ABI, full IR-type shape identity (packed scalars still differ), and adamic_object_new_in allocator entry point. Changes to files absent from this branch were omitted. The runtime test uses this branch's Flags API instead of area/runtime's LinkFlags.

adamic_union_slot_owner is unchanged. adamic_union_runtime_field now reads owner->shape->kinds[index], retains references, boxes numbers, and uses immortal boolean boxes. All four hard-coded name layouts and the string comparison include are removed. Invalid or missing kind metadata fails with a named guard. This supersedes the earlier runtime-name classification and metadata-free scalar limitation described above; generated packed-scalar handling remains in the emitter.

Three new .a fixtures, registered from class_set_property_union_shape_test.go, store RegExp constructor fields and exec result fields, stat fields, and numeric/boolean/owned-string iterator result fields into mixed union slots. The readonly views exercise runtime scalar boxing as well as reference retention. Source Node, emitted JavaScript, native ASan/UBSan, release and ownership checks pass. The counted stat fixture also passes independently before mutation.

The stat size mutant changes its kind from number to reference while leaving its ownership bitmap unchanged. TestClassSetPropertyShapeStatCounted fails at the counted runtime debug check, before field reads: `adamic: inconsistent shape kind for field size` (exit -1). The boolean-boxing mutant returns number boxes; the iterator fixture catches stdout differences (false/true becomes 0/1). Both were restored. Logs and exit codes are in mutants-shape-kinds.json. The runtime owner's own TestCountedShapeKindsCatchRuntimeMutant also passes its valid/mutant controls.

Follow-up commands, all redirected to /tmp/adamic-class-set-property-shape-<package>-final.log:

```
go test ./internal/native -run '^(TestShapeKindsDistinguishScalarLayouts|TestCountedShapeKindsCatchRuntimeMutant|TestUnionRuntimeFieldGuards|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestRegExpIteratorResultShape|TestRegExpBytecodePatternUnits)$' -count=1
go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/class_set_property_|^TestClassSetPropertyShapeStatCounted$' -count=1
go test ./internal/regexp -run '^(TestParse|TestFlags|TestQuantifierBounds)$' -count=1
go test ./internal/flow -run '^TestClassSetPropertyFlow$|^TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/class_set_property_union_shape_|^Test(LivenessHoldsOnEveryPath|EveryMutationIsInItsRange)/programs/../oracle/testdata/class_set_property_union_shape_' -count=1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
```

Outputs: native 1.844s, oracle 14.306s, regexp .225s, flow 3.810s, counts 56.827s. The initial native run failed to compile because LinkFlags is absent on this branch; the API adaptation above fixed it. Setup and toolchain from the prior unit remain present; resumed nproc=5. No additional lower, IR or backend function changes were needed. The only local runtime C edit is class_set_property_union.c; all other runtime changes are the owner's cherry-picked metadata commit.

Packed runtime scalar correction:

The exhausted numeric and boolean iterator result variants exposed a gap in the initial metadata follow-up b6eec03a: the source/JavaScript backend printed undefined, while native boxed the reserved word as NaN or true. The extended iterator fixture now includes exhaustion for numbers, booleans and references. Runtime number kinds decode the existing reserved undefined word before boxing; boolean kinds decode both that word and packed 0/1/2 slots. Actual field classification still reads only shape->kinds[index], with no name comparison or layout recognition. This also preserves ordinary numbers (including NaN), booleans and references.

Two additional mutants remove number and boolean undefined decoding separately. The same Node fixture catches each through stdout disagreement; both are restored. mutants-shape-kinds.json now records all four follow-up mutants. The counted wrong-kind check remains unchanged. The final correction is committed and pushed once after all local checks pass, following the user's updated shared-fast-gate standing rule.

Final correction outputs: native 13.485s, oracle 10.199s, flow 7.976s, counts refresh 70.980s. Commands match the follow-up selections above; logs are /tmp/adamic-class-set-property-shape-packed-<package>-final.log. The unchanged regexp package checks passed in .225s. No further code or fixture changes followed these successful checks.
