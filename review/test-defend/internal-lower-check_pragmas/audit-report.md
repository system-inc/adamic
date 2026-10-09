Starting commit: 8171b3173bdbfce1f7982d3c4f731279307ece37; all 13 requested names present in the stated files.
Verdicts: seven sacred, four subsumed, one overlapping, one untrue, based on 18 production mutants.
Empty-answer result: pragma-neighbor row is vacuous; readonly positive lowering subcase is vacuous.
Survivors: M15 drops a runtime static side effect; M17 changes the diagnostic label. Both have changed-output witnesses.
Evidence: test-audit/internal-lower-check_pragmas under review/test-audit/internal-lower-check_pragmas/; production restored.

CODE UNDER TEST was declared before mutation: Adamic lowering (pragma refusals, class layout/static initialization, accessor guards, cycle traversal), plus Adamic's loader diagnostic adapter for direct checker rows. No upstream tsgo checker or testing oracle was mutated. The complete production function list is reached-functions.txt: 398 functions across 88 files reached by the 13 rows, measured before mutation with coverage.out. functions-coverage.txt includes zero-hit functions too. Coverage is over-approximate at function granularity, not proof of every branch.

ORACLE was declared before mutation: self-written IR/refusal assertions; tsgo diagnostic fragments for readonly and private checks. Those two fragments were checked in-session against pinned upstream TS2540 and TS18013 definitions, recorded in authority-check.txt. Tsgo runs in process; no row in this unit performs a separate Node differential comparison. The Node witnesses are audit evidence, not existing test oracles.

The 13 bodies differ in assertions, so they remain distinct rows despite their shared lowerSource preparation helper. No family, helper, witness, or setup-check row was identified. All requested names appear in list-retry.log; none moved from the two named files or vanished. Starting main is newer than the brief's 8de93800f4 reference.

Rows below use the complete 13-row slice for subsumption. Whole-package mutant runs were attempted because the clean baseline fit the budget. unique_kills are restricted to completed whole-package runs with exactly one failed row. M14 and M18 aborted with Go panics: every selected row was rerun alone; outside rows remain unknown for those mutants. P1 and P2 also aborted and received the same isolated reruns. All rows are marked bounded because the combined matrix contains those restricted runs. matrix.json lists the exact observed rows per mutant and the full-package failures where known. Optional skipped rows remain unknown, and repository-wide uniqueness is for central replay.

