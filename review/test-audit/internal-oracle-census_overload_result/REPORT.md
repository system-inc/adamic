Audited 68db8ddd145281a62655452496bdc32ef848bdf3; all three requested functions exist, with no grouping or moves.
Whole-package baseline timed out at 90.251 binary seconds without an individual failure; bounded baseline and restored control pass.
Two bounded sacred rows, one cannot-judge external-oracle-only row; append proof is vacuous under empty IR.
Nine menu mutants, separate nil and empty-IR entry probes, one survivor with a changed-function witness.
Production restored; evidence-only branch for central replay, nproc=5.

```json
[
  {
    "test": "TestCensusAppendResultProof",
    "package": "internal/oracle",
    "file": "internal/oracle/census_overload_result_test.go:21",
    "seconds": 0.066,
    "oracle": "Handwritten absence of strings containing of append result:. No positive IR-content or execution assertion; an empty IR program passes P02.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M03",
      "M04",
      "M05",
      "M06"
    ],
    "unique_kills": [
      "M03",
      "M04",
      "M05",
      "M06"
    ],
    "last_proven_fail": "M06 census_overload_result_test.go:33: the real append body should prove its resolved results, got check \"overload 1 of append result: expected number[], got undefined\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestCensusAppendResultProof",
      "TestCensusOverloadResultStop",
      "TestCensusSmallStoppedSourceOnNode",
      "TestNativeAgreesWithNode family (append and lie inputs only)"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=M06 ADAMIC_BUILD_CACHE_DIR=/tmp/u056/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCensusAppendResultProof|TestCensusOverloadResultStop|TestCensusSmallStoppedSourceOnNode)$'; census_overload_result_test.go:33: the real append body should prove its resolved results, got check \"overload 1 of append result: expected number[], got undefined\"",
    "timing_samples": [
      0.066,
      0.056,
      0.072
    ],
    "empty_answer_probe": "P02",
    "nil_probe_failed_by_dereference": true,
    "empty_ir_passes": true
  },
  {
    "test": "TestCensusOverloadResultStop",
    "package": "internal/oracle",
    "file": "internal/oracle/census_overload_result_test.go:40",
    "seconds": 0.6,
    "oracle": "Untouched source actually runs on Node and must match handwritten called/undefined stdout. Native and emitted JavaScript must match handwritten stdout, exact panic stderr and exit 70. These checked backend results deliberately differ from the original source.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M02",
      "M07",
      "M08"
    ],
    "unique_kills": [
      "M07",
      "M08"
    ],
    "last_proven_fail": "M08 census_overload_result_test.go:58: native: stderr differs; got oracle.run{stdout:[]uint8{0x63, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0xa}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x6f, 0x76, 0x65, 0x72, 0x6c, 0x6f, 0x61, 0x64, 0x20, 0x31, 0x20, 0x6f, 0x66, 0x20, 0x6c, 0x69, 0x65, 0x20, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x3a, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x2c, 0x20, 0x67, 0x6f, 0x74, 0x20, 0x75, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0xa}, exitCode:70}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [
      "P01",
      "P02"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCensusAppendResultProof",
      "TestCensusOverloadResultStop",
      "TestCensusSmallStoppedSourceOnNode",
      "TestNativeAgreesWithNode family (append and lie inputs only)"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=M08 ADAMIC_BUILD_CACHE_DIR=/tmp/u056/cache/M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCensusAppendResultProof|TestCensusOverloadResultStop|TestCensusSmallStoppedSourceOnNode)$'; census_overload_result_test.go:58: native: stderr differs; got oracle.run{stdout:[]uint8{0x63, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0xa}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x6f, 0x76, 0x65, 0x72, 0x6c, 0x6f, 0x61, 0x64, 0x20, 0x31, 0x20, 0x6f, 0x66, 0x20, 0x6c, 0x69, 0x65, 0x20, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x3a, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x74, 0x79, 0x70, 0x65, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x2c, 0x20, 0x67, 0x6f, 0x74, 0x20, 0x75, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0xa}, exitCode:70}",
    "timing_samples": [
      1.212,
      0.6,
      0.476
    ],
    "empty_answer_probe": "P02"
  },
  {
    "test": "TestCensusSmallStoppedSourceOnNode",
    "package": "internal/oracle",
    "file": "internal/oracle/census_small_test.go:34",
    "seconds": 0.289,
    "oracle": "Untouched stopped-source fixtures actually run on Node and must match handwritten stdout, zero exit and empty stderr. No Adamic lowering or backend is invoked.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCensusAppendResultProof",
      "TestCensusOverloadResultStop",
      "TestCensusSmallStoppedSourceOnNode",
      "TestNativeAgreesWithNode family (append and lie inputs only)"
    ],
    "evidence": "Bounded uncached baseline and three isolated runs pass. Production mutants do not apply: this row only invokes the external Node oracle on untouched fixtures.",
    "timing_samples": [
      0.289,
      0.298,
      0.255
    ],
    "reason": "No Adamic production entry is reached. A meaningful production mutant cannot affect the asserted observation; mutating Node or its reference input would violate the oracle restriction. Not labeled untrue."
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/census_overload_proof.go:27 | !l.censusHasUndefined(produced) -> l.censusHasUndefined(produced) | TestCensusAppendResultProof, TestCensusOverloadResultStop, TestNativeAgreesWithNode family (append and lie inputs only) |
| M02 | internal/lower/census_overload_proof.go:119 | !reachable(flow) -> reachable(flow) | TestCensusOverloadResultStop, TestNativeAgreesWithNode family (append and lie inputs only) |
| M03 | internal/lower/census_overload_proof.go:144 | len(flows) > 256 -> len(flows) > 0 | TestCensusAppendResultProof |
| M04 | internal/lower/census_overload_proof.go:170 | !trusted -> trusted | TestCensusAppendResultProof |
| M05 | internal/lower/census_overload_proof.go:58 | l.typeMapper = newTypeMapper(sources, targets)  -> | TestCensusAppendResultProof |
| M06 | internal/lower/census_overload_proof.go:102 | !unchanged -> unchanged | TestCensusAppendResultProof |
| M07 | internal/lower/census_small.go:295 | ordinal := 0 -> 			ordinal := 1 | TestCensusOverloadResultStop |
| M08 | internal/lower/census_small.go:312 | "overload %d of %s result: expected %s, got undefined" -> "overload %d of %s result: expected type %s, got undefined" | TestCensusOverloadResultStop |
| M09 | internal/lower/census_overload_proof.go:250 | admitted[index].Flags()&checker.TypeFlagsUndefined == 0 -> admitted[index].Flags()&checker.TypeFlagsUndefined != 0 |  |

Survivor:

M09: censusCallReturnsUndefined on append(undefined, undefined) returns [true] before and [false] after. See M09.before.log, M09.after.log, recognition-adapter.go.txt and survivor-commands.json. Both bounded test sets pass; this is unguarded recognition loss, not proven wrong program output.

Brief feedback, unit u056

1. The supplied historical reference is 8de93800f4, while the required fetch selected 68db8ddd145281a62655452496bdc32ef848bdf3. All three requested names remain in the supplied files.
2. The clean whole package times out after 90.251 test-binary seconds. No individual failure or error diagnostic appeared before the timeout. The three-row bounded baseline passes uncached in 0.505 binary seconds. Calling the timeout red would have prevented a valid bounded audit; the explicit big-package exception was necessary.
3. A selected oracle row can invoke no Adamic production code. TestCensusSmallStoppedSourceOnNode only runs the untouched reference bodies on Node. This brief supplies no dedicated oracle-only verdict. I used cannot-judge, with the precise reason, rather than mutate the oracle input or label the row untrue. If such rows should instead be judged as fixture setup checks, the brief should explicitly permit changing their reference fixtures under that exception.
4. The nil probe is not sufficient to establish oracle strength. Both lowering rows fail when Lower returns nil because they dereference the result. An additional zero-content IR probe passes AppendResultProof while OverloadResultStop fails. P02 is the useful empty-answer probe for the vacuous flag; P01 is recorded separately as a nil-precondition failure. Neither supports sacredness.
5. The stop row has two expected answers: untouched source behavior checked on Node, and a deliberately different Adamic loud stop with handwritten stdout, exact stderr and exit 70. It is external-run plus self, not a direct source-to-native agreement oracle.
6. The shared differential family checks exit 70 and agreement between the two Adamic backends for the lying fixture. It does not pin the expected panic message. M07 and M08 preserve both conditions and pass that checker, while the dedicated stop row fails on stderr. This is an observed strength difference.
7. The append row's expectation is only absence of a diagnostic substring, not execution or a positive IR structure assertion. It catches four proof regressions uniquely in the bounded matrix but also accepts a completely empty IR program. Sacred and vacuous are independent findings under this brief.
8. The full shared differential family is much larger than the unit. I included only its append and lie fixture members as one partial-family row, timed three times, and explicitly left its other members unknown. All kills and unique kills are bounded observations, not package uniqueness claims.
9. Direct grep for a lowering function name in oracle tests cannot find the call chain: these tests call lowered, which calls Lower. The selected lowering coverage inventory and explicit fixture registrations establish the paths. There are 270 reached lowering functions, listed before planting mutants.
10. A survivor can change an internal optimization recognizer without changing the exercised program's stdout. M09 changes censusCallReturnsUndefined from true to false for append(undefined, undefined). The witness observes that function result directly through a temporary exported adapter. The normal backend checks still agree. This proves unguarded recognition behavior, not a demonstrated wrong program result.
11. The independent recognition adapter is scratch observation code, not a production mutant, test change or oracle edit. Its source and commands are saved separately. It was removed before the restored control.
12. Cold Go coverage and the first switched test build add command time beyond the test binary's own line. Those command timings are recorded separately. Native compilation within the stop row is part of its binary time; I did not invent a separate native rebuild time.
13. Seven opt-in rows outside this unit skipped in the aborted whole baseline. All three requested rows actually ran uncached. The brief is unclear whether installing tools for outside-unit skipped rows is required when a bounded slice is the deliverable; this run did not expand the unit to them.
14. The first coverage-list filter accidentally treated percentages ending in 0.0, including 100.0, as zero. I corrected it to numeric percentage > 0 before writing the fixed mutation plan or running mutants. The saved inventory includes all 270 reached functions.

Timing:

```json
{
  "setup_seconds": 0,
  "nproc": 5,
  "whole_package_baseline_binary_seconds": 90.251,
  "whole_package_baseline_timed_out": true,
  "bounded_baseline_binary_seconds": 0.505,
  "isolated_timing_and_coverage_command_seconds": 46.995381034001184,
  "matrix_and_controls_command_seconds": 94.78372184499676,
  "matrix_and_controls_binary_seconds": 15.325,
  "standalone_vet_seconds": 9.028676337009529,
  "independent_probe_build_seconds": 17.281016199998703,
  "restored_control_command_seconds": 3.6785123000008753,
  "restored_control_exit": 0,
  "npm_reported_seconds": 0.826
}
```

Not covered: package-wide or repository-wide uniqueness; unselected fixtures in the shared differential family; mutations to native or JavaScript backend implementations; a production mutation for the Node-only row. Outside bounded matrix, kills are unknown. No other package tests ran. Standalone diff compilation was checked with go vet ./internal/lower through source overlays.

Outside-unit skipped rows in the whole baseline:

- TestWASIShardPlantedFixture
- TestWASIAgreesWithNode
- TestWASIOracleCatchesMutants
- TestWASIRunnerCatchesMutants
- TestWASIEmission
- TestStage3FixtureHook
- TestEntriesAcceptance
