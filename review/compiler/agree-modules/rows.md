# Source-row census

{'refuse': 71, 'accept': 77, 'IR-only': 2}; total 150 source rows. Main-plus-import dependency is one entry row; proof-domain functions belong to one source row. Classification follows assertions, including numeric-enum acceptance in tests named as limits.

The shared enum runner in enum_agree_test.go remains because enums_open_test.go outside this unit still uses it. The enum_guards and statics_guards copies are removed. requireLoweredOutput remains for callers outside this unit. No new oracle fixtures, so counts.md needs no refresh. M12 remains held by the original paired TestParserFactoryBindingHoisting in namespaces_factory_test.go; the factory-output row alone does not observe NamespaceVar metadata.

| File | Test | Row | Class | What changed | Why |
| --- | --- | --- | --- | --- | --- |
| namespaces_test.go | TestNamespaceLimitsStayLoud | computed enum | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | reopened enum | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | early enum | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | mutable object escape | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | object value | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | reflection | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | reopening | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | function merge escape | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | class merge | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | destructured state | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | nested object var | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | rest object var | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | default object var | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | computed member | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | replace function | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | early call | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | early read | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | block var | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | nested early read | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | explicit receiver | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceLimitsStayLoud | ambient | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceReceiverRefusal | source | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceTypesErase | source | accept | shared agreement helper; named direct assertion retained | Erasure, allocation, ownership or unique native slots cannot be observed by JavaScript. |
| namespaces_test.go | TestTracingNamespaceEscapeStaysNotYet | source | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestDebugNamespaceMergedCapabilities | debug_log.a | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| namespaces_test.go | TestDebugNamespaceMergedCapabilities | class_merge.a | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestDebugNamespaceMergedCapabilities | debug_class.a | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| namespaces_test.go | TestNamespaceReturnedAssignmentLimits | 1 | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceReturnedAssignmentLimits | 2 | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceInitializationReachability | direct | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceInitializationReachability | helper | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceInitializationReachability | cycle | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceClosedCallGraphEdges | 1 | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceClosedCallGraphEdges | 2 | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceClosedCallGraphEdges | 3 | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceClosedCallGraphEdges | 4 | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestNamespaceClosedCallGraphEdges | 5 | refuse | unchanged | Explicit Refused or NotYet contract. |
| namespaces_test.go | TestEmptyNeverMapCannotGainWritableInhabitants | source | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedFunctionGapsAreLoud | block | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedFunctionGapsAreLoud | generic | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedFunctionGapsAreLoud | dynamic this | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedFunctionCycleIsRefused | source | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedEnvironmentHasOneAllocationSite | source | accept | shared agreement helper; named direct assertion retained | Erasure, allocation, ownership or unique native slots cannot be observed by JavaScript. |
| nested_functions_test.go | TestNestedEnvironmentCycleIncludesDisjointSlots | source | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedCapturedParametersAreOwned | source | accept | shared agreement helper; named direct assertion retained | Erasure, allocation, ownership or unique native slots cannot be observed by JavaScript. |
| nested_functions_test.go | TestClosedFrameInputRejectsMutation | source | accept | shared agreement helper; named direct assertion retained | Erasure, allocation, ownership or unique native slots cannot be observed by JavaScript. |
| nested_functions_test.go | TestNestedRebindingNotYet | source | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedCallbackCycleIsRefused | source | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedBodylessDeclarationsAreLoud | missing implementation | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedBodylessDeclarationsAreLoud | body lowering | refuse | unchanged | Explicit Refused or NotYet contract. |
| nested_functions_test.go | TestNestedRestIsSupported | source | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | variable | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | initializer | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | member initializer | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | argument | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | assignment | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | literal property | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | array literal | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | increment | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | compound | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | array alias | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | readonly into enum array | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | function view | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | object assign | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumSlotViews | override | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | class field | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | optional union | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | map alias | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | mutable field alias | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSlotViews | generic class invariant | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumSlotViews | structural numeric enum object | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumSlotViews | structural string enum object | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumSlotViews | structural enum object array | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumSlotViews | enum object writable view | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumSwitchExhaustiveness | 1 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumSwitchExhaustiveness | 2 | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSwitchExhaustiveness | 3 | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSwitchExhaustiveness | 4 | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumSwitchExhaustiveness | 5 | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enums_test.go | TestEnumLimitsStayLoud | 1 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 2 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 3 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 4 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 5 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 6 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 7 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 8 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 9 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 10 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 11 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 12 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 13 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 14 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestEnumLimitsStayLoud | 15 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enums_test.go | TestConstEnumErasesRuntimeObject | source | accept | shared agreement helper; named direct assertion retained | Erasure, allocation, ownership or unique native slots cannot be observed by JavaScript. |
| enums_test.go | TestEnumIdentityAcrossModules | source | refuse | unchanged | Explicit Refused or NotYet contract. |
| enum_flags_test.go | TestFlagEnumsOpen | complement | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | arithmetic | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | increment | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | shift_update | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | shift_result | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | sign_bit | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | implicit | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | numeric_shape | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | arithmetic_shape | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | switch_default | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | cross_enum_left | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | cross_enum_right | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | or_number | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | xor_number | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | and_numbers | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | narrowed_member | refuse | unchanged | Explicit Refused or NotYet contract. |
| enum_flags_test.go | TestFlagEnumsOpen | compound_or_number | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | compound_xor_number | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | mutable_alias | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | bare_number | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsOpen | member_slot | refuse | unchanged | Explicit Refused or NotYet contract. |
| enum_flags_test.go | TestFlagEnumsOpen | array_view | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | bitwise_updates | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | alias | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | or | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | xor | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | and_left | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | and_right | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | and_complement | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | nested | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | field | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | array | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | map | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | parameter | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | optional | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumsDomain | switch | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestEnumNeverDefault | source | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumLiteralSpellings | source | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumMemberAliases | source | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestEnumNameEnumeration | source | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumInlineIteration | source | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumAliasBoundaries | 1 | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumAliasBoundaries | 2 | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumAliasBoundaries | 3 | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumAliasBoundaries | 4 | accept | shared agreement helper; compute and print witness | Compares nonempty computed stdout and exit with source Node. |
| enum_flags_test.go | TestFlagEnumAliasBoundaries | 5 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enum_flags_test.go | TestFlagEnumAliasBoundaries | 6 | refuse | unchanged | Explicit Refused or NotYet contract. |
| enum_guards_test.go | TestEnumMemberValuesAndReverseNameMatchNode | source | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| enum_guards_test.go | TestEnumFlagProofsWithoutObservableLoweringEffect | literal flag proof | IR-only | retained named proof checks | Open numeric enum lowering bypasses flag classification; no emitted behavior changes. |
| enum_guards_test.go | TestEnumFlagProofsWithoutObservableLoweringEffect | domain proof (left/right/both/neither/orNumber/xorFlags) | IR-only | retained named proof checks | Open numeric enum lowering bypasses flag classification; no emitted behavior changes. |
| enum_guards_test.go | TestEnumReverseMappingUsesSingleSlot | source | accept | shared agreement helper; named direct assertion retained | Erasure, allocation, ownership or unique native slots cannot be observed by JavaScript. |
| statics_guards_test.go | TestPrivateAndPublicStaticsAgreeWithNode | source | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| statics_guards_test.go | TestNamespaceFactoryBindingsAgreeWithNode | source | accept | shared agreement helper | Compares nonempty computed stdout and exit with source Node. |
| statics_guards_test.go | TestReadinessErrorIncludesReceiverExpression | source | accept | skipped conversion; shared bounded Node runner | Checked-failure soundness probe: source returns NaN but Adamic exits 70. Preserve complete panic-message contract; ordinary agreement would reject it. |
