Unit u159: 29 top-level tests, grouped into 8 rows at cf735d9fba9e.
Enabled baseline passed in 34.372 seconds; no skips; nproc 5.
Four primary mutants were caught: 1 sacred row, 4 subsumed hints and 3 setup-check rows.
Product family passed its empty-path probe; semantic rows failed their own probes.
All source edits restored; standalone diffs, raw logs and matrices accompany this report.

```json
[
  {
    "test": "TestGapStandsWhereGapsMdSays",
    "package": "stage1/typescript/scanner",
    "file": "stage1/typescript/scanner/gaps_test.go:13",
    "seconds": 0.1,
    "oracle": "Node runs the fixture and returns ab; lower.NotYet and exact What label are self-written. Location and execution of lowered output are not checked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: gaps_test.go:34: gap changed: /workspace/adamic/stage1/typescript/scanner/gaps/1_push.ts:3:1: stage 0 can't lower unsupported yet; update GAPS.md and undo gap 1 workaround",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestBigintGapStandsWhereGapsMdSays"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 0.098,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run . > M4.log 2>&1; gaps_test.go:34: gap changed: /workspace/adamic/stage1/typescript/scanner/gaps/1_push.ts:3:1: stage 0 can't lower unsupported yet; update GAPS.md and undo gap 1 workaround"
  },
  {
    "test": "TestBigintGapStandsWhereGapsMdSays",
    "package": "stage1/typescript/scanner",
    "file": "stage1/typescript/scanner/gaps_test.go:38",
    "seconds": 0.098,
    "oracle": "Node returns 1237940039285380274899124223; lower.NotYet and exact What label are self-written. Location is not checked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: gaps_test.go:59: gap changed: /workspace/adamic/stage1/typescript/scanner/gaps/2_bigint.ts:2:7: stage 0 can't lower unsupported yet; update GAPS.md and undo gap 2 workaround",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestGapStandsWhereGapsMdSays"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": 0.1,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run . > M4.log 2>&1; gaps_test.go:59: gap changed: /workspace/adamic/stage1/typescript/scanner/gaps/2_bigint.ts:2:7: stage 0 can't lower unsupported yet; update GAPS.md and undo gap 2 workaround"
  },
  {
    "test": "TestProfileArtifacts",
    "package": "stage1/typescript/scanner",
    "file": "stage1/typescript/scanner/profile_test.go:18",
    "seconds": 7.657,
    "oracle": "Self: file writes and compiler command success. S2 proves failed profile construction is detected; artifact behavior is checked in another row.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S2: clang: error: unknown argument: '-audit-invalid-profile-option'",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run ^TestProfileArtifacts$ > S2.log 2>&1; clang: error: unknown argument: '-audit-invalid-profile-option'",
    "construction_kills": [
      "S2"
    ]
  },
  {
    "test": "TestProfileSnapshotsAgree",
    "package": "stage1/typescript/scanner",
    "file": "stage1/typescript/scanner/profile_test.go:92",
    "seconds": 6.418,
    "oracle": "Pinned TypeScript-Go full bytes compared to release, profiled and Node port across all 18791 cases. Artifact directories freshly rebuilt for each mutant.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: profile_test.go:110: /tmp/u159/artifacts/M3-replay release: line 886390: port \"Identifier 0 1 0\\ta\", Go \"Identifier 0 4 0\\ta\\\\u200d\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestScannerAgreesWithTypescriptGo family"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 12.93,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run . > M3.log 2>&1; profile_test.go:110: /tmp/u159/artifacts/M3-replay release: line 886390: port \"Identifier 0 1 0\\ta\", Go \"Identifier 0 4 0\\ta\\\\u200d\""
  },
  {
    "test": "TestPerformance",
    "package": "stage1/typescript/scanner",
    "file": "stage1/typescript/scanner/scanner_agreement_shards_test.go:292",
    "seconds": 4.157,
    "oracle": "TypeScript-Go token count compared with native and Node counts. Only total count is asserted; token kinds, values, offsets and speed have no assertion. M1 and M3 survive this row.",
    "oracle_kind": "external-run",
    "kills": [
      "M2"
    ],
    "unique_kills": [
      "M2"
    ],
    "last_proven_fail": "M2: scanner_agreement_shards_test.go:317: native count \"869580\\n\", Go \"434790\\n\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run . > M2.log 2>&1; scanner_agreement_shards_test.go:317: native count \"869580\\n\", Go \"434790\\n\""
  },
  {
    "test": "TestProduct_Scanner family",
    "package": "stage1/typescript/scanner",
    "file": "stage1/typescript/scanner/scanner_products_test.go:137",
    "seconds": 2.98,
    "oracle": "Self: build recipes must complete without t.Fatal. Return paths are discarded, with no product existence or behavior assertion.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: scanner_products_test.go:139: scanner product preparation failed",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run ^TestProduct_Scanner > S1.log 2>&1; scanner_products_test.go:139: scanner product preparation failed",
    "members": [
      "TestProduct_ScannerOracle",
      "TestProduct_ScannerNative",
      "TestProduct_ScannerRelease",
      "TestProduct_ScannerSourceIdentifier",
      "TestProduct_ScannerPunctuator",
      "TestProduct_ScannerDecimalSeparator",
      "TestProduct_ScannerRegexRescan"
    ],
    "construction_kills": [
      "S1"
    ]
  },
  {
    "test": "TestScannerShardCoverage",
    "package": "stage1/typescript/scanner",
    "file": "stage1/typescript/scanner/scanner_products_test.go:289",
    "seconds": 0.371,
    "oracle": "Self: corpus union, exactly-once ownership, mutant ownership and one planted mismatch. Git supplies tracked inputs, not expected scanner semantics.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: scanner_products_test.go:329: planted failure caught by 0 shards",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P4"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run ^(TestScannerAgreesWithTypescriptGo_|TestScannerShardCoverage$) > W1.log 2>&1; scanner_products_test.go:329: planted failure caught by 0 shards",
    "construction_kills": [
      "W1"
    ]
  },
  {
    "test": "TestScannerAgreesWithTypescriptGo family",
    "package": "stage1/typescript/scanner",
    "file": "stage1/typescript/scanner/scanner_products_test.go:334",
    "seconds": 12.93,
    "oracle": "Pinned TypeScript-Go scanner, compared byte for byte to native and Node-executed port. Includes separate built-in mismatch witnesses, verified by W1.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: scanner_products_test.go:340: native: line 37545: port \"error 1127 0 2\", Go \"Identifier 0 2 0\\t\\\\u038e\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestProfileSnapshotsAgree"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 6.418,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run . > M3.log 2>&1; scanner_products_test.go:340: native: line 37545: port \"error 1127 0 2\", Go \"Identifier 0 2 0\\t\\\\u038e\"",
    "members": [
      "TestScannerAgreesWithTypescriptGo_000",
      "TestScannerAgreesWithTypescriptGo_001",
      "TestScannerAgreesWithTypescriptGo_002",
      "TestScannerAgreesWithTypescriptGo_003",
      "TestScannerAgreesWithTypescriptGo_004",
      "TestScannerAgreesWithTypescriptGo_005",
      "TestScannerAgreesWithTypescriptGo_006",
      "TestScannerAgreesWithTypescriptGo_007",
      "TestScannerAgreesWithTypescriptGo_008",
      "TestScannerAgreesWithTypescriptGo_009",
      "TestScannerAgreesWithTypescriptGo_010",
      "TestScannerAgreesWithTypescriptGo_011",
      "TestScannerAgreesWithTypescriptGo_012",
      "TestScannerAgreesWithTypescriptGo_013",
      "TestScannerAgreesWithTypescriptGo_014",
      "TestScannerAgreesWithTypescriptGo_015"
    ],
    "witness_evidence": "W1: scanner_products_test.go:356: source scanned as Identifier Node: mutant survives comparison; weakening difference also fails the other three built-in witness owners."
  }
]
```

