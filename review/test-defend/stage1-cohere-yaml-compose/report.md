TestComposeMatchGo: defended using D1 stage1/cohere/yaml/composer.ts:62
TestCSTMatchesGo: defended using D2 stage1/cohere/yaml/cstParser.ts:717
TestFileDriver family: defended using D3 stage1/cohere/yaml/main.ts:36

[
  {
    "test": "TestComposeMatchGo",
    "package": "stage1/cohere/yaml",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestFormatterMatchesGo"
    ],
    "defense": "defended",
    "unique_mutant": "D1 stage1/cohere/yaml/composer.ts:62",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "stage1/cohere/yaml/composer.ts:62",
        "change": "this.document.warnings.push(new ComposeError(token.offset, token.offset + token.source.length, code, message)); -> this.document.warnings.push(new ComposeError(token.offset, token.offset + token.source.length + 1, code, message));",
        "rows_failed": [
          "TestComposeMatchGo"
        ]
      }
    ],
    "evidence": "ADAMIC_NATIVE_SPLIT=1 ADAMIC_YAML_LIBRARY=/tmp/defend-yaml/library ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestComposeMatchGo)$'; compose_test.go:148: native ASan/UBSan/LSan: byte 1709597: got \"02a007200650066|0|0|||||0|0|1|007200650066|||0|[]|-|-,]|-|-|errors:|warnings:3,32,TAG_RESOLVE_FAILED,0055006e007200650073006f006c0076006500640020007400610067003\", want \"02a007200650066|0|0|||||0|0|1|007200650066|||0|[]|-|-,]|-|-|errors:|warnings:3,31,TAG_RESOLVE_FAILED,0055006e007200650073006f006c0076006500640020007400610067003\"",
    "bounded": true,
    "matrix_rows": [
      "TestComposeMatchGo",
      "TestFileDriver family",
      "TestFileDriverUnion",
      "TestFormatterMatchesGo",
      "TestUnistMatchesGo"
    ],
    "rows_passed": [
      "TestFileDriver family",
      "TestFileDriverUnion",
      "TestFormatterMatchesGo",
      "TestUnistMatchesGo"
    ],
    "unknown_rows": [],
    "diff": "D1.diff",
    "failed_members": [
      "TestComposeMatchGo"
    ],
    "members": []
  },
  {
    "test": "TestCSTMatchesGo",
    "package": "stage1/cohere/yaml",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestFileDriver family"
    ],
    "defense": "defended",
    "unique_mutant": "D2 stage1/cohere/yaml/cstParser.ts:717",
    "attempts": [
      {
        "mutant": "D2",
        "file_line": "stage1/cohere/yaml/cstParser.ts:717",
        "change": "this.stack.push(this.make('doc-end', this.offset, 0, this.source, false)); -> this.stack.push(this.make('doc-end', this.offset, 0, this.source, true));",
        "rows_failed": [
          "TestCSTMatchesGo"
        ]
      }
    ],
    "evidence": "ADAMIC_NATIVE_SPLIT=1 ADAMIC_YAML_LIBRARY=/tmp/defend-yaml/library ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml/cache/D2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestCSTMatchesGo)$'; cst_test.go:78: native ASan/UBSan/LSan: byte 1395977: got \"|0062||[]|-|-|[newline|22|0|000a||[]|-|-|-|[]|[],]|[]|[]|0,]|-|[]|[]\\ndoc-end|23|0|002e002e002e||[]|-|-|[newline|26|0|000a||[]|-|-|-|[]|[],]|[]|[]\\nlines 0,14,18,\", want \"|0062||[]|-|-|[newline|22|0|000a||[]|-|-|-|[]|[],]|[]|[]|0,]|-|[]|[]\\ndoc-end|23|-|002e002e002e||[]|-|-|[newline|26|0|000a||[]|-|-|-|[]|[],]|[]|[]\\nlines 0,14,18,\"",
    "bounded": true,
    "matrix_rows": [
      "TestCSTMatchesGo",
      "TestComposeMatchGo",
      "TestFileDriver family",
      "TestFileDriverUnion",
      "TestFormatterMatchesGo",
      "TestPropsMatchGo",
      "TestScalarsMatchGo family",
      "TestUnistMatchesGo"
    ],
    "rows_passed": [
      "TestComposeMatchGo",
      "TestFileDriver family",
      "TestFileDriverUnion",
      "TestFormatterMatchesGo",
      "TestPropsMatchGo",
      "TestScalarsMatchGo family",
      "TestUnistMatchesGo"
    ],
    "unknown_rows": [],
    "diff": "D2.diff",
    "failed_members": [
      "TestCSTMatchesGo"
    ],
    "members": []
  },
  {
    "test": "TestFileDriver family",
    "package": "stage1/cohere/yaml",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestFormatterMatchesGo"
    ],
    "defense": "defended",
    "unique_mutant": "D3 stage1/cohere/yaml/main.ts:36",
    "attempts": [
      {
        "mutant": "D3",
        "file_line": "stage1/cohere/yaml/main.ts:36",
        "change": "const result = format(read.text); -> const result = format(read.text.slice(0, -1));",
        "rows_failed": [
          "TestFileDriver family"
        ]
      }
    ],
    "evidence": "ADAMIC_NATIVE_SPLIT=1 ADAMIC_YAML_LIBRARY=/tmp/defend-yaml/library ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml/cache/D3 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestFileDriver_[0-9]+)$'; file_driver_shards_test.go:260: native file 43: byte 1: got \">\\n\", want \">+\"",
    "bounded": true,
    "matrix_rows": [
      "TestFileDriver family",
      "TestFileDriverUnion",
      "TestFormatterMatchesGo"
    ],
    "rows_passed": [
      "TestFileDriverUnion",
      "TestFormatterMatchesGo"
    ],
    "unknown_rows": [],
    "diff": "D3.diff",
    "failed_members": [
      "TestFileDriver_001",
      "TestFileDriver_002",
      "TestFileDriver_003",
      "TestFileDriver_005"
    ],
    "members": [
      "TestFileDriver_000",
      "TestFileDriver_001",
      "TestFileDriver_002",
      "TestFileDriver_003",
      "TestFileDriver_004",
      "TestFileDriver_005",
      "TestFileDriver_006",
      "TestFileDriver_007"
    ]
  }
]

