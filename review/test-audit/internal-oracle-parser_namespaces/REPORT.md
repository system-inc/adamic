Unit u067: nine named tests exist; grouped into seven rows.
Starting origin/main: 6c60da091afddc9c2fe88b3a1067845b6dc79cb3; nproc 5.
Bounded results: three sacred rows, one subsumed family, three witnesses.
Fourteen production mutants: twelve caught, two witnessed survivors; probes separate.
Sources restored; evidence branch test-audit/internal-oracle-parser_namespaces.

```json
[
  {
    "test": "TestParserNamespace family",
    "package": "internal/oracle",
    "file": "internal/oracle/parser_namespaces_test.go:28",
    "seconds": 1.871,
    "oracle": "Untouched source runs on Node; sanitized native and emitted JavaScript must match stdout, stderr and exit, plus sanitizer leak check.",
    "oracle_kind": "external-run",
    "last_proven_fail": "M05 parser_namespaces_test.go:101: /workspace/adamic/internal/oracle/testdata/namespace_callable_properties.a:10:1: stage 0 can't lower observing a callable namespace object; only direct calls, fixed qualified members, typeof and canonical function identity are represented yet",
    "verdict": "subsumed",
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=M05 ADAMIC_BUILD_CACHE_DIR=/tmp/u067/cache/M05 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestParserNamespaceReceiver|TestParserNamespaceClass|TestParserCallableNamespace|TestPredicateDirectionCountsAreRecorded|TestPredicateMiscompileRefusals|TestInUnionNarrowingIsRefused)$'; parser_namespaces_test.go:101: /workspace/adamic/internal/oracle/testdata/namespace_callable_properties.a:10:1: stage 0 can't lower observing a callable namespace object; only direct calls, fixed qualified members, typeof and canonical function identity are represented yet",
    "kills": [
      "M01",
      "M03",
      "M04",
      "M05"
    ],
    "unique_kills": [],
    "probe_kills": [
      "P01",
      "P02"
    ],
    "vacuous": false,
    "subsumed_by": [
      "TestNativeAgreesWithNode family (eight selected inputs)"
    ],
    "subsumer_seconds": 4.978,
    "mutants_in_matrix": 14,
    "matrix_rows": [
      "TestParserNamespace family",
      "TestPredicateDirectionCountsAreRecorded",
      "TestPredicateMiscompileRefusals",
      "TestInUnionNarrowingIsRefused",
      "TestNativeAgreesWithNode family (eight selected inputs)"
    ],
    "empty_answer_probe": "P02",
    "members": [
      "TestParserNamespaceReceiver",
      "TestParserNamespaceClass",
      "TestParserCallableNamespace"
    ],
    "subsumption_mutants": 4
  },
  {
    "test": "TestParserCallableNamespaceMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/parser_namespaces_test.go:43",
    "seconds": 0.716,
    "oracle": "Node observation and the agreement checker applied to a built-in IR mutation. Only W01 comparison weakening determines the verdict.",
    "oracle_kind": "external-run",
    "last_proven_fail": "W01 parser_namespaces_test.go:83: native attached function mutant survived: {stdout:[108 111 103 58 102 105 114 115 116 10 108 111 103 58 115 101 99 111 110 100 10 108 111 103 58 116 104 105 114 100 10 116 114 117 101 58 116 114 117 101 58 116 114 117 101 58 102 117 110 99 116 105 111 110 58 51 10 112 114 105 110 116 58 102 111 117 114 116 104 10 55 58 116 114 117 101 10] stderr:[] exitCode:0}",
    "verdict": "witness",
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=W01 ADAMIC_BUILD_CACHE_DIR=/tmp/u067/cache/W01 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^TestParserCallableNamespaceMutant$'; parser_namespaces_test.go:83: native attached function mutant survived: {stdout:[108 111 103 58 102 105 114 115 116 10 108 111 103 58 115 101 99 111 110 100 10 108 111 103 58 116 104 105 114 100 10 116 114 117 101 58 116 114 117 101 58 116 114 117 101 58 102 117 110 99 116 105 111 110 58 51 10 112 114 105 110 116 58 102 111 117 114 116 104 10 55 58 116 114 117 101 10] stderr:[] exitCode:0}",
    "kills": [],
    "unique_kills": [],
    "probe_kills": [],
    "vacuous": null,
    "subsumed_by": [],
    "subsumer_seconds": null,
    "mutants_in_matrix": 0,
    "matrix_rows": [
      "TestParserCallableNamespaceMutant",
      "TestParserNamespaceClassRegistrationMutant",
      "TestParserNamespaceReceiverMutant"
    ],
    "witness_edits": [
      "W01"
    ]
  },
  {
    "test": "TestParserNamespaceClassRegistrationMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/parser_namespaces_test.go:117",
    "seconds": 0.645,
    "oracle": "Node observation and the agreement checker applied to a built-in IR mutation. Only W01 comparison weakening determines the verdict.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "last_proven_fail": "W01 parser_namespaces_test.go:144: native missing registration mutant survived: {stdout:[] stderr:[97 100 97 109 105 99 58 32 112 97 110 105 99 58 32 82 101 102 101 114 101 110 99 101 69 114 114 111 114 58 32 67 97 110 110 111 116 32 97 99 99 101 115 115 32 39 68 101 98 117 103 84 121 112 101 77 97 112 112 101 114 39 32 98 101 102 111 114 101 32 105 110 105 116 105 97 108 105 122 97 116 105 111 110 10] exitCode:70}",
    "verdict": "witness",
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=W01 ADAMIC_BUILD_CACHE_DIR=/tmp/u067/cache/W01 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^TestParserNamespaceClassRegistrationMutant$'; parser_namespaces_test.go:144: native missing registration mutant survived: {stdout:[] stderr:[97 100 97 109 105 99 58 32 112 97 110 105 99 58 32 82 101 102 101 114 101 110 99 101 69 114 114 111 114 58 32 67 97 110 110 111 116 32 97 99 99 101 115 115 32 39 68 101 98 117 103 84 121 112 101 77 97 112 112 101 114 39 32 98 101 102 111 114 101 32 105 110 105 116 105 97 108 105 122 97 116 105 111 110 10] exitCode:70}",
    "kills": [],
    "unique_kills": [],
    "probe_kills": [],
    "vacuous": null,
    "subsumed_by": [],
    "subsumer_seconds": null,
    "mutants_in_matrix": 0,
    "matrix_rows": [
      "TestParserCallableNamespaceMutant",
      "TestParserNamespaceClassRegistrationMutant",
      "TestParserNamespaceReceiverMutant"
    ],
    "witness_edits": [
      "W01"
    ]
  },
  {
    "test": "TestParserNamespaceReceiverMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/parser_namespaces_test.go:150",
    "seconds": 0.607,
    "oracle": "Node observation and the agreement checker applied to a built-in IR mutation. Only W01 comparison weakening determines the verdict.",
    "oracle_kind": "external-run",
    "last_proven_fail": "W01 parser_namespaces_test.go:175: native wrong receiver mutant survived: {stdout:[57 57 58 57 57 10 57 57 10] stderr:[] exitCode:0}",
    "verdict": "witness",
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=W01 ADAMIC_BUILD_CACHE_DIR=/tmp/u067/cache/W01 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^TestParserNamespaceReceiverMutant$'; parser_namespaces_test.go:175: native wrong receiver mutant survived: {stdout:[57 57 58 57 57 10 57 57 10] stderr:[] exitCode:0}",
    "kills": [],
    "unique_kills": [],
    "probe_kills": [],
    "vacuous": null,
    "subsumed_by": [],
    "subsumer_seconds": null,
    "mutants_in_matrix": 0,
    "matrix_rows": [
      "TestParserCallableNamespaceMutant",
      "TestParserNamespaceClassRegistrationMutant",
      "TestParserNamespaceReceiverMutant"
    ],
    "witness_edits": [
      "W01"
    ]
  },
  {
    "test": "TestPredicateDirectionCountsAreRecorded",
    "package": "internal/oracle",
    "file": "internal/oracle/predicate_counts_test.go:81",
    "seconds": 0.616,
    "oracle": "Own counts.md snapshot plus internal site/direction validity and aggregate consistency. No independent authority.",
    "oracle_kind": "self",
    "last_proven_fail": "M10 predicate_counts_test.go:96: predicate counts changed; update counts.md with the full count writer:",
    "verdict": "sacred",
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u067/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestParserNamespaceReceiver|TestParserNamespaceClass|TestParserCallableNamespace|TestPredicateDirectionCountsAreRecorded|TestPredicateMiscompileRefusals|TestInUnionNarrowingIsRefused)$'; predicate_counts_test.go:96: predicate counts changed; update counts.md with the full count writer:",
    "kills": [
      "M07",
      "M08",
      "M09",
      "M10"
    ],
    "unique_kills": [
      "M08",
      "M09"
    ],
    "probe_kills": [
      "P01",
      "P02"
    ],
    "vacuous": false,
    "subsumed_by": [],
    "subsumer_seconds": null,
    "mutants_in_matrix": 14,
    "matrix_rows": [
      "TestParserNamespace family",
      "TestPredicateDirectionCountsAreRecorded",
      "TestPredicateMiscompileRefusals",
      "TestInUnionNarrowingIsRefused",
      "TestNativeAgreesWithNode family (eight selected inputs)"
    ],
    "empty_answer_probe": "P02"
  },
  {
    "test": "TestPredicateMiscompileRefusals",
    "package": "internal/oracle",
    "file": "internal/oracle/predicate_refusals_test.go:30",
    "seconds": 3.693,
    "oracle": "Untouched source runs on Node against handwritten stdout; handwritten refusal/pending classes and path-bearing diagnostic substrings, or exact backend panic stdout/stderr/exit and aggregate checked counts.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "last_proven_fail": "M14 predicate_refusals_test.go:79: release: want named predicate check, exit=70 stdout=\"claimed box: box a\\ndirect: box\\n\" stderr=\"adamic: panic: overload 1 of isText result: predicate x failed\\n\"",
    "verdict": "sacred",
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u067/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestParserNamespaceReceiver|TestParserNamespaceClass|TestParserCallableNamespace|TestPredicateDirectionCountsAreRecorded|TestPredicateMiscompileRefusals|TestInUnionNarrowingIsRefused)$'; predicate_refusals_test.go:79: release: want named predicate check, exit=70 stdout=\"claimed box: box a\\ndirect: box\\n\" stderr=\"adamic: panic: overload 1 of isText result: predicate x failed\\n\"",
    "kills": [
      "M07",
      "M10",
      "M13",
      "M14"
    ],
    "unique_kills": [
      "M13",
      "M14"
    ],
    "probe_kills": [
      "P01",
      "P02"
    ],
    "vacuous": false,
    "subsumed_by": [],
    "subsumer_seconds": null,
    "mutants_in_matrix": 14,
    "matrix_rows": [
      "TestParserNamespace family",
      "TestPredicateDirectionCountsAreRecorded",
      "TestPredicateMiscompileRefusals",
      "TestInUnionNarrowingIsRefused",
      "TestNativeAgreesWithNode family (eight selected inputs)"
    ],
    "empty_answer_probe": "P02"
  },
  {
    "test": "TestInUnionNarrowingIsRefused",
    "package": "internal/oracle",
    "file": "internal/oracle/presence_refused_test.go:23",
    "seconds": 0.333,
    "oracle": "Untouched source Node exit 0 and empty stderr only, with no stdout assertion; handwritten Refused class and complete own .refused diagnostic snapshot.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "last_proven_fail": "M12 presence_refused_test.go:36: want a refusal, got <nil>",
    "verdict": "sacred",
    "bounded": true,
    "evidence": "ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT=M12 ADAMIC_BUILD_CACHE_DIR=/tmp/u067/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestParserNamespaceReceiver|TestParserNamespaceClass|TestParserCallableNamespace|TestPredicateDirectionCountsAreRecorded|TestPredicateMiscompileRefusals|TestInUnionNarrowingIsRefused)$'; presence_refused_test.go:36: want a refusal, got <nil>",
    "kills": [
      "M11",
      "M12"
    ],
    "unique_kills": [
      "M11",
      "M12"
    ],
    "probe_kills": [
      "P01",
      "P02"
    ],
    "vacuous": false,
    "subsumed_by": [],
    "subsumer_seconds": null,
    "mutants_in_matrix": 14,
    "matrix_rows": [
      "TestParserNamespace family",
      "TestPredicateDirectionCountsAreRecorded",
      "TestPredicateMiscompileRefusals",
      "TestInUnionNarrowingIsRefused",
      "TestNativeAgreesWithNode family (eight selected inputs)"
    ],
    "empty_answer_probe": "P02"
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/lower/namespace_receiver_scope.go:13 | node.Kind != ast.KindArrowFunction -> node.Kind == ast.KindArrowFunction | TestNativeAgreesWithNode family (eight selected inputs), TestParserNamespace family |
| M02 | internal/lower/namespaces.go:426 | return value at entry |  |
| M03 | internal/lower/namespaces.go:502 | body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: true}}) -> body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: false}}) | TestNativeAgreesWithNode family (eight selected inputs), TestParserNamespace family |
| M04 | internal/lower/namespaces.go:94 | if node.Kind == ast.KindElementAccessExpression { 		return nil, true -> if node.Kind == ast.KindPropertyAccessExpression { 		return nil, true | TestNativeAgreesWithNode family (eight selected inputs), TestParserNamespace family |
| M05 | internal/lower/namespace_callable.go:21 | if called(node) { -> if !called(node) { | TestNativeAgreesWithNode family (eight selected inputs), TestParserNamespace family |
| M06 | internal/lower/lower.go:43 | if err := lowering.namespaceInitialization(modules); err != nil { 		return nil, err 	}  -> |  |
| M07 | internal/lower/predicates_proof.go:1222 | counts.Checked++  -> | TestPredicateDirectionCountsAreRecorded, TestPredicateMiscompileRefusals |
| M08 | internal/lower/predicates_proof.go:1224 | "unobservable", "no narrowed read in the " -> "proven", "no narrowed read in the " | TestPredicateDirectionCountsAreRecorded |
| M09 | internal/lower/predicates_proof.go:1211 | name := "true" -> name := "false" | TestPredicateDirectionCountsAreRecorded |
| M10 | internal/lower/predicates_proof.go:703 | implementation.Body().ForEachChild(writes) 	if changed { -> implementation.Body().ForEachChild(writes) 	if !changed { | TestPredicateDirectionCountsAreRecorded, TestPredicateMiscompileRefusals |
| M11 | internal/lower/unknown.go:45 | "use a discriminant, or a Map" -> "use a discriminant" | TestInUnionNarrowingIsRefused |
| M12 | internal/lower/unknown.go:44 | objects > 1 -> objects > 2 | TestInUnionNarrowingIsRefused |
| M13 | internal/lower/predicates.go:37 | "a type predicate whose return is not proven (" -> "a predicate whose return is not proven (" | TestPredicateMiscompileRefusals |
| M14 | internal/lower/predicates_proof.go:753 | "overload %d of %s result: predicate %s is false" -> "overload %d of %s result: predicate %s failed" | TestPredicateMiscompileRefusals |

Survivors

M02: namespaceReadyValue returns the raw value. Early closure read changes Node/baseline TypeError to ReferenceError; both exit 70. See M02.independent.* and independent-commands.json.
M06: dropping namespaceInitialization changes reaching_direct.a from compile refusal (exit 1) to accepted code (exit 0), then runtime TypeError (exit 70). See M06.independent.* and independent-commands.json.

Brief feedback, unit u067

1. The required fresh fetch selects 6c60da091afddc9c2fe88b3a1067845b6dc79cb3, not historical 8de93800f4. All nine named functions remain in the supplied files.
2. Nine functions become seven rows. The three ordinary parser wrappers share parserNamespaceMatchesNode and differ only in fixture inputs, so they are one family. The three built-in mutant tests are distinct witnesses, not production defenders. This changes the appropriate mutation budget.
3. The whole package times out after 90.326 binary seconds, with no individual failure or error diagnostic before the timeout. The big-package exception was required. The bounded baseline, all isolated rows and restored control pass uncached.
4. A direct grep for a lowering function in oracle tests misses the lowered -> Lower call chain. Selected lowering coverage lists 489 reached functions, recorded before planting mutants. Fixture registrations also establish the selected shared-family paths.
5. The shared differential family is too large for full replay inside this audit. I selected its six namespace fixtures and two admitted predicate-refusal fixtures as one partial family. Its other members remain unknown. No package or repository uniqueness is claimed.
6. The parser family is subsumed by that partial differential family on four caught mutants. The subsumer median is 4.978 seconds versus 1.871 for the parser family. The present brief permits this slower subsumer; this is a four-mutant hint, not a deletion recommendation.
7. Witnesses need comparison weakening, not production mutants. W01 makes disagreement always return empty and all three witnesses fail. Their production failures would only demonstrate broken preconditions, so the production matrix excludes them.
8. I ran both a nil Lower probe and a zero-content IR probe. Nil can fail by a harness dereference rather than a useful-output assertion. All four production rows also fail the empty-IR probe, so none is vacuous in this run. Probes and W01 are never counted as production kills.
9. InUnionNarrowing checks the untouched Node source only for zero exit and empty stderr, not stdout. Its independent lowering assertion pins the full refusal snapshot. These are different oracle strengths and must be named separately.
10. PredicateMiscompileRefusals has a PREDICATE_MISCOMPILE_BASELINE bypass that only logs outcomes. I explicitly unset it for timings, coverage, matrix and restored control. A baseline run with that flag would not be an equivalent audit.
11. PredicateDirectionCountsAreRecorded has an update-counts bypass and compares to the project's own snapshot. Default update-counts=false was preserved. It also validates internal direction metadata and aggregate consistency; that is self, not an outside authority.
12. M02 survives normal fixtures but changes an unresolved early read from Node's TypeError to a ReferenceError. M06 survives them but changes a pre-initialization read from compile refusal to accepted code that panics at runtime. Normal fixture agreement cannot settle these initialization boundaries.
13. M10 preserves exit 70 in a rebound-parameter case but changes the reason to a different union guard. The exact named-predicate panic assertion catches it. Exit-code-only checking would have accepted the wrong reason.
14. The aborted whole baseline has outside-unit opt-in skips. All requested rows ran. The brief is unclear whether tooling for skipped rows outside the bounded unit must be installed; this audit did not expand its unit to them.
15. My first switch instrumentation scoped the M09 local variable inside two branches, which would not compile. I corrected the instrumentation before any matrix run; the standalone mutant itself was already valid. The saved plan generator reconstructs the corrected switch.
16. My running matrix script matched any panic: substring, including expected adamic runtime panic output. It unnecessarily repeated the four production rows individually. These were not Go panics. Original logs/commands are preserved, actual Go-panic status is recorded in matrix.json, and matrix-replay.py corrects detection for central replay. Timings separately report the wasted repeat cost.
17. Test-binary medians exclude Go compilation. Command wall time includes compilation, native tools and the unnecessary repeats. Native builds inside oracle tests cannot be attributed as a separate pure build phase without more instrumentation; no such number is invented.

Timing

```json
{
  "setup_seconds": 0,
  "nproc": 5,
  "whole_baseline_binary_seconds": 90.326,
  "bounded_baseline_binary_seconds": 14.337,
  "timing_and_coverage_command_seconds": 131.74158868700397,
  "matrix_controls_probes_witness_command_seconds": 522.0478173539923,
  "matrix_controls_probes_witness_binary_seconds": 212.64000000000001,
  "unnecessary_isolated_repeat_command_seconds": 254.78576337998675,
  "standalone_vet_seconds": 33.48618071399687,
  "independent_compiler_build_seconds": 16.523721895999188,
  "restored_control_command_seconds": 7.85556710200035,
  "restored_control_exit": 0,
  "npm_reported_seconds": 0.763
}
```

Not covered: full package after its 90-second timeout, shared differential-family inputs beyond the eight listed, other packages, repository uniqueness, and opt-in tooling outside this unit. Requested rows had no skips. Outside-unit baseline skips are in skipped-rows.json.