```json
[
  {
    "test": "TestCheckPragmasAreRefused",
    "package": "internal/lower",
    "file": "internal/lower/check_pragmas_test.go:13",
    "seconds": 0.179,
    "oracle": "Self-written Refused type, source location, pragma name, removal instruction and fix fragments.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03"
    ],
    "unique_kills": [
      "M01",
      "M02"
    ],
    "last_proven_fail": "M03: check_pragmas_test.go:22: got <nil>, want pragma refusal",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M03 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M03 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > review/test-audit/internal-lower-check_pragmas/M03.log 2>&1; check_pragmas_test.go:22: got <nil>, want pragma refusal",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestCheckPragmaNeighborsCompile",
    "package": "internal/lower",
    "file": "internal/lower/check_pragmas_test.go:41",
    "seconds": 0.146,
    "oracle": "Self-written expectation of nil error only; never inspects the returned program. Empty Lower passes.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "No production mutant killed this row. P1 isolated run passed; see P1-TestCheckPragmaNeighborsCompile.log",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestTsgoHonorsNoCheckInAdamicFiles",
    "package": "internal/lower",
    "file": "internal/lower/check_pragmas_test.go:57",
    "seconds": 0.095,
    "oracle": "Self-written acceptance/rejection expectations for in-process tsgo through Load, plus Adamic Refused type. First loader rejection accepts any error.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M16"
    ],
    "unique_kills": [],
    "last_proven_fail": "M16: check_pragmas_test.go:65: checker accepted boolean = 2 without pragma",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestCheckPragmasAreRefused",
      "TestClassFeaturesPrivateChecker"
    ],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > review/test-audit/internal-lower-check_pragmas/M16.log 2>&1; check_pragmas_test.go:65: checker accepted boolean = 2 without pragma",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestClassFeaturesReadonlyChecker",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:14",
    "seconds": 0.187,
    "oracle": "tsgo TS2540 read-only diagnostic fragment, checked against pinned upstream definition this session; nil-error expectation for positive lowering subcase. Does not assert exact diagnostic code or positive IR.",
    "oracle_kind": [
      "external-authority",
      "self"
    ],
    "kills": [
      "M16"
    ],
    "unique_kills": [],
    "last_proven_fail": "M16: class_features_test.go:29: want tsc readonly diagnostic, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestClassFeaturesPrivateChecker"
    ],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 0.034,
    "vacuous": false,
    "vacuous_subcases": [
      "positive readonly/mutable-content lowering subcase passes P1; Lower entry is vacuous"
    ],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > review/test-audit/internal-lower-check_pragmas/M16.log 2>&1; class_features_test.go:29: want tsc readonly diagnostic, got <nil>",
    "subsumption_mutant_count": 1
  },
  {
    "test": "TestClassFeaturesPrivateChecker",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:38",
    "seconds": 0.034,
    "oracle": "tsgo TS18013 private identifier diagnostic fragment, checked against pinned upstream definition this session. Does not assert exact code.",
    "oracle_kind": "external-authority",
    "kills": [
      "M16"
    ],
    "unique_kills": [],
    "last_proven_fail": "M16: class_features_test.go:46: want tsc private scope diagnostic, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestTsgoHonorsNoCheckInAdamicFiles"
    ],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 0.095,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M16 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > review/test-audit/internal-lower-check_pragmas/M16.log 2>&1; class_features_test.go:46: want tsc private scope diagnostic, got <nil>",
    "subsumption_mutant_count": 1
  },
  {
    "test": "TestClassFeaturesPrivateStorage",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:50",
    "seconds": 0.04,
    "oracle": "Self-written private-field count >= 2. Counts visibility metadata, not field identity or native privacy.",
    "oracle_kind": "self",
    "kills": [
      "M04",
      "M18"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M18: class_features_test.go:69: private storage lost its visibility metadata: 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M18 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestClassFeaturesPrivateStorage$ > review/test-audit/internal-lower-check_pragmas/M18-TestClassFeaturesPrivateStorage.log 2>&1; class_features_test.go:69: private storage lost its visibility metadata: 1",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestClassFeaturesAccessorRefusals",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:73",
    "seconds": 0.118,
    "oracle": "Self-written accessor override and getter may throw diagnostic fragments; no exact code/location.",
    "oracle_kind": "self",
    "kills": [
      "M06",
      "M07",
      "M09",
      "M18"
    ],
    "unique_kills": [
      "M07"
    ],
    "last_proven_fail": "M18: class_features_test.go:87: want throwing spread diagnosed, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M18 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestClassFeaturesAccessorRefusals$ > review/test-audit/internal-lower-check_pragmas/M18-TestClassFeaturesAccessorRefusals.log 2>&1; class_features_test.go:87: want throwing spread diagnosed, got <nil>",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestClassFeaturesStaticDeclarationsExecute",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:91",
    "seconds": 0.067,
    "oracle": "Self-written Classes nonempty, first class Static flag, Main length >= 2. Does not execute; survives M15 removal of initializer calls.",
    "oracle_kind": "self",
    "kills": [
      "M05",
      "M14",
      "M18"
    ],
    "unique_kills": [
      "M05"
    ],
    "last_proven_fail": "M18: class_features_test.go:102: static initialization was dropped",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M18 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestClassFeaturesStaticDeclarationsExecute$ > review/test-audit/internal-lower-check_pragmas/M18-TestClassFeaturesStaticDeclarationsExecute.log 2>&1; class_features_test.go:102: static initialization was dropped",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestClassFeaturesStaticSoundness",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:107",
    "seconds": 0.304,
    "oracle": "Self-written non-nil-error expectation only for ten unsafe sources; any unrelated lowering error also satisfies it.",
    "oracle_kind": "self",
    "kills": [
      "M06",
      "M08",
      "M10",
      "M18"
    ],
    "unique_kills": [
      "M08"
    ],
    "last_proven_fail": "M18: class_features_test.go:123: unsafe static program accepted: class Box { static first = this.read(); static later = 'ready'; static read(): string { return this.later; } } console.log(Box.first);",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M18 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestClassFeaturesStaticSoundness$ > review/test-audit/internal-lower-check_pragmas/M18-TestClassFeaturesStaticSoundness.log 2>&1; class_features_test.go:123: unsafe static program accepted: class Box { static first = this.read(); static later = 'ready'; static read(): string { return this.later; } } console.log(Box.first);",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestClassFeaturesAccessorCaptureCycle",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:128",
    "seconds": 0.034,
    "oracle": "Self-written cycle diagnostic fragment; no exact location/code.",
    "oracle_kind": "self",
    "kills": [
      "M09",
      "M10",
      "M11"
    ],
    "unique_kills": [
      "M11"
    ],
    "last_proven_fail": "M11: class_features_test.go:137: want captured-cell cycle refused, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M11 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > review/test-audit/internal-lower-check_pragmas/M11.log 2>&1; class_features_test.go:137: want captured-cell cycle refused, got <nil>",
    "subsumption_mutant_count": null
  },
  {
    "test": "TestClassFeaturesNarrowedAccessor",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:141",
    "seconds": 0.037,
    "oracle": "Self-written narrowed accessor reread fragment. M09 still passes when refusal moves from column 207 to the earlier read at 167; witness logs prove wrong-read rejection.",
    "oracle_kind": "self",
    "kills": [
      "M18"
    ],
    "unique_kills": [],
    "last_proven_fail": "M18: class_features_test.go:145: want changing getter reread refused, got <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.036,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M18 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestClassFeaturesNarrowedAccessor$ > review/test-audit/internal-lower-check_pragmas/M18-TestClassFeaturesNarrowedAccessor.log 2>&1; class_features_test.go:145: want changing getter reread refused, got <nil>",
    "subsumption_mutant_count": 1
  },
  {
    "test": "TestClassFeaturesStaticParentCycle",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:149",
    "seconds": 0.036,
    "oracle": "Self-written cycle diagnostic fragment; no exact location/code.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M12",
      "M14",
      "M18"
    ],
    "unique_kills": [],
    "last_proven_fail": "M18: class_features_test.go:153: constructor parent cycle accepted: <nil>",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.036,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M18 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestClassFeaturesStaticParentCycle$ > review/test-audit/internal-lower-check_pragmas/M18-TestClassFeaturesStaticParentCycle.log 2>&1; class_features_test.go:153: constructor parent cycle accepted: <nil>",
    "subsumption_mutant_count": 4
  },
  {
    "test": "TestClassFeaturesStaticInterfaceCycle",
    "package": "internal/lower",
    "file": "internal/lower/class_features_test.go:157",
    "seconds": 0.036,
    "oracle": "Self-written cycle diagnostic fragment; no exact location/code.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M12",
      "M13",
      "M14",
      "M18"
    ],
    "unique_kills": [
      "M13"
    ],
    "last_proven_fail": "M18: class_features_test.go:161: constructor interface cycle accepted: <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestCheckPragmasAreRefused",
      "TestCheckPragmaNeighborsCompile",
      "TestTsgoHonorsNoCheckInAdamicFiles",
      "TestClassFeaturesReadonlyChecker",
      "TestClassFeaturesPrivateChecker",
      "TestClassFeaturesPrivateStorage",
      "TestClassFeaturesAccessorRefusals",
      "TestClassFeaturesStaticDeclarationsExecute",
      "TestClassFeaturesStaticSoundness",
      "TestClassFeaturesAccessorCaptureCycle",
      "TestClassFeaturesNarrowedAccessor",
      "TestClassFeaturesStaticParentCycle",
      "TestClassFeaturesStaticInterfaceCycle"
    ],
    "evidence": "ADAMIC_MUTANT=M18 ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run ^TestClassFeaturesStaticInterfaceCycle$ > review/test-audit/internal-lower-check_pragmas/M18-TestClassFeaturesStaticInterfaceCycle.log 2>&1; class_features_test.go:161: constructor interface cycle accepted: <nil>",
    "subsumption_mutant_count": null
  }
]
```

