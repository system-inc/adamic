# Whole overload-result series on main

Applied all 18 own commits after fork 4885cec50290686df487b62aac47c85d871ed40c, in order. No evidence commit was skipped. Source merge commits 234467ff and e19e87c7 were excluded because they import other history. Landed main e2492670 was merged as aa8e2afe before applying the series.

## Source commits taken

```text
c68bf26c Add checked overload result entries
8f124d1c Measure overload regions on c68bf26c
0c84ce9c Serve concrete overload callback contracts
81b7a0e7 Record callback proofs and unchanged region counts
bc51bf1b Record current stops in the seven overload regions
8c7a0d84 Admit proved structural overload results
6970f0b4 Report structural proofs and remaining region stops
dacbdb8c Check overload result field storage and retain constrained parameters
ba14c913 Record overload field stops and the TNode region reveal
c09a9604 Prove fixed return contracts and name failed overload relations
9b3882f3 Record failed overload relations and remaining region proofs
0cfe794b Check TypeScript overload result fields at resolved calls
3a501d3a Record TypeScript overload checks and current hidden region stops
f709b885 Propose contracts for escaped overloads and visitor invocation domains
fdb63c7b Keep overload result checks on known closure promises
3f656458 Record overload value checks and the next hidden boundary
5a718e93 Prove or check coupled visitor inputs at invocation sites
e9fd9bbb Record visitor proofs, mutants and remaining result boundary
```

## Resolutions

Count conflicts in c68bf26c, 0c84ce9c, 8c7a0d84, dacbdb8c, c09a9604, 0cfe794b, fdb63c7b and 5a718e93 retained main's rows until regeneration. The recorder adds 26 rows and changes no existing row.

dacbdb8c extracts hidden-06's constraint-backed TNode storage into constraintStorage. Concrete substitution and main's mapper path remain first. The overlapping TNode tests retain the main-based parallel probes and add the incoming parameter-node checks. 0cfe794b combines the checked overload result entry with main's namespaceReadyCall wrapper and keeps the moved NotYet fixture paths.

The Source storage expectation moved because hidden-06 source commit 702c3ecb retained the runtime tag for constrained TNode, producing Union storage. dacbdb8c now extracts that same rule. The Source fixture still matches Node in JavaScript, release native and sanitized native.

Imported aggregate subtests were made independently selectable top-level parallel tests, preserving every case. Historical evidence was moved under this review directory; its old census measurements are not presented as measurements of this tip.

## Changed result

The existing binder-result refusal is now a checked numeric result entry. The new IR assertion pins the guard; the executable fixture exits 70 before printing ran in both backends, while Node prints ran. Dropping the entry check fails that fixture. Constraint and parameter refusals remain.

Moved-result: TestCensusOverloadBinderGuards/result Refused -> checked result entry, exit 70 before use

## Verification

All commands sourced /workspace/adamic-tools/env.sh and used GOMAXPROCS=4. Go tests used -count=1 -timeout 90s, an outer timeout, and log files. nproc reports 5; the instance CPU quota is 4. Setup succeeded: build ready 66.150 seconds, total 66.520 seconds. Stale Go cache entries were removed to recover 11,661,680,974 bytes.

| Command slice | Package seconds | Result |
| --- | ---: | --- |
| internal/lower, TestOverload/TestHidden/Source/TestCensusOverload | 1.765 | pass |
| cmd/adamic, TestExplainOverload(Field/Value/Visitor)Checks | 1.188 | pass |
| internal/oracle, initial overload and hidden controls | 3.846 | pass |
| internal/oracle, uncached overload and TNode controls | see oracle-top.json.log | pass |
| internal/oracle, uncached Node fixture comparisons | see oracle-fixtures.json.log | pass |
| TestCountsAreRecorded -args -update-counts | 76.318 | pass on retry |

