u152: 11 discovered functions, nine rows after family grouping; none moved or vanished.
Base: 3bf0a5d9e74d197a38f61982e43b2d791f17b563; nproc 5; warm setup skipped.
Enabled slice baseline passed in 50.595s; restored slice passed in 52.186s; no slice skips.
Bounded verdicts: one sacred family, four subsumed rows, one overlapping row, three witnesses.
Evidence branch: test-audit/stage1-cohere-yaml-gaps; path review/test-audit/stage1-cohere-yaml-gaps/.

```json
[
  {
    "test": "TestLexerGaps",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/gaps_test.go:18",
    "seconds": 0.114,
    "oracle": "Node output is pinned to 2\\n; NotYet type and diagnostic substring are self-written compiler-limit pins.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1 gaps_test.go:44: gap changed or closed: <nil>; update GAPS.md and remove its workaround",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestStructuralPositionRefusal"
    ],
    "mutants_in_matrix": 4,
    "completed_production_mutants": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.112,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLexerGaps",
      "TestStructuralPositionRefusal",
      "TestClosed family",
      "TestLexerMatchesGo",
      "TestPropsMatchGo",
      "TestSharedSliceAppendMatchesNode"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/M1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run ^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestPropsMatchGo|TestSharedSliceAppendMatchesNode)$ > M1.log 2>&1; gaps_test.go:44: gap changed or closed: <nil>; update GAPS.md and remove its workaround",
    "nproc": 5,
    "unknown_mutants": []
  },
  {
    "test": "TestStructuralPositionRefusal",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/gaps_test.go:51",
    "seconds": 0.112,
    "oracle": "Node output is pinned to 1\\n; Refused type and Span[] diagnostic substring are self-written refusal pins. M1 is rejected for returning a different diagnostic class.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1 gaps_test.go:71: refusal changed or closed: /workspace/adamic/stage1/cohere/yaml/gaps/structuralPosition.ts:11:37: stage 0 can't lower push with other than one value yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestLexerGaps"
    ],
    "mutants_in_matrix": 4,
    "completed_production_mutants": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.114,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLexerGaps",
      "TestStructuralPositionRefusal",
      "TestClosed family",
      "TestLexerMatchesGo",
      "TestPropsMatchGo",
      "TestSharedSliceAppendMatchesNode"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/M1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run ^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestPropsMatchGo|TestSharedSliceAppendMatchesNode)$ > M1.log 2>&1; gaps_test.go:71: refusal changed or closed: /workspace/adamic/stage1/cohere/yaml/gaps/structuralPosition.ts:11:37: stage 0 can't lower push with other than one value yet",
    "nproc": 5,
    "unknown_mutants": []
  },
  {
    "test": "TestClosed family",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/gaps_test.go:77",
    "seconds": 2.851,
    "oracle": "Live Node stdout, checked against hand-written fixture pins, is compared byte-for-byte with sanitized native and emitted JavaScript stdout. M2 fails valueConjunction.ts.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2"
    ],
    "unique_kills": [
      "M2"
    ],
    "last_proven_fail": "M2 gaps_test.go:97: native ASan/UBSan/LSan: \"TRUE\\n\", Node \"true\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "completed_production_mutants": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLexerGaps",
      "TestStructuralPositionRefusal",
      "TestClosed family",
      "TestLexerMatchesGo",
      "TestPropsMatchGo",
      "TestSharedSliceAppendMatchesNode"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/M2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run ^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestPropsMatchGo|TestSharedSliceAppendMatchesNode)$ > M2.log 2>&1; gaps_test.go:97: native ASan/UBSan/LSan: \"TRUE\\n\", Node \"true\\n\"",
    "nproc": 5,
    "unknown_mutants": [],
    "members": [
      "TestClosedStringPresenceGap",
      "TestClosedLexerGaps",
      "TestClosedValuePresenceGap"
    ]
  },
  {
    "test": "TestLexerMatchesGo",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/lexer_test.go:180",
    "seconds": 5.877,
    "oracle": "Live Go cohere lexer, Node port source, emitted JavaScript and yaml@2.9.0 exact-byte agreement. The completed M1 kill is a lowering precondition failure; the port mutant M4 times out before comparison.",
    "oracle_kind": "external-run",
    "kills": [
      "M1"
    ],
    "unique_kills": [],
    "last_proven_fail": "M1 lexer_test.go:203: /workspace/adamic/stage1/cohere/yaml/lexer.ts:4:43: stage 0 can't lower push with other than one value yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPropsMatchGo"
    ],
    "mutants_in_matrix": 4,
    "completed_production_mutants": 3,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 9.754,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLexerGaps",
      "TestStructuralPositionRefusal",
      "TestClosed family",
      "TestLexerMatchesGo",
      "TestPropsMatchGo",
      "TestSharedSliceAppendMatchesNode"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/M1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run ^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestPropsMatchGo|TestSharedSliceAppendMatchesNode)$ > M1.log 2>&1; lexer_test.go:203: /workspace/adamic/stage1/cohere/yaml/lexer.ts:4:43: stage 0 can't lower push with other than one value yet",
    "nproc": 5,
    "unknown_mutants": [
      "M4"
    ]
  },
  {
    "test": "TestLexerMutants",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/lexer_test.go:225",
    "seconds": 11.774,
    "oracle": "Witness compares built-in port-mutant native and Node bytes against live Go cohere lexer output; W1 makes the comparison always agree and all five subcases fail.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1 lexer_test.go:264: native missed mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "completed_production_mutants": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestLexerMutants"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run ^TestLexerMutants$ > W1.log 2>&1; lexer_test.go:264: native missed mutant",
    "nproc": 5,
    "unknown_mutants": [],
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestPropsMatchGo",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/props_test.go:78",
    "seconds": 9.754,
    "oracle": "Live Go cohere private resolver through an export overlay, Node port source, emitted JavaScript and yaml@2.9.0 exact-byte agreement. M1 fails lowering; M3 triggers a sanitized native exit before byte comparison.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3 props_test.go:101: /tmp/adamic-gate/TestPropsMatchGo1332184536/005/props: exit status 1; ERROR: AddressSanitizer: heap-buffer-overflow",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestStructuralPositionRefusal",
      "TestSharedSliceAppendMatchesNode"
    ],
    "mutants_in_matrix": 4,
    "completed_production_mutants": 3,
    "probe_kills": [
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLexerGaps",
      "TestStructuralPositionRefusal",
      "TestClosed family",
      "TestLexerMatchesGo",
      "TestPropsMatchGo",
      "TestSharedSliceAppendMatchesNode"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run ^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestPropsMatchGo|TestSharedSliceAppendMatchesNode)$ > M3.log 2>&1; props_test.go:101: /tmp/adamic-gate/TestPropsMatchGo1332184536/005/props: exit status 1; ERROR: AddressSanitizer: heap-buffer-overflow",
    "nproc": 5,
    "unknown_mutants": [
      "M4"
    ]
  },
  {
    "test": "TestPropsMutants",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/props_test.go:128",
    "seconds": 10.448,
    "oracle": "Witness compares built-in port-mutant native and Node bytes against live Go cohere resolver output; W2 makes the comparison always agree and all three subcases fail.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2 props_test.go:178: native missed mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "completed_production_mutants": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestPropsMutants"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/W2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run ^TestPropsMutants$ > W2.log 2>&1; props_test.go:178: native missed mutant",
    "nproc": 5,
    "unknown_mutants": [],
    "witness_kills": [
      "W2"
    ]
  },
  {
    "test": "TestSharedSliceAppendMatchesNode",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/scalar_runtime_gap_test.go:15",
    "seconds": 0.471,
    "oracle": "Live Node stdout pinned to a\\nx\\n, compared byte-for-byte against native and sanitized native at offsets 0 and 48. M3 changes native stdout to a\\n\\n.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3 scalar_runtime_gap_test.go:54: /tmp/adamic-gate/TestSharedSliceAppendMatchesNode2423637761/001/shared: native \"a\\n\\n\" Node \"a\\nx\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestPropsMatchGo"
    ],
    "mutants_in_matrix": 4,
    "completed_production_mutants": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 9.754,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestLexerGaps",
      "TestStructuralPositionRefusal",
      "TestClosed family",
      "TestLexerMatchesGo",
      "TestPropsMatchGo",
      "TestSharedSliceAppendMatchesNode"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run ^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestPropsMatchGo|TestSharedSliceAppendMatchesNode)$ > M3.log 2>&1; scalar_runtime_gap_test.go:54: /tmp/adamic-gate/TestSharedSliceAppendMatchesNode2423637761/001/shared: native \"a\\n\\n\" Node \"a\\nx\\n\"",
    "nproc": 5,
    "unknown_mutants": []
  },
  {
    "test": "TestScalarMutants",
    "package": "stage1/cohere/yaml",
    "file": "stage1/cohere/yaml/scalar_test.go:104",
    "seconds": 13.43,
    "oracle": "Witness compares built-in port-mutant native and Node bytes against live Go cohere scalar output; W3 makes the comparison always agree and all three subcases fail.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W3 scalar_test.go:154 (origin; log line 153): native missed mutant",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "completed_production_mutants": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestScalarMutants"
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/u152/cache/W3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run ^TestScalarMutants$ > W3.log 2>&1; scalar_test.go:154 (origin; log line 153): native missed mutant",
    "nproc": 5,
    "unknown_mutants": [],
    "witness_kills": [
      "W3"
    ]
  }
]
```