| ID | origin/main file:line | Change | Failed selected rows |
| --- | --- | --- | --- |
| M01 | internal/lower/refusals.go:54 | flip ts-check condition | TestCheckPragmasAreRefused |
| M02 | internal/lower/refusals.go:56 | off-by-one pragma line | TestCheckPragmasAreRefused |
| M03 | internal/lower/refusals.go:53 | drop whole pragma loop | TestCheckPragmasAreRefused, TestTsgoHonorsNoCheckInAdamicFiles |
| M04 | internal/lower/class_inheritance.go:253 | change private metadata constant | TestClassFeaturesPrivateStorage |
| M05 | internal/lower/class_static.go:90 | change static metadata constant | TestClassFeaturesStaticDeclarationsExecute |
| M06 | internal/lower/class_inheritance.go:48 | return early: nil | TestClassFeaturesAccessorRefusals, TestClassFeaturesStaticSoundness |
| M07 | internal/lower/class_accessors.go:367 | return early: nil | TestClassFeaturesAccessorRefusals |
| M08 | internal/lower/class_static.go:347 | return early: nil | TestClassFeaturesStaticSoundness |
| M09 | internal/lower/object.go:508 | flip narrowed accessor condition | TestClassFeaturesAccessorCaptureCycle, TestClassFeaturesAccessorRefusals |
| M10 | internal/lower/cycles.go:63 | return early: nil | TestClassFeaturesAccessorCaptureCycle, TestClassFeaturesStaticInterfaceCycle, TestClassFeaturesStaticParentCycle, TestClassFeaturesStaticSoundness |
| M11 | internal/lower/cycles.go:426 | drop accessor capture traversal loop | TestClassFeaturesAccessorCaptureCycle |
| M12 | internal/lower/cycles.go:421 | drop static parent traversal block | TestClassFeaturesStaticInterfaceCycle, TestClassFeaturesStaticParentCycle |
| M13 | internal/lower/cycles.go:398 | drop construct-signature traversal block | TestClassFeaturesStaticInterfaceCycle |
| M14 | internal/lower/class_static.go:26 | return early: false | TestClassFeaturesStaticDeclarationsExecute, TestClassFeaturesStaticParentCycle, TestClassFeaturesStaticInterfaceCycle |
| M15 | internal/lower/class_static.go:331 | drop static initializer call statement |  |
| M16 | internal/load/load.go:233 | drop semantic diagnostic collection | TestClassFeaturesPrivateChecker, TestClassFeaturesReadonlyChecker, TestTsgoHonorsNoCheckInAdamicFiles |
| M17 | internal/load/load.go:255 | change diagnostic label constant |  |
| M18 | internal/lower/lower.go:59 | drop whole module-body lowering loop to avoid unused body local | TestClassFeaturesPrivateStorage, TestClassFeaturesAccessorRefusals, TestClassFeaturesStaticDeclarationsExecute, TestClassFeaturesStaticSoundness, TestClassFeaturesNarrowedAccessor, TestClassFeaturesStaticParentCycle, TestClassFeaturesStaticInterfaceCycle |
| P1 | internal/lower/lower.go:20 | return early: nil, nil | TestCheckPragmasAreRefused, TestTsgoHonorsNoCheckInAdamicFiles, TestClassFeaturesPrivateStorage, TestClassFeaturesAccessorRefusals, TestClassFeaturesStaticDeclarationsExecute, TestClassFeaturesStaticSoundness, TestClassFeaturesAccessorCaptureCycle, TestClassFeaturesNarrowedAccessor, TestClassFeaturesStaticParentCycle, TestClassFeaturesStaticInterfaceCycle |
| P2 | internal/load/load.go:80 | return early: nil, nil | TestCheckPragmasAreRefused, TestCheckPragmaNeighborsCompile, TestTsgoHonorsNoCheckInAdamicFiles, TestClassFeaturesReadonlyChecker, TestClassFeaturesPrivateChecker, TestClassFeaturesPrivateStorage, TestClassFeaturesAccessorRefusals, TestClassFeaturesStaticDeclarationsExecute, TestClassFeaturesStaticSoundness, TestClassFeaturesAccessorCaptureCycle, TestClassFeaturesNarrowedAccessor, TestClassFeaturesStaticParentCycle, TestClassFeaturesStaticInterfaceCycle |

