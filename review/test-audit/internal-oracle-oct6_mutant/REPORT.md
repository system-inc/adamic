u066 starts at 6c60da091afddc9c2fe88b3a1067845b6dc79cb3; all fifteen requested names remain in their listed files.
Full package baseline timed out at 90.088s without a completed test failure; narrowed clean runs passed.
Seven witnesses proved; three bounded sacred rows; five subsumed rows based on one generic Lower-entry mutant.
Four production mutants, three weakened witness checks, and four distinct-entry probes have standalone validated diffs.
Production and harness changes restored; evidence is on test-audit/internal-oracle-oct6_mutant under review/test-audit/internal-oracle-oct6_mutant/.

```json
[
  {
    "test": "TestOct6InheritanceMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/oct6_mutant_test.go",
    "seconds": 0.124,
    "oracle": "Node stdout disagreement detects the two built-in inheritance mutants; witness weakened disagreement.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: oct6_mutant_test.go:73: mutant survived: \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/W1-rows.log 2>&1; oct6_mutant_test.go:73: mutant survived: \"\"",
    "witness_kills": [
      "W1"
    ],
    "subsumption_mutants": null,
    "entry_probe_results": {},
    "witness_check": "disagreement"
  },
  {
    "test": "TestOct6ReleaseMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/oct6_mutant_test.go",
    "seconds": 0.382,
    "oracle": "Node comparison stays clean; LeakSanitizer report text must contain LeakSanitizer or leaked. Witness weakened leakSanitizer.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: oct6_mutant_test.go:116: leak mutant survived:",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=W2 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/W2 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/W2-rows.log 2>&1; oct6_mutant_test.go:116: leak mutant survived:",
    "witness_kills": [
      "W2"
    ],
    "subsumption_mutants": null,
    "entry_probe_results": {},
    "witness_check": "leakSanitizer"
  },
  {
    "test": "TestOmittedArgumentZeroMutantIsCaught",
    "package": "internal/oracle",
    "file": "internal/oracle/omitted_arguments_test.go",
    "seconds": 0.089,
    "oracle": "Node stdout disagreement must report stdout differs for built-in present-zero padding; clean exit/stderr and leaks are preconditions.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: omitted_arguments_test.go:63: want Node to catch zero padding, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/W1-rows.log 2>&1; omitted_arguments_test.go:63: want Node to catch zero padding, got \"\"",
    "witness_kills": [
      "W1"
    ],
    "subsumption_mutants": null,
    "entry_probe_results": {},
    "witness_check": "disagreement"
  },
  {
    "test": "TestOmittedOriginalProbePolicy",
    "package": "internal/oracle",
    "file": "internal/oracle/omitted_arguments_test.go",
    "seconds": 0.117,
    "oracle": "Executes Node for the source and compares native and JavaScript results. Handwritten Node 11 and leak-free expectations also apply.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: omitted_arguments_test.go:82: lower: stage 0 compiles a program from one entry file, got 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestAPromptComesBeforeTheRead"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.093,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=M4 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/M4-rows.log 2>&1; omitted_arguments_test.go:82: lower: stage 0 compiles a program from one entry file, got 1",
    "witness_kills": [],
    "subsumption_mutants": 1,
    "entry_probe_results": {
      "P1": "fail"
    },
    "witness_check": null
  },
  {
    "test": "TestOmittedReaderZeroMutantIsCaught",
    "package": "internal/oracle",
    "file": "internal/oracle/omitted_arguments_test.go",
    "seconds": 0.082,
    "oracle": "Node stdout disagreement must report stdout differs for built-in reader zero padding; clean exit/stderr and leaks are preconditions.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: omitted_arguments_test.go:136: want Node to catch method zero padding, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/W1-rows.log 2>&1; omitted_arguments_test.go:136: want Node to catch method zero padding, got \"\"",
    "witness_kills": [
      "W1"
    ],
    "subsumption_mutants": null,
    "entry_probe_results": {},
    "witness_check": "disagreement"
  },
  {
    "test": "TestNativeAgreesWithNode",
    "package": "internal/oracle",
    "file": "internal/oracle/oracle_test.go",
    "seconds": null,
    "oracle": "Node source versus native release, sanitized, and JavaScript output. Full row also contains self-written refusal and inserted-check expectations; matrix covers twelve lowering fixtures only.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: oracle_test.go:751: Lower: lower: stage 0 compiles a program from one entry file, got 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestAPromptComesBeforeTheRead"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": 0.093,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=M4 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(interleaved\\.a|status_of_stdout\\.a|large_output\\.a|output_then_panic\\.a|write_stdout_order\\.a|write_stderr_order\\.a|prompt_then_read\\.a|killed_after_output\\.a|omitted_scanner\\.a|omitted_reader\\.a|class_oct6_deep\\.a|class_oct6_release\\.a|parameter_properties\\.a|parameter_properties_ownership\\.a)$ > logs/M4-native.log 2>&1; oracle_test.go:751: Lower: lower: stage 0 compiles a program from one entry file, got 1",
    "witness_kills": [],
    "subsumption_mutants": 1,
    "bounded_seconds": 0.407,
    "bounded_subcases": [
      "TestNativeAgreesWithNode/internal/oracle/testdata/class_oct6_deep.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/class_oct6_release.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/interleaved.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/large_output.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/omitted_reader.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/omitted_scanner.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/output_then_panic.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/parameter_properties.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/parameter_properties_ownership.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/prompt_then_read.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/write_stderr_order.a",
      "TestNativeAgreesWithNode/internal/oracle/testdata/write_stdout_order.a"
    ],
    "entry_probe_results": {
      "P1": "fail",
      "P2": "fail"
    },
    "witness_check": null
  },
  {
    "test": "TestTheOracleCatchesOneByte",
    "package": "internal/oracle",
    "file": "internal/oracle/oracle_test.go",
    "seconds": 0.085,
    "oracle": "Node stdout comparison must notice the built-in extra byte; witness weakened disagreement.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: oracle_test.go:809: got \"\", want the mutant caught as \"stdout differs\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/W1-rows.log 2>&1; oracle_test.go:809: got \"\", want the mutant caught as \"stdout differs\"",
    "witness_kills": [
      "W1"
    ],
    "subsumption_mutants": null,
    "entry_probe_results": {},
    "witness_check": "disagreement"
  },
  {
    "test": "TestOneFileHoldsNodesOrder",
    "package": "internal/oracle",
    "file": "internal/oracle/output_test.go",
    "seconds": 0.112,
    "oracle": "Node execution on a shared stdout/stderr file decides byte order and exit status. Handwritten observed-output containment checks prevent a missing case.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2",
      "M4"
    ],
    "unique_kills": [
      "M2"
    ],
    "last_proven_fail": "M2: output_test.go:111: native: exit 0, \"err 0\\nout 0\\nout 1\\nout 2\\nerr 3\\nout 3\\nout 4\\nout 5\\nerr 6\\nout 6\\nout 7\\nout 8\\nerr 9\\nout 9\\nout 10\\nout 11\\nerr 12\\nout 12\\nout 13\\nout 14\\nerr 15\\nout 15\\nout 16\\nout 17\\nerr 18\\nout 18\\nout 19\\nout 20\\nerr 21\\nout 21\\nout 22\\nout 23\\nerr 24\\nout 24\\nout 25\\nout 26\\nerr 27\\nout 27\\nout 28\\nout 29\\nerr 30\\nout 30\\nout 31\\nout 32\\nerr 33\\nout 33\\nout 34\\nout 35\\nerr 36\\nout 36\\nout 37\\nout 38\\nerr 39\\nout 39\\nout 40\\nout 41\\nerr 42\\nout 42\\nout 43\\nout 44\\nerr 45\\nout 45\\nout 46\\nout 47\\nerr 48\\nout 48\\nout 49\\nout 50\\nerr 51\\nout 51\\nout 52\\nout 53\\ne",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2",
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=M2 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/M2 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/M2-rows.log 2>&1; output_test.go:111: native: exit 0, \"err 0\\nout 0\\nout 1\\nout 2\\nerr 3\\nout 3\\nout 4\\nout 5\\nerr 6\\nout 6\\nout 7\\nout 8\\nerr 9\\nout 9\\nout 10\\nout 11\\nerr 12\\nout 12\\nout 13\\nout 14\\nerr 15\\nout 15\\nout 16\\nout 17\\nerr 18\\nout 18\\nout 19\\nout 20\\nerr 21\\nout 21\\nout 22\\nout 23\\nerr 24\\nout 24\\nout 25\\nout 26\\nerr 27\\nout 27\\nout 28\\nout 29\\nerr 30\\nout 30\\nout 31\\nout 32\\nerr 33\\nout 33\\nout 34\\nout 35\\nerr 36\\nout 36\\nout 37\\nout 38\\nerr 39\\nout 39\\nout 40\\nout 41\\nerr 42\\nout 42\\nout 43\\nout 44\\nerr 45\\nout 45\\nout 46\\nout 47\\nerr 48\\nout 48\\nout 49\\nout 50\\nerr 51\\nout 51\\nout 52\\nout 53\\ne",
    "witness_kills": [],
    "subsumption_mutants": null,
    "entry_probe_results": {
      "P2": "fail",
      "P3": "fail"
    },
    "witness_check": null
  },
  {
    "test": "TestClosedStdoutEndsAsOnNode",
    "package": "internal/oracle",
    "file": "internal/oracle/output_test.go",
    "seconds": 0.117,
    "oracle": "Node execution on closed stdout decides exit and stderr; a handwritten Node exit-70 assertion also applies. M1 produces exit 71 with identical stderr.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M4"
    ],
    "unique_kills": [
      "M1"
    ],
    "last_proven_fail": "M1: output_test.go:164: native: exit 71, stderr \"stderr, unbuffered\\n\"; Node: exit 70, stderr \"stderr, unbuffered\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2",
      "P4"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=M1 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/M1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/M1-rows.log 2>&1; output_test.go:164: native: exit 71, stderr \"stderr, unbuffered\\n\"; Node: exit 70, stderr \"stderr, unbuffered\\n\"",
    "witness_kills": [],
    "subsumption_mutants": null,
    "entry_probe_results": {
      "P2": "fail",
      "P4": "fail"
    },
    "witness_check": null
  },
  {
    "test": "TestFileWritesLandInNodesOrder",
    "package": "internal/oracle",
    "file": "internal/oracle/output_test.go",
    "seconds": 0.106,
    "oracle": "Node execution with streams sharing one pipe decides exact output order; handwritten first/second/third lines must also appear.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: output_test.go:213: Lower: lower: stage 0 compiles a program from one entry file, got 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestAPromptComesBeforeTheRead"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.093,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=M4 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/M4-rows.log 2>&1; output_test.go:213: Lower: lower: stage 0 compiles a program from one entry file, got 1",
    "witness_kills": [],
    "subsumption_mutants": 1,
    "entry_probe_results": {
      "P2": "fail",
      "P3": "fail"
    },
    "witness_check": null
  },
  {
    "test": "TestAPromptComesBeforeTheRead",
    "package": "internal/oracle",
    "file": "internal/oracle/output_test.go",
    "seconds": 0.093,
    "oracle": "Executed Node dialogue decides prompt-before-read and output; handwritten ready/got yes guard applies.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: output_test.go:238: Lower: lower: stage 0 compiles a program from one entry file, got 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestFileWritesLandInNodesOrder"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2",
      "P3"
    ],
    "subsumer_seconds": 0.106,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=M4 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/M4-rows.log 2>&1; output_test.go:238: Lower: lower: stage 0 compiles a program from one entry file, got 1",
    "witness_kills": [],
    "subsumption_mutants": 1,
    "entry_probe_results": {
      "P2": "fail",
      "P3": "fail"
    },
    "witness_check": null
  },
  {
    "test": "TestASignalLeavesWhatWasPrinted",
    "package": "internal/oracle",
    "file": "internal/oracle/output_test.go",
    "seconds": 0.177,
    "oracle": "Executed Node protocol decides signal disposition, stdout and empty stderr; CPU-based readiness prevents premature signaling.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M3",
      "M4"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M3: output_test.go:359: program: killed by terminated, stderr \"\", stdout \"started, about to work for a long time\"; Node: killed by terminated, stderr \"\", stdout \"started, about to work for a long time\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2",
      "P4"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/M3-rows.log 2>&1; output_test.go:359: program: killed by terminated, stderr \"\", stdout \"started, about to work for a long time\"; Node: killed by terminated, stderr \"\", stdout \"started, about to work for a long time\\n\"",
    "witness_kills": [],
    "subsumption_mutants": null,
    "entry_probe_results": {
      "P2": "fail",
      "P4": "fail"
    },
    "witness_check": null
  },
  {
    "test": "TestOverloadContractRulings",
    "package": "internal/oracle",
    "file": "internal/oracle/overload_contract_test.go",
    "seconds": 0.182,
    "oracle": "Executed Node outputs are checked against handwritten strings. Lower must produce lower.Refused containing a handwritten overload label. The ruling oracle is self; Node does not prove the static refusal.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: overload_contract_test.go:34: want refusal \"overload 1 of createToken result\", got lower: stage 0 compiles a program from one entry file, got 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestAPromptComesBeforeTheRead"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.093,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=M4 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/M4-rows.log 2>&1; overload_contract_test.go:34: want refusal \"overload 1 of createToken result\", got lower: stage 0 compiles a program from one entry file, got 1",
    "witness_kills": [],
    "subsumption_mutants": 1,
    "entry_probe_results": {
      "P1": "fail"
    },
    "witness_check": null
  },
  {
    "test": "TestParameterPropertyMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/parameter_properties_test.go",
    "seconds": 0.421,
    "oracle": "Node stdout disagreement must catch each built-in missing/late parameter store; exit and sanitizer cleanliness are preconditions.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: parameter_properties_test.go:79: mutant not caught by Node: \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/W1-rows.log 2>&1; parameter_properties_test.go:79: mutant not caught by Node: \"\"",
    "witness_kills": [
      "W1"
    ],
    "subsumption_mutants": null,
    "entry_probe_results": {},
    "witness_check": "disagreement"
  },
  {
    "test": "TestParameterPropertyOwnershipMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/parameter_properties_test.go",
    "seconds": 0.307,
    "oracle": "ASan stderr must contain AddressSanitizer: heap-use-after-free. Exit status alone is not checked. Disabling ASan makes the built-in mutant exit 0 with empty stderr and the witness fail.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W3: parameter_properties_test.go:113: ownership mutant not caught by ASan: exit 0, stderr",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestOct6InheritanceMutants",
      "TestOct6ReleaseMutant",
      "TestOmittedArgumentZeroMutantIsCaught",
      "TestOmittedOriginalProbePolicy",
      "TestOmittedReaderZeroMutantIsCaught",
      "TestNativeAgreesWithNode",
      "TestTheOracleCatchesOneByte",
      "TestOneFileHoldsNodesOrder",
      "TestClosedStdoutEndsAsOnNode",
      "TestFileWritesLandInNodesOrder",
      "TestAPromptComesBeforeTheRead",
      "TestASignalLeavesWhatWasPrinted",
      "TestOverloadContractRulings",
      "TestParameterPropertyMutants",
      "TestParameterPropertyOwnershipMutant"
    ],
    "evidence": "ADAMIC_MUTANT=W3 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/W3 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestOct6InheritanceMutants|TestOct6ReleaseMutant|TestOmittedArgumentZeroMutantIsCaught|TestOmittedOriginalProbePolicy|TestOmittedReaderZeroMutantIsCaught|TestTheOracleCatchesOneByte|TestOneFileHoldsNodesOrder|TestClosedStdoutEndsAsOnNode|TestFileWritesLandInNodesOrder|TestAPromptComesBeforeTheRead|TestASignalLeavesWhatWasPrinted|TestOverloadContractRulings|TestParameterPropertyMutants|TestParameterPropertyOwnershipMutant)$ > logs/W3-rows.log 2>&1; parameter_properties_test.go:113: ownership mutant not caught by ASan: exit 0, stderr",
    "witness_kills": [
      "W3"
    ],
    "subsumption_mutants": null,
    "entry_probe_results": {},
    "witness_check": "ASan enabled"
  }
]
```

