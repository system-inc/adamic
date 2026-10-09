The baseline passed in 16.008 seconds; eight replays took 73.47 seconds of wall time. Every mutant retained a valid catcher outside the set, so all three candidates meet the deletable criterion. Nothing was stale or broken; no pins, witnesses, or Go panics counted as catches. Source is restored, and [evidence](https://github.com/system-inc/adamic/blob/38fb1ca37805dd03c74837dbd36c392a93053b8a/review/test-defend/deletion-set/internal-fuzz/report.json) was pushed to `test-defend/deletion-set/internal-fuzz`.

```json
{
  "package": "internal/fuzz",
  "main": "b8bcadb2c493173855f19d7e5c508b34f5eeb5b6",
  "skipped": [
    "TestGeneratedProgramsCheckAndLower",
    "TestOneSeedOneProgram",
    "TestRegexProgramsPassTheChecker"
  ],
  "mutants": [
    {
      "mutant": "M10",
      "file_line": "internal/fuzz/generate.go:167",
      "branch": "test-audit/internal-fuzz",
      "candidates_failed": [
        "TestGeneratedProgramsCheckAndLower",
        "TestRegexProgramsPassTheChecker"
      ],
      "still_caught_by": ["TestOwnershipShapes"],
      "stale": false
    },
    {
      "mutant": "G1",
      "file_line": "internal/fuzz/generate.go:1086",
      "branch": "test-defend/internal-fuzz",
      "candidates_failed": ["TestGeneratedProgramsCheckAndLower"],
      "still_caught_by": ["TestOverridesShapesAndLower"],
      "stale": false
    },
    {
      "mutant": "G2",
      "file_line": "internal/fuzz/generate.go:952",
      "branch": "test-defend/internal-fuzz",
      "candidates_failed": ["TestGeneratedProgramsCheckAndLower"],
      "still_caught_by": ["TestUndefinedNumbersShapes"],
      "stale": false
    },
    {
      "mutant": "G3",
      "file_line": "internal/fuzz/generate.go:1093",
      "branch": "test-defend/internal-fuzz",
      "candidates_failed": ["TestGeneratedProgramsCheckAndLower"],
      "still_caught_by": ["TestUndefinedNumbersShapes"],
      "stale": false
    },
    {
      "mutant": "R1",
      "file_line": "internal/fuzz/generate.go:218",
      "branch": "test-defend/internal-fuzz",
      "candidates_failed": [
        "TestGeneratedProgramsCheckAndLower",
        "TestRegexProgramsPassTheChecker"
      ],
      "still_caught_by": ["TestOwnershipShapes"],
      "stale": false
    },
    {
      "mutant": "R2",
      "file_line": "internal/fuzz/generate.go:1226",
      "branch": "test-defend/internal-fuzz",
      "candidates_failed": [
        "TestGeneratedProgramsCheckAndLower",
        "TestRegexProgramsPassTheChecker"
      ],
      "still_caught_by": ["TestOwnershipShapes"],
      "stale": false
    },
    {
      "mutant": "R3",
      "file_line": "internal/fuzz/generate.go:1214",
      "branch": "test-defend/internal-fuzz",
      "candidates_failed": [
        "TestGeneratedProgramsCheckAndLower",
        "TestRegexProgramsPassTheChecker"
      ],
      "still_caught_by": ["TestOwnershipShapes"],
      "stale": false
    },
    {
      "mutant": "S1",
      "file_line": "internal/fuzz/generate.go:27",
      "branch": "test-defend/internal-fuzz",
      "candidates_failed": ["TestOneSeedOneProgram"],
      "still_caught_by": [
        "TestOctoberFeaturesAppear",
        "TestUndefinedNumbersShapes"
      ],
      "stale": false
    }
  ],
  "keep": [],
  "deletable": [
    "TestGeneratedProgramsCheckAndLower",
    "TestOneSeedOneProgram",
    "TestRegexProgramsPassTheChecker"
  ],
  "pin_failures": [],
  "witness_failures": [],
  "panicking_tests": [],
  "notes": [
    "Candidate-specific defender diffs use G/R/S filenames rather than D; included as supplemental evidence.",
    "The full-package baseline passed 15 top-level tests with no default skips; no narrowing was needed.",
    "Replays stopped after clean catches; later rows are unknown.",
    "No tests were deleted or changed."
  ]
}
```
