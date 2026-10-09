Unit u087: all 23 names exist, grouped into 14 rows; nproc 5.
Starting origin/main: 571e74cf555b9db994c5dec6c2f8dbee676e5111.
Bounded port matrix: recovered grammar sacred; scalar family subsumed on two kills.
Four production mutants caught; four witnesses proven; two setup checks catch broken publication.
Three product-only construction rows remain green with missing artifacts; recipe empty-IR probe passes.

```json
[
  {
    "test": "TestRecoveryLoweredRecipe",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_lowered_recipe_test.go:21",
    "seconds": 0.589,
    "oracle": "Own C and JS bytes must be identical across private source paths. This checks path invariance, not semantic output correctness.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "Three isolated -count=1 baseline logs: TestRecoveryLoweredRecipe.[123].log",
    "timing_samples": [
      0.587,
      0.589,
      0.622
    ],
    "probe_passes": [
      "P02"
    ],
    "reason": "No meaningful compiler path-invariance production mutant was built within the bounded port audit. S05 repeated one input path and passed; it is supplemental construction evidence, not a production verdict."
  },
  {
    "test": "TestRecoveryMutants family",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_mutants_product_proofs_test.go:15",
    "seconds": 7.851,
    "oracle": "Go cohere answers compared with deliberately mutated source Node and sanitized native; W01 disables the comparison and only that run decides this witness verdict.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01 recovery_mutants_product_shards_test.go:179: module-await Node: mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -overlay=/workspace/adamic/review/test-audit/stage1-cohere-estree-recovery_lowered_recipe/W01-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestRecoveryMutants|TestRecoveryMutants_000|TestRecoveryMutants_001|TestRecoveryMutants_002|TestRecoveryMutantsUnion)$' > W01.TestRecoveryMutants_family.log 2>&1; recovery_mutants_product_shards_test.go:179: module-await Node: mutant survived",
    "timing_samples": [
      7.851,
      7.751,
      8.996
    ],
    "witness_edits": [
      "W01"
    ],
    "members": [
      "TestRecoveryMutants",
      "TestRecoveryMutants_000",
      "TestRecoveryMutants_001",
      "TestRecoveryMutants_002",
      "TestRecoveryMutantsUnion"
    ]
  },
  {
    "test": "TestRecoveryMutantsShardSurvivor",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_mutants_product_proofs_test.go:21",
    "seconds": 0.092,
    "oracle": "Synthetic bytes; requires exactly the owning child shard to fail.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01 recovery_shards_test.go:251: planted empty-type-list-range/008 must fail only shard-001: exit status 1",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -overlay=/workspace/adamic/review/test-audit/stage1-cohere-estree-recovery_lowered_recipe/W01-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestRecoveryMutantsShardSurvivor$' > W01.TestRecoveryMutantsShardSurvivor.log 2>&1; recovery_shards_test.go:251: planted empty-type-list-range/008 must fail only shard-001: exit status 1",
    "timing_samples": [
      0.211,
      0.035,
      0.092
    ],
    "witness_edits": [
      "W01"
    ]
  },
  {
    "test": "TestRecoveryMutantsTopSurvivor",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_mutants_product_proofs_test.go:36",
    "seconds": 0.084,
    "oracle": "Synthetic survivor; requires exactly the owning top-level shard to fail with mutant survived.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W01 recovery_mutants_product_proofs_test.go:78: planted survivor must fail only TestRecoveryMutants_001: exit status 1",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -overlay=/workspace/adamic/review/test-audit/stage1-cohere-estree-recovery_lowered_recipe/W01-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestRecoveryMutantsTopSurvivor$' > W01.TestRecoveryMutantsTopSurvivor.log 2>&1; recovery_mutants_product_proofs_test.go:78: planted survivor must fail only TestRecoveryMutants_001: exit status 1",
    "timing_samples": [
      0.07,
      0.169,
      0.084
    ],
    "witness_edits": [
      "W01"
    ]
  },
  {
    "test": "TestRecoveryMutants_Setup",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_mutants_product_proofs_test.go:84",
    "seconds": 9.918,
    "oracle": "Own product preparation and nonempty ready path; no agreement assertion.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S06 recovery_mutants_product_proofs_test.go:87: preparation failed for mutant first-accessibility",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -overlay=/workspace/adamic/review/test-audit/stage1-cohere-estree-recovery_lowered_recipe/S06-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestRecoveryMutants_Setup$' > S06.retry.log 2>&1; recovery_mutants_product_proofs_test.go:87: preparation failed for mutant first-accessibility",
    "timing_samples": [
      9.918,
      9.255,
      21.74
    ],
    "construction_edits": [
      "S06"
    ]
  },
  {
    "test": "TestRecoveryShardProofChild",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_shards_test.go:287",
    "seconds": 0.02,
    "oracle": "Synthetic subprocess entry; returns immediately without ADAMIC_ESTREE_SHARD_PROOF.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "helper",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "Three isolated -count=1 baseline logs: TestRecoveryShardProofChild.[123].log",
    "timing_samples": [
      0.02,
      0.02,
      0.012
    ],
    "parent": "TestRecoveryMutantsShardSurvivor"
  },
  {
    "test": "TestRecoveredGrammar",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_test.go:31",
    "seconds": 1.593,
    "oracle": "Go cohere serialized trees compared byte-for-byte with source Node, sanitized native and emitted JavaScript; subprocesses require successful exit and empty stderr.",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M02",
      "M03"
    ],
    "last_proven_fail": "M04 recovery_test.go:38: Node: line 57: Go \"4 .bigint string 1\", port \"4 .bigint string 11\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRecoveredGrammar",
      "TestScalarEdges family"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u087/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestRecoveredGrammar$ > M04.RecoveredGrammar.log 2>&1; recovery_test.go:38: Node: line 57: Go \"4 .bigint string 1\", port \"4 .bigint string 11\"",
    "timing_samples": [
      1.654,
      1.593,
      1.508
    ]
  },
  {
    "test": "TestRecoveryLibraryGaps",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/recovery_test.go:45",
    "seconds": 4.966,
    "oracle": "Pinned Go cohere accepts inputs that pinned typescript-estree refuses; handwritten accepted/refused label plus original-library comparisons. It does not execute the Adamic port.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
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
    "evidence": "Three isolated -count=1 baseline logs: TestRecoveryLibraryGaps.[123].log",
    "timing_samples": [
      4.532,
      5.147,
      4.966
    ],
    "reason": "Only Go cohere and pinned libraries are checked; mutating those or the harness would violate the oracle rule."
  },
  {
    "test": "TestScalarEdges preparation family",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/scalar_edges_split_products_test.go:152",
    "seconds": 5.651,
    "oracle": "Own construction of oracle, native and JS paths, plus nonempty readiness fields.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S01 scalar_edges_split_products_test.go:154: scalar-edge preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -overlay=/workspace/adamic/review/test-audit/stage1-cohere-estree-recovery_lowered_recipe/S01-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestScalarEdges|TestScalarEdges_Setup)$' > S01.TestScalarEdges_preparation_family.log 2>&1; scalar_edges_split_products_test.go:154: scalar-edge preparation failed",
    "timing_samples": [
      5.651,
      5.06,
      5.733
    ],
    "construction_edits": [
      "S01"
    ],
    "members": [
      "TestScalarEdges_Setup",
      "TestScalarEdges"
    ]
  },
  {
    "test": "TestScalarEdges family",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/scalar_edges_split_products_test.go:181",
    "seconds": 6.081,
    "oracle": "Go cohere serialized trees compared byte-for-byte with source Node, sanitized native and emitted JavaScript; union verifies every live corpus case appears once.",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M04"
    ],
    "unique_kills": [],
    "last_proven_fail": "M04 scalar_edges_split_products_test.go:277: emitted JS: line 38: Go \"2 .bigint string 1234567890123456789012345678901234567890\", port \"2 .bigint string 11234567890123456789012345678901234567890\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestRecoveredGrammar"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 1.593,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestRecoveredGrammar",
      "TestScalarEdges family"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u087/cache/M04 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run \"^(TestScalarEdgesUnion|TestScalarEdges_000|TestScalarEdges_001|TestScalarEdges_002|TestScalarEdges_003)$\" > M04.ScalarEdges_family.log 2>&1; scalar_edges_split_products_test.go:277: emitted JS: line 38: Go \"2 .bigint string 1234567890123456789012345678901234567890\", port \"2 .bigint string 11234567890123456789012345678901234567890\"",
    "timing_samples": [
      6.081,
      41.539,
      5.523
    ],
    "members": [
      "TestScalarEdgesUnion",
      "TestScalarEdges_000",
      "TestScalarEdges_001",
      "TestScalarEdges_002",
      "TestScalarEdges_003"
    ],
    "subsumption_mutants": 2
  },
  {
    "test": "TestScalarEdgesPlantedFailure",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/scalar_edges_split_products_test.go:204",
    "seconds": 4.949,
    "oracle": "Own comparison over Go cohere observations of an edited fixture; requires one owning shard to catch it.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W02 scalar_edges_split_products_test.go:233: planted case caught 0 times in shard -1",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -overlay=/workspace/adamic/review/test-audit/stage1-cohere-estree-recovery_lowered_recipe/W02-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestScalarEdgesPlantedFailure$' > W02.TestScalarEdgesPlantedFailure.log 2>&1; scalar_edges_split_products_test.go:233: planted case caught 0 times in shard -1",
    "timing_samples": [
      4.949,
      4.91,
      5.004
    ],
    "witness_edits": [
      "W02"
    ]
  },
  {
    "test": "TestProduct_scalar_edges_go_oracle",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/scalar_edges_split_products_test.go:285",
    "seconds": 2.406,
    "oracle": "Own builder completion only; returned artifact paths are not checked by this row.",
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
    "matrix_rows": [],
    "evidence": "timeout 120 go test -overlay=/workspace/adamic/review/test-audit/stage1-cohere-estree-recovery_lowered_recipe/S03-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestProduct_scalar_edges_go_oracle$' > S03.TestProduct_scalar_edges_go_oracle.log 2>&1; PASS",
    "timing_samples": [
      2.914,
      2.217,
      2.406
    ],
    "construction_edits": [
      "S03"
    ],
    "reason": "Permitted construction edit removes the expected artifact; row still passes."
  },
  {
    "test": "TestProduct_scalar_edges_lowered",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/scalar_edges_split_products_test.go:290",
    "seconds": 1.63,
    "oracle": "Own builder completion only; returned artifact paths are not checked by this row.",
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
    "matrix_rows": [],
    "evidence": "timeout 120 go test -overlay=/workspace/adamic/review/test-audit/stage1-cohere-estree-recovery_lowered_recipe/S02-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestProduct_scalar_edges_lowered$' > S02.TestProduct_scalar_edges_lowered.log 2>&1; PASS",
    "timing_samples": [
      1.63,
      1.662,
      1.56
    ],
    "construction_edits": [
      "S02"
    ],
    "reason": "Permitted construction edit removes the expected artifact; row still passes."
  },
  {
    "test": "TestProduct_scalar_edges_sanitized_native",
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/scalar_edges_split_products_test.go:295",
    "seconds": 3.929,
    "oracle": "Own builder completion only; returned artifact paths are not checked by this row.",
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
    "matrix_rows": [],
    "evidence": "timeout 120 go test -overlay=/workspace/adamic/review/test-audit/stage1-cohere-estree-recovery_lowered_recipe/S04-overlay.json -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestProduct_scalar_edges_sanitized_native$' > S04.TestProduct_scalar_edges_sanitized_native.log 2>&1; PASS",
    "timing_samples": [
      4.041,
      3.9290000000000003,
      3.352
    ],
    "construction_edits": [
      "S04"
    ],
    "reason": "Permitted construction edit removes the expected artifact; row still passes."
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | stage1/cohere/estree/protocol.ts:20 | `unit >= 32 && unit <= 126` → `unit >= 33 && unit <= 126` | TestRecoveredGrammar, TestScalarEdges family |
| M02 | stage1/cohere/estree/values.ts:23 | `value.number = flag ? 1 : 0;` → `value.number = flag ? 0 : 1;` | TestRecoveredGrammar |
| M03 | stage1/cohere/estree/pipeline.ts:33 | `parser.awaitContext = true;` → `parser.awaitContext = false;` | TestRecoveredGrammar |
| M04 | stage1/cohere/estree/values.ts:81 | `let result = '0';` → `let result = '1';` | TestRecoveredGrammar, TestScalarEdges family |

| Edit | Origin file:line | Change | Observed result |
|---|---|---|---|
| W01 | recovery_shards_test.go:220 | comparison always sees equal bytes | Recovery mutant family and both survivor witnesses fail |
| W02 | estree_test.go:163 | firstDifference returns empty | Scalar planted-failure witness fails |
| S01 | scalar_edges_split_products_test.go:130 | drop readiness publication | Both preparation family members fail |
| S02 | scalar_edges_split_products_test.go:55 | drop C artifact write | Lowered-product row passes; no port.c |
| S03 | scalar_edges_split_products_test.go:103 | drop Go oracle build | Oracle-product row passes; no oracle executable |
| S04 | scalar_edges_split_products_test.go:73 | drop native build | Native-product row passes; no port executable |
| S05 | recovery_lowered_recipe_test.go:33 | repeat main input path | Recipe passes; supplemental equivalent candidate for observed output |
| S06 | recovery_mutants_product_shards_test.go:198 | drop prepared directory publication | Recovery setup row fails |
| P01 | pipeline.ts:10 | answer returns empty string | Both port agreement rows fail |
| P02 | internal/lower/lower.go:20 | Lower returns empty successful IR | Recipe passes |

All abbreviated edit locations are under stage1/cohere/estree unless explicitly qualified. S06's exact origin line is also recorded in S06-command.json.

Production survivors: none.
Construction survivors: S02 has no port.c; S03 has no oracle executable; S04 has no port executable, with matching artifacts present in the baseline cache. S05 is an equivalent candidate for observed compiler bytes; independence of its input paths changed, but output did not.

Not covered: full package after timeout, repository uniqueness, exact function-level runtime reach, and a meaningful compiler path-invariance production mutant. No requested rows skipped. Source restored and final bounded control passed. No pure scalar-build timing is available.

Brief feedback

1. The historical 8de93800f4 is not the fetched starting commit. All 23 requested functions still exist; grouping their bodies yields 14 rows.
2. recovery_mutants_preparation_test.go and scalar_edges_preparation_test.go vanished. Their tests moved to recovery_mutants_product_proofs_test.go/recovery_mutants_product_shards_test.go and scalar_edges_split_products_test.go. File-based scope would have missed them.
3. The whole package cooks at 90 seconds in outside-unit compiler setup. The first bounded slice also cooks after the recipe's two cold lowerings. The next port slice cooks in concurrent mutant preparation. None has an individual assertion failure before timeout; the independent requested baselines eventually pass.
4. Warm tools do not imply warm products. Native products require substantial C emission and clang time. Isolated preparation and reused completed products made all requested rows measurable under 90 seconds.
5. Native rebuild fallback limits this production plan to four mutants, although the generic target of three per row would ask for many more. Two distinct native recipes are rebuilt for each mutant. Their combined row command costs and available pure BUILD timings are preserved.
6. The three scalar product rows call different builders, so they are separate rows. TestScalarEdges and TestScalarEdges_Setup have identical preparation checks and are one family, distinct from the agreement shard family. Recovery mutant wrappers plus their union are one witness family.
7. Production edits are only in the TypeScript port. Go cohere and its answers are unchanged. Recipe path-invariance checks compiler-generated bytes instead; no meaningful compiler path-invariance mutant was built within this bounded port audit, so that verdict is cannot-judge.
8. An empty successful Lower IR passes the recipe equality test. This is vacuous under the brief's empty-answer definition, while still satisfying its narrower path-invariance property. The label should not be read as a semantic-output claim that this metamorphic test makes.
9. Returning an empty string from pipeline.answer fails both agreement rows. These probes stay out of production kills and uniqueness.
10. The pinned-library gap row never runs Adamic. Its subjects are Go cohere and the external libraries themselves. Mutating them would violate the oracle restriction. The current brief supplies no meaningful allowed production mutant for that row.
11. W01 disables recoveryComparison, rather than breaking a port precondition. The real mutant family and both synthetic witnesses fail. W02 disables firstDifference and the scalar planted-failure witness fails. Production failures are not counted for these witnesses.
12. Physical harness edits enter broad product keys and trigger irrelevant native rebuilds. Go overlays let witness/construction changes compile without altering the physical port or its cache inputs. Each replay diff still applies to the starting commit and passes vet or the native port build.
13. S02, S03 and S04 remove required construction artifacts, yet the corresponding product-only tests pass. They assert builder completion, not that the named artifact exists. Their untrue verdicts concern this tested construction behavior, not production-port mutants.
14. S05 repeats the same recipe input path twice and still passes. It is supplemental input-independence evidence, with no production verdict resting on it; no output difference was demonstrated, so it is an equivalent candidate for the observed byte invariant.
15. The package matrix is bounded to recovered grammar and the complete scalar agreement family. Unique kills are only within that matrix. All other package rows and repository uniqueness remain unknown. Scalar subsumption rests on two caught mutants and is a hint, not a deletion recommendation.
16. The port inventory is a conservative transitive declaration inventory, not an instrumented exact function-reach profile. Each planted production function's reach is demonstrated by its kills, but complete runtime reachability was not measured.
17. All requested baselines ran without skips. TestRecoveryShardProofChild returning without its parent environment is a helper behavior, not an opt-in skip. The pinned library was installed and enabled.
18. Timing medians use three separate -count=1 test-binary lines. Product caches were warm for later samples; cold preparation and rebuild times are reported separately. Some earlier timing commands overlapped preparation, so these are not machine-isolated benchmark results.
19. Scalar builders do not log pure build-phase times. Their command totals include lowering, native compilation and case work; I did not invent separate clang timings for them.
20. The brief initially requires pushing and later permits skipping it. I followed the initial explicit request. No main push or pull request.

Timings

```json
{
  "setup_seconds": 0,
  "nproc": 5,
  "initial_commands": [
    {
      "name": "list",
      "command": [
        "timeout",
        "120",
        "go",
        "test",
        "-list",
        ".",
        "./stage1/cohere/estree/"
      ],
      "exit": 0,
      "seconds": 7.9494595090000075
    },
    {
      "name": "npm",
      "command": [
        "timeout",
        "90",
        "npm",
        "ci",
        "--prefix",
        "stage3/api"
      ],
      "exit": 0,
      "seconds": 0.7077757859988196
    },
    {
      "name": "library-install",
      "command": [
        "timeout",
        "90",
        "npm",
        "install",
        "--prefix",
        "/tmp/u087-estree-library",
        "--no-audit",
        "--no-fund",
        "@typescript-eslint/typescript-estree@8.65.0",
        "typescript@6.0.3",
        "prettier@3.9.6"
      ],
      "exit": 0,
      "seconds": 4.481750666000153
    },
    {
      "name": "baseline",
      "command": [
        "timeout",
        "120",
        "go",
        "test",
        "-json",
        "-count=1",
        "-timeout",
        "90s",
        "./stage1/cohere/estree/",
        "-run",
        "."
      ],
      "exit": 1,
      "seconds": 92.58934424299878,
      "env": {
        "ADAMIC_ESTREE_LIBRARY": "/tmp/u087-estree-library",
        "ADAMIC_GATE_UNCACHED": "1"
      }
    }
  ],
  "production_matrix_command_seconds": 455.32640616799836,
  "port_probe_command_seconds": 111.49203645899615,
  "logged_production_lowering_seconds": 113.259251,
  "logged_production_native_build_seconds": 155.559024,
  "construction_measurement_command_seconds": 313.3393466430025,
  "phase_totals_are_command_sums_not_session_wall_time": true
}
```

Rebuild details: rebuilds.json. Commands and logs preserve individual native rebuild times.

Additional feedback

21. Final evidence validation found S06 initially failed to link because /tmp was full. That build failure was not a kill. I removed only this unit's disposable M01-M04/P01 caches after preserving logs and reran S06; the test then failed on preparation failed for mutant first-accessibility. Both logs are preserved.
22. The cold setup, matrix, reporting and disk retry exceeded the approximate 30-minute port budget. No further testing was performed after that evidence correction. The timestamp span from the first whole-baseline start to the final retry was about 32 minutes, excluding initial fetch/read/install work.