Full-package failed rows, including outside-unit catches, are in matrix.json and raw logs. P1/P2 are empty-answer probes, excluded from kills, unique kills and all verdict/subsumption sets. No supplemental insertion mutant was used. The switched build was compiled once; all standalone diffs independently applied to the original Git index and passed go vet with a Go overlay that restored all other scratch files to origin/main. Every compiler mutant used ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/<id>. The native tests call native.Build directly and no shared ADAMIC cache directory was created. No stage1 port was involved.

Survivors:
- M15: ADAMIC_MUTANT= go run ./review/test-audit/internal-lower-check_pragmas/witness/static prints static side effect then done. ADAMIC_MUTANT=M15 with the same command prints only done. Both run generated JavaScript through the unchanged oracle/node.mjs. All package tests survived this mutant. TestClassFeaturesStaticDeclarationsExecute only checks class metadata and a minimum Main length, so it does not prove execution.
- M17: ADAMIC_MUTANT= go run ./review/test-audit/internal-lower-check_pragmas/witness prints a TS2540 diagnostic labeled error TS2540. With ADAMIC_MUTANT=M17 it prints the same message labeled diagnostic TS2540. All package tests survived. Not an equivalent candidate: observable formatted output differs. Its central replay may find a catch in internal/load, which was not tested here.

