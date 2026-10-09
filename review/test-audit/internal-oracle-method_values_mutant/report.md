All 13 requested tests exist at the fetched starting commit.
Ten witnesses, two bounded subsumed rows, one bounded sacred row.
Ten production mutants and three witness harness faults; M09 alone survived.
Whole package cooked at 90s; scoped and selected-fixture baselines passed; nproc=5.
Sources restored; standalone diffs and all raw evidence saved for central replay.

```json
[
  {
    "test": "TestLibraryMethodValueMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/method_values_mutant_test.go",
    "seconds": 0.69,
    "oracle": "Node source observation compared by disagreement.",
    "oracle_kind": "external-run",
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01: method_values_mutant_test.go:173: want Node stdout alone to kill mutant, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W01-standalone.log: method_values_mutant_test.go:173: want Node stdout alone to kill mutant, got \"\""
  },
  {
    "test": "TestLibraryMethodOwnLoadMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/method_values_mutant_test.go",
    "seconds": 0.372,
    "oracle": "Node differential plus self-written exit 70 and missing-field panic marker precondition.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01: method_values_mutant_test.go:215: source oracle accepted an own-field load of an intrinsic",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W01-standalone.log: method_values_mutant_test.go:215: source oracle accepted an own-field load of an intrinsic"
  },
  {
    "test": "TestLibraryMethodCaptureMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/method_values_ownership_test.go",
    "seconds": 0.652,
    "oracle": "ASan heap-use-after-free and LeakSanitizer marker recognition; extra-owner behavior also agrees with source Node. Marker literals are self-written.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "W02"
    ],
    "unique_kills": [],
    "last_proven_fail": "W02: method_values_ownership_test.go:46: want ASan alone to catch borrowed capture, got exit -1 stderr =================================================================",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W02.log: method_values_ownership_test.go:46: want ASan alone to catch borrowed capture, got exit -1 stderr ================================================================="
  },
  {
    "test": "TestLibraryMethodAliasReadMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/method_values_read_test.go",
    "seconds": 0.105,
    "oracle": "Node source observation compared by disagreement.",
    "oracle_kind": "external-run",
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01: method_values_read_test.go:36: want Node's temporal dead zone to catch an otherwise finishing mutant: node 70 stderr adamic: panic: ReferenceError: Cannot access 'search' before initialization",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W01-standalone.log: method_values_read_test.go:36: want Node's temporal dead zone to catch an otherwise finishing mutant: node 70 stderr adamic: panic: ReferenceError: Cannot access 'search' before initialization"
  },
  {
    "test": "TestModuleNamespaceReadsMatchNode",
    "package": "internal/oracle",
    "file": "internal/oracle/module_namespace_reads_test.go",
    "seconds": 0.702,
    "oracle": "Node source observation compared by disagreement.",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M02"
    ],
    "unique_kills": [],
    "last_proven_fail": "M02: module_namespace_reads_test.go:31: /workspace/adamic/internal/oracle/testdata/module_namespace_reads/main.a:2:36: stage 0 can't lower reading performance yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNativeAgreesWithNode family"
    ],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 6.081,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; M02.log: module_namespace_reads_test.go:31: /workspace/adamic/internal/oracle/testdata/module_namespace_reads/main.a:2:36: stage 0 can't lower reading performance yet",
    "subsumption_mutants": 2
  },
  {
    "test": "TestModuleNamespaceReadinessMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/module_namespace_reads_test.go",
    "seconds": 0.344,
    "oracle": "Node source observation compared by disagreement.",
    "oracle_kind": "external-run",
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01: module_namespace_reads_test.go:77: native removed-readiness mutant survived: {stdout:[48 10] stderr:[] exitCode:0}",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W01-standalone.log: module_namespace_reads_test.go:77: native removed-readiness mutant survived: {stdout:[48 10] stderr:[] exitCode:0}"
  },
  {
    "test": "TestModuleNamespaceLiveBindingMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/module_namespace_reads_test.go",
    "seconds": 0.222,
    "oracle": "Node source observation compared by disagreement.",
    "oracle_kind": "external-run",
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01: module_namespace_reads_test.go:112: native frozen-export mutant survived: {stdout:[48 58 49 58 50 58 48 10 116 114 117 101 10 56 48 58 51 58 48 10] stderr:[] exitCode:0}",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W01-standalone.log: module_namespace_reads_test.go:112: native frozen-export mutant survived: {stdout:[48 58 49 58 50 58 48 10 116 114 117 101 10 56 48 58 51 58 48 10] stderr:[] exitCode:0}"
  },
  {
    "test": "TestNamespaceLiveExportBoundary",
    "package": "internal/oracle",
    "file": "internal/oracle/namespace_live_export_boundary_test.go",
    "seconds": 0.321,
    "oracle": "Node executes source and validates handwritten stdout, then both backends compare byte-for-byte with Node.",
    "oracle_kind": "external-run",
    "kills": [
      "M04",
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: namespace_live_export_boundary_test.go:23: stdout differs",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNativeAgreesWithNode family"
    ],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 6.081,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; M10.log: namespace_live_export_boundary_test.go:23: stdout differs",
    "subsumption_mutants": 2
  },
  {
    "test": "TestNamespaceRuledMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/namespace_ruled_mutants_test.go",
    "seconds": 0.352,
    "oracle": "Node source observation compared by disagreement.",
    "oracle_kind": "external-run",
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01: namespace_ruled_mutants_test.go:62: mutant survived JavaScript differential",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W01-standalone.log: namespace_ruled_mutants_test.go:62: mutant survived JavaScript differential"
  },
  {
    "test": "TestNamespaceSemanticMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/namespaces_test.go",
    "seconds": 0.841,
    "oracle": "Node source observation compared by disagreement.",
    "oracle_kind": "external-run",
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01: namespaces_test.go:196: not caught by Node: \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W01-standalone.log: namespaces_test.go:196: not caught by Node: \"\""
  },
  {
    "test": "TestNamespaceStateMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/namespaces_test.go",
    "seconds": 0.366,
    "oracle": "Node differential for assignment and hoisting; readiness mutant compares only exit code with Adamic JavaScript backend, which can accept a different failure with the same code.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "W01",
      "W03"
    ],
    "unique_kills": [],
    "last_proven_fail": "W03: namespaces_test.go:260: ready check mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W03.log: namespaces_test.go:260: ready check mutant survived"
  },
  {
    "test": "TestNarrowedUnionMemberCheck",
    "package": "internal/oracle",
    "file": "internal/oracle/narrowed_union_test.go",
    "seconds": 0.16,
    "oracle": "Node validates source stdout; handwritten exit 70 and complete Adamic panic text decide inserted-check behavior. Built-in unchecked mutant witnesses that comparison.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "W01"
    ],
    "unique_kills": [],
    "last_proven_fail": "W01: narrowed_union_test.go:68: unchecked member mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; W01-standalone.log: narrowed_union_test.go:68: unchecked member mutant survived"
  },
  {
    "test": "TestNarrowedUnionObjectTagRefusal",
    "package": "internal/oracle",
    "file": "internal/oracle/narrowed_union_test.go",
    "seconds": 0.127,
    "oracle": "Node validates source stdout 1; handwritten capability-gap text, path and rewrite decide refusal. No external authority checked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M06"
    ],
    "unique_kills": [
      "M06"
    ],
    "last_proven_fail": "M06: narrowed_union_test.go:87: want refusal naming the path and fix, got /workspace/adamic/internal/oracle/testdata/reland_refused/narrowed_union_object_tag.a:7:16: stage 0 can't lower a narrowed union member whose object tag cannot be checked with instanceof; keep differently held object kinds in separately typed variables yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLibraryMethodValueMutants",
      "TestLibraryMethodOwnLoadMutants",
      "TestLibraryMethodCaptureMutants",
      "TestLibraryMethodAliasReadMutant",
      "TestModuleNamespaceReadsMatchNode",
      "TestModuleNamespaceReadinessMutants",
      "TestModuleNamespaceLiveBindingMutant",
      "TestNamespaceLiveExportBoundary",
      "TestNamespaceRuledMutants",
      "TestNamespaceSemanticMutants",
      "TestNamespaceStateMutants",
      "TestNarrowedUnionMemberCheck",
      "TestNarrowedUnionObjectTagRefusal",
      "TestNativeAgreesWithNode family"
    ],
    "evidence": "python3 review/test-audit/internal-oracle-method_values_mutant/run.py > review/test-audit/internal-oracle-method_values_mutant/runner.log 2>&1; M06.log: narrowed_union_test.go:87: want refusal naming the path and fix, got /workspace/adamic/internal/oracle/testdata/reland_refused/narrowed_union_object_tag.a:7:16: stage 0 can't lower a narrowed union member whose object tag cannot be checked with instanceof; keep differently held object kinds in separately typed variables yet"
  }
]
```

