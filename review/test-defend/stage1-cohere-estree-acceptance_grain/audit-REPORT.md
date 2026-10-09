u084 audited origin/main 12e77e8972a2e606cab6db05d84f428246a85339.
All 20 names exist; family grouping yields 12 rows; four deep-mutant functions moved files.
Verdicts: 3 setup-check, 3 witness, 2 sacred, 1 untrue, 2 cannot-judge, 1 helper.
Four production mutants: three killed, one diagnostic-detail survivor; P1 killed both production rows.
Evidence: test-audit/stage1-cohere-estree-acceptance_grain, review/test-audit/stage1-cohere-estree-acceptance_grain/.

```json
[
  {
    "test": "TestProduct_AcceptanceOracle",
    "members": [
      "TestProduct_AcceptanceOracle"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/acceptance_grain_test.go:135",
    "seconds": 1.551,
    "oracle": "Successful Go-oracle product construction; self-written builder error checks, no artifact existence/content assertion.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S4: acceptance_grain_test.go:111: build estree-acceptance-mutants-go-oracle: open /tmp/u084/cache/S4/.building-5278deaa5725-3111034126/missing/oracle: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_AcceptanceOracle"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestProduct_AcceptanceOracle)$' > S4.log 2>&1; acceptance_grain_test.go:111: build estree-acceptance-mutants-go-oracle: open /tmp/u084/cache/S4/.building-5278deaa5725-3111034126/missing/oracle: no such file or directory",
    "timing_runs": [
      2.6390000000000002,
      1.141,
      1.5510000000000002
    ],
    "setup_kills": [
      "S4"
    ],
    "construction_survivors": [
      "S1"
    ]
  },
  {
    "test": "TestProduct_Acceptance family",
    "members": [
      "TestProduct_AcceptanceCatchLowered",
      "TestProduct_AcceptanceCatchNative",
      "TestProduct_AcceptanceClassLowered",
      "TestProduct_AcceptanceClassNative"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/acceptance_grain_test.go:136",
    "seconds": 3.443,
    "oracle": "Successful lowered/native product construction; self-written builder error checks, returned paths are not inspected.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S5: acceptance_grain_test.go:138: build estree-acceptance-mutant-lowered-catch-initializer: open /tmp/u084/cache/S5/.building-4181deb19a90-1392249814/missing/port.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestProduct_AcceptanceCatchLowered",
      "TestProduct_AcceptanceCatchNative",
      "TestProduct_AcceptanceClassLowered",
      "TestProduct_AcceptanceClassNative"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestProduct_AcceptanceCatchLowered|TestProduct_AcceptanceCatchNative|TestProduct_AcceptanceClassLowered|TestProduct_AcceptanceClassNative)$' > S5.log 2>&1; acceptance_grain_test.go:138: build estree-acceptance-mutant-lowered-catch-initializer: open /tmp/u084/cache/S5/.building-4181deb19a90-1392249814/missing/port.c: no such file or directory",
    "timing_runs": [
      40.664,
      3.271,
      3.443
    ],
    "setup_kills": [
      "S5"
    ],
    "construction_survivors": [
      "S2"
    ]
  },
  {
    "test": "TestAcceptanceMutants family",
    "members": [
      "TestAcceptanceMutants_000",
      "TestAcceptanceMutants_001",
      "TestAcceptanceMutantsUnion"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/acceptance_grain_test.go:223",
    "seconds": 6.187,
    "oracle": "Live Go cohere canonical bytes against built-in port mutants on source Node and sanitized native; self-written mutant-must-differ and shard-union checks.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: acceptance_grain_test.go:224: Node mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestAcceptanceMutants_000",
      "TestAcceptanceMutants_001",
      "TestAcceptanceMutantsUnion",
      "TestDeepMutants_000",
      "TestDeepMutants_001",
      "TestDeepMutants_002"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestAcceptanceMutants_000|TestAcceptanceMutants_001|TestAcceptanceMutantsUnion|TestDeepMutants_000|TestDeepMutants_001|TestDeepMutants_002)$' > W1.log 2>&1; acceptance_grain_test.go:224: Node mutant survived",
    "timing_runs": [
      6.187,
      5.675,
      6.902
    ],
    "witness_kills": [
      "W1"
    ]
  },
  {
    "test": "TestAcceptanceMutantsPlanted family",
    "members": [
      "TestAcceptanceMutantsPlanted_000",
      "TestAcceptanceMutantsPlanted_001"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/acceptance_grain_test.go:278",
    "seconds": 6.658,
    "oracle": "Real child acceptance check using Go cohere; self-written expectation that only shard 000 rejects the planted surviving native result.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W2: acceptance_grain_test.go:278: planted surviving mutant must fail only TestAcceptanceMutants_000: <nil>",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestAcceptanceMutantsPlanted_000",
      "TestAcceptanceMutantsPlanted_001"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestAcceptanceMutantsPlanted_000|TestAcceptanceMutantsPlanted_001)$' > W2.log 2>&1; acceptance_grain_test.go:278: planted surviving mutant must fail only TestAcceptanceMutants_000: <nil>",
    "timing_runs": [
      4.763,
      6.658,
      7.223
    ],
    "witness_kills": [
      "W2"
    ]
  },
  {
    "test": "TestAcceptanceGrammar",
    "members": [
      "TestAcceptanceGrammar"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/acceptance_test.go:17",
    "seconds": 26.562,
    "oracle": "Live Go cohere canonical bytes compared exactly with source Node, sanitized native and emitted JavaScript.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [
      "M1",
      "M2"
    ],
    "last_proven_fail": "M2: acceptance_test.go:24: emitted: line 1: Go \"0 Program 0 26 0 0 0 0 26 0\", port \"1 Program 0 26 0 0 0 0 26 0\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestAcceptanceGrammar",
      "TestAcceptanceDiagnostics"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestAcceptanceGrammar|TestAcceptanceDiagnostics)$' > M2.log 2>&1; acceptance_test.go:24: emitted: line 1: Go \"0 Program 0 26 0 0 0 0 26 0\", port \"1 Program 0 26 0 0 0 0 26 0\"",
    "timing_runs": [
      24.823,
      26.562,
      26.792
    ]
  },
  {
    "test": "TestAcceptanceDiagnostics",
    "members": [
      "TestAcceptanceDiagnostics"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/acceptance_test.go:31",
    "seconds": 28.022,
    "oracle": "Live Go cohere rejects eight inputs; port check requires nonzero exit, empty stdout, no CPU timeout and ESTree parser substring. M4 preserves these while losing the specific diagnostic.",
    "oracle_kind": "external-run",
    "kills": [
      "M3"
    ],
    "unique_kills": [
      "M3"
    ],
    "last_proven_fail": "M3: acceptance_test.go:46: [node --disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/estree/main.ts /tmp/adamic-gate/TestAcceptanceDiagnostics2586825486/001/0000.ts]: timeout=false exit=exit status 70 stdout=0 stderr=adamic: panic: ESTree validator: update operand must be a left hand side expression",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestAcceptanceGrammar",
      "TestAcceptanceDiagnostics"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestAcceptanceGrammar|TestAcceptanceDiagnostics)$' > M3.log 2>&1; acceptance_test.go:46: [node --disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/estree/main.ts /tmp/adamic-gate/TestAcceptanceDiagnostics2586825486/001/0000.ts]: timeout=false exit=exit status 70 stdout=0 stderr=adamic: panic: ESTree validator: update operand must be a left hand side expression",
    "timing_runs": [
      28.022,
      28.308,
      27.556
    ]
  },
  {
    "test": "TestAcceptanceDiagnosticControl",
    "members": [
      "TestAcceptanceDiagnosticControl"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/acceptance_test.go:53",
    "seconds": 26.244,
    "oracle": "Live Go cohere refusal plus self-written 0 Program substring for the validation-disabled port on source Node/native. The refusal checker it claims to witness is never invoked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
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
      "TestAcceptanceDiagnosticControl"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestAcceptanceDiagnosticControl)$' > W3.log 2>&1; --- PASS: TestAcceptanceDiagnosticControl (21.19s)",
    "timing_runs": [
      26.244,
      26.527,
      24.368
    ],
    "weakened_check_survivors": [
      "W3"
    ]
  },
  {
    "test": "TestRepositoryAgreement",
    "members": [
      "TestRepositoryAgreement"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/corpus_test.go:14",
    "seconds": 0.011,
    "oracle": "Frozen Go cohere canonical answers, not checked against Go this session because original snapshot/answers are absent.",
    "oracle_kind": "external-authority",
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
    "evidence": "go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestRepositoryAgreement$ > timing-7-0.log 2>&1; corpus_test.go:17: set ADAMIC_ESTREE_CORPUS to a completed Go/Node corpus audit directory",
    "timing_runs": [
      0.01,
      0.011,
      0.013
    ],
    "reason": "Completed frozen corpus unavailable; all three runs skipped. Seconds measure only skipped test-binary invocation."
  },
  {
    "test": "TestCorpusNativeRefusals",
    "members": [
      "TestCorpusNativeRefusals"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/corpus_test.go:81",
    "seconds": 0.02,
    "oracle": "Frozen Go cohere refusal statuses plus self-written panic/empty-stdout/deadline checks; specific refusal reasons are not compared. Frozen values not checked this session.",
    "oracle_kind": [
      "external-authority",
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
    "evidence": "go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run ^TestCorpusNativeRefusals$ > timing-8-0.log 2>&1; corpus_test.go:84: completed frozen corpus required",
    "timing_runs": [
      0.063,
      0.02,
      0.01
    ],
    "reason": "Completed frozen corpus unavailable; all three runs skipped. Seconds measure only skipped test-binary invocation."
  },
  {
    "test": "TestDeadlineChild",
    "members": [
      "TestDeadlineChild"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/deadlines_test.go:21",
    "seconds": 0.01,
    "oracle": "Subprocess entry only; active when parent sets ADAMIC_ESTREE_DEADLINE_CHILD=1.",
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
    "evidence": "go test -json -count=1 ./stage1/cohere/estree/ -run ^TestDeadlineChild$ > timing-9-0.log 2>&1; --- PASS: TestDeadlineChild (0.00s), dormant entry; production diagnostics invoke the active child.",
    "timing_runs": [
      0.009,
      0.01,
      0.091
    ],
    "parents_in_unit": [
      "TestAcceptanceDiagnostics",
      "TestCorpusNativeRefusals"
    ]
  },
  {
    "test": "TestDeepMutants_Setup",
    "members": [
      "TestDeepMutants_Setup"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/deep_mutants_product_shards_test.go:215",
    "seconds": 1.728,
    "oracle": "Successful frozen-fixture construction; self-written builder error checks, no want-file validation.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S6: deep_mutants_product_shards_test.go:202: build deep-mutants-oracle-output-v2: open /tmp/u084/cache/S6/.building-b1ba8a73fb0e-2060911392/missing/want: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestDeepMutants_Setup"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestDeepMutants_Setup)$' > S6.log 2>&1; deep_mutants_product_shards_test.go:202: build deep-mutants-oracle-output-v2: open /tmp/u084/cache/S6/.building-b1ba8a73fb0e-2060911392/missing/want: no such file or directory",
    "timing_runs": [
      4.008,
      1.087,
      1.728
    ],
    "setup_kills": [
      "S6"
    ],
    "construction_survivors": [
      "S3"
    ]
  },
  {
    "test": "TestDeepMutants family",
    "members": [
      "TestDeepMutants_000",
      "TestDeepMutants_001",
      "TestDeepMutants_002"
    ],
    "package": "stage1/cohere/estree",
    "file": "stage1/cohere/estree/deep_mutants_product_shards_test.go:286",
    "seconds": 9.007,
    "oracle": "Live Go cohere bytes compared to three built-in port mutants on Node/native; same byte predicate also guards planted disagreement and shard ownership.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: deep_mutants_product_shards_test.go:287: plant caught by []",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestAcceptanceMutants_000",
      "TestAcceptanceMutants_001",
      "TestAcceptanceMutantsUnion",
      "TestDeepMutants_000",
      "TestDeepMutants_001",
      "TestDeepMutants_002"
    ],
    "evidence": "timeout 100 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^(TestAcceptanceMutants_000|TestAcceptanceMutants_001|TestAcceptanceMutantsUnion|TestDeepMutants_000|TestDeepMutants_001|TestDeepMutants_002)$' > W1.log 2>&1; deep_mutants_product_shards_test.go:287: plant caught by []",
    "timing_runs": [
      59.421,
      9.007,
      8.206
    ],
    "witness_kills": [
      "W1"
    ]
  }
]
```

