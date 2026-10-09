u025: both assigned rows exist, with no moves or vanished names.
Base: origin/main 7b18d0576930caca4e22ce2eef92fcf563af52d0.
Clean baseline: PASS, 1.090 test-binary seconds; no skips; nproc=5.
Verdicts: regexp sacred (M1 and M2 package-unique); pin setup-check (S1).
Evidence: test-audit/internal-load-pin, review/test-audit/internal-load-pin/.

```json
[
  {
    "package": "internal/load",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "subsumer_seconds": null,
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
    "test": "TestTypeScriptIsThePinnedCommit",
    "file": "internal/load/pin_test.go:17",
    "seconds": 0.005,
    "oracle": "Git rev-parse HEAD versus self-recorded verified commit d92d9bfee114c80be2c375d72edae966176e3a4f. Git is run; the historical claim that the suite verified this pin was not independently checked.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1 pin_test.go:24: TypeScript is at f901d0af4d5b5a9c445b2634aa4fe34ebff51ffd, and stage 0 was verified against d92d9bfee114c80be2c375d72edae966176e3a4f.",
    "verdict": "setup-check",
    "probe_kills": [],
    "vacuous": null,
    "evidence": "python3 review/test-audit/internal-load-pin/setup-probe.py; timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run ^TestTypeScriptIsThePinnedCommit$; S1.log pin_test.go:24: TypeScript is at f901d0af4d5b5a9c445b2634aa4fe34ebff51ffd"
  },
  {
    "package": "internal/load",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "subsumer_seconds": null,
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
    "test": "TestRegExpCaptureTypes",
    "file": "internal/load/regexp_test.go:5",
    "seconds": 0.073,
    "oracle": "Self-written nonempty CheckError expectation. No outside authority is invoked or cited. M4 passes all four subcases despite a configuration error: observation of the split input changes TS18048 (value possibly undefined) into TS5052 (exactOptionalPropertyTypes requires strictNullChecks). Diagnostic identity is unchecked.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [
      "M1",
      "M2"
    ],
    "last_proven_fail": "M2 regexp_test.go:11: Load: want a CheckError, got <nil>",
    "verdict": "sacred",
    "probe_kills": [
      "P1"
    ],
    "vacuous": false,
    "vacuous_subcases": [],
    "evidence": "ADAMIC_MUTANT=M2 timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > M2.log 2>&1; regexp_test.go:11: Load: want a CheckError, got <nil>; isolated P1 also fails all four subcases."
  }
]
```

| ID | Origin file:line (internal/load/) | Change | Failed rows |
|---|---|---|---|
| M1 | regexp_library.go:20 | Drop exec-array rewrite | TestRegExpCaptureTypes |
| M2 | regexp_library.go:21 | Drop match-array rewrite | TestRegExpCaptureTypes |
| M3 | regexp_library.go:22 | Drop RegExp split rewrite |  |
| M4 | load.go:59 | Strict core.TSTrue -> core.TSFalse | TestStarExportCollisionNamesBothModules, TestExplicitExportResolvesStarCollision, TestDeclarationsCarryTheirProvenTypes, TestATypeErrorFailsTheLoad, TestAdamicOptionsAreOn, TestEveryFileIsAModule, TestAdamicFilesImportEachOther, TestThePreludeDeclaresConsoleAndPanic, TestTheProgramsInTheSpecLoad, TestOverlayPrecedesDiskForAdamicAliases, TestHouseStyleOptionsDoNotRejectValidControlFlow, TestNodeLibraryUsesPinnedDeclarations, TestNodeLibraryKeepsOfficialConsoleSignatures |
| M5 | load.go:255 | diagnostic.Code() -> 0 | TestStarExportCollisionNamesBothModules, TestATypeErrorFailsTheLoad, TestAdamicOptionsAreOn, TestAdamicFilesImportEachOther, TestThePreludeDeclaresConsoleAndPanic, TestOverlayPrecedesDiskForAdamicAliases, TestHouseStyleOptionsDoNotRejectValidControlFlow, TestNodeLibraryUsesPinnedDeclarations |
| M6 | load.go:278 | column + 1 -> + 0 | TestStarExportCollisionNamesBothModules, TestDeclarationsCarryTheirProvenTypes, TestATypeErrorFailsTheLoad, TestAdamicFilesImportEachOther |

P1 is a probe only: Load at load.go:80 returns nil,nil. The whole package aborts on a nil dereference in TestAdamicFilesImportEachOther. Rows without a completed result remain unknown in matrix.csv. Isolated assigned rows: pin passes, regexp fails all four subcases. S1 changes only the checkout construction to an identical-tree scratch commit and fails pin_test.go:24. It is setup evidence, not a production kill or standalone production diff.

Survivor M3: actual regexpLibraryFS.ReadFile output changes, witnessed by `go test -v -count=1 -timeout 90s ./internal/load/ -run ^TestAuditU025DeclarationWitness$`: explicit RegExp split undefined-union declaration is true before, false after. Saved survivor-original.log and survivor-M3.log. This is an unguarded declaration transformation; semantic equivalence remains unresolved because the separately rewritten Symbol.split overload remains. No additional survivor among M1 through M6.

Brief friction and execution limits:
- The supplied reference commit is older than fetched origin/main. Scope came from the actual list; neither assigned name moved or vanished.
- Warm env.sh did not imply initialized submodules. Initial listing failed with missing cohere/TypeScript/tsc/go.mod. npm ci succeeded before the baseline, then recursive submodule initialization took about 15.2 seconds.
- Initial cold test listing compilation ran roughly 95 seconds. I omitted its timeout and exceeded the 90-second step budget; no test binary approached 90 seconds. Subsequent full package runs fit without narrowing. The exact initial build wall time was not instrumented, so this is approximate.
- Pin reaches no production code. Treating repository submodule construction as setup is the applicable setup-check interpretation. Changing the recorded pin in the test would have violated the no-oracle-mutation rule, so it was left intact.
- Predeclared reached functions were conservative: coverage later showed Stat and Realpath uncalled; writeChain was entered but its body/recursion was unexecuted. All nonzero entries are recorded in coverage-functions.txt. The M5 line in the preplan was misnumbered; its correct origin/main line is 255.
- Empty Load caused a package panic. Isolated assigned-row reruns resolve only their results; other unfinished rows remain unknown.
- Supplemental temporary observation tests measured actual function output and diagnostics. They were not counted as mutants or kills and were removed. Scripts preserve the survivor witness; oracle logs preserve the diagnostic observation.

Costs: warm env sourcing/version check under one second; no cloud/setup.sh run. npm ci reported 782 ms. Submodules approximately 15.2 s. Initial checker compilation approximately 95 s. Isolated test binary samples: pin 0.005,0.006,0.005 s (median 0.005); regexp 0.063,0.073,0.098 s (median 0.073). Standalone vet checks total 3.739 s wall; switched binary build 4.659 s; six matrix commands 13.184 s wall, about 6.067 s in test binaries. Full per-command timings in timings.json. Compiler mutants change only Go loading, with no native product builds and no native build cache involved.

All six standalone production diffs and P1.diff apply to the starting origin/main and pass go vet ./internal/load/. The runner uses one mutation selector and excludes switch scaffolding from diffs. Production sources and submodule HEAD were restored. No other packages, repo-wide uniqueness, native products, or outside semantic authority were covered. No PR or main push. No subsumption verdict rests on this matrix.
