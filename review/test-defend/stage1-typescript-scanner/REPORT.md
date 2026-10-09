Defender scanner unit at origin/main 76c59c81e8617cea1892a01841494895927a712c: all 29 audit tests still exist, no new top-level tests.
Two gap rows defended by package-unique D1/D2; profile row not defended after three valid attempts.
Evidence: test-defend/stage1-typescript-scanner under review/test-defend/stage1-typescript-scanner/.

```json
[
  {
    "test": "TestGapStandsWhereGapsMdSays",
    "package": "stage1/typescript/scanner",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestBigintGapStandsWhereGapsMdSays"
    ],
    "defense": "defended",
    "unique_mutant": "D1 internal/lower/object.go:1266",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/lower/object.go:1266",
        "change": "\"push with other than one value\" -> \"push with two values\"",
        "rows_failed": [
          "TestGapStandsWhereGapsMdSays"
        ]
      }
    ],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/defend-scanner/typescript ADAMIC_SCANNER_BENCH=1 ADAMIC_SCANNER_PROFILE_DIR=/tmp/defend-scanner/artifacts/D1 ADAMIC_SCANNER_PROFILE_SNAPSHOTS=/tmp/defend-scanner/artifacts/D1 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-scanner/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run . > D1.log 2>&1; gaps_test.go:34: gap changed: /workspace/adamic/stage1/typescript/scanner/gaps/1_push.ts:3:1: stage 0 can't lower push with two values yet; update GAPS.md and undo gap 1 workaround",
    "rows_passed": [
      "TestBigintGapStandsWhereGapsMdSays",
      "TestPerformance",
      "TestProduct_ScannerDecimalSeparator",
      "TestProduct_ScannerNative",
      "TestProduct_ScannerOracle",
      "TestProduct_ScannerPunctuator",
      "TestProduct_ScannerRegexRescan",
      "TestProduct_ScannerRelease",
      "TestProduct_ScannerSourceIdentifier",
      "TestProfileArtifacts",
      "TestProfileSnapshotsAgree",
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
      "TestScannerAgreesWithTypescriptGo_015",
      "TestScannerShardCoverage"
    ]
  },
  {
    "test": "TestBigintGapStandsWhereGapsMdSays",
    "package": "stage1/typescript/scanner",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestGapStandsWhereGapsMdSays"
    ],
    "defense": "defended",
    "unique_mutant": "D2 internal/lower/expression.go:40",
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "internal/lower/expression.go:40",
        "change": "\"a value of type \"+l.checker.TypeToString(l.checker.GetTypeAtLocation(node)) -> \"a representation of type \"+l.checker.TypeToString(l.checker.GetTypeAtLocation(node))",
        "rows_failed": [
          "TestBigintGapStandsWhereGapsMdSays"
        ]
      }
    ],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/defend-scanner/typescript ADAMIC_SCANNER_BENCH=1 ADAMIC_SCANNER_PROFILE_DIR=/tmp/defend-scanner/artifacts/D2 ADAMIC_SCANNER_PROFILE_SNAPSHOTS=/tmp/defend-scanner/artifacts/D2 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-scanner/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run . > D2.log 2>&1; gaps_test.go:59: gap changed: /workspace/adamic/stage1/typescript/scanner/gaps/2_bigint.ts:2:7: stage 0 can't lower a representation of type 1237940039285380274899124223n yet; update GAPS.md and undo gap 2 workaround",
    "rows_passed": [
      "TestGapStandsWhereGapsMdSays",
      "TestPerformance",
      "TestProduct_ScannerDecimalSeparator",
      "TestProduct_ScannerNative",
      "TestProduct_ScannerOracle",
      "TestProduct_ScannerPunctuator",
      "TestProduct_ScannerRegexRescan",
      "TestProduct_ScannerRelease",
      "TestProduct_ScannerSourceIdentifier",
      "TestProfileArtifacts",
      "TestProfileSnapshotsAgree",
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
      "TestScannerAgreesWithTypescriptGo_015",
      "TestScannerShardCoverage"
    ]
  },
  {
    "test": "TestProfileSnapshotsAgree",
    "package": "stage1/typescript/scanner",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestScannerAgreesWithTypescriptGo family"
    ],
    "defense": "not defended",
    "unique_mutant": null,
    "attempts": [
      {
        "mutant": "D3",
        "file_line": "internal/native/runtime/heap.c:149",
        "change": "each->free = ((free_slot *)slot)->next; -> [statement dropped]",
        "rows_failed": [
          "TestPerformance",
          "TestProfileSnapshotsAgree"
        ]
      },
      {
        "mutant": "D4",
        "file_line": "internal/native/runtime/heap.c:152",
        "change": "each->fresh += size; -> each->fresh += size - 1;",
        "rows_failed": [
          "TestPerformance",
          "TestProfileSnapshotsAgree"
        ]
      },
      {
        "mutant": "D5",
        "file_line": "internal/native/native.go:106",
        "change": "return append(flags, \"-O2\") -> return append(flags, \"-O2\", \"-funsafe-math-optimizations\")",
        "rows_failed": []
      }
    ],
    "evidence": "ADAMIC_TYPESCRIPT_SOURCE=/tmp/defend-scanner/typescript ADAMIC_SCANNER_BENCH=1 ADAMIC_SCANNER_PROFILE_DIR=/tmp/defend-scanner/artifacts/D4 ADAMIC_SCANNER_PROFILE_SNAPSHOTS=/tmp/defend-scanner/artifacts/D4 ADAMIC_BUILD_CACHE_DIR=/tmp/defend-scanner/cache/D4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/scanner/ -run . > D4.log 2>&1; profile_test.go:105: /tmp/defend-scanner/artifacts/D4/scanner [--manifest /tmp/adamic-gate/TestProfileSnapshotsAgree3810207524/001/all.txt]: signal: segmentation fault",
    "name_assertion_finding": "No mismatch: the row asserts complete answer bytes against TypeScript-Go for release, profiled and Node runs. It does not compare the counted artifact."
  }
]
```

