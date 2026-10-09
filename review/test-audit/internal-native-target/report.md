u054 started clean at 0942c5169d0ea736d9dfaa19881af1ad8adad162; all 48 listed functions exist.
Two unit families give 12 audit rows; no requested function moved or vanished.
Whole package timed out at 90.060 s without an individual failure; bounded default and enabled WASI baselines passed.
Verdicts: 2 cannot-judge, 8 sacred, 1 subsumed, 1 witness. All verdicts and uniqueness are bounded.
17 production mutants, one witness weakening and 20 entry probes; all production mutants were killed.

```json
[
  {
    "test": "TestWASITargetFlags",
    "package": "internal/native",
    "file": "internal/native/target_test.go:11",
    "seconds": 0.008,
    "oracle": "Self-written required WASI flags and absence of WASI flags for native.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03"
    ],
    "unique_kills": [
      "M01",
      "M02",
      "M03"
    ],
    "last_proven_fail": "M03: target_test.go:16: missing -mno-atomics in -std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls --target=wasm32-wasi --sysroot=/tmp/adamic-gate/TestWASITargetFlags3379192136/001 -DADAMIC_TARGET_WASI=1 -matomics -O2",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "PFlags"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASITargetFlags",
      "TestWASIRefusesUnsupportedOptions",
      "TestTSGoBuildSeesTheProgramsFeatures",
      "TestTypedArrayRuntime",
      "TestUnitsPreserveSharedState",
      "TestUnitCacheFlagsHoldSanitizer",
      "TestSplitTokensDoNotRewriteLiterals",
      "TestUnitSystemHeaderProvenance",
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/M03 ADAMIC_MUTANT=M03 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestWASITargetFlags|TestWASIRefusesUnsupportedOptions|TestTSGoBuildSeesTheProgramsFeatures|TestTypedArrayRuntime|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance|TestViewMixedUnionUnknownAndUnavailable)$ > M03.log 2>&1; M03: target_test.go:16: missing -mno-atomics in -std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls --target=wasm32-wasi --sysroot=/tmp/adamic-gate/TestWASITargetFlags3379192136/001 -DADAMIC_TARGET_WASI=1 -matomics -O2"
  },
  {
    "test": "TestWASIRefusesUnsupportedOptions",
    "package": "internal/native",
    "file": "internal/native/target_test.go:25",
    "seconds": 0.008,
    "oracle": "Self-written refusal policy. Most invalid cases require only a non-nil error; the missing-sysroot case checks its diagnostic substring.",
    "oracle_kind": "self",
    "kills": [
      "M04",
      "M05",
      "M06",
      "M07"
    ],
    "unique_kills": [
      "M04",
      "M05",
      "M06",
      "M07"
    ],
    "last_proven_fail": "M07: target_test.go:34: missing sysroot: native: WASI_SYSROOT is not a directory:",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "PValidate"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASITargetFlags",
      "TestWASIRefusesUnsupportedOptions",
      "TestTSGoBuildSeesTheProgramsFeatures",
      "TestTypedArrayRuntime",
      "TestUnitsPreserveSharedState",
      "TestUnitCacheFlagsHoldSanitizer",
      "TestSplitTokensDoNotRewriteLiterals",
      "TestUnitSystemHeaderProvenance",
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/M07 ADAMIC_MUTANT=M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestWASITargetFlags|TestWASIRefusesUnsupportedOptions|TestTSGoBuildSeesTheProgramsFeatures|TestTypedArrayRuntime|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance|TestViewMixedUnionUnknownAndUnavailable)$ > M07.log 2>&1; M07: target_test.go:34: missing sysroot: native: WASI_SYSROOT is not a directory:"
  },
  {
    "test": "TestTSGoBuildSeesTheProgramsFeatures",
    "package": "internal/native",
    "file": "internal/native/tsgo_features_test.go:17",
    "seconds": 0.1,
    "oracle": "Clang syntax checks plus self-written expectation that missing feature flags must fail; W01 alone proves the witness.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01: tsgo_features_test.go:68: main.c compiled without its feature flags, so this test no longer covers the include order",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "PTSFlags",
      "PFeatureFlags"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestWASITargetFlags",
      "TestWASIRefusesUnsupportedOptions",
      "TestTSGoBuildSeesTheProgramsFeatures",
      "TestTypedArrayRuntime",
      "TestUnitsPreserveSharedState",
      "TestUnitCacheFlagsHoldSanitizer",
      "TestSplitTokensDoNotRewriteLiterals",
      "TestUnitSystemHeaderProvenance",
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/W01 ADAMIC_MUTANT=W01 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestWASITargetFlags|TestWASIRefusesUnsupportedOptions|TestTSGoBuildSeesTheProgramsFeatures|TestTypedArrayRuntime|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance|TestViewMixedUnionUnknownAndUnavailable)$ > W01.log 2>&1; W01: tsgo_features_test.go:68: main.c compiled without its feature flags, so this test no longer covers the include order"
  },
  {
    "test": "TestTypedArrayRuntime",
    "package": "internal/native",
    "file": "internal/native/typed_array_test.go:12",
    "seconds": 0.745,
    "oracle": "Node determines normal typed-array output and RangeError cases. Self-written exit-70 and exact index diagnostics; range probes check only exit 70 and a panic prefix and can accept a different panic.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M13",
      "M14"
    ],
    "unique_kills": [
      "M13",
      "M14"
    ],
    "last_proven_fail": "M14: typed_array_test.go:38: runtime: exit status 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "PT_new",
      "PT_from_numbers",
      "PT_get",
      "PT_check_write",
      "PT_set",
      "PT_length",
      "PT_fill",
      "PT_set_from",
      "PT_subarray",
      "PT_iterate",
      "PT_iterator_next"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASITargetFlags",
      "TestWASIRefusesUnsupportedOptions",
      "TestTSGoBuildSeesTheProgramsFeatures",
      "TestTypedArrayRuntime",
      "TestUnitsPreserveSharedState",
      "TestUnitCacheFlagsHoldSanitizer",
      "TestSplitTokensDoNotRewriteLiterals",
      "TestUnitSystemHeaderProvenance",
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/M14 ADAMIC_MUTANT=M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestWASITargetFlags|TestWASIRefusesUnsupportedOptions|TestTSGoBuildSeesTheProgramsFeatures|TestTypedArrayRuntime|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance|TestViewMixedUnionUnknownAndUnavailable)$ > M14.log 2>&1; M14: typed_array_test.go:38: runtime: exit status 1",
    "vacuous_subcases": [
      "PT_check_write: release normal-output comparison before invalid-write probes",
      "PT_check_write: sanitized normal-output comparison before invalid-write probes"
    ]
  },
  {
    "test": "TestUnitsPreserveSharedState",
    "package": "internal/native",
    "file": "internal/native/units_test.go:13",
    "seconds": 0.367,
    "oracle": "Self-written deterministic split and native output 1 2 2.",
    "oracle_kind": "self",
    "kills": [
      "M09",
      "M10",
      "M12"
    ],
    "unique_kills": [
      "M12"
    ],
    "last_proven_fail": "M12: units_test.go:38: native: linking units: exit status 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "PSplit",
      "PBuild"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASITargetFlags",
      "TestWASIRefusesUnsupportedOptions",
      "TestTSGoBuildSeesTheProgramsFeatures",
      "TestTypedArrayRuntime",
      "TestUnitsPreserveSharedState",
      "TestUnitCacheFlagsHoldSanitizer",
      "TestSplitTokensDoNotRewriteLiterals",
      "TestUnitSystemHeaderProvenance",
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/M12 ADAMIC_MUTANT=M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestWASITargetFlags|TestWASIRefusesUnsupportedOptions|TestTSGoBuildSeesTheProgramsFeatures|TestTypedArrayRuntime|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance|TestViewMixedUnionUnknownAndUnavailable)$ > M12.log 2>&1; M12: units_test.go:38: native: linking units: exit status 1"
  },
  {
    "test": "TestUnitCacheFlagsHoldSanitizer",
    "package": "internal/native",
    "file": "internal/native/units_test.go:48",
    "seconds": 0.163,
    "oracle": "Clang UBSan runs the signed-overflow probe; self-written diagnostic substring requires signed integer overflow.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M10",
      "M11"
    ],
    "unique_kills": [
      "M11"
    ],
    "last_proven_fail": "M11: units_test.go:83: sanitized rebuild reused uninstrumented object: exit=<nil> output=",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "PCompile"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASITargetFlags",
      "TestWASIRefusesUnsupportedOptions",
      "TestTSGoBuildSeesTheProgramsFeatures",
      "TestTypedArrayRuntime",
      "TestUnitsPreserveSharedState",
      "TestUnitCacheFlagsHoldSanitizer",
      "TestSplitTokensDoNotRewriteLiterals",
      "TestUnitSystemHeaderProvenance",
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/M11 ADAMIC_MUTANT=M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestWASITargetFlags|TestWASIRefusesUnsupportedOptions|TestTSGoBuildSeesTheProgramsFeatures|TestTypedArrayRuntime|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance|TestViewMixedUnionUnknownAndUnavailable)$ > M11.log 2>&1; M11: units_test.go:83: sanitized rebuild reused uninstrumented object: exit=<nil> output="
  },
  {
    "test": "TestSplitTokensDoNotRewriteLiterals",
    "package": "internal/native",
    "file": "internal/native/units_test.go:88",
    "seconds": 0.008,
    "oracle": "Self-written literal preservation and rewritten symbol spelling.",
    "oracle_kind": "self",
    "kills": [
      "M08",
      "M09"
    ],
    "unique_kills": [
      "M08"
    ],
    "last_proven_fail": "M09: units_test.go:93: native: split: unsupported declaration near \" \"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "PSplit"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASITargetFlags",
      "TestWASIRefusesUnsupportedOptions",
      "TestTSGoBuildSeesTheProgramsFeatures",
      "TestTypedArrayRuntime",
      "TestUnitsPreserveSharedState",
      "TestUnitCacheFlagsHoldSanitizer",
      "TestSplitTokensDoNotRewriteLiterals",
      "TestUnitSystemHeaderProvenance",
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/M09 ADAMIC_MUTANT=M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestWASITargetFlags|TestWASIRefusesUnsupportedOptions|TestTSGoBuildSeesTheProgramsFeatures|TestTypedArrayRuntime|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance|TestViewMixedUnionUnknownAndUnavailable)$ > M09.log 2>&1; M09: units_test.go:93: native: split: unsupported declaration near \" \""
  },
  {
    "test": "TestUnitSystemHeaderProvenance",
    "package": "internal/native",
    "file": "internal/native/units_test.go:104",
    "seconds": 0.223,
    "oracle": "Clang distinguishes system-header and user-header extensions; self-written cached/uncached output 1 and 2 and diagnostic substring.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: units_test.go:143: system-header provenance lost: native: preprocessing unit probe.c: exit status 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestUnitCacheFlagsHoldSanitizer"
    ],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "PCompile"
    ],
    "subsumer_seconds": 0.163,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASITargetFlags",
      "TestWASIRefusesUnsupportedOptions",
      "TestTSGoBuildSeesTheProgramsFeatures",
      "TestTypedArrayRuntime",
      "TestUnitsPreserveSharedState",
      "TestUnitCacheFlagsHoldSanitizer",
      "TestSplitTokensDoNotRewriteLiterals",
      "TestUnitSystemHeaderProvenance",
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/M10 ADAMIC_MUTANT=M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestWASITargetFlags|TestWASIRefusesUnsupportedOptions|TestTSGoBuildSeesTheProgramsFeatures|TestTypedArrayRuntime|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance|TestViewMixedUnionUnknownAndUnavailable)$ > M10.log 2>&1; M10: units_test.go:143: system-header provenance lost: native: preprocessing unit probe.c: exit status 1",
    "subsumption_mutants": 1
  },
  {
    "test": "TestViewMixedUnionUnknownAndUnavailable",
    "package": "internal/native",
    "file": "internal/native/view_unions_mixed_test.go:11",
    "seconds": 0.242,
    "oracle": "Self-written exact mixed-union panic diagnostics, exit 70 and empty stdout.",
    "oracle_kind": "self",
    "kills": [
      "M15",
      "M16"
    ],
    "unique_kills": [
      "M15",
      "M16"
    ],
    "last_proven_fail": "M16: view_unions_mixed_test.go:53: got <nil>, stdout \"\", stderr \"\"; want exit 70 and \"adamic: panic: field read failed: view.value matches no member of Left | Right; expected Left | Right, found object\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 17,
    "probe_kills": [
      "PMixed"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASITargetFlags",
      "TestWASIRefusesUnsupportedOptions",
      "TestTSGoBuildSeesTheProgramsFeatures",
      "TestTypedArrayRuntime",
      "TestUnitsPreserveSharedState",
      "TestUnitCacheFlagsHoldSanitizer",
      "TestSplitTokensDoNotRewriteLiterals",
      "TestUnitSystemHeaderProvenance",
      "TestViewMixedUnionUnknownAndUnavailable"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/M16 ADAMIC_MUTANT=M16 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestWASITargetFlags|TestWASIRefusesUnsupportedOptions|TestTSGoBuildSeesTheProgramsFeatures|TestTypedArrayRuntime|TestUnitsPreserveSharedState|TestUnitCacheFlagsHoldSanitizer|TestSplitTokensDoNotRewriteLiterals|TestUnitSystemHeaderProvenance|TestViewMixedUnionUnknownAndUnavailable)$ > M16.log 2>&1; M16: view_unions_mixed_test.go:53: got <nil>, stdout \"\", stderr \"\"; want exit 70 and \"adamic: panic: field read failed: view.value matches no member of Left | Right; expected Left | Right, found object\\n\""
  },
  {
    "test": "TestMeasureClangUnits",
    "package": "internal/native",
    "file": "internal/native/units_measure_test.go:14",
    "seconds": null,
    "oracle": "Self-written split/header invariants and exactly-one-changed-unit requirement; timing outputs measure the compiler.",
    "oracle_kind": "self",
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
    "evidence": "bounded-baseline.log: set ADAMIC_CLANG_MEASURE to emitted C evidence directory",
    "reason": "Required emitted-C benchmark dataset was not configured; repository search found none; no enabled run."
  },
  {
    "test": "TestSplitTSGoAgreesUnit family",
    "package": "internal/native",
    "file": "internal/native/units_tsgo_test.go:102",
    "seconds": null,
    "oracle": "Whole Adamic checker-archive product is the expected-output snapshot for the split product; no independent Go cohere comparison.",
    "oracle_kind": "self",
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
    "evidence": "tsgo-archive-build.log and tsgo-archive-build.json",
    "members": [
      "TestSplitTSGoAgreesUnit00",
      "TestSplitTSGoAgreesUnit01"
    ],
    "reason": "Checker archive build exceeded 90 s; no enabled run."
  },
  {
    "test": "TestWASIUnit family",
    "package": "internal/native",
    "file": "internal/native/wasm_test.go:282",
    "seconds": 11.556,
    "oracle": "Node source execution compared with WASI stdout, stderr and exit; request host additionally checks self-written memory/counter and stack-margin invariants.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M17"
    ],
    "unique_kills": [
      "M17"
    ],
    "last_proven_fail": "M17: wasm_test.go:378: WASI oracle: 0/35 equivalent",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [
      "PC"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestWASIUnit family"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u054/cache/M17 WASI_SYSROOT=/tmp/u054/sdk/share/wasi-sysroot ADAMIC_TEST_WASI=1 PATH=/tmp/u054/sdk/bin:$PATH timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^TestWASIUnit[0-9]+$ > M17-wasi-standalone.log 2>&1; M17: wasm_test.go:378: WASI oracle: 0/35 equivalent",
    "members": [
      "TestWASIUnit00",
      "TestWASIUnit01",
      "TestWASIUnit02",
      "TestWASIUnit03",
      "TestWASIUnit04",
      "TestWASIUnit05",
      "TestWASIUnit06",
      "TestWASIUnit07",
      "TestWASIUnit08",
      "TestWASIUnit09",
      "TestWASIUnit10",
      "TestWASIUnit11",
      "TestWASIUnit12",
      "TestWASIUnit13",
      "TestWASIUnit14",
      "TestWASIUnit15",
      "TestWASIUnit16",
      "TestWASIUnit17",
      "TestWASIUnit18",
      "TestWASIUnit19",
      "TestWASIUnit20",
      "TestWASIUnit21",
      "TestWASIUnit22",
      "TestWASIUnit23",
      "TestWASIUnit24",
      "TestWASIUnit25",
      "TestWASIUnit26",
      "TestWASIUnit27",
      "TestWASIUnit28",
      "TestWASIUnit29",
      "TestWASIUnit30",
      "TestWASIUnit31",
      "TestWASIUnit32",
      "TestWASIUnit33",
      "TestWASIUnit34",
      "TestWASIUnit35"
    ]
  }
]
```

