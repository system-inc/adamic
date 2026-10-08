Fresh callback returns and the unchanged fixture 25 now pass; date 2026-10-08 UTC.
Commits: root 5d1b45e1876e688012518370c8de2609e7a1f7ab, including fresh-map 52882bdf and final numeric repair 9ac8e5b9; host scratch parent a4f599fb plus this evidence commit.
Commands/results: uncached complete host oracle 20.515s, eight focused backend probes 2.289s, root full lower 15.896s, focused host controls 0.218s, vet clean, all measured allocations freed.
Mutants: every-path freshness, type/storage guards, evaluation order, dispatch, sorting and live capture mutations are caught as detailed below.
Gaps: full host lower retains predicate-marker and virtual-call-target failures; full count regeneration is blocked by project-root attribution; no additional nullable reference kind was reached.

The production branch is codex/generic-function-value. The complete host integration is on our own codex/generic-function-value-host-scratch. It contains pinned library c97402ba, nullable dependency 1a83ba1a and current main 6f16a169. Root and scratch were merged with current main and checked after that merge. No main, area or library branch was pushed; no PR was opened. The host-only overload, evolving-string, default-sort and Node-console changes require the host dependencies in this scratch branch.

Before: fixture 25 refused the fresh callback-return view and subsequent lowering sites. After: its exact original source lowers and agrees with Node in both backends. Its unchanged blob is bd24f7bd2ed712f6d07455130ab8bbe76c82b743. The permanent host_read_directory_generic_walk_test.go registers that exact file. Census-wide family totals 123 and 110 were not remeasured during this walk.

Freshness is proven at the intrinsic result binding, without changing the checker's contextual types. Every callback return path must allocate a fresh array directly or use an existing proven fresh operation. Shared returns, mixed fresh/shared returns, stores and unrecognized alias/control-flow paths retain refusal. The checker still reports never[] for [] and never[][] for map; TestContextualEmptyMap pins those observations. Number and object instantiations push into the fresh rows and agree with Node. Shared, mixed and stored witnesses each print 1 on Node and keep the exact refusal with its readonly/fresh-array fix. Named callbacks and general local-alias escape analysis remain conservative.

The walk's implementation commits and their own detailed reports are:

| Fix | Commit | Evidence report |
| --- | --- | --- |
| Every fresh callback return | 52882bdf | FRESH_MAP_RETURNS.md |
| Object parameter defaults | aa1e2961 | OBJECT_PARAMETER_DEFAULT.md |
| Primitive phantom array overload elements | 50c9f0d8, host | HOST_PHANTOM_ARRAY_ELEMENT.md |
| UTF-16 substr length | a4b8b73c | STRING_SUBSTR_LENGTH.md |
| Evolving string local storage | 8596f410, host | HOST_EVOLVING_STRING_LOCAL.md |
| Positioned lastIndexOf | 93e7fee5 | STRING_LAST_INDEX_POSITION.md |
| Finite computed data fields | b18b0275 | FINITE_PROPERTY_INDEX.md |
| Multiple push argument evaluation | 5f85ac78 | ARRAY_PUSH_VALUES.md |
| Optional slice bounds | 866b4823 | ARRAY_SLICE_OPTIONAL_BOUNDS.md |
| Undefined beside concrete object union members | 2069a265 | NARROWED_ARRAY_UNDEFINED.md |
| Homogeneous immutable overload comparisons | f178e5dd, host | HOST_OVERLOAD_PRIMITIVE_COMPARISON.md |
| Optional string comparator | 34624421, host | HOST_ARRAY_SORT_OPTIONAL.md |
| Preserve host sparse consumers after merge | 63e4155a, host | HOST_ARRAY_SORT_OPTIONAL.md |
| Compatible array-union slots | a0fae438, host | HOST_ARRAY_UNION_ELEMENTS.md |
| Automatic array storage from all observations | 583d7e31 | EVOLVING_ARRAY_STORAGE.md |
| Node-only console declaration identity | ffd5bcdc, host | HOST_NODE_CONSOLE_IDENTITY.md |
| Boxed union capture cells | 8f2d9a7a | CAPTURED_BOXED_UNION.md |
| Narrowed numeric updates fitted back to storage | 9ac8e5b9 | NARROWED_NUMBER_INCREMENT.md |