The first count recorder reached its existing 90-second timeout while building the interface-count helper after cache cleanup. The retry passed. This unchanged recorder is not a newly introduced test leaf. No whole package or whole gate was run.

## Measured test leaves


lower.json.log

| Test | Seconds |
| --- | ---: |
| TestOverloadedShorthandFunctionValueStaysNotYet | 0.08 |
| TestHiddenTNodeConstraintMutationRemainsNotYet | 0.08 |
| TestOverloadVisitorResultStorage | 0.06 |
| TestRepresentationClockSourceCheckedTypes | 0.08 |
| TestOverloadStructuralRefusesMissingBlockField | 0.16 |
| TestOverloadVisitorPathsInterveningElementEffect | 0.07 |
| TestOverloadVisitorPathsReturnedCallback | 0.07 |
| TestOverloadVisitorPathsStoredCallback | 0.04 |
| TestOverloadVisitorPathsFabricatedNode | 0.04 |
| TestOverloadVisitorResultIndependent | 0.14 |
| TestOverloadValueStopsOpaqueStorage | 0.05 |
| TestOverloadValueStopsMixedParameter | 0.06 |
| TestOverloadVisitorPathsUnguardedPresence | 0.06 |
| TestOverloadValueStopsMutableTarget | 0.04 |
| TestOverloadStructuralRefusesWritableVariance | 0.13 |
| TestOverloadStructuralRefusesReverseReadonlyVariance | 0.18 |
| TestOverloadCallbackFieldStorage | 0.06 |
| TestOverloadResultsLiarStopsEvaluateLiar | 0.53 |
| TestOverloadStructuralRefusesSharedWiderResult | 0.18 |
| TestOverloadResultsFixturesScalar | 0.05 |
| TestOverloadResultsFixturesParameter | 0.05 |
| TestOverloadResultsFixturesBinding | 0.06 |
| TestOverloadResultsFixturesEvaluate | 0.04 |
| TestOverloadResultsFixturesTransform | 0.04 |
| TestOverloadCallbackUnservedResult | 0.05 |
| TestOverloadInferenceWitnesses/one_compatible_binder | 0 |
| TestOverloadInferenceWitnesses/incompatible_known_member | 0 |
| TestOverloadInferenceWitnesses/multiple_unknown_binders | 0 |
| TestOverloadInferenceWitnesses/optional_rigid_binder | 0 |
| TestOverloadInferenceWitnesses | 0.06 |
| TestOverloadCallbackUnservedRepresentation | 0.07 |
| TestOverloadCallbackUnservedAlias | 0.05 |
| TestOverloadCallbackUnservedGeneric | 0.05 |
| TestHiddenTNodeConstraintArrayLayout | 0.03 |
| TestHiddenTNodeConstraintObject | 0.06 |
| TestHiddenTNodeGenericValueRemainsNotYet | 0.06 |
| TestHiddenTNodeConstraintAny | 0.04 |
| TestHiddenTNodeConstraintUnknown | 0.04 |
| TestHiddenTNodeConstraintUnconstrained | 0.05 |
| TestHiddenTNodeConstraintObjectUnion | 0.04 |
| TestHiddenTNodeConstraintStructuralLength | 0.05 |
| TestHiddenTNodeConstraintBoolean | 0.05 |
| TestHiddenTNodeConstraintNumber | 0.03 |
| TestHiddenTNodeConstraintString | 0.04 |
| TestOverloadCallbackServed | 0.05 |
| TestOverloadCallbackUnservedInput | 0.06 |
| TestCensusOverloadBinderResultHasCheckedEntry | 0.06 |
| TestHiddenNeverArrayRejectsWritableWidening | 0.07 |
| TestOverloadResultsLiarStopsLiar | 0.54 |
| TestOverloadStructuralRefusesMixedFieldStorage | 0.14 |
| TestOverloadStructuralRefusesBlockFactoryCovariance | 0.15 |
| TestOverloadStructuralRefusesVisitorInputContravariance | 0.14 |
| TestOverloadStructuralRefusesFactoryResultCovariance | 0.14 |
| TestOverloadStructuralRefusesVisitorDomainLiar | 0.16 |
| TestOverloadStructuralRefusesFactoryFieldStorage | 0.16 |
| TestOverloadResultsRefusesEscapingCheckedOverload | 0.05 |
| TestOverloadStructuralRefusesFactoryWritableInvariance | 0.18 |
| TestOverloadResultsRefusesSharedReadonlyResult | 0.06 |
| TestOverloadResultsRefusesUnservedGenericDomain | 0.04 |
| TestOverloadStructuralRefusesFactoryLiteralCovariance | 0.15 |
| TestOverloadResultsRefusesOptionalResultProperty | 0.05 |
| TestOverloadResultsRefusesSharedWritableResult | 0.04 |
| TestCensusOverloadRelation/parameter_contravariance | 0.04 |
| TestOverloadStructuralResults | 0.19 |
| TestCensusOverloadRelation/mutable_parameters | 0.05 |
| TestCensusOverloadRelation/generic_parameter_constraints | 0.04 |
| TestCensusOverloadRelation/result_covariance | 0.07 |
| TestCensusOverloadRelation | 0 |
| TestCensusOverloadReturnProof/parameter_return | 0.04 |
| TestCensusOverloadReturnProof/empty_return | 0.03 |
| TestCensusOverloadBinderGuards/constraint | 0.05 |
| TestCensusOverloadReturnProof/captured_write | 0.04 |
| TestCensusOverloadReturnProof/rebound_parameter | 0.03 |
| TestCensusOverloadReturnProof/lying_return | 0.03 |
| TestCensusOverloadReturnProof | 0 |
| TestCensusOverloadBinderGuards/parameter | 0.04 |
| TestCensusOverloadBinderGuards | 0 |
| TestOverloadResultsLiarStopsParameterLiar | 0.39 |

