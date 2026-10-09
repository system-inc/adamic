u058: all 13 names exist at c0a7667baadaf161d6bb9066e0838b32774caa6c; none moved or vanished.
Whole oracle timed out at 90.309 s without an earlier test failure; scoped baseline passed in 11.347 s; nproc 5.
14 eligible production mutants, M1 supplemental, three entry probes and one witness check.
Bounded verdicts: 7 subsumed, 5 sacred, 1 witness.
Evidence: test-audit/internal-oracle-checked_views_v2_source, review/test-audit/internal-oracle-checked_views_v2_source/.

```json
[
  {
    "test": "TestCheckedViewUntaggedSourceDispatch",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_source_test.go",
    "seconds": 2.596,
    "oracle": "Node exit/stdout/stderr comparisons for good controls; self-written source-refusals.json pins for wrong/nested. Absent cases only require a diagnostic substring and clean Node exit.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M6",
      "M8",
      "M9",
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: checked_views_v2_source_test.go:23: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.a:16:1: Adamic 0.1 refuses var; use const or let",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewUntaggedRecursive"
    ],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.842,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; checked_views_v2_source_test.go:23: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.a:16:1: Adamic 0.1 refuses var; use const or let",
    "subsumption_mutants": 4,
    "vacuous_subcases": [
      "initial binding-name-source-good admission control: P1 returns nil,nil and reaches the subtests"
    ]
  },
  {
    "test": "TestCheckedViewUntaggedCallableUnion",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_source_test.go",
    "seconds": 1.009,
    "oracle": "Node must print true; successful backends compare to Node, wrong/nested cases use self-written exact panic pins.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M7",
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: checked_views_v2_source_test.go:84: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/callable-union-good-number.a:4:1: Adamic 0.1 refuses var; use const or let",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewUntaggedOptionalCallableControl"
    ],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.684,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; checked_views_v2_source_test.go:84: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/callable-union-good-number.a:4:1: Adamic 0.1 refuses var; use const or let",
    "subsumption_mutants": 2
  },
  {
    "test": "TestCheckedViewUntaggedOptionalCallableControl",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_source_test.go",
    "seconds": 0.684,
    "oracle": "Node executes and backends compare exit/stdout/stderr; self-written true output also checks Node.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M7",
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: checked_views_v2_source_test.go:114: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/callable-union-optional-boundary.a:4:1: Adamic 0.1 refuses var; use const or let",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewUntaggedCallableUnion"
    ],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 1.009,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; checked_views_v2_source_test.go:114: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/callable-union-optional-boundary.a:4:1: Adamic 0.1 refuses var; use const or let",
    "subsumption_mutants": 2
  },
  {
    "test": "TestCheckedViewUntaggedSourceFlows",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_source_test.go",
    "seconds": 1.449,
    "oracle": "Node controls must print true; self-written exact output and panic expectations decide each backend result.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M6",
      "M8",
      "M9",
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: checked_views_v2_source_test.go:132: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/flow-helper-good.a:17:1: Adamic 0.1 refuses var; use const or let",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewUntaggedRecursive"
    ],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.842,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; checked_views_v2_source_test.go:132: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/flow-helper-good.a:17:1: Adamic 0.1 refuses var; use const or let",
    "subsumption_mutants": 4
  },
  {
    "test": "TestCheckedViewUntaggedOwnClassData",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_source_test.go",
    "seconds": 0.622,
    "oracle": "Node executes and good backend output agrees; wrong case uses a self-written panic pin.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M6",
      "M9",
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: checked_views_v2_source_test.go:156: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/class-data-good.a:6:1: Adamic 0.1 refuses var; use const or let",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewUntaggedRecursive"
    ],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.842,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; checked_views_v2_source_test.go:156: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/class-data-good.a:6:1: Adamic 0.1 refuses var; use const or let",
    "subsumption_mutants": 3
  },
  {
    "test": "TestCheckedViewUntaggedRecursive",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_source_test.go",
    "seconds": 0.842,
    "oracle": "Node executes and good/absent backend output agrees; wrong/nested use self-written panic pins.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M5",
      "M6",
      "M8",
      "M9",
      "M14"
    ],
    "unique_kills": [
      "M5"
    ],
    "last_proven_fail": "M14: checked_views_v2_source_test.go:180: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/recursive-good.a:6:1: Adamic 0.1 refuses var; use const or let",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1",
      "P2",
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; checked_views_v2_source_test.go:180: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/recursive-good.a:6:1: Adamic 0.1 refuses var; use const or let"
  },
  {
    "test": "TestCheckedViewUntaggedArrayPending",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_source_test.go",
    "seconds": 0.23,
    "oracle": "Self: Lower must return NotYet with an array-element or never-array reason. All five execution subcases then skip. No Node run in this row.",
    "oracle_kind": "self",
    "kills": [
      "M6",
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: checked_views_v2_source_test.go:211: array admission must stay NotYet: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/array-union-good.a:6:1: Adamic 0.1 refuses var; use const or let",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewUntaggedOwnClassData"
    ],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.622,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; checked_views_v2_source_test.go:211: array admission must stay NotYet: /workspace/adamic/stage3/interface-downcasts/untagged/fixtures/array-union-good.a:6:1: Adamic 0.1 refuses var; use const or let",
    "subsumption_mutants": 2,
    "skipped_subcases": [
      "good",
      "wrong",
      "nested",
      "empty",
      "mixed"
    ]
  },
  {
    "test": "TestClassWrongOutput103",
    "package": "internal/oracle",
    "file": "internal/oracle/class_wrong_output_test.go",
    "seconds": 0.317,
    "oracle": "Node executes and self-written stdout control plus exact refusal path and repair are checked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M10",
      "M14"
    ],
    "unique_kills": [
      "M10"
    ],
    "last_proven_fail": "M14: class_wrong_output_test.go:31: want path and repair, got /workspace/adamic/internal/oracle/testdata/class_wrong_output_refused/classfeat_init_super.a:3:7: Adamic 0.1 refuses var; use const or let",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; class_wrong_output_test.go:31: want path and repair, got /workspace/adamic/internal/oracle/testdata/class_wrong_output_refused/classfeat_init_super.a:3:7: Adamic 0.1 refuses var; use const or let"
  },
  {
    "test": "TestClassWrongOutput107",
    "package": "internal/oracle",
    "file": "internal/oracle/class_wrong_output_test.go",
    "seconds": 0.221,
    "oracle": "Node executes and self-written stdout control plus exact iterator-origin refusal and repair are checked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M11"
    ],
    "unique_kills": [
      "M11"
    ],
    "last_proven_fail": "M11: class_wrong_output_test.go:56: want pinned receiver-origin refusal, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; class_wrong_output_test.go:56: want pinned receiver-origin refusal, got <nil>"
  },
  {
    "test": "TestClassWrongOutput108",
    "package": "internal/oracle",
    "file": "internal/oracle/class_wrong_output_test.go",
    "seconds": 0.19,
    "oracle": "Node executes and self-written stdout control plus exact symbol-key-view refusal and repair are checked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M13"
    ],
    "unique_kills": [
      "M13"
    ],
    "last_proven_fail": "M13: class_wrong_output_test.go:75: want pinned symbol-key-view refusal, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; class_wrong_output_test.go:75: want pinned symbol-key-view refusal, got <nil>"
  },
  {
    "test": "TestClassWrongOutput106",
    "package": "internal/oracle",
    "file": "internal/oracle/class_wrong_output_test.go",
    "seconds": 0.243,
    "oracle": "Node executes and self-written stdout control plus exact private-storage refusal and repair are checked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M12"
    ],
    "unique_kills": [
      "M12"
    ],
    "last_proven_fail": "M12: class_wrong_output_test.go:98: want pinned private-instance-from-static refusal, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; class_wrong_output_test.go:98: want pinned private-instance-from-static refusal, got <nil>"
  },
  {
    "test": "TestClockGenericReturnsT01Mutant",
    "package": "internal/oracle",
    "file": "internal/oracle/clock_generic_returns_t_01_test.go",
    "seconds": 0.402,
    "oracle": "Witness: Node output, clean planted-mutant execution, exact stdout difference and leak check. Only W1 weakening disagreement decides its verdict.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: clock_generic_returns_t_01_test.go:58: native mutant: want stdout disagreement, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 14,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^TestClockGenericReturnsT01Mutant$'; clock_generic_returns_t_01_test.go:58: native mutant: want stdout disagreement, got \"\""
  },
  {
    "test": "TestClosureMergeRefusals",
    "package": "internal/oracle",
    "file": "internal/oracle/closure_merge_refusals_test.go",
    "seconds": 3.996,
    "oracle": "Node checks only zero exit and empty stderr, not stdout. Self-written .refused snapshots pin the complete lowering diagnostic.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14: closure_merge_refusals_test.go:50: diagnostic: got \"stage3/fixtures/nested-functions/refused/01_scanner_frame.a:49:5: Adamic 0.1 refuses var; use const or let\", want \"stage3/fixtures/nested-functions/refused/01_scanner_frame.a:69:5: Adamic 0.1 refuses var; use const or let\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewUntaggedArrayPending"
    ],
    "mutants_in_matrix": 14,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.23,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewUntaggedSourceDispatch",
      "TestCheckedViewUntaggedCallableUnion",
      "TestCheckedViewUntaggedOptionalCallableControl",
      "TestCheckedViewUntaggedSourceFlows",
      "TestCheckedViewUntaggedOwnClassData",
      "TestCheckedViewUntaggedRecursive",
      "TestCheckedViewUntaggedArrayPending",
      "TestClassWrongOutput103",
      "TestClassWrongOutput107",
      "TestClassWrongOutput108",
      "TestClassWrongOutput106",
      "TestClockGenericReturnsT01Mutant",
      "TestClosureMergeRefusals"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewUntaggedSourceDispatch|TestCheckedViewUntaggedCallableUnion|TestCheckedViewUntaggedOptionalCallableControl|TestCheckedViewUntaggedSourceFlows|TestCheckedViewUntaggedOwnClassData|TestCheckedViewUntaggedRecursive|TestCheckedViewUntaggedArrayPending|TestClassWrongOutput103|TestClassWrongOutput107|TestClassWrongOutput108|TestClassWrongOutput106|TestClockGenericReturnsT01Mutant|TestClosureMergeRefusals)$'; closure_merge_refusals_test.go:50: diagnostic: got \"stage3/fixtures/nested-functions/refused/01_scanner_frame.a:49:5: Adamic 0.1 refuses var; use const or let\", want \"stage3/fixtures/nested-functions/refused/01_scanner_frame.a:69:5: Adamic 0.1 refuses var; use const or let\"",
    "subsumption_mutants": 1
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M1 | internal/lower/view_unions_untagged.go:51 | if field.Optional { -> if !field.Optional { |  |
| M2 | internal/lower/view_unions_untagged.go:88 | contract.Of == ir.Boolean \|\| contract.Of == ir.MaybeNumber -> contract.Of == ir.Array \|\| contract.Of == ir.MaybeNumber |  |
| M3 | internal/lower/view_unions_untagged.go:182 | for index, contract := range l.result.ViewContracts { 		if contract.Unsupported == "untagged object union" && l.supportsUntaggedRead(contract) { 			l.result.ViewContracts[index].Unsupported = "" 		} 	} ->  |  |
| M4 | internal/lower/view_unions_untagged.go:257 | contract.ProducerCertified = true -> contract.ProducerCertified = false |  |
| M5 | internal/lower/view_unions_untagged.go:96 | if seen[id] { 			return true -> if seen[id] { 			return false | TestCheckedViewUntaggedRecursive |
| M6 | internal/lower/view_unions_dispatch.go:25 | if child.Kind == ir.ViewArray { -> if child.Kind == ir.ViewObject { | TestCheckedViewUntaggedSourceDispatch, TestCheckedViewUntaggedSourceFlows, TestCheckedViewUntaggedOwnClassData, TestCheckedViewUntaggedRecursive, TestCheckedViewUntaggedArrayPending |
| M7 | internal/lower/view_unions_callable_members.go:13 | len(signatures) != 1 -> len(signatures) != 2 | TestCheckedViewUntaggedCallableUnion, TestCheckedViewUntaggedOptionalCallableControl |
| M8 | internal/native/view_unions_untagged.go:59 | field.Contract, field.Optional) -> field.Contract, !field.Optional) | TestCheckedViewUntaggedSourceDispatch, TestCheckedViewUntaggedSourceFlows, TestCheckedViewUntaggedRecursive |
| M9 | internal/javascript/view_unions_untagged.go:65 | if(depth>128) return false; -> if(depth>0) return false; | TestCheckedViewUntaggedSourceDispatch, TestCheckedViewUntaggedSourceFlows, TestCheckedViewUntaggedOwnClassData, TestCheckedViewUntaggedRecursive |
| M10 | internal/lower/class_inheritance.go:646 | && !available[l.fieldName(child.Name())] { -> && available[l.fieldName(child.Name())] { | TestClassWrongOutput103 |
| M11 | internal/lower/iteration.go:229 | if !l.iterationOrigin(where, members, &presence, 0) { -> if l.iterationOrigin(where, members, &presence, 0) { | TestClassWrongOutput107 |
| M12 | internal/lower/class_static.go:516 | target.Parent == declaration && !ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic) -> target.Parent == declaration && ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic) | TestClassWrongOutput106 |
| M13 | internal/lower/class_features.go:89 | if !fresh && !isClassInstance(proven) -> if fresh && !isClassInstance(proven) | TestClassWrongOutput108 |
| M14 | internal/lower/locals.go:13 | list.Flags&ast.NodeFlagsBlockScoped == 0 -> list.Flags&ast.NodeFlagsBlockScoped != 0 | TestCheckedViewUntaggedSourceDispatch, TestCheckedViewUntaggedCallableUnion, TestCheckedViewUntaggedOptionalCallableControl, TestCheckedViewUntaggedSourceFlows, TestCheckedViewUntaggedOwnClassData, TestCheckedViewUntaggedRecursive, TestCheckedViewUntaggedArrayPending, TestClassWrongOutput103, TestClockGenericReturnsT01Mutant, TestClosureMergeRefusals |
| M15 | internal/lower/view_lazy.go:197 | if strings.Contains(family, "views-v3: array element kind") { -> if !strings.Contains(family, "views-v3: array element kind") { |  |
| W1 | internal/oracle/oracle_test.go:716 | func disagreement(oracle run, native run) string { 	switch { 	case oracle.exitCode != native.exitCode: 		return "exit codes differ" 	case !bytes.Equal(oracle.stdout, native.stdout): 		return "stdout differs" 	case !bytes.Equal(oracle.stderr, native.stderr): 		return "stderr differs" 	} 	return "" }  -> func disagreement(oracle run, native run) string { 	return "" }  | TestClockGenericReturnsT01Mutant |

Survivors:

- M1 supplemental, outside reached code: optional tag changes from no tags to tag=x in the direct exported-planner probe. Coverage says UntaggedViewMembers 0.0%; excluded from all verdicts.
- M2: supportsUntaggedRead on a synthetic Boolean field-only union changes true to false. The tested source Boolean fixture itself did not change; direct function witness establishes the changed output.
- M3: completeUntaggedRecursiveContracts on a valid incomplete descriptor changes cleared Unsupported="" to retained "untagged object union". Direct function witness, not a test kill.
- M4: Lower output for callable-union-good-number changes two ProducerCertified flags from true to false; all requested rows still pass.
- M15: direct checkLazyViewReads changes NotYet array-element metadata refusal to Refused for the same synthetic IR; no requested row fails. Logs retain both complete diagnostics.

Brief ambiguities, costs and limitations:

- Supplied historical SHA differs from the required current origin/main start. Every requested name remains in its supplied file.
- Whole-package oracle exceeds 90 seconds. I ran the 13 requested rows as a bounded slice and obtained production coverage for 629 functions across lower/native/javascript. Other rows and package-wide uniqueness are unknown. The coverage inventory is dynamic Go reachability, not a complete C-runtime call graph.
- These bodies have distinct assertions, so all 13 remain separate rows. The shared disagreement helper alone does not make different refusal and contract checks one family. Internal subcases stay grouped under their top-level test.
- Three mutants per row would require 39, beyond the ceiling of 20. I fixed 15 attempts before results; one was later excluded because its old planner is not reached. Fourteen eligible mutants remain. The M1 selection was my audit mistake, not a weakness of the brief. Its diff and observed result remain supplemental for transparency.
- Standalone validation found that ir.Void is a type, not an enum value. M2 was corrected to ir.Array before any mutant outcomes.
- All production mutants are Go compiler/emitter changes under one switch. Each selector has a distinct ADAMIC_BUILD_CACHE_DIR; ADAMIC_GATE_UNCACHED=1 bypasses oracle result caches, including the environment-selected generated JavaScript check. Runtime source archives use their content/flag cache and remain unchanged. The at-most-four fallback was unnecessary.
- TestClockGenericReturnsT01Mutant is a witness. Its failures under M14 or empty compiler entries are broken preconditions and contribute no production kills. W1 disables only disagreement in a scratch copy and proves the witness fails. Node and the test itself remain unchanged.
- TestCheckedViewUntaggedArrayPending is not missing a tool or opt-in. Every subcase asserts NotYet and then skips execution because the compiler lacks V3 array metadata. Its five skips are reported; no claim of supported native/JavaScript behavior is made.
- Empty Lower aborts parallel runs when helpers dereference nil programs. Lost rows were rerun individually. Only own-entry results determine vacuity. The witness has null vacuity because production entry probes do not judge its comparison check.
- SourceDispatch initially checks only a nil error and ignores the program; that admission control passes the empty Lower probe. Its later negative and execution subcases fail, so the row is not vacuous.
- Node controls and refusal pins have different authority. The tests actually run Node, but Adamic-only panic text, repairs and .refused snapshots are self. No external-authority value was claimed or checked. ClosureMerge checks Node exit/stderr only, so unrelated successful stdout would pass that control.
- Survivors require output witnesses, not speculation. Direct synthetic-IR calls establish M2/M3 output changes; M4 changes actual fixture IR. M1 demonstrates a change only outside the reached unit. Synthetic probes do not establish source-level runtime misbehavior.
- Subsumption is a small-matrix hint, not a deletion recommendation. Each row records how many kills support it and the fastest observed subsumer.
- Compiler rebuild time was not separately instrumented for each emitted native fixture. Command wall times include clang work; isolated binary medians remain the reported cost. No other package tests ran.

Timing: warm setup skipped (0 s); npm ci about 1 s; clean whole-package binary 90.309 s; bounded baseline 11.347 s. Command-wall totals: {"isolated_timing_wall": 143.35606850000067, "production_matrix_wall": 223.955219726, "probe_wall_with_abort_reruns": 64.63566964900019, "build_validation_wall": 36.43679748399882}. Raw commands and timings are saved. Native runtime rebuilds are included in test-command time, not separately measured. No full-package completion, repo-wide uniqueness replay, source-level witness for every survivor, exhaustive mutation coverage or complete C reachability is claimed. Production sources are restored.

Reproduce direct survivor witnesses from the repository root after sourcing env.sh: python review/test-audit/internal-oracle-checked_views_v2_source/replay-survivors.py. The script installs a temporary read-only bridge to unexported functions and applies each saved diff independently, then restores sources. Bridge Go files are stored as .go.txt so evidence does not create a package dependent on the removed scratch bridge.
