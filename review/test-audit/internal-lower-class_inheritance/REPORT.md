u029: base 8171b3173bdbfce1f7982d3c4f731279307ece37; nproc 5.
All 18 requested functions exist; acceptance grouping gives 16 scoped rows.
Clean whole-package baseline passed in 26.774 seconds; row medians 0.042 to 0.301 seconds.
Verdicts: 3 sacred, 8 subsumed, 5 overlapping; the acceptance family is vacuous.
19 verdict mutants, 1 supplemental mutant, 3 independent entry probes; production restored.

```json
[
  {
    "test": "TestInheritanceRefusesUnsoundOverrides",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.145,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04",
      "M05",
      "M09",
      "M15",
      "M16",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: class_inheritance_test.go:52: want refusal naming cycle-capable with fix \"Weak\", got /tmp/adamic-gate/TestInheritanceRefusesUnsoundOverridesinherited_private_cycle725595212/001/main.a:2:13: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestInheritance family",
      "TestInheritanceGenericSoundness"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M16.log 2>&1; class_inheritance_test.go:52: want refusal naming cycle-capable with fix \"Weak\", got <nil>",
    "row_id": "R1"
  },
  {
    "test": "TestInheritanceKeepsCheckerConstructorRules",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.085,
    "oracle": "Handwritten *load.CheckError and abstract/super substring checks; upstream checker is not independently run or compared",
    "oracle_kind": "self",
    "kills": [
      "M20"
    ],
    "unique_kills": [],
    "last_proven_fail": "M20: class_inheritance_test.go:82: want abstract, got",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestClassFeaturesPrivateChecker"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOAD"
    ],
    "subsumer_seconds": 0.104,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M20.log 2>&1; class_inheritance_test.go:82: want abstract, got",
    "row_id": "R2"
  },
  {
    "test": "TestInheritance family",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.301,
    "oracle": "Handwritten success-only err == nil checks; no IR or execution result checked",
    "oracle_kind": "self",
    "kills": [
      "M04",
      "M05",
      "M08",
      "M09",
      "M15",
      "M17"
    ],
    "unique_kills": [
      "M08"
    ],
    "last_proven_fail": "M17: class_inheritance_test.go:97: /tmp/adamic-gate/TestInheritanceAllowsSoundOverrides1518375611/001/main.a:4:9: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M08 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M08.log 2>&1; class_inheritance_test.go:97: /tmp/adamic-gate/TestInheritanceAllowsSoundOverrides302435907/001/main.a:4:87: Adamic 0.1 refuses an override that widens its return type; return the base method's result type or a subtype",
    "members": [
      "TestInheritanceAllowsSoundOverrides",
      "TestInheritanceKeepsNominalTupleDestructuring",
      "TestInheritanceNativeSignatureNeighbors"
    ],
    "row_id": "R3"
  },
  {
    "test": "TestInheritanceHasClassIdentity",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.042,
    "oracle": "Handwritten IR class identity or field-layout assertions",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M15",
      "M17"
    ],
    "unique_kills": [
      "M01"
    ],
    "last_proven_fail": "M17: class_inheritance_test.go:106: /tmp/adamic-gate/TestInheritanceHasClassIdentity1510642595/001/main.a:1:20: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M01 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M01 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M01.log 2>&1; class_inheritance_test.go:109: missing ancestry: [{Definition:1 Name:A Base:0 Constructor:0 Fields:[] OwnStart:0 Methods:[] Accessors:[] Literal:false PublicFields:[] Static:false StaticParent:0 StaticFlags:[]} {Definition:2 Name:B Base:0 Constructor:2 Fields:[] OwnStart:0 Methods:[] Accessors:[] Literal:false PublicFields:[] Static:false StaticParent:0 StaticFlags:[]}]",
    "row_id": "R4"
  },
  {
    "test": "TestInheritanceCycleFinderIncludesInheritedFields",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.054,
    "oracle": "Handwritten inherited property-symbol membership assertion",
    "oracle_kind": "self",
    "kills": [
      "M16"
    ],
    "unique_kills": [],
    "last_proven_fail": "M16: class_inheritance_test.go:135: cycle traversal lost Base.parent when visiting Child",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInheritanceGenericSoundness"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_FIELDS"
    ],
    "subsumer_seconds": 0.078,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M16.log 2>&1; class_inheritance_test.go:135: cycle traversal lost Base.parent when visiting Child",
    "row_id": "R5"
  },
  {
    "test": "TestInheritanceRejectsUnsupportedConstructorShapes",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.096,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions",
    "oracle_kind": "self",
    "kills": [
      "M17",
      "M18"
    ],
    "unique_kills": [],
    "last_proven_fail": "M18: class_inheritance_test.go:151: want explicit unsupported construct \"declare or abstract\", got <nil>",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestInheritanceGenericFactoryLayouts",
      "TestConstructorCacheOneArgumentRemainsNotYet"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M18 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M18.log 2>&1; class_inheritance_test.go:151: want explicit unsupported construct \"declare or abstract\", got <nil>",
    "row_id": "R6"
  },
  {
    "test": "TestInheritanceRefusesFalseNominalViews",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.169,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05",
      "M09",
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: class_inheritance_test.go:173: want nominal class refusal, got <nil>",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestInheritanceGenericViewsKeepNominalArguments",
      "TestInheritanceRefusesUnsoundOverrides"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M14.log 2>&1; class_inheritance_test.go:173: want nominal class refusal, got <nil>",
    "row_id": "R7"
  },
  {
    "test": "TestInheritanceRefusesThisBeforeSuperReturns",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.049,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: class_inheritance_test.go:187: want pre-super this refusal with fix, got /tmp/adamic-gate/TestInheritanceRefusesThisBeforeSuperReturns3675340471/001/main.a:1:20: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInheritanceConditionalThisRules"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": 0.098,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M10.log 2>&1; class_inheritance_test.go:187: want pre-super this refusal with fix, got <nil>",
    "row_id": "R8"
  },
  {
    "test": "TestInheritanceGenericMonomorphizations",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.049,
    "oracle": "Handwritten IR class identity or field-layout assertions",
    "oracle_kind": "self",
    "kills": [
      "M05",
      "M12",
      "M13",
      "M17"
    ],
    "unique_kills": [
      "M12"
    ],
    "last_proven_fail": "M17: class_inheritance_test.go:211: /tmp/adamic-gate/TestInheritanceGenericMonomorphizations2523021764/001/main.a:3:15: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M12 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M12.log 2>&1; class_inheritance_test.go:223: monomorphizations lost their shared erased identity",
    "row_id": "R9"
  },
  {
    "test": "TestInheritanceRefusesGrowingGenericClasses",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.044,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions; only caught M17, an earlier computed-base rejection, not a mutation of the recursion bound",
    "oracle_kind": "self",
    "kills": [
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: class_inheritance_test.go:241: want bounded monomorphization refusal with fix, got /tmp/adamic-gate/TestInheritanceRefusesGrowingGenericClasses497919312/001/main.a:2:15: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInheritanceHasClassIdentity"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": 0.042,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M17 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M17.log 2>&1; class_inheritance_test.go:241: want bounded monomorphization refusal with fix, got /tmp/adamic-gate/TestInheritanceRefusesGrowingGenericClasses497919312/001/main.a:2:15: stage 0 can't lower a computed class base; name the base class directly yet",
    "row_id": "R10"
  },
  {
    "test": "TestInheritanceGenericSoundness",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.078,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05",
      "M13",
      "M15",
      "M16",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: class_inheritance_test.go:260: want cycle-capable refusal, got /tmp/adamic-gate/TestInheritanceGenericSoundnessgeneric_inherited_cycle3570834409/001/main.a:3:16: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestInheritanceGenericFactoryLayouts",
      "TestInheritanceRefusesUnsoundOverrides"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M16.log 2>&1; class_inheritance_test.go:260: want cycle-capable refusal, got <nil>",
    "row_id": "R11"
  },
  {
    "test": "TestInheritanceConditionalThisRules",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.098,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: class_inheritance_test.go:276: want pre-super binding refusal with fix, got /tmp/adamic-gate/TestInheritanceConditionalThisRules1489028306/001/main.a:1:20: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInheritanceRefusesThisBeforeSuperReturns"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": 0.049,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M10.log 2>&1; class_inheritance_test.go:276: want pre-super binding refusal with fix, got <nil>",
    "row_id": "R12"
  },
  {
    "test": "TestInheritanceGenericViewsKeepNominalArguments",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.043,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions",
    "oracle_kind": "self",
    "kills": [
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: class_inheritance_test.go:290: want nominal generic argument refusal, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInheritanceGenericNominalConstraints"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": 0.074,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M14.log 2>&1; class_inheritance_test.go:290: want nominal generic argument refusal, got <nil>",
    "row_id": "R13"
  },
  {
    "test": "TestInheritanceGenericNominalConstraints",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.074,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions",
    "oracle_kind": "self",
    "kills": [
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: class_inheritance_test.go:303: want nominal constraint refusal with structural repair, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInheritanceGenericViewsKeepNominalArguments"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": 0.043,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M14.log 2>&1; class_inheritance_test.go:303: want nominal constraint refusal with structural repair, got <nil>",
    "row_id": "R14"
  },
  {
    "test": "TestInheritanceGenericFactoryLayouts",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.042,
    "oracle": "Handwritten IR class identity or field-layout assertions",
    "oracle_kind": "self",
    "kills": [
      "M13",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: class_inheritance_test.go:318: /tmp/adamic-gate/TestInheritanceGenericFactoryLayouts3368409751/001/main.a:3:65: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInheritanceGenericMonomorphizations"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": 0.049,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M13 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M13.log 2>&1; class_inheritance_test.go:318: /tmp/adamic-gate/TestInheritanceGenericFactoryLayouts503510385/001/main.a:3:65: stage 0 can't lower a class instantiated with T[\"native\"] yet",
    "row_id": "R15"
  },
  {
    "test": "TestInheritanceNativeSignatureLimits",
    "package": "internal/lower",
    "file": "internal/lower/class_inheritance_test.go",
    "seconds": 0.162,
    "oracle": "Handwritten acceptance/refusal and diagnostic substring assertions",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05",
      "M06",
      "M09",
      "M13",
      "M15",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: class_inheritance_test.go:383: want NotYet naming parameter \"factor\" and its repair, got /tmp/adamic-gate/TestInheritanceNativeSignatureLimitsgeneric_default_added1415935946/001/main.a:2:16: stage 0 can't lower a computed class base; name the base class directly yet",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestInheritanceGenericSoundness",
      "TestOverride family"
    ],
    "mutants_in_matrix": 19,
    "probe_kills": [
      "P_LOWER"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "ADAMIC_MUTANT=M15 ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M15.log 2>&1; class_inheritance_test.go:383: want NotYet naming parameter \"factor\" and its repair, got /tmp/adamic-gate/TestInheritanceNativeSignatureLimitsgeneric_default_added561452470/001/main.a:3:29: Adamic 0.1 refuses a value without nominal ancestry seen as Base<number>; construct that class or a subclass; use an interface for structural values (adamic/nominal-class)",
    "row_id": "R16"
  }
]
```