cli.json.log

| Test | Seconds |
| --- | ---: |
| TestExplainOverloadVisitorChecks | 0.53 |
| TestExplainOverloadValueChecks | 0.54 |
| TestExplainOverloadFieldChecks | 1.18 |

oracle-top.json.log

| Test | Seconds |
| --- | ---: |
| TestHiddenTNodeValueExactNotYet | 0.14 |
| TestOverloadStructuralFieldRefusal | 0.26 |
| TestOverloadValuesOrderLiar | 0.67 |
| TestOverloadValuesOrderValid | 0.68 |
| TestOverloadValuesNarrowLiar | 0.59 |
| TestOverloadValuesNarrowValid | 0.6 |
| TestOverloadValuesReturnedLiar | 0.57 |
| TestOverloadValuesReturnedValid | 0.57 |
| TestOverloadVisitorInstantiations | 0.57 |
| TestOverloadVisitorsHelperValid | 0.87 |
| TestOverloadVisitorsOverloadedHelperLiar | 0.6 |
| TestOverloadVisitorsDefaultLiar | 0.64 |
| TestOverloadVisitorsHelperLiar | 0.59 |
| TestOverloadVisitorsOverloadedHelperValid | 0.9 |
| TestOverloadFieldHatchValueValid | 0.61 |
| TestOverloadFieldHatchValueUndefined | 0.62 |
| TestOverloadValueProof | 0.55 |
| TestOverloadFieldHatchValueLiar | 0.72 |
| TestOverloadVisitorsOriginalLiar | 0.6 |
| TestOverloadVisitorsOriginalValid | 0.86 |
| TestOverloadValuesModuleValid | 0.62 |
| TestOverloadValuesModuleLiar | 0.66 |
| TestOverloadFieldHatchKindValid | 0.61 |
| TestOverloadContractRulings/token | 0.19 |
| TestOverloadContractRulings/serializer | 0.17 |
| TestOverloadContractRulings/trampoline | 0.18 |
| TestOverloadContractRulings | 0 |
| TestOverloadFieldHatchKindLiar | 0.6 |