| ID | Starting file:line | Change | Observed failed tests |
|---|---|---|---|
| M01 | internal/lower/load_time_reads.go:102 | `!l.provenModuleReads[node] && l.checked(local)` -> `l.provenModuleReads[node] && l.checked(local)` | TestModuleNamespaceReadinessMutants, TestModuleNamespaceReadsMatchNode, TestNativeAgreesWithNode |
| M02 | internal/lower/module_namespace.go:17 | `			return true` -> `			return false` | TestModuleNamespaceLiveBindingMutant, TestModuleNamespaceReadinessMutants, TestModuleNamespaceReadsMatchNode, TestNativeAgreesWithNode |
| M03 | internal/lower/namespaces.go:400 | `Value: ir.BooleanConstant{Value: false}})` -> `Value: ir.BooleanConstant{Value: true}})` | TestNativeAgreesWithNode |
| M04 | internal/lower/namespaces.go:502 | `body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: true}})` -> `body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: false}})` | TestNamespaceLiveExportBoundary, TestNamespaceSemanticMutants, TestNamespaceStateMutants, TestNativeAgreesWithNode |
| M05 | internal/lower/locals.go:49 | `l.result.Locals[local].NamespaceVar && declaration.Initializer() == nil` -> `l.result.Locals[local].NamespaceVar && declaration.Initializer() != nil` | TestNamespaceSemanticMutants, TestNativeAgreesWithNode |
| M06 | internal/lower/locals.go:273 | `a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables` -> `a narrowed union member whose object tag cannot be checked with instanceof; keep differently held object kinds in separately typed variables` | TestNarrowedUnionObjectTagRefusal |
| M07 | internal/lower/locals.go:279 | `Operator: ir.Equal, Left: ir.TypeOf{Value: held}` -> `Operator: ir.NotEqual, Left: ir.TypeOf{Value: held}` | TestNarrowedUnionMemberCheck, TestNativeAgreesWithNode |
| M08 | internal/lower/locals.go:256 | `name = "number"` -> `name = "string"` | TestNarrowedUnionMemberCheck, TestNativeAgreesWithNode |
| M09 | internal/lower/locals.go:272 | `known && held != narrowed && (held == ir.Object` -> `known && held == narrowed && (held == ir.Object` |  |
| M10 | internal/lower/namespaces.go:432 | `b.read(b.parameters[len(checks)])` -> `b.read(b.parameters[len(checks)-1])` | TestNamespaceLiveExportBoundary, TestNamespaceRuledMutants, TestNamespaceSemanticMutants, TestNativeAgreesWithNode |
| W01 | internal/oracle/oracle_test.go:718,720,722 | disable all three disagreement comparisons; authorized witness harness weakening | TestLibraryMethodAliasReadMutant, TestLibraryMethodOwnLoadMutants, TestLibraryMethodValueMutants, TestModuleNamespaceLiveBindingMutant, TestModuleNamespaceReadinessMutants, TestNamespaceRuledMutants, TestNamespaceSemanticMutants, TestNamespaceStateMutants, TestNarrowedUnionMemberCheck |
| W02 | internal/oracle/method_values_ownership_test.go:45,54 | disable ASan and LeakSanitizer marker recognition; authorized witness harness weakening | TestLibraryMethodCaptureMutants |
| W03 | internal/oracle/namespaces_test.go:259 | force direct exit comparison to report agreement; authorized witness harness weakening | TestNamespaceStateMutants |