| ID | Starting file:line | Change | Rows failed |
|---|---|---|---|
| M1 | stage1/typescript/scanner/scanner.ts:74 | `this.pos += code > 0xffff ? 2 : 1;` → `this.pos += code > 0xffff ? 1 : 1;` | TestProfileSnapshotsAgree, TestScannerAgreesWithTypescriptGo family |
| M2 | stage1/typescript/scanner/main.ts:58 | `count++;` → `count += 2;` | TestPerformance |
| M3 | stage1/typescript/scanner/characters.ts:250 | `else if(code > upper)` → `else if(code >= upper)` | TestProfileSnapshotsAgree, TestScannerAgreesWithTypescriptGo family |
| M4 | internal/lower/diagnostics.go:33 | `What: what` → `What: "unsupported"` | TestGapStandsWhereGapsMdSays, TestBigintGapStandsWhereGapsMdSays |

Survivors: none in the whole-package primary matrix. M1 and M3 survive the count-only benchmark but are caught by full-answer checks. No equivalent candidates.

The brief requires a clean baseline and scratch source mutations, but the repository corpus collector rejects dirty tracked TypeScript files. M1/M2/M3 first failed this guard. Those failures establish no semantic kill. Temporary local commits solved it without changing the collector; their hashes are saved. Each run was reset back to the starting commit before the next edit. Central replay needs a clean committed variant too, or the same guard will mask semantic results.