Mutants below use line numbers at the base commit. Every diff is independent and has no selector. Full run command for ID: `ADAMIC_MUTANT=ID ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/ID timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > ID.log 2>&1`.

| ID | Base file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/class.go:146 | metadata.Base = base.class -> metadata.Base = 0 | TestInheritanceHasClassIdentity |
| M02 | internal/lower/class.go:147 | metadata.Fields = append(metadata.Fields, l.result.Classes[base.class-1].Fields...) -> (drop whole statement/block) |  |
| M03 | internal/lower/class_inheritance.go:42 | if len(bases) == 0 { -> if len(bases) >= 0 { | TestClassFeaturesAccessorRefusals, TestInheritanceGenericSoundness, TestInheritanceNativeSignatureLimits, TestInheritanceRefusesFalseNominalViews, TestInheritanceRefusesUnsoundOverrides, TestOverride family, TestParameterPropertiesSoundness |
| M04 | internal/lower/class_inheritance.go:108 | !l.checker.IsReadonlySymbol(inherited) -> l.checker.IsReadonlySymbol(inherited) | TestClassFeaturesStaticSoundness, TestInheritance family, TestInheritanceRefusesUnsoundOverrides, TestParameterPropertiesSoundness |
| M05 | internal/lower/class_inheritance.go:132 | if len(old.Parameters()) != len(next.Parameters()) { -> if len(old.Parameters()) == len(next.Parameters()) { | TestClassWrongOutputIteratorReceiver, TestEnumSlotViews, TestInheritance family, TestInheritanceGenericMonomorphizations, TestInheritanceGenericSoundness, TestInheritanceNativeSignatureLimits, TestInheritanceRefusesFalseNominalViews, TestInheritanceRefusesUnsoundOverrides, TestOverride family |
| M06 | internal/lower/class_inheritance.go:178 | representation = ir.Maybe(representation) -> (drop whole statement/block) | TestInheritanceNativeSignatureLimits, TestOverride family |
| M07 | internal/lower/class_inheritance.go:186 | return 0, true -> return ir.Boolean, true |  |
| M08 | internal/lower/class_inheritance.go:154 | l.classAssignable(newResult, oldResult) -> l.classAssignable(oldResult, newResult) | TestInheritance family |
| M09 | internal/lower/class_inheritance.go:140 | if checkABI && (!knownA \|\| !knownB \|\| a != b) { 				return -> if checkABI && (!knownA \|\| !knownB \|\| a == b) { 				return | TestEnumSlotViews, TestInheritance family, TestInheritanceNativeSignatureLimits, TestInheritanceRefusesFalseNominalViews, TestInheritanceRefusesUnsoundOverrides, TestOverride family |
| M10 | internal/lower/class.go:754 | if l.instance != nil && l.instance.unreadyThis[node] { -> if false && l.instance != nil && l.instance.unreadyThis[node] { | TestInheritanceConditionalThisRules, TestInheritanceRefusesThisBeforeSuperReturns |
| M11 supplemental | internal/lower/class_inheritance.go:822 | if !tuple { -> if tuple { |  |
| M12 | internal/lower/class_inheritance.go:868 | for key, instance := range l.instances { 		if l.classKeyMatches(key, declaration) { 			return l.result.Classes[instance.class-1].Definition 		} 	} -> (drop whole statement/block) | TestInheritanceGenericMonomorphizations |
| M13 | internal/lower/class_inheritance.go:933 | concrete[index] = instantiateType(l.checker, base, mapper) -> concrete[index] = base | TestCheckedCastProofAndElision, TestInheritanceGenericFactoryLayouts, TestInheritanceGenericMonomorphizations, TestInheritanceGenericSoundness, TestInheritanceNativeSignatureLimits |
| M14 | internal/lower/class_inheritance.go:744 | if !l.nominalAncestor(from, to, seen) { -> if false && !l.nominalAncestor(from, to, seen) { | TestCheckedCastProofAndElision, TestEnumSlotViews, TestGenericIteratorViewsPreserveNativeArguments, TestInheritanceGenericNominalConstraints, TestInheritanceGenericViewsKeepNominalArguments, TestInheritanceRefusesFalseNominalViews, TestProvenRelationsRefuse, TestUncheckableCastsStayRefused, TestUnprovenPredicateReturnsAreRefused |
| M15 | internal/lower/class_inheritance.go:593 | if !l.sameClassArguments(from, to) { -> if l.sameClassArguments(from, to) { | TestCheckedCastProofAndElision, TestClassFeaturesPrivateStorage, TestClassWrongOutputIteratorReceiver, TestEnumSlotViews, TestGenericIteratorViewsPreserveNativeArguments, TestInheritance family, TestInheritanceGenericSoundness, TestInheritanceHasClassIdentity, TestInheritanceNativeSignatureLimits, TestInheritanceRefusesUnsoundOverrides, TestUncheckableCastsStayRefused |
| M16 | internal/lower/cycles.go:235 | property.Flags&ast.SymbolFlagsMethod == 0 -> property.Flags&ast.SymbolFlagsMethod != 0 | TestClassFeaturesStaticInterfaceCycle, TestClassFeaturesStaticParentCycle, TestClassFeaturesStaticSoundness, TestInheritanceCycleFinderIncludesInheritedFields, TestInheritanceGenericSoundness, TestInheritanceRefusesUnsoundOverrides, TestReadonlyFieldsAreJudgedByTheirConstructorsWrites, TestWhatZeroOneRefusesIsRefusedWithAFix |
| M17 | internal/lower/class_inheritance.go:20 | if len(types) != 1 \|\| !ast.IsIdentifier(ast.SkipParentheses(types[0].AsExpressionWithTypeArguments().Expression)) { -> if len(types) != 1 \|\| ast.IsIdentifier(ast.SkipParentheses(types[0].AsExpressionWithTypeArguments().Expression)) { | TestCheckedCastProofAndElision, TestClassFeaturesPrivateStorage, TestClassWrongOutputIteratorReceiver, TestInheritance family, TestInheritanceConditionalThisRules, TestInheritanceGenericFactoryLayouts, TestInheritanceGenericMonomorphizations, TestInheritanceGenericSoundness, TestInheritanceHasClassIdentity, TestInheritanceNativeSignatureLimits, TestInheritanceRefusesGrowingGenericClasses, TestInheritanceRefusesThisBeforeSuperReturns, TestInheritanceRefusesUnsoundOverrides, TestInheritanceRejectsUnsupportedConstructorShapes, TestParameterPropertiesSoundness |
| M18 | internal/lower/class_inheritance.go:237 | ast.ModifierFlagsAmbient\|ast.ModifierFlagsAbstract -> ast.ModifierFlagsStatic | TestConstructorCacheOneArgumentRemainsNotYet, TestInheritanceRejectsUnsupportedConstructorShapes, TestIteratorDescriptorReasons, TestPrototypeHazardsBehindObjectViewsAreNotYet |
| M19 | internal/lower/class_inheritance.go:472 | initialized, err := l.fieldInitializers(declaration, l.this) 	if err != nil { 		return nil, err 	} 	statements = append(statements, initialized...) -> (drop whole statement/block) | TestParameterPropertiesSoundness |
| M20 | internal/load/load.go:148 | &CheckError{Diagnostics: diagnostics} -> &CheckError{Diagnostics: nil} | TestClassFeaturesPrivateChecker, TestClassFeaturesReadonlyChecker, TestInheritanceKeepsCheckerConstructorRules, TestInputSpreadCoverage, TestOptionalWideningRefused |

Survivors:

- M02: unguarded metadata behavior. Independent W1: B.Fields changes from [x, y] to [y].
- M07: unguarded diagnostic classification. Independent W2: void-to-boolean override changes from NotYet native-result representation to Refused widened return type. Both reject the source.
- M11: supplemental equivalent candidate. Scoped coverage of classDestructuredView was 0%; original and derived-tuple observations did not differ. No verdict rests on it.

Code under test: reached Go lowering functions listed in reached-functions.txt, plus the loader diagnostic-propagation adapter for the direct Load row. The latter mutates Adamic's returned CheckError payload, never the upstream checker. Direct entries: Lower, cycleFinder.fields, Load. Oracle: each row's handwritten assertions, classified self; no independent implementation is run by these scoped tests.

Family decisions: AllowsSoundOverrides, KeepsNominalTupleDestructuring and NativeSignatureNeighbors form TestInheritance family because all perform the same success-only check through lowerSource. Other scoped members assert distinct IR or diagnostic properties. Outside the scope, three wrappers calling assertOverrideParameterRefusal are grouped as TestOverride family: TestAbfe962OverrideDefaultAddedRefusesNativeSignature, TestOverrideOptionalNumberRefusesNativeSignature, TestOverrideOptionalRefusesNativeSignature. Raw function results remain in matrix-raw.json.

Brief ambiguities, limitations and costs:

- The cited historical commit 8de93800f4 differs from fresh origin/main 8171b317. Fresh origin/main governs; no requested name moved or vanished.
- Three mutants per 18 functions would mean 54, but the cap is 20. The cap governs. Coverage later disqualified M11 from verdicts, leaving 19.
- The family rule is broader than generated shards. Three inline acceptance checks were grouped despite different descriptive suffixes. All member measurements are retained, and the family was also timed three times as a single run.
- Initial M12 left an unused loop binding. It was corrected to drop the entire identity-reuse loop before any matrix ran. Two scratch selector-generation errors were corrected before matrices ran. These were audit implementation costs, not defects in the brief.
- The full package fits 90 seconds but each matrix costs roughly 27 to 36 wall seconds. This dominated the audit, which exceeded the approximate 20-minute budget. No production matrix cooked or panicked.
- Empty Lower and Load probes abort the whole package on nil dereferences. All 18 scoped functions were rerun individually for those probes. Probe kills are separate and never support verdicts. Outside the scoped solo reruns, post-panic probe results remain unknown.
- Skipped: TestOriginalCycleLedger needs pristine TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8 with generated diagnostics. It was not locally available; a bounded 30-second object-availability attempt did not obtain it. TestOptionalWideningCensus requires a specified project/config and output destination; none was supplied. TestMixedUnionContractGraph/interface_Node_{readonly_ready:boolean}_type_Target=Node|readonly_Node[] skips pending compiler/views-v3. No scoped row skipped. Package uniqueness is over the executed default package rows; optional inventory results remain unknown.
- M17 causes many rows to fail before their principal check. In particular, the growing-generic row caught only M17, an early computed-base error; the recursion bound itself was not mutated. Its subsumption claim rests on one mutant. All subsumption claims are hints over this 19-mutant set, never deletion recommendations.
- These NativeSignature tests examine IR acceptance/refusal, not native execution. No native product was built. Each run nevertheless received its own ADAMIC_BUILD_CACHE_DIR.
- No other package test suite ran. go vet checked only the mutated package for each standalone diff. Repo-wide uniqueness is deferred to central replay.

Timing: warm toolchain source succeeded, setup skipped (0 seconds); nproc 5. npm ci in stage3/api reported 0.355 seconds. Baseline binary: 26.774 seconds. Switched Go build: 11.723 seconds. Whole matrix command walls: 650.960 seconds, excluding scoped panic rerun command walls. All matrix and scoped probe binary elapsed totals: 607.142 seconds. Final independent apply/vet validation: 3.563 seconds. Individual row and family binary measurements appear in timings.json and family.json. Initial validation/build retries and command startup add overhead; total session was roughly 25 minutes.
