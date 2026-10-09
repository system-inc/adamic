u068 started from origin/main 6c60da091afddc9c2fe88b3a1067845b6dc79cb3, nproc 5.
All 15 names exist in their listed files; the shared readiness checker groups three names into one family, yielding 13 rows.
Full clean baseline timed out at 90.108 seconds; the selected clean slice passed in 2.945 seconds, with no selected skips.
Four production mutants support 2 bounded sacred rows; weakened checks prove 10 witness rows; the dependency construction row is untrue.
Four empty-entry probes were caught; no production mutant survived; standalone evidence is on test-audit/internal-oracle-private_generic_mutant.

```json
[
  {
    "test": "TestPrivateGenericMethodOwnerMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/private_generic_mutant_test.go",
    "seconds": 0.09,
    "oracle": "Live source Node compared with deliberately changed native or JavaScript output.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: private_generic_mutant_test.go:54: private method owner mutant survived: \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => private_generic_mutant_test.go:54: private method owner mutant survived: \"\"",
    "members": [
      "TestPrivateGenericMethodOwnerMutant"
    ]
  },
  {
    "test": "TestReadinessMutants family",
    "package": "internal/oracle",
    "file": "internal/oracle/readiness_test.go",
    "seconds": 0.926,
    "oracle": "Self eager non-null panic contract and harmless-output detector; source Node only sanity-checked, own JavaScript backend agreement.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: readiness_test.go:147: drop-check mutant was not caught by harmless output: oracle.run{stdout:[]uint8{0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0xa, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0xa, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x20, ",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => readiness_test.go:147: drop-check mutant was not caught by harmless output: oracle.run{stdout:[]uint8{0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0xa, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x65, 0x72, 0xa, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x20, ",
    "members": [
      "TestReadinessMutants",
      "TestUninitializedIsNotNullishMutant",
      "TestLazyInitializerIsNotEagerMutant"
    ]
  },
  {
    "test": "TestNonNullWeakFreedNamesExpression",
    "package": "internal/oracle",
    "file": "internal/oracle/readiness_test.go",
    "seconds": 0.358,
    "oracle": "Live source Node for pinned tracing output; self pinned native freed-Weak diagnostic.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: readiness_test.go:136: freed Weak diagnostic mutant escaped pinned output",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => readiness_test.go:136: freed Weak diagnostic mutant escaped pinned output",
    "members": [
      "TestNonNullWeakFreedNamesExpression"
    ]
  },
  {
    "test": "TestClosedNestedConstituentMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/real_nested_functions_test.go",
    "seconds": 0.295,
    "oracle": "Live source Node compared with deliberately changed native or JavaScript output.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: real_nested_functions_test.go:72: native incorrect constituent count escaped Node",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => real_nested_functions_test.go:72: native incorrect constituent count escaped Node",
    "members": [
      "TestClosedNestedConstituentMutant"
    ]
  },
  {
    "test": "TestRegexCycleFixtureHasItsNativeDependency",
    "package": "internal/oracle",
    "file": "internal/oracle/regexp_cycle_test.go",
    "seconds": 0.008,
    "oracle": "Self reflection on ir.Program.Regexps. Body has no failure assertion; false dependency result only skips.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestRegexCycleFixtureHasItsNativeDependency"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-S01-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^TestRegexCycleFixtureHasItsNativeDependency$ => SKIP: native regex lowering/emission is on codex/stage1-css; run this fixture on the scratch merge",
    "members": [
      "TestRegexCycleFixtureHasItsNativeDependency"
    ]
  },
  {
    "test": "TestRegExpReplacementNodeMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/regexp_replace_test.go",
    "seconds": 0.363,
    "oracle": "Live source Node compared with deliberately changed native or JavaScript output.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: regexp_replace_test.go:64: Node did not catch mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => regexp_replace_test.go:64: Node did not catch mutant",
    "members": [
      "TestRegExpReplacementNodeMutants"
    ]
  },
  {
    "test": "TestRegExpReplacementTypeGuardMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/regexp_replace_test.go",
    "seconds": 0.296,
    "oracle": "Self declared-type guard policy, using our JavaScript backend output as native expectation.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: regexp_replace_test.go:108: backend comparison did not catch mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => regexp_replace_test.go:108: backend comparison did not catch mutant",
    "members": [
      "TestRegExpReplacementTypeGuardMutants"
    ]
  },
  {
    "test": "TestRepresentationClockSourceMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/representation_clock_source_test.go",
    "seconds": 0.109,
    "oracle": "Live source Node compared with deliberately changed native or JavaScript output.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: representation_clock_source_test.go:55: want stdout kill, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => representation_clock_source_test.go:55: want stdout kill, got \"\"",
    "members": [
      "TestRepresentationClockSourceMutant"
    ]
  },
  {
    "test": "TestRuntimeLastIndexOfMatchesNode",
    "package": "internal/oracle",
    "file": "internal/oracle/runtime_last_index_test.go",
    "seconds": 0.07,
    "oracle": "Live source Node stdout, stderr and exit compared with sanitized, release and our JavaScript backend; leak checking.",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M02"
    ],
    "unique_kills": [
      "M01",
      "M02"
    ],
    "last_proven_fail": "M02: runtime_last_index_test.go:22: sanitized: stdout differs; Node stdout \"0\\n0\\n-1\\n-1\\n102\\n100\\n101\\n98\\n96\\n95\\n96\\n95\\n96\\n95\\n96\\n95\\n96\\n89\\n102\\n100\\n101\\n98\\n99\\n98\\n99\\n96\\n97\\n96\\n97\\n90\\n102\\n100\\n100\\n99\\n100\\n97\\n98\\n97\\n98\\n91\\n102\\n100\\n1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01",
      "P02",
      "P03",
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRuntimeLastIndexOfMatchesNode",
      "TestScannerNestedReferences"
    ],
    "evidence": "ADAMIC_MUTANT=M02 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u068/cache/M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestRuntimeLastIndexOfMatchesNode|TestScannerNestedReferences)$ => runtime_last_index_test.go:22: sanitized: stdout differs; Node stdout \"0\\n0\\n-1\\n-1\\n102\\n100\\n101\\n98\\n96\\n95\\n96\\n95\\n96\\n95\\n96\\n95\\n96\\n89\\n102\\n100\\n101\\n98\\n99\\n98\\n99\\n96\\n97\\n96\\n97\\n90\\n102\\n100\\n100\\n99\\n100\\n97\\n98\\n97\\n98\\n91\\n102\\n100\\n1",
    "members": [
      "TestRuntimeLastIndexOfMatchesNode"
    ]
  },
  {
    "test": "TestScannerNestedOverloadImplementationMutantIsCaught",
    "package": "internal/oracle",
    "file": "internal/oracle/scanner_nested_overload_test.go",
    "seconds": 0.209,
    "oracle": "Live source Node compared with deliberately changed native or JavaScript output.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: scanner_nested_overload_test.go:57: want Node stdout catcher, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => scanner_nested_overload_test.go:57: want Node stdout catcher, got \"\"",
    "members": [
      "TestScannerNestedOverloadImplementationMutantIsCaught"
    ]
  },
  {
    "test": "TestScannerNestedReferenceMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/scanner_nested_references_test.go",
    "seconds": 0.254,
    "oracle": "Live source Node compared with deliberately changed native or JavaScript output.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: scanner_nested_references_test.go:70: native: want stdout catcher, got \"\"",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => scanner_nested_references_test.go:70: native: want stdout catcher, got \"\"",
    "members": [
      "TestScannerNestedReferenceMutants"
    ]
  },
  {
    "test": "TestScannerNestedReferences",
    "package": "internal/oracle",
    "file": "internal/oracle/scanner_nested_references_test.go",
    "seconds": 0.422,
    "oracle": "Live source Node stdout, stderr and exit compared with sanitized, release and our JavaScript backend; leak checking.",
    "oracle_kind": "external-run",
    "kills": [
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M03",
      "M04"
    ],
    "last_proven_fail": "M04: scanner_nested_references_test.go:102: JavaScript exit codes differ: output \"\", Node \"x\\n\", stderr \"adamic: panic: TypeError: Cannot read properties of undefined (reading 'code')\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01",
      "P02",
      "P03"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRuntimeLastIndexOfMatchesNode",
      "TestScannerNestedReferences"
    ],
    "evidence": "ADAMIC_MUTANT=M04 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u068/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestRuntimeLastIndexOfMatchesNode|TestScannerNestedReferences)$ => scanner_nested_references_test.go:102: JavaScript exit codes differ: output \"\", Node \"x\\n\", stderr \"adamic: panic: TypeError: Cannot read properties of undefined (reading 'code')\\n\"",
    "members": [
      "TestScannerNestedReferences"
    ]
  },
  {
    "test": "TestNestedReferenceIdentityMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/scanner_nested_references_test.go",
    "seconds": 0.308,
    "oracle": "Live source Node compared with deliberately changed native or JavaScript output.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W: scanner_nested_references_test.go:143: JavaScript identity mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPrivateGenericMethodOwnerMutant",
      "TestReadinessMutants family",
      "TestNonNullWeakFreedNamesExpression",
      "TestClosedNestedConstituentMutant",
      "TestRegExpReplacementNodeMutants",
      "TestRegExpReplacementTypeGuardMutants",
      "TestRepresentationClockSourceMutant",
      "TestScannerNestedOverloadImplementationMutantIsCaught",
      "TestScannerNestedReferenceMutants",
      "TestNestedReferenceIdentityMutant"
    ],
    "evidence": "ADAMIC_GATE_UNCACHED=1 timeout 120 go test -json -overlay /tmp/u068-W-overlay.json -count=1 -timeout 90s ./internal/oracle/ -run ^(TestPrivateGenericMethodOwnerMutant|TestReadinessMutants|TestUninitializedIsNotNullishMutant|TestNonNullWeakFreedNamesExpression|TestLazyInitializerIsNotEagerMutant|TestClosedNestedConstituentMutant|TestRegExpReplacementNodeMutants|TestRegExpReplacementTypeGuardMutants|TestRepresentationClockSourceMutant|TestScannerNestedOverloadImplementationMutantIsCaught|TestScannerNestedReferenceMutants|TestNestedReferenceIdentityMutant)$ => scanner_nested_references_test.go:143: JavaScript identity mutant survived",
    "members": [
      "TestNestedReferenceIdentityMutant"
    ]
  }
]
```

