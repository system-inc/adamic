u047 audited all 14 named rows at b902a0ccc09e97940571388a1634450da3383559; none moved or vanished.
Whole package cooked at 90.096 s; clean bounded slice passed in 11.280 s; nproc 5.
13 production mutants and 28 entry probes; uniqueness and subsumption are bounded to these 14 rows.
Verdicts: 4 subsumed, 7 sacred, 1 witness, 1 untrue, 1 setup-check.
Evidence: test-audit/internal-native-element_borrow, review/test-audit/internal-native-element_borrow/.

```json
[
  {
    "test": "TestNbodyIndexedElementsBorrow",
    "package": "internal/native",
    "file": "internal/native/element_borrow_test.go",
    "seconds": 0.099,
    "oracle": "Self: exactly five indexed declarations, owner=NULL and absence of releases in generated C. Count can admit a different set of five; no native behavior compared.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: element_borrow_test.go:38: want sun and both body/other pairs borrowed, got 0 declarations",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDevirtualizeBorrowDocClaim"
    ],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P1",
      "P2"
    ],
    "subsumer_seconds": 0.065,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; element_borrow_test.go:38: want sun and both body/other pairs borrowed, got 0 declarations",
    "subsumption_mutants": 2
  },
  {
    "test": "TestThrowElementBorrowPlan",
    "package": "internal/native",
    "file": "internal/native/element_borrow_test.go",
    "seconds": 0.068,
    "oracle": "Self: named functions must or must not borrow across throws, catches and message calls.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2",
      "M3"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M3: element_borrow_test.go:84: caughtInside: harmless Error creation prevented borrowing",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; element_borrow_test.go:84: caughtInside: harmless Error creation prevented borrowing"
  },
  {
    "test": "TestCallTargetsElementBorrowPlan",
    "package": "internal/native",
    "file": "internal/native/element_borrow_test.go",
    "seconds": 0.071,
    "oracle": "Self: bounded harmless closure borrows; virtual, unknown closure and writing targets do not.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: element_borrow_test.go:113: bounded harmless closure lost its element borrow",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDevirtualizeBorrowDocClaim"
    ],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.065,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; element_borrow_test.go:113: bounded harmless closure lost its element borrow",
    "subsumption_mutants": 3
  },
  {
    "test": "TestDevirtualizeBorrowDocClaim",
    "package": "internal/native",
    "file": "internal/native/element_borrow_test.go",
    "seconds": 0.065,
    "oracle": "Self: both named virtual and closure length-only targets must permit a borrow; docs are context, not an external authority.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M2",
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: element_borrow_test.go:142: throughClosure: length-only targets prevented borrowing",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCallTargetsElementBorrowPlan"
    ],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.071,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; element_borrow_test.go:142: throughClosure: length-only targets prevented borrowing",
    "subsumption_mutants": 3
  },
  {
    "test": "TestUniformFieldsMatchNode",
    "package": "internal/native",
    "file": "internal/native/fields_test.go",
    "seconds": 0.697,
    "oracle": "Node stdout executed and compared in release and sanitized binaries; self-written emitted-C lookup patterns also require optimization and conservative fallback.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M5",
      "M8"
    ],
    "unique_kills": [
      "M5"
    ],
    "last_proven_fail": "M8: fields_test.go:109: native: clang failed: exit status 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; fields_test.go:109: native: clang failed: exit status 1"
  },
  {
    "test": "TestRuntimeFieldLayoutsAreIncluded",
    "package": "internal/native",
    "file": "internal/native/fields_test.go",
    "seconds": 0.019,
    "oracle": "Self: runtime C shape declarations must be represented by the production field-offset proof. The two production representations agree; no external authority.",
    "oracle_kind": "self",
    "kills": [
      "M7"
    ],
    "unique_kills": [
      "M7"
    ],
    "last_proven_fail": "M7: fields_test.go:168: directory.c: runtime field \"symbolicLink\" at 3 is absent or conflicting in the layout proof",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; fields_test.go:168: directory.c: runtime field \"symbolicLink\" at 3 is absent or conflicting in the layout proof"
  },
  {
    "test": "TestRegexProgramsKeepCheckedFieldReads",
    "package": "internal/native",
    "file": "internal/native/fields_test.go",
    "seconds": 0.514,
    "oracle": "Node stdout executed and compared in release and sanitized binaries; self-written generated-C regex must retain checked field lookup.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M6",
      "M8"
    ],
    "unique_kills": [
      "M6"
    ],
    "last_proven_fail": "M8: fields_test.go:205: native: clang failed: exit status 1",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; fields_test.go:205: native: clang failed: exit status 1"
  },
  {
    "test": "TestOptionalWriteMissingSlotRemainsChecked",
    "package": "internal/native",
    "file": "internal/native/fields_test.go",
    "seconds": 0.165,
    "oracle": "Node executes the missing-property fixture and must print 2\\n2\\n; self-written NotYet type, location and diagnostic require refusal before C.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M13"
    ],
    "unique_kills": [
      "M13"
    ],
    "last_proven_fail": "M13: fields_test.go:241: missing-slot write must be refused before it reaches C, got <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P4"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; fields_test.go:241: missing-slot write must be refused before it reaches C, got <nil>"
  },
  {
    "test": "TestFreedValuesAreCaughtWithSlabs",
    "package": "internal/native",
    "file": "internal/native/heap_test.go",
    "seconds": 0.596,
    "oracle": "AddressSanitizer executes deliberately freed runtime strings; requires heap-use-after-free or use-after-poison diagnostics plus failing exit, and a clean untouched control. Classified as a planted-failure witness; production M9 failures excluded from its kills.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: heap_test.go:56: malloc, retain after the free: want AddressSanitizer's heap-use-after-free, got <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P5"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^TestFreedValuesAreCaughtWithSlabs$'; heap_test.go:56: malloc, retain after the free: want AddressSanitizer's heap-use-after-free, got <nil>",
    "vacuous_subcases": [
      "malloc/none",
      "slabs/none"
    ]
  },
  {
    "test": "TestSizeClassesShareTheirChunks",
    "package": "internal/native",
    "file": "internal/native/heap_test.go",
    "seconds": 0.155,
    "oracle": "Self: release harness validates string bytes and compares churn/control RSS <=1.5. M10 inflates both, so the relative oracle passes despite lost spare reuse.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P6"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "ADAMIC_MUTANT=M10 go test -json -count=1 -timeout 90s ./internal/native/ -run \"$(cat scope.regex)\"; TestSizeClassesShareTheirChunks passed despite RSS inflation. See survivor logs."
  },
  {
    "test": "TestResidentSetUnits",
    "package": "internal/native",
    "file": "internal/native/heap_test.go",
    "seconds": 0.021,
    "oracle": "Self: hand-written Linux and macOS getrusage-unit cases for a test-local normalization helper; no outside value checked.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: heap_test.go:146: darwin: got 8388 KiB, want 8192",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "go test -json -count=1 -timeout 90s ./internal/native/ -run '^TestResidentSetUnits$'; heap_test.go:146: darwin: got 8388 KiB, want 8192"
  },
  {
    "test": "TestIeee754MatchesNodeBitForBit",
    "package": "internal/native",
    "file": "internal/native/ieee754_test.go",
    "seconds": 5.162,
    "oracle": "Node Math runs for all 21 entry functions; exact result bits compared, with all NaN payloads treated equal. Baseline 1,281,787 answers, zero mismatches.",
    "oracle_kind": "external-run",
    "kills": [
      "M11"
    ],
    "unique_kills": [
      "M11"
    ],
    "last_proven_fail": "M11: ieee754_test.go:255: sin 1 8000000000000000: native 0000000000000000, Node 8000000000000000",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P_adamic_math_acos",
      "P_adamic_math_acosh",
      "P_adamic_math_asin",
      "P_adamic_math_asinh",
      "P_adamic_math_atan",
      "P_adamic_math_atan2",
      "P_adamic_math_cos",
      "P_adamic_math_exp",
      "P_adamic_math_atanh",
      "P_adamic_math_log",
      "P_adamic_math_log1p",
      "P_adamic_math_log2",
      "P_adamic_math_log10",
      "P_adamic_math_expm1",
      "P_adamic_math_cbrt",
      "P_adamic_math_sin",
      "P_adamic_math_tan",
      "P_adamic_math_cosh",
      "P_adamic_math_sinh",
      "P_adamic_math_tanh",
      "P_adamic_math_hypot"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; ieee754_test.go:255: sin 1 8000000000000000: native 0000000000000000, Node 8000000000000000"
  },
  {
    "test": "TestCEndsInNewline",
    "package": "internal/native",
    "file": "internal/native/library_array_test.go",
    "seconds": 0.06,
    "oracle": "Self: generated C must end in newline for small and >256 KiB programs. Field rows also catch this through clang diagnostics.",
    "oracle_kind": "self",
    "kills": [
      "M8"
    ],
    "unique_kills": [],
    "last_proven_fail": "M8: library_array_test.go:20: generated translation unit lacks a final newline",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRegexProgramsKeepCheckedFieldReads"
    ],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 0.514,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; library_array_test.go:20: generated translation unit lacks a final newline",
    "subsumption_mutants": 1
  },
  {
    "test": "TestLibraryMapSetIteratorResources",
    "package": "internal/native",
    "file": "internal/native/library_map_set_iterator_test.go",
    "seconds": 0.935,
    "oracle": "Self: RSS allowance 4096 KiB, active iterator counts through exit codes, successful numeric total and CPU bound 0.35 s. The count subcase checks only zero exit, so unrelated nonzero failures also fail it.",
    "oracle_kind": "self",
    "kills": [
      "M12"
    ],
    "unique_kills": [
      "M12"
    ],
    "last_proven_fail": "M12: library_map_set_iterator_test.go:90: held exhausted iterator peak 160152 KiB, control 660 KiB, allowance 4096 KiB",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 13,
    "probe_kills": [
      "P7"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestNbodyIndexedElementsBorrow",
      "TestThrowElementBorrowPlan",
      "TestCallTargetsElementBorrowPlan",
      "TestDevirtualizeBorrowDocClaim",
      "TestUniformFieldsMatchNode",
      "TestRuntimeFieldLayoutsAreIncluded",
      "TestRegexProgramsKeepCheckedFieldReads",
      "TestOptionalWriteMissingSlotRemainsChecked",
      "TestFreedValuesAreCaughtWithSlabs",
      "TestSizeClassesShareTheirChunks",
      "TestResidentSetUnits",
      "TestIeee754MatchesNodeBitForBit",
      "TestCEndsInNewline",
      "TestLibraryMapSetIteratorResources"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestNbodyIndexedElementsBorrow|TestThrowElementBorrowPlan|TestCallTargetsElementBorrowPlan|TestDevirtualizeBorrowDocClaim|TestUniformFieldsMatchNode|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestOptionalWriteMissingSlotRemainsChecked|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks|TestResidentSetUnits|TestIeee754MatchesNodeBitForBit|TestCEndsInNewline|TestLibraryMapSetIteratorResources)$'; library_map_set_iterator_test.go:90: held exhausted iterator peak 160152 KiB, control 660 KiB, allowance 4096 KiB"
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M1 | internal/native/element_borrow.go:30 | if function.Closure { -> if !function.Closure { | TestNbodyIndexedElementsBorrow, TestThrowElementBorrowPlan, TestCallTargetsElementBorrowPlan, TestDevirtualizeBorrowDocClaim |
| M2 | internal/native/element_borrow.go:138 | if _, isStore := statement.(ir.SetIndex); isStore { -> if _, isStore := statement.(ir.SetIndex); !isStore { | TestNbodyIndexedElementsBorrow, TestThrowElementBorrowPlan, TestCallTargetsElementBorrowPlan, TestDevirtualizeBorrowDocClaim |
| M3 | internal/native/element_borrow.go:158 | case ir.ObjectLiteral, ir.ArrayPush, ir.MakeClosure, ir.MakeError, ir.Defined: 		return true -> case ir.ObjectLiteral, ir.ArrayPush, ir.MakeClosure, ir.Defined: 		return true | TestThrowElementBorrowPlan |
| M4 | internal/native/element_borrow.go:176 | if targets.Unknown { -> if !targets.Unknown { | TestCallTargetsElementBorrowPlan, TestDevirtualizeBorrowDocClaim |
| M5 | internal/native/fields.go:69 | offsets[field.Name] = -1 -> offsets[field.Name] = 0 | TestUniformFieldsMatchNode |
| M6 | internal/native/fields.go:37 | if len(program.Regexps) != 0 \|\| recordStorage { -> if recordStorage { | TestRegexProgramsKeepCheckedFieldReads |
| M7 | internal/native/fields.go:46 | "symbolicLink": 3 -> "symbolicLink": 2 | TestRuntimeFieldLayoutsAreIncluded |
| M8 | internal/native/emit.go:111 | bodies.WriteString("\treturn 0;\n}\n") -> bodies.WriteString("\treturn 0;\n}") | TestUniformFieldsMatchNode, TestRegexProgramsKeepCheckedFieldReads, TestCEndsInNewline |
| M9 | internal/native/runtime/heap.c:172 | 	POISON(slot, size); -> 	(void)size; | TestFreedValuesAreCaughtWithSlabs |
| M10 | internal/native/runtime/heap.c:108 | chunk *each = spares; -> chunk *each = NULL; |  |
| M11 | internal/native/runtime/ieee754.c:2483 | return kernel_sin(x, z, 0); -> return kernel_sin(z, x, 0); | TestIeee754MatchesNodeBitForBit |
| M12 | internal/native/runtime/map.c:267 | 	iterator->map->iterating--; -> 	/* dropped exhausted iteration decrement */ | TestLibraryMapSetIteratorResources |
| M13 | internal/lower/class.go:464 | if l.omittedOptionals[member] && !presentBeforeWrite(target) && !l.literalGivesField(target) { 		return l.notYet(target, "writing a possibly absent optional own field") 	} ->  | TestOptionalWriteMissingSlotRemainsChecked |

Survivor M10: spare reuse disabled. Isolated clean churn/control 6280/5936 KiB becomes 49948/75816 KiB. Both runs pass. The relative gate fails to detect growth shared by its control. This is demonstrated changed behavior, not an equivalent candidate.

Brief ambiguities and costs:

- The supplied historical SHA differs from current origin/main; the start rule wins. All 14 names still exist in the same six supplied files.
- Full native package exceeds the 90 s budget. Only these 14 rows have an observed complete baseline and matrix. No full-package uniqueness is asserted. The whole-package baseline skipped 41 opt-in or missing-SDK rows outside this unit; exact names are in baseline-skips.json. No scoped row skips or needs an opt-in.
- The suggested three mutants per row conflicts with the 20-mutant ceiling for 14 rows. I fixed 13 spread across the reached code before outcomes. The allocator row rests on one honest production attempt, M10, which it missed.
- The at-most-four rebuild fallback concerns native cold builds. Four C production mutants share one environment-selected source and content-keyed runtime archive across selectors. Go selectors compile once; compiler product caches differ per selector. The runtime cache uses source bytes, compiler and flags, not ADAMIC_BUILD_CACHE_DIR, as library.go shows. No stale products are reused across differing runtime source.
- The allocation safety row deliberately plants use-after-free to prove sanitizer visibility. I classify it as witness under the brief, with W1 disabling its instrumentation. Its M9 production failure remains in the matrix but never its kills or unique_kills. Disabling sanitizer in this special witness run is a permitted weakened-check harness edit, restored afterward.
- The normalization helper is defined in heap_test.go. It is suite construction, so S1 changes its conversion divisor under the explicit setup-check exception. Production code cannot reach it.
- Early-entry Go probe diffs need whole-body replacement and removal of newly unused imports to pass go vet. Their scratch selectors return at entry; saved standalone probes remove the unreachable body.
- The empty Lower probe panics in rows using it only as preparation. Individually rerun rows lost to the abort; only the refusal row is judged by P4. Runtime-entry probes are separate for concat, allocation, map construction and every one of the 21 Math functions. The setup helper was not empty-probed, so its vacuous value is null.
- The newline row shares its kill with strict clang builds in field rows. It is subsumed on one observed production mutant, not proposed for deletion.
- The resource oracle for shared chunks admits a wrong result because the same mutant worsens its control. Its survivor witness is the clearest finding. Iterator count checks observe exit codes rather than the exact reason for failure.
- C function reachability is a conservative source inventory, not measured C coverage. Native Go reached functions are backed by scope.cover. Full transitive lower and runtime coverage was not obtained. See reached-functions.txt.
- I corrected an initially prepared direct test-binary launch before retaining results: Go tests require their package working directory. The retained matrix uses the required go test -json command. Repeated validation and the corrected build costs remain in commands.jsonl.
- No external-authority claims are made from comments alone. Node and ASan actually ran; other expectations are self.

Timing and limits:

Warm toolchain setup skipped, setup 0 s, npm ci 0.969 s. Every row ran alone three times; binary medians are in the array and raw runs in timings.json. Exact command wall times and exit statuses are in commands.jsonl. Whole-package binary budget 90.096 s; bounded coverage baseline command 32.297 s, binary 11.280 s. Runtime build time is included in test command time; no separate reliable per-runtime-archive timing was captured. Compilation, vet and clang validation wall times are itemized in commands.jsonl. No other packages tested, no repo-wide replay, no external-authority value cross-check, no exhaustive mutation coverage, no full native baseline completion.
