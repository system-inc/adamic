# Optional array replay and method destructuring on main

Applied d7ca8e55 as ee126732 and 625e7879 as 270bb5dc onto 8e29eeb5. Each source branch has one own commit: parents dcdbb909 and 4885cec5 respectively. No own commit or evidence-only commit was skipped.

Main already contains cfa29460, which lowers optional array, string and typed-array indexing and chain continuations. The object.go conflict keeps optionalIndex rather than routing arrays through the older, narrower optionalArrayElement. The unused incoming optional_array.go helper was removed. Thus optional-array support is already present on this base; this step adds its independent fixtures, semantic mutants and exact caller replay. No additional revealed bytes are claimed.

The method change proves object-literal methods receiver-independent by scanning runtime syntax, including defaults and nested arrows. Such methods use ordinary closure storage and extraction preserves identity. Receiver-dependent, explicit-this and class/prototype origins remain refused. Main's existing namespace, view and mapper checks remain intact.

Both count conflicts retained main's table pending regeneration. Incoming evidence was retained under this review directory as historical evidence. It does not measure this tip. Two old receiver-independent refusal probes now have positive tests; the old class/prototype refusals remain. Imported negative subtests are selectable top-level parallel tests.

The imported false registry entries for two Refused fixtures were removed because main's generic non-lowering registry requires NotYet. Their explicit parallel tests still require the exact named Refused reasons and Node observations. The initial broad fixture run failed only these incorrect registry entries; the corrected run passes.

## Verification

Every Go run used -count=1 -timeout 90s and an outer timeout; output goes to logs. The final oracle run uses ADAMIC_GATE_UNCACHED=1 and GOMAXPROCS=4. Comparisons cover source Node, JavaScript backend, release native, ASan/UBSan and leaks. Refused programs have source Node observations and compiler refusal assertions, not compiled-backend claims.

| Run | Seconds | Result |
| --- | ---: | --- |
| focused lower, methods/iterators/optional indexing | 0.544 | pass |
| optional-array oracle controls and semantic mutants | 0.827 | pass |
| positive and affected ABI oracle fixtures | 5.882 | pass |
| TestCountsAreRecorded -args -update-counts | 61.147 | pass |

The count recorder is an unchanged existing leaf. All new or touched fixture/control leaves are below 60 seconds; the full per-leaf table follows. No whole packages or full gate were run.

Setup succeeded with GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh. Timing lines: Node 0.023, Go 0.024, markdown ready 0.071, submodules 0.074, clang 0.174, Go build 35.480, test binaries deferred 35.613, cache warm 35.614, total 35.642 seconds. nproc is 5; CPU quota is 4.

## Counts

Three rows added: optional array evaluation order, generic optional array access and method destructuring. Nine existing rows reduce retains/releases because independent methods no longer receive and count an unused object receiver. Allocations, frees, peak live and regions do not change. Every affected fixture was included in the Node comparison selector.

| Fixture | Before | After |
| --- | --- | --- |
| internal/oracle/testdata/census_optional_values_system.a | 29, 29, 67, 91, 14, 0 | 29, 29, 52, 76, 14, 0 |
| internal/oracle/testdata/census_optional_values_forms.a | 102, 102, 52, 169, 23, 0 | 102, 102, 51, 168, 23, 0 |
| internal/oracle/testdata/census_optional_values_padding.a | 42, 42, 73, 108, 12, 0 | 42, 42, 69, 104, 12, 0 |
| internal/oracle/testdata/census_unary_numeric.a | 187, 187, 178, 349, 14, 0 | 187, 187, 169, 340, 14, 0 |
| internal/oracle/testdata/closure_convention_receiver_rest.a | 11, 11, 12, 23, 7, 0 | 11, 11, 9, 20, 7, 0 |
| internal/oracle/testdata/host_never_branches.a | 50, 50, 57, 92, 5, 0 | 50, 50, 56, 91, 5, 0 |
| internal/oracle/testdata/host_void_method.a | 5, 5, 14, 18, 5, 0 | 5, 5, 12, 16, 5, 0 |
| internal/oracle/testdata/user_iterators.a | 663, 663, 455, 904, 64, 0 | 663, 663, 408, 857, 64, 0 |
| internal/oracle/testdata/user_iterators_rest_tdz.a | 5, 0, 5, 3, 5, 0 | 5, 0, 3, 2, 5, 0 |

