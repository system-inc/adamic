u032: 14 discovered rows; 9 sacred and 5 subsumed within the bounded matrix.
Base: 8171b3173bdbfce1f7982d3c4f731279307ece37; clean package passed in 55.418s.
20 fixed-menu mutants, three entry probes; no production survivor or skipped unit row.
Ledger proof decisions pass an empty proof; module-order obligations remain checked.
Evidence pushed on test-audit/internal-lower-enums under review/test-audit/internal-lower-enums/.

```json
[
  {
    "test": "TestEnumSlotViews",
    "package": "internal/lower",
    "file": "internal/lower/enums_test.go:15",
    "seconds": 0.603,
    "oracle": "Self-written accepted/refused slot views and broad reason substrings; positive cases only require err=nil.",
    "oracle_kind": "self",
    "kills": [
      "M02",
      "M04",
      "M20"
    ],
    "unique_kills": [
      "M02"
    ],
    "last_proven_fail": "M02 enums_test.go:55: got <nil>, want Refused",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M02 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M02.log 2>&1; enums_test.go:55: got <nil>, want Refused",
    "vacuous_subcases": [
      "assignment",
      "literal_property",
      "variable",
      "array_literal",
      "array_alias",
      "readonly_into_enum_array",
      "compound",
      "function_view",
      "increment",
      "mutable_field_alias",
      "member_initializer",
      "argument",
      "class_field",
      "map_alias",
      "optional_union",
      "override",
      "initializer"
    ]
  },
  {
    "test": "TestEnumSwitchExhaustiveness",
    "package": "internal/lower",
    "file": "internal/lower/enums_test.go:67",
    "seconds": 0.192,
    "oracle": "Self-written Refused type and non-exhaustive enum switch substring; accepted neighbors only require err=nil.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04",
      "M07"
    ],
    "unique_kills": [
      "M03"
    ],
    "last_proven_fail": "M03 enums_test.go:75: got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M03 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M03 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M03.log 2>&1; enums_test.go:75: got <nil>"
  },
  {
    "test": "TestEnumLimitsStayLoud",
    "package": "internal/lower",
    "file": "internal/lower/enums_test.go:90",
    "seconds": 0.541,
    "oracle": "Self-written NotYet or Refused type, without checking the specific reason for most inputs.",
    "oracle_kind": "self",
    "kills": [
      "M04",
      "M05",
      "M06"
    ],
    "unique_kills": [
      "M05",
      "M06"
    ],
    "last_proven_fail": "M06 enums_test.go:109: got <nil>, want NotYet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M06.log 2>&1; enums_test.go:109: got <nil>, want NotYet"
  },
  {
    "test": "TestConstEnumErasesRuntimeObject",
    "package": "internal/lower",
    "file": "internal/lower/enums_test.go:125",
    "seconds": 0.054,
    "oracle": "Self-written absence of enum binding and ObjectLiteral nodes. POrder yields an empty IR program and this row passes, showing its absence-only oracle weakness; primary PLower instead panics.",
    "oracle_kind": "self",
    "kills": [
      "M04",
      "M07"
    ],
    "unique_kills": [],
    "last_proven_fail": "M07 enums_test.go:133: const enum made a runtime binding",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestEnumSwitchExhaustiveness"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": 0.192,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M07.log 2>&1; enums_test.go:133: const enum made a runtime binding"
  },
  {
    "test": "TestEnumIdentityAcrossModules",
    "package": "internal/lower",
    "file": "internal/lower/enums_test.go:144",
    "seconds": 0.058,
    "oracle": "Self-written Refused type and enum-members diagnostic substring after loader admission.",
    "oracle_kind": "self",
    "kills": [
      "M01"
    ],
    "unique_kills": [
      "M01"
    ],
    "last_proven_fail": "M01 enums_test.go:160: got <nil>, want closed enum identity refusal",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M01 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M01 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M01.log 2>&1; enums_test.go:160: got <nil>, want closed enum identity refusal"
  },
  {
    "test": "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
    "package": "internal/lower",
    "file": "internal/lower/fresh_test.go:12",
    "seconds": 0.134,
    "oracle": "Self-written exact cycle-refusal fragments, location and Weak fix; accepted neighbors require err=nil.",
    "oracle_kind": "self",
    "kills": [
      "M14",
      "M15"
    ],
    "unique_kills": [
      "M14",
      "M15"
    ],
    "last_proven_fail": "M15 fresh_test.go:27: refusal \"/tmp/adamic-gate/TestReadonlyFieldsAreJudgedByTheirConstructorsWrites3325740604/001/main.a:2:2: Adamic 0.1 refuses Node.kids, a readonly field, which its constructor writes, of type Node[], which can reach back to the Node holding it: a cycle reference counting can't free, and the write at /tmp/adamic-gate/TestReadonlyFieldsAreJudgedByTheirConstructorsWrites3325740604/001/main.a:1:7 may close one (); declare it kids: Weak<Node[]> (import type { Weak } from 'adamic'), which doesn't count and reads undefined once what it points to is freed; or write into it only values this function made, or only into what it made (adamic/cycle-capable)\" doesn't say \"main.a:3:2: Adamic 0.1 refuses Node.parent, a readonly field, which its constructor writes, of type Node | undefined\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M15 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M15.log 2>&1; fresh_test.go:27: refusal \"/tmp/adamic-gate/TestReadonlyFieldsAreJudgedByTheirConstructorsWrites3325740604/001/main.a:2:2: Adamic 0.1 refuses Node.kids, a readonly field, which its constructor writes, of type Node[], which can reach back to the Node holding it: a cycle reference counting can't free, and the write at /tmp/adamic-gate/TestReadonlyFieldsAreJudgedByTheirConstructorsWrites3325740604/001/main.a:1:7 may close one (); declare it kids: Weak<Node[]> (import type { Weak } from 'adamic'), which doesn't count and reads undefined once what it points to is freed; or write into it only values this function made, or only into what it made (adamic/cycle-capable)\" doesn't say \"main.a:3:2: Adamic 0.1 refuses Node.parent, a readonly field, which its constructor writes, of type Node | undefined\""
  },
  {
    "test": "TestFunctionValueUnionViewsStayNotYet",
    "package": "internal/lower",
    "file": "internal/lower/function_values_test.go:8",
    "seconds": 0.294,
    "oracle": "Self-written NotYet type only; no diagnostic identity or behavior comparison.",
    "oracle_kind": "self",
    "kills": [
      "M20"
    ],
    "unique_kills": [],
    "last_proven_fail": "M20 function_values_test.go:19: callable view requires a union slot adapter, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestEnumSlotViews"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": 0.603,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M20.log 2>&1; function_values_test.go:19: callable view requires a union slot adapter, got <nil>"
  },
  {
    "test": "TestGenericFunctionPolymorphicRecursionIsRefused",
    "package": "internal/lower",
    "file": "internal/lower/generic_instance_key_test.go:10",
    "seconds": 0.06,
    "oracle": "Self-written Refused type and polymorphic recursion substring.",
    "oracle_kind": "self",
    "kills": [
      "M08",
      "M10"
    ],
    "unique_kills": [
      "M10"
    ],
    "last_proven_fail": "M10 generic_instance_key_test.go:20: want a polymorphic recursion refusal, got /tmp/adamic-gate/TestGenericFunctionPolymorphicRecursionIsRefused1606290132/001/main.a:3:34: Adamic 0.1 refuses a generic function instantiated without end (recursive instantiation); call it with the same type arguments it was called with, or write a function per type",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M10.log 2>&1; generic_instance_key_test.go:20: want a polymorphic recursion refusal, got /tmp/adamic-gate/TestGenericFunctionPolymorphicRecursionIsRefused1606290132/001/main.a:3:34: Adamic 0.1 refuses a generic function instantiated without end (recursive instantiation); call it with the same type arguments it was called with, or write a function per type"
  },
  {
    "test": "TestGenericUnionFixtureHasSeparateInstances",
    "package": "internal/lower",
    "file": "internal/lower/generic_instance_key_test.go:26",
    "seconds": 0.059,
    "oracle": "Self-written count of two functions with describe_ prefix; no signature/type-identity comparison.",
    "oracle_kind": "self",
    "kills": [
      "M08",
      "M09"
    ],
    "unique_kills": [],
    "last_proven_fail": "M09 generic_instance_key_test.go:34: /tmp/adamic-gate/TestGenericUnionFixtureHasSeparateInstances329503787/001/main.a:9:13: Adamic 0.1 refuses a generic function instantiated without end (polymorphic recursion); call it with the same type arguments it was called with, or write a function per type",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestGenericJSONUnionArrayIsNotYet"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": 0.06,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M09 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M09.log 2>&1; generic_instance_key_test.go:34: /tmp/adamic-gate/TestGenericUnionFixtureHasSeparateInstances329503787/001/main.a:9:13: Adamic 0.1 refuses a generic function instantiated without end (polymorphic recursion); call it with the same type arguments it was called with, or write a function per type"
  },
  {
    "test": "TestGenericJSONUnionArrayIsNotYet",
    "package": "internal/lower",
    "file": "internal/lower/generic_instance_key_test.go:47",
    "seconds": 0.06,
    "oracle": "Self-written exact named JSON union array NotYet.What.",
    "oracle_kind": "self",
    "kills": [
      "M08",
      "M09",
      "M11"
    ],
    "unique_kills": [
      "M11"
    ],
    "last_proven_fail": "M11 generic_instance_key_test.go:56: want the named JSON union array NotYet, got /tmp/adamic-gate/TestGenericJSONUnionArrayIsNotYet1366412638/001/main.a:3:72: stage 0 can't lower JSON.stringify array-containing unions (adamic/json-union-array) yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M11.log 2>&1; generic_instance_key_test.go:56: want the named JSON union array NotYet, got /tmp/adamic-gate/TestGenericJSONUnionArrayIsNotYet1366412638/001/main.a:3:72: stage 0 can't lower JSON.stringify array-containing unions (adamic/json-union-array) yet"
  },
  {
    "test": "TestOriginalCycleLedger",
    "package": "internal/lower",
    "file": "internal/lower/import_cycle_ledger_test.go:27",
    "seconds": 2.524,
    "oracle": "git verifies upstream source pin; Go cohere findings must be nonempty. Module order compares to recorded independently verified Node order (stage3/fixtures/cycles/order.json), not a live Node run here; no order value independently rechecked this session. Self-written 914/58 totals and binding identities. Proof decision counts are exported without assertion: PProve passes with 489 checked/425 proven reads instead of 9/905.",
    "oracle_kind": [
      "external-run",
      "external-authority",
      "self"
    ],
    "kills": [
      "M18",
      "M19"
    ],
    "unique_kills": [
      "M19"
    ],
    "last_proven_fail": "M19 import_cycle_ledger_test.go:50: original program lost its import cycle",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "POrder"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M19 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M19.log 2>&1; import_cycle_ledger_test.go:50: original program lost its import cycle",
    "entry_probe_results": {
      "esmModuleOrder": "POrder: fail",
      "proveModuleReads": "PProve: pass"
    },
    "vacuous_subcases": [
      "proveModuleReads proof decisions: empty proof passes PProve; module order still fails POrder"
    ]
  },
  {
    "test": "TestUndecidedCycleReadsUseReadyChecks",
    "package": "internal/lower",
    "file": "internal/lower/import_cycle_ready_test.go:28",
    "seconds": 2.805,
    "oracle": "Original source Node, generated-JavaScript Node and sanitized native each compared to expected output/exit; self-written emitted-C absence assertions also check proof elision.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M16",
      "M17",
      "M18"
    ],
    "unique_kills": [
      "M16",
      "M17"
    ],
    "last_proven_fail": "M17 import_cycle_ready_test.go:96: /workspace/adamic-tools/bin/node: got \"undefined\\n\", want \"adamic: panic: ReferenceError: Cannot access 'value' before initialization\\n\" (error <nil>)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder",
      "PProve"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M17 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M17.log 2>&1; import_cycle_ready_test.go:96: /workspace/adamic-tools/bin/node: got \"undefined\\n\", want \"adamic: panic: ReferenceError: Cannot access 'value' before initialization\\n\" (error <nil>)"
  },
  {
    "test": "TestInputSpreadArgumentsAreNotYet",
    "package": "internal/lower",
    "file": "internal/lower/input_test.go:17",
    "seconds": 0.282,
    "oracle": "Self-written NotYet type and diagnostic naming each prelude door.",
    "oracle_kind": "self",
    "kills": [
      "M12",
      "M13"
    ],
    "unique_kills": [],
    "last_proven_fail": "M13 input_test.go:33: got lower: /tmp/adamic-gate/TestInputSpreadArgumentsAreNotYetprogramArguments2691799448/001/main.a:3:1: programArguments takes nothing, and the checker let arguments through, want NotYet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInputSpreadCoverage"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": 0.959,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M13 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M13.log 2>&1; input_test.go:33: got lower: /tmp/adamic-gate/TestInputSpreadArgumentsAreNotYetprogramArguments2691799448/001/main.a:3:1: programArguments takes nothing, and the checker let arguments through, want NotYet"
  },
  {
    "test": "TestInputSpreadCoverage",
    "package": "internal/lower",
    "file": "internal/lower/input_test.go:43",
    "seconds": 0.959,
    "oracle": "Self-written fixture count 47, Lower NotYet name/location or local NotYet; copied TypeScript array-spread diagnostic text is also matched against the running checker. The seven _array cases never call Lower and pass PLower.",
    "oracle_kind": [
      "external-authority",
      "self"
    ],
    "kills": [
      "M12",
      "M13"
    ],
    "unique_kills": [],
    "last_proven_fail": "M13 input_test.go:77 (multiline assertion): : got lower: /workspace/adamic/internal/lower/testdata/input-spread/programArguments_tuple.a:3:1: programArguments takes nothing, and the checker let arguments through, want NotYet naming programArguments",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestInputSpreadArgumentsAreNotYet"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "PLower",
      "POrder"
    ],
    "subsumer_seconds": 0.282,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestEnumSlotViews",
      "TestEnumSwitchExhaustiveness",
      "TestEnumLimitsStayLoud",
      "TestConstEnumErasesRuntimeObject",
      "TestEnumIdentityAcrossModules",
      "TestReadonlyFieldsAreJudgedByTheirConstructorsWrites",
      "TestFunctionValueUnionViewsStayNotYet",
      "TestGenericFunctionPolymorphicRecursionIsRefused",
      "TestGenericUnionFixtureHasSeparateInstances",
      "TestGenericJSONUnionArrayIsNotYet",
      "TestOriginalCycleLedger",
      "TestUndecidedCycleReadsUseReadyChecks",
      "TestInputSpreadArgumentsAreNotYet",
      "TestInputSpreadCoverage"
    ],
    "evidence": "ADAMIC_MUTANT=M13 ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run \"$pattern\" > M13.log 2>&1; input_test.go:77 (multiline assertion): : got lower: /workspace/adamic/internal/lower/testdata/input-spread/programArguments_tuple.a:3:1: programArguments takes nothing, and the checker let arguments through, want NotYet naming programArguments",
    "vacuous_subcases": [
      "readTextFile_array.a",
      "readDirectory_array.a",
      "programArguments_array.a",
      "fileStatus_array.a",
      "writeTextFile_array.a",
      "utf8Length_array.a",
      "utf8At_array.a"
    ]
  }
]
```