oracle-fixtures.json.log

| Test | Seconds |
| --- | ---: |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_results_parameter.a | 0.49 |
| TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binders.a | 0.52 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_structural_block.a | 0.55 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_values/narrow-valid.ts | 0.55 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_structural_fields.a | 0.45 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_structural_factory.a | 0.49 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_values/proven.a | 0.51 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_values/returned-valid.ts | 0.58 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_structural_evaluator.a | 0.49 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_callback_served.a | 0.52 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_results_binding.a | 0.53 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_results_evaluate.a | 0.57 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_results_transform.a | 0.5 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_field_hatch/value-undefined.ts | 0.54 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_field_hatch/value-valid.ts | 0.54 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_field_hatch/kind-valid.ts | 0.52 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_original_node.a | 0.47 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_trailing_comment_range.a | 0.48 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_leading_comment_range.a | 0.49 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_ancestor_directory.a | 0.44 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_mutate_map_skipping_new.a | 0.49 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_sort_deduplicate.a | 0.46 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_resolve_type_names.a | 0.46 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_mutate_map.a | 0.5 |
| TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_never_array.a | 0.43 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_results_scalar.a | 0.46 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_array_to_numeric_map.a | 0.47 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_array_to_map.a | 0.53 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_inference_contracts.a | 0.51 |
| TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_never_array_observations.a | 0.53 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_array_to_multimap.a | 0.52 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_visitors/helper.a | 0.52 |
| TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_contracts.a | 0.58 |
| TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_generic_tnode_constraints.a | 0.49 |
| TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_generic_tnode.a | 0.49 |
| TestNativeAgreesWithNode/internal/oracle/testdata/representation_clock_source.a | 0.41 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_visitors/overloaded-helper.ts | 0.53 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_visitors/original.ts | 0.51 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_visitors/instantiations.ts | 0.54 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_visitors/overloaded-helper.a | 0.59 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_visitors/original.a | 0.53 |
| TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_lie.a | 0.48 |
| TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binder_result_checked.a | 0.5 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_values/module-valid.ts | 0.57 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_values/order-valid.ts | 0.48 |
| TestNativeAgreesWithNode/internal/oracle/testdata/overload_visitors/helper.ts | 0.5 |
| TestNativeAgreesWithNode | 0.03 |

## Mutants

All 33 mutants fail their owning assertion. Sources are .go.txt and overlays; no product file was modified for replay. Two initial selectors were corrected after top-level splitting. The visitor-input mutant is caught by the missing variance diagnostic; the later visitor invocation guard still refuses the program.