## Mutants

| Mutant | Catcher | Result |
| --- | --- | --- |
| Return undefined for present first input | TestHiddenBoundary04PresentArrayMutant | both backends print 0\|0 instead of Node 7\|0 |
| Move index effects outside the present arm | TestHiddenBoundary04EagerIndexMutant | both backends increment the absent-receiver counter |
| Select the caller sibling | replay_check.py caller-selected assertion | worker exits 1 naming exact attempt mismatch |
| Omit overlapping region attempts | replay_check.py region roster assertion | empty units fail expected caller roster |
| Drop destructurableMethod proof | TestMethodDestructuringSafety receiver, arrow, default, explicit and structural cases | five assertion failures on actual unsafe admission; class controls still refuse |

Compiler mutants use Go overlays and .go.txt evidence; production files remain unchanged for replay. The region mutant is a tiny-project coverage proof, not a real-project byte remeasurement.

## Exact caller replay

The original replay_check.py passes against newly built scratch-only measurement workers. Dependency diagnostic and caller selection are separate. Wrong reason, wrong caller and non-header position are rejected. Caller and partial-region findings equal the tiny unfiltered full census; sibling units are excluded. Production Load/Lower remain disabled in these overlay workers. No real TypeScript or cohere corpus remeasurement was run; that is the next turn.

## Final merge and leaf timings

Merged main 0942c516 as d99dadcd without conflicts. Only landed stage1 tests changed. Final uncached oracle run passed in 5.490 seconds.