Code under test: Adamic lowering, not load.Load or cohere rule_runner. The coverage profile and reached-functions.txt enumerate every executed lower package function across these rows before mutation. Lower is the primary entry for thirteen rows. TestOriginalCycleLedger directly reaches esmModuleOrder, runtimeModuleStatement and proveModuleReads. Mutants are fixed in menu.json before matrix observations, spread over enums, module scheduling, generic specialization, input-spread diagnostics, cycle checking, fresh-write proofs, and callable views.

Oracle: self-written Refused/NotYet types and diagnostic text, successful lowering and IR structure. TestUndecidedCycleReadsUseReadyChecks actually runs original source Node, backend Node and sanitized native and compares each with expected output/exit, plus self-written emitted-C absence assertions. TestOriginalCycleLedger runs git for an external source pin, calls Go cohere for nonzero findings, and compares esmModuleOrder to recorded independently verified Node order. Its proof decision counts are exported but not asserted. TestInputSpreadCoverage also checks TypeScript checker diagnostic text; this is a preparation assertion, not the lowering code under test. It has array cases that never call Lower.

Matrix is bounded to the exact fourteen supplied rows despite a 55.418s clean whole package run. Full-package mutant repetitions would exceed the unit budget. Kills outside this set and package-wide uniqueness are unknown. All rows are separate: bodies have distinct assertions, not wrappers varying only one checker input. No witnesses, helpers or setup-only rows among the fourteen.

