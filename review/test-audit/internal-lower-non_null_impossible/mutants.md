| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/non_null.go:106 | ".ts" -> ".a" | TestAdamicNullishAssertionsAreRefused, TestNonNullAssertionLowersToNullishPanic, TestNonNullAssertionOnPresentTypeIsErased, TestPredicateOverloadRuntime, TestReadinessElisionRequiresDominatingAssignment, TestWhatZeroOneRefusesIsRefusedWithAFix |
| M02 | internal/lower/non_null.go:81 | value.Type() == ir.Number \|\| value.Type() == ir.Boolean -> value.Type() == ir.MaybeNumber \|\| value.Type() == ir.Boolean | TestNonNullAssertionLowersToNullishPanic, TestReadinessElisionRequiresDominatingAssignment |
| M03 | internal/lower/non_null.go:96 | "non-null assertion failed at " -> "" | TestNonNullAssertionLowersToNullishPanic |
| M04 supplemental | internal/lower/non_null.go:83 | return value, nil -> return ir.NumberConstant{Value: 1}, nil |  |
| M05 | internal/lower/refusals.go:28 | "use a Map, which keeps keys in the order they were added" -> "use a Map" | TestOptionalIndexingKeepsIndexSignaturesRefused |
| M06 | internal/lower/optional_indexing_map.go:99 | return hazard -> return false | TestOptionalIndexingMapShapeRefused |
| M07 | internal/lower/optional_indexing.go:35 | index.Type() != ir.Number -> index.Type() != ir.String | TestOptionalIndexingKeepsUnsupportedStorageNotYet |
| M08 | internal/lower/optional_widening.go:86 | !isClassInstance(source) -> isClassInstance(source) | TestClassWrongOutputIteratorReceiver, TestIteratorMapperIndexHasNumberRepresentation, TestOptionalWideningAllowed, TestOptionalWideningRefused |
| M09 | internal/lower/optional_widening.go:81 | skip[property.Name] -> false | TestOptionalWideningSpreadOverwrite |
| M10 | internal/lower/optional_widening.go:179 | target = l.impliedTarget(node) -> target = nil |  |
| M11 | internal/lower/optional_widening.go:196 | " on the source type, or build a fresh object with known fields (adamic/no-optional-widening)" -> " on the source type, or build a fresh object with known fields (adamic/no-optional-view)" | TestOptionalWideningRefused |
| M12 | internal/lower/generic.go:259 | len(present) == 1 && given.Flags()&checker.TypeFlagsUndefined == 0 -> len(present) == 0 && given.Flags()&checker.TypeFlagsUndefined == 0 | TestOverloadInferenceWitnesses |
| M13 | internal/lower/generic.go:269 | missing != nil -> missing == nil | TestOverloadInferenceWitnesses |
| M14 | internal/lower/generic.go:288 | if _, isSet := into[declared]; !isSet { -> if _, isSet := into[declared]; isSet { | TestAMutableLocationSeenWiderIsRefused, TestCensusOverloadBinderGuards, TestDefaultTaggedInterfaceAdmission, TestEmptyLiteralGenericReturnUsesSamePath, TestGenericFunctionPolymorphicRecursionIsRefused, TestGenericJSONUnionArrayIsNotYet, TestGenericUnionFixtureHasSeparateInstances, TestOverloadInferenceWitnesses, TestPredicateOverloadRuntime |
| M15 | internal/lower/class_inheritance.go:140 | checkABI && (!knownA \|\| !knownB \|\| a != b) -> checkABI && (!knownA \|\| !knownB \|\| a == b) | TestEnumSlotViews, TestInheritanceAllowsSoundOverrides, TestInheritanceNativeSignatureLimits, TestInheritanceNativeSignatureNeighbors, TestInheritanceRefusesFalseNominalViews, TestInheritanceRefusesUnsoundOverrides, TestOverrideNativeSignature family |
| M16 | internal/lower/class_inheritance.go:98 | !l.classAssignable(accepts, override) -> !l.classAssignable(override, accepts) | TestClassFeaturesAccessorRefusals |