| ID | File:line at starting commit | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/estree/protocol.ts:20 | `unit >= 32 && unit <= 126 && unit !== 92` -> `unit >= 33 && unit <= 126 && unit !== 92` | TestAcceptanceGrammar |
| M2 | stage1/cohere/estree/protocol.ts:36 | `depth = 0` -> `depth = 1` | TestAcceptanceGrammar |
| M3 | stage1/cohere/estree/pipeline.ts:45 | `ESTree parser: ${syntax}` -> `ESTree validator: ${syntax}` | TestAcceptanceDiagnostics |
| M4 | stage1/cohere/estree/pipeline.ts:45 | `return panic(`ESTree parser: ${syntax}`);` -> `return panic('ESTree parser: rejected');` |  |

M4 survivor: adamic: panic: ESTree parser: update operand must be a left hand side expression -> adamic: panic: ESTree parser: rejected. Both direct Node runs have the same nonzero exit; both production rows pass. This leaves diagnostic detail unguarded in this bounded matrix.

P1 is an empty-answer probe, never a mutant kill. W1 disables firstDifference; W2 disables the mutant-survival comparison; W3 bypasses the refusal checker. S1/S3 omit artifact writes; S2 returns nonexistent product paths. S4/S5/S6 route artifact writes into missing subdirectories. Every experiment has a standalone diff and log.