| ID | Origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/native/runtime/string_search_impl.h:115 | return (double)adamic_string_units(string); → return 0; | TestRuntimeLastIndexOfMatchesNode |
| M02 | internal/native/runtime/string_search_impl.h:123 | string->length - search->length + 1 → string->length - search->length | TestRuntimeLastIndexOfMatchesNode |
| M03 | internal/native/emit_expressions.go:288 | if identity > 0 { → if identity < 0 { | TestScannerNestedReferences |
| M04 | internal/javascript/javascript.go:51 | if (!values.has(code)) → if (values.has(code)) | TestScannerNestedReferences |

Survivors: none. Every production mutant was caught by a semantic row.

The reference commit 8de93800f4 is older than the fetched starting commit. None of the fifteen names moved or vanished. The brief calls them fifteen rows, but three readiness wrappers differ only by inputs to assertMigratedNonNullCheck and must be grouped. This family has three additional count=1 timing runs; individual member timing logs are retained too. There is no common name prefix, so the family takes its first member's name.

Twelve original functions are built-in-mutant witnesses. Counting their production failures as kills would be wrong. The Readiness helper now tests one removed non-null check per input, despite historical subcase names such as initialize-to-zero or miss-exception-path. The probe struct's stdout and stderr fields are not used by its body. The expected eager panic text is self-written policy. The replacement type-guard witness takes its expected result from our JavaScript backend, so its expected answer is self, not an independent Node source oracle.

The regex dependency row has no assertion after its skip guard. S01 makes nativeRegexSlicePresent return false; the row skips and the binary exits zero. Therefore the construction break is not caught. The skip text still says native regex is on codex/stage1-css even though the clean starting commit has the Regexps representation and the row passes. This is an observed untrue setup row, not a failed semantic test.

The first npm/list/baseline shell was launched from stage3/api. npm ci succeeded there, but relative Go package paths did not resolve. Those commands never executed tests. They were rerun from the repository root before any audit. The 90 second full-package timeout was not an individual failure. The matrix was narrowed to the two ordinary semantic rows in this slice; all witness and construction checks were run separately. Other package rows, including the general fixture family, were not replayed. Every unique kill is bounded to this two-row production matrix. Central replay must settle package and repository uniqueness.

Full baseline skips are listed in baseline-skips.json. They are WASI opt-ins outside the requested slice; they were not enabled or installed for this audit. All selected rows ran without skips in the clean baseline, including the dependency check. Node dependencies were refreshed with npm ci. The warm toolchain worked, so setup.sh was skipped.

reached-functions.txt lists 386 measured Go functions across oracle, lower, native and javascript. These measurements include oracle cache/harness production files; harness helpers defined in _test.go are not instrumented by Go coverage. All nine requested test files and the migrated-check helper were read whole. Runtime C definitions are conservatively listed in runtime-functions-static-superset.json. Dynamic C reachability was not instrumented because llvm-cov and llvm-profdata were unavailable. Thus the exact exhaustive dynamic function inventory remains a limitation; mutants themselves come from read, reached compiler/runtime code, not oracle or fixture expectations.

The fixed menu was written in mutant-plan.json before production outcomes. M01/M02 cover last-index constants and bounds; M03 reverses the positive frame-identity condition; M04 flips canonical-closure cache construction. No supplemental inserted-statement mutant contributes to a verdict. Selector instrumentation is isolated in switch.diff. Every production/probe standalone diff applies to the starting sources. Go changes passed go vet against original-source overlays; runtime changes compiled with native C11 warnings, count and sanitizer flags. The Lower empty probe drops the whole body and removes its now-unused fmt/filepath imports. Witness edits only weaken comparison checks; they pass go vet separately. Production files were restored before commit.

Compiler and runtime selectors use separate ADAMIC_BUILD_CACHE_DIR paths and ADAMIC_GATE_UNCACHED=1. This second setting matters: cached oracle observations can otherwise conceal a runtime environment selector even when the compiled bytes are identical. The switched clean control passed in 19.812 binary seconds. Production mutants completed in 0.666, 0.593, 0.677 and 0.604 binary seconds. The Lower nil probe panics, so each ordinary row was run alone; only observed row failures are recorded, and untouched subcases remain unknown. Witness vacuity is null because production empty-answer probes do not judge their check.

Setup.sh time was zero; npm reported 389ms. The full baseline cost 90.108 binary seconds, slice baseline 2.945, measured Go coverage 4.208. All grouped rows have three separate count=1 timing runs. build-and-run-times.json records each matrix/probe wall and binary time. Individual native rebuild time was not separately instrumented and remains included in binary elapsed time. The fixed switched runtime was reused; the per-selector build-cache roots and uncached oracle results avoided stale answers. Total session time was about 19 minutes including reading, reporting and push. No selected row exceeded 60 seconds alone.

Not covered: production kills outside the two ordinary rows, repository-wide uniqueness, exact dynamic C/test-helper coverage, separately measured native rebuild times, and unrelated WASI opt-ins. No survivor needs an equivalence witness. No PR, main push, production source change, or deletion recommendation is included.
