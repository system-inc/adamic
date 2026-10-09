Audited u024 at origin/main 7b18d0576930caca4e22ce2eef92fcf563af52d0.
All 15 requested rows exist in their listed files; no families, witnesses, setup checks or helpers in this slice.
Clean package baseline: 1.225s; 17 package rows, no skips; nproc=5.
20 production mutants: 19 killed, one equivalent candidate; verdicts: 6 sacred, 6 subsumed, 3 overlapping.
Five vacuous rows; production source restored; all 25 diffs apply and pass go vet.

```json
[
  {
    "test": "TestStarExportCollisionNamesBothModules",
    "package": "internal/load",
    "file": "internal/load/export_collision_test.go",
    "seconds": 0.036,
    "oracle": "Exact custom Adamic refusal string; both module names, name and position.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05",
      "M07",
      "M08",
      "M09",
      "M10",
      "M11",
      "M12",
      "M19"
    ],
    "unique_kills": [
      "M10",
      "M11"
    ],
    "last_proven_fail": "M11 --- FAIL: TestStarExportCollisionNamesBothModules (0.11s)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M11.log 2>&1; export_collision_test.go:18: want exact refusal \"/tmp/adamic-gate/TestStarExportCollisionNamesBothModules348022329/001/main.a:2:1: error TS2308: Adamic 0.1 refuses export * collision for 'shared' between './left.a' and './right.a'; explicitly re-export one binding\", got /tmp/adamic-gate/TestStarExportCollisionNamesBothModules348022329/001/main.a:2:1: error TS2308: Adamic 0.1 refuses export * collision for 'shared' between './right.a' and './left.a'; explicitly re-export one binding"
  },
  {
    "test": "TestExplicitExportResolvesStarCollision",
    "package": "internal/load",
    "file": "internal/load/export_collision_test.go",
    "seconds": 0.039,
    "oracle": "Acceptance only; no returned program or resolved binding assertion. Broad configuration errors also kill it.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05",
      "M12",
      "M19"
    ],
    "unique_kills": [],
    "last_proven_fail": "M19 --- FAIL: TestExplicitExportResolvesStarCollision (0.00s)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStarExportCollisionNamesBothModules"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": 0.036,
    "vacuous": true,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M19.log 2>&1; export_collision_test.go:30: load: /tmp/adamic-gate/TestExplicitExportResolvesStarCollision4237300823/001/main.a and /tmp/adamic-gate/TestExplicitExportResolvesStarCollision4237300823/001/main.a.ts both exist, and an import of main.a could mean either; rename one"
  },
  {
    "test": "TestDeclarationsCarryTheirProvenTypes",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.039,
    "oracle": "Handwritten declaration types and positions, including expanded alias; no outside authority cited.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05",
      "M09",
      "M17",
      "M18"
    ],
    "unique_kills": [
      "M18"
    ],
    "last_proven_fail": "M18 --- FAIL: TestDeclarationsCarryTheirProvenTypes (0.06s)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M18.log 2>&1; load_test.go:83: type Shape: got \"Shape\", want the union it stands for"
  },
  {
    "test": "TestATypeErrorFailsTheLoad",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.038,
    "oracle": "tsc 6.0.3 diagnostic TS2322 text and position; independently checked this session in authority-ts2322.log.",
    "oracle_kind": "external-authority",
    "kills": [
      "M03",
      "M05",
      "M07",
      "M08",
      "M09"
    ],
    "unique_kills": [],
    "last_proven_fail": "M09 --- FAIL: TestATypeErrorFailsTheLoad (0.08s)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStarExportCollisionNamesBothModules"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.036,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M09.log 2>&1; load_test.go:94: got [\"/tmp/adamic-gate/TestATypeErrorFailsTheLoad1615825898/001/main.ts:2:7: error TS2322: Type 'string' is not assignable to type 'number'.\"], want one TS2322 at main.ts:1:7 with its message"
  },
  {
    "test": "TestAdamicOptionsAreOn",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.076,
    "oracle": "TypeScript diagnostic codes TS2322/TS2375/TS2584 and documented option behavior; no independent check of these exact three fixtures. Codes only, so a different error with the same code could pass.",
    "oracle_kind": "external-authority",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M05",
      "M07",
      "M08"
    ],
    "unique_kills": [
      "M01",
      "M02"
    ],
    "last_proven_fail": "M02 --- FAIL: TestAdamicOptionsAreOn (0.00s)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M02.log 2>&1; load_test.go:113: Load: want a CheckError, got <nil>"
  },
  {
    "test": "TestEveryFileIsAModule",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.037,
    "oracle": "Acceptance of isolated module scope; no program inspection. Broad configuration errors also kill it.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04",
      "M05",
      "M12",
      "M19"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M04 --- FAIL: TestEveryFileIsAModule (0.10s)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M04.log 2>&1; load_test.go:131: Load: /adamic-prelude/adamic.d.ts:4:15: error TS2451: Cannot redeclare block-scoped variable 'console'."
  },
  {
    "test": "TestAdamicFilesImportEachOther",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.064,
    "oracle": "TypeScript number inference and TS2322 codes/positions for imports; exact fixture values not independently checked. Negative assertions inspect codes and positions, not text.",
    "oracle_kind": "external-authority",
    "kills": [
      "M03",
      "M05",
      "M08",
      "M09",
      "M12",
      "M17",
      "M19"
    ],
    "unique_kills": [],
    "last_proven_fail": "M19 --- FAIL: TestAdamicFilesImportEachOther (0.00s)",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestStarExportCollisionNamesBothModules",
      "TestDeclarationsCarryTheirProvenTypes"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M19.log 2>&1; load_test.go:141: Load: load: /tmp/adamic-gate/TestAdamicFilesImportEachOther4198956552/001/main.a and /tmp/adamic-gate/TestAdamicFilesImportEachOther4198956552/001/main.a.ts both exist, and an import of main.a could mean either; rename one"
  },
  {
    "test": "TestThePreludeDeclaresConsoleAndPanic",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.066,
    "oracle": "Adamic prelude contract: never-return panic and one-string console; negative checks only TS2345/count.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05",
      "M07",
      "M08",
      "M15",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17 --- FAIL: TestThePreludeDeclaresConsoleAndPanic (0.22s)",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestStarExportCollisionNamesBothModules",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestNodeLibraryUsesPinnedDeclarations"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M17.log 2>&1; load_test.go:165: variable age: got \"\", want number (panic returns never)"
  },
  {
    "test": "TestLoadRefusesWhatIsNotAdamic",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.007,
    "oracle": "Handwritten Adamic input refusal substrings, plus Load(nil) rejection.",
    "oracle_kind": "self",
    "kills": [
      "M19"
    ],
    "unique_kills": [],
    "last_proven_fail": "M19 --- FAIL: TestLoadRefusesWhatIsNotAdamic (0.00s)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStarExportCollisionNamesBothModules"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.036,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M19.log 2>&1; load_test.go:193: got <nil>, want an error containing \"both exist\""
  },
  {
    "test": "TestTheProgramsInTheSpecLoad",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.304,
    "oracle": "Test comment cites tsc 6.0.3 acceptance of spec programs; no spec fixture independently rerun through tsc this session. Acceptance only, no program inspection.",
    "oracle_kind": "external-authority",
    "kills": [
      "M03",
      "M05"
    ],
    "unique_kills": [],
    "last_proven_fail": "M05 --- FAIL: TestTheProgramsInTheSpecLoad (0.00s)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStarExportCollisionNamesBothModules"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": 0.036,
    "vacuous": true,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M05.log 2>&1; load_test.go:226: Load: /adamic-prelude/adamic.d.ts:4:15: error TS2451: Cannot redeclare block-scoped variable 'console'."
  },
  {
    "test": "TestOverlayPrecedesDiskForAdamicAliases",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.092,
    "oracle": "Overlay acceptance plus disk TS2322; no returned overlay program inspection.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05",
      "M07",
      "M08",
      "M12",
      "M13",
      "M19"
    ],
    "unique_kills": [
      "M13"
    ],
    "last_proven_fail": "M13 --- FAIL: TestOverlayPrecedesDiskForAdamicAliases (0.00s)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M13.log 2>&1; load_test.go:241: overlay not checked: /tmp/adamic-gate/TestOverlayPrecedesDiskForAdamicAliases.ts2827234625/001/main.ts:1:7: error TS2322: Type 'string' is not assignable to type 'number'."
  },
  {
    "test": "TestHouseStyleOptionsDoNotRejectValidControlFlow",
    "package": "internal/load",
    "file": "internal/load/load_test.go",
    "seconds": 0.08,
    "oracle": "TypeScript acceptance and TS2366 implicit return rule; exact fixtures not independently checked. Negative checks only diagnostic code.",
    "oracle_kind": "external-authority",
    "kills": [
      "M03",
      "M05",
      "M06",
      "M07",
      "M08",
      "M12",
      "M19"
    ],
    "unique_kills": [
      "M06"
    ],
    "last_proven_fail": "M06 --- FAIL: TestHouseStyleOptionsDoNotRejectValidControlFlow (0.11s)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [
      "TestHouseStyleOptionsDoNotRejectValidControlFlow/function_sign(value:_number)_{_if_(value_>_0)_{_return_1;_}_}",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow/function_choose(value:_number):_number_{_let_result=0;_switch(value)_{_case_1:_result+=1;_default:_result+=2;_}_return_result;_}"
    ],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M06.log 2>&1; load_test.go:261: /tmp/adamic-gate/TestHouseStyleOptionsDoNotRejectValidControlFlowfunction_sign(v576655709/001/main.a:1:10: error TS7030: Not all code paths return a value."
  },
  {
    "test": "TestNodeLibraryUsesPinnedDeclarations",
    "package": "internal/load",
    "file": "internal/load/node_library_test.go",
    "seconds": 0.472,
    "oracle": "Official @types/node 25.3.3 fs signatures; checked existsSync boolean at fs.d.ts:3762. Asserts declaration identity, then only TS2322 for boolean rejection.",
    "oracle_kind": "external-authority",
    "kills": [
      "M03",
      "M07",
      "M08",
      "M12",
      "M14",
      "M15",
      "M16",
      "M19"
    ],
    "unique_kills": [],
    "last_proven_fail": "M19 --- FAIL: TestNodeLibraryUsesPinnedDeclarations (0.00s)",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestStarExportCollisionNamesBothModules"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01",
      "P05"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M19.log 2>&1; node_library_test.go:23: load: /tmp/adamic-gate/TestNodeLibraryUsesPinnedDeclarations1919771430/001/main.a and /tmp/adamic-gate/TestNodeLibraryUsesPinnedDeclarations1919771430/001/main.a.ts both exist, and an import of main.a could mean either; rename one"
  },
  {
    "test": "TestNodeLibraryRejectsDifferentVersion",
    "package": "internal/load",
    "file": "internal/load/node_library_test.go",
    "seconds": 0.005,
    "oracle": "Repository-owned pin 25.3.3; only substring rejection, so another error mentioning the pin could pass.",
    "oracle_kind": "self",
    "kills": [
      "M14"
    ],
    "unique_kills": [],
    "last_proven_fail": "M14 --- FAIL: TestNodeLibraryRejectsDifferentVersion (0.00s)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNodeLibraryKeepsOfficialConsoleSignatures"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P03"
    ],
    "subsumer_seconds": 0.248,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M14.log 2>&1; node_library_test.go:55: want pinned version refusal, got <nil>"
  },
  {
    "test": "TestNodeLibraryKeepsOfficialConsoleSignatures",
    "package": "internal/load",
    "file": "internal/load/node_library_test.go",
    "seconds": 0.248,
    "oracle": "Official @types/node 25.3.3 Console.log(...data:any[]):void at console.d.ts:107 checked this session. Acceptance only, no program inspection.",
    "oracle_kind": "external-authority",
    "kills": [
      "M03",
      "M12",
      "M14",
      "M15",
      "M16",
      "M19"
    ],
    "unique_kills": [],
    "last_proven_fail": "M19 --- FAIL: TestNodeLibraryKeepsOfficialConsoleSignatures (0.00s)",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNodeLibraryUsesPinnedDeclarations"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": 0.472,
    "vacuous": true,
    "vacuous_subcases": [],
    "bounded": false,
    "matrix_rows": [
      "TestStarExportCollisionNamesBothModules",
      "TestExplicitExportResolvesStarCollision",
      "TestDeclarationsCarryTheirProvenTypes",
      "TestATypeErrorFailsTheLoad",
      "TestAdamicOptionsAreOn",
      "TestEveryFileIsAModule",
      "TestAdamicFilesImportEachOther",
      "TestThePreludeDeclaresConsoleAndPanic",
      "TestLoadRefusesWhatIsNotAdamic",
      "TestTheProgramsInTheSpecLoad",
      "TestOverlayPrecedesDiskForAdamicAliases",
      "TestHouseStyleOptionsDoNotRejectValidControlFlow",
      "TestNodeLibraryUsesPinnedDeclarations",
      "TestNodeLibraryRejectsDifferentVersion",
      "TestNodeLibraryKeepsOfficialConsoleSignatures",
      "TestTypeScriptIsThePinnedCommit",
      "TestRegExpCaptureTypes"
    ],
    "evidence": "ADAMIC_MUTANT=M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M19.log 2>&1; node_library_test.go:62: load: /tmp/adamic-gate/TestNodeLibraryKeepsOfficialConsoleSignatures3055616020/001/main.a and /tmp/adamic-gate/TestNodeLibraryKeepsOfficialConsoleSignatures3055616020/001/main.a.ts both exist, and an import of main.a could mean either; rename one"
  }
]
```