Survivor M09: survivor-clean.log reports `ADMISSION accepted: 2 functions`; survivor-M09.log reports the object-tag capability-gap refusal for the same scratch source.

The fetched origin/main was 09769cb5ddc8067d2fa052a8c760d813894c8919, not the historical 8de93800f4 cited in the brief. All 13 requested names exist in the cited files. Bodies have distinct checks, so no scoped rows collapse into a family. The added TestNativeAgreesWithNode family is bounded to 35 selected fixtures, listed in family-members.json.

Ten scoped rows are witnesses, including TestNarrowedUnionMemberCheck with its positive checks and built-in negative mutant. Production failures in witnesses are recorded but excluded from their verdicts and production kill sets. W01 demonstrates the differential comparison; W02 demonstrates ASan/LeakSanitizer marker recognition; W03 separately demonstrates the namespace-state exit-only comparison. W03 and W02 edit only the check expressions in witness tests, under the brief's explicit harness exception. W01-standalone.log reproduces the selector result using the actual standalone comparator diff. Witness vacuity is null: production-entry probes do not judge witnesses.

The full package baseline timed out at 90.131 binary seconds, with no assertion failure before timeout. It is cooked, not red. The scoped baseline passed at 8.165 binary seconds and the final uncached scoped run passed at 13.188 seconds. The additional family had three clean runs, median 6.081 seconds. No narrowed matrix run cooked or aborted. Every production column observed all 13 requested tests and all 35 selected family fixtures. Other package rows are unknown. Bounded sacred and subsumption verdicts are provisional for central replay.

