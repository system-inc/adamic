u112: 50 named tests present, grouped into 15 rows.
Start: ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2; warm toolchain; nproc 5.
Verdicts: 2 subsumed, 2 witness, 3 setup-check, 4 untrue, 4 cannot-judge.
Production matrix bounded to two parity families; full-corpus ShardsAgree preparation timed out.
Evidence branch: test-audit/stage1-cohere-lint-shards_agree_split; production/test sources restored.

```json
[
  {
    "test": "TestShardsAgree family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/shards_agree_split_test.go",
    "seconds": null,
    "oracle": "Same native lint driver run serially, exact byte and count comparison; self consistency does not establish semantic parity.",
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
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestShardsAgree_[0-9]+$; panic: test timed out after 1m30s; no completed clean baseline or mutant observation",
    "members": [
      "TestShardsAgree_000",
      "TestShardsAgree_001",
      "TestShardsAgree_002",
      "TestShardsAgree_003",
      "TestShardsAgree_004",
      "TestShardsAgree_005",
      "TestShardsAgree_006",
      "TestShardsAgree_007",
      "TestShardsAgree_008",
      "TestShardsAgree_009",
      "TestShardsAgree_010",
      "TestShardsAgree_011",
      "TestShardsAgree_012",
      "TestShardsAgree_013",
      "TestShardsAgree_014",
      "TestShardsAgree_015",
      "TestShardsAgree_016",
      "TestShardsAgree_017",
      "TestShardsAgree_018",
      "TestShardsAgree_019",
      "TestShardsAgree_020",
      "TestShardsAgree_021",
      "TestShardsAgree_022",
      "TestShardsAgree_023",
      "TestShardsAgree_024",
      "TestShardsAgree_025",
      "TestShardsAgree_026",
      "TestShardsAgree_027",
      "TestShardsAgree_028",
      "TestShardsAgree_029",
      "TestShardsAgree_030",
      "TestShardsAgree_031"
    ]
  },
  {
    "test": "TestShardsAgree_Union",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/shards_agree_split_test.go",
    "seconds": null,
    "oracle": "Hand-written enumeration, exact-once ownership, and planted byte disagreement checked with difference; witness could not reach its check within budget.",
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
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestShardsAgree_Union$; panic: test timed out after 1m30s; no completed clean baseline or mutant observation",
    "members": [
      "TestShardsAgree_Union"
    ]
  },
  {
    "test": "TestShardsAgree_Setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_setup_clock_regression_test.go",
    "seconds": null,
    "oracle": "Suite product/corpus construction, build success; preparation did not complete.",
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
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestShardsAgree_Setup$; panic: test timed out after 1m30s; no completed clean baseline or mutant observation",
    "members": [
      "TestShardsAgree_Setup"
    ]
  },
  {
    "test": "TestShardsAgree_SetupRequired",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/lint_setup_clock_regression_test.go",
    "seconds": null,
    "oracle": "Hand-written child PASS, named prepared-product build and cache-miss checks; preparation did not complete.",
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
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestShardsAgree_SetupRequired$; panic: test timed out after 1m30s; no completed clean baseline or mutant observation",
    "members": [
      "TestShardsAgree_SetupRequired"
    ]
  },
  {
    "test": "TestSuggestionAlongsideAutomaticFix_Setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/suggestion_alongside_shards_test.go",
    "seconds": 7.145,
    "oracle": "Construction success only: setting ready=false still passes; no assertion requires the ready flag.",
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
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u112/weak/S2.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestSuggestionAlongsideAutomaticFix_Setup$; --- PASS: TestSuggestionAlongsideAutomaticFix_Setup (7.93s)",
    "members": [
      "TestSuggestionAlongsideAutomaticFix_Setup"
    ]
  },
  {
    "test": "TestSuggestionAlongsideAutomaticFix family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/suggestion_alongside_shards_test.go",
    "seconds": 5.307,
    "oracle": "Go cohere executes the pinned rule adapter; exact output bytes across source Node, emitted JavaScript and sanitized native, plus hand-written automatic-fix substring.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: suggestion_alongside_shards_test.go:206: native: case 0 line 5: port \"range 9 17 unexpectedDebugger suggestions\\t\\t\\t8 17\", Go \"range 8 17 unexpectedDebugger suggestions\\t\\t\\t8 17\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestWitnessScriptKind family"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 8.129,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestSuggestionAlongsideAutomaticFix_[0-9]+$; selector=M2; suggestion_alongside_shards_test.go:206: native: case 0 line 5: port \"range 9 17 unexpectedDebugger suggestions\\t\\t\\t8 17\", Go \"range 8 17 unexpectedDebugger suggestions\\t\\t\\t8 17\"",
    "members": [
      "TestSuggestionAlongsideAutomaticFix_000",
      "TestSuggestionAlongsideAutomaticFix_001",
      "TestSuggestionAlongsideAutomaticFix_002"
    ],
    "subsumption_mutants": 1
  },
  {
    "test": "TestSuggestionAlongsideAutomaticFixPlantedFailure",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/suggestion_alongside_shards_test.go",
    "seconds": 9.748,
    "oracle": "Built-in native-only byte mismatch; parent requires exactly shard 002 failure and specific planted text. W1 makes difference accept all bytes.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1-suggestion: suggestion_alongside_shards_test.go:283: wrong planted failure: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u112/weak/W1.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestSuggestionAlongsideAutomaticFixPlantedFailure$; suggestion_alongside_shards_test.go:283: wrong planted failure: <nil>",
    "members": [
      "TestSuggestionAlongsideAutomaticFixPlantedFailure"
    ]
  },
  {
    "test": "TestSuggestionAlongsideAutomaticFixSetupIsRequired",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/suggestion_alongside_shards_test.go",
    "seconds": 25.747,
    "oracle": "Fresh-process child PASS and cache-miss assertions; S5 disables selected leaf preparation and parent fails.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S5-required: suggestion_alongside_shards_test.go:294: standalone cold shard: exit status 1 (scratch line 294 maps to origin/main line 295)",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u112/weak/S5.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestSuggestionAlongsideAutomaticFixSetupIsRequired$; suggestion_alongside_shards_test.go:294: standalone cold shard: exit status 1 (scratch line 294 maps to origin/main line 295)",
    "members": [
      "TestSuggestionAlongsideAutomaticFixSetupIsRequired"
    ]
  },
  {
    "test": "TestProduct_WitnessScriptKindLowered",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/witness_script_kind_products_test.go",
    "seconds": 3.003,
    "oracle": "Product construction only; S1 returns empty directory from all product makers and this wrapper still passes. No output comparison.",
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
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u112/weak/S1.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestProduct_WitnessScriptKind; --- PASS: TestProduct_WitnessScriptKindLowered (0.00s)",
    "members": [
      "TestProduct_WitnessScriptKindLowered"
    ]
  },
  {
    "test": "TestProduct_WitnessScriptKindNative",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/witness_script_kind_products_test.go",
    "seconds": 4.369,
    "oracle": "Product construction only; S1 returns empty directory from all product makers and this wrapper still passes. No output comparison.",
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
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u112/weak/S1.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestProduct_WitnessScriptKind; --- PASS: TestProduct_WitnessScriptKindLowered (0.00s)",
    "members": [
      "TestProduct_WitnessScriptKindNative"
    ]
  },
  {
    "test": "TestProduct_WitnessScriptKindGoOracle",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/witness_script_kind_products_test.go",
    "seconds": 2.18,
    "oracle": "Product construction only; S1 returns empty directory from all product makers and this wrapper still passes. No output comparison.",
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
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u112/weak/S1.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestProduct_WitnessScriptKind; --- PASS: TestProduct_WitnessScriptKindLowered (0.00s)",
    "members": [
      "TestProduct_WitnessScriptKindGoOracle"
    ]
  },
  {
    "test": "TestWitnessScriptKind family",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/witness_script_kind_shards_test.go,stage1/cohere/lint/witness_script_kind_union_test.go",
    "seconds": 8.129,
    "oracle": "Go cohere executes no-debugger; exact output bytes across source Node, emitted JavaScript and sanitized native; union has hand-written exact-once live-corpus coverage.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: witness_script_kind_shards_test.go:153: Node: case 0 line 5: port \"range 17 25 unexpectedDebugger fix\\t\\t\\t16 25\", Go \"range 16 25 unexpectedDebugger fix\\t\\t\\t16 25\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestSuggestionAlongsideAutomaticFix family"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 5.307,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestWitnessScriptKind(?:_[0-9]+)?$; selector=M2; witness_script_kind_shards_test.go:153: Node: case 0 line 5: port \"range 17 25 unexpectedDebugger fix\\t\\t\\t16 25\", Go \"range 16 25 unexpectedDebugger fix\\t\\t\\t16 25\"",
    "members": [
      "TestWitnessScriptKind_000",
      "TestWitnessScriptKind_001",
      "TestWitnessScriptKind"
    ],
    "subsumption_mutants": 1,
    "vacuous_subcases": [
      "TestWitnessScriptKind (coverage union)"
    ]
  },
  {
    "test": "TestWitnessScriptKindPlantedFailure",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/witness_script_kind_shards_test.go",
    "seconds": 16.161,
    "oracle": "Built-in emitted-JavaScript-only byte mismatch; parent requires exactly assigned leaf failure and emitted JavaScript diagnostic. W1 accepts all bytes, exposing planted mismatch survived instead.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1-kind: witness_script_kind_shards_test.go:199: wrong planted failure: exit status 1",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u112/weak/W1.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestWitnessScriptKindPlantedFailure$; witness_script_kind_shards_test.go:199: wrong planted failure: exit status 1",
    "members": [
      "TestWitnessScriptKindPlantedFailure"
    ]
  },
  {
    "test": "TestWitnessScriptKind_Setup",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/witness_script_kind_shards_test.go",
    "seconds": 7.674,
    "oracle": "Hand-written witness extension and ready/products/source-count construction checks; S3 constructs .ts filenames for .tsx witnesses and the unchanged extension assertion fails.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S3-kind-setup: witness_script_kind_shards_test.go:81: witness script kind lost",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u112/weak/S3.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestWitnessScriptKind_Setup$; witness_script_kind_shards_test.go:81: witness script kind lost",
    "members": [
      "TestWitnessScriptKind_Setup"
    ]
  },
  {
    "test": "TestWitnessScriptKindStandalone",
    "package": "stage1/cohere/lint",
    "file": "stage1/cohere/lint/witness_script_kind_shards_test.go",
    "seconds": 17.752,
    "oracle": "Parent requires child named-leaf PASS; S4 selects a nonexistent child row and fails that assertion.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S4-kind-standalone: witness_script_kind_shards_test.go:306: standalone leaf failed: <nil>",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestSuggestionAlongsideAutomaticFix family",
      "TestWitnessScriptKind family"
    ],
    "evidence": "timeout 120 go test -overlay=/tmp/u112/weak/S4.json -json -count=1 -timeout 90s ./stage1/cohere/lint/ -run ^TestWitnessScriptKindStandalone$; witness_script_kind_shards_test.go:306: standalone leaf failed: <nil>",
    "members": [
      "TestWitnessScriptKindStandalone"
    ]
  }
]
```