| Test | Seconds |
| --- | ---: |
| TestOptionalIndexingKeepsIndexSignaturesRefused | 0.07 |
| TestMethodDestructuringSafetyReceiver | 0.07 |
| TestOptionalIndexingKeepsUnsupportedStorageNotYet/typed_ordinary_property_continuation | 0.09 |
| TestReceiverIndependentMethodDestructuring | 0.13 |
| TestMethodDestructuringSafetyStructuralClass | 0.07 |
| TestOptionalIndexingKeepsUnsupportedStorageNotYet/string_numeric_key | 0.07 |
| TestOptionalIndexingMapShapeRefused | 0.1 |
| TestMethodDestructuringSafetyStructuralReceiver | 0.05 |
| TestMethodDestructuringSafetyClassOrigin | 0.06 |
| TestOptionalIndexingKeepsUnsupportedStorageNotYet/sparse_literal | 0.04 |
| TestMethodDestructuringSafetyExplicitReceiver | 0.04 |
| TestMethodDestructuringSafetyDefaultReceiver | 0.04 |
| TestMethodDestructuringSafetyArrowReceiver | 0.04 |
| TestIteratorSymbolKeysAreNotStringKeys | 0.05 |
| TestOptionalIndexingKeepsUnsupportedStorageNotYet/open_record | 0.05 |
| TestIteratorMapperIndexHasNumberRepresentation | 0.04 |
| TestLiteralMethodCapturesCannotMakeCycles | 0.03 |
| TestOptionalIndexingKeepsUnsupportedStorageNotYet/index_after_ordinary_continuation | 0.07 |
| TestOptionalIndexingKeepsUnsupportedStorageNotYet | 0.32 |
| TestLiteralMethodViewsDoNotLoseThis | 0.08 |
| TestIteratorDestructuringDoesNotLieAboutExhaustion | 0.03 |
| TestIteratorViewsCannotEraseReceivers | 0.04 |
| TestIteratorViewsCannotHideReturn | 0.04 |
| TestDestructuredMethodsCannotLoadOwnSlots | 0.14 |
| TestIteratorGapsAreExplicit/interface_Step_{value:number;done?:boolean}_const_source={_[Symbol.iterator](){return{next():Step{return{value:1};}}}};for(const_value_of_source){console.log(`${value}`);break;} | 0.04 |
| TestIteratorGapsAreExplicit/const_source={[Symbol.iterator](){return{next:()=>({value:1,done:false}),return:()=>({value:0,done:true})};}};const_iterator=source[Symbol.iterator]();iterator.next=():{value:number;done:boolean}=>({value:2,done:true});for(const_value_of_source){break;} | 0.04 |
| TestIteratorGapsAreExplicit/class_Result_{declare_value:number;declare_done:boolean;}_const_source={[Symbol.iterator](){return{next(){return_new_Result();}}}};for(const_value_of_source){break;} | 0.04 |
| TestIteratorGapsAreExplicit/const_source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};const_copy={...source};for(const_value_of_copy){break;} | 0.04 |
| TestIteratorGapsAreExplicit/const_source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};const_kept:(number|undefined)[]=[...source]; | 0.05 |
| TestIteratorDescriptorReasons/required_boolean_done | 0.03 |
| TestIteratorGapsAreExplicit/interface_Source_{[Symbol.iterator]():{next():{value:number;done:boolean}}}_function_consume(source:Source):void_{for(const_value_of_source){break;}}_consume({[Symbol.iterator](){return{next(){return{value:1,done:false};}}}}); | 0.04 |
| TestIteratorGapsAreExplicit/interface_Step_{value:number;done:boolean}_function_closing():{return():Step}|undefined{return_undefined;}_const_source={[Symbol.iterator](){return{...closing(),next():Step{return{value:1,done:false};}}}};for(const_value_of_source){break;} | 0.06 |
| TestIteratorGapsAreExplicit | 0 |
| TestIteratorDescriptorReasons/arguments_or_overloads | 0.05 |
| TestIteratorDescriptorReasons/declare_or_abstract_class_field | 0.03 |
| TestIteratorDescriptorReasons/replacing_an_iterator_protocol | 0.05 |
| TestIteratorDescriptorReasons/object_spread | 0.05 |
| TestIteratorDescriptorReasons/represented_value_field | 0.03 |
| TestIteratorDescriptorReasons/optional_iterator_method | 0.03 |
| TestIteratorDescriptorReasons | 0 |
| TestHiddenBoundary04MutationRefused | 0.28 |
| TestHiddenBoundary04AliasCastRefused | 0.28 |
| TestHiddenBoundary04PresentArrayMutant | 0.41 |
| TestHiddenBoundary04EvaluationOrder | 0.56 |
| TestHiddenBoundary04Node | 0.4 |
| TestHiddenBoundary04EagerIndexMutant | 0.42 |
| TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_generic_optional_array.a | 0.5 |
| TestNativeAgreesWithNode/internal/oracle/testdata/census_optional_values_system.a | 0.57 |
| TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_optional_array_order.a | 0.55 |
| TestNativeAgreesWithNode/internal/oracle/testdata/closure_convention_receiver_rest.a | 0.54 |
| TestNativeAgreesWithNode/internal/oracle/testdata/user_iterators_rest_tdz.a | 0.51 |
| TestNativeAgreesWithNode/internal/oracle/testdata/host_never_branches.a | 0.64 |
| TestNativeAgreesWithNode/internal/oracle/testdata/hidden_wave2_method_destructuring.a | 0.57 |
| TestNativeAgreesWithNode/internal/oracle/testdata/host_void_method.a | 0.54 |
| TestNativeAgreesWithNode/internal/oracle/testdata/census_optional_values_padding.a | 0.68 |
| TestNativeAgreesWithNode/internal/oracle/testdata/census_optional_values_forms.a | 0.83 |
| TestNativeAgreesWithNode/internal/oracle/testdata/census_unary_numeric.a | 0.57 |
| TestNativeAgreesWithNode/internal/oracle/testdata/user_iterators.a | 4.24 |
| TestNativeAgreesWithNode | 0.14 |
