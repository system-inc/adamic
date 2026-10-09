Audited u033 at origin/main 8171b3173bdbfce1f7982d3c4f731279307ece37; all 15 requested names exist, none moved or vanished.
Verdicts: three sacred, ten subsumed, two overlapping; no families, helpers, witnesses or setup-check rows in this unit.
Twelve fixed production mutants ran over the whole package; two entry probes ran over the scoped rows.
Two rows are vacuous; 12 admitted subcases of the mixed admission row also pass its empty-answer probe.
Evidence is on test-audit/internal-lower-interface_cast under review/test-audit/internal-lower-interface_cast/; production code is restored.

CODE UNDER TEST, named before mutation: Adamic Go lowering. Fourteen rows enter Lower through lowerSource; TestRepresentedMethodReplacementIsNotYet directly enters lowering.setProperty after load, checker, declareModule and statements preparation. The reached-function inventory from clean slice coverage is code-under-test-functions.txt, 428 functions with nonzero coverage. scope.cover and scope-functions-all.txt preserve the full coverage evidence. The inventory also includes initialization functions; Go coverage identifies functions, with closure blocks retained in scope.cover.

ORACLE, named before mutation: self-written admission/refusal expectations, Refused/NotYet categories, and diagnostic substrings. No scoped row runs native output, Node, tsc, or Go cohere as an independent comparison. load.Load is preparation and was never mutated.

The starting SHA is newer than the brief's 8de93800f4 location reference. Scope was verified against go test -list . ./internal/lower/. Each body has its own assertions; sharing lowerSource does not make these tests one family.

All production matrix commands used ADAMIC_MUTANT=MID ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/MID timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > MID.log 2>&1. Every production run completed within 90 s, without a panic. All 239 top-level package Test names have observed pass/fail/skip results in matrix.json. Two out-of-scope project inventories skipped throughout, detailed below. Package uniqueness uses every observed failed top-level Test, not just the 15 scoped rows. Consequently bounded=false: the assignment is a slice, but this production matrix sees the whole package. No repository-wide uniqueness is claimed.

Each standalone diff applies to the starting origin/main and passed go vet ./internal/lower/. Fourteen apply checks passed after restoring production. switch.diff and switch-helper.go.txt preserve the scratch instrumentation. No tests, harness or oracle were mutated. No native product was built; unique build-cache directories were nevertheless set per mutant.