| Mutant | Failing tests and seconds |
| --- | --- |
| results-trust-optional-property | TestOverloadResultsRefusesOptionalResultProperty (0.03) |
| results-trust-result | TestOverloadResultsLiarStopsLiar (0.89), TestOverloadResultsLiarStopsEvaluateLiar (1.35) |
| results-drop-specialization | TestNativeAgreesWithNode/internal/oracle/testdata/overload_results_binding.a (0.55), TestNativeAgreesWithNode (0.04) |
| results-drop-parameter-check | TestOverloadResultsLiarStopsParameterLiar (0.53) |
| results-trust-shared-result | TestOverloadResultsRefusesSharedReadonlyResult (0.28), TestOverloadResultsRefusesSharedWritableResult (0.29) |
| callback-input | TestOverloadCallbackUnservedInput (0.07) |
| callback-result | TestOverloadCallbackUnservedResult (0.09) |
| callback-representation | TestOverloadCallbackUnservedRepresentation (0.09) |
| callback-field-storage | TestOverloadCallbackFieldStorage (0.08) |
| structural-drop-proof | TestOverloadStructuralResults (0.07) |
| structural-trust-shape | TestOverloadStructuralRefusesMissingBlockField (0.17) |
| structural-reverse-covariance | TestOverloadStructuralRefusesReverseReadonlyVariance (0.19) |
| structural-mutable-variance | TestOverloadStructuralRefusesWritableVariance (0.2) |
| fields-drop-field-storage | TestOverloadStructuralRefusesMixedFieldStorage (0.18) |
| fields-drop-constraint-storage | TestHiddenTNodeConstraintNumber (0.17), TestHiddenTNodeConstraintObjectUnion (0.14), TestHiddenTNodeConstraintString (0.12), TestHiddenTNodeConstraintObject (0.15), TestHiddenTNodeConstraintBoolean (0.2), TestHiddenTNodeConstraintMutationRemainsNotYet (0.09) |
| fields-object-layout | TestHiddenTNodeConstraintMutationRemainsNotYet (0.09) |
| fields-array-layout | TestHiddenTNodeConstraintArrayLayout (0.09) |
| fields-unconstrained-layout | TestHiddenTNodeConstraintUnconstrained (0.27), TestHiddenTNodeConstraintAny (0.14), TestHiddenTNodeConstraintUnknown (0.09) |
| fields-ignore-substitution | TestHiddenTNodeConstraintStructuralLength (0.21), TestHiddenTNodeConstraintObject (0.22), TestHiddenTNodeConstraintNumber (0.24), TestHiddenTNodeConstraintObjectUnion (0.3), TestHiddenTNodeConstraintString (0.17), TestHiddenTNodeConstraintAny (0.17), TestHiddenTNodeConstraintBoolean (0.2), TestHiddenTNodeConstraintUnconstrained (0.11), TestHiddenTNodeConstraintArrayLayout (0.14), TestHiddenTNodeConstraintUnknown (0.16) |
| fields-ignore-mapper | TestHiddenTNodeConstraintBoolean (0.23), TestHiddenTNodeConstraintString (0.28), TestHiddenTNodeConstraintStructuralLength (0.31), TestHiddenTNodeConstraintObjectUnion (0.24), TestHiddenTNodeConstraintObject (0.28), TestHiddenTNodeConstraintAny (0.24), TestHiddenTNodeConstraintArrayLayout (0.17), TestHiddenTNodeConstraintUnknown (0.11), TestHiddenTNodeConstraintUnconstrained (0.14) |
| admission-drop-fixed-return-proof | TestOverloadStructuralResults (0.35) |
| admission-reverse-result-covariance | TestOverloadStructuralRefusesFactoryLiteralCovariance (0.39) |
| admission-trust-visitor-input | TestOverloadStructuralRefusesVisitorDomainLiar (0.31) |
| admission-trust-writable-result | TestOverloadStructuralRefusesFactoryWritableInvariance (0.24) |
| admission-drop-return-storage | TestOverloadStructuralRefusesFactoryFieldStorage (0.51) |
| admission-drop-relation-diagnostic | TestOverloadStructuralRefusesFactoryResultCovariance (0.37) |
| field-hatch-kind | TestOverloadFieldHatchKindLiar (0.57) |
| field-hatch-value | TestOverloadFieldHatchValueLiar (0.62) |
| values-drop-indirect-check | TestOverloadValuesReturnedLiar (0.89) |
| values-skip-single-signature-check | TestOverloadValuesNarrowLiar (0.93) |
| visitors-erase-TIn-to-Node | TestOverloadVisitorsHelperLiar (0.88) |
| visitors-admit-unproven-invocation | TestOverloadVisitorsOverloadedHelperLiar (0.9) |
| binder-result-drop-check | TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binder_result_checked.a (0.78), TestNativeAgreesWithNode (0.08) |

## Scope

This lands the whole overload-result series toward step 30. No new full TypeScript or cohere census was run. Optional array access and receiver-independent method destructuring remain later steps. Raw .ts fixtures are the original TypeScript-only unchecked controls paired with .a proof/refusal fixtures.