| ID | origin/main file:line | change | failed ordinary/witness rows |
|---|---|---|---|
| M1 | internal/native/runtime/adamic.c:83 | `		_exit(70);` to `		_exit(71);` | TestClosedStdoutEndsAsOnNode |
| M2 | internal/native/runtime/adamic.c:177 | `if (stream == adamic_stderr) {` to `if (stream != adamic_stderr) {` | TestOneFileHoldsNodesOrder |
| M3 | internal/native/runtime/adamic.c:113 | `size_t whole = output_whole;` to `size_t whole = output_whole - 1;` | TestASignalLeavesWhatWasPrinted |
| M4 | internal/lower/lower.go:22 | `if len(files) != 1 {` to `if len(files) == 1 {` | TestOmittedOriginalProbePolicy, TestNativeAgreesWithNode, TestOneFileHoldsNodesOrder, TestClosedStdoutEndsAsOnNode, TestFileWritesLandInNodesOrder, TestAPromptComesBeforeTheRead, TestASignalLeavesWhatWasPrinted, TestOverloadContractRulings |
| W1 | internal/oracle/oracle_test.go:716 | `func disagreement(oracle run, native run) string {` to `func disagreement(oracle run, native run) string { 	if oracle.exitCode == oracle.exitCode { return "" }` | TestOct6InheritanceMutants, TestOmittedArgumentZeroMutantIsCaught, TestOmittedReaderZeroMutantIsCaught, TestTheOracleCatchesOneByte, TestParameterPropertyMutants |
| W2 | internal/oracle/oracle_test.go:706 | `func leakSanitizer(t *testing.T, binary string) string {` to `func leakSanitizer(t *testing.T, binary string) string { 	if binary == binary { return "" }` | TestOct6ReleaseMutant |
| W3 | internal/oracle/parameter_properties_test.go:108 | `if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {` to `if err := native.Build(code, binary, native.Options{Sanitize: false}); err != nil {` | TestParameterPropertyOwnershipMutant |