| Harness/probe ID | Starting location | Observed failed functions |
|---|---|---|
| P1 | stage1/cohere/estree/pipeline.ts:10 | TestAcceptanceDiagnostics, TestAcceptanceGrammar |
| W1 | stage1/cohere/estree/estree_test.go:164 | TestAcceptanceMutants_000, TestAcceptanceMutants_001, TestDeepMutants_000, TestDeepMutants_001, TestDeepMutants_002 |
| W2 | stage1/cohere/estree/acceptance_grain_test.go:38 | TestAcceptanceMutantsPlanted_000 |
| W3 | stage1/cohere/estree/stalls_test.go:14 | none |
| S1 | stage1/cohere/estree/acceptance_grain_test.go:118 | none |
| S2 | stage1/cohere/estree/acceptance_grain_test.go:44 | none |
| S3 | stage1/cohere/estree/deep_mutants_product_shards_test.go:206 | none |
| S4 | stage1/cohere/estree/acceptance_grain_test.go:118 | TestProduct_AcceptanceOracle |
| S5 | stage1/cohere/estree/acceptance_grain_test.go:61 | TestProduct_AcceptanceCatchLowered, TestProduct_AcceptanceCatchNative, TestProduct_AcceptanceClassLowered, TestProduct_AcceptanceClassNative |
| S6 | stage1/cohere/estree/deep_mutants_product_shards_test.go:206 | TestDeepMutants_Setup |