An unconditional early return in run made the remaining TypeScript body unreachable and disabled its flow narrowing, producing TS2339 rather than an empty answer. Removing the whole body fixed typing but removed a built-in regex mutation site, causing an unrelated preparation failure. Both attempts are retained as invalid. The final P1 uses an always-satisfied path.length >= 0 guard and preserves all mutation sites. Its product builds pass and semantic comparisons fail.

Family rules overlap with setup rules. All sixteen agreement wrappers call the same scannerShard checker, so they are one family. Coverage has its own body and distinct assertions, so it is a separate setup-check. Seven product wrappers share scannerProductTime with different build recipes, so they form one construction family. The initial agreement timings included Coverage; corrected family-only timings are saved and used.

The instruction to enable installable opt-ins required supplying a pinned TypeScript checkout, enabling the benchmark, creating profile artifacts and giving the snapshot row their directory. The deliberate ADAMIC_SCANNER_PLANT_FAILURE option is intentionally failing behavior, not a skipped row, and was not enabled for the baseline. Artifact directories are rebuilt for each mutation before snapshot execution; a clean snapshot directory reused under mutations would silently test the wrong executable.

The benchmark has a useful unique count guard but no performance threshold. Its name alone does not establish a speed regression gate. It passes both M1 and M3 even though full-answer rows show changed semantics. Conversely, the gap rows assert the same diagnostic constructor option and share only one measured kill. Their mutual subsumption rests on one mutant, not an argument to delete either. Agreement/snapshot subsumption rests on two mutants.

The product family detects a dropped build operation, yet every wrapper passes when scannerFetch returns an empty path. Its construction guard therefore does not prove that a usable product was returned. The profile-artifact row was checked by invalidating its profile compiler option; its empty-answer status is null because no single construction-entry empty probe was run for that row. This is an explicit coverage limit, not an inferred non-vacuity claim.

The desired three mutants per row conflicts with the four-rebuild limit for ports. Four primary changes were fixed before their results and spread over advance, run, contains and notYet. Construction edits and comparison weakening have separate S/W identifiers; entry probes have P identifiers. None contributes to primary uniqueness or subsumption. The single-switch suggestion was not used because each port mutation required a separate fresh native build; the explicit four-mutant rebuild allowance was followed.

Timing compiler preparation, semantic execution and external process work separately is not directly supplied by the brief's Go package line. runs.json records command wall time, and the requested medians use the test binary's own elapsed line. Buildcache miss timings and profile-build elapsed events remain in raw JSON. Their parallel durations must not be summed as a package wall time. Native products and the fetched corpus were not pushed; reproducible source diffs, commands and logs were.

Warm tool setup was skipped (0 setup-script seconds); npm ci reports 460 ms. Timed commands total 474.648 wall seconds, including invalid and repeated runs but excluding initial fetch/npm/baseline and separately unmeasured vet commands. The baseline adds 34.372 binary seconds. Corrected primary run wall times: M1 33.778 s, M2 28.89 s, M3 35.02 s, M4 40.757 s. Their profile construction test elapsed times: M1 8 s, M2 7.87 s, M3 7.76 s, M4 7.7 s. Total session about 18 minutes.

Every primary matrix ran all 29 top-level tests with the benchmark and profile options enabled; all rows have observed pass/fail results. No run exceeded 90 seconds, no Go panic truncated the matrix, and no row skipped. M1/M2/M3 compiled through the actual sanitized/release native recipes in their corrected whole-package runs; M4 passed go vet ./internal/lower/ and native builds. P/S/W Go changes passed go vet for their mutated package. Diffs apply to the starting origin/main commit. No other packages were audited, and repo-wide uniqueness remains for central replay. The port functions and scanner-package helpers are listed in function-inventory.txt; lower.cover/lower-functions.log list actual gap reaches. The entire transitive compiler graph used as port build preparation was not separately inventoried or mutated.