Bounded matrices inspect functional agreement guards. Other packages and production-independent setup/witness rows remain outside these uniqueness claims. All current names were compared with the audit: 72 names, no additions or disappearances.

Friction and limits

The outer whole-package baseline cooked at 120 seconds; file-driver setup passed at 76.89 seconds before that. Clean solo composer/CST/file coverage runs passed. The optional library install initially omitted yaml-unist-parser@3.2.0; TestUnistMatchesGo failed solely with ERR_MODULE_NOT_FOUND, then passed after installation, before mutations. Formatter clean runs cooked at 90 seconds; supported split compilation reached its independent original-library comparison after port agreement had completed. Mutant formatter runs and final restored runs are recorded separately. No timeout is counted as a kill.

Go -coverpkg profiles do not instrument TypeScript. The file-driver parent misses its cached setup child, so its apparent lack of compiler coverage is not proof of exclusive port execution. Semantic differences and mutation evidence establish the defenses. The supplied audit evidence excerpts were truncated; the full prior report, matrix, inventory, and plan are preserved.

These rows name semantic agreement, and assert exact external observations; no missing performance threshold or name/assertion mismatch was found. No separate executor twins exist among the three target rows. No tests, oracle adapters, or original library code were changed.

Timing: warm toolchain, setup skipped, nproc=5. Initial npm install times were not separately instrumented. Matrix shell wall total 842.21 seconds, including lowering, native rebuilds, Go oracle builds, and external executors. Per-command times and distinct caches are in matrix.json. Go coverage commands and timings are in coverage-runs.json. No separate lowering-versus-clang timing is claimed.

Final restored formatter: PASS, 64.865 test-binary seconds, full optional original-library comparison enabled. Cold file-driver setup coverage: PASS, 41.941 seconds. go vet ./stage1/cohere/yaml/: PASS. Source diff after restoration: empty. All three standalone diffs apply independently to the starting source.

The file-combined.cover profile merges cold setup-child coverage with the family parent. coverage-differences-final.json uses that merged profile and the final passing formatter profile. Earlier partial profiles are retained as evidence of the measurement limitation. GOGC=200 reduced Go build overhead for the D3 formatter retry and final clean coverage checks; it changed no assertions or program sources.
