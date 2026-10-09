u050: 18 assigned functions, one family; all names exist at the starting commit.
Base: 0942c5169d0ea736d9dfaa19881af1ad8adad162.
Verdict: sacred in the bounded matrix; M1 is the family’s unique kill.
Median: 20.836 seconds; nproc: 5; no vacuous family members.
Evidence: review/test-audit/internal-native-normalize/ on test-audit/internal-native-normalize.

```json
[
  {
    "test": "TestNormalizeMatchesNode family",
    "members": [
      "TestNormalizeMatchesNodePoints00",
      "TestNormalizeMatchesNodePoints01",
      "TestNormalizeMatchesNodePoints02",
      "TestNormalizeMatchesNodePoints03",
      "TestNormalizeMatchesNodePoints04",
      "TestNormalizeMatchesNodePoints05",
      "TestNormalizeMatchesNodePoints06",
      "TestNormalizeMatchesNodePoints07",
      "TestNormalizeMatchesNodePoints08",
      "TestNormalizeMatchesNodePoints09",
      "TestNormalizeMatchesNodePoints10",
      "TestNormalizeMatchesNodePoints11",
      "TestNormalizeMatchesNodePoints12",
      "TestNormalizeMatchesNodePoints13",
      "TestNormalizeMatchesNodePoints14",
      "TestNormalizeMatchesNodePoints15",
      "TestNormalizeMatchesNodePoints16",
      "TestNormalizeMatchesNodeContexts"
    ],
    "package": "internal/native",
    "file": "internal/native/normalize_test.go:246",
    "seconds": 20.836,
    "oracle": "Node String.prototype.normalize executes in NFC, NFD, NFKC, NFKD; every WTF-8 output line must match byte for byte. A separate expected line count guards incomplete streams.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [
      "M1"
    ],
    "last_proven_fail": "M3: normalize_test.go:288: 65536 of 65536 lines differ",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNormalizeMatchesNode family",
      "TestNormalizeRandomMatchesNode"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u050/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^TestNormalize(MatchesNode|RandomMatchesNode$)' > review/test-audit/internal-native-normalize/M3.log 2>&1; normalize_test.go:288: 65536 of 65536 lines differ",
    "uniqueness_scope": "bounded matrix only"
  }
]
```

| ID | File:line at base | Change | Failed grouped rows |
|---|---|---|---|
| M1 | internal/native/runtime/normalize.c:47 | hangul_t_count = 28 → 27 | TestNormalizeMatchesNode family |
| M2 | internal/native/runtime/normalize.c:165 | decompose: !compatibility → compatibility | TestNormalizeMatchesNode family, TestNormalizeRandomMatchesNode |
| M3 | internal/native/runtime/normalize.c:246 | compose: composite(starter, point) → composite(point, starter) | TestNormalizeMatchesNode family, TestNormalizeRandomMatchesNode |
| P1 | internal/native/runtime/normalize.c:567 | entry returns adamic_string_allocate(0); probe only | TestNormalizeMatchesNode family, TestNormalizeRandomMatchesNode |

M1 evidence: normalize_test.go:248: 11238 of 65536 lines differ. Random passed M1. M2 evidence: normalize_test.go:333: 100790 of 346200 lines differ. M3 evidence: normalize_test.go:288: 65536 of 65536 lines differ. P1 evidence: normalize_test.go:293: 65536 of 65536 lines differ. The adjacent native and Node lines in the raw logs preserve the actual differing bytes.

Survivors: none among the three production mutants in the bounded matrix. M1 survived only the random comparator, not the family.

Brief ambiguities, costs, and limits:

- The heading calls these 18 rows, but the family rule makes them one row. Points and Contexts use the same checker with a different index; treating Contexts separately would falsely split an input family. Three mutants meet the grouped-row target.
- The brief’s file reference is 8de93800f4; the required fresh origin/main was 0942c516. All names and line references were checked against the actual starting commit. None moved or vanished.
- Whole-package timing consumed 90.029 seconds and ended in the mandated timeout, with no prior test failure. It cannot provide a completed green full-package baseline. The clean bounded rows passed before mutation. Unrun or unfinished package rows remain unknown.
- Compile-once switching conflicts with M1’s enum constant feeding other enum constants. The permitted separate-build fallback was used, with three production rebuilds and one probe rebuild. Build times include go run compilation, runtime compilation and native harness linking; they do not isolate clang time. Each variant uses its own ADAMIC_BUILD_CACHE_DIR; the runtime archive also keys on source contents.
- The caller search found the family, random comparison and opt-in long measurement. The long measurement constructs enormous strings and measures performance rather than adding a normal correctness fixture. It was excluded from this bounded correctness matrix, so its kills are unknown.
- No assigned test skipped or needed an opt-in. The full-package attempt recorded 86 unrelated skips, listed with reasons in baseline-skips.md. Enabling those unrelated SDK/corpus/product checks was not part of this slice.
- Family timing is the complete family’s test-binary ok line, measured three times with -count=1. Member timings inside a parallel family are not a substitute for that number. Cache warmth changes build cost, so rebuild wall times are reported separately.
- Sacred and unique_kills refer only to the two grouped rows in this bounded matrix. Package-wide and repository-wide uniqueness await central replay. A three-mutant result proves these catches, not all possible normalization behavior.
- The probe is an allocated empty string, not NULL: this preserves the normalization return type and permits byte comparison without an artificial null-dereference. All 18 members rejected it.

Timing: warm setup 0 seconds; npm ci reported 0.515 seconds; nproc 5. Family binary runs [20.705, 22.815, 20.836]; random comparator runs [10.763, 11.187, 11.092]. Rebuild wall seconds {'M1-build': 6.827, 'M2-build': 6.606, 'M3-build': 6.555, 'P1-build': 6.475}. Recorded binary running time 275.561 seconds. Exact commands, wall durations and exit statuses are in commands.json.

Not covered: other packages, repository uniqueness, completed full-package parity, skipped unrelated rows, long-string performance, native allocation failures and invalid normalization forms. No source changes remain and no pull request was opened. Standalone M1.diff, M2.diff, M3.diff and probe P1.diff apply to the starting commit and passed sanitizer-flags clang checks plus actual native builds.