All file:line references use the base commit. The complete pre-mutation reached-function list is reached-functions.txt (273 declarations), backed by coverage.out and functions-coverage.txt. Every requested name exists in list.log at the starting commit and remains in the file named by the brief; none moved or vanished. No families, witnesses, setup-only rows or subprocess helpers were found after reading all seven complete test files.

Matrix pattern is `pattern="^($(paste -sd '|' review/test-audit/internal-lower-enums/rows.txt))$"`. Source env.sh in every shell. Set ADAMIC_CYCLE_LEDGER_ROOT=/tmp/u032/typescript and ADAMIC_CYCLE_LEDGER_OUTPUT=/tmp/u032/ledger-<id>.json. Each matrix command sets ADAMIC_MUTANT=<id> and ADAMIC_BUILD_CACHE_DIR=/tmp/u032/cache/<id>, then runs `timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run "$pattern" > <id>.log 2>&1`. Paths in JSON evidence are shortened relative to the evidence directory. Timings run `timeout 95 go test -count=1 -timeout 90s ./internal/lower/ -run ^<row>$ > <row>-<round>.log 2>&1` three separate times, with no concurrent test workload. No native product shares a cache directory across mutants.

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/enums.go:39 | flip closed enum identity comparison | TestEnumIdentityAcrossModules |
| M02 | internal/lower/enums.go:95 | early return enum object relation check | TestEnumSlotViews |
| M03 | internal/lower/enums.go:436 | early return enum exhaustiveness check | TestEnumSwitchExhaustiveness |
| M04 | internal/lower/enums.go:157 | flip non-finite rejection condition | TestEnumSlotViews, TestEnumSwitchExhaustiveness, TestEnumLimitsStayLoud, TestConstEnumErasesRuntimeObject |
| M05 | internal/lower/enums.go:183 | change forbidden member name constant | TestEnumLimitsStayLoud |
| M06 | internal/lower/enums.go:400 | change ambient modifier option | TestEnumLimitsStayLoud |
| M07 | internal/lower/modules.go:120 | flip const-enum registration condition | TestEnumSwitchExhaustiveness, TestConstEnumErasesRuntimeObject |
| M08 | internal/lower/generic.go:172 | drop representative registration statement | TestGenericFunctionPolymorphicRecursionIsRefused, TestGenericUnionFixtureHasSeparateInstances, TestGenericJSONUnionArrayIsNotYet |
| M09 | internal/lower/generic.go:14 | change generic depth constant | TestGenericUnionFixtureHasSeparateInstances, TestGenericJSONUnionArrayIsNotYet |
| M10 | internal/lower/generic.go:100 | change recursion diagnostic constant | TestGenericFunctionPolymorphicRecursionIsRefused |
| M11 | internal/lower/generic.go:349 | change JSON union diagnostic constant | TestGenericJSONUnionArrayIsNotYet |
| M12 | internal/lower/input.go:36 | change spread diagnostic constant | TestInputSpreadArgumentsAreNotYet, TestInputSpreadCoverage |
| M13 | internal/lower/input.go:30 | off-by-one input function bound | TestInputSpreadArgumentsAreNotYet, TestInputSpreadCoverage |
| M14 | internal/lower/cycles.go:338 | change cycle diagnostic constant | TestReadonlyFieldsAreJudgedByTheirConstructorsWrites |
| M15 | internal/lower/fresh.go:43 | flip fresh-write proof condition | TestReadonlyFieldsAreJudgedByTheirConstructorsWrites |
| M16 | internal/lower/load_time_reads.go:102 | flip module read readiness condition | TestUndecidedCycleReadsUseReadyChecks |
| M17 | internal/lower/load_time_reads.go:88 | flip provider completion order comparison | TestUndecidedCycleReadsUseReadyChecks |
| M18 | internal/lower/modules.go:35 | change back-edge cycle flag constant | TestOriginalCycleLedger, TestUndecidedCycleReadsUseReadyChecks |
| M19 | internal/lower/modules.go:194 | flip export runtime-edge condition | TestOriginalCycleLedger |
| M20 | internal/lower/expression.go:305 | flip callable union ABI guard | TestEnumSlotViews, TestFunctionValueUnionViewsStayNotYet |