Additional wrong-reason evidence:
ADAMIC_MUTANT= go run ./review/test-audit/internal-lower-check_pragmas/witness/narrowed refuses the source read at column 207. ADAMIC_MUTANT=M09 refuses the earlier read at column 167 instead. The existing TestClassFeaturesNarrowedAccessor passes both because it checks only the diagnostic fragment. Witness logs are saved. This survives that particular row, while other rows kill M09.

Brief friction, ambiguity, and costs:
- The supplied files were tied to 8de93800f4, but the mandated fresh origin/main was 8171b317. Listing, filenames and line numbers were verified against the latter. No row moved or vanished.
- Warm env.sh worked but dependency compilation was cold. The first listing took about 90 seconds and was interrupted before the test binary ran. A capped retry completed after cached compilation progress; there is no meaningful row narrowing for compilation of a shared dependency. The initial empty listing log and successful retry log are both preserved.
- An ordinary un-escalated attempt to redirect a code-read log was refused by the read-only filesystem. Subsequent writes used approved tool escalation. It did not affect production or test results.
- Building the scratch switch initially hit overlapping-text replacement and missing-newline errors. It was regenerated from the starting commit and passed vet before any matrix run. No verdict uses these failed scratch builds.
- Three mutants per 13 rows would require 39, exceeding the cap of 20. Used 18 distributed over reached production functions, fixed before matrix outcomes. This is a finite experiment, not exhaustive proof. In particular, untrue means no mutant in this menu failed that row; it is not a universal claim that no possible bug could fail it.
- Dropping only the Main append would leave an unused body local. M18 drops the whole module-body lowering loop, as instructed. The early-return diffs use an if true block so unchanged code remains compile-checkable without unreachable-code vet errors.
- Checker rows call Adamic Load directly; loader diagnostic collection/formatting is code under test. Upstream checker implementation is an external authority and was left unchanged. Lower-only rows are judged only by P1, never by P2 failures in their preparation helpers.
- Readonly and tsgo rows call both Load and Lower, but the brief supplies only one vacuous field. For readonly, P1 passes the positive lower subcase and P2 fails negative checker cases: aggregate vacuous=false, vacuous_subcases records the positive Lower weakness. PrivateChecker calls only Load and is judged only by P2.
- Full-package panics prevent complete uniqueness measurement for M14/M18. All selected rows were isolated as required; no inference was made for rows that never ran.
- Baseline skips outside this unit: TestOriginalCycleLedger (requires a specific pristine TS Git corpus plus generated diagnostics), TestOptionalWideningCensus (requires a user-selected project/output), and the array-contract subcase of TestMixedUnionContractGraph (explicitly awaits compiler/views-v3). None of the 13 rows skipped. The external census/corpus workflows were not provisioned as part of this 13-row unit; their catches remain unknown.
- Bare Node initially failed to resolve adamic for the static witness. The first-attempt log is saved; rerunning through the repository's unchanged oracle/node.mjs succeeded. The first attempt supports no verdict.
- Native rebuild durations and exact total setup/build wall time were not individually instrumented. Whole mutant command wall times are recorded in run-results.json; package binary times are recorded in each JSON log. No precise native-build-time claim is possible.
- Subsumption is based on the observed 18-mutant slice, with counts recorded per row. It is a retention hint, not a deletion recommendation. Production code was restored exactly; no test changes are proposed.

Timing and coverage limits:
Warm toolchain setup skipped; Go 1.27.1, nproc 5. npm ci command took 0.320 s. Fetch and branch creation command took 3.298 s. Initial listing was interrupted around 90 s; capped retry then completed. Clean whole-package binary 27.374 s; restored whole-package binary 22.296 s. Coverage command took 15.101 s. All 39 isolated row binaries total 3.933 s; their medians/runs are in timings.json. Standalone apply/vet validations total 18.041 s. The matrix, switched-clean check and all panic-isolated reruns total 498.200 command wall seconds. No test binary cooked at 90 s. No other package test was run. No installed-tool setup, stage1 port, repo-wide replay, missing corpus provisioning, new tests, or production fix was included.