Mutant and probe table. Every coordinate is at the starting commit.
| ID | File:line | Operator and change | Failed package rows |
|---|---|---|---|
| M01 | internal/load/load.go:60 | change constant or option: NoUncheckedIndexedAccess:   core.TSTrue -> NoUncheckedIndexedAccess:   core.TSFalse | TestAdamicOptionsAreOn |
| M02 | internal/load/load.go:61 | change constant or option: ExactOptionalPropertyTypes: core.TSTrue -> ExactOptionalPropertyTypes: core.TSFalse | TestAdamicOptionsAreOn |
| M03 | internal/load/load.go:59 | change constant or option: Strict:                     core.TSTrue -> Strict:                     core.TSFalse | TestATypeErrorFailsTheLoad, TestAdamicFilesImportEachOther, TestAdamicOptionsAreOn, TestDeclarationsCarryTheirProvenTypes, TestEveryFileIsAModule, TestExplicitExportResolvesStarCollision, TestHouseStyleOptionsDoNotRejectValidControlFlow, TestNodeLibraryKeepsOfficialConsoleSignatures, TestNodeLibraryUsesPinnedDeclarations, TestOverlayPrecedesDiskForAdamicAliases, TestStarExportCollisionNamesBothModules, TestThePreludeDeclaresConsoleAndPanic, TestTheProgramsInTheSpecLoad |
| M04 | internal/load/load.go:69 | change constant or option: core.ModuleDetectionKindForce -> core.ModuleDetectionKindAuto | TestEveryFileIsAModule |
| M05 | internal/load/load.go:72 | change constant or option: []string{"lib.es2024.d.ts"} -> []string{"lib.es2024.d.ts", "lib.dom.d.ts"} | TestATypeErrorFailsTheLoad, TestAdamicFilesImportEachOther, TestAdamicOptionsAreOn, TestDeclarationsCarryTheirProvenTypes, TestEveryFileIsAModule, TestExplicitExportResolvesStarCollision, TestHouseStyleOptionsDoNotRejectValidControlFlow, TestOverlayPrecedesDiskForAdamicAliases, TestStarExportCollisionNamesBothModules, TestThePreludeDeclaresConsoleAndPanic, TestTheProgramsInTheSpecLoad |
| M06 | internal/load/load.go:65 | change constant or option: NoEmit:                     core.TSTrue, -> NoEmit:                     core.TSTrue, 		NoImplicitReturns: core.TSTrue, | TestHouseStyleOptionsDoNotRejectValidControlFlow |
| M07 | internal/load/load.go:147 | off-by-one bound: len(diagnostics) > 0 -> len(diagnostics) > 1 | TestATypeErrorFailsTheLoad, TestAdamicOptionsAreOn, TestHouseStyleOptionsDoNotRejectValidControlFlow, TestNodeLibraryUsesPinnedDeclarations, TestOverlayPrecedesDiskForAdamicAliases, TestRegExpCaptureTypes, TestStarExportCollisionNamesBothModules, TestThePreludeDeclaresConsoleAndPanic |
| M08 | internal/load/load.go:233 | drop statement: all = append(all, p.compiler.GetSemanticDiagnostics(ctx, nil)...) ->  | TestATypeErrorFailsTheLoad, TestAdamicFilesImportEachOther, TestAdamicOptionsAreOn, TestHouseStyleOptionsDoNotRejectValidControlFlow, TestNodeLibraryUsesPinnedDeclarations, TestOverlayPrecedesDiskForAdamicAliases, TestRegExpCaptureTypes, TestStarExportCollisionNamesBothModules, TestThePreludeDeclaresConsoleAndPanic |
| M09 | internal/load/load.go:278 | off-by-one: return line + 1, len(utf16.Encode([]rune(prefix))) + 1 -> return line + 2, len(utf16.Encode([]rune(prefix))) + 1 | TestATypeErrorFailsTheLoad, TestAdamicFilesImportEachOther, TestDeclarationsCarryTheirProvenTypes, TestStarExportCollisionNamesBothModules |
| M10 | internal/load/export_collision.go:13 | change constant or option: diagnostic.Code() != 2308 -> diagnostic.Code() != 2309 | TestStarExportCollisionNamesBothModules |
| M11 | internal/load/export_collision.go:25 | swap two arguments: arguments[1], strings.Trim(arguments[0], "'\""), declaration.ModuleSpecifier.Text() -> arguments[1], declaration.ModuleSpecifier.Text(), strings.Trim(arguments[0], "'\"") | TestStarExportCollisionNamesBothModules |
| M12 | internal/load/source_fs.go:40 | change constant or option: strings.HasSuffix(path.AsString(), ".a.ts") -> strings.HasSuffix(path.AsString(), ".b.ts") | TestAdamicFilesImportEachOther, TestEveryFileIsAModule, TestExplicitExportResolvesStarCollision, TestHouseStyleOptionsDoNotRejectValidControlFlow, TestNodeLibraryKeepsOfficialConsoleSignatures, TestNodeLibraryUsesPinnedDeclarations, TestOverlayPrecedesDiskForAdamicAliases, TestStarExportCollisionNamesBothModules |
| M13 | internal/load/load.go:102 | drop whole loop to avoid unused variables: for name, source := range overlay { 		normalizedOverlay[currentDirectory.ResolveFile(name)] = source 	} ->  | TestOverlayPrecedesDiskForAdamicAliases |
| M14 | internal/load/node_library.go:31 | flip condition: pin.Version != NodeTypesVersion -> pin.Version == NodeTypesVersion | TestNodeLibraryKeepsOfficialConsoleSignatures, TestNodeLibraryRejectsDifferentVersion, TestNodeLibraryUsesPinnedDeclarations |
| M15 | internal/load/load.go:130 | flip condition: usesNodeModules(program) -> !usesNodeModules(program) | TestNodeLibraryKeepsOfficialConsoleSignatures, TestNodeLibraryUsesPinnedDeclarations, TestThePreludeDeclaresConsoleAndPanic |
| M16 | internal/load/node_library.go:78 | return early: func nodePrelude() string { -> func nodePrelude() string { 	return prelude | TestNodeLibraryKeepsOfficialConsoleSignatures, TestNodeLibraryUsesPinnedDeclarations |
| M17 | internal/load/declarations.go:54 | drop statement: node.ForEachChild(visit) ->  | TestAdamicFilesImportEachOther, TestDeclarationsCarryTheirProvenTypes, TestThePreludeDeclaresConsoleAndPanic |
| M18 | internal/load/declarations.go:65 | flip condition: node.Kind == ast.KindTypeAliasDeclaration -> node.Kind != ast.KindTypeAliasDeclaration | TestDeclarationsCarryTheirProvenTypes |
| M19 | internal/load/load.go:214 | flip condition: fs.FS.FileExists(fileName.AppendSuffix(".ts")) -> !fs.FS.FileExists(fileName.AppendSuffix(".ts")) | TestAdamicFilesImportEachOther, TestEveryFileIsAModule, TestExplicitExportResolvesStarCollision, TestHouseStyleOptionsDoNotRejectValidControlFlow, TestLoadRefusesWhatIsNotAdamic, TestNodeLibraryKeepsOfficialConsoleSignatures, TestNodeLibraryUsesPinnedDeclarations, TestOverlayPrecedesDiskForAdamicAliases, TestRegExpCaptureTypes, TestStarExportCollisionNamesBothModules |
| M20 | internal/load/node_library.go:73 | change constant or option: strings.HasSuffix(file.FileName().AsString(), ".d.ts") -> strings.HasSuffix(file.FileName().AsString(), ".ts") |  |
| P01 | internal/load/load.go:80 | empty-answer probe: func Load(paths []string) (*Program, error) { -> func Load(paths []string) (*Program, error) { 	return nil, nil | TestATypeErrorFailsTheLoad, TestAdamicFilesImportEachOther, TestAdamicOptionsAreOn, TestDeclarationsCarryTheirProvenTypes, TestHouseStyleOptionsDoNotRejectValidControlFlow, TestLoadRefusesWhatIsNotAdamic, TestNodeLibraryUsesPinnedDeclarations, TestOverlayPrecedesDiskForAdamicAliases, TestRegExpCaptureTypes, TestStarExportCollisionNamesBothModules, TestThePreludeDeclaresConsoleAndPanic |
| P02 | internal/load/load.go:87 | empty-answer probe: func LoadOverlay(paths []string, overlay map[string]string) (*Program, error) { -> func LoadOverlay(paths []string, overlay map[string]string) (*Program, error) { 	return nil, nil |  |
| P03 | internal/load/node_library.go:19 | empty-answer probe: func nodeTypesIndex(directory string) (string, error) { -> func nodeTypesIndex(directory string) (string, error) { 	return "", nil | TestNodeLibraryKeepsOfficialConsoleSignatures, TestNodeLibraryRejectsDifferentVersion, TestNodeLibraryUsesPinnedDeclarations |
| P04 | internal/load/declarations.go:43 | empty-answer probe: whole function body -> return nil | TestAdamicFilesImportEachOther, TestDeclarationsCarryTheirProvenTypes, TestThePreludeDeclaresConsoleAndPanic |
| P05 | internal/load/node_library.go:72 | empty-answer probe: whole function body -> return false | TestNodeLibraryUsesPinnedDeclarations |

Survivor M20: equivalent candidate. Changing the final suffix predicate from .d.ts to .ts leaves every loaded pinned declaration accepted. No changed-output witness was obtained; not classified as unguarded.
Probe survivors do not count as production survivors, kills, uniqueness or subsumption. P02 passed every package row. P01 individually proved the five Load-only acceptance rows or irrelevant entries pass while the relevant negative/type-output checks fail. Probe judgments use the row's own entry: Declarations for the declaration row, LoadOverlay for overlay, nodeTypesIndex for version rejection, Load elsewhere, with additional IsNodeLibrary probe for pinned declarations.
The house-control-flow positive implicit-return and switch-fallthrough subcases passed P01 while its TS2366 negative failed. They are in vacuous_subcases.

Commands:
- Baseline/list: source /workspace/adamic-tools/env.sh; go test -list . ./internal/load/; timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > LOG 2>&1.
- Mutants: ADAMIC_MUTANT=ID timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > ID.log 2>&1. Switched scratch source uses one build; final repository has no switch.
- Panic recovery: ADAMIC_MUTANT=P01 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run ^ROW$ > P01-ROW.log 2>&1, all 17 rows.
- Cost: timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run ^ROW$ > ROW-timing-N.log 2>&1, three runs each; package ok line seconds, not CLI wall time or parallel-parent PASS time.
- Standalone validation: go vet ./internal/load/ with each standalone mutant, then restore; git apply --check EACH.diff on restored starting commit.
- Authority: node stage3/api/node_modules/typescript/bin/tsc --ignoreConfig --noEmit --strict --lib es2024 /tmp/u024-authority.ts; source const broken: number = 'x'; log confirms TS2322 at 1,7. Official declarations independently inspected with rg, saved in authority-node.log.

Unclear instructions and costs:
- The reference commit 8de93800f4 differs from fetched origin/main. Followed the explicit fresh-origin/main instruction, recording 7b18d057. No requested rows moved or vanished.
- env.sh provided working tools but the required cohere/TypeScript source was absent. The tools-only warm check did not establish repository build readiness. npm ci stage3/api succeeded in 0.766s. Prescribed setup restored submodules in 20.329s but broad cache warming exceeded 90s and was stopped, around 115s total observed elapsed. Narrowed to internal/load; initial discovery build was stopped after 97s and retried using cached dependencies. The retry finished within its hard 90s limit. These are build costs, not a red test baseline.
- Three mutants per each of 15 rows asks for about 45 mutants, conflicting with the 20-mutant cap. Honored 20, spread across production functions; each subsumption rests only on its listed 1 to 6 kills. No deletion recommendation.
- Entries were interpreted by the answer under test, rather than every internal preparation helper or filesystem interface method. Added direct Declarations and IsNodeLibrary probes after the initial three entry probes; five probes total. They are not production mutants.
- M06 changes a default-false compiler option by adding its explicit true field. This is the permitted option-change operator, not a supplemental inserted executable statement.
- Early-return standalone diffs must remove their unreachable tails to pass go vet. M16 initially failed unreachable-code vet. P03 initially failed unused-import vet; removed its now unused imports. Only corrected compiling diffs count.
- Strict=false conflicts with exactOptionalPropertyTypes=true, yielding TS5052. DOM library addition conflicts with prelude console. Their catches demonstrate rejection sensitivity, not specifically correct positive-output interpretation. Assertions that inspect only acceptance, diagnostic codes, counts or substrings can pass for the wrong reason; descriptions record that weakness.
- The brief's package uniqueness rule and big-slice bounded exception are consistent here: every production mutant ran all 17 package rows within budget, so package uniqueness is observed, bounded=false. Repo-wide uniqueness remains unknown.
- An empty nil Program makes consumers panic. Reran every package row individually, never inferred results for rows after an aborted binary.
- Function inventory is a static superset of functions on these paths, including sourceFS interface capabilities. It is not a dynamic coverage proof of every callback.

Timing and uncovered work:
- Main switched test build: 4.110s. Final standalone vet passes total 2.689s; failed attempts and earlier repeated vet batches are additional unmeasured wall costs. Additional two probes took 7.382s including their build and runs.
- Initial main matrix/probes test-binary time totals 29.922s; excludes P01 individual recovery and additional P04/P05 runs. All 45 isolated clean binary runs total 4.800s. Full session about 12 minutes, dominated early by cold dependencies and later by Go command overhead. No matrix binary exceeded 90s.
- No native products or ports are involved, so native rebuild/cache isolation is inapplicable. No other package tests, repo-wide replay, native behavior, or independent authority checks for every handwritten value were run.
- setup.log, baseline/list logs, per-mutant full JSON logs, isolated panic logs, timing logs, authority logs, validation.json, timings.json, matrix.json and results.json are preserved. runner.py, finish.py and report.py record session automation; runner.py assumes the starting repository and warm environment.