No production survivors. Every production mutant uses the fixed menu; no supplemental mutant is needed. Diagnostic changes are production output changes, not oracle edits. M08 drops the entire representative append statement. M02/M03 standalone diffs replace their full check bodies with early returns. Probes replace full bodies and remove unused imports for standalone compilation. Switch scaffolding is stored under switch/*.txt and is absent from every standalone diff.

PLower returns nil, nil at Lower entry; POrder returns nil, false from esmModuleOrder; PProve returns immediately from proveModuleReads. PLower panicked during the whole slice, so every row was rerun alone for it, in PLower-<row>.log. Only those isolated observations establish its results; ledger passes because it does not call Lower. Thirteen Lower rows fail, with nil-pointer panics in const-enum, generic-instance and readiness rows. For the ledger, POrder fails but PProve passes. This multi-entry row is overall non-vacuous because it rejects empty scheduling; its proof-decision portion is explicitly vacuous. PProve changes read counts from checked=9/proven=905 to checked=489/proven=425, and statement counts from checked=3/proven=55 to checked=29/proven=29, without a failure. Both complete ledger artifacts are saved.

Under PLower, all 17 open enum-slot positive subcases pass while six closed/refusal subcases fail; exact names are in rows.json. All seven InputSpreadCoverage _array.a subcases pass because they stop at the loader checker before Lower, while lowering subcases fail. ConstEnumErasesRuntimeObject also passes POrder because its checks assert only absence of a runtime binding/object in an empty IR program; this supplemental entry is not used to set primary-Lower vacuity. No test or oracle was edited.

Subsumption rests only on this fixed bounded menu: ConstEnumErasesRuntimeObject by EnumSwitchExhaustiveness (2 kills); FunctionValueUnionViewsStayNotYet by EnumSlotViews (1 kill); GenericUnionFixtureHasSeparateInstances by GenericJSONUnionArrayIsNotYet (2 kills); the two input-spread rows mutually subsume each other (2 kills each). These are hints, not deletion recommendations. All unique_kills mean unique within these fourteen rows, not the package or repository.

Brief issues and time costs:

- origin/main is newer than the brief's 8de93800f4 file-location pin. The fetched starting commit and actual discovery define scope; all fourteen names and file locations still match.
- Warm env.sh does not imply warm Go dependencies. Cold test discovery ran about 102 seconds before I stopped it; a bounded 90-second retry also cooked, and the next bounded retry succeeded. This was compilation, not a red test baseline. The first discovery was insufficiently bounded, a process-control error documented here. Completed dependency compilation was reused; no reduced row selection could avoid compiling this common package dependency.
- The ledger opt-in was installable, so pristine upstream v6.0.3 was cloned at 050880ce59e30b356b686bd3144efe24f875ebc8 and processDiagnosticMessages.mjs generated diagnostics. All fourteen rows then ran with no skip. No npm install was required in this source-only upstream directory; stage3/api npm ci reported 548ms.
- Whole-package baseline fits 90 seconds but costs 55.418s. Repeating it for twenty mutants would exceed the unit budget, so the matrix is the explicitly requested fourteen-row slice and bounded. Outside-slice kills are unknown. This uses the brief's slice exception rather than claiming package uniqueness.
- The ledger has multiple lowering entries. The singular entry-probe rule is ambiguous here. Both were probed; overall vacuous=false records failure on empty order, and vacuous_subcases records acceptance of empty proof.
- Nil Lower output aborts the test binary; all fourteen isolated recovery runs were necessary. Panics are observed failures, not inferred failures for unseen rows.
- InputSpreadCoverage has seven preparation-only checker subcases. Mutating Load or TypeScript would violate the declared lowering scope; those cases are reported as passing the lowering probe.
- The absence-only const-enum oracle passes an empty IR through POrder. This does not make it sacred or establish primary-entry vacuity; it is an additional demonstrated weakness.
- The first npm redirect used a relative evidence directory from stage3/api and failed before installation. It was corrected to the absolute root evidence path.
- Logs are ignored by repository policy and were force-added only under the authorized evidence directory.

Timings: nproc=5; Go 1.27.1, Node 24.19.0. Full toolchain setup skipped because env.sh worked. Exact total setup and cold compile-only durations were not instrumented, so cannot precisely report them; cooked discovery observations are above. Clean coverage slice wall: 18.921s. Clean coverage plus 42 isolated timing commands: 136.789s wall total. Matrix/probes: 226.804s wall total, 167.416s reported binary elapsed total. First switched matrix command: 18.504s including Go rebuild, native builds and tests; subsequent command walls are in matrix-times.txt. Panic recovery: 34.723s wall. All 23 standalone diff vet checks: 14.959s wall total, each exit zero. Native build-only times were not separately instrumented; per-subcase elapsed times include lowering, builds and executions and are in each JSON log. Three-round binary medians are in rows.json/timings.json. No unit row exceeds 60s.

Validation: every standalone diff applies to the exact base and compiles with go vet ./internal/lower/; production source is restored exactly, and restored.log records a clean passing slice. No main push or PR. Not covered: kills outside the fourteen rows, repo-wide uniqueness, exhaustive mutation coverage, external validation of the recorded Node order, live native tsc for the ledger, or isolated native compile-only timings. The full clean package skipped TestOptionalWideningCensus (outside this unit, opt-in source/corpus root) and one explicitly pending MixedUnionContractGraph subcase; no requested row skipped.