```json
[
  {
    "test": "TestDefaultTaggedInterfaceAdmission",
    "package": "internal/lower",
    "file": "internal/lower/interface_cast_test.go:14",
    "seconds": 0.446,
    "oracle": "Self-written admission for 12 sources and three diagnostic-substring refusals; accepted sources have no IR or runtime assertion.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M08"
    ],
    "unique_kills": [],
    "last_proven_fail": "M08 interface_cast_test.go:38: /tmp/adamic-gate/TestDefaultTaggedInterfaceAdmissionalias_write3603935960/001/main.a:4:85: stage 0 can't lower replacing a represented method at runtime yet",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestPredicateBodyProof",
      "TestEnumSlotViews"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M08 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M08.log 2>&1; interface_cast_test.go:38: /tmp/adamic-gate/TestDefaultTaggedInterfaceAdmissionalias_write3603935960/001/main.a:4:85: stage 0 can't lower replacing a represented method at runtime yet",
    "vacuous_subcases": [
      "complete",
      "missing payload",
      "wrong payload",
      "broad tag",
      "alias write",
      "unused bad factory",
      "different tag",
      "dynamic complete",
      "optional target",
      "spread",
      "generic",
      "staged class"
    ],
    "subsumption_mutants": null
  },
  {
    "test": "TestDefaultTaggedInterfaceNeedsNoFlag",
    "package": "internal/lower",
    "file": "internal/lower/interface_cast_test.go:48",
    "seconds": 0.082,
    "oracle": "Self-written nil-error admission; no IR, output, or environment-switch assertion.",
    "oracle_kind": "self",
    "kills": [
      "M01"
    ],
    "unique_kills": [],
    "last_proven_fail": "M01 interface_cast_test.go:52: default admission: /tmp/adamic-gate/TestDefaultTaggedInterfaceNeedsNoFlag1601758792/001/main.a:4:13: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation conversion yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDefaultTaggedInterfaceAdmission"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [],
    "subsumer_seconds": 0.446,
    "vacuous": true,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M01 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M01 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M01.log 2>&1; interface_cast_test.go:52: default admission: /tmp/adamic-gate/TestDefaultTaggedInterfaceNeedsNoFlag1601758792/001/main.a:4:13: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation conversion yet",
    "vacuous_subcases": [],
    "subsumption_mutants": 1
  },
  {
    "test": "TestIteratorGapsAreExplicit",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:14",
    "seconds": 0.31,
    "oracle": "Self-written NotYet error category only; accepts unrelated NotYet reasons, as M03 demonstrates.",
    "oracle_kind": "self",
    "kills": [
      "M04"
    ],
    "unique_kills": [],
    "last_proven_fail": "M04 iteration_test.go:30: got <nil>, want an explicit iterator NotYet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestIteratorViewsCannotEraseReceivers"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.067,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M04.log 2>&1; iteration_test.go:30: got <nil>, want an explicit iterator NotYet",
    "vacuous_subcases": [],
    "subsumption_mutants": 1
  },
  {
    "test": "TestIteratorViewsCannotHideReturn",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:36",
    "seconds": 0.079,
    "oracle": "Self-written NotYet plus hide-a-return substring.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04"
    ],
    "unique_kills": [],
    "last_proven_fail": "M04 iteration_test.go:45: got <nil>, want a hidden-return refusal",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestIteratorViewsCannotEraseReceivers"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.067,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M04.log 2>&1; iteration_test.go:45: got <nil>, want a hidden-return refusal",
    "vacuous_subcases": [],
    "subsumption_mutants": 2
  },
  {
    "test": "TestIteratorViewsCannotEraseReceivers",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:49",
    "seconds": 0.067,
    "oracle": "Self-written NotYet plus receiver-convention substring.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04"
    ],
    "unique_kills": [],
    "last_proven_fail": "M04 iteration_test.go:57: got <nil>, want an erased-receiver refusal",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestIteratorViewsCannotHideReturn"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.079,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M04.log 2>&1; iteration_test.go:57: got <nil>, want an erased-receiver refusal",
    "vacuous_subcases": [],
    "subsumption_mutants": 2
  },
  {
    "test": "TestIteratorDestructuringDoesNotLieAboutExhaustion",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:61",
    "seconds": 0.062,
    "oracle": "Self-written Refused plus default substring in Fix.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M12"
    ],
    "unique_kills": [
      "M12"
    ],
    "last_proven_fail": "M12 iteration_test.go:66: got <nil>, want a default or undefined fix",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M12 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M12.log 2>&1; iteration_test.go:66: got <nil>, want a default or undefined fix",
    "vacuous_subcases": [],
    "subsumption_mutants": null
  },
  {
    "test": "TestGeneratorsAreRefusedEvenWithoutYield",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:70",
    "seconds": 0.196,
    "oracle": "Self-written Refused plus suspended-frames substring in Fix, for three generator syntaxes.",
    "oracle_kind": "self",
    "kills": [
      "M05"
    ],
    "unique_kills": [
      "M05"
    ],
    "last_proven_fail": "M05 iteration_test.go:76: got <nil>, want the generator ownership refusal",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M05.log 2>&1; iteration_test.go:76: got <nil>, want the generator ownership refusal",
    "vacuous_subcases": [],
    "subsumption_mutants": null
  },
  {
    "test": "TestLiteralMethodCapturesCannotMakeCycles",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:81",
    "seconds": 0.063,
    "oracle": "Self-written Refused plus cycle substring in What.",
    "oracle_kind": "self",
    "kills": [
      "M06"
    ],
    "unique_kills": [],
    "last_proven_fail": "M06 iteration_test.go:91: got <nil>, want a captured literal-method cycle refusal",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInheritanceRefusesUnsoundOverrides"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.226,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M06.log 2>&1; iteration_test.go:91: got <nil>, want a captured literal-method cycle refusal",
    "vacuous_subcases": [],
    "subsumption_mutants": 1
  },
  {
    "test": "TestLiteralMethodViewsDoNotLoseThis",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:95",
    "seconds": 0.11,
    "oracle": "Self-written nil-error admission for two method views; no receiver IR or executed-output assertion.",
    "oracle_kind": "self",
    "kills": [
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11 iteration_test.go:103: literal-method call must preserve its receiver: /tmp/adamic-gate/TestLiteralMethodViewsDoNotLoseThis776041794/001/main.a:1:125: stage 0 can't lower a literal method through a view that erases its receiver yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDestructuredMethodsCannotLoadOwnSlots"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [],
    "subsumer_seconds": 0.224,
    "vacuous": true,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M11.log 2>&1; iteration_test.go:103: literal-method call must preserve its receiver: /tmp/adamic-gate/TestLiteralMethodViewsDoNotLoseThis776041794/001/main.a:1:125: stage 0 can't lower a literal method through a view that erases its receiver yet",
    "vacuous_subcases": [],
    "subsumption_mutants": 1
  },
  {
    "test": "TestIteratorDescriptorReasons",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:108",
    "seconds": 0.295,
    "oracle": "Self-written NotYet plus seven descriptor-reason substrings.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04",
      "M08"
    ],
    "unique_kills": [],
    "last_proven_fail": "M08 iteration_test.go:125: got /tmp/adamic-gate/TestIteratorDescriptorReasonsreplacing_an_iterator_protocol3424751618/001/main.a:1:119: stage 0 can't lower replacing a represented method at runtime yet, want iterator reason replacing an iterator protocol",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestIteratorMapperIndexHasNumberRepresentation",
      "TestIteratorViewsCannotEraseReceivers",
      "TestEnumSlotViews"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M08 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M08.log 2>&1; iteration_test.go:125: got /tmp/adamic-gate/TestIteratorDescriptorReasonsreplacing_an_iterator_protocol3424751618/001/main.a:1:119: stage 0 can't lower replacing a represented method at runtime yet, want iterator reason replacing an iterator protocol",
    "vacuous_subcases": [],
    "subsumption_mutants": null
  },
  {
    "test": "TestRepresentedMethodReplacementIsNotYet",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:131",
    "seconds": 0.062,
    "oracle": "Self-written NotYet plus represented-method-replacement substring from direct setProperty call.",
    "oracle_kind": "self",
    "kills": [
      "M08"
    ],
    "unique_kills": [],
    "last_proven_fail": "M08 iteration_test.go:156: got <nil>, want a method-replacement refusal",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestIteratorDescriptorReasons"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P02"
    ],
    "subsumer_seconds": 0.295,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M08 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M08.log 2>&1; iteration_test.go:156: got <nil>, want a method-replacement refusal",
    "vacuous_subcases": [],
    "subsumption_mutants": 1
  },
  {
    "test": "TestDestructuredMethodsCannotLoadOwnSlots",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:160",
    "seconds": 0.224,
    "oracle": "Self-written acceptance of either Refused or NotYet for five inputs; reason and fix are unchecked.",
    "oracle_kind": "self",
    "kills": [
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11 iteration_test.go:173: got <nil>, want a refusal before a method becomes an own-slot load",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestLiteralMethodViewsDoNotLoseThis"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.11,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M11.log 2>&1; iteration_test.go:173: got <nil>, want a refusal before a method becomes an own-slot load",
    "vacuous_subcases": [],
    "subsumption_mutants": 1
  },
  {
    "test": "TestGenericIteratorViewsPreserveNativeArguments",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:178",
    "seconds": 0.061,
    "oracle": "Self-written Refused plus nominal-ancestry substring.",
    "oracle_kind": "self",
    "kills": [
      "M07"
    ],
    "unique_kills": [],
    "last_proven_fail": "M07 iteration_test.go:185: got /tmp/adamic-gate/TestGenericIteratorViewsPreserveNativeArguments798416316/001/main.a:2:79: stage 0 can't lower a generic iterable view without proven native type arguments yet, want the earlier nominal generic-view refusal",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestUncheckableCastsStayRefused"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.58,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M07.log 2>&1; iteration_test.go:185: got /tmp/adamic-gate/TestGenericIteratorViewsPreserveNativeArguments798416316/001/main.a:2:79: stage 0 can't lower a generic iterable view without proven native type arguments yet, want the earlier nominal generic-view refusal",
    "vacuous_subcases": [],
    "subsumption_mutants": 1
  },
  {
    "test": "TestIteratorMapperIndexHasNumberRepresentation",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:189",
    "seconds": 0.06,
    "oracle": "Self-written NotYet plus index-representation substring.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M09"
    ],
    "unique_kills": [
      "M09"
    ],
    "last_proven_fail": "M09 iteration_test.go:195: got <nil>, want a mapper index representation refusal",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M09 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M09.log 2>&1; iteration_test.go:195: got <nil>, want a mapper index representation refusal",
    "vacuous_subcases": [],
    "subsumption_mutants": null
  },
  {
    "test": "TestIteratorSymbolKeysAreNotStringKeys",
    "package": "internal/lower",
    "file": "internal/lower/iteration_test.go:199",
    "seconds": 0.064,
    "oracle": "Self-written NotYet plus symbol-key-storage substring.",
    "oracle_kind": "self",
    "kills": [
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10 iteration_test.go:204: got <nil>, want explicit symbol-key storage refusal",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestClassWrongOutputKeysRepair"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.084,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M10.log 2>&1; iteration_test.go:204: got <nil>, want explicit symbol-key storage refusal",
    "vacuous_subcases": [],
    "subsumption_mutants": 1
  }
]
```

