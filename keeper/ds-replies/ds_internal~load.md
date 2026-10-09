With the candidate skipped, the clean baseline passed (2.015 test-binary seconds), and all seven recorded mutants remained caught. Replays took 9.089–15.903 wall seconds each, 72.444 seconds total; none were stale, broken, or panicking. The restored check passed in 3.937 wall seconds. No tests were deleted, the tree is clean, and evidence was pushed to `test-defend/deletion-set/internal-load` at `59df14e9`, under `review/test-defend/deletion-set/internal-load/`.

```json
{
  "package": "internal/load",
  "main": "d29d80ceb5d5d42d9b0ffb7b528a57a2272c76f4",
  "skipped": ["TestExplicitExportResolvesStarCollision"],
  "mutants": [
    {
      "mutant": "M03",
      "file_line": "internal/load/load.go:59",
      "branch": "test-audit/internal-load-export_collision",
      "candidates_failed": ["TestExplicitExportResolvesStarCollision"],
      "still_caught_by": [
        "TestATypeErrorFailsTheLoad",
        "TestAdamicFilesImportEachOther",
        "TestAdamicOptionsAreOn",
        "TestDeclarationsCarryTheirProvenTypes",
        "TestEveryFileIsAModule",
        "TestHouseStyleOptionsDoNotRejectValidControlFlow",
        "TestNodeLibraryKeepsOfficialConsoleSignatures",
        "TestNodeLibraryUsesPinnedDeclarations",
        "TestOverlayPrecedesDiskForAdamicAliases",
        "TestStarExportCollisionNamesBothModules",
        "TestThePreludeDeclaresConsoleAndPanic",
        "TestTheProgramsInTheSpecLoad"
      ],
      "stale": false
    },
    {
      "mutant": "M05",
      "file_line": "internal/load/load.go:72",
      "branch": "test-audit/internal-load-export_collision",
      "candidates_failed": ["TestExplicitExportResolvesStarCollision"],
      "still_caught_by": [
        "TestATypeErrorFailsTheLoad",
        "TestAdamicFilesImportEachOther",
        "TestAdamicOptionsAreOn",
        "TestDeclarationsCarryTheirProvenTypes",
        "TestEveryFileIsAModule",
        "TestHouseStyleOptionsDoNotRejectValidControlFlow",
        "TestOverlayPrecedesDiskForAdamicAliases",
        "TestStarExportCollisionNamesBothModules",
        "TestThePreludeDeclaresConsoleAndPanic",
        "TestTheProgramsInTheSpecLoad"
      ],
      "stale": false
    },
    {
      "mutant": "M12",
      "file_line": "internal/load/source_fs.go:40",
      "branch": "test-audit/internal-load-export_collision",
      "candidates_failed": ["TestExplicitExportResolvesStarCollision"],
      "still_caught_by": [
        "TestAdamicFilesImportEachOther",
        "TestEveryFileIsAModule",
        "TestHouseStyleOptionsDoNotRejectValidControlFlow",
        "TestNodeLibraryKeepsOfficialConsoleSignatures",
        "TestNodeLibraryUsesPinnedDeclarations",
        "TestOverlayPrecedesDiskForAdamicAliases",
        "TestStarExportCollisionNamesBothModules"
      ],
      "stale": false
    },
    {
      "mutant": "M19",
      "file_line": "internal/load/load.go:214",
      "branch": "test-audit/internal-load-export_collision",
      "candidates_failed": ["TestExplicitExportResolvesStarCollision"],
      "still_caught_by": [
        "TestAdamicFilesImportEachOther",
        "TestEveryFileIsAModule",
        "TestHouseStyleOptionsDoNotRejectValidControlFlow",
        "TestLoadRefusesWhatIsNotAdamic",
        "TestNodeLibraryKeepsOfficialConsoleSignatures",
        "TestNodeLibraryUsesPinnedDeclarations",
        "TestOverlayPrecedesDiskForAdamicAliases",
        "TestRegExpCaptureTypes",
        "TestStarExportCollisionNamesBothModules"
      ],
      "stale": false
    },
    {
      "mutant": "D04",
      "file_line": "internal/load/load.go:71",
      "branch": "test-defend/internal-load-export_collision",
      "candidates_failed": ["TestExplicitExportResolvesStarCollision"],
      "still_caught_by": [
        "TestATypeErrorFailsTheLoad",
        "TestAdamicFilesImportEachOther",
        "TestDeclarationsCarryTheirProvenTypes",
        "TestEveryFileIsAModule",
        "TestHouseStyleOptionsDoNotRejectValidControlFlow",
        "TestNodeLibraryKeepsOfficialConsoleSignatures",
        "TestNodeLibraryUsesPinnedDeclarations",
        "TestOverlayPrecedesDiskForAdamicAliases",
        "TestStarExportCollisionNamesBothModules",
        "TestThePreludeDeclaresConsoleAndPanic",
        "TestTheProgramsInTheSpecLoad"
      ],
      "stale": false
    },
    {
      "mutant": "D07",
      "file_line": "internal/load/load.go:70",
      "branch": "test-defend/internal-load-export_collision",
      "candidates_failed": ["TestExplicitExportResolvesStarCollision"],
      "still_caught_by": [
        "TestATypeErrorFailsTheLoad",
        "TestAdamicFilesImportEachOther",
        "TestDeclarationsCarryTheirProvenTypes",
        "TestEveryFileIsAModule",
        "TestHouseStyleOptionsDoNotRejectValidControlFlow",
        "TestNodeLibraryKeepsOfficialConsoleSignatures",
        "TestNodeLibraryUsesPinnedDeclarations",
        "TestOverlayPrecedesDiskForAdamicAliases",
        "TestStarExportCollisionNamesBothModules",
        "TestThePreludeDeclaresConsoleAndPanic",
        "TestTheProgramsInTheSpecLoad"
      ],
      "stale": false
    },
    {
      "mutant": "D08",
      "file_line": "internal/load/load.go:66",
      "branch": "test-defend/internal-load-export_collision",
      "candidates_failed": ["TestExplicitExportResolvesStarCollision"],
      "still_caught_by": [
        "TestAdamicFilesImportEachOther",
        "TestNodeLibraryUsesPinnedDeclarations",
        "TestStarExportCollisionNamesBothModules",
        "TestThePreludeDeclaresConsoleAndPanic",
        "TestTheProgramsInTheSpecLoad"
      ],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": ["TestExplicitExportResolvesStarCollision"]
}
```