No new checker context, coercion policy or null tag was introduced. The merged String sentinel helper is reused. No additional kind needing distinct null and undefined appeared during the complete fixture lowering/execution; Array/Object/Map/Closure null-and-undefined support is not broadly claimed. Existing header-tagged Union and captured strong reference cells suffice for the reached union values.

All credited mutants executed during this batch are listed here. Semantic mutants compile and exit normally unless the row explicitly names a diagnostic/check test. Every mutation was restored.

| Mutant | What caught it | Log in /tmp |
| --- | --- | --- |
| Prove only the first conditional callback return | Mixed shared-return refusal becomes nil | fresh-map-every-path-mutant.log |
| Evaluate object-default fallback eagerly/twice | Both backends disagree with Node output | object-parameter-default-mutant-double-fallback.log |
| Treat null as an undefined-only default | JavaScript output disagrees with Node | object-parameter-default-mutant-null-default.log |
| Accept phantom elements by representation only | Literal-result refusal becomes nil | host-phantom-overload-literal-mutant.log |
| Use substr count as absolute end | Both backends disagree with Node output | string-substr-length-end-mutant-valid.log |
| Ignore evolving-string storage disagreement | Mixed-write refusal becomes nil | evolving-string-mixed-mutant.log |
| Omit needle length from backward-search prefix | Both backends disagree with Node output | string-last-index-length-mutant.log |
| Always select the first computed field | Both backends disagree with Node output | finite-property-selection-mutant.log |
| Return a field for an invalid computed key | Both backends wrongly finish instead of the pinned check | finite-property-key-mutant.log |
| Push before later arguments and evaluate receiver twice | Both backends disagree on observed length and receiver count | array-push-evaluation-mutant-running.log |
| Default an undefined slice end to zero | Both backends disagree with Node output | array-slice-undefined-end-mutant.log |
| Ignore concrete reference-kind union conflicts | Array/Map conflict refusal becomes nil | union-reference-kind-mutant.log |
| Remove the stale union member IR check | Both backends wrongly finish instead of the pinned check | narrowed-array-undefined-oracle.log |
| Ignore overload primitive-pair equality | Mixed comparison refusal becomes nil | overload-comparison-pair-mutant.log |
| Ignore parameter writes in the overload proof | Reassigned comparison refusal becomes nil | overload-comparison-write-mutant-fixed.log |
| Use greater-than for the proven comparison | Both backends disagree with Node output | overload-comparison-operator-mutant.log |
| Always use default order for optional comparator | Both backends disagree with Node output | array-sort-optional-callback-mutant.log |
| Evaluate the optional comparator factory twice | Both backends disagree with Node output | array-sort-optional-evaluation-mutant.log |
| Sort by code point rather than UTF-16 | Clean native stdout disagrees with Node | array-sort-sparse-reconcile-mutants.log |
| Reverse string sort ordering | Clean native stdout disagrees with Node | array-sort-sparse-reconcile-mutants.log |
| Reverse equal sort elements | Clean native stdout disagrees with Node | array-sort-sparse-reconcile-mutants.log |
| Drop explicit undefined as though it were a hole | Clean native stdout disagrees with Node | array-sort-sparse-reconcile-mutants.log |
| Ignore array-union slot compatibility | Scalar/reference refusal becomes nil | array-union-storage-mutant.log |
| Keep only the final automatic-array observation | Changing-storage refusal becomes nil | evolving-array-root-mutant.log, evolving-array-observations-mutant.log |
| Recognize console by spelling alone | Both backends add output for a shadowed console | node-console-identity-mutant.log |
| Remove the live captured Union write | Both backends repeatedly return the old value, disagreeing with Node | captured-union-write-mutant.log |
| Add instead of subtracting in optional-number decrement | Both backends output 5,6,7,8 instead of 3,4,3,4 | narrowed-increment-operator-mutant.log |

Discarded mutation attempts are not counted as proof: an early substr witness failed Load; early push mutations corrupted an IR slice or produced cyclic IR and failed before execution; the first overload-write witness met a different closure gap so the intended mutant survived. Each was replaced with the running or pinned-diagnostic witness listed above. Their logs remain string-substr-length-end-mutant.log, array-push-evaluation-mutant.log, array-push-evaluation-mutant-valid.log and overload-comparison-write-mutant.log.

Final validation used source /workspace/adamic-tools/env.sh in every toolchain shell and wrote all test output to logs. Exact commands are:

```sh
# Root, after main merge
 go test ./internal/lower -count=1 -timeout 10m
 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/(fresh_map_returns|evolving_array_storage|captured_boxed_union|narrowed_number_increment).a$' -count=1 -timeout 10m -v
 go vet ./internal/lower ./internal/oracle
# Host, after main merge
 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/fixtures/host/(25_readDirectory|console_identity).a$' -count=1 -timeout 10m -v
 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/(fresh_map_returns|array_union_reference_elements|captured_boxed_union|evolving_array_storage|narrowed_number_increment|array_sort_optional_comparator|overload_primitive_comparison|phantom_array_element_overload).a$' -count=1 -timeout 10m -v
 go test ./internal/oracle -run '^TestFreshMap(SharedReturnsRefused|MixedReturnRefused)$' -count=1 -v
 go test ./internal/oracle -run '^Test(CapturedUnionWriteMutant|FreshMapSharedReturnRefused|FreshMapMixedReturnRefused|FreshMapStoredReturnRefused|ArrayUnionIncompatibleStorageRefused|EvolvingArrayChangingStorageOnNode|ArrayStringSortMutants)$' -count=1 -timeout 10m -v
 go test ./internal/lower -run '^Test(ContextualEmptyMap|FreshCallbackKeepsSharedNestedViews|EvolvingArrayRejectsChangingStorage|OverloadMixedComparisonRemainsNotYet|OverloadReassignedComparisonRemainsNotYet|OptionalSortKeepsNumericDefaultLimit|PhantomArrayElementOverloadLiteralRefused)$' -count=1 -v
 go test ./internal/lower -run '^TestCensusSmallFiniteKeysLowered$' -count=1 -v
 go test ./internal/fuzz -run '^TestExecute' -count=1 -timeout 10m -v
 go vet ./internal/lower ./internal/oracle ./internal/fuzz
```

The corrected commands actually selected four root positive fixtures (0.692s), two host fixtures (20.515s, fixture 25 alone 20.49s) and eight host probes (2.289s), with source Node, native ASan/UBSan/release, JavaScript and leak checks. Shared/mixed/stored negative Node/pinned-refusal controls passed 0.477s; other oracle controls/mutants 2.778s; seven focused host lower controls 0.218s; finite-key expectation repair 0.044s. The main merge preserves the host's isolated fuzzer environments and adds main's CPU deadline implementation; inherited CPU tests passed 3.643s. An initial combined-path regex selected zero positive fixtures; it is not credited and was replaced by the explicit commands above.

Source Node's complete fixture output, matched by both backends, has four nonempty lines and a final empty line:

```text
a.ts,z.ts,alias/b.ts,alias/deep/c.ts
a.ts,z.ts
src/b.ts,src/deep/c.ts,a.ts
plain.js

```

Measured complete fixture counts: allocations/frees 10191/10191, retains/releases 5869/6417, peak live 2415, region allocations 0, graph counters 0/0. The ordinary counted helper also remeasured evolving-array 12/12, captured-union 13/13 and narrowed-increment 9/9 on the final host code. Logs host-walk-final-{host-backends,probe-backends,root-lower,root-backends,root-vet,fresh-controls,controls,host-lower-controls,finite-control,fuzz-merge,vet,counts}.log preserve results.

The full host lower attempt still fails TestCensusPredicateMarkerKeepsProofBoundaries with an accepted escaped marker and TestClassFeaturesAccessorRefusals with ir: virtual call has no target set. Both were observed earlier in the scratch integration; they remain outside the fixture walk. That run also had a finite-key NotYet expectation made obsolete by our proven finite-key lowering; the expectation is now updated and its focused test passes. The root full lower package is green. A full repository gate is not claimed. Full counts regeneration was attempted earlier and fails on internal/fresh/testdata/regexp_tree.ts outside its project's root files; rows here were measured through the same counted helper individually. No checks were weakened to make the full fixture pass.

Toolchain setup was bash cloud/setup.sh, log /tmp/host-walk-setup.log: clang ready 0.186s, Go build ready 30.106s, cache warm 30.340s, total 30.366s. nproc prints 5; cgroup CPU budget is 4. Go 1.27.1, Node 24.19.0, clang 20.1.8. Later the Go build cache filled the 32 GB overlay; go clean -cache reclaimed space without removing sources or evidence, and required builds were rerun. Cohere remained at 7945d102a6c18dd36adf9114a758ce646e8b2359; only existing shim API was referenced, no cohere code was copied.
