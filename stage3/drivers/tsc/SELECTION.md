# Selected compiler tests

All source paths are relative to TypeScript 6.0.3. Baseline paths and hashes
are in `selection.json`; a clean case has no `.errors.txt` baseline.

| # | Compiler test | Feature | Codes |
|---:|---|---|---|
| 1 | `varianceCantBeStrictWhileStructureIsnt.ts` | control | clean |
| 2 | `conditionalEqualityTestingNullability.ts` | types | clean |
| 3 | `commentOnClassAccessor1.ts` | classes | clean |
| 4 | `metadataOfUnion.ts` | unions | clean |
| 5 | `ambientEnumElementInitializer3.ts` | enums | clean |
| 6 | `arrayDestructuringInSwitch1.ts` | arrays | clean |
| 7 | `nestedBlockScopedBindings15.ts` | other | clean |
| 8 | `library_StringSlice.ts` | literals | clean |
| 9 | `es6ModuleWithModuleGenTargetCommonjs.ts` | modules | clean |
| 10 | `nestedTypeVariableInfersLiteral.ts` | generics | clean |
| 11 | `namespaces2.ts` | namespaces | clean |
| 12 | `expandoFunctionExpressionsWithDynamicNames2.ts` | functions | clean |
| 13 | `parseEntityNameWithReservedWord.ts` | syntax | clean |
| 14 | `keyofObjectWithGlobalSymbolIncluded.ts` | objects | clean |
| 15 | `decoratorReferences.ts` | decorators | clean |
| 16 | `yieldStarContextualType.ts` | async | clean |
| 17 | `binopAssignmentShouldHaveType.ts` | operators | clean |
| 18 | `typeofUsedBeforeBlockScoped.ts` | types | clean |
| 19 | `superHasMethodsFromMergedInterface.ts` | classes | clean |
| 20 | `intersectionApparentTypeCaching.ts` | unions | clean |
| 21 | `commentOnArrayElement12.ts` | arrays | clean |
| 22 | `nestedBlockScopedBindings4.ts` | other | clean |
| 23 | `capturedLetConstInLoop10_ES6.ts` | control | clean |
| 24 | `genericTypeParameterEquivalence2.ts` | generics | clean |
| 25 | `exportArrayBindingPattern.ts` | modules | clean |
| 26 | `functionAssignmentError.ts` | functions | clean |
| 27 | `multiExtendsSplitInterfaces1.ts` | objects | clean |
| 28 | `stringMatchAll.ts` | literals | clean |
| 29 | `enumLiteralUnionNotWidened.ts` | enums | clean |
| 30 | `asyncYieldStarContextualType.ts` | async | clean |
| 31 | `commentInNamespaceDeclarationWithIdentifierPathName.ts` | namespaces | clean |
| 32 | `templateLiteralsAndDecoratorMetadata.ts` | decorators | clean |
| 33 | `assignmentCompatability6.ts` | operators | clean |
| 34 | `parseReplacementCharacter.ts` | syntax | clean |
| 35 | `classOrder1.ts` | classes | clean |
| 36 | `conditionalExpressions2.ts` | types | clean |
| 37 | `singletonLabeledTuple.ts` | arrays | clean |
| 38 | `breakInIterationOrSwitchStatement1.ts` | control | clean |
| 39 | `narrowByClauseExpressionInSwitchTrue2.ts` | unions | clean |
| 40 | `emitOneLineVariableDeclarationRemoveCommentsFalse.ts` | other | clean |
| 41 | `nongenericConditionalNotPartiallyComputed.ts` | generics | clean |
| 42 | `exportDefaultForNonInstantiatedModule.ts` | modules | clean |
| 43 | `collisionThisExpressionAndLocalVarInProperty.ts` | objects | clean |
| 44 | `functionWithDefaultParameterWithNoStatements10.ts` | functions | clean |
| 45 | `discriminantUsingEvaluatableTemplateExpression.ts` | literals | clean |
| 46 | `assignmentCompatForEnums.ts` | enums | clean |
| 47 | `promiseWithResolvers.ts` | async | clean |
| 48 | `uniqueSymbolAssignmentOnGlobalAugmentationSuceeds.ts` | operators | clean |
| 49 | `decoratorMetadataNoStrictNull.ts` | decorators | clean |
| 50 | `unusedInterfaceinNamespace4.ts` | namespaces | clean |
| 51 | `extendedUnicodeEscapeSequenceIdentifiers.ts` | syntax | clean |
| 52 | `commentOnClassAccessor2.ts` | classes | clean |
| 53 | `contextualTypeBasedOnIntersectionWithAnyInTheMix5.ts` | unions | clean |
| 54 | `contextualTyping19.ts` | other | clean |
| 55 | `modularizeLibrary_UsingES5LibES6ArrayLibES6WellknownSymbolLib.ts` | arrays | clean |
| 56 | `genericCallInferenceInConditionalTypes1.ts` | generics | clean |
| 57 | `localImportNameVsGlobalName.ts` | modules | clean |
| 58 | `contextualTypingOfConditionalExpression.ts` | types | clean |
| 59 | `functionOverloads31.ts` | functions | clean |
| 60 | `numberAsInLHS.ts` | literals | clean |
| 61 | `constEnumErrors.ts` | enums | TS2339, TS2474, TS2475, TS2476, TS2477, TS2478, TS2567, TS2651 |
| 62 | `interfaceDeclaration1.ts` | objects | TS2300, TS2310, TS2320, TS2420, TS2454, TS2717 |
| 63 | `didYouMeanSuggestionErrors.ts` | other | TS2583, TS2584, TS2585, TS2591, TS2592, TS2593 |
| 64 | `recursiveFunctionTypes.ts` | functions | TS2322, TS2345, TS2355, TS2394, TS2554, TS2769 |
| 65 | `callOverloads2.ts` | functions | TS2389, TS2391, TS2393, TS2813, TS2814 |
| 66 | `decoratorUsedBeforeDeclaration.ts` | decorators | TS2448, TS2450, TS2454, TS2729, TS7006 |
| 67 | `undefinedTypeAssignment4.ts` | operators | TS2397, TS2414, TS2427, TS2564 |
| 68 | `emitCapturingThisInTupleDestructuring1.ts` | arrays | TS2493, TS7010, TS7017, TS7041 |
| 69 | `computedPropertiesInDestructuring1.ts` | other | TS2339, TS2349, TS2365, TS2537, TS2538 |
| 70 | `assignmentToReferenceTypes.ts` | operators | TS2628, TS2629, TS2630, TS2708 |
| 71 | `overloadModifiersMustAgree.ts` | functions | TS2383, TS2384, TS2385, TS2386 |
| 72 | `noImplicitAnyForIn.ts` | control | TS2405, TS2872, TS7053 |
| 73 | `superNewCall1.ts` | classes | TS2351, TS2377, TS17011 |
| 74 | `parseInvalidNullableTypes.ts` | syntax | TS2322, TS2677, TS17019, TS17020 |
| 75 | `recursiveConditionalCrash4.ts` | types | TS2304, TS2503, TS2589 |
| 76 | `interfacedeclWithIndexerErrors.ts` | objects | TS2411, TS2693, TS2840 |
| 77 | `objectLitIndexerContextualType.ts` | objects | TS2353, TS2362, TS2363 |
| 78 | `weakType.ts` | other | TS2322, TS2352, TS2559, TS2560 |
| 79 | `nullableFunctionError.ts` | functions | TS2721, TS2722, TS2723 |
| 80 | `builtinIterator.ts` | other | TS2322, TS2345, TS2416, TS2511, TS2515 |
| 81 | `errorForUsingPropertyOfTypeAsType03.ts` | objects | TS2339, TS2702, TS2713, TS2749 |
| 82 | `unusedDestructuring.ts` | other | TS6133, TS6198, TS6199 |
| 83 | `indirectSelfReferenceGeneric.ts` | generics | TS2449, TS2506 |
| 84 | `thisInModule.ts` | modules | TS2331, TS2683 |
| 85 | `narrowSwitchOptionalChainContainmentEvolvingArrayNoCrash1.ts` | unions | TS7005, TS7034 |
| 86 | `noImplicitSymbolToString.ts` | literals | TS2469, TS2731 |
| 87 | `strictModeInConstructor.ts` | classes | TS2376, TS17009 |
| 88 | `baseConstraintOfDecorator.ts` | decorators | TS2322, TS2507, TS2545 |
| 89 | `moduleVisibilityTest3.ts` | modules | TS2709, TS2724 |
| 90 | `constraintWithIndexedAccess.ts` | generics | TS2344, TS2536 |
| 91 | `excessivelyLargeTupleSpread.ts` | arrays | TS2799, TS2800 |
| 92 | `omittedExpressionForOfLoop.ts` | control | TS2304, TS2488, TS18050 |
| 93 | `typeOfEnumAndVarRedeclarations.ts` | enums | TS2374, TS2403 |
| 94 | `assignmentToParenthesizedExpression1.ts` | operators | TS2364, TS2695 |
| 95 | `interfaceExtendsClassWithPrivate1.ts` | classes | TS2739, TS2741 |
| 96 | `importDeclRefereingExternalModuleWithNoResolve.ts` | modules | TS2307, TS2664 |
| 97 | `mappedTypeGenericWithKnownKeys.ts` | generics | TS2551, TS2862 |
| 98 | `moduleAugmentationGlobal7_1.ts` | modules | TS2669, TS2670 |
| 99 | `untypedFunctionCallsWithTypeParameters1.ts` | generics | TS2347, TS2349, TS2420, TS2558 |
| 100 | `scopeCheckExtendedClassInsidePublicMethod2.ts` | classes | TS2662, TS2663 |
| 101 | `noUncheckedIndexedAccessCompoundAssignments.ts` | operators | TS2532, TS18048 |
| 102 | `implicitAnyDeclareTypePropertyWithoutType.ts` | objects | TS7006, TS7008, TS7013 |
| 103 | `circularReferenceInReturnType2.ts` | functions | TS7022, TS7023 |
| 104 | `mixedStaticAndInstanceClassMembers.ts` | classes | TS2387, TS2388 |
| 105 | `exportDefaultTypeClassAndValue.ts` | modules | TS2323, TS2528 |
| 106 | `interfaceMergeWithNonGenericTypeArguments.ts` | generics | TS2315, TS2346 |
| 107 | `objectFreeze.ts` | objects | TS2322, TS2540, TS2542 |
| 108 | `redeclareParameterInCatchBlock.ts` | functions | TS2451, TS2492 |
| 109 | `classExtendsNull2.ts` | classes | TS2417, TS17005 |
| 110 | `importDeclWithExportModifierAndExportAssignment.ts` | modules | TS2309, TS2694, TS2708 |
| 111 | `circularlyReferentialInterfaceAccessNoCrash.ts` | objects | TS4109, TS4110 |
| 112 | `noImplicitAnyParametersInInterface.ts` | functions | TS7006, TS7019, TS7020 |
| 113 | `duplicateSymbolsExportMatching.ts` | modules | TS2395, TS2434 |
| 114 | `truthinessPromiseCoercion.ts` | async | TS2801 |
| 115 | `unusedClassesinNamespace4.ts` | namespaces | TS6196 |
| 116 | `recursivelyExpandingUnionNoStackoverflow.ts` | unions | TS2589, TS2615 |
| 117 | `typeUsedAsTypeLiteralIndex.ts` | literals | TS2690, TS2693 |
| 118 | `anyMappedTypesError.ts` | types | TS7039 |
| 119 | `yieldExpressionInFlowLoop.ts` | async | TS7057 |
| 120 | `excessPropertyCheckWithSpread.ts` | arrays | TS2842 |
| 121 | `switchCasesExpressionTypeMismatch.ts` | control | TS2678 |
| 122 | `literalsInComputedProperties1.ts` | literals | TS2452 |
| 123 | `narrowByEquality.ts` | unions | TS2322, TS2839 |
| 124 | `this_inside-enum-should-not-be-allowed.ts` | enums | TS2332, TS2683 |
| 125 | `noMappedGetSet.ts` | types | TS2304, TS2464 |
| 126 | `restInvalidArgumentType.ts` | arrays | TS2454, TS2700 |
| 127 | `forInStatement4.ts` | control | TS2404 |
| 128 | `noImplicitAnyUnionNormalizedObjectLiteral1.ts` | unions | TS7018 |
| 129 | `enumWithPrimitiveName.ts` | enums | TS2431 |
| 130 | `inheritedStringIndexersFromDifferentBaseTypes2.ts` | literals | TS2413 |
| 131 | `errorsForCallAndAssignmentAreSimilar.ts` | operators | TS2820 |
| 132 | `missingCommaInTemplateStringsArray.ts` | arrays | TS2796 |
| 133 | `forInStatement2.ts` | control | TS2407 |
| 134 | `augmentedTypesEnum.ts` | enums | TS2300, TS2432, TS2567 |
| 135 | `narrowingTruthyObject.ts` | unions | TS18047 |
| 136 | `nestedFreshLiteral.ts` | literals | TS2561 |
| 137 | `duplicateErrorNameNotFound.ts` | other | TS2552 |
| 138 | `voidAsOperator.ts` | operators | TS2873 |
| 139 | `typeParametersInStaticAccessors.ts` | generics | TS2302, TS2322 |
| 140 | `arrayAssignmentTest5.ts` | arrays | TS2322, TS2366 |
| 141 | `forInStrictNullChecksNoError.ts` | control | TS18049 |
| 142 | `spreadUnionPropOverride.ts` | unions | TS2783 |
| 143 | `enumPropertyAccessBeforeInitalisation.ts` | enums | TS2565 |
| 144 | `inheritedStringIndexersFromDifferentBaseTypes.ts` | literals | TS2430 |
| 145 | `ClassDeclaration10.ts` | classes | TS2390, TS2391 |
| 146 | `noImplicitAnyIndexing.ts` | other | TS2339, TS7015, TS7053 |
| 147 | `genericMappedTypeAsClause.ts` | generics | TS2312, TS2322, TS2353 |
| 148 | `nonArrayRestArgs.ts` | arrays | TS2370 |
| 149 | `undefinedTypeAssignment1.ts` | operators | TS2457 |
| 150 | `useUnknownInCatchVariables01.ts` | control | TS18046 |
| 151 | `normalizedIntersectionTooComplex.ts` | unions | TS2590, TS7006 |
| 152 | `errorOnEnumReferenceInCondition.ts` | enums | TS2845 |
| 153 | `returnTypeTypeArguments.ts` | functions | TS2314, TS2564 |
| 154 | `emitThisInSuperMethodCall.ts` | classes | TS2660 |
| 155 | `varianceReferences.ts` | other | TS2322, TS2637 |
| 156 | `misspelledNewMetaProperty.ts` | objects | TS17012 |
| 157 | `varNameConflictsWithImportInDifferentPartOfModule.ts` | modules | TS2440 |
| 158 | `incorrectRecursiveMappedTypeConstraint.ts` | generics | TS2313, TS2365 |
| 159 | `objectBindingPattern_restElementWithPropertyName.ts` | arrays | TS2566 |
| 160 | `incompatibleAssignmentOfIdenticallyNamedTypes.ts` | operators | TS2564, TS2719 |
| 161 | `didYouMeanElaborationsForExpressionsWhichCouldBeCalled.ts` | control | TS2322, TS2345, TS2740, TS2741 |
| 162 | `instanceofWithPrimitiveUnion.ts` | unions | TS2358 |
| 163 | `ArrowFunctionExpression1.ts` | functions | TS2369 |
| 164 | `staticOffOfInstance1.ts` | other | TS2576 |
| 165 | `defaultValueInConstructorOverload1.ts` | classes | TS2371 |
| 166 | `inheritanceMemberFuncOverridingProperty.ts` | objects | TS2425 |
| 167 | `compareTypeParameterConstrainedByLiteralToLiteral.ts` | generics | TS2367 |
| 168 | `importedModuleAddToGlobal.ts` | modules | TS2833 |
| 169 | `spreadInvalidArgumentType.ts` | arrays | TS2698 |
| 170 | `decrementAndIncrementOperators.ts` | operators | TS2357 |
| 171 | `extendPrivateConstructorClass.ts` | classes | TS2675 |
| 172 | `constDeclarationShadowedByVarDeclaration.ts` | other | TS2481 |
| 173 | `selfReferencesInFunctionParameters.ts` | functions | TS2372 |
| 174 | `baseExpressionTypeParameters.ts` | generics | TS2562 |
| 175 | `propertyOrdering.ts` | objects | TS2301, TS2339, TS2551 |
| 176 | `exportSpecifierReferencingOuterDeclaration1.ts` | modules | TS2661 |
| 177 | `readonlyTupleAndArrayElaboration.ts` | arrays | TS2322, TS2345, TS4104 |
| 178 | `bitwiseCompoundAssignmentOperators.ts` | operators | TS2362, TS2363, TS2447 |
| 179 | `callOnClass.ts` | classes | TS2348 |
| 180 | `keywordExpressionInternalComments.ts` | other | TS2790 |
| 181 | `noImplicitAnyDestructuringParameterDeclaration.ts` | functions | TS7006, TS7008, TS7031 |
| 182 | `propertyAccessibility1.ts` | objects | TS2341 |
| 183 | `internalImportInstantiatedModuleNotReferencingInstance.ts` | modules | TS2437 |
| 184 | `bindingPatternCannotBeOnlyInferenceSource.ts` | generics | TS2339, TS2488, TS2571 |
| 185 | `restParamsWithNonRestParams.ts` | arrays | TS2555 |
| 186 | `superInConstructorParam1.ts` | classes | TS2336, TS2377, TS17011 |
| 187 | `deleteReadonlyInStrictNullChecks.ts` | other | TS2704 |
| 188 | `uncalledFunctionChecksInConditional.ts` | functions | TS2774 |
| 189 | `circularModuleImports.ts` | modules | TS2303 |
| 190 | `interfaceMayNotBeExtendedWitACall.ts` | objects | TS2499 |
| 191 | `extendedInterfacesWithDuplicateTypeParameters.ts` | generics | TS2300, TS2428 |
| 192 | `constructorOverloads1.ts` | classes | TS2392, TS2769 |
| 193 | `capturedParametersInInitializers1.ts` | functions | TS2373 |
| 194 | `reachabilityChecks1.ts` | other | TS7027 |
| 195 | `inheritanceMemberAccessorOverridingProperty.ts` | objects | TS2611 |
| 196 | `typeArgumentDefaultUsesConstraintOnCircularDefault.ts` | generics | TS2353, TS2744 |
| 197 | `classExtendsInterfaceInModule.ts` | modules | TS2689 |
| 198 | `shadowPrivateMembers.ts` | classes | TS2415 |
| 199 | `getterMissingReturnError.ts` | functions | TS2378 |
| 200 | `instantiationExpressionErrorNoCrash.ts` | other | TS2344, TS2635 |
| 201 | `thisPredicateInObjectLiteral.ts` | objects | TS2526 |
| 202 | `reservedNameOnInterfaceImport.ts` | modules | TS2438 |
| 203 | `incrementOnTypeParameter.ts` | generics | TS2356 |
| 204 | `classExpressionExtendingAbstractClass.ts` | classes | TS2653 |
| 205 | `noImplicitReturnsInAsync2.ts` | functions | TS7030 |
| 206 | `recursiveResolveTypeMembers.ts` | other | TS2304, TS2577 |
| 207 | `uniqueSymbolAllowsIndexInObjectWithIndexSignature.ts` | objects | TS2418 |
| 208 | `importAliasInModuleAugmentation.ts` | modules | TS2591, TS2667 |
| 209 | `genericSpecializations2.ts` | generics | TS2322, TS2368, TS2416 |
| 210 | `illegalSuperCallsInConstructor.ts` | classes | TS2337, TS2377, TS2564 |
| 211 | `functionParameterArityMismatch.ts` | functions | TS2554, TS2575 |
| 212 | `duplicateIdentifierDifferentModifiers.ts` | other | TS2687 |
| 213 | `inheritanceMemberPropertyOverridingAccessor.ts` | objects | TS2610 |
| 214 | `ambientExternalModuleWithRelativeModuleName.ts` | modules | TS2436 |
| 215 | `classImplementsPrimitive.ts` | classes | TS2864 |
| 216 | `functionTypeArgumentArityErrors.ts` | functions | TS2558, TS2743 |
| 217 | `fallFromLastCase2.ts` | other | TS7029 |
| 218 | `shorthandPropertyUndefined.ts` | objects | TS18004 |
| 219 | `classImplementsClass4.ts` | classes | TS2720, TS2741 |
| 220 | `bigIntWithTargetLessThanES2016.ts` | other | TS2791 |
| 221 | `unusedParameterProperty2.ts` | functions | TS6133, TS6138 |
| 222 | `staticPrototypeProperty.ts` | objects | TS2300, TS2699 |
| 223 | `instanceofOnInstantiationExpression.ts` | other | TS2848 |
| 224 | `newFunctionImplicitAny.ts` | functions | TS7009 |
| 225 | `returnInConstructor1.ts` | classes | TS2322, TS2409, TS2564 |
| 226 | `assigningFromObjectToAnythingElse.ts` | objects | TS2558, TS2696 |
| 227 | `classFieldSuperNotAccessible.ts` | classes | TS2855 |
| 228 | `inKeywordAndUnknown.ts` | other | TS2638 |
| 229 | `awaitCallExpressionInSyncFunction.ts` | functions | TS2311 |
| 230 | `parameterPropertyInConstructor3.ts` | classes | TS2398 |
| 231 | `circularReferenceInReturnType.ts` | functions | TS7022, TS7024 |
| 232 | `neverNullishThroughParentheses.ts` | other | TS2869 |
| 233 | `noImplicitAnyNamelessParameter.ts` | functions | TS7051 |
| 234 | `superCallFromClassThatHasNoBaseType1.ts` | classes | TS2335 |
| 235 | `reachabilityChecks3.ts` | other | TS7028 |
| 236 | `returnValueInSetter.ts` | functions | TS2408 |
| 237 | `gettersAndSettersAccessibility.ts` | other | TS2808 |
| 238 | `typeParameterAsBaseClass.ts` | classes | TS2304, TS2422 |
| 239 | `expressionWithJSDocTypeArguments.ts` | other | TS8020, TS17019, TS17020 |
| 240 | `indexedAccessPrivateMemberOfGenericConstraint.ts` | classes | TS2564, TS4105 |
| 241 | `implicitAnyGetAndSetAccessorWithAnyReturnType.ts` | functions | TS7006, TS7032 |
| 242 | `constDeclarations-access2.ts` | other | TS2588 |
| 243 | `classExtendsNull3.ts` | classes | TS2531 |
| 244 | `modularizeLibrary_ErrorFromUsingES6FeaturesWithOnlyES5Lib.ts` | other | TS2304, TS2339, TS2550, TS2583, TS2585 |
| 245 | `inheritanceMemberFuncOverridingAccessor.ts` | other | TS2416, TS2426 |
| 246 | `noCrashOnMixin.ts` | other | TS2674 |
| 247 | `missingDomElements.ts` | other | TS2339, TS2812 |
| 248 | `relationComplexityError.ts` | other | TS2859 |
| 249 | `mergedClassNamespaceRecordCast.ts` | namespaces | TS2339, TS2352 |
| 250 | `parseInvalidNonNullableTypes.ts` | syntax | TS2355, TS17019, TS17020 |
| 251 | `decoratorMetadataGenericTypeVariable.ts` | decorators | TS2304, TS2564 |
| 252 | `unusedFunctionsinNamespaces3.ts` | namespaces | TS6133 |
| 253 | `correctOrderOfPromiseMethod.ts` | async | TS2352 |
| 254 | `parseTypes.ts` | syntax | TS2322, TS2352 |
| 255 | `keyRemappingKeyofResult.ts` | types | TS2322 |
| 256 | `namespaceDisambiguationInUnion.ts` | namespaces | TS2322 |
| 257 | `generatorES6_5.ts` | async | TS2304 |
| 258 | `decoratorMetadataGenericTypeVariableInScope.ts` | decorators | TS2304, TS2564 |
| 259 | `conditionalAnyCheckTypePicksBothBranches.ts` | types | TS2322 |
| 260 | `unusedFunctionsinNamespaces6.ts` | namespaces | TS6133 |
| 261 | `promiseIdentity2.ts` | async | TS2403 |
| 262 | `decoratorMetadataConditionalType.ts` | decorators | TS2564 |
| 263 | `keyofIsLiteralContexualType.ts` | types | TS2322, TS2339 |
| 264 | `unusedVariablesinNamespaces3.ts` | namespaces | TS6133 |
| 265 | `promiseIdentity.ts` | async | TS2403 |
| 266 | `decoratorMetadataGenericTypeVariableDefault.ts` | decorators | TS2304, TS2564 |
| 267 | `reverseMappedTypeContextualTypeNotCircular.ts` | types | TS2322 |
| 268 | `incorrectNumberOfTypeArgumentsDuringErrorReporting.ts` | literals | TS2559 |
| 269 | `unusedFunctionsinNamespaces5.ts` | namespaces | TS6133 |
| 270 | `promiseChaining1.ts` | async | TS2322 |
| 271 | `enumUsedBeforeDeclaration.ts` | enums | TS2450 |
| 272 | `conditionalExpressionNewLine3.ts` | types | TS2304 |
| 273 | `stringMappingAssignability.ts` | literals | TS2322 |
| 274 | `ipromise3.ts` | async | TS2454 |
| 275 | `unusedInterfaceinNamespace1.ts` | namespaces | TS6196 |
| 276 | `intersectionsAndOptionalProperties.ts` | unions | TS2322 |
| 277 | `useBeforeDeclaration_destructuring.ts` | control | TS2448, TS2454 |
| 278 | `typeOfOperator1.ts` | types | TS2322 |
| 279 | `enumAssignmentCompat.ts` | enums | TS2322 |
| 280 | `unusedClassesinNamespace2.ts` | namespaces | TS6196 |
| 281 | `doubleUnderStringLiteralAssignability.ts` | literals | TS2322 |
| 282 | `ipromise4.ts` | async | TS2322 |
| 283 | `errorsWithInvokablesInUnions01.ts` | unions | TS2322 |
| 284 | `exhaustiveSwitchCheckCircularity.ts` | control | TS2345 |
| 285 | `numberAssignableToEnumInsideUnion.ts` | enums | TS2454 |
| 286 | `conditionalExpressionNewLine9.ts` | types | TS2304 |
| 287 | `unusedFunctionsinNamespaces2.ts` | namespaces | TS6133 |
| 288 | `numberOnLeftSideOfInExpression.ts` | literals | TS2454 |
| 289 | `awaitedTypeStrictNull.ts` | async | TS2589, TS7010 |
| 290 | `contextuallyTypingOrOperator.ts` | operators | TS2872 |
| 291 | `nonexistentPropertyOnUnion.ts` | unions | TS2339 |
| 292 | `scopingInCatchBlocks.ts` | control | TS2304 |
| 293 | `augmentedTypesEnum2.ts` | enums | TS2567 |
| 294 | `mappedTypeWithCombinedTypeMappers.ts` | types | TS2322 |
| 295 | `unusedFunctionsinNamespaces4.ts` | namespaces | TS6133 |
| 296 | `numberToString.ts` | literals | TS2322, TS2345 |
| 297 | `promiseIdentityWithAny.ts` | async | TS2403 |
| 298 | `destructuringAssignmentWithDefault2.ts` | operators | TS2322 |
| 299 | `contextualTypingOfArrayLiterals1.ts` | arrays | TS2322 |
| 300 | `errorMessageOnIntersectionsWithDiscriminants01.ts` | unions | TS2322 |

Feature buckets are selection heuristics based on filenames, not an exhaustive
AST feature inventory. The full input bytes remain the diagnostic authority.
