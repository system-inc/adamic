# Class acceptance row census

106 source rows: 2 IR-only, 30 accept, 74 refuse.

Scope: all internal/lower/class_*_test.go plus parameter_properties_test.go and its local-helper guard parameter_property_guard_test.go. Rows count source programs, including table entries, not helper call sites. No golden snapshots or rows deleted. The existing static guard file is inventoried but excluded from conversion by the brief. No local source/backend Node helper remains in this scope. Native layout assertions are retained with behavior-limit comments. No new oracle fixtures, so counts.md needs no refresh.

| File | Test | Row | Class | What changed | Why |
|---|---|---|---|---|---|
| class_features_test.go | TestClassFeaturesReadonlyChecker | external write | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesReadonlyChecker | method write | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesReadonlyChecker | derived constructor write | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesReadonlyChecker | static block write | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesReadonlyChecker | private write | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesReadonlyChecker | mutable contents through readonly field | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_features_test.go | TestClassFeaturesPrivateChecker | outside private field and method access | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesPrivateStorage | source | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Object.keys and spread observe private metadata; remove redundant IR visibility count. |
| class_features_test.go | TestClassFeaturesAccessorRefusals | narrow setter override | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesAccessorRefusals | getter replaces getter/setter | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesAccessorRefusals | narrow literal setter | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesAccessorRefusals | throwing spread getter | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticDeclarationsExecute | static side effect | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_features_test.go | TestClassFeaturesStaticDeclarationsExecute | static field and block | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_features_test.go | TestClassFeaturesStaticSoundness | early method read | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticSoundness | early helper read | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticSoundness | early helper alias read | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticSoundness | early Array.from callback | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticSoundness | narrow method override | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticSoundness | narrow field override | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticSoundness | constructor self-cycle | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticSoundness | constructor escapes during block | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticSoundness | parent constructor cycle | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticSoundness | private constructor self-cycle | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesAccessorCaptureCycle | source | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesNarrowedAccessor | source | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticParentCycle | source | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_features_test.go | TestClassFeaturesStaticInterfaceCycle | source | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesUnsoundOverrides | parameter bivariance | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesUnsoundOverrides | mutable field narrowing | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesUnsoundOverrides | mutable parameter contents | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesUnsoundOverrides | readonly mutable contents | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesUnsoundOverrides | inherited private cycle | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesUnsoundOverrides | inherited cycle | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceKeepsCheckerConstructorRules | abstract construction | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceKeepsCheckerConstructorRules | this before super | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceKeepsCheckerConstructorRules | abstract method missing | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceAllowsSoundOverrides | source | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_inheritance_test.go | TestInheritanceHasClassIdentity | source | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Replace ancestry IR assertion with instanceof observations. |
| class_inheritance_test.go | TestInheritanceCycleFinderIncludesInheritedFields | source | IR-only | retained internal assertion | No acceptance claim: inspects cycle-finder traversal, before Lower. |
| class_inheritance_test.go | TestInheritanceRejectsUnsupportedConstructorShapes | replacement object | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRejectsUnsupportedConstructorShapes | union dispatch | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRejectsUnsupportedConstructorShapes | computed base | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRejectsUnsupportedConstructorShapes | declare field | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesFalseNominalViews | weak structural | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesFalseNominalViews | source union | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesFalseNominalViews | target union | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesFalseNominalViews | nested array | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesFalseNominalViews | unrelated override | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesFalseNominalViews | structural object | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesFalseNominalViews | unrelated class | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceRefusesThisBeforeSuperReturns | source | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceKeepsNominalTupleDestructuring | source | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_inheritance_test.go | TestInheritanceGenericMonomorphizations | source | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_inheritance_test.go | TestInheritanceRefusesGrowingGenericClasses | source | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceGenericSoundness | generic narrowing | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceGenericSoundness | generic inherited cycle | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceConditionalThisRules | conditional missing super | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceConditionalThisRules | super method before super | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceConditionalThisRules | this in constructor default | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceGenericViewsKeepNominalArguments | source | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceGenericNominalConstraints | class constraint | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceGenericNominalConstraints | function constraint | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceGenericFactoryLayouts | source | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_inheritance_test.go | TestInheritanceNativeSignatureNeighbors | identical default | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_inheritance_test.go | TestInheritanceNativeSignatureNeighbors | wider same representation | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_inheritance_test.go | TestInheritanceNativeSignatureNeighbors | identical optional | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_inheritance_test.go | TestInheritanceNativeSignatureNeighbors | void result | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Add overridden method output to the previously silent row. |
| class_inheritance_test.go | TestInheritanceNativeSignatureLimits | optional added | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceNativeSignatureLimits | boolean default added | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceNativeSignatureLimits | generic default added | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceNativeSignatureLimits | parameter count | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceNativeSignatureLimits | void to number result | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_inheritance_test.go | TestInheritanceNativeSignatureLimits | number to optional result | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | separate literals | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | property order | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | nested arrays | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | tuples | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | unions | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | functions | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | recursive interfaces | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | different fields | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | optional field | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_instance_key_test.go | TestClassArgumentsUseCheckerIdentity | same representation | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. Keep native-layout assertions: Node cannot observe layout count/types or erased-definition sharing. |
| class_static_guard_test.go | TestClassStaticInitializerCallIsEmitted | source | IR-only | retained internal assertion | Already-converted file excluded by brief; retain its separate initializer adjacency/call-count guard unchanged. |
| class_static_guard_test.go | TestClassStaticSideEffectMatchesNode | source | accept | already converted; unchanged | Compare computed stdout and exit with source Node. Existing native sanitizer comparison preserved. |
| class_wrong_output_test.go | TestClassWrongOutputIteratorReceiver | for-of override | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_wrong_output_test.go | TestClassWrongOutputIteratorReceiver | spread override | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_wrong_output_test.go | TestClassWrongOutputIteratorReceiver | Array.from base view | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_wrong_output_test.go | TestClassWrongOutputIteratorReceiver | destructure override | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| class_wrong_output_test.go | TestClassWrongOutputIteratorReceiver | base spread | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_wrong_output_test.go | TestClassWrongOutputIteratorReceiver | unchanged subclass spread | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_wrong_output_test.go | TestClassWrongOutputKeysRepair | source | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| class_wrong_output_test.go | TestClassWrongOutputPrivateRepair | source | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| parameter_properties_test.go | TestParameterPropertiesSoundness | readonly view | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_properties_test.go | TestParameterPropertiesSoundness | mutable override | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_properties_test.go | TestParameterPropertiesSoundness | readonly override | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_properties_test.go | TestParameterPropertiesSoundness | inherited initializer | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_properties_test.go | TestParameterPropertiesSoundness | early default | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_properties_test.go | TestParameterPropertiesSoundness | field initializer | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_properties_test.go | TestParameterPropertyCheckerContracts | readonly external write | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_properties_test.go | TestParameterPropertyCheckerContracts | private external read | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_properties_test.go | TestParameterPropertyCheckerContracts | protected external read | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_properties_test.go | TestParameterPropertyCheckerContracts | public write | accept | lowersAndAgreesWithNode | Compare computed stdout and exit with source Node. |
| parameter_properties_test.go | TestParameterPropertyCallbackReceiver | source | refuse | unchanged | Preserve checker, Refused, or NotYet rejection and diagnostic. |
| parameter_property_guard_test.go | TestParameterPropertyValueMatchesNode | parameterPropertyValueSource | accept | consolidate local Node comparison into lowersAndAgreesWithNode | Preserve native ASan/UBSan stdout check; shared helper adds exit comparison and empty-answer refusal. |