| ID | origin/main file:line | Change | Failing audit rows |
|---|---|---|---|
| M01 | internal/native/native.go:106 | "-O2" -> "-O0" | TestWASITargetFlags |
| M02 | internal/native/native.go:86 | "-ffp-contract=off" -> "-ffp-contract=fast" | TestWASITargetFlags |
| M03 | internal/native/native.go:92 | "-mno-atomics" -> "-matomics" | TestWASITargetFlags |
| M04 | internal/native/target.go:11 | options.Request && options.Target != "wasm32-wasi" -> options.Request && options.Target == "wasm32-wasi" | TestWASIRefusesUnsupportedOptions |
| M05 | internal/native/target.go:20 | if options.Sanitize { -> if !options.Sanitize { | TestWASIRefusesUnsupportedOptions |
| M06 | internal/native/target.go:23 | if options.cpu != "" { -> if options.cpu == "" { | TestWASIRefusesUnsupportedOptions |
| M07 | internal/native/target.go:27 | if sysroot == "" { -> if sysroot != "" { | TestWASIRefusesUnsupportedOptions |
| M08 | internal/native/units.go:186 | names[d.name] = "adamic_unit_" + d.name -> names[d.name] = "adamic_mutant_" + d.name | TestSplitTokensDoNotRewriteLiterals |
| M09 | internal/native/units.go:27 | if strings.ContainsRune(" \t\r\n", rune(c)) { -> if !strings.ContainsRune(" \t\r\n", rune(c)) { | TestSplitTokensDoNotRewriteLiterals, TestUnitsPreserveSharedState |
| M10 | internal/native/units.go:300 | "-E", unit.name -> "-EP", unit.name | TestUnitCacheFlagsHoldSanitizer, TestUnitSystemHeaderProvenance, TestUnitsPreserveSharedState |
| M11 | internal/native/units.go:281 | append([]string{"adamic-units-v2"}, flags...) -> []string{"adamic-units-v2"} | TestUnitCacheFlagsHoldSanitizer |
| M12 | internal/native/units.go:244 | state.WriteString(rewrite(tokens) + "\n") -> (drop statement) | TestUnitsPreserveSharedState |
| M13 | internal/native/runtime/typed_array.c:68 | ? 256.0 : 4294967296.0 -> ? 255.0 : 4294967296.0 | TestTypedArrayRuntime |
| M14 | internal/native/runtime/typed_array.c:88 | index < (double)array->length -> index <= (double)array->length | TestTypedArrayRuntime |
| M15 | internal/native/runtime/view_unions_mixed.c:41 | return "unsupported representation"; -> return "unknown"; | TestViewMixedUnionUnknownAndUnavailable |
| M16 | internal/native/runtime/view_unions_mixed.c:64 | member->contract == 0 \|\| match == NULL -> member->contract != 0 \|\| match == NULL | TestViewMixedUnionUnknownAndUnavailable |
| M17 | internal/native/runtime/weak.c:153 | ((adamic_weak *)reveal(*entry))->target = 0; -> (drop statement) | TestWASIUnit family |

Survivors: none among valid activated production runs.

The supplied reference 8de93800f4 differs from the fetched starting commit. All requested tests remain in their named files; source references and diffs use the actual starting commit. The 48 names group into the advertised 12 rows: the two SplitTSGo wrappers are one family, and the 36 WASI wrappers are one family. WASI's common checker also runs the request-host assertions in its last unit; those extra checks are explicitly included in its oracle description.

The whole-package baseline exceeded the binary's 90-second budget, but had no individual test failure. It was narrowed to the named rows. The default bounded baseline passed in 1.480 seconds and measured 35.0% Go statement coverage. The enabled WASI baseline passed in 30.596 seconds. Every enabled row then passed three clean isolated runs. Kills outside the bounded matrix are unknown, including other callers of Flags, splitC and the runtime. Sacred and unique_kills mean uniqueness in the bounded observations, not proven package or repository uniqueness.

The warm environment lacked a WASI SDK. SDK 27 was installed under /tmp/u054/sdk and used only for separate WASI runs; the shared env.sh and default native clang were not changed. The checker archive build reached 90 seconds and was stopped; no archive was produced or used. TestMeasureClangUnits needs an emitted-C benchmark dataset, not just an installable compiler. No matching dataset was found in the repository, and I did not invent a replacement benchmark. Those two rows remain cannot-judge, with null costs and vacuity. Initial skips, enabled outcomes and all other whole-package skips are in baseline-summary.json.

WASI's host passes env:{} to its module. This prevents a C getenv-based mutation switch from activating there. The first M17 WASI run is marked invalid, excluded from kills and survivor claims. M17 was rerun as its unconditional standalone diff, without editing the host or oracle. That run failed the weak-parent and two reuse-weak fixtures. It has a separate cache environment and build/run timing. All other runtime mutants and probes use native subprocesses that inherit the selector. PC is a Go emitter probe and activates in the compiler process before the WASI module exists.

The Go matrix covered nine default-enabled rows for every production mutant; M17 additionally ran the WASI family. The WASI harness uses its own fixed clang flags and links unsplit emitted C, so its runtime test does not directly call Flags, ValidateOptions, splitC or compileUnit. The relevant source/caller evidence is retained. C runtime coverage is a conservative source inventory, not measured coverage of native or WASI processes.

TSGoFeatures deliberately checks that removing features makes compilation fail. It is judged as a witness: W01 disabled its compile check and the missing-feature assertion failed. Its production-mutant precondition failures do not count. UnitCacheFlagsHoldSanitizer checks the production cache via a signed-overflow program; M11 removed flags from the key and the runtime sanitizer observation failed. Header provenance is subsumed by that row on one observed mutant only, which is a small-matrix hint rather than a deletion recommendation.

PSplit panicked in the literal-inspection row after returning no units. All nine default rows were rerun individually; their observed results replace the aborted run. Probe failures never count as production kills. Typed-array check_write's empty probe permits the normal-output comparisons and is caught by invalid-write probes; those positive portions are listed in vacuous_subcases. Range-error probes only require exit 70 and a panic prefix, so a different panic could satisfy them. No whole enabled row passed its own entry probe.

Warm setup script: 0 seconds; nproc=5. npm ci: 0.413s. SDK installation: 30.477s. Checker archive attempt: 90.007s, timeout status 124. Whole-package build/run wall: 91.737s. Thirty clean timing invocations totaled 92.453s wall. Matrix, witness and probe invocations totaled 245.494s wall. Per-diff validation times and exact commands are in validation.json and probe-validation.json. Native C switch products were built once and selected at execution; compiler flag variants rebuilt their needed products. Build-only times were not separately isolated from test overhead. Per-production run totals: M01: 18.875s wall/13.209s binary; M02: 16.543s wall/14.857s binary; M03: 6.021s wall/4.372s binary; M04: 3.106s wall/1.317s binary; M05: 3.002s wall/1.276s binary; M06: 2.952s wall/1.270s binary; M07: 2.890s wall/1.259s binary; M08: 2.965s wall/1.310s binary; M09: 2.633s wall/0.934s binary; M10: 2.375s wall/0.766s binary; M11: 2.968s wall/1.305s binary; M12: 2.793s wall/1.062s binary; M13: 2.583s wall/0.915s binary; M14: 2.537s wall/0.891s binary; M15: 2.957s wall/1.271s binary; M16: 2.946s wall/1.280s binary; M17: 3.001s wall/1.339s binary; M17 WASI invalid: 32.558s wall/30.864s binary; M17 WASI standalone: 37.809s wall/31.801s binary.

No survivors remain among the 17 valid production mutants. Not covered: the timed-out checker archive family, absent real clang benchmark corpus, package-wide or repo-wide uniqueness, and runtime function coverage outside the observed cases. Production and harness sources were restored; no PR or main push.

W01 failure source maps to origin/main tsgo_features_test.go:65; the permitted scratch guard shifts its logged line to 68.
