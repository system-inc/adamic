u050 started clean at origin/main 83f3940ec77b8ba779da05ded113e9d9e7e710b6.
All 18 requested functions exist in normalize_test.go; one TestNormalizeMatchesNode family.
Clean package baseline is red: record_test.go:242 reports text file busy; package later timed out at 90.082 s.
Audit stopped before mutants, probes or median timing runs; verdict cannot-judge and vacuous=null.
Evidence destination: test-audit/internal-native-normalize, review/test-audit/internal-native-normalize/.

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
    "seconds": null,
    "oracle": "external-run: Node String.prototype.normalize in NFC, NFD, NFKC and NFKD; compares complete WTF-8 output lines, stream lengths, and process success.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run . > review/test-audit/internal-native-normalize/baseline.log 2>&1; baseline blocker: record_test.go:242: want LeakSanitizer to catch mutant: fork/exec /tmp/adamic-gate/TestRecordMutantsUnit03overwrite-key-leaked1446908639/001/main: text file busy",
    "reason": "Clean package baseline failed before its later 90-second timeout. The brief requires stopping on red, so no mutants, probes, or timing trials were run."
  }
]
```

Mutant table: empty. No mutants were planted.

Survivors: none tested.

The baseline command was `timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run . > review/test-audit/internal-native-normalize/baseline.log 2>&1`. Its concrete failure was `record_test.go:242: want LeakSanitizer to catch mutant: fork/exec /tmp/adamic-gate/TestRecordMutantsUnit03overwrite-key-leaked1446908639/001/main: text file busy`. This was an execution failure of a pre-existing record-mutant witness, not a mutation planted during u050 and not evidence against normalization. The failure preceded the timeout, so the brief's red-baseline stop rule applies; treating the run as only an over-budget baseline would discard the observed failure. No narrowed matrix was run.

The supplied reference 8de93800f4 differs from fetched origin/main 83f3940ec77b8ba779da05ded113e9d9e7e710b6. The requested test file has no diff between those commits. All 18 wrappers call the same checker with only the input selector changed, so the family rule makes this one row rather than 18 independent verdicts. The Contexts wrapper shares that checker and adds no independent assertion. All member names are retained in results.json.

Every skipped row is enumerated in baseline-summary.json. They are TestNormalizeLongMeasurements, TestRecordBenchmark, TestMeasureClangUnits, TestSplitTSGoAgreesUnit00 and Unit01, and TestWASIUnit00 through Unit35. None of the requested 18 skipped. Optional benchmark, checker-archive and WASI enablement was not pursued after the red-baseline stop. Therefore this session does not establish an enabled full-package baseline. The timeout leaves later test outcomes unknown; all 18 requested functions passed in this baseline, but that does not override the package failure. Their observed start/pass events are retained.

The mandatory stop prevents the requested three-run median, code-drawn mutants, empty-entry probe, uniqueness, subsumption and survivor witnesses. Those fields remain null or empty rather than inferred from source. The source inventory is conservative and is not a claim of measured function coverage. No production source changed, no PR was opened, and nothing was pushed to main.

Warm setup: 0 setup-script seconds, nproc=5. npm ci: 0.378 s, status 0. Test-list build/startup: 6.312 s, status 0. Clean package build-and-run wall: 91.793 s; test binary: 90.082 s, status 1. Build-only time was not isolated. No mutation rebuilds or further test runs occurred.