The first combined fixture regex selected the parent but zero children because Go splits names at slash boundaries. This cost ten parent-only selections. Corrected per-component family replays supply the real observations; both logs are preserved. Matrix completion was checked explicitly. Cache-status lines were excluded when extracting failing assertions.

The fixed plan has ten production mutations for the three ordinary rows, plus three shared harness faults for the ten witnesses. It does not pursue twenty production faults against witness preconditions. Subsumption rests on two production kills per ordinary runtime row and is a hint, not a deletion recommendation.

Coverage was collected on the clean scoped tests before mutants. coverage-functions.txt lists all measured production functions; reached-production-functions.txt filters positive function coverage. Coverage attribution is aggregate across the scope, not an exact per-row call graph. The source-based conservative inventory was frozen before edits. No claim of exhaustive mutation coverage is made.

M09 is a survivor with changed admission: a scratch object-or-number program lowers cleanly before the mutant and is refused afterwards. This is a changed compiler answer, not a measured native performance or code-generation difference. The scratch witness and raw logs are saved. M03 survived the requested rows but was caught by the corrected fixture-family replay.

The reported timing medians are test-binary package lines from independent count=1 runs, using normal warm oracle caches. A separate uncached clean run verifies fresh Node/native/sanitizer execution. Initial instrumented and coverage compilation costs are included in command overhead and were not independently timed. The later clean warm build was measured separately. No other packages, complete repo gate, SDK/corpus opt-in installation or repo-wide uniqueness replay was run. Scoped rows and selected family members did not skip. Full-baseline skip information is saved in skipped.json; the full run did not reach every package row.

Production and harness files were restored. No PR was opened and no main push was made.

Measured command totals:
```json
{
  "runs.jsonl": {
    "commands": 17,
    "command_seconds": 197.6860907310038
  },
  "family-runs.jsonl": {
    "commands": 13,
    "command_seconds": 91.65139207099855
  },
  "timing-commands.jsonl": {
    "commands": 39,
    "command_seconds": 94.74672758000088
  },
  "vet-times.jsonl": {
    "commands": 10,
    "command_seconds": 5.312340096003027
  },
  "finished_utc": "2026-10-09T11:19:51.101746+00:00",
  "reached_production_functions": 562
}
```
Setup, clean build and baseline timing records are separate JSON files; timings include a final uncached clean validation.

Seven skips observed outside the selected scope in the cooked whole-package baseline: TestWASIShardPlantedFixture, TestWASIAgreesWithNode, TestWASIOracleCatchesMutants, TestWASIRunnerCatchesMutants, TestWASIEmission, TestStage3FixtureHook, TestEntriesAcceptance. None were part of the bounded matrix. Their messages are preserved in skipped.json; no SDK or external corpus was installed for these out-of-scope rows.
