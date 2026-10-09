# agree-rest row census

Source rows on base 571e74cf555b9db994c5dec6c2f8dbee676e5111. Numbered rows use original table order. Refuse includes NotYet and direct proof diagnostics. IR-only means assertions of internal facts, including proof analysis. Helper definitions and shared source prefixes are not rows. Fixture extensions are separate source rows.

{'accept': 71, 'refuse': 194, 'IR-only': 33}; total 298.

Scope is the eight named files; helper-copy examples outside this territory were not edited. No new oracle fixture, so counts.md needs no refresh. Skipped conversions preserve every original row and assertion.

| file | test | row | class | what I did | why |
|---|---|---|---|---|---|
| array_predicate_test.go | TestArrayPredicatePreservesDeclaredElementContract | 1 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| array_predicate_test.go | TestArrayPredicatePreservesDeclaredElementContract | 2 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| array_predicate_test.go | TestArrayPredicatePreservesDeclaredElementContract | 3 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| array_predicate_test.go | TestArrayPredicatePreservesDeclaredElementContract | 4 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| array_predicate_test.go | TestArrayPredicatePreservesDeclaredElementContract | 5 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| array_predicate_test.go | TestArrayPredicateCannotInventAnElementContract | 1 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestArrayPredicateCannotInventAnElementContract | 2 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestArrayPredicateCannotInventAnElementContract | 3 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestArrayPredicateCannotInventAnElementContract | 4 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestArrayPredicateCannotInventAnElementContract | 5 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestArrayPredicateCannotInventAnElementContract | 6 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestArrayPredicateCannotInventAnElementContract | 7 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestArrayPredicateDoesNotMisclassifyNativeTuples | 1 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestArrayPredicateDoesNotMisclassifyNativeTuples | 2 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestUnknownArrayPredicateRefusesUnrepresentedObservations | 1 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestUnknownArrayPredicateRefusesUnrepresentedObservations | 2 | refuse | unchanged | Pinned lowering or proof refusal |
| array_predicate_test.go | TestArrayPredicateCoexistsWithUnknownReflection | 1 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| array_predicate_test.go | TestArrayPredicateCoexistsWithUnknownReflection | 2 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| array_predicate_test.go | TestArrayPredicateCoexistsWithUnknownReflection | 3 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| array_predicate_test.go | TestArrayPredicateCoexistsWithUnknownReflection | 4 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| array_predicate_test.go | TestArrayPredicateCoexistsWithUnknownReflection | 5 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| proven_relations_test.go | TestProvenRelationsRefuse | satisfies mutable array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | upcast mutable array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | satisfies mutable field | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | satisfies nested mutable literal | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | satisfies readonly to writable | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | satisfies function parameter | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | upcast function parameter | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | upcast nominal identity | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | satisfies nominal identity | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | satisfies hidden optional | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | upcast hidden optional | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | upcast mutable optional elements | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | satisfies mutable optional field | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | upcast invariant optional class argument | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsRefuse | satisfies nested optional | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| proven_relations_test.go | TestProvenRelationsErase | plain operand | accept | agreement | Preserve operand behavior; removed emitted artifact equality snapshots; runtime cost no longer claimed |
| proven_relations_test.go | TestProvenRelationsErase | satisfies relation | IR-only | agreement | Preserve operand behavior; removed emitted artifact equality snapshots; runtime cost no longer claimed |
| proven_relations_test.go | TestProvenRelationsErase | as relation | IR-only | agreement | Preserve operand behavior; removed emitted artifact equality snapshots; runtime cost no longer claimed |
| arguments_length_test.go | TestArgumentsLengthRefusals | indexing | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | aliasing | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | passing | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | returning | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | spreading | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | writing object | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | writing length | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | incrementing length | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | compound length | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | arrow | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthRefusals | parenthesized write | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| arguments_length_test.go | TestArgumentsLengthReadNeighbors | 1 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| arguments_length_test.go | TestArgumentsLengthReadNeighbors | 2 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| arguments_length_test.go | TestArgumentsLengthReadNeighbors | 3 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| arguments_length_test.go | TestArgumentsLengthReadNeighbors | 4 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| arguments_length_test.go | TestArgumentsLengthReadNeighbors | 5 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | indexing.a | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | indexing.ts | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | aliasing.a | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | aliasing.ts | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | passing.a | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | passing.ts | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | returning.a | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | returning.ts | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | spreading.a | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | spreading.ts | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | writing.a | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | writing.ts | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | arrow.a | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthRefusalFixtures | arrow.ts | refuse | unchanged | Existing fixture copied under each extension; pinned diagnostic |
| arguments_length_test.go | TestArgumentsLengthReadKeepsReaderFact | 1 | IR-only | agreement plus reader convention assertion | Zero-argument behavior cannot observe hidden calling convention |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | complete | accept | agreement | Execute checked view with valid payload; optional target now prints its tag |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | missing payload | accept | skipped conversion | Malformed payload or tag intentionally admitted; checked read can panic while source Node succeeds |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | wrong payload | accept | skipped conversion | Malformed payload or tag intentionally admitted; checked read can panic while source Node succeeds |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | broad tag | accept | skipped conversion | Malformed payload or tag intentionally admitted; checked read can panic while source Node succeeds |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | alias write | accept | agreement | Execute checked view with valid payload; optional target now prints its tag |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | unused bad factory | accept | agreement | Execute checked view with valid payload; optional target now prints its tag |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | different tag | accept | skipped conversion | Malformed payload or tag intentionally admitted; checked read can panic while source Node succeeds |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | dynamic complete | accept | agreement | Execute checked view with valid payload; optional target now prints its tag |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | optional target | accept | agreement | Execute checked view with valid payload; optional target now prints its tag |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | optional target read | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | spread | accept | agreement | Execute checked view with valid payload; optional target now prints its tag |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | reflection alias | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | generic | accept | agreement | Execute checked view with valid payload; optional target now prints its tag |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | staged class | accept | skipped conversion | Malformed payload or tag intentionally admitted; checked read can panic while source Node succeeds |
| interface_cast_test.go | TestDefaultTaggedInterfaceAdmission | forged payload | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| interface_cast_test.go | TestDefaultTaggedInterfaceNeedsNoFlag | 1 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| definite_assignment_test.go | TestDefiniteAssignmentUsesReadiness | 1 | IR-only | skipped conversion; preserve readiness assertion | Readiness metadata after initialization is not observable; uninitialized reads deliberately diverge from unchecked source NaN |
| definite_assignment_test.go | TestDefiniteAssignmentUsesReadiness | 2 | IR-only | skipped conversion; preserve readiness assertion | Readiness metadata after initialization is not observable; uninitialized reads deliberately diverge from unchecked source NaN |
| definite_assignment_test.go | TestDefiniteAssignmentUsesReadiness | 3 | IR-only | agreement plus readiness assertion | Readiness metadata after initialization is not observable; uninitialized reads deliberately diverge from unchecked source NaN |
| definite_assignment_test.go | TestDefiniteAssignmentUsesReadiness | 4 | IR-only | agreement plus readiness assertion | Readiness metadata after initialization is not observable; uninitialized reads deliberately diverge from unchecked source NaN |
| definite_assignment_test.go | TestDefiniteAssignmentSoundNeighbors | 1 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| definite_assignment_test.go | TestDefiniteAssignmentSoundNeighbors | 2 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| definite_assignment_test.go | TestDefiniteAssignmentSoundNeighbors | 3 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| definite_assignment_test.go | TestDefiniteAssignmentSoundNeighbors | 4 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| census_small_test.go | TestCensusOverloadRelation | parameter contravariance | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| census_small_test.go | TestCensusOverloadRelation | generic parameter constraints | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| census_small_test.go | TestCensusOverloadRelation | mutable parameters | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| census_small_test.go | TestCensusOverloadRelation | result covariance | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| census_small_test.go | TestCensusSmallFiniteKeyRead | 1 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| census_small_test.go | TestCensusRestMutableElements | 1 | refuse | unchanged | Pinned lowering or proof refusal |
| census_small_test.go | TestCensusBooleanDeadBranch | 1 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| predicates_proof_test.go | TestPredicateBodyProof | condition assertion | IR-only | preserved with behavior-unobservable comment | Proof summary and checked-field handoff, no full lowered executable |
| predicates_proof_test.go | TestPredicateBodyProof | returning fail | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | empty condition assertion | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | condition mutation | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | diagnostic work on failure | IR-only | preserved with behavior-unobservable comment | Proof summary and checked-field handoff, no full lowered executable |
| predicates_proof_test.go | TestPredicateBodyProof | typeof | IR-only | preserved with behavior-unobservable comment | Proof summary and checked-field handoff, no full lowered executable |
| predicates_proof_test.go | TestPredicateBodyProof | kind | IR-only | preserved with behavior-unobservable comment | Proof summary and checked-field handoff, no full lowered executable |
| predicates_proof_test.go | TestPredicateBodyProof | kind helpers fixed point | IR-only | preserved with behavior-unobservable comment | Proof summary and checked-field handoff, no full lowered executable |
| predicates_proof_test.go | TestPredicateBodyProof | helper | IR-only | preserved with behavior-unobservable comment | Proof summary and checked-field handoff, no full lowered executable |
| predicates_proof_test.go | TestPredicateBodyProof | false branch lies | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | true branch lies | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | mutation | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | alias mutation | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | recursive | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | default changes input | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | empty assertion | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestPredicateBodyProof | assertion | IR-only | preserved with behavior-unobservable comment | Proof summary and checked-field handoff, no full lowered executable |
| predicates_proof_test.go | TestPredicateBodyProof | disabled assertion | refuse | unchanged | Pinned predicate proof diagnostic |
| predicates_proof_test.go | TestConditionAssertionAdmission | 1 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| predicates_proof_test.go | TestPredicateCallbackContracts | inferred arrow | accept | agreement with printed callback result | Execute original predicate callback application |
| predicates_proof_test.go | TestPredicateCallbackContracts | inferred opaque | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| predicates_proof_test.go | TestPredicateCallbackContracts | inferred named | accept | agreement with printed callback result | Execute original predicate callback application |
| predicates_proof_test.go | TestPredicateCallbackContracts | arrow | accept | agreement with printed callback result | Execute original predicate callback application |
| predicates_proof_test.go | TestPredicateCallbackContracts | named | accept | agreement with printed callback result | Execute original predicate callback application |
| predicates_proof_test.go | TestPredicateCallbackContracts | unproven | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| predicates_proof_test.go | TestPredicateCallbackContracts | lying | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| predicates_proof_test.go | TestPredicateCallbackContracts | reassigned | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| predicates_proof_test.go | TestEveryNeedsCallbackEffects | 1 | refuse | unchanged | Pinned lowering or proof refusal |
| predicates_proof_test.go | TestPredicateOverloadCallback | 1 | accept | agreement | Execute original computation; silent accepted functions now called and printed |
| predicates_proof_test.go | TestPredicateOverloadCallback | 2 | refuse | unchanged | Pinned lowering or proof refusal |
| predicates_proof_test.go | TestPredicateUseRegions | unused | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | true | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | false | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | both | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | return continuation | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | throw continuation | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | false continuation | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | const result alias | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | compound result alias | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | short circuit | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | closure read | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | indexed spelling | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | indexed argument | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | computed index | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | loop read | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | later read | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUseRegions | replacement | IR-only | preserved with behavior-unobservable comment | Assert per-call proof use directions, not executable acceptance |
| predicates_proof_test.go | TestPredicateUsesBelongToEachCall | 1 | IR-only | preserved with behavior-unobservable comment | Assert which proof use direction belongs to each call |
| lower_test.go | TestConsoleLowersToWriteLine | 1 | IR-only | skipped conversion; preserve existing assertions | Existing stream and source metadata probe emits nonempty success stderr, refused by common helper; no new golden |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a class inside a function | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | join on an array of functions | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | concat on an array of arrays | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | an arrow returning never | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | Array.from of an array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | Array.from without a callback | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | Array.from with a named callback | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | assigning Array.from's undefined | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | an array of boolean \| undefined | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a captured boolean \| undefined | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a class field of a union | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a template of a union with an object | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | an array of targets seen as an array of Weak | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a function value capturing what it initializes | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a tuple passed where an array goes | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a tuple given to a function value taking an array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | tuples where arrays of arrays go | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a callback taking arrays given tuples | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a function value called through ?. | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a generic function as a value | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a try around repeat | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a try around a call that reaches toFixed | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | rethrowing a mutable stored Error | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | an object seen with a field weak in one view only | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a function seen with a parameter weak in one view only | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a function seen with a result weak in one view only | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a spread seen with a field weak in one view only | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | a template interpolating an object | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | var | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | async | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | throwing a string | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | throwing a number | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | ! | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | == | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | delete | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a constructor calling a method before its fields are set (R2's, and R's half_built) | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a constructor handing this out before its fields are set (R's early_this) | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a constructor storing this before its fields are set | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a constructor storing this, a field set only in a branch | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a filter that decides by truthiness | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | Array.from's undefined typed as a number | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a parent pointer not declared weak | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a doubly linked list with a strong next | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a closure in a cell it captures | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a mutable array of children | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a callback that captures what holds it | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a map whose values reach back | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a map whose keys reach back | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a set whose elements reach back | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a map of sets that reach back | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a cycle through a subtype | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a cycle only a later instantiation closes | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a cycle a generic class makes but never names | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a generic cell holding a function that captures it | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | reduce without an initial value | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | Math.random | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | Math.random, not called | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | a method read off its object | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | this.method read as a value | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestWhatZeroOneRefusesIsRefusedWithAFix | an unchecked cast | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an array seen as a wider array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a mutable field seen as a wider field | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a map's values seen wider, passed | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a function seen as taking a narrower array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an array of arrays seen wider behind readonly | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a type parameter's constraint: a Narrow into a Pack slot | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an explicit type argument widens a mutable parameter | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an inferred type argument makes a generic write unsafe | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a type parameter's constraint behind a readonly property | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a type parameter whose constraint is narrower, seen as the wider array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an intersection's array seen as an intersection's | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an array of targets seen as an array of Weak | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an object seen with a mutable field weak in one view only | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a readonly field turned back into a writable one | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a readonly field made writable by a mapped type | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a readonly array field turned writable | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a readonly field turned writable inside a readonly array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a readonly field of a primitive turned writable | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a shorthand property | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a parameter's default | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a class field's initializer | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an object spread | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an as | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a union target | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a union source | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a conditional tsc reduced to the wider branch | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a literal's element tsc reduced to the wider | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a return tsc reduced to the wider | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a method's return | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a method's parameter, bivariant in tsc | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a tuple's later element | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a class target | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a plain object seen as a class | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a fresh copy's elements seen wider | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an annotated destructuring | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | an intersection's array seen as a plain array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a destructuring assignment's element seen wider | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMutableLocationSeenWiderIsRefused | a destructuring assignment's element seen wider inside a readonly array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a destructuring declaration of a readonly field | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a destructuring assignment's elements going into wider names | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | an array seen as a readonly wider array | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a readonly field seen wider | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a readonly constraint: a Narrow into a Pack slot | accept | skipped conversion | Unused generic or intersection body; original partial admission intentionally permits NotYet; preserve original source |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a type parameter as itself | accept | skipped conversion | Unused generic or intersection body; original partial admission intentionally permits NotYet; preserve original source |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a type parameter narrowed from undefined, as itself | accept | skipped conversion | Unused generic or intersection body; original partial admission intentionally permits NotYet; preserve original source |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | an intersection's array seen as a readonly array with the same tag | accept | skipped conversion | Unused generic or intersection body; original partial admission intentionally permits NotYet; preserve original source |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a readonly field seen as a readonly field | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a writable field seen as a readonly one | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a readonly field seen as its own type | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a readonly field copied into a writable one | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a copy made by slice | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a copy made by map | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a copy made by filter | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a conditional of fresh arrays | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a Map made from pairs | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a conditional of fresh copies | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a new Map passed | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a new Map | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a union target that can't write | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a shorthand property that can't write | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a destructuring with no type | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a method returning the same type | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | a class seen as itself | accept | agreement | Execute sound view and print its value, length, or contents |
| lower_test.go | TestAViewThatCantWriteIsNotRefused | an intersection seen as itself | accept | skipped conversion | Unused generic or intersection body; original partial admission intentionally permits NotYet; preserve original source |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | an initializer | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | a cast literal | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | an assignment | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | an argument | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | a return | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | a field | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | an element | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | an arrow's body | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | a map's value | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | a slot that may be undefined | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | an array of tuples seen as an array of arrays | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | a tuple field seen as an array field | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | a function returning a tuple seen as returning an array | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestATupleSeenAsAnArrayIsNotYet | an array method called on a tuple | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMethodReadAsAValueIsRefused | held in a variable | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMethodReadAsAValueIsRefused | passed as a callback | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMethodReadAsAValueIsRefused | in parentheses | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMethodReadAsAValueIsRefused | read through this | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMethodReadAsAValueIsRefused | an interface's method | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMethodReadAsAValueIsRefused | an array's method | refuse | unchanged | Pinned lowering diagnostic or NotYet |
| lower_test.go | TestAMethodReadAsAValueIsRefused | called on its object | accept | agreement | Execute original method call and print result |
| lower_test.go | TestAMethodReadAsAValueIsRefused | called through parentheses | accept | agreement | Execute original method call and print result |
| lower_test.go | TestAMethodReadAsAValueIsRefused | called in an arrow | accept | agreement | Execute original method call and print result |
| lower_test.go | TestAMethodReadAsAValueIsRefused | a field holding a function | accept | skipped conversion | Observed NotYet: method through view erases prototype origin; old assertion permits it |