The brief's scope count is inconsistent: it says 15 rows but lists 20 functions. All 20 exist at the fetched starting commit. Applying its shared-checker family rule produces 12 rows: four product wrappers share one builder, three acceptance shards/union share one family, two planted proofs share one checker, and three deep-mutant shards share one checker. Lowered/native booleans are builder inputs. The separate oracle builder and fixture-only setup have different bodies and stay separate.

The pinned file list is stale. The fetched origin/main is 12e77e8972a2e606cab6db05d84f428246a85339, not the brief's 8de93800f4. TestDeepMutants_Setup and its three shard functions are now in deep_mutants_product_shards_test.go. The required test listing prevented these moved tests from disappearing from the audit.

The default whole package exhausted its 90-second binary budget after grammar passed in 58.76 seconds. No preceding test failure made that baseline red. Explicit native splitting brought the bounded grammar/diagnostics baseline to 52.116 seconds. I retained the original timed-out run and used a bounded matrix rather than calling the package baseline a pass. Unassigned and corpus-dependent kills remain unknown. The matrix's outer timeout is 100 seconds rather than the prescribed 120; this is a stricter compile backstop, and no bounded run reached either limit.

Most assigned functions are witnesses, product builders, a helper, or corpus opt-ins. Production mutants would not establish their claimed witness/setup behavior. Those rows therefore have separate harness experiments, no production kills, and null entry-vacuity results. The production scope has two runnable rows. I used four predeclared production mutants and one probe, following the stage1 per-mutant rebuild fallback rather than adding a runtime selector to the port. Each experiment has its own ADAMIC_BUILD_CACHE_DIR and rebuild-inclusive timing. The 326-function TypeScript inventory is a conservative static superset; I did not produce exact dynamic function reachability. Mutations are limited to answer, dump and written, so no claim is made about unmutated parser/converter branches.