| ID | origin/main file:line | Change | All failed package rows |
|---|---|---|---|
| M01 | internal/lower/interface_cast.go:69 | `field.Flags&ast.SymbolFlagsOptional != 0 -> field.Flags&ast.SymbolFlagsOptional == 0` | TestDefaultTaggedInterfaceAdmission, TestDefaultTaggedInterfaceNeedsNoFlag, TestPredicateOverloadRuntime, TestPredicateBodyProof, TestViewObjectContractsAreAvailableToEraser |
| M02 | internal/lower/iteration.go:54 | `Receiver: true -> Receiver: false` |  |
| M03 | internal/lower/iteration.go:192 | `done == nil -> done != nil` | TestClassWrongOutputIteratorReceiver, TestClassWrongOutputKeysRepair, TestIteratorViewsCannotHideReturn, TestIteratorViewsCannotEraseReceivers, TestIteratorDestructuringDoesNotLieAboutExhaustion, TestIteratorDescriptorReasons, TestIteratorMapperIndexHasNumberRepresentation |
| M04 | internal/lower/iteration.go:308 | `for _, module := range modules { 		module.AsNode().ForEachChild(visit) 	} -> (drop whole module-traversal loop)` | TestIteratorGapsAreExplicit, TestIteratorViewsCannotHideReturn, TestIteratorViewsCannotEraseReceivers, TestIteratorDescriptorReasons |
| M05 | internal/lower/refusals.go:150 | `if generator { -> if generator && false {` | TestGeneratorsAreRefusedEvenWithoutYield |
| M06 | internal/lower/cycles.go:63 | `func (l *lowering) findCycles(modules []*ast.SourceFile) error { -> func (l *lowering) findCycles(modules []*ast.SourceFile) error {  if true { return nil }` | TestClassFeaturesStaticSoundness, TestClassFeaturesAccessorCaptureCycle, TestClassFeaturesStaticParentCycle, TestClassFeaturesStaticInterfaceCycle, TestInheritanceRefusesUnsoundOverrides, TestInheritanceGenericSoundness, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestLiteralMethodCapturesCannotMakeCycles, TestWhatZeroOneRefusesIsRefusedWithAFix, TestNamespaceAmbientHostInitialization, TestNestedFunctionCycleIsRefused, TestNestedEnvironmentCycleIncludesDisjointSlots, TestNestedCallbackCycleIsRefused |
| M07 | internal/lower/class_inheritance.go:690 | `func (l *lowering) nominalMismatch(from, to *checker.Type, seen map[[2]*checker.Type]bool) *checker.Type { -> func (l *lowering) nominalMismatch(from, to *checker.Type, seen map[[2]*checker.Type]bool) *checker.Type {  if true { return nil }` | TestUncheckableCastsStayRefused, TestCheckedCastProofAndElision, TestInheritanceRefusesFalseNominalViews, TestInheritanceGenericViewsKeepNominalArguments, TestInheritanceGenericNominalConstraints, TestEnumSlotViews, TestGenericIteratorViewsPreserveNativeArguments, TestUnprovenPredicateReturnsAreRefused, TestProvenRelationsRefuse |
| M08 | internal/lower/class.go:417 | `member.Flags&ast.SymbolFlagsMethod != 0 -> member.Flags&ast.SymbolFlagsMethod == 0` | TestClassFeaturesReadonlyChecker, TestClassFeaturesPrivateStorage, TestClassFeaturesStaticParentCycle, TestClassFeaturesStaticInterfaceCycle, TestInheritanceRefusesUnsoundOverrides, TestInheritanceGenericMonomorphizations, TestInheritanceRefusesGrowingGenericClasses, TestInheritanceGenericSoundness, TestInheritanceGenericFactoryLayouts, TestClassWrongOutputPrivateRepair, TestConstructorCacheOneArgumentRemainsNotYet, TestDefiniteAssignmentUsesReadiness, TestDefiniteAssignmentSoundNeighbors, TestFlagEnumsDomain, TestEnumSlotViews, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestDefaultTaggedInterfaceAdmission, TestIteratorDescriptorReasons, TestRepresentedMethodReplacementIsNotYet, TestWhatZeroOneRefusesIsRefusedWithAFix, TestParameterPropertyCheckerContracts, TestReadinessElisionRequiresDominatingAssignment, TestTasteRepresentationLimitsStayExplicit |
| M09 | internal/lower/iteration_consume.go:42 | `parameter != ir.Number -> parameter == ir.Number` | TestIteratorMapperIndexHasNumberRepresentation |
| M10 | internal/lower/class_features.go:77 | `!isClassInstance(proven) -> isClassInstance(proven)` | TestClassWrongOutputKeysRepair, TestIteratorSymbolKeysAreNotStringKeys |
| M11 | internal/lower/iteration_origin.go:80 | `if isCallee(where) { -> if !isCallee(where) {` | TestLiteralMethodViewsDoNotLoseThis, TestDestructuredMethodsCannotLoadOwnSlots |
| M12 | internal/lower/iteration_consume.go:167 | `!l.includesUndefined(valueType) -> l.includesUndefined(valueType)` | TestIteratorDestructuringDoesNotLieAboutExhaustion |
| P01 | internal/lower/lower.go:20 | `func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) { -> func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {  if true { return nil, nil }` | TestDefaultTaggedInterfaceAdmission, TestIteratorGapsAreExplicit, TestIteratorViewsCannotHideReturn, TestIteratorViewsCannotEraseReceivers, TestIteratorDestructuringDoesNotLieAboutExhaustion, TestGeneratorsAreRefusedEvenWithoutYield, TestLiteralMethodCapturesCannotMakeCycles, TestIteratorDescriptorReasons, TestDestructuredMethodsCannotLoadOwnSlots, TestGenericIteratorViewsPreserveNativeArguments, TestIteratorMapperIndexHasNumberRepresentation, TestIteratorSymbolKeysAreNotStringKeys |
| P02 | internal/lower/class.go:410 | `func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) { -> func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {  if true { return nil, nil }` | TestRepresentedMethodReplacementIsNotYet |