CODE UNDER TEST AND ORACLES

Push gap: Adamic lowering of a two-argument array push. Node runs the unchanged fixture and supplies ab; the exact NotYet category/What diagnostic is a self-written expectation. Bigint gap: Adamic lowering of a bigint literal representation. Node supplies the unchanged bigint result; NotYet and its exact What are self-written. Profile snapshots: the port executed in native release and separately compiled profiling products, plus Node, including Adamic native allocator/runtime and compiler options; pinned TypeScript-Go supplies the complete expected protocol. No oracle, test, harness, dispatch, corpus or copied-file list was mutated.

Coverage

Per-row go test -coverprofile runs use -coverpkg=github.com/system-inc/adamic/internal/lower for both gap rows and -coverpkg=github.com/system-inc/adamic/internal/native for snapshots and all sixteen agreement members. Exact commands and passes are in coverage-runs.json. Push has 489 exclusive covered blocks versus bigint; bigint has 13 versus push. These are blocks, not an invented statement-line count. push-exclusive.txt and bigint-exclusive.txt give source ranges, including object.go:1266 and expression.go:40 respectively.

Snapshots has no exclusive Go covered block versus the agreement family (one package-initialization block versus 1154 covered blocks). Go coverage cannot instrument already built C/Node subprocesses, so that comparison does not establish that the semantic executions are redundant. The visible build recipes establish a real difference: agreement uses sanitized native products while snapshots also executes release/profiling products. heap.c sets SLABS=0 under ASan by default and SLABS=1 in release; take is bypassed in the sanitized allocation path. D3/D4 target that release-only behavior, and D5 changes only the non-sanitized branch of native.Flags. Count-only TestPerformance also uses a release product, which explains its observed catches of both allocator corruptions.

| Mutant | Origin file:line | Intent | Failed tests |
|---|---|---|---|
| D1 | internal/lower/object.go:1266 | exclusive multi-argument push refusal diagnostic | TestGapStandsWhereGapsMdSays |
| D2 | internal/lower/expression.go:40 | exclusive unsupported bigint representation diagnostic | TestBigintGapStandsWhereGapsMdSays |
| D3 | internal/native/runtime/heap.c:149 | release slab free-list reuse, bypassed by sanitizer allocator | TestPerformance, TestProfileSnapshotsAgree |
| D4 | internal/native/runtime/heap.c:152 | release slab fresh-slot bound, bypassed by sanitizer allocator | TestPerformance, TestProfileSnapshotsAgree |
| D5 | internal/native/native.go:106 | release-only compiler floating-point option; sanitized branch returns before it | none |