The product/setup rows catch construction errors when writes target nonexistent subdirectories. They do not validate the successful products: dropping oracle/want writes still passes, and returning nonexistent lowered/native paths passes all four product wrappers. The report calls these setup-checks because the error-propagation experiments fail them; it separately preserves the missing-artifact survivors and filesystem witnesses. These outcomes must not be interpreted as semantic agreement tests or counted as production mutant kills.

TestAcceptanceDiagnosticControl demonstrates actual source-Node/native acceptance of a Go-refused input after its built-in syntax-validation mutation. It then logs that the acceptance check catches the mutant, but invokes no refusal comparison. Bypassing refusedBeforeDeadline leaves it passing. Its verdict is untrue as a witness of that check, not a claim that the positive native-acceptance assertion can never fail. The acceptance and deep-mutant families did fail when their comparison was weakened. The deep family fails its planted-predicate proof before rebuilding mutants; that proves the shared comparison predicate, not a newly executed production mutation.

Both corpus rows skipped all three times. The metadata archive names a vanished /workspace/scratch/estree-corpus-final snapshot and answer tree. Installing dependencies cannot recreate those frozen artifacts. I did not substitute a tiny freshly generated corpus and present it as the promised repository audit. Their oracle provenance comes from the test's frozen-Go descriptions, with no outside value checked this session. Their reported seconds measure skipped invocation cost only.

Diagnostics checks error count and an error-class substring, not the specific Go refusal reason. M3 proves the substring guard is active; M4 proves diagnostic detail can change while it passes. TestCorpusNativeRefusals likewise checks generic panic/exit/deadline conditions by inspection, but its missing corpus prevents a produced mutant finding. No vacuity claim is made about either skipped corpus row.

Independent -count=1 runs still reuse content-addressed products. This is why median warm-family times are much lower than cold preparation. Both are recorded; cache hits are not omitted. Family seconds are the whole test binary's elapsed line when selecting that family's members, with no unrelated tests selected. Native rebuild times are reported as enclosing test durations because the ordinary build helper does not expose an isolated clang timer. Setup uses the already working environment. The /usr/bin/time utility was absent, so warm-env verification used Python's monotonic clock instead.


Warm setup skipped. go version go1.27.1 linux/amd64; 5; warm env check wall 0.009004s; setup skipped;  npm ci reported 449 ms. Whole baseline: 90.025 s, timed out after grammar passed in 58.76 s; no preceding test failure. Bounded clean baseline: 52.116 s. All bounded runs use ADAMIC_NATIVE_SPLIT=1 and ADAMIC_NATIVE_JOBS=5.

Native rebuild-inclusive test times are upper bounds, not isolated clang measurements: M1 25.65 s, M2 25.6 s, M3 34.03 s, M4 28.75 s, P1 28.69 s. Exact cold product build lines are in timings.json.

Recorded timing command wall total: 500.31 s. Experiment command wall total: 459.521 s. Go vet overhead is additional and was not separately timed.

Not covered: absent frozen corpus, unassigned package rows, package/repo-wide uniqueness, dynamic per-function reachability, and production-entry probes for witness/setup/helper rows. Their vacuous fields remain null. No production mutant edits the Go cohere oracle. Production source is restored before the evidence commit.