Survivor: M02 is an equivalent candidate. The saved witness produces `name=object_method receiver=true parameters=1` both before and after. functions.go:71 sets Receiver=true during signature lowering, overwriting the altered initialization at iteration.go:54. No differing final output was demonstrated, so this is not called unguarded behavior. Commands: ADAMIC_MUTANT= ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/witness-clean go run ./review/test-audit/internal-lower-interface_cast/witness > M02.witness-before.log 2>&1, then the same command with ADAMIC_MUTANT=M02 and its mutant cache, saving M02.witness-after.log. The witness source is preserved. No other production mutant survived.

Probe findings: P01 returns nil,nil at Lower entry. TestDefaultTaggedInterfaceNeedsNoFlag and TestLiteralMethodViewsDoNotLoseThis pass, so vacuous=true. TestDefaultTaggedInterfaceAdmission fails on optional-target-read, reflection-alias and forged-payload negatives, but its 12 admitted subcases pass; those are vacuous_subcases, not a vacuous whole row. P02 returns nil,nil at setProperty entry, and its direct row fails. P01 passing the direct row is ignored because that row does not enter Lower. Only a row's own entry probe decides vacuity. Probe failures are never production kills or uniqueness evidence.

Subsumption is a finite-menu hint, not deletion advice. Counts of supporting mutants are in subsumption_mutants, usually one or two. Both literal-view/destructured-method rows subsume each other on M11; both iterator view rows subsume each other on M03 and M04. Out-of-unit common subsumers were measured alone three times after restoring production:
- TestInheritanceRefusesUnsoundOverrides: median 0.226 s, trials [0.257, 0.226, 0.219].
- TestUncheckableCastsStayRefused: median 0.580 s, trials [0.572, 0.58, 0.601].
- TestClassWrongOutputKeysRepair: median 0.084 s, trials [0.084, 0.09, 0.069].