| ID | Origin/main file:line | Change | Observed failed rows |
| --- | --- | --- | --- |
| M1 | internal/lower/object.go:1265 | flip len(arguments) != 1 to == 1 in the push arity guard | TestLexerGaps, TestStructuralPositionRefusal, TestLexerMatchesGo, TestPropsMatchGo |
| M2 | internal/native/runtime/string_build_impl.h:131 | change true text constant from true to TRUE | TestClosed family |
| M3 | internal/native/runtime/string_append.c:86 | drop result->length = written | TestPropsMatchGo, TestSharedSliceAppendMatchesNode |
| M4 | stage1/cohere/yaml/lexer.ts:21 | change flow indicator comma code 44 to hyphen code 45 | unknown: LexerMatchesGo and PropsMatchGo over budget |

Witness diffs: W1 lexer_test.go:263, W2 props_test.go:177, W3 scalar_test.go:153. Each comparison changes bytes.Equal(...) to true; W3 removes its unused import.

Empty-entry diffs: P1 internal/lower/lower.go:20; P2 lex_main.ts:25; P3 props_main.ts:99. Probes do not count as mutants.

Survivors: none among completed matrices. M4 is over budget; its Node witness changes a terminating token stream into nontermination.

Observed limits and brief friction:

1. Current origin/main is 3bf0a5d9e74d197a38f61982e43b2d791f17b563, newer than the brief's 8de93800f4. All 11 requested functions remain in their listed files. Discovery is saved; no moved or vanished row.
2. The three TestClosed wrappers differ only in inputs to closedGap. They form one TestClosed family with nine fixture inputs, reducing 11 functions to nine rows. Its three timing runs select all three members together.
3. The first whole-package run used the specified outer 120-second backstop and exited 124 without any observed failed test. TestMain ran a setup child for 41.877 seconds before m.Run, so the outer clock cut off the suite before its own 90-second test clock finished. The initial run skipped two out-of-scope library rows, TestComposeMatchGo and TestCSTMatchesGo. The enabled slice baseline and restored baseline had no skips and no failures. Kills outside the slice remain unknown.
4. ADAMIC_YAML_LIBRARY is required. LexerMatchesGo skips after several substantive comparisons if it is absent; PropsMatchGo skips before comparing any of its gathered outputs. Both were enabled for every slice baseline, timing, mutation and probe using yaml@2.9.0 and prettier@3.9.6 in /tmp/u152/library. API npm ci and the library installation succeeded, but their duration was not instrumented.
5. M4 makes the flow lexer fail to advance on '-x'. The initial matrix timed out at 90.016 seconds. The two reached rows were then run separately and each exhausted 90 seconds. All other production rows were replayed separately and passed. M4 has unknown kills, not zero observed kills, and is neither a survivor nor an equivalent candidate. A two-second Node behavior witness returns the token stream before M4 and exits 124 after M4. It supports changed behavior only, not a test kill.
6. The childguard defaults are 30 minutes to first output, two minutes of stall and 60 minutes total. The Go binary's timeout panics before these guards terminate the hanging native child, which runs in its own process group. Three audit-owned orphan groups required explicit cleanup. The first orphan also competed with the witness runs. The three-run Good timings happened before any mutation or orphan and are unaffected.
7. The completed M1 kills for the lexer and property agreement tests are lowering precondition failures. M3's property kill is an ASan heap-buffer-overflow before the equality assertion. These are real integration catches; they do not demonstrate a completed nonempty port-output disagreement. P2 and P3 separately demonstrate rejection of empty driver output, but probes are excluded from worthiness and subsumption.
8. The two negative gap guards pin self-written diagnostic classes and text while Node proves the inputs execute. M1 makes the structural refusal become a NotYet error and the structural row rejects it. These oracle pins will also fail if the compiler intentionally closes the limitations, as their messages explain.
9. Subsumption rests on three completed production matrices: one shared M1 kill for the gap guards and lexer row, and two shared kills for the property row. This is evidence for a defender's review, not a deletion recommendation. M4 cannot settle any additional lexer/property relation. Sacred means exclusive within these six production rows only; central replay must settle wider uniqueness.
10. The four-mutant rebuild limit takes precedence over aiming for three mutations per row. The mutations were fixed before outcomes and spread over an arity guard, a runtime constant, runtime string append and a port delimiter function. They were replayed individually with fresh per-id ADAMIC_BUILD_CACHE_DIR values, rather than a source switch, so each standalone diff is exactly the version run. Native Build also keys its runtime library on source contents; it rebuilt changed runtime versions.
11. Go and TypeScript function inventories were generated before mutations. Go coverage records 562 reached lower/native functions; the port signature list contains 125 functions. The two C implementation files were read in full before choosing their mutations. Coverage is at Go function/block level, not C or TypeScript dynamic coverage; the TypeScript inventory is a conservative imported call surface.
12. W1/W2/W3 are the allowed witness comparison edits. Each forces the equality check to report agreement and every built-in mutant subcase rejects that result. W3 also removes an otherwise-unused bytes import to compile. Its failure is log line 153, mapped to origin/main line 154. Production-mutant precondition failures in witnesses were deliberately excluded from the production matrix.
13. P1 empties the Lower entry used by the compiler integration rows. Closed gaps recover the ensuing C-emission nil panic as a test failure; the shared-slice row does not, so it was run alone and its panic was observed. P2/P3 empty the port driver mains, leaving the Go and YAML-library oracles unchanged. Every production row failed its own entry probe. Witness vacuity is null because production probes do not apply to their checks.
14. There were no completed-matrix survivors. No repo-wide tests, unrequested YAML rows, C dynamic coverage, exhaustive semantic mutants, or changes to production behavior were retained. Native build timings in builds.json are additional warm-cache validation rebuilds, not measurements of the first cold runtime compilation within each matrix.

Timing: warm toolchain setup 0 seconds, nproc 5. Twenty-seven timing runs total 164.801 binary seconds and 217.773 command wall seconds. Whole-package setup products reported Go driver 3.47s, lowering 20.91s, native build 17.19s. Matrix, witness, probe, narrowed replay and rebuild durations and commands are in runs.json, narrow-runs.json, m4-other-runs.json and builds.json. The initial npm installs were not timed. Separate warm native builds measured M2 0.192494s, M3 0.245134s, M4 0.904854s. Overall unit work began about 13:12 UTC and finished about 13:40 UTC.