Survivors: none among M1-M4. All three weakened witness checks were caught. Probe passes are neither survivors nor production kills.

Friction, ambiguity, and limits:

- The requested historical commit 8de93800f4 differs from current origin/main 6c60da09. No requested row moved or vanished.
- The package and TestNativeAgreesWithNode alone both exceed the prescribed 90-second budget. Its whole-row median is null; three bounded timings cover twelve actual fixtures. The selector initially named fourteen fixture patterns, but status_of_stdout and killed_after_output are absent from this row's fixture list. Actual matched subcases are explicit in scope.json and results.json.
- Go coverage listed 579 reached lower/native/oracle production functions before the fixed menu. Test-file helper functions were read and listed using rg. Exact runtime C branch coverage was not obtained; the C support inventory is conservative.
- The seven planted-failure tests are witnesses. Their M4 failures are broken preparation and cannot prove the comparison or sanitizer check. Those failures remain in raw matrix results but never count toward their kills or uniqueness. W1, W2, W3 supply their verdicts.
- W1 and W2 return the empty comparison/report value, under the explicit witness-check exception. They are not production mutants. W3 disables the check by changing the sanitizer option, leaving the planted retain mutation and expected report predicate intact. With ASan disabled the observed mutant exits 0 with empty stderr, demonstrating why an exit-only check would miss it.
- Five subsumption findings rest on only M4, which breaks the general one-file entry requirement. This does not measure omitted-slot, overload-proof, or individual fixture sensitivity. Reciprocal FileWrites/Prompt judgments and the other single-mutant findings are hints, not deletion recommendations.
- Runtime mutations use a native-process environment selector; the runtime source/header content remains identical for every selection, so cached libraries execute the selected behavior rather than stale compiled branches. ADAMIC_GATE_UNCACHED=1 forces actual Node/native/protocol observations for every matrix run; ADAMIC_BUILD_CACHE_DIR is separate per selection.
- The nil Lower probe caused a Go panic. Every assigned row was rerun alone, retaining the twelve-fixture restriction for NativeAgrees. Probe results on witnesses are precondition effects and vacuity is null there. Ordinary output rows are judged by their own runtime entries, not by Lower preparation.
- Three-per-row timing uses normal warm result caches, as the brief's commands specify. The first observation is often slower than subsequent cached observations. The matrix bypasses those caches. The full NativeAgrees row was stopped after its first 90.039-second timeout, not repeated twice more.
- M1 is caught on exit status, but the test also compares exact stderr; the mutant preserved stderr. M2 and M3 change actual byte order/content. The overload ruling oracle is a self-written refusal label even though Node runs each program: Node runtime behavior cannot establish the static refusal policy.
- All requested rows ran without skips. Outside-slice skip names from the unfinished full baseline are listed below. No outside-slice SDK or corpus opt-in was installed. Other package rows, unselected NativeAgrees fixtures, repo-wide uniqueness, and exhaustive C branch behavior remain unknown.

Outside-slice observed skips: TestEntriesAcceptance, TestStage3FixtureHook, TestWASIAgreesWithNode, TestWASIEmission, TestWASIOracleCatchesMutants, TestWASIRunnerCatchesMutants, TestWASIShardPlantedFixture.

Warm setup reused; setup.sh not run; nproc=5. npm ci: 0.537s. Clean timing runs: 191.13995869999962 wall seconds, including the 90.039s whole-row timeout. Standalone validation: 2.8523778379994837 wall seconds. Matrix: 199.67929995699978 wall seconds. Per-run and rerun times are in timings.json. The clean switched control was 28.502 wall seconds, including compilation, runtime-library rebuilding, and 20.759 binary seconds; build-only native time was not separately isolated. The native library builds once per flag set because the selector bytes are shared. All final standalone validations passed and all diffs apply to the starting commit. Restored rows passed; restored bounded NativeAgrees passed three times.