Brief ambiguities, costs and limits:

- Warm env.sh did not mean warm Go build dependencies for this starting commit. Discovery's cold dependency build went over 90 s and was stopped; a 90 s retry also timed out while building. Completed dependency artifacts were reused. Subsequent scope discovery succeeded and the first executable whole-package baseline was green. These were build-budget failures, not red tests. Their exact combined build wall time was not instrumented.
- An initial dependency/log command used stage3/api as cwd with repository-relative paths. Redirections failed before npm or tests ran. Corrected to the repository root, then npm ci succeeded. This was an execution mistake, not a red baseline.
- The brief's file-reference commit differs from current origin/main. Name-based discovery resolved the scope without relying on the reference file contents.
- The bounded-slice instruction assumes a slice-only matrix, while the mandatory matrix command asks for the whole package. Since the clean binary completed in 53.394 s, this audit ran all production mutants over the whole package and reports actual package uniqueness. bounded=false reflects the observed matrix, not an extrapolation from 15 rows.
- About three mutants per 15 rows would require 45 mutants, exceeding the 20 cap. Whole-package runs cost around 40 to 60 s, so the fixed code-derived menu contains 12 mutants spread across 11 reached functions and eight production areas. This does not exhaust the behavior of 428 reached functions. No new mutant was added after inspecting kills.
- Compile once refers to the switched source used by all selectors. Standalone vet validation necessarily checks 14 separate variants. Go's compiler cache keeps that validation inexpensive. No native rebuilding or ADAMIC_NATIVE_SPLIT behavior was involved.
- Entry probes use constant-true return blocks in standalone diffs so vet does not reject ordinary unreachable code following a bare return.
- TestIteratorGapsAreExplicit accepts any NotYet. It survives M03, which changes descriptor refusal causes, but TestIteratorDescriptorReasons rejects the wrong reasons. TestDestructuredMethodsCannotLoadOwnSlots accepts either Refused or NotYet without checking a reason. Admission rows do not inspect IR or run output. NeedsNoFlag also does not set or assert any environment flag. These are observed oracle limits, not inferred redundancy.
- Two out-of-scope inventories skipped: TestOriginalCycleLedger requires ADAMIC_CYCLE_LEDGER_ROOT pointing to pristine TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8 with generated diagnostics; TestOptionalWideningCensus requires OPTIONAL_WIDENING_CONFIG and OPTIONAL_WIDENING_OUTPUT for a project. No requested row skipped. Those inventory datasets and caller-selected project targets were not prepared, so uniqueness claims concern the default executable package gate, not those opt-in inventories.
- All oracles are self. No outside-authority value was checked and no runtime agreement was demonstrated by these rows. Three sacred verdicts prove finite-menu package uniqueness, not external correctness.
- Other packages and repository-wide uniqueness were not run. Central replay has every standalone diff. Most reached functions and many branches have no mutant in this menu. There were no panic/timeout matrix runs requiring narrower reruns.

Timing:
Toolchain setup skipped, 0 setup work; env.sh works, Go 1.27.1 linux/amd64, nproc=5. npm ci: 0.835 wall s. Clean whole-package binary: 53.394 s. Coverage slice: 1.184 binary s, 15.979 command wall s including instrumentation build. Forty-five clean timing commands: 113.515 wall s; per-row medians use the test binary's ok line, not Go command wall. Standalone vet validation: 14.460 wall s. Switched binary build: 10.348 wall s. Production matrix and probes: 596.135 command wall s. Additional subsumer runs are saved separately with all nine binary times; aggregate command wall was not instrumented. No native rebuilds. Total unit work roughly 25 minutes including cold Go dependency discovery and evidence preparation. Production restored; no main push and no pull request.