D1 and D2 each have exactly one observed failing top-level test. All other 28 passed; their exact names are in rows.json and defense-runs.json, including all sixteen agreement members, all seven product wrappers, coverage, performance, profile construction/snapshots and the other gap row. Every valid matrix ran all 29 tests with all profile/benchmark opt-ins enabled and fresh mutant cache/profile paths. No skips, cooking or unobserved cells occurred. Runtime crashes happened in child binaries and did not abort the Go test process.

D3 drops the entire free-list advance assignment, causing reuse of the allocated slot. D4 advances fresh slots one byte less than their size. Both produce segmentation faults caught by snapshots and performance. D5 adds -funsafe-math-optimizations only to release options; all 29 tests pass it. A separate unchanged-source rounding program with runtime input 1e20 prints 0 in normal native and Node, but 1 with D5. Its source, exact CLI build/run commands, and outputs are saved as D5-witness-*; this is a witnessed survivor, not a defense of snapshots. The three valid profile attempts are D3/D4/D5.

Instruction ambiguities and costs

The audit report was split across REPORT.md, rows.json, REPLAY.md and inventory.md; all were read before test/source inspection. Copies are retained under audit-source. Its starting commit was cf735d9f; this defense starts at fetched main 76c59c81. All 29 names match, so there were no added tests to omit from the matrix. Families affect row uniqueness; here D1/D2 each fail a single top-level test even before grouping, so the uniqueness result does not depend on that grouping.

Go coverage profiles are required, but they do not cover native C or Node. I report that limitation and use the actual release-versus-sanitized execution difference rather than pretending zero Go coverage proves the port was untouched. This is the main coverage limitation; no exclusive native C line count is claimed.

The repository corpus collector refuses dirty tracked inputs. Runtime C mutants were committed as temporary variants and followed by restoration commits, so the collector stayed intact. There was no history rewrite. I paused the first runner before the runtime attempts to add this requirement. Every final diff applies independently to the starting origin/main, and the final source tree matches it. Artifacts were freshly rebuilt per mutant; original baseline snapshots were not reused to judge a mutated compiler.

The first D5 candidate used -ffast-math. Go vet passed, but actual native runtime compilation failed under -Werror,-Wnan-infinity-disabled. Those failures are invalid and not kills. The rejected candidate and log are labeled D5-fastmath-invalid; invalid-runs.json records it separately. The corrected D5 keeps NaN/Infinity support and successfully builds every product. Both forms change production compiler options, not the clang harness or warning policy.

The standalone rounding witness initially passed a number to console.log, which Adamic types reject. Its final unchanged source interpolates a string, then builds successfully in both compiler variants. This witness does not run another package test or change a test oracle. Its Node comparison is a direct external execution.

The bias-to-keep requirement and not-defended verdict are compatible: not defended means these three targeted mutations establish no unique catch, not that the row is useless or may be deleted. No test was deleted, rewritten or weakened. For the undefended row, the name promises snapshots agreeing and the assertion checks complete answer bytes against TypeScript-Go. There is no missing threshold promised by that name. The counted binary is built by Artifacts but omitted from snapshots comparison. The two defended gap names mention where a gap stands, but their assertions check the fixture result and diagnostic What/category rather than asserting diagnostic Where or reading GAPS.md.

Warm tools were used, with no setup script; npm ci ran in stage3/api. The baseline took 39.026 binary seconds. Coverage commands took 39.092 wall seconds. Valid matrix commands including validation took 216.266 wall seconds; per-run durations are in defense-runs.json. No binary run exceeded 90 seconds; the longest valid mutant binary was 40.335 seconds. nproc=5. Total session approximately 16 minutes. Other package tests and repository-wide uniqueness were not covered.