| ID | Origin/main location | Allowed-menu change | Failed production rows |
| --- | --- | --- | --- |
| M1 | stage1/cohere/lint/main.ts:178 | if(caseNumber % (shardCount + 1) === shardIndex) | None in bounded matrix; outside unknown |
| M2 | stage1/cohere/lint/main.ts:113 | `range ${start + 1} ${end} | TestSuggestionAlongsideAutomaticFix family, TestWitnessScriptKind family |
| M3 | stage1/cohere/lint/shards/shards.go:65 | total := 1 | None in bounded matrix; outside unknown |
| M4 | stage1/cohere/lint/shards/shards.go:132 | drop the entire final assembly loop, avoiding an unused block variable | None in bounded matrix; outside unknown |

Survivors in the bounded matrix:

- M1: native-witness-clean.log prints case 0 and case 1; native-witness-M1.log prints only case 0 on the same two-row manifest, both exit zero. Native binary and arguments are in native-witness.json. Full-corpus ShardsAgree coverage is unknown.
- M3: the direct production shards.Run witness changes count from "7\n" to "8\n", error nil. These parity families never call Run; full-corpus ShardsAgree coverage is unknown.
- M4: the direct production shards.Merge witness changes "case 0\nbody\n" to an empty byte slice, error nil. These parity families never call Merge; full-corpus ShardsAgree coverage is unknown.

These are changed behaviors, not equivalent candidates. This bounded survivor observation does not establish that the behavior is unguarded in the package. No package-unique or repository-unique kill was established.

P1 empties main.run at entry. It kills both parity families; every runtime leaf fails. TestWitnessScriptKind, the coverage-union member, passes, so the family is not vacuous but that member is listed in vacuous_subcases. P2 empties shards.Run. Neither completed parity family calls that entry; both pass, and those passes decide no row's vacuity. ShardsAgree has null vacuity because its own entry could not be meaningfully tested within budget. Empty probes do not count as production kills.

W1 returns an empty difference for all byte comparisons. Both planted-failure witnesses now fail their parent assertions, confirming that they distinguish a real comparison failure from a silent survivor. S1 returns empty paths from all three script-kind product makers; their TestProduct wrappers still pass. S2 leaves suggestion setup's ready flag false; the setup wrapper still passes. These are untrue under the requested construction-weakening rubric, not proof that the wrappers cannot fail on an arbitrary toolchain or filesystem error. S3 constructs .ts files for .tsx witnesses and the unchanged extension assertion fails. S4 selects a nonexistent child leaf and the standalone PASS assertion fails. S5 drops the selected suggestion leaf's preparation and its cold-setup parent fails. Every overlay passes go vet.

Brief ambiguities, corrections, and costs:

1. The historical SHA 8de93800f4 does not identify current origin/main. The required fresh branch started at ce1c5a2f. All 50 requested names still exist. TestShardsAgree_Setup and TestShardsAgree_SetupRequired moved from the historical file list to lint_setup_clock_regression_test.go. Nothing vanished. scope.json records every current location against the starting commit.
2. The 15-row header and 50-function list require family grouping. The 32 ShardsAgree leaves are one family. Its union is separate because it additionally witnesses an intentionally planted disagreement and requires exact attribution. The three suggestion runtimes are one shared-check family. The two script-kind leaves and their live-corpus coverage union are one family. The three product wrappers call different lowering, native-building, and Go-oracle recipes and are separate construction rows. Generic buildcache.Product alone is not their checker. Family medians are three complete family runs, not a fastest member's timing.
3. Warm env.sh did not configure ADAMIC_TYPESCRIPT_SOURCE. The tests require a corpus and fail rather than skip when it is absent. I fetched the exact pin named in cloud/lint-wave-check.md, taking 19.874 seconds, before baseline. Mandatory npm ci took 1.057 seconds. No completed scoped row skipped. Rows that timed out never reached their assertions; a scope listing is not evidence that they executed.
4. The whole package consumed 90.662 binary seconds and generated about 20 MB of JSON, mostly unrelated test scheduling and a timeout stack. It did not provide an assertion-red baseline. The requested combined slice consumed 90.069 binary seconds while preparing products. I narrowed to grouped rows. Four independent ShardsAgree rows also timed out in full-corpus preparation. Their clean medians, mutant failures, and vacuity are unknown. Changing their corpus or injecting a smaller prepared product would change the harness, so I did not use that to manufacture production kills.
5. The cold product recipe still exceeds the audit's deadline for those rows. Repeated independent selection confirmed the same preparation limit but did not make that family auditable. Prepared-product phases could be replayed centrally on a warmer machine. I did not grant a larger timeout or audit on a red assertion baseline.
6. The completed production matrix covers two parity families, each run independently per mutant. They reach the TypeScript main entry but do not call the Go shard coordinator. M3 and M4 are therefore bounded survivors in this matrix, supported by direct changed-output witnesses. They are not evidence of package-wide unguarded behavior. Kills in other rows and packages remain unknown. The two subsumption hints each rest on one caught mutant, M2; neither supports deleting a test.
7. ShardsAgree's expected bytes come from the same native driver run serially. Its oracle is self consistency, not Go cohere, even though preparing its corpus runs Go. The parity families actually execute Go cohere and compare exact output, including finding offsets and serialized repair data. Their construction and planted-attribution expectations are handwritten. Count or exit-only assertions in setup parents are identified in rows.json and were checked by their own construction probes.
8. I initially implemented S3 by flipping its extension assertion. That was invalid evidence for a construction check. I discarded it and reran S3 by changing the constructed filename extension, leaving the assertion intact. The excluded observation and explanation remain in invalid-control.json and S3-invalid-* files. No verdict rests on that edit. S5's observed scratch line 294 maps to starting-commit line 295; the JSON explicitly records that mapping.
9. The 462-function V8 reach inventory is observed on a source-Node no-debugger witness, including anonymous callback ranges, plus the coordinator's statically read Run/Merge/isCaseLine path. It is not an exhaustive reach inventory for the full compiler corpus that could not finish its baseline. That remaining inventory is unknown. Raw filtered V8 coverage is retained. No mutation changed Go cohere, the Go rule adapter, or external expected outputs.
10. The port's runtime selector permits one switched source build per existing product recipe. Compiler sources were never mutated, so there was no compiler cache-key hazard. Cache identities include actual port contents. Each standalone TypeScript diff was also rebuilt independently through the port's no-debugger native recipe; this compile evidence is narrower than compiling the complete live rule registry. M3, M4 and P2 pass go vet on the coordinator package. All standalone production/probe/allowed-harness diffs apply cleanly to the starting commit.
11. The empty-answer rule is applied only to a row's own semantic entry. Passing P2 on a family that never calls Run would falsely label it vacuous, so no such inference appears. Construction rows and witnesses have null production-entry vacuity. Their empty construction results and disabled comparisons are kept separate from production mutations.

Timing and coverage limits:

Tool setup: 0 seconds, Go 1.27.1, Node 24.19.0, nproc 5. Clean isolated command wall: 882.701s. Mutation/control/standalone command wall: 330.246s. Valid witness/construction command wall: 121.186s. The latter excludes the discarded S3 observation. Initial whole-package and combined-slice binary times are 90.662s and 90.069s. Command walls include Go compilation and cache lookup; row seconds use only each binary's own reported elapsed value.

Switched control builds: suggestion lowering 19.42s, sanitized native 21.24s, Go oracle 4.17s; script-kind lowering 18.38s, sanitized native 24.81s, Go oracle 4.90s. These rebuild once per switched product, not per selector value.

Standalone compile commands: M1-compile 25.038s (exit 0), M2-compile 25.890s (exit 0), M3-compile 0.297s (exit 0), M4-compile 0.147s (exit 0), P1-compile 9.673s (exit 0), P2-compile 0.103s (exit 0).

No repository-wide uniqueness replay, complete full-rule-registry native rebuild, or mutation verdict for the four cooked ShardsAgree rows was completed. The full-corpus function inventory remains incomplete. Production and test sources were restored before committing evidence. No PR was opened and main was not pushed.
